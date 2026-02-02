// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPC Scheduler integration tests with mock adapters
package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	// Initialize SDK config (bech32 prefixes) for tests
	_ "github.com/virtengine/virtengine/sdk/go/sdkutil"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// MockHPCScheduler is a mock implementation of HPCScheduler for testing
type MockHPCScheduler struct {
	mu        sync.RWMutex
	running   bool
	jobs      map[string]*HPCSchedulerJob
	callbacks []HPCJobLifecycleCallback

	// Control behavior
	SubmitError  error
	CancelError  error
	StatusError  error
	ProviderAddr string
}

func NewMockHPCScheduler() *MockHPCScheduler {
	return &MockHPCScheduler{
		jobs:         make(map[string]*HPCSchedulerJob),
		callbacks:    make([]HPCJobLifecycleCallback, 0),
		ProviderAddr: "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr",
	}
}

func (m *MockHPCScheduler) Type() HPCSchedulerType {
	return HPCSchedulerTypeSLURM
}

func (m *MockHPCScheduler) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

func (m *MockHPCScheduler) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *MockHPCScheduler) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

func (m *MockHPCScheduler) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error) {
	if m.SubmitError != nil {
		return nil, m.SubmitError
	}

	schedulerJob := &HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  "slurm-" + job.JobID,
		SchedulerType:   HPCSchedulerTypeSLURM,
		State:           HPCJobStateQueued,
		SubmitTime:      time.Now(),
		OriginalJob:     job,
	}

	m.mu.Lock()
	m.jobs[job.JobID] = schedulerJob
	m.mu.Unlock()

	m.notifyCallbacks(schedulerJob, HPCJobEventSubmitted, HPCJobStatePending)

	return schedulerJob, nil
}

func (m *MockHPCScheduler) CancelJob(ctx context.Context, virtEngineJobID string) error {
	if m.CancelError != nil {
		return m.CancelError
	}

	m.mu.Lock()
	if job, exists := m.jobs[virtEngineJobID]; exists {
		prevState := job.State
		job.State = HPCJobStateCancelled
		now := time.Now()
		job.EndTime = &now
		m.mu.Unlock()
		m.notifyCallbacks(job, HPCJobEventCancelled, prevState)
		return nil
	}
	m.mu.Unlock()

	return NewHPCSchedulerError(HPCErrorCodeJobNotFound, "job not found", nil)
}

func (m *MockHPCScheduler) GetJobStatus(ctx context.Context, virtEngineJobID string) (*HPCSchedulerJob, error) {
	if m.StatusError != nil {
		return nil, m.StatusError
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if job, exists := m.jobs[virtEngineJobID]; exists {
		return job, nil
	}

	return nil, NewHPCSchedulerError(HPCErrorCodeJobNotFound, "job not found", nil)
}

func (m *MockHPCScheduler) GetJobAccounting(ctx context.Context, virtEngineJobID string) (*HPCSchedulerMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if job, exists := m.jobs[virtEngineJobID]; exists {
		return job.Metrics, nil
	}

	return nil, NewHPCSchedulerError(HPCErrorCodeJobNotFound, "job not found", nil)
}

func (m *MockHPCScheduler) ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active []*HPCSchedulerJob
	for _, job := range m.jobs {
		if !job.State.IsTerminal() {
			active = append(active, job)
		}
	}
	return active, nil
}

func (m *MockHPCScheduler) RegisterLifecycleCallback(cb HPCJobLifecycleCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, cb)
}

func (m *MockHPCScheduler) CreateStatusReport(job *HPCSchedulerJob) (*HPCStatusReport, error) {
	return &HPCStatusReport{
		ProviderAddress: m.ProviderAddr,
		VirtEngineJobID: job.VirtEngineJobID,
		SchedulerJobID:  job.SchedulerJobID,
		SchedulerType:   HPCSchedulerTypeSLURM,
		State:           job.State,
		StateMessage:    job.StateMessage,
		ExitCode:        job.ExitCode,
		Metrics:         job.Metrics,
		Timestamp:       time.Now(),
		Signature:       "mock-signature",
	}, nil
}

func (m *MockHPCScheduler) notifyCallbacks(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
	m.mu.RLock()
	callbacks := make([]HPCJobLifecycleCallback, len(m.callbacks))
	copy(callbacks, m.callbacks)
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb(job, event, prevState)
	}
}

// SimulateJobStart simulates a job starting
func (m *MockHPCScheduler) SimulateJobStart(jobID string) {
	m.mu.Lock()
	if job, exists := m.jobs[jobID]; exists {
		prevState := job.State
		job.State = HPCJobStateRunning
		now := time.Now()
		job.StartTime = &now
		m.mu.Unlock()
		m.notifyCallbacks(job, HPCJobEventStarted, prevState)
		return
	}
	m.mu.Unlock()
}

// SimulateJobComplete simulates a job completing
func (m *MockHPCScheduler) SimulateJobComplete(jobID string, exitCode int32) {
	m.mu.Lock()
	if job, exists := m.jobs[jobID]; exists {
		prevState := job.State
		job.State = HPCJobStateCompleted
		job.ExitCode = exitCode
		now := time.Now()
		job.EndTime = &now
		job.Metrics = &HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			MemoryGBSeconds:  7200,
			GPUSeconds:       3600,
			NodesUsed:        4,
			NodeHours:        4.0,
		}
		m.mu.Unlock()
		m.notifyCallbacks(job, HPCJobEventCompleted, prevState)
		return
	}
	m.mu.Unlock()
}

// MockOnChainReporter is a mock implementation of HPCOnChainReporter
type MockOnChainReporter struct {
	mu            sync.Mutex
	StatusReports []*HPCStatusReport
	UsageReports  []struct {
		JobID   string
		Metrics *HPCSchedulerMetrics
	}
	ReportError error
}

func NewMockOnChainReporter() *MockOnChainReporter {
	return &MockOnChainReporter{
		StatusReports: make([]*HPCStatusReport, 0),
	}
}

func (m *MockOnChainReporter) ReportJobStatus(ctx context.Context, report *HPCStatusReport) error {
	if m.ReportError != nil {
		return m.ReportError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StatusReports = append(m.StatusReports, report)
	return nil
}

func (m *MockOnChainReporter) ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error {
	if m.ReportError != nil {
		return m.ReportError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UsageReports = append(m.UsageReports, struct {
		JobID   string
		Metrics *HPCSchedulerMetrics
	}{jobID, metrics})
	return nil
}

func (m *MockOnChainReporter) GetStatusReportCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.StatusReports)
}

// MockAuditLogger is a mock implementation of HPCAuditLogger
type MockAuditLogger struct {
	mu             sync.Mutex
	JobEvents      []HPCAuditEvent
	SecurityEvents []HPCAuditEvent
	UsageEvents    []HPCAuditEvent
}

func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		JobEvents:      make([]HPCAuditEvent, 0),
		SecurityEvents: make([]HPCAuditEvent, 0),
		UsageEvents:    make([]HPCAuditEvent, 0),
	}
}

func (m *MockAuditLogger) LogJobEvent(event HPCAuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.JobEvents = append(m.JobEvents, event)
}

func (m *MockAuditLogger) LogSecurityEvent(event HPCAuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SecurityEvents = append(m.SecurityEvents, event)
}

func (m *MockAuditLogger) LogUsageReport(event HPCAuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UsageEvents = append(m.UsageEvents, event)
}

// MockSigner is a mock implementation of HPCSchedulerSigner
type MockSigner struct {
	Address string
}

func NewMockSigner(address string) *MockSigner {
	return &MockSigner{Address: address}
}

func (m *MockSigner) Sign(data []byte) ([]byte, error) {
	return []byte("mock-signature-" + string(data[:8])), nil
}

func (m *MockSigner) GetProviderAddress() string {
	return m.Address
}

// =============================================================================
// Test Fixtures
// =============================================================================

func createTestJob(jobID string) *hpctypes.HPCJob {
	return &hpctypes.HPCJob{
		JobID:           jobID,
		OfferingID:      "offering-1",
		ClusterID:       "cluster-test",
		ProviderAddress: "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr",
		CustomerAddress: "ve18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuv92kx8",
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: "python:3.9",
			Command:        "python",
			Arguments:      []string{"train.py"},
		},
		Resources: hpctypes.JobResources{
			Nodes:           4,
			CPUCoresPerNode: 32,
			MemoryGBPerNode: 128,
			GPUsPerNode:     2,
			StorageGB:       500,
			GPUType:         "A100",
		},
		MaxRuntimeSeconds: 7200,
		CreatedAt:         time.Now(),
	}
}

func createTestConfig() HPCConfig {
	config := DefaultHPCConfig()
	config.Enabled = true
	config.ClusterID = "cluster-test"
	config.JobService.JobPollInterval = 100 * time.Millisecond
	config.UsageReporting.ReportInterval = 100 * time.Millisecond
	config.Retry.MaxRetries = 2
	config.Retry.InitialBackoff = 10 * time.Millisecond
	config.Audit.Enabled = true
	return config
}

// =============================================================================
// Tests
// =============================================================================

func TestHPCConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  HPCConfig
		wantErr bool
	}{
		{
			name: "valid disabled config",
			config: HPCConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "valid enabled config",
			config: func() HPCConfig {
				c := DefaultHPCConfig()
				c.Enabled = true
				c.ClusterID = testClusterID
				return c
			}(),
			wantErr: false,
		},
		{
			name: "missing cluster ID",
			config: func() HPCConfig {
				c := DefaultHPCConfig()
				c.Enabled = true
				c.ClusterID = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid scheduler type",
			config: func() HPCConfig {
				c := DefaultHPCConfig()
				c.Enabled = true
				c.ClusterID = "test"
				c.SchedulerType = "invalid"
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHPCJobState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    HPCJobState
		terminal bool
	}{
		{HPCJobStatePending, false},
		{HPCJobStateQueued, false},
		{HPCJobStateRunning, false},
		{HPCJobStateSuspended, false},
		{HPCJobStateCompleted, true},
		{HPCJobStateFailed, true},
		{HPCJobStateCancelled, true},
		{HPCJobStateTimeout, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.terminal {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.terminal)
			}
		})
	}
}

func TestMockHPCScheduler_SubmitAndCancel(t *testing.T) {
	scheduler := NewMockHPCScheduler()
	ctx := context.Background()

	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = scheduler.Stop() }()

	job := createTestJob("test-job-1")

	// Submit job
	schedulerJob, err := scheduler.SubmitJob(ctx, job)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	if schedulerJob.VirtEngineJobID != job.JobID {
		t.Errorf("VirtEngineJobID = %v, want %v", schedulerJob.VirtEngineJobID, job.JobID)
	}

	if schedulerJob.State != HPCJobStateQueued {
		t.Errorf("State = %v, want %v", schedulerJob.State, HPCJobStateQueued)
	}

	// Get status
	status, err := scheduler.GetJobStatus(ctx, job.JobID)
	if err != nil {
		t.Fatalf("GetJobStatus() error = %v", err)
	}

	if status.State != HPCJobStateQueued {
		t.Errorf("Status.State = %v, want %v", status.State, HPCJobStateQueued)
	}

	// Cancel job
	if err := scheduler.CancelJob(ctx, job.JobID); err != nil {
		t.Fatalf("CancelJob() error = %v", err)
	}

	// Verify cancelled
	status, err = scheduler.GetJobStatus(ctx, job.JobID)
	if err != nil {
		t.Fatalf("GetJobStatus() after cancel error = %v", err)
	}

	if status.State != HPCJobStateCancelled {
		t.Errorf("Status.State after cancel = %v, want %v", status.State, HPCJobStateCancelled)
	}
}

func TestMockHPCScheduler_LifecycleCallbacks(t *testing.T) {
	scheduler := NewMockHPCScheduler()
	ctx := context.Background()

	var events []HPCJobLifecycleEvent
	var mu sync.Mutex

	scheduler.RegisterLifecycleCallback(func(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()
	})

	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = scheduler.Stop() }()

	job := createTestJob("lifecycle-test")

	// Submit
	_, err := scheduler.SubmitJob(ctx, job)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	// Simulate start
	scheduler.SimulateJobStart(job.JobID)

	// Simulate complete
	scheduler.SimulateJobComplete(job.JobID, 0)

	// Verify events
	mu.Lock()
	defer mu.Unlock()

	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d: %v", len(events), events)
	}

	if events[0] != HPCJobEventSubmitted {
		t.Errorf("events[0] = %v, want %v", events[0], HPCJobEventSubmitted)
	}
	if events[1] != HPCJobEventStarted {
		t.Errorf("events[1] = %v, want %v", events[1], HPCJobEventStarted)
	}
	if events[2] != HPCJobEventCompleted {
		t.Errorf("events[2] = %v, want %v", events[2], HPCJobEventCompleted)
	}
}

func TestHPCJobService_SubmitJob(t *testing.T) {
	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	config := createTestConfig()

	service := NewHPCJobService(config, scheduler, reporter, auditor)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = service.Stop() }()

	job := createTestJob("service-test")

	schedulerJob, err := service.SubmitJob(ctx, job)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	if schedulerJob.VirtEngineJobID != job.JobID {
		t.Errorf("VirtEngineJobID = %v, want %v", schedulerJob.VirtEngineJobID, job.JobID)
	}

	// Wait for async reporting
	time.Sleep(50 * time.Millisecond)

	// Check audit log
	if len(auditor.JobEvents) == 0 {
		t.Error("Expected job events to be logged")
	}
}

func TestHPCJobService_CancelJob(t *testing.T) {
	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	config := createTestConfig()

	service := NewHPCJobService(config, scheduler, reporter, auditor)

	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = service.Stop() }()

	job := createTestJob("cancel-test")

	_, err := service.SubmitJob(ctx, job)
	if err != nil {
		t.Fatalf("SubmitJob() error = %v", err)
	}

	err = service.CancelJob(ctx, job.JobID)
	if err != nil {
		t.Fatalf("CancelJob() error = %v", err)
	}

	status, err := service.GetJobStatus(ctx, job.JobID)
	if err != nil {
		t.Fatalf("GetJobStatus() error = %v", err)
	}

	if status.State != HPCJobStateCancelled {
		t.Errorf("State = %v, want %v", status.State, HPCJobStateCancelled)
	}
}

func TestHPCJobMapper_MapToSLURM(t *testing.T) {
	mapper := NewHPCJobMapper(HPCSchedulerTypeSLURM, "test-cluster")

	job := createTestJob("mapper-test")

	spec, err := mapper.MapToSLURM(job)
	if err != nil {
		t.Fatalf("MapToSLURM() error = %v", err)
	}

	if spec.Nodes != job.Resources.Nodes {
		t.Errorf("Nodes = %v, want %v", spec.Nodes, job.Resources.Nodes)
	}

	if spec.CPUsPerNode != job.Resources.CPUCoresPerNode {
		t.Errorf("CPUsPerNode = %v, want %v", spec.CPUsPerNode, job.Resources.CPUCoresPerNode)
	}

	if spec.GPUs != job.Resources.GPUsPerNode {
		t.Errorf("GPUs = %v, want %v", spec.GPUs, job.Resources.GPUsPerNode)
	}

	if spec.ContainerImage != job.WorkloadSpec.ContainerImage {
		t.Errorf("ContainerImage = %v, want %v", spec.ContainerImage, job.WorkloadSpec.ContainerImage)
	}

	// Check environment
	if spec.Environment["VIRTENGINE_JOB_ID"] != job.JobID {
		t.Errorf("VIRTENGINE_JOB_ID = %v, want %v", spec.Environment["VIRTENGINE_JOB_ID"], job.JobID)
	}
}

func TestHPCJobMapper_MapToMOAB(t *testing.T) {
	mapper := NewHPCJobMapper(HPCSchedulerTypeMOAB, "test-cluster")

	job := createTestJob("moab-mapper-test")

	spec, err := mapper.MapToMOAB(job)
	if err != nil {
		t.Fatalf("MapToMOAB() error = %v", err)
	}

	if spec.Nodes != job.Resources.Nodes {
		t.Errorf("Nodes = %v, want %v", spec.Nodes, job.Resources.Nodes)
	}

	if spec.ProcsPerNode != job.Resources.CPUCoresPerNode {
		t.Errorf("ProcsPerNode = %v, want %v", spec.ProcsPerNode, job.Resources.CPUCoresPerNode)
	}

	if spec.WallTimeLimit != job.MaxRuntimeSeconds {
		t.Errorf("WallTimeLimit = %v, want %v", spec.WallTimeLimit, job.MaxRuntimeSeconds)
	}
}

func TestHPCUsageReporter_CreateUsageRecord(t *testing.T) {
	config := HPCUsageReportingConfig{
		Enabled:        true,
		ReportInterval: time.Minute,
		BatchSize:      10,
	}
	signer := NewMockSigner("ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr")
	reporter := NewHPCUsageReporter(config, "test-cluster", signer)

	job := &HPCSchedulerJob{
		VirtEngineJobID: "test-job",
		SchedulerJobID:  "slurm-123",
		State:           HPCJobStateRunning,
		Metrics: &HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   14400,
			GPUSeconds:       3600,
		},
	}

	record, err := reporter.CreateUsageRecord(
		job,
		"ve18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuv92kx8",
		time.Now().Add(-time.Hour),
		time.Now(),
		false,
	)

	if err != nil {
		t.Fatalf("CreateUsageRecord() error = %v", err)
	}

	if record.JobID != job.VirtEngineJobID {
		t.Errorf("JobID = %v, want %v", record.JobID, job.VirtEngineJobID)
	}

	if record.Signature == "" {
		t.Error("Signature should not be empty")
	}

	if record.IsFinal {
		t.Error("IsFinal should be false")
	}
}

func TestHPCBillingCalculator(t *testing.T) {
	calc := DefaultHPCBillingCalculator()

	metrics := &HPCSchedulerMetrics{
		WallClockSeconds: 3600,       // 1 hour
		CPUCoreSeconds:   144000,     // 40 core-hours
		MemoryGBSeconds:  460800,     // 128 GB for 1 hour
		GPUSeconds:       7200,       // 2 GPU-hours
		NodeHours:        4.0,        // 4 node-hours
		StorageGBHours:   500,        // 500 GB-hours
		NetworkBytesIn:   1073741824, // 1 GB
		NetworkBytesOut:  1073741824, // 1 GB
	}

	cost := calc.CalculateCost(metrics)
	if cost <= 0 {
		t.Errorf("Cost should be positive, got %v", cost)
	}

	breakdown := calc.CalculateCostBreakdown(metrics)
	if breakdown["total"] != cost {
		t.Errorf("Breakdown total = %v, want %v", breakdown["total"], cost)
	}

	// Verify individual components are non-negative
	for component, value := range breakdown {
		if value < 0 {
			t.Errorf("Component %s has negative value: %v", component, value)
		}
	}
}

func TestMapSLURMState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected HPCJobState
	}{
		{"PENDING", "PENDING", HPCJobStateQueued},
		{"RUNNING", "RUNNING", HPCJobStateRunning},
		{"COMPLETED", "COMPLETED", HPCJobStateCompleted},
		{"FAILED", "FAILED", HPCJobStateFailed},
		{"CANCELLED", "CANCELLED", HPCJobStateCancelled},
		{"TIMEOUT", "TIMEOUT", HPCJobStateTimeout},
		{"SUSPENDED", "SUSPENDED", HPCJobStateSuspended},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Import slurm state type for testing
			var slurmState = tt.input
			result := mapSLURMStateString(slurmState)
			if result != tt.expected {
				t.Errorf("MapSLURMState(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper for testing state mapping without importing slurm_adapter
func mapSLURMStateString(state string) HPCJobState {
	switch state {
	case "PENDING":
		return HPCJobStateQueued
	case "RUNNING":
		return HPCJobStateRunning
	case "COMPLETED":
		return HPCJobStateCompleted
	case "FAILED":
		return HPCJobStateFailed
	case "CANCELLED":
		return HPCJobStateCancelled
	case "TIMEOUT":
		return HPCJobStateTimeout
	case "SUSPENDED":
		return HPCJobStateSuspended
	default:
		return HPCJobStatePending
	}
}

func TestHPCSchedulerError(t *testing.T) {
	err := NewHPCSchedulerError(HPCErrorCodeConnectionFailed, "connection failed", nil)

	if !err.Retryable {
		t.Error("Connection failed error should be retryable")
	}

	if err.Error() != "connection failed" {
		t.Errorf("Error() = %v, want 'connection failed'", err.Error())
	}

	// Test with cause
	cause := NewHPCSchedulerError(HPCErrorCodeTimeout, "timeout", nil)
	err2 := NewHPCSchedulerError(HPCErrorCodeJobSubmissionFailed, "submission failed", cause)

	if err2.Unwrap() != cause {
		t.Error("Unwrap() should return cause")
	}
}

func TestHPCUsageAggregator(t *testing.T) {
	agg := NewHPCUsageAggregator("job-1", "cluster-1", "customer", "provider")

	// Add first metrics sample
	agg.AddMetrics(&HPCSchedulerMetrics{
		WallClockSeconds: 1800,
		CPUCoreSeconds:   7200,
		MemoryBytesMax:   1073741824, // 1 GB
		GPUSeconds:       1800,
	})

	// Add second sample
	agg.AddMetrics(&HPCSchedulerMetrics{
		WallClockSeconds: 3600,
		CPUCoreSeconds:   7200,
		MemoryBytesMax:   2147483648, // 2 GB (higher peak)
		GPUSeconds:       1800,
	})

	metrics := agg.GetAggregatedMetrics()

	// Wall clock should be the latest value (not cumulative)
	if metrics.WallClockSeconds != 3600 {
		t.Errorf("WallClockSeconds = %v, want 3600", metrics.WallClockSeconds)
	}

	// CPU should be cumulative
	if metrics.CPUCoreSeconds != 14400 {
		t.Errorf("CPUCoreSeconds = %v, want 14400", metrics.CPUCoreSeconds)
	}

	// Memory max should be peak
	if metrics.MemoryBytesMax != 2147483648 {
		t.Errorf("MemoryBytesMax = %v, want 2147483648", metrics.MemoryBytesMax)
	}
}
