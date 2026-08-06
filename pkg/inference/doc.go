// Package inference provides TensorFlow-based ML inference for VEID identity
// scoring in the VirtEngine blockchain.
//
// # Overview
//
// This package implements deterministic inference integration for the VEID
// module's identity verification pipeline. It provides:
//
//   - Model bundle loading and verification
//   - Feature extraction from identity verification inputs
//   - Deterministic inference execution for blockchain consensus
//   - Production sidecar execution plus explicit non-production stub paths
//
// # Architecture
//
// The inference package supports two execution modes:
//
// 1. Production Mode: Inference via gRPC to an external sidecar service
//   - Advantages: explicit deployable runtime, model isolation, consensus-safe
//   - Disadvantages: Higher latency, additional deployment complexity
//
// 2. Non-production Stub Mode: deterministic simulated inference
//   - Advantages: local development and tests without serving infrastructure
//   - Disadvantages: not real inference, must be explicitly enabled
//
// Both modes ensure deterministic execution for blockchain consensus by:
//   - Forcing CPU-only execution (GPUs can be non-deterministic)
//   - Setting fixed random seeds
//   - Using single-threaded execution
//   - Computing input/output hashes for verification
//
// # Usage
//
// Basic usage with production sidecar inference:
//
//	config := inference.DefaultInferenceConfig()
//	config.Enabled = true
//	config.UseSidecar = true
//	config.ModelPath = "/path/to/saved_model"
//	config.SidecarAddress = "localhost:50051"
//
//	scorer, err := inference.NewScorer(config)
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
// Using explicit non-production stub mode in tests:
//
//	config := inference.DefaultInferenceConfig()
//	config.AllowFallbackToStub = true
//	config.ModelPath = "/path/to/test/model"
//
//	scorer, err := inference.NewTensorFlowScorer(config)
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
