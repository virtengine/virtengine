package v1beta4

import (
	"net/url"
	"reflect"
	"regexp"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	attr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

var (
	msgTypeCreateProvider = ""
	msgTypeUpdateProvider = ""
	msgTypeDeleteProvider = ""
)

var (
	_ sdk.Msg = &MsgCreateProvider{}
	_ sdk.Msg = &MsgUpdateProvider{}
	_ sdk.Msg = &MsgDeleteProvider{}
)

var (
	attributeNameRegexp = regexp.MustCompile(attr.AttributeNameRegexpString)
)

func init() {
	msgTypeCreateProvider = reflect.TypeOf(&MsgCreateProvider{}).Elem().Name()
	msgTypeUpdateProvider = reflect.TypeOf(&MsgUpdateProvider{}).Elem().Name()
	msgTypeDeleteProvider = reflect.TypeOf(&MsgDeleteProvider{}).Elem().Name()
}

// NewMsgCreateProvider creates a new MsgCreateProvider instance
func NewMsgCreateProvider(owner sdk.AccAddress, hostURI string, attributes attr.Attributes) *MsgCreateProvider {
	return &MsgCreateProvider{
		Owner:      owner.String(),
		HostURI:    hostURI,
		Attributes: attributes,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgCreateProvider) Type() string { return msgTypeCreateProvider }

// ValidateBasic does basic validation of a HostURI
func (msg *MsgCreateProvider) ValidateBasic() error {
	if err := validateProviderURI(msg.HostURI); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgCreate: Invalid Provider Address")
	}
	if err := msg.Attributes.ValidateWithRegex(attributeNameRegexp); err != nil {
		return err
	}
	if err := msg.Info.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgCreateProvider) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgUpdateProvider creates a new MsgUpdateProvider instance
func NewMsgUpdateProvider(owner sdk.AccAddress, hostURI string, attributes attr.Attributes) *MsgUpdateProvider {
	return &MsgUpdateProvider{
		Owner:      owner.String(),
		HostURI:    hostURI,
		Attributes: attributes,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgUpdateProvider) Type() string { return msgTypeUpdateProvider }

// ValidateBasic does basic validation of a ProviderURI
func (msg *MsgUpdateProvider) ValidateBasic() error {
	if err := validateProviderURI(msg.HostURI); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgUpdate: Invalid Provider Address")
	}
	if err := msg.Attributes.ValidateWithRegex(attributeNameRegexp); err != nil {
		return err
	}
	if err := msg.Info.Validate(); err != nil {
		return err
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgUpdateProvider) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

// NewMsgDeleteProvider creates a new MsgDeleteProvider instance
func NewMsgDeleteProvider(owner sdk.AccAddress) *MsgDeleteProvider {
	return &MsgDeleteProvider{
		Owner: owner.String(),
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgDeleteProvider) Type() string { return msgTypeDeleteProvider }

// ValidateBasic does basic validation
func (msg *MsgDeleteProvider) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgDelete: Invalid Provider Address")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgDeleteProvider) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{owner}
}

func validateProviderURI(val string) error {
	u, err := url.Parse(val)
	if err != nil {
		return ErrInvalidProviderURI
	}
	if !u.IsAbs() {
		return ErrNotAbsProviderURI.Wrapf("validating %q for absolute URI", val)
	}

	if u.Scheme != "https" {
		return ErrInvalidProviderURI.Wrapf("scheme in %q should be https", val)
	}

	if u.Host == "" {
		return ErrInvalidProviderURI.Wrapf("validating %q for valid host", val)
	}

	if u.Path != "" {
		return ErrInvalidProviderURI.Wrapf("path in %q should be empty", val)
	}

	return nil
}

// ============= GetSignBytes =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgCreateProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgUpdateProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgDeleteProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// ============= Route =============
// ModuleCdc is defined in codec.go
// TODO @troian to check if we need them at all since sdk.Msg does not not have Route defined anymore

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgCreateProvider) Route() string { return RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgUpdateProvider) Route() string { return RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgDeleteProvider) Route() string { return RouterKey }

