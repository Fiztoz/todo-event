package adapter

import (
	"context"
	"fmt"

	"todoe/domain/task/domain"
	"todoe/internal/event"
)

type SaveHandler struct {
	repo *MongoRepository
}

func NewSaveHandler(repo *MongoRepository) *SaveHandler {
	return &SaveHandler{repo: repo}
}

func (h *SaveHandler) Handle(ctx context.Context, e event.Event) error {
	task, ok := e.Payload.(domain.Task)
	if e.Type == domain.EventStatusChanged || e.Type == domain.EventCreated {
		if !ok {
			return fmt.Errorf("unexpected payload type %T", e.Payload)
		}
		result := h.repo.Save(ctx, task)
		if result.IsError() {
			return result.Error()
		}
		return nil
	}
	return fmt.Errorf("unsupported event type %s", e.Type)
}
