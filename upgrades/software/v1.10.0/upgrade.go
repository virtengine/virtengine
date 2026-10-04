// Package v1_10_0 introduces per-account assurance vectors in x/veid.
//
// The upgrade is additive and idempotent: it introduces the new state layout and
// deliberately back-fills NOTHING per account. See
// utypes.AssuranceVectorUpgradeName for why absence is the correct migration
// target rather than a zeroed vector.
package v1_10_0

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
)

const UpgradeName = utypes.AssuranceVectorUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{App: app, log: logger.With("module", fmt.Sprintf("upgrade/%s", UpgradeName))}, nil
}

// StoreLoader adds no new store: the assurance vector reuses the existing x/veid
// store key, so no KVStore or transient store is introduced and an empty loader
// is correct.
//
// Upgrading an existing chain needs no state migration for the new prefixes —
// absent keys already read as "no assurance claim" — which is exactly the
// behaviour this module wants.
func (up *upgrade) StoreLoader() *storetypes.StoreUpgrades { return &storetypes.StoreUpgrades{} }

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// No per-account writes. Existing accounts keep an absent vector until
		// they are re-verified, at which point ComputeAndStoreCompositeScore
		// derives one from real measurements.
		//
		// Writing a vector here derived from the scalar score would be a
		// consensus-visible fabrication: the per-factor values a relying party
		// reads (document, biometric, device, ...) were never measured, and
		// inventing them would let an unverified factor pass a factor-level
		// policy. So this handler is intentionally empty of state writes.
		up.log.Info("assurance vectors introduced; existing accounts hold no vector until re-verified",
			"upgrade", UpgradeName,
			"height", sdkCtx.BlockHeight(),
			"absence_means", "no_assurance_claim",
		)

		return cloneVersionMap(fromVM), nil
	}
}

// cloneVersionMap copies the module version map so the handler cannot mutate the
// caller's map.
func cloneVersionMap(source module.VersionMap) module.VersionMap {
	clone := make(module.VersionMap, len(source))
	for name, version := range source {
		clone[name] = version
	}
	return clone
}
