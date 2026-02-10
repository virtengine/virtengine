# Code Review Checklist

This checklist helps reviewers and authors ensure code meets VirtEngine quality standards.

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [General Code Quality](#general-code-quality)
3. [Module-Specific Checks](#module-specific-checks)
4. [Security Review](#security-review)
5. [Performance Review](#performance-review)
6. [Testing Review](#testing-review)
7. [Documentation Review](#documentation-review)
8. [Consensus Safety](#consensus-safety)

---

## Quick Reference

### Must-Have (Blocking)

- [ ] Tests pass locally
- [ ] Linting passes (`make lint-go`)
- [ ] No security vulnerabilities introduced
- [ ] Consensus-safe (deterministic where required)
- [ ] Commits signed off
- [ ] Commit messages follow conventions

### Should-Have (Non-Blocking)

- [ ] Adequate test coverage
- [ ] Documentation updated
- [ ] Error messages are helpful
- [ ] Logging is appropriate
- [ ] Performance considered

---

## General Code Quality

### Code Structure

- [ ] **Single Responsibility**: Each function/method does one thing
- [ ] **Appropriate Abstraction**: Not too abstract, not too concrete
- [ ] **Consistent Style**: Follows existing patterns in the codebase
- [ ] **Reasonable Size**: Functions under 50 lines, files under 500 lines

### Naming

- [ ] **Descriptive Names**: Variables and functions clearly convey purpose
- [ ] **Consistent Conventions**: Follows Go naming conventions
- [ ] **No Abbreviations**: Unless universally understood (e.g., `ctx`, `err`)
- [ ] **Package Names**: Lowercase, short, no underscores

```go
// Good
func (k Keeper) CreateOrder(ctx sdk.Context, msg *types.MsgCreateOrder) error

// Bad
func (k Keeper) CO(c sdk.Context, m *types.MsgCreateOrder) error
```

### Error Handling

- [ ] **Errors Wrapped**: Include context when wrapping
- [ ] **Sentinel Errors**: Used for expected conditions
- [ ] **No Silent Failures**: All errors handled or explicitly ignored
- [ ] **Helpful Messages**: Error messages aid debugging

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create order for customer %s: %w", msg.Customer, err)
}

// Bad
if err != nil {
    return err
}
```

### Comments

- [ ] **Doc Comments**: All exported types, functions, methods
- [ ] **Why, Not What**: Comments explain reasoning, not obvious code
- [ ] **No Stale Comments**: Comments match the code
- [ ] **No Commented-Out Code**: Remove dead code

```go
// Good
// CreateOrder creates a new marketplace order and holds payment in escrow.
// It validates the customer's identity score before proceeding.
func (k Keeper) CreateOrder(...) error

// Bad
// This function creates an order
func (k Keeper) CreateOrder(...) error
```

---

## Module-Specific Checks

### Keeper Implementation

- [ ] **IKeeper Interface**: All public methods in interface
- [ ] **Authority Field**: Uses `x/gov` module account
- [ ] **Store Key**: Uses `storetypes.StoreKey`
- [ ] **Context Validation**: Checks context deadlines for long operations

```go
// Pattern to follow
type IKeeper interface {
    CreateOrder(ctx sdk.Context, msg *types.MsgCreateOrder) (*types.Order, error)
    GetOrder(ctx sdk.Context, id string) (types.Order, bool)
}

type Keeper struct {
    cdc       codec.BinaryCodec
    storeKey  storetypes.StoreKey
    authority string  // x/gov module account
}
```

### Message Handlers

- [ ] **ValidateBasic**: Input validation in message
- [ ] **Authorization**: Checks in handler
- [ ] **Events Emitted**: Significant state changes emit events
- [ ] **Idempotency**: Consider repeated submissions

### Genesis

- [ ] **DefaultGenesisState**: Returns sensible defaults
- [ ] **Validate**: Checks all state invariants
- [ ] **Import/Export**: State properly serialized

### Types

- [ ] **Validation Methods**: All types have validation
- [ ] **Proto Definitions**: Match Go types
- [ ] **Store Keys**: Properly defined in `keys.go`

---

## Security Review

### Critical Checks

- [ ] **No Secrets in Code**: API keys, passwords, mnemonics
- [ ] **Input Validation**: All external input validated
- [ ] **Authorization**: Actions properly gated
- [ ] **No SQL/Command Injection**: Parameterized queries

### Sensitive Data

- [ ] **Encryption Required**: Sensitive data encrypted before storage
- [ ] **No Logging Secrets**: Private keys, passwords not logged
- [ ] **Proper Redaction**: Logs use observability package redaction

```go
// Bad - logs sensitive data
log.Printf("Processing with key: %s", privateKey)

// Good - logs fingerprint only
log.Printf("Processing with key fingerprint: %s", computeFingerprint(publicKey))
```

### Access Control

- [ ] **Role Checks**: RBAC properly enforced
- [ ] **MFA Gating**: High-risk operations require MFA
- [ ] **Authority Validation**: Only authorized accounts can execute

### Cryptography

- [ ] **Standard Algorithms**: Using approved algorithms
- [ ] **Proper Random**: Using crypto/rand for secrets
- [ ] **Key Management**: Keys properly stored and rotated

---

## Performance Review

### Query Performance

- [ ] **Pagination**: Large result sets paginated
- [ ] **Index Usage**: Queries use appropriate indexes
- [ ] **Bounded Iterations**: No unbounded loops over store

```go
// Good - bounded iteration
func (k Keeper) GetOrders(ctx sdk.Context, pagination *query.PageRequest) ([]Order, error)

// Bad - unbounded
func (k Keeper) GetAllOrders(ctx sdk.Context) []Order
```

### Memory Usage

- [ ] **No Memory Leaks**: Resources properly cleaned up
- [ ] **Reasonable Allocations**: No excessive allocations
- [ ] **Goroutine Safety**: No leaked goroutines

### Store Operations

- [ ] **Batch Operations**: Multiple writes batched when possible
- [ ] **Efficient Serialization**: Using proto encoding
- [ ] **Key Design**: Store keys designed for efficient queries

---

## Testing Review

### Test Coverage

- [ ] **Happy Path**: Normal operation tested
- [ ] **Error Paths**: All error conditions tested
- [ ] **Edge Cases**: Boundary conditions covered
- [ ] **Integration**: Cross-module interactions tested

### Test Quality

- [ ] **Meaningful Assertions**: Tests verify behavior, not implementation
- [ ] **Independent**: Tests don't depend on each other
- [ ] **Deterministic**: No flaky tests
- [ ] **Fast**: Unit tests complete quickly

### Test Patterns

- [ ] **Table-Driven**: Multiple cases in one test function
- [ ] **Subtests**: Related cases grouped with `t.Run()`
- [ ] **Goroutine Leaks**: Uses `goleak.VerifyNoLeaks(t)`

```go
// Good - table-driven test
func TestValidation(t *testing.T) {
    tests := []struct{
        name    string
        input   Input
        wantErr bool
    }{
        {"valid input", validInput, false},
        {"empty field", emptyField, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validate(tt.input)
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

## Documentation Review

### Code Documentation

- [ ] **Package Comment**: Describes package purpose
- [ ] **Exported Symbols**: All have doc comments
- [ ] **Examples**: Non-obvious usage has examples
- [ ] **README Updates**: New features documented

### API Documentation

- [ ] **Endpoint Docs**: REST/gRPC endpoints documented
- [ ] **Request/Response**: Examples provided
- [ ] **Error Codes**: Documented with meanings

### Changelog

- [ ] **Entry Added**: For user-visible changes
- [ ] **Correct Category**: Added/Fixed/Changed/Deprecated
- [ ] **Issue Reference**: Links to related issue

---

## Consensus Safety

### Determinism Requirements

All operations in consensus must be deterministic:

- [ ] **No Floating Point**: Use integer math or `sdk.Dec`
- [ ] **No Map Iteration**: Use sorted keys
- [ ] **No Time Functions**: Use block time from context
- [ ] **No Random**: Fixed seeds for any randomness
- [ ] **No External Calls**: No network/file in consensus path

```go
// Bad - non-deterministic
for k, v := range myMap {  // map iteration order is random
    process(k, v)
}

// Good - deterministic
keys := make([]string, 0, len(myMap))
for k := range myMap {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    process(k, myMap[k])
}
```

### ML Inference Determinism

For ML scoring (critical for VEID):

- [ ] **CPU Only**: ForceCPU = true
- [ ] **Fixed Seed**: RandomSeed = 42
- [ ] **Deterministic Ops**: TensorFlow deterministic mode
- [ ] **Pinned Versions**: requirements-deterministic.txt

```go
// Required settings for consensus
config := DeterminismConfig{
    ForceCPU:         true,
    RandomSeed:       42,
    DeterministicOps: true,
}
```

### State Machine Safety

- [ ] **Idempotent**: Repeated execution safe
- [ ] **Atomic**: State changes all-or-nothing
- [ ] **No Panics**: Error handling instead of panics
- [ ] **Bounded**: No unbounded computation

---

## Review Workflow

### As a Reviewer

1. **Check CI Status**: Ensure automated checks pass
2. **Read PR Description**: Understand intent
3. **Review File by File**: Use checklist for each
4. **Test Locally**: For significant changes
5. **Provide Constructive Feedback**: Explain why

### As an Author

1. **Self-Review First**: Use this checklist before requesting review
2. **Keep PRs Small**: Easier to review, faster to merge
3. **Respond Promptly**: Address feedback quickly
4. **Don't Take It Personally**: Reviews improve code quality

### Review Comments

Use these prefixes for clarity:

| Prefix | Meaning |
|--------|---------|
| `blocking:` | Must be fixed before merge |
| `nit:` | Minor suggestion, not blocking |
| `question:` | Seeking clarification |
| `suggestion:` | Alternative approach to consider |
| `praise:` | Something done well |

---

## Printable Checklist

Copy this for PR reviews:

```markdown
## Code Review Checklist

### Required
- [ ] Tests pass
- [ ] Linting passes
- [ ] No security issues
- [ ] Consensus-safe
- [ ] Commits signed off

### Code Quality
- [ ] Clear naming
- [ ] Proper error handling
- [ ] Appropriate comments
- [ ] Follows patterns

### Security
- [ ] Input validated
- [ ] No secrets exposed
- [ ] Access controlled
- [ ] Sensitive data encrypted

### Testing
- [ ] Happy path covered
- [ ] Error paths covered
- [ ] No flaky tests

### Documentation
- [ ] Code documented
- [ ] Changelog updated
- [ ] README updated (if needed)
```

---

## Related Documentation

- [Code Contribution](./03-code-contribution.md) - PR process
- [Testing Guide](./04-testing-guide.md) - Writing tests
- [Patterns & Anti-patterns](./07-patterns-antipatterns.md) - Best practices
- [Security Guide](../threat-model.md) - Security considerations
