package types

// Event types for the MFA module
const (
	// EventTypeFactorEnrolled is emitted when a factor is enrolled
	EventTypeFactorEnrolled = "factor_enrolled"
	// EventTypeFactorRevoked is emitted when a factor is revoked
	EventTypeFactorRevoked = "factor_revoked"
	// EventTypeFactorVerified is emitted when a factor is verified
	EventTypeFactorVerified = "factor_verified"
	// EventTypePolicyUpdated is emitted when an MFA policy is updated
	EventTypePolicyUpdated = "mfa_policy_updated"
	// EventTypeChallengeCreated is emitted when a challenge is created
	EventTypeChallengeCreated = "challenge_created"
	// EventTypeChallengeVerified is emitted when a challenge is verified
	EventTypeChallengeVerified = "challenge_verified"
	// EventTypeChallengeFailed is emitted when a challenge verification fails
	EventTypeChallengeFailed = "challenge_failed"
	// EventTypeChallengeExpired is emitted when a challenge expires
	EventTypeChallengeExpired = "challenge_expired"
	// EventTypeSessionCreated is emitted when an authorization session is created
	EventTypeSessionCreated = "session_created"
	// EventTypeSessionUsed is emitted when an authorization session is used
	EventTypeSessionUsed = "session_used"
	// EventTypeSessionExpired is emitted when an authorization session expires
	EventTypeSessionExpired = "session_expired"
	// EventTypeTrustedDeviceAdded is emitted when a trusted device is added
	EventTypeTrustedDeviceAdded = "trusted_device_added"
	// EventTypeTrustedDeviceRemoved is emitted when a trusted device is removed
	EventTypeTrustedDeviceRemoved = "trusted_device_removed"
	// EventTypeMFARequired is emitted when MFA is required for a transaction
	EventTypeMFARequired = "mfa_required"
	// EventTypeMFABypassed is emitted when MFA is bypassed (trusted device)
	EventTypeMFABypassed = "mfa_bypassed"
)

// Event attribute keys
const (
	AttributeKeyAccountAddress    = "account_address"
	AttributeKeyFactorType        = "factor_type"
	AttributeKeyFactorID          = "factor_id"
	AttributeKeyChallengeID       = "challenge_id"
	AttributeKeySessionID         = "session_id"
	AttributeKeyTransactionType   = "transaction_type"
	AttributeKeyStatus            = "status"
	AttributeKeyDeviceFingerprint = "device_fingerprint"
	AttributeKeyTimestamp         = "timestamp"
	AttributeKeyExpiresAt         = "expires_at"
	AttributeKeyReason            = "reason"
	AttributeKeyAttemptCount      = "attempt_count"
	AttributeKeyVerifiedFactors   = "verified_factors"
	AttributeKeyVEIDScore         = "veid_score"
	AttributeKeyThreshold         = "threshold"
)

// Event attribute values
const (
	AttributeValueSuccess = "success"
	AttributeValueFailure = "failure"
	AttributeValueExpired = "expired"
	AttributeValuePending = "pending"
)
