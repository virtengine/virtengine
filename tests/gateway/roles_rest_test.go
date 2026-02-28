//go:build e2e.integration

package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	roleskeeper "github.com/virtengine/virtengine/x/roles/keeper"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
)

func TestRolesRESTGatewayRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping gateway integration test in short mode")
	}

	appInstance := app.Setup(app.WithChainID("virtengine-gateway-roles-1"))
	ctx := appInstance.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(time.Unix(1_700_000_100, 0).UTC())

	accountAddr := sdk.AccAddress([]byte("roles_rest_account_01"))
	address := accountAddr.String()
	assignedBy := sdk.AccAddress([]byte("roles_rest_admin_01"))

	require.NoError(t, appInstance.Keepers.VirtEngine.Roles.AssignRole(
		ctx,
		accountAddr,
		rolestypes.RoleAdministrator,
		assignedBy,
	))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, appInstance.InterfaceRegistry())
	rolestypes.RegisterQueryServer(queryHelper, roleskeeper.GRPCQuerier{Keeper: appInstance.Keepers.VirtEngine.Roles})

	mux := runtime.NewServeMux()
	require.NoError(t, rolestypes.RegisterQueryHandlerClient(context.Background(), mux, rolestypes.NewQueryClient(queryHelper)))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/virtengine/roles/v1/account/" + address + "/roles")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var accountRoles map[string]any
	require.NoError(t, json.Unmarshal(body, &accountRoles))
	require.Equal(t, address, accountRoles["address"])

	roleEntries, ok := accountRoles["roles"].([]any)
	require.True(t, ok)
	require.Len(t, roleEntries, 1)

	roleAssignment, ok := roleEntries[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, address, roleAssignment["address"])

	hasRoleResp, err := http.Get(server.URL + "/virtengine/roles/v1/account/" + address + "/has_role/administrator")
	require.NoError(t, err)
	defer hasRoleResp.Body.Close()
	require.Equal(t, http.StatusOK, hasRoleResp.StatusCode)

	hasRoleBody, err := io.ReadAll(hasRoleResp.Body)
	require.NoError(t, err)

	var hasRole map[string]any
	require.NoError(t, json.Unmarshal(hasRoleBody, &hasRole))
	require.Equal(t, true, firstPresent(hasRole, "has_role", "hasRole"))

	assignment, ok := firstPresent(hasRole, "assignment").(map[string]any)
	require.True(t, ok)
	require.Equal(t, address, assignment["address"])
}

func firstPresent(payload map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			return value
		}
	}
	return nil
}
