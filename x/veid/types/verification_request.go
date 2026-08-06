package types

import (
	"crypto/sha256"
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

	// InferenceProfileSnapshot is the immutable production inference profile
	// captured when the request is created. It is internal JSON state only; it
	// intentionally does not change protobuf/API contracts.
	InferenceProfileSnapshot *InferenceProfileSnapshot `json:"inference_profile_snapshot,omitempty"`
}

// InferenceProfileSnapshot captures the exact committed inference profile a
// verification request was issued under. Receipts for the request must match
// this snapshot, and vote-extension staging additionally requires the same
// profile to still be the single active bundle commitment.
type InferenceProfileSnapshot struct {
	PipelineVersion         string `json:"pipeline_version"`
	RuntimeImageDigest      []byte `json:"runtime_image_digest"`
	RuntimeDigest           []byte `json:"runtime_digest"`
	ModelManifestDigest     []byte `json:"model_manifest_digest"`
	ModelDigest             []byte `json:"model_digest"`
	DeterminismConfigDigest []byte `json:"determinism_config_digest"`
	FeatureSchemaDigest     []byte `json:"feature_schema_digest"`
	ActivationHeight        int64  `json:"activation_height"`
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

	if r.InferenceProfileSnapshot != nil {
		if err := r.InferenceProfileSnapshot.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SetInferenceProfileSnapshot stores a defensive copy of a validated profile
// snapshot on the request.
func (r *VerificationRequest) SetInferenceProfileSnapshot(snapshot *InferenceProfileSnapshot) error {
	if snapshot == nil {
		return ErrInvalidVerificationRequest.Wrap("inference profile snapshot is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	copied := snapshot.DeepCopy()
	if err := copied.Validate(); err != nil {
		return err
	}
	r.InferenceProfileSnapshot = copied
	return nil
}

// SetInferenceProfileSnapshotForTest explicitly backfills legacy direct test
// fixtures. Production request creation must use SetInferenceProfileSnapshot
// from committed active profile state instead of implicit current-state repair.
func (r *VerificationRequest) SetInferenceProfileSnapshotForTest(snapshot *InferenceProfileSnapshot) error {
	return r.SetInferenceProfileSnapshot(snapshot)
}

// RequireInferenceProfileSnapshot returns a defensive copy of the required
// profile snapshot for production receipt validation.
func (r *VerificationRequest) RequireInferenceProfileSnapshot() (*InferenceProfileSnapshot, error) {
	if r == nil || r.InferenceProfileSnapshot == nil {
		return nil, ErrInvalidVerificationRequest.Wrap("inference profile snapshot is required")
	}
	copied := r.InferenceProfileSnapshot.DeepCopy()
	if err := copied.Validate(); err != nil {
		return nil, err
	}
	return copied, nil
}

// DeepCopy returns a snapshot copy with independent digest storage.
func (s *InferenceProfileSnapshot) DeepCopy() *InferenceProfileSnapshot {
	if s == nil {
		return nil
	}
	return &InferenceProfileSnapshot{
		PipelineVersion:         s.PipelineVersion,
		RuntimeImageDigest:      append([]byte(nil), s.RuntimeImageDigest...),
		RuntimeDigest:           append([]byte(nil), s.RuntimeDigest...),
		ModelManifestDigest:     append([]byte(nil), s.ModelManifestDigest...),
		ModelDigest:             append([]byte(nil), s.ModelDigest...),
		DeterminismConfigDigest: append([]byte(nil), s.DeterminismConfigDigest...),
		FeatureSchemaDigest:     append([]byte(nil), s.FeatureSchemaDigest...),
		ActivationHeight:        s.ActivationHeight,
	}
}

// Validate rejects missing, malformed, or aliased digest fields and invalid
// activation metadata. RuntimeImageDigest and RuntimeDigest may have identical
// contents because the current receipt semantics commit both runtime fields to
// the same image hash, but they must not share mutable backing storage.
func (s *InferenceProfileSnapshot) Validate() error {
	if s == nil {
		return ErrInvalidVerificationRequest.Wrap("inference profile snapshot is required")
	}
	if s.PipelineVersion == "" {
		return ErrInvalidVerificationRequest.Wrap("inference profile pipeline_version is required")
	}
	if s.ActivationHeight <= 0 {
		return ErrInvalidVerificationRequest.Wrap("inference profile activation_height must be positive")
	}
	digests := []struct {
		name  string
		value []byte
	}{
		{"runtime_image_digest", s.RuntimeImageDigest},
		{"runtime_digest", s.RuntimeDigest},
		{"model_manifest_digest", s.ModelManifestDigest},
		{"model_digest", s.ModelDigest},
		{"determinism_config_digest", s.DeterminismConfigDigest},
		{"feature_schema_digest", s.FeatureSchemaDigest},
	}
	for _, digest := range digests {
		if len(digest.value) != sha256.Size {
			return ErrInvalidVerificationRequest.Wrapf("inference profile %s must be SHA-256", digest.name)
		}
	}
	for i := range digests {
		for j := i + 1; j < len(digests); j++ {
			if digestSlicesAlias(digests[i].value, digests[j].value) {
				return ErrInvalidVerificationRequest.Wrap("inference profile digest fields must not alias")
			}
		}
	}
	return nil
}

func digestSlicesAlias(a, b []byte) bool {
	return len(a) > 0 && len(b) > 0 && &a[0] == &b[0]
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
