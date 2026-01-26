// Package keeper implements the delegation module keeper tests.
//
// VE-922: Delegation operation tests
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

	"pkg.akt.dev/node/x/delegation/types"
)

// DelegationOperationsTestSuite is the test suite for delegation operations
type DelegationOperationsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
	skey   *storetypes.KVStoreKey
}

// SetupTest sets up the test suite
func (s *DelegationOperationsTestSuite) SetupTest() {
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
		nil, // bankKeeper - not needed for these tests
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

// TestDelegationOperationsTestSuite runs the test suite
func TestDelegationOperationsTestSuite(t *testing.T) {
	suite.Run(t, new(DelegationOperationsTestSuite))
}

// TestDelegateBasic tests basic delegation
func (s *DelegationOperationsTestSuite) TestDelegateBasic() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(2000000)) // 2 tokens

	// Delegate
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Verify delegation was created
	del, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().True(found)
	s.Require().Equal(delegatorAddr, del.DelegatorAddress)
	s.Require().Equal(validatorAddr, del.ValidatorAddress)
	s.Require().Equal("2000000", del.InitialAmount)

	// Verify validator shares were created
	valShares, found := s.keeper.GetValidatorShares(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Equal("2000000", valShares.TotalStake)
}

// TestDelegateIncrease tests increasing an existing delegation
func (s *DelegationOperationsTestSuite) TestDelegateIncrease() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount1 := sdk.NewCoin("uve", sdk.NewInt(1000000)) // 1 token
	amount2 := sdk.NewCoin("uve", sdk.NewInt(2000000)) // 2 tokens

	// First delegation
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount1)
	s.Require().NoError(err)

	// Get initial shares
	del1, _ := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	initialShares := del1.Shares

	// Second delegation (increase)
	err = s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount2)
	s.Require().NoError(err)

	// Verify shares increased
	del2, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().True(found)
	s.Require().NotEqual(initialShares, del2.Shares) // Shares should have increased

	// Verify validator total stake
	valShares, _ := s.keeper.GetValidatorShares(s.ctx, validatorAddr)
	s.Require().Equal("3000000", valShares.TotalStake) // 1 + 2 = 3 tokens
}

// TestDelegateMinimumAmount tests minimum delegation amount enforcement
func (s *DelegationOperationsTestSuite) TestDelegateMinimumAmount() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(100)) // Below minimum

	// Should fail with minimum amount error
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "minimum delegation")
}

// TestDelegateMaxValidators tests max validators per delegator enforcement
func (s *DelegationOperationsTestSuite) TestDelegateMaxValidators() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// Set max validators to 2 for testing
	params := s.keeper.GetParams(s.ctx)
	params.MaxValidatorsPerDelegator = 2
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Delegate to first validator
	err = s.keeper.Delegate(s.ctx, delegatorAddr, "cosmos1validator1", amount)
	s.Require().NoError(err)

	// Delegate to second validator
	err = s.keeper.Delegate(s.ctx, delegatorAddr, "cosmos1validator2", amount)
	s.Require().NoError(err)

	// Third validator should fail
	err = s.keeper.Delegate(s.ctx, delegatorAddr, "cosmos1validator3", amount)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max validators")

	// But increasing existing delegation should work
	err = s.keeper.Delegate(s.ctx, delegatorAddr, "cosmos1validator1", amount)
	s.Require().NoError(err)
}

// TestUndelegate tests basic undelegation
func (s *DelegationOperationsTestSuite) TestUndelegate() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	delegateAmount := sdk.NewCoin("uve", sdk.NewInt(3000000))  // 3 tokens
	undelegateAmount := sdk.NewCoin("uve", sdk.NewInt(1000000)) // 1 token

	// First delegate
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, delegateAmount)
	s.Require().NoError(err)

	// Then undelegate
	completionTime, err := s.keeper.Undelegate(s.ctx, delegatorAddr, validatorAddr, undelegateAmount)
	s.Require().NoError(err)
	s.Require().False(completionTime.IsZero())

	// Verify unbonding delegation was created
	unbondings := s.keeper.GetDelegatorUnbondingDelegations(s.ctx, delegatorAddr)
	s.Require().Len(unbondings, 1)
	s.Require().Equal(validatorAddr, unbondings[0].ValidatorAddress)

	// Verify remaining delegation
	del, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().True(found)

	// Verify validator shares decreased
	valShares, _ := s.keeper.GetValidatorShares(s.ctx, validatorAddr)
	s.Require().Equal("2000000", valShares.TotalStake) // 3 - 1 = 2 tokens
	s.Require().NotEmpty(del.Shares)
}

// TestUndelegateAll tests undelegating entire delegation
func (s *DelegationOperationsTestSuite) TestUndelegateAll() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(2000000))

	// Delegate
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Undelegate all
	_, err = s.keeper.Undelegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Delegation should be deleted
	_, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, validatorAddr)
	s.Require().False(found)
}

// TestUndelegateNotFound tests undelegating non-existent delegation
func (s *DelegationOperationsTestSuite) TestUndelegateNotFound() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// Should fail
	_, err := s.keeper.Undelegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")
}

// TestRedelegate tests basic redelegation
func (s *DelegationOperationsTestSuite) TestRedelegate() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	srcValidator := "cosmos1validator1111111111111111"
	dstValidator := "cosmos1validator2222222222222222"
	delegateAmount := sdk.NewCoin("uve", sdk.NewInt(3000000))
	redelegateAmount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// First delegate to source validator
	err := s.keeper.Delegate(s.ctx, delegatorAddr, srcValidator, delegateAmount)
	s.Require().NoError(err)

	// Redelegate to destination validator
	completionTime, err := s.keeper.Redelegate(s.ctx, delegatorAddr, srcValidator, dstValidator, redelegateAmount)
	s.Require().NoError(err)
	s.Require().False(completionTime.IsZero())

	// Verify source delegation decreased
	srcDel, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, srcValidator)
	s.Require().True(found)
	srcValShares, _ := s.keeper.GetValidatorShares(s.ctx, srcValidator)
	s.Require().Equal("2000000", srcValShares.TotalStake)
	s.Require().NotEmpty(srcDel.Shares)

	// Verify destination delegation created
	dstDel, found := s.keeper.GetDelegation(s.ctx, delegatorAddr, dstValidator)
	s.Require().True(found)
	s.Require().NotEmpty(dstDel.Shares)
	dstValShares, _ := s.keeper.GetValidatorShares(s.ctx, dstValidator)
	s.Require().Equal("1000000", dstValShares.TotalStake)

	// Verify redelegation record created
	redelegations := s.keeper.GetDelegatorRedelegations(s.ctx, delegatorAddr)
	s.Require().Len(redelegations, 1)
}

// TestRedelegateSelfDelegation tests that self-redelegation is prevented
func (s *DelegationOperationsTestSuite) TestRedelegateSelfDelegation() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// Delegate
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Try to redelegate to same validator - should fail
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, validatorAddr, validatorAddr, amount)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "same validator")
}

// TestRedelegateTransitive tests that transitive redelegation is prevented
func (s *DelegationOperationsTestSuite) TestRedelegateTransitive() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validator1 := "cosmos1validator1111111111111111"
	validator2 := "cosmos1validator2222222222222222"
	validator3 := "cosmos1validator3333333333333333"
	amount := sdk.NewCoin("uve", sdk.NewInt(2000000))
	redAmount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// Delegate to validator1
	err := s.keeper.Delegate(s.ctx, delegatorAddr, validator1, amount)
	s.Require().NoError(err)

	// Redelegate from validator1 to validator2
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, validator1, validator2, redAmount)
	s.Require().NoError(err)

	// Delegate to validator2 (fresh delegation, not from redelegation)
	err = s.keeper.Delegate(s.ctx, delegatorAddr, validator2, amount)
	s.Require().NoError(err)

	// Try to redelegate from validator1 to validator3 (should work, validator1 hasn't received redelegation)
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, validator1, validator3, redAmount)
	// This would fail because there's an active redelegation from validator1
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "transitive")
}

// TestRedelegateMaxRedelegations tests max redelegations enforcement
func (s *DelegationOperationsTestSuite) TestRedelegateMaxRedelegations() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(10000000))
	redAmount := sdk.NewCoin("uve", sdk.NewInt(1000000))

	// Set max redelegations to 2
	params := s.keeper.GetParams(s.ctx)
	params.MaxRedelegations = 2
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Delegate to multiple validators
	for i := 1; i <= 4; i++ {
		validator := "cosmos1validator" + string(rune('0'+i)) + "0000000000000"
		err := s.keeper.Delegate(s.ctx, delegatorAddr, validator, amount)
		s.Require().NoError(err)
	}

	// First redelegation - should work
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, "cosmos1validator10000000000000", "cosmos1validatora0000000000000", redAmount)
	s.Require().NoError(err)

	// Second redelegation - should work
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, "cosmos1validator20000000000000", "cosmos1validatorb0000000000000", redAmount)
	s.Require().NoError(err)

	// Third redelegation - should fail (max reached)
	_, err = s.keeper.Redelegate(s.ctx, delegatorAddr, "cosmos1validator30000000000000", "cosmos1validatorc0000000000000", redAmount)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "max redelegations")
}

// TestCompleteUnbonding tests completing an unbonding delegation
func (s *DelegationOperationsTestSuite) TestCompleteUnbonding() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"

	// Create a mature unbonding delegation
	pastTime := s.ctx.BlockTime().Add(-1 * time.Hour)
	ubd := types.NewUnbondingDelegation(
		"ubd-complete-test",
		delegatorAddr,
		validatorAddr,
		s.ctx.BlockHeight()-100,
		pastTime,
		pastTime.Add(-21*24*time.Hour),
		"1000000",
		"1000000000000000000",
	)

	err := s.keeper.SetUnbondingDelegation(s.ctx, *ubd)
	s.Require().NoError(err)

	// Complete unbonding
	err = s.keeper.CompleteUnbonding(s.ctx, "ubd-complete-test")
	s.Require().NoError(err)

	// Unbonding should be deleted
	_, found := s.keeper.GetUnbondingDelegation(s.ctx, "ubd-complete-test")
	s.Require().False(found)
}

// TestCompleteRedelegation tests completing a redelegation
func (s *DelegationOperationsTestSuite) TestCompleteRedelegation() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	srcValidator := "cosmos1validator1111111111111111"
	dstValidator := "cosmos1validator2222222222222222"

	// Create a mature redelegation
	pastTime := s.ctx.BlockTime().Add(-1 * time.Hour)
	red := types.NewRedelegation(
		"red-complete-test",
		delegatorAddr,
		srcValidator,
		dstValidator,
		s.ctx.BlockHeight()-100,
		pastTime,
		pastTime.Add(-21*24*time.Hour),
		"1000000",
		"1000000000000000000",
	)

	err := s.keeper.SetRedelegation(s.ctx, *red)
	s.Require().NoError(err)

	// Complete redelegation
	err = s.keeper.CompleteRedelegation(s.ctx, "red-complete-test")
	s.Require().NoError(err)

	// Redelegation should be deleted
	_, found := s.keeper.GetRedelegation(s.ctx, "red-complete-test")
	s.Require().False(found)
}

// TestUnbondingPeriod tests that unbonding period is correctly applied
func (s *DelegationOperationsTestSuite) TestUnbondingPeriod() {
	delegatorAddr := "cosmos1delegator123456789abcdef"
	validatorAddr := "cosmos1validator123456789abcdef"
	amount := sdk.NewCoin("uve", sdk.NewInt(2000000))

	// Set unbonding period to 7 days
	params := s.keeper.GetParams(s.ctx)
	params.UnbondingPeriod = 7 * 24 * 60 * 60 // 7 days in seconds
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	// Delegate
	err = s.keeper.Delegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Undelegate
	completionTime, err := s.keeper.Undelegate(s.ctx, delegatorAddr, validatorAddr, amount)
	s.Require().NoError(err)

	// Verify completion time is 7 days from now
	expectedTime := s.ctx.BlockTime().Add(7 * 24 * time.Hour)
	s.Require().Equal(expectedTime.Unix(), completionTime.Unix())
}
