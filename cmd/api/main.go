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
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	healthadapter "todoe/internal/health/adapter"
	healthhttp "todoe/internal/health/adapter/http"
	healthapp "todoe/internal/health/application"

	taskadapter "todoe/domain/task/adapter"
	taskhttp "todoe/domain/task/adapter/http"
	taskapplication "todoe/domain/task/application"
	taskdomain "todoe/domain/task/domain"
	"todoe/internal/event"
	"todoe/internal/messaging"
)

type natsPublisher struct{ conn *nats.Conn }

func (p *natsPublisher) Publish(_ context.Context, e event.Event) {
	payload, _ := json.Marshal(e.Payload)
	data, _ := json.Marshal(messaging.Message{Type: e.Type, Payload: payload})
	if err := p.conn.Publish(messaging.TaskSubject, data); err != nil {
		slog.Error("nats: publish error", "err", err)
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

	bus := event.NewEventBus()
	taskRepo := taskadapter.NewMongoRepository(clientIO)
	projectionHandler := taskadapter.NewProjectionHandler(taskRepo)
	bus.Subscribe(taskdomain.EventCreated, projectionHandler)
	bus.Subscribe(taskdomain.EventStatusChanged, projectionHandler)

	publisher := &multiPublisher{publishers: []event.Publisher{bus, &natsPublisher{nc}}}
	taskService := taskapplication.NewService(taskRepo, publisher)
	taskHandler := taskhttp.NewHandler(taskService)

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)
	app.Post("/tasks", taskHandler.Create)
	app.Get("/tasks", taskHandler.List)
	app.Get("/tasks/:id", taskHandler.Detail)
	app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)

	log.Fatal(app.Listen(":3000"))
}
