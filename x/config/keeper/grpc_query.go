package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/config/types"
)

// GRPCQuerier provides gRPC query capabilities
type GRPCQuerier struct {
	Keeper Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// ApprovedClient returns a single approved client
func (q GRPCQuerier) ApprovedClient(ctx sdk.Context, req *types.QueryApprovedClientRequest) (*types.QueryApprovedClientResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidClientID.Wrap("request cannot be nil")
	}

	if req.ClientID == "" {
		return nil, types.ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	client, found := q.Keeper.GetClient(ctx, req.ClientID)
	if !found {
		return nil, types.ErrClientNotFound.Wrapf("client %s not found", req.ClientID)
	}

	return &types.QueryApprovedClientResponse{
		Client: client,
	}, nil
}

// ApprovedClients returns all approved clients
func (q GRPCQuerier) ApprovedClients(ctx sdk.Context, req *types.QueryApprovedClientsRequest) (*types.QueryApprovedClientsResponse, error) {
	clients := q.Keeper.ListClients(ctx)
	return &types.QueryApprovedClientsResponse{
		Clients: clients,
	}, nil
}

// ApprovedClientsByStatus returns all clients with a specific status
func (q GRPCQuerier) ApprovedClientsByStatus(ctx sdk.Context, req *types.QueryApprovedClientsByStatusRequest) (*types.QueryApprovedClientsByStatusResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidClientStatus.Wrap("request cannot be nil")
	}

	if !req.Status.IsValid() {
		return nil, types.ErrInvalidClientStatus.Wrapf("invalid status: %s", req.Status)
	}

	clients := q.Keeper.ListClientsByStatus(ctx, req.Status)
	return &types.QueryApprovedClientsByStatusResponse{
		Clients: clients,
	}, nil
}

// ValidateClientSignature validates a client signature
func (q GRPCQuerier) ValidateClientSignature(ctx sdk.Context, req *types.QueryValidateClientSignatureRequest) (*types.QueryValidateClientSignatureResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidSignature.Wrap("request cannot be nil")
	}

	if req.ClientID == "" {
		return nil, types.ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if len(req.Signature) == 0 {
		return nil, types.ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	if len(req.PayloadHash) == 0 {
		return nil, types.ErrInvalidPayloadHash.Wrap("payload_hash cannot be empty")
	}

	err := q.Keeper.ValidateClientSignature(ctx, req.ClientID, req.Signature, req.PayloadHash)
	if err != nil {
		return &types.QueryValidateClientSignatureResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &types.QueryValidateClientSignatureResponse{
		Valid:   true,
		Message: "signature is valid",
	}, nil
}

// ValidateClientVersion validates a client version
func (q GRPCQuerier) ValidateClientVersion(ctx sdk.Context, req *types.QueryValidateClientVersionRequest) (*types.QueryValidateClientVersionResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidVersionConstraint.Wrap("request cannot be nil")
	}

	if req.ClientID == "" {
		return nil, types.ErrInvalidClientID.Wrap("client_id cannot be empty")
	}

	if req.Version == "" {
		return nil, types.ErrInvalidVersionConstraint.Wrap("version cannot be empty")
	}

	err := q.Keeper.ValidateClientVersion(ctx, req.ClientID, req.Version)
	if err != nil {
		return &types.QueryValidateClientVersionResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &types.QueryValidateClientVersionResponse{
		Valid:   true,
		Message: "version is valid",
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := q.Keeper.GetParams(ctx)
	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
