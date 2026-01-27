package v1beta3

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	DefaultMinCommissionRate = sdkmath.LegacyNewDecWithPrec(5, 2)
)

var (
	KeyMinCommissionRate = []byte("MinCommissionRate")
)

func NewParams(minCommissionRate sdkmath.LegacyDec) Params {
	return Params{
		MinCommissionRate: minCommissionRate,
	}
}

func DefaultParams() Params {
	return NewParams(DefaultMinCommissionRate)
}

// MustUnmarshalParams the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.LegacyAmino, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// UnmarshalParams the current staking params value from store key
func UnmarshalParams(cdc *codec.LegacyAmino, value []byte) (params Params, err error) {
	err = cdc.Unmarshal(value, &params)
	if err != nil {
		return
	}

	return
}

func (p Params) Validate() error {
	if err := validateMinCommissionRate(p.MinCommissionRate); err != nil {
		return err
	}

	return nil
}

// ParamKeyTable for astaking module
// Deprecated: now params can be accessed via cosmos-sdk staking store
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinCommissionRate, &p.MinCommissionRate, validateMinCommissionRate),
	}
}

func validateMinCommissionRate(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min commission rate cannot be negative: %s", v)
	}
	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("min commission rate is too large: %s", v)
	}

	return nil
}
