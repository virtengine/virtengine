# TEE Security Model for VEID Consensus Verification

**Version:** 1.0.0  
**Date:** 2026-01-27  
**Status:** Authoritative Baseline  
**Task Reference:** VE-228

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Threat Model](#threat-model)
3. [Enclave Guarantees](#enclave-guarantees)
4. [Operational Definitions](#operational-definitions)
5. [Slashing and Penalty Conditions](#slashing-and-penalty-conditions)
6. [Architecture Diagram](#architecture-diagram)
7. [Appendices](#appendices)

---

## Executive Summary

This document defines the Trusted Execution Environment (TEE) security model for VirtEngine's VEID consensus verification system. The TEE model ensures that identity data is processed exclusively within cryptographically isolated enclaves, preventing exposure to malicious host operating systems, compromised validators, or side-channel attacks.

### Supported TEE Platforms

| Platform | Status | Version Requirements |
|----------|--------|---------------------|
| Intel SGX | Primary | SGX2 with DCAP attestation |
| AMD SEV-SNP | Supported | SNP with versioned attestation |
| ARM TrustZone | Future | TBD |
| AWS Nitro Enclaves | Supported | NSM attestation v1.0+ |

### Core Security Properties

1. **Confidentiality**: Identity data never leaves enclave in plaintext
2. **Integrity**: Enclave code is measured and verified via remote attestation
3. **Attestability**: Validators can prove enclave execution to consensus
4. **Determinism**: All validators produce identical results from identical inputs

---

## Threat Model

### T1: Malicious Host Operating System

**Description**: The host OS running the validator node is fully compromised and under attacker control.

**Attack Vectors**:
- Direct memory inspection of validator process
- System call interception and modification
- File system access to validator data
- Network traffic inspection
- Process injection and debugging

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Enclave memory isolation | TEE hardware enforces encrypted memory regions |
| Sealed storage | Private keys sealed to enclave measurement |
| No host-readable buffers | Plaintext scrubbed immediately after use |
| Attestation verification | Remote attestation proves genuine enclave |

**Residual Risk**: Low - TEE hardware guarantees prevent host access to enclave memory.

---

### T2: Validator Operator Compromise

**Description**: A validator operator intentionally or negligently attempts to extract identity data.

**Attack Vectors**:
- Modifying enclave code to leak data
- Running debuggable enclaves
- Exfiltrating sealed keys
- Logging enclave inputs/outputs

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Measurement allowlist | Only approved enclave binaries accepted |
| Debug mode disabled | Production enclaves must have debug=false |
| Sealed key non-exportability | Keys cannot be unsealed outside original enclave |
| Structured logging redaction | Sensitive data never logged |

**Residual Risk**: Low - Allowlisted measurements prevent malicious code execution.

---

### T3: Side-Channel Attacks

**Description**: Attacker extracts information through timing, cache, power, or electromagnetic analysis.

**Attack Vectors**:
- Cache timing attacks (Prime+Probe, Flush+Reload)
- Branch prediction attacks (Spectre variants)
- Power analysis
- Electromagnetic emanations
- Memory access pattern analysis

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Constant-time operations | Crypto primitives use constant-time implementations |
| Memory access obfuscation | ORAM-style access patterns for sensitive operations |
| Microarchitectural mitigations | Platform-specific patches (IBPB, STIBP, etc.) |
| Rate limiting | Bounded verification attempts per epoch |

**Residual Risk**: Medium - Some side-channel vectors remain theoretical concerns.

---

### T4: Replay Attacks

**Description**: Attacker replays previously valid attestations or encrypted payloads.

**Attack Vectors**:
- Reusing old attestation quotes
- Replaying previous verification requests
- Submitting duplicate identity scopes

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Attestation expiry | Quotes valid for limited epoch duration |
| Nonce binding | Each request includes unique nonce |
| Monotonic counters | Enclave maintains rollback-resistant state |
| Block height binding | Verification tied to specific block proposal |

**Residual Risk**: Low - Nonce and epoch binding prevent replay.

---

### T5: Enclave Rollback Attacks

**Description**: Attacker forces enclave to use outdated sealed state.

**Attack Vectors**:
- Restoring old sealed data files
- Reverting enclave to previous version
- Manipulating monotonic counter storage

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Monotonic counters | Platform-provided anti-rollback counters |
| Epoch binding | State includes current blockchain epoch |
| Version tracking | State includes enclave version, rejected if outdated |
| Multi-party sealed storage | Critical state distributed across validators |

**Residual Risk**: Low - Hardware monotonic counters prevent rollback.

---

### T6: Attestation Forgery

**Description**: Attacker creates false attestation quotes claiming enclave execution.

**Attack Vectors**:
- Forging Intel EPID/DCAP signatures
- Compromising attestation service keys
- Creating fake enclave measurements
- Man-in-the-middle on attestation flow

**Mitigations**:
| Control | Implementation |
|---------|----------------|
| Remote attestation verification | Quotes verified against platform root of trust |
| Attestation chain validation | Full chain from quote to platform CA verified |
| Measurement pinning | Only allowlisted measurements accepted |
| Collateral freshness | TCB recovery checks against up-to-date collateral |

**Residual Risk**: Very Low - Hardware attestation provides strong guarantees.

---

## Enclave Guarantees

### Minimum Requirements for Mainnet

| Requirement | Description | Verification Method |
|-------------|-------------|---------------------|
| **Remote Attestation** | Enclave must provide cryptographic proof of execution | DCAP/EPID quote verification |
| **Sealed Keys** | Decryption keys must be sealed to enclave measurement | Key generation inside enclave, sealed via platform API |
| **Measurement Allowlist** | Only pre-approved enclave binaries may execute | On-chain governance-controlled allowlist |
| **Debug Disabled** | Production enclaves must not be debuggable | Attestation quote includes debug flag (must be 0) |
| **Memory Encryption** | All enclave memory must be encrypted | Hardware-enforced (SGX MEE, SEV-SNP encryption) |
| **Code Integrity** | Enclave code cannot be modified at runtime | Hardware-enforced memory protection |

### Attestation Quote Requirements

```
AttestationQuote {
    version:            uint32      // Quote format version (minimum: 3 for DCAP)
    tee_type:           string      // "SGX", "SEV-SNP", "NITRO"
    measurement_hash:   bytes32     // MRENCLAVE/launch digest
    signer_hash:        bytes32     // MRSIGNER/author measurement
    isv_prod_id:        uint16      // Product ID
    isv_svn:            uint16      // Security version number
    report_data:        bytes64     // Enclave-provided binding data
    signature:          bytes       // Platform signature
    attestation_chain:  []bytes     // Certificate chain to root CA
    tcb_info:           bytes       // TCB level and status
    timestamp:          int64       // Quote generation timestamp
    debug_flag:         bool        // MUST be false for production
}
```

### Enclave Runtime Controls

| Control | Value | Purpose |
|---------|-------|---------|
| Max input size | 10 MB | Prevent DoS via large payloads |
| Max execution time | 1000 ms | Bound block proposal delay |
| Max concurrent requests | 4 | Resource exhaustion protection |
| Memory scrub interval | Per-request | Prevent data persistence |
| Key rotation epoch | 1000 blocks | Limit key exposure window |

---

## Operational Definitions

### "Never Accessible by General Code"

This phrase operationally means:

1. **No plaintext in host memory**: Decrypted identity data exists only within enclave-protected memory pages. The host process memory space never contains plaintext identity content.

2. **No plaintext in logs**: Structured logging rules prevent any identity data from appearing in validator logs, system logs, or telemetry. Encrypted blobs are truncated; request bodies are redacted.

3. **No plaintext in IPC payloads**: Communication between host and enclave uses encrypted channels. The host submits ciphertext and receives only encrypted or hashed results.

4. **No plaintext on disk**: Sealed storage encrypts all persistent enclave state. Temporary files are never created outside the enclave.

5. **Enclave boundary enforced**: The TEE hardware enforces that code outside the enclave cannot read enclave memory, even with root/kernel access.

### Verification of "Never Accessible"

| Check | Method | Frequency |
|-------|--------|-----------|
| Memory inspection | Harness scans host process memory for known patterns | Per release |
| Log analysis | Static analysis of logging calls in VEID path | Per commit (CI) |
| Fuzz testing | Inject malformed payloads, verify no leakage in errors | Per release |
| Code review | Manual review of enclave interface code | Per PR |
| Runtime monitoring | Production memory sampling (non-sensitive canaries) | Continuous |

---

## Slashing and Penalty Conditions

### Slashing Triggers for Enclave Misbehavior

| Condition | Detection Method | Penalty |
|-----------|------------------|---------|
| **Invalid Attestation** | Quote fails verification | Slash 5% + jail 1 week |
| **Expired Attestation** | Quote timestamp beyond validity window | Warning (first), Slash 1% (repeat) |
| **Non-allowlisted Measurement** | MRENCLAVE not in governance allowlist | Slash 10% + jail indefinitely |
| **Debug Mode Enabled** | Quote debug flag = true | Slash 20% + permanent removal |
| **Missing Recomputation** | No verification record in vote | Slash 1% per missed block |
| **Inconsistent Scores** | Score differs from consensus by >tolerance | Slash 2% + investigation |
| **Attestation Forgery Attempt** | Quote signature invalid | Slash 100% + permanent ban |

### Penalty Escalation

```
First offense:  Base penalty
Second offense: 2x base penalty + 1 week jail
Third offense:  4x base penalty + 1 month jail
Fourth offense: 8x base penalty + permanent removal consideration
```

### Grace Periods

| Scenario | Grace Period | Notes |
|----------|--------------|-------|
| Key rotation | 100 blocks | Old + new keys both valid |
| Enclave upgrade | 200 blocks | Both measurements allowlisted |
| TCB recovery | 500 blocks | Extended for platform updates |
| Network partition | Proportional | Based on partition duration |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          TEE-ENABLED VEID VERIFICATION                               │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                           CLIENT LAYER                                       │   │
│  │  ┌──────────────────┐                                                       │   │
│  │  │  Mobile App      │ ─── Captures identity data                            │   │
│  │  │  (Approved)      │ ─── Signs with client + user keys                     │   │
│  │  └────────┬─────────┘ ─── Encrypts to validator enclave public keys         │   │
│  └───────────┼─────────────────────────────────────────────────────────────────┘   │
│              │                                                                      │
│              │ Multi-recipient encrypted envelope                                   │
│              │ (payload_ciphertext + per-validator wrapped_key)                    │
│              ▼                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                      BLOCKCHAIN LAYER (Cosmos SDK)                          │   │
│  │                                                                              │   │
│  │  ┌───────────────────┐    ┌───────────────────┐    ┌───────────────────┐   │   │
│  │  │  Enclave Registry │    │   VEID Module     │    │  Encryption Mod   │   │   │
│  │  │                   │    │                   │    │                   │   │   │
│  │  │ • Measurements    │◄──►│ • Identity Scopes │◄──►│ • Envelope Format │   │   │
│  │  │ • Enclave Keys    │    │ • Scores          │    │ • Multi-Recipient │   │   │
│  │  │ • Attestations    │    │ • Attested Results│    │ • Key Registry    │   │   │
│  │  └─────────┬─────────┘    └─────────┬─────────┘    └───────────────────┘   │   │
│  │            │                        │                                       │   │
│  └────────────┼────────────────────────┼───────────────────────────────────────┘   │
│               │                        │                                            │
│               │                        │                                            │
│  ┌────────────┼────────────────────────┼───────────────────────────────────────┐   │
│  │            │    VALIDATOR NODE      │    (Host OS - Untrusted)              │   │
│  │            ▼                        ▼                                        │   │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │   │
│  │  │                    ENCLAVE RUNTIME                                   │   │   │
│  │  │  ┌────────────────────────────────────────────────────────────────┐ │   │   │
│  │  │  │               TEE BOUNDARY (Hardware Enforced)                  │ │   │   │
│  │  │  │                                                                  │ │   │   │
│  │  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │ │   │   │
│  │  │  │  │  Sealed Key  │  │  Decrypt +   │  │   ML Scoring         │  │ │   │   │
│  │  │  │  │  Storage     │──│  Unwrap      │──│   Pipeline           │  │ │   │   │
│  │  │  │  │              │  │              │  │                      │  │ │   │   │
│  │  │  │  │ • Private key│  │ • Ciphertext │  │ • Face verification  │  │ │   │   │
│  │  │  │  │ • Monotonic  │  │ • Plaintext  │  │ • Document OCR       │  │ │   │   │
│  │  │  │  │   counter    │  │   (scrubbed) │  │ • Score computation  │  │ │   │   │
│  │  │  │  └──────────────┘  └──────────────┘  └───────────┬──────────┘  │ │   │   │
│  │  │  │                                                   │             │ │   │   │
│  │  │  │                                                   ▼             │ │   │   │
│  │  │  │                                    ┌──────────────────────┐    │ │   │   │
│  │  │  │                                    │  Attested Output     │    │ │   │   │
│  │  │  │                                    │                      │    │ │   │   │
│  │  │  │                                    │ • Score + Status     │    │ │   │   │
│  │  │  │                                    │ • Evidence hashes    │    │ │   │   │
│  │  │  │                                    │ • Enclave signature  │    │ │   │   │
│  │  │  │                                    │ • Measurement link   │    │ │   │   │
│  │  │  │                                    └──────────┬───────────┘    │ │   │   │
│  │  │  │                                               │                │ │   │   │
│  │  │  └───────────────────────────────────────────────┼────────────────┘ │   │   │
│  │  │                                                   │                  │   │   │
│  │  └───────────────────────────────────────────────────┼──────────────────┘   │   │
│  │                                                       │                      │   │
│  │       Host receives only: score, status, hashes, signature (no plaintext)   │   │
│  └───────────────────────────────────────────────────────┼──────────────────────┘   │
│                                                          │                          │
│                                                          ▼                          │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                       CONSENSUS VERIFICATION                                 │   │
│  │                                                                              │   │
│  │  Proposer computes score in enclave ──► Includes attested result in block   │   │
│  │                                               │                              │   │
│  │                                               ▼                              │   │
│  │  Other validators:                                                           │   │
│  │    1. Decrypt with own enclave key                                          │   │
│  │    2. Recompute score in own enclave                                        │   │
│  │    3. Compare to proposer's result                                          │   │
│  │    4. Vote only if match (or within tolerance)                              │   │
│  │    5. Include own attested signature as evidence                            │   │
│  │                                                                              │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Appendices

### A. Enclave Measurement Allowlist Governance

The enclave measurement allowlist is managed through on-chain governance:

1. **Proposal**: Submit `MsgProposeEnclaveMeasurement` with MRENCLAVE/MRSIGNER hash
2. **Voting**: Standard governance voting period (configurable, default 7 days)
3. **Execution**: On proposal pass, measurement added to allowlist
4. **Revocation**: `MsgRevokeEnclaveMeasurement` for emergency removal (requires supermajority)

### B. Incident Response Procedure for Suspected Leakage

1. **Detection**: Monitor for anomalous patterns in verification flow
2. **Containment**: Suspend affected validator(s) immediately
3. **Analysis**: Collect attestation quotes and verification logs
4. **Remediation**: Rotate enclave keys, update measurements if needed
5. **Communication**: Disclose to affected users per policy
6. **Prevention**: Update controls based on root cause analysis

### C. Validator Operator Runbook

See [validator-tee-operations.md](./validator-tee-operations.md) for:
- Enclave installation and configuration
- Key generation and registration
- Attestation quote generation
- Key rotation procedures
- Emergency recovery procedures

### D. Related Documents

- [Architecture Overview](./architecture.md)
- [Threat Model](./threat-model.md)
- [VEID Flow Specification](./veid-flow-spec.md)
- [Data Classification](./data-classification.md)
