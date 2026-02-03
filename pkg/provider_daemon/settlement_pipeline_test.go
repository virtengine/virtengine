package provider_daemon

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

func TestSettlementPipeline_AddPendingUsage(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	record := &UsageRecord{
		ID:           "test-record-1",
		WorkloadID:   "workload-1",
		DeploymentID: "deployment-1",
		LeaseID:      "lease-1",
		ProviderID:   "provider-1",
		Type:         UsageRecordTypePeriodic,
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now(),
		Metrics: ResourceMetrics{
			CPUMilliSeconds:    3600000,
			MemoryByteSeconds:  1073741824 * 3600,
			StorageByteSeconds: 10737418240 * 3600,
			GPUSeconds:         3600,
		},
	}

	pipeline.AddPendingUsage(record)

	if pipeline.GetPendingCount() != 1 {
		t.Errorf("expected 1 pending record, got %d", pipeline.GetPendingCount())
	}
}

func TestSettlementPipeline_CreateDispute(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	record := &UsageRecord{
		ID:           "test-record-1",
		WorkloadID:   "workload-1",
		DeploymentID: "deployment-1",
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 3600000,
		},
	}
	pipeline.AddPendingUsage(record)

	expectedMetrics := &ResourceMetrics{
		CPUMilliSeconds: 1800000, // Expected half of reported
	}

	dispute, err := pipeline.CreateDispute(
		"test-record-1",
		"order-1",
		"customer-address",
		"CPU usage seems too high",
		"",
		expectedMetrics,
	)

	if err != nil {
		t.Fatalf("failed to create dispute: %v", err)
	}

	if dispute.Status != DisputeStatusPending {
		t.Errorf("expected pending status, got %s", dispute.Status)
	}

	if dispute.UsageRecordID != "test-record-1" {
		t.Errorf("expected usage record ID 'test-record-1', got %s", dispute.UsageRecordID)
	}

	activeDisputes := pipeline.GetActiveDisputes()
	if len(activeDisputes) != 1 {
		t.Errorf("expected 1 active dispute, got %d", len(activeDisputes))
	}
}

func TestSettlementPipeline_ResolveDispute(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	record := &UsageRecord{
		ID:           "test-record-1",
		WorkloadID:   "workload-1",
		DeploymentID: "deployment-1",
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 3600000,
		},
	}
	pipeline.AddPendingUsage(record)

	expectedMetrics := &ResourceMetrics{
		CPUMilliSeconds: 1800000,
	}

	dispute, _ := pipeline.CreateDispute(
		"test-record-1",
		"order-1",
		"customer-address",
		"CPU usage too high",
		"",
		expectedMetrics,
	)

	err := pipeline.ResolveDispute(dispute.DisputeID, "Corrected to expected value", true)
	if err != nil {
		t.Fatalf("failed to resolve dispute: %v", err)
	}

	activeDisputes := pipeline.GetActiveDisputes()
	if len(activeDisputes) != 0 {
		t.Errorf("expected 0 active disputes after resolution, got %d", len(activeDisputes))
	}
}

func TestSettlementPipeline_ProcessUsageToLineItems(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	now := time.Now()
	record := &UsageRecord{
		ID:           "test-record-1",
		WorkloadID:   "workload-1",
		DeploymentID: "deployment-1",
		LeaseID:      "lease-1",
		StartTime:    now.Add(-time.Hour),
		EndTime:      now,
		Metrics: ResourceMetrics{
			CPUMilliSeconds:    3600000,            // 1 CPU-hour
			MemoryByteSeconds:  1073741824 * 3600,  // 1 GB-hour
			StorageByteSeconds: 10737418240 * 3600, // 10 GB-hours
			GPUSeconds:         3600,               // 1 GPU-hour
			NetworkBytesIn:     1073741824,         // 1 GB in
			NetworkBytesOut:    1073741824,         // 1 GB out
		},
		PricingInputs: PricingInputs{
			AgreedCPURate:     "0.01",
			AgreedMemoryRate:  "0.005",
			AgreedStorageRate: "0.001",
			AgreedGPURate:     "0.5",
			AgreedNetworkRate: "0.02",
		},
	}

	lineItems, err := pipeline.ProcessUsageToLineItems(record)
	if err != nil {
		t.Fatalf("failed to process usage: %v", err)
	}

	if len(lineItems) != 5 {
		t.Errorf("expected 5 line items, got %d", len(lineItems))
	}

	// Check that line items have correct resource types
	resourceTypes := make(map[string]bool)
	for _, item := range lineItems {
		resourceTypes[item.ResourceType] = true
		if item.TotalCost.IsZero() {
			t.Errorf("line item %s has zero cost", item.ResourceType)
		}
	}

	expectedTypes := []string{"cpu", "memory", "storage", "gpu", "network"}
	for _, rt := range expectedTypes {
		if !resourceTypes[rt] {
			t.Errorf("missing line item for resource type: %s", rt)
		}
	}
}

func TestSettlementPipeline_DetectAnomalies(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	tests := []struct {
		name          string
		record        *UsageRecord
		allocated     *ResourceMetrics
		expectedCount int
		expectedTypes []string
	}{
		{
			name: "duration too short",
			record: &UsageRecord{
				ID:        "test-1",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(30 * time.Second), // 30 seconds
				Metrics:   ResourceMetrics{CPUMilliSeconds: 1000},
			},
			expectedCount: 1,
			expectedTypes: []string{"duration_too_short"},
		},
		{
			name: "duration too long",
			record: &UsageRecord{
				ID:        "test-2",
				StartTime: time.Now().Add(-30 * time.Hour),
				EndTime:   time.Now(),
				Metrics:   ResourceMetrics{CPUMilliSeconds: 1000},
			},
			expectedCount: 1,
			expectedTypes: []string{"duration_too_long"},
		},
		{
			name: "negative values",
			record: &UsageRecord{
				ID:        "test-3",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now(),
				Metrics:   ResourceMetrics{CPUMilliSeconds: -1000},
			},
			expectedCount: 1,
			expectedTypes: []string{"negative_values"},
		},
		{
			name: "normal record",
			record: &UsageRecord{
				ID:        "test-4",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now(),
				Metrics:   ResourceMetrics{CPUMilliSeconds: 3600000},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anomalies := pipeline.DetectAnomalies(tt.record, tt.allocated)
			if len(anomalies) != tt.expectedCount {
				t.Errorf("expected %d anomalies, got %d", tt.expectedCount, len(anomalies))
			}

			for i, expectedType := range tt.expectedTypes {
				if i < len(anomalies) && anomalies[i].AnomalyType != expectedType {
					t.Errorf("expected anomaly type %s, got %s", expectedType, anomalies[i].AnomalyType)
				}
			}
		})
	}
}

func TestBillableLineItem_Hash(t *testing.T) {
	item1 := &BillableLineItem{
		OrderID:      "order-1",
		LeaseID:      "lease-1",
		ResourceType: "cpu",
		Quantity:     sdkmath.LegacyNewDec(100),
		PeriodStart:  time.Unix(1000, 0),
		PeriodEnd:    time.Unix(2000, 0),
	}

	item2 := &BillableLineItem{
		OrderID:      "order-1",
		LeaseID:      "lease-1",
		ResourceType: "cpu",
		Quantity:     sdkmath.LegacyNewDec(100),
		PeriodStart:  time.Unix(1000, 0),
		PeriodEnd:    time.Unix(2000, 0),
	}

	item3 := &BillableLineItem{
		OrderID:      "order-2", // Different order
		LeaseID:      "lease-1",
		ResourceType: "cpu",
		Quantity:     sdkmath.LegacyNewDec(100),
		PeriodStart:  time.Unix(1000, 0),
		PeriodEnd:    time.Unix(2000, 0),
	}

	hash1 := item1.Hash()
	hash2 := item2.Hash()
	hash3 := item3.Hash()

	if string(hash1) != string(hash2) {
		t.Error("identical items should have the same hash")
	}

	if string(hash1) == string(hash3) {
		t.Error("different items should have different hashes")
	}
}

func TestUsageDispute_Workflow(t *testing.T) {
	cfg := DefaultSettlementConfig()
	cfg.DisputeWindow = time.Hour
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	// Add usage record
	record := &UsageRecord{
		ID:           "usage-1",
		DeploymentID: "order-1",
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 3600000,
		},
	}
	pipeline.AddPendingUsage(record)

	// Create dispute
	dispute, err := pipeline.CreateDispute(
		"usage-1",
		"order-1",
		"customer-1",
		"Overcharged for CPU",
		"Screenshot attached",
		&ResourceMetrics{CPUMilliSeconds: 1800000},
	)
	if err != nil {
		t.Fatalf("create dispute failed: %v", err)
	}

	// Verify dispute is pending
	if dispute.Status != DisputeStatusPending {
		t.Errorf("expected pending, got %s", dispute.Status)
	}

	// Verify expiration is set
	if dispute.ExpiresAt.Before(time.Now()) {
		t.Error("dispute should not be expired yet")
	}

	// Resolve dispute with acceptance
	err = pipeline.ResolveDispute(dispute.DisputeID, "Correction applied", true)
	if err != nil {
		t.Fatalf("resolve dispute failed: %v", err)
	}

	// Verify dispute is resolved
	pipeline.mu.RLock()
	resolvedDispute := pipeline.disputes[dispute.DisputeID]
	pipeline.mu.RUnlock()

	if resolvedDispute.Status != DisputeStatusResolved {
		t.Errorf("expected resolved, got %s", resolvedDispute.Status)
	}

	if resolvedDispute.ResolvedAt == nil {
		t.Error("resolved_at should be set")
	}
}

func TestUsageCorrection_AppliesCorrectMetrics(t *testing.T) {
	cfg := DefaultSettlementConfig()
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), nil)

	// Add usage record with original metrics
	originalMetrics := ResourceMetrics{
		CPUMilliSeconds:   3600000,
		MemoryByteSeconds: 1073741824 * 3600,
	}
	record := &UsageRecord{
		ID:           "usage-1",
		DeploymentID: "order-1",
		Metrics:      originalMetrics,
	}
	pipeline.AddPendingUsage(record)

	// Create dispute with expected metrics
	expectedMetrics := &ResourceMetrics{
		CPUMilliSeconds:   1800000,           // Half of original
		MemoryByteSeconds: 1073741824 * 1800, // Half
	}

	dispute, _ := pipeline.CreateDispute(
		"usage-1",
		"order-1",
		"customer-1",
		"Metrics incorrect",
		"",
		expectedMetrics,
	)

	// Resolve with acceptance (applies correction)
	err := pipeline.ResolveDispute(dispute.DisputeID, "Applying correction", true)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	// Verify correction was applied
	pipeline.mu.RLock()
	correctedRecord := pipeline.pending["usage-1"]
	numCorrections := len(pipeline.corrections)
	pipeline.mu.RUnlock()

	if numCorrections != 1 {
		t.Errorf("expected 1 correction, got %d", numCorrections)
	}

	if correctedRecord.Metrics.CPUMilliSeconds != expectedMetrics.CPUMilliSeconds {
		t.Errorf("expected corrected CPU %d, got %d",
			expectedMetrics.CPUMilliSeconds, correctedRecord.Metrics.CPUMilliSeconds)
	}
}

// Mock chain submitter for testing
type mockChainSubmitter struct {
	usageReports       []*ChainUsageReport
	settlementRequests []mockSettlementRequest
}

type mockSettlementRequest struct {
	OrderID        string
	UsageRecordIDs []string
	IsFinal        bool
}

func (m *mockChainSubmitter) SubmitUsageReport(ctx context.Context, report *ChainUsageReport) error {
	m.usageReports = append(m.usageReports, report)
	return nil
}

func (m *mockChainSubmitter) SubmitSettlementRequest(ctx context.Context, orderID string, usageRecordIDs []string, isFinal bool) error {
	m.settlementRequests = append(m.settlementRequests, mockSettlementRequest{
		OrderID:        orderID,
		UsageRecordIDs: usageRecordIDs,
		IsFinal:        isFinal,
	})
	return nil
}

func TestSettlementPipeline_SubmitUsageToChain(t *testing.T) {
	cfg := DefaultSettlementConfig()
	cfg.ProviderAddress = "provider123"

	mockSubmitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(cfg, nil, nil, NewUsageSnapshotStore(), mockSubmitter)

	now := time.Now()
	record := &UsageRecord{
		ID:           "test-record-1",
		DeploymentID: "order-1",
		LeaseID:      "lease-1",
		StartTime:    now.Add(-time.Hour),
		EndTime:      now,
		Metrics: ResourceMetrics{
			CPUMilliSeconds: 3600000,
		},
		PricingInputs: PricingInputs{
			AgreedCPURate: "0.01",
		},
	}

	err := pipeline.SubmitUsageToChain(context.Background(), record)
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	if len(mockSubmitter.usageReports) != 1 {
		t.Errorf("expected 1 usage report, got %d", len(mockSubmitter.usageReports))
	}

	submitted := mockSubmitter.usageReports[0]
	if submitted.OrderID != "order-1" {
		t.Errorf("expected order ID 'order-1', got %s", submitted.OrderID)
	}
	if submitted.UsageUnits == 0 {
		t.Error("expected non-zero usage units")
	}
}

func TestSettlementConfig_Defaults(t *testing.T) {
	cfg := DefaultSettlementConfig()

	if cfg.SettlementInterval != time.Hour {
		t.Errorf("expected 1 hour settlement interval, got %v", cfg.SettlementInterval)
	}

	if cfg.DisputeWindow != 24*time.Hour {
		t.Errorf("expected 24 hour dispute window, got %v", cfg.DisputeWindow)
	}

	if cfg.MaxPendingRecords != 100 {
		t.Errorf("expected 100 max pending records, got %d", cfg.MaxPendingRecords)
	}
}

func TestAnomalyThresholds_Defaults(t *testing.T) {
	thresholds := DefaultAnomalyThresholds()

	if thresholds.MaxCPUVariance != 50.0 {
		t.Errorf("expected 50%% CPU variance threshold, got %.2f", thresholds.MaxCPUVariance)
	}

	if thresholds.MinRecordDuration != time.Minute {
		t.Errorf("expected 1 minute min duration, got %v", thresholds.MinRecordDuration)
	}

	if thresholds.MaxRecordDuration != 25*time.Hour {
		t.Errorf("expected 25 hour max duration, got %v", thresholds.MaxRecordDuration)
	}
}
