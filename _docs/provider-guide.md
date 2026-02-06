# VirtEngine Provider Guide

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-803

---

## Table of Contents

1. [Overview](#overview)
2. [Provider Daemon Setup](#provider-daemon-setup)
3. [Orchestration Adapters](#orchestration-adapters)
4. [Key Management](#key-management)
5. [Benchmarking](#benchmarking)
6. [Dispute Handling](#dispute-handling)
7. [Operational Best Practices](#operational-best-practices)

---

## Overview

VirtEngine providers offer compute resources through the decentralized marketplace. This guide covers:

- Setting up and configuring the Provider Daemon
- Connecting to orchestration platforms (Kubernetes, SLURM)
- Managing cryptographic keys
- Participating in benchmarking
- Handling disputes

### Provider Responsibilities

1. **Resource Provisioning**: Deploy customer workloads on allocated resources
2. **Usage Reporting**: Submit signed usage records to the chain
3. **Benchmarking**: Participate in performance verification
4. **Availability**: Maintain high uptime for allocated resources
5. **Security**: Protect customer data and workloads

## Provider Daemon Setup

### Prerequisites

- Ubuntu 22.04 LTS or equivalent
- Go 1.21+
- Access to VirtEngine node (RPC endpoint)
- Kubernetes cluster or SLURM installation
- Provider wallet with sufficient stake

### Installation

```bash
# Download provider daemon
wget https://github.com/virtengine/virtengine/releases/download/v1.0.0/provider-daemon_linux_amd64.tar.gz
tar -xzf provider-daemon_linux_amd64.tar.gz
sudo mv provider-daemon /usr/local/bin/

# Verify installation
provider-daemon version
```

### Configuration

Create configuration file at `~/.provider-daemon/config.yaml`:

```yaml
# Provider identity
provider:
  # On-chain address
  address: "virtengine1provider..."
  # Display name
  name: "MyProvider"
  # Contact info
  contact_email: "ops@myprovider.com"

# Chain connection
chain:
  node_url: "http://localhost:26657"
  grpc_url: "localhost:9090"
  chain_id: "virtengine-1"
  gas_prices: "0.025uve"
  gas_adjustment: 1.5

# Key management
keys:
  # Path to keyring
  keyring_path: "/home/provider/.provider-daemon/keyring"
  # Keyring backend: file, os, test
  keyring_backend: "file"
  # Key name for signing transactions
  signing_key: "provider"
  # Key name for encryption operations
  encryption_key: "provider-encryption"

# Orchestration
orchestration:
  # Adapter type: kubernetes, slurm
  adapter: "kubernetes"
  # Kubernetes-specific config
  kubernetes:
    kubeconfig: "/home/provider/.kube/config"
    namespace: "virtengine-workloads"
    resource_quota:
      cpu: "100"
      memory: "512Gi"
      gpu: "8"
  # SLURM-specific config (if using SLURM)
  slurm:
    controller_host: "slurm-controller.local"
    partition: "compute"
    account: "virtengine"

# Bid engine
bidding:
  # Enable automatic bidding
  enabled: true
  # Pricing strategy: fixed, dynamic
  strategy: "dynamic"
  # Base prices (per hour)
  base_prices:
    cpu_core: 10      # uve per CPU core
    memory_gb: 5      # uve per GB RAM
    gpu_unit: 500     # uve per GPU
    storage_gb: 1     # uve per GB storage
  # Price multipliers
  multipliers:
    peak_hours: 1.5   # 9am-5pm multiplier
    gpu_scarcity: 2.0 # When GPU utilization > 80%

# Workload management
workloads:
  # Maximum concurrent workloads
  max_concurrent: 50
  # Health check interval
  health_check_interval: "30s"
  # Eviction policy for failed workloads
  eviction_timeout: "10m"

# Usage reporting
usage:
  # Reporting interval
  report_interval: "1h"
  # Batch size for on-chain submission
  batch_size: 100
  # Retry configuration
  max_retries: 5
  retry_delay: "30s"

# Logging
logging:
  level: "info"
  format: "json"
  # IMPORTANT: Enable redaction for sensitive data
  redact_sensitive: true

# Metrics
metrics:
  enabled: true
  port: 9091
  path: "/metrics"
```

### Starting the Daemon

```bash
# Start in foreground
provider-daemon start --config ~/.provider-daemon/config.yaml

# Or as systemd service
sudo systemctl enable provider-daemon
sudo systemctl start provider-daemon
```

### Registering as Provider

```bash
# Create provider registration transaction
virtengine tx provider create \
    --name "MyProvider" \
    --contact "ops@myprovider.com" \
    --website "https://myprovider.com" \
    --deposit 100000000000uve \
    --from provider \
    --keyring-backend file

# Register encryption key for receiving encrypted order details
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(provider-daemon keys show provider-encryption --pubkey) \
    --label "Provider Order Decryption Key" \
    --from provider
```

## Orchestration Adapters

### Kubernetes Adapter

The Kubernetes adapter provisions workloads as Kubernetes Deployments.

#### Configuration

```yaml
orchestration:
  adapter: "kubernetes"
  kubernetes:
    kubeconfig: "/path/to/kubeconfig"
    namespace: "virtengine-workloads"
    # Namespace labels for isolation
    namespace_labels:
      "virtengine.com/managed": "true"
    # Default resource limits
    default_limits:
      cpu: "4"
      memory: "8Gi"
    # Storage class for persistent volumes
    storage_class: "fast-ssd"
    # Enable network policies
    network_policies: true
    # Pod security context
    security_context:
      run_as_non_root: true
      read_only_root_filesystem: true
```

#### Required RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: provider-daemon
rules:
  - apiGroups: [""]
    resources: ["namespaces", "pods", "services", "configmaps", "secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["networkpolicies", "ingresses"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
```

### SLURM Adapter

The SLURM adapter submits jobs to a SLURM cluster for HPC workloads.

#### Configuration

```yaml
orchestration:
  adapter: "slurm"
  slurm:
    controller_host: "slurm-controller.local"
    controller_port: 6817
    partition: "compute"
    account: "virtengine"
    # SSH configuration for job submission
    ssh:
      user: "slurm-submit"
      key_path: "/home/provider/.ssh/slurm_key"
    # Resource mapping
    resource_mapping:
      cpu_to_cores: 1      # 1 VE CPU = 1 SLURM core
      memory_to_mb: 1024   # 1 VE memory unit = 1024 MB
      gpu_to_gres: "gpu:1" # 1 VE GPU = 1 SLURM GPU
    # Job defaults
    defaults:
      time_limit: "24:00:00"
      output_dir: "/scratch/virtengine/%j"
```

#### SLURM Script Template

```bash
#!/bin/bash
#SBATCH --job-name=ve-{{.JobID}}
#SBATCH --partition={{.Partition}}
#SBATCH --account={{.Account}}
#SBATCH --nodes={{.Nodes}}
#SBATCH --ntasks={{.Tasks}}
#SBATCH --cpus-per-task={{.CPUsPerTask}}
#SBATCH --mem={{.Memory}}
#SBATCH --time={{.TimeLimit}}
#SBATCH --output={{.OutputDir}}/output.log
#SBATCH --error={{.OutputDir}}/error.log

# Load required modules
module load singularity

# Run containerized workload
singularity exec {{.Container}} {{.Command}}
```

### SLURM-on-Kubernetes Bootstrap (Provider Daemon)

For SLURM clusters deployed on Kubernetes, the provider daemon can bootstrap a minimal SLURM stack and wire node agents
to the on-chain node lifecycle. This flow uses `helm` and `kubectl` CLIs on the provider host.

#### Configuration

```yaml
hpc:
  enabled: true
  scheduler_type: "slurm"
  cluster_id: "HPC-1"
  provider_address: "virtengine1provider..."
  slurm_k8s:
    enabled: true
    bootstrap_on_start: true
    namespace: "slurm-system"
    helm_chart_path: "/opt/virtengine/charts/slurm"
    helm_release_name: "slurm-hpc-1"
    ready_timeout: "15m"
    rollback_on_failure: true
    min_compute_ready: 1
    provider_endpoint: "http://provider-daemon:8081"
    helm:
      binary: "helm"
      kubeconfig: "/home/provider/.kube/config"
    kube:
      binary: "kubectl"
      kubeconfig: "/home/provider/.kube/config"
  node_aggregator:
    enabled: true
    listen_addr: ":8081"
    provider_address: "virtengine1provider..."
    cluster_id: "HPC-1"
    heartbeat_timeout: "2m"
    checkpoint_file: "/var/lib/virtengine/hpc-node-checkpoint.json"
    chain_submit_enabled: true
    max_submit_retries: 5
    retry_backoff: "5s"
    stale_miss_threshold: 5
    default_region: "us-west-1"
    default_datacenter: "pdx-1"
```

#### Operational Notes

- `slurm_k8s.provider_endpoint` must be reachable from the SLURM node agents (often the provider daemon service).
- The node aggregator persists checkpoints so restarts do not lose heartbeat sequences or pending chain updates.
- If bootstrap readiness fails and `rollback_on_failure` is enabled, the Helm release is uninstalled.

## Key Management

### Key Types

| Key Type | Purpose | Rotation Frequency |
|----------|---------|-------------------|
| Signing Key | Transaction signing | Yearly or on compromise |
| Encryption Key | Decrypting order details | Yearly |
| TLS Key | API/webhook authentication | Yearly |

### Key Generation

```bash
# Generate signing key
provider-daemon keys add provider --keyring-backend file

# Generate encryption key
provider-daemon keys add provider-encryption --keyring-backend file

# Generate TLS certificates
provider-daemon tls generate --output ~/.provider-daemon/tls/
```

### Key Rotation

```bash
# 1. Generate new key
provider-daemon keys add provider-v2 --keyring-backend file

# 2. Register new encryption key on-chain
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(provider-daemon keys show provider-encryption-v2 --pubkey) \
    --label "Provider Encryption Key v2" \
    --from provider

# 3. Update daemon config to use new key
# Edit config.yaml: encryption_key: "provider-encryption-v2"

# 4. Restart daemon
sudo systemctl restart provider-daemon

# 5. After grace period (24-48 hours), revoke old key
virtengine tx encryption revoke-recipient-key \
    --fingerprint <old_key_fingerprint> \
    --from provider
```

### Backup Procedures

> ⚠️ **CRITICAL**: Always backup keys before rotation

```bash
# Backup keyring
tar -czf keyring-backup-$(date +%Y%m%d).tar.gz ~/.provider-daemon/keyring/

# Store backup securely (encrypted, offsite)
gpg --encrypt --recipient backup@myprovider.com keyring-backup-*.tar.gz
```

## Benchmarking

### Automatic Benchmarking

The provider daemon automatically participates in benchmarking:

```yaml
benchmarking:
  enabled: true
  # Run benchmarks at these intervals
  schedule: "0 */6 * * *"  # Every 6 hours
  # Benchmark types to run
  types:
    - compute
    - network
    - storage
  # Submit results to chain
  auto_submit: true
```

### Manual Benchmarking

```bash
# Run compute benchmark
provider-daemon benchmark run --type compute

# Run network benchmark
provider-daemon benchmark run --type network

# Run storage benchmark
provider-daemon benchmark run --type storage

# Run all benchmarks
provider-daemon benchmark run --all

# Submit results to chain
provider-daemon benchmark submit --latest
```

### Benchmark Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| CPU Score | Compute performance | > 1000 |
| Memory Bandwidth | GB/s throughput | > 50 |
| Network Latency | P99 latency to reference | < 50ms |
| Network Throughput | Gbps | > 1 |
| Storage IOPS | Random read/write | > 10000 |
| Storage Throughput | Sequential MB/s | > 500 |

### Anti-Gaming Protection

VirtEngine uses several mechanisms to prevent benchmark gaming:

1. **Random Challenges**: Benchmarks include random workloads
2. **Cross-Validation**: Results verified by multiple validators
3. **Anomaly Detection**: Sudden changes trigger investigation
4. **Real-World Correlation**: Benchmarks compared to actual usage

## Dispute Handling

### Dispute Types

| Type | Description | Resolution |
|------|-------------|------------|
| Usage Disagreement | Customer disputes usage records | Evidence review |
| SLA Violation | Provider failed availability commitment | Automatic penalty |
| Resource Mismatch | Actual resources differ from offering | Arbitration |
| Malicious Workload | Customer ran prohibited workload | Evidence review |

### Responding to Disputes

```bash
# View pending disputes
virtengine query disputes list --provider $(virtengine keys show provider -a)

# View dispute details
virtengine query disputes show dispute_123

# Submit evidence
virtengine tx disputes submit-evidence \
    --dispute dispute_123 \
    --evidence-file evidence.json \
    --from provider

# Accept dispute (if at fault)
virtengine tx disputes accept \
    --dispute dispute_123 \
    --from provider
```

### Evidence Format

```json
{
  "dispute_id": "dispute_123",
  "provider_response": "Usage records are accurate",
  "evidence": [
    {
      "type": "usage_log",
      "description": "Raw usage logs from orchestrator",
      "hash": "abc123...",
      "url": "https://evidence.myprovider.com/logs/12345"
    },
    {
      "type": "metric_snapshot",
      "description": "Prometheus metrics at time of dispute",
      "hash": "def456...",
      "data": {...}
    }
  ],
  "submitted_at": "2026-01-24T12:00:00Z"
}
```

## Operational Best Practices

### Monitoring

```yaml
# Prometheus scrape config
scrape_configs:
  - job_name: 'provider-daemon'
    static_configs:
      - targets: ['localhost:9091']
```

#### Key Metrics

| Metric | Alert Threshold |
|--------|----------------|
| `provider_workloads_active` | > max_concurrent |
| `provider_bid_success_rate` | < 0.5 |
| `provider_usage_submission_errors` | > 0 for 15m |
| `provider_orchestration_errors` | > 0 for 5m |

### Logging

```yaml
logging:
  level: "info"
  format: "json"
  output: "/var/log/provider-daemon/daemon.log"
  rotation:
    max_size_mb: 100
    max_backups: 10
    max_age_days: 30
  # CRITICAL: Always enable redaction
  redact_sensitive: true
  redact_patterns:
    - "password"
    - "secret"
    - "key"
    - "token"
```

### High Availability

For production deployments:

1. **Multiple Daemon Instances**: Run active-passive for failover
2. **Health Checks**: Kubernetes liveness/readiness probes
3. **Automatic Restart**: Systemd restart policies
4. **State Persistence**: External state store (etcd/Redis)

```yaml
# docker-compose.yml for HA setup
services:
  provider-daemon-primary:
    image: virtengine/provider-daemon:v1.0.0
    volumes:
      - ./config.yaml:/config.yaml
    command: start --config /config.yaml --leader-election

  provider-daemon-secondary:
    image: virtengine/provider-daemon:v1.0.0
    volumes:
      - ./config.yaml:/config.yaml
    command: start --config /config.yaml --leader-election
```

### Capacity Planning

| Metric | Warning | Critical |
|--------|---------|----------|
| CPU Utilization | > 70% | > 90% |
| Memory Utilization | > 75% | > 90% |
| GPU Utilization | > 80% | > 95% |
| Storage Utilization | > 70% | > 85% |

### Incident Response

1. **Detection**: Automated alerts trigger on-call
2. **Triage**: Assess impact and urgency
3. **Communication**: Notify affected customers
4. **Resolution**: Fix issue and restore service
5. **Post-Mortem**: Document and improve

```bash
# Emergency: Pause new allocations
virtengine tx provider set-status \
    --status PAUSED \
    --reason "Emergency maintenance" \
    --from provider

# Resume operations
virtengine tx provider set-status \
    --status ACTIVE \
    --from provider
```

---

## Support

- Provider Forum: [providers.virtengine.com](https://providers.virtengine.com)
- Discord: [#providers](https://discord.gg/virtengine)
- Emergency: providers-emergency@virtengine.com
