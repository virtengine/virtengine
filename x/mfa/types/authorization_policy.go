// Package types provides types for the MFA module.
//
// VE-221: Authorization policy for high-value purchase thresholds
// This file defines authorization policies that trigger biometric/MFA requirements
// for high-stakes operations like purchases above a user-defined threshold.
package types

import (
	"fmt"
	"time"
)

// AuthorizationPolicyVersion is the current version of the authorization policy format
const AuthorizationPolicyVersion uint32 = 1

// PolicyTriggerType defines what condition triggers the policy
type PolicyTriggerType string

const (
	// PolicyTriggerThreshold triggers when a transaction value exceeds a threshold
	PolicyTriggerThreshold PolicyTriggerType = "threshold"

	// PolicyTriggerCategory triggers for specific transaction categories
	PolicyTriggerCategory PolicyTriggerType = "category"

	// PolicyTriggerRecipient triggers for specific recipient addresses
	PolicyTriggerRecipient PolicyTriggerType = "recipient"

	// PolicyTriggerFrequency triggers based on transaction frequency
	PolicyTriggerFrequency PolicyTriggerType = "frequency"

	// PolicyTriggerTimeWindow triggers based on time-of-day restrictions
	PolicyTriggerTimeWindow PolicyTriggerType = "time_window"
)

// AllPolicyTriggerTypes returns all valid policy trigger types
func AllPolicyTriggerTypes() []PolicyTriggerType {
	return []PolicyTriggerType{
		PolicyTriggerThreshold,
		PolicyTriggerCategory,
		PolicyTriggerRecipient,
		PolicyTriggerFrequency,
		PolicyTriggerTimeWindow,
	}
}

// IsValidPolicyTriggerType checks if a trigger type is valid
func IsValidPolicyTriggerType(t PolicyTriggerType) bool {
	for _, valid := range AllPolicyTriggerTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// AuthorizationRequirement defines what authorization is needed
type AuthorizationRequirement string

const (
	// AuthReqBiometric requires biometric verification (VEID face match)
	AuthReqBiometric AuthorizationRequirement = "biometric"

	// AuthReqMFA requires MFA verification (any configured factors)
	AuthReqMFA AuthorizationRequirement = "mfa"

	// AuthReqBiometricAndMFA requires both biometric and MFA
	AuthReqBiometricAndMFA AuthorizationRequirement = "biometric_and_mfa"

	// AuthReqElevatedMFA requires additional MFA factors beyond normal
	AuthReqElevatedMFA AuthorizationRequirement = "elevated_mfa"

	// AuthReqManualReview requires manual approval from authorized party
	AuthReqManualReview AuthorizationRequirement = "manual_review"
)

// AllAuthorizationRequirements returns all valid authorization requirements
func AllAuthorizationRequirements() []AuthorizationRequirement {
	return []AuthorizationRequirement{
		AuthReqBiometric,
		AuthReqMFA,
		AuthReqBiometricAndMFA,
		AuthReqElevatedMFA,
		AuthReqManualReview,
	}
}

// IsValidAuthorizationRequirement checks if a requirement is valid
func IsValidAuthorizationRequirement(r AuthorizationRequirement) bool {
	for _, valid := range AllAuthorizationRequirements() {
		if r == valid {
			return true
		}
	}
	return false
}

// ThresholdConfig defines threshold-based authorization triggers
type ThresholdConfig struct {
	// Amount is the threshold amount in the smallest token denomination
	Amount uint64 `json:"amount"`

	// Denom is the token denomination (e.g., "uve")
	Denom string `json:"denom"`

	// PerTransaction if true, applies per transaction; if false, applies to rolling window
	PerTransaction bool `json:"per_transaction"`

	// WindowDurationSeconds is the rolling window duration (if PerTransaction is false)
	WindowDurationSeconds int64 `json:"window_duration_seconds,omitempty"`
}

// Validate validates the threshold configuration
func (c *ThresholdConfig) Validate() error {
	if c.Amount == 0 {
		return ErrInvalidPolicy.Wrap("threshold amount must be positive")
	}
	if c.Denom == "" {
		return ErrInvalidPolicy.Wrap("threshold denom cannot be empty")
	}
	if !c.PerTransaction && c.WindowDurationSeconds <= 0 {
		return ErrInvalidPolicy.Wrap("window duration must be positive for rolling window thresholds")
	}
	return nil
}

// FrequencyConfig defines frequency-based authorization triggers
type FrequencyConfig struct {
	// MaxTransactions is the maximum number of transactions in the window
	MaxTransactions uint32 `json:"max_transactions"`

	// WindowDurationSeconds is the time window for counting transactions
	WindowDurationSeconds int64 `json:"window_duration_seconds"`

	// TransactionTypes is the list of transaction types to count (empty = all)
	TransactionTypes []SensitiveTransactionType `json:"transaction_types,omitempty"`
}

// Validate validates the frequency configuration
func (c *FrequencyConfig) Validate() error {
	if c.MaxTransactions == 0 {
		return ErrInvalidPolicy.Wrap("max transactions must be positive")
	}
	if c.WindowDurationSeconds <= 0 {
		return ErrInvalidPolicy.Wrap("window duration must be positive")
	}
	return nil
}

// TimeWindowConfig defines time-of-day based authorization triggers
type TimeWindowConfig struct {
	// StartHourUTC is the start hour (0-23) of the restricted window
	StartHourUTC uint8 `json:"start_hour_utc"`

	// EndHourUTC is the end hour (0-23) of the restricted window
	EndHourUTC uint8 `json:"end_hour_utc"`

	// DaysOfWeek is the list of days (0=Sunday, 6=Saturday) the restriction applies
	// Empty means all days
	DaysOfWeek []uint8 `json:"days_of_week,omitempty"`
}

// Validate validates the time window configuration
func (c *TimeWindowConfig) Validate() error {
	if c.StartHourUTC > 23 {
		return ErrInvalidPolicy.Wrap("start hour must be 0-23")
	}
	if c.EndHourUTC > 23 {
		return ErrInvalidPolicy.Wrap("end hour must be 0-23")
	}
	for _, day := range c.DaysOfWeek {
		if day > 6 {
			return ErrInvalidPolicy.Wrap("day of week must be 0-6")
		}
	}
	return nil
}

// AuthorizationPolicyTrigger defines a single trigger condition
type AuthorizationPolicyTrigger struct {
	// TriggerType identifies what kind of trigger this is
	TriggerType PolicyTriggerType `json:"trigger_type"`

	// Threshold contains threshold configuration (if TriggerType is threshold)
	Threshold *ThresholdConfig `json:"threshold,omitempty"`

	// Frequency contains frequency configuration (if TriggerType is frequency)
	Frequency *FrequencyConfig `json:"frequency,omitempty"`

	// TimeWindow contains time window configuration (if TriggerType is time_window)
	TimeWindow *TimeWindowConfig `json:"time_window,omitempty"`

	// Categories is the list of transaction categories (if TriggerType is category)
	Categories []SensitiveTransactionType `json:"categories,omitempty"`

	// Recipients is the list of recipient addresses (if TriggerType is recipient)
	Recipients []string `json:"recipients,omitempty"`
}

// Validate validates the trigger
func (t *AuthorizationPolicyTrigger) Validate() error {
	if !IsValidPolicyTriggerType(t.TriggerType) {
		return ErrInvalidPolicy.Wrapf("invalid trigger type: %s", t.TriggerType)
	}

	switch t.TriggerType {
	case PolicyTriggerThreshold:
		if t.Threshold == nil {
			return ErrInvalidPolicy.Wrap("threshold config required for threshold trigger")
		}
		return t.Threshold.Validate()
	case PolicyTriggerFrequency:
		if t.Frequency == nil {
			return ErrInvalidPolicy.Wrap("frequency config required for frequency trigger")
		}
		return t.Frequency.Validate()
	case PolicyTriggerTimeWindow:
		if t.TimeWindow == nil {
			return ErrInvalidPolicy.Wrap("time window config required for time_window trigger")
		}
		return t.TimeWindow.Validate()
	case PolicyTriggerCategory:
		if len(t.Categories) == 0 {
			return ErrInvalidPolicy.Wrap("categories required for category trigger")
		}
	case PolicyTriggerRecipient:
		if len(t.Recipients) == 0 {
			return ErrInvalidPolicy.Wrap("recipients required for recipient trigger")
		}
	}

	return nil
}

// AuthorizationPolicy defines a user's authorization requirements for high-stakes actions
type AuthorizationPolicy struct {
	// Version is the policy format version
	Version uint32 `json:"version"`

	// PolicyID is a unique identifier for this policy
	PolicyID string `json:"policy_id"`

	// AccountAddress is the account this policy applies to
	AccountAddress string `json:"account_address"`

	// Enabled indicates if the policy is active
	Enabled bool `json:"enabled"`

	// Triggers defines the conditions that activate this policy
	Triggers []AuthorizationPolicyTrigger `json:"triggers"`

	// Requirements defines what authorization is needed when triggered
	Requirements AuthorizationRequirement `json:"requirements"`

	// RequiredFactors specifies which factor combinations satisfy the requirement
	// Used when Requirements is MFA, ElevatedMFA, or BiometricAndMFA
	RequiredFactors []FactorCombination `json:"required_factors,omitempty"`

	// MinVEIDScore is the minimum VEID score required for biometric authorization
	MinVEIDScore uint32 `json:"min_veid_score,omitempty"`

	// GracePeriodSeconds allows a grace period after last authorization
	GracePeriodSeconds int64 `json:"grace_period_seconds,omitempty"`

	// CreatedAt is when this policy was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this policy was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Description is a user-provided description
	Description string `json:"description,omitempty"`
}

// NewAuthorizationPolicy creates a new authorization policy
func NewAuthorizationPolicy(
	policyID string,
	accountAddress string,
	triggers []AuthorizationPolicyTrigger,
	requirements AuthorizationRequirement,
	createdAt time.Time,
) *AuthorizationPolicy {
	return &AuthorizationPolicy{
		Version:        AuthorizationPolicyVersion,
		PolicyID:       policyID,
		AccountAddress: accountAddress,
		Enabled:        true,
		Triggers:       triggers,
		Requirements:   requirements,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
}

// NewThresholdAuthorizationPolicy creates a simple threshold-based policy
func NewThresholdAuthorizationPolicy(
	policyID string,
	accountAddress string,
	amount uint64,
	denom string,
	requirements AuthorizationRequirement,
	createdAt time.Time,
) *AuthorizationPolicy {
	triggers := []AuthorizationPolicyTrigger{
		{
			TriggerType: PolicyTriggerThreshold,
			Threshold: &ThresholdConfig{
				Amount:         amount,
				Denom:          denom,
				PerTransaction: true,
			},
		},
	}
	return NewAuthorizationPolicy(policyID, accountAddress, triggers, requirements, createdAt)
}

// Validate validates the authorization policy
func (p *AuthorizationPolicy) Validate() error {
	if p.Version == 0 || p.Version > AuthorizationPolicyVersion {
		return ErrInvalidPolicy.Wrapf("unsupported version: %d", p.Version)
	}

	if p.PolicyID == "" {
		return ErrInvalidPolicy.Wrap("policy_id cannot be empty")
	}

	if p.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}

	if len(p.Triggers) == 0 {
		return ErrInvalidPolicy.Wrap("at least one trigger is required")
	}

	for i, trigger := range p.Triggers {
		if err := trigger.Validate(); err != nil {
			return ErrInvalidPolicy.Wrapf("invalid trigger[%d]: %v", i, err)
		}
	}

	if !IsValidAuthorizationRequirement(p.Requirements) {
		return ErrInvalidPolicy.Wrapf("invalid requirements: %s", p.Requirements)
	}

	if p.CreatedAt.IsZero() {
		return ErrInvalidPolicy.Wrap("created_at cannot be zero")
	}

	return nil
}

// HasThresholdTrigger returns true if the policy has a threshold trigger
func (p *AuthorizationPolicy) HasThresholdTrigger() bool {
	for _, t := range p.Triggers {
		if t.TriggerType == PolicyTriggerThreshold {
			return true
		}
	}
	return false
}

// GetThresholdConfig returns the threshold configuration if present
func (p *AuthorizationPolicy) GetThresholdConfig() *ThresholdConfig {
	for _, t := range p.Triggers {
		if t.TriggerType == PolicyTriggerThreshold && t.Threshold != nil {
			return t.Threshold
		}
	}
	return nil
}

// String returns a string representation
func (p *AuthorizationPolicy) String() string {
	return fmt.Sprintf("AuthorizationPolicy{ID: %s, Account: %s, Enabled: %t, Requirements: %s}",
		p.PolicyID, p.AccountAddress, p.Enabled, p.Requirements)
}

// AuthorizationResult represents the result of an authorization check
type AuthorizationResult struct {
	// Authorized indicates if the action is authorized
	Authorized bool `json:"authorized"`

	// RequiredAction specifies what action is needed (if not authorized)
	RequiredAction AuthorizationRequirement `json:"required_action,omitempty"`

	// TriggeredPolicies lists the policy IDs that were triggered
	TriggeredPolicies []string `json:"triggered_policies,omitempty"`

	// TriggerReasons describes why each policy was triggered
	TriggerReasons []string `json:"trigger_reasons,omitempty"`

	// SessionID is the authorization session ID (if MFA session was used)
	SessionID string `json:"session_id,omitempty"`

	// AuthorizedAt is when authorization was granted
	AuthorizedAt *time.Time `json:"authorized_at,omitempty"`

	// ExpiresAt is when the authorization expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// NewAuthorizationResult creates a new authorization result
func NewAuthorizationResult(authorized bool) *AuthorizationResult {
	return &AuthorizationResult{
		Authorized:        authorized,
		TriggeredPolicies: make([]string, 0),
		TriggerReasons:    make([]string, 0),
	}
}

// AddTriggeredPolicy adds a triggered policy to the result
func (r *AuthorizationResult) AddTriggeredPolicy(policyID string, reason string) {
	r.TriggeredPolicies = append(r.TriggeredPolicies, policyID)
	r.TriggerReasons = append(r.TriggerReasons, reason)
}

// AuthorizationAuditEvent represents an audit event for authorization
type AuthorizationAuditEvent struct {
	// EventID is a unique identifier for this event
	EventID string `json:"event_id"`

	// AccountAddress is the account that triggered the policy
	AccountAddress string `json:"account_address"`

	// TransactionType is the type of transaction
	TransactionType SensitiveTransactionType `json:"transaction_type"`

	// TransactionValue is the value of the transaction (if applicable)
	TransactionValue uint64 `json:"transaction_value,omitempty"`

	// TransactionDenom is the denomination of the value
	TransactionDenom string `json:"transaction_denom,omitempty"`

	// TriggeredPolicyIDs lists which policies were triggered
	TriggeredPolicyIDs []string `json:"triggered_policy_ids"`

	// TriggerTypes lists the trigger types that matched
	TriggerTypes []PolicyTriggerType `json:"trigger_types"`

	// RequiredAuthorization is what was required
	RequiredAuthorization AuthorizationRequirement `json:"required_authorization"`

	// AuthorizationSatisfied indicates if authorization was satisfied
	AuthorizationSatisfied bool `json:"authorization_satisfied"`

	// SatisfiedBy indicates how authorization was satisfied (factor types used)
	SatisfiedBy []string `json:"satisfied_by,omitempty"`

	// FailureReason is the reason authorization failed (if applicable)
	FailureReason string `json:"failure_reason,omitempty"`

	// BlockHeight is when this event occurred
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when this event occurred
	Timestamp time.Time `json:"timestamp"`
}

// NewAuthorizationAuditEvent creates a new audit event
func NewAuthorizationAuditEvent(
	eventID string,
	accountAddress string,
	txType SensitiveTransactionType,
	blockHeight int64,
	timestamp time.Time,
) *AuthorizationAuditEvent {
	return &AuthorizationAuditEvent{
		EventID:            eventID,
		AccountAddress:     accountAddress,
		TransactionType:    txType,
		TriggeredPolicyIDs: make([]string, 0),
		TriggerTypes:       make([]PolicyTriggerType, 0),
		BlockHeight:        blockHeight,
		Timestamp:          timestamp,
	}
}

// Validate validates the audit event
func (e *AuthorizationAuditEvent) Validate() error {
	if e.EventID == "" {
		return ErrInvalidPolicy.Wrap("event_id cannot be empty")
	}
	if e.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account address cannot be empty")
	}
	if e.Timestamp.IsZero() {
		return ErrInvalidPolicy.Wrap("timestamp cannot be zero")
	}
	return nil
}
