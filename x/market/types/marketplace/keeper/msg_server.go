package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

func (ms msgServer) ResizeAllocation(goCtx context.Context, msg *marketplace.MsgResizeAllocation) (*marketplace.MsgResizeAllocationResponse, error) {
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

	if len(msg.ResourceUnits) == 0 {
		return nil, marketplace.ErrInvalidRequest.Wrap("resource_units must be provided")
	}

	targetState, err := marketplace.ValidateLifecycleTransition(allocation.State, marketplace.LifecycleActionResize)
	if err != nil {
		return nil, marketplace.ErrInvalidStateTransition.Wrap(err.Error())
	}

	reason := msg.Reason
	if reason == "" {
		reason = "customer requested resize"
	}

	if targetState != allocation.State {
		if err := allocation.SetStateAt(targetState, reason, ctx.BlockTime()); err != nil {
			return nil, err
		}
		if err := ms.keeper.UpdateAllocation(ctx, allocation); err != nil {
			return nil, err
		}
	}

	operationID := marketplace.GenerateOperationID(allocation.ID.String(), marketplace.LifecycleActionResize, ctx.BlockTime())
	resourceUnits := make(map[string]uint64, len(msg.ResourceUnits))
	for _, unit := range msg.ResourceUnits {
		if unit.ResourceType == "" {
			return nil, marketplace.ErrInvalidRequest.Wrap("resource_type is required")
		}
		if unit.Units == 0 {
			return nil, marketplace.ErrInvalidRequest.Wrap("resource unit quantity must be greater than 0")
		}
		resourceUnits[unit.ResourceType] = unit.Units
	}
	parameters := map[string]interface{}{
		"resource_units": resourceUnits,
	}
	seq := ms.keeper.IncrementEventSequence(ctx)
	event := marketplace.NewLifecycleActionRequestedEventAt(
		allocation.ID.String(),
		order.ID.String(),
		allocation.ProviderAddress,
		marketplace.LifecycleActionResize,
		operationID,
		msg.Customer,
		targetState,
		parameters,
		marketplace.RollbackPolicyAutomatic,
		ctx.BlockHeight(),
		seq,
		ctx.BlockTime(),
	)
	if err := ms.keeper.EmitMarketplaceEvent(ctx, event); err != nil {
		return nil, err
	}

	return &marketplace.MsgResizeAllocationResponse{}, nil
}

func (ms msgServer) PauseAllocation(goCtx context.Context, msg *marketplace.MsgPauseAllocation) (*marketplace.MsgPauseAllocationResponse, error) {
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

	targetState, err := marketplace.ValidateLifecycleTransition(allocation.State, marketplace.LifecycleActionSuspend)
	if err != nil {
		return nil, marketplace.ErrInvalidStateTransition.Wrap(err.Error())
	}

	reason := msg.Reason
	if reason == "" {
		reason = "customer requested pause"
	}

	if targetState != allocation.State {
		if err := allocation.SetStateAt(targetState, reason, ctx.BlockTime()); err != nil {
			return nil, err
		}
		if err := ms.keeper.UpdateAllocation(ctx, allocation); err != nil {
			return nil, err
		}
	}

	operationID := marketplace.GenerateOperationID(allocation.ID.String(), marketplace.LifecycleActionSuspend, ctx.BlockTime())
	seq := ms.keeper.IncrementEventSequence(ctx)
	event := marketplace.NewLifecycleActionRequestedEventAt(
		allocation.ID.String(),
		order.ID.String(),
		allocation.ProviderAddress,
		marketplace.LifecycleActionSuspend,
		operationID,
		msg.Customer,
		targetState,
		nil,
		marketplace.RollbackPolicyAutomatic,
		ctx.BlockHeight(),
		seq,
		ctx.BlockTime(),
	)
	if err := ms.keeper.EmitMarketplaceEvent(ctx, event); err != nil {
		return nil, err
	}

	return &marketplace.MsgPauseAllocationResponse{}, nil
}

func (ms msgServer) WaldurCallback(goCtx context.Context, msg *marketplace.MsgWaldurCallback) (*marketplace.MsgWaldurCallbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var callback marketplace.WaldurCallback
	if msg.Payload != "" {
		if err := json.Unmarshal([]byte(msg.Payload), &callback); err != nil {
			return nil, marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
		}
	}

	if callback.ActionType == "" {
		callback.ActionType = marketplace.WaldurActionType(msg.CallbackType)
	}
	if callback.ChainEntityID == "" {
		callback.ChainEntityID = msg.ResourceId
	}
	if callback.ChainEntityType == "" && msg.Status != "" {
		callback.ChainEntityType = marketplace.WaldurSyncType(msg.Status)
	}
	if callback.SignerID == "" {
		callback.SignerID = msg.Sender
	}
	if len(callback.Signature) == 0 {
		callback.Signature = msg.Signature
	}
	if callback.Payload == nil {
		callback.Payload = map[string]string{}
	}
	if msg.Payload != "" && len(callback.Payload) == 0 {
		callback.Payload["payload"] = msg.Payload
	}
	if callback.Nonce == "" {
		callback.Nonce = fmt.Sprintf("nonce_%d", ctx.BlockTime().UnixNano())
	}
	if callback.Timestamp.IsZero() {
		callback.Timestamp = ctx.BlockTime().UTC()
	}
	if callback.ExpiresAt.IsZero() {
		callback.ExpiresAt = ctx.BlockTime().Add(time.Hour).UTC()
	}

	if err := ms.keeper.ProcessWaldurCallback(ctx, &callback); err != nil {
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
