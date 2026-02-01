package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// validateStatePath checks for path traversal attacks in state file paths
func validateStatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	cleanPath := filepath.Clean(path)
	// Check for traversal sequences
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}
	// Check for null bytes
	if strings.ContainsRune(path, '\x00') {
		return fmt.Errorf("null byte in path: %s", path)
	}
	return nil
}

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
func NewEventCheckpointStore(path string) (*EventCheckpointStore, error) {
	if err := validateStatePath(path); err != nil {
		return nil, fmt.Errorf("invalid checkpoint path: %w", err)
	}
	return &EventCheckpointStore{path: filepath.Clean(path)}, nil
}

// Load reads the checkpoint from disk. If no file exists, returns a zero state.
func (s *EventCheckpointStore) Load(subscriberID string) (*EventCheckpointState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Path already validated in constructor, use cleaned path
	data, err := os.ReadFile(s.path) // #nosec G304 -- path validated in constructor
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

