// Package v1_5_0 activates Task 84B through registered module migrations.
//
//nolint:revive
package v1_5_0

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

const UpgradeName = utypes.AuthenticatedMeteringUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{App: app, log: logger.With("module", fmt.Sprintf("upgrade/%s", UpgradeName))}, nil
}

func (up *upgrade) StoreLoader() *storetypes.StoreUpgrades {
	return &storetypes.StoreUpgrades{}
}

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		toVM, err := up.MM.RunMigrations(ctx, up.Configurator, fromVM)
		if err != nil {
			return nil, err
		}
		up.log.Info("authenticated usage metering and provider signing-key epochs activated")
		return toVM, nil
	}
}
