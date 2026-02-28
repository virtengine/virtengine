package provider_daemon

import (
	"context"
	"encoding/json"
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
	require.True(t, result.InSync)
	require.Equal(t, 100, result.Score)
	require.Empty(t, result.Discrepancies)
}
