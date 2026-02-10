package handler

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"

	"github.com/virtengine/virtengine/x/audit/keeper"
)

// NewHandler returns a handler for "provider" type messages.
func NewHandler(keeper keeper.Keeper) baseapp.MsgServiceHandler {
	ms := NewMsgServerImpl(keeper)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case *types.MsgSignProviderAttributes:
			res, err := ms.SignProviderAttributes(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteProviderAttributes:
			res, err := ms.DeleteProviderAttributes(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		}

		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message type: %T", msg)
	}
}
