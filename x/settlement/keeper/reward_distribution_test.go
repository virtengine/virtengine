package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func (s *KeeperTestSuite) TestProviderRewardCalculationAccuracy() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.UsageRewardRateBps = 1000
	params.UsageRewardCPUMultiplierBps = 10000
	params.UsageRewardSLAOnTimeMultiplierBps = 10000
	params.UsageRewardSLALateMultiplierBps = 10000
	params.UsageRewardAcknowledgedMultiplierBps = 10000
	params.UsageRewardUnacknowledgedMultiplierBps = 10000
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	now := s.ctx.BlockTime()
	usages := []types.UsageRecord{
		{
			UsageID:              "usage-provider-1",
			OrderID:              "order-1",
			Provider:             s.provider.String(),
			Customer:             s.depositor.String(),
			UsageUnits:           250,
			UsageType:            "compute",
			TotalCost:            sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			PeriodStart:          now.Add(-time.Hour),
			PeriodEnd:            now,
			SubmittedAt:          now,
			CustomerAcknowledged: true,
		},
		{
			UsageID:              "usage-provider-2",
			OrderID:              "order-2",
			Provider:             s.depositor.String(),
			Customer:             s.provider.String(),
			UsageUnits:           500,
			UsageType:            "compute",
			TotalCost:            sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(2000))),
			PeriodStart:          now.Add(-time.Hour),
			PeriodEnd:            now,
			SubmittedAt:          now,
			CustomerAcknowledged: true,
		},
	}

	dist, err := s.keeper.DistributeProviderRewards(s.ctx, usages)
	require.NoError(t, err)
	require.Equal(t, types.RewardSourceProvider, dist.Source)

	var providerReward sdkmath.Int
	var depositorReward sdkmath.Int
	for _, recipient := range dist.Recipients {
		if recipient.Address == s.provider.String() {
			providerReward = recipient.Amount.AmountOf("uve")
		}
		if recipient.Address == s.depositor.String() {
			depositorReward = recipient.Amount.AmountOf("uve")
		}
	}

	require.Equal(t, sdkmath.NewInt(100), providerReward)
	require.Equal(t, sdkmath.NewInt(200), depositorReward)
}

func (s *KeeperTestSuite) TestUsageRewardsMultiDenomAndEvents() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.UsageRewardRateBps = 1000
	params.UsageRewardCPUMultiplierBps = 10000
	params.UsageRewardMemoryMultiplierBps = 10000
	params.UsageRewardStorageMultiplierBps = 10000
	params.UsageRewardGPUMultiplierBps = 10000
	params.UsageRewardNetworkMultiplierBps = 10000
	params.UsageRewardSLAOnTimeMultiplierBps = 10000
	params.UsageRewardSLALateMultiplierBps = 10000
	params.UsageRewardAcknowledgedMultiplierBps = 10000
	params.UsageRewardUnacknowledgedMultiplierBps = 10000
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	usages := []types.UsageRecord{
		{
			UsageID:              "usage-multi-1",
			OrderID:              "order-multi",
			Provider:             s.provider.String(),
			Customer:             s.depositor.String(),
			UsageUnits:           1000,
			UsageType:            "cpu",
			TotalCost:            sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)), sdk.NewCoin("uusdc", sdkmath.NewInt(500))),
			PeriodStart:          s.ctx.BlockTime().Add(-time.Hour),
			PeriodEnd:            s.ctx.BlockTime(),
			SubmittedAt:          s.ctx.BlockTime(),
			CustomerAcknowledged: true,
		},
	}

	dist, err := s.keeper.DistributeUsageRewards(s.ctx, usages)
	require.NoError(t, err)
	require.Equal(t, types.RewardSourceUsage, dist.Source)
	require.Equal(t, sdkmath.NewInt(100), dist.TotalRewards.AmountOf("uve"))
	require.Equal(t, sdkmath.NewInt(50), dist.TotalRewards.AmountOf("uusdc"))

	events := s.ctx.EventManager().Events()
	require.NotEmpty(t, events)
}

func (s *KeeperTestSuite) TestStakingRewardAllocation() {
	t := s.T()

	epoch := uint64(5)
	dist, err := s.keeper.DistributeStakingRewards(s.ctx, epoch)
	require.NoError(t, err)
	require.Equal(t, types.RewardSourceStaking, dist.Source)
	require.Equal(t, epoch, dist.EpochNumber)
	require.Len(t, dist.Recipients, 1)
	require.NotEmpty(t, dist.Recipients[0].Address)
}
