package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veidregistry/types"
)

type msgServer struct {
	keeper Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) UpsertVerifierVersion(goCtx context.Context, msg *types.MsgUpsertVerifierVersion) (*types.MsgUpsertVerifierVersionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.UpsertProposedVerifier(ctx, msg.Verifier); err != nil {
		return nil, err
	}
	return &types.MsgUpsertVerifierVersionResponse{}, nil
}

func (ms msgServer) ApproveVerifierVersion(goCtx context.Context, msg *types.MsgApproveVerifierVersion) (*types.MsgApproveVerifierVersionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.ApproveVerifier(ctx, msg.VerifierID, msg.GovernanceProposalID, msg.ActivationHeight, msg.SecurityFix); err != nil {
		return nil, err
	}
	return &types.MsgApproveVerifierVersionResponse{}, nil
}

func (ms msgServer) CancelVerifierVersion(goCtx context.Context, msg *types.MsgCancelVerifierVersion) (*types.MsgCancelVerifierVersionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.CancelVerifier(ctx, msg.VerifierID); err != nil {
		return nil, err
	}
	return &types.MsgCancelVerifierVersionResponse{}, nil
}

func (ms msgServer) RetireVerifierVersion(goCtx context.Context, msg *types.MsgRetireVerifierVersion) (*types.MsgRetireVerifierVersionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.keeper.GetAuthority() != msg.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}
	if err := ms.keeper.RetireVerifier(ctx, msg.VerifierID); err != nil {
		return nil, err
	}
	return &types.MsgRetireVerifierVersionResponse{}, nil
}

func (ms msgServer) ReportValidatorReadiness(goCtx context.Context, msg *types.MsgReportValidatorReadiness) (*types.MsgReportValidatorReadinessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keeper.SetValidatorReadiness(ctx, types.ValidatorReadiness{
		ValidatorAddress:  msg.ValidatorAddress,
		VerifierID:        msg.VerifierID,
		ConformancePassed: msg.ConformancePassed,
		ImplementationID:  msg.ImplementationID,
		Organization:      msg.Organization,
		ReportedHeight:    ctx.BlockHeight(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgReportValidatorReadinessResponse{}, nil
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
