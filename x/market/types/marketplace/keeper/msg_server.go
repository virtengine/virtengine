package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type msgServer struct {
	keeper IKeeper
}

var _ marketplace.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the marketplace MsgServer interface.
func NewMsgServerImpl(k IKeeper) marketplace.MsgServer {
	return msgServer{keeper: k}
}

func (ms msgServer) SubmitWaldurCallback(ctx sdk.Context, msg *marketplace.MsgWaldurCallback) (*marketplace.MsgWaldurCallbackResponse, error) {
	if err := ms.keeper.ProcessWaldurCallback(ctx, msg.Callback); err != nil {
		return &marketplace.MsgWaldurCallbackResponse{
			Accepted: false,
			Message:  err.Error(),
		}, err
	}

	return &marketplace.MsgWaldurCallbackResponse{
		Accepted: true,
	}, nil
}
