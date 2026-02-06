// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Provider - aggregates all HPC components for provider daemon
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCChainSubscriberConfig configures the on-chain event subscriber
type HPCChainSubscriberConfig struct {
	// Enabled enables chain event subscription
	Enabled bool `json:"enabled" yaml:"enabled"`

	// SubscriptionBufferSize is the buffer size for event subscription
	SubscriptionBufferSize int `json:"subscription_buffer_size" yaml:"subscription_buffer_size"`

	// ReconnectInterval is how often to retry connection on failure
	ReconnectInterval time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`

	// MaxReconnectAttempts is the maximum reconnection attempts (0 = infinite)
	MaxReconnectAttempts int `json:"max_reconnect_attempts" yaml:"max_reconnect_attempts"`
}

// DefaultHPCChainSubscriberConfig returns the default chain subscriber config
func DefaultHPCChainSubscriberConfig() HPCChainSubscriberConfig {
	return HPCChainSubscriberConfig{
		Enabled:                true,
		SubscriptionBufferSize: 100,
		ReconnectInterval:      10 * time.Second,
		MaxReconnectAttempts:   0, // Infinite
	}
}

// HPCCredentialConfig configures credential management
type HPCCredentialConfig struct {
	// CredentialManager configuration
	Manager HPCCredentialManagerConfig `json:"manager" yaml:"manager"`

	// RequireEncryption requires all credentials to be encrypted
	RequireEncryption bool `json:"require_encryption" yaml:"require_encryption"`

	// AutoRotateCredentials enables automatic credential rotation
	AutoRotateCredentials bool `json:"auto_rotate_credentials" yaml:"auto_rotate_credentials"`

	// RotationCheckInterval is how often to check for rotation needs
	RotationCheckInterval time.Duration `json:"rotation_check_interval" yaml:"rotation_check_interval"`
}

// DefaultHPCCredentialConfig returns the default credential config
func DefaultHPCCredentialConfig() HPCCredentialConfig {
	return HPCCredentialConfig{
		Manager:               DefaultHPCCredentialManagerConfig(),
		RequireEncryption:     true,
		AutoRotateCredentials: true,
		RotationCheckInterval: 24 * time.Hour,
	}
}

// HPCProviderConfig contains all configuration for the HPC provider
type HPCProviderConfig struct {
	// HPC contains core HPC configuration
	HPC HPCConfig `json:"hpc" yaml:"hpc"`

	// Chain contains chain subscriber configuration
	Chain HPCChainSubscriberConfig `json:"chain" yaml:"chain"`

	// Settlement contains settlement pipeline configuration
	Settlement HPCBatchSettlementConfig `json:"settlement" yaml:"settlement"`

	// Credentials contains credential management configuration
	Credentials HPCCredentialConfig `json:"credentials" yaml:"credentials"`

	// Accounting contains accounting service configuration
	Accounting HPCAccountingConfig `json:"accounting" yaml:"accounting"`

	// Reconciliation contains reconciliation service configuration
	Reconciliation HPCReconciliationConfig `json:"reconciliation" yaml:"reconciliation"`
}

// DefaultHPCProviderConfig returns the default HPC provider configuration
func DefaultHPCProviderConfig() HPCProviderConfig {
	return HPCProviderConfig{
		HPC:            DefaultHPCConfig(),
		Chain:          DefaultHPCChainSubscriberConfig(),
		Settlement:     DefaultHPCBatchSettlementConfig(),
		Credentials:    DefaultHPCCredentialConfig(),
		Accounting:     DefaultHPCAccountingConfig(),
		Reconciliation: DefaultHPCReconciliationConfig(),
	}
}

// Validate validates the HPC provider configuration
func (c *HPCProviderConfig) Validate() error {
	if err := c.HPC.Validate(); err != nil {
		return fmt.Errorf("invalid HPC config: %w", err)
	}

	if c.Chain.Enabled && c.Chain.SubscriptionBufferSize < 1 {
		return errors.New("subscription_buffer_size must be at least 1")
	}

	if err := c.Settlement.Validate(); err != nil {
		return fmt.Errorf("invalid settlement config: %w", err)
	}

	return nil
}

// HPCChainClient defines the interface for chain operations
type HPCChainClient interface {
	// Job event subscription
	HPCJobEventSubscriber

	// On-chain reporting
	HPCOnChainReporter

	// Accounting submission
	HPCAccountingSubmitter

	// Get current block height
	GetCurrentBlockHeight(ctx context.Context) (int64, error)
}

// HPCComponentHealth represents the health of an HPC component
type HPCComponentHealth struct {
	Name      string                 `json:"name"`
	Healthy   bool                   `json:"healthy"`
	Message   string                 `json:"message"`
	LastCheck time.Time              `json:"last_check"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HPCProviderHealth aggregates health from all HPC components
type HPCProviderHealth struct {
	Overall          bool                 `json:"overall"`
	Message          string               `json:"message"`
	LastCheck        time.Time            `json:"last_check"`
	Components       []HPCComponentHealth `json:"components"`
	ActiveJobs       int                  `json:"active_jobs"`
	PendingRecords   int                  `json:"pending_records"`
	CredentialHealth []CredentialHealth   `json:"credential_health,omitempty"`
}

// HPCProvider aggregates all HPC components for the provider daemon
type HPCProvider struct {
	config             HPCProviderConfig
	backendFactory     *HPCBackendFactory
	jobService         *HPCJobService
	chainSubscriber    *HPCChainSubscriberWithStats
	settlementPipeline *HPCBatchSettlementPipeline
	usageReporter      *HPCUsageReporter
	credManager        *HPCCredentialManager
	routingEnforcer    *RoutingEnforcer
	nodeAggregator     *HPCNodeAggregator
	slurmK8sManager    *HPCSlurmK8sManager
	auditor            HPCAuditLogger

	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewHPCProvider creates a new HPC provider with all components wired together
func NewHPCProvider(
	config HPCProviderConfig,
	chainClient HPCChainClient,
	auditor HPCAuditLogger,
) (*HPCProvider, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create credential manager
	credManager, err := NewHPCCredentialManager(config.Credentials.Manager)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential manager: %w", err)
	}

	// Create a basic signer implementation that uses the credential manager
	signer := &providerCredentialSigner{credManager: credManager}

	// Create backend factory (uses the existing HPCBackendFactory from hpc_backend_manager.go)
	backendFactory, err := NewHPCBackendFactory(config.HPC, credManager, signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend factory: %w", err)
	}

	// Get the scheduler from the backend factory
	scheduler := backendFactory.GetScheduler()

	// Create job service
	jobService := NewHPCJobService(config.HPC, scheduler, chainClient, auditor)

	// Create chain subscriber (uses existing HPCChainSubscriberWithStats from hpc_chain_subscriber.go)
	chainSubscriber, err := NewHPCChainSubscriberWithStats(
		config.Chain,
		config.HPC.ClusterID,
		signer.GetProviderAddress(),
		chainClient,
		jobService,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain subscriber: %w", err)
	}

	// Create usage reporter
	usageReporter := NewHPCUsageReporter(config.HPC.UsageReporting, config.HPC.ClusterID, signer)

	// Create settlement pipeline (uses existing HPCBatchSettlementPipeline from hpc_settlement_pipeline.go)
	settlementPipeline := NewHPCBatchSettlementPipeline(config.Settlement, chainClient, signer)

	// Create routing enforcer if enabled
	var routingEnforcer *RoutingEnforcer
	if config.HPC.Routing.Enabled {
		routingEnforcerConfig := RoutingEnforcerConfig{
			EnforcementMode:              hpctypes.RoutingEnforcementMode(config.HPC.Routing.EnforcementMode),
			MaxDecisionAgeBlocks:         config.HPC.Routing.MaxDecisionAgeBlocks,
			MaxDecisionAgeSeconds:        config.HPC.Routing.MaxDecisionAgeSeconds,
			AllowAutomaticFallback:       config.HPC.Routing.AllowAutomaticFallback,
			RequireDecisionForSubmission: config.HPC.Routing.RequireDecisionForSubmission,
			AutoRefreshStaleDecisions:    config.HPC.Routing.AutoRefreshStaleDecisions,
			ViolationAlertThreshold:      config.HPC.Routing.ViolationAlertThreshold,
		}
		routingEnforcer = NewRoutingEnforcer(routingEnforcerConfig, nil, nil, auditor)
		jobService.SetRoutingEnforcer(routingEnforcer)
	}

	var nodeAggregator *HPCNodeAggregator
	if config.HPC.NodeAggregator.Enabled {
		nodeCfg := config.HPC.NodeAggregator
		if nodeCfg.ProviderAddress == "" {
			nodeCfg.ProviderAddress = config.HPC.ProviderAddress
		}
		if nodeCfg.ClusterID == "" {
			nodeCfg.ClusterID = config.HPC.ClusterID
		}
		if nodeCfg.ChainReporter == nil {
			if reporter, ok := chainClient.(HPCNodeChainReporter); ok {
				nodeCfg.ChainReporter = reporter
			}
		}

		aggregator, err := NewHPCNodeAggregator(nodeCfg, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create node aggregator: %w", err)
		}
		nodeAggregator = aggregator
	}

	slurmManager, err := NewHPCSlurmK8sManager(
		config.HPC.SlurmK8s,
		config.HPC.ClusterID,
		config.HPC.ProviderAddress,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create slurm_k8s manager: %w", err)
	}

	return &HPCProvider{
		config:             config,
		backendFactory:     backendFactory,
		jobService:         jobService,
		chainSubscriber:    chainSubscriber,
		settlementPipeline: settlementPipeline,
		usageReporter:      usageReporter,
		credManager:        credManager,
		routingEnforcer:    routingEnforcer,
		nodeAggregator:     nodeAggregator,
		slurmK8sManager:    slurmManager,
		auditor:            auditor,
	}, nil
}

// Start starts all HPC components
func (p *HPCProvider) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	// Start components in order
	if p.slurmK8sManager != nil {
		if err := p.slurmK8sManager.Start(p.ctx); err != nil {
			return fmt.Errorf("failed to start slurm_k8s manager: %w", err)
		}
	}

	if p.nodeAggregator != nil {
		if err := p.nodeAggregator.Start(p.ctx); err != nil {
			if p.slurmK8sManager != nil {
				_ = p.slurmK8sManager.Stop()
			}
			return fmt.Errorf("failed to start node aggregator: %w", err)
		}
	}

	if err := p.backendFactory.Start(p.ctx); err != nil {
		if p.nodeAggregator != nil {
			p.nodeAggregator.Stop()
		}
		if p.slurmK8sManager != nil {
			_ = p.slurmK8sManager.Stop()
		}
		return fmt.Errorf("failed to start backend factory: %w", err)
	}

	if err := p.usageReporter.Start(); err != nil {
		if p.nodeAggregator != nil {
			p.nodeAggregator.Stop()
		}
		if p.slurmK8sManager != nil {
			_ = p.slurmK8sManager.Stop()
		}
		_ = p.backendFactory.Stop()
		return fmt.Errorf("failed to start usage reporter: %w", err)
	}

	if err := p.jobService.Start(p.ctx); err != nil {
		if p.nodeAggregator != nil {
			p.nodeAggregator.Stop()
		}
		if p.slurmK8sManager != nil {
			_ = p.slurmK8sManager.Stop()
		}
		_ = p.usageReporter.Stop()
		_ = p.backendFactory.Stop()
		return fmt.Errorf("failed to start job service: %w", err)
	}

	if p.config.Chain.Enabled {
		if err := p.chainSubscriber.Start(p.ctx); err != nil {
			if p.nodeAggregator != nil {
				p.nodeAggregator.Stop()
			}
			if p.slurmK8sManager != nil {
				_ = p.slurmK8sManager.Stop()
			}
			_ = p.jobService.Stop()
			_ = p.usageReporter.Stop()
			_ = p.backendFactory.Stop()
			return fmt.Errorf("failed to start chain subscriber: %w", err)
		}
	}

	if p.config.Settlement.Enabled {
		if err := p.settlementPipeline.Start(p.ctx); err != nil {
			if p.nodeAggregator != nil {
				p.nodeAggregator.Stop()
			}
			if p.slurmK8sManager != nil {
				_ = p.slurmK8sManager.Stop()
			}
			_ = p.chainSubscriber.Stop()
			_ = p.jobService.Stop()
			_ = p.usageReporter.Stop()
			_ = p.backendFactory.Stop()
			return fmt.Errorf("failed to start settlement pipeline: %w", err)
		}
	}

	p.mu.Lock()
	p.running = true
	p.mu.Unlock()

	p.logAuditEvent("provider_started", "", map[string]interface{}{
		"scheduler_type": string(p.config.HPC.SchedulerType),
		"cluster_id":     p.config.HPC.ClusterID,
	}, true)

	return nil
}

// Stop stops all HPC components
func (p *HPCProvider) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = false
	p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	var errs []error

	// Stop components in reverse order
	if p.config.Settlement.Enabled {
		if err := p.settlementPipeline.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("settlement pipeline: %w", err))
		}
	}

	if p.config.Chain.Enabled {
		if err := p.chainSubscriber.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("chain subscriber: %w", err))
		}
	}

	if err := p.jobService.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("job service: %w", err))
	}

	if err := p.usageReporter.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("usage reporter: %w", err))
	}

	if err := p.backendFactory.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("backend factory: %w", err))
	}

	if p.nodeAggregator != nil {
		p.nodeAggregator.Stop()
	}

	if p.slurmK8sManager != nil {
		if err := p.slurmK8sManager.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("slurm_k8s manager: %w", err))
		}
	}

	// Lock credential manager
	p.credManager.Lock()

	p.logAuditEvent("provider_stopped", "", nil, true)

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// IsRunning returns true if the provider is running
func (p *HPCProvider) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// GetHealth returns aggregated health from all components
func (p *HPCProvider) GetHealth() *HPCProviderHealth {
	p.mu.RLock()
	defer p.mu.RUnlock()

	components := []HPCComponentHealth{}

	// Add backend factory health
	backendHealth := p.backendFactory.GetHealth()
	components = append(components, HPCComponentHealth{
		Name:      "backend_factory",
		Healthy:   backendHealth.Running && backendHealth.CredentialsValid,
		Message:   backendHealth.Message,
		LastCheck: backendHealth.LastHealthCheck,
		Details: map[string]interface{}{
			"active_jobs":       backendHealth.ActiveJobs,
			"scheduler_type":    string(backendHealth.SchedulerType),
			"credentials_valid": backendHealth.CredentialsValid,
		},
	})

	if p.config.Chain.Enabled {
		components = append(components, p.chainSubscriber.GetHealth())
	}

	if p.config.Settlement.Enabled {
		stats := p.settlementPipeline.GetStats()
		components = append(components, HPCComponentHealth{
			Name:      "settlement_pipeline",
			Healthy:   p.settlementPipeline.IsRunning(),
			Message:   "running",
			LastCheck: time.Now(),
			Details: map[string]interface{}{
				"pending_count":   stats.PendingCount,
				"submitted_count": stats.SubmittedCount,
				"total_confirmed": stats.TotalConfirmed,
				"total_failed":    stats.TotalFailed,
			},
		})
	}

	if p.nodeAggregator != nil {
		components = append(components, HPCComponentHealth{
			Name:      "node_aggregator",
			Healthy:   true,
			Message:   "running",
			LastCheck: time.Now(),
			Details: map[string]interface{}{
				"node_count":      p.nodeAggregator.GetNodeCount(),
				"pending_updates": p.nodeAggregator.GetPendingUpdateCount(),
			},
		})
	}

	if p.slurmK8sManager != nil {
		components = append(components, HPCComponentHealth{
			Name:      "slurm_k8s_bootstrap",
			Healthy:   p.slurmK8sManager.IsRunning(),
			Message:   "running",
			LastCheck: time.Now(),
		})
	}

	// Add job service health
	jobServiceHealth := HPCComponentHealth{
		Name:      "job_service",
		Healthy:   p.jobService.IsRunning(),
		Message:   "running",
		LastCheck: time.Now(),
	}
	if !p.jobService.IsRunning() {
		jobServiceHealth.Message = "not running"
	}
	components = append(components, jobServiceHealth)

	// Add usage reporter health
	usageReporterHealth := HPCComponentHealth{
		Name:      "usage_reporter",
		Healthy:   true,
		Message:   "running",
		LastCheck: time.Now(),
		Details: map[string]interface{}{
			"pending_count": p.usageReporter.GetPendingCount(),
		},
	}
	components = append(components, usageReporterHealth)

	// Calculate overall health
	overall := true
	for _, c := range components {
		if !c.Healthy {
			overall = false
			break
		}
	}

	message := "healthy"
	if !overall {
		message = "degraded"
	}
	if !p.running {
		overall = false
		message = "not running"
	}

	// Get active job count
	activeJobs := 0
	if p.jobService.IsRunning() {
		jobs, err := p.jobService.ListActiveJobs(context.Background())
		if err == nil {
			activeJobs = len(jobs)
		}
	}

	return &HPCProviderHealth{
		Overall:          overall,
		Message:          message,
		LastCheck:        time.Now(),
		Components:       components,
		ActiveJobs:       activeJobs,
		PendingRecords:   p.usageReporter.GetPendingCount(),
		CredentialHealth: p.credManager.CheckHealth(),
	}
}

// SubmitJob submits a new job
func (p *HPCProvider) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	if !p.IsRunning() {
		return nil, errors.New("provider not running")
	}
	return p.jobService.SubmitJob(ctx, job)
}

// CancelJob cancels a job
func (p *HPCProvider) CancelJob(ctx context.Context, jobID string) error {
	if !p.IsRunning() {
		return errors.New("provider not running")
	}
	return p.jobService.CancelJob(ctx, jobID)
}

// GetJobStatus gets the status of a job
func (p *HPCProvider) GetJobStatus(ctx context.Context, jobID string) (*HPCSchedulerJob, error) {
	if !p.IsRunning() {
		return nil, errors.New("provider not running")
	}
	return p.jobService.GetJobStatus(ctx, jobID)
}

// ListActiveJobs lists all active jobs
func (p *HPCProvider) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	if !p.IsRunning() {
		return nil, errors.New("provider not running")
	}
	return p.jobService.ListActiveJobs(ctx)
}

// GetJobAccounting gets accounting metrics for a job
func (p *HPCProvider) GetJobAccounting(ctx context.Context, jobID string) (*HPCSchedulerMetrics, error) {
	if !p.IsRunning() {
		return nil, errors.New("provider not running")
	}
	return p.jobService.GetJobAccounting(ctx, jobID)
}

// GetCredentialManager returns the credential manager
func (p *HPCProvider) GetCredentialManager() *HPCCredentialManager {
	return p.credManager
}

// GetBackendFactory returns the backend factory
func (p *HPCProvider) GetBackendFactory() *HPCBackendFactory {
	return p.backendFactory
}

// GetJobService returns the job service
func (p *HPCProvider) GetJobService() *HPCJobService {
	return p.jobService
}

// GetUsageReporter returns the usage reporter
func (p *HPCProvider) GetUsageReporter() *HPCUsageReporter {
	return p.usageReporter
}

// GetRoutingEnforcer returns the routing enforcer
func (p *HPCProvider) GetRoutingEnforcer() *RoutingEnforcer {
	return p.routingEnforcer
}

// GetSettlementPipeline returns the settlement pipeline
func (p *HPCProvider) GetSettlementPipeline() *HPCBatchSettlementPipeline {
	return p.settlementPipeline
}

// GetChainSubscriber returns the chain subscriber
func (p *HPCProvider) GetChainSubscriber() *HPCChainSubscriberWithStats {
	return p.chainSubscriber
}

func (p *HPCProvider) logAuditEvent(eventType, jobID string, details map[string]interface{}, success bool) {
	if p.auditor == nil || !p.config.HPC.Audit.Enabled {
		return
	}

	event := HPCAuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		JobID:     jobID,
		ClusterID: p.config.HPC.ClusterID,
		Details:   details,
		Success:   success,
	}

	if p.config.HPC.Audit.LogJobEvents {
		p.auditor.LogJobEvent(event)
	}
}

// providerCredentialSigner implements HPCSchedulerSigner using HPCCredentialManager
type providerCredentialSigner struct {
	credManager     *HPCCredentialManager
	providerAddress string
}

func (s *providerCredentialSigner) Sign(data []byte) ([]byte, error) {
	return s.credManager.Sign(data)
}

func (s *providerCredentialSigner) GetProviderAddress() string {
	return s.providerAddress
}

// SetProviderAddress sets the provider address
func (s *providerCredentialSigner) SetProviderAddress(addr string) {
	s.providerAddress = addr
}
