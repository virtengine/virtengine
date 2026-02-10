// Package keeper implements the staking module keeper.
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
	"github.com/virtengine/virtengine/x/staking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the staking MsgServer interface.
func NewMsgServerImpl(k Keeper) stakingv1.MsgServer {
	return &msgServer{Keeper: k}
}

var _ stakingv1.MsgServer = (*msgServer)(nil)

func (m msgServer) UpdateParams(goCtx context.Context, msg *stakingv1.MsgUpdateParams) (*stakingv1.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("invalid authority %s", msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.SetParams(ctx, types.Params(msg.Params)); err != nil {
		return nil, err
	}

	return &stakingv1.MsgUpdateParamsResponse{}, nil
}

func (m msgServer) SlashValidator(goCtx context.Context, msg *stakingv1.MsgSlashValidator) (*stakingv1.MsgSlashValidatorResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("invalid authority %s", msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := m.Keeper.SlashValidator(ctx, msg.ValidatorAddress, types.SlashReason(msg.Reason), msg.InfractionHeight, msg.Evidence)
	if err != nil {
		return nil, err
	}

	return &stakingv1.MsgSlashValidatorResponse{}, nil
}

func (m msgServer) UnjailValidator(goCtx context.Context, msg *stakingv1.MsgUnjailValidator) (*stakingv1.MsgUnjailValidatorResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.Keeper.UnjailValidator(ctx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	return &stakingv1.MsgUnjailValidatorResponse{}, nil
}

func (m msgServer) RecordPerformance(goCtx context.Context, msg *stakingv1.MsgRecordPerformance) (*stakingv1.MsgRecordPerformanceResponse, error) {
	if msg == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("invalid authority %s", msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch := m.GetCurrentEpoch(ctx)
	perf, found := m.GetValidatorPerformance(ctx, msg.ValidatorAddress, epoch)
	if !found {
		perf = *types.NewValidatorPerformance(msg.ValidatorAddress, epoch)
	}

	perf.BlocksProposed = msg.BlocksProposed
	perf.TotalSignatures = msg.BlocksSigned
	perf.VEIDVerificationsCompleted = msg.VEIDVerificationsCompleted
	perf.VEIDVerificationScore = msg.VEIDVerificationScore
	blockTime := ctx.BlockTime()
	perf.UpdatedAt = &blockTime
	types.ComputeOverallScore(&perf)

	if err := m.SetValidatorPerformance(ctx, perf); err != nil {
		return nil, err
	}

	return &stakingv1.MsgRecordPerformanceResponse{}, nil
}
