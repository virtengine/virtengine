package provider

import (
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	"github.com/virtengine/virtengine/x/provider/keeper"
)

const (
	// StoreKey represents storekey of provider module
	StoreKey = types.StoreKey
	// ModuleName represents current module name
	ModuleName = types.ModuleName
)

type (
	// Keeper defines keeper of provider module
	Keeper = keeper.Keeper
)

var (
	// NewKeeper creates new keeper instance of provider module
	NewKeeper = keeper.NewKeeper
)
