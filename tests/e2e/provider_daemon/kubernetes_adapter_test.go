//go:build e2e.integration

package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
)

type mockKubernetesClient struct {
	namespaces      map[string]bool
	deployments     map[string]*pd.K8sDeploymentSpec
	services        map[string]*pd.K8sServiceSpec
	secrets         map[string]map[string][]byte
	pvcs            map[string]*pd.K8sPVCSpec
	networkPolicies map[string]*pd.K8sNetworkPolicySpec
	endpoints       map[string][]pd.WorkloadEndpoint
	podStatuses     map[string][]pd.PodStatus
	failOnCreate    bool
}

func newMockKubernetesClient() *mockKubernetesClient {
	return &mockKubernetesClient{
		namespaces:      make(map[string]bool),
		deployments:     make(map[string]*pd.K8sDeploymentSpec),
		services:        make(map[string]*pd.K8sServiceSpec),
		secrets:         make(map[string]map[string][]byte),
		pvcs:            make(map[string]*pd.K8sPVCSpec),
		networkPolicies: make(map[string]*pd.K8sNetworkPolicySpec),
		endpoints:       make(map[string][]pd.WorkloadEndpoint),
		podStatuses:     make(map[string][]pd.PodStatus),
	}
}

func (m *mockKubernetesClient) CreateNamespace(_ context.Context, name string, _ map[string]string) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.namespaces[name] = true
	return nil
}

func (m *mockKubernetesClient) DeleteNamespace(_ context.Context, name string) error {
	delete(m.namespaces, name)
	return nil
}

func (m *mockKubernetesClient) CreateDeployment(_ context.Context, namespace string, spec *pd.K8sDeploymentSpec) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.deployments[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *mockKubernetesClient) UpdateDeployment(_ context.Context, namespace string, spec *pd.K8sDeploymentSpec) error {
	key := namespace + "/" + spec.Name
	if existing, ok := m.deployments[key]; ok {
		existing.Replicas = spec.Replicas
	}
	return nil
}

func (m *mockKubernetesClient) DeleteDeployment(_ context.Context, namespace, name string) error {
	delete(m.deployments, namespace+"/"+name)
	return nil
}

func (m *mockKubernetesClient) CreateService(_ context.Context, namespace string, spec *pd.K8sServiceSpec) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.services[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *mockKubernetesClient) DeleteService(_ context.Context, namespace, name string) error {
	delete(m.services, namespace+"/"+name)
	return nil
}

func (m *mockKubernetesClient) CreateSecret(_ context.Context, namespace, name string, data map[string][]byte) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.secrets[namespace+"/"+name] = data
	return nil
}

func (m *mockKubernetesClient) DeleteSecret(_ context.Context, namespace, name string) error {
	delete(m.secrets, namespace+"/"+name)
	return nil
}

func (m *mockKubernetesClient) CreatePVC(_ context.Context, namespace string, spec *pd.K8sPVCSpec) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.pvcs[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *mockKubernetesClient) DeletePVC(_ context.Context, namespace, name string) error {
	delete(m.pvcs, namespace+"/"+name)
	return nil
}

func (m *mockKubernetesClient) GetPodStatus(_ context.Context, namespace, deploymentName string) ([]pd.PodStatus, error) {
	key := namespace + "/" + deploymentName
	if status, ok := m.podStatuses[key]; ok {
		return status, nil
	}
	return []pd.PodStatus{{Name: "pod", Phase: "Running", Ready: true, StartTime: time.Now()}}, nil
}

func (m *mockKubernetesClient) ApplyNetworkPolicy(_ context.Context, namespace string, spec *pd.K8sNetworkPolicySpec) error {
	if m.failOnCreate {
		return errors.New("mock failure")
	}
	m.networkPolicies[namespace+"/"+spec.Name] = spec
	return nil
}

func (m *mockKubernetesClient) GetServiceEndpoints(_ context.Context, namespace, serviceName string) ([]pd.WorkloadEndpoint, error) {
	key := namespace + "/" + serviceName
	if endpoints, ok := m.endpoints[key]; ok {
		return endpoints, nil
	}
	return []pd.WorkloadEndpoint{{Service: serviceName, Port: 8080, Protocol: "TCP", InternalAddress: "10.0.0.1"}}, nil
}

func TestKubernetesAdapterE2E(t *testing.T) {
	ctx := context.Background()
	h := newWaldurHarness(t)

	manifest := &pd.Manifest{
		Version: pd.ManifestVersionV1,
		Name:    "k8s-e2e",
		Services: []pd.ServiceSpec{
			{
				Name:  "api",
				Type:  "container",
				Image: "nginx",
				Tag:   "latest",
				Resources: pd.ResourceSpec{
					CPU:    500,
					Memory: 512 * 1024 * 1024,
					GPU:    1,
				},
				Ports:   []pd.PortSpec{{Name: "http", ContainerPort: 8080, Expose: true}},
				Volumes: []pd.VolumeMountSpec{{Name: "data", MountPath: "/data"}},
			},
		},
		Volumes: []pd.VolumeSpec{{Name: "data", Type: "persistent", Size: 10 * 1024 * 1024 * 1024}},
	}

	order := h.createOrder(ctx, "k8s-order", map[string]interface{}{"backend": "k8s"})
	resource := h.waitForResource(order.UUID)

	statusCh := make(chan pd.WorkloadStatusUpdate, 10)
	client := newMockKubernetesClient()
	adapter := pd.NewKubernetesAdapter(pd.KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       "provider-e2e",
		ResourcePrefix:   "e2e",
		StatusUpdateChan: statusCh,
	})

	workload, err := adapter.Deploy(ctx, manifest, "deployment-1", "lease-1", pd.DeploymentOptions{})
	require.NoError(t, err)
	require.Equal(t, pd.WorkloadStateRunning, workload.State)

	err = adapter.Pause(ctx, workload.ID)
	require.NoError(t, err)
	err = adapter.Resume(ctx, workload.ID)
	require.NoError(t, err)

	h.submitUsage(ctx, resource.UUID, workload.ID)
	require.Greater(t, len(h.mock.GetUsageRecords(resource.UUID)), 0)

	_, err = h.lifecycle.Stop(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	assert.Equal(t, "Stopped", h.mock.GetResource(resource.UUID).State)

	_, err = h.lifecycle.Start(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	assert.Equal(t, "OK", h.mock.GetResource(resource.UUID).State)

	err = adapter.Terminate(ctx, workload.ID)
	require.NoError(t, err)
	_, err = h.lifecycle.Terminate(ctx, waldur.LifecycleRequest{ResourceUUID: resource.UUID})
	require.NoError(t, err)
	assert.Equal(t, "Terminated", h.mock.GetResource(resource.UUID).State)

	t.Run("DeploymentFailure", func(t *testing.T) {
		failingClient := newMockKubernetesClient()
		failingClient.failOnCreate = true
		failAdapter := pd.NewKubernetesAdapter(pd.KubernetesAdapterConfig{Client: failingClient, ProviderID: "provider-e2e"})

		_, err := failAdapter.Deploy(ctx, manifest, "deployment-fail", "lease-fail", pd.DeploymentOptions{})
		require.Error(t, err)
	})
}
