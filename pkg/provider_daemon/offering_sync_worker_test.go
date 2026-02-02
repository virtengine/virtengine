// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-14D: Tests for offering sync worker and chain-to-Waldur synchronization.
package provider_daemon

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// sampleCompute returns a sample compute offering for testing.
func sampleCompute() *marketplace.Offering {
	now := time.Now().UTC()
	return &marketplace.Offering{
		ID: marketplace.OfferingID{
			ProviderAddress: "virtengine1provider123",
			Sequence:        1,
		},
		Name:        "Basic Compute Instance",
		Description: "A basic compute instance with 2 vCPUs and 4GB RAM",
		Category:    marketplace.OfferingCategoryCompute,
		State:       marketplace.OfferingStateActive,
		Version:     "1.0.0",
		Pricing: marketplace.PricingInfo{
			Model:      marketplace.PricingModelHourly,
			BasePrice:  50000,
			Currency:   "uvirt",
			UsageRates: map[string]uint64{},
		},
		IdentityRequirement: marketplace.IdentityRequirement{
			MinScore:   10,
			RequireMFA: false,
		},
		Tags:                []string{"compute", "basic"},
		Regions:             []string{"us-east-1", "eu-west-1"},
		MaxConcurrentOrders: 10,
		RequireMFAForOrders: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// sampleHPC returns a sample HPC offering for testing.
func sampleHPC() *marketplace.Offering {
	now := time.Now().UTC()
	return &marketplace.Offering{
		ID: marketplace.OfferingID{
			ProviderAddress: "virtengine1provider123",
			Sequence:        2,
		},
		Name:        "HPC Cluster Access",
		Description: "High-performance computing cluster with SLURM scheduling",
		Category:    marketplace.OfferingCategoryHPC,
		State:       marketplace.OfferingStateActive,
		Version:     "1.0.0",
		Pricing: marketplace.PricingInfo{
			Model:     marketplace.PricingModelUsageBased,
			BasePrice: 0,
			Currency:  "uvirt",
			UsageRates: map[string]uint64{
				"cpu_hours": 10000,
				"gpu_hours": 1500000,
			},
		},
		IdentityRequirement: marketplace.IdentityRequirement{
			MinScore:   50,
			RequireMFA: true,
		},
		Tags:                []string{"hpc", "slurm", "research"},
		Regions:             []string{"us-west-2"},
		MaxConcurrentOrders: 5,
		RequireMFAForOrders: true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func TestOffering_ToWaldurCreate(t *testing.T) {
	cfg := marketplace.DefaultOfferingSyncConfig()
	cfg.WaldurCustomerUUID = "waldur-customer-uuid-123"
	cfg.WaldurCategoryMap = map[string]string{
		"compute": "waldur-cat-compute",
		"hpc":     "waldur-cat-hpc",
	}

	tests := []struct {
		name         string
		offering     *marketplace.Offering
		wantName     string
		wantType     string
		wantState    string
		wantShared   bool
		wantBillable bool
	}{
		{
			name:         "compute offering",
			offering:     sampleCompute(),
			wantName:     "Basic Compute Instance",
			wantType:     "VirtEngine.Compute",
			wantState:    "Active",
			wantShared:   true,
			wantBillable: true,
		},
		{
			name:         "hpc offering",
			offering:     sampleHPC(),
			wantName:     "HPC Cluster Access",
			wantType:     "VirtEngine.HPC",
			wantState:    "Active",
			wantShared:   true,
			wantBillable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waldurCreate := tt.offering.ToWaldurCreate(cfg)

			if waldurCreate.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", waldurCreate.Name, tt.wantName)
			}
			if waldurCreate.Type != tt.wantType {
				t.Errorf("Type = %s, want %s", waldurCreate.Type, tt.wantType)
			}
			if waldurCreate.State != tt.wantState {
				t.Errorf("State = %s, want %s", waldurCreate.State, tt.wantState)
			}
			if waldurCreate.Shared != tt.wantShared {
				t.Errorf("Shared = %t, want %t", waldurCreate.Shared, tt.wantShared)
			}
			if waldurCreate.Billable != tt.wantBillable {
				t.Errorf("Billable = %t, want %t", waldurCreate.Billable, tt.wantBillable)
			}
			if waldurCreate.CustomerUUID != cfg.WaldurCustomerUUID {
				t.Errorf("CustomerUUID = %s, want %s", waldurCreate.CustomerUUID, cfg.WaldurCustomerUUID)
			}
			if waldurCreate.BackendID != tt.offering.ID.String() {
				t.Errorf("BackendID = %s, want %s", waldurCreate.BackendID, tt.offering.ID.String())
			}
		})
	}
}

func TestOffering_ToWaldurUpdate(t *testing.T) {
	cfg := marketplace.DefaultOfferingSyncConfig()
	offering := sampleCompute()
	offering.Name = "Updated Compute Instance"
	offering.Description = "Updated description"

	waldurUpdate := offering.ToWaldurUpdate(cfg)

	if waldurUpdate.Name != "Updated Compute Instance" {
		t.Errorf("Name = %s, want 'Updated Compute Instance'", waldurUpdate.Name)
	}
	if waldurUpdate.Description != "Updated description" {
		t.Errorf("Description = %s, want 'Updated description'", waldurUpdate.Description)
	}
	if waldurUpdate.State != "Active" {
		t.Errorf("State = %s, want 'Active'", waldurUpdate.State)
	}
}

func TestOffering_SyncChecksum(t *testing.T) {
	offering := sampleCompute()

	checksum1 := offering.SyncChecksum()
	checksum2 := offering.SyncChecksum()

	if checksum1 != checksum2 {
		t.Error("checksums should be deterministic")
	}

	// Modify offering and check checksum changes
	offering.Name = "Modified Name"
	checksum3 := offering.SyncChecksum()

	if checksum1 == checksum3 {
		t.Error("checksum should differ for modified offering")
	}
}

func TestOfferingSyncState(t *testing.T) {
	state := NewOfferingSyncState("virtengine1provider123")

	if state.ProviderAddress != "virtengine1provider123" {
		t.Errorf("ProviderAddress = %s, want virtengine1provider123", state.ProviderAddress)
	}
	if len(state.Records) != 0 {
		t.Errorf("Records should be empty, got %d", len(state.Records))
	}

	t.Run("mark synced", func(t *testing.T) {
		state.MarkSynced("offering/1", "waldur-uuid-1", "checksum123", 1)
		record := state.GetRecord("offering/1")
		if record == nil {
			t.Fatal("record should exist")
		}
		if record.State != SyncStateSynced {
			t.Errorf("state = %v, want synced", record.State)
		}
		if record.WaldurUUID != "waldur-uuid-1" {
			t.Errorf("WaldurUUID = %s, want waldur-uuid-1", record.WaldurUUID)
		}
		if record.Checksum != "checksum123" {
			t.Errorf("Checksum = %s, want checksum123", record.Checksum)
		}
		if record.SyncedVersion != 1 {
			t.Errorf("SyncedVersion = %d, want 1", record.SyncedVersion)
		}
	})

	t.Run("mark failed with retry", func(t *testing.T) {
		state.GetOrCreateRecord("offering/2")
		deadLettered := state.MarkFailed("offering/2", "test error", 3, time.Second, time.Minute)
		if deadLettered {
			t.Error("should not be dead-lettered on first failure")
		}
		record := state.GetRecord("offering/2")
		if record.State != SyncStateRetrying {
			t.Errorf("state = %v, want retrying", record.State)
		}
		if record.RetryCount != 1 {
			t.Errorf("RetryCount = %d, want 1", record.RetryCount)
		}
		if record.NextRetryAt == nil {
			t.Error("NextRetryAt should be set")
		}
	})

	t.Run("dead letter after max retries", func(t *testing.T) {
		state.GetOrCreateRecord("offering/3")
		for i := 0; i <= 3; i++ {
			state.MarkFailed("offering/3", "persistent error", 3, time.Second, time.Minute)
		}
		record := state.GetRecord("offering/3")
		if record.State != SyncStateDeadLettered {
			t.Errorf("state = %v, want dead_lettered", record.State)
		}
		if len(state.DeadLetterQueue) != 1 {
			t.Errorf("dead letter queue size = %d, want 1", len(state.DeadLetterQueue))
		}
	})

	t.Run("mark out of sync", func(t *testing.T) {
		state.MarkSynced("offering/4", "waldur-uuid-4", "old-checksum", 1)
		state.MarkOutOfSync("offering/4", 2, "new-checksum")
		record := state.GetRecord("offering/4")
		if record.State != SyncStateOutOfSync {
			t.Errorf("state = %v, want out_of_sync", record.State)
		}
		if record.ChainVersion != 2 {
			t.Errorf("ChainVersion = %d, want 2", record.ChainVersion)
		}
		if record.Checksum != "new-checksum" {
			t.Errorf("Checksum = %s, want new-checksum", record.Checksum)
		}
	})

	t.Run("needs sync offerings", func(t *testing.T) {
		state.GetOrCreateRecord("offering/5") // Pending state
		needs := state.NeedsSyncOfferings()
		found := false
		for _, id := range needs {
			if id == "offering/5" {
				found = true
				break
			}
		}
		if !found {
			t.Error("offering/5 should need sync")
		}
	})

	t.Run("reprocess dead letter", func(t *testing.T) {
		ok := state.ReprocessDeadLetter("offering/3")
		if !ok {
			t.Error("reprocess should succeed")
		}
		record := state.GetRecord("offering/3")
		if record.State != SyncStatePending {
			t.Errorf("state = %v, want pending after reprocess", record.State)
		}
		if record.RetryCount != 0 {
			t.Errorf("RetryCount = %d, want 0 after reprocess", record.RetryCount)
		}
		if len(state.DeadLetterQueue) != 0 {
			t.Errorf("dead letter queue size = %d, want 0", len(state.DeadLetterQueue))
		}
	})
}

func TestOfferingSyncStateStore(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "offering_sync_state.json")
	store := NewOfferingSyncStateStore(statePath)

	// Load non-existent creates new
	state, err := store.Load("virtengine1provider123")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if state.ProviderAddress != "virtengine1provider123" {
		t.Errorf("ProviderAddress = %s, want virtengine1provider123", state.ProviderAddress)
	}

	// Add data and save
	state.MarkSynced("offering/1", "waldur-uuid-1", "checksum", 1)
	if err := store.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("state file should exist")
	}

	// Load and verify
	loaded, err := store.Load("virtengine1provider123")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.GetRecord("offering/1") == nil {
		t.Error("loaded state should contain offering/1")
	}

	// Delete
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("state file should be deleted")
	}
}

func TestOfferingSyncWorkerConfig(t *testing.T) {
	cfg := DefaultOfferingSyncWorkerConfig()

	if cfg.EventBuffer != 100 {
		t.Errorf("EventBuffer = %d, want 100", cfg.EventBuffer)
	}
	if cfg.SyncIntervalSeconds != 300 {
		t.Errorf("SyncIntervalSeconds = %d, want 300", cfg.SyncIntervalSeconds)
	}
	if !cfg.ReconcileOnStartup {
		t.Error("ReconcileOnStartup should be true")
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.CurrencyDenominator != 1000000 {
		t.Errorf("CurrencyDenominator = %d, want 1000000", cfg.CurrencyDenominator)
	}
}

func TestOfferingSyncTask(t *testing.T) {
	offering := sampleCompute()
	task := &OfferingSyncTask{
		OfferingID: offering.ID.String(),
		Action:     SyncActionCreate,
		Offering:   offering,
		Timestamp:  time.Now().UTC(),
	}

	if task.OfferingID == "" {
		t.Error("OfferingID should not be empty")
	}
	if task.Action != SyncActionCreate {
		t.Errorf("Action = %s, want create", task.Action)
	}
	if task.Offering == nil {
		t.Error("Offering should not be nil")
	}
}

func TestSyncActions(t *testing.T) {
	tests := []struct {
		action SyncAction
		want   string
	}{
		{SyncActionCreate, "create"},
		{SyncActionUpdate, "update"},
		{SyncActionDisable, "disable"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.action) != tt.want {
				t.Errorf("action = %s, want %s", tt.action, tt.want)
			}
		})
	}
}

func TestSyncStates(t *testing.T) {
	tests := []struct {
		state SyncState
		want  string
	}{
		{SyncStatePending, "pending"},
		{SyncStateSynced, "synced"},
		{SyncStateFailed, "failed"},
		{SyncStateRetrying, "retrying"},
		{SyncStateDeadLettered, "dead_lettered"},
		{SyncStateOutOfSync, "out_of_sync"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.state) != tt.want {
				t.Errorf("state = %s, want %s", tt.state, tt.want)
			}
		})
	}
}

func TestOfferingSyncAuditEntry(t *testing.T) {
	entry := OfferingSyncAuditEntry{
		Timestamp:       time.Now().UTC(),
		OfferingID:      "offering/1",
		WaldurUUID:      "waldur-uuid-1",
		Action:          SyncActionCreate,
		Success:         true,
		Duration:        time.Millisecond * 100,
		RetryCount:      0,
		ProviderAddress: "virtengine1provider123",
		Checksum:        "checksum123",
	}

	if entry.OfferingID != "offering/1" {
		t.Errorf("OfferingID = %s, want offering/1", entry.OfferingID)
	}
	if !entry.Success {
		t.Error("Success should be true")
	}
}

func TestReconciliationAuditEntry(t *testing.T) {
	entry := ReconciliationAuditEntry{
		Timestamp:        time.Now().UTC(),
		ProviderAddress:  "virtengine1provider123",
		OfferingsChecked: 100,
		DriftDetected:    5,
		OfferingsQueued:  5,
		Duration:         time.Second * 2,
	}

	if entry.OfferingsChecked != 100 {
		t.Errorf("OfferingsChecked = %d, want 100", entry.OfferingsChecked)
	}
	if entry.DriftDetected != 5 {
		t.Errorf("DriftDetected = %d, want 5", entry.DriftDetected)
	}
}

func TestDeadLetterAuditEntry(t *testing.T) {
	entry := DeadLetterAuditEntry{
		Timestamp:       time.Now().UTC(),
		OfferingID:      "offering/1",
		ProviderAddress: "virtengine1provider123",
		TotalAttempts:   5,
		LastError:       "connection timeout",
		Action:          "dead_lettered",
	}

	if entry.TotalAttempts != 5 {
		t.Errorf("TotalAttempts = %d, want 5", entry.TotalAttempts)
	}
	if entry.Action != "dead_lettered" {
		t.Errorf("Action = %s, want dead_lettered", entry.Action)
	}
}

func TestDefaultAuditLogger(t *testing.T) {
	logger := NewDefaultAuditLogger("[test-audit]")

	// Should not panic
	logger.LogSyncAttempt(OfferingSyncAuditEntry{
		OfferingID: "test-offering",
		Action:     SyncActionCreate,
		Success:    true,
	})

	logger.LogReconciliation(ReconciliationAuditEntry{
		ProviderAddress:  "test-provider",
		OfferingsChecked: 10,
	})

	logger.LogDeadLetter(DeadLetterAuditEntry{
		OfferingID: "test-offering",
		Action:     "dead_lettered",
	})
}

func TestOfferingSyncWorkerMetrics(t *testing.T) {
	metrics := &OfferingSyncWorkerMetrics{
		WorkerUptime: time.Now().UTC(),
	}

	if metrics.SyncsTotal != 0 {
		t.Errorf("SyncsTotal = %d, want 0", metrics.SyncsTotal)
	}
	if metrics.WorkerUptime.IsZero() {
		t.Error("WorkerUptime should be set")
	}
}

func TestOfferingSyncPrometheusMetrics(t *testing.T) {
	metrics := &OfferingSyncPrometheusMetrics{}

	metrics.SyncsTotal.Add(1)
	metrics.SyncsSuccessful.Add(1)
	metrics.ReconciliationsRun.Add(1)

	if metrics.SyncsTotal.Load() != 1 {
		t.Errorf("SyncsTotal = %d, want 1", metrics.SyncsTotal.Load())
	}
	if metrics.SyncsSuccessful.Load() != 1 {
		t.Errorf("SyncsSuccessful = %d, want 1", metrics.SyncsSuccessful.Load())
	}
	if metrics.ReconciliationsRun.Load() != 1 {
		t.Errorf("ReconciliationsRun = %d, want 1", metrics.ReconciliationsRun.Load())
	}
}

func TestEventTypeToAction(t *testing.T) {
	w := &OfferingSyncWorker{}

	tests := []struct {
		eventType marketplace.MarketplaceEventType
		want      SyncAction
	}{
		{marketplace.EventOfferingCreated, SyncActionCreate},
		{marketplace.EventOfferingUpdated, SyncActionUpdate},
		{marketplace.EventOfferingTerminated, SyncActionDisable},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			got := w.eventTypeToAction(tt.eventType)
			if got != tt.want {
				t.Errorf("eventTypeToAction(%s) = %s, want %s", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestStateToAction(t *testing.T) {
	w := &OfferingSyncWorker{}

	tests := []struct {
		state string
		want  string
	}{
		{"Active", "activate"},
		{"Paused", "pause"},
		{"Archived", "archive"},
		{"Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := w.stateToAction(tt.state)
			if got != tt.want {
				t.Errorf("stateToAction(%s) = %s, want %s", tt.state, got, tt.want)
			}
		})
	}
}

func TestWaldurOfferingTypeMapping(t *testing.T) {
	tests := []struct {
		category marketplace.OfferingCategory
		want     string
	}{
		{marketplace.OfferingCategoryCompute, "VirtEngine.Compute"},
		{marketplace.OfferingCategoryStorage, "VirtEngine.Storage"},
		{marketplace.OfferingCategoryNetwork, "VirtEngine.Network"},
		{marketplace.OfferingCategoryHPC, "VirtEngine.HPC"},
		{marketplace.OfferingCategoryGPU, "VirtEngine.GPU"},
		{marketplace.OfferingCategoryML, "VirtEngine.ML"},
		{marketplace.OfferingCategoryOther, "VirtEngine.Generic"},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			got := marketplace.WaldurOfferingType[tt.category]
			if got != tt.want {
				t.Errorf("WaldurOfferingType[%s] = %s, want %s", tt.category, got, tt.want)
			}
		})
	}
}

func TestWaldurOfferingStateMapping(t *testing.T) {
	tests := []struct {
		state marketplace.OfferingState
		want  string
	}{
		{marketplace.OfferingStateActive, "Active"},
		{marketplace.OfferingStatePaused, "Paused"},
		{marketplace.OfferingStateSuspended, "Archived"},
		{marketplace.OfferingStateDeprecated, "Paused"},
		{marketplace.OfferingStateTerminated, "Archived"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			got := marketplace.WaldurOfferingState[tt.state]
			if got != tt.want {
				t.Errorf("WaldurOfferingState[%s] = %s, want %s", tt.state, got, tt.want)
			}
		})
	}
}

// TestOfferingSyncWorkerCreation tests worker creation with nil marketplace client.
func TestOfferingSyncWorkerCreation(t *testing.T) {
	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.ProviderAddress = "virtengine1provider123"

	// Should fail with nil marketplace client
	_, err := NewOfferingSyncWorker(cfg, nil)
	if err == nil {
		t.Error("should fail with nil marketplace client")
	}
}

// TestOfferingSyncWorkerWithLogger tests worker creation with custom logger.
func TestOfferingSyncWorkerWithLogger(t *testing.T) {
	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.ProviderAddress = "virtengine1provider123"

	// Should fail with nil marketplace client
	customLogger := NewDefaultAuditLogger("[custom]")
	_, err := NewOfferingSyncWorkerWithLogger(cfg, nil, customLogger)
	if err == nil {
		t.Error("should fail with nil marketplace client")
	}
}

// TestSyncMetrics tests the sync metrics structure.
func TestSyncMetrics(t *testing.T) {
	metrics := &SyncMetrics{
		TotalSyncs:         100,
		SuccessfulSyncs:    95,
		FailedSyncs:        5,
		DeadLettered:       2,
		DriftDetections:    10,
		ReconciliationsRun: 20,
	}

	if metrics.TotalSyncs != 100 {
		t.Errorf("TotalSyncs = %d, want 100", metrics.TotalSyncs)
	}
	if metrics.SuccessfulSyncs != 95 {
		t.Errorf("SuccessfulSyncs = %d, want 95", metrics.SuccessfulSyncs)
	}
	successRate := float64(metrics.SuccessfulSyncs) / float64(metrics.TotalSyncs)
	if successRate < 0.95 {
		t.Errorf("success rate = %.2f, want >= 0.95", successRate)
	}
}

// TestDeadLetterItem tests the dead letter item structure.
func TestDeadLetterItem(t *testing.T) {
	now := time.Now().UTC()
	item := &DeadLetterItem{
		OfferingID:     "offering/1",
		Action:         "update",
		LastError:      "connection timeout",
		RetryCount:     5,
		FirstAttemptAt: now.Add(-time.Hour),
		DeadLetteredAt: now,
		Checksum:       "checksum123",
		ChainVersion:   3,
	}

	if item.OfferingID != "offering/1" {
		t.Errorf("OfferingID = %s, want offering/1", item.OfferingID)
	}
	if item.RetryCount != 5 {
		t.Errorf("RetryCount = %d, want 5", item.RetryCount)
	}
	if item.ChainVersion != 3 {
		t.Errorf("ChainVersion = %d, want 3", item.ChainVersion)
	}
}

// TestOfferingSyncRecord tests the sync record structure.
func TestOfferingSyncRecord(t *testing.T) {
	now := time.Now().UTC()
	record := &OfferingSyncRecord{
		OfferingID:    "offering/1",
		WaldurUUID:    "waldur-uuid-1",
		State:         SyncStateSynced,
		ChainVersion:  1,
		SyncedVersion: 1,
		Checksum:      "checksum123",
		LastSyncedAt:  &now,
		CreatedAt:     now,
	}

	if record.OfferingID != "offering/1" {
		t.Errorf("OfferingID = %s, want offering/1", record.OfferingID)
	}
	if record.State != SyncStateSynced {
		t.Errorf("State = %s, want synced", record.State)
	}
	if record.LastSyncedAt == nil {
		t.Error("LastSyncedAt should be set")
	}
}

// TestQueueSync tests the queue sync functionality.
func TestQueueSync(t *testing.T) {
	// Create a minimal worker for testing queue operations
	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.EventBuffer = 10

	worker := &OfferingSyncWorker{
		cfg:       cfg,
		syncQueue: make(chan *OfferingSyncTask, cfg.EventBuffer),
		metrics:   &OfferingSyncWorkerMetrics{},
	}

	offering := sampleCompute()
	err := worker.QueueSync(offering.ID.String(), SyncActionUpdate, offering)
	if err != nil {
		t.Errorf("QueueSync failed: %v", err)
	}

	// Verify task was queued
	select {
	case task := <-worker.syncQueue:
		if task.OfferingID != offering.ID.String() {
			t.Errorf("queued task OfferingID = %s, want %s", task.OfferingID, offering.ID.String())
		}
		if task.Action != SyncActionUpdate {
			t.Errorf("queued task Action = %s, want update", task.Action)
		}
	default:
		t.Error("no task was queued")
	}
}

// TestQueueSyncFull tests queue full behavior.
func TestQueueSyncFull(t *testing.T) {
	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.EventBuffer = 1 // Small buffer for testing

	worker := &OfferingSyncWorker{
		cfg:       cfg,
		syncQueue: make(chan *OfferingSyncTask, cfg.EventBuffer),
		metrics:   &OfferingSyncWorkerMetrics{},
	}

	offering := sampleCompute()

	// Fill the queue
	_ = worker.QueueSync("offering/1", SyncActionCreate, offering)

	// Next queue should fail
	err := worker.QueueSync("offering/2", SyncActionCreate, offering)
	if err == nil {
		t.Error("QueueSync should fail when queue is full")
	}
	if !errors.Is(err, errors.New("sync queue full")) && err.Error() != "sync queue full" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestReconcile tests the reconcile method.
func TestReconcile(t *testing.T) {
	ctx := context.Background()

	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.EventBuffer = 100

	state := NewOfferingSyncState("virtengine1provider123")
	// Add some offerings that need sync
	state.GetOrCreateRecord("offering/1")
	state.MarkOutOfSync("offering/1", 2, "new-checksum")

	worker := &OfferingSyncWorker{
		cfg:         cfg,
		state:       state,
		syncQueue:   make(chan *OfferingSyncTask, cfg.EventBuffer),
		metrics:     &OfferingSyncWorkerMetrics{},
		promMetrics: &OfferingSyncPrometheusMetrics{},
		auditLogger: NewDefaultAuditLogger("[test]"),
	}

	err := worker.Reconcile(ctx)
	if err != nil {
		t.Errorf("Reconcile failed: %v", err)
	}

	// Verify task was queued
	select {
	case task := <-worker.syncQueue:
		if task.OfferingID != "offering/1" {
			t.Errorf("queued task OfferingID = %s, want offering/1", task.OfferingID)
		}
	default:
		t.Error("reconcile should have queued a task")
	}
}

// TestWorkerMetricsSnapshot tests the Metrics method.
func TestWorkerMetricsSnapshot(t *testing.T) {
	cfg := DefaultOfferingSyncWorkerConfig()
	cfg.EventBuffer = 10

	worker := &OfferingSyncWorker{
		cfg:       cfg,
		syncQueue: make(chan *OfferingSyncTask, cfg.EventBuffer),
		metrics: &OfferingSyncWorkerMetrics{
			SyncsTotal:      100,
			SyncsSuccessful: 95,
			SyncsFailed:     5,
			WorkerUptime:    time.Now().UTC().Add(-time.Hour),
		},
	}

	snapshot := worker.Metrics()

	if snapshot.SyncsTotal != 100 {
		t.Errorf("SyncsTotal = %d, want 100", snapshot.SyncsTotal)
	}
	if snapshot.SyncsSuccessful != 95 {
		t.Errorf("SyncsSuccessful = %d, want 95", snapshot.SyncsSuccessful)
	}
	if snapshot.SyncsFailed != 5 {
		t.Errorf("SyncsFailed = %d, want 5", snapshot.SyncsFailed)
	}
	if snapshot.QueueDepth != 0 {
		t.Errorf("QueueDepth = %d, want 0", snapshot.QueueDepth)
	}
}
