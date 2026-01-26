package types

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// ChallengeStatus represents the status of an MFA challenge
type ChallengeStatus uint8

const (
	// ChallengeStatusUnspecified represents an unspecified status
	ChallengeStatusUnspecified ChallengeStatus = 0
	// ChallengeStatusPending represents a challenge awaiting response
	ChallengeStatusPending ChallengeStatus = 1
	// ChallengeStatusVerified represents a successfully verified challenge
	ChallengeStatusVerified ChallengeStatus = 2
	// ChallengeStatusFailed represents a failed challenge
	ChallengeStatusFailed ChallengeStatus = 3
	// ChallengeStatusExpired represents an expired challenge
	ChallengeStatusExpired ChallengeStatus = 4
	// ChallengeStatusCancelled represents a cancelled challenge
	ChallengeStatusCancelled ChallengeStatus = 5
)

// String returns the string representation of a challenge status
func (s ChallengeStatus) String() string {
	switch s {
	case ChallengeStatusPending:
		return "pending"
	case ChallengeStatusVerified:
		return "verified"
	case ChallengeStatusFailed:
		return "failed"
	case ChallengeStatusExpired:
		return "expired"
	case ChallengeStatusCancelled:
		return "cancelled"
	default:
		return "unspecified"
	}
}

// Challenge represents an MFA challenge issued to a user
type Challenge struct {
	// ChallengeID is the unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account this challenge is for
	AccountAddress string `json:"account_address"`

	// FactorType is the type of factor this challenge is for
	FactorType FactorType `json:"factor_type"`

	// FactorID is the specific factor enrollment being challenged
	FactorID string `json:"factor_id"`

	// TransactionType is the sensitive transaction this challenge is for
	TransactionType SensitiveTransactionType `json:"transaction_type"`

	// Status is the current status of the challenge
	Status ChallengeStatus `json:"status"`

	// ChallengeData contains factor-specific challenge data
	// For FIDO2: contains the challenge bytes to sign
	// For OTP factors: contains hash of expected OTP (for verification tracking)
	// Never contains raw secrets
	ChallengeData []byte `json:"challenge_data,omitempty"`

	// CreatedAt is when the challenge was created
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the challenge expires
	ExpiresAt int64 `json:"expires_at"`

	// VerifiedAt is when the challenge was verified (if successful)
	VerifiedAt int64 `json:"verified_at,omitempty"`

	// AttemptCount tracks the number of verification attempts
	AttemptCount uint32 `json:"attempt_count"`

	// MaxAttempts is the maximum number of verification attempts allowed
	MaxAttempts uint32 `json:"max_attempts"`

	// Nonce is a random value for replay protection
	Nonce string `json:"nonce"`

	// SessionID links this challenge to an authorization session
	SessionID string `json:"session_id,omitempty"`

	// Metadata contains additional challenge-specific data
	Metadata *ChallengeMetadata `json:"metadata,omitempty"`
}

// ChallengeMetadata contains additional challenge information
type ChallengeMetadata struct {
	// FIDO2Challenge contains FIDO2-specific challenge data
	FIDO2Challenge *FIDO2ChallengeData `json:"fido2_challenge,omitempty"`

	// OTPInfo contains OTP-specific tracking data
	OTPInfo *OTPChallengeInfo `json:"otp_info,omitempty"`

	// ClientInfo contains information about the requesting client
	ClientInfo *ClientInfo `json:"client_info,omitempty"`

	// HardwareKeyChallenge contains hardware key/X.509/smart card challenge data (VE-925)
	HardwareKeyChallenge *HardwareKeyChallenge `json:"hardware_key_challenge,omitempty"`
}

// FIDO2ChallengeData contains FIDO2-specific challenge information
type FIDO2ChallengeData struct {
	// Challenge is the random challenge bytes
	Challenge []byte `json:"challenge"`

	// RelyingPartyID is the relying party identifier
	RelyingPartyID string `json:"relying_party_id"`

	// AllowedCredentials is the list of allowed credential IDs
	AllowedCredentials [][]byte `json:"allowed_credentials,omitempty"`

	// UserVerificationRequirement specifies if user verification is required
	UserVerificationRequirement string `json:"user_verification"`
}

// OTPChallengeInfo contains OTP tracking information
// NOTE: Never stores the actual OTP value
type OTPChallengeInfo struct {
	// DeliveryMethod indicates how the OTP was delivered
	DeliveryMethod string `json:"delivery_method"` // "sms", "email", "app"

	// DeliveryDestinationMasked is the masked delivery destination
	DeliveryDestinationMasked string `json:"delivery_destination_masked"`

	// SentAt is when the OTP was sent
	SentAt int64 `json:"sent_at"`

	// ResendCount tracks how many times the OTP was resent
	ResendCount uint32 `json:"resend_count"`

	// LastResendAt is when the OTP was last resent
	LastResendAt int64 `json:"last_resend_at,omitempty"`
}

// ClientInfo contains information about the requesting client
type ClientInfo struct {
	// DeviceFingerprint is a hash identifying the device
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// IPHash is a hash of the client IP
	IPHash string `json:"ip_hash,omitempty"`

	// UserAgent is the sanitized user agent string
	UserAgent string `json:"user_agent,omitempty"`

	// RequestedAt is when the request was made
	RequestedAt int64 `json:"requested_at"`
}

// NewChallenge creates a new challenge
func NewChallenge(
	accountAddress string,
	factorType FactorType,
	factorID string,
	txType SensitiveTransactionType,
	ttlSeconds int64,
	maxAttempts uint32,
) (*Challenge, error) {
	// Generate challenge ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, ErrChallengeCreationFailed.Wrapf("failed to generate challenge ID: %v", err)
	}

	// Generate nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, ErrChallengeCreationFailed.Wrapf("failed to generate nonce: %v", err)
	}

	now := time.Now().Unix()

	return &Challenge{
		ChallengeID:     hex.EncodeToString(idBytes),
		AccountAddress:  accountAddress,
		FactorType:      factorType,
		FactorID:        factorID,
		TransactionType: txType,
		Status:          ChallengeStatusPending,
		CreatedAt:       now,
		ExpiresAt:       now + ttlSeconds,
		AttemptCount:    0,
		MaxAttempts:     maxAttempts,
		Nonce:           hex.EncodeToString(nonceBytes),
	}, nil
}

// Validate validates the challenge
func (c *Challenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidChallenge.Wrap("challenge_id cannot be empty")
	}

	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}

	if !c.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", c.FactorType)
	}

	if c.CreatedAt == 0 {
		return ErrInvalidChallenge.Wrap("created_at cannot be zero")
	}

	if c.ExpiresAt <= c.CreatedAt {
		return ErrInvalidChallenge.Wrap("expires_at must be after created_at")
	}

	if c.MaxAttempts == 0 {
		return ErrInvalidChallenge.Wrap("max_attempts must be greater than zero")
	}

	if c.Nonce == "" {
		return ErrInvalidChallenge.Wrap("nonce cannot be empty")
	}

	return nil
}

// IsExpired returns true if the challenge has expired
func (c *Challenge) IsExpired(now time.Time) bool {
	return now.Unix() > c.ExpiresAt
}

// IsPending returns true if the challenge is still pending
func (c *Challenge) IsPending() bool {
	return c.Status == ChallengeStatusPending
}

// CanAttempt returns true if another attempt can be made
func (c *Challenge) CanAttempt(now time.Time) bool {
	return c.IsPending() && !c.IsExpired(now) && c.AttemptCount < c.MaxAttempts
}

// RecordAttempt increments the attempt counter
func (c *Challenge) RecordAttempt() {
	c.AttemptCount++
}

// MarkVerified marks the challenge as verified
func (c *Challenge) MarkVerified(timestamp int64) {
	c.Status = ChallengeStatusVerified
	c.VerifiedAt = timestamp
}

// MarkFailed marks the challenge as failed
func (c *Challenge) MarkFailed() {
	c.Status = ChallengeStatusFailed
}

// MarkExpired marks the challenge as expired
func (c *Challenge) MarkExpired() {
	c.Status = ChallengeStatusExpired
}

// MarkCancelled marks the challenge as cancelled
func (c *Challenge) MarkCancelled() {
	c.Status = ChallengeStatusCancelled
}

// ChallengeResponse represents a response to an MFA challenge
type ChallengeResponse struct {
	// ChallengeID is the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// FactorType is the type of factor used
	FactorType FactorType `json:"factor_type"`

	// ResponseData contains the verification data
	// For FIDO2: signature and authenticator data
	// For OTP: the OTP code (transmitted securely, not stored)
	ResponseData []byte `json:"response_data"`

	// ClientInfo contains information about the responding client
	ClientInfo *ClientInfo `json:"client_info,omitempty"`

	// Timestamp is when the response was created
	Timestamp int64 `json:"timestamp"`
}

// Validate validates the challenge response
func (r *ChallengeResponse) Validate() error {
	if r.ChallengeID == "" {
		return ErrInvalidChallenge.Wrap("challenge_id cannot be empty")
	}

	if !r.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", r.FactorType)
	}

	if len(r.ResponseData) == 0 {
		return ErrInvalidChallengeResponse.Wrap("response_data cannot be empty")
	}

	return nil
}

// AuthorizationSession represents a temporary elevated session after MFA verification
type AuthorizationSession struct {
	// SessionID is the unique identifier for this session
	SessionID string `json:"session_id"`

	// AccountAddress is the account this session belongs to
	AccountAddress string `json:"account_address"`

	// TransactionType is the type of transaction authorized
	TransactionType SensitiveTransactionType `json:"transaction_type"`

	// VerifiedFactors are the factors that were verified for this session
	VerifiedFactors []FactorType `json:"verified_factors"`

	// CreatedAt is when the session was created
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is when the session expires
	ExpiresAt int64 `json:"expires_at"`

	// UsedAt is when the session was used (if single-use)
	UsedAt int64 `json:"used_at,omitempty"`

	// IsSingleUse indicates if the session can only be used once
	IsSingleUse bool `json:"is_single_use"`

	// DeviceFingerprint is the device this session is bound to
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

// IsValid returns true if the session is valid for use
func (s *AuthorizationSession) IsValid(now time.Time) bool {
	if now.Unix() > s.ExpiresAt {
		return false
	}

	if s.IsSingleUse && s.UsedAt > 0 {
		return false
	}

	return true
}

// MarkUsed marks the session as used
func (s *AuthorizationSession) MarkUsed(timestamp int64) {
	s.UsedAt = timestamp
}
