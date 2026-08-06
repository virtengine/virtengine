// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

//go:build e2e.integration

package settlement_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestAuthenticatedFiatConversionObservationPipeline(t *testing.T) {
	fixture := setupTask85BFiatFixture(t)
	keeper := &fixture.suite.App().Keepers.VirtEngine.Settlement
	require.Equal(t, fixture.conversion.ConversionID, fixture.payout.FiatConversionID)
	require.Equal(t, fixture.payout.PayoutID, fixture.conversion.PayoutID)
	require.Equal(t, task85BDEXProfileID, fixture.conversion.DEXProfileID)
	require.Equal(t, task85BPayoutProfileID, fixture.conversion.PayoutProfileID)
	require.Len(t, fixture.conversion.RequestDigest, sha256.Size)
	require.Len(t, fixture.conversion.ComplianceDecisionHash, sha256.Size)

	retry, err := keeper.RequestFiatConversion(fixture.ctx, fixture.request)
	require.NoError(t, err)
	require.Equal(t, fixture.conversion.ConversionID, retry.ConversionID)
	require.Equal(t, fixture.conversion.RequestDigest, retry.RequestDigest)

	observations := task85BSuccessfulObservations(t, fixture)
	for index, observation := range observations {
		result, recordErr := keeper.RecordFiatConversionObservation(fixture.ctx, observation)
		require.NoError(t, recordErr, "stage %d", index+1)
		require.Equal(t, uint64(index+1), result.Conversion.ObservationSequence)
		if index == 0 {
			exact, replayErr := keeper.RecordFiatConversionObservation(fixture.ctx, observation)
			require.NoError(t, replayErr)
			require.True(t, exact.ExactDuplicate)
			conflict := *observation
			conflict.Status = "changed"
			_, replayErr = keeper.RecordFiatConversionObservation(fixture.ctx, &conflict)
			require.ErrorIs(t, replayErr, types.ErrFiatObservationReplayConflict)
		}
	}

	conversion, found := keeper.GetFiatConversion(fixture.ctx, fixture.conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, types.FiatConversionStatePayoutCompleted, conversion.State)
	require.Equal(t, uint64(6), conversion.ObservationSequence)
	require.Len(t, conversion.Observations, 6)
	require.Equal(t, "external_fiat_completed_custody_sink", conversion.TerminalPolicy)
	require.True(t, conversion.ValueMovementApplied)
	require.True(t, conversion.CustodySinkAmount.IsEqual(conversion.CryptoAmount))
	require.Len(t, conversion.CustodySinkEffectHash, sha256.Size)

	payout, found := keeper.GetPayout(fixture.ctx, fixture.payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStateCompleted, payout.State)
	require.Empty(t, payout.TxHash)
	require.Equal(t, conversion.PayoutFinalityHash, payout.ExternalFinalityHash)
	require.Equal(t, conversion.CustodySinkEffectHash, payout.ValueMovementEffectHash)
	require.True(t, fixture.suite.App().Keepers.Cosmos.Bank.GetBalance(fixture.ctx, fixture.provider, "uve").IsZero())
	require.True(t, keeper.GetFiatConversionCustodyBalance(fixture.ctx).Equal(sdk.NewCoins(conversion.CryptoAmount)))
	require.Empty(t, keeper.ValidateFiatConversionInvariants(fixture.ctx))
}

func TestAuthenticatedFiatConversionRejectsOwnerProfileDriftAndExternalIO(t *testing.T) {
	fixture := setupTask85BFiatFixture(t)
	keeper := &fixture.suite.App().Keepers.VirtEngine.Settlement
	quote := task85BSuccessfulObservations(t, fixture)[0]
	unauthorized := *quote
	unauthorized.Sender = fixture.customer.String()
	_, err := keeper.RecordFiatConversionObservation(fixture.ctx, &unauthorized)
	require.ErrorIs(t, err, types.ErrUnauthorized)
	wrongProfile := *quote
	wrongProfile.DexProfileDigest = bytes.Repeat([]byte{99}, sha256.Size)
	_, err = keeper.RecordFiatConversionObservation(fixture.ctx, &wrongProfile)
	require.ErrorIs(t, err, types.ErrFiatProfileCommitment)
	result, err := keeper.RecordFiatConversionObservation(fixture.ctx, quote)
	require.NoError(t, err)
	require.Equal(t, types.FiatConversionStateSwapPending, result.Conversion.State)
	reconciled, err := keeper.ReconcileFiatConversion(fixture.ctx, fixture.conversion.ConversionID)
	require.ErrorIs(t, err, types.ErrExternalIOForbidden)
	require.Equal(t, types.FiatConversionStateSwapPending, reconciled.State)
	payout, found := keeper.GetPayout(fixture.ctx, fixture.payout.PayoutID)
	require.True(t, found)
	require.Equal(t, types.PayoutStatePending, payout.State)
	require.False(t, payout.ValueMovementApplied)
	require.True(t, keeper.GetFiatConversionCustodyBalance(fixture.ctx).IsZero())
}
