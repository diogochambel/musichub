package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestStatus represents the state of a version request.
type RequestStatus string

const (
	RequestStatusPending  RequestStatus = "em análise"
	RequestStatusAccepted RequestStatus = "aceite"
	RequestStatusRejected RequestStatus = "recusado"
)

// VersionRequest represents a user-submitted request for a new album version.
type VersionRequest struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	AlbumID     primitive.ObjectID `json:"album_id" bson:"album_id"`
	EAN13       string             `json:"ean_13" bson:"ean_13"`
	Support     SupportType        `json:"support" bson:"support"`
	VersionName string             `json:"version_name,omitempty" bson:"version_name,omitempty"`
	Status      RequestStatus      `json:"status" bson:"status"`
	RequestedAt time.Time          `json:"requested_at" bson:"requested_at"`
}

// CreateVersionRequestRequest is the payload for submitting a new version request.
type CreateVersionRequestRequest struct {
	AlbumID     string      `json:"album_id"`
	EAN13       string      `json:"ean13"`
	Support     SupportType `json:"support"`
	VersionName string      `json:"version_name"`
}

// ValidateCreateVersionRequestRequest validates the payload for a version request.
func ValidateCreateVersionRequestRequest(req *CreateVersionRequestRequest) error {
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

// ValidateRequestStatus checks that a status value is allowed.
func ValidateRequestStatus(status RequestStatus) error {
	switch status {
	case RequestStatusPending, RequestStatusAccepted, RequestStatusRejected:
		return nil
	default:
		return fmt.Errorf("status must be 'em análise', 'aceite', or 'recusado'")
	}
}
