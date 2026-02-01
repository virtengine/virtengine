// Package v1 provides SDK interface implementations for fraud proto message types.
//
// VE-3053: Added SDK interface methods for proto types
// This file is not auto-generated and provides methods required by the Cosmos SDK
// message interface for the generated proto message types.
package v1

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// RouterKey is the router key for the fraud module
	RouterKey = "fraud"

	// TypeMsgSubmitFraudReport is the type for MsgSubmitFraudReport
	TypeMsgSubmitFraudReport = "submit_fraud_report"
	// TypeMsgAssignModerator is the type for MsgAssignModerator
	TypeMsgAssignModerator = "assign_moderator"
	// TypeMsgUpdateReportStatus is the type for MsgUpdateReportStatus
	TypeMsgUpdateReportStatus = "update_report_status"
	// TypeMsgResolveFraudReport is the type for MsgResolveFraudReport
	TypeMsgResolveFraudReport = "resolve_fraud_report"
	// TypeMsgRejectFraudReport is the type for MsgRejectFraudReport
	TypeMsgRejectFraudReport = "reject_fraud_report"
	// TypeMsgEscalateFraudReport is the type for MsgEscalateFraudReport
	TypeMsgEscalateFraudReport = "escalate_fraud_report"
	// TypeMsgUpdateParams is the type for MsgUpdateParams
	TypeMsgUpdateParams = "update_params"

	// MinDescriptionLength is the minimum description length
	MinDescriptionLength = 10
	// MaxDescriptionLength is the maximum description length
	MaxDescriptionLength = 10000
)

// Error codes for validation
var (
	ErrInvalidReporter       = errors.Register("fraud", 2000, "invalid reporter address")
	ErrInvalidReportedParty  = errors.Register("fraud", 2001, "invalid reported party address")
	ErrSelfReport            = errors.Register("fraud", 2008, "invalid report: cannot report yourself")
	ErrInvalidCategory       = errors.Register("fraud", 2013, "invalid fraud category")
	ErrDescriptionTooShort   = errors.Register("fraud", 2016, "description too short")
	ErrDescriptionTooLong    = errors.Register("fraud", 2015, "description too long")
	ErrMissingEvidence       = errors.Register("fraud", 2017, "evidence is required for fraud reports")
	ErrUnauthorizedModerator = errors.Register("fraud", 2006, "unauthorized: moderator role required")
	ErrInvalidReportID       = errors.Register("fraud", 2009, "invalid report ID")
	ErrInvalidStatus         = errors.Register("fraud", 2010, "invalid status transition")
	ErrInvalidResolution     = errors.Register("fraud", 2014, "invalid resolution")
	ErrInvalidResolutionNotes = errors.Register("fraud", 2018, "invalid resolution notes")
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

	if m.Category == FraudCategoryUnspecified {
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

	if m.NewStatus == FraudReportStatusUnspecified {
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

	if m.Resolution == ResolutionTypeUnspecified {
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

