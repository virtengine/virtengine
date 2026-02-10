// Package types provides types for the VEID module.
//
// VE-910: VoIP detection for SMS verification
// This file defines the VoIP detection interface for detecting virtual numbers.
package types

import (
	"time"
)

// CarrierType represents the type of phone carrier
type CarrierType string

const (
	// CarrierTypeMobile represents a mobile carrier
	CarrierTypeMobile CarrierType = "mobile"

	// CarrierTypeLandline represents a landline carrier
	CarrierTypeLandline CarrierType = "landline"

	// CarrierTypeVoIP represents a VoIP carrier
	CarrierTypeVoIP CarrierType = "voip"

	// CarrierTypeTollFree represents a toll-free number
	CarrierTypeTollFree CarrierType = "toll_free"

	// CarrierTypePremium represents a premium rate number
	CarrierTypePremium CarrierType = "premium"

	// CarrierTypeUnknown represents an unknown carrier type
	CarrierTypeUnknown CarrierType = "unknown"
)

// AllCarrierTypes returns all carrier types
func AllCarrierTypes() []CarrierType {
	return []CarrierType{
		CarrierTypeMobile,
		CarrierTypeLandline,
		CarrierTypeVoIP,
		CarrierTypeTollFree,
		CarrierTypePremium,
		CarrierTypeUnknown,
	}
}

// IsValidCarrierType checks if a carrier type is valid
func IsValidCarrierType(ct CarrierType) bool {
	for _, valid := range AllCarrierTypes() {
		if ct == valid {
			return true
		}
	}
	return false
}

// IsBlockedCarrierType returns true if this carrier type should be blocked
func IsBlockedCarrierType(ct CarrierType) bool {
	switch ct {
	case CarrierTypeVoIP, CarrierTypeTollFree, CarrierTypePremium:
		return true
	default:
		return false
	}
}

// VoIPRiskLevel represents the risk level of a phone number
type VoIPRiskLevel string

const (
	// VoIPRiskLow indicates low VoIP/fraud risk
	VoIPRiskLow VoIPRiskLevel = "low"

	// VoIPRiskMedium indicates medium VoIP/fraud risk
	VoIPRiskMedium VoIPRiskLevel = "medium"

	// VoIPRiskHigh indicates high VoIP/fraud risk
	VoIPRiskHigh VoIPRiskLevel = "high"

	// VoIPRiskCritical indicates critical VoIP/fraud risk (block)
	VoIPRiskCritical VoIPRiskLevel = "critical"
)

// CarrierLookupResult represents the result of a carrier lookup
type CarrierLookupResult struct {
	// PhoneHashRef is a reference to the phone hash (not the actual number)
	PhoneHashRef string `json:"phone_hash_ref"`

	// CountryCode is the ISO country code
	CountryCode string `json:"country_code"`

	// CarrierName is the name of the carrier (may be empty)
	CarrierName string `json:"carrier_name,omitempty"`

	// CarrierType is the type of carrier
	CarrierType CarrierType `json:"carrier_type"`

	// IsVoIP indicates if this is a VoIP number
	IsVoIP bool `json:"is_voip"`

	// IsPorted indicates if the number has been ported
	IsPorted bool `json:"is_ported"`

	// IsRoaming indicates if the number is currently roaming
	IsRoaming bool `json:"is_roaming"`

	// RiskLevel is the assessed VoIP/fraud risk level
	RiskLevel VoIPRiskLevel `json:"risk_level"`

	// RiskScore is a numeric risk score (0-100, higher = more risk)
	RiskScore uint32 `json:"risk_score"`

	// RiskFactors lists the factors contributing to the risk score
	RiskFactors []string `json:"risk_factors,omitempty"`

	// LookupProvider is the provider that performed the lookup
	LookupProvider string `json:"lookup_provider"`

	// LookupTimestamp is when the lookup was performed
	LookupTimestamp time.Time `json:"lookup_timestamp"`

	// IsCached indicates if this result was from cache
	IsCached bool `json:"is_cached"`

	// CacheExpiresAt is when the cached result expires
	CacheExpiresAt *time.Time `json:"cache_expires_at,omitempty"`

	// Success indicates if the lookup was successful
	Success bool `json:"success"`

	// ErrorCode is the error code if lookup failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if lookup failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// ShouldBlock returns true if this phone number should be blocked
func (r *CarrierLookupResult) ShouldBlock(blockVoIP bool) bool {
	if !r.Success {
		// If lookup failed, allow but with caution
		return false
	}

	// Block VoIP numbers if configured
	if blockVoIP && r.IsVoIP {
		return true
	}

	// Block if carrier type is blocked
	if IsBlockedCarrierType(r.CarrierType) {
		return true
	}

	// Block if critical risk
	if r.RiskLevel == VoIPRiskCritical {
		return true
	}

	// Block if high risk score
	if r.RiskScore >= 80 {
		return true
	}

	return false
}

// GetBlockReason returns the reason for blocking (if applicable)
func (r *CarrierLookupResult) GetBlockReason(blockVoIP bool) string {
	if !r.ShouldBlock(blockVoIP) {
		return ""
	}

	if blockVoIP && r.IsVoIP {
		return "VoIP numbers are not allowed for verification"
	}

	if IsBlockedCarrierType(r.CarrierType) {
		return "phone number type not allowed: " + string(r.CarrierType)
	}

	if r.RiskLevel == VoIPRiskCritical {
		return "phone number flagged as high fraud risk"
	}

	if r.RiskScore >= 80 {
		return "phone number has elevated risk score"
	}

	return "phone number blocked"
}

// Validate validates the carrier lookup result
func (r *CarrierLookupResult) Validate() error {
	if r.PhoneHashRef == "" {
		return ErrInvalidPhone.Wrap("phone_hash_ref cannot be empty")
	}
	if r.LookupTimestamp.IsZero() {
		return ErrInvalidPhone.Wrap("lookup_timestamp cannot be zero")
	}
	if !IsValidCarrierType(r.CarrierType) {
		return ErrInvalidPhone.Wrapf("invalid carrier type: %s", r.CarrierType)
	}
	return nil
}

// VoIPDetector defines the interface for VoIP detection services
// Implementations can use various carrier lookup APIs (Twilio, Plivo, etc.)
type VoIPDetector interface {
	// LookupCarrier performs a carrier lookup for a phone number
	// CRITICAL: The phone number should only be used for the API call,
	// never logged or stored. Return the result with only the hash reference.
	LookupCarrier(phoneNumber string, phoneHashRef string) (*CarrierLookupResult, error)

	// GetProviderName returns the name of the lookup provider
	GetProviderName() string

	// IsAvailable checks if the detector is available and configured
	IsAvailable() bool

	// GetRateLimit returns the rate limit for lookups (per minute)
	GetRateLimit() uint32

	// GetCacheTTL returns how long results should be cached (seconds)
	GetCacheTTL() int64
}

// VoIPDetectorConfig contains configuration for VoIP detection
type VoIPDetectorConfig struct {
	// Provider is the lookup provider (twilio, plivo, numverify, etc.)
	Provider string `json:"provider"`

	// Enabled indicates if VoIP detection is enabled
	Enabled bool `json:"enabled"`

	// BlockVoIP indicates if VoIP numbers should be blocked
	BlockVoIP bool `json:"block_voip"`

	// CacheTTLSeconds is how long to cache lookup results
	CacheTTLSeconds int64 `json:"cache_ttl_seconds"`

	// RateLimitPerMinute is the max lookups per minute
	RateLimitPerMinute uint32 `json:"rate_limit_per_minute"`

	// HighRiskThreshold is the risk score threshold for high risk
	HighRiskThreshold uint32 `json:"high_risk_threshold"`

	// BlockThreshold is the risk score threshold for blocking
	BlockThreshold uint32 `json:"block_threshold"`

	// AllowedCountries is the list of allowed country codes (empty = all)
	AllowedCountries []string `json:"allowed_countries,omitempty"`

	// BlockedCountries is the list of blocked country codes
	BlockedCountries []string `json:"blocked_countries,omitempty"`
}

// DefaultVoIPDetectorConfig returns the default VoIP detector configuration
func DefaultVoIPDetectorConfig() VoIPDetectorConfig {
	return VoIPDetectorConfig{
		Provider:           "numverify",
		Enabled:            true,
		BlockVoIP:          true,
		CacheTTLSeconds:    86400, // 24 hours
		RateLimitPerMinute: 60,
		HighRiskThreshold:  60,
		BlockThreshold:     80,
		AllowedCountries:   []string{}, // All allowed
		BlockedCountries:   []string{}, // None blocked
	}
}

// Validate validates the VoIP detector configuration
func (c *VoIPDetectorConfig) Validate() error {
	if c.Enabled && c.Provider == "" {
		return ErrInvalidPhone.Wrap("provider cannot be empty when enabled")
	}
	if c.CacheTTLSeconds < 0 {
		return ErrInvalidPhone.Wrap("cache_ttl_seconds cannot be negative")
	}
	if c.RateLimitPerMinute == 0 {
		return ErrInvalidPhone.Wrap("rate_limit_per_minute must be positive")
	}
	if c.BlockThreshold < c.HighRiskThreshold {
		return ErrInvalidPhone.Wrap("block_threshold must be >= high_risk_threshold")
	}
	return nil
}

// IsCountryAllowed checks if a country code is allowed
func (c *VoIPDetectorConfig) IsCountryAllowed(countryCode string) bool {
	// Check blocked list first
	for _, blocked := range c.BlockedCountries {
		if blocked == countryCode {
			return false
		}
	}

	// If allowed list is empty, all countries are allowed
	if len(c.AllowedCountries) == 0 {
		return true
	}

	// Check allowed list
	for _, allowed := range c.AllowedCountries {
		if allowed == countryCode {
			return true
		}
	}

	return false
}

// MockVoIPDetector is a mock implementation for testing
type MockVoIPDetector struct {
	providerName string
	available    bool
	rateLimit    uint32
	cacheTTL     int64
	results      map[string]*CarrierLookupResult
}

// NewMockVoIPDetector creates a new mock VoIP detector
func NewMockVoIPDetector() *MockVoIPDetector {
	return &MockVoIPDetector{
		providerName: "mock",
		available:    true,
		rateLimit:    100,
		cacheTTL:     3600,
		results:      make(map[string]*CarrierLookupResult),
	}
}

// SetResult sets a mock result for a phone hash
func (m *MockVoIPDetector) SetResult(phoneHashRef string, result *CarrierLookupResult) {
	m.results[phoneHashRef] = result
}

// LookupCarrier performs a mock carrier lookup
func (m *MockVoIPDetector) LookupCarrier(phoneNumber string, phoneHashRef string) (*CarrierLookupResult, error) {
	if result, ok := m.results[phoneHashRef]; ok {
		return result, nil
	}

	// Default to mobile carrier
	return &CarrierLookupResult{
		PhoneHashRef:    phoneHashRef,
		CountryCode:     "US",
		CarrierName:     "Mock Carrier",
		CarrierType:     CarrierTypeMobile,
		IsVoIP:          false,
		IsPorted:        false,
		IsRoaming:       false,
		RiskLevel:       VoIPRiskLow,
		RiskScore:       10,
		LookupProvider:  m.providerName,
		LookupTimestamp: time.Unix(0, 0),
		Success:         true,
	}, nil
}

// GetProviderName returns the mock provider name
func (m *MockVoIPDetector) GetProviderName() string {
	return m.providerName
}

// IsAvailable returns if the mock is available
func (m *MockVoIPDetector) IsAvailable() bool {
	return m.available
}

// GetRateLimit returns the mock rate limit
func (m *MockVoIPDetector) GetRateLimit() uint32 {
	return m.rateLimit
}

// GetCacheTTL returns the mock cache TTL
func (m *MockVoIPDetector) GetCacheTTL() int64 {
	return m.cacheTTL
}
