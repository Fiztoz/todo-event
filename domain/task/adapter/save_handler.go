package adapter

import (
	"context"
	"log"

	"todoe/domain/task/domain"
)

func NewSaveHandler(repo *MongoRepository) func(any) {
	return func(payload any) {
		var task domain.Task
		switch p := payload.(type) {
		case domain.CreatedPayload:
			task = p.Task
		case domain.StatusChangedPayload:
			task = p.Task
		default:
			return
		}
		if result := repo.Save(context.Background(), task); result.IsError() {
			log.Printf("save task failed: %v", result.Error())
		}
	}
}
