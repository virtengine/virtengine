package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pkg.akt.dev/node/x/veid/types"
)

// Error message constants
const (
	errMsgEmptyRequest           = "empty request"
	errMsgAccountAddressEmpty    = "account address cannot be empty"
	errMsgInvalidAccountAddress  = "invalid account address"
	errMsgScopeIDEmpty           = "scope_id cannot be empty"
)

// GRPCQuerier is used as Keeper will have duplicate methods if used directly
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// IdentityRecord returns the identity record for an address
func (q GRPCQuerier) IdentityRecord(ctx sdk.Context, req *types.QueryIdentityRecordRequest) (*types.QueryIdentityRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	record, found := q.Keeper.GetIdentityRecord(ctx, address)
	if !found {
		return &types.QueryIdentityRecordResponse{
			Record: nil,
		}, nil
	}

	return &types.QueryIdentityRecordResponse{
		Record: &record,
	}, nil
}

// Scope returns a specific scope for an address
func (q GRPCQuerier) Scope(ctx sdk.Context, req *types.QueryScopeRequest) (*types.QueryScopeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if req.ScopeID == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgScopeIDEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	scope, found := q.Keeper.GetScope(ctx, address, req.ScopeID)
	if !found {
		return &types.QueryScopeResponse{
			Scope: nil,
		}, nil
	}

	return &types.QueryScopeResponse{
		Scope: &scope,
	}, nil
}

// ScopesByType returns all scopes of a specific type for an address
func (q GRPCQuerier) ScopesByType(ctx sdk.Context, req *types.QueryScopesByTypeRequest) (*types.QueryScopesByTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	if !types.IsValidScopeType(req.ScopeType) {
		return nil, status.Error(codes.InvalidArgument, "invalid scope type")
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	scopes := q.Keeper.GetScopesByType(ctx, address, req.ScopeType)

	return &types.QueryScopesByTypeResponse{
		Scopes: scopes,
	}, nil
}

// VerificationHistory returns the verification history for an address
func (q GRPCQuerier) VerificationHistory(ctx sdk.Context, req *types.QueryVerificationHistoryRequest) (*types.QueryVerificationHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	if req.AccountAddress == "" {
		return nil, status.Error(codes.InvalidArgument, errMsgAccountAddressEmpty)
	}

	address, err := sdk.AccAddressFromBech32(req.AccountAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errMsgInvalidAccountAddress)
	}

	limit := req.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	events := q.Keeper.GetVerificationHistory(ctx, address, limit)

	return &types.QueryVerificationHistoryResponse{
		Events: events,
	}, nil
}

// ApprovedClients returns all approved clients
func (q GRPCQuerier) ApprovedClients(ctx sdk.Context, req *types.QueryApprovedClientsRequest) (*types.QueryApprovedClientsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var clients []types.ApprovedClient

	q.Keeper.WithApprovedClients(ctx, func(client types.ApprovedClient) bool {
		clients = append(clients, client)
		return false
	})

	return &types.QueryApprovedClientsResponse{
		Clients: clients,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(ctx sdk.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
