package v1

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable for wasm module
// Deprecated: now params can be accessed on key `0x01` on the wasm store.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

func DefaultParams() Params {
	return Params{}
}

func (p Params) Validate() error {
	for _, addr := range p.BlockedAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.Wrapf(err, "invalid blocked address: %s", addr)
		}
	}

	return nil
}
