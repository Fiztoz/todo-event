package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	auditadapter "todoe/internal/audit/adapter"
	auditdomain "todoe/internal/audit/domain"
	"todoe/internal/messaging"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:root@localhost:27017"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	clientIO := mo.NewIOEither(func() (*mongo.Client, error) {
		return mongo.Connect(options.Client().ApplyURI(mongoURI))
	})
	auditRepo := auditadapter.NewMongoRepository(clientIO)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("nats:", err)
	}
	defer nc.Drain()

	nc.Subscribe(messaging.TaskSubject, func(m *nats.Msg) {
		var msg messaging.Message
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			slog.Error("audit: unmarshal", "err", err)
			return
		}
		var payload any
		json.Unmarshal(msg.Payload, &payload)
		entry := auditdomain.AuditEntry{
			ID:        bson.NewObjectID(),
			EventType: msg.Type,
			Payload:   payload,
			CreatedAt: time.Now(),
		}
		slog.Info("audit: received event", "type", msg.Type)
		if r := auditRepo.Save(context.Background(), entry); r.IsError() {
			slog.Error("audit: save", "err", r.Error())
		}
	})

	slog.Info("audit service listening", "subject", messaging.TaskSubject)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("audit service stopping")
}
