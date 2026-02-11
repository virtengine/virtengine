package types

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"time"
)

// RetentionAction identifies retention queue actions.
type RetentionAction string

const (
	RetentionActionArchive RetentionAction = "archive"
	RetentionActionPurge   RetentionAction = "purge"
)

// IsValid returns true if action is supported.
func (a RetentionAction) IsValid() bool {
	return a == RetentionActionArchive || a == RetentionActionPurge
}

// RetentionQueueEntry represents an entry in the retention queue.
type RetentionQueueEntry struct {
	RequestID     string          `json:"request_id"`
	Action        RetentionAction `json:"action"`
	ScheduledAt   time.Time       `json:"scheduled_at"`
	Attempts      uint32          `json:"attempts"`
	LastAttemptAt *time.Time      `json:"last_attempt_at,omitempty"`
	LastError     string          `json:"last_error,omitempty"`
}

// RetentionStatus summarizes retention timing and queue state for a request.
type RetentionStatus struct {
	TicketID     string               `json:"ticket_id"`
	Archived     bool                 `json:"archived"`
	Purged       bool                 `json:"purged"`
	ArchiveAt    *time.Time           `json:"archive_at,omitempty"`
	PurgeAt      *time.Time           `json:"purge_at,omitempty"`
	ArchiveQueue *RetentionQueueEntry `json:"archive_queue,omitempty"`
	PurgeQueue   *RetentionQueueEntry `json:"purge_queue,omitempty"`
}

// ParseRetentionQueueKey parses a retention queue key for the given prefix.
func ParseRetentionQueueKey(prefix []byte, key []byte) (int64, string, bool) {
	if len(key) <= len(prefix) {
		return 0, "", false
	}
	if !bytes.HasPrefix(key, prefix) {
		return 0, "", false
	}
	remaining := key[len(prefix):]
	if len(remaining) == 0 {
		return 0, "", false
	}

	// New format uses 8-byte big endian timestamp.
	if len(remaining) >= 9 && remaining[8] == '/' {
		raw := binary.BigEndian.Uint64(remaining[:8])
		if raw > math.MaxInt64 {
			return 0, "", false
		}
		timestamp := int64(raw)
		requestID := string(remaining[9:])
		if requestID == "" {
			return 0, "", false
		}
		return timestamp, requestID, true
	}

	// Legacy format fallback: timestamp as string before '/'.
	for i, b := range remaining {
		if b == '/' {
			tsStr := string(remaining[:i])
			requestID := string(remaining[i+1:])
			if requestID == "" {
				return 0, "", false
			}
			ts, err := parseRetentionTimestamp(tsStr)
			if err != nil {
				return 0, "", false
			}
			return ts, requestID, true
		}
	}
	return 0, "", false
}

func parseRetentionTimestamp(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
