# Debugging Guide

This guide covers debugging techniques for VirtEngine development.

## Table of Contents

1. [Debugging Tools](#debugging-tools)
2. [Common Issues](#common-issues)
3. [Debugging Tests](#debugging-tests)
4. [Debugging the Chain](#debugging-the-chain)
5. [Debugging Provider Daemon](#debugging-provider-daemon)
6. [Debugging ML Inference](#debugging-ml-inference)
7. [Log Analysis](#log-analysis)
8. [Performance Debugging](#performance-debugging)

---

## Debugging Tools

### Essential Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| Delve | Go debugger | `go install github.com/go-delve/delve/cmd/dlv@latest` |
| pprof | Profiling | Built into Go |
| go-torch | Flame graphs | `go install github.com/uber/go-torch@latest` |
| grpcurl | gRPC testing | `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` |
| jq | JSON processing | Package manager |

### IDE Debugging

#### VS Code

```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${fileDirname}",
            "args": ["-test.run", "TestYourFunction"]
        },
        {
            "name": "Debug Chain",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/virtengine",
            "args": ["start", "--home", ".virtengine"]
        }
    ]
}
```

#### GoLand

1. Right-click test function → "Debug"
2. Or create a Run Configuration for main binary

---

## Common Issues

### Build Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `package not found` | Missing dependency | `go mod tidy` |
| `cannot find module` | Module path mismatch | Check `go.mod` and imports |
| `cgo: C compiler not found` | Missing C compiler | Install build-essential (Linux) or Xcode CLI (macOS) |
| `undefined: ...` | Import missing | Add import or check spelling |

### Runtime Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `nil pointer dereference` | Uninitialized pointer | Add nil checks |
| `index out of range` | Array bounds | Check length before access |
| `concurrent map write` | Race condition | Use sync.Map or mutex |
| `context deadline exceeded` | Timeout | Increase timeout or optimize |

### Chain Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `invalid chain-id` | Wrong chain ID | Check config |
| `account not found` | Missing account | Create account or fund it |
| `signature verification failed` | Wrong key | Check keyring |
| `insufficient funds` | Low balance | Fund account |

---

## Debugging Tests

### Verbose Output

```bash
# Show test names and timing
go test -v ./x/veid/...

# Include all log output
go test -v -count=1 ./x/veid/...
```

### Run Single Test

```bash
# By exact name
go test -v -run TestCreateOrder ./x/market/keeper/...

# By pattern
go test -v -run "TestOrder.*" ./x/market/...

# Subtests
go test -v -run "TestOrder/with_escrow" ./x/market/...
```

### Using Delve

```bash
# Debug a specific test
dlv test ./x/veid/keeper/... -- -test.run TestVerifyIdentity

# Delve commands:
# break main.go:42    - Set breakpoint
# break TestFunc      - Break at function
# continue (c)        - Continue execution
# next (n)            - Step over
# step (s)            - Step into
# print var (p var)   - Print variable
# locals              - Show local variables
# stack               - Show call stack
# quit (q)            - Exit
```

### Debug Output

```go
// Add debug output in tests
func TestMyFunction(t *testing.T) {
    t.Logf("Input: %+v", input)
    
    result, err := myFunction(input)
    
    t.Logf("Result: %+v, Error: %v", result, err)
}
```

### Race Detection

```bash
# Run with race detector
go test -race ./x/...

# The race detector will report:
# - Where the race occurred
# - Which goroutines are involved
# - Stack traces for both accesses
```

---

## Debugging the Chain

### Increase Log Level

```bash
# Start chain with debug logging
virtengine start --log_level debug

# Or specific modules
virtengine start --log_level "x/market:debug,x/veid:debug"
```

### Query Chain State

```bash
# Check chain status
curl http://localhost:26657/status | jq

# Query account
virtengine query auth account <address>

# Query module state
virtengine query veid params
virtengine query market orders

# Get block details
virtengine query block 12345
```

### Inspect Transactions

```bash
# Query transaction by hash
virtengine query tx <hash>

# Query transaction events
virtengine query txs --events 'message.action=/virtengine.market.v1.MsgCreateOrder'

# Decode transaction
virtengine tx decode <base64-encoded-tx>
```

### Simulate Transactions

```bash
# Dry-run transaction
virtengine tx market create-order \
    --offering offering_123 \
    --quantity 1 \
    --from customer \
    --dry-run

# Generate unsigned transaction
virtengine tx market create-order \
    --offering offering_123 \
    --quantity 1 \
    --from customer \
    --generate-only > unsigned.json
```

### Debug with gRPC

```bash
# List available services
grpcurl -plaintext localhost:9090 list

# Describe service
grpcurl -plaintext localhost:9090 describe virtengine.market.v1.Query

# Call method
grpcurl -plaintext \
    -d '{"order_id": "order_123"}' \
    localhost:9090 \
    virtengine.market.v1.Query/Order
```

### Genesis Debugging

```bash
# Export current state
virtengine export > genesis_export.json

# Validate genesis
virtengine genesis validate genesis.json

# Debug genesis import
virtengine genesis import genesis.json --log_level debug
```

---

## Debugging Provider Daemon

### Enable Debug Logging

```yaml
# provider-config.yaml
log:
  level: debug
  format: json
```

### Check Daemon Status

```bash
# Health check
curl http://localhost:8443/health

# Metrics
curl http://localhost:8443/metrics

# Active leases
curl http://localhost:8443/leases
```

### Debug Bid Engine

```go
// Add debug logging in bid_engine.go
func (be *BidEngine) ProcessOrder(ctx context.Context, order Order) (*Bid, error) {
    be.logger.Debug("processing order",
        "order_id", order.ID,
        "customer", order.Customer,
        "requirements", order.Requirements,
    )
    
    // ... processing logic
    
    be.logger.Debug("computed bid",
        "order_id", order.ID,
        "price", bid.Price,
        "resources", bid.Resources,
    )
    
    return bid, nil
}
```

### Debug Kubernetes Adapter

```bash
# Check adapter logs
kubectl logs -f deployment/provider-daemon -n virtengine

# Check pod status
kubectl get pods -n virtengine-<lease-id>

# Describe pod for events
kubectl describe pod <pod-name> -n virtengine-<lease-id>

# Get pod logs
kubectl logs <pod-name> -n virtengine-<lease-id>
```

### Debug Workload States

```go
// Valid state transitions
var validTransitions = map[WorkloadState][]WorkloadState{
    WorkloadStatePending:   {WorkloadStateDeploying, WorkloadStateFailed},
    WorkloadStateDeploying: {WorkloadStateRunning, WorkloadStateFailed, WorkloadStateStopped},
    WorkloadStateRunning:   {WorkloadStatePaused, WorkloadStateStopping, WorkloadStateFailed},
    WorkloadStatePaused:    {WorkloadStateRunning, WorkloadStateStopping},
    WorkloadStateStopping:  {WorkloadStateStopped, WorkloadStateFailed},
    WorkloadStateStopped:   {WorkloadStateTerminated},
}

// Debug invalid transition
func (w *Workload) SetState(newState WorkloadState) error {
    valid := validTransitions[w.State]
    for _, s := range valid {
        if s == newState {
            log.Printf("State transition: %s -> %s", w.State, newState)
            w.State = newState
            return nil
        }
    }
    return fmt.Errorf("invalid state transition: %s -> %s", w.State, newState)
}
```

---

## Debugging ML Inference

### Determinism Issues

ML scoring must be deterministic for consensus:

```go
// Check determinism config
config := DeterminismConfig{
    ForceCPU:         true,   // Must be true
    RandomSeed:       42,     // Must be fixed
    DeterministicOps: true,   // Must be true
}

// Debug: Compare scores across runs
func TestScoreDeterminism(t *testing.T) {
    scorer := NewScorer(config)
    
    var scores []float64
    for i := 0; i < 10; i++ {
        score, err := scorer.Score(testInput)
        require.NoError(t, err)
        scores = append(scores, score)
    }
    
    // All scores must be identical
    for i := 1; i < len(scores); i++ {
        require.Equal(t, scores[0], scores[i],
            "Score differed on run %d: %f vs %f", i, scores[0], scores[i])
    }
}
```

### Model Loading Issues

```go
func (s *Scorer) loadModel(path string) error {
    // Debug: Check file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Errorf("model file not found: %s", path)
    }
    
    // Debug: Check file size
    info, _ := os.Stat(path)
    s.logger.Debug("loading model",
        "path", path,
        "size_mb", info.Size()/(1024*1024),
    )
    
    // Load model
    model, err := tf.LoadSavedModel(path, []string{"serve"}, nil)
    if err != nil {
        return fmt.Errorf("failed to load model: %w", err)
    }
    
    s.logger.Debug("model loaded successfully",
        "operations", len(model.Graph.Operations()),
    )
    
    return nil
}
```

### Inference Errors

```go
func (s *Scorer) Score(input []byte) (float64, error) {
    // Debug: Log input shape
    tensor, err := tf.NewTensor(input)
    if err != nil {
        return 0, fmt.Errorf("failed to create tensor: %w", err)
    }
    s.logger.Debug("input tensor", "shape", tensor.Shape())
    
    // Run inference
    result, err := s.model.Session.Run(
        map[tf.Output]*tf.Tensor{s.input: tensor},
        []tf.Output{s.output},
        nil,
    )
    if err != nil {
        return 0, fmt.Errorf("inference failed: %w", err)
    }
    
    // Debug: Log output
    s.logger.Debug("inference result",
        "shape", result[0].Shape(),
        "dtype", result[0].DataType(),
    )
    
    return result[0].Value().([][]float32)[0][0], nil
}
```

---

## Log Analysis

### Structured Logging

VirtEngine uses structured logging:

```go
// Using observability package
logger := observability.NewLogger(observability.Config{
    Level:  "debug",
    Format: "json",
})

logger.Info("processing order",
    "order_id", order.ID,
    "customer", order.Customer,
    "amount", order.Amount,
)
```

### Log Filtering

```bash
# Filter JSON logs with jq
cat logs.json | jq 'select(.level == "error")'

# Filter by module
cat logs.json | jq 'select(.module == "market")'

# Filter by time range
cat logs.json | jq 'select(.timestamp > "2026-01-01T00:00:00Z")'

# Extract specific fields
cat logs.json | jq '{time: .timestamp, msg: .message, error: .error}'
```

### Common Log Patterns

```bash
# Find errors
grep -r "error" logs/ | grep -v "no error"

# Find panics
grep -r "panic" logs/

# Find slow operations
grep -r "slow" logs/ | grep -E "[0-9]+ms"

# Find failed transactions
grep -r "failed" logs/ | grep -i "tx"
```

### Log Aggregation

For production, use log aggregation:

```yaml
# fluent-bit config
[INPUT]
    Name tail
    Path /var/log/virtengine/*.log
    Parser json

[FILTER]
    Name grep
    Match *
    Regex level error|warn

[OUTPUT]
    Name elasticsearch
    Match *
    Host elasticsearch
    Port 9200
    Index virtengine-logs
```

---

## Performance Debugging

### CPU Profiling

```go
import _ "net/http/pprof"

func main() {
    // Enable pprof server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    // ... rest of application
}
```

```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Commands in pprof:
# top          - Show top functions
# top -cum     - Show by cumulative time
# list FuncName - Show annotated source
# web          - Open in browser
```

### Memory Profiling

```bash
# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Allocs profile
go tool pprof http://localhost:6060/debug/pprof/allocs

# Compare profiles
go tool pprof -base profile1.prof profile2.prof
```

### Goroutine Analysis

```bash
# Get goroutine dump
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Analyze in pprof
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Benchmark Debugging

```bash
# Run benchmarks
go test -bench=. -benchmem ./x/market/...

# With CPU profile
go test -bench=. -cpuprofile=cpu.prof ./x/market/...

# With memory profile
go test -bench=. -memprofile=mem.prof ./x/market/...

# Analyze profile
go tool pprof cpu.prof
```

### Trace Analysis

```bash
# Collect trace
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# View trace
go tool trace trace.out
```

---

## Quick Reference

### Debug Commands Cheatsheet

```bash
# Tests
go test -v -run TestName ./path/...     # Single test
dlv test ./path/... -- -test.run Test   # Debug test

# Chain
virtengine start --log_level debug      # Debug logging
virtengine query tx <hash>              # Inspect tx
grpcurl -plaintext localhost:9090 list  # List gRPC

# Provider
curl localhost:8443/health              # Health check
kubectl logs -f deploy/provider-daemon  # K8s logs

# Profiling
go tool pprof http://localhost:6060/debug/pprof/profile
go tool trace trace.out
```

### Environment Variables

```bash
# Enable verbose logging
export COSMOS_SDK_LOG_LEVEL=debug

# Enable goroutine tracebacks
export GOTRACEBACK=all

# Enable race detector
export GORACE="log_path=/tmp/race.log"

# Profile memory allocations
export GODEBUG=allocfreetrace=1
```

---

## Related Documentation

- [Testing Guide](./04-testing-guide.md) - Test debugging
- [Patterns & Anti-patterns](./07-patterns-antipatterns.md) - Common issues
- [Architecture Overview](./02-architecture-overview.md) - System understanding
- [SLOs and Playbooks](../slos-and-playbooks.md) - Production debugging
