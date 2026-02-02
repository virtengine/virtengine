package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/x/bme/keeper"
)

func TestMsgServerUpdateParams(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServer(k)

	// Get authority
	authority := k.GetAuthority()

	// Valid params update
	customParams := types.Params{
		CircuitBreakerWarnThreshold: 9600,
		CircuitBreakerHaltThreshold: 9100,
		MinEpochBlocks:              20,
		EpochBlocksBackoff:          15,
		MintSpreadBps:               30,
		SettleSpreadBps:             5,
	}

	resp, err := ms.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authority,
		Params:    customParams,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify params were updated
	storedParams := k.GetParams(ctx)
	require.Equal(t, customParams, storedParams)
}

func TestMsgServerUpdateParamsUnauthorized(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServer(k)

	// Unauthorized sender
	_, err := ms.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "cosmos1invalid",
		Params:    types.DefaultParams(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMsgServerBurnMint(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServer(k)

	resp, err := ms.BurnMint(ctx, &types.MsgBurnMint{
		Owner:       "cosmos1test",
		To:          "cosmos1test",
		CoinsToBurn: sdk.NewCoin("uve", math.NewInt(100)),
		DenomToMint: "uvact",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// Current stub returns pending status
	require.Equal(t, types.LedgerRecordSatusPending, resp.Status)
}

func TestMsgServerMintACT(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServer(k)

	resp, err := ms.MintACT(ctx, &types.MsgMintACT{
		Owner:       "cosmos1test",
		To:          "cosmos1test",
		CoinsToBurn: sdk.NewCoin("uve", math.NewInt(100)),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.LedgerRecordSatusPending, resp.Status)
}

func TestMsgServerBurnACT(t *testing.T) {
	k, ctx := setupKeeper(t)
	ms := keeper.NewMsgServer(k)

	resp, err := ms.BurnACT(ctx, &types.MsgBurnACT{
		Owner:       "cosmos1test",
		To:          "cosmos1test",
		CoinsToBurn: sdk.NewCoin("uvact", math.NewInt(50)),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.LedgerRecordSatusPending, resp.Status)
}
