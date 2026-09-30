package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionItemRepo interface {
	Insert(ctx context.Context, item *models.CollectionItem) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.CollectionItem, error)
	FindByUserAndAlbumAndEAN(ctx context.Context, userID, albumID primitive.ObjectID, ean13 string) (*models.CollectionItem, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.CollectionItem, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type CollectionAlbumRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error)
}

type CollectionArtistRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
}

type CollectionHandler struct {
	CollectionRepo CollectionItemRepo
	AlbumRepo      CollectionAlbumRepo
	ArtistRepo     CollectionArtistRepo
}

func NewCollectionHandler(collectionRepo CollectionItemRepo, albumRepo CollectionAlbumRepo, artistRepo CollectionArtistRepo) *CollectionHandler {
	return &CollectionHandler{
		CollectionRepo: collectionRepo,
		AlbumRepo:      albumRepo,
		ArtistRepo:     artistRepo,
	}
}

type CollectionItemResponse struct {
	ID          primitive.ObjectID `json:"id"`
	AlbumID     primitive.ObjectID `json:"album_id"`
	AlbumTitle  string             `json:"album_title"`
	ReleaseYear int                `json:"release_year"`
	ArtistName  string             `json:"artist_name"`
	EAN13       string             `json:"ean_13"`
	Support     models.SupportType `json:"support"`
	VersionName string             `json:"version_name,omitempty"`
	AddedAt     time.Time          `json:"added_at"`
}

func (h *CollectionHandler) HandleCollections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCollection(w, r)
	case http.MethodPost:
		h.addToCollection(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CollectionHandler) HandleCollectionItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.removeFromCollection(w, r)
}

func (h *CollectionHandler) getCollection(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(authData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	items, err := h.CollectionRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get collection")
		return
	}

	resp := make([]CollectionItemResponse, 0, len(items))
	for _, item := range items {
		entry := CollectionItemResponse{
			ID:          item.ID,
			AlbumID:     item.AlbumID,
			EAN13:       item.EAN13,
			Support:     item.Support,
			VersionName: item.VersionName,
			AddedAt:     item.AddedAt,
		}
		album, err := h.AlbumRepo.FindByID(r.Context(), item.AlbumID)
		if err == nil {
			entry.AlbumTitle = album.Title
			entry.ReleaseYear = album.ReleaseYear
			if album.ArtistID != nil {
				artist, err := h.ArtistRepo.FindByID(r.Context(), *album.ArtistID)
				if err == nil {
					entry.ArtistName = artist.Name
				}
			}
		}
		resp = append(resp, entry)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *CollectionHandler) addToCollection(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	var req models.CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := models.ValidateCreateCollectionRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := primitive.ObjectIDFromHex(authData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	albumID, err := primitive.ObjectIDFromHex(req.AlbumID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid album_id")
		return
	}

	_, err = h.CollectionRepo.FindByUserAndAlbumAndEAN(r.Context(), userID, albumID, req.EAN13)
	if err == nil {
		writeError(w, http.StatusConflict, "this version is already in your collection")
		return
	}

	item := &models.CollectionItem{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		AlbumID:     albumID,
		EAN13:       req.EAN13,
		Support:     req.Support,
		VersionName: req.VersionName,
		AddedAt:     time.Now(),
	}

	if err := h.CollectionRepo.Insert(r.Context(), item); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add to collection")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *CollectionHandler) removeFromCollection(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/collections/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid collection item id")
		return
	}

	item, err := h.CollectionRepo.FindByID(r.Context(), id)
	if err == mongo.ErrNoDocuments || item == nil {
		writeError(w, http.StatusNotFound, "collection item not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find collection item")
		return
	}

	if item.UserID.Hex() != authData.UserID {
		writeError(w, http.StatusForbidden, "you can only remove items from your own collection")
		return
	}

	if err := h.CollectionRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove from collection")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "removed from collection"})
}
