package provider_daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ContainerProvisioner provisions container workloads using KubernetesAdapter.
type ContainerProvisioner struct {
	adapter *KubernetesAdapter
	timeout time.Duration
	dryRun  bool
}

// NewContainerProvisioner creates a container provisioner.
func NewContainerProvisioner(adapter *KubernetesAdapter, timeout time.Duration, dryRun bool) *ContainerProvisioner {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &ContainerProvisioner{adapter: adapter, timeout: timeout, dryRun: dryRun}
}

// Name returns the provisioner name.
func (p *ContainerProvisioner) Name() string {
	return "kubernetes"
}

// CanHandle returns true for container service types.
func (p *ContainerProvisioner) CanHandle(serviceType marketplace.ServiceType) bool {
	return serviceType == marketplace.ServiceTypeContainer
}

// Provision deploys a container workload.
func (p *ContainerProvisioner) Provision(ctx context.Context, req ProvisioningRequest) (*ProvisioningResult, error) {
	if p.adapter == nil {
		return nil, fmt.Errorf("kubernetes adapter not configured")
	}

	spec, err := marketplace.ParseContainerServiceSpec(req.Specifications)
	if err != nil {
		return nil, err
	}

	if workload, err := p.adapter.GetWorkloadByLease(req.AllocationID); err == nil && workload != nil {
		status, err := p.adapter.GetStatus(ctx, workload.ID)
		if err != nil {
			return nil, err
		}
		return resultFromWorkload(workload, status), nil
	}

	manifest := manifestFromContainerSpec(req.AllocationID, spec)
	opts := DeploymentOptions{
		DryRun:  p.dryRun,
		Timeout: p.timeout,
	}

	workload, err := p.adapter.Deploy(ctx, manifest, req.AllocationID, req.AllocationID, opts)
	if err != nil {
		return &ProvisioningResult{
			State:    marketplace.AllocationStateFailed,
			Phase:    marketplace.ProvisioningPhaseFailed,
			Message:  err.Error(),
			Progress: 0,
		}, err
	}

	status, err := p.adapter.GetStatus(ctx, workload.ID)
	if err != nil {
		return &ProvisioningResult{
			State:      marketplace.AllocationStateProvisioning,
			Phase:      marketplace.ProvisioningPhaseProvisioning,
			Message:    err.Error(),
			Progress:   10,
			ResourceID: workload.ID,
		}, err
	}

	return resultFromWorkload(workload, status), nil
}

func manifestFromContainerSpec(allocationID string, spec marketplace.ContainerServiceSpec) *Manifest {
	service := ServiceSpec{
		Name:    fmt.Sprintf("alloc-%s", allocationID),
		Type:    "container",
		Image:   spec.Image,
		Command: spec.Command,
		Args:    spec.Args,
		Env:     spec.Env,
		Resources: ResourceSpec{
			CPU:    int64(spec.CPU) * 1000,
			Memory: int64(spec.MemoryMB) * 1024 * 1024,
		},
	}

	for _, port := range spec.Ports {
		if port <= 0 || port > 65535 {
			continue
		}
		service.Ports = append(service.Ports, PortSpec{
			ContainerPort: int32(port),
			Expose:        true,
		})
	}

	return &Manifest{
		Version:  ManifestVersionV1,
		Name:     fmt.Sprintf("allocation-%s", allocationID),
		Services: []ServiceSpec{service},
	}
}

func resultFromWorkload(workload *DeployedWorkload, status *WorkloadStatusUpdate) *ProvisioningResult {
	result := &ProvisioningResult{
		ResourceID: workload.ID,
		Endpoints:  map[string]string{},
	}
	if status != nil {
		result.Message = status.Message
	}

	switch workload.State {
	case WorkloadStateRunning:
		result.State = marketplace.AllocationStateActive
		result.Phase = marketplace.ProvisioningPhaseActive
		result.Progress = 100
	case WorkloadStateFailed:
		result.State = marketplace.AllocationStateFailed
		result.Phase = marketplace.ProvisioningPhaseFailed
	default:
		result.State = marketplace.AllocationStateProvisioning
		result.Phase = marketplace.ProvisioningPhaseProvisioning
		result.Progress = 50
	}

	for _, endpoint := range workload.Endpoints {
		key := endpoint.Service
		if key == "" {
			key = fmt.Sprintf("port-%d", endpoint.Port)
		}
		address := endpoint.ExternalAddress
		if address == "" {
			address = endpoint.InternalAddress
		}
		if address != "" {
			result.Endpoints[key] = fmt.Sprintf("%s:%d", address, endpoint.Port)
		}
	}

	return result
}
