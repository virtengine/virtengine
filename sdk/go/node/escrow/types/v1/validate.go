package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/node/escrow/module"
)

func (obj *PaymentState) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(obj.Owner)
	if err != nil {
		return err
	}

	if !obj.State.IsValid() {
		return module.ErrInvalidPayment.Wrapf("invalid payment state %d", obj.State)
	}

	if obj.Rate.IsZero() {
		return module.ErrInvalidPayment.Wrap("payment rate zero")
	}
	if obj.State == StateInvalid {
		return module.ErrInvalidPayment.Wrap("invalid state")
	}

	return nil
}

