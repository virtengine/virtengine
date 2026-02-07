package adapters

import (
	"context"
	"image"
	"strings"
	"time"

	"github.com/virtengine/virtengine/pkg/veid/documents"
	"github.com/virtengine/virtengine/pkg/veid/documents/mrz"
)

type MRZAdapter struct {
	countries map[documents.CountryCode]struct{}
}

func NewMRZAdapter(countries []documents.CountryCode) *MRZAdapter {
	set := make(map[documents.CountryCode]struct{}, len(countries))
	for _, code := range countries {
		set[code] = struct{}{}
	}
	return &MRZAdapter{countries: set}
}

func (a *MRZAdapter) SupportedTypes() []documents.DocumentType {
	return []documents.DocumentType{
		documents.DocumentTypePassport,
		documents.DocumentTypeIDCard,
		documents.DocumentTypeResidencePermit,
		documents.DocumentTypeVisa,
	}
}

func (a *MRZAdapter) SupportedCountries() []documents.CountryCode {
	countries := make([]documents.CountryCode, 0, len(a.countries))
	for code := range a.countries {
		countries = append(countries, code)
	}
	return countries
}

func (a *MRZAdapter) CanProcess(docType documents.DocumentType, country documents.CountryCode) bool {
	if _, ok := a.countries[country]; !ok {
		return false
	}
	for _, supported := range a.SupportedTypes() {
		if supported == docType {
			return true
		}
	}
	return false
}

func (a *MRZAdapter) Extract(ctx context.Context, img image.Image) (*documents.DocumentData, error) {
	return nil, documents.ErrMRZRequired
}

func (a *MRZAdapter) ExtractWithMRZ(ctx context.Context, img image.Image, mrzValue string) (*documents.DocumentData, error) {
	parsed, err := mrz.Parse(mrzValue)
	if err != nil {
		return nil, err
	}

	docType := mapDocumentType(parsed.DocumentType)

	data := &documents.DocumentData{
		GivenNames:        parsed.GivenNames,
		Surname:           parsed.Surname,
		DateOfBirth:       parsed.DateOfBirth,
		Sex:               parsed.Sex,
		Nationality:       documents.CountryCode(parsed.Nationality),
		DocumentType:      docType,
		DocumentNumber:    parsed.DocumentNumber,
		IssuingCountry:    documents.CountryCode(parsed.IssuingCountry),
		ExpiryDate:        parsed.ExpiryDate,
		MRZData:           parsed,
		OverallConfidence: 0.92,
		FieldConfidences: map[string]float64{
			"name":            0.95,
			"document_number": 0.94,
			"date_of_birth":   0.93,
			"expiry_date":     0.93,
		},
	}

	return data, nil
}

func (a *MRZAdapter) Validate(data *documents.DocumentData) ([]documents.ValidationError, error) {
	if data == nil {
		return nil, documents.ErrInvalidDocument
	}

	var errs []documents.ValidationError
	if strings.TrimSpace(data.Surname) == "" {
		errs = append(errs, documents.ValidationError{Field: "surname", Message: "missing surname"})
	}
	if strings.TrimSpace(data.GivenNames) == "" {
		errs = append(errs, documents.ValidationError{Field: "given_names", Message: "missing given names"})
	}
	if strings.TrimSpace(data.DocumentNumber) == "" {
		errs = append(errs, documents.ValidationError{Field: "document_number", Message: "missing document number"})
	}
	if data.DateOfBirth.IsZero() {
		errs = append(errs, documents.ValidationError{Field: "date_of_birth", Message: "missing date of birth"})
	}
	if data.ExpiryDate.IsZero() {
		errs = append(errs, documents.ValidationError{Field: "expiry_date", Message: "missing expiry date"})
	} else if data.ExpiryDate.Before(time.Now().AddDate(0, 0, -1)) {
		errs = append(errs, documents.ValidationError{Field: "expiry_date", Message: "document expired"})
	}
	if data.MRZData != nil && !data.MRZData.IsValid {
		errs = append(errs, documents.ValidationError{Field: "mrz", Message: "mrz check digits failed"})
	}

	if len(errs) > 0 {
		return errs, documents.ErrInvalidDocument
	}
	return nil, nil
}

func mapDocumentType(code string) documents.DocumentType {
	if code == "" {
		return documents.DocumentTypePassport
	}
	switch code[0] {
	case 'P':
		return documents.DocumentTypePassport
	case 'I':
		return documents.DocumentTypeIDCard
	case 'V':
		return documents.DocumentTypeVisa
	case 'A':
		return documents.DocumentTypeResidencePermit
	default:
		return documents.DocumentTypeIDCard
	}
}
