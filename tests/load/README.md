# VirtEngine Load Testing Suite

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-801

---

## Overview

This directory contains load testing scenarios for the VirtEngine blockchain platform. These tests verify system performance under load for identity scoring, marketplace operations, and HPC scheduling.

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

## Performance Baselines

| Metric | Target | Critical |
|--------|--------|----------|
| Identity verification p95 latency | < 5s | < 10s |
| Order provisioning p95 latency | < 30s | < 60s |
| HPC job scheduling p95 latency | < 10s | < 30s |
| Daemon event processing lag | < 100ms | < 500ms |
| Chain throughput (TPS) | > 100 | > 50 |

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
