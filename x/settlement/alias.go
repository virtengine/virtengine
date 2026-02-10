package settlement

import (
	"github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// Type aliases for external use
type (
	// Keeper alias
	Keeper  = keeper.Keeper
	IKeeper = keeper.IKeeper

	// Escrow types
	EscrowAccount    = types.EscrowAccount
	EscrowState      = types.EscrowState
	ReleaseCondition = types.ReleaseCondition

	// Settlement types
	SettlementRecord = types.SettlementRecord
	SettlementType   = types.SettlementType
	UsageRecord      = types.UsageRecord

	// Reward types
	RewardDistribution = types.RewardDistribution
	RewardRecipient    = types.RewardRecipient
	RewardSource       = types.RewardSource
	RewardEntry        = types.RewardEntry
	ClaimableRewards   = types.ClaimableRewards

	// Genesis types
	GenesisState = types.GenesisState
	Params       = types.Params

	// Message types
	MsgCreateEscrow     = types.MsgCreateEscrow
	MsgActivateEscrow   = types.MsgActivateEscrow
	MsgReleaseEscrow    = types.MsgReleaseEscrow
	MsgRefundEscrow     = types.MsgRefundEscrow
	MsgDisputeEscrow    = types.MsgDisputeEscrow
	MsgSettleOrder      = types.MsgSettleOrder
	MsgRecordUsage      = types.MsgRecordUsage
	MsgAcknowledgeUsage = types.MsgAcknowledgeUsage
	MsgClaimRewards     = types.MsgClaimRewards

	// Query types
	QueryEscrowRequest               = types.QueryEscrowRequest
	QueryEscrowResponse              = types.QueryEscrowResponse
	QueryEscrowsByOrderRequest       = types.QueryEscrowsByOrderRequest
	QueryEscrowsByOrderResponse      = types.QueryEscrowsByOrderResponse
	QueryEscrowsByStateRequest       = types.QueryEscrowsByStateRequest
	QueryEscrowsByStateResponse      = types.QueryEscrowsByStateResponse
	QuerySettlementRequest           = types.QuerySettlementRequest
	QuerySettlementResponse          = types.QuerySettlementResponse
	QuerySettlementsByOrderRequest   = types.QuerySettlementsByOrderRequest
	QuerySettlementsByOrderResponse  = types.QuerySettlementsByOrderResponse
	QueryUsageRecordRequest          = types.QueryUsageRecordRequest
	QueryUsageRecordResponse         = types.QueryUsageRecordResponse
	QueryUsageRecordsByOrderRequest  = types.QueryUsageRecordsByOrderRequest
	QueryUsageRecordsByOrderResponse = types.QueryUsageRecordsByOrderResponse
	QueryRewardDistributionRequest   = types.QueryRewardDistributionRequest
	QueryRewardDistributionResponse  = types.QueryRewardDistributionResponse
	QueryRewardsByEpochRequest       = types.QueryRewardsByEpochRequest
	QueryRewardsByEpochResponse      = types.QueryRewardsByEpochResponse
	QueryClaimableRewardsRequest     = types.QueryClaimableRewardsRequest
	QueryClaimableRewardsResponse    = types.QueryClaimableRewardsResponse
	QueryParamsRequest               = types.QueryParamsRequest
	QueryParamsResponse              = types.QueryParamsResponse
)

// Constants aliases
const (
	// Module constants
	RouterKey = types.RouterKey
	StoreKey  = types.StoreKey

	// Escrow states
	EscrowStatePending  = types.EscrowStatePending
	EscrowStateActive   = types.EscrowStateActive
	EscrowStateReleased = types.EscrowStateReleased
	EscrowStateRefunded = types.EscrowStateRefunded
	EscrowStateDisputed = types.EscrowStateDisputed
	EscrowStateExpired  = types.EscrowStateExpired

	// Settlement types
	SettlementTypePeriodic   = types.SettlementTypePeriodic
	SettlementTypeUsageBased = types.SettlementTypeUsageBased
	SettlementTypeFinal      = types.SettlementTypeFinal
	SettlementTypeRefund     = types.SettlementTypeRefund

	// Reward sources
	RewardSourceStaking      = types.RewardSourceStaking
	RewardSourceUsage        = types.RewardSourceUsage
	RewardSourceVerification = types.RewardSourceVerification
	RewardSourceProvider     = types.RewardSourceProvider

	// Release condition types
	ConditionTypeTimelock       = types.ConditionTypeTimelock
	ConditionTypeSignature      = types.ConditionTypeSignature
	ConditionTypeUsageThreshold = types.ConditionTypeUsageThreshold
	ConditionTypeVerification   = types.ConditionTypeVerification
	ConditionTypeMultisig       = types.ConditionTypeMultisig
)

// Variable aliases
var (
	// Key functions
	EscrowKey             = types.EscrowKey
	EscrowByOrderKey      = types.EscrowByOrderKey
	EscrowByStateKey      = types.EscrowByStateKey
	SettlementKey         = types.SettlementKey
	SettlementByOrderKey  = types.SettlementByOrderKey
	UsageRecordKey        = types.UsageRecordKey
	UsageByOrderKey       = types.UsageByOrderKey
	RewardDistributionKey = types.RewardDistributionKey
	RewardByEpochKey      = types.RewardByEpochKey
	ClaimableRewardsKey   = types.ClaimableRewardsKey

	// Codec functions
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
	RegisterInterfaces       = types.RegisterInterfaces

	// Genesis functions
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams       = types.DefaultParams

	// Error variables
	ErrEscrowNotFound         = types.ErrEscrowNotFound
	ErrEscrowExists           = types.ErrEscrowExists
	ErrInvalidStateTransition = types.ErrInvalidStateTransition
	ErrSettlementNotFound     = types.ErrSettlementNotFound
	ErrUsageRecordNotFound    = types.ErrUsageRecordNotFound
	ErrRewardNotFound         = types.ErrRewardNotFound
	ErrInsufficientFunds      = types.ErrInsufficientFunds
	ErrUnauthorized           = types.ErrUnauthorized
)
