# VirtEngine Production Readiness Tasks

## Analysis Summary

Based on comprehensive analysis of:

- `_docs/veid-flow-spec.md` - VEID three-layer model specification
- `_docs/ralph/prd.json` - User stories VE-200 through VE-804
- `x/veid/`, `x/mfa/`, `x/roles/`, `x/market/` implementations
- `pkg/enclave_runtime/`, `pkg/govdata/`, `pkg/edugain/`, `pkg/payment/` implementations

### Critical Blockers Identified

| Blocker ID | Description                                | Current Status                                                                       |
| ---------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| BLOCK-004  | ✅ TEE Implementation Complete             | SGX, SEV-SNP, Nitro adapters + EnclaveManager + Attestation Verification implemented |
| BLOCK-006  | ✅ SAML signature verification implemented | pkg/edugain/saml_verifier.go with goxmldsig (667 lines)                              |
| BLOCK-007  | ✅ Government verification implemented     | pkg/govdata/aamva_adapter.go with DLDV API (1044 lines)                              |
| BLOCK-008  | 🟡 Proto stubs remain (14 files)           | Hand-written ProtoMessage stubs in 14 modules need buf migration                     |

### Module Completion Matrix (Updated 2026-01-30)

| Module              | Completion | Status                                                                |
| ------------------- | ---------- | --------------------------------------------------------------------- |
| x/veid              | 90%        | ✅ Tier transitions, scoring, wallet, salt-binding complete           |
| x/mfa               | 90%        | ✅ Verification flows, session management, ante handler complete      |
| x/roles             | 70%        | ✅ MsgServer/QueryServer enabled; proto stub remains                  |
| x/market            | 90%        | ✅ VEID gating integration complete                                   |
| x/escrow            | 85%        | ✅ Settlement automation complete                                     |
| pkg/enclave_runtime | 85%        | ✅ SGX/SEV-SNP/Nitro adapters + Manager (hardware deployment pending) |
| pkg/govdata         | 95%        | ✅ AAMVA DLDV + DVS + eIDAS + GOV.UK adapters                         |
| pkg/edugain         | 90%        | ✅ goxmldsig XML-DSig verification complete                           |
| pkg/payment         | 90%        | ✅ Stripe SDK integration complete                                    |

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
**Current State:** Salt validation stub exists

**Implementation Path:**

1. File: `x/veid/keeper/capture_validation.go` (new)
2. Implement `ValidateSaltBinding(ctx, salt, metadata, payloadHash) error`
3. Implement `ValidateClientSignature(ctx, clientID, sig, payload) error`
4. Implement `ValidateUserSignature(ctx, addr, sig, payload) error`
5. Update `UploadScope` to require all three validations

**Acceptance Criteria:**

- Per-upload salt bound into file metadata and signed payload hash
- Server/chain validation rejects missing salt binding or invalid signatures
- Protocol spec includes key rotation strategy for approved clients

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

### Category 3: TEE Enclave Integration (CRITICAL BLOCKER)

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
**Current State:** Config exists but not implemented

**Implementation Path:**

1. File: `pkg/enclave_runtime/sev_enclave.go` (new)
2. Implement using AMD SEV-SNP attestation SDK
3. Same interface as SGXEnclaveService
4. Remote attestation integration

---

#### TEE-IMPL-003: Implement Enclave Registry On-Chain Module

**Priority:** CRITICAL  
**Spec Reference:** VE-229 - On-chain Enclave Registry  
**Current State:** Not implemented

**Implementation Path:**

1. New module: `x/enclave/`
2. Store per-validator enclave identity:
   - enclave_measurement_hash
   - enclave_enc_pubkey
   - enclave_sign_pubkey
   - attestation_quote
3. Governance-controlled measurement allowlist
4. MsgRegisterEnclaveIdentity, MsgRotateEnclaveIdentity

**Acceptance Criteria:**

- Validators can register enclave identities
- Clients can query active validator enclave keys
- Rejected: expired attestations, non-allowlisted measurements

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

### Category 5: Proto Generation (SYSTEMIC BLOCKER)

#### PROTO-GEN-001: Complete VEID Proto Generation

**Priority:** CRITICAL  
**Spec Reference:** All VEID messages use stubs  
**Current State:** Proto stubs in x/veid/types/proto_stub.go

**Implementation Path:**

1. Create `proto/veid/` directory
2. Define all message types:
   - MsgUploadScope, MsgRevokeScope
   - MsgRequestVerification, MsgUpdateVerificationStatus
   - MsgUpdateScore
   - MsgCreateIdentityWallet, MsgAddScopeToWallet
3. Run `buf generate`
4. Remove proto_stub.go

**Acceptance Criteria:**

- All VEID messages properly generated
- Cosmos SDK autocli integration works
- Build and tests pass

---

#### PROTO-GEN-002: Complete MFA Proto Generation

**Priority:** CRITICAL  
**Current State:** Proto stubs in x/mfa/types/

**Implementation Path:**

1. Create `proto/mfa/` directory
2. Define: MsgEnrollFactor, MsgRevokeFactor, MsgInitiateChallenge, etc.
3. Run `buf generate`

---

#### PROTO-GEN-003: Complete Staking Extension Proto

**Priority:** HIGH  
**Current State:** Proto stubs in x/staking/types/codec.go

---

#### PROTO-GEN-004: Complete HPC Proto Generation

**Priority:** HIGH  
**Current State:** Proto stubs in x/hpc/types/codec.go

---

### Category 6: Government Data Integration (BLOCKER)

#### GOVDATA-001: Implement AAMVA Production Adapter

**Priority:** CRITICAL (BLOCKER)  
**Spec Reference:** VE-909 - Government data integration  
**Current State:** Mock adapter only

**Implementation Path:**

1. File: `pkg/govdata/aamva_adapter.go` - complete implementation
2. Integrate with AAMVA DLDV service
3. Handle all response codes and error cases
4. Add comprehensive logging (no PII)

**Acceptance Criteria:**

- Real AAMVA API calls with proper error handling
- Credential management via secrets
- Rate limiting and retry logic
- Audit trail for all verifications

---

#### GOVDATA-002: Add Additional Jurisdiction Adapters

**Priority:** HIGH

Implement adapters for:

- Australia (DVS)
- UK (GOV.UK Verify)
- EU (eIDAS)
- Canada (PCTF)

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
**Current State:** Partially implemented in app/ante.go

**Implementation Path:**

1. File: `app/ante_veid.go` (new)
2. Implement per-message requirement lookup
3. Check tier and score thresholds
4. Integrate with MFA decorator for authorization

**Acceptance Criteria:**

- All messages have defined requirements (or defaults)
- Clear rejection messages
- Bypass for genesis and governance

---

### Category 10: Event System

#### EVENTS-001: Implement Complete VEID Events

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Events section

Events to implement:

- EventTypeVerificationSubmitted
- EventTypeVerificationCompleted
- EventTypeTierChanged
- EventTypeAuthorizationGranted
- EventTypeAuthorizationConsumed
- EventTypeAuthorizationExpired

---

#### EVENTS-002: Implement Marketplace Events for Provider Daemon

**Priority:** HIGH  
**Spec Reference:** VE-304

Events:

- OrderCreated
- BidPlaced
- AllocationCreated
- ProvisionRequested
- TerminateRequested
- UsageUpdateRequested

---

## TASK IMPORT SUMMARY

Total tasks identified for vibe-kanban import:

| Category           | Count | Priority      |
| ------------------ | ----- | ------------- |
| VEID Core          | 4     | CRITICAL/HIGH |
| MFA Core           | 3     | CRITICAL/HIGH |
| TEE Implementation | 4     | CRITICAL      |
| Marketplace VEID   | 3     | HIGH/CRITICAL |
| Proto Generation   | 4     | CRITICAL/HIGH |
| Government Data    | 2     | CRITICAL/HIGH |
| ML Determinism     | 2     | CRITICAL/HIGH |
| Testing            | 3     | HIGH          |
| Ante Handlers      | 1     | CRITICAL      |
| Events             | 2     | HIGH          |

**Total: 28 detailed implementation tasks**

---

## NEXT STEPS

1. When vibe-kanban becomes available, import these tasks
2. Prioritize CRITICAL tasks (TEE, Proto, VEID Core)
3. Run subagents on independent tasks in parallel
4. Validate with `make lint-go && make test`

---

_Generated: $(date)_
_Based on: veid-flow-spec.md, prd.json VE-200 through VE-804_
