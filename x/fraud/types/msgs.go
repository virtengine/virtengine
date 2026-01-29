// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Message types
package types

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
