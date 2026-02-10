package handler

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/provider/keeper"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

type domainVerificationMsgServer interface {
	RequestDomainVerification(context.Context, *types.MsgRequestDomainVerification) (*types.MsgRequestDomainVerificationResponse, error)
	ConfirmDomainVerification(context.Context, *types.MsgConfirmDomainVerification) (*types.MsgConfirmDomainVerificationResponse, error)
	RevokeDomainVerification(context.Context, *types.MsgRevokeDomainVerification) (*types.MsgRevokeDomainVerificationResponse, error)
}

// NewHandler returns a handler for "provider" type messages.
func NewHandler(keeper keeper.IKeeper, mkeeper mkeeper.IKeeper, vkeeper veidkeeper.IKeeper, mfakeeper mfakeeper.IKeeper) baseapp.MsgServiceHandler {
	ms := NewMsgServerImpl(keeper, mkeeper, vkeeper, mfakeeper)

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
		case *types.MsgRequestDomainVerification:
			server, ok := ms.(domainVerificationMsgServer)
			if !ok {
				return nil, sdkerrors.ErrUnknownRequest.Wrapf("domain verification server unavailable for message type: %T", msg)
			}
			res, err := server.RequestDomainVerification(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgConfirmDomainVerification:
			server, ok := ms.(domainVerificationMsgServer)
			if !ok {
				return nil, sdkerrors.ErrUnknownRequest.Wrapf("domain verification server unavailable for message type: %T", msg)
			}
			res, err := server.ConfirmDomainVerification(ctx, msg)
			if err != nil {
				return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, err
			}
			return sdk.WrapServiceResult(ctx, res, nil)
		case *types.MsgRevokeDomainVerification:
			server, ok := ms.(domainVerificationMsgServer)
			if !ok {
				return nil, sdkerrors.ErrUnknownRequest.Wrapf("domain verification server unavailable for message type: %T", msg)
			}
			res, err := server.RevokeDomainVerification(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized provider message type: %T", msg)
		}
	}
}
