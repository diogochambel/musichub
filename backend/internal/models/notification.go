package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification represents a message sent to a user about a version request outcome.
type Notification struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	RequestID  primitive.ObjectID `json:"request_id" bson:"request_id"`
	AlbumTitle string             `json:"album_title" bson:"album_title"`
	Response   string             `json:"response" bson:"response"`
	Read       bool               `json:"read" bson:"read"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}
