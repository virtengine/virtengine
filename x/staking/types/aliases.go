// Package types provides type aliases to generated protobuf types for the staking module.
// This file consolidates all type definitions that map to generated protobuf types.
package types

import (
	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// Type aliases to generated protobuf types
// These types have full proto support (ProtoMessage, Marshal, Unmarshal)
type (
	// Genesis and params
	GenesisState = stakingv1.GenesisState
	Params       = stakingv1.Params

	// Domain types used in genesis
	ValidatorPerformance = stakingv1.ValidatorPerformance
	ValidatorSigningInfo = stakingv1.ValidatorSigningInfo
	SlashRecord          = stakingv1.SlashRecord
	RewardEpoch          = stakingv1.RewardEpoch
	ValidatorReward      = stakingv1.ValidatorReward

	// Enums
	SlashReason = stakingv1.SlashReason
	RewardType  = stakingv1.RewardType

	// Transaction messages
	MsgUpdateParams              = stakingv1.MsgUpdateParams
	MsgSlashValidator            = stakingv1.MsgSlashValidator
	MsgUnjailValidator           = stakingv1.MsgUnjailValidator
	MsgRecordPerformance         = stakingv1.MsgRecordPerformance
	MsgUpdateParamsResponse      = stakingv1.MsgUpdateParamsResponse
	MsgSlashValidatorResponse    = stakingv1.MsgSlashValidatorResponse
	MsgUnjailValidatorResponse   = stakingv1.MsgUnjailValidatorResponse
	MsgRecordPerformanceResponse = stakingv1.MsgRecordPerformanceResponse
)

// Re-export SlashReason constants
const (
	SlashReasonUnspecified               = stakingv1.SlashReasonUnspecified
	SlashReasonDoubleSigning             = stakingv1.SlashReasonDoubleSigning
	SlashReasonDowntime                  = stakingv1.SlashReasonDowntime
	SlashReasonInvalidVEIDAttestation    = stakingv1.SlashReasonInvalidVEIDAttestation
	SlashReasonMissedRecomputation       = stakingv1.SlashReasonMissedRecomputation
	SlashReasonInconsistentScore         = stakingv1.SlashReasonInconsistentScore
	SlashReasonExpiredAttestation        = stakingv1.SlashReasonExpiredAttestation
	SlashReasonDebugModeEnabled          = stakingv1.SlashReasonDebugModeEnabled
	SlashReasonNonAllowlistedMeasurement = stakingv1.SlashReasonNonAllowlistedMeasurement
)

// Re-export RewardType constants
const (
	RewardTypeUnspecified      = stakingv1.RewardTypeUnspecified
	RewardTypeBlockProposal    = stakingv1.RewardTypeBlockProposal
	RewardTypeVEIDVerification = stakingv1.RewardTypeVEIDVerification
	RewardTypeUptime           = stakingv1.RewardTypeUptime
	RewardTypeIdentityNetwork  = stakingv1.RewardTypeIdentityNetwork
	RewardTypeStaking          = stakingv1.RewardTypeStaking
)
