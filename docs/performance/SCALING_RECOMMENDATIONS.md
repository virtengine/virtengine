# VirtEngine Scaling Recommendations

**Version:** 1.0.0
**Date:** 2026-01-30
**Task Reference:** SCALE-001 - Load Testing - 1M Nodes Simulation

---

## Overview

This document provides scaling recommendations based on load testing and degradation analysis performed against VirtEngine's chain modules, provider daemon, and state management systems. Results are derived from the benchmark and scale test suites in `tests/benchmark/` and `tests/load/scale/`.

## Target Scale

| Metric | Target |
|--------|--------|
| Validator set size | 1,000,000 |
| Active marketplace orders | 100,000 |
| Concurrent providers | 1,000+ |
| Transaction throughput | 10,000 TPS |
| Identity verifications | 10/sec |

## Hardware Recommendations

### Small Scale (1,000 validators)

| Resource | Specification |
|----------|---------------|
| CPU | 2 cores |
| Memory | 4 GB |
| Storage | 100 GB SSD |
| Network | 100 Mbps |

Suitable for development, testing, and small testnets.

### Medium Scale (10,000 validators)

| Resource | Specification |
|----------|---------------|
| CPU | 4 cores |
| Memory | 8 GB |
| Storage | 500 GB NVMe |
| Network | 1 Gbps |

Suitable for staging environments and medium-scale testnets.

### Large Scale (100,000 validators)

| Resource | Specification |
|----------|---------------|
| CPU | 8 cores |
| Memory | 32 GB |
| Storage | 2 TB NVMe |
| Network | 10 Gbps |

Suitable for production mainnet with high validator participation.

### Production Scale (1,000,000 validators)

| Resource | Specification |
|----------|---------------|
| CPU | 32+ cores |
| Memory | 128+ GB |
| Storage | 10+ TB NVMe RAID |
| Network | 100 Gbps |

Requires horizontal sharding and distributed state sync.

## Performance Baselines

### Transaction Throughput

| Metric | Baseline | Notes |
|--------|----------|-------|
| Target TPS | 10,000 | Parallel transaction processing |
| Min TPS | 5,000 | CI failure threshold |
| P95 latency | 100ms | Per-transaction latency |
| P99 latency | 250ms | Tail latency |
| Error rate | < 0.1% | Acceptable error rate |

See `tests/benchmark/throughput_benchmark_test.go` for implementation.

### Validator Set Operations

| Operation | Target (1M) | Critical |
|-----------|-------------|----------|
| Lookup P95 | < 10ms | < 50ms |
| Lookup P99 | < 50ms | < 100ms |
| Full iteration | < 30s | < 60s |
| Consensus round | < 5s | < 10s |
| Voting power calc | < 100ms | < 500ms |
| Memory per validator | < 1 KB | < 2 KB |

Validator lookup uses O(1) hash map access. Iteration is O(n) linear. Voting power calculation is O(n) with potential for parallel reduction at large scales.

### Marketplace Operations

| Operation | Target (100k orders) | Critical |
|-----------|---------------------|----------|
| Order creation P95 | < 50ms | < 100ms |
| Bid submission P95 | < 20ms | < 50ms |
| Order matching P95 | < 100ms | < 200ms |
| Iteration rate | > 100k/sec | > 50k/sec |
| Max memory | < 4 GB | < 8 GB |

Marketplace uses indexed stores (byOwner, byStatus, byOrder, byProvider) for efficient lookups. Order matching can become a bottleneck at high bid counts per order.

### Provider Daemon

| Operation | Target (1k providers) | Critical |
|-----------|----------------------|----------|
| Bid latency P95 | < 100ms | < 250ms |
| Event processing P95 | < 50ms | < 100ms |
| Min bids/sec | > 10k | > 5k |
| Max event backlog | < 10k | < 50k |

Provider daemon uses event-driven architecture with per-provider event queues. Backpressure handling drops events when queues are full.

### State Sync

| Operation | Target | Critical |
|-----------|--------|----------|
| Snapshot creation rate | > 100 MB/s | > 50 MB/s |
| Chunk transfer rate | > 50 MB/s | > 25 MB/s |
| State apply rate | > 50k entries/s | > 25k entries/s |
| Max sync time | < 5 min | < 10 min |

State sync uses chunked snapshots (16 KB chunks) for parallel transfer.

### Network Partition Recovery

| Metric | Target | Critical |
|--------|--------|----------|
| Recovery time | < 30s | < 60s |
| State reconcile | < 60s | < 120s |
| Message loss rate | < 1% | < 5% |
| Consensus rounds lost | < 5 | < 10 |

## Identified Bottlenecks

### 1. Validator Set Iteration (Linear)

**Impact:** Full iteration over 1M validators takes significant time.

**Mitigation:**
- Use indexed subsets (active/bonded validators only) for frequent operations
- Cache voting power calculations
- Consider iterator pagination for non-critical queries

### 2. Marketplace Order Matching (Bid Count Dependent)

**Impact:** Order matching performance degrades with many bids per order.

**Mitigation:**
- Limit active bids per order (configurable max)
- Use priority queue for bid selection
- Consider batch matching for high-volume periods

### 3. State Snapshot Size (Memory Bound)

**Impact:** Large state snapshots require significant memory during creation.

**Mitigation:**
- Implement streaming snapshot creation (avoid full materialization)
- Use incremental snapshots between heights
- Compress chunks before transfer

### 4. Provider Event Broadcasting (Fan-Out)

**Impact:** Broadcasting events to 1k+ providers creates O(n) message overhead.

**Mitigation:**
- Implement pub/sub with topic-based filtering
- Use WebSocket multiplexing for event delivery
- Allow providers to subscribe to relevant order types only

### 5. Memory Growth at Scale

**Impact:** Memory usage grows linearly with validator count and state size.

**Mitigation:**
- Monitor per-entry memory cost (target < 1 KB)
- Set Go GC target (`GOGC`) based on available memory
- Use memory budgets: 4 GB heap, 10M objects, 100K goroutines

## Architecture Recommendations

### Short Term (Current Architecture)

1. **Optimize hot paths** — Profile keeper methods under load, optimize the top 5 by CPU time
2. **Add caching** — Cache voting power totals, active validator sets, and frequently queried orders
3. **Tune GC** — Set `GOGC=50` for large-state nodes to reduce GC pause times
4. **Connection pooling** — Use pooled gRPC connections for provider daemon communications

### Medium Term (10k-100k Validators)

1. **Indexed state queries** — Add secondary indexes for common query patterns
2. **Parallel consensus** — Implement parallel block proposal validation
3. **Incremental snapshots** — Only sync state deltas between heights
4. **Provider daemon sharding** — Partition providers by region or resource type

### Long Term (1M+ Validators)

1. **Horizontal sharding** — Shard validator set across multiple state partitions
2. **Distributed state sync** — Use distributed hash tables for state discovery
3. **Validator rotation** — Limit active set size with stake-weighted selection
4. **Layer 2 scaling** — Offload HPC job scheduling to purpose-built sidechains

## Degradation Thresholds

| Level | Factor | Action |
|-------|--------|--------|
| Acceptable | < 2x baseline | No action needed |
| Warning | 2x - 5x baseline | Monitor, plan optimization |
| High | 5x - 10x baseline | Prioritize optimization |
| Critical | > 10x baseline | Immediate action required |

Use `TestComprehensiveDegradationReport` in `tests/load/scale/degradation_analysis_test.go` to generate current degradation metrics.

## Running Performance Tests

```bash
# Quick validation (CI)
go test -v ./tests/load/scale/... -short

# Full scale tests
go test -v ./tests/load/scale/... -timeout 30m

# Benchmarks
go test -bench=. -benchtime=10s ./tests/load/scale/...

# Throughput benchmarks
go test -bench=. -benchtime=30s ./tests/benchmark/...

# Degradation report
go test -v ./tests/load/scale/... -run TestComprehensiveDegradationReport -timeout 10m

# k6 load tests (requires running chain)
k6 run tests/load/k6/identity_burst.js
k6 run tests/load/k6/marketplace_burst.js
k6 run tests/load/k6/hpc_burst.js
```

## References

- `tests/benchmark/` — Transaction throughput and baseline framework
- `tests/load/scale/` — Scale tests and degradation analysis
- `tests/load/k6/` — k6 JavaScript load tests
- `tests/load/scenarios_test.go` — Go-based load scenarios
- `pkg/benchmark/` — Memory profiling utilities
- `tests/load/scale/SCALE_BASELINES.json` — Baseline metrics JSON
- `docs/performance/BENCHMARKS.md` — Benchmark documentation
