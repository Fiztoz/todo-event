package event

import (
	"context"
	"log/slog"
	"sync"
)

// Event represents a domain event with a type and payload.
type Event struct {
	Type    string
	Payload any
}

// DomainEvent is the type key for event handlers.
type DomainEvent string

// HandlerFunc is the signature for event handlers.
type HandlerFunc func(context.Context, Event) error

// Publisher is the interface for publishing events.
type Publisher interface {
	Publish(ctx context.Context, e Event)
}

// EventBus dispatches events to registered handlers.
type EventBus struct {
	mu         sync.RWMutex
	handlerMap map[DomainEvent][]HandlerFunc
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		handlerMap: make(map[DomainEvent][]HandlerFunc),
	}
}

// Subscribe registers a handler for the given event type.
func (b *EventBus) Subscribe(eventType DomainEvent, fn HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlerMap[eventType] = append(b.handlerMap[eventType], fn)
}

// Publish dispatches an event to all registered handlers.
func (b *EventBus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	handlers := b.handlerMap[DomainEvent(e.Type)]
	b.mu.RUnlock()

	for _, fn := range handlers {
		if err := fn(ctx, e); err != nil {
			slog.Error("eventbus: handler error", "event_type", e.Type, "err", err)
		}
	}
}

var _ Publisher = (*EventBus)(nil)
