package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthRepository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Insert(ctx context.Context, user *models.User) error
}

type AuthHandler struct {
	Repo AuthRepository
}

func NewAuthHandler(repo AuthRepository) *AuthHandler {
	return &AuthHandler{Repo: repo}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if body.Username == "" || body.Password == "" {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user, err := h.Repo.FindByUsername(r.Context(), body.Username)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if models.HashPassword(body.Password, user.Salt) != user.Password {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := middleware.CreateToken(user.ID.Hex(), user.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error creating session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"token":   token,
		"message": "login successful",
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
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

	token, err := middleware.CreateToken(user.ID.Hex(), user.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error creating session")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status":  "success",
		"token":   token,
		"message": "registration successful",
	})
}
