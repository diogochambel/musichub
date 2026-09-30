package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionItem represents an album version owned by a user.
type CollectionItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	AlbumID     primitive.ObjectID `json:"album_id" bson:"album_id"`
	EAN13       string             `json:"ean_13" bson:"ean_13"`
	Support     SupportType        `json:"support" bson:"support"`
	VersionName string             `json:"version_name,omitempty" bson:"version_name,omitempty"`
	AddedAt     time.Time          `json:"added_at" bson:"added_at"`
}

// CreateCollectionRequest is the payload for adding an item to a collection.
type CreateCollectionRequest struct {
	AlbumID     string      `json:"album_id"`
	EAN13       string      `json:"ean_13"`
	Support     SupportType `json:"support"`
	VersionName string      `json:"version_name"`
}

// ValidateCreateCollectionRequest validates the payload for adding a collection item.
func ValidateCreateCollectionRequest(req *CreateCollectionRequest) error {
	if req.AlbumID == "" {
		return fmt.Errorf("album_id is required")
	}
	switch req.Support {
	case SupportTypeCD, SupportTypeVinyl, SupportTypeCassette:
	default:
		return fmt.Errorf("support must be 'CD', 'vinyl', or 'cassette'")
	}
	if err := ValidateEAN13(req.EAN13); err != nil {
		return err
	}
	return nil
}
