// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Dispute Record Types
// ============================================================================

// DisputeRecord represents a persisted dispute with full metadata
type DisputeRecord struct {
	// Dispute contains the core dispute data from the gateway
	Dispute

	// Gateway-specific ID (may differ from our internal ID)
	GatewayDisputeID string `json:"gateway_dispute_id"`

	// CustomerID is the customer who made the original payment
	CustomerID string `json:"customer_id,omitempty"`

	// ChargeID is the underlying charge/transaction ID
	ChargeID string `json:"charge_id,omitempty"`

	// EvidenceSubmitted indicates if evidence has been submitted
	EvidenceSubmitted bool `json:"evidence_submitted"`

	// EvidenceSubmittedAt is when evidence was submitted
	EvidenceSubmittedAt *time.Time `json:"evidence_submitted_at,omitempty"`

	// Accepted indicates if the dispute was accepted (conceded)
	Accepted bool `json:"accepted"`

	// AcceptedAt is when the dispute was accepted
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`

	// ClosedAt is when the dispute was closed (won/lost/accepted)
	ClosedAt *time.Time `json:"closed_at,omitempty"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// AuditLog contains the audit trail for this dispute
	AuditLog []DisputeAuditEntry `json:"audit_log"`
}

// DisputeAuditEntry represents a single audit log entry for dispute actions
type DisputeAuditEntry struct {
	// Timestamp of the action
	Timestamp time.Time `json:"timestamp"`

	// Action performed (e.g., "created", "evidence_submitted", "accepted", "status_updated")
	Action string `json:"action"`

	// Actor is who performed the action (e.g., "webhook", "user:addr123", "system")
	Actor string `json:"actor"`

	// PreviousStatus before the action
	PreviousStatus DisputeStatus `json:"previous_status,omitempty"`

	// NewStatus after the action
	NewStatus DisputeStatus `json:"new_status,omitempty"`

	// Details contains additional context for the action
	Details map[string]string `json:"details,omitempty"`
}

// NewDisputeRecord creates a new dispute record from a gateway dispute
func NewDisputeRecord(dispute Dispute, gateway GatewayType, actor string) *DisputeRecord {
	now := time.Now()
	record := &DisputeRecord{
		Dispute:          dispute,
		GatewayDisputeID: dispute.ID,
		UpdatedAt:        now,
		AuditLog:         make([]DisputeAuditEntry, 0, 10),
	}
	record.Dispute.Gateway = gateway

	record.AddAuditEntry(DisputeAuditEntry{
		Timestamp: now,
		Action:    "created",
		Actor:     actor,
		NewStatus: dispute.Status,
		Details: map[string]string{
			"reason":        string(dispute.Reason),
			"payment_id":    dispute.PaymentIntentID,
			"amount":        fmt.Sprintf("%d", dispute.Amount.Value),
			"currency":      string(dispute.Amount.Currency),
			"gateway":       string(gateway),
		},
	})

	return record
}

// AddAuditEntry adds an audit entry to the dispute record
func (r *DisputeRecord) AddAuditEntry(entry DisputeAuditEntry) {
	r.AuditLog = append(r.AuditLog, entry)
	r.UpdatedAt = entry.Timestamp
}

// UpdateStatus updates the dispute status with audit logging
func (r *DisputeRecord) UpdateStatus(newStatus DisputeStatus, actor string, details map[string]string) {
	now := time.Now()
	previousStatus := r.Status

	r.Status = newStatus
	r.UpdatedAt = now

	if newStatus == DisputeStatusWon || newStatus == DisputeStatusLost ||
		newStatus == DisputeStatusAccepted || newStatus == DisputeStatusExpired {
		r.ClosedAt = &now
	}

	r.AddAuditEntry(DisputeAuditEntry{
		Timestamp:      now,
		Action:         "status_updated",
		Actor:          actor,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Details:        details,
	})
}

// MarkEvidenceSubmitted marks evidence as submitted with audit logging
func (r *DisputeRecord) MarkEvidenceSubmitted(actor string, evidenceType string) {
	now := time.Now()
	r.EvidenceSubmitted = true
	r.EvidenceSubmittedAt = &now
	r.UpdatedAt = now

	if r.Status == DisputeStatusNeedsResponse {
		r.Status = DisputeStatusUnderReview
	}

	r.AddAuditEntry(DisputeAuditEntry{
		Timestamp: now,
		Action:    "evidence_submitted",
		Actor:     actor,
		NewStatus: r.Status,
		Details: map[string]string{
			"evidence_type": evidenceType,
		},
	})
}

// MarkAccepted marks the dispute as accepted (conceded)
func (r *DisputeRecord) MarkAccepted(actor string, reason string) {
	now := time.Now()
	previousStatus := r.Status

	r.Accepted = true
	r.AcceptedAt = &now
	r.Status = DisputeStatusAccepted
	r.ClosedAt = &now
	r.UpdatedAt = now

	r.AddAuditEntry(DisputeAuditEntry{
		Timestamp:      now,
		Action:         "accepted",
		Actor:          actor,
		PreviousStatus: previousStatus,
		NewStatus:      DisputeStatusAccepted,
		Details: map[string]string{
			"reason": reason,
		},
	})
}

// ============================================================================
// Dispute Store Interface
// ============================================================================

// DisputeStore defines the interface for persisting disputes
type DisputeStore interface {
	// Save persists a dispute record
	Save(record *DisputeRecord) error

	// Get retrieves a dispute by ID
	Get(disputeID string) (*DisputeRecord, bool)

	// GetByPaymentIntent retrieves disputes for a payment intent
	GetByPaymentIntent(paymentIntentID string) []*DisputeRecord

	// GetByCustomer retrieves disputes for a customer
	GetByCustomer(customerID string) []*DisputeRecord

	// GetByStatus retrieves disputes by status
	GetByStatus(status DisputeStatus) []*DisputeRecord

	// GetPendingResponse retrieves disputes needing response before deadline
	GetPendingResponse() []*DisputeRecord

	// List retrieves all disputes with optional filters
	List(opts DisputeListOptions) []*DisputeRecord

	// Delete removes a dispute record
	Delete(disputeID string) error

	// Count returns the total number of disputes
	Count() int
}

// DisputeListOptions defines filtering options for listing disputes
type DisputeListOptions struct {
	// Gateway filters by payment gateway
	Gateway *GatewayType `json:"gateway,omitempty"`

	// Status filters by dispute status
	Status *DisputeStatus `json:"status,omitempty"`

	// CustomerID filters by customer
	CustomerID string `json:"customer_id,omitempty"`

	// CreatedAfter filters disputes created after this time
	CreatedAfter *time.Time `json:"created_after,omitempty"`

	// CreatedBefore filters disputes created before this time
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// Limit limits the number of results
	Limit int `json:"limit,omitempty"`

	// Offset for pagination
	Offset int `json:"offset,omitempty"`
}

// ============================================================================
// In-Memory Dispute Store Implementation
// ============================================================================

// InMemoryDisputeStore is a thread-safe in-memory implementation of DisputeStore
// Suitable for testing and development. Production should use a persistent store.
type InMemoryDisputeStore struct {
	mu       sync.RWMutex
	disputes map[string]*DisputeRecord

	// Secondary indices
	byPaymentIntent map[string][]string // paymentIntentID -> []disputeID
	byCustomer      map[string][]string // customerID -> []disputeID
	byStatus        map[DisputeStatus][]string
}

// NewInMemoryDisputeStore creates a new in-memory dispute store
func NewInMemoryDisputeStore() *InMemoryDisputeStore {
	return &InMemoryDisputeStore{
		disputes:        make(map[string]*DisputeRecord),
		byPaymentIntent: make(map[string][]string),
		byCustomer:      make(map[string][]string),
		byStatus:        make(map[DisputeStatus][]string),
	}
}

// Save persists a dispute record
func (s *InMemoryDisputeStore) Save(record *DisputeRecord) error {
	if record == nil {
		return fmt.Errorf("cannot save nil dispute record")
	}
	if record.ID == "" {
		return fmt.Errorf("dispute ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if this is an update (need to update indices)
	existing, exists := s.disputes[record.ID]
	if exists {
		// Remove from old status index if status changed
		if existing.Status != record.Status {
			s.removeFromStatusIndex(record.ID, existing.Status)
			s.addToStatusIndex(record.ID, record.Status)
		}
	} else {
		// New record - add to all indices
		s.addToPaymentIntentIndex(record.ID, record.PaymentIntentID)
		if record.CustomerID != "" {
			s.addToCustomerIndex(record.ID, record.CustomerID)
		}
		s.addToStatusIndex(record.ID, record.Status)
	}

	// Make a copy to prevent external mutation
	recordCopy := *record
	recordCopy.AuditLog = make([]DisputeAuditEntry, len(record.AuditLog))
	copy(recordCopy.AuditLog, record.AuditLog)

	s.disputes[record.ID] = &recordCopy
	return nil
}

// Get retrieves a dispute by ID
func (s *InMemoryDisputeStore) Get(disputeID string) (*DisputeRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, exists := s.disputes[disputeID]
	if !exists {
		return nil, false
	}

	// Return a copy
	recordCopy := *record
	recordCopy.AuditLog = make([]DisputeAuditEntry, len(record.AuditLog))
	copy(recordCopy.AuditLog, record.AuditLog)

	return &recordCopy, true
}

// GetByPaymentIntent retrieves disputes for a payment intent
func (s *InMemoryDisputeStore) GetByPaymentIntent(paymentIntentID string) []*DisputeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.byPaymentIntent[paymentIntentID]
	return s.getRecordsByIDs(ids)
}

// GetByCustomer retrieves disputes for a customer
func (s *InMemoryDisputeStore) GetByCustomer(customerID string) []*DisputeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.byCustomer[customerID]
	return s.getRecordsByIDs(ids)
}

// GetByStatus retrieves disputes by status
func (s *InMemoryDisputeStore) GetByStatus(status DisputeStatus) []*DisputeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.byStatus[status]
	return s.getRecordsByIDs(ids)
}

// GetPendingResponse retrieves disputes needing response before deadline
func (s *InMemoryDisputeStore) GetPendingResponse() []*DisputeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var pending []*DisputeRecord

	for _, record := range s.disputes {
		if record.Status == DisputeStatusNeedsResponse && record.EvidenceDueBy.After(now) {
			recordCopy := *record
			recordCopy.AuditLog = make([]DisputeAuditEntry, len(record.AuditLog))
			copy(recordCopy.AuditLog, record.AuditLog)
			pending = append(pending, &recordCopy)
		}
	}

	return pending
}

// List retrieves all disputes with optional filters
func (s *InMemoryDisputeStore) List(opts DisputeListOptions) []*DisputeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*DisputeRecord
	count := 0
	skipped := 0

	for _, record := range s.disputes {
		// Apply filters
		if opts.Gateway != nil && record.Gateway != *opts.Gateway {
			continue
		}
		if opts.Status != nil && record.Status != *opts.Status {
			continue
		}
		if opts.CustomerID != "" && record.CustomerID != opts.CustomerID {
			continue
		}
		if opts.CreatedAfter != nil && record.CreatedAt.Before(*opts.CreatedAfter) {
			continue
		}
		if opts.CreatedBefore != nil && record.CreatedAt.After(*opts.CreatedBefore) {
			continue
		}

		// Apply offset
		if opts.Offset > 0 && skipped < opts.Offset {
			skipped++
			continue
		}

		// Apply limit
		if opts.Limit > 0 && count >= opts.Limit {
			break
		}

		recordCopy := *record
		recordCopy.AuditLog = make([]DisputeAuditEntry, len(record.AuditLog))
		copy(recordCopy.AuditLog, record.AuditLog)
		results = append(results, &recordCopy)
		count++
	}

	return results
}

// Delete removes a dispute record
func (s *InMemoryDisputeStore) Delete(disputeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, exists := s.disputes[disputeID]
	if !exists {
		return nil // Idempotent
	}

	// Remove from indices
	s.removeFromPaymentIntentIndex(disputeID, record.PaymentIntentID)
	if record.CustomerID != "" {
		s.removeFromCustomerIndex(disputeID, record.CustomerID)
	}
	s.removeFromStatusIndex(disputeID, record.Status)

	delete(s.disputes, disputeID)
	return nil
}

// Count returns the total number of disputes
func (s *InMemoryDisputeStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.disputes)
}

// Helper methods for index management

func (s *InMemoryDisputeStore) getRecordsByIDs(ids []string) []*DisputeRecord {
	var records []*DisputeRecord
	for _, id := range ids {
		if record, exists := s.disputes[id]; exists {
			recordCopy := *record
			recordCopy.AuditLog = make([]DisputeAuditEntry, len(record.AuditLog))
			copy(recordCopy.AuditLog, record.AuditLog)
			records = append(records, &recordCopy)
		}
	}
	return records
}

func (s *InMemoryDisputeStore) addToPaymentIntentIndex(disputeID, paymentIntentID string) {
	if paymentIntentID == "" {
		return
	}
	s.byPaymentIntent[paymentIntentID] = append(s.byPaymentIntent[paymentIntentID], disputeID)
}

func (s *InMemoryDisputeStore) removeFromPaymentIntentIndex(disputeID, paymentIntentID string) {
	if paymentIntentID == "" {
		return
	}
	ids := s.byPaymentIntent[paymentIntentID]
	for i, id := range ids {
		if id == disputeID {
			s.byPaymentIntent[paymentIntentID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

func (s *InMemoryDisputeStore) addToCustomerIndex(disputeID, customerID string) {
	if customerID == "" {
		return
	}
	s.byCustomer[customerID] = append(s.byCustomer[customerID], disputeID)
}

func (s *InMemoryDisputeStore) removeFromCustomerIndex(disputeID, customerID string) {
	if customerID == "" {
		return
	}
	ids := s.byCustomer[customerID]
	for i, id := range ids {
		if id == disputeID {
			s.byCustomer[customerID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}

func (s *InMemoryDisputeStore) addToStatusIndex(disputeID string, status DisputeStatus) {
	s.byStatus[status] = append(s.byStatus[status], disputeID)
}

func (s *InMemoryDisputeStore) removeFromStatusIndex(disputeID string, status DisputeStatus) {
	ids := s.byStatus[status]
	for i, id := range ids {
		if id == disputeID {
			s.byStatus[status] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
}
