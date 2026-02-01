// Package v1 provides SDK interface methods for generated staking message types.
package v1

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MaxPerformanceScore is the maximum performance score
const MaxPerformanceScore = 100

// ValidateBasic performs basic validation for MsgUpdateParams
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}
	return m.Params.Validate()
}

// GetSigners returns the expected signers for MsgUpdateParams
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation for MsgSlashValidator
func (m *MsgSlashValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}

	if !IsValidSlashReason(m.Reason) {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid slash reason: %s", m.Reason)
	}

	if m.InfractionHeight < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("infraction_height cannot be negative")
	}

	return nil
}

// GetSigners returns the expected signers for MsgSlashValidator
func (m *MsgSlashValidator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation for MsgUnjailValidator
func (m *MsgUnjailValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}
	return nil
}

// GetSigners returns the expected signers for MsgUnjailValidator
func (m *MsgUnjailValidator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.ValidatorAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation for MsgRecordPerformance
func (m *MsgRecordPerformance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %v", err)
	}

	if m.BlocksProposed < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("blocks_proposed cannot be negative")
	}

	if m.BlocksSigned < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("blocks_signed cannot be negative")
	}

	if m.VEIDVerificationsCompleted < 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("veid_verifications_completed cannot be negative")
	}

	if m.VEIDVerificationScore < 0 || m.VEIDVerificationScore > MaxPerformanceScore {
		return sdkerrors.ErrInvalidRequest.Wrapf("veid_verification_score must be between 0 and %d", MaxPerformanceScore)
	}

	return nil
}

// GetSigners returns the expected signers for MsgRecordPerformance
func (m *MsgRecordPerformance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// IsValidSlashReason checks if a SlashReason is valid
func IsValidSlashReason(reason SlashReason) bool {
	switch reason {
	case SlashReasonUnspecified:
		return false
	case SlashReasonDoubleSigning,
		SlashReasonDowntime,
		SlashReasonInvalidVEIDAttestation,
		SlashReasonMissedRecomputation,
		SlashReasonInconsistentScore,
		SlashReasonExpiredAttestation,
		SlashReasonDebugModeEnabled,
		SlashReasonNonAllowlistedMeasurement:
		return true
	default:
		return false
	}
}

// Validate performs basic validation for Params
func (p Params) Validate() error {
	// Add param validation as needed
	return nil
}

var _ = fmt.Sprint // silence import unused

