package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"time"
)

// ============================================================================
// Feature Extraction Metadata
// ============================================================================

// FeatureExtractionMetadata contains audit information about feature extraction
type FeatureExtractionMetadata struct {
	// SchemaVersion is the ML feature schema version used
	SchemaVersion string `json:"schema_version"`

	// ExtractionTimestamp is when extraction was performed
	ExtractionTimestamp time.Time `json:"extraction_timestamp"`

	// AccountAddress is the account being verified
	AccountAddress string `json:"account_address"`

	// BlockHeight is the block height at extraction time
	BlockHeight int64 `json:"block_height"`

	// ProcessingDurationMs is how long extraction took
	ProcessingDurationMs int64 `json:"processing_duration_ms"`

	// ModelsUsed tracks which ML models were used
	ModelsUsed []ModelUsageInfo `json:"models_used"`

	// FeatureHash is the deterministic hash of extracted features
	FeatureHash string `json:"feature_hash"`

	// QualityGatesPassed indicates if all quality gates passed
	QualityGatesPassed bool `json:"quality_gates_passed"`

	// QualityGatesSummary summarizes gate results
	QualityGatesSummary *QualityGatesSummary `json:"quality_gates_summary,omitempty"`

	// Errors encountered during extraction (non-fatal)
	Errors []ExtractionError `json:"errors,omitempty"`

	// ScopeExtractedCounts tracks how many scopes of each type were processed
	ScopeExtractedCounts map[string]int `json:"scope_extracted_counts"`
}

// ModelUsageInfo tracks information about a model used during extraction
type ModelUsageInfo struct {
	// ModelType identifies the model category (face_embedding, ocr, liveness, etc.)
	ModelType string `json:"model_type"`

	// ModelName is the specific model name
	ModelName string `json:"model_name"`

	// ModelVersion is the model version
	ModelVersion string `json:"model_version"`

	// ModelHash is the SHA256 hash of model weights (if available)
	ModelHash string `json:"model_hash,omitempty"`
}

// QualityGatesSummary provides aggregate quality gate information
type QualityGatesSummary struct {
	// TotalGates is the total number of quality gates checked
	TotalGates int `json:"total_gates"`

	// PassedGates is the number of gates that passed
	PassedGates int `json:"passed_gates"`

	// FailedGates is the number of gates that failed
	FailedGates int `json:"failed_gates"`

	// FailedGateTypes lists the types of gates that failed
	FailedGateTypes []string `json:"failed_gate_types,omitempty"`

	// OverallPassRate is the percentage of gates that passed
	OverallPassRate float32 `json:"overall_pass_rate"`
}

// ExtractionError represents a non-fatal error during extraction
type ExtractionError struct {
	// ScopeID is the scope that caused the error (if applicable)
	ScopeID string `json:"scope_id,omitempty"`

	// ScopeType is the type of scope (if applicable)
	ScopeType string `json:"scope_type,omitempty"`

	// ErrorCode is a machine-readable error code
	ErrorCode string `json:"error_code"`

	// ErrorMessage is a human-readable error message
	ErrorMessage string `json:"error_message"`
}

// ============================================================================
// Constructor and Methods
// ============================================================================

// NewFeatureExtractionMetadata creates a new metadata instance
func NewFeatureExtractionMetadata(schemaVersion, accountAddress string, blockHeight int64) *FeatureExtractionMetadata {
	return &FeatureExtractionMetadata{
		SchemaVersion:        schemaVersion,
		ExtractionTimestamp:  time.Now().UTC(),
		AccountAddress:       accountAddress,
		BlockHeight:          blockHeight,
		ModelsUsed:           make([]ModelUsageInfo, 0),
		Errors:               make([]ExtractionError, 0),
		ScopeExtractedCounts: make(map[string]int),
	}
}

// AddModelUsed adds a model usage record
func (m *FeatureExtractionMetadata) AddModelUsed(modelType, modelName, modelVersion string) {
	m.ModelsUsed = append(m.ModelsUsed, ModelUsageInfo{
		ModelType:    modelType,
		ModelName:    modelName,
		ModelVersion: modelVersion,
	})
}

// AddModelUsedWithHash adds a model usage record including model hash
func (m *FeatureExtractionMetadata) AddModelUsedWithHash(modelType, modelName, modelVersion, modelHash string) {
	m.ModelsUsed = append(m.ModelsUsed, ModelUsageInfo{
		ModelType:    modelType,
		ModelName:    modelName,
		ModelVersion: modelVersion,
		ModelHash:    modelHash,
	})
}

// AddError adds an extraction error
func (m *FeatureExtractionMetadata) AddError(scopeID, errorMessage string) {
	m.Errors = append(m.Errors, ExtractionError{
		ScopeID:      scopeID,
		ErrorCode:    "EXTRACTION_ERROR",
		ErrorMessage: errorMessage,
	})
}

// AddScopeTypeError adds an error with scope type
func (m *FeatureExtractionMetadata) AddScopeTypeError(scopeType, errorCode, errorMessage string) {
	m.Errors = append(m.Errors, ExtractionError{
		ScopeType:    scopeType,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	})
}

// IncrementScopeCount increments the count for a scope type
func (m *FeatureExtractionMetadata) IncrementScopeCount(scopeType string) {
	m.ScopeExtractedCounts[scopeType]++
}

// SetQualityGatesSummary sets the quality gates summary
func (m *FeatureExtractionMetadata) SetQualityGatesSummary(total, passed int, failedTypes []string) {
	failed := total - passed
	passRate := float32(0)
	if total > 0 {
		passRate = float32(passed) / float32(total)
	}

	m.QualityGatesSummary = &QualityGatesSummary{
		TotalGates:      total,
		PassedGates:     passed,
		FailedGates:     failed,
		FailedGateTypes: failedTypes,
		OverallPassRate: passRate,
	}

	m.QualityGatesPassed = failed == 0
}

// ComputeFeatureHash computes and sets the feature hash
func (m *FeatureExtractionMetadata) ComputeFeatureHash(faceEmbedding []float32, docQualityScore float32) {
	h := sha256.New()

	// Include schema version
	h.Write([]byte(m.SchemaVersion))

	// Include face embedding
	for _, v := range faceEmbedding {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, math.Float32bits(v))
		h.Write(b)
	}

	// Include doc quality
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, math.Float32bits(docQualityScore))
	h.Write(b)

	// Include block height for temporal binding
	blockBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockBytes, safeUint64FromInt64(m.BlockHeight))
	h.Write(blockBytes)

	m.FeatureHash = hex.EncodeToString(h.Sum(nil))
}

// HasErrors returns true if there were any extraction errors
func (m *FeatureExtractionMetadata) HasErrors() bool {
	return len(m.Errors) > 0
}

// GetModelVersions returns a map of model types to their versions
func (m *FeatureExtractionMetadata) GetModelVersions() map[string]string {
	versions := make(map[string]string)
	for _, model := range m.ModelsUsed {
		versions[model.ModelType] = model.ModelVersion
	}
	return versions
}

// Validate validates the metadata
func (m *FeatureExtractionMetadata) Validate() error {
	if m.SchemaVersion == "" {
		return ErrInvalidParams.Wrap("schema_version cannot be empty")
	}
	if m.AccountAddress == "" {
		return ErrInvalidParams.Wrap("account_address cannot be empty")
	}
	if m.BlockHeight < 0 {
		return ErrInvalidParams.Wrap("block_height cannot be negative")
	}
	return nil
}

func safeUint64FromInt64(value int64) uint64 {
	if value < 0 {
		return 0
	}
	//nolint:gosec // range checked above
	return uint64(value)
}

// ============================================================================
// Audit Trail Types
// ============================================================================

// FeatureExtractionAuditRecord is the full audit record for feature extraction
type FeatureExtractionAuditRecord struct {
	// RequestID is the verification request that triggered extraction
	RequestID string `json:"request_id"`

	// Metadata contains the extraction metadata
	Metadata *FeatureExtractionMetadata `json:"metadata"`

	// InputHash is the hash of the decrypted scope contents
	InputHash string `json:"input_hash"`

	// OutputHash is the hash of the extracted features
	OutputHash string `json:"output_hash"`

	// ValidatorAddress is the validator that performed extraction
	ValidatorAddress string `json:"validator_address"`

	// Signature is the validator's signature over this record
	Signature string `json:"signature,omitempty"`
}

// NewFeatureExtractionAuditRecord creates a new audit record
func NewFeatureExtractionAuditRecord(
	requestID string,
	metadata *FeatureExtractionMetadata,
	inputHash string,
	validatorAddress string,
) *FeatureExtractionAuditRecord {
	return &FeatureExtractionAuditRecord{
		RequestID:        requestID,
		Metadata:         metadata,
		InputHash:        inputHash,
		OutputHash:       metadata.FeatureHash,
		ValidatorAddress: validatorAddress,
	}
}

// ComputeRecordHash computes a hash of the audit record for signing
func (r *FeatureExtractionAuditRecord) ComputeRecordHash() string {
	h := sha256.New()
	h.Write([]byte(r.RequestID))
	h.Write([]byte(r.InputHash))
	h.Write([]byte(r.OutputHash))
	h.Write([]byte(r.ValidatorAddress))
	if r.Metadata != nil {
		h.Write([]byte(r.Metadata.SchemaVersion))
		h.Write([]byte(r.Metadata.FeatureHash))
	}
	return hex.EncodeToString(h.Sum(nil))
}
