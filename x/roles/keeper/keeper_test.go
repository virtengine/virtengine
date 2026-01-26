package keeper_test

import (
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

func setupKeeper(t testing.TB) (sdk.Context, keeper.Keeper) {
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

	return ctx, k
}

func TestAssignRole(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign a role
	err := k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Verify the role is assigned
	hasRole := k.HasRole(ctx, addr, types.RoleCustomer)
	require.True(t, hasRole)

	// Verify we can get the role
	roles := k.GetAccountRoles(ctx, addr)
	require.Len(t, roles, 1)
	require.Equal(t, types.RoleCustomer, roles[0].Role)
	require.Equal(t, addr.String(), roles[0].Address)
}

func TestAssignRoleDuplicate(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign a role
	err := k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Try to assign the same role again
	err = k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.ErrorIs(t, err, types.ErrRoleAlreadyAssigned)
}

func TestAssignMultipleRoles(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign multiple roles
	err := k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	err = k.AssignRole(ctx, addr, types.RoleServiceProvider, assignedBy)
	require.NoError(t, err)

	// Verify both roles are assigned
	roles := k.GetAccountRoles(ctx, addr)
	require.Len(t, roles, 2)
}

func TestRevokeRole(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign a role
	err := k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Revoke the role
	err = k.RevokeRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Verify the role is revoked
	hasRole := k.HasRole(ctx, addr, types.RoleCustomer)
	require.False(t, hasRole)
}

func TestRevokeNonExistentRole(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	revokedBy := sdk.AccAddress([]byte("revoker_address12345"))

	// Try to revoke a role that doesn't exist
	err := k.RevokeRole(ctx, addr, types.RoleCustomer, revokedBy)
	require.ErrorIs(t, err, types.ErrRoleNotFound)
}

func TestGetRoleMembers(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr1 := sdk.AccAddress([]byte("test_address_1234567"))
	addr2 := sdk.AccAddress([]byte("test_address_2345678"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign same role to multiple accounts
	err := k.AssignRole(ctx, addr1, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	err = k.AssignRole(ctx, addr2, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Get role members
	members := k.GetRoleMembers(ctx, types.RoleCustomer)
	require.Len(t, members, 2)
}

func TestSetAccountState(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	modifiedBy := sdk.AccAddress([]byte("modifier_address1234"))

	// Set account state to active first
	err := k.SetAccountState(ctx, addr, types.AccountStateActive, "initial state", modifiedBy)
	require.NoError(t, err)

	// Suspend the account
	err = k.SetAccountState(ctx, addr, types.AccountStateSuspended, "test suspension", modifiedBy)
	require.NoError(t, err)

	// Verify the state
	state, found := k.GetAccountState(ctx, addr)
	require.True(t, found)
	require.Equal(t, types.AccountStateSuspended, state.State)
	require.Equal(t, "test suspension", state.Reason)
	require.Equal(t, types.AccountStateActive, state.PreviousState)
}

func TestInvalidStateTransition(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	modifiedBy := sdk.AccAddress([]byte("modifier_address1234"))

	// Set account to active first
	err := k.SetAccountState(ctx, addr, types.AccountStateActive, "initial", modifiedBy)
	require.NoError(t, err)

	// Terminate the account
	err = k.SetAccountState(ctx, addr, types.AccountStateTerminated, "termination", modifiedBy)
	require.NoError(t, err)

	// Try to reactivate a terminated account (should fail)
	err = k.SetAccountState(ctx, addr, types.AccountStateActive, "reactivation", modifiedBy)
	require.ErrorIs(t, err, types.ErrInvalidStateTransition)
}

func TestIsAccountOperational(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	modifiedBy := sdk.AccAddress([]byte("modifier_address1234"))

	// Account without explicit state should be operational
	require.True(t, k.IsAccountOperational(ctx, addr))

	// Set to active
	err := k.SetAccountState(ctx, addr, types.AccountStateActive, "active", modifiedBy)
	require.NoError(t, err)
	require.True(t, k.IsAccountOperational(ctx, addr))

	// Suspend
	err = k.SetAccountState(ctx, addr, types.AccountStateSuspended, "suspended", modifiedBy)
	require.NoError(t, err)
	require.False(t, k.IsAccountOperational(ctx, addr))
}

func TestGenesisAccount(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("genesis_addr_1234567"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, addr)
	require.NoError(t, err)

	// Verify it's a genesis account
	require.True(t, k.IsGenesisAccount(ctx, addr))

	// Verify it has the GenesisAccount role
	require.True(t, k.HasRole(ctx, addr, types.RoleGenesisAccount))

	// Get all genesis accounts
	accounts := k.GetGenesisAccounts(ctx)
	require.Len(t, accounts, 1)
	require.Equal(t, addr, accounts[0])
}

func TestCanAssignRole(t *testing.T) {
	ctx, k := setupKeeper(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	adminAddr := sdk.AccAddress([]byte("admin_address_123456"))
	userAddr := sdk.AccAddress([]byte("user_address_1234567"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Genesis account can assign any role
	require.True(t, k.CanAssignRole(ctx, genesisAddr, types.RoleAdministrator))
	require.True(t, k.CanAssignRole(ctx, genesisAddr, types.RoleGenesisAccount))
	require.True(t, k.CanAssignRole(ctx, genesisAddr, types.RoleCustomer))

	// Assign admin role
	err = k.AssignRole(ctx, adminAddr, types.RoleAdministrator, genesisAddr)
	require.NoError(t, err)

	// Admin can assign lower roles
	require.True(t, k.CanAssignRole(ctx, adminAddr, types.RoleCustomer))
	require.True(t, k.CanAssignRole(ctx, adminAddr, types.RoleModerator))

	// Admin cannot assign admin or genesis roles
	require.False(t, k.CanAssignRole(ctx, adminAddr, types.RoleAdministrator))
	require.False(t, k.CanAssignRole(ctx, adminAddr, types.RoleGenesisAccount))

	// Regular user cannot assign any roles
	require.False(t, k.CanAssignRole(ctx, userAddr, types.RoleCustomer))
}

func TestCanModifyAccountState(t *testing.T) {
	ctx, k := setupKeeper(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	adminAddr := sdk.AccAddress([]byte("admin_address_123456"))
	userAddr := sdk.AccAddress([]byte("user_address_1234567"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Genesis can modify
	require.True(t, k.CanModifyAccountState(ctx, genesisAddr))

	// Assign admin role
	err = k.AssignRole(ctx, adminAddr, types.RoleAdministrator, genesisAddr)
	require.NoError(t, err)

	// Admin can modify
	require.True(t, k.CanModifyAccountState(ctx, adminAddr))

	// Regular user cannot modify
	require.False(t, k.CanModifyAccountState(ctx, userAddr))
}

func TestParams(t *testing.T) {
	ctx, k := setupKeeper(t)

	// Get default params
	params := k.GetParams(ctx)
	require.Equal(t, types.DefaultParams(), params)

	// Set new params
	newParams := types.Params{
		MaxRolesPerAccount: 10,
		AllowSelfRevoke:    true,
	}
	err := k.SetParams(ctx, newParams)
	require.NoError(t, err)

	// Verify new params
	params = k.GetParams(ctx)
	require.Equal(t, uint32(10), params.MaxRolesPerAccount)
	require.True(t, params.AllowSelfRevoke)
}

func TestMaxRolesPerAccount(t *testing.T) {
	ctx, k := setupKeeper(t)

	// Set max roles to 2
	err := k.SetParams(ctx, types.Params{
		MaxRolesPerAccount: 2,
		AllowSelfRevoke:    false,
	})
	require.NoError(t, err)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign first role
	err = k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Assign second role
	err = k.AssignRole(ctx, addr, types.RoleServiceProvider, assignedBy)
	require.NoError(t, err)

	// Try to assign third role (should fail)
	err = k.AssignRole(ctx, addr, types.RoleModerator, assignedBy)
	require.Error(t, err)
}

func TestEventsEmitted(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("test_address_1234567"))
	assignedBy := sdk.AccAddress([]byte("assigner_address1234"))

	// Assign a role
	err := k.AssignRole(ctx, addr, types.RoleCustomer, assignedBy)
	require.NoError(t, err)

	// Check that event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)
}
