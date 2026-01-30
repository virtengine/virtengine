package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type msgServer struct {
	marketplace.UnimplementedMsgServer
	keeper IKeeper
}

var _ marketplace.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the marketplace MsgServer interface.
func NewMsgServerImpl(k IKeeper) marketplace.MsgServer {
	return msgServer{keeper: k}
}

func (ms msgServer) WaldurCallback(goCtx context.Context, msg *marketplace.MsgWaldurCallback) (*marketplace.MsgWaldurCallbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create a WaldurCallback from the message fields
	callback := &marketplace.WaldurCallback{
		ChainEntityID: msg.ResourceId,
		SignerID:      msg.Sender,
		Payload:       map[string]string{"status": msg.Status, "payload": msg.Payload},
		Signature:     msg.Signature,
	}

	if err := ms.keeper.ProcessWaldurCallback(ctx, callback); err != nil {
		return nil, err
	}

	return &marketplace.MsgWaldurCallbackResponse{}, nil
}
