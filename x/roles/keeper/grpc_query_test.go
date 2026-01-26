package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/roles/keeper"
	"pkg.akt.dev/node/x/roles/types"
)

func setupQuerier(t testing.TB) (context.Context, keeper.Keeper, keeper.GRPCQuerier) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(1000, 0)}, false, log.NewNopLogger())
	k := keeper.NewKeeper(cdc, key, "authority")
	querier := keeper.GRPCQuerier{Keeper: k}

	return ctx, k, querier
}

func TestQueryAccountRoles(t *testing.T) {
	ctx, k, querier := setupQuerier(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign some roles
	err := k.AssignRole(ctx.(sdk.Context), addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)
	err = k.AssignRole(ctx.(sdk.Context), addr, types.RoleServiceProvider, assignedBy)
	require.NoError(t, err)

	// Query roles
	resp, err := querier.AccountRoles(ctx, &types.QueryAccountRolesRequest{
		Address: addr.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, addr.String(), resp.Address)
	require.Len(t, resp.Roles, 2)
}

func TestQueryAccountRolesEmpty(t *testing.T) {
	ctx, _, querier := setupQuerier(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))

	// Query roles for account with no roles
	resp, err := querier.AccountRoles(ctx, &types.QueryAccountRolesRequest{
		Address: addr.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Roles)
}

func TestQueryAccountRolesInvalidAddress(t *testing.T) {
	ctx, _, querier := setupQuerier(t)

	// Query with invalid address
	_, err := querier.AccountRoles(ctx, &types.QueryAccountRolesRequest{
		Address: "invalid",
	})
	require.Error(t, err)
}

func TestQueryRoleMembers(t *testing.T) {
	ctx, k, querier := setupQuerier(t)

	addr1 := sdk.AccAddress([]byte("test_address_1234567"))
	addr2 := sdk.AccAddress([]byte("test_address_2345678"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign same role to multiple accounts
	err := k.AssignRole(ctx.(sdk.Context), addr1, types.RoleCustomer, assignedBy)
	require.NoError(t, err)
	err = k.AssignRole(ctx.(sdk.Context), addr2, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Query role members
	resp, err := querier.RoleMembers(ctx, &types.QueryRoleMembersRequest{
		Role: types.RoleCustomer.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Members, 2)
}

func TestQueryRoleMembersInvalidRole(t *testing.T) {
	ctx, _, querier := setupQuerier(t)

	// Query with invalid role
	_, err := querier.RoleMembers(ctx, &types.QueryRoleMembersRequest{
		Role: "invalid_role",
	})
	require.Error(t, err)
}

func TestQueryAccountState(t *testing.T) {
	ctx, k, querier := setupQuerier(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	modifiedBy := sdk.AccAddress([]byte("modifier_address1234"))

	// Set account state
	err := k.SetAccountState(ctx.(sdk.Context), addr, types.AccountStateActive, "test", modifiedBy)
	require.NoError(t, err)

	// Query state
	resp, err := querier.AccountState(ctx, &types.QueryAccountStateRequest{
		Address: addr.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.AccountStateActive, resp.AccountState.State)
}

func TestQueryAccountStateDefault(t *testing.T) {
	ctx, _, querier := setupQuerier(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))

	// Query state for account without explicit state
	resp, err := querier.AccountState(ctx, &types.QueryAccountStateRequest{
		Address: addr.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// Default should be active
	require.Equal(t, types.AccountStateActive, resp.AccountState.State)
}

func TestQueryGenesisAccounts(t *testing.T) {
	ctx, k, querier := setupQuerier(t)

	addr1 := sdk.AccAddress([]byte("genesis_addr_1234567"))
	addr2 := sdk.AccAddress([]byte("genesis_addr_2345678"))

	// Add genesis accounts
	err := k.AddGenesisAccount(ctx.(sdk.Context), addr1)
	require.NoError(t, err)
	err = k.AddGenesisAccount(ctx.(sdk.Context), addr2)
	require.NoError(t, err)

	// Query genesis accounts
	resp, err := querier.GenesisAccounts(ctx, &types.QueryGenesisAccountsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Addresses, 2)
}

func TestQueryParams(t *testing.T) {
	ctx, k, querier := setupQuerier(t)

	// Set custom params
	customParams := types.Params{
		MaxRolesPerAccount: 10,
		AllowSelfRevoke:    true,
	}
	err := k.SetParams(ctx.(sdk.Context), customParams)
	require.NoError(t, err)

	// Query params
	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint32(10), resp.Params.MaxRolesPerAccount)
	require.True(t, resp.Params.AllowSelfRevoke)
}

func TestQueryNilRequest(t *testing.T) {
	ctx, _, querier := setupQuerier(t)

	// All queries should handle nil requests
	_, err := querier.AccountRoles(ctx, nil)
	require.Error(t, err)

	_, err = querier.RoleMembers(ctx, nil)
	require.Error(t, err)

	_, err = querier.AccountState(ctx, nil)
	require.Error(t, err)

	_, err = querier.GenesisAccounts(ctx, nil)
	require.Error(t, err)

	_, err = querier.Params(ctx, nil)
	require.Error(t, err)
}
