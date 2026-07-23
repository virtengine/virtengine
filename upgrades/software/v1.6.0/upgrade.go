// Package v1_6_0 activates Task 84C through registered module migrations.
package v1_6_0

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
)

const UpgradeName = utypes.CanonicalReservationsUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{App: app, log: logger.With("module", fmt.Sprintf("upgrade/%s", UpgradeName))}, nil
}
func (up *upgrade) StoreLoader() *storetypes.StoreUpgrades { return &storetypes.StoreUpgrades{} }
func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		if upgradeAlreadyApplied(fromVM) {
			return cloneVersionMap(fromVM), nil
		}
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if up.Keepers.VirtEngine.Marketplace.IsCanonicalLifecycleActive(sdkCtx) {
			return nil, fmt.Errorf("%s precondition: mktplace canonical write fence already active", UpgradeName)
		}
		report, err := up.reconcileLegacyLifecycle(sdkCtx)
		if err != nil {
			return nil, err
		}
		toVM, err := up.MM.RunMigrations(ctx, up.Configurator, fromVM)
		if err != nil {
			return nil, err
		}
		// This historical handler must not pre-apply later Task 84D migrations
		// merely because the running binary knows newer module versions.
		toVM[resourcetypes.ModuleName] = 2
		toVM[hpctypes.ModuleName] = 2
		if err := up.Keepers.VirtEngine.Resources.ValidateCapacityConservation(sdkCtx); err != nil {
			return nil, fmt.Errorf("post-upgrade capacity invariant: %w", err)
		}
		if err := up.ValidateReservationLineage(sdkCtx); err != nil {
			return nil, fmt.Errorf("post-upgrade lineage invariant: %w", err)
		}
		up.log.Info("canonical x/market lifecycle and authoritative x/resources reservations activated", "market_leases_scanned", report.MarketLeasesScanned, "hpc_jobs_scanned", report.HPCJobsScanned, "marketplace_orders_scanned", report.MarketplaceOrdersScanned, "marketplace_bids_scanned", report.MarketplaceBidsScanned, "marketplace_allocations_scanned", report.MarketplaceAllocationsScanned, "quarantined", report.Quarantined, "already_linked", report.AlreadyLinked, "terminal_preserved", report.TerminalPreserved)
		return toVM, nil
	}
}

func upgradeAlreadyApplied(fromVM module.VersionMap) bool {
	return fromVM[marketv1.ModuleName] >= 8 && fromVM[marketplacetypes.ModuleName] >= 2 &&
		fromVM[resourcetypes.ModuleName] >= 2 && fromVM[hpctypes.ModuleName] >= 2
}

func cloneVersionMap(source module.VersionMap) module.VersionMap {
	clone := make(module.VersionMap, len(source))
	for name, version := range source {
		clone[name] = version
	}
	return clone
}

type reconciliationReport struct {
	MarketLeasesScanned, HPCJobsScanned, MarketplaceOrdersScanned, MarketplaceBidsScanned, MarketplaceAllocationsScanned uint64
	Quarantined, AlreadyLinked, TerminalPreserved                                                                        uint64
}

func (up *upgrade) reconcileLegacyLifecycle(ctx sdk.Context) (reconciliationReport, error) {
	report := reconciliationReport{}
	var reconcileErr error
	up.Keepers.VirtEngine.Market.WithLeases(ctx, func(lease marketv1.Lease) bool {
		report.MarketLeasesScanned++
		if lease.State != marketv1.LeaseActive {
			report.TerminalPreserved++
			return false
		}
		if lease.ReservationId != "" {
			reservation, found := up.Keepers.VirtEngine.Resources.GetReservation(ctx, lease.ReservationId)
			if found && reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE &&
				reservation.ProviderAddress == lease.ID.Provider && reservation.RequesterAddress == lease.ID.Owner &&
				reservation.MarketLeaseId == lease.ID.String() {
				report.AlreadyLinked++
				return false
			}
			quarantine, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "market_lease_inconsistent", lease.ID.String(), lease.ID.Provider, "market_lease", lease.ID.String(), resourcesv1.ReservationLink{MarketOrderId: lease.ID.OrderID().String(), MarketBidId: lease.ID.BidID().String(), MarketLeaseId: lease.ID.String(), EscrowId: "market/payment/" + lease.ID.String()}, "legacy_active_lease_reservation_inconsistent")
			if err != nil {
				reconcileErr = err
				return true
			}
			if created {
				report.Quarantined++
			} else {
				report.AlreadyLinked++
			}
			if err := up.Keepers.VirtEngine.Market.SetLegacyReservationLinks(ctx, lease.ID, quarantine.ReservationId); err != nil {
				reconcileErr = err
				return true
			}
			return false
		}
		reservation, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "market_lease", lease.ID.String(), lease.ID.Provider, "market_lease", lease.ID.String(), resourcesv1.ReservationLink{MarketOrderId: lease.ID.OrderID().String(), MarketBidId: lease.ID.BidID().String(), MarketLeaseId: lease.ID.String(), EscrowId: "market/payment/" + lease.ID.String()}, "legacy_active_lease_requires_inventory_reconciliation")
		if err != nil {
			reconcileErr = err
			return true
		}
		if created {
			report.Quarantined++
		} else {
			report.AlreadyLinked++
		}
		if err := up.Keepers.VirtEngine.Market.SetLegacyReservationLinks(ctx, lease.ID, reservation.ReservationId); err != nil {
			reconcileErr = err
			return true
		}
		return false
	})
	if reconcileErr != nil {
		return report, reconcileErr
	}
	up.Keepers.VirtEngine.HPC.WithJobs(ctx, func(job hpctypes.HPCJob) bool {
		report.HPCJobsScanned++
		if hpctypes.IsTerminalJobState(job.State) {
			report.TerminalPreserved++
			return false
		}
		if job.ReservationID != "" {
			reservation, found := up.Keepers.VirtEngine.Resources.GetReservation(ctx, job.ReservationID)
			jobMatches := found && reservation.HpcJobId == job.JobID
			leaseMatches := found && job.MarketLeaseID != "" && reservation.MarketLeaseId == job.MarketLeaseID
			if found && reservation.State == resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE &&
				(jobMatches || leaseMatches) && reservation.ProviderAddress == job.ProviderAddress &&
				reservation.RequesterAddress == job.CustomerAddress {
				report.AlreadyLinked++
				return false
			}
			quarantine, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "hpc_job_inconsistent", job.JobID, job.ProviderAddress, "hpc_job", job.JobID, resourcesv1.ReservationLink{HpcJobId: job.JobID, MarketOrderId: job.MarketOrderID, MarketBidId: job.MarketBidID, MarketLeaseId: job.MarketLeaseID, EscrowId: job.EscrowID}, "legacy_hpc_job_reservation_inconsistent")
			if err != nil {
				reconcileErr = err
				return true
			}
			if created {
				report.Quarantined++
			} else {
				report.AlreadyLinked++
			}
			if err := up.Keepers.VirtEngine.HPC.SetLegacyReservationLink(ctx, job.JobID, quarantine.ReservationId); err != nil {
				reconcileErr = err
				return true
			}
			return false
		}
		reservation, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "hpc_job", job.JobID, job.ProviderAddress, "hpc_job", job.JobID, resourcesv1.ReservationLink{HpcJobId: job.JobID, MarketOrderId: job.MarketOrderID, MarketBidId: job.MarketBidID, MarketLeaseId: job.MarketLeaseID, EscrowId: job.EscrowID}, "legacy_hpc_job_requires_inventory_reconciliation")
		if err != nil {
			reconcileErr = err
			return true
		}
		if created {
			report.Quarantined++
		} else {
			report.AlreadyLinked++
		}
		if err := up.Keepers.VirtEngine.HPC.SetLegacyReservationLink(ctx, job.JobID, reservation.ReservationId); err != nil {
			reconcileErr = err
			return true
		}
		return false
	})
	if reconcileErr != nil {
		return report, reconcileErr
	}
	up.Keepers.VirtEngine.Marketplace.WithOrders(ctx, func(order marketplacetypes.Order) bool {
		report.MarketplaceOrdersScanned++
		if order.State.IsTerminal() {
			report.TerminalPreserved++
			return false
		}
		_, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "mktplace_order", order.ID.String(), order.OfferingID.ProviderAddress, "legacy_marketplace_order", order.ID.String(), resourcesv1.ReservationLink{}, "non_owner_order_requires_canonical_reconciliation")
		if err != nil {
			reconcileErr = err
			return true
		}
		if created {
			report.Quarantined++
		} else {
			report.AlreadyLinked++
		}
		return false
	})
	if reconcileErr != nil {
		return report, reconcileErr
	}
	up.Keepers.VirtEngine.Marketplace.WithBids(ctx, func(bid marketplacetypes.MarketplaceBid) bool {
		report.MarketplaceBidsScanned++
		if bid.State.IsTerminal() {
			report.TerminalPreserved++
			return false
		}
		_, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "mktplace_bid", bid.ID.String(), bid.ID.ProviderAddress, "legacy_marketplace_bid", bid.ID.String(), resourcesv1.ReservationLink{}, "non_owner_bid_requires_canonical_reconciliation")
		if err != nil {
			reconcileErr = err
			return true
		}
		if created {
			report.Quarantined++
		} else {
			report.AlreadyLinked++
		}
		return false
	})
	if reconcileErr != nil {
		return report, reconcileErr
	}
	up.Keepers.VirtEngine.Marketplace.WithAllocations(ctx, func(allocation marketplacetypes.Allocation) bool {
		report.MarketplaceAllocationsScanned++
		if allocation.State.IsTerminal() {
			report.TerminalPreserved++
			return false
		}
		_, created, err := up.Keepers.VirtEngine.Resources.ImportLegacyQuarantine(ctx, "mktplace_allocation", allocation.ID.String(), allocation.ProviderAddress, "legacy_marketplace_allocation", allocation.ID.String(), resourcesv1.ReservationLink{}, "non_owner_allocation_requires_canonical_reconciliation")
		if err != nil {
			reconcileErr = err
			return true
		}
		if created {
			report.Quarantined++
		} else {
			report.AlreadyLinked++
		}
		return false
	})
	return report, reconcileErr
}
