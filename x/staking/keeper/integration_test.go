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

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
	"github.com/virtengine/virtengine/x/staking/types"
)

func TestStakingRewardLifecycleIntegration(t *testing.T) {
	skey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(skey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	authority := sdk.AccAddress([]byte("staking-authority")).String()
	keeper := NewKeeper(cdc, skey, nil, nil, nil, authority)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msgServer := NewMsgServerImpl(keeper)
	querier := Querier{Keeper: keeper}

	validatorAddr := sdk.AccAddress([]byte("validator-lifecycle")).String()
	epoch := keeper.GetCurrentEpoch(ctx)

	epochInfo := types.NewRewardEpoch(epoch, ctx.BlockHeight(), ctx.BlockTime())
	require.NoError(t, keeper.SetRewardEpoch(ctx, *epochInfo))

	_, err := msgServer.RecordPerformance(ctx, &stakingv1.MsgRecordPerformance{
		Authority:                  authority,
		ValidatorAddress:           validatorAddr,
		BlocksProposed:             8,
		BlocksSigned:               8,
		VEIDVerificationsCompleted: 3,
		VEIDVerificationScore:      95,
	})
	require.NoError(t, err)

	require.NoError(t, keeper.DistributeRewards(ctx, epoch))

	rewardResp, err := querier.ValidatorReward(ctx, &types.QueryValidatorRewardRequest{
		ValidatorAddress: validatorAddr,
		Epoch:            epoch,
	})
	require.NoError(t, err)
	require.False(t, rewardResp.Reward.TotalReward.IsZero())
}
