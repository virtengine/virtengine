package handler

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

// NewHandler returns a handler for "provider" type messages.
func NewHandler(keeper keeper.IKeeper, mkeeper mkeeper.IKeeper) baseapp.MsgServiceHandler {
	ms := NewMsgServerImpl(keeper, mkeeper)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case *types.MsgCreateProvider:
			res, err := ms.CreateProvider(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUpdateProvider:
			res, err := ms.UpdateProvider(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDeleteProvider:
			res, err := ms.DeleteProvider(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized bank message type: %T", msg)
		}
	}
}
