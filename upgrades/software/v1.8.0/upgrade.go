// Package v1_8_0 activates Task 85B authenticated fiat observations.
package v1_8_0

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const UpgradeName = utypes.AuthenticatedFiatConversionsUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{App: app, log: logger.With("module", "upgrade/"+UpgradeName)}, nil
}

func (*upgrade) StoreLoader() *storetypes.StoreUpgrades { return &storetypes.StoreUpgrades{} }

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		cacheCtx, write := sdkCtx.CacheContext()
		report := struct {
			Scanned, TerminalPreserved, ActiveQuarantined uint64
			Digest                                        string
		}{}
		if fromVM[settlementtypes.ModuleName] < 4 {
			params := up.Keepers.VirtEngine.Settlement.GetParams(cacheCtx)
			params.FiatConversionEnabled = false
			params.FiatConversionDEXProfileID = ""
			params.FiatConversionDEXProfileDigest = nil
			params.FiatConversionDEXProfileState = settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED
			params.FiatConversionPayoutProfileID = ""
			params.FiatConversionPayoutProfileDigest = nil
			params.FiatConversionPayoutProfileState = settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED
			if err := up.Keepers.VirtEngine.Settlement.SetParams(cacheCtx, params); err != nil {
				return nil, err
			}
			migrationReport, err := up.Keepers.VirtEngine.Settlement.MigrateFiatConversions(cacheCtx)
			if err != nil {
				return nil, err
			}
			report.Scanned, report.TerminalPreserved, report.ActiveQuarantined, report.Digest = migrationReport.Scanned, migrationReport.TerminalPreserved, migrationReport.ActiveQuarantined, migrationReport.Digest
		}
		toVM, err := up.MM.RunMigrations(cacheCtx, up.Configurator, fromVM)
		if err != nil {
			return nil, err
		}
		if broken := up.Keepers.VirtEngine.Settlement.ValidateFiatConversionInvariants(cacheCtx); len(broken) != 0 {
			return nil, fmt.Errorf("post-upgrade fiat conversion invariant: %v", broken)
		}
		write()
		up.log.Info("authenticated fiat conversion protocol activated", "scanned", report.Scanned, "terminal_preserved", report.TerminalPreserved, "active_quarantined", report.ActiveQuarantined, "digest", report.Digest)
		return toVM, nil
	}
}
