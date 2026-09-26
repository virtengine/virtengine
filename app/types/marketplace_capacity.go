package types

import (
	"fmt"
	"math"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	marketplacekeeper "github.com/virtengine/virtengine/x/market/types/marketplace/keeper"
	resourceskeeper "github.com/virtengine/virtengine/x/resources/keeper"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
)

// marketplaceCapacityAdapter implements the marketplace resolution engine's
// CapacityKeeper against the canonical x/resources reservation ledger.
//
// A resolved marketplace match mints a canonical reservation with marketplace
// lineage, so the capacity-conservation invariant
// (available + nonterminal reservations = declared total) covers catalog
// demand exactly like canonical market leases.
type marketplaceCapacityAdapter struct {
	resources resourceskeeper.Keeper
}

func newMarketplaceCapacityAdapter(resources resourceskeeper.Keeper) marketplacekeeper.CapacityKeeper {
	return marketplaceCapacityAdapter{resources: resources}
}

// ReserveForOrder reserves and activates capacity for a resolved order/candidate
// pair and returns the canonical reservation identifier.
func (a marketplaceCapacityAdapter) ReserveForOrder(ctx sdk.Context, order marketplacetypes.Order, candidate marketplacetypes.Candidate) (string, error) {
	capacity, err := capacityForOrder(order, candidate)
	if err != nil {
		return "", err
	}

	reference := candidate.ProviderAddress
	if candidate.Kind == marketplacetypes.CandidateKindBid && candidate.BidID != nil {
		reference = candidate.BidID.String()
	} else if candidate.OfferingID != nil {
		reference = candidate.OfferingID.String()
	}
	orderID := order.ID.String()
	request := resourcestypes.ReservationRequest{
		IdempotencyKey:   fmt.Sprintf("mktplace/order/%s/%s/%s/%s", orderID, candidate.Kind, reference, candidate.Price.Amount.String()),
		RequestId:        fmt.Sprintf("mktplace/order/%s", orderID),
		RequesterAddress: order.ID.CustomerAddress,
		ProviderAddress:  candidate.ProviderAddress,
		ResourceClass:    resourceClassForCandidate(candidate),
		Capacity:         capacity,
		ConsumerType:     "mktplace_allocation",
		ConsumerId:       fmt.Sprintf("%s/%s", orderID, candidate.ProviderAddress),
		MarketOrderId:    orderID,
		Version:          1,
	}
	if candidate.Kind == marketplacetypes.CandidateKindBid && candidate.BidID != nil {
		request.MarketBidId = candidate.BidID.String()
	}

	reservation, err := a.resources.Reserve(ctx, request)
	if err != nil {
		return "", err
	}
	link := resourcestypes.ReservationLink{
		ConsumerType:  request.ConsumerType,
		ConsumerId:    request.ConsumerId,
		MarketOrderId: request.MarketOrderId,
		MarketBidId:   request.MarketBidId,
	}
	if _, err := a.resources.ActivateReservation(ctx, reservation.ReservationId, link); err != nil {
		return "", err
	}
	return reservation.ReservationId, nil
}

// ReleaseReservation releases a previously reserved capacity claim.
func (a marketplaceCapacityAdapter) ReleaseReservation(ctx sdk.Context, reservationID string) error {
	_, err := a.resources.ReleaseReservation(ctx, reservationID, "mktplace allocation released")
	return err
}

// Canonical unit and class names for capacity mapping.
const (
	capacityUnitCPU     = "cpu"
	capacityUnitRAM     = "ram"
	capacityUnitStorage = "storage"
	capacityUnitNetwork = "network"
	capacityUnitGPU     = "gpu"

	capacityClassStorage = "storage"
	capacityClassNetwork = "network"
)

// resourceClassForCandidate maps a candidate's coarse class to x/resources.
func resourceClassForCandidate(candidate marketplacetypes.Candidate) resourcestypes.ResourceClass {
	switch strings.ToLower(strings.TrimSpace(candidate.ResourceClass)) {
	case capacityClassStorage:
		return resourcestypes.ResourceClassStorage
	case capacityClassNetwork:
		return resourcestypes.ResourceClassNetwork
	default:
		return resourcestypes.ResourceClassCompute
	}
}

// capacityForOrder derives reservation capacity from an order's resource units
// scaled by the matched quantity, saturating at math.MaxInt64.
func capacityForOrder(order marketplacetypes.Order, candidate marketplacetypes.Candidate) (resourcestypes.ResourceCapacity, error) {
	quantity := candidate.Capacity
	if quantity == 0 {
		quantity = uint64(order.RequestedQuantity)
	}
	var capacity resourcestypes.ResourceCapacity
	for unit, amount := range order.ResourceUnits {
		scaled := saturatingMul(amount, quantity)
		switch normalizeUnit(unit) {
		case capacityUnitCPU:
			capacity.CpuCores = saturatingAdd(capacity.CpuCores, scaled)
		case capacityUnitRAM:
			capacity.MemoryGb = saturatingAdd(capacity.MemoryGb, scaled)
		case capacityUnitStorage:
			capacity.StorageGb = saturatingAdd(capacity.StorageGb, scaled)
		case capacityUnitNetwork:
			capacity.NetworkMbps = saturatingAdd(capacity.NetworkMbps, scaled)
		case capacityUnitGPU:
			capacity.Gpus = saturatingAdd(capacity.Gpus, scaled)
		}
	}
	if capacity.Gpus > 0 {
		if candidate.GPUType == "" {
			return resourcestypes.ResourceCapacity{}, fmt.Errorf("gpu capacity requires a gpu_type: set vm.gpu_type on the offering")
		}
		capacity.GpuType = candidate.GPUType
	}
	if capacity.CpuCores == 0 && capacity.MemoryGb == 0 && capacity.StorageGb == 0 && capacity.NetworkMbps == 0 && capacity.Gpus == 0 {
		return resourcestypes.ResourceCapacity{}, fmt.Errorf("order carries no reservable capacity: set resource_units")
	}
	return capacity, nil
}

func normalizeUnit(unit string) string {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "cpu", "vcpu", "cores", "cpu_cores":
		return "cpu"
	case "ram", "memory", "memory_gb", "memory_mb", "mem":
		return "ram"
	case "storage", "disk", "disk_gb", "volume":
		return "storage"
	case "network", "bandwidth", "network_mbps":
		return "network"
	case "gpu", "gpus", "gpu_count":
		return "gpu"
	default:
		return ""
	}
}

func saturatingMul(a, b uint64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	if a > uint64(math.MaxInt64)/b {
		return math.MaxInt64
	}
	// #nosec G115 -- guarded above: a*b <= MaxInt64 with a,b >= 1, so both operands fit int64 and the product cannot overflow
	return int64(a) * int64(b)
}

func saturatingAdd(a int64, b int64) int64 {
	if b > 0 && a > math.MaxInt64-b {
		return math.MaxInt64
	}
	return a + b
}
