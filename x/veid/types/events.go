package types

import "fmt"

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

	// Spec-defined VEID events (per veid-flow-spec.md)
	EventTypeVerificationSubmitted = "veid_verification_submitted"
	EventTypeVerificationCompleted = "veid_verification_completed"
	EventTypeTierChanged           = "veid_tier_changed"
	EventTypeAuthorizationGranted  = "veid_authorization_granted"
	EventTypeAuthorizationConsumed = "veid_authorization_consumed"
	EventTypeAuthorizationExpired  = "veid_authorization_expired"
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

	// Spec-defined attribute keys (per veid-flow-spec.md)
	AttributeKeyAccount     = "account"
	AttributeKeyOldTier     = "old_tier"
	AttributeKeyNewTier     = "new_tier"
	AttributeKeySessionID   = "session_id"
	AttributeKeyAction      = "action"
	AttributeKeyFactorsUsed = "factors_used"
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

// EventVerificationSubmitted is emitted when a verification is submitted (spec-defined)
type EventVerificationSubmitted struct {
	Account     string `json:"account"`
	ScopeID     string `json:"scope_id"`
	ScopeType   string `json:"scope_type"`
	RequestID   string `json:"request_id"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTierChanged is emitted when an account's tier changes (spec-defined)
type EventTierChanged struct {
	Account     string `json:"account"`
	OldTier     string `json:"old_tier"`
	NewTier     string `json:"new_tier"`
	Score       uint32 `json:"score"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventAuthorizationGranted is emitted when MFA authorization is granted (spec-defined)
type EventAuthorizationGranted struct {
	Account     string   `json:"account"`
	SessionID   string   `json:"session_id"`
	Action      string   `json:"action"`
	FactorsUsed []string `json:"factors_used"`
	ExpiresAt   int64    `json:"expires_at"`
	BlockHeight int64    `json:"block_height"`
	Timestamp   int64    `json:"timestamp"`
}

// EventAuthorizationConsumed is emitted when an MFA authorization is consumed (spec-defined)
type EventAuthorizationConsumed struct {
	Account     string `json:"account"`
	SessionID   string `json:"session_id"`
	Action      string `json:"action"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventAuthorizationExpired is emitted when an MFA authorization expires (spec-defined)
type EventAuthorizationExpired struct {
	Account     string `json:"account"`
	SessionID   string `json:"session_id"`
	Action      string `json:"action"`
	CreatedAt   int64  `json:"created_at"`
	ExpiredAt   int64  `json:"expired_at"`
	BlockHeight int64  `json:"block_height"`
}

// ============================================================================
// Proto.Message interface stubs for Event types
// ============================================================================

// EventScopeUploaded proto stubs
func (*EventScopeUploaded) ProtoMessage()    {}
func (m *EventScopeUploaded) Reset()         { *m = EventScopeUploaded{} }
func (m *EventScopeUploaded) String() string { return fmt.Sprintf("%+v", *m) }

// EventScopeRevoked proto stubs
func (*EventScopeRevoked) ProtoMessage()    {}
func (m *EventScopeRevoked) Reset()         { *m = EventScopeRevoked{} }
func (m *EventScopeRevoked) String() string { return fmt.Sprintf("%+v", *m) }

// EventScopeVerified proto stubs
func (*EventScopeVerified) ProtoMessage()    {}
func (m *EventScopeVerified) Reset()         { *m = EventScopeVerified{} }
func (m *EventScopeVerified) String() string { return fmt.Sprintf("%+v", *m) }

// EventScopeRejected proto stubs
func (*EventScopeRejected) ProtoMessage()    {}
func (m *EventScopeRejected) Reset()         { *m = EventScopeRejected{} }
func (m *EventScopeRejected) String() string { return fmt.Sprintf("%+v", *m) }

// EventStatusUpdated proto stubs
func (*EventStatusUpdated) ProtoMessage()    {}
func (m *EventStatusUpdated) Reset()         { *m = EventStatusUpdated{} }
func (m *EventStatusUpdated) String() string { return fmt.Sprintf("%+v", *m) }

// EventScoreUpdated proto stubs
func (*EventScoreUpdated) ProtoMessage()    {}
func (m *EventScoreUpdated) Reset()         { *m = EventScoreUpdated{} }
func (m *EventScoreUpdated) String() string { return fmt.Sprintf("%+v", *m) }

// EventIdentityCreated proto stubs
func (*EventIdentityCreated) ProtoMessage()    {}
func (m *EventIdentityCreated) Reset()         { *m = EventIdentityCreated{} }
func (m *EventIdentityCreated) String() string { return fmt.Sprintf("%+v", *m) }

// EventIdentityLocked proto stubs
func (*EventIdentityLocked) ProtoMessage()    {}
func (m *EventIdentityLocked) Reset()         { *m = EventIdentityLocked{} }
func (m *EventIdentityLocked) String() string { return fmt.Sprintf("%+v", *m) }

// EventIdentityUnlocked proto stubs
func (*EventIdentityUnlocked) ProtoMessage()    {}
func (m *EventIdentityUnlocked) Reset()         { *m = EventIdentityUnlocked{} }
func (m *EventIdentityUnlocked) String() string { return fmt.Sprintf("%+v", *m) }

// EventVerificationRequested proto stubs
func (*EventVerificationRequested) ProtoMessage()    {}
func (m *EventVerificationRequested) Reset()         { *m = EventVerificationRequested{} }
func (m *EventVerificationRequested) String() string { return fmt.Sprintf("%+v", *m) }

// EventVerificationCompleted proto stubs
func (*EventVerificationCompleted) ProtoMessage()    {}
func (m *EventVerificationCompleted) Reset()         { *m = EventVerificationCompleted{} }
func (m *EventVerificationCompleted) String() string { return fmt.Sprintf("%+v", *m) }

// EventVerificationFailed proto stubs
func (*EventVerificationFailed) ProtoMessage()    {}
func (m *EventVerificationFailed) Reset()         { *m = EventVerificationFailed{} }
func (m *EventVerificationFailed) String() string { return fmt.Sprintf("%+v", *m) }

// EventVerificationSubmitted proto stubs
func (*EventVerificationSubmitted) ProtoMessage()    {}
func (m *EventVerificationSubmitted) Reset()         { *m = EventVerificationSubmitted{} }
func (m *EventVerificationSubmitted) String() string { return fmt.Sprintf("%+v", *m) }

// EventTierChanged proto stubs
func (*EventTierChanged) ProtoMessage()    {}
func (m *EventTierChanged) Reset()         { *m = EventTierChanged{} }
func (m *EventTierChanged) String() string { return fmt.Sprintf("%+v", *m) }

// EventAuthorizationGranted proto stubs
func (*EventAuthorizationGranted) ProtoMessage()    {}
func (m *EventAuthorizationGranted) Reset()         { *m = EventAuthorizationGranted{} }
func (m *EventAuthorizationGranted) String() string { return fmt.Sprintf("%+v", *m) }

// EventAuthorizationConsumed proto stubs
func (*EventAuthorizationConsumed) ProtoMessage()    {}
func (m *EventAuthorizationConsumed) Reset()         { *m = EventAuthorizationConsumed{} }
func (m *EventAuthorizationConsumed) String() string { return fmt.Sprintf("%+v", *m) }

// EventAuthorizationExpired proto stubs
func (*EventAuthorizationExpired) ProtoMessage()    {}
func (m *EventAuthorizationExpired) Reset()         { *m = EventAuthorizationExpired{} }
func (m *EventAuthorizationExpired) String() string { return fmt.Sprintf("%+v", *m) }

// ============================================================================
// Verification Attestation Events (VE-1B)
// ============================================================================

// Attestation event types
const (
	EventTypeAttestationCreated  = "attestation_created"
	EventTypeAttestationRevoked  = "attestation_revoked"
	EventTypeAttestationExpired  = "attestation_expired"
	EventTypeSignerKeyRegistered = "signer_key_registered"
	EventTypeSignerKeyActivated  = "signer_key_activated"
	EventTypeSignerKeyRevoked    = "signer_key_revoked"
	EventTypeSignerKeyRotated    = "signer_key_rotated"
	EventTypeNonceUsed           = "nonce_used"
	EventTypeAttestationVerified = "attestation_verified"
)

// Attestation event attribute keys
const (
	AttributeKeyAttestationID     = "attestation_id"
	AttributeKeyAttestationType   = "attestation_type"
	AttributeKeyIssuerID          = "issuer_id"
	AttributeKeyIssuerFingerprint = "issuer_fingerprint"
	AttributeKeySubjectAddress    = "subject_address"
	AttributeKeyNonceHash         = "nonce_hash"
	// Note: AttributeKeyExpiresAt already defined in borderline_events.go
	AttributeKeyKeyID            = "key_id"
	AttributeKeyKeyFingerprint   = "key_fingerprint"
	AttributeKeyKeyState         = "key_state"
	AttributeKeySignerID         = "signer_id"
	AttributeKeyRotationID       = "rotation_id"
	AttributeKeyOldKeyID         = "old_key_id"
	AttributeKeyNewKeyID         = "new_key_id"
	AttributeKeyRevocationReason = "revocation_reason"
	AttributeKeyConfidence       = "confidence"
)

// EventAttestationCreated is emitted when a verification attestation is created
type EventAttestationCreated struct {
	AttestationID   string `json:"attestation_id"`
	AttestationType string `json:"attestation_type"`
	IssuerID        string `json:"issuer_id"`
	IssuerFP        string `json:"issuer_fingerprint"`
	SubjectAddress  string `json:"subject_address"`
	Score           uint32 `json:"score"`
	Confidence      uint32 `json:"confidence"`
	ExpiresAt       int64  `json:"expires_at"`
	BlockHeight     int64  `json:"block_height"`
	Timestamp       int64  `json:"timestamp"`
}

// EventAttestationRevoked is emitted when an attestation is revoked
type EventAttestationRevoked struct {
	AttestationID  string `json:"attestation_id"`
	IssuerID       string `json:"issuer_id"`
	SubjectAddress string `json:"subject_address"`
	Reason         string `json:"reason"`
	RevokedBy      string `json:"revoked_by"`
	BlockHeight    int64  `json:"block_height"`
	Timestamp      int64  `json:"timestamp"`
}

// EventAttestationExpired is emitted when an attestation expires
type EventAttestationExpired struct {
	AttestationID  string `json:"attestation_id"`
	IssuerID       string `json:"issuer_id"`
	SubjectAddress string `json:"subject_address"`
	CreatedAt      int64  `json:"created_at"`
	ExpiredAt      int64  `json:"expired_at"`
	BlockHeight    int64  `json:"block_height"`
}

// EventSignerKeyRegistered is emitted when a new signer key is registered
type EventSignerKeyRegistered struct {
	KeyID       string `json:"key_id"`
	SignerID    string `json:"signer_id"`
	Fingerprint string `json:"fingerprint"`
	Algorithm   string `json:"algorithm"`
	State       string `json:"state"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSignerKeyActivated is emitted when a signer key becomes active
type EventSignerKeyActivated struct {
	KeyID       string `json:"key_id"`
	SignerID    string `json:"signer_id"`
	Fingerprint string `json:"fingerprint"`
	ExpiresAt   int64  `json:"expires_at"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSignerKeyRevoked is emitted when a signer key is revoked
type EventSignerKeyRevoked struct {
	KeyID       string `json:"key_id"`
	SignerID    string `json:"signer_id"`
	Fingerprint string `json:"fingerprint"`
	Reason      string `json:"reason"`
	RevokedBy   string `json:"revoked_by"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSignerKeyRotated is emitted when a signer key rotation completes
type EventSignerKeyRotated struct {
	RotationID        string `json:"rotation_id"`
	SignerID          string `json:"signer_id"`
	OldKeyID          string `json:"old_key_id"`
	OldKeyFingerprint string `json:"old_key_fingerprint"`
	NewKeyID          string `json:"new_key_id"`
	NewKeyFingerprint string `json:"new_key_fingerprint"`
	Reason            string `json:"reason"`
	BlockHeight       int64  `json:"block_height"`
	Timestamp         int64  `json:"timestamp"`
}

// EventNonceUsed is emitted when an attestation nonce is consumed
type EventNonceUsed struct {
	NonceHash       string `json:"nonce_hash"`
	IssuerFP        string `json:"issuer_fingerprint"`
	AttestationID   string `json:"attestation_id"`
	AttestationType string `json:"attestation_type"`
	BlockHeight     int64  `json:"block_height"`
	Timestamp       int64  `json:"timestamp"`
}

// EventAttestationVerified is emitted when an attestation signature is verified
type EventAttestationVerified struct {
	AttestationID  string `json:"attestation_id"`
	IssuerID       string `json:"issuer_id"`
	SubjectAddress string `json:"subject_address"`
	KeyID          string `json:"key_id"`
	Valid          bool   `json:"valid"`
	BlockHeight    int64  `json:"block_height"`
	Timestamp      int64  `json:"timestamp"`
}

// Proto stubs for attestation events

func (*EventAttestationCreated) ProtoMessage()    {}
func (m *EventAttestationCreated) Reset()         { *m = EventAttestationCreated{} }
func (m *EventAttestationCreated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventAttestationRevoked) ProtoMessage()    {}
func (m *EventAttestationRevoked) Reset()         { *m = EventAttestationRevoked{} }
func (m *EventAttestationRevoked) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventAttestationExpired) ProtoMessage()    {}
func (m *EventAttestationExpired) Reset()         { *m = EventAttestationExpired{} }
func (m *EventAttestationExpired) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSignerKeyRegistered) ProtoMessage()    {}
func (m *EventSignerKeyRegistered) Reset()         { *m = EventSignerKeyRegistered{} }
func (m *EventSignerKeyRegistered) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSignerKeyActivated) ProtoMessage()    {}
func (m *EventSignerKeyActivated) Reset()         { *m = EventSignerKeyActivated{} }
func (m *EventSignerKeyActivated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSignerKeyRevoked) ProtoMessage()    {}
func (m *EventSignerKeyRevoked) Reset()         { *m = EventSignerKeyRevoked{} }
func (m *EventSignerKeyRevoked) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSignerKeyRotated) ProtoMessage()    {}
func (m *EventSignerKeyRotated) Reset()         { *m = EventSignerKeyRotated{} }
func (m *EventSignerKeyRotated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventNonceUsed) ProtoMessage()    {}
func (m *EventNonceUsed) Reset()         { *m = EventNonceUsed{} }
func (m *EventNonceUsed) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventAttestationVerified) ProtoMessage()    {}
func (m *EventAttestationVerified) Reset()         { *m = EventAttestationVerified{} }
func (m *EventAttestationVerified) String() string { return fmt.Sprintf("%+v", *m) }

// ============================================================================
// SSO/OIDC Linkage Events (VE-4B)
// ============================================================================

// SSO linkage event types
const (
	EventTypeSSOLinkageCreated     = "sso_linkage_created"
	EventTypeSSOLinkageRevoked     = "sso_linkage_revoked"
	EventTypeSSOLinkageUpdated     = "sso_linkage_updated"
	EventTypeSSOLinkageExpired     = "sso_linkage_expired"
	EventTypeSSOChallengeCreated   = "sso_challenge_created"
	EventTypeSSOChallengeCompleted = "sso_challenge_completed"
	EventTypeSSOChallengeFailed    = "sso_challenge_failed"
	EventTypeSSONonceUsed          = "sso_nonce_used"
)

// SSO linkage event attribute keys
const (
	AttributeKeyLinkageID      = "linkage_id"
	AttributeKeyProvider       = "provider"
	AttributeKeyIssuer         = "issuer"
	AttributeKeySSOChallengeID = "sso_challenge_id"
	AttributeKeySubjectHash    = "subject_hash"
	AttributeKeyEmailDomain    = "email_domain"
	AttributeKeyTenantID       = "tenant_id"
	AttributeKeyFailureReason  = "failure_reason"
)

// EventSSOLinkageCreated is emitted when an SSO linkage is created
type EventSSOLinkageCreated struct {
	Account     string `json:"account"`
	LinkageID   string `json:"linkage_id"`
	Provider    string `json:"provider"`
	Issuer      string `json:"issuer"`
	SubjectHash string `json:"subject_hash"`
	EmailDomain string `json:"email_domain,omitempty"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSSOLinkageRevoked is emitted when an SSO linkage is revoked
type EventSSOLinkageRevoked struct {
	Account     string `json:"account"`
	LinkageID   string `json:"linkage_id"`
	Provider    string `json:"provider"`
	Reason      string `json:"reason"`
	RevokedBy   string `json:"revoked_by"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSSOLinkageUpdated is emitted when an SSO linkage is updated
type EventSSOLinkageUpdated struct {
	Account     string `json:"account"`
	LinkageID   string `json:"linkage_id"`
	Provider    string `json:"provider"`
	UpdateType  string `json:"update_type"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSSOLinkageExpired is emitted when an SSO linkage expires
type EventSSOLinkageExpired struct {
	Account     string `json:"account"`
	LinkageID   string `json:"linkage_id"`
	Provider    string `json:"provider"`
	CreatedAt   int64  `json:"created_at"`
	ExpiredAt   int64  `json:"expired_at"`
	BlockHeight int64  `json:"block_height"`
}

// EventSSOChallengeCreated is emitted when an SSO challenge is created
type EventSSOChallengeCreated struct {
	Account     string `json:"account"`
	ChallengeID string `json:"challenge_id"`
	Provider    string `json:"provider"`
	ExpiresAt   int64  `json:"expires_at"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSSOChallengeCompleted is emitted when an SSO challenge is completed
type EventSSOChallengeCompleted struct {
	Account     string `json:"account"`
	ChallengeID string `json:"challenge_id"`
	Provider    string `json:"provider"`
	LinkageID   string `json:"linkage_id"`
	Score       uint32 `json:"score"`
	Timestamp   int64  `json:"timestamp"`
}

// EventSSOChallengeFailed is emitted when an SSO challenge fails
type EventSSOChallengeFailed struct {
	Account       string `json:"account"`
	ChallengeID   string `json:"challenge_id"`
	Provider      string `json:"provider"`
	FailureReason string `json:"failure_reason"`
	Timestamp     int64  `json:"timestamp"`
}

// Proto stubs for SSO linkage events

func (*EventSSOLinkageCreated) ProtoMessage()    {}
func (m *EventSSOLinkageCreated) Reset()         { *m = EventSSOLinkageCreated{} }
func (m *EventSSOLinkageCreated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOLinkageRevoked) ProtoMessage()    {}
func (m *EventSSOLinkageRevoked) Reset()         { *m = EventSSOLinkageRevoked{} }
func (m *EventSSOLinkageRevoked) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOLinkageUpdated) ProtoMessage()    {}
func (m *EventSSOLinkageUpdated) Reset()         { *m = EventSSOLinkageUpdated{} }
func (m *EventSSOLinkageUpdated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOLinkageExpired) ProtoMessage()    {}
func (m *EventSSOLinkageExpired) Reset()         { *m = EventSSOLinkageExpired{} }
func (m *EventSSOLinkageExpired) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOChallengeCreated) ProtoMessage()    {}
func (m *EventSSOChallengeCreated) Reset()         { *m = EventSSOChallengeCreated{} }
func (m *EventSSOChallengeCreated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOChallengeCompleted) ProtoMessage()    {}
func (m *EventSSOChallengeCompleted) Reset()         { *m = EventSSOChallengeCompleted{} }
func (m *EventSSOChallengeCompleted) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventSSOChallengeFailed) ProtoMessage()    {}
func (m *EventSSOChallengeFailed) Reset()         { *m = EventSSOChallengeFailed{} }
func (m *EventSSOChallengeFailed) String() string { return fmt.Sprintf("%+v", *m) }
