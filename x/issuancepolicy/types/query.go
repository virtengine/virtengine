package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"
)

type QueryPolicyRequest struct {
	PolicyID string `json:"policy_id"`
}

type QueryPolicyResponse struct {
	Policy IssuancePolicy `json:"policy"`
}

type QueryPoliciesRequest struct{}

type QueryPoliciesResponse struct {
	Policies []IssuancePolicy `json:"policies"`
}

type QueryProofMintRecordsRequest struct{}

type QueryProofMintRecordsResponse struct {
	Records []ProofMintRecord `json:"records"`
}

type QueryActivePolicyRequest struct{}

type QueryActivePolicyResponse struct {
	Policy *IssuancePolicy `json:"policy,omitempty"`
}

type QueryCountersRequest struct{}

type QueryCountersResponse struct {
	Counters IssuanceCounters `json:"counters"`
}

type QueryProofMintRecordRequest struct {
	ProofID string `json:"proof_id"`
}

type QueryProofMintRecordResponse struct {
	Record ProofMintRecord `json:"record"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryServer interface {
	Policy(ctx sdk.Context, req *QueryPolicyRequest) (*QueryPolicyResponse, error)
	Policies(ctx sdk.Context, req *QueryPoliciesRequest) (*QueryPoliciesResponse, error)
	ProofMintRecords(ctx sdk.Context, req *QueryProofMintRecordsRequest) (*QueryProofMintRecordsResponse, error)
	ActivePolicy(ctx sdk.Context, req *QueryActivePolicyRequest) (*QueryActivePolicyResponse, error)
	Counters(ctx sdk.Context, req *QueryCountersRequest) (*QueryCountersResponse, error)
	ProofMintRecord(ctx sdk.Context, req *QueryProofMintRecordRequest) (*QueryProofMintRecordResponse, error)
	Params(ctx sdk.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
}

func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	_ = s
	_ = srv
}
