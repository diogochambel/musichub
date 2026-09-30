package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dariolbs/PSI/internal/db"
)

type HealthHandler struct {
	DB *db.MongoDB
}

func NewHealthHandler(database *db.MongoDB) *HealthHandler {
	return &HealthHandler{DB: database}
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service,omitempty"`
	Message string `json:"message,omitempty"`
}

type DBHealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
	Message  string `json:"message,omitempty"`
}

// Health returns basic service health status
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:  "ok",
		Service: "backend",
	}

	json.NewEncoder(w).Encode(response)
}

// DBHealth checks database connection
func (h *HealthHandler) DBHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		response := DBHealthResponse{
			Status:  "error",
			Message: "Database connection failed: " + err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := DBHealthResponse{
		Status:   "ok",
		Database: "connected",
	}
	json.NewEncoder(w).Encode(response)
}
