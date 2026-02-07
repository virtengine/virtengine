package documents

import (
	"context"
	"errors"
	"image"
	"time"

	"github.com/virtengine/virtengine/pkg/veid/documents/mrz"
)

type DocumentType string

type CountryCode string

const (
	DocumentTypePassport        DocumentType = "passport"
	DocumentTypeIDCard          DocumentType = "id_card"
	DocumentTypeDriverLicense   DocumentType = "driver_license"
	DocumentTypeResidencePermit DocumentType = "residence_permit"
	DocumentTypeVisa            DocumentType = "visa"
)

var (
	ErrNoAdapter         = errors.New("no document adapter available")
	ErrUnsupportedFormat = errors.New("unsupported document format")
	ErrMRZRequired       = errors.New("mrz data required")
	ErrInvalidDocument   = errors.New("document validation failed")
	ErrOCRUnavailable    = errors.New("ocr engine unavailable")
	ErrNFCUnavailable    = errors.New("nfc reader unavailable")
	ErrNotImplemented    = errors.New("not implemented")
)

// DocumentAdapter extracts data from a specific document format.
type DocumentAdapter interface {
	SupportedTypes() []DocumentType
	SupportedCountries() []CountryCode
	CanProcess(docType DocumentType, country CountryCode) bool
	Extract(ctx context.Context, img image.Image) (*DocumentData, error)
	ExtractWithMRZ(ctx context.Context, img image.Image, mrzValue string) (*DocumentData, error)
	Validate(data *DocumentData) ([]ValidationError, error)
}

type Address struct {
	Line1      string
	Line2      string
	City       string
	Region     string
	PostalCode string
	Country    CountryCode
}

type ValidationError struct {
	Field   string
	Message string
}

// DocumentData is the standardized output from any adapter.
type DocumentData struct {
	GivenNames        string
	Surname           string
	DateOfBirth       time.Time
	Sex               string
	Nationality       CountryCode
	DocumentType      DocumentType
	DocumentNumber    string
	IssuingCountry    CountryCode
	ExpiryDate        time.Time
	IssueDate         *time.Time
	PlaceOfBirth      string
	Address           *Address
	FacialImage       []byte
	MRZData           *mrz.MRZData
	NFCData           *NFCData
	OverallConfidence float64
	FieldConfidences  map[string]float64
}

// NFCData captures data extracted from ePassport chips when available.
type NFCData struct {
	DG1           []byte
	DG2           []byte
	DG11          []byte
	DG12          []byte
	SOD           []byte
	PassiveValid  bool
	ActiveValid   bool
	PACEUsed      bool
	BACUsed       bool
	EACUsed       bool
	ChipAuthValid bool
}
