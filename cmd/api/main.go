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

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	healthapp "todoe/internal/health/application"

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

type natsPublisher struct {
	conn    *nats.Conn
	subject string
}

func (p *natsPublisher) Publish(_ context.Context, e event.Event) {
	payload, _ := json.Marshal(e.Payload)
	data, _ := json.Marshal(messaging.Message{Type: e.Type, Payload: payload})
	if err := p.conn.Publish(p.subject, data); err != nil {
		slog.Error("nats: publish error", "subject", p.subject, "err", err)
	}
}

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
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

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
	taskRepo := taskadapter.NewMongoRepository(clientIO)
	taskProjection := taskadapter.NewProjectionHandler(taskRepo)
	taskBus.Subscribe(taskdomain.EventCreated, taskProjection)
	taskBus.Subscribe(taskdomain.EventStatusChanged, taskProjection)
	taskPublisher := &multiPublisher{publishers: []event.Publisher{
		taskBus,
		&natsPublisher{nc, messaging.TaskSubject},
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
		&natsPublisher{nc, messaging.UserSubject},
	}}
	userService := userapplication.NewService(userRepo, userPublisher)
	userHandler := userhttp.NewHandler(userService)

	// ── Credit result loop-back ──────────────────────────────────────────
	// cmd/credit publishes credit.results → api calls RecordCreditScore
	nc.Subscribe(messaging.CreditResultSubject, func(m *nats.Msg) {
		var msg messaging.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			slog.Error("api: credit result unmarshal", "err", err)
			return
		}
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
	})

	// ── HTTP ─────────────────────────────────────────────────────────────
	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)
	app.Post("/tasks", taskHandler.Create)
	app.Get("/tasks", taskHandler.List)
	app.Get("/tasks/:id", taskHandler.Detail)
	app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)
	app.Post("/users/register", userHandler.Register)
	app.Get("/users/:id", userHandler.GetUser)
	app.Post("/users/:id/verify-email", userHandler.VerifyEmail)
	app.Post("/users/:id/complete-profile", userHandler.CompleteProfile)

	log.Fatal(app.Listen(":3000"))
}
