# Settlement Usage Accounting and Rewards

This document describes the usage accounting pipeline in `x/settlement` and how
usage-based rewards are calculated and distributed.

## Overview

1. Provider daemon records usage metrics per allocation/job.
2. Usage reports are submitted on-chain as `MsgRecordUsage` events.
3. Usage records are settled into invoices and payouts during `SettleOrder`.
4. Usage rewards are calculated deterministically from settled usage records and
   stored as claimable rewards.

## Usage Reporting (Provider Daemon)

The provider daemon emits **per-resource** usage reports for a metering window:

| Usage Type | Units | Notes |
| --- | --- | --- |
| `cpu` | CPU-hours | `CPUMilliSeconds / (1000 * 3600)` |
| `memory` | GB-hours | `MemoryByteSeconds / (1024^3 * 3600)` |
| `storage` | GB-hours | `StorageByteSeconds / (1024^3 * 3600)` |
| `gpu` | GPU-hours | `GPUSeconds / 3600` |
| `network` | GB | `(NetworkBytesIn + NetworkBytesOut) / 1024^3` |

If all usage units are zero, the daemon emits a single `cpu` report with `1`
unit to ensure a minimum accounting footprint.

## Usage Accounting (On-Chain)

Each `MsgRecordUsage` creates a `UsageRecord` with:

- `UsageUnits`, `UsageType`, `UnitPrice`, `TotalCost`
- `PeriodStart`, `PeriodEnd`, `SubmittedAt`
- `CustomerAcknowledged` (via `MsgAcknowledgeUsage`)

Usage records are settled by `SettleOrder`, which:

- Creates a settlement record and invoice (when billing is enabled).
- Executes payout records.
- Triggers usage reward distribution for the settlement.

## Deterministic Reward Calculation

Usage rewards are computed per **usage record** (per-job, per-resource). The
reward formula is deterministic and uses fixed-point math:

```
reward = total_cost
  × usage_reward_rate_bps
  × resource_multiplier_bps
  × sla_multiplier_bps
  × acknowledgement_multiplier_bps
  / 10_000^4
```

### SLA Modifier

`sla_multiplier_bps` is derived from report timeliness:

- **On time**: `SubmittedAt <= PeriodEnd + UsageGracePeriod`
- **Late**: reports submitted after the grace period

### Quality Modifier

`acknowledgement_multiplier_bps` is derived from customer acknowledgement:

- **Acknowledged**: `CustomerAcknowledged = true`
- **Unacknowledged**: `CustomerAcknowledged = false`

### Default Parameters (excerpt)

- `usage_reward_rate_bps`: `1000` (10%)
- `usage_reward_cpu_multiplier_bps`: `10000` (1.0x)
- `usage_reward_gpu_multiplier_bps`: `12000` (1.2x)
- `usage_reward_network_multiplier_bps`: `9000` (0.9x)
- `usage_reward_sla_ontime_multiplier_bps`: `10000` (1.0x)
- `usage_reward_sla_late_multiplier_bps`: `8000` (0.8x)
- `usage_reward_ack_multiplier_bps`: `10000` (1.0x)
- `usage_reward_unack_multiplier_bps`: `9000` (0.9x)

## Queries

### Usage Summary

Returns aggregated usage units and cost by order/provider and time range:

- `QueryUsageSummaryRequest`
- `QueryUsageSummaryResponse`

### Reward History

Returns reward distributions for an address, optionally filtered by source:

- `QueryRewardHistoryRequest`
- `QueryRewardHistoryResponse`

### Payouts

Payout execution records are queryable via:

- `QueryPayoutRequest`
- `QueryPayoutsByProviderRequest`

## Notes

- Usage rewards are added as **claimable rewards**, not transferred immediately.
- Payouts remain the primary settlement transfer to providers.
- Reward history is derived from distributions stored on-chain.
