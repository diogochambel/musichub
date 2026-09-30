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

type MongoVersionRequestRepository struct {
	collection *mongo.Collection
}

func NewMongoVersionRequestRepository(mongodb *db.MongoDB) *MongoVersionRequestRepository {
	collection := mongodb.Collection("version_requests")
	r := &MongoVersionRequestRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoVersionRequestRepository) createIndexes(ctx context.Context) {
	userIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}
	statusIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{userIdx, statusIdx})
}

func (r *MongoVersionRequestRepository) Insert(ctx context.Context, req *models.VersionRequest) error {
	_, err := r.collection.InsertOne(ctx, req)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoVersionRequestRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.VersionRequest, error) {
	var req models.VersionRequest
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *MongoVersionRequestRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.VersionRequest, error) {
	opts := options.Find().SetSort(bson.D{{Key: "requested_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "user_id", Value: userID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*models.VersionRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *MongoVersionRequestRepository) FindByUserIDAndStatus(ctx context.Context, userID primitive.ObjectID, status models.RequestStatus) ([]*models.VersionRequest, error) {
	opts := options.Find().SetSort(bson.D{{Key: "requested_at", Value: -1}})
	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "status", Value: status},
	}
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*models.VersionRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *MongoVersionRequestRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.RequestStatus) error {
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: status}}}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoVersionRequestRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
