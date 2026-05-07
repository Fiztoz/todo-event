package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	healthapp "todoe/internal/health/application"

	authenAdapter "todoe/internal/authen/adapter"
	authenhttp "todoe/internal/authen/adapter/http"
	authenapp "todoe/internal/authen/application"
	authendomain "todoe/internal/authen/domain"

	taskadapter "todoe/domain/task/adapter"
	taskhttp "todoe/domain/task/adapter/http"
	taskapplication "todoe/domain/task/application"
	taskdomain "todoe/domain/task/domain"

	useradapter "todoe/domain/user/adapter"
	userhttp "todoe/domain/user/adapter/http"
	userapplication "todoe/domain/user/application"
	userdomain "todoe/domain/user/domain"

	"todoe/internal/event"
	"todoe/internal/messaging"
)

type multiPublisher struct{ publishers []event.Publisher }

func (m *multiPublisher) Publish(ctx context.Context, e event.Event) {
	for _, p := range m.publishers {
		p.Publish(ctx, e)
	}
}

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	healthRepo := healthadapter.NewMongoRepository(clientIO)
	defer healthRepo.Disconnect(context.Background())
	healthService := healthapp.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	conn, ch, err := messaging.Connect(amqpURL)
	if err != nil {
		log.Fatal("rabbit:", err)
	}
	defer conn.Close()

	if err := messaging.DeclareTopology(ch, []messaging.Binding{
		{Exchange: messaging.TaskExchange, Queue: messaging.QueueAuditTaskEvents},
		{Exchange: messaging.UserExchange, Queue: messaging.QueueWelcomeUserEvents},
		{Exchange: messaging.UserExchange, Queue: messaging.QueueCreditUserEvents},
		{Exchange: messaging.CreditResultExchange, Queue: messaging.QueueAPICreditResults},
	}); err != nil {
		log.Fatal("rabbit topology:", err)
	}

	// ── Task domain ──────────────────────────────────────────────────────
	taskBus := event.NewEventBus()
	taskRepo := taskadapter.NewMongoRepository(clientIO)
	taskProjection := taskadapter.NewProjectionHandler(taskRepo)
	taskBus.Subscribe(taskdomain.EventCreated, taskProjection)
	taskBus.Subscribe(taskdomain.EventStatusChanged, taskProjection)
	taskPublisher := &multiPublisher{publishers: []event.Publisher{
		taskBus,
		messaging.NewPublisher(ch, messaging.TaskExchange),
	}}
	taskService := taskapplication.NewService(taskRepo, taskPublisher)
	taskHandler := taskhttp.NewHandler(taskService)

	// ── User domain ──────────────────────────────────────────────────────
	userBus := event.NewEventBus()
	userRepo := useradapter.NewMongoRepository(clientIO)
	userProjection := useradapter.NewProjectionHandler(userRepo)
	userBus.Subscribe(userdomain.EventRegistered, userProjection)
	userBus.Subscribe(userdomain.EventEmailVerified, userProjection)
	userBus.Subscribe(userdomain.EventCreditScored, userProjection)
	userBus.Subscribe(userdomain.EventProfileCompleted, userProjection)
	userPublisher := &multiPublisher{publishers: []event.Publisher{
		userBus,
		messaging.NewPublisher(ch, messaging.UserExchange),
	}}
	userService := userapplication.NewService(userRepo, userPublisher)
	userHandler := userhttp.NewHandler(userService)

	// ── Authen domain ────────────────────────────────────────────────────
	authenBus := event.NewEventBus()
	authenRepo := authenAdapter.NewMongoRepository(clientIO)
	authenProjection := authenAdapter.NewProjectionHandler(authenRepo)
	authenBus.Subscribe(authendomain.EventLoggedIn, authenProjection)
	authenBus.Subscribe(authendomain.EventLoggedOut, authenProjection)
	authenService := authenapp.NewService(authenRepo, authenBus)
	authenHandler := authenhttp.NewHandler(authenService)

	//user.activated → create auth credential for the newly onboarded user
	//subsribe domain events directly from internal bus since this is a same-process integration
	userBus.Subscribe(userdomain.EventUserActivated, func(ctx context.Context, e event.Event) error {
		p, ok := e.Payload.(userdomain.UserActivatedPayload)
		if !ok {
			slog.Error("api: user.activated unexpected payload", "type", fmt.Sprintf("%T", e.Payload))
			return nil
		}
		userID, err := bson.ObjectIDFromHex(p.UserID)
		if err != nil {
			slog.Error("api: user.activated invalid user id", "err", err)
			return nil
		}
		if r := authenService.ActivateUser(ctx, userID, p.Email, p.Name); r.IsError() {
			slog.Error("api: user.activated credential creation failed", "err", r.Error())
		}
		return nil
	})

	// ── Credit result loop-back ──────────────────────────────────────────
	// cmd/credit publishes credit.results → api calls RecordCreditScore
	if err := messaging.Subscribe(ch, messaging.CreditResultExchange, messaging.QueueAPICreditResults, func(msg messaging.Message) {
		if msg.Type != userdomain.EventCreditScored {
			return
		}
		var p userdomain.CreditScoredPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			slog.Error("api: credit scored unmarshal", "err", err)
			return
		}
		id, err := bson.ObjectIDFromHex(p.UserID)
		if err != nil {
			slog.Error("api: invalid user id in credit result", "err", err)
			return
		}
		if r := userService.RecordCreditScore(context.Background(), id, p.Score, p.Approved); r.IsError() {
			slog.Error("api: record credit score", "err", r.Error())
		}
	}); err != nil {
		log.Fatal("rabbit subscribe credit.results:", err)
	}

	// ── HTTP ─────────────────────────────────────────────────────────────
	authMiddleware := func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}
		if r := authenService.ValidateToken(c.Context(), token); r.IsError() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)

	app.Post("/auth/register", authenHandler.RegisterCredential)
	app.Post("/auth/login", authenHandler.Login)
	app.Post("/auth/logout", authenHandler.Logout)

	tasks := app.Group("/tasks", authMiddleware)
	tasks.Post("/", taskHandler.Create)
	tasks.Get("/", taskHandler.List)
	tasks.Get("/:id", taskHandler.Detail)
	tasks.Patch("/:id/status", taskHandler.ChangeStatus)

	app.Post("/users/register", userHandler.Register)
	app.Get("/users/:id", userHandler.GetUser)
	app.Post("/users/:id/verify-email", userHandler.VerifyEmail)
	app.Post("/users/:id/complete-profile", userHandler.CompleteProfile)

	log.Fatal(app.Listen(":3000"))
}
