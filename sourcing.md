# Event Sourcing in this Application

## What gets stored

Every change appends a record to `task_events`. Nothing is ever updated or deleted.

```
// Create "buy milk"
{ _id: A1, aggregate_id: T1, type: "task.created",        payload: {title: "buy milk"}, created_at: ... }

// Change status to done
{ _id: A2, aggregate_id: T1, type: "task.status_changed", payload: {status: "done"},    created_at: ... }
```

`T1` is the stable task identity. `A1`, `A2` are just insertion IDs.

---

## Reading current state

There is no "current state" document. To read a task, the repository fetches all events sharing the same `aggregate_id`, sorted by `_id`, and folds them through `domain.Apply`:

```
events → Apply → Task{id: T1, title: "buy milk", status: "done"}
```

`Apply` starts with an empty `Task` and handles each event type in sequence — `EventCreated` sets title and timestamps, `EventStatusChanged` overwrites status.

---

## The full flow for `ChangeStatus`

```
1. Service calls repo.FindByID(T1)
      → MongoDB: find all events where aggregate_id = T1, sort by _id
      → domain.Apply replays them → Task{status: "pending"}

2. Service calls task.ChangeStatus("done")
      → Pure Logic: returns Task{..., status: "done"} (in memory only)

3. Service calls repo.Append(T1, "task.status_changed", {status: "done"})
      → MongoDB: insert new event document

4. Service calls publisher.Publish(EventStatusChanged, task)
      → AuditHandler writes to audit_log
```

---

## Why this is different from snapshot storage

| | Snapshot storage | Event sourcing |
|---|---|---|
| Stored document | Full Task snapshot | Minimal delta (`{title}` or `{status}`) |
| Reading current state | Query the leaf node | Replay all events |
| History | Linked list via `OriginID` | All records share `aggregate_id` |
| Source of truth | The latest record | The entire event stream |

The event stream is permanent and complete. You can replay it to any point in time, add new event types, or build new projections without touching the store.
