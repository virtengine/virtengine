// Package inference provides TensorFlow-based ML inference for VEID identity
// scoring in the VirtEngine blockchain.
//
// # Overview
//
// This package implements TensorFlow-Go inference integration for the VEID
// module's identity verification pipeline. It provides:
//
//   - TensorFlow SavedModel loading and verification
//   - Feature extraction from identity verification inputs
//   - Deterministic inference execution for blockchain consensus
//   - Both embedded TensorFlow and gRPC sidecar modes
//
// # Architecture
//
// The inference package supports two execution modes:
//
// 1. Embedded Mode: TensorFlow-Go runs directly in the node process
//   - Advantages: Lower latency, simpler deployment
//   - Disadvantages: Requires TensorFlow C library, larger binary
//
// 2. Sidecar Mode: Inference via gRPC to an external service
//   - Advantages: Language-agnostic, can use GPU, memory isolation
//   - Disadvantages: Higher latency, additional deployment complexity
//
// Both modes ensure deterministic execution for blockchain consensus by:
//   - Forcing CPU-only execution (GPUs can be non-deterministic)
//   - Setting fixed random seeds
//   - Using single-threaded execution
//   - Computing input/output hashes for verification
//
// # Usage
//
// Basic usage with embedded TensorFlow:
//
//	config := inference.DefaultInferenceConfig()
//	config.ModelPath = "/path/to/saved_model"
//
//	scorer, err := inference.NewTensorFlowScorer(config)
//	if err != nil {
//	    return err
//	}
//	defer scorer.Close()
//
//	inputs := &inference.ScoreInputs{
//	    FaceEmbedding:   faceEmbedding,
//	    DocQualityScore: docQuality,
//	    OCRConfidences:  ocrScores,
//	    // ... other fields
//	}
//
//	result, err := scorer.ComputeScore(inputs)
//	if err != nil {
//	    return err
//	}
//
//	fmt.Printf("Score: %d, Confidence: %.2f\n", result.Score, result.Confidence)
//
// Using the sidecar client:
//
//	config := inference.DefaultInferenceConfig()
//	config.UseSidecar = true
//	config.SidecarAddress = "localhost:50051"
//
//	scorer, err := inference.NewScorer(config)  // Returns sidecar client
//	// ... rest is identical
//
// # Feature Vector Format
//
// The feature vector sent to the model has 768 dimensions:
//
//   - [0-511]:    Face embedding (512-dim from facial verification)
//   - [512-516]:  Document quality features (5 values)
//   - [517-526]:  OCR features (5 fields × 2 values each)
//   - [527-542]:  Metadata features (16 values)
//   - [543-767]:  Reserved/padding (225 values)
//
// # Determinism Guarantees
//
// For blockchain consensus, all validators must produce identical scores
// for identical inputs. This is achieved by:
//
//   - Setting TF_DETERMINISTIC_OPS=1
//   - Disabling GPU (CUDA_VISIBLE_DEVICES=-1)
//   - Using single inter-op and intra-op threads
//   - Computing SHA256 hashes of inputs and outputs
//   - Verifying model hash matches expected value
//
// # Model Requirements
//
// The TensorFlow model must be exported in SavedModel format with:
//
//   - Input tensor: "features" with shape [batch, 768] and dtype float32
//   - Output tensor: "trust_score" with shape [batch, 1] and dtype float32
//   - Output range: 0-100 (sigmoid scaled)
//
// See ml/training/model/export.py for the model export code.
//
// # Configuration
//
// Configuration can be provided via InferenceConfig struct or environment
// variables:
//
//   - VEID_INFERENCE_MODEL_PATH: Path to SavedModel directory
//   - VEID_INFERENCE_MODEL_VERSION: Expected model version
//   - VEID_INFERENCE_MODEL_HASH: Expected SHA256 hash of model (required in deterministic mode)
//   - VEID_INFERENCE_TIMEOUT: Max inference time (e.g., "2s")
//   - VEID_INFERENCE_USE_SIDECAR: Enable sidecar mode ("true"/"false")
//   - VEID_INFERENCE_SIDECAR_ADDR: Sidecar gRPC address
//   - VEID_INFERENCE_DETERMINISTIC: Force deterministic mode
//
// # Error Handling
//
// Errors are categorized with reason codes:
//
//   - SUCCESS: Score computed successfully
//   - HIGH_CONFIDENCE / LOW_CONFIDENCE: Model confidence level
//   - FACE_MISMATCH: Face verification issues
//   - LOW_DOC_QUALITY: Document quality below threshold
//   - LOW_OCR_CONFIDENCE: OCR extraction confidence low
//   - INSUFFICIENT_SCOPES: Not enough identity data
//   - TIMEOUT: Inference exceeded time limit
//   - INFERENCE_ERROR: Model execution failed
//
// When UseFallbackOnError is enabled, errors return a fallback score instead
// of failing the request.
//
// # Testing
//
// Run tests with:
//
//	go test ./pkg/inference/...
//
// Benchmarks:
//
//	go test -bench=. ./pkg/inference/...
package inference
