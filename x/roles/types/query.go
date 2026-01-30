package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
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
	Role       string `json:"role"`
	Pagination *PageRequest `json:"pagination,omitempty"`
}

// QueryRoleMembersResponse is the response for QueryRoleMembers
type QueryRoleMembersResponse struct {
	Role       string           `json:"role"`
	Members    []RoleAssignment `json:"members"`
	Pagination *PageResponse `json:"pagination,omitempty"`
}

// QueryAccountStateRequest is the request for QueryAccountState
type QueryAccountStateRequest struct {
	Address string `json:"address"`
}

// QueryAccountStateResponse is the response for QueryAccountState
type QueryAccountStateResponse struct {
	AccountState AccountStateRecord `json:"account_state"`
}

// QueryGenesisAccountsRequest is the request for QueryGenesisAccounts
type QueryGenesisAccountsRequest struct {
	Pagination *PageRequest `json:"pagination,omitempty"`
}

// QueryGenesisAccountsResponse is the response for QueryGenesisAccounts
type QueryGenesisAccountsResponse struct {
	Addresses  []string `json:"addresses"`
	Pagination *PageResponse `json:"pagination,omitempty"`
}

// QueryParamsRequest is the request for QueryParams
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParams
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// PageRequest is a simple pagination request
type PageRequest struct {
	Key    []byte `json:"key,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
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
}

// _Query_serviceDesc is the grpc.ServiceDesc for Query service.
//
//nolint:unused // Reserved for future gRPC registration
var _Query_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.roles.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "AccountRoles", Handler: nil},
		{MethodName: "RoleMembers", Handler: nil},
		{MethodName: "AccountState", Handler: nil},
		{MethodName: "GenesisAccounts", Handler: nil},
		{MethodName: "Params", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/roles/v1/query.proto",
}

// RegisterQueryServer registers the QueryServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterQueryServer(s grpc.Server, impl QueryServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}
