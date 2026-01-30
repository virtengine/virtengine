# HPC Cluster Template Specification

Version: 1.0.0
Status: Draft
Last Updated: 2026-01-30

## Overview

This document defines the schema for HPC cluster templates used to register and configure SLURM clusters on the VirtEngine blockchain. Cluster templates describe the physical and logical topology of HPC infrastructure, including partitions, QoS policies, and hardware classes.

## Cluster Template Schema

### ClusterTemplate

The `ClusterTemplate` represents the configuration template for an HPC cluster.

```json
{
  "template_id": "string (optional, auto-generated if empty)",
  "template_name": "string (required, 3-64 chars)",
  "template_version": "string (semver format, e.g., '1.0.0')",
  "description": "string (optional, max 1024 chars)",
  
  "partitions": [
    {
      "name": "string (required, SLURM partition name)",
      "display_name": "string (user-friendly name)",
      "nodes": "int32 (number of nodes, >= 1)",
      "max_nodes_per_job": "int32 (max nodes per job)",
      "max_runtime_seconds": "int64 (max wall-clock time)",
      "default_runtime_seconds": "int64 (default wall-clock time)",
      "priority": "int32 (0-1000, higher = more priority)",
      "features": ["string (e.g., 'gpu', 'highmem', 'infiniband')"],
      "state": "string (up|down|drain|inactive)",
      "access_control": {
        "allow_groups": ["string (group names)"],
        "deny_groups": ["string (group names)"],
        "require_reservation": "bool"
      }
    }
  ],
  
  "qos_policies": [
    {
      "name": "string (required, QoS name)",
      "priority": "int32 (job priority adjustment)",
      "max_jobs_per_user": "int32 (0 = unlimited)",
      "max_submit_jobs_per_user": "int32 (0 = unlimited)",
      "max_wall_duration_seconds": "int64",
      "max_cpus_per_user": "int32 (0 = unlimited)",
      "max_gpus_per_user": "int32 (0 = unlimited)",
      "max_memory_gb_per_user": "int32 (0 = unlimited)",
      "preempt_mode": "string (off|suspend|requeue|cancel)",
      "usage_factor": "string (fixed-point, 6 decimals, e.g., '1000000' = 1.0)"
    }
  ],
  
  "hardware_classes": {
    "cpu_classes": [
      {
        "class_id": "string (e.g., 'cpu-standard', 'cpu-highmem')",
        "description": "string",
        "cores_per_node": "int32",
        "memory_gb_per_node": "int32",
        "cpu_model": "string (e.g., 'Intel Xeon Gold 6248')",
        "cpu_generation": "string (e.g., 'Cascade Lake')",
        "threads_per_core": "int32 (1 or 2)",
        "numa_nodes": "int32",
        "features": ["string"]
      }
    ],
    "gpu_classes": [
      {
        "class_id": "string (e.g., 'gpu-a100', 'gpu-v100')",
        "description": "string",
        "gpu_model": "string (e.g., 'NVIDIA A100 80GB')",
        "gpu_count_per_node": "int32",
        "gpu_memory_gb": "int32",
        "cuda_compute_capability": "string (e.g., '8.0')",
        "nvlink_enabled": "bool",
        "mig_supported": "bool",
        "mig_profiles": ["string (e.g., '1g.10gb', '2g.20gb')"],
        "features": ["string"]
      }
    ],
    "storage_classes": [
      {
        "class_id": "string (e.g., 'nvme-local', 'lustre-parallel')",
        "description": "string",
        "storage_type": "string (nvme|ssd|hdd|lustre|gpfs|cephfs)",
        "capacity_tb": "int64",
        "iops_read": "int64 (estimated)",
        "iops_write": "int64 (estimated)",
        "bandwidth_gbps": "int64",
        "is_shared": "bool",
        "mount_path": "string"
      }
    ],
    "network_classes": [
      {
        "class_id": "string (e.g., 'infiniband-hdr', 'ethernet-100g')",
        "description": "string",
        "network_type": "string (infiniband|ethernet|roce)",
        "bandwidth_gbps": "int64",
        "latency_us": "int64 (estimated)",
        "rdma_enabled": "bool"
      }
    ]
  },
  
  "resource_limits": {
    "max_nodes_per_job": "int32",
    "max_cpus_per_job": "int32",
    "max_gpus_per_job": "int32",
    "max_memory_gb_per_job": "int32",
    "max_wall_time_seconds": "int64",
    "max_jobs_per_user": "int32",
    "max_concurrent_jobs_per_user": "int32"
  },
  
  "scheduling_policy": {
    "scheduler_type": "string (slurm|pbs|custom)",
    "backfill_enabled": "bool",
    "preemption_enabled": "bool",
    "fair_share_enabled": "bool",
    "fair_share_decay_half_life_days": "int32",
    "priority_weight_age": "int32",
    "priority_weight_fair_share": "int32",
    "priority_weight_job_size": "int32",
    "priority_weight_partition": "int32",
    "priority_weight_qos": "int32"
  },
  
  "maintenance_windows": [
    {
      "name": "string",
      "start_cron": "string (cron expression)",
      "duration_hours": "int32",
      "timezone": "string (e.g., 'UTC', 'America/New_York')"
    }
  ],
  
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## Field Definitions

### Partitions

Partitions map directly to SLURM partitions and define logical groupings of compute nodes:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | SLURM partition name (alphanumeric, hyphens allowed) |
| `display_name` | string | No | Human-readable name for UI display |
| `nodes` | int32 | Yes | Number of nodes in partition (≥1) |
| `max_nodes_per_job` | int32 | No | Max nodes a single job can request |
| `max_runtime_seconds` | int64 | Yes | Maximum wall-clock time (≥60) |
| `default_runtime_seconds` | int64 | Yes | Default wall-clock time |
| `priority` | int32 | No | Partition priority (0-1000, default: 100) |
| `features` | []string | No | SLURM feature tags for constraint matching |
| `state` | string | Yes | Partition state (up\|down\|drain\|inactive) |

### QoS Policies

Quality of Service policies control job scheduling and resource allocation:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | QoS name (alphanumeric, underscores) |
| `priority` | int32 | No | Priority adjustment factor |
| `max_jobs_per_user` | int32 | No | Max running jobs per user (0=unlimited) |
| `max_wall_duration_seconds` | int64 | No | Max wall-clock override |
| `preempt_mode` | string | No | Preemption behavior |
| `usage_factor` | string | No | Fair-share usage multiplier (fixed-point) |

### Hardware Classes

Hardware classes categorize compute resources for pricing and matching:

#### CPU Classes

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `class_id` | string | Yes | Unique identifier |
| `cores_per_node` | int32 | Yes | Physical CPU cores |
| `memory_gb_per_node` | int32 | Yes | RAM in GB |
| `cpu_model` | string | Yes | CPU model name |
| `numa_nodes` | int32 | No | NUMA topology |

#### GPU Classes

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `class_id` | string | Yes | Unique identifier |
| `gpu_model` | string | Yes | GPU model name |
| `gpu_count_per_node` | int32 | Yes | GPUs per node |
| `gpu_memory_gb` | int32 | Yes | GPU memory in GB |
| `cuda_compute_capability` | string | No | CUDA compute level |
| `mig_supported` | bool | No | Multi-Instance GPU support |

## Validation Rules

### Template Validation

1. **Template Name**: 3-64 characters, alphanumeric with hyphens
2. **Template Version**: Must be valid semver (e.g., `1.0.0`)
3. **Partitions**: At least one partition required
4. **Runtime**: `default_runtime_seconds` ≤ `max_runtime_seconds`
5. **Nodes**: `max_nodes_per_job` ≤ `nodes` for each partition
6. **Fixed-Point Values**: 6 decimal places (divide by 1,000,000 for real value)

### Partition Validation

```go
func (p *Partition) Validate() error {
    if p.Name == "" {
        return errors.New("partition name required")
    }
    if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`).MatchString(p.Name) {
        return errors.New("invalid partition name format")
    }
    if p.Nodes < 1 {
        return errors.New("partition must have at least 1 node")
    }
    if p.MaxRuntimeSeconds < 60 {
        return errors.New("max runtime must be at least 60 seconds")
    }
    if p.DefaultRuntimeSeconds > p.MaxRuntimeSeconds {
        return errors.New("default runtime cannot exceed max runtime")
    }
    return nil
}
```

### QoS Validation

```go
func (q *QoSPolicy) Validate() error {
    if q.Name == "" {
        return errors.New("qos name required")
    }
    if q.MaxJobsPerUser < 0 {
        return errors.New("max_jobs_per_user cannot be negative")
    }
    if q.UsageFactor != "" {
        factor, err := parseFixedPoint(q.UsageFactor, 6)
        if err != nil || factor < 0 {
            return errors.New("invalid usage_factor")
        }
    }
    return nil
}
```

## On-Chain Mapping

Cluster templates map to `x/hpc` types as follows:

| Template Field | On-Chain Type | Store Key |
|---------------|---------------|-----------|
| `ClusterTemplate` | `types.HPCCluster` | `ClusterPrefix + clusterID` |
| `partitions` | `[]types.Partition` | Embedded in `HPCCluster` |
| `qos_policies` | `types.ClusterMetadata.QoSPolicies` | Embedded in `ClusterMetadata` |
| `hardware_classes` | `types.ClusterMetadata` | Embedded in `ClusterMetadata` |

### Example On-Chain Registration

```go
cluster := &types.HPCCluster{
    ClusterID:       "hpc-cluster-1",
    ProviderAddress: providerAddr,
    Name:            template.TemplateName,
    Description:     template.Description,
    State:           types.ClusterStatePending,
    Partitions:      convertPartitions(template.Partitions),
    TotalNodes:      calculateTotalNodes(template.Partitions),
    Region:          "us-east-1",
    ClusterMetadata: types.ClusterMetadata{
        TotalCPUCores:    calculateTotalCPU(template),
        TotalMemoryGB:    calculateTotalMemory(template),
        TotalGPUs:        calculateTotalGPUs(template),
        GPUTypes:         extractGPUTypes(template),
        InterconnectType: extractNetworkType(template),
        StorageType:      extractStorageType(template),
    },
    SLURMVersion: "23.02.4",
}
```

## Example Templates

### Standard Compute Cluster

```json
{
  "template_name": "standard-compute",
  "template_version": "1.0.0",
  "description": "General-purpose CPU compute cluster",
  "partitions": [
    {
      "name": "normal",
      "display_name": "Normal Queue",
      "nodes": 100,
      "max_nodes_per_job": 32,
      "max_runtime_seconds": 86400,
      "default_runtime_seconds": 3600,
      "priority": 100,
      "features": ["cpu-standard"],
      "state": "up"
    },
    {
      "name": "debug",
      "display_name": "Debug Queue",
      "nodes": 4,
      "max_nodes_per_job": 2,
      "max_runtime_seconds": 1800,
      "default_runtime_seconds": 600,
      "priority": 200,
      "features": ["cpu-standard"],
      "state": "up"
    }
  ],
  "qos_policies": [
    {
      "name": "normal",
      "priority": 0,
      "max_jobs_per_user": 100,
      "max_wall_duration_seconds": 86400,
      "preempt_mode": "off"
    },
    {
      "name": "premium",
      "priority": 100,
      "max_jobs_per_user": 50,
      "max_wall_duration_seconds": 604800,
      "preempt_mode": "off",
      "usage_factor": "2000000"
    }
  ],
  "hardware_classes": {
    "cpu_classes": [
      {
        "class_id": "cpu-standard",
        "description": "Standard compute nodes",
        "cores_per_node": 64,
        "memory_gb_per_node": 256,
        "cpu_model": "AMD EPYC 7763",
        "cpu_generation": "Milan",
        "threads_per_core": 2,
        "numa_nodes": 2
      }
    ],
    "storage_classes": [
      {
        "class_id": "scratch-nvme",
        "storage_type": "nvme",
        "capacity_tb": 2,
        "is_shared": false,
        "mount_path": "/scratch"
      }
    ],
    "network_classes": [
      {
        "class_id": "ethernet-25g",
        "network_type": "ethernet",
        "bandwidth_gbps": 25,
        "rdma_enabled": false
      }
    ]
  },
  "resource_limits": {
    "max_nodes_per_job": 32,
    "max_cpus_per_job": 2048,
    "max_memory_gb_per_job": 8192,
    "max_wall_time_seconds": 604800,
    "max_concurrent_jobs_per_user": 50
  }
}
```

### GPU Training Cluster

```json
{
  "template_name": "gpu-training",
  "template_version": "1.0.0",
  "description": "GPU cluster for ML training workloads",
  "partitions": [
    {
      "name": "gpu",
      "display_name": "GPU Partition",
      "nodes": 32,
      "max_nodes_per_job": 16,
      "max_runtime_seconds": 259200,
      "default_runtime_seconds": 14400,
      "priority": 100,
      "features": ["gpu-a100", "nvlink"],
      "state": "up"
    }
  ],
  "hardware_classes": {
    "cpu_classes": [
      {
        "class_id": "cpu-gpu-host",
        "cores_per_node": 128,
        "memory_gb_per_node": 1024,
        "cpu_model": "AMD EPYC 7763",
        "cpu_generation": "Milan"
      }
    ],
    "gpu_classes": [
      {
        "class_id": "gpu-a100",
        "gpu_model": "NVIDIA A100 80GB",
        "gpu_count_per_node": 8,
        "gpu_memory_gb": 80,
        "cuda_compute_capability": "8.0",
        "nvlink_enabled": true,
        "mig_supported": true,
        "mig_profiles": ["1g.10gb", "2g.20gb", "3g.40gb", "7g.80gb"]
      }
    ],
    "network_classes": [
      {
        "class_id": "infiniband-hdr",
        "network_type": "infiniband",
        "bandwidth_gbps": 200,
        "latency_us": 1,
        "rdma_enabled": true
      }
    ]
  }
}
```

## Related Documents

- [Node Agent Protocol](./hpc-node-agent-protocol.md) - Node heartbeat and metrics protocol
- [HPC Module Types](../../x/hpc/types/) - On-chain type definitions
- [Provider Daemon](../provider-daemon-waldur-integration.md) - Provider integration guide
