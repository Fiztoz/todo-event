package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
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

	eventbus "todoe/internal/event"
	auditadapter "todoe/internal/audit/adapter"
	// slaadapter "todoe/internal/sla/adapter"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})

	healthRepo := healthadapter.NewMongoRepository(clientIO)
	defer healthRepo.Disconnect(context.Background())

	healthService := healthapp.NewService(healthRepo)
	healthHandler := healthhttp.NewHandler(healthService)
	eventBus := eventbus.NewEventBus()
	auditRepo := auditadapter.NewMongoRepository(clientIO)
	auditHandler := auditadapter.NewAuditHandler(auditRepo, "audit_fallback.jsonl")

	// slaRepo := slaadapter.NewMongoRepository(clientIO)
	// slaHandler := slaadapter.NewSlaHandler(slaRepo, "sla.csv")

	eventBus.Subscribe(taskdomain.EventCreated, auditHandler)
	eventBus.Subscribe(taskdomain.EventStatusChanged, auditHandler)
	// eventBus.Subscribe(taskdomain.EventStatusChanged, slaHandler)

	taskRepo := taskadapter.NewMongoRepository(clientIO, "task_fallback.jsonl")
	taskService := taskapplication.NewService(taskRepo, eventBus)
	taskHandler := taskhttp.NewHandler(taskService)

	app := fiber.New()
	app.Get("/health", healthHandler.CheckHealth)
	app.Post("/tasks", taskHandler.Create)
	app.Get("/tasks", taskHandler.List)
	app.Get("/tasks/:id", taskHandler.Detail)
	app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)

	log.Fatal(app.Listen(":3000"))
}
