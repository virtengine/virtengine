// Package types contains types for the HPC module.
//
// VE-5F: Query types for workload templates
package types

import (
	"fmt"

	query "github.com/cosmos/cosmos-sdk/types/query"
)

// Query request and response types

// QueryWorkloadTemplateRequest queries a specific workload template
type QueryWorkloadTemplateRequest struct {
	TemplateId string `json:"template_id"`
	Version    string `json:"version"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplateRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplateRequest) Reset() { *m = QueryWorkloadTemplateRequest{} }

// String implements proto.Message
func (m *QueryWorkloadTemplateRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateRequest{TemplateId: %s, Version: %s}", m.TemplateId, m.Version)
}

// QueryWorkloadTemplateResponse is the response for workload template query
type QueryWorkloadTemplateResponse struct {
	Template *WorkloadTemplate `json:"template"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplateResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplateResponse) Reset() { *m = QueryWorkloadTemplateResponse{} }

// String implements proto.Message
func (m *QueryWorkloadTemplateResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateResponse{Template: %v}", m.Template)
}

// QueryWorkloadTemplatesRequest queries all workload templates
type QueryWorkloadTemplatesRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
	TemplateId string             `json:"template_id,omitempty"` // Optional: filter by template ID
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesRequest) Reset() { *m = QueryWorkloadTemplatesRequest{} }

// String implements proto.Message
func (m *QueryWorkloadTemplatesRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesRequest{TemplateId: %s}", m.TemplateId)
}

// QueryWorkloadTemplatesResponse is the response for workload templates query
type QueryWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `json:"templates"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesResponse) Reset() { *m = QueryWorkloadTemplatesResponse{} }

// String implements proto.Message
func (m *QueryWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplatesByTypeRequest queries workload templates by type
type QueryWorkloadTemplatesByTypeRequest struct {
	Type       WorkloadType       `json:"type"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesByTypeRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesByTypeRequest) Reset() { *m = QueryWorkloadTemplatesByTypeRequest{} }

// String implements proto.Message
func (m *QueryWorkloadTemplatesByTypeRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByTypeRequest{Type: %s}", m.Type)
}

// QueryWorkloadTemplatesByTypeResponse is the response for workload templates by type query
type QueryWorkloadTemplatesByTypeResponse struct {
	Templates  []*WorkloadTemplate `json:"templates"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesByTypeResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesByTypeResponse) Reset() { *m = QueryWorkloadTemplatesByTypeResponse{} }

// String implements proto.Message
func (m *QueryWorkloadTemplatesByTypeResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByTypeResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplatesByPublisherRequest queries workload templates by publisher
type QueryWorkloadTemplatesByPublisherRequest struct {
	Publisher  string             `json:"publisher"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesByPublisherRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesByPublisherRequest) Reset() {
	*m = QueryWorkloadTemplatesByPublisherRequest{}
}

// String implements proto.Message
func (m *QueryWorkloadTemplatesByPublisherRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByPublisherRequest{Publisher: %s}", m.Publisher)
}

// QueryWorkloadTemplatesByPublisherResponse is the response for workload templates by publisher query
type QueryWorkloadTemplatesByPublisherResponse struct {
	Templates  []*WorkloadTemplate `json:"templates"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplatesByPublisherResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplatesByPublisherResponse) Reset() {
	*m = QueryWorkloadTemplatesByPublisherResponse{}
}

// String implements proto.Message
func (m *QueryWorkloadTemplatesByPublisherResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplatesByPublisherResponse{Templates: %d}", len(m.Templates))
}

// QueryApprovedWorkloadTemplatesRequest queries approved workload templates
type QueryApprovedWorkloadTemplatesRequest struct {
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryApprovedWorkloadTemplatesRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryApprovedWorkloadTemplatesRequest) Reset() {
	*m = QueryApprovedWorkloadTemplatesRequest{}
}

// String implements proto.Message
func (m *QueryApprovedWorkloadTemplatesRequest) String() string {
	return "QueryApprovedWorkloadTemplatesRequest{}"
}

// QueryApprovedWorkloadTemplatesResponse is the response for approved workload templates query
type QueryApprovedWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `json:"templates"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QueryApprovedWorkloadTemplatesResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryApprovedWorkloadTemplatesResponse) Reset() {
	*m = QueryApprovedWorkloadTemplatesResponse{}
}

// String implements proto.Message
func (m *QueryApprovedWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QueryApprovedWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// QueryWorkloadTemplateUsageRequest queries workload template usage statistics
type QueryWorkloadTemplateUsageRequest struct {
	TemplateId string `json:"template_id"`
	Version    string `json:"version"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplateUsageRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplateUsageRequest) Reset() { *m = QueryWorkloadTemplateUsageRequest{} }

// String implements proto.Message
func (m *QueryWorkloadTemplateUsageRequest) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateUsageRequest{TemplateId: %s, Version: %s}", m.TemplateId, m.Version)
}

// QueryWorkloadTemplateUsageResponse is the response for workload template usage query
type QueryWorkloadTemplateUsageResponse struct {
	TemplateId    string `json:"template_id"`
	Version       string `json:"version"`
	TotalUses     int64  `json:"total_uses"`
	ActiveJobs    int64  `json:"active_jobs"`
	CompletedJobs int64  `json:"completed_jobs"`
	FailedJobs    int64  `json:"failed_jobs"`
}

// ProtoMessage implements proto.Message
func (*QueryWorkloadTemplateUsageResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QueryWorkloadTemplateUsageResponse) Reset() { *m = QueryWorkloadTemplateUsageResponse{} }

// String implements proto.Message
func (m *QueryWorkloadTemplateUsageResponse) String() string {
	return fmt.Sprintf("QueryWorkloadTemplateUsageResponse{TemplateId: %s, Version: %s, TotalUses: %d}", m.TemplateId, m.Version, m.TotalUses)
}

// QuerySearchWorkloadTemplatesRequest searches workload templates by query string
type QuerySearchWorkloadTemplatesRequest struct {
	Query      string             `json:"query"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QuerySearchWorkloadTemplatesRequest) ProtoMessage() {}

// Reset implements proto.Message
func (m *QuerySearchWorkloadTemplatesRequest) Reset() { *m = QuerySearchWorkloadTemplatesRequest{} }

// String implements proto.Message
func (m *QuerySearchWorkloadTemplatesRequest) String() string {
	return fmt.Sprintf("QuerySearchWorkloadTemplatesRequest{Query: %s}", m.Query)
}

// QuerySearchWorkloadTemplatesResponse is the response for workload template search query
type QuerySearchWorkloadTemplatesResponse struct {
	Templates  []*WorkloadTemplate `json:"templates"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// ProtoMessage implements proto.Message
func (*QuerySearchWorkloadTemplatesResponse) ProtoMessage() {}

// Reset implements proto.Message
func (m *QuerySearchWorkloadTemplatesResponse) Reset() { *m = QuerySearchWorkloadTemplatesResponse{} }

// String implements proto.Message
func (m *QuerySearchWorkloadTemplatesResponse) String() string {
	return fmt.Sprintf("QuerySearchWorkloadTemplatesResponse{Templates: %d}", len(m.Templates))
}

// Query interface methods

// QueryClient is the interface for workload template queries
type QueryClient interface {
	// WorkloadTemplate queries a specific workload template
	WorkloadTemplate(ctx interface{}, req *QueryWorkloadTemplateRequest) (*QueryWorkloadTemplateResponse, error)

	// WorkloadTemplates queries all workload templates
	WorkloadTemplates(ctx interface{}, req *QueryWorkloadTemplatesRequest) (*QueryWorkloadTemplatesResponse, error)

	// WorkloadTemplatesByType queries workload templates by type
	WorkloadTemplatesByType(ctx interface{}, req *QueryWorkloadTemplatesByTypeRequest) (*QueryWorkloadTemplatesByTypeResponse, error)

	// WorkloadTemplatesByPublisher queries workload templates by publisher
	WorkloadTemplatesByPublisher(ctx interface{}, req *QueryWorkloadTemplatesByPublisherRequest) (*QueryWorkloadTemplatesByPublisherResponse, error)

	// ApprovedWorkloadTemplates queries approved workload templates
	ApprovedWorkloadTemplates(ctx interface{}, req *QueryApprovedWorkloadTemplatesRequest) (*QueryApprovedWorkloadTemplatesResponse, error)

	// WorkloadTemplateUsage queries workload template usage statistics
	WorkloadTemplateUsage(ctx interface{}, req *QueryWorkloadTemplateUsageRequest) (*QueryWorkloadTemplateUsageResponse, error)

	// SearchWorkloadTemplates searches workload templates by tags
	SearchWorkloadTemplates(ctx interface{}, req *QuerySearchWorkloadTemplatesRequest) (*QuerySearchWorkloadTemplatesResponse, error)
}

// NewQueryClient creates a new QueryClient - placeholder for generated code
func NewQueryClient(clientCtx interface{}) QueryClient {
	// This would normally be generated from protobuf
	// For now, return nil as this is a placeholder
	return nil
}
