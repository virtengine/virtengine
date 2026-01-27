// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Message types
package types

import (
	"fmt"

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

var (
	_ sdk.Msg = &MsgSubmitFraudReport{}
	_ sdk.Msg = &MsgAssignModerator{}
	_ sdk.Msg = &MsgUpdateReportStatus{}
	_ sdk.Msg = &MsgResolveFraudReport{}
	_ sdk.Msg = &MsgRejectFraudReport{}
	_ sdk.Msg = &MsgEscalateFraudReport{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// MsgSubmitFraudReport defines the message for submitting a fraud report
type MsgSubmitFraudReport struct {
	// Reporter is the provider address submitting the report
	Reporter string `json:"reporter"`

	// ReportedParty is the address of the party being reported
	ReportedParty string `json:"reported_party"`

	// Category is the type of fraud being reported
	Category FraudCategory `json:"category"`

	// Description is the detailed description of the fraud
	Description string `json:"description"`

	// Evidence contains encrypted evidence attachments
	Evidence []EncryptedEvidence `json:"evidence"`

	// RelatedOrderIDs are optional order IDs related to this fraud
	RelatedOrderIDs []string `json:"related_order_ids,omitempty"`
}

// NewMsgSubmitFraudReport creates a new MsgSubmitFraudReport
func NewMsgSubmitFraudReport(
	reporter string,
	reportedParty string,
	category FraudCategory,
	description string,
	evidence []EncryptedEvidence,
	relatedOrderIDs []string,
) *MsgSubmitFraudReport {
	return &MsgSubmitFraudReport{
		Reporter:        reporter,
		ReportedParty:   reportedParty,
		Category:        category,
		Description:     description,
		Evidence:        evidence,
		RelatedOrderIDs: relatedOrderIDs,
	}
}

// Route returns the message route
func (msg *MsgSubmitFraudReport) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgSubmitFraudReport) Type() string { return TypeMsgSubmitFraudReport }

// ValidateBasic performs basic validation
func (msg *MsgSubmitFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Reporter); err != nil {
		return ErrInvalidReporter.Wrap(err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(msg.ReportedParty); err != nil {
		return ErrInvalidReportedParty.Wrap(err.Error())
	}
	if msg.Reporter == msg.ReportedParty {
		return ErrSelfReport
	}
	if !msg.Category.IsValid() {
		return ErrInvalidCategory
	}
	if len(msg.Description) < MinDescriptionLength {
		return ErrDescriptionTooShort.Wrapf("minimum %d characters required", MinDescriptionLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong.Wrapf("maximum %d characters allowed", MaxDescriptionLength)
	}
	if len(msg.Evidence) == 0 {
		return ErrMissingEvidence
	}
	for i, ev := range msg.Evidence {
		if err := ev.Validate(); err != nil {
			return ErrInvalidEvidence.Wrapf("evidence %d: %v", i, err)
		}
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgSubmitFraudReport) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgSubmitFraudReport) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Reporter)
	return []sdk.AccAddress{addr}
}

// MsgAssignModerator defines the message for assigning a moderator to a report
type MsgAssignModerator struct {
	// Moderator is the moderator performing the assignment
	Moderator string `json:"moderator"`

	// ReportID is the fraud report ID to assign
	ReportID string `json:"report_id"`

	// AssignTo is the moderator to assign (can be self)
	AssignTo string `json:"assign_to"`
}

// NewMsgAssignModerator creates a new MsgAssignModerator
func NewMsgAssignModerator(moderator, reportID, assignTo string) *MsgAssignModerator {
	return &MsgAssignModerator{
		Moderator: moderator,
		ReportID:  reportID,
		AssignTo:  assignTo,
	}
}

// Route returns the message route
func (msg *MsgAssignModerator) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgAssignModerator) Type() string { return TypeMsgAssignModerator }

// ValidateBasic performs basic validation
func (msg *MsgAssignModerator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if msg.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if _, err := sdk.AccAddressFromBech32(msg.AssignTo); err != nil {
		return ErrUnauthorizedModerator.Wrap("invalid assign_to address")
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgAssignModerator) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgAssignModerator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{addr}
}

// MsgUpdateReportStatus defines the message for updating a report's status
type MsgUpdateReportStatus struct {
	// Moderator is the moderator performing the update
	Moderator string `json:"moderator"`

	// ReportID is the fraud report ID
	ReportID string `json:"report_id"`

	// NewStatus is the new status for the report
	NewStatus FraudReportStatus `json:"new_status"`

	// Notes are optional notes about the status change
	Notes string `json:"notes,omitempty"`
}

// NewMsgUpdateReportStatus creates a new MsgUpdateReportStatus
func NewMsgUpdateReportStatus(moderator, reportID string, newStatus FraudReportStatus, notes string) *MsgUpdateReportStatus {
	return &MsgUpdateReportStatus{
		Moderator: moderator,
		ReportID:  reportID,
		NewStatus: newStatus,
		Notes:     notes,
	}
}

// Route returns the message route
func (msg *MsgUpdateReportStatus) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgUpdateReportStatus) Type() string { return TypeMsgUpdateReportStatus }

// ValidateBasic performs basic validation
func (msg *MsgUpdateReportStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if msg.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if !msg.NewStatus.IsValid() {
		return ErrInvalidStatus.Wrap("invalid new status")
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgUpdateReportStatus) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgUpdateReportStatus) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{addr}
}

// MsgResolveFraudReport defines the message for resolving a fraud report
type MsgResolveFraudReport struct {
	// Moderator is the moderator resolving the report
	Moderator string `json:"moderator"`

	// ReportID is the fraud report ID
	ReportID string `json:"report_id"`

	// Resolution is the resolution type
	Resolution ResolutionType `json:"resolution"`

	// Notes are resolution notes
	Notes string `json:"notes"`
}

// NewMsgResolveFraudReport creates a new MsgResolveFraudReport
func NewMsgResolveFraudReport(moderator, reportID string, resolution ResolutionType, notes string) *MsgResolveFraudReport {
	return &MsgResolveFraudReport{
		Moderator:  moderator,
		ReportID:   reportID,
		Resolution: resolution,
		Notes:      notes,
	}
}

// Route returns the message route
func (msg *MsgResolveFraudReport) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgResolveFraudReport) Type() string { return TypeMsgResolveFraudReport }

// ValidateBasic performs basic validation
func (msg *MsgResolveFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if msg.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if !msg.Resolution.IsValid() {
		return ErrInvalidResolution
	}
	if len(msg.Notes) > MaxResolutionNotesLength {
		return ErrInvalidResolutionNotes.Wrapf("maximum %d characters allowed", MaxResolutionNotesLength)
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgResolveFraudReport) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgResolveFraudReport) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{addr}
}

// MsgRejectFraudReport defines the message for rejecting a fraud report
type MsgRejectFraudReport struct {
	// Moderator is the moderator rejecting the report
	Moderator string `json:"moderator"`

	// ReportID is the fraud report ID
	ReportID string `json:"report_id"`

	// Notes are rejection notes
	Notes string `json:"notes"`
}

// NewMsgRejectFraudReport creates a new MsgRejectFraudReport
func NewMsgRejectFraudReport(moderator, reportID string, notes string) *MsgRejectFraudReport {
	return &MsgRejectFraudReport{
		Moderator: moderator,
		ReportID:  reportID,
		Notes:     notes,
	}
}

// Route returns the message route
func (msg *MsgRejectFraudReport) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgRejectFraudReport) Type() string { return TypeMsgRejectFraudReport }

// ValidateBasic performs basic validation
func (msg *MsgRejectFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if msg.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if len(msg.Notes) > MaxResolutionNotesLength {
		return ErrInvalidResolutionNotes.Wrapf("maximum %d characters allowed", MaxResolutionNotesLength)
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgRejectFraudReport) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgRejectFraudReport) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{addr}
}

// MsgEscalateFraudReport defines the message for escalating a fraud report
type MsgEscalateFraudReport struct {
	// Moderator is the moderator escalating the report
	Moderator string `json:"moderator"`

	// ReportID is the fraud report ID
	ReportID string `json:"report_id"`

	// Reason is the escalation reason
	Reason string `json:"reason"`
}

// NewMsgEscalateFraudReport creates a new MsgEscalateFraudReport
func NewMsgEscalateFraudReport(moderator, reportID string, reason string) *MsgEscalateFraudReport {
	return &MsgEscalateFraudReport{
		Moderator: moderator,
		ReportID:  reportID,
		Reason:    reason,
	}
}

// Route returns the message route
func (msg *MsgEscalateFraudReport) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgEscalateFraudReport) Type() string { return TypeMsgEscalateFraudReport }

// ValidateBasic performs basic validation
func (msg *MsgEscalateFraudReport) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Moderator); err != nil {
		return ErrUnauthorizedModerator.Wrap(err.Error())
	}
	if msg.ReportID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if msg.Reason == "" {
		return fmt.Errorf("escalation reason is required")
	}
	return nil
}

// GetSignBytes returns the bytes for signing
func (msg *MsgEscalateFraudReport) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgEscalateFraudReport) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Moderator)
	return []sdk.AccAddress{addr}
}

// MsgUpdateParams defines the message for updating module parameters
type MsgUpdateParams struct {
	// Authority is the address that controls the module
	Authority string `json:"authority"`

	// Params are the new parameters
	Params Params `json:"params"`
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the message route
func (msg *MsgUpdateParams) Route() string { return RouterKey }

// Type returns the message type
func (msg *MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// ValidateBasic performs basic validation
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	return msg.Params.Validate()
}

// GetSignBytes returns the bytes for signing
func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the signers
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
