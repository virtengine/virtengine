// Package v1_4_0 activates deterministic ABCI++ admission without changing stores.
//
//nolint:revive
package v1_4_0

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

const UpgradeName = utypes.ConsensusAdmissionUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{
		App: app,
		log: logger.With("module", fmt.Sprintf("upgrade/%s", UpgradeName)),
	}, nil
}

func (up *upgrade) StoreLoader() *storetypes.StoreUpgrades {
	return &storetypes.StoreUpgrades{}
}

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		params, err := up.Keepers.Cosmos.ConsensusParams.ParamsStore.Get(ctx)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return nil, err
		}
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if params.Abci == nil {
			params.Abci = &cmtproto.ABCIParams{}
		}
		params.Abci.VoteExtensionsEnableHeight, err = voteExtensionActivationHeight(params.Abci.VoteExtensionsEnableHeight, sdkCtx.BlockHeight())
		if err != nil {
			return nil, err
		}
		if err := up.Keepers.Cosmos.ConsensusParams.ParamsStore.Set(ctx, params); err != nil {
			return nil, err
		}

		toVM, err := up.MM.RunMigrations(ctx, up.Configurator, fromVM)
		if err != nil {
			return nil, err
		}
		up.log.Info(fmt.Sprintf("all migrations for %s have been completed; deterministic proposal admission activates next height", UpgradeName))
		return toVM, nil
	}
}

func voteExtensionActivationHeight(current, upgradeHeight int64) (int64, error) {
	expected := upgradeHeight + 1
	if current == 0 {
		return expected, nil
	}
	if current != expected {
		return 0, fmt.Errorf("vote extensions must activate at H+1: got %d, expected %d", current, expected)
	}
	return current, nil
}
