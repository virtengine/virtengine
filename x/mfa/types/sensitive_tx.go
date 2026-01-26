package types

import (
	"fmt"
)

// SensitiveTransactionType represents types of transactions that require MFA
type SensitiveTransactionType uint8

const (
	// SensitiveTxUnspecified represents an unspecified transaction type
	SensitiveTxUnspecified SensitiveTransactionType = 0
	// SensitiveTxAccountRecovery represents account recovery operations
	SensitiveTxAccountRecovery SensitiveTransactionType = 1
	// SensitiveTxKeyRotation represents key rotation operations
	SensitiveTxKeyRotation SensitiveTransactionType = 2
	// SensitiveTxLargeWithdrawal represents large token withdrawals
	SensitiveTxLargeWithdrawal SensitiveTransactionType = 3
	// SensitiveTxProviderRegistration represents provider registration
	SensitiveTxProviderRegistration SensitiveTransactionType = 4
	// SensitiveTxValidatorRegistration represents validator registration
	SensitiveTxValidatorRegistration SensitiveTransactionType = 5
	// SensitiveTxHighValueOrder represents high-value marketplace orders
	SensitiveTxHighValueOrder SensitiveTransactionType = 6
	// SensitiveTxRoleAssignment represents role assignment operations
	SensitiveTxRoleAssignment SensitiveTransactionType = 7
	// SensitiveTxGovernanceProposal represents governance proposal creation
	SensitiveTxGovernanceProposal SensitiveTransactionType = 8
	// SensitiveTxPrimaryEmailChange represents primary email changes
	SensitiveTxPrimaryEmailChange SensitiveTransactionType = 9
	// SensitiveTxPhoneNumberChange represents phone number changes
	SensitiveTxPhoneNumberChange SensitiveTransactionType = 10
	// SensitiveTxTwoFactorDisable represents disabling 2FA
	SensitiveTxTwoFactorDisable SensitiveTransactionType = 11
	// SensitiveTxAccountDeletion represents account deletion
	SensitiveTxAccountDeletion SensitiveTransactionType = 12
	// SensitiveTxGovernanceVote represents high-stake governance votes
	SensitiveTxGovernanceVote SensitiveTransactionType = 13
	// SensitiveTxFirstOfferingCreate represents first offering creation by provider
	SensitiveTxFirstOfferingCreate SensitiveTransactionType = 14
	// SensitiveTxTransferToNewAddress represents transfers to new addresses
	SensitiveTxTransferToNewAddress SensitiveTransactionType = 15
	// SensitiveTxMediumWithdrawal represents medium-value withdrawals
	SensitiveTxMediumWithdrawal SensitiveTransactionType = 16
	// SensitiveTxAPIKeyGeneration represents API key generation
	SensitiveTxAPIKeyGeneration SensitiveTransactionType = 17
	// SensitiveTxWebhookConfiguration represents webhook configuration
	SensitiveTxWebhookConfiguration SensitiveTransactionType = 18
)

// SensitiveTransactionTypeNames maps transaction types to human-readable names
var SensitiveTransactionTypeNames = map[SensitiveTransactionType]string{
	SensitiveTxUnspecified:           "unspecified",
	SensitiveTxAccountRecovery:       "account_recovery",
	SensitiveTxKeyRotation:           "key_rotation",
	SensitiveTxLargeWithdrawal:       "large_withdrawal",
	SensitiveTxProviderRegistration:  "provider_registration",
	SensitiveTxValidatorRegistration: "validator_registration",
	SensitiveTxHighValueOrder:        "high_value_order",
	SensitiveTxRoleAssignment:        "role_assignment",
	SensitiveTxGovernanceProposal:    "governance_proposal",
	SensitiveTxPrimaryEmailChange:    "primary_email_change",
	SensitiveTxPhoneNumberChange:     "phone_number_change",
	SensitiveTxTwoFactorDisable:      "two_factor_disable",
	SensitiveTxAccountDeletion:       "account_deletion",
	SensitiveTxGovernanceVote:        "governance_vote",
	SensitiveTxFirstOfferingCreate:   "first_offering_create",
	SensitiveTxTransferToNewAddress:  "transfer_to_new_address",
	SensitiveTxMediumWithdrawal:      "medium_withdrawal",
	SensitiveTxAPIKeyGeneration:      "api_key_generation",
	SensitiveTxWebhookConfiguration:  "webhook_configuration",
}

// String returns the string representation of a SensitiveTransactionType
func (t SensitiveTransactionType) String() string {
	if name, ok := SensitiveTransactionTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid returns true if the transaction type is valid
func (t SensitiveTransactionType) IsValid() bool {
	return t >= SensitiveTxAccountRecovery && t <= SensitiveTxWebhookConfiguration
}

// SensitiveTransactionTypeFromString converts a string to SensitiveTransactionType
func SensitiveTransactionTypeFromString(s string) (SensitiveTransactionType, error) {
	for t, name := range SensitiveTransactionTypeNames {
		if name == s {
			return t, nil
		}
	}
	return SensitiveTxUnspecified, fmt.Errorf("unknown sensitive transaction type: %s", s)
}

// TransactionRiskLevel represents the risk level of a transaction
type TransactionRiskLevel uint8

const (
	// RiskLevelLow represents low-risk transactions
	RiskLevelLow TransactionRiskLevel = 1
	// RiskLevelMedium represents medium-risk transactions
	RiskLevelMedium TransactionRiskLevel = 2
	// RiskLevelHigh represents high-risk transactions
	RiskLevelHigh TransactionRiskLevel = 3
	// RiskLevelCritical represents critical-risk transactions
	RiskLevelCritical TransactionRiskLevel = 4
)

// GetRiskLevel returns the risk level of a sensitive transaction type
func (t SensitiveTransactionType) GetRiskLevel() TransactionRiskLevel {
	switch t {
	case SensitiveTxAccountRecovery, SensitiveTxKeyRotation,
		SensitiveTxAccountDeletion, SensitiveTxTwoFactorDisable:
		return RiskLevelCritical

	case SensitiveTxValidatorRegistration, SensitiveTxRoleAssignment,
		SensitiveTxLargeWithdrawal, SensitiveTxPrimaryEmailChange,
		SensitiveTxPhoneNumberChange:
		return RiskLevelHigh

	case SensitiveTxProviderRegistration, SensitiveTxGovernanceProposal,
		SensitiveTxHighValueOrder, SensitiveTxGovernanceVote,
		SensitiveTxFirstOfferingCreate, SensitiveTxWebhookConfiguration:
		return RiskLevelMedium

	case SensitiveTxMediumWithdrawal, SensitiveTxTransferToNewAddress,
		SensitiveTxAPIKeyGeneration:
		return RiskLevelLow

	default:
		return RiskLevelLow
	}
}

// IsSingleUse returns true if the transaction requires single-use authorization
func (t SensitiveTransactionType) IsSingleUse() bool {
	switch t {
	case SensitiveTxAccountRecovery, SensitiveTxKeyRotation,
		SensitiveTxAccountDeletion, SensitiveTxTwoFactorDisable,
		SensitiveTxValidatorRegistration, SensitiveTxRoleAssignment,
		SensitiveTxPrimaryEmailChange, SensitiveTxPhoneNumberChange:
		return true
	default:
		return false
	}
}

// GetDefaultSessionDuration returns the default session duration in seconds
func (t SensitiveTransactionType) GetDefaultSessionDuration() int64 {
	switch t.GetRiskLevel() {
	case RiskLevelCritical:
		return 0 // Single use, no duration
	case RiskLevelHigh:
		return 15 * 60 // 15 minutes
	case RiskLevelMedium:
		return 30 * 60 // 30 minutes
	case RiskLevelLow:
		return 60 * 60 // 1 hour
	default:
		return 15 * 60
	}
}

// SensitiveTxConfig represents the configuration for a sensitive transaction type
type SensitiveTxConfig struct {
	// TransactionType is the type of transaction
	TransactionType SensitiveTransactionType `json:"transaction_type"`

	// Enabled indicates if MFA is required for this transaction type
	Enabled bool `json:"enabled"`

	// MinVEIDScore is the minimum VEID score required
	MinVEIDScore uint32 `json:"min_veid_score"`

	// RequiredFactorCombinations are the default factor combinations required
	RequiredFactorCombinations []FactorCombination `json:"required_factor_combinations"`

	// SessionDuration is the authorization session duration in seconds
	SessionDuration int64 `json:"session_duration"`

	// IsSingleUse indicates if the authorization is single-use
	IsSingleUse bool `json:"is_single_use"`

	// AllowTrustedDeviceReduction indicates if trusted devices can reduce MFA
	AllowTrustedDeviceReduction bool `json:"allow_trusted_device_reduction"`

	// ValueThreshold is the value threshold for amount-based transactions
	ValueThreshold string `json:"value_threshold,omitempty"`

	// CooldownPeriod is the cooldown period in seconds for rate-limited operations
	CooldownPeriod int64 `json:"cooldown_period,omitempty"`

	// Description provides a human-readable description
	Description string `json:"description"`
}

// Validate validates the sensitive transaction config
func (c *SensitiveTxConfig) Validate() error {
	if !c.TransactionType.IsValid() {
		return ErrInvalidSensitiveTxType.Wrapf("invalid transaction type: %d", c.TransactionType)
	}

	for i, fc := range c.RequiredFactorCombinations {
		if err := fc.Validate(); err != nil {
			return ErrInvalidSensitiveTxConfig.Wrapf("invalid factor combination[%d]: %v", i, err)
		}
	}

	if c.SessionDuration < 0 {
		return ErrInvalidSensitiveTxConfig.Wrap("session_duration cannot be negative")
	}

	return nil
}

// GetDefaultSensitiveTxConfigs returns the default configurations for all sensitive transactions
func GetDefaultSensitiveTxConfigs() []SensitiveTxConfig {
	return []SensitiveTxConfig{
		// Critical tier
		{
			TransactionType: SensitiveTxAccountRecovery,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeSMS}},
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeEmail}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			Description:                 "Account recovery requires VEID + FIDO2 + SMS/Email",
		},
		{
			TransactionType: SensitiveTxKeyRotation,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			Description:                 "Key rotation requires VEID + FIDO2 + current key proof",
		},
		{
			TransactionType: SensitiveTxAccountDeletion,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeSMS}},
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeEmail}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			CooldownPeriod:              48 * 60 * 60, // 48 hour cooling off
			Description:                 "Account deletion with 48hr cooling off period",
		},
		{
			TransactionType: SensitiveTxTwoFactorDisable,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2, FactorTypeTOTP}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			Description:                 "Disable 2FA requires verification of existing factors",
		},

		// High tier
		{
			TransactionType: SensitiveTxProviderRegistration,
			Enabled:         true,
			MinVEIDScore:    70,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: false,
			Description:                 "Provider registration requires VEID ≥70 + FIDO2",
		},
		{
			TransactionType: SensitiveTxValidatorRegistration,
			Enabled:         true,
			MinVEIDScore:    85,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			Description:                 "Validator registration requires VEID ≥85 + FIDO2 + governance approval",
		},
		{
			TransactionType: SensitiveTxLargeWithdrawal,
			Enabled:         true,
			MinVEIDScore:    70,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: false,
			ValueThreshold:              "10000", // 10,000 VE tokens
			Description:                 "Large withdrawal (>10,000 VE) requires VEID + FIDO2",
		},
		{
			TransactionType: SensitiveTxRoleAssignment,
			Enabled:         true,
			MinVEIDScore:    85,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             0,
			IsSingleUse:                 true,
			AllowTrustedDeviceReduction: false,
			Description:                 "Role assignment requires VEID ≥85 + FIDO2",
		},

		// Medium tier
		{
			TransactionType: SensitiveTxHighValueOrder,
			Enabled:         true,
			MinVEIDScore:    70,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
				{Factors: []FactorType{FactorTypeVEID, FactorTypeTOTP}},
			},
			SessionDuration:             30 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: true,
			ValueThreshold:              "1000", // 1,000 VE tokens
			Description:                 "High-value orders (>1,000 VE) require VEID + strong factor",
		},
		{
			TransactionType: SensitiveTxGovernanceProposal,
			Enabled:         true,
			MinVEIDScore:    70,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeVEID, FactorTypeFIDO2}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: false,
			Description:                 "Governance proposal creation requires VEID + FIDO2",
		},
		{
			TransactionType: SensitiveTxGovernanceVote,
			Enabled:         true,
			MinVEIDScore:    70,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeFIDO2}},
				{Factors: []FactorType{FactorTypeTOTP}},
			},
			SessionDuration:             30 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: true,
			Description:                 "Governance votes on high-stake proposals require MFA",
		},

		// Low tier
		{
			TransactionType: SensitiveTxMediumWithdrawal,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeFIDO2}},
				{Factors: []FactorType{FactorTypeTOTP}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: true,
			ValueThreshold:              "1000", // 1,000-10,000 VE range
			Description:                 "Medium withdrawal (1,000-10,000 VE) requires single strong factor",
		},
		{
			TransactionType: SensitiveTxTransferToNewAddress,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeFIDO2}},
				{Factors: []FactorType{FactorTypeTOTP}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: true,
			Description:                 "First transfer to a new address requires verification",
		},
		{
			TransactionType: SensitiveTxAPIKeyGeneration,
			Enabled:         true,
			MinVEIDScore:    50,
			RequiredFactorCombinations: []FactorCombination{
				{Factors: []FactorType{FactorTypeFIDO2}},
				{Factors: []FactorType{FactorTypeTOTP}},
			},
			SessionDuration:             15 * 60,
			IsSingleUse:                 false,
			AllowTrustedDeviceReduction: true,
			Description:                 "API key generation requires single factor verification",
		},
	}
}

// KnownSensitiveMsgTypes maps Cosmos SDK message type URLs to sensitive transaction types
var KnownSensitiveMsgTypes = map[string]SensitiveTransactionType{
	// Account operations
	"/virtengine.roles.v1.MsgSetAccountState":  SensitiveTxAccountRecovery,
	"/virtengine.auth.v1.MsgRotateKeys":        SensitiveTxKeyRotation,
	"/virtengine.auth.v1.MsgDeleteAccount":     SensitiveTxAccountDeletion,
	"/virtengine.mfa.v1.MsgDisableMFA":         SensitiveTxTwoFactorDisable,
	"/virtengine.auth.v1.MsgUpdateEmail":       SensitiveTxPrimaryEmailChange,
	"/virtengine.auth.v1.MsgUpdatePhone":       SensitiveTxPhoneNumberChange,

	// Role operations
	"/virtengine.roles.v1.MsgAssignRole":   SensitiveTxRoleAssignment,
	"/virtengine.roles.v1.MsgNominateAdmin": SensitiveTxRoleAssignment,

	// Provider/Validator operations
	"/virtengine.provider.v1.MsgCreateProvider":   SensitiveTxProviderRegistration,
	"/cosmos.staking.v1beta1.MsgCreateValidator":  SensitiveTxValidatorRegistration,

	// Governance
	"/cosmos.gov.v1.MsgSubmitProposal":    SensitiveTxGovernanceProposal,
	"/cosmos.gov.v1beta1.MsgSubmitProposal": SensitiveTxGovernanceProposal,
}

// GetSensitiveTransactionType returns the sensitive transaction type for a message type URL
func GetSensitiveTransactionType(msgTypeURL string) (SensitiveTransactionType, bool) {
	if t, ok := KnownSensitiveMsgTypes[msgTypeURL]; ok {
		return t, true
	}
	return SensitiveTxUnspecified, false
}
