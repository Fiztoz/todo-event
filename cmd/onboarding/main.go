package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

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

	conn, ch, err := messaging.Connect(amqpURL)
	if err != nil {
		log.Fatal("rabbit:", err)
	}
	defer conn.Close()

	if err := messaging.DeclareTopology(ch, []messaging.Binding{
		{Exchange: messaging.UserExchange, Queue: messaging.QueueWelcomeUserEvents},
		{Exchange: messaging.UserExchange, Queue: messaging.QueueCreditUserEvents},
		{Exchange: messaging.UserExchange, Queue: messaging.QueueAuthenUserEvents},
		{Exchange: messaging.CreditResultExchange, Queue: messaging.QueueOnboardingCreditResults},
	}); err != nil {
		log.Fatal("rabbit topology:", err)
	}

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

	if err := messaging.Subscribe(ch, messaging.CreditResultExchange, messaging.QueueOnboardingCreditResults, func(msg messaging.Message) {
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
	}); err != nil {
		log.Fatal("rabbit subscribe credit.results:", err)
	}

	app := fiber.New()
	app.Post("/users/register", userHandler.Register)
	app.Get("/users/:id", userHandler.GetUser)
	app.Post("/users/:id/verify-email", userHandler.VerifyEmail)
	app.Post("/users/:id/complete-profile", userHandler.CompleteProfile)

	log.Fatal(app.Listen(":3002"))
}
