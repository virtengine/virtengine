package provider_daemon

import (
	"context"
	"fmt"
	"strings"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// ResourceResolver resolves allocation IDs to backend resource UUIDs.
type ResourceResolver interface {
	ResolveResourceUUID(ctx context.Context, allocationID string) (string, error)
}

// WaldurProvisioner provisions VMs via Waldur/OpenStack backend.
type WaldurProvisioner struct {
	lifecycle *waldur.LifecycleClient
	resolver  ResourceResolver
}

// NewWaldurProvisioner creates a Waldur provisioner.
func NewWaldurProvisioner(lifecycle *waldur.LifecycleClient, resolver ResourceResolver) *WaldurProvisioner {
	return &WaldurProvisioner{lifecycle: lifecycle, resolver: resolver}
}

// Name returns the provisioner name.
func (p *WaldurProvisioner) Name() string {
	return "waldur"
}

// CanHandle returns true for VM service types.
func (p *WaldurProvisioner) CanHandle(serviceType marketplace.ServiceType) bool {
	return serviceType == marketplace.ServiceTypeVM
}

// Provision resolves resource state via Waldur and returns provisioning status.
func (p *WaldurProvisioner) Provision(ctx context.Context, req ProvisioningRequest) (*ProvisioningResult, error) {
	if p.lifecycle == nil {
		return nil, fmt.Errorf("waldur lifecycle client not configured")
	}
	if p.resolver == nil {
		return nil, fmt.Errorf("resource resolver not configured")
	}

	resourceUUID := req.ResourceID
	if resourceUUID == "" {
		resolved, err := p.resolver.ResolveResourceUUID(ctx, req.AllocationID)
		if err != nil {
			return nil, err
		}
		resourceUUID = resolved
	}

	state, err := p.lifecycle.GetResourceState(ctx, resourceUUID)
	if err != nil {
		return nil, err
	}

	allocationState := mapWaldurStateToAllocationState(string(state))
	result := &ProvisioningResult{
		State:      allocationState,
		Phase:      phaseFromAllocationState(allocationState),
		ResourceID: resourceUUID,
		Message:    fmt.Sprintf("waldur state: %s", strings.ToLower(string(state))),
	}

	switch allocationState {
	case marketplace.AllocationStateActive:
		result.Progress = 100
	case marketplace.AllocationStateProvisioning:
		result.Progress = 50
	}

	return result, nil
}

// WaldurStateResolver resolves resource UUIDs from Waldur bridge state.
type WaldurStateResolver struct {
	store *WaldurBridgeStateStore
}

// NewWaldurStateResolver creates a resolver using the Waldur bridge state store.
func NewWaldurStateResolver(store *WaldurBridgeStateStore) *WaldurStateResolver {
	return &WaldurStateResolver{store: store}
}

// ResolveResourceUUID returns resource UUID for an allocation.
func (r *WaldurStateResolver) ResolveResourceUUID(_ context.Context, allocationID string) (string, error) {
	if r == nil || r.store == nil {
		return "", fmt.Errorf("waldur state store not configured")
	}
	state, err := r.store.Load()
	if err != nil {
		return "", err
	}
	if state == nil || state.Mappings == nil {
		return "", fmt.Errorf("waldur mapping not found for %s", allocationID)
	}
	mapping := state.Mappings[allocationID]
	if mapping == nil {
		return "", fmt.Errorf("waldur mapping not found for %s", allocationID)
	}
	if mapping.ResourceUUID == "" {
		return "", fmt.Errorf("waldur resource UUID not yet available for %s", allocationID)
	}
	return mapping.ResourceUUID, nil
}

func phaseFromAllocationState(state marketplace.AllocationState) marketplace.ProvisioningPhase {
	switch state {
	case marketplace.AllocationStateProvisioning:
		return marketplace.ProvisioningPhaseProvisioning
	case marketplace.AllocationStateActive:
		return marketplace.ProvisioningPhaseActive
	case marketplace.AllocationStateTerminating, marketplace.AllocationStateTerminated:
		return marketplace.ProvisioningPhaseTerminated
	case marketplace.AllocationStateFailed:
		return marketplace.ProvisioningPhaseFailed
	default:
		return marketplace.ProvisioningPhaseRequested
	}
}
