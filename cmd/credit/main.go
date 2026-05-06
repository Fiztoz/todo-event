package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	userdomain "todoe/domain/user/domain"
	"todoe/internal/event"
	"todoe/internal/messaging"
)

type multiPublisher struct{ publishers []event.Publisher }

func (m *multiPublisher) Publish(ctx context.Context, e event.Event) {
	for _, p := range m.publishers {
		p.Publish(ctx, e)
	}
}

// knownCorporateDomains has higher base trust.
var knownCorporateDomains = map[string]bool{
	"gmail.com": true,
	"microsoft.com": true,
	"apple.com": true,
	"github.com": true,
	"amazon.com": true,
}

var specialDomains = map[string]bool{
	"pea.co.th": true, // known for disposable emails
}

// disposableDomains are known temporary email providers.
var disposableDomains = map[string]bool{
	"mailinator.com": true,
	"guerrillamail.com": true,
	"tempmail.com": true,
	"throwaway.email": true,
	"yopmail.com": true,
}

// digitPattern matches emails with long numeric suffixes (often auto-generated).
var digitPattern = regexp.MustCompile(`\d{4,}`)

// creditScore computes a deterministic score (300–850) from email characteristics.
// Scores >= 600 are approved.
func creditScore(email string) int {
	score := 500 // base

	local, domain, hasParts := strings.Cut(email, "@")
	if !hasParts || domain == "" {
		return max(300, score-150)
	}

	// Domain reputation
	domain = strings.ToLower(domain)
	if knownCorporateDomains[domain] {
		score += 80
	}
	if disposableDomains[domain] {
		score -= 120
	}
	if specialDomains[domain] {
		score += 500
	}

	// Local part complexity
	if len(local) >= 8 {
		score += 20
	}
	if digitPattern.MatchString(local) {
		score -= 30
	}

	// Dot in local part suggests deliberate choice
	if strings.Contains(local, ".") {
		score += 15
	}

	// Subdomain (e.g. user@dept.corp.com) adds slight trust
	if strings.Count(domain, ".") >= 2 {
		score += 10
	}

	return clamp(300, score, 1000)
}

func clamp(min, v, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, ch, err := messaging.Connect(amqpURL)
	if err != nil {
		log.Fatal("rabbit:", err)
	}
	defer conn.Close()

	if err := messaging.DeclareTopology(ch, []messaging.Binding{
		{Exchange: messaging.CreditResultExchange, Queue: messaging.CreditResultExchange},
		{Exchange: messaging.TaskExchange, Queue: messaging.QueueAuditTaskEvents},
	}); err != nil {
		log.Fatal("rabbit topology:", err)
	}

	publisher := &multiPublisher{publishers: []event.Publisher{
		messaging.NewPublisher(ch, messaging.CreditResultExchange),
		messaging.NewPublisher(ch, messaging.TaskExchange),
	}}

	if err := messaging.Subscribe(ch, messaging.UserExchange, messaging.QueueCreditUserEvents, func(msg messaging.Message) {
		if msg.Type != userdomain.EventEmailVerified {
			return
		}

		var user userdomain.User
		if err := json.Unmarshal(msg.Payload, &user); err != nil {
			slog.Error("credit: unmarshal user", "err", err)
			return
		}

		score := creditScore(user.Email)
		approved := score >= 600
		slog.Info("credit: scored", "user_id", user.ID, "email", user.Email, "score", score, "approved", approved)

		publisher.Publish(context.Background(), event.Event{
			Type: userdomain.EventCreditScored,
			Payload: userdomain.CreditScoredPayload{
				UserID:   user.ID,
				Score:    score,
				Approved: approved,
			},
		})
	}); err != nil {
		log.Fatal("rabbit subscribe:", err)
	}

	slog.Info("credit service listening", "exchange", messaging.UserExchange)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("credit service stopping")
}
