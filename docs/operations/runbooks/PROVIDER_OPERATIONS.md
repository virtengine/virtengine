# Provider Daemon Operational Procedures

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Overview](#overview)
2. [Initial Setup](#initial-setup)
3. [Orchestration Configuration](#orchestration-configuration)
4. [Bid Engine Operations](#bid-engine-operations)
5. [Workload Management](#workload-management)
6. [Usage Reporting](#usage-reporting)
7. [Daily Operations](#daily-operations)
8. [Maintenance Procedures](#maintenance-procedures)
9. [Security Operations](#security-operations)
10. [High Availability Setup](#high-availability-setup)

---

## Overview

The Provider Daemon bridges on-chain marketplace orders to infrastructure provisioning. It handles:

- **Bid Engine**: Automatically bids on marketplace orders
- **Workload Provisioning**: Deploys workloads via Kubernetes/SLURM/VMware
- **Usage Metering**: Collects and submits usage records
- **Health Monitoring**: Tracks workload health and manages lifecycle

### Provider Lifecycle

```
Registration → Active → Bidding → Allocation → Deployment → Usage Reporting → Settlement
```

---

## Initial Setup

### Prerequisites

```bash
# System requirements
- Ubuntu 22.04 LTS (or equivalent)
- Go 1.21+
- Access to VirtEngine RPC/gRPC endpoints
- Kubernetes cluster or SLURM installation
- Provider wallet with minimum stake (100,000 UVE)
```

### Step 1: Install Provider Daemon

```bash
# Download pre-built binary
RELEASE_VERSION="v1.0.0"
wget https://github.com/virtengine/virtengine/releases/download/${RELEASE_VERSION}/provider-daemon_linux_amd64.tar.gz
tar -xzf provider-daemon_linux_amd64.tar.gz
sudo mv provider-daemon /usr/local/bin/

# Verify installation
provider-daemon version
```

### Step 2: Create Configuration

Create `~/.provider-daemon/config.yaml`:

```yaml
# Provider identity
provider:
  address: "virtengine1provider..."
  name: "MyProvider"
  contact_email: "ops@myprovider.com"
  website: "https://myprovider.com"
  region: "us-east-1"

# Chain connection
chain:
  node_url: "http://localhost:26657"
  grpc_url: "localhost:9090"
  chain_id: "virtengine-1"
  gas_prices: "0.025uve"
  gas_adjustment: 1.5
  broadcast_mode: "sync"
  timeout: "30s"

# Key management
keys:
  keyring_path: "/home/provider/.provider-daemon/keyring"
  keyring_backend: "file"
  signing_key: "provider"
  encryption_key: "provider-encryption"

# Server settings
server:
  address: "0.0.0.0:8443"
  tls:
    enabled: true
    cert_file: "/home/provider/.provider-daemon/tls/server.crt"
    key_file: "/home/provider/.provider-daemon/tls/server.key"

# Orchestration (see next section for details)
orchestration:
  adapter: "kubernetes"
  
# Bid engine
bidding:
  enabled: true
  strategy: "dynamic"
  max_bids_per_minute: 10
  
# Workload management
workloads:
  max_concurrent: 100
  health_check_interval: "30s"
  eviction_timeout: "10m"
  
# Usage reporting
usage:
  report_interval: "1h"
  batch_size: 100
  
# Logging
logging:
  level: "info"
  format: "json"
  output: "/var/log/provider-daemon/daemon.log"
  redact_sensitive: true

# Metrics
metrics:
  enabled: true
  address: "0.0.0.0:9091"
  path: "/metrics"
```

### Step 3: Create Keys

```bash
# Create signing key
provider-daemon keys add provider --keyring-backend file

# Create encryption key
provider-daemon keys add provider-encryption --keyring-backend file

# Generate TLS certificates
provider-daemon tls generate \
    --output ~/.provider-daemon/tls/ \
    --hostname provider.mydomain.com

# Backup keys immediately
tar -czf provider-keys-backup-$(date +%Y%m%d).tar.gz ~/.provider-daemon/keyring/
# Store backup securely offsite
```

### Step 4: Register Provider

```bash
# Register on-chain
virtengine tx provider create \
    --name "MyProvider" \
    --contact "ops@myprovider.com" \
    --website "https://myprovider.com" \
    --deposit 100000000000uve \
    --from provider \
    --keyring-backend file

# Register encryption key
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(provider-daemon keys show provider-encryption --pubkey) \
    --label "Provider Order Decryption Key v1" \
    --from provider

# Verify registration
virtengine query provider info $(virtengine keys show provider -a)
```

### Step 5: Create Systemd Service

```ini
# /etc/systemd/system/provider-daemon.service
[Unit]
Description=VirtEngine Provider Daemon
After=network.target

[Service]
Type=simple
User=provider
Group=provider
ExecStart=/usr/local/bin/provider-daemon start --config /home/provider/.provider-daemon/config.yaml
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

# Environment
Environment="HOME=/home/provider"

# Security
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable provider-daemon
sudo systemctl start provider-daemon

# Check status
sudo systemctl status provider-daemon

# Check health endpoint
curl -k https://localhost:8443/health
```

---

## Orchestration Configuration

### Kubernetes Adapter

```yaml
orchestration:
  adapter: "kubernetes"
  kubernetes:
    kubeconfig: "/home/provider/.kube/config"
    namespace: "virtengine-workloads"
    
    # Namespace configuration
    namespace_labels:
      "virtengine.com/managed": "true"
      "virtengine.com/provider": "provider-1"
    
    # Resource quotas
    resource_quota:
      cpu: "100"
      memory: "512Gi"
      gpu: "8"
      storage: "10Ti"
    
    # Default limits per workload
    default_limits:
      cpu: "4"
      memory: "8Gi"
    
    # Network policies
    network_policies: true
    
    # Storage
    storage_class: "fast-ssd"
    
    # Image registry
    image_pull_secrets:
      - "virtengine-registry"
    
    # Security
    security_context:
      run_as_non_root: true
      read_only_root_filesystem: true
      allow_privilege_escalation: false
```

**Required Kubernetes RBAC:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: provider-daemon
rules:
  - apiGroups: [""]
    resources: ["namespaces", "pods", "services", "configmaps", "secrets", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["networkpolicies", "ingresses"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["pods/log", "pods/exec"]
    verbs: ["get", "create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: provider-daemon
subjects:
  - kind: ServiceAccount
    name: provider-daemon
    namespace: virtengine-system
roleRef:
  kind: ClusterRole
  name: provider-daemon
  apiGroup: rbac.authorization.k8s.io
```

### SLURM Adapter

```yaml
orchestration:
  adapter: "slurm"
  slurm:
    controller_host: "slurm-controller.local"
    controller_port: 6817
    partition: "compute"
    account: "virtengine"
    
    # SSH for job submission
    ssh:
      user: "slurm-submit"
      key_path: "/home/provider/.ssh/slurm_key"
      timeout: "30s"
    
    # Resource mapping
    resource_mapping:
      cpu_to_cores: 1
      memory_to_mb: 1024
      gpu_to_gres: "gpu:1"
    
    # Job defaults
    defaults:
      time_limit: "24:00:00"
      output_dir: "/scratch/virtengine/%j"
      qos: "normal"
    
    # Container runtime
    container_runtime: "singularity"
    singularity_images_dir: "/shared/images"
```

### OpenStack/Waldur Adapter

```yaml
orchestration:
  adapter: "openstack"
  openstack:
    waldur_url: "https://waldur.example.com"
    waldur_token: "${WALDUR_TOKEN}"
    project_uuid: "xxx-xxx-xxx"
    
    # Default settings
    defaults:
      flavor: "m1.medium"
      image: "ubuntu-22.04"
      network: "provider-network"
      security_group: "virtengine-workloads"
    
    # SSH keypair for access
    ssh_keypair: "virtengine-provider"
```

### VMware Adapter

```yaml
orchestration:
  adapter: "vmware"
  vmware:
    vcenter_url: "vcenter.local"
    username: "${VCENTER_USER}"
    password: "${VCENTER_PASSWORD}"
    datacenter: "DC1"
    cluster: "Compute-Cluster"
    resource_pool: "VirtEngine"
    datastore: "Fast-Storage"
    
    # VM defaults
    defaults:
      template: "ubuntu-22.04-template"
      network: "VM Network"
      folder: "VirtEngine-Workloads"
```

---

## Bid Engine Operations

### Configuration

```yaml
bidding:
  enabled: true
  
  # Bidding strategy
  strategy: "dynamic"  # fixed, dynamic, or custom
  
  # Rate limiting
  max_bids_per_minute: 10
  max_concurrent_bids: 5
  
  # Pricing (per hour in uve)
  base_prices:
    cpu_core: 10
    memory_gb: 5
    gpu_unit: 500
    storage_gb: 1
    bandwidth_gb: 0.5
  
  # Dynamic pricing multipliers
  multipliers:
    peak_hours:        # 9am-5pm local time
      enabled: true
      factor: 1.5
    gpu_scarcity:      # When GPU > 80% utilized
      enabled: true
      threshold: 0.8
      factor: 2.0
    low_utilization:   # Discount when < 30% utilized
      enabled: true
      threshold: 0.3
      factor: 0.8
  
  # Order filtering
  filters:
    min_duration: "1h"
    max_duration: "720h"
    min_value: 1000000  # uve
    allowed_regions:
      - "us-east"
      - "us-west"
    blocked_addresses: []
    
  # Auto-accept settings
  auto_accept:
    enabled: true
    max_utilization: 0.9  # Stop accepting at 90% capacity
```

### Manual Bid Operations

```bash
# View current bids
provider-daemon bids list

# View bid history
provider-daemon bids history --limit 50

# Cancel a pending bid
provider-daemon bids cancel --bid-id bid_123

# Manually create a bid
provider-daemon bids create \
    --order-id order_456 \
    --price 1000000uve \
    --deposit 100000uve

# View bid engine status
provider-daemon bids status
```

### Bid Engine Metrics

| Metric | Description |
|--------|-------------|
| `provider_bids_submitted_total` | Total bids submitted |
| `provider_bids_won_total` | Bids that won allocation |
| `provider_bids_lost_total` | Bids that lost |
| `provider_bid_win_rate` | Rolling win rate |
| `provider_bid_latency_seconds` | Time to submit bid |

---

## Workload Management

### Workload States

```
Pending → Deploying → Running → Stopping → Stopped → Terminated
              ↓          ↓         ↓
           Failed     Failed    Failed
              ↓          ↓
           Paused    Paused
```

### Managing Workloads

```bash
# List all workloads
provider-daemon workloads list

# List by state
provider-daemon workloads list --state running
provider-daemon workloads list --state failed

# View workload details
provider-daemon workloads show workload_123

# View workload logs
provider-daemon workloads logs workload_123 --tail 100

# Force stop a workload
provider-daemon workloads stop workload_123 --force

# Restart a workload
provider-daemon workloads restart workload_123
```

### Health Checks

```yaml
workloads:
  health_check:
    interval: "30s"
    timeout: "10s"
    retries: 3
    
  # Auto-healing
  auto_heal:
    enabled: true
    max_restarts: 5
    restart_delay: "30s"
    
  # Eviction
  eviction:
    enabled: true
    timeout: "10m"
    reasons:
      - "resource_exhaustion"
      - "repeated_failures"
      - "lease_expired"
```

### Resource Monitoring

```bash
# View resource utilization
provider-daemon resources status

# Sample output:
# Resource     Allocated   Available   Utilization
# CPU          64 cores    36 cores    64%
# Memory       256 GB      144 GB      64%
# GPU          4 units     4 units     50%
# Storage      2 TB        8 TB        20%
```

---

## Usage Reporting

### Configuration

```yaml
usage:
  # Reporting interval
  report_interval: "1h"
  
  # Batch size for on-chain submission
  batch_size: 100
  
  # Retry configuration
  max_retries: 5
  retry_delay: "30s"
  retry_backoff: 2.0
  
  # Local storage for failed submissions
  failed_reports_dir: "/var/lib/provider-daemon/failed_reports"
  
  # Metrics collection
  metrics_sources:
    - "prometheus"
    - "kubernetes"
```

### Manual Usage Operations

```bash
# View pending usage reports
provider-daemon usage pending

# View submitted reports
provider-daemon usage history --limit 50

# Force submit pending reports
provider-daemon usage submit --force

# Export usage data
provider-daemon usage export \
    --start "2026-01-01T00:00:00Z" \
    --end "2026-01-31T23:59:59Z" \
    --format csv \
    --output usage-report.csv
```

### Usage Metrics Format

```json
{
  "lease_id": "lease_123",
  "provider": "virtengine1provider...",
  "period_start": "2026-01-30T00:00:00Z",
  "period_end": "2026-01-30T01:00:00Z",
  "resources": {
    "cpu_milli_seconds": 3600000,
    "memory_byte_seconds": 8589934592000,
    "storage_byte_seconds": 107374182400000,
    "gpu_seconds": 3600,
    "network_bytes_in": 1073741824,
    "network_bytes_out": 2147483648
  },
  "signature": "..."
}
```

---

## Daily Operations

### Daily Checklist

- [ ] Check daemon health: `curl -k https://localhost:8443/health`
- [ ] Verify bid engine active: `provider-daemon bids status`
- [ ] Check workload status: `provider-daemon workloads list --state failed`
- [ ] Verify usage submission: `provider-daemon usage pending`
- [ ] Check resource utilization: `provider-daemon resources status`
- [ ] Review error logs

### Daily Commands Script

```bash
#!/bin/bash
# daily-check.sh

echo "=== Provider Daemon Daily Check ==="
echo ""

echo "1. Health Check"
curl -sk https://localhost:8443/health | jq
echo ""

echo "2. Bid Engine Status"
provider-daemon bids status
echo ""

echo "3. Failed Workloads"
provider-daemon workloads list --state failed
echo ""

echo "4. Pending Usage Reports"
provider-daemon usage pending | wc -l
echo ""

echo "5. Resource Utilization"
provider-daemon resources status
echo ""

echo "6. Recent Errors (last 24h)"
journalctl -u provider-daemon --since "24 hours ago" | grep -i error | tail -10
echo ""

echo "7. Provider On-Chain Status"
virtengine query provider info $(virtengine keys show provider -a) | head -20
```

### Weekly Checklist

- [ ] Review bid win rate and adjust pricing if needed
- [ ] Check for failed usage submissions and retry
- [ ] Review capacity trends
- [ ] Verify key rotation schedule
- [ ] Check for daemon updates
- [ ] Review and resolve any pending disputes

---

## Maintenance Procedures

### Graceful Shutdown

```bash
# 1. Pause new allocations
virtengine tx provider set-status \
    --status PAUSED \
    --reason "Scheduled maintenance" \
    --from provider

# 2. Wait for existing workloads to complete or notify customers
provider-daemon workloads list --state running

# 3. Stop the daemon
sudo systemctl stop provider-daemon

# 4. Perform maintenance

# 5. Restart
sudo systemctl start provider-daemon

# 6. Resume operations
virtengine tx provider set-status \
    --status ACTIVE \
    --from provider
```

### Updating Provider Daemon

```bash
# 1. Pause operations
virtengine tx provider set-status --status PAUSED --from provider

# 2. Download new version
wget https://github.com/virtengine/virtengine/releases/download/v1.1.0/provider-daemon_linux_amd64.tar.gz
tar -xzf provider-daemon_linux_amd64.tar.gz

# 3. Stop daemon
sudo systemctl stop provider-daemon

# 4. Backup current binary
sudo cp /usr/local/bin/provider-daemon /usr/local/bin/provider-daemon.bak

# 5. Install new binary
sudo mv provider-daemon /usr/local/bin/

# 6. Verify version
provider-daemon version

# 7. Start daemon
sudo systemctl start provider-daemon

# 8. Verify health
curl -k https://localhost:8443/health

# 9. Resume operations
virtengine tx provider set-status --status ACTIVE --from provider
```

### Key Rotation

```bash
# 1. Generate new encryption key
provider-daemon keys add provider-encryption-v2 --keyring-backend file

# 2. Register new key on-chain
virtengine tx encryption register-recipient-key \
    --algorithm X25519-XSalsa20-Poly1305 \
    --public-key $(provider-daemon keys show provider-encryption-v2 --pubkey) \
    --label "Provider Encryption Key v2" \
    --from provider

# 3. Wait grace period (48 hours) for new key propagation

# 4. Update config
# Change encryption_key: "provider-encryption-v2" in config.yaml

# 5. Restart daemon
sudo systemctl restart provider-daemon

# 6. Verify new key is active
provider-daemon keys list

# 7. Revoke old key (after grace period)
virtengine tx encryption revoke-recipient-key \
    --fingerprint <old_key_fingerprint> \
    --from provider
```

### Database Maintenance (if using external state store)

```bash
# For Redis
redis-cli --cluster check localhost:6379

# For etcd
etcdctl endpoint health
etcdctl defrag
etcdctl compact $(etcdctl endpoint status --write-out="json" | jq -r '.[] | .Status.header.revision')
```

---

## Security Operations

### Access Control

```bash
# Firewall rules
sudo ufw allow 8443/tcp comment "Provider API"
sudo ufw allow 9091/tcp from MONITORING_IP comment "Metrics"
sudo ufw deny from BLOCKED_IP

# View active connections
ss -tuln | grep -E "8443|9091"
```

### Audit Logging

```yaml
logging:
  audit:
    enabled: true
    output: "/var/log/provider-daemon/audit.log"
    events:
      - "key_usage"
      - "bid_submitted"
      - "workload_deployed"
      - "workload_terminated"
      - "usage_submitted"
      - "config_changed"
```

### Incident Response

```bash
# If compromise suspected:

# 1. Immediately pause operations
virtengine tx provider set-status \
    --status PAUSED \
    --reason "Security incident" \
    --from provider

# 2. Rotate all keys
provider-daemon keys add provider-emergency --keyring-backend file
# Complete key rotation procedure

# 3. Audit recent activity
provider-daemon audit-log --since "48 hours ago" | grep -E "error|unauthorized|failed"

# 4. Notify security team
# Email: security@virtengine.com

# 5. Document incident
# Create incident report
```

---

## High Availability Setup

### Active-Passive Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  provider-daemon-primary:
    image: virtengine/provider-daemon:v1.0.0
    volumes:
      - ./config.yaml:/config.yaml:ro
      - ./keyring:/keyring:ro
    command: start --config /config.yaml --leader-election --leader-id primary
    environment:
      - ETCD_ENDPOINTS=etcd:2379
    healthcheck:
      test: ["CMD", "curl", "-k", "https://localhost:8443/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  provider-daemon-secondary:
    image: virtengine/provider-daemon:v1.0.0
    volumes:
      - ./config.yaml:/config.yaml:ro
      - ./keyring:/keyring:ro
    command: start --config /config.yaml --leader-election --leader-id secondary
    environment:
      - ETCD_ENDPOINTS=etcd:2379
    depends_on:
      - provider-daemon-primary

  etcd:
    image: quay.io/coreos/etcd:v3.5.0
    command:
      - /usr/local/bin/etcd
      - --data-dir=/etcd-data
      - --listen-client-urls=http://0.0.0.0:2379
      - --advertise-client-urls=http://etcd:2379
    volumes:
      - etcd-data:/etcd-data

volumes:
  etcd-data:
```

### Load Balancer Health Checks

```nginx
# nginx.conf
upstream provider-daemon {
    server provider-primary:8443 max_fails=3 fail_timeout=10s;
    server provider-secondary:8443 backup;
}

server {
    listen 443 ssl;
    
    location /health {
        proxy_pass https://provider-daemon/health;
        proxy_connect_timeout 5s;
    }
    
    location / {
        proxy_pass https://provider-daemon;
    }
}
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: provider-daemon
  namespace: virtengine-system
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: provider-daemon
  template:
    metadata:
      labels:
        app: provider-daemon
    spec:
      containers:
        - name: provider-daemon
          image: virtengine/provider-daemon:v1.0.0
          args:
            - start
            - --config=/config/config.yaml
            - --leader-election
          ports:
            - containerPort: 8443
              name: api
            - containerPort: 9091
              name: metrics
          livenessProbe:
            httpGet:
              path: /health
              port: 8443
              scheme: HTTPS
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8443
              scheme: HTTPS
            initialDelaySeconds: 10
            periodSeconds: 5
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2000m"
              memory: "2Gi"
          volumeMounts:
            - name: config
              mountPath: /config
            - name: keyring
              mountPath: /keyring
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: provider-daemon-config
        - name: keyring
          secret:
            secretName: provider-daemon-keyring
```

---

## Troubleshooting

### Daemon Not Starting

```bash
# Check logs
journalctl -u provider-daemon -n 100

# Common issues:
# 1. Key not found
provider-daemon keys list

# 2. Config error
provider-daemon validate-config --config ~/.provider-daemon/config.yaml

# 3. Port in use
ss -tuln | grep 8443

# 4. Permission denied
ls -la ~/.provider-daemon/
```

### Bids Not Winning

```bash
# Check bid history
provider-daemon bids history --limit 20

# Check pricing vs market
virtengine query market orders --state open --limit 10

# Adjust pricing if needed
# Edit config.yaml bidding.base_prices

# Restart to apply
sudo systemctl restart provider-daemon
```

### Workloads Failing

```bash
# Check workload logs
provider-daemon workloads logs workload_123

# Check orchestrator status
kubectl get pods -n virtengine-workloads
kubectl describe pod <pod-name>

# Check resource availability
provider-daemon resources status
```

### Usage Submission Failures

```bash
# Check pending reports
provider-daemon usage pending

# Check chain connectivity
virtengine query provider info $(virtengine keys show provider -a)

# Retry failed submissions
provider-daemon usage submit --force

# Check failed reports directory
ls -la /var/lib/provider-daemon/failed_reports/
```

---

## Appendix: Metrics Reference

| Metric | Type | Description |
|--------|------|-------------|
| `provider_daemon_up` | gauge | Daemon running |
| `provider_bids_submitted_total` | counter | Total bids submitted |
| `provider_bids_won_total` | counter | Winning bids |
| `provider_workloads_active` | gauge | Active workloads |
| `provider_workloads_total` | counter | Total workloads |
| `provider_usage_submissions_total` | counter | Usage submissions |
| `provider_usage_submission_errors_total` | counter | Submission errors |
| `provider_resource_cpu_allocated` | gauge | Allocated CPU |
| `provider_resource_memory_allocated_bytes` | gauge | Allocated memory |
| `provider_resource_gpu_allocated` | gauge | Allocated GPU |

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
