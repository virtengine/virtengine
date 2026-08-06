package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/remotecommand"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

type fakeExecCall struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
}

type fakeKubernetesRuntimeClient struct {
	mu            sync.Mutex
	namespaces    map[string]bool
	deployments   map[string]*provider_daemon.K8sDeploymentSpec
	services      map[string]*provider_daemon.K8sServiceSpec
	podStatuses   map[string][]provider_daemon.PodStatus
	logs          map[string]string
	execCalls     []fakeExecCall
	failOnCreate  bool
	failExec      error
	failPodLookup error
	failLogStream error
}

func newFakeKubernetesRuntimeClient() *fakeKubernetesRuntimeClient {
	return &fakeKubernetesRuntimeClient{
		namespaces:  make(map[string]bool),
		deployments: make(map[string]*provider_daemon.K8sDeploymentSpec),
		services:    make(map[string]*provider_daemon.K8sServiceSpec),
		podStatuses: make(map[string][]provider_daemon.PodStatus),
		logs:        make(map[string]string),
	}
}

func (f *fakeKubernetesRuntimeClient) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failOnCreate {
		return fmt.Errorf("create namespace failed")
	}
	f.namespaces[name] = true
	return nil
}

func (f *fakeKubernetesRuntimeClient) DeleteNamespace(ctx context.Context, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.namespaces, name)
	return nil
}

func (f *fakeKubernetesRuntimeClient) CreateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failOnCreate {
		return fmt.Errorf("create deployment failed")
	}
	f.deployments[namespace+"/"+spec.Name] = spec
	if _, ok := f.podStatuses[namespace+"/"+spec.Name]; !ok {
		f.podStatuses[namespace+"/"+spec.Name] = []provider_daemon.PodStatus{{
			Name:  spec.Name + "-pod-0",
			Phase: "Running",
			Ready: true,
		}}
	}
	return nil
}

func (f *fakeKubernetesRuntimeClient) UpdateDeployment(ctx context.Context, namespace string, spec *provider_daemon.K8sDeploymentSpec) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := namespace + "/" + spec.Name
	if current, ok := f.deployments[key]; ok {
		current.Replicas = spec.Replicas
	}
	return nil
}

func (f *fakeKubernetesRuntimeClient) DeleteDeployment(ctx context.Context, namespace, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.deployments, namespace+"/"+name)
	return nil
}

func (f *fakeKubernetesRuntimeClient) CreateService(ctx context.Context, namespace string, spec *provider_daemon.K8sServiceSpec) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.services[namespace+"/"+spec.Name] = spec
	return nil
}

func (f *fakeKubernetesRuntimeClient) DeleteService(ctx context.Context, namespace, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.services, namespace+"/"+name)
	return nil
}

func (f *fakeKubernetesRuntimeClient) CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	return nil
}

func (f *fakeKubernetesRuntimeClient) DeleteSecret(ctx context.Context, namespace, name string) error {
	return nil
}

func (f *fakeKubernetesRuntimeClient) CreatePVC(ctx context.Context, namespace string, spec *provider_daemon.K8sPVCSpec) error {
	return nil
}

func (f *fakeKubernetesRuntimeClient) DeletePVC(ctx context.Context, namespace, name string) error {
	return nil
}

func (f *fakeKubernetesRuntimeClient) GetPodStatus(ctx context.Context, namespace, deploymentName string) ([]provider_daemon.PodStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failPodLookup != nil {
		return nil, f.failPodLookup
	}
	if statuses, ok := f.podStatuses[namespace+"/"+deploymentName]; ok {
		return append([]provider_daemon.PodStatus(nil), statuses...), nil
	}
	return []provider_daemon.PodStatus{{
		Name:  deploymentName + "-pod-0",
		Phase: "Running",
		Ready: true,
	}}, nil
}

func (f *fakeKubernetesRuntimeClient) ApplyNetworkPolicy(ctx context.Context, namespace string, spec *provider_daemon.K8sNetworkPolicySpec) error {
	return nil
}

func (f *fakeKubernetesRuntimeClient) GetServiceEndpoints(ctx context.Context, namespace, serviceName string) ([]provider_daemon.WorkloadEndpoint, error) {
	return []provider_daemon.WorkloadEndpoint{{
		Service:         serviceName,
		Port:            80,
		Protocol:        "TCP",
		InternalAddress: serviceName + ".svc.cluster.local",
	}}, nil
}

func (f *fakeKubernetesRuntimeClient) ResolveDeploymentPods(ctx context.Context, namespace, deploymentName string) ([]deploymentPod, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failPodLookup != nil {
		return nil, f.failPodLookup
	}
	statuses := f.podStatuses[namespace+"/"+deploymentName]
	if len(statuses) == 0 {
		return []deploymentPod{{
			Name:       deploymentName + "-pod-0",
			Phase:      "Running",
			Ready:      true,
			Containers: []string{deploymentName},
			CreatedAt:  time.Now().UTC(),
		}}, nil
	}

	pods := make([]deploymentPod, 0, len(statuses))
	for _, status := range statuses {
		pods = append(pods, deploymentPod{
			Name:       status.Name,
			Phase:      status.Phase,
			Ready:      status.Ready,
			Containers: []string{deploymentName},
			CreatedAt:  time.Now().UTC(),
		})
	}
	return pods, nil
}

func (f *fakeKubernetesRuntimeClient) StreamPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64, follow bool) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failLogStream != nil {
		return nil, f.failLogStream
	}
	return io.NopCloser(strings.NewReader(f.logs[namespace+"/"+podName+"/"+containerName])), nil
}

func (f *fakeKubernetesRuntimeClient) ExecInPod(
	ctx context.Context,
	namespace, podName, containerName string,
	command []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	tty bool,
	terminalResize <-chan remotecommand.TerminalSize,
) error {
	f.mu.Lock()
	f.execCalls = append(f.execCalls, fakeExecCall{
		Namespace: namespace,
		Pod:       podName,
		Container: containerName,
		Command:   append([]string(nil), command...),
	})
	failErr := f.failExec
	f.mu.Unlock()

	if failErr != nil {
		return failErr
	}
	if stdout != nil {
		_, _ = stdout.Write([]byte("shell ready\n"))
	}
	return nil
}

func (f *fakeKubernetesRuntimeClient) SetPodReady(namespace, deploymentName string, ready bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.podStatuses[namespace+"/"+deploymentName] = []provider_daemon.PodStatus{{
		Name:  deploymentName + "-pod-0",
		Phase: "Running",
		Ready: ready,
	}}
}

func testContainerManifest() *provider_daemon.Manifest {
	return &provider_daemon.Manifest{
		Version: provider_daemon.ManifestVersionV1,
		Name:    "allocation-alloc-1",
		Services: []provider_daemon.ServiceSpec{{
			Name:      "web",
			Type:      "container",
			Image:     "nginx",
			Tag:       "latest",
			Resources: provider_daemon.ResourceSpec{CPU: 1000, Memory: 256 * 1024 * 1024},
			Ports: []provider_daemon.PortSpec{{
				Name:          "http",
				ContainerPort: 80,
				Expose:        true,
			}},
		}},
	}
}

func TestNewKubernetesWorkloadRuntimeDryRun(t *testing.T) {
	runtime, err := newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
		ProviderID:     "provider-1",
		ResourcePrefix: "ve",
		DryRun:         true,
	})
	require.NoError(t, err)
	require.NotNil(t, runtime)
	require.True(t, runtime.dryRun)
	require.NotNil(t, runtime.adapter)
	require.Nil(t, runtime.client)
}

func TestKubernetesWorkloadRuntimeTailLogsAndShell(t *testing.T) {
	fakeClient := newFakeKubernetesRuntimeClient()
	runtime, err := newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
		ProviderID:     "provider-1",
		ResourcePrefix: "ve",
		NewClient: func(kubeconfig string) (kubernetesRuntimeClient, error) {
			return fakeClient, nil
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	workload, err := runtime.adapter.Deploy(ctx, testContainerManifest(), "alloc-1", "alloc-1", provider_daemon.DeploymentOptions{})
	require.NoError(t, err)

	fakeClient.logs[workload.Namespace+"/web-pod-0/web"] = strings.Join([]string{
		"2026-04-10T11:00:00Z starting",
		"2026-04-10T11:00:01Z error failed-to-bind",
	}, "\n")

	entries, err := runtime.TailLogs(ctx, "alloc-1", "web", 20)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.Equal(t, "info", entries[0].Level)
	require.Equal(t, "error", entries[1].Level)

	stdout := &bytes.Buffer{}
	err = runtime.OpenShell(ctx, &provider_daemon.ShellExecutionRequest{
		DeploymentID: "alloc-1",
		Container:    "web",
		Stdout:       stdout,
	})
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "shell ready")
	require.Len(t, fakeClient.execCalls, 1)
	require.Equal(t, []string{"/bin/sh"}, fakeClient.execCalls[0].Command)
}
