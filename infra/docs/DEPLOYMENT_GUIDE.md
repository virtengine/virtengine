# VirtEngine Production Deployment Guide

## Overview

This guide covers the complete deployment process for VirtEngine infrastructure using Terraform IaC and GitOps with ArgoCD.

## Prerequisites

### Required Tools

```bash
# Terraform
terraform version  # >= 1.6.0

# AWS CLI
aws --version  # >= 2.0

# kubectl
kubectl version --client  # >= 1.28

# Helm
helm version  # >= 3.12

# ArgoCD CLI (optional)
argocd version
```

### CI Enforcement

The `Infrastructure` GitHub Actions workflow runs before infrastructure
security, plan, apply, and drift jobs. Its `validate` job installs Node 20 and
checksum-pinned kubectl 1.29.0, then runs:

```bash
node scripts/task85c-validate-kubernetes.mjs
```

The gate triggers on `infra/**` (including this deployment guide), `deploy/**`,
`scripts/task85c-validate-kubernetes.mjs`,
`docs/documentation/INFRA-001-IMPLEMENTATION-SUMMARY.md`, focused Task 85C
ADR/completion/recovery docs, and the infrastructure workflow itself. The
validation job uses only read access and does not require AWS, ArgoCD, or other
deployment credentials.

### AWS Permissions

The deploying IAM user/role needs:
- `AdministratorAccess` OR
- Custom policy with:
  - EC2, EKS, IAM, S3, KMS, VPC full access
  - CloudWatch, WAF, DynamoDB access

## Deployment Steps

### 1. Bootstrap Terraform State

For first-time setup, create the state bucket manually:

```bash
# Create state bucket
aws s3api create-bucket \
  --bucket virtengine-terraform-state-prod \
  --region us-west-2 \
  --create-bucket-configuration LocationConstraint=us-west-2

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket virtengine-terraform-state-prod \
  --versioning-configuration Status=Enabled

# Create DynamoDB lock table
aws dynamodb create-table \
  --table-name virtengine-terraform-locks-prod \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST
```

### 2. Deploy Infrastructure

```bash
# Navigate to environment
cd infra/terraform/environments/prod

# Initialize Terraform
terraform init

# Review plan
terraform plan -out=tfplan

# Apply (requires approval)
terraform apply tfplan

# Get cluster credentials
$(terraform output -raw kubeconfig_command)
```

### 3. Install ArgoCD

```bash
# Create namespace
kubectl create namespace argocd

# Install ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### 4. Configure ArgoCD

```bash
# Apply project configuration
kubectl apply -f infra/argocd/projects/virtengine-project.yaml

# Apply app-of-apps
kubectl apply -f infra/argocd/apps/app-of-apps.yaml
```

### 5. Render and Apply Canonical Kubernetes

`deploy/kubernetes/overlays/prod` is the canonical production application
topology. The `infra/kubernetes/overlays/prod` entry point is retained only as a
compatibility shim for older GitOps wiring and must render byte-equivalent
output to the canonical deploy overlay.

Do not apply `infra/rollouts/` as a directory. That directory is not an
application topology source. `infra/rollouts/rollback-config.yaml` may be
applied only by operators who intentionally run Argo Rollouts AnalysisTemplate
policy for out-of-band health analysis; it does not replace the StatefulSet
topology in the canonical production overlay. Never use a recursive or
directory-wide apply for `infra/rollouts`.

```bash
# Render canonical production manifests
kubectl kustomize --load-restrictor=LoadRestrictionsNone \
  deploy/kubernetes/overlays/prod > /tmp/virtengine-prod.yaml

# Optional compatibility check for legacy infra entry points
kubectl kustomize --load-restrictor=LoadRestrictionsNone \
  infra/kubernetes/overlays/prod > /tmp/virtengine-prod-infra-shim.yaml
diff -u /tmp/virtengine-prod.yaml /tmp/virtengine-prod-infra-shim.yaml

# Review and apply only the canonical Kustomize entry point
kubectl diff -k deploy/kubernetes/overlays/prod
kubectl apply -k deploy/kubernetes/overlays/prod
```

### 6. Configure External Secrets

```bash
# Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets \
  --create-namespace \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=$(terraform output -raw external_secrets_role_arn)

# Verify the canonical production render contains the expected ExternalSecret resources
kubectl get externalsecrets -n virtengine
```

External Secrets resources are owned by the canonical deploy overlay and
environment-specific patches. Do not use the removed legacy external-secrets
file from the infra compatibility tree.

## Environment Promotion

### Dev → Staging

1. Create PR to merge `dev` branch into `staging`
2. CI runs tests and builds images tagged `staging`
3. ArgoCD auto-syncs staging applications
4. Run integration tests

### Staging → Production

1. Create release tag (e.g., `v1.0.0`)
2. Build production images and publish immutable SHA-256 digests
3. Create PR to promote only digest-pinned image references in the canonical
   deploy tree
4. Manual approval required
5. Render `deploy/kubernetes/overlays/prod` and verify the infra compatibility
   shim still renders identically
6. ArgoCD or the release operator applies
  `deploy/kubernetes/overlays/prod` as the canonical Kustomize entry point
7. Observe StatefulSet rolling updates and roll back failed revisions

## Production Rolling Update Process

### Digest Promotion

Production image promotion is a manifest change, not a mutable tag retarget:

1. Build and sign the release image.
2. Resolve the registry digest.
3. Replace the matching `image:` value with `name@sha256:<digest>` in the
   canonical deploy tree.
4. Run `node scripts/task85c-validate-kubernetes.mjs`.
   CI runs the same validator with checksum-pinned kubectl before downstream
   infrastructure security, plan, apply, or drift jobs.
5. After approval, run
  `kubectl apply -k deploy/kubernetes/overlays/prod` (or configure ArgoCD with
  that exact Kustomize path).

### Observe StatefulSets

Production chain and provider workloads are StatefulSets:

```bash
# Validator: exactly one remote-signer validator replica
kubectl rollout status statefulset/virtengine-validator -n virtengine --timeout=15m
kubectl get pods -n virtengine -l app=virtengine-validator -o wide

# Sentry/full nodes
kubectl rollout status statefulset/virtengine-node -n virtengine --timeout=30m
kubectl get pods -n virtengine -l app=virtengine-node -o wide

# Provider daemon HA replicas
kubectl rollout status statefulset/provider-daemon -n virtengine --timeout=30m
kubectl get pods -n virtengine -l app=provider-daemon -o wide
```

### Rollback

```bash
# Undo the previous StatefulSet controller revision
kubectl rollout undo statefulset/virtengine-node -n virtengine
kubectl rollout undo statefulset/provider-daemon -n virtengine

# Manual rollback to previous version
kubectl rollout status statefulset/virtengine-node -n virtengine --timeout=30m
kubectl rollout status statefulset/provider-daemon -n virtengine --timeout=30m

# Rollback to specific revision
kubectl rollout undo statefulset/virtengine-node --to-revision=2 -n virtengine
```

For `virtengine-validator`, roll back only after confirming the remote signer,
signer epoch, fencing token, and validator key fingerprint are consistent with
the target revision. The Kubernetes manifest rollback does not roll back the
independent TMKMS/HSM custody boundary.

## Monitoring Deployments

### ArgoCD UI

Access via port-forward:
```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Open https://localhost:8080
```

### Prometheus Metrics

Key deployment metrics:
- `argocd_app_sync_status`
- `kube_statefulset_status_replicas_ready`
- `kube_statefulset_status_observed_generation`
- `kube_pod_container_status_restarts_total`
- `kube_pod_status_ready`

## Troubleshooting

### ArgoCD Sync Issues

```bash
# Check application status
kubectl get applications -n argocd

# View sync details
argocd app get virtengine-node-prod

# Force refresh
argocd app refresh virtengine-node-prod

# Hard refresh (re-clone repo)
argocd app refresh virtengine-node-prod --hard
```

### StatefulSet Rolling Update Stuck

```bash
# Get StatefulSet details
kubectl describe statefulset virtengine-node -n virtengine
kubectl describe statefulset provider-daemon -n virtengine

# Check pods and events
kubectl get pods -n virtengine -o wide
kubectl get events -n virtengine --sort-by=.lastTimestamp

# Inspect current and previous container logs
kubectl logs -n virtengine statefulset/virtengine-node --tail=200
kubectl logs -n virtengine statefulset/provider-daemon --tail=200
```

### Terraform State Lock

```bash
# Force unlock (use carefully)
terraform force-unlock LOCK_ID
```

## Security Considerations

### Secrets Management

- All secrets stored in AWS Secrets Manager or Vault
- External Secrets Operator syncs to Kubernetes
- Secrets never committed to Git
- Rotation handled by Vault/AWS

### Network Security

- Private EKS endpoint for production
- Network policies restrict pod-to-pod traffic
- WAF protects public endpoints
- VPC flow logs enabled

### IAM Security

- IRSA for pod-level permissions
- Minimal IAM policies
- No long-lived credentials
- Audit logging enabled

## Disaster Recovery

### Backup Procedures

Chain state backups run every 4 hours:
```bash
./scripts/dr/backup-chain-state.sh
```

Key backups run daily:
```bash
./scripts/dr/backup-keys.sh --type all
```

### Recovery Procedures

See `scripts/dr/README.md` for detailed recovery procedures.

## Contact

- **On-Call:** #virtengine-oncall (Slack)
- **Infrastructure Team:** infra@virtengine.com
- **Security Issues:** security@virtengine.com
