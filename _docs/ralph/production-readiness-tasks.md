# VirtEngine Production Readiness Tasks

## Analysis Summary

Based on comprehensive analysis of:

- `_docs/veid-flow-spec.md` - VEID three-layer model specification
- `_docs/ralph/prd.json` - User stories VE-200 through VE-804
- `x/veid/`, `x/mfa/`, `x/roles/`, `x/market/` implementations
- `pkg/enclave_runtime/`, `pkg/govdata/`, `pkg/edugain/`, `pkg/payment/` implementations

### Critical Blockers Identified

| Blocker ID | Description                                 | Current Status                                                                                |
| ---------- | ------------------------------------------- | --------------------------------------------------------------------------------------------- |
| BLOCK-004  | 🟡 TEE hardware deployment                  | SGX/SEV-SNP/Nitro adapters implemented; production hardware rollout pending                   |
| BLOCK-009  | 🟡 Payment conversion & disputes            | Price feed integration, conversion execution, and dispute lifecycle persistence pending      |
| BLOCK-010  | 🟡 Artifact store backend                   | Waldur artifact backend still stubbed (no real storage API integration)                       |
| BLOCK-011  | 🟡 NLI session storage                      | In-memory sessions + rate limits; requires Redis/distributed backing                          |
| BLOCK-012  | 🟡 Provider daemon event streaming          | Order/config polling still used; needs WebSocket/gRPC streaming for scale                     |

### Module Completion Matrix (Updated 2026-01-30)

| Module              | Completion | Status                                                                |
| ------------------- | ---------- | --------------------------------------------------------------------- |
| x/veid              | 90%        | ✅ Tier transitions, scoring, wallet, salt-binding complete           |
| x/mfa               | 90%        | ✅ Verification flows, session management, ante handler complete      |
| x/roles             | 75%        | ✅ MsgServer/QueryServer enabled; proto generation complete           |
| x/market            | 90%        | ✅ VEID gating integration complete                                   |
| x/escrow            | 85%        | ✅ Settlement automation complete                                     |
| pkg/enclave_runtime | 85%        | ✅ SGX/SEV-SNP/Nitro adapters + Manager (hardware deployment pending) |
| pkg/govdata         | 95%        | ✅ AAMVA DLDV + DVS + eIDAS + GOV.UK adapters                         |
| pkg/edugain         | 90%        | ✅ goxmldsig XML-DSig verification complete                           |
| pkg/payment         | 90%        | ✅ Stripe/Adyen gateways complete; conversion/dispute flows pending   |

---

## SPECIFICATION-DRIVEN IMPLEMENTATION TASKS

### Category 1: VEID Core Identity System (VE-200 series gaps)

#### VEID-CORE-001: Implement VEID Tier Transition Logic

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Account Tier System  
**Current State:** ✅ **COMPLETED** - x/veid/keeper/tier_transitions.go implemented

**Evidence:**

- `TierTransitionResult` struct with `UpdateAccountTier()` method
- Tier mapping: 0-49→Tier0, 50-69→Tier1, 70-84→Tier2, 85-100→Tier3
- Events emitted on tier changes

---

#### VEID-CORE-002: Implement Identity Scope Scoring Algorithm

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - ML Score Calculation  
**Current State:** ✅ **COMPLETED** - x/veid/keeper/scoring.go implemented (765 lines)

**Evidence:**

- `MLScoringConfig` with full TensorFlow integration
- `ScoringInput/Output` structs with deterministic hash computation
- Composite scoring with configurable weights
- Score version tracking

---

#### VEID-CORE-003: Implement Identity Wallet On-Chain Primitive

**Priority:** HIGH  
**Spec Reference:** VE-209 - Identity Wallet primitive  
**Current State:** ✅ **COMPLETED** - x/veid/types/identity_wallet.go implemented (755 lines)

**Evidence:**

- `IdentityWallet` struct with scope refs, derived features, consent settings
- `MsgCreateIdentityWallet` in x/veid/types/msg.go
- Keeper methods: `CreateIdentityWallet`, `UpdateIdentityWallet`, `RevokeScope`
- Full wallet store implementation

---

#### VEID-CORE-004: Implement Capture Protocol Salt-Binding

**Priority:** HIGH  
**Spec Reference:** VE-207 - Mobile capture protocol  
**Current State:** ✅ **COMPLETED** - salt binding + signature validation implemented

**Evidence:**

- `x/veid/keeper/capture_validation.go` implements `ValidateSaltBinding` + signature checks
- `x/veid/keeper/keeper.go` enforces salt binding and client/user signatures on uploads
- `x/veid/keeper/signature_crypto_test.go` covers salt binding validation paths

---

### Category 2: MFA Module Completion

#### MFA-CORE-001: Implement MFA Challenge Verification Flows

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - MFA Authorization Integration  
**Current State:** ✅ **COMPLETED** - x/mfa/keeper/verification.go implemented (1416 lines)

**Evidence:**

- TOTP verification via `verifyTOTPCode()` (RFC 6238)
- FIDO2/WebAuthn verification in x/mfa/keeper/fido2_verifier.go
- Factor combination rules with TTL and max attempts
- Audit events for all verification attempts

---

#### MFA-CORE-002: Implement Authorization Session Management

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Authorization Sessions  
**Current State:** ✅ **COMPLETED** - x/mfa/keeper/sessions.go implemented (282 lines)

**Evidence:**

- `HasValidAuthSession()` with device fingerprint validation
- `ConsumeAuthSession()` for single-use sessions
- Session durations: Critical=single-use, High=15m, Medium=30m, Low=60m
- Sessions bound to device/context hash

---

#### MFA-CORE-003: Implement Sensitive Transaction Gating

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Sensitive Actions List  
**Current State:** ✅ **COMPLETED** - app/ante_mfa.go implemented (283 lines)

**Evidence:**

- `MFAGatingDecorator` integrated into ante handler chain
- Action-to-requirement mapping for AccountRecovery, KeyRotation, ProviderReg, LargeWithdrawal, ValidatorReg
- Clear error messages with required factors
- Audit trail for all gated actions

---

### Category 3: TEE Enclave Integration (Implementation Complete)

#### TEE-IMPL-001: Implement SGX Enclave Service

**Priority:** CRITICAL (BLOCKER)  
**Spec Reference:** VE-231 - Enclave Runtime v1  
**Current State:** ✅ **COMPLETED** - pkg/enclave_runtime/sgx_adapter.go implemented (1223 lines)

**Evidence:**

- `SGXEnclaveService` with factory support
- Attestation verification in enclave_attestation.go
- EnclaveManager for lifecycle management
- POC implementation (hardware deployment pending)

---

#### TEE-IMPL-002: Implement SEV-SNP Enclave Service

**Priority:** HIGH  
**Spec Reference:** VE-228 - TEE Security Model  
**Current State:** ✅ **COMPLETED** - SEV-SNP enclave service implemented

**Evidence:**

- `pkg/enclave_runtime/sev_enclave.go` production SEV-SNP implementation
- `pkg/enclave_runtime/hardware_sev.go` + `crypto_sev.go` for attestation and crypto paths
- `pkg/enclave_runtime/sev_enclave_test.go` for validation

---

#### TEE-IMPL-003: Implement Enclave Registry On-Chain Module

**Priority:** CRITICAL  
**Spec Reference:** VE-229 - On-chain Enclave Registry  
**Current State:** ✅ **COMPLETED** - on-chain enclave registry implemented

**Evidence:**

- `x/enclave/` module with keepers, msgs, and genesis
- Measurement allowlist + governance proposal handlers in `x/enclave/keeper` and `x/enclave/client`
- Protobufs generated under `sdk/proto/node/virtengine/enclave`

---

#### TEE-IMPL-004: Multi-Recipient Encryption for Validator Enclaves

**Priority:** CRITICAL  
**Spec Reference:** VE-230 - Encrypted envelope upgrade  
**Current State:** ✅ **COMPLETED** - x/encryption/types/envelope.go extended

**Evidence:**

- Per-recipient wrapped_key entries implemented
- Full validator set and committee subset support
- Deterministic serialization with canonical encoding
- Unit tests verify identical envelope bytes for same inputs

---

### Category 4: Marketplace VEID Integration

#### MARKET-VEID-001: Implement Order VEID Score Gating

**Priority:** HIGH  
**Spec Reference:** VE-301 - Marketplace gating  
**Current State:** ✅ **COMPLETED** - x/market/keeper/veid_gating.go implemented

**Evidence:**

- CreateOrder checks customer tier >= offering.min_customer_tier
- CreateOrder checks customer score >= offering.min_customer_score
- Clear rejection reason returned if insufficient identity

---

#### MARKET-VEID-002: Implement Provider VEID Registration Check

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Provider Registration requires score ≥70  
**Current State:** ✅ **COMPLETED** - x/provider/keeper/registration.go updated

**Evidence:**

- RegisterProvider requires VEID score ≥ 70
- MFA session validation for ProviderRegistration
- Domain verification integration via provider_domain_verification.go

---

#### MARKET-VEID-003: Implement Validator VEID Registration Check

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Validator Registration requires score ≥85  
**Current State:** ✅ **COMPLETED** - x/staking/keeper/ extended

**Evidence:**

- CreateValidator requires VEID score ≥ 85
- MFA (FIDO2) requirement enforced
- Governance approval workflow integrated

**Acceptance Criteria:**

- Validator registration blocked if score < 85
- MFA + governance approval required
- Audit events for validator identity verification

---

### Category 5: Proto Generation (RESOLVED)

#### PROTO-GEN-001: Complete VEID Proto Generation

**Priority:** CRITICAL  
**Spec Reference:** VEID protobuf definitions  
**Current State:** ✅ **COMPLETED** - VEID protos generated; stubs removed

**Evidence:**

- Protos under `sdk/proto/node/virtengine/veid/`
- `x/veid/types/codec.go` uses generated registrations

---

#### PROTO-GEN-002: Complete MFA Proto Generation

**Priority:** CRITICAL  
**Current State:** ✅ **COMPLETED** - MFA protos generated; stubs removed

**Evidence:**

- Protos under `sdk/proto/node/virtengine/mfa/`

---

#### PROTO-GEN-003: Complete Staking Extension Proto

**Priority:** HIGH  
**Current State:** ✅ **COMPLETED** - Staking extension protos generated

**Evidence:**

- Protos under `sdk/proto/node/virtengine/staking/`

---

#### PROTO-GEN-004: Complete HPC Proto Generation

**Priority:** HIGH  
**Current State:** ✅ **COMPLETED** - HPC protos generated

**Evidence:**

- Protos under `sdk/proto/node/virtengine/hpc/`

---

### Category 6: Government Data Integration (RESOLVED)

#### GOVDATA-001: Implement AAMVA Production Adapter

**Priority:** CRITICAL (BLOCKER)  
**Spec Reference:** VE-909 - Government data integration  
**Current State:** ✅ **COMPLETED** - AAMVA production adapter implemented

**Evidence:**

- `pkg/govdata/aamva_adapter.go` with DLDV integration
- Rate limiting and audit logging in govdata service

---

#### GOVDATA-002: Add Additional Jurisdiction Adapters

**Priority:** HIGH
**Current State:** ✅ **COMPLETED** - DVS, GOV.UK, eIDAS adapters implemented

**Evidence:**

- `pkg/govdata/` includes DVS, GOV.UK, and eIDAS adapters

---

### Category 7: ML Pipeline Determinism

#### ML-DET-001: Pin TensorFlow-Go Determinism

**Priority:** CRITICAL  
**Spec Reference:** VE-219 - Deterministic identity verification runtime  
**Current State:** pkg/inference has determinism config but needs validation

**Implementation Path:**

1. Validate all TF ops are deterministic
2. Pin exact model weight hashes
3. Create conformance test suite
4. CPU-only mode enforced in production

**Acceptance Criteria:**

- Pinned OCI image with model hashes
- Conformance test verifies same inputs → same outputs across machines
- Integration test: proposer → validator recomputation matches exactly

---

#### ML-DET-002: DeepFace Pipeline Determinism

**Priority:** HIGH  
**Spec Reference:** VE-211 - Facial verification pipeline  
**Current State:** Python ML in ml/facial_verification

**Implementation Path:**

1. Pin exact library versions in requirements-deterministic.txt
2. Document preprocessing steps explicitly
3. Deterministic face alignment
4. Record model hash with each verification

---

### Category 8: Testing Infrastructure

#### TEST-001: E2E VEID Onboarding Flow

**Priority:** HIGH

Test path: Create account → Upload scope → Validator score → Tier change → Order placement

---

#### TEST-002: E2E MFA Gating Flow

**Priority:** HIGH

Test path: Attempt sensitive action → MFA challenge → Complete factors → Action succeeds

---

#### TEST-003: E2E Provider Daemon Flow

**Priority:** HIGH

Test path: Provider register → Order created → Bid placed → Allocation → Provision → Usage report

---

### Category 9: Ante Handler Completion

#### ANTE-001: Complete VEID Decorator

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Ante Handler Integration  
**Current State:** ✅ **COMPLETED** - VEID ante decorator implemented

**Evidence:**

- `app/ante_veid.go` enforces tier/score + MFA authorization requirements
- Governance bypass and clear rejection messages implemented

---

### Category 10: Event System

#### EVENTS-001: Implement Complete VEID Events

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Events section
**Current State:** ✅ **COMPLETED** - VEID events defined and emitted

**Evidence:**

- `x/veid/types/events.go` defines typed events
- `x/veid/keeper/events.go` and msg handlers emit typed events

---

#### EVENTS-002: Implement Marketplace Events for Provider Daemon

**Priority:** HIGH  
**Spec Reference:** VE-304
**Current State:** ✅ **COMPLETED** - marketplace events implemented

**Evidence:**

- `x/market/types/marketplace/events.go` defines events
- `x/market/types/marketplace/keeper/keeper.go` emits events on state changes

---

### Category 11: Production Hardening & Integrations (NEW)

#### TEE-HW-001: Deploy TEE Hardware & Attestation in Production

**Priority:** CRITICAL  
**Spec Reference:** VE-228/231 - TEE production deployment  
**Current State:** Adapters implemented; production hardware rollout pending

**Acceptance Criteria:**

- SGX/SEV-SNP/Nitro hardware provisioned and attestation verified
- Production runbooks and monitoring for enclave health
- Failover strategy across TEE types documented

---

#### PAY-001: Implement Real Price Feed Integration

**Priority:** CRITICAL  
**Spec Reference:** VE-906 - Fiat-to-crypto onramp  
**Current State:** Price feed placeholders in `pkg/payment/service.go`

**Acceptance Criteria:**

- Real price feed integration (CoinGecko/Chainlink/Pyth) with caching and retries
- Deterministic conversion quotes with source attribution
- Failure fallback documented and monitored

---

#### PAY-002: Implement Conversion Execution & Treasury Transfer

**Priority:** CRITICAL  
**Spec Reference:** VE-906 - Conversion execution  
**Current State:** `ExecuteConversion` is a stub (no on-chain transfer)

**Acceptance Criteria:**

- Treasury transfer executed on-chain (MsgSend or dedicated module)
- Idempotent conversion execution with ledger persistence
- Clear failure handling and audit trail

---

#### PAY-003: Dispute Lifecycle Persistence & Gateway Actions

**Priority:** HIGH  
**Spec Reference:** Payment disputes (Stripe/Adyen)  
**Current State:** Dispute retrieval/submit/accept are stubbed in `pkg/payment/service.go`

**Acceptance Criteria:**

- Dispute records persisted (DB or module store)
- Stripe/Adyen dispute actions wired via gateway APIs
- Webhook ingestion updates dispute state

---

#### ARTIFACT-001: Replace Waldur Artifact Store Stub

**Priority:** HIGH  
**Spec Reference:** VE-304 + artifact storage  
**Current State:** `pkg/artifact_store/waldur_backend.go` uses stubbed in-memory storage

**Acceptance Criteria:**

- Real Waldur storage API integration (upload/download/pin)
- Production auth and quota enforcement
- Streaming support for large artifacts

---

#### NLI-001: Redis-Backed NLI Sessions & Rate Limiting

**Priority:** HIGH  
**Spec Reference:** VE-910 - NLI service scale  
**Current State:** `pkg/nli/service.go` uses in-memory session and rate limit maps

**Acceptance Criteria:**

- Redis-backed session store with TTL and eviction policy
- Distributed rate limiting (reuse `pkg/ratelimit`)
- Metrics for session count and rate-limit hits

---

#### PROVIDER-STREAM-001: Replace Provider Daemon Polling with Streaming

**Priority:** HIGH  
**Spec Reference:** VE-304 - Provider daemon scale  
**Current State:** `pkg/provider_daemon/bid_engine.go` polls orders/config on tickers

**Acceptance Criteria:**

- WebSocket/gRPC event subscriptions to chain events
- Checkpointed replay on reconnect using existing checkpoint store
- Polling fallback only for recovery

---

#### MOAB-SSH-001: Enforce SSH Known Hosts Verification

**Priority:** HIGH  
**Spec Reference:** HPC adapter security  
**Current State:** `pkg/moab_adapter/client.go` uses `ssh.InsecureIgnoreHostKey()`

**Acceptance Criteria:**

- Host key verification via known_hosts or pinned keys
- Configurable trust store per provider
- Security tests for MITM protection

---

## TASK IMPORT SUMMARY

Total tasks identified for vibe-kanban import:

| Category                             | Count | Priority      |
| ------------------------------------ | ----- | ------------- |
| Production Hardening & Integrations  | 8     | CRITICAL/HIGH |
| ML Determinism                       | 2     | CRITICAL/HIGH |
| Testing (E2E)                        | 3     | HIGH          |

**Total: 13 detailed implementation tasks**

---

## NEXT STEPS

1. When vibe-kanban becomes available, import these tasks
2. Prioritize CRITICAL tasks (TEE, Proto, VEID Core)
3. Run subagents on independent tasks in parallel
4. Validate with `make lint-go && make test`

---

_Generated: 2026-01-30_
_Based on: veid-flow-spec.md, prd.json VE-200 through VE-804_
