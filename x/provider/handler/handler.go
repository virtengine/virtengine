package handler

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogoproto "github.com/cosmos/gogoproto/proto"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	"github.com/virtengine/virtengine/x/provider/keeper"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

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

		case *types.MsgGenerateDomainVerificationToken:
			res, err := ms.GenerateDomainVerificationToken(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgVerifyProviderDomain:
			res, err := ms.VerifyProviderDomain(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRequestDomainVerification:
			res, err := ms.RequestDomainVerification(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgConfirmDomainVerification:
			res, err := ms.ConfirmDomainVerification(ctx, msg)
			if err != nil && res != nil {
				wrapped, wrapErr := wrapServiceResult(ctx, res)
				if wrapErr != nil {
					return nil, wrapErr
				}
				return wrapped, err
			}
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgRevokeDomainVerification:
			res, err := ms.RevokeDomainVerification(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized provider message type: %T", msg)
		}
	}
}

func wrapServiceResult(ctx sdk.Context, res gogoproto.Message) (*sdk.Result, error) {
	any, err := codectypes.NewAnyWithValue(res)
	if err != nil {
		return nil, err
	}

	var data []byte
	if res != nil {
		data, err = gogoproto.Marshal(res)
		if err != nil {
			return nil, err
		}
	}

	var events []abci.Event
	if evtMgr := ctx.EventManager(); evtMgr != nil {
		events = evtMgr.ABCIEvents()
	}

	return &sdk.Result{
		Data:         data,
		Events:       events,
		MsgResponses: []*codectypes.Any{any},
	}, nil
}
