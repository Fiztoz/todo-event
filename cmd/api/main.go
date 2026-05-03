package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	"todoe/internal/event"
	"todoe/internal/messaging"
	pb "todoe/onboarding"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "todoe"
	}
	onboardingAddr := os.Getenv("ONBOARDING_GRPC_ADDR")
	if onboardingAddr == "" {
		onboardingAddr = "localhost:50051"
	}

	conn, err := grpc.NewClient(onboardingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("api: grpc dial onboarding:", err)
	}
	defer conn.Close()
	onboardingClient := pb.NewOnboardingServiceClient(conn)

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	healthRepo := healthadapter.NewMongoRepository(clientIO)
	defer healthRepo.Disconnect(context.Background())
	healthService := healthapp.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("nats:", err)
	}
	defer nc.Drain()

	// ── Task domain ──────────────────────────────────────────────────────
	taskBus := event.NewEventBus()
	taskRepo := taskadapter.NewMongoRepository(clientIO, dbName)
	taskProjection := taskadapter.NewProjectionHandler(taskRepo)
	taskBus.Subscribe(taskdomain.EventCreated, taskProjection)
	taskBus.Subscribe(taskdomain.EventStatusChanged, taskProjection)
	taskPublisher := &messaging.MultiPublisher{Publishers: []event.Publisher{
		taskBus,
		&messaging.NatsPublisher{Conn: nc, Subject: messaging.TaskSubject},
	}}
	taskService := taskapplication.NewService(taskRepo, taskPublisher)
	taskHandler := taskhttp.NewHandler(taskService)

	// ── Authen domain ────────────────────────────────────────────────────
	authenBus := event.NewEventBus()
	authenRepo := authenAdapter.NewMongoRepository(clientIO, dbName)
	authenProjection := authenAdapter.NewProjectionHandler(authenRepo)
	authenBus.Subscribe(authendomain.EventLoggedIn, authenProjection)
	authenBus.Subscribe(authendomain.EventLoggedOut, authenProjection)
	authenService := authenapp.NewService(authenRepo, authenBus)
	authenHandler := authenhttp.NewHandler(authenService)

	// user.activated (NATS notification) → fetch full user via gRPC → create auth credential
	nc.Subscribe(messaging.UserSubject, func(m *nats.Msg) {
		var msg messaging.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			slog.Error("api: user event unmarshal", "err", err)
			return
		}
		if msg.Type != "user.activated" {
			return
		}
		var p struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			slog.Error("api: user.activated unmarshal", "err", err)
			return
		}
		id, err := bson.ObjectIDFromHex(p.UserID)
		if err != nil {
			slog.Error("api: user.activated invalid user id", "err", err)
			return
		}
		resp, err := onboardingClient.GetUser(context.Background(), &pb.GetUserRequest{UserId: p.UserID})
		if err != nil {
			slog.Error("api: user.activated grpc GetUser failed", "err", err)
			return
		}
		if r := authenService.ActivateUser(context.Background(), id, resp.Email, resp.Name); r.IsError() {
			slog.Error("api: user.activated credential creation failed", "err", r.Error())
		}
	})

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

	slog.Info("api service starting", "port", 3000, "db", dbName)
	log.Fatal(app.Listen(":3000"))
}
