package adapter

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"todoe/domain/task/domain"
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

func (r *MongoRepository) Save(ctx context.Context, task domain.Task) mo.Result[struct{}] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[struct{}](err)
	}
	if _, err := col.InsertOne(ctx, task); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

func (r *MongoRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	defer cursor.Close(ctx)
	var tasks []domain.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return mo.Err[[]domain.Task](err)
	}
	return mo.Ok(tasks)
}
