package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestAuditPayoutLineageCompletesEscrowAliasAndHold(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	payout, found := s.keeper.GetPayout(s.ctx, s.payoutID)
	require.True(t, found)
	payout.EscrowID = ""
	require.NoError(t, s.keeper.SetPayout(s.ctx, payout))

	request := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "payout-only", []byte("payout-only"))
	request.Subject.EscrowId = ""
	request.Subject.OrderId = ""
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, request)
	require.NoError(t, err)
	require.Equal(t, s.escrowID, opened.Subject.EscrowId)
	require.Equal(t, uint32(2), opened.ActiveHoldCount)
	require.Empty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestAuditTerminalEffectsRequireExpectedMarkerSet(t *testing.T) {
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
	finalized.Effects = nil
	require.NoError(t, s.keeper.SetFinancialCase(s.ctx, *finalized))
	require.NotEmpty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestAuditOpenRejectsConflictingAliasStateBeforeMutation(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	first := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "first", []byte("first"))
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, first)
	require.NoError(t, err)

	corrupted := *opened
	corrupted.Subject = types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: "second-root", OrderId: "second-root", EscrowId: s.escrowID}
	secondID, err := DeterministicFinancialCaseID(corrupted.Subject)
	require.NoError(t, err)
	corrupted.CaseId = secondID
	for i := range corrupted.Claims {
		corrupted.Claims[i].ClaimId, corrupted.Claims[i].PayloadHash, err = DeterministicFinancialClaimID(secondID, corrupted.Claims[i])
		require.NoError(t, err)
	}
	corrupted.ClaimRoot = financialClaimRoot(corrupted.Claims)
	// Inject a malformed competing active record/index directly to model imported/corrupt state.
	bz, err := s.keeper.cdc.Marshal(&corrupted)
	require.NoError(t, err)
	store := s.ctx.KVStore(s.keeper.skey)
	store.Set(types.FinancialCaseKey(secondID), bz)
	store.Set(types.FinancialCaseIndexKey(types.PrefixFinancialCaseByEscrow, s.escrowID, secondID), []byte(secondID))

	third := financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "third", []byte("third"))
	third.Subject = types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: "third-root", OrderId: "third-root", EscrowId: s.escrowID}
	_, _, _, err = s.keeper.OpenFinancialCase(s.ctx, third)
	require.ErrorIs(t, err, types.ErrFinancialCaseMalformedState)
}

func TestAuditInvariantRejectsOrphanSubjectIndex(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "orphan-subject-index", []byte("orphan-subject-index")))
	require.NoError(t, err)
	store := s.ctx.KVStore(s.keeper.skey)
	store.Delete(types.FinancialCaseKey(opened.CaseId))

	require.NotEmpty(t, s.keeper.ValidateFinancialCaseInvariants(s.ctx))
}

func TestAuditFinalizeFailureRollsBackBankAndEffectState(t *testing.T) {
	s := setupFinancialCaseTest(t)
	s.keeper.ActivateFinancialCases(s.ctx)
	opened, _, _, err := s.keeper.OpenFinancialCase(s.ctx, financialCaseOpenRequest(s, types.FinancialClaimTypeBilling, "settlement", "failure", []byte("failure")))
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
}
