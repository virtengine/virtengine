# ML Consensus Requirements for VirtEngine VEID

**Version:** 2.0.0  
**Last Updated:** 2024-01-15  
**Status:** Production Ready  
**Task Reference:** 15A - ML determinism validation + conformance suite

## Executive Summary

This document specifies the requirements for achieving ML consensus in VirtEngine's VEID (Verified Encrypted Identity) system. **Consensus is achieved when all validators produce scores within an acceptable tolerance** - not necessarily bit-exact identical results.

### Key Design Principle

> **Tolerance-Based Consensus**: Validators may use different hardware (GPU/CPU) and threading configurations, as long as their scores fall within the defined consensus tolerance (±2 points, or ±0.5 near thresholds).

This approach allows for:
- GPU acceleration for faster inference
- Multi-threaded execution for performance
- Practical deployment across heterogeneous infrastructure
- Robust consensus even with minor floating-point variance

## Table of Contents

1. [Why Tolerance-Based Consensus](#why-tolerance-based-consensus)
2. [Consensus Parameters](#consensus-parameters)
3. [Score Binning for Hashing](#score-binning-for-hashing)
4. [TensorFlow Configuration](#tensorflow-configuration)
5. [Model Hash Pinning](#model-hash-pinning)
6. [Cross-Machine Validation](#cross-machine-validation)
7. [Threshold Handling](#threshold-handling)
8. [Conformance Testing](#conformance-testing)
9. [Troubleshooting](#troubleshooting)
10. [Verification Checklist](#verification-checklist)

## Why Tolerance-Based Consensus

### The Reality of ML Inference

Machine learning inference inherently has minor variance across different hardware:

| Source | Typical Variance | Impact |
|--------|-----------------|--------|
| GPU vs CPU | ±0.5 points | Normal |
| Different GPU models | ±0.3 points | Normal |
| Multi-threading | ±0.1 points | Normal |
| Floating-point rounding | ±0.01 points | Negligible |

**These variances do NOT affect identity verification outcomes** because:
- A score of 75.1 vs 75.3 both mean "PASS"
- A score of 85.2 vs 85.5 both mean "HIGH TRUST"
- Only threshold crossings matter for consensus

### Strict Determinism vs Tolerance Consensus

| Approach | Strict Determinism | Tolerance Consensus |
|----------|-------------------|---------------------|
| Score match | Bit-exact (75.123456) | Within tolerance (75.0 ± 2.0) |
| Hardware | CPU-only, single-thread | GPU allowed, parallel OK |
| Performance | Slow | Fast |
| Deployment | Complex | Flexible |
| Consensus | Fragile | Robust |
| **VirtEngine Choice** | ❌ | ✅ |

## Consensus Parameters

### Core Constants

```go
const (
    // ConsensusTolerance is the maximum allowed deviation between validator scores.
    // Scores within ±2 points are considered equivalent for consensus.
    ConsensusTolerance = 2.0

    // ConsensusToleranceStrict is used near decision thresholds.
    // Within 5 points of a threshold, we use stricter tolerance.
    ConsensusToleranceStrict = 0.5

    // ScoreThresholdPass is the minimum score for identity verification pass.
    ScoreThresholdPass = 60.0

    // ScoreThresholdHighTrust is the threshold for high-trust identity tier.
    ScoreThresholdHighTrust = 80.0

    // ThresholdBuffer defines the zone around thresholds where strict tolerance applies.
    ThresholdBuffer = 5.0

    // MinValidatorsForConsensus is the minimum validators needed for consensus.
    MinValidatorsForConsensus = 3
)
```

### Consensus Validation Function

```go
// ValidateConsensus checks if validator scores achieve consensus.
func ValidateConsensus(scores []float32) ConsensusResult {
    if len(scores) < MinValidatorsForConsensus {
        return ConsensusResult{
            Achieved: len(scores) == 1, // Single validator always achieves
            Reason:   "insufficient validators",
        }
    }

    // Calculate median score
    median := calculateMedian(scores)
    
    // Determine effective tolerance based on threshold proximity
    tolerance := GetEffectiveTolerance(median)
    
    // Check all scores within tolerance of median
    for _, score := range scores {
        deviation := math.Abs(float64(score - median))
        if deviation > tolerance {
            return ConsensusResult{
                Achieved:     false,
                Reason:       "score deviation exceeds tolerance",
                MaxDeviation: deviation,
            }
        }
    }

    return ConsensusResult{
        Achieved:    true,
        MedianScore: median,
        Tolerance:   tolerance,
    }
}
```

## Score Binning for Hashing

### Why Binning?

When storing scores on-chain or computing verification hashes, we "bin" scores to reduce minor variance:

```go
// BinScore rounds to nearest 0.5 for hash computation.
// This ensures scores like 75.1, 75.2, 75.3, 75.4 all hash to 75.0 or 75.5.
func BinScore(score float32) float32 {
    return float32(math.Round(float64(score)*2) / 2)
}

// BinScoreInt rounds to nearest integer for maximum determinism.
func BinScoreInt(score float32) int {
    return int(math.Round(float64(score)))
}
```

### Binning Examples

| Raw Score | Binned (0.5) | Binned (int) |
|-----------|--------------|--------------|
| 75.1 | 75.0 | 75 |
| 75.3 | 75.5 | 75 |
| 75.6 | 75.5 | 76 |
| 75.8 | 76.0 | 76 |

## TensorFlow Configuration

### Recommended Settings (Performance Mode)

```go
func ConfigureProductionTensorFlow() *TensorFlowConfig {
    return &TensorFlowConfig{
        // Performance settings (GPU/parallel allowed)
        UseCPUOnly:              false, // GPU allowed
        InterOpParallelism:      0,     // Auto (system decides)
        IntraOpParallelism:      0,     // Auto (system decides)
        
        // Still recommended for reduced variance
        RandomSeed:              42,
        EnableDeterministicOps:  true,  // Use deterministic ops where available
        DisableAutoTuning:       false, // Auto-tune OK for performance
        
        // Precision settings
        UseXLA:                  true,  // XLA OK for performance
        UseMixedPrecision:       false, // Full precision still recommended
    }
}
```

### Strict Mode (For Testing)

```go
// NewStrictDeterminismConfig returns CPU-only, single-threaded config
// for testing or when bit-exact matching is needed.
func NewStrictDeterminismConfig() *ProductionDeterminismConfig {
    return &ProductionDeterminismConfig{
        RandomSeed:         42,
        ForceCPU:           true,
        StrictMode:         true,
        AllowGPU:           false,
        AllowParallelism:   false,
        InterOpParallelism: 1,
        IntraOpParallelism: 1,
    }
}
```

## Model Hash Pinning

Model files must still be byte-for-byte identical across validators:

```go
type PinnedModelInfo struct {
    ModelName       string    `json:"model_name"`
    Version         string    `json:"version"`
    SHA256Hash      string    `json:"sha256_hash"`
    FileSize        int64     `json:"file_size"`
    PinnedAt        time.Time `json:"pinned_at"`
    GovernanceRef   string    `json:"governance_ref"`
}
```

### Model Loading with Verification

```go
func LoadModel(modelDir string, expectedHash string) (*Model, error) {
    actualHash, err := ComputeModelHash(modelDir)
    if err != nil {
        return nil, fmt.Errorf("hash computation failed: %w", err)
    }
    
    if actualHash != expectedHash {
        return nil, fmt.Errorf(
            "model hash mismatch: expected %s, got %s",
            expectedHash, actualHash,
        )
    }
    
    return loadVerifiedModel(modelDir)
}
```

## Cross-Machine Validation

### Conformance Test Suite

The conformance suite validates consensus across different machines:

```go
func TestMultiMachineConformanceSuite(t *testing.T) {
    suite := NewMultiMachineConformanceSuite(t)
    
    // Phase 1: Configuration validation
    t.Run("ConfigValidation", suite.TestProductionConfigurationValid)
    
    // Phase 2: Consensus tolerance validation
    t.Run("ConsensusValidation", suite.TestScoresInConsensus)
    t.Run("ScoreBinning", suite.TestScoreBinning)
    
    // Phase 3: Operation validation
    t.Run("OpsValidation", suite.TestDeterministicOpsRegistryComplete)
    
    // Phase 4: Feature extraction consistency
    t.Run("FeatureExtraction", suite.TestFeatureVectorDeterministic)
    
    // Phase 5: Golden vector tests
    t.Run("GoldenVectors", suite.TestGoldenVectorScoresConsistent)
    
    // Phase 6: Multi-validator simulation
    t.Run("ValidatorSimulation", suite.TestSimulatedValidatorVariance)
    t.Run("ConsensusDrift", suite.TestConsensusWithDrift)
    
    suite.GenerateEvidence(t)
}
```

### Simulated Validator Tests

```go
// TestSimulatedValidatorVariance simulates multiple validators with score variance.
func (s *Suite) TestSimulatedValidatorVariance(t *testing.T) {
    baseScore := float32(75.0)
    validatorScores := []float32{
        baseScore,         // Validator 1
        baseScore + 0.3,   // Validator 2 (slight drift)
        baseScore - 0.2,   // Validator 3
        baseScore + 0.5,   // Validator 4
        baseScore - 0.4,   // Validator 5
    }

    result := ValidateConsensus(validatorScores)
    
    if !result.Achieved {
        t.Errorf("Consensus should be achieved: %s", result.Reason)
    }
}
```

## Threshold Handling

### Dynamic Tolerance Near Thresholds

Near decision thresholds (60 and 80), we use stricter tolerance:

```go
func GetEffectiveTolerance(score float32) float64 {
    // Check proximity to pass threshold (60)
    if math.Abs(float64(score-ScoreThresholdPass)) <= ThresholdBuffer {
        return ConsensusToleranceStrict
    }
    
    // Check proximity to high-trust threshold (80)
    if math.Abs(float64(score-ScoreThresholdHighTrust)) <= ThresholdBuffer {
        return ConsensusToleranceStrict
    }
    
    return ConsensusTolerance
}
```

### Threshold Zones

| Score Range | Zone | Tolerance |
|-------------|------|-----------|
| 0 - 55 | FAIL zone | ±2.0 |
| 55 - 65 | **PASS threshold zone** | **±0.5** |
| 65 - 75 | PASS zone | ±2.0 |
| 75 - 85 | **HIGH_TRUST threshold zone** | **±0.5** |
| 85 - 100 | HIGH_TRUST zone | ±2.0 |

## Conformance Testing

### Running Tests

```bash
# Run conformance suite
go test -v ./pkg/inference/... -run "TestMultiMachineConformanceSuite"

# Run with evidence generation
VEID_CONFORMANCE_EVIDENCE_DIR=./evidence go test -v ./pkg/inference/...

# Run in strict mode (CPU-only, for debugging)
VEID_STRICT_DETERMINISM=true go test -v ./pkg/inference/...
```

### CI Workflow

The `ml-determinism.yaml` workflow validates:

1. **Consensus Tolerance** - Verifies tolerance functions work correctly
2. **Score Binning** - Verifies binning produces expected results
3. **Cross-Platform Consistency** - Tests on Ubuntu, Windows, macOS
4. **Multi-Validator Simulation** - Simulates validator variance
5. **Model Hash Verification** - Verifies pinned model hashes

## Troubleshooting

### Consensus Failures

**Symptom:** Validators fail to reach consensus.

**Diagnosis:**
```go
result := ValidateConsensus(scores)
if !result.Achieved {
    log.Printf("Consensus failed: %s", result.Reason)
    log.Printf("Max deviation: %.2f (tolerance: %.2f)", 
        result.MaxDeviation, result.Tolerance)
}
```

**Solutions:**
1. Check if scores are near thresholds (stricter tolerance applies)
2. Verify all validators use the same model hash
3. Check for outlier validators with hardware issues
4. Consider increasing ConsensusTolerance if justified

### Threshold Crossing Disagreement

**Symptom:** Validators disagree on PASS/FAIL outcome.

**Diagnosis:**
```go
// Check if scores straddle a threshold
for _, score := range scores {
    if (score >= 59.5 && score <= 60.5) {
        log.Printf("Score %.2f near pass threshold - strict tolerance applies", score)
    }
}
```

**Solutions:**
1. The strict tolerance (±0.5) should prevent this
2. If occurring frequently, consider adjusting ThresholdBuffer
3. May indicate model calibration issues

## Verification Checklist

### Pre-Deployment

- [ ] `ConsensusTolerance` is 2.0
- [ ] `ConsensusToleranceStrict` is 0.5
- [ ] Model hash matches pinned registry
- [ ] Conformance tests pass locally
- [ ] CI workflow passes on all platforms
- [ ] Simulated validator tests pass

### Validator Node Setup

```bash
# 1. Verify model hash
virtengine verify-model --expected-hash "a1b2c3d4..."

# 2. Run local conformance
virtengine conformance-check

# 3. Start validator (GPU/parallel allowed)
virtengined start --veid-consensus-mode=tolerance
```

### Model Upgrade Process

1. **Governance Proposal** - Submit upgrade with new model hash
2. **Network Vote** - Validators vote on upgrade
3. **Upgrade Block** - All validators switch at specified height
4. **Hash Verification** - Each validator verifies new model hash
5. **Conformance Check** - Run conformance suite with new model

---

## References

- [TensorFlow Determinism Guide](https://www.tensorflow.org/guide/random_numbers)
- [Cosmos SDK Consensus Documentation](https://docs.cosmos.network/main/learn/intro/sdk-design)
- [Practical ML Reproducibility](https://arxiv.org/abs/2307.05523)

---

*Document maintained by VirtEngine Core Team*  
*For questions, open an issue at https://github.com/virtengine/virtengine*
