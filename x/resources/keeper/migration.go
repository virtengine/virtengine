package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

const legacyResourcesAllocationSource = "resources_allocation"

// ImportLegacyQuarantine records ambiguous legacy lineage without claiming
// synthetic capacity. Exact retries return the original record.
func (k Keeper) ImportLegacyQuarantine(ctx sdk.Context, source, reference, provider, consumerType, consumerID string, link types.ReservationLink, reason string) (types.Reservation, bool, error) {
	digest := sha256.Sum256([]byte(source + "\x00" + reference))
	reservationID := "legacy/quarantine/" + hex.EncodeToString(digest[:])
	if existing, found := k.GetReservation(ctx, reservationID); found {
		if existing.LegacySource != source || existing.LegacyReference != reference || existing.ProviderAddress != provider || existing.ConsumerType != consumerType || existing.ConsumerId != consumerID || existing.MarketOrderId != link.MarketOrderId || existing.MarketBidId != link.MarketBidId || existing.MarketLeaseId != link.MarketLeaseId || existing.HpcJobId != link.HpcJobId || existing.EscrowId != link.EscrowId || existing.CollateralId != link.CollateralId || existing.Reason != reason {
			return types.Reservation{}, false, types.ErrReservationConflict.Wrap("legacy quarantine retry payload differs")
		}
		return existing, false, nil
	}
	cacheCtx, write := ctx.CacheContext()
	reservation, created, err := k.importLegacyQuarantine(cacheCtx, reservationID, source, reference, provider, consumerType, consumerID, link, reason)
	if err != nil {
		return types.Reservation{}, false, err
	}
	write()
	return reservation, created, nil
}

func (k Keeper) importLegacyQuarantine(ctx sdk.Context, reservationID, source, reference, provider, consumerType, consumerID string, link types.ReservationLink, reason string) (types.Reservation, bool, error) {
	now := ctx.BlockTime()
	reservation := types.Reservation{ReservationId: reservationID, IdempotencyKey: reservationID, RequestId: reference, ProviderAddress: provider, ConsumerType: consumerType, ConsumerId: consumerID, MarketOrderId: link.MarketOrderId, MarketBidId: link.MarketBidId, MarketLeaseId: link.MarketLeaseId, HpcJobId: link.HpcJobId, EscrowId: link.EscrowId, CollateralId: link.CollateralId, State: types.ReservationStateQuarantined, Version: 1, Reason: reason, LegacySource: source, LegacyReference: reference, CreatedAt: now, UpdatedAt: now, QuarantinedAt: &now, CreatedHeight: ctx.BlockHeight(), QuarantinedHeight: ctx.BlockHeight()}
	if err := k.SetReservation(ctx, reservation); err != nil {
		return types.Reservation{}, false, err
	}
	if err := k.setReservationIndexes(ctx, reservation); err != nil {
		return types.Reservation{}, false, err
	}
	if err := k.recordReservationEvent(ctx, reservation, types.ReservationStateUnspecified, types.ReservationStateQuarantined, reason); err != nil {
		return types.Reservation{}, false, err
	}
	return reservation, true, nil
}

// MigrationReport contains deterministic preflight/reconciliation counts.
type MigrationReport struct {
	InventoriesScanned  uint64 `json:"inventories_scanned"`
	AllocationsScanned  uint64 `json:"allocations_scanned"`
	ReservationsCreated uint64 `json:"reservations_created"`
	AlreadyLinked       uint64 `json:"already_linked"`
	TerminalPreserved   uint64 `json:"terminal_preserved"`
	Quarantined         uint64 `json:"quarantined"`
}

// MigrateReservations converts unambiguous legacy allocations and quarantines
// active records that cannot be tied to inventory. It never creates capacity.
func (k Keeper) MigrateReservations(ctx sdk.Context) (MigrationReport, error) {
	cacheCtx, write := ctx.CacheContext()
	report, err := k.migrateReservations(cacheCtx)
	if err != nil {
		return report, err
	}
	write()
	return report, nil
}

func (k Keeper) migrateReservations(ctx sdk.Context) (MigrationReport, error) {
	report := MigrationReport{}
	inventories := make([]types.ResourceInventory, 0)
	k.WithInventories(ctx, func(inventory types.ResourceInventory) bool {
		inventories = append(inventories, inventory)
		report.InventoriesScanned++
		return false
	})
	allocations := make([]types.ResourceAllocation, 0)
	k.WithAllocations(ctx, func(allocation types.ResourceAllocation) bool {
		allocations = append(allocations, allocation)
		return false
	})
	sort.Slice(allocations, func(i, j int) bool { return allocations[i].AllocationId < allocations[j].AllocationId })

	inventoryByLegacyAllocation := make(map[string]types.ResourceInventory)
	legacyTotals := make(map[string]types.ResourceCapacity)
	legacyGroupsValid := make(map[string]bool)
	for _, allocation := range allocations {
		if allocation.ReservationId != "" || (allocation.State != types.AllocationStatePending && allocation.State != types.AllocationStateActive) {
			continue
		}
		matches := matchingLegacyInventories(inventories, allocation)
		if len(matches) != 1 || !capacityEqual(allocation.Required, allocation.Assigned) || validateCapacity(allocation.Assigned, false) != nil {
			continue
		}
		inventory := matches[0]
		identity := inventoryIdentity(inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
		inventoryByLegacyAllocation[allocation.AllocationId] = inventory
		legacyGroupsValid[identity] = true
		total, err := addCapacityChecked(legacyTotals[identity], allocation.Assigned)
		if err != nil {
			legacyGroupsValid[identity] = false
			continue
		}
		legacyTotals[identity] = total
	}
	for _, inventory := range inventories {
		identity := inventoryIdentity(inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
		if !legacyGroupsValid[identity] {
			continue
		}
		committed, err := addCapacityChecked(inventory.Available, legacyTotals[identity])
		if err != nil || !capacityEqual(committed, inventory.Total) {
			legacyGroupsValid[identity] = false
		}
	}

	for _, allocation := range allocations {
		report.AllocationsScanned++
		reservationID := "legacy/allocation/" + allocation.AllocationId
		if allocation.ReservationId != "" {
			existing, found := k.GetReservation(ctx, allocation.ReservationId)
			if found && legacyAllocationReservationMatches(allocation, existing) {
				report.AlreadyLinked++
				continue
			}
			source := "resources_allocation_inconsistent"
			if allocation.LegacySource == legacyResourcesAllocationSource {
				source = legacyResourcesAllocationSource
			}
			quarantine, created, err := k.ImportLegacyQuarantine(ctx, source, allocation.AllocationId, allocation.ProviderAddress, "legacy_allocation", allocation.AllocationId, types.ReservationLink{}, "legacy_allocation_reservation_inconsistent")
			if err != nil {
				return report, err
			}
			allocation.ReservationId = quarantine.ReservationId
			allocation.LegacySource = source
			allocation.LegacyReference = allocation.AllocationId
			if err := k.SetAllocation(ctx, allocation); err != nil {
				return report, err
			}
			if created {
				report.ReservationsCreated++
				report.Quarantined++
			} else {
				report.AlreadyLinked++
			}
			continue
		}
		if existing, found := k.GetReservation(ctx, reservationID); found {
			if !legacyAllocationReservationMatches(allocation, existing) {
				return report, types.ErrReservationConflict.Wrapf("legacy allocation %s reservation payload differs", allocation.AllocationId)
			}
			allocation.ReservationId = existing.ReservationId
			allocation.LegacySource = "resources_allocation"
			allocation.LegacyReference = allocation.AllocationId
			if err := k.SetAllocation(ctx, allocation); err != nil {
				return report, err
			}
			report.AlreadyLinked++
			continue
		}
		state := types.ReservationStateQuarantined
		reason := "legacy_allocation_ambiguous"
		inventory, linkedInventory := inventoryByLegacyAllocation[allocation.AllocationId]
		inventoryID := ""
		if linkedInventory {
			inventoryID = inventory.InventoryId
		}
		switch allocation.State {
		case types.AllocationStateReleased, types.AllocationStateExpired:
			state = types.ReservationStateReleased
			if allocation.State == types.AllocationStateExpired {
				state = types.ReservationStateExpired
			}
			reason = "legacy_terminal_preserved"
			report.TerminalPreserved++
		case types.AllocationStatePending, types.AllocationStateActive:
			if linkedInventory && legacyGroupsValid[inventoryIdentity(inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)] {
				state = types.ReservationStatePending
				if allocation.State == types.AllocationStateActive {
					state = types.ReservationStateActive
				}
				reason = "legacy_allocation_linked"
			} else {
				report.Quarantined++
			}
		default:
			report.Quarantined++
		}
		capacity := allocation.Assigned
		if state == types.ReservationStateQuarantined {
			capacity = types.ResourceCapacity{}
		}
		reservation := types.Reservation{ReservationId: reservationID, IdempotencyKey: reservationID, RequestId: allocation.RequestId, RequesterAddress: allocation.RequesterAddress, ProviderAddress: allocation.ProviderAddress, InventoryId: inventoryID, ResourceClass: allocation.ResourceClass, Capacity: capacity, State: state, ConsumerType: "legacy_allocation", ConsumerId: allocation.AllocationId, Version: 1, Reason: reason, LegacySource: legacyResourcesAllocationSource, LegacyReference: allocation.AllocationId, CreatedAt: allocation.CreatedAt, ActivatedAt: allocation.ActivatedAt, ExpiresAt: allocation.ExpiresAt, UpdatedAt: ctx.BlockTime(), CreatedHeight: allocation.BlockHeight}
		if reservation.State == types.ReservationStatePending && reservation.ExpiresAt == nil {
			expiresAt := ctx.BlockTime().Add(secondsToDuration(k.GetParams(ctx).ReservationTimeoutSeconds))
			reservation.ExpiresAt = &expiresAt
		}
		if reservation.State == types.ReservationStateActive && reservation.ActivatedAt == nil {
			activatedAt := ctx.BlockTime()
			reservation.ActivatedAt = &activatedAt
			reservation.ActivatedHeight = ctx.BlockHeight()
		}
		if reservation.State == types.ReservationStateQuarantined {
			quarantinedAt := ctx.BlockTime()
			reservation.QuarantinedAt = &quarantinedAt
			reservation.QuarantinedHeight = ctx.BlockHeight()
		}
		if reservation.State == types.ReservationStateReleased {
			releasedAt := ctx.BlockTime()
			reservation.ReleasedAt = &releasedAt
			reservation.ReleasedHeight = ctx.BlockHeight()
		}
		if reservation.State == types.ReservationStateExpired {
			expiredAt := ctx.BlockTime()
			reservation.ExpiredAt = &expiredAt
		}
		if err := k.SetReservation(ctx, reservation); err != nil {
			return report, err
		}
		if err := k.setReservationIndexes(ctx, reservation); err != nil {
			return report, err
		}
		fromState := types.ReservationStateUnspecified
		if err := k.recordReservationEvent(ctx, reservation, fromState, reservation.State, reason); err != nil {
			return report, err
		}
		allocation.ReservationId = reservationID
		allocation.LegacySource = legacyResourcesAllocationSource
		allocation.LegacyReference = allocation.AllocationId
		if err := k.SetAllocation(ctx, allocation); err != nil {
			return report, err
		}
		report.ReservationsCreated++
	}
	if err := k.ValidateCapacityConservation(ctx); err != nil {
		return report, fmt.Errorf("post-migration reservation invariant: %w", err)
	}
	k.ActivateCanonicalReservations(ctx)
	return report, nil
}

func legacyAllocationReservationMatches(allocation types.ResourceAllocation, reservation types.Reservation) bool {
	expectedSource := allocation.LegacySource
	if expectedSource == "" {
		expectedSource = legacyResourcesAllocationSource
	}
	if reservation.LegacySource != expectedSource || reservation.LegacyReference != allocation.AllocationId ||
		reservation.ProviderAddress != allocation.ProviderAddress || reservation.ConsumerType != "legacy_allocation" ||
		reservation.ConsumerId != allocation.AllocationId || reservation.RequesterAddress != allocation.RequesterAddress ||
		reservation.RequestId != allocation.RequestId || reservation.ResourceClass != allocation.ResourceClass {
		return false
	}
	if reservation.State == types.ReservationStateQuarantined {
		return capacityIsZero(reservation.Capacity)
	}
	return capacityEqual(reservation.Capacity, allocation.Assigned)
}

func matchingLegacyInventories(inventories []types.ResourceInventory, allocation types.ResourceAllocation) []types.ResourceInventory {
	matches := make([]types.ResourceInventory, 0, 1)
	for _, inventory := range inventories {
		if inventory.ProviderAddress == allocation.ProviderAddress && inventory.ResourceClass == allocation.ResourceClass {
			matches = append(matches, inventory)
		}
	}
	return matches
}
