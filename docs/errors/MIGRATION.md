# Error Handling Migration Guide

This guide helps developers migrate existing code to use the new standardized error handling system.

## Overview

The new error handling system provides:

- Standardized error codes across all modules
- Rich error types with categories and metadata
- Error wrapping with context
- Panic recovery utilities
- Error metrics for observability

## Migration Steps

### Step 1: Update Imports

Add the new error package import:

```go
import "github.com/virtengine/virtengine/pkg/errors"
```

For blockchain modules (x/), continue using:

```go
import errorsmod "cosmossdk.io/errors"
```

### Step 2: Replace Error Creation

#### Before (Standard Go Errors)

```go
func DoSomething() error {
    return errors.New("something failed")
}
```

#### After (Coded Errors)

```go
func DoSomething() error {
    return errors.NewCodedError("mymodule", 100, "something failed", errors.CategoryInternal)
}
```

Or use specialized types:

```go
// Validation error
return errors.NewValidationError("mymodule", 100, "field", "invalid value")

// Not found error
return errors.NewNotFoundError("mymodule", 110, "resource", resourceID)
```

### Step 3: Update Error Wrapping

#### Before (fmt.Errorf without %w)

```go
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %v", err)
}
```

#### After (Proper Error Wrapping)

```go
if err := doSomething(); err != nil {
    return errors.Wrap(err, "failed to do something")
}
```

Or with formatting:

```go
if err := processUser(userID); err != nil {
    return errors.Wrapf(err, "failed to process user %s", userID)
}
```

### Step 4: Add Context to Errors

#### Before

```go
return fmt.Errorf("failed to create order: %v", err)
```

#### After

```go
err = errors.WithOperation(err, "create_order")
err = errors.WithResource(err, "order", orderID)
return err
```

### Step 5: Add Panic Recovery to Goroutines

#### Before

```go
go func() {
    // ... work that might panic ...
}()
```

#### After

```go
errors.SafeGo("worker-task", func() {
    // ... work that might panic ...
})
```

Or with error handling:

```go
errCh := make(chan error, 1)
errors.SafeGoWithError("worker-task", errCh, func() error {
    // ... work ...
    return nil
})

if err := <-errCh; err != nil {
    log.Error("worker failed", "error", err)
}
```

### Step 6: Update Error Checking

#### Before (String Matching - Bad!)

```go
if err != nil && strings.Contains(err.Error(), "not found") {
    // handle not found
}
```

#### After (Sentinel Errors or Types)

```go
if errors.Is(err, errors.ErrNotFound) {
    // handle not found
}

// Or check type
var nfErr *errors.NotFoundError
if errors.As(err, &nfErr) {
    // handle not found
    log.Info("resource not found", "type", nfErr.ResourceType, "id", nfErr.ResourceID)
}
```

## Module-Specific Migration

### Blockchain Modules (x/)

Blockchain modules should continue using `errorsmod.Register()` but follow the new code allocation:

#### Before

```go
var (
    ErrInvalidScope = errorsmod.Register(ModuleName, 1, "invalid scope")
    ErrScopeNotFound = errorsmod.Register(ModuleName, 2, "scope not found")
)
```

#### After

```go
var (
    // Follow code patterns: 00-09 for validation, 10-19 for not found
    ErrInvalidScope = errorsmod.Register(ModuleName, 1001, "invalid scope")
    ErrScopeNotFound = errorsmod.Register(ModuleName, 1010, "scope not found")
)
```

### Off-Chain Services (pkg/)

Off-chain services should use the new error types:

#### Before

```go
package provider_daemon

func (d *Daemon) Start() error {
    if d.config == nil {
        return errors.New("configuration not set")
    }
    // ...
}
```

#### After

```go
package provider_daemon

import "github.com/virtengine/virtengine/pkg/errors"

func (d *Daemon) Start() error {
    if d.config == nil {
        return errors.NewValidationError("provider_daemon", 100, "config", "configuration not set")
    }
    // ...
}
```

## Common Migration Patterns

### Pattern 1: Simple Error to Coded Error

```go
// Before
return errors.New("invalid input")

// After
return errors.NewValidationError("module", 100, "input", "invalid input")
```

### Pattern 2: fmt.Errorf to Wrapped Error

```go
// Before
return fmt.Errorf("failed to process: %v", err)

// After
return errors.Wrap(err, "failed to process")
```

### Pattern 3: Custom Error Type to Standard Type

```go
// Before
type MyNotFoundError struct {
    ResourceID string
}

func (e *MyNotFoundError) Error() string {
    return fmt.Sprintf("resource %s not found", e.ResourceID)
}

// After - Use standard NotFoundError
return errors.NewNotFoundError("module", 110, "resource", resourceID)
```

### Pattern 4: Panic in Goroutine

```go
// Before
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("panic", "recovered", r)
        }
    }()
    // work
}()

// After
errors.SafeGo("task-name", func() {
    // work
})
```

### Pattern 5: Error with Retry Logic

```go
// Before
for i := 0; i < 3; i++ {
    if err := doWork(); err != nil {
        if i < 2 {
            time.Sleep(time.Second)
            continue
        }
        return err
    }
    return nil
}

// After - Use error types to determine retryability
for i := 0; i < 3; i++ {
    err := doWork()
    if err == nil {
        return nil
    }
    
    if !errors.IsRetryable(err) {
        return err
    }
    
    if i < 2 {
        time.Sleep(time.Duration(1<<i) * time.Second)
    }
}
return errors.Wrap(lastErr, "max retries exceeded")
```

## Module-by-Module Checklist

Use this checklist when migrating a module:

- [ ] Identify all error creation points
- [ ] Assign error codes following the module's range
- [ ] Replace `errors.New()` with typed errors
- [ ] Replace `fmt.Errorf()` with `errors.Wrap()` or `errors.Wrapf()`
- [ ] Add context to errors (operation, resource, field)
- [ ] Wrap all goroutines with panic recovery
- [ ] Replace string matching with `errors.Is()` and `errors.As()`
- [ ] Update tests to check error types
- [ ] Add error code validation tests
- [ ] Document new error codes in ERROR_CATALOG.md
- [ ] Update module README with error handling info

## Testing

### Test Error Types

```go
func TestMyFunction(t *testing.T) {
    err := MyFunction()
    
    // Check error type
    var valErr *errors.ValidationError
    if !errors.As(err, &valErr) {
        t.Errorf("expected ValidationError, got %T", err)
    }
    
    // Check error code
    module, code := errors.GetCode(err)
    if module != "mymodule" || code != 100 {
        t.Errorf("unexpected error code: %s:%d", module, code)
    }
    
    // Check error category
    if errors.GetCategory(err) != errors.CategoryValidation {
        t.Error("wrong error category")
    }
}
```

### Test Error Wrapping

```go
func TestErrorWrapping(t *testing.T) {
    originalErr := errors.ErrNotFound
    wrappedErr := errors.Wrap(originalErr, "additional context")
    
    // Verify unwrapping works
    if !errors.Is(wrappedErr, originalErr) {
        t.Error("wrapped error should unwrap to original")
    }
}
```

### Test Panic Recovery

```go
func TestPanicRecovery(t *testing.T) {
    errCh := make(chan error, 1)
    
    errors.SafeGoWithError("test", errCh, func() error {
        panic("test panic")
    })
    
    err := <-errCh
    if err == nil {
        t.Error("expected error from panic recovery")
    }
    
    var intErr *errors.InternalError
    if !errors.As(err, &intErr) {
        t.Error("panic should be converted to InternalError")
    }
}
```

## Gradual Migration Strategy

1. **Phase 1**: New Code Only
   - All new code uses the new error system
   - Existing code remains unchanged

2. **Phase 2**: Critical Paths
   - Migrate high-traffic error paths first
   - Focus on goroutines needing panic recovery

3. **Phase 3**: Module by Module
   - Pick one module at a time
   - Complete migration before moving to next

4. **Phase 4**: Cleanup
   - Remove deprecated error patterns
   - Update all documentation
   - Add CI checks for error standards

## Common Pitfalls

### Pitfall 1: Losing Error Context

❌ **Wrong:**
```go
if err != nil {
    return errors.NewInternalError("module", 160, "component", "error occurred")
}
```

✅ **Correct:**
```go
if err != nil {
    return errors.WrapCoded(err, "module", 160, "error occurred", errors.CategoryInternal)
}
```

### Pitfall 2: Not Using Retryable Flag

❌ **Wrong:**
```go
// Timeout error but not marked retryable
return errors.NewCodedError("module", 151, "timeout", errors.CategoryTimeout)
```

✅ **Correct:**
```go
// Use TimeoutError which sets retryable=true
return errors.NewTimeoutError("module", 151, "operation", "30s")
```

### Pitfall 3: Incorrect Error Code Range

❌ **Wrong:**
```go
// Using code outside allocated range
return errors.NewCodedError("veid", 2000, "error", errors.CategoryInternal)
```

✅ **Correct:**
```go
// Use code within veid's range (1000-1099)
return errors.NewCodedError("veid", 1060, "error", errors.CategoryInternal)
```

### Pitfall 4: Forgetting Panic Recovery

❌ **Wrong:**
```go
go func() {
    // No panic recovery - will crash the program
    riskyOperation()
}()
```

✅ **Correct:**
```go
errors.SafeGo("risky-operation", func() {
    riskyOperation()
})
```

## Resources

- Error Handling Best Practices: `_docs/ERROR_HANDLING.md`
- Error Code Registry: `pkg/errors/codes.go`
- Complete Error Catalog: `docs/errors/ERROR_CATALOG.md`
- Client Error Guide: `docs/api/ERROR_HANDLING.md`
- Error Code Policy: `ERROR_CODE_POLICY.md`

## Questions?

For questions about migration:

1. Review the best practices guide
2. Check existing migrated code for examples
3. Ask in #dev-errors Slack channel
4. Open an issue with the `error-handling` label
