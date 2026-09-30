package repository

import (
	"context"
	"fmt"

	"github.com/dariolbs/PSI/internal/db"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoNotificationRepository struct {
	collection *mongo.Collection
}

func NewMongoNotificationRepository(mongodb *db.MongoDB) *MongoNotificationRepository {
	collection := mongodb.Collection("notifications")
	r := &MongoNotificationRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoNotificationRepository) createIndexes(ctx context.Context) {
	userIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}

	readIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "read", Value: 1}},
	}

	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, readIdx})
}

func (r *MongoNotificationRepository) Insert(ctx context.Context, n *models.Notification) error {
	_, err := r.collection.InsertOne(ctx, n)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}

	return nil
}

func (r *MongoNotificationRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Notification, error) {
	var n models.Notification

	err := r.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: id},
	}).Decode(&n)

	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *MongoNotificationRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Notification, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},
	})

	cursor, err := r.collection.Find(ctx, bson.D{
		{Key: "user_id", Value: userID},
	}, opts)

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *MongoNotificationRepository) FindUnreadByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Notification, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "created_at", Value: -1},
	})

	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "read", Value: false},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *MongoNotificationRepository) MarkRead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "user_id", Value: userID},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "read", Value: true},
		}},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *MongoNotificationRepository) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "read", Value: false},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "read", Value: true},
			}},
		},
	)

	return err
}

func (r *MongoNotificationRepository) CountUnread(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.D{
		{Key: "user_id", Value: userID},
		{Key: "read", Value: false},
	})
}

func (r *MongoNotificationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{
		{Key: "_id", Value: id},
	})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
