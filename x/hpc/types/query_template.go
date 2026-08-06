// Package types contains types for the HPC module.
//
// VE-5F: Query types and transport wiring for workload templates.
package types

import (
	"context"
	"fmt"

	query "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const workloadTemplateQueryServiceName = "virtengine.hpc.v1.WorkloadTemplateQuery"

// QueryWorkloadTemplateRequest queries a specific workload template.
type QueryWorkloadTemplateRequest struct {
	TemplateId string `protobuf:"bytes,1,opt,name=template_id,json=templateId,proto3" json:"template_id"`
	Version    string `protobuf:"bytes,2,opt,name=version,proto3" json:"version"`
}

func (*QueryWorkloadTemplateRequest) ProtoMessage() {}
func (m *QueryWorkloadTemplateRequest) Reset()      { *m = QueryWorkloadTemplateRequest{} }
func (m *QueryWorkloadTemplateRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateRequest{TemplateId: %s, Version: %s}", m.TemplateId, m.Version)
}

// QueryWorkloadTemplateResponse is the response for workload template query.
type QueryWorkloadTemplateResponse struct {
	Template *WorkloadTemplate `protobuf:"bytes,1,opt,name=template,proto3" json:"template"`
}

func (*QueryWorkloadTemplateResponse) ProtoMessage() {}
func (m *QueryWorkloadTemplateResponse) Reset()      { *m = QueryWorkloadTemplateResponse{} }
func (m *QueryWorkloadTemplateResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateResponse{Template: %v}", m.Template)
}

// QueryWorkloadTemplatesRequest queries all workload templates.
type QueryWorkloadTemplatesRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
	TemplateId string             `protobuf:"bytes,2,opt,name=template_id,json=templateId,proto3" json:"template_id,omitempty"`
}

func (*QueryWorkloadTemplatesRequest) ProtoMessage() {}
func (m *QueryWorkloadTemplatesRequest) Reset()      { *m = QueryWorkloadTemplatesRequest{} }
func (m *QueryWorkloadTemplatesRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesRequest{TemplateId: %s}", m.TemplateId)
}

// QueryWorkloadTemplatesResponse is the response for workload templates query.
type QueryWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryWorkloadTemplatesResponse) ProtoMessage() {}
func (m *QueryWorkloadTemplatesResponse) Reset()      { *m = QueryWorkloadTemplatesResponse{} }
func (m *QueryWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplatesByTypeRequest queries workload templates by type.
type QueryWorkloadTemplatesByTypeRequest struct {
	Type       WorkloadType       `protobuf:"bytes,1,opt,name=type,proto3" json:"type"`
	Pagination *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryWorkloadTemplatesByTypeRequest) ProtoMessage() {}
func (m *QueryWorkloadTemplatesByTypeRequest) Reset()      { *m = QueryWorkloadTemplatesByTypeRequest{} }
func (m *QueryWorkloadTemplatesByTypeRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByTypeRequest{Type: %s}", m.Type)
}

// QueryWorkloadTemplatesByTypeResponse is the response for workload templates by type query.
type QueryWorkloadTemplatesByTypeResponse struct {
	Templates  []*WorkloadTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryWorkloadTemplatesByTypeResponse) ProtoMessage() {}
func (m *QueryWorkloadTemplatesByTypeResponse) Reset()      { *m = QueryWorkloadTemplatesByTypeResponse{} }
func (m *QueryWorkloadTemplatesByTypeResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByTypeResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplatesByPublisherRequest queries workload templates by publisher.
type QueryWorkloadTemplatesByPublisherRequest struct {
	Publisher  string             `protobuf:"bytes,1,opt,name=publisher,proto3" json:"publisher"`
	Pagination *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryWorkloadTemplatesByPublisherRequest) ProtoMessage() {}
func (m *QueryWorkloadTemplatesByPublisherRequest) Reset() {
	*m = QueryWorkloadTemplatesByPublisherRequest{}
}
func (m *QueryWorkloadTemplatesByPublisherRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByPublisherRequest{Publisher: %s}", m.Publisher)
}

// QueryWorkloadTemplatesByPublisherResponse is the response for workload templates by publisher query.
type QueryWorkloadTemplatesByPublisherResponse struct {
	Templates  []*WorkloadTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryWorkloadTemplatesByPublisherResponse) ProtoMessage() {}
func (m *QueryWorkloadTemplatesByPublisherResponse) Reset() {
	*m = QueryWorkloadTemplatesByPublisherResponse{}
}
func (m *QueryWorkloadTemplatesByPublisherResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByPublisherResponse{Templates: %d}", len(m.Templates))
}

// QueryApprovedWorkloadTemplatesRequest queries approved workload templates.
type QueryApprovedWorkloadTemplatesRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryApprovedWorkloadTemplatesRequest) ProtoMessage() {}
func (m *QueryApprovedWorkloadTemplatesRequest) Reset() {
	*m = QueryApprovedWorkloadTemplatesRequest{}
}
func (m *QueryApprovedWorkloadTemplatesRequest) String() string {
	return "QueryApprovedWorkloadTemplatesRequest{}"
}

// QueryApprovedWorkloadTemplatesResponse is the response for approved workload templates query.
type QueryApprovedWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QueryApprovedWorkloadTemplatesResponse) ProtoMessage() {}
func (m *QueryApprovedWorkloadTemplatesResponse) Reset() {
	*m = QueryApprovedWorkloadTemplatesResponse{}
}
func (m *QueryApprovedWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QueryApprovedWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplateUsageRequest queries workload template usage statistics.
type QueryWorkloadTemplateUsageRequest struct {
	TemplateId string `protobuf:"bytes,1,opt,name=template_id,json=templateId,proto3" json:"template_id"`
	Version    string `protobuf:"bytes,2,opt,name=version,proto3" json:"version"`
}

func (*QueryWorkloadTemplateUsageRequest) ProtoMessage() {}
func (m *QueryWorkloadTemplateUsageRequest) Reset()      { *m = QueryWorkloadTemplateUsageRequest{} }
func (m *QueryWorkloadTemplateUsageRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateUsageRequest{TemplateId: %s, Version: %s}", m.TemplateId, m.Version)
}

// QueryWorkloadTemplateUsageResponse is the response for workload template usage query.
type QueryWorkloadTemplateUsageResponse struct {
	TemplateId    string `protobuf:"bytes,1,opt,name=template_id,json=templateId,proto3" json:"template_id"`
	Version       string `protobuf:"bytes,2,opt,name=version,proto3" json:"version"`
	TotalUses     int64  `protobuf:"varint,3,opt,name=total_uses,json=totalUses,proto3" json:"total_uses"`
	ActiveJobs    int64  `protobuf:"varint,4,opt,name=active_jobs,json=activeJobs,proto3" json:"active_jobs"`
	CompletedJobs int64  `protobuf:"varint,5,opt,name=completed_jobs,json=completedJobs,proto3" json:"completed_jobs"`
	FailedJobs    int64  `protobuf:"varint,6,opt,name=failed_jobs,json=failedJobs,proto3" json:"failed_jobs"`
}

func (*QueryWorkloadTemplateUsageResponse) ProtoMessage() {}
func (m *QueryWorkloadTemplateUsageResponse) Reset()      { *m = QueryWorkloadTemplateUsageResponse{} }
func (m *QueryWorkloadTemplateUsageResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateUsageResponse{TemplateId: %s, Version: %s, TotalUses: %d}", m.TemplateId, m.Version, m.TotalUses)
}

// QuerySearchWorkloadTemplatesRequest searches workload templates by query string.
type QuerySearchWorkloadTemplatesRequest struct {
	Query      string             `protobuf:"bytes,1,opt,name=query,proto3" json:"query"`
	Pagination *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QuerySearchWorkloadTemplatesRequest) ProtoMessage() {}
func (m *QuerySearchWorkloadTemplatesRequest) Reset()      { *m = QuerySearchWorkloadTemplatesRequest{} }
func (m *QuerySearchWorkloadTemplatesRequest) String() string {
	return fmt.Sprintf("QuerySearchWorkloadTemplatesRequest{Query: %s}", m.Query)
}

// QuerySearchWorkloadTemplatesResponse is the response for workload template search query.
type QuerySearchWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `protobuf:"bytes,1,rep,name=templates,proto3" json:"templates"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (*QuerySearchWorkloadTemplatesResponse) ProtoMessage() {}
func (m *QuerySearchWorkloadTemplatesResponse) Reset()      { *m = QuerySearchWorkloadTemplatesResponse{} }
func (m *QuerySearchWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QuerySearchWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// QueryServer is the server API for workload template queries.
type QueryServer interface {
	WorkloadTemplate(context.Context, *QueryWorkloadTemplateRequest) (*QueryWorkloadTemplateResponse, error)
	WorkloadTemplates(context.Context, *QueryWorkloadTemplatesRequest) (*QueryWorkloadTemplatesResponse, error)
	WorkloadTemplatesByType(context.Context, *QueryWorkloadTemplatesByTypeRequest) (*QueryWorkloadTemplatesByTypeResponse, error)
	WorkloadTemplatesByPublisher(context.Context, *QueryWorkloadTemplatesByPublisherRequest) (*QueryWorkloadTemplatesByPublisherResponse, error)
	ApprovedWorkloadTemplates(context.Context, *QueryApprovedWorkloadTemplatesRequest) (*QueryApprovedWorkloadTemplatesResponse, error)
	WorkloadTemplateUsage(context.Context, *QueryWorkloadTemplateUsageRequest) (*QueryWorkloadTemplateUsageResponse, error)
	SearchWorkloadTemplates(context.Context, *QuerySearchWorkloadTemplatesRequest) (*QuerySearchWorkloadTemplatesResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward-compatible implementations.
type UnimplementedQueryServer struct{}

func (*UnimplementedQueryServer) WorkloadTemplate(context.Context, *QueryWorkloadTemplateRequest) (*QueryWorkloadTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkloadTemplate not implemented")
}
func (*UnimplementedQueryServer) WorkloadTemplates(context.Context, *QueryWorkloadTemplatesRequest) (*QueryWorkloadTemplatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkloadTemplates not implemented")
}
func (*UnimplementedQueryServer) WorkloadTemplatesByType(context.Context, *QueryWorkloadTemplatesByTypeRequest) (*QueryWorkloadTemplatesByTypeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkloadTemplatesByType not implemented")
}
func (*UnimplementedQueryServer) WorkloadTemplatesByPublisher(context.Context, *QueryWorkloadTemplatesByPublisherRequest) (*QueryWorkloadTemplatesByPublisherResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkloadTemplatesByPublisher not implemented")
}
func (*UnimplementedQueryServer) ApprovedWorkloadTemplates(context.Context, *QueryApprovedWorkloadTemplatesRequest) (*QueryApprovedWorkloadTemplatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApprovedWorkloadTemplates not implemented")
}
func (*UnimplementedQueryServer) WorkloadTemplateUsage(context.Context, *QueryWorkloadTemplateUsageRequest) (*QueryWorkloadTemplateUsageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WorkloadTemplateUsage not implemented")
}
func (*UnimplementedQueryServer) SearchWorkloadTemplates(context.Context, *QuerySearchWorkloadTemplatesRequest) (*QuerySearchWorkloadTemplatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchWorkloadTemplates not implemented")
}

// QueryClient is the client API for workload template queries.
type QueryClient interface {
	WorkloadTemplate(ctx context.Context, req *QueryWorkloadTemplateRequest) (*QueryWorkloadTemplateResponse, error)
	WorkloadTemplates(ctx context.Context, req *QueryWorkloadTemplatesRequest) (*QueryWorkloadTemplatesResponse, error)
	WorkloadTemplatesByType(ctx context.Context, req *QueryWorkloadTemplatesByTypeRequest) (*QueryWorkloadTemplatesByTypeResponse, error)
	WorkloadTemplatesByPublisher(ctx context.Context, req *QueryWorkloadTemplatesByPublisherRequest) (*QueryWorkloadTemplatesByPublisherResponse, error)
	ApprovedWorkloadTemplates(ctx context.Context, req *QueryApprovedWorkloadTemplatesRequest) (*QueryApprovedWorkloadTemplatesResponse, error)
	WorkloadTemplateUsage(ctx context.Context, req *QueryWorkloadTemplateUsageRequest) (*QueryWorkloadTemplateUsageResponse, error)
	SearchWorkloadTemplates(ctx context.Context, req *QuerySearchWorkloadTemplatesRequest) (*QuerySearchWorkloadTemplatesResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

// NewQueryClient creates a new workload template query client.
func NewQueryClient(clientCtx grpc1.ClientConn) QueryClient {
	return &queryClient{cc: clientCtx}
}

func (c *queryClient) WorkloadTemplate(ctx context.Context, req *QueryWorkloadTemplateRequest) (*QueryWorkloadTemplateResponse, error) {
	resp := new(QueryWorkloadTemplateResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/WorkloadTemplate", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) WorkloadTemplates(ctx context.Context, req *QueryWorkloadTemplatesRequest) (*QueryWorkloadTemplatesResponse, error) {
	resp := new(QueryWorkloadTemplatesResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/WorkloadTemplates", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) WorkloadTemplatesByType(ctx context.Context, req *QueryWorkloadTemplatesByTypeRequest) (*QueryWorkloadTemplatesByTypeResponse, error) {
	resp := new(QueryWorkloadTemplatesByTypeResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/WorkloadTemplatesByType", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) WorkloadTemplatesByPublisher(ctx context.Context, req *QueryWorkloadTemplatesByPublisherRequest) (*QueryWorkloadTemplatesByPublisherResponse, error) {
	resp := new(QueryWorkloadTemplatesByPublisherResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/WorkloadTemplatesByPublisher", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) ApprovedWorkloadTemplates(ctx context.Context, req *QueryApprovedWorkloadTemplatesRequest) (*QueryApprovedWorkloadTemplatesResponse, error) {
	resp := new(QueryApprovedWorkloadTemplatesResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/ApprovedWorkloadTemplates", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) WorkloadTemplateUsage(ctx context.Context, req *QueryWorkloadTemplateUsageRequest) (*QueryWorkloadTemplateUsageResponse, error) {
	resp := new(QueryWorkloadTemplateUsageResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/WorkloadTemplateUsage", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *queryClient) SearchWorkloadTemplates(ctx context.Context, req *QuerySearchWorkloadTemplatesRequest) (*QuerySearchWorkloadTemplatesResponse, error) {
	resp := new(QuerySearchWorkloadTemplatesResponse)
	err := c.cc.Invoke(ctx, "/"+workloadTemplateQueryServiceName+"/SearchWorkloadTemplates", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RegisterQueryServer registers the workload template query service with the gRPC server.
func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

// RegisterQueryHandlerClient registers the gRPC gateway routes for the workload template query service.
// Gateway handlers are not generated in this build, so this remains a no-op.
func RegisterQueryHandlerClient(_ context.Context, _ *runtime.ServeMux, _ QueryClient) error {
	return nil
}

func _Query_WorkloadTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWorkloadTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WorkloadTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/WorkloadTemplate"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WorkloadTemplate(ctx, req.(*QueryWorkloadTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_WorkloadTemplates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWorkloadTemplatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WorkloadTemplates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/WorkloadTemplates"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WorkloadTemplates(ctx, req.(*QueryWorkloadTemplatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_WorkloadTemplatesByType_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWorkloadTemplatesByTypeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WorkloadTemplatesByType(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/WorkloadTemplatesByType"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WorkloadTemplatesByType(ctx, req.(*QueryWorkloadTemplatesByTypeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_WorkloadTemplatesByPublisher_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWorkloadTemplatesByPublisherRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WorkloadTemplatesByPublisher(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/WorkloadTemplatesByPublisher"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WorkloadTemplatesByPublisher(ctx, req.(*QueryWorkloadTemplatesByPublisherRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ApprovedWorkloadTemplates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryApprovedWorkloadTemplatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ApprovedWorkloadTemplates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/ApprovedWorkloadTemplates"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ApprovedWorkloadTemplates(ctx, req.(*QueryApprovedWorkloadTemplatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_WorkloadTemplateUsage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWorkloadTemplateUsageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WorkloadTemplateUsage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/WorkloadTemplateUsage"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WorkloadTemplateUsage(ctx, req.(*QueryWorkloadTemplateUsageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_SearchWorkloadTemplates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySearchWorkloadTemplatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SearchWorkloadTemplates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + workloadTemplateQueryServiceName + "/SearchWorkloadTemplates"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SearchWorkloadTemplates(ctx, req.(*QuerySearchWorkloadTemplatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: workloadTemplateQueryServiceName,
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "WorkloadTemplate", Handler: _Query_WorkloadTemplate_Handler},
		{MethodName: "WorkloadTemplates", Handler: _Query_WorkloadTemplates_Handler},
		{MethodName: "WorkloadTemplatesByType", Handler: _Query_WorkloadTemplatesByType_Handler},
		{MethodName: "WorkloadTemplatesByPublisher", Handler: _Query_WorkloadTemplatesByPublisher_Handler},
		{MethodName: "ApprovedWorkloadTemplates", Handler: _Query_ApprovedWorkloadTemplates_Handler},
		{MethodName: "WorkloadTemplateUsage", Handler: _Query_WorkloadTemplateUsage_Handler},
		{MethodName: "SearchWorkloadTemplates", Handler: _Query_SearchWorkloadTemplates_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "virtengine/hpc/v1/query_template.proto",
}
