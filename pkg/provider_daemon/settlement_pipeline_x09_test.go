package provider_daemon

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type blockingSettlementSubmitterX09 struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
}

func (s *blockingSettlementSubmitterX09) SubmitUsageReport(context.Context, *ChainUsageReport) error {
	return nil
}

func (s *blockingSettlementSubmitterX09) SubmitSettlementRequest(context.Context, string, []string, bool) error {
	s.calls.Add(1)
	s.once.Do(func() { close(s.entered) })
	<-s.release
	return nil
}

func TestSettlementPipelineSubmitUsageToChainSnapshotsBeforeEligibility(t *testing.T) {
	submitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	entered := make(chan struct{})
	release := make(chan struct{})
	observedAllocation := make(chan string, 1)
	var enteredOnce sync.Once
	pipeline.SetSettlementEligibility(func(record *UsageRecord) error {
		enteredOnce.Do(func() {
			observedAllocation <- record.AllocationID
			close(entered)
		})
		<-release
		return nil
	})
	originalStart := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	originalEnd := originalStart.Add(time.Hour)
	originalMetrics := ResourceMetrics{CPUMilliSeconds: 3_600_000}
	record := &UsageRecord{
		ID: "record-original", AllocationID: "allocation-original", DeploymentID: "order-1", LeaseID: "lease-1",
		StartTime: originalStart, EndTime: originalEnd, Metrics: originalMetrics,
		PricingInputs: PricingInputs{AgreedCPURate: "0.01"},
	}
	errors := make(chan error, 1)
	go func() { errors <- pipeline.SubmitUsageToChain(context.Background(), record) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("eligibility check did not start")
	}
	require.Equal(t, "allocation-original", <-observedAllocation)
	record.AllocationID = "allocation-mutated"
	record.StartTime = originalStart.Add(-time.Hour)
	record.EndTime = originalEnd.Add(time.Hour)
	record.Metrics.CPUMilliSeconds = 7_200_000
	close(release)
	select {
	case err := <-errors:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("submission did not complete")
	}
	require.Len(t, submitter.usageReports, 1)
	submitted := submitter.usageReports[0]
	require.Equal(t, "allocation-original", submitted.AllocationID)
	require.Equal(t, originalStart, submitted.PeriodStart)
	require.Equal(t, originalEnd, submitted.PeriodEnd)
	require.Equal(t, originalMetrics, submitted.RawMetrics)
	require.Equal(t, uint64(1), submitted.UsageUnits)
}

func TestSettlementPipelineAddPendingUsageOwnsSnapshot(t *testing.T) {
	store := NewUsageSnapshotStore()
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, store, nil)
	originalEnd := time.Date(2026, 8, 2, 11, 0, 0, 0, time.UTC)
	record := &UsageRecord{ID: "record-original", AllocationID: "allocation-original", EndTime: originalEnd, Metrics: ResourceMetrics{CPUMilliSeconds: 3_600_000}}
	pipeline.AddPendingUsage(record)
	record.AllocationID = "allocation-mutated"
	record.EndTime = originalEnd.Add(time.Hour)
	record.Metrics.CPUMilliSeconds = 7_200_000
	stored, found := store.FindLatest("record-original", nil, nil)
	require.True(t, found)
	require.Equal(t, "allocation-original", stored.AllocationID)
	require.Equal(t, originalEnd, stored.EndTime)
	require.Equal(t, int64(3_600_000), stored.Metrics.CPUMilliSeconds)
	stored.AllocationID = "lookup-mutated"
	again, found := store.FindLatest("record-original", nil, nil)
	require.True(t, found)
	require.Equal(t, "allocation-original", again.AllocationID)
}

func TestSettlementPipelineDisputeEvidenceIsImmutableAndCorrectionRetracked(t *testing.T) {
	store := NewUsageSnapshotStore()
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, store, nil)
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	record := &UsageRecord{ID: "record-disputed", AllocationID: "allocation-1", EndTime: now, Metrics: ResourceMetrics{CPUMilliSeconds: 3_600_000}}
	pipeline.AddPendingUsage(record)
	expected := &ResourceMetrics{CPUMilliSeconds: 1_800_000}
	dispute, err := pipeline.CreateDispute(record.ID, "order-1", "customer", "usage mismatch", "", expected)
	require.NoError(t, err)
	expected.CPUMilliSeconds = 9_999_999
	dispute.ExpectedUsage.CPUMilliSeconds = 8_888_888
	dispute.ReportedUsage.CPUMilliSeconds = 7_777_777
	require.NoError(t, pipeline.ResolveDispute(dispute.DisputeID, "apply expected usage", true))
	stored, found := store.FindLatest("allocation-1", nil, nil)
	require.True(t, found)
	require.Equal(t, int64(1_800_000), stored.Metrics.CPUMilliSeconds)
	pipeline.mu.RLock()
	internal := pipeline.disputes[dispute.DisputeID]
	require.Equal(t, int64(1_800_000), internal.ExpectedUsage.CPUMilliSeconds)
	require.Equal(t, int64(3_600_000), internal.ReportedUsage.CPUMilliSeconds)
	pipeline.mu.RUnlock()
}

func TestSettlementPipelineProcessSettlementsDoesNotHoldLockDuringEligibility(t *testing.T) {
	submitter := &retryChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	entered := make(chan struct{})
	release := make(chan struct{})
	var enteredOnce sync.Once
	pipeline.SetSettlementEligibility(func(*UsageRecord) error {
		enteredOnce.Do(func() { close(entered) })
		<-release
		return nil
	})
	now := time.Now()
	pipeline.AddPendingUsage(&UsageRecord{ID: "rec-blocked", DeploymentID: "order-blocked", AllocationID: "allocation-1", StartTime: now.Add(-time.Hour), EndTime: now})
	processed := make(chan struct{})
	go func() { pipeline.processSettlements(context.Background()); close(processed) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("eligibility check did not start")
	}
	added := make(chan struct{})
	go func() { pipeline.AddPendingUsage(&UsageRecord{ID: "rec-new", DeploymentID: "order-new"}); close(added) }()
	select {
	case <-added:
	case <-time.After(time.Second):
		t.Fatal("pending ingestion blocked behind eligibility")
	}
	close(release)
	select {
	case <-processed:
	case <-time.After(time.Second):
		t.Fatal("settlement processing did not complete")
	}
	require.Equal(t, 1, pipeline.GetPendingCount())
}

func TestSettlementPipelineRejectsDisputeAfterSettlementCommitStarts(t *testing.T) {
	submitter := &blockingSettlementSubmitterX09{entered: make(chan struct{}), release: make(chan struct{})}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	pipeline.SetSettlementEligibility(func(*UsageRecord) error { return nil })
	now := time.Now()
	pipeline.AddPendingUsage(&UsageRecord{ID: "rec-settling", DeploymentID: "order-settling", AllocationID: "allocation-1", StartTime: now.Add(-time.Hour), EndTime: now})
	done := make(chan struct{})
	go func() { pipeline.processSettlements(context.Background()); close(done) }()
	select {
	case <-submitter.entered:
	case <-time.After(time.Second):
		t.Fatal("settlement submission did not start")
	}
	_, err := pipeline.CreateDispute("rec-settling", "order-settling", "customer", "late dispute", "", nil)
	require.ErrorContains(t, err, "settlement is in progress")
	close(submitter.release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("settlement submission did not complete")
	}
	require.Equal(t, 0, pipeline.GetPendingCount())
}

func TestSettlementPipelineAddPendingUsageRejectsPricingConflict(t *testing.T) {
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), nil)
	first := &UsageRecord{ID: "dup-price", DeploymentID: "order-1", Metrics: ResourceMetrics{CPUMilliSeconds: 3_600_000}, PricingInputs: PricingInputs{AgreedCPURate: "0.01"}}
	conflict := *first
	conflict.PricingInputs.AgreedCPURate = "100.00"
	pipeline.AddPendingUsage(first)
	pipeline.AddPendingUsage(&conflict)
	require.Len(t, pipeline.GetUnacknowledgedAnomalies(), 1)
}

func TestSettlementPipelineRejectsNegativeCorrection(t *testing.T) {
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), nil)
	record := &UsageRecord{ID: "record-negative", Metrics: ResourceMetrics{CPUMilliSeconds: 100}}
	pipeline.AddPendingUsage(record)
	dispute, err := pipeline.CreateDispute(record.ID, "order-1", "customer", "invalid correction", "", &ResourceMetrics{CPUMilliSeconds: -1})
	require.NoError(t, err)
	require.ErrorContains(t, pipeline.ResolveDispute(dispute.DisputeID, "reject negative", true), "corrected usage metrics are invalid")
	pipeline.mu.RLock()
	require.Equal(t, DisputeStatusPending, pipeline.disputes[dispute.DisputeID].Status)
	require.Nil(t, pipeline.disputes[dispute.DisputeID].ResolvedAt)
	require.Equal(t, int64(100), pipeline.pending[record.ID].Metrics.CPUMilliSeconds)
	pipeline.mu.RUnlock()
}

func TestSettlementPipelineSortsSettlementUsageIDs(t *testing.T) {
	submitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	pipeline.SetSettlementEligibility(func(*UsageRecord) error { return nil })
	pipeline.AddPendingUsage(&UsageRecord{ID: "record-b", DeploymentID: "order-1"})
	pipeline.AddPendingUsage(&UsageRecord{ID: "record-a", DeploymentID: "order-1"})
	pipeline.processSettlements(context.Background())
	require.Len(t, submitter.settlementRequests, 1)
	require.Equal(t, []string{"record-a", "record-b"}, submitter.settlementRequests[0].UsageRecordIDs)
}

func TestSettlementPipelineCorrectionDuringEligibilityRequiresRecheck(t *testing.T) {
	submitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	entered := make(chan struct{})
	release := make(chan struct{})
	pipeline.SetSettlementEligibility(func(*UsageRecord) error {
		close(entered)
		<-release
		return nil
	})
	now := time.Now().UTC()
	record := &UsageRecord{
		ID: "record-corrected-during-check", DeploymentID: "order-1", AllocationID: "allocation-1",
		StartTime: now.Add(-time.Hour), EndTime: now, Metrics: ResourceMetrics{CPUMilliSeconds: 3_600_000},
	}
	pipeline.AddPendingUsage(record)
	done := make(chan struct{})
	go func() { pipeline.processSettlements(context.Background()); close(done) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("eligibility check did not start")
	}
	dispute, err := pipeline.CreateDispute(record.ID, record.DeploymentID, "customer", "correction", "", &ResourceMetrics{CPUMilliSeconds: 1_800_000})
	require.NoError(t, err)
	require.NoError(t, pipeline.ResolveDispute(dispute.DisputeID, "accepted", true))
	close(release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("settlement processing did not complete")
	}
	require.Empty(t, submitter.settlementRequests)
	require.Equal(t, 1, pipeline.GetPendingCount())

	pipeline.SetSettlementEligibility(func(candidate *UsageRecord) error {
		require.Equal(t, int64(1_800_000), candidate.Metrics.CPUMilliSeconds)
		return nil
	})
	pipeline.processSettlements(context.Background())
	require.Len(t, submitter.settlementRequests, 1)
	require.Equal(t, 0, pipeline.GetPendingCount())
}

func TestSettlementPipelineSortsOrdersDeterministically(t *testing.T) {
	submitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	pipeline.SetSettlementEligibility(func(*UsageRecord) error { return nil })
	pipeline.AddPendingUsage(&UsageRecord{ID: "record-z", DeploymentID: "order-z"})
	pipeline.AddPendingUsage(&UsageRecord{ID: "record-a", DeploymentID: "order-a"})
	pipeline.AddPendingUsage(&UsageRecord{ID: "record-m", DeploymentID: "order-m"})
	pipeline.processSettlements(context.Background())
	require.Len(t, submitter.settlementRequests, 3)
	require.Equal(t, "order-a", submitter.settlementRequests[0].OrderID)
	require.Equal(t, "order-m", submitter.settlementRequests[1].OrderID)
	require.Equal(t, "order-z", submitter.settlementRequests[2].OrderID)
}

func TestSettlementPipelineGeneratedIDsDoNotCollideAtSameTimestamp(t *testing.T) {
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), nil)
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	require.NotEqual(t, pipeline.generateID("dispute", now), pipeline.generateID("dispute", now))
}

func TestSettlementPipelineRechecksEligibilityAfterReservation(t *testing.T) {
	submitter := &mockChainSubmitter{}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	var checks atomic.Int32
	pipeline.SetSettlementEligibility(func(*UsageRecord) error {
		if checks.Add(1) > 1 {
			return ErrSettlementReconciliationHold
		}
		return nil
	})
	pipeline.AddPendingUsage(&UsageRecord{ID: "reauthorize", DeploymentID: "order-1"})
	pipeline.processSettlements(context.Background())
	require.Empty(t, submitter.settlementRequests)
	require.Equal(t, 1, pipeline.GetPendingCount())
}

func TestSettlementPipelineCorrectionRecordsResolutionReason(t *testing.T) {
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), nil)
	record := &UsageRecord{ID: "reason-record", Metrics: ResourceMetrics{CPUMilliSeconds: 100}}
	pipeline.AddPendingUsage(record)
	dispute, err := pipeline.CreateDispute(record.ID, "order-1", "customer", "wrong usage", "", &ResourceMetrics{CPUMilliSeconds: 50})
	require.NoError(t, err)
	require.NoError(t, pipeline.ResolveDispute(dispute.DisputeID, "independent evidence accepted", true))
	pipeline.mu.RLock()
	defer pipeline.mu.RUnlock()
	require.Len(t, pipeline.corrections, 1)
	for _, correction := range pipeline.corrections {
		require.Equal(t, "independent evidence accepted", correction.Reason)
	}
}

func TestSettlementPipelineConcurrentPassDoesNotDoubleSubmit(t *testing.T) {
	submitter := &blockingSettlementSubmitterX09{entered: make(chan struct{}), release: make(chan struct{})}
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter)
	pipeline.SetSettlementEligibility(func(*UsageRecord) error { return nil })
	pipeline.AddPendingUsage(&UsageRecord{ID: "single-submit", DeploymentID: "order-1"})
	firstDone := make(chan struct{})
	go func() { pipeline.processSettlements(context.Background()); close(firstDone) }()
	select {
	case <-submitter.entered:
	case <-time.After(time.Second):
		t.Fatal("first settlement did not start")
	}
	secondDone := make(chan struct{})
	go func() { pipeline.processSettlements(context.Background()); close(secondDone) }()
	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Fatal("overlapping settlement pass did not return")
	}
	require.Equal(t, int32(1), submitter.calls.Load())
	close(submitter.release)
	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("first settlement did not finish")
	}
}

func TestSettlementPipelineNilUsageStoreRemainsFailClosed(t *testing.T) {
	pipeline := NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, nil, nil)
	require.NotPanics(t, func() { pipeline.AddPendingUsage(&UsageRecord{ID: "nil-store"}) })
	dispute, err := pipeline.CreateDispute("nil-store", "order-1", "customer", "correction", "", &ResourceMetrics{CPUMilliSeconds: 1})
	require.NoError(t, err)
	require.NoError(t, pipeline.ResolveDispute(dispute.DisputeID, "accepted", true))
	pipeline.mu.RLock()
	require.Equal(t, int64(1), pipeline.pending["nil-store"].Metrics.CPUMilliSeconds)
	pipeline.mu.RUnlock()
}
