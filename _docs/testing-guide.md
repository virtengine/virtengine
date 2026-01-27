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

### Coverage

Generate coverage reports:

```bash
# Basic coverage
go test -cover ./x/... ./pkg/...

# Coverage with HTML report
go test -coverprofile=coverage.out ./x/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific module
go test -cover ./x/veid/...
```

### Quick Test Commands

```bash
# Fast test run (no cache)
go test -count=1 ./x/...

# With timeout
go test -timeout 5m ./x/...

# Summary output
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
