package v1beta4

import (
	"fmt"
	"reflect"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	msgTypeGenerateDomainVerificationToken = ""
	msgTypeVerifyProviderDomain            = ""
)

var (
	_ sdk.Msg = &MsgGenerateDomainVerificationToken{}
	_ sdk.Msg = &MsgVerifyProviderDomain{}
)

func init() {
	msgTypeGenerateDomainVerificationToken = reflect.TypeOf(&MsgGenerateDomainVerificationToken{}).Elem().Name()
	msgTypeVerifyProviderDomain = reflect.TypeOf(&MsgVerifyProviderDomain{}).Elem().Name()
}

// MsgGenerateDomainVerificationToken defines message for generating domain verification token
type MsgGenerateDomainVerificationToken struct {
	Owner  string `json:"owner" yaml:"owner"`
	Domain string `json:"domain" yaml:"domain"`
}

// ProtoMessage implements proto.Message
func (*MsgGenerateDomainVerificationToken) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgGenerateDomainVerificationToken) Reset() { *msg = MsgGenerateDomainVerificationToken{} }

// String implements proto.Message
func (msg *MsgGenerateDomainVerificationToken) String() string {
	return fmt.Sprintf("MsgGenerateDomainVerificationToken{Owner: %s, Domain: %s}", msg.Owner, msg.Domain)
}

// NewMsgGenerateDomainVerificationToken creates a new MsgGenerateDomainVerificationToken instance
func NewMsgGenerateDomainVerificationToken(owner sdk.AccAddress, domain string) *MsgGenerateDomainVerificationToken {
	return &MsgGenerateDomainVerificationToken{
		Owner:  owner.String(),
		Domain: domain,
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgGenerateDomainVerificationToken) Type() string {
	return msgTypeGenerateDomainVerificationToken
}

// ValidateBasic does basic validation
func (msg *MsgGenerateDomainVerificationToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgGenerateDomainVerificationToken: Invalid Owner Address")
	}
	if msg.Domain == "" {
		return ErrInvalidDomain.Wrap("domain cannot be empty")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgGenerateDomainVerificationToken) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// MsgGenerateDomainVerificationTokenResponse defines the response
type MsgGenerateDomainVerificationTokenResponse struct {
	Token     string `json:"token" yaml:"token"`
	ExpiresAt int64  `json:"expires_at" yaml:"expires_at"`
}

// ProtoMessage implements proto.Message
func (*MsgGenerateDomainVerificationTokenResponse) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgGenerateDomainVerificationTokenResponse) Reset() {
	*msg = MsgGenerateDomainVerificationTokenResponse{}
}

// String implements proto.Message
func (msg *MsgGenerateDomainVerificationTokenResponse) String() string {
	return fmt.Sprintf("MsgGenerateDomainVerificationTokenResponse{Token: %s, ExpiresAt: %d}", msg.Token, msg.ExpiresAt)
}

// MsgVerifyProviderDomain defines message for verifying provider domain
type MsgVerifyProviderDomain struct {
	Owner string `json:"owner" yaml:"owner"`
}

// ProtoMessage implements proto.Message
func (*MsgVerifyProviderDomain) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgVerifyProviderDomain) Reset() { *msg = MsgVerifyProviderDomain{} }

// String implements proto.Message
func (msg *MsgVerifyProviderDomain) String() string {
	return fmt.Sprintf("MsgVerifyProviderDomain{Owner: %s}", msg.Owner)
}

// NewMsgVerifyProviderDomain creates a new MsgVerifyProviderDomain instance
func NewMsgVerifyProviderDomain(owner sdk.AccAddress) *MsgVerifyProviderDomain {
	return &MsgVerifyProviderDomain{
		Owner: owner.String(),
	}
}

// Type implements the sdk.Msg interface
func (msg *MsgVerifyProviderDomain) Type() string {
	return msgTypeVerifyProviderDomain
}

// ValidateBasic does basic validation
func (msg *MsgVerifyProviderDomain) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgVerifyProviderDomain: Invalid Owner Address")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgVerifyProviderDomain) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// MsgVerifyProviderDomainResponse defines the response
type MsgVerifyProviderDomainResponse struct {
	Verified bool `json:"verified" yaml:"verified"`
}

// ProtoMessage implements proto.Message
func (*MsgVerifyProviderDomainResponse) ProtoMessage() {}

// Reset implements proto.Message
func (msg *MsgVerifyProviderDomainResponse) Reset() { *msg = MsgVerifyProviderDomainResponse{} }

// String implements proto.Message
func (msg *MsgVerifyProviderDomainResponse) String() string {
	return fmt.Sprintf("MsgVerifyProviderDomainResponse{Verified: %t}", msg.Verified)
}

// ============= GetSignBytes =============

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgGenerateDomainVerificationToken) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSignBytes encodes the message for signing
//
// Deprecated: GetSignBytes is deprecated
func (msg *MsgVerifyProviderDomain) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// ============= Route =============

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgGenerateDomainVerificationToken) Route() string { return RouterKey }

// Route implements the sdk.Msg interface
//
// Deprecated: Route is deprecated
func (msg *MsgVerifyProviderDomain) Route() string { return RouterKey }
