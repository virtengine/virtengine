package v1beta4

import (
	"fmt"
	"math"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	keyMinDeposits = "MinDeposits"
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte(keyMinDeposits), &p.MinDeposits, validateMinDeposits),
	}
}

func DefaultParams() Params {
	return Params{
		MinDeposits: sdk.Coins{
			sdk.NewCoin("uve", sdkmath.NewInt(500000)),
			sdk.NewCoin("uact", sdkmath.NewInt(500000)),
		},
	}
}

func (p Params) Validate() error {
	if err := validateMinDeposits(p.MinDeposits); err != nil {
		return err
	}
	return nil
}

func (p Params) ValidateDeposit(amt sdk.Coin) error {
	minDeposit, err := p.MinDepositFor(amt.Denom)

	if err != nil {
		return err
	}

	if amt.IsGTE(minDeposit) {
		return nil
	}

	return errors.Wrapf(v1.ErrInvalidDeposit, "Deposit too low - %v < %v", amt.Amount, minDeposit)
}

func (p Params) MinDepositFor(denom string) (sdk.Coin, error) {
	for _, minDeposit := range p.MinDeposits {
		if minDeposit.Denom == denom {
			return sdk.NewCoin(minDeposit.Denom, minDeposit.Amount), nil
		}
	}

	return sdk.NewInt64Coin(denom, math.MaxInt64), fmt.Errorf("%w: Invalid deposit denomination %v", v1.ErrInvalidDeposit, denom)
}

func validateMinDeposits(i interface{}) error {
	vals, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("%w: Min Deposits - invalid type: %T", v1.ErrInvalidParam, i)
	}

	check := make(map[string]bool)

	for _, minDeposit := range vals {
		if _, exists := check[minDeposit.Denom]; exists {
			return fmt.Errorf("duplicate Min Deposit for denom (%#v)", minDeposit)
		}

		check[minDeposit.Denom] = true

		if minDeposit.Amount.Uint64() >= math.MaxInt32 {
			return fmt.Errorf("%w: Min Deposit (%v) - too large: %v", v1.ErrInvalidParam, minDeposit.Denom, minDeposit.Amount.Uint64())
		}
	}

	if _, exists := check["uve"]; !exists {
		return fmt.Errorf("%w: Min Deposits - uve not given: %#v", v1.ErrInvalidParam, vals)
	}

	if _, exists := check["uact"]; !exists {
		return fmt.Errorf("%w: Min Deposits - uact not given: %#v", v1.ErrInvalidParam, vals)
	}

	return nil
}
