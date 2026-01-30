# VEID Verification Attestation Schema

**Task Reference:** VE-1B  
**Schema Version:** 1.0.0  
**Status:** Draft  
**Last Updated:** 2024-01-15

## Overview

This document specifies the verification attestation schema for VEID verification services. Attestations are cryptographically signed statements that assert the result of an identity verification process.

## Table of Contents

1. [Attestation Schema](#attestation-schema)
2. [Signature Format](#signature-format)
3. [Canonical Serialization](#canonical-serialization)
4. [Key Rotation Policy](#key-rotation-policy)
5. [Replay Protection](#replay-protection)
6. [On-Chain Linkage](#on-chain-linkage)
7. [Audit Logging](#audit-logging)
8. [Signer Service Contract](#signer-service-contract)
9. [Example Payloads](#example-payloads)

---

## Attestation Schema

### VerificationAttestation

The core attestation structure that represents a signed verification result.

```json
{
  "id": "veid:attestation:<issuer_fingerprint_prefix>:<nonce_prefix>",
  "schema_version": "1.0.0",
  "type": "facial_verification",
  "issuer": {
    "id": "did:virtengine:validator:<address>",
    "key_fingerprint": "<sha256_hex_64_chars>",
    "key_id": "validator-key-001",
    "validator_address": "virtengine1...",
    "service_endpoint": "https://veid-signer.virtengine.io/v1"
  },
  "subject": {
    "id": "did:virtengine:<account_address>",
    "account_address": "virtengine1...",
    "scope_id": "scope-facial-001",
    "request_id": "req-20240115-001"
  },
  "nonce": "<hex_encoded_32_bytes>",
  "issued_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-02-14T10:30:00Z",
  "verification_proofs": [
    {
      "proof_type": "facial_embedding_match",
      "content_hash": "sha256:abc123...",
      "score": 94,
      "passed": true,
      "threshold": 80,
      "timestamp": "2024-01-15T10:29:55Z"
    }
  ],
  "score": 92,
  "confidence": 95,
  "model_version": "facial-v2.3.0",
  "metadata": {
    "pipeline_version": "1.0.0",
    "validator_consensus": "5/5"
  },
  "proof": {
    "type": "Ed25519Signature2020",
    "created": "2024-01-15T10:30:00Z",
    "verification_method": "did:virtengine:validator:<address>#validator-key-001",
    "proof_purpose": "assertionMethod",
    "proof_value": "<base64_signature>",
    "nonce": "<same_as_attestation_nonce>",
    "domain": "veid.virtengine.io"
  }
}
```

### Attestation Types

| Type | Description | Validity Period |
|------|-------------|-----------------|
| `facial_verification` | Facial recognition match | 30 days |
| `liveness_check` | Liveness detection | 30 days |
| `document_verification` | ID document verification | 1 year |
| `email_verification` | Email ownership proof | 90 days |
| `sms_verification` | Phone number verification | 90 days |
| `domain_verification` | Domain ownership proof | 1 year |
| `sso_verification` | SSO provider verification | 90 days |
| `biometric_verification` | Biometric data verification | 30 days |
| `composite_identity` | Combined identity attestation | 30 days |

### Attestation Type to Scope Type Mapping

| Attestation Type | Scope Type |
|-----------------|------------|
| `facial_verification` | `selfie` |
| `liveness_check` | `face_video` |
| `document_verification` | `id_document` |
| `email_verification` | `email_proof` |
| `sms_verification` | `sms_proof` |
| `domain_verification` | `domain_verify` |
| `sso_verification` | `sso_metadata` |
| `biometric_verification` | `biometric` |

---

## Signature Format

### Supported Algorithms

| Algorithm | Type Identifier | Key Size | Usage |
|-----------|-----------------|----------|-------|
| Ed25519 | `Ed25519Signature2020` | 256 bits | **Recommended** |
| secp256k1 | `EcdsaSecp256k1Signature2019` | 256 bits | Cosmos compatible |
| sr25519 | `Sr25519Signature2020` | 256 bits | Substrate compatible |

### Proof Structure

```json
{
  "type": "Ed25519Signature2020",
  "created": "2024-01-15T10:30:00Z",
  "verification_method": "<issuer_id>#<key_id>",
  "proof_purpose": "assertionMethod",
  "proof_value": "<base64_encoded_signature>",
  "nonce": "<attestation_nonce_hex>",
  "domain": "veid.virtengine.io",
  "challenge": "<optional_challenge>"
}
```

### Key Fingerprint

Key fingerprints are computed as SHA256 hash of the raw public key bytes:

```
fingerprint = hex(SHA256(public_key_bytes))
```

The fingerprint is a 64-character hex string.

---

## Canonical Serialization

To ensure deterministic signing and verification, attestations are serialized using canonical JSON:

### Canonicalization Rules

1. **Exclude proof field**: The `proof` field is not included in the signed payload
2. **UTC timestamps**: All timestamps use RFC3339Nano format in UTC
3. **Sorted keys**: JSON object keys are lexicographically sorted
4. **No whitespace**: No unnecessary whitespace in serialization
5. **Sorted metadata**: Metadata map keys are sorted alphabetically

### Canonical Structure

```json
{
  "confidence": 95,
  "expires_at": "2024-02-14T10:30:00.000000000Z",
  "id": "veid:attestation:1234567890abcdef:a1b2c3d4e5f6a1b2",
  "issued_at": "2024-01-15T10:30:00.000000000Z",
  "issuer": { ... },
  "metadata": { "key1": "value1", "key2": "value2" },
  "model_version": "facial-v2.3.0",
  "nonce": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
  "schema_version": "1.0.0",
  "score": 92,
  "subject": { ... },
  "type": "facial_verification",
  "verification_proofs": [ ... ]
}
```

### Signing Process

1. Serialize attestation to canonical JSON (excluding `proof`)
2. Compute SHA256 hash of canonical bytes
3. Sign hash with signer's private key
4. Base64-encode signature
5. Attach proof to attestation

---

## Key Rotation Policy

### Default Policy Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `max_key_age_seconds` | 7,776,000 (90 days) | Maximum key lifetime |
| `rotation_overlap_seconds` | 604,800 (7 days) | Grace period for in-flight attestations |
| `min_rotation_notice_seconds` | 259,200 (3 days) | Minimum notice before rotation |
| `max_pending_keys` | 2 | Maximum successor keys |
| `require_successor_key` | true | Require successor before rotation |
| `allow_emergency_revocation` | true | Allow immediate revocation |
| `min_key_strength` | 256 bits | Minimum key strength |

### Key Lifecycle States

```
┌─────────┐     ┌────────┐     ┌──────────┐     ┌─────────┐
│ Pending │────▶│ Active │────▶│ Rotating │────▶│ Expired │
└─────────┘     └────────┘     └──────────┘     └─────────┘
                     │                              ▲
                     │         ┌─────────┐          │
                     └────────▶│ Revoked │◀─────────┘
                               └─────────┘
```

| State | Can Sign | Can Verify | Description |
|-------|----------|------------|-------------|
| `pending` | No | No | Awaiting activation |
| `active` | Yes | Yes | Currently valid |
| `rotating` | Yes | Yes | Transitioning to new key |
| `revoked` | No | No | Key has been revoked |
| `expired` | No | No | Key has expired |

### Revocation Reasons

| Reason | Description | Overlap Period |
|--------|-------------|----------------|
| `compromised` | Key was compromised | None (immediate) |
| `rotation` | Normal key rotation | Full overlap |
| `decommissioned` | Signer is decommissioned | 7 days |
| `policy_violation` | Policy violation | None (immediate) |
| `administrative` | Administrative action | 7 days |

### Key Rotation Process

1. **Generate successor key** in `pending` state
2. **Activate successor** (old key transitions to `rotating`)
3. **Overlap period** where both keys are valid
4. **Revoke old key** after overlap period

---

## Replay Protection

### Nonce Requirements

| Parameter | Value |
|-----------|-------|
| Minimum length | 16 bytes |
| Maximum length | 64 bytes |
| Default length | 32 bytes |
| Character encoding | Hex |

### Nonce Validity Window

| Parameter | Default | Range |
|-----------|---------|-------|
| Window duration | 1 hour | 5 min - 24 hours |
| Clock skew tolerance | 5 minutes | ≤ window/2 |
| Max nonces per issuer | 10,000 | N/A |

### Nonce Rules

1. **Uniqueness**: Each nonce must be unique per issuer within the validity window
2. **Timestamp binding**: Nonces must be used within clock skew of issuance time
3. **Issuer binding**: Nonces are bound to issuer key fingerprint
4. **One-time use**: Once used, a nonce cannot be reused
5. **Expiry**: Unused nonces expire after the validity window

### Weak Nonce Detection

The following nonces are rejected:

- All zeros (e.g., `0000...0000`)
- All ones (e.g., `ffff...ffff`)
- Nonces shorter than 16 bytes
- Nonces longer than 64 bytes
- Non-hex encoded nonces

---

## On-Chain Linkage

### Store Key Prefixes

| Prefix | Key Format | Value Type |
|--------|------------|------------|
| `0x7C` | `attestation_id` | `VerificationAttestation` |
| `0x7D` | `subject_address \| attestation_id` | `bool` |
| `0x7E` | `issuer_fingerprint \| attestation_id` | `bool` |
| `0x7F` | `attestation_type \| attestation_id` | `bool` |
| `0x80` | `expires_at \| attestation_id` | `bool` |
| `0x81` | `nonce_hash` | `NonceRecord` |
| `0x83` | `key_id` | `SignerKeyInfo` |
| `0x86` | `signer_id` | `SignerRegistryEntry` |
| `0x87` | `rotation_id` | `KeyRotationRecord` |

### Verification Flow

```
1. Receive attestation
2. Validate schema version
3. Validate attestation type
4. Validate issuer (registered, active key)
5. Validate nonce (unique, not expired)
6. Validate timestamp (within window)
7. Validate signature
8. Validate score/confidence bounds
9. Store attestation
10. Link to identity scope
11. Emit events
```

---

## Audit Logging

### Event Types

| Event | Description | Emitted When |
|-------|-------------|--------------|
| `attestation_created` | New attestation created | Attestation stored |
| `attestation_revoked` | Attestation revoked | Manual revocation |
| `attestation_expired` | Attestation expired | Expiry cleanup |
| `signer_key_registered` | New signer key | Key registration |
| `signer_key_activated` | Key activated | Key activation |
| `signer_key_revoked` | Key revoked | Key revocation |
| `signer_key_rotated` | Key rotation complete | Rotation complete |
| `nonce_used` | Nonce consumed | Attestation created |
| `attestation_verified` | Signature verified | Verification complete |

### Event Attributes

All attestation events include:

- `block_height`: Block when event occurred
- `timestamp`: Event timestamp (Unix)
- `attestation_id`: Attestation identifier
- `issuer_id`: Issuer DID
- `subject_address`: Subject account address

---

## Signer Service Contract

### Service Endpoints

#### POST /v1/attestations

Create a new verification attestation.

**Request:**

```json
{
  "subject_address": "virtengine1...",
  "attestation_type": "facial_verification",
  "scope_id": "scope-123",
  "request_id": "req-456",
  "verification_proofs": [...],
  "score": 92,
  "confidence": 95,
  "model_version": "facial-v2.3.0",
  "validity_duration_seconds": 2592000
}
```

**Response:**

```json
{
  "attestation": { ... },
  "attestation_hash": "<sha256_hex>",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### GET /v1/attestations/{id}

Retrieve an attestation by ID.

#### GET /v1/signers/{signer_id}/keys

List signer keys.

#### POST /v1/signers/{signer_id}/keys/rotate

Initiate key rotation.

### Authentication

All requests must include:

```
Authorization: Bearer <validator_jwt>
X-Validator-Address: virtengine1...
X-Request-Signature: <ed25519_signature>
```

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_SUBJECT` | Subject address is invalid |
| `INVALID_TYPE` | Attestation type is invalid |
| `NONCE_COLLISION` | Nonce already used |
| `KEY_NOT_FOUND` | Signing key not found |
| `KEY_REVOKED` | Signing key has been revoked |
| `KEY_EXPIRED` | Signing key has expired |
| `SIGNATURE_FAILED` | Signature generation failed |
| `RATE_LIMITED` | Too many requests |

---

## Example Payloads

### Facial Verification Attestation

```json
{
  "id": "veid:attestation:1234567890abcdef:a1b2c3d4e5f6a1b2",
  "schema_version": "1.0.0",
  "type": "facial_verification",
  "issuer": {
    "id": "did:virtengine:validator:virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
    "key_fingerprint": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "key_id": "validator-key-001",
    "validator_address": "virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
    "service_endpoint": "https://veid-signer.virtengine.io/v1"
  },
  "subject": {
    "id": "did:virtengine:virtengine1abc123def456ghi789jkl012mno345pqr678stu",
    "account_address": "virtengine1abc123def456ghi789jkl012mno345pqr678stu",
    "scope_id": "scope-facial-001",
    "request_id": "req-20240115-001"
  },
  "nonce": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
  "issued_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-02-14T10:30:00Z",
  "verification_proofs": [
    {
      "proof_type": "facial_embedding_match",
      "content_hash": "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
      "score": 94,
      "passed": true,
      "threshold": 80,
      "timestamp": "2024-01-15T10:29:55Z"
    },
    {
      "proof_type": "liveness_score",
      "content_hash": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "score": 91,
      "passed": true,
      "threshold": 85,
      "timestamp": "2024-01-15T10:29:57Z"
    }
  ],
  "score": 92,
  "confidence": 95,
  "model_version": "facial-v2.3.0",
  "metadata": {
    "pipeline_version": "1.0.0",
    "validator_consensus": "5/5"
  },
  "proof": {
    "type": "Ed25519Signature2020",
    "created": "2024-01-15T10:30:00Z",
    "verification_method": "did:virtengine:validator:virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu#validator-key-001",
    "proof_purpose": "assertionMethod",
    "proof_value": "z3MvGcY3pFmYBDPKcSQr8TnqNkQu3yqNqwBmJCHuYU2JqcXNLHw2KwgEv7mFdJx2cFmTMxY4JmNvBpQnE9E3Kw7v",
    "nonce": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
    "domain": "veid.virtengine.io"
  }
}
```

### Document Verification Attestation

```json
{
  "id": "veid:attestation:abcdef1234567890:b2c3d4e5f6a1b2c3",
  "schema_version": "1.0.0",
  "type": "document_verification",
  "issuer": {
    "id": "did:virtengine:validator:virtengine1validator123456789...",
    "key_fingerprint": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
    "validator_address": "virtengine1validator123456789..."
  },
  "subject": {
    "id": "did:virtengine:virtengine1user456...",
    "account_address": "virtengine1user456...",
    "scope_id": "scope-document-001"
  },
  "nonce": "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
  "issued_at": "2024-01-15T11:00:00Z",
  "expires_at": "2025-01-15T11:00:00Z",
  "verification_proofs": [
    {
      "proof_type": "document_authenticity",
      "content_hash": "sha256:doc123...",
      "score": 90,
      "passed": true,
      "threshold": 80,
      "timestamp": "2024-01-15T10:59:50Z"
    },
    {
      "proof_type": "ocr_extraction",
      "content_hash": "sha256:ocr456...",
      "score": 95,
      "passed": true,
      "threshold": 90,
      "timestamp": "2024-01-15T10:59:52Z"
    },
    {
      "proof_type": "document_face_match",
      "content_hash": "sha256:face789...",
      "score": 82,
      "passed": true,
      "threshold": 75,
      "timestamp": "2024-01-15T10:59:55Z"
    }
  ],
  "score": 88,
  "confidence": 85,
  "proof": {
    "type": "EcdsaSecp256k1Signature2019",
    "created": "2024-01-15T11:00:00Z",
    "verification_method": "did:virtengine:validator:virtengine1validator123456789...#keys-1",
    "proof_purpose": "assertionMethod",
    "proof_value": "...",
    "nonce": "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3"
  }
}
```

---

## Security Considerations

1. **Key Storage**: Signer keys must be stored in hardware security modules (HSMs) or secure enclaves
2. **Nonce Generation**: Nonces must be generated using cryptographically secure random number generators
3. **Timestamp Validation**: Always validate timestamps to prevent replay attacks
4. **Key Rotation**: Rotate keys regularly and immediately upon compromise
5. **Audit Trail**: Maintain complete audit logs for compliance and forensics
6. **Rate Limiting**: Implement rate limiting to prevent abuse
7. **Input Validation**: Validate all input parameters before processing

---

## References

- [W3C Verifiable Credentials Data Model 1.1](https://www.w3.org/TR/vc-data-model/)
- [RFC 8017 - PKCS #1: RSA Cryptography Specifications](https://tools.ietf.org/html/rfc8017)
- [RFC 8032 - Edwards-Curve Digital Signature Algorithm (EdDSA)](https://tools.ietf.org/html/rfc8032)
- [VirtEngine VEID Flow Specification](_docs/veid-flow-spec.md)
