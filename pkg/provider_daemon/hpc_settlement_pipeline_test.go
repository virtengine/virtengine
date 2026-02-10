package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHPCBatchSettlementConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  HPCBatchSettlementConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultHPCBatchSettlementConfig(),
			wantErr: false,
		},
		{
			name: "invalid batch size",
			config: HPCBatchSettlementConfig{
				BatchSize:         0,
				BatchInterval:     time.Minute,
				MaxRetries:        3,
				RetryBackoff:      time.Second,
				MaxPendingRecords: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid batch interval",
			config: HPCBatchSettlementConfig{
				BatchSize:         10,
				BatchInterval:     0,
				MaxRetries:        3,
				RetryBackoff:      time.Second,
				MaxPendingRecords: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid retry backoff",
			config: HPCBatchSettlementConfig{
				BatchSize:         10,
				BatchInterval:     time.Minute,
				MaxRetries:        3,
				RetryBackoff:      0,
				MaxPendingRecords: 100,
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: HPCBatchSettlementConfig{
				BatchSize:         10,
				BatchInterval:     time.Minute,
				MaxRetries:        -1,
				RetryBackoff:      time.Second,
				MaxPendingRecords: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid max pending records",
			config: HPCBatchSettlementConfig{
				BatchSize:         10,
				BatchInterval:     time.Minute,
				MaxRetries:        3,
				RetryBackoff:      time.Second,
				MaxPendingRecords: 0,
			},
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

func TestSettlementRecordStatus_IsValid(t *testing.T) {
	tests := []struct {
		status SettlementRecordStatus
		valid  bool
	}{
		{SettlementRecordStatusPending, true},
		{SettlementRecordStatusSubmitted, true},
		{SettlementRecordStatusConfirmed, true},
		{SettlementRecordStatusFailed, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestSettlementRecordStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   SettlementRecordStatus
		terminal bool
	}{
		{SettlementRecordStatusPending, false},
		{SettlementRecordStatusSubmitted, false},
		{SettlementRecordStatusConfirmed, true},
		{SettlementRecordStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.terminal)
			}
		})
	}
}

func TestHPCSettlementRecord_Validate(t *testing.T) {
	validMetrics := &HPCSchedulerMetrics{
		WallClockSeconds: 3600,
		CPUCoreSeconds:   7200,
	}

	tests := []struct {
		name    string
		record  HPCSettlementRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    validMetrics,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: false,
		},
		{
			name: "empty job id",
			record: HPCSettlementRecord{
				JobID:           "",
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    validMetrics,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: true,
		},
		{
			name: "empty cluster id",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    validMetrics,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: true,
		},
		{
			name: "empty provider address",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: "",
				CustomerAddress: "ve1customer",
				UsageMetrics:    validMetrics,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: true,
		},
		{
			name: "empty customer address",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "",
				UsageMetrics:    validMetrics,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: true,
		},
		{
			name: "nil usage metrics",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    nil,
				Status:          SettlementRecordStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			record: HPCSettlementRecord{
				JobID:           "job-1",
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    validMetrics,
				Status:          "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHPCSettlementRecord_CanRetry(t *testing.T) {
	tests := []struct {
		name       string
		status     SettlementRecordStatus
		attempts   int
		maxRetries int
		want       bool
	}{
		{"failed with attempts left", SettlementRecordStatusFailed, 1, 3, true},
		{"failed at max attempts", SettlementRecordStatusFailed, 3, 3, false},
		{"failed over max attempts", SettlementRecordStatusFailed, 5, 3, false},
		{"pending cannot retry", SettlementRecordStatusPending, 0, 3, false},
		{"submitted cannot retry", SettlementRecordStatusSubmitted, 0, 3, false},
		{"confirmed cannot retry", SettlementRecordStatusConfirmed, 0, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &HPCSettlementRecord{
				Status:   tt.status,
				Attempts: tt.attempts,
			}
			if got := record.CanRetry(tt.maxRetries); got != tt.want {
				t.Errorf("CanRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHPCSettlementRecord_Hash(t *testing.T) {
	record1 := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
	}

	record2 := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
	}

	record3 := &HPCSettlementRecord{
		JobID:           "job-2",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
	}

	hash1 := record1.Hash()
	hash2 := record2.Hash()
	hash3 := record3.Hash()

	if hash1 == "" {
		t.Error("Hash() returned empty string")
	}

	if hash1 != hash2 {
		t.Error("Same records should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different records should produce different hash")
	}
}

// MockHPCOnChainReporter is a mock implementation for testing
type MockHPCOnChainReporter struct {
	mu             sync.Mutex
	StatusReports  []*HPCStatusReport
	AccountingJobs []string
	AccountingErr  error
	StatusErr      error
}

func NewMockHPCOnChainReporter() *MockHPCOnChainReporter {
	return &MockHPCOnChainReporter{
		StatusReports:  make([]*HPCStatusReport, 0),
		AccountingJobs: make([]string, 0),
	}
}

func (m *MockHPCOnChainReporter) ReportJobStatus(ctx context.Context, report *HPCStatusReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StatusErr != nil {
		return m.StatusErr
	}
	m.StatusReports = append(m.StatusReports, report)
	return nil
}

func (m *MockHPCOnChainReporter) ReportJobAccounting(ctx context.Context, jobID string, metrics *HPCSchedulerMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.AccountingErr != nil {
		return m.AccountingErr
	}
	m.AccountingJobs = append(m.AccountingJobs, jobID)
	return nil
}

func (m *MockHPCOnChainReporter) GetAccountingCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.AccountingJobs)
}

func (m *MockHPCOnChainReporter) SetAccountingErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AccountingErr = err
}

func TestNewHPCBatchSettlementPipeline(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	if pipeline == nil {
		t.Fatal("NewHPCBatchSettlementPipeline returned nil")
	}

	if pipeline.config.BatchSize != config.BatchSize {
		t.Error("Config not set correctly")
	}

	if pipeline.pending == nil {
		t.Error("pending map not initialized")
	}

	if pipeline.submitted == nil {
		t.Error("submitted map not initialized")
	}
}

func TestHPCBatchSettlementPipeline_QueueSettlement(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	record := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		UsageMetrics: &HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   7200,
		},
		Status: SettlementRecordStatusPending,
	}

	err := pipeline.QueueSettlement(record)
	if err != nil {
		t.Fatalf("QueueSettlement failed: %v", err)
	}

	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected 1 pending record, got %d", pipeline.GetPendingCount())
	}

	// Queue same record again should be idempotent
	err = pipeline.QueueSettlement(record)
	if err != nil {
		t.Fatalf("QueueSettlement failed on duplicate: %v", err)
	}

	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected still 1 pending record after duplicate, got %d", pipeline.GetPendingCount())
	}
}

func TestHPCBatchSettlementPipeline_QueueSettlement_Validation(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Nil record
	err := pipeline.QueueSettlement(nil)
	if err == nil {
		t.Error("expected error for nil record")
	}

	// Invalid record
	err = pipeline.QueueSettlement(&HPCSettlementRecord{
		JobID: "", // Invalid: empty
	})
	if err == nil {
		t.Error("expected error for invalid record")
	}
}

func TestHPCBatchSettlementPipeline_QueueSettlement_MaxPending(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.MaxPendingRecords = 3
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	for i := 0; i < 3; i++ {
		record := &HPCSettlementRecord{
			JobID:           "job-" + string(rune('a'+i)),
			ClusterID:       "cluster-1",
			ProviderAddress: "ve1provider",
			CustomerAddress: "ve1customer",
			UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: int64(i)},
			Status:          SettlementRecordStatusPending,
		}
		err := pipeline.QueueSettlement(record)
		if err != nil {
			t.Fatalf("QueueSettlement failed for record %d: %v", i, err)
		}
	}

	// Should fail on 4th record
	record := &HPCSettlementRecord{
		JobID:           "job-d",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: 100},
		Status:          SettlementRecordStatusPending,
	}
	err := pipeline.QueueSettlement(record)
	if err == nil {
		t.Error("expected error when max pending reached")
	}
}

func TestHPCBatchSettlementPipeline_StartStop(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.BatchInterval = time.Millisecond * 100
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !pipeline.IsRunning() {
		t.Error("pipeline should be running after Start")
	}

	// Start again should be idempotent
	err = pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed on second call: %v", err)
	}

	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if pipeline.IsRunning() {
		t.Error("pipeline should not be running after Stop")
	}

	// Stop again should be idempotent
	err = pipeline.Stop()
	if err != nil {
		t.Fatalf("Stop failed on second call: %v", err)
	}
}

func TestHPCBatchSettlementPipeline_ProcessBatch(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.BatchInterval = time.Millisecond * 50
	config.BatchSize = 2
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Queue 3 records
	for i := 0; i < 3; i++ {
		record := &HPCSettlementRecord{
			JobID:           "job-" + string(rune('1'+i)),
			ClusterID:       "cluster-1",
			ProviderAddress: "ve1provider",
			CustomerAddress: "ve1customer",
			UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: int64(i * 1000)},
			Status:          SettlementRecordStatusPending,
		}
		if err := pipeline.QueueSettlement(record); err != nil {
			t.Fatalf("QueueSettlement failed: %v", err)
		}
	}

	ctx := context.Background()

	// Process one batch (should take 2 records)
	pipeline.processBatch(ctx)

	// Check that 2 records were submitted
	if reporter.GetAccountingCount() != 2 {
		t.Errorf("expected 2 accounting reports, got %d", reporter.GetAccountingCount())
	}

	submitted := pipeline.GetSubmittedRecords()
	if len(submitted) != 2 {
		t.Errorf("expected 2 submitted records, got %d", len(submitted))
	}

	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected 1 pending record, got %d", pipeline.GetPendingCount())
	}
}

func TestHPCBatchSettlementPipeline_ProcessBatch_WithErrors(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.BatchSize = 10
	config.MaxRetries = 3
	reporter := NewMockHPCOnChainReporter()
	reporter.SetAccountingErr(errors.New("chain error"))
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	record := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: 3600},
		Status:          SettlementRecordStatusPending,
	}
	if err := pipeline.QueueSettlement(record); err != nil {
		t.Fatalf("QueueSettlement failed: %v", err)
	}

	ctx := context.Background()

	// Process batch - should fail
	pipeline.processBatch(ctx)

	// Record should still be in pending (not at max retries yet)
	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected 1 pending record, got %d", pipeline.GetPendingCount())
	}

	stats := pipeline.GetStats()
	if stats.LastError == "" {
		t.Error("expected LastError to be set")
	}
}

func TestHPCBatchSettlementPipeline_ConfirmSettlement(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Queue and submit a record
	record := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: 3600},
		Status:          SettlementRecordStatusPending,
	}
	if err := pipeline.QueueSettlement(record); err != nil {
		t.Fatalf("QueueSettlement failed: %v", err)
	}

	ctx := context.Background()
	pipeline.processBatch(ctx)

	// Confirm the settlement
	err := pipeline.ConfirmSettlement("job-1", "")
	if err != nil {
		t.Fatalf("ConfirmSettlement failed: %v", err)
	}

	stats := pipeline.GetStats()
	if stats.TotalConfirmed != 1 {
		t.Errorf("expected 1 confirmed, got %d", stats.TotalConfirmed)
	}

	if len(pipeline.GetSubmittedRecords()) != 0 {
		t.Error("expected no submitted records after confirmation")
	}
}

func TestHPCBatchSettlementPipeline_ConfirmSettlement_NotFound(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	err := pipeline.ConfirmSettlement("nonexistent-job", "")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestHPCBatchSettlementPipeline_RetryFailed(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.MaxRetries = 3
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Manually add a failed record
	record := &HPCSettlementRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: 3600},
		Status:          SettlementRecordStatusFailed,
		Attempts:        1, // Still can retry
		LastAttempt:     time.Now().Add(-time.Hour),
		CreatedAt:       time.Now(),
	}

	pipeline.mu.Lock()
	pipeline.failed[record.Hash()] = record
	pipeline.mu.Unlock()

	retried := pipeline.RetryFailed()
	if retried != 1 {
		t.Errorf("expected 1 retried, got %d", retried)
	}

	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected 1 pending after retry, got %d", pipeline.GetPendingCount())
	}

	if len(pipeline.GetFailedRecords()) != 0 {
		t.Error("expected no failed records after retry")
	}
}

func TestHPCBatchSettlementPipeline_GetStats(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Queue some records
	for i := 0; i < 5; i++ {
		record := &HPCSettlementRecord{
			JobID:           "job-" + string(rune('a'+i)),
			ClusterID:       "cluster-1",
			ProviderAddress: "ve1provider",
			CustomerAddress: "ve1customer",
			UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: int64(i * 1000)},
			Status:          SettlementRecordStatusPending,
		}
		if err := pipeline.QueueSettlement(record); err != nil {
			t.Fatalf("QueueSettlement failed: %v", err)
		}
	}

	stats := pipeline.GetStats()

	if stats.TotalQueued != 5 {
		t.Errorf("expected TotalQueued=5, got %d", stats.TotalQueued)
	}

	if stats.PendingCount != 5 {
		t.Errorf("expected PendingCount=5, got %d", stats.PendingCount)
	}
}

func TestHPCBatchSettlementPipeline_ForceFlush(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.BatchSize = 2
	config.BatchInterval = time.Hour // Long interval so normal batch doesn't trigger
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Queue records
	for i := 0; i < 5; i++ {
		record := &HPCSettlementRecord{
			JobID:           "job-" + string(rune('1'+i)),
			ClusterID:       "cluster-1",
			ProviderAddress: "ve1provider",
			CustomerAddress: "ve1customer",
			UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: int64(i * 1000)},
			Status:          SettlementRecordStatusPending,
		}
		if err := pipeline.QueueSettlement(record); err != nil {
			t.Fatalf("QueueSettlement failed: %v", err)
		}
	}

	// Force flush all
	err = pipeline.ForceFlush(ctx)
	if err != nil {
		t.Fatalf("ForceFlush failed: %v", err)
	}

	if pipeline.GetPendingCount() != 0 {
		t.Errorf("expected 0 pending after flush, got %d", pipeline.GetPendingCount())
	}

	// All 5 should be submitted
	if len(pipeline.GetSubmittedRecords()) != 5 {
		t.Errorf("expected 5 submitted, got %d", len(pipeline.GetSubmittedRecords()))
	}
}

func TestHPCBatchSettlementPipeline_ForceFlush_NotRunning(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	err := pipeline.ForceFlush(context.Background())
	if err == nil {
		t.Error("expected error when pipeline not running")
	}
}

func TestHPCBatchSettlementPipeline_ClearConfirmed(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Add a confirmed record directly
	pipeline.mu.Lock()
	pipeline.confirmed["test-hash"] = &HPCSettlementRecord{
		JobID:  "job-1",
		Status: SettlementRecordStatusConfirmed,
	}
	pipeline.mu.Unlock()

	pipeline.ClearConfirmed()

	pipeline.mu.RLock()
	count := len(pipeline.confirmed)
	pipeline.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 confirmed after clear, got %d", count)
	}
}

func TestHPCBatchSettlementPipeline_ClearFailed(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	// Add a failed record directly
	pipeline.mu.Lock()
	pipeline.failed["test-hash"] = &HPCSettlementRecord{
		JobID:  "job-1",
		Status: SettlementRecordStatusFailed,
	}
	pipeline.mu.Unlock()

	pipeline.ClearFailed()

	if len(pipeline.GetFailedRecords()) != 0 {
		t.Error("expected 0 failed after clear")
	}
}

func TestHPCBatchSettlementPipeline_DisabledConfig(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.Enabled = false
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	ctx := context.Background()

	// Start should be no-op when disabled
	err := pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Should not be running
	if pipeline.IsRunning() {
		t.Error("pipeline should not be running when disabled")
	}
}

func TestHPCBatchSettlementPipeline_Concurrent(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()
	config.MaxPendingRecords = 1000
	reporter := NewMockHPCOnChainReporter()
	signer := NewMockSigner("ve1provider")

	pipeline := NewHPCBatchSettlementPipeline(config, reporter, signer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pipeline.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Concurrent queue operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			record := &HPCSettlementRecord{
				JobID:           fmt.Sprintf("job-%d", idx),
				ClusterID:       "cluster-1",
				ProviderAddress: "ve1provider",
				CustomerAddress: "ve1customer",
				UsageMetrics:    &HPCSchedulerMetrics{WallClockSeconds: int64(idx * 100)},
				Status:          SettlementRecordStatusPending,
			}
			_ = pipeline.QueueSettlement(record)
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and some records were queued
	stats := pipeline.GetStats()
	if stats.TotalQueued == 0 {
		t.Error("expected some records to be queued")
	}
}

func TestDefaultHPCBatchSettlementConfig(t *testing.T) {
	config := DefaultHPCBatchSettlementConfig()

	if !config.Enabled {
		t.Error("default config should be enabled")
	}

	if config.BatchSize != 50 {
		t.Errorf("expected BatchSize=50, got %d", config.BatchSize)
	}

	if config.BatchInterval != time.Minute*5 {
		t.Errorf("expected BatchInterval=5m, got %v", config.BatchInterval)
	}

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.RetryBackoff != time.Second*5 {
		t.Errorf("expected RetryBackoff=5s, got %v", config.RetryBackoff)
	}

	if config.MaxPendingRecords != 100 {
		t.Errorf("expected MaxPendingRecords=100, got %d", config.MaxPendingRecords)
	}

	// Validate should pass
	if err := config.Validate(); err != nil {
		t.Errorf("default config validation failed: %v", err)
	}
}

func TestCreateSettlementRecordFromUsage(t *testing.T) {
	usage := &HPCUsageRecord{
		RecordID:        "record-1",
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: "ve1provider",
		CustomerAddress: "ve1customer",
		Metrics: &HPCSchedulerMetrics{
			WallClockSeconds: 3600,
			CPUCoreSeconds:   7200,
		},
		IsFinal: true,
	}

	record := CreateSettlementRecordFromUsage(usage)

	if record == nil {
		t.Fatal("expected non-nil record")
	}

	if record.JobID != usage.JobID {
		t.Errorf("expected JobID=%s, got %s", usage.JobID, record.JobID)
	}

	if record.ClusterID != usage.ClusterID {
		t.Errorf("expected ClusterID=%s, got %s", usage.ClusterID, record.ClusterID)
	}

	if record.Status != SettlementRecordStatusPending {
		t.Errorf("expected status=pending, got %s", record.Status)
	}

	if !record.IsFinal {
		t.Error("expected IsFinal=true")
	}

	if len(record.UsageRecordIDs) != 1 || record.UsageRecordIDs[0] != usage.RecordID {
		t.Error("expected UsageRecordIDs to contain record ID")
	}
}

func TestCreateSettlementRecordFromUsage_Nil(t *testing.T) {
	record := CreateSettlementRecordFromUsage(nil)
	if record != nil {
		t.Error("expected nil for nil input")
	}
}
