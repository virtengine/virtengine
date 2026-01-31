# VEID Determinism Conformance Suite

This document describes the deterministic inference conformance suite for the VirtEngine Identity (VEID) module. The suite verifies that ML inference produces identical outputs across all validator nodes for blockchain consensus.

## Overview

The VEID module uses ML-based identity scoring that must be **deterministic** across all validators. Any variance in output would cause consensus failures and potential chain halts.

### Why Determinism Matters

1. **Consensus Integrity**: All validators must compute identical scores for the same input
2. **Chain Consistency**: Different scores would cause app hash mismatches
3. **Legal Compliance**: Identity verification results must be reproducible
4. **Audit Trail**: Evidence must be verifiable across time and platforms

## Conformance Test Suite

### Location

- **Test Files**: `pkg/inference/deterministic_conformance_suite_test.go`
- **Golden Vectors**: `pkg/inference/golden_vectors.go`
- **CI Workflow**: `.github/workflows/veid-conformance.yml`

### Running Tests Locally

```bash
# Run full conformance suite
go test -v -run "TestDeterministicConformanceSuite" ./pkg/inference/...

# Run specific tests
go test -v -run "TestGoldenVectorExecution" ./pkg/inference/...
go test -v -run "TestMultiEnvironmentSimulation" ./pkg/inference/...

# Run with race detection
go test -v -race -run "TestDeterministicConformanceSuite" ./pkg/inference/...

# Run benchmarks
go test -v -bench=. -benchmem ./pkg/inference/...
```

### Environment Variables

```bash
# Required for determinism
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export PYTHONHASHSEED=42
export OMP_NUM_THREADS=1
export TF_ENABLE_ONEDNN_OPTS=0

# Optional: Evidence storage location
export VEID_CONFORMANCE_EVIDENCE_DIR=/path/to/evidence
```

## Golden Test Vectors

The suite includes **5 golden vectors** with pinned expected outputs:

| Vector ID | Description | Expected Score |
|-----------|-------------|----------------|
| `golden_high_quality_v1` | High-quality verification (canonical) | 87 |
| `golden_medium_quality_v1` | Medium quality inputs | 62 |
| `golden_low_quality_v1` | Low document quality | 38 |
| `golden_boundary_v1` | Boundary values | 50 |
| `golden_perfect_v1` | Maximum quality | 95 |

### Vector Format

```go
type GoldenVector struct {
    ID                   string        // Unique identifier
    Name                 string        // Human-readable name
    Version              string        // Vector format version
    Inputs               *ScoreInputs  // Input data
    ExpectedInputHash    string        // SHA256 of serialized inputs
    ExpectedOutputHash   string        // SHA256 of raw model output
    ExpectedScore        uint32        // Expected quantized score (0-100)
    RequiredModelVersion string        // Model version
    RequiredModelHash    string        // Model weights hash
}
```

## Allowed Hardware/Software Matrix

### Supported Platforms

| Platform | Architecture | Support Level | Notes |
|----------|-------------|---------------|-------|
| Linux | amd64 | **Primary** | Reference platform for validators |
| Linux | arm64 | **Primary** | ARM-based validators |
| macOS | arm64 | Secondary | Development (Apple Silicon) |
| macOS | amd64 | Secondary | Development (Intel) |
| Windows | amd64 | Secondary | Development only |

### Required Software Versions

| Component | Version | Notes |
|-----------|---------|-------|
| Go | 1.22+ | Required for builds |
| TensorFlow | 2.13.0 | Pinned for determinism |
| NumPy | 1.24.3 | Pinned for floating point consistency |
| Python | 3.10+ | For ML pipeline |

### Runtime Requirements

| Setting | Value | Reason |
|---------|-------|--------|
| `TF_DETERMINISTIC_OPS` | `1` | Force deterministic TF operations |
| `TF_CUDNN_DETERMINISTIC` | `1` | Disable cuDNN non-determinism |
| `PYTHONHASHSEED` | `42` | Fixed hash seed |
| `OMP_NUM_THREADS` | `1` | Single-threaded execution |
| Random Seed | `42` | Fixed for all RNG |
| CPU Only | `true` | No GPU (variance source) |
| Inter-op Parallelism | `1` | Single thread |
| Intra-op Parallelism | `1` | Single thread |

## Test Evidence

### Storage

Test evidence is stored in JSON format with the following structure:

```json
{
  "suite_version": "1.0.0",
  "platform": {
    "os": "linux",
    "arch": "amd64",
    "go_version": "go1.22.0"
  },
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-01T00:01:30Z",
  "results": [
    {
      "vector_id": "golden_high_quality_v1",
      "passed": true,
      "actual_input_hash": "d4f5e6a7b8c9...",
      "actual_output_hash": "a1b2c3d4e5f6...",
      "execution_time_ms": 150
    }
  ],
  "total_passed": 5,
  "total_failed": 0
}
```

### CI Artifacts

The CI workflow stores evidence as artifacts:

- **Retention**: 90 days
- **Naming**: `conformance-evidence-{os}-{arch}`
- **Contents**:
  - `conformance-output.txt` - Full test output
  - `platform-info.txt` - Platform details
  - `*.json` - Structured evidence files

### Accessing Evidence

```bash
# Download from GitHub Actions
gh run download <run-id> -n conformance-evidence-linux-amd64

# View evidence
cat conformance-evidence-linux-amd64/conformance-output.txt
```

## CI Integration

### Workflow Triggers

- **Push**: To `main` or `mainnet/main` branches
- **Pull Request**: To `main` or `mainnet/main` branches
- **Schedule**: Daily at 2 AM UTC
- **Manual**: Via workflow dispatch

### Required Checks

For PRs affecting `pkg/inference/`, `x/veid/`, or `ml/`:

1. `Conformance (linux-amd64)` - Must pass
2. `Conformance (linux-arm64)` - Must pass
3. `Verify Cross-Platform Hashes` - Must pass

### Failure Response

If conformance tests fail:

1. **Do not merge** the PR
2. Check evidence artifacts for differences
3. Review `docs/operations/runbooks/veid-non-deterministic.md`
4. Fix the non-determinism source
5. Re-run the conformance suite

## Adding New Test Vectors

### Process

1. Create new vector in `pkg/inference/golden_vectors.go`
2. Compute expected hashes on reference platform (linux/amd64)
3. Run conformance suite locally
4. Submit PR and verify CI passes

### Example

```go
{
    ID:          "golden_new_case_v1",
    Name:        "New Test Case - Golden",
    Description: "Description of what this tests",
    Version:     GoldenVectorVersion,
    Inputs: &ScoreInputs{
        // ... input data
    },
    ExpectedInputHash:    "compute_on_reference_platform",
    ExpectedOutputHash:   "compute_on_reference_platform",
    ExpectedScore:        75,
    RequiredModelVersion: "v1.0.0",
    RequiredModelHash:    ExpectedModelHashV1,
},
```

## Troubleshooting

### Common Issues

#### Hash Mismatch Between Platforms

1. Check TensorFlow version matches
2. Verify environment variables are set
3. Ensure CPU-only mode is enabled
4. Check for floating point precision issues

#### Test Timeout

1. Increase timeout in test flags
2. Check for resource constraints
3. Verify model loads correctly

#### Missing Evidence Files

1. Check evidence directory permissions
2. Verify `VEID_CONFORMANCE_EVIDENCE_DIR` is set
3. Check disk space

### Debug Mode

```bash
# Enable verbose logging
go test -v -run "TestDeterministicConformanceSuite" \
  -args -debug=true \
  ./pkg/inference/...
```

## References

- [VE-219: Deterministic identity verification runtime](../../../_docs/veid-flow-spec.md)
- [ML Determinism Guide](../../../ml/docs/determinism.md)
- [TensorFlow Determinism](https://www.tensorflow.org/api_docs/python/tf/config/experimental/enable_op_determinism)
- [Non-Deterministic Runbook](../../operations/runbooks/veid-non-deterministic.md)

---

*Task 8A: Deterministic inference conformance suite*
*VE-219: Deterministic identity verification runtime*
