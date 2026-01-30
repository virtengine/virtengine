package types

import (
	"fmt"
)

// ============================================================================
// ML Feature Schema Version and Constants
// ============================================================================

// MLFeatureSchemaVersion defines the current version of the ML feature schema.
// This version must match between training and inference pipelines.
const MLFeatureSchemaVersion = "1.0.0"

// MLFeatureSchemaVersionMajor is the major version component
const MLFeatureSchemaVersionMajor = 1

// MLFeatureSchemaVersionMinor is the minor version component
const MLFeatureSchemaVersionMinor = 0

// MLFeatureSchemaVersionPatch is the patch version component
const MLFeatureSchemaVersionPatch = 0

// ============================================================================
// Feature Dimension Constants
// ============================================================================

const (
	// FaceEmbeddingDim is the dimension of face embedding vectors
	FaceEmbeddingDim = 512

	// DocQualityDim is the dimension of document quality features
	DocQualityDim = 5

	// OCRFieldCount is the number of OCR fields tracked
	OCRFieldCount = 5

	// OCRFeaturesDim is the dimension of OCR features (fields * 2)
	OCRFeaturesDim = OCRFieldCount * 2

	// MetadataFeaturesDim is the dimension of metadata features
	MetadataFeaturesDim = 16

	// TotalFeatureDim is the total combined feature vector dimension
	TotalFeatureDim = 768

	// PaddingDim is the reserved padding dimension
	PaddingDim = TotalFeatureDim - FaceEmbeddingDim - DocQualityDim - OCRFeaturesDim - MetadataFeaturesDim
)

// Validate dimension constants at compile time
var _ = [1]struct{}{}[TotalFeatureDim-768]               // TotalFeatureDim must be 768
var _ = [1]struct{}{}[PaddingDim-225]                    // PaddingDim must be 225
var _ = [1]struct{}{}[FaceEmbeddingDim+DocQualityDim+OCRFeaturesDim+MetadataFeaturesDim+PaddingDim-768] // Sum must be 768

// ============================================================================
// Feature Group Definitions
// ============================================================================

// FeatureGroup represents a group of related ML features
type FeatureGroup string

const (
	// FeatureGroupFace contains face embedding and related features
	FeatureGroupFace FeatureGroup = "face"

	// FeatureGroupDocQuality contains document quality metrics
	FeatureGroupDocQuality FeatureGroup = "doc_quality"

	// FeatureGroupOCR contains OCR-extracted field features
	FeatureGroupOCR FeatureGroup = "ocr"

	// FeatureGroupMetadata contains contextual metadata features
	FeatureGroupMetadata FeatureGroup = "metadata"

	// FeatureGroupPadding contains reserved padding dimensions
	FeatureGroupPadding FeatureGroup = "padding"
)

// AllFeatureGroups returns all valid feature groups
func AllFeatureGroups() []FeatureGroup {
	return []FeatureGroup{
		FeatureGroupFace,
		FeatureGroupDocQuality,
		FeatureGroupOCR,
		FeatureGroupMetadata,
		FeatureGroupPadding,
	}
}

// ============================================================================
// Feature Offset Map
// ============================================================================

// FeatureOffset defines the offset and size of a feature group in the feature vector
type FeatureOffset struct {
	Group      FeatureGroup
	StartIndex int
	EndIndex   int
	Dimension  int
}

// FeatureOffsets returns the offset map for all feature groups
func FeatureOffsets() []FeatureOffset {
	return []FeatureOffset{
		{Group: FeatureGroupFace, StartIndex: 0, EndIndex: FaceEmbeddingDim - 1, Dimension: FaceEmbeddingDim},
		{Group: FeatureGroupDocQuality, StartIndex: FaceEmbeddingDim, EndIndex: FaceEmbeddingDim + DocQualityDim - 1, Dimension: DocQualityDim},
		{Group: FeatureGroupOCR, StartIndex: FaceEmbeddingDim + DocQualityDim, EndIndex: FaceEmbeddingDim + DocQualityDim + OCRFeaturesDim - 1, Dimension: OCRFeaturesDim},
		{Group: FeatureGroupMetadata, StartIndex: FaceEmbeddingDim + DocQualityDim + OCRFeaturesDim, EndIndex: FaceEmbeddingDim + DocQualityDim + OCRFeaturesDim + MetadataFeaturesDim - 1, Dimension: MetadataFeaturesDim},
		{Group: FeatureGroupPadding, StartIndex: FaceEmbeddingDim + DocQualityDim + OCRFeaturesDim + MetadataFeaturesDim, EndIndex: TotalFeatureDim - 1, Dimension: PaddingDim},
	}
}

// GetFeatureOffset returns the offset for a specific feature group
func GetFeatureOffset(group FeatureGroup) (FeatureOffset, bool) {
	for _, offset := range FeatureOffsets() {
		if offset.Group == group {
			return offset, true
		}
	}
	return FeatureOffset{}, false
}

// ============================================================================
// OCR Field Definitions
// ============================================================================

// OCRFieldName represents a recognized OCR field
type OCRFieldName string

const (
	OCRFieldName_Name           OCRFieldName = "name"
	OCRFieldName_DateOfBirth    OCRFieldName = "date_of_birth"
	OCRFieldName_DocumentNumber OCRFieldName = "document_number"
	OCRFieldName_ExpiryDate     OCRFieldName = "expiry_date"
	OCRFieldName_Nationality    OCRFieldName = "nationality"
)

// OCRFieldNames returns all OCR field names in order
func OCRFieldNames() []OCRFieldName {
	return []OCRFieldName{
		OCRFieldName_Name,
		OCRFieldName_DateOfBirth,
		OCRFieldName_DocumentNumber,
		OCRFieldName_ExpiryDate,
		OCRFieldName_Nationality,
	}
}

// OCRFieldIndex returns the index of an OCR field (0-4)
func OCRFieldIndex(field OCRFieldName) (int, bool) {
	for i, f := range OCRFieldNames() {
		if f == field {
			return i, true
		}
	}
	return -1, false
}

// ============================================================================
// Consent Category Definitions
// ============================================================================

// ConsentCategory represents a category of consent for feature groups
type ConsentCategory string

const (
	// ConsentCategoryBiometricPII requires explicit consent for biometric PII
	ConsentCategoryBiometricPII ConsentCategory = "biometric_pii"

	// ConsentCategoryBiometric requires explicit consent for biometric data
	ConsentCategoryBiometric ConsentCategory = "biometric"

	// ConsentCategoryIdentityAttestation has implicit consent via SSO flow
	ConsentCategoryIdentityAttestation ConsentCategory = "identity_attestation"

	// ConsentCategoryContactVerification has implicit consent via verification
	ConsentCategoryContactVerification ConsentCategory = "contact_verification"

	// ConsentCategoryDomainOwnership has implicit consent via DNS record
	ConsentCategoryDomainOwnership ConsentCategory = "domain_ownership"

	// ConsentCategoryEnterpriseIdentity is delegated via enterprise
	ConsentCategoryEnterpriseIdentity ConsentCategory = "enterprise_identity"
)

// ScopeTypeToConsentCategory maps scope types to their consent categories
func ScopeTypeToConsentCategory(scopeType ScopeType) ConsentCategory {
	switch scopeType {
	case ScopeTypeIDDocument:
		return ConsentCategoryBiometricPII
	case ScopeTypeSelfie, ScopeTypeFaceVideo, ScopeTypeBiometric:
		return ConsentCategoryBiometric
	case ScopeTypeSSOMetadata:
		return ConsentCategoryIdentityAttestation
	case ScopeTypeEmailProof, ScopeTypeSMSProof:
		return ConsentCategoryContactVerification
	case ScopeTypeDomainVerify:
		return ConsentCategoryDomainOwnership
	case ScopeTypeADSSO:
		return ConsentCategoryEnterpriseIdentity
	default:
		return ConsentCategoryBiometricPII // Default to most restrictive
	}
}

// RequiresExplicitConsent returns true if the consent category requires explicit consent
func RequiresExplicitConsent(category ConsentCategory) bool {
	switch category {
	case ConsentCategoryBiometricPII, ConsentCategoryBiometric:
		return true
	default:
		return false
	}
}

// ============================================================================
// Feature Value Ranges
// ============================================================================

// FeatureRange defines the valid range for a feature value
type FeatureRange struct {
	Name    string
	Min     float32
	Max     float32
	Default float32
	Unit    string
}

// DocQualityFeatureRanges returns the valid ranges for document quality features
func DocQualityFeatureRanges() []FeatureRange {
	return []FeatureRange{
		{Name: "sharpness", Min: 0.0, Max: 1.0, Default: 0.5, Unit: "score"},
		{Name: "brightness", Min: 0.0, Max: 1.0, Default: 0.5, Unit: "score"},
		{Name: "contrast", Min: 0.0, Max: 1.0, Default: 0.5, Unit: "score"},
		{Name: "noise_level_inv", Min: 0.0, Max: 1.0, Default: 0.5, Unit: "score"},
		{Name: "blur_score_inv", Min: 0.0, Max: 1.0, Default: 0.5, Unit: "score"},
	}
}

// OCRFeatureRanges returns the valid ranges for OCR features (per field)
func OCRFeatureRanges() []FeatureRange {
	return []FeatureRange{
		{Name: "confidence", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "probability"},
		{Name: "validated", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
	}
}

// MetadataFeatureRanges returns the valid ranges for metadata features
func MetadataFeatureRanges() []FeatureRange {
	return []FeatureRange{
		{Name: "scope_count", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "normalized_count"},
		{Name: "has_id_document", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_selfie", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_face_video", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_biometric", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_sso_metadata", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_email_proof", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_sms_proof", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "has_domain_verify", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "binary"},
		{Name: "face_confidence", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "probability"},
		{Name: "block_height_normalized", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "ratio"},
		{Name: "reserved_11", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "reserved"},
		{Name: "reserved_12", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "reserved"},
		{Name: "reserved_13", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "reserved"},
		{Name: "reserved_14", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "reserved"},
		{Name: "reserved_15", Min: 0.0, Max: 1.0, Default: 0.0, Unit: "reserved"},
	}
}

// ============================================================================
// Feature Vector Validation
// ============================================================================

// MLFeatureVector represents a complete ML feature vector
type MLFeatureVector struct {
	// SchemaVersion is the schema version used to create this vector
	SchemaVersion string `json:"schema_version"`

	// Features is the raw feature vector
	Features []float32 `json:"features"`
}

// NewMLFeatureVector creates a new feature vector with default values
func NewMLFeatureVector() *MLFeatureVector {
	features := make([]float32, TotalFeatureDim)

	// Set default values for doc quality features
	docQualityOffset, _ := GetFeatureOffset(FeatureGroupDocQuality)
	for i, r := range DocQualityFeatureRanges() {
		features[docQualityOffset.StartIndex+i] = r.Default
	}

	return &MLFeatureVector{
		SchemaVersion: MLFeatureSchemaVersion,
		Features:      features,
	}
}

// Validate validates the feature vector
func (v *MLFeatureVector) Validate() error {
	if v.SchemaVersion == "" {
		return ErrInvalidParams.Wrap("schema_version cannot be empty")
	}

	if len(v.Features) != TotalFeatureDim {
		return ErrInvalidParams.Wrapf("feature vector dimension mismatch: expected %d, got %d", TotalFeatureDim, len(v.Features))
	}

	// Validate face embedding is normalized (L2 norm should be ~1.0 or 0.0)
	var sumSquares float64
	for i := 0; i < FaceEmbeddingDim; i++ {
		sumSquares += float64(v.Features[i]) * float64(v.Features[i])
	}
	if sumSquares > 0.01 { // Non-zero embedding
		norm := sumSquares
		if norm < 0.99 || norm > 1.01 {
			return ErrInvalidParams.Wrapf("face embedding not normalized: L2 norm = %.4f", norm)
		}
	}

	// Validate doc quality features are in range
	docQualityOffset, _ := GetFeatureOffset(FeatureGroupDocQuality)
	for i, r := range DocQualityFeatureRanges() {
		val := v.Features[docQualityOffset.StartIndex+i]
		if val < r.Min || val > r.Max {
			return ErrInvalidParams.Wrapf("doc quality feature %s out of range [%.2f, %.2f]: %.4f", r.Name, r.Min, r.Max, val)
		}
	}

	// Validate OCR features are in range
	ocrOffset, _ := GetFeatureOffset(FeatureGroupOCR)
	ocrRanges := OCRFeatureRanges()
	for i := 0; i < OCRFieldCount; i++ {
		for j, r := range ocrRanges {
			idx := ocrOffset.StartIndex + i*2 + j
			val := v.Features[idx]
			if val < r.Min || val > r.Max {
				return ErrInvalidParams.Wrapf("OCR feature %s[%d] out of range [%.2f, %.2f]: %.4f", r.Name, i, r.Min, r.Max, val)
			}
		}
	}

	// Validate metadata features are in range
	metaOffset, _ := GetFeatureOffset(FeatureGroupMetadata)
	for i, r := range MetadataFeatureRanges() {
		val := v.Features[metaOffset.StartIndex+i]
		if val < r.Min || val > r.Max {
			return ErrInvalidParams.Wrapf("metadata feature %s out of range [%.2f, %.2f]: %.4f", r.Name, r.Min, r.Max, val)
		}
	}

	return nil
}

// GetFeatureGroup extracts a specific feature group from the vector
func (v *MLFeatureVector) GetFeatureGroup(group FeatureGroup) ([]float32, error) {
	offset, found := GetFeatureOffset(group)
	if !found {
		return nil, fmt.Errorf("unknown feature group: %s", group)
	}

	if len(v.Features) < offset.EndIndex+1 {
		return nil, fmt.Errorf("feature vector too short for group %s", group)
	}

	return v.Features[offset.StartIndex : offset.EndIndex+1], nil
}

// SetFeatureGroup sets a specific feature group in the vector
func (v *MLFeatureVector) SetFeatureGroup(group FeatureGroup, values []float32) error {
	offset, found := GetFeatureOffset(group)
	if !found {
		return fmt.Errorf("unknown feature group: %s", group)
	}

	if len(values) != offset.Dimension {
		return fmt.Errorf("dimension mismatch for group %s: expected %d, got %d", group, offset.Dimension, len(values))
	}

	copy(v.Features[offset.StartIndex:offset.EndIndex+1], values)
	return nil
}

// ============================================================================
// Schema Compatibility
// ============================================================================

// IsSchemaCompatible checks if two schema versions are compatible
func IsSchemaCompatible(version1, version2 string) bool {
	// For now, only exact matches within major version are compatible
	// Parse version strings to compare major versions
	var major1, minor1, patch1 int
	var major2, minor2, patch2 int

	_, err1 := fmt.Sscanf(version1, "%d.%d.%d", &major1, &minor1, &patch1)
	_, err2 := fmt.Sscanf(version2, "%d.%d.%d", &major2, &minor2, &patch2)

	if err1 != nil || err2 != nil {
		return version1 == version2 // Fallback to exact match
	}

	// Same major version = compatible
	return major1 == major2
}

// GetSchemaInfo returns information about the current schema
func GetSchemaInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":             MLFeatureSchemaVersion,
		"major":               MLFeatureSchemaVersionMajor,
		"minor":               MLFeatureSchemaVersionMinor,
		"patch":               MLFeatureSchemaVersionPatch,
		"total_dimension":     TotalFeatureDim,
		"face_embedding_dim":  FaceEmbeddingDim,
		"doc_quality_dim":     DocQualityDim,
		"ocr_features_dim":    OCRFeaturesDim,
		"metadata_dim":        MetadataFeaturesDim,
		"padding_dim":         PaddingDim,
		"ocr_field_count":     OCRFieldCount,
		"scope_schema_version": ScopeSchemaVersion,
	}
}
