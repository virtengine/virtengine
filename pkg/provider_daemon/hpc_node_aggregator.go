// Package provider_daemon implements the VirtEngine provider daemon services.
//
// VE-500: HPC node aggregator for heartbeat ingestion and on-chain metadata updates.
package provider_daemon

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HPCNodeAggregatorConfig contains configuration for the HPC node aggregator
type HPCNodeAggregatorConfig struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string

	// ProviderID is the provider's public key
	ProviderID string

	// ListenAddr is the HTTP listen address for node heartbeats
	ListenAddr string

	// BatchSubmitInterval is how often to batch submit node metadata updates
	BatchSubmitInterval time.Duration

	// MaxBatchSize is the maximum nodes to update in one batch
	MaxBatchSize int

	// HeartbeatTimeout is when to consider a node stale
	HeartbeatTimeout time.Duration

	// ChainGRPC is the chain gRPC endpoint for on-chain submissions
	ChainGRPC string

	// AllowedNodePubkeys is the allowlist of node public keys (base64)
	AllowedNodePubkeys map[string]bool
}

// DefaultHPCNodeAggregatorConfig returns the default configuration
func DefaultHPCNodeAggregatorConfig() HPCNodeAggregatorConfig {
	return HPCNodeAggregatorConfig{
		ListenAddr:          ":8081",
		BatchSubmitInterval: 60 * time.Second,
		MaxBatchSize:        50,
		HeartbeatTimeout:    120 * time.Second,
		AllowedNodePubkeys:  make(map[string]bool),
	}
}

// HPCNodeAggregator aggregates node heartbeats and submits on-chain updates
type HPCNodeAggregator struct {
	config     HPCNodeAggregatorConfig
	keyManager *KeyManager

	nodes     map[string]*aggregatedNodeState
	nodesMu   sync.RWMutex
	pendingMu sync.Mutex
	pending   []*HPCNodeHeartbeat

	server *http.Server
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// aggregatedNodeState tracks state for a node
type aggregatedNodeState struct {
	NodeID             string
	ClusterID          string
	PublicKey          ed25519.PublicKey
	LastHeartbeat      time.Time
	LastSequence       uint64
	ConsecutiveMisses  int
	TotalHeartbeats    uint64
	Capacity           *HPCNodeCapacity
	Health             *HPCNodeHealth
	Latency            *HPCNodeLatency
	PendingChainUpdate bool
}

// HPCNodeHeartbeat is a heartbeat received from a node
type HPCNodeHeartbeat struct {
	NodeID         string          `json:"node_id"`
	ClusterID      string          `json:"cluster_id"`
	SequenceNumber uint64          `json:"sequence_number"`
	Timestamp      time.Time       `json:"timestamp"`
	AgentVersion   string          `json:"agent_version"`
	Capacity       HPCNodeCapacity `json:"capacity"`
	Health         HPCNodeHealth   `json:"health"`
	Latency        HPCNodeLatency  `json:"latency"`
	Jobs           HPCNodeJobs     `json:"jobs"`
	Services       HPCNodeServices `json:"services"`
}

// HPCNodeCapacity contains node capacity
type HPCNodeCapacity struct {
	CPUCoresTotal      int32  `json:"cpu_cores_total"`
	CPUCoresAvailable  int32  `json:"cpu_cores_available"`
	MemoryGBTotal      int32  `json:"memory_gb_total"`
	MemoryGBAvailable  int32  `json:"memory_gb_available"`
	GPUsTotal          int32  `json:"gpus_total"`
	GPUsAvailable      int32  `json:"gpus_available"`
	GPUType            string `json:"gpu_type,omitempty"`
	StorageGBTotal     int32  `json:"storage_gb_total"`
	StorageGBAvailable int32  `json:"storage_gb_available"`
}

// HPCNodeHealth contains node health
type HPCNodeHealth struct {
	Status                   string `json:"status"`
	UptimeSeconds            int64  `json:"uptime_seconds"`
	LoadAverage1m            string `json:"load_average_1m"`
	CPUUtilizationPercent    int32  `json:"cpu_utilization_percent"`
	MemoryUtilizationPercent int32  `json:"memory_utilization_percent"`
	SLURMState               string `json:"slurm_state,omitempty"`
}

// HPCNodeLatency contains latency measurements
type HPCNodeLatency struct {
	Measurements      []HPCLatencyProbe `json:"measurements,omitempty"`
	AvgClusterLatency int64             `json:"avg_cluster_latency_us"`
}

// HPCLatencyProbe is a latency probe
type HPCLatencyProbe struct {
	TargetNodeID string    `json:"target_node_id"`
	LatencyUs    int64     `json:"latency_us"`
	MeasuredAt   time.Time `json:"measured_at"`
}

// HPCNodeJobs contains job counts
type HPCNodeJobs struct {
	RunningCount int32 `json:"running_count"`
	PendingCount int32 `json:"pending_count"`
}

// HPCNodeServices contains service status
type HPCNodeServices struct {
	SLURMDRunning    bool   `json:"slurmd_running"`
	SLURMDVersion    string `json:"slurmd_version,omitempty"`
	ContainerRuntime string `json:"container_runtime,omitempty"`
}

// HPCHeartbeatAuth contains heartbeat authentication
type HPCHeartbeatAuth struct {
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
}

// HPCHeartbeatRequest is the heartbeat request payload
type HPCHeartbeatRequest struct {
	Heartbeat *HPCNodeHeartbeat `json:"heartbeat"`
	Auth      *HPCHeartbeatAuth `json:"auth"`
}

// HPCHeartbeatResponse is the heartbeat response
type HPCHeartbeatResponse struct {
	Accepted             bool                `json:"accepted"`
	SequenceAck          uint64              `json:"sequence_ack"`
	Timestamp            time.Time           `json:"timestamp"`
	NextHeartbeatSeconds int32               `json:"next_heartbeat_seconds"`
	Commands             []HPCNodeCommand    `json:"commands,omitempty"`
	Errors               []HPCHeartbeatError `json:"errors,omitempty"`
}

// HPCNodeCommand is a command for the node
type HPCNodeCommand struct {
	CommandID  string            `json:"command_id"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// HPCHeartbeatError is an error in heartbeat processing
type HPCHeartbeatError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewHPCNodeAggregator creates a new HPC node aggregator
func NewHPCNodeAggregator(config HPCNodeAggregatorConfig, keyManager *KeyManager) *HPCNodeAggregator {
	return &HPCNodeAggregator{
		config:     config,
		keyManager: keyManager,
		nodes:      make(map[string]*aggregatedNodeState),
		pending:    make([]*HPCNodeHeartbeat, 0),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the aggregator
func (a *HPCNodeAggregator) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/hpc/nodes/", a.handleHeartbeat)
	mux.HandleFunc("/api/v1/hpc/nodes/register", a.handleRegister)
	mux.HandleFunc("/health", a.handleHealth)

	a.server = &http.Server{
		Addr:         a.config.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	a.wg.Add(2)
	go a.runHTTPServer(ctx)
	go a.runBatchSubmitter(ctx)

	return nil
}

// Stop halts the aggregator
func (a *HPCNodeAggregator) Stop() {
	close(a.stopCh)
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = a.server.Shutdown(ctx)
	}
	a.wg.Wait()
}

//nolint:unparam // ctx kept for future graceful shutdown integration
func (a *HPCNodeAggregator) runHTTPServer(_ context.Context) {
	defer a.wg.Done()

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HPC-AGGREGATOR] HTTP server error: %v\n", err)
		}
	}()

	<-a.stopCh
}

func (a *HPCNodeAggregator) runBatchSubmitter(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.BatchSubmitInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.submitBatch(ctx)
			a.checkStaleNodes()
		}
	}
}

func (a *HPCNodeAggregator) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *HPCNodeAggregator) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID          string `json:"node_id"`
		ClusterID       string `json:"cluster_id"`
		ProviderAddress string `json:"provider_address"`
		AgentPubkey     string `json:"agent_pubkey"`
		Hostname        string `json:"hostname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate provider address
	if req.ProviderAddress != a.config.ProviderAddress {
		http.Error(w, "Provider address mismatch", http.StatusForbidden)
		return
	}

	// Decode public key
	pubkeyBytes, err := base64.StdEncoding.DecodeString(req.AgentPubkey)
	if err != nil {
		http.Error(w, "Invalid public key", http.StatusBadRequest)
		return
	}

	// Add to allowlist (in production, this would require provider signature)
	a.config.AllowedNodePubkeys[req.AgentPubkey] = true

	// Create node state
	a.nodesMu.Lock()
	a.nodes[req.NodeID] = &aggregatedNodeState{
		NodeID:        req.NodeID,
		ClusterID:     req.ClusterID,
		PublicKey:     ed25519.PublicKey(pubkeyBytes),
		LastHeartbeat: time.Now(),
	}
	a.nodesMu.Unlock()

	fmt.Printf("[HPC-AGGREGATOR] Registered node: %s\n", req.NodeID)

	w.Header().Set("Content-Type", "application/json")
	//nolint:errchkjson // simple response for HTTP handler
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"accepted": true,
		"node_id":  req.NodeID,
	})
}

func (a *HPCNodeAggregator) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HPCHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, "invalid_request", "Failed to decode request")
		return
	}

	if req.Heartbeat == nil || req.Auth == nil {
		a.writeError(w, "missing_fields", "Heartbeat and auth required")
		return
	}

	// Validate heartbeat
	response := a.processHeartbeat(req.Heartbeat, req.Auth)

	w.Header().Set("Content-Type", "application/json")
	//nolint:errchkjson // simple response for HTTP handler
	_ = json.NewEncoder(w).Encode(response)
}

func (a *HPCNodeAggregator) processHeartbeat(hb *HPCNodeHeartbeat, auth *HPCHeartbeatAuth) *HPCHeartbeatResponse {
	a.nodesMu.Lock()
	defer a.nodesMu.Unlock()

	nodeState, exists := a.nodes[hb.NodeID]
	if !exists {
		return &HPCHeartbeatResponse{
			Accepted:             false,
			SequenceAck:          hb.SequenceNumber,
			Timestamp:            time.Now(),
			NextHeartbeatSeconds: 30,
			Errors: []HPCHeartbeatError{{
				Code:    "node_not_registered",
				Message: "Node must be registered first",
			}},
		}
	}

	// Verify signature (if we have the public key)
	if nodeState.PublicKey != nil {
		if !a.verifySignature(hb, auth, nodeState.PublicKey) {
			return &HPCHeartbeatResponse{
				Accepted:             false,
				SequenceAck:          hb.SequenceNumber,
				Timestamp:            time.Now(),
				NextHeartbeatSeconds: 30,
				Errors: []HPCHeartbeatError{{
					Code:    "signature_invalid",
					Message: "Heartbeat signature verification failed",
				}},
			}
		}
	}

	// Validate sequence number
	if hb.SequenceNumber <= nodeState.LastSequence {
		return &HPCHeartbeatResponse{
			Accepted:             false,
			SequenceAck:          hb.SequenceNumber,
			Timestamp:            time.Now(),
			NextHeartbeatSeconds: 30,
			Errors: []HPCHeartbeatError{{
				Code:    "stale_sequence",
				Message: fmt.Sprintf("Sequence %d <= %d", hb.SequenceNumber, nodeState.LastSequence),
			}},
		}
	}

	// Update node state
	nodeState.LastHeartbeat = time.Now()
	nodeState.LastSequence = hb.SequenceNumber
	nodeState.ConsecutiveMisses = 0
	nodeState.TotalHeartbeats++
	nodeState.Capacity = &hb.Capacity
	nodeState.Health = &hb.Health
	nodeState.Latency = &hb.Latency
	nodeState.PendingChainUpdate = true

	// Queue for batch submission
	a.pendingMu.Lock()
	a.pending = append(a.pending, hb)
	a.pendingMu.Unlock()

	fmt.Printf("[HPC-AGGREGATOR] Heartbeat accepted: node=%s seq=%d\n", hb.NodeID, hb.SequenceNumber)

	// Determine next interval
	nextInterval := int32(30)
	if hb.Health.Status == "degraded" {
		nextInterval = 15
	}

	return &HPCHeartbeatResponse{
		Accepted:             true,
		SequenceAck:          hb.SequenceNumber,
		Timestamp:            time.Now(),
		NextHeartbeatSeconds: nextInterval,
	}
}

func (a *HPCNodeAggregator) verifySignature(hb *HPCNodeHeartbeat, auth *HPCHeartbeatAuth, pubkey ed25519.PublicKey) bool {
	// Serialize heartbeat
	data, err := json.Marshal(hb)
	if err != nil {
		return false
	}

	// Decode signature
	sig, err := base64.StdEncoding.DecodeString(auth.Signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubkey, data, sig)
}

func (a *HPCNodeAggregator) submitBatch(ctx context.Context) {
	a.pendingMu.Lock()
	if len(a.pending) == 0 {
		a.pendingMu.Unlock()
		return
	}

	// Take up to MaxBatchSize
	batch := a.pending
	if len(batch) > a.config.MaxBatchSize {
		batch = a.pending[:a.config.MaxBatchSize]
		a.pending = a.pending[a.config.MaxBatchSize:]
	} else {
		a.pending = make([]*HPCNodeHeartbeat, 0)
	}
	a.pendingMu.Unlock()

	fmt.Printf("[HPC-AGGREGATOR] Submitting batch of %d node updates\n", len(batch))

	// Build MsgUpdateNodeMetadata messages
	for _, hb := range batch {
		msg := a.buildNodeMetadataUpdate(hb)
		if err := a.submitOnChain(ctx, msg); err != nil {
			fmt.Printf("[HPC-AGGREGATOR] Failed to submit update for node %s: %v\n", hb.NodeID, err)
		}
	}
}

func (a *HPCNodeAggregator) buildNodeMetadataUpdate(hb *HPCNodeHeartbeat) map[string]interface{} {
	latencyMeasurements := make([]map[string]interface{}, 0, len(hb.Latency.Measurements))
	for _, m := range hb.Latency.Measurements {
		latencyMeasurements = append(latencyMeasurements, map[string]interface{}{
			"target_node_id": m.TargetNodeID,
			"latency_ms":     m.LatencyUs / 1000,
			"measured_at":    m.MeasuredAt.Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"@type":                "/virtengine.hpc.v1.MsgUpdateNodeMetadata",
		"provider_address":     a.config.ProviderAddress,
		"node_id":              hb.NodeID,
		"cluster_id":           hb.ClusterID,
		"latency_measurements": latencyMeasurements,
		"resources": map[string]interface{}{
			"cpu_cores":  hb.Capacity.CPUCoresTotal,
			"memory_gb":  hb.Capacity.MemoryGBTotal,
			"gpus":       hb.Capacity.GPUsTotal,
			"gpu_type":   hb.Capacity.GPUType,
			"storage_gb": hb.Capacity.StorageGBTotal,
		},
		"active": hb.Health.Status == "healthy" || hb.Health.Status == "degraded",
	}
}

func (a *HPCNodeAggregator) submitOnChain(ctx context.Context, msg map[string]interface{}) error {
	if a.config.ChainGRPC == "" {
		// In demo mode, just log
		fmt.Printf("[HPC-AGGREGATOR] Would submit: %v\n", msg["node_id"])
		return nil
	}

	// In production, this would use cosmos-sdk client to broadcast tx
	// For now, we'll make an HTTP call to a tx submission endpoint
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/cosmos/tx/v1beta1/txs", a.config.ChainGRPC),
		bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("chain submission failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

func (a *HPCNodeAggregator) checkStaleNodes() {
	a.nodesMu.Lock()
	defer a.nodesMu.Unlock()

	now := time.Now()
	for nodeID, state := range a.nodes {
		timeSinceHB := now.Sub(state.LastHeartbeat)
		if timeSinceHB > a.config.HeartbeatTimeout {
			state.ConsecutiveMisses++
			fmt.Printf("[HPC-AGGREGATOR] Node %s is stale (missed %d heartbeats)\n",
				nodeID, state.ConsecutiveMisses)

			// After 5 consecutive misses, consider offline
			if state.ConsecutiveMisses >= 5 {
				fmt.Printf("[HPC-AGGREGATOR] Node %s marked offline\n", nodeID)
			}
		}
	}
}

func (a *HPCNodeAggregator) writeError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	//nolint:errchkjson // simple error response for HTTP handler
	_ = json.NewEncoder(w).Encode(HPCHeartbeatResponse{
		Accepted:             false,
		Timestamp:            time.Now(),
		NextHeartbeatSeconds: 30,
		Errors: []HPCHeartbeatError{{
			Code:    code,
			Message: message,
		}},
	})
}

// GetNodeCount returns the number of tracked nodes
func (a *HPCNodeAggregator) GetNodeCount() int {
	a.nodesMu.RLock()
	defer a.nodesMu.RUnlock()
	return len(a.nodes)
}

// GetNodeStats returns stats for a specific node
func (a *HPCNodeAggregator) GetNodeStats(nodeID string) (map[string]interface{}, bool) {
	a.nodesMu.RLock()
	defer a.nodesMu.RUnlock()

	state, exists := a.nodes[nodeID]
	if !exists {
		return nil, false
	}

	return map[string]interface{}{
		"node_id":            state.NodeID,
		"cluster_id":         state.ClusterID,
		"last_heartbeat":     state.LastHeartbeat,
		"last_sequence":      state.LastSequence,
		"consecutive_misses": state.ConsecutiveMisses,
		"total_heartbeats":   state.TotalHeartbeats,
	}, true
}
