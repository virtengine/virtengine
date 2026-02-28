package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

type GRPCQuerier struct {
	Keeper Keeper
}

var _ types.QueryServer = GRPCQuerier{}

func (q GRPCQuerier) Policy(ctx sdk.Context, req *types.QueryPolicyRequest) (*types.QueryPolicyResponse, error) {
	if req == nil || req.PolicyID == "" {
		return nil, fmt.Errorf("policy_id cannot be empty")
	}
	policy, found := q.Keeper.GetPolicy(ctx, req.PolicyID)
	if !found {
		return nil, fmt.Errorf("policy %s not found", req.PolicyID)
	}
	return &types.QueryPolicyResponse{Policy: *policy}, nil
}

func (q GRPCQuerier) Policies(ctx sdk.Context, req *types.QueryPoliciesRequest) (*types.QueryPoliciesResponse, error) {
	return &types.QueryPoliciesResponse{Policies: q.Keeper.ListPolicies(ctx)}, nil
}

func (q GRPCQuerier) ProofMintRecords(ctx sdk.Context, req *types.QueryProofMintRecordsRequest) (*types.QueryProofMintRecordsResponse, error) {
	return &types.QueryProofMintRecordsResponse{Records: q.Keeper.ListProofMintRecords(ctx)}, nil
}

func (q GRPCQuerier) ActivePolicy(ctx sdk.Context, req *types.QueryActivePolicyRequest) (*types.QueryActivePolicyResponse, error) {
	policy, found := q.Keeper.GetActivePolicy(ctx)
	if !found {
		return &types.QueryActivePolicyResponse{}, nil
	}
	return &types.QueryActivePolicyResponse{Policy: policy}, nil
}

func (q GRPCQuerier) Counters(ctx sdk.Context, req *types.QueryCountersRequest) (*types.QueryCountersResponse, error) {
	return &types.QueryCountersResponse{Counters: q.Keeper.GetCounters(ctx)}, nil
}

func (q GRPCQuerier) ProofMintRecord(ctx sdk.Context, req *types.QueryProofMintRecordRequest) (*types.QueryProofMintRecordResponse, error) {
	if req == nil || req.ProofID == "" {
		return nil, fmt.Errorf("proof_id cannot be empty")
	}
	record, found := q.Keeper.GetProofMintRecord(ctx, req.ProofID)
	if !found {
		return nil, fmt.Errorf("proof record %s not found", req.ProofID)
	}
	return &types.QueryProofMintRecordResponse{Record: *record}, nil
}

func (q GRPCQuerier) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(ctx)}, nil
}
