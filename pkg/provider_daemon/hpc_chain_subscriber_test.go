// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-21C: HPC Chain Subscriber tests
package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// =============================================================================
// Mock Implementations for Chain Subscriber Tests
// =============================================================================

const (
	chainSubscriberTestClusterID      = "test-cluster"
	chainSubscriberTestProviderAddr   = "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr"
	chainSubscriberTestCustomerAddr   = "ve18qa2a2ltfyvkyj0ggj3hkvuj6twzyumuv92kx8"
	chainSubscriberTestClusterEventID = "test-cluster"
)

// MockHPCChainClient implements HPCChainClient for testing
type MockHPCChainClient struct {
	mu               sync.Mutex
	jobHandler       func(*hpctypes.HPCJob) error
	cancelHandler    func(string) error
	subscribeError   error
	blockOnSubscribe bool
	blockCh          chan struct{}

	// Track calls
	statusReports []*HPCStatusReport
	usageReports  []struct {
		JobID   string
		Metrics *HPCSchedulerMetrics
	}
}

func NewMockHPCChainClient() *MockHPCChainClient {
	return &MockHPCChainClient{
		blockCh: make(chan struct{}),
	}
}

func (m *MockHPCChainClient) SubscribeToJobRequests(ctx context.Context, clusterID string, handler func(*hpctypes.HPCJob) error) error {
	m.mu.Lock()
	m.jobHandler = handler
	subscribeError := m.subscribeError
	blockOnSubscribe := m.blockOnSubscribe
	blockCh := m.blockCh
	m.mu.Unlock()

	if subscribeError != nil {
		return subscribeError
	}

	if blockOnSubscribe {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-blockCh:
			return nil
		}
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m *MockHPCChainClient) SubscribeToJobCancellations(ctx context.Context, clusterID string, handler func(string) error) error {
	m.mu.Lock()
	m.cancelHandler = handler
	subscribeError := m.subscribeError
	blockOnSubscribe := m.blockOnSubscribe
	blockCh := m.blockCh
	m.mu.Unlock()

	if subscribeError != nil {
		return subscribeError
	}

	if blockOnSubscribe {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-blockCh:
			return nil
		}
	}

	<-ctx.Done()
	return ctx.Err()
}

func (m *MockHPCChainClient) ReportJobStatus(ctx context.Context, report *HPCStatusReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusReports = append(m.statusReports, report)
	return nil
}

func (m *MockHPCChainClient) ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usageReports = append(m.usageReports, struct {
		JobID   string
		Metrics *HPCSchedulerMetrics
	}{jobID, metrics})
	return nil
}

func (m *MockHPCChainClient) SubmitAccountingRecord(ctx context.Context, record *hpctypes.HPCAccountingRecord) error {
	return nil
}

func (m *MockHPCChainClient) SubmitUsageSnapshot(ctx context.Context, snapshot *hpctypes.HPCUsageSnapshot) error {
	return nil
}

func (m *MockHPCChainClient) GetBillingRules(ctx context.Context, providerAddr string) (*hpctypes.HPCBillingRules, error) {
	return &hpctypes.HPCBillingRules{
		FormulaVersion: "v1",
	}, nil
}

func (m *MockHPCChainClient) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	return 100, nil
}

func (m *MockHPCChainClient) SimulateJobEvent(job *hpctypes.HPCJob) error {
	m.mu.Lock()
	handler := m.jobHandler
	m.mu.Unlock()

	if handler != nil {
		return handler(job)
	}
	return nil
}

func (m *MockHPCChainClient) SimulateCancelEvent(jobID string) error {
	m.mu.Lock()
	handler := m.cancelHandler
	m.mu.Unlock()

	if handler != nil {
		return handler(jobID)
	}
	return nil
}

func (m *MockHPCChainClient) SetSubscribeError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribeError = err
}

func (m *MockHPCChainClient) Unblock() {
	close(m.blockCh)
}

// =============================================================================
// Tests
// =============================================================================

func TestNewHPCChainSubscriberWithStats_Validation(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	jobService := NewHPCJobService(createTestConfig(), scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	tests := []struct {
		name      string
		clusterID string
		service   *HPCJobService
		wantError bool
	}{
		{
			name:      "valid config",
			clusterID: chainSubscriberTestClusterID,
			service:   jobService,
			wantError: false,
		},
		{
			name:      "nil job service",
			clusterID: chainSubscriberTestClusterID,
			service:   nil,
			wantError: true,
		},
		{
			name:      "missing cluster ID",
			clusterID: "",
			service:   jobService,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHPCChainSubscriberWithStats(config, tt.clusterID, "", chainClient, tt.service)
			if (err != nil) != tt.wantError {
				t.Errorf("NewHPCChainSubscriberWithStats() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHPCChainSubscriberWithStats_StartStop(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Should not be running initially
	if sub.IsRunning() {
		t.Error("Expected subscriber to not be running initially")
	}

	// Start
	ctx := context.Background()
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Should be running
	if !sub.IsRunning() {
		t.Error("Expected subscriber to be running after Start()")
	}

	// Double start should be no-op
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Double Start() error = %v", err)
	}

	// Stop
	if err := sub.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// Should not be running
	if sub.IsRunning() {
		t.Error("Expected subscriber to not be running after Stop()")
	}

	// Double stop should be no-op
	if err := sub.Stop(); err != nil {
		t.Fatalf("Double Stop() error = %v", err)
	}
}

func TestHPCChainSubscriberWithStats_GetStats(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Get stats before start
	stats := sub.GetStats()
	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	if stats.JobsReceived != 0 {
		t.Errorf("Expected JobsReceived to be 0, got %d", stats.JobsReceived)
	}

	if stats.JobsProcessed != 0 {
		t.Errorf("Expected JobsProcessed to be 0, got %d", stats.JobsProcessed)
	}

	// Start and get stats
	ctx := context.Background()
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	stats = sub.GetStats()
	if stats.StartTime.IsZero() {
		t.Error("Expected StartTime to be set after Start()")
	}

	if stats.Uptime == "" {
		t.Error("Expected Uptime to be set after Start()")
	}
}

func TestHPCChainSubscriberWithStats_InjectJobEvent(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	// Start job service
	ctx := context.Background()
	if err := jobService.Start(ctx); err != nil {
		t.Fatalf("JobService.Start() error = %v", err)
	}
	defer func() { _ = jobService.Stop() }()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Inject before start should fail
	job := createTestJob("inject-test-1")
	if err := sub.InjectJobEvent(job); err == nil {
		t.Error("Expected error when injecting before start")
	}

	// Start subscriber
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// Inject job event
	if err := sub.InjectJobEvent(job); err != nil {
		t.Fatalf("InjectJobEvent() error = %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := sub.GetStats()
	if stats.JobsReceived != 1 {
		t.Errorf("Expected JobsReceived to be 1, got %d", stats.JobsReceived)
	}

	if stats.JobsProcessed != 1 {
		t.Errorf("Expected JobsProcessed to be 1, got %d", stats.JobsProcessed)
	}
}

func TestHPCChainSubscriberWithStats_InjectCancelEvent(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	// Start job service
	ctx := context.Background()
	if err := jobService.Start(ctx); err != nil {
		t.Fatalf("JobService.Start() error = %v", err)
	}
	defer func() { _ = jobService.Stop() }()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Inject before start should fail
	if err := sub.InjectCancelEvent("job-1"); err == nil {
		t.Error("Expected error when injecting before start")
	}

	// Start subscriber
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// First submit a job
	job := createTestJob("cancel-test-1")
	if err := sub.InjectJobEvent(job); err != nil {
		t.Fatalf("InjectJobEvent() error = %v", err)
	}

	// Wait for job to be processed
	time.Sleep(100 * time.Millisecond)

	// Now inject cancel event
	if err := sub.InjectCancelEvent(job.JobID); err != nil {
		t.Fatalf("InjectCancelEvent() error = %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := sub.GetStats()
	if stats.CancelsReceived != 1 {
		t.Errorf("Expected CancelsReceived to be 1, got %d", stats.CancelsReceived)
	}

	if stats.CancelsProcessed != 1 {
		t.Errorf("Expected CancelsProcessed to be 1, got %d", stats.CancelsProcessed)
	}
}

func TestHPCChainSubscriberWithStats_ProviderAddressFilter(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	// Start job service
	ctx := context.Background()
	if err := jobService.Start(ctx); err != nil {
		t.Fatalf("JobService.Start() error = %v", err)
	}
	defer func() { _ = jobService.Stop() }()

	// Create subscriber with provider address filter
	sub, err := NewHPCChainSubscriberWithStats(
		config,
		chainSubscriberTestClusterID,
		chainSubscriberTestProviderAddr, // Filter for this address only
		chainClient,
		jobService,
	)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// Inject job with matching provider address
	job1 := createTestJob("filter-test-1")
	job1.ProviderAddress = chainSubscriberTestProviderAddr
	if err := sub.InjectJobEvent(job1); err != nil {
		t.Fatalf("InjectJobEvent() error = %v", err)
	}

	// Inject job with non-matching provider address
	job2 := createTestJob("filter-test-2")
	job2.ProviderAddress = chainSubscriberTestCustomerAddr
	if err := sub.InjectJobEvent(job2); err != nil {
		t.Fatalf("InjectJobEvent() error = %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check stats - only matching job should be processed
	stats := sub.GetStats()
	if stats.JobsReceived != 2 {
		t.Errorf("Expected JobsReceived to be 2, got %d", stats.JobsReceived)
	}

	if stats.JobsProcessed != 1 {
		t.Errorf("Expected JobsProcessed to be 1 (filtered), got %d", stats.JobsProcessed)
	}
}

func TestHPCChainSubscriberWithStats_BufferFull(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 1 // Very small buffer

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	ctx := context.Background()
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// Fill the buffer
	for i := 0; i < 5; i++ {
		job := createTestJob("buffer-test-" + string(rune('0'+i)))
		_ = sub.InjectJobEvent(job)
	}

	// Give time for events to be processed
	time.Sleep(50 * time.Millisecond)

	stats := sub.GetStats()
	// Some events should have been received, but not all processed due to buffer
	if stats.JobsReceived == 0 {
		t.Error("Expected some jobs to be received")
	}
}

func TestHPCChainSubscriberWithStats_IsConnected(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Should not be connected initially
	if sub.IsConnected() {
		t.Error("Expected subscriber to not be connected initially")
	}

	// Mock client blocks to keep subscription active
	chainClient.blockOnSubscribe = true

	ctx := context.Background()
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Wait a bit for connection to establish
	time.Sleep(100 * time.Millisecond)

	// Should be connected
	if !sub.IsConnected() {
		t.Error("Expected subscriber to be connected after Start()")
	}

	// Stop
	if err := sub.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	// Should not be connected after stop
	if sub.IsConnected() {
		t.Error("Expected subscriber to not be connected after Stop()")
	}
}

func TestHPCChainSubscriberWithStats_GetHealth(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)
	chainClient := NewMockHPCChainClient()

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	// Health before start
	health := sub.GetHealth()
	if health.Healthy {
		t.Error("Expected health to be false before start")
	}

	if health.Name != "chain_subscriber_stats" {
		t.Errorf("Expected health.Name to be 'chain_subscriber_stats', got %s", health.Name)
	}

	ctx := context.Background()
	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// Health after start
	health = sub.GetHealth()
	if !health.Healthy {
		t.Error("Expected health to be true after start")
	}

	if health.Details == nil {
		t.Error("Expected health.Details to be set")
	}
}

func TestHPCChainSubscriberWithStats_WithChainEvents(t *testing.T) {
	config := DefaultHPCChainSubscriberConfig()
	config.SubscriptionBufferSize = 10

	scheduler := NewMockHPCScheduler()
	reporter := NewMockOnChainReporter()
	auditor := NewMockAuditLogger()
	hpcConfig := createTestConfig()
	jobService := NewHPCJobService(hpcConfig, scheduler, reporter, auditor)

	// Start job service
	ctx := context.Background()
	if err := jobService.Start(ctx); err != nil {
		t.Fatalf("JobService.Start() error = %v", err)
	}
	defer func() { _ = jobService.Stop() }()

	chainClient := NewMockHPCChainClient()
	chainClient.blockOnSubscribe = true

	sub, err := NewHPCChainSubscriberWithStats(config, chainSubscriberTestClusterID, "", chainClient, jobService)
	if err != nil {
		t.Fatalf("NewHPCChainSubscriberWithStats() error = %v", err)
	}

	if err := sub.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = sub.Stop() }()

	// Wait for subscription to be established
	time.Sleep(100 * time.Millisecond)

	// Simulate job event through chain client
	job := createTestJob("chain-test")
	if err := chainClient.SimulateJobEvent(job); err != nil {
		t.Fatalf("SimulateJobEvent() error = %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := sub.GetStats()
	if stats.JobsReceived != 1 {
		t.Errorf("Expected JobsReceived to be 1, got %d", stats.JobsReceived)
	}

	if stats.JobsProcessed != 1 {
		t.Errorf("Expected JobsProcessed to be 1, got %d", stats.JobsProcessed)
	}
}

func TestSubscriberStats_Fields(t *testing.T) {
	stats := &SubscriberStats{
		JobsReceived:     10,
		JobsProcessed:    8,
		CancelsReceived:  5,
		CancelsProcessed: 4,
		ProcessingErrors: 3,
		ReconnectCount:   1,
		LastEventTime:    time.Now(),
		LastErrorTime:    time.Now().Add(-time.Hour),
		LastError:        "connection timeout",
		StartTime:        time.Now().Add(-2 * time.Hour),
		Uptime:           "2h0m0s",
	}

	if stats.JobsReceived != 10 {
		t.Errorf("JobsReceived = %d, want 10", stats.JobsReceived)
	}

	if stats.JobsProcessed != 8 {
		t.Errorf("JobsProcessed = %d, want 8", stats.JobsProcessed)
	}

	if stats.CancelsReceived != 5 {
		t.Errorf("CancelsReceived = %d, want 5", stats.CancelsReceived)
	}

	if stats.CancelsProcessed != 4 {
		t.Errorf("CancelsProcessed = %d, want 4", stats.CancelsProcessed)
	}

	if stats.ProcessingErrors != 3 {
		t.Errorf("ProcessingErrors = %d, want 3", stats.ProcessingErrors)
	}

	if stats.ReconnectCount != 1 {
		t.Errorf("ReconnectCount = %d, want 1", stats.ReconnectCount)
	}

	if stats.LastError != "connection timeout" {
		t.Errorf("LastError = %s, want 'connection timeout'", stats.LastError)
	}

	if stats.Uptime != "2h0m0s" {
		t.Errorf("Uptime = %s, want '2h0m0s'", stats.Uptime)
	}
}

func TestHPCEventType_Constants(t *testing.T) {
	if HPCEventTypeJobCreated != "hpc_job_created" {
		t.Errorf("HPCEventTypeJobCreated = %s, want 'hpc_job_created'", HPCEventTypeJobCreated)
	}

	if HPCEventTypeJobCancelled != "hpc_job_cancelled" {
		t.Errorf("HPCEventTypeJobCancelled = %s, want 'hpc_job_cancelled'", HPCEventTypeJobCancelled)
	}

	if HPCEventTypeJobUpdated != "hpc_job_updated" {
		t.Errorf("HPCEventTypeJobUpdated = %s, want 'hpc_job_updated'", HPCEventTypeJobUpdated)
	}
}

func TestHPCChainEvent_Fields(t *testing.T) {
	now := time.Now()
	job := createTestJob("event-test")

	event := HPCChainEvent{
		Type:        HPCEventTypeJobCreated,
		JobID:       "test-job-1",
		ClusterID:   chainSubscriberTestClusterEventID,
		BlockHeight: 12345,
		Timestamp:   now,
		Job:         job,
	}

	if event.Type != HPCEventTypeJobCreated {
		t.Errorf("Type = %s, want 'hpc_job_created'", event.Type)
	}

	if event.JobID != "test-job-1" {
		t.Errorf("JobID = %s, want 'test-job-1'", event.JobID)
	}

	if event.ClusterID != chainSubscriberTestClusterEventID {
		t.Errorf("ClusterID = %s, want %s", event.ClusterID, chainSubscriberTestClusterEventID)
	}

	if event.BlockHeight != 12345 {
		t.Errorf("BlockHeight = %d, want 12345", event.BlockHeight)
	}

	if event.Job == nil {
		t.Error("Job should not be nil")
	}
}
