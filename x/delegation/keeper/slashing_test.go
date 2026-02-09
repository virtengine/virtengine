// Package keeper implements the delegation module keeper tests.
//
// VE-922: Delegator slashing tests
package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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

type SlashingTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec
	skey   *storetypes.KVStoreKey
}

func (s *SlashingTestSuite) SetupTest() {
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

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	require.NoError(s.T(), err)
}

func TestSlashingTestSuite(t *testing.T) {
	suite.Run(t, new(SlashingTestSuite))
}

func (s *SlashingTestSuite) TestSlashDelegationsProportional() {
	validator := "cosmos1validator-slash"
	delegator1 := "cosmos1delegator-slash-1"
	delegator2 := "cosmos1delegator-slash-2"

	shares := types.NewValidatorShares(validator, s.ctx.BlockTime())
	require.NoError(s.T(), shares.AddShares("1000", "1000", s.ctx.BlockTime()))
	require.NoError(s.T(), s.keeper.SetValidatorShares(s.ctx, *shares))

	del1 := types.NewDelegation(delegator1, validator, "600", "600", s.ctx.BlockTime(), s.ctx.BlockHeight())
	del2 := types.NewDelegation(delegator2, validator, "400", "400", s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del1))
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *del2))

	fraction := sdkmath.LegacyNewDecWithPrec(1, 1) // 10%
	require.NoError(s.T(), s.keeper.SlashDelegations(s.ctx, validator, fraction, s.ctx.BlockHeight()))

	updated1, found := s.keeper.GetDelegation(s.ctx, delegator1, validator)
	s.Require().True(found)
	s.Require().Equal("540", updated1.Shares)

	updated2, found := s.keeper.GetDelegation(s.ctx, delegator2, validator)
	s.Require().True(found)
	s.Require().Equal("360", updated2.Shares)

	valShares, _ := s.keeper.GetValidatorShares(s.ctx, validator)
	s.Require().Equal("900", valShares.TotalShares)
	s.Require().Equal("900", valShares.TotalStake)

	events1 := s.keeper.GetDelegatorSlashingEvents(s.ctx, delegator1)
	s.Require().Len(events1, 1)
	s.Require().Equal("60", events1[0].SlashAmount)
}

func (s *SlashingTestSuite) TestSlashDuringRedelegation() {
	srcValidator := "cosmos1validator-src"
	dstValidator := "cosmos1validator-dst"
	delegator := "cosmos1delegator-redelegate"

	// Destination validator shares and delegation
	dstShares := types.NewValidatorShares(dstValidator, s.ctx.BlockTime())
	require.NoError(s.T(), dstShares.AddShares("1000", "1000", s.ctx.BlockTime()))
	require.NoError(s.T(), s.keeper.SetValidatorShares(s.ctx, *dstShares))

	dstDel := types.NewDelegation(delegator, dstValidator, "500", "500", s.ctx.BlockTime(), s.ctx.BlockHeight())
	require.NoError(s.T(), s.keeper.SetDelegation(s.ctx, *dstDel))

	// Redelegation entry from src to dst
	red := types.NewRedelegation(
		"red-slash-test",
		delegator,
		srcValidator,
		dstValidator,
		s.ctx.BlockHeight()-1,
		s.ctx.BlockTime().Add(24*time.Hour),
		s.ctx.BlockTime(),
		"500",
		"500",
	)
	require.NoError(s.T(), s.keeper.SetRedelegation(s.ctx, *red))

	fraction := sdkmath.LegacyNewDecWithPrec(5, 1) // 50%
	require.NoError(s.T(), s.keeper.SlashDelegations(s.ctx, srcValidator, fraction, s.ctx.BlockHeight()))

	updatedDel, found := s.keeper.GetDelegation(s.ctx, delegator, dstValidator)
	s.Require().True(found)
	s.Require().Equal("250", updatedDel.Shares)

	updatedShares, _ := s.keeper.GetValidatorShares(s.ctx, dstValidator)
	s.Require().Equal("750", updatedShares.TotalShares)
	s.Require().Equal("750", updatedShares.TotalStake)

	updatedRed, _ := s.keeper.GetRedelegation(s.ctx, "red-slash-test")
	s.Require().Equal("250", updatedRed.Entries[0].SharesDst)
	s.Require().Equal("250", updatedRed.Entries[0].InitialBalance)
}
