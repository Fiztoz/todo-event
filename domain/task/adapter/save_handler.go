package adapter

import (
	"context"
	"log"

	"todoe/domain/task/domain"
)

func NewSaveHandler(repo *MongoRepository) func(any) {
	return func(payload any) {
		p, ok := payload.(domain.CreatedPayload)
		if !ok {
			return
		}
		if result := repo.Save(context.Background(), p.Task); result.IsError() {
			log.Printf("save task failed: %v", result.Error())
		}
	}
}
