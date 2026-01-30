// Package hpc contains integration tests for HPC SLURM deployment.
//
//go:build e2e.integration

package hpc

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
)

// TestSLURMDeploymentKind tests SLURM deployment on a kind cluster.
// Prerequisites:
// - kind installed and available in PATH
// - kubectl configured for the kind cluster
// - Helm 3 installed
//
// Run with: go test -tags="e2e.integration" -v ./tests/integration/hpc/...
func TestSLURMDeploymentKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check prerequisites
	checkPrerequisites(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	clusterName := "slurm-test-" + randomSuffix()
	namespace := "slurm-test"

	// Create kind cluster
	t.Log("Creating kind cluster...")
	if err := createKindCluster(ctx, clusterName); err != nil {
		t.Fatalf("failed to create kind cluster: %v", err)
	}
	defer deleteKindCluster(clusterName)

	// Create namespace
	t.Log("Creating namespace...")
	if err := runKubectl(ctx, "create", "namespace", namespace); err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}

	// Deploy SLURM cluster using Helm
	t.Log("Deploying SLURM cluster...")
	chartPath := "../../../deploy/slurm/slurm-cluster"
	releaseName := "slurm-test"

	helmArgs := []string{
		"install", releaseName, chartPath,
		"--namespace", namespace,
		"--set", "cluster.id=test-cluster",
		"--set", "cluster.name=Test SLURM Cluster",
		"--set", "compute.replicas=2",
		"--set", "controller.persistence.size=1Gi",
		"--set", "database.persistence.size=1Gi",
		"--set", "mariadb.persistence.size=1Gi",
		"--wait",
		"--timeout", "10m",
	}

	if err := runHelm(ctx, helmArgs...); err != nil {
		// Get pod status for debugging
		runKubectl(ctx, "get", "pods", "-n", namespace)
		runKubectl(ctx, "describe", "pods", "-n", namespace)
		t.Fatalf("failed to deploy SLURM cluster: %v", err)
	}

	// Verify deployment
	t.Log("Verifying deployment...")
	if err := verifyDeployment(ctx, namespace, releaseName); err != nil {
		t.Fatalf("deployment verification failed: %v", err)
	}

	// Test SLURM functionality
	t.Log("Testing SLURM functionality...")
	if err := testSLURMFunctionality(ctx, namespace, releaseName); err != nil {
		t.Fatalf("SLURM functionality test failed: %v", err)
	}

	// Test scaling
	t.Log("Testing scaling...")
	if err := testScaling(ctx, namespace, releaseName, chartPath); err != nil {
		t.Fatalf("scaling test failed: %v", err)
	}

	// Cleanup
	t.Log("Cleaning up...")
	if err := runHelm(ctx, "uninstall", releaseName, "--namespace", namespace); err != nil {
		t.Errorf("failed to uninstall Helm release: %v", err)
	}

	t.Log("All tests passed!")
}

func checkPrerequisites(t *testing.T) {
	t.Helper()

	// Check kind
	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kind not found, skipping test")
	}

	// Check kubectl
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not found, skipping test")
	}

	// Check helm
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not found, skipping test")
	}
}

func createKindCluster(ctx context.Context, name string) error {
	// Create kind cluster config
	config := `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
`
	configFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(configFile.Name())

	if _, err := configFile.WriteString(config); err != nil {
		return err
	}
	configFile.Close()

	cmd := exec.CommandContext(ctx, "kind", "create", "cluster",
		"--name", name,
		"--config", configFile.Name(),
		"--wait", "5m")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func deleteKindCluster(name string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runKubectl(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runHelm(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func verifyDeployment(ctx context.Context, namespace, releaseName string) error {
	// Wait for controller to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-controller",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	// Wait for database to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-slurmdbd",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	// Wait for compute nodes to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-compute",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	return nil
}

func testSLURMFunctionality(ctx context.Context, namespace, releaseName string) error {
	controllerPod := releaseName + "-slurm-cluster-controller-0"

	// Test scontrol ping
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"scontrol", "ping"); err != nil {
		return err
	}

	// Test sinfo
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"sinfo"); err != nil {
		return err
	}

	// Test squeue
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"squeue"); err != nil {
		return err
	}

	return nil
}

func testScaling(ctx context.Context, namespace, releaseName, chartPath string) error {
	// Scale up to 4 nodes
	if err := runHelm(ctx, "upgrade", releaseName, chartPath,
		"--namespace", namespace,
		"--set", "compute.replicas=4",
		"--wait",
		"--timeout", "5m"); err != nil {
		return err
	}

	// Verify new nodes are registered
	controllerPod := releaseName + "-slurm-cluster-controller-0"
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"sinfo", "-N"); err != nil {
		return err
	}

	// Scale back down
	if err := runHelm(ctx, "upgrade", releaseName, chartPath,
		"--namespace", namespace,
		"--set", "compute.replicas=2",
		"--wait",
		"--timeout", "5m"); err != nil {
		return err
	}

	return nil
}

func randomSuffix() string {
	return time.Now().Format("150405")
}

// TestAdapterIntegration tests the SLURM Kubernetes adapter with mocks
func TestAdapterIntegration(t *testing.T) {
	helm := &slurm_k8s.MockHelmClient{
		InstallFunc: func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
			t.Logf("Mock Helm install: %s in %s", releaseName, namespace)
			return nil
		},
		UpgradeFunc: func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
			t.Logf("Mock Helm upgrade: %s", releaseName)
			return nil
		},
		UninstallFunc: func(ctx context.Context, releaseName, namespace string) error {
			t.Logf("Mock Helm uninstall: %s", releaseName)
			return nil
		},
	}

	k8s := &slurm_k8s.MockK8sChecker{
		StatefulSetStatus: map[string]*slurm_k8s.StatefulSetStatus{
			"test-ns/slurm-integration-test-controller": {ReadyReplicas: 1},
			"test-ns/slurm-integration-test-slurmdbd":   {ReadyReplicas: 1},
			"test-ns/slurm-integration-test-compute":    {ReadyReplicas: 2, Replicas: 2},
		},
		ExecOutput: map[string]string{
			"slurm-integration-test-controller-0:scontrol": "Slurmctld(primary) at slurm-integration-test-controller-0 is UP",
			"slurm-integration-test-controller-0:sinfo":    "compute-0 64 256000 (null) idle\ncompute-1 64 256000 (null) idle",
		},
	}

	reporter := &slurm_k8s.MockReporter{}

	adapter := slurm_k8s.NewSLURMKubernetesAdapter(slurm_k8s.AdapterConfig{
		Helm:      helm,
		K8s:       k8s,
		Reporter:  reporter,
		ChartPath: "/charts/slurm",
	})

	ctx := context.Background()

	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("failed to start adapter: %v", err)
	}
	defer adapter.Stop()

	// Deploy cluster
	config := slurm_k8s.DeploymentConfig{
		ClusterID:        "integration-test",
		ClusterName:      "Integration Test Cluster",
		Namespace:        "test-ns",
		ProviderAddress:  "virtengine1test",
		ProviderEndpoint: "https://provider.test:8443",
	}

	cluster, err := adapter.Deploy(ctx, config)
	if err != nil {
		t.Fatalf("failed to deploy cluster: %v", err)
	}

	if cluster.State != slurm_k8s.ClusterStateRunning {
		t.Errorf("expected state running, got %s", cluster.State)
	}

	// Get health
	health, err := adapter.GetClusterHealth(ctx, "integration-test")
	if err != nil {
		t.Fatalf("failed to get health: %v", err)
	}

	if !health.ControllerReady {
		t.Error("expected controller to be ready")
	}

	// Scale cluster
	if err := adapter.Scale(ctx, "integration-test", slurm_k8s.ScaleRequest{TargetNodes: 4}); err != nil {
		t.Fatalf("failed to scale cluster: %v", err)
	}

	// Terminate cluster
	if err := adapter.Terminate(ctx, "integration-test"); err != nil {
		t.Fatalf("failed to terminate cluster: %v", err)
	}

	// Verify reporter was called
	if len(reporter.StatusReports) == 0 {
		t.Error("expected status reports to be submitted")
	}
}
