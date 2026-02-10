# Common Patterns and Anti-patterns

This guide documents common patterns to follow and anti-patterns to avoid in VirtEngine development.

## Table of Contents

1. [Go Patterns](#go-patterns)
2. [Cosmos SDK Patterns](#cosmos-sdk-patterns)
3. [Concurrency Patterns](#concurrency-patterns)
4. [Error Handling Patterns](#error-handling-patterns)
5. [Testing Patterns](#testing-patterns)
6. [Security Patterns](#security-patterns)
7. [Performance Patterns](#performance-patterns)
8. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Go Patterns

### Interface Segregation

✅ **Do**: Define small, focused interfaces

```go
// Good - small, focused interfaces
type OrderReader interface {
    GetOrder(ctx sdk.Context, id string) (Order, bool)
}

type OrderWriter interface {
    SetOrder(ctx sdk.Context, order Order) error
    DeleteOrder(ctx sdk.Context, id string) error
}

type OrderStore interface {
    OrderReader
    OrderWriter
}
```

❌ **Don't**: Create large, monolithic interfaces

```go
// Bad - monolithic interface with unrelated methods
type Everything interface {
    GetOrder(ctx sdk.Context, id string) (Order, bool)
    SetOrder(ctx sdk.Context, order Order) error
    GetUser(ctx sdk.Context, addr string) (User, bool)
    SendNotification(ctx sdk.Context, msg string) error
    // ... 50 more methods
}
```

### Functional Options

✅ **Do**: Use functional options for complex configuration

```go
// Good - flexible, backwards-compatible configuration
type ServerOption func(*Server)

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = d
    }
}

func WithLogger(l Logger) ServerOption {
    return func(s *Server) {
        s.logger = l
    }
}

func NewServer(opts ...ServerOption) *Server {
    s := &Server{
        timeout: defaultTimeout,
        logger:  defaultLogger,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(
    WithTimeout(30 * time.Second),
    WithLogger(myLogger),
)
```

### Constructor Pattern

✅ **Do**: Use constructors to enforce invariants

```go
// Good - constructor validates and initializes
func NewOrder(customer string, quantity int) (*Order, error) {
    if quantity <= 0 {
        return nil, ErrInvalidQuantity
    }
    addr, err := sdk.AccAddressFromBech32(customer)
    if err != nil {
        return nil, ErrInvalidAddress
    }
    return &Order{
        Customer: addr,
        Quantity: quantity,
        State:    OrderStateOpen,
    }, nil
}
```

❌ **Don't**: Allow invalid states

```go
// Bad - anyone can create invalid orders
order := Order{
    Quantity: -5,  // Invalid!
    State:    "",  // Invalid!
}
```

---

## Cosmos SDK Patterns

### Keeper Interface Pattern

✅ **Do**: Define interface before concrete Keeper

```go
// Good - interface first
type IKeeper interface {
    CreateOrder(ctx sdk.Context, msg *MsgCreateOrder) (*Order, error)
    GetOrder(ctx sdk.Context, id string) (Order, bool)
    WithOrders(ctx sdk.Context, fn func(Order) bool)
}

type Keeper struct {
    cdc       codec.BinaryCodec
    storeKey  storetypes.StoreKey
    authority string
}

var _ IKeeper = Keeper{}  // Compile-time check
```

### Authority Pattern

✅ **Do**: Use x/gov module account for MsgUpdateParams

```go
// Good - governance controls params
func NewKeeper(
    cdc codec.BinaryCodec,
    storeKey storetypes.StoreKey,
    authority string,  // Always x/gov module address
) Keeper {
    return Keeper{
        authority: authority,
    }
}

func (m msgServer) UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
    if m.GetAuthority() != msg.Authority {
        return nil, ErrUnauthorized.Wrapf("expected %s, got %s", m.GetAuthority(), msg.Authority)
    }
    // Update params
}
```

❌ **Don't**: Hardcode addresses

```go
// Bad - hardcoded address
const adminAddress = "virtengine1admin..."

func (k Keeper) UpdateParams(ctx sdk.Context, sender string, params Params) error {
    if sender != adminAddress {  // Bad!
        return ErrUnauthorized
    }
}
```

### Event Emission Pattern

✅ **Do**: Emit events for significant state changes

```go
// Good - emit events with relevant attributes
func (m msgServer) CreateOrder(ctx context.Context, msg *MsgCreateOrder) (*MsgCreateOrderResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    
    order, err := m.Keeper.CreateOrder(sdkCtx, msg)
    if err != nil {
        return nil, err
    }
    
    sdkCtx.EventManager().EmitEvent(
        sdk.NewEvent(
            EventTypeOrderCreated,
            sdk.NewAttribute(AttributeKeyOrderID, order.ID),
            sdk.NewAttribute(AttributeKeyCustomer, order.Customer),
            sdk.NewAttribute(AttributeKeyAmount, order.Amount.String()),
        ),
    )
    
    return &MsgCreateOrderResponse{OrderId: order.ID}, nil
}
```

### Store Iteration Pattern

✅ **Do**: Use bounded iteration with pagination

```go
// Good - bounded iteration
func (k Keeper) GetOrders(ctx sdk.Context, pagination *query.PageRequest) ([]Order, *query.PageResponse, error) {
    store := ctx.KVStore(k.storeKey)
    orderStore := prefix.NewStore(store, OrderKeyPrefix)
    
    var orders []Order
    pageRes, err := query.Paginate(orderStore, pagination, func(key, value []byte) error {
        var order Order
        if err := k.cdc.Unmarshal(value, &order); err != nil {
            return err
        }
        orders = append(orders, order)
        return nil
    })
    
    return orders, pageRes, err
}
```

❌ **Don't**: Unbounded iteration

```go
// Bad - unbounded iteration can be DoS vector
func (k Keeper) GetAllOrders(ctx sdk.Context) []Order {
    var orders []Order
    k.WithOrders(ctx, func(order Order) bool {
        orders = append(orders, order)  // Could be millions!
        return false
    })
    return orders
}
```

---

## Concurrency Patterns

### Context Cancellation

✅ **Do**: Respect context cancellation

```go
// Good - check context in long operations
func (k Keeper) ProcessBatch(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := k.processItem(ctx, item); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### Goroutine Lifecycle

✅ **Do**: Ensure goroutines can be stopped

```go
// Good - controllable goroutine
type Worker struct {
    done chan struct{}
}

func (w *Worker) Start(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case <-w.done:
                return
            case work := <-w.workChan:
                w.process(work)
            }
        }
    }()
}

func (w *Worker) Stop() {
    close(w.done)
}
```

❌ **Don't**: Create uncontrollable goroutines

```go
// Bad - goroutine runs forever
func startWorker() {
    go func() {
        for {
            doWork()  // No way to stop!
        }
    }()
}
```

### Channel Patterns

✅ **Do**: Use proper channel idioms

```go
// Good - buffered channel with proper closing
func produce(ctx context.Context) <-chan Item {
    out := make(chan Item, 10)
    go func() {
        defer close(out)  // Always close when done
        for {
            select {
            case <-ctx.Done():
                return
            case out <- createItem():
            }
        }
    }()
    return out
}
```

---

## Error Handling Patterns

### Error Wrapping

✅ **Do**: Wrap errors with context

```go
// Good - error wrapping with context
func (k Keeper) CreateOrder(ctx sdk.Context, msg *MsgCreateOrder) (*Order, error) {
    customer, err := sdk.AccAddressFromBech32(msg.Customer)
    if err != nil {
        return nil, fmt.Errorf("invalid customer address %s: %w", msg.Customer, err)
    }
    
    if err := k.validateOrder(ctx, msg); err != nil {
        return nil, fmt.Errorf("order validation failed: %w", err)
    }
    
    return &Order{}, nil
}
```

### Sentinel Errors

✅ **Do**: Use sentinel errors for expected conditions

```go
// Good - sentinel errors
var (
    ErrOrderNotFound   = errors.New("order not found")
    ErrInvalidQuantity = errors.New("invalid quantity")
    ErrUnauthorized    = errors.New("unauthorized")
)

// Check with errors.Is
if errors.Is(err, ErrOrderNotFound) {
    // Handle missing order
}
```

### Error Types

✅ **Do**: Use error types for rich error information

```go
// Good - error type with context
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// Check with errors.As
var valErr *ValidationError
if errors.As(err, &valErr) {
    log.Printf("Invalid field: %s", valErr.Field)
}
```

---

## Testing Patterns

### Table-Driven Tests

✅ **Do**: Use table-driven tests

```go
// Good - table-driven test
func TestValidateOrder(t *testing.T) {
    tests := []struct {
        name    string
        order   Order
        wantErr error
    }{
        {
            name:    "valid order",
            order:   Order{Quantity: 1, Customer: validAddr},
            wantErr: nil,
        },
        {
            name:    "zero quantity",
            order:   Order{Quantity: 0, Customer: validAddr},
            wantErr: ErrInvalidQuantity,
        },
        {
            name:    "negative quantity",
            order:   Order{Quantity: -1, Customer: validAddr},
            wantErr: ErrInvalidQuantity,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.order.Validate()
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Test Helpers

✅ **Do**: Create reusable test helpers

```go
// Good - test helper with t.Helper()
func setupTestKeeper(t *testing.T) (sdk.Context, Keeper) {
    t.Helper()  // Marks this as a helper function
    
    db := dbm.NewMemDB()
    storeKey := storetypes.NewKVStoreKey(types.StoreKey)
    stateStore := store.NewCommitMultiStore(db)
    stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
    require.NoError(t, stateStore.LoadLatestVersion())
    
    ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
    k := NewKeeper(codec, storeKey, authority)
    
    return ctx, k
}
```

### Goroutine Leak Detection

✅ **Do**: Check for goroutine leaks

```go
// Good - detect goroutine leaks
func TestConcurrentOperations(t *testing.T) {
    defer goleak.VerifyNoLeaks(t)
    
    // Test concurrent operations
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Do work
        }()
    }
    wg.Wait()
}
```

---

## Security Patterns

### Input Validation

✅ **Do**: Validate all external input

```go
// Good - comprehensive validation
func (msg MsgCreateOrder) ValidateBasic() error {
    // Validate address
    if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
        return ErrInvalidAddress.Wrap(err.Error())
    }
    
    // Validate quantity
    if msg.Quantity <= 0 {
        return ErrInvalidQuantity.Wrap("quantity must be positive")
    }
    if msg.Quantity > MaxOrderQuantity {
        return ErrInvalidQuantity.Wrapf("quantity exceeds max %d", MaxOrderQuantity)
    }
    
    // Validate offering ID format
    if !isValidOfferingID(msg.OfferingID) {
        return ErrInvalidInput.Wrap("invalid offering ID format")
    }
    
    return nil
}
```

### Sensitive Data Handling

✅ **Do**: Protect sensitive data

```go
// Good - use encryption for sensitive data
type SecureConfig struct {
    EncryptedAPIKey []byte
}

func (c *SecureConfig) DecryptAPIKey(privateKey []byte) ([]byte, error) {
    return crypto.Decrypt(c.EncryptedAPIKey, privateKey)
}

// Good - clear sensitive data from memory
func processSensitive(key []byte) error {
    defer func() {
        for i := range key {
            key[i] = 0
        }
    }()
    // Use key
    return nil
}
```

### Logging Safety

✅ **Do**: Redact sensitive information

```go
// Good - redact sensitive fields
logger := observability.NewLogger(observability.Config{
    RedactPatterns: []string{
        "password", "secret", "key", "token", "mnemonic",
    },
})

// Logs: "Processing request user=alice api_key=[REDACTED]"
logger.Info("Processing request",
    "user", userID,
    "api_key", apiKey,  // Will be redacted
)
```

---

## Performance Patterns

### Lazy Initialization

✅ **Do**: Initialize expensive resources lazily

```go
// Good - lazy initialization
type Service struct {
    client     *http.Client
    clientOnce sync.Once
}

func (s *Service) getClient() *http.Client {
    s.clientOnce.Do(func() {
        s.client = &http.Client{
            Timeout: 30 * time.Second,
        }
    })
    return s.client
}
```

### Pooling

✅ **Do**: Pool expensive resources

```go
// Good - buffer pool
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data []byte) error {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    buf.Write(data)
    // Process
    return nil
}
```

### Batch Operations

✅ **Do**: Batch database operations

```go
// Good - batch writes
func (k Keeper) SetOrders(ctx sdk.Context, orders []Order) error {
    store := ctx.KVStore(k.storeKey)
    
    for _, order := range orders {
        bz, err := k.cdc.Marshal(&order)
        if err != nil {
            return err
        }
        store.Set(OrderKey(order.ID), bz)
    }
    
    return nil
}
```

---

## Anti-Patterns to Avoid

### Non-Deterministic Operations in Consensus

❌ **Never** use non-deterministic operations in consensus code:

```go
// BAD - map iteration is random
for k, v := range myMap {
    process(k, v)  // Non-deterministic order!
}

// BAD - current time
timestamp := time.Now()  // Varies between validators!

// BAD - random numbers
n := rand.Intn(100)  // Different on each validator!

// BAD - floating point
result := 1.0 / 3.0  // May differ across platforms!
```

✅ **Do**: Use deterministic alternatives

```go
// GOOD - sorted keys
keys := make([]string, 0, len(myMap))
for k := range myMap {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    process(k, myMap[k])
}

// GOOD - block time from context
timestamp := ctx.BlockTime()

// GOOD - fixed seed for any randomness
rng := rand.New(rand.NewSource(42))

// GOOD - integer or sdk.Dec for precision
result := sdk.NewDec(1).Quo(sdk.NewDec(3))
```

### Panic in Business Logic

❌ **Don't** panic in normal code paths:

```go
// Bad - panic on invalid input
func getOrder(id string) Order {
    if id == "" {
        panic("invalid order ID")  // Don't panic!
    }
}
```

✅ **Do**: Return errors

```go
// Good - return error
func getOrder(id string) (Order, error) {
    if id == "" {
        return Order{}, ErrInvalidOrderID
    }
    return Order{}, nil
}
```

### Ignoring Errors

❌ **Don't** ignore errors:

```go
// Bad - ignored error
result, _ := riskyOperation()

// Bad - empty error handler
if err != nil {
    // nothing
}
```

✅ **Do**: Handle all errors explicitly

```go
// Good - handle error
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("risky operation failed: %w", err)
}

// Good - explicit ignore with reason
_ = closeResource() // Ignore close error in defer
```

### God Objects

❌ **Don't** create god objects:

```go
// Bad - does too many things
type SuperManager struct {
    // Handles users, orders, payments, notifications, reporting...
}
```

✅ **Do**: Single responsibility

```go
// Good - focused responsibilities
type OrderManager struct { /* just orders */ }
type PaymentProcessor struct { /* just payments */ }
type NotificationService struct { /* just notifications */ }
```

---

## Related Documentation

- [Module Development](./06-module-development.md) - Building modules
- [Code Review Checklist](./05-code-review-checklist.md) - Review standards
- [Testing Guide](./04-testing-guide.md) - Testing practices
- [Debugging Guide](./08-debugging-guide.md) - Troubleshooting
