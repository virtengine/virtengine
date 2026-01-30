# HPC Node Agent

The HPC Node Agent is a lightweight daemon that runs on HPC compute nodes to report health, capacity, and latency metrics to the provider daemon for on-chain node metadata updates.

## Overview

The node agent implements VE-500: Node Registration and Heartbeat Pipeline, which enables:

- Automated node registration with cryptographic identity
- Periodic heartbeat with capacity, health, and latency metrics
- Signed payload submission to the provider daemon
- Stale node detection and automatic deactivation

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────┐
│   HPC Node Agent    │────▶│   Provider Daemon   │────▶│   Blockchain    │
│   (hpc-node-agent)  │     │   (HPC Aggregator)  │     │   (x/hpc)       │
└─────────────────────┘     └─────────────────────┘     └─────────────────┘
         │                           │                          │
    Signed                     Batch                    MsgUpdateNode
   Heartbeats                Submission                   Metadata
```

## Installation

### Building from Source

```bash
# From the repository root
go build -o hpc-node-agent ./cmd/hpc-node-agent

# Install to system path
sudo mv hpc-node-agent /usr/local/bin/
```

### Configuration

Create a configuration file at `/etc/virtengine/hpc-node-agent.yaml`:

```yaml
node-id: "node-001"
cluster-id: "hpc-cluster-1"
provider-address: "virtengine1provider..."
provider-daemon-url: "http://provider-daemon:8081"
heartbeat-interval: 30s
key-file: "/etc/virtengine/node-agent.key"
region: "us-east-1"
datacenter: "dc1"
latency-targets:
  - "node-002"
  - "node-003"
log-level: "info"
```

## Usage

### Initialize Node Keys

Before running the agent, initialize the Ed25519 key pair:

```bash
hpc-node-agent init --key-file /etc/virtengine/node-agent.key
```

This generates a key pair and outputs the public key. The public key must be registered with the provider for the node to be allowed.

### Register Node

Register the node with the provider daemon:

```bash
hpc-node-agent register \
  --node-id node-001 \
  --cluster-id hpc-cluster-1 \
  --provider-address virtengine1provider...
```

### Start the Agent

```bash
hpc-node-agent start \
  --config /etc/virtengine/hpc-node-agent.yaml
```

### Check Status

View current node metrics:

```bash
hpc-node-agent status
```

## Node States

Nodes transition through the following states:

| State | Description |
|-------|-------------|
| `pending` | Node registered but not yet active |
| `active` | Node is healthy and available |
| `stale` | Node has missed heartbeat timeout (default: 120s) |
| `draining` | Node is draining jobs before maintenance |
| `drained` | Node has drained all jobs |
| `offline` | Node has exceeded offline threshold (default: 300s) |
| `deregistered` | Node has been deregistered (terminal) |

### State Transitions

```
pending ──▶ active ──▶ stale ──▶ offline ──▶ deregistered
              │          │                         ▲
              │          └───────────────────────────┘
              ▼                                      
          draining ──▶ drained ──────────────────────┘
```

## Heartbeat Format

The agent sends heartbeats in the following JSON format:

```json
{
  "heartbeat": {
    "node_id": "node-001",
    "cluster_id": "hpc-cluster-1",
    "sequence_number": 42,
    "timestamp": "2024-01-15T10:30:00Z",
    "agent_version": "0.1.0",
    "capacity": {
      "cpu_cores_total": 64,
      "cpu_cores_available": 32,
      "memory_gb_total": 512,
      "memory_gb_available": 256,
      "gpus_total": 8,
      "gpus_available": 4,
      "gpu_type": "NVIDIA A100",
      "storage_gb_total": 4000,
      "storage_gb_available": 2000
    },
    "health": {
      "status": "healthy",
      "uptime_seconds": 86400,
      "load_average_1m": "0.500000",
      "cpu_utilization_percent": 45,
      "memory_utilization_percent": 50,
      "slurm_state": "idle"
    },
    "latency": {
      "measurements": [
        {
          "target_node_id": "node-002",
          "latency_us": 150,
          "packet_loss_percent": 0,
          "measured_at": "2024-01-15T10:29:55Z"
        }
      ],
      "gateway_latency_us": 1000,
      "avg_cluster_latency_us": 200
    },
    "jobs": {
      "running_count": 4,
      "pending_count": 2,
      "completed_24h": 100,
      "failed_24h": 1
    },
    "services": {
      "slurmd_running": true,
      "slurmd_version": "23.02.4",
      "munge_running": true,
      "container_runtime": "singularity",
      "container_runtime_version": "3.10.4"
    }
  },
  "auth": {
    "signature": "base64-ed25519-signature",
    "nonce": "base64-nonce",
    "timestamp": 1705315800
  }
}
```

## Security

### Node Identity

Each node has an Ed25519 key pair for signing heartbeats:

- **Private Key**: Stored securely on the node at `/etc/virtengine/node-agent.key`
- **Public Key**: Registered with the provider for allowlist verification

### Signature Verification

The provider daemon verifies:
1. The heartbeat signature matches the registered public key
2. The sequence number is greater than the last seen (prevents replay)
3. The timestamp is within acceptable bounds (±5 minutes)

### Hardware Fingerprint

For additional security, nodes can include a hardware fingerprint (SHA256 of hardware identifiers) to detect hardware changes.

## Monitoring

### Metrics

The agent exposes the following metrics:

| Metric | Description |
|--------|-------------|
| `total_heartbeats` | Total heartbeats sent |
| `failed_heartbeats` | Failed heartbeat submissions |
| `sequence_number` | Current sequence number |
| `uptime_seconds` | Agent uptime |

### Alerts

The provider daemon's heartbeat monitor generates alerts for:

- **stale**: Node missed heartbeat timeout
- **offline**: Node exceeded offline threshold
- **recovered**: Node recovered from stale/offline
- **anomaly**: Unusual heartbeat interval patterns
- **sequence_gap**: Missing sequence numbers

## Troubleshooting

### Common Issues

**1. Heartbeat rejected: "node not registered"**

Ensure the node is registered with the provider daemon before starting:

```bash
hpc-node-agent register --node-id node-001 ...
```

**2. Heartbeat rejected: "signature invalid"**

The public key registered with the provider doesn't match the node's key. Re-register the node with the correct public key.

**3. Heartbeat rejected: "stale sequence"**

The sequence number is not greater than the last seen. This can happen after agent restart. Wait for the next heartbeat or restart with a fresh sequence.

**4. Node marked stale/offline**

Check network connectivity to the provider daemon:

```bash
curl http://provider-daemon:8081/health
```

### Debug Mode

Run with debug logging:

```bash
hpc-node-agent start --log-level debug --config /etc/virtengine/hpc-node-agent.yaml
```

## Systemd Service

Create `/etc/systemd/system/hpc-node-agent.service`:

```ini
[Unit]
Description=VirtEngine HPC Node Agent
After=network.target slurmd.service

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/hpc-node-agent start --config /etc/virtengine/hpc-node-agent.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable hpc-node-agent
sudo systemctl start hpc-node-agent
```

## Related Documentation

- [HPC Module Overview](./hpc-module.md)
- [Provider Daemon](./provider-daemon.md)
- [SLURM Deployment](./slurm-deployment.md)
