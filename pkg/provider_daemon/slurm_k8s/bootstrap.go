// Package slurm_k8s implements SLURM cluster deployment on Kubernetes.
package slurm_k8s

import (
	"context"
	"fmt"
	"time"
)

// DeployOptions controls deployment readiness behavior.
type DeployOptions struct {
	// ReadyTimeout is how long to wait for readiness.
	ReadyTimeout time.Duration

	// AllowDegraded returns a degraded cluster without failing.
	AllowDegraded bool

	// RollbackOnFailure uninstalls the release on readiness failure.
	RollbackOnFailure bool

	// MinComputeReady is the minimum number of compute nodes required.
	MinComputeReady int32
}

// BootstrapOptions controls bootstrap behavior.
type BootstrapOptions struct {
	DeployOptions

	// RetryAttempts is the number of bootstrap retries on failure.
	RetryAttempts int

	// RetryBackoff is the backoff between retries.
	RetryBackoff time.Duration
}

func (o DeployOptions) normalize() DeployOptions {
	if o.ReadyTimeout == 0 {
		o.ReadyTimeout = 10 * time.Minute
	}
	return o
}

func (o BootstrapOptions) normalize() BootstrapOptions {
	o.DeployOptions = o.DeployOptions.normalize()
	if o.RetryBackoff == 0 {
		o.RetryBackoff = 15 * time.Second
	}
	return o
}

// DeployWithOptions deploys a new SLURM cluster with readiness controls.
func (a *SLURMKubernetesAdapter) DeployWithOptions(
	ctx context.Context,
	config DeploymentConfig,
	options DeployOptions,
) (*DeployedCluster, error) {
	if !a.IsRunning() {
		return nil, fmt.Errorf("adapter not running")
	}

	if config.ClusterID == "" {
		return nil, fmt.Errorf("cluster_id required")
	}

	options = options.normalize()

	// Check if cluster already exists
	a.mu.RLock()
	_, exists := a.clusters[config.ClusterID]
	a.mu.RUnlock()
	if exists {
		return nil, fmt.Errorf("cluster %s already exists", config.ClusterID)
	}

	// Create cluster record
	cluster := &DeployedCluster{
		Config:          config,
		State:           ClusterStatePending,
		Phase:           DeploymentPhasePending,
		PhaseTimestamps: make(map[DeploymentPhase]time.Time),
		DeployedAt:      time.Now(),
		UpdatedAt:       time.Now(),
	}
	cluster.PhaseTimestamps[DeploymentPhasePending] = time.Now()

	a.mu.Lock()
	a.clusters[config.ClusterID] = cluster
	a.mu.Unlock()

	// Emit initial phase transition
	a.transitionPhase(config.ClusterID, DeploymentPhasePending, "Cluster record created, preparing deployment", nil)

	// Build Helm values
	values := a.buildHelmValues(config)

	// Update state to deploying and transition to helm installation phase
	a.updateClusterPhaseAndState(config.ClusterID, DeploymentPhaseHelmInstalling, ClusterStateDeploying, "Installing Helm chart")

	chartPath := config.HelmChartPath
	if chartPath == "" {
		chartPath = a.chartPath
	}
	releaseName := config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", config.ClusterID)
	}

	if err := a.helm.Install(ctx, releaseName, chartPath, config.Namespace, values); err != nil {
		a.updateClusterPhaseAndState(config.ClusterID, DeploymentPhaseFailed, ClusterStateFailed, fmt.Sprintf("Helm install failed: %v", err))
		a.transitionPhase(config.ClusterID, DeploymentPhaseFailed, "Helm installation failed", []string{err.Error()})
		return cluster, fmt.Errorf("failed to install Helm chart: %w", err)
	}

	readyErr := a.waitForReady(ctx, config.ClusterID, options.ReadyTimeout, options.MinComputeReady)
	if readyErr != nil {
		a.updateClusterState(config.ClusterID, ClusterStateDegraded, fmt.Sprintf("Deployment not ready: %v", readyErr))
		a.transitionPhase(config.ClusterID, DeploymentPhaseFailed, "Readiness checks failed", []string{readyErr.Error()})
		if options.RollbackOnFailure {
			_ = a.rollbackDeployment(ctx, cluster, releaseName)
		}
		if options.AllowDegraded {
			// Update to registration ready phase even if degraded (allows partial clusters)
			a.transitionPhase(config.ClusterID, DeploymentPhaseRegistrationReady, "Degraded cluster ready for registration", []string{readyErr.Error()})
			return cluster, nil
		}
		return cluster, readyErr
	}

	// All components ready - transition to complete
	a.updateClusterPhaseAndState(config.ClusterID, DeploymentPhaseComplete, ClusterStateRunning, "Cluster deployed successfully")

	if a.reporter != nil {
		a.reportStatus(ctx, config.ClusterID)
	}

	return a.GetCluster(config.ClusterID)
}

// Bootstrap deploys a new SLURM cluster with retry/rollback handling.
func (a *SLURMKubernetesAdapter) Bootstrap(
	ctx context.Context,
	config DeploymentConfig,
	options BootstrapOptions,
) (*DeployedCluster, error) {
	options = options.normalize()

	var lastErr error
	for attempt := 0; attempt <= options.RetryAttempts; attempt++ {
		cluster, err := a.DeployWithOptions(ctx, config, options.DeployOptions)
		if err == nil {
			return cluster, nil
		}
		lastErr = err

		if attempt >= options.RetryAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return cluster, ctx.Err()
		case <-time.After(options.RetryBackoff):
		}
	}

	return nil, fmt.Errorf("bootstrap failed after %d attempts: %w", options.RetryAttempts+1, lastErr)
}

// Redeploy attempts to reinstall an existing cluster.
func (a *SLURMKubernetesAdapter) Redeploy(ctx context.Context, clusterID string, options DeployOptions) (*DeployedCluster, error) {
	cluster, err := a.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	releaseName := cluster.Config.HelmReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("slurm-%s", clusterID)
	}

	a.updateClusterState(clusterID, ClusterStateUpdating, "Redeploying cluster")

	if err := a.helm.Uninstall(ctx, releaseName, cluster.Config.Namespace); err != nil {
		a.updateClusterState(clusterID, ClusterStateDegraded, fmt.Sprintf("Redeploy uninstall failed: %v", err))
		if options.RollbackOnFailure {
			return cluster, err
		}
	}

	a.mu.Lock()
	delete(a.clusters, clusterID)
	a.mu.Unlock()

	return a.DeployWithOptions(ctx, cluster.Config, options)
}

func (a *SLURMKubernetesAdapter) rollbackDeployment(ctx context.Context, cluster *DeployedCluster, releaseName string) error {
	if cluster == nil {
		return nil
	}
	if err := a.helm.Uninstall(ctx, releaseName, cluster.Config.Namespace); err != nil {
		a.updateClusterState(cluster.Config.ClusterID, ClusterStateFailed, fmt.Sprintf("Rollback failed: %v", err))
		return err
	}
	a.updateClusterState(cluster.Config.ClusterID, ClusterStateFailed, "Rolled back failed deployment")
	return nil
}
