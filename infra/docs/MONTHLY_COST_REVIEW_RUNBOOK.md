# VirtEngine Monthly Cost Review Runbook

## Overview

This runbook provides a structured process for conducting monthly infrastructure cost reviews. The goal is to identify optimization opportunities, track spending trends, and ensure efficient resource utilization.

## Schedule

- **Frequency**: First business day of each month
- **Duration**: 2-3 hours
- **Participants**: Platform Team, Finance (optional), Engineering Leads (as needed)

## Pre-Review Checklist

Before the review meeting:

- [ ] Cost recommendations Lambda has run (automatically on 1st of month)
- [ ] AWS Cost Explorer data is up to date (24-48 hour delay)
- [ ] Grafana cost dashboard is accessible
- [ ] Previous month's action items have been tracked

## Review Process

### 1. Cost Summary Review (30 min)

#### AWS Console Steps

1. **Access Cost Explorer**
   ```
   AWS Console → Billing & Cost Management → Cost Explorer
   ```

2. **Review Monthly Spend**
   - Filter by: Tags → Project = virtengine
   - Group by: Service
   - Compare: Current month vs Previous month

3. **Check Budget Status**
   ```
   AWS Console → Billing → Budgets
   ```
   - Review alerts triggered
   - Check forecast vs budget

4. **Review Anomaly Detection**
   ```
   AWS Console → Billing → Cost Anomaly Detection
   ```
   - Investigate any flagged anomalies
   - Document root causes

#### Key Metrics to Record

| Metric | Previous Month | Current Month | Change % | Notes |
|--------|---------------|---------------|----------|-------|
| Total Spend | | | | |
| EC2/EKS | | | | |
| RDS | | | | |
| Data Transfer | | | | |
| S3/EBS | | | | |
| NAT Gateway | | | | |

### 2. Service-Level Analysis (30 min)

#### Compute Costs (EC2/EKS)

```bash
# Check EC2 right-sizing recommendations
aws ce get-rightsizing-recommendation \
  --service AmazonEC2 \
  --configuration RecommendationTarget=SAME_INSTANCE_FAMILY,BenefitsConsidered=true
```

Review:
- [ ] Instance utilization (target: 60-80% CPU average)
- [ ] Spot instance usage percentage
- [ ] Reserved Instance / Savings Plan utilization
- [ ] Cluster Autoscaler efficiency

#### Storage Costs

Review:
- [ ] EBS volume utilization (unattached volumes)
- [ ] S3 storage class distribution
- [ ] Snapshot retention compliance
- [ ] gp2 → gp3 migration status

#### Network Costs

Review:
- [ ] NAT Gateway data transfer
- [ ] Cross-AZ traffic
- [ ] VPC Endpoint utilization
- [ ] Data transfer out costs

### 3. Optimization Opportunities (30 min)

#### Reserved Instances / Savings Plans

```bash
# Check RI utilization
aws ce get-reservation-utilization \
  --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
  --granularity MONTHLY

# Check Savings Plans coverage
aws ce get-savings-plans-coverage \
  --time-period Start=YYYY-MM-01,End=YYYY-MM-DD \
  --granularity MONTHLY
```

Recommendations:
- Current RI utilization: ____%
- Unused RI hours: ____
- Recommended RI purchases: ____
- Recommended SP commitment: $____/month

#### Right-Sizing Recommendations

1. Access Cost Explorer recommendations
2. Export recommendations list
3. Prioritize by potential savings
4. Create tickets for top 5 opportunities

### 4. Environment-Specific Review (20 min)

#### Development Environment

- [ ] Scheduled scaling working (8 PM - 8 AM shutdown)
- [ ] Weekend shutdown enabled
- [ ] All resources using Spot instances
- [ ] Single NAT gateway configuration

Monthly dev cost: $____
Target: < $650/month

#### Staging Environment

- [ ] Reduced replicas during off-hours
- [ ] Appropriate instance sizing
- [ ] Multi-AZ only for critical services

Monthly staging cost: $____
Target: < $2,500/month

#### Production Environment

- [ ] RI/SP coverage adequate
- [ ] Cluster Autoscaler efficient
- [ ] No orphaned resources
- [ ] Storage lifecycle policies active

Monthly prod cost: $____
Target: < $10,500/month

### 5. Cleanup Verification (20 min)

Check automated cleanup results:

```bash
# Review cleanup Lambda logs
aws logs filter-log-events \
  --log-group-name /aws/lambda/virtengine-prod-resource-cleanup \
  --start-time $(date -d '30 days ago' +%s000) \
  --filter-pattern "Cleanup report"
```

Manual checks:
- [ ] Unattached EBS volumes
- [ ] Old EBS snapshots (> 30 days)
- [ ] Unused Elastic IPs
- [ ] Terminated but not deleted resources
- [ ] Old AMIs (> 90 days)

### 6. Action Items (30 min)

#### Template for Action Items

| Priority | Action | Owner | Due Date | Estimated Savings |
|----------|--------|-------|----------|-------------------|
| P1 | | | | |
| P2 | | | | |
| P3 | | | | |

#### Common Action Categories

1. **Immediate** (This week)
   - Delete unused resources
   - Fix misconfigured scaling
   - Address budget overruns

2. **Short-term** (This month)
   - Right-size instances
   - Enable scheduled scaling
   - Migrate storage classes

3. **Medium-term** (This quarter)
   - Purchase RIs/SPs
   - Implement Karpenter
   - Evaluate Graviton instances

## Cost Optimization Checklist

### Compute
- [ ] All workload nodes using Spot where appropriate
- [ ] CPU utilization 60-80% average
- [ ] Cluster Autoscaler configured with correct thresholds
- [ ] Predictive scaling enabled for predictable workloads

### Storage
- [ ] All EBS using gp3 (not gp2)
- [ ] S3 lifecycle policies configured
- [ ] No unattached volumes > 7 days
- [ ] Snapshot retention < 30 days

### Network
- [ ] VPC endpoints for S3, ECR, STS
- [ ] Single NAT in dev environment
- [ ] Data transfer patterns optimized
- [ ] CloudFront for cacheable content

### Reserved Capacity
- [ ] RI utilization > 95%
- [ ] SP coverage > 70%
- [ ] No unused reservations
- [ ] Upcoming expiration reviewed

## Escalation Criteria

Escalate to leadership if:
- Monthly spend exceeds budget by > 20%
- Anomaly detection finds unexplained cost spike > $1,000
- RI/SP utilization drops below 80%
- Optimization opportunities > $500/month not addressed within 30 days

## Reporting

### Monthly Report Template

```markdown
# VirtEngine Cost Report - [Month Year]

## Executive Summary
- Total spend: $X,XXX
- Budget variance: +/-X%
- YoY change: +/-X%

## Key Findings
1. [Finding 1]
2. [Finding 2]
3. [Finding 3]

## Optimization Actions Taken
1. [Action 1] - Saved $XXX/month
2. [Action 2] - Saved $XXX/month

## Recommendations
1. [Recommendation 1] - Est. savings: $XXX/month
2. [Recommendation 2] - Est. savings: $XXX/month

## Next Month Focus
1. [Focus area 1]
2. [Focus area 2]
```

## Tools and Resources

### Dashboards
- Grafana Cost Dashboard: `https://grafana.internal/d/cost-optimization`
- AWS Cost Explorer: `https://console.aws.amazon.com/cost-management/home`
- CloudWatch Cost Dashboard: `virtengine-{env}-cost-dashboard`

### Lambda Functions
- Resource Cleanup: `virtengine-{env}-resource-cleanup`
- Cost Recommendations: `virtengine-{env}-cost-recommendations`

### Documentation
- Cost Optimization Guide: `infra/docs/COST_OPTIMIZATION.md`
- Infrastructure Modules: `infra/terraform/modules/`

## Appendix

### AWS Cost Categories

Ensure these cost categories are configured:
- `virtengine-dev`
- `virtengine-staging`
- `virtengine-prod`
- `virtengine-shared`

### Required IAM Permissions

For cost review access:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:*",
        "budgets:View*",
        "aws-portal:View*"
      ],
      "Resource": "*"
    }
  ]
}
```

### Contact

- Platform Team: platform-team@virtengine.io
- Finance: finance@virtengine.io
- On-call: #platform-oncall
