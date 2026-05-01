# Step 02 — Add the `Apply` projection function

## What changes and why

`Apply` is the heart of event sourcing. Given a stable aggregate ID and an ordered
slice of `StoredEvent` records it folds them left-to-right to produce the current
`Task` state. There is no mutable state, no database, and no side effects — it is a
pure function in the domain layer.

Adding it here, before touching the repository or service, lets you test the replay
logic in isolation before wiring it into any infrastructure.

The two event types handled:

| Event type | Effect |
|---|---|
| `task.created` | Sets `Title`, `Status = pending`, `CreatedAt` |
| `task.status_changed` | Overwrites `Status` |

An empty event slice means the aggregate was never created — return an error.

## File: `domain/task/domain/task.go`

### Add `"fmt"` to the import block

```go
import (
	"fmt"   // add
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)
```

### Add `Apply` after the payload type declarations

```go
func Apply(id bson.ObjectID, events []StoredEvent) (Task, error) {
	if len(events) == 0 {
		return Task{}, fmt.Errorf("task not found: %s", id.Hex())
	}
	task := Task{ID: id}
	for _, e := range events {
		switch e.Type {
		case EventCreated:
			var p TaskCreatedPayload
			bson.Unmarshal(e.Payload, &p)
			task.Title = p.Title
			task.Status = StatusPending
			task.CreatedAt = e.CreatedAt
		case EventStatusChanged:
			var p StatusChangedPayload
			bson.Unmarshal(e.Payload, &p)
			task.Status = p.Status
		}
	}
	return task, nil
}
```

### Why `bson.Unmarshal` errors are silently ignored

`StoredEvent.Payload` is `bson.Raw` — already validated BSON bytes written by this
same codebase. A failure here means the database contains corrupted data, not a
caller mistake. In production you would log and skip; for this stage silent
continuation is acceptable.

## Checklist

- [ ] `"fmt"` added to imports
- [ ] `Apply` added after the payload type declarations
- [ ] `go build ./...` passes
- [ ] Optional: write a table-driven unit test for `Apply` covering:
  - zero events → error
  - one `task.created` event → correct Title/Status/CreatedAt
  - `task.created` + `task.status_changed` → updated Status
