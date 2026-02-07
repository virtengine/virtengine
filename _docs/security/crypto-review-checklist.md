# Cryptographic Security Review Checklist

## Task Reference: 22A — Pre-mainnet Security Hardening

**Last Reviewed:** Pre-mainnet
**Status:** In Progress
**Reviewer:** Automated + Manual Review Required

---

## 1. Signature Verification (`x/veid/keeper/signature_crypto.go`)

### Ed25519 Signatures (Client Signatures)

| Check | Status | Notes |
|-------|--------|-------|
| Key size validated (32 bytes) | ✅ Pass | `Ed25519PublicKeySize = ed25519.PublicKeySize` |
| Uses `ed25519.Verify()` from Go stdlib | ✅ Pass | Constant-time comparison |
| Public key validated before use | ✅ Pass | Length check at L71 |
| No key reuse across algorithms | ✅ Pass | Separate validation paths |
| Signature malleability addressed | ✅ Pass | Go's ed25519 uses RFC 8032 |

### Secp256k1 Signatures (User/Cosmos Signatures)

| Check | Status | Notes |
|-------|--------|-------|
| Compressed public key size validated (33 bytes) | ✅ Pass | `Secp256k1PublicKeySize = 33` |
| Uses Cosmos SDK crypto primitives | ✅ Pass | `secp256k1.PubKey` |
| Signature format validated | ✅ Pass | DER decoding handled |
| No low-S malleability | ⚠️ Review | Cosmos SDK handles this internally |

### Salt Binding

| Check | Status | Notes |
|-------|--------|-------|
| Salt uniqueness enforced | ✅ Pass | `checkSaltUnused()` via store lookup |
| Salt size validated | ✅ Pass | Now enforced via `MaxSaltSize` (128 bytes) |
| Replay prevention via salt | ✅ Pass | Salt stored after use |
| Composite signature binding | ✅ Pass | `sha256(salt|clientID|payloadHash)` |

---

## 2. ZK Proof System (`x/veid/keeper/zkproofs_circuits.go`)

### Groth16 Setup

| Check | Status | Notes |
|-------|--------|-------|
| Trusted setup ceremony | ✅ Implemented | `trusted_setup.go` — MPC ceremony tooling |
| Minimum 3 participants | ✅ Enforced | `MinCeremonyParticipants = 3` |
| Contribution chain integrity | ✅ Verified | Hash chain validation in `CompleteCeremony()` |
| Verification key registry | ✅ Implemented | `VerificationKeyRecord` with ceremony linkage |
| Circuit-specific ceremonies | ✅ Pass | Separate ceremonies per circuit type |

### Circuit Security

| Check | Status | Notes |
|-------|--------|-------|
| BN254 curve (~100-bit security) | ⚠️ Acceptable | Standard for Groth16, sufficient for identity |
| Age range circuit constraints (~500) | ✅ Pass | Proper range proof |
| Residency circuit constraints (~400) | ✅ Pass | Proper membership proof |
| Score range circuit constraints (~300) | ✅ Pass | Proper range proof |
| Proof verification is deterministic | ✅ Pass | gnark verifier is deterministic |
| No proof malleability | ✅ Pass | Groth16 proofs are non-malleable |

### Determinism

| Check | Status | Notes |
|-------|--------|-------|
| Proof generation off-chain only | ✅ Pass | Verification only on-chain |
| Deterministic verification | ✅ Pass | Same inputs → same result |
| No floating point in verification | ✅ Pass | Field arithmetic only |

---

## 3. Encryption (`x/veid/types/security_controls.go`)

### Envelope Encryption

| Check | Status | Notes |
|-------|--------|-------|
| X25519-XSalsa20-Poly1305 | ✅ Pass | NaCl box construction |
| Nonce uniqueness | ✅ Pass | Random nonce per envelope |
| Authenticated encryption | ✅ Pass | Poly1305 MAC |
| Key derivation | ✅ Pass | X25519 ECDH |
| No plaintext storage | ✅ Pass | Encrypted at rest |

### Tokenization & Pseudonymization

| Check | Status | Notes |
|-------|--------|-------|
| Deterministic tokenization | ✅ Pass | HMAC-SHA256 based |
| Token-to-data unlinkability | ✅ Pass | Requires HMAC key |
| Pseudonym generation | ✅ Pass | SHA-256 with domain separation |

---

## 4. Key Management

### Approved Client Keys

| Check | Status | Notes |
|-------|--------|-------|
| Key rotation mechanism | ✅ Implemented | `key_rotation.go` — overlap period rotation |
| Overlap period for continuity | ✅ Pass | `DefaultKeyRotationOverlapBlocks = 17280` (~1 day) |
| Maximum overlap bounded | ✅ Pass | `MaxKeyRotationOverlapBlocks = 120960` (~7 days) |
| Governance-gated rotation | ✅ Pass | Authority check required |
| Auto-completion at expiry | ✅ Pass | `ProcessExpiredKeyRotations()` in EndBlock |
| No concurrent rotations | ✅ Pass | Active rotation check before initiation |

### HSM Support

| Check | Status | Notes |
|-------|--------|-------|
| HSM keyring wrapper exists | ✅ Pass | `pkg/keymanagement/keyring_hsm.go` |
| Hardware key storage | ✅ Pass | Delegated to HSM |
| Key never in plaintext memory | ⚠️ Partial | HSM operations keep key in hardware |

---

## 5. Input Validation (22A-AC4)

### Message Handler Audit

| Handler | Address Validation | Size Limits | Rate Limited |
|---------|-------------------|-------------|-------------|
| `UploadScope` | ✅ `AccAddressFromBech32` | ✅ All fields | ✅ Per-account + per-block |
| `RevokeScope` | ✅ `AccAddressFromBech32` | ✅ Reason field | ❌ Low risk |
| `RequestVerification` | ✅ `AccAddressFromBech32` | ✅ Scope ID | ✅ Per-account |
| `UpdateVerificationStatus` | ✅ `AccAddressFromBech32` | ✅ Scope ID + reason | ❌ Validator-gated |
| `UpdateScore` | ✅ `AccAddressFromBech32` | ✅ Score bounds (0-1000) | ✅ Per-block |
| `UpdateParams` | ✅ Authority check | N/A | ❌ Governance-gated |
| `CreateIdentityWallet` | ✅ `AccAddressFromBech32` | ⚠️ Review | ❌ Low frequency |

### Size Limits Enforced

| Field | Max Size | Rationale |
|-------|----------|-----------|
| `scope_id` | 128 bytes | UUID-like identifier |
| `reason` | 512 bytes | Human-readable text |
| `client_id` | 64 bytes | Short identifier |
| `device_fingerprint` | 256 bytes | Hash-based fingerprint |
| `salt` | 128 bytes | Cryptographic salt |
| `signature` | 512 bytes | Ed25519/secp256k1 |
| `payload_hash` | 64 bytes | SHA-256 hex |
| `geo_hint` | 128 bytes | Country/region code |
| `purpose` | 256 bytes | Consent purpose |

---

## 6. Rate Limiting (22A-AC5)

| Operation | Per-Account Per-Block | Per-Block Global | Cooldown |
|-----------|----------------------|-----------------|----------|
| `UploadScope` | 3 | 50 | 2 blocks |
| `RequestVerification` | 5 | N/A | N/A |
| `UpdateScore` | N/A | 100 | N/A |

---

## 7. Privilege Escalation Paths (22A-AC6)

### Identified Privilege Levels

| Level | Required By | Enforcement |
|-------|------------|-------------|
| Governance Authority | `UpdateParams`, `UpdateBorderlineParams`, Key Rotation, Ceremony | `sender == k.authority` |
| Bonded Validator | `UpdateVerificationStatus`, `UpdateScore` | `IsValidator()` → `stakingKeeper.GetValidator()` + `IsBonded()` |
| Any Account | `UploadScope`, `RevokeScope`, `CreateIdentityWallet` | Address validation only |

### Escalation Prevention

| Check | Status | Notes |
|-------|--------|-------|
| Validator check uses staking keeper | ✅ Pass | Queries live validator set |
| Nil staking keeper → DENY | ✅ Pass | Safe default at L842 |
| Authority is module account | ✅ Pass | Set to x/gov module account in `NewKeeper()` |
| No hardcoded addresses | ✅ Pass | Authority from app wiring |
| Privilege audit logging | ✅ Implemented | `ValidatePrivilegedOperation()` with store records |

---

## 8. Recommendations

### Critical (Must Fix Before Mainnet)
1. ✅ **Trusted setup ceremony** — Implemented MPC ceremony tooling
2. ✅ **Key rotation for approved clients** — Implemented with overlap periods
3. ✅ **Rate limiting on msg handlers** — Per-account and per-block limits

### High Priority
4. ⚠️ **External audit of ZK circuits** — Formal verification of constraint systems recommended
5. ⚠️ **Secp256k1 low-S check** — Verify Cosmos SDK handles signature malleability
6. ⚠️ **HSM key zeroization** — Verify memory is cleared after HSM operations

### Medium Priority
7. 📋 **Rate limit parameter governance** — Make limits configurable via params
8. 📋 **Ceremony expiry cleanup** — Add EndBlock cleanup for expired ceremonies
9. 📋 **Key rotation history** — Store completed rotations for audit trail

### Low Priority
10. 📋 **BN254 to BLS12-381 migration path** — Higher security margin if needed
11. 📋 **Rate limit telemetry** — Emit events for rate limit hits
