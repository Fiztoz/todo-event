package adapter

import (
	"context"
	"fmt"

	taskdomain "todoe/domain/task/domain"
	"todoe/internal/event"
)

func NewReaderHandler(repo *MongoRepository) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		fmt.Printf("Received event: %v\n", e)
		task, ok := e.Payload.(taskdomain.Task)
		if !ok {
			return fmt.Errorf("invalid payload type: expected taskdomain.Task, got %T", e.Payload)
		}

		result := repo.Upsert(ctx, task)
		if result.IsError() {
			return result.Error()
		}
		return nil
	}
}
