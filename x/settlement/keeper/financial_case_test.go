package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestFinancialCaseDeterministicIDs(t *testing.T) {
	subject := types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: "order/7", OrderId: "order/7"}
	id1, err := DeterministicFinancialCaseID(subject)
	require.NoError(t, err)
	id2, err := DeterministicFinancialCaseID(subject)
	require.NoError(t, err)
	require.Equal(t, id1, id2)
	require.Contains(t, id1, "financial-case/")

	claim := types.FinancialClaim{ClaimType: types.FinancialClaimTypeFraud, Claimant: testAddress(1), SourceModule: "fraud", SourceReference: "fraud-report-1", EvidenceHash: bytes.Repeat([]byte{1}, sha256.Size), IdempotencyKey: []byte("retry-1")}
	claimID1, payload1, err := DeterministicFinancialClaimID(id1, claim)
	require.NoError(t, err)
	claimID2, payload2, err := DeterministicFinancialClaimID(id1, claim)
	require.NoError(t, err)
	require.Equal(t, claimID1, claimID2)
	require.Equal(t, payload1, payload2)

	claim.EvidenceHash[0]++
	claimID3, payload3, err := DeterministicFinancialClaimID(id1, claim)
	require.NoError(t, err)
	require.NotEqual(t, claimID1, claimID3)
	require.NotEqual(t, payload1, payload3)
}

func TestFinancialCaseTransitionTable(t *testing.T) {
	tests := []struct {
		from, to types.FinancialCaseStatus
		allowed  bool
	}{
		{types.FinancialCaseStatusOpen, types.FinancialCaseStatusEvidence, true},
		{types.FinancialCaseStatusOpen, types.FinancialCaseStatusReview, true},
		{types.FinancialCaseStatusEvidence, types.FinancialCaseStatusReview, true},
		{types.FinancialCaseStatusReview, types.FinancialCaseStatusEscalated, true},
		{types.FinancialCaseStatusReview, types.FinancialCaseStatusResolvedPendingAppeal, true},
		{types.FinancialCaseStatusEscalated, types.FinancialCaseStatusResolvedPendingAppeal, true},
		{types.FinancialCaseStatusResolvedPendingAppeal, types.FinancialCaseStatusReview, true},
		{types.FinancialCaseStatusResolvedPendingAppeal, types.FinancialCaseStatusFinal, true},
		{types.FinancialCaseStatusFinal, types.FinancialCaseStatusReview, false},
		{types.FinancialCaseStatusCancelled, types.FinancialCaseStatusOpen, false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.allowed, CanTransitionFinancialCase(tt.from, tt.to), "%s -> %s", tt.from, tt.to)
	}
}

func TestValidateTerminalAllocationMultiDenomConservation(t *testing.T) {
	exposure := sdk.NewCoins(
		sdk.NewCoin("uatom", sdkmath.NewInt(100)),
		sdk.NewCoin("uve", sdkmath.NewInt(1000)),
	)
	valid := types.TerminalAllocation{
		OriginalExposure:      exposure,
		Provider:              sdk.NewCoins(sdk.NewInt64Coin("uatom", 60), sdk.NewInt64Coin("uve", 600)),
		Customer:              sdk.NewCoins(sdk.NewInt64Coin("uatom", 30), sdk.NewInt64Coin("uve", 300)),
		Platform:              sdk.NewCoins(sdk.NewInt64Coin("uatom", 10), sdk.NewInt64Coin("uve", 80)),
		SlashWitness:          sdk.NewCoins(sdk.NewInt64Coin("uve", 20)),
		SlashWitnessRecipient: testAddress(7),
		ResolutionType:        types.FinancialResolutionPartialSplit,
	}
	require.NoError(t, ValidateTerminalAllocation(exposure, valid))

	invalid := valid
	invalid.Provider = sdk.NewCoins(sdk.NewInt64Coin("uatom", 61), sdk.NewInt64Coin("uve", 600))
	require.Error(t, ValidateTerminalAllocation(exposure, invalid))

	wrongOriginal := valid
	wrongOriginal.OriginalExposure = sdk.NewCoins(sdk.NewInt64Coin("uve", 1000))
	require.Error(t, ValidateTerminalAllocation(exposure, wrongOriginal))
}

func TestFinancialAllocationConservationProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(84)) //nolint:gosec // deterministic property-test input, not security randomness
	for i := 0; i < 1000; i++ {
		totalA := int64(rng.Intn(1_000_000) + 4)
		totalB := int64(rng.Intn(1_000_000) + 4)
		split := func(total int64) [4]int64 {
			a := int64(rng.Intn(int(total + 1)))
			b := int64(rng.Intn(int(total - a + 1)))
			c := int64(rng.Intn(int(total - a - b + 1)))
			return [4]int64{a, b, c, total - a - b - c}
		}
		a, b := split(totalA), split(totalB)
		exposure := sdk.NewCoins(sdk.NewInt64Coin("ucoin-a", totalA), sdk.NewInt64Coin("ucoin-b", totalB))
		allocation := types.TerminalAllocation{
			OriginalExposure:      exposure,
			Provider:              sdk.NewCoins(sdk.NewInt64Coin("ucoin-a", a[0]), sdk.NewInt64Coin("ucoin-b", b[0])),
			Customer:              sdk.NewCoins(sdk.NewInt64Coin("ucoin-a", a[1]), sdk.NewInt64Coin("ucoin-b", b[1])),
			Platform:              sdk.NewCoins(sdk.NewInt64Coin("ucoin-a", a[2]), sdk.NewInt64Coin("ucoin-b", b[2])),
			SlashWitness:          sdk.NewCoins(sdk.NewInt64Coin("ucoin-a", a[3]), sdk.NewInt64Coin("ucoin-b", b[3])),
			SlashWitnessRecipient: testAddress(7),
			ResolutionType:        types.FinancialResolutionPartialSplit,
		}
		require.NoError(t, ValidateTerminalAllocation(exposure, allocation), "iteration %d", i)
		allocation.Provider = allocation.Provider.Add(sdk.NewInt64Coin("ucoin-a", 1))
		require.Error(t, ValidateTerminalAllocation(exposure, allocation), "iteration %d", i)
	}
}

func TestOpenFinancialCaseMergesClaimsAndHoldsPayout(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "legacy-dispute-1", []byte("open-1"))
	opened, claim, duplicate, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Len(t, opened.Claims, 1)
	require.Equal(t, claim.ClaimId, opened.Claims[0].ClaimId)

	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateHeld, payout.State)
	require.Equal(t, opened.CaseId, payout.DisputeID)
	require.Equal(t, uint32(2), opened.ActiveHoldCount)

	exact, exactClaim, exactDuplicate, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.True(t, exactDuplicate)
	require.Equal(t, opened.CaseId, exact.CaseId)
	require.Equal(t, claim.ClaimId, exactClaim.ClaimId)
	require.Len(t, exact.Claims, 1)

	second := financialCaseOpenRequest(s, types.FinancialClaimTypeFraud, "fraud", "fraud-report-1", []byte("open-2"))
	merged, _, duplicate, err := s.keeper.OpenFinancialCase(s.ctx, second)
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, opened.CaseId, merged.CaseId)
	require.Len(t, merged.Claims, 2)
	leaseAlias := financialCaseOpenRequest(s, types.FinancialClaimTypeService, "hpc", "lease-alias", []byte("open-lease"))
	leaseAlias.Subject = types.FinancialSubject{Type: types.FinancialSubjectTypeHPCJob, PrimaryId: "job-alias", HpcJobId: "job-alias", LeaseId: "lease-84d", EscrowId: s.escrowID}
	leaseAlias.TrustedAdapter = true
	mergedByLease, _, _, err := s.keeper.OpenFinancialCase(s.ctx, leaseAlias)
	require.NoError(t, err)
	require.Equal(t, opened.CaseId, mergedByLease.CaseId)
	conflictingParty := financialCaseOpenRequest(s, types.FinancialClaimTypeHPC, "hpc", "conflicting-party", []byte("open-3"))
	conflictingParty.Respondent = testAddress(33)
	conflictingParty.TrustedAdapter = true
	_, _, _, err = s.keeper.OpenFinancialCase(s.ctx, conflictingParty)
	require.ErrorIs(t, err, types.ErrFinancialCaseAuthorization)

	conflict := request
	conflict.Claim.EvidenceHash = bytes.Repeat([]byte{9}, sha256.Size)
	_, _, _, err = s.keeper.OpenFinancialCase(s.ctx, conflict)
	require.ErrorIs(t, err, types.ErrFinancialCaseIdempotencyConflict)
}

func TestFinancialCaseAuthorizationAppealAndFinalEffectsExactlyOnce(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "billing-1", []byte("case-1")))
	require.NoError(t, err)

	require.Error(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, testAddress(9)))
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))

	allocation := types.TerminalAllocation{
		OriginalExposure: sdk.NewCoins(sdk.NewInt64Coin("uve", 100)),
		Provider:         sdk.NewCoins(sdk.NewInt64Coin("uve", 60)),
		Customer:         sdk.NewCoins(sdk.NewInt64Coin("uve", 40)),
		ResolutionType:   types.FinancialResolutionPartialSplit,
	}
	require.Error(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.provider, allocation))
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))

	appeal, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, bytes.Repeat([]byte{2}, sha256.Size), "enc/ref", []byte("appeal-1"))
	require.NoError(t, err)
	require.False(t, duplicate)
	require.NotEmpty(t, appeal.AppealId)
	exactAppeal, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, bytes.Repeat([]byte{2}, sha256.Size), "enc/ref", []byte("appeal-1"))
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, appeal.AppealId, exactAppeal.AppealId)
	_, _, err = s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, bytes.Repeat([]byte{2}, sha256.Size), "enc/changed", []byte("appeal-1"))
	require.ErrorIs(t, err, types.ErrFinancialCaseIdempotencyConflict)
	caseAfterAppeal, found := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	require.True(t, found)
	require.Equal(t, types.FinancialCaseStatusReview, caseAfterAppeal.Status)
	payout, _ := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.Equal(t, types.PayoutStateHeld, payout.State)

	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	s.ctx = s.ctx.WithBlockHeight(caseAfterAppeal.ReviewDeadlineHeight + 200000).WithBlockTime(s.ctx.BlockTime().Add(15 * 24 * time.Hour))
	finalCase, err := s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, types.FinancialCaseStatusFinal, finalCase.Status)
	firstTransfers := len(s.bank.moduleToAccountCalls)
	require.Equal(t, 2, firstTransfers)

	finalRetry, err := s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, types.FinancialCaseStatusFinal, finalRetry.Status)
	require.Equal(t, firstTransfers, len(s.bank.moduleToAccountCalls))
	for _, effect := range finalRetry.Effects {
		require.Equal(t, types.FinancialEffectStatusApplied, effect.Status)
	}
	escrow, found := s.keeper.GetEscrow(s.ctx, s.escrowID)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uve", 100)), escrow.Balance)
	require.Equal(t, types.EscrowStateActive, escrow.State)
}

func TestFinancialCaseTimeoutEscalatesAndDoesNotRelease(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeHPC, "hpc", "hpc-dispute-1", []byte("timeout-1")))
	require.NoError(t, err)

	s.ctx = s.ctx.WithBlockHeight(opened.EvidenceDeadlineHeight + 1).WithBlockTime(time.Unix(opened.EvidenceDeadlineTime+1, 0))
	processed, err := s.keeper.ProcessFinancialCaseTimeouts(s.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), processed)
	timedOut, found := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	require.True(t, found)
	require.Equal(t, types.FinancialCaseStatusEscalated, timedOut.Status)
	payout, _ := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.Equal(t, types.PayoutStateHeld, payout.State)
}

func TestFinancialCaseRejectsNonPartyAndExpiredDirectClaims(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "billing-auth", []byte("auth-open"))
	request.Claimant = testAddress(44)
	request.Claim.Claimant = request.Claimant
	_, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.ErrorIs(t, err, types.ErrFinancialCaseAuthorization)

	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "billing-deadline", []byte("deadline-open")))
	require.NoError(t, err)
	s.ctx = s.ctx.WithBlockHeight(opened.FilingDeadlineHeight + 1).WithBlockTime(time.Unix(opened.FilingDeadlineTime+1, 0))
	_, _, _, err = s.keeper.AddFinancialClaim(s.ctx, opened.CaseId, types.FinancialClaim{
		ClaimType: types.FinancialClaimTypeUsage, Claimant: s.customer, SourceModule: "settlement",
		SourceReference: "late-claim", EvidenceHash: bytes.Repeat([]byte{3}, sha256.Size), IdempotencyKey: []byte("late-claim"),
	})
	require.ErrorIs(t, err, types.ErrFinancialCaseDeadline)
}

func TestFinancialCaseRewardHoldAndAllAllocationRecipients(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	reward := types.NewClaimableRewards(s.provider, s.ctx.BlockTime())
	reward.AddReward(types.RewardEntry{DistributionID: "dist-84d", Source: types.RewardSourceUsage, Amount: sdk.NewCoins(sdk.NewInt64Coin("uve", 25)), CreatedAt: s.ctx.BlockTime(), Reason: "usage"})
	require.NoError(t, s.keeper.SetClaimableRewards(s.ctx, sdk.MustAccAddressFromBech32(s.provider), *reward))

	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeUsage, "settlement", "reward-hold", []byte("reward-hold")))
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uve", 25)), opened.Exposure.UnclaimedRewards)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uve", 125)), opened.Exposure.OriginalHeld)
	require.Equal(t, uint32(3), opened.ActiveHoldCount)
	require.ErrorIs(t, func() error {
		_, err := s.keeper.ClaimRewards(s.ctx, sdk.MustAccAddressFromBech32(s.provider), "")
		return err
	}(), types.ErrDisputeActive)

	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{
		OriginalExposure: sdk.NewCoins(sdk.NewInt64Coin("uve", 125)),
		Provider:         sdk.NewCoins(sdk.NewInt64Coin("uve", 75)), Customer: sdk.NewCoins(sdk.NewInt64Coin("uve", 30)),
		Platform: sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), SlashWitness: sdk.NewCoins(sdk.NewInt64Coin("uve", 10)),
		SlashWitnessRecipient: testAddress(7), ResolutionType: types.FinancialResolutionFraudConfirmed,
	}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	caseAfterResolve, _ := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	s.ctx = s.ctx.WithBlockHeight(caseAfterResolve.AppealDeadlineHeight + 1).WithBlockTime(time.Unix(caseAfterResolve.AppealDeadlineTime+1, 0))
	finalCase, err := s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, types.FinancialCaseStatusFinal, finalCase.Status)
	require.Len(t, s.bank.moduleToAccountCalls, 4)
	providerPaid := sdk.NewCoins()
	for _, transfer := range s.bank.moduleToAccountCalls {
		if transfer.to == s.provider {
			providerPaid = providerPaid.Add(transfer.amount...)
		}
	}
	require.Equal(t, allocation.Provider, providerPaid)
	require.Len(t, s.bank.moduleToModuleCalls, 1)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), s.bank.moduleToModuleCalls[0].amount)
	remaining, found := s.keeper.GetClaimableRewards(s.ctx, sdk.MustAccAddressFromBech32(s.provider))
	require.True(t, found)
	require.True(t, remaining.TotalClaimable.IsZero())
	require.Empty(t, remaining.RewardEntries)
	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCancelled, payout.State, "partial payout plus rewards is not a completed gross payout")
}

func TestFinancialCaseProviderPayoutCompletesWhenAllocationAlsoContainsRewards(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	reward := types.NewClaimableRewards(s.provider, s.ctx.BlockTime())
	reward.AddReward(types.RewardEntry{DistributionID: "dist-payout-reward", Source: types.RewardSourceUsage, Amount: sdk.NewCoins(sdk.NewInt64Coin("uve", 25)), CreatedAt: s.ctx.BlockTime(), Reason: "usage"})
	require.NoError(t, s.keeper.SetClaimableRewards(s.ctx, sdk.MustAccAddressFromBech32(s.provider), *reward))
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeUsage, "settlement", "payout-reward-classification", []byte("payout-reward-classification")))
	require.NoError(t, err)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{OriginalExposure: opened.Exposure.OriginalHeld, Provider: opened.Exposure.OriginalHeld, ResolutionType: types.FinancialResolutionProviderWin}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	resolved, found := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	require.True(t, found)
	s.ctx = s.ctx.WithBlockHeight(resolved.AppealDeadlineHeight + 1).WithBlockTime(time.Unix(resolved.AppealDeadlineTime+1, 0))
	_, err = s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
}

func TestCanonicalPayoutCannotUseLegacyReleaseOrRefundAfterActivation(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "fence", []byte("fence")))
	require.NoError(t, err)
	require.ErrorIs(t, s.keeper.ReleasePayoutHold(s.ctx, s.payoutID), types.ErrLegacyFinancialMutationFenced)
	require.ErrorIs(t, s.keeper.RefundPayout(s.ctx, s.payoutID, "legacy"), types.ErrLegacyFinancialMutationFenced)
	require.NotEmpty(t, opened.CaseId)
}

func TestProviderFiledCaseUsesCanonicalProviderCustomerAllocationRoles(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeService, "settlement", "provider-filed", []byte("provider-filed"))
	request.Claimant, request.Respondent = s.provider, s.customer
	request.Claim.Claimant = s.provider
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.Equal(t, s.provider, opened.Provider)
	require.Equal(t, s.customer, opened.Customer)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.provider))
	allocation := types.TerminalAllocation{
		OriginalExposure: sdk.NewCoins(sdk.NewInt64Coin("uve", 100)),
		Provider:         sdk.NewCoins(sdk.NewInt64Coin("uve", 70)), Customer: sdk.NewCoins(sdk.NewInt64Coin("uve", 30)),
		ResolutionType: types.FinancialResolutionPartialSplit,
	}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	resolved, _ := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	s.ctx = s.ctx.WithBlockHeight(resolved.AppealDeadlineHeight + 1).WithBlockTime(time.Unix(resolved.AppealDeadlineTime+1, 0))
	_, err = s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, s.provider, s.bank.moduleToAccountCalls[0].to)
	require.Equal(t, s.customer, s.bank.moduleToAccountCalls[1].to)
}

func TestMigrateFinancialCasesRebindsLegacyHeldPayoutIdempotently(t *testing.T) {
	s := setupFinancialCaseTest(t)
	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.NoError(t, payout.Hold("legacy-dispute-7", "legacy hold", s.ctx.BlockTime()))
	require.NoError(t, s.keeper.SetPayout(s.ctx, payout))

	report, err := s.keeper.MigrateFinancialCases(s.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), report.PayoutsScanned)
	require.Equal(t, uint64(1), report.CasesCreated)

	migratedPayout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateHeld, migratedPayout.State)
	require.NotEqual(t, "legacy-dispute-7", migratedPayout.DisputeID)
	financialCase, found := s.keeper.GetFinancialCase(s.ctx, migratedPayout.DisputeID)
	require.True(t, found)
	require.True(t, financialCase.Migrated)
	require.Equal(t, payout.GrossAmount, financialCase.Exposure.OriginalHeld)

	retry, err := s.keeper.MigrateFinancialCases(s.ctx)
	require.NoError(t, err)
	require.Equal(t, report, retry)
}

func TestFinancialCaseActivationFenceAndExecutionGuard(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.ctx.KVStore(s.keeper.skey).Delete(types.FinancialCaseActivationKey())
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "pre-activation", []byte("pre-activation"))
	_, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.ErrorIs(t, err, types.ErrFinancialCasesNotActive)

	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	require.Equal(t, opened.CaseId, payout.DisputeID)
	require.ErrorIs(t, s.keeper.executePayoutTransfer(s.ctx, &payout), types.ErrPayoutHeld)
	require.Empty(t, s.bank.moduleToAccountCalls)
}

func TestRebuildFinancialCaseStateRestoresIndexesAndReplay(t *testing.T) {
	s := setupFinancialCaseTest(t)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "rebuild", []byte("rebuild"))
	financialCase, _, _, err := s.keeper.openFinancialCase(s.ctx, request)
	require.NoError(t, err)
	store := s.ctx.KVStore(s.keeper.skey)
	store.Delete(types.FinancialSubjectKey("1/order-84d"))
	store.Delete(types.FinancialClaimIdempotencyKey(request.IdempotencyKey))

	require.NoError(t, s.keeper.RebuildFinancialCaseState(s.ctx))
	require.Equal(t, financialCase.CaseId, string(store.Get(types.FinancialSubjectKey("1/order-84d"))))
	_, exists, err := s.keeper.getFinancialClaimReplay(s.ctx, request.IdempotencyKey)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestRebuildFinancialCaseStateRestoresAppealReplay(t *testing.T) {
	s := setupFinancialCaseTest(t)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "appeal-rebuild", []byte("appeal-rebuild")))
	require.NoError(t, err)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{OriginalExposure: opened.Exposure.OriginalHeld, Provider: opened.Exposure.OriginalHeld, ResolutionType: types.FinancialResolutionProviderWin}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	evidence := bytes.Repeat([]byte{6}, sha256.Size)
	appealKey := []byte("appeal-rebuild-key")
	appeal, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, evidence, "enc/rebuild-appeal", appealKey)
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, appealKey, appeal.IdempotencyKey)

	store := s.ctx.KVStore(s.keeper.skey)
	store.Delete(types.FinancialAppealIdempotencyKey(appealKey))
	_, exists, err := s.keeper.getFinancialAppealReplay(s.ctx, appealKey)
	require.NoError(t, err)
	require.False(t, exists)

	require.NoError(t, s.keeper.RebuildFinancialCaseState(s.ctx))
	replay, exists, err := s.keeper.getFinancialAppealReplay(s.ctx, appealKey)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, opened.CaseId, replay.CaseID)
	require.Equal(t, appeal.AppealId, replay.AppealID)
	retried, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, evidence, "enc/rebuild-appeal", appealKey)
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, appeal.AppealId, retried.AppealId)
	require.Empty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestFinancialCaseCompletesCanonicalPayoutAliasesBeforeIndexing(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "canonical-aliases", []byte("canonical-aliases"))
	request.Subject.OrderId = ""
	request.Subject.EscrowId = ""
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.Equal(t, "order-84d", opened.Subject.OrderId)
	require.Equal(t, s.escrowID, opened.Subject.EscrowId)
	require.Equal(t, "lease-84d", opened.Subject.LeaseId)
	require.Equal(t, uint32(2), opened.ActiveHoldCount)
	require.Empty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestFinancialCaseInvariantRejectsMissingRequiredTerminalEffect(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "missing-effects", []byte("missing-effects")))
	require.NoError(t, err)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{OriginalExposure: opened.Exposure.OriginalHeld, Provider: opened.Exposure.OriginalHeld, ResolutionType: types.FinancialResolutionProviderWin}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	resolved, _ := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	s.ctx = s.ctx.WithBlockHeight(resolved.AppealDeadlineHeight + 1).WithBlockTime(time.Unix(resolved.AppealDeadlineTime+1, 0))
	finalized, err := s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	finalized.Effects = finalized.Effects[1:]
	require.NoError(t, s.keeper.SetFinancialCase(s.ctx, *finalized))
	require.NotEmpty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestFinancialCaseFinalizationFailureRollsBackAndRetriesOnce(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "failure-retry", []byte("failure-retry")))
	require.NoError(t, err)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{OriginalExposure: opened.Exposure.OriginalHeld, Provider: opened.Exposure.OriginalHeld, ResolutionType: types.FinancialResolutionProviderWin}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	resolved, _ := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	s.ctx = s.ctx.WithBlockHeight(resolved.AppealDeadlineHeight + 1).WithBlockTime(time.Unix(resolved.AppealDeadlineTime+1, 0))

	s.bank.failAccountAt = 1
	_, err = s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.Error(t, err)
	stored, found := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	require.True(t, found)
	require.Equal(t, types.FinancialCaseStatusResolvedPendingAppeal, stored.Status)
	require.Empty(t, stored.Effects)
	require.Zero(t, s.bank.successfulAccountTransfers)

	s.bank.failAccountAt = 0
	_, err = s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, 1, s.bank.successfulAccountTransfers)
	_, err = s.keeper.FinalizeFinancialCase(s.ctx, opened.CaseId, s.authority)
	require.NoError(t, err)
	require.Equal(t, 1, s.bank.successfulAccountTransfers)
}

func TestFinancialCaseExactClaimRetrySucceedsAtClaimLimit(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	params := s.keeper.GetParams(s.ctx)
	params.FinancialCaseMaxClaims = 1
	require.NoError(t, s.keeper.SetParams(s.ctx, params))
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "claim-limit-retry", []byte("claim-limit-retry"))
	opened, first, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	retried, claim, duplicate, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, opened.CaseId, retried.CaseId)
	require.Equal(t, first.ClaimId, claim.ClaimId)
	require.Len(t, retried.Claims, 1)
}

func TestFinancialCaseAppealRetryAndConflictAtLimit(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "appeal-retry", []byte("appeal-retry")))
	require.NoError(t, err)
	require.NoError(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer))
	allocation := types.TerminalAllocation{OriginalExposure: opened.Exposure.OriginalHeld, Provider: opened.Exposure.OriginalHeld, ResolutionType: types.FinancialResolutionProviderWin}
	require.NoError(t, s.keeper.ResolveFinancialCase(s.ctx, opened.CaseId, s.authority, allocation))
	evidence := bytes.Repeat([]byte{5}, sha256.Size)
	appeal, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, evidence, "enc/appeal", []byte("appeal-key"))
	require.NoError(t, err)
	require.False(t, duplicate)

	retried, duplicate, err := s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, evidence, "enc/appeal", []byte("appeal-key"))
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, appeal.AppealId, retried.AppealId)

	_, _, err = s.keeper.AppealFinancialCase(s.ctx, opened.CaseId, s.customer, evidence, "enc/different", []byte("appeal-key"))
	require.ErrorIs(t, err, types.ErrFinancialCaseIdempotencyConflict)
}

func TestFinancialCaseCannotReopenTerminalDeterministicRoot(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "terminal-root", []byte("terminal-root"))
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.NoError(t, s.keeper.CancelFinancialCase(s.ctx, opened.CaseId, s.customer, bytes.Repeat([]byte{9}, sha256.Size)))
	request.IdempotencyKey = []byte("terminal-root-new-key")
	request.Claim.IdempotencyKey = request.IdempotencyKey
	_, _, _, err = s.keeper.OpenFinancialCase(s.ctx, request)
	require.ErrorIs(t, err, types.ErrFinancialCaseTransition)
	stored, found := s.keeper.GetFinancialCase(s.ctx, opened.CaseId)
	require.True(t, found)
	require.Equal(t, types.FinancialCaseStatusCancelled, stored.Status)
}

func TestFinancialCaseReviewAndEscalationDeadlines(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "review-deadline", []byte("review-deadline")))
	require.NoError(t, err)
	s.ctx = s.ctx.WithBlockHeight(opened.EvidenceDeadlineHeight + 1).WithBlockTime(time.Unix(opened.EvidenceDeadlineTime+1, 0))
	require.ErrorIs(t, s.keeper.SubmitFinancialCaseForReview(s.ctx, opened.CaseId, s.customer), types.ErrFinancialCaseDeadline)

	s = setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err = s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "escalation-deadline", []byte("escalation-deadline")))
	require.NoError(t, err)
	s.ctx = s.ctx.WithBlockHeight(opened.EscalationDeadlineHeight + 1).WithBlockTime(time.Unix(opened.EscalationDeadlineTime+1, 0))
	require.ErrorIs(t, s.keeper.EscalateFinancialCase(s.ctx, opened.CaseId, s.customer, bytes.Repeat([]byte{8}, sha256.Size)), types.ErrFinancialCaseDeadline)
}

type financialCaseFixture struct {
	keeper    Keeper
	ctx       sdk.Context
	bank      *financialCaseBankKeeper
	authority string
	customer  string
	provider  string
	escrowID  string
	payoutID  string
}

func setupFinancialCaseTest(t *testing.T) financialCaseFixture {
	t.Helper()
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 10, Time: time.Unix(1_700_000_000, 0).UTC()}, false, log.NewNopLogger())
	registry := codectypes.NewInterfaceRegistry()
	c := codec.NewProtoCodec(registry)
	bank := &financialCaseBankKeeper{}
	k := NewKeeper(c, storeKey, bank, nil, "", nil)
	authority := testAddress(90)
	k.authority = authority
	k.ActivateFinancialCases(ctx)
	customer := testAddress(1)
	provider := testAddress(2)
	escrow := types.NewEscrowAccount("escrow-84d", "order-84d", customer, sdk.NewCoins(sdk.NewInt64Coin("uve", 100)), ctx.BlockTime().Add(7*24*time.Hour), nil, ctx.BlockTime(), ctx.BlockHeight())
	require.NoError(t, escrow.Activate(provider, ctx.BlockTime()))
	require.NoError(t, k.SetEscrow(ctx, *escrow))
	payout := types.NewPayoutRecord("payout-84d", "invoice-84d", "settlement-84d", escrow.EscrowID, escrow.OrderID, "lease-84d", provider, customer, sdk.NewCoins(sdk.NewInt64Coin("uve", 100)), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), ctx.BlockTime(), ctx.BlockHeight())
	require.NoError(t, k.SetPayout(ctx, *payout))
	return financialCaseFixture{keeper: k, ctx: ctx, bank: bank, authority: authority, customer: customer, provider: provider, escrowID: escrow.EscrowID, payoutID: payout.PayoutID}
}

type financialCaseTransfer struct {
	module string
	to     string
	amount sdk.Coins
}

type financialCaseModuleTransfer struct {
	from   string
	to     string
	amount sdk.Coins
}

type financialCaseBankKeeper struct {
	moduleToAccountCalls       []financialCaseTransfer
	moduleToModuleCalls        []financialCaseModuleTransfer
	failAccountAt              int
	accountAttempts            int
	successfulAccountTransfers int
}

func (m *financialCaseBankKeeper) SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
	return nil
}

func (m *financialCaseBankKeeper) SendCoinsFromModuleToModule(_ context.Context, from, to string, amount sdk.Coins) error {
	m.moduleToModuleCalls = append(m.moduleToModuleCalls, financialCaseModuleTransfer{from: from, to: to, amount: amount})
	return nil
}
func (m *financialCaseBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, module string, to sdk.AccAddress, amount sdk.Coins) error {
	m.accountAttempts++
	if m.failAccountAt > 0 && m.accountAttempts == m.failAccountAt {
		return fmt.Errorf("injected account transfer failure")
	}
	m.moduleToAccountCalls = append(m.moduleToAccountCalls, financialCaseTransfer{module: module, to: to.String(), amount: amount})
	m.successfulAccountTransfers++
	return nil
}
func (m *financialCaseBankKeeper) SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}
func (m *financialCaseBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}
func (m *financialCaseBankKeeper) GetBalance(context.Context, sdk.AccAddress, string) sdk.Coin {
	return sdk.Coin{}
}

func financialCaseOpenRequest(s financialCaseFixture, claimType types.FinancialClaimType, source, reference string, idempotency []byte) FinancialCaseOpenRequest {
	return FinancialCaseOpenRequest{
		Subject:        types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: "order-84d", OrderId: "order-84d", InvoiceId: "invoice-84d", SettlementId: "settlement-84d", EscrowId: s.escrowID},
		Claimant:       s.customer,
		Respondent:     s.provider,
		Claim:          types.FinancialClaim{ClaimType: claimType, Claimant: s.customer, SourceModule: source, SourceReference: reference, EvidenceHash: bytes.Repeat([]byte{1}, sha256.Size), EncryptedReference: "enc/ref", IdempotencyKey: idempotency},
		IdempotencyKey: idempotency,
	}
}

func testAddress(seed byte) string {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20)).String()
}
