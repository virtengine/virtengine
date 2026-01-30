# VirtEngine Performance Benchmarks

Task Reference: PERF-001 - Performance Benchmarking Suite

This document describes the performance benchmarking suite for VirtEngine, including baseline metrics, running instructions, and CI integration guidelines.

## Overview

The performance benchmarking suite provides comprehensive coverage for all critical paths in VirtEngine:

| Category | Target | Location |
|----------|--------|----------|
| Transaction Throughput | 10k TPS | `tests/benchmark/throughput_benchmark_test.go` |
| VEID Verification Latency | <100ms | `x/veid/keeper/verification_benchmark_test.go` |
| Provider Daemon Bidding | <50ms | `pkg/provider_daemon/bid_engine_benchmark_test.go` |
| ZK Proof Generation | Documented | `x/veid/keeper/zkproofs_benchmark_test.go` |
| Memory Profiling | See budgets | `pkg/benchmark/memory_profile.go` |

## Baseline Metrics

All baseline metrics are defined in `tests/benchmark/BASELINE_METRICS.json`. These serve as the reference for regression detection.

### Transaction Throughput Baselines

| Metric | Baseline | Description |
|--------|----------|-------------|
| Target TPS | 10,000 | Target transactions per second |
| Min Acceptable TPS | 5,000 | Minimum before failing CI |
| P95 Latency | 100ms | 95th percentile transaction latency |
| P99 Latency | 250ms | 99th percentile transaction latency |
| Max Error Rate | 0.1% | Maximum acceptable error rate |

#### Component-Level Baselines
| Operation | Baseline | Notes |
|-----------|----------|-------|
| Hash Computation | 1µs | SHA-256 for transaction hash |
| Validation | 10µs | Signature verification |
| Execution | 50µs | State transition execution |
| State Write | 5µs | IAVL store write |

### VEID Verification Latency Baselines

| Metric | Baseline | Description |
|--------|----------|-------------|
| Target Latency | 100ms | Target verification time |
| P95 Latency | 150ms | 95th percentile latency |
| P99 Latency | 250ms | 99th percentile latency |
| Min Throughput | 10/sec | Minimum verifications per second |

#### Component-Level Baselines
| Operation | Baseline | Notes |
|-----------|----------|-------|
| Scope Decryption | 30ms | X25519-XSalsa20-Poly1305 |
| ML Scoring | 50ms | TensorFlow inference |
| State Update | 10ms | Identity record update |
| Identity Record Create | 100µs | Initial record creation |
| Identity Record Get | 10µs | Record retrieval |
| Score Update | 50µs | Score modification |

### Provider Daemon Bidding Latency Baselines

| Metric | Baseline | Description |
|--------|----------|-------------|
| Target Latency | 50ms | Target bid processing time |
| P95 Latency | 100ms | 95th percentile latency |
| P99 Latency | 200ms | 99th percentile latency |
| Min Bids/sec | 20 | Minimum bidding throughput |

#### Component-Level Baselines
| Operation | Baseline | Notes |
|-----------|----------|-------|
| Order Matching | 10ms | Capacity/capability matching |
| Price Calculation | 5ms | Resource pricing computation |
| Bid Signing | 20ms | Cryptographic signing |
| Rate Limiter Check | 1µs | Rate limit evaluation |

### ZK Proof Generation Baselines

| Proof Type | Baseline | Notes |
|------------|----------|-------|
| Age Proof | 500ms | Proves age over threshold |
| Residency Proof | 500ms | Proves country of residence |
| Score Threshold | 500ms | Proves score above threshold |
| Selective Disclosure | 1000ms | Multi-claim disclosure |
| Proof Verification | 100ms | Verifier-side validation |
| Circuit Compilation | 5s | One-time setup cost |

#### Auxiliary Operations
| Operation | Baseline |
|-----------|----------|
| Commitment Generation | 5µs |
| Nonce Generation | 1µs |
| Proof ID Generation | 2µs |

### Memory Baselines

| Metric | Baseline | Description |
|--------|----------|-------------|
| Max Heap Allocation | 4GB | Maximum heap size |
| Max Heap Objects | 10M | Maximum allocated objects |
| Max Goroutines | 100K | Maximum concurrent goroutines |
| Max GC Pause | 100ms | Maximum stop-the-world pause |
| Max Leak Growth | 10% | Maximum heap growth rate |
| Idle Heap | 256MB | Expected idle memory |
| Per Verification | 64KB | Memory per verification |
| Per Transaction | 8KB | Memory per transaction |

## Running Benchmarks

### All Benchmarks

```bash
# Run all benchmarks with memory stats
go test -bench=. -benchmem ./tests/benchmark/...
go test -bench=. -benchmem ./x/veid/keeper/...
go test -bench=. -benchmem ./pkg/provider_daemon/...
go test -bench=. -benchmem ./pkg/benchmark/...
```

### Transaction Throughput

```bash
# Run transaction throughput benchmarks
go test -bench=BenchmarkTransaction -benchmem ./tests/benchmark/...

# Run baseline test
go test -run=TestTransactionThroughputBaseline ./tests/benchmark/...
```

### VEID Verification

```bash
# Run verification latency benchmarks
go test -bench=BenchmarkVerification -benchmem ./x/veid/keeper/...
go test -bench=BenchmarkFull -benchmem ./x/veid/keeper/...

# Run baseline test
go test -run=TestVerificationLatencyBaseline ./x/veid/keeper/...
```

### Provider Daemon Bidding

```bash
# Run bidding latency benchmarks
go test -bench=BenchmarkBid -benchmem ./pkg/provider_daemon/...
go test -bench=BenchmarkOrderMatching -benchmem ./pkg/provider_daemon/...

# Run baseline test
go test -run=TestBiddingLatencyBaseline ./pkg/provider_daemon/...
```

### ZK Proof Generation

```bash
# Run ZK proof benchmarks
go test -bench=BenchmarkAge -benchmem ./x/veid/keeper/...
go test -bench=BenchmarkResidency -benchmem ./x/veid/keeper/...
go test -bench=BenchmarkProof -benchmem ./x/veid/keeper/...
```

### Memory Profiling

```bash
# Run memory profiler tests
go test -run=TestMemoryProfiler ./pkg/benchmark/...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmark/...

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./tests/benchmark/...
```

## CI Integration

### Regression Detection

The benchmark suite includes automatic regression detection. Use the `BenchmarkComparator` to compare results against baselines:

```go
package main

import (
    "os"
    "github.com/virtengine/virtengine/tests/benchmark"
)

func main() {
    // Load baselines
    baselines, err := benchmark.LoadBaselines("tests/benchmark/BASELINE_METRICS.json")
    if err != nil {
        panic(err)
    }
    
    // Create comparator with 10% warn and 20% fail thresholds
    comparator := benchmark.NewBenchmarkComparator(*baselines, 10.0, 20.0)
    
    // Compare your benchmark results
    results := []benchmark.BenchmarkResult{
        {Name: "hash_compute", Category: "transaction", NsPerOp: measuredNs},
        // ... more results
    }
    
    // Generate report
    report := comparator.GenerateReport(results)
    
    // Check status
    if report.Summary.OverallStatus == "fail" {
        os.Exit(1)
    }
}
```

### GitHub Actions Integration

```yaml
name: Performance Benchmarks

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run Benchmarks
        run: |
          go test -bench=. -benchmem -json ./tests/benchmark/... > benchmark_results.json
          go test -bench=. -benchmem -json ./x/veid/keeper/... >> benchmark_results.json
          go test -bench=. -benchmem -json ./pkg/provider_daemon/... >> benchmark_results.json
      
      - name: Check Baseline Compliance
        run: |
          go test -run=TestTransactionThroughputBaseline ./tests/benchmark/...
          go test -run=TestVerificationLatencyBaseline ./x/veid/keeper/...
          go test -run=TestBiddingLatencyBaseline ./pkg/provider_daemon/...
      
      - name: Upload Results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: benchmark_results.json
```

## Benchmark Categories

### Micro-benchmarks

Low-level operations that run in nanoseconds:
- Hash computation
- Serialization/deserialization
- Rate limiter checks
- Commitment generation

### Component Benchmarks

Mid-level operations that run in microseconds to milliseconds:
- Transaction validation
- Order matching
- Price calculation
- Score updates

### Integration Benchmarks

Full pipeline operations:
- Full transaction processing
- Complete verification flow
- End-to-end bidding

### Load Tests

High-concurrency stress tests:
- Parallel transaction processing
- Concurrent verifications
- Burst bidding scenarios

## Memory Analysis

### Using the Memory Profiler

```go
import "github.com/virtengine/virtengine/pkg/benchmark"

func main() {
    // Create profiler
    config := benchmark.MemoryProfilerConfig{
        Interval: 5 * time.Second,
        MaxSize:  1000,
    }
    profiler := benchmark.NewMemoryProfiler(config)
    
    // Start collecting
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    profiler.Start(ctx)
    defer profiler.Stop()
    
    // Run workload...
    
    // Analyze results
    stats := profiler.GetStats()
    fmt.Printf("Avg Heap: %s\n", benchmark.FormatBytes(stats.AvgHeapAlloc))
    fmt.Printf("Max Heap: %s\n", benchmark.FormatBytes(stats.MaxHeapAlloc))
    
    // Check for leaks
    leak := profiler.DetectLeaks(10.0) // 10% threshold
    if leak.SuspectedLeak {
        fmt.Printf("Potential leak: %d%% growth\n", int(leak.GrowthPercent))
    }
    
    // Check budget compliance
    budget := benchmark.DefaultMemoryBudget()
    violations := profiler.CheckBudget(budget)
    for _, v := range violations {
        fmt.Printf("Budget violation: %s\n", v.Type)
    }
}
```

### Memory Budget Enforcement

```go
budget := benchmark.MemoryBudget{
    MaxHeapAlloc:   4 * 1024 * 1024 * 1024, // 4GB
    MaxHeapObjects: 10_000_000,
    MaxGoroutines:  100_000,
    MaxGCPauseNs:   100_000_000, // 100ms
}

violations := profiler.CheckBudget(budget)
if len(violations) > 0 {
    // Handle budget violations
}
```

## Updating Baselines

When performance characteristics change intentionally (e.g., algorithm improvements), update the baselines:

1. Run benchmarks on the reference hardware
2. Update `tests/benchmark/BASELINE_METRICS.json`
3. Document the reason for baseline changes in the PR
4. Ensure all CI checks pass with new baselines

```bash
# Generate new baselines
go test -bench=. -benchmem -json ./... > new_results.json
# Review and update BASELINE_METRICS.json accordingly
```

## Troubleshooting

### High Variance in Results

- Run with `-count=10` to get multiple samples
- Use `-benchtime=5s` for longer measurement windows
- Ensure no other processes are competing for resources

### Memory Leaks Detected

1. Use `go tool pprof` with heap profiles
2. Check goroutine lifecycle management
3. Review resource cleanup in defer statements

### Slow CI Benchmarks

- Consider running only critical benchmarks in CI
- Use `-short` flag to skip long-running tests
- Cache benchmark binaries

## Related Documentation

- [VEID Module Architecture](../architecture/veid.md)
- [Provider Daemon Design](../architecture/provider-daemon.md)
- [ML Inference Determinism](../architecture/ml-inference.md)
- [Testing Guide](../_docs/testing-guide.md)
