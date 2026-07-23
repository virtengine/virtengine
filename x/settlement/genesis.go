package settlement

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// InitGenesis initializes the settlement module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.IKeeper, data *types.GenesisState) {
	if !k.GetFiatConversionCustodyBalance(ctx).Equal(data.FiatConversionCustodyBalance) {
		panic(types.ErrInvalidSettlement.Wrap("fiat custody module-account balance does not match genesis accounting"))
	}
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
	if data.UsageAuthenticationActive {
		if err := k.ActivateUsageAuthentication(ctx); err != nil {
			panic(err)
		}
	}
	if data.FinancialCasesActive {
		k.ActivateFinancialCases(ctx)
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

	// Import payout records before financial cases so active case hold
	// validation can resolve every referenced payout.
	for _, payout := range data.PayoutRecords {
		if err := k.SetPayout(ctx, payout); err != nil {
			panic(err)
		}
	}

	// Import fiat payout preferences
	for _, pref := range data.FiatPayoutPreferences {
		if err := k.SetFiatPayoutPreference(ctx, pref); err != nil {
			panic(err)
		}
	}

	// Import fiat conversion records
	for _, conversion := range data.FiatConversionRecords {
		if err := k.ImportFiatConversion(ctx, conversion); err != nil {
			panic(err)
		}
	}

	for _, financialCase := range data.FinancialCases {
		if err := k.SetFinancialCase(ctx, financialCase); err != nil {
			panic(err)
		}
	}

	// Preserve explicit next sequences when supplied. Legacy documents omitted
	// them or used zero, in which case derive a safe next value from stored IDs.
	var maxEscrowSeq, maxSettlementSeq, maxUsageSeq, maxDistributionSeq, maxPayoutSeq, maxFiatConversionSeq uint64

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

	for _, payout := range data.PayoutRecords {
		seq := extractSequenceFromID(payout.PayoutID)
		if seq > maxPayoutSeq {
			maxPayoutSeq = seq
		}
	}

	for _, conversion := range data.FiatConversionRecords {
		seq := extractSequenceFromID(conversion.ConversionID)
		if seq > maxFiatConversionSeq {
			maxFiatConversionSeq = seq
		}
	}

	k.SetNextEscrowSequence(ctx, genesisNextSequence(data.EscrowSequence, maxEscrowSeq))
	k.SetNextSettlementSequence(ctx, genesisNextSequence(data.SettlementSequence, maxSettlementSeq))
	k.SetNextUsageSequence(ctx, genesisNextSequence(data.UsageSequence, maxUsageSeq))
	k.SetNextDistributionSequence(ctx, genesisNextSequence(data.DistributionSequence, maxDistributionSeq))
	k.SetNextPayoutSequence(ctx, genesisNextSequence(data.PayoutSequence, maxPayoutSeq))
	k.SetNextFiatConversionSequence(ctx, genesisNextSequence(data.FiatConversionSequence, maxFiatConversionSeq))
	if err := k.RebuildFinancialCaseState(ctx); err != nil {
		panic(err)
	}
	if needsFiatConversionMigration(data.FiatConversionRecords) {
		if _, err := k.MigrateFiatConversions(ctx); err != nil {
			panic(err)
		}
	} else if err := k.RebuildFiatConversionState(ctx); err != nil {
		panic(err)
	}
}

func needsFiatConversionMigration(records []types.FiatConversionRecord) bool {
	for _, record := range records {
		if record.ProtocolVersion == 0 {
			return true
		}
	}
	return false
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

	// Export fiat conversion records
	var conversions []types.FiatConversionRecord
	k.WithFiatConversions(ctx, func(conversion types.FiatConversionRecord) bool {
		conversions = append(conversions, conversion)
		return false
	})
	var payouts []types.PayoutRecord
	k.WithPayouts(ctx, func(payout types.PayoutRecord) bool {
		payouts = append(payouts, payout)
		return false
	})

	// Export fiat payout preferences
	var preferences []types.FiatPayoutPreference
	k.WithFiatPayoutPreferences(ctx, func(pref types.FiatPayoutPreference) bool {
		preferences = append(preferences, pref)
		return false
	})
	var financialCases []types.FinancialCase
	if err := k.WithFinancialCases(ctx, func(financialCase types.FinancialCase) bool {
		financialCases = append(financialCases, financialCase)
		return false
	}); err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:                       params,
		EscrowAccounts:               escrows,
		SettlementRecords:            settlements,
		UsageRecords:                 usageRecords,
		RewardDistributions:          rewardDistributions,
		ClaimableRewards:             claimableRewards,
		PayoutRecords:                payouts,
		FiatConversionRecords:        conversions,
		FiatPayoutPreferences:        preferences,
		UsageAuthenticationActive:    k.IsUsageAuthenticationActive(ctx),
		FinancialCases:               financialCases,
		FinancialCasesActive:         k.IsFinancialCasesActive(ctx),
		EscrowSequence:               k.GetEscrowSequence(ctx),
		SettlementSequence:           k.GetSettlementSequence(ctx),
		DistributionSequence:         k.GetDistributionSequence(ctx),
		UsageSequence:                k.GetUsageSequence(ctx),
		PayoutSequence:               k.GetPayoutSequence(ctx),
		FiatConversionSequence:       k.GetFiatConversionSequence(ctx),
		FiatConversionCustodyBalance: k.GetFiatConversionCustodyBalance(ctx),
	}
}

func genesisNextSequence(explicit, maxID uint64) uint64 {
	derived := uint64(1)
	if maxID > 0 {
		derived = maxID + 1
	}
	if explicit > derived {
		return explicit
	}
	return derived
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
