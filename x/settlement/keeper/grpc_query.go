package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pkg.akt.dev/node/x/settlement/types"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// Escrow returns an escrow account by ID
func (q GRPCQuerier) Escrow(ctx sdk.Context, req *types.QueryEscrowRequest) (*types.QueryEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.EscrowID == "" {
		return nil, status.Error(codes.InvalidArgument, "escrow_id cannot be empty")
	}

	escrow, found := q.Keeper.GetEscrow(ctx, req.EscrowID)
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
func (q GRPCQuerier) EscrowsByOrder(ctx sdk.Context, req *types.QueryEscrowsByOrderRequest) (*types.QueryEscrowsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	escrow, found := q.Keeper.GetEscrowByOrder(ctx, req.OrderID)
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
func (q GRPCQuerier) EscrowsByState(ctx sdk.Context, req *types.QueryEscrowsByStateRequest) (*types.QueryEscrowsByStateResponse, error) {
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

	var escrows []types.EscrowAccount
	q.Keeper.WithEscrowsByState(ctx, state, func(escrow types.EscrowAccount) bool {
		escrows = append(escrows, escrow)
		return false
	})

	return &types.QueryEscrowsByStateResponse{
		Escrows: escrows,
	}, nil
}

// Settlement returns a settlement record by ID
func (q GRPCQuerier) Settlement(ctx sdk.Context, req *types.QuerySettlementRequest) (*types.QuerySettlementResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SettlementID == "" {
		return nil, status.Error(codes.InvalidArgument, "settlement_id cannot be empty")
	}

	settlement, found := q.Keeper.GetSettlement(ctx, req.SettlementID)
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
func (q GRPCQuerier) SettlementsByOrder(ctx sdk.Context, req *types.QuerySettlementsByOrderRequest) (*types.QuerySettlementsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	settlements := q.Keeper.GetSettlementsByOrder(ctx, req.OrderID)

	return &types.QuerySettlementsByOrderResponse{
		Settlements: settlements,
	}, nil
}

// UsageRecord returns a usage record by ID
func (q GRPCQuerier) UsageRecord(ctx sdk.Context, req *types.QueryUsageRecordRequest) (*types.QueryUsageRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UsageID == "" {
		return nil, status.Error(codes.InvalidArgument, "usage_id cannot be empty")
	}

	usage, found := q.Keeper.GetUsageRecord(ctx, req.UsageID)
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
func (q GRPCQuerier) UsageRecordsByOrder(ctx sdk.Context, req *types.QueryUsageRecordsByOrderRequest) (*types.QueryUsageRecordsByOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id cannot be empty")
	}

	usages := q.Keeper.GetUsageRecordsByOrder(ctx, req.OrderID)

	return &types.QueryUsageRecordsByOrderResponse{
		UsageRecords: usages,
	}, nil
}

// RewardDistribution returns a reward distribution by ID
func (q GRPCQuerier) RewardDistribution(ctx sdk.Context, req *types.QueryRewardDistributionRequest) (*types.QueryRewardDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DistributionID == "" {
		return nil, status.Error(codes.InvalidArgument, "distribution_id cannot be empty")
	}

	dist, found := q.Keeper.GetRewardDistribution(ctx, req.DistributionID)
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
func (q GRPCQuerier) RewardsByEpoch(ctx sdk.Context, req *types.QueryRewardsByEpochRequest) (*types.QueryRewardsByEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	distributions := q.Keeper.GetRewardsByEpoch(ctx, req.EpochNumber)

	return &types.QueryRewardsByEpochResponse{
		Distributions: distributions,
	}, nil
}

// ClaimableRewards returns claimable rewards for an address
func (q GRPCQuerier) ClaimableRewards(ctx sdk.Context, req *types.QueryClaimableRewardsRequest) (*types.QueryClaimableRewardsResponse, error) {
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

	rewards, found := q.Keeper.GetClaimableRewards(ctx, address)
	if !found {
		return &types.QueryClaimableRewardsResponse{
			Rewards: nil,
		}, nil
	}

	return &types.QueryClaimableRewardsResponse{
		Rewards: &rewards,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
