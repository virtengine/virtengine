package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/staking/types"
)

// Querier provides gRPC query handlers for staking.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (q Querier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

func (q Querier) ValidatorPerformance(goCtx context.Context, req *types.QueryValidatorPerformanceRequest) (*types.QueryValidatorPerformanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch := q.resolveEpoch(ctx, req.Epoch)
	perf, found := q.GetValidatorPerformance(ctx, req.ValidatorAddress, epoch)
	if !found {
		return nil, status.Error(codes.NotFound, "validator performance not found")
	}

	return &types.QueryValidatorPerformanceResponse{Performance: perf}, nil
}

func (q Querier) ValidatorPerformances(goCtx context.Context, req *types.QueryValidatorPerformancesRequest) (*types.QueryValidatorPerformancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch := q.resolveEpoch(ctx, req.Epoch)
	var performances []types.ValidatorPerformance
	q.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch {
			performances = append(performances, perf)
		}
		return false
	})

	return &types.QueryValidatorPerformancesResponse{Performances: performances}, nil
}

func (q Querier) ValidatorReward(goCtx context.Context, req *types.QueryValidatorRewardRequest) (*types.QueryValidatorRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch := q.resolveEpoch(ctx, req.Epoch)
	reward, found := q.GetValidatorReward(ctx, req.ValidatorAddress, epoch)
	if !found {
		return nil, status.Error(codes.NotFound, "validator reward not found")
	}

	return &types.QueryValidatorRewardResponse{Reward: reward}, nil
}

func (q Querier) ValidatorRewards(goCtx context.Context, req *types.QueryValidatorRewardsRequest) (*types.QueryValidatorRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var rewards []types.ValidatorReward
	q.WithValidatorRewards(ctx, func(reward types.ValidatorReward) bool {
		if reward.ValidatorAddress == req.ValidatorAddress {
			rewards = append(rewards, reward)
		}
		return false
	})

	return &types.QueryValidatorRewardsResponse{Rewards: rewards}, nil
}

func (q Querier) RewardEpoch(goCtx context.Context, req *types.QueryRewardEpochRequest) (*types.QueryRewardEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Epoch == 0 {
		return nil, types.ErrInvalidEpoch.Wrap("epoch must be > 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch, found := q.GetRewardEpoch(ctx, req.Epoch)
	if !found {
		return nil, status.Error(codes.NotFound, "reward epoch not found")
	}

	return &types.QueryRewardEpochResponse{RewardEpoch: epoch}, nil
}

func (q Querier) SlashRecords(goCtx context.Context, req *types.QuerySlashRecordsRequest) (*types.QuerySlashRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	records := q.GetSlashingRecordsByValidator(ctx, req.ValidatorAddress)
	return &types.QuerySlashRecordsResponse{Records: records}, nil
}

func (q Querier) SigningInfo(goCtx context.Context, req *types.QuerySigningInfoRequest) (*types.QuerySigningInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	info, found := q.GetValidatorSigningInfo(ctx, req.ValidatorAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "signing info not found")
	}

	return &types.QuerySigningInfoResponse{Info: info}, nil
}

func (q Querier) CurrentEpoch(goCtx context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryCurrentEpochResponse{Epoch: q.GetCurrentEpoch(ctx)}, nil
}

func (q Querier) resolveEpoch(ctx sdk.Context, epoch uint64) uint64 {
	if epoch == 0 {
		return q.GetCurrentEpoch(ctx)
	}
	return epoch
}
