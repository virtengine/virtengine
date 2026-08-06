package adapters

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

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

func TestMRZAdapterValidateRejectsUnsupportedCapabilityBinding(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"UTO"})
	data, err := adapter.ExtractWithMRZ(context.Background(), nil, "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10")
	if err != nil {
		t.Fatal(err)
	}
	data.IssuingCountry = "USA"
	validation, err := adapter.Validate(data)
	if !errors.Is(err, documents.ErrInvalidDocument) {
		t.Fatalf("expected unsupported capability validation error, got %v", err)
	}
	for _, issue := range validation {
		if issue.Field == "document" {
			return
		}
	}
	t.Fatalf("expected document capability issue, got %#v", validation)
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

func TestMRZAdapterRejectsUnsupportedDirectExtraction(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"USA"})
	_, err := adapter.ExtractWithMRZ(context.Background(), nil, "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<\nL898902C36UTO7408122F1204159ZE184226B<<<<<10")
	if !errors.Is(err, documents.ErrNoAdapter) {
		t.Fatalf("expected direct extraction capability rejection, got %v", err)
	}
}

func TestMRZAdapterRejectsExpiredDocumentsAndInvalidSex(t *testing.T) {
	adapter := NewMRZAdapter([]documents.CountryCode{"UTO"})
	base := &documents.DocumentData{
		GivenNames:     "Anna",
		Surname:        "Eriksson",
		DateOfBirth:    time.Date(1974, 8, 12, 0, 0, 0, 0, time.UTC),
		Sex:            "F",
		Nationality:    "UTO",
		DocumentType:   documents.DocumentTypePassport,
		DocumentNumber: "L898902C3",
		IssuingCountry: "UTO",
		ExpiryDate:     time.Now().UTC().Truncate(24 * time.Hour),
	}

	expired := *base
	expired.ExpiryDate = expired.ExpiryDate.AddDate(0, 0, -1)
	validation, err := adapter.Validate(&expired)
	if !errors.Is(err, documents.ErrInvalidDocument) || !hasValidationField(validation, "expiry_date") {
		t.Fatalf("expected expiry rejection, got %v / %#v", err, validation)
	}

	invalidSex := *base
	invalidSex.Sex = "Z"
	validation, err = adapter.Validate(&invalidSex)
	if !errors.Is(err, documents.ErrInvalidDocument) || !hasValidationField(validation, "sex") {
		t.Fatalf("expected sex marker rejection, got %v / %#v", err, validation)
	}
}

func hasValidationField(validation []documents.ValidationError, field string) bool {
	for _, issue := range validation {
		if issue.Field == field {
			return true
		}
	}
	return false
}
