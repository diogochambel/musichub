package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Song struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id"`
	ISRC            string               `json:"isrc" bson:"isrc"`
	Title           string               `json:"title" bson:"title"`
	DurationSeconds int                  `json:"duration_seconds" bson:"duration_seconds"`
	ArtistIDs       []primitive.ObjectID `json:"artist_ids" bson:"artist_ids"`
}

type CreateSongRequest struct {
	ISRC            string   `json:"isrc"`
	Title           string   `json:"title"`
	DurationSeconds int      `json:"duration_seconds"`
	ArtistIDs       []string `json:"artist_ids"`
}

type UpdateSongRequest struct {
	ISRC            *string  `json:"isrc"`
	Title           *string  `json:"title"`
	DurationSeconds *int     `json:"duration_seconds"`
	ArtistIDs       []string `json:"artist_ids,omitempty"`
}

func ValidateISRC(isrc string) error {
	if isrc == "" {
		return fmt.Errorf("isrc is required")
	}
	return nil
}

func ValidateSongTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

func ValidateDurationSeconds(d int) error {
	if d <= 0 {
		return fmt.Errorf("duration_seconds must be a positive integer")
	}
	return nil
}

func ValidateCreateSongRequest(req *CreateSongRequest) error {
	if err := ValidateISRC(req.ISRC); err != nil {
		return err
	}
	if err := ValidateSongTitle(req.Title); err != nil {
		return err
	}
	if err := ValidateDurationSeconds(req.DurationSeconds); err != nil {
		return err
	}
	return nil
}
