package adapter

import (
	"context"
	"fmt"

	"todoe/internal/captcha/domain"
	"todoe/internal/event"
)

func NewProjectionHandler(repo *MongoRepository) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		challenge, ok := e.Payload.(domain.Challenge)
		if !ok {
			return fmt.Errorf("unexpected payload type %T", e.Payload)
		}
		switch e.Type {
		case domain.EventIssued:
			result := repo.SaveChallenge(ctx, challenge)
			if result.IsError() {
				return result.Error()
			}
		case domain.EventVerified:
			result := repo.MarkVerified(ctx, challenge.ID)
			if result.IsError() {
				return result.Error()
			}
		}
		return nil
	}
}
