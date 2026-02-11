// Package keeper implements the HPC module keeper.
//
// VE-5A: Settlement integration for escrow and invoice pipeline
package keeper

import (
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Settlement Pipeline
// ============================================================================

// SettlementRequest represents a request to settle an HPC job
type SettlementRequest struct {
	// JobID is the job to settle
	JobID string `json:"job_id"`

	// AccountingRecordID is the accounting record to settle
	AccountingRecordID string `json:"accounting_record_id"`

	// EscrowID is the escrow account to settle from
	EscrowID string `json:"escrow_id"`

	// RequestedBy is who requested settlement
	RequestedBy string `json:"requested_by"`

	// RequestedAt is when settlement was requested
	RequestedAt time.Time `json:"requested_at"`
}

// SettlementResult represents the result of a settlement
type SettlementResult struct {
	// Success indicates if settlement succeeded
	Success bool `json:"success"`

	// SettlementID is the settlement record ID
	SettlementID string `json:"settlement_id"`

	// JobID is the job that was settled
	JobID string `json:"job_id"`

	// AccountingRecordID is the accounting record
	AccountingRecordID string `json:"accounting_record_id"`

	// CustomerPaid is what the customer paid
	CustomerPaid sdk.Coins `json:"customer_paid"`

	// ProviderReceived is what the provider received
	ProviderReceived sdk.Coins `json:"provider_received"`

	// PlatformFee is the platform fee collected
	PlatformFee sdk.Coins `json:"platform_fee"`

	// RefundAmount is any refund to customer
	RefundAmount sdk.Coins `json:"refund_amount"`

	// Error contains error message if failed
	Error string `json:"error,omitempty"`

	// SettledAt is when settlement completed
	SettledAt time.Time `json:"settled_at"`
}

// ProcessJobSettlement processes settlement for a completed job
func (k Keeper) ProcessJobSettlement(ctx sdk.Context, jobID string) (*SettlementResult, error) {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return nil, types.ErrJobNotFound
	}

	// Verify job is in terminal state
	if !types.IsTerminalJobState(job.State) {
		return nil, types.ErrInvalidJobState.Wrap("job must be in terminal state for settlement")
	}

	// Get or create accounting record
	records := k.GetAccountingRecordsByJob(ctx, jobID)
	var record *types.HPCAccountingRecord
	if len(records) == 0 {
		// Calculate billing for the job
		newRecord, err := k.CalculateJobBilling(ctx, jobID)
		if err != nil {
			return &SettlementResult{
				Success: false,
				JobID:   jobID,
				Error:   fmt.Sprintf("failed to calculate billing: %v", err),
			}, err
		}
		record = newRecord
	} else {
		record = &records[0]
	}

	// Ensure record is finalized
	if record.Status == types.AccountingStatusPending {
		if err := k.FinalizeAccountingRecord(ctx, record.RecordID); err != nil {
			return nil, err
		}
		// Reload record
		*record, _ = k.GetAccountingRecord(ctx, record.RecordID)
	}

	// Check if already settled
	if record.Status == types.AccountingStatusSettled {
		return &SettlementResult{
			Success:            true,
			SettlementID:       record.SettlementID,
			JobID:              jobID,
			AccountingRecordID: record.RecordID,
			CustomerPaid:       record.BillableAmount,
			ProviderReceived:   record.ProviderReward,
			PlatformFee:        record.PlatformFee,
		}, nil
	}

	// Check if disputed
	if record.Status == types.AccountingStatusDisputed {
		return &SettlementResult{
			Success:            false,
			JobID:              jobID,
			AccountingRecordID: record.RecordID,
			Error:              "cannot settle disputed record",
		}, types.ErrInvalidJobAccounting.Wrap("record is disputed")
	}

	// Perform settlement
	result, err := k.executeSettlement(ctx, &job, record)
	if err != nil {
		return result, err
	}

	// Update accounting record
	if err := k.SettleAccountingRecord(ctx, record.RecordID, result.SettlementID); err != nil {
		return nil, err
	}

	if k.billingEnabled() {
		if _, err := k.GenerateInvoiceForJob(ctx, record.RecordID); err != nil {
			return nil, err
		}
	}

	// Distribute rewards
	if _, err := k.DistributeJobRewardsFromSettlement(ctx, jobID, record); err != nil {
		k.Logger(ctx).Error("failed to distribute rewards", "job_id", jobID, "error", err)
		// Continue - settlement succeeded even if reward distribution failed
	}

	k.Logger(ctx).Info("settled HPC job",
		"job_id", jobID,
		"settlement_id", result.SettlementID,
		"customer_paid", result.CustomerPaid.String(),
		"provider_received", result.ProviderReceived.String())

	return result, nil
}

// executeSettlement executes the actual settlement
func (k Keeper) executeSettlement(ctx sdk.Context, job *types.HPCJob, record *types.HPCAccountingRecord) (*SettlementResult, error) {
	result := &SettlementResult{
		JobID:              job.JobID,
		AccountingRecordID: record.RecordID,
		RefundAmount:       sdk.NewCoins(),
		SettledAt:          ctx.BlockTime(),
	}

	if k.settlementKeeper == nil {
		result.Success = false
		result.Error = "settlement keeper not configured"
		return result, types.ErrInvalidJobAccounting.Wrap(result.Error)
	}

	orderID, escrow, found := k.resolveSettlementEscrow(ctx, job)
	if !found {
		result.Success = false
		result.Error = "settlement escrow not found"
		return result, types.ErrInvalidJobAccounting.Wrap(result.Error)
	}

	usageRecords, err := k.buildSettlementUsageRecords(ctx, job, record, orderID, escrow)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to build settlement usage records: %v", err)
		return result, err
	}

	usageIDs := make([]string, 0, len(usageRecords))
	for i := range usageRecords {
		if err := k.settlementKeeper.RecordUsage(ctx, &usageRecords[i]); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to record usage: %v", err)
			return result, err
		}
		usageIDs = append(usageIDs, usageRecords[i].UsageID)
	}

	settlementRecord, err := k.settlementKeeper.SettleOrder(ctx, orderID, usageIDs, true)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to settle order: %v", err)
		return result, err
	}

	// Calculate refund if agreed price exceeds billable
	if job.AgreedPrice.IsAllGTE(record.BillableAmount) {
		refund := job.AgreedPrice.Sub(record.BillableAmount...)
		if !refund.IsZero() {
			result.RefundAmount = refund
		}
	}

	result.SettlementID = settlementRecord.SettlementID
	result.CustomerPaid = settlementRecord.TotalAmount
	result.ProviderReceived = settlementRecord.ProviderShare
	result.PlatformFee = settlementRecord.PlatformFee

	// Emit settlement event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_settlement",
			sdk.NewAttribute("settlement_id", settlementRecord.SettlementID),
			sdk.NewAttribute("job_id", job.JobID),
			sdk.NewAttribute("accounting_record_id", record.RecordID),
			sdk.NewAttribute("customer", job.CustomerAddress),
			sdk.NewAttribute("provider", job.ProviderAddress),
			sdk.NewAttribute("customer_paid", settlementRecord.TotalAmount.String()),
			sdk.NewAttribute("provider_received", settlementRecord.ProviderShare.String()),
			sdk.NewAttribute("platform_fee", settlementRecord.PlatformFee.String()),
			sdk.NewAttribute("refund", result.RefundAmount.String()),
		),
	)

	result.Success = true
	return result, nil
}

func (k Keeper) resolveSettlementEscrow(ctx sdk.Context, job *types.HPCJob) (string, settlementtypes.EscrowAccount, bool) {
	orderID := job.JobID
	if k.settlementKeeper == nil {
		return orderID, settlementtypes.EscrowAccount{}, false
	}

	if escrow, found := k.settlementKeeper.GetEscrowByOrder(ctx, orderID); found {
		return orderID, escrow, true
	}

	if job.EscrowID != "" {
		if escrow, found := k.settlementKeeper.GetEscrow(ctx, job.EscrowID); found {
			return escrow.OrderID, escrow, true
		}
	}

	return orderID, settlementtypes.EscrowAccount{}, false
}

func (k Keeper) buildSettlementUsageRecords(
	ctx sdk.Context,
	job *types.HPCJob,
	record *types.HPCAccountingRecord,
	orderID string,
	escrow settlementtypes.EscrowAccount,
) ([]settlementtypes.UsageRecord, error) {
	metrics := record.UsageMetrics
	usageComponents := buildUsageComponents(metrics, record.BillableBreakdown)

	targetCoin := selectBillableCoin(record.BillableAmount, usageComponents)
	usageComponents = scaleComponentsToTarget(targetCoin, usageComponents)

	baseMetadata := map[string]string{
		"hpc_job_id":            job.JobID,
		"hpc_cluster_id":        record.ClusterID,
		"hpc_offering_id":       record.OfferingID,
		"hpc_accounting_id":     record.RecordID,
		"hpc_scheduler_type":    record.SchedulerType,
		"hpc_formula_version":   record.FormulaVersion,
		"hpc_job_state":         string(job.State),
		"hpc_job_escrow_id":     job.EscrowID,
		"hpc_cpu_core_seconds":  strconv.FormatInt(metrics.CPUCoreSeconds, 10),
		"hpc_memory_gb_seconds": strconv.FormatInt(metrics.MemoryGBSeconds, 10),
		"hpc_storage_gb_hours":  strconv.FormatInt(metrics.StorageGBHours, 10),
		"hpc_network_bytes_in":  strconv.FormatInt(metrics.NetworkBytesIn, 10),
		"hpc_network_bytes_out": strconv.FormatInt(metrics.NetworkBytesOut, 10),
		"hpc_node_hours":        metrics.NodeHours.String(),
		"hpc_gpu_seconds":       strconv.FormatInt(metrics.GPUSeconds, 10),
		"hpc_gpu_type":          metrics.GPUType,
	}

	if escrow.EscrowID != "" {
		baseMetadata["settlement_escrow_id"] = escrow.EscrowID
	}
	if escrow.LeaseID != "" {
		baseMetadata["settlement_lease_id"] = escrow.LeaseID
	}

	if record.CalculationHash != "" {
		baseMetadata["hpc_calculation_hash"] = record.CalculationHash
	}

	periodStart := record.PeriodStart
	periodEnd := record.PeriodEnd
	if periodEnd.Before(periodStart) {
		periodEnd = periodStart
	}

	signature := []byte(record.CalculationHash)
	if len(signature) == 0 {
		signature = []byte(record.RecordID)
	}

	usageRecords := make([]settlementtypes.UsageRecord, 0, len(usageComponents))
	for _, component := range usageComponents {
		if component.cost.IsNil() || !component.cost.IsPositive() {
			continue
		}
		usageUnits := component.units
		if usageUnits == 0 {
			usageUnits = 1
		}

		unitPrice := sdk.NewDecCoinFromDec(
			component.cost.Denom,
			sdkmath.LegacyNewDecFromInt(component.cost.Amount).QuoInt64(int64(usageUnits)),
		)

		metadata := copyStringMap(baseMetadata)
		metadata["hpc_usage_type"] = component.usageType
		metadata["hpc_usage_units"] = strconv.FormatUint(usageUnits, 10)

		usageRecords = append(usageRecords, settlementtypes.UsageRecord{
			UsageID:           "",
			OrderID:           orderID,
			LeaseID:           job.JobID,
			Provider:          record.ProviderAddress,
			Customer:          record.CustomerAddress,
			UsageUnits:        usageUnits,
			UsageType:         component.usageType,
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			UnitPrice:         unitPrice,
			TotalCost:         sdk.NewCoins(component.cost),
			ProviderSignature: signature,
			SubmittedAt:       ctx.BlockTime(),
			BlockHeight:       ctx.BlockHeight(),
			Metadata:          metadata,
		})
	}

	if len(usageRecords) == 0 && targetCoin.IsPositive() {
		usageRecords = append(usageRecords, settlementtypes.UsageRecord{
			UsageID:           "",
			OrderID:           orderID,
			LeaseID:           job.JobID,
			Provider:          record.ProviderAddress,
			Customer:          record.CustomerAddress,
			UsageUnits:        1,
			UsageType:         "fixed",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			UnitPrice:         sdk.NewDecCoinFromCoin(targetCoin),
			TotalCost:         sdk.NewCoins(targetCoin),
			ProviderSignature: signature,
			SubmittedAt:       ctx.BlockTime(),
			BlockHeight:       ctx.BlockHeight(),
			Metadata:          baseMetadata,
		})
	}

	if len(usageRecords) == 0 {
		return nil, types.ErrInvalidJobAccounting.Wrap("no billable usage to settle")
	}

	return usageRecords, nil
}

type usageComponent struct {
	usageType string
	units     uint64
	cost      sdk.Coin
}

func buildUsageComponents(metrics types.HPCDetailedMetrics, breakdown types.BillableBreakdown) []usageComponent {
	var components []usageComponent

	components = appendUsageComponent(components, "cpu_core_hours", usageUnitsFromSeconds(metrics.CPUCoreSeconds), breakdown.CPUCost)
	components = appendUsageComponent(components, "memory_gb_hours", usageUnitsFromSeconds(metrics.MemoryGBSeconds), breakdown.MemoryCost)
	components = appendUsageComponent(components, "gpu_hours", usageUnitsFromSeconds(metrics.GPUSeconds), breakdown.GPUCost)
	components = appendUsageComponent(components, "node_hours", usageUnitsFromDec(metrics.NodeHours), breakdown.NodeCost)
	components = appendUsageComponent(components, "storage_gb_hours", usageUnitsFromInt64(metrics.StorageGBHours), breakdown.StorageCost)

	networkBytes := metrics.NetworkBytesIn + metrics.NetworkBytesOut
	components = appendUsageComponent(components, "network_gb", usageUnitsFromBytes(networkBytes), breakdown.NetworkCost)

	if breakdown.QueuePenalty.IsPositive() {
		components = appendUsageComponent(components, "queue_penalty", 1, breakdown.QueuePenalty)
	}

	return components
}

func appendUsageComponent(components []usageComponent, usageType string, units uint64, cost sdk.Coin) []usageComponent {
	if !cost.IsPositive() {
		return components
	}
	if units == 0 {
		units = 1
	}
	return append(components, usageComponent{
		usageType: usageType,
		units:     units,
		cost:      cost,
	})
}

func selectBillableCoin(amount sdk.Coins, components []usageComponent) sdk.Coin {
	if len(amount) > 0 {
		return amount[0]
	}
	for _, component := range components {
		if component.cost.IsPositive() {
			return component.cost
		}
	}
	return sdk.Coin{}
}

func scaleComponentsToTarget(target sdk.Coin, components []usageComponent) []usageComponent {
	if !target.IsPositive() {
		return components
	}

	total := sdkmath.NewInt(0)
	for _, component := range components {
		if component.cost.Denom == target.Denom {
			total = total.Add(component.cost.Amount)
		}
	}

	if total.IsZero() {
		return []usageComponent{
			{
				usageType: "fixed",
				units:     1,
				cost:      target,
			},
		}
	}

	if total.Equal(target.Amount) {
		return components
	}

	ratio := sdkmath.LegacyNewDecFromInt(target.Amount).Quo(sdkmath.LegacyNewDecFromInt(total))
	adjusted := make([]usageComponent, 0, len(components))
	sumAdjusted := sdkmath.NewInt(0)
	lastIdx := -1

	for idx, component := range components {
		if component.cost.Denom != target.Denom {
			adjusted = append(adjusted, component)
			continue
		}
		amount := sdkmath.LegacyNewDecFromInt(component.cost.Amount).Mul(ratio).TruncateInt()
		adjustedCost := sdk.NewCoin(component.cost.Denom, amount)
		adjusted = append(adjusted, usageComponent{
			usageType: component.usageType,
			units:     component.units,
			cost:      adjustedCost,
		})
		sumAdjusted = sumAdjusted.Add(amount)
		lastIdx = idx
	}

	if lastIdx >= 0 && !sumAdjusted.Equal(target.Amount) {
		remainder := target.Amount.Sub(sumAdjusted)
		for idx := range adjusted {
			if adjusted[idx].usageType == components[lastIdx].usageType && adjusted[idx].cost.Denom == target.Denom {
				adjusted[idx].cost = adjusted[idx].cost.Add(sdk.NewCoin(target.Denom, remainder))
				break
			}
		}
	}

	return adjusted
}

func usageUnitsFromSeconds(seconds int64) uint64 {
	if seconds <= 0 {
		return 0
	}
	hours := seconds / 3600
	if seconds%3600 != 0 {
		hours++
	}
	if hours <= 0 {
		hours = 1
	}
	return uint64(hours)
}

func usageUnitsFromInt64(value int64) uint64 {
	if value <= 0 {
		return 0
	}
	return uint64(value)
}

func usageUnitsFromBytes(bytes int64) uint64 {
	if bytes <= 0 {
		return 0
	}
	const gb = int64(1024 * 1024 * 1024)
	gbUnits := bytes / gb
	if bytes%gb != 0 {
		gbUnits++
	}
	if gbUnits <= 0 {
		gbUnits = 1
	}
	return uint64(gbUnits)
}

func usageUnitsFromDec(value sdkmath.LegacyDec) uint64 {
	if value.IsNil() || value.IsNegative() || value.IsZero() {
		return 0
	}
	truncated := value.TruncateInt64()
	if value.Equal(sdkmath.LegacyNewDec(truncated)) {
		return uint64(truncated)
	}
	return uint64(truncated + 1)
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// DistributeJobRewardsFromSettlement distributes rewards after settlement
func (k Keeper) DistributeJobRewardsFromSettlement(ctx sdk.Context, jobID string, record *types.HPCAccountingRecord) (*types.HPCRewardRecord, error) {
	job, exists := k.GetJob(ctx, jobID)
	if !exists {
		return nil, types.ErrJobNotFound
	}

	// Get scheduling decision for node information
	// Currently clusters don't track individual node IDs, so provider gets full reward
	// TODO: Implement node-level reward distribution when node tracking is added
	var nodeRewards []types.HPCRewardRecipient

	// If no node-level breakdown, provider gets full reward
	if len(nodeRewards) == 0 {
		nodeRewards = []types.HPCRewardRecipient{
			{
				Address:            record.ProviderAddress,
				Amount:             record.ProviderReward,
				RecipientType:      "provider",
				ContributionWeight: "1.000000",
				Reason:             "Full job completion reward",
			},
		}
	}

	// Create reward record
	rewardRecord := &types.HPCRewardRecord{
		JobID:                  jobID,
		ClusterID:              job.ClusterID,
		Source:                 types.HPCRewardSourceJobCompletion,
		TotalReward:            record.ProviderReward,
		Recipients:             nodeRewards,
		ReferencedUsageRecords: record.SignedUsageRecords,
		JobCompletionStatus:    job.State,
		FormulaVersion:         record.FormulaVersion,
		CalculationDetails: types.RewardCalculationDetails{
			TotalUsageValue:         record.BillableAmount.String(),
			RewardPoolContribution:  record.ProviderReward.String(),
			PlatformFeeRate:         fmt.Sprintf("%d", k.GetParams(ctx).PlatformFeeRateBps),
			NodeContributionFormula: "proportional_time_weighted",
			InputMetrics: map[string]string{
				"wall_clock_seconds": fmt.Sprintf("%d", record.UsageMetrics.WallClockSeconds),
				"cpu_core_seconds":   fmt.Sprintf("%d", record.UsageMetrics.CPUCoreSeconds),
				"nodes_used":         fmt.Sprintf("%d", record.UsageMetrics.NodesUsed),
			},
		},
		IssuedAt: ctx.BlockTime(),
	}

	// Store reward record using existing method
	if err := k.SetHPCReward(ctx, *rewardRecord); err != nil {
		return nil, err
	}

	return rewardRecord, nil
}

// calculateNodeRewards calculates rewards for each node based on contribution
//
//nolint:unused // reserved for future per-node reward distribution
func (k Keeper) calculateNodeRewards(ctx sdk.Context, totalReward sdk.Coins, nodeIDs []string) []types.HPCRewardRecipient {
	if len(nodeIDs) == 0 {
		return nil
	}

	recipients := make([]types.HPCRewardRecipient, 0, len(nodeIDs))

	// For now, distribute equally among nodes
	// TODO: Implement weighted distribution based on actual node contribution
	nodeCount := int64(len(nodeIDs))

	for _, nodeID := range nodeIDs {
		// Get node metadata for owner address
		node, exists := k.GetNodeMetadata(ctx, nodeID)
		if !exists {
			continue
		}

		// Calculate share
		nodeReward := sdk.NewCoins()
		for _, coin := range totalReward {
			share := coin.Amount.Quo(sdkmath.NewInt(nodeCount))
			if share.IsPositive() {
				nodeReward = nodeReward.Add(sdk.NewCoin(coin.Denom, share))
			}
		}

		weight := fmt.Sprintf("%.6f", 1.0/float64(nodeCount))

		recipients = append(recipients, types.HPCRewardRecipient{
			Address:            node.ProviderAddress,
			Amount:             nodeReward,
			RecipientType:      "node_operator",
			NodeID:             nodeID,
			ContributionWeight: weight,
			Reason:             "Node contribution to job execution",
		})
	}

	return recipients
}

// ============================================================================
// Settlement Hooks for EndBlocker
// ============================================================================

// ProcessPendingSettlements processes pending settlements in EndBlocker
func (k Keeper) ProcessPendingSettlements(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Find jobs that are complete but not settled
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		// Only process terminal jobs
		if !types.IsTerminalJobState(job.State) {
			return false
		}

		// Check if job has accounting records that are finalized but not settled
		records := k.GetAccountingRecordsByJob(ctx, job.JobID)
		for _, record := range records {
			if record.Status == types.AccountingStatusFinalized {
				// Check if settlement delay has passed
				if record.FinalizedAt != nil {
					delay := time.Duration(params.SettlementDelaySec) * time.Second
					if ctx.BlockTime().Sub(*record.FinalizedAt) >= delay {
						// Process settlement
						if _, err := k.ProcessJobSettlement(ctx, job.JobID); err != nil {
							k.Logger(ctx).Error("failed to process settlement",
								"job_id", job.JobID, "error", err)
						}
					}
				}
			}
		}

		return false
	})

	return nil
}

// ============================================================================
// Dispute-Related Settlement
// ============================================================================

// ProcessDisputeResolution handles settlement after dispute resolution
func (k Keeper) ProcessDisputeResolution(ctx sdk.Context, disputeID string) error {
	dispute, exists := k.GetDispute(ctx, disputeID)
	if !exists {
		return types.ErrInvalidDispute.Wrap("dispute not found")
	}

	if dispute.Status != types.DisputeStatusResolved {
		return types.ErrInvalidDispute.Wrap("dispute not resolved")
	}

	// Get accounting record
	records := k.GetAccountingRecordsByJob(ctx, dispute.JobID)
	for _, record := range records {
		if record.DisputeID == disputeID {
			// Update record status and process settlement
			record.Status = types.AccountingStatusFinalized
			record.DisputeID = ""
			if err := k.SetAccountingRecord(ctx, record); err != nil {
				return err
			}

			// Process settlement
			if _, err := k.ProcessJobSettlement(ctx, dispute.JobID); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// CreateCorrectionRecord creates a correction accounting record after dispute
func (k Keeper) CreateCorrectionRecord(ctx sdk.Context, originalRecordID string, correctedMetrics *types.HPCDetailedMetrics, reason string) (*types.HPCAccountingRecord, error) {
	original, exists := k.GetAccountingRecord(ctx, originalRecordID)
	if !exists {
		return nil, types.ErrInvalidJobAccounting.Wrap("original record not found")
	}

	// Get billing rules
	providerAddr, _ := sdk.AccAddressFromBech32(original.ProviderAddress)
	rules := k.GetOrDefaultBillingRules(ctx, providerAddr)

	// Calculate new billing
	calculator := types.NewHPCBillingCalculator(rules)
	breakdown, billable, err := calculator.CalculateBillableAmount(correctedMetrics, nil, nil)
	if err != nil {
		return nil, err
	}

	providerReward := calculator.CalculateProviderReward(billable)
	platformFee := calculator.CalculatePlatformFee(billable)

	// Create correction record
	correction := &types.HPCAccountingRecord{
		JobID:              original.JobID,
		ClusterID:          original.ClusterID,
		ProviderAddress:    original.ProviderAddress,
		CustomerAddress:    original.CustomerAddress,
		OfferingID:         original.OfferingID,
		SchedulerType:      original.SchedulerType,
		UsageMetrics:       *correctedMetrics,
		BillableAmount:     billable,
		BillableBreakdown:  *breakdown,
		ProviderReward:     providerReward,
		PlatformFee:        platformFee,
		Status:             types.AccountingStatusCorrected,
		CorrectedFromID:    originalRecordID,
		CorrectionReason:   reason,
		PeriodStart:        original.PeriodStart,
		PeriodEnd:          original.PeriodEnd,
		FormulaVersion:     rules.FormulaVersion,
		SignedUsageRecords: original.SignedUsageRecords,
	}

	if err := k.CreateAccountingRecord(ctx, correction); err != nil {
		return nil, err
	}

	// Mark original as corrected
	original.Status = types.AccountingStatusCorrected
	if err := k.SetAccountingRecord(ctx, original); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("created correction record",
		"original_id", originalRecordID,
		"correction_id", correction.RecordID,
		"reason", reason)

	return correction, nil
}
