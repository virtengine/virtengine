package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventCheckpointState tracks the last processed event sequence.
type EventCheckpointState struct {
	SubscriberID string    `json:"subscriber_id"`
	LastSequence uint64    `json:"last_sequence"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// EventCheckpointStore persists marketplace event checkpoints.
type EventCheckpointStore struct {
	path string
	mu   sync.Mutex
}

// NewEventCheckpointStore creates a checkpoint store using a file path.
func NewEventCheckpointStore(path string) *EventCheckpointStore {
	return &EventCheckpointStore{path: path}
}

// Load reads the checkpoint from disk. If no file exists, returns a zero state.
func (s *EventCheckpointStore) Load(subscriberID string) (*EventCheckpointState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &EventCheckpointState{
				SubscriberID: subscriberID,
				LastSequence: 0,
				UpdatedAt:    time.Now().UTC(),
			}, nil
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var state EventCheckpointState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode checkpoint: %w", err)
	}

	if state.SubscriberID == "" {
		state.SubscriberID = subscriberID
	}

	return &state, nil
}

// Save writes the checkpoint to disk atomically.
func (s *EventCheckpointStore) Save(state *EventCheckpointState) error {
	if state == nil {
		return fmt.Errorf("checkpoint state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode checkpoint: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create checkpoint dir: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write checkpoint tmp: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace checkpoint: %w", err)
	}

	return nil
}
