package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veidregistry/types"
)

type GRPCQuerier struct {
	Keeper Keeper
}

var _ types.QueryServer = GRPCQuerier{}

func (q GRPCQuerier) Verifier(ctx sdk.Context, req *types.QueryVerifierRequest) (*types.QueryVerifierResponse, error) {
	if req == nil || req.VerifierID == "" {
		return nil, fmt.Errorf("verifier_id cannot be empty")
	}
	verifier, found := q.Keeper.GetVerifierVersion(ctx, req.VerifierID)
	if !found {
		return nil, fmt.Errorf("verifier %s not found", req.VerifierID)
	}
	return &types.QueryVerifierResponse{Verifier: *verifier}, nil
}

func (q GRPCQuerier) Verifiers(ctx sdk.Context, req *types.QueryVerifiersRequest) (*types.QueryVerifiersResponse, error) {
	return &types.QueryVerifiersResponse{Verifiers: q.Keeper.ListVerifierVersions(ctx)}, nil
}

func (q GRPCQuerier) QueuedVerifiers(ctx sdk.Context, req *types.QueryQueuedVerifiersRequest) (*types.QueryQueuedVerifiersResponse, error) {
	return &types.QueryQueuedVerifiersResponse{Verifiers: q.Keeper.ListQueuedVerifierVersions(ctx)}, nil
}

func (q GRPCQuerier) EligibleVerifiers(ctx sdk.Context, req *types.QueryEligibleVerifiersRequest) (*types.QueryEligibleVerifiersResponse, error) {
	return &types.QueryEligibleVerifiersResponse{Verifiers: q.Keeper.EligibleVerifierVersions(ctx)}, nil
}

func (q GRPCQuerier) ActiveVerifier(ctx sdk.Context, req *types.QueryActiveVerifierRequest) (*types.QueryActiveVerifierResponse, error) {
	active, found := q.Keeper.GetActiveVerifier(ctx)
	if !found {
		return &types.QueryActiveVerifierResponse{}, nil
	}
	return &types.QueryActiveVerifierResponse{ActiveVerifier: active}, nil
}

func (q GRPCQuerier) ValidatorReadiness(ctx sdk.Context, req *types.QueryValidatorReadinessRequest) (*types.QueryValidatorReadinessResponse, error) {
	if req == nil || req.VerifierID == "" {
		return nil, fmt.Errorf("verifier_id cannot be empty")
	}
	return &types.QueryValidatorReadinessResponse{Readiness: q.Keeper.ListValidatorReadiness(ctx, req.VerifierID)}, nil
}

func (q GRPCQuerier) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(ctx)}, nil
}
