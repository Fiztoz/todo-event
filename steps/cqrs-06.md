# CQRS Step 6 — Wire Main

## What changes and why

`cmd/api/main.go` is the composition root. Subscribe the projection handler to both task events and register the new `GET /tasks/:id` route.

## Target: `cmd/api/main.go`

```go
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

    auditadapter "todoe/internal/audit/adapter"
    taskadapter "todoe/domain/task/adapter"
    taskhttp "todoe/domain/task/adapter/http"
    taskapplication "todoe/domain/task/application"
    taskdomain "todoe/domain/task/domain"
    "todoe/internal/event"
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

    bus := event.NewEventBus()

    auditRepo := auditadapter.NewMongoRepository(clientIO)
    auditHandler := auditadapter.NewAuditHandler(auditRepo)
    bus.Subscribe(taskdomain.EventCreated, auditHandler)
    bus.Subscribe(taskdomain.EventStatusChanged, auditHandler)

    taskRepo := taskadapter.NewMongoRepository(clientIO)
    projectionHandler := taskadapter.NewProjectionHandler(taskRepo)
    bus.Subscribe(taskdomain.EventCreated, projectionHandler)
    bus.Subscribe(taskdomain.EventStatusChanged, projectionHandler)

    taskService := taskapplication.NewService(taskRepo, bus)
    taskHandler := taskhttp.NewHandler(taskService)

    app := fiber.New()
    app.Get("/health", healthHandler.CheckHealth)
    app.Post("/tasks", taskHandler.Create)
    app.Get("/tasks", taskHandler.List)
    app.Get("/tasks/:id", taskHandler.Detail)
    app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)

    log.Fatal(app.Listen(":3000"))
}
```

## Key rules

- `projectionHandler` subscribes to both events — any task mutation updates `tasks_view`
- Subscription order matters: audit handler runs first, projection handler second (bus dispatches in registration order)
- `taskRepo` is passed to both `NewProjectionHandler` and `NewService` — it is the same repository instance

## Check it compiles

```bash
go build ./...
```

## Full flow

```
POST /tasks
  → service.CreateTask
      → repo.Append     → task_events (write model)
      → bus.Publish(EventCreated)
          → auditHandler   → audit_log
          → projectionHandler → repo.Upsert → tasks_view (read model)

GET /tasks
  → service.ListTasks
      → repo.FindAll    → tasks_view (simple Find, no aggregation)

GET /tasks/:id
  → service.GetTask
      → repo.FindByID   → tasks_view (simple FindOne)

PATCH /tasks/:id/status
  → service.ChangeStatus
      → repo.FindByID   → tasks_view (read current)
      → repo.Append     → task_events (write)
      → bus.Publish(EventStatusChanged)
          → auditHandler   → audit_log
          → projectionHandler → repo.Upsert → tasks_view (update read model)
```
