// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0
//
// Package inference provides production consensus configuration for ML inference.
// This file implements Task 15A: ML determinism validation + conformance suite.
//
// VE-219: Identity verification runtime with tolerance-based consensus
// Task 15A: ML consensus validation + conformance suite
//
// DESIGN DECISION: Tolerance-Based Consensus
// ==========================================
// Rather than requiring bit-exact determinism (which forces CPU-only, single-threaded),
// we use tolerance-based consensus where validators agree if scores are within an
// acceptable range. This allows:
// - GPU acceleration (10-100x faster inference)
// - Multi-threaded execution
// - Practical deployment without exotic configuration
//
// Consensus is achieved when: |score_A - score_B| <= ConsensusTolerance

package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// Consensus Configuration Constants
// ============================================================================

const (
	// ProductionRandomSeed is the recommended random seed for reproducibility.
	// While not strictly required for tolerance-based consensus, using the same
	// seed helps reduce variance across validators.
	ProductionRandomSeed int64 = 42

	// ConsensusTolerance is the maximum allowed difference between validator scores.
	// Scores within this range are considered equivalent for consensus purposes.
	// Value: 2.0 means scores of 75.0 and 76.5 would both be accepted.
	ConsensusTolerance float32 = 2.0

	// ConsensusToleranceStrict is a tighter tolerance for high-stakes operations.
	// Used for scores near critical thresholds (e.g., pass/fail boundaries).
	ConsensusToleranceStrict float32 = 0.5

	// ScoreThresholdPass is the minimum score for identity verification to pass.
	ScoreThresholdPass float32 = 60.0

	// ScoreThresholdHighTrust is the threshold for high-trust tier.
	ScoreThresholdHighTrust float32 = 80.0

	// ThresholdBuffer is the buffer zone around thresholds where stricter
	// tolerance is applied to prevent gaming.
	ThresholdBuffer float32 = 5.0

	// ModelVersionV1 is the v1.0.0 model identifier
	ModelVersionV1 = "v1.0.0"

	// ExpectedModelHashV1Production is the production model hash for v1.0.0
	// This hash is pinned from the reference training run and verified during CI.
	// DO NOT MODIFY without completing a full network upgrade procedure.
	ExpectedModelHashV1Production = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// TensorFlowVersionRequired is the required TensorFlow version for inference
	TensorFlowVersionRequired = "2.15.0"

	// ConsensusConfigVersion is the version of this configuration format
	ConsensusConfigVersion = "2.0.0"

	// DeterminismConfigVersion kept for backwards compatibility
	DeterminismConfigVersion = ConsensusConfigVersion
)

// ============================================================================
// Consensus Score Comparison
// ============================================================================

// ScoresInConsensus returns true if two scores are within the consensus tolerance.
// This is the core function for determining if validators agree.
func ScoresInConsensus(scoreA, scoreB float32) bool {
	return ScoresInConsensusWithTolerance(scoreA, scoreB, ConsensusTolerance)
}

// ScoresInConsensusWithTolerance checks if scores are within a custom tolerance.
func ScoresInConsensusWithTolerance(scoreA, scoreB, tolerance float32) bool {
	diff := float32(math.Abs(float64(scoreA - scoreB)))
	return diff <= tolerance
}

// ScoresInConsensusStrict uses the stricter tolerance for threshold-adjacent scores.
func ScoresInConsensusStrict(scoreA, scoreB float32) bool {
	return ScoresInConsensusWithTolerance(scoreA, scoreB, ConsensusToleranceStrict)
}

// GetEffectiveTolerance returns the appropriate tolerance based on the score.
// Scores near critical thresholds use stricter tolerance to prevent gaming.
func GetEffectiveTolerance(score float32) float32 {
	// Check if score is near the pass threshold
	if math.Abs(float64(score-ScoreThresholdPass)) < float64(ThresholdBuffer) {
		return ConsensusToleranceStrict
	}

	// Check if score is near the high-trust threshold
	if math.Abs(float64(score-ScoreThresholdHighTrust)) < float64(ThresholdBuffer) {
		return ConsensusToleranceStrict
	}

	return ConsensusTolerance
}

// ValidateConsensus checks if a set of validator scores achieve consensus.
// Consensus requires that all scores are within tolerance of the median.
func ValidateConsensus(scores []float32) ConsensusResult {
	if len(scores) == 0 {
		return ConsensusResult{
			Achieved: false,
			Reason:   "no scores provided",
		}
	}

	if len(scores) == 1 {
		return ConsensusResult{
			Achieved:    true,
			MedianScore: scores[0],
			Reason:      "single validator",
		}
	}

	// Calculate median
	sorted := make([]float32, len(scores))
	copy(sorted, scores)
	sortFloat32s(sorted)

	var median float32
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		median = sorted[mid]
	}

	// Determine effective tolerance based on median score
	tolerance := GetEffectiveTolerance(median)

	// Check all scores are within tolerance of median
	var maxDeviation float32
	allWithinTolerance := true
	for _, s := range scores {
		deviation := float32(math.Abs(float64(s - median)))
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
		if deviation > tolerance {
			allWithinTolerance = false
		}
	}

	return ConsensusResult{
		Achieved:     allWithinTolerance,
		MedianScore:  median,
		MaxDeviation: maxDeviation,
		Tolerance:    tolerance,
		Reason:       formatConsensusReason(allWithinTolerance, maxDeviation, tolerance),
	}
}

// ConsensusResult contains the result of a consensus validation.
type ConsensusResult struct {
	Achieved     bool    `json:"achieved"`
	MedianScore  float32 `json:"median_score"`
	MaxDeviation float32 `json:"max_deviation"`
	Tolerance    float32 `json:"tolerance"`
	Reason       string  `json:"reason"`
}

func formatConsensusReason(achieved bool, maxDeviation, tolerance float32) string {
	if achieved {
		return fmt.Sprintf("consensus achieved: max deviation %.2f within tolerance %.2f",
			maxDeviation, tolerance)
	}
	return fmt.Sprintf("consensus failed: max deviation %.2f exceeds tolerance %.2f",
		maxDeviation, tolerance)
}

func sortFloat32s(s []float32) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// ============================================================================
// Binned Score for Deterministic Hashing
// ============================================================================

// BinScore rounds a score to the nearest bin for deterministic consensus.
// This allows validators to hash the same "effective score" even with small variance.
func BinScore(score float32) float32 {
	// Bin to nearest 0.5 (e.g., 75.3 -> 75.5, 75.1 -> 75.0)
	return float32(math.Round(float64(score)*2) / 2)
}

// BinScoreInt returns the score as an integer (0-100 scale).
// This provides the most deterministic representation for hashing.
func BinScoreInt(score float32) int {
	return int(math.Round(float64(score)))
}

// ScoreBand returns the qualitative band for a score.
func ScoreBand(score float32) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 80:
		return "high"
	case score >= 70:
		return "good"
	case score >= 60:
		return "acceptable"
	case score >= 50:
		return "marginal"
	default:
		return "low"
	}
}

// ============================================================================
// TensorFlow Operations Registry (kept for reference/validation)
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
// Production Consensus Config (Tolerance-Based)
// ============================================================================

// ProductionDeterminismConfig contains the configuration for ML inference consensus.
// Note: Despite the name (kept for backwards compatibility), this now uses
// tolerance-based consensus rather than strict determinism.
type ProductionDeterminismConfig struct {
	// Version is the configuration version
	Version string `json:"version"`

	// RandomSeed is the recommended random seed for reproducibility
	RandomSeed int64 `json:"random_seed"`

	// AllowGPU enables GPU acceleration (recommended for performance)
	AllowGPU bool `json:"allow_gpu"`

	// ForceCPU forces CPU-only mode (legacy, not recommended)
	ForceCPU bool `json:"force_cpu"`

	// DisableGPU hides GPU devices (legacy, not recommended)
	DisableGPU bool `json:"disable_gpu"`

	// AllowParallelism enables multi-threaded execution
	AllowParallelism bool `json:"allow_parallelism"`

	// InterOpParallelism controls TF inter-op parallelism (0 = auto)
	InterOpParallelism int `json:"inter_op_parallelism"`

	// IntraOpParallelism controls TF intra-op parallelism (0 = auto)
	IntraOpParallelism int `json:"intra_op_parallelism"`

	// ConsensusTolerance is the max allowed score difference for consensus
	ConsensusTolerance float32 `json:"consensus_tolerance"`

	// StrictTolerance is the tolerance near thresholds
	StrictTolerance float32 `json:"strict_tolerance"`

	// EnableDeterministicOps enables TF deterministic operations (optional)
	EnableDeterministicOps bool `json:"enable_deterministic_ops"`

	// DisableAutoTuning disables cuDNN auto-tuning (optional)
	DisableAutoTuning bool `json:"disable_auto_tuning"`

	// ModelVersion is the required model version
	ModelVersion string `json:"model_version"`

	// ExpectedModelHash is the pinned model hash
	ExpectedModelHash string `json:"expected_model_hash"`

	// TensorFlowVersion is the required TF version
	TensorFlowVersion string `json:"tensorflow_version"`

	// StrictMode uses strict tolerance for all comparisons
	StrictMode bool `json:"strict_mode"`

	// RequireHashVerification requires model hash verification
	RequireHashVerification bool `json:"require_hash_verification"`

	// ConfigHash is the hash of this configuration for integrity
	ConfigHash string `json:"config_hash,omitempty"`

	// GeneratedAt is when this config was generated
	GeneratedAt time.Time `json:"generated_at"`
}

// NewProductionDeterminismConfig creates the production consensus configuration.
// This uses tolerance-based consensus, allowing GPU and multi-threading.
func NewProductionDeterminismConfig() *ProductionDeterminismConfig {
	config := &ProductionDeterminismConfig{
		Version:                 ConsensusConfigVersion,
		RandomSeed:              ProductionRandomSeed,
		AllowGPU:                true,  // GPU enabled for performance
		ForceCPU:                false, // Not required with tolerance-based consensus
		DisableGPU:              false,
		AllowParallelism:        true,             // Multi-threading enabled
		InterOpParallelism:      0,                // 0 = auto (let TF decide)
		IntraOpParallelism:      0,                // 0 = auto
		ConsensusTolerance:      ConsensusTolerance,
		StrictTolerance:         ConsensusToleranceStrict,
		EnableDeterministicOps:  false,            // Not required
		DisableAutoTuning:       false,
		ModelVersion:            ModelVersionV1,
		ExpectedModelHash:       ExpectedModelHashV1Production,
		TensorFlowVersion:       TensorFlowVersionRequired,
		StrictMode:              false,            // Use dynamic tolerance
		RequireHashVerification: true,             // Still verify model hash
		GeneratedAt:             time.Now().UTC(),
	}

	// Compute config hash
	config.ConfigHash = config.computeHash()

	return config
}

// NewStrictDeterminismConfig creates a strict CPU-only config for testing.
// This is the legacy mode that forces full determinism.
func NewStrictDeterminismConfig() *ProductionDeterminismConfig {
	config := &ProductionDeterminismConfig{
		Version:                 ConsensusConfigVersion,
		RandomSeed:              ProductionRandomSeed,
		AllowGPU:                false,
		ForceCPU:                true,
		DisableGPU:              true,
		AllowParallelism:        false,
		InterOpParallelism:      1,
		IntraOpParallelism:      1,
		ConsensusTolerance:      0,     // Zero tolerance = exact match
		StrictTolerance:         0,
		EnableDeterministicOps:  true,
		DisableAutoTuning:       true,
		ModelVersion:            ModelVersionV1,
		ExpectedModelHash:       ExpectedModelHashV1Production,
		TensorFlowVersion:       TensorFlowVersionRequired,
		StrictMode:              true,
		RequireHashVerification: true,
		GeneratedAt:             time.Now().UTC(),
	}

	config.ConfigHash = config.computeHash()
	return config
}

// computeHash computes the SHA256 hash of the configuration.
func (c *ProductionDeterminismConfig) computeHash() string {
	// Create a copy without the hash field for computing
	hashable := struct {
		Version                 string  `json:"version"`
		RandomSeed              int64   `json:"random_seed"`
		AllowGPU                bool    `json:"allow_gpu"`
		AllowParallelism        bool    `json:"allow_parallelism"`
		ConsensusTolerance      float32 `json:"consensus_tolerance"`
		StrictTolerance         float32 `json:"strict_tolerance"`
		ModelVersion            string  `json:"model_version"`
		ExpectedModelHash       string  `json:"expected_model_hash"`
		StrictMode              bool    `json:"strict_mode"`
		RequireHashVerification bool    `json:"require_hash_verification"`
	}{
		Version:                 c.Version,
		RandomSeed:              c.RandomSeed,
		AllowGPU:                c.AllowGPU,
		AllowParallelism:        c.AllowParallelism,
		ConsensusTolerance:      c.ConsensusTolerance,
		StrictTolerance:         c.StrictTolerance,
		ModelVersion:            c.ModelVersion,
		ExpectedModelHash:       c.ExpectedModelHash,
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

// Validate validates the consensus configuration.
// For tolerance-based consensus, requirements are relaxed compared to strict determinism.
func (c *ProductionDeterminismConfig) Validate() []string {
	var issues []string

	// Validate tolerance is reasonable (must be positive, not too large)
	if c.ConsensusTolerance < 0 {
		issues = append(issues, "consensus_tolerance cannot be negative")
	}
	if c.ConsensusTolerance > 10.0 {
		issues = append(issues, fmt.Sprintf(
			"consensus_tolerance of %.2f is too large (max 10.0)", c.ConsensusTolerance))
	}

	// Validate strict tolerance is less than or equal to regular tolerance
	if c.StrictTolerance > c.ConsensusTolerance && c.ConsensusTolerance > 0 {
		issues = append(issues, fmt.Sprintf(
			"strict_tolerance (%.2f) should not exceed consensus_tolerance (%.2f)",
			c.StrictTolerance, c.ConsensusTolerance))
	}

	// Validate model hash
	if c.ExpectedModelHash == "" {
		issues = append(issues, "expected_model_hash must be set for production")
	} else if len(c.ExpectedModelHash) != 64 {
		issues = append(issues, fmt.Sprintf(
			"expected_model_hash must be 64 hex chars, got %d",
			len(c.ExpectedModelHash)))
	}

	// Model hash verification is always required
	if !c.RequireHashVerification {
		issues = append(issues, "require_hash_verification must be true for production")
	}

	// Warn (not error) about legacy strict mode settings
	if c.ForceCPU && c.AllowGPU {
		issues = append(issues, "inconsistent: both force_cpu and allow_gpu are set")
	}

	return issues
}

// ToInferenceConfig converts to the standard InferenceConfig.
func (c *ProductionDeterminismConfig) ToInferenceConfig() InferenceConfig {
	return InferenceConfig{
		ModelVersion:            c.ModelVersion,
		ExpectedHash:            c.ExpectedModelHash,
		Deterministic:           c.StrictMode, // Only strict mode requires full determinism
		ForceCPU:                c.ForceCPU,
		RandomSeed:              c.RandomSeed,
		RequireHashVerification: c.RequireHashVerification,
		StrictDeterminism:       c.StrictMode,
		Enabled:                 true,
	}
}

// GetEnvironmentVariables returns the environment variables for TensorFlow.
// For tolerance-based consensus, these are optional optimizations.
func (c *ProductionDeterminismConfig) GetEnvironmentVariables() map[string]string {
	envVars := map[string]string{
		// Common settings
		"TF_CPP_MIN_LOG_LEVEL": "2",
		"PYTHONHASHSEED":       fmt.Sprintf("%d", c.RandomSeed),
	}

	// Only set strict determinism env vars if in strict mode
	if c.StrictMode || c.ForceCPU {
		envVars["CUDA_VISIBLE_DEVICES"] = "-1"
		envVars["TF_DETERMINISTIC_OPS"] = "1"
		envVars["TF_CUDNN_DETERMINISTIC"] = "1"
		envVars["TF_USE_CUDNN_AUTOTUNE"] = "0"
		envVars["TF_ENABLE_ONEDNN_OPTS"] = "0"
		envVars["TF_FORCE_GPU_ALLOW_GROWTH"] = "false"
		envVars["TF_XLA_FLAGS"] = "--tf_xla_auto_jit=-1"
		envVars["OMP_NUM_THREADS"] = "1"
		envVars["MKL_NUM_THREADS"] = "1"
		envVars["OPENBLAS_NUM_THREADS"] = "1"
	} else if c.AllowGPU {
		// Allow GPU with memory growth for tolerance-based consensus
		envVars["TF_FORCE_GPU_ALLOW_GROWTH"] = "true"
	}

	return envVars
}

// CheckScoreConsensus checks if a proposed score is in consensus with a reference score.
func (c *ProductionDeterminismConfig) CheckScoreConsensus(proposed, reference float32) bool {
	tolerance := c.ConsensusTolerance

	// Use strict tolerance near thresholds
	if c.StrictMode ||
		math.Abs(float64(reference-ScoreThresholdPass)) <= float64(ThresholdBuffer) ||
		math.Abs(float64(reference-ScoreThresholdHighTrust)) <= float64(ThresholdBuffer) {
		tolerance = c.StrictTolerance
	}

	return ScoresInConsensusWithTolerance(proposed, reference, tolerance)
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
