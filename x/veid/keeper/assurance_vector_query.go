package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Assurance Vector Query Endpoints
// ============================================================================

// QueryAssuranceVector returns an account's current per-factor assurance vector.
//
// An account with no vector returns Found=false and a nil Vector rather than an
// error: absence is a real, meaningful state ("no assurance claim"), and a
// relying party must be able to tell it apart from a vector of zeros. Only a
// malformed request or an invalid address is an error.
func (q GRPCQuerier) QueryAssuranceVector(ctx sdk.Context, req *types.QueryAssuranceVectorRequest) (*types.QueryAssuranceVectorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	// Validate address format, matching the other score queries.
	if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	vector, recency, found := q.GetAssuranceVectorRecency(ctx, req.AccountAddress)
	if !found {
		return &types.QueryAssuranceVectorResponse{
			Vector:  nil,
			Found:   false,
			Recency: nil,
		}, nil
	}

	return &types.QueryAssuranceVectorResponse{
		Vector:       vector,
		Found:        true,
		Recency:      &recency,
		CurrentEpoch: vector.Epoch,
	}, nil
}

// QueryAssuranceVectorHistory returns an account's assurance vectors, newest
// first, so a verifier can inspect superseded epochs.
func (q GRPCQuerier) QueryAssuranceVectorHistory(ctx sdk.Context, req *types.QueryAssuranceVectorHistoryRequest) (*types.QueryAssuranceVectorHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if _, err := sdk.AccAddressFromBech32(req.AccountAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	vectors := q.GetAssuranceVectorHistory(ctx, req.AccountAddress)
	if vectors == nil {
		// Return an empty slice so clients can range over the result without a
		// nil check; the proto conversion turns it into an empty repeated field.
		vectors = []*types.AssuranceVector{}
	}

	return &types.QueryAssuranceVectorHistoryResponse{
		Vectors: vectors,
	}, nil
}
