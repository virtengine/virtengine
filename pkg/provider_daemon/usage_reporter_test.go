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
		WorkloadID:   "alloc-1",
		DeploymentID: "dep-1",
		LeaseID:      "lease-1",
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

	windowStart := end.Add(10 * time.Minute)
	if _, ok := store.FindLatest("alloc-1", &windowStart, nil); ok {
		t.Fatalf("expected record to be filtered out by period start")
	}
}
