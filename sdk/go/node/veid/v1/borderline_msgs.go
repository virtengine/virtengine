package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for borderline messages
const (
	TypeMsgCompleteBorderlineFallback = "complete_borderline_fallback"
	TypeMsgUpdateBorderlineParams     = "update_borderline_params"
)

var (
	_ sdk.Msg = &MsgCompleteBorderlineFallback{}
	_ sdk.Msg = &MsgUpdateBorderlineParams{}
)

// ============================================================================
// MsgCompleteBorderlineFallback
// ============================================================================

// Route returns the route for the message
func (msg *MsgCompleteBorderlineFallback) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgCompleteBorderlineFallback) Type() string { return TypeMsgCompleteBorderlineFallback }

// ValidateBasic validates the message
func (msg *MsgCompleteBorderlineFallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ChallengeId == "" {
		return ErrBorderlineFallbackFailed.Wrap("challenge_id cannot be empty")
	}

	if len(msg.FactorsSatisfied) == 0 {
		return ErrBorderlineFallbackFailed.Wrap("at least one satisfied factor is required")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgCompleteBorderlineFallback) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgUpdateBorderlineParams
// ============================================================================

// Route returns the route for the message
func (msg *MsgUpdateBorderlineParams) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateBorderlineParams) Type() string { return TypeMsgUpdateBorderlineParams }

// ValidateBasic validates the message
func (msg *MsgUpdateBorderlineParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.Params == nil {
		return ErrBorderlineFallbackFailed.Wrap("params cannot be nil")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateBorderlineParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}

// ============================================================================
// MsgUpdateParams (module params)
// ============================================================================

// Route returns the route for the message
func (msg *MsgUpdateParams) Route() string { return RouterKey }

// Type returns the type for the message
func (msg *MsgUpdateParams) Type() string { return "update_params" }

// ValidateBasic validates the message
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address")
	}

	if msg.Params == nil {
		return ErrInvalidScope.Wrap("params cannot be nil")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{signer}
}
