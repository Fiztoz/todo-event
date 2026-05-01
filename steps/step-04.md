# Step 04 — Rewrite `MongoRepository` to store and replay events

## What changes and why

This is the infrastructure adapter change. The repository switches from the `tasks`
collection (one document = one snapshot) to the `task_events` collection (one document
= one event). Three methods change:

| Method | Before | After |
|---|---|---|
| `collection()` | points to `tasks` | points to `task_events` |
| `Save` → `Append` | inserts full Task doc | marshals payload + inserts `StoredEvent` |
| `FindByID` | `FindOne` by `_id` | `Find` all events by `aggregate_id`, replay with `Apply` |
| `FindAll` | `$lookup + $match` to find leaf nodes | `$sort + $group` to bucket events per aggregate, replay each |

The struct scaffolding (`MongoRepository`, `NewMongoRepository`, `getClient`,
`collection`) is identical to before.

## File: `domain/task/adapter/mongo_repository.go`

### Imports — add `"time"` and `mongo/options`

```go
import (
	"context"
	"sync"
	"sync/atomic"
	"time"                                           // add

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"  // add

	"todoe/domain/task/domain"
)
```

### `collection()` — change collection name

```go
// Before
return either.MustRight().Database("todoe").Collection("tasks"), nil

// After
return either.MustRight().Database("todoe").Collection("task_events"), nil
```

### Remove `Save`, add `Append`

```go
// Remove:
func (r *MongoRepository) Save(ctx context.Context, task domain.Task) mo.Result[struct{}] { ... }

// Add:
func (r *MongoRepository) Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}] {
	col, err := r.collection()
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
	if _, err := col.InsertOne(ctx, e); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}
```

### Rewrite `FindByID`

Old `FindByID` did `FindOne` by `_id`. New `FindByID` does `Find` by `aggregate_id`
because many event documents share one logical aggregate ID.

```go
func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	cursor, err := col.Find(ctx,
		bson.D{{Key: "aggregate_id", Value: id}},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}),
	)
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	defer cursor.Close(ctx)
	var events []domain.StoredEvent
	if err := cursor.All(ctx, &events); err != nil {
		return mo.Err[domain.Task](err)
	}
	task, err := domain.Apply(id, events)
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	return mo.Ok(task)
}
```

### Rewrite `FindAll`

Old `FindAll` used `$lookup + $match` to identify leaf nodes (documents with no
children referencing them via `origin_id`). New `FindAll` groups every event by
`aggregate_id` then replays each group.

The `$sort` before `$group` guarantees that `$push` appends events in insertion order
within each bucket — critical for correct replay.

```go
func (r *MongoRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$aggregate_id"},
			{Key: "events", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}},
	}
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	defer cursor.Close(ctx)

	var tasks []domain.Task
	for cursor.Next(ctx) {
		var group struct {
			AggregateID bson.ObjectID        `bson:"_id"`
			Events      []domain.StoredEvent `bson:"events"`
		}
		if err := cursor.Decode(&group); err != nil {
			return mo.Err[[]domain.Task](err)
		}
		task, err := domain.Apply(group.AggregateID, group.Events)
		if err != nil {
			return mo.Err[[]domain.Task](err)
		}
		tasks = append(tasks, task)
	}
	return mo.Ok(tasks)
}
```

## Checklist

- [ ] `"time"` and `"mongo/options"` added to imports
- [ ] `collection()` returns `"task_events"`
- [ ] `Save` removed; `Append` added
- [ ] `FindByID` queries by `aggregate_id`, sorts by `_id` asc, calls `Apply`
- [ ] `FindAll` uses `$sort + $group` pipeline, calls `Apply` per group
- [ ] `go build ./...` passes (requires Step 03 already applied)
