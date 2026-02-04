// Package waldur provides Waldur usage reporting API methods
//
// VE-21D: Usage reporting integration for automated Waldur marketplace
package waldur

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UsageClient provides usage reporting operations for Waldur resources
type UsageClient struct {
	marketplace *MarketplaceClient
}

// NewUsageClient creates a new usage client
func NewUsageClient(m *MarketplaceClient) *UsageClient {
	return &UsageClient{marketplace: m}
}

// ComponentUsage represents usage for a single pricing component
type ComponentUsage struct {
	// Type is the component type (matches offering component)
	Type string `json:"type"`

	// Amount is the usage amount in the measured unit
	Amount float64 `json:"amount"`

	// Description provides additional usage context
	Description string `json:"description,omitempty"`
}

// ResourceUsageReport represents a usage report for a Waldur resource
type ResourceUsageReport struct {
	// ResourceUUID is the Waldur resource UUID
	ResourceUUID string `json:"resource_uuid"`

	// PeriodStart is the start of the usage period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the usage period
	PeriodEnd time.Time `json:"period_end"`

	// Components contains usage data for each pricing component
	Components []ComponentUsage `json:"components"`

	// BackendID is the VirtEngine allocation ID for cross-reference
	BackendID string `json:"backend_id,omitempty"`

	// Metadata contains additional usage metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// SubmittedAt is when the report was submitted
	SubmittedAt time.Time `json:"submitted_at,omitempty"`
}

// UsageReportResponse contains the response from submitting a usage report
type UsageReportResponse struct {
	// UUID is the Waldur usage record UUID
	UUID string `json:"uuid,omitempty"`

	// State is the usage record state
	State string `json:"state,omitempty"`

	// ResourceUUID is the resource the usage was recorded for
	ResourceUUID string `json:"resource_uuid,omitempty"`

	// BillingPeriod is the billing period for this usage
	BillingPeriod string `json:"billing_period,omitempty"`

	// TotalAmount is the calculated total amount
	TotalAmount string `json:"total_amount,omitempty"`

	// Currency is the currency for the total amount
	Currency string `json:"currency,omitempty"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created,omitempty"`
}

// SubmitUsageReport submits a usage report for a marketplace resource
func (u *UsageClient) SubmitUsageReport(ctx context.Context, report *ResourceUsageReport) (*UsageReportResponse, error) {
	if report == nil {
		return nil, fmt.Errorf("usage report is required")
	}
	if report.ResourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}
	if len(report.Components) == 0 {
		return nil, fmt.Errorf("at least one component usage is required")
	}

	var response *UsageReportResponse

	err := u.marketplace.client.doWithRetry(ctx, func() error {
		// Build request body
		body := map[string]interface{}{
			"plan_period_start": report.PeriodStart.Format(time.RFC3339),
			"plan_period_end":   report.PeriodEnd.Format(time.RFC3339),
		}

		// Build component usages
		usages := make(map[string]float64)
		for _, component := range report.Components {
			usages[component.Type] = component.Amount
		}
		body["usages"] = usages

		if report.BackendID != "" {
			body["backend_id"] = report.BackendID
		}

		if len(report.Metadata) > 0 {
			body["metadata"] = report.Metadata
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/marketplace-resources/%s/submit_usage/", report.ResourceUUID)
		respBody, statusCode, err := u.marketplace.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK && statusCode != http.StatusCreated && statusCode != http.StatusAccepted {
			return mapHTTPError(statusCode, respBody)
		}

		// Parse response
		response = &UsageReportResponse{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, response); err != nil {
				// Some endpoints return minimal response
				response.State = "submitted"
				response.ResourceUUID = report.ResourceUUID
			}
		} else {
			response.State = "submitted"
			response.ResourceUUID = report.ResourceUUID
		}

		return nil
	})

	return response, err
}

// SubmitBulkUsage submits usage reports for multiple resources
func (u *UsageClient) SubmitBulkUsage(ctx context.Context, reports []*ResourceUsageReport) ([]*UsageReportResponse, error) {
	if len(reports) == 0 {
		return nil, nil
	}

	responses := make([]*UsageReportResponse, 0, len(reports))

	for _, report := range reports {
		resp, err := u.SubmitUsageReport(ctx, report)
		if err != nil {
			// Continue with other reports, log error
			responses = append(responses, &UsageReportResponse{
				ResourceUUID: report.ResourceUUID,
				State:        "failed",
			})
			continue
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// GetResourceUsage retrieves usage records for a resource
func (u *UsageClient) GetResourceUsage(ctx context.Context, resourceUUID string, periodStart, periodEnd *time.Time) ([]UsageRecord, error) {
	if resourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}

	var records []UsageRecord

	err := u.marketplace.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/marketplace-resources/%s/usages/", resourceUUID)

		// Add date filters if provided
		queryParams := []string{}
		if periodStart != nil {
			queryParams = append(queryParams, fmt.Sprintf("date_start=%s", periodStart.Format("2006-01-02")))
		}
		if periodEnd != nil {
			queryParams = append(queryParams, fmt.Sprintf("date_end=%s", periodEnd.Format("2006-01-02")))
		}

		if len(queryParams) > 0 {
			path += "?" + joinQueryParams(queryParams)
		}

		respBody, statusCode, err := u.marketplace.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &records); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return records, err
}

// UsageRecord represents a recorded usage in Waldur
type UsageRecord struct {
	// UUID is the usage record UUID
	UUID string `json:"uuid"`

	// ResourceUUID is the resource UUID
	ResourceUUID string `json:"resource_uuid"`

	// ComponentType is the pricing component type
	ComponentType string `json:"component_type"`

	// Usage is the usage amount
	Usage float64 `json:"usage"`

	// Date is the usage date
	Date time.Time `json:"date"`

	// BillingPeriod is the billing period
	BillingPeriod string `json:"billing_period"`

	// Created is when the record was created
	Created time.Time `json:"created"`
}

// CreateUsageRecordRequest contains parameters for creating a usage record
type CreateUsageRecordRequest struct {
	// ResourceUUID is the resource UUID
	ResourceUUID string `json:"resource_uuid"`

	// ComponentType is the pricing component type
	ComponentType string `json:"component_type"`

	// Usage is the usage amount
	Usage float64 `json:"usage"`

	// Date is the usage date
	Date time.Time `json:"date"`

	// Description is an optional description
	Description string `json:"description,omitempty"`
}

// CreateUsageRecord creates a single usage record
func (u *UsageClient) CreateUsageRecord(ctx context.Context, req CreateUsageRecordRequest) (*UsageRecord, error) {
	if req.ResourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}
	if req.ComponentType == "" {
		return nil, fmt.Errorf("component type is required")
	}

	var record *UsageRecord

	err := u.marketplace.client.doWithRetry(ctx, func() error {
		body := map[string]interface{}{
			"type":  req.ComponentType,
			"usage": req.Usage,
			"date":  req.Date.Format("2006-01-02"),
		}

		if req.Description != "" {
			body["description"] = req.Description
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/marketplace-resources/%s/usages/", req.ResourceUUID)
		respBody, statusCode, err := u.marketplace.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		record = &UsageRecord{}
		if err := json.Unmarshal(respBody, record); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		return nil
	})

	return record, err
}

// UsageComponent represents a usage component in Waldur format
type UsageComponent struct {
	// Type is the component type
	Type string `json:"type"`

	// Name is the component name
	Name string `json:"name"`

	// Usage is the current usage
	Usage float64 `json:"usage"`

	// Limit is the usage limit (if any)
	Limit float64 `json:"limit,omitempty"`

	// MeasuredUnit is the unit of measurement
	MeasuredUnit string `json:"measured_unit,omitempty"`
}

// GetResourceComponentUsage gets current component usage for a resource
func (u *UsageClient) GetResourceComponentUsage(ctx context.Context, resourceUUID string) ([]UsageComponent, error) {
	if resourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}

	var components []UsageComponent

	err := u.marketplace.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/marketplace-resources/%s/", resourceUUID)
		respBody, statusCode, err := u.marketplace.client.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		// Parse resource to extract current_usages
		var resource struct {
			CurrentUsages []UsageComponent `json:"current_usages"`
		}
		if err := json.Unmarshal(respBody, &resource); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		components = resource.CurrentUsages
		return nil
	})

	return components, err
}

// UpdateResourceLimits updates the limits for a resource
func (u *UsageClient) UpdateResourceLimits(ctx context.Context, resourceUUID string, limits map[string]int) error {
	if resourceUUID == "" {
		return fmt.Errorf("resource UUID is required")
	}
	if len(limits) == 0 {
		return nil
	}

	return u.marketplace.client.doWithRetry(ctx, func() error {
		body := map[string]interface{}{
			"limits": limits,
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/marketplace-resources/%s/update_limits/", resourceUUID)
		respBody, statusCode, err := u.marketplace.client.doRequest(ctx, http.MethodPost, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK && statusCode != http.StatusAccepted {
			return mapHTTPError(statusCode, respBody)
		}

		return nil
	})
}

// SubmitComponentUsage is a convenience method to submit usage for specific components
func (u *UsageClient) SubmitComponentUsage(
	ctx context.Context,
	resourceUUID string,
	periodStart, periodEnd time.Time,
	cpuHours, gpuHours, ramGBHours, storageGBHours, networkGB float64,
) (*UsageReportResponse, error) {
	components := make([]ComponentUsage, 0)

	if cpuHours > 0 {
		components = append(components, ComponentUsage{
			Type:        "cpu_hours",
			Amount:      cpuHours,
			Description: "CPU usage in core-hours",
		})
	}
	if gpuHours > 0 {
		components = append(components, ComponentUsage{
			Type:        "gpu_hours",
			Amount:      gpuHours,
			Description: "GPU usage in GPU-hours",
		})
	}
	if ramGBHours > 0 {
		components = append(components, ComponentUsage{
			Type:        "ram_gb_hours",
			Amount:      ramGBHours,
			Description: "Memory usage in GB-hours",
		})
	}
	if storageGBHours > 0 {
		components = append(components, ComponentUsage{
			Type:        "storage_gb_hours",
			Amount:      storageGBHours,
			Description: "Storage usage in GB-hours",
		})
	}
	if networkGB > 0 {
		components = append(components, ComponentUsage{
			Type:        "network_gb",
			Amount:      networkGB,
			Description: "Network transfer in GB",
		})
	}

	if len(components) == 0 {
		return nil, fmt.Errorf("no usage data provided")
	}

	return u.SubmitUsageReport(ctx, &ResourceUsageReport{
		ResourceUUID: resourceUUID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Components:   components,
	})
}
