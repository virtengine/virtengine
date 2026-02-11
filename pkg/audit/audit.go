package audit

import "time"

// Entry captures a generic audit trail entry.
type Entry struct {
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewEntry builds a new audit entry with UTC timestamps.
func NewEntry(action, actor, details string, timestamp time.Time) Entry {
	return Entry{
		Action:    action,
		Actor:     actor,
		Details:   details,
		Timestamp: timestamp.UTC(),
	}
}
