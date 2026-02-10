# HPC Provider Operations Guide

This guide covers configuring and operating HPC scheduler integrations (SLURM, MOAB, Open OnDemand) with the VirtEngine provider daemon.

## Overview

The VirtEngine provider daemon can integrate with HPC schedulers to execute compute jobs submitted on-chain. The integration supports:

- **SLURM** - Standard HPC workload manager
- **MOAB** - Adaptive Computing's workload manager
- **Open OnDemand (OOD)** - Web-based HPC portal

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Provider Daemon                              │
│                                                                  │
│  ┌──────────────┐    ┌─────────────────┐    ┌───────────────┐  │
│  │   On-Chain   │    │  HPC Job        │    │  HPCScheduler │  │
│  │   Events     │───▶│  Service        │───▶│  Interface    │  │
│  └──────────────┘    └─────────────────┘    └───────┬───────┘  │
│                              │                       │          │
│                              ▼                       ▼          │
│                      ┌───────────────┐    ┌──────────────────┐ │
│                      │ Usage Reporter │    │ Scheduler        │ │
│                      │ & Auditor      │    │ Adapter          │ │
│                      └───────────────┘    │ (SLURM/MOAB/OOD) │ │
│                                           └──────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                                      │
                                                      ▼
                                            ┌──────────────────┐
                                            │  HPC Cluster     │
                                            │  (SLURM/MOAB)    │
                                            └──────────────────┘
```

## Configuration

### Basic Configuration

Add the HPC configuration to your provider daemon config file:

```yaml
hpc:
  enabled: true
  scheduler_type: slurm  # Options: slurm, moab, ood
  cluster_id: "my-hpc-cluster-001"
  
  job_service:
    job_poll_interval: 15s
    job_timeout_default: 24h
    max_concurrent_jobs: 100
    enable_state_recovery: true
    state_store_path: /var/lib/virtengine/hpc-state
  
  usage_reporting:
    enabled: true
    report_interval: 5m
    batch_size: 50
    retry_on_failure: true
  
  retry:
    max_retries: 3
    initial_backoff: 1s
    max_backoff: 30s
    backoff_multiplier: 2.0
  
  audit:
    enabled: true
    log_path: /var/log/virtengine/hpc-audit.log
    log_job_events: true
    log_security_events: true
    log_usage_reports: true
```

### SLURM Configuration

```yaml
hpc:
  scheduler_type: slurm
  slurm:
    cluster_name: virtengine-hpc
    controller_host: slurmctld.example.com
    controller_port: 6817
    auth_method: munge  # Options: munge, jwt
    # auth_token: "..."  # Required if auth_method is jwt
    default_partition: default
    job_poll_interval: 10s
    connection_timeout: 30s
    max_retries: 3
```

### MOAB Configuration

```yaml
hpc:
  scheduler_type: moab
  moab:
    server_host: moab-server.example.com
    server_port: 42559
    use_tls: true
    auth_method: password  # Options: password, key, kerberos
    # username and password should be provided via environment variables
    default_queue: batch
    default_account: default
    job_poll_interval: 15s
    connection_timeout: 30s
    waldur_integration: true
    waldur_endpoint: https://waldur.example.com/api
    ssh_host_key_callback: known_hosts  # Options: known_hosts, pinned, insecure
    ssh_known_hosts_path: /home/provider/.ssh/known_hosts
```

### Open OnDemand Configuration

```yaml
hpc:
  scheduler_type: ood
  ood:
    base_url: https://ondemand.example.com
    cluster: virtengine-hpc
    oidc_issuer: https://veid.virtengine.io
    oidc_client_id: provider-daemon
    # oidc_client_secret provided via environment variable
    session_poll_interval: 15s
    connection_timeout: 30s
    slurm_partition: interactive
    default_hours: 4
    enable_file_browser: true
```

## Environment Variables

Sensitive credentials should be provided via environment variables:

```bash
# SLURM JWT authentication
export SLURM_JWT_TOKEN="your-jwt-token"

# MOAB authentication
export MOAB_USERNAME="provider-service"
export MOAB_PASSWORD="secure-password"

# OOD OIDC credentials
export OOD_OIDC_CLIENT_SECRET="client-secret"

# Provider signing key (for on-chain reports)
export PROVIDER_SIGNING_KEY_PATH="/path/to/key"
```

## Job Lifecycle

### Job States

| State | Description |
|-------|-------------|
| `pending` | Job received, not yet submitted to scheduler |
| `queued` | Job submitted and queued in scheduler |
| `starting` | Job is starting on compute nodes |
| `running` | Job is actively running |
| `suspended` | Job is paused/held |
| `completed` | Job completed successfully |
| `failed` | Job failed with error |
| `cancelled` | Job was cancelled |
| `timeout` | Job exceeded time limit |

### State Flow

```
pending → queued → starting → running → completed
                                     ↘ failed
                                     ↘ timeout
           ↓
        cancelled
```

### Lifecycle Callbacks

The provider daemon fires callbacks on state transitions:

1. **submitted** - Job submitted to scheduler
2. **queued** - Job entered scheduler queue
3. **started** - Job began execution
4. **completed** - Job finished successfully
5. **failed** - Job failed
6. **cancelled** - Job was cancelled
7. **timeout** - Job timed out
8. **suspended** - Job was paused
9. **resumed** - Job was resumed

## Usage Reporting

Usage metrics are collected and reported on-chain:

### Metrics Collected

| Metric | Description |
|--------|-------------|
| `wall_clock_seconds` | Total elapsed time |
| `cpu_core_seconds` | CPU core-seconds used |
| `memory_gb_seconds` | Memory GB-seconds used |
| `gpu_seconds` | GPU-seconds used |
| `node_hours` | Total node-hours |
| `storage_gb_hours` | Storage consumption |
| `network_bytes_in/out` | Network transfer |
| `energy_joules` | Energy consumption (if available) |

### Usage Record Format

```json
{
  "record_id": "cluster-001-job-123-1",
  "job_id": "job-123",
  "cluster_id": "cluster-001",
  "provider_address": "virtengine1provider...",
  "customer_address": "virtengine1customer...",
  "period_start": "2024-01-15T10:00:00Z",
  "period_end": "2024-01-15T11:00:00Z",
  "metrics": {
    "wall_clock_seconds": 3600,
    "cpu_core_seconds": 14400,
    "gpu_seconds": 3600
  },
  "is_final": false,
  "job_state": "running",
  "timestamp": "2024-01-15T11:00:05Z",
  "signature": "abcd1234..."
}
```

## Audit Logging

Audit events are logged for compliance and debugging:

### Event Types

| Event Type | Category | Description |
|------------|----------|-------------|
| `job_submitted` | Job | Job successfully submitted |
| `job_cancelled` | Job | Job cancelled |
| `job_lifecycle_event` | Job | State transition |
| `job_validation_failed` | Job | Invalid job spec |
| `job_submission_failed` | Job | Scheduler rejected job |
| `status_reported` | Usage | Status sent on-chain |
| `usage_reported` | Usage | Usage record submitted |
| `accounting_reported` | Usage | Final accounting sent |

### Log Format

```json
{
  "timestamp": "2024-01-15T10:00:00Z",
  "event_type": "job_submitted",
  "job_id": "job-123",
  "cluster_id": "cluster-001",
  "details": {
    "scheduler_job_id": "slurm-456",
    "scheduler_type": "slurm"
  },
  "success": true
}
```

## Troubleshooting

### Common Issues

#### Connection Failures

```
ERROR: failed to connect to SLURM: connection refused
```

**Resolution:**
1. Verify scheduler host is reachable
2. Check firewall rules
3. Verify authentication credentials
4. Check if scheduler service is running

#### Job Submission Failures

```
ERROR: job submission failed: partition not available
```

**Resolution:**
1. Verify partition/queue exists
2. Check user has access to partition
3. Verify resource limits are valid
4. Check account/project is valid

#### Authentication Errors

```
ERROR: MOAB authentication failed: invalid credentials
```

**Resolution:**
1. Verify username/password in environment variables
2. Check Kerberos tickets (if using Kerberos)
3. Verify SSH key permissions (if using key auth)
4. Check host key verification settings

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: debug
  hpc_scheduler: trace
```

### Health Checks

Check scheduler connectivity:

```bash
virtengine-provider hpc health-check
```

Output:
```
Scheduler Type: SLURM
Connection: OK
Authentication: OK
Queue Access: OK
Last Job Poll: 2024-01-15T10:00:00Z
Active Jobs: 15
Pending Reports: 3
```

## Security Considerations

### Credential Management

1. **Never log credentials** - All sensitive fields are excluded from logs
2. **Use environment variables** - Don't put secrets in config files
3. **Rotate credentials regularly** - Especially for long-running daemons
4. **Limit permissions** - Use minimal scheduler privileges

### SSH Host Key Verification

For MOAB/SSH-based connections:

```yaml
moab:
  ssh_host_key_callback: known_hosts  # RECOMMENDED
  ssh_known_hosts_path: /home/provider/.ssh/known_hosts
```

**Warning:** Never use `insecure` in production - it disables MITM protection.

### Job Isolation

Jobs are isolated by:
1. Container namespaces
2. SLURM cgroups
3. User isolation
4. Network policies

## Performance Tuning

### Poll Intervals

Adjust based on your cluster size:

| Cluster Size | Recommended Poll Interval |
|--------------|---------------------------|
| < 100 nodes | 5-10 seconds |
| 100-500 nodes | 10-15 seconds |
| 500+ nodes | 15-30 seconds |

### Concurrent Jobs

Set based on available resources:

```yaml
job_service:
  max_concurrent_jobs: 100  # Adjust based on cluster capacity
```

### Usage Reporting

Batch reports to reduce on-chain transactions:

```yaml
usage_reporting:
  report_interval: 5m
  batch_size: 50
```

## Monitoring

### Prometheus Metrics

Expose metrics for monitoring:

```yaml
metrics:
  enabled: true
  port: 9090
  path: /metrics
```

Metrics exported:
- `hpc_jobs_submitted_total`
- `hpc_jobs_completed_total`
- `hpc_jobs_failed_total`
- `hpc_job_duration_seconds`
- `hpc_usage_reports_sent_total`
- `hpc_scheduler_poll_duration_seconds`

### Alerting

Recommended alerts:

```yaml
# Alert on high job failure rate
- alert: HPCHighJobFailureRate
  expr: rate(hpc_jobs_failed_total[5m]) / rate(hpc_jobs_submitted_total[5m]) > 0.1
  for: 10m

# Alert on scheduler connectivity issues
- alert: HPCSchedulerDisconnected
  expr: hpc_scheduler_connected == 0
  for: 5m

# Alert on usage reporting backlog
- alert: HPCUsageReportBacklog
  expr: hpc_pending_usage_reports > 100
  for: 15m
```

## Appendix

### Scheduler Type Comparison

| Feature | SLURM | MOAB | OOD |
|---------|-------|------|-----|
| Batch Jobs | ✓ | ✓ | ✓ (via Job Composer) |
| Interactive Apps | Limited | Limited | ✓ |
| File Browser | ✗ | ✗ | ✓ |
| Web UI | ✗ | ✗ | ✓ |
| GPU Support | ✓ | ✓ | ✓ |
| Container Support | ✓ (Singularity) | ✓ | ✓ |
| Waldur Integration | ✗ | ✓ | ✗ |

### Job Spec Mapping

VirtEngine job specs are mapped to scheduler-specific formats:

| VirtEngine Field | SLURM | MOAB |
|------------------|-------|------|
| `nodes` | `--nodes` | `-l nodes` |
| `cpu_cores_per_node` | `--cpus-per-node` | `:ppn` |
| `memory_gb_per_node` | `--mem` | `-l mem` |
| `gpus_per_node` | `--gpus` | `:gpus` |
| `max_runtime_seconds` | `--time` | `-l walltime` |
| `queue_name` | `--partition` | `-q` |
| `container_image` | Singularity | Singularity |
