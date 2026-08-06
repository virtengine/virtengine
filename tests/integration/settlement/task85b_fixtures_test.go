// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

//go:build e2e.integration

package settlement_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/testutil/state"
	"github.com/virtengine/virtengine/x/settlement/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	task85BDEXProfileID    = "integration-dex-profile"
	task85BPayoutProfileID = "integration-payout-profile"
	task85BStableDenom     = "uusdc"
)

type task85BFiatFixture struct {
	suite      *state.TestSuite
	ctx        sdk.Context
	provider   sdk.AccAddress
	customer   sdk.AccAddress
	compliance *veidtypes.ComplianceRecord
	payout     types.PayoutRecord
	conversion types.FiatConversionRecord
	request    types.FiatConversionRequest
}

func setupTask85BFiatFixture(t *testing.T) *task85BFiatFixture {
	t.Helper()
	suite := state.SetupTestSuiteWithoutModuleServices(t)
	ctx := suite.Context()
	keeper := &suite.App().Keepers.VirtEngine.Settlement
	provider := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	customer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	params := keeper.GetParams(ctx)
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	params.FiatConversionEnabled = true
	params.FiatConversionDEXProfileID = task85BDEXProfileID
	params.FiatConversionDEXProfileDigest = bytes.Repeat([]byte{0xd1}, sha256.Size)
	params.FiatConversionDEXProfileState = certified
	params.FiatConversionPayoutProfileID = task85BPayoutProfileID
	params.FiatConversionPayoutProfileDigest = bytes.Repeat([]byte{0xe1}, sha256.Size)
	params.FiatConversionPayoutProfileState = certified
	params.FiatConversionMinAmount = "1"
	params.FiatConversionMaxAmount = "1000000000"
	params.FiatConversionDailyLimit = "10000000000"
	params.FiatConversionStableDenom = task85BStableDenom
	params.FiatConversionStableSymbol = "USDC"
	params.FiatConversionStableDecimals = 6
	params.FiatConversionMaxSlippage = "0.05"
	params.FiatConversionMinComplianceStatus = "CLEARED"
	require.NoError(t, keeper.SetParams(ctx, params))

	keeper.SetComplianceKeeper(suite.App().Keepers.VirtEngine.VEID)
	compliance := veidtypes.NewComplianceRecord(provider.String(), ctx.BlockTime())
	compliance.Status = veidtypes.ComplianceStatusCleared
	compliance.RiskScore = 5
	compliance.ExpiresAt = ctx.BlockTime().Add(24 * time.Hour).Unix()
	require.NoError(t, suite.App().Keepers.VirtEngine.VEID.SetComplianceRecord(ctx, compliance))
	usageAuth := newAuthenticatedUsageFixture(t, suite, provider)
	bank := suite.App().Keepers.Cosmos.Bank
	coins := sdk.NewCoins(sdk.NewInt64Coin("uve", 100_000))
	require.NoError(t, bank.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, bank.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, customer, coins))
	require.NoError(t, keeper.SetFiatPayoutPreference(ctx, makeFiatPayoutPreference(t, suite, ctx.BlockTime(), provider, "uve", "acct-task85b")))

	escrowID, err := keeper.CreateEscrow(ctx, "order-task85b", customer, sdk.NewCoins(sdk.NewInt64Coin("uve", 1_000)), 24*time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, keeper.ActivateEscrow(ctx, escrowID, "lease-task85b", provider))
	usage := usageAuth.newUsage(t, ctx, "order-task85b", "lease-task85b", provider, customer,
		"compute", 1, 1_000, ctx.BlockTime().Add(-time.Hour), ctx.BlockTime())
	require.NoError(t, keeper.RecordUsage(ctx, usage))
	require.True(t, usage.IsAuthenticated())
	settlement, err := keeper.SettleOrder(ctx, "order-task85b", []string{usage.UsageID}, false)
	require.NoError(t, err)
	payout, found := keeper.GetPayoutBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, payout.State)
	conversion, found := keeper.GetFiatConversionBySettlement(ctx, settlement.SettlementID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStateCreated, conversion.State)

	return &task85BFiatFixture{
		suite: suite, ctx: ctx, provider: provider, customer: customer, compliance: compliance,
		payout: payout, conversion: conversion,
		request: types.FiatConversionRequest{
			InvoiceID: conversion.InvoiceID, SettlementID: conversion.SettlementID, PayoutID: conversion.PayoutID,
			Provider: conversion.Provider, Customer: conversion.Customer, RequestedBy: conversion.RequestedBy,
			CryptoAmount: conversion.CryptoAmount, FiatCurrency: conversion.FiatCurrency, PaymentMethod: conversion.PaymentMethod,
			DestinationHash: conversion.DestinationHash, DestinationRegion: conversion.DestinationRegion,
			SlippageTolerance: conversion.SlippageTolerance, SlippageToleranceExact: conversion.SlippageToleranceExact,
			CryptoToken: conversion.CryptoToken, StableToken: conversion.StableToken, EncryptedPayload: conversion.EncryptedPayload,
		},
	}
}

func task85BObservation(t *testing.T, fixture *task85BFiatFixture, sequence uint64, stage settlementv1.FiatConversionObservationStage) *types.MsgRecordFiatConversionObservation {
	t.Helper()
	encoded, err := json.Marshal(fixture.compliance)
	require.NoError(t, err)
	complianceDigest := sha256.Sum256(encoded)
	return &types.MsgRecordFiatConversionObservation{
		Sender: fixture.provider.String(), ConversionId: fixture.conversion.ConversionID,
		ObservationSequence: sequence, IdempotencyKey: bytes.Repeat([]byte{byte(sequence)}, sha256.Size), Stage: stage,
		DexProfileId: fixture.conversion.DEXProfileID, DexProfileDigest: append([]byte(nil), fixture.conversion.DEXProfileDigest...),
		PayoutProfileId: fixture.conversion.PayoutProfileID, PayoutProfileDigest: append([]byte(nil), fixture.conversion.PayoutProfileDigest...),
		ObservedAt: fixture.ctx.BlockTime().Unix(), EvidenceHash: bytes.Repeat([]byte{byte(sequence + 20)}, sha256.Size),
		ComplianceDecisionHash: complianceDigest[:], Status: "accepted",
	}
}

func task85BSuccessfulObservations(t *testing.T, fixture *task85BFiatFixture) []*types.MsgRecordFiatConversionObservation {
	t.Helper()
	quote := task85BObservation(t, fixture, 1, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	quote.QuoteDigest, quote.QuoteExpiry = bytes.Repeat([]byte{31}, sha256.Size), fixture.ctx.BlockTime().Add(10*time.Minute).Unix()
	quote.MinimumStableOutput = sdk.NewInt64Coin(task85BStableDenom, 900)
	submitted := task85BObservation(t, fixture, 2, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	submitted.QuoteDigest, submitted.QuoteExpiry, submitted.MinimumStableOutput = quote.QuoteDigest, quote.QuoteExpiry, quote.MinimumStableOutput
	submitted.SwapTxHash = "external-dex-transaction-1"
	finalized := task85BObservation(t, fixture, 3, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED)
	finalized.QuoteDigest, finalized.MinimumStableOutput, finalized.SwapTxHash = quote.QuoteDigest, quote.MinimumStableOutput, submitted.SwapTxHash
	finalized.SwapHeight, finalized.SwapFinalityConfirmations = 100, 2
	finalized.SwapBlockHash, finalized.SwapFinalityHash = bytes.Repeat([]byte{41}, sha256.Size), bytes.Repeat([]byte{42}, sha256.Size)
	finalized.StableAmount = sdk.NewInt64Coin(task85BStableDenom, 950)
	payoutQuote := task85BObservation(t, fixture, 4, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED)
	payoutQuote.QuoteDigest, payoutQuote.QuoteExpiry, payoutQuote.OffRampQuoteId = bytes.Repeat([]byte{51}, sha256.Size), fixture.ctx.BlockTime().Add(10*time.Minute).Unix(), "offramp-quote-1"
	payoutSubmitted := task85BObservation(t, fixture, 5, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED)
	payoutSubmitted.QuoteDigest, payoutSubmitted.QuoteExpiry = payoutQuote.QuoteDigest, payoutQuote.QuoteExpiry
	payoutSubmitted.OffRampQuoteId, payoutSubmitted.OffRampPayoutId, payoutSubmitted.Status = payoutQuote.OffRampQuoteId, "offramp-payout-1", "submitted"
	payoutSubmitted.PrivacySafeReferenceHash = bytes.Repeat([]byte{61}, sha256.Size)
	completed := task85BObservation(t, fixture, 6, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED)
	completed.QuoteDigest, completed.OffRampQuoteId, completed.OffRampPayoutId = payoutQuote.QuoteDigest, payoutQuote.OffRampQuoteId, payoutSubmitted.OffRampPayoutId
	completed.Status, completed.FiatAmount = "completed", "123.45"
	completed.PrivacySafeReferenceHash, completed.PayoutFinalityHash = payoutSubmitted.PrivacySafeReferenceHash, bytes.Repeat([]byte{62}, sha256.Size)
	return []*types.MsgRecordFiatConversionObservation{quote, submitted, finalized, payoutQuote, payoutSubmitted, completed}
}
