package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProvisioningTask tracks provisioning progress for an allocation.
type ProvisioningTask struct {
	AllocationID       string            `json:"allocation_id"`
	OfferingID         string            `json:"offering_id,omitempty"`
	ProviderAddress    string            `json:"provider_address,omitempty"`
	ServiceType        string            `json:"service_type,omitempty"`
	EncryptedConfigRef string            `json:"encrypted_config_ref,omitempty"`
	Specifications     map[string]string `json:"specifications,omitempty"`
	State              string            `json:"state"`
	Phase              string            `json:"phase,omitempty"`
	Progress           uint8             `json:"progress,omitempty"`
	ResourceID         string            `json:"resource_id,omitempty"`
	Endpoints          map[string]string `json:"endpoints,omitempty"`
	Attempts           int               `json:"attempts"`
	NextAttemptAt      *time.Time        `json:"next_attempt_at,omitempty"`
	LastError          string            `json:"last_error,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	CompletedAt        *time.Time        `json:"completed_at,omitempty"`
}

// ProvisioningState stores provisioning task state for a provider.
type ProvisioningState struct {
	ProviderAddress string                       `json:"provider_address"`
	Tasks           map[string]*ProvisioningTask `json:"tasks"`
	UpdatedAt       time.Time                    `json:"updated_at"`
}

// ProvisioningStateStore persists provisioning state to disk.
type ProvisioningStateStore struct {
	path string
}

// NewProvisioningStateStore creates a new state store.
func NewProvisioningStateStore(path string) *ProvisioningStateStore {
	// Validate path if provided
	if path != "" {
		if err := validateStatePath(path); err != nil {
			// Return store with empty path if validation fails
			// Caller will handle empty path case gracefully
			return &ProvisioningStateStore{path: ""}
		}
		path = filepath.Clean(path)
	}
	return &ProvisioningStateStore{path: path}
}

// Load loads state from disk.
func (s *ProvisioningStateStore) Load() (*ProvisioningState, error) {
	if s == nil || s.path == "" {
		return &ProvisioningState{Tasks: map[string]*ProvisioningTask{}}, nil
	}

	// #nosec G304 -- path validated and cleaned in constructor
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProvisioningState{Tasks: map[string]*ProvisioningTask{}}, nil
		}
		return nil, err
	}

	var state ProvisioningState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode provisioning state: %w", err)
	}
	if state.Tasks == nil {
		state.Tasks = map[string]*ProvisioningTask{}
	}
	return &state, nil
}

// Save persists state to disk.
func (s *ProvisioningStateStore) Save(state *ProvisioningState) error {
	if s == nil || s.path == "" || state == nil {
		return nil
	}
	state.UpdatedAt = time.Now().UTC()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	// #nosec G304 -- path validated and cleaned in constructor
	return os.WriteFile(s.path, data, 0o600)
}
