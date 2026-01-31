# HPC Billing Rules and Usage Accounting

> VE-5A Implementation Documentation

This document describes the billing rules, usage accounting, and settlement pipeline for HPC (High-Performance Computing) jobs on VirtEngine.

## Overview

The HPC billing system transforms scheduler usage metrics into deterministic billing and provider rewards. It integrates with the escrow module for settlement and supports dispute workflows for usage corrections.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Provider Daemon                                 │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐       │
│  │ HPCAccounting    │  │ HPCReconciliation│  │ Scheduler        │       │
│  │ Service          │  │ Service          │  │ Adapters         │       │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘       │
│           │                     │                      │                 │
│           └─────────────────────┴──────────────────────┘                 │
│                                 │                                        │
└─────────────────────────────────┼────────────────────────────────────────┘
                                  │ Submit to Chain
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          x/hpc Module                                    │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐       │
│  │ Accounting       │  │ Billing          │  │ Settlement       │       │
│  │ Keeper           │  │ Keeper           │  │ Keeper           │       │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘       │
│           │                     │                      │                 │
│           └─────────────────────┴──────────────────────┘                 │
│                                 │                                        │
└─────────────────────────────────┼────────────────────────────────────────┘
                                  │ Settlement
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          x/escrow Module                                 │
│  ┌──────────────────┐  ┌──────────────────┐                              │
│  │ Invoice          │  │ Ledger           │                              │
│  │ Generation       │  │ Records          │                              │
│  └──────────────────┘  └──────────────────┘                              │
└─────────────────────────────────────────────────────────────────────────┘
```

## Usage Metrics Schema

### HPCDetailedMetrics

The core metrics tracked for every HPC job:

| Metric | Type | Description |
|--------|------|-------------|
| `WallClockSeconds` | int64 | Total wall-clock runtime of the job |
| `QueueTimeSeconds` | int64 | Time spent waiting in scheduler queue |
| `CPUCoreSeconds` | int64 | CPU usage (cores × seconds) |
| `MemoryGBSeconds` | int64 | Memory usage (GB × seconds) |
| `GPUSeconds` | int64 | GPU usage (devices × seconds) |
| `GPUType` | string | GPU model (e.g., "nvidia-a100") |
| `StorageGBHours` | float64 | Storage consumed (GB × hours) |
| `NetworkBytesIn` | int64 | Network bytes received |
| `NetworkBytesOut` | int64 | Network bytes sent |
| `IOReadBytes` | int64 | Disk I/O read bytes |
| `IOWriteBytes` | int64 | Disk I/O write bytes |
| `NodeHours` | Dec | Node-hours consumed |
| `NodesUsed` | int32 | Number of compute nodes |

### Derived Metrics

The system calculates derived values for billing:

```go
CPUCoreHours := CPUCoreSeconds / 3600
MemoryGBHours := MemoryGBSeconds / 3600
GPUHours := GPUSeconds / 3600
NetworkGB := (NetworkBytesIn + NetworkBytesOut) / (1024^3)
```

## Billing Formula

### Version: v1.0.0

The current billing formula uses deterministic fixed-point arithmetic for consensus safety.

```
TotalBillable = CPUCost + MemoryCost + GPUCost + StorageCost + NetworkCost + NodeCost

Where:
  CPUCost     = (CPUCoreSeconds / 3600) × CPUCoreHourRate
  MemoryCost  = (MemoryGBSeconds / 3600) × MemoryGBHourRate
  GPUCost     = (GPUSeconds / 3600) × GPUHourRate[GPUType]
  StorageCost = StorageGBHours × StorageGBHourRate
  NetworkCost = NetworkGB × NetworkGBRate
  NodeCost    = NodeHours × NodeHourRate
```

### Default Resource Rates

| Resource | Rate (uvirt/unit) | Unit |
|----------|-------------------|------|
| CPU Core | 10,000 | per hour |
| Memory | 1,000 | per GB-hour |
| GPU (A100) | 500,000 | per hour |
| GPU (V100) | 250,000 | per hour |
| GPU (T4) | 100,000 | per hour |
| Storage | 50 | per GB-hour |
| Network | 100 | per GB |
| Node | 50,000 | per hour |

### Minimum Charge

All jobs have a minimum billable amount of **1,000 uvirt** to cover overhead costs.

## Discounts

### Discount Types

1. **Volume Discounts** - Based on cumulative usage within a period
2. **Commitment Discounts** - Based on prepaid commitments
3. **Promotional Discounts** - Time-limited promotional rates

### Volume Discount Tiers

| CPU Core-Hours | Discount |
|----------------|----------|
| 0-100 | 0% |
| 100-500 | 5% |
| 500-1000 | 10% |
| 1000+ | 15% |

### Discount Application Order

1. Calculate base billable amount
2. Apply volume discounts (multiplicative)
3. Apply commitment discounts (additive)
4. Apply promotional discounts (additive)
5. Enforce minimum charge

## Billing Caps

### Cap Types

1. **Monthly Cap** - Maximum spending per calendar month
2. **Daily Cap** - Maximum spending per day
3. **Job Cap** - Maximum spending per individual job

### Default Caps

| Cap Type | Default Amount |
|----------|---------------|
| Monthly | 1,000,000,000,000 uvirt (1M VIRT) |
| Daily | 100,000,000,000 uvirt (100K VIRT) |
| Job | 10,000,000,000 uvirt (10K VIRT) |

## Provider Rewards

### Distribution Formula

```
PlatformFee    = BillableAmount × (PlatformFeeRateBps / 10000)
ProviderReward = BillableAmount × (ProviderRewardRateBps / 10000)

Default rates (configurable via governance):
  PlatformFeeRateBps:     250 (2.5%)
  ProviderRewardRateBps: 9750 (97.5%)
```

### Reward Record

Each successful settlement creates an `HPCRewardRecord`:

```go
type HPCRewardRecord struct {
    RecordID        string
    JobID           string
    ProviderAddress string
    RewardAmount    sdk.Coins
    PlatformFee     sdk.Coins
    DistributedAt   time.Time
    FormulaVersion  string
}
```

## Accounting Record Lifecycle

### States

1. **Pending** - Initial state after creation
2. **Finalized** - Metrics locked, ready for settlement
3. **Disputed** - Under dispute investigation
4. **Settled** - Successfully settled with escrow
5. **Corrected** - Modified after dispute resolution

### State Transitions

```
                    ┌──────────────┐
                    │   Pending    │
                    └──────┬───────┘
                           │ finalize()
                           ▼
                    ┌──────────────┐
        ┌───────────│  Finalized   │───────────┐
        │           └──────────────┘           │
        │ dispute()                     settle()│
        ▼                                       ▼
┌──────────────┐                        ┌──────────────┐
│   Disputed   │                        │   Settled    │
└──────┬───────┘                        └──────────────┘
       │ resolve()
       ▼
┌──────────────┐
│  Corrected   │
└──────────────┘
```

### Timing Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `AccountingFinalizationDelaySec` | 3600 (1 hour) | Delay before finalizing pending records |
| `SettlementDelaySec` | 86400 (24 hours) | Delay after finalization before settlement |
| `UsageSnapshotIntervalSec` | 300 (5 min) | Interval for capturing usage snapshots |

## Usage Snapshots

### Purpose

Usage snapshots capture point-in-time usage metrics for:
- Interim billing during long-running jobs
- Audit trail for disputes
- Reconciliation with scheduler logs

### Snapshot Types

1. **Initial** - Captured at job start
2. **Interim** - Captured periodically during execution
3. **Final** - Captured at job completion
4. **Correction** - Created during dispute resolution

### Snapshot Contents

```go
type HPCUsageSnapshot struct {
    SnapshotID        string
    JobID             string
    SnapshotType      SnapshotType
    SequenceNumber    int64
    Metrics           HPCDetailedMetrics
    CumulativeMetrics HPCDetailedMetrics
    JobState          string
    SnapshotTime      time.Time
    ProviderSignature string
    ContentHash       string
}
```

## Dispute Workflow

### Dispute Initiation

Either the customer or provider can dispute an accounting record:

1. Customer disputes if billed amount seems incorrect
2. Provider disputes if usage wasn't properly captured

### Dispute Evidence

Required evidence for dispute resolution:

- Usage snapshots from the disputed period
- Scheduler logs (via reconciliation service)
- Node metrics if available

### Resolution Outcomes

1. **Upheld** - Original billing stands
2. **Corrected** - New corrected record created
3. **Voided** - Record cancelled, no billing

### Correction Records

When a dispute results in a correction:

```go
type HPCAccountingRecord {
    // ... normal fields ...
    CorrectionOf    string  // Points to original record
    CorrectionNotes string  // Explanation of correction
}
```

## Reconciliation

### Purpose

The reconciliation service compares:
- On-chain accounting records
- Scheduler logs (SLURM/MOAB/OOD)
- Provider-signed usage snapshots

### Tolerances

| Metric | Tolerance |
|--------|-----------|
| Wall Clock | 0.1% |
| CPU Seconds | 1% |
| GPU Seconds | 0.5% |
| Memory | 2% |
| Network | 5% |

### Discrepancy Handling

When discrepancies exceed tolerances:

1. `Informational` - Logged but no action
2. `Warning` - Logged and flagged for review
3. `Critical` - Automatic dispute created

## Settlement Pipeline

### Process Flow

```
1. Job Completes
   └─> HPCAccountingService.finalizeJobAccounting()

2. Create Accounting Record
   └─> keeper.CreateAccountingRecord()

3. Wait Finalization Period
   └─> EndBlocker checks PeriodEnd + FinalizationDelay

4. Finalize Record
   └─> keeper.FinalizeAccountingRecord()

5. Wait Settlement Period
   └─> EndBlocker checks FinalizedAt + SettlementDelay

6. Process Settlement
   └─> keeper.ProcessJobSettlement()
       ├─> Generate Invoice (x/escrow)
       ├─> Process Payment (x/escrow)
       └─> Distribute Rewards (x/hpc)
```

### Automatic Settlement

When `EnableAutoSettlement` is true (default):
- EndBlocker processes pending settlements
- No manual intervention required

### Manual Settlement

For special cases:
- Submit `MsgSettleAccountingRecord` transaction
- Requires appropriate authority

## Integration with Escrow

### Invoice Generation

Settlement creates an escrow invoice:

```go
invoice := billing.Invoice{
    InvoiceID:       fmt.Sprintf("HPC-INV-%s", record.RecordID),
    CustomerAddress: record.CustomerAddress,
    ProviderAddress: record.ProviderAddress,
    LineItems:       generateLineItems(record),
    TotalAmount:     record.BillableAmount,
    Status:          billing.InvoiceStatusPending,
}
```

### Line Items

Each billing component becomes a line item:

| Description | Amount | ResourceType |
|-------------|--------|--------------|
| CPU Usage (X core-hours) | $Y | compute |
| Memory Usage (X GB-hours) | $Y | memory |
| GPU Usage (X hours) | $Y | gpu |
| Storage (X GB-hours) | $Y | storage |
| Network (X GB) | $Y | network |

## Edge Cases

### Job Cancellation

- Jobs cancelled before completion use last snapshot for billing
- Minimum charge still applies
- Queue time is not billed

### Job Failure

- Failed jobs are billed for actual usage before failure
- Provider may waive charges via dispute workflow
- Infrastructure failures may result in platform credits

### Resource Preemption

- Preempted jobs are billed for actual usage
- Queue time for rescheduling is tracked separately
- Preemption penalties may apply to provider rewards

### Long-Running Jobs

- Jobs > 24 hours receive interim invoices
- Interim billing uses incremental snapshots
- Final invoice reconciles with actual completion

### Clock Skew

- All times use chain block time for consistency
- Scheduler times are advisory only
- Disputes resolved using on-chain timestamps

## Governance Parameters

The following parameters are adjustable via governance:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `PlatformFeeRate` | 50000 (5%) | Platform fee (scale: 1,000,000) |
| `ProviderRewardRate` | 800000 (80%) | Provider share (scale: 1,000,000) |
| `NodeRewardRate` | 150000 (15%) | Node operator share (scale: 1,000,000) |
| `PlatformFeeRateBps` | 250 (2.5%) | Platform fee in basis points |
| `AccountingFinalizationDelaySec` | 3600 | Finalization delay |
| `SettlementDelaySec` | 86400 | Settlement delay |
| `DefaultDenom` | "uvirt" | Default currency |

## API Reference

### Keeper Methods

```go
// Accounting
CreateAccountingRecord(ctx, record) error
GetAccountingRecord(ctx, recordID) (HPCAccountingRecord, bool)
FinalizeAccountingRecord(ctx, recordID) error
MarkAccountingRecordDisputed(ctx, recordID, disputeID) error

// Billing
CalculateJobBilling(ctx, jobID) (HPCBillableBreakdown, sdk.Coins, error)
GenerateInvoiceForJob(ctx, jobID) (billing.Invoice, error)

// Settlement
ProcessJobSettlement(ctx, recordID) error
DistributeJobRewardsFromSettlement(ctx, jobID, record) error
ProcessDisputeResolution(ctx, disputeID, outcome, newMetrics) error

// Snapshots
CreateUsageSnapshot(ctx, snapshot) error
GetJobSnapshots(ctx, jobID) ([]HPCUsageSnapshot, error)

// Reconciliation
CreateReconciliationRecord(ctx, record) error
GetPendingReconciliations(ctx) ([]HPCReconciliationRecord, error)
```

### Provider Daemon Services

```go
// HPCAccountingService
StartAccountingLoop(ctx) error
StopAccountingLoop()
HandleJobStarted(jobID, clusterID)
HandleJobCompleted(jobID, success bool)

// HPCReconciliationService
StartReconciliationLoop(ctx) error
StopReconciliationLoop()
ReconcileJob(ctx, jobID) ([]ReconciliationDiscrepancy, error)
```

## Troubleshooting

### Common Issues

1. **Settlement stuck in Pending**
   - Check `AccountingFinalizationDelaySec` hasn't elapsed
   - Verify job completion was properly signaled

2. **Discrepancy between scheduler and chain**
   - Run reconciliation service
   - Check clock sync between provider and chain

3. **Provider rewards not received**
   - Verify settlement completed successfully
   - Check reward distribution transaction

4. **Dispute not resolving**
   - Verify dispute resolution period hasn't expired
   - Check arbitration governance proposal status

## Version History

| Version | Date | Changes |
|---------|------|---------|
| v1.0.0 | 2024-01 | Initial billing formula |

---

*This document is part of the VirtEngine VE-5A feature implementation.*
