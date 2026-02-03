# TEE Incident Response Runbook

## Overview

This runbook covers incidents involving TEE hardware, enclave health, and attestation verification for the VEID pipeline. It focuses on failures impacting SGX, SEV-SNP, and AWS Nitro platforms used by the `tee-enclave` service in the `virtengine` namespace.

## Triggers

- `PrimaryTEEPlatformDown`, `TEEFailoverFailed`, or `AllEnclavesDown` alerts.
- Attestation verification errors (quote/VCEK/NSM failures, TCB below minimum).
- Enclave pods crashlooping or failing readiness/liveness checks.
- Measurement allowlist mismatch or unregistered measurement.
- Hardware device missing on nodes (`/dev/nsm`, `/dev/sev-guest`, `/dev/sgx_enclave`).

## Immediate Actions

1. Confirm blast radius and current platform:
   ```bash
   kubectl get pods -n virtengine -l app.kubernetes.io/name=tee-enclave -o wide
   kubectl get nodes -l virtengine.io/enclave-ready=true -o wide
   ```
2. Check enclave service health and recent errors:
   ```bash
   kubectl logs -n virtengine -l app.kubernetes.io/name=tee-enclave --tail=200
   kubectl describe deployment -n virtengine tee-enclave
   ```
3. Verify attestation endpoint responsiveness:
   ```bash
   kubectl exec -n virtengine -it deploy/tee-enclave -- /bin/attestation-test
   curl -X POST http://tee-enclave:8080/v1/attestation/verify -d '{"nonce":"test"}'
   ```
4. Confirm chain node is healthy:
   ```bash
   virtengined status
   ```

## Diagnostics

### Platform Health and Hardware Presence

```bash
# On affected nodes (SSH or privileged debug pod)
ls -la /dev/nsm /dev/sev-guest /dev/sgx_enclave
dmesg | grep -i "sgx\|sev"
```

### Attestation and Measurement Governance

```bash
# Check allowlist config
kubectl get configmap -n virtengine tee-measurement-allowlist -o yaml

# Validate registered measurements on-chain
virtengine query veid registered-enclaves --platform nitro
virtengine query veid registered-enclaves --platform sev-snp
virtengine query veid registered-enclaves --platform sgx
```

### Enclave Logs and Health Checks

```bash
kubectl exec -n virtengine -it deploy/tee-enclave -- /bin/tee-health-check
kubectl logs -n virtengine -l app.kubernetes.io/name=tee-enclave --since=10m | grep -i "attestation\|tcb\|measurement\|quote"
```

### SGX / SEV-SNP / Nitro Specific Checks

```bash
# SGX
sudo systemctl status aesmd
curl -k https://localhost:8081/sgx/certification/v4/rootcacrl

# SEV-SNP
ls -la /dev/sev*
snpguest report attestation.bin request.bin --random
snpguest fetch vcek der attestation.bin certs/

# Nitro
nitro-cli --version
```

## Remediation

### Restart or Reschedule Enclave Pods

```bash
kubectl delete pod -n virtengine -l app.kubernetes.io/name=tee-enclave
kubectl rollout restart deployment/tee-enclave -n virtengine
```

### Force Failover Away from a Failing Platform

```bash
kubectl cordon -l virtengine.io/tee-platform=nitro
kubectl drain -l virtengine.io/tee-platform=nitro \
  --pod-selector=app.kubernetes.io/name=tee-enclave \
  --grace-period=60 \
  --delete-emptydir-data
kubectl scale deployment/tee-enclave -n virtengine --replicas=2
```

### Patch Node Affinity to Exclude a Platform

```bash
kubectl patch deployment/tee-enclave -n virtengine --type=json -p='[
  {"op": "add", "path": "/spec/template/spec/affinity/nodeAffinity/requiredDuringSchedulingIgnoredDuringExecution/nodeSelectorTerms/0/matchExpressions/-",
   "value": {"key": "virtengine.io/tee-platform", "operator": "NotIn", "values": ["nitro"]}}
]'
```

### Update Measurement Allowlist or Re-Register Enclave

```bash
# Update allowlist ConfigMap via GitOps or kubectl apply (preferred)
kubectl apply -f deploy/kubernetes/base/tee-enclave-configmap.yaml

# Register updated measurement on-chain
virtengine tx veid register-enclave \
  --platform sgx \
  --mrenclave <mrenclave> \
  --mrsigner <mrsigner> \
  --min-isvsvn 1 \
  --from <validator-key> \
  --chain-id <chain-id> \
  --gas auto \
  --gas-adjustment 1.5 \
  -y
```

### Attestation Endpoint or Cache Issues

```bash
kubectl get configmap -n virtengine tee-enclave-config -o yaml | grep VIRTENGINE_TEE_ATTESTATION_ENDPOINT
kubectl rollout restart deployment/tee-enclave -n virtengine
```

## Recovery

1. Validate platform recovery and restore scheduling:
   ```bash
   kubectl exec -n virtengine -it deploy/tee-enclave -- /bin/tee-health-check
   kubectl exec -n virtengine -it deploy/tee-enclave -- /bin/attestation-test
   kubectl uncordon -l virtengine.io/tee-platform=nitro
   kubectl get pods -n virtengine -l app.kubernetes.io/name=tee-enclave -o wide
   ```
2. Confirm failover state cleared and attestation succeeds:
   ```bash
   curl -X POST http://tee-enclave:8080/v1/attestation/verify -d '{"nonce":"recovery"}'
   ```
3. Monitor metrics and logs for at least 30 minutes after recovery.

## Communications

- **Initial notification:** On-call engineering, security, and platform teams.
- **Status updates:** Every 30–60 minutes for P1/P2; include current platform, attestation status, and failover state.
- **Resolution message:** Summarize impact, root cause, mitigation, and recovery steps.
- **Post-incident:** File a follow-up ticket with logs, timeline, and any measurement or TCB changes.

## Related References

- `_docs/tee-deployment-guide.md`
- `_docs/tee-failover-strategy.md`
- `deploy/monitoring/alerts/enclave-health.yaml`
