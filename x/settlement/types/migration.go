package types

import "fmt"

// MigrationAuditEntry records a migration action for sensitive data.
type MigrationAuditEntry struct {
	RecordType    string   `json:"record_type"`
	RecordID      string   `json:"record_id"`
	ClearedFields []string `json:"cleared_fields,omitempty"`
	Timestamp     int64    `json:"timestamp"`
	Note          string   `json:"note,omitempty"`
}

func (e *MigrationAuditEntry) Validate() error {
	if e == nil {
		return ErrInvalidParams.Wrap("migration entry is nil")
	}
	if e.RecordType == "" {
		return ErrInvalidParams.Wrap("record_type required")
	}
	if e.RecordID == "" {
		return ErrInvalidParams.Wrap("record_id required")
	}
	if e.Timestamp == 0 {
		return ErrInvalidParams.Wrap("timestamp required")
	}
	return nil
}

func (e *MigrationAuditEntry) String() string {
	if e == nil {
		return "MigrationAuditEntry<nil>"
	}
	return fmt.Sprintf("MigrationAuditEntry{type=%s, id=%s, fields=%v}", e.RecordType, e.RecordID, e.ClearedFields)
}
