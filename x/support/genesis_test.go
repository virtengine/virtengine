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
	"github.com/virtengine/virtengine/x/support/types"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

func setupTestKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(cdc, storeKey, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now(),
		Height: 1,
	}, false, log.NewNopLogger())

	return k, ctx
}

func createTestEnvelope() encryptiontypes.MultiRecipientEnvelope {
	return encryptiontypes.MultiRecipientEnvelope{
		Version:          2,
		AlgorithmID:      "X25519-XSALSA20-POLY1305",
		AlgorithmVersion: 1,
		RecipientMode:    "specific",
		PayloadCiphertext: []byte("encrypted-content"),
		PayloadNonce:     []byte("123456789012345678901234"),
		WrappedKeys: []encryptiontypes.WrappedKeyEntry{
			{
				RecipientID: "recipient1",
				WrappedKey:  []byte("wrapped-key-1"),
			},
		},
		ClientSignature: []byte("client-signature"),
		ClientID:        "test-client",
		UserSignature:   []byte("user-signature"),
		UserPubKey:      []byte("12345678901234567890123456789012"),
	}
}

func TestInitGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	customer := sdk.AccAddress("customer1")

	genesisState := &types.GenesisState{
		Params:         types.DefaultParams(),
		TicketSequence: 100,
		Tickets: []types.SupportTicket{
			{
				TicketID:         "TKT-00000001",
				CustomerAddress:  customer.String(),
				Status:           types.TicketStatusOpen,
				Priority:         types.TicketPriorityNormal,
				Category:         "technical",
				EncryptedPayload: createTestEnvelope(),
				CreatedAt:        ctx.BlockTime(),
				UpdatedAt:        ctx.BlockTime(),
			},
		},
		Responses: []types.TicketResponse{},
	}

	// Initialize genesis
	InitGenesis(ctx, k, genesisState)

	// Verify params were set
	params := k.GetParams(ctx)
	require.Equal(t, types.DefaultParams().MaxTicketsPerCustomerPerDay, params.MaxTicketsPerCustomerPerDay)

	// Verify sequence was set
	seq := k.GetTicketSequence(ctx)
	require.Equal(t, uint64(100), seq)

	// Verify ticket was created
	ticket, found := k.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, customer.String(), ticket.CustomerAddress)
}

func TestExportGenesis(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Set up initial state
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	k.SetTicketSequence(ctx, 50)

	customer := sdk.AccAddress("customer1")
	ticket := &types.SupportTicket{
		TicketID:         "TKT-00000001",
		CustomerAddress:  customer.String(),
		Status:           types.TicketStatusOpen,
		Priority:         types.TicketPriorityNormal,
		Category:         "technical",
		EncryptedPayload: createTestEnvelope(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}
	require.NoError(t, k.CreateTicket(ctx, ticket))

	// Export genesis
	exported := ExportGenesis(ctx, k)

	// Verify export
	require.Equal(t, uint64(50), exported.TicketSequence)
	require.Len(t, exported.Tickets, 1)
	require.Equal(t, "TKT-00000001", exported.Tickets[0].TicketID)
}

func TestGenesisValidation(t *testing.T) {
	// Valid genesis
	validGenesis := types.DefaultGenesisState()
	require.NoError(t, validGenesis.Validate())

	// Invalid genesis - invalid params
	invalidGenesis := &types.GenesisState{
		Params: types.Params{
			MaxTicketsPerCustomerPerDay: 0, // Invalid
			MaxResponsesPerTicket:       100,
			MaxOpenTicketsPerCustomer:   10,
			AllowedCategories:           []string{"test"},
		},
	}
	require.Error(t, invalidGenesis.Validate())
}

func TestGenesisRoundTrip(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	// Create initial state
	customer := sdk.AccAddress("customer1")

	initialGenesis := &types.GenesisState{
		Params:         types.DefaultParams(),
		TicketSequence: 200,
		Tickets: []types.SupportTicket{
			{
				TicketID:         "TKT-00000001",
				CustomerAddress:  customer.String(),
				Status:           types.TicketStatusOpen,
				Priority:         types.TicketPriorityHigh,
				Category:         "billing",
				EncryptedPayload: createTestEnvelope(),
				CreatedAt:        ctx.BlockTime(),
				UpdatedAt:        ctx.BlockTime(),
			},
		},
		Responses: []types.TicketResponse{},
	}

	// Initialize
	InitGenesis(ctx, k, initialGenesis)

	// Export
	exported := ExportGenesis(ctx, k)

	// Verify round-trip
	require.Equal(t, initialGenesis.TicketSequence, exported.TicketSequence)
	require.Len(t, exported.Tickets, len(initialGenesis.Tickets))
	require.Equal(t, initialGenesis.Tickets[0].TicketID, exported.Tickets[0].TicketID)
	require.Equal(t, initialGenesis.Tickets[0].Priority, exported.Tickets[0].Priority)
}
