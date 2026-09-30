package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CustomList represents a user-created list of albums.
type CustomList struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Name      string             `json:"name" bson:"name"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateCustomListRequest is the payload for creating a new custom list.
type CreateCustomListRequest struct {
	Name string `json:"name"`
}

// ValidateCreateCustomListRequest validates the payload for creating a custom list.
func ValidateCreateCustomListRequest(req *CreateCustomListRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	return nil
}
