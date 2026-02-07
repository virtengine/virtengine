# VirtEngine Security Coding Guidelines

**Version:** 1.0.0  
**Last Updated:** 2025-01-20  
**Status:** Active  
**Maintainers:** VirtEngine Security Team

---

## Table of Contents

1. [Overview](#overview)
2. [Secret Management](#secret-management)
3. [Input Validation](#input-validation)
4. [Cryptography](#cryptography)
5. [Error Handling](#error-handling)
6. [Code Review Security Checklist](#code-review-security-checklist)
7. [Common Secure Coding Patterns](#common-secure-coding-patterns)
8. [nolint Directive Policy](#nolint-directive-policy)
9. [Security Testing](#security-testing)
10. [References](#references)

---

## Overview

VirtEngine is a Cosmos SDK-based blockchain for decentralized cloud computing with ML-powered identity verification. Security is paramount because:

- **Consensus Safety**: All validators must produce identical state transitions. Non-deterministic code causes chain halts.
- **Financial Assets**: The chain handles real value through escrow, staking, and marketplace transactions.
- **Identity Data**: VEID module processes sensitive biometric and identity information.
- **Infrastructure Control**: Provider daemon manages cloud resources with privileged access.

### Security Principles

1. **Defense in Depth**: Multiple layers of security controls
2. **Least Privilege**: Components have minimal required permissions
3. **Fail Secure**: Errors default to denying access
4. **Zero Trust**: Validate all inputs, even from trusted sources
5. **Determinism**: All on-chain operations must be reproducible

### Threat Model

| Threat Actor | Capabilities | Primary Targets |
|--------------|--------------|-----------------|
| Malicious Validator | Full node access, consensus participation | State manipulation, censorship |
| External Attacker | Network access, API calls | DoS, data exfiltration |
| Malicious Provider | Infrastructure control | Resource theft, data access |
| Compromised Client | Signed transactions | Identity fraud, fund theft |

---

## Secret Management

### Never Hardcode Credentials

**CRITICAL**: Secrets must never appear in source code, configuration files checked into git, or log output.

```go
// ❌ NEVER DO THIS
const apiKey = "sk-1234567890abcdef"
dbPassword := "production-password-123"

// ❌ NEVER DO THIS - even in "test" code
var testCredentials = map[string]string{
    "admin": "admin123",
}

// ✅ CORRECT: Use environment variables
apiKey := os.Getenv("VIRTENGINE_API_KEY")
if apiKey == "" {
    return errors.New("VIRTENGINE_API_KEY environment variable required")
}
```

### Environment Variables

All sensitive configuration must use environment variables with the `VIRTENGINE_` prefix.

#### Required Environment Variables

| Variable | Description | Required By |
|----------|-------------|-------------|
| `VIRTENGINE_HOME` | Base directory for chain data | All components |
| `VIRTENGINE_KEYRING_BACKEND` | Keyring type (os, file, test) | CLI, daemon |
| `VIRTENGINE_NODE` | RPC endpoint URL | CLI, provider |
| `VIRTENGINE_CHAIN_ID` | Network chain ID | All components |

#### Provider Daemon Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PROVIDER_KEY_SECRET` | Provider signing key passphrase | Yes |
| `VAULT_PASSWORD` | Ansible vault password | If using Ansible |
| `KUBECONFIG` | Kubernetes config path | If using K8s adapter |
| `AWS_ACCESS_KEY_ID` | AWS credentials | If using AWS adapter |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials | If using AWS adapter |
| `AZURE_CLIENT_SECRET` | Azure credentials | If using Azure adapter |

#### ML Inference Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TF_CPP_MIN_LOG_LEVEL` | TensorFlow log level | 2 |
| `TF_DETERMINISTIC_OPS` | Force deterministic ops | 1 |
| `CUDA_VISIBLE_DEVICES` | GPU device selection | -1 (CPU only) |

### Secret Handling in Code

```go
// ✅ Clear secrets from memory after use
func (k *KeyManager) SignTransaction(tx []byte, passphrase string) ([]byte, error) {
    // Clear passphrase from memory when done
    defer func() {
        for i := range passphrase {
            passphrase = passphrase[:i] + "\x00" + passphrase[i+1:]
        }
    }()
    
    // Use passphrase for signing...
}

// ✅ Never log secrets
func ConnectToDatabase(connStr string) error {
    // Log connection attempt without credentials
    log.Info("connecting to database", "host", extractHost(connStr))
    // NOT: log.Info("connecting", "connStr", connStr)
}
```

---

## Input Validation

### Validate All Inputs

Every external input must be validated before use. This includes:

- Transaction messages
- Query parameters
- RPC/gRPC requests
- File uploads
- Environment variables
- Configuration files

```go
// ✅ CORRECT: Comprehensive input validation
func (k Keeper) CreateOrder(ctx sdk.Context, msg *MsgCreateOrder) error {
    // Validate address format
    if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
        return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", err)
    }
    
    // Validate numeric ranges
    if msg.Price.IsNegative() || msg.Price.IsZero() {
        return sdkerrors.ErrInvalidRequest.Wrap("price must be positive")
    }
    
    // Validate string lengths
    if len(msg.Description) > MaxDescriptionLength {
        return sdkerrors.ErrInvalidRequest.Wrapf(
            "description exceeds max length: %d > %d",
            len(msg.Description), MaxDescriptionLength,
        )
    }
    
    // Validate against allowlist
    if !IsAllowedResourceType(msg.ResourceType) {
        return sdkerrors.ErrInvalidRequest.Wrapf(
            "invalid resource type: %s", msg.ResourceType,
        )
    }
    
    return nil
}
```

### Use Allowlists, Not Blocklists

```go
// ❌ WRONG: Blocklist approach - can be bypassed
func validateFilename(name string) error {
    blocked := []string{"..", "/", "\\", "\x00"}
    for _, b := range blocked {
        if strings.Contains(name, b) {
            return errors.New("invalid filename")
        }
    }
    return nil
}

// ✅ CORRECT: Allowlist approach
var validFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,254}$`)

func validateFilename(name string) error {
    if !validFilenameRegex.MatchString(name) {
        return errors.New("filename must be alphanumeric with ._- allowed")
    }
    return nil
}
```

### Size Limits

Define and enforce size limits for all variable-length inputs:

```go
// types/limits.go - Central limit definitions
const (
    // Message field limits
    MaxDescriptionLength = 1024
    MaxMemoLength        = 256
    MaxLabelLength       = 64
    MaxLabelsPerResource = 16
    
    // Upload limits
    MaxManifestSize      = 1 << 20  // 1 MB
    MaxImageUploadSize   = 10 << 20 // 10 MB
    MaxBatchSize         = 100
    
    // Query limits
    MaxPageSize          = 100
    DefaultPageSize      = 50
)

// ✅ Enforce limits in handlers
func (k Keeper) HandleManifest(ctx sdk.Context, manifest []byte) error {
    if len(manifest) > MaxManifestSize {
        return sdkerrors.ErrInvalidRequest.Wrapf(
            "manifest size %d exceeds limit %d",
            len(manifest), MaxManifestSize,
        )
    }
    // Process manifest...
}
```

---

## Cryptography

### Approved Algorithms Only

VirtEngine uses the following cryptographic algorithms exclusively:

| Purpose | Algorithm | Library |
|---------|-----------|---------|
| Key Exchange | X25519 | `golang.org/x/crypto/curve25519` |
| Symmetric Encryption | XSalsa20-Poly1305 | `golang.org/x/crypto/nacl/secretbox` |
| AEAD Encryption | AES-256-GCM | `crypto/aes` + `crypto/cipher` |
| Hashing | SHA-256 | `crypto/sha256` |
| Signing | Ed25519, secp256k1 | Cosmos SDK keyring |

```go
// ✅ CORRECT: Use approved algorithms
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    
    "golang.org/x/crypto/curve25519"
    "golang.org/x/crypto/nacl/box"
)

// Create encryption envelope using X25519 + XSalsa20-Poly1305
func EncryptEnvelope(plaintext []byte, recipientPubKey *[32]byte) (*EncryptionEnvelope, error) {
    // Generate ephemeral keypair
    var ephemeralPub, ephemeralPriv [32]byte
    if _, err := rand.Read(ephemeralPriv[:]); err != nil {
        return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
    }
    curve25519.ScalarBaseMult(&ephemeralPub, &ephemeralPriv)
    
    // Generate nonce
    var nonce [24]byte
    if _, err := rand.Read(nonce[:]); err != nil {
        return nil, fmt.Errorf("failed to generate nonce: %w", err)
    }
    
    // Encrypt using NaCl box
    ciphertext := box.Seal(nil, plaintext, &nonce, recipientPubKey, &ephemeralPriv)
    
    // Clear ephemeral private key
    for i := range ephemeralPriv {
        ephemeralPriv[i] = 0
    }
    
    return &EncryptionEnvelope{
        Algorithm:   "X25519-XSalsa20-Poly1305",
        EphemeralPK: ephemeralPub[:],
        Nonce:       nonce[:],
        Ciphertext:  ciphertext,
    }, nil
}
```

### Never Implement Custom Cryptography

```go
// ❌ NEVER DO THIS - Custom crypto implementations
func myEncrypt(data []byte, key []byte) []byte {
    result := make([]byte, len(data))
    for i, b := range data {
        result[i] = b ^ key[i%len(key)]
    }
    return result
}

// ❌ NEVER DO THIS - Custom random generation
func myRandom() int64 {
    return time.Now().UnixNano()
}

// ❌ NEVER DO THIS - Custom hashing
func myHash(data []byte) []byte {
    // Some homegrown algorithm
}
```

### Deterministic Operations for Consensus

All cryptographic operations in on-chain code must be deterministic:

```go
// ✅ CORRECT: Deterministic hash for consensus
func ComputeOrderHash(order *Order) []byte {
    h := sha256.New()
    // Use canonical encoding
    bz, _ := order.Marshal()
    h.Write(bz)
    return h.Sum(nil)
}

// ❌ WRONG: Non-deterministic in consensus code
func ComputeOrderHash(order *Order) []byte {
    // JSON encoding is NOT deterministic (map key ordering)
    bz, _ := json.Marshal(order)
    return sha256.Sum256(bz)
}

// ✅ CORRECT: ML inference with determinism config
type DeterminismConfig struct {
    ForceCPU         bool  // true - no GPU variance
    RandomSeed       int64 // fixed seed (default: 42)
    DeterministicOps bool  // true - TF deterministic ops
}

func NewConsensusScorer() *Scorer {
    return &Scorer{
        config: DeterminismConfig{
            ForceCPU:         true,
            RandomSeed:       42,
            DeterministicOps: true,
        },
    }
}
```

---

## Error Handling

### Don't Leak Internal Details

Error messages returned to users should be informative but not expose internal implementation details.

```go
// ❌ WRONG: Exposes internal details
func (k Keeper) GetUser(ctx sdk.Context, id string) (*User, error) {
    user, err := k.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        // Exposes database schema and query
        return nil, fmt.Errorf("database query failed: %v", err)
    }
    return user, nil
}

// ✅ CORRECT: Generic user-facing error, detailed internal logging
func (k Keeper) GetUser(ctx sdk.Context, id string) (*User, error) {
    user, err := k.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        // Log detailed error internally
        k.logger.Error("database query failed",
            "error", err,
            "query", "GetUser",
            "id", id,
        )
        // Return generic error to user
        return nil, sdkerrors.ErrNotFound.Wrap("user not found")
    }
    return user, nil
}
```

### Proper Error Types

Use Cosmos SDK error types consistently:

```go
import (
    sdkerrors "cosmossdk.io/errors"
    "github.com/virtengine/virtengine/x/market/types"
)

// Module-specific error codes in types/errors.go
var (
    ErrOrderNotFound     = sdkerrors.Register(ModuleName, 2, "order not found")
    ErrInvalidBid        = sdkerrors.Register(ModuleName, 3, "invalid bid")
    ErrUnauthorized      = sdkerrors.Register(ModuleName, 4, "unauthorized")
    ErrInsufficientFunds = sdkerrors.Register(ModuleName, 5, "insufficient funds")
)

// ✅ CORRECT: Use typed errors with wrapping
func (k Keeper) AcceptBid(ctx sdk.Context, orderID, bidID uint64) error {
    order, found := k.GetOrder(ctx, orderID)
    if !found {
        return types.ErrOrderNotFound.Wrapf("order %d", orderID)
    }
    
    if order.State != OrderStateOpen {
        return types.ErrInvalidBid.Wrapf(
            "order %d is in state %s, expected %s",
            orderID, order.State, OrderStateOpen,
        )
    }
    
    return nil
}
```

### Panic vs Error

```go
// ✅ Use panic only for programming errors (invariant violations)
func (k Keeper) MustGetOrder(ctx sdk.Context, id uint64) types.Order {
    order, found := k.GetOrder(ctx, id)
    if !found {
        // This should never happen if caller verified order exists
        panic(fmt.Sprintf("order %d must exist", id))
    }
    return order
}

// ✅ Use errors for expected failure cases
func (k Keeper) GetOrder(ctx sdk.Context, id uint64) (types.Order, bool) {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get(types.OrderKey(id))
    if bz == nil {
        return types.Order{}, false
    }
    var order types.Order
    k.cdc.MustUnmarshal(bz, &order)
    return order, true
}
```

---

## Code Review Security Checklist

Use this checklist when reviewing PRs:

### Authentication & Authorization

- [ ] All endpoints require appropriate authentication
- [ ] Authorization checks happen before any state changes
- [ ] Module authority is validated for admin operations
- [ ] Signer verification uses `msg.GetSigners()` correctly

### Input Validation

- [ ] All message fields are validated in `ValidateBasic()`
- [ ] Keeper methods validate inputs before state access
- [ ] Size limits enforced for variable-length fields
- [ ] Allowlists used instead of blocklists

### Cryptography

- [ ] Only approved algorithms used
- [ ] No custom crypto implementations
- [ ] Random values from `crypto/rand`, not `math/rand`
- [ ] Keys/secrets cleared from memory after use

### State Management

- [ ] No non-deterministic operations in consensus code
- [ ] Store keys are properly prefixed and documented
- [ ] Iterator resources are closed (defer iterator.Close())
- [ ] Genesis import/export is symmetric

### Error Handling

- [ ] Errors don't leak sensitive information
- [ ] All errors are properly wrapped with context
- [ ] Panic only for invariant violations
- [ ] Resources cleaned up in error paths

### Concurrency

- [ ] No data races (run `go test -race`)
- [ ] Proper mutex usage for shared state
- [ ] Context cancellation respected
- [ ] Goroutine leaks prevented (goleak in tests)

### Dependencies

- [ ] New dependencies are vetted for security
- [ ] No known vulnerabilities (`govulncheck`)
- [ ] Replace directives for forked deps documented

---

## Common Secure Coding Patterns

### Safe Type Assertions

Always use the two-value form of type assertions to prevent panics:

```go
// ❌ WRONG: Will panic if assertion fails
func handleMessage(msg sdk.Msg) {
    createOrder := msg.(*MsgCreateOrder)
    // Process...
}

// ✅ CORRECT: Safe type assertion with ok-pattern
func handleMessage(msg sdk.Msg) error {
    createOrder, ok := msg.(*MsgCreateOrder)
    if !ok {
        return sdkerrors.ErrInvalidType.Wrapf(
            "expected *MsgCreateOrder, got %T", msg,
        )
    }
    // Process...
    return nil
}

// ✅ CORRECT: Type switch for multiple message types
func handleMessage(msg sdk.Msg) error {
    switch m := msg.(type) {
    case *MsgCreateOrder:
        return handleCreateOrder(m)
    case *MsgCancelOrder:
        return handleCancelOrder(m)
    default:
        return sdkerrors.ErrUnknownRequest.Wrapf(
            "unrecognized message type: %T", msg,
        )
    }
}
```

### Proper Deferred Close() Error Handling

```go
// ❌ WRONG: Ignores Close() error
func readConfig(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close() // Error ignored!
    
    return io.ReadAll(f)
}

// ✅ CORRECT: Handle Close() error with named return
func readConfig(path string) (data []byte, err error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open config: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close config: %w", cerr)
        }
    }()
    
    data, err = io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }
    return data, nil
}

// ✅ CORRECT: For write operations, always check Close()
func writeData(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close file: %w", cerr)
        }
    }()
    
    if _, err := f.Write(data); err != nil {
        return fmt.Errorf("write data: %w", err)
    }
    
    // Sync to ensure data is flushed to disk
    if err := f.Sync(); err != nil {
        return fmt.Errorf("sync file: %w", err)
    }
    
    return nil
}
```

### HTTP vs HTTPS Usage

```go
// ❌ WRONG: Insecure HTTP for production
client := &http.Client{}
resp, err := client.Get("http://api.example.com/data")

// ✅ CORRECT: Enforce HTTPS with proper TLS config
func newSecureClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
                // Use system CA pool
                RootCAs: nil, // Uses system pool
            },
            // Timeouts to prevent resource exhaustion
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            TLSHandshakeTimeout:   10 * time.Second,
            ResponseHeaderTimeout: 10 * time.Second,
        },
        Timeout: 60 * time.Second,
    }
}

// ✅ CORRECT: Validate URL scheme
func fetchResource(urlStr string) ([]byte, error) {
    u, err := url.Parse(urlStr)
    if err != nil {
        return nil, fmt.Errorf("invalid URL: %w", err)
    }
    
    // Enforce HTTPS in production
    if os.Getenv("VIRTENGINE_ENV") == "production" && u.Scheme != "https" {
        return nil, errors.New("HTTPS required in production")
    }
    
    client := newSecureClient()
    resp, err := client.Get(urlStr)
    // ...
}

// ✅ CORRECT: Local development exception (explicit)
func isLocalDevelopment(host string) bool {
    return host == "localhost" || 
           host == "127.0.0.1" || 
           strings.HasSuffix(host, ".local")
}
```

### Context Timeout Handling

```go
// ✅ CORRECT: Respect context deadlines
func (k Keeper) ProcessWithTimeout(ctx sdk.Context, data []byte) error {
    // Check if context has deadline
    if deadline, ok := ctx.Context().Deadline(); ok {
        remaining := time.Until(deadline)
        if remaining < MinProcessingTime {
            return sdkerrors.ErrInvalidRequest.Wrap("insufficient time remaining")
        }
    }
    
    // Create cancellable context for long operations
    processCtx, cancel := context.WithTimeout(ctx.Context(), MaxProcessingTime)
    defer cancel()
    
    select {
    case result := <-k.asyncProcess(processCtx, data):
        return result
    case <-processCtx.Done():
        return sdkerrors.ErrTimeout.Wrap("processing timed out")
    }
}
```

---

## nolint Directive Policy

### General Policy

All `nolint` directives MUST include a justification explaining why the lint rule is being suppressed. Unjustified nolint directives will be rejected in code review.

### Format

```go
//nolint:lintername // REASON: explanation of why this is acceptable
```

### Acceptable Justifications

#### Security-Related

```go
// ✅ ACCEPTABLE: Documented security exception
//nolint:gosec // REASON: G404 - math/rand used for non-security shuffle of display order only
rand.Shuffle(len(items), func(i, j int) {
    items[i], items[j] = items[j], items[i]
})

// ✅ ACCEPTABLE: False positive with explanation
//nolint:gosec // REASON: G304 - filepath is validated against allowlist in validatePath()
data, err := os.ReadFile(filepath)
```

#### Error Handling

```go
// ✅ ACCEPTABLE: Intentionally ignoring error with reason
//nolint:errcheck // REASON: Best-effort cleanup, error logged but not propagated
defer func() {
    if err := cleanup(); err != nil {
        logger.Warn("cleanup failed", "error", err)
    }
}()

// ✅ ACCEPTABLE: Test code with controlled conditions
//nolint:errcheck // REASON: Test setup - panic on error is acceptable in test
f.Write(testData)
```

#### Performance

```go
// ✅ ACCEPTABLE: Intentional design decision
//nolint:prealloc // REASON: Slice grows unpredictably based on filter conditions
var results []Order
for _, o := range orders {
    if filter(o) {
        results = append(results, o)
    }
}
```

#### Cosmos SDK Patterns

```go
// ✅ ACCEPTABLE: SDK convention
//nolint:staticcheck // REASON: SA1019 - Using deprecated method for SDK v0.50 compatibility
ctx = ctx.WithBlockHeight(height)
```

### Unacceptable Justifications

```go
// ❌ UNACCEPTABLE: No justification
//nolint:errcheck

// ❌ UNACCEPTABLE: Vague justification
//nolint:gosec // TODO: fix later

// ❌ UNACCEPTABLE: Lazy justification
//nolint:errcheck // don't care about error

// ❌ UNACCEPTABLE: Blanket suppression
//nolint:all
```

### File-Level Directives

File-level nolint directives are strongly discouraged. If required, they must be documented at the top of the file:

```go
// Package legacy contains deprecated code maintained for backwards compatibility.
// This file uses patterns that trigger lint warnings but cannot be refactored
// without breaking API compatibility.
//
//nolint:staticcheck // REASON: Entire file uses deprecated SDK patterns for v0.47 compat
package legacy
```

---

## Security Testing

### Test Suite Location

Security-focused tests are located in `tests/security/`:

```
tests/security/
├── auth_test.go           # Authentication bypass tests
├── crypto_test.go         # Cryptographic operation tests
├── input_validation_test.go # Input validation fuzzing
├── authorization_test.go  # Permission boundary tests
├── injection_test.go      # Injection attack tests
└── determinism_test.go    # Consensus determinism tests
```

### Running Security Tests

```bash
# Run all security tests
go test -v ./tests/security/...

# Run with race detector
go test -race ./tests/security/...

# Run specific test category
go test -v ./tests/security/... -run TestInputValidation
```

### Security Test Patterns

```go
// tests/security/input_validation_test.go
func TestMessageValidation_MaliciousInput(t *testing.T) {
    testCases := []struct {
        name    string
        msg     sdk.Msg
        wantErr bool
    }{
        {
            name: "oversized description",
            msg: &types.MsgCreateOrder{
                Description: strings.Repeat("a", MaxDescriptionLength+1),
            },
            wantErr: true,
        },
        {
            name: "null bytes in string",
            msg: &types.MsgCreateOrder{
                Description: "valid\x00hidden",
            },
            wantErr: true,
        },
        {
            name: "path traversal attempt",
            msg: &types.MsgUploadManifest{
                Path: "../../../etc/passwd",
            },
            wantErr: true,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := tc.msg.ValidateBasic()
            if tc.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

// tests/security/determinism_test.go
func TestMLScoring_Determinism(t *testing.T) {
    scorer := inference.NewConsensusScorer()
    input := loadTestInput(t)
    
    // Run scoring multiple times
    results := make([]float64, 10)
    for i := 0; i < 10; i++ {
        score, err := scorer.Score(input)
        require.NoError(t, err)
        results[i] = score
    }
    
    // All results must be identical for consensus
    for i := 1; i < len(results); i++ {
        require.Equal(t, results[0], results[i],
            "ML scoring must be deterministic for consensus")
    }
}
```

### Vulnerability Scanning

```bash
# Run Go vulnerability checker
govulncheck ./...

# Run gosec static analysis
gosec -fmt=json -out=security-report.json ./...

# Run semgrep for additional patterns
semgrep --config=p/golang ./
```

---

## References

### Internal Documentation

- [Development Environment Setup](development-environment.md)
- [Testing Guide](testing-guide.md)
- [VEID Module Architecture](../x/veid/README.md)
- [Encryption Module](../x/encryption/README.md)
- [Provider Daemon Security](../pkg/provider_daemon/SECURITY.md)

### External Resources

- [Cosmos SDK Security Best Practices](https://docs.cosmos.network/main/build/building-modules/errors)
- [Go Secure Coding Guidelines](https://github.com/OWASP/Go-SCP)
- [CWE Top 25 Most Dangerous Software Weaknesses](https://cwe.mitre.org/top25/)
- [OWASP Cryptographic Failures](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/)

### Reporting Security Issues

For security vulnerabilities, please email security@virtengine.io with:

1. Description of the vulnerability
2. Steps to reproduce
3. Potential impact assessment
4. Any suggested mitigations

Do NOT open public GitHub issues for security vulnerabilities.

---

*This document is reviewed quarterly. Last security audit: Q4 2024.*
