# HPC Marketplace Provider E2E Test Guide

**VE-8C**: End-to-end testing for HPC and marketplace provider flows.

This document describes the E2E test suite for the complete provider workflow: registration → offering listing → job execution → usage reporting → settlement.

## Overview

The HPC Marketplace E2E test suite validates the complete provider lifecycle on VirtEngine, ensuring deterministic and reliable execution of:

1. **Provider Registration** - On-chain provider registration with Waldur sync
2. **Offering Publishing** - Marketplace offerings creation and chain synchronization
3. **Order & Allocation** - Order creation, bidding, and resource allocation
4. **Job Execution** - HPC job submission and scheduler execution (SLURM/MOAB)
5. **Usage Reporting** - Metrics capture and on-chain usage reports
6. **Settlement** - Invoice generation and provider payout

## Test Structure

```
tests/e2e/
├── hpc_marketplace_e2e_test.go     # Main E2E test suite
├── provider_daemon_e2e_test.go     # Provider daemon workflow tests
└── fixtures/
    └── hpc_provider_fixtures.go    # Reusable test fixtures
```

## Prerequisites

### Required Components

| Component | Description | Configuration |
|-----------|-------------|---------------|
| Go 1.21+ | Build and test execution | `GO_VERSION` in CI |
| Docker | Localnet container orchestration | Docker Buildx required |
| Localnet | Local chain for integration tests | `./scripts/localnet.sh` |

### Environment Variables

```bash
# Optional: Override test configuration
export E2E_PROVIDER_ADDRESS="virtengine1..."
export E2E_CUSTOMER_ADDRESS="virtengine1..."
export E2E_CLUSTER_ID="test-cluster"
export E2E_COMET_RPC="http://localhost:26657"
export E2E_WALDUR_BASE_URL="http://localhost:8080"  # Mock Waldur
```

## Running the Tests

### Local Development

```bash
# Run full E2E suite
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestHPCMarketplaceE2E"

# Run specific test phases
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestHPCMarketplaceE2E/TestA_"
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestHPCMarketplaceE2E/TestD_JobSubmission"

# Run with verbose logging
go test -v -tags="e2e.integration" ./tests/e2e/... -run "TestHPCMarketplaceE2E" -count=1
```

### With Localnet

```bash
# Start localnet
./scripts/localnet.sh start

# Wait for chain readiness
./scripts/localnet.sh status

# Run E2E tests
make test-integration

# Stop localnet
./scripts/localnet.sh stop
```

### CI Pipeline

The tests run automatically in CI via the `hpc-provider-e2e` job:

```yaml
# .github/workflows/ci.yaml
hpc-provider-e2e:
  name: HPC Provider E2E Tests
  runs-on: ubuntu-latest
  needs: [build, integration]
```

## Test Phases

### A. Staging Environment Setup

Tests the initialization of the provider daemon and scheduler backend:

```go
func (s *HPCMarketplaceE2ETestSuite) TestA_StagingEnvironmentSetup()
```

Validates:
- Scheduler backend connectivity
- Provider daemon configuration
- Waldur bridge configuration

### B. Provider Registration and Offerings

Tests provider on-boarding and offering publication:

```go
func (s *HPCMarketplaceE2ETestSuite) TestB_ProviderRegistrationAndOfferings()
```

Validates:
- Provider registration on-chain
- Offering creation (compute, GPU, etc.)
- Chain-to-Waldur synchronization

### C. Order Creation and Allocation

Tests the marketplace order workflow:

```go
func (s *HPCMarketplaceE2ETestSuite) TestC_OrderCreationAndAllocation()
```

Validates:
- Order creation by customer
- Provider bid placement
- Bid acceptance and allocation
- Resource provisioning

### D. Job Submission and Execution

Tests HPC job lifecycle:

```go
func (s *HPCMarketplaceE2ETestSuite) TestD_JobSubmissionAndExecution()
```

Validates:
- Job submission to scheduler
- State transitions (pending → queued → running → completed)
- Scheduler execution verification

### E. Usage Metrics and Reporting

Tests usage capture and on-chain reporting:

```go
func (s *HPCMarketplaceE2ETestSuite) TestE_UsageMetricsAndReporting()
```

Validates:
- Metrics capture (CPU, memory, GPU, network)
- Usage record creation
- On-chain submission
- Record integrity

### F. Invoice and Settlement

Tests billing and payout:

```go
func (s *HPCMarketplaceE2ETestSuite) TestF_InvoiceAndSettlement()
```

Validates:
- Invoice generation from usage
- Line item calculation
- Settlement trigger
- Provider payout (97.5%)
- Platform fee (2.5%)

### G. State Transitions and Events

Tests on-chain state machine:

```go
func (s *HPCMarketplaceE2ETestSuite) TestG_StateTransitionsAndEvents()
```

Validates:
- Valid job state transitions
- Event emissions
- Accounting status transitions

### H. Negative Scenarios

Tests error handling:

```go
func (s *HPCMarketplaceE2ETestSuite) TestH_NegativeScenarios()
```

Validates:
- Failed job handling
- Partial usage reporting
- Timeout handling
- Job cancellation
- Resource rejection
- Disputed settlements

## Using Test Fixtures

The `fixtures` package provides reusable test components:

```go
import "github.com/virtengine/virtengine/tests/e2e/fixtures"

func TestMyHPCFlow(t *testing.T) {
    config := fixtures.DefaultFixtureConfig()
    config.NumJobs = 10
    
    fixture := fixtures.NewHPCProviderFixture(t, config)
    require.NoError(t, fixture.Setup())
    defer fixture.Teardown()
    
    // Create and submit a job
    job := fixture.CreateJob("my-test", 4, 8192, 1)
    schedulerJob, err := fixture.SubmitJob(job)
    require.NoError(t, err)
    
    // Create an order and place a bid
    order := fixture.CreateOrder("hpc-compute-medium", "50.0")
    bid, err := fixture.PlaceBid(order.OrderID, "40.0")
    require.NoError(t, err)
}
```

### Fixture Configuration

```go
type HPCProviderFixtureConfig struct {
    ProviderAddress      string  // Test provider address
    CustomerAddress      string  // Test customer address
    ClusterID            string  // Test cluster ID
    NumOfferings         int     // Number of offerings to create
    NumOrders            int     // Number of orders to create
    NumJobs              int     // Number of jobs to create
    EnableMockScheduler  bool    // Enable mock scheduler
    EnableMockWaldur     bool    // Enable mock Waldur client
    EnableMockSettlement bool    // Enable mock settlement pipeline
    JobTimeoutSeconds    int64   // Max job runtime
    CleanupOnTeardown    bool    // Remove test data on teardown
}
```

## Mock Components

### MockHPCScheduler

Simulates SLURM/MOAB scheduler behavior:

```go
scheduler := NewMockHPCScheduler()
scheduler.Start(ctx)

// Submit job
job := &hpctypes.HPCJob{...}
schedulerJob, err := scheduler.SubmitJob(ctx, job)

// Simulate state changes
scheduler.SetJobState(job.JobID, pd.HPCJobStateRunning)
scheduler.SetJobMetrics(job.JobID, &pd.HPCSchedulerMetrics{
    WallClockSeconds: 3600,
    CPUCoreSeconds:   14400,
})
scheduler.SetJobState(job.JobID, pd.HPCJobStateCompleted)
```

### MockWaldurClient

Simulates Waldur marketplace operations:

```go
waldur := NewMockWaldurClient()
waldur.SetProviderRegistered(providerAddr, true)
waldur.PublishOffering(ctx, offering)
waldur.CreateOrder(ctx, order)
waldur.PlaceBid(ctx, bid)
waldur.AcceptBid(ctx, orderID, bidID)
```

### MockSettlementClient

Simulates settlement pipeline:

```go
settlement := NewMockSettlementClient()
settlement.CreateInvoice(ctx, invoice)
settlement.TriggerSettlement(ctx, invoiceID)
settlement.DisputeInvoice(ctx, invoiceID, reason)
```

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Tests timeout | Chain not ready | Increase localnet wait time |
| Connection refused | Localnet not running | Run `./scripts/localnet.sh start` |
| Missing dependencies | Go modules out of sync | Run `go mod tidy && go mod vendor` |
| Build tag errors | Missing build tag | Use `-tags="e2e.integration"` |

### Debug Mode

Enable verbose logging:

```bash
# Set log level
export LOG_LEVEL=debug

# Run with race detector
go test -v -race -tags="e2e.integration" ./tests/e2e/...
```

### Checking Test Coverage

```bash
# Generate coverage report
go test -v -tags="e2e.integration" -coverprofile=e2e-coverage.out ./tests/e2e/...

# View coverage
go tool cover -html=e2e-coverage.out -o e2e-coverage.html
```

## Adding New Tests

1. Add test method to `HPCMarketplaceE2ETestSuite`
2. Follow naming convention: `Test[Phase]_[Description]`
3. Use mocks from the test file or fixtures package
4. Ensure proper cleanup in test teardown

```go
func (s *HPCMarketplaceE2ETestSuite) TestX_NewFeature() {
    ctx := context.Background()
    
    s.Run("SubTestCase", func() {
        // Test implementation
        s.Require().NoError(err)
        s.Equal(expected, actual)
    })
}
```

## Acceptance Criteria

The test suite validates these acceptance criteria:

- ✅ End-to-end provider flow passes with deterministic results
- ✅ Usage reports drive invoices and settlements
- ✅ Failures handled with clear state transitions
- ✅ CI job exists for provider flow

## Related Documentation

- [Testing Guide](testing-guide.md)
- [Provider Daemon Waldur Integration](provider-daemon-waldur-integration.md)
- [HPC Billing Rules](hpc-billing-rules.md)
- [Provider Guide](provider-guide.md)
