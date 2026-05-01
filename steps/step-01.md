# Step 01 — Strip `OriginID` and introduce the event vocabulary

## What changes and why

The append-only chain model stored each state transition as a brand-new document
linked backwards via `OriginID`. In event sourcing that pointer disappears: an
aggregate's identity is stable across its entire lifetime, so `Task.ID` never changes.
The event log is the history — `OriginID` is no longer needed.

What the domain does need are names and payload structs for each kind of change:

- `StoredEvent` — the record persisted to MongoDB, holding serialised `bson.Raw` bytes
- `TaskCreatedPayload` — the data captured when a task is first created
- `StatusChangedPayload` — the data captured when the status transitions

`ChangeStatus` gets one small but important fix: because the ID is now stable, the
returned task must carry the original `ID` and `CreatedAt` forward instead of stamping
a fresh `ObjectID`.

No behaviour visible to callers changes yet — `NewTask` and `ChangeStatus` keep working
exactly as before from the service's perspective.

## File: `domain/task/domain/task.go`

### Task struct — remove `OriginID`

```go
// Before
type Task struct {
	ID        bson.ObjectID  `bson:"_id"                 json:"id"`
	OriginID  *bson.ObjectID `bson:"origin_id,omitempty" json:"origin_id,omitempty"`
	Title     string         `bson:"title"               json:"title"`
	Status    Status         `bson:"status"              json:"status"`
	CreatedAt time.Time      `bson:"created_at"          json:"created_at"`
}

// After
type Task struct {
	ID        bson.ObjectID `bson:"_id"        json:"id"`
	Title     string        `bson:"title"      json:"title"`
	Status    Status        `bson:"status"     json:"status"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}
```

### Add event storage types

```go
type StoredEvent struct {
	ID          bson.ObjectID `bson:"_id"`
	AggregateID bson.ObjectID `bson:"aggregate_id"`
	Type        string        `bson:"type"`
	Payload     bson.Raw      `bson:"payload"`
	CreatedAt   time.Time     `bson:"created_at"`
}

type TaskCreatedPayload struct {
	Title string `bson:"title"`
}

type StatusChangedPayload struct {
	Status Status `bson:"status"`
}
```

### `ChangeStatus` — keep the same ID

```go
// Before
func (t Task) ChangeStatus(status Status) Task {
	return Task{
		ID:        bson.NewObjectID(),  // new ID each time
		OriginID:  &t.ID,
		Title:     t.Title,
		Status:    status,
		CreatedAt: time.Now(),
	}
}

// After
func (t Task) ChangeStatus(status Status) Task {
	return Task{ID: t.ID, Title: t.Title, Status: status, CreatedAt: t.CreatedAt}
}
```

`NewTask` is unchanged.

## Checklist

- [ ] `OriginID` field removed from `Task`
- [ ] `StoredEvent`, `TaskCreatedPayload`, `StatusChangedPayload` added
- [ ] `ChangeStatus` propagates original `ID` and `CreatedAt`
- [ ] `go build ./...` passes

> `Apply` is added in Step 02. Until then `StoredEvent` is declared but unused —
> the compiler is satisfied because exported types are never "unused".
