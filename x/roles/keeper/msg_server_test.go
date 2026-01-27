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

	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

func setupMsgServer(t testing.TB) (sdk.Context, keeper.Keeper, types.MsgServer) {
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
	msgServer := keeper.NewMsgServerImpl(k)

	return ctx, k, msgServer
}

func TestMsgAssignRole(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Add genesis account first
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Assign role as genesis account
	msg := &types.MsgAssignRole{
		Sender:  genesisAddr.String(),
		Address: targetAddr.String(),
		Role:    types.RoleCustomer.String(),
	}

	resp, err := msgServer.AssignRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify role is assigned
	require.True(t, k.HasRole(ctx, targetAddr, types.RoleCustomer))
}

func TestMsgAssignRoleUnauthorized(t *testing.T) {
	ctx, _, msgServer := setupMsgServer(t)

	senderAddr := sdk.AccAddress([]byte("sender_addr_12345678"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Try to assign role without authorization
	msg := &types.MsgAssignRole{
		Sender:  senderAddr.String(),
		Address: targetAddr.String(),
		Role:    types.RoleCustomer.String(),
	}

	_, err := msgServer.AssignRole(ctx, msg)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestMsgRevokeRole(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Assign role first
	err = k.AssignRole(ctx, targetAddr, types.RoleCustomer, genesisAddr)
	require.NoError(t, err)

	// Revoke the role
	msg := &types.MsgRevokeRole{
		Sender:  genesisAddr.String(),
		Address: targetAddr.String(),
		Role:    types.RoleCustomer.String(),
	}

	resp, err := msgServer.RevokeRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify role is revoked
	require.False(t, k.HasRole(ctx, targetAddr, types.RoleCustomer))
}

func TestMsgSetAccountState(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Set initial state to active
	err = k.SetAccountState(ctx, targetAddr, types.AccountStateActive, "initial", genesisAddr)
	require.NoError(t, err)

	// Suspend the account
	msg := &types.MsgSetAccountState{
		Sender:  genesisAddr.String(),
		Address: targetAddr.String(),
		State:   types.AccountStateSuspended.String(),
		Reason:  "test suspension",
	}

	resp, err := msgServer.SetAccountState(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify state is changed
	state, found := k.GetAccountState(ctx, targetAddr)
	require.True(t, found)
	require.Equal(t, types.AccountStateSuspended, state.State)
}

func TestMsgSetAccountStateCannotSuspendSelf(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Set initial state
	err = k.SetAccountState(ctx, genesisAddr, types.AccountStateActive, "initial", genesisAddr)
	require.NoError(t, err)

	// Try to suspend self
	msg := &types.MsgSetAccountState{
		Sender:  genesisAddr.String(),
		Address: genesisAddr.String(),
		State:   types.AccountStateSuspended.String(),
		Reason:  "self suspension",
	}

	_, err = msgServer.SetAccountState(ctx, msg)
	require.ErrorIs(t, err, types.ErrCannotSuspendSelf)
}

func TestMsgNominateAdmin(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Nominate admin
	msg := &types.MsgNominateAdmin{
		Sender:  genesisAddr.String(),
		Address: targetAddr.String(),
	}

	resp, err := msgServer.NominateAdmin(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify admin role is assigned
	require.True(t, k.HasRole(ctx, targetAddr, types.RoleAdministrator))
}

func TestMsgNominateAdminNotGenesisAccount(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	adminAddr := sdk.AccAddress([]byte("admin_addr_123456789"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))
	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))

	// Add genesis account and assign admin role
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	err = k.AssignRole(ctx, adminAddr, types.RoleAdministrator, genesisAddr)
	require.NoError(t, err)

	// Try to nominate admin as non-genesis
	msg := &types.MsgNominateAdmin{
		Sender:  adminAddr.String(),
		Address: targetAddr.String(),
	}

	_, err = msgServer.NominateAdmin(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotGenesisAccount)
}

func TestMsgAssignRoleSuspendedSender(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)

	genesisAddr := sdk.AccAddress([]byte("genesis_addr_1234567"))
	adminAddr := sdk.AccAddress([]byte("admin_addr_123456789"))
	targetAddr := sdk.AccAddress([]byte("target_addr_12345678"))

	// Add genesis account
	err := k.AddGenesisAccount(ctx, genesisAddr)
	require.NoError(t, err)

	// Assign admin role
	err = k.AssignRole(ctx, adminAddr, types.RoleAdministrator, genesisAddr)
	require.NoError(t, err)

	// Set admin's initial state and then suspend them
	err = k.SetAccountState(ctx, adminAddr, types.AccountStateActive, "initial", genesisAddr)
	require.NoError(t, err)
	err = k.SetAccountState(ctx, adminAddr, types.AccountStateSuspended, "suspended", genesisAddr)
	require.NoError(t, err)

	// Try to assign role as suspended admin
	msg := &types.MsgAssignRole{
		Sender:  adminAddr.String(),
		Address: targetAddr.String(),
		Role:    types.RoleCustomer.String(),
	}

	_, err = msgServer.AssignRole(ctx, msg)
	require.ErrorIs(t, err, types.ErrAccountSuspended)
}
