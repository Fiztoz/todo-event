package port

import (
	"context"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/captcha/domain"
	"todoe/internal/event"
)

type UseCase interface {
	Issue(ctx context.Context) mo.Result[domain.Challenge]
	Verify(ctx context.Context, id bson.ObjectID, answer int) mo.Result[domain.Challenge]
}

type Repository interface {
	Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}]
	SaveChallenge(ctx context.Context, c domain.Challenge) mo.Result[struct{}]
	FindChallenge(ctx context.Context, id bson.ObjectID) mo.Result[domain.Challenge]
	MarkVerified(ctx context.Context, id bson.ObjectID) mo.Result[struct{}]
}

type Publisher interface {
	Publish(ctx context.Context, e event.Event)
}
