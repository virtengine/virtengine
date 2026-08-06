//go:build e2e.integration

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	slurm_k8s "github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
	"github.com/virtengine/virtengine/tests/integration/provider/providerharness"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestSLURMKubernetesBootstrapE2E(t *testing.T) {
	ctx := context.Background()
	controlPlane := startE2EControlPlane(t)
	slurmHarness := providerharness.NewSLURMClusterHarness(controlPlane)
	reporter := &providerharness.RecordingReporter{}
	statusCh := make(chan slurm_k8s.ClusterStatusUpdate, 16)
	phaseCh := make(chan slurm_k8s.PhaseTransitionEvent, 16)

	adapter := slurm_k8s.NewSLURMKubernetesAdapter(slurm_k8s.AdapterConfig{
		Helm:                slurmHarness,
		K8s:                 slurmHarness,
		Reporter:            reporter,
		StatusChan:          statusCh,
		PhaseEventChan:      phaseCh,
		ChartPath:           "/charts/slurm",
		HealthCheckInterval: 20 * time.Millisecond,
	})
	require.NoError(t, adapter.Start(ctx))
	t.Cleanup(func() {
		require.NoError(t, adapter.Stop())
	})

	config := slurm_k8s.DeploymentConfig{
		ClusterID:        "cluster-e2e",
		ClusterName:      "E2E Cluster",
		Namespace:        "slurm-e2e",
		HelmReleaseName:  "slurm-e2e-cluster",
		HelmChartPath:    "/charts/slurm",
		ProviderAddress:  "virtengine1e2eprovider",
		ProviderEndpoint: "http://provider-daemon:8081",
		Template: &hpctypes.ClusterTemplate{
			Partitions: []hpctypes.PartitionConfig{
				{Name: "gpu", Nodes: 2, State: "up"},
			},
		},
	}

	cluster, err := adapter.Bootstrap(ctx, config, slurm_k8s.BootstrapOptions{
		DeployOptions: slurm_k8s.DeployOptions{
			ReadyTimeout:      2 * time.Second,
			RollbackOnFailure: true,
			MinComputeReady:   2,
		},
	})
	require.NoError(t, err)
	require.Equal(t, slurm_k8s.ClusterStateRunning, cluster.State)
	require.Equal(t, slurm_k8s.DeploymentPhaseComplete, cluster.Phase)

	_, err = controlPlane.Clientset.AppsV1().StatefulSets(config.Namespace).Get(ctx, config.HelmReleaseName+"-controller", metav1.GetOptions{})
	require.NoError(t, err)
	_, err = controlPlane.Clientset.AppsV1().StatefulSets(config.Namespace).Get(ctx, config.HelmReleaseName+"-compute", metav1.GetOptions{})
	require.NoError(t, err)

	phases := drainPhaseTransitions(phaseCh)
	requirePhaseOrder(t, phases,
		slurm_k8s.DeploymentPhaseHelmInstalling,
		slurm_k8s.DeploymentPhaseControllerReady,
		slurm_k8s.DeploymentPhaseDatabaseReady,
		slurm_k8s.DeploymentPhaseComputeReady,
		slurm_k8s.DeploymentPhaseRegistrationReady,
		slurm_k8s.DeploymentPhaseComplete,
	)

	statuses := drainClusterStatuses(statusCh)
	require.NotEmpty(t, statuses)
	require.Equal(t, slurm_k8s.ClusterStateRunning, statuses[len(statuses)-1].State)

	health, err := adapter.GetClusterHealth(ctx, config.ClusterID)
	require.NoError(t, err)
	require.True(t, health.ControllerReady)
	require.True(t, health.DatabaseReady)
	require.EqualValues(t, 2, health.ComputeNodesReady)

	capacity, err := adapter.GetClusterCapacity(ctx, config.ClusterID)
	require.NoError(t, err)
	require.EqualValues(t, 2, capacity.TotalNodes)
	require.EqualValues(t, 2, capacity.AvailableNodes)

	require.NotEmpty(t, reporter.StatusReports)

	require.NoError(t, adapter.Terminate(ctx, config.ClusterID))
	_, err = controlPlane.Clientset.AppsV1().StatefulSets(config.Namespace).Get(ctx, config.HelmReleaseName+"-controller", metav1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
}

func drainPhaseTransitions(ch <-chan slurm_k8s.PhaseTransitionEvent) []slurm_k8s.DeploymentPhase {
	phases := make([]slurm_k8s.DeploymentPhase, 0)
	for {
		select {
		case event := <-ch:
			phases = append(phases, event.ToPhase)
		default:
			return phases
		}
	}
}

func drainClusterStatuses(ch <-chan slurm_k8s.ClusterStatusUpdate) []slurm_k8s.ClusterStatusUpdate {
	updates := make([]slurm_k8s.ClusterStatusUpdate, 0)
	for {
		select {
		case update := <-ch:
			updates = append(updates, update)
		default:
			return updates
		}
	}
}

func requirePhaseOrder(t *testing.T, phases []slurm_k8s.DeploymentPhase, required ...slurm_k8s.DeploymentPhase) {
	t.Helper()

	index := 0
	for _, phase := range phases {
		if index < len(required) && phase == required[index] {
			index++
		}
	}
	require.Equal(t, len(required), index, "missing required phase sequence in %v", phases)
}
