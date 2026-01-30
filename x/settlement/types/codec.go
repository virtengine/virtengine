package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Escrow messages
	legacy.RegisterAminoMsg(cdc, &MsgCreateEscrow{}, "settlement/MsgCreateEscrow")
	legacy.RegisterAminoMsg(cdc, &MsgActivateEscrow{}, "settlement/MsgActivateEscrow")
	legacy.RegisterAminoMsg(cdc, &MsgReleaseEscrow{}, "settlement/MsgReleaseEscrow")
	legacy.RegisterAminoMsg(cdc, &MsgRefundEscrow{}, "settlement/MsgRefundEscrow")
	legacy.RegisterAminoMsg(cdc, &MsgDisputeEscrow{}, "settlement/MsgDisputeEscrow")

	// Settlement messages
	legacy.RegisterAminoMsg(cdc, &MsgSettleOrder{}, "settlement/MsgSettleOrder")
	legacy.RegisterAminoMsg(cdc, &MsgRecordUsage{}, "settlement/MsgRecordUsage")
	legacy.RegisterAminoMsg(cdc, &MsgAcknowledgeUsage{}, "settlement/MsgAcknowledgeUsage")

	// Reward messages
	legacy.RegisterAminoMsg(cdc, &MsgClaimRewards{}, "settlement/MsgClaimRewards")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		// Escrow messages
		&MsgCreateEscrow{},
		&MsgActivateEscrow{},
		&MsgReleaseEscrow{},
		&MsgRefundEscrow{},
		&MsgDisputeEscrow{},
		// Settlement messages
		&MsgSettleOrder{},
		&MsgRecordUsage{},
		&MsgAcknowledgeUsage{},
		// Reward messages
		&MsgClaimRewards{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &settlementv1.Msg_serviceDesc)
}

// Query request/response types

// QueryEscrowRequest is the request for querying an escrow
type QueryEscrowRequest struct {
	EscrowID string `json:"escrow_id"`
}

// QueryEscrowResponse is the response for querying an escrow
type QueryEscrowResponse struct {
	Escrow *EscrowAccount `json:"escrow"`
}

// QueryEscrowsByOrderRequest is the request for querying escrows by order
type QueryEscrowsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QueryEscrowsByOrderResponse is the response for querying escrows by order
type QueryEscrowsByOrderResponse struct {
	Escrows []EscrowAccount `json:"escrows"`
}

// QueryEscrowsByStateRequest is the request for querying escrows by state
type QueryEscrowsByStateRequest struct {
	State string `json:"state"`
}

// QueryEscrowsByStateResponse is the response for querying escrows by state
type QueryEscrowsByStateResponse struct {
	Escrows []EscrowAccount `json:"escrows"`
}

// QuerySettlementRequest is the request for querying a settlement
type QuerySettlementRequest struct {
	SettlementID string `json:"settlement_id"`
}

// QuerySettlementResponse is the response for querying a settlement
type QuerySettlementResponse struct {
	Settlement *SettlementRecord `json:"settlement"`
}

// QuerySettlementsByOrderRequest is the request for querying settlements by order
type QuerySettlementsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QuerySettlementsByOrderResponse is the response for querying settlements by order
type QuerySettlementsByOrderResponse struct {
	Settlements []SettlementRecord `json:"settlements"`
}

// QueryUsageRecordRequest is the request for querying a usage record
type QueryUsageRecordRequest struct {
	UsageID string `json:"usage_id"`
}

// QueryUsageRecordResponse is the response for querying a usage record
type QueryUsageRecordResponse struct {
	UsageRecord *UsageRecord `json:"usage_record"`
}

// QueryUsageRecordsByOrderRequest is the request for querying usage records by order
type QueryUsageRecordsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QueryUsageRecordsByOrderResponse is the response for querying usage records by order
type QueryUsageRecordsByOrderResponse struct {
	UsageRecords []UsageRecord `json:"usage_records"`
}

// QueryRewardDistributionRequest is the request for querying a reward distribution
type QueryRewardDistributionRequest struct {
	DistributionID string `json:"distribution_id"`
}

// QueryRewardDistributionResponse is the response for querying a reward distribution
type QueryRewardDistributionResponse struct {
	Distribution *RewardDistribution `json:"distribution"`
}

// QueryRewardsByEpochRequest is the request for querying rewards by epoch
type QueryRewardsByEpochRequest struct {
	EpochNumber uint64 `json:"epoch_number"`
}

// QueryRewardsByEpochResponse is the response for querying rewards by epoch
type QueryRewardsByEpochResponse struct {
	Distributions []RewardDistribution `json:"distributions"`
}

// QueryClaimableRewardsRequest is the request for querying claimable rewards
type QueryClaimableRewardsRequest struct {
	Address string `json:"address"`
}

// QueryClaimableRewardsResponse is the response for querying claimable rewards
type QueryClaimableRewardsResponse struct {
	Rewards *ClaimableRewards `json:"rewards"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for querying module parameters
type QueryParamsResponse struct {
	Params Params `json:"params"`
}
