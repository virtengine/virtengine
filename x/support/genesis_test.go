package support

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

	"github.com/virtengine/virtengine/x/support/keeper"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // deprecated types retained for compatibility
)

func setupTestKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(cdc, storeKey, "authority", nil, nil)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now(),
		Height: 1,
	}, false, log.NewNopLogger())

	return k, ctx
}

func TestInitGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	owner := sdk.AccAddress("owner1")

	genesisState := &types.GenesisState{
		Params: types.DefaultParams(),
		ExternalRefs: []types.ExternalTicketRef{
			{
				ResourceID:       "deployment-123",
				ResourceType:     types.ResourceTypeDeployment,
				ExternalSystem:   types.ExternalSystemWaldur,
				ExternalTicketID: "WALDUR-456",
				ExternalURL:      "https://waldur.example.com/tickets/456",
				CreatedBy:        owner.String(),
			},
		},
	}

	// Initialize genesis
	InitGenesis(ctx, k, genesisState)

	// Verify params were set
	params := k.GetParams(ctx)
	require.NotEmpty(t, params.AllowedExternalSystems)

	// Verify ref was created
	ref, found := k.GetExternalRef(ctx, types.ResourceTypeDeployment, "deployment-123")
	require.True(t, found)
	require.Equal(t, owner.String(), ref.CreatedBy)
	require.Equal(t, "WALDUR-456", ref.ExternalTicketID)
}

func TestExportGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Set up initial state
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	owner := sdk.AccAddress("owner1")
	ref := &types.ExternalTicketRef{
		ResourceID:       "lease-789",
		ResourceType:     types.ResourceTypeLease,
		ExternalSystem:   types.ExternalSystemJira,
		ExternalTicketID: "JIRA-1234",
		ExternalURL:      "https://jira.example.com/browse/JIRA-1234",
		CreatedBy:        owner.String(),
	}
	require.NoError(t, k.RegisterExternalRef(ctx, ref))

	// Export genesis
	exported := ExportGenesis(ctx, k)

	// Verify export
	require.Len(t, exported.ExternalRefs, 1)
	require.Equal(t, "lease-789", exported.ExternalRefs[0].ResourceID)
	require.Equal(t, "JIRA-1234", exported.ExternalRefs[0].ExternalTicketID)
}

func TestGenesisValidation(t *testing.T) {
	// Valid genesis
	validGenesis := types.DefaultGenesisState()
	require.NoError(t, validGenesis.Validate())

	// Invalid genesis - invalid external system
	invalidGenesis := &types.GenesisState{
		Params: types.Params{
			AllowedExternalSystems: []string{"waldur", "jira"},
			AllowedExternalDomains: []string{"example.com"},
			MaxResponsesPerRequest: 10,
			DefaultRetentionPolicy: types.DefaultParams().DefaultRetentionPolicy,
		},
		ExternalRefs: []types.ExternalTicketRef{
			{
				ResourceID:       "test-123",
				ResourceType:     types.ResourceType("invalid"),
				ExternalSystem:   types.ExternalSystemWaldur,
				ExternalTicketID: "WALDUR-1",
			},
		},
	}
	require.Error(t, invalidGenesis.Validate())
}

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create initial state
	owner := sdk.AccAddress("owner1")

	initialGenesis := &types.GenesisState{
		Params: types.DefaultParams(),
		ExternalRefs: []types.ExternalTicketRef{
			{
				ResourceID:       "order-001",
				ResourceType:     types.ResourceTypeOrder,
				ExternalSystem:   types.ExternalSystemWaldur,
				ExternalTicketID: "WALDUR-999",
				ExternalURL:      "https://waldur.example.com/tickets/999",
				CreatedBy:        owner.String(),
			},
		},
	}

	// Initialize
	InitGenesis(ctx, k, initialGenesis)

	// Export
	exported := ExportGenesis(ctx, k)

	// Verify round-trip
	require.Len(t, exported.ExternalRefs, len(initialGenesis.ExternalRefs))
	require.Equal(t, initialGenesis.ExternalRefs[0].ResourceID, exported.ExternalRefs[0].ResourceID)
	require.Equal(t, initialGenesis.ExternalRefs[0].ExternalTicketID, exported.ExternalRefs[0].ExternalTicketID)
}
