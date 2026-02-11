package gaspricing

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateMinGasPrices(t *testing.T) {
	key := storetypes.NewKVStoreKey("gaspricing-test")
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, tmproto.Header{Height: 10}, false, log.NewNopLogger())

	baseMinGas, err := sdk.ParseDecCoins("0.001uvirt")
	require.NoError(t, err)
	params := DefaultParams(baseMinGas)
	params.TargetBlockUtilizationBPS = 5000
	params.AdjustmentRateBPS = 4000
	params.MaxChangeBPS = 2000
	params.CongestionThresholdBPS = 8000
	params.CongestionMultiplierBPS = 2000

	keeper := NewKeeper(key, log.NewNopLogger(), params)
	require.NoError(t, keeper.SetParams(ctx, params))

	highMinGas, _, err := keeper.UpdateMinGasPrices(ctx, 9000, 10000)
	require.NoError(t, err)
	require.True(t, decCoinsAllGTE(highMinGas, baseMinGas))

	lowMinGas, _, err := keeper.UpdateMinGasPrices(ctx, 1000, 10000)
	require.NoError(t, err)
	require.True(t, decCoinsAllGTE(lowMinGas, baseMinGas))
}
