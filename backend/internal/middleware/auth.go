package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Secret key used to sign the tokens
var secretKey = "psi_26"

// Data stored inside the token
type TokenData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Expires  int64  `json:"expires"`
}

// CreateToken creates a token for the user that just logged in
func CreateToken(userID string, username string) (string, error) {
	// Store the user data and when the token expires (24 hours)
	data := TokenData{
		UserID:   userID,
		Username: username,
		Expires:  time.Now().Add(24 * time.Hour).Unix(),
	}

	// Convert the data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error creating token")
	}

	// Convert JSON to base64 (so it can be sent in the HTTP header)
	dataB64 := base64.RawURLEncoding.EncodeToString(dataJSON)

	// Create the signature to make sure the token is not tampered with
	signature := createSignature(dataB64)

	// Final token is: data.signature
	token := dataB64 + "." + signature
	return token, nil
}

// VerifyToken checks if a token is valid and returns the user data
func VerifyToken(token string) (*TokenData, error) {
	// The token has the format: data.signature
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token")
	}

	dataB64 := parts[0]
	signature := parts[1]

	// Check if the signature is correct
	if !hmac.Equal([]byte(createSignature(dataB64)), []byte(signature)) {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert data from base64 to JSON
	dataJSON, err := base64.RawURLEncoding.DecodeString(dataB64)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Convert JSON to the TokenData struct
	var data TokenData
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if the token has not expired yet
	if time.Now().Unix() > data.Expires {
		return nil, fmt.Errorf("session expired, please login again")
	}

	return &data, nil
}

// RequireAuth is a middleware that blocks routes that require login
func RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the token from the HTTP request header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "you must be logged in to access this")
			return
		}

		// The header has the format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "invalid token format")
			return
		}

		// Check if the token is valid
		data, err := VerifyToken(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Store the user data in the context to use in the handler
		ctx := context.WithValue(r.Context(), "user_data", data)
		handler(w, r.WithContext(ctx))
	}
}

// GetUserData returns the authenticated user's data from the request context
func GetUserData(r *http.Request) *TokenData {
	data, _ := r.Context().Value("user_data").(*TokenData)
	return data
}

// createSignature creates an HMAC signature for the token data
func createSignature(data string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"status":"error","message":"%s"}`, message)
}
