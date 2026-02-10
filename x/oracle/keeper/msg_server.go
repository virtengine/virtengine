// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
)

// MsgServer implements the Oracle module MsgServer interface.
type MsgServer struct {
	keeper IKeeper
}

// NewMsgServer returns an implementation of the Oracle MsgServer interface.
func NewMsgServer(keeper IKeeper) types.MsgServer {
	return &MsgServer{keeper: keeper}
}

// AddPriceEntry implements the MsgAddPriceEntry handler.
func (m *MsgServer) AddPriceEntry(ctx context.Context, msg *types.MsgAddPriceEntry) (*types.MsgAddPriceEntryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get params to validate the signer is an authorized source
	params := m.keeper.GetParams(sdkCtx)

	// Find the source index for this signer
	sourceIdx := -1
	for i, source := range params.Sources {
		if source == msg.Signer {
			sourceIdx = i
			break
		}
	}

	if sourceIdx == -1 {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("signer %s is not an authorized price source", msg.Signer)
	}

	// Add the price entry
	if err := m.keeper.AddPriceEntry(
		sdkCtx,
		safeUint32FromInt(sourceIdx),
		msg.ID.Denom,
		msg.ID.BaseDenom,
		msg.Price,
	); err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "add_price_entry"),
			sdk.NewAttribute("signer", msg.Signer),
			sdk.NewAttribute("denom", msg.ID.Denom),
			sdk.NewAttribute("base_denom", msg.ID.BaseDenom),
			sdk.NewAttribute("price", msg.Price.Price.String()),
		),
	)

	return &types.MsgAddPriceEntryResponse{}, nil
}

// UpdateParams implements the MsgUpdateParams handler.
func (m *MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if m.keeper.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.keeper.SetParams(sdkCtx, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "update_params"),
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// Ensure MsgServer implements the MsgServer interface
var _ types.MsgServer = &MsgServer{}
