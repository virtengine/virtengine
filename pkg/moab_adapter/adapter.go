// Package moab_adapter implements the MOAB workload manager adapter for VirtEngine.
//
// VE-917: MOAB workload manager using Waldur
package moab_adapter

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	jobStatusFailed    = "failed"
	jobStatusCancelled = "cancelled"
	jobStatusCompleted = "completed"
)

// JobSigner signs job status updates
type JobSigner interface {
	// Sign signs data and returns the signature
	Sign(data []byte) ([]byte, error)

	// Verify verifies a signature
	Verify(data []byte, signature []byte) bool

	// GetProviderAddress returns the provider address
	GetProviderAddress() string
}

// OnChainReporter reports job status to the blockchain
type OnChainReporter interface {
	// ReportJobStatus reports job status on-chain
	ReportJobStatus(ctx context.Context, report *JobStatusReport) error
}

// RewardsIntegration handles VE rewards integration
type RewardsIntegration interface {
	// RecordJobCompletion records job completion for rewards
	RecordJobCompletion(ctx context.Context, data *VERewardsData) error

	// GetRewardRate gets the current reward rate for HPC jobs
	GetRewardRate(ctx context.Context, schedulerType string) (float64, error)
}

// JobLifecycleCallback is called during job lifecycle events
type JobLifecycleCallback func(job *MOABJob, event JobLifecycleEvent)

// JobLifecycleEvent represents a job lifecycle event
type JobLifecycleEvent string

const (
	// JobEventSubmitted is fired when job is submitted
	JobEventSubmitted JobLifecycleEvent = "submitted"

	// JobEventStarted is fired when job starts running
	JobEventStarted JobLifecycleEvent = "started"

	// JobEventCompleted is fired when job completes
	JobEventCompleted JobLifecycleEvent = "completed"

	// JobEventFailed is fired when job fails
	JobEventFailed JobLifecycleEvent = "failed"

	// JobEventCancelled is fired when job is cancelled
	JobEventCancelled JobLifecycleEvent = "cancelled"

	// JobEventHeld is fired when job is put on hold
	JobEventHeld JobLifecycleEvent = "held"

	// JobEventReleased is fired when job is released from hold
	JobEventReleased JobLifecycleEvent = "released"
)

// MOABAdapter implements the MOAB workload manager adapter
type MOABAdapter struct {
	config             MOABConfig
	client             MOABClient
	signer             JobSigner
	rewardsIntegration RewardsIntegration
	mu                 sync.RWMutex
	jobs               map[string]*MOABJob // MOAB job ID -> job
	jobMapping         map[string]string   // VirtEngine job ID -> MOAB job ID
	running            bool
	stopCh             chan struct{}
	callbacks          []JobLifecycleCallback
	clusterID          string
}

// NewMOABAdapter creates a new MOAB adapter
func NewMOABAdapter(config MOABConfig, client MOABClient, signer JobSigner) *MOABAdapter {
	return &MOABAdapter{
		config:     config,
		client:     client,
		signer:     signer,
		jobs:       make(map[string]*MOABJob),
		jobMapping: make(map[string]string),
		stopCh:     make(chan struct{}),
		callbacks:  make([]JobLifecycleCallback, 0),
	}
}

// SetRewardsIntegration sets the rewards integration handler
func (a *MOABAdapter) SetRewardsIntegration(ri RewardsIntegration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rewardsIntegration = ri
}

// SetClusterID sets the cluster ID for rewards tracking
func (a *MOABAdapter) SetClusterID(clusterID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clusterID = clusterID
}

// RegisterLifecycleCallback registers a callback for job lifecycle events
func (a *MOABAdapter) RegisterLifecycleCallback(cb JobLifecycleCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callbacks = append(a.callbacks, cb)
}

// notifyCallbacks notifies all registered callbacks
func (a *MOABAdapter) notifyCallbacks(job *MOABJob, event JobLifecycleEvent) {
	a.mu.RLock()
	callbacks := make([]JobLifecycleCallback, len(a.callbacks))
	copy(callbacks, a.callbacks)
	a.mu.RUnlock()

	for _, cb := range callbacks {
		cb(job, event)
	}
}

// Start starts the MOAB adapter
func (a *MOABAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.stopCh = make(chan struct{})
	a.mu.Unlock()

	// Connect to MOAB
	if err := a.client.Connect(ctx); err != nil {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
		return fmt.Errorf("failed to connect to MOAB: %w", err)
	}

	// Start job polling
	go a.pollJobs()

	return nil
}

// Stop stops the MOAB adapter
func (a *MOABAdapter) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = false
	close(a.stopCh)
	a.mu.Unlock()

	return a.client.Disconnect()
}

// IsRunning checks if the adapter is running
func (a *MOABAdapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// ClusterName returns the cluster name from config
func (a *MOABAdapter) ClusterName() string {
	return a.config.DefaultQueue
}

// SubmitJob submits a job to MOAB using msub
func (a *MOABAdapter) SubmitJob(ctx context.Context, virtEngineJobID string, spec *MOABJobSpec) (*MOABJob, error) {
	if !a.IsRunning() {
		return nil, ErrMOABNotConnected
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJobSpec, err)
	}

	// Set defaults from config if not specified
	if spec.Queue == "" {
		spec.Queue = a.config.DefaultQueue
	}
	if spec.Account == "" {
		spec.Account = a.config.DefaultAccount
	}

	// Submit to MOAB (msub equivalent)
	moabJobID, err := a.client.SubmitJob(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJobSubmissionFailed, err)
	}

	// Create job record
	job := &MOABJob{
		MOABJobID:       moabJobID,
		VirtEngineJobID: virtEngineJobID,
		Spec:            spec,
		State:           MOABJobStateIdle,
		SubmitTime:      time.Now(),
	}

	// Store job
	a.mu.Lock()
	a.jobs[moabJobID] = job
	a.jobMapping[virtEngineJobID] = moabJobID
	a.mu.Unlock()

	// Notify callbacks
	a.notifyCallbacks(job, JobEventSubmitted)

	return job, nil
}

// CancelJob cancels a job using mjobctl -c
func (a *MOABAdapter) CancelJob(ctx context.Context, virtEngineJobID string) error {
	a.mu.RLock()
	moabJobID, exists := a.jobMapping[virtEngineJobID]
	a.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	if err := a.client.CancelJob(ctx, moabJobID); err != nil {
		return fmt.Errorf("%w: %v", ErrJobCancellationFailed, err)
	}

	// Update local state
	a.mu.Lock()
	if job, exists := a.jobs[moabJobID]; exists {
		job.State = MOABJobStateCancelled
		job.StatusMessage = "Cancelled by user"
		now := time.Now()
		job.EndTime = &now

		// Notify callbacks
		go a.notifyCallbacks(job, JobEventCancelled)
	}
	a.mu.Unlock()

	return nil
}

// HoldJob puts a job on hold using mjobctl -h
func (a *MOABAdapter) HoldJob(ctx context.Context, virtEngineJobID string) error {
	a.mu.RLock()
	moabJobID, exists := a.jobMapping[virtEngineJobID]
	a.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	if err := a.client.HoldJob(ctx, moabJobID); err != nil {
		return fmt.Errorf("hold job failed: %w", err)
	}

	// Update local state
	a.mu.Lock()
	if job, exists := a.jobs[moabJobID]; exists {
		job.State = MOABJobStateHold
		job.StatusMessage = "Job held by user"

		// Notify callbacks
		go a.notifyCallbacks(job, JobEventHeld)
	}
	a.mu.Unlock()

	return nil
}

// ReleaseJob releases a held job using mjobctl -u
func (a *MOABAdapter) ReleaseJob(ctx context.Context, virtEngineJobID string) error {
	a.mu.RLock()
	moabJobID, exists := a.jobMapping[virtEngineJobID]
	a.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	if err := a.client.ReleaseJob(ctx, moabJobID); err != nil {
		return fmt.Errorf("release job failed: %w", err)
	}

	// Update local state
	a.mu.Lock()
	if job, exists := a.jobs[moabJobID]; exists {
		job.State = MOABJobStateIdle
		job.StatusMessage = "Job released from hold"

		// Notify callbacks
		go a.notifyCallbacks(job, JobEventReleased)
	}
	a.mu.Unlock()

	return nil
}

// GetJobStatus gets job status using checkjob
func (a *MOABAdapter) GetJobStatus(ctx context.Context, virtEngineJobID string) (*MOABJob, error) {
	a.mu.RLock()
	moabJobID, exists := a.jobMapping[virtEngineJobID]
	if !exists {
		a.mu.RUnlock()
		return nil, ErrJobNotFound
	}
	job, exists := a.jobs[moabJobID]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrJobNotFound
	}

	// Get latest status from MOAB
	updatedJob, err := a.client.GetJobStatus(ctx, moabJobID)
	if err != nil {
		// Return cached status if MOAB query fails
		return job, nil
	}

	// Preserve VirtEngineJobID
	updatedJob.VirtEngineJobID = virtEngineJobID

	// Fetch accounting data for completed jobs
	if isTerminalState(updatedJob.State) && updatedJob.UsageMetrics == nil {
		metrics, err := a.client.GetJobAccounting(ctx, moabJobID)
		if err == nil {
			updatedJob.UsageMetrics = metrics
		}
	}

	// Update cached job
	a.mu.Lock()
	a.jobs[moabJobID] = updatedJob
	a.mu.Unlock()

	return updatedJob, nil
}

// GetJobsByQueue gets all jobs in a specific queue
func (a *MOABAdapter) GetJobsByQueue(queue string) []*MOABJob {
	a.mu.RLock()
	defer a.mu.RUnlock()

	jobs := make([]*MOABJob, 0)
	for _, job := range a.jobs {
		if job.Spec != nil && job.Spec.Queue == queue {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// GetAllJobs gets all tracked jobs
func (a *MOABAdapter) GetAllJobs() []*MOABJob {
	a.mu.RLock()
	defer a.mu.RUnlock()

	jobs := make([]*MOABJob, 0, len(a.jobs))
	for _, job := range a.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListQueues lists available queues using mdiag -q
func (a *MOABAdapter) ListQueues(ctx context.Context) ([]QueueInfo, error) {
	if !a.IsRunning() {
		return nil, ErrMOABNotConnected
	}
	return a.client.ListQueues(ctx)
}

// ListNodes lists nodes using mdiag -n
func (a *MOABAdapter) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	if !a.IsRunning() {
		return nil, ErrMOABNotConnected
	}
	return a.client.ListNodes(ctx)
}

// GetClusterInfo gets cluster information
func (a *MOABAdapter) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	if !a.IsRunning() {
		return nil, ErrMOABNotConnected
	}
	return a.client.GetClusterInfo(ctx)
}

// CreateStatusReport creates a signed status report for on-chain submission
func (a *MOABAdapter) CreateStatusReport(job *MOABJob) (*JobStatusReport, error) {
	report := &JobStatusReport{
		ProviderAddress: a.signer.GetProviderAddress(),
		VirtEngineJobID: job.VirtEngineJobID,
		MOABJobID:       job.MOABJobID,
		State:           job.State,
		StatusMessage:   job.StatusMessage,
		ExitCode:        job.ExitCode,
		UsageMetrics:    job.UsageMetrics,
		Timestamp:       time.Now(),
	}

	// Sign the report
	hash := report.Hash()
	sig, err := a.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}
	report.Signature = hex.EncodeToString(sig)

	return report, nil
}

// CreateVERewardsData creates VE rewards data for a completed job
func (a *MOABAdapter) CreateVERewardsData(job *MOABJob, customerAddress string) (*VERewardsData, error) {
	if job.UsageMetrics == nil {
		return nil, fmt.Errorf("job has no usage metrics")
	}
	if job.StartTime == nil || job.EndTime == nil {
		return nil, fmt.Errorf("job timing information incomplete")
	}

	var completionStatus string
	switch job.State {
	case MOABJobStateFailed:
		completionStatus = jobStatusFailed
	case MOABJobStateCancelled:
		completionStatus = jobStatusCancelled
	default:
		completionStatus = jobStatusCompleted
	}

	a.mu.RLock()
	clusterID := a.clusterID
	a.mu.RUnlock()

	return &VERewardsData{
		JobID:            job.VirtEngineJobID,
		ClusterID:        clusterID,
		ProviderAddress:  a.signer.GetProviderAddress(),
		CustomerAddress:  customerAddress,
		SchedulerType:    "MOAB",
		Usage:            job.UsageMetrics,
		StartTime:        *job.StartTime,
		EndTime:          *job.EndTime,
		CompletionStatus: completionStatus,
	}, nil
}

// RecordJobForRewards records a completed job for VE rewards
func (a *MOABAdapter) RecordJobForRewards(ctx context.Context, job *MOABJob, customerAddress string) error {
	a.mu.RLock()
	ri := a.rewardsIntegration
	a.mu.RUnlock()

	if ri == nil {
		return fmt.Errorf("rewards integration not configured")
	}

	data, err := a.CreateVERewardsData(job, customerAddress)
	if err != nil {
		return err
	}

	return ri.RecordJobCompletion(ctx, data)
}

// pollJobs polls for job status updates
func (a *MOABAdapter) pollJobs() {
	// Use default poll interval if not set
	pollInterval := a.config.JobPollInterval
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second // Default to 15 seconds
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.updateJobStatuses()
		}
	}
}

// updateJobStatuses updates all job statuses
func (a *MOABAdapter) updateJobStatuses() {
	a.mu.RLock()
	moabJobIDs := make([]string, 0, len(a.jobs))
	previousStates := make(map[string]MOABJobState)
	for id, job := range a.jobs {
		// Only poll non-terminal jobs
		if !isTerminalState(job.State) {
			moabJobIDs = append(moabJobIDs, id)
			previousStates[id] = job.State
		}
	}
	a.mu.RUnlock()

	ctx := context.Background()
	for _, moabJobID := range moabJobIDs {
		updatedJob, err := a.client.GetJobStatus(ctx, moabJobID)
		if err != nil {
			continue
		}

		// Preserve VirtEngineJobID
		a.mu.RLock()
		if originalJob, exists := a.jobs[moabJobID]; exists {
			updatedJob.VirtEngineJobID = originalJob.VirtEngineJobID
		}
		a.mu.RUnlock()

		// Get accounting data for completed jobs
		if isTerminalState(updatedJob.State) {
			metrics, err := a.client.GetJobAccounting(ctx, moabJobID)
			if err == nil {
				updatedJob.UsageMetrics = metrics
			}
		}

		// Detect state changes for lifecycle events
		prevState := previousStates[moabJobID]
		if prevState != updatedJob.State {
			a.handleStateChange(updatedJob, prevState)
		}

		a.mu.Lock()
		a.jobs[moabJobID] = updatedJob
		a.mu.Unlock()
	}
}

// handleStateChange handles job state changes
func (a *MOABAdapter) handleStateChange(job *MOABJob, prevState MOABJobState) {
	switch job.State {
	case MOABJobStateRunning:
		if prevState == MOABJobStateIdle || prevState == MOABJobStateStarting {
			a.notifyCallbacks(job, JobEventStarted)
		}
	case MOABJobStateCompleted:
		a.notifyCallbacks(job, JobEventCompleted)
	case MOABJobStateFailed:
		a.notifyCallbacks(job, JobEventFailed)
	case MOABJobStateCancelled:
		a.notifyCallbacks(job, JobEventCancelled)
	case MOABJobStateHold:
		a.notifyCallbacks(job, JobEventHeld)
	}
}

// isTerminalState checks if a state is terminal
func isTerminalState(state MOABJobState) bool {
	switch state {
	case MOABJobStateCompleted, MOABJobStateFailed, MOABJobStateCancelled, MOABJobStateRemoved, MOABJobStateVacated:
		return true
	default:
		return false
	}
}

// MapToVirtEngineState maps MOAB state to VirtEngine job state
func MapToVirtEngineState(state MOABJobState) string {
	switch state {
	case MOABJobStateIdle:
		return "queued"
	case MOABJobStateStarting:
		return "starting"
	case MOABJobStateRunning:
		return "running"
	case MOABJobStateCompleted:
		return "completed"
	case MOABJobStateFailed:
		return "failed"
	case MOABJobStateCancelled:
		return "cancelled"
	case MOABJobStateHold:
		return "held"
	case MOABJobStateSuspended:
		return "paused"
	case MOABJobStateDeferred:
		return "deferred"
	case MOABJobStateRemoved, MOABJobStateVacated:
		return "removed"
	default:
		return "pending"
	}
}

// CalculateJobCost calculates the cost of a job based on usage metrics
func CalculateJobCost(metrics *MOABUsageMetrics, cpuRatePerHour, gpuRatePerHour, memoryRatePerGBHour float64) float64 {
	if metrics == nil {
		return 0
	}

	cost := 0.0

	// CPU cost
	cpuHours := float64(metrics.CPUTimeSeconds) / 3600.0
	cost += cpuHours * cpuRatePerHour

	// GPU cost
	gpuHours := float64(metrics.GPUSeconds) / 3600.0
	cost += gpuHours * gpuRatePerHour

	// Memory cost (using max RSS)
	memoryGB := float64(metrics.MaxRSSBytes) / (1024 * 1024 * 1024)
	memoryHours := memoryGB * float64(metrics.WallClockSeconds) / 3600.0
	cost += memoryHours * memoryRatePerGBHour

	return cost
}

