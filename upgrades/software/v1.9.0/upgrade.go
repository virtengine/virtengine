// Package v1_9_0 enables ADR-010 deterministic market resolution through a
// governance upgrade once the canonical lifecycle and reservations are active.
package v1_9_0

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
)

const UpgradeName = utypes.UnifiedMarketResolutionUpgradeName

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
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// The resolver mints canonical reservations, so the capacity ledger
		// must already be authoritative.
		if !up.Keepers.VirtEngine.Resources.IsCanonicalReservationsActive(sdkCtx) {
			return nil, fmt.Errorf("%s precondition: canonical reservations are not active", UpgradeName)
		}
		if !up.Keepers.VirtEngine.Marketplace.IsCanonicalLifecycleActive(sdkCtx) {
			return nil, fmt.Errorf("%s precondition: mktplace canonical write fence is not active", UpgradeName)
		}
		if err := up.assertNoActiveLegacyLifecycle(sdkCtx); err != nil {
			return nil, err
		}

		params := up.Keepers.VirtEngine.Marketplace.GetParams(sdkCtx)
		if !params.EnableAutoResolution {
			params.EnableAutoResolution = true
			if err := up.Keepers.VirtEngine.Marketplace.SetParams(sdkCtx, params); err != nil {
				return nil, fmt.Errorf("%s: enable auto-resolution: %w", UpgradeName, err)
			}
			up.log.Info("deterministic market resolution enabled", "upgrade", UpgradeName)
		} else {
			up.log.Info("deterministic market resolution already enabled", "upgrade", UpgradeName)
		}

		return cloneVersionMap(fromVM), nil
	}
}

// assertNoActiveLegacyLifecycle fails closed when mutable legacy marketplace
// records could still disagree with the resolver.
func (up *upgrade) assertNoActiveLegacyLifecycle(ctx sdk.Context) error {
	keeper := up.Keepers.VirtEngine.Marketplace
	var offending string
	keeper.WithOrders(ctx, func(order marketplacetypes.Order) bool {
		if !order.State.IsTerminal() {
			offending = fmt.Sprintf("order %s in state %s", order.ID.String(), order.State.String())
			return true
		}
		return false
	})
	if offending != "" {
		return fmt.Errorf("%s precondition: non-terminal legacy %s", UpgradeName, offending)
	}
	keeper.WithBids(ctx, func(bid marketplacetypes.MarketplaceBid) bool {
		if !bid.State.IsTerminal() {
			offending = fmt.Sprintf("bid %s in state %s", bid.ID.String(), bid.State.String())
			return true
		}
		return false
	})
	if offending != "" {
		return fmt.Errorf("%s precondition: non-terminal legacy %s", UpgradeName, offending)
	}
	keeper.WithAllocations(ctx, func(allocation marketplacetypes.Allocation) bool {
		if !allocation.State.IsTerminal() {
			offending = fmt.Sprintf("allocation %s in state %s", allocation.ID.String(), allocation.State.String())
			return true
		}
		return false
	})
	if offending != "" {
		return fmt.Errorf("%s precondition: non-terminal legacy %s", UpgradeName, offending)
	}
	return nil
}

func cloneVersionMap(source module.VersionMap) module.VersionMap {
	clone := make(module.VersionMap, len(source))
	for name, version := range source {
		clone[name] = version
	}
	return clone
}
