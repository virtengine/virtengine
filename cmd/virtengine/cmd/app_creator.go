package cmd

import (
	"errors"
	"io"
	"path/filepath"

	"cosmossdk.io/store"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	cflags "pkg.akt.dev/go/cli/flags"
	"pkg.akt.dev/go/sdkutil"

	virtengine "pkg.akt.dev/node/app"
)

type appCreator struct {
	encCfg sdkutil.EncodingConfig
}

func (a appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(cflags.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(cflags.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := sdkserver.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	homeDir := cast.ToString(appOpts.Get(cflags.FlagHome))
	chainID := cast.ToString(appOpts.Get(cflags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		genDocFile := filepath.Join(homeDir, cast.ToString(appOpts.Get("genesis_file")))
		appGenesis, err := genutiltypes.AppGenesisFromFile(genDocFile)
		if err != nil {
			panic(err)
		}

		chainID = appGenesis.ChainID
	}

	snapshotDir := filepath.Join(homeDir, "data", "snapshots")
	snapshotDB, err := dbm.NewDB("metadata", sdkserver.GetAppDBBackend(appOpts), snapshotDir)
	if err != nil {
		panic(err)
	}

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	// BaseApp Opts
	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(cflags.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(cflags.FlagStateSyncSnapshotKeepRecent)),
	)

	baseAppOptions := []func(*baseapp.BaseApp){
		baseapp.SetChainID(chainID),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(cflags.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(cflags.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(cflags.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(cflags.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(cflags.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(cflags.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(cflags.FlagIAVLCacheSize))),
	}

	return virtengine.NewApp(
		logger, db, traceStore, true, cast.ToUint(appOpts.Get(cflags.FlagInvCheckPeriod)), skipUpgradeHeights,
		a.encCfg,
		appOpts,
		baseAppOptions...,
	)
}

func (a appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	tio io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var VirtEngineApp *virtengine.VirtEngineApp

	homePath, ok := appOpts.Get(cflags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home is not set")
	}
	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}
	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(cflags.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	if height != -1 {
		VirtEngineApp = virtengine.NewApp(logger, db, tio, false, uint(1), map[int64]bool{}, a.encCfg, appOpts)

		if err := VirtEngineApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		VirtEngineApp = virtengine.NewApp(logger, db, tio, true, uint(1), map[int64]bool{}, a.encCfg, appOpts)
	}

	return VirtEngineApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// newTestnetApp starts by running the normal newApp method. From there, the app interface returned is modified in order
// for a testnet to be created from the provided app.
func (a appCreator) newTestnetApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	// Create an app and type cast to an VirtEngineApp
	app := a.newApp(logger, db, traceStore, appOpts)
	VirtEngineApp, ok := app.(*virtengine.VirtEngineApp)
	if !ok {
		panic("app created from newApp is not of type VirtEngineApp")
	}

	tcfg, valid := appOpts.Get(cflags.KeyTestnetConfig).(*virtengine.TestnetConfig)
	if !valid {
		panic("cflags.KeyTestnetConfig is not of type *virtengine.TestnetConfig")
	}

	// Make modifications to the normal VirtEngineApp required to run the network locally
	return virtengine.InitVirtEngineAppForTestnet(VirtEngineApp, db, tcfg)
}
