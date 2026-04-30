package adapter

import (
	"context"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
	"todoe/internal/event"
)

func NewSaveHandler(repo port.Repository) event.HandlerFunc {
	return func(ctx context.Context, e event.Event) error {
		task, ok := e.Payload.(domain.Task)
		if !ok {
			return nil
		}
		return repo.Save(ctx, task).Error()
	}
}
