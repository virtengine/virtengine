# Security Fundamentals

**Module Duration:** 4 hours  
**Level:** Foundational  
**Prerequisites:** Architecture Overview module, basic cryptography concepts  
**Version:** 1.0.0  
**Last Updated:** 2025-01-24

---

## Table of Contents

1. [Learning Objectives](#learning-objectives)
2. [Introduction to VirtEngine Security](#introduction-to-virtengine-security)
3. [Cryptographic Foundations](#cryptographic-foundations)
4. [Key Management Basics](#key-management-basics)
5. [Encryption in VirtEngine](#encryption-in-virtengine)
6. [Security Principles](#security-principles)
7. [Attack Vectors and Threats](#attack-vectors-and-threats)
8. [Defense Strategies](#defense-strategies)
9. [Key Takeaways](#key-takeaways)
10. [Hands-On Exercises](#hands-on-exercises)
11. [Assessment Questions](#assessment-questions)
12. [References](#references)

---

## Learning Objectives

By the end of this module, you will be able to:

- [ ] **Explain** VirtEngine's cryptographic architecture and approved algorithms
- [ ] **Understand** key management concepts including HSM integration and key rotation
- [ ] **Describe** the X25519-XSalsa20-Poly1305 encryption envelope format
- [ ] **Identify** common attack vectors and their mitigations
- [ ] **Apply** security principles when reviewing code or configurations
- [ ] **Recognize** secure coding patterns and anti-patterns
- [ ] **Implement** basic security controls in VirtEngine contexts

---

## Introduction to VirtEngine Security

### Why Security Matters

VirtEngine handles multiple high-value assets requiring robust security:

| Asset Type | Risk | Impact |
|------------|------|--------|
| **Financial** | Token theft, escrow manipulation | Direct monetary loss |
| **Identity** | VEID data exposure, identity fraud | Privacy breach, regulatory penalties |
| **Infrastructure** | Unauthorized resource access | Resource theft, data exfiltration |
| **Consensus** | Validator compromise, chain halt | Network-wide disruption |

### Security Context

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    VIRTENGINE SECURITY LAYERS                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  APPLICATION LAYER                                               │   │
│  │  - Input validation, authorization, MFA gating                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CRYPTOGRAPHIC LAYER                                             │   │
│  │  - Encryption (X25519-XSalsa20-Poly1305), signing (Ed25519)     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CONSENSUS LAYER                                                 │   │
│  │  - BFT consensus, validator staking, slashing                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  INFRASTRUCTURE LAYER                                            │   │
│  │  - TLS, network isolation, HSM, access controls                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Core Security Principles

VirtEngine follows these foundational principles:

| Principle | Description | VirtEngine Implementation |
|-----------|-------------|---------------------------|
| **Defense in Depth** | Multiple security layers | Encryption + MFA + RBAC + audit |
| **Least Privilege** | Minimal required permissions | Role-based access control |
| **Fail Secure** | Deny by default on errors | Transactions fail closed |
| **Zero Trust** | Verify all inputs | Validate even from trusted sources |
| **Determinism** | Reproducible operations | Required for consensus safety |

---

## Cryptographic Foundations

### Approved Algorithms

VirtEngine uses **only** the following cryptographic primitives:

| Purpose | Algorithm | Library | Security Level |
|---------|-----------|---------|----------------|
| Key Exchange | X25519 (Curve25519) | `golang.org/x/crypto/curve25519` | 128-bit |
| Symmetric Encryption | XSalsa20-Poly1305 | `golang.org/x/crypto/nacl/secretbox` | 256-bit key |
| AEAD Encryption | AES-256-GCM | `crypto/aes` + `crypto/cipher` | 256-bit key |
| Hashing | SHA-256 | `crypto/sha256` | 256-bit |
| Digital Signatures | Ed25519 | Cosmos SDK keyring | 128-bit |
| Blockchain Signatures | secp256k1 | Cosmos SDK keyring | 128-bit |

### Why These Algorithms?

**X25519 (Key Exchange)**
- Elliptic Curve Diffie-Hellman on Curve25519
- High security with small key sizes (32 bytes)
- Constant-time implementation (timing attack resistant)
- Widely audited and battle-tested

**XSalsa20-Poly1305 (Authenticated Encryption)**
- Stream cipher (XSalsa20) + MAC (Poly1305)
- Authenticated encryption (confidentiality + integrity)
- Nonce-based (24-byte nonce prevents reuse issues)
- Part of NaCl cryptographic library

**Ed25519 (Digital Signatures)**
- EdDSA signature scheme on Curve25519
- Fast signature generation and verification
- Small signatures (64 bytes)
- Deterministic signatures (no random nonce issues)

### Cryptographic Anti-Patterns

**❌ NEVER do these:**

```go
// ❌ NEVER: Custom crypto implementations
func myEncrypt(data []byte, key []byte) []byte {
    result := make([]byte, len(data))
    for i, b := range data {
        result[i] = b ^ key[i%len(key)]  // XOR is NOT encryption!
    }
    return result
}

// ❌ NEVER: Weak random generation
func myRandom() int64 {
    return time.Now().UnixNano()  // Predictable!
}

// ❌ NEVER: MD5 or SHA1 for security purposes
hash := md5.Sum(password)  // Broken!

// ❌ NEVER: ECB mode encryption
block, _ := aes.NewCipher(key)
block.Encrypt(dst, src)  // No IV, patterns visible!
```

**✅ CORRECT patterns:**

```go
// ✅ Use crypto/rand for random bytes
import "crypto/rand"

nonce := make([]byte, 24)
if _, err := rand.Read(nonce); err != nil {
    return fmt.Errorf("failed to generate nonce: %w", err)
}

// ✅ Use approved libraries
import "golang.org/x/crypto/nacl/box"

ciphertext := box.Seal(nil, plaintext, &nonce, &recipientPubKey, &senderPrivKey)

// ✅ Use SHA-256 for hashing
hash := sha256.Sum256(data)
```

### Hashing Best Practices

```go
// Computing deterministic hashes for consensus
func ComputeOrderHash(order *Order) []byte {
    h := sha256.New()
    // Use canonical (protobuf) encoding - NOT JSON!
    bz, _ := order.Marshal()
    h.Write(bz)
    return h.Sum(nil)
}

// ❌ WRONG: JSON encoding is NOT deterministic
func WrongHash(order *Order) []byte {
    bz, _ := json.Marshal(order)  // Map key order varies!
    return sha256.Sum256(bz)
}
```

---

## Key Management Basics

### Key Types in VirtEngine

| Key Type | Algorithm | Purpose | Storage |
|----------|-----------|---------|---------|
| **Validator Consensus Key** | Ed25519 | Block signing | HSM/Ledger (production) |
| **Validator Operator Key** | secp256k1 | Transactions | Keyring |
| **Provider Key** | secp256k1 | Bid signing, usage records | Keyring/HSM |
| **User Key** | secp256k1 | Transaction signing | Wallet |
| **Encryption Key** | X25519 | VEID data encryption | Per-validator |
| **Ephemeral Key** | X25519 | Per-envelope encryption | Temporary |

### Key Storage Options

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    KEY STORAGE HIERARCHY                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐                                                        │
│  │    HSM      │  Hardware Security Module                              │
│  │  (Highest)  │  - Tamper-resistant hardware                           │
│  └──────┬──────┘  - Keys never leave device                             │
│         │         - PKCS#11 interface                                    │
│         ▼                                                                │
│  ┌─────────────┐                                                        │
│  │   Ledger    │  Hardware Wallet                                       │
│  │   Device    │  - Non-custodial                                       │
│  └──────┬──────┘  - User confirmation required                          │
│         │                                                                │
│         ▼                                                                │
│  ┌─────────────┐                                                        │
│  │  OS Keyring │  Operating System Keychain                             │
│  │             │  - macOS Keychain, Windows DPAPI                       │
│  └──────┬──────┘  - Protected by OS security                            │
│         │                                                                │
│         ▼                                                                │
│  ┌─────────────┐                                                        │
│  │   File      │  Encrypted File                                        │
│  │ (Encrypted) │  - Password-protected                                  │
│  └──────┬──────┘  - Development use only                                │
│         │                                                                │
│         ▼                                                                │
│  ┌─────────────┐                                                        │
│  │   Memory    │  In-Memory (Test Only)                                 │
│  │  (Lowest)   │  - No persistence                                      │
│  └─────────────┘  - NEVER for production                                │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### HSM Integration

**Supported HSMs:**

| HSM | Type | Production Ready |
|-----|------|------------------|
| YubiHSM 2 | Hardware | ✅ Yes |
| AWS CloudHSM | Cloud | ✅ Yes |
| Azure Dedicated HSM | Cloud | ✅ Yes |
| GCP Cloud HSM | Cloud | ✅ Yes |
| Thales Luna | Enterprise | ✅ Yes |
| SoftHSM | Software | ❌ Development only |

**Configuration Example:**

```go
// HSM Provider Configuration
type HSMProviderConfig struct {
    LibraryPath string  // Path to PKCS#11 library
    SlotNumber  uint    // HSM slot number
    PIN         string  // HSM PIN (from env var)
    Label       string  // Key label
}

// Initialize HSM provider
func NewHSMProvider(config HSMProviderConfig) (*HSMProvider, error) {
    // Load PKCS#11 library
    p := pkcs11.New(config.LibraryPath)
    if err := p.Initialize(); err != nil {
        return nil, fmt.Errorf("pkcs11 init: %w", err)
    }
    
    // Open session
    session, err := p.OpenSession(config.SlotNumber, pkcs11.CKF_SERIAL_SESSION)
    if err != nil {
        return nil, fmt.Errorf("open session: %w", err)
    }
    
    // Login with PIN
    if err := p.Login(session, pkcs11.CKU_USER, config.PIN); err != nil {
        return nil, fmt.Errorf("hsm login: %w", err)
    }
    
    return &HSMProvider{ctx: p, session: session}, nil
}
```

### Key Rotation

**Rotation Schedule:**

| Key Type | Rotation Frequency | Procedure |
|----------|-------------------|-----------|
| Validator Consensus | Emergency only | Coordinated network upgrade |
| Provider Key | Annual or on compromise | Submit new key on-chain |
| Encryption Key | Annual | Re-encrypt active scopes |
| Session Keys | Per-session | Automatic |

**Key Rotation Procedure:**

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    KEY ROTATION PROCEDURE                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. Generate new keypair in HSM/secure storage                          │
│  2. Register new public key on-chain (MsgUpdateKey)                     │
│  3. Wait for confirmation (key becomes active)                          │
│  4. Re-encrypt affected data with new key                               │
│  5. Verify new key operations work correctly                            │
│  6. Revoke old key (after transition period)                            │
│  7. Securely destroy old private key material                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Secret Management

**Environment Variables:**

```bash
# Required environment variables for production
export VIRTENGINE_HOME="/opt/virtengine"
export VIRTENGINE_KEYRING_BACKEND="os"  # or "file" with passphrase
export VIRTENGINE_CHAIN_ID="virtengine-mainnet-1"

# Provider daemon secrets (from secret manager)
export PROVIDER_KEY_SECRET="$(vault read -field=secret secret/provider/key)"
export VAULT_PASSWORD="$(vault read -field=password secret/ansible/vault)"

# Never in environment for production:
# ❌ VIRTENGINE_MNEMONIC
# ❌ PROVIDER_PRIVATE_KEY
```

**Secret Handling in Code:**

```go
// ✅ Clear secrets from memory after use
func (k *KeyManager) SignTransaction(tx []byte, passphrase string) ([]byte, error) {
    // Clear passphrase from memory when done
    defer func() {
        for i := range passphrase {
            passphrase = passphrase[:i] + "\x00" + passphrase[i+1:]
        }
    }()
    
    // Sign transaction
    return k.doSign(tx, passphrase)
}

// ✅ Never log secrets
func ConnectToDatabase(connStr string) error {
    // Extract only safe portions for logging
    host := extractHost(connStr)
    log.Info("connecting to database", "host", host)
    // NOT: log.Info("connecting", "connStr", connStr)
    
    return db.Connect(connStr)
}
```

---

## Encryption in VirtEngine

### X25519-XSalsa20-Poly1305 Envelope

VirtEngine uses a standardized encryption envelope for all sensitive on-chain data:

```go
// Envelope structure for encrypted payloads
type EncryptionEnvelope struct {
    RecipientFingerprint string  // Validator's key fingerprint (SHA-256 of pubkey)
    Algorithm            string  // "X25519-XSalsa20-Poly1305"
    EphemeralPublicKey   []byte  // 32-byte X25519 public key
    Nonce                []byte  // 24-byte random nonce
    Ciphertext           []byte  // Encrypted data + Poly1305 tag
}
```

### Encryption Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    ENCRYPTION FLOW                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  CLIENT                                                                  │
│    │                                                                     │
│    ├──1. Generate ephemeral X25519 keypair                              │
│    │     ephemeralPriv, ephemeralPub = X25519.GenerateKey()            │
│    │                                                                     │
│    ├──2. Compute shared secret using ECDH                               │
│    │     sharedSecret = X25519(ephemeralPriv, recipientPub)            │
│    │                                                                     │
│    ├──3. Derive symmetric key (optional KDF)                            │
│    │     symmetricKey = SHA256(sharedSecret)                            │
│    │                                                                     │
│    ├──4. Generate random 24-byte nonce                                  │
│    │     nonce = crypto/rand.Read(24)                                   │
│    │                                                                     │
│    ├──5. Encrypt with XSalsa20-Poly1305                                 │
│    │     ciphertext = SecretBox.Seal(plaintext, nonce, key)            │
│    │                                                                     │
│    ├──6. Clear ephemeral private key                                    │
│    │     ephemeralPriv = zeros                                          │
│    │                                                                     │
│    └──7. Create envelope                                                 │
│          { recipientFingerprint, algorithm, ephemeralPub,               │
│            nonce, ciphertext }                                          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Encryption Implementation

```go
import (
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/curve25519"
    "golang.org/x/crypto/nacl/box"
)

// EncryptEnvelope creates an encryption envelope for sensitive data
func EncryptEnvelope(plaintext []byte, recipientPubKey *[32]byte) (*EncryptionEnvelope, error) {
    // 1. Generate ephemeral keypair
    var ephemeralPub, ephemeralPriv [32]byte
    if _, err := rand.Read(ephemeralPriv[:]); err != nil {
        return nil, fmt.Errorf("generate ephemeral key: %w", err)
    }
    curve25519.ScalarBaseMult(&ephemeralPub, &ephemeralPriv)
    
    // 2. Generate random nonce
    var nonce [24]byte
    if _, err := rand.Read(nonce[:]); err != nil {
        return nil, fmt.Errorf("generate nonce: %w", err)
    }
    
    // 3. Encrypt using NaCl box (X25519 + XSalsa20-Poly1305)
    ciphertext := box.Seal(nil, plaintext, &nonce, recipientPubKey, &ephemeralPriv)
    
    // 4. Clear ephemeral private key from memory
    for i := range ephemeralPriv {
        ephemeralPriv[i] = 0
    }
    
    // 5. Compute recipient fingerprint
    fingerprint := sha256.Sum256(recipientPubKey[:])
    
    return &EncryptionEnvelope{
        RecipientFingerprint: hex.EncodeToString(fingerprint[:8]),
        Algorithm:            "X25519-XSalsa20-Poly1305",
        EphemeralPublicKey:   ephemeralPub[:],
        Nonce:                nonce[:],
        Ciphertext:           ciphertext,
    }, nil
}
```

### Decryption Implementation

```go
// DecryptEnvelope decrypts an envelope using the recipient's private key
func DecryptEnvelope(envelope *EncryptionEnvelope, recipientPrivKey *[32]byte) ([]byte, error) {
    // Validate algorithm
    if envelope.Algorithm != "X25519-XSalsa20-Poly1305" {
        return nil, fmt.Errorf("unsupported algorithm: %s", envelope.Algorithm)
    }
    
    // Convert ephemeral public key
    var ephemeralPub [32]byte
    copy(ephemeralPub[:], envelope.EphemeralPublicKey)
    
    // Convert nonce
    var nonce [24]byte
    copy(nonce[:], envelope.Nonce)
    
    // Decrypt using NaCl box
    plaintext, ok := box.Open(nil, envelope.Ciphertext, &nonce, &ephemeralPub, recipientPrivKey)
    if !ok {
        return nil, errors.New("decryption failed: authentication error")
    }
    
    return plaintext, nil
}
```

### Multi-Recipient Encryption

For VEID data that multiple validators must decrypt:

```go
// MultiRecipientEnvelope encrypts for multiple recipients
type MultiRecipientEnvelope struct {
    Recipients []RecipientData  // Per-recipient encrypted key
    Payload    EncryptedPayload // Symmetrically encrypted data
}

type RecipientData struct {
    Fingerprint      string  // Recipient identifier
    EncryptedDataKey []byte  // Data key encrypted to this recipient
}

func EncryptForMultipleRecipients(plaintext []byte, recipientPubKeys []*[32]byte) (*MultiRecipientEnvelope, error) {
    // 1. Generate random data encryption key
    var dataKey [32]byte
    if _, err := rand.Read(dataKey[:]); err != nil {
        return nil, err
    }
    defer func() {
        for i := range dataKey {
            dataKey[i] = 0
        }
    }()
    
    // 2. Encrypt plaintext with data key
    var nonce [24]byte
    rand.Read(nonce[:])
    ciphertext := secretbox.Seal(nil, plaintext, &nonce, &dataKey)
    
    // 3. Encrypt data key for each recipient
    recipients := make([]RecipientData, len(recipientPubKeys))
    for i, pubKey := range recipientPubKeys {
        encKey, _ := EncryptEnvelope(dataKey[:], pubKey)
        fingerprint := sha256.Sum256(pubKey[:])
        recipients[i] = RecipientData{
            Fingerprint:      hex.EncodeToString(fingerprint[:8]),
            EncryptedDataKey: encKey.Ciphertext,
        }
    }
    
    return &MultiRecipientEnvelope{
        Recipients: recipients,
        Payload:    EncryptedPayload{Nonce: nonce[:], Ciphertext: ciphertext},
    }, nil
}
```

---

## Security Principles

### Defense in Depth

VirtEngine implements multiple security layers:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DEFENSE IN DEPTH                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Layer 1: Network Security                                              │
│  ├── TLS 1.3 for all connections                                       │
│  ├── Firewall rules and network segmentation                           │
│  └── DDoS protection at edge                                            │
│                                                                          │
│  Layer 2: Authentication                                                │
│  ├── Ed25519/secp256k1 signatures required                             │
│  ├── MFA for sensitive operations                                      │
│  └── Session management with timeouts                                   │
│                                                                          │
│  Layer 3: Authorization                                                 │
│  ├── Role-based access control (RBAC)                                  │
│  ├── Resource-level permissions                                        │
│  └── Module authority validation                                        │
│                                                                          │
│  Layer 4: Data Protection                                               │
│  ├── Encryption at rest (envelopes)                                    │
│  ├── Encryption in transit (TLS)                                       │
│  └── Key management with HSM                                            │
│                                                                          │
│  Layer 5: Audit & Detection                                             │
│  ├── Complete transaction audit trail                                  │
│  ├── Anomaly detection                                                 │
│  └── Alerting on security events                                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Least Privilege

```go
// ✅ CORRECT: Minimal permissions
type ProviderKeeper interface {
    // Only methods needed by marketplace
    GetProvider(ctx sdk.Context, addr sdk.Address) (Provider, bool)
    IsActiveProvider(ctx sdk.Context, addr sdk.Address) bool
    
    // NOT exposed: UpdateProvider, DeleteProvider, etc.
}

// ❌ WRONG: Exposing full keeper
type ProviderKeeper interface {
    *Keeper  // Exposes all methods!
}
```

### Fail Secure

```go
// ✅ CORRECT: Deny by default
func (k Keeper) CanAccessResource(ctx sdk.Context, user, resource string) bool {
    // Start with denial
    allowed := false
    
    // Check each permission explicitly
    if k.hasExplicitPermission(ctx, user, resource) {
        allowed = true
    }
    
    return allowed
}

// ❌ WRONG: Allow by default
func (k Keeper) CanAccessResource(ctx sdk.Context, user, resource string) bool {
    // Start with allow - dangerous!
    if k.isBlocked(ctx, user, resource) {
        return false
    }
    return true  // Default allow is dangerous
}
```

### Input Validation

```go
// ✅ CORRECT: Comprehensive validation
func (msg MsgCreateOrder) ValidateBasic() error {
    // Validate address format
    if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
        return sdkerrors.ErrInvalidAddress.Wrapf("owner: %s", err)
    }
    
    // Validate numeric ranges
    if msg.Price.IsNegative() || msg.Price.IsZero() {
        return sdkerrors.ErrInvalidRequest.Wrap("price must be positive")
    }
    
    // Validate string lengths
    if len(msg.Description) > MaxDescriptionLength {
        return sdkerrors.ErrInvalidRequest.Wrapf(
            "description exceeds max: %d > %d",
            len(msg.Description), MaxDescriptionLength,
        )
    }
    
    // Allowlist validation
    if !IsAllowedResourceType(msg.ResourceType) {
        return sdkerrors.ErrInvalidRequest.Wrapf(
            "invalid resource type: %s", msg.ResourceType,
        )
    }
    
    return nil
}
```

---

## Attack Vectors and Threats

### Threat Model

| Threat Actor | Capabilities | Primary Targets | Mitigation |
|--------------|--------------|-----------------|------------|
| **Malicious Validator** | Full node access, consensus participation | State manipulation, censorship | BFT consensus, slashing |
| **External Attacker** | Network access, API calls | DoS, data exfiltration | Rate limiting, encryption |
| **Malicious Provider** | Infrastructure control | Resource theft, data access | Signed attestations, auditing |
| **Compromised Client** | Signed transactions | Identity fraud, fund theft | MFA, transaction limits |

### Common Attack Vectors

#### 1. Replay Attacks

**Attack:** Resubmitting a valid signed transaction.

**Mitigation:**
```go
// Sequence numbers prevent replay
type Account struct {
    Address  sdk.AccAddress
    Sequence uint64  // Increments with each transaction
}

// Salt binding in VEID prevents scope replay
type ScopeSubmission struct {
    Salt         []byte    // Unique per submission
    Timestamp    time.Time // Time-bound validity
    Signatures   [][]byte  // Includes salt in signed data
}
```

#### 2. Man-in-the-Middle

**Attack:** Intercepting and modifying communications.

**Mitigation:**
- TLS 1.3 for all connections
- Certificate pinning for mobile apps
- End-to-end encryption for sensitive data

#### 3. Key Compromise

**Attack:** Theft of private keys.

**Mitigation:**
```
┌─────────────────────────────────────────────────────────────────────────┐
│  KEY COMPROMISE RESPONSE                                                 │
├─────────────────────────────────────────────────────────────────────────┤
│  1. Immediately revoke compromised key on-chain                         │
│  2. Rotate to backup key                                                │
│  3. Notify affected parties                                             │
│  4. Audit transactions made with compromised key                        │
│  5. Forensic analysis of compromise vector                              │
│  6. Update security controls                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 4. Denial of Service

**Attack:** Overwhelming system resources.

**Mitigation:**
```go
// Rate limiting
const (
    MaxRequestsPerSecond = 100
    MaxBatchSize         = 100
    MaxPayloadSize       = 1 << 20  // 1 MB
)

// Gas limits prevent computational DoS
func (k Keeper) ProcessHeavyOperation(ctx sdk.Context) error {
    gasLimit := ctx.GasMeter().Limit()
    if gasLimit < MinGasForOperation {
        return sdkerrors.ErrInvalidRequest.Wrap("insufficient gas")
    }
    // Consume gas proportional to work
    ctx.GasMeter().ConsumeGas(operationGas, "heavy_operation")
    return nil
}
```

#### 5. Consensus Attacks

**Attack:** Manipulating block production or transaction ordering.

**Mitigation:**
- BFT requires 2/3+ honest validators
- Slashing for double-signing
- MEV protection through encryption

### Vulnerability Categories

| Category | Description | Example | Prevention |
|----------|-------------|---------|------------|
| **Injection** | Untrusted data executed | SQL injection (unlikely in Go) | Parameterized queries, input validation |
| **Authentication** | Weak auth mechanisms | Predictable session tokens | Crypto-random tokens, proper key management |
| **Sensitive Data** | Exposed secrets | Logging credentials | Never log secrets, encrypt at rest |
| **Cryptographic** | Weak crypto | MD5 hashing | Use approved algorithms only |
| **Access Control** | Missing auth checks | Unauthenticated endpoints | Consistent auth checks |

---

## Defense Strategies

### Secure Coding Patterns

#### Safe Type Assertions

```go
// ❌ WRONG: Will panic
func handleMessage(msg sdk.Msg) {
    createOrder := msg.(*MsgCreateOrder)  // Panic if wrong type!
}

// ✅ CORRECT: Safe assertion
func handleMessage(msg sdk.Msg) error {
    createOrder, ok := msg.(*MsgCreateOrder)
    if !ok {
        return sdkerrors.ErrInvalidType.Wrapf("expected *MsgCreateOrder, got %T", msg)
    }
    return processOrder(createOrder)
}

// ✅ CORRECT: Type switch
func handleMessage(msg sdk.Msg) error {
    switch m := msg.(type) {
    case *MsgCreateOrder:
        return handleCreateOrder(m)
    case *MsgCancelOrder:
        return handleCancelOrder(m)
    default:
        return sdkerrors.ErrUnknownRequest.Wrapf("unknown message: %T", msg)
    }
}
```

#### Proper Error Handling

```go
// ❌ WRONG: Leaks internal details
func (k Keeper) GetUser(ctx sdk.Context, id string) (*User, error) {
    user, err := k.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, fmt.Errorf("database query failed: %v", err)  // Exposes schema!
    }
    return user, nil
}

// ✅ CORRECT: Generic user error, detailed internal log
func (k Keeper) GetUser(ctx sdk.Context, id string) (*User, error) {
    user, err := k.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        k.logger.Error("database query failed", "error", err, "query", "GetUser")
        return nil, sdkerrors.ErrNotFound.Wrap("user not found")
    }
    return user, nil
}
```

#### Resource Cleanup

```go
// ✅ CORRECT: Handle Close() errors
func readConfig(path string) (data []byte, err error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close: %w", cerr)
        }
    }()
    
    return io.ReadAll(f)
}
```

### Code Review Security Checklist

When reviewing code, verify:

**Authentication & Authorization:**
- [ ] All endpoints require appropriate authentication
- [ ] Authorization checks before state changes
- [ ] Module authority validated for admin operations
- [ ] Signer verification uses `msg.GetSigners()` correctly

**Input Validation:**
- [ ] All message fields validated in `ValidateBasic()`
- [ ] Size limits enforced
- [ ] Allowlists used instead of blocklists

**Cryptography:**
- [ ] Only approved algorithms
- [ ] Random values from `crypto/rand`
- [ ] Secrets cleared from memory

**State Management:**
- [ ] No non-deterministic operations
- [ ] Iterator resources closed
- [ ] Genesis import/export symmetric

**Error Handling:**
- [ ] No sensitive info in errors
- [ ] All errors wrapped with context
- [ ] Resources cleaned in error paths

---

## Key Takeaways

### Security Summary

1. **Use only approved cryptographic algorithms** - X25519, XSalsa20-Poly1305, AES-256-GCM, SHA-256, Ed25519

2. **Keys require proper protection** - HSM for validators, secure keyring for others, never in plaintext

3. **Encryption is mandatory for sensitive data** - All VEID and identity data uses envelope encryption

4. **Defense in depth** - Multiple layers (network, auth, authz, encryption, audit)

5. **Validate all inputs** - Never trust user input, use allowlists not blocklists

6. **Fail secure** - Default deny, explicit allow

7. **Determinism is critical** - All on-chain operations must be reproducible

### Quick Reference

| Topic | Key Point |
|-------|-----------|
| Encryption Algorithm | X25519-XSalsa20-Poly1305 |
| Key Exchange | X25519 (Curve25519 ECDH) |
| Signatures | Ed25519 (validators), secp256k1 (users) |
| Hashing | SHA-256 only |
| Random Generation | `crypto/rand` only |
| Key Storage | HSM for production validators |
| Secret Handling | Clear from memory, never log |

---

## Hands-On Exercises

### Exercise 1: Encryption Implementation (45 minutes)

**Objective:** Implement and test envelope encryption.

1. Write a function to encrypt a message:
```go
func EncryptMessage(message string, recipientPubKey *[32]byte) (*EncryptionEnvelope, error) {
    // Your implementation
}
```

2. Write the corresponding decryption function:
```go
func DecryptMessage(envelope *EncryptionEnvelope, privKey *[32]byte) (string, error) {
    // Your implementation
}
```

3. Write tests that verify:
   - [ ] Round-trip encryption/decryption works
   - [ ] Wrong key fails decryption
   - [ ] Modified ciphertext fails authentication

### Exercise 2: Code Review (30 minutes)

**Objective:** Identify security issues in sample code.

Review this code and list all security issues:

```go
func ProcessPayment(w http.ResponseWriter, r *http.Request) {
    amount := r.URL.Query().Get("amount")
    from := r.URL.Query().Get("from")
    to := r.URL.Query().Get("to")
    apiKey := r.Header.Get("X-API-Key")
    
    log.Printf("Processing payment: %s -> %s, amount: %s, key: %s", from, to, amount, apiKey)
    
    amountInt, _ := strconv.Atoi(amount)
    
    result := db.Exec(fmt.Sprintf("INSERT INTO payments (from_addr, to_addr, amount) VALUES ('%s', '%s', %d)", from, to, amountInt))
    
    if result != nil {
        w.Write([]byte("Success"))
    }
}
```

<details>
<summary>Click for issues</summary>

1. API key logged in plaintext
2. No input validation on amount, from, to
3. SQL injection vulnerability (string formatting)
4. Ignored error from strconv.Atoi
5. No authentication check
6. No authorization check
7. Error handling doesn't return early
8. Missing HTTPS check
9. No rate limiting
10. No CSRF protection
</details>

### Exercise 3: Key Management (30 minutes)

**Objective:** Configure secure key storage.

1. Set up a test keyring:
```bash
virtengine keys add test-user --keyring-backend file
```

2. Export and examine the key format (DO NOT do this with real keys):
```bash
virtengine keys export test-user --keyring-backend file
```

3. Document the key storage path and format
4. List the keyring backends and their security properties

### Exercise 4: Threat Modeling (45 minutes)

**Objective:** Analyze attack vectors for a scenario.

**Scenario:** A user wants to register as a provider and start accepting bids.

1. List all assets involved (keys, funds, reputation)
2. Identify threat actors (who might attack?)
3. List 5 potential attack vectors
4. Propose mitigations for each attack
5. Rate residual risk (high/medium/low)

---

## Assessment Questions

### Knowledge Check

1. **What encryption algorithm does VirtEngine use for sensitive on-chain data?**
   <details>
   <summary>Answer</summary>
   X25519-XSalsa20-Poly1305 - X25519 for key exchange, XSalsa20 for symmetric encryption, Poly1305 for authentication.
   </details>

2. **Why must ML scoring in VEID be deterministic?**
   <details>
   <summary>Answer</summary>
   All validators must produce identical scores during consensus. Non-deterministic scoring would cause validators to disagree, halting the chain or producing incorrect results.
   </details>

3. **What are the components of an EncryptionEnvelope?**
   <details>
   <summary>Answer</summary>
   RecipientFingerprint, Algorithm, EphemeralPublicKey, Nonce, and Ciphertext.
   </details>

4. **Why should you use `crypto/rand` instead of `math/rand` for security purposes?**
   <details>
   <summary>Answer</summary>
   `math/rand` is deterministic and predictable if seeded with known values. `crypto/rand` uses the operating system's cryptographically secure random number generator.
   </details>

5. **What is the principle of "Fail Secure"?**
   <details>
   <summary>Answer</summary>
   When errors occur, the system defaults to denying access rather than allowing it. Explicit permission is required for access.
   </details>

### Scenario Questions

6. **A validator's consensus key is suspected to be compromised. What are the immediate steps?**
   <details>
   <summary>Answer</summary>
   1. Immediately stop the validator node
   2. Notify other validators and the community
   3. Revoke/unbond the validator on-chain
   4. Forensic analysis of the compromise
   5. Rotate to backup key if available
   6. Review all blocks signed with the compromised key
   </details>

7. **You're reviewing code that uses `json.Marshal` for computing hashes in on-chain logic. What's the problem?**
   <details>
   <summary>Answer</summary>
   JSON marshaling is non-deterministic because map key ordering is not guaranteed. This would cause different validators to compute different hashes, breaking consensus. Use canonical (protobuf) encoding instead.
   </details>

8. **An attacker intercepts an EncryptionEnvelope and modifies the ciphertext. What prevents them from succeeding?**
   <details>
   <summary>Answer</summary>
   The Poly1305 authentication tag. XSalsa20-Poly1305 provides authenticated encryption - any modification to the ciphertext will be detected during decryption, and the operation will fail.
   </details>

---

## References

### Internal Documentation

| Document | Path | Description |
|----------|------|-------------|
| Security Guidelines | `_docs/security-guidelines.md` | Complete security coding guide |
| Key Management | `_docs/key-management.md` | Key management procedures |
| Threat Model | `_docs/threat-model.md` | System threat analysis |
| Encryption Module | `x/encryption/README.md` | Encryption implementation |

### External Resources

| Resource | URL | Description |
|----------|-----|-------------|
| NaCl Library | https://nacl.cr.yp.to/ | Crypto primitives reference |
| OWASP Go Guidelines | https://owasp.org/www-project-go-secure-coding-practices-guide/ | Secure Go coding |
| Cosmos SDK Security | https://docs.cosmos.network/main/build/building-modules/errors | SDK security practices |
| CWE Top 25 | https://cwe.mitre.org/top25/ | Common weaknesses |

### Next Modules

After completing this module, proceed to:

1. **Incident Response Basics** (`incident-response-basics.md`) - Handle security incidents
2. **Advanced Cryptography** (`../security/advanced-crypto.md`) - Deep dive into crypto
3. **Validator Security** (`../validator/security.md`) - Validator-specific security

---

*Module Version: 1.0.0 | Last Updated: 2025-01-24 | Maintainer: VirtEngine Security Team*