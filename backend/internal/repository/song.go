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

type MongoSongRepository struct {
	collection *mongo.Collection
}

func NewMongoSongRepository(mongodb *db.MongoDB) *MongoSongRepository {
	collection := mongodb.Collection("songs")
	r := &MongoSongRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoSongRepository) createIndexes(ctx context.Context) {
	isrcIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "isrc", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	titleIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "title", Value: 1}},
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{isrcIdx, titleIdx})
}

func (r *MongoSongRepository) Insert(ctx context.Context, song *models.Song) error {
	_, err := r.collection.InsertOne(ctx, song)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoSongRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Song, error) {
	var song models.Song
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *MongoSongRepository) FindAll(ctx context.Context) ([]*models.Song, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*models.Song
	if err := cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *MongoSongRepository) Update(ctx context.Context, id primitive.ObjectID, song *models.Song) error {
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: id}}, song)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoSongRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoSongRepository) FindByISRC(ctx context.Context, isrc string) (*models.Song, error) {
	var song models.Song
	err := r.collection.FindOne(ctx, bson.D{{Key: "isrc", Value: isrc}}).Decode(&song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *MongoSongRepository) SearchByTitle(ctx context.Context, title string) ([]*models.Song, error) {
	regex := primitive.Regex{Pattern: title, Options: "i"}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "title", Value: regex}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*models.Song
	if err := cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}

func (r *MongoSongRepository) FindByArtistID(ctx context.Context, artistID primitive.ObjectID) ([]*models.Song, error) {
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "artist_ids", Value: artistID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var songs []*models.Song
	if err := cursor.All(ctx, &songs); err != nil {
		return nil, err
	}
	return songs, nil
}
