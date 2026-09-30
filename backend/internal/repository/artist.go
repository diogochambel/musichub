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

type MongoArtistRepository struct {
	collection *mongo.Collection
}

func NewMongoArtistRepository(mongodb *db.MongoDB) *MongoArtistRepository {
	collection := mongodb.Collection("artists")
	r := &MongoArtistRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoArtistRepository) createIndexes(ctx context.Context) {
	isniIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "isni", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	nameIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{isniIdx, nameIdx})
}

func (r *MongoArtistRepository) Insert(ctx context.Context, artist *models.Artist) error {
	_, err := r.collection.InsertOne(ctx, artist)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoArtistRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error) {
	var artist models.Artist
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&artist)
	if err != nil {
		return nil, err
	}
	return &artist, nil
}

func (r *MongoArtistRepository) FindAll(ctx context.Context) ([]*models.Artist, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var artists []*models.Artist
	if err := cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (r *MongoArtistRepository) Update(ctx context.Context, id primitive.ObjectID, artist *models.Artist) error {
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: id}}, artist)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoArtistRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoArtistRepository) SearchByName(ctx context.Context, name string) ([]*models.Artist, error) {
	regex := primitive.Regex{Pattern: name, Options: "i"}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "name", Value: regex}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var artists []*models.Artist
	if err := cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (r *MongoArtistRepository) FindRandom(ctx context.Context, limit int) ([]*models.Artist, error) {
	pipeline := []bson.D{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit}}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var artists []*models.Artist
	if err := cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (r *MongoArtistRepository) FindByISNI(ctx context.Context, isni string) (*models.Artist, error) {
	var artist models.Artist
	err := r.collection.FindOne(ctx, bson.D{{Key: "isni", Value: isni}}).Decode(&artist)
	if err != nil {
		return nil, err
	}
	return &artist, nil
}
