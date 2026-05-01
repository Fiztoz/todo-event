# CQRS Step 5 — Projection Handler

## What changes and why

The `ProjectionHandler` is the bridge between the write model and the read model. It subscribes to domain events and upserts the denormalized task state into `tasks_view`. Every time a task is created or its status changes, `tasks_view` is updated synchronously (within the same event dispatch).

## Target: `domain/task/adapter/projection_handler.go`

```go
package adapter

import (
    "context"
    "fmt"

    "todoe/domain/task/domain"
    "todoe/internal/event"
)

func NewProjectionHandler(repo *MongoRepository) func(context.Context, event.Event) error {
    return func(ctx context.Context, e event.Event) error {
        task, ok := e.Payload.(domain.Task)
        if !ok {
            return fmt.Errorf("unexpected payload type %T", e.Payload)
        }
        result := repo.Upsert(ctx, task)
        if result.IsError() {
            return result.Error()
        }
        return nil
    }
}
```

## Key rules

- Takes `*MongoRepository` (concrete type) — `Upsert` is not on the `port.Repository` interface
- Returns an error if the payload type is wrong — unlike `SaveHandler` which silently returns nil on type mismatch, the projection handler is stricter: an unexpected payload is a programming error
- Subscribe to BOTH `EventCreated` and `EventStatusChanged` in main.go
- Handler errors are logged by the bus but do not fail the HTTP request

## Next step

Wire the projection handler and the `GET /tasks/:id` route in `cmd/api/main.go`.
