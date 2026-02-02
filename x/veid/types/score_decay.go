package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// ============================================================================
// Score Decay Types (VE-3026: Trust Score Decay Mechanism)
// ============================================================================

// DecayType defines how the score decay is calculated
type DecayType int

const (
	// DecayTypeLinear applies constant decay: score -= rate per period
	DecayTypeLinear DecayType = iota

	// DecayTypeExponential applies multiplicative decay: score *= (1 - rate) per period
	DecayTypeExponential

	// DecayTypeStepFunction applies threshold-based drops
	DecayTypeStepFunction
)

// String returns the string representation of a DecayType
func (d DecayType) String() string {
	switch d {
	case DecayTypeLinear:
		return "linear"
	case DecayTypeExponential:
		return "exponential"
	case DecayTypeStepFunction:
		return "step_function"
	default:
		return "unknown"
	}
}

// ParseDecayType parses a string into a DecayType
func ParseDecayType(s string) (DecayType, error) {
	switch s {
	case "linear":
		return DecayTypeLinear, nil
	case "exponential":
		return DecayTypeExponential, nil
	case "step_function":
		return DecayTypeStepFunction, nil
	default:
		return DecayTypeLinear, fmt.Errorf("unknown decay type: %s", s)
	}
}

// IsValid checks if the decay type is a valid value
func (d DecayType) IsValid() bool {
	return d >= DecayTypeLinear && d <= DecayTypeStepFunction
}

// StepThreshold defines a threshold for step function decay
type StepThreshold struct {
	// DaysSinceActivity is the number of days since last activity to trigger this step
	DaysSinceActivity int64 `json:"days_since_activity"`

	// ScoreMultiplier is the multiplier applied to the score at this step (0.0-1.0)
	ScoreMultiplier math.LegacyDec `json:"score_multiplier"`
}

// Validate validates a step threshold
func (s StepThreshold) Validate() error {
	if s.DaysSinceActivity < 0 {
		return fmt.Errorf("days_since_activity must be non-negative")
	}
	if s.ScoreMultiplier.IsNegative() || s.ScoreMultiplier.GT(math.LegacyOneDec()) {
		return fmt.Errorf("score_multiplier must be between 0.0 and 1.0")
	}
	return nil
}

// DecayPolicy defines the configuration for score decay
type DecayPolicy struct {
	// PolicyID is the unique identifier for this policy
	PolicyID string `json:"policy_id"`

	// DecayType is the type of decay algorithm to use
	DecayType DecayType `json:"decay_type"`

	// DecayRate is the rate of decay per period
	// For linear: absolute amount to subtract (e.g., 0.01 = 1 point per period)
	// For exponential: fraction to subtract (e.g., 0.05 = 5% per period)
	DecayRate math.LegacyDec `json:"decay_rate"`

	// DecayPeriod is the duration between decay applications
	DecayPeriod time.Duration `json:"decay_period"`

	// MinScore is the minimum score floor (decay stops at this value)
	MinScore math.LegacyDec `json:"min_score"`

	// GracePeriod is the time after verification/activity before decay starts
	GracePeriod time.Duration `json:"grace_period"`

	// LastActivityBonus is the bonus multiplier applied for recent activity
	// E.g., 1.1 means 10% bonus if active within grace period
	LastActivityBonus math.LegacyDec `json:"last_activity_bonus"`

	// StepThresholds defines thresholds for step function decay (only used if DecayType == DecayTypeStepFunction)
	StepThresholds []StepThreshold `json:"step_thresholds,omitempty"`

	// Enabled indicates whether this policy is active
	Enabled bool `json:"enabled"`

	// CreatedAt is when this policy was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this policy was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultDecayPolicy returns a sensible default decay policy
func DefaultDecayPolicy() DecayPolicy {
	return DecayPolicy{
		PolicyID:          "default",
		DecayType:         DecayTypeExponential,
		DecayRate:         math.LegacyNewDecWithPrec(2, 2),  // 2% per period
		DecayPeriod:       24 * time.Hour * 30,              // Monthly decay
		MinScore:          math.LegacyNewDecWithPrec(20, 0), // Floor at 20
		GracePeriod:       24 * time.Hour * 90,              // 90 day grace period
		LastActivityBonus: math.LegacyOneDec(),              // No bonus by default
		Enabled:           true,
		StepThresholds:    nil,
	}
}

// DefaultStepFunctionPolicy returns a default step function decay policy
func DefaultStepFunctionPolicy() DecayPolicy {
	return DecayPolicy{
		PolicyID:          "step_default",
		DecayType:         DecayTypeStepFunction,
		DecayRate:         math.LegacyZeroDec(),             // Not used for step function
		DecayPeriod:       24 * time.Hour,                   // Check daily
		MinScore:          math.LegacyNewDecWithPrec(10, 0), // Floor at 10
		GracePeriod:       24 * time.Hour * 30,              // 30 day grace period
		LastActivityBonus: math.LegacyOneDec(),              // No bonus by default
		Enabled:           true,
		StepThresholds: []StepThreshold{
			{DaysSinceActivity: 30, ScoreMultiplier: math.LegacyNewDecWithPrec(95, 2)},  // 95% at 30 days
			{DaysSinceActivity: 60, ScoreMultiplier: math.LegacyNewDecWithPrec(85, 2)},  // 85% at 60 days
			{DaysSinceActivity: 90, ScoreMultiplier: math.LegacyNewDecWithPrec(70, 2)},  // 70% at 90 days
			{DaysSinceActivity: 180, ScoreMultiplier: math.LegacyNewDecWithPrec(50, 2)}, // 50% at 180 days
			{DaysSinceActivity: 365, ScoreMultiplier: math.LegacyNewDecWithPrec(25, 2)}, // 25% at 365 days
		},
	}
}

// Validate validates the decay policy
func (p DecayPolicy) Validate() error {
	if p.PolicyID == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}

	if !p.DecayType.IsValid() {
		return fmt.Errorf("invalid decay_type: %d", p.DecayType)
	}

	if p.DecayRate.IsNegative() {
		return fmt.Errorf("decay_rate cannot be negative")
	}

	if p.DecayType == DecayTypeExponential && p.DecayRate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("exponential decay_rate must be <= 1.0")
	}

	if p.DecayPeriod <= 0 {
		return fmt.Errorf("decay_period must be positive")
	}

	if p.MinScore.IsNegative() {
		return fmt.Errorf("min_score cannot be negative")
	}

	if p.MinScore.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("min_score cannot exceed 100")
	}

	if p.GracePeriod < 0 {
		return fmt.Errorf("grace_period cannot be negative")
	}

	if p.LastActivityBonus.IsNegative() {
		return fmt.Errorf("last_activity_bonus cannot be negative")
	}

	if p.DecayType == DecayTypeStepFunction {
		if len(p.StepThresholds) == 0 {
			return fmt.Errorf("step_function decay requires at least one step_threshold")
		}
		var lastDays int64 = -1
		for i, step := range p.StepThresholds {
			if err := step.Validate(); err != nil {
				return fmt.Errorf("invalid step_threshold[%d]: %w", i, err)
			}
			if step.DaysSinceActivity <= lastDays {
				return fmt.Errorf("step_thresholds must be in ascending order by days_since_activity")
			}
			lastDays = step.DaysSinceActivity
		}
	}

	return nil
}

// ScoreSnapshot represents the current state of an account's score with decay tracking
type ScoreSnapshot struct {
	// Address is the account address
	Address string `json:"address"`

	// OriginalScore is the score at the time of last verification
	OriginalScore math.LegacyDec `json:"original_score"`

	// CurrentScore is the current effective score after decay
	CurrentScore math.LegacyDec `json:"current_score"`

	// LastDecayAt is when decay was last applied
	LastDecayAt time.Time `json:"last_decay_at"`

	// LastActivityAt is when the account was last active (resets grace period)
	LastActivityAt time.Time `json:"last_activity_at"`

	// LastVerifiedAt is when the account was last verified
	LastVerifiedAt time.Time `json:"last_verified_at"`

	// PolicyID is the decay policy applied to this account
	PolicyID string `json:"policy_id"`

	// DecayPaused indicates if decay is temporarily paused for this account
	DecayPaused bool `json:"decay_paused"`

	// TotalDecayApplied tracks cumulative decay applied
	TotalDecayApplied math.LegacyDec `json:"total_decay_applied"`
}

// NewScoreSnapshot creates a new score snapshot
func NewScoreSnapshot(
	address string,
	score math.LegacyDec,
	policyID string,
	now time.Time,
) *ScoreSnapshot {
	return &ScoreSnapshot{
		Address:           address,
		OriginalScore:     score,
		CurrentScore:      score,
		LastDecayAt:       now,
		LastActivityAt:    now,
		LastVerifiedAt:    now,
		PolicyID:          policyID,
		DecayPaused:       false,
		TotalDecayApplied: math.LegacyZeroDec(),
	}
}

// Validate validates a score snapshot
func (s ScoreSnapshot) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if s.OriginalScore.IsNegative() {
		return fmt.Errorf("original_score cannot be negative")
	}
	if s.CurrentScore.IsNegative() {
		return fmt.Errorf("current_score cannot be negative")
	}
	if s.CurrentScore.GT(s.OriginalScore) && s.TotalDecayApplied.IsPositive() {
		return fmt.Errorf("current_score cannot exceed original_score when decay has been applied")
	}
	if s.PolicyID == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}
	return nil
}

// IsInGracePeriod checks if the account is still within the grace period
func (s ScoreSnapshot) IsInGracePeriod(policy DecayPolicy, now time.Time) bool {
	if policy.GracePeriod == 0 {
		return false
	}
	graceEnd := s.LastActivityAt.Add(policy.GracePeriod)
	return now.Before(graceEnd)
}

// ShouldApplyDecay checks if decay should be applied
func (s ScoreSnapshot) ShouldApplyDecay(policy DecayPolicy, now time.Time) bool {
	if s.DecayPaused {
		return false
	}
	if !policy.Enabled {
		return false
	}
	if s.IsInGracePeriod(policy, now) {
		return false
	}
	// Check if current score is already at or below minimum
	if s.CurrentScore.LTE(policy.MinScore) {
		return false
	}
	// Check if enough time has passed since last decay
	nextDecayAt := s.LastDecayAt.Add(policy.DecayPeriod)
	return now.After(nextDecayAt) || now.Equal(nextDecayAt)
}

// DecayResult contains the result of a decay calculation
type DecayResult struct {
	// PreviousScore is the score before decay
	PreviousScore math.LegacyDec `json:"previous_score"`

	// NewScore is the score after decay
	NewScore math.LegacyDec `json:"new_score"`

	// DecayAmount is how much was decayed
	DecayAmount math.LegacyDec `json:"decay_amount"`

	// PeriodsApplied is how many decay periods were applied
	PeriodsApplied int64 `json:"periods_applied"`

	// ReachedFloor indicates if the minimum score floor was reached
	ReachedFloor bool `json:"reached_floor"`
}

// DecayEvent is emitted when decay is applied to an account
type DecayEvent struct {
	Address       string         `json:"address"`
	PolicyID      string         `json:"policy_id"`
	PreviousScore math.LegacyDec `json:"previous_score"`
	NewScore      math.LegacyDec `json:"new_score"`
	DecayAmount   math.LegacyDec `json:"decay_amount"`
	Reason        string         `json:"reason"`
	BlockHeight   int64          `json:"block_height"`
	Timestamp     time.Time      `json:"timestamp"`
}

// ActivityType defines types of activities that can reset the grace period
type ActivityType string

const (
	// ActivityTypeTransaction indicates a blockchain transaction
	ActivityTypeTransaction ActivityType = "transaction"

	// ActivityTypeVerification indicates a verification event
	ActivityTypeVerification ActivityType = "verification"

	// ActivityTypeLogin indicates a login/authentication event
	ActivityTypeLogin ActivityType = "login"

	// ActivityTypeMarketplace indicates marketplace activity
	ActivityTypeMarketplace ActivityType = "marketplace"

	// ActivityTypeStaking indicates staking activity
	ActivityTypeStaking ActivityType = "staking"

	// ActivityTypeGovernance indicates governance participation
	ActivityTypeGovernance ActivityType = "governance"
)

// AllActivityTypes returns all valid activity types
func AllActivityTypes() []ActivityType {
	return []ActivityType{
		ActivityTypeTransaction,
		ActivityTypeVerification,
		ActivityTypeLogin,
		ActivityTypeMarketplace,
		ActivityTypeStaking,
		ActivityTypeGovernance,
	}
}

// IsValid checks if the activity type is valid
func (a ActivityType) IsValid() bool {
	for _, valid := range AllActivityTypes() {
		if a == valid {
			return true
		}
	}
	return false
}

// ActivityRecord tracks account activity for grace period calculation
type ActivityRecord struct {
	Address      string       `json:"address"`
	ActivityType ActivityType `json:"activity_type"`
	Timestamp    time.Time    `json:"timestamp"`
	BlockHeight  int64        `json:"block_height"`
	TxHash       string       `json:"tx_hash,omitempty"`
}

// Validate validates an activity record
func (a ActivityRecord) Validate() error {
	if a.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if !a.ActivityType.IsValid() {
		return fmt.Errorf("invalid activity_type: %s", a.ActivityType)
	}
	return nil
}
