// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Message types
// VE-3053: Added SDK interface methods for proto types
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgSubmitFraudReport   = "submit_fraud_report"
	TypeMsgAssignModerator     = "assign_moderator"
	TypeMsgUpdateReportStatus  = "update_report_status"
	TypeMsgResolveFraudReport  = "resolve_fraud_report"
	TypeMsgRejectFraudReport   = "reject_fraud_report"
	TypeMsgEscalateFraudReport = "escalate_fraud_report"
	TypeMsgUpdateParams        = "update_params"
)

// =============================================================================
// MsgSubmitFraudReport SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgSubmitFraudReport{}

// ValidateBasic implements sdk.Msg
func (m *MsgSubmitFraudReport) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Reporter)
	if err != nil {
		return ErrInvalidReporter.Wrap("invalid reporter address")
	}

	_, err = sdk.AccAddressFromBech32(m.ReportedParty)
	if err != nil {
		return ErrInvalidReportedParty.Wrap("invalid reported party address")
	}

	if m.Reporter == m.ReportedParty {
		return ErrSelfReport
	}

	if m.Category == FraudCategoryPBUnspecified {
		return ErrInvalidCategory.Wrap("category must be specified")
	}

	if len(m.Description) < MinDescriptionLength {
		return ErrDescriptionTooShort
	}

	if len(m.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}

	if len(m.Evidence) == 0 {
		return ErrMissingEvidence
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgSubmitFraudReport) GetSigners() []sdk.AccAddress {
	reporter, _ := sdk.AccAddressFromBech32(m.Reporter)
	return []sdk.AccAddress{reporter}
}

// Route implements sdk.Msg (deprecated, but required for compatibility)
func (m *MsgSubmitFraudReport) Route() string { return RouterKey }

// Type implements sdk.Msg (deprecated, but required for compatibility)
func (m *MsgSubmitFraudReport) Type() string { return TypeMsgSubmitFraudReport }

// =============================================================================
// MsgAssignModerator SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgAssignModerator{}

// ValidateBasic implements sdk.Msg
func (m *MsgAssignModerator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid moderator address")
	}

	if m.ReportId == "" {
		return ErrInvalidReportID.Wrap("report_id cannot be empty")
	}

	_, err = sdk.AccAddressFromBech32(m.AssignTo)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid assign_to address")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgAssignModerator) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{moderator}
}

// Route implements sdk.Msg
func (m *MsgAssignModerator) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgAssignModerator) Type() string { return TypeMsgAssignModerator }

// =============================================================================
// MsgUpdateReportStatus SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgUpdateReportStatus{}

// ValidateBasic implements sdk.Msg
func (m *MsgUpdateReportStatus) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid moderator address")
	}

	if m.ReportId == "" {
		return ErrInvalidReportID.Wrap("report_id cannot be empty")
	}

	if m.NewStatus == FraudReportStatusPBUnspecified {
		return ErrInvalidStatus.Wrap("new_status must be specified")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgUpdateReportStatus) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{moderator}
}

// Route implements sdk.Msg
func (m *MsgUpdateReportStatus) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgUpdateReportStatus) Type() string { return TypeMsgUpdateReportStatus }

// =============================================================================
// MsgResolveFraudReport SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgResolveFraudReport{}

// ValidateBasic implements sdk.Msg
func (m *MsgResolveFraudReport) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid moderator address")
	}

	if m.ReportId == "" {
		return ErrInvalidReportID.Wrap("report_id cannot be empty")
	}

	if m.Resolution == ResolutionTypePBUnspecified {
		return ErrInvalidResolution.Wrap("resolution must be specified")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgResolveFraudReport) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{moderator}
}

// Route implements sdk.Msg
func (m *MsgResolveFraudReport) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgResolveFraudReport) Type() string { return TypeMsgResolveFraudReport }

// =============================================================================
// MsgRejectFraudReport SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgRejectFraudReport{}

// ValidateBasic implements sdk.Msg
func (m *MsgRejectFraudReport) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid moderator address")
	}

	if m.ReportId == "" {
		return ErrInvalidReportID.Wrap("report_id cannot be empty")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgRejectFraudReport) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{moderator}
}

// Route implements sdk.Msg
func (m *MsgRejectFraudReport) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgRejectFraudReport) Type() string { return TypeMsgRejectFraudReport }

// =============================================================================
// MsgEscalateFraudReport SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgEscalateFraudReport{}

// ValidateBasic implements sdk.Msg
func (m *MsgEscalateFraudReport) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Moderator)
	if err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid moderator address")
	}

	if m.ReportId == "" {
		return ErrInvalidReportID.Wrap("report_id cannot be empty")
	}

	if m.Reason == "" {
		return ErrInvalidResolutionNotes.Wrap("reason cannot be empty for escalation")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgEscalateFraudReport) GetSigners() []sdk.AccAddress {
	moderator, _ := sdk.AccAddressFromBech32(m.Moderator)
	return []sdk.AccAddress{moderator}
}

// Route implements sdk.Msg
func (m *MsgEscalateFraudReport) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgEscalateFraudReport) Type() string { return TypeMsgEscalateFraudReport }

// =============================================================================
// MsgUpdateParams SDK interface methods
// =============================================================================

var _ sdk.Msg = &MsgUpdateParams{}

// ValidateBasic implements sdk.Msg
func (m *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return ErrInvalidReporter.Wrap("invalid authority address")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{authority}
}

// Route implements sdk.Msg
func (m *MsgUpdateParams) Route() string { return RouterKey }

// Type implements sdk.Msg
func (m *MsgUpdateParams) Type() string { return TypeMsgUpdateParams }
