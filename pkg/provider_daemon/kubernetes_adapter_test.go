package provider_daemon

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockKubernetesClient is a mock implementation of KubernetesClient
type MockKubernetesClient struct {
	mu              sync.Mutex
	namespaces      map[string]bool
	deployments     map[string]*K8sDeploymentSpec
	services        map[string]*K8sServiceSpec
	secrets         map[string]map[string][]byte
	pvcs            map[string]*K8sPVCSpec
	networkPolicies map[string]*K8sNetworkPolicySpec
	podStatuses     map[string][]PodStatus
	endpoints       map[string][]WorkloadEndpoint
	failOnCreate    bool
}

func NewMockKubernetesClient() *MockKubernetesClient {
	return &MockKubernetesClient{
		namespaces:      make(map[string]bool),
		deployments:     make(map[string]*K8sDeploymentSpec),
		services:        make(map[string]*K8sServiceSpec),
		secrets:         make(map[string]map[string][]byte),
		pvcs:            make(map[string]*K8sPVCSpec),
		networkPolicies: make(map[string]*K8sNetworkPolicySpec),
		podStatuses:     make(map[string][]PodStatus),
		endpoints:       make(map[string][]WorkloadEndpoint),
	}
}

func (m *MockKubernetesClient) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.namespaces[name] = true
	return nil
}

func (m *MockKubernetesClient) DeleteNamespace(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.namespaces, name)
	return nil
}

func (m *MockKubernetesClient) CreateDeployment(ctx context.Context, namespace string, spec *K8sDeploymentSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.deployments[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *MockKubernetesClient) UpdateDeployment(ctx context.Context, namespace string, spec *K8sDeploymentSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := namespace + "/" + spec.Name
	if existing, ok := m.deployments[key]; ok {
		existing.Replicas = spec.Replicas
	}
	return nil
}

func (m *MockKubernetesClient) DeleteDeployment(ctx context.Context, namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.deployments, namespace+"/"+name)
	return nil
}

func (m *MockKubernetesClient) CreateService(ctx context.Context, namespace string, spec *K8sServiceSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.services[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *MockKubernetesClient) DeleteService(ctx context.Context, namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.services, namespace+"/"+name)
	return nil
}

func (m *MockKubernetesClient) CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.secrets[namespace+"/"+name] = data
	return nil
}

func (m *MockKubernetesClient) DeleteSecret(ctx context.Context, namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.secrets, namespace+"/"+name)
	return nil
}

func (m *MockKubernetesClient) CreatePVC(ctx context.Context, namespace string, spec *K8sPVCSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.pvcs[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *MockKubernetesClient) DeletePVC(ctx context.Context, namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pvcs, namespace+"/"+name)
	return nil
}

func (m *MockKubernetesClient) GetPodStatus(ctx context.Context, namespace, deploymentName string) ([]PodStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if statuses, ok := m.podStatuses[namespace+"/"+deploymentName]; ok {
		return statuses, nil
	}
	return []PodStatus{
		{Name: deploymentName + "-pod-1", Phase: "Running", Ready: true},
	}, nil
}

func (m *MockKubernetesClient) ApplyNetworkPolicy(ctx context.Context, namespace string, spec *K8sNetworkPolicySpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.networkPolicies[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *MockKubernetesClient) GetServiceEndpoints(ctx context.Context, namespace, serviceName string) ([]WorkloadEndpoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if endpoints, ok := m.endpoints[namespace+"/"+serviceName]; ok {
		return endpoints, nil
	}
	return []WorkloadEndpoint{
		{Service: serviceName, Port: 80, InternalAddress: serviceName + ".svc.cluster.local"},
	}, nil
}

func (m *MockKubernetesClient) SetFailOnCreate(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failOnCreate = fail
}

func (m *MockKubernetesClient) GetDeployment(namespace, name string) *K8sDeploymentSpec {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deployments[namespace+"/"+name]
}

func (m *MockKubernetesClient) HasNamespace(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.namespaces[name]
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     WorkloadState
		to       WorkloadState
		expected bool
	}{
		{"pending to deploying", WorkloadStatePending, WorkloadStateDeploying, true},
		{"pending to running", WorkloadStatePending, WorkloadStateRunning, false},
		{"deploying to running", WorkloadStateDeploying, WorkloadStateRunning, true},
		{"running to paused", WorkloadStateRunning, WorkloadStatePaused, true},
		{"running to stopping", WorkloadStateRunning, WorkloadStateStopping, true},
		{"paused to running", WorkloadStatePaused, WorkloadStateRunning, true},
		{"stopped to terminated", WorkloadStateStopped, WorkloadStateTerminated, true},
		{"terminated to anything", WorkloadStateTerminated, WorkloadStateRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKubernetesAdapterDeploy(t *testing.T) {
	client := NewMockKubernetesClient()
	statusChan := make(chan WorkloadStatusUpdate, 100)

	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       "provider-123",
		ResourcePrefix:   "test",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Tag:       "latest",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024 * 1024 * 1024},
				Ports: []PortSpec{
					{Name: "http", ContainerPort: 80, Protocol: "tcp"},
				},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{
		Timeout: 30 * time.Second,
	})

	require.NoError(t, err)
	require.NotNil(t, workload)
	assert.Equal(t, WorkloadStateRunning, workload.State)
	assert.NotEmpty(t, workload.ID)
	assert.Equal(t, "deployment-1", workload.DeploymentID)
	assert.Equal(t, "lease-1", workload.LeaseID)
	assert.True(t, client.HasNamespace(workload.Namespace))
}

func TestKubernetesAdapterDeployDryRun(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{
		DryRun: true,
	})

	require.NoError(t, err)
	require.NotNil(t, workload)
	assert.Equal(t, WorkloadStatePending, workload.State)
	assert.False(t, client.HasNamespace(workload.Namespace))
}

func TestKubernetesAdapterDeployInvalidManifest(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		// Missing version and services
		Name: "test-app",
	}

	ctx := context.Background()
	_, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid manifest")
}

func TestKubernetesAdapterDeployFailure(t *testing.T) {
	client := NewMockKubernetesClient()
	client.SetFailOnCreate(true)

	statusChan := make(chan WorkloadStatusUpdate, 100)
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       "provider-123",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})

	require.Error(t, err)
	require.NotNil(t, workload)
	assert.Equal(t, WorkloadStateFailed, workload.State)
}

func TestKubernetesAdapterDeployWithVolumes(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
				Volumes: []VolumeMountSpec{
					{Name: "data", MountPath: "/data"},
				},
			},
		},
		Volumes: []VolumeSpec{
			{Name: "data", Type: "persistent", Size: 10 * 1024 * 1024 * 1024},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})

	require.NoError(t, err)
	assert.Equal(t, WorkloadStateRunning, workload.State)

	// Check PVC was created
	hasPVC := false
	for _, res := range workload.Resources {
		if res.Kind == "PersistentVolumeClaim" && res.Name == "data" {
			hasPVC = true
			break
		}
	}
	assert.True(t, hasPVC, "Expected PVC to be created")
}

func TestKubernetesAdapterDeployWithSecrets(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{
		Secrets: []SecretData{
			{
				Name: "app-secrets",
				Data: map[string][]byte{
					"api-key": []byte("secret-value"),
				},
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, WorkloadStateRunning, workload.State)

	// Check secret was created
	hasSecret := false
	for _, res := range workload.Resources {
		if res.Kind == "Secret" && res.Name == "app-secrets" {
			hasSecret = true
			break
		}
	}
	assert.True(t, hasSecret, "Expected secret to be created")
}

func TestKubernetesAdapterGetWorkload(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	deployed, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)

	// Get by ID
	workload, err := adapter.GetWorkload(deployed.ID)
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, workload.ID)

	// Get non-existent
	_, err = adapter.GetWorkload("non-existent")
	require.Error(t, err)
	assert.Equal(t, ErrWorkloadNotFound, err)
}

func TestKubernetesAdapterGetWorkloadByLease(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	deployed, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)

	workload, err := adapter.GetWorkloadByLease("lease-1")
	require.NoError(t, err)
	assert.Equal(t, deployed.ID, workload.ID)
}

func TestKubernetesAdapterListWorkloads(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	_, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)
	_, err = adapter.Deploy(ctx, manifest, "deployment-2", "lease-2", DeploymentOptions{})
	require.NoError(t, err)

	workloads := adapter.ListWorkloads()
	assert.Len(t, workloads, 2)
}

func TestKubernetesAdapterTerminate(t *testing.T) {
	client := NewMockKubernetesClient()
	statusChan := make(chan WorkloadStatusUpdate, 100)
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       "provider-123",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)
	require.True(t, client.HasNamespace(workload.Namespace))

	err = adapter.Terminate(ctx, workload.ID)
	require.NoError(t, err)

	// Namespace should be deleted
	assert.False(t, client.HasNamespace(workload.Namespace))

	// Workload should be terminated
	w, _ := adapter.GetWorkload(workload.ID)
	assert.Equal(t, WorkloadStateTerminated, w.State)
}

func TestKubernetesAdapterPauseResume(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
				Replicas:  3,
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)

	// Pause
	err = adapter.Pause(ctx, workload.ID)
	require.NoError(t, err)

	w, _ := adapter.GetWorkload(workload.ID)
	assert.Equal(t, WorkloadStatePaused, w.State)

	// Check deployment was scaled to 0
	dep := client.GetDeployment(workload.Namespace, "web")
	require.NotNil(t, dep)
	assert.Equal(t, int32(0), dep.Replicas)

	// Resume
	err = adapter.Resume(ctx, workload.ID)
	require.NoError(t, err)

	w, _ = adapter.GetWorkload(workload.ID)
	assert.Equal(t, WorkloadStateRunning, w.State)

	// Check deployment was scaled back
	dep = client.GetDeployment(workload.Namespace, "web")
	assert.Equal(t, int32(3), dep.Replicas)
}

func TestKubernetesAdapterGetStatus(t *testing.T) {
	client := NewMockKubernetesClient()
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:     client,
		ProviderID: "provider-123",
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)

	status, err := adapter.GetStatus(ctx, workload.ID)
	require.NoError(t, err)
	assert.Equal(t, workload.ID, status.WorkloadID)
	assert.Equal(t, WorkloadStateRunning, status.State)
}

func TestStatusUpdateChannel(t *testing.T) {
	client := NewMockKubernetesClient()
	statusChan := make(chan WorkloadStatusUpdate, 100)

	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       "provider-123",
		StatusUpdateChan: statusChan,
	})

	manifest := &Manifest{
		Version: ManifestVersionV1,
		Name:    "test-app",
		Services: []ServiceSpec{
			{
				Name:      "web",
				Type:      "container",
				Image:     "nginx",
				Resources: ResourceSpec{CPU: 1000, Memory: 1024},
			},
		},
	}

	ctx := context.Background()
	_, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", DeploymentOptions{})
	require.NoError(t, err)

	// Should have received multiple status updates
	updateCount := 0
	done := false
	for !done {
		select {
		case <-statusChan:
			updateCount++
		default:
			done = true
		}
	}

	assert.Greater(t, updateCount, 0, "Expected status updates to be sent")
}

func TestBuildDeploymentSpec(t *testing.T) {
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		ProviderID: "provider-123",
	})

	svc := &ServiceSpec{
		Name:    "web",
		Type:    "container",
		Image:   "nginx",
		Tag:     "1.21",
		Command: []string{"nginx"},
		Args:    []string{"-g", "daemon off;"},
		Env:     map[string]string{"ENV": "prod"},
		Resources: ResourceSpec{
			CPU:    2000,
			Memory: 4 * 1024 * 1024 * 1024,
			GPU:    1,
		},
		Ports: []PortSpec{
			{Name: "http", ContainerPort: 80, Protocol: "tcp"},
		},
		Replicas: 3,
		HealthCheck: &HealthCheckSpec{
			HTTP: &HTTPAction{
				Path: "/health",
				Port: 80,
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
		},
	}

	workload := &DeployedWorkload{
		DeploymentID: "deployment-1",
		LeaseID:      "lease-1",
		Manifest:     &Manifest{},
	}

	spec := adapter.buildDeploymentSpec(svc, workload, DeploymentOptions{})

	assert.Equal(t, "web", spec.Name)
	assert.Equal(t, int32(3), spec.Replicas)
	assert.Len(t, spec.Containers, 1)

	container := spec.Containers[0]
	assert.Equal(t, "nginx:1.21", container.Image)
	assert.Equal(t, "2000m", container.Resources.CPURequest)
	assert.Equal(t, "1", container.Resources.GPULimit)
	assert.NotNil(t, container.LivenessProbe)
	assert.NotNil(t, container.ReadinessProbe)
}

func TestBuildServiceSpec(t *testing.T) {
	adapter := NewKubernetesAdapter(KubernetesAdapterConfig{
		ProviderID: "provider-123",
	})

	svc := &ServiceSpec{
		Name: "web",
		Ports: []PortSpec{
			{Name: "http", ContainerPort: 80, Expose: true, ExternalPort: 8080},
			{Name: "metrics", ContainerPort: 9090, Expose: false},
		},
	}

	workload := &DeployedWorkload{}

	spec := adapter.buildServiceSpec(svc, workload, DeploymentOptions{})

	assert.Equal(t, "web", spec.Name)
	assert.Equal(t, "LoadBalancer", spec.Type) // Because one port is exposed
	assert.Len(t, spec.Ports, 2)

	// First port should use external port
	assert.Equal(t, int32(8080), spec.Ports[0].Port)
	assert.Equal(t, int32(80), spec.Ports[0].TargetPort)
}

