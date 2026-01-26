package types

// QueryEscrowRequest is the request type for querying an escrow
type QueryEscrowRequest struct {
	EscrowID string `json:"escrow_id"`
}

// QueryEscrowResponse is the response type for querying an escrow
type QueryEscrowResponse struct {
	Escrow *EscrowAccount `json:"escrow,omitempty"`
}

// QueryEscrowsByOrderRequest is the request type for querying escrows by order
type QueryEscrowsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QueryEscrowsByOrderResponse is the response type for querying escrows by order
type QueryEscrowsByOrderResponse struct {
	Escrows []EscrowAccount `json:"escrows"`
}

// QueryEscrowsByStateRequest is the request type for querying escrows by state
type QueryEscrowsByStateRequest struct {
	State string `json:"state"`
}

// QueryEscrowsByStateResponse is the response type for querying escrows by state
type QueryEscrowsByStateResponse struct {
	Escrows []EscrowAccount `json:"escrows"`
}

// QuerySettlementRequest is the request type for querying a settlement
type QuerySettlementRequest struct {
	SettlementID string `json:"settlement_id"`
}

// QuerySettlementResponse is the response type for querying a settlement
type QuerySettlementResponse struct {
	Settlement *SettlementRecord `json:"settlement,omitempty"`
}

// QuerySettlementsByOrderRequest is the request type for querying settlements by order
type QuerySettlementsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QuerySettlementsByOrderResponse is the response type for querying settlements by order
type QuerySettlementsByOrderResponse struct {
	Settlements []SettlementRecord `json:"settlements"`
}

// QueryUsageRecordRequest is the request type for querying a usage record
type QueryUsageRecordRequest struct {
	UsageID string `json:"usage_id"`
}

// QueryUsageRecordResponse is the response type for querying a usage record
type QueryUsageRecordResponse struct {
	UsageRecord *UsageRecord `json:"usage_record,omitempty"`
}

// QueryUsageRecordsByOrderRequest is the request type for querying usage records by order
type QueryUsageRecordsByOrderRequest struct {
	OrderID string `json:"order_id"`
}

// QueryUsageRecordsByOrderResponse is the response type for querying usage records by order
type QueryUsageRecordsByOrderResponse struct {
	UsageRecords []UsageRecord `json:"usage_records"`
}

// QueryRewardDistributionRequest is the request type for querying a reward distribution
type QueryRewardDistributionRequest struct {
	DistributionID string `json:"distribution_id"`
}

// QueryRewardDistributionResponse is the response type for querying a reward distribution
type QueryRewardDistributionResponse struct {
	Distribution *RewardDistribution `json:"distribution,omitempty"`
}

// QueryRewardsByEpochRequest is the request type for querying rewards by epoch
type QueryRewardsByEpochRequest struct {
	EpochNumber uint64 `json:"epoch_number"`
}

// QueryRewardsByEpochResponse is the response type for querying rewards by epoch
type QueryRewardsByEpochResponse struct {
	Distributions []RewardDistribution `json:"distributions"`
}

// QueryClaimableRewardsRequest is the request type for querying claimable rewards
type QueryClaimableRewardsRequest struct {
	Address string `json:"address"`
}

// QueryClaimableRewardsResponse is the response type for querying claimable rewards
type QueryClaimableRewardsResponse struct {
	Rewards *ClaimableRewards `json:"rewards,omitempty"`
}

// QueryParamsRequest is the request type for querying params
type QueryParamsRequest struct{}

// QueryParamsResponse is the response type for querying params
type QueryParamsResponse struct {
	Params Params `json:"params"`
}
