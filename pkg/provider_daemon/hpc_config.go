// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPC scheduler configuration for provider daemon
package provider_daemon

import (
	"errors"
	"time"

	"github.com/virtengine/virtengine/pkg/moab_adapter"
	"github.com/virtengine/virtengine/pkg/ood_adapter"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// HPCSchedulerType represents the type of HPC scheduler
type HPCSchedulerType string

const (
	// HPCSchedulerTypeSLURM indicates SLURM scheduler
	HPCSchedulerTypeSLURM HPCSchedulerType = "slurm"

	// HPCSchedulerTypeMOAB indicates MOAB scheduler
	HPCSchedulerTypeMOAB HPCSchedulerType = "moab"

	// HPCSchedulerTypeOOD indicates Open OnDemand
	HPCSchedulerTypeOOD HPCSchedulerType = "ood"
)

// IsValid checks if the scheduler type is valid
func (t HPCSchedulerType) IsValid() bool {
	switch t {
	case HPCSchedulerTypeSLURM, HPCSchedulerTypeMOAB, HPCSchedulerTypeOOD:
		return true
	default:
		return false
	}
}

// HPCConfig contains all HPC-related configuration for the provider daemon
type HPCConfig struct {
	// Enabled enables HPC job processing
	Enabled bool `json:"enabled" yaml:"enabled"`

	// SchedulerType is the type of HPC scheduler to use
	SchedulerType HPCSchedulerType `json:"scheduler_type" yaml:"scheduler_type"`

	// ClusterID is the on-chain cluster ID this provider manages
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`

	// ProviderAddress is the provider's bech32 address
	ProviderAddress string `json:"provider_address" yaml:"provider_address"`

	// SLURM configuration (used when SchedulerType is "slurm")
	SLURM slurm_adapter.SLURMConfig `json:"slurm" yaml:"slurm"`

	// MOAB configuration (used when SchedulerType is "moab")
	MOAB moab_adapter.MOABConfig `json:"moab" yaml:"moab"`

	// OOD configuration (used when SchedulerType is "ood")
	OOD ood_adapter.OODConfig `json:"ood" yaml:"ood"`

	// SLURM-on-Kubernetes bootstrap configuration
	SlurmK8s HPCSlurmK8sConfig `json:"slurm_k8s" yaml:"slurm_k8s"`

	// Node aggregator configuration
	NodeAggregator HPCNodeAggregatorConfig `json:"node_aggregator" yaml:"node_aggregator"`

	// JobService configuration
	JobService HPCJobServiceConfig `json:"job_service" yaml:"job_service"`

	// UsageReporting configuration
	UsageReporting HPCUsageReportingConfig `json:"usage_reporting" yaml:"usage_reporting"`

	// Retry configuration
	Retry HPCRetryConfig `json:"retry" yaml:"retry"`

	// Audit configuration
	Audit HPCAuditConfig `json:"audit" yaml:"audit"`

	// Routing configuration (VE-5B)
	Routing HPCRoutingConfig `json:"routing" yaml:"routing"`

	// EnforcementMode is the routing enforcement mode (for backward compatibility)
	EnforcementMode string `json:"enforcement_mode" yaml:"enforcement_mode"`
}

// HPCJobServiceConfig configures the HPC job service
type HPCJobServiceConfig struct {
	// JobPollInterval is how often to poll for job updates
	JobPollInterval time.Duration `json:"job_poll_interval" yaml:"job_poll_interval"`

	// JobTimeoutDefault is the default job timeout
	JobTimeoutDefault time.Duration `json:"job_timeout_default" yaml:"job_timeout_default"`

	// MaxConcurrentJobs is the maximum number of concurrent jobs
	MaxConcurrentJobs int `json:"max_concurrent_jobs" yaml:"max_concurrent_jobs"`

	// EnableStateRecovery enables recovery of job state on restart
	EnableStateRecovery bool `json:"enable_state_recovery" yaml:"enable_state_recovery"`

	// StateStorePath is where to persist job state
	StateStorePath string `json:"state_store_path" yaml:"state_store_path"`
}

// HPCUsageReportingConfig configures usage reporting
type HPCUsageReportingConfig struct {
	// Enabled enables usage reporting
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ReportInterval is how often to submit usage reports
	ReportInterval time.Duration `json:"report_interval" yaml:"report_interval"`

	// BatchSize is the maximum number of reports to batch
	BatchSize int `json:"batch_size" yaml:"batch_size"`

	// RetryOnFailure enables retry on reporting failure
	RetryOnFailure bool `json:"retry_on_failure" yaml:"retry_on_failure"`
}

// HPCRetryConfig configures retry behavior
type HPCRetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration `json:"initial_backoff" yaml:"initial_backoff"`

	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration `json:"max_backoff" yaml:"max_backoff"`

	// BackoffMultiplier is the backoff multiplier
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`

	// RetryableErrors are error patterns that should be retried
	RetryableErrors []string `json:"retryable_errors" yaml:"retryable_errors"`
}

// HPCAuditConfig configures audit logging
type HPCAuditConfig struct {
	// Enabled enables audit logging
	Enabled bool `json:"enabled" yaml:"enabled"`

	// LogPath is the path to the audit log
	LogPath string `json:"log_path" yaml:"log_path"`

	// LogJobEvents logs job lifecycle events
	LogJobEvents bool `json:"log_job_events" yaml:"log_job_events"`

	// LogSecurityEvents logs security-related events
	LogSecurityEvents bool `json:"log_security_events" yaml:"log_security_events"`

	// LogUsageReports logs usage report submissions
	LogUsageReports bool `json:"log_usage_reports" yaml:"log_usage_reports"`
}

// HPCRoutingConfig configures routing enforcement (VE-5B)
type HPCRoutingConfig struct {
	// Enabled enables routing enforcement
	Enabled bool `json:"enabled" yaml:"enabled"`

	// EnforcementMode is the enforcement mode (strict, permissive, audit_only)
	EnforcementMode string `json:"enforcement_mode" yaml:"enforcement_mode"`

	// MaxDecisionAgeBlocks is the maximum age of a scheduling decision in blocks
	MaxDecisionAgeBlocks int64 `json:"max_decision_age_blocks" yaml:"max_decision_age_blocks"`

	// MaxDecisionAgeSeconds is the maximum age in seconds
	MaxDecisionAgeSeconds int64 `json:"max_decision_age_seconds" yaml:"max_decision_age_seconds"`

	// AllowAutomaticFallback indicates if automatic fallback is permitted
	AllowAutomaticFallback bool `json:"allow_automatic_fallback" yaml:"allow_automatic_fallback"`

	// RequireDecisionForSubmission requires a scheduling decision for job submission
	RequireDecisionForSubmission bool `json:"require_decision_for_submission" yaml:"require_decision_for_submission"`

	// AutoRefreshStaleDecisions automatically refreshes stale decisions
	AutoRefreshStaleDecisions bool `json:"auto_refresh_stale_decisions" yaml:"auto_refresh_stale_decisions"`

	// ViolationAlertThreshold is the number of violations before alerting
	ViolationAlertThreshold int32 `json:"violation_alert_threshold" yaml:"violation_alert_threshold"`
}

// DefaultHPCConfig returns the default HPC configuration
func DefaultHPCConfig() HPCConfig {
	return HPCConfig{
		Enabled:        false,
		SchedulerType:  HPCSchedulerTypeSLURM,
		SLURM:          slurm_adapter.DefaultSLURMConfig(),
		MOAB:           moab_adapter.DefaultMOABConfig(),
		OOD:            ood_adapter.DefaultOODConfig(),
		SlurmK8s:       DefaultHPCSlurmK8sConfig(),
		NodeAggregator: DefaultHPCNodeAggregatorConfig(),
		JobService: HPCJobServiceConfig{
			JobPollInterval:     15 * time.Second,
			JobTimeoutDefault:   24 * time.Hour,
			MaxConcurrentJobs:   100,
			EnableStateRecovery: true,
			StateStorePath:      "/var/lib/virtengine/hpc-state",
		},
		UsageReporting: HPCUsageReportingConfig{
			Enabled:        true,
			ReportInterval: 5 * time.Minute,
			BatchSize:      50,
			RetryOnFailure: true,
		},
		Retry: HPCRetryConfig{
			MaxRetries:        3,
			InitialBackoff:    1 * time.Second,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 2.0,
			RetryableErrors: []string{
				"connection refused",
				"timeout",
				"temporary failure",
			},
		},
		Audit: HPCAuditConfig{
			Enabled:           true,
			LogPath:           "/var/log/virtengine/hpc-audit.log",
			LogJobEvents:      true,
			LogSecurityEvents: true,
			LogUsageReports:   true,
		},
		Routing: HPCRoutingConfig{
			Enabled:                      true,
			EnforcementMode:              "strict",
			MaxDecisionAgeBlocks:         100,
			MaxDecisionAgeSeconds:        600,
			AllowAutomaticFallback:       true,
			RequireDecisionForSubmission: true,
			AutoRefreshStaleDecisions:    true,
			ViolationAlertThreshold:      5,
		},
	}
}

// Validate validates the HPC configuration
func (c *HPCConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if !c.SchedulerType.IsValid() {
		return errors.New("invalid scheduler type")
	}

	if c.ClusterID == "" {
		return errors.New("cluster_id is required when HPC is enabled")
	}

	if c.JobService.JobPollInterval < time.Second {
		return errors.New("job_poll_interval must be at least 1 second")
	}

	if c.JobService.MaxConcurrentJobs < 1 {
		return errors.New("max_concurrent_jobs must be at least 1")
	}

	if c.UsageReporting.Enabled && c.UsageReporting.ReportInterval < time.Minute {
		return errors.New("report_interval must be at least 1 minute")
	}

	if c.Retry.MaxRetries < 0 {
		return errors.New("max_retries cannot be negative")
	}

	if c.Retry.BackoffMultiplier < 1.0 {
		return errors.New("backoff_multiplier must be at least 1.0")
	}

	if err := c.NodeAggregator.Validate(); err != nil {
		return err
	}

	if err := c.SlurmK8s.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSchedulerConfig returns the configuration for the selected scheduler
func (c *HPCConfig) GetSchedulerConfig() interface{} {
	switch c.SchedulerType {
	case HPCSchedulerTypeSLURM:
		return c.SLURM
	case HPCSchedulerTypeMOAB:
		return c.MOAB
	case HPCSchedulerTypeOOD:
		return c.OOD
	default:
		return nil
	}
}
