package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumType string

const (
	AlbumTypeSingle AlbumType = "single"
	AlbumTypeEP     AlbumType = "EP"
	AlbumTypeLP     AlbumType = "LP"
)

type SupportType string

const (
	SupportTypeCD       SupportType = "CD"
	SupportTypeVinyl    SupportType = "vinyl"
	SupportTypeCassette SupportType = "cassette"
)

type TrackEntry struct {
	TrackNumber int                `json:"track_number" bson:"track_number"`
	SongID      primitive.ObjectID `json:"song_id" bson:"song_id"`
}

type AlbumRelease struct {
	Support     SupportType `json:"support" bson:"support"`
	VersionName string      `json:"version_name" bson:"version_name"`
	EAN13       string      `json:"ean13" bson:"ean13"`
}

type Album struct {
	ID          primitive.ObjectID  `json:"id" bson:"_id"`
	MBID        string              `json:"mbid" bson:"mbid"`
	Title       string              `json:"title" bson:"title"`
	ReleaseYear int                 `json:"release_year" bson:"release_year"`
	Type        AlbumType           `json:"type" bson:"type"`
	ArtistID    *primitive.ObjectID `json:"artist_id" bson:"artist_id,omitempty"`
	Tracks      []TrackEntry        `json:"tracks" bson:"tracks"`
	Releases    []AlbumRelease      `json:"releases" bson:"releases"`
	ImageURL    *string             `json:"image_url,omitempty" bson:"image_url,omitempty"`
}

type CreateAlbumRequest struct {
	MBID        string         `json:"mbid"`
	Title       string         `json:"title"`
	ReleaseYear int            `json:"release_year"`
	Type        string         `json:"type"`
	ArtistID    *string        `json:"artist_id,omitempty"`
	Tracks      []TrackEntry   `json:"tracks"`
	Releases    []AlbumRelease `json:"releases"`
	ImageURL    *string        `json:"image_url,omitempty"`
}

type UpdateAlbumRequest struct {
	MBID        *string        `json:"mbid"`
	Title       *string        `json:"title"`
	ReleaseYear *int           `json:"release_year"`
	Type        *string        `json:"type"`
	ArtistID    *string        `json:"artist_id"`
	Tracks      []TrackEntry   `json:"tracks"`
	Releases    []AlbumRelease `json:"releases"`
	ImageURL    *string        `json:"image_url,omitempty"`
}

func ValidateMBID(mbid string) error {
	if mbid == "" {
		return fmt.Errorf("mbid is required")
	}
	return nil
}

func ValidateAlbumTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

func ValidateReleaseYear(year int) error {
	if year <= 0 {
		return fmt.Errorf("release_year must be a positive integer")
	}
	return nil
}

func ValidateAlbumType(albumType string) error {
	if albumType == "" {
		return fmt.Errorf("type is required")
	}
	switch AlbumType(albumType) {
	case AlbumTypeSingle, AlbumTypeEP, AlbumTypeLP:
		return nil
	default:
		return fmt.Errorf("type must be 'single', 'EP', or 'LP'")
	}
}

func ValidateEAN13(ean13 string) error {
	if ean13 == "" {
		return fmt.Errorf("ean13 is required")
	}
	if len(ean13) != 13 {
		return fmt.Errorf("ean13 must be 13 digits")
	}
	for _, c := range ean13 {
		if c < '0' || c > '9' {
			return fmt.Errorf("ean13 must contain only digits")
		}
	}
	return nil
}

func ValidateAlbumRelease(r AlbumRelease) error {
	switch r.Support {
	case SupportTypeCD, SupportTypeVinyl, SupportTypeCassette:
	default:
		return fmt.Errorf("support must be 'CD', 'vinyl', or 'cassette'")
	}
	if err := ValidateEAN13(r.EAN13); err != nil {
		return err
	}
	return nil
}

func ValidateCreateAlbumRequest(req *CreateAlbumRequest) error {
	if err := ValidateMBID(req.MBID); err != nil {
		return err
	}
	if err := ValidateAlbumTitle(req.Title); err != nil {
		return err
	}
	if err := ValidateReleaseYear(req.ReleaseYear); err != nil {
		return err
	}
	if err := ValidateAlbumType(req.Type); err != nil {
		return err
	}
	for _, r := range req.Releases {
		if err := ValidateAlbumRelease(r); err != nil {
			return err
		}
	}
	return nil
}
