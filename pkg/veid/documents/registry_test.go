package documents_test

import (
	"context"
	"errors"
	"testing"

	"github.com/virtengine/virtengine/pkg/veid/documents"
	"github.com/virtengine/virtengine/pkg/veid/documents/adapters"
)

const passportMRZ = "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10"

func TestRegistryBindsExtractedMRZToRequestedDocumentIdentity(t *testing.T) {
	registry := documents.NewRegistry(adapters.NewMRZAdapter([]documents.CountryCode{"UTO", "USA"}))
	data, err := registry.Extract(context.Background(), documents.DocumentTypePassport, "UTO", nil, passportMRZ)
	if err != nil {
		t.Fatalf("extract matching MRZ: %v", err)
	}
	if data.DocumentNumber != "L898902C3" {
		t.Fatalf("unexpected document %q", data.DocumentNumber)
	}
	_, err = registry.Extract(context.Background(), documents.DocumentTypePassport, "USA", nil, passportMRZ)
	if !errors.Is(err, documents.ErrInvalidDocument) {
		t.Fatalf("expected country binding rejection, got %v", err)
	}
	_, err = registry.Extract(context.Background(), documents.DocumentTypeIDCard, "UTO", nil, passportMRZ)
	if !errors.Is(err, documents.ErrInvalidDocument) {
		t.Fatalf("expected type binding rejection, got %v", err)
	}
}
