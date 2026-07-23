package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

// ImportLegacyQuarantine records ambiguous legacy lineage without claiming
// synthetic capacity. Exact retries return the original record.
func (k Keeper) ImportLegacyQuarantine(ctx sdk.Context, source, reference, provider, consumerType, consumerID string, link types.ReservationLink, reason string) (types.Reservation, bool, error) {
	digest := sha256.Sum256([]byte(source + "\x00" + reference))
	reservationID := "legacy/quarantine/" + hex.EncodeToString(digest[:12])
	if existing, found := k.GetReservation(ctx, reservationID); found {
		return existing, false, nil
	}
	now := ctx.BlockTime()
	reservation := types.Reservation{ReservationId: reservationID, IdempotencyKey: reservationID, RequestId: reference, ProviderAddress: provider, ConsumerType: consumerType, ConsumerId: consumerID, MarketOrderId: link.MarketOrderId, MarketBidId: link.MarketBidId, MarketLeaseId: link.MarketLeaseId, HpcJobId: link.HpcJobId, EscrowId: link.EscrowId, CollateralId: link.CollateralId, State: types.ReservationStateQuarantined, Version: 1, Reason: reason, LegacySource: source, LegacyReference: reference, CreatedAt: now, UpdatedAt: now, QuarantinedAt: &now, CreatedHeight: ctx.BlockHeight(), QuarantinedHeight: ctx.BlockHeight()}
	if err := k.SetReservation(ctx, reservation); err != nil {
		return types.Reservation{}, false, err
	}
	k.setReservationIndexes(ctx, reservation)
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
	TerminalPreserved   uint64 `json:"terminal_preserved"`
	Quarantined         uint64 `json:"quarantined"`
}

// MigrateReservations converts unambiguous legacy allocations and quarantines
// active records that cannot be tied to inventory. It never creates capacity.
func (k Keeper) MigrateReservations(ctx sdk.Context) (MigrationReport, error) {
	report := MigrationReport{}
	k.WithInventories(ctx, func(types.ResourceInventory) bool { report.InventoriesScanned++; return false })
	allocations := make([]types.ResourceAllocation, 0)
	k.WithAllocations(ctx, func(allocation types.ResourceAllocation) bool {
		allocations = append(allocations, allocation)
		return false
	})
	sort.Slice(allocations, func(i, j int) bool { return allocations[i].AllocationId < allocations[j].AllocationId })
	for _, allocation := range allocations {
		report.AllocationsScanned++
		if allocation.ReservationId != "" {
			if _, found := k.GetReservation(ctx, allocation.ReservationId); found {
				continue
			}
		}
		state := types.ReservationStateQuarantined
		reason := "legacy_allocation_ambiguous"
		var inventoryID string
		matches := make([]types.ResourceInventory, 0, 1)
		k.WithInventories(ctx, func(inventory types.ResourceInventory) bool {
			if inventory.ProviderAddress == allocation.ProviderAddress && inventory.ResourceClass == allocation.ResourceClass {
				matches = append(matches, inventory)
			}
			return false
		})
		if len(matches) == 1 {
			inventoryID = matches[0].InventoryId
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
			// Legacy allocation capacity was already subtracted without a
			// reservation aggregate. Preserve it as quarantined until an operator
			// proves the inventory baseline, avoiding retroactive double counting.
			report.Quarantined++
		default:
			report.Quarantined++
		}
		reservationID := "legacy/allocation/" + allocation.AllocationId
		capacity := allocation.Assigned
		if state == types.ReservationStateQuarantined {
			capacity = types.ResourceCapacity{}
		}
		reservation := types.Reservation{ReservationId: reservationID, IdempotencyKey: reservationID, RequestId: allocation.RequestId, RequesterAddress: allocation.RequesterAddress, ProviderAddress: allocation.ProviderAddress, InventoryId: inventoryID, ResourceClass: allocation.ResourceClass, Capacity: capacity, State: state, ConsumerType: "legacy_allocation", ConsumerId: allocation.AllocationId, Version: 1, Reason: reason, LegacySource: "resources_allocation", LegacyReference: allocation.AllocationId, CreatedAt: allocation.CreatedAt, ActivatedAt: allocation.ActivatedAt, ExpiresAt: allocation.ExpiresAt, UpdatedAt: ctx.BlockTime(), CreatedHeight: allocation.BlockHeight}
		if err := k.SetReservation(ctx, reservation); err != nil {
			return report, err
		}
		k.setReservationIndexes(ctx, reservation)
		allocation.ReservationId = reservationID
		allocation.LegacySource = "resources_allocation"
		allocation.LegacyReference = allocation.AllocationId
		if err := k.SetAllocation(ctx, allocation); err != nil {
			return report, err
		}
		report.ReservationsCreated++
	}
	if err := k.ValidateCapacityConservation(ctx); err != nil && report.ReservationsCreated > 0 {
		return report, fmt.Errorf("post-migration reservation invariant: %w", err)
	}
	k.ActivateCanonicalReservations(ctx)
	return report, nil
}
