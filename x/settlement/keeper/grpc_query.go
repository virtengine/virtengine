package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// Escrow returns an escrow account by ID
func (q GRPCQuerier) Escrow(ctx context.Context, req *types.QueryEscrowRequest) (*types.QueryEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.EscrowID == "" {
		return nil, status.Error(codes.InvalidArgument, "escrow_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	escrow, found := q.GetEscrow(sdkCtx, req.EscrowID)
	if !found {
		return &types.QueryEscrowResponse{
			Escrow: nil,
		}, nil
	}

	return &types.QueryEscrowResponse{
		Escrow: &escrow,
	}, nil
}

// EscrowsByOrder returns escrows for an order
func (q GRPCQuerier) EscrowsByOrder(ctx context.Context, req *types.QueryEscrowsByOrderRequest) (*types.QueryEscrowsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	escrow, found := q.GetEscrowByOrder(sdkCtx, req.OrderID)
	if !found {
		return &types.QueryEscrowsByOrderResponse{
			Escrows: []types.EscrowAccount{},
		}, nil
	}

	return &types.QueryEscrowsByOrderResponse{
		Escrows: []types.EscrowAccount{escrow},
	}, nil
}

// EscrowsByState returns escrows in a specific state
func (q GRPCQuerier) EscrowsByState(ctx context.Context, req *types.QueryEscrowsByStateRequest) (*types.QueryEscrowsByStateResponse, error) {
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
	var escrows []types.EscrowAccount
	q.WithEscrowsByState(sdkCtx, state, func(escrow types.EscrowAccount) bool {
		escrows = append(escrows, escrow)
		return false
	})

	return &types.QueryEscrowsByStateResponse{
		Escrows: escrows,
	}, nil
}

// Settlement returns a settlement record by ID
func (q GRPCQuerier) Settlement(ctx context.Context, req *types.QuerySettlementRequest) (*types.QuerySettlementResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SettlementID == "" {
		return nil, status.Error(codes.InvalidArgument, "settlement_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	settlement, found := q.GetSettlement(sdkCtx, req.SettlementID)
	if !found {
		return &types.QuerySettlementResponse{
			Settlement: nil,
		}, nil
	}

	return &types.QuerySettlementResponse{
		Settlement: &settlement,
	}, nil
}

// SettlementsByOrder returns settlements for an order
func (q GRPCQuerier) SettlementsByOrder(ctx context.Context, req *types.QuerySettlementsByOrderRequest) (*types.QuerySettlementsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	settlements := q.GetSettlementsByOrder(sdkCtx, req.OrderID)

	return &types.QuerySettlementsByOrderResponse{
		Settlements: settlements,
	}, nil
}

// UsageRecord returns a usage record by ID
func (q GRPCQuerier) UsageRecord(ctx context.Context, req *types.QueryUsageRecordRequest) (*types.QueryUsageRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UsageID == "" {
		return nil, status.Error(codes.InvalidArgument, "usage_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	usage, found := q.GetUsageRecord(sdkCtx, req.UsageID)
	if !found {
		return &types.QueryUsageRecordResponse{
			UsageRecord: nil,
		}, nil
	}

	return &types.QueryUsageRecordResponse{
		UsageRecord: &usage,
	}, nil
}

// UsageRecordsByOrder returns usage records for an order
func (q GRPCQuerier) UsageRecordsByOrder(ctx context.Context, req *types.QueryUsageRecordsByOrderRequest) (*types.QueryUsageRecordsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	usages := q.GetUsageRecordsByOrder(sdkCtx, req.OrderID)

	return &types.QueryUsageRecordsByOrderResponse{
		UsageRecords: usages,
	}, nil
}

// UsageSummary returns usage summary for an order/provider and period.
func (q GRPCQuerier) UsageSummary(ctx context.Context, req *types.QueryUsageSummaryRequest) (*types.QueryUsageSummaryResponse, error) {
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
	summary, err := q.BuildUsageSummary(sdkCtx, req.OrderID, req.Provider, start, end)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.QueryUsageSummaryResponse{
		Summary: summary,
	}, nil
}

// RewardDistribution returns a reward distribution by ID
func (q GRPCQuerier) RewardDistribution(ctx context.Context, req *types.QueryRewardDistributionRequest) (*types.QueryRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DistributionID == "" {
		return nil, status.Error(codes.InvalidArgument, "distribution_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	dist, found := q.GetRewardDistribution(sdkCtx, req.DistributionID)
	if !found {
		return &types.QueryRewardDistributionResponse{
			Distribution: nil,
		}, nil
	}

	return &types.QueryRewardDistributionResponse{
		Distribution: &dist,
	}, nil
}

// RewardsByEpoch returns reward distributions for an epoch
func (q GRPCQuerier) RewardsByEpoch(ctx context.Context, req *types.QueryRewardsByEpochRequest) (*types.QueryRewardsByEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	distributions := q.GetRewardsByEpoch(sdkCtx, req.EpochNumber)

	return &types.QueryRewardsByEpochResponse{
		Distributions: distributions,
	}, nil
}

// RewardHistory returns reward history for an address.
func (q GRPCQuerier) RewardHistory(ctx context.Context, req *types.QueryRewardHistoryRequest) (*types.QueryRewardHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	entries, err := q.GetRewardHistory(sdkCtx, req.Address, req.Source, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.QueryRewardHistoryResponse{
		Entries: entries,
	}, nil
}

// ClaimableRewards returns claimable rewards for an address
func (q GRPCQuerier) ClaimableRewards(ctx context.Context, req *types.QueryClaimableRewardsRequest) (*types.QueryClaimableRewardsResponse, error) {
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
		return &types.QueryClaimableRewardsResponse{
			Rewards: nil,
		}, nil
	}

	return &types.QueryClaimableRewardsResponse{
		Rewards: &rewards,
	}, nil
}

// Payout returns a payout record by ID.
func (q GRPCQuerier) Payout(ctx context.Context, req *types.QueryPayoutRequest) (*types.QueryPayoutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.PayoutID == "" {
		return nil, status.Error(codes.InvalidArgument, "payout_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	payout, found := q.GetPayout(sdkCtx, req.PayoutID)
	if !found {
		return &types.QueryPayoutResponse{
			Payout: nil,
		}, nil
	}

	return &types.QueryPayoutResponse{
		Payout: &payout,
	}, nil
}

// PayoutsByProvider returns payouts for a provider.
func (q GRPCQuerier) PayoutsByProvider(ctx context.Context, req *types.QueryPayoutsByProviderRequest) (*types.QueryPayoutsByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	payouts := q.GetPayoutsByProvider(sdkCtx, req.Provider)

	return &types.QueryPayoutsByProviderResponse{
		Payouts: payouts,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// FiatConversion returns a fiat conversion by ID.
func (q GRPCQuerier) FiatConversion(ctx context.Context, req *types.QueryFiatConversionRequest) (*types.QueryFiatConversionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ConversionID == "" {
		return nil, status.Error(codes.InvalidArgument, "conversion_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	conversion, found := q.GetFiatConversion(sdkCtx, req.ConversionID)
	if !found {
		return &types.QueryFiatConversionResponse{Conversion: nil}, nil
	}
	return &types.QueryFiatConversionResponse{Conversion: &conversion}, nil
}

// FiatConversionsByProvider returns conversions for a provider.
func (q GRPCQuerier) FiatConversionsByProvider(ctx context.Context, req *types.QueryFiatConversionsByProviderRequest) (*types.QueryFiatConversionsByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	var conversions []types.FiatConversionRecord
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	q.WithFiatConversions(sdkCtx, func(conversion types.FiatConversionRecord) bool {
		if conversion.Provider == req.Provider {
			conversions = append(conversions, conversion)
		}
		return false
	})

	return &types.QueryFiatConversionsByProviderResponse{Conversions: conversions}, nil
}

// FiatPayoutPreference returns payout preference for a provider.
func (q GRPCQuerier) FiatPayoutPreference(ctx context.Context, req *types.QueryFiatPayoutPreferenceRequest) (*types.QueryFiatPayoutPreferenceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pref, found := q.GetFiatPayoutPreference(sdkCtx, req.Provider)
	if !found {
		return &types.QueryFiatPayoutPreferenceResponse{Preference: nil}, nil
	}
	return &types.QueryFiatPayoutPreferenceResponse{Preference: &pref}, nil
}
