package gdpr

import "time"

// AuditAction represents the GDPR audit action type.
type AuditAction string

const (
	AuditExportRequested   AuditAction = "export_requested"
	AuditExportReady       AuditAction = "export_ready"
	AuditExportFailed      AuditAction = "export_failed"
	AuditDeletionBlocked   AuditAction = "deletion_blocked"
	AuditDeletionFailed    AuditAction = "deletion_failed"
	AuditDeletionCompleted AuditAction = "deletion_completed"
)

// AuditEvent captures a GDPR audit trail record.
type AuditEvent struct {
	Action      AuditAction       `json:"action"`
	RequestID   string            `json:"request_id"`
	DataSubject string            `json:"data_subject"`
	Timestamp   time.Time         `json:"timestamp"`
	Details     map[string]string `json:"details,omitempty"`
}
