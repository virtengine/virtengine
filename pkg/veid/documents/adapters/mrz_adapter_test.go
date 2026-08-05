package adapters

import (
	"context"
	"errors"
	"reflect"
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

func TestMRZAdapterSupportedCountriesAreDeterministic(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"USA", "AUS", "DEU"})
	want := []documents.CountryCode{"AUS", "DEU", "USA"}
	for range 100 {
		if got := adapter.SupportedCountries(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestMRZAdapterHonorsCancellationAndRejectsTamperedDerivedFields(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"UTO"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := adapter.ExtractWithMRZ(ctx, nil, "ignored"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	data, err := adapter.ExtractWithMRZ(context.Background(), nil, "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10")
	if err != nil {
		t.Fatal(err)
	}
	data.DocumentNumber = "tampered"
	validation, err := adapter.Validate(data)
	if !errors.Is(err, documents.ErrInvalidDocument) || len(validation) == 0 {
		t.Fatalf("expected invalid tampered MRZ data, got %v / %#v", err, validation)
	}
}
