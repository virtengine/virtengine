// Package types provides types for the VEID module.
//
// VE-910: Rate limiting for SMS verification
// This file defines rate limiting types to prevent SMS abuse.
package types

import (
	"time"
)

// Rate limiting constants
const (
	// DefaultAccountRateLimitPerHour is the max SMS requests per account per hour
	DefaultAccountRateLimitPerHour uint32 = 5

	// DefaultAccountRateLimitPerDay is the max SMS requests per account per day
	DefaultAccountRateLimitPerDay uint32 = 10

	// DefaultPhoneRateLimitPerHour is the max SMS requests per phone per hour
	DefaultPhoneRateLimitPerHour uint32 = 3

	// DefaultPhoneRateLimitPerDay is the max SMS requests per phone per day
	DefaultPhoneRateLimitPerDay uint32 = 5

	// DefaultGlobalRateLimitPerMinute is the global rate limit per minute
	DefaultGlobalRateLimitPerMinute uint32 = 100

	// DefaultCooldownSeconds is the cooldown between OTP requests (same phone)
	DefaultCooldownSeconds int64 = 60

	// DefaultBlockDurationHours is how long a rate-limited entity is blocked
	DefaultBlockDurationHours int64 = 24
)

// RateLimitType represents the type of rate limit
type RateLimitType string

const (
	// RateLimitTypeAccount limits requests per account
	RateLimitTypeAccount RateLimitType = "account"

	// RateLimitTypePhone limits requests per phone hash
	RateLimitTypePhone RateLimitType = "phone"

	// RateLimitTypeIP limits requests per IP hash
	RateLimitTypeIP RateLimitType = "ip"

	// RateLimitTypeGlobal is the global rate limit
	RateLimitTypeGlobal RateLimitType = "global"
)

// SMSRateLimitConfig defines rate limiting configuration
type SMSRateLimitConfig struct {
	// AccountLimitPerHour is max requests per account per hour
	AccountLimitPerHour uint32 `json:"account_limit_per_hour"`

	// AccountLimitPerDay is max requests per account per day
	AccountLimitPerDay uint32 `json:"account_limit_per_day"`

	// PhoneLimitPerHour is max requests per phone hash per hour
	PhoneLimitPerHour uint32 `json:"phone_limit_per_hour"`

	// PhoneLimitPerDay is max requests per phone hash per day
	PhoneLimitPerDay uint32 `json:"phone_limit_per_day"`

	// GlobalLimitPerMinute is the global rate limit per minute
	GlobalLimitPerMinute uint32 `json:"global_limit_per_minute"`

	// CooldownSeconds is the cooldown between requests
	CooldownSeconds int64 `json:"cooldown_seconds"`

	// BlockDurationHours is how long to block rate-limited entities
	BlockDurationHours int64 `json:"block_duration_hours"`

	// EnableIPRateLimit enables IP-based rate limiting
	EnableIPRateLimit bool `json:"enable_ip_rate_limit"`

	// IPLimitPerHour is max requests per IP hash per hour (if enabled)
	IPLimitPerHour uint32 `json:"ip_limit_per_hour"`
}

// DefaultSMSRateLimitConfig returns the default rate limit configuration
func DefaultSMSRateLimitConfig() SMSRateLimitConfig {
	return SMSRateLimitConfig{
		AccountLimitPerHour:  DefaultAccountRateLimitPerHour,
		AccountLimitPerDay:   DefaultAccountRateLimitPerDay,
		PhoneLimitPerHour:    DefaultPhoneRateLimitPerHour,
		PhoneLimitPerDay:     DefaultPhoneRateLimitPerDay,
		GlobalLimitPerMinute: DefaultGlobalRateLimitPerMinute,
		CooldownSeconds:      DefaultCooldownSeconds,
		BlockDurationHours:   DefaultBlockDurationHours,
		EnableIPRateLimit:    true,
		IPLimitPerHour:       10,
	}
}

// Validate validates the rate limit configuration
func (c *SMSRateLimitConfig) Validate() error {
	if c.AccountLimitPerHour == 0 {
		return ErrInvalidRateLimit.Wrap("account_limit_per_hour must be positive")
	}
	if c.AccountLimitPerDay < c.AccountLimitPerHour {
		return ErrInvalidRateLimit.Wrap("account_limit_per_day must be >= account_limit_per_hour")
	}
	if c.PhoneLimitPerHour == 0 {
		return ErrInvalidRateLimit.Wrap("phone_limit_per_hour must be positive")
	}
	if c.PhoneLimitPerDay < c.PhoneLimitPerHour {
		return ErrInvalidRateLimit.Wrap("phone_limit_per_day must be >= phone_limit_per_hour")
	}
	if c.GlobalLimitPerMinute == 0 {
		return ErrInvalidRateLimit.Wrap("global_limit_per_minute must be positive")
	}
	if c.CooldownSeconds < 30 {
		return ErrInvalidRateLimit.Wrap("cooldown_seconds must be at least 30")
	}
	if c.BlockDurationHours < 1 {
		return ErrInvalidRateLimit.Wrap("block_duration_hours must be at least 1")
	}
	return nil
}

// SMSRateLimitState tracks rate limit state for an entity
type SMSRateLimitState struct {
	// EntityType is the type of entity (account, phone, ip)
	EntityType RateLimitType `json:"entity_type"`

	// EntityHash is the hash of the entity identifier
	EntityHash string `json:"entity_hash"`

	// HourlyCount is the request count in the current hour window
	HourlyCount uint32 `json:"hourly_count"`

	// HourlyWindowStart is when the current hour window started
	HourlyWindowStart time.Time `json:"hourly_window_start"`

	// DailyCount is the request count in the current day window
	DailyCount uint32 `json:"daily_count"`

	// DailyWindowStart is when the current day window started
	DailyWindowStart time.Time `json:"daily_window_start"`

	// LastRequestAt is when the last request was made
	LastRequestAt time.Time `json:"last_request_at"`

	// IsBlocked indicates if this entity is currently blocked
	IsBlocked bool `json:"is_blocked"`

	// BlockedUntil is when the block expires (if blocked)
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`

	// BlockReason is why this entity was blocked
	BlockReason string `json:"block_reason,omitempty"`

	// TotalRequests is the total lifetime request count
	TotalRequests uint64 `json:"total_requests"`

	// TotalBlocks is the total number of times this entity was blocked
	TotalBlocks uint32 `json:"total_blocks"`
}

// NewSMSRateLimitState creates a new rate limit state
func NewSMSRateLimitState(entityType RateLimitType, entityHash string, now time.Time) *SMSRateLimitState {
	return &SMSRateLimitState{
		EntityType:        entityType,
		EntityHash:        entityHash,
		HourlyCount:       0,
		HourlyWindowStart: now,
		DailyCount:        0,
		DailyWindowStart:  now,
		LastRequestAt:     time.Time{},
		IsBlocked:         false,
		TotalRequests:     0,
		TotalBlocks:       0,
	}
}

// ResetHourlyWindow resets the hourly counter if the window has expired
func (s *SMSRateLimitState) ResetHourlyWindow(now time.Time) {
	if now.Sub(s.HourlyWindowStart) >= time.Hour {
		s.HourlyCount = 0
		s.HourlyWindowStart = now
	}
}

// ResetDailyWindow resets the daily counter if the window has expired
func (s *SMSRateLimitState) ResetDailyWindow(now time.Time) {
	if now.Sub(s.DailyWindowStart) >= 24*time.Hour {
		s.DailyCount = 0
		s.DailyWindowStart = now
	}
}

// RecordRequest records a new request
func (s *SMSRateLimitState) RecordRequest(now time.Time) {
	s.ResetHourlyWindow(now)
	s.ResetDailyWindow(now)

	s.HourlyCount++
	s.DailyCount++
	s.LastRequestAt = now
	s.TotalRequests++
}

// CheckCooldown checks if cooldown period has passed
func (s *SMSRateLimitState) CheckCooldown(now time.Time, cooldownSeconds int64) bool {
	if s.LastRequestAt.IsZero() {
		return true
	}
	return now.Sub(s.LastRequestAt) >= time.Duration(cooldownSeconds)*time.Second
}

// IsBlockExpired checks if the block has expired
func (s *SMSRateLimitState) IsBlockExpired(now time.Time) bool {
	if !s.IsBlocked || s.BlockedUntil == nil {
		return true
	}
	return now.After(*s.BlockedUntil)
}

// Block blocks this entity
func (s *SMSRateLimitState) Block(now time.Time, durationHours int64, reason string) {
	blockedUntil := now.Add(time.Duration(durationHours) * time.Hour)
	s.IsBlocked = true
	s.BlockedUntil = &blockedUntil
	s.BlockReason = reason
	s.TotalBlocks++
}

// Unblock removes the block
func (s *SMSRateLimitState) Unblock() {
	s.IsBlocked = false
	s.BlockedUntil = nil
	s.BlockReason = ""
}

// SMSRateLimiter provides rate limiting functionality
type SMSRateLimiter struct {
	config SMSRateLimitConfig
}

// NewSMSRateLimiter creates a new rate limiter
func NewSMSRateLimiter(config SMSRateLimitConfig) *SMSRateLimiter {
	return &SMSRateLimiter{
		config: config,
	}
}

// RateLimitCheckResult represents the result of a rate limit check
type RateLimitCheckResult struct {
	// Allowed indicates if the request is allowed
	Allowed bool `json:"allowed"`

	// LimitType is the type of limit that was hit (if not allowed)
	LimitType RateLimitType `json:"limit_type,omitempty"`

	// LimitReason explains why the request was blocked
	LimitReason string `json:"limit_reason,omitempty"`

	// RetryAfter is when the request can be retried (seconds)
	RetryAfter int64 `json:"retry_after,omitempty"`

	// RemainingHourly is the remaining hourly quota
	RemainingHourly uint32 `json:"remaining_hourly"`

	// RemainingDaily is the remaining daily quota
	RemainingDaily uint32 `json:"remaining_daily"`
}

// CheckAccountLimit checks if an account is within rate limits
func (r *SMSRateLimiter) CheckAccountLimit(state *SMSRateLimitState, now time.Time) RateLimitCheckResult {
	// Check if blocked
	if state.IsBlocked && !state.IsBlockExpired(now) {
		return RateLimitCheckResult{
			Allowed:     false,
			LimitType:   RateLimitTypeAccount,
			LimitReason: "account is temporarily blocked: " + state.BlockReason,
			RetryAfter:  int64(state.BlockedUntil.Sub(now).Seconds()),
		}
	}

	// Unblock if expired
	if state.IsBlocked && state.IsBlockExpired(now) {
		state.Unblock()
	}

	// Reset windows
	state.ResetHourlyWindow(now)
	state.ResetDailyWindow(now)

	// Check cooldown
	if !state.CheckCooldown(now, r.config.CooldownSeconds) {
		waitTime := r.config.CooldownSeconds - int64(now.Sub(state.LastRequestAt).Seconds())
		return RateLimitCheckResult{
			Allowed:         false,
			LimitType:       RateLimitTypeAccount,
			LimitReason:     "cooldown period not elapsed",
			RetryAfter:      waitTime,
			RemainingHourly: r.config.AccountLimitPerHour - state.HourlyCount,
			RemainingDaily:  r.config.AccountLimitPerDay - state.DailyCount,
		}
	}

	// Check hourly limit
	if state.HourlyCount >= r.config.AccountLimitPerHour {
		waitTime := int64(time.Hour.Seconds()) - int64(now.Sub(state.HourlyWindowStart).Seconds())
		return RateLimitCheckResult{
			Allowed:         false,
			LimitType:       RateLimitTypeAccount,
			LimitReason:     "hourly rate limit exceeded",
			RetryAfter:      waitTime,
			RemainingHourly: 0,
			RemainingDaily:  r.config.AccountLimitPerDay - state.DailyCount,
		}
	}

	// Check daily limit
	if state.DailyCount >= r.config.AccountLimitPerDay {
		waitTime := int64(24*time.Hour.Seconds()) - int64(now.Sub(state.DailyWindowStart).Seconds())
		return RateLimitCheckResult{
			Allowed:         false,
			LimitType:       RateLimitTypeAccount,
			LimitReason:     "daily rate limit exceeded",
			RetryAfter:      waitTime,
			RemainingHourly: r.config.AccountLimitPerHour - state.HourlyCount,
			RemainingDaily:  0,
		}
	}

	return RateLimitCheckResult{
		Allowed:         true,
		RemainingHourly: r.config.AccountLimitPerHour - state.HourlyCount - 1,
		RemainingDaily:  r.config.AccountLimitPerDay - state.DailyCount - 1,
	}
}

// CheckPhoneLimit checks if a phone hash is within rate limits
func (r *SMSRateLimiter) CheckPhoneLimit(state *SMSRateLimitState, now time.Time) RateLimitCheckResult {
	// Check if blocked
	if state.IsBlocked && !state.IsBlockExpired(now) {
		return RateLimitCheckResult{
			Allowed:     false,
			LimitType:   RateLimitTypePhone,
			LimitReason: "phone number is temporarily blocked: " + state.BlockReason,
			RetryAfter:  int64(state.BlockedUntil.Sub(now).Seconds()),
		}
	}

	// Unblock if expired
	if state.IsBlocked && state.IsBlockExpired(now) {
		state.Unblock()
	}

	// Reset windows
	state.ResetHourlyWindow(now)
	state.ResetDailyWindow(now)

	// Check hourly limit
	if state.HourlyCount >= r.config.PhoneLimitPerHour {
		waitTime := int64(time.Hour.Seconds()) - int64(now.Sub(state.HourlyWindowStart).Seconds())
		return RateLimitCheckResult{
			Allowed:         false,
			LimitType:       RateLimitTypePhone,
			LimitReason:     "hourly rate limit exceeded for this phone",
			RetryAfter:      waitTime,
			RemainingHourly: 0,
			RemainingDaily:  r.config.PhoneLimitPerDay - state.DailyCount,
		}
	}

	// Check daily limit
	if state.DailyCount >= r.config.PhoneLimitPerDay {
		waitTime := int64(24*time.Hour.Seconds()) - int64(now.Sub(state.DailyWindowStart).Seconds())
		return RateLimitCheckResult{
			Allowed:         false,
			LimitType:       RateLimitTypePhone,
			LimitReason:     "daily rate limit exceeded for this phone",
			RetryAfter:      waitTime,
			RemainingHourly: r.config.PhoneLimitPerHour - state.HourlyCount,
			RemainingDaily:  0,
		}
	}

	return RateLimitCheckResult{
		Allowed:         true,
		RemainingHourly: r.config.PhoneLimitPerHour - state.HourlyCount - 1,
		RemainingDaily:  r.config.PhoneLimitPerDay - state.DailyCount - 1,
	}
}

// GlobalRateLimitState tracks global rate limit state
type GlobalRateLimitState struct {
	// MinuteCount is requests in the current minute window
	MinuteCount uint32 `json:"minute_count"`

	// MinuteWindowStart is when the current minute window started
	MinuteWindowStart time.Time `json:"minute_window_start"`

	// TotalRequests is lifetime request count
	TotalRequests uint64 `json:"total_requests"`
}

// NewGlobalRateLimitState creates a new global rate limit state
func NewGlobalRateLimitState(now time.Time) *GlobalRateLimitState {
	return &GlobalRateLimitState{
		MinuteCount:       0,
		MinuteWindowStart: now,
		TotalRequests:     0,
	}
}

// ResetMinuteWindow resets if the window has expired
func (s *GlobalRateLimitState) ResetMinuteWindow(now time.Time) {
	if now.Sub(s.MinuteWindowStart) >= time.Minute {
		s.MinuteCount = 0
		s.MinuteWindowStart = now
	}
}

// RecordRequest records a global request
func (s *GlobalRateLimitState) RecordRequest(now time.Time) {
	s.ResetMinuteWindow(now)
	s.MinuteCount++
	s.TotalRequests++
}

// CheckGlobalLimit checks the global rate limit
func (r *SMSRateLimiter) CheckGlobalLimit(state *GlobalRateLimitState, now time.Time) RateLimitCheckResult {
	state.ResetMinuteWindow(now)

	if state.MinuteCount >= r.config.GlobalLimitPerMinute {
		waitTime := int64(time.Minute.Seconds()) - int64(now.Sub(state.MinuteWindowStart).Seconds())
		return RateLimitCheckResult{
			Allowed:     false,
			LimitType:   RateLimitTypeGlobal,
			LimitReason: "global rate limit exceeded, try again shortly",
			RetryAfter:  waitTime,
		}
	}

	return RateLimitCheckResult{
		Allowed: true,
	}
}
