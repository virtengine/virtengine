package v1beta3

import (
	"reflect"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
)

var (
	msgTypeUpdateParams = ""
)

func init() {
	msgTypeUpdateParams = reflect.TypeOf(&MsgUpdateParams{}).Elem().Name()
}

// ====MsgUpdateParams====

// Type implements the sdk.Msg interface
func (m *MsgUpdateParams) Type() string { return msgTypeUpdateParams }

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cerrors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgUpdateParams) Route() string {
	return RouterKey
}
