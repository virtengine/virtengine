// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Event definitions
package types

// Event types for the fraud module
const (
	// EventTypeFraudReportSubmitted is emitted when a new fraud report is submitted
	EventTypeFraudReportSubmitted = "fraud_report_submitted"

	// EventTypeFraudReportAssigned is emitted when a report is assigned to a moderator
	EventTypeFraudReportAssigned = "fraud_report_assigned"

	// EventTypeFraudReportStatusChanged is emitted when a report status changes
	EventTypeFraudReportStatusChanged = "fraud_report_status_changed"

	// EventTypeFraudReportResolved is emitted when a report is resolved
	EventTypeFraudReportResolved = "fraud_report_resolved"

	// EventTypeFraudReportRejected is emitted when a report is rejected
	EventTypeFraudReportRejected = "fraud_report_rejected"

	// EventTypeFraudReportEscalated is emitted when a report is escalated
	EventTypeFraudReportEscalated = "fraud_report_escalated"

	// EventTypeAuditLogCreated is emitted when an audit log entry is created
	EventTypeAuditLogCreated = "audit_log_created"

	// EventTypeModeratorQueueUpdated is emitted when the moderator queue changes
	EventTypeModeratorQueueUpdated = "moderator_queue_updated"
)

// Attribute keys for fraud events
const (
	// AttributeKeyReportID is the fraud report ID attribute
	AttributeKeyReportID = "report_id"

	// AttributeKeyReporter is the reporter address attribute
	AttributeKeyReporter = "reporter"

	// AttributeKeyReportedParty is the reported party address attribute
	AttributeKeyReportedParty = "reported_party"

	// AttributeKeyCategory is the fraud category attribute
	AttributeKeyCategory = "category"

	// AttributeKeyStatus is the report status attribute
	AttributeKeyStatus = "status"

	// AttributeKeyPreviousStatus is the previous status attribute
	AttributeKeyPreviousStatus = "previous_status"

	// AttributeKeyModerator is the moderator address attribute
	AttributeKeyModerator = "moderator"

	// AttributeKeyResolution is the resolution type attribute
	AttributeKeyResolution = "resolution"

	// AttributeKeyAuditLogID is the audit log ID attribute
	AttributeKeyAuditLogID = "audit_log_id"

	// AttributeKeyAction is the action type attribute
	AttributeKeyAction = "action"

	// AttributeKeyTimestamp is the timestamp attribute
	AttributeKeyTimestamp = "timestamp"

	// AttributeKeyBlockHeight is the block height attribute
	AttributeKeyBlockHeight = "block_height"

	// AttributeKeyHasEvidence is whether evidence was provided
	AttributeKeyHasEvidence = "has_evidence"

	// AttributeKeyEvidenceHash is the hash of the encrypted evidence
	AttributeKeyEvidenceHash = "evidence_hash"

	// AttributeKeyQueuePosition is the position in moderator queue
	AttributeKeyQueuePosition = "queue_position"
)
