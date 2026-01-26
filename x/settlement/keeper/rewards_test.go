package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/settlement/types"
)

func (s *KeeperTestSuite) TestDistributeStakingRewards() {
	// Set up params with staking reward pool
	params := types.DefaultParams()
	params.StakingRewardPool = "1000000" // 1M uve
	params.StakingRewardEpochLength = 100
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Distribute staking rewards
	epoch := uint64(1)
	dist, err := s.keeper.DistributeStakingRewards(s.ctx, epoch)
	s.Require().NoError(err)
	s.Require().NotNil(dist)
	s.Require().Equal(types.RewardSourceStaking, dist.Source)
	s.Require().Equal(epoch, dist.EpochNumber)
}

func (s *KeeperTestSuite) TestDistributeProviderRewards() {
	// Create usage records for reward distribution
	usages := []types.UsageRecord{
		{
			UsageID:     "usage-reward-1",
			OrderID:     "order-1",
			Provider:    s.provider.String(),
			Customer:    s.depositor.String(),
			ComputeUsed: "1000",
			TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			PeriodStart: s.ctx.BlockTime().Add(-time.Hour),
			PeriodEnd:   s.ctx.BlockTime(),
			RecordedAt:  s.ctx.BlockTime(),
		},
	}

	// Distribute provider rewards
	dist, err := s.keeper.DistributeProviderRewards(s.ctx, usages)
	s.Require().NoError(err)
	s.Require().NotNil(dist)
	s.Require().Equal(types.RewardSourceProvider, dist.Source)
}

func (s *KeeperTestSuite) TestDistributeVerificationRewards() {
	// Create verification results
	results := []keeper.VerificationResult{
		{
			ValidatorAddress: s.validator.String(),
			AccountAddress:   s.depositor.String(),
			Score:            100,
			BlockHeight:      1,
		},
	}

	// Distribute verification rewards
	dist, err := s.keeper.DistributeVerificationRewards(s.ctx, results)
	s.Require().NoError(err)
	s.Require().NotNil(dist)
	s.Require().Equal(types.RewardSourceVerification, dist.Source)
}

func (s *KeeperTestSuite) TestClaimRewards() {
	// Add claimable rewards
	entry := types.RewardEntry{
		DistributionID: "dist-1",
		Source:         types.RewardSourceStaking,
		Amount:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		CreatedAt:      s.ctx.BlockTime(),
		Reason:         "staking reward",
	}
	err := s.keeper.AddClaimableReward(s.ctx, s.depositor, entry)
	s.Require().NoError(err)

	// Verify claimable rewards exist
	rewards, found := s.keeper.GetClaimableRewards(s.ctx, s.depositor)
	s.Require().True(found)
	s.Require().False(rewards.TotalUnclaimed.IsZero())

	// Claim rewards
	claimed, err := s.keeper.ClaimRewards(s.ctx, s.depositor, "")
	s.Require().NoError(err)
	s.Require().False(claimed.IsZero())

	// Verify rewards are claimed
	rewards, found = s.keeper.GetClaimableRewards(s.ctx, s.depositor)
	s.Require().True(found)
	s.Require().True(rewards.TotalUnclaimed.IsZero())
}

func (s *KeeperTestSuite) TestClaimRewardsBySource() {
	// Add multiple reward entries from different sources
	stakingEntry := types.RewardEntry{
		DistributionID: "dist-staking",
		Source:         types.RewardSourceStaking,
		Amount:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		CreatedAt:      s.ctx.BlockTime(),
		Reason:         "staking reward",
	}
	providerEntry := types.RewardEntry{
		DistributionID: "dist-provider",
		Source:         types.RewardSourceProvider,
		Amount:         sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(300))),
		CreatedAt:      s.ctx.BlockTime(),
		Reason:         "provider reward",
	}

	err := s.keeper.AddClaimableReward(s.ctx, s.depositor, stakingEntry)
	s.Require().NoError(err)
	err = s.keeper.AddClaimableReward(s.ctx, s.depositor, providerEntry)
	s.Require().NoError(err)

	// Claim only staking rewards
	claimed, err := s.keeper.ClaimRewards(s.ctx, s.depositor, string(types.RewardSourceStaking))
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(500), claimed.AmountOf("uve"))

	// Verify provider rewards still exist
	rewards, found := s.keeper.GetClaimableRewards(s.ctx, s.depositor)
	s.Require().True(found)
	s.Require().Equal(sdkmath.NewInt(300), rewards.TotalUnclaimed.AmountOf("uve"))
}

func (s *KeeperTestSuite) TestGetRewardsByEpoch() {
	// Distribute rewards for multiple epochs
	params := types.DefaultParams()
	params.StakingRewardPool = "1000000"
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	_, err = s.keeper.DistributeStakingRewards(s.ctx, 1)
	s.Require().NoError(err)
	_, err = s.keeper.DistributeStakingRewards(s.ctx, 2)
	s.Require().NoError(err)

	// Get rewards by epoch
	epoch1Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 1)
	s.Require().Len(epoch1Rewards, 1)

	epoch2Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 2)
	s.Require().Len(epoch2Rewards, 1)

	// Non-existent epoch
	epoch3Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 3)
	s.Require().Len(epoch3Rewards, 0)
}

func TestRewardDistributionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		dist        types.RewardDistribution
		expectError bool
	}{
		{
			name: "valid reward distribution",
			dist: types.RewardDistribution{
				DistributionID: "dist-1",
				Source:         types.RewardSourceStaking,
				EpochNumber:    1,
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Recipients: []types.RewardRecipient{
					{
						Address: "cosmos1recipient...",
						Amount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
						Share:   "1.0",
						Reason:  "staking",
					},
				},
				DistributedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty distribution ID",
			dist: types.RewardDistribution{
				DistributionID: "",
				Source:         types.RewardSourceStaking,
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			},
			expectError: true,
		},
		{
			name: "invalid source",
			dist: types.RewardDistribution{
				DistributionID: "dist-1",
				Source:         types.RewardSource("invalid"),
				TotalAmount:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dist.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClaimableRewardsValidation(t *testing.T) {
	testCases := []struct {
		name        string
		rewards     types.ClaimableRewards
		expectError bool
	}{
		{
			name: "valid claimable rewards",
			rewards: types.ClaimableRewards{
				Address:        "cosmos1address...",
				TotalUnclaimed: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				TotalClaimed:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
				LastUpdated:    time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty address",
			rewards: types.ClaimableRewards{
				Address:        "",
				TotalUnclaimed: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rewards.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
