// Package keeper implements the delegation module keeper tests.
//
// VE-922: Reward distribution tests
package keeper

import (
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

	"github.com/virtengine/virtengine/x/delegation/types"
)

// Valid bech32 test addresses (must be 20 bytes for proper encoding)
var (
	testValidatorAddr  = sdk.AccAddress([]byte("validator_address___")).String()
	testDelegator1Addr = sdk.AccAddress([]byte("delegator1_address__")).String()
	testDelegator2Addr = sdk.AccAddress([]byte("delegator2_address__")).String()
	testValidator1Addr = sdk.AccAddress([]byte("validator1_address__")).String()
	testValidator2Addr = sdk.AccAddress([]byte("validator2_address__")).String()
)

// RewardsTestSuite is the test suite for reward distribution
type RewardsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
	skey   *storetypes.KVStoreKey
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

	s.keeper = NewKeeper(
		s.cdc,
		s.skey,
		nil, // bankKeeper
		nil, // stakingRewardsKeeper
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

// TestDistributeValidatorRewardsToDelegators tests reward distribution
func (s *RewardsTestSuite) TestDistributeValidatorRewardsToDelegators() {
	validatorAddr := "cosmos1validator123456789abcdef"
	delegator1 := "cosmos1delegator1111111111111111"
	delegator2 := "cosmos1delegator2222222222222222"
	epoch := uint64(10)
	validatorReward := "10000000" // 10 tokens in reward

	// Create validator shares with total stake of 10 tokens
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	// Total shares: 10 tokens worth
	err := valShares.AddShares("10000000000000000000000000", "10000000", s.ctx.BlockTime())
	s.Require().NoError(err)
	err = s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	// Create delegations
	// Delegator1: 6 tokens (60%)
	del1 := types.NewDelegation(delegator1, validatorAddr, "6000000000000000000000000", "6000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del1)
	s.Require().NoError(err)

	// Delegator2: 4 tokens (40%)
	del2 := types.NewDelegation(delegator2, validatorAddr, "4000000000000000000000000", "4000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del2)
	s.Require().NoError(err)

	// Distribute rewards
	err = s.keeper.DistributeValidatorRewardsToDelegators(s.ctx, validatorAddr, epoch, validatorReward)
	s.Require().NoError(err)

	// Check delegator1 reward
	// Commission: 10% = 1,000,000
	// Distributable: 9,000,000
	// Delegator1 (60%): 5,400,000
	reward1, found := s.keeper.GetDelegatorReward(s.ctx, delegator1, validatorAddr, epoch)
	s.Require().True(found)
	s.Require().Equal("5400000", reward1.Reward)
	s.Require().False(reward1.Claimed)

	// Check delegator2 reward
	// Delegator2 (40%): 3,600,000
	reward2, found := s.keeper.GetDelegatorReward(s.ctx, delegator2, validatorAddr, epoch)
	s.Require().True(found)
	s.Require().Equal("3600000", reward2.Reward)
	s.Require().False(reward2.Claimed)
}

// TestDistributeRewardsNoCommission tests distribution with zero commission
func (s *RewardsTestSuite) TestDistributeRewardsNoCommission() {
	validatorAddr := "cosmos1validator123456789abcdef"
	delegator := "cosmos1delegator1111111111111111"
	epoch := uint64(10)
	validatorReward := "10000000"

	// Set commission to 0
	params := s.keeper.GetParams(s.ctx)
	params.ValidatorCommissionRate = 0
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Create validator shares
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err = valShares.AddShares("1000000000000000000000000", "1000000", s.ctx.BlockTime())
	s.Require().NoError(err)
	err = s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	// Create delegation (100%)
	del := types.NewDelegation(delegator, validatorAddr, "1000000000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Distribute rewards
	err = s.keeper.DistributeValidatorRewardsToDelegators(s.ctx, validatorAddr, epoch, validatorReward)
	s.Require().NoError(err)

	// Delegator should get 100% (no commission)
	reward, found := s.keeper.GetDelegatorReward(s.ctx, delegator, validatorAddr, epoch)
	s.Require().True(found)
	s.Require().Equal("10000000", reward.Reward)
}

// TestDistributeRewardsNoDelegations tests distribution with no delegations
func (s *RewardsTestSuite) TestDistributeRewardsNoDelegations() {
	validatorAddr := "cosmos1validator123456789abcdef"
	epoch := uint64(10)
	validatorReward := "10000000"

	// No delegations or shares - should not error
	err := s.keeper.DistributeValidatorRewardsToDelegators(s.ctx, validatorAddr, epoch, validatorReward)
	s.Require().NoError(err)
}

// TestClaimRewards tests claiming rewards from a specific validator
func (s *RewardsTestSuite) TestClaimRewards() {
	delegatorAddr := testDelegator1Addr
	validatorAddr := testValidatorAddr

	// Create unclaimed rewards for multiple epochs
	reward1 := types.NewDelegatorReward(delegatorAddr, validatorAddr, 1, "100000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err := s.keeper.SetDelegatorReward(s.ctx, *reward1)
	s.Require().NoError(err)

	reward2 := types.NewDelegatorReward(delegatorAddr, validatorAddr, 2, "200000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err = s.keeper.SetDelegatorReward(s.ctx, *reward2)
	s.Require().NoError(err)

	// Claim rewards (bank transfer would happen with real bank keeper)
	claimedCoins, err := s.keeper.ClaimRewards(s.ctx, delegatorAddr, validatorAddr)
	s.Require().NoError(err)

	// Verify claimed amount (100000 + 200000 = 300000)
	s.Require().Equal(1, len(claimedCoins))
	s.Require().Equal("300000uve", claimedCoins.String())

	// Verify rewards are marked as claimed
	r1, _ := s.keeper.GetDelegatorReward(s.ctx, delegatorAddr, validatorAddr, 1)
	s.Require().True(r1.Claimed)
	s.Require().NotNil(r1.ClaimedAt)

	r2, _ := s.keeper.GetDelegatorReward(s.ctx, delegatorAddr, validatorAddr, 2)
	s.Require().True(r2.Claimed)
	s.Require().NotNil(r2.ClaimedAt)
}

// TestClaimAllRewards tests claiming all rewards from all validators
func (s *RewardsTestSuite) TestClaimAllRewards() {
	delegatorAddr := testDelegator1Addr
	validator1 := testValidator1Addr
	validator2 := testValidator2Addr

	// Create rewards from multiple validators
	reward1 := types.NewDelegatorReward(delegatorAddr, validator1, 1, "100000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err := s.keeper.SetDelegatorReward(s.ctx, *reward1)
	s.Require().NoError(err)

	reward2 := types.NewDelegatorReward(delegatorAddr, validator2, 1, "200000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err = s.keeper.SetDelegatorReward(s.ctx, *reward2)
	s.Require().NoError(err)

	// Claim all rewards
	_, err = s.keeper.ClaimAllRewards(s.ctx, delegatorAddr)
	s.Require().NoError(err)

	// Verify all rewards are claimed
	unclaimedRewards := s.keeper.GetDelegatorUnclaimedRewards(s.ctx, delegatorAddr)
	s.Require().Len(unclaimedRewards, 0)
}

// TestClaimRewardsNoRewards tests claiming when no rewards exist
func (s *RewardsTestSuite) TestClaimRewardsNoRewards() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"

	// No rewards exist
	claimedCoins, err := s.keeper.ClaimRewards(s.ctx, delegatorAddr, validatorAddr)
	s.Require().NoError(err)
	s.Require().Len(claimedCoins, 0)
}

// TestGetDelegatorTotalRewards tests getting total unclaimed rewards
func (s *RewardsTestSuite) TestGetDelegatorTotalRewards() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validator1 := "cosmos1validator1111111111111111"
	validator2 := "cosmos1validator2222222222222222"

	// Create unclaimed rewards
	reward1 := types.NewDelegatorReward(delegatorAddr, validator1, 1, "100000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err := s.keeper.SetDelegatorReward(s.ctx, *reward1)
	s.Require().NoError(err)

	reward2 := types.NewDelegatorReward(delegatorAddr, validator2, 1, "200000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err = s.keeper.SetDelegatorReward(s.ctx, *reward2)
	s.Require().NoError(err)

	// Create claimed reward (should not be counted)
	reward3 := types.NewDelegatorReward(delegatorAddr, validator1, 2, "500000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	reward3.Claimed = true
	now := s.ctx.BlockTime()
	reward3.ClaimedAt = &now
	err = s.keeper.SetDelegatorReward(s.ctx, *reward3)
	s.Require().NoError(err)

	// Get total unclaimed rewards
	total := s.keeper.GetDelegatorTotalRewards(s.ctx, delegatorAddr)
	s.Require().Equal("300000", total) // 100000 + 200000
}

// TestGetDelegatorValidatorTotalRewards tests getting total from specific validator
func (s *RewardsTestSuite) TestGetDelegatorValidatorTotalRewards() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validator1 := "cosmos1validator1111111111111111"
	validator2 := "cosmos1validator2222222222222222"

	// Create rewards from validator1
	reward1 := types.NewDelegatorReward(delegatorAddr, validator1, 1, "100000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err := s.keeper.SetDelegatorReward(s.ctx, *reward1)
	s.Require().NoError(err)

	reward2 := types.NewDelegatorReward(delegatorAddr, validator1, 2, "150000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err = s.keeper.SetDelegatorReward(s.ctx, *reward2)
	s.Require().NoError(err)

	// Create reward from validator2 (should not be counted)
	reward3 := types.NewDelegatorReward(delegatorAddr, validator2, 1, "200000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err = s.keeper.SetDelegatorReward(s.ctx, *reward3)
	s.Require().NoError(err)

	// Get total from validator1
	total := s.keeper.GetDelegatorValidatorTotalRewards(s.ctx, delegatorAddr, validator1)
	s.Require().Equal("250000", total) // 100000 + 150000
}

// TestCalculateDelegatorRewardAmount tests reward amount calculation
func (s *RewardsTestSuite) TestCalculateDelegatorRewardAmount() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	validatorReward := "10000000" // 10 tokens

	// Create validator shares (total: 10 tokens)
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err := valShares.AddShares("10000000000000000000000000", "10000000", s.ctx.BlockTime())
	s.Require().NoError(err)
	err = s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	// Create delegation (5 tokens = 50%)
	del := types.NewDelegation(delegatorAddr, validatorAddr, "5000000000000000000000000", "5000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Calculate reward
	// Commission: 10% = 1,000,000
	// Distributable: 9,000,000
	// Delegator (50%): 4,500,000
	rewardAmount, err := s.keeper.CalculateDelegatorRewardAmount(s.ctx, delegatorAddr, validatorAddr, validatorReward)
	s.Require().NoError(err)
	s.Require().Equal("4500000", rewardAmount)
}

// TestRewardDistributionProportional tests that rewards are proportional to shares
func (s *RewardsTestSuite) TestRewardDistributionProportional() {
	validatorAddr := "cosmos1validator123456789abcdef"
	delegator1 := "cosmos1delegator1111111111111111"
	delegator2 := "cosmos1delegator2222222222222222"
	delegator3 := "cosmos1delegator3333333333333333"
	epoch := uint64(10)
	validatorReward := "100000000" // 100 tokens

	// Set no commission for easier calculation
	params := s.keeper.GetParams(s.ctx)
	params.ValidatorCommissionRate = 0
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Create validator shares (total: 100 tokens)
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err = valShares.AddShares("100000000000000000000000000", "100000000", s.ctx.BlockTime())
	s.Require().NoError(err)
	err = s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	// Create delegations
	// Delegator1: 50 tokens (50%)
	del1 := types.NewDelegation(delegator1, validatorAddr, "50000000000000000000000000", "50000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del1)
	s.Require().NoError(err)

	// Delegator2: 30 tokens (30%)
	del2 := types.NewDelegation(delegator2, validatorAddr, "30000000000000000000000000", "30000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del2)
	s.Require().NoError(err)

	// Delegator3: 20 tokens (20%)
	del3 := types.NewDelegation(delegator3, validatorAddr, "20000000000000000000000000", "20000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del3)
	s.Require().NoError(err)

	// Distribute rewards
	err = s.keeper.DistributeValidatorRewardsToDelegators(s.ctx, validatorAddr, epoch, validatorReward)
	s.Require().NoError(err)

	// Check proportional distribution
	reward1, _ := s.keeper.GetDelegatorReward(s.ctx, delegator1, validatorAddr, epoch)
	s.Require().Equal("50000000", reward1.Reward) // 50%

	reward2, _ := s.keeper.GetDelegatorReward(s.ctx, delegator2, validatorAddr, epoch)
	s.Require().Equal("30000000", reward2.Reward) // 30%

	reward3, _ := s.keeper.GetDelegatorReward(s.ctx, delegator3, validatorAddr, epoch)
	s.Require().Equal("20000000", reward3.Reward) // 20%
}
