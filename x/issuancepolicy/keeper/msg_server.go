package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

type msgServer struct {
	keeper Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) UpsertPolicy(goCtx context.Context, msg *types.MsgUpsertPolicy) (*types.MsgUpsertPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.UpsertPolicy(ctx, msg.Policy); err != nil {
		return nil, err
	}
	return &types.MsgUpsertPolicyResponse{}, nil
}

func (ms msgServer) SetActivePolicy(goCtx context.Context, msg *types.MsgSetActivePolicy) (*types.MsgSetActivePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.SetActivePolicy(ctx, msg.PolicyID); err != nil {
		return nil, err
	}
	return &types.MsgSetActivePolicyResponse{}, nil
}

func (ms msgServer) PausePolicy(goCtx context.Context, msg *types.MsgPausePolicy) (*types.MsgPausePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.PausePolicy(ctx, msg.PolicyID); err != nil {
		return nil, err
	}
	return &types.MsgPausePolicyResponse{}, nil
}

func (ms msgServer) ResumePolicy(goCtx context.Context, msg *types.MsgResumePolicy) (*types.MsgResumePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.ResumePolicy(ctx, msg.PolicyID); err != nil {
		return nil, err
	}
	return &types.MsgResumePolicyResponse{}, nil
}

func (ms msgServer) DeprecatePolicy(goCtx context.Context, msg *types.MsgDeprecatePolicy) (*types.MsgDeprecatePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.DeprecatePolicy(ctx, msg.PolicyID); err != nil {
		return nil, err
	}
	return &types.MsgDeprecatePolicyResponse{}, nil
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
