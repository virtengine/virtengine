package v1

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgSignProviderAttributes{}
	_ sdk.Msg = &MsgDeleteProviderAttributes{}
)

var (
	MsgTypeSignProviderAttributes   = ""
	MsgTypeDeleteProviderAttributes = ""
)

func init() {
	MsgTypeSignProviderAttributes = reflect.TypeOf(&MsgSignProviderAttributes{}).Elem().Name()
	MsgTypeDeleteProviderAttributes = reflect.TypeOf(&MsgDeleteProviderAttributes{}).Elem().Name()
}

// ====MsgSignProviderAttributes====

// Type implements the sdk.Msg interface
func (m *MsgSignProviderAttributes) Type() string {
	return MsgTypeSignProviderAttributes
}

// ValidateBasic does basic validation
func (m *MsgSignProviderAttributes) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgCreate: Invalid Owner Address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Auditor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgCreate: Invalid Auditor Address")
	}

	return nil
}

// GetSigners defines whose signature is required
func (m *MsgSignProviderAttributes) GetSigners() []sdk.AccAddress {
	auditor, err := sdk.AccAddressFromBech32(m.Auditor)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{auditor}
}

// ====MsgRevokeProviderAttributes====

// Type implements the sdk.Msg interface
func (m *MsgDeleteProviderAttributes) Type() string {
	return MsgTypeDeleteProviderAttributes
}

// ValidateBasic does basic validation
func (m *MsgDeleteProviderAttributes) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgCreate: Invalid Owner Address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Auditor); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgCreate: Invalid Auditor Address")
	}

	return nil
}

// GetSignBytes encodes the message for signing

// GetSigners defines whose signature is required
func (m *MsgDeleteProviderAttributes) GetSigners() []sdk.AccAddress {
	auditor, err := sdk.AccAddressFromBech32(m.Auditor)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{auditor}
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgSignProviderAttributes) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgDeleteProviderAttributes) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgSignProviderAttributes) Route() string {
	return RouterKey
}

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgDeleteProviderAttributes) Route() string {
	return RouterKey
}
