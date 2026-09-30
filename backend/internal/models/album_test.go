package models

import "testing"

func TestValidateMBID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "12345678-1234-1234-1234-123456789012", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMBID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMBID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAlbumTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Infest", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlbumTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAlbumTitle(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAlbumType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"single", "single", false},
		{"EP", "EP", false},
		{"LP", "LP", false},
		{"empty", "", true},
		{"invalid", "album", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlbumType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAlbumType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEAN13(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "1234567890123", false},
		{"empty", "", true},
		{"too short", "123456789012", true},
		{"too long", "12345678901234", true},
		{"non digits", "1234567890abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEAN13(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEAN13(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAlbumRelease(t *testing.T) {
	validRelease := AlbumRelease{
		Support:     SupportTypeCD,
		VersionName: "Standard Edition",
		EAN13:       "1234567890123",
	}
	if err := ValidateAlbumRelease(validRelease); err != nil {
		t.Errorf("ValidateAlbumRelease(valid) = %v, want nil", err)
	}

	invalidSupport := AlbumRelease{
		Support:     "invalid",
		VersionName: "Standard Edition",
		EAN13:       "1234567890123",
	}
	if err := ValidateAlbumRelease(invalidSupport); err == nil {
		t.Error("ValidateAlbumRelease(invalid support) = nil, want error")
	}

	invalidEAN := AlbumRelease{
		Support:     SupportTypeCD,
		VersionName: "Standard Edition",
		EAN13:       "abc",
	}
	if err := ValidateAlbumRelease(invalidEAN); err == nil {
		t.Error("ValidateAlbumRelease(invalid EAN13) = nil, want error")
	}
}

func TestValidateCreateAlbumRequest(t *testing.T) {
	validReq := &CreateAlbumRequest{
		MBID:        "12345678-1234-1234-1234-123456789012",
		Title:       "Infest",
		ReleaseYear: 2000,
		Type:        "LP",
	}
	if err := ValidateCreateAlbumRequest(validReq); err != nil {
		t.Errorf("ValidateCreateAlbumRequest(valid) = %v, want nil", err)
	}

	invalidReq := &CreateAlbumRequest{
		MBID:        "",
		Title:       "",
		ReleaseYear: 0,
		Type:        "album",
	}
	if err := ValidateCreateAlbumRequest(invalidReq); err == nil {
		t.Error("ValidateCreateAlbumRequest(invalid) = nil, want error")
	}
}
