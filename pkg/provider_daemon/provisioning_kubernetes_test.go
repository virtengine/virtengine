package provider_daemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestContainerProvisionerProvision(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-1",
	})
	provisioner := NewContainerProvisioner(adapter, 0, false)

	req := ProvisioningRequest{
		AllocationID: "alloc-1",
		ServiceType:  marketplace.ServiceTypeContainer,
		Specifications: map[string]string{
			marketplace.SpecKeyContainerImage:    "nginx:latest",
			marketplace.SpecKeyContainerCPU:      "1",
			marketplace.SpecKeyContainerMemoryMB: "256",
			marketplace.SpecKeyContainerPorts:    "80",
		},
	}

	result, err := provisioner.Provision(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, marketplace.AllocationStateActive, result.State)
	require.Equal(t, marketplace.ProvisioningPhaseActive, result.Phase)
}
