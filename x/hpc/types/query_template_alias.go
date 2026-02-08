// Package types contains types for the HPC module.
//
// VE-5F: Query type aliases for workload template queries
package types

// Query type aliases for backwards-compatible naming used in tests and CLI.

// Get workload template (alias)
type QueryGetWorkloadTemplateRequest = QueryWorkloadTemplateRequest
type QueryGetWorkloadTemplateResponse = QueryWorkloadTemplateResponse

// List workload templates (alias)
type QueryListWorkloadTemplatesRequest = QueryWorkloadTemplatesRequest
type QueryListWorkloadTemplatesResponse = QueryWorkloadTemplatesResponse

// List workload templates by type (alias)
type QueryListWorkloadTemplatesByTypeRequest = QueryWorkloadTemplatesByTypeRequest
type QueryListWorkloadTemplatesByTypeResponse = QueryWorkloadTemplatesByTypeResponse

// List workload templates by publisher (alias)
type QueryListWorkloadTemplatesByPublisherRequest = QueryWorkloadTemplatesByPublisherRequest
type QueryListWorkloadTemplatesByPublisherResponse = QueryWorkloadTemplatesByPublisherResponse

// List approved workload templates (alias)
type QueryListApprovedWorkloadTemplatesRequest = QueryApprovedWorkloadTemplatesRequest
type QueryListApprovedWorkloadTemplatesResponse = QueryApprovedWorkloadTemplatesResponse
