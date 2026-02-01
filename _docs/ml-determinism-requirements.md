# ML Determinism Requirements for VirtEngine VEID

**Version:** 1.0.0  
**Last Updated:** 2024-01-15  
**Status:** Production Ready  
**Task Reference:** 15A - ML determinism validation + conformance suite

## Executive Summary

This document specifies the mandatory requirements for deterministic ML inference in VirtEngine's VEID (Verified Encrypted Identity) system. **Determinism is CRITICAL for blockchain consensus** - all validators must produce identical scores for the same identity verification inputs.

## Table of Contents

1. [Why Determinism Matters](#why-determinism-matters)
2. [Production Requirements](#production-requirements)
3. [TensorFlow Configuration](#tensorflow-configuration)
4. [Environment Variables](#environment-variables)
5. [Model Hash Pinning](#model-hash-pinning)
6. [Cross-Machine Validation](#cross-machine-validation)
7. [Floating-Point Handling](#floating-point-handling)
8. [Conformance Testing](#conformance-testing)
9. [Troubleshooting](#troubleshooting)
10. [Verification Checklist](#verification-checklist)

## Why Determinism Matters

### Blockchain Consensus Requirements

In a blockchain network, all validator nodes must independently verify transactions and arrive at the **exact same result**. For identity verification:

1. **Validator A** receives identity verification request
2. **Validator A** computes ML score: `75.123456`
3. **Validator B** independently verifies: must also get `75.123456`
4. **Validators C, D, E...** must all get `75.123456`

If validators produce different scores:
- ❌ Consensus fails
- ❌ Block production halts
- ❌ Network integrity compromised
- ❌ Identity verification becomes unreliable

### Sources of Non-Determinism

| Source | Risk Level | Mitigation |
|--------|------------|------------|
| GPU computation | 🔴 Critical | **Force CPU-only mode** |
| Multi-threading | 🔴 Critical | Single-threaded execution |
| Random number generators | 🔴 Critical | Fixed seed (42) |
| Floating-point precision | 🟡 High | Truncate to 6 decimal places |
| TensorFlow auto-tuning | 🟡 High | Disable all auto-tuning |
| cuDNN algorithms | 🟡 High | Force deterministic algorithms |
| Model weight loading | 🟢 Medium | SHA256 hash verification |
| Feature extraction order | 🟢 Medium | Canonical ordering |

## Production Requirements

### Mandatory Configuration

```go
// These values MUST NEVER change without a network upgrade proposal
const (
    // ProductionRandomSeed is the ONLY acceptable random seed for production.
    // Changing this value will break consensus across all validators.
    ProductionRandomSeed = 42
    
    // ProductionHashPrecision defines decimal places for float hashing.
    // This ensures cross-platform consistency despite FPU differences.
    ProductionHashPrecision = 6
    
    // ExpectedModelHashV1Production is the SHA256 hash of the production model.
    // This MUST be updated through governance when models are upgraded.
    ExpectedModelHashV1Production = "a1b2c3d4e5f6..."
)
```

### CPU-Only Enforcement

**GPU inference is NEVER allowed in production consensus.** GPUs introduce non-determinism due to:

- Parallel reduction algorithms
- cuDNN algorithm selection
- Hardware-specific optimizations
- Memory access patterns

```go
// Validation enforces CPU-only mode
func (c *ProductionDeterminismConfig) Validate() []string {
    var issues []string
    
    if !c.ForceCPU {
        issues = append(issues, "CRITICAL: ForceCPU must be true for production consensus")
    }
    
    if os.Getenv("CUDA_VISIBLE_DEVICES") != "-1" {
        issues = append(issues, "CRITICAL: CUDA_VISIBLE_DEVICES must be set to -1")
    }
    
    return issues
}
```

### Single-Threaded Execution

Multi-threading introduces race conditions and non-deterministic ordering:

```go
type ProductionDeterminismConfig struct {
    // Both parallelism values MUST be 1
    InterOpParallelism int // Must be 1
    IntraOpParallelism int // Must be 1
}
```

## TensorFlow Configuration

### Required TensorFlow Settings

```go
func ConfigureProductionTensorFlow() *TensorFlowConfig {
    return &TensorFlowConfig{
        // Mandatory determinism settings
        RandomSeed:              42,
        UseCPUOnly:              true,
        EnableDeterministicOps:  true,
        DisableAutoTuning:       true,
        
        // Parallelism constraints
        InterOpParallelism:      1,
        IntraOpParallelism:      1,
        
        // Precision settings
        UseXLA:                  false, // XLA can introduce variance
        UseMixedPrecision:       false, // Full precision required
    }
}
```

### Deterministic Operations Registry

These TensorFlow operations are verified deterministic:

```go
var DeterministicOps = []string{
    // Math operations
    "Add", "Sub", "Mul", "Div", "Neg", "Abs", "Square", "Sqrt",
    "Exp", "Log", "Pow", "Maximum", "Minimum",
    
    // Matrix operations
    "MatMul", "BatchMatMul", "Transpose", "Reshape",
    
    // Activation functions
    "Relu", "Relu6", "Sigmoid", "Tanh", "Softmax", "LogSoftmax",
    
    // Reduction operations (on CPU with single thread)
    "Sum", "Mean", "Max", "Min", "Prod", "ArgMax", "ArgMin",
    
    // Data operations
    "Const", "Identity", "Placeholder", "Cast", "Pack", "Unpack",
    "Concat", "Split", "Slice", "Gather", "Squeeze", "ExpandDims",
    
    // Normalization (with fixed moments)
    "FusedBatchNorm", "LayerNorm",
}
```

### Prohibited Operations

These operations are NON-DETERMINISTIC and blocked in production:

```go
var NonDeterministicOpsProduction = []string{
    // GPU-specific operations
    "CudnnRNN", "CudnnRNNV2", "CudnnRNNV3",
    "CudnnRNNBackprop", "CudnnRNNBackpropV2",
    
    // Dropout (even with seed, implementation varies)
    "Dropout", "SparseDropout",
    
    // Non-deterministic reductions on GPU
    "UnsortedSegmentSum", "UnsortedSegmentProd",
    
    // Floating point atomic operations
    "ResourceScatterAdd", "ScatterNd",
    
    // Variable random operations
    "RandomUniform", "RandomNormal", "RandomShuffle",
    "Multinomial", "ParameterizedTruncatedNormal",
}
```

### Conditionally Deterministic Operations

These operations are deterministic ONLY under specific conditions:

| Operation | Determinism Condition |
|-----------|----------------------|
| `Conv2D` | CPU-only, single thread, fixed algorithm |
| `Conv2DBackprop*` | CPU-only, single thread |
| `BiasAdd` | CPU-only, single thread |
| `MaxPool` | CPU-only, single thread |
| `AvgPool` | CPU-only, single thread |

## Environment Variables

### Required Environment Setup

These environment variables MUST be set before any TensorFlow operations:

```bash
# TensorFlow determinism
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export TF_USE_CUDNN_AUTOTUNE=0

# Disable GPU entirely
export CUDA_VISIBLE_DEVICES=-1
export TF_FORCE_GPU_ALLOW_GROWTH=false

# Thread control
export OMP_NUM_THREADS=1
export MKL_NUM_THREADS=1
export OPENBLAS_NUM_THREADS=1
export NUMEXPR_NUM_THREADS=1

# Python reproducibility
export PYTHONHASHSEED=42
export PYTHONDONTWRITEBYTECODE=1
```

### Verification Script

```go
func VerifyDeterminismEnvironment() error {
    required := map[string]string{
        "TF_DETERMINISTIC_OPS":     "1",
        "TF_CUDNN_DETERMINISTIC":   "1",
        "OMP_NUM_THREADS":          "1",
        "CUDA_VISIBLE_DEVICES":     "-1",
    }
    
    for key, expected := range required {
        actual := os.Getenv(key)
        if actual != expected {
            return fmt.Errorf("env %s: expected %s, got %s", key, expected, actual)
        }
    }
    return nil
}
```

## Model Hash Pinning

### Why Pin Model Hashes?

Model files must be byte-for-byte identical across all validators. Hash pinning ensures:

1. No accidental model file corruption
2. No unauthorized model modifications
3. All validators use exact same weights
4. Model upgrades are coordinated via governance

### Pinned Model Registry

```go
type PinnedModelRegistry struct {
    Models map[string]PinnedModelInfo `json:"models"`
}

type PinnedModelInfo struct {
    ModelName       string    `json:"model_name"`
    Version         string    `json:"version"`
    SHA256Hash      string    `json:"sha256_hash"`      // Primary verification
    MD5Hash         string    `json:"md5_hash"`         // Secondary verification
    FileSize        int64     `json:"file_size"`
    PinnedAt        time.Time `json:"pinned_at"`
    GovernanceRef   string    `json:"governance_ref"`   // Upgrade proposal reference
}
```

### Hash Computation

```go
func ComputeModelHash(modelDir string) (string, error) {
    hasher := sha256.New()
    
    // Walk model files in deterministic order
    files, _ := filepath.Glob(filepath.Join(modelDir, "*"))
    sort.Strings(files)
    
    for _, file := range files {
        // Skip non-essential files
        if filepath.Base(file) == "export_metadata.json" {
            continue
        }
        
        data, err := os.ReadFile(file)
        if err != nil {
            return "", err
        }
        hasher.Write(data)
    }
    
    return hex.EncodeToString(hasher.Sum(nil)), nil
}
```

### Model Loading with Verification

```go
func LoadModel(modelDir string, expectedHash string) (*Model, error) {
    // 1. Compute model hash
    actualHash, err := ComputeModelHash(modelDir)
    if err != nil {
        return nil, fmt.Errorf("hash computation failed: %w", err)
    }
    
    // 2. Verify hash matches
    if actualHash != expectedHash {
        return nil, fmt.Errorf(
            "model hash mismatch: expected %s, got %s - CONSENSUS FAILURE RISK",
            expectedHash, actualHash,
        )
    }
    
    // 3. Load model with verified hash
    return loadVerifiedModel(modelDir)
}
```

## Cross-Machine Validation

### Conformance Test Suite

The conformance suite validates determinism across different machines:

```go
func TestMultiMachineConformanceSuite(t *testing.T) {
    suite := NewMultiMachineConformanceSuite(t)
    
    // Phase 1: Configuration validation
    t.Run("ConfigurationValidation", suite.TestProductionConfigurationValid)
    
    // Phase 2: TensorFlow operation validation
    t.Run("OperationValidation", suite.TestDeterministicOpsRegistryComplete)
    
    // Phase 3: Hash computation determinism
    t.Run("HashDeterminism", suite.TestInputHashDeterministic)
    
    // Phase 4: Feature extraction determinism
    t.Run("FeatureExtraction", suite.TestFeatureVectorDeterministic)
    
    // Phase 5: Golden vector tests
    t.Run("GoldenVectors", suite.TestGoldenVectorHashesMatch)
    
    // Phase 6: Cross-run consistency
    t.Run("CrossRunConsistency", suite.TestRepeatedHashingIdentical)
    
    // Generate evidence file for cross-machine comparison
    suite.GenerateEvidence(t)
}
```

### Evidence Files

Each machine generates an evidence file containing:

```json
{
  "report_id": "conf-linux-amd64-1705312345",
  "version": "1.0.0",
  "platform": {
    "os": "linux",
    "arch": "amd64",
    "go_version": "go1.22.0",
    "num_cpu": 16
  },
  "config": {
    "random_seed": 42,
    "force_cpu": true,
    "hash_precision": 6
  },
  "results": [
    {
      "test_name": "InputHashDeterministic",
      "passed": true,
      "hash": "a1b2c3d4e5f6..."
    }
  ]
}
```

### Cross-Platform Comparison

Evidence files from different machines are compared in CI:

```yaml
cross-platform-hashes:
  name: Cross-Platform Hash Verification
  needs: go-inference-determinism
  steps:
    - name: Download all evidence
      uses: actions/download-artifact@v4
      with:
        pattern: conformance-evidence-*
        
    - name: Compare cross-platform hashes
      run: |
        go run ./scripts/verify-cross-platform-hashes.go ./all-evidence/
```

## Floating-Point Handling

### The Floating-Point Problem

Different CPUs may produce slightly different floating-point results due to:

- FPU rounding modes
- Extended precision registers (x87 vs SSE)
- Fused multiply-add availability
- Compiler optimizations

### Solution: Hash Precision Truncation

All floating-point values are truncated to 6 decimal places before hashing:

```go
const ProductionHashPrecision = 6

func truncateForHash(value float32) float32 {
    multiplier := math.Pow10(ProductionHashPrecision)
    return float32(math.Trunc(float64(value)*multiplier) / multiplier)
}

func ComputeOutputHash(outputs []float32) string {
    hasher := sha256.New()
    for _, v := range outputs {
        truncated := truncateForHash(v)
        binary.Write(hasher, binary.LittleEndian, truncated)
    }
    return hex.EncodeToString(hasher.Sum(nil))
}
```

### Precision Trade-offs

| Precision | Risk of False Mismatch | Accuracy Impact |
|-----------|------------------------|-----------------|
| 8 decimals | High | Minimal |
| 7 decimals | Medium | Minimal |
| **6 decimals** | **Low (chosen)** | **Acceptable** |
| 5 decimals | Very Low | Noticeable |
| 4 decimals | None | Significant |

## Conformance Testing

### Running Conformance Tests

```bash
# Run full conformance suite
go test -v ./pkg/inference/... -run "TestMultiMachineConformanceSuite"

# Run with evidence generation
VEID_CONFORMANCE_EVIDENCE_DIR=./evidence go test -v ./pkg/inference/...

# Run extended conformance (stress testing)
VEID_EXTENDED_CONFORMANCE=true go test -v -count=10 ./pkg/inference/...
```

### Golden Vectors

Pre-computed test vectors with known-correct hashes:

```go
var GoldenVectors = []GoldenVector{
    {
        ID:          "standard_001",
        Description: "Standard high-quality identity verification",
        Inputs:      createStandardInputs(),
        ExpectedHash: "a1b2c3d4e5f6...",
        ExpectedScore: 85,
        ExpectedConfidence: 0.92,
    },
    // ... more vectors
}
```

### CI Workflow

The `ml-determinism.yaml` workflow runs on every PR:

1. **Go Inference Determinism** - Tests on Ubuntu, Windows, macOS
2. **Cross-Platform Hash Verification** - Compares hashes across platforms
3. **Python ML Determinism** - Validates Python pipeline
4. **Model Hash Verification** - Checks pinned model hashes
5. **Configuration Validation** - Verifies production settings

## Troubleshooting

### Common Issues

#### Hash Mismatch Across Machines

**Symptom:** Different machines produce different hashes for same input.

**Diagnosis:**
```bash
# Check environment on both machines
env | grep -E "(TF_|OMP_|MKL_|CUDA_)"

# Verify same Go version
go version

# Check CPU features
lscpu | grep -E "(Model|Flags)"
```

**Solutions:**
1. Ensure all environment variables are set identically
2. Verify same model files (check SHA256)
3. Confirm single-threaded execution
4. Check for GPU usage (must be disabled)

#### Score Drift Over Time

**Symptom:** Same inputs produce slightly different scores over time.

**Diagnosis:**
```go
for i := 0; i < 1000; i++ {
    score := computeScore(input)
    if score != expectedScore {
        log.Printf("Drift detected at iteration %d: %f", i, score)
    }
}
```

**Solutions:**
1. Check for memory corruption
2. Verify model hasn't been modified
3. Look for race conditions in multi-threaded code
4. Ensure random seed is set before each run

#### Conformance Test Failures

**Symptom:** Conformance suite fails intermittently.

**Diagnosis:**
```bash
# Run with verbose logging
go test -v -count=10 ./pkg/inference/... -run "Conformance" 2>&1 | tee conformance.log

# Look for variance patterns
grep -E "(mismatch|failed|error)" conformance.log
```

**Solutions:**
1. Run tests in isolated environment
2. Check for concurrent test interference
3. Verify no background processes affecting CPU
4. Increase test iterations to find rare failures

## Verification Checklist

### Pre-Deployment

- [ ] `ProductionRandomSeed` is 42
- [ ] `ForceCPU` is true
- [ ] `InterOpParallelism` is 1
- [ ] `IntraOpParallelism` is 1
- [ ] `TF_DETERMINISTIC_OPS=1` is set
- [ ] `CUDA_VISIBLE_DEVICES=-1` is set
- [ ] Model hash matches pinned registry
- [ ] Conformance tests pass locally
- [ ] CI workflow passes on all platforms

### Validator Node Setup

```bash
# 1. Set environment variables
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export OMP_NUM_THREADS=1
export CUDA_VISIBLE_DEVICES=-1

# 2. Verify model hash
virtengine verify-model --expected-hash "a1b2c3d4..."

# 3. Run local conformance
virtengine conformance-check

# 4. Start validator
virtengined start --veid-determinism-strict
```

### Model Upgrade Process

1. **Governance Proposal** - Submit upgrade proposal with new model hash
2. **Network Vote** - Validators vote on upgrade
3. **Upgrade Block** - At specified block height, all validators switch
4. **Hash Verification** - Each validator verifies new model hash
5. **Conformance Check** - Run conformance suite with new model

---

## References

- [TensorFlow Determinism Guide](https://www.tensorflow.org/guide/random_numbers)
- [NVIDIA Determinism Documentation](https://docs.nvidia.com/deeplearning/cudnn/developer-guide/index.html#reproducibility)
- [IEEE 754 Floating-Point Standard](https://ieeexplore.ieee.org/document/8766229)
- [Cosmos SDK Consensus Documentation](https://docs.cosmos.network/main/learn/intro/sdk-design)

---

*Document maintained by VirtEngine Core Team*  
*For questions, open an issue at https://github.com/virtengine/virtengine*
