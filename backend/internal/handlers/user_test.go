package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var testBirthDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

type MockUserRepository struct {
	users     map[primitive.ObjectID]*models.User
	nextID    primitive.ObjectID
	insertErr error
	updateErr error
	deleteErr error
}

type MockArtistRepository struct {
	artists map[primitive.ObjectID]*models.Artist
}

func NewMockArtistRepository() *MockArtistRepository {
	return &MockArtistRepository{
		artists: make(map[primitive.ObjectID]*models.Artist),
	}
}

func (m *MockArtistRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Artist, error) {
	artist, ok := m.artists[id]
	if !ok {
		return nil, fmt.Errorf("artist not found")
	}
	return artist, nil
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[primitive.ObjectID]*models.User),
		nextID: primitive.NewObjectID(),
	}
}

func (m *MockUserRepository) Insert(ctx context.Context, user *models.User) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	result := make([]*models.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *MockUserRepository) Update(ctx context.Context, id primitive.ObjectID, user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[id] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.users[id]; !ok {
		return fmt.Errorf("no documents in result")
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) SetFavoriteArtistID(ctx context.Context, id primitive.ObjectID, artistID primitive.ObjectID) error {
	user, ok := m.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.FavoriteArtistID = &artistID
	return nil
}

func (m *MockUserRepository) RemoveFavoriteArtistID(ctx context.Context, id primitive.ObjectID) error {
	user, ok := m.users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	user.FavoriteArtistID = nil
	return nil
}

func setupHandler() (*UserHandler, *MockUserRepository, *MockArtistRepository) {
	userRepo := NewMockUserRepository()
	artistRepo := NewMockArtistRepository()
	handler := NewUserHandler(userRepo, artistRepo)
	return handler, userRepo, artistRepo
}

func withAuthContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), "user_data", &middleware.TokenData{
		UserID: userID,
	})
	return req.WithContext(ctx)
}

func TestCreateUser_Success(t *testing.T) {
	handler, _, _ := setupHandler()

	body := `{"username":"john123","email":"john@example.com","password":"Password1","birth_date":"2000-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)

	if response["username"] != "john123" {
		t.Errorf("expected username john123, got %v", response["username"])
	}
	if response["email"] != "john@example.com" {
		t.Errorf("expected email, got %v", response["email"])
	}
	if _, ok := response["password"]; ok {
		t.Error("password should not be in response")
	}
	if _, ok := response["salt"]; ok {
		t.Error("salt should not be in response")
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	handler, _, _ := setupHandler()

	existingUser := &models.User{
		ID:        primitive.NewObjectID(),
		Username:  "john123",
		Email:     "other@example.com",
		Password:  "hash",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), existingUser)

	body := `{"username":"john123","email":"new@example.com","password":"Password1","birth_date":"2000-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUser(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	handler, _, _ := setupHandler()

	existingUser := &models.User{
		ID:        primitive.NewObjectID(),
		Username:  "other",
		Email:     "john@example.com",
		Password:  "hash",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), existingUser)

	body := `{"username":"newuser","email":"john@example.com","password":"Password1","birth_date":"2000-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUser(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestCreateUser_InvalidBody(t *testing.T) {
	handler, _, _ := setupHandler()

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			"empty username",
			`{"username":"","email":"a@b.com","password":"Password1","birth_date":"2000-01-01T00:00:00Z"}`,
			"username",
		},
		{
			"short password",
			`{"username":"john","email":"a@b.com","password":"Ab1","birth_date":"2000-01-01T00:00:00Z"}`,
			"password",
		},
		{
			"invalid birth date",
			`{"username":"john","email":"a@b.com","password":"Password1","birth_date":"2020-01-01T00:00:00Z"}`,
			"birth_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _ := setupHandler()
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateUser(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			var response ErrorResponse
			json.NewDecoder(w.Body).Decode(&response)
			if response.Status != "error" {
				t.Errorf("expected error status, got %s", response.Status)
			}
		})
	}
}

func TestGetUser_Success(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "hashed",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	handler, _, _ := setupHandler()

	fakeID := primitive.NewObjectID()
	req := httptest.NewRequest(http.MethodGet, "/users/"+fakeID.Hex(), nil)
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	handler, _, _ := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
	w := httptest.NewRecorder()

	handler.GetUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListUsers(t *testing.T) {
	handler, _, _ := setupHandler()

	user1 := &models.User{
		ID:        primitive.NewObjectID(),
		Username:  "user1",
		Email:     "user1@example.com",
		Password:  "hash1",
		Salt:      "salt1",
		BirthDate: testBirthDate,
	}
	user2 := &models.User{
		ID:        primitive.NewObjectID(),
		Username:  "user2",
		Email:     "user2@example.com",
		Password:  "hash2",
		Salt:      "salt2",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user1)
	handler.Repo.Insert(context.Background(), user2)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []any
	json.NewDecoder(w.Body).Decode(&response)
	if len(response) != 2 {
		t.Errorf("expected 2 users, got %d", len(response))
	}
}

func TestListUsers_Empty(t *testing.T) {
	handler, _, _ := setupHandler()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []any
	json.NewDecoder(w.Body).Decode(&response)
	if len(response) != 0 {
		t.Errorf("expected 0 users, got %d", len(response))
	}
}

func TestUpdateUser_Success(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  models.HashPassword("Password1", "salt"),
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	body := `{"username":"johnupdated","current_password":"Password1"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)
	if response["username"] != "johnupdated" {
		t.Errorf("expected username johnupdated, got %v", response["username"])
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	handler, _, _ := setupHandler()

	fakeID := primitive.NewObjectID()
	body := `{"username":"johnupdated"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+fakeID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateUser_MissingCurrentPassword(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  models.HashPassword("Password1", "salt"),
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	body := `{"username":"johnupdated"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateUser_IncorrectCurrentPassword(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  models.HashPassword("Password1", "salt"),
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	body := `{"username":"johnupdated","current_password":"WrongPassword"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUpdateUser_PasswordRehash(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  models.HashPassword("Password1", "oldsalt"),
		Salt:      "oldsalt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	body := `{"password":"NewPassword1","current_password":"Password1"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	updatedUser := handler.Repo.(*MockUserRepository).users[userID]
	if updatedUser.Password == "oldhash" {
		t.Error("password should have been rehashed")
	}
	if updatedUser.Salt == "oldsalt" {
		t.Error("salt should have been regenerated")
	}
}

func TestDeleteUser_Success(t *testing.T) {
	handler, _, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "hashed",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	handler.Repo.Insert(context.Background(), user)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	handler, _, _ := setupHandler()

	fakeID := primitive.NewObjectID()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+fakeID.Hex(), nil)
	w := httptest.NewRecorder()

	handler.DeleteUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSetFavoriteArtist_Success(t *testing.T) {
	handler, userRepo, artistRepo := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "hashed",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	userRepo.Insert(context.Background(), user)

	artistID := primitive.NewObjectID()
	artist := &models.Artist{
		ID:   artistID,
		ISNI: "1234567890123456",
		Name: "Test Artist",
	}
	artistRepo.artists[artistID] = artist

	body := fmt.Sprintf(`{"artist_id":"%s"}`, artistID.Hex())
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex()+"/favorite-artist", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.SetFavoriteArtist(w, req, userID.Hex())

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)
	if response["favorite_artist_id"] == nil {
		t.Error("expected favorite_artist_id to be set")
	}
}

func TestSetFavoriteArtist_UserNotFound(t *testing.T) {
	handler, _, _ := setupHandler()

	fakeID := primitive.NewObjectID()
	artistID := primitive.NewObjectID()

	body := fmt.Sprintf(`{"artist_id":"%s"}`, artistID.Hex())
	req := httptest.NewRequest(http.MethodPut, "/users/"+fakeID.Hex()+"/favorite-artist", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, fakeID.Hex())
	w := httptest.NewRecorder()

	handler.SetFavoriteArtist(w, req, fakeID.Hex())

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSetFavoriteArtist_ArtistNotFound(t *testing.T) {
	handler, userRepo, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "hashed",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	userRepo.Insert(context.Background(), user)

	fakeArtistID := primitive.NewObjectID()
	body := fmt.Sprintf(`{"artist_id":"%s"}`, fakeArtistID.Hex())
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex()+"/favorite-artist", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.SetFavoriteArtist(w, req, userID.Hex())

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSetFavoriteArtist_InvalidArtistID(t *testing.T) {
	handler, userRepo, _ := setupHandler()

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:        userID,
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "hashed",
		Salt:      "salt",
		BirthDate: testBirthDate,
	}
	userRepo.Insert(context.Background(), user)

	body := `{"artist_id":"invalid"}`
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.Hex()+"/favorite-artist", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.SetFavoriteArtist(w, req, userID.Hex())

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveFavoriteArtist_Success(t *testing.T) {
	handler, userRepo, artistRepo := setupHandler()

	artistID := primitive.NewObjectID()
	artist := &models.Artist{
		ID:   artistID,
		ISNI: "1234567890123456",
		Name: "Test Artist",
	}
	artistRepo.artists[artistID] = artist

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:               userID,
		Username:         "john123",
		Email:            "john@example.com",
		Password:         "hashed",
		Salt:             "salt",
		BirthDate:        testBirthDate,
		FavoriteArtistID: &artistID,
	}
	userRepo.Insert(context.Background(), user)

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID.Hex()+"/favorite-artist", nil)
	req = withAuthContext(req, userID.Hex())
	w := httptest.NewRecorder()

	handler.RemoveFavoriteArtist(w, req, userID.Hex())

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)
	if response["favorite_artist_id"] != nil {
		t.Error("expected favorite_artist_id to be nil after removal")
	}
}

func TestRemoveFavoriteArtist_UserNotFound(t *testing.T) {
	handler, _, _ := setupHandler()

	fakeID := primitive.NewObjectID()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+fakeID.Hex()+"/favorite-artist", nil)
	req = withAuthContext(req, fakeID.Hex())
	w := httptest.NewRecorder()

	handler.RemoveFavoriteArtist(w, req, fakeID.Hex())

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
