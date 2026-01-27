package v1

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSeedVault{}
)

var (
	msgTypeUpdateParams = ""
)

func init() {
	msgTypeUpdateParams = reflect.TypeOf(&MsgUpdateParams{}).Elem().Name()
}

func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	return msg.Params.Validate()
}

func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// Type implements the sdk.Msg interface
func (msg *MsgUpdateParams) Type() string { return msgTypeUpdateParams }

func (msg MsgSeedVault) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized.Wrapf("invalid authority address: %s", err)
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap("amount must be positive")
	}
	if msg.Amount.Denom != "uakt" {
		return ErrInvalidDenom.Wrapf("expected uakt, got %s", msg.Amount.Denom)
	}
	return nil
}

func (msg MsgSeedVault) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}
