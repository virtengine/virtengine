// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
package slurm_k8s

import "context"

// MockHelmClient is a mock Helm client for testing.
// Exported for use in integration tests.
type MockHelmClient struct {
	InstallFunc   func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error
	UpgradeFunc   func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error
	UninstallFunc func(ctx context.Context, releaseName, namespace string) error
	GetFunc       func(ctx context.Context, releaseName, namespace string) (*HelmRelease, error)
	ListFunc      func(ctx context.Context, namespace string) ([]*HelmRelease, error)
}

// Install implements HelmClient.Install.
func (m *MockHelmClient) Install(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(ctx, releaseName, chartPath, namespace, values)
	}
	return nil
}

// Upgrade implements HelmClient.Upgrade.
func (m *MockHelmClient) Upgrade(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
	if m.UpgradeFunc != nil {
		return m.UpgradeFunc(ctx, releaseName, chartPath, namespace, values)
	}
	return nil
}

// Uninstall implements HelmClient.Uninstall.
func (m *MockHelmClient) Uninstall(ctx context.Context, releaseName, namespace string) error {
	if m.UninstallFunc != nil {
		return m.UninstallFunc(ctx, releaseName, namespace)
	}
	return nil
}

// GetRelease implements HelmClient.GetRelease.
func (m *MockHelmClient) GetRelease(ctx context.Context, releaseName, namespace string) (*HelmRelease, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, releaseName, namespace)
	}
	return &HelmRelease{
		Name:      releaseName,
		Namespace: namespace,
		Status:    "deployed",
	}, nil
}

// ListReleases implements HelmClient.ListReleases.
func (m *MockHelmClient) ListReleases(ctx context.Context, namespace string) ([]*HelmRelease, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, namespace)
	}
	return nil, nil
}

// MockK8sChecker is a mock Kubernetes status checker.
// Exported for use in integration tests.
type MockK8sChecker struct {
	StatefulSetStatus map[string]*StatefulSetStatus
	ExecOutput        map[string]string
}

// GetStatefulSetStatus implements K8sChecker.GetStatefulSetStatus.
func (m *MockK8sChecker) GetStatefulSetStatus(ctx context.Context, namespace, name string) (*StatefulSetStatus, error) {
	key := namespace + "/" + name
	if status, ok := m.StatefulSetStatus[key]; ok {
		return status, nil
	}
	return &StatefulSetStatus{
		Name:          name,
		Replicas:      1,
		ReadyReplicas: 1,
	}, nil
}

// GetPodLogs implements K8sChecker.GetPodLogs.
func (m *MockK8sChecker) GetPodLogs(ctx context.Context, namespace, podName, containerName string, lines int) (string, error) {
	return "", nil
}

// ExecInPod implements K8sChecker.ExecInPod.
func (m *MockK8sChecker) ExecInPod(ctx context.Context, namespace, podName, containerName string, command []string) (string, error) {
	if m.ExecOutput != nil {
		key := podName + ":" + command[0]
		if output, ok := m.ExecOutput[key]; ok {
			return output, nil
		}
	}
	return "OK", nil
}

// MockReporter is a mock on-chain reporter.
// Exported for use in integration tests.
type MockReporter struct {
	StatusReports   []ClusterStatusUpdate
	CapacityReports []ClusterCapacity
	NodeJoins       []string
	NodeLeaves      []string
}

// ReportClusterStatus implements Reporter.ReportClusterStatus.
func (m *MockReporter) ReportClusterStatus(ctx context.Context, clusterID string, status *ClusterStatusUpdate) error {
	m.StatusReports = append(m.StatusReports, *status)
	return nil
}

// ReportCapacityUpdate implements Reporter.ReportCapacityUpdate.
func (m *MockReporter) ReportCapacityUpdate(ctx context.Context, clusterID string, capacity *ClusterCapacity) error {
	m.CapacityReports = append(m.CapacityReports, *capacity)
	return nil
}

// ReportNodeJoin implements Reporter.ReportNodeJoin.
func (m *MockReporter) ReportNodeJoin(ctx context.Context, clusterID, nodeID string) error {
	m.NodeJoins = append(m.NodeJoins, nodeID)
	return nil
}

// ReportNodeLeave implements Reporter.ReportNodeLeave.
func (m *MockReporter) ReportNodeLeave(ctx context.Context, clusterID, nodeID string) error {
	m.NodeLeaves = append(m.NodeLeaves, nodeID)
	return nil
}
