package types

import (
	"context"
	"fmt"
	"net/http"

	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	proto "github.com/golang/protobuf/proto" //nolint:staticcheck // generated gateway compatibility
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	filterRolesQuery = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}

	patternRolesAccountRoles = mustRolesPattern("/virtengine/roles/v1/account/{address}/roles")
	patternRolesRoleMembers  = mustRolesPattern("/virtengine/roles/v1/role/{role}/members")
	patternRolesAccountState = mustRolesPattern("/virtengine/roles/v1/account/{address}/state")
	patternRolesGenesis      = mustRolesPattern("/virtengine/roles/v1/genesis_accounts")
	patternRolesParams       = mustRolesPattern("/virtengine/roles/v1/params")
	patternRolesHasRole      = mustRolesPattern("/virtengine/roles/v1/account/{address}/has_role/{role}")
)

// QueryClient is the generated gRPC client for the roles query service.
type QueryClient = rolesv1.QueryClient

// NewQueryClient returns a new roles query client.
func NewQueryClient(cc gogogrpc.ClientConn) QueryClient {
	return rolesv1.NewQueryClient(cc)
}

type queryServerAdapter struct {
	srv QueryServer
}

var _ rolesv1.QueryServer = (*queryServerAdapter)(nil)

// RegisterQueryServer registers the local query server against the generated protobuf gRPC service.
func RegisterQueryServer(s gogogrpc.Server, srv QueryServer) {
	rolesv1.RegisterQueryServer(s, &queryServerAdapter{srv: srv})
}

// RegisterQueryHandlerServer registers REST gateway routes backed by the local server implementation.
func RegisterQueryHandlerServer(ctx context.Context, mux *runtime.ServeMux, server QueryServer) error {
	adapter := &queryServerAdapter{srv: server}
	registerRolesGatewayRoutes(ctx, mux, func(ctx context.Context, req *rolesv1.QueryAccountRolesRequest) (proto.Message, error) {
		return adapter.AccountRoles(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryRoleMembersRequest) (proto.Message, error) {
		return adapter.RoleMembers(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryAccountStateRequest) (proto.Message, error) {
		return adapter.AccountState(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryGenesisAccountsRequest) (proto.Message, error) {
		return adapter.GenesisAccounts(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryParamsRequest) (proto.Message, error) {
		return adapter.Params(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryHasRoleRequest) (proto.Message, error) {
		return adapter.HasRole(ctx, req)
	})
	return nil
}

// RegisterQueryHandlerClient registers REST gateway routes backed by a generated gRPC client.
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	registerRolesGatewayRoutes(ctx, mux, func(ctx context.Context, req *rolesv1.QueryAccountRolesRequest) (proto.Message, error) {
		return client.AccountRoles(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryRoleMembersRequest) (proto.Message, error) {
		return client.RoleMembers(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryAccountStateRequest) (proto.Message, error) {
		return client.AccountState(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryGenesisAccountsRequest) (proto.Message, error) {
		return client.GenesisAccounts(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryParamsRequest) (proto.Message, error) {
		return client.Params(ctx, req)
	}, func(ctx context.Context, req *rolesv1.QueryHasRoleRequest) (proto.Message, error) {
		return client.HasRole(ctx, req)
	})
	return nil
}

func (a *queryServerAdapter) AccountRoles(ctx context.Context, req *rolesv1.QueryAccountRolesRequest) (*rolesv1.QueryAccountRolesResponse, error) {
	resp, err := a.srv.AccountRoles(ctx, &QueryAccountRolesRequest{Address: req.Address})
	if err != nil {
		return nil, err
	}

	return &rolesv1.QueryAccountRolesResponse{
		Address: resp.Address,
		Roles:   convertRoleAssignmentsToProto(resp.Roles),
	}, nil
}

func (a *queryServerAdapter) RoleMembers(ctx context.Context, req *rolesv1.QueryRoleMembersRequest) (*rolesv1.QueryRoleMembersResponse, error) {
	resp, err := a.srv.RoleMembers(ctx, &QueryRoleMembersRequest{
		Role:       req.Role,
		Pagination: convertPageRequestFromProto(req.Pagination),
	})
	if err != nil {
		return nil, err
	}

	return &rolesv1.QueryRoleMembersResponse{
		Role:       resp.Role,
		Members:    convertRoleAssignmentsToProto(resp.Members),
		Pagination: convertPageResponseToProto(resp.Pagination),
	}, nil
}

func (a *queryServerAdapter) AccountState(ctx context.Context, req *rolesv1.QueryAccountStateRequest) (*rolesv1.QueryAccountStateResponse, error) {
	resp, err := a.srv.AccountState(ctx, &QueryAccountStateRequest{Address: req.Address})
	if err != nil {
		return nil, err
	}

	return &rolesv1.QueryAccountStateResponse{
		AccountState: convertAccountStateRecordToProto(resp.AccountState),
		Found:        resp.Found,
	}, nil
}

func (a *queryServerAdapter) GenesisAccounts(ctx context.Context, req *rolesv1.QueryGenesisAccountsRequest) (*rolesv1.QueryGenesisAccountsResponse, error) {
	resp, err := a.srv.GenesisAccounts(ctx, &QueryGenesisAccountsRequest{
		Pagination: convertPageRequestFromProto(req.Pagination),
	})
	if err != nil {
		return nil, err
	}

	return &rolesv1.QueryGenesisAccountsResponse{
		Addresses:  append([]string(nil), resp.Addresses...),
		Pagination: convertPageResponseToProto(resp.Pagination),
	}, nil
}

func (a *queryServerAdapter) Params(ctx context.Context, req *rolesv1.QueryParamsRequest) (*rolesv1.QueryParamsResponse, error) {
	resp, err := a.srv.Params(ctx, &QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	return &rolesv1.QueryParamsResponse{
		Params: convertParamsToProto(resp.Params),
	}, nil
}

func (a *queryServerAdapter) HasRole(ctx context.Context, req *rolesv1.QueryHasRoleRequest) (*rolesv1.QueryHasRoleResponse, error) {
	resp, err := a.srv.HasRole(ctx, &QueryHasRoleRequest{
		Address: req.Address,
		Role:    req.Role,
	})
	if err != nil {
		return nil, err
	}

	protoResp := &rolesv1.QueryHasRoleResponse{HasRole: resp.HasRole}
	if resp.Assignment != nil {
		assignment := convertRoleAssignmentToProto(*resp.Assignment)
		protoResp.Assignment = &assignment
	}
	return protoResp, nil
}

func registerRolesGatewayRoutes(
	ctx context.Context,
	mux *runtime.ServeMux,
	accountRoles func(context.Context, *rolesv1.QueryAccountRolesRequest) (proto.Message, error),
	roleMembers func(context.Context, *rolesv1.QueryRoleMembersRequest) (proto.Message, error),
	accountState func(context.Context, *rolesv1.QueryAccountStateRequest) (proto.Message, error),
	genesisAccounts func(context.Context, *rolesv1.QueryGenesisAccountsRequest) (proto.Message, error),
	params func(context.Context, *rolesv1.QueryParamsRequest) (proto.Message, error),
	hasRole func(context.Context, *rolesv1.QueryHasRoleRequest) (proto.Message, error),
) {
	mux.Handle("GET", patternRolesAccountRoles, rolesGatewayHandler(ctx, mux, requestRolesAccountRoles, accountRoles))

	mux.Handle("GET", patternRolesRoleMembers, rolesGatewayHandler(ctx, mux, requestRolesRoleMembers, roleMembers))

	mux.Handle("GET", patternRolesAccountState, rolesGatewayHandler(ctx, mux, requestRolesAccountState, accountState))

	mux.Handle("GET", patternRolesGenesis, rolesGatewayHandler(ctx, mux, requestRolesGenesisAccounts, genesisAccounts))

	mux.Handle("GET", patternRolesParams, rolesGatewayHandler(ctx, mux, requestRolesParams, params))

	mux.Handle("GET", patternRolesHasRole, rolesGatewayHandler(ctx, mux, requestRolesHasRole, hasRole))
}

func rolesGatewayHandler[T proto.Message](
	ctx context.Context,
	mux *runtime.ServeMux,
	request func(*http.Request, map[string]string) (T, error),
	invoke func(context.Context, T) (proto.Message, error),
) runtime.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		marshaler, outbound := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outbound, w, req, err)
			return
		}

		protoReq, err := request(req, pathParams)
		if err != nil {
			runtime.HTTPError(rctx, mux, outbound, w, req, err)
			return
		}

		msg, err := invoke(rctx, protoReq)
		if err != nil {
			runtime.HTTPError(rctx, mux, outbound, w, req, err)
			return
		}

		runtime.ForwardResponseMessage(rctx, mux, marshaler, w, req, msg)
	}
}

func requestRolesAccountRoles(_ *http.Request, pathParams map[string]string) (*rolesv1.QueryAccountRolesRequest, error) {
	address, ok := pathParams["address"]
	if !ok || address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "address")
	}
	return &rolesv1.QueryAccountRolesRequest{Address: address}, nil
}

func requestRolesRoleMembers(req *http.Request, pathParams map[string]string) (*rolesv1.QueryRoleMembersRequest, error) {
	role, ok := pathParams["role"]
	if !ok || role == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "role")
	}
	if err := req.ParseForm(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	protoReq := &rolesv1.QueryRoleMembersRequest{Role: role}
	if err := runtime.PopulateQueryParameters(protoReq, req.Form, filterRolesQuery); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return protoReq, nil
}

func requestRolesAccountState(_ *http.Request, pathParams map[string]string) (*rolesv1.QueryAccountStateRequest, error) {
	address, ok := pathParams["address"]
	if !ok || address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "address")
	}
	return &rolesv1.QueryAccountStateRequest{Address: address}, nil
}

func requestRolesGenesisAccounts(req *http.Request, _ map[string]string) (*rolesv1.QueryGenesisAccountsRequest, error) {
	if err := req.ParseForm(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	protoReq := &rolesv1.QueryGenesisAccountsRequest{}
	if err := runtime.PopulateQueryParameters(protoReq, req.Form, filterRolesQuery); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return protoReq, nil
}

func requestRolesParams(_ *http.Request, _ map[string]string) (*rolesv1.QueryParamsRequest, error) {
	return &rolesv1.QueryParamsRequest{}, nil
}

func requestRolesHasRole(_ *http.Request, pathParams map[string]string) (*rolesv1.QueryHasRoleRequest, error) {
	address, ok := pathParams["address"]
	if !ok || address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "address")
	}
	role, ok := pathParams["role"]
	if !ok || role == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing parameter %s", "role")
	}
	return &rolesv1.QueryHasRoleRequest{
		Address: address,
		Role:    role,
	}, nil
}

func convertPageRequestFromProto(page *sdkquery.PageRequest) *PageRequest {
	if page == nil {
		return nil
	}
	return &PageRequest{
		Key:        append([]byte(nil), page.Key...),
		Offset:     page.Offset,
		Limit:      page.Limit,
		CountTotal: page.CountTotal,
		Reverse:    page.Reverse,
	}
}

func convertPageResponseToProto(page *PageResponse) *sdkquery.PageResponse {
	if page == nil {
		return nil
	}
	return &sdkquery.PageResponse{
		NextKey: append([]byte(nil), page.NextKey...),
		Total:   page.Total,
	}
}

func convertRoleAssignmentsToProto(assignments []RoleAssignment) []rolesv1.RoleAssignment {
	if len(assignments) == 0 {
		return nil
	}
	protoAssignments := make([]rolesv1.RoleAssignment, len(assignments))
	for i, assignment := range assignments {
		protoAssignments[i] = convertRoleAssignmentToProto(assignment)
	}
	return protoAssignments
}

func convertRoleAssignmentToProto(assignment RoleAssignment) rolesv1.RoleAssignment {
	return rolesv1.RoleAssignment{
		Address:    assignment.Address,
		Role:       rolesv1.Role(assignment.Role),
		AssignedBy: assignment.AssignedBy,
		AssignedAt: assignment.AssignedAt,
	}
}

func convertAccountStateRecordToProto(record AccountStateRecord) rolesv1.AccountStateRecord {
	return rolesv1.AccountStateRecord{
		Address:       record.Address,
		State:         rolesv1.AccountState(record.State),
		Reason:        record.Reason,
		ModifiedBy:    record.ModifiedBy,
		ModifiedAt:    record.ModifiedAt,
		PreviousState: rolesv1.AccountState(record.PreviousState),
	}
}

func convertParamsToProto(params Params) rolesv1.Params {
	return rolesv1.Params{
		MaxRolesPerAccount: params.MaxRolesPerAccount,
		AllowSelfRevoke:    params.AllowSelfRevoke,
	}
}

func mustRolesPattern(tmpl string) runtime.Pattern {
	compiled, err := httprule.Parse(tmpl)
	if err != nil {
		panic(fmt.Sprintf("invalid roles gateway template %q: %v", tmpl, err))
	}
	template := compiled.Compile()
	return runtime.MustPattern(runtime.NewPattern(
		template.Version,
		template.OpCodes,
		template.Pool,
		template.Verb,
		runtime.AssumeColonVerbOpt(false),
	))
}
