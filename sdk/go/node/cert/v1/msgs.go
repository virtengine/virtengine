package v1

import (
	"math/big"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreateCertificate{}
	_ sdk.Msg = &MsgRevokeCertificate{}
)

var (
	MsgTypeCreateCertificate = ""
	MsgTypeRevokeCertificate = ""
)

func init() {
	MsgTypeCreateCertificate = reflect.TypeOf(&MsgCreateCertificate{}).Elem().Name()
	MsgTypeRevokeCertificate = reflect.TypeOf(&MsgRevokeCertificate{}).Elem().Name()
}

// ====MsgCreateCertificate====

// Type implements the sdk.Msg interface
func (m *MsgCreateCertificate) Type() string {
	return MsgTypeCreateCertificate
}

// ValidateBasic does basic validation
func (m *MsgCreateCertificate) ValidateBasic() error {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgCreateCertificate: Invalid Owner Address")
	}

	_, err = ParseAndValidateCertificate(owner, m.Cert, m.Pubkey)
	if err != nil {
		return err
	}

	return nil
}

// GetSigners defines whose signature is required
func (m *MsgCreateCertificate) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// ====MsgRevokeCertificate====

// Type implements the sdk.Msg interface
func (m *MsgRevokeCertificate) Type() string {
	return MsgTypeRevokeCertificate
}

// ValidateBasic does basic validation
func (m *MsgRevokeCertificate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.ID.Owner); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("MsgRevoke: Invalid Owner Address")
	}

	if _, valid := new(big.Int).SetString(m.ID.Serial, 10); !valid {
		return ErrInvalidSerialNumber
	}

	return nil
}

// GetSigners defines whose signature is required
func (m *MsgRevokeCertificate) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.ID.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgCreateCertificate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (m *MsgRevokeCertificate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgCreateCertificate) Route() string {
	return RouterKey
}

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (m *MsgRevokeCertificate) Route() string {
	return RouterKey
}
