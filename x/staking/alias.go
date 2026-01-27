// Package staking provides type aliases for the staking module.
//
// VE-921: Staking rewards module aliases
package staking

import (
	"github.com/virtengine/virtengine/x/staking/keeper"
	"github.com/virtengine/virtengine/x/staking/types"
)

const (
	// ModuleName is the module name
	ModuleName = types.ModuleName

	// StoreKey is the store key
	StoreKey = types.StoreKey

	// RouterKey is the router key
	RouterKey = types.RouterKey
)

// Type aliases for types
type (
	// Keeper aliases
	Keeper           = keeper.Keeper
	IKeeper          = keeper.IKeeper
	PerformanceUpdate = keeper.PerformanceUpdate
	BankKeeper       = keeper.BankKeeper
	VEIDKeeper       = keeper.VEIDKeeper
	StakingKeeper    = keeper.StakingKeeper

	// Types aliases
	GenesisState          = types.GenesisState
	Params                = types.Params
	ValidatorPerformance  = types.ValidatorPerformance
	ValidatorSigningInfo  = types.ValidatorSigningInfo
	SlashRecord           = types.SlashRecord
	SlashReason           = types.SlashReason
	SlashConfig           = types.SlashConfig
	DoubleSignEvidence    = types.DoubleSignEvidence
	InvalidVEIDAttestation = types.InvalidVEIDAttestation
	RewardEpoch           = types.RewardEpoch
	ValidatorReward       = types.ValidatorReward
	RewardType            = types.RewardType
	RewardCalculationInput = types.RewardCalculationInput
	IdentityNetworkRewardInput = types.IdentityNetworkRewardInput

	// Message types
	MsgUpdateParams      = types.MsgUpdateParams
	MsgSlashValidator    = types.MsgSlashValidator
	MsgUnjailValidator   = types.MsgUnjailValidator
	MsgRecordPerformance = types.MsgRecordPerformance
)

// Constants aliases
const (
	// Slash reasons
	SlashReasonDoubleSigning           = types.SlashReasonDoubleSigning
	SlashReasonDowntime                = types.SlashReasonDowntime
	SlashReasonInvalidVEIDAttestation  = types.SlashReasonInvalidVEIDAttestation
	SlashReasonMissedRecomputation     = types.SlashReasonMissedRecomputation
	SlashReasonInconsistentScore       = types.SlashReasonInconsistentScore
	SlashReasonExpiredAttestation      = types.SlashReasonExpiredAttestation
	SlashReasonDebugModeEnabled        = types.SlashReasonDebugModeEnabled
	SlashReasonNonAllowlistedMeasurement = types.SlashReasonNonAllowlistedMeasurement

	// Reward types
	RewardTypeBlockProposal     = types.RewardTypeBlockProposal
	RewardTypeVEIDVerification  = types.RewardTypeVEIDVerification
	RewardTypeUptime            = types.RewardTypeUptime
	RewardTypeIdentityNetwork   = types.RewardTypeIdentityNetwork
	RewardTypeStaking           = types.RewardTypeStaking

	// Performance constants
	MaxPerformanceScore    = types.MaxPerformanceScore
	FixedPointScale        = types.FixedPointScale
	WeightBlockProposal    = types.WeightBlockProposal
	WeightVEIDVerification = types.WeightVEIDVerification
	WeightUptime           = types.WeightUptime
)

// Function aliases
var (
	// Genesis
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams       = types.DefaultParams

	// Types
	NewValidatorPerformance   = types.NewValidatorPerformance
	NewValidatorSigningInfo   = types.NewValidatorSigningInfo
	NewSlashRecord            = types.NewSlashRecord
	NewRewardEpoch            = types.NewRewardEpoch
	NewValidatorReward        = types.NewValidatorReward
	IsValidSlashReason        = types.IsValidSlashReason
	GetSlashConfig            = types.GetSlashConfig
	DefaultSlashConfigs       = types.DefaultSlashConfigs
	CalculateRewards          = types.CalculateRewards
	CalculateIdentityNetworkReward = types.CalculateIdentityNetworkReward

	// Messages
	NewMsgUpdateParams      = types.NewMsgUpdateParams
	NewMsgSlashValidator    = types.NewMsgSlashValidator
	NewMsgUnjailValidator   = types.NewMsgUnjailValidator
	NewMsgRecordPerformance = types.NewMsgRecordPerformance

	// Keeper
	NewKeeper = keeper.NewKeeper
)

// Error aliases
var (
	ErrValidatorNotFound       = types.ErrValidatorNotFound
	ErrInvalidValidator        = types.ErrInvalidValidator
	ErrSlashingAlreadyRecorded = types.ErrSlashingAlreadyRecorded
	ErrInvalidSlashReason      = types.ErrInvalidSlashReason
	ErrInvalidEpoch            = types.ErrInvalidEpoch
	ErrRewardsAlreadyDistributed = types.ErrRewardsAlreadyDistributed
	ErrInvalidParams           = types.ErrInvalidParams
	ErrValidatorJailed         = types.ErrValidatorJailed
	ErrInvalidPerformanceMetric = types.ErrInvalidPerformanceMetric
	ErrInsufficientStake       = types.ErrInsufficientStake
	ErrDoubleSign              = types.ErrDoubleSign
	ErrInvalidAttestation      = types.ErrInvalidAttestation
	ErrDowntime                = types.ErrDowntime
)
