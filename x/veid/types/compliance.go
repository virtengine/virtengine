// Package types provides VEID module types.
//
// This file defines compliance types for KYC/AML regulatory requirements.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// Compliance Constants
// ============================================================================

// Compliance configuration defaults
const (
	// DefaultCheckExpiryBlocks is how long compliance checks remain valid (90 days @ 6s blocks)
	DefaultCheckExpiryBlocks int64 = 1296000

	// DefaultRiskScoreThreshold is the default maximum allowed risk score
	DefaultRiskScoreThreshold int32 = 75

	// DefaultMinAttestationsRequired is the minimum validator attestations needed
	DefaultMinAttestationsRequired int32 = 3

	// MaxRiskScore is the maximum risk score value
	MaxRiskScore int32 = 100

	// MinRiskScore is the minimum risk score value
	MinRiskScore int32 = 0

	// MaxProviderIDLength is the maximum length of a compliance provider ID
	MaxProviderIDLength = 64

	// MaxDetailsLength is the maximum length of check result details
	MaxDetailsLength = 1000

	// MaxRestrictedRegions is the maximum number of restricted regions
	MaxRestrictedRegions = 50

	// MaxAttestations is the maximum number of attestations per record
	MaxAttestations = 100
)

// ============================================================================
// Compliance Status
// ============================================================================

// ComplianceStatus represents the compliance state of an identity
type ComplianceStatus int32

const (
	// ComplianceStatusUnknown indicates no compliance check has been performed
	ComplianceStatusUnknown ComplianceStatus = 0

	// ComplianceStatusPending indicates compliance check is in progress
	ComplianceStatusPending ComplianceStatus = 1

	// ComplianceStatusCleared indicates identity passed all compliance checks
	ComplianceStatusCleared ComplianceStatus = 2

	// ComplianceStatusFlagged indicates identity has been flagged for review
	ComplianceStatusFlagged ComplianceStatus = 3

	// ComplianceStatusBlocked indicates identity is blocked from transactions
	ComplianceStatusBlocked ComplianceStatus = 4

	// ComplianceStatusExpired indicates compliance check has expired
	ComplianceStatusExpired ComplianceStatus = 5
)

// String returns the string representation of the compliance status
func (s ComplianceStatus) String() string {
	switch s {
	case ComplianceStatusUnknown:
		return "UNKNOWN"
	case ComplianceStatusPending:
		return "PENDING"
	case ComplianceStatusCleared:
		return "CLEARED"
	case ComplianceStatusFlagged:
		return "FLAGGED"
	case ComplianceStatusBlocked:
		return "BLOCKED"
	case ComplianceStatusExpired:
		return "EXPIRED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// IsValid returns true if the compliance status is a valid value
func (s ComplianceStatus) IsValid() bool {
	return s >= ComplianceStatusUnknown && s <= ComplianceStatusExpired
}

// IsTerminal returns true if the status requires no further action
func (s ComplianceStatus) IsTerminal() bool {
	return s == ComplianceStatusBlocked
}

// AllowsTransactions returns true if identity can perform transactions
func (s ComplianceStatus) AllowsTransactions() bool {
	return s == ComplianceStatusCleared
}

// ============================================================================
// Compliance Check Types
// ============================================================================

// ComplianceCheckType defines what type of compliance check
type ComplianceCheckType int32

const (
	// ComplianceCheckSanctionList checks against global sanction lists (OFAC, UN, EU, etc.)
	ComplianceCheckSanctionList ComplianceCheckType = 0

	// ComplianceCheckPEP checks for Politically Exposed Person status
	ComplianceCheckPEP ComplianceCheckType = 1

	// ComplianceCheckAdverseMedia checks for negative news coverage
	ComplianceCheckAdverseMedia ComplianceCheckType = 2

	// ComplianceCheckGeographic checks geographic restrictions
	ComplianceCheckGeographic ComplianceCheckType = 3

	// ComplianceCheckWatchlist checks against custom watchlists
	ComplianceCheckWatchlist ComplianceCheckType = 4

	// ComplianceCheckDocumentVerification verifies identity documents
	ComplianceCheckDocumentVerification ComplianceCheckType = 5

	// ComplianceCheckAMLRisk assesses anti-money laundering risk
	ComplianceCheckAMLRisk ComplianceCheckType = 6
)

// String returns the string representation of the check type
func (t ComplianceCheckType) String() string {
	switch t {
	case ComplianceCheckSanctionList:
		return "SANCTION_LIST"
	case ComplianceCheckPEP:
		return "PEP"
	case ComplianceCheckAdverseMedia:
		return "ADVERSE_MEDIA"
	case ComplianceCheckGeographic:
		return "GEOGRAPHIC"
	case ComplianceCheckWatchlist:
		return "WATCHLIST"
	case ComplianceCheckDocumentVerification:
		return "DOCUMENT_VERIFICATION"
	case ComplianceCheckAMLRisk:
		return "AML_RISK"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// IsValid returns true if the check type is a valid value
func (t ComplianceCheckType) IsValid() bool {
	return t >= ComplianceCheckSanctionList && t <= ComplianceCheckAMLRisk
}

// ============================================================================
// Compliance Check Result
// ============================================================================

// ComplianceCheckResult stores result of a single compliance check
type ComplianceCheckResult struct {
	// CheckType indicates what kind of check was performed
	CheckType ComplianceCheckType `json:"check_type"`

	// Passed indicates whether the check passed (true) or failed (false)
	Passed bool `json:"passed"`

	// Details provides additional context about the check result
	Details string `json:"details,omitempty"`

	// MatchScore indicates confidence of any matches found (0-100)
	// Higher score means higher confidence of match against watchlist/sanctions
	MatchScore int32 `json:"match_score,omitempty"`

	// CheckedAt is the Unix timestamp when the check was performed
	CheckedAt int64 `json:"checked_at"`

	// ProviderID identifies the compliance provider that performed the check
	ProviderID string `json:"provider_id"`

	// ReferenceID is the provider's reference for this check
	ReferenceID string `json:"reference_id,omitempty"`
}

// Validate validates the compliance check result
func (r *ComplianceCheckResult) Validate() error {
	if !r.CheckType.IsValid() {
		return ErrComplianceCheckFailed.Wrap("invalid check type")
	}

	if r.MatchScore < MinRiskScore || r.MatchScore > MaxRiskScore {
		return ErrRiskScoreExceeded.Wrapf("match score must be between %d and %d", MinRiskScore, MaxRiskScore)
	}

	if r.CheckedAt <= 0 {
		return ErrComplianceCheckFailed.Wrap("checked_at must be positive")
	}

	if r.ProviderID == "" {
		return ErrNotComplianceProvider.Wrap("provider_id is required")
	}

	if len(r.ProviderID) > MaxProviderIDLength {
		return ErrNotComplianceProvider.Wrapf("provider_id exceeds maximum length of %d", MaxProviderIDLength)
	}

	if len(r.Details) > MaxDetailsLength {
		return ErrComplianceCheckFailed.Wrapf("details exceeds maximum length of %d", MaxDetailsLength)
	}

	return nil
}

// ============================================================================
// Compliance Attestation
// ============================================================================

// ComplianceAttestation is a validator attestation of compliance status
type ComplianceAttestation struct {
	// ValidatorAddress is the address of the attesting validator
	ValidatorAddress string `json:"validator_address"`

	// AttestedAt is the Unix timestamp when attestation was made
	AttestedAt int64 `json:"attested_at"`

	// ExpiresAt is the Unix timestamp when this attestation expires
	ExpiresAt int64 `json:"expires_at"`

	// AttestationType describes what is being attested
	AttestationType string `json:"attestation_type"`

	// AttestationHash is a hash of the attestation data for verification
	AttestationHash string `json:"attestation_hash,omitempty"`
}

// Validate validates the compliance attestation
func (a *ComplianceAttestation) Validate() error {
	if a.ValidatorAddress == "" {
		return ErrInsufficientAttestations.Wrap("validator_address is required")
	}

	// Note: Address format validation is done at the keeper level

	if a.AttestedAt <= 0 {
		return ErrInsufficientAttestations.Wrap("attested_at must be positive")
	}

	if a.ExpiresAt <= a.AttestedAt {
		return ErrComplianceExpired.Wrap("expires_at must be after attested_at")
	}

	if a.AttestationType == "" {
		return ErrInsufficientAttestations.Wrap("attestation_type is required")
	}

	return nil
}

// IsExpired checks if the attestation has expired
func (a *ComplianceAttestation) IsExpired(currentTime int64) bool {
	return currentTime >= a.ExpiresAt
}

// ============================================================================
// Compliance Record
// ============================================================================

// ComplianceRecord stores the complete compliance status for an identity
type ComplianceRecord struct {
	// AccountAddress is the blockchain address of the identity
	AccountAddress string `json:"account_address"`

	// Status is the overall compliance status
	Status ComplianceStatus `json:"status"`

	// CheckResults contains results of individual compliance checks
	CheckResults []ComplianceCheckResult `json:"check_results"`

	// LastCheckedAt is the Unix timestamp of the last compliance check
	LastCheckedAt int64 `json:"last_checked_at"`

	// ExpiresAt is the Unix timestamp when the compliance record expires
	ExpiresAt int64 `json:"expires_at"`

	// RiskScore is the overall risk score (0-100, lower is better)
	RiskScore int32 `json:"risk_score"`

	// RestrictedRegions lists regions where this identity is restricted
	RestrictedRegions []string `json:"restricted_regions,omitempty"`

	// Attestations contains validator attestations of compliance
	Attestations []ComplianceAttestation `json:"attestations"`

	// CreatedAt is when this record was first created
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is when this record was last updated
	UpdatedAt int64 `json:"updated_at"`

	// Notes contains any additional notes about the compliance status
	Notes string `json:"notes,omitempty"`
}

// NewComplianceRecord creates a new compliance record
func NewComplianceRecord(accountAddress string, createdAt time.Time) *ComplianceRecord {
	timestamp := createdAt.Unix()
	return &ComplianceRecord{
		AccountAddress:    accountAddress,
		Status:            ComplianceStatusUnknown,
		CheckResults:      make([]ComplianceCheckResult, 0),
		LastCheckedAt:     0,
		ExpiresAt:         0,
		RiskScore:         0,
		RestrictedRegions: make([]string, 0),
		Attestations:      make([]ComplianceAttestation, 0),
		CreatedAt:         timestamp,
		UpdatedAt:         timestamp,
	}
}

// Validate validates the compliance record
func (r *ComplianceRecord) Validate() error {
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	// Note: Address format validation is done at the keeper level when storing,
	// matching the pattern used by IdentityRecord.

	if !r.Status.IsValid() {
		return ErrComplianceCheckFailed.Wrapf("invalid compliance status: %d", r.Status)
	}

	if r.RiskScore < MinRiskScore || r.RiskScore > MaxRiskScore {
		return ErrRiskScoreExceeded.Wrapf("risk score must be between %d and %d", MinRiskScore, MaxRiskScore)
	}

	if len(r.RestrictedRegions) > MaxRestrictedRegions {
		return ErrRestrictedRegion.Wrapf("too many restricted regions (max %d)", MaxRestrictedRegions)
	}

	if len(r.Attestations) > MaxAttestations {
		return ErrInsufficientAttestations.Wrapf("too many attestations (max %d)", MaxAttestations)
	}

	// Validate check results
	for i, result := range r.CheckResults {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("invalid check result at index %d: %w", i, err)
		}
	}

	// Validate attestations
	for i, attestation := range r.Attestations {
		if err := attestation.Validate(); err != nil {
			return fmt.Errorf("invalid attestation at index %d: %w", i, err)
		}
	}

	if r.CreatedAt <= 0 {
		return ErrComplianceCheckFailed.Wrap("created_at must be positive")
	}

	if r.UpdatedAt < r.CreatedAt {
		return ErrComplianceCheckFailed.Wrap("updated_at cannot be before created_at")
	}

	return nil
}

// IsExpired checks if the compliance record has expired
func (r *ComplianceRecord) IsExpired(currentTime int64) bool {
	return r.ExpiresAt > 0 && currentTime >= r.ExpiresAt
}

// HasValidAttestations checks if record has minimum required valid attestations
func (r *ComplianceRecord) HasValidAttestations(minRequired int32, currentTime int64) bool {
	validCount := int32(0)
	for _, attestation := range r.Attestations {
		if !attestation.IsExpired(currentTime) {
			validCount++
		}
	}
	return validCount >= minRequired
}

// GetValidatorAddresses returns the addresses of validators who have attested
func (r *ComplianceRecord) GetValidatorAddresses() []string {
	addresses := make([]string, len(r.Attestations))
	for i, attestation := range r.Attestations {
		addresses[i] = attestation.ValidatorAddress
	}
	return addresses
}

// AddCheckResult adds a new check result and updates the record
func (r *ComplianceRecord) AddCheckResult(result ComplianceCheckResult, updateTime int64) {
	r.CheckResults = append(r.CheckResults, result)
	r.LastCheckedAt = result.CheckedAt
	r.UpdatedAt = updateTime
}

// AddAttestation adds a validator attestation to the record
func (r *ComplianceRecord) AddAttestation(attestation ComplianceAttestation, updateTime int64) error {
	if err := attestation.Validate(); err != nil {
		return err
	}

	if len(r.Attestations) >= MaxAttestations {
		return ErrInsufficientAttestations.Wrapf("maximum attestations (%d) reached", MaxAttestations)
	}

	// Check for duplicate validator
	for _, existing := range r.Attestations {
		if existing.ValidatorAddress == attestation.ValidatorAddress {
			return ErrInsufficientAttestations.Wrap("validator has already attested")
		}
	}

	r.Attestations = append(r.Attestations, attestation)
	r.UpdatedAt = updateTime
	return nil
}

// CalculateRiskScore calculates overall risk score from check results
func (r *ComplianceRecord) CalculateRiskScore() int32 {
	if len(r.CheckResults) == 0 {
		return 0
	}

	var totalScore int32
	var count int32

	for _, result := range r.CheckResults {
		if !result.Passed {
			// Failed checks contribute their match score to risk
			totalScore += result.MatchScore
		}
		count++
	}

	if count == 0 {
		return 0
	}

	// Average score, weighted by failed checks
	avgScore := totalScore / count

	// Ensure within bounds
	if avgScore > MaxRiskScore {
		return MaxRiskScore
	}
	if avgScore < MinRiskScore {
		return MinRiskScore
	}

	return avgScore
}

// GenerateRecordHash generates a unique hash for this compliance record
func (r *ComplianceRecord) GenerateRecordHash() string {
	data := fmt.Sprintf("%s|%d|%d|%d|%d",
		r.AccountAddress,
		r.Status,
		r.LastCheckedAt,
		r.RiskScore,
		len(r.CheckResults),
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ============================================================================
// Compliance Parameters
// ============================================================================

// ComplianceParams configures the compliance module behavior
type ComplianceParams struct {
	// RequireSanctionCheck indicates if sanction list check is mandatory
	RequireSanctionCheck bool `json:"require_sanction_check"`

	// RequirePEPCheck indicates if PEP check is mandatory
	RequirePEPCheck bool `json:"require_pep_check"`

	// CheckExpiryBlocks is how long compliance checks remain valid
	CheckExpiryBlocks int64 `json:"check_expiry_blocks"`

	// RiskScoreThreshold is the maximum allowed risk score (0-100)
	RiskScoreThreshold int32 `json:"risk_score_threshold"`

	// RestrictedCountries is list of ISO country codes that are restricted
	RestrictedCountries []string `json:"restricted_countries"`

	// MinAttestationsRequired is minimum validator attestations needed
	MinAttestationsRequired int32 `json:"min_attestations_required"`

	// EnableAutoExpiry enables automatic expiration of compliance records
	EnableAutoExpiry bool `json:"enable_auto_expiry"`

	// RequireDocumentVerification indicates if document verification is mandatory
	RequireDocumentVerification bool `json:"require_document_verification"`
}

// DefaultComplianceParams returns default compliance parameters
func DefaultComplianceParams() ComplianceParams {
	return ComplianceParams{
		RequireSanctionCheck:        true,
		RequirePEPCheck:             true,
		CheckExpiryBlocks:           DefaultCheckExpiryBlocks,
		RiskScoreThreshold:          DefaultRiskScoreThreshold,
		RestrictedCountries:         []string{},
		MinAttestationsRequired:     DefaultMinAttestationsRequired,
		EnableAutoExpiry:            true,
		RequireDocumentVerification: false,
	}
}

// Validate validates the compliance parameters
func (p ComplianceParams) Validate() error {
	if p.CheckExpiryBlocks <= 0 {
		return ErrComplianceExpired.Wrap("check_expiry_blocks must be positive")
	}

	if p.RiskScoreThreshold < MinRiskScore || p.RiskScoreThreshold > MaxRiskScore {
		return ErrRiskScoreExceeded.Wrapf("risk_score_threshold must be between %d and %d", MinRiskScore, MaxRiskScore)
	}

	if p.MinAttestationsRequired < 0 {
		return ErrInsufficientAttestations.Wrap("min_attestations_required cannot be negative")
	}

	if len(p.RestrictedCountries) > MaxRestrictedRegions {
		return ErrRestrictedRegion.Wrapf("too many restricted countries (max %d)", MaxRestrictedRegions)
	}

	return nil
}

// ============================================================================
// Compliance Provider
// ============================================================================

// ComplianceProvider represents an authorized external compliance provider
type ComplianceProvider struct {
	// ProviderID is the unique identifier for this provider
	ProviderID string `json:"provider_id"`

	// Name is the human-readable name of the provider
	Name string `json:"name"`

	// ProviderAddress is the blockchain address authorized to submit checks
	ProviderAddress string `json:"provider_address"`

	// SupportedCheckTypes lists which check types this provider can perform
	SupportedCheckTypes []ComplianceCheckType `json:"supported_check_types"`

	// IsActive indicates if this provider is currently active
	IsActive bool `json:"is_active"`

	// RegisteredAt is when this provider was registered
	RegisteredAt int64 `json:"registered_at"`

	// LastActiveAt is when this provider last submitted a check
	LastActiveAt int64 `json:"last_active_at"`
}

// Validate validates the compliance provider
func (p *ComplianceProvider) Validate() error {
	if p.ProviderID == "" {
		return ErrNotComplianceProvider.Wrap("provider_id is required")
	}

	if len(p.ProviderID) > MaxProviderIDLength {
		return ErrNotComplianceProvider.Wrapf("provider_id exceeds maximum length of %d", MaxProviderIDLength)
	}

	if p.Name == "" {
		return ErrNotComplianceProvider.Wrap("name is required")
	}

	if p.ProviderAddress == "" {
		return ErrNotComplianceProvider.Wrap("provider_address is required")
	}

	// Note: Address format validation is done at the keeper level

	if len(p.SupportedCheckTypes) == 0 {
		return ErrNotComplianceProvider.Wrap("at least one supported check type is required")
	}

	for _, checkType := range p.SupportedCheckTypes {
		if !checkType.IsValid() {
			return ErrComplianceCheckFailed.Wrapf("invalid check type: %d", checkType)
		}
	}

	return nil
}

// SupportsCheckType checks if this provider supports a given check type
func (p *ComplianceProvider) SupportsCheckType(checkType ComplianceCheckType) bool {
	for _, supported := range p.SupportedCheckTypes {
		if supported == checkType {
			return true
		}
	}
	return false
}
