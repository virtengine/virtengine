# Provider Operator Cheat Sheet

**Quick Reference for VirtEngine Providers**

---

## Essential Commands

### Provider Status
```bash
# Check provider daemon status
systemctl status provider-daemon

# API health check
curl -s localhost:8443/api/v1/status | jq .

# List active leases
virtengine query market leases --provider $(virtengine keys show provider -a)
```

### Bid Engine
```bash
# Check bid engine status
curl -s localhost:8443/api/v1/bid-engine/status | jq .

# List pending bids
virtengine query market bids --provider $(virtengine keys show provider -a) --state open

# Bid history
virtengine query market bids --provider $(virtengine keys show provider -a) --limit 50
```

### Workload Management
```bash
# List workloads
curl -s localhost:8443/api/v1/workloads | jq .

# Workload logs
curl -s localhost:8443/api/v1/workloads/{id}/logs

# Workload metrics
curl -s localhost:8443/api/v1/workloads/{id}/metrics
```

---

## Configuration

### Key Files
```
/etc/provider-daemon/
├── config.yaml          # Main configuration
├── pricing.yaml         # Pricing configuration
└── adapters/
    ├── kubernetes.yaml  # K8s adapter config
    └── ansible.yaml     # Ansible adapter config
```

### Pricing Configuration
```yaml
# pricing.yaml
pricing:
  cpu_millicores: 0.000001uve     # Per millicore per block
  memory_mb: 0.0000001uve         # Per MB per block
  storage_mb: 0.00000001uve       # Per MB per block
  gpu: 0.001uve                   # Per GPU per block
  
  profit_margin: 0.15             # 15% margin
```

---

## Workload Lifecycle

```
Pending → Deploying → Running → Stopping → Stopped → Terminated
    │         │          │                      │
    └─────────┴──────────┴──► Failed ◄──────────┘
```

### State Transitions
| From | To | Trigger |
|------|-----|---------|
| Pending | Deploying | Lease accepted |
| Deploying | Running | Deploy success |
| Running | Stopping | Close request |
| Running | Paused | Pause request |
| Any | Failed | Error condition |

---

## Monitoring

### Key Metrics
```promql
# Active leases
virtengine_provider_active_leases

# Bid success rate
rate(virtengine_bid_success_total[1h]) / rate(virtengine_bid_total[1h])

# Workload count by state
virtengine_provider_workloads{state="running"}

# Revenue rate
sum(rate(virtengine_escrow_settlement_amount[1d]))
```

### Resource Usage
```bash
# Kubernetes resources
kubectl get pods -n provider-workloads -o wide

# Resource quotas
kubectl describe resourcequota -n provider-workloads
```

---

## Escrow & Settlement

### Check Escrow
```bash
# Query escrow balance
virtengine query escrow accounts --provider $(virtengine keys show provider -a)

# Check specific lease escrow
virtengine query escrow account --account-id <escrow-id>
```

### Usage Submission
```bash
# Usage is automatic, but can check:
journalctl -u provider-daemon | grep "usage submitted"
```

---

## Emergency Procedures

### Bid Engine Stuck
```bash
# Check logs
journalctl -u provider-daemon -n 100 | grep -i bid

# Restart daemon
sudo systemctl restart provider-daemon

# Verify recovery
curl -s localhost:8443/api/v1/bid-engine/status
```

### Workload Deployment Failed
```bash
# Check adapter logs
journalctl -u provider-daemon | grep -i adapter

# Check Kubernetes events
kubectl get events -n provider-workloads --sort-by=.lastTimestamp

# Manual cleanup if needed
kubectl delete pod <pod-name> -n provider-workloads
```

### Low Escrow Balance
```bash
# Check escrow
virtengine query escrow account --account-id <escrow-id>

# Notify tenant (if low)
# Consider graceful shutdown if critically low
```

---

## Adapter Quick Reference

### Kubernetes
```bash
# Verify connection
kubectl cluster-info

# Check provider namespace
kubectl get all -n provider-workloads

# Pod logs
kubectl logs <pod-name> -n provider-workloads
```

### Common Adapter Issues
| Issue | Symptom | Fix |
|-------|---------|-----|
| Auth failure | "unauthorized" errors | Refresh kubeconfig |
| Quota exceeded | Deploy fails | Increase quotas |
| Image pull fail | ImagePullBackOff | Check registry access |
| Network issues | CrashLoopBackOff | Check network policies |

---

## Key Ports

| Port | Service | Access |
|------|---------|--------|
| 8443 | Provider API | Internal |
| 8444 | Metrics | Monitoring |
| 6443 | K8s API | Adapter |

---

## Contact

- **Support**: #providers channel
- **Emergency**: ops@virtengine.com
- **Docs**: _docs/provider-guide.md

---

*Keep this handy during operations!*
