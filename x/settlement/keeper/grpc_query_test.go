package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type grpcTestSuite struct {
	t           *testing.T
	ctx         sdk.Context
	keeper      keeper.Keeper
	queryClient settlementv1.QueryClient
}

type seededData struct {
	escrow      types.EscrowAccount
	settlement  types.SettlementRecord
	usage       types.UsageRecord
	rewardDist  types.RewardDistribution
	claimable   types.ClaimableRewards
	payout      types.PayoutRecord
	conversion  types.FiatConversionRecord
	preference  types.FiatPayoutPreference
	orderID     string
	provider    sdk.AccAddress
	customer    sdk.AccAddress
	usagePeriod struct {
		start time.Time
		end   time.Time
	}
}

func setupGRPCTest(t *testing.T) *grpcTestSuite {
	cfg := testutilmod.MakeTestEncodingConfig()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	now := time.Now().UTC()
	ctx := sdk.NewContext(stateStore, tmproto.Header{Height: 1, Time: now}, false, testutil.Logger(t))

	bankKeeper := NewMockBankKeeper()
	settlementKeeper := keeper.NewKeeper(cfg.Codec, storeKey, bankKeeper, "authority", mockEncryptionKeeper{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, cfg.InterfaceRegistry)
	settlementv1.RegisterQueryServer(queryHelper, keeper.GRPCQuerier{IKeeper: settlementKeeper})
	queryClient := settlementv1.NewQueryClient(queryHelper)

	return &grpcTestSuite{
		t:           t,
		ctx:         ctx,
		keeper:      settlementKeeper,
		queryClient: queryClient,
	}
}

func seedSettlementData(t *testing.T, suite *grpcTestSuite) seededData {
	now := suite.ctx.BlockTime()
	orderID := "order-1"
	leaseID := "lease-1"
	escrowID := "escrow-1"
	usageID := "usage-1"
	settlementID := "settlement-1"
	distributionID := "dist-1"
	payoutID := "payout-1"
	conversionID := "conv-1"

	provider := testutil.AccAddress(t)
	customer := testutil.AccAddress(t)

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	conditions := []types.ReleaseCondition{{
		Type:        types.ConditionTypeTimelock,
		UnlockAfter: ptrTime(now.Add(2 * time.Hour)),
		Satisfied:   true,
		SatisfiedAt: ptrTime(now.Add(time.Hour)),
	}}

	escrow := types.NewEscrowAccount(escrowID, orderID, customer.String(), amount, now.Add(24*time.Hour), conditions, now, suite.ctx.BlockHeight())
	escrow.LeaseID = leaseID
	escrow.Recipient = provider.String()
	escrow.State = types.EscrowStateActive
	escrow.ActivatedAt = ptrTime(now.Add(time.Minute))
	escrow.TotalSettled = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200)))
	escrow.Balance = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(800)))
	escrow.SettlementCount = 1
	require.NoError(t, suite.keeper.SetEscrow(suite.ctx, *escrow))

	periodStart := now.Add(-2 * time.Hour)
	periodEnd := now.Add(-time.Hour)
	usageRecord := types.NewUsageRecord(
		usageID,
		orderID,
		leaseID,
		provider.String(),
		customer.String(),
		10,
		"compute",
		periodStart,
		periodEnd,
		sdk.NewDecCoinFromCoin(sdk.NewCoin("uve", sdkmath.NewInt(2))),
		[]byte("provider-sig"),
		now,
		suite.ctx.BlockHeight(),
	)
	usageRecord.CustomerAcknowledged = true
	usageRecord.CustomerSignature = []byte("customer-sig")
	usageRecord.Metadata["region"] = "us-east"
	require.NoError(t, suite.keeper.SetUsageRecord(suite.ctx, *usageRecord))

	providerShare := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(700)))
	platformFee := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200)))
	validatorFee := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100)))
	settlement := types.NewSettlementRecord(
		settlementID,
		escrowID,
		orderID,
		leaseID,
		provider.String(),
		customer.String(),
		amount,
		providerShare,
		platformFee,
		validatorFee,
		[]string{usageID},
		usageRecord.UsageUnits,
		periodStart,
		periodEnd,
		types.SettlementTypeFinal,
		true,
		now,
		suite.ctx.BlockHeight(),
	)
	require.NoError(t, suite.keeper.SetSettlement(suite.ctx, *settlement))

	recipient := types.RewardRecipient{
		Address:     provider.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
		Reason:      "usage reward",
		UsageUnits:  usageRecord.UsageUnits,
		ReferenceID: orderID,
	}
	rewardDist := types.NewRewardDistribution(
		distributionID,
		1,
		types.RewardSourceUsage,
		[]types.RewardRecipient{recipient},
		now,
		suite.ctx.BlockHeight(),
	)
	rewardDist.ReferenceTxHashes = []string{"txhash"}
	rewardDist.Metadata["batch"] = "1"
	require.NoError(t, suite.keeper.SetRewardDistribution(suite.ctx, *rewardDist))

	claimable := types.NewClaimableRewards(provider.String(), now)
	claimable.AddReward(types.RewardEntry{
		DistributionID: distributionID,
		Source:         types.RewardSourceUsage,
		Amount:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50))),
		CreatedAt:      now,
		ExpiresAt:      ptrTime(now.Add(24 * time.Hour)),
		Reason:         "usage reward",
	})
	require.NoError(t, suite.keeper.SetClaimableRewards(suite.ctx, provider, *claimable))

	payout := types.NewPayoutRecord(
		payoutID,
		"invoice-1",
		settlementID,
		escrowID,
		orderID,
		leaseID,
		provider.String(),
		customer.String(),
		amount,
		platformFee,
		validatorFee,
		sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(0))),
		now,
		suite.ctx.BlockHeight(),
	)
	payout.FiatConversionID = conversionID
	require.NoError(t, suite.keeper.SetPayout(suite.ctx, *payout))

	conversion := types.FiatConversionRecord{
		ConversionID:      conversionID,
		InvoiceID:         "invoice-1",
		SettlementID:      settlementID,
		PayoutID:          payoutID,
		EscrowID:          escrowID,
		OrderID:           orderID,
		LeaseID:           leaseID,
		Provider:          provider.String(),
		Customer:          customer.String(),
		RequestedBy:       provider.String(),
		RequestedAt:       now,
		UpdatedAt:         now,
		State:             types.FiatConversionStateRequested,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: "USD", Denom: "uusd", Decimals: 6},
		CryptoAmount:      sdk.NewCoin("uve", sdkmath.NewInt(500)),
		StableAmount:      sdk.NewCoin("uusd", sdkmath.NewInt(1000)),
		FiatCurrency:      "USD",
		FiatAmount:        "100.00",
		PaymentMethod:     "bank",
		DestinationHash:   types.HashDestination("acct-123"),
		DestinationRegion: "US",
		SlippageTolerance: 0.01,
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key", "customer-key"}),
		AuditTrail: []types.FiatConversionAuditEntry{{
			Action:    "requested",
			Actor:     provider.String(),
			Timestamp: now.Unix(),
			Metadata:  map[string]string{"source": "test"},
		}},
	}
	require.NoError(t, suite.keeper.SetFiatConversion(suite.ctx, conversion))

	preference := types.FiatPayoutPreference{
		Provider:          provider.String(),
		Enabled:           true,
		FiatCurrency:      "USD",
		PaymentMethod:     "bank",
		DestinationHash:   types.HashDestination("acct-123"),
		DestinationRegion: "US",
		PreferredDEX:      "dex",
		PreferredOffRamp:  "offramp",
		SlippageTolerance: 0.02,
		CryptoToken:       types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken:       types.TokenSpec{Symbol: "USD", Denom: "uusd", Decimals: 6},
		EncryptedPayload:  makeEncryptedSettlementPayload(t, []string{"provider-key"}),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	require.NoError(t, suite.keeper.SetFiatPayoutPreference(suite.ctx, preference))

	data := seededData{
		escrow:     *escrow,
		settlement: *settlement,
		usage:      *usageRecord,
		rewardDist: *rewardDist,
		claimable:  *claimable,
		payout:     *payout,
		conversion: conversion,
		preference: preference,
		orderID:    orderID,
		provider:   provider,
		customer:   customer,
	}
	data.usagePeriod.start = periodStart
	data.usagePeriod.end = periodEnd
	return data
}

func TestGRPCSettlementQueries(t *testing.T) {
	suite := setupGRPCTest(t)
	data := seedSettlementData(t, suite)

	ctx := suite.ctx

	escrowResp, err := suite.queryClient.Escrow(ctx, &settlementv1.QueryEscrowRequest{EscrowId: data.escrow.EscrowID})
	require.NoError(t, err)
	require.NotNil(t, escrowResp.Escrow)
	require.Equal(t, data.escrow.EscrowID, escrowResp.Escrow.EscrowId)

	escrowsByOrderResp, err := suite.queryClient.EscrowsByOrder(ctx, &settlementv1.QueryEscrowsByOrderRequest{OrderId: data.orderID})
	require.NoError(t, err)
	require.Len(t, escrowsByOrderResp.Escrows, 1)
	require.Equal(t, data.escrow.EscrowID, escrowsByOrderResp.Escrows[0].EscrowId)

	escrowsByStateResp, err := suite.queryClient.EscrowsByState(ctx, &settlementv1.QueryEscrowsByStateRequest{State: string(data.escrow.State)})
	require.NoError(t, err)
	require.Len(t, escrowsByStateResp.Escrows, 1)
	require.Equal(t, data.escrow.EscrowID, escrowsByStateResp.Escrows[0].EscrowId)

	settlementResp, err := suite.queryClient.Settlement(ctx, &settlementv1.QuerySettlementRequest{SettlementId: data.settlement.SettlementID})
	require.NoError(t, err)
	require.NotNil(t, settlementResp.Settlement)
	require.Equal(t, data.settlement.SettlementID, settlementResp.Settlement.SettlementId)

	settlementsByOrderResp, err := suite.queryClient.SettlementsByOrder(ctx, &settlementv1.QuerySettlementsByOrderRequest{OrderId: data.orderID})
	require.NoError(t, err)
	require.Len(t, settlementsByOrderResp.Settlements, 1)
	require.Equal(t, data.settlement.SettlementID, settlementsByOrderResp.Settlements[0].SettlementId)

	usageResp, err := suite.queryClient.UsageRecord(ctx, &settlementv1.QueryUsageRecordRequest{UsageId: data.usage.UsageID})
	require.NoError(t, err)
	require.NotNil(t, usageResp.UsageRecord)
	require.Equal(t, data.usage.UsageID, usageResp.UsageRecord.UsageId)

	usageByOrderResp, err := suite.queryClient.UsageRecordsByOrder(ctx, &settlementv1.QueryUsageRecordsByOrderRequest{OrderId: data.orderID})
	require.NoError(t, err)
	require.Len(t, usageByOrderResp.UsageRecords, 1)
	require.Equal(t, data.usage.UsageID, usageByOrderResp.UsageRecords[0].UsageId)

	usageSummaryResp, err := suite.queryClient.UsageSummary(ctx, &settlementv1.QueryUsageSummaryRequest{
		OrderId:     data.orderID,
		Provider:    data.provider.String(),
		PeriodStart: data.usagePeriod.start.Unix(),
		PeriodEnd:   data.usagePeriod.end.Unix(),
	})
	require.NoError(t, err)
	require.Equal(t, data.usage.UsageUnits, usageSummaryResp.Summary.TotalUsageUnits)
	require.Len(t, usageSummaryResp.Summary.UsageRecordIds, 1)

	rewardDistResp, err := suite.queryClient.RewardDistribution(ctx, &settlementv1.QueryRewardDistributionRequest{DistributionId: data.rewardDist.DistributionID})
	require.NoError(t, err)
	require.NotNil(t, rewardDistResp.Distribution)
	require.Equal(t, data.rewardDist.DistributionID, rewardDistResp.Distribution.DistributionId)

	rewardsByEpochResp, err := suite.queryClient.RewardsByEpoch(ctx, &settlementv1.QueryRewardsByEpochRequest{EpochNumber: data.rewardDist.EpochNumber})
	require.NoError(t, err)
	require.Len(t, rewardsByEpochResp.Distributions, 1)
	require.Equal(t, data.rewardDist.DistributionID, rewardsByEpochResp.Distributions[0].DistributionId)

	rewardHistoryResp, err := suite.queryClient.RewardHistory(ctx, &settlementv1.QueryRewardHistoryRequest{Address: data.provider.String()})
	require.NoError(t, err)
	require.Len(t, rewardHistoryResp.Entries, 1)
	require.Equal(t, data.rewardDist.DistributionID, rewardHistoryResp.Entries[0].DistributionId)

	claimableResp, err := suite.queryClient.ClaimableRewards(ctx, &settlementv1.QueryClaimableRewardsRequest{Address: data.provider.String()})
	require.NoError(t, err)
	require.NotNil(t, claimableResp.Rewards)
	require.Equal(t, data.provider.String(), claimableResp.Rewards.Address)

	payoutResp, err := suite.queryClient.Payout(ctx, &settlementv1.QueryPayoutRequest{PayoutId: data.payout.PayoutID})
	require.NoError(t, err)
	require.NotNil(t, payoutResp.Payout)
	require.Equal(t, data.payout.PayoutID, payoutResp.Payout.PayoutId)

	payoutsByProviderResp, err := suite.queryClient.PayoutsByProvider(ctx, &settlementv1.QueryPayoutsByProviderRequest{Provider: data.provider.String()})
	require.NoError(t, err)
	require.Len(t, payoutsByProviderResp.Payouts, 1)
	require.Equal(t, data.payout.PayoutID, payoutsByProviderResp.Payouts[0].PayoutId)

	paramsResp, err := suite.queryClient.Params(ctx, &settlementv1.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, paramsResp.Params.PlatformFeeRate)

	conversionResp, err := suite.queryClient.FiatConversion(ctx, &settlementv1.QueryFiatConversionRequest{ConversionId: data.conversion.ConversionID})
	require.NoError(t, err)
	require.NotNil(t, conversionResp.Conversion)
	require.Equal(t, data.conversion.ConversionID, conversionResp.Conversion.ConversionId)

	conversionsByProviderResp, err := suite.queryClient.FiatConversionsByProvider(ctx, &settlementv1.QueryFiatConversionsByProviderRequest{Provider: data.provider.String()})
	require.NoError(t, err)
	require.Len(t, conversionsByProviderResp.Conversions, 1)
	require.Equal(t, data.conversion.ConversionID, conversionsByProviderResp.Conversions[0].ConversionId)

	preferenceResp, err := suite.queryClient.FiatPayoutPreference(ctx, &settlementv1.QueryFiatPayoutPreferenceRequest{Provider: data.provider.String()})
	require.NoError(t, err)
	require.NotNil(t, preferenceResp.Preference)
	require.Equal(t, data.preference.Provider, preferenceResp.Preference.Provider)
}
