package v1beta4

import (
	"fmt"

	cerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

// VerificationMethod defines the enum for provider domain verification methods.
type VerificationMethod = int32

const (
	VERIFICATION_METHOD_UNKNOWN         VerificationMethod = 0
	VERIFICATION_METHOD_DNS_TXT         VerificationMethod = 1
	VERIFICATION_METHOD_DNS_CNAME       VerificationMethod = 2
	VERIFICATION_METHOD_HTTP_WELL_KNOWN VerificationMethod = 3
)

// MsgRequestDomainVerification requests a domain verification token.
type MsgRequestDomainVerification struct {
	Owner  string             `json:"owner"`
	Domain string             `json:"domain"`
	Method VerificationMethod `json:"method"`
}

// MsgRequestDomainVerificationResponse returns a verification token.
type MsgRequestDomainVerificationResponse struct {
	Token              string `json:"token"`
	ExpiresAt          int64  `json:"expires_at"`
	VerificationTarget string `json:"verification_target"`
}

// MsgConfirmDomainVerification confirms a domain verification.
type MsgConfirmDomainVerification struct {
	Owner string `json:"owner"`
	Proof string `json:"proof"`
}

// MsgConfirmDomainVerificationResponse returns confirmation status.
type MsgConfirmDomainVerificationResponse struct {
	Verified   bool  `json:"verified"`
	VerifiedAt int64 `json:"verified_at"`
}

// MsgRevokeDomainVerification revokes a domain verification.
type MsgRevokeDomainVerification struct {
	Owner string `json:"owner"`
}

func (msg *MsgRequestDomainVerification) Reset()         { *msg = MsgRequestDomainVerification{} }
func (msg *MsgRequestDomainVerification) String() string { return "MsgRequestDomainVerification" }
func (*MsgRequestDomainVerification) ProtoMessage()      {}

func (msg *MsgConfirmDomainVerification) Reset()         { *msg = MsgConfirmDomainVerification{} }
func (msg *MsgConfirmDomainVerification) String() string { return "MsgConfirmDomainVerification" }
func (*MsgConfirmDomainVerification) ProtoMessage()      {}

func (msg *MsgRevokeDomainVerification) Reset()         { *msg = MsgRevokeDomainVerification{} }
func (msg *MsgRevokeDomainVerification) String() string { return "MsgRevokeDomainVerification" }
func (*MsgRevokeDomainVerification) ProtoMessage()      {}

// MsgRevokeDomainVerificationResponse is the revoke response.
type MsgRevokeDomainVerificationResponse struct{}

// ValidateBasic validates MsgRequestDomainVerification.
func (msg *MsgRequestDomainVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgRequestDomainVerification: Invalid Provider Address")
	}
	if err := validateDomain(msg.Domain); err != nil {
		return ErrInvalidDomain.Wrapf("invalid domain: %v", err)
	}
	if msg.Method == VERIFICATION_METHOD_UNKNOWN {
		return ErrInvalidDomain.Wrap("verification method cannot be unknown")
	}
	return nil
}

// Type implements the sdk.Msg interface.
func (msg *MsgRequestDomainVerification) Type() string { return msgTypeRequestDomainVerification }

// GetSigners defines whose signature is required.
func (msg *MsgRequestDomainVerification) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// GetSignBytes encodes the message for signing.
//
// Deprecated: GetSignBytes is deprecated.
func (msg *MsgRequestDomainVerification) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Route implements the sdk.Msg interface.
//
// Deprecated: Route is deprecated.
func (msg *MsgRequestDomainVerification) Route() string { return RouterKey }

// ValidateBasic validates MsgConfirmDomainVerification.
func (msg *MsgConfirmDomainVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgConfirmDomainVerification: Invalid Provider Address")
	}
	if msg.Proof == "" {
		return ErrDomainVerificationFailed.Wrap("proof cannot be empty")
	}
	return nil
}

// Type implements the sdk.Msg interface.
func (msg *MsgConfirmDomainVerification) Type() string { return msgTypeConfirmDomainVerification }

// GetSigners defines whose signature is required.
func (msg *MsgConfirmDomainVerification) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// GetSignBytes encodes the message for signing.
//
// Deprecated: GetSignBytes is deprecated.
func (msg *MsgConfirmDomainVerification) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Route implements the sdk.Msg interface.
//
// Deprecated: Route is deprecated.
func (msg *MsgConfirmDomainVerification) Route() string { return RouterKey }

// ValidateBasic validates MsgRevokeDomainVerification.
func (msg *MsgRevokeDomainVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return cerrors.Wrap(sdkerrors.ErrInvalidAddress, "MsgRevokeDomainVerification: Invalid Provider Address")
	}
	return nil
}

// Type implements the sdk.Msg interface.
func (msg *MsgRevokeDomainVerification) Type() string { return msgTypeRevokeDomainVerification }

// GetSigners defines whose signature is required.
func (msg *MsgRevokeDomainVerification) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// GetSignBytes encodes the message for signing.
//
// Deprecated: GetSignBytes is deprecated.
func (msg *MsgRevokeDomainVerification) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Route implements the sdk.Msg interface.
//
// Deprecated: Route is deprecated.
func (msg *MsgRevokeDomainVerification) Route() string { return RouterKey }

// EventProviderDomainVerificationRequested is emitted when verification is requested.
type EventProviderDomainVerificationRequested struct {
	Owner  string `json:"owner"`
	Domain string `json:"domain"`
	Method string `json:"method"`
	Token  string `json:"token"`
}

// EventProviderDomainVerificationConfirmed is emitted when verification is confirmed.
type EventProviderDomainVerificationConfirmed struct {
	Owner  string `json:"owner"`
	Domain string `json:"domain"`
	Method string `json:"method"`
}

// EventProviderDomainVerificationRevoked is emitted when verification is revoked.
type EventProviderDomainVerificationRevoked struct {
	Owner  string `json:"owner"`
	Domain string `json:"domain"`
}

func (*EventProviderDomainVerificationRequested) Reset() {}
func (*EventProviderDomainVerificationRequested) String() string {
	return "EventProviderDomainVerificationRequested"
}
func (*EventProviderDomainVerificationRequested) ProtoMessage() {}
func (*EventProviderDomainVerificationConfirmed) Reset()        {}
func (*EventProviderDomainVerificationConfirmed) String() string {
	return "EventProviderDomainVerificationConfirmed"
}
func (*EventProviderDomainVerificationConfirmed) ProtoMessage() {}
func (*EventProviderDomainVerificationRevoked) Reset()          {}
func (*EventProviderDomainVerificationRevoked) String() string {
	return "EventProviderDomainVerificationRevoked"
}
func (*EventProviderDomainVerificationRevoked) ProtoMessage() {}

var (
	_ sdk.Msg       = (*MsgRequestDomainVerification)(nil)
	_ sdk.Msg       = (*MsgConfirmDomainVerification)(nil)
	_ sdk.Msg       = (*MsgRevokeDomainVerification)(nil)
	_ proto.Message = (*EventProviderDomainVerificationRequested)(nil)
	_ proto.Message = (*EventProviderDomainVerificationConfirmed)(nil)
	_ proto.Message = (*EventProviderDomainVerificationRevoked)(nil)
)

func init() {
	if msgTypeRequestDomainVerification == "" {
		msgTypeRequestDomainVerification = "request_domain_verification"
	}
	if msgTypeConfirmDomainVerification == "" {
		msgTypeConfirmDomainVerification = "confirm_domain_verification"
	}
	if msgTypeRevokeDomainVerification == "" {
		msgTypeRevokeDomainVerification = "revoke_domain_verification"
	}
}

// String implements fmt.Stringer for responses to avoid nil pointer logging.
func (msg *MsgRequestDomainVerificationResponse) String() string {
	return fmt.Sprintf("MsgRequestDomainVerificationResponse{%s}", msg.Token)
}
