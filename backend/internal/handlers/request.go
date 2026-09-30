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
)

type VersionRequestRepository interface {
	Insert(ctx context.Context, req *models.VersionRequest) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.VersionRequest, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.VersionRequest, error)
	FindByUserIDAndStatus(ctx context.Context, userID primitive.ObjectID, status models.RequestStatus) ([]*models.VersionRequest, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.RequestStatus) error
}

type NotificationRepository interface {
	Insert(ctx context.Context, n *models.Notification) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Notification, error)
	CountUnread(ctx context.Context, userID primitive.ObjectID) (int64, error)
	MarkAllRead(ctx context.Context, userID primitive.ObjectID) error
	MarkRead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

type RequestAlbumFinder interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error)
	Update(ctx context.Context, id primitive.ObjectID, album *models.Album) error
}

type RequestHandler struct {
	RequestRepo      VersionRequestRepository
	NotificationRepo NotificationRepository
	AlbumRepo        RequestAlbumFinder
}

func NewRequestHandler(requestRepo VersionRequestRepository, notificationRepo NotificationRepository, albumRepo RequestAlbumFinder) *RequestHandler {
	return &RequestHandler{
		RequestRepo:      requestRepo,
		NotificationRepo: notificationRepo,
		AlbumRepo:        albumRepo,
	}
}

func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body models.CreateVersionRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := models.ValidateCreateVersionRequestRequest(&body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	albumID, err := primitive.ObjectIDFromHex(body.AlbumID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid album_id")
		return
	}

	_, err = h.AlbumRepo.FindByID(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	req := &models.VersionRequest{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		AlbumID:     albumID,
		EAN13:       body.EAN13,
		Support:     body.Support,
		VersionName: body.VersionName,
		Status:      models.RequestStatusPending,
		RequestedAt: time.Now(),
	}

	if err := h.RequestRepo.Insert(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create request")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "request submitted successfully",
		"request": req,
	})
}

func (h *RequestHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	statusFilter := r.URL.Query().Get("status")

	var requests []*models.VersionRequest

	if statusFilter != "" {
		status := models.RequestStatus(statusFilter)
		if err := models.ValidateRequestStatus(status); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		requests, err = h.RequestRepo.FindByUserIDAndStatus(r.Context(), userID, status)
	} else {
		requests, err = h.RequestRepo.FindByUserID(r.Context(), userID)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch requests")
		return
	}

	type RequestWithAlbum struct {
		ID          primitive.ObjectID   `json:"id"`
		AlbumID     primitive.ObjectID   `json:"album_id"`
		AlbumTitle  string               `json:"album_title"`
		EAN13       string               `json:"ean13"`
		Support     models.SupportType   `json:"support"`
		VersionName string               `json:"version_name"`
		Status      models.RequestStatus `json:"status"`
		RequestedAt time.Time            `json:"requested_at"`
	}

	result := make([]RequestWithAlbum, 0, len(requests))
	for _, req := range requests {
		albumTitle := "Unknown album"
		album, err := h.AlbumRepo.FindByID(r.Context(), req.AlbumID)
		if err == nil {
			albumTitle = album.Title
		}

		result = append(result, RequestWithAlbum{
			ID:          req.ID,
			AlbumID:     req.AlbumID,
			AlbumTitle:  albumTitle,
			EAN13:       req.EAN13,
			Support:     req.Support,
			VersionName: req.VersionName,
			Status:      req.Status,
			RequestedAt: req.RequestedAt,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *RequestHandler) RespondToRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/requests/")
	path = strings.TrimSuffix(path, "/respond")

	requestID, err := primitive.ObjectIDFromHex(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	newStatus := models.RequestStatus(body.Status)
	if newStatus != models.RequestStatusAccepted && newStatus != models.RequestStatusRejected {
		writeError(w, http.StatusBadRequest, "status must be 'aceite' or 'recusado'")
		return
	}

	req, err := h.RequestRepo.FindByID(r.Context(), requestID)
	if err != nil {
		writeError(w, http.StatusNotFound, "request not found")
		return
	}

	if req.Status != models.RequestStatusPending {
		writeError(w, http.StatusConflict, "request already has a final status")
		return
	}

	album, err := h.AlbumRepo.FindByID(r.Context(), req.AlbumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "album not found")
		return
	}

	if err := h.RequestRepo.UpdateStatus(r.Context(), requestID, newStatus); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update request")
		return
	}

	if newStatus == models.RequestStatusAccepted {
		newRelease := models.AlbumRelease{
			Support:     req.Support,
			VersionName: req.VersionName,
			EAN13:       req.EAN13,
		}

		album.Releases = append(album.Releases, newRelease)

		if err := h.AlbumRepo.Update(r.Context(), album.ID, album); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to add album version")
			return
		}
	}

	notification := &models.Notification{
		ID:         primitive.NewObjectID(),
		UserID:     req.UserID,
		RequestID:  req.ID,
		AlbumTitle: album.Title,
		Response:   string(newStatus),
		Read:       false,
		CreatedAt:  time.Now(),
	}

	if err := h.NotificationRepo.Insert(r.Context(), notification); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create notification")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "success",
		"message":      "request updated and notification sent",
		"notification": notification,
	})
}

func (h *RequestHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	notifications, err := h.NotificationRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch notifications")
		return
	}

	writeJSON(w, http.StatusOK, notifications)
}

func (h *RequestHandler) CountUnreadNotifications(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	count, err := h.NotificationRepo.CountUnread(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{
		"unread": count,
	})
}

func (h *RequestHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	if err := h.NotificationRepo.MarkAllRead(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark notifications as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "notifications marked as read",
	})
}

func (h *RequestHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userData := middleware.GetUserData(r)
	if userData == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userData.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user session")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/notifications/")
	path = strings.TrimSuffix(path, "/read")

	notificationID, err := primitive.ObjectIDFromHex(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.NotificationRepo.MarkRead(r.Context(), notificationID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "notification marked as read",
	})
}
