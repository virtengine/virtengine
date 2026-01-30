// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: Scheduler adapter wrappers - implement HPCScheduler interface for each adapter
package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/moab_adapter"
	"github.com/virtengine/virtengine/pkg/ood_adapter"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCSchedulerSigner is the interface for signing status reports
type HPCSchedulerSigner interface {
	Sign(data []byte) ([]byte, error)
	GetProviderAddress() string
}

// =============================================================================
// SLURM Scheduler Wrapper
// =============================================================================

// SLURMSchedulerWrapper wraps the SLURM adapter to implement HPCScheduler
type SLURMSchedulerWrapper struct {
	adapter   *slurm_adapter.SLURMAdapter
	mapper    *HPCJobMapper
	signer    HPCSchedulerSigner
	clusterID string

	mu        sync.RWMutex
	callbacks []HPCJobLifecycleCallback
	jobs      map[string]*HPCSchedulerJob // VirtEngine job ID -> job
}

// NewSLURMSchedulerWrapper creates a new SLURM scheduler wrapper
func NewSLURMSchedulerWrapper(
	adapter *slurm_adapter.SLURMAdapter,
	signer HPCSchedulerSigner,
	clusterID string,
) *SLURMSchedulerWrapper {
	wrapper := &SLURMSchedulerWrapper{
		adapter:   adapter,
		mapper:    NewHPCJobMapper(HPCSchedulerTypeSLURM, clusterID),
		signer:    signer,
		clusterID: clusterID,
		callbacks: make([]HPCJobLifecycleCallback, 0),
		jobs:      make(map[string]*HPCSchedulerJob),
	}
	return wrapper
}

func (w *SLURMSchedulerWrapper) Type() HPCSchedulerType {
	return HPCSchedulerTypeSLURM
}

func (w *SLURMSchedulerWrapper) Start(ctx context.Context) error {
	return w.adapter.Start(ctx)
}

func (w *SLURMSchedulerWrapper) Stop() error {
	return w.adapter.Stop()
}

func (w *SLURMSchedulerWrapper) IsRunning() bool {
	return w.adapter.IsRunning()
}

func (w *SLURMSchedulerWrapper) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	// Map to SLURM spec
	spec, err := w.mapper.MapToSLURM(job)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeInvalidJobSpec, "failed to map job spec", err)
	}

	// Submit to SLURM
	slurmJob, err := w.adapter.SubmitJob(ctx, job.JobID, spec)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobSubmissionFailed, "SLURM job submission failed", err)
	}

	// Create unified job record
	hpcJob := &HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  slurmJob.SLURMJobID,
		SchedulerType:   HPCSchedulerTypeSLURM,
		State:           MapSLURMState(slurmJob.State),
		SubmitTime:      slurmJob.SubmitTime,
		OriginalJob:     job,
	}

	// Store job
	w.mu.Lock()
	w.jobs[job.JobID] = hpcJob
	w.mu.Unlock()

	// Notify callbacks
	w.notifyCallbacks(hpcJob, HPCJobEventSubmitted, HPCJobStatePending)

	return hpcJob, nil
}

func (w *SLURMSchedulerWrapper) CancelJob(ctx context.Context, virtEngineJobID string) error {
	err := w.adapter.CancelJob(ctx, virtEngineJobID)
	if err != nil {
		return NewHPCSchedulerError(HPCErrorCodeJobCancellationFailed, "SLURM job cancellation failed", err)
	}

	// Update local state
	w.mu.Lock()
	if job, exists := w.jobs[virtEngineJobID]; exists {
		prevState := job.State
		job.State = HPCJobStateCancelled
		now := time.Now()
		job.EndTime = &now
		w.mu.Unlock()
		w.notifyCallbacks(job, HPCJobEventCancelled, prevState)
	} else {
		w.mu.Unlock()
	}

	return nil
}

func (w *SLURMSchedulerWrapper) GetJobStatus(ctx context.Context, virtEngineJobID string) (*HPCSchedulerJob, error) {
	slurmJob, err := w.adapter.GetJobStatus(ctx, virtEngineJobID)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobNotFound, "job not found", err)
	}

	w.mu.Lock()
	hpcJob, exists := w.jobs[virtEngineJobID]
	if !exists {
		// Create new tracking record
		hpcJob = &HPCSchedulerJob{
			VirtEngineJobID: virtEngineJobID,
			SchedulerJobID:  slurmJob.SLURMJobID,
			SchedulerType:   HPCSchedulerTypeSLURM,
		}
		w.jobs[virtEngineJobID] = hpcJob
	}

	// Update state
	prevState := hpcJob.State
	hpcJob.State = MapSLURMState(slurmJob.State)
	hpcJob.StateMessage = slurmJob.StatusMessage
	hpcJob.ExitCode = slurmJob.ExitCode
	hpcJob.NodeList = slurmJob.NodeList
	hpcJob.SubmitTime = slurmJob.SubmitTime
	hpcJob.StartTime = slurmJob.StartTime
	hpcJob.EndTime = slurmJob.EndTime

	if slurmJob.UsageMetrics != nil {
		hpcJob.Metrics = MapSLURMMetrics(slurmJob.UsageMetrics, int32(len(slurmJob.NodeList)))
	}

	w.mu.Unlock()

	// Notify on state change
	if prevState != hpcJob.State {
		w.notifyStateChange(hpcJob, prevState)
	}

	return hpcJob, nil
}

func (w *SLURMSchedulerWrapper) GetJobAccounting(ctx context.Context, virtEngineJobID string) (*HPCSchedulerMetrics, error) {
	job, err := w.GetJobStatus(ctx, virtEngineJobID)
	if err != nil {
		return nil, err
	}
	return job.Metrics, nil
}

func (w *SLURMSchedulerWrapper) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var activeJobs []*HPCSchedulerJob
	for _, job := range w.jobs {
		if !job.State.IsTerminal() {
			activeJobs = append(activeJobs, job)
		}
	}
	return activeJobs, nil
}

func (w *SLURMSchedulerWrapper) RegisterLifecycleCallback(cb HPCJobLifecycleCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

func (w *SLURMSchedulerWrapper) CreateStatusReport(job *HPCSchedulerJob) (*HPCStatusReport, error) {
	report := &HPCStatusReport{
		ProviderAddress: w.signer.GetProviderAddress(),
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   HPCSchedulerTypeSLURM,
		State:           job.State,
		StateMessage:    job.StateMessage,
		ExitCode:        job.ExitCode,
		Metrics:         job.Metrics,
		Timestamp:       time.Now(),
	}

	// Sign the report
	hash := hashStatusReport(report)
	sig, err := w.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}
	report.Signature = hex.EncodeToString(sig)

	return report, nil
}

func (w *SLURMSchedulerWrapper) notifyCallbacks(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	w.mu.RLock()
	callbacks := make([]HPCJobLifecycleCallback, len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, cb := range callbacks {
		cb(job, event, prevState)
	}
}

func (w *SLURMSchedulerWrapper) notifyStateChange(job *HPCSchedulerJob, prevState HPCJobState) {
	event := HPCJobEventStateChanged
	switch job.State {
	case HPCJobStateQueued:
		event = HPCJobEventQueued
	case HPCJobStateRunning:
		event = HPCJobEventStarted
	case HPCJobStateCompleted:
		event = HPCJobEventCompleted
	case HPCJobStateFailed:
		event = HPCJobEventFailed
	case HPCJobStateCancelled:
		event = HPCJobEventCancelled
	case HPCJobStateTimeout:
		event = HPCJobEventTimeout
	case HPCJobStateSuspended:
		event = HPCJobEventSuspended
	}
	w.notifyCallbacks(job, event, prevState)
}

// =============================================================================
// MOAB Scheduler Wrapper
// =============================================================================

// MOABSchedulerWrapper wraps the MOAB adapter to implement HPCScheduler
type MOABSchedulerWrapper struct {
	adapter   *moab_adapter.MOABAdapter
	mapper    *HPCJobMapper
	signer    HPCSchedulerSigner
	clusterID string

	mu        sync.RWMutex
	callbacks []HPCJobLifecycleCallback
	jobs      map[string]*HPCSchedulerJob
}

// NewMOABSchedulerWrapper creates a new MOAB scheduler wrapper
func NewMOABSchedulerWrapper(
	adapter *moab_adapter.MOABAdapter,
	signer HPCSchedulerSigner,
	clusterID string,
) *MOABSchedulerWrapper {
	wrapper := &MOABSchedulerWrapper{
		adapter:   adapter,
		mapper:    NewHPCJobMapper(HPCSchedulerTypeMOAB, clusterID),
		signer:    signer,
		clusterID: clusterID,
		callbacks: make([]HPCJobLifecycleCallback, 0),
		jobs:      make(map[string]*HPCSchedulerJob),
	}
	return wrapper
}

func (w *MOABSchedulerWrapper) Type() HPCSchedulerType {
	return HPCSchedulerTypeMOAB
}

func (w *MOABSchedulerWrapper) Start(ctx context.Context) error {
	return w.adapter.Start(ctx)
}

func (w *MOABSchedulerWrapper) Stop() error {
	return w.adapter.Stop()
}

func (w *MOABSchedulerWrapper) IsRunning() bool {
	return w.adapter.IsRunning()
}

func (w *MOABSchedulerWrapper) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	// Map to MOAB spec
	spec, err := w.mapper.MapToMOAB(job)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeInvalidJobSpec, "failed to map job spec", err)
	}

	// Submit to MOAB
	moabJob, err := w.adapter.SubmitJob(ctx, job.JobID, spec)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobSubmissionFailed, "MOAB job submission failed", err)
	}

	// Create unified job record
	hpcJob := &HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  moabJob.MOABJobID,
		SchedulerType:   HPCSchedulerTypeMOAB,
		State:           MapMOABState(moabJob.State),
		SubmitTime:      moabJob.SubmitTime,
		OriginalJob:     job,
	}

	// Store job
	w.mu.Lock()
	w.jobs[job.JobID] = hpcJob
	w.mu.Unlock()

	// Notify callbacks
	w.notifyCallbacks(hpcJob, HPCJobEventSubmitted, HPCJobStatePending)

	return hpcJob, nil
}

func (w *MOABSchedulerWrapper) CancelJob(ctx context.Context, virtEngineJobID string) error {
	err := w.adapter.CancelJob(ctx, virtEngineJobID)
	if err != nil {
		return NewHPCSchedulerError(HPCErrorCodeJobCancellationFailed, "MOAB job cancellation failed", err)
	}

	// Update local state
	w.mu.Lock()
	if job, exists := w.jobs[virtEngineJobID]; exists {
		prevState := job.State
		job.State = HPCJobStateCancelled
		now := time.Now()
		job.EndTime = &now
		w.mu.Unlock()
		w.notifyCallbacks(job, HPCJobEventCancelled, prevState)
	} else {
		w.mu.Unlock()
	}

	return nil
}

func (w *MOABSchedulerWrapper) GetJobStatus(ctx context.Context, virtEngineJobID string) (*HPCSchedulerJob, error) {
	moabJob, err := w.adapter.GetJobStatus(ctx, virtEngineJobID)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobNotFound, "job not found", err)
	}

	w.mu.Lock()
	hpcJob, exists := w.jobs[virtEngineJobID]
	if !exists {
		hpcJob = &HPCSchedulerJob{
			VirtEngineJobID: virtEngineJobID,
			SchedulerJobID:  moabJob.MOABJobID,
			SchedulerType:   HPCSchedulerTypeMOAB,
		}
		w.jobs[virtEngineJobID] = hpcJob
	}

	prevState := hpcJob.State
	hpcJob.State = MapMOABState(moabJob.State)
	hpcJob.StateMessage = moabJob.StatusMessage
	hpcJob.ExitCode = moabJob.ExitCode
	hpcJob.NodeList = moabJob.NodeList
	hpcJob.SubmitTime = moabJob.SubmitTime
	hpcJob.StartTime = moabJob.StartTime
	hpcJob.EndTime = moabJob.EndTime

	if moabJob.UsageMetrics != nil {
		hpcJob.Metrics = MapMOABMetrics(moabJob.UsageMetrics)
	}

	w.mu.Unlock()

	if prevState != hpcJob.State {
		w.notifyStateChange(hpcJob, prevState)
	}

	return hpcJob, nil
}

func (w *MOABSchedulerWrapper) GetJobAccounting(ctx context.Context, virtEngineJobID string) (*HPCSchedulerMetrics, error) {
	job, err := w.GetJobStatus(ctx, virtEngineJobID)
	if err != nil {
		return nil, err
	}
	return job.Metrics, nil
}

func (w *MOABSchedulerWrapper) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var activeJobs []*HPCSchedulerJob
	for _, job := range w.jobs {
		if !job.State.IsTerminal() {
			activeJobs = append(activeJobs, job)
		}
	}
	return activeJobs, nil
}

func (w *MOABSchedulerWrapper) RegisterLifecycleCallback(cb HPCJobLifecycleCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

func (w *MOABSchedulerWrapper) CreateStatusReport(job *HPCSchedulerJob) (*HPCStatusReport, error) {
	report := &HPCStatusReport{
		ProviderAddress: w.signer.GetProviderAddress(),
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   HPCSchedulerTypeMOAB,
		State:           job.State,
		StateMessage:    job.StateMessage,
		ExitCode:        job.ExitCode,
		Metrics:         job.Metrics,
		Timestamp:       time.Now(),
	}

	hash := hashStatusReport(report)
	sig, err := w.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}
	report.Signature = hex.EncodeToString(sig)

	return report, nil
}

func (w *MOABSchedulerWrapper) notifyCallbacks(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	w.mu.RLock()
	callbacks := make([]HPCJobLifecycleCallback, len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, cb := range callbacks {
		cb(job, event, prevState)
	}
}

func (w *MOABSchedulerWrapper) notifyStateChange(job *HPCSchedulerJob, prevState HPCJobState) {
	event := HPCJobEventStateChanged
	switch job.State {
	case HPCJobStateQueued:
		event = HPCJobEventQueued
	case HPCJobStateRunning:
		event = HPCJobEventStarted
	case HPCJobStateCompleted:
		event = HPCJobEventCompleted
	case HPCJobStateFailed:
		event = HPCJobEventFailed
	case HPCJobStateCancelled:
		event = HPCJobEventCancelled
	case HPCJobStateTimeout:
		event = HPCJobEventTimeout
	case HPCJobStateSuspended:
		event = HPCJobEventSuspended
	}
	w.notifyCallbacks(job, event, prevState)
}

// =============================================================================
// OOD Scheduler Wrapper
// =============================================================================

// OODSchedulerWrapper wraps the OOD adapter to implement HPCScheduler
type OODSchedulerWrapper struct {
	adapter   *ood_adapter.OODAdapter
	mapper    *HPCJobMapper
	signer    HPCSchedulerSigner
	clusterID string

	mu        sync.RWMutex
	callbacks []HPCJobLifecycleCallback
	sessions  map[string]*HPCSchedulerJob // VirtEngine session ID -> job
}

// NewOODSchedulerWrapper creates a new OOD scheduler wrapper
func NewOODSchedulerWrapper(
	adapter *ood_adapter.OODAdapter,
	signer HPCSchedulerSigner,
	clusterID string,
) *OODSchedulerWrapper {
	return &OODSchedulerWrapper{
		adapter:   adapter,
		mapper:    NewHPCJobMapper(HPCSchedulerTypeOOD, clusterID),
		signer:    signer,
		clusterID: clusterID,
		callbacks: make([]HPCJobLifecycleCallback, 0),
		sessions:  make(map[string]*HPCSchedulerJob),
	}
}

func (w *OODSchedulerWrapper) Type() HPCSchedulerType {
	return HPCSchedulerTypeOOD
}

func (w *OODSchedulerWrapper) Start(ctx context.Context) error {
	return w.adapter.Start(ctx)
}

func (w *OODSchedulerWrapper) Stop() error {
	return w.adapter.Stop()
}

func (w *OODSchedulerWrapper) IsRunning() bool {
	return w.adapter.IsRunning()
}

func (w *OODSchedulerWrapper) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	// Map to OOD spec
	spec, err := w.mapper.MapToOOD(job)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeInvalidJobSpec, "failed to map job spec", err)
	}

	// For OOD, we need a VEID address - extract from job
	veidAddress := job.CustomerAddress

	// Launch interactive app session
	session, err := w.adapter.LaunchInteractiveApp(ctx, job.JobID, veidAddress, spec)
	if err != nil {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobSubmissionFailed, "OOD session launch failed", err)
	}

	// Create unified job record
	hpcJob := &HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  session.SessionID,
		SchedulerType:   HPCSchedulerTypeOOD,
		State:           MapOODState(session.State),
		SubmitTime:      session.CreatedAt,
		OriginalJob:     job,
	}

	// Store session
	w.mu.Lock()
	w.sessions[job.JobID] = hpcJob
	w.mu.Unlock()

	w.notifyCallbacks(hpcJob, HPCJobEventSubmitted, HPCJobStatePending)

	return hpcJob, nil
}

func (w *OODSchedulerWrapper) CancelJob(ctx context.Context, virtEngineJobID string) error {
	w.mu.RLock()
	job, exists := w.sessions[virtEngineJobID]
	w.mu.RUnlock()

	if !exists {
		return NewHPCSchedulerError(HPCErrorCodeJobNotFound, "session not found", nil)
	}

	err := w.adapter.TerminateSession(ctx, job.SchedulerJobID)
	if err != nil {
		return NewHPCSchedulerError(HPCErrorCodeJobCancellationFailed, "OOD session termination failed", err)
	}

	w.mu.Lock()
	prevState := job.State
	job.State = HPCJobStateCancelled
	now := time.Now()
	job.EndTime = &now
	w.mu.Unlock()

	w.notifyCallbacks(job, HPCJobEventCancelled, prevState)

	return nil
}

func (w *OODSchedulerWrapper) GetJobStatus(ctx context.Context, virtEngineJobID string) (*HPCSchedulerJob, error) {
	w.mu.RLock()
	hpcJob, exists := w.sessions[virtEngineJobID]
	w.mu.RUnlock()

	if !exists {
		return nil, NewHPCSchedulerError(HPCErrorCodeJobNotFound, "session not found", nil)
	}

	session, err := w.adapter.GetSession(ctx, hpcJob.SchedulerJobID)
	if err != nil {
		return hpcJob, nil // Return cached state on error
	}

	w.mu.Lock()
	prevState := hpcJob.State
	hpcJob.State = MapOODState(session.State)
	hpcJob.StateMessage = session.StatusMessage
	if session.StartedAt != nil {
		hpcJob.StartTime = session.StartedAt
	}
	if session.EndedAt != nil {
		hpcJob.EndTime = session.EndedAt
	}
	w.mu.Unlock()

	if prevState != hpcJob.State {
		w.notifyStateChange(hpcJob, prevState)
	}

	return hpcJob, nil
}

func (w *OODSchedulerWrapper) GetJobAccounting(ctx context.Context, virtEngineJobID string) (*HPCSchedulerMetrics, error) {
	job, err := w.GetJobStatus(ctx, virtEngineJobID)
	if err != nil {
		return nil, err
	}

	// Calculate basic metrics from session timing
	if job.StartTime != nil && job.EndTime != nil {
		duration := job.EndTime.Sub(*job.StartTime)
		return &HPCSchedulerMetrics{
			WallClockSeconds: int64(duration.Seconds()),
		}, nil
	}

	return job.Metrics, nil
}

func (w *OODSchedulerWrapper) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var activeJobs []*HPCSchedulerJob
	for _, job := range w.sessions {
		if !job.State.IsTerminal() {
			activeJobs = append(activeJobs, job)
		}
	}
	return activeJobs, nil
}

func (w *OODSchedulerWrapper) RegisterLifecycleCallback(cb HPCJobLifecycleCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

func (w *OODSchedulerWrapper) CreateStatusReport(job *HPCSchedulerJob) (*HPCStatusReport, error) {
	report := &HPCStatusReport{
		ProviderAddress: w.signer.GetProviderAddress(),
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   HPCSchedulerTypeOOD,
		State:           job.State,
		StateMessage:    job.StateMessage,
		ExitCode:        job.ExitCode,
		Metrics:         job.Metrics,
		Timestamp:       time.Now(),
	}

	hash := hashStatusReport(report)
	sig, err := w.signer.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}
	report.Signature = hex.EncodeToString(sig)

	return report, nil
}

func (w *OODSchedulerWrapper) notifyCallbacks(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	w.mu.RLock()
	callbacks := make([]HPCJobLifecycleCallback, len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, cb := range callbacks {
		cb(job, event, prevState)
	}
}

func (w *OODSchedulerWrapper) notifyStateChange(job *HPCSchedulerJob, prevState HPCJobState) {
	event := HPCJobEventStateChanged
	switch job.State {
	case HPCJobStateRunning:
		event = HPCJobEventStarted
	case HPCJobStateCompleted:
		event = HPCJobEventCompleted
	case HPCJobStateFailed:
		event = HPCJobEventFailed
	case HPCJobStateCancelled:
		event = HPCJobEventCancelled
	case HPCJobStateSuspended:
		event = HPCJobEventSuspended
	}
	w.notifyCallbacks(job, event, prevState)
}

// =============================================================================
// Helper Functions
// =============================================================================

// hashStatusReport generates a hash of the status report for signing
func hashStatusReport(report *HPCStatusReport) []byte {
	data := struct {
		ProviderAddress string `json:"provider_address"`
		VirtEngineJobID string `json:"virtengine_job_id"`
		SchedulerJobID  string `json:"scheduler_job_id"`
		SchedulerType   string `json:"scheduler_type"`
		State           string `json:"state"`
		ExitCode        int32  `json:"exit_code"`
		Timestamp       int64  `json:"timestamp"`
	}{
		ProviderAddress: report.ProviderAddress,
		VirtEngineJobID: report.VirtEngineJobID,
		SchedulerJobID:  report.SchedulerJobID,
		SchedulerType:   string(report.SchedulerType),
		State:           string(report.State),
		ExitCode:        report.ExitCode,
		Timestamp:       report.Timestamp.Unix(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(bytes)
	return hash[:]
}
