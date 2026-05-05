package adapter

import (
	"context"
	"log"

	"todoe/domain/task/domain"
)

func NewUpdateStatusHandler(repo *MongoRepository) func(any) {
	return func(payload any) {
		p, ok := payload.(domain.StatusChangedPayload)
		if !ok {
			return
		}
		if result := repo.UpdateStatus(context.Background(), p.TaskID, p.Status); result.IsError() {
			log.Printf("update task status failed: %v", result.Error())
		}
	}
}
