# ADR 003: Technical Debt Refactoring (REFACTOR-001)

## Changelog

* 2026-01-30: Initial implementation

## Status

Accepted

## Context

As part of ongoing code quality improvements (REFACTOR-001), we conducted a comprehensive analysis of technical debt across the VirtEngine codebase. This ADR documents the findings, decisions made, and provides guidance for future maintenance.

## Analysis Summary

### TODO Comments Categorization

#### Category 1: Excluded Test Files (Low Priority)
Multiple test files are excluded via build tags waiting for API stabilization:
- `x/config/keeper/keeper_test.go` - sdk.NewContext API
- `x/settlement/keeper/{settlement,rewards,events,escrow}_test.go` - settlement API
- `x/staking/{types,keeper}_test.go` - staking API
- `x/enclave/keeper/keeper_test.go` - sdk.NewContext API
- `x/hpc/keeper/{scheduling,rewards,keeper}_test.go` - HPC API
- `x/escrow/keeper/keeper_test.go` - testutil.VECoinRandom API
- `x/mfa/{types,keeper}_test.go` - MFA API
- `x/market/handler/handler_test.go` - market handler API
- `x/veid/types/{email,domain}_verification_test.go` - VE-1007

**Decision**: These tests are intentionally excluded and should be re-enabled once their respective APIs stabilize. No action required at this time.

#### Category 2: Implementation TODOs (Addressed)
The following implementation TODOs were addressed:
- `x/escrow/keeper/key.go:163,185` - Added validation for scope, xid, pid parameters

The following implementation TODOs are deferred:
- `x/enclave/keeper/heartbeat.go:218,241` - Signature/attestation verification (requires TEE hardware)
- `x/staking/module.go:111` - Invariants addition (performance monitoring)
- `x/veid/keeper/market_integration.go:467` - Delegation age checking
- `x/veid/keeper/score.go:598,611` - Verification count/score tracking per validator
- `x/market/keeper/keeper.go:803,868` - String ID parsing, completion time tracking

#### Category 3: Enclave Runtime TODOs (Documentation Only)
The `pkg/enclave_runtime/` files contain TODOs that are intentional placeholders for hardware-specific implementations. These require:
- AWS Nitro NSM library integration
- Intel SGX SDK integration
- AMD SEV-SNP SDK integration

**Decision**: These are documented architectural placeholders, not technical debt. They will be implemented when hardware support is added.

### Dead Code Elimination

The following unused code was removed or marked:

1. **Removed**:
   - `x/config/types/keys.go`: `decodeInt64()` - unused helper function
   - `x/settlement/keeper/keeper.go`: `parseAmount()` - unused helper function  
   - `x/market/types/marketplace/waldur_bridge.go`: `generateNonce()` - wrapper function
   - `x/veid/types/verification_property_test.go`: `detectCycle()` - unused test helper
   - `x/delegation/types/types_test.go`: `testDelegatorAddr2` - unused test variable
   - `x/delegation/keeper/rewards_test.go`: `testDelegator2Addr` - unused test variable

2. **Marked with nolint (Reserved for Future)**:
   - `x/delegation/keeper/keeper.go`: Sequence management functions (getNextDelegationSequence, getNextUnbondingSequence, getNextRedelegationSequence)
   - `x/roles/types/grpc_handlers.go`: gRPC service descriptors and handlers (reserved for direct gRPC registration)
   - `x/roles/types/query.go`: Query service descriptor
   - `x/veid/types/grpc_handlers.go`: Msg and Query service descriptors

### Deprecated API Patterns

The "deprecated" patterns found are legitimate domain concepts:
- `pkg/capture_protocol/types.go`: Key rotation support (DeprecatedKey, DeprecatedKeyExpiry)
- `x/veid/types/`: Model lifecycle states (ModelStatusDeprecated, PipelineVersionStatusDeprecated)

**Decision**: These are part of the domain model for versioning and key rotation. No changes required.

## Consequences

### Positive
- Reduced dead code improves maintainability
- Clear documentation of intentional placeholder code
- Input validation added to escrow key functions improves robustness
- nolint comments explain why certain code is kept

### Negative
- Some technical debt remains in deferred TODOs
- Test files remain excluded until APIs stabilize

### Neutral
- Enclave runtime TODOs remain as architectural documentation

## Future Work

1. Re-enable excluded test files as APIs stabilize
2. Implement validator verification tracking (x/veid/keeper/score.go)
3. Add order completion time tracking (x/market/keeper/keeper.go)
4. Implement enclave hardware integrations when hardware support is needed

## References

- REFACTOR-001 ticket
- [Cosmos SDK Best Practices](https://docs.cosmos.network/main/build/building-modules/intro)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
