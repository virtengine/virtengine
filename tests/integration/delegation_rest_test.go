//go:build e2e.integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
	delegationkeeper "github.com/virtengine/virtengine/x/delegation/keeper"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
)

func TestDelegationRESTGateway(t *testing.T) {
	skipIfShort(t)

	appInstance := app.Setup(app.WithChainID("virtengine-integration-1"))
	ctx := appInstance.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(time.Unix(1_700_000_000, 0).UTC())

	delegator := sdk.AccAddress([]byte("rest_delegator_01")).String()
	validator := sdk.AccAddress([]byte("rest_validator_01")).String()

	require.NoError(t, appInstance.Keepers.VirtEngine.Delegation.SetParams(ctx, delegationtypes.DefaultParams()))

	delegation := delegationtypes.NewDelegation(
		delegator,
		validator,
		"1000000000000000000",
		"1000000",
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	require.NoError(t, appInstance.Keepers.VirtEngine.Delegation.SetDelegation(ctx, *delegation))

	querier := delegationkeeper.NewQuerier(appInstance.Keepers.VirtEngine.Delegation)
	expectedPath := "/virtengine/delegation/v1/delegator/" + delegator + "/validator/" + validator

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
		}

		resp, err := querier.Delegation(sdk.WrapSDKContext(ctx), &delegationv1.QueryDelegationRequest{
			DelegatorAddress: delegator,
			ValidatorAddress: validator,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + expectedPath)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var payload delegationv1.QueryDelegationResponse
	require.NoError(t, json.Unmarshal(body, &payload))
	require.True(t, payload.Found)
	require.Equal(t, delegator, payload.Delegation.DelegatorAddress)
	require.Equal(t, validator, payload.Delegation.ValidatorAddress)
}
