package provider_daemon

import (
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCSlurmK8sConfig configures SLURM-on-Kubernetes bootstrap.
type HPCSlurmK8sConfig struct {
	Enabled          bool                      `json:"enabled" yaml:"enabled"`
	BootstrapOnStart bool                      `json:"bootstrap_on_start" yaml:"bootstrap_on_start"`
	Namespace        string                    `json:"namespace" yaml:"namespace"`
	ClusterName      string                    `json:"cluster_name" yaml:"cluster_name"`
	HelmChartPath    string                    `json:"helm_chart_path" yaml:"helm_chart_path"`
	HelmReleaseName  string                    `json:"helm_release_name" yaml:"helm_release_name"`
	ValuesOverrides  map[string]interface{}    `json:"values_overrides" yaml:"values_overrides"`
	ImageRegistry    string                    `json:"image_registry" yaml:"image_registry"`
	StorageClass     string                    `json:"storage_class" yaml:"storage_class"`
	ProviderEndpoint string                    `json:"provider_endpoint" yaml:"provider_endpoint"`
	Template         *hpctypes.ClusterTemplate `json:"template" yaml:"template"`

	ReadyTimeout           time.Duration `json:"ready_timeout" yaml:"ready_timeout"`
	RollbackOnFailure      bool          `json:"rollback_on_failure" yaml:"rollback_on_failure"`
	AllowDegraded          bool          `json:"allow_degraded" yaml:"allow_degraded"`
	MinComputeReady        int32         `json:"min_compute_ready" yaml:"min_compute_ready"`
	HealthCheckInterval    time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	BootstrapRetryAttempts int           `json:"bootstrap_retry_attempts" yaml:"bootstrap_retry_attempts"`
	BootstrapRetryBackoff  time.Duration `json:"bootstrap_retry_backoff" yaml:"bootstrap_retry_backoff"`

	Helm slurm_k8s.HelmCLIConfig `json:"helm" yaml:"helm"`
	Kube slurm_k8s.KubeCLIConfig `json:"kube" yaml:"kube"`
}

// DefaultHPCSlurmK8sConfig returns default SLURM K8s config.
func DefaultHPCSlurmK8sConfig() HPCSlurmK8sConfig {
	return HPCSlurmK8sConfig{
		Enabled:                false,
		BootstrapOnStart:       true,
		Namespace:              "slurm-system",
		ReadyTimeout:           10 * time.Minute,
		RollbackOnFailure:      true,
		AllowDegraded:          false,
		MinComputeReady:        1,
		HealthCheckInterval:    60 * time.Second,
		BootstrapRetryAttempts: 1,
		BootstrapRetryBackoff:  20 * time.Second,
	}
}

// Validate validates the SLURM K8s config.
func (c *HPCSlurmK8sConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Namespace == "" {
		return fmt.Errorf("slurm_k8s.namespace required")
	}
	if c.HelmChartPath == "" {
		return fmt.Errorf("slurm_k8s.helm_chart_path required")
	}
	if c.ReadyTimeout < time.Minute {
		return fmt.Errorf("slurm_k8s.ready_timeout must be >= 1m")
	}
	if c.BootstrapRetryAttempts < 0 {
		return fmt.Errorf("slurm_k8s.bootstrap_retry_attempts cannot be negative")
	}
	if c.BootstrapRetryBackoff != 0 && c.BootstrapRetryBackoff < time.Second {
		return fmt.Errorf("slurm_k8s.bootstrap_retry_backoff must be >= 1s")
	}
	return nil
}
