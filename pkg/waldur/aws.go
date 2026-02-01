// Package waldur provides AWS operations via Waldur API
//
// VE-2024: AWS instance/volume management via Waldur
package waldur

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	client "github.com/waldur/go-client"
)

// AWSClient provides AWS operations through Waldur
type AWSClient struct {
	client *Client
}

// NewAWSClient creates a new AWS client
func NewAWSClient(c *Client) *AWSClient {
	return &AWSClient{client: c}
}

// AWSInstance represents an AWS EC2 instance in Waldur
type AWSInstance struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	BackendID           string    `json:"backend_id"`
	Cores               int       `json:"cores"`
	RAM                 int       `json:"ram"`  // MiB
	Disk                int       `json:"disk"` // MiB
	ExternalIPs         []string  `json:"external_ips"`
	InternalIPs         []string  `json:"internal_ips"`
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	ErrorMessage        string    `json:"error_message"`
	CreatedAt           time.Time `json:"created"`
}

// AWSVolume represents an AWS EBS volume in Waldur
type AWSVolume struct {
	UUID                string    `json:"uuid"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	State               string    `json:"state"`
	BackendID           string    `json:"backend_id"`
	Size                int       `json:"size"` // GiB
	ServiceSettingsUUID string    `json:"service_settings_uuid"`
	ProjectUUID         string    `json:"project_uuid"`
	CreatedAt           time.Time `json:"created"`
}

// ListAWSInstancesParams contains parameters for listing AWS instances
type ListAWSInstancesParams struct {
	ServiceSettingsUUID string
	ProjectUUID         string
	CustomerUUID        string
	State               string
	Name                string
	Page                int
	PageSize            int
}

// ListAWSInstances lists AWS instances via Waldur
func (a *AWSClient) ListAWSInstances(ctx context.Context, params ListAWSInstancesParams) ([]AWSInstance, error) {
	var instances []AWSInstance

	err := a.client.doWithRetry(ctx, func() error {
		apiParams := &client.AwsInstancesListParams{}

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

		resp, err := a.client.api.AwsInstancesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		instances = make([]AWSInstance, 0, len(*resp.JSON200))
		for _, i := range *resp.JSON200 {
			instance := mapAWSInstance(&i)
			instances = append(instances, instance)
		}

		return nil
	})

	return instances, err
}

// GetAWSInstance retrieves a specific AWS instance
func (a *AWSClient) GetAWSInstance(ctx context.Context, instanceUUID string) (*AWSInstance, error) {
	var instance *AWSInstance

	err := a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := a.client.api.AwsInstancesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		i := mapAWSInstance(resp.JSON200)
		instance = &i
		return nil
	})

	return instance, err
}

// DeleteAWSInstance deletes an AWS instance
func (a *AWSClient) DeleteAWSInstance(ctx context.Context, instanceUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := a.client.api.AwsInstancesDestroyWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StartAWSInstance starts a stopped AWS instance
func (a *AWSClient) StartAWSInstance(ctx context.Context, instanceUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := a.client.api.AwsInstancesStartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// StopAWSInstance stops a running AWS instance
func (a *AWSClient) StopAWSInstance(ctx context.Context, instanceUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := a.client.api.AwsInstancesStopWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// RestartAWSInstance restarts an AWS instance
func (a *AWSClient) RestartAWSInstance(ctx context.Context, instanceUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(instanceUUID)
		resp, err := a.client.api.AwsInstancesRestartWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// ListAWSVolumesParams contains parameters for listing AWS volumes
type ListAWSVolumesParams struct {
	Page     int
	PageSize int
}

// ListAWSVolumes lists AWS volumes via Waldur
func (a *AWSClient) ListAWSVolumes(ctx context.Context, params ListAWSVolumesParams) ([]AWSVolume, error) {
	var volumes []AWSVolume

	err := a.client.doWithRetry(ctx, func() error {
		apiParams := &client.AwsVolumesListParams{}

		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := a.client.api.AwsVolumesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		volumes = make([]AWSVolume, 0, len(*resp.JSON200))
		for _, v := range *resp.JSON200 {
			volume := mapAWSVolume(&v)
			volumes = append(volumes, volume)
		}

		return nil
	})

	return volumes, err
}

// GetAWSVolume retrieves a specific AWS volume
func (a *AWSClient) GetAWSVolume(ctx context.Context, volumeUUID string) (*AWSVolume, error) {
	var volume *AWSVolume

	err := a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(volumeUUID)
		resp, err := a.client.api.AwsVolumesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		v := mapAWSVolume(resp.JSON200)
		volume = &v
		return nil
	})

	return volume, err
}

// DeleteAWSVolume deletes an AWS volume
func (a *AWSClient) DeleteAWSVolume(ctx context.Context, volumeUUID string) error {
	return a.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(volumeUUID)
		resp, err := a.client.api.AwsVolumesDestroyWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		return nil
	})
}

// WaitForAWSInstanceState waits for an instance to reach the desired state
func (a *AWSClient) WaitForAWSInstanceState(ctx context.Context, instanceUUID string, desiredState string, pollInterval time.Duration) (*AWSInstance, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			instance, err := a.GetAWSInstance(ctx, instanceUUID)
			if err != nil {
				return nil, err
			}

			if instance.State == desiredState {
				return instance, nil
			}

			// Check for error states
			if instance.State == "Erred" {
				return instance, fmt.Errorf("instance entered error state: %s", instance.ErrorMessage)
			}
		}
	}
}

// mapAWSInstance maps Waldur AwsInstance to our type
func mapAWSInstance(i *client.AwsInstance) AWSInstance {
	instance := AWSInstance{
		Name:         safeString(i.Name),
		Description:  safeString(i.Description),
		BackendID:    safeString(i.BackendId),
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

// mapAWSVolume maps Waldur AwsVolume to our type
func mapAWSVolume(v *client.AwsVolume) AWSVolume {
	volume := AWSVolume{
		Name:        safeString(v.Name),
		Description: safeString(v.Description),
		BackendID:   safeString(v.BackendId),
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

