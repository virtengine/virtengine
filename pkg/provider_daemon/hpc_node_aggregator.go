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

	"github.com/virtengine/virtengine/pkg/observability"
	"github.com/virtengine/virtengine/pkg/security"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// Health status constants
const (
	healthStatusHealthy  = "healthy"
	healthStatusDegraded = "degraded"
	healthStatusOffline  = "offline"
)

// HPCNodeAggregatorConfig contains configuration for the HPC node aggregator
type HPCNodeAggregatorConfig struct {
	// Enabled toggles the aggregator.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address" yaml:"provider_address"`

	// ProviderID is the provider's public key
	ProviderID string `json:"provider_id" yaml:"provider_id"`

	// ClusterID is the default cluster ID for registrations
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`

	// ListenAddr is the HTTP listen address for node heartbeats
	ListenAddr string `json:"listen_addr" yaml:"listen_addr"`

	// BatchSubmitInterval is how often to batch submit node metadata updates
	BatchSubmitInterval time.Duration `json:"batch_submit_interval" yaml:"batch_submit_interval"`

	// MaxBatchSize is the maximum nodes to update in one batch
	MaxBatchSize int `json:"max_batch_size" yaml:"max_batch_size"`

	// HeartbeatTimeout is when to consider a node stale
	HeartbeatTimeout time.Duration `json:"heartbeat_timeout" yaml:"heartbeat_timeout"`

	// ChainGRPC is the chain gRPC endpoint for on-chain submissions
	ChainGRPC string `json:"chain_grpc" yaml:"chain_grpc"`

	// AllowedNodePubkeys is the allowlist of node public keys (base64)
	AllowedNodePubkeys map[string]bool `json:"allowed_node_pubkeys" yaml:"allowed_node_pubkeys"`

	// ChainSubmitEnabled toggles on-chain submissions.
	ChainSubmitEnabled bool `json:"chain_submit_enabled" yaml:"chain_submit_enabled"`

	// CheckpointFile is the path to persist node state.
	CheckpointFile string `json:"checkpoint_file" yaml:"checkpoint_file"`

	// CheckpointInterval is how often to persist checkpoints.
	CheckpointInterval time.Duration `json:"checkpoint_interval" yaml:"checkpoint_interval"`

	// MaxSubmitRetries is the number of retries per update.
	MaxSubmitRetries int `json:"max_submit_retries" yaml:"max_submit_retries"`

	// RetryBackoff is the base backoff between retries.
	RetryBackoff time.Duration `json:"retry_backoff" yaml:"retry_backoff"`

	// StaleMissThreshold is the consecutive miss count to mark offline.
	StaleMissThreshold int `json:"stale_miss_threshold" yaml:"stale_miss_threshold"`

	// DefaultRegion is used when heartbeats don't include region.
	DefaultRegion string `json:"default_region" yaml:"default_region"`

	// DefaultDatacenter is used when heartbeats don't include datacenter.
	DefaultDatacenter string `json:"default_datacenter" yaml:"default_datacenter"`

	// DiscoveryEnabled enables scheduler-based node discovery.
	DiscoveryEnabled bool `json:"discovery_enabled" yaml:"discovery_enabled"`

	// DiscoveryInterval controls how often to discover nodes.
	DiscoveryInterval time.Duration `json:"discovery_interval" yaml:"discovery_interval"`

	// ChainReporter is the on-chain reporter (optional).
	ChainReporter HPCNodeChainReporter `json:"-" yaml:"-"`

	// CheckpointStore overrides the checkpoint store (optional).
	CheckpointStore *HPCNodeCheckpointStore `json:"-" yaml:"-"`

	// NodeDiscoverer provides node discovery results (optional).
	NodeDiscoverer HPCNodeDiscoverer `json:"-" yaml:"-"`
}

// HPCNodeDiscoverer defines scheduler-backed node discovery.
type HPCNodeDiscoverer interface {
	ListNodes(ctx context.Context) ([]HPCDiscoveredNode, error)
}

// HPCDiscoveredNode describes a node discovered from the scheduler.
type HPCDiscoveredNode struct {
	NodeID      string
	ClusterID   string
	Region      string
	Datacenter  string
	Capacity    *HPCNodeCapacity
	Hardware    *HPCNodeHardware
	Topology    *HPCNodeTopology
	Locality    *HPCNodeLocality
	AgentPubkey string
}

// DefaultHPCNodeAggregatorConfig returns the default configuration
func DefaultHPCNodeAggregatorConfig() HPCNodeAggregatorConfig {
	return HPCNodeAggregatorConfig{
		Enabled:             false,
		ListenAddr:          ":8081",
		BatchSubmitInterval: 60 * time.Second,
		MaxBatchSize:        50,
		HeartbeatTimeout:    120 * time.Second,
		AllowedNodePubkeys:  make(map[string]bool),
		ChainSubmitEnabled:  true,
		CheckpointInterval:  30 * time.Second,
		MaxSubmitRetries:    5,
		RetryBackoff:        5 * time.Second,
		StaleMissThreshold:  5,
		DiscoveryEnabled:    true,
		DiscoveryInterval:   2 * time.Minute,
	}
}

// HPCNodeAggregator aggregates node heartbeats and submits on-chain updates
type HPCNodeAggregator struct {
	config     HPCNodeAggregatorConfig
	keyManager *KeyManager

	nodes     map[string]*aggregatedNodeState
	nodesMu   sync.RWMutex
	pendingMu sync.Mutex
	pending   []*nodeUpdate

	server *http.Server
	stopCh chan struct{}
	wg     sync.WaitGroup

	chainReporter   HPCNodeChainReporter
	checkpointStore *HPCNodeCheckpointStore
	httpClient      *http.Client
}

// aggregatedNodeState tracks state for a node
type aggregatedNodeState struct {
	NodeID                string
	ClusterID             string
	PublicKey             ed25519.PublicKey
	AgentPubkey           string
	HardwareFingerprint   string
	AgentVersion          string
	LastHeartbeat         time.Time
	LastSequence          uint64
	ConsecutiveMisses     int
	TotalHeartbeats       uint64
	Capacity              *HPCNodeCapacity
	Health                *HPCNodeHealth
	Latency               *HPCNodeLatency
	Hardware              *HPCNodeHardware
	Topology              *HPCNodeTopology
	Locality              *HPCNodeLocality
	PendingChainUpdate    bool
	OnChainRegistered     bool
	LastSubmittedSequence uint64
	LastChainUpdate       time.Time
	LastSubmitError       string
	RegisteredAt          time.Time
}

type nodeUpdateType string

const (
	nodeUpdateRegistration nodeUpdateType = "registration"
	nodeUpdateHeartbeat    nodeUpdateType = "heartbeat"
	nodeUpdateOffline      nodeUpdateType = "offline"
)

type nodeUpdate struct {
	NodeID         string
	ClusterID      string
	UpdateType     nodeUpdateType
	Heartbeat      *HPCNodeHeartbeat
	Active         bool
	SequenceNumber uint64
	Attempts       int
	NextAttempt    time.Time
	EnqueuedAt     time.Time
}

type nodeRegistration struct {
	NodeID              string           `json:"node_id"`
	ClusterID           string           `json:"cluster_id"`
	ProviderAddress     string           `json:"provider_address"`
	AgentPubkey         string           `json:"agent_pubkey"`
	Hostname            string           `json:"hostname,omitempty"`
	HardwareFingerprint string           `json:"hardware_fingerprint,omitempty"`
	AgentVersion        string           `json:"agent_version,omitempty"`
	Region              string           `json:"region,omitempty"`
	Datacenter          string           `json:"datacenter,omitempty"`
	Capacity            *HPCNodeCapacity `json:"capacity,omitempty"`
	Health              *HPCNodeHealth   `json:"health,omitempty"`
	Hardware            *HPCNodeHardware `json:"hardware,omitempty"`
	Topology            *HPCNodeTopology `json:"topology,omitempty"`
	Locality            *HPCNodeLocality `json:"locality,omitempty"`
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
	CPUCoresAllocated  int32  `json:"cpu_cores_allocated,omitempty"`
	MemoryGBTotal      int32  `json:"memory_gb_total"`
	MemoryGBAvailable  int32  `json:"memory_gb_available"`
	MemoryGBAllocated  int32  `json:"memory_gb_allocated,omitempty"`
	GPUsTotal          int32  `json:"gpus_total"`
	GPUsAvailable      int32  `json:"gpus_available"`
	GPUsAllocated      int32  `json:"gpus_allocated,omitempty"`
	GPUType            string `json:"gpu_type,omitempty"`
	StorageGBTotal     int32  `json:"storage_gb_total"`
	StorageGBAvailable int32  `json:"storage_gb_available"`
	StorageGBAllocated int32  `json:"storage_gb_allocated,omitempty"`
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

// HPCNodeHardware contains node hardware details
type HPCNodeHardware struct {
	CPUModel       string   `json:"cpu_model,omitempty"`
	CPUVendor      string   `json:"cpu_vendor,omitempty"`
	CPUArch        string   `json:"cpu_arch,omitempty"`
	Sockets        int32    `json:"sockets,omitempty"`
	CoresPerSocket int32    `json:"cores_per_socket,omitempty"`
	ThreadsPerCore int32    `json:"threads_per_core,omitempty"`
	MemoryType     string   `json:"memory_type,omitempty"`
	MemorySpeedMHz int32    `json:"memory_speed_mhz,omitempty"`
	GPUModel       string   `json:"gpu_model,omitempty"`
	GPUMemoryGB    int32    `json:"gpu_memory_gb,omitempty"`
	StorageType    string   `json:"storage_type,omitempty"`
	Features       []string `json:"features,omitempty"`
}

// HPCNodeTopology describes node topology
type HPCNodeTopology struct {
	NUMANodes     int32  `json:"numa_nodes,omitempty"`
	NUMAMemoryGB  int32  `json:"numa_memory_gb,omitempty"`
	Interconnect  string `json:"interconnect,omitempty"`
	NetworkFabric string `json:"network_fabric,omitempty"`
	TopologyHint  string `json:"topology_hint,omitempty"`
}

// HPCNodeLocality describes node locality
type HPCNodeLocality struct {
	Region     string `json:"region,omitempty"`
	Datacenter string `json:"datacenter,omitempty"`
	Zone       string `json:"zone,omitempty"`
	Rack       string `json:"rack,omitempty"`
	Row        string `json:"row,omitempty"`
	Position   string `json:"position,omitempty"`
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

// NewHPCNodeAggregator creates a new HPC node aggregator.
func NewHPCNodeAggregator(config HPCNodeAggregatorConfig, keyManager *KeyManager) (*HPCNodeAggregator, error) {
	if config.AllowedNodePubkeys == nil {
		config.AllowedNodePubkeys = make(map[string]bool)
	}
	if config.BatchSubmitInterval == 0 {
		config.BatchSubmitInterval = 60 * time.Second
	}
	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = 50
	}
	if config.HeartbeatTimeout == 0 {
		config.HeartbeatTimeout = 120 * time.Second
	}
	if config.CheckpointInterval == 0 {
		config.CheckpointInterval = 30 * time.Second
	}
	if config.MaxSubmitRetries == 0 {
		config.MaxSubmitRetries = 5
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = 5 * time.Second
	}
	if config.StaleMissThreshold == 0 {
		config.StaleMissThreshold = 5
	}

	checkpointStore := config.CheckpointStore
	if checkpointStore == nil && config.CheckpointFile != "" {
		store, err := NewHPCNodeCheckpointStore(config.CheckpointFile)
		if err != nil {
			return nil, err
		}
		checkpointStore = store
	}

	return &HPCNodeAggregator{
		config:          config,
		keyManager:      keyManager,
		nodes:           make(map[string]*aggregatedNodeState),
		pending:         make([]*nodeUpdate, 0),
		stopCh:          make(chan struct{}),
		chainReporter:   config.ChainReporter,
		checkpointStore: checkpointStore,
		httpClient:      security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
	}, nil
}

// Start begins the aggregator
func (a *HPCNodeAggregator) Start(ctx context.Context) error {
	if a.checkpointStore != nil {
		if err := a.loadCheckpoint(); err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/hpc/nodes/", a.handleHeartbeat)
	mux.HandleFunc("/api/v1/hpc/nodes/register", a.handleRegister)
	mux.HandleFunc("/health", a.handleHealth)

	a.server = &http.Server{
		Addr:         a.config.ListenAddr,
		Handler:      observability.HTTPTracingHandler(mux, "provider.hpc_node_aggregator"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	discoveryEnabled := a.config.DiscoveryEnabled && a.config.NodeDiscoverer != nil
	workers := 2
	if discoveryEnabled {
		workers++
	}
	a.wg.Add(workers)
	go a.runHTTPServer(ctx)
	go a.runBatchSubmitter(ctx)
	if discoveryEnabled {
		go a.runDiscoveryLoop(ctx)
	}

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
	if a.checkpointStore != nil {
		_ = a.persistCheckpoint()
	}
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
	checkpointTicker := time.NewTicker(a.config.CheckpointInterval)
	defer checkpointTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = a.persistCheckpoint()
			return
		case <-a.stopCh:
			_ = a.persistCheckpoint()
			return
		case <-ticker.C:
			a.submitBatch(ctx)
			a.checkStaleNodes()
		case <-checkpointTicker.C:
			_ = a.persistCheckpoint()
		}
	}
}

func (a *HPCNodeAggregator) runDiscoveryLoop(ctx context.Context) {
	defer a.wg.Done()

	interval := a.config.DiscoveryInterval
	if interval <= 0 {
		interval = 2 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	a.discoverNodes(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.discoverNodes(ctx)
		}
	}
}

func (a *HPCNodeAggregator) discoverNodes(ctx context.Context) {
	if a.config.NodeDiscoverer == nil {
		return
	}

	nodes, err := a.config.NodeDiscoverer.ListNodes(ctx)
	if err != nil {
		fmt.Printf("[HPC-AGGREGATOR] Node discovery failed: %v\n", err)
		return
	}

	for _, node := range nodes {
		a.nodesMu.RLock()
		_, exists := a.nodes[node.NodeID]
		a.nodesMu.RUnlock()
		if exists {
			continue
		}

		reg := nodeRegistration{
			NodeID:          node.NodeID,
			ClusterID:       node.ClusterID,
			ProviderAddress: a.config.ProviderAddress,
			Region:          node.Region,
			Datacenter:      node.Datacenter,
			Capacity:        node.Capacity,
			Hardware:        node.Hardware,
			Topology:        node.Topology,
			Locality:        node.Locality,
			AgentPubkey:     node.AgentPubkey,
		}

		if err := a.registerNode(reg); err != nil {
			fmt.Printf("[HPC-AGGREGATOR] Discovery registration failed for %s: %v\n", node.NodeID, err)
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

	var req nodeRegistration

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := a.registerNode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	//nolint:errchkjson // simple response for HTTP handler
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"accepted": true,
		"node_id":  req.NodeID,
	})
}

func (a *HPCNodeAggregator) registerNode(reg nodeRegistration) error {
	if reg.ProviderAddress != a.config.ProviderAddress {
		return fmt.Errorf("provider address mismatch")
	}
	if reg.NodeID == "" {
		return fmt.Errorf("node_id required")
	}
	if reg.ClusterID == "" {
		reg.ClusterID = a.config.ClusterID
	}
	if reg.ClusterID == "" {
		return fmt.Errorf("cluster_id required")
	}

	var pubkeyBytes []byte
	if reg.AgentPubkey != "" {
		var err error
		pubkeyBytes, err = base64.StdEncoding.DecodeString(reg.AgentPubkey)
		if err != nil {
			return fmt.Errorf("invalid public key")
		}
	}

	if reg.AgentPubkey != "" {
		a.config.AllowedNodePubkeys[reg.AgentPubkey] = true
	}

	now := time.Now()
	a.nodesMu.Lock()
	state, exists := a.nodes[reg.NodeID]
	if !exists {
		state = &aggregatedNodeState{
			NodeID:        reg.NodeID,
			ClusterID:     reg.ClusterID,
			PublicKey:     ed25519.PublicKey(pubkeyBytes),
			AgentPubkey:   reg.AgentPubkey,
			LastHeartbeat: now,
			RegisteredAt:  now,
		}
		a.nodes[reg.NodeID] = state
	} else {
		state.ClusterID = reg.ClusterID
		if len(pubkeyBytes) > 0 {
			state.PublicKey = ed25519.PublicKey(pubkeyBytes)
			state.AgentPubkey = reg.AgentPubkey
		}
		state.LastHeartbeat = now
		if state.RegisteredAt.IsZero() {
			state.RegisteredAt = now
		}
	}
	if reg.HardwareFingerprint != "" {
		state.HardwareFingerprint = reg.HardwareFingerprint
	}
	if reg.AgentVersion != "" {
		state.AgentVersion = reg.AgentVersion
	}
	if reg.Capacity != nil {
		state.Capacity = reg.Capacity
	}
	if reg.Health != nil {
		state.Health = reg.Health
	}
	if reg.Hardware != nil {
		state.Hardware = reg.Hardware
	}
	if reg.Topology != nil {
		state.Topology = reg.Topology
	}
	if reg.Locality != nil {
		state.Locality = reg.Locality
	} else if reg.Region != "" || reg.Datacenter != "" {
		state.Locality = &HPCNodeLocality{
			Region:     reg.Region,
			Datacenter: reg.Datacenter,
		}
	}
	state.PendingChainUpdate = true
	a.nodesMu.Unlock()

	a.enqueueUpdate(&nodeUpdate{
		NodeID:     reg.NodeID,
		ClusterID:  reg.ClusterID,
		UpdateType: nodeUpdateRegistration,
		Active:     false,
		EnqueuedAt: now,
	})

	fmt.Printf("[HPC-AGGREGATOR] Registered node: %s\n", reg.NodeID)
	return nil
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

	if nodeState.ClusterID != "" && hb.ClusterID != nodeState.ClusterID {
		return &HPCHeartbeatResponse{
			Accepted:             false,
			SequenceAck:          hb.SequenceNumber,
			Timestamp:            time.Now(),
			NextHeartbeatSeconds: 30,
			Errors: []HPCHeartbeatError{{
				Code:    "cluster_mismatch",
				Message: fmt.Sprintf("cluster %s does not match %s", hb.ClusterID, nodeState.ClusterID),
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
	nodeState.AgentVersion = hb.AgentVersion
	nodeState.Capacity = &hb.Capacity
	nodeState.Health = &hb.Health
	nodeState.Latency = &hb.Latency
	nodeState.PendingChainUpdate = true
	nodeState.LastSubmitError = ""

	// Queue for batch submission
	a.enqueueUpdate(&nodeUpdate{
		NodeID:         hb.NodeID,
		ClusterID:      hb.ClusterID,
		UpdateType:     nodeUpdateHeartbeat,
		Heartbeat:      hb,
		Active:         a.isActiveHealth(hb.Health.Status),
		SequenceNumber: hb.SequenceNumber,
		EnqueuedAt:     time.Now(),
	})

	fmt.Printf("[HPC-AGGREGATOR] Heartbeat accepted: node=%s seq=%d\n", hb.NodeID, hb.SequenceNumber)

	// Determine next interval
	nextInterval := int32(30)
	if hb.Health.Status == healthStatusDegraded {
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
	now := time.Now()
	updates := a.dequeueUpdates(now)
	if len(updates) == 0 {
		return
	}

	fmt.Printf("[HPC-AGGREGATOR] Submitting batch of %d node updates\n", len(updates))

	for _, update := range updates {
		msg := a.buildNodeMetadataUpdate(update)
		if msg == nil {
			a.markUpdateFailed(update, fmt.Errorf("failed to build update"))
			continue
		}

		if err := a.submitOnChain(ctx, msg); err != nil {
			a.markUpdateFailed(update, err)
			continue
		}

		a.markUpdateSuccess(update)
	}
}

func (a *HPCNodeAggregator) buildNodeMetadataUpdate(update *nodeUpdate) *hpcv1.MsgUpdateNodeMetadata {
	if update == nil {
		return nil
	}

	a.nodesMu.RLock()
	state := a.nodes[update.NodeID]
	a.nodesMu.RUnlock()

	if state == nil {
		return nil
	}
	if update.UpdateType == nodeUpdateHeartbeat && update.SequenceNumber > 0 &&
		update.SequenceNumber <= state.LastSubmittedSequence {
		return nil
	}

	heartbeat := update.Heartbeat
	latency := make([]hpcv1.LatencyMeasurement, 0)
	if heartbeat != nil {
		latency = make([]hpcv1.LatencyMeasurement, 0, len(heartbeat.Latency.Measurements))
		for _, m := range heartbeat.Latency.Measurements {
			latency = append(latency, hpcv1.LatencyMeasurement{
				TargetNodeId: m.TargetNodeID,
				LatencyMs:    m.LatencyUs / 1000,
				MeasuredAt:   m.MeasuredAt,
			})
		}
	} else if state.Latency != nil {
		latency = make([]hpcv1.LatencyMeasurement, 0, len(state.Latency.Measurements))
		for _, m := range state.Latency.Measurements {
			latency = append(latency, hpcv1.LatencyMeasurement{
				TargetNodeId: m.TargetNodeID,
				LatencyMs:    m.LatencyUs / 1000,
				MeasuredAt:   m.MeasuredAt,
			})
		}
	}

	capacity := &HPCNodeCapacity{}
	health := &HPCNodeHealth{}
	agentVersion := state.AgentVersion
	if heartbeat != nil {
		capacity = &heartbeat.Capacity
		health = &heartbeat.Health
		if heartbeat.AgentVersion != "" {
			agentVersion = heartbeat.AgentVersion
		}
	} else if state.Capacity != nil {
		capacity = state.Capacity
	}
	if heartbeat == nil && state.Health != nil {
		health = state.Health
	}

	healthStatus := healthStatusFromString(health.Status)
	nodeState := nodeStateForUpdate(update, healthStatus)

	region := a.config.DefaultRegion
	datacenter := a.config.DefaultDatacenter
	if state.Locality != nil {
		if state.Locality.Region != "" {
			region = state.Locality.Region
		}
		if state.Locality.Datacenter != "" {
			datacenter = state.Locality.Datacenter
		}
	}

	return &hpcv1.MsgUpdateNodeMetadata{
		ProviderAddress:     a.config.ProviderAddress,
		NodeId:              update.NodeID,
		ClusterId:           update.ClusterID,
		Region:              region,
		Datacenter:          datacenter,
		LatencyMeasurements: latency,
		Resources: &hpcv1.NodeResources{
			CpuCores:  capacity.CPUCoresTotal,
			MemoryGb:  capacity.MemoryGBTotal,
			Gpus:      capacity.GPUsTotal,
			GpuType:   capacity.GPUType,
			StorageGb: capacity.StorageGBTotal,
		},
		Active:              update.Active,
		State:               nodeState,
		HealthStatus:        healthStatus,
		AgentPubkey:         state.AgentPubkey,
		HardwareFingerprint: state.HardwareFingerprint,
		AgentVersion:        agentVersion,
		LastSequenceNumber:  update.SequenceNumber,
		Capacity:            nodeCapacityToProto(capacity),
		Health:              nodeHealthToProto(health),
		Hardware:            nodeHardwareToProto(state.Hardware),
		Topology:            nodeTopologyToProto(state.Topology),
		Locality:            nodeLocalityToProto(state.Locality, a.config.DefaultRegion, a.config.DefaultDatacenter),
	}
}

func (a *HPCNodeAggregator) submitOnChain(ctx context.Context, msg *hpcv1.MsgUpdateNodeMetadata) error {
	if !a.config.ChainSubmitEnabled {
		return nil
	}

	if a.chainReporter != nil {
		return a.chainReporter.SubmitNodeMetadata(ctx, msg)
	}

	if a.config.ChainGRPC == "" {
		fmt.Printf("[HPC-AGGREGATOR] Would submit: %s\n", msg.NodeId)
		return nil
	}

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

	resp, err := a.httpClient.Do(req)
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

			if state.ConsecutiveMisses >= a.config.StaleMissThreshold {
				if state.Health == nil {
					state.Health = &HPCNodeHealth{Status: healthStatusOffline}
				} else {
					state.Health.Status = healthStatusOffline
				}
				state.PendingChainUpdate = true
				state.LastSubmitError = ""

				a.enqueueUpdate(&nodeUpdate{
					NodeID:         nodeID,
					ClusterID:      state.ClusterID,
					UpdateType:     nodeUpdateOffline,
					Active:         false,
					SequenceNumber: state.LastSequence,
					EnqueuedAt:     time.Now(),
				})

				fmt.Printf("[HPC-AGGREGATOR] Node %s marked offline\n", nodeID)
			}
		}
	}
}

func (a *HPCNodeAggregator) enqueueUpdate(update *nodeUpdate) {
	if update == nil {
		return
	}
	if update.ClusterID == "" {
		a.nodesMu.RLock()
		if state, ok := a.nodes[update.NodeID]; ok {
			update.ClusterID = state.ClusterID
		}
		a.nodesMu.RUnlock()
		if update.ClusterID == "" {
			update.ClusterID = a.config.ClusterID
		}
	}
	if update.EnqueuedAt.IsZero() {
		update.EnqueuedAt = time.Now()
	}

	a.pendingMu.Lock()
	a.pending = append(a.pending, update)
	a.pendingMu.Unlock()
}

func (a *HPCNodeAggregator) dequeueUpdates(now time.Time) []*nodeUpdate {
	a.pendingMu.Lock()
	defer a.pendingMu.Unlock()

	if len(a.pending) == 0 {
		return nil
	}

	batchCap := len(a.pending)
	if a.config.MaxBatchSize > 0 && a.config.MaxBatchSize < batchCap {
		batchCap = a.config.MaxBatchSize
	}
	batch := make([]*nodeUpdate, 0, batchCap)
	remaining := make([]*nodeUpdate, 0, len(a.pending))

	for i, update := range a.pending {
		if !update.NextAttempt.IsZero() && update.NextAttempt.After(now) {
			remaining = append(remaining, update)
			continue
		}

		batch = append(batch, update)
		if a.config.MaxBatchSize > 0 && len(batch) >= a.config.MaxBatchSize {
			remaining = append(remaining, a.pending[i+1:]...)
			break
		}
	}

	a.pending = remaining

	return batch
}

func (a *HPCNodeAggregator) markUpdateFailed(update *nodeUpdate, err error) {
	if update == nil {
		return
	}

	update.Attempts++
	if update.Attempts > a.config.MaxSubmitRetries {
		a.nodesMu.Lock()
		if state, ok := a.nodes[update.NodeID]; ok {
			state.PendingChainUpdate = false
			if err != nil {
				state.LastSubmitError = err.Error()
			}
		}
		a.nodesMu.Unlock()
		return
	}

	backoff := a.config.RetryBackoff
	if backoff <= 0 {
		backoff = 5 * time.Second
	}
	delay := backoff * time.Duration(1<<minInt(update.Attempts-1, 5))
	update.NextAttempt = time.Now().Add(delay)

	if err != nil {
		a.nodesMu.Lock()
		if state, ok := a.nodes[update.NodeID]; ok {
			state.LastSubmitError = err.Error()
			state.PendingChainUpdate = true
		}
		a.nodesMu.Unlock()
	}

	a.enqueueUpdate(update)
	fmt.Printf("[HPC-AGGREGATOR] Failed to submit update for node %s: %v\n", update.NodeID, err)
}

func (a *HPCNodeAggregator) markUpdateSuccess(update *nodeUpdate) {
	if update == nil {
		return
	}

	a.nodesMu.Lock()
	if state, ok := a.nodes[update.NodeID]; ok {
		if update.UpdateType == nodeUpdateRegistration {
			state.OnChainRegistered = true
		}
		if update.SequenceNumber > state.LastSubmittedSequence {
			state.LastSubmittedSequence = update.SequenceNumber
		}
		state.LastChainUpdate = time.Now()
		state.PendingChainUpdate = false
		state.LastSubmitError = ""
	}
	a.nodesMu.Unlock()
}

func (a *HPCNodeAggregator) isActiveHealth(status string) bool {
	switch status {
	case healthStatusHealthy, healthStatusDegraded:
		return true
	default:
		return false
	}
}

func healthStatusFromString(status string) hpcv1.HealthStatus {
	switch status {
	case healthStatusHealthy:
		return hpcv1.HealthStatusHealthy
	case healthStatusDegraded:
		return hpcv1.HealthStatusDegraded
	case "unhealthy":
		return hpcv1.HealthStatusUnhealthy
	case "draining":
		return hpcv1.HealthStatusDraining
	case healthStatusOffline:
		return hpcv1.HealthStatusOffline
	default:
		return hpcv1.HealthStatusUnspecified
	}
}

func nodeStateForUpdate(update *nodeUpdate, healthStatus hpcv1.HealthStatus) hpcv1.NodeState {
	if update == nil {
		return hpcv1.NodeStateUnspecified
	}

	switch update.UpdateType {
	case nodeUpdateRegistration:
		return hpcv1.NodeStatePending
	case nodeUpdateOffline:
		return hpcv1.NodeStateOffline
	}

	switch healthStatus {
	case hpcv1.HealthStatusOffline:
		return hpcv1.NodeStateOffline
	case hpcv1.HealthStatusDraining:
		return hpcv1.NodeStateDraining
	}

	if update.Active {
		return hpcv1.NodeStateActive
	}

	return hpcv1.NodeStateStale
}

func nodeCapacityToProto(capacity *HPCNodeCapacity) *hpcv1.NodeCapacity {
	if capacity == nil {
		return nil
	}
	return &hpcv1.NodeCapacity{
		CpuCoresTotal:      capacity.CPUCoresTotal,
		CpuCoresAvailable:  capacity.CPUCoresAvailable,
		CpuCoresAllocated:  capacity.CPUCoresAllocated,
		MemoryGbTotal:      capacity.MemoryGBTotal,
		MemoryGbAvailable:  capacity.MemoryGBAvailable,
		MemoryGbAllocated:  capacity.MemoryGBAllocated,
		GpusTotal:          capacity.GPUsTotal,
		GpusAvailable:      capacity.GPUsAvailable,
		GpusAllocated:      capacity.GPUsAllocated,
		GpuType:            capacity.GPUType,
		StorageGbTotal:     capacity.StorageGBTotal,
		StorageGbAvailable: capacity.StorageGBAvailable,
		StorageGbAllocated: capacity.StorageGBAllocated,
	}
}

func nodeHealthToProto(health *HPCNodeHealth) *hpcv1.NodeHealth {
	if health == nil {
		return nil
	}
	return &hpcv1.NodeHealth{
		Status:                   healthStatusFromString(health.Status),
		UptimeSeconds:            health.UptimeSeconds,
		LoadAverage_1M:           health.LoadAverage1m,
		CpuUtilizationPercent:    health.CPUUtilizationPercent,
		MemoryUtilizationPercent: health.MemoryUtilizationPercent,
		SlurmState:               health.SLURMState,
	}
}

func nodeHardwareToProto(hardware *HPCNodeHardware) *hpcv1.NodeHardware {
	if hardware == nil {
		return nil
	}
	return &hpcv1.NodeHardware{
		CpuModel:       hardware.CPUModel,
		CpuVendor:      hardware.CPUVendor,
		CpuArch:        hardware.CPUArch,
		Sockets:        hardware.Sockets,
		CoresPerSocket: hardware.CoresPerSocket,
		ThreadsPerCore: hardware.ThreadsPerCore,
		MemoryType:     hardware.MemoryType,
		MemorySpeedMhz: hardware.MemorySpeedMHz,
		GpuModel:       hardware.GPUModel,
		GpuMemoryGb:    hardware.GPUMemoryGB,
		StorageType:    hardware.StorageType,
		Features:       hardware.Features,
	}
}

func nodeTopologyToProto(topology *HPCNodeTopology) *hpcv1.NodeTopology {
	if topology == nil {
		return nil
	}
	return &hpcv1.NodeTopology{
		NumaNodes:     topology.NUMANodes,
		NumaMemoryGb:  topology.NUMAMemoryGB,
		Interconnect:  topology.Interconnect,
		NetworkFabric: topology.NetworkFabric,
		TopologyHint:  topology.TopologyHint,
	}
}

func nodeLocalityToProto(locality *HPCNodeLocality, defaultRegion, defaultDatacenter string) *hpcv1.NodeLocality {
	if locality == nil {
		if defaultRegion == "" && defaultDatacenter == "" {
			return nil
		}
		return &hpcv1.NodeLocality{
			Region:     defaultRegion,
			Datacenter: defaultDatacenter,
		}
	}
	return &hpcv1.NodeLocality{
		Region:     locality.Region,
		Datacenter: locality.Datacenter,
		Zone:       locality.Zone,
		Rack:       locality.Rack,
		Row:        locality.Row,
		Position:   locality.Position,
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *HPCNodeAggregator) loadCheckpoint() error {
	if a.checkpointStore == nil {
		return nil
	}

	state, err := a.checkpointStore.Load()
	if err != nil {
		return err
	}

	a.nodesMu.Lock()
	a.nodes = make(map[string]*aggregatedNodeState, len(state.Nodes))
	for _, node := range state.Nodes {
		var pubkey ed25519.PublicKey
		if node.PublicKey != "" {
			if decoded, err := base64.StdEncoding.DecodeString(node.PublicKey); err == nil {
				pubkey = ed25519.PublicKey(decoded)
				a.config.AllowedNodePubkeys[node.PublicKey] = true
			}
		}
		if node.AgentPubkey != "" {
			a.config.AllowedNodePubkeys[node.AgentPubkey] = true
		}

		a.nodes[node.NodeID] = &aggregatedNodeState{
			NodeID:                node.NodeID,
			ClusterID:             node.ClusterID,
			PublicKey:             pubkey,
			AgentPubkey:           node.AgentPubkey,
			HardwareFingerprint:   node.HardwareFingerprint,
			AgentVersion:          node.AgentVersion,
			LastHeartbeat:         node.LastHeartbeat,
			LastSequence:          node.LastSequence,
			ConsecutiveMisses:     node.ConsecutiveMisses,
			TotalHeartbeats:       node.TotalHeartbeats,
			Capacity:              node.Capacity,
			Health:                node.Health,
			Latency:               node.Latency,
			Hardware:              node.Hardware,
			Topology:              node.Topology,
			Locality:              node.Locality,
			PendingChainUpdate:    node.PendingChainUpdate,
			OnChainRegistered:     node.OnChainRegistered,
			LastSubmittedSequence: node.LastSubmittedSequence,
			LastChainUpdate:       node.LastChainUpdate,
			LastSubmitError:       node.LastSubmitError,
			RegisteredAt:          node.RegisteredAt,
		}
	}
	a.nodesMu.Unlock()

	a.pendingMu.Lock()
	a.pending = make([]*nodeUpdate, 0, len(state.Pending))
	for _, update := range state.Pending {
		a.pending = append(a.pending, &nodeUpdate{
			NodeID:         update.NodeID,
			ClusterID:      update.ClusterID,
			UpdateType:     update.UpdateType,
			Heartbeat:      update.Heartbeat,
			Active:         update.Active,
			SequenceNumber: update.SequenceNumber,
			Attempts:       update.Attempts,
			NextAttempt:    update.NextAttempt,
			EnqueuedAt:     update.EnqueuedAt,
		})
	}
	a.pendingMu.Unlock()

	return nil
}

func (a *HPCNodeAggregator) persistCheckpoint() error {
	if a.checkpointStore == nil {
		return nil
	}

	a.nodesMu.RLock()
	nodes := make([]HPCNodeCheckpointNode, 0, len(a.nodes))
	for _, state := range a.nodes {
		publicKey := ""
		if state.PublicKey != nil {
			publicKey = base64.StdEncoding.EncodeToString(state.PublicKey)
		}
		nodes = append(nodes, HPCNodeCheckpointNode{
			NodeID:                state.NodeID,
			ClusterID:             state.ClusterID,
			PublicKey:             publicKey,
			AgentPubkey:           state.AgentPubkey,
			HardwareFingerprint:   state.HardwareFingerprint,
			AgentVersion:          state.AgentVersion,
			LastHeartbeat:         state.LastHeartbeat,
			LastSequence:          state.LastSequence,
			ConsecutiveMisses:     state.ConsecutiveMisses,
			TotalHeartbeats:       state.TotalHeartbeats,
			Capacity:              state.Capacity,
			Health:                state.Health,
			Latency:               state.Latency,
			Hardware:              state.Hardware,
			Topology:              state.Topology,
			Locality:              state.Locality,
			PendingChainUpdate:    state.PendingChainUpdate,
			OnChainRegistered:     state.OnChainRegistered,
			LastSubmittedSequence: state.LastSubmittedSequence,
			LastChainUpdate:       state.LastChainUpdate,
			LastSubmitError:       state.LastSubmitError,
			RegisteredAt:          state.RegisteredAt,
		})
	}
	a.nodesMu.RUnlock()

	a.pendingMu.Lock()
	pending := make([]HPCNodeCheckpointUpdate, 0, len(a.pending))
	for _, update := range a.pending {
		pending = append(pending, HPCNodeCheckpointUpdate{
			NodeID:         update.NodeID,
			ClusterID:      update.ClusterID,
			UpdateType:     update.UpdateType,
			Heartbeat:      update.Heartbeat,
			Active:         update.Active,
			SequenceNumber: update.SequenceNumber,
			Attempts:       update.Attempts,
			NextAttempt:    update.NextAttempt,
			EnqueuedAt:     update.EnqueuedAt,
		})
	}
	a.pendingMu.Unlock()

	return a.checkpointStore.Save(&HPCNodeCheckpointState{
		Version: hpcNodeCheckpointVersion,
		Nodes:   nodes,
		Pending: pending,
	})
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
		"node_id":                 state.NodeID,
		"cluster_id":              state.ClusterID,
		"last_heartbeat":          state.LastHeartbeat,
		"last_sequence":           state.LastSequence,
		"consecutive_misses":      state.ConsecutiveMisses,
		"total_heartbeats":        state.TotalHeartbeats,
		"on_chain_registered":     state.OnChainRegistered,
		"last_submitted_sequence": state.LastSubmittedSequence,
		"last_chain_update":       state.LastChainUpdate,
		"last_submit_error":       state.LastSubmitError,
	}, true
}

// GetPendingUpdateCount returns the number of queued chain updates.
func (a *HPCNodeAggregator) GetPendingUpdateCount() int {
	a.pendingMu.Lock()
	defer a.pendingMu.Unlock()
	return len(a.pending)
}
