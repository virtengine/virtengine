package keeper

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"

	"github.com/virtengine/virtengine/util/query"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Deployments returns deployments based on filters
func (k Querier) Deployments(c context.Context, req *types.QueryDeploymentsRequest) (*types.QueryDeploymentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	} else if req.Pagination != nil && req.Pagination.Offset > 0 && req.Filters.State == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request parameters. if offset is set, filter.state must be provided")
	}

	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	states := make([]byte, 0, 2)

	var searchPrefix []byte

	// setup for case 3 - cross-index search
	// nolint: gocritic
	if len(req.Pagination.Key) > 0 {
		var key []byte
		var err error
		states, searchPrefix, key, _, err = query.DecodePaginationKey(req.Pagination.Key)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Key = key
	} else if req.Filters.State != "" {
		stateVal := v1.Deployment_State(v1.Deployment_State_value[req.Filters.State])

		if req.Filters.State != "" && stateVal == v1.DeploymentStateInvalid {
			return nil, status.Error(codes.InvalidArgument, "invalid state value")
		}

		states = append(states, byte(stateVal))
	} else {
		// request does not have a pagination set. Start from active store
		states = append(states, byte(v1.DeploymentActive))
		states = append(states, byte(v1.DeploymentClosed))
	}

	// Phase 1: Collect matching deployments (without fetching related data)
	var collectedDeployments []v1.Deployment
	var pageRes *sdkquery.PageResponse

	total := uint64(0)

	for idx := range states {
		state := v1.Deployment_State(states[idx])

		var err error
		if idx > 0 {
			req.Pagination.Key = nil
		}

		if len(req.Pagination.Key) == 0 {
			req.Filters.State = state.String()

			searchPrefix, err = deploymentPrefixFromFilter(req.Filters)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		searchStore := prefix.NewStore(ctx.KVStore(k.skey), searchPrefix)

		count := uint64(0)

		pageRes, err = sdkquery.FilteredPaginate(searchStore, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
			var deployment v1.Deployment

			err := k.cdc.Unmarshal(value, &deployment)
			if err != nil {
				return false, err
			}

			// filter deployments with provided filters
			if req.Filters.Accept(deployment, state) {
				if accumulate {
					collectedDeployments = append(collectedDeployments, deployment)
					count++
				}

				return true, nil
			}

			return false, nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Limit -= count
		total += count

		if req.Pagination.Limit == 0 {
			if len(pageRes.NextKey) > 0 {
				pageRes.NextKey, err = query.EncodePaginationKey(states[idx:], searchPrefix, pageRes.NextKey, nil)
				if err != nil {
					pageRes.Total = total
					return &types.QueryDeploymentsResponse{
						Deployments: nil,
						Pagination:  pageRes,
					}, status.Error(codes.Internal, err.Error())
				}
			}

			break
		}
	}

	// Phase 2: Batch fetch related data for all collected deployments
	deployments := make(types.DeploymentResponses, 0, len(collectedDeployments))

	if len(collectedDeployments) > 0 {
		// Collect deployment IDs for batch operations
		deploymentIDs := make([]v1.DeploymentID, len(collectedDeployments))

		for i, d := range collectedDeployments {
			deploymentIDs[i] = d.ID
		}

		// Batch fetch groups for all deployments
		groupsMap := k.GetGroupsBatch(ctx, deploymentIDs)

		// Build responses with batched data
		// Note: Account fetching still happens individually as escrow accounts use different ID structure
		for _, deployment := range collectedDeployments {
			account, err := k.ekeeper.GetAccount(ctx, deployment.ID.ToEscrowAccountID())
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("fetching escrow account for DeploymentID=%s: %v", deployment.ID, err))
			}

			groups := groupsMap[deployment.ID.String()]

			deployments = append(deployments, types.QueryDeploymentResponse{
				Deployment:    deployment,
				Groups:        groups,
				EscrowAccount: account,
			})
		}
	}

	pageRes.Total = total

	return &types.QueryDeploymentsResponse{
		Deployments: deployments,
		Pagination:  pageRes,
	}, nil
}

// Deployment returns deployment details based on DeploymentID
func (k Querier) Deployment(c context.Context, req *types.QueryDeploymentRequest) (*types.QueryDeploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if _, err := sdk.AccAddressFromBech32(req.ID.Owner); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	deployment, found := k.GetDeployment(ctx, req.ID)
	if !found {
		return nil, v1.ErrDeploymentNotFound
	}

	account, err := k.ekeeper.GetAccount(ctx, req.ID.ToEscrowAccountID())
	if err != nil {
		return &types.QueryDeploymentResponse{}, err
	}

	value := &types.QueryDeploymentResponse{
		Deployment:    deployment,
		Groups:        k.GetGroups(ctx, req.ID),
		EscrowAccount: account,
	}

	return value, nil
}

// Group returns group details based on GroupID
func (k Querier) Group(c context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if _, err := sdk.AccAddressFromBech32(req.ID.Owner); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	ctx := sdk.UnwrapSDKContext(c)

	group, found := k.GetGroup(ctx, req.ID)
	if !found {
		return nil, v1.ErrGroupNotFound
	}

	return &types.QueryGroupResponse{Group: group}, nil
}

func (k Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}
