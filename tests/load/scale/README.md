# VirtEngine Scale Testing Suite

**Version:** 1.0.0  
**Date:** 2026-01-30  
**Task Reference:** SCALE-001 - Load Testing - 1M Nodes Simulation

---

## Overview

This directory contains scale testing scenarios for VirtEngine, designed to validate system performance at production scale including:

- **1M validators simulation**
- **100k active marketplace orders**
- **1k+ concurrent providers**
- **Network partition recovery**
- **State sync at scale**
- **Performance degradation analysis**

## Test Files

| File | Description |
|------|-------------|
| `validator_scale_test.go` | 1M validator set simulation and consensus testing |
| `marketplace_scale_test.go` | 100k orders, bid processing, lease lifecycle |
| `provider_stress_test.go` | 1k+ provider concurrent bidding and event processing |
| `network_partition_test.go` | Network partition scenarios and recovery testing |
| `state_sync_test.go` | State snapshot creation and sync at scale |
| `degradation_analysis_test.go` | Performance bottleneck identification |
| `SCALE_BASELINES.json` | Baseline metrics and thresholds |

## Running Scale Tests

### Quick Validation (CI Mode)

```bash
# Run all scale tests with reduced scale (faster)
go test -v ./tests/load/scale/... -short

# Run specific test suite
go test -v ./tests/load/scale/... -run TestValidatorScaleBaseline -short
```

### Full Scale Testing

```bash
# Run all scale tests (may take 10+ minutes)
go test -v ./tests/load/scale/... -timeout 30m

# Run with verbose output and race detection
go test -v -race ./tests/load/scale/... -timeout 30m

# Run benchmarks
go test -bench=. -benchtime=10s ./tests/load/scale/...

# Run specific scale (override via environment)
SCALE_TEST_VALIDATORS=1000000 go test -v ./tests/load/scale/... -run TestValidatorScaleBaseline
```

### Parallel Benchmarks

```bash
# Run parallel benchmarks with custom CPU count
go test -bench=Parallel -cpu=1,2,4,8,16 ./tests/load/scale/...
```

## Performance Baselines

### Validator Operations (Target: 1M validators)

| Metric | Target | Critical |
|--------|--------|----------|
| Lookup P95 latency | < 10ms | < 50ms |
| Lookup P99 latency | < 50ms | < 100ms |
| Full iteration | < 30s | < 60s |
| Consensus round | < 5s | < 10s |
| Voting power calc | < 100ms | < 500ms |
| Memory per validator | < 1KB | < 2KB |

### Marketplace Operations (Target: 100k orders)

| Metric | Target | Critical |
|--------|--------|----------|
| Order creation P95 | < 50ms | < 100ms |
| Bid submission P95 | < 20ms | < 50ms |
| Order matching P95 | < 100ms | < 200ms |
| Order iteration rate | > 100k/sec | > 50k/sec |
| Max memory | < 4GB | < 8GB |

### Provider Daemon (Target: 1k+ providers)

| Metric | Target | Critical |
|--------|--------|----------|
| Bid latency P95 | < 100ms | < 250ms |
| Event processing P95 | < 50ms | < 100ms |
| Min bids/sec | > 10k | > 5k |
| Max event backlog | < 10k | < 50k |

### Network Partition Recovery

| Metric | Target | Critical |
|--------|--------|----------|
| Recovery time | < 30s | < 60s |
| State reconcile | < 60s | < 120s |
| Message loss rate | < 1% | < 5% |
| Consensus rounds lost | < 5 | < 10 |

### State Sync

| Metric | Target | Critical |
|--------|--------|----------|
| Snapshot creation rate | > 100MB/s | > 50MB/s |
| Chunk transfer rate | > 50MB/s | > 25MB/s |
| State apply rate | > 50k entries/s | > 25k entries/s |
| Max sync time | < 5min | < 10min |

## Degradation Thresholds

| Level | Factor | Action |
|-------|--------|--------|
| Acceptable | < 2x | No action needed |
| Warning | 2x - 5x | Monitor and plan optimization |
| High | 5x - 10x | Prioritize optimization |
| Critical | > 10x | Immediate action required |

## Scaling Recommendations

### Small Scale (1,000 validators)

- CPU: 2 cores
- Memory: 4 GB
- Storage: 100 GB SSD
- Network: 100 Mbps

### Medium Scale (10,000 validators)

- CPU: 4 cores
- Memory: 8 GB
- Storage: 500 GB NVMe
- Network: 1 Gbps

### Large Scale (100,000 validators)

- CPU: 8 cores
- Memory: 32 GB
- Storage: 2 TB NVMe
- Network: 10 Gbps

### Production Scale (1,000,000 validators)

- CPU: 32+ cores
- Memory: 128+ GB
- Storage: 10+ TB NVMe RAID
- Network: 100 Gbps
- **Note:** Requires horizontal sharding and distributed state sync

## Test Scenarios

### Scenario 1: Validator Scale Test

Tests validator set operations at various scales:
1. Store population performance
2. Single validator lookup latency
3. Full set iteration time
4. Voting power calculation
5. Bonded validator retrieval
6. Memory pressure under load

### Scenario 2: Marketplace Scale Test

Tests marketplace operations at 100k order scale:
1. Order creation throughput
2. Bid submission latency
3. Order matching performance
4. Concurrent operations
5. Complete order lifecycle

### Scenario 3: Provider Stress Test

Tests provider daemon under stress:
1. Pool creation and startup
2. Event broadcasting to 1k+ providers
3. Concurrent bidding simulation
4. Backpressure handling
5. Connection pool stress
6. Resource contention

### Scenario 4: Network Partition Test

Tests network partition scenarios:
1. Pre-partition consensus
2. Partition creation (2/3 - 1/3 split)
3. Partitioned consensus behavior
4. Partition healing
5. State reconciliation
6. Chaos testing (random partitions)

### Scenario 5: State Sync Test

Tests state synchronization:
1. State population at scale
2. Snapshot creation performance
3. Snapshot application
4. Incremental sync
5. Concurrent access patterns
6. Crash recovery simulation

### Scenario 6: Degradation Analysis

Comprehensive performance analysis:
1. Multi-scale measurements
2. Scaling factor calculation
3. Bottleneck identification
4. Memory analysis
5. Concurrent load testing
6. Report generation

## CI Integration

Scale tests can be integrated into CI pipelines:

```yaml
# Example GitHub Actions workflow
scale-tests:
  runs-on: ubuntu-latest
  timeout-minutes: 30
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - name: Run Scale Tests
      run: go test -v ./tests/load/scale/... -short -timeout 20m
    - name: Run Scale Benchmarks
      run: go test -bench=. -benchtime=5s ./tests/load/scale/... | tee benchmark.txt
    - name: Upload Results
      uses: actions/upload-artifact@v4
      with:
        name: scale-test-results
        path: benchmark.txt
```

## Interpreting Results

### Scaling Factor

The scaling factor indicates how performance degrades with scale:

- **1.0** = Linear scaling (O(n)) - acceptable
- **< 1.0** = Sub-linear (O(log n)) - excellent
- **2.0** = Quadratic (O(n²)) - needs optimization
- **> 2.0** = Super-linear - critical, requires algorithmic changes

### Memory Scaling

Memory efficiency is measured as:
- Bytes per entry at different scales
- Total heap allocation
- GC pressure indicators

Good memory scaling shows constant bytes-per-entry across scales.

### Bottleneck Categories

1. **CPU-bound**: Look for high goroutine counts, consider parallelization
2. **Memory-bound**: Check allocation patterns, consider pooling
3. **I/O-bound**: Optimize disk/network access patterns
4. **Lock contention**: Profile mutex usage, consider lock-free structures

## Troubleshooting

### Tests Timeout

Reduce scale or increase timeout:
```bash
go test -v ./tests/load/scale/... -timeout 60m
```

### Out of Memory

Run with memory limits:
```bash
GOGC=50 go test -v ./tests/load/scale/... -short
```

### Flaky Results

Increase benchmark time for more stable results:
```bash
go test -bench=. -benchtime=30s ./tests/load/scale/...
```

## Contributing

When adding new scale tests:

1. Define baseline metrics in `SCALE_BASELINES.json`
2. Implement tests with both short-mode (CI) and full-scale modes
3. Include benchmarks for key operations
4. Add scaling analysis where applicable
5. Update this README with new scenarios

## Security Warnings

> ⚠️ Scale tests use significant system resources
> ⚠️ Never run against production systems
> ⚠️ Monitor system resources during tests
> ⚠️ Use dedicated test infrastructure for full-scale runs
