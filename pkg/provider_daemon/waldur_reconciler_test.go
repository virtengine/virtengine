package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/waldur"
)

type reconciliationCompletionTestStore struct {
	ReconciliationJobStore
	beforeComplete func()
	completeErr    error
	loadErr        error
	pendingErr     error
	projection     *ReconciliationProjection
	duplicate      bool
}

func (s *reconciliationCompletionTestStore) LoadProjection(ctx context.Context) (*ReconciliationProjection, error) {
	if s.loadErr != nil {
		return nil, s.loadErr
	}
	if s.projection != nil {
		return s.projection, nil
	}
	return s.ReconciliationJobStore.LoadProjection(ctx)
}

func (s *reconciliationCompletionTestStore) PendingJobs(ctx context.Context) ([]ReconciliationJob, error) {
	if s.pendingErr != nil {
		return nil, s.pendingErr
	}
	return s.ReconciliationJobStore.PendingJobs(ctx)
}

func (s *reconciliationCompletionTestStore) CompleteAttempt(
	ctx context.Context,
	result DurableReconciliationResult,
	intents []ReconciliationActionIntent,
	cursor ReconciliationCursor,
) error {
	if s.beforeComplete != nil {
		s.beforeComplete()
	}
	if s.completeErr != nil {
		return s.completeErr
	}
	if err := s.ReconciliationJobStore.CompleteAttempt(ctx, result, intents, cursor); err != nil {
		return err
	}
	if s.duplicate {
		return s.ReconciliationJobStore.CompleteAttempt(ctx, result, intents, cursor)
	}
	return nil
}

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
	jobStore := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	reconciler.SetJobStore(jobStore)

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
	projection, err := jobStore.LoadProjection(context.Background())
	require.NoError(t, err)
	require.Len(t, projection.Results, 1)
	for _, durable := range projection.Results {
		require.True(t, validReconciliationResultDigest(durable.Result, durable.ResultDigest))
	}
	require.Len(t, projection.Cursors, 1)
	require.Empty(t, projection.Intents)
}

func TestWaldurReconcilerStartRequiresDurableStore(t *testing.T) {
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	require.ErrorIs(t, reconciler.Start(context.Background()), ErrReconciliationUnavailable)
}

func TestWaldurReconcilerStartLoadFailureWithoutMetrics(t *testing.T) {
	fileStore := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	reconciler.SetJobStore(&reconciliationCompletionTestStore{
		ReconciliationJobStore: fileStore,
		loadErr:                errors.New("injected startup load failure"),
	})
	require.ErrorContains(t, reconciler.Start(context.Background()), "injected startup load failure")
}

func TestWaldurReconcilerStartHydratesDurableResults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(context.Background(), job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(context.Background(), job.ID)
	require.NoError(t, err)
	durable := testDurableReconciliationResult(job, attempt.Number)
	cursor := ReconciliationCursor{StreamID: "waldur/default", LastCompletedJobSequence: 1, JobID: job.ID, ResultDigest: durable.ResultDigest}
	require.NoError(t, store.CompleteAttempt(context.Background(), durable, nil, cursor))
	require.NoError(t, store.Close())

	reopened, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	reconciler.SetJobStore(reopened)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, reconciler.Start(ctx))
	reconciler.Stop()
	result, found := reconciler.GetResult(job.AllocationID)
	require.True(t, found)
	require.Equal(t, durable.Result.State, result.State)
	require.Equal(t, durable.Result.ReasonCode, result.ReasonCode)
}

func TestWaldurReconcilerPersistsUnavailableIntentBeforePublishing(t *testing.T) {
	now := time.Now().UTC()
	usageStore := NewUsageSnapshotStore()
	usageStore.Track(&UsageRecord{
		ID: "record-1", DeploymentID: "alloc-1", StartTime: now.Add(-time.Hour), EndTime: now,
		Metrics: ResourceMetrics{CPUMilliSeconds: 100},
	})
	stateStore := NewWaldurBridgeStateStore(filepath.Join(t.TempDir(), "waldur-state.json"))
	require.NoError(t, stateStore.Save(&WaldurBridgeState{Mappings: map[string]*WaldurAllocationMapping{
		"alloc-1": {AllocationID: "alloc-1", ResourceUUID: "resource-1", UpdatedAt: now},
	}}))
	jobStore := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	cfg := DefaultWaldurReconcilerConfig()
	cfg.ReconciliationInterval = time.Hour
	reconciler := NewWaldurReconciler(cfg, nil, usageStore, nil, stateStore)
	reconciler.SetJobStore(jobStore)

	reconciler.runReconciliation(context.Background())
	result, found := reconciler.GetResult("alloc-1")
	require.True(t, found)
	require.Equal(t, ReconciliationStateUnavailable, result.State)
	require.Equal(t, ReconciliationReasonIndependentEvidenceUnavailable, result.ReasonCode)

	projection, err := jobStore.LoadProjection(context.Background())
	require.NoError(t, err)
	require.Len(t, projection.Results, 1)
	require.Len(t, projection.Intents, 1)
	require.Len(t, projection.Cursors, 1)
	for _, intent := range projection.Intents {
		require.Equal(t, "alert_discrepancy", intent.Kind)
		require.Equal(t, "pending", intent.Status)
	}
}

func TestWaldurReconcilerDurableTransitionsUpdateMetrics(t *testing.T) {
	ctx := context.Background()
	stateStore, usageStore := unavailableReconciliationInputs(t)
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	fileStore := openReconciliationStore(t, path)
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	wrapper := &reconciliationCompletionTestStore{ReconciliationJobStore: fileStore, duplicate: true}
	wrapper.beforeComplete = func() {
		require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_backlog", nil))
		require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_last_completed_timestamp_seconds", nil))
		for _, state := range []ReconciliationState{
			ReconciliationStateMatched, ReconciliationStateMismatched, ReconciliationStateUnavailable,
			ReconciliationStateStale, ReconciliationStateUnresolved,
		} {
			require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_results", map[string]string{"state": string(state)}))
		}
	}
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, usageStore, nil, stateStore)
	reconciler.SetJobStore(wrapper)
	reconciler.SetMetrics(metrics)
	reconciler.runReconciliation(ctx)

	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_backlog", nil))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_results", map[string]string{"state": string(ReconciliationStateUnavailable)}))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_action_intents", map[string]string{"kind": "alert_discrepancy", "severity": "high", "status": "pending"}))
	projection, err := fileStore.LoadProjection(ctx)
	require.NoError(t, err)
	require.Len(t, projection.Results, 1)
	require.Len(t, projection.Intents, 1)
	require.Len(t, projection.Cursors, 1)
	var completedAt time.Time
	for _, result := range projection.Results {
		completedAt = result.CompletedAt
	}
	require.Equal(t, float64(completedAt.Unix()), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_last_completed_timestamp_seconds", nil))

	require.NoError(t, fileStore.Close())
	reopened, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	restartRegistry := prometheus.NewRegistry()
	restartMetrics, err := NewReconciliationMetrics(restartRegistry)
	require.NoError(t, err)
	restarted := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, usageStore, nil, stateStore)
	restarted.SetJobStore(reopened)
	restarted.SetMetrics(restartMetrics)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	require.NoError(t, restarted.Start(cancelled))
	restarted.Stop()
	require.Equal(t, float64(0), reconciliationMetricValue(t, restartRegistry, "virtengine_provider_reconciliation_backlog", nil))
	require.Equal(t, float64(1), reconciliationMetricValue(t, restartRegistry, "virtengine_provider_reconciliation_results", map[string]string{"state": string(ReconciliationStateUnavailable)}))
	require.Equal(t, float64(1), reconciliationMetricValue(t, restartRegistry, "virtengine_provider_reconciliation_action_intents", map[string]string{"kind": "alert_discrepancy", "severity": "high", "status": "pending"}))
	require.Equal(t, float64(completedAt.Unix()), reconciliationMetricValue(t, restartRegistry, "virtengine_provider_reconciliation_last_completed_timestamp_seconds", nil))
}

func TestWaldurReconcilerFailedCompletionDoesNotMutateMetrics(t *testing.T) {
	ctx := context.Background()
	stateStore, usageStore := unavailableReconciliationInputs(t)
	fileStore := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	wrapper := &reconciliationCompletionTestStore{
		ReconciliationJobStore: fileStore,
		completeErr:            errors.New("injected completion failure"),
	}
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, usageStore, nil, stateStore)
	reconciler.SetJobStore(wrapper)
	reconciler.SetMetrics(metrics)
	reconciler.runReconciliation(ctx)

	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_backlog", nil))
	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_last_completed_timestamp_seconds", nil))
	for _, state := range []ReconciliationState{
		ReconciliationStateMatched, ReconciliationStateMismatched, ReconciliationStateUnavailable,
		ReconciliationStateStale, ReconciliationStateUnresolved,
	} {
		require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_results", map[string]string{"state": string(state)}))
	}
	projection, err := fileStore.LoadProjection(ctx)
	require.NoError(t, err)
	require.Empty(t, projection.Results)
	require.Empty(t, projection.Intents)
	require.Empty(t, projection.Cursors)
	_, found := reconciler.GetResult("alloc-1")
	require.False(t, found)
}

func TestWaldurReconcilerRefreshMetricsInvalidatesOnProjectionLoadFailure(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	require.NoError(t, metrics.ObserveProjection(validMetricProjection()))
	require.Equal(t, float64(1), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))

	reconciler := &WaldurReconciler{
		metrics: metrics,
		jobStore: &reconciliationCompletionTestStore{
			loadErr: errors.New("injected projection failure"),
		},
	}
	reconciler.refreshMetrics(context.Background())

	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	for _, name := range []string{
		"virtengine_provider_reconciliation_results",
		"virtengine_provider_reconciliation_backlog",
		"virtengine_provider_reconciliation_last_completed_timestamp_seconds",
		"virtengine_provider_reconciliation_action_intents",
	} {
		require.Equal(t, 0, reconciliationMetricSeriesCount(t, registry, name), name)
	}
}

func TestWaldurReconcilerPendingJobsFailureInvalidatesMetrics(t *testing.T) {
	ctx := context.Background()
	stateStore, usageStore := unavailableReconciliationInputs(t)
	fileStore := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	registry := prometheus.NewRegistry()
	metrics, err := NewReconciliationMetrics(registry)
	require.NoError(t, err)
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, usageStore, nil, stateStore)
	reconciler.SetJobStore(&reconciliationCompletionTestStore{
		ReconciliationJobStore: fileStore,
		pendingErr:             errors.New("injected pending jobs failure"),
	})
	reconciler.SetMetrics(metrics)
	reconciler.runReconciliation(ctx)

	require.Equal(t, float64(0), reconciliationMetricValue(t, registry, "virtengine_provider_reconciliation_projection_available", nil))
	require.Equal(t, 0, reconciliationMetricSeriesCount(t, registry, "virtengine_provider_reconciliation_backlog"))
}

func unavailableReconciliationInputs(t *testing.T) (*WaldurBridgeStateStore, *UsageSnapshotStore) {
	t.Helper()
	now := time.Now().UTC()
	usageStore := NewUsageSnapshotStore()
	usageStore.Track(&UsageRecord{
		ID: "record-1", DeploymentID: "alloc-1", StartTime: now.Add(-time.Hour), EndTime: now,
		Metrics: ResourceMetrics{CPUMilliSeconds: 100},
	})
	stateStore := NewWaldurBridgeStateStore(filepath.Join(t.TempDir(), "waldur-state.json"))
	require.NoError(t, stateStore.Save(&WaldurBridgeState{Mappings: map[string]*WaldurAllocationMapping{
		"alloc-1": {AllocationID: "alloc-1", ResourceUUID: "resource-1", UpdatedAt: now},
	}}))
	return stateStore, usageStore
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

func TestWaldurReconcilerSettlementEligibilityRequiresMatched(t *testing.T) {
	var absent *WaldurReconciler
	require.ErrorIs(t, absent.SettlementEligibility("alloc-1"), ErrSettlementReconciliationHold)

	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, nil, nil, nil)
	require.ErrorIs(t, reconciler.SettlementEligibility("alloc-1"), ErrSettlementReconciliationHold)
	for _, state := range []ReconciliationState{
		ReconciliationStateMismatched,
		ReconciliationStateUnavailable,
		ReconciliationStateStale,
		ReconciliationStateUnresolved,
	} {
		reconciler.storeResult(&ReconciliationResult{AllocationID: "alloc-1", State: state})
		require.ErrorIs(t, reconciler.SettlementEligibility("alloc-1"), ErrSettlementReconciliationHold)
	}
	reconciler.storeResult(&ReconciliationResult{AllocationID: "alloc-1", State: ReconciliationStateMatched})
	require.NoError(t, reconciler.SettlementEligibility("alloc-1"))
}

func TestWaldurReconcilerDurableSettlementEligibilityRequiresCoveringMatchedResult(t *testing.T) {
	ctx := context.Background()
	job := testReconciliationJob()
	record := &UsageRecord{
		ID: "usage-1", AllocationID: job.AllocationID,
		StartTime: job.PeriodStart.Add(time.Minute), EndTime: job.PeriodEnd.Add(-time.Minute),
	}
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)

	store := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	reconciler.SetJobStore(store)
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	result := testDurableReconciliationResult(job, attempt.Number)
	result.Result.State = ReconciliationStateMatched
	result.Result.ReasonCode = ReconciliationReasonExactMatch
	result.Result.Score = 100
	result.Result.ProviderRecordDigest, err = reconciliationProviderRecordDigest(record)
	require.NoError(t, err)
	result.ResultDigest, err = canonicalReconciliationResultDigest(result.Result)
	require.NoError(t, err)
	result.Evidence.ProviderRecordDigest = result.Result.ProviderRecordDigest
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.NoError(t, store.CompleteAttempt(ctx, result, nil, cursor))
	require.NoError(t, reconciler.DurableSettlementEligibility(record))

	otherRecord := *record
	otherRecord.ID = "usage-other"
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(&otherRecord), ErrSettlementReconciliationHold)
	otherRecord = *record
	otherRecord.Metrics.CPUMilliSeconds++
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(&otherRecord), ErrSettlementReconciliationHold)
	otherRecord = *record
	otherRecord.AllocationID = "other-allocation"
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(&otherRecord), ErrSettlementReconciliationHold)
	otherRecord = *record
	otherRecord.StartTime = otherRecord.StartTime.Add(time.Second)
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(&otherRecord), ErrSettlementReconciliationHold)

	newerJob := newReconciliationJob(job.AllocationID, job.ResourceUUID, job.PeriodStart, job.PeriodEnd.Add(time.Minute))
	_, _, err = store.PutJobIfAbsent(ctx, newerJob)
	require.NoError(t, err)
	newerAttempt, err := store.BeginAttempt(ctx, newerJob.ID)
	require.NoError(t, err)
	newerResult := testDurableReconciliationResult(newerJob, newerAttempt.Number)
	newerCursor := ReconciliationCursor{StreamID: "waldur/default", JobID: newerJob.ID, ResultDigest: newerResult.ResultDigest}
	require.NoError(t, store.CompleteAttempt(ctx, newerResult, nil, newerCursor))
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)

	record.EndTime = job.PeriodEnd.Add(time.Second)
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)
}

func TestWaldurReconcilerDurableSettlementEligibilityRejectsAuthorityTie(t *testing.T) {
	job := testReconciliationJob()
	otherJob := job
	otherJob.ID = "job-2"
	matched := testDurableReconciliationResult(job, 1)
	matched.Result.State = ReconciliationStateMatched
	matched.Result.ReasonCode = ReconciliationReasonExactMatch
	matched.Result.Score = 100
	mismatched := testDurableReconciliationResult(otherJob, 1)
	projection := &ReconciliationProjection{
		Jobs:    map[string]ReconciliationJob{job.ID: job, otherJob.ID: otherJob},
		Results: map[string]DurableReconciliationResult{job.ID: matched, otherJob.ID: mismatched},
	}
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	reconciler.SetJobStore(&reconciliationCompletionTestStore{projection: projection})
	record := &UsageRecord{AllocationID: job.AllocationID, StartTime: job.PeriodStart, EndTime: job.PeriodEnd}
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)

	record.EndTime = record.StartTime
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)
}

func TestWaldurReconcilerDurableSettlementEligibilityRejectsRecordDigestTie(t *testing.T) {
	job := testReconciliationJob()
	otherJob := job
	otherJob.ID = "job-2"
	record := &UsageRecord{ID: "usage-1", AllocationID: job.AllocationID, StartTime: job.PeriodStart, EndTime: job.PeriodEnd}
	first := testDurableReconciliationResult(job, 1)
	first.Result.State, first.Result.ReasonCode, first.Result.Score = ReconciliationStateMatched, ReconciliationReasonExactMatch, 100
	first.Result.ProviderRecordDigest, _ = reconciliationProviderRecordDigest(record)
	first.Evidence.ProviderRecordDigest = first.Result.ProviderRecordDigest
	second := first
	second.JobID = otherJob.ID
	second.Result.ProviderRecordDigest = strings.Repeat("a", 64)
	second.Evidence.ProviderRecordDigest = second.Result.ProviderRecordDigest
	projection := &ReconciliationProjection{
		Jobs:    map[string]ReconciliationJob{job.ID: job, otherJob.ID: otherJob},
		Results: map[string]DurableReconciliationResult{job.ID: first, otherJob.ID: second},
	}
	reconciler := NewWaldurReconciler(DefaultWaldurReconcilerConfig(), nil, NewUsageSnapshotStore(), nil, nil)
	reconciler.SetJobStore(&reconciliationCompletionTestStore{projection: projection})
	require.ErrorIs(t, reconciler.DurableSettlementEligibility(record), ErrSettlementReconciliationHold)
}

func TestReconciliationResultJSONUsesExplicitContract(t *testing.T) {
	bz, err := json.Marshal(&ReconciliationResult{State: ReconciliationStateUnavailable, ReasonCode: ReconciliationReasonIndependentEvidenceUnavailable})
	require.NoError(t, err)
	require.Contains(t, string(bz), `"state":"unavailable"`)
	require.Contains(t, string(bz), `"reason_code":"independent_evidence_unavailable"`)
	require.NotContains(t, string(bz), "in_sync")
}
