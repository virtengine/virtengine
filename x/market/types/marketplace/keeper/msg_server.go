package keeper

import (
	"context"
	"encoding/json"
	"fmt"

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

func (ms msgServer) CreateOffering(goCtx context.Context, msg *marketplace.MsgCreateOffering) (*marketplace.MsgCreateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}

	var offering marketplace.Offering
	if err := json.Unmarshal(msg.Offering, &offering); err != nil {
		return nil, fmt.Errorf("decode offering: %w", err)
	}

	if offering.ID.ProviderAddress == "" {
		return nil, fmt.Errorf("offering provider address is required")
	}
	if msg.Sender != offering.ID.ProviderAddress {
		return nil, fmt.Errorf("sender does not match offering provider")
	}

	if err := ms.keeper.CreateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	return &marketplace.MsgCreateOfferingResponse{OfferingId: offering.ID.String()}, nil
}

func (ms msgServer) UpdateOffering(goCtx context.Context, msg *marketplace.MsgUpdateOffering) (*marketplace.MsgUpdateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}
	if msg.OfferingId == "" {
		return nil, fmt.Errorf("offering id is required")
	}

	var offering marketplace.Offering
	if err := json.Unmarshal(msg.Offering, &offering); err != nil {
		return nil, fmt.Errorf("decode offering: %w", err)
	}

	offeringID, err := marketplace.ParseOfferingID(msg.OfferingId)
	if err != nil {
		return nil, fmt.Errorf("invalid offering id: %w", err)
	}

	if offering.ID.ProviderAddress == "" {
		offering.ID = offeringID
	}

	if offering.ID.String() != msg.OfferingId {
		return nil, fmt.Errorf("offering id mismatch")
	}
	if msg.Sender != offering.ID.ProviderAddress {
		return nil, fmt.Errorf("sender does not match offering provider")
	}

	if err := ms.keeper.UpdateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	return &marketplace.MsgUpdateOfferingResponse{}, nil
}

func (ms msgServer) DeprecateOffering(goCtx context.Context, msg *marketplace.MsgDeprecateOffering) (*marketplace.MsgDeprecateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("message is nil")
	}
	if msg.OfferingId == "" {
		return nil, fmt.Errorf("offering id is required")
	}

	offeringID, err := marketplace.ParseOfferingID(msg.OfferingId)
	if err != nil {
		return nil, fmt.Errorf("invalid offering id: %w", err)
	}
	if msg.Sender != offeringID.ProviderAddress {
		return nil, fmt.Errorf("sender does not match offering provider")
	}

	if err := ms.keeper.TerminateOffering(ctx, offeringID, msg.Reason); err != nil {
		return nil, err
	}

	return &marketplace.MsgDeprecateOfferingResponse{}, nil
}
