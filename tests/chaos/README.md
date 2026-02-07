# Chaos Testing Framework

## Overview

Chaos testing framework for VirtEngine using Chaos Mesh to validate system resilience under failures.

## Prerequisites

- Kubernetes cluster
- Chaos Mesh installed
- kubectl configured

## Install Chaos Mesh

```bash
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace chaos-testing --create-namespace
```

## Available Experiments

### Network Partition

Simulates network partition between validator nodes.

```bash
kubectl apply -f infra/chaos-mesh/network-partition.yaml
```

### Pod Failure

Kills provider daemon pods randomly.

```bash
kubectl apply -f infra/chaos-mesh/pod-failure.yaml
```

### Memory Stress

Applies memory pressure to nodes.

```bash
kubectl apply -f infra/chaos-mesh/stress-test.yaml
```

### Latency Injection

Adds latency to ML inference service.

```bash
kubectl apply -f infra/chaos-mesh/latency-injection.yaml
```

## Health Checkers

The framework includes health checkers to monitor system behavior:

- **ChainHealthChecker**: Monitors block production
- **APIEndpointChecker**: Monitors API availability
- **TransactionSubmitChecker**: Monitors transaction submission

## Running Experiments

Experiments are scheduled automatically or can be run manually:

```bash
kubectl apply -f infra/chaos-mesh/network-partition.yaml
sleep 120
kubectl delete -f infra/chaos-mesh/network-partition.yaml
```

## Resilience Targets

| Scenario | Max Downtime | Recovery Time |
|----------|--------------|---------------|
| Pod Kill | 0s | < 30s |
| Network Partition | N/A (consensus) | < 60s after heal |
| Memory Stress | 0s (degraded) | < 30s after release |

## CI/CD Integration

Chaos tests run weekly via GitHub Actions. See `.github/workflows/chaos-test.yaml`.
