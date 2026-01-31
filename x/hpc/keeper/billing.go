// Package keeper implements the HPC module keeper.
//
// VE-5A: Billing keeper methods for cost calculation and invoice generation
package keeper

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Billing Calculation
// ============================================================================

// CalculateJobBilling calculates billing for a completed job
func (k Keeper) CalculateJobBilling(ctx sdk.Context, jobID string) (*types.HPCAccountingRecord, error) {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return nil, types.ErrJobNotFound
	}

	// Get billing rules for the provider
	providerAddr, err := sdk.AccAddressFromBech32(job.ProviderAddress)
	if err != nil {
		return nil, err
	}
	rules := k.GetOrDefaultBillingRules(ctx, providerAddr)

	// Get usage snapshots for the job
	snapshots := k.GetUsageSnapshotsByJob(ctx, jobID)
	if len(snapshots) == 0 {
		// If no snapshots, create metrics from job timing
		return k.calculateBillingFromJobTiming(ctx, &job, rules)
	}

	// Get final snapshot
	finalSnapshot, _ := k.GetLatestUsageSnapshot(ctx, jobID)

	// Calculate billable amount
	calculator := types.NewHPCBillingCalculator(rules)

	// Get historical usage for discount evaluation
	customerAddr, _ := sdk.AccAddressFromBech32(job.CustomerAddress)
	monthAgo := ctx.BlockTime().Add(-30 * 24 * time.Hour)
	aggregations := k.GetAggregationsByCustomer(ctx, customerAddr, monthAgo, ctx.BlockTime())
	var historicalUsage *types.AccountingAggregation
	if len(aggregations) > 0 {
		historicalUsage = &aggregations[0]
	}

	// Evaluate discounts
	appliedDiscounts := calculator.EvaluateDiscounts(&finalSnapshot.CumulativeMetrics, job.CustomerAddress, historicalUsage)

	// Calculate billable amount
	breakdown, billable, err := calculator.CalculateBillableAmount(&finalSnapshot.CumulativeMetrics, appliedDiscounts, nil)
	if err != nil {
		return nil, err
	}

	// Calculate provider reward and platform fee
	providerReward := calculator.CalculateProviderReward(billable)
	platformFee := calculator.CalculatePlatformFee(billable)

	// Collect signed usage record IDs from snapshots
	var signedRecordIDs []string
	for _, s := range snapshots {
		if s.ProviderSignature != "" {
			signedRecordIDs = append(signedRecordIDs, s.SnapshotID)
		}
	}

	// Create accounting record
	record := &types.HPCAccountingRecord{
		JobID:             jobID,
		ClusterID:         job.ClusterID,
		ProviderAddress:   job.ProviderAddress,
		CustomerAddress:   job.CustomerAddress,
		OfferingID:        job.OfferingID,
		SchedulerType:     "", // Will be set by provider daemon
		UsageMetrics:      finalSnapshot.CumulativeMetrics,
		BillableAmount:    billable,
		BillableBreakdown: *breakdown,
		AppliedDiscounts:  appliedDiscounts,
		ProviderReward:    providerReward,
		PlatformFee:       platformFee,
		SignedUsageRecords: signedRecordIDs,
		Status:            types.AccountingStatusPending,
		PeriodStart:       finalSnapshot.Metrics.SubmitTime,
		PeriodEnd:         ctx.BlockTime(),
		FormulaVersion:    rules.FormulaVersion,
	}

	if finalSnapshot.Metrics.EndTime != nil {
		record.PeriodEnd = *finalSnapshot.Metrics.EndTime
	}

	// Create the record
	if err := k.CreateAccountingRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// calculateBillingFromJobTiming calculates billing from job timing when no snapshots exist
func (k Keeper) calculateBillingFromJobTiming(ctx sdk.Context, job *types.HPCJob, rules types.HPCBillingRules) (*types.HPCAccountingRecord, error) {
	// Calculate metrics from job timing
	var wallClock int64
	if job.StartedAt != nil && job.CompletedAt != nil {
		wallClock = int64(job.CompletedAt.Sub(*job.StartedAt).Seconds())
	}

	var queueTime int64
	if job.QueuedAt != nil && job.StartedAt != nil {
		queueTime = int64(job.StartedAt.Sub(*job.QueuedAt).Seconds())
	}

	// Estimate metrics from resource request
	cpuCoreSeconds := wallClock * int64(job.Resources.CPUCoresPerNode) * int64(job.Resources.Nodes)
	memGBSeconds := wallClock * int64(job.Resources.MemoryGBPerNode) * int64(job.Resources.Nodes)
	gpuSeconds := wallClock * int64(job.Resources.GPUsPerNode) * int64(job.Resources.Nodes)
	nodeHours := sdkmath.LegacyNewDec(wallClock * int64(job.Resources.Nodes)).Quo(sdkmath.LegacyNewDec(3600))

	metrics := types.HPCDetailedMetrics{
		WallClockSeconds: wallClock,
		QueueTimeSeconds: queueTime,
		CPUCoreSeconds:   cpuCoreSeconds,
		MemoryGBSeconds:  memGBSeconds,
		GPUSeconds:       gpuSeconds,
		GPUType:          job.Resources.GPUType,
		NodeHours:        nodeHours,
		NodesUsed:        job.Resources.Nodes,
		SubmitTime:       job.CreatedAt,
	}

	if job.StartedAt != nil {
		metrics.StartTime = job.StartedAt
	}
	if job.CompletedAt != nil {
		metrics.EndTime = job.CompletedAt
	}

	// Calculate billing
	calculator := types.NewHPCBillingCalculator(rules)
	breakdown, billable, err := calculator.CalculateBillableAmount(&metrics, nil, nil)
	if err != nil {
		return nil, err
	}

	providerReward := calculator.CalculateProviderReward(billable)
	platformFee := calculator.CalculatePlatformFee(billable)

	record := &types.HPCAccountingRecord{
		JobID:             job.JobID,
		ClusterID:         job.ClusterID,
		ProviderAddress:   job.ProviderAddress,
		CustomerAddress:   job.CustomerAddress,
		OfferingID:        job.OfferingID,
		UsageMetrics:      metrics,
		BillableAmount:    billable,
		BillableBreakdown: *breakdown,
		ProviderReward:    providerReward,
		PlatformFee:       platformFee,
		Status:            types.AccountingStatusPending,
		PeriodStart:       job.CreatedAt,
		PeriodEnd:         ctx.BlockTime(),
		FormulaVersion:    rules.FormulaVersion,
	}

	if job.CompletedAt != nil {
		record.PeriodEnd = *job.CompletedAt
	}

	if err := k.CreateAccountingRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// CalculateInterimBilling calculates interim billing for a running job
func (k Keeper) CalculateInterimBilling(ctx sdk.Context, jobID string, metrics *types.HPCDetailedMetrics) (sdk.Coins, *types.BillableBreakdown, error) {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return nil, nil, types.ErrJobNotFound
	}

	providerAddr, err := sdk.AccAddressFromBech32(job.ProviderAddress)
	if err != nil {
		return nil, nil, err
	}
	rules := k.GetOrDefaultBillingRules(ctx, providerAddr)

	calculator := types.NewHPCBillingCalculator(rules)
	breakdown, billable, err := calculator.CalculateBillableAmount(metrics, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	return billable, breakdown, nil
}

// ============================================================================
// Invoice Generation
// ============================================================================

// GenerateInvoiceForJob generates an invoice from an accounting record
func (k Keeper) GenerateInvoiceForJob(ctx sdk.Context, accountingRecordID string) (string, error) {
	record, exists := k.GetAccountingRecord(ctx, accountingRecordID)
	if !exists {
		return "", types.ErrInvalidJobAccounting.Wrap("accounting record not found")
	}

	if record.Status != types.AccountingStatusFinalized && record.Status != types.AccountingStatusPending {
		return "", types.ErrInvalidJobAccounting.Wrapf("cannot generate invoice for %s record", record.Status)
	}

	// Get job for additional context
	job, _ := k.GetJob(ctx, record.JobID)

	// Generate invoice ID
	invoiceID := fmt.Sprintf("hpc-inv-%s", record.RecordID)

	// Update accounting record with invoice ID
	record.InvoiceID = invoiceID
	if err := k.SetAccountingRecord(ctx, record); err != nil {
		return "", err
	}

	k.Logger(ctx).Info("generated invoice for HPC job",
		"invoice_id", invoiceID,
		"job_id", record.JobID,
		"amount", record.BillableAmount.String(),
		"customer", record.CustomerAddress)

	// Emit event for invoice generation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_invoice_generated",
			sdk.NewAttribute("invoice_id", invoiceID),
			sdk.NewAttribute("job_id", record.JobID),
			sdk.NewAttribute("accounting_record_id", record.RecordID),
			sdk.NewAttribute("customer", record.CustomerAddress),
			sdk.NewAttribute("provider", record.ProviderAddress),
			sdk.NewAttribute("amount", record.BillableAmount.String()),
			sdk.NewAttribute("escrow_id", job.EscrowID),
		),
	)

	return invoiceID, nil
}

// ============================================================================
// Period Aggregation
// ============================================================================

// AggregateAccountingForPeriod aggregates accounting records for a billing period
func (k Keeper) AggregateAccountingForPeriod(
	ctx sdk.Context,
	customerAddr sdk.AccAddress,
	providerAddr sdk.AccAddress,
	periodStart, periodEnd time.Time,
) (*types.AccountingAggregation, error) {
	params := k.GetParams(ctx)

	aggregation := types.AccountingAggregation{
		CustomerAddress:     customerAddr.String(),
		ProviderAddress:     providerAddr.String(),
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
		TotalCPUCoreHours:   sdkmath.LegacyZeroDec(),
		TotalGPUHours:       sdkmath.LegacyZeroDec(),
		TotalMemoryGBHours:  sdkmath.LegacyZeroDec(),
		TotalStorageGBHours: sdkmath.LegacyZeroDec(),
		TotalNodeHours:      sdkmath.LegacyZeroDec(),
		TotalBillableAmount: sdk.NewCoins(),
		TotalDiscounts:      sdk.NewCoins(),
		AccountingRecordIDs: make([]string, 0),
	}

	// Iterate through accounting records
	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		// Filter by customer and provider
		if record.CustomerAddress != customerAddr.String() {
			return false
		}
		if record.ProviderAddress != providerAddr.String() {
			return false
		}

		// Filter by period
		if record.PeriodEnd.Before(periodStart) || record.PeriodStart.After(periodEnd) {
			return false
		}

		// Aggregate metrics
		aggregation.TotalJobs++
		aggregation.AccountingRecordIDs = append(aggregation.AccountingRecordIDs, record.RecordID)

		// Convert to hours
		cpuHours := sdkmath.LegacyNewDec(record.UsageMetrics.CPUCoreSeconds).Quo(sdkmath.LegacyNewDec(3600))
		gpuHours := sdkmath.LegacyNewDec(record.UsageMetrics.GPUSeconds).Quo(sdkmath.LegacyNewDec(3600))
		memHours := sdkmath.LegacyNewDec(record.UsageMetrics.MemoryGBSeconds).Quo(sdkmath.LegacyNewDec(3600))
		storageHours := sdkmath.LegacyNewDec(record.UsageMetrics.StorageGBHours)

		aggregation.TotalCPUCoreHours = aggregation.TotalCPUCoreHours.Add(cpuHours)
		aggregation.TotalGPUHours = aggregation.TotalGPUHours.Add(gpuHours)
		aggregation.TotalMemoryGBHours = aggregation.TotalMemoryGBHours.Add(memHours)
		aggregation.TotalStorageGBHours = aggregation.TotalStorageGBHours.Add(storageHours)
		aggregation.TotalNodeHours = aggregation.TotalNodeHours.Add(record.UsageMetrics.NodeHours)

		aggregation.TotalBillableAmount = aggregation.TotalBillableAmount.Add(record.BillableAmount...)

		// Sum discounts
		for _, d := range record.AppliedDiscounts {
			aggregation.TotalDiscounts = aggregation.TotalDiscounts.Add(d.DiscountAmount...)
		}

		return false
	})

	// Set default currency if no billable amount
	if aggregation.TotalBillableAmount.IsZero() {
		aggregation.TotalBillableAmount = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdkmath.ZeroInt()))
	}
	if aggregation.TotalDiscounts.Empty() {
		aggregation.TotalDiscounts = sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, sdkmath.ZeroInt()))
	}

	// Save aggregation
	if err := k.CreateAccountingAggregation(ctx, &aggregation); err != nil {
		return nil, err
	}

	return &aggregation, nil
}

// ============================================================================
// Spending Tracking
// ============================================================================

// GetCustomerSpendingInPeriod gets a customer's spending in a period
func (k Keeper) GetCustomerSpendingInPeriod(ctx sdk.Context, customerAddr sdk.AccAddress, start, end time.Time) sdk.Coins {
	spending := sdk.NewCoins()

	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.CustomerAddress != customerAddr.String() {
			return false
		}

		// Check if record is in period
		if record.PeriodEnd.Before(start) || record.PeriodStart.After(end) {
			return false
		}

		spending = spending.Add(record.BillableAmount...)
		return false
	})

	return spending
}

// GetProviderEarningsInPeriod gets a provider's earnings in a period
func (k Keeper) GetProviderEarningsInPeriod(ctx sdk.Context, providerAddr sdk.AccAddress, start, end time.Time) sdk.Coins {
	earnings := sdk.NewCoins()

	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.ProviderAddress != providerAddr.String() {
			return false
		}

		// Check if record is in period
		if record.PeriodEnd.Before(start) || record.PeriodStart.After(end) {
			return false
		}

		earnings = earnings.Add(record.ProviderReward...)
		return false
	})

	return earnings
}

// ============================================================================
// Pending Records Processing
// ============================================================================

// ProcessPendingAccountingRecords processes pending accounting records
func (k Keeper) ProcessPendingAccountingRecords(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	finalizationDelay := time.Duration(params.AccountingFinalizationDelaySec) * time.Second

	k.WithAccountingRecords(ctx, func(record types.HPCAccountingRecord) bool {
		if record.Status != types.AccountingStatusPending {
			return false
		}

		// Check if record is old enough to finalize
		if ctx.BlockTime().Sub(record.CreatedAt) < finalizationDelay {
			return false
		}

		// Finalize the record
		if err := k.FinalizeAccountingRecord(ctx, record.RecordID); err != nil {
			k.Logger(ctx).Error("failed to finalize accounting record",
				"record_id", record.RecordID, "error", err)
		}

		return false
	})

	return nil
}
