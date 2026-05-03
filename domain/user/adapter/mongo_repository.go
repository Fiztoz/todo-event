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

	"todoe/domain/user/domain"
)

type StoredEvent struct {
	ID          bson.ObjectID `bson:"_id"`
	AggregateID bson.ObjectID `bson:"aggregate_id"`
	Type        string        `bson:"type"`
	Payload     bson.Raw      `bson:"payload"`
	CreatedAt   time.Time     `bson:"created_at"`
}

type MongoRepository struct {
	clientIO    mo.IOEither[*mongo.Client]
	dbName      string
	once        sync.Once
	cached      mo.Either[error, *mongo.Client]
	initialized atomic.Bool
}

func NewMongoRepository(clientIO mo.IOEither[*mongo.Client], dbName string) *MongoRepository {
	return &MongoRepository{clientIO: clientIO, dbName: dbName}
}

func (r *MongoRepository) getClient() mo.Either[error, *mongo.Client] {
	r.once.Do(func() {
		r.cached = r.clientIO.Run()
		r.initialized.Store(true)
	})
	return r.cached
}

func (r *MongoRepository) db() (*mongo.Database, error) {
	either := r.getClient()
	if either.IsLeft() {
		return nil, either.MustLeft()
	}
	return either.MustRight().Database(r.dbName), nil
}

func (r *MongoRepository) Append(ctx context.Context, aggregateID bson.ObjectID, eventType string, payload any) mo.Result[struct{}] {
	db, err := r.db()
	if err != nil {
		return mo.Err[struct{}](err)
	}
	raw, err := bson.Marshal(payload)
	if err != nil {
		return mo.Err[struct{}](err)
	}
	e := StoredEvent{
		ID:          bson.NewObjectID(),
		AggregateID: aggregateID,
		Type:        eventType,
		Payload:     raw,
		CreatedAt:   time.Now(),
	}
	if _, err := db.Collection("users_events").InsertOne(ctx, e); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

func (r *MongoRepository) Upsert(ctx context.Context, user domain.User) mo.Result[struct{}] {
	db, err := r.db()
	if err != nil {
		return mo.Err[struct{}](err)
	}
	filter := bson.D{{Key: "_id", Value: user.ID}}
	update := bson.D{{Key: "$set", Value: user}}
	if _, err := db.Collection("users_view").UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return mo.Err[struct{}](err)
	}
	return mo.Ok(struct{}{})
}

func (r *MongoRepository) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.User] {
	db, err := r.db()
	if err != nil {
		return mo.Err[domain.User](err)
	}
	var user domain.User
	if err := db.Collection("users_view").FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user); err != nil {
		return mo.Err[domain.User](err)
	}
	return mo.Ok(user)
}
