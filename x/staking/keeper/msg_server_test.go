// Package keeper implements the staking module message server tests.
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

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
	"github.com/virtengine/virtengine/x/staking/types"
)

type MsgServerTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    Keeper
	msgServer stakingv1.MsgServer
	cdc       codec.BinaryCodec
	skey      *storetypes.KVStoreKey
	authority string
}

func (s *MsgServerTestSuite) SetupTest() {
	s.skey = storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(s.skey, storetypes.StoreTypeIAVL, db)
	require.NoError(s.T(), stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(registry)

	s.authority = sdk.AccAddress([]byte("staking-authority")).String()
	s.keeper = NewKeeper(
		s.cdc,
		s.skey,
		nil,
		nil,
		nil,
		s.authority,
	)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	require.NoError(s.T(), s.keeper.SetParams(s.ctx, types.DefaultParams()))
	s.msgServer = NewMsgServerImpl(s.keeper)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestUpdateParams() {
	params := s.keeper.GetParams(s.ctx)
	params.EpochLength = 250

	_, err := s.msgServer.UpdateParams(s.ctx, &stakingv1.MsgUpdateParams{
		Authority: s.authority,
		Params:    params,
	})
	s.Require().NoError(err)

	updated := s.keeper.GetParams(s.ctx)
	s.Require().Equal(uint64(250), updated.EpochLength)
}

func (s *MsgServerTestSuite) TestUpdateParamsUnauthorized() {
	_, err := s.msgServer.UpdateParams(s.ctx, &stakingv1.MsgUpdateParams{
		Authority: sdk.AccAddress([]byte("other-authority")).String(),
		Params:    s.keeper.GetParams(s.ctx),
	})
	s.Require().Error(err)
}

func (s *MsgServerTestSuite) TestSlashValidator() {
	validatorAddr := sdk.AccAddress([]byte("validator-slash")).String()

	_, err := s.msgServer.SlashValidator(s.ctx, &stakingv1.MsgSlashValidator{
		Authority:        s.authority,
		ValidatorAddress: validatorAddr,
		Reason:           types.SlashReasonDowntime,
		InfractionHeight: 90,
		Evidence:         "downtime evidence",
	})
	s.Require().NoError(err)

	records := s.keeper.GetSlashingRecordsByValidator(s.ctx, validatorAddr)
	s.Require().Len(records, 1)
}

func (s *MsgServerTestSuite) TestUnjailValidator() {
	validatorAddr := sdk.AccAddress([]byte("validator-unjail")).String()

	info := types.NewValidatorSigningInfo(validatorAddr, s.ctx.BlockHeight())
	past := s.ctx.BlockTime().Add(-time.Minute)
	info.JailedUntil = &past
	require.NoError(s.T(), s.keeper.SetValidatorSigningInfo(s.ctx, *info))

	_, err := s.msgServer.UnjailValidator(s.ctx, &stakingv1.MsgUnjailValidator{
		ValidatorAddress: validatorAddr,
	})
	s.Require().NoError(err)

	updated, found := s.keeper.GetValidatorSigningInfo(s.ctx, validatorAddr)
	s.Require().True(found)
	s.Require().Nil(updated.JailedUntil)
}

func (s *MsgServerTestSuite) TestRecordPerformance() {
	validatorAddr := sdk.AccAddress([]byte("validator-perf")).String()

	_, err := s.msgServer.RecordPerformance(s.ctx, &stakingv1.MsgRecordPerformance{
		Authority:                  s.authority,
		ValidatorAddress:           validatorAddr,
		BlocksProposed:             5,
		BlocksSigned:               5,
		VEIDVerificationsCompleted: 2,
		VEIDVerificationScore:      95,
	})
	s.Require().NoError(err)

	perf, found := s.keeper.GetValidatorPerformance(s.ctx, validatorAddr, s.keeper.GetCurrentEpoch(s.ctx))
	s.Require().True(found)
	s.Require().Equal(int64(5), perf.BlocksProposed)
	s.Require().Equal(int64(5), perf.TotalSignatures)
	s.Require().Equal(int64(2), perf.VEIDVerificationsCompleted)
}
