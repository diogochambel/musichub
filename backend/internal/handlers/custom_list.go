package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomListRepository interface {
	Insert(ctx context.Context, list *models.CustomList) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.CustomList, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.CustomList, error)
	FindByUserAndName(ctx context.Context, userID primitive.ObjectID, name string) (*models.CustomList, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	Update(ctx context.Context, list *models.CustomList) error
}

type CustomListItemRepo interface {
	Insert(ctx context.Context, item *models.CustomListItem) error
	FindByListID(ctx context.Context, listID primitive.ObjectID) ([]*models.CustomListItem, error)
	CountByListID(ctx context.Context, listID primitive.ObjectID) (int64, error)
	DeleteByListID(ctx context.Context, listID primitive.ObjectID) error
	DeleteByListIDAndAlbumID(ctx context.Context, listID, albumID primitive.ObjectID) error
}

type CustomListAlbumRepo interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error)
}

type CustomListHandler struct {
	CustomListRepo     CustomListRepository
	CustomListItemRepo CustomListItemRepo
	AlbumRepo          CustomListAlbumRepo
}

func NewCustomListHandler(
	customListRepo CustomListRepository,
	customListItemRepo CustomListItemRepo,
	albumRepo CustomListAlbumRepo,
) *CustomListHandler {
	return &CustomListHandler{
		CustomListRepo:     customListRepo,
		CustomListItemRepo: customListItemRepo,
		AlbumRepo:          albumRepo,
	}
}

type CustomListResponse struct {
	ID         primitive.ObjectID `json:"id"`
	Name       string             `json:"name"`
	AlbumCount int64              `json:"album_count"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CustomListItemResponse struct {
	AlbumID    primitive.ObjectID `json:"album_id"`
	AlbumTitle string             `json:"album_title"`
	AddedAt    time.Time          `json:"added_at"`
}

func (h *CustomListHandler) HandleCustomLists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCustomLists(w, r)
	case http.MethodPost:
		h.createCustomList(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CustomListHandler) HandleCustomListItem(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/items") {
		switch r.Method {
		case http.MethodGet:
			h.listItems(w, r)
		case http.MethodPost:
			h.addItem(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.Contains(path, "/items/") {
		if r.Method == http.MethodDelete {
			h.removeItem(w, r)
			return
		}

		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Method == http.MethodDelete {
		h.deleteCustomList(w, r)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *CustomListHandler) createCustomList(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	var req models.CreateCustomListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := models.ValidateCreateCustomListRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := primitive.ObjectIDFromHex(authData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	_, err = h.CustomListRepo.FindByUserAndName(r.Context(), userID, req.Name)
	if err == nil {
		writeError(w, http.StatusConflict, "you already have a list with this name")
		return
	}

	list := &models.CustomList{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.CustomListRepo.Insert(r.Context(), list); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create list")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":  "success",
		"message": "List created successfully!",
		"list":    list,
	})
}

func (h *CustomListHandler) listCustomLists(w http.ResponseWriter, r *http.Request) {
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

	lists, err := h.CustomListRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get lists")
		return
	}

	resp := make([]CustomListResponse, 0, len(lists))

	for _, list := range lists {
		count, _ := h.CustomListItemRepo.CountByListID(r.Context(), list.ID)

		resp = append(resp, CustomListResponse{
			ID:         list.ID,
			Name:       list.Name,
			AlbumCount: count,
			UpdatedAt:  list.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *CustomListHandler) deleteCustomList(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/lists/")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}

	list, err := h.CustomListRepo.FindByID(r.Context(), id)
	if err == mongo.ErrNoDocuments || list == nil {
		writeError(w, http.StatusNotFound, "list not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find list")
		return
	}

	if list.UserID.Hex() != authData.UserID {
		writeError(w, http.StatusForbidden, "you can only delete your own lists")
		return
	}

	if err := h.CustomListItemRepo.DeleteByListID(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete list items")
		return
	}

	if err := h.CustomListRepo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete list")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "list deleted successfully",
	})
}

func (h *CustomListHandler) listItems(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/lists/")
	idStr := strings.TrimSuffix(path, "/items")

	listID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}

	list, err := h.CustomListRepo.FindByID(r.Context(), listID)
	if err != nil {
		writeError(w, http.StatusNotFound, "list not found")
		return
	}

	if list.UserID.Hex() != authData.UserID {
		writeError(w, http.StatusForbidden, "you can only view your own lists")
		return
	}

	items, err := h.CustomListItemRepo.FindByListID(r.Context(), listID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get list items")
		return
	}

	response := make([]CustomListItemResponse, 0, len(items))

	for _, item := range items {
		albumTitle := "Unknown album"

		album, err := h.AlbumRepo.FindByID(r.Context(), item.AlbumID)
		if err == nil && album != nil {
			albumTitle = album.Title
		}

		response = append(response, CustomListItemResponse{
			AlbumID:    item.AlbumID,
			AlbumTitle: albumTitle,
			AddedAt:    item.AddedAt,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *CustomListHandler) addItem(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/lists/")
	idStr := strings.TrimSuffix(path, "/items")

	listID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}

	list, err := h.CustomListRepo.FindByID(r.Context(), listID)
	if err != nil {
		writeError(w, http.StatusNotFound, "list not found")
		return
	}

	if list.UserID.Hex() != authData.UserID {
		writeError(w, http.StatusForbidden, "you can only add items to your own lists")
		return
	}

	var body struct {
		AlbumID string `json:"album_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.AlbumID == "" {
		writeError(w, http.StatusBadRequest, "album_id is required")
		return
	}

	albumID, err := primitive.ObjectIDFromHex(body.AlbumID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid album_id")
		return
	}

	if _, err := h.AlbumRepo.FindByID(r.Context(), albumID); err != nil {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	item := &models.CustomListItem{
		ID:      primitive.NewObjectID(),
		ListID:  listID,
		AlbumID: albumID,
		AddedAt: time.Now(),
	}

	if err := h.CustomListItemRepo.Insert(r.Context(), item); err != nil {
		if err.Error() == "duplicate key error" {
			writeError(w, http.StatusConflict, "album already in list")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to add item")
		return
	}

	list.UpdatedAt = time.Now()
	if err := h.CustomListRepo.Update(r.Context(), list); err != nil {
		log.Printf("failed to update list timestamp: %v", err)
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *CustomListHandler) removeItem(w http.ResponseWriter, r *http.Request) {
	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/lists/")
	parts := strings.Split(path, "/items/")

	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "invalid url")
		return
	}

	listID, err := primitive.ObjectIDFromHex(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}

	albumID, err := primitive.ObjectIDFromHex(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid album id")
		return
	}

	list, err := h.CustomListRepo.FindByID(r.Context(), listID)
	if err != nil {
		writeError(w, http.StatusNotFound, "list not found")
		return
	}

	if list.UserID.Hex() != authData.UserID {
		writeError(w, http.StatusForbidden, "you can only remove items from your own lists")
		return
	}

	if err := h.CustomListItemRepo.DeleteByListIDAndAlbumID(r.Context(), listID, albumID); err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "item not found in list")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to remove item")
		return
	}

	list.UpdatedAt = time.Now()
	h.CustomListRepo.Update(r.Context(), list)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "item removed successfully",
	})
}