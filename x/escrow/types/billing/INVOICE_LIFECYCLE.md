# Invoice Lifecycle and API Documentation

## Overview

The VirtEngine invoice system provides deterministic invoice generation, immutable on-chain ledger records, and artifact storage for full invoice documents. This enables transparent billing for marketplace and HPC usage.

## Invoice Lifecycle

### Status States

| Status | Description | Terminal |
|--------|-------------|----------|
| `draft` | Invoice being prepared, not yet issued | No |
| `pending` | Awaiting payment | No |
| `paid` | Fully paid | Yes |
| `partially_paid` | Partial payment received | No |
| `overdue` | Past due date | No |
| `disputed` | Under dispute | No |
| `cancelled` | Cancelled/voided | Yes |
| `refunded` | Refunded | Yes |

### Valid Transitions

```
draft → pending (issue invoice)
draft → cancelled (cancel draft)

pending → paid (full payment)
pending → partially_paid (partial payment)
pending → overdue (mark overdue)
pending → disputed (dispute invoice)
pending → cancelled (cancel pending)

partially_paid → paid (complete payment)
partially_paid → overdue (mark overdue)
partially_paid → disputed (dispute invoice)

overdue → paid (pay overdue)
overdue → partially_paid (partial pay overdue)
overdue → disputed (dispute overdue)
overdue → cancelled (write off)

disputed → pending (resolve - pending)
disputed → paid (resolve - paid)
disputed → cancelled (resolve - cancel)
disputed → refunded (resolve - refund)

paid → refunded (refund paid)
```

## Invoice Generation

### From Usage Records

```go
gen := billing.NewInvoiceGenerator(billing.DefaultInvoiceGeneratorConfig())

req := billing.InvoiceGenerationRequest{
    EscrowID:    "escrow-001",
    OrderID:     "order-001",
    LeaseID:     "lease-001",
    Provider:    "virtengine1provider...",
    Customer:    "virtengine1customer...",
    UsageInputs: []billing.UsageInput{
        {
            UsageRecordID: "usage-001",
            UsageType:     billing.UsageTypeCPU,
            Quantity:      sdkmath.LegacyNewDec(10),
            Unit:          "core-hour",
            UnitPrice:     sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(100)),
            Description:   "CPU usage for 10 core-hours",
            PeriodStart:   periodStart,
            PeriodEnd:     periodEnd,
        },
    },
    BillingPeriod: billing.BillingPeriod{
        StartTime:       periodStart,
        EndTime:         periodEnd,
        DurationSeconds: 86400,
        PeriodType:      billing.BillingPeriodTypeDaily,
    },
    Currency: "uvirt",
}

invoice, err := gen.GenerateInvoice(req, blockHeight, now)
```

### From HPC Usage

```go
hpcUsage := billing.HPCUsageInput{
    JobID:           "job-001",
    CPUHours:        sdkmath.LegacyNewDec(100),
    GPUHours:        sdkmath.LegacyNewDec(10),
    MemoryGBHours:   sdkmath.LegacyNewDec(512),
    StorageGBMonths: sdkmath.LegacyNewDec(100),
    NetworkGB:       sdkmath.LegacyNewDec(50),
    PeriodStart:     periodStart,
    PeriodEnd:       periodEnd,
}

invoice, err := gen.GenerateHPCInvoice(
    escrowID, orderID, leaseID,
    provider, customer,
    hpcUsage, pricingPolicy,
    blockHeight, now,
)
```

## On-Chain Storage

### Invoice Ledger Record

The `InvoiceLedgerRecord` is the immutable on-chain record containing:

- Invoice metadata (ID, number, provider, customer, amounts)
- Content hash (SHA-256) for artifact verification
- Artifact CID for document retrieval
- Status and timestamps

```go
record, err := billing.NewInvoiceLedgerRecord(invoice, artifactCID, blockHeight, now)
```

### Ledger Entries

Each state change creates an `InvoiceLedgerEntry` for audit trail. Entries form an immutable hash chain:

```go
// Get previous entry hash and sequence from the chain
previousHash, seqNum := keeper.GetLastEntryHashAndSeq(invoiceID)

entry := billing.NewInvoiceLedgerEntry(
    entryID,
    invoiceID,
    billing.LedgerEntryTypePayment,
    previousStatus,
    newStatus,
    amount,
    description,
    initiator,
    txHash,
    previousHash,   // Hash chain link
    seqNum + 1,     // Next sequence number
    blockHeight,
    timestamp,
)
```

### Hash Chain Verification

The ledger chain can be verified for integrity:

```go
chain, err := keeper.GetInvoiceLedgerChain(ctx, invoiceID)
if err := chain.Validate(); err != nil {
    // Chain integrity compromised
}
```

## Artifact Storage

Invoice documents are stored off-chain with on-chain verification:

```go
store := billing.NewMemoryArtifactStore() // or production implementation
docGen := billing.NewInvoiceDocumentGenerator(store)

artifact, err := docGen.GenerateJSONDocument(ctx, invoice, createdBy)
// Returns CID and content hash for on-chain reference
```

### Verification

```go
valid, err := record.VerifyContentHash(fullInvoice)
// or
valid, err := store.Verify(ctx, cid, content)
```

## Query APIs

### Get Invoice by ID

```go
record, err := invoiceKeeper.GetInvoice(ctx, invoiceID)
```

### Get Invoices by Provider

```go
records, pageRes, err := invoiceKeeper.GetInvoicesByProvider(ctx, provider, pagination)
```

### Get Invoices by Customer

```go
records, pageRes, err := invoiceKeeper.GetInvoicesByCustomer(ctx, customer, pagination)
```

### Get Invoices by Status

```go
records, pageRes, err := invoiceKeeper.GetInvoicesByStatus(ctx, billing.InvoiceStatusPending, pagination)
```

### Get Invoice Ledger History

```go
entries, err := invoiceKeeper.GetInvoiceLedgerEntries(ctx, invoiceID)
```

## Status Transitions

Use the `InvoiceStatusMachine` for safe state transitions:

```go
sm := billing.NewInvoiceStatusMachine(invoice, blockHeight, now)

// Issue invoice
err := sm.MarkIssued(initiator)

// Record payment
err := sm.TransitionWithPayment(amount, initiator)

// Mark disputed
err := sm.MarkDisputed(disputeWindow, initiator, reason)

// Get ledger entry
entry := sm.GetLedgerEntry()
```

## Reconciliation

Generate reconciliation reports for billing verification:

```go
report := &billing.ReconciliationReport{
    ReportID:    "report-001",
    ReportType:  billing.ReconciliationReportTypeDaily,
    PeriodStart: periodStart,
    PeriodEnd:   periodEnd,
    Status:      billing.ReconciliationStatusComplete,
    Summary:     summary,
    InvoiceIDs:  invoiceIDs,
    GeneratedAt: now,
    GeneratedBy: "system",
    BlockHeight: blockHeight,
}

err := invoiceKeeper.SaveReconciliationReport(ctx, report)
```

## Deterministic Properties

1. **Invoice IDs**: Generated deterministically from escrow ID, order ID, addresses, period, and block height
2. **Content Hashes**: SHA-256 of canonical invoice JSON
3. **Rounding**: Configurable banker's rounding (half-even) for amounts

## Store Keys

| Prefix | Description |
|--------|-------------|
| `0x01` | Invoices |
| `0x02` | Invoice by provider index |
| `0x03` | Invoice by customer index |
| `0x04` | Invoice by status index |
| `0x05` | Invoice by escrow index |
| `0x60` | Invoice ledger records |
| `0x61` | Invoice ledger entries |
| `0x62` | Ledger entries by invoice index |
| `0x63` | Invoice artifacts |
| `0x70` | Reconciliation reports |

## Usage Types

| Type | Unit | Description |
|------|------|-------------|
| `cpu` | core-hour | CPU usage |
| `memory` | gb-hour | Memory usage |
| `storage` | gb-month | Storage usage |
| `network` | gb | Network bandwidth |
| `gpu` | gpu-hour | GPU usage |
| `fixed` | unit | Fixed charges |
| `setup` | unit | One-time setup fees |
| `other` | unit | Other charges |

## Configuration

```go
config := billing.InvoiceGeneratorConfig{
    InvoiceNumberPrefix:    "VE-INV",
    DefaultPaymentTermDays: 7,
    RoundingMode:           billing.RoundingModeHalfEven,
    DefaultCurrency:        "uvirt",
    ApplyTax:               false,
    DefaultTaxJurisdiction: "US",
}
```

## Events

Invoice operations emit the following events:

- `invoice.created` - New invoice created
- `invoice.issued` - Invoice issued (draft → pending)
- `invoice.payment` - Payment recorded
- `invoice.paid` - Invoice fully paid
- `invoice.disputed` - Invoice disputed
- `invoice.cancelled` - Invoice cancelled
- `invoice.refunded` - Invoice refunded
