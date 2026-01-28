// Package keeper implements the delegation module keeper tests.
//
// VE-2017: MsgServer tests for delegation module
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

// MsgServerTestSuite is the test suite for the delegation MsgServer
type MsgServerTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    Keeper
	msgServer types.MsgServer
	cdc       codec.BinaryCodec
	skey      *storetypes.KVStoreKey
}

// SetupTest sets up the test suite
func (s *MsgServerTestSuite) SetupTest() {
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
		nil, // bankKeeper - nil for unit tests
		nil, // stakingRewardsKeeper
		"cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
	)

	s.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Initialize default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	require.NoError(s.T(), err)

	// Create MsgServer
	s.msgServer = NewMsgServerImpl(s.keeper)
}

// TestMsgServerTestSuite runs the test suite
func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

// TestNewMsgServerImpl verifies MsgServer creation
func (s *MsgServerTestSuite) TestNewMsgServerImpl() {
	msgServer := NewMsgServerImpl(s.keeper)
	s.Require().NotNil(msgServer)
}

// TestDelegate tests the Delegate message handler
func (s *MsgServerTestSuite) TestDelegate() {
	delegatorAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validatorAddr := "cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn"

	// Create validator shares first
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err := s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		msg       *types.MsgDelegate
		expectErr bool
		errType   error
	}{
		{
			name: "valid delegation",
			msg: types.NewMsgDelegate(
				delegatorAddr,
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: false,
		},
		{
			name: "invalid delegator address",
			msg: types.NewMsgDelegate(
				"invalid",
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "invalid validator address",
			msg: types.NewMsgDelegate(
				delegatorAddr,
				"invalid",
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "below minimum delegation",
			msg: types.NewMsgDelegate(
				delegatorAddr,
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(100)), // Below default minimum of 1000000
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.Delegate(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}
		})
	}
}

// TestUndelegate tests the Undelegate message handler
func (s *MsgServerTestSuite) TestUndelegate() {
	delegatorAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validatorAddr := "cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn"

	// Create validator shares and delegation first
	valShares := types.NewValidatorShares(validatorAddr, s.ctx.BlockTime())
	err := s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	// Create a delegation directly to set up state
	del := types.NewDelegation(
		delegatorAddr,
		validatorAddr,
		"2000000000000000000", // 2e18 shares
		"2000000",
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
	err = s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Update validator shares to match
	_ = valShares.AddShares("2000000000000000000", "2000000", s.ctx.BlockTime())
	err = s.keeper.SetValidatorShares(s.ctx, *valShares)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		msg       *types.MsgUndelegate
		expectErr bool
	}{
		{
			name: "valid undelegation",
			msg: types.NewMsgUndelegate(
				delegatorAddr,
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: false,
		},
		{
			name: "invalid delegator address",
			msg: types.NewMsgUndelegate(
				"invalid",
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "invalid validator address",
			msg: types.NewMsgUndelegate(
				delegatorAddr,
				"invalid",
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "non-existent delegation",
			msg: types.NewMsgUndelegate(
				"cosmos1x5wgh6vwye8xqxev3q7dqq4u2c3qv0rj5qqpzs",
				validatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.Undelegate(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().Greater(resp.CompletionTime, int64(0))
			}
		})
	}
}

// TestRedelegate tests the Redelegate message handler
func (s *MsgServerTestSuite) TestRedelegate() {
	delegatorAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	srcValidatorAddr := "cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn"
	dstValidatorAddr := "cosmos1wn2fmyewy43qj3d2tpa82wq8ndehqvzzxhcm0d"

	// Create validator shares for both validators
	srcValShares := types.NewValidatorShares(srcValidatorAddr, s.ctx.BlockTime())
	err := s.keeper.SetValidatorShares(s.ctx, *srcValShares)
	s.Require().NoError(err)

	dstValShares := types.NewValidatorShares(dstValidatorAddr, s.ctx.BlockTime())
	err = s.keeper.SetValidatorShares(s.ctx, *dstValShares)
	s.Require().NoError(err)

	// Create a delegation to source validator
	del := types.NewDelegation(
		delegatorAddr,
		srcValidatorAddr,
		"2000000000000000000", // 2e18 shares
		"2000000",
		s.ctx.BlockTime(),
		s.ctx.BlockHeight(),
	)
	err = s.keeper.SetDelegation(s.ctx, *del)
	s.Require().NoError(err)

	// Update source validator shares
	_ = srcValShares.AddShares("2000000000000000000", "2000000", s.ctx.BlockTime())
	err = s.keeper.SetValidatorShares(s.ctx, *srcValShares)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		msg       *types.MsgRedelegate
		expectErr bool
	}{
		{
			name: "valid redelegation",
			msg: types.NewMsgRedelegate(
				delegatorAddr,
				srcValidatorAddr,
				dstValidatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: false,
		},
		{
			name: "invalid delegator address",
			msg: types.NewMsgRedelegate(
				"invalid",
				srcValidatorAddr,
				dstValidatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "invalid source validator address",
			msg: types.NewMsgRedelegate(
				delegatorAddr,
				"invalid",
				dstValidatorAddr,
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
		{
			name: "invalid destination validator address",
			msg: types.NewMsgRedelegate(
				delegatorAddr,
				srcValidatorAddr,
				"invalid",
				sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.Redelegate(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Require().Greater(resp.CompletionTime, int64(0))
			}
		})
	}
}

// TestClaimRewards tests the ClaimRewards message handler
func (s *MsgServerTestSuite) TestClaimRewards() {
	delegatorAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validatorAddr := "cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn"

	testCases := []struct {
		name      string
		msg       *types.MsgClaimRewards
		expectErr bool
	}{
		{
			name: "claim from validator with no rewards",
			msg: types.NewMsgClaimRewards(
				delegatorAddr,
				validatorAddr,
			),
			expectErr: false, // Should succeed but return 0 rewards
		},
		{
			name: "invalid delegator address",
			msg: types.NewMsgClaimRewards(
				"invalid",
				validatorAddr,
			),
			expectErr: true,
		},
		{
			name: "invalid validator address",
			msg: types.NewMsgClaimRewards(
				delegatorAddr,
				"invalid",
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.ClaimRewards(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}
		})
	}
}

// TestClaimAllRewards tests the ClaimAllRewards message handler
func (s *MsgServerTestSuite) TestClaimAllRewards() {
	delegatorAddr := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

	testCases := []struct {
		name      string
		msg       *types.MsgClaimAllRewards
		expectErr bool
	}{
		{
			name: "claim all with no rewards",
			msg: types.NewMsgClaimAllRewards(
				delegatorAddr,
			),
			expectErr: false, // Should succeed but return 0 rewards
		},
		{
			name: "invalid delegator address",
			msg: types.NewMsgClaimAllRewards(
				"invalid",
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.ClaimAllRewards(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}
		})
	}
}

// TestUpdateParams tests the UpdateParams message handler
func (s *MsgServerTestSuite) TestUpdateParams() {
	authorityAddr := "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
	wrongAuthorityAddr := "cosmos1x5wgh6vwye8xqxev3q7dqq4u2c3qv0rj5qqpzs"

	testCases := []struct {
		name      string
		msg       *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "valid params update from authority",
			msg: types.NewMsgUpdateParams(
				authorityAddr,
				types.Params{
					UnbondingPeriod:           14 * 24 * 60 * 60, // 14 days
					MinDelegationAmount:       2000000,
					MaxValidatorsPerDelegator: 20,
					MaxRedelegations:          10,
					ValidatorCommissionRate:   1500, // 15%
					RewardDenom:               "uve",
					StakeDenom:                "uve",
				},
			),
			expectErr: false,
		},
		{
			name: "invalid authority address",
			msg: types.NewMsgUpdateParams(
				"invalid",
				types.DefaultParams(),
			),
			expectErr: true,
		},
		{
			name: "wrong authority",
			msg: types.NewMsgUpdateParams(
				wrongAuthorityAddr,
				types.DefaultParams(),
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.UpdateParams(sdk.WrapSDKContext(s.ctx), tc.msg)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)

				// Verify params were updated
				params := s.keeper.GetParams(s.ctx)
				s.Require().Equal(tc.msg.Params.UnbondingPeriod, params.UnbondingPeriod)
			}
		})
	}
}

// TestMsgServerDelegateValidation tests message validation
func (s *MsgServerTestSuite) TestMsgServerDelegateValidation() {
	// Test that msg.ValidateBasic is properly called by the message with valid inputs
	msg := types.NewMsgDelegate(
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
		"cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn",
		sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
	)
	s.Require().NoError(msg.ValidateBasic())

	// Test with zero amount (should fail validation)
	msgZero := types.NewMsgDelegate(
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
		"cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn",
		sdk.NewCoin("uve", sdkmath.ZeroInt()),
	)
	s.Require().Error(msgZero.ValidateBasic())

	// Test with invalid delegator address
	msgInvalidDelegator := types.NewMsgDelegate(
		"invalid_address",
		"cosmos1sn6yerd80wqhwlm8ykw337vj6hurs688ucg9tn",
		sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
	)
	s.Require().Error(msgInvalidDelegator.ValidateBasic())
}
