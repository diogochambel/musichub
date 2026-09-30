package models

import "testing"

func TestValidateISNI(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "1234567890123456", false},
		{"empty", "", true},
		{"too short", "12345", true},
		{"too long", "12345678901234567", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateISNI(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateISNI(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateArtistName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Papa Roach", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArtistName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArtistName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateStartYear(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid", 1993, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStartYear(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStartYear(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateArtistType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"solo", "solo", false},
		{"group", "group", false},
		{"empty", "", true},
		{"invalid", "band", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArtistType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArtistType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreateArtistRequest(t *testing.T) {
	validReq := &CreateArtistRequest{
		ISNI:      "1234567890123456",
		Name:      "Papa Roach",
		StartYear: 1993,
		Type:      "group",
	}
	if err := ValidateCreateArtistRequest(validReq); err != nil {
		t.Errorf("ValidateCreateArtistRequest(valid) = %v, want nil", err)
	}

	invalidReq := &CreateArtistRequest{
		ISNI:      "",
		Name:      "",
		StartYear: 0,
		Type:      "band",
	}
	if err := ValidateCreateArtistRequest(invalidReq); err == nil {
		t.Error("ValidateCreateArtistRequest(invalid) = nil, want error")
	}
}
