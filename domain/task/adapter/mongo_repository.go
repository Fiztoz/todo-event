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

// FindAll returns only current task state — leaf records not referenced as origin_id by any other task.
func (r *MongoRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "tasks"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "origin_id"},
			{Key: "as", Value: "children"},
		}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "children", Value: bson.D{{Key: "$size", Value: 0}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{{Key: "children", Value: 0}}}},
	}
	cursor, err := col.Aggregate(ctx, pipeline)
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

func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	var task domain.Task
	if err := col.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&task); err != nil {
		return mo.Err[domain.Task](err)
	}
	return mo.Ok(task)
}
