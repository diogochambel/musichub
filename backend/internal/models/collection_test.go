package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateCreateCollectionRequest(t *testing.T) {
	validReq := &CreateCollectionRequest{
		AlbumID:     primitive.NewObjectID().Hex(),
		EAN13:       "1234567890123",
		Support:     SupportTypeCD,
		VersionName: "Standard Edition",
	}
	if err := ValidateCreateCollectionRequest(validReq); err != nil {
		t.Errorf("ValidateCreateCollectionRequest(valid) = %v, want nil", err)
	}

	missingAlbum := &CreateCollectionRequest{
		AlbumID: "",
		EAN13:   "1234567890123",
		Support: SupportTypeCD,
	}
	if err := ValidateCreateCollectionRequest(missingAlbum); err == nil {
		t.Error("ValidateCreateCollectionRequest(missing album_id) = nil, want error")
	}

	invalidSupport := &CreateCollectionRequest{
		AlbumID: primitive.NewObjectID().Hex(),
		EAN13:   "1234567890123",
		Support: "dvd",
	}
	if err := ValidateCreateCollectionRequest(invalidSupport); err == nil {
		t.Error("ValidateCreateCollectionRequest(invalid support) = nil, want error")
	}

	invalidEAN := &CreateCollectionRequest{
		AlbumID: primitive.NewObjectID().Hex(),
		EAN13:   "abc",
		Support: SupportTypeVinyl,
	}
	if err := ValidateCreateCollectionRequest(invalidEAN); err == nil {
		t.Error("ValidateCreateCollectionRequest(invalid EAN13) = nil, want error")
	}
}
