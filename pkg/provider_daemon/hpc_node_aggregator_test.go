package provider_daemon

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNodeAggregatorCheckpointRestore(t *testing.T) {
	const clusterID = "cluster-1"

	dir := t.TempDir()
	checkpointPath := filepath.Join(dir, "node-checkpoint.json")

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	pubkey := base64.StdEncoding.EncodeToString(pub)

	cfg := DefaultHPCNodeAggregatorConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "virtengine1provider"
	cfg.ClusterID = clusterID
	cfg.CheckpointFile = checkpointPath
	cfg.ChainSubmitEnabled = false

	agg, err := NewHPCNodeAggregator(cfg, nil)
	if err != nil {
		t.Fatalf("failed to create aggregator: %v", err)
	}

	if err := agg.registerNode("node-1", cfg.ClusterID, cfg.ProviderAddress, pubkey); err != nil {
		t.Fatalf("failed to register node: %v", err)
	}

	hb := &HPCNodeHeartbeat{
		NodeID:         "node-1",
		ClusterID:      cfg.ClusterID,
		SequenceNumber: 1,
		Timestamp:      time.Now().UTC(),
		AgentVersion:   "1.0.0",
		Capacity: HPCNodeCapacity{
			CPUCoresTotal: 8,
			MemoryGBTotal: 64,
		},
		Health: HPCNodeHealth{Status: healthStatusHealthy},
	}
	auth := signHeartbeat(t, hb, priv)

	resp := agg.processHeartbeat(hb, auth)
	if !resp.Accepted {
		t.Fatalf("expected heartbeat accepted, got %+v", resp.Errors)
	}

	if err := agg.persistCheckpoint(); err != nil {
		t.Fatalf("failed to persist checkpoint: %v", err)
	}

	restoredCfg := cfg
	restoredCfg.AllowedNodePubkeys = make(map[string]bool)
	restoredAgg, err := NewHPCNodeAggregator(restoredCfg, nil)
	if err != nil {
		t.Fatalf("failed to create restored aggregator: %v", err)
	}

	if err := restoredAgg.loadCheckpoint(); err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if restoredAgg.GetNodeCount() != 1 {
		t.Fatalf("expected 1 node after restore, got %d", restoredAgg.GetNodeCount())
	}

	stats, ok := restoredAgg.GetNodeStats("node-1")
	if !ok {
		t.Fatalf("expected restored node stats")
	}
	if stats["last_sequence"].(uint64) != 1 {
		t.Fatalf("expected last_sequence 1, got %v", stats["last_sequence"])
	}

	restoredAgg.pendingMu.Lock()
	pendingCount := len(restoredAgg.pending)
	restoredAgg.pendingMu.Unlock()
	if pendingCount == 0 {
		t.Fatalf("expected pending updates to be restored")
	}

	if _, err := os.Stat(checkpointPath); err != nil {
		t.Fatalf("expected checkpoint file on disk: %v", err)
	}
}

func signHeartbeat(t *testing.T, hb *HPCNodeHeartbeat, priv ed25519.PrivateKey) *HPCHeartbeatAuth {
	t.Helper()
	data, err := json.Marshal(hb)
	if err != nil {
		t.Fatalf("marshal heartbeat: %v", err)
	}
	sig := ed25519.Sign(priv, data)
	return &HPCHeartbeatAuth{
		Signature: base64.StdEncoding.EncodeToString(sig),
		Nonce:     "test",
		Timestamp: time.Now().Unix(),
	}
}
