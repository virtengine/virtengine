package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const hpcNodeCheckpointVersion = 1

// HPCNodeCheckpointState persists node aggregator state.
type HPCNodeCheckpointState struct {
	Version   int                       `json:"version"`
	UpdatedAt time.Time                 `json:"updated_at"`
	Nodes     []HPCNodeCheckpointNode   `json:"nodes"`
	Pending   []HPCNodeCheckpointUpdate `json:"pending"`
}

// HPCNodeCheckpointNode stores node state.
type HPCNodeCheckpointNode struct {
	NodeID                string           `json:"node_id"`
	ClusterID             string           `json:"cluster_id"`
	PublicKey             string           `json:"public_key"`
	AgentPubkey           string           `json:"agent_pubkey,omitempty"`
	HardwareFingerprint   string           `json:"hardware_fingerprint,omitempty"`
	AgentVersion          string           `json:"agent_version,omitempty"`
	LastHeartbeat         time.Time        `json:"last_heartbeat"`
	LastSequence          uint64           `json:"last_sequence"`
	ConsecutiveMisses     int              `json:"consecutive_misses"`
	TotalHeartbeats       uint64           `json:"total_heartbeats"`
	Capacity              *HPCNodeCapacity `json:"capacity,omitempty"`
	Health                *HPCNodeHealth   `json:"health,omitempty"`
	Latency               *HPCNodeLatency  `json:"latency,omitempty"`
	Hardware              *HPCNodeHardware `json:"hardware,omitempty"`
	Topology              *HPCNodeTopology `json:"topology,omitempty"`
	Locality              *HPCNodeLocality `json:"locality,omitempty"`
	PendingChainUpdate    bool             `json:"pending_chain_update"`
	OnChainRegistered     bool             `json:"on_chain_registered"`
	LastSubmittedSequence uint64           `json:"last_submitted_sequence"`
	LastChainUpdate       time.Time        `json:"last_chain_update"`
	LastSubmitError       string           `json:"last_submit_error,omitempty"`
	RegisteredAt          time.Time        `json:"registered_at"`
}

// HPCNodeCheckpointUpdate stores pending chain updates.
type HPCNodeCheckpointUpdate struct {
	NodeID         string            `json:"node_id"`
	ClusterID      string            `json:"cluster_id"`
	UpdateType     nodeUpdateType    `json:"update_type"`
	Heartbeat      *HPCNodeHeartbeat `json:"heartbeat,omitempty"`
	Active         bool              `json:"active"`
	SequenceNumber uint64            `json:"sequence_number"`
	Attempts       int               `json:"attempts"`
	NextAttempt    time.Time         `json:"next_attempt"`
	EnqueuedAt     time.Time         `json:"enqueued_at"`
}

// HPCNodeCheckpointStore persists node checkpoints.
type HPCNodeCheckpointStore struct {
	path string
}

// NewHPCNodeCheckpointStore creates a checkpoint store.
func NewHPCNodeCheckpointStore(path string) (*HPCNodeCheckpointStore, error) {
	if err := validateStatePath(path); err != nil {
		return nil, fmt.Errorf("invalid checkpoint path: %w", err)
	}
	return &HPCNodeCheckpointStore{path: filepath.Clean(path)}, nil
}

// Load reads the checkpoint from disk.
func (s *HPCNodeCheckpointStore) Load() (*HPCNodeCheckpointState, error) {
	data, err := os.ReadFile(s.path) // #nosec G304 -- path validated in constructor
	if err != nil {
		if os.IsNotExist(err) {
			return &HPCNodeCheckpointState{
				Version:   hpcNodeCheckpointVersion,
				UpdatedAt: time.Now().UTC(),
				Nodes:     []HPCNodeCheckpointNode{},
				Pending:   []HPCNodeCheckpointUpdate{},
			}, nil
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var state HPCNodeCheckpointState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode checkpoint: %w", err)
	}
	if state.Version == 0 {
		state.Version = hpcNodeCheckpointVersion
	}
	return &state, nil
}

// Save writes the checkpoint to disk.
func (s *HPCNodeCheckpointStore) Save(state *HPCNodeCheckpointState) error {
	if state == nil {
		return fmt.Errorf("checkpoint state is nil")
	}

	state.Version = hpcNodeCheckpointVersion
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
