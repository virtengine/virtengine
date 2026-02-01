package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WaldurAllocationMapping tracks Waldur order/resource IDs for an allocation.
type WaldurAllocationMapping struct {
	AllocationID string    `json:"allocation_id"`
	OrderUUID    string    `json:"order_uuid,omitempty"`
	ResourceUUID string    `json:"resource_uuid,omitempty"`
	OfferingUUID string    `json:"offering_uuid,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WaldurBridgeState holds allocation mappings.
type WaldurBridgeState struct {
	Mappings map[string]*WaldurAllocationMapping `json:"mappings"`
}

// WaldurBridgeStateStore persists Waldur bridge state to disk.
type WaldurBridgeStateStore struct {
	path string
	mu   sync.Mutex
}

// NewWaldurBridgeStateStore creates a new store.
func NewWaldurBridgeStateStore(path string) *WaldurBridgeStateStore {
	return &WaldurBridgeStateStore{path: path}
}

// Load reads state from disk or returns a new state.
func (s *WaldurBridgeStateStore) Load() (*WaldurBridgeState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WaldurBridgeState{Mappings: map[string]*WaldurAllocationMapping{}}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	var state WaldurBridgeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	if state.Mappings == nil {
		state.Mappings = map[string]*WaldurAllocationMapping{}
	}
	return &state, nil
}

// Save writes state to disk atomically.
func (s *WaldurBridgeStateStore) Save(state *WaldurBridgeState) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
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

