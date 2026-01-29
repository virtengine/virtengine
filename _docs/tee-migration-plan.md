# TEE Migration Plan: SimulatedEnclaveService to Production TEE

**Version:** 1.0.0  
**Date:** 2026-01-29  
**Status:** Planning Document  
**Task Reference:** VE-2023

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Migration Phases](#migration-phases)
4. [Validator Migration Checklist](#validator-migration-checklist)
5. [Rollback Procedures](#rollback-procedures)
6. [Success Criteria](#success-criteria)
7. [Timeline](#timeline)

---

## Executive Summary

This document outlines the migration plan from VirtEngine's current `SimulatedEnclaveService` (which provides **NO security guarantees**) to production-ready Trusted Execution Environment (TEE) implementations using Intel SGX and AMD SEV-SNP.

### Critical Warning

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ⚠️  SECURITY ALERT: SimulatedEnclaveService is NOT SECURE                  ║
║                                                                              ║
║  The current SimulatedEnclaveService:                                        ║
║  - Does NOT provide memory isolation                                         ║
║  - Does NOT encrypt identity data                                            ║
║  - Does NOT prevent host access to plaintext                                 ║
║  - Does NOT provide valid attestation                                        ║
║                                                                              ║
║  This service MUST be replaced before mainnet launch.                        ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

### Migration Goals

1. **Zero Downtime**: Validators can migrate without consensus interruption
2. **Gradual Rollout**: Phased approach with clear checkpoints
3. **Reversibility**: Each phase includes rollback procedures
4. **Verification**: Automated testing at each phase gate
5. **Security**: No regression in security posture during migration

---

## Current State Analysis

### SimulatedEnclaveService Capabilities

| Capability | Simulated | Production TEE |
|------------|-----------|----------------|
| Interface Compliance | ✅ | ✅ |
| Memory Isolation | ❌ | ✅ |
| Encrypted Memory | ❌ | ✅ |
| Remote Attestation | ❌ (mock) | ✅ |
| Sealed Key Storage | ❌ (plaintext) | ✅ |
| Anti-Rollback | ❌ | ✅ |
| Production Use | ❌ | ✅ |

### Files to Migrate

```
pkg/enclave_runtime/
├── enclave_service.go      # Interface + SimulatedEnclaveService
├── enclave_service_test.go # Tests
├── real_enclave.go         # POC stubs (SGX, SEV-SNP, Nitro)
├── real_enclave_test.go    # POC tests
├── sgx_enclave.go          # NEW: SGX implementation (POC)
├── sev_enclave.go          # NEW: SEV-SNP implementation (POC)
├── memory_scrub.go         # Memory scrubbing utilities
├── memory_scrub_test.go    # Scrub tests
├── privacy_controls.go     # Privacy configuration
└── privacy_controls_test.go
```

### Integration Points

```
x/veid/
├── keeper/keeper.go        # Uses EnclaveService interface
└── keeper/scoring.go       # Calls enclave.Score()

app/
├── app.go                  # Enclave configuration
└── modules.go              # Module wiring
```

---

## Migration Phases

### Phase 1: Dual Mode Infrastructure (Week 1-2)

**Goal**: Enable running both simulated and real TEE services simultaneously

#### 1.1 Configuration Schema

```yaml
# app.toml - Enhanced enclave configuration

[enclave]
# Primary enclave type: "simulated" | "sgx" | "sev-snp" | "nitro"
type = "simulated"

# Fallback enclave (for gradual migration)
fallback_type = ""

# Dual mode: Run both and compare results (testnet only)
dual_mode = false
dual_mode_compare = true  # Log discrepancies

# SGX-specific settings
[enclave.sgx]
enabled = false
enclave_path = "/opt/virtengine/enclave/veid_scorer.signed.so"
dcap_enabled = true
quote_provider = "azure"  # "azure" | "intel" | "custom"
debug = false  # MUST be false for production

# SEV-SNP-specific settings
[enclave.sev_snp]
enabled = false
endpoint = "unix:///var/run/veid-enclave.sock"
cert_chain_path = "/opt/virtengine/certs/amd_chain.pem"
vcek_cache_path = "/var/cache/virtengine/vcek"
min_tcb_version = "2.0.8.115"
allow_debug_policy = false  # MUST be false for production

# Attestation settings
[enclave.attestation]
# Verification mode: "strict" | "permissive" | "disabled"
# - strict: Reject unknown measurements (production)
# - permissive: Log warning but accept (testnet)
# - disabled: Skip attestation (development only)
mode = "permissive"

# Allowed enclave measurements (hex-encoded)
allowed_measurements = []

# Cache attestation reports for this duration (seconds)
cache_duration = 300

# Maximum attestation age (seconds)
max_report_age = 3600
```

#### 1.2 Implementation Changes

```go
// pkg/enclave_runtime/factory.go (new file)

// EnclaveFactory creates enclave services based on configuration
type EnclaveFactory struct {
    config    EnclaveConfig
    primary   EnclaveService
    fallback  EnclaveService
    dualMode  bool
    logger    log.Logger
}

// NewEnclaveFactory creates a factory with the given configuration
func NewEnclaveFactory(config EnclaveConfig, logger log.Logger) (*EnclaveFactory, error) {
    factory := &EnclaveFactory{
        config:   config,
        dualMode: config.DualMode,
        logger:   logger,
    }

    // Create primary enclave
    primary, err := CreateEnclaveService(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create primary enclave: %w", err)
    }
    factory.primary = primary

    // Create fallback enclave if configured
    if config.FallbackType != "" {
        fallbackConfig := config
        fallbackConfig.Platform = config.FallbackType
        fallback, err := CreateEnclaveService(fallbackConfig)
        if err != nil {
            logger.Error("failed to create fallback enclave", "error", err)
        } else {
            factory.fallback = fallback
        }
    }

    return factory, nil
}

// Score performs scoring with optional dual-mode comparison
func (f *EnclaveFactory) Score(ctx context.Context, req *ScoringRequest) (*ScoringResult, error) {
    // Score with primary enclave
    result, err := f.primary.Score(ctx, req)
    if err != nil && f.fallback != nil {
        // Fallback on primary failure
        f.logger.Warn("primary enclave failed, using fallback",
            "error", err,
            "request_id", req.RequestID)
        return f.fallback.Score(ctx, req)
    }

    // Dual mode comparison
    if f.dualMode && f.fallback != nil && err == nil {
        go f.compareDualMode(ctx, req, result)
    }

    return result, err
}

// compareDualMode compares results between primary and fallback
func (f *EnclaveFactory) compareDualMode(ctx context.Context, req *ScoringRequest, primaryResult *ScoringResult) {
    fallbackResult, err := f.fallback.Score(ctx, req)
    if err != nil {
        f.logger.Warn("dual mode fallback failed", "error", err)
        return
    }

    // Compare scores
    if primaryResult.Score != fallbackResult.Score {
        f.logger.Error("DUAL MODE SCORE MISMATCH",
            "request_id", req.RequestID,
            "primary_score", primaryResult.Score,
            "fallback_score", fallbackResult.Score)
    } else {
        f.logger.Debug("dual mode scores match",
            "request_id", req.RequestID,
            "score", primaryResult.Score)
    }
}
```

#### 1.3 Phase 1 Deliverables

- [ ] `EnclaveFactory` with dual-mode support
- [ ] Updated configuration schema
- [ ] Migration tooling CLI commands
- [ ] Dual-mode logging and metrics
- [ ] Unit tests for factory

### Phase 2: Testnet Validation (Week 3-5)

**Goal**: Validate TEE integration on testnet with real hardware

#### 2.1 Testnet Requirements

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     TESTNET TEE REQUIREMENTS                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Minimum Validator Setup:                                                   │
│  ├─ 5 validators with AMD SEV-SNP                                          │
│  ├─ 3 validators with Intel SGX                                            │
│  └─ 2 validators with SimulatedEnclaveService (control group)              │
│                                                                             │
│  Testing Scenarios:                                                         │
│  ├─ Normal operation: All validators score identical inputs                 │
│  ├─ Measurement mismatch: One validator with wrong enclave                  │
│  ├─ Attestation failure: Simulate quote verification failure                │
│  ├─ Key rotation: Test key rotation across all validators                   │
│  ├─ Enclave upgrade: Hot-swap enclave measurement                           │
│  └─ Recovery: Test enclave restart and state recovery                       │
│                                                                             │
│  Success Metrics:                                                           │
│  ├─ Score consistency: 100% match across TEE types                          │
│  ├─ Attestation success: >99.9% valid attestations                          │
│  ├─ Latency impact: <50ms additional latency                                │
│  └─ Availability: >99.9% enclave availability                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 2.2 Testnet Configuration

```yaml
# testnet-config.yaml

network:
  chain_id: "virtengine-testnet-tee-1"
  min_validators: 10

enclave_requirements:
  # During testnet, allow mixed modes
  allowed_modes:
    - simulated
    - sgx
    - sev-snp
    - nitro

  # Testnet measurements (not production!)
  allowed_measurements:
    sgx:
      - "abc123..."  # v0.1.0-testnet
      - "def456..."  # v0.1.1-testnet
    sev_snp:
      - "789abc..."  # v0.1.0-testnet
    nitro:
      - "xyz789..."  # v0.1.0-testnet

  # Permissive mode for testing
  attestation_mode: "permissive"
  require_attestation: false  # Set true in phase 3

governance:
  # Proposal to require TEE
  tee_requirement_proposal:
    voting_period: "72h"  # 3 days for testnet
    threshold: "0.5"      # 50% for testnet
```

#### 2.3 Phase 2 Deliverables

- [ ] Testnet deployment with mixed TEE types
- [ ] Automated test suite for TEE scenarios
- [ ] Performance benchmarks (latency, throughput)
- [ ] Monitoring dashboards for TEE metrics
- [ ] Incident response runbook

### Phase 3: Mainnet Preparation (Week 6-8)

**Goal**: Prepare mainnet validators for TEE requirement

#### 3.1 Validator Migration Steps

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    VALIDATOR TEE MIGRATION WORKFLOW                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Step 1: Hardware Verification                                              │
│  ├─ Check CPU support (Intel SGX / AMD SEV-SNP)                            │
│  ├─ Verify firmware/BIOS settings                                          │
│  ├─ Enable TEE features (sgx_enable, sev_snp)                              │
│  └─ Run hardware compatibility check: `virtengine tee verify-hardware`     │
│                                                                             │
│  Step 2: Software Installation                                              │
│  ├─ Install TEE drivers (SGX PSW / SEV kernel)                             │
│  ├─ Install enclave runtime dependencies                                    │
│  ├─ Download signed enclave binary                                          │
│  └─ Configure enclave service                                               │
│                                                                             │
│  Step 3: Key Generation                                                     │
│  ├─ Initialize enclave                                                      │
│  ├─ Generate enclave keys (inside TEE)                                     │
│  ├─ Export public keys for on-chain registration                           │
│  └─ Verify keys with: `virtengine tee verify-keys`                         │
│                                                                             │
│  Step 4: Attestation Test                                                   │
│  ├─ Generate test attestation report                                        │
│  ├─ Submit attestation to testnet                                           │
│  ├─ Verify attestation accepted                                             │
│  └─ Check measurement in allowlist                                          │
│                                                                             │
│  Step 5: Configuration Switch                                               │
│  ├─ Update app.toml: type = "sgx" or "sev-snp"                             │
│  ├─ Disable fallback: fallback_type = ""                                    │
│  ├─ Enable strict mode: attestation.mode = "strict"                         │
│  └─ Restart validator node                                                  │
│                                                                             │
│  Step 6: Verification                                                       │
│  ├─ Check enclave status: `virtengine query enclave status`                │
│  ├─ Verify scoring works: Submit test identity                              │
│  ├─ Check attestation: `virtengine query enclave attestation`              │
│  └─ Monitor for 24 hours                                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 3.2 CLI Commands for Migration

```bash
# Hardware verification
virtengine tee verify-hardware

# Expected output:
# ┌──────────────────────────────────────────────────────────────┐
# │                 TEE Hardware Verification                    │
# ├──────────────────────────────────────────────────────────────┤
# │ Platform:    AMD SEV-SNP                                     │
# │ CPU:         AMD EPYC 7443 24-Core Processor                │
# │ Firmware:    1.55.23 (AGESA 1.0.0.c)                        │
# │ SNP Status:  Enabled                                         │
# │ Memory:      64 GB (encrypted)                               │
# │ Verdict:     ✅ COMPATIBLE                                   │
# └──────────────────────────────────────────────────────────────┘

# Initialize enclave
virtengine tee init --platform sev-snp

# Generate and register keys
virtengine tee keygen
virtengine tx enclave register-key [pubkey] --from validator

# Generate attestation
virtengine tee attest --nonce $(openssl rand -hex 32)

# Verify configuration
virtengine tee verify-config
```

#### 3.3 Governance Proposal

```json
{
  "title": "Enable Mandatory TEE for VEID Scoring",
  "description": "This proposal enables mandatory Trusted Execution Environment (TEE) for all VEID identity scoring operations. After passage:\n\n1. All validators must run approved TEE enclaves\n2. Scoring without valid attestation will be rejected\n3. Grace period of 200 blocks for migration\n4. Non-compliant validators will be jailed\n\nSee: https://docs.virtengine.io/tee-migration",
  "type": "SoftwareUpgrade",
  "changes": [
    {
      "subspace": "enclave",
      "key": "RequireTEE",
      "value": "true"
    },
    {
      "subspace": "enclave",
      "key": "AttestationMode",
      "value": "strict"
    },
    {
      "subspace": "enclave",
      "key": "GracePeriodBlocks",
      "value": "200"
    }
  ],
  "deposit": "10000000uve"
}
```

#### 3.4 Phase 3 Deliverables

- [ ] Validator migration documentation
- [ ] Hardware compatibility guide
- [ ] CLI tools for migration
- [ ] Governance proposal templates
- [ ] Support escalation procedures

### Phase 4: Mainnet Activation (Week 9-10)

**Goal**: Activate mandatory TEE on mainnet

#### 4.1 Activation Timeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MAINNET TEE ACTIVATION TIMELINE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Day 0: Governance Proposal                                                 │
│  └─ Submit proposal for mandatory TEE                                       │
│                                                                             │
│  Day 1-7: Voting Period                                                     │
│  ├─ Community discussion                                                    │
│  └─ Validator commitments                                                   │
│                                                                             │
│  Day 8: Proposal Passes                                                     │
│  └─ Grace period begins (200 blocks ≈ 20 minutes)                          │
│                                                                             │
│  Day 8 + 200 blocks: Enforcement Begins                                     │
│  ├─ Scoring requires valid attestation                                      │
│  ├─ Non-compliant validators: scoring rejected                              │
│  └─ Warnings logged for edge cases                                          │
│                                                                             │
│  Day 8 + 1000 blocks: Full Enforcement                                      │
│  ├─ Slashing enabled for invalid attestation                                │
│  ├─ SimulatedEnclaveService disabled                                        │
│  └─ All scoring must have TEE attestation                                   │
│                                                                             │
│  Day 15: Post-Activation Review                                             │
│  ├─ Analyze activation metrics                                              │
│  ├─ Document issues encountered                                             │
│  └─ Plan for future enclave upgrades                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

#### 4.2 Emergency Procedures

```yaml
# Emergency disable TEE requirement (governance fast-track)

emergency_proposal:
  title: "Emergency: Disable Mandatory TEE"
  reason: "Critical vulnerability discovered in TEE implementation"
  voting_period: "24h"  # Emergency shortened period
  threshold: "0.67"     # 2/3 supermajority required

  changes:
    - key: "RequireTEE"
      value: "false"
    - key: "AttestationMode"
      value: "permissive"

  # Automatic safeguards
  safeguards:
    - "Halt new VEID submissions during transition"
    - "Preserve existing scores (no re-scoring)"
    - "Require post-incident review within 7 days"
```

#### 4.3 Phase 4 Deliverables

- [ ] Mainnet governance proposal
- [ ] Emergency response procedures
- [ ] Post-activation monitoring plan
- [ ] Communication templates

### Phase 5: Deprecation of SimulatedEnclaveService (Week 11-12)

**Goal**: Remove simulated enclave from codebase

#### 5.1 Code Removal Plan

```go
// pkg/enclave_runtime/enclave_service.go

// BEFORE: SimulatedEnclaveService included
type SimulatedEnclaveService struct { ... }

// AFTER: Removed, replaced with:

// DeprecatedSimulatedStub returns an error indicating simulation is no longer supported
func NewSimulatedEnclaveService() (EnclaveService, error) {
    return nil, errors.New(
        "SimulatedEnclaveService has been deprecated and removed. " +
        "Production TEE (SGX, SEV-SNP, or Nitro) is now required. " +
        "See https://docs.virtengine.io/tee-migration for migration guide.",
    )
}
```

#### 5.2 Phase 5 Deliverables

- [ ] Remove SimulatedEnclaveService
- [ ] Update all tests to use TEE stubs
- [ ] Archive simulation code (for reference)
- [ ] Update documentation
- [ ] Final security audit

---

## Validator Migration Checklist

### Pre-Migration

- [ ] Review hardware requirements document
- [ ] Verify CPU TEE support (SGX or SEV-SNP)
- [ ] Check firmware/BIOS settings
- [ ] Allocate budget for hardware upgrades (if needed)
- [ ] Schedule maintenance window

### Hardware Setup

- [ ] Enable TEE in BIOS/UEFI
- [ ] Update CPU microcode
- [ ] Verify kernel support (Linux 6.0+ for SEV-SNP)
- [ ] Install TEE drivers and SDK
- [ ] Run hardware verification: `virtengine tee verify-hardware`

### Software Setup

- [ ] Upgrade virtengine binary
- [ ] Download signed enclave binary
- [ ] Configure enclave in app.toml
- [ ] Initialize enclave: `virtengine tee init`
- [ ] Generate keys: `virtengine tee keygen`

### Registration

- [ ] Register enclave public key on-chain
- [ ] Verify measurement in allowlist
- [ ] Generate test attestation
- [ ] Submit attestation to testnet

### Activation

- [ ] Switch enclave type in config
- [ ] Restart validator
- [ ] Verify enclave status
- [ ] Monitor for 24 hours
- [ ] Confirm scoring success

### Post-Migration

- [ ] Disable fallback mode
- [ ] Enable strict attestation
- [ ] Document lessons learned
- [ ] Update runbooks

---

## Rollback Procedures

### Phase 1-2 Rollback (Testnet)

```bash
# Revert to simulated enclave
virtengine config set enclave.type "simulated"
virtengine config set enclave.fallback_type ""
systemctl restart virtengine
```

### Phase 3-4 Rollback (Mainnet - Requires Governance)

```bash
# Submit emergency proposal
virtengine tx gov submit-proposal \
  --title "Emergency: Revert TEE Requirement" \
  --type emergency-config \
  --changes '[{"key":"RequireTEE","value":"false"}]' \
  --from validator \
  --deposit 100000000uve
```

### Validator-Level Rollback

```bash
# If individual validator needs to revert (will be jailed if TEE required)
virtengine config set enclave.type "simulated"
systemctl restart virtengine

# NOTE: Validator will be jailed until TEE restored or governance reverts requirement
```

---

## Success Criteria

### Phase Completion Gates

| Phase | Criteria | Threshold |
|-------|----------|-----------|
| Phase 1 | Dual mode operational | 100% feature complete |
| Phase 2 | Testnet stability | >99.9% uptime, 0 critical bugs |
| Phase 3 | Validator readiness | >80% validators TEE-ready |
| Phase 4 | Mainnet activation | Governance passes, activation smooth |
| Phase 5 | Deprecation complete | Zero simulated enclave usage |

### Key Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Score Consistency | 100% | Cross-platform score matching |
| Attestation Success | >99.9% | Valid attestations / total |
| Latency Impact | <50ms | Additional scoring latency |
| Availability | >99.9% | Enclave uptime |
| Migration Time | <4 hours | Per-validator migration |

---

## Timeline

| Week | Phase | Key Activities |
|------|-------|----------------|
| 1-2 | Phase 1 | Dual mode infrastructure |
| 3-5 | Phase 2 | Testnet validation |
| 6-8 | Phase 3 | Mainnet preparation |
| 9-10 | Phase 4 | Mainnet activation |
| 11-12 | Phase 5 | Deprecation |

**Total Duration: 12 weeks**

---

## Appendix A: Reference Commands

```bash
# Check TEE status
virtengine query enclave status

# Get enclave measurement
virtengine query enclave measurement

# Generate attestation
virtengine tee attest --nonce $(openssl rand -hex 32)

# Verify attestation
virtengine tee verify-attestation [attestation_file]

# List allowed measurements
virtengine query enclave allowed-measurements

# Propose new measurement
virtengine tx gov submit-proposal add-measurement \
  --platform sgx \
  --measurement "abc123..." \
  --version "v1.0.0" \
  --from validator
```

## Appendix B: Related Documents

- [TEE Security Model](./tee-security-model.md)
- [TEE Integration Architecture](./tee-integration-architecture.md)
- [TEE Integration Plan](./tee-integration-plan.md)
- [Validator Hardware Guide](./validator-hardware-guide.md)
- [Slashing Conditions](./slashing-conditions.md)
