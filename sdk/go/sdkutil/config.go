package sdkutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// init atm SDK configs is instantiated with const values, so it sealed within init below
// it helps for all tests as well as packages relying on this api to always have the same config
// as soon as sdkutil is imported
func init() {
	sdk.DefaultBondDenom = BondDenom

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
	config.Seal()
}
