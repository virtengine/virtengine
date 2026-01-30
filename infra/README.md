# VirtEngine Infrastructure as Code

Production-grade infrastructure automation using Terraform, ArgoCD, and GitOps principles.

## Directory Structure

```
infra/
├── terraform/                 # Terraform IaC modules and environments
│   ├── modules/              # Reusable Terraform modules
│   │   ├── vpc/             # VPC and networking
│   │   ├── eks/             # EKS Kubernetes cluster
│   │   ├── s3/              # S3 storage buckets
│   │   └── iam/             # IAM roles and policies
│   ├── environments/         # Environment-specific configurations
│   │   ├── dev/             # Development environment
│   │   ├── staging/         # Staging environment
│   │   └── prod/            # Production environment
│   └── tests/               # Terratest infrastructure tests
├── argocd/                   # ArgoCD GitOps configurations
│   ├── apps/                # Application manifests
│   ├── projects/            # ArgoCD projects
│   └── applicationsets/     # ApplicationSet patterns
├── kubernetes/               # Kubernetes manifests
│   ├── base/                # Base Kustomize configurations
│   └── overlays/            # Environment overlays
│       ├── dev/
│       ├── staging/
│       └── prod/
├── vault/                    # HashiCorp Vault configurations
│   └── policies/            # Vault policies
└── rollouts/                 # Argo Rollouts for blue/green deployments
```

## Quick Start

### Prerequisites

- Terraform >= 1.6.0
- AWS CLI configured with appropriate credentials
- kubectl configured for cluster access
- Helm >= 3.12.0

### Deploy Infrastructure

```bash
# Initialize and plan for development
cd infra/terraform/environments/dev
terraform init
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan
```

### Deploy Applications via GitOps

```bash
# Bootstrap ArgoCD
kubectl apply -k infra/argocd/

# Deploy app-of-apps
kubectl apply -f infra/argocd/apps/app-of-apps.yaml
```

### Run Infrastructure Tests

```bash
cd infra/terraform/tests
go test -v -timeout 30m
```

## Multi-Environment Strategy

| Environment | Purpose | Auto-deploy | Approval Required |
|-------------|---------|-------------|-------------------|
| dev | Development/testing | Yes (on PR merge) | No |
| staging | Pre-production validation | Yes (on release tag) | No |
| prod | Production workloads | No | Yes (manual) |

## Blue/Green Deployments

Blue/green deployments are managed via Argo Rollouts with:
- Automated traffic shifting
- Health check gates
- Automatic rollback on failure
- Manual promotion option for production

## Secrets Management

Secrets are managed via HashiCorp Vault with External Secrets Operator:
- Vault paths: `secret/virtengine/{env}/{service}`
- Auto-rotation for database credentials
- Integration with AWS Secrets Manager for cloud resources

## Cost Optimization

See [COST_OPTIMIZATION.md](./docs/COST_OPTIMIZATION.md) for detailed analysis.
