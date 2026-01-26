// Package marketplace provides types for the marketplace on-chain module.
//
// VE-302: Marketplace sensitive action gating via MFA module
// This file implements MFA gating hooks for sensitive marketplace actions.
package marketplace

import (
	"fmt"
	"time"
)

// MarketplaceActionType represents types of marketplace actions that may require MFA
type MarketplaceActionType uint8

const (
	// ActionUnspecified represents an unspecified action
	ActionUnspecified MarketplaceActionType = 0

	// ActionPlaceOrder represents placing a new order
	ActionPlaceOrder MarketplaceActionType = 1

	// ActionModifyOrder represents modifying an existing order
	ActionModifyOrder MarketplaceActionType = 2

	// ActionCancelOrder represents cancelling an order
	ActionCancelOrder MarketplaceActionType = 3

	// ActionTerminateAllocation represents terminating an allocation
	ActionTerminateAllocation MarketplaceActionType = 4

	// ActionWithdrawFunds represents withdrawing funds from marketplace
	ActionWithdrawFunds MarketplaceActionType = 5

	// ActionCreateOffering represents creating a new offering
	ActionCreateOffering MarketplaceActionType = 6

	// ActionModifyOffering represents modifying an offering
	ActionModifyOffering MarketplaceActionType = 7

	// ActionTerminateOffering represents terminating an offering
	ActionTerminateOffering MarketplaceActionType = 8

	// ActionAcceptBid represents accepting a bid
	ActionAcceptBid MarketplaceActionType = 9

	// ActionPlaceBid represents placing a bid
	ActionPlaceBid MarketplaceActionType = 10

	// ActionSettlement represents settlement operations
	ActionSettlement MarketplaceActionType = 11

	// ActionKeyRotation represents key rotation for marketplace
	ActionKeyRotation MarketplaceActionType = 12
)

// MarketplaceActionTypeNames maps action types to human-readable names
var MarketplaceActionTypeNames = map[MarketplaceActionType]string{
	ActionUnspecified:         "unspecified",
	ActionPlaceOrder:          "place_order",
	ActionModifyOrder:         "modify_order",
	ActionCancelOrder:         "cancel_order",
	ActionTerminateAllocation: "terminate_allocation",
	ActionWithdrawFunds:       "withdraw_funds",
	ActionCreateOffering:      "create_offering",
	ActionModifyOffering:      "modify_offering",
	ActionTerminateOffering:   "terminate_offering",
	ActionAcceptBid:           "accept_bid",
	ActionPlaceBid:            "place_bid",
	ActionSettlement:          "settlement",
	ActionKeyRotation:         "key_rotation",
}

// String returns the string representation of an action type
func (t MarketplaceActionType) String() string {
	if name, ok := MarketplaceActionTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid returns true if the action type is valid
func (t MarketplaceActionType) IsValid() bool {
	return t >= ActionPlaceOrder && t <= ActionKeyRotation
}

// MFARequirementLevel represents the MFA requirement level
type MFARequirementLevel uint8

const (
	// MFANotRequired indicates MFA is not required
	MFANotRequired MFARequirementLevel = 0

	// MFAOptional indicates MFA is optional but recommended
	MFAOptional MFARequirementLevel = 1

	// MFARequiredSingleFactor indicates single-factor MFA is required
	MFARequiredSingleFactor MFARequirementLevel = 2

	// MFARequiredMultiFactor indicates multi-factor MFA is required
	MFARequiredMultiFactor MFARequirementLevel = 3

	// MFARequiredBiometric indicates biometric verification is required
	MFARequiredBiometric MFARequirementLevel = 4
)

// MFARequirementLevelNames maps requirement levels to names
var MFARequirementLevelNames = map[MFARequirementLevel]string{
	MFANotRequired:          "not_required",
	MFAOptional:             "optional",
	MFARequiredSingleFactor: "single_factor",
	MFARequiredMultiFactor:  "multi_factor",
	MFARequiredBiometric:    "biometric",
}

// String returns the string representation of an MFA requirement level
func (l MFARequirementLevel) String() string {
	if name, ok := MFARequirementLevelNames[l]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", l)
}

// MFAActionConfig defines MFA requirements for a specific action type
type MFAActionConfig struct {
	// ActionType is the action this config applies to
	ActionType MarketplaceActionType `json:"action_type"`

	// RequirementLevel is the base MFA requirement level
	RequirementLevel MFARequirementLevel `json:"requirement_level"`

	// ValueThreshold is the value threshold above which MFA is required (0 = always)
	ValueThreshold uint64 `json:"value_threshold,omitempty"`

	// TrustedDeviceReducesRequirement indicates if trusted device reduces requirement
	TrustedDeviceReducesRequirement bool `json:"trusted_device_reduces_requirement"`

	// RecentMFASatisfies indicates if recent MFA (within session) satisfies requirement
	RecentMFASatisfies bool `json:"recent_mfa_satisfies"`

	// SessionWindowSeconds is how long a recent MFA is valid
	SessionWindowSeconds int64 `json:"session_window_seconds"`

	// AllowedFactorTypes specifies which factor types are accepted
	AllowedFactorTypes []string `json:"allowed_factor_types,omitempty"`
}

// DefaultMFAActionConfigs returns default MFA configurations for marketplace actions
func DefaultMFAActionConfigs() []MFAActionConfig {
	return []MFAActionConfig{
		{
			ActionType:                      ActionPlaceOrder,
			RequirementLevel:                MFAOptional,
			ValueThreshold:                  1000000, // High value threshold
			TrustedDeviceReducesRequirement: true,
			RecentMFASatisfies:              true,
			SessionWindowSeconds:            3600, // 1 hour
		},
		{
			ActionType:                      ActionCancelOrder,
			RequirementLevel:                MFARequiredSingleFactor,
			TrustedDeviceReducesRequirement: true,
			RecentMFASatisfies:              true,
			SessionWindowSeconds:            3600,
		},
		{
			ActionType:                      ActionTerminateAllocation,
			RequirementLevel:                MFARequiredSingleFactor,
			TrustedDeviceReducesRequirement: true,
			RecentMFASatisfies:              true,
			SessionWindowSeconds:            1800, // 30 minutes
		},
		{
			ActionType:                      ActionWithdrawFunds,
			RequirementLevel:                MFARequiredMultiFactor,
			TrustedDeviceReducesRequirement: false, // No reduction for withdrawals
			RecentMFASatisfies:              false, // Always require fresh MFA
			SessionWindowSeconds:            0,
		},
		{
			ActionType:                      ActionCreateOffering,
			RequirementLevel:                MFARequiredSingleFactor,
			TrustedDeviceReducesRequirement: true,
			RecentMFASatisfies:              true,
			SessionWindowSeconds:            3600,
		},
		{
			ActionType:                      ActionKeyRotation,
			RequirementLevel:                MFARequiredMultiFactor,
			TrustedDeviceReducesRequirement: false,
			RecentMFASatisfies:              false,
			SessionWindowSeconds:            0,
		},
		{
			ActionType:                      ActionSettlement,
			RequirementLevel:                MFARequiredSingleFactor,
			TrustedDeviceReducesRequirement: true,
			RecentMFASatisfies:              true,
			SessionWindowSeconds:            1800,
		},
	}
}

// MFAGatingContext provides context for MFA gating decisions
type MFAGatingContext struct {
	// ActionType is the action being performed
	ActionType MarketplaceActionType `json:"action_type"`

	// AccountAddress is the account performing the action
	AccountAddress string `json:"account_address"`

	// TransactionValue is the value of the transaction (if applicable)
	TransactionValue uint64 `json:"transaction_value,omitempty"`

	// IsTrustedDevice indicates if the request is from a trusted device
	IsTrustedDevice bool `json:"is_trusted_device"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// LastMFAVerifiedAt is when MFA was last verified
	LastMFAVerifiedAt *time.Time `json:"last_mfa_verified_at,omitempty"`

	// AccountMFAPolicy is the account's MFA policy
	AccountMFAPolicy *AccountMFAPolicy `json:"account_mfa_policy,omitempty"`

	// OfferingRequiresMFA indicates if the offering requires MFA
	OfferingRequiresMFA bool `json:"offering_requires_mfa"`
}

// AccountMFAPolicy represents an account's MFA policy for marketplace
type AccountMFAPolicy struct {
	// Enabled indicates if MFA is enabled
	Enabled bool `json:"enabled"`

	// EnrolledFactors lists enrolled factor types
	EnrolledFactors []string `json:"enrolled_factors"`

	// RequireForHighValue indicates if MFA is required for high-value transactions
	RequireForHighValue bool `json:"require_for_high_value"`

	// HighValueThreshold is the threshold for high-value transactions
	HighValueThreshold uint64 `json:"high_value_threshold"`

	// TrustedDevices lists trusted device fingerprints
	TrustedDevices []string `json:"trusted_devices"`
}

// MFAGatingResult represents the result of MFA gating check
type MFAGatingResult struct {
	// Required indicates if MFA is required
	Required bool `json:"required"`

	// RequirementLevel is the required MFA level
	RequirementLevel MFARequirementLevel `json:"requirement_level"`

	// Satisfied indicates if MFA requirements are satisfied
	Satisfied bool `json:"satisfied"`

	// ChallengeID is the MFA challenge ID if a challenge is needed
	ChallengeID string `json:"challenge_id,omitempty"`

	// AcceptedFactorTypes lists accepted factor types
	AcceptedFactorTypes []string `json:"accepted_factor_types,omitempty"`

	// Reason explains why MFA is required or satisfied
	Reason string `json:"reason"`

	// SessionValid indicates if a recent MFA session is valid
	SessionValid bool `json:"session_valid"`

	// TrustedDeviceUsed indicates if trusted device rule was applied
	TrustedDeviceUsed bool `json:"trusted_device_used"`
}

// MFAGatingChecker performs MFA gating checks
type MFAGatingChecker struct {
	configs map[MarketplaceActionType]*MFAActionConfig
}

// NewMFAGatingChecker creates a new MFA gating checker with default configs
func NewMFAGatingChecker() *MFAGatingChecker {
	checker := &MFAGatingChecker{
		configs: make(map[MarketplaceActionType]*MFAActionConfig),
	}

	for _, config := range DefaultMFAActionConfigs() {
		c := config // Create a copy
		checker.configs[config.ActionType] = &c
	}

	return checker
}

// SetConfig sets the config for an action type
func (c *MFAGatingChecker) SetConfig(config *MFAActionConfig) {
	c.configs[config.ActionType] = config
}

// Check performs MFA gating check for the given context
func (c *MFAGatingChecker) Check(ctx *MFAGatingContext) *MFAGatingResult {
	config, ok := c.configs[ctx.ActionType]
	if !ok {
		// No config means MFA not required
		return &MFAGatingResult{
			Required:  false,
			Satisfied: true,
			Reason:    "no MFA configuration for action",
		}
	}

	result := &MFAGatingResult{
		RequirementLevel: config.RequirementLevel,
	}

	// Determine if MFA is required based on config and context
	required := c.determineIfRequired(config, ctx)
	result.Required = required

	if !required {
		result.Satisfied = true
		result.Reason = "MFA not required for this action"
		return result
	}

	// Check if offering requires MFA
	if ctx.OfferingRequiresMFA {
		result.Required = true
		result.RequirementLevel = MFARequiredSingleFactor
		if config.RequirementLevel > MFARequiredSingleFactor {
			result.RequirementLevel = config.RequirementLevel
		}
	}

	// Check if trusted device reduces requirement
	if config.TrustedDeviceReducesRequirement && ctx.IsTrustedDevice {
		if result.RequirementLevel == MFARequiredSingleFactor {
			result.Required = false
			result.Satisfied = true
			result.TrustedDeviceUsed = true
			result.Reason = "trusted device reduces MFA requirement"
			return result
		}
		result.TrustedDeviceUsed = true
	}

	// Check if recent MFA satisfies requirement
	if config.RecentMFASatisfies && ctx.LastMFAVerifiedAt != nil {
		windowDuration := time.Duration(config.SessionWindowSeconds) * time.Second
		if time.Since(*ctx.LastMFAVerifiedAt) < windowDuration {
			result.Satisfied = true
			result.SessionValid = true
			result.Reason = "recent MFA verification still valid"
			return result
		}
	}

	// MFA is required and not satisfied
	result.Satisfied = false
	result.AcceptedFactorTypes = config.AllowedFactorTypes
	result.Reason = fmt.Sprintf("MFA required for %s (level: %s)", ctx.ActionType, result.RequirementLevel)

	return result
}

// determineIfRequired determines if MFA is required based on config and context
func (c *MFAGatingChecker) determineIfRequired(config *MFAActionConfig, ctx *MFAGatingContext) bool {
	if config.RequirementLevel == MFANotRequired {
		return false
	}

	if config.RequirementLevel >= MFARequiredSingleFactor {
		return true
	}

	// For optional, check value threshold
	if config.RequirementLevel == MFAOptional && config.ValueThreshold > 0 {
		if ctx.TransactionValue >= config.ValueThreshold {
			return true
		}
	}

	// Check account policy
	if ctx.AccountMFAPolicy != nil && ctx.AccountMFAPolicy.RequireForHighValue {
		if ctx.TransactionValue >= ctx.AccountMFAPolicy.HighValueThreshold {
			return true
		}
	}

	return false
}

// MFAAuditRecord records MFA verification for audit purposes
type MFAAuditRecord struct {
	// ActionType is the action that triggered MFA
	ActionType MarketplaceActionType `json:"action_type"`

	// AccountAddress is the account that performed MFA
	AccountAddress string `json:"account_address"`

	// ChallengeID is the MFA challenge ID
	ChallengeID string `json:"challenge_id"`

	// FactorTypesUsed lists the factor types used
	FactorTypesUsed []string `json:"factor_types_used"`

	// Success indicates if MFA was successful
	Success bool `json:"success"`

	// Timestamp is when MFA was performed
	Timestamp time.Time `json:"timestamp"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// TrustedDevice indicates if the device was trusted
	TrustedDevice bool `json:"trusted_device"`

	// TransactionRef is a reference to the related transaction
	TransactionRef string `json:"transaction_ref,omitempty"`

	// FailureReason is set if MFA failed
	FailureReason string `json:"failure_reason,omitempty"`

	// AttemptCount is the number of attempts made
	AttemptCount uint32 `json:"attempt_count"`

	// IPAddressHash is a hashed representation of the IP (privacy-preserving)
	IPAddressHash string `json:"ip_address_hash,omitempty"`
}

// NewMFAAuditRecord creates a new MFA audit record
func NewMFAAuditRecord(actionType MarketplaceActionType, accountAddress, challengeID string) *MFAAuditRecord {
	return &MFAAuditRecord{
		ActionType:      actionType,
		AccountAddress:  accountAddress,
		ChallengeID:     challengeID,
		FactorTypesUsed: make([]string, 0),
		Timestamp:       time.Now().UTC(),
		AttemptCount:    1,
	}
}

// RecordSuccess records a successful MFA verification
func (r *MFAAuditRecord) RecordSuccess(factorsUsed []string) {
	r.Success = true
	r.FactorTypesUsed = factorsUsed
}

// RecordFailure records a failed MFA verification
func (r *MFAAuditRecord) RecordFailure(reason string) {
	r.Success = false
	r.FailureReason = reason
}

// MFAVerificationState represents the current state of MFA verification for a marketplace action
type MFAVerificationState uint8

const (
	// MFAStateNotRequired indicates MFA is not required
	MFAStateNotRequired MFAVerificationState = 0

	// MFAStatePending indicates MFA is pending
	MFAStatePending MFAVerificationState = 1

	// MFAStateVerified indicates MFA is verified
	MFAStateVerified MFAVerificationState = 2

	// MFAStateFailed indicates MFA verification failed
	MFAStateFailed MFAVerificationState = 3

	// MFAStateExpired indicates MFA challenge expired
	MFAStateExpired MFAVerificationState = 4
)

// MFAVerificationStateNames maps states to human-readable names
var MFAVerificationStateNames = map[MFAVerificationState]string{
	MFAStateNotRequired: "not_required",
	MFAStatePending:     "pending",
	MFAStateVerified:    "verified",
	MFAStateFailed:      "failed",
	MFAStateExpired:     "expired",
}

// String returns the string representation of an MFA verification state
func (s MFAVerificationState) String() string {
	if name, ok := MFAVerificationStateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}
