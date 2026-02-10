# VirtEngine Infrastructure Cost Optimization Analysis

## Executive Summary

This document provides cost optimization strategies for VirtEngine's AWS infrastructure across development, staging, and production environments. Implementing these recommendations can reduce infrastructure costs by 30-50% while maintaining performance and reliability.

## Current Cost Breakdown (Estimated Monthly)

| Environment | Compute | Storage | Networking | Total |
|-------------|---------|---------|------------|-------|
| Development | $500 | $50 | $100 | $650 |
| Staging | $2,000 | $200 | $300 | $2,500 |
| Production | $8,000 | $1,000 | $1,500 | $10,500 |
| **Total** | **$10,500** | **$1,250** | **$1,900** | **$13,650** |

## Cost Optimization Strategies

### 1. Compute Optimization

#### Spot Instances for Non-Critical Workloads

**Current State:** All node groups use On-Demand instances
**Recommendation:** Use Spot instances for workload and development node groups

```hcl
# Recommended node group configuration
node_groups = {
  system = {
    capacity_type = "ON_DEMAND"  # Keep system nodes on-demand
    # ...
  }
  workload = {
    capacity_type = "SPOT"  # Use spot for workloads
    instance_types = ["t3.xlarge", "t3a.xlarge", "m5.xlarge"]  # Multiple types for availability
    # ...
  }
}
```

**Savings:** 60-70% on workload compute costs
**Monthly Impact:** -$2,000

#### Right-Sizing Instances

| Node Group | Current | Recommended | Justification |
|------------|---------|-------------|---------------|
| Dev System | t3.medium | t3.small | Dev has low utilization |
| Staging Workload | t3.xlarge | t3.large | 40% CPU headroom observed |
| Prod Validators | m5.2xlarge | m5.xlarge | Memory usage under 50% |

**Monthly Impact:** -$1,500

#### Reserved Instances / Savings Plans

For production workloads with predictable usage:

| Component | Commitment | Discount | Annual Savings |
|-----------|------------|----------|----------------|
| 3x Validator nodes | 1-year RI | 40% | $11,520 |
| 2x System nodes | Compute SP | 30% | $4,320 |
| NAT Gateways | N/A | N/A | N/A |

**Recommendation:** Purchase 1-year Compute Savings Plan at $5,000/month commitment

**Annual Savings:** $18,000

### 2. Storage Optimization

#### EBS Volume Optimization

| Current | Recommended | Impact |
|---------|-------------|--------|
| gp2 100GB per node | gp3 100GB | -20% cost, +20% baseline IOPS |
| No lifecycle policies | 30-day snapshot retention | -$200/month |

```hcl
# Use gp3 for better price/performance
resource "aws_ebs_volume" "example" {
  type = "gp3"
  iops = 3000  # Baseline (no extra cost)
  throughput = 125  # Baseline (no extra cost)
}
```

#### S3 Lifecycle Policies

Implemented in the S3 module:

```hcl
lifecycle_configuration {
  rule {
    id = "transition-to-ia"
    transition {
      days = 30
      storage_class = "STANDARD_IA"  # 45% cheaper
    }
    transition {
      days = 90
      storage_class = "GLACIER"  # 80% cheaper
    }
  }
}
```

**Monthly Impact:** -$400

### 3. Network Optimization

#### NAT Gateway Consolidation

**Current:** 3 NAT gateways (one per AZ) in each environment
**Recommendation:** 
- Dev: 1 NAT gateway (non-HA acceptable)
- Staging: 2 NAT gateways
- Prod: 3 NAT gateways (maintain HA)

**Monthly Impact:** -$200

#### VPC Endpoints

Already implemented - saves ~$50/month in NAT gateway data transfer for AWS service traffic.

#### Data Transfer Optimization

- Use VPC endpoints for S3, ECR, STS (implemented)
- Enable S3 Transfer Acceleration for large uploads: Not needed currently
- Use CloudFront for API caching: Consider for public endpoints

### 4. Environment-Specific Optimizations

#### Development Environment

| Optimization | Implementation | Savings |
|-------------|----------------|---------|
| Shutdown off-hours | Scheduled scaling to 0 from 8PM-8AM | -$250/month |
| Spot-only nodes | All node groups use SPOT | -$200/month |
| Reduced redundancy | Single NAT, no multi-AZ | -$100/month |

**Dev Auto-Shutdown Script:**
```bash
# Scale down dev cluster at 8 PM
kubectl scale deployment --all --replicas=0 -n virtengine-dev

# Scale up at 8 AM
kubectl scale deployment virtengine-node --replicas=2 -n virtengine-dev
```

#### Staging Environment

| Optimization | Implementation | Savings |
|-------------|----------------|---------|
| Reduced replicas | 2 validators instead of 3 | -$400/month |
| Smaller instance types | t3.large instead of t3.xlarge | -$300/month |

#### Production Environment

| Optimization | Implementation | Savings |
|-------------|----------------|---------|
| Savings Plans | 1-year commitment | -$1,500/month |
| GP3 volumes | Already using | $0 |
| Cluster Autoscaler | Scale based on demand | -$500/month (average) |

### 5. Monitoring & Cost Visibility

#### AWS Cost Allocation Tags

All resources are tagged with:
```hcl
default_tags {
  tags = {
    Project     = "virtengine"
    Environment = var.environment
    ManagedBy   = "terraform"
    CostCenter  = "infrastructure"
  }
}
```

#### Budget Alerts

Set up AWS Budgets:
```
- Monthly budget: $15,000 (total)
- Alert at 80%: $12,000
- Alert at 100%: $15,000
- Forecast alert at 110%
```

#### Cost Explorer Recommendations

Review monthly:
- Right-sizing recommendations
- Idle resource detection
- Reserved Instance utilization

## Implementation Roadmap

### Phase 1: Quick Wins (Week 1-2)
- [x] Enable gp3 volumes in Terraform modules
- [ ] Configure S3 lifecycle policies
- [ ] Implement dev environment auto-shutdown
- [ ] Set up cost allocation tags

**Expected Savings:** $600/month

### Phase 2: Spot Instances (Week 3-4)
- [ ] Enable Spot for dev workload nodes
- [ ] Enable Spot for staging workload nodes
- [ ] Configure Spot instance interruption handling
- [ ] Set up Spot Fleet diversification

**Expected Savings:** $1,500/month

### Phase 3: Capacity Planning (Month 2)
- [ ] Analyze 30-day metrics for right-sizing
- [ ] Purchase Savings Plans for production
- [ ] Implement Cluster Autoscaler tuning

**Expected Savings:** $2,000/month

### Phase 4: Advanced Optimization (Month 3+)
- [ ] Implement Karpenter for better node provisioning
- [ ] Evaluate Graviton instances for workloads
- [ ] Consider EKS Anywhere for on-premises workloads

**Expected Savings:** $500/month

## Cost Monitoring Dashboard

### Key Metrics to Track

1. **Cost per transaction**
   ```promql
   aws_cost_daily / rate(virtengine_transactions_total[24h])
   ```

2. **Cost per validator**
   ```promql
   aws_ec2_cost{role="validator"} / count(virtengine_validators_active)
   ```

3. **Infrastructure efficiency**
   ```promql
   (cpu_utilization + memory_utilization) / 2
   ```

### Grafana Dashboard

Import the cost dashboard from: `deploy/monitoring/grafana/dashboards/cost-optimization.json`

## Summary of Expected Savings

| Category | Monthly Savings | Annual Savings |
|----------|-----------------|----------------|
| Spot Instances | $2,000 | $24,000 |
| Right-sizing | $1,500 | $18,000 |
| Savings Plans | $1,500 | $18,000 |
| Storage | $400 | $4,800 |
| Networking | $200 | $2,400 |
| Dev Shutdown | $350 | $4,200 |
| **Total** | **$5,950** | **$71,400** |

**Cost Reduction:** 43% reduction in infrastructure costs

## Appendix

### A. Instance Type Comparison

| Type | vCPU | Memory | Hourly (On-Demand) | Hourly (Spot Avg) |
|------|------|--------|-------------------|-------------------|
| t3.medium | 2 | 4 GB | $0.0416 | $0.0125 |
| t3.large | 2 | 8 GB | $0.0832 | $0.0250 |
| t3.xlarge | 4 | 16 GB | $0.1664 | $0.0499 |
| m5.large | 2 | 8 GB | $0.096 | $0.0384 |
| m5.xlarge | 4 | 16 GB | $0.192 | $0.0768 |
| m5.2xlarge | 8 | 32 GB | $0.384 | $0.1536 |

### B. Savings Plan Calculator

Use AWS Savings Plan Calculator: https://calculator.aws/#/addService/SavingsPlans

Recommended commitment based on baseline:
- Compute Savings Plan: $3,000/month (3-year) = 66% discount
- Or: $4,000/month (1-year) = 40% discount

### C. Spot Instance Best Practices

1. Use multiple instance types in node groups
2. Spread across all AZs
3. Implement graceful termination handling
4. Use Spot Instance Advisor for availability
5. Set appropriate `spotInterruptionBehavior`

### D. Related Documentation

- [AWS Cost Optimization Pillar](https://docs.aws.amazon.com/wellarchitected/latest/cost-optimization-pillar/)
- [EKS Best Practices - Cost](https://aws.github.io/aws-eks-best-practices/cost_optimization/)
- [Karpenter Documentation](https://karpenter.sh/)
