# Step 05 — Update the application service to write events

## What changes and why

The service is where business logic lives. It no longer calls `repo.Save(task)`;
instead it calls `repo.Append(aggregateID, eventType, payload)` — writing only the
minimal delta for each operation.

The aggregate ID is now allocated by the service (not inside `domain.NewTask`) so the
service controls what ID it returns to the caller and publishes to the event bus before
any round-trip to the database.

Three methods change:

**`CreateTask`** — allocates `id` with `bson.NewObjectID()`, appends a
`TaskCreatedPayload`, then constructs the in-memory `Task` locally. `domain.NewTask`
is not called here because the service needs to control the ID and timestamp for the
event payload.

**`GetTask`** (new) — one-line delegation to `repo.FindByID`. This satisfies the
`UseCase.GetTask` method added to the port in Step 03.

**`ChangeStatus`** — replays the current state via `repo.FindByID`, calls the pure
`task.ChangeStatus` to compute the next state in memory, then appends only a
`StatusChangedPayload`. The `id` argument (the stable aggregate ID from the caller) is
what gets recorded — not `next.ID`, which is the same value after Step 01 but naming
it explicitly avoids confusion.

## File: `domain/task/application/service.go`

### Imports — add `"time"`

```go
import (
	"context"
	"errors"
	"time"         // add

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
	"todoe/internal/event"
)
```

### Replace `CreateTask`

```go
// Before
func (s *Service) CreateTask(ctx context.Context, title string) mo.Result[domain.Task] {
	if err := validateTitle(title); err != nil {
		return mo.Err[domain.Task](err)
	}
	task := domain.NewTask(title)
	if result := s.repo.Save(ctx, task); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	s.publisher.Publish(ctx, event.Event{Type: domain.EventCreated, Payload: task})
	return mo.Ok(task)
}

// After
func (s *Service) CreateTask(ctx context.Context, title string) mo.Result[domain.Task] {
	if err := validateTitle(title); err != nil {
		return mo.Err[domain.Task](err)
	}
	id := bson.NewObjectID()
	if result := s.repo.Append(ctx, id, domain.EventCreated, domain.TaskCreatedPayload{Title: title}); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	task := domain.Task{ID: id, Title: title, Status: domain.StatusPending, CreatedAt: time.Now()}
	s.publisher.Publish(ctx, event.Event{Type: domain.EventCreated, Payload: task})
	return mo.Ok(task)
}
```

### Add `GetTask`

```go
func (s *Service) GetTask(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	return s.repo.FindByID(ctx, id)
}
```

### Replace `ChangeStatus`

```go
// Before
func (s *Service) ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task] {
	if err := validateStatus(status); err != nil {
		return mo.Err[domain.Task](err)
	}
	current := s.repo.FindByID(ctx, id)
	if current.IsError() {
		return mo.Err[domain.Task](current.Error())
	}
	next := current.MustGet().ChangeStatus(status)
	if result := s.repo.Save(ctx, next); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	s.publisher.Publish(ctx, event.Event{Type: domain.EventStatusChanged, Payload: next})
	return mo.Ok(next)
}

// After
func (s *Service) ChangeStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[domain.Task] {
	if err := validateStatus(status); err != nil {
		return mo.Err[domain.Task](err)
	}
	current := s.repo.FindByID(ctx, id)
	if current.IsError() {
		return mo.Err[domain.Task](current.Error())
	}
	next := current.MustGet().ChangeStatus(status)
	if result := s.repo.Append(ctx, id, domain.EventStatusChanged, domain.StatusChangedPayload{Status: status}); result.IsError() {
		return mo.Err[domain.Task](result.Error())
	}
	s.publisher.Publish(ctx, event.Event{Type: domain.EventStatusChanged, Payload: next})
	return mo.Ok(next)
}
```

## Checklist

- [ ] `"time"` added to imports
- [ ] `CreateTask` uses `bson.NewObjectID()` + `repo.Append` + inline `Task` literal
- [ ] `GetTask` added (one line)
- [ ] `ChangeStatus` uses `repo.Append` with `StatusChangedPayload`
- [ ] `go build ./...` passes (requires Steps 03 and 04 already applied)
