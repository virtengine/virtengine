# VirtEngine Encryption Documentation

**Version:** 1.0.0  
**Date:** 2026-01-30  
**Status:** Authoritative Baseline  
**Task Reference:** DOCS-003  
**Classification:** Internal

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Encryption Standards](#encryption-standards)
3. [Encryption at Rest](#encryption-at-rest)
4. [Encryption in Transit](#encryption-in-transit)
5. [End-to-End Encryption](#end-to-end-encryption)
6. [Key Management](#key-management)
7. [Cryptographic Agility](#cryptographic-agility)
8. [Implementation Guidelines](#implementation-guidelines)
9. [Compliance Requirements](#compliance-requirements)

---

## Executive Summary

This document defines VirtEngine's encryption strategy for protecting data at rest and in transit. All sensitive data is encrypted using industry-standard algorithms, with key management following best practices for blockchain and decentralized systems.

### Encryption Coverage

| Data State | Coverage | Algorithm | Key Management |
|------------|----------|-----------|----------------|
| At Rest (On-chain) | 100% of sensitive data | X25519-XSalsa20-Poly1305 | Per-recipient |
| At Rest (Off-chain) | 100% | AES-256-GCM | AWS KMS / Vault |
| In Transit (External) | 100% | TLS 1.3 | Certificate rotation |
| In Transit (P2P) | 100% | Noise Protocol | Peer Ed25519 keys |
| End-to-End | All identity data | Envelope encryption | Threshold decryption |

---

## Encryption Standards

### Approved Algorithms

| Purpose | Algorithm | Key Size | Standard | Status |
|---------|-----------|----------|----------|--------|
| **Symmetric Encryption** |||||
| Authenticated encryption | XSalsa20-Poly1305 | 256-bit | NaCl | ✅ Primary |
| Authenticated encryption | AES-256-GCM | 256-bit | NIST | ✅ Alternative |
| **Asymmetric Encryption** |||||
| Key exchange | X25519 | 256-bit | RFC 7748 | ✅ Primary |
| Key exchange | ECDH P-256 | 256-bit | NIST | ✅ Alternative |
| **Digital Signatures** |||||
| Signing | Ed25519 | 256-bit | RFC 8032 | ✅ Primary |
| Wallet signatures | secp256k1 | 256-bit | SEC 2 | ✅ Required |
| Signing | ECDSA P-256 | 256-bit | NIST | ⚠️ Legacy |
| **Hashing** |||||
| General purpose | SHA-256 | 256-bit | FIPS 180-4 | ✅ Primary |
| High security | SHA-512 | 512-bit | FIPS 180-4 | ✅ Alternative |
| Password hashing | Argon2id | N/A | RFC 9106 | ✅ Required |
| **Key Derivation** |||||
| Key derivation | HKDF-SHA256 | 256-bit | RFC 5869 | ✅ Primary |
| Password-based | Argon2id | N/A | RFC 9106 | ✅ Required |

### Deprecated/Prohibited Algorithms

| Algorithm | Status | Reason | Migration |
|-----------|--------|--------|-----------|
| MD5 | ❌ Prohibited | Broken | Use SHA-256 |
| SHA-1 | ❌ Prohibited | Weak | Use SHA-256 |
| DES/3DES | ❌ Prohibited | Weak | Use AES-256 |
| RC4 | ❌ Prohibited | Broken | Use AES-GCM |
| RSA-1024 | ❌ Prohibited | Weak | Use RSA-2048+ or Ed25519 |
| TLS 1.0/1.1 | ❌ Prohibited | Vulnerable | Use TLS 1.3 |
| TLS 1.2 | ⚠️ Deprecated | Upgrade | Use TLS 1.3 |

---

## Encryption at Rest

### On-Chain Data

All sensitive on-chain data uses the **EncryptedPayloadEnvelope** format.

#### Envelope Structure

```go
// EncryptedPayloadEnvelope is the standard format for on-chain encrypted data
type EncryptedPayloadEnvelope struct {
    // Algorithm identifies the encryption algorithm
    // Currently: "X25519-XSalsa20-Poly1305"
    Algorithm     string `json:"algorithm"`
    
    // Version for protocol evolution
    Version       uint32 `json:"version"`
    
    // EphemeralKey is the sender's ephemeral public key (32 bytes)
    EphemeralKey  []byte `json:"ephemeral_key"`
    
    // Nonce is the encryption nonce (24 bytes)
    Nonce         []byte `json:"nonce"`
    
    // Ciphertext is the encrypted payload
    Ciphertext    []byte `json:"ciphertext"`
    
    // RecipientKeys contains per-recipient wrapped DEKs
    RecipientKeys []WrappedKey `json:"recipient_keys"`
    
    // Signature binds the envelope to the sender
    Signature     []byte `json:"signature"`
}

type WrappedKey struct {
    // Fingerprint identifies the recipient key
    Fingerprint   string `json:"fingerprint"`
    
    // EncryptedDEK is the data encryption key encrypted to recipient
    EncryptedDEK  []byte `json:"encrypted_dek"`
}
```

#### Encryption Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ON-CHAIN ENCRYPTION FLOW                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   SENDER                                                                     │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 1. Generate Data Encryption Key (DEK) - 32 bytes from crypto/rand   │  │
│   │ 2. Generate ephemeral X25519 keypair                                 │  │
│   │ 3. For each recipient:                                               │  │
│   │    a. Perform X25519 ECDH with recipient public key                 │  │
│   │    b. Derive shared secret using HKDF-SHA256                        │  │
│   │    c. Encrypt DEK with XSalsa20-Poly1305 using shared secret        │  │
│   │    d. Store as WrappedKey with recipient fingerprint                │  │
│   │ 4. Encrypt plaintext with DEK using XSalsa20-Poly1305               │  │
│   │ 5. Generate 24-byte random nonce                                     │  │
│   │ 6. Sign envelope (SHA256 binding to ephemeral key + payload)        │  │
│   │ 7. Zero DEK from memory                                              │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│   BLOCKCHAIN                                                                 │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ • Envelope stored in transaction/state                               │  │
│   │ • Plaintext never visible on-chain                                   │  │
│   │ • Only encrypted ciphertext persisted                                │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│   RECIPIENT                                                                  │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 1. Find WrappedKey with matching fingerprint                         │  │
│   │ 2. Perform X25519 ECDH with ephemeral key from envelope             │  │
│   │ 3. Derive shared secret using HKDF-SHA256                            │  │
│   │ 4. Decrypt DEK using XSalsa20-Poly1305                               │  │
│   │ 5. Decrypt ciphertext using DEK                                      │  │
│   │ 6. Verify signature binding                                          │  │
│   │ 7. Zero DEK from memory                                              │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### On-Chain Encrypted Data Types

| Data Type | Module | Recipients | Retention |
|-----------|--------|------------|-----------|
| Identity documents | x/veid | Validators (threshold) | Until erasure |
| Selfie/video captures | x/veid | Validators (threshold) | Until erasure |
| Order credentials | x/market | Provider + Customer | Lease duration |
| Order configuration | x/market | Provider | Lease duration |
| TOTP secrets | x/mfa | User only | Until revoked |

### Off-Chain Data

#### Database Encryption

| Database | Encryption | Key Management | Key Rotation |
|----------|------------|----------------|--------------|
| PostgreSQL | TDE (Transparent Data Encryption) | AWS RDS encryption | Automatic |
| Redis | TLS + at-rest | AWS ElastiCache | Automatic |
| Elasticsearch | Index encryption | AWS Elasticsearch | Automatic |

**Configuration Example (PostgreSQL on AWS RDS):**

```yaml
# terraform/rds.tf
resource "aws_db_instance" "virtengine" {
  engine               = "postgres"
  engine_version       = "15"
  instance_class       = "db.r6g.large"
  
  # Encryption at rest
  storage_encrypted    = true
  kms_key_id          = aws_kms_key.rds.arn
  
  # Encryption in transit
  parameter_group_name = aws_db_parameter_group.encrypted.name
}

resource "aws_db_parameter_group" "encrypted" {
  family = "postgres15"
  
  parameter {
    name  = "rds.force_ssl"
    value = "1"
  }
}
```

#### Backup Encryption

| Backup Type | Encryption | Key | Rotation |
|-------------|------------|-----|----------|
| Database snapshots | AES-256 | AWS KMS | Per-snapshot |
| File backups | AES-256-GCM | Backup key | Monthly |
| Log archives | AES-256-GCM | Log key | Daily |

**Backup Encryption Process:**

```bash
# Encrypted backup creation
openssl enc -aes-256-gcm \
    -in backup.tar.gz \
    -out backup.tar.gz.enc \
    -pass file:/secrets/backup-key \
    -pbkdf2

# Backup key rotation
aws kms generate-data-key --key-id alias/backup-key --key-spec AES_256
```

#### Log Encryption

- All logs encrypted at rest in SIEM
- Log forwarding uses TLS 1.3
- PII automatically redacted before logging
- Encryption key rotated daily

---

## Encryption in Transit

### TLS Configuration

#### Supported Configuration

```yaml
# TLS 1.3 Configuration
tls:
  min_version: "TLS 1.3"
  cipher_suites:
    - TLS_AES_256_GCM_SHA384
    - TLS_CHACHA20_POLY1305_SHA256
  key_exchange:
    - X25519
    - P-256
  certificate:
    type: ECDSA
    curve: P-256
  hsts:
    enabled: true
    max_age: 31536000
    include_subdomains: true
    preload: true
```

#### Endpoint Configuration

| Endpoint | Port | TLS Version | Certificate | mTLS |
|----------|------|-------------|-------------|------|
| api.virtengine.io | 443 | TLS 1.3 | ECDSA P-256 | No |
| rpc.virtengine.io | 443 | TLS 1.3 | ECDSA P-256 | No |
| grpc.virtengine.io | 443 | TLS 1.3 | ECDSA P-256 | Optional |
| Internal services | 443 | TLS 1.3 | Internal CA | Yes |

### P2P Network Encryption

Validator-to-validator communication uses the Noise Protocol framework.

#### Noise Protocol Configuration

```go
// Noise Protocol configuration for P2P
type NoiseConfig struct {
    // Protocol pattern
    Pattern: "XX"  // Full mutual authentication
    
    // DH function
    DH: "25519"  // X25519
    
    // Cipher
    Cipher: "ChaChaPoly"  // ChaCha20-Poly1305
    
    // Hash
    Hash: "BLAKE2b"
}
```

#### P2P Encryption Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    P2P NOISE PROTOCOL HANDSHAKE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   INITIATOR                                                    RESPONDER    │
│   (Validator A)                                               (Validator B) │
│                                                                              │
│   1. Generate ephemeral keypair                                             │
│      ┌──────────────┐                                                       │
│      │ e = X25519() │                                                       │
│      └──────────────┘                                                       │
│                                                                              │
│   2. Send ephemeral public key                                              │
│      ──────────────────────────────────────────────────────────────────────▶│
│      [e.public, encrypted(s.public)]                                        │
│                                                                              │
│   3. Responder generates ephemeral keypair                                  │
│                                                                 ┌──────────┐│
│                                                                 │e = X25519││
│                                                                 └──────────┘│
│                                                                              │
│   4. Responder sends response                                               │
│      ◀──────────────────────────────────────────────────────────────────────│
│      [e.public, encrypted(s.public)]                                        │
│                                                                              │
│   5. Both derive shared secrets:                                            │
│      - ee = DH(initiator.e, responder.e)                                   │
│      - se = DH(initiator.s, responder.e)                                   │
│      - es = DH(initiator.e, responder.s)                                   │
│      - ss = DH(initiator.s, responder.s)                                   │
│                                                                              │
│   6. Derive transport keys using HKDF                                       │
│      ┌────────────────────────────────────────────────────────────────────┐│
│      │ tx_key = HKDF(ee || se || es || ss, "noise-transport-keys")        ││
│      └────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│   7. All subsequent messages encrypted with ChaCha20-Poly1305              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### gRPC Encryption

```go
// gRPC client with TLS
func NewGRPCClient(endpoint string) (*grpc.ClientConn, error) {
    // Load TLS credentials
    creds, err := credentials.NewClientTLSFromFile("certs/ca.pem", "")
    if err != nil {
        return nil, err
    }
    
    // Connect with TLS
    conn, err := grpc.Dial(
        endpoint,
        grpc.WithTransportCredentials(creds),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             3 * time.Second,
            PermitWithoutStream: true,
        }),
    )
    
    return conn, err
}
```

---

## End-to-End Encryption

### Identity Data (VEID)

Identity documents are encrypted end-to-end from capture to verification.

#### Threshold Decryption

For identity verification, data is encrypted to multiple validators using threshold cryptography.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    THRESHOLD DECRYPTION (t-of-n)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ENCRYPTION (Client-side)                                                   │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 1. Generate DEK                                                       │  │
│   │ 2. Encrypt identity data with DEK                                     │  │
│   │ 3. For each of n validators:                                          │  │
│   │    - Wrap DEK with validator's public key                            │  │
│   │ 4. Store envelope on-chain                                            │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│   DECRYPTION (Validator-side)                                                │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │ 1. Each validator decrypts their share of DEK                         │  │
│   │ 2. Validators run ML scoring locally                                  │  │
│   │ 3. Validators submit signed scores to consensus                       │  │
│   │ 4. Consensus determines final score (median or weighted)             │  │
│   │ 5. Plaintext never leaves validator TEE                               │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│   SECURITY PROPERTIES                                                        │
│   • No single validator can access identity data alone                      │
│   • Collusion of t-1 validators reveals nothing                             │
│   • Plaintext processed in TEE only                                         │
│   • Score is public, documents remain encrypted                             │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Provider-Customer Communication

Order credentials (SSH keys, passwords) are encrypted end-to-end between customer and provider.

```go
// Customer encrypts credentials for provider
envelope, err := crypto.Encrypt(
    credentials,
    []crypto.RecipientPublicKey{
        providerPublicKey,  // Provider can decrypt
        customerPublicKey,  // Customer can decrypt (for reference)
    },
    customerPrivateKey,
)

// Submit order with encrypted envelope
msg := &markettypes.MsgCreateOrder{
    Buyer:      customerAddress,
    Offering:   offeringID,
    Envelope:   envelope.Bytes(),
}
```

---

## Key Management

### Key Hierarchy

| Level | Key Type | Storage | Rotation | Access |
|-------|----------|---------|----------|--------|
| L0 | Genesis Authority | Air-gapped HSM | Never | Multi-sig 3-of-5 |
| L0 | Emergency Recovery | Geographically distributed | Never | Multi-sig 5-of-7 |
| L1 | Validator Consensus | HSM | Annual | Validator operator |
| L1 | Validator Identity Decrypt | HSM | Annual | Threshold (validators) |
| L2 | Provider Daemon | Secure enclave | Quarterly | Provider operator |
| L2 | API TLS | Certificate manager | 90 days | Automated |
| L3 | User Account | User wallet | User-controlled | User |
| L4 | Ephemeral | Memory only | Per-operation | System |

### Key Storage Requirements

| Key Classification | Storage Requirement | Examples |
|--------------------|---------------------|----------|
| Critical (C4) | HSM or hardware wallet | Validator keys, genesis keys |
| Restricted (C3) | Encrypted + access controlled | Provider keys, TLS keys |
| Confidential (C2) | Encrypted at rest | Session keys, API keys |
| Internal (C1) | Standard protection | Non-sensitive keys |

### Key Rotation Procedures

See [_docs/key-management.md](../../_docs/key-management.md) for detailed procedures.

---

## Cryptographic Agility

### Algorithm Migration Plan

| Current | Future (PQC) | Migration Path | Timeline |
|---------|--------------|----------------|----------|
| X25519 | Kyber-768 (ML-KEM) | Hybrid mode first | 2027+ |
| Ed25519 | Dilithium-3 (ML-DSA) | Hybrid mode first | 2027+ |
| secp256k1 | To be determined | Cosmos SDK upgrade | TBD |
| SHA-256 | SHA-3 or BLAKE3 | Soft fork | As needed |

### Hybrid Mode Implementation

```go
// Future: Hybrid key exchange (classical + PQC)
type HybridKeyExchange struct {
    Classical   X25519KeyExchange   // Current
    PostQuantum Kyber768KeyExchange // Future
}

func (h *HybridKeyExchange) DeriveSharedSecret(peer HybridPublicKey) []byte {
    // Combine both shared secrets
    classicalSecret := h.Classical.DeriveSecret(peer.Classical)
    pqcSecret := h.PostQuantum.DeriveSecret(peer.PostQuantum)
    
    // Use HKDF to combine
    return hkdf.Extract(sha256.New, classicalSecret, pqcSecret)
}
```

---

## Implementation Guidelines

### Developer Checklist

#### Encryption

- [ ] Use `x/encryption/crypto` package for all encryption
- [ ] Never implement custom cryptography
- [ ] Use `crypto/rand` for randomness (off-chain only)
- [ ] Zero sensitive data after use (`crypto.ZeroBytes()`)
- [ ] Verify all recipients can decrypt before submitting

#### Key Handling

- [ ] Never log key material
- [ ] Never store keys in plaintext
- [ ] Use secure key storage (keyring, HSM, secure enclave)
- [ ] Implement key rotation
- [ ] Follow least privilege for key access

#### Code Examples

```go
// CORRECT: Use the encryption package
import "virtengine/x/encryption/crypto"

func encryptData(plaintext []byte, recipient crypto.PublicKey) (*types.Envelope, error) {
    return crypto.Encrypt(plaintext, []crypto.PublicKey{recipient}, nil)
}

// INCORRECT: Do not implement custom crypto
func badEncrypt(plaintext []byte, key []byte) []byte {
    // DON'T DO THIS
    cipher, _ := aes.NewCipher(key)
    // ... custom implementation
}
```

### Security Review Checklist

| Check | Description | Required For |
|-------|-------------|--------------|
| Algorithm review | Verify approved algorithms | All crypto code |
| Key management | Verify key storage/rotation | All key handling |
| Random generation | Verify `crypto/rand` usage | All random values |
| Memory zeroing | Verify sensitive data cleared | All key material |
| Error handling | Verify no info leakage | All crypto operations |
| Constant-time | Verify timing-safe comparisons | All verification |

---

## Compliance Requirements

### GDPR (Article 32)

| Requirement | Implementation | Evidence |
|-------------|----------------|----------|
| Encryption of personal data | Envelope encryption | x/encryption/ |
| Pseudonymization | Feature hashes only | x/veid/ |
| Integrity and confidentiality | AEAD encryption | XSalsa20-Poly1305 |
| Restoration capability | Key backup | key-management.md |

### SOC 2 (CC6.1, CC6.6)

| Criterion | Implementation | Evidence |
|-----------|----------------|----------|
| Logical access to encryption keys | RBAC, HSM | Access logs |
| Encryption of data at rest | AES-256, envelope | Configuration |
| Encryption of data in transit | TLS 1.3 | Network config |

### ISO 27001 (A.10)

| Control | Implementation | Evidence |
|---------|----------------|----------|
| A.10.1.1 Cryptographic policy | This document | ENCRYPTION.md |
| A.10.1.2 Key management | Key management system | key-management.md |

---

## Related Documentation

| Document | Purpose | Location |
|----------|---------|----------|
| Key Management | Key lifecycle procedures | [_docs/key-management.md](../../_docs/key-management.md) |
| Cryptographic Audit | Security review | [SECURITY-001-CRYPTOGRAPHIC-AUDIT.md](../../SECURITY-001-CRYPTOGRAPHIC-AUDIT.md) |
| Security Architecture | Overall security | [SECURITY_ARCHITECTURE.md](./SECURITY_ARCHITECTURE.md) |
| Data Classification | Data sensitivity | [_docs/data-classification.md](../../_docs/data-classification.md) |

---

**Document Owner**: Security Team  
**Last Updated**: 2026-01-30  
**Version**: 1.0.0  
**Next Review**: 2026-04-30
