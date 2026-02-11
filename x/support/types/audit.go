package types

import (
	"fmt"
	"time"
)

// SupportAuditAction identifies audit trail actions for support requests.
type SupportAuditAction string

const (
	SupportAuditActionRequestCreated   SupportAuditAction = "support_request_created"
	SupportAuditActionRequestUpdated   SupportAuditAction = "support_request_updated"
	SupportAuditActionStatusChanged    SupportAuditAction = "support_request_status_changed"
	SupportAuditActionResponseAdded    SupportAuditAction = "support_response_added"
	SupportAuditActionRequestArchived  SupportAuditAction = "support_request_archived"
	SupportAuditActionRequestPurged    SupportAuditAction = "support_request_purged"
	SupportAuditActionRetentionEnqueue SupportAuditAction = "support_retention_enqueued"
	SupportAuditActionRetentionRetry   SupportAuditAction = "support_retention_retry"
)

// SupportAuditEntry represents an audit trail entry for a support request.
type SupportAuditEntry struct {
	Action      SupportAuditAction `json:"action"`
	PerformedBy string             `json:"performed_by"`
	Details     string             `json:"details,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
	BlockHeight int64              `json:"block_height"`
}

// NewSupportAuditEntry constructs a new audit entry.
func NewSupportAuditEntry(action SupportAuditAction, performedBy string, details string, timestamp time.Time, height int64) SupportAuditEntry {
	return SupportAuditEntry{
		Action:      action,
		PerformedBy: performedBy,
		Details:     details,
		Timestamp:   timestamp.UTC(),
		BlockHeight: height,
	}
}

// Validate validates the audit entry.
func (e *SupportAuditEntry) Validate() error {
	if e == nil {
		return fmt.Errorf("audit entry is nil")
	}
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}
	if len(e.Action) > 128 {
		return fmt.Errorf("action exceeds maximum length of 128")
	}
	if e.PerformedBy == "" {
		return fmt.Errorf("performed_by is required")
	}
	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}
