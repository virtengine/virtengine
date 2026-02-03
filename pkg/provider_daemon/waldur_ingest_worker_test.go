// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-3D: Tests for Waldur ingestion worker and state management.
package provider_daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

const testComputeTag = "compute"

// MockOfferingSubmitter implements OfferingSubmitter for testing.
type MockOfferingSubmitter struct {
	CreateOfferingFn          func(ctx context.Context, offering *marketplace.Offering) (string, error)
	UpdateOfferingFn          func(ctx context.Context, offeringID string, offering *marketplace.Offering) error
	DeprecateOfferingFn       func(ctx context.Context, offeringID string) error
	GetNextOfferingSequenceFn func(ctx context.Context, providerAddress string) (uint64, error)
	ValidateProviderVEIDFn    func(ctx context.Context, providerAddress string, minScore uint32) error

	CreateCalls    int
	UpdateCalls    int
	DeprecateCalls int
	SequenceNumber uint64
}

func (m *MockOfferingSubmitter) CreateOffering(ctx context.Context, offering *marketplace.Offering) (string, error) {
	m.CreateCalls++
	if m.CreateOfferingFn != nil {
		return m.CreateOfferingFn(ctx, offering)
	}
	return offering.ID.String(), nil
}

func (m *MockOfferingSubmitter) UpdateOffering(ctx context.Context, offeringID string, offering *marketplace.Offering) error {
	m.UpdateCalls++
	if m.UpdateOfferingFn != nil {
		return m.UpdateOfferingFn(ctx, offeringID, offering)
	}
	return nil
}

func (m *MockOfferingSubmitter) DeprecateOffering(ctx context.Context, offeringID string) error {
	m.DeprecateCalls++
	if m.DeprecateOfferingFn != nil {
		return m.DeprecateOfferingFn(ctx, offeringID)
	}
	return nil
}

func (m *MockOfferingSubmitter) GetNextOfferingSequence(ctx context.Context, providerAddress string) (uint64, error) {
	m.SequenceNumber++
	if m.GetNextOfferingSequenceFn != nil {
		return m.GetNextOfferingSequenceFn(ctx, providerAddress)
	}
	return m.SequenceNumber, nil
}

func (m *MockOfferingSubmitter) ValidateProviderVEID(ctx context.Context, providerAddress string, minScore uint32) error {
	if m.ValidateProviderVEIDFn != nil {
		return m.ValidateProviderVEIDFn(ctx, providerAddress, minScore)
	}
	return nil
}

// Sample Waldur offering payloads for testing
var sampleWaldurCompute = &marketplace.WaldurOfferingImport{
	UUID:         "uuid-compute-001",
	Name:         "Basic Compute Instance",
	Description:  "A basic compute instance with 2 vCPUs and 4GB RAM",
	Type:         "VirtEngine.Compute",
	State:        "Active",
	CategoryUUID: "cat-compute-uuid",
	CustomerUUID: "cust-provider-001",
	Shared:       true,
	Billable:     true,
	Attributes: map[string]interface{}{
		"tags":                  []interface{}{testComputeTag, "basic"},
		"regions":               []interface{}{"us-east-1", "eu-west-1"},
		"spec_vcpu":             "2",
		"spec_memory_gb":        "4",
		"spec_disk_gb":          "50",
		"ve_min_identity_score": float64(10),
		"ve_require_mfa":        true,
	},
	Components: []marketplace.WaldurPricingComponent{
		{
			Type:         "usage",
			Name:         "base",
			MeasuredUnit: "hour",
			BillingType:  "usage",
			Price:        "0.050000",
		},
	},
	Created:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	Modified: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
}

var sampleWaldurHPC = &marketplace.WaldurOfferingImport{
	UUID:         "uuid-hpc-001",
	Name:         "HPC Cluster Access",
	Description:  "High-performance computing cluster with SLURM scheduling",
	Type:         "VirtEngine.HPC",
	State:        "Active",
	CategoryUUID: "cat-hpc-uuid",
	CustomerUUID: "cust-provider-001",
	Shared:       true,
	Billable:     true,
	Attributes: map[string]interface{}{
		"tags":           []interface{}{"hpc", "slurm", "research"},
		"regions":        []interface{}{"us-west-2"},
		"spec_cpu_limit": "1000",
		"spec_gpu_limit": "8",
		"spec_ram_gb":    "512",
	},
	Components: []marketplace.WaldurPricingComponent{
		{
			Type:         "usage",
			Name:         "cpu_hours",
			MeasuredUnit: "cpu_hour",
			BillingType:  "usage",
			Price:        "0.010000",
		},
		{
			Type:         "usage",
			Name:         "gpu_hours",
			MeasuredUnit: "gpu_hour",
			BillingType:  "usage",
			Price:        "1.500000",
		},
	},
	Created:  time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
	Modified: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
}

var sampleWaldurStorage = &marketplace.WaldurOfferingImport{
	UUID:         "uuid-storage-001",
	Name:         "Block Storage",
	Description:  "High-performance block storage for VMs",
	Type:         "VirtEngine.Storage",
	State:        "Paused",
	CategoryUUID: "cat-storage-uuid",
	CustomerUUID: "cust-provider-001",
	Shared:       true,
	Billable:     true,
	Attributes: map[string]interface{}{
		"tags":            []interface{}{"storage", "ssd"},
		"regions":         []interface{}{"us-east-1"},
		"spec_type":       "ssd",
		"spec_iops":       "3000",
		"spec_throughput": "250",
	},
	Components: []marketplace.WaldurPricingComponent{
		{
			Type:         "usage",
			Name:         "base",
			MeasuredUnit: "gb_month",
			BillingType:  "monthly",
			Price:        "0.100000",
		},
	},
	Created:  time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
	Modified: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
}

func TestWaldurOfferingImport_Validate(t *testing.T) {
	cfg := marketplace.DefaultIngestConfig()
	cfg.CustomerProviderMap = map[string]string{
		"cust-provider-001": "virtengine1provider123",
	}

	tests := []struct {
		name       string
		offering   *marketplace.WaldurOfferingImport
		wantValid  bool
		wantErrors int
	}{
		{
			name:       "valid compute offering",
			offering:   sampleWaldurCompute,
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name:       "valid hpc offering",
			offering:   sampleWaldurHPC,
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name:       "valid storage offering",
			offering:   sampleWaldurStorage,
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "missing UUID",
			offering: &marketplace.WaldurOfferingImport{
				Name:         "Test",
				CustomerUUID: "cust-provider-001",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "missing name",
			offering: &marketplace.WaldurOfferingImport{
				UUID:         "test-uuid",
				CustomerUUID: "cust-provider-001",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "unmapped customer",
			offering: &marketplace.WaldurOfferingImport{
				UUID:         "test-uuid",
				Name:         "Test",
				CustomerUUID: "unknown-customer",
			},
			wantValid:  false,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.offering.Validate(cfg)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if len(result.Errors) < tt.wantErrors {
				t.Errorf("Errors count = %d, want >= %d", len(result.Errors), tt.wantErrors)
			}
		})
	}
}

func TestWaldurOfferingImport_ResolveCategory(t *testing.T) {
	cfg := marketplace.DefaultIngestConfig()

	tests := []struct {
		name     string
		offering *marketplace.WaldurOfferingImport
		want     marketplace.OfferingCategory
	}{
		{
			name:     "compute type",
			offering: &marketplace.WaldurOfferingImport{Type: "VirtEngine.Compute"},
			want:     marketplace.OfferingCategoryCompute,
		},
		{
			name:     "hpc type",
			offering: &marketplace.WaldurOfferingImport{Type: "VirtEngine.HPC"},
			want:     marketplace.OfferingCategoryHPC,
		},
		{
			name:     "storage type",
			offering: &marketplace.WaldurOfferingImport{Type: "VirtEngine.Storage"},
			want:     marketplace.OfferingCategoryStorage,
		},
		{
			name:     "gpu type",
			offering: &marketplace.WaldurOfferingImport{Type: "VirtEngine.GPU"},
			want:     marketplace.OfferingCategoryGPU,
		},
		{
			name:     "infer from vm keyword",
			offering: &marketplace.WaldurOfferingImport{Type: "Custom.VirtualMachine"},
			want:     marketplace.OfferingCategoryCompute,
		},
		{
			name:     "infer from slurm keyword",
			offering: &marketplace.WaldurOfferingImport{Type: "Custom.SlurmCluster"},
			want:     marketplace.OfferingCategoryHPC,
		},
		{
			name:     "unknown type",
			offering: &marketplace.WaldurOfferingImport{Type: "Custom.Unknown"},
			want:     marketplace.OfferingCategoryOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.offering.ResolveCategory(cfg)
			if got != tt.want {
				t.Errorf("ResolveCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWaldurOfferingImport_ResolveState(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  marketplace.OfferingState
	}{
		{"active", "Active", marketplace.OfferingStateActive},
		{"paused", "Paused", marketplace.OfferingStatePaused},
		{"archived", "Archived", marketplace.OfferingStateTerminated},
		{"unknown", "Unknown", marketplace.OfferingStatePaused},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &marketplace.WaldurOfferingImport{State: tt.state}
			got := o.ResolveState()
			if got != tt.want {
				t.Errorf("ResolveState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWaldurOfferingImport_ResolvePricing(t *testing.T) {
	cfg := marketplace.DefaultIngestConfig()

	t.Run("base price only", func(t *testing.T) {
		o := &marketplace.WaldurOfferingImport{
			Components: []marketplace.WaldurPricingComponent{
				{Name: "base", Price: "0.050000", BillingType: "usage"},
			},
		}
		pricing := o.ResolvePricing(cfg)
		if pricing.BasePrice != 50000 {
			t.Errorf("BasePrice = %d, want 50000", pricing.BasePrice)
		}
		if pricing.Model != marketplace.PricingModelHourly {
			t.Errorf("Model = %v, want hourly", pricing.Model)
		}
	})

	t.Run("with usage rates", func(t *testing.T) {
		o := sampleWaldurHPC
		pricing := o.ResolvePricing(cfg)
		if len(pricing.UsageRates) != 2 {
			t.Errorf("UsageRates count = %d, want 2", len(pricing.UsageRates))
		}
		if pricing.UsageRates["cpu_hours"] != 10000 {
			t.Errorf("cpu_hours rate = %d, want 10000", pricing.UsageRates["cpu_hours"])
		}
		if pricing.UsageRates["gpu_hours"] != 1500000 {
			t.Errorf("gpu_hours rate = %d, want 1500000", pricing.UsageRates["gpu_hours"])
		}
	})

	t.Run("monthly billing", func(t *testing.T) {
		o := sampleWaldurStorage
		pricing := o.ResolvePricing(cfg)
		if pricing.Model != marketplace.PricingModelMonthly {
			t.Errorf("Model = %v, want monthly", pricing.Model)
		}
	})
}

func TestWaldurOfferingImport_ExtractTags(t *testing.T) {
	tags := sampleWaldurCompute.ExtractTags()
	if len(tags) != 2 {
		t.Errorf("tags count = %d, want 2", len(tags))
	}
	if tags[0] != testComputeTag {
		t.Errorf("first tag = %s, want '%s'", tags[0], testComputeTag)
	}
}

func TestWaldurOfferingImport_ExtractRegions(t *testing.T) {
	cfg := marketplace.DefaultIngestConfig()
	regions := sampleWaldurCompute.ExtractRegions(cfg)
	if len(regions) != 2 {
		t.Errorf("regions count = %d, want 2", len(regions))
	}
}

func TestWaldurOfferingImport_ExtractSpecifications(t *testing.T) {
	specs := sampleWaldurCompute.ExtractSpecifications()
	if specs["vcpu"] != "2" {
		t.Errorf("vcpu = %s, want '2'", specs["vcpu"])
	}
	if specs["memory_gb"] != "4" {
		t.Errorf("memory_gb = %s, want '4'", specs["memory_gb"])
	}
}

func TestWaldurOfferingImport_ToOffering(t *testing.T) {
	cfg := marketplace.DefaultIngestConfig()
	cfg.CustomerProviderMap = map[string]string{
		"cust-provider-001": "virtengine1provider123",
	}

	offering := sampleWaldurCompute.ToOffering("virtengine1provider123", 1, cfg)

	if offering.ID.ProviderAddress != "virtengine1provider123" {
		t.Errorf("ProviderAddress = %s, want virtengine1provider123", offering.ID.ProviderAddress)
	}
	if offering.ID.Sequence != 1 {
		t.Errorf("Sequence = %d, want 1", offering.ID.Sequence)
	}
	if offering.Name != sampleWaldurCompute.Name {
		t.Errorf("Name = %s, want %s", offering.Name, sampleWaldurCompute.Name)
	}
	if offering.Category != marketplace.OfferingCategoryCompute {
		t.Errorf("Category = %v, want compute", offering.Category)
	}
	if offering.State != marketplace.OfferingStateActive {
		t.Errorf("State = %v, want active", offering.State)
	}
	if offering.Pricing.BasePrice != 50000 {
		t.Errorf("BasePrice = %d, want 50000", offering.Pricing.BasePrice)
	}
	if offering.IdentityRequirement.MinScore != 10 {
		t.Errorf("MinScore = %d, want 10", offering.IdentityRequirement.MinScore)
	}
}

func TestWaldurOfferingImport_IngestChecksum(t *testing.T) {
	checksum1 := sampleWaldurCompute.IngestChecksum()
	checksum2 := sampleWaldurCompute.IngestChecksum()

	if checksum1 != checksum2 {
		t.Errorf("checksums should be deterministic")
	}

	modified := *sampleWaldurCompute
	modified.Name = "Modified Name"
	checksum3 := modified.IngestChecksum()

	if checksum1 == checksum3 {
		t.Errorf("checksums should differ for modified offerings")
	}
}

func TestWaldurIngestState(t *testing.T) {
	state := NewWaldurIngestState("cust-001", "provider-001")

	t.Run("mark ingested", func(t *testing.T) {
		state.MarkIngested("waldur-001", "chain/1", "checksum123", 1)
		record := state.GetRecord("waldur-001")
		if record == nil {
			t.Fatal("record should exist")
		}
		if record.State != IngestRecordStateIngested {
			t.Errorf("state = %v, want ingested", record.State)
		}
		if record.ChainOfferingID != "chain/1" {
			t.Errorf("ChainOfferingID = %s, want chain/1", record.ChainOfferingID)
		}
	})

	t.Run("mark failed with retry", func(t *testing.T) {
		state.GetOrCreateRecord("waldur-002", "Test Offering")
		deadLettered := state.MarkFailed("waldur-002", "test error", 3, time.Second, time.Minute)
		if deadLettered {
			t.Error("should not be dead-lettered on first failure")
		}
		record := state.GetRecord("waldur-002")
		if record.State != IngestRecordStateRetrying {
			t.Errorf("state = %v, want retrying", record.State)
		}
		if record.RetryCount != 1 {
			t.Errorf("RetryCount = %d, want 1", record.RetryCount)
		}
	})

	t.Run("dead letter after max retries", func(t *testing.T) {
		state.GetOrCreateRecord("waldur-003", "Test Offering")
		for i := 0; i <= 3; i++ {
			state.MarkFailed("waldur-003", "persistent error", 3, time.Second, time.Minute)
		}
		record := state.GetRecord("waldur-003")
		if record.State != IngestRecordStateDeadLettered {
			t.Errorf("state = %v, want dead_lettered", record.State)
		}
		if len(state.DeadLetterQueue) != 1 {
			t.Errorf("dead letter queue size = %d, want 1", len(state.DeadLetterQueue))
		}
	})

	t.Run("needs ingest", func(t *testing.T) {
		state.GetOrCreateRecord("waldur-004", "Pending Offering")
		needs := state.NeedsIngestOfferings()
		found := false
		for _, uuid := range needs {
			if uuid == "waldur-004" {
				found = true
				break
			}
		}
		if !found {
			t.Error("waldur-004 should need ingestion")
		}
	})

	t.Run("reprocess dead letter", func(t *testing.T) {
		ok := state.ReprocessDeadLetter("waldur-003")
		if !ok {
			t.Error("reprocess should succeed")
		}
		record := state.GetRecord("waldur-003")
		if record.State != IngestRecordStatePending {
			t.Errorf("state = %v, want pending after reprocess", record.State)
		}
		if len(state.DeadLetterQueue) != 0 {
			t.Errorf("dead letter queue size = %d, want 0", len(state.DeadLetterQueue))
		}
	})
}

func TestWaldurIngestStateStore(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "ingest_state.json")
	store := NewWaldurIngestStateStore(statePath)

	// Load non-existent creates new
	state, err := store.Load("cust-001", "provider-001")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if state.WaldurCustomerUUID != "cust-001" {
		t.Errorf("WaldurCustomerUUID = %s, want cust-001", state.WaldurCustomerUUID)
	}

	// Add data and save
	state.MarkIngested("waldur-001", "chain/1", "checksum", 1)
	if err := store.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("state file should exist")
	}

	// Load and verify
	loaded, err := store.Load("cust-001", "provider-001")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.GetRecord("waldur-001") == nil {
		t.Error("loaded state should contain waldur-001")
	}

	// Delete
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("state file should be deleted")
	}
}

func TestWaldurIngestStats(t *testing.T) {
	state := NewWaldurIngestState("cust-001", "provider-001")

	state.MarkIngested("w1", "c1", "cs1", 1)
	state.MarkIngested("w2", "c2", "cs2", 1)
	state.GetOrCreateRecord("w3", "Pending")
	state.MarkSkipped("w3", "skipped")

	stats := state.GetStats()
	if stats.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d, want 3", stats.TotalRecords)
	}
	if stats.Ingested != 2 {
		t.Errorf("Ingested = %d, want 2", stats.Ingested)
	}
	if stats.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", stats.Skipped)
	}
}

// TestMockAuditLogger tests the audit logger interface.
func TestMockAuditLogger(t *testing.T) {
	logger := NewDefaultIngestAuditLogger("[test]")

	// Should not panic
	logger.LogIngestAttempt(IngestAuditEntry{
		WaldurUUID: "test-uuid",
		Action:     "create",
		Success:    true,
	})

	logger.LogReconciliation(IngestReconciliationAuditEntry{
		WaldurCustomer:   "test-customer",
		OfferingsChecked: 10,
	})

	logger.LogDeadLetter(IngestDeadLetterAuditEntry{
		WaldurUUID: "test-uuid",
		Action:     "dead_lettered",
	})
}
