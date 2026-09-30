package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Insert(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	FindAll(ctx context.Context) ([]*models.User, error)
	Update(ctx context.Context, id primitive.ObjectID, user *models.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	SetFavoriteArtistID(ctx context.Context, id primitive.ObjectID, artistID primitive.ObjectID) error
	RemoveFavoriteArtistID(ctx context.Context, id primitive.ObjectID) error
}

type ArtistFinder interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error)
}

type UserHandler struct {
	Repo       UserRepository
	ArtistRepo ArtistFinder
}

func NewUserHandler(repo UserRepository, artistRepo ArtistFinder) *UserHandler {
	return &UserHandler{Repo: repo, ArtistRepo: artistRepo}
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Status: "error", Message: message})
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := models.ValidateCreateRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing, _ := h.Repo.FindByUsername(r.Context(), req.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}

	existing, _ = h.Repo.FindByEmail(r.Context(), req.Email)
	if existing != nil {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	salt, err := models.GenerateSalt()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	user := &models.User{
		ID:        primitive.NewObjectID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  models.HashPassword(req.Password, salt),
		Salt:      salt,
		BirthDate: req.BirthDate,
	}

	if err := h.Repo.Insert(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.Repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	users, err := h.Repo.FindAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	if users == nil {
		users = []*models.User{}
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	existing, err := h.Repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentPassword == nil || *req.CurrentPassword == "" {
		writeError(w, http.StatusUnauthorized, "current password is required")
		return
	}

	if existing.Password != models.HashPassword(*req.CurrentPassword, existing.Salt) {
		writeError(w, http.StatusUnauthorized, "incorrect current password")
		return
	}

	if req.Username != nil {
		if err := models.ValidateUsername(*req.Username); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		duplicate, _ := h.Repo.FindByUsername(r.Context(), *req.Username)
		if duplicate != nil && duplicate.ID != id {
			writeError(w, http.StatusConflict, "username already exists")
			return
		}
		existing.Username = *req.Username
	}

	if req.Email != nil {
		if err := models.ValidateEmail(*req.Email); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		duplicate, _ := h.Repo.FindByEmail(r.Context(), *req.Email)
		if duplicate != nil && duplicate.ID != id {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		existing.Email = *req.Email
	}

	if req.Password != nil {
		if err := models.ValidatePassword(*req.Password); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		salt, err := models.GenerateSalt()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		existing.Salt = salt
		existing.Password = models.HashPassword(*req.Password, salt)
	}

	if req.BirthDate != nil {
		if err := models.ValidateBirthDate(*req.BirthDate); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		existing.BirthDate = *req.BirthDate
	}

	if err := h.Repo.Update(r.Context(), id, existing); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(w, http.StatusConflict, "username or email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "no documents") {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) SetFavoriteArtist(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	authUserData := middleware.GetUserData(r)
	if authUserData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	if authUserData.UserID != userID.Hex() {
		writeError(w, http.StatusForbidden, "authenticated user id must match the user id")
		return
	}

	user, err := h.Repo.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if user.FavoriteArtistID != nil {
		writeError(w, http.StatusBadRequest, "user already has a favorite artist")
		return
	}

	var req models.SetFavoriteArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	artistID, err := primitive.ObjectIDFromHex(req.ArtistID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid artist id")
		return
	}

	if _, err = h.ArtistRepo.FindByID(r.Context(), artistID); err != nil {
		writeError(w, http.StatusNotFound, "artist not found")
		return
	}

	if user.FavoriteArtistID != nil && *user.FavoriteArtistID != artistID {
		writeError(w, http.StatusConflict, "you already have a different favorite artist")
		return
	}

	if err := h.Repo.SetFavoriteArtistID(r.Context(), userID, artistID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set favorite artist")
		return
	}

	user, _ = h.Repo.FindByID(r.Context(), userID)
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) RemoveFavoriteArtist(w http.ResponseWriter, r *http.Request, idStr string) {
	userID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	authData := middleware.GetUserData(r)
	if authData == nil {
		writeError(w, http.StatusUnauthorized, "user must be authenticated")
		return
	}

	if authData.UserID != idStr {
		writeError(w, http.StatusForbidden, "authenticated user id must match the user id")
		return
	}

	if _, err = h.Repo.FindByID(r.Context(), userID); err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.Repo.RemoveFavoriteArtistID(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove favorite artist")
		return
	}

	user, _ := h.Repo.FindByID(r.Context(), userID)
	writeJSON(w, http.StatusOK, user)
}
