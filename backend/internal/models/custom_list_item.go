package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomListItem represents an album inside a custom list.
type CustomListItem struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	ListID  primitive.ObjectID `json:"list_id" bson:"list_id"`
	AlbumID primitive.ObjectID `json:"album_id" bson:"album_id"`
	AddedAt time.Time          `json:"added_at" bson:"added_at"`
}
