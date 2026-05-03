package messaging

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"

	"todoe/internal/event"
)

// NatsPublisher publishes domain events to a NATS subject as JSON.
type NatsPublisher struct {
	Conn    *nats.Conn
	Subject string
}

func (p *NatsPublisher) Publish(_ context.Context, e event.Event) {
	payload, _ := json.Marshal(e.Payload)
	data, _ := json.Marshal(Message{Type: e.Type, Payload: payload})
	if err := p.Conn.Publish(p.Subject, data); err != nil {
		slog.Error("nats: publish error", "subject", p.Subject, "err", err)
	}
}

// MultiPublisher fans out a single Publish call to multiple publishers.
type MultiPublisher struct {
	Publishers []event.Publisher
}

func (m *MultiPublisher) Publish(ctx context.Context, e event.Event) {
	for _, p := range m.Publishers {
		p.Publish(ctx, e)
	}
}
