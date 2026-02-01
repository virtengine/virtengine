// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Tests for dispute lifecycle persistence and gateway actions
package payment

import (
	"testing"
	"time"
)

// ============================================================================
// DisputeStore Tests
// ============================================================================

func TestInMemoryDisputeStore_SaveAndGet(t *testing.T) {
	store := NewInMemoryDisputeStore()

	dispute := Dispute{
		ID:              "dp_test123",
		Gateway:         GatewayStripe,
		PaymentIntentID: "pi_test456",
		ChargeID:        "ch_test789",
		Amount:          NewAmount(10000, CurrencyUSD),
		Status:          DisputeStatusNeedsResponse,
		Reason:          DisputeReasonFraudulent,
		EvidenceDueBy:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:       time.Now(),
	}

	record := NewDisputeRecord(dispute, GatewayStripe, "test:create")

	// Save
	err := store.Save(record)
	if err != nil {
		t.Fatalf("Failed to save dispute record: %v", err)
	}

	// Get
	retrieved, found := store.Get("dp_test123")
	if !found {
		t.Fatal("Failed to retrieve dispute record")
	}

	if retrieved.ID != dispute.ID {
		t.Errorf("Expected ID %s, got %s", dispute.ID, retrieved.ID)
	}
	if retrieved.Status != dispute.Status {
		t.Errorf("Expected status %s, got %s", dispute.Status, retrieved.Status)
	}
	if retrieved.Gateway != dispute.Gateway {
		t.Errorf("Expected gateway %s, got %s", dispute.Gateway, retrieved.Gateway)
	}
	if len(retrieved.AuditLog) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(retrieved.AuditLog))
	}
	if retrieved.AuditLog[0].Action != "created" {
		t.Errorf("Expected action 'created', got '%s'", retrieved.AuditLog[0].Action)
	}
}

func TestInMemoryDisputeStore_GetByPaymentIntent(t *testing.T) {
	store := NewInMemoryDisputeStore()

	// Create multiple disputes for the same payment intent
	dispute1 := Dispute{
		ID:              "dp_test1",
		Gateway:         GatewayStripe,
		PaymentIntentID: "pi_shared",
		Status:          DisputeStatusNeedsResponse,
	}
	dispute2 := Dispute{
		ID:              "dp_test2",
		Gateway:         GatewayStripe,
		PaymentIntentID: "pi_shared",
		Status:          DisputeStatusWon,
	}
	dispute3 := Dispute{
		ID:              "dp_test3",
		Gateway:         GatewayStripe,
		PaymentIntentID: "pi_other",
		Status:          DisputeStatusLost,
	}

	_ = store.Save(NewDisputeRecord(dispute1, GatewayStripe, "test"))
	_ = store.Save(NewDisputeRecord(dispute2, GatewayStripe, "test"))
	_ = store.Save(NewDisputeRecord(dispute3, GatewayStripe, "test"))

	// Query by payment intent
	records := store.GetByPaymentIntent("pi_shared")
	if len(records) != 2 {
		t.Errorf("Expected 2 disputes for pi_shared, got %d", len(records))
	}

	records = store.GetByPaymentIntent("pi_other")
	if len(records) != 1 {
		t.Errorf("Expected 1 dispute for pi_other, got %d", len(records))
	}
}

func TestInMemoryDisputeStore_GetByStatus(t *testing.T) {
	store := NewInMemoryDisputeStore()

	disputes := []Dispute{
		{ID: "dp_1", Status: DisputeStatusNeedsResponse},
		{ID: "dp_2", Status: DisputeStatusNeedsResponse},
		{ID: "dp_3", Status: DisputeStatusUnderReview},
		{ID: "dp_4", Status: DisputeStatusWon},
	}

	for _, d := range disputes {
		_ = store.Save(NewDisputeRecord(d, GatewayStripe, "test"))
	}

	needsResponse := store.GetByStatus(DisputeStatusNeedsResponse)
	if len(needsResponse) != 2 {
		t.Errorf("Expected 2 disputes needing response, got %d", len(needsResponse))
	}

	underReview := store.GetByStatus(DisputeStatusUnderReview)
	if len(underReview) != 1 {
		t.Errorf("Expected 1 dispute under review, got %d", len(underReview))
	}

	won := store.GetByStatus(DisputeStatusWon)
	if len(won) != 1 {
		t.Errorf("Expected 1 won dispute, got %d", len(won))
	}
}

func TestInMemoryDisputeStore_UpdateWithAuditTrail(t *testing.T) {
	store := NewInMemoryDisputeStore()

	dispute := Dispute{
		ID:     "dp_audit",
		Status: DisputeStatusNeedsResponse,
	}

	record := NewDisputeRecord(dispute, GatewayStripe, "test:create")
	_ = store.Save(record)

	// Retrieve and update
	record, _ = store.Get("dp_audit")

	// Mark evidence submitted
	record.MarkEvidenceSubmitted("user:test@example.com", "product_description")
	_ = store.Save(record)

	// Verify audit trail
	record, _ = store.Get("dp_audit")
	if len(record.AuditLog) != 2 {
		t.Errorf("Expected 2 audit entries, got %d", len(record.AuditLog))
	}
	if record.AuditLog[1].Action != "evidence_submitted" {
		t.Errorf("Expected action 'evidence_submitted', got '%s'", record.AuditLog[1].Action)
	}
	if !record.EvidenceSubmitted {
		t.Error("Expected EvidenceSubmitted to be true")
	}
	if record.EvidenceSubmittedAt == nil {
		t.Error("Expected EvidenceSubmittedAt to be set")
	}
}

func TestInMemoryDisputeStore_StatusTransition(t *testing.T) {
	store := NewInMemoryDisputeStore()

	dispute := Dispute{
		ID:     "dp_transition",
		Status: DisputeStatusNeedsResponse,
	}

	record := NewDisputeRecord(dispute, GatewayStripe, "test:create")
	_ = store.Save(record)

	// Transition to under review
	record, _ = store.Get("dp_transition")
	record.UpdateStatus(DisputeStatusUnderReview, "webhook:evidence_submitted", nil)
	_ = store.Save(record)

	// Verify status changed in index
	needsResponse := store.GetByStatus(DisputeStatusNeedsResponse)
	if len(needsResponse) != 0 {
		t.Errorf("Expected 0 disputes needing response after transition, got %d", len(needsResponse))
	}

	underReview := store.GetByStatus(DisputeStatusUnderReview)
	if len(underReview) != 1 {
		t.Errorf("Expected 1 dispute under review, got %d", len(underReview))
	}
}

func TestInMemoryDisputeStore_AcceptDispute(t *testing.T) {
	store := NewInMemoryDisputeStore()

	dispute := Dispute{
		ID:     "dp_accept",
		Status: DisputeStatusNeedsResponse,
	}

	record := NewDisputeRecord(dispute, GatewayStripe, "test:create")
	_ = store.Save(record)

	// Accept the dispute
	record, _ = store.Get("dp_accept")
	record.MarkAccepted("api:accept", "merchant_decision")
	_ = store.Save(record)

	// Verify
	record, _ = store.Get("dp_accept")
	if record.Status != DisputeStatusAccepted {
		t.Errorf("Expected status %s, got %s", DisputeStatusAccepted, record.Status)
	}
	if !record.Accepted {
		t.Error("Expected Accepted to be true")
	}
	if record.AcceptedAt == nil {
		t.Error("Expected AcceptedAt to be set")
	}
	if record.ClosedAt == nil {
		t.Error("Expected ClosedAt to be set")
	}
	if len(record.AuditLog) != 2 {
		t.Errorf("Expected 2 audit entries, got %d", len(record.AuditLog))
	}
}

func TestInMemoryDisputeStore_Delete(t *testing.T) {
	store := NewInMemoryDisputeStore()

	dispute := Dispute{
		ID:              "dp_delete",
		PaymentIntentID: "pi_delete",
		Status:          DisputeStatusNeedsResponse,
	}

	record := NewDisputeRecord(dispute, GatewayStripe, "test:create")
	record.CustomerID = "cust_delete"
	_ = store.Save(record)

	// Verify exists
	if store.Count() != 1 {
		t.Errorf("Expected count 1, got %d", store.Count())
	}

	// Delete
	err := store.Delete("dp_delete")
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify deleted
	if store.Count() != 0 {
		t.Errorf("Expected count 0 after delete, got %d", store.Count())
	}

	_, found := store.Get("dp_delete")
	if found {
		t.Error("Expected dispute to be deleted")
	}

	// Verify indices cleaned up
	if len(store.GetByPaymentIntent("pi_delete")) != 0 {
		t.Error("Expected payment intent index to be cleaned up")
	}
	if len(store.GetByStatus(DisputeStatusNeedsResponse)) != 0 {
		t.Error("Expected status index to be cleaned up")
	}
}

func TestInMemoryDisputeStore_ListWithFilters(t *testing.T) {
	store := NewInMemoryDisputeStore()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	disputes := []struct {
		dispute    Dispute
		customerID string
		createdAt  time.Time
	}{
		{Dispute{ID: "dp_1", Status: DisputeStatusNeedsResponse, Gateway: GatewayStripe}, "cust_1", now},
		{Dispute{ID: "dp_2", Status: DisputeStatusWon, Gateway: GatewayStripe}, "cust_1", yesterday},
		{Dispute{ID: "dp_3", Status: DisputeStatusNeedsResponse, Gateway: GatewayAdyen}, "cust_2", now},
	}

	for _, d := range disputes {
		record := NewDisputeRecord(d.dispute, d.dispute.Gateway, "test")
		record.CustomerID = d.customerID
		record.CreatedAt = d.createdAt
		_ = store.Save(record)
	}

	// Filter by gateway
	gatewayFilter := GatewayStripe
	opts := DisputeListOptions{Gateway: &gatewayFilter}
	results := store.List(opts)
	if len(results) != 2 {
		t.Errorf("Expected 2 Stripe disputes, got %d", len(results))
	}

	// Filter by status
	statusFilter := DisputeStatusNeedsResponse
	opts = DisputeListOptions{Status: &statusFilter}
	results = store.List(opts)
	if len(results) != 2 {
		t.Errorf("Expected 2 disputes needing response, got %d", len(results))
	}

	// Filter by customer
	opts = DisputeListOptions{CustomerID: "cust_1"}
	results = store.List(opts)
	if len(results) != 2 {
		t.Errorf("Expected 2 disputes for cust_1, got %d", len(results))
	}

	// Filter with limit
	opts = DisputeListOptions{Limit: 1}
	results = store.List(opts)
	if len(results) != 1 {
		t.Errorf("Expected 1 dispute with limit, got %d", len(results))
	}
}

func TestInMemoryDisputeStore_GetPendingResponse(t *testing.T) {
	store := NewInMemoryDisputeStore()

	now := time.Now()
	futureDeadline := now.Add(7 * 24 * time.Hour)
	pastDeadline := now.Add(-24 * time.Hour)

	disputes := []Dispute{
		{ID: "dp_pending1", Status: DisputeStatusNeedsResponse, EvidenceDueBy: futureDeadline},
		{ID: "dp_pending2", Status: DisputeStatusNeedsResponse, EvidenceDueBy: futureDeadline},
		{ID: "dp_expired", Status: DisputeStatusNeedsResponse, EvidenceDueBy: pastDeadline},
		{ID: "dp_closed", Status: DisputeStatusWon, EvidenceDueBy: futureDeadline},
	}

	for _, d := range disputes {
		_ = store.Save(NewDisputeRecord(d, GatewayStripe, "test"))
	}

	pending := store.GetPendingResponse()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending disputes, got %d", len(pending))
	}
}

// ============================================================================
// DisputeRecord Tests
// ============================================================================

func TestDisputeRecord_NewDisputeRecord(t *testing.T) {
	dispute := Dispute{
		ID:              "dp_new",
		PaymentIntentID: "pi_123",
		Amount:          NewAmount(5000, CurrencyEUR),
		Status:          DisputeStatusOpen,
		Reason:          DisputeReasonDuplicate,
	}

	record := NewDisputeRecord(dispute, GatewayAdyen, "system:webhook")

	if record.ID != dispute.ID {
		t.Errorf("Expected ID %s, got %s", dispute.ID, record.ID)
	}
	if record.Gateway != GatewayAdyen {
		t.Errorf("Expected gateway %s, got %s", GatewayAdyen, record.Gateway)
	}
	if record.GatewayDisputeID != dispute.ID {
		t.Errorf("Expected gateway dispute ID %s, got %s", dispute.ID, record.GatewayDisputeID)
	}
	if len(record.AuditLog) != 1 {
		t.Fatalf("Expected 1 audit entry, got %d", len(record.AuditLog))
	}

	entry := record.AuditLog[0]
	if entry.Action != "created" {
		t.Errorf("Expected action 'created', got '%s'", entry.Action)
	}
	if entry.Actor != "system:webhook" {
		t.Errorf("Expected actor 'system:webhook', got '%s'", entry.Actor)
	}
	if entry.NewStatus != DisputeStatusOpen {
		t.Errorf("Expected new status %s, got %s", DisputeStatusOpen, entry.NewStatus)
	}
}

func TestDisputeRecord_UpdateStatus(t *testing.T) {
	dispute := Dispute{ID: "dp_status", Status: DisputeStatusNeedsResponse}
	record := NewDisputeRecord(dispute, GatewayStripe, "test")

	record.UpdateStatus(DisputeStatusUnderReview, "user:admin", map[string]string{
		"reason": "evidence_submitted",
	})

	if record.Status != DisputeStatusUnderReview {
		t.Errorf("Expected status %s, got %s", DisputeStatusUnderReview, record.Status)
	}
	if len(record.AuditLog) != 2 {
		t.Fatalf("Expected 2 audit entries, got %d", len(record.AuditLog))
	}

	entry := record.AuditLog[1]
	if entry.PreviousStatus != DisputeStatusNeedsResponse {
		t.Errorf("Expected previous status %s, got %s", DisputeStatusNeedsResponse, entry.PreviousStatus)
	}
	if entry.Details["reason"] != "evidence_submitted" {
		t.Errorf("Expected reason 'evidence_submitted', got '%s'", entry.Details["reason"])
	}
}

func TestDisputeStatus_IsFinal(t *testing.T) {
	tests := []struct {
		status   DisputeStatus
		expected bool
	}{
		{DisputeStatusOpen, false},
		{DisputeStatusNeedsResponse, false},
		{DisputeStatusUnderReview, false},
		{DisputeStatusWon, true},
		{DisputeStatusLost, true},
		{DisputeStatusAccepted, true},
		{DisputeStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsFinal() != tt.expected {
				t.Errorf("Expected IsFinal()=%v for status %s", tt.expected, tt.status)
			}
		})
	}
}

