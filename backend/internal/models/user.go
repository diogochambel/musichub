package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID               primitive.ObjectID  `json:"id" bson:"_id"`
	Username         string              `json:"username" bson:"username"`
	Email            string              `json:"email" bson:"email"`
	Password         string              `json:"-" bson:"password"`
	Salt             string              `json:"-" bson:"salt"`
	BirthDate        time.Time           `bson:"birth_date" json:"birth_date" `
	FavoriteArtistID *primitive.ObjectID `json:"favorite_artist_id,omitempty" bson:"favorite_artist_id,omitempty"`
}

type SetFavoriteArtistRequest struct {
	ArtistID string `json:"artist_id"`
}

type CreateUserRequest struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	BirthDate time.Time `json:"birth_date"`
}

type UpdateUserRequest struct {
	Username        *string    `json:"username"`
	Email           *string    `json:"email"`
	Password        *string    `json:"password"`
	CurrentPassword *string    `json:"current_password"`
	BirthDate       *time.Time `json:"birth_date,omitempty"`
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username must contain only letters and numbers")
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		if unicode.IsUpper(c) {
			hasUpper = true
		}
		if unicode.IsLower(c) {
			hasLower = true
		}
		if unicode.IsDigit(c) {
			hasNumber = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	return nil
}

func ValidateBirthDate(birthDate time.Time) error {
	now := time.Now()

	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	if age < 13 {
		return errors.New("user must be at least 13 years old")
	}

	return nil
}

func ValidateCreateRequest(req *CreateUserRequest) error {
	if err := ValidateUsername(req.Username); err != nil {
		return err
	}
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return err
	}
	if err := ValidateBirthDate(req.BirthDate); err != nil {
		return err
	}
	return nil
}

func GenerateSalt() (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	return hex.EncodeToString(salt), nil
}

func HashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + password))
	return hex.EncodeToString(h.Sum(nil))
}
