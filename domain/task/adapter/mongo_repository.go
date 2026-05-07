package adapter

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"todoe/domain/task/domain"
)

const (
	maxRetries    = 3
	retryInterval = 1 * time.Second
)

type MongoRepository struct {
	clientIO    mo.IOEither[*mongo.Client]
	once        sync.Once
	cached      mo.Either[error, *mongo.Client]
	initialized atomic.Bool
	fileWriter  *FileWriter
}

func NewMongoRepository(clientIO mo.IOEither[*mongo.Client], fallbackPath string) *MongoRepository {
	return &MongoRepository{
		clientIO:   clientIO,
		fileWriter: NewFileWriter(fallbackPath),
	}
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
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		col, err := r.collection()
		if err != nil {
			lastErr = err
		} else if _, err := col.InsertOne(ctx, task); err != nil {
			lastErr = err
		} else {
			return mo.Ok(struct{}{})
		}

		slog.Warn("task: mongo save failed",
			"attempt", attempt,
			"max_retries", maxRetries,
			"err", lastErr,
		)
		if attempt < maxRetries {
			time.Sleep(retryInterval)
		}
	}

	slog.Error("task: all mongo retries exhausted, falling back to file",
		"err", lastErr,
	)
	if err := r.fileWriter.Write(task); err != nil {
		slog.Error("task: file fallback also failed", "err", err)
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

// FindAll returns only the current state of each task (leaf records — not referenced as origin_id).
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
