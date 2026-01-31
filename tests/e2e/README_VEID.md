# VEID E2E Test Suite

This directory contains end-to-end tests for the VirtEngine Identity (VEID) module, covering the complete onboarding and verification flows.

## Overview

The VEID E2E test suite validates:

1. **Account Onboarding**: Identity record creation and scope uploads
2. **Verification Flows**: SSO, Email, and SMS verification with OTP
3. **ML Scoring**: Deterministic scoring and tier transitions
4. **Attestations**: Verification attestation recording and validation
5. **Rejection Paths**: Expired OTP, low scores, VoIP blocking, invalid signatures

## Test Files

| File | Description |
|------|-------------|
| `veid_fixtures.go` | Deterministic test fixtures and mock data |
| `veid_e2e_test.go` | Main E2E test suite |

## Running Tests

### Prerequisites

1. **Go 1.22+** installed
2. **CGO enabled** (for crypto dependencies)
3. Build dependencies installed: `make deps`

### Run All VEID E2E Tests

```bash
# Standard run
go test -v -tags="e2e.integration" -timeout 20m ./tests/e2e/... -run "TestVEIDE2E"

# With race detection
go test -v -tags="e2e.integration" -race -timeout 30m ./tests/e2e/... -run "TestVEIDE2E"

# With coverage
go test -v -tags="e2e.integration" -coverprofile=coverage.txt ./tests/e2e/... -run "TestVEIDE2E"
```

### Run Specific Tests

```bash
# Onboarding flow only
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestVEIDE2E/TestCompleteOnboardingFlow"

# Email verification only
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestVEIDE2E/TestEmailVerification"

# SSO verification only
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestVEIDE2E/TestSSOVerificationFlow"

# Rejection paths only
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestVEIDE2E/Test.*Rejection"
```

### Run via Make

```bash
# All integration tests
make test-integration

# VEID-specific (if added to Makefile)
make test-veid-e2e
```

## Test Structure

### Fixtures (`veid_fixtures.go`)

The fixtures provide deterministic test data with a fixed seed (42) for reproducibility:

```go
// Deterministic seed for all tests
const DeterministicSeed = 42

// Fixed timestamp for reproducible tests
const TestBlockTimeUnix = 1700000000
```

#### Available Fixtures

| Category | Fixtures |
|----------|----------|
| **Scopes** | `SelfieScope()`, `IDDocumentScope()`, `FaceVideoScope()`, `LowScoreScope()` |
| **Email** | `ValidEmailVerification()`, `OrgEmailVerification()`, `ExpiredEmailVerification()` |
| **SMS** | `ValidSMSVerification()`, `VoIPSMSVerification()` |
| **SSO** | `GoogleSSOVerification()`, `MicrosoftSSOVerification()`, `GitHubSSOVerification()` |
| **Attestations** | `FacialVerificationAttestation()`, `LivenessCheckAttestation()`, `DocumentVerificationAttestation()` |
| **Transitions** | `UnverifiedToBasic()`, `BasicToVerified()`, `VerifiedToTrusted()` |

### Test Suite (`veid_e2e_test.go`)

The main test suite includes:

| Test | Description |
|------|-------------|
| `TestCompleteOnboardingFlow` | Full onboarding: create account → upload scopes → verify → tier up |
| `TestEmailVerificationFlow` | Email OTP verification with challenge/response |
| `TestEmailVerificationExpiredOTP` | Expired OTP rejection |
| `TestSMSVerificationFlow` | SMS OTP verification with mobile number |
| `TestSMSVerificationVoIPBlocking` | VoIP number detection and blocking |
| `TestSSOVerificationFlow` | SSO verification for Google, Microsoft, GitHub |
| `TestMLScoringAndTierTransitions` | Score updates and tier progression |
| `TestAttestationRecording` | Attestation creation and validation |
| `TestMarketplaceVEIDGating` | Marketplace order gating by VEID score |
| `TestLowScoreRejection` | Low score tier assignment |
| `TestInvalidClientSignatureRejection` | Invalid signature rejection |
| `TestMaxAttemptsExceeded` | OTP max attempts enforcement |
| `TestScopeRevocation` | Scope revocation flow |

## Determinism Requirements

For blockchain consensus, all ML scoring must be deterministic. The tests enforce:

```go
// Environment variables for deterministic inference
TF_DETERMINISTIC_OPS=1
TF_CUDNN_DETERMINISTIC=1
PYTHONHASHSEED=42
CUDA_VISIBLE_DEVICES=''  // Force CPU-only
```

The `DeterminismController` in `pkg/inference` ensures:
- Fixed random seeds
- CPU-only execution (no GPU variance)
- Reproducible floating-point operations

## CI Integration

The VEID E2E tests run automatically in CI via `.github/workflows/veid-e2e.yaml`:

- **On push/PR**: Runs when `x/veid/**`, `tests/e2e/veid_*.go`, or `pkg/inference/**` are modified
- **Manual dispatch**: Can be triggered manually with verbose logging option

### CI Jobs

| Job | Purpose |
|-----|---------|
| `veid-e2e` | Full E2E integration tests |
| `veid-unit` | VEID module unit tests with coverage |
| `veid-determinism` | Cross-machine determinism verification |

## Adding New Tests

1. **Add fixtures** to `veid_fixtures.go` if new test data is needed
2. **Add test method** to `VEIDE2ETestSuite` in `veid_e2e_test.go`
3. **Follow naming convention**: `Test<Feature><Scenario>`
4. **Use deterministic data**: Always use fixtures or `DeterministicNonce()`
5. **Log progress**: Use `s.T().Log()` for test traceability

### Example: Adding a new verification test

```go
func (s *VEIDE2ETestSuite) TestNewVerificationFlow() {
    ctx := s.ctx
    customer := sdktestutil.AccAddress(s.T())

    // Use fixtures
    fixture := NewVerificationFixture()

    // Test logic...

    // Assert expectations
    require.NoError(s.T(), err)
    require.Equal(s.T(), expected, actual)

    s.T().Log("✅ New verification flow test passed")
}
```

## Troubleshooting

### Test Failures

1. **Signature validation errors**: Ensure `TestClientID` matches approved client in genesis
2. **Tier transition failures**: Check `ComputeTierFromScore()` thresholds
3. **Timing issues**: Use `FixedTimestamp()` and `FixedTimestampPlus()` for deterministic time

### Build Issues

```bash
# Clear cache and rebuild
make clean
go clean -cache
make virtengine
```

### Verbose Logging

```bash
# Enable verbose test output
VEID_E2E_VERBOSE=true go test -v -tags="e2e.integration" ./tests/e2e/...
```

## Related Documentation

- [VEID Module Documentation](../../x/veid/README.md)
- [Inference Package](../../pkg/inference/README.md)
- [Testing Guide](../../_docs/testing-guide.md)
- [Development Environment](../../_docs/development-environment.md)
