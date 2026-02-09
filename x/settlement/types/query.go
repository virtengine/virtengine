package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
)

// QueryServer defines the settlement module's gRPC query service interface.
type QueryServer interface {
	Escrow(context.Context, *QueryEscrowRequest) (*QueryEscrowResponse, error)
	EscrowsByOrder(context.Context, *QueryEscrowsByOrderRequest) (*QueryEscrowsByOrderResponse, error)
	EscrowsByState(context.Context, *QueryEscrowsByStateRequest) (*QueryEscrowsByStateResponse, error)
	Settlement(context.Context, *QuerySettlementRequest) (*QuerySettlementResponse, error)
	SettlementsByOrder(context.Context, *QuerySettlementsByOrderRequest) (*QuerySettlementsByOrderResponse, error)
	UsageRecord(context.Context, *QueryUsageRecordRequest) (*QueryUsageRecordResponse, error)
	UsageRecordsByOrder(context.Context, *QueryUsageRecordsByOrderRequest) (*QueryUsageRecordsByOrderResponse, error)
	UsageSummary(context.Context, *QueryUsageSummaryRequest) (*QueryUsageSummaryResponse, error)
	RewardDistribution(context.Context, *QueryRewardDistributionRequest) (*QueryRewardDistributionResponse, error)
	RewardsByEpoch(context.Context, *QueryRewardsByEpochRequest) (*QueryRewardsByEpochResponse, error)
	RewardHistory(context.Context, *QueryRewardHistoryRequest) (*QueryRewardHistoryResponse, error)
	ClaimableRewards(context.Context, *QueryClaimableRewardsRequest) (*QueryClaimableRewardsResponse, error)
	Payout(context.Context, *QueryPayoutRequest) (*QueryPayoutResponse, error)
	PayoutsByProvider(context.Context, *QueryPayoutsByProviderRequest) (*QueryPayoutsByProviderResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	FiatConversion(context.Context, *QueryFiatConversionRequest) (*QueryFiatConversionResponse, error)
	FiatConversionsByProvider(context.Context, *QueryFiatConversionsByProviderRequest) (*QueryFiatConversionsByProviderResponse, error)
	FiatPayoutPreference(context.Context, *QueryFiatPayoutPreferenceRequest) (*QueryFiatPayoutPreferenceResponse, error)
}

// RegisterQueryServer registers the QueryServer.
// This stub keeps module wiring consistent until protobuf generation is added.
func RegisterQueryServer(s grpc.Server, impl QueryServer) {
	_ = s
	_ = impl
}
