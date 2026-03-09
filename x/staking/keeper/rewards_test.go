// Package keeper implements the staking module keeper tests.
//
// VE-921: Reward calculation tests
package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/staking/types"
)

// RewardsTestSuite is the test suite for reward calculations
type RewardsTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      Keeper
	cdc         codec.BinaryCodec
	skey        *storetypes.KVStoreKey
	stakeKeeper *mockStakeKeeper
}

// SetupTest sets up the test suite
func (s *RewardsTestSuite) SetupTest() {
	s.skey = storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(s.skey, storetypes.StoreTypeIAVL, db)
	require.NoError(s.T(), stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(registry)

	s.stakeKeeper = newMockStakeKeeper()
	s.keeper = NewKeeper(
		s.cdc,
		s.skey,
		nil, // bankKeeper
		nil, // veidKeeper
		s.stakeKeeper,
		"authority",
	)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Initialize default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	require.NoError(s.T(), err)
}

// TestRewardsTestSuite runs the test suite
func TestRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(RewardsTestSuite))
}

// TestCalculateRewards tests reward calculation
func (s *RewardsTestSuite) TestCalculateRewards() {
	// Create a performance record
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.VEIDVerificationsCompleted = 5
	perf.VEIDVerificationsExpected = 5
	perf.VEIDVerificationScore = 10000
	perf.UptimeSeconds = 86400
	perf.DowntimeSeconds = 0
	types.ComputeOverallScore(perf)

	input := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perf,
		StakeAmount:      1000000000,
		TotalStake:       1000000000,
		EpochRewardPool:  100000000,
		BlocksInEpoch:    100,
	}

	reward := types.CalculateRewards(input, "uve")

	s.Require().NotNil(reward)
	s.Require().False(reward.TotalReward.IsZero())
	s.Require().False(reward.BlockProposalReward.IsZero())
	s.Require().False(reward.VEIDReward.IsZero())
	s.Require().False(reward.UptimeReward.IsZero())
}

// TestCalculateRewardsZeroStake tests reward calculation with zero stake
func (s *RewardsTestSuite) TestCalculateRewardsZeroStake() {
	input := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      nil,
		StakeAmount:      0,
		TotalStake:       0,
		EpochRewardPool:  100000000,
		BlocksInEpoch:    100,
	}

	reward := types.CalculateRewards(input, "uve")

	s.Require().NotNil(reward)
	s.Require().True(reward.TotalReward.IsZero())
}

// TestCalculateRewardsProportional tests proportional reward distribution
func (s *RewardsTestSuite) TestCalculateRewardsProportional() {
	// Create two validators with different stakes
	perf1 := types.NewValidatorPerformance("validator1", 1)
	perf1.BlocksProposed = 10
	perf1.BlocksExpected = 10
	perf1.UptimeSeconds = 86400
	types.ComputeOverallScore(perf1)

	perf2 := types.NewValidatorPerformance("validator2", 1)
	perf2.BlocksProposed = 10
	perf2.BlocksExpected = 10
	perf2.UptimeSeconds = 86400
	types.ComputeOverallScore(perf2)

	totalStake := int64(1000000000)
	epochRewardPool := int64(100000000)

	// Validator 1 has 70% stake
	input1 := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perf1,
		StakeAmount:      700000000,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	// Validator 2 has 30% stake
	input2 := types.RewardCalculationInput{
		ValidatorAddress: "validator2",
		Performance:      perf2,
		StakeAmount:      300000000,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	reward1 := types.CalculateRewards(input1, "uve")
	reward2 := types.CalculateRewards(input2, "uve")

	// Validator 1 should get more rewards due to higher stake
	total1 := reward1.TotalReward.AmountOf("uve").Int64()
	total2 := reward2.TotalReward.AmountOf("uve").Int64()

	s.Require().Greater(total1, total2)

	// Ratio should be approximately 70:30
	ratio := float64(total1) / float64(total2)
	s.Require().InDelta(2.33, ratio, 0.5) // 70/30 ≈ 2.33
}

// TestCalculateRewardsPerformanceMultiplier tests performance impact on rewards
func (s *RewardsTestSuite) TestCalculateRewardsPerformanceMultiplier() {
	// High performer
	perfHigh := types.NewValidatorPerformance("validator1", 1)
	perfHigh.BlocksProposed = 10
	perfHigh.BlocksExpected = 10
	perfHigh.VEIDVerificationsCompleted = 5
	perfHigh.VEIDVerificationsExpected = 5
	perfHigh.VEIDVerificationScore = 10000
	perfHigh.UptimeSeconds = 86400
	types.ComputeOverallScore(perfHigh)

	// Low performer
	perfLow := types.NewValidatorPerformance("validator2", 1)
	perfLow.BlocksProposed = 2
	perfLow.BlocksExpected = 10
	perfLow.VEIDVerificationsCompleted = 1
	perfLow.VEIDVerificationsExpected = 5
	perfLow.VEIDVerificationScore = 5000
	perfLow.UptimeSeconds = 43200
	perfLow.DowntimeSeconds = 43200
	types.ComputeOverallScore(perfLow)

	stake := int64(500000000)
	totalStake := int64(1000000000)
	epochRewardPool := int64(100000000)

	inputHigh := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perfHigh,
		StakeAmount:      stake,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	inputLow := types.RewardCalculationInput{
		ValidatorAddress: "validator2",
		Performance:      perfLow,
		StakeAmount:      stake,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	rewardHigh := types.CalculateRewards(inputHigh, "uve")
	rewardLow := types.CalculateRewards(inputLow, "uve")

	// High performer should get more rewards despite equal stake
	totalHigh := rewardHigh.TotalReward.AmountOf("uve").Int64()
	totalLow := rewardLow.TotalReward.AmountOf("uve").Int64()

	s.Require().Greater(totalHigh, totalLow)
}

// TestCalculateIdentityNetworkReward tests identity network reward calculation
func (s *RewardsTestSuite) TestCalculateIdentityNetworkReward() {
	input := types.IdentityNetworkRewardInput{
		ValidatorAddress:         "validator1",
		VerificationsCompleted:   10,
		TotalVerifications:       100,
		AverageVerificationScore: 9000,
		RewardPool:               100000000,
	}

	reward := types.CalculateIdentityNetworkReward(input, "uve")

	s.Require().False(reward.IsZero())

	// 10% of verifications should yield ~10% of pool (plus quality bonus)
	amount := reward.AmountOf("uve").Int64()
	s.Require().Equal(int64(9000000), amount) // 10% share discounted by 90% quality
}

// TestCalculateIdentityNetworkRewardZero tests zero verification case
func (s *RewardsTestSuite) TestCalculateIdentityNetworkRewardZero() {
	input := types.IdentityNetworkRewardInput{
		ValidatorAddress:         "validator1",
		VerificationsCompleted:   0,
		TotalVerifications:       100,
		AverageVerificationScore: 9000,
		RewardPool:               100000000,
	}

	reward := types.CalculateIdentityNetworkReward(input, "uve")

	s.Require().True(reward.IsZero())
}

// TestCalculateEpochRewards tests epoch reward calculation
func (s *RewardsTestSuite) TestCalculateEpochRewards() {
	epoch := uint64(1)

	// Create epoch
	epochInfo := types.NewRewardEpoch(epoch, 1, s.ctx.BlockTime())
	epochInfo.EndHeight = 100
	err := s.keeper.SetRewardEpoch(s.ctx, *epochInfo)
	s.Require().NoError(err)

	// Create validator performances
	for i := 1; i <= 3; i++ {
		validatorAddr := sdk.AccAddress([]byte(fmt.Sprintf("validator-%d-address", i))).String()
		s.stakeKeeper.SetStake(validatorAddr, int64(i)*1_000_000)
		perf := types.NewValidatorPerformance(validatorAddr, epoch)
		perf.BlocksProposed = int64(i * 10)
		perf.BlocksExpected = 30
		perf.VEIDVerificationsCompleted = int64(i * 2)
		perf.VEIDVerificationsExpected = 6
		perf.VEIDVerificationScore = 8000
		perf.UptimeSeconds = 86400
		types.ComputeOverallScore(perf)
		err := s.keeper.SetValidatorPerformance(s.ctx, *perf)
		s.Require().NoError(err)
	}

	// Calculate rewards
	rewards, err := s.keeper.CalculateEpochRewards(s.ctx, epoch)
	s.Require().NoError(err)
	s.Require().Len(rewards, 3)

	// All rewards should be non-zero
	for _, reward := range rewards {
		s.Require().False(reward.TotalReward.IsZero())
	}

	totalDistributed := int64(0)
	for _, reward := range rewards {
		totalDistributed += reward.TotalReward.AmountOf("uve").Int64()
	}

	epochRewardPool := s.keeper.GetParams(s.ctx).BaseRewardPerBlock * types.EpochDuration(epochInfo)
	s.Require().Greater(totalDistributed, int64(0))
	s.Require().Less(totalDistributed, epochRewardPool)
}

func (s *RewardsTestSuite) TestDistributeRewardsRoutesDelegatorShare() {
	epoch := uint64(1)
	validatorAddr := sdk.AccAddress([]byte("validator-route")).String()

	bankKeeper := &mockBankKeeper{}
	stakeKeeper := newMockStakeKeeper()
	stakeKeeper.SetStake(validatorAddr, 1_000_000)
	stakeKeeper.SetValidatorPayoutBPS(validatorAddr, 8000)

	s.keeper = NewKeeper(
		s.cdc,
		s.skey,
		bankKeeper,
		nil,
		stakeKeeper,
		"authority",
	)
	require.NoError(s.T(), s.keeper.SetParams(s.ctx, types.DefaultParams()))

	epochInfo := types.NewRewardEpoch(epoch, 1, s.ctx.BlockTime())
	epochInfo.EndHeight = 100
	require.NoError(s.T(), s.keeper.SetRewardEpoch(s.ctx, *epochInfo))

	perf := types.NewValidatorPerformance(validatorAddr, epoch)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.UptimeSeconds = 3600
	perf.VEIDVerificationsCompleted = 2
	perf.VEIDVerificationsExpected = 2
	perf.VEIDVerificationScore = 9000
	types.ComputeOverallScore(perf)
	require.NoError(s.T(), s.keeper.SetValidatorPerformance(s.ctx, *perf))

	require.NoError(s.T(), s.keeper.DistributeRewards(s.ctx, epoch))
	require.Len(s.T(), bankKeeper.minted, 1)
	require.Len(s.T(), bankKeeper.moduleTransfers, 1)
	require.Len(s.T(), bankKeeper.accountTransfers, 1)

	totalMinted := bankKeeper.minted[0].AmountOf("uve")
	delegatorAmount := bankKeeper.moduleTransfers[0].Amount.AmountOf("uve")
	validatorAmount := bankKeeper.accountTransfers[0].Amount.AmountOf("uve")
	require.True(s.T(), totalMinted.Equal(delegatorAmount.Add(validatorAmount)))
	require.Equal(s.T(), "virt_delegation", bankKeeper.moduleTransfers[0].RecipientModule)
	require.Len(s.T(), stakeKeeper.rewardDistCalls, 1)
}

// TestCalculateVEIDRewards tests VEID reward calculation
func (s *RewardsTestSuite) TestCalculateVEIDRewards() {
	epoch := uint64(1)

	// Create epoch
	epochInfo := types.NewRewardEpoch(epoch, 1, s.ctx.BlockTime())
	epochInfo.EndHeight = 100
	err := s.keeper.SetRewardEpoch(s.ctx, *epochInfo)
	s.Require().NoError(err)

	// Create validator performances with VEID work
	for i := 1; i <= 2; i++ {
		validatorAddr := "validator" + string(rune('0'+i))
		perf := types.NewValidatorPerformance(validatorAddr, epoch)
		perf.VEIDVerificationsCompleted = int64(i * 5)
		perf.VEIDVerificationScore = int64(8000 + i*500)
		err := s.keeper.SetValidatorPerformance(s.ctx, *perf)
		s.Require().NoError(err)
	}

	// Calculate VEID rewards
	rewards, err := s.keeper.CalculateVEIDRewards(s.ctx, epoch)
	s.Require().NoError(err)
	s.Require().Len(rewards, 2)

	// Validator with more verifications should get more rewards
	s.Require().Greater(
		rewards[1].VEIDReward.AmountOf("uve").Int64(),
		rewards[0].VEIDReward.AmountOf("uve").Int64(),
	)
}

// TestRewardDeterminism tests that reward calculations are deterministic
func (s *RewardsTestSuite) TestRewardDeterminism() {
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.VEIDVerificationsCompleted = 5
	perf.VEIDVerificationsExpected = 5
	perf.VEIDVerificationScore = 9500
	perf.UptimeSeconds = 86400
	types.ComputeOverallScore(perf)

	input := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perf,
		StakeAmount:      500000000,
		TotalStake:       1000000000,
		EpochRewardPool:  100000000,
		BlocksInEpoch:    100,
	}

	// Calculate rewards multiple times
	reward1 := types.CalculateRewards(input, "uve")
	reward2 := types.CalculateRewards(input, "uve")
	reward3 := types.CalculateRewards(input, "uve")

	// All results should be identical (deterministic)
	s.Require().True(reward1.TotalReward.Equal(reward2.TotalReward))
	s.Require().True(reward2.TotalReward.Equal(reward3.TotalReward))
	s.Require().Equal(reward1.PerformanceScore, reward2.PerformanceScore)
	s.Require().Equal(reward2.PerformanceScore, reward3.PerformanceScore)
}

func (s *RewardsTestSuite) TestCalculateRewardsCapsSingleValidatorToEpochPool() {
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 100
	perf.BlocksExpected = 100
	perf.VEIDVerificationsCompleted = 10
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = types.MaxPerformanceScore
	perf.UptimeSeconds = 3600
	types.ComputeOverallScore(perf)

	input := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perf,
		StakeAmount:      1_000_000,
		TotalStake:       1_000_000,
		EpochRewardPool:  100_000_000,
	}

	reward := types.CalculateRewards(input, "uve")
	s.Require().Equal(int64(30_000_000), reward.BlockProposalReward.AmountOf("uve").Int64())
	s.Require().Equal(int64(40_000_000), reward.VEIDReward.AmountOf("uve").Int64())
	s.Require().Equal(int64(30_000_000), reward.UptimeReward.AmountOf("uve").Int64())
	s.Require().Equal(int64(100_000_000), reward.TotalReward.AmountOf("uve").Int64())
}

func (s *RewardsTestSuite) TestCalculateRewardsPenalizesLowPerformanceWithoutOverpaying() {
	perf := types.NewValidatorPerformance("validator1", 1)
	perf.BlocksProposed = 20
	perf.BlocksExpected = 100
	perf.VEIDVerificationsCompleted = 1
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = 5000
	perf.UptimeSeconds = 1800
	perf.DowntimeSeconds = 1800
	types.ComputeOverallScore(perf)

	input := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perf,
		StakeAmount:      1_000_000,
		TotalStake:       1_000_000,
		EpochRewardPool:  100_000_000,
	}

	reward := types.CalculateRewards(input, "uve")
	total := reward.TotalReward.AmountOf("uve").Int64()
	s.Require().Greater(total, int64(0))
	s.Require().Less(total, int64(100_000_000))
}

func (s *RewardsTestSuite) TestCalculateRewardsAllocatesRemainderDeterministically() {
	perf := types.NewValidatorPerformance("validator-remainder", 1)
	perf.BlocksProposed = 10
	perf.BlocksExpected = 10
	perf.VEIDVerificationsCompleted = 10
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = types.MaxPerformanceScore
	perf.UptimeSeconds = 3600
	perf.DowntimeSeconds = 0
	types.ComputeOverallScore(perf)

	reward := types.CalculateRewards(types.RewardCalculationInput{
		ValidatorAddress: "validator-remainder",
		Performance:      perf,
		StakeAmount:      1,
		TotalStake:       2,
		EpochRewardPool:  3,
		BlocksInEpoch:    3,
	}, "uve")

	s.Require().Equal(int64(1), reward.TotalReward.AmountOf("uve").Int64())
	s.Require().True(reward.BlockProposalReward.IsZero())
	s.Require().Equal(int64(1), reward.VEIDReward.AmountOf("uve").Int64())
	s.Require().True(reward.UptimeReward.IsZero())
}

// TestVEIDBonusForHighScore tests score-sensitive VEID rewards.
func (s *RewardsTestSuite) TestVEIDBonusForHighScore() {
	// High VEID score (>= 9000)
	perfHigh := types.NewValidatorPerformance("validator1", 1)
	perfHigh.VEIDVerificationScore = 9500
	types.ComputeOverallScore(perfHigh)

	// Medium VEID score (< 9000)
	perfMed := types.NewValidatorPerformance("validator2", 1)
	perfMed.VEIDVerificationScore = 8000
	types.ComputeOverallScore(perfMed)

	stake := int64(500000000)
	totalStake := int64(1000000000)
	epochRewardPool := int64(100000000)

	inputHigh := types.RewardCalculationInput{
		ValidatorAddress: "validator1",
		Performance:      perfHigh,
		StakeAmount:      stake,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	inputMed := types.RewardCalculationInput{
		ValidatorAddress: "validator2",
		Performance:      perfMed,
		StakeAmount:      stake,
		TotalStake:       totalStake,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    100,
	}

	rewardHigh := types.CalculateRewards(inputHigh, "uve")
	rewardMed := types.CalculateRewards(inputMed, "uve")

	// High performer should get VEID bonus
	veidHigh := rewardHigh.VEIDReward.AmountOf("uve").Int64()
	veidMed := rewardMed.VEIDReward.AmountOf("uve").Int64()

	s.Require().Greater(veidHigh, veidMed)
}

func (s *RewardsTestSuite) TestCalculateEpochRewardsMatchesDeterministicAbsoluteEconomics() {
	epoch := uint64(1)
	validatorAddr := sdk.AccAddress([]byte("reward-keeper-validator-01")).String()
	otherValidatorAddr := sdk.AccAddress([]byte("reward-keeper-validator-02")).String()

	epochInfo := types.NewRewardEpoch(epoch, 1, s.ctx.BlockTime())
	epochInfo.EndHeight = 101
	s.Require().NoError(s.keeper.SetRewardEpoch(s.ctx, *epochInfo))

	s.stakeKeeper.SetStake(validatorAddr, 700_000)
	s.stakeKeeper.SetStake(otherValidatorAddr, 300_000)

	perf := types.NewValidatorPerformance(validatorAddr, epoch)
	perf.BlocksProposed = 20
	perf.BlocksExpected = 100
	perf.VEIDVerificationsCompleted = 1
	perf.VEIDVerificationsExpected = 10
	perf.VEIDVerificationScore = 5_000
	perf.UptimeSeconds = 1_800
	perf.DowntimeSeconds = 1_800
	types.ComputeOverallScore(perf)
	s.Require().NoError(s.keeper.SetValidatorPerformance(s.ctx, *perf))

	params := s.keeper.GetParams(s.ctx)
	epochRewardPool := params.BaseRewardPerBlock * types.EpochDuration(epochInfo)
	expected := types.CalculateRewards(types.RewardCalculationInput{
		ValidatorAddress: validatorAddr,
		Performance:      perf,
		StakeAmount:      700_000,
		TotalStake:       1_000_000,
		EpochRewardPool:  epochRewardPool,
		BlocksInEpoch:    types.EpochDuration(epochInfo),
	}, params.RewardDenom)

	rewards, err := s.keeper.CalculateEpochRewards(s.ctx, epoch)
	s.Require().NoError(err)
	s.Require().Len(rewards, 1)

	actual := rewards[0]
	s.Require().True(expected.TotalReward.Equal(actual.TotalReward))
	s.Require().True(expected.BlockProposalReward.Equal(actual.BlockProposalReward))
	s.Require().True(expected.VEIDReward.Equal(actual.VEIDReward))
	s.Require().True(expected.UptimeReward.Equal(actual.UptimeReward))
	s.Require().Less(actual.TotalReward.AmountOf(params.RewardDenom).Int64(), epochRewardPool)
}
