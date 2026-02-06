# Task 30C: Multi-Region Deployment & Disaster Recovery

**vibe-kanban ID:** `2de509ce-096a-4493-b6cb-374d023b2c80`

## Problem Statement

VirtEngine requires production-grade multi-region infrastructure for:
- High availability (validator set distribution)
- Disaster recovery (geographic redundancy)
- Compliance (data residency requirements)
- Performance (low-latency provider access)

Current infrastructure is single-region which creates:
- Single point of failure
- No disaster recovery capability
- High latency for global users
- No compliance with data residency laws

## Acceptance Criteria

### AC-1: Multi-Region Kubernetes Clusters
- [ ] Primary region cluster (US-East)
- [ ] Secondary region cluster (EU-West)
- [ ] Tertiary region cluster (APAC)
- [ ] Cross-cluster networking (VPN mesh)
- [ ] Regional load balancers

### AC-2: Database Replication
- [ ] CockroachDB multi-region setup
- [ ] Cross-region replication
- [ ] Regional read replicas
- [ ] Automatic failover
- [ ] Backup to multiple regions

### AC-3: State Sync & Blockchain
- [ ] Validator geographic distribution policy
- [ ] State sync for fast bootstrap
- [ ] Archive nodes per region
- [ ] RPC endpoint per region

### AC-4: Observability Multi-Region
- [ ] Centralized logging (cross-region)
- [ ] Multi-region Prometheus federation
- [ ] Cross-region alerting
- [ ] Region health dashboards

### AC-5: Disaster Recovery Automation
- [ ] RTO < 15 minutes
- [ ] RPO < 5 minutes
- [ ] Automated failover runbooks
- [ ] DR testing automation
- [ ] Recovery validation tests

### AC-6: CI/CD Multi-Region
- [ ] Regional deployment stages
- [ ] Canary deployments
- [ ] Regional rollback capability
- [ ] Blue-green per region

## Technical Requirements

### Infrastructure Architecture

```
                    ┌─────────────────────────────────────────┐
                    │          Global Load Balancer           │
                    │        (Cloudflare/AWS Global)          │
                    └────────────┬────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   US-East-1     │    │   EU-West-1     │    │   AP-Southeast  │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Kubernetes  │ │    │ │ Kubernetes  │ │    │ │ Kubernetes  │ │
│ │  Cluster    │ │    │ │  Cluster    │ │    │ │  Cluster    │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Validators  │ │    │ │ Validators  │ │    │ │ Validators  │ │
│ │   (33%)     │ │    │ │   (33%)     │ │    │ │   (34%)     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ CockroachDB │◄┼────┼►│ CockroachDB │◄┼────┼►│ CockroachDB │ │
│ │  Regional   │ │    │ │  Regional   │ │    │ │  Regional   │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │   Archive   │ │    │ │   Archive   │ │    │ │   Archive   │ │
│ │    Node     │ │    │ │    Node     │ │    │ │    Node     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌────────────┴────────────┐
                    │    Cross-Region VPN     │
                    │       (WireGuard)       │
                    └─────────────────────────┘
```

### Terraform Multi-Region Structure

```
infra/terraform/
├── modules/
│   ├── vpc/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── kubernetes/
│   │   ├── main.tf
│   │   ├── node_pools.tf
│   │   ├── networking.tf
│   │   └── outputs.tf
│   ├── database/
│   │   ├── cockroachdb.tf
│   │   ├── replication.tf
│   │   └── outputs.tf
│   ├── observability/
│   │   ├── prometheus.tf
│   │   ├── loki.tf
│   │   └── grafana.tf
│   └── dns/
│       ├── global_lb.tf
│       └── regional.tf
├── regions/
│   ├── us-east-1/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   ├── eu-west-1/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   └── ap-southeast-1/
│       ├── main.tf
│       ├── terraform.tfvars
│       └── backend.tf
├── global/
│   ├── dns.tf
│   ├── iam.tf
│   └── state.tf
└── dr/
    ├── failover.tf
    ├── runbooks/
    └── tests/
```

### Regional Kubernetes Module

```hcl
# infra/terraform/modules/kubernetes/main.tf

variable "region" {
  description = "AWS region"
  type        = string
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "node_groups" {
  description = "Node group configurations"
  type = map(object({
    instance_types = list(string)
    min_size       = number
    max_size       = number
    desired_size   = number
    labels         = map(string)
    taints         = list(object({
      key    = string
      value  = string
      effect = string
    }))
  }))
}

# EKS Cluster
resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  version  = var.kubernetes_version
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true # Controlled via security groups
    security_group_ids      = [aws_security_group.cluster.id]
  }

  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }

  enabled_cluster_log_types = [
    "api",
    "audit",
    "authenticator",
    "controllerManager",
    "scheduler"
  ]

  tags = {
    Region      = var.region
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# Node Groups
resource "aws_eks_node_group" "nodes" {
  for_each = var.node_groups

  cluster_name    = aws_eks_cluster.main.name
  node_group_name = each.key
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.private_subnet_ids

  instance_types = each.value.instance_types

  scaling_config {
    min_size     = each.value.min_size
    max_size     = each.value.max_size
    desired_size = each.value.desired_size
  }

  labels = merge(
    each.value.labels,
    {
      "virtengine.io/region" = var.region
    }
  )

  dynamic "taint" {
    for_each = each.value.taints
    content {
      key    = taint.value.key
      value  = taint.value.value
      effect = taint.value.effect
    }
  }

  tags = {
    Region = var.region
  }
}

# Cross-Region VPC Peering
resource "aws_vpc_peering_connection" "cross_region" {
  for_each = var.peer_regions

  vpc_id        = var.vpc_id
  peer_vpc_id   = each.value.vpc_id
  peer_region   = each.value.region
  auto_accept   = false

  tags = {
    Name = "virtengine-${var.region}-to-${each.key}"
  }
}
```

### CockroachDB Multi-Region

```yaml
# deploy/cockroachdb/multi-region-values.yaml

cockroachdb:
  multiRegion:
    enabled: true
    regions:
      - name: us-east-1
        zone: us-east-1a,us-east-1b,us-east-1c
        primary: true
      - name: eu-west-1
        zone: eu-west-1a,eu-west-1b,eu-west-1c
        primary: false
      - name: ap-southeast-1
        zone: ap-southeast-1a,ap-southeast-1b,ap-southeast-1c
        primary: false

  conf:
    # Geographic data distribution
    localities:
      - region=us-east-1
      - region=eu-west-1
      - region=ap-southeast-1
    
    # Survival goals
    survival_goal: region
    
    # Locality-aware queries
    experimental_enable_implicit_column_partitioning: true

  storage:
    persistentVolume:
      size: 500Gi
      storageClass: gp3-encrypted

  replication:
    factor: 3
    
  backup:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    retention: 30d
    destinations:
      - s3://virtengine-backup-us-east-1/cockroachdb
      - s3://virtengine-backup-eu-west-1/cockroachdb
      - gs://virtengine-backup-ap/cockroachdb
```

### Disaster Recovery Runbook

```yaml
# infra/dr/runbooks/regional-failover.yaml

name: Regional Failover Runbook
version: "1.0"
rto: 15m
rpo: 5m

triggers:
  - name: region_health_check_failed
    threshold: 3 consecutive failures
    window: 5m
  - name: manual_trigger
    authorized_roles:
      - sre-lead
      - platform-engineer

pre_checks:
  - name: verify_backup_freshness
    command: |
      crdb sql --execute "SELECT now() - max(backup_end_time) AS age FROM [SHOW BACKUP LATEST IN 's3://...']"
    threshold: "age < interval '5 minutes'"
  
  - name: verify_target_region_health
    command: |
      kubectl --context=${TARGET_REGION} get nodes -o wide
    expect: "Ready"

steps:
  - name: pause_traffic
    action: cloudflare_disable_origin
    params:
      origin: ${FAILING_REGION}
    timeout: 30s
    
  - name: verify_replication_caught_up
    action: cockroachdb_verify_replication
    params:
      cluster: virtengine-${TARGET_REGION}
    timeout: 2m
    
  - name: promote_secondary
    action: cockroachdb_set_primary
    params:
      region: ${TARGET_REGION}
    timeout: 1m
    rollback:
      action: cockroachdb_set_primary
      params:
        region: ${ORIGINAL_REGION}
    
  - name: scale_target_validators
    action: kubernetes_scale
    params:
      context: ${TARGET_REGION}
      deployment: validator
      replicas: ${VALIDATOR_COUNT}
    timeout: 5m
    
  - name: update_dns
    action: route53_failover
    params:
      record: api.virtengine.io
      target: ${TARGET_REGION_LB}
    timeout: 30s
    
  - name: enable_traffic
    action: cloudflare_enable_origin
    params:
      origin: ${TARGET_REGION}
    timeout: 30s

post_checks:
  - name: verify_api_health
    command: |
      curl -f https://api.virtengine.io/cosmos/base/tendermint/v1beta1/node_info
    retries: 5
    interval: 10s
    
  - name: verify_block_production
    command: |
      virtengine query block --height latest
    expect: "height increasing"
    
  - name: notify_team
    action: pagerduty_resolve
    params:
      incident: ${INCIDENT_ID}
      message: "Regional failover complete to ${TARGET_REGION}"

rollback:
  trigger: any_post_check_failed
  steps:
    - name: revert_dns
      action: route53_failover
      params:
        record: api.virtengine.io
        target: ${ORIGINAL_REGION_LB}
    - name: demote_target
      action: cockroachdb_set_primary
      params:
        region: ${ORIGINAL_REGION}
```

### DR Testing Automation

```go
// infra/dr/tests/failover_test.go
package dr_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/virtengine/virtengine/infra/dr"
)

func TestRegionalFailover(t *testing.T) {
    if testing.Short() {
        t.Skip("DR tests require full infrastructure")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    runner := dr.NewRunbookRunner(dr.Config{
        RunbookPath: "runbooks/regional-failover.yaml",
        DryRun:      false,
        Regions: []string{
            "us-east-1",
            "eu-west-1",
        },
    })

    // Record initial state
    initialState := runner.CaptureState(ctx)
    require.NotNil(t, initialState)

    // Simulate failure in us-east-1
    t.Log("Simulating us-east-1 failure")
    err := runner.SimulateRegionFailure(ctx, "us-east-1")
    require.NoError(t, err)

    // Execute failover
    t.Log("Executing failover to eu-west-1")
    result, err := runner.Execute(ctx, dr.FailoverParams{
        FailingRegion: "us-east-1",
        TargetRegion:  "eu-west-1",
    })
    require.NoError(t, err)

    // Verify RTO
    rto := result.EndTime.Sub(result.StartTime)
    t.Logf("Actual RTO: %v", rto)
    require.Less(t, rto, 15*time.Minute, "RTO exceeded 15 minutes")

    // Verify system health
    health := runner.VerifyHealth(ctx, "eu-west-1")
    require.True(t, health.API.Healthy)
    require.True(t, health.Database.Healthy)
    require.True(t, health.Blockchain.Producing)

    // Verify data integrity (RPO)
    dataCheck := runner.VerifyDataIntegrity(ctx)
    require.Less(t, dataCheck.DataLoss, 5*time.Minute, "RPO exceeded 5 minutes")

    // Recover original region
    t.Log("Recovering us-east-1")
    err = runner.RecoverRegion(ctx, "us-east-1")
    require.NoError(t, err)

    // Failback
    t.Log("Executing failback to us-east-1")
    _, err = runner.Execute(ctx, dr.FailoverParams{
        FailingRegion: "eu-west-1",
        TargetRegion:  "us-east-1",
    })
    require.NoError(t, err)

    // Verify restored state
    finalState := runner.CaptureState(ctx)
    require.Equal(t, initialState.Region, finalState.Region)
}

func TestBackupRestoration(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
    defer cancel()

    backup := dr.NewBackupManager(dr.BackupConfig{
        Source:      "us-east-1",
        Destination: "eu-west-1",
    })

    // Get latest backup
    latest, err := backup.GetLatest(ctx)
    require.NoError(t, err)
    require.NotNil(t, latest)

    // Verify backup age
    age := time.Since(latest.Timestamp)
    t.Logf("Backup age: %v", age)
    require.Less(t, age, 6*time.Hour, "Backup too old")

    // Restore to test cluster
    testCluster := "virtengine-dr-test"
    err = backup.RestoreTo(ctx, latest, testCluster)
    require.NoError(t, err)

    // Verify data integrity
    integrity, err := backup.VerifyIntegrity(ctx, testCluster)
    require.NoError(t, err)
    require.True(t, integrity.Valid)
    require.Zero(t, integrity.CorruptedRows)

    // Cleanup
    err = backup.CleanupTestCluster(ctx, testCluster)
    require.NoError(t, err)
}
```

### Multi-Region CI/CD Pipeline

```yaml
# .github/workflows/multi-region-deploy.yaml

name: Multi-Region Deployment

on:
  push:
    branches: [main]
    tags: ['v*']

env:
  REGIONS: "us-east-1,eu-west-1,ap-southeast-1"
  PRIMARY_REGION: "us-east-1"

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      image_tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          push: true
          tags: ghcr.io/virtengine/virtengine:${{ github.sha }}

  deploy-primary:
    needs: build
    runs-on: ubuntu-latest
    environment: production-us-east-1
    steps:
      - name: Deploy to US-East-1
        uses: ./.github/actions/deploy-region
        with:
          region: us-east-1
          image_tag: ${{ needs.build.outputs.image_tag }}
          is_primary: true
          
      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh us-east-1
          
      - name: Verify block production
        run: |
          kubectl --context us-east-1 exec -it validator-0 -- \
            virtengine status | jq '.sync_info.catching_up == false'

  deploy-secondary:
    needs: [build, deploy-primary]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        region: [eu-west-1, ap-southeast-1]
      fail-fast: false
    environment: production-${{ matrix.region }}
    steps:
      - name: Deploy to ${{ matrix.region }}
        uses: ./.github/actions/deploy-region
        with:
          region: ${{ matrix.region }}
          image_tag: ${{ needs.build.outputs.image_tag }}
          is_primary: false
          
      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh ${{ matrix.region }}

  verify-global:
    needs: [deploy-primary, deploy-secondary]
    runs-on: ubuntu-latest
    steps:
      - name: Verify cross-region connectivity
        run: |
          ./scripts/verify-cross-region.sh
          
      - name: Verify database replication
        run: |
          ./scripts/verify-db-replication.sh
          
      - name: Update global load balancer
        run: |
          ./scripts/update-global-lb.sh enable-all
```

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `infra/terraform/modules/vpc/` | VPC module | 300 |
| `infra/terraform/modules/kubernetes/` | K8s module | 500 |
| `infra/terraform/modules/database/` | CockroachDB module | 400 |
| `infra/terraform/modules/observability/` | Monitoring module | 300 |
| `infra/terraform/regions/us-east-1/` | US region config | 150 |
| `infra/terraform/regions/eu-west-1/` | EU region config | 150 |
| `infra/terraform/regions/ap-southeast-1/` | APAC region config | 150 |
| `infra/terraform/global/` | Global resources | 200 |
| `deploy/cockroachdb/multi-region-values.yaml` | DB config | 200 |
| `infra/dr/runbooks/*.yaml` | DR runbooks | 400 |
| `infra/dr/tests/*.go` | DR tests | 600 |
| `.github/workflows/multi-region-deploy.yaml` | CI/CD | 300 |
| `scripts/verify-*.sh` | Verification scripts | 400 |

**Total Estimated:** 4,050 lines

## Validation Checklist

- [ ] Cross-region VPN connectivity verified
- [ ] CockroachDB replication lag < 100ms
- [ ] Validators distributed evenly across regions
- [ ] RTO < 15 minutes in failover test
- [ ] RPO < 5 minutes data loss
- [ ] Global load balancer health checks working
- [ ] DR runbook executed successfully
- [ ] Backup restoration tested
- [ ] Cross-region monitoring federation working
- [ ] Regional rollback capability verified

## Dependencies

- 30B (HSM) - Key management for regional validators

## Risk Mitigation

1. **Network Partitions**
   - CockroachDB handles partitions gracefully
   - Validators use gossip for state sync
   - Timeouts tuned for cross-region latency

2. **Data Residency**
   - Zone survival goal maintains regional data
   - GDPR compliance via EU-only table partitions
   - Audit logging for cross-region access

3. **Cost Management**
   - Cross-region data transfer costs monitored
   - Auto-scaling based on regional demand
   - Reserved instances for baseline capacity
