package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateCreateVersionRequestRequest(t *testing.T) {
	validReq := &CreateVersionRequestRequest{
		AlbumID:     primitive.NewObjectID().Hex(),
		EAN13:       "1234567890123",
		Support:     SupportTypeCD,
		VersionName: "Deluxe Edition",
	}
	if err := ValidateCreateVersionRequestRequest(validReq); err != nil {
		t.Errorf("ValidateCreateVersionRequestRequest(valid) = %v, want nil", err)
	}

	missingAlbum := &CreateVersionRequestRequest{
		AlbumID: "",
		EAN13:   "1234567890123",
		Support: SupportTypeCD,
	}
	if err := ValidateCreateVersionRequestRequest(missingAlbum); err == nil {
		t.Error("ValidateCreateVersionRequestRequest(missing album_id) = nil, want error")
	}

	invalidSupport := &CreateVersionRequestRequest{
		AlbumID: primitive.NewObjectID().Hex(),
		EAN13:   "1234567890123",
		Support: "digital",
	}
	if err := ValidateCreateVersionRequestRequest(invalidSupport); err == nil {
		t.Error("ValidateCreateVersionRequestRequest(invalid support) = nil, want error")
	}

	invalidEAN := &CreateVersionRequestRequest{
		AlbumID: primitive.NewObjectID().Hex(),
		EAN13:   "short",
		Support: SupportTypeCassette,
	}
	if err := ValidateCreateVersionRequestRequest(invalidEAN); err == nil {
		t.Error("ValidateCreateVersionRequestRequest(invalid EAN13) = nil, want error")
	}
}

func TestValidateRequestStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  RequestStatus
		wantErr bool
	}{
		{"pending", RequestStatusPending, false},
		{"accepted", RequestStatusAccepted, false},
		{"rejected", RequestStatusRejected, false},
		{"invalid", "em progresso", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequestStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequestStatus(%q) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}
