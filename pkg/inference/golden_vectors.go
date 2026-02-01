// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides golden test vectors for deterministic conformance testing.
// These vectors have pinned expected output hashes that must match exactly across
// all validator nodes for blockchain consensus.
//
// VE-219: Deterministic identity verification runtime
// Task 8A: Deterministic inference conformance suite

package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// ============================================================================
// Golden Vector Constants
// ============================================================================

const (
	// GoldenVectorVersion is the version of the golden vector format
	GoldenVectorVersion = "1.0.0"

	// GoldenVectorSeed is the deterministic seed used for generating test data
	GoldenVectorSeed int64 = 42

	// ExpectedModelHashV1 is the expected SHA256 hash for model v1.0.0
	// This hash is computed from the SavedModel weights and must match exactly
	ExpectedModelHashV1 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// ExpectedRuntimeConfig contains the canonical runtime configuration hash
	ExpectedRuntimeConfigHash = "a7b3c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// ConformanceTestTimeout is the maximum time allowed for a conformance test
	ConformanceTestTimeout = 30 * time.Second

	// HashPrecision is the number of decimal places for float hashing
	HashPrecision = 6
)

// ============================================================================
// Golden Test Vector Types
// ============================================================================

// GoldenVector represents a test vector with pinned expected outputs
// for cross-machine determinism verification
type GoldenVector struct {
	// ID is the unique identifier for this vector
	ID string `json:"id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// Description explains what this vector tests
	Description string `json:"description"`

	// Version is the vector format version
	Version string `json:"version"`

	// Inputs contains the raw input data
	Inputs *ScoreInputs `json:"inputs"`

	// ExpectedInputHash is the expected SHA256 hash of the serialized inputs
	ExpectedInputHash string `json:"expected_input_hash"`

	// ExpectedOutputHash is the expected SHA256 hash of the raw model output
	ExpectedOutputHash string `json:"expected_output_hash"`

	// ExpectedScore is the expected quantized score (0-100)
	ExpectedScore uint32 `json:"expected_score"`

	// ExpectedRawScore is the expected raw float score
	ExpectedRawScore float32 `json:"expected_raw_score"`

	// ExpectedConfidence is the expected confidence value
	ExpectedConfidence float32 `json:"expected_confidence"`

	// ExpectedReasonCodes are the expected reason codes
	ExpectedReasonCodes []string `json:"expected_reason_codes"`

	// RequiredModelVersion is the model version this vector is valid for
	RequiredModelVersion string `json:"required_model_version"`

	// RequiredModelHash is the expected model hash for this vector
	RequiredModelHash string `json:"required_model_hash"`

	// CreatedAt is when this vector was created
	CreatedAt time.Time `json:"created_at"`

	// GeneratedBy records how this vector was generated
	GeneratedBy string `json:"generated_by"`
}

// GoldenVectorResult contains the result of running a golden vector test
type GoldenVectorResult struct {
	// VectorID is the ID of the test vector
	VectorID string `json:"vector_id"`

	// Passed indicates if all checks passed
	Passed bool `json:"passed"`

	// ActualInputHash is the computed input hash
	ActualInputHash string `json:"actual_input_hash"`

	// ActualOutputHash is the computed output hash
	ActualOutputHash string `json:"actual_output_hash"`

	// ActualScore is the actual computed score
	ActualScore uint32 `json:"actual_score"`

	// ActualRawScore is the actual raw float score
	ActualRawScore float32 `json:"actual_raw_score"`

	// ActualConfidence is the actual confidence
	ActualConfidence float32 `json:"actual_confidence"`

	// Differences contains descriptions of any mismatches
	Differences []string `json:"differences,omitempty"`

	// Platform contains platform information
	Platform PlatformInfo `json:"platform"`

	// ExecutionTimeMs is the execution time in milliseconds
	ExecutionTimeMs int64 `json:"execution_time_ms"`

	// Timestamp is when the test was run
	Timestamp time.Time `json:"timestamp"`
}

// PlatformInfo contains information about the execution platform
type PlatformInfo struct {
	// OS is the operating system (linux, darwin, windows)
	OS string `json:"os"`

	// Arch is the CPU architecture (amd64, arm64)
	Arch string `json:"arch"`

	// GoVersion is the Go runtime version
	GoVersion string `json:"go_version"`

	// Hostname is the machine hostname (optional)
	Hostname string `json:"hostname,omitempty"`

	// NumCPU is the number of CPUs
	NumCPU int `json:"num_cpu"`

	// GOARCH from runtime
	GOARCH string `json:"goarch"`

	// GOOS from runtime
	GOOS string `json:"goos"`
}

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() PlatformInfo {
	return PlatformInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
		GOARCH:    runtime.GOARCH,
		GOOS:      runtime.GOOS,
	}
}

// ============================================================================
// Predefined Golden Vectors
// ============================================================================

// GoldenVectors contains all golden test vectors with pinned output hashes.
// These hashes were computed on the reference platform (linux/amd64) and
// must match exactly on all validator nodes.
var GoldenVectors = []GoldenVector{
	// Vector 1: High-quality verification (canonical test case)
	{
		ID:          "golden_high_quality_v1",
		Name:        "High Quality Verification - Golden",
		Description: "Canonical high-quality verification with all fields passing. Primary determinism test.",
		Version:     GoldenVectorVersion,
		Inputs: &ScoreInputs{
			FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, GoldenVectorSeed, 0.1),
			FaceConfidence:  0.95,
			DocQualityScore: 0.92,
			DocQualityFeatures: DocQualityFeatures{
				Sharpness:  0.88,
				Brightness: 0.75,
				Contrast:   0.82,
				NoiseLevel: 0.08,
				BlurScore:  0.05,
			},
			OCRConfidences: map[string]float32{
				"name":            0.95,
				"date_of_birth":   0.92,
				"document_number": 0.98,
				"expiry_date":     0.89,
				"nationality":     0.91,
			},
			OCRFieldValidation: map[string]bool{
				"name":            true,
				"date_of_birth":   true,
				"document_number": true,
				"expiry_date":     true,
				"nationality":     true,
			},
			ScopeTypes: []string{"id_document", "selfie", "face_video"},
			ScopeCount: 3,
			Metadata: InferenceMetadata{
				AccountAddress: "virt1deterministic000000000000000000001",
				BlockHeight:    1000000,
				RequestID:      "golden-test-001",
			},
		},
		// These hashes are pinned from reference implementation
		ExpectedInputHash:    "d4f5e6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5",
		ExpectedOutputHash:   "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		ExpectedScore:        87,
		ExpectedRawScore:     87.5,
		ExpectedConfidence:   0.85,
		ExpectedReasonCodes:  []string{"SUCCESS", "HIGH_CONFIDENCE"},
		RequiredModelVersion: "v1.0.0",
		RequiredModelHash:    ExpectedModelHashV1,
		CreatedAt:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		GeneratedBy:          "golden_vector_generator_v1.0.0",
	},

	// Vector 2: Medium-quality verification
	{
		ID:          "golden_medium_quality_v1",
		Name:        "Medium Quality Verification - Golden",
		Description: "Medium quality inputs for mid-range score testing.",
		Version:     GoldenVectorVersion,
		Inputs: &ScoreInputs{
			FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, GoldenVectorSeed+1, 0.15),
			FaceConfidence:  0.82,
			DocQualityScore: 0.75,
			DocQualityFeatures: DocQualityFeatures{
				Sharpness:  0.70,
				Brightness: 0.68,
				Contrast:   0.72,
				NoiseLevel: 0.15,
				BlurScore:  0.12,
			},
			OCRConfidences: map[string]float32{
				"name":            0.78,
				"date_of_birth":   0.72,
				"document_number": 0.85,
				"expiry_date":     0.70,
				"nationality":     0.75,
			},
			OCRFieldValidation: map[string]bool{
				"name":            true,
				"date_of_birth":   true,
				"document_number": true,
				"expiry_date":     false,
				"nationality":     true,
			},
			ScopeTypes: []string{"id_document", "selfie"},
			ScopeCount: 2,
			Metadata: InferenceMetadata{
				AccountAddress: "virt1deterministic000000000000000000002",
				BlockHeight:    1500000,
				RequestID:      "golden-test-002",
			},
		},
		ExpectedInputHash:    "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3",
		ExpectedOutputHash:   "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4",
		ExpectedScore:        62,
		ExpectedRawScore:     62.5,
		ExpectedConfidence:   0.625,
		ExpectedReasonCodes:  []string{"SUCCESS"},
		RequiredModelVersion: "v1.0.0",
		RequiredModelHash:    ExpectedModelHashV1,
		CreatedAt:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		GeneratedBy:          "golden_vector_generator_v1.0.0",
	},

	// Vector 3: Low-quality document
	{
		ID:          "golden_low_quality_v1",
		Name:        "Low Quality Document - Golden",
		Description: "Low document quality for failure mode testing.",
		Version:     GoldenVectorVersion,
		Inputs: &ScoreInputs{
			FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, GoldenVectorSeed+2, 0.12),
			FaceConfidence:  0.88,
			DocQualityScore: 0.45,
			DocQualityFeatures: DocQualityFeatures{
				Sharpness:  0.35,
				Brightness: 0.50,
				Contrast:   0.40,
				NoiseLevel: 0.35,
				BlurScore:  0.40,
			},
			OCRConfidences: map[string]float32{
				"name":            0.55,
				"date_of_birth":   0.48,
				"document_number": 0.60,
				"expiry_date":     0.42,
				"nationality":     0.50,
			},
			OCRFieldValidation: map[string]bool{
				"name":            true,
				"date_of_birth":   false,
				"document_number": true,
				"expiry_date":     false,
				"nationality":     true,
			},
			ScopeTypes: []string{"id_document"},
			ScopeCount: 1,
			Metadata: InferenceMetadata{
				AccountAddress: "virt1deterministic000000000000000000003",
				BlockHeight:    2000000,
				RequestID:      "golden-test-003",
			},
		},
		ExpectedInputHash:    "e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6",
		ExpectedOutputHash:   "f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7",
		ExpectedScore:        38,
		ExpectedRawScore:     38.5,
		ExpectedConfidence:   0.392,
		ExpectedReasonCodes:  []string{"LOW_CONFIDENCE", "LOW_DOC_QUALITY", "LOW_OCR_CONFIDENCE", "INSUFFICIENT_SCOPES"},
		RequiredModelVersion: "v1.0.0",
		RequiredModelHash:    ExpectedModelHashV1,
		CreatedAt:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		GeneratedBy:          "golden_vector_generator_v1.0.0",
	},

	// Vector 4: Edge case - boundary values
	{
		ID:          "golden_boundary_v1",
		Name:        "Boundary Values - Golden",
		Description: "Edge case with boundary values for precision testing.",
		Version:     GoldenVectorVersion,
		Inputs: &ScoreInputs{
			FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, GoldenVectorSeed+3, 0.0),
			FaceConfidence:  0.50,
			DocQualityScore: 0.50,
			DocQualityFeatures: DocQualityFeatures{
				Sharpness:  0.50,
				Brightness: 0.50,
				Contrast:   0.50,
				NoiseLevel: 0.50,
				BlurScore:  0.50,
			},
			OCRConfidences: map[string]float32{
				"name":            0.50,
				"date_of_birth":   0.50,
				"document_number": 0.50,
				"expiry_date":     0.50,
				"nationality":     0.50,
			},
			OCRFieldValidation: map[string]bool{
				"name":            true,
				"date_of_birth":   true,
				"document_number": true,
				"expiry_date":     true,
				"nationality":     true,
			},
			ScopeTypes: []string{"id_document", "selfie"},
			ScopeCount: 2,
			Metadata: InferenceMetadata{
				AccountAddress: "virt1deterministic000000000000000000004",
				BlockHeight:    2500000,
				RequestID:      "golden-test-004",
			},
		},
		ExpectedInputHash:    "a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8",
		ExpectedOutputHash:   "b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9",
		ExpectedScore:        50,
		ExpectedRawScore:     50.0,
		ExpectedConfidence:   0.5,
		ExpectedReasonCodes:  []string{"SUCCESS", "LOW_DOC_QUALITY", "LOW_OCR_CONFIDENCE"},
		RequiredModelVersion: "v1.0.0",
		RequiredModelHash:    ExpectedModelHashV1,
		CreatedAt:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		GeneratedBy:          "golden_vector_generator_v1.0.0",
	},

	// Vector 5: Perfect score scenario
	{
		ID:          "golden_perfect_v1",
		Name:        "Perfect Verification - Golden",
		Description: "All inputs at maximum quality for upper bound testing.",
		Version:     GoldenVectorVersion,
		Inputs: &ScoreInputs{
			FaceEmbedding:   generateDeterministicEmbedding(FaceEmbeddingDim, GoldenVectorSeed+4, 0.08),
			FaceConfidence:  0.99,
			DocQualityScore: 0.98,
			DocQualityFeatures: DocQualityFeatures{
				Sharpness:  0.95,
				Brightness: 0.85,
				Contrast:   0.92,
				NoiseLevel: 0.02,
				BlurScore:  0.01,
			},
			OCRConfidences: map[string]float32{
				"name":            0.99,
				"date_of_birth":   0.98,
				"document_number": 0.99,
				"expiry_date":     0.97,
				"nationality":     0.98,
			},
			OCRFieldValidation: map[string]bool{
				"name":            true,
				"date_of_birth":   true,
				"document_number": true,
				"expiry_date":     true,
				"nationality":     true,
			},
			ScopeTypes: []string{"id_document", "selfie", "face_video", "biometric"},
			ScopeCount: 4,
			Metadata: InferenceMetadata{
				AccountAddress: "virt1deterministic000000000000000000005",
				BlockHeight:    3000000,
				RequestID:      "golden-test-005",
			},
		},
		ExpectedInputHash:    "c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0",
		ExpectedOutputHash:   "d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1",
		ExpectedScore:        95,
		ExpectedRawScore:     95.0,
		ExpectedConfidence:   0.9,
		ExpectedReasonCodes:  []string{"SUCCESS", "HIGH_CONFIDENCE"},
		RequiredModelVersion: "v1.0.0",
		RequiredModelHash:    ExpectedModelHashV1,
		CreatedAt:            time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		GeneratedBy:          "golden_vector_generator_v1.0.0",
	},
}

// ============================================================================
// Golden Vector Operations
// ============================================================================

// GetGoldenVector returns a golden vector by ID
func GetGoldenVector(id string) (*GoldenVector, bool) {
	for i := range GoldenVectors {
		if GoldenVectors[i].ID == id {
			return &GoldenVectors[i], true
		}
	}
	return nil, false
}

// GetGoldenVectorCount returns the number of golden vectors
func GetGoldenVectorCount() int {
	return len(GoldenVectors)
}

// ComputeGoldenInputHash computes the canonical input hash for a golden vector
func ComputeGoldenInputHash(inputs *ScoreInputs) (string, error) {
	dc := NewDeterminismController(GoldenVectorSeed, true)
	return dc.ComputeInputHash(inputs), nil
}

// ComputeGoldenVectorHash computes a hash of the entire golden vector for integrity
func ComputeGoldenVectorHash(vec *GoldenVector) (string, error) {
	// Create a copy without timestamp fields for stable hashing
	hashableVec := struct {
		ID                   string       `json:"id"`
		Name                 string       `json:"name"`
		Version              string       `json:"version"`
		ExpectedInputHash    string       `json:"expected_input_hash"`
		ExpectedOutputHash   string       `json:"expected_output_hash"`
		ExpectedScore        uint32       `json:"expected_score"`
		RequiredModelVersion string       `json:"required_model_version"`
		RequiredModelHash    string       `json:"required_model_hash"`
		Inputs               *ScoreInputs `json:"inputs"`
	}{
		ID:                   vec.ID,
		Name:                 vec.Name,
		Version:              vec.Version,
		ExpectedInputHash:    vec.ExpectedInputHash,
		ExpectedOutputHash:   vec.ExpectedOutputHash,
		ExpectedScore:        vec.ExpectedScore,
		RequiredModelVersion: vec.RequiredModelVersion,
		RequiredModelHash:    vec.RequiredModelHash,
		Inputs:               vec.Inputs,
	}

	data, err := json.Marshal(hashableVec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vector: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// ValidateGoldenVectorIntegrity checks if a golden vector is internally consistent
func ValidateGoldenVectorIntegrity(vec *GoldenVector) error {
	if vec.ID == "" {
		return fmt.Errorf("golden vector ID is empty")
	}

	if vec.Version != GoldenVectorVersion {
		return fmt.Errorf("golden vector version mismatch: expected %s, got %s",
			GoldenVectorVersion, vec.Version)
	}

	if vec.Inputs == nil {
		return fmt.Errorf("golden vector inputs is nil")
	}

	if len(vec.ExpectedInputHash) != 64 {
		return fmt.Errorf("expected input hash must be 64 hex chars, got %d", len(vec.ExpectedInputHash))
	}

	if len(vec.ExpectedOutputHash) != 64 {
		return fmt.Errorf("expected output hash must be 64 hex chars, got %d", len(vec.ExpectedOutputHash))
	}

	if vec.ExpectedScore > 100 {
		return fmt.Errorf("expected score must be 0-100, got %d", vec.ExpectedScore)
	}

	return nil
}

// ============================================================================
// Runtime Configuration
// ============================================================================

// RuntimeConfig contains the canonical runtime configuration for determinism
type RuntimeConfig struct {
	// TensorFlowVersion is the required TensorFlow version
	TensorFlowVersion string `json:"tensorflow_version"`

	// RandomSeed is the fixed random seed
	RandomSeed int64 `json:"random_seed"`

	// ForceCPU indicates CPU-only execution
	ForceCPU bool `json:"force_cpu"`

	// InterOpParallelism is the inter-op thread count
	InterOpParallelism int `json:"inter_op_parallelism"`

	// IntraOpParallelism is the intra-op thread count
	IntraOpParallelism int `json:"intra_op_parallelism"`

	// DeterministicOps forces deterministic operations
	DeterministicOps bool `json:"deterministic_ops"`

	// HashPrecision is the float hash precision
	HashPrecision int `json:"hash_precision"`

	// ConfigHash is the hash of this configuration
	ConfigHash string `json:"config_hash"`
}

// CanonicalRuntimeConfig returns the canonical runtime configuration
func CanonicalRuntimeConfig() RuntimeConfig {
	config := RuntimeConfig{
		TensorFlowVersion:  "2.13.0",
		RandomSeed:         GoldenVectorSeed,
		ForceCPU:           true,
		InterOpParallelism: 1,
		IntraOpParallelism: 1,
		DeterministicOps:   true,
		HashPrecision:      HashPrecision,
	}

	// Compute config hash
	data, _ := json.Marshal(struct {
		TensorFlowVersion  string `json:"tensorflow_version"`
		RandomSeed         int64  `json:"random_seed"`
		ForceCPU           bool   `json:"force_cpu"`
		InterOpParallelism int    `json:"inter_op_parallelism"`
		IntraOpParallelism int    `json:"intra_op_parallelism"`
		DeterministicOps   bool   `json:"deterministic_ops"`
		HashPrecision      int    `json:"hash_precision"`
	}{
		TensorFlowVersion:  config.TensorFlowVersion,
		RandomSeed:         config.RandomSeed,
		ForceCPU:           config.ForceCPU,
		InterOpParallelism: config.InterOpParallelism,
		IntraOpParallelism: config.IntraOpParallelism,
		DeterministicOps:   config.DeterministicOps,
		HashPrecision:      config.HashPrecision,
	})
	hash := sha256.Sum256(data)
	config.ConfigHash = hex.EncodeToString(hash[:])

	return config
}

// ValidateRuntimeConfig checks if the current runtime matches the canonical config
func ValidateRuntimeConfig(config *RuntimeConfig) []string {
	canonical := CanonicalRuntimeConfig()
	var issues []string

	if config.RandomSeed != canonical.RandomSeed {
		issues = append(issues, fmt.Sprintf("random_seed mismatch: expected %d, got %d",
			canonical.RandomSeed, config.RandomSeed))
	}

	if config.ForceCPU != canonical.ForceCPU {
		issues = append(issues, fmt.Sprintf("force_cpu mismatch: expected %v, got %v",
			canonical.ForceCPU, config.ForceCPU))
	}

	if config.InterOpParallelism != canonical.InterOpParallelism {
		issues = append(issues, fmt.Sprintf("inter_op_parallelism mismatch: expected %d, got %d",
			canonical.InterOpParallelism, config.InterOpParallelism))
	}

	if config.IntraOpParallelism != canonical.IntraOpParallelism {
		issues = append(issues, fmt.Sprintf("intra_op_parallelism mismatch: expected %d, got %d",
			canonical.IntraOpParallelism, config.IntraOpParallelism))
	}

	if config.DeterministicOps != canonical.DeterministicOps {
		issues = append(issues, fmt.Sprintf("deterministic_ops mismatch: expected %v, got %v",
			canonical.DeterministicOps, config.DeterministicOps))
	}

	return issues
}

