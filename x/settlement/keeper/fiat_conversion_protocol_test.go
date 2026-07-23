package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"math"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	testFiatCancellationCode       = "USER_CANCELLED"
	terminalPolicyCompletedForTest = "external_fiat_completed_custody_sink"
	testPayoutCompletedStatus      = "completed"
	testPayoutFiatAmount           = "123.45"
)

func setupAuthenticatedFiatConversion(t *testing.T) (*KeeperTestSuite, types.FiatConversionRecord, types.PayoutRecord, *veidtypes.ComplianceRecord) {
	t.Helper()
	s := new(KeeperTestSuite)
	s.SetT(t)
	s.SetupTest()

	params := s.keeper.GetParams(s.ctx)
	configureCertifiedFiatProfiles(&params)
	params.FiatConversionMinAmount = "1"
	params.FiatConversionMaxAmount = "1000000"
	params.FiatConversionDailyLimit = "2000000"
	params.FiatConversionStableDenom = testStableDenom
	params.FiatConversionStableSymbol = testStableSymbol
	params.FiatConversionStableDecimals = 6
	params.FiatConversionMinComplianceStatus = testComplianceCleared
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	compliance := veidtypes.NewComplianceRecord(s.provider.String(), s.ctx.BlockTime())
	compliance.Status = veidtypes.ComplianceStatusCleared
	compliance.RiskScore = 5
	compliance.ExpiresAt = s.ctx.BlockTime().Add(time.Hour).Unix()
	s.keeper.SetComplianceKeeper(mockComplianceKeeper{record: compliance})

	settlement := s.buildSettlement(t, "task85b-observation")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))
	payout := types.NewPayoutRecord(
		"payout-task85b", "invoice-task85b", settlement.SettlementID, settlement.EscrowID,
		settlement.OrderID, settlement.LeaseID, settlement.Provider, settlement.Customer,
		settlement.TotalAmount, settlement.PlatformFee, settlement.ValidatorFee, sdk.NewCoins(),
		s.ctx.BlockTime(), s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
	s.bankKeeper.SetModuleBalance(types.ModuleAccountName, payout.NetAmount)
	request := types.FiatConversionRequest{
		InvoiceID: "invoice-task85b", SettlementID: settlement.SettlementID, PayoutID: payout.PayoutID,
		Provider: settlement.Provider, Customer: settlement.Customer, RequestedBy: settlement.Provider,
		CryptoAmount: payout.NetAmount[0], FiatCurrency: "USD", PaymentMethod: "bank_transfer",
		DestinationHash: types.HashDestination("opaque-destination"), SlippageTolerance: 0.01,
		CryptoToken:      types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:      types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	conversion, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)
	storedConversion, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	return s, storedConversion, *payout, compliance
}

func observationFor(t *testing.T, conversion types.FiatConversionRecord, compliance *veidtypes.ComplianceRecord, sequence uint64, stage settlementv1.FiatConversionObservationStage) *types.MsgRecordFiatConversionObservation {
	t.Helper()
	complianceHash, err := testComplianceDigest(compliance)
	require.NoError(t, err)
	return &types.MsgRecordFiatConversionObservation{
		Sender: conversion.Provider, ConversionId: conversion.ConversionID, ObservationSequence: sequence,
		IdempotencyKey: bytes.Repeat([]byte{byte(sequence)}, 32), Stage: stage,
		DexProfileId: conversion.DEXProfileID, DexProfileDigest: append([]byte(nil), conversion.DEXProfileDigest...),
		PayoutProfileId: conversion.PayoutProfileID, PayoutProfileDigest: append([]byte(nil), conversion.PayoutProfileDigest...),
		ObservedAt: conversion.RequestedAt.Unix(), EvidenceHash: bytes.Repeat([]byte{byte(sequence + 20)}, 32),
		ComplianceDecisionHash: complianceHash, Status: "accepted",
	}
}

func testComplianceDigest(record *veidtypes.ComplianceRecord) ([]byte, error) {
	bz, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(bz)
	return digest[:], nil
}

func runSuccessfulFiatObservationFlow(t *testing.T, s *KeeperTestSuite, conversion types.FiatConversionRecord, compliance *veidtypes.ComplianceRecord) types.FiatConversionRecord {
	t.Helper()
	completed := runSuccessfulFiatObservationFlowUntilCompletion(t, s, conversion, compliance)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, completed)
	require.NoError(t, err)
	result, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	return result
}

func runSuccessfulFiatObservationFlowUntilCompletion(t *testing.T, s *KeeperTestSuite, conversion types.FiatConversionRecord, compliance *veidtypes.ComplianceRecord) *types.MsgRecordFiatConversionObservation {
	t.Helper()
	quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	quote.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	quote.QuoteExpiry = s.ctx.BlockTime().Add(10 * time.Minute).Unix()
	quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, quote)
	require.NoError(t, err)

	swapSubmitted := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	swapSubmitted.QuoteDigest = quote.QuoteDigest
	swapSubmitted.QuoteExpiry = quote.QuoteExpiry
	swapSubmitted.MinimumStableOutput = quote.MinimumStableOutput
	swapSubmitted.SwapTxHash = "DEX-TX-1"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, swapSubmitted)
	require.NoError(t, err)

	swapFinalized := observationFor(t, conversion, compliance, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED)
	swapFinalized.QuoteDigest = quote.QuoteDigest
	swapFinalized.MinimumStableOutput = quote.MinimumStableOutput
	swapFinalized.SwapTxHash = swapSubmitted.SwapTxHash
	swapFinalized.SwapHeight = 100
	swapFinalized.SwapBlockHash = bytes.Repeat([]byte{41}, 32)
	swapFinalized.SwapFinalityConfirmations = 2
	swapFinalized.SwapFinalityHash = bytes.Repeat([]byte{42}, 32)
	swapFinalized.StableAmount = sdk.NewInt64Coin(testStableDenom, 950)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, swapFinalized)
	require.NoError(t, err)

	payoutQuoted := observationFor(t, conversion, compliance, 4, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED)
	payoutQuoted.QuoteDigest = bytes.Repeat([]byte{51}, 32)
	payoutQuoted.QuoteExpiry = s.ctx.BlockTime().Add(10 * time.Minute).Unix()
	payoutQuoted.OffRampQuoteId = "OFFRAMP-QUOTE-1"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, payoutQuoted)
	require.NoError(t, err)

	payoutSubmitted := observationFor(t, conversion, compliance, 5, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED)
	payoutSubmitted.QuoteDigest = payoutQuoted.QuoteDigest
	payoutSubmitted.QuoteExpiry = payoutQuoted.QuoteExpiry
	payoutSubmitted.OffRampQuoteId = payoutQuoted.OffRampQuoteId
	payoutSubmitted.OffRampPayoutId = "OFFRAMP-PAYOUT-1"
	payoutSubmitted.Status = "submitted"
	payoutSubmitted.PrivacySafeReferenceHash = bytes.Repeat([]byte{61}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, payoutSubmitted)
	require.NoError(t, err)

	completed := observationFor(t, conversion, compliance, 6, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED)
	completed.OffRampQuoteId = payoutQuoted.OffRampQuoteId
	completed.OffRampPayoutId = payoutSubmitted.OffRampPayoutId
	completed.QuoteDigest = payoutQuoted.QuoteDigest
	completed.Status = testPayoutCompletedStatus
	completed.FiatAmount = testPayoutFiatAmount
	completed.PrivacySafeReferenceHash = bytes.Repeat([]byte{61}, 32)
	completed.PayoutFinalityHash = bytes.Repeat([]byte{62}, 32)
	return completed
}

func TestFiatObservationExactReplayAndConflict(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	msg := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	msg.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	msg.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	msg.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	first, err := s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.NoError(t, err)
	require.False(t, first.ExactDuplicate)
	duplicate, err := s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.NoError(t, err)
	require.True(t, duplicate.ExactDuplicate)
	msg.Status = "changed"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.ErrorIs(t, err, types.ErrFiatObservationReplayConflict)
}

func TestFiatObservationAuthorizationSequenceProfileAndCompliance(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	base := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	base.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	base.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	base.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	unauthorized := *base
	unauthorized.Sender = s.depositor.String()
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, &unauthorized)
	require.ErrorIs(t, err, types.ErrUnauthorized)
	gap := *base
	gap.ObservationSequence = 2
	gap.IdempotencyKey = bytes.Repeat([]byte{2}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &gap)
	require.ErrorIs(t, err, types.ErrFiatObservationSequence)
	wrongProfile := *base
	wrongProfile.DexProfileDigest = bytes.Repeat([]byte{99}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &wrongProfile)
	require.ErrorIs(t, err, types.ErrFiatProfileCommitment)
	wrongCompliance := *base
	wrongCompliance.ComplianceDecisionHash = bytes.Repeat([]byte{98}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &wrongCompliance)
	require.ErrorIs(t, err, types.ErrComplianceRequired)
}

func TestFiatObservationEveryLegalStageAndTerminalExactlyOnce(t *testing.T) {
	s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
	providerBefore := s.bankKeeper.GetBalance(s.ctx, s.provider, "uve").Amount
	final := runSuccessfulFiatObservationFlow(t, s, conversion, compliance)
	require.Equal(t, types.FiatConversionStatePayoutCompleted, final.State)
	require.Equal(t, uint64(6), final.ObservationSequence)
	require.True(t, final.ValueMovementApplied)
	require.Equal(t, terminalPolicyCompletedForTest, final.TerminalPolicy)
	require.True(t, final.CustodySinkAmount.IsEqual(final.CryptoAmount))
	require.Len(t, final.CustodySinkEffectHash, sha256.Size)
	updatedPayout, found := s.keeper.GetPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, updatedPayout.State)
	require.Empty(t, updatedPayout.TxHash)
	require.Equal(t, final.PayoutFinalityHash, updatedPayout.ExternalFinalityHash)
	require.Equal(t, providerBefore, s.bankKeeper.GetBalance(s.ctx, s.provider, "uve").Amount)
	require.True(t, s.bankKeeper.ModuleBalance(types.ModuleAccountName).IsZero())
	require.True(t, s.bankKeeper.ModuleBalance(types.FiatConversionCustodyAccountName).Equal(sdk.NewCoins(final.CryptoAmount)))
	entries := s.keeper.GetPayoutLedgerEntries(s.ctx, payout.PayoutID)
	require.Equal(t, payout.NetAmount, entries[len(entries)-1].Amount)

	terminalReplay := observationFor(t, conversion, compliance, 7, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED)
	terminalReplay.OffRampQuoteId = final.OffRampQuoteID
	terminalReplay.OffRampPayoutId = final.OffRampID
	terminalReplay.QuoteDigest = final.QuoteDigest
	terminalReplay.Status = testPayoutCompletedStatus
	terminalReplay.FiatAmount = testPayoutFiatAmount
	terminalReplay.PrivacySafeReferenceHash = bytes.Repeat([]byte{61}, 32)
	terminalReplay.PayoutFinalityHash = bytes.Repeat([]byte{62}, 32)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, terminalReplay)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)
	require.Equal(t, providerBefore, s.bankKeeper.GetBalance(s.ctx, s.provider, "uve").Amount)
}

func TestFiatObservationPayoutReferenceLineageImmutable(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)

	quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	quote.QuoteDigest = bytes.Repeat([]byte{31}, sha256.Size)
	quote.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, quote)
	require.NoError(t, err)

	swapSubmitted := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	swapSubmitted.QuoteDigest, swapSubmitted.QuoteExpiry, swapSubmitted.MinimumStableOutput = quote.QuoteDigest, quote.QuoteExpiry, quote.MinimumStableOutput
	swapSubmitted.SwapTxHash = "DEX-REFERENCE-LINEAGE"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, swapSubmitted)
	require.NoError(t, err)

	swapFinalized := observationFor(t, conversion, compliance, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED)
	swapFinalized.QuoteDigest, swapFinalized.MinimumStableOutput, swapFinalized.SwapTxHash = quote.QuoteDigest, quote.MinimumStableOutput, swapSubmitted.SwapTxHash
	swapFinalized.SwapHeight = 10
	swapFinalized.SwapBlockHash = bytes.Repeat([]byte{41}, sha256.Size)
	swapFinalized.SwapFinalityConfirmations = 2
	swapFinalized.SwapFinalityHash = bytes.Repeat([]byte{42}, sha256.Size)
	swapFinalized.StableAmount = sdk.NewInt64Coin(testStableDenom, 950)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, swapFinalized)
	require.NoError(t, err)

	payoutQuoted := observationFor(t, conversion, compliance, 4, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED)
	payoutQuoted.QuoteDigest = bytes.Repeat([]byte{51}, sha256.Size)
	payoutQuoted.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	payoutQuoted.OffRampQuoteId = "OFFRAMP-REFERENCE-QUOTE"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, payoutQuoted)
	require.NoError(t, err)

	submitted := observationFor(t, conversion, compliance, 5, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED)
	submitted.QuoteDigest, submitted.QuoteExpiry = payoutQuoted.QuoteDigest, payoutQuoted.QuoteExpiry
	submitted.OffRampQuoteId, submitted.OffRampPayoutId, submitted.Status = payoutQuoted.OffRampQuoteId, "OFFRAMP-REFERENCE-PAYOUT", "submitted"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, submitted)
	require.Error(t, err)

	referenceHash := bytes.Repeat([]byte{61}, sha256.Size)
	submitted.PrivacySafeReferenceHash = referenceHash
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, submitted)
	require.NoError(t, err)

	completed := observationFor(t, conversion, compliance, 6, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED)
	completed.OffRampQuoteId, completed.OffRampPayoutId, completed.QuoteDigest = submitted.OffRampQuoteId, submitted.OffRampPayoutId, submitted.QuoteDigest
	completed.Status, completed.FiatAmount = testPayoutCompletedStatus, testPayoutFiatAmount
	completed.PrivacySafeReferenceHash = bytes.Repeat([]byte{99}, sha256.Size)
	completed.PayoutFinalityHash = bytes.Repeat([]byte{62}, sha256.Size)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, completed)
	require.ErrorIs(t, err, types.ErrFiatObservationEvidence)
	stored, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.Equal(t, types.FiatConversionStatePayoutSubmitted, stored.State)
	require.Equal(t, referenceHash, stored.PrivacySafeReferenceHash)
}

func TestFiatObservationCompletionInsufficientBalanceIsAtomic(t *testing.T) {
	s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
	s.bankKeeper.SetModuleBalance(types.ModuleAccountName, sdk.NewCoins())

	finalBefore := runSuccessfulFiatObservationFlowUntilCompletion(t, s, conversion, compliance)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, finalBefore)
	require.ErrorIs(t, err, types.ErrPayoutExecutionFailed)

	stored, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.Equal(t, types.FiatConversionStatePayoutSubmitted, stored.State)
	require.Equal(t, uint64(5), stored.ObservationSequence)
	linked, _ := s.keeper.GetPayout(s.ctx, payout.PayoutID)
	require.Equal(t, types.PayoutStatePending, linked.State)
	require.False(t, linked.ValueMovementApplied)
	require.Empty(t, s.bankKeeper.ModuleBalance(types.FiatConversionCustodyAccountName))
	require.Empty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))
}

func TestFiatObservationRejectsExpiryWrongDenomMinOutputAndFinality(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*KeeperTestSuite, *types.MsgRecordFiatConversionObservation)
	}{
		{"expired_quote", func(s *KeeperTestSuite, msg *types.MsgRecordFiatConversionObservation) {
			msg.QuoteExpiry = s.ctx.BlockTime().Add(-time.Second).Unix()
		}},
		{"wrong_denom", func(_ *KeeperTestSuite, msg *types.MsgRecordFiatConversionObservation) {
			msg.MinimumStableOutput = sdk.NewInt64Coin("uwrong", 900)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
			msg := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
			msg.QuoteDigest = bytes.Repeat([]byte{31}, 32)
			msg.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
			msg.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
			test.mutate(s, msg)
			_, err := s.keeper.RecordFiatConversionObservation(s.ctx, msg)
			require.Error(t, err)
			stored, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
			require.Zero(t, stored.ObservationSequence)
		})
	}
}

func TestFiatObservationExpiredQuoteCanBeReplacedWithoutReleasingHold(t *testing.T) {
	s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
	first := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	first.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	first.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	first.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, first)
	require.NoError(t, err)

	before, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.True(t, before.DailyQuotaReserved)
	requestDigest := append([]byte(nil), before.RequestDigest...)
	complianceDigest := append([]byte(nil), before.ComplianceDecisionHash...)
	dexDigest := append([]byte(nil), before.DEXProfileDigest...)
	payoutDigest := append([]byte(nil), before.PayoutProfileDigest...)

	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Unix(first.QuoteExpiry+1, 0).UTC())
	replacement := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	replacement.ObservedAt = s.ctx.BlockTime().Unix()
	replacement.QuoteDigest = bytes.Repeat([]byte{32}, 32)
	replacement.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	replacement.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 910)
	wrongProfile := *replacement
	wrongProfile.DexProfileDigest = bytes.Repeat([]byte{91}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &wrongProfile)
	require.ErrorIs(t, err, types.ErrFiatProfileCommitment)
	wrongCompliance := *replacement
	wrongCompliance.ComplianceDecisionHash = bytes.Repeat([]byte{92}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &wrongCompliance)
	require.ErrorIs(t, err, types.ErrComplianceRequired)
	result, err := s.keeper.RecordFiatConversionObservation(s.ctx, replacement)
	require.NoError(t, err)
	require.False(t, result.ExactDuplicate)
	require.Equal(t, types.FiatConversionStateSwapPending, result.Conversion.State)
	require.Equal(t, replacement.QuoteDigest, result.Conversion.QuoteDigest)
	require.Equal(t, replacement.QuoteExpiry, result.Conversion.QuoteExpiry)
	require.True(t, replacement.MinimumStableOutput.IsEqual(result.Conversion.MinimumStableOutput))
	require.Equal(t, uint64(2), result.Conversion.ObservationSequence)
	require.Len(t, result.Conversion.Observations, 2)
	require.Equal(t, requestDigest, result.Conversion.RequestDigest)
	require.Equal(t, complianceDigest, result.Conversion.ComplianceDecisionHash)
	require.Equal(t, dexDigest, result.Conversion.DEXProfileDigest)
	require.Equal(t, payoutDigest, result.Conversion.PayoutProfileDigest)
	require.True(t, result.Conversion.DailyQuotaReserved)
	linked, found := s.keeper.GetPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, linked.State)
}

func TestFiatObservationReplacementQuoteRejectedBeforeExpiry(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	first := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	first.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	first.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	first.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, first)
	require.NoError(t, err)
	replacement := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	replacement.QuoteDigest = bytes.Repeat([]byte{32}, 32)
	replacement.QuoteExpiry = s.ctx.BlockTime().Add(2 * time.Minute).Unix()
	replacement.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 910)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, replacement)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)
	stored, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, uint64(1), stored.ObservationSequence)
	require.Equal(t, first.QuoteDigest, stored.QuoteDigest)
}

func TestFiatObservationReplacementQuoteRejectedAfterSwapSubmission(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	first := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	first.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	first.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	first.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, first)
	require.NoError(t, err)
	submitted := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	submitted.QuoteDigest, submitted.QuoteExpiry, submitted.MinimumStableOutput = first.QuoteDigest, first.QuoteExpiry, first.MinimumStableOutput
	submitted.SwapTxHash = "DEX-TX-REQUOTE-GUARD"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, submitted)
	require.NoError(t, err)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Unix(first.QuoteExpiry+1, 0).UTC())
	replacement := observationFor(t, conversion, compliance, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	replacement.ObservedAt = s.ctx.BlockTime().Unix()
	replacement.QuoteDigest = bytes.Repeat([]byte{32}, 32)
	replacement.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	replacement.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 910)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, replacement)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)
	stored, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, uint64(2), stored.ObservationSequence)
	require.Equal(t, submitted.SwapTxHash, stored.SwapTxHash)
}

func TestFiatObservationReplacementQuoteReplayIsExactOrConflict(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	first := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	first.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	first.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	first.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, first)
	require.NoError(t, err)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(time.Unix(first.QuoteExpiry+1, 0).UTC())
	replacement := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	replacement.ObservedAt = s.ctx.BlockTime().Unix()
	replacement.QuoteDigest = bytes.Repeat([]byte{32}, 32)
	replacement.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	replacement.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 910)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, replacement)
	require.NoError(t, err)
	exact, err := s.keeper.RecordFiatConversionObservation(s.ctx, replacement)
	require.NoError(t, err)
	require.True(t, exact.ExactDuplicate)
	conflict := *replacement
	conflict.QuoteDigest = bytes.Repeat([]byte{33}, 32)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, &conflict)
	require.ErrorIs(t, err, types.ErrFiatObservationReplayConflict)
	stored, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, replacement.QuoteDigest, stored.QuoteDigest)
	require.Equal(t, uint64(2), stored.ObservationSequence)
}

func TestFiatObservationRejectsIllegalTransitionFinalityAndMinOutput(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	illegal := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	illegal.SwapTxHash = "DETACHED"
	illegal.QuoteDigest = bytes.Repeat([]byte{30}, 32)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, illegal)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)

	quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	quote.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	quote.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, quote)
	require.NoError(t, err)
	submitted := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	submitted.QuoteDigest, submitted.QuoteExpiry, submitted.MinimumStableOutput = quote.QuoteDigest, quote.QuoteExpiry, quote.MinimumStableOutput
	submitted.SwapTxHash = "DEX-TX"
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, submitted)
	require.NoError(t, err)

	finalized := observationFor(t, conversion, compliance, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED)
	finalized.QuoteDigest = quote.QuoteDigest
	finalized.MinimumStableOutput = quote.MinimumStableOutput
	finalized.SwapTxHash = submitted.SwapTxHash
	finalized.SwapHeight = 1
	finalized.SwapBlockHash = bytes.Repeat([]byte{1}, 32)
	finalized.SwapFinalityHash = bytes.Repeat([]byte{2}, 32)
	finalized.SwapFinalityConfirmations = 1
	finalized.StableAmount = sdk.NewInt64Coin(testStableDenom, 899)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, finalized)
	require.ErrorIs(t, err, types.ErrFiatObservationEvidence)
	stored, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.Equal(t, uint64(2), stored.ObservationSequence)

	finalized.IdempotencyKey = bytes.Repeat([]byte{4}, 32)
	finalized.SwapFinalityConfirmations = 2
	finalized.StableAmount = sdk.NewInt64Coin("uwrong", 950)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, finalized)
	require.ErrorIs(t, err, types.ErrFiatObservationEvidence)
}

func TestFiatObservationComplianceRevocationFailsClosed(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	msg := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	msg.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	msg.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	msg.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	compliance.Status = veidtypes.ComplianceStatusBlocked
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.ErrorIs(t, err, types.ErrComplianceRequired)
}

func TestFiatObservationFailureCancellationAndObservedTimeBounds(t *testing.T) {
	t.Run("cancel before swap", func(t *testing.T) {
		s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
		cancelled := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED)
		cancelled.FailureCode = testFiatCancellationCode
		cancelled.Status = "cancelled"
		result, err := s.keeper.RecordFiatConversionObservation(s.ctx, cancelled)
		require.NoError(t, err)
		require.Equal(t, types.FiatConversionStateCancelled, result.Conversion.State)
		linked, _ := s.keeper.GetPayout(s.ctx, payout.PayoutID)
		require.Equal(t, types.PayoutStatePending, linked.State)
	})

	t.Run("failure after swap submission", func(t *testing.T) {
		s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
		quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
		quote.QuoteDigest = bytes.Repeat([]byte{31}, 32)
		quote.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
		quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
		_, err := s.keeper.RecordFiatConversionObservation(s.ctx, quote)
		require.NoError(t, err)
		submitted := observationFor(t, conversion, compliance, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
		submitted.QuoteDigest, submitted.QuoteExpiry, submitted.MinimumStableOutput = quote.QuoteDigest, quote.QuoteExpiry, quote.MinimumStableOutput
		submitted.SwapTxHash = "DEX-TX-FAILED"
		_, err = s.keeper.RecordFiatConversionObservation(s.ctx, submitted)
		require.NoError(t, err)
		failed := observationFor(t, conversion, compliance, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED)
		failed.FailureCode = "DEX_FINALITY_TIMEOUT"
		failed.Status = "failed"
		result, err := s.keeper.RecordFiatConversionObservation(s.ctx, failed)
		require.NoError(t, err)
		require.Equal(t, types.FiatConversionStateFailed, result.Conversion.State)
		linked, _ := s.keeper.GetPayout(s.ctx, payout.PayoutID)
		require.Equal(t, types.PayoutStatePending, linked.State)

		illegalCancel := observationFor(t, conversion, compliance, 4, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED)
		illegalCancel.FailureCode = testFiatCancellationCode
		_, err = s.keeper.RecordFiatConversionObservation(s.ctx, illegalCancel)
		require.ErrorIs(t, err, types.ErrInvalidStateTransition)
	})

	t.Run("observed time outside block bounds", func(t *testing.T) {
		s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
		quote := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
		quote.QuoteDigest = bytes.Repeat([]byte{31}, 32)
		quote.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
		quote.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
		quote.ObservedAt = s.ctx.BlockTime().Add(-2 * time.Hour).Unix()
		_, err := s.keeper.RecordFiatConversionObservation(s.ctx, quote)
		require.ErrorIs(t, err, types.ErrFiatObservationEvidence)
	})
}

func TestFiatRequestDailyLimitAndFinancialCaseHold(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	params := s.keeper.GetParams(s.ctx)
	params.FiatConversionDailyLimit = "1000"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))
	secondSettlement := s.buildSettlement(t, "task85b-daily")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, secondSettlement))
	secondPayout := types.NewPayoutRecord("payout-daily-2", "invoice-daily-2", secondSettlement.SettlementID, secondSettlement.EscrowID, secondSettlement.OrderID, secondSettlement.LeaseID, secondSettlement.Provider, secondSettlement.Customer, secondSettlement.TotalAmount, secondSettlement.PlatformFee, secondSettlement.ValidatorFee, sdk.NewCoins(), s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(t, s.keeper.SetPayout(s.ctx, *secondPayout))
	second := types.FiatConversionRequest{
		InvoiceID: "invoice-daily-2", SettlementID: secondSettlement.SettlementID,
		PayoutID: secondPayout.PayoutID,
		Provider: secondSettlement.Provider, Customer: secondSettlement.Customer, RequestedBy: secondSettlement.Provider,
		CryptoAmount: secondPayout.NetAmount[0], FiatCurrency: "USD", PaymentMethod: "bank_transfer",
		DestinationHash: types.HashDestination("opaque-two"), SlippageTolerance: 0.01,
		CryptoToken:      types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:      types.TokenSpec{Symbol: testStableSymbol, Denom: testStableDenom, Decimals: 6},
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, second)
	require.ErrorIs(t, err, types.ErrFiatLimitExceeded)

	s.keeper.ActivateFinancialCases(s.ctx)
	idempotency := bytes.Repeat([]byte{71}, 32)
	_, _, _, err = s.keeper.OpenFinancialCase(s.ctx, keeper.FinancialCaseOpenRequest{
		Subject:  types.FinancialSubject{Type: types.FinancialSubjectTypeSettlement, PrimaryId: conversion.SettlementID, SettlementId: conversion.SettlementID, InvoiceId: conversion.InvoiceID, OrderId: conversion.OrderID, EscrowId: conversion.EscrowID, LeaseId: conversion.LeaseID},
		Claimant: conversion.Customer, Respondent: conversion.Provider, IdempotencyKey: idempotency,
		Claim: types.FinancialClaim{ClaimType: types.FinancialClaimTypeBilling, Claimant: conversion.Customer, SourceModule: "settlement", SourceReference: conversion.ConversionID, EvidenceHash: bytes.Repeat([]byte{72}, 32), EncryptedReference: "settlement://task85b/evidence", IdempotencyKey: idempotency},
	})
	require.NoError(t, err)
	msg := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	msg.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	msg.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	msg.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err = s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.ErrorIs(t, err, types.ErrDisputeActive)
}

func TestFiatRequestPayloadConflictDailyAndAtomicRollback(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	settlement, _ := s.keeper.GetSettlement(s.ctx, conversion.SettlementID)
	request := types.FiatConversionRequest{
		InvoiceID: conversion.InvoiceID, SettlementID: conversion.SettlementID,
		Provider: conversion.Provider, Customer: conversion.Customer, RequestedBy: conversion.Provider,
		CryptoAmount: conversion.CryptoAmount, FiatCurrency: conversion.FiatCurrency, PaymentMethod: conversion.PaymentMethod,
		DestinationHash: conversion.DestinationHash, SlippageTolerance: conversion.SlippageTolerance,
		CryptoToken: conversion.CryptoToken, StableToken: conversion.StableToken,
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_ = settlement
	request.FiatCurrency = "EUR"
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrFiatConversionIdempotencyConflict)
	stored, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.Equal(t, "USD", stored.FiatCurrency)
}

func TestFiatRequestMinimumLimitRejectsWithoutState(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	params := s.keeper.GetParams(s.ctx)
	params.FiatConversionMinAmount = "1001"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))
	settlement := s.buildSettlement(t, "task85b-minimum")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))
	request := types.FiatConversionRequest{
		InvoiceID: "invoice-minimum", SettlementID: settlement.SettlementID,
		Provider: conversion.Provider, Customer: conversion.Customer, RequestedBy: conversion.Provider,
		CryptoAmount: sdk.NewInt64Coin("uve", 1000), FiatCurrency: "USD", PaymentMethod: "bank_transfer",
		DestinationHash: types.HashDestination("minimum"), SlippageTolerance: 0.01,
		CryptoToken: types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6}, StableToken: conversion.StableToken,
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrFiatLimitExceeded)
	_, found := s.keeper.GetFiatConversionByInvoice(s.ctx, request.InvoiceID)
	require.False(t, found)
}

func TestFiatRequestRequiresExistingPendingHeldPayoutAndExactAmount(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	settlement := s.buildSettlement(t, "task85b-lineage-fresh")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))
	payout := types.NewPayoutRecord("payout-lineage-fresh", "invoice-lineage-fresh", settlement.SettlementID, settlement.EscrowID, settlement.OrderID, settlement.LeaseID, settlement.Provider, settlement.Customer, settlement.TotalAmount, settlement.PlatformFee, settlement.ValidatorFee, sdk.NewCoins(), s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))

	request := types.FiatConversionRequest{
		InvoiceID: payout.InvoiceID, SettlementID: payout.SettlementID, PayoutID: "missing-payout",
		Provider: payout.Provider, Customer: payout.Customer, RequestedBy: payout.Provider,
		CryptoAmount: payout.NetAmount[0], FiatCurrency: conversion.FiatCurrency, PaymentMethod: conversion.PaymentMethod,
		DestinationHash: conversion.DestinationHash, SlippageToleranceExact: "0.01",
		CryptoToken: conversion.CryptoToken, StableToken: conversion.StableToken,
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrPayoutNotFound)

	request.PayoutID = payout.PayoutID
	request.CryptoAmount = sdk.NewCoin(payout.NetAmount[0].Denom, payout.NetAmount[0].Amount.AddRaw(1))
	_, err = s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrInvalidAmount)

	require.NoError(t, payout.MarkProcessing(s.ctx.BlockTime()))
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
	request.CryptoAmount = payout.NetAmount[0]
	_, err = s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrPayoutHeld)
	require.Equal(t, settlement.SettlementID, payout.SettlementID)
}

func TestFiatRequestCanonicalDigestRejectsFloatAndNilPayloadWithoutPanic(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	settlement := s.buildSettlement(t, "task85b-digest-fresh")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))
	payout := types.NewPayoutRecord("payout-digest-fresh", "invoice-digest-fresh", settlement.SettlementID, settlement.EscrowID, settlement.OrderID, settlement.LeaseID, settlement.Provider, settlement.Customer, settlement.TotalAmount, settlement.PlatformFee, settlement.ValidatorFee, sdk.NewCoins(), s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
	request := types.FiatConversionRequest{
		InvoiceID: payout.InvoiceID, SettlementID: payout.SettlementID, PayoutID: payout.PayoutID,
		Provider: payout.Provider, Customer: payout.Customer, RequestedBy: payout.Provider,
		CryptoAmount: payout.NetAmount[0], FiatCurrency: conversion.FiatCurrency, PaymentMethod: conversion.PaymentMethod,
		DestinationHash: conversion.DestinationHash, SlippageToleranceExact: "0.010000000000000000",
		CryptoToken: conversion.CryptoToken, StableToken: conversion.StableToken,
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	created, err := s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)
	require.Equal(t, "0.010000000000000000", created.SlippageToleranceExact)

	request.SlippageToleranceExact = "0.02"
	request.SlippageTolerance = math.NaN()
	_, err = s.keeper.RequestFiatConversion(s.ctx, request)
	require.ErrorIs(t, err, types.ErrInvalidParams)

	request.InvoiceID = "nil-payload"
	request.PayoutID = ""
	request.EncryptedPayload = nil
	require.NotPanics(t, func() {
		_, err = s.keeper.RequestFiatConversion(s.ctx, request)
	})
	require.Error(t, err)
}

func TestFiatRequestExactSlippageReplayProperty(t *testing.T) {
	values := []string{"0", "0.000001", "0.01", "0.500000000000000000", "1"}
	for index, exact := range values {
		t.Run(exact, func(t *testing.T) {
			s, template, _, _ := setupAuthenticatedFiatConversion(t)
			settlement := s.buildSettlement(t, "task85b-slippage-property-"+exact)
			require.NoError(t, s.keeper.SetSettlement(s.ctx, settlement))
			payout := types.NewPayoutRecord(
				"payout-slippage-property-"+exact, "invoice-slippage-property-"+exact, settlement.SettlementID,
				settlement.EscrowID, settlement.OrderID, settlement.LeaseID, settlement.Provider, settlement.Customer,
				settlement.TotalAmount, settlement.PlatformFee, settlement.ValidatorFee, sdk.NewCoins(), s.ctx.BlockTime(), s.ctx.BlockHeight(),
			)
			require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))
			request := types.FiatConversionRequest{
				InvoiceID: payout.InvoiceID, SettlementID: payout.SettlementID, PayoutID: payout.PayoutID,
				Provider: payout.Provider, Customer: payout.Customer, RequestedBy: payout.Provider,
				CryptoAmount: payout.NetAmount[0], FiatCurrency: template.FiatCurrency, PaymentMethod: template.PaymentMethod,
				DestinationHash: types.HashDestination("slippage-property-" + exact), SlippageToleranceExact: exact,
				PreferredDEX: "caller-cannot-select-dex-" + string(rune('a'+index)), PreferredOffRamp: "caller-cannot-select-payout",
				CryptoToken: template.CryptoToken, StableToken: template.StableToken,
				EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
			}
			first, err := s.keeper.RequestFiatConversion(s.ctx, request)
			require.NoError(t, err)
			require.Equal(t, exact, first.SlippageToleranceExact)
			require.Equal(t, s.keeper.GetParams(s.ctx).FiatConversionDEXProfileID, first.DEXProfileID)
			second, err := s.keeper.RequestFiatConversion(s.ctx, request)
			require.NoError(t, err)
			require.Equal(t, first.ConversionID, second.ConversionID)
			require.Equal(t, first.RequestDigest, second.RequestDigest)
		})
	}
}

func TestFiatQuotaReleasedOnlyForPreSwapCancellation(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	params := s.keeper.GetParams(s.ctx)
	params.FiatConversionDailyLimit = conversion.CryptoAmount.Amount.String()
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	cancelled := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_CANCELLED)
	cancelled.FailureCode = testFiatCancellationCode
	cancelled.Status = "cancelled"
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, cancelled)
	require.NoError(t, err)
	require.Empty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))

	secondSettlement := s.buildSettlement(t, "task85b-quota-reuse")
	require.NoError(t, s.keeper.SetSettlement(s.ctx, secondSettlement))
	secondPayout := types.NewPayoutRecord(
		"payout-quota-reuse", "invoice-quota-reuse", secondSettlement.SettlementID, secondSettlement.EscrowID,
		secondSettlement.OrderID, secondSettlement.LeaseID, secondSettlement.Provider, secondSettlement.Customer,
		secondSettlement.TotalAmount, secondSettlement.PlatformFee, secondSettlement.ValidatorFee, sdk.NewCoins(),
		s.ctx.BlockTime(), s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetPayout(s.ctx, *secondPayout))
	request := types.FiatConversionRequest{
		InvoiceID: secondPayout.InvoiceID, SettlementID: secondPayout.SettlementID, PayoutID: secondPayout.PayoutID,
		Provider: secondPayout.Provider, Customer: secondPayout.Customer, RequestedBy: secondPayout.Provider,
		CryptoAmount: secondPayout.NetAmount[0], FiatCurrency: conversion.FiatCurrency, PaymentMethod: conversion.PaymentMethod,
		DestinationHash: types.HashDestination("quota-reuse"), SlippageToleranceExact: "0.01",
		CryptoToken: conversion.CryptoToken, StableToken: conversion.StableToken,
		EncryptedPayload: makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
	}
	_, err = s.keeper.RequestFiatConversion(s.ctx, request)
	require.NoError(t, err)
}

func TestFiatUpdateParamsRequiresAuthorityAndExactCommitments(t *testing.T) {
	s, _, _, _ := setupAuthenticatedFiatConversion(t)
	server := keeper.NewMsgServerImpl(s.keeper)
	params := s.keeper.GetParams(s.ctx)
	protoParams := settlementv1.Params{
		PlatformFeeRate: params.PlatformFeeRate, ValidatorFeeRate: params.ValidatorFeeRate,
		MinEscrowDuration: params.MinEscrowDuration, MaxEscrowDuration: params.MaxEscrowDuration,
		SettlementPeriod: params.SettlementPeriod, RewardClaimExpiry: params.RewardClaimExpiry,
		MinSettlementAmount: params.MinSettlementAmount, UsageGracePeriod: params.UsageGracePeriod,
		StakingRewardEpochLength: params.StakingRewardEpochLength, VerificationRewardAmount: params.VerificationRewardAmount,
		PayoutHoldbackRate: params.PayoutHoldbackRate, MaxPayoutRetries: params.MaxPayoutRetries,
		DisputeWindowDuration: params.DisputeWindowDuration, UsageRewardRateBps: params.UsageRewardRateBps,
		UsageRewardCpuMultiplierBps: params.UsageRewardCPUMultiplierBps, UsageRewardMemoryMultiplierBps: params.UsageRewardMemoryMultiplierBps,
		UsageRewardStorageMultiplierBps: params.UsageRewardStorageMultiplierBps, UsageRewardGpuMultiplierBps: params.UsageRewardGPUMultiplierBps,
		UsageRewardNetworkMultiplierBps: params.UsageRewardNetworkMultiplierBps, UsageRewardSlaOntimeMultiplierBps: params.UsageRewardSLAOnTimeMultiplierBps,
		UsageRewardSlaLateMultiplierBps: params.UsageRewardSLALateMultiplierBps, UsageRewardAckMultiplierBps: params.UsageRewardAcknowledgedMultiplierBps,
		UsageRewardUnackMultiplierBps: params.UsageRewardUnacknowledgedMultiplierBps, FiatConversionEnabled: true,
		FiatConversionMinAmount: params.FiatConversionMinAmount, FiatConversionMaxAmount: params.FiatConversionMaxAmount,
		FiatConversionDailyLimit: params.FiatConversionDailyLimit, FiatConversionStableDenom: params.FiatConversionStableDenom,
		FiatConversionStableSymbol: params.FiatConversionStableSymbol, FiatConversionStableDecimals: params.FiatConversionStableDecimals,
		FiatConversionDefaultFiat: params.FiatConversionDefaultFiat, FiatConversionDefaultMethod: params.FiatConversionDefaultMethod,
		FiatConversionMaxSlippage: params.FiatConversionMaxSlippage, FiatConversionRiskScoreThreshold: params.FiatConversionRiskScoreThreshold,
		FiatConversionMinComplianceStatus: params.FiatConversionMinComplianceStatus,
		FiatConversionDexProfileId:        params.FiatConversionDEXProfileID, FiatConversionDexProfileDigest: append([]byte(nil), params.FiatConversionDEXProfileDigest...),
		FiatConversionDexProfileState: params.FiatConversionDEXProfileState,
		FiatConversionPayoutProfileId: params.FiatConversionPayoutProfileID, FiatConversionPayoutProfileDigest: append([]byte(nil), params.FiatConversionPayoutProfileDigest...),
		FiatConversionPayoutProfileState:           params.FiatConversionPayoutProfileState,
		FiatConversionMinSwapFinalityConfirmations: params.FiatConversionMinSwapFinalityConfirmations,
		FiatConversionObservationMaxPastSeconds:    params.FiatConversionObservationMaxPastSeconds,
		FiatConversionObservationMaxFutureSeconds:  params.FiatConversionObservationMaxFutureSeconds,
		FiatConversionMaxObservations:              params.FiatConversionMaxObservations,
		FinancialCaseFilingWindowSeconds:           params.FinancialCaseFilingWindowSeconds, FinancialCaseEvidenceWindowSeconds: params.FinancialCaseEvidenceWindowSeconds,
		FinancialCaseReviewWindowSeconds: params.FinancialCaseReviewWindowSeconds, FinancialCaseAppealWindowSeconds: params.FinancialCaseAppealWindowSeconds,
		FinancialCaseEscalationWindowSeconds: params.FinancialCaseEscalationWindowSeconds,
		FinancialCaseFilingWindowBlocks:      params.FinancialCaseFilingWindowBlocks, FinancialCaseEvidenceWindowBlocks: params.FinancialCaseEvidenceWindowBlocks,
		FinancialCaseReviewWindowBlocks: params.FinancialCaseReviewWindowBlocks, FinancialCaseAppealWindowBlocks: params.FinancialCaseAppealWindowBlocks,
		FinancialCaseEscalationWindowBlocks: params.FinancialCaseEscalationWindowBlocks, FinancialCaseMaxClaims: params.FinancialCaseMaxClaims,
		FinancialCaseMaxAppeals: params.FinancialCaseMaxAppeals, FinancialCaseMaxEvidenceReferenceBytes: params.FinancialCaseMaxEvidenceReferenceBytes,
		FinancialCaseTimeoutBatchLimit: params.FinancialCaseTimeoutBatchLimit,
	}
	_, err := server.UpdateParams(s.ctx, &settlementv1.MsgUpdateParams{Authority: s.provider.String(), Params: protoParams})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	protoParams.FiatConversionDexProfileDigest = bytes.Repeat([]byte{1}, 31)
	_, err = server.UpdateParams(s.ctx, &settlementv1.MsgUpdateParams{Authority: s.keeper.GetAuthority(), Params: protoParams})
	require.ErrorIs(t, err, types.ErrInvalidParams)
}

func TestFiatConversionInvariantDetectsMalformedAndDuplicateIndexes(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	store := s.ctx.KVStore(s.storeKey)
	store.Set(append(append([]byte(nil), types.PrefixFiatObservationReplay...), []byte("malformed")...), bytes.Repeat([]byte{1}, 32))
	store.Set(types.FiatDailyTotalKey(conversion.Provider, "19990101"), []byte("1"))
	broken := s.keeper.ValidateFiatConversionInvariants(s.ctx)
	require.NotEmpty(t, broken)
}

func TestFiatConversionInvariantDetectsPayoutAndLedgerCorruption(t *testing.T) {
	s, _, payout, _ := setupAuthenticatedFiatConversion(t)
	store := s.ctx.KVStore(s.storeKey)
	store.Delete(types.PayoutByStateKey(payout.State, payout.PayoutID))
	store.Set(types.PayoutByProviderKey(s.depositor.String(), payout.PayoutID), []byte{})
	store.Set(types.PayoutLedgerByPayoutKey(payout.PayoutID, "missing-entry"), []byte("missing-entry"))
	require.NotEmpty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))
}

func TestFiatConversionInvariantDetectsCustodySinkBalanceMismatch(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	completed := runSuccessfulFiatObservationFlow(t, s, conversion, compliance)
	require.Empty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))

	s.bankKeeper.SetModuleBalance(types.FiatConversionCustodyAccountName, sdk.NewCoins())
	broken := s.keeper.ValidateFiatConversionInvariants(s.ctx)
	require.NotEmpty(t, broken)
	require.True(t, completed.ValueMovementApplied)
}

func TestFiatConversionInvariantDetectsReplayCorruption(t *testing.T) {
	s, conversion, _, compliance := setupAuthenticatedFiatConversion(t)
	msg := observationFor(t, conversion, compliance, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	msg.QuoteDigest = bytes.Repeat([]byte{31}, 32)
	msg.QuoteExpiry = s.ctx.BlockTime().Add(time.Minute).Unix()
	msg.MinimumStableOutput = sdk.NewInt64Coin(testStableDenom, 900)
	_, err := s.keeper.RecordFiatConversionObservation(s.ctx, msg)
	require.NoError(t, err)
	require.Empty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))
	s.ctx.KVStore(s.storeKey).Set(types.FiatObservationSequenceKey(conversion.ConversionID, 1), bytes.Repeat([]byte{99}, 32))
	require.NotEmpty(t, s.keeper.ValidateFiatConversionInvariants(s.ctx))
}

func TestMigrateFiatConversionsQuarantinesActivePreservesTerminalIdempotently(t *testing.T) {
	s, conversion, _, _ := setupAuthenticatedFiatConversion(t)
	conversion.ProtocolVersion = 0
	conversion.LegacyQuarantined = false
	require.NoError(t, s.keeper.SetFiatConversion(s.ctx, conversion))
	report, err := s.keeper.MigrateFiatConversions(s.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), report.ActiveQuarantined)
	migrated, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, migrated.LegacyQuarantined)
	retry, err := s.keeper.MigrateFiatConversions(s.ctx)
	require.NoError(t, err)
	require.Equal(t, report.Digest, retry.Digest)
	require.Equal(t, uint64(1), retry.AlreadyMigrated)
	retryRecord, _ := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, retryRecord.LegacyQuarantined)
}

func TestMigrateFiatConversionsPreservesTerminalHistoricalRecord(t *testing.T) {
	s, conversion, payout, compliance := setupAuthenticatedFiatConversion(t)
	completed := runSuccessfulFiatObservationFlow(t, s, conversion, compliance)
	completed.ProtocolVersion = 0
	require.NoError(t, s.keeper.SetFiatConversion(s.ctx, completed))
	report, err := s.keeper.MigrateFiatConversions(s.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), report.TerminalPreserved)
	migrated, found := s.keeper.GetFiatConversion(s.ctx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStatePayoutCompleted, migrated.State)
	require.True(t, migrated.LegacyQuarantined)
	require.False(t, migrated.ValueMovementApplied)
	require.Empty(t, migrated.CustodySinkEffectHash)
	linked, found := s.keeper.GetPayout(s.ctx, payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, linked.State)
	require.False(t, linked.ValueMovementApplied)
}
