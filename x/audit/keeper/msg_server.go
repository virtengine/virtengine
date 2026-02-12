package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

// MsgServer implements the audit log message server
type MsgServer struct {
	keeper Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(keeper Keeper) types.MsgServiceServer {
	return &MsgServer{keeper: keeper}
}

var _ types.MsgServiceServer = &MsgServer{}

// CreateExportJob creates a new export job
func (m *MsgServer) CreateExportJob(goCtx context.Context, msg *types.MsgCreateExportJob) (*types.MsgCreateExportJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate requester address
	_, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, err
	}

	// Validate format
	if msg.Format != "json" && msg.Format != "csv" {
		return nil, types.ErrInvalidExportFormat.Wrapf("format must be json or csv, got %s", msg.Format)
	}

	// Create export job
	jobID, err := m.keeper.CreateExportJob(ctx, msg.Requester, *msg.Filter, msg.Format)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateExportJobResponse{
		JobId: jobID,
	}, nil
}

// UpdateParams updates the audit log module parameters
func (m *MsgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority (should be gov module account)
	if m.keeper.GetAuthority() != msg.Authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Set params
	err := m.keeper.SetAuditLogParams(ctx, msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// GetAuthority returns the authority address (gov module account)
func (k Keeper) GetAuthority() string {
	// For now, return empty. This will be set during keeper initialization
	// in the main app wiring
	return ""
}
