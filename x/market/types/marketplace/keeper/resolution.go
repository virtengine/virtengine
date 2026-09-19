package keeper

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// CapacityKeeper is the minimal capacity interface the resolution engine needs.
// It is satisfied by the x/resources reservation keeper when wired.
type CapacityKeeper interface {
	// ReserveForOrder reserves capacity for a resolved order/candidate pair and
	// returns the reservation identifier.
	ReserveForOrder(ctx sdk.Context, order marketplace.Order, candidate marketplace.Candidate) (string, error)
	// ReleaseReservation releases a previously reserved capacity claim.
	ReleaseReservation(ctx sdk.Context, reservationID string) error
}

// SetCapacityKeeper wires the optional capacity keeper into the resolver.
func (k *Keeper) SetCapacityKeeper(capacity CapacityKeeper) {
	k.capacityKeeper = capacity
}

// ResolveOpenOrders runs one deterministic resolution pass over open orders.
//
// The function is consensus-safe: it performs no external calls, reads only
// committed state, and orders all iteration deterministically. It is intended to
// be invoked from EndBlock.
func (k Keeper) ResolveOpenOrders(ctx sdk.Context) (int, error) {
	params := k.GetParams(ctx)
	if !params.EnableAutoResolution {
		return 0, nil
	}

	now := ctx.BlockTime()
	candidates := make([]marketplace.Order, 0)
	k.WithOrders(ctx, func(order marketplace.Order) bool {
		if order.State != marketplace.OrderStateOpen {
			return false
		}
		if order.IsBidOrder() && !order.BidWindowClosed(now) {
			return false
		}
		candidates = append(candidates, order)
		return false
	})

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].ID.String() < candidates[j].ID.String()
	})

	resolved := 0
	for i := range candidates {
		ok, err := k.resolveOrder(ctx, &candidates[i], params)
		if err != nil {
			return resolved, err
		}
		if ok {
			resolved++
		}
	}
	return resolved, nil
}

func (k Keeper) resolveOrder(ctx sdk.Context, order *marketplace.Order, params marketplace.Params) (bool, error) {
	policy := marketplace.ResolutionPolicy{
		AllowPartialFill: params.AllowPartialFill,
		PreferNative:     params.PreferNativeSupply,
		PreferListings:   true,
	}

	candidates, err := k.buildResolutionCandidates(ctx, order, params)
	if err != nil {
		return false, err
	}
	result := marketplace.SelectCandidates(candidates, uint64(order.RequestedQuantity), policy)

	if !result.Matched {
		// A bid order whose window closed without a match is failed. Direct
		// orders remain open and are retried on subsequent blocks until expiry.
		if order.IsBidOrder() || (order.ExpiresAt != nil && ctx.BlockTime().After(*order.ExpiresAt)) {
			if err := order.SetStateAt(marketplace.OrderStateFailed, "no eligible supply", ctx.BlockTime()); err != nil {
				return false, err
			}
			if err := k.updateOrder(ctx, order); err != nil {
				return false, err
			}
			k.rejectOpenBids(ctx, order.ID, nil)
			return false, nil
		}
		return false, nil
	}

	now := ctx.BlockTime().UTC()
	acceptedBids := make(map[string]struct{})
	var totalAccepted uint64
	var firstProvider string

	sequence := uint64(0)
	for _, match := range result.Matches {
		sequence++
		allocationID := marketplace.AllocationID{
			OrderID:  order.ID,
			Sequence: sequence,
		}
		bidID := marketplace.BidID{}
		if match.Kind == marketplace.CandidateKindBid && match.BidID != nil {
			bidID = *match.BidID
		}

		allocation := marketplace.NewAllocationAt(
			allocationID,
			order.OfferingID,
			match.ProviderAddress,
			bidID,
			match.Price.Amount.Uint64(),
			now,
		)

		if k.capacityKeeper != nil {
			reservationID, err := k.capacityKeeper.ReserveForOrder(ctx, *order, match)
			if err != nil {
				return false, err
			}
			if reservationID != "" {
				if allocation.PublicMetadata == nil {
					allocation.PublicMetadata = map[string]string{}
				}
				allocation.PublicMetadata["reservation_id"] = reservationID
			}
		}

		if err := k.createAllocation(ctx, allocation); err != nil {
			return false, err
		}

		if match.Kind == marketplace.CandidateKindBid && match.BidID != nil {
			acceptedBids[match.BidID.String()] = struct{}{}
			if bid, found := k.GetBid(ctx, *match.BidID); found {
				bid.State = marketplace.BidStateAccepted
				if err := k.putBid(ctx, bid); err != nil {
					return false, err
				}
			}
		}

		if firstProvider == "" {
			firstProvider = match.ProviderAddress
		}
		totalAccepted += match.Price.Amount.Uint64()

		seq := k.IncrementEventSequence(ctx)
		event := marketplace.NewAllocationCreatedEventAt(allocation, order.ID.CustomerAddress, ctx.BlockHeight(), seq, ctx.BlockTime())
		if err := k.EmitMarketplaceEvent(ctx, event); err != nil {
			return false, err
		}
	}

	if err := order.SetStateAt(marketplace.OrderStateMatched, "resolved by engine", now); err != nil {
		return false, err
	}
	order.AllocatedProviderAddress = firstProvider
	order.AcceptedPrice = totalAccepted
	if err := k.updateOrder(ctx, order); err != nil {
		return false, err
	}

	// Reject all losing open bids.
	k.rejectOpenBids(ctx, order.ID, acceptedBids)

	// Emit a durable Waldur command when the winning supply is Waldur-backed.
	if err := k.enqueueWaldurLeaseCommand(ctx, order, firstProvider); err != nil {
		return false, err
	}

	return true, nil
}

func (k Keeper) rejectOpenBids(ctx sdk.Context, orderID marketplace.OrderID, accepted map[string]struct{}) {
	k.WithBidsForOrder(ctx, orderID, func(bid marketplace.MarketplaceBid) bool {
		if bid.State != marketplace.BidStateOpen {
			return false
		}
		if accepted != nil {
			if _, ok := accepted[bid.ID.String()]; ok {
				return false
			}
		}
		bid.State = marketplace.BidStateRejected
		_ = k.putBid(ctx, &bid)
		return false
	})
}

func (k Keeper) buildResolutionCandidates(
	ctx sdk.Context,
	order *marketplace.Order,
	params marketplace.Params,
) ([]marketplace.Candidate, error) {
	if order == nil {
		return nil, fmt.Errorf("order is required")
	}

	mode := order.EffectiveAcquisitionMode()
	candidates := make([]marketplace.Candidate, 0)

	addOffering := func(offering marketplace.Offering) {
		if order.HasOffering() && offering.ID != order.OfferingID {
			return
		}
		if !offering.AdmitsOrder(mode, order.Selector) {
			return
		}
		quote, err := marketplace.CalculateOfferingPrice(&offering, order.ResourceUnits, order.RequestedQuantity)
		if err != nil {
			return
		}
		if !quote.Total.IsValid() || !quote.Total.Amount.IsPositive() || !quote.Total.Amount.IsUint64() {
			return
		}
		total := quote.Total.Amount.Uint64()
		if order.MaxBidPrice > 0 && total > order.MaxBidPrice {
			return
		}
		if order.Selector != nil && order.Selector.MaxPrice != nil {
			if order.Selector.MaxPrice.Denom != quote.Total.Denom ||
				order.Selector.MaxPrice.Amount.LT(quote.Total.Amount) {
				return
			}
		}
		id := offering.ID
		candidates = append(candidates, marketplace.Candidate{
			Kind:            marketplace.CandidateKindListing,
			OfferingID:      &id,
			ProviderAddress: offering.ID.ProviderAddress,
			Price:           quote.Total,
			Capacity:        uint64(order.RequestedQuantity),
			CapacityFit:     1,
			Sequence:        offering.ID.Sequence,
			Source:          offering.Source,
			GPUType:         gpuTypeFromSpecs(offering.Specifications),
			ResourceClass:   resourceClassFromCategory(offering.Category),
		})
	}

	if order.HasOffering() {
		if offering, found := k.GetOffering(ctx, order.OfferingID); found {
			addOffering(*offering)
		}
	} else {
		k.WithOfferings(ctx, func(offering marketplace.Offering) bool {
			addOffering(offering)
			return false
		})
	}

	denom := k.matchingDenomFor(ctx, order, params)

	k.WithBidsForOrder(ctx, order.ID, func(bid marketplace.MarketplaceBid) bool {
		if bid.State != marketplace.BidStateOpen {
			return false
		}
		if order.MaxBidPrice > 0 && bid.Price > order.MaxBidPrice {
			return false
		}
		bidID := bid.ID
		candidates = append(candidates, marketplace.Candidate{
			Kind:            marketplace.CandidateKindBid,
			BidID:           &bidID,
			ProviderAddress: bid.ID.ProviderAddress,
			Price:           sdk.NewCoin(denom, sdkmath.NewIntFromUint64(bid.Price)),
			Capacity:        uint64(order.RequestedQuantity),
			CapacityFit:     1,
			Sequence:        bid.ID.Sequence,
		})
		return false
	})

	return candidates, nil
}

func (k Keeper) matchingDenomFor(ctx sdk.Context, order *marketplace.Order, params marketplace.Params) string {
	if order.HasOffering() {
		if offering, found := k.GetOffering(ctx, order.OfferingID); found {
			if offering.Pricing.Currency != "" {
				return offering.Pricing.Currency
			}
			if len(offering.Prices) > 0 && offering.Prices[0].Price.Denom != "" {
				return offering.Prices[0].Price.Denom
			}
		}
	}
	if params.DefaultMatchingDenom != "" {
		return params.DefaultMatchingDenom
	}
	return "uvirt"
}

// gpuTypeFromSpecs extracts an accelerator type hint from specifications.
func gpuTypeFromSpecs(specs map[string]string) string {
	for _, key := range []string{"vm.gpu_type", "gpu_type", "container.gpu_type"} {
		if value := specs[key]; value != "" {
			return value
		}
	}
	return ""
}

// resourceClassFromCategory maps an offering category to a coarse capacity
// class understood by x/resources.
func resourceClassFromCategory(category marketplace.OfferingCategory) string {
	switch category {
	case marketplace.OfferingCategoryStorage:
		return "storage"
	case marketplace.OfferingCategoryNetwork:
		return "network"
	default:
		return "compute"
	}
}

// enqueueWaldurLeaseCommand emits a durable Waldur order-creation command when a
// resolved order is backed by a Waldur listing.
func (k Keeper) enqueueWaldurLeaseCommand(ctx sdk.Context, order *marketplace.Order, provider string) error {
	if !order.HasOffering() {
		return nil
	}
	offering, found := k.GetOffering(ctx, order.OfferingID)
	if !found || offering.Source.Effective() != marketplace.OfferingSourceWaldur || offering.Waldur == nil {
		return nil
	}

	command, err := marketplace.NewWaldurCommandAt(
		marketplace.WaldurCommandCreateOrder,
		offering.Waldur.InstanceID,
		order.ID.String(),
		ctx.BlockTime(),
	)
	if err != nil {
		return err
	}
	command.WaldurOfferingUUID = offering.Waldur.OfferingUUID
	command.ChainEntityType = marketplace.SyncTypeOrder
	command.ChainEntityID = order.ID.String()
	command.BackendID = order.ID.String()
	command.Payload["provider"] = provider
	command.Payload["customer"] = order.ID.CustomerAddress
	return k.EnqueueWaldurCommand(ctx, command)
}
