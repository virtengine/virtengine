# VirtEngine Deployment and Recovery Guide

This guide is the operator-facing path that matches the A18 infrastructure
automation and the B18 repo-owned deploy and recovery assets.

## Canonical Paths

- Infrastructure parity gate: [`infra/scripts/check-environment-parity.sh`](../../infra/scripts/check-environment-parity.sh)
- Reviewed Terraform plan/apply wrapper: [`infra/scripts/terraform-run.sh`](../../infra/scripts/terraform-run.sh)
- Kubernetes platform overlays: [`deploy/kubernetes/overlays`](../../deploy/kubernetes/overlays)
- ArgoCD bootstrap and applications: [`deploy/argocd`](../../deploy/argocd)
- DR drill wrapper: [`infra/dr/run-failover-drill.sh`](../../infra/dr/run-failover-drill.sh)
- Backup and restore scripts: [`scripts/dr`](../../scripts/dr)
- Rollback scripts: [`scripts/rollback`](../../scripts/rollback)

## Deployment Model

- `dev`, `staging`, and `prod` are separate environments with their own
  Terraform state backends and reviewed-plan workflow approvals.
- The runtime namespace is `virtengine` in every environment. Environment
  separation happens at the cluster and Terraform layer, not via suffixed
  runtime namespaces.
- Validator workloads schedule onto nodes labeled `role=validator` and
  `virtengine.com/chain=true`, with the `dedicated=validator:NoSchedule` taint.
- Provider workloads schedule onto nodes labeled `role=workload`.
- External secrets resolve from `secret/virtengine/{env}/{service}` and are
  patched in each Kustomize overlay.

## 1. Validate Infra Parity

Run this before any cluster bootstrap or environment rollout:

```bash
./infra/scripts/check-environment-parity.sh
```

The parity gate fails if the Terraform env layout, region layout, workflow
versions, or critical infra trust configuration drift from the A18 contract.

## 2. Produce a Reviewed Terraform Plan

Use the wrapper instead of direct `terraform plan` or `terraform apply` so the
artifact set is reproducible and checksum-backed.

```bash
./infra/scripts/terraform-run.sh plan infra/terraform/environments/staging output/infra/staging-plan
```

Review:

- `output/infra/staging-plan/plan.txt`
- `output/infra/staging-plan/plan.json`
- `output/infra/staging-plan/tfplan.sha256`
- `output/infra/staging-plan/manifest.env`

Apply only the reviewed plan artifact:

```bash
./infra/scripts/terraform-run.sh apply infra/terraform/environments/staging output/infra/staging-plan
```

For drift-only checks:

```bash
./infra/scripts/terraform-run.sh drift infra/terraform/environments/prod output/infra/prod-drift
```

## 3. Bootstrap ArgoCD

The repo-owned ArgoCD base now pulls the pinned upstream HA install manifest and
applies the local config, project, and ingress resources in one step:

```bash
kubectl apply -k deploy/argocd/base
kubectl apply -f deploy/argocd/apps/applicationsets.yaml
```

What this creates:

- ArgoCD control plane in `argocd`
- `virtengine-platform-{env}` applications that point at
  `deploy/kubernetes/overlays/{env}`
- `monitoring-{env}` applications for `deploy/monitoring/overlays/{env}`

## 4. Deploy the Platform Manifests

Direct apply remains available for emergency or break-glass use:

```bash
kubectl apply -k deploy/kubernetes/overlays/dev
kubectl apply -k deploy/kubernetes/overlays/staging
kubectl apply -k deploy/kubernetes/overlays/prod
```

The overlays now align with the infrastructure model:

- the chain node is a valid `StatefulSet`
- service discovery uses stable names with no overlay name prefixes
- readiness and liveness probes are defined for validator, provider, and TEE
  paths
- production blue/green routing references the real `virtengine` namespace
- prod-only autoscaling resources no longer point at non-existent RPC workloads

## 5. Secrets and Runtime Dependencies

Before syncing an environment, ensure these secret paths exist:

- `secret/virtengine/dev/virtengine-node`
- `secret/virtengine/dev/provider-daemon`
- `secret/virtengine/dev/database`
- `secret/virtengine/staging/virtengine-node`
- `secret/virtengine/staging/provider-daemon`
- `secret/virtengine/staging/database`
- `secret/virtengine/prod/virtengine-node`
- `secret/virtengine/prod/provider-daemon`
- `secret/virtengine/prod/database`

Operator expectation:

- ESO or Vault wiring supplies the secret material.
- Pods do not rely on blank IRSA stubs in the manifest layer.
- Provider ingress TLS is handled by the deployed service/backend path instead
  of an unset ACM annotation.

## 6. Backup, Restore, and DR Validation

Run the backup scripts directly when needed:

```bash
./scripts/dr/backup-chain-state.sh
./scripts/dr/backup-provider-state.sh
./scripts/dr/backup-keys.sh --type all
```

Run the continuous DR validation suite:

```bash
./scripts/dr/dr-test.sh --environment staging --report
```

Run a failover drill and keep the emitted evidence bundle:

```bash
./infra/dr/run-failover-drill.sh --mode rehearsal --output-dir output/drill/rehearsal
./infra/dr/run-failover-drill.sh --mode live --live-validation --output-dir output/drill/live
```

Evidence produced by the drill wrapper:

- `failover-drill.log`
- `failover-drill-summary.md`
- `failover-drill-evidence.json`

## 7. Rollback Paths

### ArgoCD Revision Rollback

```bash
./scripts/rollback/argocd-rollback.sh virtengine-platform-prod --dry-run
./scripts/rollback/argocd-rollback.sh virtengine-platform-prod --yes
```

Artifacts land under `output/rollback/argocd/...` unless `--artifact-dir` is
provided.

### Blue/Green Traffic Shift

Gradual shifts require Prometheus so the script can fail closed on elevated 5xx
rates:

```bash
PROMETHEUS_URL=https://prometheus.monitoring.svc.cluster.local:9090 \
./scripts/rollback/blue-green-switch.sh provider-daemon green --mode gradual --yes
```

For emergency break-glass:

```bash
./scripts/rollback/blue-green-switch.sh provider-daemon blue --mode instant --yes
```

### Terraform Backend State Rollback

```bash
./scripts/rollback/terraform-rollback.sh prod --steps-back 1 --dry-run
./scripts/rollback/terraform-rollback.sh prod --steps-back 1 --yes
```

This script now:

- resolves the real backend bucket/key from `infra/terraform/environments/{env}`
- downloads the exact target object version
- captures the current state first
- pushes the rollback state only on non-dry-run execution
- generates a reviewed post-rollback plan via `infra/scripts/terraform-run.sh`

## 8. Operator Validation Checklist

Before calling an environment deployable, verify:

1. `infra/scripts/check-environment-parity.sh` passes.
2. The reviewed Terraform plan artifact exists and matches its checksum.
3. `kubectl kustomize deploy/kubernetes/overlays/{env}` renders cleanly.
4. `kubectl kustomize deploy/argocd/base` renders cleanly.
5. `scripts/dr/dr-test.sh --environment {env} --report` passes for the target
   environment or has an explicit, recorded blocker.
6. Rollback operators can run the `--dry-run` mode of each rollback helper
   without missing dependencies or unresolved inputs.
