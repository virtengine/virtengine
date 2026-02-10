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

	"github.com/virtengine/virtengine/x/staking/types"
)

func TestStakingQueryServer(t *testing.T) {
	skey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(skey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(
		cdc,
		skey,
		nil,
		nil,
		nil,
		sdk.AccAddress([]byte("staking-authority")).String(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	querier := Querier{Keeper: keeper}
	validatorAddr := sdk.AccAddress([]byte("validator-query")).String()
	epoch := keeper.GetCurrentEpoch(ctx)

	perf := types.NewValidatorPerformance(validatorAddr, epoch)
	perf.BlocksProposed = 4
	require.NoError(t, keeper.SetValidatorPerformance(ctx, *perf))

	reward := types.NewValidatorReward(validatorAddr, epoch)
	reward.TotalReward = sdk.NewCoins(sdk.NewInt64Coin(types.DefaultParams().RewardDenom, 100))
	require.NoError(t, keeper.SetValidatorReward(ctx, *reward))

	epochInfo := types.NewRewardEpoch(epoch, ctx.BlockHeight(), ctx.BlockTime())
	require.NoError(t, keeper.SetRewardEpoch(ctx, *epochInfo))

	info := types.NewValidatorSigningInfo(validatorAddr, ctx.BlockHeight())
	require.NoError(t, keeper.SetValidatorSigningInfo(ctx, *info))

	_, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)

	perfResp, err := querier.ValidatorPerformance(ctx, &types.QueryValidatorPerformanceRequest{
		ValidatorAddress: validatorAddr,
	})
	require.NoError(t, err)
	require.Equal(t, int64(4), perfResp.Performance.BlocksProposed)

	rewardResp, err := querier.ValidatorReward(ctx, &types.QueryValidatorRewardRequest{
		ValidatorAddress: validatorAddr,
	})
	require.NoError(t, err)
	require.False(t, rewardResp.Reward.TotalReward.IsZero())

	epochResp, err := querier.RewardEpoch(ctx, &types.QueryRewardEpochRequest{Epoch: epoch})
	require.NoError(t, err)
	require.Equal(t, epoch, epochResp.RewardEpoch.EpochNumber)

	signingResp, err := querier.SigningInfo(ctx, &types.QuerySigningInfoRequest{ValidatorAddress: validatorAddr})
	require.NoError(t, err)
	require.Equal(t, validatorAddr, signingResp.Info.ValidatorAddress)

	currentResp, err := querier.CurrentEpoch(ctx, &types.QueryCurrentEpochRequest{})
	require.NoError(t, err)
	require.Equal(t, epoch, currentResp.Epoch)
}
