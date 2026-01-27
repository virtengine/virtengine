package types

import (
	"time"
)

// VerificationStatus represents the verification state of an identity scope
type VerificationStatus string

// Verification status constants
const (
	// VerificationStatusUnknown indicates an uninitialized or unknown status
	VerificationStatusUnknown VerificationStatus = "unknown"

	// VerificationStatusPending indicates the scope is awaiting verification
	VerificationStatusPending VerificationStatus = "pending"

	// VerificationStatusInProgress indicates verification is actively being processed
	VerificationStatusInProgress VerificationStatus = "in_progress"

	// VerificationStatusVerified indicates the scope has been successfully verified
	VerificationStatusVerified VerificationStatus = "verified"

	// VerificationStatusRejected indicates the scope failed verification
	VerificationStatusRejected VerificationStatus = "rejected"

	// VerificationStatusExpired indicates the verification has expired
	VerificationStatusExpired VerificationStatus = "expired"

	// VerificationStatusNeedsAdditionalFactor indicates borderline score requires MFA
	// This is set when facial verification score is in the borderline band
	VerificationStatusNeedsAdditionalFactor VerificationStatus = "needs_additional_factor"

	// VerificationStatusAdditionalFactorPending indicates MFA challenge is in progress
	// This is set after MFA challenge is created but not yet completed
	VerificationStatusAdditionalFactorPending VerificationStatus = "additional_factor_pending"
)

// AllVerificationStatuses returns all valid verification statuses
func AllVerificationStatuses() []VerificationStatus {
	return []VerificationStatus{
		VerificationStatusUnknown,
		VerificationStatusPending,
		VerificationStatusInProgress,
		VerificationStatusVerified,
		VerificationStatusRejected,
		VerificationStatusExpired,
		VerificationStatusNeedsAdditionalFactor,
		VerificationStatusAdditionalFactorPending,
	}
}

// IsValidVerificationStatus checks if a status is valid
func IsValidVerificationStatus(status VerificationStatus) bool {
	for _, s := range AllVerificationStatuses() {
		if s == status {
			return true
		}
	}
	return false
}

// IsFinalStatus checks if the status is a final state
func IsFinalStatus(status VerificationStatus) bool {
	switch status {
	case VerificationStatusVerified, VerificationStatusRejected, VerificationStatusExpired:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if a status can transition to another status
func (s VerificationStatus) CanTransitionTo(target VerificationStatus) bool {
	// Define valid state transitions
	transitions := map[VerificationStatus][]VerificationStatus{
		VerificationStatusUnknown:                 {VerificationStatusPending},
		VerificationStatusPending:                 {VerificationStatusInProgress, VerificationStatusRejected, VerificationStatusExpired},
		VerificationStatusInProgress:              {VerificationStatusVerified, VerificationStatusRejected, VerificationStatusPending, VerificationStatusNeedsAdditionalFactor},
		VerificationStatusVerified:                {VerificationStatusExpired},
		VerificationStatusRejected:                {VerificationStatusPending}, // Allow re-verification after rejection
		VerificationStatusExpired:                 {VerificationStatusPending}, // Allow re-verification after expiry
		VerificationStatusNeedsAdditionalFactor:   {VerificationStatusAdditionalFactorPending, VerificationStatusRejected, VerificationStatusExpired},
		VerificationStatusAdditionalFactorPending: {VerificationStatusVerified, VerificationStatusRejected, VerificationStatusExpired, VerificationStatusNeedsAdditionalFactor},
	}

	validTargets, exists := transitions[s]
	if !exists {
		return false
	}

	for _, valid := range validTargets {
		if valid == target {
			return true
		}
	}
	return false
}

// VerificationEvent represents a verification status change event
type VerificationEvent struct {
	// EventID is a unique identifier for this event
	EventID string `json:"event_id"`

	// ScopeID is the scope this event relates to (empty for account-level events)
	ScopeID string `json:"scope_id,omitempty"`

	// PreviousStatus is the status before this event
	PreviousStatus VerificationStatus `json:"previous_status"`

	// NewStatus is the status after this event
	NewStatus VerificationStatus `json:"new_status"`

	// Timestamp is when this event occurred
	Timestamp time.Time `json:"timestamp"`

	// Reason is the reason for the status change
	Reason string `json:"reason,omitempty"`

	// ValidatorAddress is the address of the validator that processed this (if applicable)
	ValidatorAddress string `json:"validator_address,omitempty"`

	// Score is the identity score at the time of this event (if applicable)
	Score *uint32 `json:"score,omitempty"`

	// Metadata contains additional event-specific data
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewVerificationEvent creates a new verification event
func NewVerificationEvent(
	eventID string,
	scopeID string,
	previousStatus VerificationStatus,
	newStatus VerificationStatus,
	timestamp time.Time,
	reason string,
) *VerificationEvent {
	return &VerificationEvent{
		EventID:        eventID,
		ScopeID:        scopeID,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Timestamp:      timestamp,
		Reason:         reason,
		Metadata:       make(map[string]string),
	}
}

// Validate validates the verification event
func (e *VerificationEvent) Validate() error {
	if e.EventID == "" {
		return ErrInvalidVerificationEvent.Wrap("event_id cannot be empty")
	}

	if !IsValidVerificationStatus(e.PreviousStatus) {
		return ErrInvalidVerificationStatus.Wrapf("invalid previous status: %s", e.PreviousStatus)
	}

	if !IsValidVerificationStatus(e.NewStatus) {
		return ErrInvalidVerificationStatus.Wrapf("invalid new status: %s", e.NewStatus)
	}

	if e.Timestamp.IsZero() {
		return ErrInvalidVerificationEvent.Wrap("timestamp cannot be zero")
	}

	return nil
}

// SimpleVerificationResult represents a simplified result of a verification process
// Used for internal workflow tracking; see VerificationResult in verification_result.go for full result type
type SimpleVerificationResult struct {
	// Success indicates if verification was successful
	Success bool `json:"success"`

	// Status is the resulting verification status
	Status VerificationStatus `json:"status"`

	// Score is the computed identity score (0-100)
	Score uint32 `json:"score"`

	// ScoreVersion is the ML model version used for scoring
	ScoreVersion string `json:"score_version"`

	// Confidence is the confidence level of the verification (0-100)
	Confidence uint32 `json:"confidence"`

	// Details contains human-readable verification details
	Details string `json:"details,omitempty"`

	// Flags contains any flags raised during verification
	Flags []string `json:"flags,omitempty"`

	// ProcessedAt is when the verification was processed
	ProcessedAt time.Time `json:"processed_at"`

	// ValidatorConsensus is the percentage of validators that agreed
	ValidatorConsensus uint32 `json:"validator_consensus,omitempty"`
}

// NewSimpleVerificationResult creates a new simplified verification result
func NewSimpleVerificationResult(success bool, status VerificationStatus, score uint32, scoreVersion string) *SimpleVerificationResult {
	return &SimpleVerificationResult{
		Success:      success,
		Status:       status,
		Score:        score,
		ScoreVersion: scoreVersion,
		ProcessedAt:  time.Now(),
		Flags:        make([]string, 0),
	}
}

// Validate validates the verification result
func (r *SimpleVerificationResult) Validate() error {
	if !IsValidVerificationStatus(r.Status) {
		return ErrInvalidVerificationStatus.Wrapf("invalid status: %s", r.Status)
	}

	if r.Score > 100 {
		return ErrInvalidScore.Wrap("score cannot exceed 100")
	}

	if r.Confidence > 100 {
		return ErrInvalidScore.Wrap("confidence cannot exceed 100")
	}

	if r.ScoreVersion == "" {
		return ErrInvalidScore.Wrap("score version cannot be empty")
	}

	return nil
}
