package adapter

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	taskdomain "todoe/domain/task/domain"
)

type MongoRepository struct {
	clientIO    mo.IOEither[*mongo.Client]
	once        sync.Once
	cached      mo.Either[error, *mongo.Client]
	initialized atomic.Bool
}

func NewMongoRepository(clientIO mo.IOEither[*mongo.Client]) *MongoRepository {
	return &MongoRepository{clientIO: clientIO}
}

func (r *MongoRepository) getClient() mo.Either[error, *mongo.Client] {
	r.once.Do(func() {
		r.cached = r.clientIO.Run()
		r.initialized.Store(true)
	})
	return r.cached
}

func (r *MongoRepository) collection() (*mongo.Collection, error) {
	either := r.getClient()
	if either.IsLeft() {
		return nil, either.MustLeft()
	}
	return either.MustRight().Database("todoe").Collection("tasks"), nil
}

func (r *MongoRepository) Upsert(ctx context.Context, task taskdomain.Task) mo.Result[struct{}] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[struct{}](err)
	}

	filter := bson.D{{Key: "_id", Value: task.ID}}
	update := bson.D{{Key: "$set", Value: task}}
	opts := options.UpdateOne().SetUpsert(true)

	if _, err := col.UpdateOne(ctx, filter, update, opts); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}
