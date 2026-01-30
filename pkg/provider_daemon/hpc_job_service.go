// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPC Job Service - manages job lifecycle and on-chain reporting
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCOnChainReporter reports job status and accounting to the blockchain
type HPCOnChainReporter interface {
	// ReportJobStatus reports job status update on-chain
	ReportJobStatus(ctx context.Context, report *HPCStatusReport) error

	// ReportJobAccounting reports job accounting/usage on-chain
	ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error
}

// HPCJobEventSubscriber subscribes to on-chain job events
type HPCJobEventSubscriber interface {
	// SubscribeToJobRequests subscribes to new job request events
	SubscribeToJobRequests(ctx context.Context, clusterID string, handler func(*hpctypes.HPCJob) error) error

	// SubscribeToJobCancellations subscribes to job cancellation events
	SubscribeToJobCancellations(ctx context.Context, clusterID string, handler func(jobID string) error) error
}

// HPCAuditLogger logs audit events
type HPCAuditLogger interface {
	// LogJobEvent logs a job lifecycle event
	LogJobEvent(event HPCAuditEvent)

	// LogSecurityEvent logs a security event
	LogSecurityEvent(event HPCAuditEvent)

	// LogUsageReport logs a usage report submission
	LogUsageReport(event HPCAuditEvent)
}

// HPCAuditEvent represents an audit event
type HPCAuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	JobID     string                 `json:"job_id,omitempty"`
	ClusterID string                 `json:"cluster_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Success   bool                   `json:"success"`
	ErrorMsg  string                 `json:"error_msg,omitempty"`
}

// HPCJobService manages HPC job lifecycle
type HPCJobService struct {
	config          HPCConfig
	scheduler       HPCScheduler
	reporter        HPCOnChainReporter
	auditor         HPCAuditLogger
	routingEnforcer *RoutingEnforcer

	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
	pendingJobs map[string]*hpctypes.HPCJob // Jobs waiting to be submitted
	activeJobs  map[string]*HPCSchedulerJob // Currently active jobs
	retryQueue  map[string]*retryItem       // Jobs pending retry
}

type retryItem struct {
	job       *hpctypes.HPCJob
	attempts  int
	nextRetry time.Time
	lastError error
}

// NewHPCJobService creates a new HPC job service
func NewHPCJobService(
	config HPCConfig,
	scheduler HPCScheduler,
	reporter HPCOnChainReporter,
	auditor HPCAuditLogger,
) *HPCJobService {
	svc := &HPCJobService{
		config:      config,
		scheduler:   scheduler,
		reporter:    reporter,
		auditor:     auditor,
		stopCh:      make(chan struct{}),
		pendingJobs: make(map[string]*hpctypes.HPCJob),
		activeJobs:  make(map[string]*HPCSchedulerJob),
		retryQueue:  make(map[string]*retryItem),
	}

	// Register lifecycle callback
	scheduler.RegisterLifecycleCallback(svc.handleJobLifecycleEvent)

	return svc
}

// SetRoutingEnforcer sets the routing enforcer for the job service
func (s *HPCJobService) SetRoutingEnforcer(enforcer *RoutingEnforcer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routingEnforcer = enforcer
}

// Start starts the HPC job service
func (s *HPCJobService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	// Start scheduler
	if err := s.scheduler.Start(ctx); err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Recover state if enabled
	if s.config.JobService.EnableStateRecovery {
		if err := s.recoverState(); err != nil {
			// Log but don't fail - state recovery is best effort
			s.logAuditEvent("state_recovery_failed", "", map[string]interface{}{
				"error": err.Error(),
			}, false)
		}
	}

	// Start background workers
	s.wg.Add(3)
	go s.pollJobStatusLoop()
	go s.processRetryQueueLoop()
	go s.reportUsageLoop()

	s.logAuditEvent("service_started", "", map[string]interface{}{
		"scheduler_type": string(s.scheduler.Type()),
		"cluster_id":     s.config.ClusterID,
	}, true)

	return nil
}

// Stop stops the HPC job service
func (s *HPCJobService) Stop() error {
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

	// Persist state
	if s.config.JobService.EnableStateRecovery {
		if err := s.persistState(); err != nil {
			s.logAuditEvent("state_persist_failed", "", map[string]interface{}{
				"error": err.Error(),
			}, false)
		}
	}

	// Stop scheduler
	if err := s.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	s.logAuditEvent("service_stopped", "", nil, true)

	return nil
}

// IsRunning checks if the service is running
func (s *HPCJobService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// SubmitJob submits a new job
func (s *HPCJobService) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	if !s.IsRunning() {
		return nil, fmt.Errorf("job service not running")
	}

	// Validate job
	if err := job.Validate(); err != nil {
		s.logAuditEvent("job_validation_failed", job.JobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return nil, fmt.Errorf("invalid job: %w", err)
	}

	// VE-5B: Enforce routing before submission
	s.mu.RLock()
	enforcer := s.routingEnforcer
	s.mu.RUnlock()

	if enforcer != nil {
		result, err := enforcer.EnforceRouting(ctx, job)
		if err != nil {
			s.logAuditEvent("routing_enforcement_failed", job.JobID, map[string]interface{}{
				"error": err.Error(),
			}, false)
			return nil, fmt.Errorf("routing enforcement failed: %w", err)
		}

		if !result.Allowed {
			s.logAuditEvent("job_routing_rejected", job.JobID, map[string]interface{}{
				"reason":    result.AuditRecord.Reason,
				"violation": result.Violation,
			}, false)
			return nil, fmt.Errorf("job routing rejected: %s", result.AuditRecord.Reason)
		}

		// Update job with enforced routing info
		if result.Decision != nil {
			job.SchedulingDecisionID = result.Decision.DecisionID
			job.ClusterID = result.TargetClusterID
		}

		s.logAuditEvent("routing_enforcement_passed", job.JobID, map[string]interface{}{
			"decision_id":     job.SchedulingDecisionID,
			"cluster_id":      result.TargetClusterID,
			"is_fallback":     result.IsFallback,
			"was_rescheduled": result.WasRescheduled,
		}, true)
	}

	// Check concurrent job limit
	s.mu.RLock()
	activeCount := len(s.activeJobs)
	s.mu.RUnlock()

	if activeCount >= s.config.JobService.MaxConcurrentJobs {
		s.logAuditEvent("job_queue_full", job.JobID, map[string]interface{}{
			"active_jobs": activeCount,
			"max_jobs":    s.config.JobService.MaxConcurrentJobs,
		}, false)
		return nil, NewHPCSchedulerError(HPCErrorCodeQuotaExceeded, "maximum concurrent jobs reached", nil)
	}

	// Submit to scheduler
	schedulerJob, err := s.scheduler.SubmitJob(ctx, job)
	if err != nil {
		// Check if retryable
		if hpcErr, ok := err.(*HPCSchedulerError); ok && hpcErr.Retryable {
			s.enqueueRetry(job, err)
			s.logAuditEvent("job_submission_queued_for_retry", job.JobID, map[string]interface{}{
				"error": err.Error(),
			}, false)
			return nil, err
		}

		s.logAuditEvent("job_submission_failed", job.JobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return nil, err
	}

	// Track active job
	s.mu.Lock()
	s.activeJobs[job.JobID] = schedulerJob
	s.mu.Unlock()

	s.logAuditEvent("job_submitted", job.JobID, map[string]interface{}{
		"scheduler_job_id":       schedulerJob.SchedulerJobID,
		"scheduler_type":         string(schedulerJob.SchedulerType),
		"scheduling_decision_id": job.SchedulingDecisionID,
		"cluster_id":             job.ClusterID,
	}, true)

	// Report status on-chain with decision linkage
	go s.reportJobStatusWithDecision(context.Background(), schedulerJob, job.SchedulingDecisionID)

	return schedulerJob, nil
}

// CancelJob cancels a job
func (s *HPCJobService) CancelJob(ctx context.Context, jobID string) error {
	if !s.IsRunning() {
		return fmt.Errorf("job service not running")
	}

	err := s.scheduler.CancelJob(ctx, jobID)
	if err != nil {
		s.logAuditEvent("job_cancellation_failed", jobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return err
	}

	s.logAuditEvent("job_cancelled", jobID, nil, true)

	return nil
}

// GetJobStatus gets job status
func (s *HPCJobService) GetJobStatus(ctx context.Context, jobID string) (*HPCSchedulerJob, error) {
	return s.scheduler.GetJobStatus(ctx, jobID)
}

// GetJobAccounting gets job accounting metrics
func (s *HPCJobService) GetJobAccounting(ctx context.Context, jobID string) (*HPCSchedulerMetrics, error) {
	return s.scheduler.GetJobAccounting(ctx, jobID)
}

// ListActiveJobs lists all active jobs
func (s *HPCJobService) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	return s.scheduler.ListActiveJobs(ctx)
}

// HandleJobRequest handles a new job request from on-chain events
func (s *HPCJobService) HandleJobRequest(job *hpctypes.HPCJob) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.SubmitJob(ctx, job)
	return err
}

// HandleJobCancellation handles a job cancellation from on-chain events
func (s *HPCJobService) HandleJobCancellation(jobID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.CancelJob(ctx, jobID)
}

// handleJobLifecycleEvent handles job lifecycle events from the scheduler
func (s *HPCJobService) handleJobLifecycleEvent(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	s.logAuditEvent("job_lifecycle_event", job.VirtEngineJobID, map[string]interface{}{
		"event":         string(event),
		"prev_state":    string(prevState),
		"current_state": string(job.State),
	}, true)

	// Report status change on-chain
	go s.reportJobStatus(context.Background(), job)

	// Handle terminal states
	if job.State.IsTerminal() {
		s.mu.Lock()
		delete(s.activeJobs, job.VirtEngineJobID)
		s.mu.Unlock()

		// Report final accounting
		go s.reportJobAccounting(context.Background(), job)
	}
}

// pollJobStatusLoop polls for job status updates
func (s *HPCJobService) pollJobStatusLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.JobService.JobPollInterval)
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

// pollActiveJobs polls status for all active jobs
func (s *HPCJobService) pollActiveJobs() {
	s.mu.RLock()
	jobIDs := make([]string, 0, len(s.activeJobs))
	for id := range s.activeJobs {
		jobIDs = append(jobIDs, id)
	}
	s.mu.RUnlock()

	ctx := context.Background()
	for _, jobID := range jobIDs {
		_, err := s.scheduler.GetJobStatus(ctx, jobID)
		if err != nil {
			// Log but continue polling other jobs
			continue
		}
	}
}

// processRetryQueueLoop processes the retry queue
func (s *HPCJobService) processRetryQueueLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processRetryQueue()
		}
	}
}

// processRetryQueue processes pending retries
func (s *HPCJobService) processRetryQueue() {
	now := time.Now()

	s.mu.Lock()
	var toRetry []*retryItem
	for id, item := range s.retryQueue {
		if now.After(item.nextRetry) {
			toRetry = append(toRetry, item)
			delete(s.retryQueue, id)
		}
	}
	s.mu.Unlock()

	ctx := context.Background()
	for _, item := range toRetry {
		_, err := s.SubmitJob(ctx, item.job)
		if err != nil {
			// Will be re-queued if retryable
			continue
		}
	}
}

// enqueueRetry adds a job to the retry queue
func (s *HPCJobService) enqueueRetry(job *hpctypes.HPCJob, lastError error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.retryQueue[job.JobID]
	attempts := 0
	if exists {
		attempts = existing.attempts
	}

	if attempts >= s.config.Retry.MaxRetries {
		// Max retries exceeded
		s.logAuditEvent("job_max_retries_exceeded", job.JobID, map[string]interface{}{
			"attempts":   attempts,
			"last_error": lastError.Error(),
		}, false)
		return
	}

	// Calculate next retry time with exponential backoff
	backoff := s.config.Retry.InitialBackoff
	for i := 0; i < attempts; i++ {
		backoff = time.Duration(float64(backoff) * s.config.Retry.BackoffMultiplier)
		if backoff > s.config.Retry.MaxBackoff {
			backoff = s.config.Retry.MaxBackoff
			break
		}
	}

	s.retryQueue[job.JobID] = &retryItem{
		job:       job,
		attempts:  attempts + 1,
		nextRetry: time.Now().Add(backoff),
		lastError: lastError,
	}
}

// reportUsageLoop periodically reports usage for active jobs
func (s *HPCJobService) reportUsageLoop() {
	defer s.wg.Done()

	if !s.config.UsageReporting.Enabled {
		return
	}

	ticker := time.NewTicker(s.config.UsageReporting.ReportInterval)
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

// reportActiveJobsUsage reports usage for all active jobs
func (s *HPCJobService) reportActiveJobsUsage() {
	ctx := context.Background()

	activeJobs, err := s.scheduler.ListActiveJobs(ctx)
	if err != nil {
		return
	}

	for _, job := range activeJobs {
		metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
		if err != nil || metrics == nil {
			continue
		}

		if s.reporter != nil {
			err = s.reporter.ReportJobAccounting(ctx, job.VirtEngineJobID, metrics)
			if err != nil {
				s.logAuditEvent("usage_report_failed", job.VirtEngineJobID, map[string]interface{}{
					"error": err.Error(),
				}, false)
			} else {
				s.logAuditEvent("usage_reported", job.VirtEngineJobID, map[string]interface{}{
					"wall_clock_seconds": metrics.WallClockSeconds,
					"cpu_core_seconds":   metrics.CPUCoreSeconds,
				}, true)
			}
		}
	}
}

// reportJobStatus reports job status on-chain
func (s *HPCJobService) reportJobStatus(ctx context.Context, job *HPCSchedulerJob) {
	if s.reporter == nil {
		return
	}

	report, err := s.scheduler.CreateStatusReport(job)
	if err != nil {
		s.logAuditEvent("status_report_creation_failed", job.VirtEngineJobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return
	}

	err = s.reporter.ReportJobStatus(ctx, report)
	if err != nil {
		s.logAuditEvent("status_report_failed", job.VirtEngineJobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return
	}

	s.logAuditEvent("status_reported", job.VirtEngineJobID, map[string]interface{}{
		"state": string(job.State),
	}, true)
}

// reportJobStatusWithDecision reports job status on-chain with scheduling decision linkage
func (s *HPCJobService) reportJobStatusWithDecision(ctx context.Context, job *HPCSchedulerJob, decisionID string) {
	if s.reporter == nil {
		return
	}

	report, err := s.scheduler.CreateStatusReport(job)
	if err != nil {
		s.logAuditEvent("status_report_creation_failed", job.VirtEngineJobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return
	}

	// VE-5B: Add scheduling decision linkage to report
	// The report already has fields, but we add decision context in audit
	err = s.reporter.ReportJobStatus(ctx, report)
	if err != nil {
		s.logAuditEvent("status_report_failed", job.VirtEngineJobID, map[string]interface{}{
			"error":       err.Error(),
			"decision_id": decisionID,
		}, false)
		return
	}

	s.logAuditEvent("status_reported_with_decision", job.VirtEngineJobID, map[string]interface{}{
		"state":                  string(job.State),
		"scheduling_decision_id": decisionID,
	}, true)
}

// reportJobAccounting reports final job accounting on-chain
func (s *HPCJobService) reportJobAccounting(ctx context.Context, job *HPCSchedulerJob) {
	if s.reporter == nil {
		return
	}

	metrics, err := s.scheduler.GetJobAccounting(ctx, job.VirtEngineJobID)
	if err != nil || metrics == nil {
		return
	}

	err = s.reporter.ReportJobAccounting(ctx, job.VirtEngineJobID, metrics)
	if err != nil {
		s.logAuditEvent("accounting_report_failed", job.VirtEngineJobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
		return
	}

	s.logAuditEvent("accounting_reported", job.VirtEngineJobID, map[string]interface{}{
		"wall_clock_seconds": metrics.WallClockSeconds,
		"cpu_core_seconds":   metrics.CPUCoreSeconds,
		"gpu_seconds":        metrics.GPUSeconds,
		"node_hours":         metrics.NodeHours,
	}, true)
}

// logAuditEvent logs an audit event
func (s *HPCJobService) logAuditEvent(eventType, jobID string, details map[string]interface{}, success bool) {
	if s.auditor == nil || !s.config.Audit.Enabled {
		return
	}

	event := HPCAuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		JobID:     jobID,
		ClusterID: s.config.ClusterID,
		Details:   details,
		Success:   success,
	}

	switch {
	case s.config.Audit.LogJobEvents && isJobEvent(eventType):
		s.auditor.LogJobEvent(event)
	case s.config.Audit.LogSecurityEvents && isSecurityEvent(eventType):
		s.auditor.LogSecurityEvent(event)
	case s.config.Audit.LogUsageReports && isUsageEvent(eventType):
		s.auditor.LogUsageReport(event)
	}
}

// State persistence

type persistedState struct {
	ActiveJobs  []*persistedJob `json:"active_jobs"`
	RetryQueue  []*persistedJob `json:"retry_queue"`
	LastUpdated time.Time       `json:"last_updated"`
}

type persistedJob struct {
	JobID          string    `json:"job_id"`
	SchedulerJobID string    `json:"scheduler_job_id"`
	SubmitTime     time.Time `json:"submit_time"`
	RetryAttempts  int       `json:"retry_attempts,omitempty"`
}

func (s *HPCJobService) persistState() error {
	if s.config.JobService.StateStorePath == "" {
		return nil
	}

	s.mu.RLock()
	state := persistedState{
		ActiveJobs:  make([]*persistedJob, 0, len(s.activeJobs)),
		RetryQueue:  make([]*persistedJob, 0, len(s.retryQueue)),
		LastUpdated: time.Now(),
	}

	for id, job := range s.activeJobs {
		state.ActiveJobs = append(state.ActiveJobs, &persistedJob{
			JobID:          id,
			SchedulerJobID: job.SchedulerJobID,
			SubmitTime:     job.SubmitTime,
		})
	}

	for id, item := range s.retryQueue {
		state.RetryQueue = append(state.RetryQueue, &persistedJob{
			JobID:         id,
			SubmitTime:    item.job.CreatedAt,
			RetryAttempts: item.attempts,
		})
	}
	s.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.config.JobService.StateStorePath), 0750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write state file
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.config.JobService.StateStorePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (s *HPCJobService) recoverState() error {
	if s.config.JobService.StateStorePath == "" {
		return nil
	}

	file, err := os.Open(s.config.JobService.StateStorePath)
	if os.IsNotExist(err) {
		return nil // No state to recover
	}
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Recover active jobs by polling scheduler
	ctx := context.Background()
	for _, pj := range state.ActiveJobs {
		job, err := s.scheduler.GetJobStatus(ctx, pj.JobID)
		if err != nil {
			continue // Job may have completed or been removed
		}

		s.mu.Lock()
		s.activeJobs[pj.JobID] = job
		s.mu.Unlock()
	}

	s.logAuditEvent("state_recovered", "", map[string]interface{}{
		"active_jobs_recovered": len(s.activeJobs),
		"state_timestamp":       state.LastUpdated,
	}, true)

	return nil
}

// Helper functions

func isJobEvent(eventType string) bool {
	switch eventType {
	case "job_submitted", "job_cancelled", "job_lifecycle_event",
		"job_validation_failed", "job_submission_failed",
		"job_cancellation_failed", "job_max_retries_exceeded":
		return true
	}
	return false
}

func isSecurityEvent(eventType string) bool {
	switch eventType {
	case "status_report_creation_failed", "authentication_failed",
		"unauthorized_access":
		return true
	}
	return false
}

func isUsageEvent(eventType string) bool {
	switch eventType {
	case "usage_reported", "usage_report_failed",
		"status_reported", "status_report_failed",
		"accounting_reported", "accounting_report_failed":
		return true
	}
	return false
}
