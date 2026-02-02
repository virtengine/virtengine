package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles/types"
)

// Error message constants
const (
	errMsgInvalidSenderAddr     = "invalid sender address"
	errMsgInvalidTargetAddr     = "invalid target address"
	errMsgAccountNotOperational = "sender account is not operational"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the roles MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// AssignRole assigns a role to an account
func (ms msgServer) AssignRole(goCtx context.Context, msg *types.MsgAssignRole) (*types.MsgAssignRoleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	target, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidTargetAddr)
	}

	role, err := types.RoleFromString(msg.Role)
	if err != nil {
		return nil, types.ErrInvalidRole.Wrap(err.Error())
	}

	// Check if sender is authorized to assign this role
	if !ms.keeper.CanAssignRole(ctx, sender, role) {
		return nil, types.ErrUnauthorized.Wrapf(
			"sender %s is not authorized to assign role %s",
			sender.String(),
			role.String(),
		)
	}

	// Check if sender's account is operational
	if !ms.keeper.IsAccountOperational(ctx, sender) {
		return nil, types.ErrAccountSuspended.Wrap(errMsgAccountNotOperational)
	}

	// Assign the role
	if err := ms.keeper.AssignRole(ctx, target, role, sender); err != nil {
		return nil, err
	}

	return &types.MsgAssignRoleResponse{}, nil
}

// RevokeRole revokes a role from an account
func (ms msgServer) RevokeRole(goCtx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	target, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidTargetAddr)
	}

	role, err := types.RoleFromString(msg.Role)
	if err != nil {
		return nil, types.ErrInvalidRole.Wrap(err.Error())
	}

	// Check if trying to revoke own role
	if sender.Equals(target) {
		params := ms.keeper.GetParams(ctx)
		if !params.AllowSelfRevoke {
			return nil, types.ErrCannotRevokeOwnRole
		}
	}

	// Check if sender is authorized to revoke this role
	if !ms.keeper.CanRevokeRole(ctx, sender, role) {
		return nil, types.ErrUnauthorized.Wrapf(
			"sender %s is not authorized to revoke role %s",
			sender.String(),
			role.String(),
		)
	}

	// Check if sender's account is operational
	if !ms.keeper.IsAccountOperational(ctx, sender) {
		return nil, types.ErrAccountSuspended.Wrap(errMsgAccountNotOperational)
	}

	// Cannot revoke GenesisAccount role from a genesis account
	if role == types.RoleGenesisAccount && ms.keeper.IsGenesisAccount(ctx, target) {
		return nil, types.ErrCannotModifyGenesisAccount.Wrap("cannot revoke genesis account role")
	}

	// Revoke the role
	if err := ms.keeper.RevokeRole(ctx, target, role, sender); err != nil {
		return nil, err
	}

	return &types.MsgRevokeRoleResponse{}, nil
}

// SetAccountState sets the state of an account
func (ms msgServer) SetAccountState(goCtx context.Context, msg *types.MsgSetAccountState) (*types.MsgSetAccountStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	target, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidTargetAddr)
	}

	state, err := types.AccountStateFromString(msg.State)
	if err != nil {
		return nil, types.ErrInvalidAccountState.Wrap(err.Error())
	}

	// Check if sender is authorized to modify account states
	if !ms.keeper.CanModifyAccountState(ctx, sender) {
		return nil, types.ErrUnauthorized.Wrap("sender is not authorized to modify account states")
	}

	// Cannot suspend self
	if sender.Equals(target) && state == types.AccountStateSuspended {
		return nil, types.ErrCannotSuspendSelf
	}

	// Cannot modify genesis account states (only other genesis accounts can)
	if ms.keeper.IsGenesisAccount(ctx, target) && !ms.keeper.IsGenesisAccount(ctx, sender) {
		return nil, types.ErrCannotModifyGenesisAccount.Wrap("only genesis accounts can modify other genesis accounts")
	}

	// Set the account state
	if err := ms.keeper.SetAccountState(ctx, target, state, msg.Reason, sender); err != nil {
		return nil, err
	}

	return &types.MsgSetAccountStateResponse{}, nil
}

// NominateAdmin nominates an account as an administrator (GenesisAccount only)
func (ms msgServer) NominateAdmin(goCtx context.Context, msg *types.MsgNominateAdmin) (*types.MsgNominateAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}

	target, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidTargetAddr)
	}

	// Only genesis accounts can nominate administrators
	if !ms.keeper.IsGenesisAccount(ctx, sender) {
		return nil, types.ErrNotGenesisAccount
	}

	// Check if sender's account is operational
	if !ms.keeper.IsAccountOperational(ctx, sender) {
		return nil, types.ErrAccountSuspended.Wrap(errMsgAccountNotOperational)
	}

	// Assign Administrator role
	if err := ms.keeper.AssignRole(ctx, target, types.RoleAdministrator, sender); err != nil {
		return nil, err
	}

	// Emit nomination event
	err = ctx.EventManager().EmitTypedEvent(&types.EventAdminNominated{
		Address:     target.String(),
		NominatedBy: sender.String(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgNominateAdminResponse{}, nil
}

// UpdateParams updates the module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority matches the module's expected authority
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	// Validate params
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// Set the new params
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
