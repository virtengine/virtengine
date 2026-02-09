package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func (s *KeeperTestSuite) seedQueryData(t require.TestingT) (types.EscrowAccount, types.SettlementRecord, types.UsageRecord, types.RewardDistribution, types.PayoutRecord, types.FiatConversionRecord, types.FiatPayoutPreference) {
	now := s.ctx.BlockTime()

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	escrow := types.NewEscrowAccount(
		"escrow-1",
		"order-1",
		s.depositor.String(),
		amount,
		now.Add(time.Hour),
		nil,
		now,
		s.ctx.BlockHeight(),
	)
	escrow.State = types.EscrowStateActive
	escrow.Recipient = s.provider.String()
	require.NoError(t, s.keeper.SetEscrow(s.ctx, *escrow))

	usage := types.NewUsageRecord(
		"usage-1",
		"order-1",
		"lease-1",
		s.provider.String(),
		s.depositor.String(),
		100,
		"compute",
		now.Add(-time.Hour),
		now,
		sdk.NewDecCoin("uve", sdkmath.NewInt(1)),
		[]byte("signature"),
		now,
		s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetUsageRecord(s.ctx, *usage))

	settlement := types.NewSettlementRecord(
		"settlement-1",
		escrow.EscrowID,
		escrow.OrderID,
		"lease-1",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(800))),
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200))),
		sdk.NewCoins(),
		[]string{usage.UsageID},
		usage.UsageUnits,
		now.Add(-time.Hour),
		now,
		types.SettlementTypeUsageBased,
		false,
		now,
		s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetSettlement(s.ctx, *settlement))

	recipient := types.RewardRecipient{
		Address: s.provider.String(),
		Amount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10))),
		Reason:  "test-reward",
	}
	distribution := types.NewRewardDistribution(
		"dist-1",
		1,
		types.RewardSourceUsage,
		[]types.RewardRecipient{recipient},
		now,
		s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetRewardDistribution(s.ctx, *distribution))

	claimable := types.NewClaimableRewards(s.provider.String(), now)
	claimable.AddReward(types.RewardEntry{
		DistributionID: distribution.DistributionID,
		Source:         distribution.Source,
		Amount:         recipient.Amount,
		CreatedAt:      now,
		Reason:         recipient.Reason,
	})
	require.NoError(t, s.keeper.SetClaimableRewards(s.ctx, s.provider, *claimable))

	payout := types.NewPayoutRecord(
		"payout-1",
		"invoice-1",
		settlement.SettlementID,
		escrow.EscrowID,
		escrow.OrderID,
		"lease-1",
		s.provider.String(),
		s.depositor.String(),
		amount,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
		sdk.NewCoins(),
		sdk.NewCoins(),
		now,
		s.ctx.BlockHeight(),
	)
	require.NoError(t, s.keeper.SetPayout(s.ctx, *payout))

	request := types.FiatConversionRequest{
		InvoiceID:         payout.InvoiceID,
		SettlementID:      settlement.SettlementID,
		PayoutID:          payout.PayoutID,
		Provider:          s.provider.String(),
		Customer:          s.depositor.String(),
		RequestedBy:       s.provider.String(),
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(100)),
		FiatCurrency:      "USD",
		PaymentMethod:     "bank",
		Destination:       "acct-1",
		SlippageTolerance: 0.1,
		CryptoToken: types.TokenSpec{
			Symbol:   "UVE",
			Denom:    "uve",
			Decimals: 6,
		},
		StableToken: types.TokenSpec{
			Symbol:   "USDC",
			Denom:    "uusdc",
			Decimals: 6,
		},
	}
	conversion := types.NewFiatConversionRecord("conv-1", request, request.CryptoAmount, now)
	require.NoError(t, s.keeper.SetFiatConversion(s.ctx, *conversion))

	preference := types.FiatPayoutPreference{
		Provider:  s.provider.String(),
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.keeper.SetFiatPayoutPreference(s.ctx, preference))

	return *escrow, *settlement, *usage, *distribution, *payout, *conversion, preference
}

func (s *KeeperTestSuite) TestGRPCQuerierQueries() {
	escrow, settlement, usage, distribution, payout, conversion, preference := s.seedQueryData(s.T())

	querier := keeper.GRPCQuerier{Keeper: s.keeper}
	ctx := sdk.WrapSDKContext(s.ctx)

	respEscrow, err := querier.Escrow(ctx, &types.QueryEscrowRequest{EscrowID: escrow.EscrowID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), escrow.EscrowID, respEscrow.Escrow.EscrowID)

	respEscrowsByOrder, err := querier.EscrowsByOrder(ctx, &types.QueryEscrowsByOrderRequest{OrderID: escrow.OrderID})
	require.NoError(s.T(), err)
	require.Len(s.T(), respEscrowsByOrder.Escrows, 1)

	respEscrowsByState, err := querier.EscrowsByState(ctx, &types.QueryEscrowsByStateRequest{State: string(types.EscrowStateActive)})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), respEscrowsByState.Escrows)

	respSettlement, err := querier.Settlement(ctx, &types.QuerySettlementRequest{SettlementID: settlement.SettlementID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), settlement.SettlementID, respSettlement.Settlement.SettlementID)

	respSettlementsByOrder, err := querier.SettlementsByOrder(ctx, &types.QuerySettlementsByOrderRequest{OrderID: settlement.OrderID})
	require.NoError(s.T(), err)
	require.Len(s.T(), respSettlementsByOrder.Settlements, 1)

	respUsageRecord, err := querier.UsageRecord(ctx, &types.QueryUsageRecordRequest{UsageID: usage.UsageID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), usage.UsageID, respUsageRecord.UsageRecord.UsageID)

	respUsageByOrder, err := querier.UsageRecordsByOrder(ctx, &types.QueryUsageRecordsByOrderRequest{OrderID: usage.OrderID})
	require.NoError(s.T(), err)
	require.Len(s.T(), respUsageByOrder.UsageRecords, 1)

	respSummary, err := querier.UsageSummary(ctx, &types.QueryUsageSummaryRequest{OrderID: usage.OrderID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), usage.OrderID, respSummary.Summary.OrderID)

	respDist, err := querier.RewardDistribution(ctx, &types.QueryRewardDistributionRequest{DistributionID: distribution.DistributionID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), distribution.DistributionID, respDist.Distribution.DistributionID)

	respEpoch, err := querier.RewardsByEpoch(ctx, &types.QueryRewardsByEpochRequest{EpochNumber: distribution.EpochNumber})
	require.NoError(s.T(), err)
	require.Len(s.T(), respEpoch.Distributions, 1)

	respHistory, err := querier.RewardHistory(ctx, &types.QueryRewardHistoryRequest{Address: s.provider.String()})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), respHistory.Entries)

	respClaimable, err := querier.ClaimableRewards(ctx, &types.QueryClaimableRewardsRequest{Address: s.provider.String()})
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.provider.String(), respClaimable.Rewards.Address)

	respPayout, err := querier.Payout(ctx, &types.QueryPayoutRequest{PayoutID: payout.PayoutID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), payout.PayoutID, respPayout.Payout.PayoutID)

	respPayoutsByProvider, err := querier.PayoutsByProvider(ctx, &types.QueryPayoutsByProviderRequest{Provider: s.provider.String()})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), respPayoutsByProvider.Payouts)

	respParams, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), respParams.Params)

	respFiatConversion, err := querier.FiatConversion(ctx, &types.QueryFiatConversionRequest{ConversionID: conversion.ConversionID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), conversion.ConversionID, respFiatConversion.Conversion.ConversionID)

	respFiatConversions, err := querier.FiatConversionsByProvider(ctx, &types.QueryFiatConversionsByProviderRequest{Provider: s.provider.String()})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), respFiatConversions.Conversions)

	respFiatPreference, err := querier.FiatPayoutPreference(ctx, &types.QueryFiatPayoutPreferenceRequest{Provider: preference.Provider})
	require.NoError(s.T(), err)
	require.Equal(s.T(), preference.Provider, respFiatPreference.Preference.Provider)
}

func (s *KeeperTestSuite) TestGRPCQuerierNilRequests() {
	querier := keeper.GRPCQuerier{Keeper: s.keeper}
	ctx := sdk.WrapSDKContext(s.ctx)

	requests := []struct {
		name string
		call func() error
	}{
		{"Escrow", func() error { _, err := querier.Escrow(ctx, nil); return err }},
		{"EscrowsByOrder", func() error { _, err := querier.EscrowsByOrder(ctx, nil); return err }},
		{"EscrowsByState", func() error { _, err := querier.EscrowsByState(ctx, nil); return err }},
		{"Settlement", func() error { _, err := querier.Settlement(ctx, nil); return err }},
		{"SettlementsByOrder", func() error { _, err := querier.SettlementsByOrder(ctx, nil); return err }},
		{"UsageRecord", func() error { _, err := querier.UsageRecord(ctx, nil); return err }},
		{"UsageRecordsByOrder", func() error { _, err := querier.UsageRecordsByOrder(ctx, nil); return err }},
		{"UsageSummary", func() error { _, err := querier.UsageSummary(ctx, nil); return err }},
		{"RewardDistribution", func() error { _, err := querier.RewardDistribution(ctx, nil); return err }},
		{"RewardsByEpoch", func() error { _, err := querier.RewardsByEpoch(ctx, nil); return err }},
		{"RewardHistory", func() error { _, err := querier.RewardHistory(ctx, nil); return err }},
		{"ClaimableRewards", func() error { _, err := querier.ClaimableRewards(ctx, nil); return err }},
		{"Payout", func() error { _, err := querier.Payout(ctx, nil); return err }},
		{"PayoutsByProvider", func() error { _, err := querier.PayoutsByProvider(ctx, nil); return err }},
		{"Params", func() error { _, err := querier.Params(ctx, nil); return err }},
		{"FiatConversion", func() error { _, err := querier.FiatConversion(ctx, nil); return err }},
		{"FiatConversionsByProvider", func() error { _, err := querier.FiatConversionsByProvider(ctx, nil); return err }},
		{"FiatPayoutPreference", func() error { _, err := querier.FiatPayoutPreference(ctx, nil); return err }},
	}

	for _, tc := range requests {
		s.T().Run(tc.name, func(t *testing.T) {
			require.Error(t, tc.call())
		})
	}
}
