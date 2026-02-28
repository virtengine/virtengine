//go:build integration

package provider_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/tests/integration/provider/providerharness"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestKubernetesProvisionerClusterLifecycle(t *testing.T) {
	ctx := context.Background()
	controlPlane := startProviderControlPlane(t)

	adapter := provider_daemon.NewKubernetesAdapter(provider_daemon.KubernetesAdapterConfig{
		Client:         controlPlane.NewKubernetesClient(),
		ProviderID:     "provider-integration",
		ResourcePrefix: "ve",
	})
	provisioner := provider_daemon.NewContainerProvisioner(adapter, 10*time.Second, false)

	req := provider_daemon.ProvisioningRequest{
		AllocationID: "alloc-int",
		ServiceType:  marketplace.ServiceTypeContainer,
		Specifications: map[string]string{
			marketplace.SpecKeyContainerImage:    "nginx:1.27",
			marketplace.SpecKeyContainerCPU:      "2",
			marketplace.SpecKeyContainerMemoryMB: "512",
			marketplace.SpecKeyContainerPorts:    "8080",
		},
	}

	initial, err := provisioner.Provision(ctx, req)
	require.NoError(t, err)
	require.Equal(t, marketplace.AllocationStateProvisioning, initial.State)
	require.Equal(t, marketplace.ProvisioningPhaseProvisioning, initial.Phase)

	workload, err := adapter.GetWorkloadByLease(req.AllocationID)
	require.NoError(t, err)

	exists, err := controlPlane.NamespaceExists(ctx, workload.Namespace)
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"alloc-"+req.AllocationID,
		corev1.PodRunning,
		true,
		"",
		"running",
		"",
	))

	active, err := provisioner.Provision(ctx, req)
	require.NoError(t, err)
	require.Equal(t, marketplace.AllocationStateActive, active.State)
	require.Equal(t, marketplace.ProvisioningPhaseActive, active.Phase)
	require.NotEmpty(t, active.Endpoints)

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"alloc-"+req.AllocationID,
		corev1.PodFailed,
		false,
		"pod exited unexpectedly",
		"terminated",
		"container crashed",
	))

	failed, err := provisioner.Provision(ctx, req)
	require.NoError(t, err)
	require.Equal(t, marketplace.AllocationStateFailed, failed.State)
	require.Equal(t, marketplace.ProvisioningPhaseFailed, failed.Phase)
	require.Contains(t, failed.Message, "container crashed")

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"alloc-"+req.AllocationID,
		corev1.PodRunning,
		true,
		"",
		"running",
		"",
	))

	recovered, err := provisioner.Provision(ctx, req)
	require.NoError(t, err)
	require.Equal(t, marketplace.AllocationStateActive, recovered.State)
	require.Equal(t, marketplace.ProvisioningPhaseActive, recovered.Phase)

	require.NoError(t, adapter.Terminate(ctx, workload.ID))
	require.NoError(t, controlPlane.WaitForNamespaceDeleted(ctx, workload.Namespace))
}

func startProviderControlPlane(t *testing.T) *providerharness.ControlPlane {
	t.Helper()

	controlPlane, err := providerharness.StartControlPlane()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, controlPlane.Stop())
	})
	return controlPlane
}
