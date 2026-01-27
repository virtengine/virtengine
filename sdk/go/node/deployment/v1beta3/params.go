package v1beta3

import (
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	sdk "github.com/virtengine/virtengine/sdk/go/node/types/sdk"
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
			sdk.NewCoin("uve", sdk.NewInt(500000)),
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

	return errorsmod.Wrapf(ErrInvalidDeposit, "Deposit too low - %v < %v", amt.Amount, minDeposit)
}

func (p Params) MinDepositFor(denom string) (sdk.Coin, error) {
	for _, minDeposit := range p.MinDeposits {
		if minDeposit.Denom == denom {
			return sdk.NewCoin(minDeposit.Denom, minDeposit.Amount), nil
		}
	}

	return sdk.NewInt64Coin(denom, math.MaxInt64), errorsmod.Wrapf(ErrInvalidDeposit, "Invalid deposit denomination %v", denom)
}

func validateMinDeposits(i interface{}) error {
	vals, ok := i.(sdk.Coins)
	if !ok {
		return errorsmod.Wrapf(ErrInvalidParam, "Min Deposits - invalid type: %T", i)
	}

	check := make(map[string]bool)

	for _, minDeposit := range vals {
		if _, exists := check[minDeposit.Denom]; exists {
			return fmt.Errorf("duplicate Min Deposit for denom (%#v)", minDeposit)
		}

		check[minDeposit.Denom] = true

		if minDeposit.Amount.Uint64() >= math.MaxInt32 {
			return errorsmod.Wrapf(ErrInvalidParam, "Min Deposit (%v) - too large: %v", minDeposit.Denom, minDeposit.Amount.Uint64())
		}
	}

	if _, exists := check["uve"]; !exists {
		return errorsmod.Wrapf(ErrInvalidParam, "Min Deposits - uve not given: %#v", vals)
	}

	return nil
}
