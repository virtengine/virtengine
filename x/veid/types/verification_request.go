package types

import (
	"fmt"
	"time"
)

// ============================================================================
// Request Status Types
// ============================================================================

// RequestStatus represents the status of a verification request
type RequestStatus string

const (
	// RequestStatusPending indicates the request is waiting to be processed
	RequestStatusPending RequestStatus = "pending"

	// RequestStatusInProgress indicates the request is being processed
	RequestStatusInProgress RequestStatus = "in_progress"

	// RequestStatusCompleted indicates the request was successfully completed
	RequestStatusCompleted RequestStatus = "completed"

	// RequestStatusFailed indicates the request failed due to an error
	RequestStatusFailed RequestStatus = "failed"

	// RequestStatusTimeout indicates the request timed out
	RequestStatusTimeout RequestStatus = "timeout"

	// RequestStatusRejected indicates the request was rejected (e.g., invalid scopes)
	RequestStatusRejected RequestStatus = "rejected"
)

// AllRequestStatuses returns all valid request statuses
func AllRequestStatuses() []RequestStatus {
	return []RequestStatus{
		RequestStatusPending,
		RequestStatusInProgress,
		RequestStatusCompleted,
		RequestStatusFailed,
		RequestStatusTimeout,
		RequestStatusRejected,
	}
}

// IsValidRequestStatus checks if a status is valid
func IsValidRequestStatus(status RequestStatus) bool {
	for _, s := range AllRequestStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// IsFinalRequestStatus checks if a request status is a terminal state
func IsFinalRequestStatus(status RequestStatus) bool {
	switch status {
	case RequestStatusCompleted, RequestStatusFailed, RequestStatusRejected:
		return true
	default:
		return false
	}
}

// ============================================================================
// Verification Request
// ============================================================================

// VerificationRequest represents a request for identity verification
type VerificationRequest struct {
	// RequestID is a unique identifier for this verification request
	RequestID string `json:"request_id"`

	// AccountAddress is the address of the account being verified
	AccountAddress string `json:"account_address"`

	// ScopeIDs are the specific scope identifiers to verify
	ScopeIDs []string `json:"scope_ids"`

	// RequestedAt is when the verification was requested
	RequestedAt time.Time `json:"requested_at"`

	// RequestedBlock is the block height at which the request was created
	RequestedBlock int64 `json:"requested_block"`

	// Status is the current status of the request
	Status RequestStatus `json:"status"`

	// RetryCount tracks how many times this request has been retried
	RetryCount uint32 `json:"retry_count"`

	// LastAttemptAt is when the last processing attempt occurred
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// Priority indicates processing priority (higher = more urgent)
	Priority uint32 `json:"priority"`

	// Metadata contains additional request-specific data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewVerificationRequest creates a new verification request
func NewVerificationRequest(
	requestID string,
	accountAddress string,
	scopeIDs []string,
	requestedAt time.Time,
	requestedBlock int64,
) *VerificationRequest {
	return &VerificationRequest{
		RequestID:      requestID,
		AccountAddress: accountAddress,
		ScopeIDs:       scopeIDs,
		RequestedAt:    requestedAt,
		RequestedBlock: requestedBlock,
		Status:         RequestStatusPending,
		RetryCount:     0,
		Priority:       0,
		Metadata:       make(map[string]string),
	}
}

// Validate validates the verification request
func (r *VerificationRequest) Validate() error {
	if r.RequestID == "" {
		return ErrInvalidVerificationRequest.Wrap("request_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidVerificationRequest.Wrap("account_address cannot be empty")
	}

	if len(r.ScopeIDs) == 0 {
		return ErrInvalidVerificationRequest.Wrap("at least one scope_id required")
	}

	if r.RequestedAt.IsZero() {
		return ErrInvalidVerificationRequest.Wrap("requested_at cannot be zero")
	}

	if r.RequestedBlock < 0 {
		return ErrInvalidVerificationRequest.Wrap("requested_block cannot be negative")
	}

	if !IsValidRequestStatus(r.Status) {
		return ErrInvalidVerificationRequest.Wrapf("invalid status: %s", r.Status)
	}

	return nil
}

// IsRetryable checks if the request can be retried
func (r *VerificationRequest) IsRetryable(maxRetries uint32) bool {
	if IsFinalRequestStatus(r.Status) {
		return false
	}
	return r.RetryCount < maxRetries
}

// IncrementRetry increments the retry count and updates last attempt time
func (r *VerificationRequest) IncrementRetry(attemptTime time.Time) {
	r.RetryCount++
	r.LastAttemptAt = &attemptTime
}

// SetInProgress marks the request as in progress
func (r *VerificationRequest) SetInProgress(attemptTime time.Time) {
	r.Status = RequestStatusInProgress
	r.LastAttemptAt = &attemptTime
}

// SetCompleted marks the request as completed
func (r *VerificationRequest) SetCompleted() {
	r.Status = RequestStatusCompleted
}

// SetFailed marks the request as failed with a reason
func (r *VerificationRequest) SetFailed(reason string) {
	r.Status = RequestStatusFailed
	r.Metadata["failure_reason"] = reason
}

// SetTimeout marks the request as timed out
func (r *VerificationRequest) SetTimeout() {
	r.Status = RequestStatusTimeout
}

// SetRejected marks the request as rejected with a reason
func (r *VerificationRequest) SetRejected(reason string) {
	r.Status = RequestStatusRejected
	r.Metadata["rejection_reason"] = reason
}

// String returns a string representation of the request
func (r *VerificationRequest) String() string {
	return fmt.Sprintf("VerificationRequest{ID: %s, Account: %s, Status: %s, Scopes: %d}",
		r.RequestID, r.AccountAddress, r.Status, len(r.ScopeIDs))
}

// ============================================================================
// Store Keys
// ============================================================================

var (
	// PrefixVerificationRequest is the prefix for verification request storage
	// Key: PrefixVerificationRequest | request_id -> VerificationRequest
	PrefixVerificationRequest = []byte{0x10}

	// PrefixVerificationRequestByAccount is the prefix for lookup by account
	// Key: PrefixVerificationRequestByAccount | account_address -> []request_id
	PrefixVerificationRequestByAccount = []byte{0x11}

	// PrefixPendingVerificationRequest is the prefix for pending request queue
	// Key: PrefixPendingVerificationRequest | block_height | request_id -> nil
	PrefixPendingVerificationRequest = []byte{0x12}
)

// VerificationRequestKey returns the store key for a verification request
func VerificationRequestKey(requestID string) []byte {
	key := make([]byte, 0, len(PrefixVerificationRequest)+len(requestID))
	key = append(key, PrefixVerificationRequest...)
	key = append(key, []byte(requestID)...)
	return key
}

// VerificationRequestByAccountKey returns the key for requests by account
func VerificationRequestByAccountKey(accountAddress string) []byte {
	key := make([]byte, 0, len(PrefixVerificationRequestByAccount)+len(accountAddress))
	key = append(key, PrefixVerificationRequestByAccount...)
	key = append(key, []byte(accountAddress)...)
	return key
}

// PendingVerificationRequestKey returns the key for pending request queue entry
func PendingVerificationRequestKey(blockHeight int64, requestID string) []byte {
	heightBytes := encodeInt64(blockHeight)
	key := make([]byte, 0, len(PrefixPendingVerificationRequest)+8+1+len(requestID))
	key = append(key, PrefixPendingVerificationRequest...)
	key = append(key, heightBytes...)
	key = append(key, byte('/'))
	key = append(key, []byte(requestID)...)
	return key
}

// PendingVerificationRequestPrefixKey returns the prefix for all pending requests
func PendingVerificationRequestPrefixKey() []byte {
	return PrefixPendingVerificationRequest
}
