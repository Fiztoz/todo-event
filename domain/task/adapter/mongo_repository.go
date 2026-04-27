package adapter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

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
	return either.MustRight().Database("todoe").Collection("task_events"), nil
}

func (r *MongoRepository) Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[struct{}](err)
	}
	raw, err := bson.Marshal(payload)
	if err != nil {
		return mo.Err[struct{}](err)
	}
	e := domain.StoredEvent{
		ID:          bson.NewObjectID(),
		AggregateID: aggregateID,
		Type:        eventType,
		Payload:     raw,
		CreatedAt:   time.Now(),
	}
	if _, err := col.InsertOne(ctx, e); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	cursor, err := col.Find(ctx,
		bson.D{{Key: "aggregate_id", Value: id}},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}),
	)
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	defer cursor.Close(ctx)
	var events []domain.StoredEvent
	if err := cursor.All(ctx, &events); err != nil {
		return mo.Err[domain.Task](err)
	}
	task, err := domain.Apply(id, events)
	if err != nil {
		return mo.Err[domain.Task](err)
	}
	return mo.Ok(task)
}

func (r *MongoRepository) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	col, err := r.collection()
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$aggregate_id"},
			{Key: "events", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}}},
	}
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return mo.Err[[]domain.Task](err)
	}
	defer cursor.Close(ctx)

	var tasks []domain.Task
	for cursor.Next(ctx) {
		var group struct {
			AggregateID bson.ObjectID       `bson:"_id"`
			Events      []domain.StoredEvent `bson:"events"`
		}
		if err := cursor.Decode(&group); err != nil {
			return mo.Err[[]domain.Task](err)
		}
		task, err := domain.Apply(group.AggregateID, group.Events)
		if err != nil {
			return mo.Err[[]domain.Task](err)
		}
		tasks = append(tasks, task)
	}
	return mo.Ok(tasks)
}
