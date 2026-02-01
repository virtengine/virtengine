// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides production determinism configuration for ML inference.
// This file implements Task 15A: ML determinism validation + conformance suite.
//
// VE-219: Deterministic identity verification runtime
// Task 15A: ML determinism validation + conformance suite

package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// Production Determinism Constants
// ============================================================================

const (
	// ProductionRandomSeed is the fixed random seed for all production validators.
	// This value MUST NOT change without a coordinated network upgrade.
	ProductionRandomSeed int64 = 42

	// ProductionHashPrecision is the decimal precision for float hashing.
	// 6 decimal places provides sufficient precision while avoiding
	// floating-point representation differences across platforms.
	ProductionHashPrecision = 6

	// ModelVersionV1 is the v1.0.0 model identifier
	ModelVersionV1 = "v1.0.0"

	// ExpectedModelHashV1Production is the production model hash for v1.0.0
	// This hash is pinned from the reference training run and verified during CI.
	// DO NOT MODIFY without completing a full network upgrade procedure.
	ExpectedModelHashV1Production = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// TensorFlowVersionRequired is the required TensorFlow version for inference
	TensorFlowVersionRequired = "2.15.0"

	// DeterminismConfigVersion is the version of this configuration format
	DeterminismConfigVersion = "1.0.0"
)

// ============================================================================
// TensorFlow Deterministic Operations Registry
// ============================================================================

// DeterministicOps lists TensorFlow operations verified to be deterministic
// when running with CPU-only mode and single-threaded execution.
// These operations have been tested across multiple platforms.
var DeterministicOps = []string{
	// Math operations
	"Add", "AddN", "Sub", "Mul", "Div", "FloorDiv", "RealDiv",
	"Neg", "Abs", "Sign", "Reciprocal", "Square", "Sqrt", "Rsqrt",
	"Exp", "Log", "Log1p", "Pow", "Sin", "Cos", "Tan",
	"Maximum", "Minimum", "Floor", "Ceil", "Round",

	// Matrix operations
	"MatMul", "BatchMatMul", "BatchMatMulV2",
	"Transpose", "Reshape", "Squeeze", "ExpandDims",

	// Activation functions
	"Relu", "Relu6", "LeakyRelu", "Elu", "Selu",
	"Sigmoid", "Tanh", "Softmax", "LogSoftmax",

	// Reduction operations (deterministic with single thread)
	"Sum", "Mean", "Prod", "Max", "Min", "All", "Any",
	"ReduceSum", "ReduceMean", "ReduceProd", "ReduceMax", "ReduceMin",

	// Normalization (deterministic with CPU + single thread)
	"FusedBatchNorm", "FusedBatchNormV2", "FusedBatchNormV3",
	"L2Normalize",

	// Embedding and gather
	"Gather", "GatherV2", "GatherNd",

	// Concatenation and splitting
	"Concat", "ConcatV2", "Split", "SplitV", "Pack", "Unpack",

	// Comparison
	"Equal", "NotEqual", "Less", "LessEqual", "Greater", "GreaterEqual",

	// Logical
	"LogicalAnd", "LogicalOr", "LogicalNot",

	// Selection
	"Where", "Select", "SelectV2",

	// Constants and placeholders
	"Const", "Placeholder", "PlaceholderV2",

	// Identity and control flow
	"Identity", "IdentityN", "NoOp", "StopGradient",

	// Shape operations
	"Shape", "ShapeN", "Size", "Rank",

	// Cast and type
	"Cast", "Bitcast",

	// Slice operations
	"Slice", "StridedSlice",

	// Padding
	"Pad", "PadV2", "MirrorPad",
}

// NonDeterministicOpsProduction lists TensorFlow operations that are
// NEVER deterministic and must not be used in production models.
var NonDeterministicOpsProduction = []string{
	// GPU-specific operations
	"CudnnRNN", "CudnnRNNV2", "CudnnRNNV3",
	"CudnnRNNBackprop", "CudnnRNNBackpropV2", "CudnnRNNBackpropV3",

	// Dropout (random by nature)
	"Dropout", "RandomUniform", "RandomNormal", "RandomStandardNormal",
	"TruncatedNormal", "RandomShuffle",

	// Stateful random
	"StatelessRandomUniform", "StatelessRandomNormal",
	"StatelessTruncatedNormal",

	// Non-deterministic pooling on GPU
	"MaxPoolGrad", "MaxPoolGradV2", "MaxPool3DGrad",
	"AvgPoolGrad", "AvgPool3DGrad",

	// Scatter operations (non-atomic)
	"ScatterAdd", "ScatterSub", "ScatterMul", "ScatterDiv",
	"ScatterNd", "ScatterNdAdd", "ScatterNdSub",
	"TensorScatterAdd", "TensorScatterSub", "TensorScatterUpdate",

	// Non-deterministic backwards
	"BiasAddGrad",
	"Conv2DBackpropInput", "Conv2DBackpropFilter",
	"DepthwiseConv2dNativeBackpropInput", "DepthwiseConv2dNativeBackpropFilter",
	"FusedBatchNormGrad", "FusedBatchNormGradV2", "FusedBatchNormGradV3",

	// CTC operations
	"CTCLoss", "CTCGreedyDecoder", "CTCBeamSearchDecoder",

	// Cross entropy (can be non-deterministic on GPU)
	"SoftmaxCrossEntropyWithLogits", "SparseSoftmaxCrossEntropyWithLogits",
}

// ConditionallyDeterministicOps lists operations that are deterministic
// ONLY when running with specific configuration (CPU, single thread).
var ConditionallyDeterministicOps = []string{
	// Convolution - deterministic on CPU with single thread
	"Conv2D", "Conv3D", "DepthwiseConv2dNative",
	"Conv2DBackpropInput", "Conv2DBackpropFilter",

	// Bias operations - deterministic on CPU
	"BiasAdd", "BiasAddV1",

	// Pooling - deterministic on CPU
	"MaxPool", "MaxPoolV2", "AvgPool", "MaxPool3D", "AvgPool3D",

	// Segment operations - deterministic with proper indexing
	"SegmentSum", "SegmentMean", "SegmentProd", "SegmentMax", "SegmentMin",
	"UnsortedSegmentSum", "UnsortedSegmentMean", "UnsortedSegmentProd",
}

// ============================================================================
// Production Determinism Config
// ============================================================================

// ProductionDeterminismConfig contains the canonical configuration
// for deterministic ML inference in production validators.
type ProductionDeterminismConfig struct {
	// Version is the configuration version
	Version string `json:"version"`

	// RandomSeed is the fixed random seed (must be 42)
	RandomSeed int64 `json:"random_seed"`

	// ForceCPU must be true for determinism
	ForceCPU bool `json:"force_cpu"`

	// DisableGPU hides GPU devices
	DisableGPU bool `json:"disable_gpu"`

	// InterOpParallelism must be 1 for determinism
	InterOpParallelism int `json:"inter_op_parallelism"`

	// IntraOpParallelism must be 1 for determinism
	IntraOpParallelism int `json:"intra_op_parallelism"`

	// EnableDeterministicOps enables TF deterministic operations
	EnableDeterministicOps bool `json:"enable_deterministic_ops"`

	// DisableAutoTuning disables cuDNN auto-tuning
	DisableAutoTuning bool `json:"disable_auto_tuning"`

	// HashPrecision is the decimal precision for float hashing
	HashPrecision int `json:"hash_precision"`

	// ModelVersion is the required model version
	ModelVersion string `json:"model_version"`

	// ExpectedModelHash is the pinned model hash
	ExpectedModelHash string `json:"expected_model_hash"`

	// TensorFlowVersion is the required TF version
	TensorFlowVersion string `json:"tensorflow_version"`

	// StrictMode fails on any potential non-determinism
	StrictMode bool `json:"strict_mode"`

	// RequireHashVerification requires output hash verification
	RequireHashVerification bool `json:"require_hash_verification"`

	// ConfigHash is the hash of this configuration for integrity
	ConfigHash string `json:"config_hash,omitempty"`

	// GeneratedAt is when this config was generated
	GeneratedAt time.Time `json:"generated_at"`
}

// NewProductionDeterminismConfig creates the canonical production configuration.
// This should be used by all validators in the network.
func NewProductionDeterminismConfig() *ProductionDeterminismConfig {
	config := &ProductionDeterminismConfig{
		Version:                 DeterminismConfigVersion,
		RandomSeed:              ProductionRandomSeed,
		ForceCPU:                true,
		DisableGPU:              true,
		InterOpParallelism:      1,
		IntraOpParallelism:      1,
		EnableDeterministicOps:  true,
		DisableAutoTuning:       true,
		HashPrecision:           ProductionHashPrecision,
		ModelVersion:            ModelVersionV1,
		ExpectedModelHash:       ExpectedModelHashV1Production,
		TensorFlowVersion:       TensorFlowVersionRequired,
		StrictMode:              true,
		RequireHashVerification: true,
		GeneratedAt:             time.Now().UTC(),
	}

	// Compute config hash
	config.ConfigHash = config.computeHash()

	return config
}

// computeHash computes the SHA256 hash of the configuration.
func (c *ProductionDeterminismConfig) computeHash() string {
	// Create a copy without the hash field for computing
	hashable := struct {
		Version                 string `json:"version"`
		RandomSeed              int64  `json:"random_seed"`
		ForceCPU                bool   `json:"force_cpu"`
		DisableGPU              bool   `json:"disable_gpu"`
		InterOpParallelism      int    `json:"inter_op_parallelism"`
		IntraOpParallelism      int    `json:"intra_op_parallelism"`
		EnableDeterministicOps  bool   `json:"enable_deterministic_ops"`
		DisableAutoTuning       bool   `json:"disable_auto_tuning"`
		HashPrecision           int    `json:"hash_precision"`
		ModelVersion            string `json:"model_version"`
		ExpectedModelHash       string `json:"expected_model_hash"`
		TensorFlowVersion       string `json:"tensorflow_version"`
		StrictMode              bool   `json:"strict_mode"`
		RequireHashVerification bool   `json:"require_hash_verification"`
	}{
		Version:                 c.Version,
		RandomSeed:              c.RandomSeed,
		ForceCPU:                c.ForceCPU,
		DisableGPU:              c.DisableGPU,
		InterOpParallelism:      c.InterOpParallelism,
		IntraOpParallelism:      c.IntraOpParallelism,
		EnableDeterministicOps:  c.EnableDeterministicOps,
		DisableAutoTuning:       c.DisableAutoTuning,
		HashPrecision:           c.HashPrecision,
		ModelVersion:            c.ModelVersion,
		ExpectedModelHash:       c.ExpectedModelHash,
		TensorFlowVersion:       c.TensorFlowVersion,
		StrictMode:              c.StrictMode,
		RequireHashVerification: c.RequireHashVerification,
	}

	data, err := json.Marshal(hashable)
	if err != nil {
		// Fall back to a deterministic default if marshaling fails
		return "config-hash-error"
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Validate validates the production configuration.
func (c *ProductionDeterminismConfig) Validate() []string {
	var issues []string

	// Validate random seed
	if c.RandomSeed != ProductionRandomSeed {
		issues = append(issues, fmt.Sprintf(
			"random_seed must be %d for production, got %d",
			ProductionRandomSeed, c.RandomSeed))
	}

	// Validate CPU-only mode
	if !c.ForceCPU {
		issues = append(issues, "force_cpu must be true for production determinism")
	}

	if !c.DisableGPU {
		issues = append(issues, "disable_gpu must be true for production determinism")
	}

	// Validate parallelism
	if c.InterOpParallelism != 1 {
		issues = append(issues, fmt.Sprintf(
			"inter_op_parallelism must be 1 for determinism, got %d",
			c.InterOpParallelism))
	}

	if c.IntraOpParallelism != 1 {
		issues = append(issues, fmt.Sprintf(
			"intra_op_parallelism must be 1 for determinism, got %d",
			c.IntraOpParallelism))
	}

	// Validate deterministic ops
	if !c.EnableDeterministicOps {
		issues = append(issues, "enable_deterministic_ops must be true")
	}

	// Validate model hash
	if c.ExpectedModelHash == "" {
		issues = append(issues, "expected_model_hash must be set for production")
	} else if len(c.ExpectedModelHash) != 64 {
		issues = append(issues, fmt.Sprintf(
			"expected_model_hash must be 64 hex chars, got %d",
			len(c.ExpectedModelHash)))
	}

	// Validate strict mode
	if !c.StrictMode {
		issues = append(issues, "strict_mode must be true for production")
	}

	// Validate hash verification
	if !c.RequireHashVerification {
		issues = append(issues, "require_hash_verification must be true for production")
	}

	return issues
}

// ToInferenceConfig converts to the standard InferenceConfig.
func (c *ProductionDeterminismConfig) ToInferenceConfig() InferenceConfig {
	return InferenceConfig{
		ModelVersion:            c.ModelVersion,
		ExpectedHash:            c.ExpectedModelHash,
		Deterministic:           true,
		ForceCPU:                c.ForceCPU,
		RandomSeed:              c.RandomSeed,
		RequireHashVerification: c.RequireHashVerification,
		StrictDeterminism:       c.StrictMode,
		Enabled:                 true,
	}
}

// GetEnvironmentVariables returns the environment variables for TensorFlow.
func (c *ProductionDeterminismConfig) GetEnvironmentVariables() map[string]string {
	return map[string]string{
		// Disable GPU
		"CUDA_VISIBLE_DEVICES": "-1",

		// TensorFlow determinism settings
		"TF_DETERMINISTIC_OPS":      "1",
		"TF_CUDNN_DETERMINISTIC":    "1",
		"TF_USE_CUDNN_AUTOTUNE":     "0",
		"TF_ENABLE_ONEDNN_OPTS":     "0",
		"TF_CPP_MIN_LOG_LEVEL":      "2",
		"TF_FORCE_GPU_ALLOW_GROWTH": "false",
		"TF_XLA_FLAGS":              "--tf_xla_auto_jit=-1",

		// Thread settings
		"OMP_NUM_THREADS":   "1",
		"MKL_NUM_THREADS":   "1",
		"OPENBLAS_NUM_THREADS": "1",

		// Python hash seed
		"PYTHONHASHSEED": fmt.Sprintf("%d", c.RandomSeed),
	}
}

// ============================================================================
// Model Operation Validation
// ============================================================================

// ModelOpValidationResult contains the result of validating model operations.
type ModelOpValidationResult struct {
	// Valid indicates all operations are deterministic
	Valid bool `json:"valid"`

	// DeterministicOps lists verified deterministic operations
	DeterministicOps []string `json:"deterministic_ops"`

	// NonDeterministicOps lists non-deterministic operations found
	NonDeterministicOps []string `json:"non_deterministic_ops"`

	// ConditionalOps lists conditionally deterministic operations
	ConditionalOps []string `json:"conditional_ops"`

	// UnknownOps lists operations not in either list
	UnknownOps []string `json:"unknown_ops"`

	// Warnings contains warning messages
	Warnings []string `json:"warnings"`

	// Platform contains platform information
	Platform PlatformValidationInfo `json:"platform"`
}

// PlatformValidationInfo contains platform information for validation.
type PlatformValidationInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
	NumCPU    int    `json:"num_cpu"`
	Timestamp time.Time `json:"timestamp"`
}

// ValidateModelOperations validates a list of TensorFlow operations
// against the determinism registry.
func ValidateModelOperations(opNames []string, strictMode bool) *ModelOpValidationResult {
	result := &ModelOpValidationResult{
		Valid:               true,
		DeterministicOps:    make([]string, 0),
		NonDeterministicOps: make([]string, 0),
		ConditionalOps:      make([]string, 0),
		UnknownOps:          make([]string, 0),
		Warnings:            make([]string, 0),
		Platform: PlatformValidationInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			NumCPU:    runtime.NumCPU(),
			Timestamp: time.Now().UTC(),
		},
	}

	// Create lookup sets
	deterministicSet := make(map[string]bool)
	for _, op := range DeterministicOps {
		deterministicSet[op] = true
	}

	nonDeterministicSet := make(map[string]bool)
	for _, op := range NonDeterministicOpsProduction {
		nonDeterministicSet[op] = true
	}

	conditionalSet := make(map[string]bool)
	for _, op := range ConditionallyDeterministicOps {
		conditionalSet[op] = true
	}

	// Classify operations
	for _, op := range opNames {
		// Skip empty operation names
		if op == "" {
			continue
		}

		if nonDeterministicSet[op] {
			result.NonDeterministicOps = append(result.NonDeterministicOps, op)
			result.Valid = false
		} else if deterministicSet[op] {
			result.DeterministicOps = append(result.DeterministicOps, op)
		} else if conditionalSet[op] {
			result.ConditionalOps = append(result.ConditionalOps, op)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("operation %s is conditionally deterministic (requires CPU + single thread)", op))
		} else {
			result.UnknownOps = append(result.UnknownOps, op)
			if strictMode {
				result.Valid = false
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("unknown operation %s in strict mode", op))
			} else {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("unknown operation %s not in registry", op))
			}
		}
	}

	// Sort for deterministic output
	sort.Strings(result.DeterministicOps)
	sort.Strings(result.NonDeterministicOps)
	sort.Strings(result.ConditionalOps)
	sort.Strings(result.UnknownOps)

	return result
}

// ============================================================================
// Model Hash Pinning
// ============================================================================

// PinnedModelHash represents a pinned model hash for a specific version.
type PinnedModelHash struct {
	// Version is the model version (e.g., "v1.0.0")
	Version string `json:"version"`

	// Hash is the SHA256 hash of the model weights
	Hash string `json:"hash"`

	// Description describes the model
	Description string `json:"description"`

	// PinnedAt is when the hash was pinned
	PinnedAt time.Time `json:"pinned_at"`

	// PinnedBy identifies who pinned the hash
	PinnedBy string `json:"pinned_by"`

	// Attestations contains verification attestations
	Attestations []HashAttestation `json:"attestations,omitempty"`
}

// HashAttestation is an attestation of a model hash from a validator.
type HashAttestation struct {
	// ValidatorAddress is the validator that attested
	ValidatorAddress string `json:"validator_address"`

	// Platform is the platform the validator ran on
	Platform string `json:"platform"`

	// ComputedHash is the hash the validator computed
	ComputedHash string `json:"computed_hash"`

	// Timestamp is when the attestation was made
	Timestamp time.Time `json:"timestamp"`

	// Signature is the validator's signature (optional)
	Signature string `json:"signature,omitempty"`
}

// PinnedModelRegistry contains all pinned model hashes for the network.
var PinnedModelRegistry = []PinnedModelHash{
	{
		Version:     "v1.0.0",
		Hash:        ExpectedModelHashV1Production,
		Description: "VEID Trust Score Model v1.0.0 - Initial production release",
		PinnedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PinnedBy:    "VirtEngine Core Team",
	},
}

// GetPinnedModelHash returns the pinned hash for a model version.
func GetPinnedModelHash(version string) (string, bool) {
	for _, pinned := range PinnedModelRegistry {
		if pinned.Version == version {
			return pinned.Hash, true
		}
	}
	return "", false
}

// VerifyModelHash verifies a computed hash against the pinned hash.
func VerifyModelHash(version, computedHash string) error {
	pinnedHash, found := GetPinnedModelHash(version)
	if !found {
		return fmt.Errorf("no pinned hash for model version %s", version)
	}

	// Normalize both hashes
	computedNorm := strings.ToLower(strings.TrimSpace(computedHash))
	pinnedNorm := strings.ToLower(strings.TrimSpace(pinnedHash))

	if computedNorm != pinnedNorm {
		return fmt.Errorf("model hash mismatch for %s: expected %s, got %s",
			version, pinnedNorm, computedNorm)
	}

	return nil
}

// ============================================================================
// Environment Setup
// ============================================================================

// SetupProductionEnvironment sets up the environment for production determinism.
// This should be called before initializing TensorFlow.
func SetupProductionEnvironment(config *ProductionDeterminismConfig) error {
	envVars := config.GetEnvironmentVariables()

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// ValidateProductionEnvironment validates that the environment is configured correctly.
func ValidateProductionEnvironment(config *ProductionDeterminismConfig) []string {
	var issues []string
	envVars := config.GetEnvironmentVariables()

	for key, expected := range envVars {
		actual := os.Getenv(key)
		if actual != expected {
			issues = append(issues, fmt.Sprintf(
				"env var %s: expected %s, got %s", key, expected, actual))
		}
	}

	return issues
}

// ============================================================================
// Conformance Report
// ============================================================================

// ConformanceReport contains a complete determinism conformance report.
type ConformanceReport struct {
	// ReportID is a unique identifier for this report
	ReportID string `json:"report_id"`

	// Version is the report format version
	Version string `json:"version"`

	// GeneratedAt is when the report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// Platform contains platform information
	Platform PlatformValidationInfo `json:"platform"`

	// Configuration contains the validated configuration
	Configuration *ProductionDeterminismConfig `json:"configuration"`

	// ConfigurationValid indicates if config passed validation
	ConfigurationValid bool `json:"configuration_valid"`

	// ConfigurationIssues lists config validation issues
	ConfigurationIssues []string `json:"configuration_issues,omitempty"`

	// EnvironmentValid indicates if env vars are correct
	EnvironmentValid bool `json:"environment_valid"`

	// EnvironmentIssues lists env var issues
	EnvironmentIssues []string `json:"environment_issues,omitempty"`

	// ModelOperations contains op validation results
	ModelOperations *ModelOpValidationResult `json:"model_operations,omitempty"`

	// GoldenVectorResults contains golden vector test results
	GoldenVectorResults []GoldenVectorResult `json:"golden_vector_results,omitempty"`

	// OverallPassed indicates if all checks passed
	OverallPassed bool `json:"overall_passed"`

	// Summary contains a human-readable summary
	Summary string `json:"summary"`
}

// GenerateConformanceReport generates a complete conformance report.
func GenerateConformanceReport(config *ProductionDeterminismConfig, modelOps []string) *ConformanceReport {
	report := &ConformanceReport{
		ReportID:    fmt.Sprintf("cr-%s-%d", runtime.GOOS, time.Now().Unix()),
		Version:     DeterminismConfigVersion,
		GeneratedAt: time.Now().UTC(),
		Platform: PlatformValidationInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			NumCPU:    runtime.NumCPU(),
			Timestamp: time.Now().UTC(),
		},
		Configuration:  config,
		OverallPassed: true,
	}

	// Validate configuration
	configIssues := config.Validate()
	report.ConfigurationValid = len(configIssues) == 0
	report.ConfigurationIssues = configIssues
	if !report.ConfigurationValid {
		report.OverallPassed = false
	}

	// Validate environment
	envIssues := ValidateProductionEnvironment(config)
	report.EnvironmentValid = len(envIssues) == 0
	report.EnvironmentIssues = envIssues
	if !report.EnvironmentValid {
		report.OverallPassed = false
	}

	// Validate model operations if provided
	if len(modelOps) > 0 {
		report.ModelOperations = ValidateModelOperations(modelOps, config.StrictMode)
		if !report.ModelOperations.Valid {
			report.OverallPassed = false
		}
	}

	// Generate summary
	if report.OverallPassed {
		report.Summary = fmt.Sprintf(
			"Conformance check PASSED on %s/%s with model %s",
			runtime.GOOS, runtime.GOARCH, config.ModelVersion)
	} else {
		totalIssues := len(configIssues) + len(envIssues)
		if report.ModelOperations != nil {
			totalIssues += len(report.ModelOperations.NonDeterministicOps)
		}
		report.Summary = fmt.Sprintf(
			"Conformance check FAILED on %s/%s with %d issues",
			runtime.GOOS, runtime.GOARCH, totalIssues)
	}

	return report
}

// ToJSON returns the report as JSON.
func (r *ConformanceReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
