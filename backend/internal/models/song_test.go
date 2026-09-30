package models

import "testing"

func TestValidateISRC(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "USUM72000000", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateISRC(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateISRC(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSongTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Last Resort", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSongTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSongTitle(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDurationSeconds(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid", 240, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDurationSeconds(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDurationSeconds(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateSongRequest(t *testing.T) {
	validReq := &CreateSongRequest{
		ISRC:            "USUM72000000",
		Title:           "Last Resort",
		DurationSeconds: 240,
		ArtistIDs:       []string{},
	}
	if err := ValidateCreateSongRequest(validReq); err != nil {
		t.Errorf("ValidateCreateSongRequest(valid) = %v, want nil", err)
	}

	invalidReq := &CreateSongRequest{
		ISRC:            "",
		Title:           "",
		DurationSeconds: 0,
	}
	if err := ValidateCreateSongRequest(invalidReq); err == nil {
		t.Error("ValidateCreateSongRequest(invalid) = nil, want error")
	}
}
