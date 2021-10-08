package app

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	"github.com/virtengine/virtengine/x/audit"
	"github.com/virtengine/virtengine/x/cert"
	"github.com/virtengine/virtengine/x/deployment"
	"github.com/virtengine/virtengine/x/escrow"
	ekeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	"github.com/virtengine/virtengine/x/market"
	mhooks "github.com/virtengine/virtengine/x/market/hooks"
	"github.com/virtengine/virtengine/x/provider"
)

func virtengineModuleBasics() []module.AppModuleBasic {
	return []module.AppModuleBasic{
		escrow.AppModuleBasic{},
		deployment.AppModuleBasic{},
		market.AppModuleBasic{},
		provider.AppModuleBasic{},
		audit.AppModuleBasic{},
		cert.AppModuleBasic{},
	}
}

func virtengineKVStoreKeys() []string {
	return []string{
		escrow.StoreKey,
		deployment.StoreKey,
		market.StoreKey,
		provider.StoreKey,
		audit.StoreKey,
		cert.StoreKey,
	}
}

func virtengineSubspaces(k paramskeeper.Keeper) paramskeeper.Keeper {
	k.Subspace(deployment.ModuleName)
	k.Subspace(market.ModuleName)
	return k
}

func (app *VirtEngineApp) setVirtEngineKeepers() {

	app.keeper.escrow = ekeeper.NewKeeper(
		app.appCodec,
		app.keys[escrow.StoreKey],
		app.keeper.bank,
	)

	app.keeper.deployment = deployment.NewKeeper(
		app.appCodec,
		app.keys[deployment.StoreKey],
		app.GetSubspace(deployment.ModuleName),
		app.keeper.escrow,
	)

	app.keeper.market = market.NewKeeper(
		app.appCodec,
		app.keys[market.StoreKey],
		app.GetSubspace(market.ModuleName),
		app.keeper.escrow,
	)

	hook := mhooks.New(app.keeper.deployment, app.keeper.market)

	app.keeper.escrow.AddOnAccountClosedHook(hook.OnEscrowAccountClosed)
	app.keeper.escrow.AddOnPaymentClosedHook(hook.OnEscrowPaymentClosed)

	app.keeper.provider = provider.NewKeeper(
		app.appCodec,
		app.keys[provider.StoreKey],
	)

	app.keeper.audit = audit.NewKeeper(
		app.appCodec,
		app.keys[audit.StoreKey],
	)

	app.keeper.cert = cert.NewKeeper(
		app.appCodec,
		app.keys[cert.StoreKey],
	)
}

func (app *VirtEngineApp) virtengineAppModules() []module.AppModule {
	return []module.AppModule{

		escrow.NewAppModule(
			app.appCodec,
			app.keeper.escrow,
		),

		deployment.NewAppModule(
			app.appCodec,
			app.keeper.deployment,
			app.keeper.market,
			app.keeper.escrow,
			app.keeper.bank,
		),

		market.NewAppModule(
			app.appCodec,
			app.keeper.market,
			app.keeper.escrow,
			app.keeper.audit,
			app.keeper.deployment,
			app.keeper.provider,
			app.keeper.bank,
		),

		provider.NewAppModule(
			app.appCodec,
			app.keeper.provider,
			app.keeper.bank,
			app.keeper.market,
		),

		audit.NewAppModule(
			app.appCodec,
			app.keeper.audit,
		),

		cert.NewAppModule(
			app.appCodec,
			app.keeper.cert,
		),
	}
}

func (app *VirtEngineApp) virtengineEndBlockModules() []string {
	return []string{
		deployment.ModuleName, market.ModuleName,
	}
}

func (app *VirtEngineApp) virtengineInitGenesisOrder() []string {
	return []string{
		cert.ModuleName,
		escrow.ModuleName,
		deployment.ModuleName,
		provider.ModuleName,
		market.ModuleName,
	}
}

func (app *VirtEngineApp) virtengineSimModules() []module.AppModuleSimulation {
	return []module.AppModuleSimulation{
		deployment.NewAppModuleSimulation(
			app.keeper.deployment,
			app.keeper.acct,
			app.keeper.bank,
		),

		market.NewAppModuleSimulation(
			app.keeper.market,
			app.keeper.acct,
			app.keeper.deployment,
			app.keeper.provider,
			app.keeper.bank,
		),

		provider.NewAppModuleSimulation(
			app.keeper.provider,
			app.keeper.acct,
			app.keeper.bank,
		),

		cert.NewAppModuleSimulation(
			app.keeper.cert,
		),
	}
}
