package handlers

import (
	"context"
	"net/http"

	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumRepository interface {
	FindRandom(ctx context.Context, limit int) ([]*models.Album, error)
	FindRecent(ctx context.Context, sinceYear int, limit int) ([]*models.Album, error)
}

type ArtistRepository interface {
	FindRandom(ctx context.Context, limit int) ([]*models.Artist, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
}

type HomeHandler struct {
	AlbumRepo  AlbumRepository
	ArtistRepo ArtistRepository
}

func NewHomeHandler(albumRepo AlbumRepository, artistRepo ArtistRepository) *HomeHandler {
	return &HomeHandler{AlbumRepo: albumRepo, ArtistRepo: artistRepo}
}

type AlbumWithArtist struct {
	ID          primitive.ObjectID    `json:"id"`
	MBID        string                `json:"mbid"`
	Title       string                `json:"title"`
	ReleaseYear int                   `json:"release_year"`
	Type        models.AlbumType      `json:"type"`
	ArtistName  string                `json:"artist_name,omitempty"`
	Tracks      []models.TrackEntry   `json:"tracks"`
	Releases    []models.AlbumRelease `json:"releases"`
	ImageURL    *string               `json:"image_url,omitempty"`
}

func (h *HomeHandler) albumsWithArtistNames(r *http.Request, albums []*models.Album) []AlbumWithArtist {
	results := make([]AlbumWithArtist, 0, len(albums))
	for _, album := range albums {
		item := AlbumWithArtist{
			ID:          album.ID,
			MBID:        album.MBID,
			Title:       album.Title,
			ReleaseYear: album.ReleaseYear,
			Type:        album.Type,
			Tracks:      album.Tracks,
			Releases:    album.Releases,
			ImageURL:    album.ImageURL,
		}
		if album.ArtistID != nil {
			artist, err := h.ArtistRepo.FindByID(r.Context(), *album.ArtistID)
			if err == nil {
				item.ArtistName = artist.Name
			}
		}
		results = append(results, item)
	}
	return results
}

func (h *HomeHandler) FeaturedAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	albums, err := h.AlbumRepo.FindRandom(r.Context(), 4)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch featured albums")
		return
	}

	if albums == nil {
		albums = []*models.Album{}
	}

	writeJSON(w, http.StatusOK, h.albumsWithArtistNames(r, albums))
}

func (h *HomeHandler) FeaturedArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	artists, err := h.ArtistRepo.FindRandom(r.Context(), 4)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch featured artists")
		return
	}

	if artists == nil {
		artists = []*models.Artist{}
	}

	writeJSON(w, http.StatusOK, artists)
}

func (h *HomeHandler) RecentReleases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	albums, err := h.AlbumRepo.FindRecent(r.Context(), 2025, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch recent releases")
		return
	}

	if albums == nil {
		albums = []*models.Album{}
	}

	writeJSON(w, http.StatusOK, h.albumsWithArtistNames(r, albums))
}
