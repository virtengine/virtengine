package market

import (
	v1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"

	"github.com/virtengine/virtengine/x/market/keeper"
)

const (
	// StoreKey represents storekey of market module
	StoreKey = v1.StoreKey
	// ModuleName represents current module name
	ModuleName = v1.ModuleName
)

type (
	// Keeper defines keeper of market module
	Keeper = keeper.Keeper
)

var (
	// NewKeeper creates new keeper instance of market module
	NewKeeper = keeper.NewKeeper
)
