// Package keeper implements the delegation module keeper.
//
// VE-922: gRPC query server for delegation module
package keeper

import (
	"context"
	"encoding/json"
	"math/big"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
)

// Querier implements the gRPC QueryServer for the delegation module.
type Querier struct {
	Keeper
	delegationv1.UnimplementedQueryServer
}

// NewQuerier returns a new delegation querier.
func NewQuerier(k Keeper) *Querier {
	return &Querier{Keeper: k}
}

var _ delegationv1.QueryServer = (*Querier)(nil)

// Params returns the module parameters.
func (q *Querier) Params(ctx context.Context, req *delegationv1.QueryParamsRequest) (*delegationv1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)
	return &delegationv1.QueryParamsResponse{
		Params: params,
	}, nil
}

// Delegation queries a specific delegation.
func (q *Querier) Delegation(ctx context.Context, req *delegationv1.QueryDelegationRequest) (*delegationv1.QueryDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address and validator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidValidator.Wrap(err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	del, found := q.GetDelegation(sdkCtx, req.DelegatorAddress, req.ValidatorAddress)
	if !found {
		return &delegationv1.QueryDelegationResponse{
			Found: false,
		}, nil
	}

	return &delegationv1.QueryDelegationResponse{
		Delegation: delegationToProto(del),
		Found:      true,
	}, nil
}

// DelegatorDelegations queries all delegations for a delegator.
func (q *Querier) DelegatorDelegations(ctx context.Context, req *delegationv1.QueryDelegatorDelegationsRequest) (*delegationv1.QueryDelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.DelegationPrefix)

	delegations := make([]delegationv1.Delegation, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var del delegationtypes.Delegation
		if err := json.Unmarshal(value, &del); err != nil {
			return false, err
		}

		match := del.DelegatorAddress == req.DelegatorAddress
		if accumulate && match {
			delegations = append(delegations, delegationToProto(del))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryDelegatorDelegationsResponse{
		Delegations: delegations,
		Pagination:  pageRes,
	}, nil
}

// ValidatorDelegations queries all delegations for a validator.
func (q *Querier) ValidatorDelegations(ctx context.Context, req *delegationv1.QueryValidatorDelegationsRequest) (*delegationv1.QueryValidatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidValidator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.DelegationPrefix)

	delegations := make([]delegationv1.Delegation, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var del delegationtypes.Delegation
		if err := json.Unmarshal(value, &del); err != nil {
			return false, err
		}

		match := del.ValidatorAddress == req.ValidatorAddress
		if accumulate && match {
			delegations = append(delegations, delegationToProto(del))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryValidatorDelegationsResponse{
		Delegations: delegations,
		Pagination:  pageRes,
	}, nil
}

// UnbondingDelegation queries a specific unbonding delegation.
func (q *Querier) UnbondingDelegation(ctx context.Context, req *delegationv1.QueryUnbondingDelegationRequest) (*delegationv1.QueryUnbondingDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.UnbondingId == "" {
		return nil, status.Error(codes.InvalidArgument, "unbonding_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ubd, found := q.GetUnbondingDelegation(sdkCtx, req.UnbondingId)
	if !found {
		return &delegationv1.QueryUnbondingDelegationResponse{
			Found: false,
		}, nil
	}

	return &delegationv1.QueryUnbondingDelegationResponse{
		UnbondingDelegation: unbondingDelegationToProto(ubd),
		Found:               true,
	}, nil
}

// DelegatorUnbondingDelegations queries all unbonding delegations for a delegator.
func (q *Querier) DelegatorUnbondingDelegations(ctx context.Context, req *delegationv1.QueryDelegatorUnbondingDelegationsRequest) (*delegationv1.QueryDelegatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.UnbondingDelegationPrefix)

	unbondings := make([]delegationv1.UnbondingDelegation, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var ubd delegationtypes.UnbondingDelegation
		if err := json.Unmarshal(value, &ubd); err != nil {
			return false, err
		}

		match := ubd.DelegatorAddress == req.DelegatorAddress
		if accumulate && match {
			unbondings = append(unbondings, unbondingDelegationToProto(ubd))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingDelegations: unbondings,
		Pagination:           pageRes,
	}, nil
}

// Redelegation queries a specific redelegation.
func (q *Querier) Redelegation(ctx context.Context, req *delegationv1.QueryRedelegationRequest) (*delegationv1.QueryRedelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.RedelegationId == "" {
		return nil, status.Error(codes.InvalidArgument, "redelegation_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	red, found := q.GetRedelegation(sdkCtx, req.RedelegationId)
	if !found {
		return &delegationv1.QueryRedelegationResponse{
			Found: false,
		}, nil
	}

	return &delegationv1.QueryRedelegationResponse{
		Redelegation: redelegationToProto(red),
		Found:        true,
	}, nil
}

// DelegatorRedelegations queries all redelegations for a delegator.
func (q *Querier) DelegatorRedelegations(ctx context.Context, req *delegationv1.QueryDelegatorRedelegationsRequest) (*delegationv1.QueryDelegatorRedelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.RedelegationPrefix)

	redelegations := make([]delegationv1.Redelegation, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var red delegationtypes.Redelegation
		if err := json.Unmarshal(value, &red); err != nil {
			return false, err
		}

		match := red.DelegatorAddress == req.DelegatorAddress
		if accumulate && match {
			redelegations = append(redelegations, redelegationToProto(red))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryDelegatorRedelegationsResponse{
		Redelegations: redelegations,
		Pagination:    pageRes,
	}, nil
}

// Redelegations queries all active redelegations.
func (q *Querier) Redelegations(ctx context.Context, req *delegationv1.QueryRedelegationsRequest) (*delegationv1.QueryRedelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.RedelegationPrefix)
	now := sdkCtx.BlockTime()

	redelegations := make([]delegationv1.Redelegation, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var red delegationtypes.Redelegation
		if err := json.Unmarshal(value, &red); err != nil {
			return false, err
		}

		active := false
		for _, entry := range red.Entries {
			if entry.CompletionTime.After(now) {
				active = true
				break
			}
		}

		if accumulate && active {
			redelegations = append(redelegations, redelegationToProto(red))
		}

		return active, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryRedelegationsResponse{
		Redelegations: redelegations,
		Pagination:    pageRes,
	}, nil
}

// DelegatorRewards queries unclaimed rewards for a delegator from a specific validator.
func (q *Querier) DelegatorRewards(ctx context.Context, req *delegationv1.QueryDelegatorRewardsRequest) (*delegationv1.QueryDelegatorRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address and validator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidValidator.Wrap(err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	rewards := q.GetDelegatorValidatorUnclaimedRewards(sdkCtx, req.DelegatorAddress, req.ValidatorAddress)
	responseRewards := make([]delegationv1.DelegatorReward, 0, len(rewards))
	for _, reward := range rewards {
		responseRewards = append(responseRewards, delegatorRewardToProto(reward))
	}

	totalReward := q.GetDelegatorValidatorTotalRewards(sdkCtx, req.DelegatorAddress, req.ValidatorAddress)
	return &delegationv1.QueryDelegatorRewardsResponse{
		Rewards:     responseRewards,
		TotalReward: totalReward,
	}, nil
}

// HistoricalRewards queries historical rewards for a delegator from a validator.
func (q *Querier) HistoricalRewards(ctx context.Context, req *delegationv1.QueryHistoricalRewardsRequest) (*delegationv1.QueryHistoricalRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" || req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address and validator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidValidator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	startHeight := req.StartHeight
	endHeight := req.EndHeight
	if endHeight != 0 && startHeight > endHeight {
		return nil, status.Error(codes.InvalidArgument, "start_height cannot be greater than end_height")
	}
	if endHeight == 0 {
		endHeight = int64(^uint64(0) >> 1)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.DelegatorRewardsPrefix)

	rewards := make([]delegationv1.DelegatorReward, 0)
	totalReward := big.NewInt(0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var reward delegationtypes.DelegatorReward
		if err := json.Unmarshal(value, &reward); err != nil {
			return false, err
		}

		match := reward.DelegatorAddress == req.DelegatorAddress &&
			reward.ValidatorAddress == req.ValidatorAddress &&
			!reward.Claimed &&
			reward.Height >= startHeight &&
			reward.Height <= endHeight

		if accumulate && match {
			rewards = append(rewards, delegatorRewardToProto(reward))
			totalReward.Add(totalReward, reward.GetRewardBigInt())
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QueryHistoricalRewardsResponse{
		Rewards:     rewards,
		TotalReward: totalReward.String(),
		Pagination:  pageRes,
	}, nil
}

// SlashingEvents queries slashing events for a delegator.
func (q *Querier) SlashingEvents(ctx context.Context, req *delegationv1.QuerySlashingEventsRequest) (*delegationv1.QuerySlashingEventsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	prefixKey := append([]byte{}, delegationtypes.DelegatorSlashingEventPrefix...)
	prefixKey = append(prefixKey, []byte(req.DelegatorAddress+":")...)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), prefixKey)

	events := make([]delegationv1.DelegatorSlashingEvent, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var event delegationtypes.DelegatorSlashingEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return false, err
		}

		if accumulate {
			events = append(events, delegatorSlashingEventToProto(event))
		}
		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &delegationv1.QuerySlashingEventsResponse{
		Events:     events,
		Pagination: pageRes,
	}, nil
}

// DelegatorAllRewards queries all unclaimed rewards for a delegator.
func (q *Querier) DelegatorAllRewards(ctx context.Context, req *delegationv1.QueryDelegatorAllRewardsRequest) (*delegationv1.QueryDelegatorAllRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidDelegator.Wrap(err.Error())
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), delegationtypes.DelegatorRewardsPrefix)

	rewards := make([]delegationv1.DelegatorReward, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var reward delegationtypes.DelegatorReward
		if err := json.Unmarshal(value, &reward); err != nil {
			return false, err
		}

		match := reward.DelegatorAddress == req.DelegatorAddress && !reward.Claimed
		if accumulate && match {
			rewards = append(rewards, delegatorRewardToProto(reward))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	totalReward := q.GetDelegatorTotalRewards(sdkCtx, req.DelegatorAddress)
	return &delegationv1.QueryDelegatorAllRewardsResponse{
		Rewards:     rewards,
		TotalReward: totalReward,
		Pagination:  pageRes,
	}, nil
}

// ValidatorShares queries the total shares for a validator.
func (q *Querier) ValidatorShares(ctx context.Context, req *delegationv1.QueryValidatorSharesRequest) (*delegationv1.QueryValidatorSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator_address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, delegationtypes.ErrInvalidValidator.Wrap(err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	shares, found := q.GetValidatorShares(sdkCtx, req.ValidatorAddress)
	if !found {
		return &delegationv1.QueryValidatorSharesResponse{
			Found: false,
		}, nil
	}

	return &delegationv1.QueryValidatorSharesResponse{
		ValidatorShares: validatorSharesToProto(shares),
		Found:           true,
	}, nil
}

func delegationToProto(del delegationtypes.Delegation) delegationv1.Delegation {
	return delegationv1.Delegation{
		DelegatorAddress: del.DelegatorAddress,
		ValidatorAddress: del.ValidatorAddress,
		Shares:           del.Shares,
		InitialAmount:    del.InitialAmount,
		CreatedAt:        del.CreatedAt,
		UpdatedAt:        del.UpdatedAt,
		Height:           del.Height,
	}
}

func unbondingDelegationToProto(ubd delegationtypes.UnbondingDelegation) delegationv1.UnbondingDelegation {
	entries := make([]delegationv1.UnbondingDelegationEntry, 0, len(ubd.Entries))
	for _, entry := range ubd.Entries {
		entries = append(entries, unbondingEntryToProto(entry))
	}

	return delegationv1.UnbondingDelegation{
		Id:               ubd.ID,
		DelegatorAddress: ubd.DelegatorAddress,
		ValidatorAddress: ubd.ValidatorAddress,
		Entries:          entries,
		CreatedAt:        ubd.CreatedAt,
		Height:           ubd.Height,
	}
}

func unbondingEntryToProto(entry delegationtypes.UnbondingDelegationEntry) delegationv1.UnbondingDelegationEntry {
	return delegationv1.UnbondingDelegationEntry{
		CreationHeight:  entry.CreationHeight,
		CompletionTime:  entry.CompletionTime,
		InitialBalance:  entry.InitialBalance,
		Balance:         entry.Balance,
		UnbondingShares: entry.UnbondingShares,
	}
}

func redelegationToProto(red delegationtypes.Redelegation) delegationv1.Redelegation {
	entries := make([]delegationv1.RedelegationEntry, 0, len(red.Entries))
	for _, entry := range red.Entries {
		entries = append(entries, redelegationEntryToProto(entry))
	}

	return delegationv1.Redelegation{
		Id:                  red.ID,
		DelegatorAddress:    red.DelegatorAddress,
		ValidatorSrcAddress: red.ValidatorSrcAddress,
		ValidatorDstAddress: red.ValidatorDstAddress,
		Entries:             entries,
		CreatedAt:           red.CreatedAt,
		Height:              red.Height,
	}
}

func redelegationEntryToProto(entry delegationtypes.RedelegationEntry) delegationv1.RedelegationEntry {
	return delegationv1.RedelegationEntry{
		CreationHeight: entry.CreationHeight,
		CompletionTime: entry.CompletionTime,
		InitialBalance: entry.InitialBalance,
		SharesDst:      entry.SharesDst,
	}
}

func validatorSharesToProto(shares delegationtypes.ValidatorShares) delegationv1.ValidatorShares {
	return delegationv1.ValidatorShares{
		ValidatorAddress: shares.ValidatorAddress,
		TotalShares:      shares.TotalShares,
		TotalStake:       shares.TotalStake,
		UpdatedAt:        shares.UpdatedAt,
	}
}

func delegatorRewardToProto(reward delegationtypes.DelegatorReward) delegationv1.DelegatorReward {
	return delegationv1.DelegatorReward{
		DelegatorAddress:            reward.DelegatorAddress,
		ValidatorAddress:            reward.ValidatorAddress,
		EpochNumber:                 reward.EpochNumber,
		Reward:                      reward.Reward,
		SharesAtEpoch:               reward.SharesAtEpoch,
		ValidatorTotalSharesAtEpoch: reward.ValidatorTotalSharesAtEpoch,
		CalculatedAt:                reward.CalculatedAt,
		Height:                      reward.Height,
		Claimed:                     reward.Claimed,
		ClaimedAt:                   reward.ClaimedAt,
	}
}

func delegatorSlashingEventToProto(event delegationtypes.DelegatorSlashingEvent) delegationv1.DelegatorSlashingEvent {
	return delegationv1.DelegatorSlashingEvent{
		Id:               event.ID,
		DelegatorAddress: event.DelegatorAddress,
		ValidatorAddress: event.ValidatorAddress,
		SlashFraction:    event.SlashFraction,
		SlashAmount:      event.SlashAmount,
		SharesSlashed:    event.SharesSlashed,
		InfractionHeight: event.InfractionHeight,
		BlockHeight:      event.BlockHeight,
		BlockTime:        event.BlockTime,
	}
}
