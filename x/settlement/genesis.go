package settlement

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// InitGenesis initializes the settlement module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.IKeeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Import escrow accounts
	for _, escrow := range data.EscrowAccounts {
		if err := k.SetEscrow(ctx, escrow); err != nil {
			panic(err)
		}
	}

	// Import settlement records
	for _, settlement := range data.SettlementRecords {
		if err := k.SetSettlement(ctx, settlement); err != nil {
			panic(err)
		}
	}

	// Import usage records
	for _, usage := range data.UsageRecords {
		if err := k.SetUsageRecord(ctx, usage); err != nil {
			panic(err)
		}
	}

	// Import reward distributions
	for _, distribution := range data.RewardDistributions {
		if err := k.SetRewardDistribution(ctx, distribution); err != nil {
			panic(err)
		}
	}

	// Import claimable rewards
	for _, claimable := range data.ClaimableRewards {
		addr, err := sdk.AccAddressFromBech32(claimable.Address)
		if err != nil {
			panic(err)
		}
		if err := k.SetClaimableRewards(ctx, addr, claimable); err != nil {
			panic(err)
		}
	}

	// Set the next sequences from the highest existing IDs
	var maxEscrowSeq, maxSettlementSeq, maxUsageSeq, maxDistributionSeq uint64

	for _, escrow := range data.EscrowAccounts {
		seq := extractSequenceFromID(escrow.EscrowID)
		if seq > maxEscrowSeq {
			maxEscrowSeq = seq
		}
	}

	for _, settlement := range data.SettlementRecords {
		seq := extractSequenceFromID(settlement.SettlementID)
		if seq > maxSettlementSeq {
			maxSettlementSeq = seq
		}
	}

	for _, usage := range data.UsageRecords {
		seq := extractSequenceFromID(usage.UsageID)
		if seq > maxUsageSeq {
			maxUsageSeq = seq
		}
	}

	for _, dist := range data.RewardDistributions {
		seq := extractSequenceFromID(dist.DistributionID)
		if seq > maxDistributionSeq {
			maxDistributionSeq = seq
		}
	}

	// Set sequences (they will be incremented before use)
	if maxEscrowSeq > 0 {
		k.SetNextEscrowSequence(ctx, maxEscrowSeq+1)
	}
	if maxSettlementSeq > 0 {
		k.SetNextSettlementSequence(ctx, maxSettlementSeq+1)
	}
	if maxUsageSeq > 0 {
		k.SetNextUsageSequence(ctx, maxUsageSeq+1)
	}
	if maxDistributionSeq > 0 {
		k.SetNextDistributionSequence(ctx, maxDistributionSeq+1)
	}
}

// ExportGenesis returns the settlement module's genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.IKeeper) *types.GenesisState {
	params := k.GetParams(ctx)

	// Export all escrows
	var escrows []types.EscrowAccount
	k.WithEscrows(ctx, func(escrow types.EscrowAccount) bool {
		escrows = append(escrows, escrow)
		return false
	})

	// Export all settlements
	var settlements []types.SettlementRecord
	k.WithSettlements(ctx, func(settlement types.SettlementRecord) bool {
		settlements = append(settlements, settlement)
		return false
	})

	// Export all usage records
	var usageRecords []types.UsageRecord
	k.WithUsageRecords(ctx, func(usage types.UsageRecord) bool {
		usageRecords = append(usageRecords, usage)
		return false
	})

	// Export all reward distributions
	var rewardDistributions []types.RewardDistribution
	k.WithRewardDistributions(ctx, func(dist types.RewardDistribution) bool {
		rewardDistributions = append(rewardDistributions, dist)
		return false
	})

	// Export all claimable rewards
	var claimableRewards []types.ClaimableRewards
	k.WithClaimableRewards(ctx, func(rewards types.ClaimableRewards) bool {
		claimableRewards = append(claimableRewards, rewards)
		return false
	})

	return &types.GenesisState{
		Params:              params,
		EscrowAccounts:      escrows,
		SettlementRecords:   settlements,
		UsageRecords:        usageRecords,
		RewardDistributions: rewardDistributions,
		ClaimableRewards:    claimableRewards,
	}
}

// extractSequenceFromID extracts the numeric sequence from an ID string
// Expected format: "prefix-<sequence>" e.g., "escrow-123"
func extractSequenceFromID(id string) uint64 {
	var seq uint64
	// Simple extraction - find last dash and parse number after it
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '-' {
			// Parse the number after the dash
			numStr := id[i+1:]
			for _, c := range numStr {
				if c >= '0' && c <= '9' {
					seq = seq*10 + uint64(c-'0')
				}
			}
			break
		}
	}
	return seq
}
