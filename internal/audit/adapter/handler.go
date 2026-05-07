package adapter

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/audit/domain"
	"todoe/internal/event"
)

const (
	maxRetries    = 3
	retryInterval = 1 * time.Second
)

func NewAuditHandler(repo *MongoRepository, fallbackPath string) func(context.Context, event.Event) error {
	fileWriter := NewFileWriter(fallbackPath)

	return func(ctx context.Context, e event.Event) error {
		entry := domain.AuditEntry{
			ID:        bson.NewObjectID(),
			EventType: string(e.Type),
			Payload:   e.Payload,
			CreatedAt: time.Now(),
		}
		slog.Info("audit: handling event", "event_type", e.Type)

		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			result := repo.Save(ctx, entry)
			if !result.IsError() {
				return nil
			}
			lastErr = result.Error()
			slog.Warn("audit: mongo save failed",
				"attempt", attempt,
				"max_retries", maxRetries,
				"err", lastErr,
			)
			if attempt < maxRetries {
				time.Sleep(retryInterval)
			}
		}

		slog.Error("audit: all mongo retries exhausted, falling back to file",
			"file", fallbackPath,
			"err", lastErr,
		)
		if err := fileWriter.Write(entry); err != nil {
			slog.Error("audit: file fallback also failed", "err", err)
			return err
		}
		return nil
	}
}
