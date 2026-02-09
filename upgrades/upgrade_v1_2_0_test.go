package upgrades_test

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	bmetypes "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	oracletypes "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	"github.com/virtengine/virtengine/testutil/state"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	benchtypes "github.com/virtengine/virtengine/x/benchmark/types"
	configtypes "github.com/virtengine/virtengine/x/config/types"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
	reviewtypes "github.com/virtengine/virtengine/x/review/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	virtstakingtypes "github.com/virtengine/virtengine/x/staking/types"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const upgradeV120 = "v1.2.0"

func TestUpgradeV120InitializesNewModules(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	ctx := suite.Context()

	upgradeInit, ok := utypes.GetUpgradesList()[upgradeV120]
	require.True(t, ok, "upgrade %s not registered", upgradeV120)

	up, err := upgradeInit(log.NewNopLogger(), app.App)
	require.NoError(t, err)

	newModules := map[string]struct{}{
		veidtypes.ModuleName:        {},
		mfatypes.ModuleName:         {},
		encryptiontypes.ModuleName:  {},
		hpctypes.ModuleName:         {},
		settlementtypes.ModuleName:  {},
		delegationtypes.ModuleName:  {},
		rolestypes.ModuleName:       {},
		supporttypes.ModuleName:     {},
		enclavetypes.ModuleName:     {},
		fraudtypes.ModuleName:       {},
		benchtypes.ModuleName:       {},
		resourcestypes.ModuleName:   {},
		configtypes.ModuleName:      {},
		oracletypes.ModuleName:      {},
		reviewtypes.ModuleName:      {},
		bmetypes.ModuleName:         {},
		virtstakingtypes.ModuleName: {},
	}

	newStoreKeys := []string{
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
	}

	for _, storeKey := range newStoreKeys {
		clearStore(ctx, app.GetKey(storeKey))
	}

	fromVM := module.VersionMap{}
	for name, mod := range app.MM.Modules {
		if _, isNew := newModules[name]; isNew {
			continue
		}

		if versioned, ok := mod.(module.HasConsensusVersion); ok {
			fromVM[name] = versioned.ConsensusVersion()
		} else {
			fromVM[name] = 0
		}
	}

	_, err = up.UpgradeHandler()(sdk.WrapSDKContext(ctx), upgradetypes.Plan{Name: upgradeV120}, fromVM)
	require.NoError(t, err)

	require.Equal(t, veidtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.VEID.GetParams(ctx))
	require.Equal(t, mfatypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.MFA.GetParams(ctx))
	require.Equal(t, encryptiontypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Encryption.GetParams(ctx))
	require.Equal(t, hpctypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.HPC.GetParams(ctx))
	require.Equal(t, settlementtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Settlement.GetParams(ctx))
	require.Equal(t, delegationtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Delegation.GetParams(ctx))
	require.Equal(t, rolestypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Roles.GetParams(ctx))
	expectedSupport := normalizeSupportParams(supporttypes.DefaultGenesisState().Params)
	actualSupport := normalizeSupportParams(app.Keepers.VirtEngine.Support.GetParams(ctx))
	require.Equal(t, expectedSupport, actualSupport)
	require.Equal(t, enclavetypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Enclave.GetParams(ctx))
	require.Equal(t, fraudtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Fraud.GetParams(ctx))
	require.Equal(t, benchtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Benchmark.GetParams(ctx))
	require.Equal(t, resourcestypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Resources.GetParams(ctx))
	require.Equal(t, configtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Config.GetParams(ctx))
	require.Equal(t, reviewtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Review.GetParams(ctx))
	require.Equal(t, virtstakingtypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.VirtStaking.GetParams(ctx))
	require.Equal(t, bmetypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.BME.GetParams(ctx))
	require.Equal(t, oracletypes.DefaultGenesisState().Params, app.Keepers.VirtEngine.Oracle.GetParams(ctx))
}

func clearStore(ctx sdk.Context, storeKey *storetypes.KVStoreKey) {
	store := ctx.KVStore(storeKey)
	iter := store.Iterator(nil, nil)

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys = append(keys, key)
	}

	_ = iter.Close()

	for _, key := range keys {
		store.Delete(key)
	}
}

func normalizeSupportParams(params supporttypes.Params) supporttypes.Params {
	if params.AllowedExternalSystems == nil {
		params.AllowedExternalSystems = []string{}
	}
	if params.AllowedExternalDomains == nil {
		params.AllowedExternalDomains = []string{}
	}
	if params.SupportRecipientKeyIDs == nil {
		params.SupportRecipientKeyIDs = []string{}
	}
	return params
}
