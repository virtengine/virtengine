package keeper

import (
	"bytes"
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

	"github.com/virtengine/virtengine/x/staking/types"
)

type mockStakingKeeper struct {
	stakeByValidator map[string]int64
	totalStake       int64
}

func (m mockStakingKeeper) GetAllValidators(_ sdk.Context) []sdk.AccAddress {
	return nil
}

func (m mockStakingKeeper) GetValidatorStake(_ sdk.Context, validatorAddr sdk.AccAddress) int64 {
	return m.stakeByValidator[validatorAddr.String()]
}

func (m mockStakingKeeper) GetTotalStake(_ sdk.Context) int64 {
	return m.totalStake
}

func (m mockStakingKeeper) SlashDelegations(_ sdk.Context, _ string, _ sdkmath.LegacyDec, _ int64) error {
	return nil
}

func (m mockStakingKeeper) DistributeValidatorRewardsToDelegators(_ sdk.Context, _ string, _ uint64, validatorReward sdk.Coins) (sdk.Coins, sdk.Coins, error) {
	return validatorReward, sdk.NewCoins(), nil
}

func TestBuildStakeDistributionUsesResolvedStakeSnapshotTotal(t *testing.T) {
	skey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(skey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	addrWithStake := sdk.AccAddress(bytes.Repeat([]byte{0x1}, 20)).String()
	addrFallback := sdk.AccAddress(bytes.Repeat([]byte{0x2}, 20)).String()

	k := NewKeeper(
		cdc,
		skey,
		nil,
		nil,
		mockStakingKeeper{
			stakeByValidator: map[string]int64{
				addrWithStake: 100,
				// addrFallback intentionally missing to trigger fallback logic.
			},
			totalStake: 100, // intentionally stale/non-inclusive
		},
		"authority",
	)

	performances := []types.ValidatorPerformance{
		{
			ValidatorAddress:          addrWithStake,
			BlocksExpected:            10,
			VEIDVerificationsExpected: 2,
			OverallScore:              8000,
		},
		{
			ValidatorAddress:          addrFallback,
			BlocksExpected:            100,
			VEIDVerificationsExpected: 10,
			OverallScore:              8000,
		},
	}

	stakes, totalStake := k.buildStakeDistribution(ctx, performances)
	require.Len(t, stakes, 2)
require.Greater(t, stakes[addrFallback], int64(0))
require.Equal(t, stakes[addrWithStake]+stakes[addrFallback], totalStake)
require.Greater(t, totalStake, int64(100))
}

func TestBuildStakeDistributionUsesNetworkTotalWhenHigher(t *testing.T) {
	skey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(skey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	addrWithStake := sdk.AccAddress(bytes.Repeat([]byte{0x3}, 20)).String()
	addrMissingPerf := sdk.AccAddress(bytes.Repeat([]byte{0x4}, 20)).String()

	k := NewKeeper(
		cdc,
		skey,
		nil,
		nil,
		mockStakingKeeper{
			stakeByValidator: map[string]int64{
				addrWithStake: 700,
				// addrMissingPerf has stake but no performance sample this epoch.
				addrMissingPerf: 300,
			},
			totalStake: 1000,
		},
		"authority",
	)

	performances := []types.ValidatorPerformance{
		{
			ValidatorAddress:          addrWithStake,
			BlocksExpected:            10,
			VEIDVerificationsExpected: 2,
			OverallScore:              8000,
		},
	}

	stakes, totalStake := k.buildStakeDistribution(ctx, performances)
	require.Len(t, stakes, 1)
	require.Equal(t, int64(700), stakes[addrWithStake])
	require.Equal(t, int64(1000), totalStake)
}

func TestBuildStakeDistributionUsesNetworkTotalWhenHigher(t *testing.T) {
	skey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(skey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	addrWithStake := sdk.AccAddress(bytes.Repeat([]byte{0x3}, 20)).String()
	addrMissingPerf := sdk.AccAddress(bytes.Repeat([]byte{0x4}, 20)).String()

	k := NewKeeper(
		cdc,
		skey,
		nil,
		nil,
		mockStakingKeeper{
			stakeByValidator: map[string]int64{
				addrWithStake: 700,
				// addrMissingPerf has stake but no performance sample this epoch.
				addrMissingPerf: 300,
			},
			totalStake: 1000,
		},
		"authority",
	)

	performances := []types.ValidatorPerformance{
		{
			ValidatorAddress:          addrWithStake,
			BlocksExpected:            10,
			VEIDVerificationsExpected: 2,
			OverallScore:              8000,
		},
	}

	stakes, totalStake := k.buildStakeDistribution(ctx, performances)
	require.Len(t, stakes, 1)
	require.Equal(t, int64(700), stakes[addrWithStake])
	require.Equal(t, int64(1000), totalStake)
}
