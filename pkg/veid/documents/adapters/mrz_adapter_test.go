package adapters

import (
	"context"
	"testing"

	"github.com/virtengine/virtengine/pkg/veid/documents"
)

func TestMRZAdapterDoesNotInventConfidence(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"UTO"})
	data, err := adapter.ExtractWithMRZ(context.Background(), nil, "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10")
	if err != nil {
		t.Fatalf("extract MRZ: %v", err)
	}
	if data.OverallConfidence != 0 || data.FieldConfidences != nil {
		t.Fatalf("MRZ parsing must not manufacture calibrated confidence: %#v", data)
	}
	if data.MRZData == nil || !data.MRZData.IsValid {
		t.Fatal("expected valid deterministic MRZ evidence")
	}
}
