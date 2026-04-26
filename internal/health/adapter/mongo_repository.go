package adapter

import (
	"context"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoRepository struct {
	client *mongo.Client
}

func NewMongoRepository(client *mongo.Client) *MongoRepository {
	return &MongoRepository{client: client}
}

func (r *MongoRepository) Ping(ctx context.Context) mo.Result[struct{}] {
	if err := r.client.Ping(ctx, nil); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}
