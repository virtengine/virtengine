# HPC SLURM SLA Requirements and Capacity Planning

## Overview

This document defines the SLA requirements and capacity planning guidelines for SLURM HPC clusters deployed on Kubernetes via the VirtEngine platform.

## SLA Requirements

### Availability Targets

| Component | Target Availability | Max Monthly Downtime | RTO | RPO |
|-----------|-------------------|---------------------|-----|-----|
| SLURM Controller (slurmctld) | 99.9% | 43.8 minutes | 15 min | 0 |
| SLURM Database (slurmdbd) | 99.9% | 43.8 minutes | 30 min | 1 hour |
| Compute Nodes (slurmd) | 99.5% | 3.6 hours | 5 min | N/A |
| Job Scheduler | 99.9% | 43.8 minutes | 15 min | 0 |
| Accounting System | 99.5% | 3.6 hours | 30 min | 1 hour |

### Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Job submission latency | < 500ms | Time from sbatch to job ID returned |
| Job startup latency | < 60s | Time from scheduled to running (idle node) |
| sinfo response time | < 100ms | 95th percentile |
| squeue response time | < 200ms | 95th percentile |
| Scheduler cycle time | < 5s | Time between scheduling passes |
| Node registration time | < 30s | Time from pod ready to SLURM idle |

### Reliability Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Job completion rate | > 99% | Excluding user-cancelled jobs |
| False node failures | < 0.1% | Nodes marked DOWN incorrectly |
| Data integrity | 100% | No job accounting data loss |
| Secret rotation success | 100% | Munge key rotations without disruption |

## Capacity Planning

### Node Sizing Guidelines

#### Controller Node (slurmctld)

| Cluster Size | CPU | Memory | Storage | Notes |
|-------------|-----|--------|---------|-------|
| < 100 nodes | 2 cores | 4 GB | 10 GB | Single controller |
| 100-500 nodes | 4 cores | 8 GB | 50 GB | Consider HA |
| 500-1000 nodes | 8 cores | 16 GB | 100 GB | HA required |
| > 1000 nodes | 16 cores | 32 GB | 200 GB | HA + backup controller |

#### Database Node (slurmdbd + MariaDB)

| Cluster Size | CPU | Memory | DB Storage | Retention |
|-------------|-----|--------|------------|-----------|
| < 100 nodes | 2 cores | 4 GB | 20 GB | 1 year |
| 100-500 nodes | 4 cores | 8 GB | 100 GB | 1 year |
| 500-1000 nodes | 8 cores | 16 GB | 500 GB | 1 year |
| > 1000 nodes | 16 cores | 32 GB | 1 TB | 2 years |

#### Compute Nodes

Compute node sizing depends on workload characteristics:

**CPU-Only Workloads:**
```
Recommended: 64-128 cores per node, 4-8 GB RAM per core
Overcommit ratio: 1:1 (no overcommit for HPC)
```

**GPU Workloads:**
```
Recommended: 8 GPUs per node (matches NVIDIA DGX topology)
CPU ratio: 8-16 cores per GPU
Memory ratio: 64-128 GB per GPU
```

### Storage Requirements

| Storage Type | Use Case | Size Formula | IOPS |
|-------------|----------|--------------|------|
| Controller state | Job state, checkpoints | 1 GB per 100 nodes | 100 |
| Database | Accounting data | 10 GB + (jobs/day × 1 KB × retention) | 500 |
| Compute scratch | Job temporary files | 100 GB - 2 TB per node | 10000 |
| Shared filesystem | User home, shared data | Based on user requirements | Varies |

### Network Requirements

| Path | Bandwidth | Latency | Notes |
|------|-----------|---------|-------|
| Controller ↔ Compute | 1 Gbps | < 10 ms | Control plane traffic |
| Compute ↔ Compute | 25-100 Gbps | < 5 μs | For MPI workloads |
| Controller ↔ Database | 1 Gbps | < 5 ms | Accounting writes |
| Compute ↔ Storage | 10-100 Gbps | < 1 ms | I/O-bound workloads |

### Scaling Formulas

#### Maximum Nodes per Controller

```
Max nodes = (Controller CPU cores × 50) × (1 + RAM_GB / 16)

Example: 8 cores, 16 GB RAM
Max nodes = (8 × 50) × (1 + 16/16) = 400 × 2 = 800 nodes
```

#### Database Connection Pool Size

```
Pool size = min(max_connections, (compute_nodes × 2) + (controllers × 10) + buffer)

Example: 100 compute nodes, 2 controllers
Pool size = min(200, (100 × 2) + (2 × 10) + 20) = min(200, 240) = 200
```

#### Scheduler Cycle Time Estimation

```
Cycle time (ms) = base_time + (pending_jobs × 0.1) + (running_jobs × 0.05) + (nodes × 0.5)

Example: 1000 pending, 500 running, 100 nodes
Cycle time = 100 + (1000 × 0.1) + (500 × 0.05) + (100 × 0.5) = 275 ms
```

## Capacity Planning Checklist

### Pre-Deployment

- [ ] Determine maximum node count for next 12 months
- [ ] Calculate storage requirements for job accounting
- [ ] Identify network bandwidth requirements for workloads
- [ ] Plan for HA if > 100 nodes
- [ ] Define backup and recovery procedures
- [ ] Set up monitoring and alerting

### Ongoing Monitoring

- [ ] Track scheduler cycle time trends
- [ ] Monitor database growth rate
- [ ] Review node utilization monthly
- [ ] Check for job queuing trends
- [ ] Validate backup integrity quarterly

### Scaling Triggers

| Metric | Warning Threshold | Critical Threshold | Action |
|--------|------------------|-------------------|--------|
| Scheduler cycle time | > 5s | > 15s | Add controller resources |
| Database disk usage | > 70% | > 85% | Expand storage or archive |
| Controller CPU | > 70% avg | > 90% avg | Scale controller |
| Node wait time | > 1 hour avg | > 4 hours avg | Add compute nodes |
| Job failure rate | > 2% | > 5% | Investigate root cause |

## Resource Quotas

### Default Resource Limits

```yaml
# Per-user defaults (customize per organization)
defaultLimits:
  maxJobsPerUser: 100
  maxSubmitJobsPerUser: 500
  maxCPUsPerUser: 1024
  maxGPUsPerUser: 32
  maxMemoryGBPerUser: 4096
  maxWallTimeHours: 168  # 1 week

# Per-account defaults
accountLimits:
  maxJobsPerAccount: 500
  maxCPUsPerAccount: 4096
  maxGPUsPerAccount: 128
```

### QoS Tier Definitions

| Tier | Priority | Max Wall Time | Max Nodes | Preemption | Cost Factor |
|------|----------|---------------|-----------|------------|-------------|
| debug | 200 | 30 min | 2 | No | 1.0x |
| normal | 100 | 24 hours | 32 | No | 1.0x |
| long | 50 | 168 hours | 16 | Yes (requeue) | 1.5x |
| premium | 150 | 72 hours | 64 | No | 2.0x |
| preemptible | 10 | 4 hours | 128 | Yes (cancel) | 0.5x |

## Disaster Recovery

### Backup Strategy

| Component | Frequency | Retention | Method |
|-----------|-----------|-----------|--------|
| MariaDB full backup | Daily | 30 days | mysqldump |
| MariaDB incremental | Hourly | 7 days | Binary log |
| Controller state | Hourly | 7 days | PVC snapshot |
| Helm values | Per change | Forever | Git |
| Munge key | Per rotation | Forever | Vault |

### Recovery Procedures

1. **Controller Loss**: Automatic recovery from PVC, RTO < 15 min
2. **Database Corruption**: Restore from backup, RTO < 30 min
3. **Full Cluster Loss**: Redeploy from Helm + restore DB, RTO < 2 hours
4. **Namespace Deletion**: Restore PVCs + redeploy, RTO < 4 hours

### Multi-Region Considerations

For clusters requiring geographic distribution:

- Deploy controllers in primary region only
- Compute nodes can span regions with latency < 50ms
- Database replication for DR (async, RPO 1 hour)
- Consider job affinity to minimize cross-region traffic

## Cost Optimization

### Right-Sizing Recommendations

1. **Use preemptible nodes** for fault-tolerant workloads (50% cost savings)
2. **Implement job packing** to maximize node utilization
3. **Set appropriate time limits** to prevent runaway jobs
4. **Archive old accounting data** to reduce storage costs
5. **Use node pools** with autoscaling for variable workloads

### Cost Allocation

Tag resources for chargeback:

```yaml
# Recommended labels for cost allocation
labels:
  virtengine.com/cluster-id: "{{ .Values.cluster.id }}"
  virtengine.com/cost-center: "{{ .Values.costCenter }}"
  virtengine.com/department: "{{ .Values.department }}"
  virtengine.com/environment: "{{ .Values.environment }}"
```

## Compliance Considerations

### Data Residency

- Ensure PVCs are provisioned in required regions
- Configure network policies to prevent cross-region data flow
- Implement data classification labels for sensitive workloads

### Audit Logging

- Enable SLURM job accounting for all jobs
- Configure Kubernetes audit logging
- Retain logs per compliance requirements (typically 7 years for financial)

### Access Control

- Implement RBAC for Kubernetes access
- Use SLURM accounts for job submission authorization
- Rotate munge keys annually (or per security policy)

---

*Last updated: 2026-01-30*
*Version: 1.0.0*
