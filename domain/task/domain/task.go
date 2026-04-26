package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	EventCreated       = "task.created"
	EventStatusChanged = "task.status_changed"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID        bson.ObjectID  `bson:"_id" json:"id"`
	OriginID  *bson.ObjectID `bson:"origin_id,omitempty" json:"origin_id,omitempty"`
	Title     string         `bson:"title" json:"title"`
	Status    Status         `bson:"status" json:"status"`
	CreatedAt time.Time      `bson:"created_at" json:"created_at"`
}

type CreatedPayload struct {
	Task Task
}

type StatusChangedPayload struct {
	Task Task
}
