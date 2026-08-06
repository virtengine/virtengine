//go:build integration

package provider_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	slurm_k8s "github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
	"github.com/virtengine/virtengine/tests/integration/provider/providerharness"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestSLURMKubernetesClusterReconcileAndCleanup(t *testing.T) {
	ctx := context.Background()
	controlPlane := startProviderControlPlane(t)
	slurmHarness := providerharness.NewSLURMClusterHarness(controlPlane)
	reporter := &providerharness.RecordingReporter{}

	adapter := slurm_k8s.NewSLURMKubernetesAdapter(slurm_k8s.AdapterConfig{
		Helm:                slurmHarness,
		K8s:                 slurmHarness,
		Reporter:            reporter,
		ChartPath:           "/charts/slurm",
		HealthCheckInterval: 20 * time.Millisecond,
	})
	require.NoError(t, adapter.Start(ctx))
	t.Cleanup(func() {
		require.NoError(t, adapter.Stop())
	})

	config := slurm_k8s.DeploymentConfig{
		ClusterID:        "cluster-int",
		ClusterName:      "Integration Cluster",
		Namespace:        "slurm-integration",
		HelmReleaseName:  "slurm-int-cluster",
		HelmChartPath:    "/charts/slurm",
		ProviderAddress:  "virtengine1integrationprovider",
		ProviderEndpoint: "http://provider-daemon:8081",
		Template: &hpctypes.ClusterTemplate{
			Partitions: []hpctypes.PartitionConfig{
				{Name: "cpu", Nodes: 2, State: "up"},
			},
		},
	}

	cluster, err := adapter.Bootstrap(ctx, config, slurm_k8s.BootstrapOptions{
		DeployOptions: slurm_k8s.DeployOptions{
			ReadyTimeout:    2 * time.Second,
			MinComputeReady: 2,
		},
	})
	require.NoError(t, err)
	require.Equal(t, slurm_k8s.ClusterStateRunning, cluster.State)

	_, err = controlPlane.Clientset.AppsV1().StatefulSets(config.Namespace).Get(ctx, config.HelmReleaseName+"-compute", metav1.GetOptions{})
	require.NoError(t, err)

	health, err := adapter.GetClusterHealth(ctx, config.ClusterID)
	require.NoError(t, err)
	require.True(t, health.ControllerReady)
	require.True(t, health.DatabaseReady)
	require.EqualValues(t, 2, health.ComputeNodesReady)
	require.EqualValues(t, 2, health.ComputeNodesTotal)

	lifecycle := slurm_k8s.NewLifecycleManager(adapter)
	slurmHarness.SetNodeState(config.Namespace, config.HelmReleaseName, "compute-1", "down")
	require.NoError(t, slurmHarness.SetStatefulSetReadyReplicas(
		ctx,
		config.Namespace,
		config.HelmReleaseName+"-compute",
		1,
	))
	require.NoError(t, lifecycle.ReconcileCluster(ctx, config.ClusterID))

	degradedCluster, err := adapter.GetCluster(config.ClusterID)
	require.NoError(t, err)
	require.Equal(t, slurm_k8s.ClusterStateDegraded, degradedCluster.State)
	require.Contains(t, degradedCluster.StatusMessage, "Partially ready")

	calls := slurmHarness.ExecCalls()
	var resumed bool
	for _, call := range calls {
		if strings.Contains(strings.Join(call.Command, " "), "NodeName=compute-1") &&
			strings.Contains(strings.Join(call.Command, " "), "State=RESUME") {
			resumed = true
			break
		}
	}
	require.True(t, resumed, "expected reconcile to issue a resume for the down node")

	slurmHarness.SetNodeState(config.Namespace, config.HelmReleaseName, "compute-1", "idle")
	require.NoError(t, slurmHarness.SetStatefulSetReadyReplicas(
		ctx,
		config.Namespace,
		config.HelmReleaseName+"-compute",
		2,
	))
	require.NoError(t, lifecycle.ReconcileCluster(ctx, config.ClusterID))

	reconciledCluster, err := adapter.GetCluster(config.ClusterID)
	require.NoError(t, err)
	require.Equal(t, slurm_k8s.ClusterStateRunning, reconciledCluster.State)

	capacity, err := adapter.GetClusterCapacity(ctx, config.ClusterID)
	require.NoError(t, err)
	require.EqualValues(t, 2, capacity.TotalNodes)
	require.EqualValues(t, 2, capacity.AvailableNodes)
	require.EqualValues(t, 128, capacity.TotalCPUs)

	require.NoError(t, adapter.Terminate(ctx, config.ClusterID))
	_, err = adapter.GetCluster(config.ClusterID)
	require.Error(t, err)

	_, err = controlPlane.Clientset.AppsV1().StatefulSets(config.Namespace).Get(ctx, config.HelmReleaseName+"-controller", metav1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
}
