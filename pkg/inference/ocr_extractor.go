package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// OCR Extractor
// ============================================================================

// OCRExtractor extracts OCR and document quality features from ID document scopes.
// It interfaces with the Python OCR pipeline via gRPC sidecar or provides
// deterministic stub features when unavailable.
type OCRExtractor struct {
	config OCRExtractorConfig
	client OCRExtractionClient

	mu              sync.RWMutex
	isHealthy       bool
	extractionCount uint64
	errorCount      uint64
}

// OCRExtractorConfig contains configuration for OCR extraction
type OCRExtractorConfig struct {
	// SidecarAddress is the gRPC address of the OCR extraction sidecar
	SidecarAddress string

	// Timeout for extraction operations
	Timeout time.Duration

	// ExpectedFields is the list of expected OCR field names
	ExpectedFields []string

	// MinConfidence is the minimum confidence threshold for valid extractions
	MinConfidence float32

	// UseFallbackOnError returns stub values on extraction errors
	UseFallbackOnError bool

	// ValidateFieldFormats enables field format validation
	ValidateFieldFormats bool

	// MaxRetries for transient failures
	MaxRetries int

	// RetryDelay between retry attempts
	RetryDelay time.Duration
}

// DefaultOCRExtractorConfig returns sensible defaults
func DefaultOCRExtractorConfig() OCRExtractorConfig {
	return OCRExtractorConfig{
		SidecarAddress:       "localhost:50053",
		Timeout:              5 * time.Second,
		ExpectedFields:       OCRFieldNames, // from types.go
		MinConfidence:        0.5,
		UseFallbackOnError:   true,
		ValidateFieldFormats: true,
		MaxRetries:           2,
		RetryDelay:           100 * time.Millisecond,
	}
}

// OCRExtractionClient defines the interface for OCR extraction backends
type OCRExtractionClient interface {
	// ExtractText extracts text and fields from document image
	ExtractText(ctx context.Context, imageData []byte, documentType string) (*OCRExtractionResult, error)

	// IsHealthy checks if the client is ready
	IsHealthy() bool

	// Close releases resources
	Close() error
}

// OCRExtractionResult contains the complete result of OCR extraction
type OCRExtractionResult struct {
	// Fields contains extracted field values with metadata
	Fields map[string]*ExtractedField

	// DocumentQuality contains document image quality metrics
	DocumentQuality DocumentQualityResult

	// RawText is the full extracted text (for hashing, not storage)
	RawText string

	// DocumentType detected document type
	DocumentType string

	// OverallConfidence is the aggregate OCR confidence
	OverallConfidence float32

	// IdentityHash is a hash of the extracted identity fields
	IdentityHash string

	// ModelVersion used for extraction
	ModelVersion string

	// ProcessingTimeMs is the extraction time in milliseconds
	ProcessingTimeMs int64

	// ReasonCodes provide explanations for extraction outcome
	ReasonCodes []string

	// Success indicates if extraction was successful
	Success bool
}

// ExtractedField contains a single extracted OCR field
type ExtractedField struct {
	// Value is the extracted text value
	Value string

	// Confidence is the OCR confidence for this field (0.0-1.0)
	Confidence float32

	// Validated indicates if the field passed format validation
	Validated bool

	// ValidationError if validation failed
	ValidationError string

	// BoundingBox of the field in the image
	BoundingBox FieldBoundingBox

	// Hash of the field value for on-chain storage
	Hash string
}

// FieldBoundingBox represents a field location in the document
type FieldBoundingBox struct {
	X      int
	Y      int
	Width  int
	Height int
}

// DocumentQualityResult contains document image quality metrics
type DocumentQualityResult struct {
	// Sharpness of the document image (0.0-1.0)
	Sharpness float32

	// Brightness of the document (0.0-1.0, 0.5 is ideal)
	Brightness float32

	// Contrast of the document (0.0-1.0)
	Contrast float32

	// NoiseLevel of the document (0.0-1.0, lower is better)
	NoiseLevel float32

	// BlurScore indicating blur amount (0.0-1.0, lower is better)
	BlurScore float32

	// GlareScore indicating glare/reflection (0.0-1.0, lower is better)
	GlareScore float32

	// SkewAngle in degrees
	SkewAngle float32

	// OverallScore combining all quality metrics (0.0-1.0)
	OverallScore float32
}

// ============================================================================
// OCR Extraction Reason Codes
// ============================================================================

const (
	OCRReasonCodeSuccess          = "OCR_EXTRACTION_SUCCESS"
	OCRReasonCodeNoTextFound      = "NO_TEXT_FOUND"
	OCRReasonCodeLowConfidence    = "LOW_OCR_CONFIDENCE"
	OCRReasonCodeLowQuality       = "LOW_DOCUMENT_QUALITY"
	OCRReasonCodeMissingFields    = "MISSING_REQUIRED_FIELDS"
	OCRReasonCodeValidationFailed = "FIELD_VALIDATION_FAILED"
	OCRReasonCodeExtractionError  = "OCR_EXTRACTION_ERROR"
	OCRReasonCodeSidecarUnavail   = "OCR_SIDECAR_UNAVAILABLE"
	OCRReasonCodeTimeout          = "OCR_EXTRACTION_TIMEOUT"
	OCRReasonCodeHighNoise        = "HIGH_DOCUMENT_NOISE"
	OCRReasonCodeHighBlur         = "HIGH_DOCUMENT_BLUR"
	OCRReasonCodeUnknownDocType   = "UNKNOWN_DOCUMENT_TYPE"
)

// ============================================================================
// Field Validation Patterns
// ============================================================================

var (
	// namePattern matches valid name characters
	namePattern = regexp.MustCompile(`^[A-Za-z\s\-'.]+$`)

	// datePattern matches common date formats (DD/MM/YYYY, YYYY-MM-DD, etc.)
	datePattern = regexp.MustCompile(`^(\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4}|\d{4}[/\-\.]\d{1,2}[/\-\.]\d{1,2})$`)

	// documentNumberPattern matches alphanumeric document numbers
	documentNumberPattern = regexp.MustCompile(`^[A-Z0-9\-]+$`)

	// countryPattern matches country codes or names
	countryPattern = regexp.MustCompile(`^[A-Za-z\s]+$`)
)

// ============================================================================
// Constructor and Interface Implementation
// ============================================================================

// NewOCRExtractor creates a new OCR extractor
func NewOCRExtractor(config OCRExtractorConfig) *OCRExtractor {
	return &OCRExtractor{
		config:    config,
		isHealthy: false,
	}
}

// NewOCRExtractorWithClient creates an OCR extractor with a specific client
func NewOCRExtractorWithClient(config OCRExtractorConfig, client OCRExtractionClient) *OCRExtractor {
	oe := NewOCRExtractor(config)
	oe.client = client
	oe.isHealthy = client != nil && client.IsHealthy()
	return oe
}

// Extract extracts OCR features from a document image
func (oe *OCRExtractor) Extract(ctx context.Context, imageData []byte, documentType string) (*OCRExtractionResult, error) {
	if len(imageData) == 0 {
		return oe.createFailureResult(OCRReasonCodeExtractionError, "empty image data"), nil
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, oe.config.Timeout)
	defer cancel()

	oe.mu.Lock()
	oe.extractionCount++
	oe.mu.Unlock()

	// Try extraction with retries
	var result *OCRExtractionResult
	var lastErr error

	for attempt := 0; attempt <= oe.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return oe.handleError(ctx.Err(), "timeout during retry")
			case <-time.After(oe.config.RetryDelay):
			}
		}

		result, lastErr = oe.doExtraction(ctx, imageData, documentType)
		if lastErr == nil && result.Success {
			break
		}
	}

	if lastErr != nil {
		return oe.handleError(lastErr, "extraction failed after retries")
	}

	// Validate and post-process result
	oe.postProcessResult(result)

	return result, nil
}

// doExtraction performs the actual extraction via client or fallback
func (oe *OCRExtractor) doExtraction(ctx context.Context, imageData []byte, documentType string) (*OCRExtractionResult, error) {
	startTime := time.Now()

	if oe.client == nil || !oe.client.IsHealthy() {
		// Use fallback stub extraction
		return oe.extractFallback(imageData, documentType, startTime), nil
	}

	result, err := oe.client.ExtractText(ctx, imageData, documentType)
	if err != nil {
		return nil, err
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// extractFallback generates deterministic stub OCR results when sidecar is unavailable
func (oe *OCRExtractor) extractFallback(imageData []byte, documentType string, startTime time.Time) *OCRExtractionResult {
	// Generate deterministic field values from image hash
	fields := oe.generateDeterministicFields(imageData)

	// Generate quality metrics
	quality := oe.generateDeterministicQuality(imageData)

	// Compute identity hash
	identityHash := oe.computeIdentityHash(fields)

	return &OCRExtractionResult{
		Fields:            fields,
		DocumentQuality:   quality,
		RawText:           "[stub extraction]",
		DocumentType:      documentType,
		OverallConfidence: 0.85,
		IdentityHash:      identityHash,
		ModelVersion:      "stub-v1.0.0",
		ProcessingTimeMs:  time.Since(startTime).Milliseconds(),
		ReasonCodes:       []string{OCRReasonCodeSuccess, OCRReasonCodeSidecarUnavail},
		Success:           true,
	}
}

// generateDeterministicFields creates deterministic stub field values
func (oe *OCRExtractor) generateDeterministicFields(imageData []byte) map[string]*ExtractedField {
	fields := make(map[string]*ExtractedField)

	// Use hash to generate deterministic but varied values
	hash := sha256.Sum256(imageData)

	for i, fieldName := range oe.config.ExpectedFields {
		// Generate confidence based on hash byte
		byteIdx := i % len(hash)
		confidence := 0.7 + float32(hash[byteIdx]%30)/100.0 // 0.7-0.99

		// Generate field hash
		fieldHash := sha256.Sum256(append(hash[:], []byte(fieldName)...))

		fields[fieldName] = &ExtractedField{
			Value:      fmt.Sprintf("[%s_stub]", fieldName),
			Confidence: confidence,
			Validated:  true,
			Hash:       hex.EncodeToString(fieldHash[:16]),
		}
	}

	return fields
}

// generateDeterministicQuality creates deterministic document quality metrics
func (oe *OCRExtractor) generateDeterministicQuality(imageData []byte) DocumentQualityResult {
	hash := sha256.Sum256(imageData)

	return DocumentQualityResult{
		Sharpness:    0.7 + float32(hash[0]%30)/100.0,
		Brightness:   0.4 + float32(hash[1]%20)/100.0,
		Contrast:     0.6 + float32(hash[2]%30)/100.0,
		NoiseLevel:   float32(hash[3]%30) / 100.0,
		BlurScore:    float32(hash[4]%20) / 100.0,
		GlareScore:   float32(hash[5]%15) / 100.0,
		SkewAngle:    float32(hash[6]%5) - 2.5,
		OverallScore: 0.75 + float32(hash[7]%20)/100.0,
	}
}

// postProcessResult validates and enriches the extraction result
func (oe *OCRExtractor) postProcessResult(result *OCRExtractionResult) {
	if result == nil {
		return
	}

	// Validate field formats if enabled
	if oe.config.ValidateFieldFormats {
		oe.validateFields(result)
	}

	// Clamp confidence values
	result.OverallConfidence = clampFloat32(result.OverallConfidence, 0.0, 1.0)

	// Clamp quality metrics
	oe.clampQualityMetrics(&result.DocumentQuality)

	// Add reason codes based on quality
	if result.OverallConfidence < oe.config.MinConfidence {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeLowConfidence)
	}

	if result.DocumentQuality.OverallScore < 0.5 {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeLowQuality)
	}

	if result.DocumentQuality.NoiseLevel > 0.3 {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeHighNoise)
	}

	if result.DocumentQuality.BlurScore > 0.3 {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeHighBlur)
	}

	// Check for missing required fields
	missingFields := oe.getMissingFields(result)
	if len(missingFields) > 0 {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeMissingFields)
	}

	// Compute identity hash if not present
	if result.IdentityHash == "" {
		result.IdentityHash = oe.computeIdentityHash(result.Fields)
	}
}

// validateFields validates field formats and updates validation status
func (oe *OCRExtractor) validateFields(result *OCRExtractionResult) {
	validationFailed := false

	for fieldName, field := range result.Fields {
		if field == nil || field.Value == "" {
			continue
		}

		var valid bool
		var errMsg string

		switch fieldName {
		case "name", "full_name", "surname", "given_names":
			valid, errMsg = oe.validateName(field.Value)
		case "date_of_birth", "expiry_date", "issue_date":
			valid, errMsg = oe.validateDate(field.Value)
		case "document_number", "passport_number", "license_number":
			valid, errMsg = oe.validateDocumentNumber(field.Value)
		case "nationality", "country":
			valid, errMsg = oe.validateCountry(field.Value)
		default:
			// Accept other fields without specific validation
			valid = true
		}

		field.Validated = valid
		if !valid {
			field.ValidationError = errMsg
			validationFailed = true
		}
	}

	if validationFailed {
		result.ReasonCodes = append(result.ReasonCodes, OCRReasonCodeValidationFailed)
	}
}

// validateName validates a name field
func (oe *OCRExtractor) validateName(value string) (bool, string) {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false, "name too short"
	}
	if len(value) > 100 {
		return false, "name too long"
	}
	if !namePattern.MatchString(value) {
		return false, "invalid name characters"
	}
	return true, ""
}

// validateDate validates a date field
func (oe *OCRExtractor) validateDate(value string) (bool, string) {
	value = strings.TrimSpace(value)
	if len(value) < 6 {
		return false, "date too short"
	}
	if !datePattern.MatchString(value) {
		return false, "invalid date format"
	}
	return true, ""
}

// validateDocumentNumber validates a document number field
func (oe *OCRExtractor) validateDocumentNumber(value string) (bool, string) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) < 4 {
		return false, "document number too short"
	}
	if len(value) > 20 {
		return false, "document number too long"
	}
	if !documentNumberPattern.MatchString(value) {
		return false, "invalid document number characters"
	}
	return true, ""
}

// validateCountry validates a country field
func (oe *OCRExtractor) validateCountry(value string) (bool, string) {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false, "country too short"
	}
	if !countryPattern.MatchString(value) {
		return false, "invalid country characters"
	}
	return true, ""
}

// getMissingFields returns list of expected fields that are missing
func (oe *OCRExtractor) getMissingFields(result *OCRExtractionResult) []string {
	var missing []string
	for _, fieldName := range oe.config.ExpectedFields {
		field, exists := result.Fields[fieldName]
		if !exists || field == nil || field.Value == "" {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

// clampQualityMetrics ensures all quality metrics are in valid range
func (oe *OCRExtractor) clampQualityMetrics(quality *DocumentQualityResult) {
	quality.Sharpness = clampFloat32(quality.Sharpness, 0.0, 1.0)
	quality.Brightness = clampFloat32(quality.Brightness, 0.0, 1.0)
	quality.Contrast = clampFloat32(quality.Contrast, 0.0, 1.0)
	quality.NoiseLevel = clampFloat32(quality.NoiseLevel, 0.0, 1.0)
	quality.BlurScore = clampFloat32(quality.BlurScore, 0.0, 1.0)
	quality.GlareScore = clampFloat32(quality.GlareScore, 0.0, 1.0)
	quality.SkewAngle = clampFloat32(quality.SkewAngle, -45.0, 45.0)
	quality.OverallScore = clampFloat32(quality.OverallScore, 0.0, 1.0)
}

// computeIdentityHash computes a hash of the identity fields
func (oe *OCRExtractor) computeIdentityHash(fields map[string]*ExtractedField) string {
	h := sha256.New()

	// Hash fields in deterministic order
	for _, fieldName := range oe.config.ExpectedFields {
		if field, exists := fields[fieldName]; exists && field != nil {
			h.Write([]byte(fieldName))
			h.Write([]byte(":"))
			h.Write([]byte(field.Value))
			h.Write([]byte("|"))
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// handleError handles extraction errors with optional fallback
func (oe *OCRExtractor) handleError(err error, msg string) (*OCRExtractionResult, error) {
	oe.mu.Lock()
	oe.errorCount++
	oe.mu.Unlock()

	if oe.config.UseFallbackOnError {
		return oe.createFailureResult(OCRReasonCodeExtractionError, msg), nil
	}
	return nil, fmt.Errorf("%s: %w", msg, err)
}

// createFailureResult creates a result indicating extraction failure
//
//nolint:unparam // message kept for future logging or error details in result
func (oe *OCRExtractor) createFailureResult(reasonCode, _ string) *OCRExtractionResult {
	return &OCRExtractionResult{
		Fields:            make(map[string]*ExtractedField),
		DocumentQuality:   DocumentQualityResult{},
		OverallConfidence: 0.0,
		ModelVersion:      "unknown",
		ReasonCodes:       []string{reasonCode},
		Success:           false,
	}
}

// ============================================================================
// Feature Vector Conversion
// ============================================================================

// ToFeatureInputs converts OCR result to ScoreInputs format
func (oe *OCRExtractor) ToFeatureInputs(result *OCRExtractionResult) (float32, DocQualityFeatures, map[string]float32, map[string]bool) {
	// Document quality score
	docQualityScore := result.DocumentQuality.OverallScore

	// Document quality features
	docQualityFeatures := DocQualityFeatures{
		Sharpness:  result.DocumentQuality.Sharpness,
		Brightness: result.DocumentQuality.Brightness,
		Contrast:   result.DocumentQuality.Contrast,
		NoiseLevel: result.DocumentQuality.NoiseLevel,
		BlurScore:  result.DocumentQuality.BlurScore,
	}

	// OCR confidences
	ocrConfidences := make(map[string]float32)
	for fieldName, field := range result.Fields {
		if field != nil {
			ocrConfidences[fieldName] = field.Confidence
		}
	}

	// OCR field validation
	ocrValidation := make(map[string]bool)
	for fieldName, field := range result.Fields {
		if field != nil {
			ocrValidation[fieldName] = field.Validated
		}
	}

	return docQualityScore, docQualityFeatures, ocrConfidences, ocrValidation
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateResult validates that an OCR result meets quality requirements
func (oe *OCRExtractor) ValidateResult(result *OCRExtractionResult) []string {
	var issues []string

	if result == nil {
		issues = append(issues, "nil OCR result")
		return issues
	}

	// Check overall confidence
	if result.OverallConfidence < 0 || result.OverallConfidence > 1 {
		issues = append(issues, fmt.Sprintf(
			"overall confidence out of range [0,1]: %.4f",
			result.OverallConfidence,
		))
	}

	// Check document quality
	if result.DocumentQuality.OverallScore < 0 || result.DocumentQuality.OverallScore > 1 {
		issues = append(issues, fmt.Sprintf(
			"document quality score out of range [0,1]: %.4f",
			result.DocumentQuality.OverallScore,
		))
	}

	// Check field confidences
	for fieldName, field := range result.Fields {
		if field == nil {
			continue
		}
		if field.Confidence < 0 || field.Confidence > 1 {
			issues = append(issues, fmt.Sprintf(
				"field '%s' confidence out of range [0,1]: %.4f",
				fieldName, field.Confidence,
			))
		}
	}

	// Check for NaN values
	if math.IsNaN(float64(result.OverallConfidence)) {
		issues = append(issues, "overall confidence is NaN")
	}
	if math.IsNaN(float64(result.DocumentQuality.OverallScore)) {
		issues = append(issues, "document quality score is NaN")
	}

	return issues
}

// SanitizeResult ensures all values in an OCR result are valid
func (oe *OCRExtractor) SanitizeResult(result *OCRExtractionResult) {
	if result == nil {
		return
	}

	// Sanitize confidence
	if math.IsNaN(float64(result.OverallConfidence)) || math.IsInf(float64(result.OverallConfidence), 0) {
		result.OverallConfidence = 0.0
	}
	result.OverallConfidence = clampFloat32(result.OverallConfidence, 0.0, 1.0)

	// Sanitize document quality
	oe.clampQualityMetrics(&result.DocumentQuality)

	// Sanitize field confidences
	for _, field := range result.Fields {
		if field == nil {
			continue
		}
		if math.IsNaN(float64(field.Confidence)) || math.IsInf(float64(field.Confidence), 0) {
			field.Confidence = 0.0
		}
		field.Confidence = clampFloat32(field.Confidence, 0.0, 1.0)
	}
}

// ============================================================================
// Statistics and Health
// ============================================================================

// IsHealthy returns whether the extractor is ready
func (oe *OCRExtractor) IsHealthy() bool {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	if oe.client != nil {
		return oe.client.IsHealthy()
	}
	// Stub mode is always healthy
	return true
}

// GetStats returns extraction statistics
func (oe *OCRExtractor) GetStats() OCRExtractorStats {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	return OCRExtractorStats{
		ExtractionCount: oe.extractionCount,
		ErrorCount:      oe.errorCount,
		IsHealthy:       oe.IsHealthy(),
		UsingStub:       oe.client == nil,
	}
}

// OCRExtractorStats contains extractor statistics
type OCRExtractorStats struct {
	ExtractionCount uint64
	ErrorCount      uint64
	IsHealthy       bool
	UsingStub       bool
}

// Close releases resources
func (oe *OCRExtractor) Close() error {
	if oe.client != nil {
		return oe.client.Close()
	}
	return nil
}
