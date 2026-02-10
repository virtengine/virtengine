# SLURM Cluster Deployment Runbook

## Overview

This runbook covers the deployment, scaling, and recovery procedures for SLURM HPC clusters on Kubernetes using the VirtEngine provider daemon.

## Prerequisites

- Kubernetes cluster (v1.25+) with:
  - StorageClass for persistent volumes
  - Network policies enabled (optional but recommended)
  - RBAC enabled
- Helm v3.10+
- `kubectl` configured for target cluster
- Provider daemon running with HPC module enabled

## Deployment Procedures

### 1. Initial Cluster Deployment

#### Using Helm Directly

```bash
# Create namespace
kubectl create namespace slurm-prod

# Add VirtEngine Helm repository (if using remote charts)
helm repo add virtengine https://charts.virtengine.dev
helm repo update

# Deploy SLURM cluster
helm install slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --set cluster.id=hpc-cluster-prod \
  --set cluster.name="Production HPC Cluster" \
  --set cluster.providerAddress=virtengine1provider123 \
  --set compute.replicas=8 \
  --set controller.persistence.size=50Gi \
  --set database.persistence.size=100Gi \
  --set mariadb.persistence.size=200Gi \
  --set global.storageClass=fast-ssd \
  --set nodeAgent.enabled=true \
  --set nodeAgent.config.providerEndpoint=https://provider.example.com:8443 \
  --wait \
  --timeout 15m
```

#### Using Provider Daemon API

```bash
# Deploy via provider daemon gRPC
grpcurl -d '{
  "cluster_id": "hpc-cluster-prod",
  "cluster_name": "Production HPC Cluster",
  "namespace": "slurm-prod",
  "template": {
    "partitions": [
      {"name": "normal", "nodes": 8, "max_runtime_seconds": 86400, "state": "up"},
      {"name": "gpu", "nodes": 4, "max_runtime_seconds": 259200, "features": ["gpu"], "state": "up"}
    ]
  },
  "storage_class": "fast-ssd",
  "provider_endpoint": "https://provider.example.com:8443"
}' provider.example.com:8443 virtengine.provider.v1.HPCService/DeployCluster
```

### 2. Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n slurm-prod

# Expected output:
# NAME                                    READY   STATUS    RESTARTS   AGE
# slurm-cluster-controller-0              2/2     Running   0          5m
# slurm-cluster-slurmdbd-0                2/2     Running   0          5m
# slurm-cluster-mariadb-0                 1/1     Running   0          5m
# slurm-cluster-compute-0                 3/3     Running   0          4m
# slurm-cluster-compute-1                 3/3     Running   0          4m
# ... (more compute nodes)

# Verify SLURM controller is responding
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- scontrol ping

# Expected: Slurmctld(primary) at slurm-cluster-controller-0 is UP

# Check node status
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- sinfo

# Check partition status
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- sinfo -p normal,gpu
```

### 3. Post-Deployment Configuration

#### Configure QoS Policies

```bash
# Create QoS policy
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sacctmgr -i add qos premium Priority=100 MaxJobsPerUser=50

# Associate QoS with accounts
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sacctmgr -i modify account root set qos=normal,premium
```

#### Configure Accounts and Users

```bash
# Create cluster account
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sacctmgr -i add cluster virtengine

# Create organization account
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sacctmgr -i add account research Cluster=virtengine

# Add users
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sacctmgr -i add user researcher1 Account=research
```

## Scaling Procedures

### Scale Up Compute Nodes

```bash
# Using Helm upgrade
helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --set compute.replicas=16 \
  --reuse-values \
  --wait \
  --timeout 10m

# Verify new nodes are registered
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- sinfo -N
```

### Scale Down Compute Nodes

```bash
# First, drain nodes that will be removed
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol update NodeName=slurm-cluster-compute-[12-15] State=DRAIN Reason="Scaling down"

# Wait for jobs to complete
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  squeue -w slurm-cluster-compute-[12-15]

# Scale down
helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --set compute.replicas=12 \
  --reuse-values \
  --wait
```

### Add GPU Node Pool

```bash
# Add GPU nodes via Helm values
cat > gpu-values.yaml << EOF
nodePools:
  - name: gpu-a100
    replicas: 4
    cpus: 128
    memory: 1048576
    gpus: 8
    gpuType: nvidia
    features:
      - gpu
      - a100
      - nvlink
    nodeSelector:
      node-type: gpu
    tolerations:
      - key: nvidia.com/gpu
        operator: Exists
        effect: NoSchedule
    resources:
      requests:
        nvidia.com/gpu: "8"
      limits:
        nvidia.com/gpu: "8"
EOF

helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  -f gpu-values.yaml \
  --reuse-values \
  --wait
```

## Upgrade Procedures

### Rolling Upgrade

```bash
# 1. Check current version
helm list -n slurm-prod

# 2. Backup database
kubectl exec -n slurm-prod slurm-cluster-mariadb-0 -- \
  mysqldump -u root -p$MYSQL_ROOT_PASSWORD slurm_acct_db > slurm_backup_$(date +%Y%m%d).sql

# 3. Drain nodes in batches
for node in slurm-cluster-compute-{0..3}; do
  kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
    scontrol update NodeName=$node State=DRAIN Reason="Upgrade"
done

# 4. Perform upgrade
helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --set global.slurmVersion=23.11.0 \
  --reuse-values \
  --wait \
  --timeout 20m

# 5. Resume nodes
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol update NodeName=ALL State=RESUME

# 6. Verify upgrade
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol show config | grep SlurmVersion
```

## Recovery Procedures

### Controller Recovery

```bash
# 1. Check controller status
kubectl get pod -n slurm-prod slurm-cluster-controller-0 -o wide
kubectl logs -n slurm-prod slurm-cluster-controller-0 -c slurmctld --tail=100

# 2. If controller is in CrashLoopBackOff, check state files
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  ls -la /var/spool/slurm/ctld/

# 3. Force restart controller
kubectl delete pod -n slurm-prod slurm-cluster-controller-0

# 4. If state is corrupted, reconfigure from scratch
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol reconfigure
```

### Database Recovery

```bash
# 1. Check database status
kubectl get pod -n slurm-prod slurm-cluster-mariadb-0 -o wide

# 2. If database is corrupted, restore from backup
kubectl cp slurm_backup.sql slurm-prod/slurm-cluster-mariadb-0:/tmp/

kubectl exec -n slurm-prod slurm-cluster-mariadb-0 -- \
  mysql -u root -p$MYSQL_ROOT_PASSWORD slurm_acct_db < /tmp/slurm_backup.sql

# 3. Restart slurmdbd
kubectl delete pod -n slurm-prod slurm-cluster-slurmdbd-0
```

### Node Recovery

```bash
# 1. Check node status
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  sinfo -N -l

# 2. Resume nodes stuck in DOWN state
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol update NodeName=slurm-cluster-compute-[0-7] State=RESUME

# 3. Force node restart if unresponsive
kubectl delete pod -n slurm-prod slurm-cluster-compute-3

# 4. Check for hardware issues
kubectl exec -n slurm-prod slurm-cluster-compute-3 -c slurmd -- \
  slurmd -C
```

### Munge Key Issues

```bash
# 1. Verify munge key consistency
kubectl get secret -n slurm-prod slurm-cluster-munge -o jsonpath='{.data.munge\.key}' | base64 -d | md5sum

# 2. If keys are mismatched, regenerate
kubectl delete secret -n slurm-prod slurm-cluster-munge

# 3. Re-deploy to generate new key
helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --reuse-values

# 4. Restart all pods to pick up new key
kubectl rollout restart statefulset -n slurm-prod
```

## Monitoring and Alerts

### Key Metrics to Monitor

| Metric | Threshold | Action |
|--------|-----------|--------|
| Controller uptime | < 99.9% | Investigate crashes |
| Node availability | < 90% | Check node health |
| Job wait time | > 1 hour | Scale up or review QoS |
| Database connections | > 80% max | Increase connection pool |
| PVC usage | > 80% | Expand storage |

### Prometheus Alerts

```yaml
# Example alert rules
groups:
  - name: slurm-alerts
    rules:
      - alert: SLURMControllerDown
        expr: up{job="slurm-controller"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "SLURM controller is down"
          
      - alert: SLURMNodesUnavailable
        expr: slurm_nodes_idle + slurm_nodes_mixed < 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "No SLURM nodes available for jobs"
          
      - alert: SLURMDatabaseConnectionHigh
        expr: slurm_dbd_connections > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SLURM database connection count high"
```

## Maintenance Windows

### Scheduled Maintenance

```bash
# 1. Announce maintenance
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol create reservation ReservationName=maint StartTime=now+1hour Duration=4:00:00 Users=root Flags=MAINT,IGNORE_JOBS Nodes=ALL

# 2. Drain all nodes
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol update NodeName=ALL State=DRAIN Reason="Scheduled maintenance"

# 3. Perform maintenance...

# 4. Resume nodes
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol update NodeName=ALL State=RESUME

# 5. Delete reservation
kubectl exec -n slurm-prod slurm-cluster-controller-0 -c slurmctld -- \
  scontrol delete reservation maint
```

## Troubleshooting

### Common Issues

| Issue | Symptoms | Resolution |
|-------|----------|------------|
| Jobs stuck in pending | `squeue` shows many PD jobs | Check node availability, QoS limits |
| Nodes not registering | `sinfo` shows nodes in UNKNOWN | Check slurmd logs, munge auth |
| Accounting errors | Job completion not recorded | Check slurmdbd connection, DB space |
| Auth failures | "Authentication failure" in logs | Verify munge key sync |

### Debug Commands

```bash
# Check slurmctld logs
kubectl logs -n slurm-prod slurm-cluster-controller-0 -c slurmctld -f

# Check slurmd logs on compute node
kubectl logs -n slurm-prod slurm-cluster-compute-0 -c slurmd -f

# Check munge status
kubectl exec -n slurm-prod slurm-cluster-compute-0 -c munge -- \
  munge -n | unmunge

# Check network connectivity
kubectl exec -n slurm-prod slurm-cluster-compute-0 -c slurmd -- \
  nc -zv slurm-cluster-controller-0 6817
```

## Contact and Escalation

| Level | Team | Contact | Response Time |
|-------|------|---------|---------------|
| L1 | On-call SRE | #hpc-oncall | 15 min |
| L2 | HPC Platform | #hpc-platform | 1 hour |
| L3 | Core Engineering | #virtengine-core | 4 hours |

---

*Last updated: 2026-01-30*
*Version: 1.0.0*
