package event

import (
	"context"
	"log/slog"
	"sync"
)

type Event struct {
	Type    EventType
	Payload any
}

type Publisher interface {
	Publish(ctx context.Context, e Event)
}

type EventHandler func(context.Context, Event) error

type EventType string
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[EventType][]EventHandler)}
}

func (b *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	slog.Debug("eventbus: handler subscribed", "event_type", eventType)
}

func (b *EventBus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	handlers := b.handlers[e.Type]

	for _, handler := range handlers {
		if err := handler(ctx, e); err != nil {
			slog.Error("eventbus: handler error", "event_type", e.Type, "err", err)
		}
		slog.Debug("eventbus: event published", "event_type", e.Type)
	}
}

var _ Publisher = (*EventBus)(nil)