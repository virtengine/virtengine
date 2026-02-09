// Package v1_2_0
// nolint revive
package v1_2_0

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	bmetypes "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	oracletypes "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	benchmark "github.com/virtengine/virtengine/x/benchmark"
	benchtypes "github.com/virtengine/virtengine/x/benchmark/types"
	bmekeeper "github.com/virtengine/virtengine/x/bme/keeper"
	config "github.com/virtengine/virtengine/x/config"
	configtypes "github.com/virtengine/virtengine/x/config/types"
	delegation "github.com/virtengine/virtengine/x/delegation"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	enclave "github.com/virtengine/virtengine/x/enclave"
	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
	encryption "github.com/virtengine/virtengine/x/encryption"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	fraud "github.com/virtengine/virtengine/x/fraud"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	hpc "github.com/virtengine/virtengine/x/hpc"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	mfa "github.com/virtengine/virtengine/x/mfa"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	oraclekeeper "github.com/virtengine/virtengine/x/oracle/keeper"
	resources "github.com/virtengine/virtengine/x/resources"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
	review "github.com/virtengine/virtengine/x/review"
	reviewtypes "github.com/virtengine/virtengine/x/review/types"
	roles "github.com/virtengine/virtengine/x/roles"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	settlement "github.com/virtengine/virtengine/x/settlement"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	virtstaking "github.com/virtengine/virtengine/x/staking"
	virtstakingtypes "github.com/virtengine/virtengine/x/staking/types"
	support "github.com/virtengine/virtengine/x/support"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
	veid "github.com/virtengine/virtengine/x/veid"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	UpgradeName = "v1.2.0"
)

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(log log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	up := &upgrade{
		App: app,
		log: log.With("module", fmt.Sprintf("upgrade/%s", UpgradeName)),
	}

	return up, nil
}

func (up *upgrade) StoreLoader() *storetypes.StoreUpgrades {
	return &storetypes.StoreUpgrades{
		Added: []string{
			veidtypes.StoreKey,
			mfatypes.StoreKey,
			encryptiontypes.StoreKey,
			hpctypes.StoreKey,
			settlementtypes.StoreKey,
			delegationtypes.StoreKey,
			rolestypes.StoreKey,
			supporttypes.StoreKey,
			enclavetypes.StoreKey,
			fraudtypes.StoreKey,
			benchtypes.StoreKey,
			resourcestypes.StoreKey,
			configtypes.StoreKey,
			oracletypes.StoreKey,
			reviewtypes.StoreKey,
			bmetypes.StoreKey,
			virtstakingtypes.StoreKey,
		},
	}
}

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		toVM, err := up.MM.RunMigrations(ctx, up.Configurator, fromVM)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		up.initNewModuleGeneses(sdkCtx)

		up.log.Info(fmt.Sprintf("all migrations for %s have been completed", UpgradeName))

		return toVM, nil
	}
}

func (up *upgrade) initNewModuleGeneses(ctx sdk.Context) {
	up.initIfEmpty(ctx, veidtypes.StoreKey, func() {
		veid.InitGenesis(ctx, up.Keepers.VirtEngine.VEID, veidtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, mfatypes.StoreKey, func() {
		mfa.InitGenesis(ctx, up.Keepers.VirtEngine.MFA, mfatypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, encryptiontypes.StoreKey, func() {
		encryption.InitGenesis(ctx, up.Keepers.VirtEngine.Encryption, encryptiontypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, hpctypes.StoreKey, func() {
		hpc.InitGenesis(ctx, up.Keepers.VirtEngine.HPC, hpctypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, settlementtypes.StoreKey, func() {
		settlement.InitGenesis(ctx, up.Keepers.VirtEngine.Settlement, settlementtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, delegationtypes.StoreKey, func() {
		delegation.InitGenesis(ctx, up.Keepers.VirtEngine.Delegation, delegationtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, rolestypes.StoreKey, func() {
		roles.InitGenesis(ctx, up.Keepers.VirtEngine.Roles, rolestypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, supporttypes.StoreKey, func() {
		support.InitGenesis(ctx, up.Keepers.VirtEngine.Support, supporttypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, enclavetypes.StoreKey, func() {
		enclave.InitGenesis(ctx, up.Keepers.VirtEngine.Enclave, enclavetypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, fraudtypes.StoreKey, func() {
		fraud.InitGenesis(ctx, up.Keepers.VirtEngine.Fraud, fraudtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, benchtypes.StoreKey, func() {
		benchmark.InitGenesis(ctx, up.Keepers.VirtEngine.Benchmark, benchtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, resourcestypes.StoreKey, func() {
		resources.InitGenesis(ctx, up.Keepers.VirtEngine.Resources, resourcestypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, configtypes.StoreKey, func() {
		config.InitGenesis(ctx, up.Keepers.VirtEngine.Config, configtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, reviewtypes.StoreKey, func() {
		review.InitGenesis(ctx, up.Keepers.VirtEngine.Review, reviewtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, virtstakingtypes.StoreKey, func() {
		virtstaking.InitGenesis(ctx, up.Keepers.VirtEngine.VirtStaking, virtstakingtypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, bmetypes.StoreKey, func() {
		bmekeeper.InitGenesis(ctx, up.Keepers.VirtEngine.BME, bmetypes.DefaultGenesisState())
	})
	up.initIfEmpty(ctx, oracletypes.StoreKey, func() {
		oraclekeeper.InitGenesis(ctx, up.Keepers.VirtEngine.Oracle, oracletypes.DefaultGenesisState())
	})
}

func (up *upgrade) initIfEmpty(ctx sdk.Context, storeKey string, initFn func()) {
	store := ctx.KVStore(up.GetKey(storeKey))
	iter := store.Iterator(nil, nil)
	defer func() {
		_ = iter.Close()
	}()

	if !iter.Valid() {
		initFn()
	}
}
