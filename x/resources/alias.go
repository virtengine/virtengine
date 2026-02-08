package resources

import (
	"github.com/virtengine/virtengine/x/resources/keeper"
	"github.com/virtengine/virtengine/x/resources/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

type (
	Keeper       = keeper.Keeper
	Params       = types.Params
	GenesisState = types.GenesisState

	ResourceInventory  = types.ResourceInventory
	ResourceRequest    = types.ResourceRequest
	ResourceAllocation = types.ResourceAllocation
)

var (
	NewKeeper                = keeper.NewKeeper
	DefaultParams            = types.DefaultParams
	DefaultGenesisState      = types.DefaultGenesisState
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
	RegisterInterfaces       = types.RegisterInterfaces
)
