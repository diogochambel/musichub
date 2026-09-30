package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArtistDetailRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
	FindAll(ctx context.Context) ([]*models.Artist, error)
	SearchByName(ctx context.Context, name string) ([]*models.Artist, error)
}

type ArtistDetailAlbumRepo interface {
	FindByArtistID(ctx context.Context, artistID primitive.ObjectID) ([]*models.Album, error)
}

type ArtistDetailSongRepo interface {
	FindByArtistID(ctx context.Context, artistID primitive.ObjectID) ([]*models.Song, error)
}

type ArtistHandler struct {
	ArtistRepo ArtistDetailRepo
	AlbumRepo  ArtistDetailAlbumRepo
	SongRepo   ArtistDetailSongRepo
}

func NewArtistHandler(artistRepo ArtistDetailRepo, albumRepo ArtistDetailAlbumRepo, songRepo ArtistDetailSongRepo) *ArtistHandler {
	return &ArtistHandler{ArtistRepo: artistRepo, AlbumRepo: albumRepo, SongRepo: songRepo}
}

type MemberInfo struct {
	ID   primitive.ObjectID `json:"id"`
	Name string             `json:"name"`
}

type AlbumBrief struct {
	ID          primitive.ObjectID `json:"id"`
	Title       string             `json:"title"`
	ReleaseYear int                `json:"release_year"`
	Type        models.AlbumType   `json:"type"`
	ImageURL    *string            `json:"image_url,omitempty"`
}

type ArtistDetailResponse struct {
	ID        primitive.ObjectID   `json:"id"`
	ISNI      string               `json:"isni"`
	Name      string               `json:"name"`
	StartYear int                  `json:"start_year"`
	Type      models.ArtistType    `json:"type"`
	MemberIDs []primitive.ObjectID `json:"member_ids,omitempty"`
	Members   []MemberInfo         `json:"members,omitempty"`
	Albums    []AlbumBrief         `json:"albums"`
}

func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/artists/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid artist id")
		return
	}

	artist, err := h.ArtistRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "artist not found")
		return
	}

	resp := ArtistDetailResponse{
		ID:        artist.ID,
		ISNI:      artist.ISNI,
		Name:      artist.Name,
		StartYear: artist.StartYear,
		Type:      artist.Type,
		MemberIDs: artist.MemberIDs,
	}

	if len(artist.MemberIDs) > 0 {
		members := make([]MemberInfo, 0, len(artist.MemberIDs))
		for _, mid := range artist.MemberIDs {
			m, err := h.ArtistRepo.FindByID(r.Context(), mid)
			if err == nil {
				members = append(members, MemberInfo{ID: m.ID, Name: m.Name})
			}
		}
		resp.Members = members
	}

	albums, err := h.AlbumRepo.FindByArtistID(r.Context(), id)
	if err == nil && albums != nil {
		sort.Slice(albums, func(i, j int) bool {
			return albums[i].ReleaseYear > albums[j].ReleaseYear
		})
		resp.Albums = make([]AlbumBrief, 0, len(albums))
		for _, a := range albums {
			resp.Albums = append(resp.Albums, AlbumBrief{
				ID:          a.ID,
				Title:       a.Title,
				ReleaseYear: a.ReleaseYear,
				Type:        a.Type,
				ImageURL:    a.ImageURL,
			})
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

type ArtistSearchResult struct {
	ID        primitive.ObjectID `json:"id"`
	ISNI      string             `json:"isni"`
	Name      string             `json:"name"`
	StartYear int                `json:"start_year"`
	Type      models.ArtistType  `json:"type"`
}

func (h *ArtistHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "missing search query parameter 'q'")
		return
	}

	artists, err := h.ArtistRepo.SearchByName(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search artists")
		return
	}

	if artists == nil {
		artists = []*models.Artist{}
	}

	results := make([]ArtistSearchResult, 0, len(artists))
	for _, a := range artists {
		results = append(results, ArtistSearchResult{
			ID:        a.ID,
			ISNI:      a.ISNI,
			Name:      a.Name,
			StartYear: a.StartYear,
			Type:      a.Type,
		})
	}

	writeJSON(w, http.StatusOK, results)
}
