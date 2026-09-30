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

type MongoCustomListItemRepository struct {
	collection *mongo.Collection
}

func NewMongoCustomListItemRepository(mongodb *db.MongoDB) *MongoCustomListItemRepository {
	collection := mongodb.Collection("custom_list_items")
	r := &MongoCustomListItemRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoCustomListItemRepository) createIndexes(ctx context.Context) {
	uniqueIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "list_id", Value: 1},
			{Key: "album_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{uniqueIdx})
}

func (r *MongoCustomListItemRepository) Insert(ctx context.Context, item *models.CustomListItem) error {
	_, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoCustomListItemRepository) CountByListID(ctx context.Context, listID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.D{{Key: "list_id", Value: listID}})
}

func (r *MongoCustomListItemRepository) DeleteByListID(ctx context.Context, listID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.D{{Key: "list_id", Value: listID}})
	return err
}

func (r *MongoCustomListItemRepository) DeleteByListIDAndAlbumID(ctx context.Context, listID, albumID primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{
		{Key: "list_id", Value: listID},
		{Key: "album_id", Value: albumID},
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoCustomListItemRepository) FindByListID(ctx context.Context, listID primitive.ObjectID) ([]*models.CustomListItem, error) {
	opts := options.Find().SetSort(bson.D{{Key: "added_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "list_id", Value: listID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*models.CustomListItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
