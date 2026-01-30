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

	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	"github.com/virtengine/virtengine/x/support/types"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// mockRolesKeeper is a mock implementation of RolesKeeper for testing
type mockRolesKeeper struct {
	roles  map[string][]rolestypes.Role
	admins map[string]bool
}

func newMockRolesKeeper() *mockRolesKeeper {
	return &mockRolesKeeper{
		roles:  make(map[string][]rolestypes.Role),
		admins: make(map[string]bool),
	}
}

func (m *mockRolesKeeper) HasRole(ctx sdk.Context, address sdk.AccAddress, role rolestypes.Role) bool {
	roles := m.roles[address.String()]
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func (m *mockRolesKeeper) IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	return m.admins[addr.String()]
}

func (m *mockRolesKeeper) SetRole(addr sdk.AccAddress, role rolestypes.Role) {
	m.roles[addr.String()] = append(m.roles[addr.String()], role)
}

func (m *mockRolesKeeper) SetAdmin(addr sdk.AccAddress, isAdmin bool) {
	m.admins[addr.String()] = isAdmin
}

// setupKeeper creates a test keeper with mock dependencies
func setupKeeper(t *testing.T) (Keeper, sdk.Context, *mockRolesKeeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(cdc, storeKey, "authority")

	mockRoles := newMockRolesKeeper()
	keeper.SetRolesKeeper(mockRoles)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now(),
		Height: 1,
	}, false, log.NewNopLogger())

	// Set default params
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	return keeper, ctx, mockRoles
}

// createTestEnvelope creates a minimal valid MultiRecipientEnvelope for testing
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

func TestCreateTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

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

	err := keeper.CreateTicket(ctx, ticket)
	require.NoError(t, err)

	// Verify ticket was stored
	retrieved, found := keeper.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, ticket.TicketID, retrieved.TicketID)
	require.Equal(t, ticket.CustomerAddress, retrieved.CustomerAddress)
	require.Equal(t, ticket.Status, retrieved.Status)
	require.Equal(t, ticket.Priority, retrieved.Priority)
	require.Equal(t, ticket.Category, retrieved.Category)
}

func TestCreateTicketInvalidCategory(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	
	ticket := &types.SupportTicket{
		TicketID:         "TKT-00000001",
		CustomerAddress:  customer.String(),
		Status:           types.TicketStatusOpen,
		Priority:         types.TicketPriorityNormal,
		Category:         "invalid-category",
		EncryptedPayload: createTestEnvelope(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}

	err := keeper.CreateTicket(ctx, ticket)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestAssignTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	// Set up roles
	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create ticket
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
	require.NoError(t, keeper.CreateTicket(ctx, ticket))

	// Assign ticket
	err := keeper.AssignTicket(ctx, "TKT-00000001", agent, admin)
	require.NoError(t, err)

	// Verify assignment
	retrieved, found := keeper.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, agent.String(), retrieved.AssignedTo)
	require.Equal(t, types.TicketStatusAssigned, retrieved.Status)
}

func TestResolveTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	// Set up roles
	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create and assign ticket
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
	require.NoError(t, keeper.CreateTicket(ctx, ticket))
	require.NoError(t, keeper.AssignTicket(ctx, "TKT-00000001", agent, admin))

	// Move to in progress (required state transition)
	retrieved, _ := keeper.GetTicket(ctx, "TKT-00000001")
	retrieved.Status = types.TicketStatusInProgress
	require.NoError(t, keeper.SetTicket(ctx, &retrieved))

	// Resolve ticket
	err := keeper.ResolveTicket(ctx, "TKT-00000001", agent, "issue-fixed")
	require.NoError(t, err)

	// Verify resolution
	retrieved, found := keeper.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, types.TicketStatusResolved, retrieved.Status)
	require.NotNil(t, retrieved.ResolvedAt)
	require.Equal(t, "issue-fixed", retrieved.ResolutionRef)
}

func TestCloseTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	customer := sdk.AccAddress("customer1")

	// Create ticket
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
	require.NoError(t, keeper.CreateTicket(ctx, ticket))

	// Close ticket
	err := keeper.CloseTicket(ctx, "TKT-00000001", customer, "no longer needed")
	require.NoError(t, err)

	// Verify closure
	retrieved, found := keeper.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, types.TicketStatusClosed, retrieved.Status)
	require.NotNil(t, retrieved.ClosedAt)
}

func TestReopenTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	customer := sdk.AccAddress("customer1")

	// Create and close ticket
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
	require.NoError(t, keeper.CreateTicket(ctx, ticket))
	require.NoError(t, keeper.CloseTicket(ctx, "TKT-00000001", customer, "test"))

	// Reopen ticket
	err := keeper.ReopenTicket(ctx, "TKT-00000001", customer, "issue recurred")
	require.NoError(t, err)

	// Verify reopening
	retrieved, found := keeper.GetTicket(ctx, "TKT-00000001")
	require.True(t, found)
	require.Equal(t, types.TicketStatusOpen, retrieved.Status)
	require.Nil(t, retrieved.ClosedAt)
}

func TestAddResponse(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	// Set up roles
	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create and assign ticket
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
	require.NoError(t, keeper.CreateTicket(ctx, ticket))
	require.NoError(t, keeper.AssignTicket(ctx, "TKT-00000001", agent, admin))

	// Add customer response
	response := &types.TicketResponse{
		TicketID:         "TKT-00000001",
		ResponderAddress: customer.String(),
		IsAgent:          false,
		EncryptedPayload: createTestEnvelope(),
		CreatedAt:        ctx.BlockTime(),
	}
	err := keeper.AddResponse(ctx, "TKT-00000001", response)
	require.NoError(t, err)
	require.Equal(t, uint32(0), response.ResponseIndex)

	// Verify response was stored
	retrieved, found := keeper.GetResponse(ctx, "TKT-00000001", 0)
	require.True(t, found)
	require.Equal(t, customer.String(), retrieved.ResponderAddress)
	require.False(t, retrieved.IsAgent)

	// Verify ticket response count
	ticketRetrieved, _ := keeper.GetTicket(ctx, "TKT-00000001")
	require.Equal(t, uint32(1), ticketRetrieved.ResponseCount)
}

func TestGetTicketsByCustomer(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	customer1 := sdk.AccAddress("customer1")
	customer2 := sdk.AccAddress("customer2")

	// Create tickets for customer1
	for i := 0; i < 3; i++ {
		ticket := &types.SupportTicket{
			TicketID:         keeper.GetNextTicketID(ctx),
			CustomerAddress:  customer1.String(),
			Status:           types.TicketStatusOpen,
			Priority:         types.TicketPriorityNormal,
			Category:         "technical",
			EncryptedPayload: createTestEnvelope(),
			CreatedAt:        ctx.BlockTime(),
			UpdatedAt:        ctx.BlockTime(),
		}
		require.NoError(t, keeper.CreateTicket(ctx, ticket))
	}

	// Create tickets for customer2
	for i := 0; i < 2; i++ {
		ticket := &types.SupportTicket{
			TicketID:         keeper.GetNextTicketID(ctx),
			CustomerAddress:  customer2.String(),
			Status:           types.TicketStatusOpen,
			Priority:         types.TicketPriorityNormal,
			Category:         "billing",
			EncryptedPayload: createTestEnvelope(),
			CreatedAt:        ctx.BlockTime(),
			UpdatedAt:        ctx.BlockTime(),
		}
		require.NoError(t, keeper.CreateTicket(ctx, ticket))
	}

	// Verify counts
	customer1Tickets := keeper.GetTicketsByCustomer(ctx, customer1)
	require.Len(t, customer1Tickets, 3)

	customer2Tickets := keeper.GetTicketsByCustomer(ctx, customer2)
	require.Len(t, customer2Tickets, 2)
}

func TestGetTicketsByStatus(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create open ticket
	ticket1 := &types.SupportTicket{
		TicketID:         "TKT-00000001",
		CustomerAddress:  customer.String(),
		Status:           types.TicketStatusOpen,
		Priority:         types.TicketPriorityNormal,
		Category:         "technical",
		EncryptedPayload: createTestEnvelope(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}
	require.NoError(t, keeper.CreateTicket(ctx, ticket1))

	// Create assigned ticket
	ticket2 := &types.SupportTicket{
		TicketID:         "TKT-00000002",
		CustomerAddress:  customer.String(),
		Status:           types.TicketStatusOpen,
		Priority:         types.TicketPriorityHigh,
		Category:         "billing",
		EncryptedPayload: createTestEnvelope(),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}
	require.NoError(t, keeper.CreateTicket(ctx, ticket2))
	require.NoError(t, keeper.AssignTicket(ctx, "TKT-00000002", agent, admin))

	// Verify open tickets
	openTickets := keeper.GetTicketsByStatus(ctx, types.TicketStatusOpen)
	require.Len(t, openTickets, 1)

	// Verify assigned tickets
	assignedTickets := keeper.GetTicketsByStatus(ctx, types.TicketStatusAssigned)
	require.Len(t, assignedTickets, 1)
}

func TestRateLimiting(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	customer := sdk.AccAddress("customer1")

	// Should pass initially
	err := keeper.CheckRateLimit(ctx, customer)
	require.NoError(t, err)

	// Create max tickets
	params := keeper.GetParams(ctx)
	for i := uint32(0); i < params.MaxTicketsPerCustomerPerDay; i++ {
		keeper.IncrementRateLimit(ctx, customer)
	}

	// Should fail now
	err = keeper.CheckRateLimit(ctx, customer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeded")
}

func TestCanViewTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)

	customer := sdk.AccAddress("customer1")
	provider := sdk.AccAddress("provider1")
	agent := sdk.AccAddress("agent1")
	otherUser := sdk.AccAddress("other1")
	admin := sdk.AccAddress("admin1")

	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	ticket := &types.SupportTicket{
		TicketID:        "TKT-00000001",
		CustomerAddress: customer.String(),
		ProviderAddress: provider.String(),
		AssignedTo:      agent.String(),
	}

	// Customer can view own ticket
	require.True(t, keeper.CanViewTicket(ctx, customer, ticket))

	// Provider can view related ticket
	require.True(t, keeper.CanViewTicket(ctx, provider, ticket))

	// Agent can view assigned ticket
	require.True(t, keeper.CanViewTicket(ctx, agent, ticket))

	// Admin can view any ticket
	require.True(t, keeper.CanViewTicket(ctx, admin, ticket))

	// Other user cannot view
	require.False(t, keeper.CanViewTicket(ctx, otherUser, ticket))
}

func TestParams(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	// Get default params
	params := keeper.GetParams(ctx)
	require.Equal(t, types.DefaultParams().MaxTicketsPerCustomerPerDay, params.MaxTicketsPerCustomerPerDay)

	// Update params
	newParams := types.Params{
		MaxTicketsPerCustomerPerDay: 10,
		MaxResponsesPerTicket:       50,
		TicketCooldownSeconds:       120,
		AutoCloseAfterDays:          14,
		MaxOpenTicketsPerCustomer:   20,
		ReopenWindowDays:            60,
		AllowedCategories:           []string{"test", "other"},
	}
	require.NoError(t, keeper.SetParams(ctx, newParams))

	// Verify update
	retrievedParams := keeper.GetParams(ctx)
	require.Equal(t, uint32(10), retrievedParams.MaxTicketsPerCustomerPerDay)
	require.Equal(t, uint32(50), retrievedParams.MaxResponsesPerTicket)
}

func TestTicketSequence(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)

	// Initial sequence
	seq := keeper.GetTicketSequence(ctx)
	require.Equal(t, uint64(1), seq)

	// Get next ticket ID
	id1 := keeper.GetNextTicketID(ctx)
	require.Equal(t, "TKT-00000001", id1)

	id2 := keeper.GetNextTicketID(ctx)
	require.Equal(t, "TKT-00000002", id2)

	// Verify sequence was incremented
	seq = keeper.GetTicketSequence(ctx)
	require.Equal(t, uint64(3), seq)
}
