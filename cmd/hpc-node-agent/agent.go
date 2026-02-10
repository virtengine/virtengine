// Package main implements the VirtEngine HPC Node Agent.
//
// VE-500: Node agent implementation with heartbeat and metrics collection.
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/pkg/security"
)

// AgentConfig contains configuration for the node agent
type AgentConfig struct {
	NodeID            string
	ClusterID         string
	ProviderAddress   string
	ProviderDaemonURL string
	HeartbeatInterval time.Duration
	PrivateKey        ed25519.PrivateKey
	PublicKey         ed25519.PublicKey
	Hostname          string
	Region            string
	Datacenter        string
	Zone              string
	Rack              string
	Row               string
	Position          string
	LatencyTargets    []string
}

// Agent is the HPC node agent
type Agent struct {
	config           AgentConfig
	metricsCollector *MetricsCollector
	messageHandler   *MessageHandler
	sequenceNumber   uint64
	httpClient       *http.Client
	stopCh           chan struct{}
	wg               sync.WaitGroup
	running          int32
}

const agentVersion = "0.1.0"

// NewAgent creates a new node agent
func NewAgent(config AgentConfig) *Agent {
	agent := &Agent{
		config:           config,
		metricsCollector: NewMetricsCollector(),
		httpClient:       security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second)),
		stopCh:           make(chan struct{}),
	}
	agent.messageHandler = NewMessageHandler(agent)
	return agent
}

// Start begins the agent's heartbeat loop
func (a *Agent) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&a.running, 0, 1) {
		return fmt.Errorf("agent already running")
	}

	if err := a.registerNode(ctx); err != nil {
		fmt.Printf("[REGISTER] Error registering node: %v\n", err)
	}

	// Start message handler
	a.messageHandler.Start(ctx)

	a.wg.Add(1)
	go a.heartbeatLoop(ctx)

	return nil
}

// Stop halts the agent
func (a *Agent) Stop() {
	if !atomic.CompareAndSwapInt32(&a.running, 1, 0) {
		return
	}
	close(a.stopCh)
	a.messageHandler.Stop()
	a.wg.Wait()
}

func (a *Agent) heartbeatLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.HeartbeatInterval)
	defer ticker.Stop()

	// Send initial heartbeat
	a.sendHeartbeat(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.sendHeartbeat(ctx)
		}
	}
}

func (a *Agent) sendHeartbeat(ctx context.Context) {
	heartbeat, err := a.buildHeartbeat()
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error building heartbeat: %v\n", err)
		return
	}

	auth, err := a.signHeartbeat(heartbeat)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error signing heartbeat: %v\n", err)
		return
	}

	resp, err := a.submitHeartbeat(ctx, heartbeat, auth)
	if err != nil {
		fmt.Printf("[HEARTBEAT] Error submitting heartbeat: %v\n", err)
		return
	}

	if resp.Accepted {
		fmt.Printf("[HEARTBEAT] Accepted (seq=%d, next=%ds)\n",
			resp.SequenceAck, resp.NextHeartbeatSeconds)
	} else {
		fmt.Printf("[HEARTBEAT] Rejected: %v\n", resp.Errors)
	}

	// Process any commands from the response
	for _, cmd := range resp.Commands {
		a.processCommand(cmd)
	}
}

func (a *Agent) buildHeartbeat() (*NodeHeartbeat, error) {
	// Collect metrics
	capacity, err := a.metricsCollector.CollectCapacity()
	if err != nil {
		return nil, fmt.Errorf("failed to collect capacity: %w", err)
	}

	health, err := a.metricsCollector.CollectHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to collect health: %w", err)
	}

	latency := a.metricsCollector.CollectLatency(a.config.LatencyTargets)
	jobs := a.metricsCollector.CollectJobs()
	services := a.metricsCollector.CollectServices()

	// Increment sequence number
	a.sequenceNumber++

	return &NodeHeartbeat{
		NodeID:         a.config.NodeID,
		ClusterID:      a.config.ClusterID,
		SequenceNumber: a.sequenceNumber,
		Timestamp:      time.Now().UTC(),
		AgentVersion:   agentVersion,
		Capacity:       *capacity,
		Health:         *health,
		Latency:        *latency,
		Jobs:           *jobs,
		Services:       *services,
	}, nil
}

func (a *Agent) signHeartbeat(heartbeat *NodeHeartbeat) (*HeartbeatAuth, error) {
	// Serialize heartbeat for signing
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	// Sign with Ed25519
	signature := ed25519.Sign(a.config.PrivateKey, data)

	return &HeartbeatAuth{
		Signature: base64.StdEncoding.EncodeToString(signature),
		Nonce:     base64.StdEncoding.EncodeToString(data[:16]), // Use first 16 bytes as nonce
		Timestamp: time.Now().Unix(),
	}, nil
}

func (a *Agent) submitHeartbeat(ctx context.Context, heartbeat *NodeHeartbeat, auth *HeartbeatAuth) (*HeartbeatResponse, error) {
	payload := struct {
		Heartbeat *NodeHeartbeat `json:"heartbeat"`
		Auth      *HeartbeatAuth `json:"auth"`
	}{
		Heartbeat: heartbeat,
		Auth:      auth,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/hpc/nodes/%s/heartbeat", a.config.ProviderDaemonURL, a.config.NodeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-ID", a.config.NodeID)
	req.Header.Set("X-Cluster-ID", a.config.ClusterID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		// Return a synthetic response for offline operation
		return &HeartbeatResponse{
			Accepted:             false,
			SequenceAck:          heartbeat.SequenceNumber,
			Timestamp:            time.Now(),
			NextHeartbeatSeconds: 30,
			Errors: []HeartbeatError{{
				Code:    "connection_failed",
				Message: err.Error(),
			}},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &HeartbeatResponse{
			Accepted:             false,
			SequenceAck:          heartbeat.SequenceNumber,
			Timestamp:            time.Now(),
			NextHeartbeatSeconds: 30,
			Errors: []HeartbeatError{{
				Code:    "http_error",
				Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
			}},
		}, nil
	}

	var response HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (a *Agent) registerNode(ctx context.Context) error {
	registration, err := a.buildRegistration()
	if err != nil {
		return err
	}
	return a.submitRegistration(ctx, registration)
}

func (a *Agent) buildRegistration() (*NodeRegistrationRequest, error) {
	hardware, err := a.metricsCollector.CollectHardware()
	if err != nil {
		return nil, err
	}

	capacity, err := a.metricsCollector.CollectCapacity()
	if err != nil {
		return nil, err
	}

	health, err := a.metricsCollector.CollectHealth()
	if err != nil {
		return nil, err
	}

	locality := &NodeLocality{
		Region:     a.config.Region,
		Datacenter: a.config.Datacenter,
		Zone:       a.config.Zone,
		Rack:       a.config.Rack,
		Row:        a.config.Row,
		Position:   a.config.Position,
	}

	agentPubkey := base64.StdEncoding.EncodeToString(a.config.PublicKey)
	fingerprint := computeHardwareFingerprint(capacity, hardware)

	return &NodeRegistrationRequest{
		NodeID:              a.config.NodeID,
		ClusterID:           a.config.ClusterID,
		ProviderAddress:     a.config.ProviderAddress,
		AgentPubkey:         agentPubkey,
		Hostname:            a.config.Hostname,
		HardwareFingerprint: fingerprint,
		AgentVersion:        agentVersion,
		Region:              a.config.Region,
		Datacenter:          a.config.Datacenter,
		Capacity:            capacity,
		Health:              health,
		Hardware:            hardware,
		Locality:            locality,
	}, nil
}

func (a *Agent) submitRegistration(ctx context.Context, registration *NodeRegistrationRequest) error {
	if registration == nil {
		return fmt.Errorf("registration is nil")
	}

	data, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/hpc/nodes/register", a.config.ProviderDaemonURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

func computeHardwareFingerprint(capacity *NodeCapacity, hardware *NodeHardware) string {
	h := sha256.New()
	if hardware != nil {
		h.Write([]byte(hardware.CPUModel))
		h.Write([]byte(hardware.CPUVendor))
		h.Write([]byte(hardware.CPUArch))
		h.Write([]byte(hardware.GPUModel))
		h.Write([]byte(hardware.StorageType))
	}
	if capacity != nil {
		_, _ = fmt.Fprintf(h, "%d/%d/%d/%d",
			capacity.CPUCoresTotal,
			capacity.MemoryGBTotal,
			capacity.GPUsTotal,
			capacity.StorageGBTotal,
		)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (a *Agent) processCommand(cmd NodeCommand) {
	fmt.Printf("[COMMAND] Received: %s (%s)\n", cmd.Type, cmd.CommandID)

	switch cmd.Type {
	case "drain":
		fmt.Println("[COMMAND] Initiating node drain...")
		// In production, this would trigger SLURM drain
	case "resume":
		fmt.Println("[COMMAND] Resuming node...")
	case "shutdown":
		fmt.Println("[COMMAND] Shutdown requested...")
	case "update_agent":
		fmt.Println("[COMMAND] Agent update requested...")
	case "run_diagnostic":
		fmt.Println("[COMMAND] Running diagnostics...")
	default:
		fmt.Printf("[COMMAND] Unknown command type: %s\n", cmd.Type)
	}
}

// NodeIdentity represents the identity of a node agent
type NodeIdentity struct {
	NodeID              string `json:"node_id"`
	ClusterID           string `json:"cluster_id"`
	ProviderAddress     string `json:"provider_address"`
	AgentPubkey         string `json:"agent_pubkey"`
	Hostname            string `json:"hostname,omitempty"`
	HardwareFingerprint string `json:"hardware_fingerprint,omitempty"`
}

// NodeHeartbeat represents a heartbeat from a node agent
type NodeHeartbeat struct {
	NodeID         string       `json:"node_id"`
	ClusterID      string       `json:"cluster_id"`
	SequenceNumber uint64       `json:"sequence_number"`
	Timestamp      time.Time    `json:"timestamp"`
	AgentVersion   string       `json:"agent_version"`
	Capacity       NodeCapacity `json:"capacity"`
	Health         NodeHealth   `json:"health"`
	Latency        NodeLatency  `json:"latency"`
	Jobs           NodeJobs     `json:"jobs"`
	Services       NodeServices `json:"services"`
}

// NodeCapacity contains node capacity information
type NodeCapacity struct {
	CPUCoresTotal      int32  `json:"cpu_cores_total"`
	CPUCoresAvailable  int32  `json:"cpu_cores_available"`
	CPUCoresAllocated  int32  `json:"cpu_cores_allocated"`
	MemoryGBTotal      int32  `json:"memory_gb_total"`
	MemoryGBAvailable  int32  `json:"memory_gb_available"`
	MemoryGBAllocated  int32  `json:"memory_gb_allocated"`
	GPUsTotal          int32  `json:"gpus_total"`
	GPUsAvailable      int32  `json:"gpus_available"`
	GPUsAllocated      int32  `json:"gpus_allocated"`
	GPUType            string `json:"gpu_type,omitempty"`
	StorageGBTotal     int32  `json:"storage_gb_total"`
	StorageGBAvailable int32  `json:"storage_gb_available"`
	StorageGBAllocated int32  `json:"storage_gb_allocated"`
}

// NodeHealth contains node health information
type NodeHealth struct {
	Status                      string `json:"status"`
	UptimeSeconds               int64  `json:"uptime_seconds"`
	LoadAverage1m               string `json:"load_average_1m"`
	LoadAverage5m               string `json:"load_average_5m"`
	LoadAverage15m              string `json:"load_average_15m"`
	CPUUtilizationPercent       int32  `json:"cpu_utilization_percent"`
	MemoryUtilizationPercent    int32  `json:"memory_utilization_percent"`
	GPUUtilizationPercent       int32  `json:"gpu_utilization_percent,omitempty"`
	GPUMemoryUtilizationPercent int32  `json:"gpu_memory_utilization_percent,omitempty"`
	DiskIOUtilizationPercent    int32  `json:"disk_io_utilization_percent"`
	NetworkUtilizationPercent   int32  `json:"network_utilization_percent"`
	TemperatureCelsius          int32  `json:"temperature_celsius,omitempty"`
	GPUTemperatureCelsius       int32  `json:"gpu_temperature_celsius,omitempty"`
	ErrorCount24h               int32  `json:"error_count_24h"`
	WarningCount24h             int32  `json:"warning_count_24h"`
	LastErrorMessage            string `json:"last_error_message,omitempty"`
	SLURMState                  string `json:"slurm_state,omitempty"`
}

// NodeHardware contains node hardware details
type NodeHardware struct {
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

// NodeTopology describes node topology
type NodeTopology struct {
	NUMANodes     int32  `json:"numa_nodes,omitempty"`
	NUMAMemoryGB  int32  `json:"numa_memory_gb,omitempty"`
	Interconnect  string `json:"interconnect,omitempty"`
	NetworkFabric string `json:"network_fabric,omitempty"`
	TopologyHint  string `json:"topology_hint,omitempty"`
}

// NodeLocality describes node locality
type NodeLocality struct {
	Region     string `json:"region,omitempty"`
	Datacenter string `json:"datacenter,omitempty"`
	Zone       string `json:"zone,omitempty"`
	Rack       string `json:"rack,omitempty"`
	Row        string `json:"row,omitempty"`
	Position   string `json:"position,omitempty"`
}

// NodeLatency contains latency measurements
type NodeLatency struct {
	Measurements      []LatencyProbe `json:"measurements,omitempty"`
	GatewayLatencyUs  int64          `json:"gateway_latency_us"`
	ChainLatencyMs    int64          `json:"chain_latency_ms"`
	AvgClusterLatency int64          `json:"avg_cluster_latency_us"`
}

// LatencyProbe represents a latency measurement to another node
type LatencyProbe struct {
	TargetNodeID      string    `json:"target_node_id"`
	LatencyUs         int64     `json:"latency_us"`
	PacketLossPercent int32     `json:"packet_loss_percent"`
	MeasuredAt        time.Time `json:"measured_at"`
}

// NodeJobs contains job information
type NodeJobs struct {
	RunningCount int32    `json:"running_count"`
	PendingCount int32    `json:"pending_count"`
	Completed24h int32    `json:"completed_24h"`
	Failed24h    int32    `json:"failed_24h"`
	ActiveJobIDs []string `json:"active_job_ids,omitempty"`
}

// NodeServices contains service status
type NodeServices struct {
	SLURMDRunning           bool   `json:"slurmd_running"`
	SLURMDVersion           string `json:"slurmd_version,omitempty"`
	MungeRunning            bool   `json:"munge_running"`
	ContainerRuntime        string `json:"container_runtime,omitempty"`
	ContainerRuntimeVersion string `json:"container_runtime_version,omitempty"`
}

// HeartbeatAuth contains authentication for a heartbeat
type HeartbeatAuth struct {
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// NodeRegistrationRequest is the registration payload
type NodeRegistrationRequest struct {
	NodeID              string        `json:"node_id"`
	ClusterID           string        `json:"cluster_id"`
	ProviderAddress     string        `json:"provider_address"`
	AgentPubkey         string        `json:"agent_pubkey"`
	Hostname            string        `json:"hostname,omitempty"`
	HardwareFingerprint string        `json:"hardware_fingerprint,omitempty"`
	AgentVersion        string        `json:"agent_version,omitempty"`
	Region              string        `json:"region,omitempty"`
	Datacenter          string        `json:"datacenter,omitempty"`
	Capacity            *NodeCapacity `json:"capacity,omitempty"`
	Health              *NodeHealth   `json:"health,omitempty"`
	Hardware            *NodeHardware `json:"hardware,omitempty"`
	Topology            *NodeTopology `json:"topology,omitempty"`
	Locality            *NodeLocality `json:"locality,omitempty"`
}

// HeartbeatResponse is the response to a heartbeat
type HeartbeatResponse struct {
	Accepted             bool                   `json:"accepted"`
	SequenceAck          uint64                 `json:"sequence_ack"`
	Timestamp            time.Time              `json:"timestamp"`
	NextHeartbeatSeconds int32                  `json:"next_heartbeat_seconds"`
	Commands             []NodeCommand          `json:"commands,omitempty"`
	ConfigUpdates        *HeartbeatConfigUpdate `json:"config_updates,omitempty"`
	Errors               []HeartbeatError       `json:"errors,omitempty"`
}

// NodeCommand represents a command for the node agent
type NodeCommand struct {
	CommandID  string            `json:"command_id"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Deadline   time.Time         `json:"deadline,omitempty"`
}

// HeartbeatConfigUpdate contains configuration updates
type HeartbeatConfigUpdate struct {
	SamplingIntervalSeconds int32    `json:"sampling_interval_seconds,omitempty"`
	LatencyProbeTargets     []string `json:"latency_probe_targets,omitempty"`
	MetricsRetentionHours   int32    `json:"metrics_retention_hours,omitempty"`
}

// HeartbeatError represents an error in heartbeat processing
type HeartbeatError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
