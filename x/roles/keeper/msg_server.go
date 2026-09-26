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

	// Punitive states are reachable only through a sanction record.
	//
	// This message carries a bare reason string and no scope, duration, notice
	// or second party, so allowing it to reach Suspended or Terminated would
	// leave a single-administrator, permanent, account-wide lockout intact
	// alongside the sanction model and make that model advisory. Reactivation
	// to Active is still permitted here, because restoring access is the safe
	// direction and the reviewed revoke path is not the only way back.
	//
	// Genesis is unaffected: it calls the keeper directly, and a chain importing
	// a genesis snapshot must be able to restore an account's recorded state
	// without inventing a sanction history.
	if state == types.AccountStateSuspended || state == types.AccountStateTerminated {
		return nil, types.ErrSanctionRequired.Wrapf(
			"refusing to set %s via MsgSetAccountState for %s; impose a sanction instead",
			state, target)
	}

	// Set the account state
	if err := ms.keeper.SetAccountState(ctx, target, state, msg.Reason, sender); err != nil {
		return nil, err
	}

	return &types.MsgSetAccountStateResponse{}, nil
}

// ImposeSanction records a scoped, time-limited sanction against an account.
//
// This is the on-chain entry point to the sanction state machine. A suspension
// or termination it proposes is recorded PendingReview and has no effect until a
// second, distinct moderator confirms it via ConfirmSanction; the response
// reports the recorded status so the proposer is never told the sanction is in
// force when it is not.
func (ms msgServer) ImposeSanction(goCtx context.Context, msg *types.MsgImposeSanction) (*types.MsgImposeSanctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// ImposeSanction writes a record and re-projects the subject's account state,
	// so it runs in a cache context: a later validation failure must not leave a
	// half-written sanction behind.
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.imposeSanction(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms msgServer) imposeSanction(ctx sdk.Context, msg *types.MsgImposeSanction) (*types.MsgImposeSanctionResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}
	subject, err := sdk.AccAddressFromBech32(msg.Subject)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidTargetAddr)
	}
	scope, err := types.SanctionScopeFromString(msg.Scope)
	if err != nil {
		return nil, types.ErrInvalidSanctionScope.Wrap(err.Error())
	}
	kind, err := types.SanctionKindFromString(msg.Kind)
	if err != nil {
		return nil, types.ErrInvalidSanctionKind.Wrap(err.Error())
	}
	reasonCode, err := types.SanctionReasonCodeFromString(msg.ReasonCode)
	if err != nil {
		return nil, types.ErrInvalidSanctionReason.Wrap(err.Error())
	}
	if msg.DurationSeconds < 0 {
		return nil, types.ErrInvalidSanction.Wrap("duration_seconds must not be negative")
	}

	// A moderator may not sanction themselves: the appeal path assumes the
	// subject is a party who can answer back, and self-sanction is a coercion
	// surface the second-reviewer rule does not cover.
	if sender.Equals(subject) {
		return nil, types.ErrCannotSuspendSelf
	}

	// A genesis account may only be sanctioned by another genesis account.
	if ms.keeper.IsGenesisAccount(ctx, subject) && !ms.keeper.IsGenesisAccount(ctx, sender) {
		return nil, types.ErrCannotModifyGenesisAccount.Wrap("only genesis accounts can sanction other genesis accounts")
	}

	proposal := types.Sanction{
		Subject:       subject.String(),
		Scope:         scope,
		ScopeRef:      msg.ScopeRef,
		Kind:          kind,
		ReasonCode:    reasonCode,
		Justification: msg.Justification,
		Notice:        msg.Notice,
	}
	// A requested duration is resolved against block time, never wall clock.
	// Zero means "the kind's own default", which ImposeSanction applies.
	if msg.DurationSeconds > 0 {
		proposal.ExpiresAt = ctx.BlockTime().Unix() + msg.DurationSeconds
	}

	record, err := ms.keeper.ImposeSanction(ctx, proposal, sender)
	if err != nil {
		return nil, err
	}

	return &types.MsgImposeSanctionResponse{
		SanctionID: record.ID,
		Status:     record.Status.String(),
	}, nil
}

// ConfirmSanction applies a pending sanction as its second, distinct reviewer.
//
// This is the only on-chain path by which a suspension or termination takes
// effect, and the only way to convert an emergency hold into a reviewed
// sanction. The keeper refuses a reviewer identical to the imposing actor.
func (ms msgServer) ConfirmSanction(goCtx context.Context, msg *types.MsgConfirmSanction) (*types.MsgConfirmSanctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	if _, err := ms.confirmSanction(cacheCtx, msg); err != nil {
		return nil, err
	}
	write()
	return &types.MsgConfirmSanctionResponse{}, nil
}

func (ms msgServer) confirmSanction(ctx sdk.Context, msg *types.MsgConfirmSanction) (types.Sanction, error) {
	reviewer, err := sdk.AccAddressFromBech32(msg.Reviewer)
	if err != nil {
		return types.Sanction{}, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}
	return ms.keeper.ConfirmSanction(ctx, msg.SanctionID, reviewer, msg.ReviewedUntil)
}

// RevokeSanction clears an in-force or pending sanction.
//
// Revocation is de-escalation, so it needs one moderator-or-above actor, not
// two. It is also the corrective for a mistaken emergency hold, which binds
// immediately: without this message a mistaken hold would be unliftable until
// its window closed.
func (ms msgServer) RevokeSanction(goCtx context.Context, msg *types.MsgRevokeSanction) (*types.MsgRevokeSanctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	if _, err := ms.revokeSanction(cacheCtx, msg); err != nil {
		return nil, err
	}
	write()
	return &types.MsgRevokeSanctionResponse{}, nil
}

func (ms msgServer) revokeSanction(ctx sdk.Context, msg *types.MsgRevokeSanction) (types.Sanction, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return types.Sanction{}, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}
	return ms.keeper.RevokeSanction(ctx, msg.SanctionID, sender, msg.Reason)
}

// OpenSanctionAppeal opens an appeal against an in-force sanction.
//
// While the appeal is pending it blocks any harsher escalation against the same
// subject. Only the sanctioned party may appeal, and the keeper signs the
// message against the subject address, so a third party cannot appeal for them.
func (ms msgServer) OpenSanctionAppeal(goCtx context.Context, msg *types.MsgOpenSanctionAppeal) (*types.MsgOpenSanctionAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.openSanctionAppeal(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms msgServer) openSanctionAppeal(ctx sdk.Context, msg *types.MsgOpenSanctionAppeal) (*types.MsgOpenSanctionAppealResponse, error) {
	subject, err := sdk.AccAddressFromBech32(msg.Subject)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}
	record, err := ms.keeper.OpenAppeal(ctx, msg.SanctionID, subject, msg.Justification)
	if err != nil {
		return nil, err
	}
	return &types.MsgOpenSanctionAppealResponse{AppealID: record.ID}, nil
}

// ResolveSanctionAppeal resolves an open appeal.
//
// Granting it revokes the challenged sanction, which the account-state
// projection then reflects without any further action. The keeper refuses a
// reviewer who is the appellant or who took part in the challenged decision.
func (ms msgServer) ResolveSanctionAppeal(goCtx context.Context, msg *types.MsgResolveSanctionAppeal) (*types.MsgResolveSanctionAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	if _, err := ms.resolveSanctionAppeal(cacheCtx, msg); err != nil {
		return nil, err
	}
	write()
	return &types.MsgResolveSanctionAppealResponse{}, nil
}

func (ms msgServer) resolveSanctionAppeal(ctx sdk.Context, msg *types.MsgResolveSanctionAppeal) (types.Sanction, error) {
	reviewer, err := sdk.AccAddressFromBech32(msg.Reviewer)
	if err != nil {
		return types.Sanction{}, types.ErrInvalidAddress.Wrap(errMsgInvalidSenderAddr)
	}
	return ms.keeper.ResolveAppeal(ctx, msg.AppealID, reviewer, msg.Grant, msg.Notes)
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
