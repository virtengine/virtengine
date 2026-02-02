// Package keeper implements the delegation module keeper tests.
//
// VE-922: Delegation keeper tests
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

// Test address constants
const (
	testDelegatorAddr = "cosmos1delegator123"
	testValidatorAddr = "cosmos1validator456"
	testValidator123  = "cosmos1validator123"
)

// DelegationKeeperTestSuite is the test suite for the delegation keeper
type DelegationKeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
	skey   *storetypes.KVStoreKey
}

// SetupTest sets up the test suite
func (s *DelegationKeeperTestSuite) SetupTest() {
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

// TestDelegationKeeperTestSuite runs the test suite
func TestDelegationKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(DelegationKeeperTestSuite))
}

// TestParams tests parameter management
func (s *DelegationKeeperTestSuite) TestParams() {
	// Get default params
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultUnbondingPeriod, params.UnbondingPeriod)
	s.Require().Equal(types.DefaultMinDelegationAmount, params.MinDelegationAmount)
	s.Require().Equal(types.DefaultMaxValidatorsPerDelegator, params.MaxValidatorsPerDelegator)
	s.Require().Equal("uve", params.RewardDenom)

	// Update params
	params.UnbondingPeriod = 14 * 24 * 60 * 60 // 14 days
	params.MinDelegationAmount = 2000000
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Verify update
	updatedParams := s.keeper.GetParams(s.ctx)
	s.Require().Equal(int64(14*24*60*60), updatedParams.UnbondingPeriod)
	s.Require().Equal(int64(2000000), updatedParams.MinDelegationAmount)
}

// TestDelegationStorage tests delegation storage
func (s *DelegationKeeperTestSuite) TestDelegationStorage() {
	delegatorAddr := testDelegatorAddr
	validatorAddr := testValidatorAddr

	// Initially no delegation
	_, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().False(found)

	// Create delegation
	del := types.NewDelegation(
		delegatorAddr,
		validatorAddr,
		"1000000000000000000", // 1 share with 18 decimals
		"1000000",             // 1 token
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)

	err := s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Get delegation
	retrieved, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().True(found)
	s.Require().Equal(delegatorAddr, retrieved.DelegatorAddress)
	s.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	s.Require().Equal("1000000000000000000", retrieved.Shares)
	s.Require().Equal("1000000", retrieved.InitialAmount)
}

// TestValidatorShares tests validator shares management
func (s *DelegationKeeperTestSuite) TestValidatorShares() {
	validatorAddr := testValidator123

	// Initially no shares
	_, found := s.keeper.GetValidatorShares(s.ctx, validatorAddr)
	s.Require().False(found)

	// Create validator shares
	shares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err := shares.AddShares("1000000000000000000", "1000000", s.ctx.BlockTime())
	s.Require().NoError(err)

	err = s.keeper.SetValidatorShares(s.ctx, *shares)
	s.Require().NoError(err)

	// Get validator shares
	retrieved, found := s.keeper.GetValidatorShares(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Equal("1000000000000000000", retrieved.TotalShares)
	s.Require().Equal("1000000", retrieved.TotalStake)
}

// TestCalculateSharesForAmount tests share calculation
func (s *DelegationKeeperTestSuite) TestCalculateSharesForAmount() {
	validatorAddr := "cosmos1validator123"

	// Test with no existing shares (first delegation)
	shares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	newShares, err := shares.CalculateSharesForAmount("1000000")
	s.Require().NoError(err)
	// Should be amount * 10^18
	s.Require().Equal("1000000000000000000000000", newShares)

	// Add shares and stake
	err = shares.AddShares(newShares, "1000000", s.ctx.BlockTime())
	s.Require().NoError(err)

	// Test with existing shares (second delegation)
	// Same amount should get same shares
	secondShares, err := shares.CalculateSharesForAmount("1000000")
	s.Require().NoError(err)
	s.Require().Equal(newShares, secondShares)

	// Double amount should get double shares
	doubleShares, err := shares.CalculateSharesForAmount("2000000")
	s.Require().NoError(err)
	s.Require().Equal("2000000000000000000000000", doubleShares)
}

// TestCalculateAmountForShares tests amount calculation from shares
func (s *DelegationKeeperTestSuite) TestCalculateAmountForShares() {
	validatorAddr := "cosmos1validator123"

	shares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	// Add 1M shares for 1M stake
	err := shares.AddShares("1000000000000000000000000", "1000000", s.ctx.BlockTime())
	s.Require().NoError(err)

	// Calculate amount for half the shares
	amount, err := shares.CalculateAmountForShares("500000000000000000000000")
	s.Require().NoError(err)
	s.Require().Equal("500000", amount)
}

// TestUnbondingDelegation tests unbonding delegation storage
func (s *DelegationKeeperTestSuite) TestUnbondingDelegation() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"
	completionTime := s.ctx.BlockTime().Add(21 * 24 * time.Hour)

	// Create unbonding delegation
	ubd := types.NewUnbondingDelegation(
		"ubd-12345678",
		delegatorAddr,
		validatorAddr,
		s.ctx.BlockHeight(),
		completionTime,
		s.ctx.BlockTime(),
		"1000000",
		"1000000000000000000",
	)

	err := s.keeper.SetUnbondingDelegation(s.ctx, *ubd)
	s.Require().NoError(err)

	// Get unbonding delegation
	retrieved, found := s.keeper.GetUnbondingDelegation(s.ctx, "ubd-12345678")
	s.Require().True(found)
	s.Require().Equal(delegatorAddr, retrieved.DelegatorAddress)
	s.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	s.Require().Len(retrieved.Entries, 1)
	s.Require().Equal("1000000", retrieved.Entries[0].Balance)
}

// TestRedelegation tests redelegation storage
func (s *DelegationKeeperTestSuite) TestRedelegation() {
	delegatorAddr := "cosmos1delegator123"
	srcValidator := "cosmos1validator1"
	dstValidator := "cosmos1validator2"
	completionTime := s.ctx.BlockTime().Add(21 * 24 * time.Hour)

	// Create redelegation
	red := types.NewRedelegation(
		"red-12345678",
		delegatorAddr,
		srcValidator,
		dstValidator,
		s.ctx.BlockHeight(),
		completionTime,
		s.ctx.BlockTime(),
		"1000000",
		"1000000000000000000",
	)

	err := s.keeper.SetRedelegation(s.ctx, *red)
	s.Require().NoError(err)

	// Get redelegation
	retrieved, found := s.keeper.GetRedelegation(s.ctx, "red-12345678")
	s.Require().True(found)
	s.Require().Equal(delegatorAddr, retrieved.DelegatorAddress)
	s.Require().Equal(srcValidator, retrieved.ValidatorSrcAddress)
	s.Require().Equal(dstValidator, retrieved.ValidatorDstAddress)

	// Check HasRedelegation
	hasRed := s.keeper.HasRedelegation(s.ctx, delegatorAddr, srcValidator)
	s.Require().True(hasRed)

	hasRed = s.keeper.HasRedelegation(s.ctx, delegatorAddr, dstValidator)
	s.Require().False(hasRed)
}

// TestDelegatorReward tests delegator reward storage
func (s *DelegationKeeperTestSuite) TestDelegatorReward() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"
	epoch := uint64(10)

	// Create delegator reward
	reward := types.NewDelegatorReward(
		delegatorAddr,
		validatorAddr,
		epoch,
		"500000",
		"1000000000000000000",
		"10000000000000000000",
		s.ctx.BlockTime(),
	)

	err := s.keeper.SetDelegatorReward(s.ctx, *reward)
	s.Require().NoError(err)

	// Get delegator reward
	retrieved, found := s.keeper.GetDelegatorReward(s.ctx, delegatorAddr, validatorAddr, epoch)
	s.Require().True(found)
	s.Require().Equal(delegatorAddr, retrieved.DelegatorAddress)
	s.Require().Equal(validatorAddr, retrieved.ValidatorAddress)
	s.Require().Equal("500000", retrieved.Reward)
	s.Require().False(retrieved.Claimed)
}

// TestGetDelegatorDelegations tests getting all delegations for a delegator
func (s *DelegationKeeperTestSuite) TestGetDelegatorDelegations() {
	delegatorAddr := "cosmos1delegator123"
	validator1 := "cosmos1validator1"
	validator2 := "cosmos1validator2"

	// Create delegations
	del1 := types.NewDelegation(delegatorAddr, validator1, "1000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del2 := types.NewDelegation(delegatorAddr, validator2, "2000000000000000000", "2000000", s.ctx.BlockTime(), s.ctx.BlockHeight())

	err := s.keeper.SetDelegation(s.ctx, *del1)
	s.Require().NoError(err)
	err = s.keeper.SetDelegation(s.ctx, *del2)
	s.Require().NoError(err)

	// Get all delegations
	delegations := s.keeper.GetDelegatorDelegations(s.ctx, delegatorAddr)
	s.Require().Len(delegations, 2)
}

// TestGetValidatorDelegations tests getting all delegations for a validator
func (s *DelegationKeeperTestSuite) TestGetValidatorDelegations() {
	validatorAddr := "cosmos1validator123"
	delegator1 := "cosmos1delegator1"
	delegator2 := "cosmos1delegator2"

	// Create delegations
	del1 := types.NewDelegation(delegator1, validatorAddr, "1000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del2 := types.NewDelegation(delegator2, validatorAddr, "2000000000000000000", "2000000", s.ctx.BlockTime(), s.ctx.BlockHeight())

	err := s.keeper.SetDelegation(s.ctx, *del1)
	s.Require().NoError(err)
	err = s.keeper.SetDelegation(s.ctx, *del2)
	s.Require().NoError(err)

	// Get all delegations for validator
	delegations := s.keeper.GetValidatorDelegations(s.ctx, validatorAddr)
	s.Require().Len(delegations, 2)
}

// TestMatureUnbondings tests getting mature unbonding delegations
func (s *DelegationKeeperTestSuite) TestMatureUnbondings() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"

	// Create unbonding that's already mature
	pastTime := s.ctx.BlockTime().Add(-1 * time.Hour)
	ubd1 := types.NewUnbondingDelegation(
		"ubd-mature",
		delegatorAddr,
		validatorAddr,
		s.ctx.BlockHeight()-100,
		pastTime,
		pastTime.Add(-21*24*time.Hour),
		"1000000",
		"1000000000000000000",
	)

	// Create unbonding that's not yet mature
	futureTime := s.ctx.BlockTime().Add(21 * 24 * time.Hour)
	ubd2 := types.NewUnbondingDelegation(
		"ubd-pending",
		delegatorAddr,
		validatorAddr,
		s.ctx.BlockHeight(),
		futureTime,
		s.ctx.BlockTime(),
		"2000000",
		"2000000000000000000",
	)

	err := s.keeper.SetUnbondingDelegation(s.ctx, *ubd1)
	s.Require().NoError(err)
	err = s.keeper.SetUnbondingDelegation(s.ctx, *ubd2)
	s.Require().NoError(err)

	// Get mature unbondings
	matureUnbondings := s.keeper.GetMatureUnbondings(s.ctx)
	s.Require().Len(matureUnbondings, 1)
	s.Require().Equal("ubd-mature", matureUnbondings[0].ID)
}

// TestCalculateDelegatorProportion tests delegator proportion calculation
func (s *DelegationKeeperTestSuite) TestCalculateDelegatorProportion() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"

	// Set up validator shares (total 10 tokens)
	shares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err := shares.AddShares("10000000000000000000000000", "10000000", s.ctx.BlockTime())
	s.Require().NoError(err)
	err = s.keeper.SetValidatorShares(s.ctx, *shares)
	s.Require().NoError(err)

	// Set up delegation (1 token = 10% of total)
	del := types.NewDelegation(delegatorAddr, validatorAddr, "1000000000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err = s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Calculate proportion (should be 1000 basis points = 10%)
	proportion, err := s.keeper.CalculateDelegatorProportion(s.ctx, delegatorAddr, validatorAddr)
	s.Require().NoError(err)
	s.Require().Equal(int64(1000), proportion)
}

// TestCountDelegatorRedelegations tests counting redelegations
func (s *DelegationKeeperTestSuite) TestCountDelegatorRedelegations() {
	delegatorAddr := "cosmos1delegator123"
	completionTime := s.ctx.BlockTime().Add(21 * 24 * time.Hour)

	// Create redelegations
	red1 := types.NewRedelegation("red-1", delegatorAddr, "validator1", "validator2", s.ctx.BlockHeight(), completionTime, s.ctx.BlockTime(), "1000000", "1000000000000000000")
	red2 := types.NewRedelegation("red-2", delegatorAddr, "validator2", "validator3", s.ctx.BlockHeight(), completionTime, s.ctx.BlockTime(), "1000000", "1000000000000000000")

	err := s.keeper.SetRedelegation(s.ctx, *red1)
	s.Require().NoError(err)
	err = s.keeper.SetRedelegation(s.ctx, *red2)
	s.Require().NoError(err)

	// Count redelegations
	count := s.keeper.CountDelegatorRedelegations(s.ctx, delegatorAddr)
	s.Require().Equal(2, count)
}

// TestDeleteDelegation tests delegation deletion
func (s *DelegationKeeperTestSuite) TestDeleteDelegation() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"

	// Create delegation
	del := types.NewDelegation(delegatorAddr, validatorAddr, "1000000000000000000", "1000000", s.ctx.BlockTime(), s.ctx.BlockHeight())
	err := s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Verify exists
	_, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().True(found)

	// Delete
	s.keeper.DeleteDelegation(s.ctx, delegatorAddr, validatorAddr)

	// Verify deleted
	_, found = s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().False(found)
}

// TestGetDelegatorUnclaimedRewards tests getting unclaimed rewards
func (s *DelegationKeeperTestSuite) TestGetDelegatorUnclaimedRewards() {
	delegatorAddr := "cosmos1delegator123"
	validatorAddr := "cosmos1validator456"

	// Create unclaimed reward
	reward1 := types.NewDelegatorReward(delegatorAddr, validatorAddr, 1, "100000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	err := s.keeper.SetDelegatorReward(s.ctx, *reward1)
	s.Require().NoError(err)

	// Create claimed reward
	reward2 := types.NewDelegatorReward(delegatorAddr, validatorAddr, 2, "200000", "1000000000000000000", "10000000000000000000", s.ctx.BlockTime())
	reward2.Claimed = true
	now := s.ctx.BlockTime()
	reward2.ClaimedAt = &now
	err = s.keeper.SetDelegatorReward(s.ctx, *reward2)
	s.Require().NoError(err)

	// Get unclaimed rewards
	unclaimedRewards := s.keeper.GetDelegatorUnclaimedRewards(s.ctx, delegatorAddr)
	s.Require().Len(unclaimedRewards, 1)
	s.Require().Equal("100000", unclaimedRewards[0].Reward)
}
