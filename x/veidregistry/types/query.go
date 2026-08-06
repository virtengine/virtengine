package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"
)

type QueryVerifierRequest struct {
	VerifierID string `json:"verifier_id"`
}

type QueryVerifierResponse struct {
	Verifier VerifierVersion `json:"verifier"`
}

type QueryVerifiersRequest struct{}

type QueryVerifiersResponse struct {
	Verifiers []VerifierVersion `json:"verifiers"`
}

type QueryQueuedVerifiersRequest struct{}

type QueryQueuedVerifiersResponse struct {
	Verifiers []VerifierVersion `json:"verifiers"`
}

type QueryEligibleVerifiersRequest struct{}

type QueryEligibleVerifiersResponse struct {
	Verifiers []VerifierVersion `json:"verifiers"`
}

type QueryActiveVerifierRequest struct{}

type QueryActiveVerifierResponse struct {
	ActiveVerifier *ActiveVerifierPointer `json:"active_verifier,omitempty"`
}

type QueryValidatorReadinessRequest struct {
	VerifierID string `json:"verifier_id"`
}

type QueryValidatorReadinessResponse struct {
	Readiness []ValidatorReadiness `json:"readiness"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryServer interface {
	Verifier(ctx sdk.Context, req *QueryVerifierRequest) (*QueryVerifierResponse, error)
	Verifiers(ctx sdk.Context, req *QueryVerifiersRequest) (*QueryVerifiersResponse, error)
	QueuedVerifiers(ctx sdk.Context, req *QueryQueuedVerifiersRequest) (*QueryQueuedVerifiersResponse, error)
	EligibleVerifiers(ctx sdk.Context, req *QueryEligibleVerifiersRequest) (*QueryEligibleVerifiersResponse, error)
	ActiveVerifier(ctx sdk.Context, req *QueryActiveVerifierRequest) (*QueryActiveVerifierResponse, error)
	ValidatorReadiness(ctx sdk.Context, req *QueryValidatorReadinessRequest) (*QueryValidatorReadinessResponse, error)
	Params(ctx sdk.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
}

func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	_ = s
	_ = srv
}
