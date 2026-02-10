package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/take/v1"
)

func TestQueryParams_Success(t *testing.T) {
	k, ctx := setupTakeKeeper(t)

	params := types.Params{
		DefaultTakeRate: 18,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 4},
			{Denom: "ufoo", Rate: 6},
		},
	}
	require.NoError(t, k.SetParams(ctx, params))

	querier := k.NewQuerier()
	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, params, resp.Params)
}

func TestQueryParams_NilRequest(t *testing.T) {
	k, ctx := setupTakeKeeper(t)

	querier := k.NewQuerier()
	resp, err := querier.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "empty request")
}

func TestQueryParams_DefaultParams(t *testing.T) {
	k, ctx := setupTakeKeeper(t)

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	querier := k.NewQuerier()
	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.DefaultParams(), resp.Params)
}
