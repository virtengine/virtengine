//go:build integration

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func TestKubernetesRuntimeReconcilesPodReadiness(t *testing.T) {
	fakeClient := newFakeKubernetesRuntimeClient()
	runtime, err := newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
		ProviderID:        "provider-1",
		ResourcePrefix:    "ve",
		ReconcileInterval: 10 * time.Millisecond,
		NewClient: func(kubeconfig string) (kubernetesRuntimeClient, error) {
			return fakeClient, nil
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manifest := testContainerManifest()
	namespace := testWorkloadNamespace("ve", "alloc-1", "alloc-1")
	fakeClient.SetPodReady(namespace, "web", false)

	workload, err := runtime.adapter.Deploy(ctx, manifest, "alloc-1", "alloc-1", provider_daemon.DeploymentOptions{})
	require.NoError(t, err)
	require.Equal(t, provider_daemon.WorkloadStateDeploying, workload.State)

	runtime.Start(ctx)
	fakeClient.SetPodReady(workload.Namespace, "web", true)

	require.Eventually(t, func() bool {
		status, err := runtime.adapter.GetStatus(ctx, workload.ID)
		return err == nil && status.State == provider_daemon.WorkloadStateRunning
	}, 2*time.Second, 20*time.Millisecond)
}

func TestContainerProvisionerRetriesFailedWorkload(t *testing.T) {
	fakeClient := newFakeKubernetesRuntimeClient()
	fakeClient.failOnCreate = true

	runtime, err := newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
		ProviderID:     "provider-1",
		ResourcePrefix: "ve",
		NewClient: func(kubeconfig string) (kubernetesRuntimeClient, error) {
			return fakeClient, nil
		},
	})
	require.NoError(t, err)

	provisioner := provider_daemon.NewContainerProvisioner(runtime.adapter, time.Minute, false)
	req := provider_daemon.ProvisioningRequest{
		AllocationID: "alloc-1",
		ServiceType:  marketplace.ServiceTypeContainer,
		Specifications: map[string]string{
			marketplace.SpecKeyContainerImage:    "nginx:latest",
			marketplace.SpecKeyContainerCPU:      "1",
			marketplace.SpecKeyContainerMemoryMB: "256",
			marketplace.SpecKeyContainerPorts:    "80",
		},
	}

	_, err = provisioner.Provision(context.Background(), req)
	require.Error(t, err)

	fakeClient.failOnCreate = false
	result, err := provisioner.Provision(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, marketplace.AllocationStateActive, result.State)
	require.Equal(t, marketplace.ProvisioningPhaseActive, result.Phase)
}

func testWorkloadNamespace(prefix, deploymentID, leaseID string) string {
	sum := sha256.Sum256([]byte(deploymentID + ":" + leaseID))
	return prefix + "-" + hex.EncodeToString(sum[:8])
}
