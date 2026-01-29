# VirtEngine Production Readiness Tasks

## Analysis Summary

Based on comprehensive analysis of:

- `_docs/veid-flow-spec.md` - VEID three-layer model specification
- `_docs/ralph/prd.json` - User stories VE-200 through VE-804
- `x/veid/`, `x/mfa/`, `x/roles/`, `x/market/` implementations
- `pkg/enclave_runtime/`, `pkg/govdata/`, `pkg/edugain/`, `pkg/payment/` implementations

### Critical Blockers Identified

| Blocker ID | Description                                                                  | Current Status                                                                  |
| ---------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| BLOCK-004  | Only SimulatedEnclaveService exists - no real TEE integration                | pkg/enclave_runtime has SGX/SEV/Nitro configs but only simulated implementation |
| BLOCK-006  | SAML signature verification uses goxmldsig but may need production hardening | Verification code exists but needs security audit                               |
| BLOCK-007  | Government verification returns mock data                                    | pkg/govdata has adapters but needs real API integration                         |

### Module Completion Matrix (Updated)

| Module              | Completion | Key Gaps                                          |
| ------------------- | ---------- | ------------------------------------------------- |
| x/veid              | 25%        | Proto stubs, tier transition logic incomplete     |
| x/mfa               | 40%        | Proto stubs, factor verification flows incomplete |
| x/roles             | 35%        | Proto stubs, role assignment governance missing   |
| x/market            | 85%        | VEID gating integration needed                    |
| x/escrow            | 85%        | Settlement automation needed                      |
| pkg/enclave_runtime | 20%        | Only SimulatedEnclaveService, no real TEE         |
| pkg/govdata         | 25%        | Mock adapters only                                |
| pkg/edugain         | 30%        | SAML verification exists but needs hardening      |
| pkg/payment         | 35%        | Stripe adapter exists but incomplete              |

---

## SPECIFICATION-DRIVEN IMPLEMENTATION TASKS

### Category 1: VEID Core Identity System (VE-200 series gaps)

#### VEID-CORE-001: Implement VEID Tier Transition Logic

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Account Tier System  
**Current State:** Types defined but transition logic not implemented

**Implementation Path:**

1. File: `x/veid/keeper/tier_transitions.go` (new)
2. Implement `UpdateAccountTier(ctx, addr)` that:
   - Fetches all verified scopes for account
   - Calculates composite score from scope scores
   - Maps score to tier: 0-49→Tier0, 50-69→Tier1, 70-84→Tier2, 85-100→Tier3
   - Stores new tier and emits `VEIDTierChanged` event
3. Add `GetAccountTier(ctx, addr) AccountTier` query
4. Add `MeetsScoreThreshold(ctx, addr, threshold) bool` helper

**Acceptance Criteria (from spec):**

- Account tiers: Tier 0 (Unverified), Tier 1 (Basic 50-69), Tier 2 (Standard 70-84), Tier 3 (Premium 85-100)
- Tier changes emit events with old/new tier
- Query endpoints return current tier and score

---

#### VEID-CORE-002: Implement Identity Scope Scoring Algorithm

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - ML Score Calculation  
**Current State:** Scoring placeholders only

**Implementation Path:**

1. File: `x/veid/keeper/scoring.go` (new)
2. Implement composite scoring with weights:
   - Document Authenticity: 25%
   - Face Match: 25%
   - Liveness Detection: 20%
   - Data Consistency: 15%
   - Historical Signals: 10%
   - Risk Indicators: 5%
3. Store score version with each computed score
4. Determinism: All computations must be reproducible

**Acceptance Criteria:**

- Score is 0-100 with reason codes
- Score version recorded in verification state
- Unit tests cover scoring boundaries and invariants

---

#### VEID-CORE-003: Implement Identity Wallet On-Chain Primitive

**Priority:** HIGH  
**Spec Reference:** VE-209 - Identity Wallet primitive  
**Current State:** Not implemented

**Implementation Path:**

1. File: `x/veid/types/identity_wallet.go` (new)
2. Define `IdentityWallet` struct with:
   - Encrypted scope envelope references
   - Derived feature hashes
   - Verification history
   - Current score/status
3. Implement in keeper: `CreateIdentityWallet`, `UpdateIdentityWallet`, `RevokeScope`
4. Add consent toggle per scope

**Acceptance Criteria:**

- Wallet cryptographically bound to user account
- Scope-level consent toggles and revocation flags
- Query returns only non-sensitive metadata

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
**Current State:** Keeper interface defined, verification logic incomplete

**Implementation Path:**

1. File: `x/mfa/keeper/verification.go`
2. Complete `VerifyMFAChallenge(ctx, challengeID, response)`:
   - FIDO2/WebAuthn: Validate using go-webauthn library
   - TOTP: Validate using pquerna/otp library
   - SMS/Email OTP: Validate against stored challenge
3. Implement factor combination rules from spec

**Acceptance Criteria:**

- Support factor types: FIDO2, TOTP, SMS, Email, VEID biometric
- Verification within TTL, max attempts enforced
- Audit events emitted for all verification attempts

---

#### MFA-CORE-002: Implement Authorization Session Management

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Authorization Sessions  
**Current State:** Session types defined, lifecycle incomplete

**Implementation Path:**

1. File: `x/mfa/keeper/sessions.go`
2. Implement session durations per action type:
   - Critical (AccountRecovery, KeyRotation): Single use
   - High (ProviderReg, LargeWithdrawal): 15 minutes
   - Medium (HighValueOrder): 30 minutes
3. Implement `HasValidAuthSession(ctx, addr, action) bool`
4. Implement `ConsumeAuthSession(ctx, addr, action) error` for single-use

**Acceptance Criteria:**

- Session expires after configured duration
- Single-use sessions consumed after first use
- Sessions bound to device/context hash

---

#### MFA-CORE-003: Implement Sensitive Transaction Gating

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Sensitive Actions List  
**Current State:** SensitiveTxConfig exists, gating not enforced

**Implementation Path:**

1. File: `app/ante_mfa.go` - extend existing decorator
2. Implement full action-to-requirement mapping per spec:
   - AccountRecovery: VEID + FIDO2 + SMS/Email, Single use
   - ProviderRegistration: VEID + FIDO2, 15 min
   - LargeWithdrawal (>10k VE): VEID + FIDO2, 15 min
   - HighValueOrder (>1k VE): VEID + FIDO2, 30 min
3. Return explicit MFA requirement errors

**Acceptance Criteria:**

- All designated sensitive transactions gated per spec
- Clear error messages with required factors
- Audit trail for all gated actions

---

### Category 3: TEE Enclave Integration (CRITICAL BLOCKER)

#### TEE-IMPL-001: Implement SGX Enclave Service

**Priority:** CRITICAL (BLOCKER)  
**Spec Reference:** VE-231 - Enclave Runtime v1  
**Current State:** Only SimulatedEnclaveService exists

**Implementation Path:**

1. File: `pkg/enclave_runtime/sgx_enclave.go` (new)
2. Implement `SGXEnclaveService` using ego-go or gramine-direct SDK
3. Key generation inside enclave, sealed storage
4. Score computation with no plaintext escape
5. Enclave signature on results

**Acceptance Criteria:**

- Enclave generates keys inside and seals them
- Host cannot export private keys
- Plaintext scrubbed after processing
- Integration test verifies no plaintext outside enclave

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
**Current State:** Single recipient encryption only

**Implementation Path:**

1. File: `x/encryption/types/envelope.go` - extend
2. Add per-recipient `wrapped_key` entries
3. Support full validator set or committee subset
4. Deterministic serialization

**Acceptance Criteria:**

- Envelope includes multiple wrapped keys
- Canonical encoding, deterministic bytes
- Unit tests: same inputs produce identical envelope bytes

---

### Category 4: Marketplace VEID Integration

#### MARKET-VEID-001: Implement Order VEID Score Gating

**Priority:** HIGH  
**Spec Reference:** VE-301 - Marketplace gating  
**Current State:** Not implemented

**Implementation Path:**

1. File: `x/market/keeper/veid_gating.go` (new)
2. In CreateOrder: check customer tier >= offering.min_customer_tier
3. In CreateOrder: check customer score >= offering.min_customer_score
4. Return clear rejection reason if insufficient

**Acceptance Criteria:**

- Offerings can declare minimum identity score/status requirement
- Order creation rejected if insufficient identity
- UI/API surfaces reason and required steps

---

#### MARKET-VEID-002: Implement Provider VEID Registration Check

**Priority:** HIGH  
**Spec Reference:** veid-flow-spec.md - Provider Registration requires score ≥70  
**Current State:** Not enforced

**Implementation Path:**

1. File: `x/provider/keeper/registration.go`
2. In RegisterProvider: require VEID score ≥ 70
3. Require MFA session valid for ProviderRegistration
4. Domain verification integration

**Acceptance Criteria:**

- Provider registration blocked if score < 70
- MFA required for provider registration
- Domain verification linked to provider profile

---

#### MARKET-VEID-003: Implement Validator VEID Registration Check

**Priority:** CRITICAL  
**Spec Reference:** veid-flow-spec.md - Validator Registration requires score ≥85  
**Current State:** Not enforced

**Implementation Path:**

1. File: `x/staking/keeper/` - extend CreateValidator
2. Require VEID score ≥ 85
3. Require MFA (FIDO2) + Governance approval

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
