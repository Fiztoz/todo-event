package domain

import "time"

type SLAEntry struct {
	ID        string
	EventType string
	CreatedAt time.Time
	Duration  time.Duration
}
