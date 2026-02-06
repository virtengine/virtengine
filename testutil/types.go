package testutil

import (
	"encoding/json"
	"fmt"
	"os"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"

	"github.com/virtengine/virtengine/app"
	"github.com/virtengine/virtengine/testutil/network"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

// NewTestNetworkFixture returns a new simapp AppConstructor for network simulation tests
func NewTestNetworkFixture(opts ...network.TestnetFixtureOption) network.TestFixture {
	dir, err := os.MkdirTemp("", "simapp")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cfgOpts := &network.TestnetFixtureOptions{}

	for _, opt := range opts {
		opt(cfgOpts)
	}

	if cfgOpts.EncCfg.InterfaceRegistry == nil {
		cfgOpts.EncCfg = sdkutil.MakeEncodingConfig()
		app.ModuleBasics().RegisterInterfaces(cfgOpts.EncCfg.InterfaceRegistry)
	}

	tapp := app.NewApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		0,
		make(map[int64]bool),
		cfgOpts.EncCfg,
		simtestutil.NewAppOptionsWithFlagHome(dir),
	)

	appCtr := func(val network.ValidatorI) servertypes.Application {
		return app.NewApp(
			val.GetCtx().Logger,
			dbm.NewMemDB(),
			nil,
			true,
			0,
			make(map[int64]bool),
			cfgOpts.EncCfg,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			bam.SetChainID(val.GetCtx().Viper.GetString(cflags.FlagChainID)),
		)
	}

	genesisState := app.NewDefaultGenesisState(tapp.AppCodec())
	if mfaGenesisBz, ok := genesisState[mfatypes.ModuleName]; ok {
		var mfaGenesis mfatypes.GenesisState
		if err := json.Unmarshal(mfaGenesisBz, &mfaGenesis); err != nil {
			panic(fmt.Sprintf("failed to unmarshal mfa genesis: %v", err))
		}
		for i, config := range mfaGenesis.SensitiveTxConfigs {
			if config.TransactionType != mfatypes.SensitiveTxAccountRecovery {
				config.Enabled = false
			}
			mfaGenesis.SensitiveTxConfigs[i] = config
		}
		mfaGenesisBz, err := json.Marshal(&mfaGenesis)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal mfa genesis: %v", err))
		}
		genesisState[mfatypes.ModuleName] = mfaGenesisBz
	}

	return network.TestFixture{
		AppConstructor: appCtr,
		GenesisState:   genesisState,
		EncodingConfig: cfgOpts.EncCfg,
	}
}
