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

type MongoAlbumRepository struct {
	collection *mongo.Collection
}

func NewMongoAlbumRepository(mongodb *db.MongoDB) *MongoAlbumRepository {
	collection := mongodb.Collection("albums")
	r := &MongoAlbumRepository{collection: collection}
	r.createIndexes(context.Background())
	return r
}

func (r *MongoAlbumRepository) createIndexes(ctx context.Context) {
	mbidIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "mbid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	artistIdx := mongo.IndexModel{
		Keys: bson.D{{Key: "artist_id", Value: 1}},
	}
	r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{mbidIdx, artistIdx})
}

func (r *MongoAlbumRepository) Insert(ctx context.Context, album *models.Album) error {
	_, err := r.collection.InsertOne(ctx, album)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoAlbumRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error) {
	var album models.Album
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&album)
	if err != nil {
		return nil, err
	}
	return &album, nil
}

func (r *MongoAlbumRepository) FindAll(ctx context.Context) ([]*models.Album, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *MongoAlbumRepository) Update(ctx context.Context, id primitive.ObjectID, album *models.Album) error {
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: id}}, album)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return err
	}
	return nil
}

func (r *MongoAlbumRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoAlbumRepository) FindRandom(ctx context.Context, limit int) ([]*models.Album, error) {
	pipeline := []bson.D{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit}}}},
	}
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *MongoAlbumRepository) FindRecent(ctx context.Context, sinceYear int, limit int) ([]*models.Album, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "release_year", Value: -1}}).
		SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "release_year", Value: bson.D{{Key: "$gte", Value: sinceYear}}}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *MongoAlbumRepository) FindByArtistID(ctx context.Context, artistID primitive.ObjectID) ([]*models.Album, error) {
	opts := options.Find().SetSort(bson.D{{Key: "release_year", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "artist_id", Value: artistID}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *MongoAlbumRepository) FindByMBID(ctx context.Context, mbid string) (*models.Album, error) {
	var album models.Album
	err := r.collection.FindOne(ctx, bson.D{{Key: "mbid", Value: mbid}}).Decode(&album)
	if err != nil {
		return nil, err
	}
	return &album, nil
}

func (r *MongoAlbumRepository) SearchByTitle(ctx context.Context, query string) ([]*models.Album, error) {
	filter := bson.D{{Key: "title", Value: bson.D{{Key: "$regex", Value: query}, {Key: "$options", Value: "i"}}}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *MongoAlbumRepository) FindBySongID(ctx context.Context, songID primitive.ObjectID) ([]*models.Album, error) {
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "tracks.song_id", Value: songID}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var albums []*models.Album
	if err := cursor.All(ctx, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}
