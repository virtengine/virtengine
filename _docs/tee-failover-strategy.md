# VirtEngine TEE Failover Strategy

## Overview

This document describes the failover strategy for VirtEngine's Trusted Execution Environment (TEE) infrastructure across multiple hardware platforms: AWS Nitro, AMD SEV-SNP, and Intel SGX.

The goal is to ensure continuous availability of TEE-protected identity verification even when individual platforms or nodes experience failures.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Platform Priority Order](#platform-priority-order)
3. [Failover Triggers](#failover-triggers)
4. [Automatic Failover](#automatic-failover)
5. [Manual Failover](#manual-failover)
6. [Recovery Procedures](#recovery-procedures)
7. [Testing Failover](#testing-failover)
8. [Monitoring and Alerts](#monitoring-and-alerts)
9. [Runbooks](#runbooks)

---

## Architecture Overview

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    Load Balancer                        │
                    │              (TEE-aware routing)                        │
                    └─────────────────────┬───────────────────────────────────┘
                                          │
                    ┌─────────────────────┼───────────────────────────────────┐
                    │                     │                                    │
         ┌──────────▼──────────┐ ┌────────▼────────┐ ┌────────────────────────▼─┐
         │   AWS Nitro Pool    │ │ AMD SEV-SNP Pool│ │    Intel SGX Pool        │
         │    (Primary)        │ │   (Secondary)   │ │    (Tertiary)            │
         ├─────────────────────┤ ├─────────────────┤ ├──────────────────────────┤
         │ nitro-node-1        │ │ sev-node-1      │ │ sgx-node-1               │
         │ nitro-node-2        │ │ sev-node-2      │ │ sgx-node-2               │
         │ nitro-node-3        │ │ sev-node-3      │ │ sgx-node-3               │
         └─────────────────────┘ └─────────────────┘ └──────────────────────────┘
```

### Key Components

| Component | Description |
|-----------|-------------|
| TEE Enclave Service | Kubernetes Deployment with multi-platform support |
| Platform Selector | Kubernetes node affinity rules for platform preference |
| Health Checker | Continuous monitoring of enclave health and attestation |
| Failover Controller | Automated failover logic in enclave manager |
| Attestation Verifier | Platform-agnostic attestation verification |

---

## Platform Priority Order

VirtEngine uses a prioritized platform selection strategy:

| Priority | Platform | Use Case | AWS Instance Types |
|----------|----------|----------|-------------------|
| 1 (Primary) | AWS Nitro | AWS deployments | c5, c6i, m5, r5 |
| 2 (Secondary) | AMD SEV-SNP | Multi-cloud, Azure/GCP | c6a, m6a |
| 3 (Tertiary) | Intel SGX | On-premises, legacy | c5, c6i (SGX-enabled) |

### Selection Criteria

1. **Hardware Availability**: Platform must have functioning hardware
2. **Attestation Health**: Platform must pass attestation verification
3. **TCB Version**: Platform must meet minimum TCB requirements
4. **Capacity**: Platform must have available enclave slots

---

## Failover Triggers

### Automatic Triggers

| Trigger | Condition | Action |
|---------|-----------|--------|
| Hardware Failure | Device not responding | Immediate failover |
| Attestation Failure | Quote generation fails | Retry 3x, then failover |
| TCB Outdated | Version below minimum | Graceful failover |
| Node Failure | Kubernetes node unhealthy | Reschedule to healthy node |
| Health Check Failure | 3 consecutive failures | Pod restart, then failover |

### Manual Triggers

- Platform maintenance (security patches)
- Capacity rebalancing
- Disaster recovery testing
- Compliance requirements

---

## Automatic Failover

### Kubernetes-Level Failover

The Kubernetes Deployment handles node-level failures automatically:

```yaml
# Pod Anti-Affinity ensures distribution across nodes
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: tee-enclave
          topologyKey: kubernetes.io/hostname

# Node Affinity with preference order
  nodeAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100  # Prefer Nitro
        preference:
          matchExpressions:
            - key: virtengine.io/tee-platform
              operator: In
              values: ["nitro"]
      - weight: 80   # Then SEV-SNP
        preference:
          matchExpressions:
            - key: virtengine.io/tee-platform
              operator: In
              values: ["sev-snp"]
      - weight: 60   # Then SGX
        preference:
          matchExpressions:
            - key: virtengine.io/tee-platform
              operator: In
              values: ["sgx"]
```

### Application-Level Failover

The enclave manager handles platform-level failover:

```go
// Failover sequence in EnclaveManager
func (m *EnclaveManager) handlePlatformFailure(failedPlatform PlatformType) error {
    // 1. Mark failed platform as unavailable
    m.platformHealth[failedPlatform] = false
    
    // 2. Select next available platform by priority
    nextPlatform := m.selectHealthyPlatform()
    if nextPlatform == "" {
        return ErrNoPlatformAvailable
    }
    
    // 3. Initialize new platform
    if err := m.initializePlatform(nextPlatform); err != nil {
        return fmt.Errorf("failover to %s failed: %w", nextPlatform, err)
    }
    
    // 4. Verify attestation on new platform
    if err := m.verifyAttestation(nextPlatform); err != nil {
        return fmt.Errorf("attestation on %s failed: %w", nextPlatform, err)
    }
    
    // 5. Emit failover event
    m.emitEvent(FailoverComplete{
        From: failedPlatform,
        To:   nextPlatform,
    })
    
    return nil
}
```

### Failover Timing

| Phase | Duration | Description |
|-------|----------|-------------|
| Detection | 10-30s | Health check failure detection |
| Decision | <1s | Platform selection |
| Initialization | 5-30s | New enclave startup |
| Attestation | 2-10s | Attestation verification |
| Total | 17-71s | End-to-end failover time |

---

## Manual Failover

### Procedure: Drain Platform

To gracefully drain a TEE platform for maintenance:

```bash
# 1. Cordon nodes on the platform
kubectl cordon -l virtengine.io/tee-platform=nitro

# 2. Drain enclave pods (respects PDB)
kubectl drain -l virtengine.io/tee-platform=nitro \
  --pod-selector=app.kubernetes.io/name=tee-enclave \
  --grace-period=60 \
  --delete-emptydir-data

# 3. Verify pods have migrated
kubectl get pods -l app.kubernetes.io/name=tee-enclave -o wide

# 4. Perform maintenance...

# 5. Uncordon nodes
kubectl uncordon -l virtengine.io/tee-platform=nitro
```

### Procedure: Force Failover

To force immediate failover (emergency):

```bash
# 1. Scale down on failing platform
kubectl scale deployment/tee-enclave --replicas=0

# 2. Update node affinity to exclude platform
kubectl patch deployment/tee-enclave --type=json -p='[
  {"op": "add", "path": "/spec/template/spec/affinity/nodeAffinity/requiredDuringSchedulingIgnoredDuringExecution/nodeSelectorTerms/0/matchExpressions/-", 
   "value": {"key": "virtengine.io/tee-platform", "operator": "NotIn", "values": ["nitro"]}}
]'

# 3. Scale back up
kubectl scale deployment/tee-enclave --replicas=2

# 4. Verify attestation on new platform
curl -X POST http://tee-enclave:8080/v1/attestation/verify -d '{"nonce": "test"}'
```

---

## Recovery Procedures

### Platform Recovery

After a failed platform becomes available again:

```bash
# 1. Verify hardware health
kubectl exec -it tee-enclave-xxx -- /bin/tee-health-check

# 2. Verify attestation
kubectl exec -it tee-enclave-xxx -- /bin/attestation-test

# 3. Uncordon nodes
kubectl uncordon -l virtengine.io/tee-platform=nitro

# 4. Kubernetes will automatically prefer recovered platform due to affinity weights
```

### State Recovery

TEE enclaves are stateless - all state is:
- Derived from blockchain state
- Sealed with platform-independent keys
- Reconstructible from on-chain data

Recovery steps:
1. New enclave initializes
2. Loads model weights from sealed storage
3. Derives session keys from blockchain state
4. Begins processing requests

---

## Testing Failover

### Chaos Testing

```bash
# Test node failure (using chaos-mesh)
kubectl apply -f - <<EOF
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: tee-enclave-pod-kill
spec:
  action: pod-kill
  mode: one
  selector:
    labelSelectors:
      app.kubernetes.io/name: tee-enclave
      virtengine.io/tee-platform: nitro
  duration: '30s'
EOF
```

### Attestation Failure Test

```bash
# Simulate attestation failure
kubectl exec -it tee-enclave-xxx -- \
  curl -X POST localhost:8080/debug/simulate-attestation-failure

# Verify failover occurred
kubectl logs tee-enclave-xxx | grep -i failover
```

### Full Platform Failure Test

```bash
# 1. Cordon all Nitro nodes
kubectl cordon -l virtengine.io/tee-platform=nitro

# 2. Delete Nitro pods
kubectl delete pods -l app.kubernetes.io/name=tee-enclave,virtengine.io/tee-platform=nitro

# 3. Verify pods scheduled on SEV-SNP/SGX
kubectl get pods -l app.kubernetes.io/name=tee-enclave -o wide

# 4. Verify attestation works
curl -X POST http://tee-enclave:8080/v1/attestation/generate

# 5. Restore Nitro nodes
kubectl uncordon -l virtengine.io/tee-platform=nitro
```

---

## Monitoring and Alerts

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `virtengine_enclave_platform_available` | Platform availability | 0 for any platform |
| `virtengine_enclave_failover_active` | Failover in progress | 1 (informational) |
| `virtengine_enclave_failover_failures_total` | Failed failovers | > 0 |
| `virtengine_enclave_failover_duration_seconds` | Failover duration | > 120s |
| `virtengine_enclave_platform_transitions_total` | Platform switches | Unusual spikes |

### Grafana Dashboard

Dashboard available at: `https://grafana.virtengine.com/d/tee-failover`

Panels:
- Platform availability heatmap
- Failover events timeline
- Attestation success rate by platform
- Node distribution across platforms

### Alert Rules

See `deploy/monitoring/alerts/enclave-health.yml` for:
- `PrimaryTEEPlatformDown`
- `TEEFailoverInProgress`
- `TEEFailoverFailed`
- `AllEnclavesDown`

---

## Runbooks

### Runbook: Platform Unavailable

**Trigger**: `PrimaryTEEPlatformDown` alert

**Steps**:
1. Check AWS/cloud provider status page
2. Verify node group health: `kubectl get nodes -l virtengine.io/tee-platform=nitro`
3. Check node events: `kubectl describe node <node-name>`
4. If hardware issue, escalate to cloud provider
5. If software issue, restart node group

### Runbook: Failover Failed

**Trigger**: `TEEFailoverFailed` alert

**Steps**:
1. Check enclave logs: `kubectl logs -l app.kubernetes.io/name=tee-enclave`
2. Verify alternative platforms are healthy
3. Check attestation services (PCCS, KDS) connectivity
4. Manual failover if needed (see above)
5. If all platforms unavailable, escalate immediately

### Runbook: All Enclaves Down

**Trigger**: `AllEnclavesDown` alert

**Priority**: P0 - Immediate response required

**Steps**:
1. Check cluster health: `kubectl get nodes`
2. Check namespace: `kubectl get all -n virtengine`
3. Check deployment: `kubectl describe deployment tee-enclave`
4. Check events: `kubectl get events --sort-by=.lastTimestamp`
5. If node issue, scale up node groups
6. If image issue, rollback: `kubectl rollout undo deployment/tee-enclave`
7. Notify incident team

### Runbook: TCB Update Required

**Trigger**: `TCBVersionOutOfDate` alert

**Steps**:
1. Schedule maintenance window
2. Update node group AMIs with new firmware
3. Rolling restart of nodes
4. Verify new TCB version in attestation
5. Update minimum TCB version in config

---

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VIRTENGINE_TEE_FAILOVER_ENABLED` | `true` | Enable automatic failover |
| `VIRTENGINE_TEE_FAILOVER_THRESHOLD` | `3` | Failures before failover |
| `VIRTENGINE_TEE_FAILOVER_COOLDOWN` | `300s` | Cooldown between failovers |
| `VIRTENGINE_TEE_PLATFORM_PRIORITY` | `nitro,sev-snp,sgx` | Platform priority order |

### Terraform Variables

```hcl
# Enable/disable platforms
enable_nitro   = true   # Primary
enable_sev_snp = true   # Secondary
enable_sgx     = false  # Tertiary (on-demand)

# Minimum nodes per platform (for HA)
nitro_min_size   = 2
sev_snp_min_size = 2
sgx_min_size     = 2
```

---

## Appendix: Platform Compatibility Matrix

| Feature | Nitro | SEV-SNP | SGX |
|---------|-------|---------|-----|
| Memory Encryption | Yes | Yes | Yes (EPC) |
| Attestation | NSM | VCEK | DCAP |
| Cloud Support | AWS | Azure, GCP | Limited |
| On-Prem Support | No | Yes | Yes |
| Key Derivation | KMS | Hardware | Hardware |
| vTPM Support | Yes | Yes | No |

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-30 | VirtEngine Team | Initial release (TEE-HW-001) |
