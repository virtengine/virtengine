package provider_daemon

import (
	"testing"
	"time"
)

func TestUsageSnapshotStoreTrackAndFind(t *testing.T) {
	store := NewUsageSnapshotStore()

	start := time.Now().Add(-2 * time.Hour).UTC()
	end := start.Add(time.Hour)
	record := &UsageRecord{
		ID:           "r1",
		WorkloadID:   "workload-1",
		DeploymentID: "dep-1",
		LeaseID:      "lease-1",
		AllocationID: "alloc-1",
		StartTime:    start,
		EndTime:      end,
		CreatedAt:    end,
	}

	store.Track(record)

	found, ok := store.FindLatest("alloc-1", nil, nil)
	if !ok || found == nil {
		t.Fatalf("expected record to be found")
	}
	if found.ID != "r1" {
		t.Fatalf("unexpected record ID: %s", found.ID)
	}

	if _, ok := store.FindLatest("dep-1", nil, nil); !ok {
		t.Fatalf("expected record to be found by deployment ID")
	}
	if _, ok := store.FindLatest("lease-1", nil, nil); !ok {
		t.Fatalf("expected record to be found by lease ID")
	}
	if _, ok := store.FindLatest("workload-1", nil, nil); !ok {
		t.Fatalf("expected record to be found by workload ID")
	}

	windowStart := end.Add(10 * time.Minute)
	if _, ok := store.FindLatest("alloc-1", &windowStart, nil); ok {
		t.Fatalf("expected record to be filtered out by period start")
	}
}

func TestUsageSnapshotStoreEqualTimeSelectionIsDeterministic(t *testing.T) {
	store := NewUsageSnapshotStore()
	end := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	created := end.Add(time.Minute)
	store.Track(&UsageRecord{ID: "record-a", AllocationID: "alloc-tie", EndTime: end, CreatedAt: created})
	store.Track(&UsageRecord{ID: "record-b", AllocationID: "alloc-tie", EndTime: end, CreatedAt: created})
	selected, found := store.FindLatest("alloc-tie", nil, nil)
	if !found {
		t.Fatal("expected tied snapshot")
	}
	if selected.ID != "record-b" {
		t.Fatalf("expected deterministic record-b tie winner, got %s", selected.ID)
	}

	reverse := NewUsageSnapshotStore()
	reverse.Track(&UsageRecord{ID: "record-b", AllocationID: "alloc-tie", EndTime: end, CreatedAt: created})
	reverse.Track(&UsageRecord{ID: "record-a", AllocationID: "alloc-tie", EndTime: end, CreatedAt: created})
	selected, found = reverse.FindLatest("alloc-tie", nil, nil)
	if !found || selected.ID != "record-b" {
		t.Fatalf("expected order-independent record-b tie winner")
	}
}
