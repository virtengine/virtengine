// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines anti-manipulation safeguards for the marketplace.
package marketplace

import (
	"fmt"
	"time"
)

// ManipulationType represents the type of market manipulation
type ManipulationType uint8

const (
	// ManipulationTypeNone indicates no manipulation detected
	ManipulationTypeNone ManipulationType = 0

	// ManipulationTypeWashTrading is self-dealing to inflate volume
	ManipulationTypeWashTrading ManipulationType = 1

	// ManipulationTypeSpoofing is placing orders with intent to cancel
	ManipulationTypeSpoofing ManipulationType = 2

	// ManipulationTypeLayering is placing multiple orders to manipulate price
	ManipulationTypeLayering ManipulationType = 3

	// ManipulationTypeFrontRunning is trading ahead of known orders
	ManipulationTypeFrontRunning ManipulationType = 4

	// ManipulationTypePriceManipulation is artificial price inflation/deflation
	ManipulationTypePriceManipulation ManipulationType = 5

	// ManipulationTypeOrderSpamming is excessive order placement
	ManipulationTypeOrderSpamming ManipulationType = 6

	// ManipulationTypeSybilAttack is creating multiple identities
	ManipulationTypeSybilAttack ManipulationType = 7
)

// ManipulationTypeNames maps types to human-readable names
var ManipulationTypeNames = map[ManipulationType]string{
	ManipulationTypeNone:              "none",
	ManipulationTypeWashTrading:       "wash_trading",
	ManipulationTypeSpoofing:          "spoofing",
	ManipulationTypeLayering:          "layering",
	ManipulationTypeFrontRunning:      "front_running",
	ManipulationTypePriceManipulation: "price_manipulation",
	ManipulationTypeOrderSpamming:     "order_spamming",
	ManipulationTypeSybilAttack:       "sybil_attack",
}

// ManipulationTypeSeverity maps types to severity levels (1-10)
var ManipulationTypeSeverity = map[ManipulationType]uint8{
	ManipulationTypeNone:              0,
	ManipulationTypeWashTrading:       7,
	ManipulationTypeSpoofing:          6,
	ManipulationTypeLayering:          6,
	ManipulationTypeFrontRunning:      9,
	ManipulationTypePriceManipulation: 8,
	ManipulationTypeOrderSpamming:     4,
	ManipulationTypeSybilAttack:       10,
}

// String returns the string representation of a ManipulationType
func (t ManipulationType) String() string {
	if name, ok := ManipulationTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// Severity returns the severity level of this manipulation type
func (t ManipulationType) Severity() uint8 {
	if sev, ok := ManipulationTypeSeverity[t]; ok {
		return sev
	}
	return 0
}

// PenaltyAction represents the action to take for violations
type PenaltyAction uint8

const (
	// PenaltyActionNone indicates no penalty
	PenaltyActionNone PenaltyAction = 0

	// PenaltyActionWarning is a warning without penalty
	PenaltyActionWarning PenaltyAction = 1

	// PenaltyActionFee is a fee penalty
	PenaltyActionFee PenaltyAction = 2

	// PenaltyActionCooldown is a trading cooldown
	PenaltyActionCooldown PenaltyAction = 3

	// PenaltyActionOrderCancel is forced order cancellation
	PenaltyActionOrderCancel PenaltyAction = 4

	// PenaltyActionSuspension is temporary suspension
	PenaltyActionSuspension PenaltyAction = 5

	// PenaltyActionBan is permanent ban
	PenaltyActionBan PenaltyAction = 6
)

// PenaltyActionNames maps actions to human-readable names
var PenaltyActionNames = map[PenaltyAction]string{
	PenaltyActionNone:        "none",
	PenaltyActionWarning:     "warning",
	PenaltyActionFee:         "fee",
	PenaltyActionCooldown:    "cooldown",
	PenaltyActionOrderCancel: "order_cancel",
	PenaltyActionSuspension:  "suspension",
	PenaltyActionBan:         "ban",
}

// String returns the string representation of a PenaltyAction
func (a PenaltyAction) String() string {
	if name, ok := PenaltyActionNames[a]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", a)
}

// CircuitBreakerStatus represents the circuit breaker status
type CircuitBreakerStatus uint8

const (
	// CircuitBreakerStatusClosed indicates normal operation
	CircuitBreakerStatusClosed CircuitBreakerStatus = 0

	// CircuitBreakerStatusOpen indicates trading is halted
	CircuitBreakerStatusOpen CircuitBreakerStatus = 1

	// CircuitBreakerStatusHalfOpen indicates limited trading during recovery
	CircuitBreakerStatusHalfOpen CircuitBreakerStatus = 2
)

// CircuitBreakerStatusNames maps statuses to human-readable names
var CircuitBreakerStatusNames = map[CircuitBreakerStatus]string{
	CircuitBreakerStatusClosed:   "closed",
	CircuitBreakerStatusOpen:     "open",
	CircuitBreakerStatusHalfOpen: "half_open",
}

// String returns the string representation of a CircuitBreakerStatus
func (s CircuitBreakerStatus) String() string {
	if name, ok := CircuitBreakerStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// WashTradingConfig defines wash trading detection configuration
type WashTradingConfig struct {
	// Enabled indicates if wash trading detection is enabled
	Enabled bool `json:"enabled"`

	// SelfDealingWindowBlocks is the window for self-dealing detection
	SelfDealingWindowBlocks int64 `json:"self_dealing_window_blocks"`

	// MinSuspiciousVolumeRatioBps is min volume ratio to flag (basis points)
	MinSuspiciousVolumeRatioBps uint32 `json:"min_suspicious_volume_ratio_bps"`

	// MaxAllowedSelfDealingPct is max allowed self-dealing percentage
	MaxAllowedSelfDealingPct uint32 `json:"max_allowed_self_dealing_pct"`

	// RelatedAddressDepthCheck is how many hops to check for related addresses
	RelatedAddressDepthCheck uint32 `json:"related_address_depth_check"`

	// PenaltyBps is the penalty in basis points of suspicious volume
	PenaltyBps uint32 `json:"penalty_bps"`
}

// DefaultWashTradingConfig returns default wash trading configuration
func DefaultWashTradingConfig() WashTradingConfig {
	return WashTradingConfig{
		Enabled:                     true,
		SelfDealingWindowBlocks:     1000, // ~1.6 hours
		MinSuspiciousVolumeRatioBps: 5000, // 50% of volume
		MaxAllowedSelfDealingPct:    5,    // 5% allowed
		RelatedAddressDepthCheck:    2,
		PenaltyBps:                  1000, // 10% penalty
	}
}

// SpoofingConfig defines spoofing detection configuration
type SpoofingConfig struct {
	// Enabled indicates if spoofing detection is enabled
	Enabled bool `json:"enabled"`

	// MinOrderLifetimeBlocks is minimum expected order lifetime
	MinOrderLifetimeBlocks int64 `json:"min_order_lifetime_blocks"`

	// MaxCancellationRatePct is max allowed cancellation rate
	MaxCancellationRatePct uint32 `json:"max_cancellation_rate_pct"`

	// CancellationWindowBlocks is the window for measuring cancellations
	CancellationWindowBlocks int64 `json:"cancellation_window_blocks"`

	// MinOrdersForEvaluation is minimum orders to evaluate
	MinOrdersForEvaluation uint32 `json:"min_orders_for_evaluation"`

	// PenaltyPerCancelledOrderBps is penalty per cancelled order
	PenaltyPerCancelledOrderBps uint32 `json:"penalty_per_cancelled_order_bps"`
}

// DefaultSpoofingConfig returns default spoofing configuration
func DefaultSpoofingConfig() SpoofingConfig {
	return SpoofingConfig{
		Enabled:                     true,
		MinOrderLifetimeBlocks:      5,   // ~30 seconds
		MaxCancellationRatePct:      80,  // 80% max cancellation
		CancellationWindowBlocks:    100, // ~10 minutes
		MinOrdersForEvaluation:      5,
		PenaltyPerCancelledOrderBps: 100, // 1% per cancelled order
	}
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	// Enabled indicates if rate limiting is enabled
	Enabled bool `json:"enabled"`

	// MaxOrdersPerBlock is max orders per block per address
	MaxOrdersPerBlock uint32 `json:"max_orders_per_block"`

	// MaxOrdersPerMinute is max orders per minute per address
	MaxOrdersPerMinute uint32 `json:"max_orders_per_minute"`

	// MaxBidsPerBlock is max bids per block per address
	MaxBidsPerBlock uint32 `json:"max_bids_per_block"`

	// MaxBidsPerMinute is max bids per minute per address
	MaxBidsPerMinute uint32 `json:"max_bids_per_minute"`

	// MaxCancelsPerBlock is max cancellations per block per address
	MaxCancelsPerBlock uint32 `json:"max_cancels_per_block"`

	// BurstAllowance is the burst allowance multiplier (100 = 1x)
	BurstAllowance uint32 `json:"burst_allowance"`

	// CooldownBlocks is cooldown after rate limit hit
	CooldownBlocks int64 `json:"cooldown_blocks"`
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:            true,
		MaxOrdersPerBlock:  5,
		MaxOrdersPerMinute: 20,
		MaxBidsPerBlock:    10,
		MaxBidsPerMinute:   50,
		MaxCancelsPerBlock: 10,
		BurstAllowance:     200, // 2x burst allowed
		CooldownBlocks:     10,
	}
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	// Enabled indicates if circuit breaker is enabled
	Enabled bool `json:"enabled"`

	// PriceMovementThresholdBps is the threshold for triggering
	PriceMovementThresholdBps uint32 `json:"price_movement_threshold_bps"`

	// VolumeSpikethresholdPct is threshold for volume spike
	VolumeSpikeThresholdPct uint32 `json:"volume_spike_threshold_pct"`

	// TripDurationBlocks is how long circuit stays open
	TripDurationBlocks int64 `json:"trip_duration_blocks"`

	// HalfOpenDurationBlocks is half-open period duration
	HalfOpenDurationBlocks int64 `json:"half_open_duration_blocks"`

	// ConsecutiveTripsForExtension is trips needed for extended halt
	ConsecutiveTripsForExtension uint32 `json:"consecutive_trips_for_extension"`

	// ExtensionMultiplier multiplies duration on consecutive trips
	ExtensionMultiplier uint32 `json:"extension_multiplier"`

	// MaxOrdersDuringHalfOpen limits orders during recovery
	MaxOrdersDuringHalfOpen uint32 `json:"max_orders_during_half_open"`
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:                      true,
		PriceMovementThresholdBps:    2000, // 20% move
		VolumeSpikeThresholdPct:      500,  // 5x normal volume
		TripDurationBlocks:           100,  // ~10 minutes
		HalfOpenDurationBlocks:       50,   // ~5 minutes
		ConsecutiveTripsForExtension: 3,
		ExtensionMultiplier:          2,
		MaxOrdersDuringHalfOpen:      10,
	}
}

// PenaltyConfig defines penalty configuration for violations
type PenaltyConfig struct {
	// WarningThreshold is violations before warning
	WarningThreshold uint32 `json:"warning_threshold"`

	// FeeThreshold is violations before fee penalty
	FeeThreshold uint32 `json:"fee_threshold"`

	// CooldownThreshold is violations before cooldown
	CooldownThreshold uint32 `json:"cooldown_threshold"`

	// SuspensionThreshold is violations before suspension
	SuspensionThreshold uint32 `json:"suspension_threshold"`

	// BanThreshold is violations before ban
	BanThreshold uint32 `json:"ban_threshold"`

	// BasePenaltyBps is base penalty in basis points
	BasePenaltyBps uint32 `json:"base_penalty_bps"`

	// EscalationMultiplier increases penalty per violation
	EscalationMultiplier uint32 `json:"escalation_multiplier"`

	// CooldownDurationBlocks is cooldown duration
	CooldownDurationBlocks int64 `json:"cooldown_duration_blocks"`

	// SuspensionDurationBlocks is suspension duration
	SuspensionDurationBlocks int64 `json:"suspension_duration_blocks"`

	// ViolationDecayBlocks is blocks for violation count to decay
	ViolationDecayBlocks int64 `json:"violation_decay_blocks"`
}

// DefaultPenaltyConfig returns default penalty configuration
func DefaultPenaltyConfig() PenaltyConfig {
	return PenaltyConfig{
		WarningThreshold:         1,
		FeeThreshold:             3,
		CooldownThreshold:        5,
		SuspensionThreshold:      10,
		BanThreshold:             20,
		BasePenaltyBps:           100,    // 1%
		EscalationMultiplier:     150,    // 1.5x per violation
		CooldownDurationBlocks:   100,    // ~10 minutes
		SuspensionDurationBlocks: 14400,  // ~1 day
		ViolationDecayBlocks:     100800, // ~1 week
	}
}

// SafeguardConfig holds all safeguard configurations
type SafeguardConfig struct {
	// Enabled indicates if safeguards are enabled
	Enabled bool `json:"enabled"`

	// WashTrading is wash trading detection config
	WashTrading WashTradingConfig `json:"wash_trading"`

	// Spoofing is spoofing detection config
	Spoofing SpoofingConfig `json:"spoofing"`

	// RateLimit is rate limiting config
	RateLimit RateLimitConfig `json:"rate_limit"`

	// CircuitBreaker is circuit breaker config
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`

	// Penalty is penalty config
	Penalty PenaltyConfig `json:"penalty"`

	// LogSuspiciousActivity enables logging of suspicious activity
	LogSuspiciousActivity bool `json:"log_suspicious_activity"`

	// AutomaticEnforcement enables automatic penalty enforcement
	AutomaticEnforcement bool `json:"automatic_enforcement"`
}

// DefaultSafeguardConfig returns the default safeguard configuration
func DefaultSafeguardConfig() SafeguardConfig {
	return SafeguardConfig{
		Enabled:               true,
		WashTrading:           DefaultWashTradingConfig(),
		Spoofing:              DefaultSpoofingConfig(),
		RateLimit:             DefaultRateLimitConfig(),
		CircuitBreaker:        DefaultCircuitBreakerConfig(),
		Penalty:               DefaultPenaltyConfig(),
		LogSuspiciousActivity: true,
		AutomaticEnforcement:  true,
	}
}

// Validate validates the safeguard configuration
func (c *SafeguardConfig) Validate() error {
	if c.WashTrading.MaxAllowedSelfDealingPct > 100 {
		return fmt.Errorf("max_allowed_self_dealing_pct cannot exceed 100")
	}
	if c.Spoofing.MaxCancellationRatePct > 100 {
		return fmt.Errorf("max_cancellation_rate_pct cannot exceed 100")
	}
	if c.Penalty.WarningThreshold == 0 {
		return fmt.Errorf("warning_threshold must be positive")
	}
	if c.Penalty.FeeThreshold <= c.Penalty.WarningThreshold {
		return fmt.Errorf("fee_threshold must be greater than warning_threshold")
	}
	return nil
}

// ViolationRecord records a detected violation
type ViolationRecord struct {
	// Address is the violator's address
	Address string `json:"address"`

	// Type is the manipulation type
	Type ManipulationType `json:"type"`

	// Severity is the violation severity (1-10)
	Severity uint8 `json:"severity"`

	// Description describes the violation
	Description string `json:"description"`

	// Evidence contains evidence of the violation
	Evidence map[string]string `json:"evidence,omitempty"`

	// DetectedAt is when the violation was detected
	DetectedAt time.Time `json:"detected_at"`

	// BlockHeight is the block height when detected
	BlockHeight int64 `json:"block_height"`

	// Action is the penalty action taken
	Action PenaltyAction `json:"action"`

	// PenaltyAmount is the penalty amount if applicable
	PenaltyAmount uint64 `json:"penalty_amount,omitempty"`

	// ExpiresAt is when the record expires for counting
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// AccountSafeguardState tracks safeguard state for an account
type AccountSafeguardState struct {
	// Address is the account address
	Address string `json:"address"`

	// ViolationCount is the current violation count
	ViolationCount uint32 `json:"violation_count"`

	// Violations are the recorded violations
	Violations []ViolationRecord `json:"violations"`

	// IsSuspended indicates if the account is suspended
	IsSuspended bool `json:"is_suspended"`

	// SuspendedUntil is when suspension ends
	SuspendedUntil *time.Time `json:"suspended_until,omitempty"`

	// IsBanned indicates if the account is permanently banned
	IsBanned bool `json:"is_banned"`

	// BannedAt is when the account was banned
	BannedAt *time.Time `json:"banned_at,omitempty"`

	// CooldownUntil is when cooldown ends
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`

	// TotalPenaltiesPaid is total penalties paid
	TotalPenaltiesPaid uint64 `json:"total_penalties_paid"`

	// OrdersThisBlock is orders placed this block
	OrdersThisBlock uint32 `json:"orders_this_block"`

	// OrdersThisMinute is orders placed this minute
	OrdersThisMinute uint32 `json:"orders_this_minute"`

	// BidsThisBlock is bids placed this block
	BidsThisBlock uint32 `json:"bids_this_block"`

	// BidsThisMinute is bids placed this minute
	BidsThisMinute uint32 `json:"bids_this_minute"`

	// CancelsThisBlock is cancellations this block
	CancelsThisBlock uint32 `json:"cancels_this_block"`

	// LastActivityBlock is the last activity block
	LastActivityBlock int64 `json:"last_activity_block"`

	// LastMinuteReset is when minute counters were reset
	LastMinuteReset time.Time `json:"last_minute_reset"`

	// RelatedAddresses are addresses linked to this account
	RelatedAddresses []string `json:"related_addresses,omitempty"`

	// TrustScore is the account's trust score (0-100)
	TrustScore uint32 `json:"trust_score"`

	// LastUpdated is when the state was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewAccountSafeguardState creates a new account safeguard state
func NewAccountSafeguardState(address string) *AccountSafeguardState {
	return &AccountSafeguardState{
		Address:    address,
		Violations: make([]ViolationRecord, 0),
		TrustScore: 100, // Start with full trust
	}
}

// CanTrade checks if the account can trade
func (s *AccountSafeguardState) CanTrade(now time.Time) error {
	if s.IsBanned {
		return fmt.Errorf("account is permanently banned")
	}
	if s.IsSuspended && s.SuspendedUntil != nil && now.Before(*s.SuspendedUntil) {
		return fmt.Errorf("account is suspended until %s", s.SuspendedUntil.Format(time.RFC3339))
	}
	if s.CooldownUntil != nil && now.Before(*s.CooldownUntil) {
		return fmt.Errorf("account is in cooldown until %s", s.CooldownUntil.Format(time.RFC3339))
	}
	return nil
}

// RecordViolation records a violation and returns the penalty action
func (s *AccountSafeguardState) RecordViolation(violation ViolationRecord, config PenaltyConfig) PenaltyAction {
	s.Violations = append(s.Violations, violation)
	s.ViolationCount++
	s.TrustScore = s.calculateTrustScore()
	s.LastUpdated = violation.DetectedAt

	// Determine penalty action
	switch {
	case s.ViolationCount >= config.BanThreshold:
		s.IsBanned = true
		s.BannedAt = &violation.DetectedAt
		return PenaltyActionBan
	case s.ViolationCount >= config.SuspensionThreshold:
		s.IsSuspended = true
		endTime := violation.DetectedAt.Add(time.Duration(config.SuspensionDurationBlocks*6) * time.Second)
		s.SuspendedUntil = &endTime
		return PenaltyActionSuspension
	case s.ViolationCount >= config.CooldownThreshold:
		endTime := violation.DetectedAt.Add(time.Duration(config.CooldownDurationBlocks*6) * time.Second)
		s.CooldownUntil = &endTime
		return PenaltyActionCooldown
	case s.ViolationCount >= config.FeeThreshold:
		return PenaltyActionFee
	case s.ViolationCount >= config.WarningThreshold:
		return PenaltyActionWarning
	default:
		return PenaltyActionNone
	}
}

// calculateTrustScore calculates trust score based on violations
func (s *AccountSafeguardState) calculateTrustScore() uint32 {
	if s.IsBanned {
		return 0
	}

	// Start at 100, deduct for violations
	score := int32(100)
	for _, v := range s.Violations {
		score -= int32(v.Severity) * 2
	}

	if score < 0 {
		return 0
	}
	return uint32(score)
}

// ResetBlockCounters resets per-block counters
func (s *AccountSafeguardState) ResetBlockCounters() {
	s.OrdersThisBlock = 0
	s.BidsThisBlock = 0
	s.CancelsThisBlock = 0
}

// ResetMinuteCounters resets per-minute counters
func (s *AccountSafeguardState) ResetMinuteCounters(now time.Time) {
	s.OrdersThisMinute = 0
	s.BidsThisMinute = 0
	s.LastMinuteReset = now
}

// CheckRateLimit checks if an action would exceed rate limits
func (s *AccountSafeguardState) CheckRateLimit(actionType string, config RateLimitConfig) error {
	if !config.Enabled {
		return nil
	}

	switch actionType {
	case "order":
		if s.OrdersThisBlock >= config.MaxOrdersPerBlock {
			return fmt.Errorf("order rate limit exceeded: %d per block", config.MaxOrdersPerBlock)
		}
		if s.OrdersThisMinute >= config.MaxOrdersPerMinute {
			return fmt.Errorf("order rate limit exceeded: %d per minute", config.MaxOrdersPerMinute)
		}
	case "bid":
		if s.BidsThisBlock >= config.MaxBidsPerBlock {
			return fmt.Errorf("bid rate limit exceeded: %d per block", config.MaxBidsPerBlock)
		}
		if s.BidsThisMinute >= config.MaxBidsPerMinute {
			return fmt.Errorf("bid rate limit exceeded: %d per minute", config.MaxBidsPerMinute)
		}
	case "cancel":
		if s.CancelsThisBlock >= config.MaxCancelsPerBlock {
			return fmt.Errorf("cancel rate limit exceeded: %d per block", config.MaxCancelsPerBlock)
		}
	}

	return nil
}

// CircuitBreakerState tracks circuit breaker state for an offering
type CircuitBreakerState struct {
	// OfferingID is the offering ID
	OfferingID OfferingID `json:"offering_id"`

	// Status is the current circuit breaker status
	Status CircuitBreakerStatus `json:"status"`

	// TrippedAt is when the circuit was tripped
	TrippedAt *time.Time `json:"tripped_at,omitempty"`

	// TrippedAtBlock is the block when tripped
	TrippedAtBlock int64 `json:"tripped_at_block,omitempty"`

	// ResetAt is when the circuit will reset
	ResetAt *time.Time `json:"reset_at,omitempty"`

	// ResetAtBlock is the block when circuit resets
	ResetAtBlock int64 `json:"reset_at_block,omitempty"`

	// ConsecutiveTrips is the number of consecutive trips
	ConsecutiveTrips uint32 `json:"consecutive_trips"`

	// TripReason describes why the circuit tripped
	TripReason string `json:"trip_reason,omitempty"`

	// OrdersDuringHalfOpen is orders during half-open state
	OrdersDuringHalfOpen uint32 `json:"orders_during_half_open"`

	// LastUpdated is when state was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewCircuitBreakerState creates a new circuit breaker state
func NewCircuitBreakerState(offeringID OfferingID) *CircuitBreakerState {
	return &CircuitBreakerState{
		OfferingID: offeringID,
		Status:     CircuitBreakerStatusClosed,
	}
}

// Trip trips the circuit breaker
func (s *CircuitBreakerState) Trip(reason string, config CircuitBreakerConfig, currentBlock int64, now time.Time) {
	s.Status = CircuitBreakerStatusOpen
	s.TrippedAt = &now
	s.TrippedAtBlock = currentBlock
	s.TripReason = reason
	s.ConsecutiveTrips++

	// Calculate reset time with escalation
	duration := config.TripDurationBlocks
	if s.ConsecutiveTrips >= config.ConsecutiveTripsForExtension {
		duration *= int64(config.ExtensionMultiplier)
	}

	resetBlock := currentBlock + duration
	s.ResetAtBlock = resetBlock
	resetTime := now.Add(time.Duration(duration*6) * time.Second)
	s.ResetAt = &resetTime
	s.LastUpdated = now
}

// Update updates the circuit breaker state
func (s *CircuitBreakerState) Update(currentBlock int64, config CircuitBreakerConfig, now time.Time) {
	switch s.Status {
	case CircuitBreakerStatusOpen:
		if currentBlock >= s.ResetAtBlock {
			// Transition to half-open
			s.Status = CircuitBreakerStatusHalfOpen
			s.OrdersDuringHalfOpen = 0
			s.ResetAtBlock = currentBlock + config.HalfOpenDurationBlocks
			resetTime := now.Add(time.Duration(config.HalfOpenDurationBlocks*6) * time.Second)
			s.ResetAt = &resetTime
		}
	case CircuitBreakerStatusHalfOpen:
		if currentBlock >= s.ResetAtBlock {
			// Transition to closed
			s.Status = CircuitBreakerStatusClosed
			s.ConsecutiveTrips = 0
			s.TrippedAt = nil
			s.ResetAt = nil
		}
	}
	s.LastUpdated = now
}

// AllowOrder checks if an order is allowed under current circuit breaker state
func (s *CircuitBreakerState) AllowOrder(config CircuitBreakerConfig) error {
	switch s.Status {
	case CircuitBreakerStatusOpen:
		return fmt.Errorf("circuit breaker is open: trading halted")
	case CircuitBreakerStatusHalfOpen:
		if s.OrdersDuringHalfOpen >= config.MaxOrdersDuringHalfOpen {
			return fmt.Errorf("max orders during half-open reached: %d", config.MaxOrdersDuringHalfOpen)
		}
		s.OrdersDuringHalfOpen++
	}
	return nil
}

// ManipulationDetector detects various forms of market manipulation
type ManipulationDetector struct {
	Config SafeguardConfig `json:"config"`
}

// NewManipulationDetector creates a new manipulation detector
func NewManipulationDetector(config SafeguardConfig) *ManipulationDetector {
	return &ManipulationDetector{Config: config}
}

// WashTradingCheck checks for potential wash trading
type WashTradingCheck struct {
	// BuyerAddress is the buyer's address
	BuyerAddress string `json:"buyer_address"`

	// SellerAddress is the seller's address
	SellerAddress string `json:"seller_address"`

	// Amount is the trade amount
	Amount uint64 `json:"amount"`

	// Timestamp is when the trade occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`
}

// DetectWashTrading detects potential wash trading
func (d *ManipulationDetector) DetectWashTrading(check WashTradingCheck, accountState *AccountSafeguardState) *ViolationRecord {
	if !d.Config.Enabled || !d.Config.WashTrading.Enabled {
		return nil
	}

	// Check for self-dealing
	if check.BuyerAddress == check.SellerAddress {
		return &ViolationRecord{
			Address:     check.BuyerAddress,
			Type:        ManipulationTypeWashTrading,
			Severity:    ManipulationTypeWashTrading.Severity(),
			Description: "self-dealing detected",
			Evidence: map[string]string{
				"buyer":  check.BuyerAddress,
				"seller": check.SellerAddress,
				"amount": fmt.Sprintf("%d", check.Amount),
			},
			DetectedAt:  check.Timestamp,
			BlockHeight: check.BlockHeight,
		}
	}

	// Check for related addresses
	for _, related := range accountState.RelatedAddresses {
		if check.SellerAddress == related {
			return &ViolationRecord{
				Address:     check.BuyerAddress,
				Type:        ManipulationTypeWashTrading,
				Severity:    ManipulationTypeWashTrading.Severity(),
				Description: "trading with related address",
				Evidence: map[string]string{
					"buyer":           check.BuyerAddress,
					"seller":          check.SellerAddress,
					"related_address": related,
				},
				DetectedAt:  check.Timestamp,
				BlockHeight: check.BlockHeight,
			}
		}
	}

	return nil
}

// SpoofingCheck checks for potential spoofing
type SpoofingCheck struct {
	// Address is the account address
	Address string `json:"address"`

	// TotalOrders is total orders placed
	TotalOrders uint32 `json:"total_orders"`

	// CancelledOrders is cancelled orders
	CancelledOrders uint32 `json:"cancelled_orders"`

	// AverageOrderLifetime is average order lifetime in blocks
	AverageOrderLifetime int64 `json:"average_order_lifetime"`

	// WindowStart is the start of measurement window
	WindowStart time.Time `json:"window_start"`

	// WindowEnd is the end of measurement window
	WindowEnd time.Time `json:"window_end"`

	// BlockHeight is the current block height
	BlockHeight int64 `json:"block_height"`
}

// DetectSpoofing detects potential spoofing
func (d *ManipulationDetector) DetectSpoofing(check SpoofingCheck) *ViolationRecord {
	if !d.Config.Enabled || !d.Config.Spoofing.Enabled {
		return nil
	}

	// Check minimum orders
	if check.TotalOrders < d.Config.Spoofing.MinOrdersForEvaluation {
		return nil
	}

	// Calculate cancellation rate
	cancellationRate := (check.CancelledOrders * 100) / check.TotalOrders

	if cancellationRate > d.Config.Spoofing.MaxCancellationRatePct {
		return &ViolationRecord{
			Address:     check.Address,
			Type:        ManipulationTypeSpoofing,
			Severity:    ManipulationTypeSpoofing.Severity(),
			Description: fmt.Sprintf("excessive cancellation rate: %d%%", cancellationRate),
			Evidence: map[string]string{
				"total_orders":      fmt.Sprintf("%d", check.TotalOrders),
				"cancelled_orders":  fmt.Sprintf("%d", check.CancelledOrders),
				"cancellation_rate": fmt.Sprintf("%d%%", cancellationRate),
			},
			DetectedAt:  check.WindowEnd,
			BlockHeight: check.BlockHeight,
		}
	}

	// Check average order lifetime
	if check.AverageOrderLifetime < d.Config.Spoofing.MinOrderLifetimeBlocks {
		return &ViolationRecord{
			Address:     check.Address,
			Type:        ManipulationTypeSpoofing,
			Severity:    ManipulationTypeSpoofing.Severity() - 1, // Lower severity
			Description: fmt.Sprintf("short average order lifetime: %d blocks", check.AverageOrderLifetime),
			Evidence: map[string]string{
				"average_lifetime": fmt.Sprintf("%d", check.AverageOrderLifetime),
				"min_expected":     fmt.Sprintf("%d", d.Config.Spoofing.MinOrderLifetimeBlocks),
			},
			DetectedAt:  check.WindowEnd,
			BlockHeight: check.BlockHeight,
		}
	}

	return nil
}

// SafeguardParams holds all safeguard parameters
type SafeguardParams struct {
	// Config is the safeguard configuration
	Config SafeguardConfig `json:"config"`

	// ViolationExpiryBlocks is when violations expire
	ViolationExpiryBlocks int64 `json:"violation_expiry_blocks"`

	// MinTrustScoreForTrading is minimum trust score to trade
	MinTrustScoreForTrading uint32 `json:"min_trust_score_for_trading"`

	// PenaltyCollectorAddress is where penalties are sent
	PenaltyCollectorAddress string `json:"penalty_collector_address"`
}

// DefaultSafeguardParams returns default safeguard parameters
func DefaultSafeguardParams() SafeguardParams {
	return SafeguardParams{
		Config:                  DefaultSafeguardConfig(),
		ViolationExpiryBlocks:   604800, // ~1 week
		MinTrustScoreForTrading: 20,
	}
}

// Validate validates the safeguard parameters
func (p *SafeguardParams) Validate() error {
	if err := p.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if p.MinTrustScoreForTrading > 100 {
		return fmt.Errorf("min_trust_score_for_trading cannot exceed 100")
	}
	return nil
}
