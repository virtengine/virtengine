package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"k8s.io/client-go/tools/remotecommand"
)

const defaultWorkloadReconcileInterval = 15 * time.Second

type deploymentPod struct {
	Name       string
	Phase      string
	Ready      bool
	Containers []string
	CreatedAt  time.Time
}

type kubernetesRuntimeClient interface {
	provider_daemon.KubernetesClient
	ResolveDeploymentPods(ctx context.Context, namespace, deploymentName string) ([]deploymentPod, error)
	StreamPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64, follow bool) (io.ReadCloser, error)
	ExecInPod(
		ctx context.Context,
		namespace, podName, containerName string,
		command []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		tty bool,
		terminalResize <-chan remotecommand.TerminalSize,
	) error
}

type kubernetesRuntimeConfig struct {
	ProviderID        string
	ResourcePrefix    string
	Kubeconfig        string
	DryRun            bool
	ReconcileInterval time.Duration
	StatusUpdateChan  chan<- provider_daemon.WorkloadStatusUpdate
	NewClient         func(kubeconfig string) (kubernetesRuntimeClient, error)
}

type kubernetesWorkloadRuntime struct {
	adapter           *provider_daemon.KubernetesAdapter
	client            kubernetesRuntimeClient
	dryRun            bool
	reconcileInterval time.Duration
}

// CollectMetrics implements the metering boundary. The core Kubernetes client
// currently exposes lifecycle/status APIs but not metrics-server counters, so
// production collection fails closed instead of inventing or billing requested
// resources. A metrics-server adapter can replace this method without changing
// the authenticated submission pipeline.
func (r *kubernetesWorkloadRuntime) CollectMetrics(_ context.Context, workloadID string) (*provider_daemon.ResourceMetrics, error) {
	return nil, fmt.Errorf("workload %s metrics are unavailable: kubernetes metrics-server collector is not configured", workloadID)
}

type workloadTarget struct {
	Workload      *provider_daemon.DeployedWorkload
	Namespace     string
	Deployment    string
	PodName       string
	ContainerName string
}

func newKubernetesWorkloadRuntime(cfg kubernetesRuntimeConfig) (*kubernetesWorkloadRuntime, error) {
	interval := cfg.ReconcileInterval
	if interval <= 0 {
		interval = defaultWorkloadReconcileInterval
	}

	runtime := &kubernetesWorkloadRuntime{
		dryRun:            cfg.DryRun,
		reconcileInterval: interval,
	}

	if cfg.DryRun {
		runtime.adapter = provider_daemon.NewKubernetesAdapter(provider_daemon.KubernetesAdapterConfig{
			Client:           provider_daemon.NewNoopKubernetesClient(),
			ProviderID:       cfg.ProviderID,
			ResourcePrefix:   cfg.ResourcePrefix,
			StatusUpdateChan: cfg.StatusUpdateChan,
		})
		return runtime, nil
	}

	clientFactory := cfg.NewClient
	if clientFactory == nil {
		clientFactory = func(kubeconfig string) (kubernetesRuntimeClient, error) {
			return newKubernetesAPIClient(kubeconfig)
		}
	}

	client, err := clientFactory(cfg.Kubeconfig)
	if err != nil {
		return nil, err
	}

	runtime.client = client
	runtime.adapter = provider_daemon.NewKubernetesAdapter(provider_daemon.KubernetesAdapterConfig{
		Client:           client,
		ProviderID:       cfg.ProviderID,
		ResourcePrefix:   cfg.ResourcePrefix,
		StatusUpdateChan: cfg.StatusUpdateChan,
	})

	return runtime, nil
}

func (r *kubernetesWorkloadRuntime) Start(ctx context.Context) {
	if r == nil || r.client == nil || r.dryRun {
		return
	}

	ticker := time.NewTicker(r.reconcileInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.reconcileAll(ctx)
			}
		}
	}()
}

func (r *kubernetesWorkloadRuntime) reconcileAll(ctx context.Context) {
	if r == nil || r.adapter == nil {
		return
	}

	for _, workload := range r.adapter.ListWorkloads() {
		if workload == nil {
			continue
		}
		_, _ = r.adapter.GetStatus(ctx, workload.ID)
	}
}

func (r *kubernetesWorkloadRuntime) TailLogs(
	ctx context.Context,
	deploymentID, containerName string,
	tail int,
) ([]provider_daemon.LogEntry, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("kubernetes workload runtime unavailable")
	}

	target, err := r.resolvePodTarget(ctx, deploymentID, containerName)
	if err != nil {
		return nil, err
	}

	stream, err := r.client.StreamPodLogs(ctx, target.Namespace, target.PodName, target.ContainerName, int64(tail), false)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return readKubernetesLogEntries(stream)
}

func (r *kubernetesWorkloadRuntime) StreamLogs(
	ctx context.Context,
	deploymentID, containerName string,
	tail int,
) (<-chan provider_daemon.LogEntry, func(), error) {
	if r == nil || r.client == nil {
		return nil, nil, fmt.Errorf("kubernetes workload runtime unavailable")
	}

	target, err := r.resolvePodTarget(ctx, deploymentID, containerName)
	if err != nil {
		return nil, nil, err
	}

	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := r.client.StreamPodLogs(streamCtx, target.Namespace, target.PodName, target.ContainerName, int64(tail), true)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	ch := make(chan provider_daemon.LogEntry, 32)
	go func() {
		defer close(ch)
		defer cancel()
		defer stream.Close()

		scanner := bufio.NewScanner(stream)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			select {
			case <-streamCtx.Done():
				return
			case ch <- parseKubernetesLogLine(scanner.Text()):
			}
		}
	}()

	return ch, cancel, nil
}

func (r *kubernetesWorkloadRuntime) OpenShell(ctx context.Context, req *provider_daemon.ShellExecutionRequest) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("kubernetes workload runtime unavailable")
	}
	if req == nil {
		return fmt.Errorf("shell execution request is required")
	}

	target, err := r.resolvePodTarget(ctx, req.DeploymentID, req.Container)
	if err != nil {
		return err
	}

	return r.client.ExecInPod(
		ctx,
		target.Namespace,
		target.PodName,
		target.ContainerName,
		[]string{"/bin/sh"},
		req.Stdin,
		req.Stdout,
		req.Stderr,
		req.TTY,
		req.TerminalResize,
	)
}

func (r *kubernetesWorkloadRuntime) resolvePodTarget(
	ctx context.Context,
	deploymentID, containerHint string,
) (*workloadTarget, error) {
	if r == nil || r.adapter == nil {
		return nil, fmt.Errorf("kubernetes adapter unavailable")
	}

	workload, err := r.adapter.GetWorkloadByDeployment(deploymentID)
	if err != nil {
		workload, err = r.adapter.GetWorkloadByLease(deploymentID)
		if err != nil {
			return nil, fmt.Errorf("unknown workload %s: %w", deploymentID, err)
		}
	}
	if workload.Manifest == nil || len(workload.Manifest.Services) == 0 {
		return nil, fmt.Errorf("workload %s has no recorded services", deploymentID)
	}

	service, err := selectWorkloadService(workload.Manifest.Services, containerHint)
	if err != nil {
		return nil, err
	}

	deploymentName := sanitizeWorkloadServiceName(service.Name)
	pods, err := r.client.ResolveDeploymentPods(ctx, workload.Namespace, deploymentName)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods available for deployment %s", deploymentID)
	}

	selectedPod := selectDeploymentPod(pods)
	containerName := service.Name
	if containerHint != "" {
		containerName = containerHint
	}

	return &workloadTarget{
		Workload:      workload,
		Namespace:     workload.Namespace,
		Deployment:    deploymentName,
		PodName:       selectedPod.Name,
		ContainerName: containerName,
	}, nil
}

func selectWorkloadService(services []provider_daemon.ServiceSpec, containerHint string) (provider_daemon.ServiceSpec, error) {
	if len(services) == 0 {
		return provider_daemon.ServiceSpec{}, fmt.Errorf("workload has no services")
	}
	if containerHint == "" {
		return services[0], nil
	}

	for _, svc := range services {
		if svc.Name == containerHint || sanitizeWorkloadServiceName(svc.Name) == sanitizeWorkloadServiceName(containerHint) {
			return svc, nil
		}
	}

	return provider_daemon.ServiceSpec{}, fmt.Errorf("container %s not found in workload", containerHint)
}

func selectDeploymentPod(pods []deploymentPod) deploymentPod {
	selected := pods[0]
	for _, pod := range pods {
		if pod.Ready && strings.EqualFold(pod.Phase, "running") {
			return pod
		}
		if !selected.Ready && pod.Ready {
			selected = pod
			continue
		}
		if selected.CreatedAt.Before(pod.CreatedAt) && strings.EqualFold(pod.Phase, "running") {
			selected = pod
		}
	}
	return selected
}

func readKubernetesLogEntries(reader io.Reader) ([]provider_daemon.LogEntry, error) {
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	entries := make([]provider_daemon.LogEntry, 0)
	for scanner.Scan() {
		entries = append(entries, parseKubernetesLogLine(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func parseKubernetesLogLine(line string) provider_daemon.LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return provider_daemon.LogEntry{Timestamp: time.Now().UTC(), Level: "info", Message: ""}
	}

	timestamp := time.Now().UTC()
	message := line
	if parts := strings.SplitN(line, " ", 2); len(parts) == 2 {
		if parsed, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
			timestamp = parsed
			message = strings.TrimSpace(parts[1])
		}
	}

	return provider_daemon.LogEntry{
		Timestamp: timestamp,
		Level:     guessKubernetesLogLevel(message),
		Message:   message,
	}
}

func guessKubernetesLogLevel(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "fatal"), strings.Contains(lower, "panic"), strings.Contains(lower, "error"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warn"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}

func sanitizeWorkloadServiceName(name string) string {
	sanitized := strings.ToLower(name)
	return strings.ReplaceAll(sanitized, "_", "-")
}
