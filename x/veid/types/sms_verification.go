// Package types provides types for the VEID module.
//
// VE-910: SMS verification scope (decentralized)
// This file defines types for SMS-based phone verification with OTP,
// phone number hashing (never store plaintext), and validator-triggered SMS.
package types

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// SMSVerificationVersion is the current version of SMS verification format
const SMSVerificationVersion uint32 = 1

// Error message constants for SMS verification
const (
	errMsgCreatedAtZero       = "created_at cannot be zero"
	errMsgExpiresAtZero       = "expires_at cannot be zero"
	errMsgMaxAttemptsZero     = "max_attempts must be positive"
	errMsgOTPHashEmpty        = "otp_hash cannot be empty"
	errMsgChallengeIDEmpty    = "challenge_id cannot be empty"
	errMsgVerificationIDEmpty = "verification_id cannot be empty"
)

// OTP configuration constants
const (
	// DefaultOTPLength is the default length of OTP codes
	DefaultOTPLength = 6

	// DefaultOTPTTLSeconds is the default OTP time-to-live in seconds (5 minutes)
	DefaultOTPTTLSeconds int64 = 300

	// DefaultMaxOTPAttempts is the default maximum OTP verification attempts
	DefaultMaxOTPAttempts uint32 = 3

	// DefaultMaxResends is the default maximum OTP resends per session
	DefaultMaxResends uint32 = 3

	// DefaultPhoneHashSaltLength is the length of phone hash salt
	DefaultPhoneHashSaltLength = 32

	// DefaultVerificationExpiryDays is how long a verification remains valid
	DefaultVerificationExpiryDays = 365
)

// SMSVerificationStatus represents the status of SMS verification
type SMSVerificationStatus string

const (
	// SMSStatusPending indicates OTP has been sent, awaiting verification
	SMSStatusPending SMSVerificationStatus = "pending"

	// SMSStatusVerified indicates phone ownership is verified
	SMSStatusVerified SMSVerificationStatus = "verified"

	// SMSStatusFailed indicates verification failed (max attempts exceeded)
	SMSStatusFailed SMSVerificationStatus = "failed"

	// SMSStatusRevoked indicates verification was revoked
	SMSStatusRevoked SMSVerificationStatus = "revoked"

	// SMSStatusExpired indicates verification or OTP has expired
	SMSStatusExpired SMSVerificationStatus = "expired"

	// SMSStatusBlocked indicates phone number is blocked (VoIP/fraud detected)
	SMSStatusBlocked SMSVerificationStatus = "blocked"
)

// AllSMSVerificationStatuses returns all valid verification statuses
func AllSMSVerificationStatuses() []SMSVerificationStatus {
	return []SMSVerificationStatus{
		SMSStatusPending,
		SMSStatusVerified,
		SMSStatusFailed,
		SMSStatusRevoked,
		SMSStatusExpired,
		SMSStatusBlocked,
	}
}

// IsValidSMSVerificationStatus checks if a status is valid
func IsValidSMSVerificationStatus(s SMSVerificationStatus) bool {
	for _, valid := range AllSMSVerificationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// PhoneNumberHash represents a hashed phone number with salt
// CRITICAL: Never store or log plaintext phone numbers
type PhoneNumberHash struct {
	// Hash is the SHA-256 HMAC of the normalized phone number
	Hash string `json:"hash"`

	// Salt is the unique salt used for this hash
	Salt string `json:"salt"`

	// CountryCodeHash is a hash of just the country code (for analytics)
	CountryCodeHash string `json:"country_code_hash"`

	// CreatedAt is when this hash was created
	CreatedAt time.Time `json:"created_at"`
}

// NewPhoneNumberHash creates a hashed phone number
// The phone number should be normalized to E.164 format before hashing
// CRITICAL: The plaintext phoneNumber is immediately hashed and never stored
func NewPhoneNumberHash(phoneNumber string, countryCode string) (*PhoneNumberHash, error) {
	if phoneNumber == "" {
		return nil, ErrInvalidPhone.Wrap("phone number cannot be empty")
	}

	// Generate random salt
	saltBytes := make([]byte, DefaultPhoneHashSaltLength)
	if _, err := rand.Read(saltBytes); err != nil {
		return nil, ErrInvalidPhone.Wrapf("failed to generate salt: %v", err)
	}
	salt := hex.EncodeToString(saltBytes)

	// Create HMAC-SHA256 hash of the phone number
	h := hmac.New(sha256.New, saltBytes)
	h.Write([]byte(phoneNumber))
	phoneHash := hex.EncodeToString(h.Sum(nil))

	// Hash country code separately (for regional analytics)
	var countryCodeHash string
	if countryCode != "" {
		ccHash := sha256.Sum256([]byte(countryCode))
		countryCodeHash = hex.EncodeToString(ccHash[:])
	}

	return &PhoneNumberHash{
		Hash:            phoneHash,
		Salt:            salt,
		CountryCodeHash: countryCodeHash,
		CreatedAt:       time.Now(),
	}, nil
}

// VerifyPhoneHash verifies if a phone number matches this hash
// CRITICAL: Only use for verification, never store the plaintext
func (p *PhoneNumberHash) VerifyPhoneHash(phoneNumber string) bool {
	if p.Salt == "" || p.Hash == "" {
		return false
	}

	saltBytes, err := hex.DecodeString(p.Salt)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, saltBytes)
	h.Write([]byte(phoneNumber))
	computedHash := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(p.Hash), []byte(computedHash))
}

// Validate validates the phone number hash
func (p *PhoneNumberHash) Validate() error {
	if p.Hash == "" {
		return ErrInvalidPhone.Wrap("hash cannot be empty")
	}
	if len(p.Hash) != 64 { // SHA-256 hex = 64 chars
		return ErrInvalidPhone.Wrap("hash must be a valid SHA-256 hex string")
	}
	if p.Salt == "" {
		return ErrInvalidPhone.Wrap("salt cannot be empty")
	}
	if len(p.Salt) != DefaultPhoneHashSaltLength*2 { // hex encoded
		return ErrInvalidPhone.Wrap("salt has invalid length")
	}
	if p.CreatedAt.IsZero() {
		return ErrInvalidPhone.Wrap(errMsgCreatedAtZero)
	}
	return nil
}

// SMSVerificationRecord represents on-chain SMS verification metadata
// Note: Actual phone numbers are never stored on-chain, only hashes
type SMSVerificationRecord struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// VerificationID is a unique identifier for this verification
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account that owns this phone verification
	AccountAddress string `json:"account_address"`

	// PhoneHash contains the hashed phone number (never plaintext)
	PhoneHash PhoneNumberHash `json:"phone_hash"`

	// Status is the current verification status
	Status SMSVerificationStatus `json:"status"`

	// VerifiedAt is when this phone was verified
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when this verification expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CreatedAt is when this record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// VerificationAttempts counts total verification attempts
	VerificationAttempts uint32 `json:"verification_attempts"`

	// IsVoIP indicates if VoIP detection flagged this number
	IsVoIP bool `json:"is_voip"`

	// CarrierType indicates the type of carrier (mobile, landline, voip)
	CarrierType string `json:"carrier_type,omitempty"`

	// ValidatorAddress is the validator that performed the verification
	ValidatorAddress string `json:"validator_address,omitempty"`

	// AccountSignature binds this verification to the account
	AccountSignature []byte `json:"account_signature"`
}

// NewSMSVerificationRecord creates a new SMS verification record
// CRITICAL: phoneNumber is hashed immediately and never stored
func NewSMSVerificationRecord(
	verificationID string,
	accountAddress string,
	phoneNumber string, // Will be hashed, not stored
	countryCode string,
	createdAt time.Time,
) (*SMSVerificationRecord, error) {
	phoneHash, err := NewPhoneNumberHash(phoneNumber, countryCode)
	if err != nil {
		return nil, err
	}

	return &SMSVerificationRecord{
		Version:        SMSVerificationVersion,
		VerificationID: verificationID,
		AccountAddress: accountAddress,
		PhoneHash:      *phoneHash,
		Status:         SMSStatusPending,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}, nil
}

// Validate validates the SMS verification record
func (r *SMSVerificationRecord) Validate() error {
	if r.Version == 0 || r.Version > SMSVerificationVersion {
		return ErrInvalidPhone.Wrapf("unsupported version: %d", r.Version)
	}

	if r.VerificationID == "" {
		return ErrInvalidPhone.Wrap("verification_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if err := r.PhoneHash.Validate(); err != nil {
		return err
	}

	if !IsValidSMSVerificationStatus(r.Status) {
		return ErrInvalidPhone.Wrapf("invalid status: %s", r.Status)
	}

	if r.CreatedAt.IsZero() {
		return ErrInvalidPhone.Wrap(errMsgCreatedAtZero)
	}

	return nil
}

// IsActive returns true if the verification is currently valid
func (r *SMSVerificationRecord) IsActive() bool {
	if r.Status != SMSStatusVerified {
		return false
	}
	if r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt) {
		return false
	}
	return true
}

// MarkVerified marks the record as verified
func (r *SMSVerificationRecord) MarkVerified(verifiedAt time.Time, expiresAt *time.Time, validatorAddr string) {
	r.Status = SMSStatusVerified
	r.VerifiedAt = &verifiedAt
	r.ExpiresAt = expiresAt
	r.UpdatedAt = verifiedAt
	r.ValidatorAddress = validatorAddr
}

// MarkBlocked marks the record as blocked (VoIP/fraud detected)
func (r *SMSVerificationRecord) MarkBlocked(blockedAt time.Time, reason string) {
	r.Status = SMSStatusBlocked
	r.UpdatedAt = blockedAt
}

// MarkFailed marks the record as failed
func (r *SMSVerificationRecord) MarkFailed(failedAt time.Time) {
	r.Status = SMSStatusFailed
	r.UpdatedAt = failedAt
}

// String returns a string representation (non-sensitive)
func (r *SMSVerificationRecord) String() string {
	return fmt.Sprintf("SMSVerification{ID: %s, Status: %s, IsVoIP: %t}",
		r.VerificationID, r.Status, r.IsVoIP)
}

// SMSOTPChallenge represents a pending SMS OTP challenge
// This is used for the actual OTP verification flow
type SMSOTPChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// VerificationID links to the parent verification record
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account requesting SMS verification
	AccountAddress string `json:"account_address"`

	// PhoneHashRef is a reference to the phone hash (not the full hash)
	PhoneHashRef string `json:"phone_hash_ref"`

	// OTPHash is a hash of the OTP (never store plaintext OTP)
	OTPHash string `json:"otp_hash"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status is the challenge status
	Status SMSVerificationStatus `json:"status"`

	// Attempts is the number of verification attempts
	Attempts uint32 `json:"attempts"`

	// MaxAttempts is the maximum allowed attempts
	MaxAttempts uint32 `json:"max_attempts"`

	// ResendCount tracks how many times OTP was resent
	ResendCount uint32 `json:"resend_count"`

	// MaxResends is the maximum allowed resends
	MaxResends uint32 `json:"max_resends"`

	// LastAttemptAt is when the last attempt was made
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastResendAt is when the OTP was last resent
	LastResendAt *time.Time `json:"last_resend_at,omitempty"`

	// ValidatorAddress is the validator assigned to send the SMS
	ValidatorAddress string `json:"validator_address"`

	// SMSSentAt is when the SMS was sent by the validator
	SMSSentAt *time.Time `json:"sms_sent_at,omitempty"`

	// IsConsumed indicates if the OTP has been consumed
	IsConsumed bool `json:"is_consumed"`

	// MaskedPhone is the masked phone number for display (e.g., +1***...1234)
	MaskedPhone string `json:"masked_phone"`
}

// GenerateOTP generates a cryptographically secure OTP code
// Returns the OTP (to send via SMS) and its hash (to store)
func GenerateOTP(length int) (otp string, otpHash string, err error) {
	if length < 4 || length > 10 {
		length = DefaultOTPLength
	}

	// Generate random bytes for OTP
	otpBytes := make([]byte, length)
	if _, err := rand.Read(otpBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Convert to numeric OTP
	digits := make([]byte, length)
	for i := range otpBytes {
		digits[i] = '0' + (otpBytes[i] % 10)
	}
	otp = string(digits)

	// Hash the OTP for storage
	hash := sha256.Sum256([]byte(otp))
	otpHash = hex.EncodeToString(hash[:])

	return otp, otpHash, nil
}

// HashOTP creates a SHA-256 hash of an OTP
func HashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

// SMSOTPChallengeConfig contains configuration for creating an SMS OTP challenge
type SMSOTPChallengeConfig struct {
	ChallengeID      string
	VerificationID   string
	AccountAddress   string
	PhoneHashRef     string
	OTPHash          string
	MaskedPhone      string
	ValidatorAddress string
	CreatedAt        time.Time
	TTLSeconds       int64
	MaxAttempts      uint32
	MaxResends       uint32
}

// NewSMSOTPChallenge creates a new SMS OTP challenge from config
func NewSMSOTPChallenge(cfg SMSOTPChallengeConfig) *SMSOTPChallenge {
	expiresAt := cfg.CreatedAt.Add(time.Duration(cfg.TTLSeconds) * time.Second)

	return &SMSOTPChallenge{
		ChallengeID:      cfg.ChallengeID,
		VerificationID:   cfg.VerificationID,
		AccountAddress:   cfg.AccountAddress,
		PhoneHashRef:     cfg.PhoneHashRef,
		OTPHash:          cfg.OTPHash,
		MaskedPhone:      cfg.MaskedPhone,
		ValidatorAddress: cfg.ValidatorAddress,
		CreatedAt:        cfg.CreatedAt,
		ExpiresAt:        expiresAt,
		Status:           SMSStatusPending,
		MaxAttempts:      cfg.MaxAttempts,
		MaxResends:       cfg.MaxResends,
	}
}

// Validate validates the SMS OTP challenge
func (c *SMSOTPChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidPhone.Wrap(errMsgChallengeIDEmpty)
	}
	if c.VerificationID == "" {
		return ErrInvalidPhone.Wrap(errMsgVerificationIDEmpty)
	}
	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if c.OTPHash == "" {
		return ErrInvalidPhone.Wrap(errMsgOTPHashEmpty)
	}
	if len(c.OTPHash) != 64 { // SHA-256 hex
		return ErrInvalidPhone.Wrap("otp_hash must be a valid SHA-256 hex string")
	}
	if c.CreatedAt.IsZero() {
		return ErrInvalidPhone.Wrap(errMsgCreatedAtZero)
	}
	if c.ExpiresAt.IsZero() {
		return ErrInvalidPhone.Wrap(errMsgExpiresAtZero)
	}
	if c.MaxAttempts == 0 {
		return ErrInvalidPhone.Wrap(errMsgMaxAttemptsZero)
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *SMSOTPChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// CanAttempt returns true if another verification attempt is allowed
func (c *SMSOTPChallenge) CanAttempt() bool {
	return !c.IsConsumed && c.Attempts < c.MaxAttempts
}

// CanResend returns true if OTP can be resent
func (c *SMSOTPChallenge) CanResend() bool {
	return !c.IsConsumed && c.ResendCount < c.MaxResends
}

// RecordAttempt records a verification attempt
func (c *SMSOTPChallenge) RecordAttempt(attemptTime time.Time, success bool) {
	c.Attempts++
	c.LastAttemptAt = &attemptTime
	if success {
		c.IsConsumed = true
		c.Status = SMSStatusVerified
	} else if c.Attempts >= c.MaxAttempts {
		c.Status = SMSStatusFailed
	}
}

// RecordResend records an OTP resend
func (c *SMSOTPChallenge) RecordResend(resendTime time.Time, newOTPHash string, newExpiresAt time.Time) error {
	if !c.CanResend() {
		return ErrSMSResendLimitExceeded
	}
	c.ResendCount++
	c.LastResendAt = &resendTime
	c.OTPHash = newOTPHash
	c.ExpiresAt = newExpiresAt
	c.Attempts = 0 // Reset attempts on resend
	return nil
}

// VerifyOTP checks if the provided OTP matches
func (c *SMSOTPChallenge) VerifyOTP(providedOTP string) bool {
	if c.IsConsumed {
		return false
	}
	providedHash := HashOTP(providedOTP)
	return providedHash == c.OTPHash
}

// MaskPhoneNumber masks a phone number for display
// Example: +14155551234 -> +1***...1234
func MaskPhoneNumber(phoneNumber string) string {
	if len(phoneNumber) < 6 {
		return "****"
	}

	// Keep first 2-3 chars (country code) and last 4 digits
	prefix := phoneNumber[:3]
	suffix := phoneNumber[len(phoneNumber)-4:]

	return fmt.Sprintf("%s***...%s", prefix, suffix)
}

// SMSScoringWeight defines the weight of SMS verification in VEID scoring
type SMSScoringWeight struct {
	// BaseWeight is the base score weight in basis points
	BaseWeight uint32 `json:"base_weight"`

	// MobileBonus is additional weight for mobile (non-VoIP) numbers
	MobileBonus uint32 `json:"mobile_bonus"`

	// VerificationAgeBonus is bonus per year of verification age
	VerificationAgeBonusPerYear uint32 `json:"verification_age_bonus_per_year"`

	// MaxAgeBonus is the maximum age bonus
	MaxAgeBonus uint32 `json:"max_age_bonus"`
}

// DefaultSMSScoringWeight returns the default SMS scoring weight
func DefaultSMSScoringWeight() SMSScoringWeight {
	return SMSScoringWeight{
		BaseWeight:                  300, // 3% weight (same as email)
		MobileBonus:                 200, // +2% for non-VoIP mobile number
		VerificationAgeBonusPerYear: 25,  // +0.25% per year
		MaxAgeBonus:                 100, // max +1% from age
	}
}

// CalculateSMSScore calculates the score contribution for an SMS verification
func CalculateSMSScore(record *SMSVerificationRecord, weight SMSScoringWeight, now time.Time) uint32 {
	if !record.IsActive() {
		return 0
	}

	score := weight.BaseWeight

	// Bonus for non-VoIP mobile numbers
	if !record.IsVoIP && record.CarrierType == "mobile" {
		score += weight.MobileBonus
	}

	// Age bonus
	if record.VerifiedAt != nil {
		ageYears := uint32(now.Sub(*record.VerifiedAt).Hours() / (24 * 365))
		ageBonus := ageYears * weight.VerificationAgeBonusPerYear
		if ageBonus > weight.MaxAgeBonus {
			ageBonus = weight.MaxAgeBonus
		}
		score += ageBonus
	}

	return score
}

// ValidatorSMSGateway represents a validator's SMS gateway configuration
// Validators are responsible for sending SMS messages in a decentralized manner
type ValidatorSMSGateway struct {
	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// GatewayType is the type of SMS gateway (twilio, aws_sns, custom, etc.)
	GatewayType string `json:"gateway_type"`

	// IsActive indicates if this gateway is currently active
	IsActive bool `json:"is_active"`

	// SupportedRegions is the list of country codes this gateway supports
	SupportedRegions []string `json:"supported_regions"`

	// LastHealthCheck is when the gateway was last health-checked
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`

	// SuccessRate is the delivery success rate (0-100)
	SuccessRate uint32 `json:"success_rate"`

	// CreatedAt is when this gateway was registered
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this gateway was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the validator SMS gateway configuration
func (g *ValidatorSMSGateway) Validate() error {
	if g.ValidatorAddress == "" {
		return ErrInvalidAddress.Wrap("validator address cannot be empty")
	}
	if g.GatewayType == "" {
		return ErrInvalidPhone.Wrap("gateway_type cannot be empty")
	}
	if g.CreatedAt.IsZero() {
		return ErrInvalidPhone.Wrap(errMsgCreatedAtZero)
	}
	return nil
}

// SMSDeliveryResult represents the result of an SMS delivery attempt
type SMSDeliveryResult struct {
	// ChallengeID is the challenge this delivery is for
	ChallengeID string `json:"challenge_id"`

	// ValidatorAddress is the validator that sent the SMS
	ValidatorAddress string `json:"validator_address"`

	// Success indicates if the SMS was delivered successfully
	Success bool `json:"success"`

	// ErrorCode is the error code if delivery failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if delivery failed (sanitized)
	ErrorMessage string `json:"error_message,omitempty"`

	// DeliveredAt is when the SMS was delivered
	DeliveredAt time.Time `json:"delivered_at"`

	// GatewayMessageID is the gateway's message ID for tracking
	GatewayMessageID string `json:"gateway_message_id,omitempty"`
}

// SMSVerificationParams contains configurable parameters for SMS verification
type SMSVerificationParams struct {
	// OTPLength is the length of OTP codes
	OTPLength int `json:"otp_length"`

	// OTPTTLSeconds is how long an OTP is valid
	OTPTTLSeconds int64 `json:"otp_ttl_seconds"`

	// MaxOTPAttempts is max verification attempts per OTP
	MaxOTPAttempts uint32 `json:"max_otp_attempts"`

	// MaxResends is max OTP resends per verification
	MaxResends uint32 `json:"max_resends"`

	// VerificationExpiryDays is how long a verification remains valid
	VerificationExpiryDays int `json:"verification_expiry_days"`

	// BlockVoIPNumbers indicates if VoIP numbers should be blocked
	BlockVoIPNumbers bool `json:"block_voip_numbers"`

	// RequireCarrierLookup indicates if carrier lookup is required
	RequireCarrierLookup bool `json:"require_carrier_lookup"`
}

// DefaultSMSVerificationParams returns the default SMS verification parameters
func DefaultSMSVerificationParams() SMSVerificationParams {
	return SMSVerificationParams{
		OTPLength:              DefaultOTPLength,
		OTPTTLSeconds:          DefaultOTPTTLSeconds,
		MaxOTPAttempts:         DefaultMaxOTPAttempts,
		MaxResends:             DefaultMaxResends,
		VerificationExpiryDays: DefaultVerificationExpiryDays,
		BlockVoIPNumbers:       true,
		RequireCarrierLookup:   true,
	}
}

// Validate validates the SMS verification parameters
func (p *SMSVerificationParams) Validate() error {
	if p.OTPLength < 4 || p.OTPLength > 10 {
		return ErrInvalidPhone.Wrap("otp_length must be between 4 and 10")
	}
	if p.OTPTTLSeconds < 60 || p.OTPTTLSeconds > 600 {
		return ErrInvalidPhone.Wrap("otp_ttl_seconds must be between 60 and 600")
	}
	if p.MaxOTPAttempts < 1 || p.MaxOTPAttempts > 10 {
		return ErrInvalidPhone.Wrap("max_otp_attempts must be between 1 and 10")
	}
	if p.MaxResends > 5 {
		return ErrInvalidPhone.Wrap("max_resends must not exceed 5")
	}
	if p.VerificationExpiryDays < 1 || p.VerificationExpiryDays > 730 {
		return ErrInvalidPhone.Wrap("verification_expiry_days must be between 1 and 730")
	}
	return nil
}
