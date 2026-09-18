# VirtEngine Testing Guide

This guide covers test execution, debugging, and coverage practices for the VirtEngine blockchain.

## Overview

VirtEngine tests are organized into three tiers:

| Tier | Location | Purpose |
|------|----------|---------|
| Unit Tests | `x/*/types/`, `x/*/keeper/`, `pkg/*/` | Module logic, types validation |
| Integration Tests | `tests/integration/` | Cross-module interactions |
| E2E Tests | `tests/e2e/` | CLI and gRPC endpoint testing |

## Running Tests

### Unit Tests

Run all module and package unit tests:

```bash
# All unit tests
go test ./x/... ./pkg/...

# Specific module
go test ./x/veid/...
go test ./x/market/...

# Specific package
go test ./pkg/provider_daemon/...
go test ./pkg/inference/...
```

### Integration Tests

Integration tests require the `e2e.integration` build tag:

```bash
go test -tags="e2e.integration" ./tests/integration/...
```

### E2E Tests

Use the Makefile target for full end-to-end testing:

```bash
make test-integration
```

### Complete Test Target Reference (make)

Every target below was confirmed with `make -n <target>` from the repo root; the source column
cites the recipe that runs. Two variables shape most of them:

- `TEST_MODULES` (`make/test-integration.mk:3`) — `$(shell $(GO) list ./... | grep -v '/mocks')`,
  i.e. every package in the main module **except** those under a `/mocks` path. It is expanded
  when make parses the file, so newly added packages are picked up automatically.
- `BUILD_TAGS` (`Makefile:49`) — defaults to `osusergo,netgo,hidraw,ledger`.

| Target | Source | What it actually runs |
|--------|--------|-----------------------|
| `make test` | `make/test-integration.mk:10` | `go test -v -timeout 600s $(TEST_MODULES)` — verbose, 10-minute timeout |
| `make test-nocache` | `make/test-integration.mk:14` | `go test -count=1 $(TEST_MODULES)` — `-count=1` bypasses the Go test cache |
| `make test-full` | `make/test-integration.mk:18` | `go test -v -tags=$(BUILD_TAGS) $(TEST_MODULES)` — adds the default build tags |
| `make test-integration` | `make/test-integration.mk:22` | `go test -v -tags="e2e.integration" $(TEST_MODULES)` |
| `make test-coverage` | `make/test-integration.mk:26` | `CGO_ENABLED=1 go test -tags=$(BUILD_TAGS) -coverprofile=coverage.txt -covermode=atomic -timeout=20m $(TEST_MODULES)` |
| `make test-vet` | `make/test-integration.mk:33` | `go vet -mod=readonly ./...` |
| `make test-compatibility` | `make/test-integration.mk:41` | `go test -v -tags="e2e.compatibility" ./tests/compatibility/...` **then** `go test -v ./pkg/compatibility/...` |
| `make test-compatibility-full` | `make/test-integration.mk:47` | The same two suites, with `-coverprofile=coverage-compatibility.txt` and `-coverprofile=coverage-pkg-compatibility.txt` |
| `make test-bins` | `make/releasing.mk:68` | `docker run ... goreleaser-cross` with `.goreleaser-test-bins.yaml`, `--snapshot --skip=publish,validate --clean` — a release dry-run build, not a Go test |

Details that are easy to get wrong:

- **`test-vet` is not a test run, and it is broader than the test targets.** It runs `go vet` over
  `./...`, which includes the `/mocks` packages that `TEST_MODULES` deliberately excludes.
- **`test-coverage` is the only target that forces `CGO_ENABLED=1`**, and the only one using
  `-covermode=atomic`. It is genuinely slow — `-timeout=20m` is the intended budget, not a guard
  against hangs.
- **`test-compatibility` takes the `e2e.compatibility` tag on the first invocation only**; the
  second (`./pkg/compatibility/...`) runs untagged.
- **`make test-bins` needs Docker plus an image tag from the environment.**
  `make/releasing.mk:4` builds the image reference as
  `ghcr.io/goreleaser/goreleaser-cross:$(GOTOOLCHAIN_SEMVER)`. `GOTOOLCHAIN_SEMVER` is exported by
  `.envrc` (from `script/tools.sh gotoolchain`) and is **not** set by the Makefile. In a shell
  without direnv the tag expands to nothing, giving the invalid reference
  `ghcr.io/goreleaser/goreleaser-cross:` — this fails before any build runs. With direnv active on
  the current toolchain the tag resolves to `v1.25.8`.
- `COVER_PACKAGES` (`make/test-integration.mk:1`) is defined but referenced by no target. It is
  dead code; nothing reads it.

### Simulation Tests (make)

`make/test-simulation.mk` drives the Cosmos SDK application simulations under `./app`
(`APP_DIR := ./app`, `Makefile:1`). All four share the same shape — `go test -mod=readonly ./app`
with `-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true`:

| Target | Source | Simulation test | Notable flags |
|--------|--------|-----------------|---------------|
| `make test-sim-fullapp` | `make/test-simulation.mk:5` | `TestFullAppSimulation` | `-Seed=99 -Period=5 -timeout 10m` |
| `make test-sim-import-export` | `make/test-simulation.mk:15` | `TestAppImportExport` | `-Seed=99 -Period=5 -timeout 10m` |
| `make test-sim-after-import` | `make/test-simulation.mk:20` | `TestAppSimulationAfterImport` | `-Seed=99 -Period=5 -timeout 10m` |
| `make test-sim-nondeterminism` | `make/test-simulation.mk:10` | `TestAppStateDeterminism` | `-Period=0 -timeout 24h` — **can take up to a day** |
| `make test-sims` | `make/test-simulation.mk:25` | all four, in the order above | aggregator; has no recipe of its own |

`test-sim-nondeterminism` is the outlier: `-Period=0` and a 24-hour `-timeout` because it hunts
consensus non-determinism, which is deliberately slow. Do not expect the 10-minute turnaround of
its siblings.

Unlike the targets in `make/test-integration.mk`, these four are **not** declared `.PHONY`
(`make/test-simulation.mk:5`). They work only because no file of the same name exists in the repo
root; a stray file named e.g. `test-sims` would silently suppress the target.

## Upgrade Testing

VirtEngine upgrade testing is split into unit checks and multi-validator e2e drills.

### Upgrade Handler Framework (Unit)

- Upgrade handler tests live under `upgrades/software/*` and validate state migrations at the keeper/store level.
- Upgrade test cases are defined in `tests/upgrade/test-cases.json` and validated by:

```bash
go test ./tests/upgrade
```

### Multi-Validator Upgrade Coordination (E2E)

This runs a local multi-validator network and submits a software upgrade proposal:

```bash
make -C tests/upgrade test
```

Key inputs:
- `tests/upgrade/test-cases.json` (module add/remove + migration expectations)
- `tests/upgrade/test-config.json` (validator topology)
- `tests/upgrade/testnet.json` (testnet state shaping for upgrade simulation)

### Upgrade Simulation on Testnet State

Use the upgrade harness to snapshot a network and replay an upgrade against the testnetified state:

```bash
make -C tests/upgrade prepare-state
make -C tests/upgrade test
```

Both commands require `ROOT_DIR` to be exported: `tests/upgrade/Makefile:1` is
`include $(ROOT_DIR)/make/init.mk`, evaluated before `make/init.mk` can default it. `ROOT_DIR` is
supplied by `.env` (`ROOT_DIR=${VIRTENGINE_ROOT}`) via direnv. In a bare shell the command fails
immediately with `Makefile:1: /make/init.mk: No such file or directory` — that is an unconfigured
environment, not a broken target.

### Rollback Procedure Drill

Simulate a failed upgrade and verify recovery:

1. Run the e2e upgrade test with the upgrade binary missing or invalid.
2. Confirm nodes halt at the upgrade height (cosmovisor logs show `UPGRADE NEEDED`).
3. Restore the correct upgrade binary under `cosmovisor/upgrades/<name>/bin`.
4. Restart validators and confirm block production resumes.

### Emergency Upgrade Procedure

For critical incidents requiring immediate action:

1. Add a height patch via `upgrades/types.RegisterHeightPatch` and gate behavior by height.
2. Cut a hotfix release and publish upgrade binaries.
3. Submit a governance upgrade or coordinate a privileged upgrade window (test on the upgrade harness first).
4. Validate post-upgrade invariants using the post-upgrade test worker in `tests/upgrade/types`.

### Upgrade Failure Scenarios (Documented)

- Missing or non-executable upgrade binary.
- Store key changes without corresponding `StoreLoader` entries.
- Migration handler panics or missing migration versions.
- State migration errors (e.g., invalid params or missing subspaces).
- Incompatible binary on a subset of validators (multi-validator desync).

Expected outcomes:
- Nodes halt at upgrade height or fail fast during migration.
- Logs capture the failing module/migration and the chain does not continue in a partially-migrated state.

## Coverage

Generate coverage reports. The Makefile form is `make test-coverage` (see the test target
reference above); the raw `go test` equivalents are:

```bash
# Basic coverage
go test -cover ./x/... ./pkg/...

# Coverage with HTML report
go test -coverprofile=coverage.out ./x/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific module
go test -cover ./x/veid/...
```

## Quick Test Commands

```bash
# Fast test run (no cache)
go test -count=1 ./x/...

# With timeout
go test -timeout 5m ./x/...

# Verbose run of every main-module package, 10-minute timeout
make test
```

## Test Organization

### Module Tests (`x/*/`)

Each blockchain module follows this structure:

```
x/veid/
├── keeper/
│   ├── keeper.go
│   ├── keeper_test.go      # Keeper unit tests
│   └── msg_server_test.go  # Message handler tests
├── types/
│   ├── keys.go
│   ├── keys_test.go        # Type tests
│   ├── genesis.go
│   └── genesis_test.go     # Genesis validation tests
└── module.go
```

### Integration Tests (`tests/integration/`)

Test cross-module interactions:

```
tests/integration/
├── escrow_market_test.go   # Escrow + Market integration
├── veid_mfa_test.go        # VEID + MFA integration
└── ...
```

### E2E Tests (`tests/e2e/`)

Test CLI and gRPC endpoints:

```
tests/e2e/
├── market_cli_test.go      # Market CLI commands
├── market_grpc_test.go     # Market gRPC queries
├── veid_cli_test.go        # VEID CLI commands
└── ...
```

Golden path marketplace → provision → usage → invoice → payout:

```
go test -tags="e2e.integration" ./tests/e2e/golden -run TestGoldenPathMarketplaceProvisionUsageInvoicePayout
```

Failure artifacts are written to `_build/artifacts/e2e` (override with `VE_E2E_ARTIFACTS_DIR`).

## Excluded Tests

### Build Tag: `// +build ignore`

Some test files are excluded from normal test runs using the ignore build tag:

```go
//go:build ignore
// +build ignore

package keeper_test
```

### Why Tests Are Excluded

1. **API Mismatches** - Tests depend on APIs that are still being stabilized
2. **Missing Dependencies** - External service dependencies not yet available
3. **Pending Refactoring** - Tests need updates after module restructuring

### Currently Excluded

Check for excluded tests:

```bash
# Find files with ignore tag
Get-ChildItem -Recurse -Filter "*_test.go" | Select-String -Pattern "//go:build ignore" -List | Select-Object Path
```

### Re-enabling Tests

To re-enable an excluded test:

1. Remove the build tag comments at the top of the file:
   ```go
   // REMOVE THESE LINES:
   //go:build ignore
   // +build ignore
   ```

2. Fix any API mismatches or missing dependencies

3. Run the test to verify:
   ```bash
   go test -v ./path/to/module/...
   ```

## Debugging Tests

### Verbose Output

```bash
# Verbose mode shows test names and timing
go test -v ./x/veid/...

# Very verbose with all logs
go test -v -count=1 ./x/veid/keeper/...
```

### Run Specific Tests

```bash
# Single test by name
go test -v -run TestCreateOrder ./x/market/keeper/...

# Pattern matching
go test -v -run "TestOrder.*" ./x/market/...

# Subtests
go test -v -run "TestOrder/with_escrow" ./x/market/...
```

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a specific test
dlv test ./x/veid/keeper/... -- -test.run TestVerifyIdentity

# Common dlv commands:
# b <func>    - set breakpoint
# c           - continue
# n           - next line
# s           - step into
# p <var>     - print variable
# q           - quit
```

### Debug Environment Variables

```bash
# Enable verbose SDK logging
$env:COSMOS_SDK_LOG_LEVEL = "debug"
go test -v ./x/...

# Trace goroutines
$env:GOTRACEBACK = "all"
go test -v ./x/...
```

### Test Timeouts

```bash
# Increase timeout for slow tests
go test -timeout 10m ./tests/integration/...

# Identify slow tests
go test -v -timeout 30s ./x/... 2>&1 | Select-String "SLOW"
```

## Coverage Goals

### Current Status

| Category | Passing | Total | Status |
|----------|---------|-------|--------|
| Module packages (`x/`) | 14 | 24 | 58% |
| Utility packages (`pkg/`) | ~8 | 12 | 67% |

### Target Coverage

- **Overall**: 80%+ line coverage
- **Keeper methods**: 90%+ coverage (critical business logic)
- **Type validation**: 100% coverage (all Validate() methods)
- **Genesis**: 100% coverage (genesis import/export)

### Priority Modules for Coverage

1. **x/veid** - Identity verification (critical security)
2. **x/escrow** - Payment handling (financial operations)
3. **x/market** - Order/bid/lease lifecycle
4. **x/mfa** - Authentication gating
5. **pkg/inference** - ML scoring (determinism critical)

### Checking Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./x/...

# View coverage by function
go tool cover -func=coverage.out | Select-String -Pattern "total:|veid|market"

# Find uncovered code
go tool cover -func=coverage.out | Where-Object { $_ -match "0.0%" }
```

### Coverage in CI

The project uses codecov for coverage tracking. See `codecov.yml` for configuration.

```bash
# Upload to codecov (CI only)
bash <(curl -s https://codecov.io/bash) -f coverage.out
```

## Best Practices

### Writing Tests

1. **Use table-driven tests** for multiple scenarios
2. **Add `goleak.VerifyNoLeaks(t)`** for goroutine tests
3. **Mock external dependencies** using interfaces
4. **Test error paths** not just happy paths

### Test Naming

```go
// Good: descriptive function names
func TestCreateOrder_WithInsufficientFunds_ReturnsError(t *testing.T)
func TestVerifyIdentity_ValidScope_Succeeds(t *testing.T)

// Use subtests for related cases
func TestOrderLifecycle(t *testing.T) {
    t.Run("create", func(t *testing.T) { ... })
    t.Run("match", func(t *testing.T) { ... })
    t.Run("close", func(t *testing.T) { ... })
}
```

### Mocking

Mocks are generated with mockery. See `make generate`:

```bash
# Regenerate mocks
make generate

# Mocks are in **/mocks/ directories
```

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `package not found` | Run `go mod tidy` |
| `timeout` | Increase with `-timeout 10m` |
| `build constraints exclude` | Check build tags |
| `mock not found` | Run `make generate` |

### Test Cache

```bash
# Clear test cache
go clean -testcache

# Run without cache
go test -count=1 ./...
```

### Parallel Test Issues

```bash
# Run tests sequentially
go test -p 1 ./x/...

# Limit parallelism within package
go test -parallel 1 ./x/veid/...
```
