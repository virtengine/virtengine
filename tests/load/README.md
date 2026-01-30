# VirtEngine Load Testing Suite

**Version:** 1.1.0  
**Date:** 2026-01-30  
**Task Reference:** VE-801, SCALE-001

---

## Overview

This directory contains load testing scenarios for the VirtEngine blockchain platform. These tests verify system performance under load for identity scoring, marketplace operations, HPC scheduling, and **production-scale simulations**.

## Directory Structure

```
tests/load/
├── k6/                    # k6 JavaScript load tests
├── scale/                 # SCALE-001: Production scale tests (1M validators)
│   ├── validator_scale_test.go
│   ├── marketplace_scale_test.go
│   ├── provider_stress_test.go
│   ├── network_partition_test.go
│   ├── state_sync_test.go
│   ├── degradation_analysis_test.go
│   └── README.md
├── scenarios_test.go      # Standard load scenarios
└── README.md              # This file
```

## Test Scenarios

### Scenario A: Identity Scope Upload Burst
Tests the identity verification pipeline under high load.
- Burst of identity scope uploads
- Pending → finalized scoring transitions
- Chain stability under load

### Scenario B: Marketplace Order Burst
Tests marketplace operations under high transaction volume.
- Burst of order creations
- Provider bid submissions
- Allocation processing
- Consistent state transitions

### Scenario C: HPC Job Submissions
Tests the SLURM scheduling pipeline under load.
- Job submission bursts
- Queue management
- Accounting updates
- Reward distribution

### Scenario D: Scale Testing (SCALE-001)
Tests production-scale scenarios. See `scale/README.md` for details.
- **1M Validators**: Validator set operations at 1M node scale
- **100k Orders**: Marketplace with 100k concurrent orders
- **1k+ Providers**: Provider daemon stress testing
- **Network Partitions**: Partition and recovery testing
- **State Sync**: Large state snapshot and sync operations
- **Degradation Analysis**: Performance bottleneck identification

## Tools

- **k6**: JavaScript load testing tool for API and chain endpoints
- **Locust**: Python-based load testing for complex scenarios
- **Go Benchmarks**: Native Go benchmarks for internal components

## Running Load Tests

### k6 Tests

```bash
# Install k6
# Windows: winget install k6
# macOS: brew install k6
# Linux: apt-get install k6

# Run identity load test
k6 run tests/load/k6/identity_burst.js

# Run marketplace load test
k6 run tests/load/k6/marketplace_burst.js

# Run HPC load test
k6 run tests/load/k6/hpc_burst.js

# Run with custom options
k6 run --vus 100 --duration 5m tests/load/k6/identity_burst.js
```

### Locust Tests

```bash
# Install locust
pip install locust

# Run identity load test
locust -f tests/load/locust/identity_load.py --host=http://localhost:26657

# Run marketplace load test
locust -f tests/load/locust/marketplace_load.py --host=http://localhost:26657

# Run HPC load test
locust -f tests/load/locust/hpc_load.py --host=http://localhost:26657
```

### Go Benchmarks

```bash
# Run all load benchmarks
go test -bench=. -benchtime=30s ./tests/load/...

# Run specific benchmark
go test -bench=BenchmarkIdentityScoring -benchtime=1m ./tests/load/...
```

### Scale Tests (SCALE-001)

```bash
# Run scale tests (short mode for CI)
go test -v ./tests/load/scale/... -short

# Run full scale tests (production simulation)
go test -v ./tests/load/scale/... -timeout 30m

# Run scale benchmarks
go test -bench=. -benchtime=10s ./tests/load/scale/...

# Run specific scale test
go test -v ./tests/load/scale/... -run TestValidatorScaleBaseline

# Run degradation analysis
go test -v ./tests/load/scale/... -run TestComprehensiveDegradationReport
```

## Performance Baselines

| Metric | Target | Critical |
|--------|--------|----------|
| Identity verification p95 latency | < 5s | < 10s |
| Order provisioning p95 latency | < 30s | < 60s |
| HPC job scheduling p95 latency | < 10s | < 30s |
| Daemon event processing lag | < 100ms | < 500ms |
| Chain throughput (TPS) | > 100 | > 50 |

### Scale Test Baselines (SCALE-001)

| Metric | Target | Critical |
|--------|--------|----------|
| 1M validator lookup P95 | < 10ms | < 50ms |
| 100k order iteration | < 30s | < 60s |
| 1k provider bid latency | < 100ms | < 250ms |
| Partition recovery | < 30s | < 60s |
| State sync rate | > 100MB/s | > 50MB/s |

See `scale/SCALE_BASELINES.json` for complete baseline definitions.

## Resource Limits

| Component | CPU Limit | Memory Limit |
|-----------|-----------|--------------|
| Chain Node | 4 cores | 8 GB |
| Provider Daemon | 2 cores | 4 GB |
| Benchmarking Daemon | 1 core | 2 GB |
| Bridge Service | 1 core | 2 GB |

## CI Integration

Load tests are run as part of CI on release branches:
1. Nightly performance regression tests
2. Pre-release load validation
3. Baseline comparison with previous release

## Results Storage

Load test results are stored in:
- `tests/load/results/` (local)
- Prometheus metrics (production)
- Grafana dashboards (visualization)

## Security Warnings

> ⚠️ Never run load tests against production systems  
> ⚠️ Use dedicated test accounts and keys  
> ⚠️ Monitor resource usage during tests  
> ⚠️ Clean up test data after runs
