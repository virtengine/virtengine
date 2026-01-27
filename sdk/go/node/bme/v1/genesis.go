package v1

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		State: GenesisVaultState{
			TotalBurned: sdk.Coins{
				sdk.NewCoin(sdkutil.DenomUact, sdkmath.ZeroInt()),
				sdk.NewCoin(sdkutil.Denomuve, sdkmath.ZeroInt()),
			},
			TotalMinted: sdk.Coins{
				sdk.NewCoin(sdkutil.DenomUact, sdkmath.ZeroInt()),
				sdk.NewCoin(sdkutil.Denomuve, sdkmath.ZeroInt()),
			},
			// do not uses sdk.NewCoins as it's sanitize removes zero coins
			RemintCredits: sdk.Coins{
				sdk.NewCoin(sdkutil.Denomuve, sdkmath.ZeroInt()),
			},
		},
	}
}

// Validate validates the genesis state.
func (gs *GenesisState) Validate() error {
	return gs.Params.Validate()
}
