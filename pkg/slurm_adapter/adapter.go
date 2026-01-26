// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-501: SLURM orchestration adapter in Provider Daemon (v1)
package slurm_adapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// SLURMAdapter implements the SLURM orchestration adapter
type SLURMAdapter struct {
	config     SLURMConfig
	client     SLURMClient
	signer     JobSigner
	mu         sync.RWMutex
	jobs       map[string]*SLURMJob // SLURM job ID -> job
	jobMapping map[string]string    // VirtEngine job ID -> SLURM job ID
	running    bool
	stopCh     chan struct{}
}

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

// JobStatusReport contains job status for on-chain reporting
type JobStatusReport struct {
	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// SLURMJobID is the SLURM job ID
	SLURMJobID string `json:"slurm_job_id"`

	// State is the job state
	State SLURMJobState `json:"state"`

	// StatusMessage is the status message
	StatusMessage string `json:"status_message,omitempty"`

	// ExitCode is the exit code
	ExitCode int32 `json:"exit_code,omitempty"`

	// UsageMetrics are the usage metrics
	UsageMetrics *SLURMUsageMetrics `json:"usage_metrics,omitempty"`

	// Timestamp is when the report was created
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature
	Signature string `json:"signature"`
}

// Hash generates a hash for signing
func (r *JobStatusReport) Hash() []byte {
	data := struct {
		ProviderAddress string `json:"provider_address"`
		VirtEngineJobID string `json:"virtengine_job_id"`
		SLURMJobID      string `json:"slurm_job_id"`
		State           string `json:"state"`
		ExitCode        int32  `json:"exit_code"`
		Timestamp       int64  `json:"timestamp"`
	}{
		ProviderAddress: r.ProviderAddress,
		VirtEngineJobID: r.VirtEngineJobID,
		SLURMJobID:      r.SLURMJobID,
		State:           string(r.State),
		ExitCode:        r.ExitCode,
		Timestamp:       r.Timestamp.Unix(),
	}
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// NewSLURMAdapter creates a new SLURM adapter
func NewSLURMAdapter(config SLURMConfig, client SLURMClient, signer JobSigner) *SLURMAdapter {
	return &SLURMAdapter{
		config:     config,
		client:     client,
		signer:     signer,
		jobs:       make(map[string]*SLURMJob),
		jobMapping: make(map[string]string),
		stopCh:     make(chan struct{}),
	}
}

// Start starts the SLURM adapter
func (a *SLURMAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.running = true
	a.mu.Unlock()

	// Connect to SLURM
	if err := a.client.Connect(ctx); err != nil {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
		return fmt.Errorf("failed to connect to SLURM: %w", err)
	}

	// Start job polling
	go a.pollJobs()

	return nil
}

// Stop stops the SLURM adapter
func (a *SLURMAdapter) Stop() error {
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
func (a *SLURMAdapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// SubmitJob submits a job to SLURM
func (a *SLURMAdapter) SubmitJob(ctx context.Context, virtEngineJobID string, spec *SLURMJobSpec) (*SLURMJob, error) {
	if !a.IsRunning() {
		return nil, ErrSLURMNotConnected
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	// Submit to SLURM
	slurmJobID, err := a.client.SubmitJob(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJobSubmissionFailed, err)
	}

	// Create job record
	job := &SLURMJob{
		SLURMJobID:      slurmJobID,
		VirtEngineJobID: virtEngineJobID,
		Spec:            spec,
		State:           SLURMJobStatePending,
		SubmitTime:      time.Now(),
	}

	// Store job
	a.mu.Lock()
	a.jobs[slurmJobID] = job
	a.jobMapping[virtEngineJobID] = slurmJobID
	a.mu.Unlock()

	return job, nil
}

// CancelJob cancels a job
func (a *SLURMAdapter) CancelJob(ctx context.Context, virtEngineJobID string) error {
	a.mu.RLock()
	slurmJobID, exists := a.jobMapping[virtEngineJobID]
	a.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	if err := a.client.CancelJob(ctx, slurmJobID); err != nil {
		return fmt.Errorf("%w: %v", ErrJobCancellationFailed, err)
	}

	// Update local state
	a.mu.Lock()
	if job, exists := a.jobs[slurmJobID]; exists {
		job.State = SLURMJobStateCancelled
		job.StatusMessage = "Cancelled by user"
		now := time.Now()
		job.EndTime = &now
	}
	a.mu.Unlock()

	return nil
}

// GetJobStatus gets job status
func (a *SLURMAdapter) GetJobStatus(ctx context.Context, virtEngineJobID string) (*SLURMJob, error) {
	a.mu.RLock()
	slurmJobID, exists := a.jobMapping[virtEngineJobID]
	if !exists {
		a.mu.RUnlock()
		return nil, ErrJobNotFound
	}
	job, exists := a.jobs[slurmJobID]
	a.mu.RUnlock()

	if !exists {
		return nil, ErrJobNotFound
	}

	// Get latest status from SLURM
	updatedJob, err := a.client.GetJobStatus(ctx, slurmJobID)
	if err != nil {
		// Return cached status if SLURM query fails
		return job, nil
	}

	// Update cached job
	a.mu.Lock()
	a.jobs[slurmJobID] = updatedJob
	a.mu.Unlock()

	return updatedJob, nil
}

// GetJobsByCluster gets all jobs for the cluster
func (a *SLURMAdapter) GetJobsByCluster() []*SLURMJob {
	a.mu.RLock()
	defer a.mu.RUnlock()

	jobs := make([]*SLURMJob, 0, len(a.jobs))
	for _, job := range a.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// CreateStatusReport creates a signed status report for on-chain submission
func (a *SLURMAdapter) CreateStatusReport(job *SLURMJob) (*JobStatusReport, error) {
	report := &JobStatusReport{
		ProviderAddress: a.signer.GetProviderAddress(),
		VirtEngineJobID: job.VirtEngineJobID,
		SLURMJobID:      job.SLURMJobID,
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

// pollJobs polls for job status updates
func (a *SLURMAdapter) pollJobs() {
	ticker := time.NewTicker(a.config.JobPollInterval)
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
func (a *SLURMAdapter) updateJobStatuses() {
	a.mu.RLock()
	slurmJobIDs := make([]string, 0, len(a.jobs))
	for id, job := range a.jobs {
		// Only poll non-terminal jobs
		if !isTerminalState(job.State) {
			slurmJobIDs = append(slurmJobIDs, id)
		}
	}
	a.mu.RUnlock()

	ctx := context.Background()
	for _, slurmJobID := range slurmJobIDs {
		updatedJob, err := a.client.GetJobStatus(ctx, slurmJobID)
		if err != nil {
			continue
		}

		// Get accounting data for completed jobs
		if isTerminalState(updatedJob.State) {
			metrics, err := a.client.GetJobAccounting(ctx, slurmJobID)
			if err == nil {
				updatedJob.UsageMetrics = metrics
			}
		}

		a.mu.Lock()
		a.jobs[slurmJobID] = updatedJob
		a.mu.Unlock()
	}
}

// isTerminalState checks if a state is terminal
func isTerminalState(state SLURMJobState) bool {
	switch state {
	case SLURMJobStateCompleted, SLURMJobStateFailed, SLURMJobStateCancelled, SLURMJobStateTimeout:
		return true
	default:
		return false
	}
}

// MapToVirtEngineState maps SLURM state to VirtEngine job state
func MapToVirtEngineState(state SLURMJobState) string {
	switch state {
	case SLURMJobStatePending:
		return "queued"
	case SLURMJobStateRunning:
		return "running"
	case SLURMJobStateCompleted:
		return "completed"
	case SLURMJobStateFailed:
		return "failed"
	case SLURMJobStateCancelled:
		return "cancelled"
	case SLURMJobStateTimeout:
		return "timeout"
	case SLURMJobStateSuspended:
		return "paused"
	default:
		return "pending"
	}
}
