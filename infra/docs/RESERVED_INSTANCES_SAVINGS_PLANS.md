# VirtEngine Reserved Instances and Savings Plans Guide

## Overview

This document provides guidance on purchasing and managing Reserved Instances (RIs) and Savings Plans (SPs) for VirtEngine infrastructure. Proper use of these commitment-based pricing models can reduce compute costs by 30-72%.

## Current Recommendations

Based on the VirtEngine infrastructure analysis:

### Production Environment

| Resource Type | Recommended Commitment | Discount | Annual Savings |
|---------------|------------------------|----------|----------------|
| 3x Validator Nodes (m5.2xlarge) | 1-year Standard RI | 40% | $11,520 |
| 2x System Nodes (t3.large) | Compute Savings Plan | 30% | $4,320 |
| RDS Primary (db.r5.large) | 1-year Reserved DB | 35% | $2,100 |

**Recommended Monthly Commitment**: $4,000-5,000 (1-year Compute Savings Plan)

### When to Use Each Option

#### Reserved Instances (RIs)
Best for:
- Specific instance types that won't change
- Database instances (RDS)
- Validators and chain nodes with known sizes

Pros:
- Highest discount (up to 72% for 3-year)
- Capacity reservation option
- Predictable pricing

Cons:
- Less flexible
- Tied to specific instance family
- Unused capacity is wasted

#### Compute Savings Plans
Best for:
- Workloads that may change instance types
- EKS node groups with variable sizing
- Flexibility is important

Pros:
- Applies across instance families
- Works with EC2, Fargate, Lambda
- Automatic application to usage

Cons:
- Lower discount than RIs (up to 66%)
- No capacity reservation

#### EC2 Instance Savings Plans
Best for:
- Committed to specific instance family
- Want higher discount than Compute SP
- Regional flexibility

Pros:
- Higher discount than Compute SP
- Flexible within instance family
- Works across sizes and tenancy

## Purchase Process

### Step 1: Analyze Usage (2-4 weeks before)

```bash
# Get cost and usage for last 3 months
aws ce get-cost-and-usage \
  --time-period Start=YYYY-MM-DD,End=YYYY-MM-DD \
  --granularity MONTHLY \
  --metrics "BlendedCost" "UnblendedCost" "UsageQuantity" \
  --group-by Type=DIMENSION,Key=INSTANCE_TYPE \
  --filter '{"Tags":{"Key":"Project","Values":["virtengine"]}}'

# Get reservation recommendations
aws ce get-reservation-purchase-recommendation \
  --service "Amazon Elastic Compute Cloud - Compute" \
  --account-scope PAYER \
  --lookback-period-in-days SIXTY_DAYS \
  --term-in-years ONE_YEAR \
  --payment-option NO_UPFRONT
```

### Step 2: Review Recommendations

Access AWS Cost Explorer:
1. Navigate to: **Cost Management → Reservations → Recommendations**
2. Review recommended purchases
3. Note estimated savings and break-even point

### Step 3: Approval Process

Before purchasing:
1. [ ] Review with Finance team
2. [ ] Get budget approval for commitment
3. [ ] Document business justification
4. [ ] Confirm workload will persist for commitment term

### Step 4: Purchase

#### For Reserved Instances:

```bash
# List available offerings
aws ec2 describe-reserved-instances-offerings \
  --instance-type m5.2xlarge \
  --product-description "Linux/UNIX" \
  --instance-tenancy default \
  --offering-type "No Upfront" \
  --duration 31536000  # 1 year

# Purchase RI
aws ec2 purchase-reserved-instances-offering \
  --reserved-instances-offering-id [OFFERING_ID] \
  --instance-count 3
```

#### For Savings Plans:

```bash
# Get savings plan recommendations
aws ce get-savings-plans-purchase-recommendation \
  --savings-plans-type COMPUTE_SP \
  --lookback-period-in-days SIXTY_DAYS \
  --term-in-years ONE_YEAR \
  --payment-option NO_UPFRONT
```

Then purchase via AWS Console:
1. Navigate to: **Cost Management → Savings Plans → Purchase Savings Plans**
2. Select commitment amount
3. Review and confirm

## Management and Monitoring

### Monthly Monitoring Tasks

1. **Check Utilization**
   ```bash
   aws ce get-reservation-utilization \
     --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
     --granularity MONTHLY
   
   aws ce get-savings-plans-utilization \
     --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
     --granularity MONTHLY
   ```

2. **Review Coverage**
   ```bash
   aws ce get-reservation-coverage \
     --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
     --granularity MONTHLY
   
   aws ce get-savings-plans-coverage \
     --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
     --granularity MONTHLY
   ```

### Utilization Targets

| Metric | Target | Action if Below |
|--------|--------|-----------------|
| RI Utilization | > 95% | Sell/modify RIs |
| SP Utilization | > 95% | Right-size workloads |
| Coverage | > 70% | Consider additional purchases |

### Expiration Management

Track expiring commitments:

```bash
# List expiring RIs
aws ec2 describe-reserved-instances \
  --filters "Name=state,Values=active" \
  --query 'ReservedInstances[?End<`2024-06-01`]'
```

Set up CloudWatch Alarms:
- Alert 90 days before expiration
- Alert 30 days before expiration
- Review and renew or let expire based on usage

## Selling or Modifying RIs

### Selling on RI Marketplace

If RIs are no longer needed:

1. Navigate to: **EC2 → Reserved Instances → Sell Reserved Instances**
2. List RIs with competitive pricing
3. Note: Marketplace fees apply (12% of upfront cost)

### Modifying RIs

You can modify RIs within the same instance family:

```bash
aws ec2 modify-reserved-instances \
  --reserved-instances-ids ri-1234567890abcdef0 \
  --target-configurations AvailabilityZone=us-east-1b,InstanceCount=2,InstanceType=m5.xlarge
```

## Cost Calculation Examples

### Example 1: Validator Nodes

Current: 3x m5.2xlarge On-Demand
- Hourly: $0.384 × 3 = $1.152
- Monthly: $1.152 × 730 = $840.96
- Annually: $10,091.52

With 1-year Standard RI (No Upfront):
- Hourly: $0.230 × 3 = $0.69
- Monthly: $503.70
- Annually: $6,044.40
- **Savings: $4,047/year (40%)**

### Example 2: Compute Savings Plan

$3,000/month commitment:
- Covers ~$4,500/month On-Demand equivalent
- Discount: ~33%
- Annual savings: ~$18,000

## Best Practices

1. **Start Conservative**
   - Begin with 60-70% of baseline usage
   - Increase coverage as workloads stabilize

2. **Match Term to Confidence**
   - 1-year: Standard workloads
   - 3-year: Core infrastructure unlikely to change

3. **Prefer Flexibility**
   - Savings Plans over RIs for variable workloads
   - Regional over zonal RIs

4. **Regular Reviews**
   - Monthly utilization checks
   - Quarterly optimization reviews
   - Annual renewal planning

5. **Document Everything**
   - Purchase justification
   - Expected vs actual savings
   - Lessons learned

## Appendix

### RI Payment Options

| Option | Upfront | Discount | Best For |
|--------|---------|----------|----------|
| All Upfront | 100% | Highest | Large capital budget |
| Partial Upfront | 50% | Medium | Balanced approach |
| No Upfront | 0% | Lowest | Cash flow optimization |

### Savings Plan Types

| Type | Coverage | Flexibility |
|------|----------|-------------|
| Compute SP | EC2, Fargate, Lambda | Highest |
| EC2 Instance SP | EC2 only | Medium |

### Contacts

- Cloud Economics: cloud-economics@virtengine.io
- Finance: finance@virtengine.io
- Platform Team: platform-team@virtengine.io
