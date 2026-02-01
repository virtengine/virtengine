// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-14B: SLURM Integration Service - wires SLURM adapter into provider daemon bid engine
package provider_daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/slurm_adapter"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// SLURMIntegrationConfig configures the SLURM integration service
type SLURMIntegrationConfig struct {
	// Enabled enables SLURM integration
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ClusterID is the on-chain HPC cluster ID
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`

	// ProviderAddress is the provider's bech32 address
	ProviderAddress string `json:"provider_address" yaml:"provider_address"`

	// SSHConfig is the SSH configuration for SLURM access
	SSHConfig slurm_adapter.SSHConfig `json:"ssh" yaml:"ssh"`

	// SLURMConfig is the SLURM-specific configuration
	SLURMConfig slurm_adapter.SLURMConfig `json:"slurm" yaml:"slurm"`

	// AutoSubmitOnLease enables automatic job submission when a lease is created
	AutoSubmitOnLease bool `json:"auto_submit_on_lease" yaml:"auto_submit_on_lease"`

	// JobPollInterval is how often to poll for job status updates
	JobPollInterval time.Duration `json:"job_poll_interval" yaml:"job_poll_interval"`

	// StatusReportInterval is how often to report job status on-chain
	StatusReportInterval time.Duration `json:"status_report_interval" yaml:"status_report_interval"`

	// UsageReportInterval is how often to report usage metrics
	UsageReportInterval time.Duration `json:"usage_report_interval" yaml:"usage_report_interval"`

	// MaxConcurrentJobs limits the number of concurrent SLURM jobs
	MaxConcurrentJobs int `json:"max_concurrent_jobs" yaml:"max_concurrent_jobs"`

	// RetryConfig configures retry behavior
	RetryConfig SLURMRetryConfig `json:"retry" yaml:"retry"`
}

// SLURMRetryConfig configures retry behavior for SLURM operations
type SLURMRetryConfig struct {
	MaxRetries        int           `json:"max_retries" yaml:"max_retries"`
	InitialBackoff    time.Duration `json:"initial_backoff" yaml:"initial_backoff"`
	MaxBackoff        time.Duration `json:"max_backoff" yaml:"max_backoff"`
	BackoffMultiplier float64       `json:"backoff_multiplier" yaml:"backoff_multiplier"`
}

// DefaultSLURMIntegrationConfig returns the default configuration
func DefaultSLURMIntegrationConfig() SLURMIntegrationConfig {
	return SLURMIntegrationConfig{
		Enabled:              false,
		AutoSubmitOnLease:    true,
		JobPollInterval:      15 * time.Second,
		StatusReportInterval: 30 * time.Second,
		UsageReportInterval:  5 * time.Minute,
		MaxConcurrentJobs:    100,
		SSHConfig:            slurm_adapter.DefaultSSHConfig(),
		SLURMConfig:          slurm_adapter.DefaultSLURMConfig(),
		RetryConfig: SLURMRetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Second,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
}

// Validate validates the SLURM integration configuration
func (c *SLURMIntegrationConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.ClusterID == "" {
		return fmt.Errorf("cluster_id is required when SLURM integration is enabled")
	}

	if c.ProviderAddress == "" {
		return fmt.Errorf("provider_address is required when SLURM integration is enabled")
	}

	if c.SSHConfig.Host == "" {
		return fmt.Errorf("ssh.host is required for SLURM integration")
	}

	if c.SSHConfig.User == "" {
		return fmt.Errorf("ssh.user is required for SLURM integration")
	}

	if c.JobPollInterval < time.Second {
		return fmt.Errorf("job_poll_interval must be at least 1 second")
	}

	if c.MaxConcurrentJobs < 1 {
		return fmt.Errorf("max_concurrent_jobs must be at least 1")
	}

	return nil
}

// LeaseHandler handles lease lifecycle events
type LeaseHandler interface {
	// OnLeaseCreated is called when a new lease is created
	OnLeaseCreated(ctx context.Context, lease *LeaseInfo) error

	// OnLeaseTerminated is called when a lease is terminated
	OnLeaseTerminated(ctx context.Context, leaseID string) error
}

// LeaseInfo contains lease information for job submission
type LeaseInfo struct {
	LeaseID         string                 `json:"lease_id"`
	OrderID         string                 `json:"order_id"`
	ProviderAddress string                 `json:"provider_address"`
	CustomerAddress string                 `json:"customer_address"`
	OfferingID      string                 `json:"offering_id"`
	ClusterID       string                 `json:"cluster_id"`
	JobSpec         *hpctypes.HPCJob       `json:"job_spec,omitempty"`
	Resources       *hpctypes.JobResources `json:"resources,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// SLURMIntegrationService integrates SLURM adapter with the provider daemon
type SLURMIntegrationService struct {
	config        SLURMIntegrationConfig
	credManager   *HPCCredentialManager
	slurmAdapter  *slurm_adapter.SLURMAdapter
	scheduler     HPCScheduler
	reporter      HPCOnChainReporter
	usageReporter *HPCUsageReporter

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup

	// Tracking
	leaseToJob    map[string]string // lease ID -> job ID
	jobToLease    map[string]string // job ID -> lease ID
	pendingLeases map[string]*LeaseInfo
	activeJobs    map[string]*HPCSchedulerJob

	// Lifecycle callbacks
	callbacks []SLURMIntegrationCallback
}

// SLURMIntegrationCallback is called on integration events
type SLURMIntegrationCallback func(event SLURMIntegrationEvent)

// SLURMIntegrationEvent represents an integration event
type SLURMIntegrationEvent struct {
	Type      SLURMIntegrationEventType `json:"type"`
	LeaseID   string                    `json:"lease_id,omitempty"`
	JobID     string                    `json:"job_id,omitempty"`
	State     HPCJobState               `json:"state,omitempty"`
	Message   string                    `json:"message,omitempty"`
	Timestamp time.Time                 `json:"timestamp"`
	Error     error                     `json:"-"`
}

// SLURMIntegrationEventType represents the type of integration event
type SLURMIntegrationEventType string

const (
	SLURMEventLeaseReceived     SLURMIntegrationEventType = "lease_received"
	SLURMEventJobSubmitted      SLURMIntegrationEventType = "job_submitted"
	SLURMEventJobStarted        SLURMIntegrationEventType = "job_started"
	SLURMEventJobCompleted      SLURMIntegrationEventType = "job_completed"
	SLURMEventJobFailed         SLURMIntegrationEventType = "job_failed"
	SLURMEventJobCancelled      SLURMIntegrationEventType = "job_cancelled"
	SLURMEventStatusReported    SLURMIntegrationEventType = "status_reported"
	SLURMEventUsageReported     SLURMIntegrationEventType = "usage_reported"
	SLURMEventConnectionLost    SLURMIntegrationEventType = "connection_lost"
	SLURMEventConnectionRestore SLURMIntegrationEventType = "connection_restored"
	SLURMEventError             SLURMIntegrationEventType = "error"
)

// NewSLURMIntegrationService creates a new SLURM integration service
func NewSLURMIntegrationService(
	config SLURMIntegrationConfig,
	credManager *HPCCredentialManager,
	reporter HPCOnChainReporter,
) (*SLURMIntegrationService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	service := &SLURMIntegrationService{
		config:        config,
		credManager:   credManager,
		reporter:      reporter,
		stopCh:        make(chan struct{}),
		leaseToJob:    make(map[string]string),
		jobToLease:    make(map[string]string),
		pendingLeases: make(map[string]*LeaseInfo),
		activeJobs:    make(map[string]*HPCSchedulerJob),
		callbacks:     make([]SLURMIntegrationCallback, 0),
	}

	return service, nil
}

// Start starts the SLURM integration service
func (s *SLURMIntegrationService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	// Initialize SLURM adapter with credentials
	if err := s.initializeSLURMAdapter(ctx); err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("failed to initialize SLURM adapter: %w", err)
	}

	// Create scheduler wrapper
	signer := &integrationSigner{
		credManager: s.credManager,
		address:     s.config.ProviderAddress,
	}
	s.scheduler = NewSLURMSchedulerWrapper(s.slurmAdapter, signer, s.config.ClusterID)

	// Register lifecycle callback
	s.scheduler.RegisterLifecycleCallback(s.handleJobLifecycle)

	// Start scheduler
	if err := s.scheduler.Start(ctx); err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Initialize usage reporter
	usageConfig := HPCUsageReportingConfig{
		Enabled:        true,
		ReportInterval: s.config.UsageReportInterval,
		BatchSize:      50,
		RetryOnFailure: true,
	}
	s.usageReporter = NewHPCUsageReporter(usageConfig, s.config.ClusterID, signer)

	// Start background workers
	s.wg.Add(3)
	go s.jobStatusPollerLoop()
	go s.statusReportingLoop()
	go s.usageReportingLoop()

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventConnectionRestore,
		Message:   "SLURM integration service started",
		Timestamp: time.Now(),
	})

	return nil
}

// Stop stops the SLURM integration service
func (s *SLURMIntegrationService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	// Wait for background workers
	s.wg.Wait()

	// Stop scheduler
	if s.scheduler != nil {
		if err := s.scheduler.Stop(); err != nil {
			return fmt.Errorf("failed to stop scheduler: %w", err)
		}
	}

	// Stop usage reporter
	if s.usageReporter != nil {
		if err := s.usageReporter.Stop(); err != nil {
			return fmt.Errorf("failed to stop usage reporter: %w", err)
		}
	}

	return nil
}

// IsRunning returns true if the service is running
func (s *SLURMIntegrationService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// OnLeaseCreated implements LeaseHandler
func (s *SLURMIntegrationService) OnLeaseCreated(ctx context.Context, lease *LeaseInfo) error {
	if !s.IsRunning() {
		return fmt.Errorf("SLURM integration service not running")
	}

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventLeaseReceived,
		LeaseID:   lease.LeaseID,
		Message:   fmt.Sprintf("Received lease for order %s", lease.OrderID),
		Timestamp: time.Now(),
	})

	// Check concurrent job limit
	s.mu.RLock()
	activeCount := len(s.activeJobs)
	s.mu.RUnlock()

	if activeCount >= s.config.MaxConcurrentJobs {
		return fmt.Errorf("maximum concurrent jobs (%d) reached", s.config.MaxConcurrentJobs)
	}

	// Store pending lease
	s.mu.Lock()
	s.pendingLeases[lease.LeaseID] = lease
	s.mu.Unlock()

	// Submit job if auto-submit is enabled and job spec is provided
	if s.config.AutoSubmitOnLease && lease.JobSpec != nil {
		return s.submitJobForLease(ctx, lease)
	}

	return nil
}

// OnLeaseTerminated implements LeaseHandler
func (s *SLURMIntegrationService) OnLeaseTerminated(ctx context.Context, leaseID string) error {
	if !s.IsRunning() {
		return fmt.Errorf("SLURM integration service not running")
	}

	s.mu.RLock()
	jobID, exists := s.leaseToJob[leaseID]
	s.mu.RUnlock()

	if !exists {
		// Lease may not have had a job submitted yet
		s.mu.Lock()
		delete(s.pendingLeases, leaseID)
		s.mu.Unlock()
		return nil
	}

	// Cancel the job
	if err := s.scheduler.CancelJob(ctx, jobID); err != nil {
		return fmt.Errorf("failed to cancel job %s: %w", jobID, err)
	}

	// Clean up mappings
	s.mu.Lock()
	delete(s.leaseToJob, leaseID)
	delete(s.jobToLease, jobID)
	delete(s.activeJobs, jobID)
	s.mu.Unlock()

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventJobCancelled,
		LeaseID:   leaseID,
		JobID:     jobID,
		Message:   "Job cancelled due to lease termination",
		Timestamp: time.Now(),
	})

	return nil
}

// SubmitJob submits a job for an existing lease
func (s *SLURMIntegrationService) SubmitJob(ctx context.Context, leaseID string, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	if !s.IsRunning() {
		return nil, fmt.Errorf("SLURM integration service not running")
	}

	s.mu.RLock()
	lease, exists := s.pendingLeases[leaseID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("lease %s not found", leaseID)
	}

	lease.JobSpec = job
	return s.submitJobForLeaseInternal(ctx, lease)
}

// GetJobStatus returns the status of a job
func (s *SLURMIntegrationService) GetJobStatus(ctx context.Context, jobID string) (*HPCSchedulerJob, error) {
	if s.scheduler == nil {
		return nil, fmt.Errorf("scheduler not initialized")
	}
	return s.scheduler.GetJobStatus(ctx, jobID)
}

// GetJobByLease returns the job for a lease
func (s *SLURMIntegrationService) GetJobByLease(leaseID string) (*HPCSchedulerJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobID, exists := s.leaseToJob[leaseID]
	if !exists {
		return nil, fmt.Errorf("no job found for lease %s", leaseID)
	}

	job, exists := s.activeJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found in active jobs", jobID)
	}

	return job, nil
}

// ListActiveJobs returns all active jobs
func (s *SLURMIntegrationService) ListActiveJobs() []*HPCSchedulerJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*HPCSchedulerJob, 0, len(s.activeJobs))
	for _, job := range s.activeJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// RegisterCallback registers a callback for integration events
func (s *SLURMIntegrationService) RegisterCallback(cb SLURMIntegrationCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks = append(s.callbacks, cb)
}

// GetActiveJobCount returns the number of active jobs
func (s *SLURMIntegrationService) GetActiveJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeJobs)
}

// GetPendingLeaseCount returns the number of pending leases
func (s *SLURMIntegrationService) GetPendingLeaseCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingLeases)
}

// Internal methods

func (s *SLURMIntegrationService) initializeSLURMAdapter(ctx context.Context) error {
	// Get credentials from credential manager
	creds, err := s.credManager.GetCredentials(ctx, s.config.ClusterID, CredentialTypeSLURM)
	if err != nil {
		return fmt.Errorf("failed to get SLURM credentials: %w", err)
	}

	// Apply credentials to SSH config
	sshConfig := s.config.SSHConfig
	if creds.SSHPrivateKey != "" {
		sshConfig.PrivateKey = creds.SSHPrivateKey
	} else if creds.SSHPrivateKeyPath != "" {
		sshConfig.PrivateKeyPath = creds.SSHPrivateKeyPath
	}
	if creds.Password != "" {
		sshConfig.Password = creds.Password
	}

	// Create SSH client
	sshClient, err := slurm_adapter.NewSSHSLURMClient(
		sshConfig,
		s.config.SLURMConfig.ClusterName,
		s.config.SLURMConfig.DefaultPartition,
	)
	if err != nil {
		return fmt.Errorf("failed to create SSH SLURM client: %w", err)
	}

	// Create job signer
	signer := &adapterJobSigner{
		credManager: s.credManager,
		address:     s.config.ProviderAddress,
	}

	// Create SLURM adapter
	s.slurmAdapter = slurm_adapter.NewSLURMAdapter(s.config.SLURMConfig, sshClient, signer)

	return nil
}

func (s *SLURMIntegrationService) submitJobForLease(ctx context.Context, lease *LeaseInfo) error {
	_, err := s.submitJobForLeaseInternal(ctx, lease)
	return err
}

func (s *SLURMIntegrationService) submitJobForLeaseInternal(ctx context.Context, lease *LeaseInfo) (*HPCSchedulerJob, error) {
	if lease.JobSpec == nil {
		return nil, fmt.Errorf("no job spec provided for lease %s", lease.LeaseID)
	}

	// Enrich job spec with lease info
	job := lease.JobSpec
	job.ClusterID = s.config.ClusterID
	job.ProviderAddress = s.config.ProviderAddress
	job.CustomerAddress = lease.CustomerAddress

	// Submit to scheduler
	schedulerJob, err := s.scheduler.SubmitJob(ctx, job)
	if err != nil {
		s.emitEvent(SLURMIntegrationEvent{
			Type:      SLURMEventError,
			LeaseID:   lease.LeaseID,
			Message:   fmt.Sprintf("Failed to submit job: %v", err),
			Error:     err,
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("failed to submit job: %w", err)
	}

	// Track mappings
	s.mu.Lock()
	s.leaseToJob[lease.LeaseID] = job.JobID
	s.jobToLease[job.JobID] = lease.LeaseID
	s.activeJobs[job.JobID] = schedulerJob
	delete(s.pendingLeases, lease.LeaseID)
	s.mu.Unlock()

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventJobSubmitted,
		LeaseID:   lease.LeaseID,
		JobID:     job.JobID,
		State:     schedulerJob.State,
		Message:   fmt.Sprintf("Job submitted with SLURM ID %s", schedulerJob.SchedulerJobID),
		Timestamp: time.Now(),
	})

	// Report initial status on-chain
	go s.reportJobStatus(context.Background(), schedulerJob)

	return schedulerJob, nil
}

func (s *SLURMIntegrationService) handleJobLifecycle(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	s.mu.Lock()
	if existingJob, exists := s.activeJobs[job.VirtEngineJobID]; exists {
		// Update cached job
		existingJob.State = job.State
		existingJob.StateMessage = job.StateMessage
		existingJob.ExitCode = job.ExitCode
		existingJob.StartTime = job.StartTime
		existingJob.EndTime = job.EndTime
		existingJob.Metrics = job.Metrics
		existingJob.NodeList = job.NodeList
	}
	leaseID := s.jobToLease[job.VirtEngineJobID]
	s.mu.Unlock()

	// Emit appropriate event
	var eventType SLURMIntegrationEventType
	switch event {
	case HPCJobEventStarted:
		eventType = SLURMEventJobStarted
	case HPCJobEventCompleted:
		eventType = SLURMEventJobCompleted
	case HPCJobEventFailed:
		eventType = SLURMEventJobFailed
	case HPCJobEventCancelled:
		eventType = SLURMEventJobCancelled
	default:
		return // Don't emit for other events
	}

	s.emitEvent(SLURMIntegrationEvent{
		Type:      eventType,
		LeaseID:   leaseID,
		JobID:     job.VirtEngineJobID,
		State:     job.State,
		Message:   job.StateMessage,
		Timestamp: time.Now(),
	})

	// Report status on-chain
	go s.reportJobStatus(context.Background(), job)

	// Handle terminal states
	if job.State.IsTerminal() {
		// Report final usage
		go s.reportFinalUsage(context.Background(), job, leaseID)

		// Clean up after a delay (allow final reports to complete)
		go func() {
			time.Sleep(5 * time.Second)
			s.mu.Lock()
			delete(s.activeJobs, job.VirtEngineJobID)
			delete(s.leaseToJob, leaseID)
			delete(s.jobToLease, job.VirtEngineJobID)
			s.mu.Unlock()
		}()
	}
}

func (s *SLURMIntegrationService) jobStatusPollerLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.JobPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.pollActiveJobs()
		}
	}
}

func (s *SLURMIntegrationService) pollActiveJobs() {
	s.mu.RLock()
	jobIDs := make([]string, 0, len(s.activeJobs))
	for id := range s.activeJobs {
		jobIDs = append(jobIDs, id)
	}
	s.mu.RUnlock()

	ctx := context.Background()
	for _, jobID := range jobIDs {
		_, _ = s.scheduler.GetJobStatus(ctx, jobID)
	}
}

func (s *SLURMIntegrationService) statusReportingLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.StatusReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.reportAllJobStatuses()
		}
	}
}

func (s *SLURMIntegrationService) reportAllJobStatuses() {
	s.mu.RLock()
	jobs := make([]*HPCSchedulerJob, 0, len(s.activeJobs))
	for _, job := range s.activeJobs {
		jobs = append(jobs, job)
	}
	s.mu.RUnlock()

	ctx := context.Background()
	for _, job := range jobs {
		s.reportJobStatus(ctx, job)
	}
}

func (s *SLURMIntegrationService) reportJobStatus(ctx context.Context, job *HPCSchedulerJob) {
	if s.reporter == nil || s.scheduler == nil {
		return
	}

	report, err := s.scheduler.CreateStatusReport(job)
	if err != nil {
		s.emitEvent(SLURMIntegrationEvent{
			Type:      SLURMEventError,
			JobID:     job.VirtEngineJobID,
			Message:   fmt.Sprintf("Failed to create status report: %v", err),
			Error:     err,
			Timestamp: time.Now(),
		})
		return
	}

	if err := s.reporter.ReportJobStatus(ctx, report); err != nil {
		s.emitEvent(SLURMIntegrationEvent{
			Type:      SLURMEventError,
			JobID:     job.VirtEngineJobID,
			Message:   fmt.Sprintf("Failed to report status: %v", err),
			Error:     err,
			Timestamp: time.Now(),
		})
		return
	}

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventStatusReported,
		JobID:     job.VirtEngineJobID,
		State:     job.State,
		Timestamp: time.Now(),
	})
}

func (s *SLURMIntegrationService) usageReportingLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.UsageReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.reportActiveJobsUsage()
		}
	}
}

func (s *SLURMIntegrationService) reportActiveJobsUsage() {
	if s.reporter == nil || s.scheduler == nil {
		return
	}

	s.mu.RLock()
	jobs := make([]*HPCSchedulerJob, 0, len(s.activeJobs))
	leaseMap := make(map[string]string)
	for jobID, job := range s.activeJobs {
		jobs = append(jobs, job)
		leaseMap[jobID] = s.jobToLease[jobID]
	}
	s.mu.RUnlock()

	ctx := context.Background()
	for _, job := range jobs {
		metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
		if err != nil || metrics == nil {
			continue
		}

		if err := s.reporter.ReportJobAccounting(ctx, job.VirtEngineJobID, metrics); err != nil {
			continue
		}

		s.emitEvent(SLURMIntegrationEvent{
			Type:      SLURMEventUsageReported,
			LeaseID:   leaseMap[job.VirtEngineJobID],
			JobID:     job.VirtEngineJobID,
			Timestamp: time.Now(),
		})
	}
}

func (s *SLURMIntegrationService) reportFinalUsage(ctx context.Context, job *HPCSchedulerJob, leaseID string) {
	if s.reporter == nil || s.scheduler == nil || s.usageReporter == nil {
		return
	}

	// Get final accounting
	metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
	if err != nil || metrics == nil {
		return
	}

	// Create final usage record
	record, err := s.usageReporter.CreateUsageRecord(
		job,
		job.OriginalJob.CustomerAddress,
		job.SubmitTime,
		time.Now(),
		true, // isFinal
	)
	if err != nil {
		return
	}

	// Report to chain
	if err := s.reporter.ReportJobAccounting(ctx, job.VirtEngineJobID, metrics); err != nil {
		return
	}

	// Queue for settlement
	s.usageReporter.QueueRecord(record)

	s.emitEvent(SLURMIntegrationEvent{
		Type:      SLURMEventUsageReported,
		LeaseID:   leaseID,
		JobID:     job.VirtEngineJobID,
		Message:   "Final usage reported",
		Timestamp: time.Now(),
	})
}

func (s *SLURMIntegrationService) emitEvent(event SLURMIntegrationEvent) {
	s.mu.RLock()
	callbacks := make([]SLURMIntegrationCallback, len(s.callbacks))
	copy(callbacks, s.callbacks)
	s.mu.RUnlock()

	for _, cb := range callbacks {
		cb(event)
	}
}

// integrationSigner implements HPCSchedulerSigner for the integration service
type integrationSigner struct {
	credManager *HPCCredentialManager
	address     string
}

func (s *integrationSigner) Sign(data []byte) ([]byte, error) {
	return s.credManager.Sign(data)
}

func (s *integrationSigner) GetProviderAddress() string {
	return s.address
}

// adapterJobSigner implements slurm_adapter.JobSigner
type adapterJobSigner struct {
	credManager *HPCCredentialManager
	address     string
}

func (s *adapterJobSigner) Sign(data []byte) ([]byte, error) {
	return s.credManager.Sign(data)
}

func (s *adapterJobSigner) Verify(data []byte, signature []byte) bool {
	return s.credManager.Verify(data, signature)
}

func (s *adapterJobSigner) GetProviderAddress() string {
	return s.address
}

