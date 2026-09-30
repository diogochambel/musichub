package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SongDetailRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Song, error)
}

type SongDetailArtistRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
}

type SongDetailAlbumRepo interface {
	FindBySongID(ctx context.Context, songID primitive.ObjectID) ([]*models.Album, error)
}

type SongHandler struct {
	SongRepo   SongDetailRepo
	ArtistRepo SongDetailArtistRepo
	AlbumRepo  SongDetailAlbumRepo
}

func NewSongHandler(songRepo SongDetailRepo, artistRepo SongDetailArtistRepo, albumRepo SongDetailAlbumRepo) *SongHandler {
	return &SongHandler{SongRepo: songRepo, ArtistRepo: artistRepo, AlbumRepo: albumRepo}
}

type SongAlbumBrief struct {
	ID          primitive.ObjectID `json:"id"`
	Title       string             `json:"title"`
	ReleaseYear int                `json:"release_year"`
}

type SongDetailResponse struct {
	ID              primitive.ObjectID   `json:"id"`
	ISRC            string               `json:"isrc"`
	Title           string               `json:"title"`
	DurationSeconds int                  `json:"duration_seconds"`
	ArtistIDs       []primitive.ObjectID `json:"artist_ids"`
	Artists         []MemberInfo         `json:"artists"`
	Albums          []SongAlbumBrief     `json:"albums"`
}

func (h *SongHandler) GetSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/songs/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid song id")
		return
	}

	song, err := h.SongRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "song not found")
		return
	}

	resp := SongDetailResponse{
		ID:              song.ID,
		ISRC:            song.ISRC,
		Title:           song.Title,
		DurationSeconds: song.DurationSeconds,
		ArtistIDs:       song.ArtistIDs,
	}

	if len(song.ArtistIDs) > 0 {
		resp.Artists = make([]MemberInfo, 0, len(song.ArtistIDs))
		for _, aid := range song.ArtistIDs {
			artist, err := h.ArtistRepo.FindByID(r.Context(), aid)
			if err == nil {
				resp.Artists = append(resp.Artists, MemberInfo{ID: artist.ID, Name: artist.Name})
			}
		}
	}

	albums, err := h.AlbumRepo.FindBySongID(r.Context(), id)
	if err == nil && albums != nil {
		resp.Albums = make([]SongAlbumBrief, 0, len(albums))
		for _, a := range albums {
			resp.Albums = append(resp.Albums, SongAlbumBrief{
				ID:          a.ID,
				Title:       a.Title,
				ReleaseYear: a.ReleaseYear,
			})
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
