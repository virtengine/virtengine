package provider_daemon

import (
	"context"
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

const testPubProviderAddress = "provider123"

// publicationMockOfferingSubmitter implements OfferingSubmitter for publication testing.
type publicationMockOfferingSubmitter struct {
	offerings       map[string]*marketplace.Offering
	sequence        uint64
	createErr       error
	updateErr       error
	deprecateErr    error
	validateVEIDErr error
}

func newPublicationMockOfferingSubmitter() *publicationMockOfferingSubmitter {
	return &publicationMockOfferingSubmitter{
		offerings: make(map[string]*marketplace.Offering),
		sequence:  1,
	}
}

func (m *publicationMockOfferingSubmitter) CreateOffering(_ context.Context, offering *marketplace.Offering) (string, error) {
	if m.createErr != nil {
		return "", m.createErr
	}
	id := offering.ID.String()
	m.offerings[id] = offering
	return id, nil
}

func (m *publicationMockOfferingSubmitter) UpdateOffering(_ context.Context, offeringID string, offering *marketplace.Offering) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.offerings[offeringID] = offering
	return nil
}

func (m *publicationMockOfferingSubmitter) DeprecateOffering(_ context.Context, offeringID string) error {
	if m.deprecateErr != nil {
		return m.deprecateErr
	}
	if o, ok := m.offerings[offeringID]; ok {
		o.State = marketplace.OfferingStateDeprecated
	}
	return nil
}

func (m *publicationMockOfferingSubmitter) GetNextOfferingSequence(_ context.Context, _ string) (uint64, error) {
	seq := m.sequence
	m.sequence++
	return seq, nil
}

func (m *publicationMockOfferingSubmitter) ValidateProviderVEID(_ context.Context, _ string, _ uint32) error {
	return m.validateVEIDErr
}

func TestOfferingPublicationState_BasicOperations(t *testing.T) {
	state := NewOfferingPublicationState(testPubProviderAddress)

	// Test empty state
	if len(state.ListPublications()) != 0 {
		t.Error("expected empty publications list")
	}

	// Test set and get
	pub := &OfferingPublication{
		WaldurUUID: "uuid-1",
		Name:       "Test Offering",
		Category:   "compute",
		Status:     PublicationStatusPending,
		CreatedAt:  time.Now().UTC(),
	}
	state.SetPublication(pub)

	retrieved := state.GetPublication("uuid-1")
	if retrieved == nil {
		t.Fatal("expected to retrieve publication")
	}
	if retrieved.Name != "Test Offering" {
		t.Errorf("expected name 'Test Offering', got %s", retrieved.Name)
	}

	// Test list returns publication
	pubs := state.ListPublications()
	if len(pubs) != 1 {
		t.Errorf("expected 1 publication, got %d", len(pubs))
	}
}

func TestOfferingPublicationState_FilterByStatus(t *testing.T) {
	state := NewOfferingPublicationState(testPubProviderAddress)

	// Add publications with different statuses
	state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-1",
		Name:       "Pending 1",
		Status:     PublicationStatusPending,
		CreatedAt:  time.Now().UTC(),
	})
	state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-2",
		Name:       "Published 1",
		Status:     PublicationStatusPublished,
		CreatedAt:  time.Now().UTC(),
	})
	state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-3",
		Name:       "Pending 2",
		Status:     PublicationStatusPending,
		CreatedAt:  time.Now().UTC(),
	})

	// Filter pending
	pending := state.FilterByStatus(PublicationStatusPending)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending publications, got %d", len(pending))
	}

	// Filter published
	published := state.FilterByStatus(PublicationStatusPublished)
	if len(published) != 1 {
		t.Errorf("expected 1 published publication, got %d", len(published))
	}

	// Filter non-existent
	deprecated := state.FilterByStatus(PublicationStatusDeprecated)
	if len(deprecated) != 0 {
		t.Errorf("expected 0 deprecated publications, got %d", len(deprecated))
	}
}

func TestOfferingPublicationService_GetOfferingStats(t *testing.T) {
	cfg := DefaultOfferingPublicationConfig()
	cfg.ProviderAddress = testPubProviderAddress
	cfg.AutoPublish = false

	// Create service without marketplace client for stats testing
	svc := &OfferingPublicationService{
		cfg:   cfg,
		state: NewOfferingPublicationState(cfg.ProviderAddress),
	}

	// Add publications with different statuses
	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID:   "uuid-1",
		Status:       PublicationStatusPublished,
		ActiveLeases: 5,
		TotalOrders:  10,
	})
	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID:   "uuid-2",
		Status:       PublicationStatusPending,
		ActiveLeases: 0,
		TotalOrders:  0,
	})
	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID:   "uuid-3",
		Status:       PublicationStatusPaused,
		ActiveLeases: 2,
		TotalOrders:  5,
	})

	stats := svc.GetOfferingStats()

	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
	if stats.Published != 1 {
		t.Errorf("expected published 1, got %d", stats.Published)
	}
	if stats.Pending != 1 {
		t.Errorf("expected pending 1, got %d", stats.Pending)
	}
	if stats.Paused != 1 {
		t.Errorf("expected paused 1, got %d", stats.Paused)
	}
	if stats.TotalLeases != 7 {
		t.Errorf("expected total leases 7, got %d", stats.TotalLeases)
	}
	if stats.TotalOrders != 15 {
		t.Errorf("expected total orders 15, got %d", stats.TotalOrders)
	}
}

func TestOfferingPublicationStatus_Constants(t *testing.T) {
	// Verify status constants are unique
	statuses := map[OfferingPublicationStatus]bool{
		PublicationStatusPending:    true,
		PublicationStatusPublished:  true,
		PublicationStatusFailed:     true,
		PublicationStatusPaused:     true,
		PublicationStatusDeprecated: true,
		PublicationStatusDraft:      true,
	}

	if len(statuses) != 6 {
		t.Error("expected 6 unique status constants")
	}
}

func TestMapWaldurTypeToCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VirtEngine.Compute", "compute"},
		{"VirtEngine.Storage", "storage"},
		{"VirtEngine.Network", "network"},
		{"VirtEngine.HPC", "hpc"},
		{"VirtEngine.GPU", "gpu"},
		{"VirtEngine.ML", "ml"},
		{"VirtEngine.Unknown", "other"},
		{"SomeOther.Type", "other"},
		{"", "other"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := mapWaldurTypeToCategory(tc.input)
			if result != tc.expected {
				t.Errorf("mapWaldurTypeToCategory(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestOfferingPublicationService_ListOfferings(t *testing.T) {
	cfg := DefaultOfferingPublicationConfig()
	cfg.ProviderAddress = testPubProviderAddress
	cfg.AutoPublish = false

	svc := &OfferingPublicationService{
		cfg:   cfg,
		state: NewOfferingPublicationState(cfg.ProviderAddress),
	}

	// Empty list
	offerings := svc.ListOfferings()
	if len(offerings) != 0 {
		t.Errorf("expected 0 offerings, got %d", len(offerings))
	}

	// Add some offerings
	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-1",
		Name:       "Beta Offering",
		Status:     PublicationStatusPending,
	})
	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-2",
		Name:       "Alpha Offering",
		Status:     PublicationStatusPublished,
	})

	offerings = svc.ListOfferings()
	if len(offerings) != 2 {
		t.Errorf("expected 2 offerings, got %d", len(offerings))
	}

	// Verify sorted by name
	if offerings[0].Name != "Alpha Offering" {
		t.Errorf("expected first offering to be 'Alpha Offering', got %s", offerings[0].Name)
	}
}

func TestOfferingPublicationService_GetOffering(t *testing.T) {
	cfg := DefaultOfferingPublicationConfig()
	cfg.ProviderAddress = testPubProviderAddress
	cfg.AutoPublish = false

	svc := &OfferingPublicationService{
		cfg:   cfg,
		state: NewOfferingPublicationState(cfg.ProviderAddress),
	}

	svc.state.SetPublication(&OfferingPublication{
		WaldurUUID: "uuid-1",
		Name:       "Test Offering",
		Status:     PublicationStatusPending,
	})

	// Get existing
	offering := svc.GetOffering("uuid-1")
	if offering == nil {
		t.Fatal("expected to find offering")
	}
	if offering.Name != "Test Offering" {
		t.Errorf("expected name 'Test Offering', got %s", offering.Name)
	}

	// Get non-existent
	offering = svc.GetOffering("non-existent")
	if offering != nil {
		t.Error("expected nil for non-existent offering")
	}
}

func TestDefaultOfferingPublicationConfig(t *testing.T) {
	cfg := DefaultOfferingPublicationConfig()

	if !cfg.AutoPublish {
		t.Error("expected AutoPublish to be true by default")
	}
	if cfg.PollIntervalSeconds != 300 {
		t.Errorf("expected PollIntervalSeconds 300, got %d", cfg.PollIntervalSeconds)
	}
	if cfg.OperationTimeout != 60*time.Second {
		t.Errorf("expected OperationTimeout 60s, got %v", cfg.OperationTimeout)
	}
	if cfg.WaldurCategoryMap == nil {
		t.Error("expected WaldurCategoryMap to be initialized")
	}
}

func TestPublicationMockOfferingSubmitter(t *testing.T) {
	mock := newPublicationMockOfferingSubmitter()
	ctx := context.Background()

	// Test CreateOffering
	offering := &marketplace.Offering{
		ID:   marketplace.OfferingID{ProviderAddress: testPubProviderAddress, Sequence: 1},
		Name: "Test Offering",
	}
	id, err := mock.CreateOffering(ctx, offering)
	if err != nil {
		t.Fatalf("CreateOffering failed: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty offering ID")
	}

	// Verify offering was stored
	if _, exists := mock.offerings[id]; !exists {
		t.Error("offering not stored in mock")
	}

	// Test GetNextOfferingSequence
	seq, err := mock.GetNextOfferingSequence(ctx, testPubProviderAddress)
	if err != nil {
		t.Fatalf("GetNextOfferingSequence failed: %v", err)
	}
	if seq != 1 {
		t.Errorf("expected sequence 1, got %d", seq)
	}

	// Test sequence increments
	seq2, _ := mock.GetNextOfferingSequence(ctx, testPubProviderAddress)
	if seq2 != 2 {
		t.Errorf("expected sequence 2, got %d", seq2)
	}

	// Test ValidateProviderVEID
	err = mock.ValidateProviderVEID(ctx, testPubProviderAddress, 1)
	if err != nil {
		t.Errorf("expected no error from ValidateProviderVEID, got %v", err)
	}
}
