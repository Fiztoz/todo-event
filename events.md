# Event Sourcing

The repository **never updates or deletes** a record. Every state change is a new record.

## Traditional vs Event Sourcing

**Traditional approach:**
```
tasks collection:
{ _id: "abc", title: "buy milk", status: "pending" }

// Update → overwrites in place
{ _id: "abc", title: "buy milk", status: "done" }   ← history gone
```

**Event sourcing approach:**
```
tasks collection:
{ _id: "abc", title: "buy milk", status: "pending", created_at: "..." }
{ _id: "xyz", origin_id: "abc", status: "done",     created_at: "..." }  ← new record

// "abc" is never touched again
```

The second record *associates* with the origin via `origin_id`. The current state of a task is derived by reading the latest record for a given origin.

## Why

- Full audit trail — you can see every state the entity ever had
- Nothing is ever lost — rollback = ignore the latest record
- Aligns with the domain events pattern — each event that changes state produces a new document

## In This Codebase

Right now the `tasks` collection only has `Save` (insert). When we add "complete a task" or "delete a task", the `Repository` interface will have no `Update` or `Delete` — only another `Save` with a new record referencing the original `_id`.

```go
// What we will NOT add:
UpdateStatus(ctx, id, status) // ← forbidden

// What we WILL add:
SaveStatusChange(ctx, event TaskStatusChangedEvent) // ← new record
```

The current `Task` struct will likely gain an `OriginID *bson.ObjectID` field (nil for the first record, set for all subsequent ones).
