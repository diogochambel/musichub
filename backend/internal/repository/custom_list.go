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

type MongoCustomListRepository struct {
	collection *mongo.Collection
}

func NewMongoCustomListRepository(mongodb *db.MongoDB) *MongoCustomListRepository {
	collection := mongodb.Collection("custom_lists")
	r := &MongoCustomListRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoCustomListRepository) createIndexes(ctx context.Context) {
	uniqueIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{uniqueIdx})
}

func (r *MongoCustomListRepository) Insert(ctx context.Context, list *models.CustomList) error {
	_, err := r.collection.InsertOne(ctx, list)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoCustomListRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.CustomList, error) {
	var list models.CustomList
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *MongoCustomListRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.CustomList, error) {
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "user_id", Value: userID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var lists []*models.CustomList
	if err := cursor.All(ctx, &lists); err != nil {
		return nil, err
	}
	return lists, nil
}

func (r *MongoCustomListRepository) FindByUserAndName(ctx context.Context, userID primitive.ObjectID, name string) (*models.CustomList, error) {
	var list models.CustomList
	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "name", Value: name},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *MongoCustomListRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoCustomListRepository) Update(ctx context.Context, list *models.CustomList) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: list.ID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "updated_at", Value: list.UpdatedAt},
		}}},
	)
	return err
}
