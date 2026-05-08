package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"todoe/internal/event"
)

func Connect(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return conn, ch, nil
}

func DeclareExchange(ch *amqp.Channel, name string) error {
	return ch.ExchangeDeclare(name, "fanout", true, false, false, false, nil)
}

const (
	QueueAuditTaskEvents    = "audit.task.events"
	QueueAuthenUserEvents   = "authen.user.events"
	QueueAuditUserEvents    = "audit.user.events"
	QueueAPIUserEvents      = "api.user.events"
	QueueRPCUserLookup      = "rpc.user.lookup"
	QueueAPIUserLookupReply = "api.user.lookup.reply"
)

type Binding struct {
	Exchange string
	Queue    string
}

func DeclareTopology(ch *amqp.Channel, bindings []Binding) error {
	for _, b := range bindings {
		if err := DeclareExchange(ch, b.Exchange); err != nil {
			return err
		}
		if _, err := ch.QueueDeclare(b.Queue, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(b.Queue, "", b.Exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

type Publisher struct {
	ch       *amqp.Channel
	exchange string
}

func NewPublisher(ch *amqp.Channel, exchange string) *Publisher {
	return &Publisher{ch: ch, exchange: exchange}
}

func (p *Publisher) Publish(ctx context.Context, e event.Event) {
	payload, _ := json.Marshal(e.Payload)
	data, _ := json.Marshal(Message{Type: e.Type, Payload: payload})
	err := p.ch.PublishWithContext(ctx, p.exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         data,
	})
	if err != nil {
		slog.Error("rabbit: publish error", "exchange", p.exchange, "err", err)
	}
}

var _ event.Publisher = (*Publisher)(nil)

func Subscribe(ch *amqp.Channel, exchange, queue string, handler func(Message)) error {
	if err := DeclareExchange(ch, exchange); err != nil {
		return err
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := ch.QueueBind(q.Name, "", exchange, false, nil); err != nil {
		return err
	}
	deliveries, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			var msg Message
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				slog.Error("rabbit: unmarshal", "queue", queue, "err", err)
				d.Nack(false, false)
				continue
			}
			handler(msg)
			d.Ack(false)
		}
	}()
	return nil
}

func ConsumeQueue(ch *amqp.Channel, queue string, handler func(body []byte)) error {
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	deliveries, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			handler(d.Body)
			d.Ack(false)
		}
	}()
	return nil
}

// RPCServer starts a goroutine that consumes requests from queue, calls handler
// with the raw request body, and publishes the returned bytes to the ReplyTo queue.
func RPCServer(ch *amqp.Channel, queue string, handler func(body []byte) []byte) error {
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	deliveries, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			response := handler(d.Body)
			if err := ch.PublishWithContext(context.Background(), "", d.ReplyTo, false, false, amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: d.CorrelationId,
				Body:          response,
			}); err != nil {
				slog.Error("rpc: reply publish failed", "queue", queue, "err", err)
			}
			d.Ack(false)
		}
	}()
	return nil
}

// PublishRPCRequest sends a request to the given queue and sets the ReplyTo queue.
func PublishRPCRequest(ch *amqp.Channel, queue string, replyTo string, body []byte) error {
	return ch.PublishWithContext(context.Background(), "", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		ReplyTo:     replyTo,
		Body:        body,
	})
}

// RPCCall sends request body to queue and blocks until a correlated reply arrives
// or the timeout elapses. A dedicated exclusive reply queue is created per call.
func RPCCall(ch *amqp.Channel, queue string, body []byte, timeout time.Duration) ([]byte, error) {
	replyQ, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, err
	}
	corrID := fmt.Sprintf("%d", time.Now().UnixNano())
	if err := ch.PublishWithContext(context.Background(), "", queue, false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		ReplyTo:       replyQ.Name,
		Body:          body,
	}); err != nil {
		return nil, err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return nil, errors.New("rpc: reply channel closed")
			}
			if msg.CorrelationId == corrID {
				return msg.Body, nil
			}
		case <-timer.C:
			return nil, errors.New("rpc: timeout waiting for reply")
		}
	}
}

