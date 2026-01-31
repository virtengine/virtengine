# Security Review Report: ML and Verification Services

**Review ID:** VE-8D  
**Date:** 2026-01-25  
**Status:** COMPLETED  
**Reviewer:** VirtEngine Security Team  

---

## Executive Summary

This security review covers ML inference and verification services including attestation, signing, anti-fraud, and replay protection controls. The review found the implementation to be **robust with no HIGH or CRITICAL unresolved findings**.

### Key Findings Summary

| Severity | Count | Status |
|----------|-------|--------|
| CRITICAL | 0 | N/A |
| HIGH | 0 | N/A |
| MEDIUM | 2 | Mitigated |
| LOW | 3 | Acceptable |
| INFO | 5 | Documented |

---

## Scope

### Components Reviewed

| Component | Path | Description |
|-----------|------|-------------|
| ML Inference Determinism | `pkg/inference/determinism.go` | Consensus-safe ML inference |
| ML Scorer | `pkg/inference/scorer.go` | TensorFlow scoring engine |
| Attestation Signer | `pkg/verification/signer/` | Ed25519 signing with key rotation |
| Nonce Store | `pkg/verification/nonce/` | Replay protection |
| Key Storage | `pkg/verification/keystorage/` | Secure key storage backends |
| Rate Limiter | `pkg/ratelimit/` | Token bucket rate limiting |
| SMS Anti-Fraud | `pkg/verification/sms/antifraud.go` | VoIP detection, velocity checks |
| Email Verification | `pkg/verification/email/` | OTP and magic link verification |
| OIDC Verifier | `pkg/verification/oidc/` | SSO/OIDC token verification |
| Audit Logger | `pkg/verification/audit/` | Immutable audit logging |

### Out of Scope

- Third-party dependencies (covered by separate SBOM review)
- Network-level attacks (covered by infrastructure security)
- Physical security of validator hardware

---

## Detailed Findings

### F-001: Memory-Based Nonce Store Not Production-Ready (MEDIUM - Mitigated)

**Location:** `pkg/verification/nonce/memory.go`

**Description:** The in-memory nonce store is not suitable for production use in multi-node deployments.

**Evidence:** Code comment at line 15: "NOTE: Memory store is NOT suitable for production use in multi-node deployments."

**Mitigation:** Redis-based nonce store (`pkg/verification/nonce/redis.go`) is available and recommended for production.

**Status:** MITIGATED - Production deployments should use Redis backend.

---

### F-002: OIDC Discovery Cache TTL (MEDIUM - Mitigated)

**Location:** `pkg/verification/oidc/verifier.go`

**Description:** OIDC JWKS discovery cache has 1-hour TTL which balances security vs. performance.

**Risk:** If a key is compromised, tokens signed with that key could be valid for up to 1 hour after revocation.

**Mitigation:** 
- TTL is configurable per-deployment
- Token expiry is typically shorter than cache TTL
- Critical key rotations can trigger cache invalidation

**Status:** MITIGATED - Acceptable trade-off with monitoring.

---

### F-003: SMS Inherent Vulnerabilities (LOW - Acceptable)

**Location:** `pkg/verification/sms/antifraud.go`

**Description:** SMS verification is inherently less secure than cryptographic methods.

**Risk:** SIM swap attacks, SMS interception.

**Mitigation:**
- VoIP carrier detection blocks most automation
- Velocity checks limit abuse rate
- SMS is supplementary, not sole verification factor

**Status:** ACCEPTABLE - Multiple mitigations in place.

---

### F-004: File Key Storage for Development (LOW - Acceptable)

**Location:** `pkg/verification/keystorage/file.go`

**Description:** File-based key storage uses AES-GCM encryption but is less secure than HSM.

**Risk:** Key extraction if filesystem is compromised.

**Mitigation:**
- Production should use Vault or HSM backends
- File permissions enforced (0600/0700)
- Encryption key required for decryption

**Status:** ACCEPTABLE - Development only; production uses Vault/HSM.

---

### F-005: Audit Log Memory Buffer (LOW - Acceptable)

**Location:** `pkg/verification/audit/logger.go`

**Description:** Memory-backed audit logger could lose entries on crash.

**Risk:** Audit trail gaps during system failures.

**Mitigation:**
- File-based and remote loggers available
- Production uses persistent storage
- Crash recovery procedures documented

**Status:** ACCEPTABLE - Memory logger for development/testing only.

---

### F-006: Determinism Config Documentation (INFO)

**Location:** `pkg/inference/determinism.go`

**Description:** Determinism configuration well-documented in code.

**Positive Finding:** 
- Fixed random seed (42) enforced
- CPU-only mode for reproducibility
- 6-decimal precision for hash normalization

**Status:** DOCUMENTED

---

### F-007: Key Lifecycle States (INFO)

**Location:** `pkg/verification/signer/signer.go`

**Description:** Clear key state machine: pending → active → rotating → revoked.

**Positive Finding:**
- No invalid state transitions possible
- Private keys cleared from memory after use
- Overlapping validity for smooth rotation

**Status:** DOCUMENTED

---

### F-008: Rate Limiting Architecture (INFO)

**Location:** `pkg/ratelimit/redis_limiter.go`

**Description:** Multi-tier token bucket rate limiting with Lua scripts.

**Positive Finding:**
- Atomic operations via Lua scripts
- Bypass detection and auto-banning
- Graceful degradation under load
- Priority endpoints not degraded

**Status:** DOCUMENTED

---

### F-009: Nonce Binding Security (INFO)

**Location:** `pkg/verification/nonce/memory.go`

**Description:** Comprehensive nonce binding prevents replay attacks.

**Positive Finding:**
- Issuer fingerprint binding
- Subject address binding
- Configurable expiry window
- Atomic validate-and-use

**Status:** DOCUMENTED

---

### F-010: Input Validation (INFO)

**Location:** `pkg/security/`

**Description:** Comprehensive input validation utilities.

**Positive Finding:**
- Path traversal protection
- Integer overflow protection
- Input sanitization helpers

**Status:** DOCUMENTED

---

## Security Controls Verified

### 1. ML Inference Determinism

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Fixed random seed (42) | Code review | ✅ PASS |
| CPU-only inference | Code review, test | ✅ PASS |
| Deterministic TF ops | Code review | ✅ PASS |
| Hash precision (6 dec) | Unit test | ✅ PASS |
| Cross-validator conformance | Integration test | ✅ PASS |

### 2. Replay Protection

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Nonce uniqueness | Unit test | ✅ PASS |
| Issuer binding | Unit test | ✅ PASS |
| Subject binding | Unit test | ✅ PASS |
| Expiry enforcement | Unit test | ✅ PASS |
| Concurrent replay prevention | Unit test | ✅ PASS |

### 3. Key Management

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Ed25519 signing | Code review | ✅ PASS |
| Key state machine | Unit test | ✅ PASS |
| Private key clearing | Code review | ✅ PASS |
| Overlapping rotation | Unit test | ✅ PASS |
| Revoked key rejection | Unit test | ✅ PASS |

### 4. Rate Limiting

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Token bucket algorithm | Code review | ✅ PASS |
| Multi-tier limits | Unit test | ✅ PASS |
| Bypass detection | Unit test | ✅ PASS |
| Auto-banning | Unit test | ✅ PASS |
| Graceful degradation | Unit test | ✅ PASS |

### 5. Anti-Fraud Controls

| Control | Verification Method | Result |
|---------|---------------------|--------|
| VoIP detection | Unit test | ✅ PASS |
| Per-phone velocity | Unit test | ✅ PASS |
| Per-IP velocity | Unit test | ✅ PASS |
| Device fingerprinting | Code review | ✅ PASS |
| Risk scoring | Unit test | ✅ PASS |

### 6. OIDC/SSO Verification

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Token expiry validation | Unit test | ✅ PASS |
| Issuer validation | Unit test | ✅ PASS |
| Audience validation | Unit test | ✅ PASS |
| Signature verification | Code review | ✅ PASS |
| JWKS caching | Code review | ✅ PASS |

### 7. Audit Logging

| Control | Verification Method | Result |
|---------|---------------------|--------|
| Append-only logging | Unit test | ✅ PASS |
| Hash chain integrity | Unit test | ✅ PASS |
| Timestamp ordering | Unit test | ✅ PASS |
| Delete rejection | Unit test | ✅ PASS |

---

## Security Tests Added

New security tests added to `tests/security/ml_verification_test.go`:

| Test Suite | Test Cases | Coverage |
|------------|------------|----------|
| TestMLDeterminismSecurity | 4 | ML determinism, seed, precision, GPU |
| TestMLOutputValidation | 8 | Score bounds, overflow protection |
| TestNonceReplayProtection | 5 | Replay, binding, expiry, concurrent |
| TestRateLimitingAbusePrevention | 4 | Token bucket, tiers, bypass, degradation |
| TestSMSAntiFraud | 5 | VoIP, velocity, device, risk scoring |
| TestKeyRotationSecurity | 4 | State transitions, revocation, memory |
| TestOIDCVerificationSecurity | 5 | Expiry, issuer, audience, signature |
| TestAuditLogIntegrity | 3 | Append-only, hash chain, ordering |
| TestInjectionPrevention | 2 | Path traversal, input sanitization |

**Total: 40 security test cases**

Run with: `go test -tags=security ./tests/security/... -run "TestMLVerification"`

---

## Threat Model Updates

Added new threat category **T9: ML Inference and Verification Services** with 6 sub-threats:

| ID | Name | Impact | Likelihood | Status |
|----|------|--------|------------|--------|
| T9.1 | ML Model Tampering | CRITICAL | LOW | Mitigated |
| T9.2 | Non-Deterministic Inference | HIGH | MEDIUM | Mitigated |
| T9.3 | Verification Service Bypass | HIGH | LOW | Mitigated |
| T9.4 | Attestation Replay Attack | MEDIUM | VERY LOW | Mitigated |
| T9.5 | Key Rotation Failures | MEDIUM | LOW | Mitigated |
| T9.6 | SMS Fraud Abuse | MEDIUM | MEDIUM | Mitigated |

See `_docs/threat-model.md` for full details.

---

## Recommendations

### Immediate (Before Production)

1. **Use Redis nonce store** - Replace memory store with Redis for production deployments.
2. **Configure Vault/HSM key storage** - Use HashiCorp Vault or HSM for production key storage.
3. **Enable persistent audit logging** - Configure file or remote audit log backend.

### Short-Term (Next Release)

1. **Add automated security scans** - Integrate security tests into CI pipeline.
2. **Implement OIDC cache invalidation** - Add endpoint for emergency JWKS cache refresh.
3. **Enhanced SMS monitoring** - Add alerting for unusual SMS verification patterns.

### Long-Term (Roadmap)

1. **Hardware attestation** - Consider TPM-based attestation for mobile capture apps.
2. **Threshold signing** - Implement threshold signatures for critical attestations.
3. **Zero-knowledge proofs** - Explore ZK proofs for privacy-preserving verification.

---

## Compliance Evidence

| Requirement | Evidence | Location |
|-------------|----------|----------|
| Input validation | Security tests | `tests/security/` |
| Replay protection | Nonce tests | `pkg/verification/nonce/` |
| Key management | Rotation tests | `pkg/verification/signer/` |
| Rate limiting | Limiter tests | `pkg/ratelimit/` |
| Audit logging | Logger tests | `pkg/verification/audit/` |
| Threat modeling | Updated model | `_docs/threat-model.md` |

---

## Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Security Reviewer | VirtEngine Security Team | 2026-01-25 | ✅ |
| Technical Lead | Pending | | |
| Product Owner | Pending | | |

---

*This report was generated as part of security review VE-8D.*  
*All findings have been documented and tracked for remediation.*
