# Runbook: VEID Non-Deterministic Inference

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | VEIDInferenceNonDeterministic |
| Severity | Critical |
| Service | virtengine-veid |
| Tier | Tier 0 |
| SLO Impact | Chain consensus integrity |

## Summary

This alert fires when the VEID (VirtEngine Identity) ML inference system produces non-deterministic results. This is a **critical consensus issue** because all validators must produce identical inference results for the same input.

## Impact

- **Critical**: Consensus failure if validators disagree on verification results
- **Critical**: Chain could halt due to app hash mismatch
- **High**: Identity verifications may be rejected
- **High**: Legal/compliance implications for identity verification

## Root Cause Categories

1. **Hardware variance**: GPU vs CPU, different hardware producing different results
2. **Software variance**: Different TensorFlow/model versions across nodes
3. **Floating point variance**: Non-deterministic floating point operations
4. **Random state**: Unseeded random number generators
5. **Concurrency**: Race conditions in inference pipeline

## Prerequisites

- Access to ML inference logs
- Understanding of determinism requirements
- Access to validator nodes for comparison

## Diagnostic Steps

### 1. Identify the Scope

```bash
# Check recent inference results on multiple nodes
for node in validator1 validator2 validator3; do
  echo "=== $node ==="
  ssh user@$node "grep 'inference_result' /var/log/virtengine/veid.log | tail -5"
done

# Compare hashes
curl -s http://node1:26657/abci_query?path="/veid/inference/$REQUEST_ID" | jq '.result.value' | base64 -d
curl -s http://node2:26657/abci_query?path="/veid/inference/$REQUEST_ID" | jq '.result.value' | base64 -d
```

### 2. Check Determinism Configuration

```bash
# Verify determinism settings in config
grep -A 10 "\[inference\]" /etc/virtengine/config/app.toml

# Expected settings:
# force_cpu = true
# random_seed = 42
# deterministic_ops = true
# tensorflow_determinism = "1"

# Check environment variables
env | grep -E "(TF_|PYTHONHASHSEED|CUDA)"
```

### 3. Check Model Versions

```bash
# Get model checksums on each node
for node in validator1 validator2 validator3; do
  echo "=== $node ==="
  ssh user@$node "sha256sum /var/lib/virtengine/ml/models/*"
done

# Compare versions
ssh validator1 "virtengined q veid model-info"
```

### 4. Analyze Inference Logs

```bash
# Check for variance warnings
grep -i "variance\|determinism\|mismatch" /var/log/virtengine/inference.log

# Get inference timing
grep "inference_duration" /var/log/virtengine/inference.log | tail -20

# Check for GPU usage (should be disabled)
nvidia-smi 2>/dev/null && echo "WARNING: GPU detected"
```

### 5. Compare Specific Results

```bash
# Run determinism check tool
virtengined debug inference-check \
  --input-file /tmp/test_input.json \
  --expected-hash "abc123..." \
  --verbose
```

## Resolution Steps

### Scenario 1: Force CPU Mode Not Enabled

```bash
# 1. Update app.toml on affected node
cat >> /etc/virtengine/config/app.toml << EOF

[inference]
force_cpu = true
random_seed = 42
deterministic_ops = true
EOF

# 2. Set environment variables
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export PYTHONHASHSEED=42

# 3. Update systemd service
sudo systemctl edit virtengined
# Add to [Service]:
# Environment="TF_DETERMINISTIC_OPS=1"
# Environment="PYTHONHASHSEED=42"

# 4. Restart service
sudo systemctl daemon-reload
sudo systemctl restart virtengined
```

### Scenario 2: Model Version Mismatch

```bash
# 1. Get canonical model from trusted source
CANONICAL_MODEL_HASH="sha256:abc123..."
curl -O https://models.virtengine.io/veid/v1.2.3/model.pb

# 2. Verify checksum
sha256sum model.pb | grep -q "$CANONICAL_MODEL_HASH" || echo "CHECKSUM MISMATCH"

# 3. Deploy to affected nodes
for node in $AFFECTED_NODES; do
  scp model.pb user@$node:/var/lib/virtengine/ml/models/
  ssh user@$node "sudo systemctl restart virtengined"
done

# 4. Verify consistency
virtengined q veid model-info
```

### Scenario 3: Library Version Mismatch

```bash
# 1. Check Python/TensorFlow versions
python3 -c "import tensorflow as tf; print(tf.__version__)"

# 2. Compare with expected
# Expected: tensorflow==2.13.0 (pinned version)

# 3. Reinstall from requirements-deterministic.txt
pip install -r /opt/virtengine/ml/requirements-deterministic.txt --force-reinstall

# 4. Verify numpy/scipy versions (affect floating point)
pip freeze | grep -E "numpy|scipy"
```

### Scenario 4: Emergency Bypass (Last Resort)

If consensus is failing and cannot be fixed quickly:

```bash
# 1. Enable inference bypass mode (governance required)
virtengined tx gov submit-proposal param-change \
  --param "veid/InferenceBypass=true" \
  --title "Emergency: Disable VEID inference" \
  --description "Temporary bypass due to non-determinism issue"

# 2. This requires governance vote - coordinate with validators

# 3. Once passed, inference returns deterministic placeholder
# WARNING: This disables identity verification temporarily
```

## Recovery Verification

```bash
# 1. Run determinism test suite
./scripts/test-determinism.sh

# 2. Compare inference across nodes
REQUEST_ID=$(uuidgen)
for node in validator1 validator2 validator3; do
  echo "=== $node ==="
  ssh user@$node "virtengined debug inference-test --request-id $REQUEST_ID" | sha256sum
done
# All hashes should match

# 3. Submit test verification
virtengined tx veid verify-test --from test-account --yes

# 4. Monitor for 10 minutes
watch -n 30 'grep -c "inference_mismatch" /var/log/virtengine/veid.log'
```

## Prevention

### Configuration Checklist

All validator nodes MUST have:

```toml
# app.toml
[inference]
force_cpu = true
random_seed = 42
deterministic_ops = true
inter_op_parallelism = 1
intra_op_parallelism = 1
```

```bash
# Environment
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export PYTHONHASHSEED=42
export OMP_NUM_THREADS=1
```

### Model Update Procedure

1. New models require governance proposal
2. All validators must update simultaneously
3. Verify checksum before deployment
4. Test determinism before activation

## Escalation

**Immediate escalation to ML Team if**:
- Root cause unknown after 15 minutes
- Multiple nodes affected
- Model corruption suspected

**Escalate to Security if**:
- Potential tampering detected
- Unauthorized model changes
- Checksum mismatches unexplained

## Post-Incident

1. **Mandatory postmortem** - this is a consensus integrity issue
2. Review and audit:
   - All node configurations
   - Model deployment process
   - Determinism test coverage
3. Consider:
   - Automated configuration validation
   - Pre-block determinism checks
   - Model signing requirements

## Related Alerts

- `VEIDVerificationLatencyHigh` - May indicate issues
- `InferenceScorerError` - Scorer failures
- `AppHashMismatch` - Consensus divergence
- `ValidatorJailed` - Consequence of divergence

## References

- [VEID Architecture](../../../_docs/VEID_Architecture.md)
- [ML Determinism Guide](../../../ml/docs/determinism.md)
- [TensorFlow Determinism](https://www.tensorflow.org/api_docs/python/tf/config/experimental/enable_op_determinism)
- [Inference Configuration](../../../pkg/inference/README.md)
