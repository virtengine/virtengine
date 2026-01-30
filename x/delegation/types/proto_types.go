// Package types provides type aliases and extensions for the delegation module.
//
// This file imports generated protobuf types from sdk/go/node/delegation/v1 and creates
// type aliases for use throughout x/delegation. This approach ensures:
// - Generated types are the source of truth for on-chain data structures
// - Additional methods (Validate, GetSigners, etc.) can be added via extension functions
// - Backward compatibility with existing keeper code
package types

import (
	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
)

// ============================================================================
// Proto Type Aliases - Enums from types.pb.go
// ============================================================================

// DelegationStatusPB is the protobuf-generated enum for delegation status
type DelegationStatusPB = delegationv1.DelegationStatus

// Proto enum constants for DelegationStatus
const (
	DelegationStatusPBUnspecified = delegationv1.DelegationStatusUnspecified
	DelegationStatusPBActive      = delegationv1.DelegationStatusActive
	DelegationStatusPBUnbonding   = delegationv1.DelegationStatusUnbonding
)

// ============================================================================
// Proto Type Aliases - Data Types from types.pb.go
// ============================================================================

// DelegationPB is the generated proto type for delegation records
type DelegationPB = delegationv1.Delegation

// UnbondingDelegationEntryPB is the generated proto type for unbonding entries
type UnbondingDelegationEntryPB = delegationv1.UnbondingDelegationEntry

// UnbondingDelegationPB is the generated proto type for unbonding delegation records
type UnbondingDelegationPB = delegationv1.UnbondingDelegation

// RedelegationEntryPB is the generated proto type for redelegation entries
type RedelegationEntryPB = delegationv1.RedelegationEntry

// RedelegationPB is the generated proto type for redelegation records
type RedelegationPB = delegationv1.Redelegation

// ValidatorSharesPB is the generated proto type for validator shares records
type ValidatorSharesPB = delegationv1.ValidatorShares

// DelegatorRewardPB is the generated proto type for delegator rewards
type DelegatorRewardPB = delegationv1.DelegatorReward

// ParamsPB is the generated proto type for module params
type ParamsPB = delegationv1.Params

// ============================================================================
// Proto Type Aliases - Message Types from tx.pb.go
// ============================================================================

// MsgDelegatePB is the generated proto type for delegate message
type MsgDelegatePB = delegationv1.MsgDelegate

// MsgDelegateResponsePB is the generated proto response type
type MsgDelegateResponsePB = delegationv1.MsgDelegateResponse

// MsgUndelegatePB is the generated proto type for undelegate message
type MsgUndelegatePB = delegationv1.MsgUndelegate

// MsgUndelegateResponsePB is the generated proto response type
type MsgUndelegateResponsePB = delegationv1.MsgUndelegateResponse

// MsgRedelegatePB is the generated proto type for redelegate message
type MsgRedelegatePB = delegationv1.MsgRedelegate

// MsgRedelegateResponsePB is the generated proto response type
type MsgRedelegateResponsePB = delegationv1.MsgRedelegateResponse

// MsgClaimRewardsPB is the generated proto type for claim rewards message
type MsgClaimRewardsPB = delegationv1.MsgClaimRewards

// MsgClaimRewardsResponsePB is the generated proto response type
type MsgClaimRewardsResponsePB = delegationv1.MsgClaimRewardsResponse

// MsgClaimAllRewardsPB is the generated proto type for claim all rewards message
type MsgClaimAllRewardsPB = delegationv1.MsgClaimAllRewards

// MsgClaimAllRewardsResponsePB is the generated proto response type
type MsgClaimAllRewardsResponsePB = delegationv1.MsgClaimAllRewardsResponse

// MsgUpdateParamsPB is the generated proto type for update params message
type MsgUpdateParamsPB = delegationv1.MsgUpdateParams

// MsgUpdateParamsResponsePB is the generated proto response type
type MsgUpdateParamsResponsePB = delegationv1.MsgUpdateParamsResponse

// ============================================================================
// Proto Type Aliases - Genesis from genesis.pb.go
// ============================================================================

// GenesisStatePB is the generated proto type for genesis state
type GenesisStatePB = delegationv1.GenesisState

// ============================================================================
// Proto Service Descriptors
// ============================================================================

// Msg_serviceDesc is the gRPC service descriptor for the Msg service
var Msg_serviceDesc = delegationv1.Msg_serviceDesc

// MsgServerPB is the generated proto server interface
type MsgServerPB = delegationv1.MsgServer

// RegisterMsgServerPB registers the protobuf-generated MsgServer
var RegisterMsgServerPB = delegationv1.RegisterMsgServer
