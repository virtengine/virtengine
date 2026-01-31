# VEID Identity Verification Operations

> **Duration:** 8 hours (2 days × 4 hours/day)  
> **Prerequisites:** VE-CVO certification or equivalent validator experience  
> **Target Audience:** Validator operators handling VEID verification requests

---

## Table of Contents

1. [Module Overview](#module-overview)
2. [Day 1: VEID Fundamentals](#day-1-veid-fundamentals)
3. [Day 2: Advanced Operations](#day-2-advanced-operations)
4. [Practical Exercises](#practical-exercises)
5. [Troubleshooting Guide](#troubleshooting-guide)
6. [Assessment](#assessment)

---

## Module Overview

### What is VEID?

VEID (VirtEngine Identity) is a decentralized identity verification system that combines:

- **Cryptographic Envelopes**: X25519-XSalsa20-Poly1305 encryption for data privacy
- **ML-Powered Scoring**: TensorFlow models for identity verification
- **Multi-Validator Verification**: Distributed trust through validator consensus
- **Deterministic Processing**: Reproducible results for blockchain consensus

### Validator's Role in VEID

```
┌──────────────────────────────────────────────────────────────────┐
│                    VEID Verification Pipeline                     │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│   User Device         Blockchain           Validator Nodes       │
│   ┌─────────┐        ┌─────────┐         ┌─────────────────┐    │
│   │ Capture │──────> │ Submit  │ ──────> │ 1. Receive      │    │
│   │ Identity│        │ Envelope│         │ 2. Decrypt      │    │
│   │ Data    │        │   Tx    │         │ 3. Score (ML)   │    │
│   └─────────┘        └─────────┘         │ 4. Vote         │    │
│        │                  │              │ 5. Finalize     │    │
│        │                  │              └─────────────────┘    │
│        │                  │                      │               │
│        v                  v                      v               │
│   ┌─────────┐        ┌─────────┐         ┌─────────────────┐    │
│   │ 3 Sigs: │        │ On-Chain│         │ Verification    │    │
│   │ -Client │        │ Envelope│         │ Result:         │    │
│   │ -User   │        │ Storage │         │ -Score 0.0-1.0  │    │
│   │ -Salt   │        │         │         │ -Confidence     │    │
│   └─────────┘        └─────────┘         │ -Status         │    │
│                                          └─────────────────┘    │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Learning Objectives

By the end of this module, operators will be able to:

- [ ] Configure VEID module parameters for optimal operation
- [ ] Manage validator identity keys securely
- [ ] Perform decryption operations on identity envelopes
- [ ] Understand and monitor ML model scoring
- [ ] Troubleshoot common VEID verification issues
- [ ] Execute model version upgrades safely

---

## Day 1: VEID Fundamentals

### Session 1: Identity Verification Pipeline (2 hours)

#### 1.1 Pipeline Architecture

The VEID pipeline processes identity verification requests through multiple stages:

**Stage 1: Client Capture**

```go
// Identity data captured by approved client apps
type IdentityCapture struct {
    // Facial image data (encrypted)
    FacialData []byte
    
    // Liveness detection frames
    LivenessFrames [][]byte
    
    // Document image (if applicable)
    DocumentData []byte
    
    // Device attestation
    DeviceAttestation DeviceAttestation
    
    // Capture timestamp
    CapturedAt time.Time
}
```

**Stage 2: Signature Requirements**

Every VEID submission requires three signatures for validity:

| Signature | Source | Purpose |
|-----------|--------|---------|
| Client Signature | Approved capture app | Attests capture device integrity |
| User Signature | User's wallet | Proves user consent |
| Salt Binding | Random nonce | Prevents replay attacks |

```go
// Signature verification structure
type SignatureBundle struct {
    ClientSignature struct {
        AppID       string `json:"app_id"`
        Signature   []byte `json:"signature"`
        Certificate []byte `json:"certificate"`
    }
    UserSignature struct {
        Address   string `json:"address"`
        Signature []byte `json:"signature"`
        PubKey    []byte `json:"pub_key"`
    }
    SaltBinding struct {
        Salt      []byte    `json:"salt"`
        Timestamp time.Time `json:"timestamp"`
        Signature []byte    `json:"signature"`
    }
}
```

**Stage 3: Envelope Creation**

```go
// Encryption envelope for validator processing
type EncryptionEnvelope struct {
    // Validator's key fingerprint (who can decrypt)
    RecipientFingerprint string `json:"recipient_fingerprint"`
    
    // Encryption algorithm identifier
    Algorithm string `json:"algorithm"` // "X25519-XSalsa20-Poly1305"
    
    // Encrypted identity payload
    Ciphertext []byte `json:"ciphertext"`
    
    // Encryption nonce (unique per envelope)
    Nonce []byte `json:"nonce"`
    
    // Ephemeral public key for key exchange
    EphemeralPubKey []byte `json:"ephemeral_pub_key"`
}
```

#### 1.2 Encryption Deep Dive: X25519-XSalsa20-Poly1305

VirtEngine uses NaCl's authenticated encryption for all sensitive VEID data:

**Key Exchange (X25519):**

```
User                           Validator
─────                          ─────────
Generate ephemeral keypair     Has permanent identity keypair
(ephemeral_priv, ephemeral_pub)    (validator_priv, validator_pub)
        │                              │
        └──── ephemeral_pub ──────────>│
        │                              │
        │ shared_secret = X25519(      │
        │   ephemeral_priv,            │
        │   validator_pub)             │
        │                              │
                                       │ shared_secret = X25519(
                                       │   validator_priv,
                                       │   ephemeral_pub)
                                       │
        [SAME shared_secret]           [SAME shared_secret]
```

**Symmetric Encryption (XSalsa20-Poly1305):**

```go
// Encryption process
func EncryptForValidator(data []byte, validatorPubKey [32]byte) (*EncryptionEnvelope, error) {
    // Generate ephemeral keypair
    ephemeralPub, ephemeralPriv, _ := box.GenerateKey(rand.Reader)
    
    // Generate random nonce
    var nonce [24]byte
    rand.Read(nonce[:])
    
    // Encrypt using NaCl box
    ciphertext := box.Seal(nil, data, &nonce, &validatorPubKey, ephemeralPriv)
    
    return &EncryptionEnvelope{
        Algorithm:       "X25519-XSalsa20-Poly1305",
        Ciphertext:      ciphertext,
        Nonce:           nonce[:],
        EphemeralPubKey: ephemeralPub[:],
    }, nil
}
```

#### 1.3 Validator Decryption Process

```go
// Validator decryption (requires identity private key)
func DecryptEnvelope(env *EncryptionEnvelope, validatorPrivKey [32]byte) ([]byte, error) {
    var nonce [24]byte
    var ephemeralPub [32]byte
    
    copy(nonce[:], env.Nonce)
    copy(ephemeralPub[:], env.EphemeralPubKey)
    
    // Decrypt using NaCl box.Open
    plaintext, ok := box.Open(nil, env.Ciphertext, &nonce, &ephemeralPub, &validatorPrivKey)
    if !ok {
        return nil, errors.New("decryption failed: invalid ciphertext or wrong key")
    }
    
    return plaintext, nil
}
```

#### 1.4 Hands-On Exercise: Envelope Inspection

```bash
# Query a VEID verification request
.cache/bin/virtengine query veid verification-request REQ_ID --output json

# Output structure:
# {
#   "request_id": "veid-req-12345",
#   "status": "pending",
#   "envelope": {
#     "recipient_fingerprint": "val1abc123...",
#     "algorithm": "X25519-XSalsa20-Poly1305",
#     "ciphertext": "base64...",
#     "nonce": "base64...",
#     "ephemeral_pub_key": "base64..."
#   },
#   "signatures": {
#     "client_verified": true,
#     "user_verified": true,
#     "salt_verified": true
#   },
#   "created_at": "2024-01-15T10:00:00Z"
# }
```

---

### Session 2: ML Model Scoring (2 hours)

#### 2.1 Scoring Architecture

VEID uses TensorFlow models for identity verification scoring:

```
┌─────────────────────────────────────────────────────────────────┐
│                    ML Scoring Pipeline                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────────┐    │
│  │   Facial     │   │   Liveness   │   │    Document      │    │
│  │ Verification │   │  Detection   │   │   Extraction     │    │
│  │    Model     │   │    Model     │   │     Model        │    │
│  └──────┬───────┘   └──────┬───────┘   └────────┬─────────┘    │
│         │                  │                     │              │
│         v                  v                     v              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                  Score Aggregation                        │  │
│  │  final_score = w1*facial + w2*liveness + w3*document     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│                              v                                  │
│                     ┌─────────────────┐                        │
│                     │ Threshold Check │                        │
│                     │  score >= 0.85  │                        │
│                     └─────────────────┘                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### 2.2 Determinism Requirements

**CRITICAL**: All ML inference must be deterministic for consensus. Different validators must produce identical scores for the same input.

```go
// Determinism configuration (pkg/inference/config.go)
type DeterminismConfig struct {
    // Force CPU execution (GPU introduces variance)
    ForceCPU bool `json:"force_cpu"` // MUST be true
    
    // Fixed random seed for reproducibility
    RandomSeed int64 `json:"random_seed"` // Default: 42
    
    // Enable TensorFlow deterministic operations
    DeterministicOps bool `json:"deterministic_ops"` // MUST be true
    
    // Model precision (float32 for consistency)
    Precision string `json:"precision"` // "float32"
    
    // Thread count (affects numerical stability)
    NumThreads int `json:"num_threads"` // Fixed value
}
```

**Determinism Checklist:**

- [ ] TensorFlow environment variable: `TF_DETERMINISTIC_OPS=1`
- [ ] NumPy random seed: Fixed at initialization
- [ ] CUDA disabled: `CUDA_VISIBLE_DEVICES=""`
- [ ] Single-threaded inference: Consistent thread count
- [ ] Model checksum: Verified against network consensus

#### 2.3 Model Configuration

```yaml
# ~/.virtengine/config/veid_models.yaml

models:
  facial_verification:
    path: "/opt/virtengine/models/facial_v1.2.0.pb"
    version: "1.2.0"
    checksum: "sha256:abc123..."
    input_shape: [1, 224, 224, 3]
    output_type: "float32"
    
  liveness_detection:
    path: "/opt/virtengine/models/liveness_v1.1.0.pb"
    version: "1.1.0"
    checksum: "sha256:def456..."
    input_shape: [1, 5, 112, 112, 3]  # 5 frames
    output_type: "float32"
    
  document_extraction:
    path: "/opt/virtengine/models/ocr_v1.0.0.pb"
    version: "1.0.0"
    checksum: "sha256:ghi789..."
    input_shape: [1, 512, 512, 3]
    output_type: "float32"

inference:
  force_cpu: true
  deterministic_ops: true
  random_seed: 42
  num_threads: 4
  timeout_seconds: 30
```

#### 2.4 Scoring Output Structure

```go
// Verification result from ML scoring
type VerificationScore struct {
    // Overall verification score (0.0 - 1.0)
    Score float64 `json:"score"`
    
    // Confidence level in the score
    Confidence float64 `json:"confidence"`
    
    // Individual component scores
    Components struct {
        FacialMatch  float64 `json:"facial_match"`
        Liveness     float64 `json:"liveness"`
        DocumentOCR  float64 `json:"document_ocr"`
    } `json:"components"`
    
    // Model versions used
    ModelVersions struct {
        Facial   string `json:"facial"`
        Liveness string `json:"liveness"`
        Document string `json:"document"`
    } `json:"model_versions"`
    
    // Processing metadata
    ProcessedAt   time.Time `json:"processed_at"`
    ProcessingMs  int64     `json:"processing_ms"`
    ValidatorAddr string    `json:"validator_addr"`
}
```

#### 2.5 Hands-On Exercise: Model Verification

```bash
# Check model checksums
for model in /opt/virtengine/models/*.pb; do
    echo "Model: $model"
    sha256sum "$model"
done

# Query expected checksums from chain
.cache/bin/virtengine query veid model-registry --output json | jq '.models'

# Verify determinism with test input
.cache/bin/virtengine debug veid test-inference \
    --model=facial_verification \
    --input=/opt/virtengine/test/sample_face.png \
    --iterations=5

# All iterations must produce identical scores
```

---

## Day 2: Advanced Operations

### Session 3: Identity Key Management (2 hours)

#### 3.1 Validator Identity Keys

Validators maintain a separate identity key for VEID operations:

```
~/.virtengine/config/
├── priv_validator_key.json    # Consensus signing (Ed25519)
├── node_key.json              # P2P identity (Ed25519)
└── veid_identity_key.json     # VEID decryption (X25519)
```

**Key Generation:**

```bash
# Generate new VEID identity key
.cache/bin/virtengine keys add veid-identity \
    --keyring-backend=file \
    --algo=x25519

# Export public key for registration
.cache/bin/virtengine keys show veid-identity \
    --keyring-backend=file \
    --pubkey

# Register identity key on chain
.cache/bin/virtengine tx veid register-identity-key \
    --pubkey=$(.cache/bin/virtengine keys show veid-identity --pubkey) \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto \
    --gas-prices=0.025uvirt
```

#### 3.2 Key Storage Security

| Storage Method | Security Level | Use Case |
|----------------|----------------|----------|
| File-based | Low | Development only |
| Encrypted File | Medium | Small validators |
| HSM (Hardware) | High | Production validators |
| Ledger Device | High | Operators with hardware wallets |

**HSM Integration Example (SoftHSM for testing):**

```bash
# Initialize SoftHSM slot
softhsm2-util --init-token --slot 0 --label "veid-keys" --pin 1234 --so-pin 4321

# Generate key in HSM
pkcs11-tool --module /usr/lib/softhsm/libsofthsm2.so \
    --login --pin 1234 \
    --keypairgen --key-type EC:prime256v1 \
    --id 01 --label "veid-identity"

# Configure virtengine to use HSM
cat >> ~/.virtengine/config/app.toml << EOF
[veid.hsm]
enabled = true
library_path = "/usr/lib/softhsm/libsofthsm2.so"
slot_id = 0
pin = "1234"
key_label = "veid-identity"
EOF
```

#### 3.3 Key Rotation Procedure

```bash
# Step 1: Generate new identity key
.cache/bin/virtengine keys add veid-identity-new \
    --keyring-backend=file \
    --algo=x25519

# Step 2: Register new key (keep old key active during transition)
.cache/bin/virtengine tx veid add-identity-key \
    --pubkey=$(virtengine keys show veid-identity-new --pubkey) \
    --from=operator \
    --chain-id=virtengine-1

# Step 3: Wait for confirmation (2-3 blocks)
sleep 20

# Step 4: Verify new key is active
.cache/bin/virtengine query veid validator-keys $(virtengine keys show operator --bech val -a)

# Step 5: Deprecate old key after transition period (24h recommended)
.cache/bin/virtengine tx veid remove-identity-key \
    --key-fingerprint=OLD_KEY_FINGERPRINT \
    --from=operator \
    --chain-id=virtengine-1

# Step 6: Securely delete old key material
shred -u ~/.virtengine/keyring-file/veid-identity.info
```

---

### Session 4: Model Management and Upgrades (2 hours)

#### 4.1 Model Version Governance

Model upgrades are governed on-chain to ensure all validators use identical versions:

```bash
# Query current model registry
.cache/bin/virtengine query veid model-registry --output json

# Output:
# {
#   "models": [
#     {
#       "name": "facial_verification",
#       "version": "1.2.0",
#       "checksum": "sha256:abc123...",
#       "effective_height": 5000000,
#       "status": "active"
#     },
#     {
#       "name": "liveness_detection",
#       "version": "1.1.0",
#       "checksum": "sha256:def456...",
#       "effective_height": 4500000,
#       "status": "active"
#     }
#   ]
# }
```

#### 4.2 Model Upgrade Process

**Phase 1: Governance Proposal**

```bash
# Submit model upgrade proposal
.cache/bin/virtengine tx gov submit-proposal upgrade-veid-model \
    --model-name="facial_verification" \
    --new-version="1.3.0" \
    --checksum="sha256:newchecksum..." \
    --download-url="https://models.virtengine.io/facial_v1.3.0.pb" \
    --effective-height=5500000 \
    --title="Upgrade Facial Verification Model to v1.3.0" \
    --description="Improved accuracy and liveness detection" \
    --deposit=1000000uvirt \
    --from=operator
```

**Phase 2: Validator Preparation**

```bash
# Download new model before effective height
wget -O /opt/virtengine/models/facial_v1.3.0.pb \
    "https://models.virtengine.io/facial_v1.3.0.pb"

# Verify checksum
echo "sha256:newchecksum... /opt/virtengine/models/facial_v1.3.0.pb" | \
    sha256sum -c -

# Update model configuration
cat >> ~/.virtengine/config/veid_models.yaml << EOF
  facial_verification_v1.3.0:
    path: "/opt/virtengine/models/facial_v1.3.0.pb"
    version: "1.3.0"
    checksum: "sha256:newchecksum..."
    effective_height: 5500000
EOF

# Restart node to load new model
sudo systemctl restart virtengine
```

**Phase 3: Activation**

At the effective height, the node automatically switches to the new model.

```bash
# Monitor activation
.cache/bin/virtengine query veid model-registry --output json | \
    jq '.models[] | select(.name=="facial_verification")'

# Verify inference with new model
.cache/bin/virtengine debug veid test-inference \
    --model=facial_verification \
    --input=/opt/virtengine/test/sample_face.png
```

#### 4.3 Model Pinning

For testing or rollback scenarios, operators can pin specific model versions:

```toml
# ~/.virtengine/config/app.toml

[veid.model_pinning]
# Enable model pinning (use with caution)
enabled = false

# Pinned versions (overrides network consensus)
# WARNING: Using different versions will cause consensus failures
[veid.model_pinning.versions]
facial_verification = "1.2.0"
liveness_detection = "1.1.0"
document_extraction = "1.0.0"
```

**⚠️ Warning**: Model pinning should only be used for:
- Testing new models before network adoption
- Emergency rollback during model failures
- Development and debugging

---

## Practical Exercises

### Exercise 1: VEID Request Processing Simulation

**Objective**: Understand the complete VEID verification flow

```bash
# Step 1: Create test identity data
cat > /tmp/test_identity.json << EOF
{
  "facial_data": "$(base64 /opt/virtengine/test/sample_face.png)",
  "liveness_frames": [
    "$(base64 /opt/virtengine/test/frame1.png)",
    "$(base64 /opt/virtengine/test/frame2.png)",
    "$(base64 /opt/virtengine/test/frame3.png)"
  ],
  "document_data": "$(base64 /opt/virtengine/test/sample_id.png)"
}
EOF

# Step 2: Encrypt for validator
.cache/bin/virtengine debug veid encrypt-for-validator \
    --input=/tmp/test_identity.json \
    --validator=$(virtengine keys show operator --bech val -a) \
    --output=/tmp/test_envelope.json

# Step 3: Submit verification request (testnet only)
.cache/bin/virtengine tx veid submit-verification \
    --envelope=/tmp/test_envelope.json \
    --from=test-user \
    --chain-id=virtengine-testnet-1

# Step 4: Query request status
.cache/bin/virtengine query veid verification-request REQ_ID

# Step 5: View processing result
.cache/bin/virtengine query veid verification-result REQ_ID
```

### Exercise 2: Decryption Operation

**Objective**: Perform manual envelope decryption

```bash
# Step 1: Fetch pending verification for your validator
.cache/bin/virtengine query veid pending-verifications \
    --validator=$(virtengine keys show operator --bech val -a) \
    --limit=1

# Step 2: Decrypt envelope manually (debugging only)
.cache/bin/virtengine debug veid decrypt-envelope \
    --envelope=/tmp/pending_envelope.json \
    --identity-key=veid-identity \
    --keyring-backend=file \
    --output=/tmp/decrypted_identity.json

# Step 3: Verify decrypted data structure
cat /tmp/decrypted_identity.json | jq '.facial_data | length'

# Step 4: Clean up sensitive data
shred -u /tmp/decrypted_identity.json
```

### Exercise 3: ML Model Scoring Test

**Objective**: Verify deterministic scoring

```bash
# Step 1: Run inference multiple times
for i in {1..5}; do
    .cache/bin/virtengine debug veid test-inference \
        --model=facial_verification \
        --input=/opt/virtengine/test/sample_face.png \
        --output=/tmp/score_$i.json
done

# Step 2: Compare all scores (must be identical)
for i in {1..5}; do
    jq '.score' /tmp/score_$i.json
done | sort | uniq -c

# Expected output: "5 0.9234567" (single unique score, 5 occurrences)

# Step 3: Test with different model
.cache/bin/virtengine debug veid test-inference \
    --model=liveness_detection \
    --input=/opt/virtengine/test/liveness_frames/ \
    --output=/tmp/liveness_score.json

# Step 4: Validate output structure
cat /tmp/liveness_score.json | jq '.'
```

### Exercise 4: Key Rotation Practice

**Objective**: Safely rotate VEID identity keys

```bash
# Step 1: Backup current key
cp ~/.virtengine/keyring-file/veid-identity.info \
    ~/.virtengine/keyring-backup/veid-identity-$(date +%Y%m%d).info

# Step 2: Generate new key
.cache/bin/virtengine keys add veid-identity-rotation \
    --keyring-backend=file \
    --algo=x25519

# Step 3: Register new key (testnet)
.cache/bin/virtengine tx veid add-identity-key \
    --pubkey=$(virtengine keys show veid-identity-rotation --pubkey) \
    --from=operator \
    --chain-id=virtengine-testnet-1

# Step 4: Verify both keys are active
.cache/bin/virtengine query veid validator-keys \
    $(virtengine keys show operator --bech val -a) --output json | jq '.keys'

# Step 5: Process a verification with new key
# (Wait for new verifications to come in)

# Step 6: Deprecate old key
.cache/bin/virtengine tx veid remove-identity-key \
    --key-fingerprint=OLD_FINGERPRINT \
    --from=operator \
    --chain-id=virtengine-testnet-1
```

---

## Troubleshooting Guide

### Common Issues and Solutions

#### Issue 1: Decryption Failure

**Symptoms:**
- Error: `decryption failed: invalid ciphertext or wrong key`
- VEID verifications stuck in "pending" state

**Diagnosis:**
```bash
# Check registered identity keys
.cache/bin/virtengine query veid validator-keys $(virtengine keys show operator --bech val -a)

# Verify envelope recipient fingerprint matches your key
.cache/bin/virtengine query veid verification-request REQ_ID | jq '.envelope.recipient_fingerprint'

# Check local key fingerprint
.cache/bin/virtengine keys show veid-identity --keyring-backend=file | grep fingerprint
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Key not registered | Register identity key on chain |
| Wrong key configured | Update veid_identity_key.json |
| Key rotated during request | Wait for new requests with current key |
| Corrupted envelope | Request resubmission from user |

---

#### Issue 2: ML Model Scoring Variance

**Symptoms:**
- Consensus failures on VEID transactions
- Different validators reporting different scores
- Error: `verification score mismatch`

**Diagnosis:**
```bash
# Verify model checksums
sha256sum /opt/virtengine/models/*.pb

# Query expected checksums
.cache/bin/virtengine query veid model-registry --output json | \
    jq '.models[] | {name, checksum}'

# Check determinism settings
grep -A5 "inference:" ~/.virtengine/config/veid_models.yaml

# Test reproducibility
for i in {1..3}; do
    .cache/bin/virtengine debug veid test-inference \
        --model=facial_verification \
        --input=/opt/virtengine/test/sample_face.png
done
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Model checksum mismatch | Re-download official model |
| GPU enabled | Set `force_cpu: true` |
| Non-deterministic ops | Set `TF_DETERMINISTIC_OPS=1` |
| Thread count variance | Fix `num_threads` in config |
| Wrong model version | Update to network-consensus version |

---

#### Issue 3: Verification Timeout

**Symptoms:**
- VEID requests timing out
- Error: `verification processing timeout exceeded`
- Slow block production during VEID processing

**Diagnosis:**
```bash
# Check processing queue
.cache/bin/virtengine query veid pending-verifications --count-only

# Monitor processing time
journalctl -u virtengine | grep -E "veid.*processing.*ms"

# Check system resources
top -b -n1 | head -20
iostat -x 1 5
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| High queue backlog | Increase processing threads |
| Slow disk I/O | Use NVMe SSD, optimize LevelDB |
| CPU bottleneck | Upgrade CPU or optimize config |
| Large identity data | Check max payload size settings |
| Model loading slow | Pre-load models at startup |

---

#### Issue 4: Signature Verification Failure

**Symptoms:**
- Error: `client signature verification failed`
- Error: `user signature verification failed`
- Error: `salt binding verification failed`

**Diagnosis:**
```bash
# Check signature bundle
.cache/bin/virtengine query veid verification-request REQ_ID | \
    jq '.signatures'

# Verify client certificate
.cache/bin/virtengine query veid approved-clients | jq '.clients'

# Check timestamp validity
.cache/bin/virtengine query veid verification-request REQ_ID | \
    jq '.signatures.salt_binding.timestamp'
```

**Solutions:**

| Cause | Solution |
|-------|----------|
| Expired client certificate | User must use updated app |
| Revoked client | Check approved clients list |
| Timestamp too old | User must resubmit within window |
| Replay attack detected | Salt already used, reject request |
| Wrong user pubkey | User must sign with registered key |

---

### Emergency Procedures

#### Emergency: Identity Key Compromise

```bash
# IMMEDIATE ACTIONS:

# 1. Stop validator to prevent further decryptions
sudo systemctl stop virtengine

# 2. Revoke compromised key on chain
.cache/bin/virtengine tx veid emergency-revoke-key \
    --key-fingerprint=COMPROMISED_KEY_FINGERPRINT \
    --from=operator \
    --chain-id=virtengine-1 \
    --gas=auto \
    --gas-prices=0.025uvirt

# 3. Generate new identity key
.cache/bin/virtengine keys add veid-identity-emergency \
    --keyring-backend=file \
    --algo=x25519

# 4. Register new key
.cache/bin/virtengine tx veid register-identity-key \
    --pubkey=$(virtengine keys show veid-identity-emergency --pubkey) \
    --from=operator

# 5. Securely destroy old key material
shred -u ~/.virtengine/keyring-file/veid-identity.info

# 6. Restart validator with new key
sudo systemctl start virtengine

# 7. Notify security team
# security@virtengine.io

# 8. Review all recent decryptions for the compromised key
```

#### Emergency: Model Corruption

```bash
# IMMEDIATE ACTIONS:

# 1. Stop VEID processing
.cache/bin/virtengine tx veid pause-processing \
    --from=operator \
    --chain-id=virtengine-1

# 2. Verify corruption
sha256sum /opt/virtengine/models/*.pb
.cache/bin/virtengine query veid model-registry

# 3. Re-download official models
MODEL_URL="https://models.virtengine.io"
for model in facial_v1.2.0 liveness_v1.1.0 document_v1.0.0; do
    wget -O /opt/virtengine/models/${model}.pb "$MODEL_URL/${model}.pb"
done

# 4. Verify new checksums
sha256sum /opt/virtengine/models/*.pb

# 5. Resume processing
.cache/bin/virtengine tx veid resume-processing \
    --from=operator \
    --chain-id=virtengine-1
```

---

## Assessment

### Knowledge Check Questions

1. What encryption algorithm does VEID use for identity data?
2. Why must ML inference be deterministic in VEID?
3. What three signatures are required for a valid VEID submission?
4. How does the model upgrade governance process work?
5. What is the procedure for identity key rotation?

### Practical Assessment

Complete the following tasks:

- [ ] Configure a validator for VEID operations
- [ ] Generate and register an identity key
- [ ] Verify ML model determinism across multiple runs
- [ ] Successfully process a test VEID verification
- [ ] Perform a simulated key rotation
- [ ] Diagnose and resolve a provided misconfiguration

### Certification Requirements

| Requirement | Minimum |
|-------------|---------|
| Knowledge Check | 80% correct |
| Practical Tasks | All completed |
| Troubleshooting Scenarios | 3/4 resolved |

---

## Additional Resources

- [VEID Technical Specification](./../veid-flow-spec.md)
- [VEID ML Feature Schema](./../veid-ml-feature-schema.md)
- [Encryption Key Management](./../key-management.md)
- [Security Guidelines](./../security-guidelines.md)
- [VEID ZK Proofs Security](./../veid-zkproofs-security.md)

---

*Document Version: 1.0.0*  
*Last Updated: 2024-01-15*  
*Maintainer: VirtEngine VEID Operations Team*