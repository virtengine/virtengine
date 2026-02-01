// Package sms provides anti-fraud controls for SMS verification.
//
// This file implements:
// - VoIP detection and blocking
// - Phone number blocklist management
// - Velocity checks (per phone, per IP, per account)
// - Device fingerprint tracking
// - Risk scoring
//
// Task Reference: VE-4C - SMS Verification Delivery + Anti-Fraud
package sms

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/errors"
)

// ============================================================================
// Anti-Fraud Configuration
// ============================================================================

// AntiFraudConfig contains anti-fraud configuration
type AntiFraudConfig struct {
	// EnableVoIPBlocking blocks VoIP numbers
	EnableVoIPBlocking bool `json:"enable_voip_blocking"`

	// EnableVelocityChecks enables velocity-based rate limiting
	EnableVelocityChecks bool `json:"enable_velocity_checks"`

	// EnableDeviceTracking tracks device fingerprints
	EnableDeviceTracking bool `json:"enable_device_tracking"`

	// EnableRiskScoring enables risk scoring
	EnableRiskScoring bool `json:"enable_risk_scoring"`

	// MaxRequestsPerPhonePerHour is the max SMS requests per phone per hour
	MaxRequestsPerPhonePerHour int `json:"max_requests_per_phone_per_hour"`

	// MaxRequestsPerIPPerHour is the max SMS requests per IP per hour
	MaxRequestsPerIPPerHour int `json:"max_requests_per_ip_per_hour"`

	// MaxRequestsPerAccountPerDay is the max SMS requests per account per day
	MaxRequestsPerAccountPerDay int `json:"max_requests_per_account_per_day"`

	// MaxAccountsPerDevicePerDay limits accounts per device per day
	MaxAccountsPerDevicePerDay int `json:"max_accounts_per_device_per_day"`

	// RiskScoreThreshold is the threshold for blocking (0-100)
	RiskScoreThreshold uint32 `json:"risk_score_threshold"`

	// BlockedCountryCodes is a list of blocked country codes
	BlockedCountryCodes []string `json:"blocked_country_codes,omitempty"`

	// BlockedCarriers is a list of blocked carrier names/patterns
	BlockedCarriers []string `json:"blocked_carriers,omitempty"`

	// VoIPCarrierPatterns is a list of known VoIP carrier patterns
	VoIPCarrierPatterns []string `json:"voip_carrier_patterns,omitempty"`

	// SuspiciousIPRanges is a list of suspicious IP ranges (CIDR)
	SuspiciousIPRanges []string `json:"suspicious_ip_ranges,omitempty"`

	// BlockDurationMinutes is how long to block after fraud detection
	BlockDurationMinutes int `json:"block_duration_minutes"`

	// RedisURL is the Redis connection URL for state storage
	RedisURL string `json:"redis_url"`

	// KeyPrefix is the Redis key prefix
	KeyPrefix string `json:"key_prefix"`
}

// DefaultAntiFraudConfig returns the default anti-fraud configuration
func DefaultAntiFraudConfig() AntiFraudConfig {
	return AntiFraudConfig{
		EnableVoIPBlocking:          true,
		EnableVelocityChecks:        true,
		EnableDeviceTracking:        true,
		EnableRiskScoring:           true,
		MaxRequestsPerPhonePerHour:  3,
		MaxRequestsPerIPPerHour:     10,
		MaxRequestsPerAccountPerDay: 10,
		MaxAccountsPerDevicePerDay:  3,
		RiskScoreThreshold:          70,
		BlockDurationMinutes:        1440, // 24 hours
		KeyPrefix:                   "sms_antifraud",
		VoIPCarrierPatterns: []string{
			"google voice",
			"bandwidth",
			"twilio",
			"nexmo",
			"plivo",
			"sinch",
			"vonage",
			"textfree",
			"textnow",
			"pinger",
			"burner",
			"hushed",
		},
	}
}

// ============================================================================
// Anti-Fraud Engine Interface
// ============================================================================

// AntiFraudEngine defines the interface for anti-fraud checks
type AntiFraudEngine interface {
	// CheckPhone performs anti-fraud checks on a phone number
	CheckPhone(ctx context.Context, req *AntiFraudRequest) (*AntiFraudResult, error)

	// RecordRequest records a request for velocity tracking
	RecordRequest(ctx context.Context, req *AntiFraudRequest) error

	// BlockPhone blocks a phone number
	BlockPhone(ctx context.Context, phoneHash string, reason string, duration time.Duration) error

	// IsPhoneBlocked checks if a phone is blocked
	IsPhoneBlocked(ctx context.Context, phoneHash string) (bool, string, error)

	// BlockIP blocks an IP address
	BlockIP(ctx context.Context, ipHash string, reason string, duration time.Duration) error

	// IsIPBlocked checks if an IP is blocked
	IsIPBlocked(ctx context.Context, ipHash string) (bool, string, error)

	// GetVelocityStats returns velocity statistics for an entity
	GetVelocityStats(ctx context.Context, entityType string, entityHash string) (*VelocityStats, error)

	// Close closes the engine
	Close() error
}

// AntiFraudRequest contains information for anti-fraud checks
type AntiFraudRequest struct {
	// AccountAddress is the requesting account
	AccountAddress string `json:"account_address"`

	// PhoneNumber is the phone number (E.164)
	PhoneNumber string `json:"phone_number"`

	// PhoneHash is the hash of the phone number
	PhoneHash string `json:"phone_hash"`

	// CountryCode is the country code
	CountryCode string `json:"country_code"`

	// IPAddress is the client IP address
	IPAddress string `json:"ip_address"`

	// IPHash is the hash of the IP address
	IPHash string `json:"ip_hash"`

	// DeviceFingerprint is a hash of device identifiers
	DeviceFingerprint string `json:"device_fingerprint"`

	// UserAgent is the user agent string
	UserAgent string `json:"user_agent"`

	// CarrierInfo is the carrier lookup result (if available)
	CarrierInfo *CarrierLookupResult `json:"carrier_info,omitempty"`

	// Timestamp is when the request was made
	Timestamp time.Time `json:"timestamp"`
}

// AntiFraudResult contains the result of anti-fraud checks
type AntiFraudResult struct {
	// Allowed indicates if the request is allowed
	Allowed bool `json:"allowed"`

	// RiskScore is the overall risk score (0-100)
	RiskScore uint32 `json:"risk_score"`

	// RiskLevel is the risk level (low, medium, high, critical)
	RiskLevel RiskLevel `json:"risk_level"`

	// RiskFactors lists detected risk factors
	RiskFactors []RiskFactor `json:"risk_factors"`

	// BlockReason is the reason if blocked
	BlockReason string `json:"block_reason,omitempty"`

	// RetryAfter is when the request can be retried (if rate limited)
	RetryAfter *time.Time `json:"retry_after,omitempty"`

	// Recommendations contains recommended actions
	Recommendations []string `json:"recommendations,omitempty"`
}

// RiskLevel represents the risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RiskFactor represents a detected risk factor
type RiskFactor struct {
	// Type is the type of risk factor
	Type string `json:"type"`

	// Description is a human-readable description
	Description string `json:"description"`

	// Score is the score contribution (0-100)
	Score uint32 `json:"score"`

	// Severity is the severity level
	Severity string `json:"severity"`
}

// VelocityStats contains velocity statistics for an entity
type VelocityStats struct {
	// EntityType is the type of entity (phone, ip, account)
	EntityType string `json:"entity_type"`

	// EntityHash is the hash of the entity
	EntityHash string `json:"entity_hash"`

	// RequestsLastHour is the number of requests in the last hour
	RequestsLastHour int `json:"requests_last_hour"`

	// RequestsLastDay is the number of requests in the last day
	RequestsLastDay int `json:"requests_last_day"`

	// UniqueAccountsLastDay is unique accounts for this entity (phone/IP)
	UniqueAccountsLastDay int `json:"unique_accounts_last_day"`

	// LastRequestAt is when the last request was made
	LastRequestAt *time.Time `json:"last_request_at,omitempty"`

	// IsBlocked indicates if this entity is blocked
	IsBlocked bool `json:"is_blocked"`

	// BlockedUntil is when the block expires
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`

	// BlockReason is why this entity is blocked
	BlockReason string `json:"block_reason,omitempty"`
}

// ============================================================================
// Redis-Based Anti-Fraud Engine
// ============================================================================

// RedisAntiFraudEngine implements AntiFraudEngine using Redis
type RedisAntiFraudEngine struct {
	config AntiFraudConfig
	client *redis.Client
	logger zerolog.Logger
	mu     sync.RWMutex
}

// NewRedisAntiFraudEngine creates a new Redis-based anti-fraud engine
func NewRedisAntiFraudEngine(ctx context.Context, config AntiFraudConfig, logger zerolog.Logger) (*RedisAntiFraudEngine, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidConfig, "invalid redis URL: %v", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrapf(ErrCacheError, "failed to connect to redis: %v", err)
	}

	return &RedisAntiFraudEngine{
		config: config,
		client: client,
		logger: logger.With().Str("component", "antifraud").Logger(),
	}, nil
}

// CheckPhone performs comprehensive anti-fraud checks
func (e *RedisAntiFraudEngine) CheckPhone(ctx context.Context, req *AntiFraudRequest) (*AntiFraudResult, error) {
	result := &AntiFraudResult{
		Allowed:     true,
		RiskScore:   0,
		RiskLevel:   RiskLevelLow,
		RiskFactors: make([]RiskFactor, 0),
	}

	// 1. Check if phone is blocked
	if blocked, reason, err := e.IsPhoneBlocked(ctx, req.PhoneHash); err == nil && blocked {
		result.Allowed = false
		result.BlockReason = reason
		result.RiskLevel = RiskLevelCritical
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "phone_blocked",
			Description: reason,
			Score:       100,
			Severity:    "critical",
		})
		return result, nil
	}

	// 2. Check if IP is blocked
	if req.IPHash != "" {
		if blocked, reason, err := e.IsIPBlocked(ctx, req.IPHash); err == nil && blocked {
			result.Allowed = false
			result.BlockReason = reason
			result.RiskLevel = RiskLevelCritical
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "ip_blocked",
				Description: reason,
				Score:       100,
				Severity:    "critical",
			})
			return result, nil
		}
	}

	// 3. Check VoIP status
	if e.config.EnableVoIPBlocking && req.CarrierInfo != nil {
		if req.CarrierInfo.IsVoIP || e.isKnownVoIPCarrier(req.CarrierInfo.CarrierName) {
			result.Allowed = false
			result.BlockReason = "VoIP numbers are not allowed"
			result.RiskLevel = RiskLevelHigh
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "voip_detected",
				Description: fmt.Sprintf("VoIP carrier detected: %s", req.CarrierInfo.CarrierName),
				Score:       80,
				Severity:    "high",
			})
			return result, nil
		}
	}

	// 4. Check country code
	if e.isBlockedCountry(req.CountryCode) {
		result.Allowed = false
		result.BlockReason = "Country not supported"
		result.RiskLevel = RiskLevelHigh
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "blocked_country",
			Description: fmt.Sprintf("Country code %s is blocked", req.CountryCode),
			Score:       90,
			Severity:    "high",
		})
		return result, nil
	}

	// 5. Velocity checks
	if e.config.EnableVelocityChecks {
		velocityResult := e.checkVelocity(ctx, req)
		result.RiskFactors = append(result.RiskFactors, velocityResult.RiskFactors...)
		result.RiskScore += velocityResult.RiskScore
		if !velocityResult.Allowed {
			result.Allowed = false
			result.BlockReason = velocityResult.BlockReason
			result.RetryAfter = velocityResult.RetryAfter
		}
	}

	// 6. Device tracking
	if e.config.EnableDeviceTracking && req.DeviceFingerprint != "" {
		deviceResult := e.checkDevice(ctx, req)
		result.RiskFactors = append(result.RiskFactors, deviceResult.RiskFactors...)
		result.RiskScore += deviceResult.RiskScore
		if !deviceResult.Allowed {
			result.Allowed = false
			result.BlockReason = deviceResult.BlockReason
		}
	}

	// 7. Apply carrier risk score
	if req.CarrierInfo != nil {
		result.RiskScore += req.CarrierInfo.RiskScore / 2 // Weight carrier risk at 50%
		if req.CarrierInfo.RiskScore > 50 {
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "carrier_risk",
				Description: fmt.Sprintf("Carrier risk score: %d", req.CarrierInfo.RiskScore),
				Score:       req.CarrierInfo.RiskScore / 2,
				Severity:    "medium",
			})
		}
	}

	// 8. Calculate final risk level
	result.RiskLevel = e.calculateRiskLevel(result.RiskScore)

	// 9. Block if risk score exceeds threshold
	if result.RiskScore >= e.config.RiskScoreThreshold {
		result.Allowed = false
		result.BlockReason = fmt.Sprintf("Risk score %d exceeds threshold %d", result.RiskScore, e.config.RiskScoreThreshold)
	}

	return result, nil
}

// RecordRequest records a request for velocity tracking
func (e *RedisAntiFraudEngine) RecordRequest(ctx context.Context, req *AntiFraudRequest) error {
	now := time.Now()
	pipe := e.client.Pipeline()

	// Record phone request
	phoneKey := fmt.Sprintf("%s:velocity:phone:%s", e.config.KeyPrefix, req.PhoneHash)
	pipe.ZAdd(ctx, phoneKey, redis.Z{Score: float64(now.Unix()), Member: req.AccountAddress})
	pipe.Expire(ctx, phoneKey, 25*time.Hour)

	// Record IP request
	if req.IPHash != "" {
		ipKey := fmt.Sprintf("%s:velocity:ip:%s", e.config.KeyPrefix, req.IPHash)
		pipe.ZAdd(ctx, ipKey, redis.Z{Score: float64(now.Unix()), Member: req.AccountAddress})
		pipe.Expire(ctx, ipKey, 25*time.Hour)
	}

	// Record account request
	accountKey := fmt.Sprintf("%s:velocity:account:%s", e.config.KeyPrefix, hashString(req.AccountAddress))
	pipe.ZAdd(ctx, accountKey, redis.Z{Score: float64(now.Unix()), Member: req.PhoneHash})
	pipe.Expire(ctx, accountKey, 25*time.Hour)

	// Record device to account mapping
	if req.DeviceFingerprint != "" {
		deviceKey := fmt.Sprintf("%s:device:accounts:%s", e.config.KeyPrefix, req.DeviceFingerprint)
		pipe.SAdd(ctx, deviceKey, req.AccountAddress)
		pipe.Expire(ctx, deviceKey, 25*time.Hour)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BlockPhone blocks a phone number
func (e *RedisAntiFraudEngine) BlockPhone(ctx context.Context, phoneHash string, reason string, duration time.Duration) error {
	key := fmt.Sprintf("%s:blocked:phone:%s", e.config.KeyPrefix, phoneHash)
	data := map[string]interface{}{
		"reason":     reason,
		"blocked_at": time.Now().Unix(),
		"expires_at": time.Now().Add(duration).Unix(),
	}
	jsonData, _ := json.Marshal(data)
	return e.client.Set(ctx, key, jsonData, duration).Err()
}

// IsPhoneBlocked checks if a phone is blocked
func (e *RedisAntiFraudEngine) IsPhoneBlocked(ctx context.Context, phoneHash string) (bool, string, error) {
	key := fmt.Sprintf("%s:blocked:phone:%s", e.config.KeyPrefix, phoneHash)
	data, err := e.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	var blockData map[string]interface{}
	if err := json.Unmarshal(data, &blockData); err != nil {
		return true, "blocked", nil
	}

	reason, _ := blockData["reason"].(string)
	return true, reason, nil
}

// BlockIP blocks an IP address
func (e *RedisAntiFraudEngine) BlockIP(ctx context.Context, ipHash string, reason string, duration time.Duration) error {
	key := fmt.Sprintf("%s:blocked:ip:%s", e.config.KeyPrefix, ipHash)
	data := map[string]interface{}{
		"reason":     reason,
		"blocked_at": time.Now().Unix(),
		"expires_at": time.Now().Add(duration).Unix(),
	}
	jsonData, _ := json.Marshal(data)
	return e.client.Set(ctx, key, jsonData, duration).Err()
}

// IsIPBlocked checks if an IP is blocked
func (e *RedisAntiFraudEngine) IsIPBlocked(ctx context.Context, ipHash string) (bool, string, error) {
	key := fmt.Sprintf("%s:blocked:ip:%s", e.config.KeyPrefix, ipHash)
	data, err := e.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	var blockData map[string]interface{}
	if err := json.Unmarshal(data, &blockData); err != nil {
		return true, "blocked", nil
	}

	reason, _ := blockData["reason"].(string)
	return true, reason, nil
}

// GetVelocityStats returns velocity statistics for an entity
func (e *RedisAntiFraudEngine) GetVelocityStats(ctx context.Context, entityType string, entityHash string) (*VelocityStats, error) {
	key := fmt.Sprintf("%s:velocity:%s:%s", e.config.KeyPrefix, entityType, entityHash)
	now := time.Now()
	hourAgo := now.Add(-time.Hour).Unix()
	dayAgo := now.Add(-24 * time.Hour).Unix()

	// Get hourly count
	hourlyCount, err := e.client.ZCount(ctx, key, fmt.Sprintf("%d", hourAgo), "+inf").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Get daily count
	dailyCount, err := e.client.ZCount(ctx, key, fmt.Sprintf("%d", dayAgo), "+inf").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Get unique accounts
	uniqueCount, err := e.client.ZCard(ctx, key).Result()
	if err != nil && err != redis.Nil {
		uniqueCount = 0
	}

	// Get last request time
	lastRequest, err := e.client.ZRevRangeWithScores(ctx, key, 0, 0).Result()
	var lastRequestAt *time.Time
	if err == nil && len(lastRequest) > 0 {
		t := time.Unix(int64(lastRequest[0].Score), 0)
		lastRequestAt = &t
	}

	// Check if blocked
	var isBlocked bool
	var blockedUntil *time.Time
	var blockReason string
	if entityType == "phone" {
		isBlocked, blockReason, _ = e.IsPhoneBlocked(ctx, entityHash)
	} else if entityType == "ip" {
		isBlocked, blockReason, _ = e.IsIPBlocked(ctx, entityHash)
	}

	return &VelocityStats{
		EntityType:            entityType,
		EntityHash:            entityHash,
		RequestsLastHour:      int(hourlyCount),
		RequestsLastDay:       int(dailyCount),
		UniqueAccountsLastDay: int(uniqueCount),
		LastRequestAt:         lastRequestAt,
		IsBlocked:             isBlocked,
		BlockedUntil:          blockedUntil,
		BlockReason:           blockReason,
	}, nil
}

// Close closes the engine
func (e *RedisAntiFraudEngine) Close() error {
	return e.client.Close()
}

// checkVelocity performs velocity checks
func (e *RedisAntiFraudEngine) checkVelocity(ctx context.Context, req *AntiFraudRequest) *AntiFraudResult {
	result := &AntiFraudResult{
		Allowed:     true,
		RiskFactors: make([]RiskFactor, 0),
	}

	now := time.Now()
	hourAgo := now.Add(-time.Hour).Unix()
	dayAgo := now.Add(-24 * time.Hour).Unix()

	// Check phone velocity
	phoneKey := fmt.Sprintf("%s:velocity:phone:%s", e.config.KeyPrefix, req.PhoneHash)
	phoneCount, _ := e.client.ZCount(ctx, phoneKey, fmt.Sprintf("%d", hourAgo), "+inf").Result()
	if int(phoneCount) >= e.config.MaxRequestsPerPhonePerHour {
		result.Allowed = false
		result.BlockReason = "Too many requests for this phone number"
		result.RiskScore += 50
		retryAt := now.Add(time.Hour)
		result.RetryAfter = &retryAt
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "phone_velocity_exceeded",
			Description: fmt.Sprintf("Phone has %d requests in the last hour (limit: %d)", phoneCount, e.config.MaxRequestsPerPhonePerHour),
			Score:       50,
			Severity:    "high",
		})
	}

	// Check IP velocity
	if req.IPHash != "" {
		ipKey := fmt.Sprintf("%s:velocity:ip:%s", e.config.KeyPrefix, req.IPHash)
		ipCount, _ := e.client.ZCount(ctx, ipKey, fmt.Sprintf("%d", hourAgo), "+inf").Result()
		if int(ipCount) >= e.config.MaxRequestsPerIPPerHour {
			result.Allowed = false
			result.BlockReason = "Too many requests from this IP"
			result.RiskScore += 40
			retryAt := now.Add(time.Hour)
			result.RetryAfter = &retryAt
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "ip_velocity_exceeded",
				Description: fmt.Sprintf("IP has %d requests in the last hour (limit: %d)", ipCount, e.config.MaxRequestsPerIPPerHour),
				Score:       40,
				Severity:    "medium",
			})
		}
	}

	// Check account velocity
	accountKey := fmt.Sprintf("%s:velocity:account:%s", e.config.KeyPrefix, hashString(req.AccountAddress))
	accountCount, _ := e.client.ZCount(ctx, accountKey, fmt.Sprintf("%d", dayAgo), "+inf").Result()
	if int(accountCount) >= e.config.MaxRequestsPerAccountPerDay {
		result.Allowed = false
		result.BlockReason = "Too many requests for this account"
		result.RiskScore += 30
		retryAt := now.Add(24 * time.Hour)
		result.RetryAfter = &retryAt
		result.RiskFactors = append(result.RiskFactors, RiskFactor{
			Type:        "account_velocity_exceeded",
			Description: fmt.Sprintf("Account has %d requests in the last day (limit: %d)", accountCount, e.config.MaxRequestsPerAccountPerDay),
			Score:       30,
			Severity:    "medium",
		})
	}

	return result
}

// checkDevice performs device fingerprint checks
func (e *RedisAntiFraudEngine) checkDevice(ctx context.Context, req *AntiFraudRequest) *AntiFraudResult {
	result := &AntiFraudResult{
		Allowed:     true,
		RiskFactors: make([]RiskFactor, 0),
	}

	deviceKey := fmt.Sprintf("%s:device:accounts:%s", e.config.KeyPrefix, req.DeviceFingerprint)
	accountCount, err := e.client.SCard(ctx, deviceKey).Result()
	if err != nil && err != redis.Nil {
		return result
	}

	// Check if this device has been used with too many accounts
	if int(accountCount) >= e.config.MaxAccountsPerDevicePerDay {
		// Check if current account is already in the set
		isMember, _ := e.client.SIsMember(ctx, deviceKey, req.AccountAddress).Result()
		if !isMember {
			result.Allowed = false
			result.BlockReason = "Too many accounts from this device"
			result.RiskScore += 60
			result.RiskFactors = append(result.RiskFactors, RiskFactor{
				Type:        "device_account_limit",
				Description: fmt.Sprintf("Device used by %d accounts (limit: %d)", accountCount, e.config.MaxAccountsPerDevicePerDay),
				Score:       60,
				Severity:    "high",
			})
		}
	}

	return result
}

// isKnownVoIPCarrier checks if a carrier is a known VoIP provider
func (e *RedisAntiFraudEngine) isKnownVoIPCarrier(carrierName string) bool {
	if carrierName == "" {
		return false
	}
	lowerCarrier := strings.ToLower(carrierName)
	for _, pattern := range e.config.VoIPCarrierPatterns {
		if strings.Contains(lowerCarrier, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isBlockedCountry checks if a country code is blocked
func (e *RedisAntiFraudEngine) isBlockedCountry(countryCode string) bool {
	for _, blocked := range e.config.BlockedCountryCodes {
		if strings.EqualFold(blocked, countryCode) {
			return true
		}
	}
	return false
}

// calculateRiskLevel calculates the risk level from score
func (e *RedisAntiFraudEngine) calculateRiskLevel(score uint32) RiskLevel {
	switch {
	case score >= 80:
		return RiskLevelCritical
	case score >= 50:
		return RiskLevelHigh
	case score >= 25:
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

// hashString creates a SHA256 hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Ensure RedisAntiFraudEngine implements AntiFraudEngine
var _ AntiFraudEngine = (*RedisAntiFraudEngine)(nil)

// ============================================================================
// In-Memory Anti-Fraud Engine (for testing)
// ============================================================================

// InMemoryAntiFraudEngine implements AntiFraudEngine using in-memory storage
type InMemoryAntiFraudEngine struct {
	config      AntiFraudConfig
	logger      zerolog.Logger
	mu          sync.RWMutex
	phoneBlocks map[string]*blockEntry
	ipBlocks    map[string]*blockEntry
	velocity    map[string][]velocityEntry
	devices     map[string]map[string]bool
}

type blockEntry struct {
	Reason    string
	ExpiresAt time.Time
}

type velocityEntry struct {
	Account   string
	Timestamp time.Time
}

// NewInMemoryAntiFraudEngine creates a new in-memory anti-fraud engine
func NewInMemoryAntiFraudEngine(config AntiFraudConfig, logger zerolog.Logger) *InMemoryAntiFraudEngine {
	return &InMemoryAntiFraudEngine{
		config:      config,
		logger:      logger.With().Str("component", "antifraud_memory").Logger(),
		phoneBlocks: make(map[string]*blockEntry),
		ipBlocks:    make(map[string]*blockEntry),
		velocity:    make(map[string][]velocityEntry),
		devices:     make(map[string]map[string]bool),
	}
}

// CheckPhone performs anti-fraud checks
func (e *InMemoryAntiFraudEngine) CheckPhone(ctx context.Context, req *AntiFraudRequest) (*AntiFraudResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := &AntiFraudResult{
		Allowed:     true,
		RiskScore:   0,
		RiskLevel:   RiskLevelLow,
		RiskFactors: make([]RiskFactor, 0),
	}

	// Check phone block
	if block, ok := e.phoneBlocks[req.PhoneHash]; ok && time.Now().Before(block.ExpiresAt) {
		result.Allowed = false
		result.BlockReason = block.Reason
		result.RiskLevel = RiskLevelCritical
		return result, nil
	}

	// Check IP block
	if block, ok := e.ipBlocks[req.IPHash]; ok && time.Now().Before(block.ExpiresAt) {
		result.Allowed = false
		result.BlockReason = block.Reason
		result.RiskLevel = RiskLevelCritical
		return result, nil
	}

	// Check VoIP
	if e.config.EnableVoIPBlocking && req.CarrierInfo != nil && req.CarrierInfo.IsVoIP {
		result.Allowed = false
		result.BlockReason = "VoIP numbers not allowed"
		result.RiskLevel = RiskLevelHigh
		return result, nil
	}

	return result, nil
}

// RecordRequest records a request
func (e *InMemoryAntiFraudEngine) RecordRequest(ctx context.Context, req *AntiFraudRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Record phone velocity
	key := "phone:" + req.PhoneHash
	e.velocity[key] = append(e.velocity[key], velocityEntry{
		Account:   req.AccountAddress,
		Timestamp: time.Now(),
	})

	// Record IP velocity
	if req.IPHash != "" {
		key = "ip:" + req.IPHash
		e.velocity[key] = append(e.velocity[key], velocityEntry{
			Account:   req.AccountAddress,
			Timestamp: time.Now(),
		})
	}

	// Record device
	if req.DeviceFingerprint != "" {
		if e.devices[req.DeviceFingerprint] == nil {
			e.devices[req.DeviceFingerprint] = make(map[string]bool)
		}
		e.devices[req.DeviceFingerprint][req.AccountAddress] = true
	}

	return nil
}

// BlockPhone blocks a phone
func (e *InMemoryAntiFraudEngine) BlockPhone(ctx context.Context, phoneHash string, reason string, duration time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.phoneBlocks[phoneHash] = &blockEntry{
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
	return nil
}

// IsPhoneBlocked checks if phone is blocked
func (e *InMemoryAntiFraudEngine) IsPhoneBlocked(ctx context.Context, phoneHash string) (bool, string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if block, ok := e.phoneBlocks[phoneHash]; ok && time.Now().Before(block.ExpiresAt) {
		return true, block.Reason, nil
	}
	return false, "", nil
}

// BlockIP blocks an IP
func (e *InMemoryAntiFraudEngine) BlockIP(ctx context.Context, ipHash string, reason string, duration time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ipBlocks[ipHash] = &blockEntry{
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
	return nil
}

// IsIPBlocked checks if IP is blocked
func (e *InMemoryAntiFraudEngine) IsIPBlocked(ctx context.Context, ipHash string) (bool, string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if block, ok := e.ipBlocks[ipHash]; ok && time.Now().Before(block.ExpiresAt) {
		return true, block.Reason, nil
	}
	return false, "", nil
}

// GetVelocityStats returns velocity stats
func (e *InMemoryAntiFraudEngine) GetVelocityStats(ctx context.Context, entityType string, entityHash string) (*VelocityStats, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := entityType + ":" + entityHash
	entries := e.velocity[key]

	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	hourCount := 0
	dayCount := 0
	accounts := make(map[string]bool)
	var lastRequest *time.Time

	for _, entry := range entries {
		if entry.Timestamp.After(hourAgo) {
			hourCount++
		}
		if entry.Timestamp.After(dayAgo) {
			dayCount++
			accounts[entry.Account] = true
		}
		if lastRequest == nil || entry.Timestamp.After(*lastRequest) {
			t := entry.Timestamp
			lastRequest = &t
		}
	}

	isBlocked := false
	blockReason := ""
	if entityType == "phone" {
		isBlocked, blockReason, _ = e.IsPhoneBlocked(ctx, entityHash)
	} else if entityType == "ip" {
		isBlocked, blockReason, _ = e.IsIPBlocked(ctx, entityHash)
	}

	return &VelocityStats{
		EntityType:            entityType,
		EntityHash:            entityHash,
		RequestsLastHour:      hourCount,
		RequestsLastDay:       dayCount,
		UniqueAccountsLastDay: len(accounts),
		LastRequestAt:         lastRequest,
		IsBlocked:             isBlocked,
		BlockReason:           blockReason,
	}, nil
}

// Close closes the engine
func (e *InMemoryAntiFraudEngine) Close() error {
	return nil
}

// Ensure InMemoryAntiFraudEngine implements AntiFraudEngine
var _ AntiFraudEngine = (*InMemoryAntiFraudEngine)(nil)

