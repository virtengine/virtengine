# Mobile Capture Protocol v1

**Version:** 1.0.0  
**Status:** Active  
**Last Updated:** 2026-01-24  
**Related Tasks:** VE-207, VE-201, VE-210

## Overview

The Mobile Capture Protocol defines a secure method for uploading identity documents and selfies from approved mobile clients to the VirtEngine blockchain. This protocol ensures:

1. **Authenticity** - Uploads originate from approved capture interfaces
2. **Integrity** - Data has not been tampered with during transmission
3. **Non-repudiation** - Dual signatures prove both client and user consent
4. **Freshness** - Salt binding and timestamps prevent replay attacks
5. **Anti-gallery** - Measures prevent uploading pre-existing images

## Protocol Version

```
Protocol:        VirtEngine Mobile Capture Protocol
Version:         1.0.0
Identifier:      VMCP/1.0
Minimum Client:  1.0.0
```

## Security Goals

| Goal | Mechanism |
|------|-----------|
| Prevent replay attacks | Per-upload salt + timestamp binding |
| Prevent gallery uploads | Device/session binding + timestamp freshness |
| Verify client authenticity | Approved client signature + allowlist |
| Verify user consent | User account signature chain |
| Ensure data integrity | SHA-256 payload hash |
| Enable key rotation | Multi-key support with deprecation windows |

---

## Salt Generation Requirements

### Salt Properties

| Property | Requirement |
|----------|-------------|
| Length | 32 bytes (256 bits) minimum |
| Entropy | Cryptographically secure random |
| Uniqueness | MUST be unique per upload |
| Freshness | Generated at capture time, not upload time |

### Salt Generation Algorithm

```
function generateSalt():
    1. random_bytes = CSPRNG(32)          // 32 cryptographically secure random bytes
    2. timestamp = current_unix_time_ms()  // Millisecond precision
    3. timestamp_bytes = int64_to_bytes(timestamp)
    4. salt = XOR(random_bytes[0:8], timestamp_bytes) || random_bytes[8:32]
    5. return salt
```

### Salt Binding

The salt MUST be cryptographically bound to:

1. **Device Identifier** - Hardware/software fingerprint
2. **Session Identifier** - Unique capture session ID
3. **Timestamp** - Unix timestamp of binding creation

```
SaltBinding = {
    salt:           bytes[32],      // The generated salt
    device_id:      string,         // Device fingerprint hash
    session_id:     string,         // UUID of capture session
    timestamp:      int64,          // Unix timestamp (seconds)
    binding_hash:   bytes[32]       // SHA256(salt || device_id || session_id || timestamp)
}
```

### Binding Hash Computation

```
binding_hash = SHA256(
    salt ||
    UTF8_ENCODE(device_id) ||
    UTF8_ENCODE(session_id) ||
    INT64_TO_BYTES(timestamp)
)
```

---

## Signature Scheme

### Dual Signature Architecture

The protocol requires two signatures in a chain:

```
┌─────────────────────────────────────────────────────────────┐
│                     SIGNATURE CHAIN                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Client Signature (Approved App)                         │
│     signed_data = salt || payload_hash                      │
│     client_sig = SIGN(client_private_key, signed_data)      │
│                                                             │
│  2. User Signature (Account Holder)                         │
│     signed_data = salt || payload_hash || client_sig        │
│     user_sig = SIGN(user_private_key, signed_data)          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Supported Algorithms

| Algorithm | Key Size | Use Case |
|-----------|----------|----------|
| Ed25519 | 256-bit | Preferred for clients |
| secp256k1 | 256-bit | Compatible with Cosmos accounts |

### Signature Proof Structure

```
SignatureProof = {
    public_key:     bytes,          // Signer's public key
    signature:      bytes,          // The signature bytes
    algorithm:      string,         // "ed25519" or "secp256k1"
    key_id:         string,         // Client ID or account address
    signed_data:    bytes           // Data that was signed
}
```

### Client Signature Requirements

1. Client MUST be registered in the approved client allowlist
2. Client key MUST match the registered public key
3. Signature MUST cover: `salt || payload_hash`
4. Algorithm MUST match client's registered algorithm

### User Signature Requirements

1. User MUST have a valid VirtEngine account
2. Signature MUST cover: `salt || payload_hash || client_signature`
3. This creates a signature chain proving user saw client's attestation

---

## Payload Structure

### CapturePayload

```
CapturePayload = {
    // Protocol identification
    version:            uint32,             // Protocol version (1)
    
    // Core content
    payload_hash:       bytes[32],          // SHA256 of encrypted content
    encrypted_content:  bytes,              // The actual encrypted data
    
    // Salt binding
    salt:               bytes[32],          // Per-upload salt
    salt_binding:       SaltBinding,        // Full binding structure
    
    // Signatures
    client_signature:   SignatureProof,     // Approved client signature
    user_signature:     SignatureProof,     // User account signature
    
    // Metadata
    capture_metadata:   CaptureMetadata,    // Capture context
    timestamp:          int64               // Upload timestamp
}
```

### CaptureMetadata

```
CaptureMetadata = {
    device_fingerprint: string,     // Device identifier hash
    client_id:          string,     // Approved client identifier
    client_version:     string,     // Client app version
    session_id:         string,     // Capture session UUID
    document_type:      string,     // "id_card" | "passport" | "selfie"
    quality_score:      uint32,     // Quality validation score (0-100)
    capture_timestamp:  int64,      // When image was captured
    geo_hint:           string      // Optional: country code
}
```

### Payload Hash Computation

```
payload_hash = SHA256(
    encrypted_content ||
    UTF8_ENCODE(JSON.stringify(capture_metadata))
)
```

---

## Anti-Replay Mechanisms

### 1. Salt Uniqueness

- Each salt MUST be globally unique
- Server maintains a cache of recently used salts
- Duplicate salt submission is rejected immediately

### 2. Timestamp Windows

| Window | Duration | Purpose |
|--------|----------|---------|
| Salt Age | 5 minutes | Maximum time from salt generation to upload |
| Replay Window | 10 minutes | Duration salts are cached for replay detection |
| Clock Skew | ±30 seconds | Allowed difference from server time |

### 3. Salt Cache Strategy

```
SaltCache = {
    storage:        LRU_Cache,      // Least Recently Used eviction
    max_entries:    1,000,000,      // Maximum cached salts
    ttl:            10 minutes,     // Time-to-live per entry
    key:            SHA256(salt),   // Lookup key (hashed salt)
    value:          timestamp       // When salt was used
}
```

### 4. Replay Detection Flow

```
function checkNotReplayed(payload):
    salt_hash = SHA256(payload.salt)
    
    // Check if salt was already used
    if cache.exists(salt_hash):
        return ERROR("Salt already used - replay detected")
    
    // Check timestamp freshness
    age = current_time - payload.salt_binding.timestamp
    if age > MAX_SALT_AGE:
        return ERROR("Salt too old - expired")
    
    if age < -MAX_CLOCK_SKEW:
        return ERROR("Salt from future - clock skew exceeded")
    
    // Record salt as used
    cache.set(salt_hash, current_time, ttl=REPLAY_WINDOW)
    
    return OK
```

### 5. Anti-Gallery Measures

To prevent uploading images from device gallery:

1. **Capture Timestamp Binding** - Salt is generated at capture time
2. **Session Binding** - Session ID ties capture to active session
3. **Device Binding** - Device fingerprint prevents cross-device replay
4. **Quality Checks** - Live capture quality differs from gallery photos
5. **Liveness Detection** - For selfies, requires active challenges

---

## Key Rotation Strategy

### Overview

Client keys can be rotated without disrupting service through a managed deprecation process.

### Key States

```
┌──────────────┐     ┌────────────────┐     ┌───────────────┐     ┌──────────┐
│   PENDING    │────►│     ACTIVE     │────►│  DEPRECATED   │────►│  REVOKED │
└──────────────┘     └────────────────┘     └───────────────┘     └──────────┘
     New key           Valid for            Valid but              Invalid,
     registered        all operations       overlap only           rejected
```

### Rotation Timeline

```
Day 0:  New key registered via governance (state: PENDING)
Day 1:  New key activated (state: ACTIVE)
        Old key marked deprecated (state: DEPRECATED)
Day 7:  Overlap period - both keys valid
Day 8:  Old key revoked (state: REVOKED)
```

### Key Registration Process

1. **Proposal** - Governance proposal to add new client key
2. **Voting** - Token holders vote on key registration
3. **Activation** - Upon passing, key becomes active after activation delay
4. **Overlap** - Previous key remains valid during overlap period
5. **Revocation** - Old key automatically revoked after overlap

### Multi-Key Validation

During overlap periods, signature validation accepts either:
- Current active key, OR
- Previous deprecated (but not yet revoked) key

```go
func validateClientKey(clientID string, publicKey []byte) error {
    client := registry.GetClient(clientID)
    
    // Check active key
    if bytes.Equal(client.ActiveKey, publicKey) {
        return nil
    }
    
    // Check deprecated key (if in overlap period)
    if client.DeprecatedKey != nil && client.DeprecatedKeyExpiry.After(time.Now()) {
        if bytes.Equal(client.DeprecatedKey, publicKey) {
            return nil
        }
    }
    
    return ErrInvalidClientKey
}
```

### Key Revocation

Keys are revoked immediately in case of:
- Compromise detection
- Client deactivation
- Security incident

Emergency revocation bypasses normal overlap period via governance emergency action.

---

## Validation Rules

### Server-Side Validation Sequence

```
1. PARSE payload structure
   └─► Reject malformed payloads

2. VERIFY protocol version
   └─► Reject unsupported versions

3. VALIDATE salt binding
   ├─► Check salt length (≥32 bytes)
   ├─► Verify binding hash
   ├─► Check timestamp freshness
   └─► Check salt not replayed

4. VALIDATE client signature
   ├─► Lookup client in approved list
   ├─► Verify client is active
   ├─► Verify public key matches
   └─► Verify signature

5. VALIDATE user signature
   ├─► Verify account exists
   ├─► Verify signature chain
   └─► Verify signature

6. RECORD salt as used
   └─► Prevent future replays

7. ACCEPT payload for processing
```

### Error Responses

| Code | Error | Description |
|------|-------|-------------|
| 1001 | INVALID_SALT | Salt too short or malformed |
| 1002 | SALT_EXPIRED | Salt timestamp too old |
| 1003 | SALT_REPLAYED | Salt was already used |
| 1004 | BINDING_MISMATCH | Salt binding hash invalid |
| 1005 | CLIENT_NOT_APPROVED | Client ID not in allowlist |
| 1006 | CLIENT_KEY_INVALID | Client public key mismatch |
| 1007 | CLIENT_SIG_INVALID | Client signature verification failed |
| 1008 | USER_SIG_INVALID | User signature verification failed |
| 1009 | TIMESTAMP_FUTURE | Timestamp too far in future |
| 1010 | PROTOCOL_VERSION | Unsupported protocol version |

---

## Implementation Notes

### Client Implementation (TypeScript)

See `lib/capture/` for reference implementation:
- `utils/salt-generator.ts` - Salt generation
- `utils/signature.ts` - Signature creation
- `types/capture.ts` - Type definitions

### Server Implementation (Go)

See `pkg/capture_protocol/` for validation implementation:
- `types.go` - Protocol types
- `salt.go` - Salt validation
- `signature.go` - Signature validation
- `validator.go` - Full protocol validator

### Chain Integration

The `x/veid` module integrates with this protocol:
- `types/upload.go` - Upload metadata types
- `keeper/msg_server.go` - Upload message handling

---

## Security Considerations

### Threat Mitigations

| Threat | Mitigation |
|--------|------------|
| Replay attack | Unique salt + cache |
| Gallery upload | Timestamp + session binding |
| Rogue client | Approved client allowlist |
| Man-in-the-middle | Dual signature chain |
| Key compromise | Key rotation + revocation |
| Clock manipulation | Server-side timestamp validation |

### Audit Logging

All validation failures SHOULD be logged with:
- Client ID
- User address (if available)
- Failure reason
- Timestamp
- Salt hash (for correlation)

---

## Changelog

### v1.0.0 (2026-01-24)
- Initial protocol specification
- Salt binding requirements
- Dual signature scheme
- Anti-replay mechanisms
- Key rotation strategy
