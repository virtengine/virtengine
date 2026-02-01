// Package waldur provides SLURM operations via Waldur API
//
// VE-2024: SLURM allocation management via Waldur
package waldur

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	client "github.com/waldur/go-client"
)

// SLURMClient provides SLURM operations through Waldur
type SLURMClient struct {
	client *Client
}

// NewSLURMClient creates a new SLURM client
func NewSLURMClient(c *Client) *SLURMClient {
	return &SLURMClient{client: c}
}

// SLURMAllocation represents a SLURM allocation in Waldur
type SLURMAllocation struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	BackendID           string    `json:"backend_id"`
	CPULimit            int       `json:"cpu_limit"`
	GPULimit            int       `json:"gpu_limit"`
	RAMLimit            int       `json:"ram_limit"` // MiB
	CPUUsage            int       `json:"cpu_usage"`
	GPUUsage            int       `json:"gpu_usage"`
	RAMUsage            int       `json:"ram_usage"` // MiB
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	CustomerUUID        string    `json:"customer_uuid"`
	ErrorMessage        string    `json:"error_message"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created"`
}

// SLURMAssociation represents a SLURM user association
type SLURMAssociation struct {
	UUID           string `json:"uuid"`
	Username       string `json:"username"`
	AllocationUUID string `json:"allocation_uuid"`
}

// SLURMJob represents a SLURM job (mapped from Waldur's FirecrestJob)
type SLURMJob struct {
	UUID         string     `json:"uuid"`
	Name         string     `json:"name"`
	BackendID    string     `json:"backend_id"`
	State        string     `json:"state"`
	RuntimeState string     `json:"runtime_state"`
	ProjectUUID  string     `json:"project_uuid"`
	CustomerUUID string     `json:"customer_uuid"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created"`
	ModifiedAt   *time.Time `json:"modified"`
}

// ListSLURMAllocationsParams contains parameters for listing SLURM allocations
type ListSLURMAllocationsParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	Name                string
	IsActive            *bool
	Page                int
	PageSize            int
}

// ListSLURMAllocations lists SLURM allocations via Waldur
func (s *SLURMClient) ListSLURMAllocations(ctx context.Context, params ListSLURMAllocationsParams) ([]SLURMAllocation, error) {
	var allocations []SLURMAllocation

	err := s.client.doWithRetry(ctx, func() error {
		apiParams := &client.SlurmAllocationsListParams{}

		if params.ServiceSettingsUUID != "" {
			u := uuid.MustParse(params.ServiceSettingsUUID)
			apiParams.ServiceSettingsUuid = &u
		}
		if params.ProjectUUID != "" {
			u := uuid.MustParse(params.ProjectUUID)
			apiParams.ProjectUuid = &u
		}
		if params.CustomerUUID != "" {
			u := uuid.MustParse(params.CustomerUUID)
			apiParams.CustomerUuid = &u
		}
		if params.Name != "" {
			apiParams.Name = &params.Name
		}
		if params.IsActive != nil {
			apiParams.IsActive = params.IsActive
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := s.client.api.SlurmAllocationsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		allocations = make([]SLURMAllocation, 0, len(*resp.JSON200))
		for _, a := range *resp.JSON200 {
			allocation := mapSLURMAllocation(&a)
			allocations = append(allocations, allocation)
		}

		return nil
	})

	return allocations, err
}

// GetSLURMAllocation retrieves a specific SLURM allocation
func (s *SLURMClient) GetSLURMAllocation(ctx context.Context, allocationUUID string) (*SLURMAllocation, error) {
	var allocation *SLURMAllocation

	err := s.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(allocationUUID)
		resp, err := s.client.api.SlurmAllocationsRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		a := mapSLURMAllocation(resp.JSON200)
		allocation = &a
		return nil
	})

	return allocation, err
}

// DeleteSLURMAllocation deletes a SLURM allocation
func (s *SLURMClient) DeleteSLURMAllocation(ctx context.Context, allocationUUID string) error {
	return s.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(allocationUUID)
		resp, err := s.client.api.SlurmAllocationsDestroyWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// SetSLURMAllocationLimits sets resource limits on a SLURM allocation
func (s *SLURMClient) SetSLURMAllocationLimits(ctx context.Context, allocationUUID string, cpuLimit, gpuLimit, ramLimit int64) error {
	return s.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(allocationUUID)
		body := client.SlurmAllocationsSetLimitsJSONRequestBody{
			CpuLimit: int(cpuLimit),
			GpuLimit: int(gpuLimit),
			RamLimit: int(ramLimit),
		}
		resp, err := s.client.api.SlurmAllocationsSetLimitsWithResponse(ctx, uuidType, body)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// ListSLURMAssociationsParams contains parameters for listing SLURM associations
type ListSLURMAssociationsParams struct {
	AllocationUUID string
	Page           int
	PageSize       int
}

// ListSLURMAssociations lists SLURM associations via Waldur
func (s *SLURMClient) ListSLURMAssociations(ctx context.Context, params ListSLURMAssociationsParams) ([]SLURMAssociation, error) {
	var associations []SLURMAssociation

	err := s.client.doWithRetry(ctx, func() error {
		apiParams := &client.SlurmAssociationsListParams{}

		if params.AllocationUUID != "" {
			u := uuid.MustParse(params.AllocationUUID)
			apiParams.AllocationUuid = &u
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := s.client.api.SlurmAssociationsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		associations = make([]SLURMAssociation, 0, len(*resp.JSON200))
		for _, a := range *resp.JSON200 {
			assoc := SLURMAssociation{
				Username:       a.Username,
				AllocationUUID: a.Allocation,
			}
			if a.Uuid != nil {
				assoc.UUID = a.Uuid.String()
			}
			associations = append(associations, assoc)
		}

		return nil
	})

	return associations, err
}

// ListSLURMJobsParams contains parameters for listing SLURM jobs
// Note: The Waldur SLURM Jobs API only supports pagination filters
type ListSLURMJobsParams struct {
	Page     int
	PageSize int
}

// ListSLURMJobs lists SLURM jobs via Waldur
func (s *SLURMClient) ListSLURMJobs(ctx context.Context, params ListSLURMJobsParams) ([]SLURMJob, error) {
	var jobs []SLURMJob

	err := s.client.doWithRetry(ctx, func() error {
		apiParams := &client.SlurmJobsListParams{}

		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := s.client.api.SlurmJobsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		jobs = make([]SLURMJob, 0, len(*resp.JSON200))
		for _, j := range *resp.JSON200 {
			job := mapSLURMJob(&j)
			jobs = append(jobs, job)
		}

		return nil
	})

	return jobs, err
}

// GetSLURMJob retrieves a specific SLURM job
func (s *SLURMClient) GetSLURMJob(ctx context.Context, jobUUID string) (*SLURMJob, error) {
	var job *SLURMJob

	err := s.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(jobUUID)
		resp, err := s.client.api.SlurmJobsRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		j := mapSLURMJob(resp.JSON200)
		job = &j
		return nil
	})

	return job, err
}

// DeleteSLURMJob cancels/deletes a SLURM job
func (s *SLURMClient) DeleteSLURMJob(ctx context.Context, jobUUID string) error {
	return s.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(jobUUID)
		resp, err := s.client.api.SlurmJobsDestroyWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// WaitForSLURMAllocationState waits for a SLURM allocation to reach the desired state
func (s *SLURMClient) WaitForSLURMAllocationState(ctx context.Context, allocationUUID string, desiredState string, pollInterval time.Duration) (*SLURMAllocation, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			allocation, err := s.GetSLURMAllocation(ctx, allocationUUID)
			if err != nil {
				return nil, err
			}

			if allocation.State == desiredState {
				return allocation, nil
			}

			// Check for error states
			if allocation.State == "Erred" {
				return allocation, fmt.Errorf("allocation entered error state: %s", allocation.ErrorMessage)
			}
		}
	}
}

// mapSLURMAllocation maps Waldur SlurmAllocation to our SLURMAllocation type
func mapSLURMAllocation(a *client.SlurmAllocation) SLURMAllocation {
	allocation := SLURMAllocation{
		Name:         safeString(a.Name),
		Description:  safeString(a.Description),
		BackendID:    safeString(a.BackendId),
		ErrorMessage: safeString(a.ErrorMessage),
	}
	if a.Uuid != nil {
		allocation.UUID = a.Uuid.String()
	}
	if a.State != nil {
		allocation.State = string(*a.State)
	}
	if a.CpuLimit != nil {
		allocation.CPULimit = *a.CpuLimit
	}
	if a.GpuLimit != nil {
		allocation.GPULimit = *a.GpuLimit
	}
	if a.RamLimit != nil {
		allocation.RAMLimit = *a.RamLimit
	}
	if a.CpuUsage != nil {
		allocation.CPUUsage = *a.CpuUsage
	}
	if a.GpuUsage != nil {
		allocation.GPUUsage = *a.GpuUsage
	}
	if a.RamUsage != nil {
		allocation.RAMUsage = *a.RamUsage
	}
	if a.ServiceSettingsUuid != nil {
		allocation.ServiceSettingsUUID = a.ServiceSettingsUuid.String()
	}
	if a.ProjectUuid != nil {
		allocation.ProjectUUID = a.ProjectUuid.String()
	}
	if a.CustomerUuid != nil {
		allocation.CustomerUUID = a.CustomerUuid.String()
	}
	if a.IsActive != nil {
		allocation.IsActive = *a.IsActive
	}
	if a.Created != nil {
		allocation.CreatedAt = *a.Created
	}
	return allocation
}

// mapSLURMJob maps Waldur FirecrestJob to our SLURMJob type
func mapSLURMJob(j *client.FirecrestJob) SLURMJob {
	job := SLURMJob{
		Name:         safeString(j.Name),
		BackendID:    safeString(j.BackendId),
		RuntimeState: safeString(j.RuntimeState),
		ErrorMessage: safeString(j.ErrorMessage),
	}
	if j.Uuid != nil {
		job.UUID = j.Uuid.String()
	}
	if j.State != nil {
		job.State = string(*j.State)
	}
	if j.ProjectUuid != nil {
		job.ProjectUUID = j.ProjectUuid.String()
	}
	if j.CustomerUuid != nil {
		job.CustomerUUID = j.CustomerUuid.String()
	}
	if j.Created != nil {
		job.CreatedAt = *j.Created
	}
	if j.Modified != nil {
		job.ModifiedAt = j.Modified
	}
	return job
}

