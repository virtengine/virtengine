package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for appeal messages
const (
	TypeMsgSubmitAppeal   = "submit_appeal"
	TypeMsgClaimAppeal    = "claim_appeal"
	TypeMsgResolveAppeal  = "resolve_appeal"
	TypeMsgWithdrawAppeal = "withdraw_appeal"
)

var (
	_ sdk.Msg = &MsgSubmitAppeal{}
	_ sdk.Msg = &MsgClaimAppeal{}
	_ sdk.Msg = &MsgResolveAppeal{}
	_ sdk.Msg = &MsgWithdrawAppeal{}
)

// ============================================================================
// MsgSubmitAppeal
// ============================================================================

// Route returns the route for the message
func (msg *MsgSubmitAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgSubmitAppeal) Type() string { return TypeMsgSubmitAppeal }

// ValidateBasic validates the message
func (msg *MsgSubmitAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrInvalidAddress.Wrap("invalid submitter address")
	}

	if msg.ScopeId == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidAppeal.Wrap("appeal reason cannot be empty")
	}

	if len(msg.Reason) > 2000 {
		return ErrInvalidAppeal.Wrap("appeal reason exceeds maximum length (2000 characters)")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgSubmitAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgClaimAppeal
// ============================================================================

// Route returns the route for the message
func (msg *MsgClaimAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgClaimAppeal) Type() string { return TypeMsgClaimAppeal }

// ValidateBasic validates the message
func (msg *MsgClaimAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Reviewer); err != nil {
		return ErrInvalidAddress.Wrap("invalid reviewer address")
	}

	if msg.AppealId == "" {
		return ErrInvalidAppeal.Wrap("appeal_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgClaimAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Reviewer)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgResolveAppeal
// ============================================================================

// Route returns the route for the message
func (msg *MsgResolveAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgResolveAppeal) Type() string { return TypeMsgResolveAppeal }

// ValidateBasic validates the message
func (msg *MsgResolveAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Resolver); err != nil {
		return ErrInvalidAddress.Wrap("invalid resolver address")
	}

	if msg.AppealId == "" {
		return ErrInvalidAppeal.Wrap("appeal_id cannot be empty")
	}

	// Resolution must be Approved, Rejected, or Withdrawn
	if msg.Resolution != AppealStatusApproved &&
		msg.Resolution != AppealStatusRejected &&
		msg.Resolution != AppealStatusWithdrawn {
		return ErrInvalidAppeal.Wrap("resolution must be approved, rejected, or withdrawn")
	}

	if msg.Reason == "" {
		return ErrInvalidAppeal.Wrap("resolution reason cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgResolveAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Resolver)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgWithdrawAppeal
// ============================================================================

// Route returns the route for the message
func (msg *MsgWithdrawAppeal) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgWithdrawAppeal) Type() string { return TypeMsgWithdrawAppeal }

// ValidateBasic validates the message
func (msg *MsgWithdrawAppeal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrInvalidAddress.Wrap("invalid submitter address")
	}

	if msg.AppealId == "" {
		return ErrInvalidAppeal.Wrap("appeal_id cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgWithdrawAppeal) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{signer}
}
