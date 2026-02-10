package provider_daemon

import "context"

// NoopKubernetesClient is a no-op implementation for dry-run provisioning.
type NoopKubernetesClient struct{}

// NewNoopKubernetesClient creates a new no-op Kubernetes client.
func NewNoopKubernetesClient() *NoopKubernetesClient {
	return &NoopKubernetesClient{}
}

func (n *NoopKubernetesClient) CreateNamespace(_ context.Context, _ string, _ map[string]string) error {
	return nil
}

func (n *NoopKubernetesClient) DeleteNamespace(_ context.Context, _ string) error {
	return nil
}

func (n *NoopKubernetesClient) CreateDeployment(_ context.Context, _ string, _ *K8sDeploymentSpec) error {
	return nil
}

func (n *NoopKubernetesClient) UpdateDeployment(_ context.Context, _ string, _ *K8sDeploymentSpec) error {
	return nil
}

func (n *NoopKubernetesClient) DeleteDeployment(_ context.Context, _ string, _ string) error {
	return nil
}

func (n *NoopKubernetesClient) CreateService(_ context.Context, _ string, _ *K8sServiceSpec) error {
	return nil
}

func (n *NoopKubernetesClient) DeleteService(_ context.Context, _ string, _ string) error {
	return nil
}

func (n *NoopKubernetesClient) CreateSecret(_ context.Context, _ string, _ string, _ map[string][]byte) error {
	return nil
}

func (n *NoopKubernetesClient) DeleteSecret(_ context.Context, _ string, _ string) error {
	return nil
}

func (n *NoopKubernetesClient) CreatePVC(_ context.Context, _ string, _ *K8sPVCSpec) error {
	return nil
}

func (n *NoopKubernetesClient) DeletePVC(_ context.Context, _ string, _ string) error {
	return nil
}

func (n *NoopKubernetesClient) GetPodStatus(_ context.Context, _ string, _ string) ([]PodStatus, error) {
	return []PodStatus{}, nil
}

func (n *NoopKubernetesClient) ApplyNetworkPolicy(_ context.Context, _ string, _ *K8sNetworkPolicySpec) error {
	return nil
}

func (n *NoopKubernetesClient) GetServiceEndpoints(_ context.Context, _ string, _ string) ([]WorkloadEndpoint, error) {
	return []WorkloadEndpoint{}, nil
}
