// Package types provides types for the VEID module.
//
// VE-219: Deterministic identity verification runtime - pinned containers + reproducible builds
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"time"
)

// ============================================================================
// Pipeline Version Types
// ============================================================================

// PipelineVersion represents a versioned identity verification pipeline
// that validators must use for consensus. All validators must run the exact
// same pipeline version to produce reproducible verification results.
type PipelineVersion struct {
	// Version is the semantic version of the pipeline (e.g., "1.0.0")
	Version string `json:"version" protobuf:"bytes,1,opt,name=version,proto3"`

	// ImageHash is the SHA256 hash of the OCI container image
	// Format: sha256:<64-char-hex>
	ImageHash string `json:"image_hash" protobuf:"bytes,2,opt,name=image_hash,json=imageHash,proto3"`

	// ImageRef is the full OCI image reference (e.g., "ghcr.io/virtengine/veid-pipeline:v1.0.0")
	ImageRef string `json:"image_ref" protobuf:"bytes,3,opt,name=image_ref,json=imageRef,proto3"`

	// ModelManifest contains hashes of all ML model weights
	ModelManifest ModelManifest `json:"model_manifest" protobuf:"bytes,4,opt,name=model_manifest,json=modelManifest,proto3"`

	// CreatedAt is when this pipeline version was registered
	CreatedAt time.Time `json:"created_at" protobuf:"bytes,5,opt,name=created_at,json=createdAt,proto3,stdtime"`

	// CreatedAtHeight is the block height when registered
	CreatedAtHeight int64 `json:"created_at_height" protobuf:"varint,6,opt,name=created_at_height,json=createdAtHeight,proto3"`

	// ActivatedAt is when this pipeline version became active (nil if pending)
	ActivatedAt *time.Time `json:"activated_at,omitempty" protobuf:"bytes,7,opt,name=activated_at,json=activatedAt,proto3,stdtime"`

	// ActivatedAtHeight is the block height when activated (0 if pending)
	ActivatedAtHeight int64 `json:"activated_at_height,omitempty" protobuf:"varint,8,opt,name=activated_at_height,json=activatedAtHeight,proto3"`

	// DeprecatedAt is when this pipeline version was deprecated (nil if active)
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty" protobuf:"bytes,9,opt,name=deprecated_at,json=deprecatedAt,proto3,stdtime"`

	// Status is the current status of the pipeline version
	Status string `json:"status" protobuf:"bytes,10,opt,name=status,proto3"`

	// MinimumValidatorVersion is the minimum validator software version required
	MinimumValidatorVersion string `json:"minimum_validator_version,omitempty" protobuf:"bytes,11,opt,name=minimum_validator_version,json=minimumValidatorVersion,proto3"`

	// DeterminismConfig contains configuration for deterministic execution
	DeterminismConfig PipelineDeterminismConfig `json:"determinism_config" protobuf:"bytes,12,opt,name=determinism_config,json=determinismConfig,proto3"`
}

// PipelineVersionStatus represents the status of a pipeline version
type PipelineVersionStatus string

const (
	// PipelineVersionStatusPending indicates the version is pending activation
	PipelineVersionStatusPending PipelineVersionStatus = "pending"

	// PipelineVersionStatusActive indicates the version is active and can be used
	PipelineVersionStatusActive PipelineVersionStatus = "active"

	// PipelineVersionStatusDeprecated indicates the version is deprecated
	PipelineVersionStatusDeprecated PipelineVersionStatus = "deprecated"

	// PipelineVersionStatusRetired indicates the version is no longer valid for new verifications
	PipelineVersionStatusRetired PipelineVersionStatus = "retired"
)

// AllPipelineVersionStatuses returns all valid pipeline version statuses
func AllPipelineVersionStatuses() []PipelineVersionStatus {
	return []PipelineVersionStatus{
		PipelineVersionStatusPending,
		PipelineVersionStatusActive,
		PipelineVersionStatusDeprecated,
		PipelineVersionStatusRetired,
	}
}

// IsValidPipelineVersionStatus checks if a status is valid
func IsValidPipelineVersionStatus(status PipelineVersionStatus) bool {
	for _, s := range AllPipelineVersionStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// ============================================================================
// Model Manifest Types
// ============================================================================

// ModelManifest contains versioned hashes for all ML models used in the pipeline
type ModelManifest struct {
	// Version is the manifest format version
	Version string `json:"version" protobuf:"bytes,1,opt,name=version,proto3"`

	// Models is a list of model info (keyed by Name field)
	Models []ModelInfo `json:"models" protobuf:"bytes,2,rep,name=models,proto3"`

	// ManifestHash is the combined hash of all model hashes
	ManifestHash string `json:"manifest_hash" protobuf:"bytes,3,opt,name=manifest_hash,json=manifestHash,proto3"`

	// CreatedAt is when the manifest was created
	CreatedAt time.Time `json:"created_at" protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3,stdtime"`
}

// ModelInfo contains information about a single ML model
type ModelInfo struct {
	// Name is the model name (e.g., "deepface_facenet512")
	Name string `json:"name" protobuf:"bytes,1,opt,name=name,proto3"`

	// Version is the model version (e.g., "1.0.0")
	Version string `json:"version" protobuf:"bytes,2,opt,name=version,proto3"`

	// WeightsHash is the SHA256 hash of the model weights file
	WeightsHash string `json:"weights_hash" protobuf:"bytes,3,opt,name=weights_hash,json=weightsHash,proto3"`

	// ConfigHash is the SHA256 hash of the model configuration
	ConfigHash string `json:"config_hash,omitempty" protobuf:"bytes,4,opt,name=config_hash,json=configHash,proto3"`

	// Framework is the ML framework used (e.g., "tensorflow", "onnx")
	Framework string `json:"framework" protobuf:"bytes,5,opt,name=framework,proto3"`

	// InputShape describes the expected input tensor shape
	InputShape []int32 `json:"input_shape,omitempty" protobuf:"varint,6,rep,packed,name=input_shape,json=inputShape,proto3"`

	// OutputShape describes the output tensor shape
	OutputShape []int32 `json:"output_shape,omitempty" protobuf:"varint,7,rep,packed,name=output_shape,json=outputShape,proto3"`

	// Purpose describes what the model is used for
	Purpose string `json:"purpose" protobuf:"bytes,8,opt,name=purpose,proto3"`
}

// ModelPurpose describes the purpose of a model in the pipeline
type ModelPurpose string

const (
	// ModelPurposeFaceDetection for face detection models
	ModelPurposeFaceDetection ModelPurpose = "face_detection"

	// ModelPurposeFaceRecognition for face recognition/embedding models
	ModelPurposeFaceRecognition ModelPurpose = "face_recognition"

	// ModelPurposeFaceVerification for face comparison models
	ModelPurposeFaceVerification ModelPurpose = "face_verification"

	// ModelPurposeTextDetection for text/ROI detection (CRAFT)
	ModelPurposeTextDetection ModelPurpose = "text_detection"

	// ModelPurposeOCR for optical character recognition (Tesseract)
	ModelPurposeOCR ModelPurpose = "ocr"

	// ModelPurposeDocumentSegmentation for document region segmentation (U-Net)
	ModelPurposeDocumentSegmentation ModelPurpose = "document_segmentation"

	// ModelPurposeFaceExtraction for extracting face from ID documents
	ModelPurposeFaceExtraction ModelPurpose = "face_extraction"

	// ModelPurposeIdentityScoring for final identity score computation
	ModelPurposeIdentityScoring ModelPurpose = "identity_scoring"
)

// ============================================================================
// Pipeline Determinism Config
// ============================================================================

// PipelineDeterminismConfig contains configuration for deterministic execution
type PipelineDeterminismConfig struct {
	// RandomSeed is the fixed random seed for all operations
	RandomSeed int64 `json:"random_seed" protobuf:"varint,1,opt,name=random_seed,json=randomSeed,proto3"`

	// ForceCPU ensures CPU-only execution (GPUs can introduce non-determinism)
	ForceCPU bool `json:"force_cpu" protobuf:"varint,2,opt,name=force_cpu,json=forceCpu,proto3"`

	// SingleThread disables multi-threading for determinism
	SingleThread bool `json:"single_thread" protobuf:"varint,3,opt,name=single_thread,json=singleThread,proto3"`

	// FloatPrecision is the number of decimal places for float comparisons
	FloatPrecision int32 `json:"float_precision" protobuf:"varint,4,opt,name=float_precision,json=floatPrecision,proto3"`

	// TensorFlowDeterministic enables TensorFlow deterministic mode
	TensorFlowDeterministic bool `json:"tensorflow_deterministic" protobuf:"varint,5,opt,name=tensorflow_deterministic,json=tensorflowDeterministic,proto3"`

	// DisableCUDNN disables cuDNN which can be non-deterministic
	DisableCUDNN bool `json:"disable_cudnn" protobuf:"varint,6,opt,name=disable_cudnn,json=disableCudnn,proto3"`

	// ONNXDeterministic enables ONNX runtime deterministic mode
	ONNXDeterministic bool `json:"onnx_deterministic" protobuf:"varint,7,opt,name=onnx_deterministic,json=onnxDeterministic,proto3"`
}

// DefaultPipelineDeterminismConfig returns the default determinism configuration
func DefaultPipelineDeterminismConfig() PipelineDeterminismConfig {
	return PipelineDeterminismConfig{
		RandomSeed:              42,
		ForceCPU:                true,
		SingleThread:            true,
		FloatPrecision:          6,
		TensorFlowDeterministic: true,
		DisableCUDNN:            true,
		ONNXDeterministic:       true,
	}
}

// ============================================================================
// Verification Record Extension
// ============================================================================

// PipelineExecutionRecord records the pipeline version and model hashes
// used during a specific verification. This is stored with the verification
// result to enable consensus verification.
type PipelineExecutionRecord struct {
	// PipelineVersion is the version of the pipeline used
	PipelineVersion string `json:"pipeline_version" protobuf:"bytes,1,opt,name=pipeline_version,json=pipelineVersion,proto3"`

	// PipelineImageHash is the SHA256 hash of the OCI image used
	PipelineImageHash string `json:"pipeline_image_hash" protobuf:"bytes,2,opt,name=pipeline_image_hash,json=pipelineImageHash,proto3"`

	// ModelManifestHash is the combined hash of all model hashes
	ModelManifestHash string `json:"model_manifest_hash" protobuf:"bytes,3,opt,name=model_manifest_hash,json=modelManifestHash,proto3"`

	// ExecutedAt is when the pipeline was executed
	ExecutedAt time.Time `json:"executed_at" protobuf:"bytes,4,opt,name=executed_at,json=executedAt,proto3,stdtime"`

	// ExecutionDurationMs is how long execution took in milliseconds
	ExecutionDurationMs int64 `json:"execution_duration_ms" protobuf:"varint,5,opt,name=execution_duration_ms,json=executionDurationMs,proto3"`

	// DeterminismVerified indicates if determinism checks passed
	DeterminismVerified bool `json:"determinism_verified" protobuf:"varint,6,opt,name=determinism_verified,json=determinismVerified,proto3"`

	// InputHash is the SHA256 hash of all inputs (for consensus)
	InputHash string `json:"input_hash" protobuf:"bytes,7,opt,name=input_hash,json=inputHash,proto3"`

	// OutputHash is the SHA256 hash of all outputs (for consensus)
	OutputHash string `json:"output_hash" protobuf:"bytes,8,opt,name=output_hash,json=outputHash,proto3"`
}

// ============================================================================
// Validation Functions
// ============================================================================

var (
	// semverRegex validates semantic versioning format
	semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(-[a-zA-Z0-9]+)?$`)

	// sha256Regex validates SHA256 hash format
	sha256Regex = regexp.MustCompile(`^(sha256:)?[a-fA-F0-9]{64}$`)

	// imageRefRegex validates OCI image reference format
	imageRefRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*(\/[a-zA-Z0-9._-]+)+:[a-zA-Z0-9._-]+(@sha256:[a-fA-F0-9]{64})?$`)
)

// Validate validates a PipelineVersion
func (pv *PipelineVersion) Validate() error {
	if pv.Version == "" {
		return ErrInvalidPipelineVersion.Wrap("version cannot be empty")
	}

	if !semverRegex.MatchString(pv.Version) {
		return ErrInvalidPipelineVersion.Wrapf("invalid version format: %s", pv.Version)
	}

	if pv.ImageHash == "" {
		return ErrInvalidPipelineVersion.Wrap("image_hash cannot be empty")
	}

	if !sha256Regex.MatchString(pv.ImageHash) {
		return ErrInvalidPipelineVersion.Wrapf("invalid image_hash format: %s", pv.ImageHash)
	}

	if pv.ImageRef == "" {
		return ErrInvalidPipelineVersion.Wrap("image_ref cannot be empty")
	}

	if !imageRefRegex.MatchString(pv.ImageRef) {
		return ErrInvalidPipelineVersion.Wrapf("invalid image_ref format: %s", pv.ImageRef)
	}

	if err := pv.ModelManifest.Validate(); err != nil {
		return ErrInvalidPipelineVersion.Wrapf("invalid model manifest: %v", err)
	}

	if !IsValidPipelineVersionStatus(PipelineVersionStatus(pv.Status)) {
		return ErrInvalidPipelineVersion.Wrapf("invalid status: %s", pv.Status)
	}

	return nil
}

// Validate validates a ModelManifest
func (mm *ModelManifest) Validate() error {
	if mm.Version == "" {
		return fmt.Errorf("manifest version cannot be empty")
	}

	if len(mm.Models) == 0 {
		return fmt.Errorf("manifest must contain at least one model")
	}

	seen := make(map[string]bool)
	for _, model := range mm.Models {
		if model.Name == "" {
			return fmt.Errorf("model name cannot be empty")
		}
		if seen[model.Name] {
			return fmt.Errorf("duplicate model name: %s", model.Name)
		}
		seen[model.Name] = true
		if err := model.Validate(); err != nil {
			return fmt.Errorf("invalid model %s: %w", model.Name, err)
		}
	}

	// Verify manifest hash
	computedHash := mm.ComputeHash()
	if mm.ManifestHash != "" && mm.ManifestHash != computedHash {
		return fmt.Errorf("manifest hash mismatch: expected %s, got %s", mm.ManifestHash, computedHash)
	}

	return nil
}

// Validate validates a ModelInfo
func (mi *ModelInfo) Validate() error {
	if mi.Name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	if mi.Version == "" {
		return fmt.Errorf("model version cannot be empty")
	}

	if mi.WeightsHash == "" {
		return fmt.Errorf("weights_hash cannot be empty")
	}

	if !sha256Regex.MatchString(mi.WeightsHash) {
		return fmt.Errorf("invalid weights_hash format: %s", mi.WeightsHash)
	}

	if mi.Framework == "" {
		return fmt.Errorf("framework cannot be empty")
	}

	return nil
}

// ============================================================================
// Hash Computation Functions
// ============================================================================

// ComputeHash computes the combined SHA256 hash of all model hashes
func (mm *ModelManifest) ComputeHash() string {
	h := sha256.New()

	// Write manifest version
	h.Write([]byte(mm.Version))

	// Get sorted model names for deterministic ordering
	names := make([]string, 0, len(mm.Models))
	for _, model := range mm.Models {
		names = append(names, model.Name)
	}
	sort.Strings(names)

	// Create a lookup map for getting model by name
	modelByName := make(map[string]ModelInfo)
	for _, model := range mm.Models {
		modelByName[model.Name] = model
	}

	// Hash each model's info
	for _, name := range names {
		model := modelByName[name]
		h.Write([]byte(name))
		h.Write([]byte(model.Version))
		h.Write([]byte(model.WeightsHash))
		if model.ConfigHash != "" {
			h.Write([]byte(model.ConfigHash))
		}
		h.Write([]byte(model.Framework))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ComputePipelineHash computes the combined hash of pipeline image and models
func (pv *PipelineVersion) ComputePipelineHash() string {
	h := sha256.New()
	h.Write([]byte(pv.Version))
	h.Write([]byte(pv.ImageHash))
	h.Write([]byte(pv.ModelManifest.ComputeHash()))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Constructor Functions
// ============================================================================

// NewPipelineVersion creates a new PipelineVersion
func NewPipelineVersion(
	version string,
	imageHash string,
	imageRef string,
	modelManifest ModelManifest,
	createdAt time.Time,
	createdAtHeight int64,
) *PipelineVersion {
	return &PipelineVersion{
		Version:           version,
		ImageHash:         imageHash,
		ImageRef:          imageRef,
		ModelManifest:     modelManifest,
		CreatedAt:         createdAt,
		CreatedAtHeight:   createdAtHeight,
		Status:            string(PipelineVersionStatusPending),
		DeterminismConfig: DefaultPipelineDeterminismConfig(),
	}
}

// NewModelManifest creates a new ModelManifest
func NewModelManifest(version string, models []ModelInfo, createdAt time.Time) *ModelManifest {
	mm := &ModelManifest{
		Version:   version,
		Models:    models,
		CreatedAt: createdAt,
	}
	mm.ManifestHash = mm.ComputeHash()
	return mm
}

// NewModelInfo creates a new ModelInfo
func NewModelInfo(
	name string,
	version string,
	weightsHash string,
	framework string,
	purpose ModelPurpose,
) *ModelInfo {
	return &ModelInfo{
		Name:        name,
		Version:     version,
		WeightsHash: weightsHash,
		Framework:   framework,
		Purpose:     string(purpose),
	}
}

// NewPipelineExecutionRecord creates a new PipelineExecutionRecord
func NewPipelineExecutionRecord(
	pipelineVersion string,
	pipelineImageHash string,
	modelManifestHash string,
	executedAt time.Time,
) *PipelineExecutionRecord {
	return &PipelineExecutionRecord{
		PipelineVersion:   pipelineVersion,
		PipelineImageHash: pipelineImageHash,
		ModelManifestHash: modelManifestHash,
		ExecutedAt:        executedAt,
	}
}

// ============================================================================
// Comparison Functions
// ============================================================================

// MatchesPipelineVersion checks if an execution record matches a pipeline version
func (per *PipelineExecutionRecord) MatchesPipelineVersion(pv *PipelineVersion) bool {
	if per.PipelineVersion != pv.Version {
		return false
	}

	if per.PipelineImageHash != pv.ImageHash {
		return false
	}

	if per.ModelManifestHash != pv.ModelManifest.ManifestHash {
		return false
	}

	return true
}

// CompareExecutionRecords compares two execution records for consensus
func CompareExecutionRecords(a, b *PipelineExecutionRecord) *PipelineComparisonResult {
	result := &PipelineComparisonResult{
		Match:       true,
		Differences: make([]string, 0),
	}

	if a.PipelineVersion != b.PipelineVersion {
		result.Match = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("pipeline version mismatch: %s vs %s", a.PipelineVersion, b.PipelineVersion))
	}

	if a.PipelineImageHash != b.PipelineImageHash {
		result.Match = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("image hash mismatch: %s vs %s", a.PipelineImageHash, b.PipelineImageHash))
	}

	if a.ModelManifestHash != b.ModelManifestHash {
		result.Match = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("model manifest hash mismatch: %s vs %s", a.ModelManifestHash, b.ModelManifestHash))
	}

	if a.InputHash != b.InputHash {
		result.Match = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("input hash mismatch: %s vs %s", a.InputHash, b.InputHash))
	}

	if a.OutputHash != b.OutputHash {
		result.Match = false
		result.Differences = append(result.Differences,
			fmt.Sprintf("output hash mismatch: %s vs %s", a.OutputHash, b.OutputHash))
	}

	return result
}

// PipelineComparisonResult contains the result of comparing two execution records
type PipelineComparisonResult struct {
	// Match indicates if the records match
	Match bool `json:"match"`

	// Differences contains descriptions of any differences found
	Differences []string `json:"differences,omitempty"`
}

// ============================================================================
// Proto.Message interface stubs for pipeline types
// ============================================================================

// PipelineVersion proto stubs
func (*PipelineVersion) ProtoMessage()    {}
func (m *PipelineVersion) Reset()         { *m = PipelineVersion{} }
func (m *PipelineVersion) String() string { return fmt.Sprintf("%+v", *m) }

// ModelManifest proto stubs
func (*ModelManifest) ProtoMessage()    {}
func (m *ModelManifest) Reset()         { *m = ModelManifest{} }
func (m *ModelManifest) String() string { return fmt.Sprintf("%+v", *m) }

// ModelInfo proto stubs
func (*ModelInfo) ProtoMessage()    {}
func (m *ModelInfo) Reset()         { *m = ModelInfo{} }
func (m *ModelInfo) String() string { return fmt.Sprintf("%+v", *m) }

// PipelineDeterminismConfig proto stubs
func (*PipelineDeterminismConfig) ProtoMessage()    {}
func (m *PipelineDeterminismConfig) Reset()         { *m = PipelineDeterminismConfig{} }
func (m *PipelineDeterminismConfig) String() string { return fmt.Sprintf("%+v", *m) }

// PipelineExecutionRecord proto stubs
func (*PipelineExecutionRecord) ProtoMessage()    {}
func (m *PipelineExecutionRecord) Reset()         { *m = PipelineExecutionRecord{} }
func (m *PipelineExecutionRecord) String() string { return fmt.Sprintf("%+v", *m) }

// PipelineComparisonResult proto stubs
func (*PipelineComparisonResult) ProtoMessage()    {}
func (m *PipelineComparisonResult) Reset()         { *m = PipelineComparisonResult{} }
func (m *PipelineComparisonResult) String() string { return fmt.Sprintf("%+v", *m) }
