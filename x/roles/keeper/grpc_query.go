package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pkg.akt.dev/node/x/roles/types"
)

// Error message constant
const errMsgEmptyRequest = "empty request"

// Querier is used as Keeper will have duplicate methods if used directly
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// AccountRoles returns all roles assigned to an account
func (q Querier) AccountRoles(req *types.QueryAccountRolesRequest) (*types.QueryAccountRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Note: This method needs a context, but the interface doesn't provide one
	// In a real implementation, this would use gRPC context
	return &types.QueryAccountRolesResponse{
		Address: req.Address,
		Roles:   nil, // Would be populated with ctx
	}, nil
}

// RoleMembers returns all accounts with a specific role
func (q Querier) RoleMembers(req *types.QueryRoleMembersRequest) (*types.QueryRoleMembersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	role, err := types.RoleFromString(req.Role)
	if err != nil {
		return nil, types.ErrInvalidRole.Wrap(err.Error())
	}

	return &types.QueryRoleMembersResponse{
		Role:    role.String(),
		Members: nil, // Would be populated with ctx
	}, nil
}

// AccountState returns the state of an account
func (q Querier) AccountState(req *types.QueryAccountStateRequest) (*types.QueryAccountStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	return &types.QueryAccountStateResponse{
		AccountState: types.AccountStateRecord{},
	}, nil
}

// GenesisAccounts returns all genesis accounts
func (q Querier) GenesisAccounts(req *types.QueryGenesisAccountsRequest) (*types.QueryGenesisAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	return &types.QueryGenesisAccountsResponse{
		Addresses: nil, // Would be populated with ctx
	}, nil
}

// Params returns the module parameters
func (q Querier) Params(req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	return &types.QueryParamsResponse{
		Params: types.DefaultParams(),
	}, nil
}

// GRPCQuerier implements the gRPC query interface with proper context handling
type GRPCQuerier struct {
	Keeper
}

// AccountRoles returns all roles assigned to an account
func (q GRPCQuerier) AccountRoles(c context.Context, req *types.QueryAccountRolesRequest) (*types.QueryAccountRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	roles := q.Keeper.GetAccountRoles(ctx, addr)

	return &types.QueryAccountRolesResponse{
		Address: req.Address,
		Roles:   roles,
	}, nil
}

// RoleMembers returns all accounts with a specific role
func (q GRPCQuerier) RoleMembers(c context.Context, req *types.QueryRoleMembersRequest) (*types.QueryRoleMembersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	role, err := types.RoleFromString(req.Role)
	if err != nil {
		return nil, types.ErrInvalidRole.Wrap(err.Error())
	}

	members := q.Keeper.GetRoleMembers(ctx, role)

	return &types.QueryRoleMembersResponse{
		Role:    role.String(),
		Members: members,
	}, nil
}

// AccountState returns the state of an account
func (q GRPCQuerier) AccountState(c context.Context, req *types.QueryAccountStateRequest) (*types.QueryAccountStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	state, found := q.Keeper.GetAccountState(ctx, addr)
	if !found {
		// Return default active state for accounts without explicit state
		state = types.DefaultAccountStateRecord(req.Address)
	}

	return &types.QueryAccountStateResponse{
		AccountState: state,
	}, nil
}

// GenesisAccounts returns all genesis accounts
func (q GRPCQuerier) GenesisAccounts(c context.Context, req *types.QueryGenesisAccountsRequest) (*types.QueryGenesisAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	accounts := q.Keeper.GetGenesisAccounts(ctx)
	addresses := make([]string, len(accounts))
	for i, acc := range accounts {
		addresses[i] = acc.String()
	}

	return &types.QueryGenesisAccountsResponse{
		Addresses: addresses,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
