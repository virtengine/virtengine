# TEE Integration Architecture for VirtEngine VEID

**Version:** 1.0.0  
**Date:** 2026-01-29  
**Status:** POC Architecture Document  
**Task Reference:** VE-2023

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Intel SGX vs AMD SEV-SNP Comparison](#intel-sgx-vs-amd-sev-snp-comparison)
3. [Remote Attestation Flows](#remote-attestation-flows)
4. [Key Derivation Design](#key-derivation-design)
5. [Sealed Storage Architecture](#sealed-storage-architecture)
6. [Hardware Requirements Matrix](#hardware-requirements-matrix)
7. [Implementation Strategy](#implementation-strategy)

---

## Executive Summary

This document defines the technical architecture for integrating Trusted Execution Environments (TEE) into VirtEngine's VEID identity verification system. The integration enables validators to process sensitive identity data in hardware-isolated enclaves, ensuring:

- **Confidentiality**: Identity data decrypted only within TEE boundaries
- **Integrity**: Enclave code measured and verified via remote attestation
- **Determinism**: All validators produce identical verification scores
- **Non-repudiation**: Results cryptographically bound to enclave execution

### Recommended Platform Strategy

| Use Case | Primary Platform | Rationale |
|----------|-----------------|-----------|
| Identity Scoring (ML) | AMD SEV-SNP | No memory limitations, near-native performance |
| Key Management | Intel SGX | Smaller TCB, proven key protection |
| Cloud Deployment | AWS Nitro / Azure CVM | Managed TEE infrastructure |

---

## Intel SGX vs AMD SEV-SNP Comparison

### Architecture Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        INTEL SGX ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         HOST APPLICATION                            │   │
│  │  ┌─────────────┐                           ┌─────────────────────┐  │   │
│  │  │  Untrusted  │ ──── ECALL/OCALL ────────▶│      SGX ENCLAVE    │  │   │
│  │  │  Runtime    │                           │  ┌───────────────┐  │  │   │
│  │  │             │                           │  │ TRUSTED CODE  │  │  │   │
│  │  └─────────────┘                           │  │  - Key mgmt   │  │  │   │
│  │                                            │  │  - ML scoring │  │  │   │
│  │                                            │  │  - Signing    │  │  │   │
│  │                                            │  └───────────────┘  │  │   │
│  │                                            │  EPC Memory (128MB) │  │   │
│  │                                            └─────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Memory Encryption: MEE (Memory Encryption Engine)                          │
│  Isolation Level: Application (process within enclave)                      │
│  TCB Size: ~100KB (enclave code + SGX SDK)                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                      AMD SEV-SNP ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                           HYPERVISOR                                │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │                   CONFIDENTIAL VM (SNP)                     │   │   │
│  │  │  ┌───────────────────────────────────────────────────────┐  │   │   │
│  │  │  │                     GUEST OS (Linux)                   │  │   │   │
│  │  │  │  ┌─────────────────┐  ┌─────────────────────────────┐ │  │   │   │
│  │  │  │  │  VEID Service   │  │  ML Scoring Service         │ │  │   │   │
│  │  │  │  │  - gRPC server  │  │  - TensorFlow Lite          │ │  │   │   │
│  │  │  │  │  - Key storage  │  │  - Face verification        │ │  │   │   │
│  │  │  │  │  - Attestation  │  │  - Document OCR             │ │  │   │   │
│  │  │  │  └─────────────────┘  └─────────────────────────────┘ │  │   │   │
│  │  │  └───────────────────────────────────────────────────────┘  │   │   │
│  │  │  Full VM Memory (No EPC limits)                             │   │   │
│  │  └─────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Memory Encryption: SME/SEV (hardware AES-128 per-VM key)                   │
│  Isolation Level: Virtual Machine (entire guest OS)                         │
│  TCB Size: ~50MB (kernel + critical services)                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Feature Comparison Matrix

| Feature | Intel SGX | AMD SEV-SNP | VEID Requirement |
|---------|-----------|-------------|------------------|
| **Memory Encryption** | MEE (proprietary) | AES-128 | ✅ Both sufficient |
| **Memory Size** | 128-512MB EPC | Unlimited | SEV-SNP wins (ML models >200MB) |
| **Performance** | 5-20% overhead | <5% overhead | SEV-SNP wins |
| **Attestation** | DCAP (Intel IAS) | AMD PSP | Both available |
| **Memory Integrity** | Limited | Full (SNP) | SEV-SNP wins |
| **Side-Channel Resistance** | Vulnerable (L1TF, etc.) | More resistant | SEV-SNP wins |
| **TCB Size** | Small (~100KB) | Large (~50MB) | SGX wins |
| **Hardware Availability** | Intel Xeon 3rd Gen+ | AMD EPYC Milan+ | Both available |
| **Cloud Support** | Azure DCsv3, IBM | Azure DCasv5, AWS, GCP | Both available |
| **Development Complexity** | High (special SDK) | Low (standard Linux) | SEV-SNP wins |

### Recommendation for VEID Identity Verification

**Primary: AMD SEV-SNP** for production identity scoring:
1. ML models (face verification, OCR) require >200MB memory
2. Near-native performance critical for consensus timing
3. Standard Linux development (no SDK learning curve)
4. Strong memory integrity with SNP extensions

**Secondary: Intel SGX** for key management operations:
1. Smaller TCB for cryptographic operations only
2. Proven sealing mechanism for key protection
3. Fine-grained ECALL/OCALL control

---

## Remote Attestation Flows

### Intel SGX DCAP Attestation Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SGX DCAP ATTESTATION FLOW                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────┐                           ┌───────────────────────────┐  │
│  │   Validator   │                           │    Intel PCS / PCCS       │  │
│  │   (Relying    │                           │  (Provisioning Service)   │  │
│  │    Party)     │                           └───────────────────────────┘  │
│  └───────┬───────┘                                      ▲                   │
│          │                                              │                   │
│          │ 1. Request attestation                       │                   │
│          │    (nonce for freshness)                     │                   │
│          ▼                                              │                   │
│  ┌───────────────┐     2. Generate Quote                │                   │
│  │  SGX Enclave  │ ─────────────────────────────────────┘                   │
│  │               │     (EREPORT → QE → Quote)                               │
│  │  MRENCLAVE:   │                                                          │
│  │  0xabc123...  │     Quote Contents:                                      │
│  │               │     ├─ MRENCLAVE (code hash)                             │
│  │  MRSIGNER:    │     ├─ MRSIGNER (signer hash)                            │
│  │  0xdef456...  │     ├─ ISV SVN (security version)                        │
│  └───────┬───────┘     ├─ ReportData (user data + nonce)                    │
│          │             ├─ TCB Info (platform TCB level)                     │
│          │             └─ QE Signature (ECDSA P-256)                        │
│          │                                                                  │
│          │ 3. Return Quote + Collateral                                     │
│          ▼                                                                  │
│  ┌───────────────┐                                                          │
│  │   Validator   │     4. Verification Steps:                               │
│  │   Verifier    │     ├─ Verify QE signature                               │
│  │               │     ├─ Verify certificate chain to Intel root            │
│  │               │     ├─ Check TCB status (not revoked)                    │
│  │               │     ├─ Validate MRENCLAVE against allowlist              │
│  │               │     ├─ Check nonce in ReportData                         │
│  │               │     └─ Verify debug flag = 0                             │
│  └───────────────┘                                                          │
│                                                                             │
│  Quote Format (SGX DCAP v3):                                                │
│  ┌──────────────────────────────────────────────────────────┐              │
│  │ Version (2B) │ AttKeyType (2B) │ Reserved (4B) │ QE SVN  │              │
│  ├──────────────────────────────────────────────────────────┤              │
│  │                    QE Vendor ID (16B)                    │              │
│  ├──────────────────────────────────────────────────────────┤              │
│  │                      User Data (20B)                     │              │
│  ├──────────────────────────────────────────────────────────┤              │
│  │                     Report Body (384B)                   │              │
│  │   ├─ CPUSVN (16B)    ├─ MiscSelect (4B)                 │              │
│  │   ├─ Attributes (16B) ├─ MRENCLAVE (32B)                │              │
│  │   ├─ MRSIGNER (32B)   ├─ ISV ProdID (2B)                │              │
│  │   ├─ ISV SVN (2B)     └─ ReportData (64B)               │              │
│  ├──────────────────────────────────────────────────────────┤              │
│  │                Signature Data (Variable)                 │              │
│  └──────────────────────────────────────────────────────────┘              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### AMD SEV-SNP Attestation Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SEV-SNP ATTESTATION FLOW                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────┐                           ┌───────────────────────────┐  │
│  │   Validator   │                           │     AMD KDS (Key          │  │
│  │   (Relying    │                           │     Distribution Server)  │  │
│  │    Party)     │                           └───────────────────────────┘  │
│  └───────┬───────┘                                      ▲                   │
│          │                                              │                   │
│          │ 1. Request attestation report                │                   │
│          │    (challenge nonce)                         │                   │
│          ▼                                              │                   │
│  ┌───────────────────────────────────────────────────────────────────┐     │
│  │                    CONFIDENTIAL VM (SNP)                          │     │
│  │  ┌────────────────────────────────────────────────────────────┐   │     │
│  │  │                     GUEST OS                                │   │     │
│  │  │  ┌───────────────┐                                         │   │     │
│  │  │  │ VEID Service  │  2. SNP_GUEST_REQUEST ioctl              │   │     │
│  │  │  │               │     └─▶ MSG_REPORT_REQ                   │   │     │
│  │  │  └───────┬───────┘                                         │   │     │
│  │  │          │                                                  │   │     │
│  │  │          ▼                                                  │   │     │
│  │  │  ┌───────────────┐                                         │   │     │
│  │  │  │ /dev/sev-guest│  3. Firmware generates report            │   │     │
│  │  │  │  device       │     └─▶ Signed by VCEK                   │   │     │
│  │  │  └───────────────┘                                         │   │     │
│  │  └────────────────────────────────────────────────────────────┘   │     │
│  │                                                                    │     │
│  │  Measurement Inputs:                                               │     │
│  │  ├─ Guest Policy (migration, debug, SMT policies)                 │     │
│  │  ├─ Launch Digest (initial memory hash at launch)                 │     │
│  │  ├─ ID Block/Auth (optional identity binding)                     │     │
│  │  └─ Host Data (hypervisor-provided data)                          │     │
│  └───────────────────────────────────────────────────────────────────┘     │
│          │                                                                  │
│          │ 4. Return Attestation Report                                     │
│          ▼                                                                  │
│  ┌───────────────┐                                                          │
│  │   Validator   │     5. Verification Steps:                               │
│  │   Verifier    │     ├─ Fetch VCEK cert from AMD KDS                      │
│  │               │     ├─ Verify VCEK chain to AMD ASK/ARK                  │
│  │               │     ├─ Verify report signature with VCEK                 │
│  │               │     ├─ Validate launch measurement                       │
│  │               │     ├─ Check guest policy (debug=0, etc.)                │
│  │               │     ├─ Verify TCB version ≥ minimum                      │
│  │               │     └─ Check nonce in REPORT_DATA                        │
│  └───────────────┘                                                          │
│                                                                             │
│  SNP Attestation Report Format (Version 2):                                 │
│  ┌──────────────────────────────────────────────────────────────┐          │
│  │ Version (4B) │ Guest SVN (4B) │ Policy (8B) │ Family ID (16B)│          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                      Image ID (16B)                          │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                 Current TCB Version (8B)                     │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │              Platform Info / Flags (8B)                      │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                  Launch Measurement (48B)                    │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                     Report Data (64B)                        │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                       Chip ID (64B)                          │          │
│  ├──────────────────────────────────────────────────────────────┤          │
│  │                   ECDSA Signature (512B)                     │          │
│  └──────────────────────────────────────────────────────────────┘          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Attestation Verification in Consensus

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              CONSENSUS ATTESTATION VERIFICATION FLOW                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                         BLOCK PROPOSAL                                │ │
│  │                                                                       │ │
│  │  Proposer Validator:                                                  │ │
│  │  1. Receive encrypted identity payload                                │ │
│  │  2. Process in TEE enclave:                                          │ │
│  │     ├─ Decrypt payload                                               │ │
│  │     ├─ Run ML scoring                                                │ │
│  │     ├─ Generate result hash                                          │ │
│  │     └─ Sign with enclave key                                         │ │
│  │  3. Generate attestation report (bind result to TEE)                 │ │
│  │  4. Include in block proposal:                                       │ │
│  │     ├─ Score result                                                  │ │
│  │     ├─ Enclave signature                                             │ │
│  │     └─ Attestation reference                                         │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                               │                                             │
│                               ▼                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      OTHER VALIDATORS                                 │ │
│  │                                                                       │ │
│  │  For each validator:                                                  │ │
│  │  1. Verify proposer's attestation report:                            │ │
│  │     ├─ Check measurement in governance allowlist                     │ │
│  │     ├─ Verify attestation signature                                  │ │
│  │     └─ Confirm report freshness (nonce/timestamp)                    │ │
│  │                                                                       │ │
│  │  2. Recompute in own TEE enclave:                                    │ │
│  │     ├─ Decrypt same payload with own key                             │ │
│  │     ├─ Run same ML scoring pipeline                                  │ │
│  │     └─ Compare result hash                                           │ │
│  │                                                                       │ │
│  │  3. Vote decision:                                                   │ │
│  │     ├─ PREVOTE if: attestation valid AND results match               │ │
│  │     └─ NO VOTE if: attestation invalid OR results mismatch           │ │
│  │                                                                       │ │
│  │  4. Include own attestation in vote (for auditability)              │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  Slashing Conditions:                                                       │
│  ├─ Invalid attestation (bad signature): 5% slash + jail                   │
│  ├─ Non-allowlisted measurement: 10% slash + indefinite jail               │
│  ├─ Debug mode enabled: 20% slash + removal                                │
│  └─ Consistent result mismatch: 2% slash + investigation                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Key Derivation Design

### Key Hierarchy Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         KEY DERIVATION HIERARCHY                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    HARDWARE ROOT OF TRUST                             │ │
│  │                                                                       │ │
│  │  ┌─────────────────────────┐    ┌─────────────────────────┐          │ │
│  │  │   SGX: Root Seal Key   │    │  SEV: AMD Root Key (ARK) │          │ │
│  │  │   (CPU fuse-based)     │    │  (PSP-protected)         │          │ │
│  │  └───────────┬─────────────┘    └───────────┬─────────────┘          │ │
│  │              │                               │                        │ │
│  │              ▼                               ▼                        │ │
│  │  ┌─────────────────────────┐    ┌─────────────────────────┐          │ │
│  │  │ MRENCLAVE-bound Seal   │    │   VCEK (per-CPU/TCB)     │          │ │
│  │  │ Key (per-enclave)      │    │   Signing Key            │          │ │
│  │  └───────────┬─────────────┘    └───────────┬─────────────┘          │ │
│  └──────────────┼──────────────────────────────┼────────────────────────┘ │
│                 │                               │                          │
│                 ▼                               ▼                          │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    ENCLAVE MASTER KEY (EMK)                           │ │
│  │                                                                       │ │
│  │  Generation:                                                          │ │
│  │  - SGX: EGETKEY with key_policy = MRENCLAVE | MRSIGNER                │ │
│  │  - SEV: Derived from launch secret + platform key                     │ │
│  │                                                                       │ │
│  │  Properties:                                                          │ │
│  │  - Never leaves TEE boundary                                          │ │
│  │  - Bound to specific enclave measurement                              │ │
│  │  - Rotated on enclave version upgrade                                 │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                    │                                       │
│          ┌─────────────────────────┼─────────────────────────┐             │
│          │                         │                         │             │
│          ▼                         ▼                         ▼             │
│  ┌───────────────┐       ┌───────────────┐       ┌───────────────┐        │
│  │ Encryption    │       │ Signing       │       │ Sealing       │        │
│  │ Key Pair      │       │ Key Pair      │       │ Key           │        │
│  │               │       │               │       │               │        │
│  │ Algorithm:    │       │ Algorithm:    │       │ Algorithm:    │        │
│  │ X25519        │       │ Ed25519       │       │ AES-256-GCM   │        │
│  │               │       │               │       │               │        │
│  │ Purpose:      │       │ Purpose:      │       │ Purpose:      │        │
│  │ Decrypt       │       │ Sign scoring  │       │ Seal state    │        │
│  │ identity      │       │ results for   │       │ to disk for   │        │
│  │ payloads      │       │ consensus     │       │ persistence   │        │
│  │               │       │               │       │               │        │
│  │ Derivation:   │       │ Derivation:   │       │ Derivation:   │        │
│  │ HKDF(EMK,     │       │ HKDF(EMK,     │       │ HKDF(EMK,     │        │
│  │  "enc"|epoch) │       │  "sig"|epoch) │       │  "seal")      │        │
│  └───────────────┘       └───────────────┘       └───────────────┘        │
│                                                                             │
│  Key Rotation:                                                              │
│  - Triggered every 1000 blocks (configurable)                              │
│  - Old keys kept for 100-block grace period                                │
│  - Public keys registered on-chain for multi-recipient encryption          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Derivation Algorithm

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      KEY DERIVATION FUNCTION (KDF)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Input:                                                                     │
│    EMK         = Enclave Master Key (32 bytes, hardware-derived)           │
│    purpose     = "encryption" | "signing" | "sealing"                       │
│    epoch       = Current key rotation epoch (uint64)                        │
│    context     = Platform-specific context (measurement, etc.)              │
│                                                                             │
│  Algorithm (HKDF-SHA256):                                                   │
│                                                                             │
│    // Step 1: Extract                                                       │
│    salt = SHA256(platform || measurement || version)                        │
│    PRK = HKDF-Extract(salt, EMK)                                           │
│                                                                             │
│    // Step 2: Expand for each key type                                     │
│    info = purpose || epoch || context                                       │
│    derived_key = HKDF-Expand(PRK, info, key_length)                        │
│                                                                             │
│  Key Lengths:                                                               │
│    X25519 secret:  32 bytes → compute public key                           │
│    Ed25519 seed:   32 bytes → derive keypair                               │
│    AES-256-GCM:    32 bytes                                                │
│                                                                             │
│  Code Example:                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  func DeriveKeyPair(emk []byte, purpose string, epoch uint64) {     │   │
│  │      salt := sha256(platform + measurement + version)               │   │
│  │      prk := hkdf.Extract(sha256.New, emk, salt)                     │   │
│  │      info := []byte(purpose + strconv.FormatUint(epoch, 10))        │   │
│  │      reader := hkdf.Expand(sha256.New, prk, info)                   │   │
│  │      seed := make([]byte, 32)                                       │   │
│  │      reader.Read(seed)                                              │   │
│  │      return ed25519.NewKeyFromSeed(seed)                            │   │
│  │  }                                                                   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Sealed Storage Architecture

### Sealed Data Format

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SEALED STORAGE FORMAT                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      SEALED BLOB HEADER (64 bytes)                    │ │
│  │                                                                       │ │
│  │  ┌─────────────┬─────────────┬─────────────┬─────────────────────┐   │ │
│  │  │ Magic (4B)  │ Version (2B)│ Platform(2B)│ Key Policy (8B)     │   │ │
│  │  │ "SEAL"      │ 0x0001      │ SGX/SEV     │ Seal to MRENCLAVE   │   │ │
│  │  └─────────────┴─────────────┴─────────────┴─────────────────────┘   │ │
│  │  ┌─────────────┬─────────────┬─────────────────────────────────────┐ │ │
│  │  │ Epoch (8B)  │ Nonce (12B) │ Measurement Hash (32B)              │ │ │
│  │  └─────────────┴─────────────┴─────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                      ENCRYPTED PAYLOAD                                │ │
│  │                                                                       │ │
│  │  Algorithm: AES-256-GCM                                               │ │
│  │  Key: Derived from sealing key + epoch                               │ │
│  │                                                                       │ │
│  │  ┌───────────────────────────────────────────────────────────────┐   │ │
│  │  │                   Ciphertext (Variable)                       │   │ │
│  │  │  ┌─────────────────────────────────────────────────────────┐  │   │ │
│  │  │  │  Plaintext Contents:                                    │  │   │ │
│  │  │  │  - EMK (32 bytes)                                       │  │   │ │
│  │  │  │  - Encryption private key (32 bytes)                    │  │   │ │
│  │  │  │  - Signing private key (64 bytes)                       │  │   │ │
│  │  │  │  - Monotonic counter value (8 bytes)                    │  │   │ │
│  │  │  │  - State checksum (32 bytes)                            │  │   │ │
│  │  │  └─────────────────────────────────────────────────────────┘  │   │ │
│  │  └───────────────────────────────────────────────────────────────┘   │ │
│  │  ┌───────────────────────────────────────────────────────────────┐   │ │
│  │  │                   Auth Tag (16 bytes)                         │   │ │
│  │  └───────────────────────────────────────────────────────────────┘   │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  Unsealing Requirements:                                                    │
│  ├─ Same enclave measurement (MRENCLAVE/launch digest)                     │
│  ├─ Same security version number (SVN) or higher                           │
│  ├─ Valid platform key hierarchy                                           │
│  └─ Monotonic counter ≥ sealed value (anti-rollback)                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Anti-Rollback Mechanism

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      ANTI-ROLLBACK PROTECTION                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Problem: Attacker restores old sealed state to replay old keys/data       │
│                                                                             │
│  Solution: Bind state to monotonic counter                                  │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                    MONOTONIC COUNTER SOURCES                          │ │
│  │                                                                       │ │
│  │  Platform   │ Source                  │ Security                      │ │
│  │  ───────────┼─────────────────────────┼────────────────────────────── │ │
│  │  SGX        │ Intel ME Monotonic      │ Hardware-backed               │ │
│  │             │ Counter (RPDB)          │                               │ │
│  │  SEV-SNP    │ Platform State          │ Hypervisor-backed             │ │
│  │             │ + Blockchain height     │ (supplemented)                │ │
│  │  Hybrid     │ Blockchain block height │ Consensus-backed              │ │
│  │             │ as minimum bound        │                               │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  State Update Flow:                                                         │
│                                                                             │
│  1. On every key operation:                                                 │
│     counter = max(platform_counter, last_blockchain_height)                 │
│     state.counter = counter + 1                                            │
│     seal(state)                                                            │
│                                                                             │
│  2. On unseal:                                                              │
│     if state.counter < platform_counter:                                   │
│         reject("rollback detected")                                        │
│     if state.counter < last_blockchain_height - grace_period:              │
│         reject("state too old")                                            │
│                                                                             │
│  3. On enclave restart:                                                     │
│     fetch current_block_height from consensus                              │
│     unseal(state)                                                          │
│     if state.last_height < current_block_height - max_gap:                 │
│         require_reregistration()                                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Hardware Requirements Matrix

### Validator Hardware Requirements

| Component | Intel SGX (Minimum) | Intel SGX (Recommended) | AMD SEV-SNP (Minimum) | AMD SEV-SNP (Recommended) |
|-----------|---------------------|------------------------|----------------------|--------------------------|
| **CPU** | Xeon Scalable 3rd Gen | Xeon 4th Gen (Sapphire Rapids) | EPYC 7003 (Milan) | EPYC 9004 (Genoa) |
| **CPU Cores** | 8 cores | 16+ cores | 8 cores | 16+ cores |
| **EPC Memory** | 128 MB | 512 MB | N/A | N/A |
| **System RAM** | 32 GB | 64 GB | 32 GB | 64 GB |
| **Storage** | 500 GB NVMe | 1 TB NVMe | 500 GB NVMe | 1 TB NVMe |
| **Network** | 1 Gbps | 10 Gbps | 1 Gbps | 10 Gbps |
| **Firmware** | SGX DCAP support | Latest microcode | AGESA 1.0.0.9+ | Latest firmware |
| **OS** | Ubuntu 22.04 LTS | Ubuntu 24.04 LTS | Ubuntu 24.04 LTS | Ubuntu 24.04 LTS |
| **Kernel** | 5.11+ with SGX driver | 6.6+ | 6.0+ with SNP patches | 6.6+ |

### Cloud Provider Compatibility

| Provider | TEE Type | Instance Family | vCPUs | Memory | Monthly Cost |
|----------|----------|----------------|-------|--------|--------------|
| **Azure** | SGX | DCsv3 | 4-48 | 16-384 GB | $350-$4,200 |
| **Azure** | SEV-SNP | DCasv5 | 4-96 | 16-384 GB | $300-$3,600 |
| **AWS** | Nitro | c6a.xlarge+ | 4-192 | 8-384 GB | $250-$3,000 |
| **GCP** | SEV | n2d-standard-4+ | 4-224 | 4-896 GB | $280-$3,400 |
| **IBM Cloud** | SGX | bx2d | 2-128 | 8-512 GB | $200-$2,800 |

### TEE Feature Availability by Platform

| Feature | SGX (DCsv3) | SEV-SNP (DCasv5) | Nitro | Notes |
|---------|-------------|------------------|-------|-------|
| Memory Encryption | ✅ | ✅ | ✅ | All platforms |
| Remote Attestation | ✅ DCAP | ✅ SNP Report | ✅ NSM | Different protocols |
| Sealed Storage | ✅ | ✅ | ✅ KMS | KMS-backed for Nitro |
| Memory Integrity | ❌ | ✅ (SNP) | ✅ | SGX lacks integrity |
| Hot Migration | ❌ | ⚠️ (planned) | ❌ | Limited support |
| Nested Virt | ❌ | ⚠️ | ❌ | Limited support |
| Debug Support | ⚠️ (disabled prod) | ⚠️ (disabled prod) | ⚠️ | Disabled in production |

---

## Implementation Strategy

### Phase 1: Interface Definition (Week 1-2)

```go
// pkg/enclave_runtime/tee_interface.go

// TEEService defines the complete interface for TEE operations
type TEEService interface {
    // Lifecycle
    Initialize(config TEEConfig) error
    Shutdown() error
    GetStatus() TEEStatus

    // Attestation
    GenerateAttestationReport(challenge []byte) (*AttestationReport, error)
    VerifyAttestationReport(report *AttestationReport) error

    // Key Management
    GetEncryptionPublicKey() ([]byte, error)
    GetSigningPublicKey() ([]byte, error)
    RotateKeys(epoch uint64) error

    // Identity Processing
    DecryptAndScore(ctx context.Context, req *ScoringRequest) (*ScoringResult, error)

    // Sealed Storage
    SealState(data []byte) ([]byte, error)
    UnsealState(sealed []byte) ([]byte, error)
}
```

### Phase 2: SGX POC (Week 3-5)

```
Tasks:
├─ Set up Gramine development environment
├─ Create enclave manifest for TensorFlow Lite
├─ Implement DCAP attestation flow
├─ Implement key derivation using EGETKEY
├─ Implement sealed storage using sgx_seal_data
├─ Benchmark EPC pressure with ML model
└─ Create integration tests with simulated QE
```

### Phase 3: SEV-SNP POC (Week 4-6)

```
Tasks:
├─ Set up Azure DCasv5 development VM
├─ Configure SNP attestation (sev-guest device)
├─ Implement launch measurement verification
├─ Implement key derivation from vTPM
├─ Implement sealed storage via encrypted disk
├─ Port ML scoring pipeline to CVM
└─ Create attestation verification client
```

### Phase 4: Integration (Week 7-9)

```
Tasks:
├─ Integrate TEEService with VEID keeper
├─ Add enclave measurement governance
├─ Implement dual-mode operation (simulated + real)
├─ Update genesis for TEE configuration
├─ Add attestation to consensus votes
└─ Implement slashing for invalid attestation
```

### Phase 5: Hardening (Week 10-12)

```
Tasks:
├─ Security audit preparation
├─ Side-channel mitigation review
├─ Performance benchmarking
├─ Documentation for validators
├─ Testnet deployment
└─ Migration dry-run
```

---

## Appendix A: Platform-Specific API References

### Intel SGX SDK

```c
// Key derivation
sgx_status_t sgx_get_key(
    const sgx_key_request_t *key_request,
    sgx_key_128bit_t *key
);

// Sealing
sgx_status_t sgx_seal_data(
    uint32_t additional_MACtext_length,
    const uint8_t *p_additional_MACtext,
    uint32_t text2encrypt_length,
    const uint8_t *p_text2encrypt,
    uint32_t sealed_data_size,
    sgx_sealed_data_t *p_sealed_data
);

// Attestation (DCAP)
quote3_error_t sgx_qe_get_quote(
    const sgx_report_t *p_app_report,
    uint32_t quote_size,
    uint8_t *p_quote
);
```

### AMD SEV-SNP (Linux)

```c
// /dev/sev-guest ioctl interface
struct snp_guest_request_ioctl {
    __u8 msg_version;
    __u64 req_data;
    __u64 resp_data;
    __u64 fw_err;
};

// MSG_REPORT_REQ
struct snp_report_req {
    __u8 user_data[64];
    __u32 vmpl;
    __u8 reserved[28];
};

// MSG_REPORT_RESP
struct snp_report_resp {
    struct snp_attestation_report report;
    __u8 reserved[64];
};
```

---

## Appendix B: Related Documents

- [TEE Security Model](./tee-security-model.md)
- [TEE Integration Plan](./tee-integration-plan.md)
- [TEE Migration Plan](./tee-migration-plan.md)
- [VEID Flow Specification](./veid-flow-spec.md)
- [Validator Hardware Guide](./validator-hardware-guide.md)
