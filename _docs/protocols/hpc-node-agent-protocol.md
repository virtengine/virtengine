# HPC Node Agent Protocol

Version: 1.0.0
Status: Draft
Last Updated: 2026-01-30

## Overview

This document defines the protocol for HPC node agents to communicate with the VirtEngine blockchain. Node agents are lightweight daemons running on compute nodes that report health, capacity, and metrics to enable intelligent job scheduling.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      VirtEngine Blockchain                       │
│                         (x/hpc module)                          │
├─────────────────────────────────────────────────────────────────┤
│    NodeMetadata Store    │    HeartbeatProcessor    │   Health  │
│    (per-node state)      │    (TTL management)      │  Monitor  │
└─────────────────────────────────────────────────────────────────┘
                                    ▲
                                    │ gRPC/REST
                                    │
┌─────────────────────────────────────────────────────────────────┐
│                     Provider Daemon (Gateway)                    │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐    │
│   │ Auth Proxy  │  │  Aggregator │  │  Chain Submitter    │    │
│   │ (mTLS/JWT)  │  │  (batching) │  │  (signed tx)        │    │
│   └─────────────┘  └─────────────┘  └─────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                    ▲
                         ┌──────────┴──────────┐
                         │                     │
              ┌──────────┴──────────┐   ┌──────┴───────┐
              │     Node Agent      │   │  Node Agent  │   ...
              │     (compute-01)    │   │  (compute-02)│
              └─────────────────────┘   └──────────────┘
```

## Node Agent Identity and Authentication

### Identity Model

Each node agent has a unique identity derived from:

1. **Node ID**: Unique identifier (e.g., `node-hpc1-compute-042`)
2. **Cluster ID**: Parent cluster identifier
3. **Provider Address**: Owning provider's blockchain address
4. **Agent Public Key**: Ed25519 public key for signing heartbeats

### Identity Registration

```json
{
  "node_identity": {
    "node_id": "string (required, unique within cluster, 4-64 chars)",
    "cluster_id": "string (required, parent cluster ID)",
    "provider_address": "string (required, bech32 provider address)",
    "agent_pubkey": "string (required, base64-encoded Ed25519 pubkey)",
    "hostname": "string (FQDN or short hostname)",
    "hardware_fingerprint": "string (SHA256 of hardware identifiers)",
    "registration_nonce": "string (base64, 32 bytes, one-time use)",
    "registration_timestamp": "timestamp (RFC3339)"
  },
  "provider_signature": "string (base64, provider signs the identity blob)"
}
```

### Authentication Flow

1. **Bootstrap**: Node agent generates Ed25519 keypair on first run
2. **Registration**: Provider signs node identity with provider key
3. **Heartbeat Auth**: Each heartbeat is signed with node's agent key
4. **Verification**: Provider daemon verifies signature before relaying

```go
// Signature verification for heartbeats
type HeartbeatAuth struct {
    NodeID        string `json:"node_id"`
    Timestamp     int64  `json:"timestamp"`      // Unix timestamp
    Nonce         string `json:"nonce"`          // Base64, 16 bytes
    Signature     string `json:"signature"`      // Base64 Ed25519 signature
    PublicKey     string `json:"public_key"`     // Base64 Ed25519 pubkey
}

// Message to sign: SHA256(node_id || timestamp || nonce || payload_hash)
```

### Key Storage

| Storage Type | Location | Use Case |
|-------------|----------|----------|
| File-based | `/etc/virtengine/node-agent/keys/` | Standard deployment |
| TPM 2.0 | TPM sealed storage | High-security nodes |
| HSM | PKCS#11 interface | Enterprise deployment |

## Heartbeat Protocol

### Heartbeat Payload Schema

```json
{
  "heartbeat": {
    "node_id": "string (required)",
    "cluster_id": "string (required)",
    "sequence_number": "uint64 (monotonically increasing)",
    "timestamp": "string (RFC3339 with nanoseconds)",
    "agent_version": "string (semver)",
    
    "capacity": {
      "cpu_cores_total": "int32",
      "cpu_cores_available": "int32",
      "cpu_cores_allocated": "int32",
      "memory_gb_total": "int32",
      "memory_gb_available": "int32",
      "memory_gb_allocated": "int32",
      "gpus_total": "int32",
      "gpus_available": "int32",
      "gpus_allocated": "int32",
      "gpu_type": "string",
      "storage_gb_total": "int32",
      "storage_gb_available": "int32",
      "storage_gb_allocated": "int32"
    },
    
    "health": {
      "status": "string (healthy|degraded|unhealthy|draining|offline)",
      "uptime_seconds": "int64",
      "load_average_1m": "string (fixed-point, 6 decimals)",
      "load_average_5m": "string (fixed-point, 6 decimals)",
      "load_average_15m": "string (fixed-point, 6 decimals)",
      "cpu_utilization_percent": "int32 (0-100)",
      "memory_utilization_percent": "int32 (0-100)",
      "gpu_utilization_percent": "int32 (0-100, optional)",
      "gpu_memory_utilization_percent": "int32 (0-100, optional)",
      "disk_io_utilization_percent": "int32 (0-100)",
      "network_utilization_percent": "int32 (0-100)",
      "temperature_celsius": "int32 (optional, CPU temp)",
      "gpu_temperature_celsius": "int32 (optional)",
      "error_count_24h": "int32",
      "warning_count_24h": "int32",
      "last_error_message": "string (optional, max 256 chars)",
      "slurm_state": "string (idle|allocated|mixed|down|drain|unknown)"
    },
    
    "latency": {
      "measurements": [
        {
          "target_node_id": "string",
          "latency_us": "int64 (microseconds)",
          "packet_loss_percent": "int32 (0-100)",
          "measured_at": "string (RFC3339)"
        }
      ],
      "gateway_latency_us": "int64 (to provider daemon)",
      "chain_latency_ms": "int64 (estimated, to chain)",
      "avg_cluster_latency_us": "int64"
    },
    
    "jobs": {
      "running_count": "int32",
      "pending_count": "int32",
      "completed_24h": "int32",
      "failed_24h": "int32",
      "active_job_ids": ["string (list of SLURM job IDs)"]
    },
    
    "services": {
      "slurmd_running": "bool",
      "slurmd_version": "string",
      "munge_running": "bool",
      "container_runtime": "string (singularity|docker|podman|none)",
      "container_runtime_version": "string"
    }
  },
  
  "auth": {
    "signature": "string (base64)",
    "nonce": "string (base64, 16 bytes)"
  }
}
```

### Heartbeat Response

```json
{
  "response": {
    "accepted": "bool",
    "sequence_ack": "uint64",
    "timestamp": "string (RFC3339)",
    "next_heartbeat_seconds": "int32 (suggested interval)",
    
    "commands": [
      {
        "command_id": "string",
        "type": "string (drain|resume|shutdown|update_agent|run_diagnostic)",
        "parameters": {},
        "deadline": "string (RFC3339)"
      }
    ],
    
    "config_updates": {
      "sampling_interval_seconds": "int32",
      "latency_probe_targets": ["string (node IDs to measure)"],
      "metrics_retention_hours": "int32"
    },
    
    "errors": [
      {
        "code": "string",
        "message": "string"
      }
    ]
  }
}
```

## TTL and Expiry Semantics

### Heartbeat TTL Configuration

| Parameter | Default | Range | Description |
|-----------|---------|-------|-------------|
| `heartbeat_interval` | 30s | 10s-120s | Normal heartbeat interval |
| `heartbeat_timeout` | 120s | 60s-600s | Time before node marked stale |
| `offline_threshold` | 300s | 120s-1800s | Time before node marked offline |
| `deregistration_delay` | 3600s | 600s-86400s | Time before automatic deregistration |

### Node State Transitions

```
                     ┌─────────────────────────────────────────┐
                     │                                         ▼
┌────────┐   register   ┌────────┐   heartbeat   ┌──────────┐
│ UNKNOWN├─────────────►│ PENDING├──────────────►│  ACTIVE  │
└────────┘               └────────┘               └────┬─────┘
                                                       │
                         ┌─────────────────────────────┤
                         │                             │
              timeout > TTL                   drain command
                         │                             │
                         ▼                             ▼
                   ┌─────────┐              ┌──────────────┐
                   │  STALE  │              │   DRAINING   │
                   └────┬────┘              └──────┬───────┘
                        │                          │
           timeout > offline_threshold      jobs complete
                        │                          │
                        ▼                          ▼
                   ┌──────────┐             ┌──────────┐
                   │ OFFLINE  │◄────────────│ DRAINED  │
                   └────┬─────┘             └──────────┘
                        │
           timeout > deregistration_delay
                        │
                        ▼
                 ┌──────────────┐
                 │ DEREGISTERED │
                 └──────────────┘
```

### Deactivation Rules

1. **Stale Detection**: 
   - No heartbeat for `heartbeat_timeout` → mark `STALE`
   - Node excluded from new job scheduling
   
2. **Offline Transition**:
   - No heartbeat for `offline_threshold` → mark `OFFLINE`
   - Running jobs flagged for potential failure
   - Alert sent to provider

3. **Automatic Deregistration**:
   - No heartbeat for `deregistration_delay` → mark `DEREGISTERED`
   - Node metadata retained for 30 days (configurable)
   - Jobs on this node marked as failed

4. **Manual Drain**:
   - Provider sends `drain` command
   - Node stops accepting new jobs
   - Existing jobs allowed to complete
   - After all jobs complete → `DRAINED` state

### Recovery Handling

```go
// Recovery after offline period
func (k Keeper) HandleNodeRecovery(ctx sdk.Context, nodeID string, heartbeat *Heartbeat) error {
    node, exists := k.GetNodeMetadata(ctx, nodeID)
    if !exists {
        return ErrNodeNotFound
    }
    
    // Validate recovery is allowed
    if node.State == NodeStateDeregistered {
        return ErrNodeDeregistered
    }
    
    // Check sequence number continuity
    if heartbeat.SequenceNumber <= node.LastSequence {
        return ErrStaleHeartbeat
    }
    
    // Calculate downtime
    downtime := heartbeat.Timestamp.Sub(node.LastHeartbeat)
    
    // Update node state
    node.State = NodeStateActive
    node.LastHeartbeat = heartbeat.Timestamp
    node.LastSequence = heartbeat.SequenceNumber
    node.RecoveryCount++
    node.TotalDowntimeSeconds += int64(downtime.Seconds())
    
    return k.SetNodeMetadata(ctx, node)
}
```

## On-Chain Metadata Mapping

### Heartbeat to NodeMetadata

| Heartbeat Field | NodeMetadata Field | Notes |
|-----------------|-------------------|-------|
| `node_id` | `NodeID` | Immutable |
| `cluster_id` | `ClusterID` | Immutable |
| `capacity.cpu_cores_available` | `Resources.CPUCores` | Available, not total |
| `capacity.memory_gb_available` | `Resources.MemoryGB` | Available, not total |
| `capacity.gpus_available` | `Resources.GPUs` | Available, not total |
| `capacity.gpu_type` | `Resources.GPUType` | - |
| `health.status` | `Active` | mapped to bool |
| `latency.avg_cluster_latency_us` | `AvgLatencyMs` | converted us→ms |
| `timestamp` | `LastHeartbeat` | - |
| `latency.measurements` | `LatencyMeasurements` | - |

### Aggregation for Cluster Metadata

The provider daemon aggregates node heartbeats to update cluster-level metadata:

```go
type ClusterAggregation struct {
    TotalNodes        int32
    ActiveNodes       int32
    AvailableCPUCores int64
    AvailableMemoryGB int64
    AvailableGPUs     int64
    AvgClusterLatency int64  // weighted average in ms
    HealthScore       string // fixed-point 0-1000000
}

// Update x/hpc cluster state from aggregation
func updateClusterFromNodes(cluster *types.HPCCluster, agg *ClusterAggregation) {
    cluster.TotalNodes = agg.TotalNodes
    cluster.AvailableNodes = agg.ActiveNodes
    cluster.ClusterMetadata.TotalCPUCores = agg.AvailableCPUCores
    cluster.ClusterMetadata.TotalMemoryGB = agg.AvailableMemoryGB
    cluster.ClusterMetadata.TotalGPUs = agg.AvailableGPUs
    cluster.UpdatedAt = time.Now()
}
```

## Metrics Sampling Intervals

### Minimum Sampling Intervals

| Metric Category | Min Interval | Recommended | Max Staleness |
|----------------|--------------|-------------|---------------|
| **Capacity** | 10s | 30s | 60s |
| **Health Status** | 10s | 30s | 60s |
| **CPU/Memory Utilization** | 5s | 15s | 30s |
| **GPU Metrics** | 5s | 15s | 30s |
| **Latency Measurements** | 60s | 300s | 600s |
| **Job Counts** | 30s | 60s | 120s |
| **Temperature** | 30s | 60s | 300s |

### Sampling Configuration

```json
{
  "sampling_config": {
    "base_interval_seconds": 30,
    "capacity_interval_seconds": 30,
    "health_interval_seconds": 30,
    "utilization_interval_seconds": 15,
    "latency_interval_seconds": 300,
    "latency_probe_count": 3,
    "latency_probe_timeout_ms": 1000,
    "metrics_buffer_size": 100,
    "batch_submit_interval_seconds": 60,
    "batch_max_size": 50
  }
}
```

### Adaptive Sampling

The agent adjusts sampling based on load:

```go
func (a *Agent) calculateSamplingInterval(baseInterval time.Duration) time.Duration {
    // Increase interval under high load
    if a.cpuUtilization > 90 {
        return baseInterval * 2
    }
    if a.cpuUtilization > 75 {
        return baseInterval + baseInterval/2
    }
    
    // Decrease interval during state changes
    if a.stateChanged {
        return baseInterval / 2
    }
    
    return baseInterval
}
```

## Example Payloads

### Healthy Node Heartbeat

```json
{
  "heartbeat": {
    "node_id": "node-hpc1-compute-042",
    "cluster_id": "hpc-cluster-1",
    "sequence_number": 145823,
    "timestamp": "2026-01-30T14:05:00.123456789Z",
    "agent_version": "0.9.1",
    "capacity": {
      "cpu_cores_total": 64,
      "cpu_cores_available": 32,
      "cpu_cores_allocated": 32,
      "memory_gb_total": 256,
      "memory_gb_available": 128,
      "memory_gb_allocated": 128,
      "gpus_total": 4,
      "gpus_available": 2,
      "gpus_allocated": 2,
      "gpu_type": "NVIDIA A100 80GB",
      "storage_gb_total": 2000,
      "storage_gb_available": 1500,
      "storage_gb_allocated": 500
    },
    "health": {
      "status": "healthy",
      "uptime_seconds": 2592000,
      "load_average_1m": "16500000",
      "load_average_5m": "15200000",
      "load_average_15m": "14800000",
      "cpu_utilization_percent": 52,
      "memory_utilization_percent": 48,
      "gpu_utilization_percent": 85,
      "gpu_memory_utilization_percent": 72,
      "disk_io_utilization_percent": 15,
      "network_utilization_percent": 23,
      "temperature_celsius": 62,
      "gpu_temperature_celsius": 71,
      "error_count_24h": 0,
      "warning_count_24h": 2,
      "slurm_state": "mixed"
    },
    "latency": {
      "measurements": [
        {
          "target_node_id": "node-hpc1-compute-041",
          "latency_us": 45,
          "packet_loss_percent": 0,
          "measured_at": "2026-01-30T14:04:55.000Z"
        },
        {
          "target_node_id": "node-hpc1-compute-043",
          "latency_us": 42,
          "packet_loss_percent": 0,
          "measured_at": "2026-01-30T14:04:55.000Z"
        }
      ],
      "gateway_latency_us": 1200,
      "chain_latency_ms": 150,
      "avg_cluster_latency_us": 44
    },
    "jobs": {
      "running_count": 2,
      "pending_count": 0,
      "completed_24h": 15,
      "failed_24h": 0,
      "active_job_ids": ["slurm-12345", "slurm-12346"]
    },
    "services": {
      "slurmd_running": true,
      "slurmd_version": "23.02.4",
      "munge_running": true,
      "container_runtime": "singularity",
      "container_runtime_version": "4.1.0"
    }
  },
  "auth": {
    "signature": "MEUCIQDKZo8x9B7fG...<base64>...",
    "nonce": "a1b2c3d4e5f6g7h8"
  }
}
```

### Degraded Node Heartbeat

```json
{
  "heartbeat": {
    "node_id": "node-hpc1-compute-099",
    "cluster_id": "hpc-cluster-1",
    "sequence_number": 89421,
    "timestamp": "2026-01-30T14:05:30.000000000Z",
    "agent_version": "0.9.1",
    "capacity": {
      "cpu_cores_total": 64,
      "cpu_cores_available": 0,
      "cpu_cores_allocated": 64,
      "memory_gb_total": 256,
      "memory_gb_available": 12,
      "memory_gb_allocated": 244,
      "gpus_total": 4,
      "gpus_available": 0,
      "gpus_allocated": 4,
      "gpu_type": "NVIDIA A100 80GB"
    },
    "health": {
      "status": "degraded",
      "uptime_seconds": 1296000,
      "load_average_1m": "72000000",
      "load_average_5m": "68000000",
      "load_average_15m": "65000000",
      "cpu_utilization_percent": 98,
      "memory_utilization_percent": 95,
      "gpu_utilization_percent": 100,
      "gpu_memory_utilization_percent": 98,
      "disk_io_utilization_percent": 85,
      "network_utilization_percent": 45,
      "temperature_celsius": 78,
      "gpu_temperature_celsius": 83,
      "error_count_24h": 3,
      "warning_count_24h": 12,
      "last_error_message": "GPU ECC error corrected on device 2",
      "slurm_state": "allocated"
    },
    "latency": {
      "measurements": [],
      "gateway_latency_us": 2500,
      "chain_latency_ms": 180,
      "avg_cluster_latency_us": 0
    },
    "jobs": {
      "running_count": 4,
      "pending_count": 0,
      "completed_24h": 8,
      "failed_24h": 1,
      "active_job_ids": ["slurm-12350", "slurm-12351", "slurm-12352", "slurm-12353"]
    },
    "services": {
      "slurmd_running": true,
      "slurmd_version": "23.02.4",
      "munge_running": true,
      "container_runtime": "singularity",
      "container_runtime_version": "4.1.0"
    }
  },
  "auth": {
    "signature": "MEQCIGjK2p9x...<base64>...",
    "nonce": "h8g7f6e5d4c3b2a1"
  }
}
```

## Validation Rules

### Heartbeat Validation

```go
func ValidateHeartbeat(hb *Heartbeat) error {
    // Required fields
    if hb.NodeID == "" {
        return errors.New("node_id required")
    }
    if hb.ClusterID == "" {
        return errors.New("cluster_id required")
    }
    if hb.SequenceNumber == 0 {
        return errors.New("sequence_number must be > 0")
    }
    
    // Timestamp validation (not too far in past or future)
    now := time.Now()
    if hb.Timestamp.Before(now.Add(-5 * time.Minute)) {
        return errors.New("timestamp too old")
    }
    if hb.Timestamp.After(now.Add(1 * time.Minute)) {
        return errors.New("timestamp in future")
    }
    
    // Capacity validation
    if err := validateCapacity(&hb.Capacity); err != nil {
        return fmt.Errorf("capacity: %w", err)
    }
    
    // Health validation
    if err := validateHealth(&hb.Health); err != nil {
        return fmt.Errorf("health: %w", err)
    }
    
    return nil
}

func validateCapacity(c *Capacity) error {
    if c.CPUCoresTotal < 1 {
        return errors.New("cpu_cores_total must be >= 1")
    }
    if c.CPUCoresAvailable < 0 || c.CPUCoresAvailable > c.CPUCoresTotal {
        return errors.New("cpu_cores_available out of range")
    }
    if c.MemoryGBTotal < 1 {
        return errors.New("memory_gb_total must be >= 1")
    }
    if c.MemoryGBAvailable < 0 || c.MemoryGBAvailable > c.MemoryGBTotal {
        return errors.New("memory_gb_available out of range")
    }
    return nil
}

func validateHealth(h *Health) error {
    validStatuses := map[string]bool{
        "healthy": true, "degraded": true, "unhealthy": true,
        "draining": true, "offline": true,
    }
    if !validStatuses[h.Status] {
        return fmt.Errorf("invalid status: %s", h.Status)
    }
    if h.CPUUtilizationPercent < 0 || h.CPUUtilizationPercent > 100 {
        return errors.New("cpu_utilization_percent out of range")
    }
    if h.MemoryUtilizationPercent < 0 || h.MemoryUtilizationPercent > 100 {
        return errors.New("memory_utilization_percent out of range")
    }
    return nil
}
```

### Authentication Validation

```go
func ValidateHeartbeatAuth(hb *Heartbeat, auth *HeartbeatAuth, knownPubkeys map[string]ed25519.PublicKey) error {
    // Get registered public key for node
    pubkey, exists := knownPubkeys[hb.NodeID]
    if !exists {
        return errors.New("unknown node_id")
    }
    
    // Verify timestamp freshness (prevent replay)
    authTime := time.Unix(auth.Timestamp, 0)
    if time.Since(authTime) > 5*time.Minute {
        return errors.New("auth timestamp expired")
    }
    
    // Construct message to verify
    payloadHash := sha256.Sum256(encodeHeartbeat(hb))
    message := fmt.Sprintf("%s|%d|%s|%x",
        auth.NodeID,
        auth.Timestamp,
        auth.Nonce,
        payloadHash[:],
    )
    
    // Verify signature
    sig, err := base64.StdEncoding.DecodeString(auth.Signature)
    if err != nil {
        return errors.New("invalid signature encoding")
    }
    
    if !ed25519.Verify(pubkey, []byte(message), sig) {
        return errors.New("signature verification failed")
    }
    
    return nil
}
```

## Node Onboarding and Bootstrap

### Prerequisites

1. Provider has registered an HPC cluster on-chain
2. Node has network connectivity to provider daemon
3. Node has SLURM worker (slurmd) installed and configured

### Bootstrap Steps

#### Step 1: Install Node Agent

```bash
# Download agent binary
curl -L https://releases.virtengine.dev/node-agent/v0.9.1/virtengine-node-agent-linux-amd64 \
  -o /usr/local/bin/virtengine-node-agent
chmod +x /usr/local/bin/virtengine-node-agent

# Create directories
mkdir -p /etc/virtengine/node-agent/keys
mkdir -p /var/lib/virtengine/node-agent
mkdir -p /var/log/virtengine
```

#### Step 2: Generate Node Identity

```bash
# Generate keypair (creates agent.key and agent.pub)
virtengine-node-agent keygen \
  --output-dir /etc/virtengine/node-agent/keys

# Display public key for registration
cat /etc/virtengine/node-agent/keys/agent.pub
```

#### Step 3: Request Registration Token

Provider creates a registration token:

```bash
# On provider system
virtengine-provider node-token create \
  --cluster-id hpc-cluster-1 \
  --node-id node-hpc1-compute-042 \
  --pubkey "$(cat agent.pub)" \
  --expires 24h
  
# Output: registration token (JWT)
```

#### Step 4: Configure Node Agent

```yaml
# /etc/virtengine/node-agent/config.yaml
node:
  id: "node-hpc1-compute-042"
  cluster_id: "hpc-cluster-1"
  
provider_daemon:
  endpoint: "https://provider.example.com:8443"
  tls:
    ca_cert: "/etc/virtengine/node-agent/ca.crt"
    client_cert: "/etc/virtengine/node-agent/client.crt"
    client_key: "/etc/virtengine/node-agent/client.key"

agent:
  key_path: "/etc/virtengine/node-agent/keys/agent.key"
  registration_token: "eyJhbGciOiJFZDI1NTE5..."

sampling:
  base_interval: "30s"
  utilization_interval: "15s"
  latency_interval: "5m"
  
logging:
  level: "info"
  file: "/var/log/virtengine/node-agent.log"
  max_size_mb: 100
  max_backups: 5
```

#### Step 5: Register and Start Agent

```bash
# Register node with provider (one-time)
virtengine-node-agent register \
  --config /etc/virtengine/node-agent/config.yaml

# Start agent service
systemctl enable virtengine-node-agent
systemctl start virtengine-node-agent

# Verify registration
virtengine-node-agent status
```

#### Step 6: Verify On-Chain Registration

```bash
# Query node metadata from chain
virtengine query hpc node node-hpc1-compute-042

# Expected output:
# node_id: node-hpc1-compute-042
# cluster_id: hpc-cluster-1
# state: active
# last_heartbeat: 2026-01-30T14:05:00Z
# ...
```

### Systemd Service Unit

```ini
# /etc/systemd/system/virtengine-node-agent.service
[Unit]
Description=VirtEngine HPC Node Agent
After=network-online.target slurmd.service
Wants=network-online.target

[Service]
Type=simple
User=virtengine
Group=virtengine
ExecStart=/usr/local/bin/virtengine-node-agent run \
  --config /etc/virtengine/node-agent/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=virtengine-node-agent

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/virtengine /var/log/virtengine
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

### Troubleshooting

| Issue | Check | Resolution |
|-------|-------|------------|
| Registration fails | Token expired | Generate new token |
| Heartbeat rejected | Clock skew | Sync NTP |
| Auth error | Key mismatch | Re-register with correct pubkey |
| No chain updates | Provider daemon offline | Check provider daemon logs |
| Node stuck in PENDING | Cluster not active | Activate cluster first |

## Related Documents

- [HPC Cluster Template Spec](./hpc-cluster-template-spec.md) - Cluster configuration schema
- [Provider Daemon Integration](../provider-daemon-waldur-integration.md) - Provider daemon guide
- [HPC Module Types](../../x/hpc/types/) - On-chain type definitions
