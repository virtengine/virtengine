package types

import (
	"time"
)

// ============================================================================
// Borderline Params
// ============================================================================

// DefaultBorderlineLowerThreshold is 85% (scaled 0-100)
const DefaultBorderlineLowerThreshold uint32 = 85

// DefaultBorderlineUpperThreshold is 90% (scaled 0-100)
const DefaultBorderlineUpperThreshold uint32 = 90

// DefaultBorderlineChallengeTimeoutSeconds is 5 minutes
const DefaultBorderlineChallengeTimeoutSeconds int64 = 300

// BorderlineParams defines the parameters for borderline identity verification
// When facial verification confidence falls in the borderline band, MFA fallback is triggered
type BorderlineParams struct {
	// LowerThreshold is the lower bound of the borderline band (e.g., 85 = 85% similarity)
	// Scores below this are rejected outright
	LowerThreshold uint32 `json:"lower_threshold"`

	// UpperThreshold is the upper bound of the borderline band (e.g., 90 = 90% similarity)
	// Scores at or above this are verified directly
	UpperThreshold uint32 `json:"upper_threshold"`

	// Enabled indicates whether borderline fallback is active
	// If false, scores below UpperThreshold are rejected
	Enabled bool `json:"enabled"`

	// RequiredFactors specifies which factor types can satisfy borderline verification
	// e.g., ["totp", "fido2", "email", "sms"]
	RequiredFactors []string `json:"required_factors"`

	// ChallengeTimeoutSeconds is the duration in seconds for MFA challenge validity
	ChallengeTimeoutSeconds int64 `json:"challenge_timeout_seconds"`

	// MinFactorsSatisfied is the minimum number of factors that must be satisfied
	// Defaults to 1 if not set
	MinFactorsSatisfied uint32 `json:"min_factors_satisfied"`
}

// DefaultBorderlineParams returns the default borderline parameters
func DefaultBorderlineParams() BorderlineParams {
	return BorderlineParams{
		LowerThreshold:          DefaultBorderlineLowerThreshold,
		UpperThreshold:          DefaultBorderlineUpperThreshold,
		Enabled:                 true,
		RequiredFactors:         []string{"totp", "fido2", "email", "sms"},
		ChallengeTimeoutSeconds: DefaultBorderlineChallengeTimeoutSeconds,
		MinFactorsSatisfied:     1,
	}
}

// Validate validates the borderline parameters
func (p BorderlineParams) Validate() error {
	if p.LowerThreshold > p.UpperThreshold {
		return ErrInvalidParams.Wrap("lower_threshold cannot exceed upper_threshold")
	}

	if p.UpperThreshold > MaxScore {
		return ErrInvalidParams.Wrapf("upper_threshold cannot exceed %d", MaxScore)
	}

	if p.LowerThreshold == 0 && p.UpperThreshold == 0 && p.Enabled {
		return ErrInvalidParams.Wrap("thresholds cannot both be zero when enabled")
	}

	if p.Enabled && len(p.RequiredFactors) == 0 {
		return ErrInvalidParams.Wrap("required_factors cannot be empty when enabled")
	}

	if p.ChallengeTimeoutSeconds <= 0 && p.Enabled {
		return ErrInvalidParams.Wrap("challenge_timeout_seconds must be positive when enabled")
	}

	return nil
}

// GetChallengeTimeout returns the challenge timeout as a duration
func (p BorderlineParams) GetChallengeTimeout() time.Duration {
	return time.Duration(p.ChallengeTimeoutSeconds) * time.Second
}

// IsScoreInBorderlineBand checks if a score falls within the borderline band
func (p BorderlineParams) IsScoreInBorderlineBand(score uint32) bool {
	return score >= p.LowerThreshold && score < p.UpperThreshold
}

// IsScoreAboveUpperThreshold checks if a score is at or above the upper threshold
func (p BorderlineParams) IsScoreAboveUpperThreshold(score uint32) bool {
	return score >= p.UpperThreshold
}

// IsScoreBelowLowerThreshold checks if a score is below the lower threshold
func (p BorderlineParams) IsScoreBelowLowerThreshold(score uint32) bool {
	return score < p.LowerThreshold
}

// ============================================================================
// Borderline Fallback Record
// ============================================================================

// BorderlineFallbackStatus represents the status of a borderline fallback attempt
type BorderlineFallbackStatus string

const (
	// BorderlineFallbackStatusPending indicates fallback is awaiting MFA completion
	BorderlineFallbackStatusPending BorderlineFallbackStatus = "pending"

	// BorderlineFallbackStatusCompleted indicates fallback was successfully completed
	BorderlineFallbackStatusCompleted BorderlineFallbackStatus = "completed"

	// BorderlineFallbackStatusFailed indicates fallback failed (MFA not satisfied)
	BorderlineFallbackStatusFailed BorderlineFallbackStatus = "failed"

	// BorderlineFallbackStatusExpired indicates fallback expired before completion
	BorderlineFallbackStatusExpired BorderlineFallbackStatus = "expired"

	// BorderlineFallbackStatusCancelled indicates fallback was cancelled
	BorderlineFallbackStatusCancelled BorderlineFallbackStatus = "cancelled"
)

// BorderlineFallbackRecord tracks a borderline fallback attempt
type BorderlineFallbackRecord struct {
	// FallbackID is the unique identifier for this fallback attempt
	FallbackID string `json:"fallback_id"`

	// AccountAddress is the account this fallback is for
	AccountAddress string `json:"account_address"`

	// BorderlineScore is the facial verification score that triggered fallback
	BorderlineScore uint32 `json:"borderline_score"`

	// ChallengeID is the MFA challenge ID created for this fallback
	ChallengeID string `json:"challenge_id"`

	// Status is the current status of the fallback
	Status BorderlineFallbackStatus `json:"status"`

	// RequiredFactors are the factor types that can satisfy this fallback
	RequiredFactors []string `json:"required_factors"`

	// SatisfiedFactors are the factors that were successfully verified
	SatisfiedFactors []string `json:"satisfied_factors,omitempty"`

	// CreatedAt is when the fallback was created (Unix timestamp)
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the fallback expires (Unix timestamp)
	ExpiresAt int64 `json:"expires_at"`

	// CompletedAt is when the fallback was completed (Unix timestamp)
	CompletedAt int64 `json:"completed_at,omitempty"`

	// BlockHeight is the block height when fallback was created
	BlockHeight int64 `json:"block_height"`

	// FinalVerificationStatus is the verification status after fallback completion
	FinalVerificationStatus VerificationStatus `json:"final_verification_status,omitempty"`
}

// NewBorderlineFallbackRecord creates a new borderline fallback record
func NewBorderlineFallbackRecord(
	fallbackID string,
	accountAddress string,
	borderlineScore uint32,
	challengeID string,
	requiredFactors []string,
	createdAt int64,
	expiresAt int64,
	blockHeight int64,
) *BorderlineFallbackRecord {
	return &BorderlineFallbackRecord{
		FallbackID:       fallbackID,
		AccountAddress:   accountAddress,
		BorderlineScore:  borderlineScore,
		ChallengeID:      challengeID,
		Status:           BorderlineFallbackStatusPending,
		RequiredFactors:  requiredFactors,
		SatisfiedFactors: make([]string, 0),
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		BlockHeight:      blockHeight,
	}
}

// Validate validates the borderline fallback record
func (r *BorderlineFallbackRecord) Validate() error {
	if r.FallbackID == "" {
		return ErrInvalidBorderlineFallback.Wrap("fallback_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}

	if r.BorderlineScore > MaxScore {
		return ErrInvalidScore.Wrapf("borderline_score cannot exceed %d", MaxScore)
	}

	if r.ChallengeID == "" {
		return ErrInvalidBorderlineFallback.Wrap("challenge_id cannot be empty")
	}

	if r.CreatedAt <= 0 {
		return ErrInvalidBorderlineFallback.Wrap("created_at must be positive")
	}

	if r.ExpiresAt <= r.CreatedAt {
		return ErrInvalidBorderlineFallback.Wrap("expires_at must be after created_at")
	}

	return nil
}

// IsExpired returns true if the fallback has expired
func (r *BorderlineFallbackRecord) IsExpired(now int64) bool {
	return now > r.ExpiresAt
}

// IsPending returns true if the fallback is still pending
func (r *BorderlineFallbackRecord) IsPending() bool {
	return r.Status == BorderlineFallbackStatusPending
}

// MarkCompleted marks the fallback as completed
func (r *BorderlineFallbackRecord) MarkCompleted(satisfiedFactors []string, finalStatus VerificationStatus, completedAt int64) {
	r.Status = BorderlineFallbackStatusCompleted
	r.SatisfiedFactors = satisfiedFactors
	r.FinalVerificationStatus = finalStatus
	r.CompletedAt = completedAt
}

// MarkFailed marks the fallback as failed
func (r *BorderlineFallbackRecord) MarkFailed(completedAt int64) {
	r.Status = BorderlineFallbackStatusFailed
	r.CompletedAt = completedAt
}

// MarkExpired marks the fallback as expired
func (r *BorderlineFallbackRecord) MarkExpired(expiredAt int64) {
	r.Status = BorderlineFallbackStatusExpired
	r.CompletedAt = expiredAt
}

// MarkCancelled marks the fallback as cancelled
func (r *BorderlineFallbackRecord) MarkCancelled(cancelledAt int64) {
	r.Status = BorderlineFallbackStatusCancelled
	r.CompletedAt = cancelledAt
}
