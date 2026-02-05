// Package keeper implements the HPC module keeper.
//
// VE-5A: Settlement integration for escrow and invoice pipeline
package keeper

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		CustomerPaid:       record.BillableAmount,
		ProviderReceived:   record.ProviderReward,
		PlatformFee:        record.PlatformFee,
		RefundAmount:       sdk.NewCoins(),
		SettledAt:          ctx.BlockTime(),
	}

	// Generate settlement ID
	result.SettlementID = fmt.Sprintf("hpc-settle-%s", record.RecordID)

	// If job has an escrow ID, process escrow settlement
	if job.EscrowID != "" {
		// Calculate refund if agreed price exceeds billable
		if job.AgreedPrice.IsAllGTE(record.BillableAmount) {
			refund := job.AgreedPrice.Sub(record.BillableAmount...)
			if !refund.IsZero() {
				result.RefundAmount = refund
			}
		}
	}

	// Transfer provider reward
	providerAddr, err := sdk.AccAddressFromBech32(job.ProviderAddress)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("invalid provider address: %v", err)
		return result, err
	}

	// Transfer from module to provider
	if !record.ProviderReward.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, providerAddr, record.ProviderReward); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to transfer provider reward: %v", err)
			return result, err
		}
	}

	// Emit settlement event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_settlement",
			sdk.NewAttribute("settlement_id", result.SettlementID),
			sdk.NewAttribute("job_id", job.JobID),
			sdk.NewAttribute("accounting_record_id", record.RecordID),
			sdk.NewAttribute("customer", job.CustomerAddress),
			sdk.NewAttribute("provider", job.ProviderAddress),
			sdk.NewAttribute("customer_paid", result.CustomerPaid.String()),
			sdk.NewAttribute("provider_received", result.ProviderReceived.String()),
			sdk.NewAttribute("platform_fee", result.PlatformFee.String()),
			sdk.NewAttribute("refund", result.RefundAmount.String()),
		),
	)

	result.Success = true
	return result, nil
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
