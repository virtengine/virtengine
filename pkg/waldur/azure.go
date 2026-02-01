// Package waldur provides Azure operations via Waldur API
//
// VE-2024: Azure VM management via Waldur
package waldur

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	client "github.com/waldur/go-client"
)

// AzureClient provides Azure operations through Waldur
type AzureClient struct {
	client *Client
}

// NewAzureClient creates a new Azure client
func NewAzureClient(c *Client) *AzureClient {
	return &AzureClient{client: c}
}

// AzureVirtualMachine represents an Azure VM in Waldur
type AzureVirtualMachine struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	RuntimeState        string    `json:"runtime_state"`
	BackendID           string    `json:"backend_id"`
	Cores               int       `json:"cores"`
	RAM                 int       `json:"ram"`  // MiB
	Disk                int       `json:"disk"` // MiB
	ImageName           string    `json:"image_name"`
	SizeName            string    `json:"size_name"`
	LocationName        string    `json:"location_name"`
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	ErrorMessage        string    `json:"error_message"`
	CreatedAt           time.Time `json:"created"`
}

// AzureLocation represents an Azure location/region
type AzureLocation struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// AzureSize represents an Azure VM size
type AzureSize struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Cores int    `json:"cores"`
	RAM   int    `json:"ram"`  // MiB
	Disk  int    `json:"disk"` // MiB
}

// ListAzureVMsParams contains parameters for listing Azure VMs
type ListAzureVMsParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	Name                string
	Page                int
	PageSize            int
}

// ListAzureVMs lists Azure virtual machines via Waldur
func (a *AzureClient) ListAzureVMs(ctx context.Context, params ListAzureVMsParams) ([]AzureVirtualMachine, error) {
	var vms []AzureVirtualMachine

	err := a.client.doWithRetry(ctx, func() error {
		apiParams := &client.AzureVirtualmachinesListParams{}

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
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := a.client.api.AzureVirtualmachinesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		vms = make([]AzureVirtualMachine, 0, len(*resp.JSON200))
		for _, v := range *resp.JSON200 {
			vm := mapAzureVirtualMachine(&v)
			vms = append(vms, vm)
		}

		return nil
	})

	return vms, err
}

// GetAzureVM retrieves a specific Azure VM
func (a *AzureClient) GetAzureVM(ctx context.Context, vmUUID string) (*AzureVirtualMachine, error) {
	var vm *AzureVirtualMachine

	err := a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(vmUUID)
		resp, err := a.client.api.AzureVirtualmachinesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		v := mapAzureVirtualMachine(resp.JSON200)
		vm = &v
		return nil
	})

	return vm, err
}

// DeleteAzureVM deletes an Azure VM
func (a *AzureClient) DeleteAzureVM(ctx context.Context, vmUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(vmUUID)
		resp, err := a.client.api.AzureVirtualmachinesDestroyWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StartAzureVM starts a stopped Azure VM
func (a *AzureClient) StartAzureVM(ctx context.Context, vmUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(vmUUID)
		resp, err := a.client.api.AzureVirtualmachinesStartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StopAzureVM stops a running Azure VM
func (a *AzureClient) StopAzureVM(ctx context.Context, vmUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(vmUUID)
		resp, err := a.client.api.AzureVirtualmachinesStopWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// RestartAzureVM restarts an Azure VM
func (a *AzureClient) RestartAzureVM(ctx context.Context, vmUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(vmUUID)
		resp, err := a.client.api.AzureVirtualmachinesRestartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// ListAzureLocations lists available Azure locations
func (a *AzureClient) ListAzureLocations(ctx context.Context, settingsUUID string) ([]AzureLocation, error) {
	var locations []AzureLocation

	err := a.client.doWithRetry(ctx, func() error {
		apiParams := &client.AzureLocationsListParams{}

		if settingsUUID != "" {
			u := uuid.MustParse(settingsUUID)
			apiParams.SettingsUuid = &u
		}

		resp, err := a.client.api.AzureLocationsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		locations = make([]AzureLocation, 0, len(*resp.JSON200))
		for _, l := range *resp.JSON200 {
			location := AzureLocation{
				Name: l.Name,
			}
			if l.Uuid != nil {
				location.UUID = l.Uuid.String()
			}
			locations = append(locations, location)
		}

		return nil
	})

	return locations, err
}

// ListAzureSizes lists available Azure VM sizes
func (a *AzureClient) ListAzureSizes(ctx context.Context, settingsUUID string) ([]AzureSize, error) {
	var sizes []AzureSize

	err := a.client.doWithRetry(ctx, func() error {
		apiParams := &client.AzureSizesListParams{}

		if settingsUUID != "" {
			u := uuid.MustParse(settingsUUID)
			apiParams.SettingsUuid = &u
		}

		resp, err := a.client.api.AzureSizesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		sizes = make([]AzureSize, 0, len(*resp.JSON200))
		for _, s := range *resp.JSON200 {
			size := AzureSize{
				Name:  s.Name,
				Cores: s.NumberOfCores,
				RAM:   s.MemoryInMb,
				Disk:  s.OsDiskSizeInMb,
			}
			if s.Uuid != nil {
				size.UUID = s.Uuid.String()
			}
			sizes = append(sizes, size)
		}

		return nil
	})

	return sizes, err
}

// WaitForAzureVMState waits for a VM to reach the desired state
func (a *AzureClient) WaitForAzureVMState(ctx context.Context, vmUUID string, desiredState string, pollInterval time.Duration) (*AzureVirtualMachine, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			vm, err := a.GetAzureVM(ctx, vmUUID)
			if err != nil {
				return nil, err
			}

			if vm.RuntimeState == desiredState || vm.State == desiredState {
				return vm, nil
			}

			// Check for error states
			if vm.State == "Erred" {
				return vm, fmt.Errorf("VM entered error state: %s", vm.ErrorMessage)
			}
		}
	}
}

// mapAzureVirtualMachine maps Waldur AzureVirtualMachine to our type
func mapAzureVirtualMachine(v *client.AzureVirtualMachine) AzureVirtualMachine {
	vm := AzureVirtualMachine{
		Name:         safeString(v.Name),
		Description:  safeString(v.Description),
		BackendID:    safeString(v.BackendId),
		RuntimeState: safeString(v.RuntimeState),
		ImageName:    safeString(v.ImageName),
		SizeName:     safeString(v.SizeName),
		LocationName: safeString(v.LocationName),
		ErrorMessage: safeString(v.ErrorMessage),
	}
	if v.Uuid != nil {
		vm.UUID = v.Uuid.String()
	}
	if v.State != nil {
		vm.State = string(*v.State)
	}
	if v.Cores != nil {
		vm.Cores = *v.Cores
	}
	if v.Ram != nil {
		vm.RAM = *v.Ram
	}
	if v.Disk != nil {
		vm.Disk = *v.Disk
	}
	if v.ServiceSettingsUuid != nil {
		vm.ServiceSettingsUUID = v.ServiceSettingsUuid.String()
	}
	if v.ProjectUuid != nil {
		vm.ProjectUUID = v.ProjectUuid.String()
	}
	if v.Created != nil {
		vm.CreatedAt = *v.Created
	}
	return vm
}

