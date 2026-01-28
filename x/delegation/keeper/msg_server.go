// Package keeper implements the delegation module keeper.
//
// VE-2017: MsgServer implementation for delegation module
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

// Error message constants for msg_server
const (
	errMsgInvalidDelegatorAddr = "invalid delegator address"
	errMsgInvalidValidatorAddr = "invalid validator address"
	errMsgInvalidAuthorityAddr = "invalid authority address"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the delegation MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// Delegate handles delegating tokens to a validator
func (ms msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegatorAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidDelegatorAddr)
	}

	// Validate validator address
	_, err = sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidValidator.Wrap(errMsgInvalidValidatorAddr)
	}

	// Perform the delegation through the keeper
	if err := ms.keeper.Delegate(ctx, msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount); err != nil {
		return nil, err
	}

	// Emit event (keeper already emits event, but we log at msg level)
	ms.keeper.Logger(ctx).Info("delegation message processed",
		"delegator", delegatorAddr.String(),
		"validator", msg.ValidatorAddress,
		"amount", msg.Amount.String(),
	)

	return &types.MsgDelegateResponse{}, nil
}

// Undelegate handles undelegating tokens from a validator
func (ms msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegatorAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidDelegatorAddr)
	}

	// Validate validator address
	_, err = sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidValidator.Wrap(errMsgInvalidValidatorAddr)
	}

	// Perform the undelegation through the keeper
	completionTime, err := ms.keeper.Undelegate(ctx, msg.DelegatorAddress, msg.ValidatorAddress, msg.Amount)
	if err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("undelegation message processed",
		"delegator", delegatorAddr.String(),
		"validator", msg.ValidatorAddress,
		"amount", msg.Amount.String(),
		"completion_time", completionTime.String(),
	)

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime.Unix(),
	}, nil
}

// Redelegate handles redelegating tokens between validators
func (ms msgServer) Redelegate(goCtx context.Context, msg *types.MsgRedelegate) (*types.MsgRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegatorAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidDelegatorAddr)
	}

	// Validate source validator address
	_, err = sdk.AccAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid source validator address")
	}

	// Validate destination validator address
	_, err = sdk.AccAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid destination validator address")
	}

	// Perform the redelegation through the keeper
	completionTime, err := ms.keeper.Redelegate(ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress, msg.Amount)
	if err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("redelegation message processed",
		"delegator", delegatorAddr.String(),
		"src_validator", msg.ValidatorSrcAddress,
		"dst_validator", msg.ValidatorDstAddress,
		"amount", msg.Amount.String(),
		"completion_time", completionTime.String(),
	)

	return &types.MsgRedelegateResponse{
		CompletionTime: completionTime.Unix(),
	}, nil
}

// ClaimRewards handles claiming delegation rewards from a specific validator
func (ms msgServer) ClaimRewards(goCtx context.Context, msg *types.MsgClaimRewards) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegatorAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidDelegatorAddr)
	}

	// Validate validator address
	_, err = sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidValidator.Wrap(errMsgInvalidValidatorAddr)
	}

	// Claim rewards through the keeper
	rewards, err := ms.keeper.ClaimRewards(ctx, msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("claim rewards message processed",
		"delegator", delegatorAddr.String(),
		"validator", msg.ValidatorAddress,
		"amount", rewards.String(),
	)

	return &types.MsgClaimRewardsResponse{
		Amount: rewards.String(),
	}, nil
}

// ClaimAllRewards handles claiming all delegation rewards from all validators
func (ms msgServer) ClaimAllRewards(goCtx context.Context, msg *types.MsgClaimAllRewards) (*types.MsgClaimAllRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate delegator address
	delegatorAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidDelegatorAddr)
	}

	// Claim all rewards through the keeper
	rewards, err := ms.keeper.ClaimAllRewards(ctx, msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("claim all rewards message processed",
		"delegator", delegatorAddr.String(),
		"amount", rewards.String(),
	)

	return &types.MsgClaimAllRewardsResponse{
		Amount: rewards.String(),
	}, nil
}

// UpdateParams handles updating module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, types.ErrInvalidDelegator.Wrap(errMsgInvalidAuthorityAddr)
	}

	// Check authority is the governance module account
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrInvalidParams.Wrapf("invalid authority: expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	// Update parameters through the keeper
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_params",
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	ms.keeper.Logger(ctx).Info("params updated",
		"authority", msg.Authority,
	)

	return &types.MsgUpdateParamsResponse{}, nil
}
