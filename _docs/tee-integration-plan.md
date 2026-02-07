# VirtEngine TEE Integration Plan

**Version:** 1.0.0  
**Date:** 2026-01-29  
**Status:** Planning Document  
**Task Reference:** VE-2023

---

## Executive Summary

This document outlines the plan for integrating real Trusted Execution Environment (TEE) support into VirtEngine's VEID identity verification system. The current `SimulatedEnclaveService` provides **NO security guarantees** and must be replaced with actual TEE implementations before mainnet launch.

---

## 1. Technology Comparison: Intel SGX vs AMD SEV-SNP

### 1.1 Intel SGX (Software Guard Extensions)

| Aspect                | Details                                                          |
| --------------------- | ---------------------------------------------------------------- |
| **Security Model**    | Application-level enclaves with encrypted memory regions         |
| **Memory Encryption** | MEE (Memory Encryption Engine) protects enclave pages            |
| **Attestation**       | DCAP (Data Center Attestation Primitives) for remote attestation |
| **Enclave Size**      | Limited EPC (Enclave Page Cache) - typically 128MB-256MB         |
| **Performance**       | Overhead on enclave entry/exit, but fast once inside             |
| **SDK**               | Intel SGX SDK (C/C++), Rust SGX SDK (Teaclave), Gramine          |
| **Cloud Support**     | Azure Confidential Computing, Alibaba, IBM                       |
| **Hardware**          | Intel Xeon (3rd Gen+), select client CPUs                        |

**Pros:**

- Mature ecosystem with extensive documentation
- Small TCB (Trusted Computing Base)
- Strong isolation at application level
- Gramine LibOS enables running unmodified applications

**Cons:**

- Limited EPC memory (paging overhead for large datasets)
- Complex SDK with C/C++ focus
- Recent side-channel vulnerabilities (patched but ongoing research)
- Intel-only hardware lock-in

### 1.2 AMD SEV-SNP (Secure Encrypted Virtualization - Secure Nested Paging)

| Aspect                | Details                                              |
| --------------------- | ---------------------------------------------------- |
| **Security Model**    | VM-level encryption with memory integrity protection |
| **Memory Encryption** | SME/SEV encrypts all VM memory transparently         |
| **Attestation**       | SNP attestation reports with versioned TCB           |
| **Memory Size**       | Full VM memory (no EPC limitations)                  |
| **Performance**       | Near-native performance, minimal overhead            |
| **SDK**               | Linux kernel support, SVSM, Coconut-SVSM             |
| **Cloud Support**     | AWS (Nitro), Azure, Google Cloud                     |
| **Hardware**          | AMD EPYC (3rd Gen Milan+)                            |

**Pros:**

- No memory size limitations
- Simpler porting (standard Linux/VM)
- Better performance for large workloads
- Stronger memory integrity guarantees (SNP)

**Cons:**

- Larger TCB (entire guest OS)
- Newer technology with less ecosystem maturity
- Requires VM-based deployment model

### 1.3 Recommendation for VEID

**Primary: AMD SEV-SNP** for the following reasons:

1. **Memory Requirements**: VEID ML models and identity processing may exceed SGX EPC limits
2. **Deployment Simplicity**: Validators can run standard Linux VMs with SEV-SNP enabled
3. **Performance**: Near-native performance critical for consensus timing
4. **Cloud Availability**: Broad cloud support (AWS Nitro, Azure CVM, GCP Confidential VM)

**Secondary: Intel SGX (Gramine)** for:

- Validators with Intel hardware only
- Smaller enclaves for key management operations
- Environments where VM-level isolation is insufficient

---

## 2. Architecture Design

### 2.1 High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Validator Node Host                       │
├──────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐  ┌────────────────────────────────┐ │
│  │   VirtEngine Node   │  │     TEE Enclave (SEV-SNP)      │ │
│  │   (Untrusted)       │  │     (Trusted)                  │ │
│  │                     │  │                                │ │
│  │  ┌───────────────┐  │  │  ┌──────────────────────────┐  │ │
│  │  │ VEID Keeper   │──┼──┼─▶│ Identity Scorer          │  │ │
│  │  │ (gRPC Client) │  │  │  │ - Decrypt payloads       │  │ │
│  │  └───────────────┘  │  │  │ - Run ML model           │  │ │
│  │                     │  │  │ - Sign results           │  │ │
│  │  ┌───────────────┐  │  │  └──────────────────────────┘  │ │
│  │  │ Attestation   │◀─┼──┼──│ Attestation Generator    │  │ │
│  │  │ Verifier      │  │  │  └──────────────────────────┘  │ │
│  │  └───────────────┘  │  │                                │ │
│  │                     │  │  ┌──────────────────────────┐  │ │
│  │                     │  │  │ Sealed Key Storage       │  │ │
│  │                     │  │  │ - Encryption keys        │  │ │
│  │                     │  │  │ - Signing keys           │  │ │
│  │                     │  │  └──────────────────────────┘  │ │
│  └─────────────────────┘  └────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Communication Protocol

```
┌─────────────┐     gRPC/TLS      ┌─────────────┐
│ VEID Keeper │◀───────────────▶ │ TEE Service │
└─────────────┘                   └─────────────┘
       │                                │
       ▼                                ▼
 1. ScoringRequest              2. Decrypt in enclave
    - encrypted_payload            - unwrap DEK
    - wrapped_key                  - decrypt payload
    - attestation_nonce            - run ML scorer
                                   - generate signature
       ▲                                │
       │                                ▼
 4. Verify attestation          3. ScoringResult
    - check measurement             - score
    - validate signature            - attestation_quote
    - store on-chain                - enclave_signature
```

### 2.3 Key Management

```
┌─────────────────────────────────────────────────────────────┐
│                    Key Hierarchy                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Platform Key (HW-bound, never exported)               │  │
│  │ - AMD PSP / Intel ME derived                          │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│                           ▼                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Enclave Master Key (Sealed to measurement)            │  │
│  │ - Generated on first boot                             │  │
│  │ - Unsealed only by matching enclave                   │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│           ┌───────────────┼───────────────┐                 │
│           ▼               ▼               ▼                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Encryption  │  │ Signing     │  │ Attestation │          │
│  │ Key Pair    │  │ Key Pair    │  │ Key Pair    │          │
│  │ (X25519)    │  │ (Ed25519)   │  │ (Platform)  │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Implementation Plan

### Phase 1: Foundation (2 weeks)

| Task  | Description                                 | Deliverable                                    |
| ----- | ------------------------------------------- | ---------------------------------------------- |
| 3.1.1 | Define `RealEnclaveService` interface       | Go interface matching current API              |
| 3.1.2 | Create gRPC proto for enclave communication | `enclave.proto` with scoring, attestation RPCs |
| 3.1.3 | Implement attestation verification          | Verify SEV-SNP/SGX quotes                      |
| 3.1.4 | Set up TEE development environment          | Azure CVM or local AMD EPYC server             |

### Phase 2: SEV-SNP POC (3 weeks)

| Task  | Description                            | Deliverable                          |
| ----- | -------------------------------------- | ------------------------------------ |
| 3.2.1 | Create minimal SEV-SNP guest image     | Linux image with attestation support |
| 3.2.2 | Implement key derivation in guest      | Platform-bound key generation        |
| 3.2.3 | Implement X25519 encryption/decryption | Decrypt identity payloads            |
| 3.2.4 | Port TensorFlow Lite scorer to guest   | Run ML model in confidential VM      |
| 3.2.5 | Generate SNP attestation reports       | Bind results to attestation          |

### Phase 3: SGX POC (2 weeks, parallel track)

| Task  | Description                       | Deliverable                  |
| ----- | --------------------------------- | ---------------------------- |
| 3.3.1 | Set up Gramine environment        | Gramine manifest for enclave |
| 3.3.2 | Package scorer as Gramine enclave | TensorFlow Lite in SGX       |
| 3.3.3 | Implement DCAP attestation        | Remote attestation flow      |
| 3.3.4 | Benchmark EPC pressure            | Validate memory requirements |

### Phase 4: Integration (2 weeks)

| Task  | Description                          | Deliverable                      |
| ----- | ------------------------------------ | -------------------------------- |
| 3.4.1 | Integrate with VEID keeper           | Replace SimulatedEnclaveService  |
| 3.4.2 | Add enclave measurement to genesis   | Allowlist valid measurements     |
| 3.4.3 | Implement enclave upgrade flow       | Hot-swap with measurement update |
| 3.4.4 | Add slashing for invalid attestation | Governance-controlled penalties  |

### Phase 5: Hardening (3 weeks)

| Task  | Description                    | Deliverable                |
| ----- | ------------------------------ | -------------------------- |
| 3.5.1 | Side-channel mitigations       | Constant-time operations   |
| 3.5.2 | Memory scrubbing audit         | Verify no plaintext leaks  |
| 3.5.3 | Fuzz testing enclave interface | Find edge cases            |
| 3.5.4 | Security audit preparation     | Documentation for auditors |
| 3.5.5 | Performance benchmarking       | Measure latency impact     |

---

## 4. Hardware Requirements for Validators

### 4.1 AMD SEV-SNP Requirements

| Component | Minimum                     | Recommended           |
| --------- | --------------------------- | --------------------- |
| CPU       | AMD EPYC 7003 (Milan)       | AMD EPYC 9004 (Genoa) |
| Memory    | 32 GB                       | 64 GB                 |
| Storage   | 500 GB NVMe                 | 1 TB NVMe             |
| Firmware  | Latest SEV firmware         | AGESA 1.0.0.9+        |
| OS        | Linux 6.0+ with SNP patches | Ubuntu 24.04 LTS      |

### 4.2 Intel SGX Requirements

| Component | Minimum                     | Recommended                          |
| --------- | --------------------------- | ------------------------------------ |
| CPU       | Intel Xeon Scalable 3rd Gen | Intel Xeon 4th Gen (Sapphire Rapids) |
| EPC       | 128 MB                      | 512 MB (with EPC expansion)          |
| Memory    | 32 GB                       | 64 GB                                |
| Storage   | 500 GB NVMe                 | 1 TB NVMe                            |
| Firmware  | SGX DCAP support            | Latest microcode                     |
| OS        | Linux 5.11+ with SGX driver | Ubuntu 22.04 LTS                     |

### 4.3 Cloud Provider Options

| Provider | TEE Type | Instance Type    | Monthly Cost (Est.) |
| -------- | -------- | ---------------- | ------------------- |
| Azure    | SEV-SNP  | DCasv5 (4 vCPU)  | ~$300               |
| Azure    | SGX      | DCsv3 (4 vCPU)   | ~$350               |
| AWS      | Nitro    | c6a.xlarge (SEV) | ~$250               |
| GCP      | SEV      | n2d-standard-4   | ~$280               |

---

## 5. Migration Plan

### 5.1 Migration Phases

```
Phase 1: Dual Mode (Simulated + Real)
├── SimulatedEnclaveService remains default
├── RealEnclaveService opt-in via config
├── Both produce compatible results
└── Validators can test TEE without commitment

Phase 2: Testnet Validation
├── Testnet requires TEE for scoring
├── SimulatedEnclaveService disabled on testnet
├── Measure performance and reliability
└── Fix issues before mainnet

Phase 3: Mainnet Transition
├── Grace period: Both modes accepted
├── Governance proposal to require TEE
├── SimulatedEnclaveService deprecated
└── Non-TEE validators lose scoring rights

Phase 4: TEE-Only
├── Remove SimulatedEnclaveService
├── All scoring requires valid attestation
├── Slashing for invalid attestations
└── Full security guarantees active
```

### 5.2 Configuration Changes

```yaml
# config/app.toml

[enclave]
# "simulated" | "sgx" | "sev-snp" | "nitro"
type = "sev-snp"

# gRPC endpoint for TEE service (if remote)
endpoint = "unix:///var/run/veid-enclave.sock"

# Required measurements (hex-encoded)
allowed_measurements = [
  "a1b2c3d4...",  # v1.0.0
  "e5f6g7h8...",  # v1.0.1
]

# Attestation verification mode
# "strict" = reject unknown measurements
# "permissive" = log warning but accept (testnet only)
attestation_mode = "strict"

# Cache attestation reports for this duration
attestation_cache_seconds = 300
```

---

## 6. Timeline Summary

| Phase       | Duration     | End Date        | Milestone              |
| ----------- | ------------ | --------------- | ---------------------- |
| Foundation  | 2 weeks      | Feb 12, 2026    | Interfaces defined     |
| SEV-SNP POC | 3 weeks      | Mar 5, 2026     | POC running in CVM     |
| SGX POC     | 2 weeks      | Mar 5, 2026     | POC in Gramine         |
| Integration | 2 weeks      | Mar 19, 2026    | Integrated with keeper |
| Hardening   | 3 weeks      | Apr 9, 2026     | Security audit ready   |
| **Total**   | **12 weeks** | **Apr 9, 2026** | **Production TEE**     |

---

## 7. Risks and Mitigations

| Risk                           | Probability | Impact   | Mitigation                 |
| ------------------------------ | ----------- | -------- | -------------------------- |
| SEV-SNP attestation complexity | Medium      | High     | Fallback to SGX if needed  |
| ML model too large for SGX EPC | High        | Medium   | Use SEV-SNP for scoring    |
| Cloud provider availability    | Low         | High     | Support multiple TEE types |
| Side-channel vulnerabilities   | Medium      | Critical | Follow vendor mitigations  |
| Validator hardware adoption    | Medium      | High     | Long migration period      |

---

## 8. References

1. [AMD SEV-SNP Architecture](https://www.amd.com/en/developer/sev.html)
2. [Intel SGX Documentation](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)
3. [Gramine LibOS](https://gramine.readthedocs.io/)
4. [Azure Confidential Computing](https://azure.microsoft.com/en-us/solutions/confidential-compute/)
5. [AWS Nitro Enclaves](https://aws.amazon.com/ec2/nitro/nitro-enclaves/)
6. [VirtEngine TEE Security Model](tee-security-model.md)

---

## Appendix A: Enclave Interface Definition

```go
// pkg/enclave_runtime/real_enclave.go

// RealEnclaveService connects to an actual TEE implementation
type RealEnclaveService interface {
    EnclaveService

    // GetAttestationReport generates a platform-specific attestation report
    GetAttestationReport(nonce []byte) (*AttestationReport, error)

    // VerifyPeerAttestation verifies another enclave's attestation
    VerifyPeerAttestation(report *AttestationReport) error

    // GetPlatformType returns the TEE platform type
    GetPlatformType() PlatformType
}

type PlatformType string

const (
    PlatformSimulated PlatformType = "simulated"
    PlatformSGX       PlatformType = "sgx"
    PlatformSEVSNP    PlatformType = "sev-snp"
    PlatformNitro     PlatformType = "nitro"
)

type AttestationReport struct {
    Platform        PlatformType
    Measurement     []byte
    ReportData      []byte
    PlatformInfo    []byte
    Signature       []byte
    CertChain       [][]byte
    Timestamp       time.Time
}
```
