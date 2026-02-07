# VEID Testing Infrastructure Guide

**Version:** 1.0.0  
**Date:** 2026-02-02  
**Task Reference:** VEID Testing Infrastructure Setup

---

## Table of Contents

1. [VEID Module Structure](#veid-module-structure)
2. [Identity Score System](#identity-score-system)
3. [Available Messages (Tx Commands)](#available-messages-tx-commands)
4. [Available Queries](#available-queries)
5. [MFA Module Overview](#mfa-module-overview)
6. [Testing Infrastructure](#testing-infrastructure)
7. [Setting Up Test Identities](#setting-up-test-identities)
8. [Admin/Bypass Mechanisms](#adminbypass-mechanisms)
9. [What Needs to Be Implemented](#what-needs-to-be-implemented)

---d

## VEID Module Structure

The VEID module is located at `x/veid/` with the following structure:

```
x/veid/
├── alias.go              # Module aliases
├── genesis.go            # Genesis state initialization
├── genesis_test.go       # Genesis tests
├── module.go             # Module definition (AppModuleBasic, AppModule)
├── keeper/               # Keeper implementation
│   ├── keeper.go         # Main keeper with IKeeper interface
│   ├── msg_server.go     # Message handlers (MsgUploadScope, MsgUpdateScore, etc.)
│   ├── grpc_query.go     # Query server (11 query methods)
│   ├── score.go          # Score management (SetScore, GetScore, UpdateScore)
│   ├── verification.go   # Verification pipeline
│   ├── scoring.go        # ML scoring integration
│   ├── wallet.go         # Identity wallet management
│   └── ...               # Many other keeper files
└── types/
    ├── msgs.go           # Message type definitions
    ├── genesis.go        # Genesis state types
    ├── identity.go       # Identity record types
    ├── score.go          # Score types and tiers
    ├── scope.go          # Scope types (Selfie, IDDocument, FaceVideo)
    └── ...               # Many other type files
```

### Key Interfaces

```go
// IKeeper defines the interface for the veid keeper
type IKeeper interface {
    // Identity record management
    GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (types.IdentityRecord, bool)
    SetIdentityRecord(ctx sdk.Context, record types.IdentityRecord) error
    CreateIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityRecord, error)

    // Score retrieval and updates
    GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool)
    UpdateScore(ctx sdk.Context, address sdk.AccAddress, score uint32, scoreVersion string) error

    // Scope management
    UploadScope(ctx sdk.Context, address sdk.AccAddress, scope *types.IdentityScope) error
    GetScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) (types.IdentityScope, bool)

    // Identity Wallet (VE-209)
    CreateIdentityWallet(ctx sdk.Context, accountAddr sdk.AccAddress, ...) (*types.IdentityWallet, error)
    GetWallet(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityWallet, bool)
    UpdateWalletScore(ctx sdk.Context, accountAddr sdk.AccAddress, newScore uint32, ...) error
}
```

---

## Identity Score System

### Score Range

- **Minimum:** 0
- **Maximum:** 100

### Identity Tiers

| Tier | Name       | Minimum Score | Capabilities                                   |
| ---- | ---------- | ------------- | ---------------------------------------------- |
| 0    | Unverified | 0             | Browse marketplace, view public offerings      |
| 0.5  | Pending    | 0             | Identity scopes submitted, awaiting ML scoring |
| 1    | Basic      | 1             | Basic marketplace access                       |
| 2    | Standard   | 30            | Standard orders                                |
| 3    | Verified   | 60            | Most operations                                |
| 4    | Trusted    | 85            | Provider registration, validator ops           |

### Role Requirements

| Role            | Min VEID Score | Additional Requirements   |
| --------------- | -------------- | ------------------------- |
| GenesisAccount  | N/A            | Pre-verified              |
| Administrator   | 85             | MFA required              |
| Moderator       | 70             | MFA for sensitive ops     |
| Validator       | 85             | MFA + Governance approval |
| ServiceProvider | 70             | MFA (FIDO2) required      |
| Customer        | 50             | For marketplace access    |
| SupportAgent    | 70             | MFA for data access       |

### Sensitive Transaction Requirements

| Transaction                  | Min Score | MFA Required                |
| ---------------------------- | --------- | --------------------------- |
| ProviderRegistration         | 70        | VEID + FIDO2                |
| ValidatorRegistration        | 85        | VEID + FIDO2 + Gov approval |
| LargeWithdrawal (>10,000 VE) | 70        | VEID + FIDO2                |
| GovernanceProposalCreate     | 70        | VEID + FIDO2                |
| HighValueOrder (>1,000 VE)   | 70        | VEID + FIDO2                |

---

## Available Messages (Tx Commands)

### VEID Module Messages

| Message                       | Description                                   | Authorization          |
| ----------------------------- | --------------------------------------------- | ---------------------- |
| `MsgUploadScope`              | Upload identity scope (selfie, ID doc, video) | User + Approved Client |
| `MsgRevokeScope`              | Revoke an identity scope                      | Scope owner            |
| `MsgRequestVerification`      | Request verification for a scope              | Scope owner            |
| `MsgUpdateVerificationStatus` | Update verification status                    | **Validators only**    |
| `MsgUpdateScore`              | Update identity score                         | **Validators only**    |
| `MsgCreateIdentityWallet`     | Create identity wallet                        | User                   |
| `MsgAddScopeToWallet`         | Add scope to wallet                           | Wallet owner           |
| `MsgRevokeScopeFromWallet`    | Revoke scope from wallet                      | Wallet owner           |
| `MsgUpdateParams`             | Update module params                          | **Governance only**    |

### CLI Commands (from runbooks)

```bash
# Query commands
virtengine query veid verification-request REQ_ID
virtengine query veid verification-result REQ_ID
virtengine query veid pending-verifications --limit 10
virtengine query veid model-status
virtengine query veid model-registry
virtengine query veid approved-clients
virtengine query veid validator-keys VALIDATOR_ADDR

# Transaction commands (require validator/governance)
virtengine tx gov submit-proposal upgrade-veid-model ...
```

---

## Available Queries

The VEID module provides 11 query methods via gRPC:

| Query                       | Description                        |
| --------------------------- | ---------------------------------- |
| `QueryIdentityRecord`       | Get identity record for an address |
| `QueryIdentityScore`        | Get current identity score         |
| `QueryVerificationRequest`  | Get verification request status    |
| `QueryVerificationResult`   | Get verification result            |
| `QueryPendingVerifications` | List pending verification requests |
| `QueryApprovedClients`      | List approved capture clients      |
| `QueryModelStatus`          | Get current ML model status        |
| `QueryModelRegistry`        | Get registered ML models           |
| `QueryValidatorKeys`        | Get validator VEID keys            |
| `QueryWallet`               | Get identity wallet                |
| `QueryWalletPublicMetadata` | Get public wallet info             |

---

## MFA Module Overview

### Structure

```
x/mfa/
├── keeper/
│   ├── keeper.go           # Main keeper with IKeeper interface
│   ├── msg_server.go       # Message handlers
│   ├── gating.go           # MFA gating logic
│   ├── fido2_verify.go     # FIDO2/WebAuthn verification
│   ├── sessions.go         # Authorization session management
│   └── verification.go     # Challenge verification
└── types/
    ├── msgs.go             # Message types
    ├── factors.go          # Factor types (FIDO2, TOTP, SMS, Email)
    ├── challenge.go        # Challenge types
    ├── authorization_policy.go
    └── sensitive_tx.go     # Sensitive transaction configs
```

### Factor Types

| Factor Type    | Security Level | Description           |
| -------------- | -------------- | --------------------- |
| FIDO2/WebAuthn | Very High      | Hardware security key |
| TOTP           | High           | Time-based OTP        |
| SMS            | Medium         | SMS verification      |
| Email          | Medium         | Email verification    |
| BackupCodes    | Medium         | Recovery codes        |

### MFA Messages

| Message                  | Description               |
| ------------------------ | ------------------------- |
| `MsgEnrollFactor`        | Enroll a new MFA factor   |
| `MsgRevokeFactor`        | Revoke an enrolled factor |
| `MsgSetMFAPolicy`        | Set account MFA policy    |
| `MsgCreateChallenge`     | Create MFA challenge      |
| `MsgVerifyChallenge`     | Verify challenge response |
| `MsgAddTrustedDevice`    | Add trusted device        |
| `MsgRemoveTrustedDevice` | Remove trusted device     |

---

## Testing Infrastructure

### Existing Test Fixtures

The project has extensive test fixtures in:

- `tests/e2e/veid_fixtures.go` - Deterministic test fixtures
- `tests/e2e/helpers/veid_helpers.go` - Test helper functions
- `tests/e2e/veid_e2e_test.go` - E2E test cases
- `tests/e2e/veid_onboarding_test.go` - Onboarding flow tests

### Key Test Helpers

```go
// Create identity record for test account
CreateIdentityRecordForAccount(t, app, ctx, account)

// Upload encrypted scope
UploadScope(t, msgServer, ctx, customer, client, params)

// Update account score (direct keeper call - bypasses validator check)
UpdateAccountScore(t, app, ctx, account, score)

// Verify account tier
VerifyAccountTier(t, app, ctx, account, expectedTier)
```

### Deterministic Test Constants

```go
const (
    DeterministicSeed     = 42
    TestChainID           = "virtengine-e2e-1"
    TestBlockTimeUnix     = 1700000000
    TestClientID          = "ve-e2e-capture-app"
    TestDeviceFingerprint = "e2e-device-fingerprint-001"
    TestModelVersion      = "veid-score-v1.0.0-e2e"
)
```

---

## Setting Up Test Identities

### Method 1: Direct Keeper Call (Unit/Integration Tests)

In tests, you can directly call the keeper's `SetScore` method which bypasses the validator authorization check:

```go
// In test code (bypasses validator check)
err := app.Keepers.VirtEngine.VEID.SetScore(ctx, accountAddr, 80, "test-model-v1")
require.NoError(t, err)

// Or using the test helper
UpdateAccountScore(t, app, ctx, account, 80)
```

### Method 2: Localnet Genesis Patching

The `scripts/init-chain.sh` already patches genesis to disable MFA and identity gating:

```bash
# From init-chain.sh - patch_genesis_for_localnet()
jq '.app_state.mfa.params.require_at_least_one_factor = false |
    .app_state.mfa.sensitive_tx_configs = [] |
    .app_state.mktplace.params.enable_mfa_gating = false |
    .app_state.mktplace.params.enable_identity_gating = false' genesis.json
```

### Method 3: Pre-seeding Genesis with Identity Records

You can add identity records to genesis:

```go
// In genesis state
veidGenesis := veidtypes.GenesisState{
    IdentityRecords: []veidtypes.IdentityRecord{
        {
            AccountAddress: "virtengine1provider...",
            CurrentScore:   80,
            Tier:           veidtypes.IdentityTierVerified,
            Status:         veidtypes.AccountStatusVerified,
        },
    },
    ApprovedClients: []veidtypes.ApprovedClient{
        testClient.ToApprovedClient(),
    },
}
```

### Method 4: Mock Validator for MsgUpdateScore

For integration tests, you can mock the staking keeper to make an address appear as a validator:

```go
// Create mock staking keeper
stakingKeeper := NewMockStakingKeeper()
stakingKeeper.AddValidator(valAddr, stakingtypes.Bonded)
keeper.SetStakingKeeper(stakingKeeper)

// Now this address can submit MsgUpdateScore
msg := types.NewMsgUpdateScore(valAddr.String(), accountAddr.String(), 80, "v1.0.0")
_, err := msgServer.UpdateScore(ctx, msg)
```

---

## Admin/Bypass Mechanisms

### Governance Authority

The governance module account can bypass VEID checks for governance messages:

```go
// In ante_veid.go
func (d VEIDDecorator) isGovernanceAuthority(signer sdk.AccAddress) bool {
    return signer.String() == d.govAuthority
}
```

### Genesis Accounts

Genesis accounts are pre-verified and can nominate administrators:

```go
// RoleGenesisAccount has trust level 100 (highest)
RoleGenesisAccount.TrustLevel() // returns 100
```

### Localnet Configuration

The localnet disables identity gating for development:

```json
{
  "mfa": {
    "params": {
      "require_at_least_one_factor": false
    },
    "sensitive_tx_configs": []
  },
  "mktplace": {
    "params": {
      "enable_mfa_gating": false,
      "enable_identity_gating": false
    }
  }
}
```

---

## What Needs to Be Implemented

### Missing CLI Commands for Testing

1. **Admin Score Override CLI**
   - Currently, `MsgUpdateScore` requires validator authorization
   - Need: `virtengine tx veid admin-set-score [address] [score] --from governance`
   - Requires governance proposal or admin message

2. **Test Identity Setup Script**
   - Script to create identity records with pre-set scores for localnet
   - Should be integrated into `init-chain.sh`

3. **VEID CLI Commands**
   - The module panics on `GetQueryCmd()` and `GetTxCmd()`
   - CLI commands exist in runbooks but may not be fully implemented
   - Need verification that all query commands work

### Recommended Implementations

#### Option A: Genesis Pre-seeding Script

Create a script to add pre-verified identities to genesis:

```bash
#!/bin/bash
# scripts/seed-test-identities.sh

# Add provider identity with score 80
ve genesis add-veid-identity $(ve keys show provider -a) 80 verified

# Add customer identity with score 50
ve genesis add-veid-identity $(ve keys show alice -a) 50 verified
```

#### Option B: Governance Proposal for Admin Score

Create a governance proposal type for setting scores:

```protobuf
message MsgAdminSetScore {
  string authority = 1;  // Must be x/gov module account
  string account_address = 2;
  uint32 score = 3;
  string reason = 4;
}
```

#### Option C: Test Mode Flag

Add a test mode flag that bypasses validator checks:

```go
type Params struct {
    // ... existing params
    TestModeEnabled bool `json:"test_mode_enabled"`
}

// In msg_server.go
if !params.TestModeEnabled && !ms.keeper.IsValidator(ctx, sender) {
    return nil, types.ErrUnauthorized
}
```

---

## Quick Reference: Testing Provider Registration

To test provider registration (requires VEID ≥70 + FIDO2):

### 1. Start Localnet (MFA/Identity gating disabled)

```bash
./scripts/localnet.sh start
```

### 2. Check Identity Gating is Disabled

```bash
virtengine query mktplace params | grep -i gating
# Should show: enable_identity_gating: false
```

### 3. Register Provider

```bash
virtengine tx provider create-provider \
  --from provider \
  --keyring-backend test \
  --chain-id virtengine-localnet-1
```

### For Full Identity Testing

Use the E2E test suite:

```bash
# Run VEID onboarding tests
go test -v -tags="e2e.integration" ./tests/e2e/... -run TestVEIDOnboarding
```

---

## References

- [VEID Flow Specification](veid-flow-spec.md)
- [Architecture Overview](architecture.md)
- [Testing Guide](testing-guide.md)
- [Validator VEID Operations](training/validator/veid-operations.md)
