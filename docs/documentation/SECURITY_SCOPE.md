# VirtEngine Security Audit Scope

**Version:** 1.0.0  
**Date:** 2026-01-29  
**Prepared For:** External Security Audit  
**Status:** Ready for Audit

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Modules In Scope](#modules-in-scope)
3. [Assets to Protect](#assets-to-protect)
4. [Attack Surfaces](#attack-surfaces)
5. [Cryptographic Operations](#cryptographic-operations)
6. [Authentication & Authorization Flows](#authentication--authorization-flows)
7. [External Trust Boundaries](#external-trust-boundaries)
8. [Out of Scope](#out-of-scope)
9. [Known Issues](#known-issues)
10. [Audit Questionnaire Responses](#audit-questionnaire-responses)

---

## Executive Summary

VirtEngine is a Cosmos SDK-based blockchain for decentralized cloud computing with ML-powered identity verification (VEID). This document defines the scope for a comprehensive security audit of the platform's core security-critical components.

### Audit Objectives

1. Validate cryptographic implementations are correct and secure
2. Verify consensus-safety of state machine transitions
3. Assess identity verification security model
4. Review authentication and authorization controls
5. Identify vulnerabilities in key management
6. Evaluate threat mitigations effectiveness

### Primary Security Concerns

| Concern                       | Priority | Relevant Components                       |
| ----------------------------- | -------- | ----------------------------------------- |
| Identity data confidentiality | CRITICAL | x/veid, x/encryption, pkg/enclave_runtime |
| Cryptographic correctness     | CRITICAL | x/encryption/crypto, pkg/capture_protocol |
| Consensus safety              | CRITICAL | All x/ modules, ante handlers             |
| Access control bypass         | HIGH     | x/roles, x/mfa, ante handlers             |
| Key management security       | HIGH     | x/encryption, pkg/provider_daemon         |
| Replay attack prevention      | HIGH     | x/veid, pkg/capture_protocol              |

---

## Modules In Scope

### Tier 1: Security-Critical (Full Audit Required)

| Module               | Path                    | Description                                                | LOC    |
| -------------------- | ----------------------- | ---------------------------------------------------------- | ------ |
| **veid**             | `x/veid/`               | Identity verification with encrypted scopes and ML scoring | ~4,500 |
| **encryption**       | `x/encryption/`         | Public-key encryption (X25519-XSalsa20-Poly1305)           | ~2,800 |
| **mfa**              | `x/mfa/`                | Multi-factor authentication gating                         | ~1,800 |
| **roles**            | `x/roles/`              | Role-based access control                                  | ~1,500 |
| **enclave_runtime**  | `pkg/enclave_runtime/`  | TEE runtime for secure processing                          | ~1,200 |
| **capture_protocol** | `pkg/capture_protocol/` | Salt-binding and signature verification                    | ~1,400 |

### Tier 2: Security-Relevant (Review Required)

| Module              | Path                   | Description                        | LOC    |
| ------------------- | ---------------------- | ---------------------------------- | ------ |
| **cert**            | `x/cert/`              | Certificate management             | ~1,200 |
| **escrow**          | `x/escrow/`            | Payment escrow with settlement     | ~2,000 |
| **market**          | `x/market/`            | Marketplace orders and bids        | ~3,500 |
| **provider_daemon** | `pkg/provider_daemon/` | Off-chain bidding and provisioning | ~5,000 |
| **ante handlers**   | `app/ante*.go`         | Transaction preprocessing          | ~800   |

### Tier 3: Supporting (Spot Check)

| Module         | Path             | Description                   |
| -------------- | ---------------- | ----------------------------- |
| **audit**      | `x/audit/`       | Provider attribute auditing   |
| **config**     | `x/config/`      | Approved client configuration |
| **settlement** | `x/settlement/`  | Payment settlement processing |
| **inference**  | `pkg/inference/` | Deterministic ML scoring      |

---

## Assets to Protect

### Critical Assets

| Asset                     | Classification | Storage Location     | Protection Mechanism              |
| ------------------------- | -------------- | -------------------- | --------------------------------- |
| Identity document images  | PII/CRITICAL   | On-chain (encrypted) | X25519-XSalsa20-Poly1305 envelope |
| Facial embeddings         | PII/CRITICAL   | On-chain (encrypted) | Validator threshold encryption    |
| Face images for liveness  | PII/CRITICAL   | On-chain (encrypted) | Per-upload encryption envelope    |
| User private keys         | SECRET         | Client-side          | Never transmitted, wallet-managed |
| Validator decryption keys | SECRET         | Validator nodes      | HSM/TEE sealed storage            |
| MFA secrets (TOTP seeds)  | SECRET         | On-chain (encrypted) | Per-user encryption               |
| Consent records           | PII            | On-chain             | Signed, immutable audit trail     |

### Sensitive Assets

| Asset                 | Classification | Storage Location | Protection Mechanism         |
| --------------------- | -------------- | ---------------- | ---------------------------- |
| User wallet addresses | INTERNAL       | On-chain         | Public but pseudonymous      |
| Verification scores   | INTERNAL       | On-chain         | Public after verification    |
| Provider credentials  | SECRET         | Provider daemon  | Sealed storage, key rotation |
| Session tokens        | SECRET         | Memory only      | Short TTL, secure deletion   |

### Integrity-Critical Assets

| Asset                | Description                 | Integrity Control               |
| -------------------- | --------------------------- | ------------------------------- |
| Genesis state        | Initial chain configuration | Multi-sig governance            |
| Module parameters    | Runtime configuration       | x/gov proposals only            |
| Approved client list | Trusted capture apps        | Governance-controlled allowlist |
| ML model versions    | Identity scoring models     | Pinned in chain params          |

---

## Attack Surfaces

### 1. gRPC/REST API Endpoints

| Endpoint Category       | Risk Level | Key Controls                                           |
| ----------------------- | ---------- | ------------------------------------------------------ |
| `veid.MsgSubmitScope`   | CRITICAL   | Salt-binding, approved-client signature, rate limiting |
| `veid.MsgVerifyScope`   | HIGH       | Validator-only, threshold consensus                    |
| `mfa.MsgRegisterDevice` | HIGH       | Account ownership, device attestation                  |
| `mfa.MsgVerify`         | HIGH       | Rate limiting, attempt tracking, lockout               |
| `encryption.MsgEncrypt` | MEDIUM     | Key validation, envelope format                        |
| `market.MsgCreateOrder` | MEDIUM     | Authorization, escrow locking                          |

### 2. Ante Handlers (Transaction Preprocessing)

| Handler                    | File                    | Security Function                       |
| -------------------------- | ----------------------- | --------------------------------------- |
| `MFADecorator`             | `app/ante_mfa.go`       | Enforces MFA for sensitive transactions |
| `RateLimitDecorator`       | `app/ante_ratelimit.go` | Prevents transaction flooding           |
| `SigVerificationDecorator` | `app/ante.go` (SDK)     | Cryptographic signature verification    |
| `DeductFeeDecorator`       | `app/ante.go` (SDK)     | Gas fee enforcement                     |

### 3. Cryptographic Operations

| Operation              | Implementation                      | Risk                         |
| ---------------------- | ----------------------------------- | ---------------------------- |
| Envelope encryption    | `x/encryption/crypto/envelope.go`   | Key confusion, nonce reuse   |
| Signature verification | `pkg/capture_protocol/signature.go` | Forgery, replay              |
| Key derivation         | `x/encryption/crypto/mnemonic.go`   | Weak entropy, predictability |
| Hash computation       | `x/encryption/types/fingerprint.go` | Collision, length extension  |

### 4. State Machine Transitions

| State Machine       | Location                       | Critical Transitions                  |
| ------------------- | ------------------------------ | ------------------------------------- |
| Verification status | `x/veid/types/verification.go` | pending→verified, expired transitions |
| MFA challenge       | `x/mfa/types/challenge.go`     | challenge→verified, lockout           |
| Order lifecycle     | `x/market/types/order.go`      | open→matched→closed                   |
| Escrow settlement   | `x/escrow/types/escrow.go`     | active→settled, dispute               |

### 5. Inter-Module Calls

| Caller   | Callee       | Trust Boundary         |
| -------- | ------------ | ---------------------- |
| x/veid   | x/encryption | Decrypt identity data  |
| x/veid   | x/mfa        | Check MFA requirement  |
| x/market | x/escrow     | Lock/release funds     |
| x/mfa    | x/roles      | Check permission level |

---

## Cryptographic Operations

### Algorithms Used

| Algorithm         | Purpose                  | Implementation                    | Standard   |
| ----------------- | ------------------------ | --------------------------------- | ---------- |
| X25519            | Key exchange             | `golang.org/x/crypto/curve25519`  | RFC 7748   |
| XSalsa20-Poly1305 | Authenticated encryption | `golang.org/x/crypto/nacl/box`    | NaCl       |
| SHA-256           | Hashing, fingerprints    | `crypto/sha256`                   | FIPS 180-4 |
| Ed25519           | Signatures (Cosmos)      | `crypto/ed25519`                  | RFC 8032   |
| secp256k1         | Signatures (optional)    | `btcec`                           | SEC 2      |
| BIP-39            | Mnemonic generation      | `x/encryption/crypto/mnemonic.go` | BIP-39     |

### Encryption Envelope Format

```
┌─────────────────────────────────────────────────────────────────┐
│                    EncryptedPayloadEnvelope                      │
├─────────────────────────────────────────────────────────────────┤
│ Version:          uint32 (currently 1)                           │
│ AlgorithmID:      string ("X25519-XSalsa20-Poly1305")           │
│ AlgorithmVersion: string ("v1")                                  │
│ RecipientKeyIDs:  []string (SHA-256 fingerprints)               │
│ RecipientPubKeys: [][]byte (32-byte X25519 public keys)         │
│ EncryptedKeys:    [][]byte (for multi-recipient mode)           │
│ Nonce:            []byte (24 bytes, unique per encryption)      │
│ Ciphertext:       []byte (encrypted data + Poly1305 tag)        │
│ SenderPubKey:     []byte (32 bytes)                             │
│ SenderSignature:  []byte (envelope signature)                   │
│ Metadata:         map[string]string                              │
└─────────────────────────────────────────────────────────────────┘
```

### Key Management

| Key Type                  | Generation             | Storage        | Rotation             |
| ------------------------- | ---------------------- | -------------- | -------------------- |
| User wallet keys          | BIP-39 mnemonic        | Client wallet  | User-managed         |
| Validator consensus keys  | `virtengine init`      | Keyring/HSM    | Manual rotation      |
| Validator encryption keys | Key ceremony           | HSM/TEE sealed | Governance-triggered |
| Approved client keys      | Build pipeline         | Secure enclave | Per-release          |
| Provider daemon keys      | `provider-daemon init` | Keyring/Ledger | Operator-managed     |

### Nonce Management

| Context             | Nonce Source     | Size     | Uniqueness Guarantee      |
| ------------------- | ---------------- | -------- | ------------------------- |
| Envelope encryption | `crypto/rand`    | 24 bytes | Random per operation      |
| MFA challenges      | `crypto/rand`    | 32 bytes | Random per challenge      |
| Salt-binding        | `crypto/rand`    | 32 bytes | Bound to session + wallet |
| Transaction nonces  | Account sequence | uint64   | Cosmos SDK managed        |

---

## Authentication & Authorization Flows

### 1. Identity Verification Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   User      │    │  Approved   │    │  Blockchain │    │  Validators │
│   Device    │    │   Client    │    │    Node     │    │   (TEE)     │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                   │                  │
       │ 1. Initiate      │                   │                  │
       │    Verification  │                   │                  │
       │─────────────────>│                   │                  │
       │                  │                   │                  │
       │ 2. Request Salt  │ 3. MsgRequestSalt │                  │
       │<─────────────────│──────────────────>│                  │
       │                  │                   │                  │
       │ 4. Salt + Nonce  │                   │                  │
       │<─────────────────│<──────────────────│                  │
       │                  │                   │                  │
       │ 5. Capture       │                   │                  │
       │    Document +    │                   │                  │
       │    Selfie        │                   │                  │
       │─────────────────>│                   │                  │
       │                  │                   │                  │
       │                  │ 6. Sign with      │                  │
       │                  │    Client Key     │                  │
       │                  │    + User Key     │                  │
       │                  │    + Salt Binding │                  │
       │                  │                   │                  │
       │                  │ 7. MsgSubmitScope │                  │
       │                  │    (encrypted)    │                  │
       │                  │──────────────────>│                  │
       │                  │                   │                  │
       │                  │                   │ 8. Distribute    │
       │                  │                   │    to Validators │
       │                  │                   │─────────────────>│
       │                  │                   │                  │
       │                  │                   │ 9. TEE Decrypt   │
       │                  │                   │    + ML Score    │
       │                  │                   │<─────────────────│
       │                  │                   │                  │
       │                  │                   │ 10. Consensus    │
       │                  │                   │     Vote         │
       │                  │                   │<─────────────────│
       │                  │                   │                  │
       │ 11. Verification │                   │                  │
       │     Result       │                   │                  │
       │<─────────────────│<──────────────────│                  │
       │                  │                   │                  │
```

### 2. MFA Verification Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    User     │    │   Client    │    │  Blockchain │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                   │
       │ 1. Submit        │                   │
       │    Sensitive Tx  │                   │
       │─────────────────>│ 2. Broadcast Tx   │
       │                  │──────────────────>│
       │                  │                   │
       │                  │ 3. MFA Required   │
       │ 4. MFA Challenge │    (ante handler) │
       │<─────────────────│<──────────────────│
       │                  │                   │
       │ 5. TOTP/WebAuthn │                   │
       │    Response      │                   │
       │─────────────────>│ 6. MsgVerifyMFA   │
       │                  │──────────────────>│
       │                  │                   │
       │                  │ 7. Verify Token   │
       │                  │    + Execute Tx   │
       │                  │                   │
       │ 8. Tx Result     │                   │
       │<─────────────────│<──────────────────│
       │                  │                   │
```

### 3. Role-Based Access Control

| Role        | Permissions                            | Assignment                |
| ----------- | -------------------------------------- | ------------------------- |
| `admin`     | All operations, parameter updates      | Genesis or governance     |
| `validator` | Identity verification, vote submission | Staking module            |
| `provider`  | Provider registration, bid submission  | Self-registration + audit |
| `user`      | Standard transactions, identity upload | Default                   |
| `auditor`   | Provider audits, attribute signing     | Governance approval       |

### 4. Transaction Authorization Matrix

| Transaction Type  | Signer Required    | MFA Required          | Additional Checks                       |
| ----------------- | ------------------ | --------------------- | --------------------------------------- |
| MsgSend           | Account owner      | If amount > threshold | Rate limiting                           |
| MsgSubmitScope    | Account owner      | No                    | Approved client signature, salt binding |
| MsgVerifyScope    | Validator          | No                    | Validator set membership                |
| MsgRegisterDevice | Account owner      | Yes (if re-register)  | Device attestation                      |
| MsgUpdateParams   | Gov module account | No                    | Proposal passed                         |
| MsgCreateOrder    | Account owner      | If sensitive          | Balance check                           |

---

## External Trust Boundaries

### 1. Client Application Boundary

```
┌─────────────────────────────────────────────────────────────────┐
│                    UNTRUSTED: User Device                        │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   Approved Client App                        ││
│  │  ┌─────────────────────────────────────────────────────────┐││
│  │  │              TRUSTED: Secure Enclave (TEE)              │││
│  │  │  • Client signing key                                   │││
│  │  │  • Capture metadata generation                          │││
│  │  │  • Salt binding computation                             │││
│  │  └─────────────────────────────────────────────────────────┘││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

**Trust Assumptions:**

- Client app binary has not been tampered with
- Secure enclave correctly isolates signing key
- User's device OS is not compromised (best effort)

### 2. Validator Node Boundary

```
┌─────────────────────────────────────────────────────────────────┐
│                 SEMI-TRUSTED: Validator Host OS                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                TRUSTED: Validator Enclave (TEE)              ││
│  │  • Identity decryption                                       ││
│  │  • ML model execution                                        ││
│  │  • Consensus vote signing                                    ││
│  │  • Sealed key storage                                        ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

**Trust Assumptions:**

- TEE hardware (SGX/SEV-SNP) provides memory isolation
- Validator operator cannot access enclave memory
- Attestation correctly verifies enclave integrity

### 3. External Service Boundaries

| Service               | Trust Level | Data Exposed                | Mitigation                   |
| --------------------- | ----------- | --------------------------- | ---------------------------- |
| Gov data APIs (AAMVA) | MEDIUM      | Name, DOB, address (hashed) | Minimal data, hashed queries |
| EduGAIN SAML IdP      | MEDIUM      | Email, institution          | Signed assertions verified   |
| Payment processors    | MEDIUM      | User ID, amount             | No PII exposed               |
| Cloud providers       | LOW         | Encrypted blobs only        | Client-side encryption       |

---

## Out of Scope

### Explicitly Excluded

| Component                | Reason                        | Notes                              |
| ------------------------ | ----------------------------- | ---------------------------------- |
| Web portal UI            | Client-side, separate review  | Basic XSS/CSRF review elsewhere    |
| Mobile app UI            | Client-side, app store review | Security review during app release |
| Deployment scripts       | Infrastructure, not code      | Covered by ops security review     |
| CI/CD pipelines          | DevSecOps, separate review    | Covered by supply chain audit      |
| Third-party dependencies | Separate SBOM review          | Go modules, npm packages           |
| Load balancer config     | Infrastructure                | Covered by network security review |
| DNS/TLS certificates     | Infrastructure                | Covered by ops security review     |

### Deferred to Phase 2

| Component                   | Reason              | Target Date |
| --------------------------- | ------------------- | ----------- |
| TEE hardware implementation | Hardware dependency | Q2 2026     |
| Cross-chain (IBC) security  | IBC audit separate  | Q2 2026     |
| Advanced DeFi integrations  | Future feature      | TBD         |

---

## Known Issues

### Critical (Must Fix Before Mainnet)

| ID    | Issue                                                | Location                          | Status       |
| ----- | ---------------------------------------------------- | --------------------------------- | ------------ |
| K-001 | Proto stubs in VEID may cause serialization mismatch | `x/veid/types/proto_stub.go`      | In Progress  |
| K-002 | VEID validator authorization not implemented         | `x/veid/keeper/msg_server.go:193` | Acknowledged |

### High (Should Fix)

| ID    | Issue                                       | Location                          | Status                 |
| ----- | ------------------------------------------- | --------------------------------- | ---------------------- |
| K-003 | Provider public key stub                    | `x/provider/keeper/keeper.go:156` | Fixed                  |
| K-004 | time.Now() usage in queries (non-consensus) | Various                           | Acceptable for queries |

### Medium (Track)

| ID    | Issue                                   | Location                                | Status      |
| ----- | --------------------------------------- | --------------------------------------- | ----------- |
| K-005 | Test suites disabled due to API changes | Various `*_test.go`                     | In Progress |
| K-006 | Some adapters are stubs                 | `pkg/ood_adapter/`, `pkg/moab_adapter/` | By Design   |

---

## Audit Questionnaire Responses

### Q1: What cryptographic libraries are used?

| Library                          | Version | Purpose                               |
| -------------------------------- | ------- | ------------------------------------- |
| `golang.org/x/crypto`            | v0.24.0 | X25519, XSalsa20-Poly1305, curve25519 |
| `crypto/sha256`                  | stdlib  | Hashing, fingerprints                 |
| `crypto/rand`                    | stdlib  | Cryptographic random generation       |
| `crypto/ed25519`                 | stdlib  | Signature verification                |
| `github.com/btcsuite/btcd/btcec` | v2.3.0  | secp256k1 (optional)                  |

### Q2: How is random number generation handled?

- All cryptographic randomness uses `crypto/rand.Reader`
- Nonces are generated fresh for each encryption operation
- No custom PRNG implementations
- Fallback to `/dev/urandom` on Unix systems

### Q3: How are keys stored and managed?

| Context              | Storage                          | Protection                         |
| -------------------- | -------------------------------- | ---------------------------------- |
| User keys            | Client wallet (Keplr, etc.)      | User-managed, BIP-39 backup        |
| Validator keys       | Keyring (`file`, `os`, `ledger`) | HSM recommended for production     |
| Provider keys        | Keyring with optional Ledger     | Hardware wallet supported          |
| Approved client keys | App binary (obfuscated)          | Code signing, rotation per release |

### Q4: What input validation is performed?

- All protobuf messages have `ValidateBasic()` methods
- Key sizes validated before cryptographic operations
- Nonce sizes verified
- Signature format validation before verification
- Rate limiting on all public endpoints

### Q5: How is consensus safety ensured?

- No `time.Now()` in state-changing operations (use `ctx.BlockTime()`)
- No floating-point arithmetic in state machines
- Deterministic iteration order (sorted keys)
- No external I/O in BeginBlock/EndBlock
- ML inference uses fixed seed, CPU-only, deterministic ops

### Q6: How is data classified and protected?

See [data-classification.md](../../_docs/data-classification.md) for full classification policy.

| Classification | Examples                        | Encryption          | Access Control       |
| -------------- | ------------------------------- | ------------------- | -------------------- |
| CRITICAL/PII   | Identity documents, face images | Envelope encryption | Threshold decryption |
| SECRET         | Private keys, MFA seeds         | Sealed storage      | Never transmitted    |
| INTERNAL       | Wallet addresses, scores        | None (public)       | Pseudonymous         |
| PUBLIC         | Chain parameters, docs          | None                | Unrestricted         |

### Q7: What are the recovery procedures for key compromise?

| Key Type             | Recovery Procedure                            |
| -------------------- | --------------------------------------------- |
| User wallet          | BIP-39 mnemonic recovery, new account if lost |
| Validator consensus  | Key rotation, migration transaction           |
| Validator encryption | Governance proposal for key rotation          |
| Approved client      | App update, allowlist update via governance   |

### Q8: How is the audit trail maintained?

- All state changes emit Cosmos SDK events
- Events include: tx hash, sender, action, timestamp
- On-chain immutable log
- No event can be emitted without state change

---

## Appendix A: File Inventory for Tier 1 Modules

### x/veid/

```
x/veid/
├── alias.go
├── genesis.go
├── module.go
├── keeper/
│   ├── grpc_query.go
│   ├── keeper.go
│   ├── msg_server.go
│   └── vote_extension.go
└── types/
    ├── codec.go
    ├── consent.go
    ├── derived_features.go
    ├── embedding_envelope.go
    ├── errors.go
    ├── events.go
    ├── genesis.go
    ├── identity.go
    ├── keys.go
    ├── msgs.go
    ├── scope.go
    ├── score.go
    ├── security_controls.go
    ├── upload.go
    ├── verification.go
    └── wallet.go
```

### x/encryption/

```
x/encryption/
├── alias.go
├── genesis.go
├── module.go
├── crypto/
│   ├── algorithms.go
│   ├── envelope.go
│   ├── ledger.go
│   └── mnemonic.go
├── keeper/
│   ├── grpc_query.go
│   ├── keeper.go
│   └── msg_server.go
└── types/
    ├── codec.go
    ├── envelope.go
    ├── errors.go
    ├── events.go
    ├── genesis.go
    ├── keys.go
    ├── msgs.go
    └── types.go
```

### x/mfa/

```
x/mfa/
├── alias.go
├── genesis.go
├── module.go
├── keeper/
│   ├── grpc_query.go
│   ├── keeper.go
│   └── msg_server.go
└── types/
    ├── challenge.go
    ├── codec.go
    ├── device.go
    ├── errors.go
    ├── events.go
    ├── genesis.go
    ├── keys.go
    ├── msgs.go
    └── sensitive_tx.go
```

---

## Appendix B: Testing Coverage

| Module               | Unit Tests   | Integration Tests | Fuzz Tests | Property Tests |
| -------------------- | ------------ | ----------------- | ---------- | -------------- |
| x/veid/types         | ✅ 288 lines | ✅                | ✅         | ✅             |
| x/encryption/crypto  | ✅ 302 lines | ✅                | ✅         | N/A            |
| x/mfa/types          | ✅           | ⚠️ Partial        | ⚠️ Partial | N/A            |
| x/roles/types        | ✅           | ✅                | N/A        | N/A            |
| pkg/enclave_runtime  | ✅           | ✅                | N/A        | N/A            |
| pkg/capture_protocol | ✅           | ✅                | ✅         | N/A            |

---

_Document prepared for VirtEngine external security audit_  
_Contact: security@virtengine.com_
