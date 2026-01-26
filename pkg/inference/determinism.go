package inference

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
)

// ============================================================================
// Determinism Controller
// ============================================================================

// DeterminismController ensures deterministic inference execution
// across all validators in the network.
//
// Key responsibilities:
// - Configure TensorFlow for deterministic execution
// - Compute input/output hashes for consensus verification
// - Verify same inputs produce same outputs
type DeterminismController struct {
	// randomSeed is the fixed random seed for all operations
	randomSeed int64

	// forceCPU ensures CPU-only execution (GPUs can introduce non-determinism)
	forceCPU bool

	// disableParallelism disables multi-threaded execution for determinism
	disableParallelism bool

	// hashPrecision is the number of decimal places to use when hashing floats
	hashPrecision int
}

// NewDeterminismController creates a new determinism controller
func NewDeterminismController(seed int64, forceCPU bool) *DeterminismController {
	return &DeterminismController{
		randomSeed:         seed,
		forceCPU:           forceCPU,
		disableParallelism: true,
		hashPrecision:      6, // 6 decimal places
	}
}

// ============================================================================
// TensorFlow Configuration
// ============================================================================

// GetTensorFlowEnvVars returns environment variables for deterministic TensorFlow
func (dc *DeterminismController) GetTensorFlowEnvVars() map[string]string {
	envVars := map[string]string{
		// Disable GPU for determinism
		"CUDA_VISIBLE_DEVICES": "",
		// Use single thread for determinism
		"OMP_NUM_THREADS": "1",
		// TensorFlow determinism settings
		"TF_DETERMINISTIC_OPS":       "1",
		"TF_CUDNN_DETERMINISTIC":     "1",
		"TF_USE_CUDNN_AUTOTUNE":      "0",
		"TF_ENABLE_ONEDNN_OPTS":      "0",
		"TF_CPP_MIN_LOG_LEVEL":       "2",
		"TF_FORCE_GPU_ALLOW_GROWTH":  "false",
		"TF_XLA_FLAGS":               "--tf_xla_auto_jit=-1",
		"PYTHONHASHSEED":             fmt.Sprintf("%d", dc.randomSeed),
	}

	if dc.forceCPU {
		envVars["CUDA_VISIBLE_DEVICES"] = "-1"
	}

	return envVars
}

// ConfigureTensorFlow configures TensorFlow for deterministic execution
// Returns operations/settings that should be applied to the TF session
func (dc *DeterminismController) ConfigureTensorFlow() TFDeterminismConfig {
	return TFDeterminismConfig{
		RandomSeed:              dc.randomSeed,
		InterOpParallelism:      1, // Single inter-op thread
		IntraOpParallelism:      1, // Single intra-op thread
		UseCPUOnly:              dc.forceCPU,
		DisableGPU:              dc.forceCPU,
		EnableDeterministicOps:  true,
		DisableAutoTuning:       true,
		UseFixedRandomGenerator: true,
	}
}

// TFDeterminismConfig contains TensorFlow session configuration for determinism
type TFDeterminismConfig struct {
	// RandomSeed is the global random seed
	RandomSeed int64

	// InterOpParallelism is the number of inter-op threads
	InterOpParallelism int

	// IntraOpParallelism is the number of intra-op threads
	IntraOpParallelism int

	// UseCPUOnly forces CPU execution only
	UseCPUOnly bool

	// DisableGPU hides GPU devices
	DisableGPU bool

	// EnableDeterministicOps uses deterministic operation implementations
	EnableDeterministicOps bool

	// DisableAutoTuning disables cuDNN auto-tuning
	DisableAutoTuning bool

	// UseFixedRandomGenerator uses a fixed random number generator
	UseFixedRandomGenerator bool
}

// ============================================================================
// Hash Computation
// ============================================================================

// ComputeInputHash computes a deterministic SHA256 hash of the score inputs
// This hash is used for consensus verification - all validators should
// produce the same hash for the same inputs
func (dc *DeterminismController) ComputeInputHash(inputs *ScoreInputs) string {
	h := sha256.New()

	// Hash account address and metadata
	h.Write([]byte(inputs.Metadata.AccountAddress))

	blockHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockHeightBytes, uint64(inputs.Metadata.BlockHeight))
	h.Write(blockHeightBytes)

	// Hash face embedding (normalize to fixed precision)
	for _, val := range inputs.FaceEmbedding {
		dc.writeFloat32(h, val)
	}

	// Hash face confidence
	dc.writeFloat32(h, inputs.FaceConfidence)

	// Hash document quality score
	dc.writeFloat32(h, inputs.DocQualityScore)

	// Hash document quality features
	dc.writeFloat32(h, inputs.DocQualityFeatures.Sharpness)
	dc.writeFloat32(h, inputs.DocQualityFeatures.Brightness)
	dc.writeFloat32(h, inputs.DocQualityFeatures.Contrast)
	dc.writeFloat32(h, inputs.DocQualityFeatures.NoiseLevel)
	dc.writeFloat32(h, inputs.DocQualityFeatures.BlurScore)

	// Hash OCR confidences (sorted by key for determinism)
	ocrKeys := make([]string, 0, len(inputs.OCRConfidences))
	for k := range inputs.OCRConfidences {
		ocrKeys = append(ocrKeys, k)
	}
	sort.Strings(ocrKeys)

	for _, key := range ocrKeys {
		h.Write([]byte(key))
		dc.writeFloat32(h, inputs.OCRConfidences[key])
	}

	// Hash OCR field validation (sorted by key)
	validationKeys := make([]string, 0, len(inputs.OCRFieldValidation))
	for k := range inputs.OCRFieldValidation {
		validationKeys = append(validationKeys, k)
	}
	sort.Strings(validationKeys)

	for _, key := range validationKeys {
		h.Write([]byte(key))
		if inputs.OCRFieldValidation[key] {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	}

	// Hash scope information
	sort.Strings(inputs.ScopeTypes)
	for _, scopeType := range inputs.ScopeTypes {
		h.Write([]byte(scopeType))
	}

	scopeCountBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(scopeCountBytes, uint32(inputs.ScopeCount))
	h.Write(scopeCountBytes)

	return hex.EncodeToString(h.Sum(nil))
}

// ComputeOutputHash computes a deterministic SHA256 hash of the raw model output
// This is used to verify that all validators produce identical inference results
func (dc *DeterminismController) ComputeOutputHash(rawOutput []float32) string {
	h := sha256.New()

	for _, val := range rawOutput {
		dc.writeFloat32(h, val)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ComputeResultHash computes a hash of the complete score result
func (dc *DeterminismController) ComputeResultHash(result *ScoreResult) string {
	h := sha256.New()

	// Hash score
	scoreBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(scoreBytes, result.Score)
	h.Write(scoreBytes)

	// Hash confidence
	dc.writeFloat32(h, result.Confidence)

	// Hash model version
	h.Write([]byte(result.ModelVersion))

	// Hash model hash
	h.Write([]byte(result.ModelHash))

	// Hash raw score
	dc.writeFloat32(h, result.RawScore)

	// Hash reason codes (sorted)
	sortedReasons := make([]string, len(result.ReasonCodes))
	copy(sortedReasons, result.ReasonCodes)
	sort.Strings(sortedReasons)
	for _, reason := range sortedReasons {
		h.Write([]byte(reason))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// writeFloat32 writes a float32 to the hash in a deterministic way
func (dc *DeterminismController) writeFloat32(h interface{ Write([]byte) (int, error) }, val float32) {
	// Round to fixed precision to handle floating point variations
	multiplier := math.Pow(10, float64(dc.hashPrecision))
	rounded := math.Round(float64(val)*multiplier) / multiplier

	// Convert to bytes using IEEE 754 representation
	bits := math.Float32bits(float32(rounded))
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, bits)
	h.Write(bytes)
}

// ============================================================================
// Verification
// ============================================================================

// VerifySameOutput checks if two score results are deterministically equivalent
func (dc *DeterminismController) VerifySameOutput(result1, result2 *ScoreResult) bool {
	// Compare scores (must match exactly)
	if result1.Score != result2.Score {
		return false
	}

	// Compare model versions
	if result1.ModelVersion != result2.ModelVersion {
		return false
	}

	// Compare model hashes
	if result1.ModelHash != result2.ModelHash {
		return false
	}

	// Compare raw scores with tolerance for floating point
	if !dc.floatsEqual(result1.RawScore, result2.RawScore) {
		return false
	}

	// Compare confidence with tolerance
	if !dc.floatsEqual(result1.Confidence, result2.Confidence) {
		return false
	}

	// Compare input hashes
	if result1.InputHash != result2.InputHash {
		return false
	}

	// Compare output hashes
	if result1.OutputHash != result2.OutputHash {
		return false
	}

	return true
}

// floatsEqual compares two floats with precision tolerance
func (dc *DeterminismController) floatsEqual(a, b float32) bool {
	multiplier := math.Pow(10, float64(dc.hashPrecision))
	roundedA := math.Round(float64(a)*multiplier) / multiplier
	roundedB := math.Round(float64(b)*multiplier) / multiplier
	return roundedA == roundedB
}

// ============================================================================
// Non-Deterministic Operation Detection
// ============================================================================

// NonDeterministicOps lists TensorFlow operations that can be non-deterministic
var NonDeterministicOps = []string{
	"CudnnRNN",
	"BiasAdd", // Can be non-deterministic with cuDNN
	"Conv2D",  // Can be non-deterministic with cuDNN
	"MaxPool",
	"AvgPool",
	"CTCLoss",
	"CTCGreedyDecoder",
	"SoftmaxCrossEntropyWithLogits",
	"SparseSoftmaxCrossEntropyWithLogits",
}

// CheckModelDeterminism validates that a model uses only deterministic operations
// Note: This is a best-effort check; some ops may be deterministic with proper configuration
func (dc *DeterminismController) CheckModelDeterminism(opNames []string) (bool, []string) {
	var nonDetOps []string

	for _, op := range opNames {
		for _, nonDet := range NonDeterministicOps {
			if op == nonDet {
				nonDetOps = append(nonDetOps, op)
			}
		}
	}

	return len(nonDetOps) == 0, nonDetOps
}
