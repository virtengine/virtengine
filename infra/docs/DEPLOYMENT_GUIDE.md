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

### 5. Install Argo Rollouts

```bash
# Install Argo Rollouts
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml

# Apply rollout configurations
kubectl apply -f infra/rollouts/
```

### 6. Configure External Secrets

```bash
# Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets \
  --create-namespace \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=$(terraform output -raw external_secrets_role_arn)

# Apply secret store configurations
kubectl apply -f infra/kubernetes/base/external-secrets.yaml
```

## Environment Promotion

### Dev → Staging

1. Create PR to merge `dev` branch into `staging`
2. CI runs tests and builds images tagged `staging`
3. ArgoCD auto-syncs staging applications
4. Run integration tests

### Staging → Production

1. Create release tag (e.g., `v1.0.0`)
2. Build production images
3. Create PR to update production image tags
4. Manual approval required
5. ArgoCD syncs with manual promotion
6. Blue/green deployment activates

## Blue/Green Deployment Process

### Automatic Promotion

For services with `autoPromotionEnabled: true`:

1. New ReplicaSet deployed as preview
2. Pre-promotion analysis runs
3. If healthy for 5 minutes, auto-promotes
4. Old ReplicaSet scaled down after delay

### Manual Promotion

For production validators:

```bash
# Check rollout status
kubectl argo rollouts get rollout virtengine-node -n virtengine

# Promote manually after verification
kubectl argo rollouts promote virtengine-node -n virtengine

# Or abort if issues found
kubectl argo rollouts abort virtengine-node -n virtengine
```

### Rollback

```bash
# Automatic rollback (analysis failure)
# Rollouts automatically rolls back if analysis fails

# Manual rollback to previous version
kubectl argo rollouts undo virtengine-node -n virtengine

# Rollback to specific revision
kubectl argo rollouts undo virtengine-node --to-revision=2 -n virtengine
```

## Monitoring Deployments

### ArgoCD UI

Access via port-forward:
```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Open https://localhost:8080
```

### Argo Rollouts Dashboard

```bash
kubectl argo rollouts dashboard
# Open http://localhost:3100
```

### Prometheus Metrics

Key deployment metrics:
- `argocd_app_sync_status`
- `rollout_info`
- `rollout_phase`
- `analysis_run_metric_*`

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

### Rollout Stuck

```bash
# Get rollout details
kubectl argo rollouts get rollout virtengine-node -n virtengine -w

# Check analysis runs
kubectl get analysisruns -n virtengine

# View analysis logs
kubectl logs -l app.kubernetes.io/name=argo-rollouts -n argo-rollouts
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
- **Infrastructure Team:** infra@virtengine.io
- **Security Issues:** security@virtengine.io
