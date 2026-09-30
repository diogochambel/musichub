package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func contextWithTestData(req *http.Request, data *TokenData) *http.Request {
	ctx := context.WithValue(req.Context(), "user_data", data)
	return req.WithContext(ctx)
}

func TestCreateToken_Success(t *testing.T) {
	token, err := CreateToken("user123", "john")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("expected token format data.signature, got %d parts", len(parts))
	}

	dataJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("expected valid base64 payload, got error %v", err)
	}

	var data TokenData
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		t.Fatalf("expected valid JSON payload, got error %v", err)
	}

	if data.UserID != "user123" {
		t.Errorf("expected user_id user123, got %s", data.UserID)
	}
	if data.Username != "john" {
		t.Errorf("expected username john, got %s", data.Username)
	}
}

func TestCreateToken_ExpirationInFuture(t *testing.T) {
	beforeCreate := time.Now().Add(23*time.Hour + 55*time.Minute).Unix()
	token, err := CreateToken("user123", "john")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	parts := strings.Split(token, ".")
	var data TokenData
	dataJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	json.Unmarshal(dataJSON, &data)

	if data.Expires <= beforeCreate {
		t.Errorf("expected expiration in the future, got %d (before threshold %d)", data.Expires, beforeCreate)
	}

	maxExpiry := time.Now().Add(24*time.Hour + 1*time.Minute).Unix()
	if data.Expires > maxExpiry {
		t.Errorf("expected expiration within ~24h, got %d", data.Expires)
	}
}

func TestVerifyToken_Success(t *testing.T) {
	token, _ := CreateToken("user123", "john")

	data, err := VerifyToken(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if data.UserID != "user123" {
		t.Errorf("expected user_id user123, got %s", data.UserID)
	}
	if data.Username != "john" {
		t.Errorf("expected username john, got %s", data.Username)
	}
}

func TestVerifyToken_InvalidFormat(t *testing.T) {
	_, err := VerifyToken("notokenformat")
	if err == nil {
		t.Error("expected error for token without dot separator, got nil")
	}
}

func TestVerifyToken_TamperedSignature(t *testing.T) {
	token, _ := CreateToken("user123", "john")
	parts := strings.Split(token, ".")

	tampered := parts[0] + "." + "tamperedsignature12345"
	_, err := VerifyToken(tampered)
	if err == nil {
		t.Error("expected error for tampered signature, got nil")
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	data := TokenData{
		UserID:   "user123",
		Username: "john",
		Expires:  time.Now().Add(-1 * time.Hour).Unix(),
	}
	dataJSON, _ := json.Marshal(data)
	dataB64 := base64.RawURLEncoding.EncodeToString(dataJSON)
	signature := createSignature(dataB64)
	token := dataB64 + "." + signature

	_, err := VerifyToken(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
	if !strings.Contains(err.Error(), "session expired") {
		t.Errorf("expected 'session expired' error, got %v", err)
	}
}

func TestVerifyToken_InvalidBase64(t *testing.T) {
	signature := createSignature("!!!invalid!!!")
	token := "!!!invalid!!!" + "." + signature

	_, err := VerifyToken(token)
	if err == nil {
		t.Error("expected error for invalid base64, got nil")
	}
}

func TestVerifyToken_InvalidJSON(t *testing.T) {
	dataB64 := base64.RawURLEncoding.EncodeToString([]byte("not json"))
	signature := createSignature(dataB64)
	token := dataB64 + "." + signature

	_, err := VerifyToken(token)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestRequireAuth_Success(t *testing.T) {
	token, _ := CreateToken("user123", "john")

	called := false
	var receivedData *TokenData

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		receivedData = GetUserData(r)
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Error("expected handler to be called")
	}
	if receivedData == nil {
		t.Fatal("expected user data in context, got nil")
	}
	if receivedData.UserID != "user123" {
		t.Errorf("expected user_id user123, got %s", receivedData.UserID)
	}
	if receivedData.Username != "john" {
		t.Errorf("expected username john, got %s", receivedData.Username)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without auth header")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuth_InvalidFormat_NoBearer(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid auth format")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token sometoken")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuth_InvalidFormat_NoToken(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token value")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	data := TokenData{
		UserID:   "user123",
		Username: "john",
		Expires:  time.Now().Add(-1 * time.Hour).Unix(),
	}
	dataJSON, _ := json.Marshal(data)
	dataB64 := base64.RawURLEncoding.EncodeToString(dataJSON)
	signature := createSignature(dataB64)
	token := dataB64 + "." + signature

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with expired token")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "session expired") {
		t.Errorf("expected 'session expired' in response body, got %s", body)
	}
}

func TestGetUserData_Present(t *testing.T) {
	data := &TokenData{UserID: "user123", Username: "john"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = contextWithTestData(req, data)

	result := GetUserData(req)
	if result == nil {
		t.Fatal("expected TokenData, got nil")
	}
	if result.UserID != "user123" {
		t.Errorf("expected user_id user123, got %s", result.UserID)
	}
	if result.Username != "john" {
		t.Errorf("expected username john, got %s", result.Username)
	}
}

func TestGetUserData_Absent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	result := GetUserData(req)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestCreateSignature_Deterministic(t *testing.T) {
	sig1 := createSignature("testdata")
	sig2 := createSignature("testdata")

	if sig1 != sig2 {
		t.Errorf("expected same signature for same input, got %s and %s", sig1, sig2)
	}
}

func TestCreateSignature_DifferentInputs(t *testing.T) {
	sig1 := createSignature("data1")
	sig2 := createSignature("data2")

	if sig1 == sig2 {
		t.Error("expected different signatures for different inputs")
	}
}

func TestRequireAuth_ErrorResponseFormat(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without auth header")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Errorf("expected JSON with status error, got %s", body)
	}
	if !strings.Contains(body, `"message"`) {
		t.Errorf("expected JSON with message field, got %s", body)
	}
}

func TestRequireAuth_ContextPropagation(t *testing.T) {
	token, _ := CreateToken("user123", "john")

	var capturedUserID string
	var capturedUsername string

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := GetUserData(r)
		if data != nil {
			capturedUserID = data.UserID
			capturedUsername = data.Username
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "ok")
	})

	protectedHandler := RequireAuth(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	if capturedUserID != "user123" {
		t.Errorf("expected user_id user123 in context, got %s", capturedUserID)
	}
	if capturedUsername != "john" {
		t.Errorf("expected username john in context, got %s", capturedUsername)
	}
}
