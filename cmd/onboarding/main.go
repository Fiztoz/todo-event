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

	useradapter "todoe/domain/user/adapter"
	userhttp "todoe/domain/user/adapter/http"
	userapplication "todoe/domain/user/application"
	userdomain "todoe/domain/user/domain"

	"todoe/internal/event"
	"todoe/internal/messaging"
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
		dbName = "todoe_onboarding"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
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

	// ── User domain ──────────────────────────────────────────────────────
	userBus := event.NewEventBus()
	userRepo := useradapter.NewMongoRepository(clientIO, dbName)
	userProjection := useradapter.NewProjectionHandler(userRepo)
	userBus.Subscribe(userdomain.EventRegistered, userProjection)
	userBus.Subscribe(userdomain.EventEmailVerified, userProjection)
	userBus.Subscribe(userdomain.EventCreditScored, userProjection)
	userBus.Subscribe(userdomain.EventProfileCompleted, userProjection)
	userPublisher := &messaging.MultiPublisher{Publishers: []event.Publisher{
		userBus,
		&messaging.NatsPublisher{Conn: nc, Subject: messaging.UserSubject},
	}}
	userService := userapplication.NewService(userRepo, userPublisher)
	userHandler := userhttp.NewHandler(userService)

	// credit.results → RecordCreditScore (result from the credit service)
	nc.Subscribe(messaging.CreditResultSubject, func(m *nats.Msg) {
		var msg messaging.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			slog.Error("onboarding: credit result unmarshal", "err", err)
			return
		}
		if msg.Type != userdomain.EventCreditScored {
			return
		}
		var p userdomain.CreditScoredPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			slog.Error("onboarding: credit scored unmarshal", "err", err)
			return
		}
		id, err := bson.ObjectIDFromHex(p.UserID)
		if err != nil {
			slog.Error("onboarding: invalid user id in credit result", "err", err)
			return
		}
		if r := userService.RecordCreditScore(context.Background(), id, p.Score, p.Approved); r.IsError() {
			slog.Error("onboarding: record credit score", "err", r.Error())
		}
	})

	// ── HTTP ─────────────────────────────────────────────────────────────
	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)
	app.Post("/users/register", userHandler.Register)
	app.Get("/users/:id", userHandler.GetUser)
	app.Post("/users/:id/verify-email", userHandler.VerifyEmail)
	app.Post("/users/:id/complete-profile", userHandler.CompleteProfile)

	slog.Info("onboarding service starting", "port", port, "db", dbName)
	log.Fatal(app.Listen(":" + port))
}
