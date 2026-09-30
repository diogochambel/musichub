package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArtistType string

const (
	ArtistTypeSolo  ArtistType = "solo"
	ArtistTypeGroup ArtistType = "group"
)

type Artist struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id"`
	ISNI      string               `json:"isni" bson:"isni"`
	Name      string               `json:"name" bson:"name"`
	StartYear int                  `json:"start_year" bson:"start_year"`
	Type      ArtistType           `json:"type" bson:"type"`
	MemberIDs []primitive.ObjectID `json:"member_ids,omitempty" bson:"member_ids,omitempty"`
}

type CreateArtistRequest struct {
	ISNI      string   `json:"isni"`
	Name      string   `json:"name"`
	StartYear int      `json:"start_year"`
	Type      string   `json:"type"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

type UpdateArtistRequest struct {
	ISNI      *string  `json:"isni"`
	Name      *string  `json:"name"`
	StartYear *int     `json:"start_year"`
	Type      *string  `json:"type"`
	MemberIDs []string `json:"member_ids,omitempty"`
}

func ValidateISNI(isni string) error {
	if isni == "" {
		return fmt.Errorf("isni is required")
	}
	if len(isni) != 16 {
		return fmt.Errorf("isni must be 16 characters")
	}
	return nil
}

func ValidateArtistName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func ValidateStartYear(year int) error {
	if year <= 0 {
		return fmt.Errorf("start_year must be a positive integer")
	}
	return nil
}

func ValidateArtistType(artistType string) error {
	if artistType == "" {
		return fmt.Errorf("type is required")
	}
	switch ArtistType(artistType) {
	case ArtistTypeSolo, ArtistTypeGroup:
		return nil
	default:
		return fmt.Errorf("type must be 'solo' or 'group'")
	}
}

func ValidateCreateArtistRequest(req *CreateArtistRequest) error {
	if err := ValidateISNI(req.ISNI); err != nil {
		return err
	}
	if err := ValidateArtistName(req.Name); err != nil {
		return err
	}
	if err := ValidateStartYear(req.StartYear); err != nil {
		return err
	}
	if err := ValidateArtistType(req.Type); err != nil {
		return err
	}
	return nil
}
