package domain

import (
	"fmt"
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
	ID        bson.ObjectID `bson:"_id"        json:"id"`
	Title     string        `bson:"title"      json:"title"`
	Status    Status        `bson:"status"     json:"status"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

type StoredEvent struct {
	ID          bson.ObjectID `bson:"_id"`
	AggregateID bson.ObjectID `bson:"aggregate_id"`
	Type        string        `bson:"type"`
	Payload     bson.Raw      `bson:"payload"`
	CreatedAt   time.Time     `bson:"created_at"`
}

type TaskCreatedPayload struct {
	Title string `bson:"title"`
}

type StatusChangedPayload struct {
	Status Status `bson:"status"`
}

func Apply(id bson.ObjectID, events []StoredEvent) (Task, error) {
	if len(events) == 0 {
		return Task{}, fmt.Errorf("task not found: %s", id.Hex())
	}
	task := Task{ID: id}
	for _, e := range events {
		switch e.Type {
		case EventCreated:
			var p TaskCreatedPayload
			bson.Unmarshal(e.Payload, &p)
			task.Title = p.Title
			task.Status = StatusPending
			task.CreatedAt = e.CreatedAt
		case EventStatusChanged:
			var p StatusChangedPayload
			bson.Unmarshal(e.Payload, &p)
			task.Status = p.Status
		}
	}
	return task, nil
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
	return Task{ID: t.ID, Title: t.Title, Status: status, CreatedAt: t.CreatedAt}
}
