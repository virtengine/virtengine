package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"

	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	cmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	audittypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	certtypes "github.com/virtengine/virtengine/sdk/go/node/cert/v1"
	deploymenttypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	taketypes "github.com/virtengine/virtengine/sdk/go/node/take/v1"

	bmetypes "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	benchmarktypes "github.com/virtengine/virtengine/x/benchmark/types"
	configtypes "github.com/virtengine/virtengine/x/config/types"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	oracletypes "github.com/virtengine/virtengine/x/oracle/types"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
	reviewtypes "github.com/virtengine/virtengine/x/review/types"
	rolestypes "github.com/virtengine/virtengine/x/roles/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
	virtstakingtypes "github.com/virtengine/virtengine/x/staking/types"
	supporttypes "github.com/virtengine/virtengine/x/support/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/virtengine/virtengine/client/docs/statik"
)

const (
	AppName = "virtengine"
)

var (
	DefaultHome = os.ExpandEnv("$HOME/.virtengine")

	_ runtime.AppI            = (*VirtEngineApp)(nil)
	_ servertypes.Application = (*VirtEngineApp)(nil)

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{}
)

// VirtEngineApp extends ABCI application
type VirtEngineApp struct {
	*baseapp.BaseApp
	*apptypes.App

	aminoCdc          *codec.LegacyAmino
	cdc               codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry
	sm                *module.SimulationManager
	invCheckPeriod    uint
}

// NewApp creates and returns a new VirtEngine App.
func NewApp(
	logger log.Logger,
	db dbm.DB,
	tio io.Writer,
	loadLatest bool,
	invCheckPeriod uint,
	skipUpgradeHeights map[int64]bool,
	encodingConfig sdkutil.EncodingConfig,
	appOpts servertypes.AppOptions,
	options ...func(*baseapp.BaseApp),
) *VirtEngineApp {
	appCodec := encodingConfig.Codec
	aminoCdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bapp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), options...)

	bapp.SetCommitMultiStoreTracer(tio)
	bapp.SetVersion(version.Version)
	bapp.SetInterfaceRegistry(interfaceRegistry)
	bapp.SetTxEncoder(txConfig.TxEncoder())

	homePath := cast.ToString(appOpts.Get(cflags.FlagHome))
	if homePath == "" {
		homePath = DefaultHome
	}

	app := &VirtEngineApp{
		BaseApp: bapp,
		App: &apptypes.App{
			Cdc: appCodec,
			Log: logger,
		},
		aminoCdc:          aminoCdc,
		cdc:               appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	app.InitSpecialKeepers(
		app.cdc,
		aminoCdc,
		app.BaseApp,
		skipUpgradeHeights,
		homePath,
	)

	app.InitNormalKeepers(
		app.cdc,
		encodingConfig,
		app.BaseApp,
		ModuleAccountPerms(),
		app.BlockedAddrs(),
		invCheckPeriod,
	)

	// TODO: There is a bug here, where we register the govRouter routes in InitNormalKeepers and then
	// call setupHooks afterwards. Therefore, if a gov proposal needs to call a method and that method calls a
	// hook, we will get a nil pointer dereference error due to the hooks in the keeper not being
	// setup yet. I will refrain from creating an issue in the sdk for now until after we unfork to 0.47,
	// because I believe the concept of Routes is going away.
	app.SetupHooks()

	// NOTE: All module / keeper changes should happen prior to this module.NewManager line being called.
	// However, in the event any changes do need to happen after this call, ensure that that keeper
	// is only passed in its keeper form (not de-ref'd anywhere)
	//
	// Generally NewAppModule will require the keeper that module defines to be passed in as an exact struct,
	// but should take in every other keeper as long as it matches a certain interface. (So no need to be de-ref'd)
	//
	// Any time a module requires a keeper de-ref'd that's not its native one,
	// its code-smell and should probably change. We should get the staking keeper dependencies fixed.
	modules := appModules(app, encodingConfig)

	app.MM = module.NewManager(modules...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's begin-blocker must come before any modules using capabilities (e.g. IBC)

	// Upgrades from v0.50.x onwards happen in pre block
	app.MM.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authtypes.ModuleName,
	)

	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.MM.SetOrderBeginBlockers(orderBeginBlockers(app.MM.ModuleNames())...)
	app.MM.SetOrderEndBlockers(OrderEndBlockers(app.MM.ModuleNames())...)
	app.MM.SetOrderInitGenesis(OrderInitGenesis(app.MM.ModuleNames())...)

	app.Configurator = module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.MM.RegisterServices(app.Configurator)
	if err != nil {
		panic(err)
	}

	// register the upgrade handler
	if err := app.registerUpgradeHandlers(); err != nil {
		panic(err)
	}

	app.sm = module.NewSimulationManager(appSimModules(app, encodingConfig)...)
	app.sm.RegisterStoreDecoders()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.MM.Modules))

	reflectionSvc := getReflectionService()
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())

	anteOpts := HandlerOptions{
		HandlerOptions: ante.HandlerOptions{
			AccountKeeper:   app.Keepers.Cosmos.Acct,
			BankKeeper:      app.Keepers.Cosmos.Bank,
			FeegrantKeeper:  app.Keepers.Cosmos.FeeGrant,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
		CDC:             app.cdc,
		GovKeeper:       app.Keepers.Cosmos.Gov,
		MFAGatingKeeper: &app.Keepers.VirtEngine.MFA,
		VEIDKeeper:      &app.Keepers.VirtEngine.VEID,
		RolesKeeper:     &app.Keepers.VirtEngine.Roles,
	}

	anteHandler, err := NewAnteHandler(anteOpts)
	if err != nil {
		panic(err)
	}

	app.SetPrepareProposal(baseapp.NoOpPrepareProposal())

	// we use a no-op ProcessProposal, this way, we accept all proposals in avoidance
	// of liveness failures due to Prepare / Process inconsistency. In other words,
	// this ProcessProposal always returns ACCEPT.
	app.SetProcessProposal(baseapp.NoOpProcessProposal())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)
	app.SetPrecommiter(app.Precommitter)
	app.SetPrepareCheckStater(app.PrepareCheckStater)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			cmos.Exit("app initialization:" + err.Error())
		}
	}

	return app
}

// orderBeginBlockers returns the order of BeginBlockers, by module name.
func orderBeginBlockers(_ []string) []string {
	return []string{
		upgradetypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		deploymenttypes.ModuleName,
		govtypes.ModuleName,
		providertypes.ModuleName,
		certtypes.ModuleName,
		markettypes.ModuleName,
		marketplacetypes.ModuleName,
		audittypes.ModuleName,
		genutiltypes.ModuleName,
		vestingtypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		taketypes.ModuleName,
		emodule.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		transfertypes.ModuleName,
		consensusparamtypes.ModuleName,
		ibctm.ModuleName,
		ibchost.ModuleName,
		feegrant.ModuleName,
		// VirtEngine patent modules (AU2024203136A1)
		veidtypes.ModuleName,
		mfatypes.ModuleName,
		encryptiontypes.ModuleName,
		rolestypes.ModuleName,
		supporttypes.ModuleName,
		configtypes.ModuleName,
		hpctypes.ModuleName,
		resourcestypes.ModuleName,
		benchmarktypes.ModuleName,
		enclavetypes.ModuleName,
		settlementtypes.ModuleName,
		fraudtypes.ModuleName,
		reviewtypes.ModuleName,
		delegationtypes.ModuleName,
		virtstakingtypes.ModuleName,
		bmetypes.ModuleName,
		oracletypes.ModuleName,
	}
}

// OrderEndBlockers returns EndBlockers (crisis, govtypes, staking) with no relative order.
func OrderEndBlockers(_ []string) []string {
	return []string{
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		upgradetypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		deploymenttypes.ModuleName,
		providertypes.ModuleName,
		certtypes.ModuleName,
		markettypes.ModuleName,
		marketplacetypes.ModuleName,
		audittypes.ModuleName,
		genutiltypes.ModuleName,
		vestingtypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		taketypes.ModuleName,
		emodule.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		transfertypes.ModuleName,
		ibchost.ModuleName,
		feegrant.ModuleName,
		consensusparamtypes.ModuleName,
		ibctm.ModuleName,
		// VirtEngine patent modules (AU2024203136A1)
		veidtypes.ModuleName,
		mfatypes.ModuleName,
		encryptiontypes.ModuleName,
		rolestypes.ModuleName,
		supporttypes.ModuleName,
		configtypes.ModuleName,
		hpctypes.ModuleName,
		resourcestypes.ModuleName,
		benchmarktypes.ModuleName,
		enclavetypes.ModuleName,
		settlementtypes.ModuleName,
		fraudtypes.ModuleName,
		reviewtypes.ModuleName,
		delegationtypes.ModuleName,
		virtstakingtypes.ModuleName,
		bmetypes.ModuleName,
		oracletypes.ModuleName,
	}
}

func getGenesisTime(appOpts servertypes.AppOptions, homePath string) time.Time { // nolint: unused
	if v := appOpts.Get("GenesisTime"); v != nil {
		// in tests, GenesisTime is supplied using appOpts
		genTime, ok := v.(time.Time)
		if !ok {
			panic("expected GenesisTime to be a Time value")
		}
		return genTime
	}

	genDoc, err := tmtypes.GenesisDocFromFile(filepath.Join(homePath, "config/genesis.json"))
	if err != nil {
		panic(err)
	}

	return genDoc.GenesisTime
}

// Name returns the name of the App
func (app *VirtEngineApp) Name() string { return app.BaseApp.Name() }

// InitChainer application update at chain initialization
func (app *VirtEngineApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	err := app.Keepers.Cosmos.Upgrade.SetModuleVersionMap(ctx, app.MM.GetVersionMap())
	if err != nil {
		panic(err)
	}

	return app.MM.InitGenesis(ctx, app.cdc, genesisState)
}

// PreBlocker application updates before each begin block.
func (app *VirtEngineApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	// Set gas meter to the free gas meter.
	// This is because there is currently non-deterministic gas usage in the
	// pre-blocker, e.g. due to hydration of in-memory data structures.
	//
	// Note that we don't need to reset the gas meter after the pre-blocker
	// because Go is pass by value.
	ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	return app.MM.PreBlock(ctx)
}

// BeginBlocker is a function in which application updates every begin block
func (app *VirtEngineApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	if patch, exists := utypes.GetHeightPatchesList()[ctx.BlockHeight()]; exists {
		app.Logger().Info(fmt.Sprintf("found patch %s for current height %d. applying...", patch.Name(), ctx.BlockHeight()))
		patch.Begin(ctx, &app.Keepers)
		app.Logger().Info(fmt.Sprintf("patch %s applied successfully at height %d", patch.Name(), ctx.BlockHeight()))
	}

	return app.MM.BeginBlock(ctx)
}

// EndBlocker is a function in which application updates every end block
func (app *VirtEngineApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.MM.EndBlock(ctx)
}

// Precommitter application updates before the commital of a block after all transactions have been delivered.
func (app *VirtEngineApp) Precommitter(ctx sdk.Context) {
	if err := app.MM.Precommit(ctx); err != nil {
		panic(err)
	}
}

func (app *VirtEngineApp) PrepareCheckStater(ctx sdk.Context) {
	if err := app.MM.PrepareCheckState(ctx); err != nil {
		panic(err)
	}
}

// LegacyAmino returns VirtEngineApp's amino codec.
func (app *VirtEngineApp) LegacyAmino() *codec.LegacyAmino {
	return app.aminoCdc
}

// AppCodec returns VirtEngineApp's app codec.
func (app *VirtEngineApp) AppCodec() codec.Codec {
	return app.cdc
}

// TxConfig returns SimApp's TxConfig
func (app *VirtEngineApp) TxConfig() client.TxConfig {
	return app.txConfig
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *VirtEngineApp) ModuleAccountAddrs() map[string]bool {
	return ModuleAccountAddrs()
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *VirtEngineApp) BlockedAddrs() map[string]bool {
	perms := ModuleAccountAddrs()
	blockedAddrs := make(map[string]bool)
	for acc := range perms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

// InterfaceRegistry returns VirtEngineApp's InterfaceRegistry
func (app *VirtEngineApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetSubspace returns a param subspace for a given module name.
func (app *VirtEngineApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.Keepers.Cosmos.Params.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *VirtEngineApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *VirtEngineApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	cctx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway
	authtx.RegisterGRPCGatewayRoutes(cctx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(cctx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics().RegisterGRPCGatewayRoutes(cctx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(cctx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(cctx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *VirtEngineApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *VirtEngineApp) RegisterTendermintService(cctx client.Context) {
	cmtservice.RegisterTendermintService(
		cctx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *VirtEngineApp) RegisterNodeService(cctx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(cctx, app.GRPCQueryRouter(), cfg)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// LoadHeight method of VirtEngineApp loads baseapp application version with given height
func (app *VirtEngineApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// cache the reflectionService to save us time within tests.
var cachedReflectionService *runtimeservices.ReflectionService

func getReflectionService() *runtimeservices.ReflectionService {
	if cachedReflectionService != nil {
		return cachedReflectionService
	}
	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	cachedReflectionService = reflectionSvc
	return reflectionSvc
}

// NewProposalContext returns a context with a branched version of the state
// that is safe to query during ProcessProposal.
func (app *VirtEngineApp) NewProposalContext(header tmproto.Header) sdk.Context {
	// use custom query multistore if provided
	ms := app.CommitMultiStore().CacheMultiStore()
	ctx := sdk.NewContext(ms, header, false, app.Logger()).
		WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).
		WithBlockHeader(header)
	ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))

	return ctx
}
