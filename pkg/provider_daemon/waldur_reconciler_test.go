package provider_daemon

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/waldur"
)

func TestWaldurReconcilerFetchWaldurUsageUsesUsageAPI(t *testing.T) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/users/me/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"uuid": "user-1"})
		case strings.Contains(r.URL.Path, "/marketplace-resources/resource-1/usages/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]waldur.UsageRecord{
				{UUID: "cpu-1", ResourceUUID: "resource-1", ComponentType: "cpu_hours", Usage: 2, Date: now, Created: now},
				{UUID: "ram-1", ResourceUUID: "resource-1", ComponentType: "ram_gb_hours", Usage: 4, Date: now, Created: now},
				{UUID: "storage-1", ResourceUUID: "resource-1", ComponentType: "storage_gb_hours", Usage: 1.5, Date: now, Created: now},
				{UUID: "gpu-1", ResourceUUID: "resource-1", ComponentType: "gpu_hours", Usage: 3, Date: now, Created: now},
				{UUID: "network-1", ResourceUUID: "resource-1", ComponentType: "network_gb", Usage: 8, Date: now, Created: now},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := waldur.NewClient(waldur.Config{BaseURL: server.URL, Token: "test-token", MaxRetries: 0})
	require.NoError(t, err)

	usageStore := NewUsageSnapshotStore()
	record := &UsageRecord{
		ID:           "record-1",
		DeploymentID: "alloc-1",
		StartTime:    now.Add(-30 * time.Minute),
		EndTime:      now,
		Metrics: ResourceMetrics{
			CPUMilliSeconds:    int64(2 * 3600 * 1000),
			MemoryByteSeconds:  int64(4 * float64(1024*1024*1024) * 3600),
			StorageByteSeconds: int64(1.5 * float64(1024*1024*1024) * 3600),
			GPUSeconds:         int64(3 * 3600),
			NetworkBytesIn:     4 * 1024 * 1024 * 1024,
			NetworkBytesOut:    4 * 1024 * 1024 * 1024,
		},
	}
	usageStore.Track(record)

	stateStore := NewWaldurBridgeStateStore(filepath.Join(t.TempDir(), "waldur-state.json"))
	require.NoError(t, stateStore.Save(&WaldurBridgeState{
		Mappings: map[string]*WaldurAllocationMapping{
			"alloc-1": {
				AllocationID: "alloc-1",
				ResourceUUID: "resource-1",
				UpdatedAt:    now,
			},
		},
	}))

	cfg := DefaultWaldurReconcilerConfig()
	cfg.ReconciliationInterval = time.Hour
	cfg.DiscrepancyThreshold = 1

	reconciler := NewWaldurReconciler(cfg, waldur.NewMarketplaceClient(client), usageStore, nil, stateStore)

	stats, err := reconciler.fetchWaldurUsage(context.Background(), "resource-1", now.Add(-time.Hour), now)
	require.NoError(t, err)
	require.Equal(t, 2.0, stats.CPUHours)
	require.Equal(t, 4.0, stats.RAMGBHours)
	require.Equal(t, 1.5, stats.StorageGBHours)
	require.Equal(t, 3.0, stats.GPUHours)
	require.Equal(t, 8.0, stats.NetworkGB)
	require.Len(t, stats.Components, 5)

	reconciler.runReconciliation(context.Background())
	result, ok := reconciler.GetResult("alloc-1")
	require.True(t, ok)
	require.Equal(t, ReconciliationStateMatched, result.State)
	require.Equal(t, ReconciliationReasonExactMatch, result.ReasonCode)
	require.Equal(t, 100, result.Score)
	require.Empty(t, result.Discrepancies)
}

func TestWaldurReconciliationStateClassification(t *testing.T) {
	now := time.Now().UTC()
	currentRecord := func(end time.Time) *UsageRecord {
		return &UsageRecord{
			ID: "record-1", DeploymentID: "alloc-1", StartTime: end.Add(-time.Hour), EndTime: end,
			Metrics: ResourceMetrics{CPUMilliSeconds: 100},
		}
	}

	t.Run("provider evidence unavailable", func(t *testing.T) {
		reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
		result, err := reconciler.ReconcileAllocation(context.Background(), "alloc-1", "resource-1")
		require.NoError(t, err)
		require.Equal(t, ReconciliationStateUnavailable, result.State)
		require.Equal(t, ReconciliationReasonProviderEvidenceUnavailable, result.ReasonCode)
		require.Zero(t, result.Score)
	})

	t.Run("independent evidence unavailable", func(t *testing.T) {
		store := NewUsageSnapshotStore()
		store.Track(currentRecord(now))
		reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, store, nil, nil)
		result, err := reconciler.ReconcileAllocation(context.Background(), "alloc-1", "resource-1")
		require.NoError(t, err)
		require.Equal(t, ReconciliationStateUnavailable, result.State)
		require.Equal(t, ReconciliationReasonIndependentEvidenceUnavailable, result.ReasonCode)
	})

	t.Run("provider evidence stale", func(t *testing.T) {
		store := NewUsageSnapshotStore()
		store.Track(currentRecord(now.Add(-2 * time.Hour)))
		cfg := DefaultWaldurReconcilerConfig()
		cfg.ReconciliationInterval = 24 * time.Hour
		cfg.MaxAgeForReconciliation = time.Hour
		reconciler := NewWaldurReconciler(cfg, nil, store, nil, nil)
		result, err := reconciler.ReconcileAllocation(context.Background(), "alloc-1", "resource-1")
		require.NoError(t, err)
		require.Equal(t, ReconciliationStateStale, result.State)
		require.Equal(t, ReconciliationReasonProviderEvidenceStale, result.ReasonCode)
	})
}

func TestWaldurReconciliationToleranceAndCompleteness(t *testing.T) {
	cfg := DefaultWaldurReconcilerConfig()
	cfg.DiscrepancyThreshold = 10
	reconciler := NewWaldurReconciler(cfg, nil, nil, nil, nil)
	require.Nil(t, reconciler.calculateDiscrepancy("cpu", 109, 100))
	require.NotNil(t, reconciler.calculateDiscrepancy("cpu", 110, 100))

	provider := ResourceMetrics{CPUMilliSeconds: 100, GPUSeconds: 50}
	require.False(t, hasCompleteIndependentEvidence(provider, []WaldurUsageComponent{{Type: "cpu_hours"}}))
	require.True(t, hasCompleteIndependentEvidence(provider, []WaldurUsageComponent{{Type: "cpu_hours"}, {Type: "gpu_hours"}}))
}

func TestValidateWaldurUsageRecordRejectsMalformedEvidence(t *testing.T) {
	now := time.Now().UTC()
	valid := waldur.UsageRecord{ResourceUUID: "resource-1", ComponentType: "cpu_hours", Usage: 1, Date: now}
	require.NoError(t, validateWaldurUsageRecord(valid, "resource-1"))

	tests := []waldur.UsageRecord{
		{ResourceUUID: "other", ComponentType: "cpu_hours", Usage: 1, Date: now},
		{ResourceUUID: "resource-1", ComponentType: "unknown", Usage: 1, Date: now},
		{ResourceUUID: "resource-1", ComponentType: "cpu_hours", Usage: -1, Date: now},
		{ResourceUUID: "resource-1", ComponentType: "cpu_hours", Usage: math.NaN(), Date: now},
		{ResourceUUID: "resource-1", ComponentType: "cpu_hours", Usage: 1},
	}
	for _, record := range tests {
		var validationErr *reconciliationEvidenceError
		require.ErrorAs(t, validateWaldurUsageRecord(record, "resource-1"), &validationErr)
		require.Equal(t, ReconciliationReasonMalformedEvidence, validationErr.reason)
	}
}

func TestReconciliationSyncStatusUsesExplicitStates(t *testing.T) {
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, nil, nil, nil)
	states := []ReconciliationState{
		ReconciliationStateMatched, ReconciliationStateMismatched, ReconciliationStateUnavailable,
		ReconciliationStateStale, ReconciliationStateUnresolved,
	}
	for index, state := range states {
		reconciler.storeResult(&ReconciliationResult{AllocationID: string(rune('a' + index)), State: state, Score: 20})
	}
	status := reconciler.GetSyncStatus()
	require.Equal(t, 5, status.TotalAllocations)
	require.Equal(t, 1, status.MatchedCount)
	require.Equal(t, 1, status.MismatchedCount)
	require.Equal(t, 1, status.UnavailableCount)
	require.Equal(t, 1, status.StaleCount)
	require.Equal(t, 1, status.UnresolvedCount)
}

func TestReconciliationResultJSONUsesExplicitContract(t *testing.T) {
	bz, err := json.Marshal(&ReconciliationResult{State: ReconciliationStateUnavailable, ReasonCode: ReconciliationReasonIndependentEvidenceUnavailable})
	require.NoError(t, err)
	require.Contains(t, string(bz), `"state":"unavailable"`)
	require.Contains(t, string(bz), `"reason_code":"independent_evidence_unavailable"`)
	require.NotContains(t, string(bz), "in_sync")
}
