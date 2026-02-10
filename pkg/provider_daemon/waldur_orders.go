package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WaldurOrderRecord tracks a chain order mapped into Waldur.
type WaldurOrderRecord struct {
	OrderID         string            `json:"order_id"`
	CustomerAddress string            `json:"customer_address,omitempty"`
	OfferingID      string            `json:"offering_id,omitempty"`
	ProviderAddress string            `json:"provider_address,omitempty"`
	OfferingUUID    string            `json:"offering_uuid,omitempty"`
	ProjectUUID     string            `json:"project_uuid,omitempty"`
	WaldurOrderUUID string            `json:"waldur_order_uuid,omitempty"`
	WaldurResource  string            `json:"waldur_resource_uuid,omitempty"`
	WaldurState     string            `json:"waldur_state,omitempty"`
	ChainState      string            `json:"chain_state,omitempty"`
	Attributes      map[string]string `json:"attributes,omitempty"`
	LastError       string            `json:"last_error,omitempty"`
	RetryCount      int               `json:"retry_count,omitempty"`
	LastAttemptAt   *time.Time        `json:"last_attempt_at,omitempty"`
	NextAttemptAt   *time.Time        `json:"next_attempt_at,omitempty"`
	DeadLettered    bool              `json:"dead_lettered,omitempty"`
	DeadLetteredAt  *time.Time        `json:"dead_lettered_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// WaldurOrderDeadLetter captures a dead-lettered order routing event.
type WaldurOrderDeadLetter struct {
	OrderID       string            `json:"order_id"`
	Reason        string            `json:"reason"`
	Error         string            `json:"error,omitempty"`
	RetryCount    int               `json:"retry_count,omitempty"`
	LastAttemptAt *time.Time        `json:"last_attempt_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// WaldurOrderState holds all order routing state.
type WaldurOrderState struct {
	Orders          map[string]*WaldurOrderRecord `json:"orders"`
	DeadLetterQueue []*WaldurOrderDeadLetter      `json:"dead_letter_queue,omitempty"`
	UpdatedAt       time.Time                     `json:"updated_at"`
}

// WaldurOrderStore persists order routing state to disk.
type WaldurOrderStore struct {
	path string
	mu   sync.Mutex
}

// NewWaldurOrderStore creates a new order store for the given file path.
func NewWaldurOrderStore(path string) (*WaldurOrderStore, error) {
	if err := validateStatePath(path); err != nil {
		return nil, fmt.Errorf("invalid order state path: %w", err)
	}
	return &WaldurOrderStore{path: filepath.Clean(path)}, nil
}

// Load reads order state from disk or returns an empty state.
func (s *WaldurOrderStore) Load() (*WaldurOrderState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path) // #nosec G304 -- path validated in constructor
	if err != nil {
		if os.IsNotExist(err) {
			return &WaldurOrderState{
				Orders:    map[string]*WaldurOrderRecord{},
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, fmt.Errorf("read order state: %w", err)
	}

	var state WaldurOrderState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode order state: %w", err)
	}
	if state.Orders == nil {
		state.Orders = map[string]*WaldurOrderRecord{}
	}
	return &state, nil
}

// Save writes order state to disk atomically.
func (s *WaldurOrderStore) Save(state *WaldurOrderState) error {
	if state == nil {
		return fmt.Errorf("order state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode order state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create order state dir: %w", err)
	}

	tmp := s.path + ".tmp"
	// #nosec G304 -- path validated in constructor and cleaned
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write order state tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace order state: %w", err)
	}
	return nil
}

// Get returns the record for an order ID.
func (s *WaldurOrderState) Get(orderID string) *WaldurOrderRecord {
	if s == nil || orderID == "" {
		return nil
	}
	return s.Orders[orderID]
}

// Upsert adds or updates an order record.
func (s *WaldurOrderState) Upsert(record *WaldurOrderRecord) {
	if s == nil || record == nil || record.OrderID == "" {
		return
	}
	if s.Orders == nil {
		s.Orders = map[string]*WaldurOrderRecord{}
	}
	now := time.Now().UTC()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	s.Orders[record.OrderID] = record
}

// FindByWaldurUUID finds an order record by Waldur order UUID.
func (s *WaldurOrderState) FindByWaldurUUID(uuid string) *WaldurOrderRecord {
	if s == nil || uuid == "" {
		return nil
	}
	for _, record := range s.Orders {
		if record != nil && record.WaldurOrderUUID == uuid {
			return record
		}
	}
	return nil
}

// MarkFailed updates record with failure metadata.
func (s *WaldurOrderState) MarkFailed(orderID, errMsg string) *WaldurOrderRecord {
	record := s.Get(orderID)
	if record == nil {
		record = &WaldurOrderRecord{OrderID: orderID}
	}
	now := time.Now().UTC()
	record.LastError = errMsg
	record.RetryCount++
	record.LastAttemptAt = &now
	record.UpdatedAt = now
	s.Upsert(record)
	return record
}

// MarkSynced updates record with successful mapping metadata.
func (s *WaldurOrderState) MarkSynced(orderID, waldurUUID, resourceUUID, state string) *WaldurOrderRecord {
	record := s.Get(orderID)
	if record == nil {
		record = &WaldurOrderRecord{OrderID: orderID}
	}
	record.WaldurOrderUUID = waldurUUID
	record.WaldurResource = resourceUUID
	record.WaldurState = state
	record.LastError = ""
	record.UpdatedAt = time.Now().UTC()
	s.Upsert(record)
	return record
}

// DeadLetter moves an order record into the dead letter queue.
func (s *WaldurOrderState) DeadLetter(record *WaldurOrderRecord, reason string) {
	if s == nil || record == nil {
		return
	}
	now := time.Now().UTC()
	record.DeadLettered = true
	record.DeadLetteredAt = &now
	record.UpdatedAt = now
	s.Upsert(record)

	entry := &WaldurOrderDeadLetter{
		OrderID:       record.OrderID,
		Reason:        reason,
		Error:         record.LastError,
		RetryCount:    record.RetryCount,
		LastAttemptAt: record.LastAttemptAt,
		CreatedAt:     now,
		Attributes:    record.Attributes,
	}
	s.DeadLetterQueue = append(s.DeadLetterQueue, entry)
}
