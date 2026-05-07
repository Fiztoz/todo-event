package adapter

import (
	"context"
	"log/slog"

	taskdomain "todoe/domain/task/domain"
	"todoe/internal/event"
	sladomain "todoe/internal/sla/domain"
)

func NewSlaHandler(repo *MongoRepository, csvPath string) func(context.Context, event.Event) error {
	csv, err := NewCSVWriter(csvPath)
	if err != nil {
		slog.Error("sla: failed to create csv writer", "err", err)
		return func(context.Context, event.Event) error { return nil }
	}

	return func(ctx context.Context, e event.Event) error {
		task, ok := e.Payload.(taskdomain.Task)
		if !ok {
			return nil
		}

		// Only compute SLA when the task reaches "done"
		if task.Status != taskdomain.StatusDone {
			return nil
		}

		if task.OriginID == nil {
			slog.Warn("sla: done task has no origin_id", "task_id", task.ID)
			return nil
		}

		pendingCreatedAt, found := repo.GetPendingCreatedAt(ctx, *task.OriginID)
		if !found {
			slog.Warn("sla: pending task not found", "origin_id", task.OriginID)
			return nil
		}

		duration := task.CreatedAt.Sub(pendingCreatedAt)

		slog.Info("sla: recorded",
			"task_id", task.ID,
			"origin_id", task.OriginID,
			"pending_created_at", pendingCreatedAt,
			"done_created_at", task.CreatedAt,
			"duration_ms", duration.Milliseconds())

		entry := sladomain.SLAEntry{
			ID:        task.OriginID.Hex(),
			EventType: e.Type,
			CreatedAt: pendingCreatedAt,
			Duration:  duration,
		}
		csv.Write(entry)
		return nil
	}
}
