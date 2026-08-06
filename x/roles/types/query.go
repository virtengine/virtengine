package types

import (
	"context"
)

// QueryAccountRolesRequest is the request for QueryAccountRoles
type QueryAccountRolesRequest struct {
	Address string `json:"address"`
}

// QueryAccountRolesResponse is the response for QueryAccountRoles
type QueryAccountRolesResponse struct {
	Address string           `json:"address"`
	Roles   []RoleAssignment `json:"roles"`
}

// QueryRoleMembersRequest is the request for QueryRoleMembers
type QueryRoleMembersRequest struct {
	Role       string       `json:"role"`
	Pagination *PageRequest `json:"pagination,omitempty"`
}

// QueryRoleMembersResponse is the response for QueryRoleMembers
type QueryRoleMembersResponse struct {
	Role       string           `json:"role"`
	Members    []RoleAssignment `json:"members"`
	Pagination *PageResponse    `json:"pagination,omitempty"`
}

// QueryAccountStateRequest is the request for QueryAccountState
type QueryAccountStateRequest struct {
	Address string `json:"address"`
}

// QueryAccountStateResponse is the response for QueryAccountState
type QueryAccountStateResponse struct {
	AccountState AccountStateRecord `json:"account_state"`
	Found        bool               `json:"found"`
}

// QueryGenesisAccountsRequest is the request for QueryGenesisAccounts
type QueryGenesisAccountsRequest struct {
	Pagination *PageRequest `json:"pagination,omitempty"`
}

// QueryGenesisAccountsResponse is the response for QueryGenesisAccounts
type QueryGenesisAccountsResponse struct {
	Addresses  []string      `json:"addresses"`
	Pagination *PageResponse `json:"pagination,omitempty"`
}

// QueryParamsRequest is the request for QueryParams
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParams
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryHasRoleRequest is the request for QueryHasRole.
type QueryHasRoleRequest struct {
	Address string `json:"address"`
	Role    string `json:"role"`
}

// QueryHasRoleResponse is the response for QueryHasRole.
type QueryHasRoleResponse struct {
	HasRole    bool            `json:"has_role"`
	Assignment *RoleAssignment `json:"assignment,omitempty"`
}

// PageRequest is a simple pagination request
type PageRequest struct {
	Key        []byte `json:"key,omitempty"`
	Offset     uint64 `json:"offset,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
	CountTotal bool   `json:"count_total,omitempty"`
	Reverse    bool   `json:"reverse,omitempty"`
}

// PageResponse is a simple pagination response
type PageResponse struct {
	NextKey []byte `json:"next_key,omitempty"`
	Total   uint64 `json:"total,omitempty"`
}

// QueryServer is the interface for the query server
type QueryServer interface {
	AccountRoles(ctx context.Context, req *QueryAccountRolesRequest) (*QueryAccountRolesResponse, error)
	RoleMembers(ctx context.Context, req *QueryRoleMembersRequest) (*QueryRoleMembersResponse, error)
	AccountState(ctx context.Context, req *QueryAccountStateRequest) (*QueryAccountStateResponse, error)
	GenesisAccounts(ctx context.Context, req *QueryGenesisAccountsRequest) (*QueryGenesisAccountsResponse, error)
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	HasRole(ctx context.Context, req *QueryHasRoleRequest) (*QueryHasRoleResponse, error)
}
