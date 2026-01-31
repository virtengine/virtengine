# Marketplace Operations Training

**Module Duration:** 4 hours  
**Prerequisites:** Provider Daemon Training  
**Track:** Provider Operator

---

## Learning Objectives

By the end of this module, you will be able to:

- [ ] Explain the order/bid/lease lifecycle
- [ ] Configure bid engine pricing strategies
- [ ] Monitor and manage active leases
- [ ] Handle escrow and settlement operations
- [ ] Troubleshoot marketplace issues

---

## Table of Contents

1. [Marketplace Overview](#marketplace-overview)
2. [Order Lifecycle](#order-lifecycle)
3. [Bid Engine Configuration](#bid-engine-configuration)
4. [Lease Management](#lease-management)
5. [Escrow and Settlement](#escrow-and-settlement)
6. [Monitoring and Metrics](#monitoring-and-metrics)
7. [Troubleshooting](#troubleshooting)
8. [Exercises](#exercises)

---

## Marketplace Overview

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  VirtEngine Marketplace                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Tenant                    Chain                    Provider   │
│   ┌─────────┐              ┌─────────┐              ┌─────────┐│
│   │ SDL     │──Create─────▶│  Order  │◀────Watch────│ Bid     ││
│   │ Deploy  │  Order       │  Book   │              │ Engine  ││
│   └─────────┘              └────┬────┘              └────┬────┘│
│                                 │                        │      │
│   ┌─────────┐              ┌────▼────┐              ┌────▼────┐│
│   │ Accept  │◀─────────────│  Bids   │◀─────────────│ Submit  ││
│   │ Bid     │              │         │              │ Bid     ││
│   └────┬────┘              └─────────┘              └─────────┘│
│        │                                                        │
│        │                   ┌─────────┐                         │
│        └──────────────────▶│  Lease  │─────────────────────────▶│
│                            │ Created │           Deploy         │
│                            └────┬────┘           Workload       │
│                                 │                               │
│                            ┌────▼────┐                         │
│                            │ Escrow  │                         │
│                            │ Account │                         │
│                            └─────────┘                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Components

| Component | Purpose | Location |
|-----------|---------|----------|
| **Order** | Resource request from tenant | `x/market/types/order.go` |
| **Bid** | Provider's price quote | `x/market/types/bid.go` |
| **Lease** | Active deployment agreement | `x/market/types/lease.go` |
| **Escrow** | Payment locking and settlement | `x/escrow/` |

---

## Order Lifecycle

### Order States

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Open    │────▶│  Matched │────▶│  Active  │────▶│  Closed  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                                                   ▲
     └───────────────────────────────────────────────────┘
                        (Timeout/Cancel)
```

### Order Creation (Tenant Side)

```bash
# Create deployment from SDL
virtengine tx deployment create deployment.yaml \
    --from tenant \
    --keyring-backend file

# Query orders
virtengine query market orders --owner $(virtengine keys show tenant -a)
```

### Order Fields

| Field | Description |
|-------|-------------|
| `OrderID` | Unique identifier (owner/dseq/gseq/oseq) |
| `State` | Open, Matched, Active, Closed |
| `Spec` | Resource requirements |
| `CreatedAt` | Block height of creation |

---

## Bid Engine Configuration

### Configuration File

```yaml
# provider-config.yaml
bid_engine:
  # Enable/disable bidding
  enabled: true
  
  # Minimum profit margin (percentage)
  min_profit_margin: 10
  
  # Maximum concurrent bids
  max_concurrent_bids: 50
  
  # Bid timeout (duration)
  bid_timeout: 5m
  
  # Resource pricing (per unit per block)
  pricing:
    cpu_millicores: 0.000001uve
    memory_mb: 0.0000001uve
    storage_mb: 0.00000001uve
    gpu: 0.001uve
```

### Pricing Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                  Bid Price Calculation                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Base Cost Calculation:                                        │
│   ─────────────────────                                         │
│   base_cost = (cpu_units × cpu_price) +                        │
│               (memory_units × memory_price) +                   │
│               (storage_units × storage_price) +                 │
│               (gpu_units × gpu_price)                           │
│                                                                 │
│   Profit Margin:                                                │
│   ──────────────                                                │
│   bid_price = base_cost × (1 + profit_margin)                  │
│                                                                 │
│   Example:                                                      │
│   ─────────                                                     │
│   Resources: 1000m CPU, 2048MB RAM, 10GB storage               │
│   Base cost: (1000 × 0.000001) + (2048 × 0.0000001) +          │
│              (10240 × 0.00000001) = 0.001307 uve/block         │
│   With 10% margin: 0.001437 uve/block                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Bidding Filters

```yaml
# Only bid on certain orders
filters:
  # Minimum order value
  min_order_value: 0.01uve
  
  # Maximum resource request
  max_cpu: 64000      # 64 cores
  max_memory: 262144  # 256 GB
  max_storage: 10485760  # 10 TB
  
  # Allowed regions (if specified in order)
  allowed_regions:
    - us-east-1
    - eu-west-1
  
  # Excluded attributes
  excluded_attributes:
    - gpu-type=nvidia-a100  # Don't have this GPU
```

---

## Lease Management

### Lease States

| State | Description | Transitions |
|-------|-------------|-------------|
| `Active` | Workload running | → Closed |
| `InsufficientFunds` | Escrow depleted | → Active, Closed |
| `Closed` | Lease terminated | Final |

### Managing Active Leases

```bash
# List all leases for provider
virtengine query market leases --provider $(virtengine keys show provider -a)

# Get specific lease details
virtengine query market lease \
    --owner <tenant-address> \
    --dseq <deployment-seq> \
    --gseq <group-seq> \
    --oseq <order-seq> \
    --provider <provider-address>

# Close a lease (provider side)
virtengine tx market lease-close \
    --dseq <deployment-seq> \
    --gseq <group-seq> \
    --oseq <order-seq> \
    --from provider
```

### Lease Monitoring

Key metrics to monitor:

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `virtengine_provider_active_leases` | Count of active leases | Warning: > 90% capacity |
| `virtengine_provider_lease_revenue` | Revenue from leases | Info only |
| `virtengine_provider_lease_close_rate` | Lease termination rate | Warning: > 10%/hour |
| `virtengine_escrow_balance` | Remaining funds per lease | Warning: < 24h remaining |

---

## Escrow and Settlement

### Escrow Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      Escrow Flow                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. Lease Created                                              │
│      ┌─────────┐        ┌─────────┐                            │
│      │ Tenant  │───────▶│ Escrow  │  Funds locked              │
│      │ Account │ Deposit│ Account │                            │
│      └─────────┘        └────┬────┘                            │
│                              │                                  │
│   2. Usage Submitted                                            │
│      ┌─────────┐             │                                  │
│      │ Usage   │─────────────┤                                  │
│      │ Record  │  Metrics    │                                  │
│      └─────────┘             │                                  │
│                              │                                  │
│   3. Settlement (Periodic)                                      │
│                         ┌────▼────┐                            │
│                         │ Settle  │                            │
│                         │ Payment │                            │
│                         └────┬────┘                            │
│                              │                                  │
│      ┌─────────┐        ┌────▼────┐                            │
│      │ Provider│◀───────│ Payment │  Funds released            │
│      │ Account │        │ Release │                            │
│      └─────────┘        └─────────┘                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### ResourceMetrics

Usage is tracked using ResourceMetrics:

```go
type ResourceMetrics struct {
    CPUMilliSeconds    int64  // CPU usage in milliseconds
    MemoryByteSeconds  int64  // Memory usage in byte-seconds
    StorageByteSeconds int64  // Storage usage in byte-seconds
    GPUSeconds         int64  // GPU usage in seconds
}
```

### Submitting Usage Records

```bash
# Usage is submitted automatically by provider daemon
# Manual submission (for testing):
virtengine tx escrow submit-usage \
    --lease-id <lease-id> \
    --cpu-milliseconds 3600000 \
    --memory-byte-seconds 8589934592000 \
    --storage-byte-seconds 107374182400000 \
    --gpu-seconds 0 \
    --from provider
```

### Settlement Process

Settlement occurs:
- **Periodically**: Every N blocks (configurable)
- **On Lease Close**: Final settlement
- **On Insufficient Funds**: Remaining balance distributed

```bash
# Query escrow balance
virtengine query escrow account --account-id <escrow-id>

# Force settlement (governance action)
virtengine tx gov submit-proposal escrow-settlement \
    --escrow-id <escrow-id> \
    --from validator
```

---

## Monitoring and Metrics

### Key Metrics Dashboard

```
┌─────────────────────────────────────────────────────────────────┐
│                  Marketplace Metrics                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Active Leases: 47/50 (94%)    Revenue: 1,234.56 UVE/day      │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░        ▲ +12% from yesterday         │
│                                                                 │
│   Bid Success Rate: 78%         Avg Bid Time: 2.3s             │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░        Target: < 5s ✓                │
│                                                                 │
│   Order Fill Rate: 85%          Escrow Warnings: 3             │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░          Low balance alerts            │
│                                                                 │
│   Lease Close Reasons:                                          │
│   ├─ Tenant Terminated: 45%                                    │
│   ├─ Lease Completed: 40%                                      │
│   ├─ Insufficient Funds: 10%                                   │
│   └─ Provider Closed: 5%                                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Prometheus Metrics

```promql
# Active leases
virtengine_market_active_leases{provider="..."}

# Bid success rate
rate(virtengine_bid_success_total[1h]) / rate(virtengine_bid_total[1h])

# Revenue rate
sum(rate(virtengine_escrow_settlement_amount[1d]))

# Low escrow balance warning
virtengine_escrow_balance < virtengine_lease_rate * 86400
```

---

## Troubleshooting

### Common Issues

| Issue | Symptoms | Resolution |
|-------|----------|------------|
| **Bids not submitting** | No bids appearing | Check bid engine logs, verify connectivity |
| **Low bid success** | Bids rejected | Adjust pricing, check filters |
| **Lease failures** | Workloads not starting | Check infrastructure adapters |
| **Settlement delays** | Payments late | Verify on-chain connectivity |

### Diagnostic Commands

```bash
# Check bid engine status
curl -s localhost:8443/api/v1/status | jq .

# List pending bids
virtengine query market bids --provider $(virtengine keys show provider -a) --state open

# Check escrow health
virtengine query escrow accounts --provider $(virtengine keys show provider -a)

# View bid engine logs
journalctl -u provider-daemon -n 100 | grep -i bid
```

### Bid Rejection Reasons

| Code | Reason | Action |
|------|--------|--------|
| `insufficient_resources` | Provider lacks capacity | Scale infrastructure |
| `price_too_high` | Bid exceeds order max | Lower pricing |
| `invalid_attributes` | Attribute mismatch | Update provider attributes |
| `provider_blocked` | Provider on blocklist | Contact governance |

---

## Exercises

### Exercise 1: Configure Bid Pricing

Create a pricing configuration that:
- Sets CPU at 0.000002 uve/millicores/block
- Sets memory at 0.0000002 uve/MB/block
- Adds 15% profit margin
- Filters out orders requiring > 32 cores

### Exercise 2: Monitor Lease Health

Set up alerts for:
- Active leases > 80% capacity
- Escrow balance < 24 hours remaining
- Bid success rate < 50%

### Exercise 3: Troubleshoot Low Revenue

Given scenario:
- Bid success rate is 90%
- But revenue is below target

Investigate and propose solutions.

---

## Key Takeaways

1. **Order → Bid → Lease** is the core marketplace flow
2. **Pricing strategy** directly impacts profitability
3. **Escrow** protects both tenant and provider
4. **Usage metrics** determine settlement amounts
5. **Monitor bid success rate** and adjust pricing
6. **Settlement is automatic** but monitor for issues

---

**Document Owner**: Training Team  
**Last Updated**: 2026-01-31  
**Version**: 1.0.0
