# CQRS Step 3 — Detail Endpoint + Error Differentiation

## What changes and why

Add `GET /tasks/:id` to the HTTP surface. Also differentiate validation errors (400) from infrastructure failures (500) in `Create` and `ChangeStatus` — previously all errors returned 400 indiscriminately.

## Target: `domain/task/adapter/http/handler.go`

```go
package httpadapter

import (
    "errors"

    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/v2/bson"

    "todoe/domain/task/application"
    "todoe/domain/task/domain"
    "todoe/domain/task/port"
)

type Handler struct {
    useCase port.UseCase
}

func NewHandler(useCase port.UseCase) *Handler {
    return &Handler{useCase: useCase}
}

func (h *Handler) Create(c *fiber.Ctx) error {
    var body struct {
        Title string `json:"title"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }
    result := h.useCase.CreateTask(c.Context(), body.Title)
    if result.IsError() {
        if errors.Is(result.Error(), application.ErrInvalidTitle) {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": result.Error().Error()})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
    }
    return c.Status(fiber.StatusCreated).JSON(result.MustGet())
}

func (h *Handler) List(c *fiber.Ctx) error {
    result := h.useCase.ListTasks(c.Context())
    if result.IsError() {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
    }
    tasks := result.OrElse([]domain.Task{})
    return c.JSON(tasks)
}

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

func (h *Handler) ChangeStatus(c *fiber.Ctx) error {
    id, err := bson.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
    }
    var body struct {
        Status domain.Status `json:"status"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }
    result := h.useCase.ChangeStatus(c.Context(), id, body.Status)
    if result.IsError() {
        if errors.Is(result.Error(), application.ErrInvalidStatus) {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": result.Error().Error()})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
    }
    return c.JSON(result.MustGet())
}
```

## Key rules

- `Detail` returns 404 on any repository error — the caller asked for a specific ID and got nothing
- `Create` / `ChangeStatus` differentiate: validation errors → 400, anything else → 500
- Handler imports `application` only for the sentinel errors (`ErrInvalidTitle`, `ErrInvalidStatus`) — never for the `Service` type

## Next step

Add `Upsert` to the repository and switch reads to `tasks_view` in `domain/task/adapter/mongo_repository.go`.
