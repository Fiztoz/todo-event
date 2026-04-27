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

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	}
	return false
}

type Task struct {
	ID        bson.ObjectID  `bson:"_id"                 json:"id"`
	OriginID  *bson.ObjectID `bson:"origin_id,omitempty" json:"origin_id,omitempty"`
	Title     string         `bson:"title"               json:"title"`
	Status    Status         `bson:"status"              json:"status"`
	CreatedAt time.Time      `bson:"created_at"          json:"created_at"`
}

func NewTask(title string) Task {
	return Task{
		ID:        bson.NewObjectID(),
		Title:     title,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

func (t Task) ChangeStatus(status Status) Task {
	return Task{
		ID:        bson.NewObjectID(),
		OriginID:  &t.ID,
		Title:     t.Title,
		Status:    status,
		CreatedAt: time.Now(),
	}
}
