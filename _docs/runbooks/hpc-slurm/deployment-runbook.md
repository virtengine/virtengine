# SLURM Cluster Deployment Runbook

## Overview

This runbook covers the deployment, scaling, and recovery procedures for SLURM HPC clusters on Kubernetes using the VirtEngine provider daemon.

## Readiness and Isolation Status

This chart is a rootless prototype and is blocked for production or multi-tenant use. Its default SLURM profile uses `proctrack/linuxproc`, `task/affinity`, and `jobacct_gather/linux`; all SLURM cgroup constraints are explicitly disabled because ordinary non-privileged Kubernetes pods do not receive a delegated writable cgroup v2 subtree. Kubernetes pod requests and limits can bound a pod, but SLURM cannot enforce per-job CPU, memory, swap, or device isolation inside that pod. Do not claim tenant isolation from this profile.

Least privilege remains unverified until a rendered deployment is validated and each digest-pinned image is independently confirmed to run with the declared identity. `securityIdentities.slurm` supplies one username and numeric UID/GID for slurmctld, slurmdbd, and slurmd because shared authentication and filesystems require consistent ownership; `munge`, `mariadb`, `nodeAgent`, and `utility` retain component-specific UID/GID values. A production profile requires a reviewed cgroup delegation mechanism or an equivalent independently validated isolation boundary; changing plugins or enabling `Constrain*` settings alone is not sufficient.

## Prerequisites

- Kubernetes cluster (v1.25+) with:
  - StorageClass for persistent volumes
  - Network policies enabled (optional but recommended)
  - RBAC enabled
- Helm v3.10+
- `kubectl` configured for target cluster
- Provider daemon running with HPC module enabled

### Stable Secret Prerequisites

The chart never creates credentials. Before every install, provision the following Kubernetes Secrets in the release namespace from an approved secret manager or protected files. Do not place secret material in Helm values, command-line `--set` arguments, or source control.

```bash
kubectl create namespace slurm-prototype

kubectl create secret generic slurm-prod-munge \
  --namespace slurm-prototype \
  --from-file=munge.key=/secure/path/munge.key

kubectl create secret generic slurm-prod-database \
  --namespace slurm-prototype \
  --from-file=password=/secure/path/database-password

kubectl create secret generic slurm-prod-mariadb \
  --namespace slurm-prototype \
  --from-file=root-password=/secure/path/mariadb-root-password

kubectl create secret generic slurm-prod-node-agent-tls \
  --namespace slurm-prototype \
  --from-file=ca.crt=/secure/path/ca.crt \
  --from-file=tls.crt=/secure/path/tls.crt \
  --from-file=tls.key=/secure/path/tls.key
```

An External Secrets controller may materialize the same Secret names and keys. Wait for all four Kubernetes Secrets to exist before installing. The chart deliberately uses no cluster `lookup`: missing references fail rendering, and upgrades cannot rotate credentials accidentally.

### Durable State Prerequisites

The authoritative state paths are `/var/spool/slurm` in `slurmctld`, `/var/spool/slurm` in `slurmdbd`, and `/var/lib/mysql` in MariaDB. Each path is mounted from a PVC. Runtime sockets, logs, and temporary files may use `emptyDir`; an `emptyDir` must never be mounted over an authoritative path.

By default, each StatefulSet creates a claim with `persistentVolumeClaimRetentionPolicy.whenDeleted=Retain` and `whenScaled=Retain`. The backing StorageClass and PV reclaim policy are separate controls and must also preserve data. Before install or upgrade, require `reclaimPolicy: Retain` or an approved storage lifecycle with equivalent independently verified protection:

```bash
kubectl get storageclass fast-ssd -o jsonpath='{.reclaimPolicy}{"\n"}'
kubectl get pv -o custom-columns=NAME:.metadata.name,CLAIM:.spec.claimRef.name,RECLAIM:.spec.persistentVolumeReclaimPolicy
```

To bind retained claims created outside this release, set `controller.persistence.existingClaim`, `database.persistence.existingClaim`, and `mariadb.persistence.existingClaim`. A non-empty value suppresses that component's `volumeClaimTemplates` and requires that component to have exactly one replica. The chart's `persistence.accessMode` applies only to generated per-replica claims; it cannot describe or make an existing claim safe for shared writers. The operator owns existing-claim creation, access mode, encryption, snapshots, expansion, namespace placement, and deletion protection. HA replicas must use generated per-replica claims unless a future chart version implements and validates an application-level shared-state protocol.

## Deployment Procedures

### 1. Initial Cluster Deployment

#### Using Helm Directly

```bash
# Add VirtEngine Helm repository (if using remote charts)
helm repo add virtengine https://charts.virtengine.dev
helm repo update

# Deploy a non-tenant prototype rehearsal only
helm install slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prototype \
  --set cluster.id=hpc-cluster-prototype \
  --set cluster.name="Prototype HPC Cluster" \
  --set cluster.providerAddress=virtengine1provider123 \
  --set munge.existingSecret=slurm-prod-munge \
  --set munge.secretKeyName=munge.key \
  --set database.config.existingSecret=slurm-prod-database \
  --set database.config.secretPasswordKey=password \
  --set mariadb.existingSecret=slurm-prod-mariadb \
  --set mariadb.secretRootPasswordKey=root-password \
  --set nodeAgent.tls.existingSecret=slurm-prod-node-agent-tls \
  --set nodeAgent.tls.caCertKey=ca.crt \
  --set nodeAgent.tls.clientCertKey=tls.crt \
  --set nodeAgent.tls.clientKeyKey=tls.key \
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

Default `helm template`, `helm install`, and `helm lint` fail until all enabled components have explicit Secret names and key mappings. For upgrades, use `--reuse-values` as shown below or provide the same references again; never replace them with inline values.

#### Using Provider Daemon API

```bash
# Deploy via provider daemon gRPC
grpcurl -d '{
  "cluster_id": "hpc-cluster-prototype",
  "cluster_name": "Prototype HPC Cluster",
  "namespace": "slurm-prototype",
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

### Migrate Legacy Generated Secrets

Releases installed before T4-07A generated credentials during rendering. Before
the first upgrade, provision replacement Secrets and pass their names and keys
explicitly. `--reuse-values` alone is rejected because legacy releases have no
`existingSecret` values.

```bash
helm upgrade slurm-cluster deploy/slurm/slurm-cluster \
  --namespace slurm-prod \
  --reuse-values \
  --set munge.existingSecret=slurm-prod-munge \
  --set munge.secretKeyName=munge.key \
  --set database.config.existingSecret=slurm-prod-database \
  --set database.config.secretPasswordKey=password \
  --set mariadb.existingSecret=slurm-prod-mariadb \
  --set mariadb.secretRootPasswordKey=root-password \
  --set nodeAgent.tls.existingSecret=slurm-prod-node-agent-tls \
  --set nodeAgent.tls.caCertKey=ca.crt \
  --set nodeAgent.tls.clientCertKey=tls.crt \
  --set nodeAgent.tls.clientKeyKey=tls.key \
  --wait
```

Verify every referenced Secret and key exists before invoking Helm. Do not
reuse credentials generated by a previous chart render.

### Rolling Upgrade

```bash
# 1. Check current version
helm list -n slurm-prod

# 2. Complete the checksum-gated durable-state backup procedure below

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

### Durable-State Backup

This procedure requires an approved encrypted backup destination and a maintenance window. Do not use a container writable layer as the destination. Record the release revision, image digests, PVC/PV identities, StorageClass, reclaim policy, and backup timestamp with the bundle.

1. Disable new submissions and wait for or administratively resolve active jobs.
2. Capture a transaction-consistent MariaDB dump while MariaDB is healthy.
3. Stop writers in dependency order: `slurmctld`, then `slurmdbd`, then MariaDB. Keep all three stopped until file copies or CSI snapshots finish.
4. From approved one-shot backup Jobs or the storage snapshot/export system, archive the complete slurmctld PVC path `/var/spool/slurm`, the complete slurmdbd PVC path `/var/spool/slurm`, and the MariaDB dump. Do not archive only the `ctld/` subdirectory.
5. Compute SHA-256 for every archive and write a signed or access-controlled manifest containing exactly `mariadb`, `slurmdbd`, and `slurmctld`, their hashes, source PVC UIDs, and restore order `mariadb, slurmdbd, slurmctld`.
6. Verify each hash from the final backup destination before restarting MariaDB, slurmdbd, and slurmctld in that order.

Example logical dump and local checksum commands are shown below. Redirect directly to the approved encrypted destination in production:

```bash
mkdir -m 0700 slurm-durable-backup
kubectl exec -n slurm-prod slurm-cluster-mariadb-0 -c mariadb -- sh -ec \
  'mariadb-dump --single-transaction --routines --events -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' \
  > slurm-durable-backup/mariadb.sql

# After the approved backup Jobs or snapshot exports produce these archives:
sha256sum slurm-durable-backup/mariadb.sql \
  slurm-durable-backup/slurmdbd.tar.gz \
  slurm-durable-backup/slurmctld.tar.gz \
  > slurm-durable-backup/SHA256SUMS
sha256sum --check slurm-durable-backup/SHA256SUMS
```

### Fail-Closed Durable-State Restore

Restore into three newly provisioned, empty PVCs. Never restore over an active release or reuse its claims, and never allow a partial checksum pass to start a component. Record the current Helm revision and the old MariaDB, slurmdbd, and slurmctld claim UIDs as one rollback set; retain them without mutation until the restored cluster has passed its observation window.

1. Close submissions, scale `slurmctld`, `slurmdbd`, MariaDB, and compute to zero, and verify no pods, Jobs, or external consumers mount either the old claims or the new restore claims. Provision distinct new PVCs with the required capacity, encryption, StorageClass, `Retain` lifecycle, and deletion protection, but do not change the production release's claim bindings yet.
2. Verify manifest authenticity, all three SHA-256 values, archive member types and paths, source PVC identity, expected release/image compatibility, and available capacity **before writing any new PVC**. Any mismatch aborts the entire restore and leaves both claim sets unchanged.
3. For a physical MariaDB snapshot, restore it to the new MariaDB claim using the approved CSI/storage procedure. For a logical dump, start an isolated temporary MariaDB workload using the production digest-pinned image and credentials with only the new MariaDB claim mounted. It must have a unique Service, deny production `slurmdbd` access, and accept traffic only from the restore operator/Job. Stream the verified SQL to the database client over standard input; do **not** copy or redirect the SQL file into the PVC:

   ```bash
   kubectl exec -i -n slurm-prod slurm-mariadb-restore-0 -c mariadb -- sh -ec \
     'exec mariadb -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' \
     < slurm-durable-backup/mariadb.sql
   ```

4. While MariaDB remains isolated, run `mariadb-check --all-databases`, compare expected schemas and tables, verify recorded row counts and logical data checksums, and inspect logs. On any mismatch, stop the temporary workload and preserve the new claim for investigation. On success, perform a clean database shutdown and remove the temporary workload and Service before production can mount the claim.
5. Using isolated one-shot restore Jobs, restore the complete slurmdbd and slurmctld `/var/spool/slurm` archives into their respective new claims. Verify ownership, modes, final file inventories, and checksums from the mounted destinations. Keep all production components stopped throughout this step.
6. Perform one controlled cutover of the release configuration that binds all three new claims. Do not mix an old claim from one component with a new claim from another. Keep workloads held at zero by the approved maintenance/deployment gate while applying and recording the new release revision and claim UIDs.
7. Start production MariaDB alone on the restored claim and repeat database health, schema, row-count, and checksum checks. Then start `slurmdbd` alone, require a healthy restored-database connection, clean logs, and expected accounting queries. Only then start `slurmctld` and require `scontrol ping`, `sacct`, job/node state, and scheduler-state checks to pass. Start compute nodes last and keep submissions closed until reconciliation completes.
8. If cutover or any later startup check fails, stop every component, atomically roll the release configuration back to the recorded revision and all three old claim bindings as one set, and validate the old stack in the same dependency order before reopening submissions. Do not retry against, overwrite, or delete either claim set. Preserve failed new claims for investigation and retain old claims until the restored cluster passes the approved observation and backup checkpoints.

The local fixture exercises archive checksums and ordering only; it is not Kubernetes, CSI, StorageClass, failover, or production restore evidence:

```bash
python scripts/hpc/slurm-durable-state-drill-test.py -v
python scripts/hpc/slurm-durable-state-drill.py
```

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

# 4. If state is corrupted, keep the controller stopped and use the
# fail-closed durable-state restore procedure above. Reconfigure is not restore.
```

### Database Recovery

```bash
# 1. Check database status
kubectl get pod -n slurm-prod slurm-cluster-mariadb-0 -o wide

# 2. If database state is corrupted, keep MariaDB, slurmdbd, and slurmctld
# stopped and use the fail-closed durable-state restore procedure above.
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
kubectl get secret -n slurm-prod slurm-cluster-munge -o jsonpath='{.data.munge\.key}' | base64 -d | sha256sum

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
