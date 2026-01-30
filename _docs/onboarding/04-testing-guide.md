# Testing Strategy and Guidelines

This guide covers VirtEngine's testing strategy, how to write tests, and best practices.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Tiers](#test-tiers)
3. [Running Tests](#running-tests)
4. [Writing Tests](#writing-tests)
5. [Mocking](#mocking)
6. [Coverage](#coverage)
7. [Debugging Tests](#debugging-tests)
8. [Best Practices](#best-practices)

---

## Testing Philosophy

VirtEngine follows these testing principles:

| Principle | Description |
|-----------|-------------|
| **Test behavior, not implementation** | Tests should verify what code does, not how |
| **Pyramid structure** | Many unit tests, fewer integration, few E2E |
| **Deterministic** | Tests must produce consistent results |
| **Fast feedback** | Unit tests should complete in seconds |
| **Independent** | Tests should not depend on each other |

### Test Pyramid

```
            ▲
           /│\        E2E Tests
          / │ \       (Few, slow, high confidence)
         /  │  \
        /───┼───\     Integration Tests
       /    │    \    (Medium, cross-module)
      /     │     \
     /──────┼──────\  Unit Tests
    /       │       \ (Many, fast, focused)
   ─────────┴─────────
```

---

## Test Tiers

### Tier 1: Unit Tests

**Location**: `x/*/types/`, `x/*/keeper/`, `pkg/*/`

**Purpose**: Test individual functions and types in isolation.

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

**Run**:
```bash
go test ./x/... ./pkg/...
```

### Tier 2: Integration Tests

**Location**: `tests/integration/`

**Purpose**: Test interactions between modules.

```
tests/integration/
├── escrow_market_test.go   # Escrow + Market integration
├── veid_mfa_test.go        # VEID + MFA integration
└── ...
```

**Run**:
```bash
go test -tags="e2e.integration" ./tests/integration/...
```

### Tier 3: E2E Tests

**Location**: `tests/e2e/`

**Purpose**: Test CLI and gRPC endpoints against a running chain.

```
tests/e2e/
├── market_cli_test.go      # Market CLI commands
├── market_grpc_test.go     # Market gRPC queries
├── veid_cli_test.go        # VEID CLI commands
└── ...
```

**Run**:
```bash
make test-integration
```

---

## Running Tests

### Quick Commands

```bash
# All unit tests
go test ./x/... ./pkg/...

# Specific module
go test ./x/veid/...
go test ./x/market/...

# Specific package
go test ./pkg/provider_daemon/...

# With verbose output
go test -v ./x/veid/...

# Without cache
go test -count=1 ./x/...

# With timeout
go test -timeout 5m ./x/...
```

### Integration Tests

```bash
# Run integration tests (requires localnet)
./scripts/localnet.sh start
go test -tags="e2e.integration" ./tests/integration/...

# Using make
make test-integration
```

### Coverage

```bash
# Basic coverage
go test -cover ./x/... ./pkg/...

# With HTML report
go test -coverprofile=coverage.out ./x/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html

# View coverage by function
go tool cover -func=coverage.out
```

### Test Specific Functions

```bash
# Single test by name
go test -v -run TestCreateOrder ./x/market/keeper/...

# Pattern matching
go test -v -run "TestOrder.*" ./x/market/...

# Subtests
go test -v -run "TestOrder/with_escrow" ./x/market/...
```

---

## Writing Tests

### Unit Test Structure

```go
package keeper_test

import (
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/virtengine/virtengine/x/market/keeper"
    "github.com/virtengine/virtengine/x/market/types"
)

func TestCreateOrder(t *testing.T) {
    // Setup
    ctx, k := setupKeeper(t)
    
    // Execute
    order, err := k.CreateOrder(ctx, &types.MsgCreateOrder{
        Customer:   "virtengine1customer...",
        OfferingID: "offering_123",
        Quantity:   1,
    })
    
    // Assert
    require.NoError(t, err)
    require.NotNil(t, order)
    require.Equal(t, "offering_123", order.OfferingID)
}
```

### Table-Driven Tests

```go
func TestOrderValidation(t *testing.T) {
    tests := []struct {
        name    string
        order   types.MsgCreateOrder
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid order",
            order: types.MsgCreateOrder{
                Customer:   "virtengine1...",
                OfferingID: "offering_123",
                Quantity:   1,
            },
            wantErr: false,
        },
        {
            name: "invalid quantity",
            order: types.MsgCreateOrder{
                Customer:   "virtengine1...",
                OfferingID: "offering_123",
                Quantity:   0,
            },
            wantErr: true,
            errMsg:  "quantity must be positive",
        },
        {
            name: "missing offering",
            order: types.MsgCreateOrder{
                Customer:   "virtengine1...",
                OfferingID: "",
                Quantity:   1,
            },
            wantErr: true,
            errMsg:  "offering ID required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.order.ValidateBasic()
            if tt.wantErr {
                require.Error(t, err)
                require.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Testing Keepers

```go
func setupKeeper(t *testing.T) (sdk.Context, keeper.Keeper) {
    t.Helper()
    
    // Create store
    storeKey := storetypes.NewKVStoreKey(types.StoreKey)
    
    // Create context
    db := dbm.NewMemDB()
    stateStore := store.NewCommitMultiStore(db)
    stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
    require.NoError(t, stateStore.LoadLatestVersion())
    
    ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
    
    // Create keeper
    k := keeper.NewKeeper(
        codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
        storeKey,
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    )
    
    return ctx, k
}
```

### Testing Message Handlers

```go
func TestMsgServer_CreateOrder(t *testing.T) {
    ctx, k := setupKeeper(t)
    msgServer := keeper.NewMsgServerImpl(k)
    
    // Create valid message
    msg := &types.MsgCreateOrder{
        Customer:   "virtengine1customer...",
        OfferingID: "offering_123",
        Quantity:   1,
    }
    
    // Execute
    resp, err := msgServer.CreateOrder(sdk.WrapSDKContext(ctx), msg)
    
    // Verify
    require.NoError(t, err)
    require.NotEmpty(t, resp.OrderId)
    
    // Verify state was updated
    order, found := k.GetOrder(ctx, resp.OrderId)
    require.True(t, found)
    require.Equal(t, types.OrderStateOpen, order.State)
}
```

### Testing Genesis

```go
func TestGenesisState_Validate(t *testing.T) {
    tests := []struct {
        name     string
        genesis  *types.GenesisState
        wantErr  bool
    }{
        {
            name:    "default genesis is valid",
            genesis: types.DefaultGenesis(),
            wantErr: false,
        },
        {
            name: "duplicate order",
            genesis: &types.GenesisState{
                Orders: []types.Order{
                    {Id: "order_1"},
                    {Id: "order_1"}, // duplicate
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.genesis.Validate()
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Mocking

### Generated Mocks

Mocks are generated using mockery:

```bash
# Regenerate all mocks
make generate
```

Mocks are placed in `**/mocks/` directories.

### Using Mocks

```go
func TestBidEngine_WithMockedKeeper(t *testing.T) {
    // Create mock
    mockKeeper := mocks.NewIKeeper(t)
    
    // Setup expectations
    mockKeeper.EXPECT().
        GetOrder(mock.Anything, "order_123").
        Return(types.Order{ID: "order_123"}, true)
    
    mockKeeper.EXPECT().
        CreateBid(mock.Anything, mock.Anything).
        Return(types.Bid{ID: "bid_456"}, nil)
    
    // Create engine with mock
    engine := NewBidEngine(mockKeeper)
    
    // Test
    bid, err := engine.ProcessOrder(ctx, "order_123")
    require.NoError(t, err)
    require.Equal(t, "bid_456", bid.ID)
}
```

### Manual Mocks

For simple interfaces, manual mocks may be clearer:

```go
type mockOrderStore struct {
    orders map[string]types.Order
}

func (m *mockOrderStore) GetOrder(ctx sdk.Context, id string) (types.Order, bool) {
    order, ok := m.orders[id]
    return order, ok
}

func (m *mockOrderStore) SetOrder(ctx sdk.Context, order types.Order) error {
    m.orders[order.ID] = order
    return nil
}
```

---

## Coverage

### Coverage Targets

| Category | Target |
|----------|--------|
| Overall | 80%+ |
| Keeper methods | 90%+ |
| Type validation | 100% |
| Genesis | 100% |
| Error paths | 80%+ |

### Priority Modules

Focus coverage on critical paths:

1. **x/veid** - Identity verification (security critical)
2. **x/escrow** - Payment handling (financial)
3. **x/market** - Order lifecycle (core business)
4. **x/mfa** - Authentication (security)
5. **pkg/inference** - ML scoring (consensus critical)

### Checking Coverage

```bash
# Generate report
go test -coverprofile=coverage.out ./x/...

# View by function
go tool cover -func=coverage.out | grep -E "total:|veid|market"

# Find uncovered code
go tool cover -func=coverage.out | grep "0.0%"
```

---

## Debugging Tests

### Verbose Output

```bash
# Show test names and timing
go test -v ./x/veid/...

# Very verbose with all logs
go test -v -count=1 ./x/veid/keeper/...
```

### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a specific test
dlv test ./x/veid/keeper/... -- -test.run TestVerifyIdentity

# Common commands:
# b <func>    - set breakpoint
# c           - continue
# n           - next line
# s           - step into
# p <var>     - print variable
# q           - quit
```

### Environment Variables

```bash
# Enable verbose SDK logging
export COSMOS_SDK_LOG_LEVEL=debug
go test -v ./x/...

# Trace goroutines
export GOTRACEBACK=all
go test -v ./x/...
```

### Test Timeouts

```bash
# Increase timeout for slow tests
go test -timeout 10m ./tests/integration/...

# Default timeout is 10 minutes
```

---

## Best Practices

### Do's

✅ **Use table-driven tests** for multiple scenarios

```go
tests := []struct{
    name string
    input string
    want string
}{...}
```

✅ **Add goleak for goroutine tests**

```go
func TestConcurrentOperation(t *testing.T) {
    defer goleak.VerifyNoLeaks(t)
    // test code
}
```

✅ **Test error paths**, not just happy paths

```go
func TestCreateOrder_InvalidInput(t *testing.T) {
    _, err := k.CreateOrder(ctx, nil)
    require.Error(t, err)
}
```

✅ **Use descriptive test names**

```go
func TestCreateOrder_WithInsufficientFunds_ReturnsError(t *testing.T)
func TestVerifyIdentity_ValidScope_Succeeds(t *testing.T)
```

✅ **Use subtests for related cases**

```go
func TestOrderLifecycle(t *testing.T) {
    t.Run("create", func(t *testing.T) { ... })
    t.Run("match", func(t *testing.T) { ... })
    t.Run("close", func(t *testing.T) { ... })
}
```

### Don'ts

❌ **Don't test implementation details**

```go
// Bad: tests internal structure
require.Equal(t, "expected_internal_state", k.internalMap["key"])

// Good: tests behavior
order, found := k.GetOrder(ctx, orderID)
require.True(t, found)
```

❌ **Don't use sleep for synchronization**

```go
// Bad
go doAsync()
time.Sleep(100 * time.Millisecond)
checkResult()

// Good
done := make(chan struct{})
go func() {
    doAsync()
    close(done)
}()
<-done
checkResult()
```

❌ **Don't share state between tests**

```go
// Bad: modifies package-level state
var globalCounter int

// Good: isolated state per test
func TestFoo(t *testing.T) {
    counter := 0
    // ...
}
```

❌ **Don't ignore errors**

```go
// Bad
result, _ := k.GetOrder(ctx, id)

// Good
result, err := k.GetOrder(ctx, id)
require.NoError(t, err)
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `package not found` | Run `go mod tidy` |
| `timeout` | Increase with `-timeout 10m` |
| `build constraints exclude` | Check build tags |
| `mock not found` | Run `make generate` |
| Race detected | Use `-race` flag to debug |

### Test Cache

```bash
# Clear test cache
go clean -testcache

# Run without cache
go test -count=1 ./...
```

### Parallel Issues

```bash
# Run tests sequentially
go test -p 1 ./x/...

# Limit parallelism within package
go test -parallel 1 ./x/veid/...
```

---

## Related Documentation

- [Testing Guide Reference](../testing-guide.md) - Full testing documentation
- [Code Contribution](./03-code-contribution.md) - PR requirements
- [Debugging Guide](./08-debugging-guide.md) - Advanced debugging
