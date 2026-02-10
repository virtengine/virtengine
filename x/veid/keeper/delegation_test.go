package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// testDelegationSetup creates a test environment for delegation tests
type testDelegationSetup struct {
	ctx              sdk.Context
	keeper           Keeper
	stateStore       store.CommitMultiStore
	delegatorAddress sdk.AccAddress
	delegateAddress  sdk.AccAddress
	delegate2Address sdk.AccAddress
}

func setupDelegationTest(t *testing.T) *testDelegationSetup {
	t.Helper()

	// Create codec
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)

	// Register cleanup to close the IAVL store and stop background pruning goroutines
	t.Cleanup(func() {
		closeStoreIfNeeded(stateStore)
	})

	// Create context with store
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	// Create keeper
	keeper := NewKeeper(cdc, storeKey, "authority")

	// Generate test addresses
	delegatorAddress := sdk.AccAddress([]byte("delegator_address___"))
	delegateAddress := sdk.AccAddress([]byte("delegate_address____"))
	delegate2Address := sdk.AccAddress([]byte("delegate2_address___"))

	return &testDelegationSetup{
		ctx:              ctx,
		keeper:           keeper,
		stateStore:       stateStore,
		delegatorAddress: delegatorAddress,
		delegateAddress:  delegateAddress,
		delegate2Address: delegate2Address,
	}
}

// ============================================================================
// Delegation Creation Tests
// ============================================================================

func TestCreateDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	// Create delegation with all permissions
	permissions := []types.DelegationPermission{
		types.PermissionViewIdentity,
		types.PermissionProveIdentity,
	}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx,
		ts.delegatorAddress,
		ts.delegateAddress,
		permissions,
		expiresAt,
		10, // max 10 uses
	)
	require.NoError(t, err)
	require.NotNil(t, delegation)

	// Verify delegation properties
	require.NotEmpty(t, delegation.DelegationID)
	require.Equal(t, ts.delegatorAddress.String(), delegation.DelegatorAddress)
	require.Equal(t, ts.delegateAddress.String(), delegation.DelegateAddress)
	require.Len(t, delegation.Permissions, 2)
	require.Equal(t, types.DelegationActive, delegation.Status)
	require.Equal(t, uint64(10), delegation.MaxUses)
	require.Equal(t, uint64(10), delegation.UsesRemaining)

	// Verify delegation can be retrieved
	retrieved, found := ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.True(t, found)
	require.Equal(t, delegation.DelegationID, retrieved.DelegationID)
}

func TestCreateDelegation_UnlimitedUses(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create with maxUses = 0 (unlimited)
	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx,
		ts.delegatorAddress,
		ts.delegateAddress,
		permissions,
		expiresAt,
		0, // unlimited
	)
	require.NoError(t, err)
	require.NotNil(t, delegation)
	require.Equal(t, uint64(0), delegation.MaxUses)
	require.Equal(t, ^uint64(0), delegation.UsesRemaining) // max uint64
}

func TestCreateDelegation_SelfDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Try to delegate to self
	_, err := ts.keeper.CreateDelegation(
		ts.ctx,
		ts.delegatorAddress,
		ts.delegatorAddress, // same address
		permissions,
		expiresAt,
		10,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot delegate to self")
}

func TestCreateDelegation_NoPermissions(t *testing.T) {
	ts := setupDelegationTest(t)

	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Try to create without permissions
	_, err := ts.keeper.CreateDelegation(
		ts.ctx,
		ts.delegatorAddress,
		ts.delegateAddress,
		[]types.DelegationPermission{}, // empty
		expiresAt,
		10,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one permission is required")
}

func TestCreateDelegation_ExpiredTime(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(-1 * time.Hour) // past time

	// Try to create with past expiration
	_, err := ts.keeper.CreateDelegation(
		ts.ctx,
		ts.delegatorAddress,
		ts.delegateAddress,
		permissions,
		expiresAt,
		10,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expires_at must be in the future")
}

// ============================================================================
// Delegation Retrieval Tests
// ============================================================================

func TestGetDelegation_NotFound(t *testing.T) {
	ts := setupDelegationTest(t)

	delegation, found := ts.keeper.GetDelegation(ts.ctx, "nonexistent_id")
	require.False(t, found)
	require.Nil(t, delegation)
}

func TestListDelegationsForDelegator(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create multiple delegations from same delegator
	_, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	_, err = ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegate2Address,
		permissions, expiresAt, 5,
	)
	require.NoError(t, err)

	// List delegations for delegator
	delegations, err := ts.keeper.ListDelegationsForDelegator(ts.ctx, ts.delegatorAddress, false)
	require.NoError(t, err)
	require.Len(t, delegations, 2)
}

func TestListDelegationsForDelegate(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create a delegation
	_, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// List delegations for delegate
	delegations, err := ts.keeper.ListDelegationsForDelegate(ts.ctx, ts.delegateAddress, false)
	require.NoError(t, err)
	require.Len(t, delegations, 1)
	require.Equal(t, ts.delegatorAddress.String(), delegations[0].DelegatorAddress)
}

func TestListDelegationsForDelegate_ActiveOnly(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create an active delegation
	_, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Create and revoke another delegation
	revoked, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegate2Address, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegate2Address, revoked.DelegationID, "testing")
	require.NoError(t, err)

	// List all delegations
	allDelegations, err := ts.keeper.ListDelegationsForDelegate(ts.ctx, ts.delegateAddress, false)
	require.NoError(t, err)
	require.Len(t, allDelegations, 2)

	// List only active delegations
	activeDelegations, err := ts.keeper.ListDelegationsForDelegate(ts.ctx, ts.delegateAddress, true)
	require.NoError(t, err)
	require.Len(t, activeDelegations, 1)
}

// ============================================================================
// Delegation Expiration Tests
// ============================================================================

func TestDelegationExpiration(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(1 * time.Hour)

	// Create delegation that expires in 1 hour
	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Verify it's active now
	require.True(t, delegation.IsActive(ts.ctx.BlockTime()))

	// Advance time past expiration
	futureCtx := ts.ctx.WithBlockTime(ts.ctx.BlockTime().Add(2 * time.Hour))

	// Retrieve delegation - status should be updated
	retrieved, found := ts.keeper.GetDelegation(futureCtx, delegation.DelegationID)
	require.True(t, found)
	require.False(t, retrieved.IsActive(futureCtx.BlockTime()))
}

func TestExpireDelegations(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	shortExpiry := ts.ctx.BlockTime().Add(1 * time.Hour)
	longExpiry := ts.ctx.BlockTime().Add(48 * time.Hour)

	// Create delegations with different expiry times
	shortDelegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, shortExpiry, 10,
	)
	require.NoError(t, err)

	_, err = ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegate2Address,
		permissions, longExpiry, 10,
	)
	require.NoError(t, err)

	// Advance time past short expiry but before long expiry
	futureCtx := ts.ctx.WithBlockTime(ts.ctx.BlockTime().Add(2 * time.Hour))

	// Run expiration
	err = ts.keeper.ExpireDelegations(futureCtx)
	require.NoError(t, err)

	// Check that short delegation is expired
	retrieved, found := ts.keeper.GetDelegation(futureCtx, shortDelegation.DelegationID)
	require.True(t, found)
	require.Equal(t, types.DelegationExpired, retrieved.Status)
}

// ============================================================================
// Delegation Use Counting Tests
// ============================================================================

func TestUseDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create delegation with 3 uses
	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 3,
	)
	require.NoError(t, err)
	require.Equal(t, uint64(3), delegation.UsesRemaining)

	// Use the delegation
	err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
	require.NoError(t, err)

	// Check uses remaining
	retrieved, found := ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.True(t, found)
	require.Equal(t, uint64(2), retrieved.UsesRemaining)

	// Use again
	err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
	require.NoError(t, err)

	retrieved, _ = ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.Equal(t, uint64(1), retrieved.UsesRemaining)

	// Use one more time
	err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
	require.NoError(t, err)

	// Delegation should now be exhausted
	retrieved, _ = ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.Equal(t, uint64(0), retrieved.UsesRemaining)
	require.Equal(t, types.DelegationExhausted, retrieved.Status)

	// Try to use again - should fail
	err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no uses remaining")
}

func TestUseDelegation_UnlimitedUses(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create delegation with unlimited uses
	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 0, // unlimited
	)
	require.NoError(t, err)

	// Use many times
	for i := 0; i < 100; i++ {
		err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
		require.NoError(t, err)
	}

	// Should still be active
	retrieved, found := ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.True(t, found)
	require.Equal(t, types.DelegationActive, retrieved.Status)
}

func TestUseDelegation_WrongPermission(t *testing.T) {
	ts := setupDelegationTest(t)

	// Create delegation with only ViewIdentity permission
	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Try to use with different permission
	err = ts.keeper.UseDelegation(ts.ctx, delegation.DelegationID, types.PermissionSignOnBehalf)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not grant")
}

// ============================================================================
// Delegation Revocation Tests
// ============================================================================

func TestRevokeDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Revoke the delegation
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegatorAddress, delegation.DelegationID, "no longer needed")
	require.NoError(t, err)

	// Check status
	retrieved, found := ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.True(t, found)
	require.Equal(t, types.DelegationRevoked, retrieved.Status)
	require.NotNil(t, retrieved.RevokedAt)
	require.Equal(t, "no longer needed", retrieved.RevocationReason)
}

func TestRevokeDelegation_Unauthorized(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Try to revoke as delegate (not delegator)
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegateAddress, delegation.DelegationID, "trying to revoke")
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the delegator can revoke")
}

func TestRevokeDelegation_AlreadyRevoked(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Revoke once
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegatorAddress, delegation.DelegationID, "first revoke")
	require.NoError(t, err)

	// Try to revoke again
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegatorAddress, delegation.DelegationID, "second revoke")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already")
}

func TestRevokeDelegation_NotFound(t *testing.T) {
	ts := setupDelegationTest(t)

	err := ts.keeper.RevokeDelegation(ts.ctx, ts.delegatorAddress, "nonexistent_id", "reason")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ============================================================================
// Delegation Permission Validation Tests
// ============================================================================

func TestValidateDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{
		types.PermissionViewIdentity,
		types.PermissionProveIdentity,
	}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Validate with granted permission
	result := ts.keeper.ValidateDelegation(ts.ctx, delegation.DelegationID, types.PermissionViewIdentity)
	require.True(t, result.Valid)
	require.NotNil(t, result.Delegation)

	// Validate with another granted permission
	result = ts.keeper.ValidateDelegation(ts.ctx, delegation.DelegationID, types.PermissionProveIdentity)
	require.True(t, result.Valid)

	// Validate with non-granted permission
	result = ts.keeper.ValidateDelegation(ts.ctx, delegation.DelegationID, types.PermissionSignOnBehalf)
	require.False(t, result.Valid)
	require.Contains(t, result.Reason, "does not grant")
}

func TestValidateDelegation_NotFound(t *testing.T) {
	ts := setupDelegationTest(t)

	result := ts.keeper.ValidateDelegation(ts.ctx, "nonexistent", types.PermissionViewIdentity)
	require.False(t, result.Valid)
	require.Contains(t, result.Reason, "not found")
}

func TestValidateDelegation_Expired(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(1 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Move time forward past expiration
	futureCtx := ts.ctx.WithBlockTime(ts.ctx.BlockTime().Add(2 * time.Hour))

	result := ts.keeper.ValidateDelegation(futureCtx, delegation.DelegationID, types.PermissionViewIdentity)
	require.False(t, result.Valid)
	require.Contains(t, result.Reason, "expired")
}

// ============================================================================
// Delegation Lookup Tests
// ============================================================================

func TestGetDelegationForDelegate(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{
		types.PermissionViewIdentity,
		types.PermissionProveIdentity,
	}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	_, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Find delegation with matching permission
	found, err := ts.keeper.GetDelegationForDelegate(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress, types.PermissionViewIdentity,
	)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Find delegation with non-matching permission should fail
	_, err = ts.keeper.GetDelegationForDelegate(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress, types.PermissionSignOnBehalf,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ============================================================================
// Delegation Deletion Tests
// ============================================================================

func TestDeleteDelegation(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	delegation, err := ts.keeper.CreateDelegation(
		ts.ctx, ts.delegatorAddress, ts.delegateAddress,
		permissions, expiresAt, 10,
	)
	require.NoError(t, err)

	// Try to delete active delegation - should fail
	err = ts.keeper.DeleteDelegation(ts.ctx, delegation.DelegationID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot delete active")

	// Revoke first
	err = ts.keeper.RevokeDelegation(ts.ctx, ts.delegatorAddress, delegation.DelegationID, "cleanup")
	require.NoError(t, err)

	// Now deletion should succeed
	err = ts.keeper.DeleteDelegation(ts.ctx, delegation.DelegationID)
	require.NoError(t, err)

	// Verify it's gone
	_, found := ts.keeper.GetDelegation(ts.ctx, delegation.DelegationID)
	require.False(t, found)
}

// ============================================================================
// Delegation Iterator Tests
// ============================================================================

func TestWithDelegations(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create several delegations
	for i := 0; i < 5; i++ {
		delegateAddr := sdk.AccAddress([]byte("delegate_addr_" + string(rune('A'+i)) + "____"))
		_, err := ts.keeper.CreateDelegation(
			ts.ctx, ts.delegatorAddress, delegateAddr,
			permissions, expiresAt, 10,
		)
		require.NoError(t, err)
	}

	// Count delegations using iterator
	count := 0
	ts.keeper.WithDelegations(ts.ctx, func(d *types.DelegationRecord) bool {
		count++
		return false // continue iteration
	})
	require.Equal(t, 5, count)
}

func TestWithDelegations_StopEarly(t *testing.T) {
	ts := setupDelegationTest(t)

	permissions := []types.DelegationPermission{types.PermissionViewIdentity}
	expiresAt := ts.ctx.BlockTime().Add(24 * time.Hour)

	// Create several delegations
	for i := 0; i < 10; i++ {
		delegateAddr := sdk.AccAddress([]byte("delegate_addr_" + string(rune('A'+i)) + "____"))
		_, err := ts.keeper.CreateDelegation(
			ts.ctx, ts.delegatorAddress, delegateAddr,
			permissions, expiresAt, 10,
		)
		require.NoError(t, err)
	}

	// Stop after 3 delegations
	count := 0
	ts.keeper.WithDelegations(ts.ctx, func(d *types.DelegationRecord) bool {
		count++
		return count >= 3 // stop after 3
	})
	require.Equal(t, 3, count)
}

// ============================================================================
// Types Tests
// ============================================================================

func TestDelegationPermission_String(t *testing.T) {
	tests := []struct {
		permission types.DelegationPermission
		expected   string
	}{
		{types.PermissionViewIdentity, "view_identity"},
		{types.PermissionProveIdentity, "prove_identity"},
		{types.PermissionSignOnBehalf, "sign_on_behalf"},
		{types.PermissionManageScopes, "manage_scopes"},
	}

	for _, tc := range tests {
		require.Equal(t, tc.expected, tc.permission.String())
	}
}

func TestParseDelegationPermission(t *testing.T) {
	tests := []struct {
		input    string
		expected types.DelegationPermission
		hasError bool
	}{
		{"view_identity", types.PermissionViewIdentity, false},
		{"prove_identity", types.PermissionProveIdentity, false},
		{"sign_on_behalf", types.PermissionSignOnBehalf, false},
		{"manage_scopes", types.PermissionManageScopes, false},
		{"invalid", 0, true},
	}

	for _, tc := range tests {
		result, err := types.ParseDelegationPermission(tc.input)
		if tc.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		}
	}
}

func TestDelegationStatus_String(t *testing.T) {
	tests := []struct {
		status   types.DelegationStatus
		expected string
	}{
		{types.DelegationActive, "active"},
		{types.DelegationExpired, "expired"},
		{types.DelegationRevoked, "revoked"},
		{types.DelegationExhausted, "exhausted"},
	}

	for _, tc := range tests {
		require.Equal(t, tc.expected, tc.status.String())
	}
}

func TestDelegationStatus_IsTerminal(t *testing.T) {
	require.False(t, types.DelegationActive.IsTerminal())
	require.False(t, types.DelegationExpired.IsTerminal())
	require.True(t, types.DelegationRevoked.IsTerminal())
	require.True(t, types.DelegationExhausted.IsTerminal())
}

func TestDelegationRecord_Validate(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		record   *types.DelegationRecord
		hasError bool
		errMsg   string
	}{
		{
			name: "valid delegation",
			record: &types.DelegationRecord{
				DelegationID:     "del_123",
				DelegatorAddress: "cosmos1abc",
				DelegateAddress:  "cosmos1def",
				Permissions:      []types.DelegationPermission{types.PermissionViewIdentity},
				ExpiresAt:        now.Add(time.Hour),
				CreatedAt:        now,
				Status:           types.DelegationActive,
			},
			hasError: false,
		},
		{
			name: "empty delegation ID",
			record: &types.DelegationRecord{
				DelegationID:     "",
				DelegatorAddress: "cosmos1abc",
				DelegateAddress:  "cosmos1def",
				Permissions:      []types.DelegationPermission{types.PermissionViewIdentity},
				ExpiresAt:        now.Add(time.Hour),
				CreatedAt:        now,
				Status:           types.DelegationActive,
			},
			hasError: true,
			errMsg:   "delegation_id cannot be empty",
		},
		{
			name: "self delegation",
			record: &types.DelegationRecord{
				DelegationID:     "del_123",
				DelegatorAddress: "cosmos1abc",
				DelegateAddress:  "cosmos1abc",
				Permissions:      []types.DelegationPermission{types.PermissionViewIdentity},
				ExpiresAt:        now.Add(time.Hour),
				CreatedAt:        now,
				Status:           types.DelegationActive,
			},
			hasError: true,
			errMsg:   "cannot delegate to self",
		},
		{
			name: "no permissions",
			record: &types.DelegationRecord{
				DelegationID:     "del_123",
				DelegatorAddress: "cosmos1abc",
				DelegateAddress:  "cosmos1def",
				Permissions:      []types.DelegationPermission{},
				ExpiresAt:        now.Add(time.Hour),
				CreatedAt:        now,
				Status:           types.DelegationActive,
			},
			hasError: true,
			errMsg:   "at least one permission is required",
		},
		{
			name: "expires before created",
			record: &types.DelegationRecord{
				DelegationID:     "del_123",
				DelegatorAddress: "cosmos1abc",
				DelegateAddress:  "cosmos1def",
				Permissions:      []types.DelegationPermission{types.PermissionViewIdentity},
				ExpiresAt:        now.Add(-time.Hour),
				CreatedAt:        now,
				Status:           types.DelegationActive,
			},
			hasError: true,
			errMsg:   "expires_at cannot be before created_at",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.hasError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
