package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v10/modules/core/05-port/keeper"
)

// IBCPortKeeperAdapter adapts IBC-Go v10 PortKeeper to settlement module's expected interface.
// In IBC-Go v10, ports are registered during app initialization and don't require runtime binding.
type IBCPortKeeperAdapter struct {
	keeper *keeper.Keeper
}

// NewIBCPortKeeperAdapter creates a new adapter.
func NewIBCPortKeeperAdapter(k *keeper.Keeper) *IBCPortKeeperAdapter {
	return &IBCPortKeeperAdapter{keeper: k}
}

// BindPort is a no-op in IBC-Go v10 - ports are bound during app initialization.
func (a *IBCPortKeeperAdapter) BindPort(_ sdk.Context, _ string) {
	// Ports are registered during router setup in app initialization
	// No runtime binding needed in IBC-Go v10
}

// IsBound always returns true since ports are statically registered.
// In IBC-Go v10, if a port isn't registered, the module won't be routed to.
func (a *IBCPortKeeperAdapter) IsBound(_ sdk.Context, _ string) bool {
	// Ports are statically registered in the router during app init
	return true
}
