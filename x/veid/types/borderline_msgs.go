package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TypeMsgCompleteBorderlineFallback is the type for MsgCompleteBorderlineFallback
	TypeMsgCompleteBorderlineFallback = "complete_borderline_fallback"
)

// MsgCompleteBorderlineFallback is the message for completing a borderline fallback
// after the MFA challenge has been satisfied.
type MsgCompleteBorderlineFallback struct {
	// Sender is the account that is completing the fallback
	Sender string `json:"sender"`

	// ChallengeID is the MFA challenge ID that was satisfied
	ChallengeID string `json:"challenge_id"`

	// FactorsSatisfied are the factor types that were successfully verified
	// This is provided by the client and verified against the challenge record
	FactorsSatisfied []string `json:"factors_satisfied"`
}

// NewMsgCompleteBorderlineFallback creates a new MsgCompleteBorderlineFallback
func NewMsgCompleteBorderlineFallback(
	sender string,
	challengeID string,
	factorsSatisfied []string,
) *MsgCompleteBorderlineFallback {
	return &MsgCompleteBorderlineFallback{
		Sender:           sender,
		ChallengeID:      challengeID,
		FactorsSatisfied: factorsSatisfied,
	}
}

// Route implements sdk.Msg
func (msg MsgCompleteBorderlineFallback) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgCompleteBorderlineFallback) Type() string {
	return TypeMsgCompleteBorderlineFallback
}

// GetSignBytes implements sdk.Msg
func (msg MsgCompleteBorderlineFallback) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgCompleteBorderlineFallback) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements sdk.Msg
func (msg MsgCompleteBorderlineFallback) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidAddress.Wrap("sender cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	if msg.ChallengeID == "" {
		return ErrInvalidBorderlineFallback.Wrap("challenge_id cannot be empty")
	}

	if len(msg.FactorsSatisfied) == 0 {
		return ErrInvalidBorderlineFallback.Wrap("factors_satisfied cannot be empty")
	}

	return nil
}

// MsgCompleteBorderlineFallbackResponse is the response for MsgCompleteBorderlineFallback
type MsgCompleteBorderlineFallbackResponse struct {
	// FallbackID is the ID of the completed fallback
	FallbackID string `json:"fallback_id"`

	// FinalStatus is the resulting verification status
	FinalStatus VerificationStatus `json:"final_status"`

	// FactorClass is the security class of the satisfied factors
	FactorClass string `json:"factor_class"`
}

// ============================================================================
// Update Borderline Params Message
// ============================================================================

const (
	// TypeMsgUpdateBorderlineParams is the type for MsgUpdateBorderlineParams
	TypeMsgUpdateBorderlineParams = "update_borderline_params"
)

// MsgUpdateBorderlineParams is the message for updating borderline parameters
type MsgUpdateBorderlineParams struct {
	// Authority is the governance module account address
	Authority string `json:"authority"`

	// Params are the new borderline parameters
	Params BorderlineParams `json:"params"`
}

// NewMsgUpdateBorderlineParams creates a new MsgUpdateBorderlineParams
func NewMsgUpdateBorderlineParams(authority string, params BorderlineParams) *MsgUpdateBorderlineParams {
	return &MsgUpdateBorderlineParams{
		Authority: authority,
		Params:    params,
	}
}

// Route implements sdk.Msg
func (msg MsgUpdateBorderlineParams) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateBorderlineParams) Type() string {
	return TypeMsgUpdateBorderlineParams
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpdateBorderlineParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateBorderlineParams) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateBorderlineParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	return msg.Params.Validate()
}

// MsgUpdateBorderlineParamsResponse is the response for MsgUpdateBorderlineParams
type MsgUpdateBorderlineParamsResponse struct{}
