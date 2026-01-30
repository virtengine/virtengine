# Runbook: Provider Deployment Failures

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | ProviderDeploymentFailures |
| Severity | Warning |
| Service | provider-daemon |
| Tier | Tier 1 |
| SLO Impact | SLO-PROVIDER-001 (Deployment success rate) |

## Summary

This alert fires when the provider deployment success rate falls below threshold or when there's a significant increase in deployment failures. Provider deployments are workloads (containers, VMs) provisioned by providers for marketplace leases.

## Impact

- **Medium**: Tenants cannot deploy workloads
- **Medium**: Revenue loss for providers
- **Low**: Does not affect chain consensus
- **Low**: Existing deployments continue running

## Prerequisites

- SSH access to provider nodes
- `provider-daemon` CLI access
- Access to infrastructure adapters (K8s, OpenStack, etc.)

## Diagnostic Steps

### 1. Check Deployment Failure Rate

```bash
# Check recent failure rate
curl -s "http://prometheus:9090/api/v1/query?query=rate(virtengine_provider_deployment_failures_total[5m])" | jq '.data.result'

# Check failure reasons
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_provider_deployment_failures_total[5m]))by(reason)" | jq '.data.result'
```

### 2. Check Provider Daemon Status

```bash
# SSH to provider
ssh user@provider-node

# Check daemon status
systemctl status provider-daemon
provider-daemon status

# Check recent deployments
provider-daemon list deployments --limit 20 --status failed
```

### 3. Check Logs

```bash
# Recent deployment errors
journalctl -u provider-daemon -n 200 | grep -i "error\|fail"

# Check specific deployment
provider-daemon logs deployment <deployment-id>

# Check Loki for patterns
# Query: {service="provider-daemon"} |= "deployment_failed" | json
```

### 4. Check Infrastructure Adapter

```bash
# Kubernetes adapter
kubectl get pods -n virtengine-workloads
kubectl get events -n virtengine-workloads --sort-by='.lastTimestamp' | tail -20

# Check for resource issues
kubectl describe nodes | grep -A 5 "Allocated resources"
```

## Resolution Steps

### Scenario 1: Resource Exhaustion

**Symptoms**: Failures mention "insufficient resources", "no available nodes"

```bash
# 1. Check cluster resources
kubectl top nodes
kubectl describe nodes | grep -E "Allocatable|Allocated" -A 5

# 2. Check pending pods
kubectl get pods -n virtengine-workloads --field-selector=status.phase=Pending

# 3. If node pressure, add capacity or evict low-priority workloads
kubectl cordon <node>  # Prevent new scheduling
kubectl drain <node> --ignore-daemonsets  # If maintenance needed

# 4. For immediate relief, increase cluster autoscaler max
kubectl patch configmap cluster-autoscaler -n kube-system --type merge -p '{"data":{"max-nodes":"20"}}'
```

### Scenario 2: Image Pull Failures

**Symptoms**: Failures mention "ImagePullBackOff", "ErrImagePull"

```bash
# 1. Check failed pods
kubectl get pods -n virtengine-workloads | grep -E "ImagePull|ErrImage"

# 2. Check image details
kubectl describe pod <pod-name> -n virtengine-workloads | grep -A 10 "Events:"

# 3. Verify registry connectivity
curl -I https://registry.example.com/v2/

# 4. Check registry credentials
kubectl get secret regcred -n virtengine-workloads -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d

# 5. Update registry credentials if expired
kubectl delete secret regcred -n virtengine-workloads
kubectl create secret docker-registry regcred \
  --docker-server=registry.example.com \
  --docker-username=$USER \
  --docker-password=$PASS \
  -n virtengine-workloads
```

### Scenario 3: Network Issues

**Symptoms**: Failures mention "network", "timeout", "connection refused"

```bash
# 1. Check pod networking
kubectl run test-net --rm -it --image=busybox -- wget -qO- http://kubernetes.default.svc

# 2. Check CNI status
kubectl get pods -n kube-system -l k8s-app=calico-node  # or your CNI

# 3. Check DNS
kubectl run test-dns --rm -it --image=busybox -- nslookup kubernetes.default

# 4. If CNI issues, restart CNI pods
kubectl delete pods -n kube-system -l k8s-app=calico-node
```

### Scenario 4: Manifest Validation Failures

**Symptoms**: Failures at "manifest validation" stage

```bash
# 1. Check rejected manifests
provider-daemon list manifests --status rejected --limit 10

# 2. Get validation errors
provider-daemon manifest validate --file /tmp/failed-manifest.yaml

# 3. Common issues:
# - Invalid resource requests
# - Unsupported features
# - Security policy violations

# 4. Update manifest parser if needed
provider-daemon config set manifest.strict_validation false  # Temporary
```

### Scenario 5: Bid Engine Issues

**Symptoms**: Deployments stuck at "bidding" stage

```bash
# 1. Check bid engine status
provider-daemon bid-engine status

# 2. Check pending bids
provider-daemon list bids --status pending

# 3. Check on-chain order matching
virtengined q market orders --state open --limit 10

# 4. Restart bid engine
provider-daemon bid-engine restart

# 5. If stuck bids, clear queue
provider-daemon bid-engine clear-pending --confirm
```

### Scenario 6: Infrastructure Adapter Errors

```bash
# For Kubernetes adapter
# 1. Check adapter logs
provider-daemon adapter-logs kubernetes --tail 100

# 2. Verify kubeconfig
kubectl cluster-info

# 3. Restart adapter
provider-daemon adapter restart kubernetes

# For OpenStack/AWS/Azure adapters
# 1. Check Waldur connectivity
curl -s http://waldur-api:8080/api/health | jq

# 2. Verify credentials
provider-daemon adapter test-auth openstack

# 3. Check quotas
openstack quota show
```

## Recovery Verification

```bash
# 1. Verify failure rate is decreasing
watch -n 30 'curl -s "http://prometheus:9090/api/v1/query?query=rate(virtengine_provider_deployment_failures_total[5m])" | jq ".data.result[0].value[1]"'

# 2. Test deployment
provider-daemon deploy test --manifest /etc/provider-daemon/test-manifest.yaml

# 3. Check deployment success
provider-daemon list deployments --limit 5 --status active

# 4. Verify metrics
curl -s http://localhost:26661/metrics | grep virtengine_provider_deployment
```

## Escalation

**Escalate to L2 if**:
- Failure rate > 50% for more than 15 minutes
- Infrastructure adapter completely down
- Cannot identify root cause

**Escalate to Infrastructure Team if**:
- Cluster resource exhaustion
- Network infrastructure issues
- Cloud provider issues

## Communication

### Provider Status Update

```
Provider Deployment Alert

Provider ID: [provider-address]
Status: [Investigating | Degraded | Resolved]
Failure Rate: [X]%
Impact: New deployments may fail

Actions:
- Investigating root cause
- [Current mitigation steps]

Existing deployments are not affected.
```

## Prevention

### Regular Health Checks

```bash
# Add to cron/scheduled job
#!/bin/bash
provider-daemon health-check --full
provider-daemon adapter test-connectivity --all
provider-daemon resources check --warn-threshold 80
```

### Capacity Monitoring

Set up alerts for:
- Node CPU/memory utilization > 80%
- Disk usage > 85%
- Pending pods > 10 for > 5 minutes

## Related Alerts

- `ProviderNodeDown` - Provider node offline
- `ProviderResourceExhausted` - Resource capacity issues
- `ProviderBidEngineStuck` - Bid engine problems
- `LeaseEndingWithoutReplacement` - Lease management

## References

- [Provider Operations Guide](../../provider-operations.md)
- [Deployment Troubleshooting](../../deployment-troubleshooting.md)
- [Infrastructure Adapters](../../../pkg/provider_daemon/README.md)
