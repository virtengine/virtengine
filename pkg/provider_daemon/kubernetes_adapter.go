// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-403: Provider Daemon Kubernetes orchestration adapter (v1)
package provider_daemon

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrWorkloadNotFound is returned when a workload is not found
var ErrWorkloadNotFound = errors.New("workload not found")

// ErrInvalidTransition is returned when a state transition is invalid
var ErrInvalidTransition = errors.New("invalid state transition")

// ErrDeploymentFailed is returned when deployment fails
var ErrDeploymentFailed = errors.New("deployment failed")

// WorkloadState represents the state of a workload
type WorkloadState string

const (
	// WorkloadStatePending indicates the workload is pending deployment
	WorkloadStatePending WorkloadState = "pending"

	// WorkloadStateDeploying indicates the workload is being deployed
	WorkloadStateDeploying WorkloadState = "deploying"

	// WorkloadStateRunning indicates the workload is running
	WorkloadStateRunning WorkloadState = "running"

	// WorkloadStatePaused indicates the workload is paused
	WorkloadStatePaused WorkloadState = "paused"

	// WorkloadStateStopping indicates the workload is stopping
	WorkloadStateStopping WorkloadState = "stopping"

	// WorkloadStateStopped indicates the workload is stopped
	WorkloadStateStopped WorkloadState = "stopped"

	// WorkloadStateFailed indicates the workload has failed
	WorkloadStateFailed WorkloadState = "failed"

	// WorkloadStateTerminated indicates the workload is terminated
	WorkloadStateTerminated WorkloadState = "terminated"
)

// validTransitions defines valid state transitions
var validTransitions = map[WorkloadState][]WorkloadState{
	WorkloadStatePending:    {WorkloadStateDeploying, WorkloadStateFailed},
	WorkloadStateDeploying:  {WorkloadStateRunning, WorkloadStateFailed, WorkloadStateStopped},
	WorkloadStateRunning:    {WorkloadStatePaused, WorkloadStateStopping, WorkloadStateFailed},
	WorkloadStatePaused:     {WorkloadStateRunning, WorkloadStateStopping, WorkloadStateFailed},
	WorkloadStateStopping:   {WorkloadStateStopped, WorkloadStateFailed},
	WorkloadStateStopped:    {WorkloadStateTerminated, WorkloadStateDeploying},
	WorkloadStateFailed:     {WorkloadStateTerminated, WorkloadStateDeploying},
	WorkloadStateTerminated: {},
}

// IsValidTransition checks if a state transition is valid
func IsValidTransition(from, to WorkloadState) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// DeployedWorkload represents a deployed workload
type DeployedWorkload struct {
	// ID is the workload ID
	ID string

	// DeploymentID is the on-chain deployment ID
	DeploymentID string

	// LeaseID is the on-chain lease ID
	LeaseID string

	// Namespace is the Kubernetes namespace
	Namespace string

	// State is the current state
	State WorkloadState

	// Manifest is the manifest used for deployment
	Manifest *Manifest

	// CreatedAt is when the workload was created
	CreatedAt time.Time

	// UpdatedAt is when the workload was last updated
	UpdatedAt time.Time

	// StatusMessage contains status details
	StatusMessage string

	// Resources contains deployed resource names
	Resources []DeployedResource

	// Endpoints contains exposed endpoints
	Endpoints []WorkloadEndpoint
}

// DeployedResource represents a deployed Kubernetes resource
type DeployedResource struct {
	// Kind is the resource kind
	Kind string

	// Name is the resource name
	Name string

	// Namespace is the resource namespace
	Namespace string
}

// WorkloadEndpoint represents an exposed endpoint
type WorkloadEndpoint struct {
	// Service is the service name
	Service string

	// Port is the port number
	Port int32

	// Protocol is the protocol
	Protocol string

	// ExternalAddress is the external address (if exposed)
	ExternalAddress string

	// InternalAddress is the internal address
	InternalAddress string
}

// SecretData represents secret data for a workload
type SecretData struct {
	// Name is the secret name
	Name string

	// Data contains the secret key-value pairs
	Data map[string][]byte

	// Type is the secret type (opaque, tls, etc.)
	Type string
}

// DeploymentOptions contains deployment options
type DeploymentOptions struct {
	// DryRun if true, validates without deploying
	DryRun bool

	// Timeout is the deployment timeout
	Timeout time.Duration

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// Labels are additional labels to apply
	Labels map[string]string

	// Annotations are additional annotations to apply
	Annotations map[string]string

	// Secrets contains secrets to inject
	Secrets []SecretData
}

// KubernetesClient is the interface for Kubernetes operations
type KubernetesClient interface {
	// CreateNamespace creates a namespace
	CreateNamespace(ctx context.Context, name string, labels map[string]string) error

	// DeleteNamespace deletes a namespace
	DeleteNamespace(ctx context.Context, name string) error

	// CreateDeployment creates a Kubernetes deployment
	CreateDeployment(ctx context.Context, namespace string, spec *K8sDeploymentSpec) error

	// UpdateDeployment updates a deployment
	UpdateDeployment(ctx context.Context, namespace string, spec *K8sDeploymentSpec) error

	// DeleteDeployment deletes a deployment
	DeleteDeployment(ctx context.Context, namespace, name string) error

	// CreateService creates a Kubernetes service
	CreateService(ctx context.Context, namespace string, spec *K8sServiceSpec) error

	// DeleteService deletes a service
	DeleteService(ctx context.Context, namespace, name string) error

	// CreateSecret creates a Kubernetes secret
	CreateSecret(ctx context.Context, namespace, name string, data map[string][]byte) error

	// DeleteSecret deletes a secret
	DeleteSecret(ctx context.Context, namespace, name string) error

	// CreatePVC creates a persistent volume claim
	CreatePVC(ctx context.Context, namespace string, spec *K8sPVCSpec) error

	// DeletePVC deletes a PVC
	DeletePVC(ctx context.Context, namespace, name string) error

	// GetPodStatus gets the status of pods in a deployment
	GetPodStatus(ctx context.Context, namespace, deploymentName string) ([]PodStatus, error)

	// ApplyNetworkPolicy applies a network policy
	ApplyNetworkPolicy(ctx context.Context, namespace string, spec *K8sNetworkPolicySpec) error

	// GetServiceEndpoints gets service endpoints
	GetServiceEndpoints(ctx context.Context, namespace, serviceName string) ([]WorkloadEndpoint, error)
}

// K8sDeploymentSpec represents a Kubernetes deployment specification
type K8sDeploymentSpec struct {
	Name        string
	Replicas    int32
	Labels      map[string]string
	Annotations map[string]string
	Containers  []K8sContainerSpec
	Volumes     []K8sVolumeSpec
}

// K8sContainerSpec represents a container specification
type K8sContainerSpec struct {
	Name            string
	Image           string
	Command         []string
	Args            []string
	Env             map[string]string
	EnvFromSecrets  []string
	Ports           []K8sPortSpec
	Resources       K8sResourceSpec
	VolumeMounts    []K8sVolumeMountSpec
	LivenessProbe   *K8sProbeSpec
	ReadinessProbe  *K8sProbeSpec
	SecurityContext *K8sSecurityContext
}

// K8sPortSpec represents a port specification
type K8sPortSpec struct {
	Name          string
	ContainerPort int32
	Protocol      string
}

// K8sResourceSpec represents resource requirements
type K8sResourceSpec struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
	GPULimit      string
}

// K8sVolumeSpec represents a volume specification
type K8sVolumeSpec struct {
	Name      string
	PVCName   string
	EmptyDir  bool
	SecretRef string
}

// K8sVolumeMountSpec represents a volume mount
type K8sVolumeMountSpec struct {
	Name      string
	MountPath string
	ReadOnly  bool
	SubPath   string
}

// K8sProbeSpec represents a probe specification
type K8sProbeSpec struct {
	HTTPGet             *K8sHTTPGetAction
	Exec                *K8sExecAction
	TCPSocket           *K8sTCPSocketAction
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	FailureThreshold    int32
	SuccessThreshold    int32
}

// K8sHTTPGetAction represents an HTTP GET action
type K8sHTTPGetAction struct {
	Path   string
	Port   int32
	Scheme string
}

// K8sExecAction represents an exec action
type K8sExecAction struct {
	Command []string
}

// K8sTCPSocketAction represents a TCP socket action
type K8sTCPSocketAction struct {
	Port int32
}

// K8sSecurityContext represents security context
type K8sSecurityContext struct {
	RunAsNonRoot bool
	RunAsUser    int64
	ReadOnlyRoot bool
	Capabilities []string
}

// K8sServiceSpec represents a Kubernetes service specification
type K8sServiceSpec struct {
	Name        string
	Labels      map[string]string
	Selector    map[string]string
	Ports       []K8sServicePortSpec
	Type        string // ClusterIP, NodePort, LoadBalancer
	Annotations map[string]string
}

// K8sServicePortSpec represents a service port
type K8sServicePortSpec struct {
	Name       string
	Port       int32
	TargetPort int32
	Protocol   string
	NodePort   int32
}

// K8sPVCSpec represents a PVC specification
type K8sPVCSpec struct {
	Name         string
	StorageClass string
	Size         string
	AccessModes  []string
}

// K8sNetworkPolicySpec represents a network policy
type K8sNetworkPolicySpec struct {
	Name          string
	PodSelector   map[string]string
	IngressRules  []K8sNetworkPolicyRule
	EgressRules   []K8sNetworkPolicyRule
	PolicyTypes   []string
}

// K8sNetworkPolicyRule represents a network policy rule
type K8sNetworkPolicyRule struct {
	Ports        []K8sNetworkPolicyPort
	FromSelector map[string]string
	ToSelector   map[string]string
}

// K8sNetworkPolicyPort represents a network policy port
type K8sNetworkPolicyPort struct {
	Protocol string
	Port     int32
}

// PodStatus represents the status of a pod
type PodStatus struct {
	Name       string
	Phase      string
	Ready      bool
	Restarts   int32
	Message    string
	StartTime  time.Time
	Containers []ContainerStatus
}

// ContainerStatus represents container status
type ContainerStatus struct {
	Name         string
	Ready        bool
	RestartCount int32
	State        string
	Message      string
}

// KubernetesAdapter manages workload deployments to Kubernetes
type KubernetesAdapter struct {
	mu        sync.RWMutex
	client    KubernetesClient
	parser    *ManifestParser
	workloads map[string]*DeployedWorkload

	// providerID is the provider's on-chain ID
	providerID string

	// resourcePrefix is the prefix for all resources
	resourcePrefix string

	// defaultLabels are applied to all resources
	defaultLabels map[string]string

	// statusUpdateChan receives status updates
	statusUpdateChan chan<- WorkloadStatusUpdate
}

// WorkloadStatusUpdate is sent when workload status changes
type WorkloadStatusUpdate struct {
	WorkloadID    string
	DeploymentID  string
	LeaseID       string
	State         WorkloadState
	Message       string
	Timestamp     time.Time
}

// KubernetesAdapterConfig configures the adapter
type KubernetesAdapterConfig struct {
	// Client is the Kubernetes client
	Client KubernetesClient

	// ProviderID is the provider's on-chain ID
	ProviderID string

	// ResourcePrefix is a prefix for resource names
	ResourcePrefix string

	// StatusUpdateChan receives status updates
	StatusUpdateChan chan<- WorkloadStatusUpdate
}

// NewKubernetesAdapter creates a new Kubernetes adapter
func NewKubernetesAdapter(cfg KubernetesAdapterConfig) *KubernetesAdapter {
	return &KubernetesAdapter{
		client:           cfg.Client,
		parser:           NewManifestParser(),
		workloads:        make(map[string]*DeployedWorkload),
		providerID:       cfg.ProviderID,
		resourcePrefix:   cfg.ResourcePrefix,
		statusUpdateChan: cfg.StatusUpdateChan,
		defaultLabels: map[string]string{
			"virtengine.com/managed-by": "provider-daemon",
			"virtengine.com/provider":   cfg.ProviderID,
		},
	}
}

// Deploy deploys a workload from a manifest
func (ka *KubernetesAdapter) Deploy(ctx context.Context, manifest *Manifest, deploymentID, leaseID string, opts DeploymentOptions) (*DeployedWorkload, error) {
	// Validate manifest
	result := ka.parser.Validate(manifest)
	if !result.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidManifest, result.Errors)
	}

	// Generate workload ID
	workloadID := ka.generateWorkloadID(deploymentID, leaseID)

	// Generate namespace
	namespace := ka.generateNamespace(workloadID)

	// Create workload record
	workload := &DeployedWorkload{
		ID:           workloadID,
		DeploymentID: deploymentID,
		LeaseID:      leaseID,
		Namespace:    namespace,
		State:        WorkloadStatePending,
		Manifest:     manifest,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Resources:    make([]DeployedResource, 0),
		Endpoints:    make([]WorkloadEndpoint, 0),
	}

	ka.mu.Lock()
	ka.workloads[workloadID] = workload
	ka.mu.Unlock()

	// Dry run mode
	if opts.DryRun {
		return workload, nil
	}

	// Deploy with timeout
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	deployCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform deployment
	if err := ka.performDeployment(deployCtx, workload, opts); err != nil {
		ka.updateWorkloadState(workloadID, WorkloadStateFailed, err.Error())
		return workload, err
	}

	ka.updateWorkloadState(workloadID, WorkloadStateRunning, "Deployment successful")
	return workload, nil
}

func (ka *KubernetesAdapter) performDeployment(ctx context.Context, workload *DeployedWorkload, opts DeploymentOptions) error {
	ka.updateWorkloadState(workload.ID, WorkloadStateDeploying, "Creating namespace")

	// Merge labels
	labels := ka.mergeLabels(opts.Labels, map[string]string{
		"virtengine.com/deployment": workload.DeploymentID,
		"virtengine.com/lease":      workload.LeaseID,
	})

	// Create namespace
	if err := ka.client.CreateNamespace(ctx, workload.Namespace, labels); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	workload.Resources = append(workload.Resources, DeployedResource{
		Kind: "Namespace", Name: workload.Namespace,
	})

	// Create secrets
	for _, secret := range opts.Secrets {
		if err := ka.client.CreateSecret(ctx, workload.Namespace, secret.Name, secret.Data); err != nil {
			return fmt.Errorf("failed to create secret %s: %w", secret.Name, err)
		}
		workload.Resources = append(workload.Resources, DeployedResource{
			Kind: "Secret", Name: secret.Name, Namespace: workload.Namespace,
		})
	}

	// Create PVCs for volumes
	for _, vol := range workload.Manifest.Volumes {
		if vol.Type == "persistent" {
			pvcSpec := ka.buildPVCSpec(vol, opts)
			if err := ka.client.CreatePVC(ctx, workload.Namespace, pvcSpec); err != nil {
				return fmt.Errorf("failed to create PVC %s: %w", vol.Name, err)
			}
			workload.Resources = append(workload.Resources, DeployedResource{
				Kind: "PersistentVolumeClaim", Name: pvcSpec.Name, Namespace: workload.Namespace,
			})
		}
	}

	// Deploy services
	for _, svc := range workload.Manifest.Services {
		// Create deployment
		deploySpec := ka.buildDeploymentSpec(&svc, workload, opts)
		if err := ka.client.CreateDeployment(ctx, workload.Namespace, deploySpec); err != nil {
			return fmt.Errorf("failed to create deployment %s: %w", svc.Name, err)
		}
		workload.Resources = append(workload.Resources, DeployedResource{
			Kind: "Deployment", Name: deploySpec.Name, Namespace: workload.Namespace,
		})

		// Create service if ports are exposed
		if len(svc.Ports) > 0 {
			svcSpec := ka.buildServiceSpec(&svc, workload, opts)
			if err := ka.client.CreateService(ctx, workload.Namespace, svcSpec); err != nil {
				return fmt.Errorf("failed to create service %s: %w", svc.Name, err)
			}
			workload.Resources = append(workload.Resources, DeployedResource{
				Kind: "Service", Name: svcSpec.Name, Namespace: workload.Namespace,
			})

			// Get endpoints
			endpoints, err := ka.client.GetServiceEndpoints(ctx, workload.Namespace, svcSpec.Name)
			if err == nil {
				workload.Endpoints = append(workload.Endpoints, endpoints...)
			}
		}
	}

	// Apply network policies
	if len(workload.Manifest.Networks) > 0 {
		policy := ka.buildNetworkPolicy(workload, opts)
		if err := ka.client.ApplyNetworkPolicy(ctx, workload.Namespace, policy); err != nil {
			return fmt.Errorf("failed to apply network policy: %w", err)
		}
		workload.Resources = append(workload.Resources, DeployedResource{
			Kind: "NetworkPolicy", Name: policy.Name, Namespace: workload.Namespace,
		})
	}

	return nil
}

// GetWorkload retrieves a deployed workload
func (ka *KubernetesAdapter) GetWorkload(workloadID string) (*DeployedWorkload, error) {
	ka.mu.RLock()
	defer ka.mu.RUnlock()

	workload, ok := ka.workloads[workloadID]
	if !ok {
		return nil, ErrWorkloadNotFound
	}
	return workload, nil
}

// GetWorkloadByLease retrieves a workload by lease ID
func (ka *KubernetesAdapter) GetWorkloadByLease(leaseID string) (*DeployedWorkload, error) {
	ka.mu.RLock()
	defer ka.mu.RUnlock()

	for _, w := range ka.workloads {
		if w.LeaseID == leaseID {
			return w, nil
		}
	}
	return nil, ErrWorkloadNotFound
}

// ListWorkloads lists all workloads
func (ka *KubernetesAdapter) ListWorkloads() []*DeployedWorkload {
	ka.mu.RLock()
	defer ka.mu.RUnlock()

	result := make([]*DeployedWorkload, 0, len(ka.workloads))
	for _, w := range ka.workloads {
		result = append(result, w)
	}
	return result
}

// Terminate terminates a workload
func (ka *KubernetesAdapter) Terminate(ctx context.Context, workloadID string) error {
	ka.mu.Lock()
	workload, ok := ka.workloads[workloadID]
	if !ok {
		ka.mu.Unlock()
		return ErrWorkloadNotFound
	}
	ka.mu.Unlock()

	// Check valid transition
	if !IsValidTransition(workload.State, WorkloadStateStopping) {
		// If already stopped or terminated, that's fine
		if workload.State == WorkloadStateStopped || workload.State == WorkloadStateTerminated {
			return nil
		}
		return fmt.Errorf("%w: cannot transition from %s to stopping", ErrInvalidTransition, workload.State)
	}

	ka.updateWorkloadState(workloadID, WorkloadStateStopping, "Terminating workload")

	// Delete namespace (cascades to all resources)
	if err := ka.client.DeleteNamespace(ctx, workload.Namespace); err != nil {
		ka.updateWorkloadState(workloadID, WorkloadStateFailed, fmt.Sprintf("Failed to delete namespace: %v", err))
		return err
	}

	ka.updateWorkloadState(workloadID, WorkloadStateStopped, "Workload stopped")
	ka.updateWorkloadState(workloadID, WorkloadStateTerminated, "Workload terminated")

	return nil
}

// Pause pauses a workload (scales to 0)
func (ka *KubernetesAdapter) Pause(ctx context.Context, workloadID string) error {
	workload, err := ka.GetWorkload(workloadID)
	if err != nil {
		return err
	}

	if !IsValidTransition(workload.State, WorkloadStatePaused) {
		return fmt.Errorf("%w: cannot transition from %s to paused", ErrInvalidTransition, workload.State)
	}

	// Scale all deployments to 0
	for _, svc := range workload.Manifest.Services {
		spec := &K8sDeploymentSpec{
			Name:     ka.resourceName(svc.Name),
			Replicas: 0,
		}
		if err := ka.client.UpdateDeployment(ctx, workload.Namespace, spec); err != nil {
			return fmt.Errorf("failed to pause deployment %s: %w", svc.Name, err)
		}
	}

	ka.updateWorkloadState(workloadID, WorkloadStatePaused, "Workload paused")
	return nil
}

// Resume resumes a paused workload
func (ka *KubernetesAdapter) Resume(ctx context.Context, workloadID string) error {
	workload, err := ka.GetWorkload(workloadID)
	if err != nil {
		return err
	}

	if !IsValidTransition(workload.State, WorkloadStateRunning) {
		return fmt.Errorf("%w: cannot transition from %s to running", ErrInvalidTransition, workload.State)
	}

	// Scale deployments back to original replicas
	for _, svc := range workload.Manifest.Services {
		replicas := svc.Replicas
		if replicas == 0 {
			replicas = 1
		}
		spec := &K8sDeploymentSpec{
			Name:     ka.resourceName(svc.Name),
			Replicas: replicas,
		}
		if err := ka.client.UpdateDeployment(ctx, workload.Namespace, spec); err != nil {
			return fmt.Errorf("failed to resume deployment %s: %w", svc.Name, err)
		}
	}

	ka.updateWorkloadState(workloadID, WorkloadStateRunning, "Workload resumed")
	return nil
}

// GetStatus gets the current status of a workload
func (ka *KubernetesAdapter) GetStatus(ctx context.Context, workloadID string) (*WorkloadStatusUpdate, error) {
	workload, err := ka.GetWorkload(workloadID)
	if err != nil {
		return nil, err
	}

	// Get pod statuses
	allReady := true
	var messages []string

	for _, svc := range workload.Manifest.Services {
		pods, err := ka.client.GetPodStatus(ctx, workload.Namespace, ka.resourceName(svc.Name))
		if err != nil {
			messages = append(messages, fmt.Sprintf("%s: error getting status", svc.Name))
			allReady = false
			continue
		}

		for _, pod := range pods {
			if !pod.Ready {
				allReady = false
				if pod.Message != "" {
					messages = append(messages, fmt.Sprintf("%s: %s", pod.Name, pod.Message))
				}
			}
		}
	}

	message := workload.StatusMessage
	if len(messages) > 0 {
		message = strings.Join(messages, "; ")
	}

	// Update state based on pod status
	if workload.State == WorkloadStateRunning && !allReady {
		// Pods are not ready, might be deploying or failing
		// Don't change state automatically, just report
	}

	return &WorkloadStatusUpdate{
		WorkloadID:   workloadID,
		DeploymentID: workload.DeploymentID,
		LeaseID:      workload.LeaseID,
		State:        workload.State,
		Message:      message,
		Timestamp:    time.Now(),
	}, nil
}

func (ka *KubernetesAdapter) updateWorkloadState(workloadID string, state WorkloadState, message string) {
	ka.mu.Lock()
	workload, ok := ka.workloads[workloadID]
	if ok {
		workload.State = state
		workload.StatusMessage = message
		workload.UpdatedAt = time.Now()
	}
	ka.mu.Unlock()

	// Send status update
	if ka.statusUpdateChan != nil && ok {
		select {
		case ka.statusUpdateChan <- WorkloadStatusUpdate{
			WorkloadID:   workloadID,
			DeploymentID: workload.DeploymentID,
			LeaseID:      workload.LeaseID,
			State:        state,
			Message:      message,
			Timestamp:    time.Now(),
		}:
		default:
			// Channel full, drop update
		}
	}
}

func (ka *KubernetesAdapter) generateWorkloadID(deploymentID, leaseID string) string {
	hash := sha256.Sum256([]byte(deploymentID + ":" + leaseID))
	return hex.EncodeToString(hash[:8])
}

func (ka *KubernetesAdapter) generateNamespace(workloadID string) string {
	prefix := ka.resourcePrefix
	if prefix == "" {
		prefix = "ve"
	}
	return fmt.Sprintf("%s-%s", prefix, workloadID)
}

func (ka *KubernetesAdapter) resourceName(name string) string {
	// Sanitize and prefix resource names
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	return sanitized
}

func (ka *KubernetesAdapter) mergeLabels(custom, base map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range ka.defaultLabels {
		result[k] = v
	}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range custom {
		result[k] = v
	}
	return result
}

func (ka *KubernetesAdapter) buildDeploymentSpec(svc *ServiceSpec, workload *DeployedWorkload, opts DeploymentOptions) *K8sDeploymentSpec {
	replicas := svc.Replicas
	if replicas == 0 {
		replicas = 1
	}

	labels := ka.mergeLabels(opts.Labels, map[string]string{
		"virtengine.com/service": svc.Name,
	})

	container := K8sContainerSpec{
		Name:    svc.Name,
		Image:   svc.Image + ":" + svc.Tag,
		Command: svc.Command,
		Args:    svc.Args,
		Env:     svc.Env,
		Resources: K8sResourceSpec{
			CPURequest:    fmt.Sprintf("%dm", svc.Resources.CPU),
			CPULimit:      fmt.Sprintf("%dm", svc.Resources.CPU),
			MemoryRequest: fmt.Sprintf("%d", svc.Resources.Memory),
			MemoryLimit:   fmt.Sprintf("%d", svc.Resources.Memory),
		},
		SecurityContext: &K8sSecurityContext{
			RunAsNonRoot: true,
			ReadOnlyRoot: false,
		},
	}

	// Add GPU if specified
	if svc.Resources.GPU > 0 {
		container.Resources.GPULimit = fmt.Sprintf("%d", svc.Resources.GPU)
	}

	// Add ports
	for _, port := range svc.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		container.Ports = append(container.Ports, K8sPortSpec{
			Name:          port.Name,
			ContainerPort: port.ContainerPort,
			Protocol:      strings.ToUpper(protocol),
		})
	}

	// Add volume mounts
	for _, mount := range svc.Volumes {
		container.VolumeMounts = append(container.VolumeMounts, K8sVolumeMountSpec{
			Name:      mount.Name,
			MountPath: mount.MountPath,
			ReadOnly:  mount.ReadOnly,
			SubPath:   mount.SubPath,
		})
	}

	// Add health check
	if svc.HealthCheck != nil {
		container.LivenessProbe = ka.buildProbeSpec(svc.HealthCheck)
		container.ReadinessProbe = ka.buildProbeSpec(svc.HealthCheck)
	}

	// Build volumes
	var volumes []K8sVolumeSpec
	for _, mount := range svc.Volumes {
		// Find volume spec
		for _, vol := range workload.Manifest.Volumes {
			if vol.Name == mount.Name {
				if vol.Type == "persistent" {
					volumes = append(volumes, K8sVolumeSpec{
						Name:    mount.Name,
						PVCName: mount.Name,
					})
				} else {
					volumes = append(volumes, K8sVolumeSpec{
						Name:     mount.Name,
						EmptyDir: true,
					})
				}
				break
			}
		}
	}

	return &K8sDeploymentSpec{
		Name:        ka.resourceName(svc.Name),
		Replicas:    replicas,
		Labels:      labels,
		Annotations: opts.Annotations,
		Containers:  []K8sContainerSpec{container},
		Volumes:     volumes,
	}
}

func (ka *KubernetesAdapter) buildServiceSpec(svc *ServiceSpec, workload *DeployedWorkload, opts DeploymentOptions) *K8sServiceSpec {
	labels := ka.mergeLabels(opts.Labels, map[string]string{
		"virtengine.com/service": svc.Name,
	})

	selector := map[string]string{
		"virtengine.com/service": svc.Name,
	}

	var ports []K8sServicePortSpec
	serviceType := "ClusterIP"

	for _, port := range svc.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "tcp"
		}

		sp := K8sServicePortSpec{
			Name:       port.Name,
			Port:       port.ContainerPort,
			TargetPort: port.ContainerPort,
			Protocol:   strings.ToUpper(protocol),
		}

		if port.Expose {
			serviceType = "LoadBalancer"
			if port.ExternalPort > 0 {
				sp.Port = port.ExternalPort
			}
		}

		ports = append(ports, sp)
	}

	return &K8sServiceSpec{
		Name:        ka.resourceName(svc.Name),
		Labels:      labels,
		Selector:    selector,
		Ports:       ports,
		Type:        serviceType,
		Annotations: opts.Annotations,
	}
}

func (ka *KubernetesAdapter) buildPVCSpec(vol VolumeSpec, opts DeploymentOptions) *K8sPVCSpec {
	storageClass := vol.StorageClass
	if storageClass == "" {
		storageClass = "standard"
	}

	return &K8sPVCSpec{
		Name:         vol.Name,
		StorageClass: storageClass,
		Size:         fmt.Sprintf("%d", vol.Size),
		AccessModes:  []string{"ReadWriteOnce"},
	}
}

func (ka *KubernetesAdapter) buildNetworkPolicy(workload *DeployedWorkload, opts DeploymentOptions) *K8sNetworkPolicySpec {
	// Default deny all ingress except from within namespace
	return &K8sNetworkPolicySpec{
		Name:        "default-network-policy",
		PodSelector: map[string]string{},
		IngressRules: []K8sNetworkPolicyRule{
			{
				FromSelector: map[string]string{
					"virtengine.com/deployment": workload.DeploymentID,
				},
			},
		},
		PolicyTypes: []string{"Ingress"},
	}
}

func (ka *KubernetesAdapter) buildProbeSpec(check *HealthCheckSpec) *K8sProbeSpec {
	probe := &K8sProbeSpec{
		InitialDelaySeconds: check.InitialDelaySeconds,
		PeriodSeconds:       check.PeriodSeconds,
		TimeoutSeconds:      check.TimeoutSeconds,
		FailureThreshold:    check.FailureThreshold,
		SuccessThreshold:    check.SuccessThreshold,
	}

	if check.HTTP != nil {
		scheme := check.HTTP.Scheme
		if scheme == "" {
			scheme = "HTTP"
		}
		probe.HTTPGet = &K8sHTTPGetAction{
			Path:   check.HTTP.Path,
			Port:   check.HTTP.Port,
			Scheme: strings.ToUpper(scheme),
		}
	} else if check.Exec != nil {
		probe.Exec = &K8sExecAction{
			Command: check.Exec.Command,
		}
	} else if check.TCP != nil {
		probe.TCPSocket = &K8sTCPSocketAction{
			Port: check.TCP.Port,
		}
	}

	return probe
}
