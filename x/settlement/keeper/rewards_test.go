package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func (s *KeeperTestSuite) seedStakeRewardRoutes(t *testing.T) {
	t.Helper()

	params := s.delegationKeeper.GetParams(s.ctx)
	params.ValidatorCommissionRate = 1000
	require.NoError(t, s.delegationKeeper.SetParams(s.ctx, params))

	validatorOneShares := delegationtypes.NewValidatorShares(s.validator.String(), s.ctx.BlockTime())
	validatorOneShares.TotalStake = "700"
	validatorOneShares.TotalShares = "700"
	require.NoError(t, s.delegationKeeper.SetValidatorShares(s.ctx, *validatorOneShares))

	validatorTwoShares := delegationtypes.NewValidatorShares(s.validatorTwo.String(), s.ctx.BlockTime())
	validatorTwoShares.TotalStake = "300"
	validatorTwoShares.TotalShares = "300"
	require.NoError(t, s.delegationKeeper.SetValidatorShares(s.ctx, *validatorTwoShares))

	require.NoError(t, s.delegationKeeper.SetDelegation(s.ctx, *delegationtypes.NewDelegation(
		s.validator.String(),
		s.validator.String(),
		"400",
		"400",
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)))
	require.NoError(t, s.delegationKeeper.SetDelegation(s.ctx, *delegationtypes.NewDelegation(
		s.delegatorOne.String(),
		s.validator.String(),
		"300",
		"300",
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)))
	require.NoError(t, s.delegationKeeper.SetDelegation(s.ctx, *delegationtypes.NewDelegation(
		s.delegatorTwo.String(),
		s.validatorTwo.String(),
		"300",
		"300",
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)))
}

func rewardTotalsByAddress(recipients []types.RewardRecipient, denom string) map[string]sdkmath.Int {
	totals := make(map[string]sdkmath.Int, len(recipients))
	for _, recipient := range recipients {
		total, found := totals[recipient.Address]
		if !found {
			total = sdkmath.ZeroInt()
		}
		totals[recipient.Address] = total.Add(recipient.Amount.AmountOf(denom))
	}
	return totals
}

func (s *KeeperTestSuite) TestDistributeStakingRewards() {
	t := s.T()

	s.seedStakeRewardRoutes(t)

	params := s.keeper.GetParams(s.ctx)
	params.StakingRewardEpochLength = 100
	params.PayoutHoldbackRate = "0.10"
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	firstSettlement := s.buildSettlement(t, "staking-epoch-one")
	firstSettlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000)))
	firstSettlement.PlatformFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50)))
	firstSettlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100)))
	firstSettlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(850)))
	require.NoError(t, s.keeper.SetSettlement(s.ctx, firstSettlement))

	firstPayout, err := s.keeper.ExecutePayout(s.ctx, "inv-staking-epoch-one", firstSettlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, firstPayout.State)

	firstDist, err := s.keeper.DistributeStakingRewards(s.ctx, 1)
	require.NoError(t, err)
	require.Equal(t, types.RewardSourceStaking, firstDist.Source)
	require.Equal(t, sdkmath.NewInt(100), firstDist.TotalRewards.AmountOf("uve"))

	replayed, err := s.keeper.DistributeStakingRewards(s.ctx, 1)
	require.NoError(t, err)
	require.Equal(t, firstDist.DistributionID, replayed.DistributionID)

	restarted := keeper.NewKeeper(s.cdc, s.keeper.StoreKey(), s.bankKeeper, s.escrow, "authority", mockEncryptionKeeper{})
	restarted.SetStakeRoutingKeeper(s.delegationKeeper)

	secondSettlement := s.buildSettlement(t, "staking-epoch-two")
	secondSettlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(400)))
	secondSettlement.PlatformFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(20)))
	secondSettlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(40)))
	secondSettlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(340)))
	require.NoError(t, restarted.SetSettlement(s.ctx, secondSettlement))

	secondPayout, err := restarted.ExecutePayout(s.ctx, "inv-staking-epoch-two", secondSettlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, secondPayout.State)

	secondDist, err := restarted.DistributeStakingRewards(s.ctx, 2)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(40), secondDist.TotalRewards.AmountOf("uve"))

	validatorRewards, found := restarted.GetClaimableRewards(s.ctx, s.validator)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(60), validatorRewards.TotalClaimable.AmountOf("uve"))

	delegatorOneRewards, found := restarted.GetClaimableRewards(s.ctx, s.delegatorOne)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(38), delegatorOneRewards.TotalClaimable.AmountOf("uve"))

	validatorTwoRewards, found := restarted.GetClaimableRewards(s.ctx, s.validatorTwo)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(4), validatorTwoRewards.TotalClaimable.AmountOf("uve"))

	delegatorTwoRewards, found := restarted.GetClaimableRewards(s.ctx, s.delegatorTwo)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(38), delegatorTwoRewards.TotalClaimable.AmountOf("uve"))
}

func (s *KeeperTestSuite) TestDistributeStakingRewardsRequiresStakeRoutingKeeper() {
	t := s.T()

	params := s.keeper.GetParams(s.ctx)
	params.StakingRewardEpochLength = 100
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	standalone := keeper.NewKeeper(s.cdc, s.storeKey, s.bankKeeper, s.escrow, "authority", mockEncryptionKeeper{})
	require.NoError(t, standalone.SetParams(s.ctx, params))

	settlement := s.buildSettlement(t, "staking-missing-router")
	settlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500)))
	settlement.PlatformFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(25)))
	settlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(50)))
	settlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(425)))
	require.NoError(t, standalone.SetSettlement(s.ctx, settlement))

	payout, err := standalone.ExecutePayout(s.ctx, "inv-staking-missing-router", settlement.SettlementID)
	require.NoError(t, err)
	require.Equal(t, types.PayoutStateCompleted, payout.State)

	dist, err := standalone.DistributeStakingRewards(s.ctx, 1)
	require.Nil(t, dist)
	require.ErrorContains(t, err, "stake routing keeper not configured")
}

func (s *KeeperTestSuite) TestDistributeProviderRewards() {
	// Create usage records for reward distribution
	usages := []types.UsageRecord{
		{
			UsageID:     "usage-reward-1",
			OrderID:     "order-1",
			Provider:    s.provider.String(),
			Customer:    s.depositor.String(),
			UsageUnits:  1000,
			UsageType:   "compute",
			TotalCost:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			PeriodStart: s.ctx.BlockTime().Add(-time.Hour),
			PeriodEnd:   s.ctx.BlockTime(),
			SubmittedAt: s.ctx.BlockTime(),
		},
	}

	// Distribute provider rewards
	dist, err := s.keeper.DistributeProviderRewards(s.ctx, usages)
	s.Require().NoError(err)
	s.Require().NotNil(dist)
	s.Require().Equal(types.RewardSourceProvider, dist.Source)
}

func (s *KeeperTestSuite) TestDistributeUsageRewards() {
	now := s.ctx.BlockTime()
	usages := []types.UsageRecord{
		{
			UsageID:              "usage-reward-usage-1",
			OrderID:              "order-usage-1",
			Provider:             s.provider.String(),
			Customer:             s.depositor.String(),
			UsageUnits:           1000,
			UsageType:            "cpu",
			TotalCost:            sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			PeriodStart:          now.Add(-time.Hour),
			PeriodEnd:            now,
			SubmittedAt:          now,
			CustomerAcknowledged: true,
			ProviderSignature:    []byte("provider-signature"),
		},
	}

	dist, err := s.keeper.DistributeUsageRewards(s.ctx, usages)
	s.Require().NoError(err)
	s.Require().NotNil(dist)
	s.Require().Equal(types.RewardSourceUsage, dist.Source)
	s.Require().Equal(sdkmath.NewInt(100), dist.TotalRewards.AmountOf("uve"))
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
	s.Require().False(rewards.TotalClaimable.IsZero())

	// Claim rewards
	claimed, err := s.keeper.ClaimRewards(s.ctx, s.depositor, "")
	s.Require().NoError(err)
	s.Require().False(claimed.IsZero())

	// Verify rewards are claimed
	rewards, found = s.keeper.GetClaimableRewards(s.ctx, s.depositor)
	s.Require().True(found)
	s.Require().True(rewards.TotalClaimable.IsZero())
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
	s.Require().Equal(sdkmath.NewInt(300), rewards.TotalClaimable.AmountOf("uve"))
}

func (s *KeeperTestSuite) TestGetRewardsByEpoch() {
	t := s.T()

	s.seedStakeRewardRoutes(t)

	params := s.keeper.GetParams(s.ctx)
	params.StakingRewardEpochLength = 100
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	firstSettlement := s.buildSettlement(t, "rewards-epoch-one")
	firstSettlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100)))
	firstSettlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))
	firstSettlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(90)))
	require.NoError(t, s.keeper.SetSettlement(s.ctx, firstSettlement))
	_, err := s.keeper.ExecutePayout(s.ctx, "inv-rewards-epoch-one", firstSettlement.SettlementID)
	require.NoError(t, err)

	_, err = s.keeper.DistributeStakingRewards(s.ctx, 1)
	require.NoError(t, err)

	secondSettlement := s.buildSettlement(t, "rewards-epoch-two")
	secondSettlement.TotalAmount = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(200)))
	secondSettlement.ValidatorFee = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(20)))
	secondSettlement.ProviderShare = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(180)))
	require.NoError(t, s.keeper.SetSettlement(s.ctx, secondSettlement))
	_, err = s.keeper.ExecutePayout(s.ctx, "inv-rewards-epoch-two", secondSettlement.SettlementID)
	require.NoError(t, err)

	_, err = s.keeper.DistributeStakingRewards(s.ctx, 2)
	require.NoError(t, err)

	// Get rewards by epoch
	epoch1Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 1)
	s.Require().Len(epoch1Rewards, 1)

	epoch2Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 2)
	s.Require().Len(epoch2Rewards, 1)

	// Non-existent epoch
	epoch3Rewards := s.keeper.GetRewardsByEpoch(s.ctx, 3)
	s.Require().Len(epoch3Rewards, 0)
}

func (s *KeeperTestSuite) TestGetRewardHistory() {
	recipients := []types.RewardRecipient{
		{
			Address: s.provider.String(),
			Amount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(123))),
			Reason:  "usage_cpu_reward",
		},
	}

	dist := types.NewRewardDistribution(
		"dist-usage-1",
		1,
		types.RewardSourceUsage,
		recipients,
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)

	err := s.keeper.SetRewardDistribution(s.ctx, *dist)
	s.Require().NoError(err)

	entries, err := s.keeper.GetRewardHistory(s.ctx, s.provider.String(), "", 0, 0)
	s.Require().NoError(err)
	s.Require().Len(entries, 1)
	s.Require().Equal("dist-usage-1", entries[0].DistributionID)
}

func TestRewardDistributionValidation(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_recipient______")).String()

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
				TotalRewards:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				Recipients: []types.RewardRecipient{
					{
						Address: validAddr,
						Amount:  sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
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
				TotalRewards:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			},
			expectError: true,
		},
		{
			name: "invalid source",
			dist: types.RewardDistribution{
				DistributionID: "dist-1",
				Source:         types.RewardSource("invalid"),
				TotalRewards:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
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

func TestClaimableRewardsStructure(t *testing.T) {
	testCases := []struct {
		name    string
		rewards types.ClaimableRewards
		valid   bool
	}{
		{
			name: "valid claimable rewards",
			rewards: types.ClaimableRewards{
				Address:        "cosmos1address...",
				TotalClaimable: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				TotalClaimed:   sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
				LastUpdated:    time.Now(),
			},
			valid: true,
		},
		{
			name: "empty address",
			rewards: types.ClaimableRewards{
				Address:        "",
				TotalClaimable: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify the struct can be created - no Validate method exists
			if tc.valid {
				require.NotEmpty(t, tc.rewards.Address)
				require.True(t, tc.rewards.TotalClaimable.IsValid())
			} else {
				require.Empty(t, tc.rewards.Address)
			}
		})
	}
}
