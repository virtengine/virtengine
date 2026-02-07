# Data Vault Package

Unified, off-chain encrypted data vault service for VirtEngine sensitive payloads.

## Overview

The data vault implements secure storage for all sensitive payloads referenced by on-chain records:
- VEID identity documents and attestations
- Support ticket attachments
- Marketplace deployment artifacts
- Audit logs and compliance artifacts

## Architecture

### Components

- **VaultService**: Main service interface for encrypted blob CRUD operations
- **EncryptedBlobStore**: Storage layer wrapping `artifact_store` with encryption
- **KeyManager**: DEK/KEK hierarchy with rotation support
- **AccessControl**: Wallet-scoped auth + role/org permissions
- **AuditLogger**: Immutable audit trail for compliance

### Security Properties

- All sensitive data encrypted at rest with unique DEKs (Data Encryption Keys)
- DEKs wrapped with KEKs (Key Encryption Keys) managed by KeyManager
- On-chain references contain only blob IDs + content hashes (SHA-256)
- Cross-org access strictly denied
- Every decrypt operation logged with full context

### Encryption

- **Algorithm**: X25519-XSalsa20-Poly1305 (NaCl box)
- **Envelope format**: `EncryptedPayloadEnvelope` from `x/encryption/types`
- **Key hierarchy**: Scope-specific KEKs, per-blob DEKs
- **Key rotation**: Backward-compatible decryption with overlap periods

## Usage

### Initialize Key Manager

```go
keyMgr := keys.NewKeyManager()
if err := keyMgr.Initialize(); err != nil {
    return err
}
```

### Create Encrypted Blob Store

```go
backend := // Initialize artifact_store backend (filesystem, IPFS, Waldur)
store := NewEncryptedBlobStore(backend, keyMgr)
```

### Upload Encrypted Blob

```go
req := &UploadRequest{
    Scope:     ScopeVEID,
    Plaintext: sensitiveData,
    Owner:     "virtengine1abc...",
    OrgID:     "org-123",
    Tags:      map[string]string{"type": "passport"},
}

blob, err := store.Store(ctx, req)
if err != nil {
    return err
}

// Store blob.Metadata.ID on-chain
```

### Retrieve and Decrypt Blob

```go
plaintext, metadata, err := store.Retrieve(ctx, blobID)
if err != nil {
    return err
}

// Verify content hash matches on-chain reference
```

### Portal API (Provider Daemon)

The provider daemon exposes HTTP endpoints under `/api/v1/vault/*` for upload, retrieve,
metadata, deletion, and audit search. All requests require wallet-signed or HMAC auth.

### Key Rotation

```go
// Initiate rotation with 24-hour overlap period
err := keyMgr.RotateKey(keys.ScopeVEID, 24*time.Hour)
if err != nil {
    return err
}

// During overlap: new keys used for encryption, old keys still work for decryption

// Complete rotation after overlap expires
err = keyMgr.CompleteRotation(keys.ScopeVEID)
if err != nil {
    return err
}
```

## Scopes

| Scope          | Description                                  |
|----------------|----------------------------------------------|
| `ScopeVEID`    | VEID identity documents and attestations    |
| `ScopeSupport` | Support ticket attachments                   |
| `ScopeMarket`  | Marketplace deployment artifacts             |
| `ScopeAudit`   | Audit logs and compliance artifacts          |

## Key Rotation

### Rotation Lifecycle

1. **Active**: Key used for encryption and decryption
2. **Deprecated**: Key no longer used for encryption, but still decrypts
3. **Retired**: Key no longer used (historical data only)

### Rotation Policy

```go
policy := &keys.RotationPolicy{
    MaxAge:           90 * 24 * time.Hour, // 90 days
    MaxVersions:      5,
    AutoRotate:       true,
    RotationSchedule: "0 0 * * 0", // Weekly on Sunday
}
keyMgr.SetRotationPolicy(keys.ScopeVEID, policy)
```

### Backward Compatibility

The KeyManager maintains all key versions, ensuring:
- Data encrypted with old keys can still be decrypted
- Supports 2+ historical key versions
- Lazy re-encryption on access (optional)

## Testing

```bash
# Run all tests
go test ./...

# Run key manager tests
go test ./keys -v

# Run store tests
go test -run TestEncryptedBlobStore
```

## Status

### Implemented (Phase 1)

- [x] Core types and interfaces
- [x] KeyManager with rotation support
- [x] EncryptedBlobStore basic implementation
- [x] Key manager tests (8 tests passing)

### TODO (Remaining)

- [x] Access control integration (wallet auth, org isolation)
- [x] Audit logging for all decrypt operations
- [x] Migration helpers for existing plaintext storage
- [ ] Streaming support for large blobs (>10MB)
- [x] API endpoints (HTTP)
- [x] Provider daemon integration
- [ ] HSM integration for KEK storage (task 30B)
- [ ] E2E tests
- [x] Documentation and runbooks

## Related

- `pkg/artifact_store`: Backend storage abstraction
- `x/encryption`: Encryption primitives and envelope format
- Task 32A: Encrypted data vault + key rotation
- Task 30B: HSM integration
- AU2024203136A1-LIVE: Regulatory compliance requirements
