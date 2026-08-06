//go:build e2e.integration

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/integration/provider/providerharness"
)

func TestKubernetesAdapterWaldurLifecycleE2E(t *testing.T) {
	ctx := context.Background()
	controlPlane := startE2EControlPlane(t)
	harness := newWaldurHarness(t)

	manifest := &provider_daemon.Manifest{
		Version: provider_daemon.ManifestVersionV1,
		Name:    "k8s-e2e",
		Services: []provider_daemon.ServiceSpec{
			{
				Name:  "api",
				Type:  "container",
				Image: "nginx",
				Tag:   "latest",
				Resources: provider_daemon.ResourceSpec{
					CPU:    500,
					Memory: 512 * 1024 * 1024,
				},
				Ports:   []provider_daemon.PortSpec{{Name: "http", ContainerPort: 8080, Expose: true}},
				Volumes: []provider_daemon.VolumeMountSpec{{Name: "data", MountPath: "/data"}},
			},
		},
		Volumes: []provider_daemon.VolumeSpec{{Name: "data", Type: "persistent", Size: 10 * 1024 * 1024 * 1024}},
	}

	order := harness.createOrder(ctx, "k8s-order", map[string]interface{}{"backend": "kubernetes"})
	resource := harness.waitForResource(order.UUID)

	statusCh := make(chan provider_daemon.WorkloadStatusUpdate, 16)
	adapter := provider_daemon.NewKubernetesAdapter(provider_daemon.KubernetesAdapterConfig{
		Client:           controlPlane.NewKubernetesClient(),
		ProviderID:       "provider-e2e",
		ResourcePrefix:   "e2e",
		StatusUpdateChan: statusCh,
	})

	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", provider_daemon.DeploymentOptions{})
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateDeploying, workload.State)

	exists, err := controlPlane.NamespaceExists(ctx, workload.Namespace)
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"api",
		corev1.PodRunning,
		true,
		"",
		"running",
		"",
	))

	readyStatus, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateRunning, readyStatus.State)
	require.NotEmpty(t, readyStatus.Message)

	harness.submitUsage(ctx, resource.UUID, workload.ID)
	require.Greater(t, len(harness.mock.GetUsageRecords(resource.UUID)), 0)

	require.NoError(t, adapter.Pause(ctx, workload.ID))
	pausedStatus, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStatePaused, pausedStatus.State)

	_, err = harness.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	require.Equal(t, "Stopped", harness.mock.GetResource(resource.UUID).State)

	_, err = harness.lifecycle.Start(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	require.Equal(t, "OK", harness.mock.GetResource(resource.UUID).State)

	require.NoError(t, adapter.Resume(ctx, workload.ID))
	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"api",
		corev1.PodRunning,
		true,
		"",
		"running",
		"",
	))

	resumedStatus, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateRunning, resumedStatus.State)

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"api",
		corev1.PodFailed,
		false,
		"readiness probe failed",
		"terminated",
		"panic: bootstrap error",
	))

	failedStatus, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateFailed, failedStatus.State)
	require.Contains(t, failedStatus.Message, "panic: bootstrap error")

	require.NoError(t, controlPlane.ReplaceDeploymentPods(
		ctx,
		workload.Namespace,
		"api",
		corev1.PodRunning,
		true,
		"",
		"running",
		"",
	))

	workload, err = adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", provider_daemon.DeploymentOptions{})
	require.NoError(t, err)

	recoveredStatus, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateRunning, recoveredStatus.State)

	require.NoError(t, adapter.Terminate(ctx, workload.ID))
	require.NoError(t, controlPlane.WaitForNamespaceDeleted(ctx, workload.Namespace))

	_, err = harness.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	require.Equal(t, "Terminated", harness.mock.GetResource(resource.UUID).State)
}

func startE2EControlPlane(t *testing.T) *providerharness.ControlPlane {
	t.Helper()

	controlPlane, err := providerharness.StartControlPlane()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, controlPlane.Stop())
	})
	return controlPlane
}
