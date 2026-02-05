package app

import (
	"cosmossdk.io/x/evidence"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"

	"github.com/virtengine/virtengine/x/audit"
	"github.com/virtengine/virtengine/x/benchmark"
	"github.com/virtengine/virtengine/x/bme"
	"github.com/virtengine/virtengine/x/cert"
	"github.com/virtengine/virtengine/x/config"
	"github.com/virtengine/virtengine/x/delegation"
	"github.com/virtengine/virtengine/x/deployment"
	"github.com/virtengine/virtengine/x/enclave"
	"github.com/virtengine/virtengine/x/encryption"
	"github.com/virtengine/virtengine/x/escrow"
	"github.com/virtengine/virtengine/x/fraud"
	"github.com/virtengine/virtengine/x/hpc"
	"github.com/virtengine/virtengine/x/market"
	"github.com/virtengine/virtengine/x/marketplace"
	"github.com/virtengine/virtengine/x/mfa"
	"github.com/virtengine/virtengine/x/oracle"
	"github.com/virtengine/virtengine/x/provider"
	"github.com/virtengine/virtengine/x/review"
	"github.com/virtengine/virtengine/x/roles"
	"github.com/virtengine/virtengine/x/settlement"
	virtstaking "github.com/virtengine/virtengine/x/staking"
	"github.com/virtengine/virtengine/x/support"
	"github.com/virtengine/virtengine/x/take"
	"github.com/virtengine/virtengine/x/veid"
)

func appModules(
	app *VirtEngineApp,
	encodingConfig sdkutil.EncodingConfig,
) []module.AppModule {
	cdc := encodingConfig.Codec

	return []module.AppModule{
		genutil.NewAppModule(
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Staking,
			app.BaseApp,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Acct,
			nil,
			app.GetSubspace(authtypes.ModuleName),
		),
		vesting.NewAppModule(
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
		),
		bank.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Acct,
			app.GetSubspace(banktypes.ModuleName),
		),
		gov.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Gov,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.GetSubspace(govtypes.ModuleName),
		),
		mint.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Mint,
			app.Keepers.Cosmos.Acct,
			nil, // todo virtengine/support#4
			app.GetSubspace(minttypes.ModuleName),
		),
		slashing.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Slashing,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Staking,
			app.GetSubspace(slashingtypes.ModuleName),
			encodingConfig.InterfaceRegistry,
		),
		distr.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Distr,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Staking,
			app.GetSubspace(distrtypes.ModuleName),
		),
		staking.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Staking,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.GetSubspace(stakingtypes.ModuleName),
		),
		upgrade.NewAppModule(
			app.Keepers.Cosmos.Upgrade,
			addresscodec.NewBech32Codec(sdkutil.Bech32PrefixAccAddr),
		),
		evidence.NewAppModule(
			*app.Keepers.Cosmos.Evidence,
		),
		authzmodule.NewAppModule(
			cdc, app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.interfaceRegistry,
		),
		feegrantmodule.NewAppModule(
			cdc,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.FeeGrant,
			app.interfaceRegistry,
		),
		ibc.NewAppModule(
			app.Keepers.Cosmos.IBC,
		),
		transfer.NewAppModule(
			app.Keepers.Cosmos.Transfer,
		),
		ibctm.NewAppModule(
			app.Keepers.Modules.TMLight,
		),
		params.NewAppModule( //nolint: staticcheck
			app.Keepers.Cosmos.Params,
		),
		consensus.NewAppModule(
			cdc,
			*app.Keepers.Cosmos.ConsensusParams,
		),
		// virtengine modules
		take.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Take,
		),
		escrow.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Escrow,
			app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Bank,
		),
		deployment.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Deployment,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.Escrow,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Authz,
		),
		market.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.Escrow,
			app.Keepers.VirtEngine.Audit,
			app.Keepers.VirtEngine.Deployment,
			app.Keepers.VirtEngine.Provider,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Bank,
		),
		marketplace.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Marketplace,
		),
		provider.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Provider,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.VEID,
			app.Keepers.VirtEngine.MFA,
		),
		audit.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Audit,
		),
		cert.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Cert,
		),
		// VirtEngine patent-specific modules (AU2024203136A1)
		encryption.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Encryption,
		),
		roles.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Roles,
		),
		support.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Support,
		),
		veid.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.VEID,
		),
		mfa.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.MFA,
		),
		config.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Config,
		),
		hpc.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.HPC,
		),
		benchmark.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Benchmark,
		),
		enclave.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Enclave,
		),
		settlement.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Settlement,
		),
		fraud.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Fraud,
		),
		review.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Review,
		),
		delegation.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Delegation,
		),
		virtstaking.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.VirtStaking,
		),
		bme.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.BME,
		),
		oracle.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Oracle,
		),
	}
}

func appSimModules(
	app *VirtEngineApp,
	encodingConfig sdkutil.EncodingConfig,
) []module.AppModuleSimulation {
	return []module.AppModuleSimulation{
		auth.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Acct,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
		authzmodule.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.interfaceRegistry,
		),
		bank.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Acct,
			app.GetSubspace(banktypes.ModuleName),
		),
		feegrantmodule.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.FeeGrant,
			app.interfaceRegistry,
		),
		authzmodule.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.interfaceRegistry,
		),
		gov.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Gov,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.GetSubspace(govtypes.ModuleName),
		),
		mint.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Mint,
			app.Keepers.Cosmos.Acct,
			nil, // todo virtengine/support#4
			app.GetSubspace(minttypes.ModuleName),
		),
		staking.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Staking,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.GetSubspace(stakingtypes.ModuleName),
		),
		distr.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Distr,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Staking,
			app.GetSubspace(distrtypes.ModuleName),
		),
		slashing.NewAppModule(
			app.cdc,
			app.Keepers.Cosmos.Slashing,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Staking,
			app.GetSubspace(slashingtypes.ModuleName),
			encodingConfig.InterfaceRegistry,
		),
		params.NewAppModule( //nolint: staticcheck
			app.Keepers.Cosmos.Params,
		),
		evidence.NewAppModule(
			*app.Keepers.Cosmos.Evidence,
		),
		ibc.NewAppModule(
			app.Keepers.Cosmos.IBC,
		),
		transfer.NewAppModule(
			app.Keepers.Cosmos.Transfer,
		),
		// virtengine sim modules
		take.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Take,
		),

		deployment.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Deployment,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.Escrow,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.Cosmos.Authz,
		),

		market.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.Escrow,
			app.Keepers.VirtEngine.Audit,
			app.Keepers.VirtEngine.Deployment,
			app.Keepers.VirtEngine.Provider,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Authz,
			app.Keepers.Cosmos.Bank,
		),
		marketplace.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Marketplace,
		),

		provider.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Provider,
			app.Keepers.Cosmos.Acct,
			app.Keepers.Cosmos.Bank,
			app.Keepers.VirtEngine.Market,
			app.Keepers.VirtEngine.VEID,
			app.Keepers.VirtEngine.MFA,
		),

		cert.NewAppModule(
			app.cdc,
			app.Keepers.VirtEngine.Cert,
		),
	}
}
