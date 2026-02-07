# Chaos Engineering Runbook

**Version:** 1.0.0  
**Date:** 2026-01-31  
**Task Reference:** TEST-INFRA-9A

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Experiment Catalog](#experiment-catalog)
4. [Execution Procedures](#execution-procedures)
5. [Rollback Procedures](#rollback-procedures)
6. [Post-Experiment Analysis](#post-experiment-analysis)
7. [Resilience Metrics and SLOs](#resilience-metrics-and-slos)
8. [Troubleshooting](#troubleshooting)

---

## Overview

This runbook documents the chaos engineering program for VirtEngine, providing procedures for executing controlled failure experiments to validate system resilience.

### Purpose

- Validate system behavior under adverse conditions
- Identify weaknesses before they cause production incidents
- Build confidence in the system's ability to withstand turbulent conditions
- Improve incident response procedures

### Scope

The chaos engineering program covers:

- **Network Chaos**: Partitions, latency, packet loss, bandwidth limits
- **Node Failures**: Pod crashes, container failures, node drains
- **Resource Exhaustion**: CPU, memory, disk, and I/O stress
- **Byzantine Behavior**: Double-signing, equivocation, invalid blocks
- **Time Chaos**: Clock skew, time jumps

### Chaos Engineering Principles

1. **Start with a hypothesis**: Define what you expect to happen
2. **Minimize blast radius**: Start small, in non-production environments
3. **Automate experiments**: Make experiments repeatable
4. **Measure everything**: Collect metrics before, during, and after
5. **Rollback quickly**: Have clear abort criteria and rollback procedures

---

## Prerequisites

### Infrastructure Requirements

- Kubernetes cluster (1.26+) with Chaos Mesh or Litmus installed
- Prometheus/Grafana for metrics collection
- Access to staging environment
- kubectl configured with appropriate permissions

### Install Chaos Mesh

```bash
# Add Chaos Mesh Helm repo
helm repo add chaos-mesh https://charts.chaos-mesh.org
helm repo update

# Install Chaos Mesh
helm install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace=chaos-mesh \
  --create-namespace \
  --set chaosDaemon.runtime=containerd \
  --set chaosDaemon.socketPath=/run/containerd/containerd.sock

# Verify installation
kubectl get pods -n chaos-mesh
```

### Install LitmusChaos

```bash
# Add Litmus Helm repo
helm repo add litmuschaos https://litmuschaos.github.io/litmus-helm/
helm repo update

# Install Litmus
helm install chaos litmuschaos/litmus \
  --namespace=litmus \
  --create-namespace

# Install Litmus experiments
kubectl apply -f https://hub.litmuschaos.io/api/chaos/3.0.0?file=charts/generic/experiments.yaml \
  -n litmus
```

### Apply VirtEngine Chaos Resources

```bash
# Apply Chaos Mesh experiments
kubectl apply -k infra/kubernetes/chaos/chaos-mesh/

# Apply Litmus workflows
kubectl apply -k infra/kubernetes/chaos/litmus/
```

---

## Experiment Catalog

### Network Chaos Experiments

| Experiment                  | Description                             | Duration | Blast Radius | Risk Level |
| --------------------------- | --------------------------------------- | -------- | ------------ | ---------- |
| `validator-partition`       | Split validators into isolated groups   | 5 min    | Medium       | Medium     |
| `validator-latency`         | Inject 500ms latency between validators | 5 min    | Low          | Low        |
| `validator-packet-loss`     | Drop 10% of packets                     | 5 min    | Low          | Low        |
| `validator-bandwidth-limit` | Limit to 1 Mbps                         | 5 min    | Low          | Low        |
| `provider-isolation`        | Isolate provider daemon from network    | 2 min    | Low          | Low        |
| `cross-zone-latency`        | Add 100ms cross-zone latency            | 10 min   | Medium       | Low        |

### Pod Chaos Experiments

| Experiment                | Description                      | Duration | Blast Radius | Risk Level |
| ------------------------- | -------------------------------- | -------- | ------------ | ---------- |
| `validator-pod-kill`      | Kill a single validator pod      | Instant  | Low          | Medium     |
| `validator-pod-failure`   | Fail a validator pod             | 1 min    | Low          | Medium     |
| `provider-pod-kill`       | Kill provider daemon pod         | Instant  | Low          | Low        |
| `multi-validator-failure` | Kill 2 validators simultaneously | Instant  | Medium       | High       |
| `rolling-failure`         | Sequential 25% pod failures      | 5 min    | Medium       | Medium     |

### Stress Chaos Experiments

| Experiment                | Description                   | Duration | Blast Radius | Risk Level |
| ------------------------- | ----------------------------- | -------- | ------------ | ---------- |
| `validator-cpu-stress`    | 80% CPU load on one validator | 2 min    | Low          | Low        |
| `validator-memory-stress` | Consume 512MB memory          | 2 min    | Low          | Medium     |
| `provider-cpu-stress`     | 90% CPU load on provider      | 3 min    | Low          | Low        |
| `ml-inference-stress`     | Stress ML scoring containers  | 1 min    | Low          | Low        |
| `oom-test`                | Push memory to trigger OOM    | 1 min    | Low          | High       |

### Time Chaos Experiments

| Experiment            | Description                           | Duration | Blast Radius | Risk Level |
| --------------------- | ------------------------------------- | -------- | ------------ | ---------- |
| `clock-skew-forward`  | Advance time by 1 hour                | 2 min    | Low          | Medium     |
| `clock-skew-backward` | Rewind time by 30 min                 | 2 min    | Low          | Medium     |
| `clock-drift`         | Add 5s drift to all validators        | 5 min    | Medium       | Low        |
| `consensus-time-test` | 10s clock offset for block validation | 1 min    | Low          | Medium     |

### Byzantine Behavior Experiments

| Experiment                | Description                    | Duration | Blast Radius | Risk Level |
| ------------------------- | ------------------------------ | -------- | ------------ | ---------- |
| `double-signing-test`     | Simulate double-signing attack | 5 min    | Low          | High       |
| `equivocation-prevote`    | Send conflicting prevotes      | 5 min    | Low          | Medium     |
| `equivocation-precommit`  | Send conflicting precommits    | 5 min    | Low          | Medium     |
| `invalid-block-malformed` | Propose malformed blocks       | 5 min    | Low          | Medium     |
| `message-tampering`       | Corrupt consensus messages     | 5 min    | Low          | Medium     |

---

## Execution Procedures

### Pre-Experiment Checklist

Before running any chaos experiment:

- [ ] Verify experiment is approved for the target environment
- [ ] Confirm current system health (all SLOs green)
- [ ] Notify relevant stakeholders (on-call, team leads)
- [ ] Verify rollback procedures are ready
- [ ] Enable enhanced monitoring/alerting
- [ ] Take baseline metrics snapshot
- [ ] Confirm abort criteria are defined

### Running Experiments with Chaos Mesh

```bash
# List available experiments
kubectl get networkchaos,podchaos,stresschaos,timechaos -n chaos-testing

# Start a network partition experiment
kubectl apply -f infra/kubernetes/chaos/chaos-mesh/network-chaos.yaml

# Check experiment status
kubectl get networkchaos validator-partition -n chaos-testing -o yaml

# View experiment events
kubectl describe networkchaos validator-partition -n chaos-testing

# Stop experiment (delete the resource)
kubectl delete networkchaos validator-partition -n chaos-testing
```

### Running Experiments with Litmus

```bash
# List available experiments
kubectl get chaosengine,chaosexperiment -n litmus

# Start a chaos workflow
kubectl apply -f infra/kubernetes/chaos/litmus/workflows/validator-chaos.yaml

# Check workflow status
kubectl get cronworkflow -n litmus

# Manually trigger a workflow
kubectl create workflow --from=cronworkflow/validator-chaos-workflow -n litmus

# View workflow logs
kubectl logs -n litmus -l workflows.argoproj.io/workflow=validator-chaos-workflow
```

### Running Experiments Programmatically

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/virtengine/virtengine/pkg/chaos"
    "github.com/virtengine/virtengine/pkg/chaos/runner"
    "github.com/virtengine/virtengine/pkg/chaos/scenarios"
)

func main() {
    // Create a Kubernetes runner
    k8sRunner := runner.NewKubernetesRunner(
        runner.WithNamespace("chaos-testing"),
        runner.WithChaosProvider("chaos-mesh"),
    )

    // Create controller
    controller := chaos.NewController(
        chaos.WithRunner(k8sRunner),
    )

    // Build a network partition scenario
    scenario := scenarios.NewValidatorPartition(
        [][]string{
            {"validator-0", "validator-1"},
            {"validator-2", "validator-3"},
        },
        5*time.Minute,
    )

    exp, err := scenario.Build()
    if err != nil {
        log.Fatal(err)
    }

    // Convert to ExperimentSpec
    spec := &chaos.ExperimentSpec{
        ID:       "exp-" + time.Now().Format("20060102-150405"),
        Name:     exp.Name,
        Type:     chaos.ExperimentTypeNetworkPartition,
        Duration: exp.Duration,
    }

    // Execute experiment
    ctx := context.Background()
    results, err := controller.Execute(ctx, spec)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Experiment completed: success=%v, duration=%v",
        results.Success, results.Duration)
}
```

### Abort Criteria

Immediately abort experiments if:

- [ ] Chain halt detected (no blocks for > 30 seconds)
- [ ] More than 1/3 validators offline
- [ ] SLO breach exceeds error budget
- [ ] Unexpected cascading failures observed
- [ ] On-call escalation triggered

---

## Rollback Procedures

### Immediate Rollback (Chaos Mesh)

```bash
# Delete all active chaos experiments
kubectl delete networkchaos,podchaos,stresschaos,timechaos --all -n chaos-testing

# Verify cleanup
kubectl get networkchaos,podchaos,stresschaos,timechaos -n chaos-testing
```

### Immediate Rollback (Litmus)

```bash
# Stop all chaos engines
kubectl patch chaosengine --all -n virtengine -p '{"spec":{"engineState":"stop"}}' --type=merge

# Delete chaos engines
kubectl delete chaosengine --all -n virtengine

# Delete chaos results
kubectl delete chaosresult --all -n virtengine
```

### Manual Recovery Procedures

If automatic recovery fails:

1. **For Pod Failures**:

   ```bash
   # Force restart affected pods
   kubectl delete pod -l app.kubernetes.io/component=validator -n virtengine --force

   # Verify pod recovery
   kubectl get pods -l app.kubernetes.io/component=validator -n virtengine -w
   ```

2. **For Network Partitions**:

   ```bash
   # Check network policies
   kubectl get networkpolicy -n virtengine

   # Remove any blocking policies
   kubectl delete networkpolicy chaos-partition -n virtengine --ignore-not-found

   # Restart affected pods to clear network state
   kubectl rollout restart statefulset/virtengine-validator -n virtengine
   ```

3. **For Resource Exhaustion**:

   ```bash
   # Identify stressed pods
   kubectl top pods -n virtengine

   # Restart affected pods
   kubectl delete pod <stressed-pod> -n virtengine
   ```

---

## Post-Experiment Analysis

### Metrics to Collect

| Metric                       | Source             | Purpose                                |
| ---------------------------- | ------------------ | -------------------------------------- |
| Block production rate        | Prometheus         | Measure chain health during experiment |
| Consensus round duration     | Prometheus         | Detect consensus slowdown              |
| Message drop rate            | Prometheus         | Quantify network impact                |
| MTTR (Mean Time To Recovery) | Experiment results | Recovery performance                   |
| MTTD (Mean Time To Detect)   | Alertmanager       | Detection effectiveness                |
| SLO burn rate                | Prometheus         | Error budget consumption               |

### Analysis Checklist

After each experiment:

- [ ] Collect timeline of events
- [ ] Calculate MTTR and MTTD
- [ ] Analyze SLO impact
- [ ] Review alerting effectiveness (Did we get paged?)
- [ ] Document unexpected behaviors
- [ ] Identify improvement opportunities
- [ ] Update runbooks with findings

### Report Template

```markdown
# Chaos Experiment Report

**Experiment**: [Name]
**Date**: [YYYY-MM-DD]
**Duration**: [X minutes]
**Environment**: [staging/production]
**Operator**: [Name]

## Hypothesis

[What we expected to happen]

## Results

- **Success**: [Yes/No]
- **MTTR**: [X seconds]
- **MTTD**: [X seconds]
- **SLO Impact**: [X% error budget consumed]

## Timeline

| Time | Event              |
| ---- | ------------------ |
| T+0  | Experiment started |
| T+X  | [Event]            |
| T+Y  | Experiment ended   |

## Observations

[What actually happened, unexpected behaviors]

## Lessons Learned

[What we learned, what to improve]

## Action Items

- [ ] [Action item 1]
- [ ] [Action item 2]
```

---

## Resilience Metrics and SLOs

### Target SLOs

| Service      | Metric                | Target       | Error Budget/Month |
| ------------ | --------------------- | ------------ | ------------------ |
| Chain        | Availability          | 99.9%        | 43.2 minutes       |
| Chain        | Block time P99        | < 6 seconds  | -                  |
| VEID Scoring | Processing P95        | < 5 minutes  | -                  |
| VEID Scoring | Availability          | 99.5%        | 3.6 hours          |
| Marketplace  | Order fulfillment P95 | < 10 minutes | -                  |
| Marketplace  | Availability          | 99.5%        | 3.6 hours          |
| HPC          | Job scheduling P95    | < 15 minutes | -                  |
| HPC          | Availability          | 99.0%        | 7.2 hours          |

### Resilience Baselines

| Baseline        | MTTR Target | MTTD Target | Recovery Rate |
| --------------- | ----------- | ----------- | ------------- |
| Chain           | 30 seconds  | 5 seconds   | 99%           |
| VEID Scoring    | 2 minutes   | 30 seconds  | 95%           |
| Marketplace     | 5 minutes   | 1 minute    | 95%           |
| HPC             | 10 minutes  | 2 minutes   | 90%           |
| Provider Daemon | 1 minute    | 30 seconds  | 95%           |

### Prometheus Queries

```promql
# Chain availability
1 - (
  sum(rate(tendermint_consensus_rounds_total{result="timeout"}[5m]))
  /
  sum(rate(tendermint_consensus_rounds_total[5m]))
)

# Block time P99
histogram_quantile(0.99,
  sum(rate(tendermint_consensus_block_time_seconds_bucket[1h])) by (le)
)

# Chaos experiment success rate
sum(rate(virtengine_chaos_experiments_total{status="success"}[24h]))
/
sum(rate(virtengine_chaos_experiments_total[24h]))

# Mean Time To Recovery
histogram_quantile(0.50,
  sum(rate(virtengine_chaos_mean_time_to_recovery_seconds_bucket[24h])) by (le)
)

# Active chaos experiments
virtengine_chaos_experiments_active
```

---

## Troubleshooting

### Common Issues

#### Chaos Mesh Controller Not Running

```bash
# Check controller status
kubectl get pods -n chaos-mesh

# Check controller logs
kubectl logs -n chaos-mesh -l app.kubernetes.io/component=controller-manager

# Restart controller
kubectl rollout restart deployment/chaos-controller-manager -n chaos-mesh
```

#### Experiment Not Applying

```bash
# Check experiment status
kubectl describe networkchaos <name> -n chaos-testing

# Check for webhook issues
kubectl logs -n chaos-mesh -l app.kubernetes.io/component=chaos-daemon

# Verify target pods exist
kubectl get pods -n virtengine -l <label-selector>
```

#### Experiments Not Cleaning Up

```bash
# Force delete stuck experiments
kubectl delete networkchaos <name> -n chaos-testing --force --grace-period=0

# Check for finalizers
kubectl get networkchaos <name> -n chaos-testing -o jsonpath='{.metadata.finalizers}'

# Remove finalizers if stuck
kubectl patch networkchaos <name> -n chaos-testing -p '{"metadata":{"finalizers":null}}' --type=merge
```

#### Recovery Taking Too Long

1. Check pod health: `kubectl get pods -n virtengine`
2. Check for pending restarts: `kubectl describe pod <pod> -n virtengine`
3. Check resource constraints: `kubectl top pods -n virtengine`
4. Verify network connectivity: `kubectl exec -n virtengine <pod> -- ping <target>`

### Emergency Contacts

| Role               | Contact   | Escalation Time |
| ------------------ | --------- | --------------- |
| On-call            | PagerDuty | Immediate       |
| Platform Team Lead | [Name]    | 15 minutes      |
| SRE Manager        | [Name]    | 30 minutes      |
| CTO                | [Name]    | Critical only   |

---

## Appendix

### Useful Commands

```bash
# Quick status check
kubectl get pods -n virtengine -o wide
kubectl get networkchaos,podchaos,stresschaos -n chaos-testing

# Live watch of chaos experiments
watch -n 2 kubectl get networkchaos,podchaos,stresschaos,timechaos -n chaos-testing

# Export experiment results
kubectl get chaosresult -n virtengine -o json > chaos-results-$(date +%Y%m%d).json

# Clean up all chaos resources
kubectl delete networkchaos,podchaos,stresschaos,timechaos,chaosengine --all -A
```

### References

- [Chaos Mesh Documentation](https://chaos-mesh.org/docs/)
- [LitmusChaos Documentation](https://docs.litmuschaos.io/)
- [Principles of Chaos Engineering](https://principlesofchaos.org/)
- [VirtEngine SLOs and Playbooks](../../slos-and-playbooks.md)
