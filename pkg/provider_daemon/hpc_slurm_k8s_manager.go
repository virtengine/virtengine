package provider_daemon

import (
	"context"
	"fmt"

	"github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
)

// HPCSlurmK8sManager manages SLURM-on-K8s bootstrap lifecycle.
type HPCSlurmK8sManager struct {
	config    HPCSlurmK8sConfig
	clusterID string
	provider  string
	adapter   *slurm_k8s.SLURMKubernetesAdapter
	running   bool
}

// NewHPCSlurmK8sManager creates a new manager.
func NewHPCSlurmK8sManager(
	config HPCSlurmK8sConfig,
	clusterID string,
	providerAddress string,
	reporter slurm_k8s.OnChainReporter,
) (*HPCSlurmK8sManager, error) {
	if !config.Enabled {
		return nil, nil
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, fmt.Errorf("cluster_id required for slurm_k8s bootstrap")
	}
	if providerAddress == "" {
		return nil, fmt.Errorf("provider_address required for slurm_k8s bootstrap")
	}

	helmClient := slurm_k8s.NewHelmCLIClient(config.Helm)
	k8sClient := slurm_k8s.NewKubeCLIStatusChecker(config.Kube)

	adapter := slurm_k8s.NewSLURMKubernetesAdapter(slurm_k8s.AdapterConfig{
		Helm:                helmClient,
		K8s:                 k8sClient,
		Reporter:            reporter,
		ChartPath:           config.HelmChartPath,
		HealthCheckInterval: config.HealthCheckInterval,
	})

	return &HPCSlurmK8sManager{
		config:    config,
		clusterID: clusterID,
		provider:  providerAddress,
		adapter:   adapter,
	}, nil
}

// Start starts the manager and performs bootstrap when enabled.
func (m *HPCSlurmK8sManager) Start(ctx context.Context) error {
	if m == nil {
		return nil
	}
	if m.running {
		return nil
	}
	if err := m.adapter.Start(ctx); err != nil {
		return fmt.Errorf("slurm_k8s adapter start: %w", err)
	}

	if m.config.BootstrapOnStart {
		_, err := m.adapter.Bootstrap(ctx, slurm_k8s.DeploymentConfig{
			ClusterID:        m.clusterID,
			ClusterName:      m.config.ClusterName,
			ProviderAddress:  m.provider,
			Namespace:        m.config.Namespace,
			Template:         m.config.Template,
			HelmReleaseName:  m.config.HelmReleaseName,
			HelmChartPath:    m.config.HelmChartPath,
			ValuesOverrides:  m.config.ValuesOverrides,
			ImageRegistry:    m.config.ImageRegistry,
			StorageClass:     m.config.StorageClass,
			ProviderEndpoint: m.config.ProviderEndpoint,
		}, slurm_k8s.BootstrapOptions{
			DeployOptions: slurm_k8s.DeployOptions{
				ReadyTimeout:      m.config.ReadyTimeout,
				AllowDegraded:     m.config.AllowDegraded,
				RollbackOnFailure: m.config.RollbackOnFailure,
				MinComputeReady:   m.config.MinComputeReady,
			},
			RetryAttempts: m.config.BootstrapRetryAttempts,
			RetryBackoff:  m.config.BootstrapRetryBackoff,
		})
		if err != nil {
			_ = m.adapter.Stop()
			return fmt.Errorf("slurm_k8s bootstrap failed: %w", err)
		}
	}

	m.running = true
	return nil
}

// Stop stops the manager.
func (m *HPCSlurmK8sManager) Stop() error {
	if m == nil || !m.running {
		return nil
	}
	m.running = false
	return m.adapter.Stop()
}

// IsRunning returns true if the manager is running.
func (m *HPCSlurmK8sManager) IsRunning() bool {
	if m == nil {
		return false
	}
	return m.running
}
