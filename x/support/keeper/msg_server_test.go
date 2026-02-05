package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // Deprecated types retained for compatibility.
)

func TestMsgServer_RegisterExternalTicket(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")

	msg := &types.MsgRegisterExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "deployment-123",
		ResourceType:     string(types.ResourceTypeDeployment),
		ExternalSystem:   string(types.ExternalSystemWaldur),
		ExternalTicketID: "WALDUR-456",
		ExternalURL:      "https://waldur.example.com/tickets/456",
	}

	resp, err := msgServer.RegisterExternalTicket(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify ref was created
	ref, found := keeper.GetExternalRef(ctx, types.ResourceTypeDeployment, "deployment-123")
	require.True(t, found)
	require.Equal(t, owner.String(), ref.CreatedBy)
	require.Equal(t, "WALDUR-456", ref.ExternalTicketID)
}

func TestMsgServer_RegisterExternalTicketInvalidAddress(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	msg := &types.MsgRegisterExternalTicket{
		Sender:           "invalid-address",
		ResourceID:       "deployment-123",
		ResourceType:     string(types.ResourceTypeDeployment),
		ExternalSystem:   string(types.ExternalSystemWaldur),
		ExternalTicketID: "WALDUR-456",
	}

	_, err := msgServer.RegisterExternalTicket(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid")
}

func TestMsgServer_UpdateExternalTicket(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")

	// Register first
	registerMsg := &types.MsgRegisterExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "lease-789",
		ResourceType:     string(types.ResourceTypeLease),
		ExternalSystem:   string(types.ExternalSystemJira),
		ExternalTicketID: "JIRA-100",
		ExternalURL:      "https://jira.example.com/browse/JIRA-100",
	}
	_, err := msgServer.RegisterExternalTicket(ctx, registerMsg)
	require.NoError(t, err)

	// Update
	updateMsg := &types.MsgUpdateExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "lease-789",
		ResourceType:     string(types.ResourceTypeLease),
		ExternalTicketID: "JIRA-100-UPDATED",
		ExternalURL:      "https://jira.example.com/browse/JIRA-100-UPDATED",
	}

	resp, err := msgServer.UpdateExternalTicket(ctx, updateMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify update
	ref, found := keeper.GetExternalRef(ctx, types.ResourceTypeLease, "lease-789")
	require.True(t, found)
	require.Equal(t, "JIRA-100-UPDATED", ref.ExternalTicketID)
}

func TestMsgServer_UpdateExternalTicketUnauthorized(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")
	otherUser := sdk.AccAddress("other1")

	// Register
	registerMsg := &types.MsgRegisterExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "order-001",
		ResourceType:     string(types.ResourceTypeOrder),
		ExternalSystem:   string(types.ExternalSystemWaldur),
		ExternalTicketID: "WALDUR-001",
	}
	_, err := msgServer.RegisterExternalTicket(ctx, registerMsg)
	require.NoError(t, err)

	// Try to update with different user
	updateMsg := &types.MsgUpdateExternalTicket{
		Sender:           otherUser.String(),
		ResourceID:       "order-001",
		ResourceType:     string(types.ResourceTypeOrder),
		ExternalTicketID: "WALDUR-HACKED",
	}

	_, err = msgServer.UpdateExternalTicket(ctx, updateMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMsgServer_UpdateExternalTicketNotFound(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")

	updateMsg := &types.MsgUpdateExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "nonexistent",
		ResourceType:     string(types.ResourceTypeDeployment),
		ExternalTicketID: "WALDUR-999",
	}

	_, err := msgServer.UpdateExternalTicket(ctx, updateMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestMsgServer_RemoveExternalTicket(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")

	// Register first
	registerMsg := &types.MsgRegisterExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "provider-001",
		ResourceType:     string(types.ResourceTypeProvider),
		ExternalSystem:   string(types.ExternalSystemWaldur),
		ExternalTicketID: "WALDUR-P001",
	}
	_, err := msgServer.RegisterExternalTicket(ctx, registerMsg)
	require.NoError(t, err)

	// Remove
	removeMsg := &types.MsgRemoveExternalTicket{
		Sender:       owner.String(),
		ResourceID:   "provider-001",
		ResourceType: string(types.ResourceTypeProvider),
	}

	resp, err := msgServer.RemoveExternalTicket(ctx, removeMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify removal
	_, found := keeper.GetExternalRef(ctx, types.ResourceTypeProvider, "provider-001")
	require.False(t, found)
}

func TestMsgServer_RemoveExternalTicketUnauthorized(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	owner := sdk.AccAddress("owner1")
	otherUser := sdk.AccAddress("other1")

	// Register
	registerMsg := &types.MsgRegisterExternalTicket{
		Sender:           owner.String(),
		ResourceID:       "deployment-456",
		ResourceType:     string(types.ResourceTypeDeployment),
		ExternalSystem:   string(types.ExternalSystemJira),
		ExternalTicketID: "JIRA-456",
	}
	_, err := msgServer.RegisterExternalTicket(ctx, registerMsg)
	require.NoError(t, err)

	// Try to remove with different user
	removeMsg := &types.MsgRemoveExternalTicket{
		Sender:       otherUser.String(),
		ResourceID:   "deployment-456",
		ResourceType: string(types.ResourceTypeDeployment),
	}

	_, err = msgServer.RemoveExternalTicket(ctx, removeMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestMsgServer_UpdateParams(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	// Update with correct authority
	updateMsg := &types.MsgUpdateParams{
		Authority: "authority",
		Params: types.Params{
			AllowedExternalSystems: []string{"waldur"},
			AllowedExternalDomains: []string{"custom.example.com"},
			MaxResponsesPerRequest: 10,
			DefaultRetentionPolicy: types.DefaultParams().DefaultRetentionPolicy,
		},
	}

	_, err := msgServer.UpdateParams(ctx, updateMsg)
	require.NoError(t, err)

	// Verify update
	params := keeper.GetParams(ctx)
	require.Equal(t, []string{"waldur"}, params.AllowedExternalSystems)
}

func TestMsgServer_UpdateParamsUnauthorized(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	msgServer := NewMsgServerImpl(keeper)

	// Update with wrong authority
	updateMsg := &types.MsgUpdateParams{
		Authority: "wrong-authority",
		Params:    types.DefaultParams(),
	}

	_, err := msgServer.UpdateParams(ctx, updateMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authority")
}
