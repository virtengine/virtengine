package types

// Event types for the veid module
const (
	EventTypeScopeUploaded         = "scope_uploaded"
	EventTypeScopeRevoked          = "scope_revoked"
	EventTypeScopeVerified         = "scope_verified"
	EventTypeScopeRejected         = "scope_rejected"
	EventTypeStatusUpdated         = "status_updated"
	EventTypeScoreUpdated          = "score_updated"
	EventTypeIdentityCreated       = "identity_created"
	EventTypeIdentityLocked        = "identity_locked"
	EventTypeIdentityUnlocked      = "identity_unlocked"
	EventTypeVerificationRequested = "verification_requested"
	EventTypeVerificationMetrics   = "verification_metrics"
	EventTypeConsensusVerification = "consensus_verification"
	EventTypeVoteExtension         = "vote_extension"
)

// Event attribute keys
const (
	AttributeKeyAccountAddress    = "account_address"
	AttributeKeyScopeID           = "scope_id"
	AttributeKeyScopeType         = "scope_type"
	AttributeKeyVersion           = "version"
	AttributeKeyClientID          = "client_id"
	AttributeKeyPayloadHash       = "payload_hash"
	AttributeKeyStatus            = "status"
	AttributeKeyPreviousStatus    = "previous_status"
	AttributeKeyNewStatus         = "new_status"
	AttributeKeyScore             = "score"
	AttributeKeyScoreVersion      = "score_version"
	AttributeKeyTier              = "tier"
	AttributeKeyTimestamp         = "timestamp"
	AttributeKeyReason            = "reason"
	AttributeKeyValidatorAddress  = "validator_address"
	AttributeKeyDeviceFingerprint = "device_fingerprint"
	AttributeKeySaltHash          = "salt_hash"
	AttributeKeyRequestID         = "request_id"
	AttributeKeyMatch             = "match"
	AttributeKeyModelVersion      = "model_version"
	AttributeKeyComputeTime       = "compute_time_ms"
	AttributeKeyBlockHeight       = "block_height"
	AttributeKeyScoreDifference   = "score_difference"
	AttributeKeyProposerScore     = "proposer_score"
	AttributeKeyComputedScore     = "computed_score"
	AttributeKeyInputHashMatch    = "input_hash_match"
)

// EventScopeUploaded is emitted when a new identity scope is uploaded
type EventScopeUploaded struct {
	AccountAddress    string `json:"account_address"`
	ScopeID           string `json:"scope_id"`
	ScopeType         string `json:"scope_type"`
	Version           uint32 `json:"version"`
	ClientID          string `json:"client_id"`
	PayloadHash       string `json:"payload_hash"`
	DeviceFingerprint string `json:"device_fingerprint"`
	UploadedAt        int64  `json:"uploaded_at"`
}

// EventScopeRevoked is emitted when an identity scope is revoked
type EventScopeRevoked struct {
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
	ScopeType      string `json:"scope_type"`
	Reason         string `json:"reason,omitempty"`
	RevokedAt      int64  `json:"revoked_at"`
}

// EventScopeVerified is emitted when an identity scope is verified
type EventScopeVerified struct {
	AccountAddress   string `json:"account_address"`
	ScopeID          string `json:"scope_id"`
	ScopeType        string `json:"scope_type"`
	ValidatorAddress string `json:"validator_address,omitempty"`
	Score            uint32 `json:"score,omitempty"`
	VerifiedAt       int64  `json:"verified_at"`
}

// EventScopeRejected is emitted when an identity scope is rejected
type EventScopeRejected struct {
	AccountAddress   string `json:"account_address"`
	ScopeID          string `json:"scope_id"`
	ScopeType        string `json:"scope_type"`
	Reason           string `json:"reason"`
	ValidatorAddress string `json:"validator_address,omitempty"`
	RejectedAt       int64  `json:"rejected_at"`
}

// EventStatusUpdated is emitted when a scope's verification status is updated
type EventStatusUpdated struct {
	AccountAddress   string `json:"account_address"`
	ScopeID          string `json:"scope_id"`
	PreviousStatus   string `json:"previous_status"`
	NewStatus        string `json:"new_status"`
	Reason           string `json:"reason,omitempty"`
	ValidatorAddress string `json:"validator_address,omitempty"`
	UpdatedAt        int64  `json:"updated_at"`
}

// EventScoreUpdated is emitted when an identity's score is updated
type EventScoreUpdated struct {
	AccountAddress string `json:"account_address"`
	PreviousScore  uint32 `json:"previous_score"`
	NewScore       uint32 `json:"new_score"`
	ScoreVersion   string `json:"score_version"`
	PreviousTier   string `json:"previous_tier"`
	NewTier        string `json:"new_tier"`
	UpdatedAt      int64  `json:"updated_at"`
}

// EventIdentityCreated is emitted when a new identity record is created
type EventIdentityCreated struct {
	AccountAddress string `json:"account_address"`
	CreatedAt      int64  `json:"created_at"`
}

// EventIdentityLocked is emitted when an identity is locked
type EventIdentityLocked struct {
	AccountAddress string `json:"account_address"`
	Reason         string `json:"reason"`
	LockedAt       int64  `json:"locked_at"`
}

// EventIdentityUnlocked is emitted when an identity is unlocked
type EventIdentityUnlocked struct {
	AccountAddress string `json:"account_address"`
	UnlockedAt     int64  `json:"unlocked_at"`
}

// EventVerificationRequested is emitted when verification is requested for a scope
type EventVerificationRequested struct {
	AccountAddress string `json:"account_address"`
	ScopeID        string `json:"scope_id"`
	ScopeType      string `json:"scope_type"`
	RequestedAt    int64  `json:"requested_at"`
}

// EventVerificationCompleted is emitted when a verification request is completed
type EventVerificationCompleted struct {
	RequestID      string `json:"request_id"`
	AccountAddress string `json:"account_address"`
	Score          uint32 `json:"score"`
	Status         string `json:"status"`
	BlockHeight    int64  `json:"block_height"`
}

// EventVerificationFailed is emitted when a verification request fails
type EventVerificationFailed struct {
	RequestID      string   `json:"request_id"`
	AccountAddress string   `json:"account_address"`
	ReasonCodes    []string `json:"reason_codes"`
	BlockHeight    int64    `json:"block_height"`
}
