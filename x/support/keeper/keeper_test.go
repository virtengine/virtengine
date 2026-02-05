package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

// setupKeeper creates a test keeper
func setupKeeper(t *testing.T) (Keeper, sdk.Context) {
	return setupKeeperWithDeps(t, nil, nil)
}

func setupKeeperWithDeps(t *testing.T, enc EncryptionKeeper, roles RolesKeeper) (Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(cdc, storeKey, "authority", enc, roles)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now(),
		Height: 1,
	}, false, log.NewNopLogger())

	// Set default params
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	return keeper, ctx
}

type mockEncryptionKeeper struct {
	activeByKeyID   map[string]encryptiontypes.RecipientKeyRecord
	activeByAddress map[string]encryptiontypes.RecipientKeyRecord
}

func (m mockEncryptionKeeper) ValidateEnvelope(_ sdk.Context, envelope *encryptiontypes.EncryptedPayloadEnvelope) error {
	if envelope == nil {
		return types.ErrInvalidPayload.Wrap("envelope is nil")
	}
	return envelope.Validate()
}

func (m mockEncryptionKeeper) ValidateEnvelopeRecipients(_ sdk.Context, envelope *encryptiontypes.EncryptedPayloadEnvelope) ([]string, error) {
	if envelope == nil {
		return nil, types.ErrInvalidPayload.Wrap("envelope is nil")
	}
	var missing []string
	for _, keyID := range envelope.RecipientKeyIDs {
		if _, ok := m.activeByKeyID[keyID]; !ok {
			missing = append(missing, keyID)
		}
	}
	return missing, nil
}

func (m mockEncryptionKeeper) GetActiveRecipientKey(_ sdk.Context, address sdk.AccAddress) (encryptiontypes.RecipientKeyRecord, bool) {
	key, ok := m.activeByAddress[address.String()]
	return key, ok
}

type mockRolesKeeper struct {
	admins        map[string]bool
	supportAgents map[string]bool
}

func (m mockRolesKeeper) HasRole(_ sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool {
	if role == rolestypes.RoleSupportAgent {
		return m.supportAgents[address.String()]
	}
	return false
}

func (m mockRolesKeeper) IsAdmin(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.admins[addr.String()]
}

func TestRegisterExternalRef(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	ref := &types.ExternalTicketRef{
		ResourceID:       "deployment-123",
		ResourceType:     types.ResourceTypeDeployment,
		ExternalSystem:   types.ExternalSystemWaldur,
		ExternalTicketID: "WALDUR-456",
		ExternalURL:      "https://waldur.example.com/tickets/456",
		CreatedBy:        owner.String(),
	}

	err := keeper.RegisterExternalRef(ctx, ref)
	require.NoError(t, err)

	// Verify ref was stored
	retrieved, found := keeper.GetExternalRef(ctx, types.ResourceTypeDeployment, "deployment-123")
	require.True(t, found)
	require.Equal(t, ref.ResourceID, retrieved.ResourceID)
	require.Equal(t, ref.ExternalTicketID, retrieved.ExternalTicketID)
	require.Equal(t, ref.CreatedBy, retrieved.CreatedBy)
	require.NotZero(t, retrieved.CreatedAt)
}

func TestRegisterExternalRefDuplicate(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	ref := &types.ExternalTicketRef{
		ResourceID:       "deployment-123",
		ResourceType:     types.ResourceTypeDeployment,
		ExternalSystem:   types.ExternalSystemWaldur,
		ExternalTicketID: "WALDUR-456",
		CreatedBy:        owner.String(),
	}

	err := keeper.RegisterExternalRef(ctx, ref)
	require.NoError(t, err)

	// Try to register again - should fail
	err = keeper.RegisterExternalRef(ctx, ref)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrRefAlreadyExists)
}

func TestUpdateExternalRef(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	// Register first
	ref := &types.ExternalTicketRef{
		ResourceID:       "lease-789",
		ResourceType:     types.ResourceTypeLease,
		ExternalSystem:   types.ExternalSystemJira,
		ExternalTicketID: "JIRA-100",
		ExternalURL:      "https://jira.example.com/browse/JIRA-100",
		CreatedBy:        owner.String(),
	}
	require.NoError(t, keeper.RegisterExternalRef(ctx, ref))

	// Update
	updatedRef := &types.ExternalTicketRef{
		ResourceID:       "lease-789",
		ResourceType:     types.ResourceTypeLease,
		ExternalSystem:   types.ExternalSystemJira,
		ExternalTicketID: "JIRA-100-UPDATED",
		ExternalURL:      "https://jira.example.com/browse/JIRA-100-UPDATED",
		CreatedBy:        owner.String(),
	}
	err := keeper.UpdateExternalRef(ctx, updatedRef)
	require.NoError(t, err)

	// Verify update
	retrieved, found := keeper.GetExternalRef(ctx, types.ResourceTypeLease, "lease-789")
	require.True(t, found)
	require.Equal(t, "JIRA-100-UPDATED", retrieved.ExternalTicketID)
	require.NotZero(t, retrieved.UpdatedAt)
}

func TestUpdateExternalRefNotFound(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	ref := &types.ExternalTicketRef{
		ResourceID:       "nonexistent",
		ResourceType:     types.ResourceTypeDeployment,
		ExternalSystem:   types.ExternalSystemWaldur,
		ExternalTicketID: "WALDUR-999",
		CreatedBy:        owner.String(),
	}

	err := keeper.UpdateExternalRef(ctx, ref)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrRefNotFound)
}

func TestRemoveExternalRef(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	// Register first
	ref := &types.ExternalTicketRef{
		ResourceID:       "order-001",
		ResourceType:     types.ResourceTypeOrder,
		ExternalSystem:   types.ExternalSystemWaldur,
		ExternalTicketID: "WALDUR-001",
		CreatedBy:        owner.String(),
	}
	require.NoError(t, keeper.RegisterExternalRef(ctx, ref))

	// Remove
	err := keeper.RemoveExternalRef(ctx, types.ResourceTypeOrder, "order-001")
	require.NoError(t, err)

	// Verify removal
	_, found := keeper.GetExternalRef(ctx, types.ResourceTypeOrder, "order-001")
	require.False(t, found)
}

func TestRemoveExternalRefNotFound(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	err := keeper.RemoveExternalRef(ctx, types.ResourceTypeDeployment, "nonexistent")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrRefNotFound)
}

func TestGetExternalRefsByOwner(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner1 := sdk.AccAddress("owner1")
	owner2 := sdk.AccAddress("owner2")

	// Register refs for owner1
	for i := 0; i < 3; i++ {
		ref := &types.ExternalTicketRef{
			ResourceID:       "deployment-" + string(rune('a'+i)),
			ResourceType:     types.ResourceTypeDeployment,
			ExternalSystem:   types.ExternalSystemWaldur,
			ExternalTicketID: "WALDUR-" + string(rune('0'+i)),
			CreatedBy:        owner1.String(),
		}
		require.NoError(t, keeper.RegisterExternalRef(ctx, ref))
	}

	// Register refs for owner2
	for i := 0; i < 2; i++ {
		ref := &types.ExternalTicketRef{
			ResourceID:       "lease-" + string(rune('a'+i)),
			ResourceType:     types.ResourceTypeLease,
			ExternalSystem:   types.ExternalSystemJira,
			ExternalTicketID: "JIRA-" + string(rune('0'+i)),
			CreatedBy:        owner2.String(),
		}
		require.NoError(t, keeper.RegisterExternalRef(ctx, ref))
	}

	// Verify counts
	owner1Refs := keeper.GetExternalRefsByOwner(ctx, owner1)
	require.Len(t, owner1Refs, 3)

	owner2Refs := keeper.GetExternalRefsByOwner(ctx, owner2)
	require.Len(t, owner2Refs, 2)
}

func TestWithExternalRefs(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	owner := sdk.AccAddress("owner1")

	// Register some refs
	for i := 0; i < 5; i++ {
		ref := &types.ExternalTicketRef{
			ResourceID:       "resource-" + string(rune('a'+i)),
			ResourceType:     types.ResourceTypeDeployment,
			ExternalSystem:   types.ExternalSystemWaldur,
			ExternalTicketID: "WALDUR-" + string(rune('0'+i)),
			CreatedBy:        owner.String(),
		}
		require.NoError(t, keeper.RegisterExternalRef(ctx, ref))
	}

	// Count all refs
	count := 0
	keeper.WithExternalRefs(ctx, func(ref types.ExternalTicketRef) bool {
		count++
		return false
	})
	require.Equal(t, 5, count)

	// Test early exit
	count = 0
	keeper.WithExternalRefs(ctx, func(ref types.ExternalTicketRef) bool {
		count++
		return count >= 2 // Stop after 2
	})
	require.Equal(t, 2, count)
}

func TestParams(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Get default params
	params := keeper.GetParams(ctx)
	require.NotEmpty(t, params.AllowedExternalSystems)
	// AllowedExternalDomains is empty by default (configured in production)

	// Update params
	newParams := types.Params{
		AllowedExternalSystems: []string{"waldur"},
		AllowedExternalDomains: []string{"custom.example.com"},
		MaxResponsesPerRequest: 50,
		DefaultRetentionPolicy: types.DefaultParams().DefaultRetentionPolicy,
	}
	require.NoError(t, keeper.SetParams(ctx, newParams))

	// Verify update
	retrievedParams := keeper.GetParams(ctx)
	require.Equal(t, []string{"waldur"}, retrievedParams.AllowedExternalSystems)
	require.Equal(t, []string{"custom.example.com"}, retrievedParams.AllowedExternalDomains)
}

func TestGetAuthority(t *testing.T) {
	keeper, _ := setupKeeper(t)

	authority := keeper.GetAuthority()
	require.Equal(t, "authority", authority)
}
