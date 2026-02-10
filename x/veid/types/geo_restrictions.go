// Package types provides VEID module types.
//
// This file defines geographic restriction types for VEID compliance.
//
// Task Reference: VE-3032 - Add Geographic Restriction Rules for VEID
package types

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// Geographic Restriction Constants
// ============================================================================

const (
	// MaxGeoRestrictionPolicies is the maximum number of geo restriction policies
	MaxGeoRestrictionPolicies = 100

	// MaxCountriesPerPolicy is the maximum number of countries in a single policy
	MaxCountriesPerPolicy = 250

	// MaxRegionsPerPolicy is the maximum number of regions in a single policy
	MaxRegionsPerPolicy = 500

	// MaxPolicyIDLength is the maximum length of a policy ID
	MaxPolicyIDLength = 64

	// MaxPolicyNameLength is the maximum length of a policy name
	MaxPolicyNameLength = 256

	// MaxBlockReasonLength is the maximum length of a block reason
	MaxBlockReasonLength = 1024

	// DefaultGeoPolicyPriority is the default priority for new policies
	DefaultGeoPolicyPriority = 100
)

// ISO 3166-1 alpha-2 country code pattern
var iso3166Alpha2Pattern = regexp.MustCompile(`^[A-Z]{2}$`)

// ISO 3166-2 subdivision pattern (e.g., US-CA, GB-ENG)
var iso3166Alpha2SubdivisionPattern = regexp.MustCompile(`^[A-Z]{2}-[A-Z0-9]{1,3}$`)

// ============================================================================
// Enforcement Level
// ============================================================================

// EnforcementLevel determines how strictly a policy is enforced
type EnforcementLevel int32

const (
	// EnforcementWarn logs a warning but allows the transaction
	EnforcementWarn EnforcementLevel = 0

	// EnforcementSoftBlock blocks the transaction but allows override with MFA
	EnforcementSoftBlock EnforcementLevel = 1

	// EnforcementHardBlock blocks the transaction with no override possible
	EnforcementHardBlock EnforcementLevel = 2
)

// String returns the string representation of the enforcement level
func (e EnforcementLevel) String() string {
	switch e {
	case EnforcementWarn:
		return "WARN"
	case EnforcementSoftBlock:
		return "SOFT_BLOCK"
	case EnforcementHardBlock:
		return "HARD_BLOCK"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", e)
	}
}

// IsValid returns true if the enforcement level is valid
func (e EnforcementLevel) IsValid() bool {
	return e >= EnforcementWarn && e <= EnforcementHardBlock
}

// AllowsOverride returns true if the enforcement level allows MFA override
func (e EnforcementLevel) AllowsOverride() bool {
	return e == EnforcementWarn || e == EnforcementSoftBlock
}

// ============================================================================
// Policy Status
// ============================================================================

// PolicyStatus represents the current state of a geo restriction policy
type PolicyStatus int32

const (
	// PolicyStatusDraft indicates the policy is being drafted
	PolicyStatusDraft PolicyStatus = 0

	// PolicyStatusActive indicates the policy is actively enforced
	PolicyStatusActive PolicyStatus = 1

	// PolicyStatusDisabled indicates the policy is temporarily disabled
	PolicyStatusDisabled PolicyStatus = 2

	// PolicyStatusArchived indicates the policy is archived (soft deleted)
	PolicyStatusArchived PolicyStatus = 3
)

// String returns the string representation of the policy status
func (s PolicyStatus) String() string {
	switch s {
	case PolicyStatusDraft:
		return "DRAFT"
	case PolicyStatusActive:
		return "ACTIVE"
	case PolicyStatusDisabled:
		return "DISABLED"
	case PolicyStatusArchived:
		return "ARCHIVED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// IsValid returns true if the policy status is valid
func (s PolicyStatus) IsValid() bool {
	return s >= PolicyStatusDraft && s <= PolicyStatusArchived
}

// IsEnforceable returns true if the policy can be enforced
func (s PolicyStatus) IsEnforceable() bool {
	return s == PolicyStatusActive
}

// ============================================================================
// Geographic Restriction Policy
// ============================================================================

// GeoRestrictionPolicy defines geographic access restrictions
type GeoRestrictionPolicy struct {
	// PolicyID is the unique identifier for this policy
	PolicyID string `json:"policy_id"`

	// Name is the human-readable name of the policy
	Name string `json:"name"`

	// Description provides additional context about the policy
	Description string `json:"description,omitempty"`

	// AllowedCountries are ISO 3166-1 alpha-2 codes that are explicitly allowed
	// If non-empty, only these countries are permitted (allowlist mode)
	AllowedCountries []string `json:"allowed_countries,omitempty"`

	// BlockedCountries are ISO 3166-1 alpha-2 codes that are explicitly blocked
	// Applied after allowed countries check (blocklist mode)
	BlockedCountries []string `json:"blocked_countries,omitempty"`

	// AllowedRegions are ISO 3166-2 subdivision codes that are explicitly allowed
	// More specific than country-level, e.g., US-CA for California
	AllowedRegions []string `json:"allowed_regions,omitempty"`

	// BlockedRegions are ISO 3166-2 subdivision codes that are explicitly blocked
	BlockedRegions []string `json:"blocked_regions,omitempty"`

	// RequireIPMatch requires the user's IP geolocation to match document country
	RequireIPMatch bool `json:"require_ip_match"`

	// EnforcementLevel determines how strictly the policy is enforced
	EnforcementLevel EnforcementLevel `json:"enforcement_level"`

	// Status is the current state of the policy
	Status PolicyStatus `json:"status"`

	// Priority determines order of evaluation (lower = higher priority)
	Priority int32 `json:"priority"`

	// ApplicableScopes limits which verification scopes this policy applies to
	// Empty means applies to all scopes
	ApplicableScopes []string `json:"applicable_scopes,omitempty"`

	// ApplicableMarkets limits which markets this policy applies to
	// Empty means applies to all markets
	ApplicableMarkets []string `json:"applicable_markets,omitempty"`

	// CreatedAt is when the policy was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the policy was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// CreatedBy is the address that created this policy
	CreatedBy string `json:"created_by"`

	// UpdatedBy is the address that last updated this policy
	UpdatedBy string `json:"updated_by,omitempty"`

	// Notes are optional admin notes about the policy
	Notes string `json:"notes,omitempty"`
}

// NewGeoRestrictionPolicy creates a new geographic restriction policy
func NewGeoRestrictionPolicy(policyID, name, createdBy string, now time.Time) *GeoRestrictionPolicy {
	return &GeoRestrictionPolicy{
		PolicyID:         policyID,
		Name:             name,
		EnforcementLevel: EnforcementWarn,
		Status:           PolicyStatusDraft,
		Priority:         DefaultGeoPolicyPriority,
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedBy:        createdBy,
	}
}

// Validate validates the geo restriction policy
func (p *GeoRestrictionPolicy) Validate() error {
	if p.PolicyID == "" {
		return ErrGeoRestrictionInvalid.Wrap("policy_id is required")
	}
	if len(p.PolicyID) > MaxPolicyIDLength {
		return ErrGeoRestrictionInvalid.Wrapf("policy_id exceeds max length %d", MaxPolicyIDLength)
	}
	if p.Name == "" {
		return ErrGeoRestrictionInvalid.Wrap("name is required")
	}
	if len(p.Name) > MaxPolicyNameLength {
		return ErrGeoRestrictionInvalid.Wrapf("name exceeds max length %d", MaxPolicyNameLength)
	}
	if !p.EnforcementLevel.IsValid() {
		return ErrGeoRestrictionInvalid.Wrap("invalid enforcement_level")
	}
	if !p.Status.IsValid() {
		return ErrGeoRestrictionInvalid.Wrap("invalid status")
	}
	if len(p.AllowedCountries) > MaxCountriesPerPolicy {
		return ErrGeoRestrictionInvalid.Wrapf("allowed_countries exceeds max %d", MaxCountriesPerPolicy)
	}
	if len(p.BlockedCountries) > MaxCountriesPerPolicy {
		return ErrGeoRestrictionInvalid.Wrapf("blocked_countries exceeds max %d", MaxCountriesPerPolicy)
	}
	if len(p.AllowedRegions) > MaxRegionsPerPolicy {
		return ErrGeoRestrictionInvalid.Wrapf("allowed_regions exceeds max %d", MaxRegionsPerPolicy)
	}
	if len(p.BlockedRegions) > MaxRegionsPerPolicy {
		return ErrGeoRestrictionInvalid.Wrapf("blocked_regions exceeds max %d", MaxRegionsPerPolicy)
	}

	// Validate country codes
	for _, code := range p.AllowedCountries {
		if err := ValidateCountryCode(code); err != nil {
			return err
		}
	}
	for _, code := range p.BlockedCountries {
		if err := ValidateCountryCode(code); err != nil {
			return err
		}
	}

	// Validate region codes
	for _, code := range p.AllowedRegions {
		if err := ValidateRegionCode(code); err != nil {
			return err
		}
	}
	for _, code := range p.BlockedRegions {
		if err := ValidateRegionCode(code); err != nil {
			return err
		}
	}

	if p.CreatedBy == "" {
		return ErrGeoRestrictionInvalid.Wrap("created_by is required")
	}

	return nil
}

// IsActive returns true if the policy is actively enforced
func (p *GeoRestrictionPolicy) IsActive() bool {
	return p.Status.IsEnforceable()
}

// HasAllowlist returns true if the policy uses allowlist mode
func (p *GeoRestrictionPolicy) HasAllowlist() bool {
	return len(p.AllowedCountries) > 0 || len(p.AllowedRegions) > 0
}

// HasBlocklist returns true if the policy uses blocklist mode
func (p *GeoRestrictionPolicy) HasBlocklist() bool {
	return len(p.BlockedCountries) > 0 || len(p.BlockedRegions) > 0
}

// AppliesToScope returns true if this policy applies to the given scope
func (p *GeoRestrictionPolicy) AppliesToScope(scopeType string) bool {
	if len(p.ApplicableScopes) == 0 {
		return true // Empty means applies to all
	}
	for _, s := range p.ApplicableScopes {
		if s == scopeType {
			return true
		}
	}
	return false
}

// AppliesToMarket returns true if this policy applies to the given market
func (p *GeoRestrictionPolicy) AppliesToMarket(marketID string) bool {
	if len(p.ApplicableMarkets) == 0 {
		return true // Empty means applies to all
	}
	for _, m := range p.ApplicableMarkets {
		if m == marketID {
			return true
		}
	}
	return false
}

// ============================================================================
// Geographic Check Result
// ============================================================================

// GeoCheckResult stores the result of a geographic compliance check
type GeoCheckResult struct {
	// Address is the account address being checked
	Address string `json:"address"`

	// Country is the ISO 3166-1 alpha-2 country code detected
	Country string `json:"country"`

	// Region is the ISO 3166-2 subdivision code detected (if available)
	Region string `json:"region,omitempty"`

	// IPCountry is the country detected from IP geolocation (if checked)
	IPCountry string `json:"ip_country,omitempty"`

	// IsAllowed indicates whether the geographic location is allowed
	IsAllowed bool `json:"is_allowed"`

	// MatchedPolicyID is the ID of the policy that determined the result
	MatchedPolicyID string `json:"matched_policy_id,omitempty"`

	// MatchedPolicyName is the name of the matched policy
	MatchedPolicyName string `json:"matched_policy_name,omitempty"`

	// BlockReason explains why access was blocked (if blocked)
	BlockReason string `json:"block_reason,omitempty"`

	// EnforcementLevel is the enforcement level of the matched policy
	EnforcementLevel EnforcementLevel `json:"enforcement_level"`

	// AllowsOverride indicates if the block can be overridden with MFA
	AllowsOverride bool `json:"allows_override"`

	// IPMismatch indicates if IP geolocation didn't match document country
	IPMismatch bool `json:"ip_mismatch,omitempty"`

	// CheckedAt is when the check was performed
	CheckedAt time.Time `json:"checked_at"`

	// EvaluatedPolicies is the count of policies evaluated
	EvaluatedPolicies int32 `json:"evaluated_policies"`
}

// NewGeoCheckResult creates a new geographic check result
func NewGeoCheckResult(address, country string, now time.Time) *GeoCheckResult {
	return &GeoCheckResult{
		Address:   address,
		Country:   country,
		IsAllowed: true, // Default to allowed
		CheckedAt: now,
	}
}

// Validate validates the geo check result
func (r *GeoCheckResult) Validate() error {
	if r.Address == "" {
		return ErrGeoRestrictionInvalid.Wrap("address is required")
	}
	if r.Country != "" {
		if err := ValidateCountryCode(r.Country); err != nil {
			return err
		}
	}
	if r.Region != "" {
		if err := ValidateRegionCode(r.Region); err != nil {
			return err
		}
	}
	if len(r.BlockReason) > MaxBlockReasonLength {
		return ErrGeoRestrictionInvalid.Wrapf("block_reason exceeds max length %d", MaxBlockReasonLength)
	}
	return nil
}

// ============================================================================
// Geographic Location Data
// ============================================================================

// GeoLocation represents geographic location data for an identity
type GeoLocation struct {
	// Country is the ISO 3166-1 alpha-2 country code
	Country string `json:"country"`

	// Region is the ISO 3166-2 subdivision code
	Region string `json:"region,omitempty"`

	// City is the city name (optional)
	City string `json:"city,omitempty"`

	// PostalCode is the postal/zip code (optional)
	PostalCode string `json:"postal_code,omitempty"`

	// Source indicates where this location data came from
	Source GeoLocationSource `json:"source"`

	// Confidence is the confidence level (0-100) of the location data
	Confidence int32 `json:"confidence"`

	// DetectedAt is when this location was detected
	DetectedAt time.Time `json:"detected_at"`
}

// GeoLocationSource indicates the source of location data
type GeoLocationSource int32

const (
	// GeoLocationSourceUnknown indicates unknown source
	GeoLocationSourceUnknown GeoLocationSource = 0

	// GeoLocationSourceDocument indicates location from identity document
	GeoLocationSourceDocument GeoLocationSource = 1

	// GeoLocationSourceIP indicates location from IP geolocation
	GeoLocationSourceIP GeoLocationSource = 2

	// GeoLocationSourceUserDeclared indicates user-declared location
	GeoLocationSourceUserDeclared GeoLocationSource = 3

	// GeoLocationSourceSSO indicates location from SSO provider
	GeoLocationSourceSSO GeoLocationSource = 4
)

// String returns the string representation of the location source
func (s GeoLocationSource) String() string {
	switch s {
	case GeoLocationSourceUnknown:
		return "UNKNOWN"
	case GeoLocationSourceDocument:
		return "DOCUMENT"
	case GeoLocationSourceIP:
		return "IP"
	case GeoLocationSourceUserDeclared:
		return "USER_DECLARED"
	case GeoLocationSourceSSO:
		return "SSO"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// IsValid returns true if the location source is valid
func (s GeoLocationSource) IsValid() bool {
	return s >= GeoLocationSourceUnknown && s <= GeoLocationSourceSSO
}

// Validate validates the geo location
func (l *GeoLocation) Validate() error {
	if l.Country == "" {
		return ErrGeoRestrictionInvalid.Wrap("country is required")
	}
	if err := ValidateCountryCode(l.Country); err != nil {
		return err
	}
	if l.Region != "" {
		if err := ValidateRegionCode(l.Region); err != nil {
			return err
		}
	}
	if !l.Source.IsValid() {
		return ErrGeoRestrictionInvalid.Wrap("invalid location source")
	}
	if l.Confidence < 0 || l.Confidence > 100 {
		return ErrGeoRestrictionInvalid.Wrap("confidence must be between 0 and 100")
	}
	return nil
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code
func ValidateCountryCode(code string) error {
	if code == "" {
		return ErrGeoRestrictionInvalid.Wrap("country code is required")
	}
	upperCode := strings.ToUpper(code)
	if !iso3166Alpha2Pattern.MatchString(upperCode) {
		return ErrGeoRestrictionInvalid.Wrapf("invalid ISO 3166-1 alpha-2 country code: %s", code)
	}
	return nil
}

// ValidateRegionCode validates an ISO 3166-2 subdivision code
func ValidateRegionCode(code string) error {
	if code == "" {
		return ErrGeoRestrictionInvalid.Wrap("region code is required")
	}
	upperCode := strings.ToUpper(code)
	if !iso3166Alpha2SubdivisionPattern.MatchString(upperCode) {
		return ErrGeoRestrictionInvalid.Wrapf("invalid ISO 3166-2 region code: %s", code)
	}
	return nil
}

// NormalizeCountryCode normalizes a country code to uppercase
func NormalizeCountryCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// NormalizeRegionCode normalizes a region code to uppercase
func NormalizeRegionCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// ExtractCountryFromRegion extracts the country code from a region code
// e.g., "US-CA" -> "US"
func ExtractCountryFromRegion(regionCode string) string {
	parts := strings.Split(regionCode, "-")
	if len(parts) >= 1 {
		return NormalizeCountryCode(parts[0])
	}
	return ""
}

// ============================================================================
// Geographic Restriction Parameters
// ============================================================================

// GeoRestrictionParams defines module parameters for geographic restrictions
type GeoRestrictionParams struct {
	// Enabled indicates if geographic restrictions are enforced
	Enabled bool `json:"enabled"`

	// DefaultEnforcementLevel is the default level for new policies
	DefaultEnforcementLevel EnforcementLevel `json:"default_enforcement_level"`

	// RequireIPVerification requires IP geolocation for all checks
	RequireIPVerification bool `json:"require_ip_verification"`

	// AllowOverrideWithMFA allows soft blocks to be overridden with MFA
	AllowOverrideWithMFA bool `json:"allow_override_with_mfa"`

	// MinConfidenceScore is the minimum confidence required for location data
	MinConfidenceScore int32 `json:"min_confidence_score"`

	// GlobalBlockedCountries are countries blocked across all policies
	GlobalBlockedCountries []string `json:"global_blocked_countries,omitempty"`
}

// DefaultGeoRestrictionParams returns default geographic restriction parameters
func DefaultGeoRestrictionParams() GeoRestrictionParams {
	return GeoRestrictionParams{
		Enabled:                 true,
		DefaultEnforcementLevel: EnforcementSoftBlock,
		RequireIPVerification:   false,
		AllowOverrideWithMFA:    true,
		MinConfidenceScore:      50,
		GlobalBlockedCountries:  []string{}, // No global blocks by default
	}
}

// Validate validates the geo restriction parameters
func (p *GeoRestrictionParams) Validate() error {
	if !p.DefaultEnforcementLevel.IsValid() {
		return ErrGeoRestrictionInvalid.Wrap("invalid default_enforcement_level")
	}
	if p.MinConfidenceScore < 0 || p.MinConfidenceScore > 100 {
		return ErrGeoRestrictionInvalid.Wrap("min_confidence_score must be between 0 and 100")
	}
	for _, code := range p.GlobalBlockedCountries {
		if err := ValidateCountryCode(code); err != nil {
			return err
		}
	}
	return nil
}
