# Performance Tuning Guide

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Overview](#overview)
2. [System-Level Tuning](#system-level-tuning)
3. [Validator Node Tuning](#validator-node-tuning)
4. [Provider Daemon Tuning](#provider-daemon-tuning)
5. [Database Tuning](#database-tuning)
6. [Network Tuning](#network-tuning)
7. [ML Inference Tuning](#ml-inference-tuning)
8. [Monitoring and Benchmarking](#monitoring-and-benchmarking)

---

## Overview

This guide provides performance tuning recommendations for VirtEngine production infrastructure. All recommendations should be tested in staging before production deployment.

### Performance Goals

| Component | Metric | Target |
|-----------|--------|--------|
| Block time | P99 latency | < 6 seconds |
| Transaction throughput | TPS | > 1000 |
| VEID scoring | P95 latency | < 5 minutes |
| API response | P99 latency | < 500ms |
| Node sync | Blocks/second | > 100 (catching up) |

---

## System-Level Tuning

### Kernel Parameters

```bash
# /etc/sysctl.d/99-virtengine.conf

# Network
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_keepalive_intvl = 15
net.ipv4.tcp_tw_reuse = 1

# File descriptors
fs.file-max = 2097152
fs.nr_open = 2097152

# Virtual memory
vm.swappiness = 10
vm.dirty_ratio = 60
vm.dirty_background_ratio = 2
vm.max_map_count = 262144

# Apply changes
sudo sysctl -p /etc/sysctl.d/99-virtengine.conf
```

### File Descriptor Limits

```bash
# /etc/security/limits.d/virtengine.conf
validator soft nofile 1048576
validator hard nofile 1048576
validator soft nproc 65535
validator hard nproc 65535

# Verify
ulimit -n
```

### Systemd Service Limits

```ini
# /etc/systemd/system/virtengine.service.d/limits.conf
[Service]
LimitNOFILE=1048576
LimitNPROC=65535
LimitCORE=infinity
LimitMEMLOCK=infinity
```

### CPU Governor

```bash
# Set performance governor
sudo cpupower frequency-set -g performance

# Verify
cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

# Persistent via /etc/default/cpufrequtils
GOVERNOR="performance"
```

### Disk I/O Scheduler

```bash
# For NVMe SSDs
echo "none" | sudo tee /sys/block/nvme0n1/queue/scheduler

# For SATA SSDs
echo "mq-deadline" | sudo tee /sys/block/sda/queue/scheduler

# Persistent via udev rule
# /etc/udev/rules.d/60-io-scheduler.rules
ACTION=="add|change", KERNEL=="nvme*", ATTR{queue/scheduler}="none"
ACTION=="add|change", KERNEL=="sd*", ATTR{queue/scheduler}="mq-deadline"
```

---

## Validator Node Tuning

### config.toml Optimization

```toml
# ~/.virtengine/config/config.toml

#######################################################
# P2P Configuration
#######################################################
[p2p]
# Larger peer limits for better connectivity
max_num_inbound_peers = 100
max_num_outbound_peers = 50

# Faster peer exchange
flush_throttle_timeout = "10ms"

# Larger send/receive buffers
send_rate = 20480000  # 20 MB/s
recv_rate = 20480000  # 20 MB/s

# Faster peer discovery
pex_reactor = true

#######################################################
# Mempool Configuration
#######################################################
[mempool]
# Larger mempool for high throughput
size = 10000
max_txs_bytes = 2147483648  # 2 GB
max_tx_bytes = 1048576      # 1 MB

# Faster recheck
recheck = true
broadcast = true

# Cache size
cache_size = 10000

#######################################################
# Consensus Configuration
#######################################################
[consensus]
# Optimized timeouts (adjust based on network latency)
timeout_propose = "3s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "3s"

# Skip timeouts when possible
skip_timeout_commit = false

# Create empty blocks for consistency
create_empty_blocks = true
create_empty_blocks_interval = "5s"

# Peer gossip optimization
peer_gossip_sleep_duration = "10ms"

#######################################################
# Storage Configuration
#######################################################
[storage]
discard_abci_responses = false

#######################################################
# Transaction Indexer
#######################################################
[tx_index]
# Use psql for better query performance in production
indexer = "kv"
# For high-volume: indexer = "psql"
```

### app.toml Optimization

```toml
# ~/.virtengine/config/app.toml

#######################################################
# Base Configuration
#######################################################
# Minimum gas price (adjust based on network)
minimum-gas-prices = "0.025uve"

# Faster block processing
halt-height = 0
halt-time = 0

#######################################################
# Pruning
#######################################################
# Aggressive pruning for non-archive nodes
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "0"
pruning-interval = "10"

#######################################################
# State Sync
#######################################################
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2

#######################################################
# gRPC Configuration
#######################################################
[grpc]
enable = true
address = "0.0.0.0:9090"
max-recv-msg-size = "10485760"
max-send-msg-size = "2147483647"

#######################################################
# API Configuration
#######################################################
[api]
enable = true
address = "tcp://127.0.0.1:1317"
max-open-connections = 1000
rpc-read-timeout = 10
rpc-write-timeout = 0
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = false

#######################################################
# Telemetry
#######################################################
[telemetry]
enabled = true
service-name = "virtengine-validator"
enable-hostname = true
enable-hostname-label = true
enable-service-label = true
prometheus-retention-time = 3600
```

### Go Runtime Optimization

```bash
# In systemd service file

[Service]
# Optimize garbage collection
Environment="GOGC=100"

# Use more OS threads for I/O
Environment="GOMAXPROCS=16"

# Memory ballast for reduced GC pressure (optional)
# Requires code modification
```

---

## Provider Daemon Tuning

### config.yaml Optimization

```yaml
# ~/.provider-daemon/config.yaml

# Server configuration
server:
  address: "0.0.0.0:8443"
  read_timeout: "30s"
  write_timeout: "60s"
  max_header_bytes: 1048576
  
  # Connection pooling
  max_connections: 10000
  keep_alive: true
  keep_alive_timeout: "120s"

# Chain connection
chain:
  node_url: "http://localhost:26657"
  grpc_url: "localhost:9090"
  
  # Connection pool
  grpc_pool_size: 10
  
  # Retry configuration
  max_retries: 5
  retry_delay: "1s"
  retry_backoff: 2.0
  
  # Timeouts
  dial_timeout: "10s"
  request_timeout: "30s"

# Orchestration
orchestration:
  adapter: "kubernetes"
  kubernetes:
    # Connection pooling
    qps: 100
    burst: 150
    
    # Informer cache
    resync_period: "10m"
    
    # Parallel operations
    max_concurrent_reconciles: 10

# Bid engine
bidding:
  enabled: true
  
  # Rate limiting
  max_bids_per_minute: 30
  max_concurrent_bids: 10
  
  # Caching
  order_cache_ttl: "30s"
  price_cache_ttl: "60s"

# Workload management
workloads:
  max_concurrent: 100
  
  # Health checks
  health_check_interval: "30s"
  health_check_timeout: "10s"
  health_check_workers: 10
  
  # Status sync
  sync_interval: "5s"
  sync_workers: 5

# Usage reporting
usage:
  report_interval: "1h"
  batch_size: 100
  
  # Parallel collection
  collection_workers: 4
  
  # Buffer size
  buffer_size: 10000

# Caching
cache:
  # In-memory cache
  max_size: 10000
  ttl: "5m"
  
  # Redis (optional, for HA)
  redis:
    enabled: false
    address: "localhost:6379"
    pool_size: 100
```

### Kubernetes Adapter Tuning

```yaml
# Provider daemon Kubernetes configuration

orchestration:
  kubernetes:
    # Higher QPS for busy clusters
    qps: 200
    burst: 300
    
    # Resource request/limit tuning
    default_requests:
      cpu: "100m"
      memory: "128Mi"
    default_limits:
      cpu: "4"
      memory: "8Gi"
    
    # Pod scheduling optimization
    scheduling:
      # Spread workloads across nodes
      topology_spread_constraints: true
      max_skew: 1
      
      # Priority classes
      priority_class: "virtengine-workload"
      
      # Preemption
      preemption_policy: "PreemptLowerPriority"
```

---

## Database Tuning

### PostgreSQL Configuration

```ini
# /etc/postgresql/14/main/postgresql.conf

#######################################################
# Memory
#######################################################
# 25% of system RAM
shared_buffers = 16GB

# 75% of system RAM for query caching
effective_cache_size = 48GB

# Working memory per operation
work_mem = 256MB

# Maintenance operations
maintenance_work_mem = 2GB

#######################################################
# Write Ahead Log
#######################################################
wal_level = replica
wal_buffers = 64MB
checkpoint_completion_target = 0.9
checkpoint_timeout = 10min
max_wal_size = 4GB
min_wal_size = 1GB

#######################################################
# Query Planner
#######################################################
random_page_cost = 1.1  # For SSDs
effective_io_concurrency = 200
default_statistics_target = 100

#######################################################
# Connections
#######################################################
max_connections = 500
superuser_reserved_connections = 3

#######################################################
# Parallelism
#######################################################
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
max_parallel_maintenance_workers = 4
parallel_leader_participation = on

#######################################################
# Logging
#######################################################
log_min_duration_statement = 1000  # Log queries > 1s
log_checkpoints = on
log_connections = off
log_disconnections = off
log_lock_waits = on
log_temp_files = 0

#######################################################
# Autovacuum
#######################################################
autovacuum = on
autovacuum_max_workers = 4
autovacuum_naptime = 10s
autovacuum_vacuum_threshold = 50
autovacuum_analyze_threshold = 50
autovacuum_vacuum_scale_factor = 0.05
autovacuum_analyze_scale_factor = 0.025
```

### Connection Pooling with PgBouncer

```ini
# /etc/pgbouncer/pgbouncer.ini

[databases]
virtengine = host=localhost port=5432 dbname=virtengine_db

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt

# Pool mode
pool_mode = transaction

# Pool sizing
default_pool_size = 100
min_pool_size = 10
reserve_pool_size = 10
reserve_pool_timeout = 5

# Connection limits
max_client_conn = 10000
max_db_connections = 200

# Timeouts
server_reset_query = DISCARD ALL
server_check_delay = 30
server_check_query = SELECT 1
server_lifetime = 3600
server_idle_timeout = 600
client_idle_timeout = 0

# Logging
log_connections = 0
log_disconnections = 0
log_pooler_errors = 1
stats_period = 60
```

### Index Optimization

```sql
-- Create indexes for common queries
CREATE INDEX CONCURRENTLY idx_blocks_height ON blocks(height);
CREATE INDEX CONCURRENTLY idx_blocks_time ON blocks(time);
CREATE INDEX CONCURRENTLY idx_txs_hash ON transactions(hash);
CREATE INDEX CONCURRENTLY idx_txs_height ON transactions(height);
CREATE INDEX CONCURRENTLY idx_events_tx_hash ON events(tx_hash);
CREATE INDEX CONCURRENTLY idx_events_type ON events(type);

-- Composite indexes for common joins
CREATE INDEX CONCURRENTLY idx_txs_height_hash ON transactions(height, hash);

-- Partial indexes for specific queries
CREATE INDEX CONCURRENTLY idx_orders_pending 
ON orders(created_at) 
WHERE state = 'pending';

-- Analyze after creating indexes
ANALYZE;
```

---

## Network Tuning

### TCP Optimization

```bash
# /etc/sysctl.d/99-network.conf

# TCP buffer sizes
net.core.rmem_default = 262144
net.core.rmem_max = 16777216
net.core.wmem_default = 262144
net.core.wmem_max = 16777216

# TCP tuning
net.ipv4.tcp_window_scaling = 1
net.ipv4.tcp_timestamps = 1
net.ipv4.tcp_sack = 1
net.ipv4.tcp_fack = 1
net.ipv4.tcp_fastopen = 3

# Connection tracking
net.netfilter.nf_conntrack_max = 2097152
net.netfilter.nf_conntrack_tcp_timeout_established = 86400

# Reduce TIME_WAIT
net.ipv4.tcp_max_tw_buckets = 2000000
```

### Network Interface Tuning

```bash
# Increase ring buffer size
ethtool -G eth0 rx 4096 tx 4096

# Enable receive side scaling
ethtool -K eth0 rxhash on

# Increase interrupt coalescing
ethtool -C eth0 rx-usecs 100 tx-usecs 100

# Set MTU (if network supports jumbo frames)
ip link set eth0 mtu 9000
```

### Load Balancer Configuration (HAProxy)

```haproxy
# /etc/haproxy/haproxy.cfg

global
    maxconn 100000
    nbproc 4
    nbthread 8
    cpu-map auto:1/1-8 0-7
    
defaults
    mode tcp
    option tcplog
    option tcp-check
    timeout connect 10s
    timeout client 30s
    timeout server 30s
    
frontend rpc_frontend
    bind *:26657
    default_backend rpc_backend
    
backend rpc_backend
    balance roundrobin
    option tcp-check
    server node1 10.0.0.1:26657 check inter 5s fall 3 rise 2
    server node2 10.0.0.2:26657 check inter 5s fall 3 rise 2
    server node3 10.0.0.3:26657 check inter 5s fall 3 rise 2

frontend grpc_frontend
    bind *:9090
    default_backend grpc_backend
    
backend grpc_backend
    balance roundrobin
    option tcp-check
    server node1 10.0.0.1:9090 check inter 5s fall 3 rise 2
    server node2 10.0.0.2:9090 check inter 5s fall 3 rise 2
```

---

## ML Inference Tuning

### TensorFlow Optimization

```python
# Deterministic inference settings
import os
import tensorflow as tf

# Force CPU for determinism
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"

# Enable XLA JIT compilation
tf.config.optimizer.set_jit(True)

# Set thread parallelism
tf.config.threading.set_inter_op_parallelism_threads(4)
tf.config.threading.set_intra_op_parallelism_threads(8)

# Memory growth
gpus = tf.config.list_physical_devices('GPU')
if gpus:
    for gpu in gpus:
        tf.config.experimental.set_memory_growth(gpu, True)

# Enable deterministic operations
tf.config.experimental.enable_op_determinism()
os.environ["TF_DETERMINISTIC_OPS"] = "1"
```

### Inference Service Configuration

```yaml
# ML inference service config

inference:
  # Model loading
  model_path: "/models/veid_scorer_v1.0.0.h5"
  preload: true
  
  # Batching
  batch_size: 32
  batch_timeout_ms: 100
  max_batch_size: 64
  
  # Threading
  num_threads: 8
  inter_op_threads: 4
  intra_op_threads: 8
  
  # Memory
  gpu_memory_fraction: 0.8
  allow_growth: true
  
  # Caching
  prediction_cache_size: 10000
  cache_ttl_seconds: 300
  
  # Timeouts
  inference_timeout_ms: 5000
  warmup_requests: 100
```

### GPU Optimization (if using)

```bash
# NVIDIA settings for ML inference

# Set persistence mode
nvidia-smi -pm 1

# Lock GPU clocks for consistency
nvidia-smi -ac 5001,1590

# Set compute mode to exclusive process
nvidia-smi -c EXCLUSIVE_PROCESS

# Monitor GPU usage
watch -n 1 nvidia-smi
```

---

## Monitoring and Benchmarking

### Performance Metrics to Monitor

```yaml
# Prometheus alerting rules for performance

groups:
  - name: performance
    rules:
      - alert: HighBlockTime
        expr: histogram_quantile(0.99, tendermint_consensus_block_time_seconds_bucket) > 6
        for: 5m
        labels:
          severity: warning
          
      - alert: LowTPS
        expr: rate(tendermint_consensus_total_txs[5m]) < 100
        for: 10m
        labels:
          severity: info
          
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / 1024 / 1024 / 1024 > 50
        for: 10m
        labels:
          severity: warning
          
      - alert: HighCPUUsage
        expr: rate(process_cpu_seconds_total[5m]) > 0.9
        for: 10m
        labels:
          severity: warning
          
      - alert: SlowQueryResponse
        expr: histogram_quantile(0.99, http_request_duration_seconds_bucket{handler="/api/v1/query"}) > 0.5
        for: 5m
        labels:
          severity: warning
```

### Benchmarking Commands

```bash
# Load testing with vegeta
echo "GET http://localhost:26657/status" | \
  vegeta attack -rate=1000 -duration=60s | \
  vegeta report

# Transaction throughput test
virtengine tx benchmark \
  --txs 10000 \
  --concurrency 100 \
  --from operator

# Disk I/O benchmark
fio --name=randread --ioengine=libaio --rw=randread \
    --bs=4k --numjobs=4 --size=4G --runtime=30 \
    --direct=1 --group_reporting

# Network latency benchmark
qperf -t 30 <peer-ip> tcp_bw tcp_lat

# Database query benchmark
pgbench -c 100 -j 4 -T 60 virtengine_db
```

### Performance Dashboard Panels

```json
{
  "panels": [
    {
      "title": "Block Time P99",
      "expr": "histogram_quantile(0.99, sum(rate(tendermint_consensus_block_time_seconds_bucket[5m])) by (le))"
    },
    {
      "title": "TPS",
      "expr": "rate(tendermint_consensus_total_txs[5m])"
    },
    {
      "title": "Mempool Size",
      "expr": "tendermint_mempool_size"
    },
    {
      "title": "Peer Count",
      "expr": "tendermint_p2p_peers"
    },
    {
      "title": "VEID Scoring Latency P95",
      "expr": "histogram_quantile(0.95, sum(rate(veid_scoring_duration_seconds_bucket[5m])) by (le))"
    },
    {
      "title": "Memory Usage",
      "expr": "process_resident_memory_bytes / 1024 / 1024 / 1024"
    },
    {
      "title": "CPU Usage",
      "expr": "rate(process_cpu_seconds_total[5m])"
    },
    {
      "title": "Disk I/O",
      "expr": "rate(node_disk_io_time_seconds_total[5m])"
    }
  ]
}
```

---

## Appendix: Tuning Checklist

### Before Tuning

- [ ] Establish baseline metrics
- [ ] Identify performance bottlenecks
- [ ] Document current configuration
- [ ] Plan rollback procedure
- [ ] Test in staging environment

### During Tuning

- [ ] Change one parameter at a time
- [ ] Monitor impact for 15-30 minutes
- [ ] Document each change and result
- [ ] Rollback if performance degrades

### After Tuning

- [ ] Run benchmark suite
- [ ] Compare with baseline
- [ ] Document final configuration
- [ ] Update monitoring thresholds
- [ ] Schedule periodic review

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
