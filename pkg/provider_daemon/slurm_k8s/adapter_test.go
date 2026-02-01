// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
package slurm_k8s

import (
	"context"
	"testing"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestNewSLURMKubernetesAdapter(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm:                &MockHelmClient{},
		K8s:                 &MockK8sChecker{},
		ChartPath:           "/charts/slurm",
		HealthCheckInterval: 30 * time.Second,
	})

	if adapter == nil {
		t.Fatal("expected adapter to be created")
	}

	if adapter.chartPath != "/charts/slurm" {
		t.Errorf("expected chart path /charts/slurm, got %s", adapter.chartPath)
	}

	if adapter.healthCheckInterval != 30*time.Second {
		t.Errorf("expected health check interval 30s, got %v", adapter.healthCheckInterval)
	}
}

func TestAdapterStartStop(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	ctx := context.Background()

	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("failed to start adapter: %v", err)
	}

	if !adapter.IsRunning() {
		t.Error("expected adapter to be running")
	}

	// Start again should be no-op
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("second start should not error: %v", err)
	}

	if err := adapter.Stop(); err != nil {
		t.Fatalf("failed to stop adapter: %v", err)
	}

	if adapter.IsRunning() {
		t.Error("expected adapter to be stopped")
	}
}

func TestDeploy(t *testing.T) {
	helmCalled := false
	helmClient := &MockHelmClient{
		InstallFunc: func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
			helmCalled = true
			if releaseName != "slurm-test-cluster" {
				t.Errorf("expected release name slurm-test-cluster, got %s", releaseName)
			}
			if namespace != "slurm-ns" {
				t.Errorf("expected namespace slurm-ns, got %s", namespace)
			}
			return nil
		},
	}

	k8s := &MockK8sChecker{
		StatefulSetStatus: map[string]*StatefulSetStatus{
			"slurm-ns/slurm-test-cluster-controller": {ReadyReplicas: 1},
			"slurm-ns/slurm-test-cluster-slurmdbd":   {ReadyReplicas: 1},
			"slurm-ns/slurm-test-cluster-compute":    {ReadyReplicas: 2, Replicas: 2},
		},
	}

	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm:      helmClient,
		K8s:       k8s,
		ChartPath: "/charts/slurm",
	})

	ctx := context.Background()
	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("failed to start adapter: %v", err)
	}
	defer func() { _ = adapter.Stop() }()

	config := DeploymentConfig{
		ClusterID:   "test-cluster",
		ClusterName: "Test Cluster",
		Namespace:   "slurm-ns",
		Template: &hpctypes.ClusterTemplate{
			TemplateName: "test-template",
			Partitions: []hpctypes.PartitionConfig{
				{Name: "normal", Nodes: 2, MaxRuntimeSeconds: 3600, State: "up"},
			},
		},
	}

	cluster, err := adapter.Deploy(ctx, config)
	if err != nil {
		t.Fatalf("failed to deploy: %v", err)
	}

	if !helmCalled {
		t.Error("expected Helm install to be called")
	}

	if cluster.Config.ClusterID != "test-cluster" {
		t.Errorf("expected cluster ID test-cluster, got %s", cluster.Config.ClusterID)
	}
}

func TestGetCluster(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	// Manually add a cluster
	adapter.clusters["test-id"] = &DeployedCluster{
		Config: DeploymentConfig{ClusterID: "test-id"},
		State:  ClusterStateRunning,
	}

	cluster, err := adapter.GetCluster("test-id")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}

	if cluster.Config.ClusterID != "test-id" {
		t.Errorf("expected cluster ID test-id, got %s", cluster.Config.ClusterID)
	}

	// Non-existent cluster
	_, err = adapter.GetCluster("non-existent")
	if err == nil {
		t.Error("expected error for non-existent cluster")
	}
}

func TestListClusters(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	adapter.clusters["cluster-1"] = &DeployedCluster{Config: DeploymentConfig{ClusterID: "cluster-1"}}
	adapter.clusters["cluster-2"] = &DeployedCluster{Config: DeploymentConfig{ClusterID: "cluster-2"}}

	clusters := adapter.ListClusters()
	if len(clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(clusters))
	}
}

func TestBuildHelmValues(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	const testPartitionNameGPU = "gpu"

	config := DeploymentConfig{
		ClusterID:        "test-cluster",
		ClusterName:      "Test Cluster",
		ProviderAddress:  "virtengine1abc123",
		ImageRegistry:    "ghcr.io/custom",
		StorageClass:     "fast-ssd",
		ProviderEndpoint: "https://provider.example.com:8443",
		Template: &hpctypes.ClusterTemplate{
			Partitions: []hpctypes.PartitionConfig{
				{Name: testPartitionNameGPU, Nodes: 4, MaxRuntimeSeconds: 86400, Priority: 100, State: "up"},
			},
			QoSPolicies: []hpctypes.QoSPolicy{
				{Name: "premium", Priority: 100, MaxJobsPerUser: 50},
			},
		},
	}

	values := adapter.buildHelmValues(config)

	// Check cluster section
	cluster, ok := values["cluster"].(map[string]interface{})
	if !ok {
		t.Fatal("expected cluster section in values")
	}
	if cluster["id"] != "test-cluster" {
		t.Errorf("expected cluster id test-cluster, got %v", cluster["id"])
	}

	// Check global section
	global, ok := values["global"].(map[string]interface{})
	if !ok {
		t.Fatal("expected global section in values")
	}
	if global["imageRegistry"] != "ghcr.io/custom" {
		t.Errorf("expected image registry ghcr.io/custom, got %v", global["imageRegistry"])
	}
	if global["storageClass"] != "fast-ssd" {
		t.Errorf("expected storage class fast-ssd, got %v", global["storageClass"])
	}

	// Check partitions
	partitions, ok := values["partitions"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected partitions in values")
	}
	if len(partitions) != 1 {
		t.Errorf("expected 1 partition, got %d", len(partitions))
	}
	if partitions[0]["name"] != testPartitionNameGPU {
		t.Errorf("expected partition name gpu, got %v", partitions[0]["name"])
	}

	// Check node agent
	nodeAgent, ok := values["nodeAgent"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nodeAgent in values")
	}
	if nodeAgent["enabled"] != true {
		t.Error("expected nodeAgent to be enabled")
	}
}

func TestParseSinfoOutput(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	output := `compute-0 64 256000 gpu:nvidia:8 idle
compute-1 64 256000 gpu:nvidia:8 mixed
compute-2 64 256000 gpu:nvidia:8 allocated
compute-3 64 256000 (null) down`

	capacity := adapter.parseSinfoOutput(output)

	if capacity.TotalNodes != 4 {
		t.Errorf("expected 4 total nodes, got %d", capacity.TotalNodes)
	}
	if capacity.AvailableNodes != 2 {
		t.Errorf("expected 2 available nodes, got %d", capacity.AvailableNodes)
	}
	if capacity.TotalCPUs != 256 {
		t.Errorf("expected 256 total CPUs, got %d", capacity.TotalCPUs)
	}
	if capacity.TotalGPUs != 24 {
		t.Errorf("expected 24 total GPUs, got %d", capacity.TotalGPUs)
	}
	if len(capacity.GPUTypes) != 1 || capacity.GPUTypes[0] != "nvidia" {
		t.Errorf("expected GPU type nvidia, got %v", capacity.GPUTypes)
	}
}

func TestParseGPUCount(t *testing.T) {
	tests := []struct {
		gres     string
		expected int
	}{
		{"gpu:nvidia:8", 8},
		{"gpu:4", 4},
		{"gpu:a100:4", 4},
		{"(null)", 0},
		{"", 0},
	}

	for _, test := range tests {
		result := parseGPUCount(test.gres)
		if result != test.expected {
			t.Errorf("parseGPUCount(%q) = %d, expected %d", test.gres, result, test.expected)
		}
	}
}

func TestParseGPUType(t *testing.T) {
	tests := []struct {
		gres     string
		expected string
	}{
		{"gpu:nvidia:8", "nvidia"},
		{"gpu:a100:4", "a100"},
		{"gpu:4", ""},
		{"(null)", ""},
	}

	for _, test := range tests {
		result := parseGPUType(test.gres)
		if result != test.expected {
			t.Errorf("parseGPUType(%q) = %q, expected %q", test.gres, result, test.expected)
		}
	}
}

func TestClusterStateTransitions(t *testing.T) {
	adapter := NewSLURMKubernetesAdapter(AdapterConfig{
		Helm: &MockHelmClient{},
		K8s:  &MockK8sChecker{},
	})

	// Add a test cluster
	adapter.clusters["test"] = &DeployedCluster{
		Config: DeploymentConfig{ClusterID: "test"},
		State:  ClusterStatePending,
	}

	// Test state update
	adapter.updateClusterState("test", ClusterStateDeploying, "Deploying")
	cluster, _ := adapter.GetCluster("test")
	if cluster.State != ClusterStateDeploying {
		t.Errorf("expected state deploying, got %s", cluster.State)
	}

	adapter.updateClusterState("test", ClusterStateRunning, "Running")
	cluster, _ = adapter.GetCluster("test")
	if cluster.State != ClusterStateRunning {
		t.Errorf("expected state running, got %s", cluster.State)
	}
}

