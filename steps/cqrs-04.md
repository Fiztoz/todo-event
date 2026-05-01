# CQRS Step 4 — Repository: Upsert + Switch Reads to tasks_view

## What changes and why

The read side needs a materialized view. Two things happen here:

1. Add `Upsert` — writes current task state to `tasks_view` (called by the projection handler)
2. Switch `FindAll` and `FindByID` to read from `tasks_view` instead of aggregating `task_events`

After this step, reads are a simple `Find`/`FindOne` — no aggregation pipeline, no event replay.

The `db()` helper replaces the per-method `collection()` call so both collections can be reached from one place.

## Target: `domain/task/adapter/mongo_repository.go`

```go
package adapter

import (
    "context"
    "sync"
    "sync/atomic"
    "time"

    "github.com/samber/mo"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"

    "todoe/domain/task/domain"
)

type MongoRepository struct {
    clientIO    mo.IOEither[*mongo.Client]
    once        sync.Once
    cached      mo.Either[error, *mongo.Client]
    initialized atomic.Bool
}

func NewMongoRepository(clientIO mo.IOEither[*mongo.Client]) *MongoRepository {
    return &MongoRepository{clientIO: clientIO}
}

func (r *MongoRepository) getClient() mo.Either[error, *mongo.Client] {
    r.once.Do(func() {
        r.cached = r.clientIO.Run()
        r.initialized.Store(true)
    })
    return r.cached
}

func (r *MongoRepository) db() (*mongo.Database, error) {
    either := r.getClient()
    if either.IsLeft() {
        return nil, either.MustLeft()
    }
    return either.MustRight().Database("todoe"), nil
}

// Append writes a domain event to the event store (task_events).
func (r *MongoRepository) Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}] {
    db, err := r.db()
    if err != nil {
        return mo.Err[struct{}](err)
    }
    raw, err := bson.Marshal(payload)
    if err != nil {
        return mo.Err[struct{}](err)
    }
    e := domain.StoredEvent{
        ID:          bson.NewObjectID(),
        AggregateID: aggregateID,
        Type:        eventType,
        Payload:     raw,
        CreatedAt:   time.Now(),
    }
    if _, err := db.Collection("task_events").InsertOne(ctx, e); err != nil {
        return mo.Err[struct{}](err)
    }
    return mo.Ok(struct{}{})
}

// Upsert writes the current task state to the read model (tasks_view).
func (r *MongoRepository) Upsert(ctx context.Context, task domain.Task) mo.Result[struct{}] {
    db, err := r.db()
    if err != nil {
        return mo.Err[struct{}](err)
    }
    filter := bson.D{{Key: "_id", Value: task.ID}}
    update := bson.D{{Key: "$set", Value: task}}
    if _, err := db.Collection("tasks_view").UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
        return mo.Err[struct{}](err)
    }
    return mo.Ok(struct{}{})
}

// FindByID reads current task state from the read model (tasks_view).
func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
    db, err := r.db()
    if err != nil {
        return mo.Err[domain.Task](err)
    }
    var task domain.Task
    if err := db.Collection("tasks_view").FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&task); err != nil {
        return mo.Err[domain.Task](err)
    }
    return mo.Ok(task)
}

// FindAll reads all current task states from the read model (tasks_view).
func (r *MongoRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
    db, err := r.db()
    if err != nil {
        return mo.Err[[]domain.Task](err)
    }
    cursor, err := db.Collection("tasks_view").Find(ctx, bson.D{})
    if err != nil {
        return mo.Err[[]domain.Task](err)
    }
    defer cursor.Close(ctx)
    tasks := []domain.Task{}
    if err := cursor.All(ctx, &tasks); err != nil {
        return mo.Err[[]domain.Task](err)
    }
    return mo.Ok(tasks)
}
```

## Key rules

- `Append` → `task_events` (write model, append-only)
- `Upsert` → `tasks_view` (read model, current state)
- `FindAll` / `FindByID` → `tasks_view` (simple queries, no aggregation)
- `Upsert` is NOT in the `port.Repository` interface — only `MongoRepository` knows about it; the projection handler receives `*MongoRepository` directly

## Next step

Create `domain/task/adapter/projection_handler.go` to keep `tasks_view` up to date.
