# Settlement & Billing System Integration Tests

This directory contains integration tests for the VirtEngine billing and settlement system, which implements a complete usage→invoice→settlement→payment pipeline.

## Test Suites

### Invoice Integration Tests (`invoice_test.go`)

Tests the complete invoice lifecycle and hash-chained ledger system:

- **Invoice Creation**: Validates invoice creation with proper ledger records
- **State Machine**: Tests invoice status transitions (Pending → Paid → Disputed → etc.)
- **Payment Recording**: Tests partial and full payment recording
- **Ledger Chain Integrity**: Verifies hash-chained ledger entries for audit trail
- **Query Operations**: Tests filtering invoices by provider, customer, status, escrow
- **Dispute Workflow**: Tests invoice dispute and cancellation flows
- **Idempotency**: Ensures invoices can be safely re-submitted

Key test patterns:
```go
// Test state transition
_, err := invoiceKeeper.UpdateInvoiceStatus(ctx, invoiceID, newStatus, initiator)

// Record payment
_, err := invoiceKeeper.RecordPayment(ctx, invoiceID, amount, processor)

// Verify ledger chain
err := invoiceKeeper.VerifyLedgerChain(ctx, invoiceID)
```

### Usage Pipeline Tests (`usage_pipeline_test.go`)

Tests the end-to-end usage reporting and settlement pipeline:

- **Usage Report Submission**: Provider daemon submits metered usage
- **Invoice Generation**: Accumulates usage records into invoices
- **Full Pipeline**: Tests usage→invoice→settlement in one flow
- **Invoice Approval**: Customer/admin approval workflow
- **Dispute Handling**: Dispute initiation and holdback processing
- **Validation**: Input validation for usage reports

Key test patterns:
```go
// Submit usage
record, err := usagePipeline.SubmitUsageReport(ctx, report)

// Generate invoice from usage
invoice, err := usagePipeline.GenerateInvoiceFromUsage(ctx, leaseID, periodEnd)

// Full pipeline
result, err := usagePipeline.ProcessUsageSettlement(ctx, leaseID)
```

## Architecture Overview

### Components

1. **Invoice Keeper** (`x/escrow/keeper/invoice.go`)
   - Invoice CRUD operations
   - Hash-chained ledger for audit trail
   - State machine enforcement
   - Payment recording

2. **Usage Pipeline Keeper** (`x/escrow/keeper/usage.go`)
   - Usage report ingestion
   - Invoice generation from usage
   - Settlement pipeline orchestration
   - Approval/dispute workflow

3. **Settlement Integration Keeper** (`x/escrow/keeper/settlement_integration.go`)
   - Escrow release on settlement
   - Treasury accounting
   - Fee distribution
   - Batch settlement

4. **Reconciliation Keeper** (`x/escrow/keeper/reconciliation.go`)
   - Usage record storage
   - Payout tracking
   - Reconciliation reports

### Data Flow

```
Provider Daemon
    ↓ (UsageReport)
Usage Pipeline Keeper
    ↓ (UsageRecord stored)
    ↓ (Accumulates over period)
    ↓ (GenerateInvoice)
Invoice Keeper
    ↓ (Invoice + Ledger Record)
    ↓ (Approval/Payment)
Settlement Integration Keeper
    ↓ (Escrow Release)
Treasury Allocation
```

### State Machine

Invoices follow this state machine:

```
Pending → Paid ✓
       → PartiallyPaid → Paid ✓
       → Disputed → Resolved → Paid ✓
       → Cancelled ✓
       → Overdue → Paid ✓
```

Terminal states: `Paid`, `Cancelled`, `Refunded`

## Running Tests

```bash
# Run all settlement integration tests
go test -v -tags="e2e.integration" ./tests/integration/settlement/...

# Run specific test suite
go test -v -tags="e2e.integration" ./tests/integration/settlement/ -run TestInvoiceIntegrationTestSuite

# Run specific test
go test -v -tags="e2e.integration" ./tests/integration/settlement/ -run TestInvoiceIntegrationTestSuite/TestInvoiceCreation
```

## Test Helpers

Located in `testutil/escrow_test_helpers.go`:

- `SetupEscrowKeeper(t)`: Creates test escrow keeper with mocked dependencies
- `AccAddress(t)`: Generates test account addresses
- `CreateTestUsageRecord(...)`: Creates test usage records

## Implementation Notes

### Deterministic Invoice Generation

Invoice IDs and ledger entries must be deterministic for consensus:
- Invoice IDs derived from escrow/order/lease IDs + block height
- Ledger entry hashes computed from previous entry hash + content
- Sequence numbers enforce ordering

### Escrow Integration

Invoices tie to escrow accounts for payment settlement:
- Invoice creation requires valid escrow ID
- Settlement releases funds from escrow to provider
- Disputes trigger escrow holdbacks

### Tax and Currency Handling

The billing system supports:
- Multiple currencies per invoice
- Tax/VAT metadata storage
- Currency conversion hooks (for display/reporting)
- Region-specific tax rules

### Audit Trail

Every invoice has a hash-chained ledger:
- Genesis entry when invoice created
- Transition entries for status changes
- Payment entries for recorded payments
- Chain validates via previous hash verification

## Related Files

- `x/escrow/types/billing/*.go` - Type definitions
- `x/escrow/keeper/*_integration.go` - Integration layer implementations
- `pkg/billing/*.go` - Off-chain billing services
- `INVOICE_LIFECYCLE.md` - Detailed invoice state machine documentation
- `RECONCILIATION_SOP.md` - Reconciliation standard operating procedures

## Test Coverage

Current coverage:
- Invoice lifecycle: 11 test cases
- Usage pipeline: 6 test cases
- Total: 17 integration tests

All tests verify both success and failure paths, including:
- Invalid state transitions
- Invalid input validation
- Idempotency guarantees
- Ledger chain integrity
