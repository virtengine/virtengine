package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
// and implements the generated gRPC query interface.
type GRPCQuerier struct {
	IKeeper
}

var _ settlementv1.QueryServer = GRPCQuerier{}

// Escrow returns an escrow account by ID
func (q GRPCQuerier) Escrow(ctx context.Context, req *settlementv1.QueryEscrowRequest) (*settlementv1.QueryEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.EscrowId == "" {
		return nil, status.Error(codes.InvalidArgument, "escrow_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	escrow, found := q.GetEscrow(sdkCtx, req.EscrowId)
	if !found {
		return &settlementv1.QueryEscrowResponse{Escrow: nil}, nil
	}

	protoEscrow := toProtoEscrowAccount(escrow)
	return &settlementv1.QueryEscrowResponse{Escrow: &protoEscrow}, nil
}

// EscrowsByOrder returns escrows for an order
func (q GRPCQuerier) EscrowsByOrder(ctx context.Context, req *settlementv1.QueryEscrowsByOrderRequest) (*settlementv1.QueryEscrowsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	escrow, found := q.GetEscrowByOrder(sdkCtx, req.OrderId)
	if !found {
		return &settlementv1.QueryEscrowsByOrderResponse{Escrows: []settlementv1.EscrowAccount{}}, nil
	}

	return &settlementv1.QueryEscrowsByOrderResponse{Escrows: []settlementv1.EscrowAccount{toProtoEscrowAccount(escrow)}}, nil
}

// EscrowsByState returns escrows in a specific state
func (q GRPCQuerier) EscrowsByState(ctx context.Context, req *settlementv1.QueryEscrowsByStateRequest) (*settlementv1.QueryEscrowsByStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.State == "" {
		return nil, status.Error(codes.InvalidArgument, "state cannot be empty")
	}

	state := types.EscrowState(req.State)
	if !types.IsValidEscrowState(state) {
		return nil, status.Error(codes.InvalidArgument, "invalid escrow state")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var escrows []settlementv1.EscrowAccount
	q.WithEscrowsByState(sdkCtx, state, func(escrow types.EscrowAccount) bool {
		escrows = append(escrows, toProtoEscrowAccount(escrow))
		return false
	})

	return &settlementv1.QueryEscrowsByStateResponse{Escrows: escrows}, nil
}

// Settlement returns a settlement record by ID
func (q GRPCQuerier) Settlement(ctx context.Context, req *settlementv1.QuerySettlementRequest) (*settlementv1.QuerySettlementResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SettlementId == "" {
		return nil, status.Error(codes.InvalidArgument, "settlement_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	settlement, found := q.GetSettlement(sdkCtx, req.SettlementId)
	if !found {
		return &settlementv1.QuerySettlementResponse{Settlement: nil}, nil
	}

	protoSettlement := toProtoSettlementRecord(settlement)
	return &settlementv1.QuerySettlementResponse{Settlement: &protoSettlement}, nil
}

// SettlementsByOrder returns settlements for an order
func (q GRPCQuerier) SettlementsByOrder(ctx context.Context, req *settlementv1.QuerySettlementsByOrderRequest) (*settlementv1.QuerySettlementsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	settlements := q.GetSettlementsByOrder(sdkCtx, req.OrderId)
	resp := make([]settlementv1.SettlementRecord, 0, len(settlements))
	for _, settlement := range settlements {
		resp = append(resp, toProtoSettlementRecord(settlement))
	}

	return &settlementv1.QuerySettlementsByOrderResponse{Settlements: resp}, nil
}

// UsageRecord returns a usage record by ID
func (q GRPCQuerier) UsageRecord(ctx context.Context, req *settlementv1.QueryUsageRecordRequest) (*settlementv1.QueryUsageRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UsageId == "" {
		return nil, status.Error(codes.InvalidArgument, "usage_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	usage, found := q.GetUsageRecord(sdkCtx, req.UsageId)
	if !found {
		return &settlementv1.QueryUsageRecordResponse{UsageRecord: nil}, nil
	}

	protoUsage := toProtoUsageRecord(usage)
	return &settlementv1.QueryUsageRecordResponse{UsageRecord: &protoUsage}, nil
}

// UsageRecordsByOrder returns usage records for an order
func (q GRPCQuerier) UsageRecordsByOrder(ctx context.Context, req *settlementv1.QueryUsageRecordsByOrderRequest) (*settlementv1.QueryUsageRecordsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	usages := q.GetUsageRecordsByOrder(sdkCtx, req.OrderId)
	resp := make([]settlementv1.UsageRecord, 0, len(usages))
	for _, usage := range usages {
		resp = append(resp, toProtoUsageRecord(usage))
	}

	return &settlementv1.QueryUsageRecordsByOrderResponse{UsageRecords: resp}, nil
}

// UsageSummary returns usage summary for an order/provider and period.
func (q GRPCQuerier) UsageSummary(ctx context.Context, req *settlementv1.QueryUsageSummaryRequest) (*settlementv1.QueryUsageSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var start, end time.Time
	if req.PeriodStart > 0 {
		start = time.Unix(req.PeriodStart, 0).UTC()
	}
	if req.PeriodEnd > 0 {
		end = time.Unix(req.PeriodEnd, 0).UTC()
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	summary, err := q.BuildUsageSummary(sdkCtx, req.OrderId, req.Provider, start, end)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	protoSummary := toProtoUsageSummary(summary)
	return &settlementv1.QueryUsageSummaryResponse{Summary: protoSummary}, nil
}

// RewardDistribution returns a reward distribution by ID
func (q GRPCQuerier) RewardDistribution(ctx context.Context, req *settlementv1.QueryRewardDistributionRequest) (*settlementv1.QueryRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DistributionId == "" {
		return nil, status.Error(codes.InvalidArgument, "distribution_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	dist, found := q.GetRewardDistribution(sdkCtx, req.DistributionId)
	if !found {
		return &settlementv1.QueryRewardDistributionResponse{Distribution: nil}, nil
	}

	protoDist := toProtoRewardDistribution(dist)
	return &settlementv1.QueryRewardDistributionResponse{Distribution: &protoDist}, nil
}

// RewardsByEpoch returns reward distributions for an epoch
func (q GRPCQuerier) RewardsByEpoch(ctx context.Context, req *settlementv1.QueryRewardsByEpochRequest) (*settlementv1.QueryRewardsByEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	distributions := q.GetRewardsByEpoch(sdkCtx, req.EpochNumber)
	resp := make([]settlementv1.RewardDistribution, 0, len(distributions))
	for _, dist := range distributions {
		resp = append(resp, toProtoRewardDistribution(dist))
	}

	return &settlementv1.QueryRewardsByEpochResponse{Distributions: resp}, nil
}

// RewardHistory returns reward history for an address.
func (q GRPCQuerier) RewardHistory(ctx context.Context, req *settlementv1.QueryRewardHistoryRequest) (*settlementv1.QueryRewardHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	entries, err := q.GetRewardHistory(sdkCtx, req.Address, req.Source, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp := make([]settlementv1.RewardHistoryEntry, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, toProtoRewardHistoryEntry(entry))
	}

	return &settlementv1.QueryRewardHistoryResponse{Entries: resp}, nil
}

// ClaimableRewards returns claimable rewards for an address
func (q GRPCQuerier) ClaimableRewards(ctx context.Context, req *settlementv1.QueryClaimableRewardsRequest) (*settlementv1.QueryClaimableRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewards, found := q.GetClaimableRewards(sdkCtx, address)
	if !found {
		return &settlementv1.QueryClaimableRewardsResponse{Rewards: nil}, nil
	}

	protoRewards := toProtoClaimableRewards(rewards)
	return &settlementv1.QueryClaimableRewardsResponse{Rewards: &protoRewards}, nil
}

// Payout returns a payout record by ID.
func (q GRPCQuerier) Payout(ctx context.Context, req *settlementv1.QueryPayoutRequest) (*settlementv1.QueryPayoutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PayoutId == "" {
		return nil, status.Error(codes.InvalidArgument, "payout_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	payout, found := q.GetPayout(sdkCtx, req.PayoutId)
	if !found {
		return &settlementv1.QueryPayoutResponse{Payout: nil}, nil
	}

	protoPayout := toProtoPayoutRecord(payout)
	return &settlementv1.QueryPayoutResponse{Payout: &protoPayout}, nil
}

// PayoutsByProvider returns payouts for a provider.
func (q GRPCQuerier) PayoutsByProvider(ctx context.Context, req *settlementv1.QueryPayoutsByProviderRequest) (*settlementv1.QueryPayoutsByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	payouts := q.GetPayoutsByProvider(sdkCtx, req.Provider)
	resp := make([]settlementv1.PayoutRecord, 0, len(payouts))
	for _, payout := range payouts {
		resp = append(resp, toProtoPayoutRecord(payout))
	}

	return &settlementv1.QueryPayoutsByProviderResponse{Payouts: resp}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(ctx context.Context, req *settlementv1.QueryParamsRequest) (*settlementv1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)

	protoParams := toProtoParams(params)
	return &settlementv1.QueryParamsResponse{Params: protoParams}, nil
}

// FiatConversion returns a fiat conversion by ID.
func (q GRPCQuerier) FiatConversion(ctx context.Context, req *settlementv1.QueryFiatConversionRequest) (*settlementv1.QueryFiatConversionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ConversionId == "" {
		return nil, status.Error(codes.InvalidArgument, "conversion_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	conversion, found := q.GetFiatConversion(sdkCtx, req.ConversionId)
	if !found {
		return &settlementv1.QueryFiatConversionResponse{Conversion: nil}, nil
	}

	protoConversion := toProtoFiatConversionRecord(conversion)
	return &settlementv1.QueryFiatConversionResponse{Conversion: &protoConversion}, nil
}

// FiatConversionsByProvider returns conversions for a provider.
func (q GRPCQuerier) FiatConversionsByProvider(ctx context.Context, req *settlementv1.QueryFiatConversionsByProviderRequest) (*settlementv1.QueryFiatConversionsByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var conversions []settlementv1.FiatConversionRecord
	q.WithFiatConversions(sdkCtx, func(conversion types.FiatConversionRecord) bool {
		if conversion.Provider == req.Provider {
			conversions = append(conversions, toProtoFiatConversionRecord(conversion))
		}
		return false
	})

	return &settlementv1.QueryFiatConversionsByProviderResponse{Conversions: conversions}, nil
}

// FiatPayoutPreference returns payout preference for a provider.
func (q GRPCQuerier) FiatPayoutPreference(ctx context.Context, req *settlementv1.QueryFiatPayoutPreferenceRequest) (*settlementv1.QueryFiatPayoutPreferenceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pref, found := q.GetFiatPayoutPreference(sdkCtx, req.Provider)
	if !found {
		return &settlementv1.QueryFiatPayoutPreferenceResponse{Preference: nil}, nil
	}

	protoPref := toProtoFiatPayoutPreference(pref)
	return &settlementv1.QueryFiatPayoutPreferenceResponse{Preference: &protoPref}, nil
}
