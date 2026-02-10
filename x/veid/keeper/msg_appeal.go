// Package keeper provides VEID module keeper implementation.
//
// This file implements the message server handlers for appeal operations.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// SubmitAppeal handles the MsgSubmitAppeal message
func (ms msgServer) SubmitAppeal(goCtx context.Context, msg *types.MsgSubmitAppeal) (*types.MsgSubmitAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Submit the appeal
	appeal, err := ms.keeper.SubmitAppeal(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitAppealResponse{
		AppealId:     appeal.AppealID,
		Status:       types.AppealStatusToProto(appeal.Status),
		AppealNumber: appeal.AppealNumber,
		SubmittedAt:  appeal.SubmittedAt,
	}, nil
}

// ClaimAppeal handles the MsgClaimAppeal message
func (ms msgServer) ClaimAppeal(goCtx context.Context, msg *types.MsgClaimAppeal) (*types.MsgClaimAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Claim the appeal
	if err := ms.keeper.ClaimAppeal(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgClaimAppealResponse{
		AppealId:  msg.AppealId,
		ClaimedAt: ctx.BlockHeight(),
	}, nil
}

// ResolveAppeal handles the MsgResolveAppeal message
func (ms msgServer) ResolveAppeal(goCtx context.Context, msg *types.MsgResolveAppeal) (*types.MsgResolveAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Resolve the appeal
	if err := ms.keeper.ResolveAppeal(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgResolveAppealResponse{
		AppealId:   msg.AppealId,
		Resolution: msg.Resolution,
		ResolvedAt: ctx.BlockHeight(),
	}, nil
}

// WithdrawAppeal handles the MsgWithdrawAppeal message
func (ms msgServer) WithdrawAppeal(goCtx context.Context, msg *types.MsgWithdrawAppeal) (*types.MsgWithdrawAppealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Withdraw the appeal
	if err := ms.keeper.WithdrawAppeal(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawAppealResponse{
		AppealId:    msg.AppealId,
		WithdrawnAt: ctx.BlockHeight(),
	}, nil
}
