package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// MsgServer implements the BME module MsgServer interface.
type MsgServer struct {
	keeper IKeeper
}

// NewMsgServer returns an implementation of the BME MsgServer interface.
func NewMsgServer(keeper IKeeper) types.MsgServer {
	return &MsgServer{keeper: keeper}
}

// UpdateParams implements the MsgUpdateParams handler.
func (m *MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if m.keeper.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := m.keeper.SetParams(sdkCtx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// BurnMint implements the MsgBurnMint handler.
func (m *MsgServer) BurnMint(ctx context.Context, msg *types.MsgBurnMint) (*types.MsgBurnMintResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get current params to check circuit breaker status
	params := m.keeper.GetParams(sdkCtx)
	_ = params // Will be used for circuit breaker checks

	// Create ledger record ID
	// Sequence is 0 for now - will be tracked via transient store when full implementation is added
	recordID := types.LedgerRecordID{
		Height:   sdkCtx.BlockHeight(),
		Sequence: 0,
	}

	// For now, return a pending status as actual burn/mint logic
	// requires oracle integration and bank keeper
	return &types.MsgBurnMintResponse{
		ID:     recordID,
		Status: types.LedgerRecordSatusPending,
	}, nil
}

// MintACT implements the MsgMintACT handler.
func (m *MsgServer) MintACT(ctx context.Context, msg *types.MsgMintACT) (*types.MsgMintACTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get current params to check circuit breaker status
	params := m.keeper.GetParams(sdkCtx)
	_ = params // Will be used for circuit breaker checks

	// Create ledger record ID
	recordID := types.LedgerRecordID{
		Height:   sdkCtx.BlockHeight(),
		Sequence: 0,
	}

	// For now, return a pending status as actual mint logic
	// requires oracle integration and bank keeper
	return &types.MsgMintACTResponse{
		ID:     recordID,
		Status: types.LedgerRecordSatusPending,
	}, nil
}

// BurnACT implements the MsgBurnACT handler.
func (m *MsgServer) BurnACT(ctx context.Context, msg *types.MsgBurnACT) (*types.MsgBurnACTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get current params and state
	params := m.keeper.GetParams(sdkCtx)
	_ = params // Will be used for calculations

	// Create ledger record ID
	recordID := types.LedgerRecordID{
		Height:   sdkCtx.BlockHeight(),
		Sequence: 0,
	}

	// For now, return a pending status as actual burn logic
	// requires oracle integration and bank keeper
	return &types.MsgBurnACTResponse{
		ID:     recordID,
		Status: types.LedgerRecordSatusPending,
	}, nil
}

// Ensure MsgServer implements the MsgServer interface
var _ types.MsgServer = &MsgServer{}
