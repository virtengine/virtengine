package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OrderRoutingState tracks Waldur order mappings and retry metadata.
type OrderRoutingState struct {
	Records         map[string]*OrderRoutingRecord `json:"records"`
	DeadLetterQueue []*OrderDeadLetterItem         `json:"dead_letter_queue,omitempty"`
	Metrics         OrderRoutingMetrics            `json:"metrics"`
}

// OrderRoutingMetrics tracks high-level routing stats.
type OrderRoutingMetrics struct {
	Processed     int64  `json:"processed"`
	Skipped       int64  `json:"skipped"`
	Retried       int64  `json:"retried"`
	Failed        int64  `json:"failed"`
	DeadLettered  int64  `json:"dead_lettered"`
	Recovered     int64  `json:"recovered"`
	LastSequence  uint64 `json:"last_sequence"`
	LastProcessed string `json:"last_processed,omitempty"`
}

// OrderRoutingRecord tracks a single chain order mapped to Waldur.
type OrderRoutingRecord struct {
	OrderID            string    `json:"order_id"`
	OfferingID         string    `json:"offering_id"`
	ProviderAddress    string    `json:"provider_address"`
	CustomerAddress    string    `json:"customer_address"`
	WaldurOfferingUUID string    `json:"waldur_offering_uuid,omitempty"`
	WaldurOrderUUID    string    `json:"waldur_order_uuid,omitempty"`
	WaldurResourceUUID string    `json:"waldur_resource_uuid,omitempty"`
	LastSequence       uint64    `json:"last_sequence"`
	LastState          string    `json:"last_state,omitempty"`
	RetryCount         int       `json:"retry_count"`
	LastError          string    `json:"last_error,omitempty"`
	DeadLettered       bool      `json:"dead_lettered"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	LastAttemptAt      time.Time `json:"last_attempt_at"`
}

// OrderDeadLetterItem represents a permanently failed order routing attempt.
type OrderDeadLetterItem struct {
	OrderID         string    `json:"order_id"`
	OfferingID      string    `json:"offering_id"`
	ProviderAddress string    `json:"provider_address"`
	Reason          string    `json:"reason"`
	Attempts        int       `json:"attempts"`
	DeadLetteredAt  time.Time `json:"dead_lettered_at"`
	LastError       string    `json:"last_error"`
}

// OrderRoutingStateStore persists order routing state to disk.
type OrderRoutingStateStore struct {
	path string
	mu   sync.Mutex
}

// NewOrderRoutingStateStore creates a new state store.
func NewOrderRoutingStateStore(path string) *OrderRoutingStateStore {
	return &OrderRoutingStateStore{path: path}
}

// Load reads routing state from disk or returns a new state.
func (s *OrderRoutingStateStore) Load() (*OrderRoutingState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &OrderRoutingState{
				Records:         map[string]*OrderRoutingRecord{},
				DeadLetterQueue: make([]*OrderDeadLetterItem, 0),
			}, nil
		}
		return nil, fmt.Errorf("read order routing state: %w", err)
	}

	var state OrderRoutingState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode order routing state: %w", err)
	}
	if state.Records == nil {
		state.Records = map[string]*OrderRoutingRecord{}
	}
	if state.DeadLetterQueue == nil {
		state.DeadLetterQueue = make([]*OrderDeadLetterItem, 0)
	}
	return &state, nil
}

// Save writes routing state to disk atomically.
func (s *OrderRoutingStateStore) Save(state *OrderRoutingState) error {
	if state == nil {
		return fmt.Errorf("order routing state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode order routing state: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write state tmp: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	return nil
}
