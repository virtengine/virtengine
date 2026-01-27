package v1

import (
	"reflect"

	cerrors "cosmossdk.io/errors"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgAddPriceEntry{}
	_ sdk.Msg = &MsgUpdateParams{}
)

var (
	msgTypeAddPriceEntry = ""
	msgTypeUpdateParams  = ""
)

func init() {
	msgTypeAddPriceEntry = reflect.TypeOf(&MsgAddPriceEntry{}).Elem().Name()
	msgTypeUpdateParams = reflect.TypeOf(&MsgUpdateParams{}).Elem().Name()
}

// ====MsgAddPriceEntry====

// Type implements the sdk.Msg interface
func (m *MsgAddPriceEntry) Type() string { return msgTypeAddPriceEntry }

// GetSigners returns the expected signers for a MsgAddPriceEntry message.
func (m *MsgAddPriceEntry) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Signer)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgAddPriceEntry) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return cerrors.Wrap(err, "invalid authority address")
	}

	return nil
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgAddPriceEntry) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgAddPriceEntry) Route() string {
	return RouterKey
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

	if err := m.Params.ValidateBasic(); err != nil {
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

// UnpackInterfaces implements UnpackInterfacesMessage
func (p MsgUpdateParams) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return p.Params.UnpackInterfaces(unpacker)
}
