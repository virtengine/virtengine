package v1beta3

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// var _ paramtypes.ParamSet = (*DepositParams)(nil)

var (
	DefaultMinInitialDepositRate = sdkmath.LegacyNewDecWithPrec(40, 2)
)

var (
	KeyDepositParams = []byte("depositparams")
)

// ParamKeyTable for agov module
// Deprecated: now params can be accessed via cosmos-sdk gov store
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(KeyDepositParams, DepositParams{}, validateDepositParams),
	)
}

// NewDepositParams creates a new DepositParams object
func NewDepositParams(minInitialDepositRate sdkmath.LegacyDec) DepositParams {
	return DepositParams{
		MinInitialDepositRate: minInitialDepositRate,
	}
}

func DefaultDepositParams() DepositParams {
	return NewDepositParams(
		DefaultMinInitialDepositRate,
	)
}

func (p DepositParams) Validate() error {
	if err := validateMinInitialDepositRate(p.MinInitialDepositRate); err != nil {
		return err
	}

	return nil
}

func validateDepositParams(i interface{}) error {
	v, ok := i.(DepositParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := validateMinInitialDepositRate(v.MinInitialDepositRate); err != nil {
		return err
	}

	return nil
}

func validateMinInitialDepositRate(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min deposit on proposal create cannot be negative: %s", v)
	}

	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("min deposit on proposal create is too large: %s", v)
	}

	return nil
}
