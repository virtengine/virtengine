package inference

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// Inference Configuration
// ============================================================================

// InferenceConfig holds all configuration for ML inference
type InferenceConfig struct {
	// Model Configuration
	// ModelPath is the path to the TensorFlow SavedModel directory
	ModelPath string

	// ModelVersion is the expected model version string
	ModelVersion string

	// ExpectedHash is the expected SHA256 hash of the model weights
	// If set, model loading will fail if the hash doesn't match
	ExpectedHash string

	// Resource Limits
	// Timeout is the maximum time allowed for a single inference call
	Timeout time.Duration

	// MaxMemoryMB is the maximum memory allowed for inference in megabytes
	MaxMemoryMB int

	// MaxBatchSize is the maximum number of inputs to process in one call
	MaxBatchSize int

	// Sidecar Configuration
	// UseSidecar enables the gRPC sidecar client instead of embedded TensorFlow
	UseSidecar bool

	// SidecarAddress is the gRPC address of the inference sidecar
	SidecarAddress string

	// SidecarTimeout is the timeout for sidecar RPC calls
	SidecarTimeout time.Duration

	// Determinism Configuration
	// Deterministic forces deterministic inference mode
	Deterministic bool

	// ForceCPU forces CPU-only execution (no GPU)
	ForceCPU bool

	// RandomSeed is the random seed for deterministic execution
	RandomSeed int64

	// Feature Configuration
	// ExpectedInputDim is the expected input feature dimension
	ExpectedInputDim int

	// Fallback Configuration
	// UseFallbackOnError returns fallback score on inference errors
	UseFallbackOnError bool

	// FallbackScore is the score returned when fallback is triggered
	FallbackScore uint32

	// Logging Configuration
	// LogInferenceDetails enables detailed inference logging
	LogInferenceDetails bool
}

// DefaultInferenceConfig returns the default inference configuration
func DefaultInferenceConfig() InferenceConfig {
	return InferenceConfig{
		// Model defaults
		ModelPath:    "models/trust_score",
		ModelVersion: "v1.0.0",
		ExpectedHash: "", // No hash verification by default

		// Resource limits
		Timeout:      2 * time.Second,
		MaxMemoryMB:  512,
		MaxBatchSize: 1,

		// Sidecar defaults
		UseSidecar:     false,
		SidecarAddress: "localhost:50051",
		SidecarTimeout: 5 * time.Second,

		// Determinism defaults
		Deterministic: true,
		ForceCPU:      true,
		RandomSeed:    42,

		// Feature defaults
		ExpectedInputDim: TotalFeatureDim,

		// Fallback defaults
		UseFallbackOnError: true,
		FallbackScore:      0,

		// Logging defaults
		LogInferenceDetails: false,
	}
}

// Validate validates the inference configuration
func (c *InferenceConfig) Validate() error {
	// Normalize expected hash if provided
	if c.ExpectedHash != "" {
		c.ExpectedHash = normalizeExpectedHash(c.ExpectedHash)
		if !isValidSHA256Hex(c.ExpectedHash) {
			return fmt.Errorf("expected_hash must be 64 hex chars, got %s", c.ExpectedHash)
		}
	}

	if c.UseSidecar {
		if c.SidecarAddress == "" {
			return fmt.Errorf("sidecar_address is required when use_sidecar is true")
		}
		if c.SidecarTimeout <= 0 {
			return fmt.Errorf("sidecar_timeout must be positive")
		}
	} else if c.ModelPath == "" {
		return fmt.Errorf("model_path is required when not using sidecar")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	if c.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be positive")
	}

	if c.ExpectedInputDim <= 0 {
		return fmt.Errorf("expected_input_dim must be positive")
	}

	if c.ExpectedInputDim != TotalFeatureDim {
		return fmt.Errorf("expected_input_dim must be %d, got %d", TotalFeatureDim, c.ExpectedInputDim)
	}

	if c.FallbackScore > 100 {
		return fmt.Errorf("fallback_score must be 0-100, got %d", c.FallbackScore)
	}

	if c.Deterministic {
		if !c.ForceCPU {
			return fmt.Errorf("force_cpu must be true when deterministic mode is enabled")
		}
		if c.ExpectedHash == "" {
			return fmt.Errorf("expected_hash must be set when deterministic mode is enabled")
		}
	}

	return nil
}

// WithModelPath returns a copy of the config with the model path set
func (c InferenceConfig) WithModelPath(path string) InferenceConfig {
	c.ModelPath = path
	return c
}

// WithTimeout returns a copy of the config with the timeout set
func (c InferenceConfig) WithTimeout(timeout time.Duration) InferenceConfig {
	c.Timeout = timeout
	return c
}

// WithSidecar returns a copy of the config configured for sidecar mode
func (c InferenceConfig) WithSidecar(address string) InferenceConfig {
	c.UseSidecar = true
	c.SidecarAddress = address
	return c
}

// WithDeterministic returns a copy of the config with deterministic mode set
func (c InferenceConfig) WithDeterministic(deterministic bool) InferenceConfig {
	c.Deterministic = deterministic
	return c
}

// ============================================================================
// Environment Variables
// ============================================================================

// Environment variable names for configuration
const (
	// EnvInferenceModelPath is the environment variable for model path
	EnvInferenceModelPath = "VEID_INFERENCE_MODEL_PATH"

	// EnvInferenceModelVersion is the environment variable for model version
	EnvInferenceModelVersion = "VEID_INFERENCE_MODEL_VERSION"

	// EnvInferenceModelHash is the environment variable for expected model hash
	EnvInferenceModelHash = "VEID_INFERENCE_MODEL_HASH"

	// EnvInferenceTimeout is the environment variable for timeout
	EnvInferenceTimeout = "VEID_INFERENCE_TIMEOUT"

	// EnvInferenceMaxMemory is the environment variable for max memory
	EnvInferenceMaxMemory = "VEID_INFERENCE_MAX_MEMORY_MB"

	// EnvInferenceUseSidecar is the environment variable for sidecar mode
	EnvInferenceUseSidecar = "VEID_INFERENCE_USE_SIDECAR"

	// EnvInferenceSidecarAddr is the environment variable for sidecar address
	EnvInferenceSidecarAddr = "VEID_INFERENCE_SIDECAR_ADDR"

	// EnvInferenceDeterministic is the environment variable for deterministic mode
	EnvInferenceDeterministic = "VEID_INFERENCE_DETERMINISTIC"

	// EnvInferenceForceCPU is the environment variable for CPU-only mode
	EnvInferenceForceCPU = "VEID_INFERENCE_FORCE_CPU"
)

// normalizeExpectedHash strips optional sha256: prefix and normalizes casing.
func normalizeExpectedHash(hash string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(hash)), "sha256:")
	return trimmed
}

var sha256HexRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)

func isValidSHA256Hex(hash string) bool {
	return sha256HexRegex.MatchString(hash)
}
