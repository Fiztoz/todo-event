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

	taskdomain "todoe/domain/task/domain"
)

// MongoRepository provides read-only access to the tasks collection
// for SLA duration calculations.
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

// GetPendingCreatedAt follows the origin_id chain starting from id until it
// reaches the root task (origin_id == nil, status "pending") and returns
// its created_at.
func (r *MongoRepository) GetPendingCreatedAt(ctx context.Context, id bson.ObjectID) (time.Time, bool) {
	col, err := r.collection()
	if err != nil {
		slog.Error("sla: get collection", "err", err)
		return time.Now(), false
	}

	currentID := id
	for {
		var task taskdomain.Task
		if err := col.FindOne(ctx, bson.D{{Key: "_id", Value: currentID}}).Decode(&task); err != nil {
			slog.Error("sla: find by id", "err", err, "id", currentID)
			return time.Now(), false
		}

		if task.OriginID == nil {
			// Root task — this is the pending one
			return task.CreatedAt, true
		}
		currentID = *task.OriginID
	}
}
