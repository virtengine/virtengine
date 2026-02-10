// Package sms provides VoIP detection for SMS verification fraud prevention.
//
// This file implements VoIP detection with carrier lookup integration,
// pattern matching, and caching to identify virtual/disposable phone numbers.
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/security"
)

// ============================================================================
// VoIP Detector Interface
// ============================================================================

// VoIPDetector defines the interface for detecting VoIP and virtual numbers
type VoIPDetector interface {
	// Detect performs VoIP detection on a phone number
	Detect(ctx context.Context, phoneNumber string) (*VoIPDetectionResult, error)

	// IsVoIP returns true if the phone number is likely VoIP
	IsVoIP(ctx context.Context, phoneNumber string) (bool, error)

	// GetRiskScore returns a risk score for the phone number (0-100)
	GetRiskScore(ctx context.Context, phoneNumber string) (uint32, error)

	// Close closes the detector and releases resources
	Close() error
}

// VoIPDetectionResult contains the result of VoIP detection
type VoIPDetectionResult struct {
	// PhoneNumber is the analyzed phone number (masked)
	PhoneNumber string `json:"phone_number"`

	// PhoneHash is the hash of the phone number
	PhoneHash string `json:"phone_hash"`

	// IsVoIP indicates if the number is likely a VoIP number
	IsVoIP bool `json:"is_voip"`

	// IsVirtual indicates if this is a virtual number
	IsVirtual bool `json:"is_virtual"`

	// IsDisposable indicates if this is a disposable number
	IsDisposable bool `json:"is_disposable"`

	// IsPrepaid indicates if this is a prepaid number
	IsPrepaid bool `json:"is_prepaid"`

	// CarrierType is the type of carrier
	CarrierType CarrierType `json:"carrier_type"`

	// CarrierName is the name of the carrier
	CarrierName string `json:"carrier_name,omitempty"`

	// CountryCode is the ISO country code
	CountryCode string `json:"country_code"`

	// RiskScore is the overall risk score (0-100)
	RiskScore uint32 `json:"risk_score"`

	// RiskLevel is the risk level (low, medium, high, critical)
	RiskLevel RiskLevel `json:"risk_level"`

	// RiskFactors lists detected risk factors
	RiskFactors []string `json:"risk_factors,omitempty"`

	// ShouldBlock indicates if this number should be blocked
	ShouldBlock bool `json:"should_block"`

	// BlockReason is the reason for blocking (if applicable)
	BlockReason string `json:"block_reason,omitempty"`

	// DetectionTimestamp is when detection was performed
	DetectionTimestamp time.Time `json:"detection_timestamp"`

	// IsCached indicates if this result was from cache
	IsCached bool `json:"is_cached"`

	// Provider is the detection provider used
	Provider string `json:"provider"`
}

// ============================================================================
// VoIP Detection Configuration
// ============================================================================

// VoIPDetectorConfig contains configuration for VoIP detection
type VoIPDetectorConfig struct {
	// Enabled indicates if VoIP detection is enabled
	Enabled bool `json:"enabled"`

	// BlockVoIP indicates if VoIP numbers should be blocked
	BlockVoIP bool `json:"block_voip"`

	// BlockVirtual indicates if virtual numbers should be blocked
	BlockVirtual bool `json:"block_virtual"`

	// BlockDisposable indicates if disposable numbers should be blocked
	BlockDisposable bool `json:"block_disposable"`

	// RiskThreshold is the risk score threshold for blocking (0-100)
	RiskThreshold uint32 `json:"risk_threshold"`

	// CacheTTLSeconds is how long to cache detection results
	CacheTTLSeconds int64 `json:"cache_ttl_seconds"`

	// UseCarrierLookup indicates if carrier lookup should be used
	UseCarrierLookup bool `json:"use_carrier_lookup"`

	// UsePatternMatching indicates if pattern matching should be used
	UsePatternMatching bool `json:"use_pattern_matching"`

	// Provider is the carrier lookup provider (twilio, vonage, etc.)
	Provider string `json:"provider"`

	// ProviderConfig is the provider-specific configuration
	ProviderConfig ProviderConfig `json:"provider_config"`

	// AllowedCountries is the list of allowed country codes (empty = all)
	AllowedCountries []string `json:"allowed_countries,omitempty"`

	// BlockedCountries is the list of blocked country codes
	BlockedCountries []string `json:"blocked_countries,omitempty"`

	// CustomVoIPPatterns are additional VoIP carrier patterns to detect
	CustomVoIPPatterns []string `json:"custom_voip_patterns,omitempty"`
}

// DefaultVoIPDetectorConfig returns the default configuration
func DefaultVoIPDetectorConfig() VoIPDetectorConfig {
	return VoIPDetectorConfig{
		Enabled:            true,
		BlockVoIP:          true,
		BlockVirtual:       true,
		BlockDisposable:    true,
		RiskThreshold:      70,
		CacheTTLSeconds:    86400, // 24 hours
		UseCarrierLookup:   true,
		UsePatternMatching: true,
		Provider:           "twilio",
	}
}

// ============================================================================
// Default VoIP Detector Implementation
// ============================================================================

// DefaultVoIPDetector implements VoIPDetector with carrier lookup and pattern matching
type DefaultVoIPDetector struct {
	config   VoIPDetectorConfig
	gateway  SMSGateway
	logger   zerolog.Logger
	cache    *voipCache
	patterns *voipPatterns
	mu       sync.RWMutex
}

// voipCache is a simple in-memory cache for VoIP detection results
type voipCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	result    *VoIPDetectionResult
	expiresAt time.Time
}

func newVoIPCache(ttl time.Duration) *voipCache {
	c := &voipCache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
	// Start cleanup goroutine
	go c.cleanupLoop()
	return c
}

func (c *voipCache) get(key string) (*VoIPDetectionResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	result := *entry.result
	result.IsCached = true
	return &result, true
}

func (c *voipCache) set(key string, result *VoIPDetectionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *voipCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *voipCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// voipPatterns contains compiled regex patterns for VoIP detection
type voipPatterns struct {
	knownVoIPCarriers     []string
	disposablePatterns    []*regexp.Regexp
	virtualNumberPrefixes map[string][]string // Country code -> prefixes
}

func newVoIPPatterns(customPatterns []string) *voipPatterns {
	p := &voipPatterns{
		knownVoIPCarriers: []string{
			"google voice",
			"google fi",
			"bandwidth",
			"bandwidth.com",
			"twilio",
			"nexmo",
			"vonage",
			"plivo",
			"sinch",
			"messagebird",
			"textfree",
			"textnow",
			"pinger",
			"burner",
			"hushed",
			"talkatone",
			"sideline",
			"2ndline",
			"line2",
			"dingtone",
			"textplus",
			"freetone",
			"nextplus",
			"textnow",
			"grasshopper",
			"ringcentral",
			"vonage business",
			"8x8",
			"ooma",
			"magicjack",
			"skype",
			"whatsapp",
			"signal",
			"telegram",
		},
		virtualNumberPrefixes: map[string][]string{
			"US": {"900", "976", "844", "855", "866", "877", "888"},
			"CA": {"900", "976"},
			"GB": {"070", "076"},
			"AU": {"1900"},
		},
	}

	// Add custom patterns
	p.knownVoIPCarriers = append(p.knownVoIPCarriers, customPatterns...)

	// Compile disposable number patterns
	disposablePatterns := []string{
		`(?i)temp\s*phone`,
		`(?i)disposable`,
		`(?i)virtual\s*number`,
		`(?i)sms\s*receive`,
		`(?i)free\s*sms`,
	}
	for _, pattern := range disposablePatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			p.disposablePatterns = append(p.disposablePatterns, re)
		}
	}

	return p
}

func (p *voipPatterns) isKnownVoIPCarrier(carrierName string) bool {
	if carrierName == "" {
		return false
	}
	carrierLower := strings.ToLower(carrierName)
	for _, known := range p.knownVoIPCarriers {
		if strings.Contains(carrierLower, strings.ToLower(known)) {
			return true
		}
	}
	return false
}

func (p *voipPatterns) isVirtualPrefix(countryCode, phoneNumber string) bool {
	prefixes, ok := p.virtualNumberPrefixes[countryCode]
	if !ok {
		return false
	}

	// Extract national number (simplified)
	number := phoneNumber
	if strings.HasPrefix(number, "+") {
		// Remove country code
		number = number[len(countryCode)+1:]
	}
	number = strings.TrimLeft(number, "0")

	for _, prefix := range prefixes {
		if strings.HasPrefix(number, prefix) {
			return true
		}
	}
	return false
}

// NewVoIPDetector creates a new VoIP detector
func NewVoIPDetector(
	config VoIPDetectorConfig,
	gateway SMSGateway,
	logger zerolog.Logger,
) (*DefaultVoIPDetector, error) {
	if !config.Enabled {
		return nil, errors.Wrap(ErrInvalidConfig, "VoIP detection is disabled")
	}

	cacheTTL := time.Duration(config.CacheTTLSeconds) * time.Second
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour
	}

	return &DefaultVoIPDetector{
		config:   config,
		gateway:  gateway,
		logger:   logger.With().Str("component", "voip_detector").Logger(),
		cache:    newVoIPCache(cacheTTL),
		patterns: newVoIPPatterns(config.CustomVoIPPatterns),
	}, nil
}

// Detect performs comprehensive VoIP detection
func (d *DefaultVoIPDetector) Detect(ctx context.Context, phoneNumber string) (*VoIPDetectionResult, error) {
	startTime := time.Now()

	// Normalize and hash phone number
	phoneHash := HashPhoneNumber(phoneNumber)

	// Check cache first
	if cached, ok := d.cache.get(phoneHash); ok {
		d.logger.Debug().
			Str("phone", MaskPhoneNumber(phoneNumber)).
			Msg("VoIP detection cache hit")
		return cached, nil
	}

	result := &VoIPDetectionResult{
		PhoneNumber:        MaskPhoneNumber(phoneNumber),
		PhoneHash:          phoneHash,
		DetectionTimestamp: time.Now(),
		RiskFactors:        make([]string, 0),
	}

	// Extract country code
	result.CountryCode = extractCountryCode(phoneNumber)

	// Check if country is allowed
	if !d.isCountryAllowed(result.CountryCode) {
		result.ShouldBlock = true
		result.BlockReason = "Country not allowed"
		result.RiskScore = 100
		result.RiskLevel = RiskLevelCritical
		result.RiskFactors = append(result.RiskFactors, "blocked_country")
		return result, nil
	}

	// Perform carrier lookup if enabled and gateway is available
	if d.config.UseCarrierLookup && d.gateway != nil {
		carrierResult, err := d.gateway.LookupCarrier(ctx, phoneNumber)
		if err != nil {
			d.logger.Warn().Err(err).Str("phone", MaskPhoneNumber(phoneNumber)).Msg("carrier lookup failed")
		} else {
			result.CarrierType = carrierResult.CarrierType
			result.CarrierName = carrierResult.CarrierName
			result.IsVoIP = carrierResult.IsVoIP
			result.IsPrepaid = carrierResult.IsPrepaid
			result.Provider = d.gateway.Name()

			if carrierResult.IsVoIP {
				result.RiskFactors = append(result.RiskFactors, "carrier_voip")
			}
			if carrierResult.IsPrepaid {
				result.RiskFactors = append(result.RiskFactors, "prepaid_number")
			}
		}
	}

	// Perform pattern matching if enabled
	if d.config.UsePatternMatching {
		d.applyPatternMatching(result, phoneNumber)
	}

	// Calculate final risk score
	d.calculateRiskScore(result)

	// Determine if should block
	d.determineBlocking(result)

	// Cache the result
	d.cache.set(phoneHash, result)

	d.logger.Info().
		Str("phone", MaskPhoneNumber(phoneNumber)).
		Bool("is_voip", result.IsVoIP).
		Uint32("risk_score", result.RiskScore).
		Bool("should_block", result.ShouldBlock).
		Dur("latency", time.Since(startTime)).
		Msg("VoIP detection completed")

	return result, nil
}

// applyPatternMatching applies pattern-based VoIP detection
func (d *DefaultVoIPDetector) applyPatternMatching(result *VoIPDetectionResult, phoneNumber string) {
	// Check carrier name against known VoIP carriers
	if result.CarrierName != "" && d.patterns.isKnownVoIPCarrier(result.CarrierName) {
		result.IsVoIP = true
		result.RiskFactors = append(result.RiskFactors, "known_voip_carrier")
	}

	// Check for virtual number prefixes
	if d.patterns.isVirtualPrefix(result.CountryCode, phoneNumber) {
		result.IsVirtual = true
		result.RiskFactors = append(result.RiskFactors, "virtual_prefix")
	}

	// Check carrier type
	if result.CarrierType == CarrierTypeVoIP {
		result.IsVoIP = true
		if !containsRiskFactor(result.RiskFactors, "carrier_voip") {
			result.RiskFactors = append(result.RiskFactors, "carrier_voip")
		}
	}
}

// calculateRiskScore calculates the overall risk score
func (d *DefaultVoIPDetector) calculateRiskScore(result *VoIPDetectionResult) {
	var score uint32 = 0

	// Base risk factors
	if result.IsVoIP {
		score += 50
	}
	if result.IsVirtual {
		score += 40
	}
	if result.IsDisposable {
		score += 60
	}
	if result.IsPrepaid {
		score += 15
	}

	// Carrier type risk
	switch result.CarrierType {
	case CarrierTypeVoIP:
		score += 30
	case CarrierTypeLandline:
		score += 10
	case CarrierTypeUnknown:
		score += 20
	}

	// Risk factor accumulation
	score += safeUint32FromInt(len(result.RiskFactors) * 5)

	// Cap at 100
	if score > 100 {
		score = 100
	}

	result.RiskScore = score

	// Determine risk level
	switch {
	case score >= 80:
		result.RiskLevel = RiskLevelCritical
	case score >= 50:
		result.RiskLevel = RiskLevelHigh
	case score >= 25:
		result.RiskLevel = RiskLevelMedium
	default:
		result.RiskLevel = RiskLevelLow
	}
}

// determineBlocking determines if the number should be blocked
func (d *DefaultVoIPDetector) determineBlocking(result *VoIPDetectionResult) {
	// Check VoIP blocking
	if d.config.BlockVoIP && result.IsVoIP {
		result.ShouldBlock = true
		result.BlockReason = "VoIP numbers are not allowed"
		return
	}

	// Check virtual number blocking
	if d.config.BlockVirtual && result.IsVirtual {
		result.ShouldBlock = true
		result.BlockReason = "Virtual numbers are not allowed"
		return
	}

	// Check disposable number blocking
	if d.config.BlockDisposable && result.IsDisposable {
		result.ShouldBlock = true
		result.BlockReason = "Disposable numbers are not allowed"
		return
	}

	// Check risk threshold
	if result.RiskScore >= d.config.RiskThreshold {
		result.ShouldBlock = true
		result.BlockReason = fmt.Sprintf("Risk score %d exceeds threshold %d", result.RiskScore, d.config.RiskThreshold)
		return
	}
}

// isCountryAllowed checks if a country code is allowed
func (d *DefaultVoIPDetector) isCountryAllowed(countryCode string) bool {
	// Check blocked list first
	for _, blocked := range d.config.BlockedCountries {
		if strings.EqualFold(blocked, countryCode) {
			return false
		}
	}

	// If allowed list is empty, all non-blocked countries are allowed
	if len(d.config.AllowedCountries) == 0 {
		return true
	}

	// Check allowed list
	for _, allowed := range d.config.AllowedCountries {
		if strings.EqualFold(allowed, countryCode) {
			return true
		}
	}

	return false
}

// IsVoIP returns true if the phone number is likely VoIP
func (d *DefaultVoIPDetector) IsVoIP(ctx context.Context, phoneNumber string) (bool, error) {
	result, err := d.Detect(ctx, phoneNumber)
	if err != nil {
		return false, err
	}
	return result.IsVoIP, nil
}

// GetRiskScore returns the risk score for a phone number
func (d *DefaultVoIPDetector) GetRiskScore(ctx context.Context, phoneNumber string) (uint32, error) {
	result, err := d.Detect(ctx, phoneNumber)
	if err != nil {
		return 0, err
	}
	return result.RiskScore, nil
}

// Close closes the detector
func (d *DefaultVoIPDetector) Close() error {
	// Clean up cache
	d.mu.Lock()
	defer d.mu.Unlock()
	return nil
}

// Ensure DefaultVoIPDetector implements VoIPDetector
var _ VoIPDetector = (*DefaultVoIPDetector)(nil)

// ============================================================================
// NumVerify VoIP Detector
// ============================================================================

const numVerifyBaseURL = "http://apilayer.net/api/validate"

// NumVerifyDetector implements VoIP detection using NumVerify API
type NumVerifyDetector struct {
	apiKey     string
	httpClient *http.Client
	logger     zerolog.Logger
	cache      *voipCache
}

// NewNumVerifyDetector creates a new NumVerify detector
func NewNumVerifyDetector(apiKey string, logger zerolog.Logger) (*NumVerifyDetector, error) {
	if apiKey == "" {
		return nil, errors.Wrap(ErrInvalidConfig, "NumVerify API key is required")
	}

	return &NumVerifyDetector{
		apiKey:     apiKey,
		httpClient: security.NewSecureHTTPClient(security.WithTimeout(15 * time.Second)),
		logger:     logger.With().Str("component", "numverify_detector").Logger(),
		cache:      newVoIPCache(24 * time.Hour),
	}, nil
}

// Detect performs VoIP detection using NumVerify
func (d *NumVerifyDetector) Detect(ctx context.Context, phoneNumber string) (*VoIPDetectionResult, error) {
	phoneHash := HashPhoneNumber(phoneNumber)

	// Check cache
	if cached, ok := d.cache.get(phoneHash); ok {
		return cached, nil
	}

	// Build request URL
	params := url.Values{}
	params.Set("access_key", d.apiKey)
	params.Set("number", phoneNumber)
	params.Set("format", "1")

	apiURL := numVerifyBaseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrapf(ErrCarrierLookupFailed, "failed to create request: %v", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(ErrCarrierLookupFailed, "request failed: %v", err)
	}
	defer resp.Body.Close()

	var numVerifyResp struct {
		Valid               bool   `json:"valid"`
		Number              string `json:"number"`
		LocalFormat         string `json:"local_format"`
		InternationalFormat string `json:"international_format"`
		CountryPrefix       string `json:"country_prefix"`
		CountryCode         string `json:"country_code"`
		CountryName         string `json:"country_name"`
		Location            string `json:"location"`
		Carrier             string `json:"carrier"`
		LineType            string `json:"line_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&numVerifyResp); err != nil {
		return nil, errors.Wrapf(ErrCarrierLookupFailed, "failed to parse response: %v", err)
	}

	result := &VoIPDetectionResult{
		PhoneNumber:        MaskPhoneNumber(phoneNumber),
		PhoneHash:          phoneHash,
		CountryCode:        numVerifyResp.CountryCode,
		CarrierName:        numVerifyResp.Carrier,
		DetectionTimestamp: time.Now(),
		Provider:           "numverify",
		RiskFactors:        make([]string, 0),
	}

	// Parse line type
	switch strings.ToLower(numVerifyResp.LineType) {
	case "mobile":
		result.CarrierType = CarrierTypeMobile
	case "landline":
		result.CarrierType = CarrierTypeLandline
	case "voip":
		result.CarrierType = CarrierTypeVoIP
		result.IsVoIP = true
		result.RiskFactors = append(result.RiskFactors, "voip_line_type")
	case "toll_free", "special_services":
		result.CarrierType = CarrierTypeUnknown
		result.IsVirtual = true
		result.RiskFactors = append(result.RiskFactors, "special_number")
	default:
		result.CarrierType = CarrierTypeUnknown
	}

	// Calculate risk score
	result.RiskScore = calculateRiskFromResult(result)

	// Determine risk level
	switch {
	case result.RiskScore >= 80:
		result.RiskLevel = RiskLevelCritical
	case result.RiskScore >= 50:
		result.RiskLevel = RiskLevelHigh
	case result.RiskScore >= 25:
		result.RiskLevel = RiskLevelMedium
	default:
		result.RiskLevel = RiskLevelLow
	}

	// Cache result
	d.cache.set(phoneHash, result)

	return result, nil
}

// IsVoIP returns true if the phone number is likely VoIP
func (d *NumVerifyDetector) IsVoIP(ctx context.Context, phoneNumber string) (bool, error) {
	result, err := d.Detect(ctx, phoneNumber)
	if err != nil {
		return false, err
	}
	return result.IsVoIP, nil
}

// GetRiskScore returns the risk score
func (d *NumVerifyDetector) GetRiskScore(ctx context.Context, phoneNumber string) (uint32, error) {
	result, err := d.Detect(ctx, phoneNumber)
	if err != nil {
		return 0, err
	}
	return result.RiskScore, nil
}

// Close closes the detector
func (d *NumVerifyDetector) Close() error {
	return nil
}

// Ensure NumVerifyDetector implements VoIPDetector
var _ VoIPDetector = (*NumVerifyDetector)(nil)

// ============================================================================
// Helper Functions
// ============================================================================

// extractCountryCode extracts the country code from a phone number
func extractCountryCode(phoneNumber string) string {
	if !strings.HasPrefix(phoneNumber, "+") {
		return "US" // Default to US
	}

	// Common country codes by prefix length
	number := strings.TrimPrefix(phoneNumber, "+")

	// Try 1-digit codes first
	if strings.HasPrefix(number, "1") {
		return "US" // or CA
	}

	// Try 2-digit codes
	twoDigit := number[:2]
	switch twoDigit {
	case "44":
		return "GB"
	case "49":
		return "DE"
	case "33":
		return "FR"
	case "39":
		return "IT"
	case "34":
		return "ES"
	case "91":
		return "IN"
	case "86":
		return "CN"
	case "81":
		return "JP"
	case "82":
		return "KR"
	case "61":
		return "AU"
	case "55":
		return "BR"
	case "52":
		return "MX"
	case "31":
		return "NL"
	case "65":
		return "SG"
	case "60":
		return "MY"
	case "63":
		return "PH"
	case "66":
		return "TH"
	case "84":
		return "VN"
	case "62":
		return "ID"
	case "27":
		return "ZA"
	case "20":
		return "EG"
	case "48":
		return "PL"
	case "47":
		return "NO"
	case "46":
		return "SE"
	case "45":
		return "DK"
	case "41":
		return "CH"
	case "43":
		return "AT"
	case "32":
		return "BE"
	case "30":
		return "GR"
	case "36":
		return "HU"
	case "42":
		return "CZ"
	case "40":
		return "RO"
	case "38":
		return "RS" // Serbia or Slovenia depending on next digit
	case "35":
		return "PT"
	case "37":
		return "LT" // or LV or EE depending on next digit
	case "98":
		return "IR"
	case "90":
		return "TR"
	case "92":
		return "PK"
	case "93":
		return "AF"
	case "94":
		return "LK"
	case "95":
		return "MM"
	case "96":
		return "IQ"
	case "97":
		return "AE" // or others
	case "99":
		return "UZ" // or others
	case "21":
		return "MA" // or DZ
	case "22":
		return "NG" // or others
	case "23":
		return "NG" // or others
	case "24":
		return "AO" // or others
	case "25":
		return "KE" // or others
	case "26":
		return "ZM" // or others
	case "67":
		return "PG"
	case "68":
		return "NZ"
	case "69":
		return "FM"
	}

	return "US" // Default fallback
}

// containsRiskFactor checks if a risk factor is already in the list
func containsRiskFactor(factors []string, factor string) bool {
	for _, f := range factors {
		if f == factor {
			return true
		}
	}
	return false
}

// calculateRiskFromResult calculates risk score from detection result
func calculateRiskFromResult(result *VoIPDetectionResult) uint32 {
	var score uint32 = 0

	if result.IsVoIP {
		score += 60
	}
	if result.IsVirtual {
		score += 40
	}
	if result.IsDisposable {
		score += 70
	}

	switch result.CarrierType {
	case CarrierTypeVoIP:
		score += 20
	case CarrierTypeUnknown:
		score += 15
	case CarrierTypeLandline:
		score += 5
	}

	score += safeUint32FromInt(len(result.RiskFactors) * 5)

	if score > 100 {
		score = 100
	}

	return score
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > int(^uint32(0)) {
		return ^uint32(0)
	}
	//nolint:gosec // range checked above
	return uint32(value)
}
