package v1

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	params := DefaultParams()

	return &GenesisState{
		Params: params,
		Prices: make([]PriceData, 0),
	}
}

// Validate validates the genesis state.
func (gs *GenesisState) Validate() error {
	return gs.Params.ValidateBasic()
}
