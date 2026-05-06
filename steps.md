
//Port
type Publisher interface {
	Publish(ctx context.Context, e event.Event)
}

//Bus
import (
	"context"
	"log/slog"
	"sync"
)

type Event struct {
	Type    string
	Payload any
}

type Publisher interface {
	Publish(ctx context.Context, e Event)
}

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]func(context.Context, Event) error
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[string][]func(context.Context, Event) error)}
}

func (b *EventBus) Subscribe(eventType string, fn func(context.Context, Event) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], fn)
	slog.Debug("eventbus: handler subscribed", "event_type", eventType)
}

func (b *EventBus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	handlers := b.handlers[e.Type]
	b.mu.RUnlock()

	for _, fn := range handlers {
		if err := fn(ctx, e); err != nil {
			slog.Error("eventbus: handler error", "event_type", e.Type, "err", err)
		}
		slog.Debug("eventbus: event published", "event_type", e.Type)
	}
}

var _ Publisher = (*EventBus)(nil)


//Audit Adapter
import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/internal/audit/domain"
	"todoe/internal/event"
)

func NewAuditHandler(repo *MongoRepository) func(context.Context, event.Event) error {
	return func(ctx context.Context, e event.Event) error {
		entry := domain.AuditEntry{
			ID:        bson.NewObjectID(),
			EventType: e.Type,
			Payload:   e.Payload,
			CreatedAt: time.Now(),
		}
		slog.Info("audit: handling event", "event_type", e.Type)
		result := repo.Save(ctx, entry)
		if result.IsError() {
			return result.Error()
		}
		return nil
	}
}

//Audit Mongo Adapter

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"todoe/internal/audit/domain"
)

type MongoRepository struct {
	clientIO    mo.IOEither[*mongo.Client]
	once        sync.Once
	cached      mo.Either[error, *mongo.Client]
	initialized atomic.Bool
}

func NewMongoRepository(clientIO mo.IOEither[*mongo.Client]) *MongoRepository {
	return &MongoRepository{clientIO: clientIO}
}

func (r *MongoRepository) getClient() mo.Either[error, *mongo.Client] {
	r.once.Do(func() {
		r.cached = r.clientIO.Run()
		r.initialized.Store(true)
	})
	return r.cached
}

func (r *MongoRepository) collection() (*mongo.Collection, error) {
	either := r.getClient()
	if either.IsLeft() {
		return nil, either.MustLeft()
	}
	return either.MustRight().Database("todoe").Collection("audit_log"), nil
}

func (r *MongoRepository) Save(ctx context.Context, entry domain.AuditEntry) mo.Result[struct{}] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[struct{}](err)
	}
	if _, err := col.InsertOne(ctx, entry); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

//Audit Entity
import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditEntry struct {
	ID        bson.ObjectID `bson:"_id"`
	EventType string        `bson:"event_type"`
	Payload   any           `bson:"payload"`
	CreatedAt time.Time     `bson:"created_at"`
}