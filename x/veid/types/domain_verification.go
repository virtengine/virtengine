// Package types provides types for the VEID module.
//
// VE-223: Domain verification scope v1
// This file defines types for domain ownership verification for providers/organizations.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// DomainVerificationMethod defines how domain ownership is verified
type DomainVerificationMethod string

const (
	// DomainVerifyDNSTXT verifies via DNS TXT record
	DomainVerifyDNSTXT DomainVerificationMethod = "dns_txt"

	// DomainVerifyHTTPWellKnown verifies via HTTP /.well-known/ endpoint
	DomainVerifyHTTPWellKnown DomainVerificationMethod = "http_well_known"

	// DomainVerifyEmailAdmin verifies via email to admin contacts
	DomainVerifyEmailAdmin DomainVerificationMethod = "email_admin"
)

// AllDomainVerificationMethods returns all valid verification methods
func AllDomainVerificationMethods() []DomainVerificationMethod {
	return []DomainVerificationMethod{
		DomainVerifyDNSTXT,
		DomainVerifyHTTPWellKnown,
		DomainVerifyEmailAdmin,
	}
}

// IsValidDomainVerificationMethod checks if a method is valid
func IsValidDomainVerificationMethod(m DomainVerificationMethod) bool {
	for _, valid := range AllDomainVerificationMethods() {
		if m == valid {
			return true
		}
	}
	return false
}

// DomainVerificationVersion is the current version of domain verification format
const DomainVerificationVersion uint32 = 1

// DomainVerificationStatus represents the status of domain verification
type DomainVerificationStatus string

const (
	// DomainStatusPending indicates verification is pending
	DomainStatusPending DomainVerificationStatus = "pending"

	// DomainStatusVerified indicates domain ownership is verified
	DomainStatusVerified DomainVerificationStatus = "verified"

	// DomainStatusFailed indicates verification failed
	DomainStatusFailed DomainVerificationStatus = "failed"

	// DomainStatusRevoked indicates verification was revoked
	DomainStatusRevoked DomainVerificationStatus = "revoked"

	// DomainStatusExpired indicates verification has expired
	DomainStatusExpired DomainVerificationStatus = "expired"
)

// AllDomainVerificationStatuses returns all valid verification statuses
func AllDomainVerificationStatuses() []DomainVerificationStatus {
	return []DomainVerificationStatus{
		DomainStatusPending,
		DomainStatusVerified,
		DomainStatusFailed,
		DomainStatusRevoked,
		DomainStatusExpired,
	}
}

// IsValidDomainVerificationStatus checks if a status is valid
func IsValidDomainVerificationStatus(s DomainVerificationStatus) bool {
	for _, valid := range AllDomainVerificationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// DomainVerificationRecord represents on-chain domain verification metadata
type DomainVerificationRecord struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// VerificationID is a unique identifier for this verification
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account that owns this domain verification
	AccountAddress string `json:"account_address"`

	// Domain is the verified domain (e.g., "example.com")
	Domain string `json:"domain"`

	// DomainHash is a SHA256 hash of the domain for indexing
	DomainHash string `json:"domain_hash"`

	// Method is the verification method used
	Method DomainVerificationMethod `json:"method"`

	// ChallengeToken is the verification token (can be stored for re-verification)
	ChallengeToken string `json:"challenge_token"`

	// Status is the current verification status
	Status DomainVerificationStatus `json:"status"`

	// VerifiedAt is when this domain was verified
	VerifiedAt *time.Time `json:"verified_at,omitempty"`

	// ExpiresAt is when this verification expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// LastCheckedAt is when the verification was last re-checked
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`

	// RevokedAt is when the verification was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the reason for revocation
	RevokedReason string `json:"revoked_reason,omitempty"`

	// CreatedAt is when this record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// OrganizationName is the organization name (optional, for display)
	OrganizationName string `json:"organization_name,omitempty"`

	// IsWildcard indicates if this covers all subdomains
	IsWildcard bool `json:"is_wildcard"`

	// AccountSignature is the signature binding this verification
	AccountSignature []byte `json:"account_signature"`
}

// NewDomainVerificationRecord creates a new domain verification record
func NewDomainVerificationRecord(
	verificationID string,
	accountAddress string,
	domain string,
	method DomainVerificationMethod,
	challengeToken string,
	createdAt time.Time,
) *DomainVerificationRecord {
	return &DomainVerificationRecord{
		Version:        DomainVerificationVersion,
		VerificationID: verificationID,
		AccountAddress: accountAddress,
		Domain:         domain,
		DomainHash:     HashDomain(domain),
		Method:         method,
		ChallengeToken: challengeToken,
		Status:         DomainStatusPending,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
		IsWildcard:     false,
	}
}

// HashDomain creates a SHA256 hash of a domain name
func HashDomain(domain string) string {
	// Normalize domain to lowercase
	normalized := []byte(domain)
	for i := range normalized {
		if normalized[i] >= 'A' && normalized[i] <= 'Z' {
			normalized[i] += 32
		}
	}
	hash := sha256.Sum256(normalized)
	return hex.EncodeToString(hash[:])
}

// Validate validates the domain verification record
func (r *DomainVerificationRecord) Validate() error {
	if r.Version == 0 || r.Version > DomainVerificationVersion {
		return ErrInvalidDomain.Wrapf("unsupported version: %d", r.Version)
	}

	if r.VerificationID == "" {
		return ErrInvalidDomain.Wrap("verification_id cannot be empty")
	}

	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if r.Domain == "" {
		return ErrInvalidDomain.Wrap("domain cannot be empty")
	}

	if r.DomainHash == "" {
		return ErrInvalidDomain.Wrap("domain_hash cannot be empty")
	}

	if !IsValidDomainVerificationMethod(r.Method) {
		return ErrInvalidDomain.Wrapf("invalid verification method: %s", r.Method)
	}

	if r.ChallengeToken == "" {
		return ErrInvalidDomain.Wrap("challenge_token cannot be empty")
	}

	if !IsValidDomainVerificationStatus(r.Status) {
		return ErrInvalidDomain.Wrapf("invalid status: %s", r.Status)
	}

	if r.CreatedAt.IsZero() {
		return ErrInvalidDomain.Wrap("created_at cannot be zero")
	}

	return nil
}

// IsActive returns true if the verification is currently valid
func (r *DomainVerificationRecord) IsActive() bool {
	if r.Status != DomainStatusVerified {
		return false
	}
	if r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt) {
		return false
	}
	return true
}

// MarkVerified marks the record as verified
func (r *DomainVerificationRecord) MarkVerified(verifiedAt time.Time, expiresAt *time.Time) {
	r.Status = DomainStatusVerified
	r.VerifiedAt = &verifiedAt
	r.LastCheckedAt = &verifiedAt
	r.ExpiresAt = expiresAt
	r.UpdatedAt = verifiedAt
}

// MarkRevoked marks the record as revoked
func (r *DomainVerificationRecord) MarkRevoked(revokedAt time.Time, reason string) {
	r.Status = DomainStatusRevoked
	r.RevokedAt = &revokedAt
	r.RevokedReason = reason
	r.UpdatedAt = revokedAt
}

// String returns a string representation
func (r *DomainVerificationRecord) String() string {
	return fmt.Sprintf("DomainVerification{ID: %s, Domain: %s, Status: %s}",
		r.VerificationID, r.Domain, r.Status)
}

// DomainVerificationChallenge represents a pending domain verification challenge
type DomainVerificationChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting domain verification
	AccountAddress string `json:"account_address"`

	// Domain is the domain to verify
	Domain string `json:"domain"`

	// Method is the verification method to use
	Method DomainVerificationMethod `json:"method"`

	// Token is the verification token to place in DNS/HTTP
	Token string `json:"token"`

	// ExpectedValue is the full expected record value
	// e.g., "virtengine-verification=<token>"
	ExpectedValue string `json:"expected_value"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status is the challenge status
	Status DomainVerificationStatus `json:"status"`

	// VerificationAttempts is the number of verification attempts made
	VerificationAttempts uint32 `json:"verification_attempts"`

	// LastAttemptAt is when the last verification attempt was made
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastAttemptError is the error from the last attempt (if any)
	LastAttemptError string `json:"last_attempt_error,omitempty"`
}

// NewDomainVerificationChallenge creates a new domain verification challenge
func NewDomainVerificationChallenge(
	challengeID string,
	accountAddress string,
	domain string,
	method DomainVerificationMethod,
	token string,
	createdAt time.Time,
	ttlSeconds int64,
) *DomainVerificationChallenge {
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	expectedValue := fmt.Sprintf("virtengine-verification=%s", token)

	return &DomainVerificationChallenge{
		ChallengeID:    challengeID,
		AccountAddress: accountAddress,
		Domain:         domain,
		Method:         method,
		Token:          token,
		ExpectedValue:  expectedValue,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         DomainStatusPending,
	}
}

// Validate validates the domain verification challenge
func (c *DomainVerificationChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidDomain.Wrap("challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if c.Domain == "" {
		return ErrInvalidDomain.Wrap("domain cannot be empty")
	}
	if !IsValidDomainVerificationMethod(c.Method) {
		return ErrInvalidDomain.Wrapf("invalid method: %s", c.Method)
	}
	if c.Token == "" {
		return ErrInvalidDomain.Wrap("token cannot be empty")
	}
	if c.CreatedAt.IsZero() {
		return ErrInvalidDomain.Wrap("created_at cannot be zero")
	}
	if c.ExpiresAt.IsZero() {
		return ErrInvalidDomain.Wrap("expires_at cannot be zero")
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *DomainVerificationChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// GetDNSRecordName returns the DNS record name to check
func (c *DomainVerificationChallenge) GetDNSRecordName() string {
	return fmt.Sprintf("_virtengine-verification.%s", c.Domain)
}

// GetHTTPWellKnownPath returns the HTTP well-known path to check
func (c *DomainVerificationChallenge) GetHTTPWellKnownPath() string {
	return fmt.Sprintf("/.well-known/virtengine-verification/%s", c.Token)
}

// DomainRevocationEvent represents a domain revocation audit event
type DomainRevocationEvent struct {
	// EventID is a unique identifier for this event
	EventID string `json:"event_id"`

	// VerificationID is the verification that was revoked
	VerificationID string `json:"verification_id"`

	// AccountAddress is the account that owned the verification
	AccountAddress string `json:"account_address"`

	// Domain is the domain that was revoked
	Domain string `json:"domain"`

	// RevokedBy is the address that initiated revocation
	RevokedBy string `json:"revoked_by"`

	// RevokedAt is when the revocation occurred
	RevokedAt time.Time `json:"revoked_at"`

	// Reason is the revocation reason
	Reason string `json:"reason"`

	// BlockHeight is the block height when this occurred
	BlockHeight int64 `json:"block_height"`
}

// NewDomainRevocationEvent creates a new domain revocation event
func NewDomainRevocationEvent(
	eventID string,
	verificationID string,
	accountAddress string,
	domain string,
	revokedBy string,
	reason string,
	blockHeight int64,
	revokedAt time.Time,
) *DomainRevocationEvent {
	return &DomainRevocationEvent{
		EventID:        eventID,
		VerificationID: verificationID,
		AccountAddress: accountAddress,
		Domain:         domain,
		RevokedBy:      revokedBy,
		RevokedAt:      revokedAt,
		Reason:         reason,
		BlockHeight:    blockHeight,
	}
}

// DomainScoringWeight defines the weight of domain verification in VEID scoring
type DomainScoringWeight struct {
	// BaseWeight is the base score weight in basis points
	BaseWeight uint32 `json:"base_weight"`

	// WildcardBonus is additional weight for wildcard verifications
	WildcardBonus uint32 `json:"wildcard_bonus"`

	// VerificationAgeBonus is bonus per year of verification age
	VerificationAgeBonusPerYear uint32 `json:"verification_age_bonus_per_year"`

	// MaxAgeBonus is the maximum age bonus
	MaxAgeBonus uint32 `json:"max_age_bonus"`
}

// DefaultDomainScoringWeight returns the default domain scoring weight
func DefaultDomainScoringWeight() DomainScoringWeight {
	return DomainScoringWeight{
		BaseWeight:                  750, // 7.5% weight
		WildcardBonus:               100, // +1% for wildcard
		VerificationAgeBonusPerYear: 50,  // +0.5% per year
		MaxAgeBonus:                 200, // max +2% from age
	}
}

// CalculateDomainScore calculates the score contribution for a domain verification
func CalculateDomainScore(record *DomainVerificationRecord, weight DomainScoringWeight, now time.Time) uint32 {
	if !record.IsActive() {
		return 0
	}

	score := weight.BaseWeight

	if record.IsWildcard {
		score += weight.WildcardBonus
	}

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
