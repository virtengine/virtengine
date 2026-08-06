package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	marketplace "github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ValidateReservationLineage checks executable market/HPC work and duplicate owner state.
func (app *App) ValidateReservationLineage(ctx sdk.Context) error {
	var invariantErr error
	app.Keepers.VirtEngine.Market.WithLeases(ctx, func(lease marketv1.Lease) bool {
		if lease.State != marketv1.LeaseActive {
			return false
		}
		if lease.ReservationId == "" {
			invariantErr = fmt.Errorf("active lease %s has no reservation", lease.ID.String())
			return true
		}
		reservation, found := app.Keepers.VirtEngine.Resources.GetReservation(ctx, lease.ReservationId)
		legacyQuarantine := found && reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED && reservation.LegacySource != ""
		executable := reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE || reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_CONSUMED
		if !found || (!executable && !legacyQuarantine) || reservation.MarketLeaseId != lease.ID.String() {
			invariantErr = fmt.Errorf("active lease %s reservation mismatch", lease.ID.String())
			return true
		}
		return false
	})
	if invariantErr != nil {
		return invariantErr
	}
	app.Keepers.VirtEngine.HPC.WithJobs(ctx, func(job hpctypes.HPCJob) bool {
		if job.State != hpctypes.JobStateQueued && job.State != hpctypes.JobStateRunning {
			return false
		}
		reservation, found := app.Keepers.VirtEngine.Resources.GetReservation(ctx, job.ReservationID)
		jobMatches := found && reservation.HpcJobId == job.JobID
		leaseMatches := found && job.MarketLeaseID != "" && reservation.MarketLeaseId == job.MarketLeaseID
		legacyQuarantine := found && reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED && reservation.LegacySource != ""
		executable := reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE || reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_CONSUMED
		if !found || (!executable && !legacyQuarantine) || (!jobMatches && !leaseMatches) || reservation.ProviderAddress != job.ProviderAddress || (reservation.RequesterAddress != "" && reservation.RequesterAddress != job.CustomerAddress) {
			invariantErr = fmt.Errorf("executable HPC job %s reservation mismatch", job.JobID)
			return true
		}
		return false
	})
	if invariantErr != nil {
		return invariantErr
	}
	if app.Keepers.VirtEngine.Marketplace.IsCanonicalLifecycleActive(ctx) {
		app.Keepers.VirtEngine.Marketplace.WithOrders(ctx, func(order marketplace.Order) bool {
			if !order.State.IsTerminal() {
				reservation, found := app.Keepers.VirtEngine.Resources.GetReservationByConsumer(ctx, "legacy_marketplace_order", order.ID.String())
				if !validLegacyQuarantine(found, reservation, "mktplace_order", order.ID.String()) {
					invariantErr = fmt.Errorf("non-owner order %s remains mutable without quarantine after activation", order.ID.String())
					return true
				}
			}
			return false
		})
		if invariantErr != nil {
			return invariantErr
		}
		app.Keepers.VirtEngine.Marketplace.WithBids(ctx, func(bid marketplace.MarketplaceBid) bool {
			if !bid.State.IsTerminal() {
				reservation, found := app.Keepers.VirtEngine.Resources.GetReservationByConsumer(ctx, "legacy_marketplace_bid", bid.ID.String())
				if !validLegacyQuarantine(found, reservation, "mktplace_bid", bid.ID.String()) {
					invariantErr = fmt.Errorf("non-owner bid %s remains mutable without quarantine after activation", bid.ID.String())
					return true
				}
			}
			return false
		})
		if invariantErr != nil {
			return invariantErr
		}
		app.Keepers.VirtEngine.Marketplace.WithAllocations(ctx, func(allocation marketplace.Allocation) bool {
			if !allocation.State.IsTerminal() {
				reservation, found := app.Keepers.VirtEngine.Resources.GetReservationByConsumer(ctx, "legacy_marketplace_allocation", allocation.ID.String())
				if !validLegacyQuarantine(found, reservation, "mktplace_allocation", allocation.ID.String()) {
					invariantErr = fmt.Errorf("non-owner allocation %s remains mutable without quarantine after activation", allocation.ID.String())
					return true
				}
			}
			return false
		})
	}
	return invariantErr
}

func validLegacyQuarantine(found bool, reservation resourcesv1.Reservation, source, reference string) bool {
	return found && reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED &&
		reservation.LegacySource == source && reservation.LegacyReference == reference
}
