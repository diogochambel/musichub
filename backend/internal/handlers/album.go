package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlbumDetailRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error)
	FindAll(ctx context.Context) ([]*models.Album, error)
	SearchByTitle(ctx context.Context, query string) ([]*models.Album, error)
}

type AlbumDetailArtistRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
}

type AlbumDetailSongRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Song, error)
}

type AlbumHandler struct {
	AlbumRepo  AlbumDetailRepo
	ArtistRepo AlbumDetailArtistRepo
	SongRepo   AlbumDetailSongRepo
}

func NewAlbumHandler(albumRepo AlbumDetailRepo, artistRepo AlbumDetailArtistRepo, songRepo AlbumDetailSongRepo) *AlbumHandler {
	return &AlbumHandler{AlbumRepo: albumRepo, ArtistRepo: artistRepo, SongRepo: songRepo}
}

type TrackWithSong struct {
	TrackNumber     int                `json:"track_number"`
	SongID          primitive.ObjectID `json:"song_id"`
	Title           string             `json:"title"`
	DurationSeconds int                `json:"duration_seconds"`
}

type AlbumDetailResponse struct {
	ID          primitive.ObjectID    `json:"id"`
	MBID        string                `json:"mbid"`
	Title       string                `json:"title"`
	ReleaseYear int                   `json:"release_year"`
	Type        models.AlbumType      `json:"type"`
	ArtistID    *primitive.ObjectID   `json:"artist_id,omitempty"`
	ArtistName  string                `json:"artist_name,omitempty"`
	Tracks      []TrackWithSong       `json:"tracks"`
	Releases    []models.AlbumRelease `json:"releases"`
	ImageURL    *string               `json:"image_url,omitempty"`
}

func (h *AlbumHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/albums/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid album id")
		return
	}

	album, err := h.AlbumRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	resp := AlbumDetailResponse{
		ID:          album.ID,
		MBID:        album.MBID,
		Title:       album.Title,
		ReleaseYear: album.ReleaseYear,
		Type:        album.Type,
		ArtistID:    album.ArtistID,
		Releases:    album.Releases,
		ImageURL:    album.ImageURL,
	}

	if album.ArtistID != nil {
		artist, err := h.ArtistRepo.FindByID(r.Context(), *album.ArtistID)
		if err == nil {
			resp.ArtistName = artist.Name
		}
	}

	if len(album.Tracks) > 0 {
		resp.Tracks = make([]TrackWithSong, 0, len(album.Tracks))
		for _, t := range album.Tracks {
			track := TrackWithSong{
				TrackNumber: t.TrackNumber,
				SongID:      t.SongID,
			}
			song, err := h.SongRepo.FindByID(r.Context(), t.SongID)
			if err == nil {
				track.Title = song.Title
				track.DurationSeconds = song.DurationSeconds
			}
			resp.Tracks = append(resp.Tracks, track)
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AlbumHandler) SearchAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) == 0 {
		writeError(w, http.StatusBadRequest, "query param 'q' is required")
		return
	}

	albums, err := h.AlbumRepo.SearchByTitle(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	writeJSON(w, http.StatusOK, albums)
}

type AlbumListItem struct {
	ID         primitive.ObjectID `json:"id"`
	Title      string             `json:"title"`
	ArtistName string             `json:"artist_name,omitempty"`
	ImageURL   *string            `json:"image_url,omitempty"`
}

func (h *AlbumHandler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	albums, err := h.AlbumRepo.FindAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch albums")
		return
	}

	results := make([]AlbumListItem, 0, len(albums))
	for _, album := range albums {
		item := AlbumListItem{
			ID:       album.ID,
			Title:    album.Title,
			ImageURL: album.ImageURL,
		}
		if album.ArtistID != nil {
			artist, err := h.ArtistRepo.FindByID(r.Context(), *album.ArtistID)
			if err == nil {
				item.ArtistName = artist.Name
			}
		}
		results = append(results, item)
	}

	writeJSON(w, http.StatusOK, results)
}
