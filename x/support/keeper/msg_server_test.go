package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	"github.com/virtengine/virtengine/x/support/types"
)

func TestMsgServer_CreateTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")

	msg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}

	resp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.NotEmpty(t, resp.TicketID)

	// Verify ticket was created
	ticket, found := keeper.GetTicket(ctx, resp.TicketID)
	require.True(t, found)
	require.Equal(t, customer.String(), ticket.CustomerAddress)
	require.Equal(t, "technical", ticket.Category)
	require.Equal(t, types.TicketStatusOpen, ticket.Status)
}

func TestMsgServer_CreateTicketInvalidCategory(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")

	msg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "invalid",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}

	_, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed")
}

func TestMsgServer_AssignTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	// Set up roles
	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create ticket first
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	// Assign ticket
	assignMsg := &types.MsgAssignTicket{
		Sender:   admin.String(),
		TicketID: createResp.TicketID,
		AssignTo: agent.String(),
	}

	_, err = msgServer.AssignTicket(sdk.WrapSDKContext(ctx), assignMsg)
	require.NoError(t, err)

	// Verify assignment
	ticket, found := keeper.GetTicket(ctx, createResp.TicketID)
	require.True(t, found)
	require.Equal(t, agent.String(), ticket.AssignedTo)
	require.Equal(t, types.TicketStatusAssigned, ticket.Status)
}

func TestMsgServer_AssignTicketUnauthorized(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	nonAdmin := sdk.AccAddress("nonadmin1")

	// Only set agent role, not admin for sender
	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)

	// Create ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	// Try to assign without admin role
	assignMsg := &types.MsgAssignTicket{
		Sender:   nonAdmin.String(),
		TicketID: createResp.TicketID,
		AssignTo: agent.String(),
	}

	_, err = msgServer.AssignTicket(sdk.WrapSDKContext(ctx), assignMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMsgServer_RespondToTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create and assign ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	// Customer responds
	respondMsg := &types.MsgRespondToTicket{
		Responder:        customer.String(),
		TicketID:         createResp.TicketID,
		EncryptedPayload: createTestEnvelope(),
	}

	resp, err := msgServer.RespondToTicket(sdk.WrapSDKContext(ctx), respondMsg)
	require.NoError(t, err)
	require.Equal(t, uint32(0), resp.ResponseIndex)

	// Verify response was added
	ticket, _ := keeper.GetTicket(ctx, createResp.TicketID)
	require.Equal(t, uint32(1), ticket.ResponseCount)
}

func TestMsgServer_RespondToTicketUnauthorized(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")
	otherUser := sdk.AccAddress("other1")

	// Create ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	// Other user tries to respond
	respondMsg := &types.MsgRespondToTicket{
		Responder:        otherUser.String(),
		TicketID:         createResp.TicketID,
		EncryptedPayload: createTestEnvelope(),
	}

	_, err = msgServer.RespondToTicket(sdk.WrapSDKContext(ctx), respondMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMsgServer_ResolveTicket(t *testing.T) {
	keeper, ctx, mockRoles := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")
	agent := sdk.AccAddress("agent1")
	admin := sdk.AccAddress("admin1")

	mockRoles.SetRole(agent, rolestypes.RoleSupportAgent)
	mockRoles.SetAdmin(admin, true)

	// Create and assign ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	assignMsg := &types.MsgAssignTicket{
		Sender:   admin.String(),
		TicketID: createResp.TicketID,
		AssignTo: agent.String(),
	}
	_, err = msgServer.AssignTicket(sdk.WrapSDKContext(ctx), assignMsg)
	require.NoError(t, err)

	// Move to in progress manually (need to update status)
	ticket, _ := keeper.GetTicket(ctx, createResp.TicketID)
	ticket.Status = types.TicketStatusInProgress
	require.NoError(t, keeper.SetTicket(ctx, &ticket))

	// Resolve ticket
	resolveMsg := &types.MsgResolveTicket{
		Sender:        agent.String(),
		TicketID:      createResp.TicketID,
		ResolutionRef: "fixed-the-issue",
	}

	_, err = msgServer.ResolveTicket(sdk.WrapSDKContext(ctx), resolveMsg)
	require.NoError(t, err)

	// Verify resolution
	ticket, _ = keeper.GetTicket(ctx, createResp.TicketID)
	require.Equal(t, types.TicketStatusResolved, ticket.Status)
	require.NotNil(t, ticket.ResolvedAt)
}

func TestMsgServer_CloseTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")

	// Create ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	// Close ticket
	closeMsg := &types.MsgCloseTicket{
		Sender:   customer.String(),
		TicketID: createResp.TicketID,
		Reason:   "no longer needed",
	}

	_, err = msgServer.CloseTicket(sdk.WrapSDKContext(ctx), closeMsg)
	require.NoError(t, err)

	// Verify closure
	ticket, _ := keeper.GetTicket(ctx, createResp.TicketID)
	require.Equal(t, types.TicketStatusClosed, ticket.Status)
	require.NotNil(t, ticket.ClosedAt)
}

func TestMsgServer_ReopenTicket(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")

	// Create and close ticket
	createMsg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "technical",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	createResp, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), createMsg)
	require.NoError(t, err)

	closeMsg := &types.MsgCloseTicket{
		Sender:   customer.String(),
		TicketID: createResp.TicketID,
	}
	_, err = msgServer.CloseTicket(sdk.WrapSDKContext(ctx), closeMsg)
	require.NoError(t, err)

	// Reopen ticket
	reopenMsg := &types.MsgReopenTicket{
		Sender:   customer.String(),
		TicketID: createResp.TicketID,
		Reason:   "issue recurred",
	}

	_, err = msgServer.ReopenTicket(sdk.WrapSDKContext(ctx), reopenMsg)
	require.NoError(t, err)

	// Verify reopening
	ticket, _ := keeper.GetTicket(ctx, createResp.TicketID)
	require.Equal(t, types.TicketStatusOpen, ticket.Status)
	require.Nil(t, ticket.ClosedAt)
}

func TestMsgServer_UpdateParams(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	// Update with correct authority
	updateMsg := &types.MsgUpdateParams{
		Authority: "authority",
		Params: types.Params{
			MaxTicketsPerCustomerPerDay: 10,
			MaxResponsesPerTicket:       50,
			TicketCooldownSeconds:       120,
			AutoCloseAfterDays:          14,
			MaxOpenTicketsPerCustomer:   20,
			ReopenWindowDays:            60,
			AllowedCategories:           []string{"test", "other"},
		},
	}

	_, err := msgServer.UpdateParams(ctx, updateMsg)
	require.NoError(t, err)

	// Verify update
	params := keeper.GetParams(ctx)
	require.Equal(t, uint32(10), params.MaxTicketsPerCustomerPerDay)
}

func TestMsgServer_UpdateParamsUnauthorized(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	// Update with wrong authority
	updateMsg := &types.MsgUpdateParams{
		Authority: "wrong-authority",
		Params:    types.DefaultParams(),
	}

	_, err := msgServer.UpdateParams(sdk.WrapSDKContext(ctx), updateMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}

func TestMsgServer_RateLimiting(t *testing.T) {
	keeper, ctx, _ := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	customer := sdk.AccAddress("customer1")

	params := keeper.GetParams(ctx)

	// Create max tickets
	for i := uint32(0); i < params.MaxTicketsPerCustomerPerDay; i++ {
		msg := &types.MsgCreateTicket{
			Customer:         customer.String(),
			Category:         "technical",
			Priority:         "normal",
			EncryptedPayload: createTestEnvelope(),
		}
		_, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), msg)
		require.NoError(t, err, "ticket %d should succeed", i)
	}

	// Next one should fail due to rate limit
	msg := &types.MsgCreateTicket{
		Customer:         customer.String(),
		Category:         "billing",
		Priority:         "normal",
		EncryptedPayload: createTestEnvelope(),
	}
	_, err := msgServer.CreateTicket(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeded")
}
