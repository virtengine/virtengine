//go:build e2e.integration

// Package mocks provides mock implementations for E2E testing.
//
// VE-15D: Mock SLURM integration for HPC E2E flow tests.
package mocks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// SLURMPartition represents a mock SLURM partition.
type SLURMPartition struct {
	Name         string
	Nodes        int32
	MaxRuntime   int64
	MaxNodes     int32
	Features     []string
	Priority     int32
	State        string
	AvailableGPU int32
	AvailableCPU int32
}

// SLURMCluster represents a mock SLURM cluster configuration.
type SLURMCluster struct {
	ClusterID     string
	Name          string
	Region        string
	SLURMVersion  string
	Partitions    []SLURMPartition
	TotalNodes    int32
	TotalCPU      int32
	TotalMemoryGB int64
	TotalGPUs     int32
	Endpoint      string
}

// RoutingDecision records a routing decision for audit purposes.
type RoutingDecision struct {
	JobID             string
	SelectedCluster   string
	CandidateClusters []string
	ScoringFactors    map[string]float64
	Timestamp         time.Time
	Reason            string
	DecisionHash      string
}

// JobExecutionRecord tracks job execution details.
type JobExecutionRecord struct {
	JobID         string
	ClusterID     string
	PartitionName string
	SLURMJobID    string
	NodeList      []string
	SubmitTime    time.Time
	StartTime     *time.Time
	EndTime       *time.Time
	ExitCode      int32
	Metrics       *pd.HPCSchedulerMetrics
}

// MockSLURMIntegration provides a comprehensive mock of SLURM for E2E testing.
type MockSLURMIntegration struct {
	mu               sync.RWMutex
	running          bool
	clusters         map[string]*SLURMCluster
	jobs             map[string]*pd.HPCSchedulerJob
	metrics          map[string]*pd.HPCSchedulerMetrics
	routingDecisions []RoutingDecision
	executionRecords map[string]*JobExecutionRecord
	callbacks        []pd.HPCJobLifecycleCallback

	// Configuration
	simulateLatency bool
	latencyMs       int
	failureRate     float64
	maxCapacity     ResourceCapacity
}

// ResourceCapacity defines resource limits for the mock cluster.
type ResourceCapacity struct {
	MaxCPU      int32
	MaxMemoryMB int64
	MaxGPUs     int32
}

// NewMockSLURMIntegration creates a new mock SLURM integration.
func NewMockSLURMIntegration() *MockSLURMIntegration {
	return &MockSLURMIntegration{
		clusters:         make(map[string]*SLURMCluster),
		jobs:             make(map[string]*pd.HPCSchedulerJob),
		metrics:          make(map[string]*pd.HPCSchedulerMetrics),
		routingDecisions: make([]RoutingDecision, 0),
		executionRecords: make(map[string]*JobExecutionRecord),
		callbacks:        make([]pd.HPCJobLifecycleCallback, 0),
		maxCapacity: ResourceCapacity{
			MaxCPU:      10000,
			MaxMemoryMB: 1024 * 1024, // 1 TB
			MaxGPUs:     1000,
		},
	}
}

// Type returns the scheduler type.
func (m *MockSLURMIntegration) Type() pd.HPCSchedulerType {
	return pd.HPCSchedulerTypeSLURM
}

// Start starts the mock SLURM integration.
func (m *MockSLURMIntegration) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

// Stop stops the mock SLURM integration.
func (m *MockSLURMIntegration) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

// IsRunning checks if the mock is running.
func (m *MockSLURMIntegration) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// RegisterCluster registers a mock SLURM cluster.
func (m *MockSLURMIntegration) RegisterCluster(cluster *SLURMCluster) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clusters[cluster.ClusterID] = cluster
}

// GetCluster returns a cluster by ID.
func (m *MockSLURMIntegration) GetCluster(clusterID string) (*SLURMCluster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cluster, ok := m.clusters[clusterID]
	return cluster, ok
}

// GetClusters returns all registered clusters.
func (m *MockSLURMIntegration) GetClusters() []*SLURMCluster {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clusters := make([]*SLURMCluster, 0, len(m.clusters))
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}
	return clusters
}

// SubmitJob submits a job to the mock SLURM scheduler.
func (m *MockSLURMIntegration) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*pd.HPCSchedulerJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil, fmt.Errorf("scheduler not running")
	}

	// Validate resource requirements
	memoryMB := int64(job.Resources.MemoryGBPerNode) * 1024
	if job.Resources.CPUCoresPerNode > m.maxCapacity.MaxCPU {
		return nil, fmt.Errorf("insufficient CPU: requested %d, available %d", job.Resources.CPUCoresPerNode, m.maxCapacity.MaxCPU)
	}
	if memoryMB > m.maxCapacity.MaxMemoryMB {
		return nil, fmt.Errorf("insufficient memory: requested %d MB, available %d MB", memoryMB, m.maxCapacity.MaxMemoryMB)
	}
	if job.Resources.GPUsPerNode > m.maxCapacity.MaxGPUs {
		return nil, fmt.Errorf("insufficient GPUs: requested %d, available %d", job.Resources.GPUsPerNode, m.maxCapacity.MaxGPUs)
	}

	// Verify cluster exists
	cluster, exists := m.clusters[job.ClusterID]
	if !exists {
		return nil, fmt.Errorf("cluster not found: %s", job.ClusterID)
	}

	// Create routing decision
	decision := m.createRoutingDecision(job, cluster)
	m.routingDecisions = append(m.routingDecisions, decision)

	// Generate SLURM job ID
	slurmJobID := fmt.Sprintf("slurm-%s-%d", job.JobID, time.Now().UnixNano()%10000)

	// Create scheduler job
	schedulerJob := &pd.HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  slurmJobID,
		SchedulerType:   pd.HPCSchedulerTypeSLURM,
		State:           pd.HPCJobStatePending,
		SubmitTime:      time.Now(),
		OriginalJob:     job,
	}

	m.jobs[job.JobID] = schedulerJob
	m.metrics[job.JobID] = &pd.HPCSchedulerMetrics{}

	// Create execution record
	m.executionRecords[job.JobID] = &JobExecutionRecord{
		JobID:         job.JobID,
		ClusterID:     job.ClusterID,
		PartitionName: job.QueueName,
		SLURMJobID:    slurmJobID,
		SubmitTime:    time.Now(),
	}

	// Notify lifecycle callbacks
	m.notifyCallbacks(schedulerJob, pd.HPCJobStatePending)

	return schedulerJob, nil
}

// CancelJob cancels a job.
func (m *MockSLURMIntegration) CancelJob(ctx context.Context, virtEngineJobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[virtEngineJobID]
	if !ok {
		return fmt.Errorf("job not found: %s", virtEngineJobID)
	}

	if job.State.IsTerminal() {
		return fmt.Errorf("cannot cancel terminal job")
	}

	prevState := job.State
	job.State = pd.HPCJobStateCancelled
	now := time.Now()
	job.EndTime = &now

	if record, ok := m.executionRecords[virtEngineJobID]; ok {
		record.EndTime = &now
		record.ExitCode = 130 // SIGINT
	}

	m.notifyCallbacks(job, prevState)

	return nil
}

// GetJobStatus returns the current job status.
func (m *MockSLURMIntegration) GetJobStatus(ctx context.Context, virtEngineJobID string) (*pd.HPCSchedulerJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[virtEngineJobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", virtEngineJobID)
	}

	return job, nil
}

// GetJobAccounting returns job accounting metrics.
func (m *MockSLURMIntegration) GetJobAccounting(ctx context.Context, virtEngineJobID string) (*pd.HPCSchedulerMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, ok := m.metrics[virtEngineJobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", virtEngineJobID)
	}

	return metrics, nil
}

// ListActiveJobs returns all active (non-terminal) jobs.
func (m *MockSLURMIntegration) ListActiveJobs(ctx context.Context) ([]*pd.HPCSchedulerJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*pd.HPCSchedulerJob
	for _, job := range m.jobs {
		if !job.State.IsTerminal() {
			active = append(active, job)
		}
	}
	return active, nil
}

// RegisterLifecycleCallback registers a callback for job lifecycle events.
func (m *MockSLURMIntegration) RegisterLifecycleCallback(cb pd.HPCJobLifecycleCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, cb)
}

// CreateStatusReport creates a signed status report.
func (m *MockSLURMIntegration) CreateStatusReport(job *pd.HPCSchedulerJob) (*pd.HPCStatusReport, error) {
	report := &pd.HPCStatusReport{
		ProviderAddress: job.OriginalJob.ProviderAddress,
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   job.SchedulerType,
		State:           job.State,
		StateMessage:    job.StateMessage,
		ExitCode:        job.ExitCode,
		Timestamp:       time.Now(),
		Signature:       m.generateSignature(job),
	}
	return report, nil
}

// SetJobState sets the job state (for test control).
func (m *MockSLURMIntegration) SetJobState(virtEngineJobID string, state pd.HPCJobState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[virtEngineJobID]
	if !ok {
		return
	}

	prevState := job.State
	job.State = state

	now := time.Now()
	switch state {
	case pd.HPCJobStateRunning:
		if job.StartTime == nil {
			job.StartTime = &now
		}
		if record, ok := m.executionRecords[virtEngineJobID]; ok {
			record.StartTime = &now
			record.NodeList = []string{"node-001", "node-002"}
		}
	case pd.HPCJobStateCompleted, pd.HPCJobStateFailed, pd.HPCJobStateCancelled, pd.HPCJobStateTimeout:
		job.EndTime = &now
		if record, ok := m.executionRecords[virtEngineJobID]; ok {
			record.EndTime = &now
		}
	}

	m.notifyCallbacks(job, prevState)
}

// SetJobMetrics sets job metrics (for test control).
func (m *MockSLURMIntegration) SetJobMetrics(virtEngineJobID string, metrics *pd.HPCSchedulerMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[virtEngineJobID] = metrics
	if record, ok := m.executionRecords[virtEngineJobID]; ok {
		record.Metrics = metrics
	}
}

// SetJobExitCode sets the job exit code (for test control).
func (m *MockSLURMIntegration) SetJobExitCode(virtEngineJobID string, exitCode int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[virtEngineJobID]; ok {
		job.ExitCode = exitCode
	}
	if record, ok := m.executionRecords[virtEngineJobID]; ok {
		record.ExitCode = exitCode
	}
}

// SetMaxCapacity sets resource capacity limits (for test control).
func (m *MockSLURMIntegration) SetMaxCapacity(cpu int32, memoryMB int64, gpus int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxCapacity = ResourceCapacity{
		MaxCPU:      cpu,
		MaxMemoryMB: memoryMB,
		MaxGPUs:     gpus,
	}
}

// GetRoutingDecisions returns all routing decisions for audit.
func (m *MockSLURMIntegration) GetRoutingDecisions() []RoutingDecision {
	m.mu.RLock()
	defer m.mu.RUnlock()
	decisions := make([]RoutingDecision, len(m.routingDecisions))
	copy(decisions, m.routingDecisions)
	return decisions
}

// GetRoutingDecisionForJob returns the routing decision for a specific job.
func (m *MockSLURMIntegration) GetRoutingDecisionForJob(jobID string) (*RoutingDecision, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := len(m.routingDecisions) - 1; i >= 0; i-- {
		if m.routingDecisions[i].JobID == jobID {
			return &m.routingDecisions[i], true
		}
	}
	return nil, false
}

// GetExecutionRecord returns the execution record for a job.
func (m *MockSLURMIntegration) GetExecutionRecord(jobID string) (*JobExecutionRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, ok := m.executionRecords[jobID]
	return record, ok
}

// GetExecutionRecords returns all execution records.
func (m *MockSLURMIntegration) GetExecutionRecords() []*JobExecutionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	records := make([]*JobExecutionRecord, 0, len(m.executionRecords))
	for _, r := range m.executionRecords {
		records = append(records, r)
	}
	return records
}

// IsValidTransition checks if a state transition is valid.
func (m *MockSLURMIntegration) IsValidTransition(from, to pd.HPCJobState) bool {
	validTransitions := map[pd.HPCJobState][]pd.HPCJobState{
		pd.HPCJobStatePending:   {pd.HPCJobStateQueued, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateQueued:    {pd.HPCJobStateStarting, pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateStarting:  {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
		pd.HPCJobStateRunning:   {pd.HPCJobStateCompleted, pd.HPCJobStateFailed, pd.HPCJobStateCancelled, pd.HPCJobStateSuspended, pd.HPCJobStateTimeout},
		pd.HPCJobStateSuspended: {pd.HPCJobStateRunning, pd.HPCJobStateFailed, pd.HPCJobStateCancelled},
	}

	if allowed, ok := validTransitions[from]; ok {
		for _, v := range allowed {
			if v == to {
				return true
			}
		}
	}
	return false
}

// SimulateJobExecution simulates a complete job execution lifecycle.
func (m *MockSLURMIntegration) SimulateJobExecution(ctx context.Context, jobID string, durationMs int, success bool, metrics *pd.HPCSchedulerMetrics) error {
	// Progress through states
	m.SetJobState(jobID, pd.HPCJobStateQueued)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Millisecond * time.Duration(durationMs/4)):
	}

	m.SetJobState(jobID, pd.HPCJobStateRunning)

	select {
	case <-ctx.Done():
		m.SetJobState(jobID, pd.HPCJobStateCancelled)
		return ctx.Err()
	case <-time.After(time.Millisecond * time.Duration(durationMs/2)):
	}

	if metrics != nil {
		m.SetJobMetrics(jobID, metrics)
	}

	if success {
		m.SetJobExitCode(jobID, 0)
		m.SetJobState(jobID, pd.HPCJobStateCompleted)
	} else {
		m.SetJobExitCode(jobID, 1)
		m.SetJobState(jobID, pd.HPCJobStateFailed)
	}

	return nil
}

// createRoutingDecision creates a routing decision record.
func (m *MockSLURMIntegration) createRoutingDecision(job *hpctypes.HPCJob, cluster *SLURMCluster) RoutingDecision {
	candidates := make([]string, 0, len(m.clusters))
	for id := range m.clusters {
		candidates = append(candidates, id)
	}

	scoringFactors := map[string]float64{
		"resource_availability": 0.95,
		"queue_depth":           0.80,
		"geographic_proximity":  1.0,
		"price_competitiveness": 0.90,
	}

	decision := RoutingDecision{
		JobID:             job.JobID,
		SelectedCluster:   cluster.ClusterID,
		CandidateClusters: candidates,
		ScoringFactors:    scoringFactors,
		Timestamp:         time.Now(),
		Reason:            fmt.Sprintf("Cluster %s selected based on resource availability and pricing", cluster.ClusterID),
	}

	// Generate decision hash for auditability
	hashInput := fmt.Sprintf("%s:%s:%d", decision.JobID, decision.SelectedCluster, decision.Timestamp.UnixNano())
	hash := sha256.Sum256([]byte(hashInput))
	decision.DecisionHash = hex.EncodeToString(hash[:])

	return decision
}

// notifyCallbacks notifies all registered callbacks of state changes.
func (m *MockSLURMIntegration) notifyCallbacks(job *pd.HPCSchedulerJob, prevState pd.HPCJobState) {
	event := m.stateToEvent(job.State)
	for _, cb := range m.callbacks {
		cb(job, event, prevState)
	}
}

// stateToEvent converts a job state to a lifecycle event.
func (m *MockSLURMIntegration) stateToEvent(state pd.HPCJobState) pd.HPCJobLifecycleEvent {
	switch state {
	case pd.HPCJobStatePending:
		return pd.HPCJobEventSubmitted
	case pd.HPCJobStateQueued:
		return pd.HPCJobEventQueued
	case pd.HPCJobStateRunning:
		return pd.HPCJobEventStarted
	case pd.HPCJobStateCompleted:
		return pd.HPCJobEventCompleted
	case pd.HPCJobStateFailed:
		return pd.HPCJobEventFailed
	case pd.HPCJobStateCancelled:
		return pd.HPCJobEventCancelled
	case pd.HPCJobStateTimeout:
		return pd.HPCJobEventTimeout
	default:
		return pd.HPCJobEventSubmitted
	}
}

// generateSignature generates a mock signature for status reports.
func (m *MockSLURMIntegration) generateSignature(job *pd.HPCSchedulerJob) string {
	data := fmt.Sprintf("%s:%s:%s:%d", job.VirtEngineJobID, job.SchedulerJobID, job.State, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]) // Truncated for mock
}

// DefaultTestCluster returns a default cluster for testing.
func DefaultTestCluster() *SLURMCluster {
	return &SLURMCluster{
		ClusterID:     "e2e-slurm-cluster",
		Name:          "E2E Test SLURM Cluster",
		Region:        "us-east",
		SLURMVersion:  "23.02.4",
		TotalNodes:    100,
		TotalCPU:      6400,
		TotalMemoryGB: 25600,
		TotalGPUs:     400,
		Endpoint:      "slurm://e2e-cluster.example.com:6817",
		Partitions: []SLURMPartition{
			{
				Name:         "default",
				Nodes:        50,
				MaxRuntime:   86400,
				MaxNodes:     10,
				Features:     []string{"cpu"},
				Priority:     50,
				State:        "up",
				AvailableCPU: 3200,
			},
			{
				Name:         "gpu",
				Nodes:        30,
				MaxRuntime:   172800,
				MaxNodes:     5,
				Features:     []string{"gpu", "a100"},
				Priority:     100,
				State:        "up",
				AvailableCPU: 1920,
				AvailableGPU: 240,
			},
			{
				Name:         "highmem",
				Nodes:        20,
				MaxRuntime:   86400,
				MaxNodes:     4,
				Features:     []string{"highmem", "1tb"},
				Priority:     75,
				State:        "up",
				AvailableCPU: 1280,
			},
		},
	}
}
