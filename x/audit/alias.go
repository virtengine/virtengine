package audit

import (
	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"

	"github.com/virtengine/virtengine/x/audit/keeper"
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
