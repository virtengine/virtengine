package deployment

import (
	"github.com/virtengine/virtengine/x/deployment/keeper"
	"github.com/virtengine/virtengine/x/deployment/types"
)

const (
	// StoreKey represents storekey of deployment module
	StoreKey = types.StoreKey
	// ModuleName represents current module name
	ModuleName = types.ModuleName
)

type (
	// Keeper defines keeper of deployment module
	Keeper = keeper.Keeper
)

var (
	// NewKeeper creates new keeper instance of deployment module
	NewKeeper = keeper.NewKeeper
)
