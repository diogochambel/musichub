package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockCustomListRepo struct {
	lists    map[primitive.ObjectID]*models.CustomList
	updateErr error
}

type MockCustomListItemRepo struct {
	items     map[string]*models.CustomListItem
	insertErr error
}

type MockListAlbumRepo struct {
	albums map[primitive.ObjectID]*models.Album
}

func NewMockCustomListRepo() *MockCustomListRepo {
	return &MockCustomListRepo{
		lists: make(map[primitive.ObjectID]*models.CustomList),
	}
}

func NewMockCustomListItemRepo() *MockCustomListItemRepo {
	return &MockCustomListItemRepo{
		items: make(map[string]*models.CustomListItem),
	}
}

func NewMockListAlbumRepo() *MockListAlbumRepo {
	return &MockListAlbumRepo{
		albums: make(map[primitive.ObjectID]*models.Album),
	}
}

func (m *MockCustomListRepo) Insert(ctx context.Context, list *models.CustomList) error {
	m.lists[list.ID] = list
	return nil
}

func (m *MockCustomListRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.CustomList, error) {
	list, ok := m.lists[id]
	if !ok {
		return nil, fmt.Errorf("list not found")
	}
	return list, nil
}

func (m *MockCustomListRepo) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.CustomList, error) {
	var result []*models.CustomList
	for _, l := range m.lists {
		if l.UserID == userID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *MockCustomListRepo) FindByUserAndName(ctx context.Context, userID primitive.ObjectID, name string) (*models.CustomList, error) {
	for _, l := range m.lists {
		if l.UserID == userID && l.Name == name {
			return l, nil
		}
	}
	return nil, fmt.Errorf("list not found")
}

func (m *MockCustomListRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	if _, ok := m.lists[id]; !ok {
		return fmt.Errorf("list not found")
	}
	delete(m.lists, id)
	return nil
}

func (m *MockCustomListRepo) Update(ctx context.Context, list *models.CustomList) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.lists[list.ID] = list
	return nil
}

func (m *MockCustomListItemRepo) Insert(ctx context.Context, item *models.CustomListItem) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	key := item.ListID.Hex() + "_" + item.AlbumID.Hex()
	if _, exists := m.items[key]; exists {
		return fmt.Errorf("duplicate key error")
	}
	m.items[key] = item
	return nil
}

func (m *MockCustomListItemRepo) FindByListID(ctx context.Context, listID primitive.ObjectID) ([]*models.CustomListItem, error) {
	var result []*models.CustomListItem
	for _, item := range m.items {
		if item.ListID == listID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *MockCustomListItemRepo) CountByListID(ctx context.Context, listID primitive.ObjectID) (int64, error) {
	var count int64
	for _, item := range m.items {
		if item.ListID == listID {
			count++
		}
	}
	return count, nil
}

func (m *MockCustomListItemRepo) DeleteByListID(ctx context.Context, listID primitive.ObjectID) error {
	for key, item := range m.items {
		if item.ListID == listID {
			delete(m.items, key)
		}
	}
	return nil
}

func (m *MockCustomListItemRepo) DeleteByListIDAndAlbumID(ctx context.Context, listID, albumID primitive.ObjectID) error {
	key := listID.Hex() + "_" + albumID.Hex()
	if _, ok := m.items[key]; !ok {
		return fmt.Errorf("no documents in result")
	}
	delete(m.items, key)
	return nil
}

func (m *MockListAlbumRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Album, error) {
	album, ok := m.albums[id]
	if !ok {
		return nil, fmt.Errorf("album not found")
	}
	return album, nil
}

func setupCustomListHandler() (*CustomListHandler, *MockCustomListRepo, *MockCustomListItemRepo, *MockListAlbumRepo) {
	listRepo := NewMockCustomListRepo()
	itemRepo := NewMockCustomListItemRepo()
	albumRepo := NewMockListAlbumRepo()
	handler := NewCustomListHandler(listRepo, itemRepo, albumRepo)
	return handler, listRepo, itemRepo, albumRepo
}

func TestAddItem_Success(t *testing.T) {
	handler, listRepo, _, albumRepo := setupCustomListHandler()

	userID := primitive.NewObjectID()
	albumID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}
	albumRepo.albums[albumID] = &models.Album{
		ID:    albumID,
		Title: "Test Album",
	}

	body := fmt.Sprintf(`{"album_id":"%s"}`, albumID.Hex())
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp models.CustomListItem
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ListID != listID {
		t.Errorf("expected list_id %s, got %s", listID.Hex(), resp.ListID.Hex())
	}
	if resp.AlbumID != albumID {
		t.Errorf("expected album_id %s, got %s", albumID.Hex(), resp.AlbumID.Hex())
	}
}

func TestAddItem_Unauthenticated(t *testing.T) {
	handler, _, _, _ := setupCustomListHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/lists/123/items", nil)
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAddItem_InvalidListID(t *testing.T) {
	handler, _, _, _ := setupCustomListHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/lists/invalid/items", nil)
	req = withAuthContext(req, primitive.NewObjectID().Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddItem_ListNotFound(t *testing.T) {
	handler, _, _, _ := setupCustomListHandler()

	unknownID := primitive.NewObjectID()
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+unknownID.Hex()+"/items", nil)
	req = withAuthContext(req, primitive.NewObjectID().Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAddItem_Forbidden(t *testing.T) {
	handler, listRepo, _, albumRepo := setupCustomListHandler()

	ownerID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: ownerID,
		Name:   "My List",
	}

	albumID := primitive.NewObjectID()
	albumRepo.albums[albumID] = &models.Album{
		ID:    albumID,
		Title: "Test Album",
	}

	body := fmt.Sprintf(`{"album_id":"%s"}`, albumID.Hex())
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, otherUserID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestAddItem_MissingAlbumID(t *testing.T) {
	handler, listRepo, _, _ := setupCustomListHandler()

	userID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}

	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddItem_InvalidAlbumID(t *testing.T) {
	handler, listRepo, _, _ := setupCustomListHandler()

	userID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}

	body := `{"album_id":"not-a-hex-string"}`
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddItem_AlbumNotFound(t *testing.T) {
	handler, listRepo, _, _ := setupCustomListHandler()

	userID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}

	unknownAlbumID := primitive.NewObjectID()
	body := fmt.Sprintf(`{"album_id":"%s"}`, unknownAlbumID.Hex())
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestAddItem_DuplicateAlbum(t *testing.T) {
	handler, listRepo, itemRepo, albumRepo := setupCustomListHandler()

	userID := primitive.NewObjectID()
	albumID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}
	albumRepo.albums[albumID] = &models.Album{
		ID:    albumID,
		Title: "Test Album",
	}

	key := listID.Hex() + "_" + albumID.Hex()
	itemRepo.items[key] = &models.CustomListItem{
		ListID:  listID,
		AlbumID: albumID,
	}

	body := fmt.Sprintf(`{"album_id":"%s"}`, albumID.Hex())
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestAddItem_DatabaseError(t *testing.T) {
	handler, listRepo, itemRepo, albumRepo := setupCustomListHandler()

	userID := primitive.NewObjectID()
	albumID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}
	albumRepo.albums[albumID] = &models.Album{
		ID:    albumID,
		Title: "Test Album",
	}

	itemRepo.insertErr = fmt.Errorf("connection lost")

	body := fmt.Sprintf(`{"album_id":"%s"}`, albumID.Hex())
	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestAddItem_InvalidJSON(t *testing.T) {
	handler, listRepo, _, _ := setupCustomListHandler()

	userID := primitive.NewObjectID()
	listID := primitive.NewObjectID()

	listRepo.lists[listID] = &models.CustomList{
		ID:     listID,
		UserID: userID,
		Name:   "My List",
	}

	req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.Hex()+"/items", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.addItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
