package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// Model Loader
// ============================================================================

// ModelLoader handles loading TensorFlow SavedModel from disk and verifying
// model integrity via hash comparison.
type ModelLoader struct {
	// config holds the loader configuration
	config InferenceConfig

	// determinism controls deterministic execution
	determinism *DeterminismController

	// mu protects loaded model state
	mu sync.RWMutex

	// loadedModel is the currently loaded model (if any)
	loadedModel *TFModel
}

const defaultInputName = "features"

// NewModelLoader creates a new model loader
func NewModelLoader(config InferenceConfig) *ModelLoader {
	return &ModelLoader{
		config:      config,
		determinism: NewDeterminismController(config.RandomSeed, config.ForceCPU),
	}
}

// ============================================================================
// Model Loading
// ============================================================================

// Load loads a TensorFlow SavedModel from the configured path
func (ml *ModelLoader) Load() (*TFModel, error) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if err := ml.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid inference config: %w", err)
	}

	// Validate model path against traversal attacks
	modelPath := ml.config.ModelPath
	if err := security.ValidateCLIPath(modelPath); err != nil {
		return nil, fmt.Errorf("invalid model path: %w", err)
	}
	cleanPath := filepath.Clean(modelPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("invalid model path: %w", err)
	}
	modelPath = absPath

	// Check if model path exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model path does not exist: %s", modelPath)
	}

	// Load model metadata
	metadata, err := ml.loadMetadata(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load model metadata: %w", err)
	}

	// Compute model hash for verification
	computedHash, err := ml.computeModelHash(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute model hash: %w", err)
	}

	// Verify hash if expected hash is configured
	if ml.config.ExpectedHash != "" {
		expectedHash := normalizeExpectedHash(ml.config.ExpectedHash)
		if computedHash != expectedHash {
			return nil, fmt.Errorf(
				"model hash mismatch: expected %s, got %s",
				expectedHash,
				computedHash,
			)
		}
	}

	// Verify metadata hash if provided
	if metadata.Hash != "" {
		metaHash := normalizeExpectedHash(metadata.Hash)
		if computedHash != metaHash {
			return nil, fmt.Errorf("model metadata hash mismatch: expected %s, got %s", metaHash, computedHash)
		}
	}

	// Validate deterministic ops if provided
	if len(metadata.OpNames) > 0 {
		ok, nonDet := ml.determinism.CheckModelDeterminism(metadata.OpNames)
		if !ok {
			return nil, fmt.Errorf("model uses non-deterministic ops: %s", strings.Join(nonDet, ", "))
		}
	}

	// Verify model version if configured
	if ml.config.ModelVersion != "" && metadata.Version != "" {
		if metadata.Version != ml.config.ModelVersion {
			return nil, fmt.Errorf(
				"model version mismatch: expected %s, got %s",
				ml.config.ModelVersion,
				metadata.Version,
			)
		}
	}

	// Create TF model struct
	// Note: Actual TensorFlow session creation is handled by TFModel.Initialize()
	model := &TFModel{
		modelPath:   modelPath,
		modelHash:   computedHash,
		version:     ml.getVersion(metadata),
		metadata:    metadata,
		config:      ml.config,
		determinism: ml.determinism,
		isLoaded:    false,
		inputName:   ml.getInputName(metadata),
		outputName:  ml.getOutputName(metadata),
		inputShape:  metadata.InputShape,
		outputShape: metadata.OutputShape,
	}

	// Initialize the TensorFlow session
	if err := model.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize TF session: %w", err)
	}

	ml.loadedModel = model
	return model, nil
}

// GetLoadedModel returns the currently loaded model (if any)
func (ml *ModelLoader) GetLoadedModel() *TFModel {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.loadedModel
}

// Unload unloads the current model and releases resources
func (ml *ModelLoader) Unload() error {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.loadedModel != nil {
		if err := ml.loadedModel.Close(); err != nil {
			return fmt.Errorf("failed to close model: %w", err)
		}
		ml.loadedModel = nil
	}

	return nil
}

// ============================================================================
// Metadata Loading
// ============================================================================

// loadMetadata loads the export_metadata.json file from the model directory
func (ml *ModelLoader) loadMetadata(modelPath string) (*ModelMetadata, error) {
	// Look for metadata in the parent directory (version directory)
	baseDir := filepath.Dir(modelPath)
	metadataPath := filepath.Join(baseDir, "export_metadata.json")
	metadataPath, err := security.CleanPathWithinBase(baseDir, metadataPath)
	if err != nil {
		return nil, fmt.Errorf("invalid metadata path: %w", err)
	}

	// If not found, try in the model directory itself
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		metadataPath = filepath.Join(modelPath, "export_metadata.json")
		metadataPath, err = security.CleanPathWithinBase(baseDir, metadataPath)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata path: %w", err)
		}
	}

	// If still not found, return default metadata
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return &ModelMetadata{
			Version:     ml.config.ModelVersion,
			InputName:   defaultInputName,
			OutputName:  "trust_score",
			InputShape:  []int64{-1, TotalFeatureDim},
			OutputShape: []int64{-1, 1},
		}, nil
	}

	// Read and parse metadata
	//nolint:gosec // G304: metadataPath validated against model base directory
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	// Parse JSON
	var rawMetadata struct {
		ModelPath       string                 `json:"model_path"`
		ModelHash       string                 `json:"model_hash"`
		Version         string                 `json:"version"`
		InputSignature  map[string]interface{} `json:"input_signature"`
		OutputSignature map[string]interface{} `json:"output_signature"`
		ExportTimestamp string                 `json:"export_timestamp"`
		TFVersion       string                 `json:"tensorflow_version"`
		OpNames         []string               `json:"op_names"`
		Operations      []string               `json:"operations"`
	}

	if err := json.Unmarshal(data, &rawMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	// Extract input/output names from signatures
	inputName := defaultInputName
	outputName := "trust_score"
	var inputShape, outputShape []int64

	if rawMetadata.InputSignature != nil {
		if name, ok := rawMetadata.InputSignature["name"].(string); ok {
			inputName = name
		}
		if shape, ok := rawMetadata.InputSignature["shape"].([]interface{}); ok {
			inputShape = parseShape(shape)
		}
	}

	if rawMetadata.OutputSignature != nil {
		if name, ok := rawMetadata.OutputSignature["name"].(string); ok {
			outputName = name
		}
		if shape, ok := rawMetadata.OutputSignature["shape"].([]interface{}); ok {
			outputShape = parseShape(shape)
		}
	}

	return &ModelMetadata{
		Version:           rawMetadata.Version,
		Hash:              rawMetadata.ModelHash,
		InputShape:        inputShape,
		OutputShape:       outputShape,
		InputName:         inputName,
		OutputName:        outputName,
		ExportTimestamp:   rawMetadata.ExportTimestamp,
		TensorFlowVersion: rawMetadata.TFVersion,
		OpNames:           mergeOpNames(rawMetadata.OpNames, rawMetadata.Operations),
	}, nil
}

// parseShape converts a JSON shape array to int64 slice
func parseShape(shape []interface{}) []int64 {
	result := make([]int64, len(shape))
	for i, v := range shape {
		switch val := v.(type) {
		case float64:
			result[i] = int64(val)
		case int:
			result[i] = int64(val)
		case nil:
			result[i] = -1 // Null represents dynamic dimension
		}
	}
	return result
}

func mergeOpNames(primary []string, secondary []string) []string {
	if len(primary) == 0 && len(secondary) == 0 {
		return nil
	}

	opSet := make(map[string]struct{})
	merged := make([]string, 0, len(primary)+len(secondary))
	for _, op := range append(primary, secondary...) {
		if op == "" {
			continue
		}
		if _, seen := opSet[op]; seen {
			continue
		}
		opSet[op] = struct{}{}
		merged = append(merged, op)
	}
	sort.Strings(merged)
	return merged
}

// getVersion returns the version string, preferring metadata over config
func (ml *ModelLoader) getVersion(metadata *ModelMetadata) string {
	if metadata != nil && metadata.Version != "" {
		return metadata.Version
	}
	return ml.config.ModelVersion
}

// getInputName returns the input tensor name
func (ml *ModelLoader) getInputName(metadata *ModelMetadata) string {
	if metadata != nil && metadata.InputName != "" {
		return metadata.InputName
	}
	return defaultInputName
}

// getOutputName returns the output tensor name
func (ml *ModelLoader) getOutputName(metadata *ModelMetadata) string {
	if metadata != nil && metadata.OutputName != "" {
		return metadata.OutputName
	}
	return "trust_score"
}

// ============================================================================
// Hash Computation
// ============================================================================

// computeModelHash computes SHA256 hash of all model files
func (ml *ModelLoader) computeModelHash(modelPath string) (string, error) {
	h := sha256.New()

	// Walk through all files in the model directory and sort for determinism
	var files []string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip metadata files (not part of model weights)
		if filepath.Base(path) == "export_metadata.json" {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return "", err
	}

	sort.Strings(files)

	for _, path := range files {
		// Read and hash file contents
		//nolint:gosec // G304: path is from trusted model directory
		file, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer func() { _ = file.Close() }()

		if _, err := io.Copy(h, file); err != nil {
			return "", fmt.Errorf("failed to hash %s: %w", path, err)
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ============================================================================
// TF Model Struct
// ============================================================================

// TFModel represents a loaded TensorFlow SavedModel
type TFModel struct {
	// Model identity
	modelPath string
	modelHash string
	version   string
	metadata  *ModelMetadata

	// Configuration
	config      InferenceConfig
	determinism *DeterminismController

	// TensorFlow session (when using embedded TF)
	// Note: These would be actual TensorFlow types when tf package is imported
	// For now, we use interface{} as placeholders
	session interface{} // *tf.Session
	graph   interface{} // *tf.Graph

	// Input/output tensors
	inputName   string
	outputName  string
	inputShape  []int64
	outputShape []int64

	// State
	isLoaded bool
	mu       sync.RWMutex
}

// Initialize initializes the TensorFlow session
// This is where actual TensorFlow loading would happen
func (m *TFModel) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isLoaded {
		return nil
	}

	// Configure TensorFlow for determinism
	tfConfig := m.determinism.ConfigureTensorFlow()

	// Apply environment variables
	for key, value := range m.determinism.GetTensorFlowEnvVars() {
		_ = os.Setenv(key, value)
	}

	// Note: Actual TensorFlow model loading would happen here
	// Using tensorflow/tensorflow/go package:
	//
	// model, err := tf.LoadSavedModel(m.modelPath, []string{"serve"}, nil)
	// if err != nil {
	//     return fmt.Errorf("failed to load SavedModel: %w", err)
	// }
	// m.session = model.Session
	// m.graph = model.Graph
	//
	// For now, we mark as loaded without actual TF session
	// The actual implementation would require TensorFlow C library

	_ = tfConfig // Use config to avoid unused variable error

	m.isLoaded = true
	return nil
}

// Run executes inference on the given input features
func (m *TFModel) Run(features []float32) ([]float32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isLoaded {
		return nil, fmt.Errorf("model not loaded")
	}

	// Validate input dimension
	if len(features) != TotalFeatureDim {
		return nil, fmt.Errorf(
			"feature dimension mismatch: expected %d, got %d",
			TotalFeatureDim,
			len(features),
		)
	}

	// Note: Actual TensorFlow inference would happen here
	// Using tensorflow/tensorflow/go package:
	//
	// inputTensor, err := tf.NewTensor([][]float32{features})
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create input tensor: %w", err)
	// }
	//
	// result, err := m.session.Run(
	//     map[tf.Output]*tf.Tensor{
	//         m.graph.Operation(m.inputName).Output(0): inputTensor,
	//     },
	//     []tf.Output{
	//         m.graph.Operation(m.outputName).Output(0),
	//     },
	//     nil,
	// )
	// if err != nil {
	//     return nil, fmt.Errorf("inference failed: %w", err)
	// }
	//
	// output := result[0].Value().([][]float32)[0]
	// return output, nil

	// Placeholder: Return stub output for testing without TF library
	// This will be replaced by actual TensorFlow inference
	return m.stubInference(features), nil
}

// stubInference provides a deterministic placeholder for testing
// This will be replaced by actual TensorFlow inference
func (m *TFModel) stubInference(features []float32) []float32 {
	// Compute a deterministic "score" based on feature values
	// This mimics the model's behavior for testing purposes
	var sum float32
	var count float32

	// Weight face embedding more heavily
	for i := 0; i < FaceEmbeddingDim && i < len(features); i++ {
		sum += absFloat32(features[i]) * 0.5
		count++
	}

	// Add document quality contribution
	docOffset := FaceEmbeddingDim
	if docOffset+DocQualityDim <= len(features) {
		for i := 0; i < DocQualityDim; i++ {
			sum += features[docOffset+i] * 1.5
			count++
		}
	}

	// Add OCR contribution
	ocrOffset := FaceEmbeddingDim + DocQualityDim
	if ocrOffset+OCRFieldsDim <= len(features) {
		for i := 0; i < OCRFieldsDim; i++ {
			sum += features[ocrOffset+i]
			count++
		}
	}

	// Normalize to 0-100 range
	if count > 0 {
		rawScore := (sum / count) * 100
		if rawScore > 100 {
			rawScore = 100
		}
		if rawScore < 0 {
			rawScore = 0
		}
		return []float32{rawScore}
	}

	return []float32{0.0}
}

// GetModelHash returns the model's SHA256 hash
func (m *TFModel) GetModelHash() string {
	return m.modelHash
}

// GetVersion returns the model version
func (m *TFModel) GetVersion() string {
	return m.version
}

// GetMetadata returns the model metadata
func (m *TFModel) GetMetadata() *ModelMetadata {
	return m.metadata
}

// IsLoaded returns whether the model is loaded
func (m *TFModel) IsLoaded() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isLoaded
}

// Close releases TensorFlow resources
func (m *TFModel) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isLoaded {
		return nil
	}

	// Note: Actual TensorFlow cleanup would happen here
	// if m.session != nil {
	//     m.session.Close()
	// }

	m.isLoaded = false
	m.session = nil
	m.graph = nil

	return nil
}
