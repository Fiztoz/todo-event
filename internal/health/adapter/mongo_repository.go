package adapter

import (
	"context"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/mongo"

	pkgio "todoe/pkg/io"
)

type MongoRepository struct {
	clientIO *pkgio.IO[*mongo.Client]
}

func NewMongoRepository(clientIO *pkgio.IO[*mongo.Client]) *MongoRepository {
	return &MongoRepository{clientIO: clientIO}
}

func (r *MongoRepository) Ping(ctx context.Context) mo.Result[struct{}] {
	clientResult := r.clientIO.Run()
	if clientResult.IsError() {
		return mo.Err[struct{}](clientResult.Error())
	}
	if err := clientResult.MustGet().Ping(ctx, nil); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}
