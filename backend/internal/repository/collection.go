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

type MongoCollectionRepository struct {
	collection *mongo.Collection
}

func NewMongoCollectionRepository(mongodb *db.MongoDB) *MongoCollectionRepository {
	collection := mongodb.Collection("user_collections")
	r := &MongoCollectionRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoCollectionRepository) createIndexes(ctx context.Context) {
	uniqueIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "album_id", Value: 1},
			{Key: "ean_13", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{uniqueIdx})
}

func (r *MongoCollectionRepository) Insert(ctx context.Context, item *models.CollectionItem) error {
	_, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoCollectionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.CollectionItem, error) {
	var item models.CollectionItem
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MongoCollectionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.CollectionItem, error) {
	opts := options.Find().SetSort(bson.D{{Key: "added_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "user_id", Value: userID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*models.CollectionItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MongoCollectionRepository) FindByUserAndAlbumAndEAN(ctx context.Context, userID, albumID primitive.ObjectID, ean13 string) (*models.CollectionItem, error) {
	var item models.CollectionItem
	filter := bson.D{
		{Key: "user_id", Value: userID},
		{Key: "album_id", Value: albumID},
		{Key: "ean_13", Value: ean13},
	}
	err := r.collection.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MongoCollectionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
