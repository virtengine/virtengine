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
│   │   ├── iam/             # IAM roles and policies
│   │   ├── monitoring/      # CloudWatch and Prometheus monitoring
│   │   ├── scaling/         # Multi-region scaling infrastructure
│   │   ├── cost_optimization/  # Cost optimization automation (PERF-9B)
│   │   └── autoscaling_optimization/  # Auto-scaling policies (PERF-9B)
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

### Cost Optimization Features (PERF-9B)

The infrastructure includes comprehensive cost optimization automation:

#### Terraform Modules

- **cost_optimization**: AWS Budgets, Cost Anomaly Detection, unused resource cleanup Lambda, cost recommendation Lambda
- **autoscaling_optimization**: Tuned auto-scaling policies, scheduled scaling for non-prod, Cluster Autoscaler configuration

#### Provider Daemon Integration

The provider daemon includes cost tracking (`pkg/provider_daemon/cost_tracker.go`):
- Real-time workload cost tracking
- Resource usage to cost conversion
- Cost threshold alerts
- Right-sizing recommendations
- Cost anomaly detection

#### Documentation

- [Monthly Cost Review Runbook](./docs/MONTHLY_COST_REVIEW_RUNBOOK.md)
- [Reserved Instances & Savings Plans Guide](./docs/RESERVED_INSTANCES_SAVINGS_PLANS.md)
- [Cost Optimization Analysis](./docs/COST_OPTIMIZATION.md)

#### Key Features

| Feature | Description | Savings |
|---------|-------------|---------|
| Scheduled Scaling | Dev/staging scale-down off-hours | ~$550/month |
| Spot Instances | Workload nodes use Spot | ~$2,000/month |
| Resource Cleanup | Weekly cleanup of unused resources | ~$200/month |
| Cost Anomaly Detection | Alert on unusual spending | Early detection |
| Budget Alerts | Multi-tier threshold alerts | Budget compliance |
| Right-sizing | Automated recommendations | ~$1,500/month |
