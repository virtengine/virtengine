// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
package slurm_k8s

import (
	"context"
	"fmt"
	"time"
)

// LifecycleManager handles cluster lifecycle operations
type LifecycleManager struct {
	adapter *SLURMKubernetesAdapter
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(adapter *SLURMKubernetesAdapter) *LifecycleManager {
	return &LifecycleManager{adapter: adapter}
}

// JoinRequest represents a node join request
type JoinRequest struct {
	// ClusterID is the cluster to join
	ClusterID string `json:"cluster_id"`

	// NodeID is the node identifier
	NodeID string `json:"node_id"`

	// NodePool is the node pool (empty for default)
	NodePool string `json:"node_pool,omitempty"`

	// CPUs is the number of CPUs
	CPUs int32 `json:"cpus"`

	// MemoryGB is the memory in GB
	MemoryGB int32 `json:"memory_gb"`

	// GPUs is the number of GPUs
	GPUs int32 `json:"gpus,omitempty"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// Features are the node features
	Features []string `json:"features,omitempty"`
}

// LeaveRequest represents a node leave request
type LeaveRequest struct {
	// ClusterID is the cluster to leave
	ClusterID string `json:"cluster_id"`

	// NodeID is the node identifier
	NodeID string `json:"node_id"`

	// Drain if true, drain jobs before leaving
	Drain bool `json:"drain"`

	// DrainTimeout is the drain timeout
	DrainTimeout time.Duration `json:"drain_timeout,omitempty"`

	// Force if true, force leave even with running jobs
	Force bool `json:"force"`
}

// DrainRequest represents a node drain request
type DrainRequest struct {
	// ClusterID is the cluster ID
	ClusterID string `json:"cluster_id"`

	// NodeID is the node to drain
	NodeID string `json:"node_id"`

	// Reason is the drain reason
	Reason string `json:"reason"`

	// Timeout is the drain timeout
	Timeout time.Duration `json:"timeout,omitempty"`
}

// ResumeRequest represents a node resume request
type ResumeRequest struct {
	// ClusterID is the cluster ID
	ClusterID string `json:"cluster_id"`

	// NodeID is the node to resume
	NodeID string `json:"node_id"`
}

// RollingUpgradeConfig configures rolling upgrade behavior
type RollingUpgradeConfig struct {
	// MaxUnavailable is the max nodes that can be unavailable
	MaxUnavailable int32 `json:"max_unavailable"`

	// DrainTimeout is the timeout for draining each node
	DrainTimeout time.Duration `json:"drain_timeout"`

	// WaitBetweenNodes is the wait time between node upgrades
	WaitBetweenNodes time.Duration `json:"wait_between_nodes"`

	// VerifyAfterUpgrade if true, verify node health after upgrade
	VerifyAfterUpgrade bool `json:"verify_after_upgrade"`
}

// NodeJoin adds a node to the cluster (for dynamic scaling)
func (lm *LifecycleManager) NodeJoin(ctx context.Context, req JoinRequest) error {
	cluster, err := lm.adapter.GetCluster(req.ClusterID)
	if err != nil {
		return err
	}

	if cluster.State != ClusterStateRunning && cluster.State != ClusterStateDegraded {
		return fmt.Errorf("cannot join node to cluster in state %s", cluster.State)
	}

	fullname := fmt.Sprintf("slurm-%s", req.ClusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Build node definition
	gres := ""
	if req.GPUs > 0 {
		gpuType := req.GPUType
		if gpuType == "" {
			gpuType = "nvidia"
		}
		gres = fmt.Sprintf("Gres=gpu:%s:%d", gpuType, req.GPUs)
	}

	features := ""
	if len(req.Features) > 0 {
		features = fmt.Sprintf("Feature=%s", joinStrings(req.Features, ","))
	}

	// Execute scontrol to create node configuration
	cmd := []string{
		"scontrol", "create", "node",
		fmt.Sprintf("NodeName=%s", req.NodeID),
		fmt.Sprintf("CPUs=%d", req.CPUs),
		fmt.Sprintf("RealMemory=%d", req.MemoryGB*1024),
		"State=FUTURE",
	}
	if gres != "" {
		cmd = append(cmd, gres)
	}
	if features != "" {
		cmd = append(cmd, features)
	}

	_, err = lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld", cmd)
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	// Resume the node to make it available
	_, err = lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"scontrol", "update", "NodeName=" + req.NodeID, "State=RESUME"})
	if err != nil {
		return fmt.Errorf("failed to resume node: %w", err)
	}

	// Report node join on-chain
	if lm.adapter.reporter != nil {
		_ = lm.adapter.reporter.ReportNodeJoin(ctx, req.ClusterID, req.NodeID)
	}

	return nil
}

// NodeLeave removes a node from the cluster
func (lm *LifecycleManager) NodeLeave(ctx context.Context, req LeaveRequest) error {
	cluster, err := lm.adapter.GetCluster(req.ClusterID)
	if err != nil {
		return err
	}

	fullname := fmt.Sprintf("slurm-%s", req.ClusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Drain the node if requested
	if req.Drain {
		drainReq := DrainRequest{
			ClusterID: req.ClusterID,
			NodeID:    req.NodeID,
			Reason:    "Node leaving cluster",
			Timeout:   req.DrainTimeout,
		}
		if err := lm.DrainNode(ctx, drainReq); err != nil && !req.Force {
			return fmt.Errorf("failed to drain node: %w", err)
		}
	}

	// Delete the node from SLURM
	_, err = lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"scontrol", "delete", "NodeName=" + req.NodeID})
	if err != nil && !req.Force {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	// Report node leave on-chain
	if lm.adapter.reporter != nil {
		_ = lm.adapter.reporter.ReportNodeLeave(ctx, req.ClusterID, req.NodeID)
	}

	return nil
}

// DrainNode drains a node (waits for jobs to complete)
func (lm *LifecycleManager) DrainNode(ctx context.Context, req DrainRequest) error {
	cluster, err := lm.adapter.GetCluster(req.ClusterID)
	if err != nil {
		return err
	}

	fullname := fmt.Sprintf("slurm-%s", req.ClusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	reason := req.Reason
	if reason == "" {
		reason = "Draining node"
	}

	// Set node to drain state
	_, err = lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"scontrol", "update", "NodeName=" + req.NodeID, "State=DRAIN", "Reason=" + reason})
	if err != nil {
		return fmt.Errorf("failed to drain node: %w", err)
	}

	// Wait for node to be drained
	timeout := req.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for node to drain")
			}

			// Check if node is drained
			output, err := lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
				[]string{"sinfo", "-n", req.NodeID, "-h", "-o", "%t"})
			if err != nil {
				continue
			}

			// If node is in drained state (no jobs running)
			if output == "drained" || output == "drain" {
				return nil
			}
		}
	}
}

// ResumeNode resumes a drained node
func (lm *LifecycleManager) ResumeNode(ctx context.Context, req ResumeRequest) error {
	cluster, err := lm.adapter.GetCluster(req.ClusterID)
	if err != nil {
		return err
	}

	fullname := fmt.Sprintf("slurm-%s", req.ClusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	_, err = lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"scontrol", "update", "NodeName=" + req.NodeID, "State=RESUME"})
	if err != nil {
		return fmt.Errorf("failed to resume node: %w", err)
	}

	return nil
}

// RollingUpgrade performs a rolling upgrade of compute nodes
func (lm *LifecycleManager) RollingUpgrade(ctx context.Context, clusterID string, upgrade UpgradeRequest, config RollingUpgradeConfig) error {
	cluster, err := lm.adapter.GetCluster(clusterID)
	if err != nil {
		return err
	}

	if cluster.State != ClusterStateRunning {
		return fmt.Errorf("cannot perform rolling upgrade on cluster in state %s", cluster.State)
	}

	lm.adapter.updateClusterState(clusterID, ClusterStateUpdating, "Starting rolling upgrade")

	fullname := fmt.Sprintf("slurm-%s", clusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Get list of compute nodes
	output, err := lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"sinfo", "-N", "-h", "-o", "%n"})
	if err != nil {
		return fmt.Errorf("failed to get node list: %w", err)
	}

	nodes := splitLines(output)

	// Upgrade nodes in batches
	batchSize := int(config.MaxUnavailable)
	if batchSize <= 0 {
		batchSize = 1
	}

	for i := 0; i < len(nodes); i += batchSize {
		end := i + batchSize
		if end > len(nodes) {
			end = len(nodes)
		}
		batch := nodes[i:end]

		// Drain batch nodes
		for _, node := range batch {
			if node == "" {
				continue
			}
			if err := lm.DrainNode(ctx, DrainRequest{
				ClusterID: clusterID,
				NodeID:    node,
				Reason:    "Rolling upgrade",
				Timeout:   config.DrainTimeout,
			}); err != nil {
				lm.adapter.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Drain failed for %s: %v", node, err))
				// Continue with upgrade anyway
			}
		}

		// Perform Helm upgrade (this will restart pods)
		if i == 0 {
			// Only need to upgrade once, pods will restart with new image
			values := lm.adapter.buildHelmValues(cluster.Config)
			values["global"] = map[string]interface{}{
				"slurmVersion": upgrade.TargetVersion,
			}

			chartPath := cluster.Config.HelmChartPath
			if chartPath == "" {
				chartPath = lm.adapter.chartPath
			}

			if err := lm.adapter.helm.Upgrade(ctx, fullname, chartPath, cluster.Config.Namespace, values); err != nil {
				lm.adapter.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Helm upgrade failed: %v", err))
				return err
			}
		}

		// Wait for batch nodes to be ready
		if config.VerifyAfterUpgrade {
			for _, node := range batch {
				if node == "" {
					continue
				}
				if err := lm.waitForNodeReady(ctx, clusterID, node, 5*time.Minute); err != nil {
					lm.adapter.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Node %s not ready: %v", node, err))
				}
			}
		}

		// Resume batch nodes
		for _, node := range batch {
			if node == "" {
				continue
			}
			_ = lm.ResumeNode(ctx, ResumeRequest{
				ClusterID: clusterID,
				NodeID:    node,
			})
		}

		// Wait between batches
		if config.WaitBetweenNodes > 0 && end < len(nodes) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(config.WaitBetweenNodes):
			}
		}
	}

	lm.adapter.updateClusterState(clusterID, ClusterStateRunning, fmt.Sprintf("Rolling upgrade to %s completed", upgrade.TargetVersion))
	return nil
}

// waitForNodeReady waits for a node to be ready
func (lm *LifecycleManager) waitForNodeReady(ctx context.Context, clusterID, nodeID string, timeout time.Duration) error {
	cluster, err := lm.adapter.GetCluster(clusterID)
	if err != nil {
		return err
	}

	fullname := fmt.Sprintf("slurm-%s", clusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for node %s to be ready", nodeID)
			}

			output, err := lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
				[]string{"sinfo", "-n", nodeID, "-h", "-o", "%t"})
			if err != nil {
				continue
			}

			// Node is ready if idle, mixed, or allocated
			if output == "idle" || output == "mixed" || output == "alloc" || output == "allocated" {
				return nil
			}
		}
	}
}

// ReconcileCluster reconciles cluster state (recovery after failure)
func (lm *LifecycleManager) ReconcileCluster(ctx context.Context, clusterID string) error {
	cluster, err := lm.adapter.GetCluster(clusterID)
	if err != nil {
		return err
	}

	// Get current health
	health, err := lm.adapter.GetClusterHealth(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster health: %w", err)
	}

	// If controller is not ready, we can't do much
	if !health.ControllerReady {
		lm.adapter.updateClusterState(clusterID, ClusterStateDegraded, "Controller not ready, waiting for recovery")
		return nil
	}

	fullname := fmt.Sprintf("slurm-%s", clusterID)
	if cluster.Config.HelmReleaseName != "" {
		fullname = cluster.Config.HelmReleaseName
	}

	// Get list of nodes and their states
	output, err := lm.adapter.k8s.ExecInPod(ctx, cluster.Config.Namespace, fullname+"-controller-0", "slurmctld",
		[]string{"sinfo", "-N", "-h", "-o", "%n %t"})
	if err != nil {
		return fmt.Errorf("failed to get node status: %w", err)
	}

	// Resume any nodes stuck in DOWN state that should be running
	for _, line := range splitLines(output) {
		if line == "" {
			continue
		}

		var nodeName, state string
		if _, err := fmt.Sscanf(line, "%s %s", &nodeName, &state); err != nil {
			continue
		}

		// If node is down but pod is running, resume it
		if state == "down" || state == "down*" {
			// Check if pod exists and is ready
			podStatus, err := lm.adapter.k8s.GetStatefulSetStatus(ctx, cluster.Config.Namespace, fullname+"-compute")
			if err == nil && podStatus.ReadyReplicas > 0 {
				// Try to resume the node
				_ = lm.ResumeNode(ctx, ResumeRequest{
					ClusterID: clusterID,
					NodeID:    nodeName,
				})
			}
		}
	}

	// Update cluster state based on health
	if health.ControllerReady && health.DatabaseReady && health.ComputeNodesReady == health.ComputeNodesTotal {
		lm.adapter.updateClusterState(clusterID, ClusterStateRunning, "Cluster reconciled successfully")
	} else if health.ControllerReady && health.ComputeNodesReady > 0 {
		lm.adapter.updateClusterState(clusterID, ClusterStateDegraded,
			fmt.Sprintf("Partially ready: %d/%d compute nodes", health.ComputeNodesReady, health.ComputeNodesTotal))
	}

	return nil
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

