// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Message types and MsgServer interface
package types

import (
	"context"
)

// MsgServer is the server API for the Fraud module's messages
type MsgServer interface {
	// SubmitFraudReport submits a new fraud report
	SubmitFraudReport(ctx context.Context, msg *MsgSubmitFraudReport) (*MsgSubmitFraudReportResponse, error)
	// AssignModerator assigns a moderator to a fraud report
	AssignModerator(ctx context.Context, msg *MsgAssignModerator) (*MsgAssignModeratorResponse, error)
	// UpdateReportStatus updates the status of a fraud report
	UpdateReportStatus(ctx context.Context, msg *MsgUpdateReportStatus) (*MsgUpdateReportStatusResponse, error)
	// ResolveFraudReport resolves a fraud report
	ResolveFraudReport(ctx context.Context, msg *MsgResolveFraudReport) (*MsgResolveFraudReportResponse, error)
	// RejectFraudReport rejects a fraud report
	RejectFraudReport(ctx context.Context, msg *MsgRejectFraudReport) (*MsgRejectFraudReportResponse, error)
	// EscalateFraudReport escalates a fraud report
	EscalateFraudReport(ctx context.Context, msg *MsgEscalateFraudReport) (*MsgEscalateFraudReportResponse, error)
	// UpdateParams updates module parameters (governance only)
	UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// MsgSubmitFraudReport is the message for submitting a fraud report
type MsgSubmitFraudReport struct {
	// Reporter is the provider address submitting the report
	Reporter string `json:"reporter"`

	// ReportedParty is the address of the party being reported
	ReportedParty string `json:"reported_party"`

	// Category is the fraud category
	Category FraudCategory `json:"category"`

	// Description is the detailed description of the fraud
	Description string `json:"description"`

	// Evidence contains encrypted evidence attachments
	Evidence []EncryptedEvidence `json:"evidence"`

	// RelatedOrderIDs are optional order IDs related to this fraud
	RelatedOrderIDs []string `json:"related_order_ids,omitempty"`
}

// MsgSubmitFraudReportResponse is the response for MsgSubmitFraudReport
type MsgSubmitFraudReportResponse struct {
	// ReportID is the unique identifier for the created report
	ReportID string `json:"report_id"`
}

// MsgAssignModerator is the message for assigning a moderator to a report
type MsgAssignModerator struct {
	// Moderator is the address of the moderator making the assignment
	Moderator string `json:"moderator"`

	// ReportID is the ID of the fraud report
	ReportID string `json:"report_id"`

	// AssignTo is the address of the moderator to assign
	AssignTo string `json:"assign_to"`
}

// MsgAssignModeratorResponse is the response for MsgAssignModerator
type MsgAssignModeratorResponse struct{}

// MsgUpdateReportStatus is the message for updating report status
type MsgUpdateReportStatus struct {
	// Moderator is the address of the moderator updating the status
	Moderator string `json:"moderator"`

	// ReportID is the ID of the fraud report
	ReportID string `json:"report_id"`

	// NewStatus is the new status for the report
	NewStatus FraudReportStatus `json:"new_status"`

	// Notes are optional notes about the status change
	Notes string `json:"notes,omitempty"`
}

// MsgUpdateReportStatusResponse is the response for MsgUpdateReportStatus
type MsgUpdateReportStatusResponse struct{}

// MsgResolveFraudReport is the message for resolving a fraud report
type MsgResolveFraudReport struct {
	// Moderator is the address of the moderator resolving the report
	Moderator string `json:"moderator"`

	// ReportID is the ID of the fraud report
	ReportID string `json:"report_id"`

	// Resolution is the resolution type
	Resolution ResolutionType `json:"resolution"`

	// Notes are the resolution notes
	Notes string `json:"notes,omitempty"`
}

// MsgResolveFraudReportResponse is the response for MsgResolveFraudReport
type MsgResolveFraudReportResponse struct{}

// MsgRejectFraudReport is the message for rejecting a fraud report
type MsgRejectFraudReport struct {
	// Moderator is the address of the moderator rejecting the report
	Moderator string `json:"moderator"`

	// ReportID is the ID of the fraud report
	ReportID string `json:"report_id"`

	// Notes are the rejection notes
	Notes string `json:"notes,omitempty"`
}

// MsgRejectFraudReportResponse is the response for MsgRejectFraudReport
type MsgRejectFraudReportResponse struct{}

// MsgEscalateFraudReport is the message for escalating a fraud report
type MsgEscalateFraudReport struct {
	// Moderator is the address of the moderator escalating the report
	Moderator string `json:"moderator"`

	// ReportID is the ID of the fraud report
	ReportID string `json:"report_id"`

	// Reason is the reason for escalation
	Reason string `json:"reason"`
}

// MsgEscalateFraudReportResponse is the response for MsgEscalateFraudReport
type MsgEscalateFraudReportResponse struct{}

// MsgUpdateParams is the message for updating module parameters
type MsgUpdateParams struct {
	// Authority is the governance account address
	Authority string `json:"authority"`

	// Params are the new module parameters
	Params Params `json:"params"`
}

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}
