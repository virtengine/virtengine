// Package waldur provides OpenStack operations via Waldur API
//
// VE-2024: OpenStack instance/volume/tenant management via Waldur
package waldur

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	client "github.com/waldur/go-client"
)

// OpenStackClient provides OpenStack operations through Waldur
type OpenStackClient struct {
	client *Client
}

// NewOpenStackClient creates a new OpenStack client
func NewOpenStackClient(c *Client) *OpenStackClient {
	return &OpenStackClient{client: c}
}

// OpenStackInstance represents an OpenStack instance in Waldur
type OpenStackInstance struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	RuntimeState        string    `json:"runtime_state"`
	BackendID           string    `json:"backend_id"`
	Cores               int       `json:"cores"`
	RAM                 int       `json:"ram"`  // MiB
	Disk                int       `json:"disk"` // MiB
	ExternalIPs         []string  `json:"external_ips"`
	InternalIPs         []string  `json:"internal_ips"`
	ImageName           string    `json:"image_name"`
	FlavorName          string    `json:"flavor_name"`
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	ErrorMessage        string    `json:"error_message"`
	CreatedAt           time.Time `json:"created"`
}

// OpenStackVolume represents an OpenStack volume in Waldur
type OpenStackVolume struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	RuntimeState        string    `json:"runtime_state"`
	BackendID           string    `json:"backend_id"`
	Size                int       `json:"size"` // GiB
	Bootable            bool      `json:"bootable"`
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	CreatedAt           time.Time `json:"created"`
}

// OpenStackTenant represents an OpenStack tenant/project in Waldur
type OpenStackTenant struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	BackendID           string    `json:"backend_id"`
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	CreatedAt           time.Time `json:"created"`
}

// ListOpenStackInstancesParams contains parameters for listing instances
type ListOpenStackInstancesParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	RuntimeState        string
	Name                string
	Page                int
	PageSize            int
}

// ListOpenStackInstances lists OpenStack instances via Waldur
func (o *OpenStackClient) ListOpenStackInstances(ctx context.Context, params ListOpenStackInstancesParams) ([]OpenStackInstance, error) {
	var instances []OpenStackInstance

	err := o.client.doWithRetry(ctx, func() error {
		apiParams := &client.OpenstackInstancesListParams{}

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
		if params.RuntimeState != "" {
			apiParams.RuntimeState = &params.RuntimeState
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := o.client.api.OpenstackInstancesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		instances = make([]OpenStackInstance, 0, len(*resp.JSON200))
		for _, i := range *resp.JSON200 {
			instance := mapOpenStackInstance(&i)
			instances = append(instances, instance)
		}

		return nil
	})

	return instances, err
}

// GetOpenStackInstance retrieves a specific OpenStack instance
func (o *OpenStackClient) GetOpenStackInstance(ctx context.Context, instanceUUID string) (*OpenStackInstance, error) {
	var instance *OpenStackInstance

	err := o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := o.client.api.OpenstackInstancesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		i := mapOpenStackInstance(resp.JSON200)
		instance = &i
		return nil
	})

	return instance, err
}

// DeleteOpenStackInstance deletes an OpenStack instance via unlink
// Note: Waldur uses unlink rather than direct destroy for OpenStack instances
func (o *OpenStackClient) DeleteOpenStackInstance(ctx context.Context, instanceUUID string) error {
	return o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := o.client.api.OpenstackInstancesUnlinkWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StartOpenStackInstance starts a stopped OpenStack instance
func (o *OpenStackClient) StartOpenStackInstance(ctx context.Context, instanceUUID string) error {
	return o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := o.client.api.OpenstackInstancesStartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StopOpenStackInstance stops a running OpenStack instance
func (o *OpenStackClient) StopOpenStackInstance(ctx context.Context, instanceUUID string) error {
	return o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := o.client.api.OpenstackInstancesStopWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// RestartOpenStackInstance restarts an OpenStack instance
func (o *OpenStackClient) RestartOpenStackInstance(ctx context.Context, instanceUUID string) error {
	return o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := o.client.api.OpenstackInstancesRestartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// ListOpenStackVolumesParams contains parameters for listing volumes
type ListOpenStackVolumesParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	Name                string
	Page                int
	PageSize            int
}

// ListOpenStackVolumes lists OpenStack volumes via Waldur
func (o *OpenStackClient) ListOpenStackVolumes(ctx context.Context, params ListOpenStackVolumesParams) ([]OpenStackVolume, error) {
	var volumes []OpenStackVolume

	err := o.client.doWithRetry(ctx, func() error {
		apiParams := &client.OpenstackVolumesListParams{}

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

		resp, err := o.client.api.OpenstackVolumesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		volumes = make([]OpenStackVolume, 0, len(*resp.JSON200))
		for _, v := range *resp.JSON200 {
			volume := mapOpenStackVolume(&v)
			volumes = append(volumes, volume)
		}

		return nil
	})

	return volumes, err
}

// GetOpenStackVolume retrieves a specific OpenStack volume
func (o *OpenStackClient) GetOpenStackVolume(ctx context.Context, volumeUUID string) (*OpenStackVolume, error) {
	var volume *OpenStackVolume

	err := o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(volumeUUID)
		resp, err := o.client.api.OpenstackVolumesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		v := mapOpenStackVolume(resp.JSON200)
		volume = &v
		return nil
	})

	return volume, err
}

// DeleteOpenStackVolume deletes an OpenStack volume via unlink
// Note: Waldur uses unlink rather than direct destroy for OpenStack volumes
func (o *OpenStackClient) DeleteOpenStackVolume(ctx context.Context, volumeUUID string) error {
	return o.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(volumeUUID)
		resp, err := o.client.api.OpenstackVolumesUnlinkWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// ListOpenStackTenantsParams contains parameters for listing tenants
type ListOpenStackTenantsParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	Name                string
	Page                int
	PageSize            int
}

// ListOpenStackTenants lists OpenStack tenants via Waldur
func (o *OpenStackClient) ListOpenStackTenants(ctx context.Context, params ListOpenStackTenantsParams) ([]OpenStackTenant, error) {
	var tenants []OpenStackTenant

	err := o.client.doWithRetry(ctx, func() error {
		apiParams := &client.OpenstackTenantsListParams{}

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

		resp, err := o.client.api.OpenstackTenantsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		tenants = make([]OpenStackTenant, 0, len(*resp.JSON200))
		for _, t := range *resp.JSON200 {
			tenant := OpenStackTenant{
				Name:        safeString(t.Name),
				Description: safeString(t.Description),
				BackendID:   safeString(t.BackendId),
			}
			if t.Uuid != nil {
				tenant.UUID = t.Uuid.String()
			}
			if t.State != nil {
				tenant.State = string(*t.State)
			}
			if t.ServiceSettingsUuid != nil {
				tenant.ServiceSettingsUUID = t.ServiceSettingsUuid.String()
			}
			if t.ProjectUuid != nil {
				tenant.ProjectUUID = t.ProjectUuid.String()
			}
			if t.Created != nil {
				tenant.CreatedAt = *t.Created
			}
			tenants = append(tenants, tenant)
		}

		return nil
	})

	return tenants, err
}

// WaitForOpenStackInstanceState waits for an instance to reach the desired state
func (o *OpenStackClient) WaitForOpenStackInstanceState(ctx context.Context, instanceUUID string, desiredState string, pollInterval time.Duration) (*OpenStackInstance, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			instance, err := o.GetOpenStackInstance(ctx, instanceUUID)
			if err != nil {
				return nil, err
			}

			if instance.RuntimeState == desiredState {
				return instance, nil
			}

			// Check for error states
			if instance.State == "Erred" {
				return instance, fmt.Errorf("instance entered error state: %s", instance.ErrorMessage)
			}
		}
	}
}

// mapOpenStackInstance maps Waldur OpenStackInstance to our type
func mapOpenStackInstance(i *client.OpenStackInstance) OpenStackInstance {
	instance := OpenStackInstance{
		Name:         safeString(i.Name),
		Description:  safeString(i.Description),
		BackendID:    safeString(i.BackendId),
		RuntimeState: safeString(i.RuntimeState),
		ImageName:    safeString(i.ImageName),
		FlavorName:   safeString(i.FlavorName),
		ErrorMessage: safeString(i.ErrorMessage),
	}
	if i.Uuid != nil {
		instance.UUID = i.Uuid.String()
	}
	if i.State != nil {
		instance.State = string(*i.State)
	}
	if i.Cores != nil {
		instance.Cores = *i.Cores
	}
	if i.Ram != nil {
		instance.RAM = *i.Ram
	}
	if i.Disk != nil {
		instance.Disk = *i.Disk
	}
	if i.ExternalIps != nil {
		instance.ExternalIPs = *i.ExternalIps
	}
	if i.InternalIps != nil {
		instance.InternalIPs = *i.InternalIps
	}
	if i.ServiceSettingsUuid != nil {
		instance.ServiceSettingsUUID = i.ServiceSettingsUuid.String()
	}
	if i.ProjectUuid != nil {
		instance.ProjectUUID = i.ProjectUuid.String()
	}
	if i.Created != nil {
		instance.CreatedAt = *i.Created
	}
	return instance
}

// mapOpenStackVolume maps Waldur OpenStackVolume to our type
func mapOpenStackVolume(v *client.OpenStackVolume) OpenStackVolume {
	volume := OpenStackVolume{
		Name:         safeString(v.Name),
		Description:  safeString(v.Description),
		BackendID:    safeString(v.BackendId),
		RuntimeState: safeString(v.RuntimeState),
	}
	if v.Uuid != nil {
		volume.UUID = v.Uuid.String()
	}
	if v.State != nil {
		volume.State = string(*v.State)
	}
	if v.Size != nil {
		volume.Size = *v.Size
	}
	if v.Bootable != nil {
		volume.Bootable = *v.Bootable
	}
	if v.ServiceSettingsUuid != nil {
		volume.ServiceSettingsUUID = v.ServiceSettingsUuid.String()
	}
	if v.ProjectUuid != nil {
		volume.ProjectUUID = v.ProjectUuid.String()
	}
	if v.Created != nil {
		volume.CreatedAt = *v.Created
	}
	return volume
}
