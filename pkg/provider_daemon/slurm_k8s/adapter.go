// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
//
// VE-502: SLURM Kubernetes deployment automation for provider daemon
package slurm_k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// ClusterState represents the state of a SLURM cluster deployment
type ClusterState string

const (
	// ClusterStatePending indicates the cluster is pending deployment
	ClusterStatePending ClusterState = "pending"

	// ClusterStateDeploying indicates the cluster is being deployed
	ClusterStateDeploying ClusterState = "deploying"

	// ClusterStateRunning indicates the cluster is running
	ClusterStateRunning ClusterState = "running"

	// ClusterStateUpdating indicates the cluster is being updated
	ClusterStateUpdating ClusterState = "updating"

	// ClusterStateDegraded indicates the cluster is running with issues
	ClusterStateDegraded ClusterState = "degraded"

	// ClusterStateStopping indicates the cluster is being stopped
	ClusterStateStopping ClusterState = "stopping"

	// ClusterStateStopped indicates the cluster is stopped
	ClusterStateStopped ClusterState = "stopped"

	// ClusterStateFailed indicates the cluster deployment failed
	ClusterStateFailed ClusterState = "failed"
)

// DeploymentConfig contains configuration for SLURM cluster deployment
type DeploymentConfig struct {
	// ClusterID is the unique identifier for the cluster
	ClusterID string `json:"cluster_id"`

	// ClusterName is a human-readable name
	ClusterName string `json:"cluster_name"`

	// ProviderAddress is the blockchain provider address
	ProviderAddress string `json:"provider_address"`

	// Namespace is the Kubernetes namespace for deployment
	Namespace string `json:"namespace"`

	// Template is the cluster template configuration
	Template *hpctypes.ClusterTemplate `json:"template"`

	// HelmReleaseName is the Helm release name
	HelmReleaseName string `json:"helm_release_name"`

	// HelmChartPath is the path to the Helm chart
	HelmChartPath string `json:"helm_chart_path"`

	// ValuesOverrides contains Helm values overrides
	ValuesOverrides map[string]interface{} `json:"values_overrides"`

	// ImageRegistry is the container image registry
	ImageRegistry string `json:"image_registry"`

	// StorageClass is the Kubernetes storage class for PVCs
	StorageClass string `json:"storage_class"`

	// ProviderEndpoint is the provider daemon gRPC endpoint
	ProviderEndpoint string `json:"provider_endpoint"`

	// TLSConfig contains TLS configuration
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`
}

// TLSConfig contains TLS configuration for secure communication
type TLSConfig struct {
	// Enabled indicates if TLS is enabled
	Enabled bool `json:"enabled"`

	// CACert is the CA certificate (PEM)
	CACert string `json:"ca_cert,omitempty"`

	// ClientCert is the client certificate (PEM)
	ClientCert string `json:"client_cert,omitempty"`

	// ClientKey is the client private key (PEM)
	ClientKey string `json:"client_key,omitempty"`
}

// DeployedCluster represents a deployed SLURM cluster
type DeployedCluster struct {
	// Config is the deployment configuration
	Config DeploymentConfig `json:"config"`

	// State is the current state
	State ClusterState `json:"state"`

	// StatusMessage contains additional status information
	StatusMessage string `json:"status_message"`

	// DeployedAt is when the cluster was deployed
	DeployedAt time.Time `json:"deployed_at"`

	// UpdatedAt is when the cluster was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// HealthStatus contains component health status
	HealthStatus *ClusterHealthStatus `json:"health_status,omitempty"`

	// Capacity contains cluster capacity information
	Capacity *ClusterCapacity `json:"capacity,omitempty"`
}

// ClusterHealthStatus contains health status for cluster components
type ClusterHealthStatus struct {
	// ControllerReady indicates if slurmctld is ready
	ControllerReady bool `json:"controller_ready"`

	// DatabaseReady indicates if slurmdbd is ready
	DatabaseReady bool `json:"database_ready"`

	// ComputeNodesReady is the number of ready compute nodes
	ComputeNodesReady int32 `json:"compute_nodes_ready"`

	// ComputeNodesTotal is the total number of compute nodes
	ComputeNodesTotal int32 `json:"compute_nodes_total"`

	// MungeHealthy indicates if munge authentication is healthy
	MungeHealthy bool `json:"munge_healthy"`

	// LastHealthCheck is when health was last checked
	LastHealthCheck time.Time `json:"last_health_check"`

	// Errors contains any current errors
	Errors []string `json:"errors,omitempty"`
}

// ClusterCapacity contains cluster resource capacity
type ClusterCapacity struct {
	// TotalNodes is the total number of nodes
	TotalNodes int32 `json:"total_nodes"`

	// AvailableNodes is the number of available nodes
	AvailableNodes int32 `json:"available_nodes"`

	// TotalCPUs is the total CPU cores
	TotalCPUs int64 `json:"total_cpus"`

	// AvailableCPUs is the available CPU cores
	AvailableCPUs int64 `json:"available_cpus"`

	// TotalMemoryGB is the total memory in GB
	TotalMemoryGB int64 `json:"total_memory_gb"`

	// AvailableMemoryGB is the available memory in GB
	AvailableMemoryGB int64 `json:"available_memory_gb"`

	// TotalGPUs is the total GPU count
	TotalGPUs int64 `json:"total_gpus"`

	// AvailableGPUs is the available GPU count
	AvailableGPUs int64 `json:"available_gpus"`

	// GPUTypes lists available GPU types
	GPUTypes []string `json:"gpu_types,omitempty"`

	// LastUpdate is when capacity was last updated
	LastUpdate time.Time `json:"last_update"`
}

// ScaleRequest represents a request to scale the cluster
type ScaleRequest struct {
	// TargetNodes is the target number of compute nodes
	TargetNodes int32 `json:"target_nodes"`

	// NodePool is the node pool to scale (empty for default)
	NodePool string `json:"node_pool,omitempty"`

	// Timeout is the scaling timeout
	Timeout time.Duration `json:"timeout"`
}

// UpgradeRequest represents a request to upgrade the cluster
type UpgradeRequest struct {
	// TargetVersion is the target SLURM version
	TargetVersion string `json:"target_version"`

	// RollingUpdate if true, performs rolling update
	RollingUpdate bool `json:"rolling_update"`

	// DrainTimeout is the timeout for draining nodes
	DrainTimeout time.Duration `json:"drain_timeout"`
}

// ClusterStatusUpdate is sent when cluster status changes
type ClusterStatusUpdate struct {
	// ClusterID is the cluster identifier
	ClusterID string `json:"cluster_id"`

	// State is the current state
	State ClusterState `json:"state"`

	// StatusMessage contains additional status
	StatusMessage string `json:"status_message"`

	// HealthStatus contains health information
	HealthStatus *ClusterHealthStatus `json:"health_status,omitempty"`

	// Capacity contains capacity information
	Capacity *ClusterCapacity `json:"capacity,omitempty"`

	// Timestamp is when the update was generated
	Timestamp time.Time `json:"timestamp"`
}

// HelmClient is the interface for Helm operations
type HelmClient interface {
	// Install installs a Helm chart
	Install(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error

	// Upgrade upgrades a Helm release
	Upgrade(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error

	// Uninstall uninstalls a Helm release
	Uninstall(ctx context.Context, releaseName, namespace string) error

	// GetRelease gets a Helm release
	GetRelease(ctx context.Context, releaseName, namespace string) (*HelmRelease, error)

	// ListReleases lists Helm releases in a namespace
	ListReleases(ctx context.Context, namespace string) ([]*HelmRelease, error)
}

// HelmRelease represents a Helm release
type HelmRelease struct {
	// Name is the release name
	Name string

	// Namespace is the namespace
	Namespace string

	// Chart is the chart name
	Chart string

	// Version is the chart version
	Version string

	// AppVersion is the app version
	AppVersion string

	// Status is the release status
	Status string

	// Values are the release values
	Values map[string]interface{}
}

// KubernetesStatusChecker checks Kubernetes resource status
type KubernetesStatusChecker interface {
	// GetStatefulSetStatus gets StatefulSet status
	GetStatefulSetStatus(ctx context.Context, namespace, name string) (*StatefulSetStatus, error)

	// GetPodLogs gets pod logs
	GetPodLogs(ctx context.Context, namespace, podName, containerName string, lines int) (string, error)

	// ExecInPod executes a command in a pod
	ExecInPod(ctx context.Context, namespace, podName, containerName string, command []string) (string, error)
}

// StatefulSetStatus contains StatefulSet status information
type StatefulSetStatus struct {
	// Name is the StatefulSet name
	Name string

	// Replicas is the desired replica count
	Replicas int32

	// ReadyReplicas is the ready replica count
	ReadyReplicas int32

	// CurrentReplicas is the current replica count
	CurrentReplicas int32

	// UpdatedReplicas is the updated replica count
	UpdatedReplicas int32

	// Conditions contains the conditions
	Conditions []string
}

// OnChainReporter reports cluster status to the blockchain
type OnChainReporter interface {
	// ReportClusterStatus reports cluster status on-chain
	ReportClusterStatus(ctx context.Context, clusterID string, status *ClusterStatusUpdate) error

	// ReportCapacityUpdate reports capacity update on-chain
	ReportCapacityUpdate(ctx context.Context, clusterID string, capacity *ClusterCapacity) error

	// ReportNodeJoin reports a node joining the cluster
	ReportNodeJoin(ctx context.Context, clusterID, nodeID string) error

	// ReportNodeLeave reports a node leaving the cluster
	ReportNodeLeave(ctx context.Context, clusterID, nodeID string) error
}

// SLURMKubernetesAdapter orchestrates SLURM cluster deployments on Kubernetes
type SLURMKubernetesAdapter struct {
	mu       sync.RWMutex
	helm     HelmClient
	k8s      KubernetesStatusChecker
	reporter OnChainReporter

	// clusters maps cluster ID to deployed cluster
	clusters map[string]*DeployedCluster

	// statusChan receives status updates
	statusChan chan<- ClusterStatusUpdate

	// stopCh is used to stop background workers
	stopCh chan struct{}

	// running indicates if the adapter is running
	running bool

	// chartPath is the default Helm chart path
	chartPath string

	// healthCheckInterval is the health check interval
	healthCheckInterval time.Duration
}

// AdapterConfig configures the SLURM Kubernetes adapter
type AdapterConfig struct {
	// Helm is the Helm client
	Helm HelmClient

	// K8s is the Kubernetes status checker
	K8s KubernetesStatusChecker

	// Reporter is the on-chain reporter
	Reporter OnChainReporter

	// StatusChan receives status updates
	StatusChan chan<- ClusterStatusUpdate

	// ChartPath is the default Helm chart path
	ChartPath string

	// HealthCheckInterval is the health check interval
	HealthCheckInterval time.Duration
}

// NewSLURMKubernetesAdapter creates a new SLURM Kubernetes adapter
func NewSLURMKubernetesAdapter(cfg AdapterConfig) *SLURMKubernetesAdapter {
	interval := cfg.HealthCheckInterval
	if interval == 0 {
		interval = 60 * time.Second
	}

	return &SLURMKubernetesAdapter{
		helm:                cfg.Helm,
		k8s:                 cfg.K8s,
		reporter:            cfg.Reporter,
		clusters:            make(map[string]*DeployedCluster),
		statusChan:          cfg.StatusChan,
		stopCh:              make(chan struct{}),
		chartPath:           cfg.ChartPath,
		healthCheckInterval: interval,
	}
}

// Start starts the adapter
func (a *SLURMKubernetesAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.mu.Unlock()

	// Start health check loop
	go a.healthCheckLoop()

	return nil
}

// Stop stops the adapter
func (a *SLURMKubernetesAdapter) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = false
	close(a.stopCh)
	a.mu.Unlock()

	return nil
}

// Deploy deploys a new SLURM cluster
func (a *SLURMKubernetesAdapter) Deploy(ctx context.Context, config DeploymentConfig) (*DeployedCluster, error) {
	if !a.IsRunning() {
		return nil, fmt.Errorf("adapter not running")
	}

	if config.ClusterID == "" {
		return nil, fmt.Errorf("cluster_id required")
	}

	// Check if cluster already exists
	a.mu.RLock()
	_, exists := a.clusters[config.ClusterID]
	a.mu.RUnlock()
	if exists {
		return nil, fmt.Errorf("cluster %s already exists", config.ClusterID)
	}

	// Create cluster record
	cluster := &DeployedCluster{
		Config:     config,
		State:      ClusterStatePending,
		DeployedAt: time.Now(),
		UpdatedAt:  time.Now(),
	}

	a.mu.Lock()
	a.clusters[config.ClusterID] = cluster
	a.mu.Unlock()

	// Build Helm values
	values := a.buildHelmValues(config)

	// Update state to deploying
	a.updateClusterState(config.ClusterID, ClusterStateDeploying, "Installing Helm chart")

	// Install Helm chart
	chartPath := config.HelmChartPath
	if chartPath == "" {
		chartPath = a.chartPath
	}
	releaseName := config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", config.ClusterID)
	}

	if err := a.helm.Install(ctx, releaseName, chartPath, config.Namespace, values); err != nil {
		a.updateClusterState(config.ClusterID, ClusterStateFailed, fmt.Sprintf("Helm install failed: %v", err))
		return cluster, fmt.Errorf("failed to install Helm chart: %w", err)
	}

	// Wait for deployment to be ready
	if err := a.waitForReady(ctx, config.ClusterID, 10*time.Minute); err != nil {
		a.updateClusterState(config.ClusterID, ClusterStateDegraded, fmt.Sprintf("Deployment not ready: %v", err))
		return cluster, nil // Return cluster even if not fully ready
	}

	a.updateClusterState(config.ClusterID, ClusterStateRunning, "Cluster deployed successfully")

	// Report to on-chain
	if a.reporter != nil {
		a.reportStatus(ctx, config.ClusterID)
	}

	return a.GetCluster(config.ClusterID)
}

// Upgrade upgrades an existing SLURM cluster
func (a *SLURMKubernetesAdapter) Upgrade(ctx context.Context, clusterID string, req UpgradeRequest) error {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return err
	}

	if cluster.State != ClusterStateRunning && cluster.State != ClusterStateDegraded {
		return fmt.Errorf("cannot upgrade cluster in state %s", cluster.State)
	}

	a.updateClusterState(clusterID, ClusterStateUpdating, fmt.Sprintf("Upgrading to version %s", req.TargetVersion))

	// Build updated values
	values := a.buildHelmValues(cluster.Config)
	values["global"] = map[string]interface{}{
		"slurmVersion": req.TargetVersion,
	}

	// Get release name
	releaseName := cluster.Config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", clusterID)
	}
	chartPath := cluster.Config.HelmChartPath
	if chartPath == "" {
		chartPath = a.chartPath
	}

	// Perform Helm upgrade
	if err := a.helm.Upgrade(ctx, releaseName, chartPath, cluster.Config.Namespace, values); err != nil {
		a.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Upgrade failed: %v", err))
		return fmt.Errorf("failed to upgrade Helm release: %w", err)
	}

	// Wait for upgrade to complete
	timeout := 15 * time.Minute
	if req.DrainTimeout > 0 {
		timeout = req.DrainTimeout + 5*time.Minute
	}
	if err := a.waitForReady(ctx, clusterID, timeout); err != nil {
		a.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Upgrade not ready: %v", err))
		return nil
	}

	a.updateClusterState(clusterID, ClusterStateRunning, fmt.Sprintf("Upgraded to version %s", req.TargetVersion))
	return nil
}

// Scale scales the compute nodes in a cluster
func (a *SLURMKubernetesAdapter) Scale(ctx context.Context, clusterID string, req ScaleRequest) error {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return err
	}

	if cluster.State != ClusterStateRunning && cluster.State != ClusterStateDegraded {
		return fmt.Errorf("cannot scale cluster in state %s", cluster.State)
	}

	a.updateClusterState(clusterID, ClusterStateUpdating, fmt.Sprintf("Scaling to %d nodes", req.TargetNodes))

	// Build updated values
	values := a.buildHelmValues(cluster.Config)
	if req.NodePool == "" {
		values["compute"] = map[string]interface{}{
			"replicas": req.TargetNodes,
		}
	} else {
		// Scale specific node pool
		// This would need to modify the nodePools section
	}

	// Get release name
	releaseName := cluster.Config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", clusterID)
	}
	chartPath := cluster.Config.HelmChartPath
	if chartPath == "" {
		chartPath = a.chartPath
	}

	// Perform Helm upgrade for scaling
	if err := a.helm.Upgrade(ctx, releaseName, chartPath, cluster.Config.Namespace, values); err != nil {
		a.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Scale failed: %v", err))
		return fmt.Errorf("failed to scale cluster: %w", err)
	}

	// Wait for scaling to complete
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	if err := a.waitForReady(ctx, clusterID, timeout); err != nil {
		a.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Scale not ready: %v", err))
		return nil
	}

	a.updateClusterState(clusterID, ClusterStateRunning, fmt.Sprintf("Scaled to %d nodes", req.TargetNodes))

	// Report capacity update
	if a.reporter != nil {
		a.reportCapacity(ctx, clusterID)
	}

	return nil
}

// Terminate terminates a SLURM cluster
func (a *SLURMKubernetesAdapter) Terminate(ctx context.Context, clusterID string) error {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return err
	}

	a.updateClusterState(clusterID, ClusterStateStopping, "Terminating cluster")

	// Get release name
	releaseName := cluster.Config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", clusterID)
	}

	// Uninstall Helm release
	if err := a.helm.Uninstall(ctx, releaseName, cluster.Config.Namespace); err != nil {
		a.updateClusterState(clusterID, ClusterStateFailed, fmt.Sprintf("Uninstall failed: %v", err))
		return fmt.Errorf("failed to uninstall Helm release: %w", err)
	}

	a.updateClusterState(clusterID, ClusterStateStopped, "Cluster terminated")

	// Remove from clusters map
	a.mu.Lock()
	delete(a.clusters, clusterID)
	a.mu.Unlock()

	return nil
}

// GetCluster gets a deployed cluster by ID
func (a *SLURMKubernetesAdapter) GetCluster(clusterID string) (*DeployedCluster, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cluster, exists := a.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}
	return cluster, nil
}

// ListClusters lists all deployed clusters
func (a *SLURMKubernetesAdapter) ListClusters() []*DeployedCluster {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*DeployedCluster, 0, len(a.clusters))
	for _, c := range a.clusters {
		result = append(result, c)
	}
	return result
}

// GetClusterHealth gets cluster health status
func (a *SLURMKubernetesAdapter) GetClusterHealth(ctx context.Context, clusterID string) (*ClusterHealthStatus, error) {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	health := &ClusterHealthStatus{
		LastHealthCheck: time.Now(),
	}

	fullname := fmt.Sprintf("slurm-%s", clusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Check controller
	ctrlStatus, err := a.k8s.GetStatefulSetStatus(ctx, cluster.Config.Namespace, fullname+"-controller")
	if err != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("controller check failed: %v", err))
	} else {
		health.ControllerReady = ctrlStatus.ReadyReplicas >= 1
	}

	// Check database
	dbStatus, err := a.k8s.GetStatefulSetStatus(ctx, cluster.Config.Namespace, fullname+"-slurmdbd")
	if err != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("database check failed: %v", err))
	} else {
		health.DatabaseReady = dbStatus.ReadyReplicas >= 1
	}

	// Check compute nodes
	computeStatus, err := a.k8s.GetStatefulSetStatus(ctx, cluster.Config.Namespace, fullname+"-compute")
	if err != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("compute check failed: %v", err))
	} else {
		health.ComputeNodesReady = computeStatus.ReadyReplicas
		health.ComputeNodesTotal = computeStatus.Replicas
	}

	// Check munge (via scontrol ping on controller)
	if health.ControllerReady {
		output, err := a.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld", []string{"scontrol", "ping"})
		if err == nil && output != "" {
			health.MungeHealthy = true
		}
	}

	// Update cluster health
	a.mu.Lock()
	if c, exists := a.clusters[clusterID]; exists {
		c.HealthStatus = health
		c.UpdatedAt = time.Now()
	}
	a.mu.Unlock()

	return health, nil
}

// GetClusterCapacity gets cluster capacity
func (a *SLURMKubernetesAdapter) GetClusterCapacity(ctx context.Context, clusterID string) (*ClusterCapacity, error) {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	fullname := fmt.Sprintf("slurm-%s", clusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Get capacity from sinfo
	output, err := a.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"sinfo", "-N", "-h", "-o", "%n %c %m %G %t"})
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster capacity: %w", err)
	}

	// Parse sinfo output
	capacity := a.parseSinfoOutput(output)

	// Update cluster capacity
	a.mu.Lock()
	if c, exists := a.clusters[clusterID]; exists {
		c.Capacity = capacity
		c.UpdatedAt = time.Now()
	}
	a.mu.Unlock()

	return capacity, nil
}

// IsRunning checks if the adapter is running
func (a *SLURMKubernetesAdapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// buildHelmValues builds Helm values from deployment config
func (a *SLURMKubernetesAdapter) buildHelmValues(config DeploymentConfig) map[string]interface{} {
	values := map[string]interface{}{
		"cluster": map[string]interface{}{
			"id":              config.ClusterID,
			"name":            config.ClusterName,
			"providerAddress": config.ProviderAddress,
		},
	}

	if config.ImageRegistry != "" {
		values["global"] = map[string]interface{}{
			"imageRegistry": config.ImageRegistry,
		}
	}

	if config.StorageClass != "" {
		if global, ok := values["global"].(map[string]interface{}); ok {
			global["storageClass"] = config.StorageClass
		} else {
			values["global"] = map[string]interface{}{
				"storageClass": config.StorageClass,
			}
		}
	}

	// Apply template configuration
	if config.Template != nil {
		values["partitions"] = a.convertPartitions(config.Template.Partitions)
		if len(config.Template.QoSPolicies) > 0 {
			values["qos"] = a.convertQoS(config.Template.QoSPolicies)
		}

		// Apply compute configuration from template
		if len(config.Template.Partitions) > 0 {
			totalNodes := int32(0)
			for _, p := range config.Template.Partitions {
				totalNodes += p.Nodes
			}
			values["compute"] = map[string]interface{}{
				"replicas": totalNodes,
			}
		}

		// Apply controller config from template
		if config.Template.SchedulingPolicy.SchedulerType != "" {
			values["controller"] = map[string]interface{}{
				"config": map[string]interface{}{
					"schedulerType":            config.Template.SchedulingPolicy.SchedulerType,
					"backfillEnabled":          config.Template.SchedulingPolicy.BackfillEnabled,
					"preemptionEnabled":        config.Template.SchedulingPolicy.PreemptionEnabled,
					"priorityWeightAge":        config.Template.SchedulingPolicy.PriorityWeightAge,
					"priorityWeightFairshare":  config.Template.SchedulingPolicy.PriorityWeightFairShare,
					"priorityWeightJobSize":    config.Template.SchedulingPolicy.PriorityWeightJobSize,
					"priorityWeightPartition":  config.Template.SchedulingPolicy.PriorityWeightPartition,
					"priorityWeightQOS":        config.Template.SchedulingPolicy.PriorityWeightQoS,
				},
			}
		}
	}

	// Apply node agent configuration
	if config.ProviderEndpoint != "" {
		values["nodeAgent"] = map[string]interface{}{
			"enabled": true,
			"config": map[string]interface{}{
				"providerEndpoint": config.ProviderEndpoint,
			},
		}

		if config.TLSConfig != nil && config.TLSConfig.Enabled {
			nodeAgent := values["nodeAgent"].(map[string]interface{})
			nodeAgent["tls"] = map[string]interface{}{
				"enabled":    true,
				"caCert":     config.TLSConfig.CACert,
				"clientCert": config.TLSConfig.ClientCert,
				"clientKey":  config.TLSConfig.ClientKey,
			}
		}
	}

	// Merge values overrides
	for k, v := range config.ValuesOverrides {
		values[k] = v
	}

	return values
}

// convertPartitions converts HPC partitions to Helm values
func (a *SLURMKubernetesAdapter) convertPartitions(partitions []hpctypes.PartitionConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(partitions))
	for _, p := range partitions {
		partition := map[string]interface{}{
			"name":     p.Name,
			"maxNodes": p.MaxNodesPerJob,
			"state":    p.State,
		}

		if p.MaxRuntimeSeconds > 0 {
			hours := p.MaxRuntimeSeconds / 3600
			mins := (p.MaxRuntimeSeconds % 3600) / 60
			partition["maxTime"] = fmt.Sprintf("%02d:%02d:00", hours, mins)
		}

		if p.DefaultRuntimeSeconds > 0 {
			hours := p.DefaultRuntimeSeconds / 3600
			mins := (p.DefaultRuntimeSeconds % 3600) / 60
			partition["defaultTime"] = fmt.Sprintf("%02d:%02d:00", hours, mins)
		}

		if p.Priority > 0 {
			partition["priority"] = p.Priority
		}

		result = append(result, partition)
	}
	return result
}

// convertQoS converts HPC QoS policies to Helm values
func (a *SLURMKubernetesAdapter) convertQoS(qos []hpctypes.QoSPolicy) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(qos))
	for _, q := range qos {
		policy := map[string]interface{}{
			"name":     q.Name,
			"priority": q.Priority,
		}

		if q.MaxJobsPerUser > 0 {
			policy["maxJobsPerUser"] = q.MaxJobsPerUser
		}

		if q.MaxWallDurationSeconds > 0 {
			hours := q.MaxWallDurationSeconds / 3600
			mins := (q.MaxWallDurationSeconds % 3600) / 60
			policy["maxWallDuration"] = fmt.Sprintf("%02d:%02d:00", hours, mins)
		}

		if q.PreemptMode != "" {
			policy["preemptMode"] = q.PreemptMode
		}

		result = append(result, policy)
	}
	return result
}

// parseSinfoOutput parses sinfo output to ClusterCapacity
func (a *SLURMKubernetesAdapter) parseSinfoOutput(output string) *ClusterCapacity {
	capacity := &ClusterCapacity{
		LastUpdate: time.Now(),
		GPUTypes:   make([]string, 0),
	}

	// Parse each line: nodename cpus memory gres state
	// Example: compute-0 64 256000 gpu:nvidia:8 idle
	lines := splitLines(output)
	gpuTypes := make(map[string]bool)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var nodeName, gres, state string
		var cpus int64
		var memory int64

		_, err := fmt.Sscanf(line, "%s %d %d %s %s", &nodeName, &cpus, &memory, &gres, &state)
		if err != nil {
			continue
		}

		capacity.TotalNodes++
		capacity.TotalCPUs += cpus
		capacity.TotalMemoryGB += memory / 1024 // Convert MB to GB

		// Parse GPU count from gres (format: gpu:type:count or (null))
		if gres != "(null)" && gres != "" {
			gpuCount := parseGPUCount(gres)
			capacity.TotalGPUs += int64(gpuCount)
			gpuType := parseGPUType(gres)
			if gpuType != "" {
				gpuTypes[gpuType] = true
			}
		}

		// Check if node is available
		if state == "idle" || state == "mixed" {
			capacity.AvailableNodes++
			capacity.AvailableCPUs += cpus
			capacity.AvailableMemoryGB += memory / 1024
			if gres != "(null)" && gres != "" {
				capacity.AvailableGPUs += int64(parseGPUCount(gres))
			}
		}
	}

	for gpuType := range gpuTypes {
		capacity.GPUTypes = append(capacity.GPUTypes, gpuType)
	}

	return capacity
}

// updateClusterState updates cluster state and sends status update
func (a *SLURMKubernetesAdapter) updateClusterState(clusterID string, state ClusterState, message string) {
	a.mu.Lock()
	cluster, exists := a.clusters[clusterID]
	if exists {
		cluster.State = state
		cluster.StatusMessage = message
		cluster.UpdatedAt = time.Now()
	}
	a.mu.Unlock()

	if a.statusChan != nil && exists {
		select {
		case a.statusChan <- ClusterStatusUpdate{
			ClusterID:     clusterID,
			State:         state,
			StatusMessage: message,
			Timestamp:     time.Now(),
		}:
		default:
			// Channel full, drop update
		}
	}
}

// waitForReady waits for cluster to be ready
func (a *SLURMKubernetesAdapter) waitForReady(ctx context.Context, clusterID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for cluster to be ready")
			}

			health, err := a.GetClusterHealth(ctx, clusterID)
			if err != nil {
				continue
			}

			if health.ControllerReady && health.DatabaseReady && health.ComputeNodesReady > 0 {
				return nil
			}
		}
	}
}

// healthCheckLoop periodically checks cluster health
func (a *SLURMKubernetesAdapter) healthCheckLoop() {
	ticker := time.NewTicker(a.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.checkAllClustersHealth()
		}
	}
}

// checkAllClustersHealth checks health of all clusters
func (a *SLURMKubernetesAdapter) checkAllClustersHealth() {
	ctx := context.Background()

	a.mu.RLock()
	clusterIDs := make([]string, 0, len(a.clusters))
	for id := range a.clusters {
		clusterIDs = append(clusterIDs, id)
	}
	a.mu.RUnlock()

	for _, id := range clusterIDs {
		health, err := a.GetClusterHealth(ctx, id)
		if err != nil {
			continue
		}

		// Update state based on health
		a.mu.Lock()
		cluster, exists := a.clusters[id]
		if exists && cluster.State == ClusterStateRunning {
			if len(health.Errors) > 0 || !health.ControllerReady || !health.DatabaseReady {
				cluster.State = ClusterStateDegraded
				cluster.StatusMessage = "Health check detected issues"
			}
		} else if exists && cluster.State == ClusterStateDegraded {
			if len(health.Errors) == 0 && health.ControllerReady && health.DatabaseReady {
				cluster.State = ClusterStateRunning
				cluster.StatusMessage = "Cluster recovered"
			}
		}
		a.mu.Unlock()
	}
}

// reportStatus reports cluster status on-chain
func (a *SLURMKubernetesAdapter) reportStatus(ctx context.Context, clusterID string) {
	if a.reporter == nil {
		return
	}

	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return
	}

	update := &ClusterStatusUpdate{
		ClusterID:     clusterID,
		State:         cluster.State,
		StatusMessage: cluster.StatusMessage,
		HealthStatus:  cluster.HealthStatus,
		Capacity:      cluster.Capacity,
		Timestamp:     time.Now(),
	}

	_ = a.reporter.ReportClusterStatus(ctx, clusterID, update)
}

// reportCapacity reports cluster capacity on-chain
func (a *SLURMKubernetesAdapter) reportCapacity(ctx context.Context, clusterID string) {
	if a.reporter == nil {
		return
	}

	capacity, err := a.GetClusterCapacity(ctx, clusterID)
	if err != nil {
		return
	}

	_ = a.reporter.ReportCapacityUpdate(ctx, clusterID, capacity)
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// parseGPUCount parses GPU count from GRES string
func parseGPUCount(gres string) int {
	// Format: gpu:type:count or gpu:count
	parts := splitByColon(gres)
	if len(parts) < 2 || parts[0] != "gpu" {
		return 0
	}

	// Try gpu:type:count format
	if len(parts) >= 3 {
		var count int
		if _, err := fmt.Sscanf(parts[2], "%d", &count); err == nil {
			return count
		}
	}

	// Try gpu:count format
	if len(parts) == 2 {
		var count int
		if _, err := fmt.Sscanf(parts[1], "%d", &count); err == nil {
			return count
		}
	}
	return 0
}

// parseGPUType parses GPU type from GRES string
func parseGPUType(gres string) string {
	// Format: gpu:type:count
	parts := splitByColon(gres)
	if len(parts) >= 3 && parts[0] == "gpu" {
		return parts[1]
	}
	return ""
}

// splitByColon splits a string by colon
func splitByColon(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
