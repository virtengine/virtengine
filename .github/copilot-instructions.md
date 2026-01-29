# VirtEngine Copilot Instructions

## Project Overview

VirtEngine is a Cosmos SDK-based blockchain for decentralized cloud computing with ML-powered identity verification (VEID). The module system in `x/` follows strict patterns for consensus safety.

## Architecture Quick Reference

```
app/          → Cosmos SDK app wiring, ante handlers, genesis
x/            → Custom blockchain modules (veid, mfa, encryption, market, escrow, roles, hpc)
pkg/          → Off-chain services (provider_daemon, inference, workflow)
ml/           → Python ML pipelines (facial_verification, liveness_detection, ocr_extraction)
cmd/          → CLI binaries (virtengine main binary, provider-daemon, benchmark-daemon)
tests/e2e/    → Integration tests for CLI and gRPC endpoints
```

**Key Module Boundaries:**

- `x/veid` - Identity verification with encrypted scopes and ML scoring
- `x/mfa` - Multi-factor authentication gating for sensitive transactions
- `x/encryption` - Public-key encryption (X25519-XSalsa20-Poly1305 envelopes)
- `x/market` - Marketplace orders, bids, leases with escrow integration
- `pkg/provider_daemon` - Off-chain bidding engine with K8s/SLURM/VMware adapters
- `pkg/inference` - TensorFlow scorer with determinism controls for consensus

## Build & Development Commands

```bash
make                    # Build virtengine binary to .cache/bin
make test               # Run unit tests (excludes mocks)
make test-integration   # Run e2e tests with -tags="e2e.integration"
make lint-go            # Run golangci-lint with all linters
make generate           # Run go generate (mockery for mocks)
./scripts/localnet.sh start   # Start local dev network (Docker)
```

**Build tools are cached in `.cache/bin`** - see [make/setup-cache.mk](make/setup-cache.mk) for versions.

## Cosmos SDK Module Patterns

### Keeper Interface Pattern

Every module defines `IKeeper` interface before the concrete `Keeper` struct:

```go
// x/market/keeper/keeper.go - pattern to follow
type IKeeper interface {
    CreateOrder(ctx sdk.Context, gid dtypes.GroupID, spec dtypesBeta.GroupSpec) (types.Order, error)
    GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool)
    WithOrders(ctx sdk.Context, fn func(types.Order) bool)
    // ... all public methods
}

type Keeper struct {
    cdc       codec.BinaryCodec
    skey      storetypes.StoreKey
    authority string  // Always x/gov module account for MsgUpdateParams
}
```

### Module Registration

Modules panic on `GetQueryCmd()`/`GetTxCmd()` - CLI commands are handled separately:

```go
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
    panic("virtengine modules do not export cli commands via cosmos interface")
}
```

### Genesis and Validation

Every module type must implement:

- `DefaultGenesisState()` returning sensible defaults
- `GenesisState.Validate()` for genesis validation
- Store keys defined in `types/keys.go`

## VEID Identity System

**Encryption is mandatory** - sensitive identity data uses the envelope format:

```go
// Envelope structure for all encrypted payloads
type EncryptionEnvelope struct {
    RecipientFingerprint string  // Validator's key fingerprint
    Algorithm            string  // "X25519-XSalsa20-Poly1305"
    Ciphertext           []byte
    Nonce                []byte
}
```

**Signature validation** - all scope uploads require three signatures:

1. Client signature (from approved capture app)
2. User signature (from wallet)
3. Salt binding (prevents replay)

## Provider Daemon Architecture

The `pkg/provider_daemon/` implements off-chain provider services that bridge on-chain marketplace orders to infrastructure provisioning.

### Core Components

| Component       | File             | Purpose                                          |
| --------------- | ---------------- | ------------------------------------------------ |
| Bid Engine      | `bid_engine.go`  | Watches orders, computes pricing, submits bids   |
| Key Manager     | `key_manager.go` | Provider key storage (hardware/ledger support)   |
| Manifest Parser | `manifest.go`    | Validates deployment manifests (v1, v2beta1)     |
| Usage Meter     | `usage_meter.go` | Collects metrics, submits on-chain usage records |

### Infrastructure Adapters

All adapters implement the same lifecycle: `Pending → Deploying → Running → Stopping → Stopped → Terminated`

```go
// Valid state transitions defined in kubernetes_adapter.go
var validTransitions = map[WorkloadState][]WorkloadState{
    WorkloadStatePending:   {WorkloadStateDeploying, WorkloadStateFailed},
    WorkloadStateDeploying: {WorkloadStateRunning, WorkloadStateFailed, WorkloadStateStopped},
    WorkloadStateRunning:   {WorkloadStatePaused, WorkloadStateStopping, WorkloadStateFailed},
    // ...
}
```

| Adapter    | File                    | Backend                                 |
| ---------- | ----------------------- | --------------------------------------- |
| Kubernetes | `kubernetes_adapter.go` | K8s namespace isolation, pod deployment |
| OpenStack  | `openstack_adapter.go`  | VM provisioning via Waldur              |
| AWS        | `aws_adapter.go`        | EC2/VPC/EBS/S3 via Waldur               |
| Azure      | `azure_adapter.go`      | Azure VMs via Waldur                    |
| VMware     | `vmware_adapter.go`     | vSphere integration                     |
| Ansible    | `ansible_adapter.go`    | Playbook execution with vault secrets   |

### Usage Metering

Usage records are submitted on-chain for escrow settlement:

```go
type ResourceMetrics struct {
    CPUMilliSeconds    int64  // CPU usage in milliseconds
    MemoryByteSeconds  int64  // Memory usage in byte-seconds
    StorageByteSeconds int64  // Storage usage in byte-seconds
    GPUSeconds         int64  // GPU usage in seconds
}
```

### Security Properties

- Provider keys support hardware/ledger/non-custodial storage
- Secrets are never logged or stored in plaintext
- Vault passwords are cleared from memory after use
- All on-chain submissions are cryptographically signed

## ML Inference Determinism (Critical)

All ML scoring must be deterministic for consensus. The `pkg/inference` package enforces:

```go
// ALWAYS use these settings for blockchain consensus
type DeterminismConfig struct {
    ForceCPU         bool   // true - no GPU variance
    RandomSeed       int64  // fixed seed (default: 42)
    DeterministicOps bool   // true - TF deterministic ops
}
```

**Python ML pipelines** in `ml/` use `requirements-deterministic.txt` with pinned versions.

## Testing Conventions

- Unit tests: `*_test.go` alongside source files
- E2E tests: `tests/e2e/*_{cli,grpc}_test.go` - test CLI and gRPC separately
- Integration: `tests/integration/` - full module interaction tests
- Always add `goleak.VerifyNoLeaks(t)` for goroutine tests

```bash
# Run specific module tests
go test -v ./x/veid/...
go test -v ./pkg/provider_daemon/...
```

## External Dependencies

- **Cosmos SDK v0.53.x** - upstream `github.com/cosmos/cosmos-sdk`
- **CometBFT v0.38.x** - upstream `github.com/cometbft/cometbft`
- **IBC-Go v10** - for cross-chain communication
- **goleveldb** - pinned to specific commit (see `go.mod` replace directives)
- **Shared Go libraries** - `github.com/virtengine/go` (SDK utilities, types, CLI helpers)

### Required Shared Repositories

The codebase depends on shared Go libraries that must exist at:

- `github.com/virtengine/go` - SDK utilities, types, and CLI helpers
- `github.com/virtengine/virtengine/sdk/go/cli` - CLI flag definitions
- `github.com/virtengine/virtengine/sdk/go/sdl` - SDL parsing utilities

These contain types for: audit, cert, deployment, escrow, gov, market, provider, staking, take modules.

## Branch Strategy

- `main` - active development (odd minor versions like v0.9.x)
- `mainnet/main` - stable releases (even minor versions like v0.8.x)

## Commit Message Conventions

VirtEngine uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages and PR titles.

### Format

```
type(scope): description
```

### Valid Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Code style (formatting, semicolons, etc.)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `build` - Build system or dependencies
- `ci` - CI/CD configuration
- `chore` - Other changes (maintenance)
- `revert` - Reverting a previous commit

### Valid Scopes

- `veid` - Identity verification module
- `mfa` - Multi-factor authentication module
- `encryption` - Encryption module
- `market` - Marketplace module
- `escrow` - Escrow module
- `roles` - Roles module
- `hpc` - HPC module
- `provider` - Provider daemon
- `sdk` - SDK packages
- `cli` - CLI commands
- `app` - Application wiring
- `deps` - Dependencies
- `ci` - CI/CD
- `api` - API changes

### Examples

```bash
feat(veid): add identity verification flow
fix(market): resolve bid race condition
docs: update contributing guidelines
chore(deps): bump cosmos-sdk to v0.53.1
ci: fix failing CI checks
```

### Breaking Changes

Add `!` after type/scope for breaking changes:

```bash
feat(api)!: change response format
```

## Common Pitfalls

1. **Don't use standard `make` on macOS** - requires GNU Make 4+ (install via Homebrew)
2. **Module authority must be x/gov account** - never hardcode addresses for `MsgUpdateParams`
3. **Always validate context deadlines** in keeper methods that do ML inference
4. **Use `storetypes.StoreKey`** not deprecated `sdk.StoreKey`
5. **Forked dependencies** - check `go.mod` replace directives before updating deps
