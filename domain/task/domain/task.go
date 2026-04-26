package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const EventCreated = "task.created"

type Task struct {
	ID        bson.ObjectID `bson:"_id" json:"id"`
	Title     string             `bson:"title" json:"title"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type CreatedPayload struct {
	Task Task
}
