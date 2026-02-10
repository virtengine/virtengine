package types

import "fmt"

// GenesisState defines the module genesis state.
type GenesisState struct {
	Params      Params
	Inventories []ResourceInventory
	Allocations []ResourceAllocation
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		Inventories: []ResourceInventory{},
		Allocations: []ResourceAllocation{},
	}
}

// Validate validates genesis state.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}
	return nil
}
