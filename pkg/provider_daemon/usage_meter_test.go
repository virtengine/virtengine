package provider_daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMetricsCollector is a mock implementation of MetricsCollector
type MockMetricsCollector struct {
	mu      sync.Mutex
	metrics map[string]*ResourceMetrics
	calls   int
}

func NewMockMetricsCollector() *MockMetricsCollector {
	return &MockMetricsCollector{
		metrics: make(map[string]*ResourceMetrics),
	}
}

func (m *MockMetricsCollector) CollectMetrics(ctx context.Context, workloadID string) (*ResourceMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++

	if metrics, ok := m.metrics[workloadID]; ok {
		return metrics, nil
	}

	// Return default metrics
	return &ResourceMetrics{
		CPUMilliSeconds:    1000,
		MemoryByteSeconds:  1024 * 1024,
		StorageByteSeconds: 1024 * 1024,
		NetworkBytesIn:     10000,
		NetworkBytesOut:    5000,
		GPUSeconds:         0,
	}, nil
}

func (m *MockMetricsCollector) SetMetrics(workloadID string, metrics *ResourceMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[workloadID] = metrics
}

func (m *MockMetricsCollector) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// MockChainRecorder is a mock implementation of ChainRecorder
type MockChainRecorder struct {
	mu           sync.Mutex
	records      []*UsageRecord
	finalRecords []*UsageRecord
	failOnSubmit bool
}

func NewMockChainRecorder() *MockChainRecorder {
	return &MockChainRecorder{
		records:      make([]*UsageRecord, 0),
		finalRecords: make([]*UsageRecord, 0),
	}
}

func (m *MockChainRecorder) SubmitUsageRecord(ctx context.Context, record *UsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnSubmit {
		return ErrMeteringNotStarted
	}
	m.records = append(m.records, record)
	return nil
}

func (m *MockChainRecorder) SubmitFinalSettlement(ctx context.Context, record *UsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failOnSubmit {
		return ErrMeteringNotStarted
	}
	m.finalRecords = append(m.finalRecords, record)
	return nil
}

func (m *MockChainRecorder) GetRecords() []*UsageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.records
}

func (m *MockChainRecorder) GetFinalRecords() []*UsageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.finalRecords
}

func TestUsageMeterStartStop(t *testing.T) {
	collector := NewMockMetricsCollector()
	recorder := NewMockChainRecorder()

	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		Interval:         MeteringIntervalMinute,
		MetricsCollector: collector,
		ChainRecorder:    recorder,
	})

	ctx := context.Background()
	err := meter.Start(ctx)
	require.NoError(t, err)

	// Starting again should be a no-op
	err = meter.Start(ctx)
	require.NoError(t, err)

	meter.Stop()
}

func TestUsageMeterStartMetering(t *testing.T) {
	collector := NewMockMetricsCollector()
	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
	})

	err := meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{
		AgreedCPURate:    "0.001",
		AgreedMemoryRate: "0.0001",
	})
	require.NoError(t, err)

	state, err := meter.GetMeteringState("workload-1")
	require.NoError(t, err)
	assert.Equal(t, "workload-1", state.WorkloadID)
	assert.Equal(t, "deployment-1", state.DeploymentID)
	assert.Equal(t, "lease-1", state.LeaseID)
	assert.True(t, state.Active)
}

func TestUsageMeterStopMetering(t *testing.T) {
	collector := NewMockMetricsCollector()
	recorder := NewMockChainRecorder()

	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
		ChainRecorder:    recorder,
	})

	// Start metering
	err := meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{})
	require.NoError(t, err)

	ctx := context.Background()

	// Stop metering
	record, err := meter.StopMetering(ctx, "workload-1")
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, UsageRecordTypeFinal, record.Type)
	assert.Equal(t, "workload-1", record.WorkloadID)
	assert.Equal(t, "deployment-1", record.DeploymentID)

	// Final record should be submitted
	finalRecords := recorder.GetFinalRecords()
	assert.Len(t, finalRecords, 1)

	// Workload should no longer be metered
	_, err = meter.GetMeteringState("workload-1")
	assert.Error(t, err)
	assert.Equal(t, ErrWorkloadNotMetered, err)
}

func TestUsageMeterStopMeteringNotMetered(t *testing.T) {
	collector := NewMockMetricsCollector()
	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
	})

	ctx := context.Background()
	_, err := meter.StopMetering(ctx, "non-existent")
	require.Error(t, err)
	assert.Equal(t, ErrWorkloadNotMetered, err)
}

func TestUsageMeterListMeteredWorkloads(t *testing.T) {
	collector := NewMockMetricsCollector()
	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
	})

	_ = meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{})
	_ = meter.StartMetering("workload-2", "deployment-2", "lease-2", PricingInputs{})
	_ = meter.StartMetering("workload-3", "deployment-3", "lease-3", PricingInputs{})

	workloads := meter.ListMeteredWorkloads()
	assert.Len(t, workloads, 3)
}

func TestUsageMeterForceCollect(t *testing.T) {
	collector := NewMockMetricsCollector()
	collector.SetMetrics("workload-1", &ResourceMetrics{
		CPUMilliSeconds:    5000,
		MemoryByteSeconds:  2 * 1024 * 1024,
		StorageByteSeconds: 1024 * 1024,
		NetworkBytesIn:     20000,
		NetworkBytesOut:    10000,
	})

	recorder := NewMockChainRecorder()
	recordChan := make(chan *UsageRecord, 10)

	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
		ChainRecorder:    recorder,
		RecordChan:       recordChan,
	})

	// Start metering
	err := meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{
		AgreedCPURate: "0.001",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Force collect
	record, err := meter.ForceCollect(ctx, "workload-1")
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, UsageRecordTypePeriodic, record.Type)
	assert.Equal(t, int64(5000), record.Metrics.CPUMilliSeconds)
	assert.Equal(t, "provider-123", record.ProviderID)

	// Record should be submitted to chain
	records := recorder.GetRecords()
	assert.Len(t, records, 1)

	// Record should be sent to channel
	select {
	case r := <-recordChan:
		assert.Equal(t, record.ID, r.ID)
	default:
		t.Error("Expected record in channel")
	}
}

func TestUsageMeterForceCollectNotMetered(t *testing.T) {
	collector := NewMockMetricsCollector()
	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
	})

	ctx := context.Background()
	_, err := meter.ForceCollect(ctx, "non-existent")
	require.Error(t, err)
	assert.Equal(t, ErrWorkloadNotMetered, err)
}

func TestUsageMeterCumulativeMetrics(t *testing.T) {
	collector := NewMockMetricsCollector()
	collector.SetMetrics("workload-1", &ResourceMetrics{
		CPUMilliSeconds:   1000,
		MemoryByteSeconds: 1024,
	})

	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
	})

	err := meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{})
	require.NoError(t, err)

	ctx := context.Background()

	// Force collect multiple times
	for i := 0; i < 5; i++ {
		_, err := meter.ForceCollect(ctx, "workload-1")
		require.NoError(t, err)
	}

	// Check cumulative metrics
	state, err := meter.GetMeteringState("workload-1")
	require.NoError(t, err)

	assert.Equal(t, int64(5000), state.CumulativeMetrics.CPUMilliSeconds)
	assert.Equal(t, int64(5120), state.CumulativeMetrics.MemoryByteSeconds)
}

func TestUsageRecordHash(t *testing.T) {
	record := &UsageRecord{
		WorkloadID:   "workload-1",
		DeploymentID: "deployment-1",
		LeaseID:      "lease-1",
		ProviderID:   "provider-123",
		Type:         UsageRecordTypePeriodic,
		StartTime:    time.Unix(1000, 0),
		EndTime:      time.Unix(2000, 0),
		Metrics: ResourceMetrics{
			CPUMilliSeconds:   5000,
			MemoryByteSeconds: 1024,
		},
	}

	hash1 := record.Hash()
	hash2 := record.Hash()

	// Same input should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different input should produce different hash
	record.Metrics.CPUMilliSeconds = 6000
	hash3 := record.Hash()
	assert.NotEqual(t, hash1, hash3)
}

func TestUsageMeterWithKeyManager(t *testing.T) {
	collector := NewMockMetricsCollector()
	keyManager, err := NewKeyManager(KeyManagerConfig{
		StorageType:      KeyStorageTypeMemory,
		DefaultAlgorithm: "ed25519",
	})
	require.NoError(t, err)

	// Unlock the key manager (memory storage doesn't need passphrase)
	err = keyManager.Unlock("")
	require.NoError(t, err)

	// Generate a key
	_, err = keyManager.GenerateKey("test-provider-address")
	require.NoError(t, err)

	meter := NewUsageMeter(UsageMeterConfig{
		ProviderID:       "provider-123",
		MetricsCollector: collector,
		KeyManager:       keyManager,
	})

	ctx := context.Background()
	err = meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{})
	require.NoError(t, err)

	record, err := meter.ForceCollect(ctx, "workload-1")
	require.NoError(t, err)

	// Record should be signed
	assert.NotEmpty(t, record.Signature)
}

func TestFraudCheckerValid(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Metrics: ResourceMetrics{
			CPUMilliSeconds:    3600000, // 1 hour of CPU usage
			MemoryByteSeconds:  1024 * 1024 * 3600,
			StorageByteSeconds: 1024 * 1024 * 3600,
			NetworkBytesIn:     1000000,
			NetworkBytesOut:    500000,
		},
	}

	// Allocated resources should make the usage ratio reasonable (< 2x)
	// Expected CPU = allocated * duration / 1000 = 1000000 * 3600 / 1000 = 3600000
	// So ratio = 3600000 / 3600000 = 1.0 (valid)
	allocated := &ResourceMetrics{
		CPUMilliSeconds:   1000000, // 1000 CPUs allocated (matching usage)
		MemoryByteSeconds: 1024 * 1024,
	}

	result := checker.CheckRecord(record, allocated)
	assert.True(t, result.Valid)
	assert.Equal(t, 0, result.Score)
	assert.Empty(t, result.Flags)
}

func TestFraudCheckerDurationTooShort(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now().Add(-30 * time.Second),
		EndTime:   time.Now(),
		Metrics:   ResourceMetrics{},
	}

	result := checker.CheckRecord(record, nil)
	// Duration too short is a warning (score 30), not invalid (threshold 50)
	assert.True(t, result.Valid)
	assert.Contains(t, result.Flags, "DURATION_TOO_SHORT")
	assert.Equal(t, 30, result.Score)
}

func TestFraudCheckerDurationTooLong(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now().Add(-48 * time.Hour),
		EndTime:   time.Now(),
		Metrics:   ResourceMetrics{},
	}

	result := checker.CheckRecord(record, nil)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Flags, "DURATION_TOO_LONG")
}

func TestFraudCheckerFutureTimestamp(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		Metrics:   ResourceMetrics{},
	}

	result := checker.CheckRecord(record, nil)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Flags, "FUTURE_TIMESTAMP")
}

func TestFraudCheckerExcessiveCPU(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 10000000, // Way more than allocated
		},
	}

	allocated := &ResourceMetrics{
		CPUMilliSeconds: 1000,
	}

	result := checker.CheckRecord(record, allocated)
	// EXCESSIVE_CPU_USAGE adds score 40, which is below invalidity threshold 50
	assert.True(t, result.Valid)
	assert.Contains(t, result.Flags, "EXCESSIVE_CPU_USAGE")
	assert.Equal(t, 40, result.Score)
}

func TestFraudCheckerNegativeMetrics(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now(),
		Metrics: ResourceMetrics{
			CPUMilliSeconds: -1000,
		},
	}

	result := checker.CheckRecord(record, nil)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Flags, "NEGATIVE_METRICS")
}

func TestFraudCheckerZeroDurationWithUsage(t *testing.T) {
	checker := NewFraudChecker()

	now := time.Now()
	record := &UsageRecord{
		StartTime: now,
		EndTime:   now,
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 1000,
		},
	}

	result := checker.CheckRecord(record, nil)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Flags, "ZERO_DURATION_WITH_USAGE")
}

func TestFraudCheckerScoreCap(t *testing.T) {
	checker := NewFraudChecker()

	now := time.Now()
	record := &UsageRecord{
		StartTime: now,
		EndTime:   now.Add(time.Hour), // Future
		Metrics: ResourceMetrics{
			CPUMilliSeconds: -1000, // Negative
		},
	}

	result := checker.CheckRecord(record, nil)
	assert.False(t, result.Valid)
	assert.LessOrEqual(t, result.Score, 100)
}

func TestFraudCheckerSignatureVerification(t *testing.T) {
	checker := NewFraudChecker()

	record := &UsageRecord{
		Signature: "abcd1234",
	}

	// Valid hex signature
	assert.True(t, checker.CheckRecordSignature(record, []byte("public-key")))

	// Empty signature
	record.Signature = ""
	assert.False(t, checker.CheckRecordSignature(record, []byte("public-key")))

	// Invalid hex
	record.Signature = "not-hex!"
	assert.False(t, checker.CheckRecordSignature(record, []byte("public-key")))
}

func TestMeteringLoopIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	collector := NewMockMetricsCollector()
	recorder := NewMockChainRecorder()

	// Use a very short interval for testing
	meter := &UsageMeter{
		providerID: "provider-123",
		interval:   100 * time.Millisecond,
		collector:  collector,
		recorder:   recorder,
		workloads:  make(map[string]*WorkloadMetering),
		stopChan:   make(chan struct{}),
	}

	// Start metering for a workload
	_ = meter.StartMetering("workload-1", "deployment-1", "lease-1", PricingInputs{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = meter.Start(ctx)

	// Wait for a few collection cycles
	time.Sleep(350 * time.Millisecond)

	meter.Stop()

	// Should have collected at least 2-3 times
	records := recorder.GetRecords()
	assert.GreaterOrEqual(t, len(records), 2)
}
