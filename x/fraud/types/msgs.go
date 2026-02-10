// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Message types
// VE-3053: SDK interface methods are defined in sdk/go/node/fraud/v1/msgs.go
//
// Note: SDK Msg interface methods (ValidateBasic, GetSigners, Route, Type)
// are defined directly on the proto types in the sdk/go/node/fraud/v1 package.
// This file only provides re-exports of message type constants for local use.
package types

// Message type constants - defined locally to avoid import cycles
// (same values as in sdk/go/node/fraud/v1/msgs.go)
const (
	TypeMsgSubmitFraudReport   = "submit_fraud_report"
	TypeMsgAssignModerator     = "assign_moderator"
	TypeMsgUpdateReportStatus  = "update_report_status"
	TypeMsgResolveFraudReport  = "resolve_fraud_report"
	TypeMsgRejectFraudReport   = "reject_fraud_report"
	TypeMsgEscalateFraudReport = "escalate_fraud_report"
	TypeMsgUpdateParams        = "update_params"
)
