# VirtEngine Production Infrastructure

This directory contains the Infrastructure as Code (IaC) for deploying VirtEngine blockchain infrastructure to AWS.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              VPC (10.x.0.0/16)                         │
├────────────────────────────┬────────────────────────────────────────────┤
│     Public Subnets         │           Private Subnets                  │
│   ┌─────────────────┐      │     ┌──────────────────────┐              │
│   │   NAT Gateway   │      │     │   EKS Node Groups    │              │
│   │   ALB/NLB       │◄─────┼────►│   - System nodes     │              │
│   │   Bastion       │      │     │   - App nodes        │              │
│   └─────────────────┘      │     │   - Chain nodes      │              │
│                            │     └──────────────────────┘              │
│                            │                                            │
│                            │     ┌──────────────────────┐              │
│                            │     │   Database Subnets   │              │
│                            │     │   - RDS PostgreSQL   │              │
│                            │     │   - Read Replica     │              │
│                            │     └──────────────────────┘              │
└────────────────────────────┴────────────────────────────────────────────┘
```

## Directory Structure

```
infra/
├── terraform/
│   ├── modules/              # Reusable Terraform modules
│   │   ├── networking/       # VPC, subnets, NAT, security groups
│   │   ├── eks/              # EKS cluster and node groups
│   │   ├── rds/              # RDS PostgreSQL
│   │   ├── vault/            # HashiCorp Vault on K8s
│   │   └── monitoring/       # CloudWatch, Prometheus, Grafana
│   ├── environments/         # Environment-specific configurations
│   │   ├── dev/             
│   │   ├── staging/         
│   │   └── prod/            
│   └── terragrunt.hcl        # DRY configuration management
└── tests/                    # Terratest infrastructure tests
```

## Quick Start

### Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform >= 1.5.0
- Terragrunt >= 0.50.0
- kubectl
- helm

### Deploy Development Environment

```bash
cd infra/terraform/environments/dev
terragrunt init
terragrunt plan
terragrunt apply
```

### Deploy Production Environment

```bash
cd infra/terraform/environments/prod
terragrunt init
terragrunt plan -out=tfplan
# Review the plan carefully
terragrunt apply tfplan
```

## Modules

### Networking

Creates VPC with public, private, and database subnets across multiple AZs.

| Resource | Description |
|----------|-------------|
| VPC | Main network with DNS enabled |
| Public Subnets | Internet-facing resources, load balancers |
| Private Subnets | EKS nodes, internal services |
| Database Subnets | Isolated subnets for RDS |
| NAT Gateway | Outbound internet access for private subnets |
| Security Groups | Network-level access control |

### EKS

Managed Kubernetes cluster with specialized node groups.

| Node Group | Purpose | Instance Types |
|------------|---------|----------------|
| System | Cluster addons, monitoring | t3.large |
| Application | VirtEngine services | m5.xlarge |
| Chain | Blockchain nodes | m5.2xlarge |

### RDS

PostgreSQL database for chain state backup and application data.

| Feature | Dev | Staging | Prod |
|---------|-----|---------|------|
| Instance | db.t3.medium | db.r5.large | db.r5.xlarge |
| Multi-AZ | No | Yes | Yes |
| Encryption | Yes | Yes | Yes |
| Read Replica | No | No | Yes |

### Vault

HashiCorp Vault for secrets management with AWS KMS auto-unseal.

Features:
- HA deployment with DynamoDB backend
- AWS KMS auto-unseal
- IRSA for AWS authentication
- External Secrets Operator integration

### Monitoring

Comprehensive monitoring with CloudWatch and Prometheus stack.

Components:
- CloudWatch dashboards and alarms
- Prometheus with persistent storage
- Grafana with pre-configured dashboards
- Alertmanager with SNS integration

## GitOps Workflow

Deployments are managed via ArgoCD with the following flow:

```
Git Push → ArgoCD Sync → Kubernetes Apply → Health Check
```

### Environments

| Environment | Branch | Auto-sync | Prune |
|-------------|--------|-----------|-------|
| Dev | main | Yes | Yes |
| Staging | release/staging | Yes | Yes |
| Prod | release/prod | No | No |

### Blue/Green Deployments

Production uses Istio-based blue/green deployments:

1. Deploy new version to "green" pods
2. Gradually shift traffic (10% increments)
3. Monitor error rates
4. Complete switch or rollback

```bash
# Switch traffic to green
./scripts/rollback/blue-green-switch.sh virtengine-node green

# Rollback to blue
./scripts/rollback/blue-green-switch.sh virtengine-node blue
```

## Rollback Procedures

### ArgoCD Rollback

```bash
./scripts/rollback/argocd-rollback.sh virtengine-prod
```

### Terraform State Rollback

```bash
./scripts/rollback/terraform-rollback.sh prod 1
```

## Testing

Run infrastructure tests:

```bash
cd infra/tests
go test -v -timeout 30m -run TestNetworkingModule
```

## Security

- All data encrypted at rest (KMS)
- TLS for all in-transit data
- Network policies for pod-to-pod communication
- IRSA for AWS service authentication
- Vault for secrets management
- VPC Flow Logs enabled in production

## Cost Optimization

| Environment | Monthly Estimate |
|-------------|------------------|
| Dev | ~$500 |
| Staging | ~$1,500 |
| Prod | ~$5,000+ |

Tips:
- Use SPOT instances for dev/staging app nodes
- Single NAT gateway for non-prod
- Reserved instances for production

## Troubleshooting

### EKS Access Issues

```bash
aws eks update-kubeconfig --region us-east-1 --name virtengine-<env>
```

### Terraform State Lock

```bash
aws dynamodb delete-item \
  --table-name virtengine-terraform-locks \
  --key '{"LockID": {"S": "<lock-id>"}}'
```

### ArgoCD Sync Issues

```bash
argocd app sync <app-name> --force --prune
```

## Support

For infrastructure issues, contact:
- On-call: oncall@virtengine.io
- Slack: #infra-support
