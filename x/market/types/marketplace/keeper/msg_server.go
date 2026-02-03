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

func (ms msgServer) CreateOffering(goCtx context.Context, msg *marketplace.MsgCreateOffering) (*marketplace.MsgCreateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, marketplace.ErrUnauthorized.Wrap("empty message")
	}

	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, marketplace.ErrUnauthorized.Wrap("invalid provider address")
	}

	if !ms.keeper.IsProvider(ctx, providerAddr) {
		return nil, marketplace.ErrNotProvider
	}

	if msg.Offering == nil {
		return nil, marketplace.ErrInvalidEncryptedPayload.Wrap("offering is required")
	}

	offering := offeringFromProto(msg.Offering)
	if offering.ID.ProviderAddress == "" {
		offering.ID.ProviderAddress = msg.Provider
	}

	if offering.ID.ProviderAddress != msg.Provider {
		return nil, marketplace.ErrUnauthorized.Wrap("provider does not match offering id")
	}

	if offering.ID.Sequence == 0 {
		offering.ID.Sequence = nextOfferingSequence(ctx, ms.keeper, msg.Provider)
	}

	if offering.State == marketplace.OfferingStateUnspecified {
		offering.State = marketplace.OfferingStateActive
	}

	if offering.IdentityRequirement == (marketplace.IdentityRequirement{}) {
		offering.IdentityRequirement = marketplace.DefaultIdentityRequirement()
	}

	now := ctx.BlockTime().UTC()
	if offering.CreatedAt.IsZero() {
		offering.CreatedAt = now
	}
	if offering.UpdatedAt.IsZero() {
		offering.UpdatedAt = now
	}
	if offering.State == marketplace.OfferingStateActive && offering.ActivatedAt == nil {
		activated := now
		offering.ActivatedAt = &activated
	}

	if err := ms.keeper.CreateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	return &marketplace.MsgCreateOfferingResponse{OfferingId: offering.ID.String()}, nil
}

func (ms msgServer) UpdateOffering(goCtx context.Context, msg *marketplace.MsgUpdateOffering) (*marketplace.MsgUpdateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, marketplace.ErrUnauthorized.Wrap("empty message")
	}

	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, marketplace.ErrUnauthorized.Wrap("invalid provider address")
	}

	if !ms.keeper.IsProvider(ctx, providerAddr) {
		return nil, marketplace.ErrNotProvider
	}

	if msg.OfferingId == "" {
		return nil, marketplace.ErrOfferingNotFound.Wrap("offering_id is required")
	}

	offeringID, err := marketplace.ParseOfferingID(msg.OfferingId)
	if err != nil {
		return nil, marketplace.ErrOfferingNotFound.Wrap(err.Error())
	}

	if offeringID.ProviderAddress != msg.Provider {
		return nil, marketplace.ErrUnauthorized.Wrap("provider does not match offering id")
	}

	existing, found := ms.keeper.GetOffering(ctx, offeringID)
	if !found {
		return nil, marketplace.ErrOfferingNotFound
	}

	if msg.Updates == nil {
		return nil, marketplace.ErrInvalidEncryptedPayload.Wrap("updates are required")
	}

	updates := offeringFromProto(msg.Updates)
	updates.ID = existing.ID
	if updates.State == marketplace.OfferingStateUnspecified {
		updates.State = existing.State
	}
	if updates.Category == "" {
		updates.Category = existing.Category
	}
	if updates.Name == "" {
		updates.Name = existing.Name
	}
	if updates.Description == "" {
		updates.Description = existing.Description
	}
	if updates.Version == "" {
		updates.Version = existing.Version
	}
	if pricingInfoEmpty(updates.Pricing) {
		updates.Pricing = existing.Pricing
	}
	if updates.IdentityRequirement == (marketplace.IdentityRequirement{}) {
		updates.IdentityRequirement = existing.IdentityRequirement
	}
	if updates.PublicMetadata == nil {
		updates.PublicMetadata = existing.PublicMetadata
	}
	if updates.Specifications == nil {
		updates.Specifications = existing.Specifications
	}
	if updates.Tags == nil {
		updates.Tags = existing.Tags
	}
	if updates.Regions == nil {
		updates.Regions = existing.Regions
	}
	if updates.EncryptedSecrets == nil {
		updates.EncryptedSecrets = existing.EncryptedSecrets
	}

	updates.CreatedAt = existing.CreatedAt
	updates.TotalOrderCount = existing.TotalOrderCount
	updates.ActiveOrderCount = existing.ActiveOrderCount

	if updates.ActivatedAt == nil {
		updates.ActivatedAt = existing.ActivatedAt
	}
	if updates.TerminatedAt == nil {
		updates.TerminatedAt = existing.TerminatedAt
	}

	now := ctx.BlockTime().UTC()
	updates.UpdatedAt = now
	if updates.State == marketplace.OfferingStateActive && updates.ActivatedAt == nil {
		activated := now
		updates.ActivatedAt = &activated
	}

	if err := ms.keeper.UpdateOffering(ctx, &updates); err != nil {
		return nil, err
	}

	return &marketplace.MsgUpdateOfferingResponse{}, nil
}

func (ms msgServer) DeactivateOffering(goCtx context.Context, msg *marketplace.MsgDeactivateOffering) (*marketplace.MsgDeactivateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, marketplace.ErrUnauthorized.Wrap("empty message")
	}

	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, marketplace.ErrUnauthorized.Wrap("invalid provider address")
	}

	if !ms.keeper.IsProvider(ctx, providerAddr) {
		return nil, marketplace.ErrNotProvider
	}

	if msg.OfferingId == "" {
		return nil, marketplace.ErrOfferingNotFound.Wrap("offering_id is required")
	}

	offeringID, err := marketplace.ParseOfferingID(msg.OfferingId)
	if err != nil {
		return nil, marketplace.ErrOfferingNotFound.Wrap(err.Error())
	}

	if offeringID.ProviderAddress != msg.Provider {
		return nil, marketplace.ErrUnauthorized.Wrap("provider does not match offering id")
	}

	if err := ms.keeper.TerminateOffering(ctx, offeringID, "offering deactivated"); err != nil {
		return nil, err
	}

	return &marketplace.MsgDeactivateOfferingResponse{}, nil
}

func (ms msgServer) AcceptBid(goCtx context.Context, msg *marketplace.MsgAcceptBid) (*marketplace.MsgAcceptBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, marketplace.ErrUnauthorized.Wrap("empty message")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return nil, marketplace.ErrUnauthorized.Wrap("invalid customer address")
	}

	if msg.BidId == "" {
		return nil, marketplace.ErrBidNotFound.Wrap("bid_id is required")
	}

	bidID, err := marketplace.ParseBidID(msg.BidId)
	if err != nil {
		return nil, marketplace.ErrBidNotFound.Wrap(err.Error())
	}

	orderID := bidID.OrderID
	if msg.OrderId != "" {
		parsedOrderID, err := marketplace.ParseOrderID(msg.OrderId)
		if err != nil {
			return nil, marketplace.ErrOrderNotFound.Wrap(err.Error())
		}
		if parsedOrderID != bidID.OrderID {
			return nil, marketplace.ErrInvalidOrderState.Wrap("order_id does not match bid")
		}
		orderID = parsedOrderID
	}

	order, found := ms.keeper.GetOrder(ctx, orderID)
	if !found {
		return nil, marketplace.ErrOrderNotFound
	}

	if order.ID.CustomerAddress != msg.Customer {
		return nil, marketplace.ErrNotCustomer
	}

	allocation, err := ms.keeper.AcceptBid(ctx, bidID)
	if err != nil {
		return nil, err
	}

	return &marketplace.MsgAcceptBidResponse{AllocationId: allocation.ID.String()}, nil
}

func (ms msgServer) TerminateAllocation(goCtx context.Context, msg *marketplace.MsgTerminateAllocation) (*marketplace.MsgTerminateAllocationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, marketplace.ErrUnauthorized.Wrap("empty message")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return nil, marketplace.ErrUnauthorized.Wrap("invalid customer address")
	}

	if msg.AllocationId == "" {
		return nil, marketplace.ErrAllocationNotFound.Wrap("allocation_id is required")
	}

	allocationID, err := marketplace.ParseAllocationID(msg.AllocationId)
	if err != nil {
		return nil, marketplace.ErrAllocationNotFound.Wrap(err.Error())
	}

	allocation, found := ms.keeper.GetAllocation(ctx, allocationID)
	if !found {
		return nil, marketplace.ErrAllocationNotFound
	}

	order, found := ms.keeper.GetOrder(ctx, allocationID.OrderID)
	if !found {
		return nil, marketplace.ErrOrderNotFound
	}

	if order.ID.CustomerAddress != msg.Customer {
		return nil, marketplace.ErrNotCustomer
	}

	if allocation.State.IsTerminal() {
		return nil, marketplace.ErrInvalidStateTransition.Wrap("allocation already terminal")
	}

	reason := msg.Reason
	if reason == "" {
		reason = "customer requested termination"
	}

	if allocation.State != marketplace.AllocationStateTerminating {
		if err := allocation.SetStateAt(marketplace.AllocationStateTerminating, reason, ctx.BlockTime()); err != nil {
			return nil, err
		}
		if err := ms.keeper.UpdateAllocation(ctx, allocation); err != nil {
			return nil, err
		}
	}

	if !order.State.IsTerminal() && order.State != marketplace.OrderStatePendingTermination {
		if err := order.SetStateAt(marketplace.OrderStatePendingTermination, reason, ctx.BlockTime()); err != nil {
			return nil, marketplace.ErrInvalidStateTransition.Wrap(err.Error())
		}
		if err := ms.keeper.UpdateOrder(ctx, order); err != nil {
			return nil, err
		}
	}

	seq := ms.keeper.IncrementEventSequence(ctx)
	event := marketplace.NewTerminateRequestedEventAt(
		allocation.ID.String(),
		order.ID.String(),
		allocation.ProviderAddress,
		msg.Customer,
		reason,
		false,
		ctx.BlockHeight(),
		seq,
		ctx.BlockTime(),
	)
	if err := ms.keeper.EmitMarketplaceEvent(ctx, event); err != nil {
		return nil, err
	}

	return &marketplace.MsgTerminateAllocationResponse{}, nil
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

func nextOfferingSequence(ctx sdk.Context, k IKeeper, providerAddress string) uint64 {
	maxSeq := uint64(0)
	offerings := k.GetOfferingsByProvider(ctx, providerAddress)
	for _, offering := range offerings {
		if offering.ID.Sequence > maxSeq {
			maxSeq = offering.ID.Sequence
		}
	}
	return maxSeq + 1
}

func pricingInfoEmpty(pricing marketplace.PricingInfo) bool {
	return pricing.Model == "" &&
		pricing.BasePrice == 0 &&
		pricing.Currency == "" &&
		len(pricing.UsageRates) == 0 &&
		pricing.MinimumCommitment == 0
}
