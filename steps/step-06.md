# Step 06 — Add the detail endpoint: HTTP handler and route

## What changes and why

With event sourcing in place, each task has a stable ID that clients can bookmark and
query directly. `GET /tasks/:id` returns the current state of a single task,
reconstructed by replaying its event stream from `task_events`.

Two files change:

1. The HTTP handler gets a `Detail` method.
2. `main.go` registers the new route.

The `UseCase` port already has `GetTask` from Step 03, so no interface change is
needed here.

## File: `domain/task/adapter/http/handler.go`

Add `Detail` between `List` and `ChangeStatus`:

```go
func (h *Handler) Detail(c *fiber.Ctx) error {
	id, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	result := h.useCase.GetTask(c.Context(), id)
	if result.IsError() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "task not found"})
	}
	return c.JSON(result.MustGet())
}
```

Error handling note: any error from `GetTask` is mapped to 404. This is accurate
because `domain.Apply` returns an error only when the event slice is empty, which
means the aggregate was never created (i.e. not found). Infrastructure errors would
surface as 500 in a more complete implementation by inspecting error types.

## File: `cmd/api/main.go`

Add one route line after `app.Get("/tasks", ...)`:

```go
// Before
app.Post("/tasks", taskHandler.Create)
app.Get("/tasks", taskHandler.List)
app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)

// After
app.Post("/tasks", taskHandler.Create)
app.Get("/tasks", taskHandler.List)
app.Get("/tasks/:id", taskHandler.Detail)           // add
app.Patch("/tasks/:id/status", taskHandler.ChangeStatus)
```

Route ordering matters in Fiber: the exact `/tasks` listing route must come before
the parameterised `/tasks/:id`. `PATCH /tasks/:id/status` is unaffected because it
uses a different HTTP method.

## Smoke test

```bash
# 1. Create a task — note the returned id
curl -s -X POST http://localhost:3000/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"buy milk"}' | jq .

# 2. Fetch it by id
curl -s http://localhost:3000/tasks/<id> | jq .

# 3. Change its status
curl -s -X PATCH http://localhost:3000/tasks/<id>/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"in_progress"}' | jq .

# 4. Fetch again — status should be in_progress
curl -s http://localhost:3000/tasks/<id> | jq .
```

## Checklist

- [ ] `Detail` method added to `Handler`
- [ ] `GET /tasks/:id` registered in `main.go`, placed between List and ChangeStatus
- [ ] `go build ./...` passes
- [ ] Smoke test: create → fetch by id → change status → fetch again returns updated status
