# VirtEngine Error Handling Best Practices

This guide provides comprehensive best practices for error handling in VirtEngine.

## Table of Contents

1. [Error Types](#error-types)
2. [Error Codes](#error-codes)
3. [Error Wrapping](#error-wrapping)
4. [Panic Recovery](#panic-recovery)
5. [Error Metrics](#error-metrics)
6. [Best Practices](#best-practices)
7. [Examples](#examples)

## Error Types

VirtEngine provides a rich error type system in `pkg/errors` for consistent error handling across all modules.

### Coded Errors

Coded errors include module name, error code, category, and severity:

```go
import "github.com/virtengine/virtengine/pkg/errors"

err := errors.NewCodedError("veid", 1001, "invalid scope format", errors.CategoryValidation)
```

### Specialized Error Types

Use specialized error types for common scenarios:

#### Validation Errors

```go
err := errors.NewValidationError("veid", 1001, "email", "invalid email format")
```

#### Not Found Errors

```go
err := errors.NewNotFoundError("veid", 1010, "identity", "user123")
```

#### Conflict Errors

```go
err := errors.NewConflictError("veid", 1020, "scope", "scope456", "")
```

#### Unauthorized Errors

```go
err := errors.NewUnauthorizedError("veid", 1030, "delete_scope", "insufficient permissions")
```

#### Timeout Errors

```go
err := errors.NewTimeoutError("inference", 251, "model inference", "30s")
```

#### Internal Errors

```go
err := errors.NewInternalError("veid", 1060, "ml_scorer", "model loading failed")
```

#### External Service Errors

```go
err := errors.NewExternalError("waldur", 650, "openstack", "create_vm", "API unavailable")
```

#### Rate Limit Errors

```go
err := errors.NewRateLimitError("nli", 980, 100, "2024-01-01T00:00:00Z")
```

## Error Codes

### Code Allocation

Each module has a 100-code range (see `pkg/errors/codes.go`):

- **x/ modules (on-chain)**: 1000-3099
  - veid: 1000-1099
  - mfa: 1200-1299
  - encryption: 1300-1399
  - market: 1400-1499
  - etc.

- **pkg/ modules (off-chain)**: 100-999, 3200-4299
  - provider_daemon: 100-199
  - inference: 200-299
  - workflow: 300-399
  - etc.

### Code Categories

Within each module's range, follow these patterns:

- **00-09**: Invalid input/validation errors
- **10-19**: Not found errors
- **20-29**: Conflict/already exists errors
- **30-39**: Unauthorized/permission errors
- **40-49**: State/lifecycle errors
- **50-59**: External service errors
- **60-69**: Internal errors
- **70-79**: Verification/validation errors
- **80-89**: Rate limiting/quota errors
- **90-99**: Reserved for future use

### Validating Codes

```go
if !errors.ValidateCode("veid", 1050) {
    // Code is out of range
}
```

## Error Wrapping

### Basic Wrapping

Always wrap errors with context using `%w`:

```go
if err := doSomething(); err != nil {
    return errors.Wrap(err, "failed to do something")
}
```

### Formatted Wrapping

```go
if err := processUser(userID); err != nil {
    return errors.Wrapf(err, "failed to process user %s", userID)
}
```

### Coded Wrapping

Wrap existing errors with coded errors:

```go
if err := externalAPI.Call(); err != nil {
    return errors.WrapCoded(err, "provider_daemon", 150, "external API call failed", errors.CategoryExternal)
}
```

### Adding Context

Add structured context to errors:

```go
err := someError()
err = errors.WithOperation(err, "database.query")
err = errors.WithResource(err, "user", userID)
err = errors.WithField(err, "email", email)
```

### Unwrapping

Use standard `errors.Unwrap()` or `errors.Is()`:

```go
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found
}

var codedErr *errors.CodedError
if errors.As(err, &codedErr) {
    log.Error("coded error", "module", codedErr.Module, "code", codedErr.Code)
}
```

## Panic Recovery

### Goroutine Safety

**Always** wrap goroutines with panic recovery:

```go
import "github.com/virtengine/virtengine/pkg/errors"

// Simple goroutine with logging
errors.SafeGo("worker-task", func() {
    // ... work that might panic ...
})

// Goroutine with error channel
errCh := make(chan error, 1)
errors.SafeGoWithError("worker-task", errCh, func() error {
    // ... work that might panic or error ...
    return nil
})

// Wait for result
if err := <-errCh; err != nil {
    log.Error("worker failed", "error", err)
}
```

### Defer Recovery

For non-goroutine code that might panic:

```go
func DoWork() (err error) {
    defer func() {
        if recErr := errors.RecoverToError("DoWork"); recErr != nil {
            err = recErr
        }
    }()
    
    // ... work that might panic ...
    return nil
}
```

### Recovery with Cleanup

```go
func ProcessData() {
    defer errors.RecoverWithCleanup("process-data", func() {
        // Always runs, even on panic
        cleanup()
    })
    
    // ... work ...
}
```

### Panic Groups

For multiple goroutines:

```go
pg := errors.NewPanicGroup()

pg.Go("task-1", func() error {
    // ... work ...
    return nil
})

pg.Go("task-2", func() error {
    // ... work ...
    return nil
})

// Wait for all and get first error (if any)
if err := pg.Wait(); err != nil {
    log.Error("panic group failed", "error", err)
}
```

### Custom Panic Handlers

Set a global panic handler:

```go
errors.SetDefaultPanicHandler(func(recovered interface{}, stack []byte) {
    log.Error("PANIC", "recovered", recovered, "stack", string(stack))
    // Send to monitoring service, etc.
})
```

## Error Metrics

### Recording Errors

Errors are automatically recorded when using `errors.As()`:

```go
var codedErr *errors.CodedError
if errors.AsWithMetrics(err, &codedErr) {
    // Error is recorded in Prometheus metrics
}
```

Or record manually:

```go
errors.RecordError(err)
```

### Recording Panics

Panics are automatically recorded when using recovery utilities:

```go
errors.SafeGo("worker", func() {
    panic("something went wrong")  // Automatically recorded
})
```

### Available Metrics

- `virtengine_errors_total{module, code, category, severity}` - Total error count
- `virtengine_panics_recovered_total{context}` - Total panics recovered
- `virtengine_retryable_errors_total{module, category}` - Retryable errors

## Best Practices

### DO

✅ **Use typed errors** for common scenarios (validation, not found, etc.)

✅ **Always wrap errors** with context using `%w`

✅ **Add structured context** to errors using `WithOperation()`, `WithResource()`, etc.

✅ **Use sentinel errors** for well-known error conditions

✅ **Wrap all goroutines** with panic recovery

✅ **Record errors in metrics** for observability

✅ **Return errors, don't log and return** - let the caller decide

✅ **Check error categories** to determine handling (retryable, fatal, etc.)

### DON'T

❌ **Don't use `panic()` in production** code except for truly unrecoverable errors

❌ **Don't ignore errors** - always handle or propagate them

❌ **Don't use string matching** for error checking - use `errors.Is()` and `errors.As()`

❌ **Don't lose error context** - always wrap with additional information

❌ **Don't log and return** the same error multiple times

❌ **Don't create new error types** without adding them to `pkg/errors`

❌ **Don't use arbitrary error codes** - follow the module code allocation

## Examples

### Example 1: Keeper Method with Error Handling

```go
func (k Keeper) CreateOrder(ctx sdk.Context, params OrderParams) (Order, error) {
    // Validate input
    if err := params.Validate(); err != nil {
        return Order{}, errors.NewValidationError(
            "market", 
            1400, 
            "params", 
            "invalid order parameters",
        ).WithContext("error", err.Error())
    }
    
    // Check if order exists
    if k.HasOrder(ctx, params.OrderID) {
        return Order{}, errors.NewConflictError(
            "market",
            1420,
            "order",
            params.OrderID.String(),
            "order already exists",
        )
    }
    
    // Create order
    order, err := k.createOrderInternal(ctx, params)
    if err != nil {
        return Order{}, errors.WrapCoded(
            err,
            "market",
            1460,
            "failed to create order",
            errors.CategoryInternal,
        )
    }
    
    return order, nil
}
```

### Example 2: Provider Daemon with Panic Recovery

```go
func (d *Daemon) Start(ctx context.Context) error {
    // Bid engine goroutine
    errors.SafeGo("bid-engine", func() {
        d.bidEngine.Run(ctx)
    })
    
    // Usage meter goroutine
    errors.SafeGo("usage-meter", func() {
        d.usageMeter.Run(ctx)
    })
    
    // Event processor with error handling
    errCh := make(chan error, 1)
    errors.SafeGoWithError("event-processor", errCh, func() error {
        return d.eventProcessor.Run(ctx)
    })
    
    // Wait for error or context cancellation
    select {
    case err := <-errCh:
        return errors.Wrap(err, "event processor failed")
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### Example 3: External Service Call with Retries

```go
func (c *WaldurClient) CreateVM(ctx context.Context, spec VMSpec) (*VM, error) {
    var lastErr error
    
    for i := 0; i < 3; i++ {
        vm, err := c.createVMAttempt(ctx, spec)
        if err == nil {
            return vm, nil
        }
        
        lastErr = err
        
        // Check if retryable
        if !errors.IsRetryable(err) {
            break
        }
        
        // Exponential backoff
        time.Sleep(time.Duration(1<<i) * time.Second)
    }
    
    return nil, errors.WrapCoded(
        lastErr,
        "waldur",
        650,
        "failed to create VM after retries",
        errors.CategoryExternal,
    )
}

func (c *WaldurClient) createVMAttempt(ctx context.Context, spec VMSpec) (*VM, error) {
    resp, err := c.client.Post(ctx, "/vms", spec)
    if err != nil {
        return nil, errors.NewExternalError(
            "waldur",
            651,
            "waldur_api",
            "create_vm",
            err.Error(),
        )
    }
    
    if resp.StatusCode >= 500 {
        return nil, errors.NewExternalError(
            "waldur",
            652,
            "waldur_api",
            "create_vm",
            fmt.Sprintf("server error: %d", resp.StatusCode),
        )
    }
    
    if resp.StatusCode == 429 {
        return nil, errors.NewRateLimitError(
            "waldur",
            680,
            100,
            resp.Header.Get("X-RateLimit-Reset"),
        )
    }
    
    var vm VM
    if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
        return nil, errors.Wrap(err, "failed to decode response")
    }
    
    return &vm, nil
}
```

### Example 4: ML Inference with Timeout

```go
func (s *Scorer) ScoreIdentity(ctx context.Context, features []float64) (float64, error) {
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Channel for result
    resultCh := make(chan float64, 1)
    errCh := make(chan error, 1)
    
    // Run inference with panic recovery
    errors.SafeGoWithError("ml-inference", errCh, func() error {
        score, err := s.model.Predict(features)
        if err != nil {
            return err
        }
        resultCh <- score
        return nil
    })
    
    // Wait for result or timeout
    select {
    case score := <-resultCh:
        return score, nil
    case err := <-errCh:
        return 0, errors.WrapCoded(
            err,
            "inference",
            261,
            "ML inference failed",
            errors.CategoryInternal,
        )
    case <-ctx.Done():
        return 0, errors.NewTimeoutError(
            "inference",
            251,
            "ML inference",
            "30s",
        )
    }
}
```

## Migration Guide

See [docs/errors/MIGRATION.md](../docs/errors/MIGRATION.md) for guidance on migrating existing code to the new error handling system.

## References

- Error Code Registry: `pkg/errors/codes.go`
- Error Types: `pkg/errors/types.go`
- Error Wrapping: `pkg/errors/wrap.go`
- Sentinel Errors: `pkg/errors/sentinel.go`
- Panic Recovery: `pkg/errors/recovery.go`
- Error Metrics: `pkg/errors/metrics.go`
