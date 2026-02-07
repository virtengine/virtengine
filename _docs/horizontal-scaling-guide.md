# VirtEngine Horizontal Scaling Architecture Guide

**Version:** 1.0.0  
**Date:** 2026-01-30  
**Task Reference:** SCALE-002

---

## Implementation Summary

This document accompanies the SCALE-002 implementation which includes:

| Component              | File                                                     | Description                                      |
| ---------------------- | -------------------------------------------------------- | ------------------------------------------------ |
| Bid Deduplication      | `pkg/provider_daemon/scaling.go`                         | Distributed order partitioning and deduplication |
| Scaling Metrics        | `pkg/provider_daemon/scaling_metrics.go`                 | Prometheus metrics for scaling observability     |
| Enhanced HPA           | `deploy/kubernetes/overlays/prod/hpa-enhanced.yaml`      | Custom metrics-based autoscaling                 |
| KEDA Scaler            | `deploy/kubernetes/overlays/prod/keda-scaledobject.yaml` | Event-driven autoscaling                         |
| State Sync Script      | `scripts/state-sync-bootstrap.sh`                        | Fast validator bootstrap                         |
| Scaling Alerts         | `deploy/monitoring/prometheus/rules/scaling_alerts.yml`  | Alerting for scaling issues                      |
| Multi-Region Terraform | `infra/terraform/modules/scaling/main.tf`                | Global load balancing infrastructure             |

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture Summary](#architecture-summary)
3. [Provider Daemon Horizontal Scaling](#provider-daemon-horizontal-scaling)
4. [RPC Endpoint Load Balancing](#rpc-endpoint-load-balancing)
5. [State Sync Optimization](#state-sync-optimization)
6. [Sharding Strategy](#sharding-strategy)
7. [Multi-Region Deployment](#multi-region-deployment)
8. [Auto-Scaling Policies](#auto-scaling-policies)
9. [Monitoring and Observability](#monitoring-and-observability)
10. [Operational Runbooks](#operational-runbooks)

---

## Overview

This document defines the horizontal scaling architecture for VirtEngine, enabling the platform to handle increased load through the addition of compute resources rather than upgrading existing ones. The strategy covers all critical components: provider daemons, RPC endpoints, validators, and supporting infrastructure.

### Design Principles

| Principle                    | Description                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------- |
| **Stateless Where Possible** | Components designed to be stateless or externalize state for horizontal scaling |
| **Graceful Degradation**     | System continues operating under partial failures                               |
| **Geographic Distribution**  | Multi-region deployment for latency and availability                            |
| **Automated Scaling**        | Kubernetes HPA and custom metrics for demand-responsive scaling                 |
| **Observability First**      | Comprehensive metrics enable informed scaling decisions                         |

### Component Scaling Characteristics

| Component       | Scaling Model | State              | Coordination Required   |
| --------------- | ------------- | ------------------ | ----------------------- |
| Provider Daemon | Horizontal    | External (chain)   | Low - bid deduplication |
| Full Node (RPC) | Horizontal    | Local (catch-up)   | None - read replicas    |
| Validator       | Horizontal    | Shared (consensus) | High - BFT coordination |
| API Gateway     | Horizontal    | Stateless          | None                    |
| ML Inference    | Horizontal    | Model cache        | Low - determinism sync  |

---

## Architecture Summary

```
                                   ┌─────────────────────────────────────────────┐
                                   │           Global Load Balancer               │
                                   │         (CloudFlare / AWS Global Acc.)       │
                                   └─────────────────────────────────────────────┘
                                                         │
                 ┌───────────────────────────────────────┼───────────────────────────────────────┐
                 │                                       │                                       │
                 ▼                                       ▼                                       ▼
    ┌────────────────────────┐         ┌────────────────────────┐         ┌────────────────────────┐
    │      Region: US-EAST   │         │     Region: EU-WEST    │         │    Region: AP-SOUTH    │
    │      (Primary)         │         │     (Secondary)        │         │    (Tertiary)          │
    ├────────────────────────┤         ├────────────────────────┤         ├────────────────────────┤
    │ ┌────────────────────┐ │         │ ┌────────────────────┐ │         │ ┌────────────────────┐ │
    │ │   Istio Ingress    │ │         │ │   Istio Ingress    │ │         │ │   Istio Ingress    │ │
    │ └─────────┬──────────┘ │         │ └─────────┬──────────┘ │         │ └─────────┬──────────┘ │
    │           │            │         │           │            │         │           │            │
    │  ┌────────┴────────┐   │         │  ┌────────┴────────┐   │         │  ┌────────┴────────┐   │
    │  │                 │   │         │  │                 │   │         │  │                 │   │
    │  ▼                 ▼   │         │  ▼                 ▼   │         │  ▼                 ▼   │
    │ ┌───┐┌───┐┌───┐┌───┐  │         │ ┌───┐┌───┐┌───┐     │  │         │ ┌───┐┌───┐          │  │
    │ │FN1││FN2││FN3││FN4│  │         │ │FN1││FN2││FN3│     │  │         │ │FN1││FN2│          │  │
    │ └───┘└───┘└───┘└───┘  │         │ └───┘└───┘└───┘     │  │         │ └───┘└───┘          │  │
    │  Full Nodes (RPC)     │         │  Full Nodes (RPC)   │  │         │  Full Nodes (RPC)   │  │
    │                       │         │                     │  │         │                     │  │
    │ ┌───┐┌───┐┌───┐┌───┐  │         │ ┌───┐┌───┐          │  │         │ ┌───┐               │  │
    │ │PD1││PD2││PD3││PD4│  │         │ │PD1││PD2│          │  │         │ │PD1│               │  │
    │ └───┘└───┘└───┘└───┘  │         │ └───┘└───┘          │  │         │ └───┘               │  │
    │  Provider Daemons     │         │  Provider Daemons   │  │         │  Provider Daemons   │  │
    │         (HPA)         │         │       (HPA)         │  │         │       (HPA)         │  │
    │                       │         │                     │  │         │                     │  │
    │  ┌───┐    ┌───┐       │         │  ┌───┐    ┌───┐     │  │         │  ┌───┐              │  │
    │  │V1 │    │V2 │       │         │  │V1 │    │V2 │     │  │         │  │V1 │              │  │
    │  └───┘    └───┘       │         │  └───┘    └───┘     │  │         │  └───┘              │  │
    │    Validators         │         │    Validators       │  │         │    Validators       │  │
    └────────────────────────┘         └────────────────────────┘         └────────────────────────┘
```

---

## Provider Daemon Horizontal Scaling

The provider daemon (`pkg/provider_daemon/`) is designed for horizontal scaling with minimal coordination requirements.

### Architecture for Horizontal Scaling

```
                    ┌─────────────────────────────────────┐
                    │         Chain (Order Events)         │
                    └─────────────────────────────────────┘
                                      │
                                      │ WebSocket / gRPC
                                      ▼
         ┌────────────────────────────────────────────────────────────────┐
         │                    Provider Daemon Fleet                       │
         │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
         │  │  Instance 1  │  │  Instance 2  │  │  Instance N  │          │
         │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │          │
         │  │ │Bid Engine│ │  │ │Bid Engine│ │  │ │Bid Engine│ │          │
         │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │          │
         │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │          │
         │  │ │K8s Adapt.│ │  │ │K8s Adapt.│ │  │ │K8s Adapt.│ │          │
         │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │          │
         │  └──────────────┘  └──────────────┘  └──────────────┘          │
         └────────────────────────────────────────────────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              │                       │                       │
              ▼                       ▼                       ▼
    ┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
    │ Redis Cluster    │   │ Kubernetes API   │   │ External Storage │
    │ (Bid Dedup/Lock) │   │ (Workload Mgmt)  │   │ (S3/GCS)         │
    └──────────────────┘   └──────────────────┘   └──────────────────┘
```

### Scaling Configuration

#### Kubernetes HPA (Existing)

The existing HPA configuration in `deploy/kubernetes/base/hpa.yaml`:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: provider-daemon-hpa
  namespace: virtengine
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: provider-daemon
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
```

#### Enhanced HPA with Custom Metrics

For production deployments, add custom metrics based on business logic:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: provider-daemon-hpa-enhanced
  namespace: virtengine
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: provider-daemon
  minReplicas: 2
  maxReplicas: 20
  metrics:
    # CPU-based scaling
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    # Memory-based scaling
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    # Custom metric: pending orders queue depth
    - type: Pods
      pods:
        metric:
          name: provider_daemon_pending_orders
        target:
          type: AverageValue
          averageValue: "50"
    # Custom metric: active leases per instance
    - type: Pods
      pods:
        metric:
          name: provider_daemon_active_leases
        target:
          type: AverageValue
          averageValue: "100"
    # External metric: orders in marketplace
    - type: External
      external:
        metric:
          name: virtengine_marketplace_open_orders
          selector:
            matchLabels:
              region: "us-east-1"
        target:
          type: AverageValue
          averageValue: "500"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
        - type: Pods
          value: 1
          periodSeconds: 60
      selectPolicy: Min
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
        - type: Pods
          value: 4
          periodSeconds: 15
      selectPolicy: Max
```

### Bid Deduplication Strategy

When running multiple provider daemon instances, implement bid deduplication to prevent duplicate bids on the same order:

#### Option 1: Redis-Based Distributed Locking

```go
// pkg/provider_daemon/bid_dedup.go

type BidDeduplicator struct {
    redis    *redis.Client
    ttl      time.Duration
    instance string
}

func NewBidDeduplicator(redisAddr, instanceID string) *BidDeduplicator {
    return &BidDeduplicator{
        redis:    redis.NewClient(&redis.Options{Addr: redisAddr}),
        ttl:      5 * time.Minute,
        instance: instanceID,
    }
}

// TryClaimOrder attempts to claim exclusive processing rights for an order
func (bd *BidDeduplicator) TryClaimOrder(ctx context.Context, orderID string) (bool, error) {
    key := fmt.Sprintf("bid:claim:%s", orderID)

    // SET NX with TTL - atomic operation
    set, err := bd.redis.SetNX(ctx, key, bd.instance, bd.ttl).Result()
    if err != nil {
        return false, fmt.Errorf("redis setnx failed: %w", err)
    }

    return set, nil
}

// ReleaseClaim releases the claim on an order
func (bd *BidDeduplicator) ReleaseClaim(ctx context.Context, orderID string) error {
    key := fmt.Sprintf("bid:claim:%s", orderID)

    // Only delete if we own the claim (Lua script for atomicity)
    script := redis.NewScript(`
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `)

    _, err := script.Run(ctx, bd.redis, []string{key}, bd.instance).Result()
    return err
}
```

#### Option 2: Chain-Native Deduplication

Leverage on-chain bid tracking - the chain rejects duplicate bids from the same provider:

```go
// In BidEngine.processBid - check existing bids before submitting
func (be *BidEngine) processBid(order Order) BidResult {
    // Check if we already have a bid on this order
    existingBids, err := be.chainClient.GetProviderBids(be.ctx, be.config.ProviderAddress)
    if err != nil {
        return BidResult{Error: err}
    }

    for _, bid := range existingBids {
        if bid.OrderID == order.OrderID && bid.State == "open" {
            // Already have an active bid - skip
            return BidResult{
                OrderID: order.OrderID,
                BidID:   bid.BidID,
                Success: true,
            }
        }
    }

    // Proceed with bid submission
    // ...
}
```

### Workload Partitioning

For large-scale deployments, partition workload management across instances:

```yaml
# Provider daemon deployment with workload partitioning
apiVersion: apps/v1
kind: Deployment
metadata:
  name: provider-daemon
  namespace: virtengine
spec:
  replicas: 4
  template:
    spec:
      containers:
        - name: provider-daemon
          env:
            # Instance identifier for partitioning
            - name: INSTANCE_ID
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            # Total instances for consistent hashing
            - name: TOTAL_INSTANCES
              value: "4"
            # Partition mode: orders assigned by hash(orderID) % TOTAL_INSTANCES
            - name: PARTITION_MODE
              value: "consistent-hash"
```

---

## RPC Endpoint Load Balancing

### Full Node Fleet Architecture

```
                         ┌─────────────────────────────────────┐
                         │        External Clients              │
                         │  (SDKs, Wallets, Block Explorers)   │
                         └─────────────────────────────────────┘
                                           │
                                           ▼
                         ┌─────────────────────────────────────┐
                         │      Global Load Balancer            │
                         │   (Anycast DNS / CloudFlare)         │
                         └─────────────────────────────────────┘
                                           │
              ┌────────────────────────────┼────────────────────────────┐
              │                            │                            │
              ▼                            ▼                            ▼
    ┌──────────────────┐        ┌──────────────────┐        ┌──────────────────┐
    │ Regional LB      │        │ Regional LB      │        │ Regional LB      │
    │ (us-east)        │        │ (eu-west)        │        │ (ap-south)       │
    └──────────────────┘        └──────────────────┘        └──────────────────┘
              │                            │                            │
              ▼                            ▼                            ▼
    ┌──────────────────┐        ┌──────────────────┐        ┌──────────────────┐
    │ Istio Ingress    │        │ Istio Ingress    │        │ Istio Ingress    │
    │ Gateway          │        │ Gateway          │        │ Gateway          │
    └──────────────────┘        └──────────────────┘        └──────────────────┘
              │                            │                            │
    ┌─────────┴─────────┐        ┌─────────┴─────────┐        ┌─────────┴─────────┐
    │                   │        │                   │        │                   │
    ▼                   ▼        ▼                   ▼        ▼                   ▼
┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐
│ FN-1  │ │ FN-2  │ │ FN-3  │ │ FN-1  │ │ FN-2  │ │ FN-3  │ │ FN-1  │ │ FN-2  │
└───────┘ └───────┘ └───────┘ └───────┘ └───────┘ └───────┘ └───────┘ └───────┘
```

### Istio Configuration for RPC Load Balancing

#### DestinationRule with Load Balancing

```yaml
# deploy/istio/traffic/destinationrule-rpc.yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: virtengine-rpc
  namespace: virtengine
spec:
  host: virtengine-node.virtengine.svc.cluster.local
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
    loadBalancer:
      simple: LEAST_REQUEST # Prefer least loaded nodes
    connectionPool:
      tcp:
        maxConnections: 1000
        connectTimeout: 5s
      http:
        http1MaxPendingRequests: 500
        http2MaxRequests: 1000
        maxRequestsPerConnection: 100
        maxRetries: 3
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 10s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
      minHealthPercent: 30
  subsets:
    # Subset for query-heavy workloads (archive nodes)
    - name: archive
      labels:
        node-type: archive
      trafficPolicy:
        connectionPool:
          tcp:
            maxConnections: 500
          http:
            http1MaxPendingRequests: 200
    # Subset for real-time queries (pruned nodes)
    - name: realtime
      labels:
        node-type: pruned
      trafficPolicy:
        loadBalancer:
          simple: ROUND_ROBIN
```

#### VirtualService with Routing Rules

```yaml
# deploy/istio/traffic/virtualservice-rpc.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: virtengine-rpc
  namespace: virtengine
spec:
  hosts:
    - rpc.virtengine.network
    - virtengine-node.virtengine.svc.cluster.local
  gateways:
    - virtengine-gateway
  http:
    # Historical queries route to archive nodes
    - match:
        - uri:
            prefix: /cosmos/tx/v1beta1/txs
        - uri:
            prefix: /cosmos/base/tendermint/v1beta1/blocks
        - headers:
            x-query-type:
              exact: historical
      route:
        - destination:
            host: virtengine-node.virtengine.svc.cluster.local
            subset: archive
      timeout: 30s
      retries:
        attempts: 3
        perTryTimeout: 10s
        retryOn: gateway-error,connect-failure,refused-stream

    # Real-time queries route to pruned nodes
    - match:
        - uri:
            prefix: /cosmos/bank
        - uri:
            prefix: /cosmos/staking
        - uri:
            prefix: /virtengine
      route:
        - destination:
            host: virtengine-node.virtengine.svc.cluster.local
            subset: realtime
      timeout: 10s
      retries:
        attempts: 3
        perTryTimeout: 3s
        retryOn: gateway-error,connect-failure,refused-stream

    # Default routing with weighted distribution
    - route:
        - destination:
            host: virtengine-node.virtengine.svc.cluster.local
            subset: realtime
          weight: 80
        - destination:
            host: virtengine-node.virtengine.svc.cluster.local
            subset: archive
          weight: 20
      timeout: 15s
      retries:
        attempts: 3
        perTryTimeout: 5s
```

### External Load Balancer Configuration

#### AWS Application Load Balancer (Terraform)

```hcl
# infra/terraform/modules/alb/main.tf

resource "aws_lb" "rpc" {
  name               = "virtengine-rpc-${var.environment}"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = var.environment == "production"
  enable_http2               = true

  access_logs {
    bucket  = aws_s3_bucket.alb_logs.id
    prefix  = "rpc-alb"
    enabled = true
  }

  tags = {
    Name        = "virtengine-rpc-${var.environment}"
    Environment = var.environment
    Component   = "rpc-load-balancer"
  }
}

resource "aws_lb_target_group" "rpc" {
  name                 = "virtengine-rpc-tg-${var.environment}"
  port                 = 26657
  protocol             = "HTTP"
  vpc_id               = var.vpc_id
  target_type          = "ip"
  deregistration_delay = 30

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 10
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 3
  }

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400
    enabled         = false  # Disabled for round-robin
  }
}

resource "aws_lb_listener" "rpc_https" {
  load_balancer_arn = aws_lb.rpc.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.rpc.arn
  }
}

# Rate limiting with WAF
resource "aws_wafv2_web_acl" "rpc" {
  name  = "virtengine-rpc-waf-${var.environment}"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rate-limit"
    priority = 1

    override_action {
      none {}
    }

    statement {
      rate_based_statement {
        limit              = 2000  # requests per 5 minutes
        aggregate_key_type = "IP"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "VirtEngineRPCWAF"
    sampled_requests_enabled   = true
  }
}
```

#### gRPC Load Balancing (Network Load Balancer)

```hcl
# infra/terraform/modules/nlb/main.tf

resource "aws_lb" "grpc" {
  name               = "virtengine-grpc-${var.environment}"
  internal           = false
  load_balancer_type = "network"
  subnets            = var.public_subnet_ids

  enable_cross_zone_load_balancing = true

  tags = {
    Name        = "virtengine-grpc-${var.environment}"
    Environment = var.environment
    Component   = "grpc-load-balancer"
  }
}

resource "aws_lb_target_group" "grpc" {
  name                 = "virtengine-grpc-tg-${var.environment}"
  port                 = 9090
  protocol             = "TCP"
  vpc_id               = var.vpc_id
  target_type          = "ip"
  deregistration_delay = 30

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    unhealthy_threshold = 2
  }
}

resource "aws_lb_listener" "grpc" {
  load_balancer_arn = aws_lb.grpc.arn
  port              = "9090"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpc.arn
  }
}
```

---

## State Sync Optimization

State sync enables new validators and full nodes to quickly catch up to the current chain state without replaying all historical blocks.

### State Sync Architecture

```
                    ┌─────────────────────────────────────┐
                    │         Active Validator Set         │
                    │   (Producing blocks, creating       │
                    │    state sync snapshots)            │
                    └─────────────────────────────────────┘
                                      │
                                      │ Snapshot every 1000 blocks
                                      ▼
         ┌────────────────────────────────────────────────────────────────┐
         │                    State Sync Providers                        │
         │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
         │  │  Full Node 1 │  │  Full Node 2 │  │  Full Node N │          │
         │  │ (Snapshot)   │  │ (Snapshot)   │  │ (Snapshot)   │          │
         │  └──────────────┘  └──────────────┘  └──────────────┘          │
         └────────────────────────────────────────────────────────────────┘
                                      │
                                      │ P2P Discovery
                                      ▼
         ┌────────────────────────────────────────────────────────────────┐
         │                    New Node / Validator                        │
         │  1. Discover peers with snapshots                              │
         │  2. Download snapshot chunks in parallel                       │
         │  3. Verify snapshot against trust height                       │
         │  4. Apply snapshot and resume block sync                       │
         └────────────────────────────────────────────────────────────────┘
```

### Configuration for State Sync Providers

Configure full nodes to serve state sync snapshots:

```toml
# ~/.virtengine/config/app.toml

[state-sync]
# Create snapshots every N blocks
snapshot-interval = 1000

# Keep last N snapshots
snapshot-keep-recent = 5

# Discovery time for new nodes
discovery-time = "15s"

# Chunk fetch timeout
chunk-fetchers = 4
chunk-request-timeout = "10s"
```

### Configuration for New Validators

Fast bootstrap using state sync:

```toml
# ~/.virtengine/config/config.toml

[statesync]
enable = true

# RPC endpoints of state sync providers (at least 2)
rpc_servers = "https://rpc1.virtengine.network:443,https://rpc2.virtengine.network:443"

# Trust height and hash (from a trusted source)
trust_height = 5000000
trust_hash = "ABCD1234..."

# Trust period (typically unbonding period)
trust_period = "168h"

# Discovery timeout
discovery_time = "15s"

# Chunk fetching configuration
chunk_fetchers = "4"
chunk_request_timeout = "10s"
```

### Automated State Sync Bootstrap Script

```bash
#!/bin/bash
# scripts/bootstrap/state-sync-bootstrap.sh
# Automatically bootstrap a new node with state sync

set -euo pipefail

NODE_HOME="${NODE_HOME:-$HOME/.virtengine}"
RPC_SERVERS="${RPC_SERVERS:-https://rpc1.virtengine.network:443,https://rpc2.virtengine.network:443}"

# Fetch latest trusted block
echo "Fetching trust height and hash from RPC..."
LATEST=$(curl -s "${RPC_SERVERS%%,*}/block" | jq -r '.result.block.header.height')
TRUST_HEIGHT=$((LATEST - 1000))  # Trust a block ~1000 behind latest

TRUST_HASH=$(curl -s "${RPC_SERVERS%%,*}/block?height=${TRUST_HEIGHT}" | \
    jq -r '.result.block_id.hash')

echo "Trust Height: ${TRUST_HEIGHT}"
echo "Trust Hash: ${TRUST_HASH}"

# Update config
sed -i.bak \
    -e "s/^enable = false/enable = true/" \
    -e "s|^rpc_servers = .*|rpc_servers = \"${RPC_SERVERS}\"|" \
    -e "s/^trust_height = .*/trust_height = ${TRUST_HEIGHT}/" \
    -e "s/^trust_hash = .*/trust_hash = \"${TRUST_HASH}\"/" \
    "${NODE_HOME}/config/config.toml"

# Clear existing data (but keep keys)
echo "Clearing existing chain data..."
virtengine tendermint unsafe-reset-all --home "${NODE_HOME}" --keep-addr-book

echo "State sync configuration complete. Start the node to begin sync."
echo "Run: virtengine start --home ${NODE_HOME}"
```

### Snapshot Manager Configuration

Using the `pkg/pruning` snapshot manager:

```go
import "github.com/virtengine/virtengine/pkg/pruning"

// Configure snapshot manager for state sync providers
cfg := pruning.Config{
    Strategy:   pruning.StrategyDefault,
    KeepRecent: 362880,  // ~3 weeks
    Interval:   10,
    Snapshot: pruning.SnapshotConfig{
        Enabled:          true,
        Interval:         1000,  // Snapshot every 1000 blocks
        KeepRecent:       5,     // Keep 5 most recent
        Compression:      true,
        CompressionLevel: 6,
    },
}

// Validate state sync compatibility
if !cfg.IsStateSyncCompatible() {
    return fmt.Errorf("configuration not compatible with state sync")
}
```

### Monitoring State Sync Health

```yaml
# deploy/monitoring/prometheus/rules/statesync_alerts.yml
groups:
  - name: statesync
    rules:
      - alert: StateSyncSnapshotStale
        expr: time() - virtengine_statesync_snapshot_last_created_timestamp > 7200
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "State sync snapshot is stale"
          description: "No new snapshot created in the last 2 hours on {{ $labels.instance }}"

      - alert: StateSyncProviderDown
        expr: up{job="virtengine-statesync"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "State sync provider is down"
          description: "{{ $labels.instance }} is not responding"

      - alert: StateSyncChunksFailing
        expr: rate(virtengine_statesync_chunk_fetch_failures_total[5m]) > 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "State sync chunk fetches are failing"
          description: "High rate of chunk fetch failures on {{ $labels.instance }}"
```

---

## Sharding Strategy

### Current State: No Application-Level Sharding

VirtEngine uses CometBFT (Tendermint) consensus which requires all validators to process all transactions. True sharding at the blockchain level is not currently implemented.

### Available Scaling Approaches

#### 1. Read Replicas (Implemented)

- Multiple full nodes serve RPC queries
- Validators focus on block production
- Horizontal scaling of read capacity

#### 2. Workload Partitioning

- Provider daemons partition orders by consistent hashing
- ML inference distributed across workers

#### 3. Future: IBC-Based Sharding

For future scaling beyond current limits, consider IBC (Inter-Blockchain Communication):

```
                    ┌─────────────────────────────────────┐
                    │        VirtEngine Hub Chain          │
                    │   (Coordination, Settlement)         │
                    └─────────────────────────────────────┘
                                      │
                        IBC Connections
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
         ▼                            ▼                            ▼
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│ VE-Market Zone   │      │ VE-Identity Zone │      │ VE-Compute Zone  │
│ (Order/Bid       │      │ (VEID Processing)│      │ (HPC Workloads)  │
│  Processing)     │      │                  │      │                  │
└──────────────────┘      └──────────────────┘      └──────────────────┘
```

### Data Partitioning Strategies

#### By Module (Recommended for Future)

| Zone          | Modules               | Purpose                         |
| ------------- | --------------------- | ------------------------------- |
| Market Zone   | market, escrow        | High-throughput order matching  |
| Identity Zone | veid, encryption, mfa | Privacy-sensitive identity ops  |
| Compute Zone  | hpc                   | Resource-intensive compute jobs |

#### By Geography

```yaml
# Region-specific provider partitioning
providers:
  us-east:
    regions: ["us-east-1", "us-east-2", "us-west-1", "us-west-2"]
    validators: 3
  eu-west:
    regions: ["eu-west-1", "eu-west-2", "eu-central-1"]
    validators: 2
  ap-south:
    regions: ["ap-south-1", "ap-southeast-1", "ap-northeast-1"]
    validators: 2
```

---

## Multi-Region Deployment

### Region Topology

Based on the existing disaster recovery architecture:

| Region                | Role      | Validators | Full Nodes | Provider Daemons | Priority |
| --------------------- | --------- | ---------- | ---------- | ---------------- | -------- |
| US-EAST (us-east-1)   | Primary   | 2          | 4          | 4                | 1        |
| EU-WEST (eu-west-1)   | Secondary | 2          | 3          | 2                | 2        |
| AP-SOUTH (ap-south-1) | Tertiary  | 1          | 2          | 1                | 3        |

### Terraform Multi-Region Setup

```hcl
# infra/terraform/environments/production/main.tf

module "us_east" {
  source = "../../modules/region"

  region              = "us-east-1"
  environment         = "production"
  role                = "primary"

  validators          = 2
  full_nodes          = 4
  provider_daemons    = 4

  instance_types = {
    validator       = "r6i.2xlarge"
    full_node       = "r6i.xlarge"
    provider_daemon = "m6i.xlarge"
  }

  vpc_cidr = "10.0.0.0/16"
}

module "eu_west" {
  source = "../../modules/region"

  region              = "eu-west-1"
  environment         = "production"
  role                = "secondary"

  validators          = 2
  full_nodes          = 3
  provider_daemons    = 2

  instance_types = {
    validator       = "r6i.2xlarge"
    full_node       = "r6i.xlarge"
    provider_daemon = "m6i.xlarge"
  }

  vpc_cidr = "10.1.0.0/16"
}

module "ap_south" {
  source = "../../modules/region"

  region              = "ap-south-1"
  environment         = "production"
  role                = "tertiary"

  validators          = 1
  full_nodes          = 2
  provider_daemons    = 1

  instance_types = {
    validator       = "r6i.2xlarge"
    full_node       = "r6i.xlarge"
    provider_daemon = "m6i.xlarge"
  }

  vpc_cidr = "10.2.0.0/16"
}

# Cross-region VPC peering
module "vpc_peering" {
  source = "../../modules/vpc-peering"

  regions = {
    us_east = module.us_east.vpc_id
    eu_west = module.eu_west.vpc_id
    ap_south = module.ap_south.vpc_id
  }
}

# Global Accelerator for latency-based routing
resource "aws_globalaccelerator_accelerator" "virtengine" {
  name            = "virtengine-production"
  ip_address_type = "IPV4"
  enabled         = true
}

resource "aws_globalaccelerator_listener" "rpc" {
  accelerator_arn = aws_globalaccelerator_accelerator.virtengine.id
  protocol        = "TCP"

  port_range {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_endpoint_group" "us_east" {
  listener_arn = aws_globalaccelerator_listener.rpc.id

  endpoint_configuration {
    endpoint_id                    = module.us_east.alb_arn
    weight                         = 40
    client_ip_preservation_enabled = true
  }

  health_check_interval_seconds = 10
  health_check_path             = "/health"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}

resource "aws_globalaccelerator_endpoint_group" "eu_west" {
  listener_arn          = aws_globalaccelerator_listener.rpc.id
  endpoint_group_region = "eu-west-1"

  endpoint_configuration {
    endpoint_id                    = module.eu_west.alb_arn
    weight                         = 35
    client_ip_preservation_enabled = true
  }

  health_check_interval_seconds = 10
  health_check_path             = "/health"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}

resource "aws_globalaccelerator_endpoint_group" "ap_south" {
  listener_arn          = aws_globalaccelerator_listener.rpc.id
  endpoint_group_region = "ap-south-1"

  endpoint_configuration {
    endpoint_id                    = module.ap_south.alb_arn
    weight                         = 25
    client_ip_preservation_enabled = true
  }

  health_check_interval_seconds = 10
  health_check_path             = "/health"
  threshold_count               = 3
  traffic_dial_percentage       = 100
}
```

### DNS-Based Failover

```hcl
# Route53 health checks and failover routing
resource "aws_route53_health_check" "us_east" {
  fqdn              = "us-east.rpc.virtengine.network"
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 10

  tags = {
    Name = "virtengine-us-east-health"
  }
}

resource "aws_route53_record" "rpc_primary" {
  zone_id = aws_route53_zone.virtengine.zone_id
  name    = "rpc.virtengine.network"
  type    = "A"

  alias {
    name                   = aws_globalaccelerator_accelerator.virtengine.dns_name
    zone_id                = aws_globalaccelerator_accelerator.virtengine.hosted_zone_id
    evaluate_target_health = true
  }
}

# Geo-routing for regional endpoints
resource "aws_route53_record" "rpc_us" {
  zone_id        = aws_route53_zone.virtengine.zone_id
  name           = "rpc.virtengine.network"
  type           = "A"
  set_identifier = "us-east"

  geolocation_routing_policy {
    continent = "NA"
  }

  alias {
    name                   = module.us_east.alb_dns_name
    zone_id                = module.us_east.alb_zone_id
    evaluate_target_health = true
  }

  health_check_id = aws_route53_health_check.us_east.id
}
```

### Kubernetes Multi-Cluster Setup

```yaml
# deploy/argocd/multi-region.yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: virtengine-multi-region
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - cluster: us-east-1
            url: https://eks-us-east-1.virtengine.internal
            role: primary
            replicas:
              fullnode: 4
              provider: 4
          - cluster: eu-west-1
            url: https://eks-eu-west-1.virtengine.internal
            role: secondary
            replicas:
              fullnode: 3
              provider: 2
          - cluster: ap-south-1
            url: https://eks-ap-south-1.virtengine.internal
            role: tertiary
            replicas:
              fullnode: 2
              provider: 1
  template:
    metadata:
      name: "virtengine-{{cluster}}"
    spec:
      project: virtengine
      source:
        repoURL: https://github.com/virtengine/virtengine
        targetRevision: HEAD
        path: deploy/kubernetes/overlays/{{cluster}}
        helm:
          values: |
            region: {{cluster}}
            role: {{role}}
            fullNodeReplicas: {{replicas.fullnode}}
            providerReplicas: {{replicas.provider}}
      destination:
        server: "{{url}}"
        namespace: virtengine
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
```

---

## Auto-Scaling Policies

### Provider Daemon Auto-Scaling

#### Time-Based Scaling

```yaml
# deploy/kubernetes/overlays/production/cron-hpa.yaml
apiVersion: autoscaling.k8s.io/v1alpha1
kind: CronHorizontalPodAutoscaler
metadata:
  name: provider-daemon-time-scale
  namespace: virtengine
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: provider-daemon
  jobs:
    # Scale up during business hours (UTC)
    - name: "scale-up-business-hours"
      schedule: "0 8 * * 1-5" # 8 AM UTC, Mon-Fri
      targetSize: 6
      runOnce: false
    # Scale down after business hours
    - name: "scale-down-evening"
      schedule: "0 20 * * 1-5" # 8 PM UTC, Mon-Fri
      targetSize: 3
      runOnce: false
    # Minimal weekend scaling
    - name: "scale-down-weekend"
      schedule: "0 0 * * 0,6" # Midnight Saturday/Sunday
      targetSize: 2
      runOnce: false
```

#### Event-Driven Scaling with KEDA

```yaml
# deploy/kubernetes/overlays/production/keda-scaler.yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: provider-daemon-keda
  namespace: virtengine
spec:
  scaleTargetRef:
    name: provider-daemon
  pollingInterval: 15
  cooldownPeriod: 300
  minReplicaCount: 2
  maxReplicaCount: 20
  fallback:
    failureThreshold: 3
    replicas: 4
  triggers:
    # Scale based on Prometheus metrics
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring:9090
        metricName: virtengine_open_orders_total
        query: sum(virtengine_open_orders_total{status="pending"})
        threshold: "100"

    # Scale based on Redis queue depth
    - type: redis
      metadata:
        address: redis.virtengine:6379
        listName: provider:order:queue
        listLength: "50"

    # Scale based on active leases
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring:9090
        metricName: virtengine_active_leases_total
        query: sum(virtengine_active_leases_total{provider=~".+"})
        threshold: "500"
```

### Full Node Auto-Scaling

```yaml
# deploy/kubernetes/overlays/production/fullnode-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: fullnode-hpa
  namespace: virtengine
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: virtengine-fullnode
  minReplicas: 3
  maxReplicas: 12
  metrics:
    # Network traffic (RPC requests)
    - type: Pods
      pods:
        metric:
          name: rpc_requests_per_second
        target:
          type: AverageValue
          averageValue: "100"
    # Query latency
    - type: Pods
      pods:
        metric:
          name: rpc_latency_p99_ms
        target:
          type: AverageValue
          averageValue: "200"
    # Connection count
    - type: Pods
      pods:
        metric:
          name: active_connections
        target:
          type: AverageValue
          averageValue: "500"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 600 # 10 minutes
      policies:
        - type: Pods
          value: 1
          periodSeconds: 300
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60
```

### Capacity Planning Thresholds

| Metric             | Warning Threshold | Scale Trigger | Max Capacity |
| ------------------ | ----------------- | ------------- | ------------ |
| CPU Utilization    | 60%               | 70%           | 85%          |
| Memory Utilization | 70%               | 80%           | 90%          |
| Pending Orders     | 50/instance       | 100/instance  | 200/instance |
| Active Leases      | 75/instance       | 100/instance  | 150/instance |
| RPC Requests/s     | 80/instance       | 100/instance  | 150/instance |
| P99 Latency        | 150ms             | 200ms         | 500ms        |

---

## Monitoring and Observability

### Scaling Metrics Dashboard

```yaml
# deploy/monitoring/grafana/dashboards/scaling.json (excerpt)
{
  "dashboard":
    {
      "title": "VirtEngine Horizontal Scaling",
      "panels":
        [
          {
            "title": "Provider Daemon Replicas",
            "type": "timeseries",
            "targets":
              [
                {
                  "expr": 'kube_deployment_status_replicas{deployment="provider-daemon"}',
                  "legendFormat": "Current",
                },
                {
                  "expr": 'kube_deployment_spec_replicas{deployment="provider-daemon"}',
                  "legendFormat": "Desired",
                },
                {
                  "expr": 'kube_horizontalpodautoscaler_spec_min_replicas{horizontalpodautoscaler="provider-daemon-hpa"}',
                  "legendFormat": "Min",
                },
                {
                  "expr": 'kube_horizontalpodautoscaler_spec_max_replicas{horizontalpodautoscaler="provider-daemon-hpa"}',
                  "legendFormat": "Max",
                },
              ],
          },
          {
            "title": "Scaling Events",
            "type": "logs",
            "targets":
              [
                {
                  "expr": '{namespace="virtengine"} |= "Scaled" |= "HorizontalPodAutoscaler"',
                },
              ],
          },
          {
            "title": "Resource Utilization Heatmap",
            "type": "heatmap",
            "targets":
              [
                {
                  "expr": 'sum by (pod) (rate(container_cpu_usage_seconds_total{namespace="virtengine"}[5m])) / sum by (pod) (kube_pod_container_resource_limits{resource="cpu", namespace="virtengine"})',
                },
              ],
          },
        ],
    },
}
```

### Alerting Rules for Scaling

```yaml
# deploy/monitoring/prometheus/rules/scaling_alerts.yml
groups:
  - name: scaling
    rules:
      - alert: HPAMaxedOut
        expr: |
          kube_horizontalpodautoscaler_status_current_replicas{namespace="virtengine"} 
          == kube_horizontalpodautoscaler_spec_max_replicas{namespace="virtengine"}
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "HPA is at maximum replicas"
          description: "{{ $labels.horizontalpodautoscaler }} has been at max replicas for 15 minutes"
          runbook_url: "https://docs.virtengine.network/runbooks/hpa-maxed"

      - alert: HighPendingOrders
        expr: sum(virtengine_open_orders_total{status="pending"}) > 500
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High number of pending orders"
          description: "{{ $value }} orders pending - may need to increase provider capacity"

      - alert: ScalingFailure
        expr: |
          increase(kube_horizontalpodautoscaler_status_condition{status="false", condition="ScalingActive"}[5m]) > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "HPA scaling is failing"
          description: "{{ $labels.horizontalpodautoscaler }} is unable to scale"

      - alert: InsufficientClusterCapacity
        expr: |
          sum(kube_pod_status_phase{phase="Pending", namespace="virtengine"}) > 0
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Pods pending due to insufficient capacity"
          description: "{{ $value }} pods are pending - cluster may need more nodes"
```

---

## Operational Runbooks

### Runbook: Manual Scale-Up

````markdown
## Manual Provider Daemon Scale-Up

### When to Use

- Anticipated high-traffic event
- HPA is at max but more capacity is needed
- Proactive scaling before maintenance

### Steps

1. **Verify current state**
   ```bash
   kubectl get hpa provider-daemon-hpa -n virtengine
   kubectl get pods -l app.kubernetes.io/name=provider-daemon -n virtengine
   ```
````

2. **Increase HPA max if needed**

   ```bash
   kubectl patch hpa provider-daemon-hpa -n virtengine \
     --patch '{"spec":{"maxReplicas":25}}'
   ```

3. **Scale deployment directly (if immediate scaling needed)**

   ```bash
   kubectl scale deployment provider-daemon -n virtengine --replicas=15
   ```

4. **Verify scaling**

   ```bash
   kubectl rollout status deployment/provider-daemon -n virtengine
   kubectl get pods -l app.kubernetes.io/name=provider-daemon -n virtengine
   ```

5. **Monitor health**
   ```bash
   kubectl logs -l app.kubernetes.io/name=provider-daemon -n virtengine --tail=100 -f
   ```

### Rollback

```bash
kubectl scale deployment provider-daemon -n virtengine --replicas=4
kubectl patch hpa provider-daemon-hpa -n virtengine \
  --patch '{"spec":{"maxReplicas":10}}'
```

````

### Runbook: Adding a New Region

```markdown
## Adding a New Deployment Region

### Prerequisites
- AWS/GCP account with required permissions
- Terraform state access
- ArgoCD access

### Steps

1. **Create Terraform configuration**
   ```bash
   cd infra/terraform/environments/production

   # Add new region module
   cat >> main.tf << 'EOF'
   module "new_region" {
     source = "../../modules/region"
     region = "sa-east-1"
     environment = "production"
     role = "quaternary"
     # ... configuration
   }
   EOF
````

2. **Plan and apply infrastructure**

   ```bash
   terraform plan -out=new-region.plan
   terraform apply new-region.plan
   ```

3. **Create Kubernetes overlay**

   ```bash
   mkdir -p deploy/kubernetes/overlays/sa-east-1
   # Copy and customize from existing region
   cp -r deploy/kubernetes/overlays/us-east-1/* deploy/kubernetes/overlays/sa-east-1/
   # Edit kustomization.yaml with region-specific values
   ```

4. **Add to ArgoCD ApplicationSet**

   ```yaml
   # Add to generators.list.elements in multi-region.yaml
   - cluster: sa-east-1
     url: https://eks-sa-east-1.virtengine.internal
     role: quaternary
     replicas:
       fullnode: 2
       provider: 1
   ```

5. **Sync and verify**

   ```bash
   argocd app sync virtengine-sa-east-1
   kubectl --context=sa-east-1 get pods -n virtengine
   ```

6. **Update DNS and load balancer**

   ```bash
   cd infra/terraform/environments/production
   terraform apply -target=aws_route53_record.rpc_sa
   ```

7. **Verify connectivity**
   ```bash
   curl https://sa-east-1.rpc.virtengine.network/health
   ```

````

### Runbook: State Sync Recovery

```markdown
## Fast Node Recovery with State Sync

### When to Use
- New validator setup
- Node data corruption
- Fast region failover

### Steps

1. **Stop the node**
   ```bash
   systemctl stop virtengine
````

2. **Get trust height and hash**

   ```bash
   TRUSTED_RPC="https://rpc.virtengine.network"
   LATEST=$(curl -s ${TRUSTED_RPC}/block | jq -r '.result.block.header.height')
   TRUST_HEIGHT=$((LATEST - 1000))
   TRUST_HASH=$(curl -s "${TRUSTED_RPC}/block?height=${TRUST_HEIGHT}" | jq -r '.result.block_id.hash')

   echo "Trust Height: ${TRUST_HEIGHT}"
   echo "Trust Hash: ${TRUST_HASH}"
   ```

3. **Configure state sync**

   ```bash
   sed -i.bak \
     -e "s/^enable = false/enable = true/" \
     -e "s|^rpc_servers = .*|rpc_servers = \"${TRUSTED_RPC}:443,${TRUSTED_RPC}:443\"|" \
     -e "s/^trust_height = .*/trust_height = ${TRUST_HEIGHT}/" \
     -e "s/^trust_hash = .*/trust_hash = \"${TRUST_HASH}\"/" \
     ~/.virtengine/config/config.toml
   ```

4. **Reset state (keep keys)**

   ```bash
   virtengine tendermint unsafe-reset-all --keep-addr-book
   ```

5. **Start node and monitor**

   ```bash
   systemctl start virtengine
   journalctl -u virtengine -f | grep -E "(state-sync|Applied snapshot|Completed)"
   ```

6. **Verify sync completion**
   ```bash
   virtengine status | jq '.sync_info'
   # catching_up should be false
   ```

### Expected Timeline

- Snapshot discovery: 1-2 minutes
- Chunk download: 5-15 minutes (depending on state size)
- Block catch-up: 1-5 minutes
- Total: ~20 minutes vs hours for full sync

```

---

## Related Documentation

- [Disaster Recovery Plan](disaster-recovery.md)
- [Business Continuity Plan](business-continuity.md)
- [Validator Onboarding Guide](validator-onboarding.md)
- [State Pruning Package](../pkg/pruning/README.md)
- [Istio Service Mesh](../deploy/istio/README.md)
- [Provider Daemon Architecture](../pkg/provider_daemon/doc.go)

---

**Document Owner**: Infrastructure Team
**Last Updated**: 2026-01-30
**Next Review**: 2026-04-30
**Classification**: Internal
```
