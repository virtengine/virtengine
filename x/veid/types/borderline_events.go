package types

// Borderline fallback event types
const (
	// EventTypeBorderlineFallbackTriggered is emitted when borderline fallback is triggered
	EventTypeBorderlineFallbackTriggered = "borderline_fallback_triggered"

	// EventTypeBorderlineFallbackCompleted is emitted when borderline fallback is completed
	EventTypeBorderlineFallbackCompleted = "borderline_fallback_completed"

	// EventTypeBorderlineFallbackFailed is emitted when borderline fallback fails
	EventTypeBorderlineFallbackFailed = "borderline_fallback_failed"

	// EventTypeBorderlineFallbackExpired is emitted when borderline fallback expires
	EventTypeBorderlineFallbackExpired = "borderline_fallback_expired"
)

// Borderline fallback event attribute keys
const (
	AttributeKeyFallbackID        = "fallback_id"
	AttributeKeyBorderlineScore   = "borderline_score"
	AttributeKeyChallengeID       = "challenge_id"
	AttributeKeyRequiredFactors   = "required_factors"
	AttributeKeySatisfiedFactors  = "satisfied_factors"
	AttributeKeyFinalStatus       = "final_status"
	AttributeKeyExpiresAt         = "expires_at"
	AttributeKeyFactorClass       = "factor_class"
)

// EventBorderlineFallbackTriggered represents the event emitted when borderline fallback is triggered
type EventBorderlineFallbackTriggered struct {
	// AccountAddress is the account triggering fallback
	AccountAddress string `json:"account_address"`

	// FallbackID is the unique identifier for this fallback attempt
	FallbackID string `json:"fallback_id"`

	// BorderlineScore is the facial verification score that triggered fallback
	BorderlineScore uint32 `json:"borderline_score"`

	// ChallengeID is the MFA challenge ID created for this fallback
	ChallengeID string `json:"challenge_id"`

	// RequiredFactors are the factor types that can satisfy this fallback
	RequiredFactors []string `json:"required_factors"`

	// ExpiresAt is when the fallback expires (Unix timestamp)
	ExpiresAt int64 `json:"expires_at"`

	// BlockHeight is the block height when fallback was triggered
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the event occurred
	Timestamp int64 `json:"timestamp"`
}

// EventBorderlineFallbackCompleted represents the event emitted when borderline fallback succeeds
type EventBorderlineFallbackCompleted struct {
	// AccountAddress is the account that completed fallback
	AccountAddress string `json:"account_address"`

	// FallbackID is the unique identifier for this fallback attempt
	FallbackID string `json:"fallback_id"`

	// ChallengeID is the MFA challenge ID that was satisfied
	ChallengeID string `json:"challenge_id"`

	// FactorsSatisfied are the factors that were successfully verified
	// This is auditable but does NOT contain secrets
	FactorsSatisfied []string `json:"factors_satisfied"`

	// FactorClass indicates the security level of satisfied factors (high/medium/low)
	FactorClass string `json:"factor_class"`

	// FinalStatus is the resulting verification status
	FinalStatus VerificationStatus `json:"final_status"`

	// BorderlineScore is the original borderline score
	BorderlineScore uint32 `json:"borderline_score"`

	// BlockHeight is the block height when fallback completed
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the event occurred
	Timestamp int64 `json:"timestamp"`
}

// EventBorderlineFallbackFailed represents the event emitted when borderline fallback fails
type EventBorderlineFallbackFailed struct {
	// AccountAddress is the account that failed fallback
	AccountAddress string `json:"account_address"`

	// FallbackID is the unique identifier for this fallback attempt
	FallbackID string `json:"fallback_id"`

	// ChallengeID is the MFA challenge ID that failed
	ChallengeID string `json:"challenge_id"`

	// Reason describes why the fallback failed
	Reason string `json:"reason"`

	// AttemptCount is the number of attempts made
	AttemptCount uint32 `json:"attempt_count"`

	// BlockHeight is the block height when fallback failed
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the event occurred
	Timestamp int64 `json:"timestamp"`
}

// EventBorderlineFallbackExpired represents the event emitted when borderline fallback expires
type EventBorderlineFallbackExpired struct {
	// AccountAddress is the account whose fallback expired
	AccountAddress string `json:"account_address"`

	// FallbackID is the unique identifier for this fallback attempt
	FallbackID string `json:"fallback_id"`

	// ChallengeID is the MFA challenge ID that expired
	ChallengeID string `json:"challenge_id"`

	// BorderlineScore is the original borderline score
	BorderlineScore uint32 `json:"borderline_score"`

	// CreatedAt is when the fallback was created
	CreatedAt int64 `json:"created_at"`

	// ExpiredAt is when the fallback expired
	ExpiredAt int64 `json:"expired_at"`

	// BlockHeight is the block height when expiry was recorded
	BlockHeight int64 `json:"block_height"`
}
