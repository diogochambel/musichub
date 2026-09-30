package models

import (
	"strings"
	"testing"
	"time"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid letters", "john", false},
		{"valid numbers", "user123", false},
		{"valid mixed", "john123", false},
		{"empty", "", true},
		{"spaces", "john doe", true},
		{"special chars", "john@doe", true},
		{"underscores", "john_doe", true},
		{"hyphens", "john-doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"valid with dots", "user.name@example.co.uk", false},
		{"valid with plus", "user+tag@example.com", false},
		{"empty", "", true},
		{"no domain", "user@", true},
		{"no at", "userexample.com", true},
		{"no tld", "user@example", true},
		{"spaces", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Password1", false},
		{"valid longer", "MySecure123!", false},
		{"empty", "", true},
		{"too short", "Pass1", true},
		{"no uppercase", "password1", true},
		{"no lowercase", "PASSWORD1", true},
		{"no number", "Password", true},
		{"exactly 8", "Passwor1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBirthDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   time.Time
		wantErr bool
	}{
		{"valid 14", now.AddDate(-14, 0, 0), false},
		{"valid adult", now.AddDate(-30, 0, 0), false},
		{"boundary 13", now.AddDate(-13, 0, 0), false},
		{"under 13", now.AddDate(-12, 0, 0), true},
		{"future date", now.AddDate(1, 0, 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBirthDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBirthDate(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateRequest(t *testing.T) {
	validReq := &CreateUserRequest{
		Username:  "john123",
		Email:     "john@example.com",
		Password:  "Password1",
		BirthDate: time.Now().AddDate(-20, 0, 0),
	}
	if err := ValidateCreateRequest(validReq); err != nil {
		t.Errorf("ValidateCreateRequest(valid) = %v, want nil", err)
	}

	invalidReq := &CreateUserRequest{
		Username:  "jo",
		Email:     "invalid",
		Password:  "short",
		BirthDate: time.Now().AddDate(-10, 0, 0),
	}
	if err := ValidateCreateRequest(invalidReq); err == nil {
		t.Error("ValidateCreateRequest(invalid) = nil, want error")
	}
}

func TestHashPassword(t *testing.T) {
	salt := "testsalt123"
	password := "Password1"

	hash1 := HashPassword(password, salt)
	hash2 := HashPassword(password, salt)

	if hash1 != hash2 {
		t.Error("HashPassword should be deterministic")
	}

	if len(hash1) != 64 {
		t.Errorf("HashPassword hash length = %d, want 64", len(hash1))
	}

	differentSalt := "othersalt456"
	hash3 := HashPassword(password, differentSalt)
	if hash1 == hash3 {
		t.Error("Different salts should produce different hashes")
	}

	differentPassword := "Password2"
	hash4 := HashPassword(differentPassword, salt)
	if hash1 == hash4 {
		t.Error("Different passwords should produce different hashes")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}

	if len(salt1) != 64 {
		t.Errorf("GenerateSalt() length = %d, want 64 (hex of 32 bytes)", len(salt1))
	}

	if strings.TrimSpace(salt1) != salt1 {
		t.Error("GenerateSalt() should not contain whitespace")
	}

	salt2, _ := GenerateSalt()
	if salt1 == salt2 {
		t.Error("Two generated salts should not be equal")
	}
}
