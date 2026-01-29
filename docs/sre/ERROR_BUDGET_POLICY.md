# Error Budget Policy

## Overview

This document defines the Error Budget Policy for VirtEngine services. Error budgets balance reliability with innovation velocity by quantifying acceptable unreliability and establishing clear decision-making criteria for operational changes.

## Philosophy

> "100% is the wrong reliability target for basically everything." - Google SRE Book

Perfect reliability is:
- **Expensive**: Exponential cost for diminishing returns
- **Unnecessary**: Users can't perceive 99.999% vs 100%
- **Innovation-Blocking**: Zero tolerance for risk prevents improvements

Error budgets solve this by:
1. Quantifying acceptable unreliability (100% - SLO)
2. Providing a shared currency between SRE and Engineering
3. Enabling data-driven decision making
4. Balancing stability with feature velocity

---

## Error Budget Calculation

### Formula

```
Error Budget = (1 - SLO_Target) × Total_Time_Period
```

### Examples

**99.90% SLO over 28 days**:
```
Error Budget = (1 - 0.9990) × 28 days × 24 hours × 60 minutes
            = 0.001 × 40,320 minutes
            = 40.32 minutes of allowed downtime
```

**99.50% SLO over 28 days**:
```
Error Budget = (1 - 0.9950) × 28 days × 24 hours × 60 minutes
            = 0.005 × 40,320 minutes
            = 201.6 minutes (3.36 hours) of allowed downtime
```

### Budget Period

- **Standard**: 28 days (4 weeks) rolling window
- **Rationale**:
  - Long enough to smooth out noise
  - Short enough to be actionable
  - Aligns with monthly planning cycles
- **Reset**: Automatic at end of period

---

## Budget Status Levels

Error budgets have four status levels based on remaining budget:

| Status | Remaining Budget | Operational Posture | Change Policy |
|--------|-----------------|---------------------|---------------|
| 🟢 **Healthy** | > 50% | Normal operations | All changes allowed |
| 🟡 **Warning** | 25% - 50% | Cautious | Feature releases require approval |
| 🔴 **Critical** | 5% - 25% | Defensive | Only bug fixes and stability work |
| ⚫ **Depleted** | < 5% | Emergency | Change freeze, emergency fixes only |

---

## Operational Policies by Status

### 🟢 Healthy (> 50% Remaining)

**Allowed Actions**:
- ✅ Feature releases
- ✅ Experimental features and A/B tests
- ✅ Performance optimizations
- ✅ Refactoring and technical debt work
- ✅ Infrastructure changes
- ✅ Aggressive deployment cadence (multiple times per day)

**Deployment Approval**: Engineering lead approval only

**Monitoring**: Weekly error budget review in team sync

**Rationale**: Sufficient budget to absorb potential issues. Encourage innovation and velocity.

---

### 🟡 Warning (25% - 50% Remaining)

**Allowed Actions**:
- ⚠️ Feature releases (SRE approval required)
- 🚫 Experimental features (postponed until budget recovers)
- ⚠️ Infrastructure changes (minimize risk)
- ✅ Bug fixes
- ✅ Stability improvements
- ⚠️ Reduced deployment cadence (1-2x per day)

**Deployment Approval**: Engineering lead + SRE approval required

**Monitoring**: Daily error budget review in SRE standup

**Required Actions**:
1. Review recent incidents for patterns
2. Identify toil automation opportunities
3. Update runbooks for common issues
4. Increase monitoring and alerting sensitivity

**Rationale**: Budget running low. Prioritize stability over new features until budget recovers.

---

### 🔴 Critical (5% - 25% Remaining)

**Allowed Actions**:
- 🚫 Feature releases (frozen)
- 🚫 Experimental features
- 🚫 Non-critical infrastructure changes
- ✅ Critical bug fixes
- ✅ Stability improvements
- ✅ Performance improvements
- ⚠️ Emergency changes only (VP approval)

**Deployment Approval**: VP Engineering approval required for all changes

**Monitoring**: Real-time error budget monitoring, hourly reviews

**Required Actions**:
1. **Immediate**: Root cause analysis of budget consumption
2. **Within 24h**: Incident review meeting with leadership
3. **Within 48h**: Remediation plan with timeline
4. **Ongoing**: Daily progress updates to leadership

**Escalation**: VP Engineering and CTO notified immediately

**Rationale**: Very limited budget remaining. All effort focused on stabilization.

---

### ⚫ Depleted (< 5% Remaining)

**Allowed Actions**:
- 🚫 All changes frozen (complete lockdown)
- ⚠️ Emergency security patches (executive approval)
- ⚠️ Emergency stability fixes (executive approval)
- ✅ Incident response and remediation

**Deployment Approval**: CTO or CEO approval required for ANY change

**Monitoring**: Continuous real-time monitoring, war room established

**Required Actions**:
1. **Immediate**: Declare Severity 1 incident
2. **Immediate**: Establish incident war room
3. **Within 1h**: Executive briefing (CTO, CEO)
4. **Within 4h**: External communication plan (if user-impacting)
5. **Within 24h**: Comprehensive post-incident review
6. **Within 48h**: Recovery plan with executive approval

**Escalation**: Executive team (CTO, CEO) on-call immediately

**Communication**:
- Internal: All engineering teams notified
- External: Customer communication if user-impacting
- Status page: Updated with current status

**Rationale**: SLO violated. Complete focus on restoration of service reliability.

---

## Burn Rate Alerts

Error budgets can be consumed slowly (chronic issues) or quickly (acute incidents). Burn rate monitoring detects both.

### Burn Rate Definition

```
Burn Rate = (Budget Consumed per Hour) / (Sustainable Rate per Hour)

Sustainable Rate = 1.0x
Fast Burn = > 2.0x
Critical Burn = > 5.0x
Emergency Burn = > 10.0x
```

### Alert Thresholds

| Burn Rate | Duration | Severity | Action | Notification |
|-----------|----------|----------|--------|--------------|
| 2x | 1 hour | Warning | Investigate | Slack alert |
| 5x | 15 minutes | Critical | Page on-call SRE | PagerDuty |
| 10x | 5 minutes | Emergency | Page incident commander | PagerDuty + Phone |

### Multi-Window Alerting

To reduce false positives, use multiple time windows:

**Example Alert Rule**:
```
Alert if:
  (Burn rate > 5x over last 1 hour) AND (Burn rate > 2x over last 6 hours)

Rationale:
  - Short window (1h) detects acute problems
  - Long window (6h) confirms it's not a transient spike
```

**Recommended Windows**:
- **Tier 0 (Critical)**: 1h + 6h
- **Tier 1 (High)**: 2h + 12h
- **Tier 2 (Standard)**: 4h + 24h

---

## Decision Making Framework

### Should We Deploy This Change?

**Decision Tree**:

```
1. What is the current error budget status?

   Healthy (> 50%) → Proceed to #2
   Warning (25-50%) → Proceed to #3
   Critical (5-25%) → Proceed to #4
   Depleted (< 5%) → STOP, see Depleted policy

2. Healthy - What type of change?

   Feature release → Approved (eng lead sign-off)
   Experimental → Approved (eng lead sign-off)
   Bug fix → Approved (eng lead sign-off)
   Infrastructure → Approved (eng lead sign-off)

3. Warning - What type of change?

   Feature release → Requires SRE approval + risk assessment
   Experimental → BLOCKED, postpone
   Bug fix → Approved (eng lead sign-off)
   Infrastructure → Requires SRE approval + rollback plan

4. Critical - What type of change?

   Feature release → BLOCKED
   Experimental → BLOCKED
   Bug fix (critical) → Requires VP approval
   Stability improvement → Approved (SRE sign-off)
```

### Risk Assessment Checklist

For changes requiring approval in Warning/Critical status:

- [ ] Clear rollback plan documented
- [ ] Canary deployment strategy defined
- [ ] Monitoring dashboards ready
- [ ] Alert rules configured
- [ ] On-call SRE briefed and available
- [ ] Change window scheduled (off-peak hours)
- [ ] Stakeholders notified

---

## Budget Consumption Sources

Error budgets can be consumed by:

### 1. Incidents (Downtime)

**Calculation**:
```
Budget Consumed = Downtime Duration (minutes)
```

**Example**:
- Service down for 15 minutes
- 15 minutes consumed from 40.32-minute budget
- 37% of budget consumed in one incident

**Tracking**: Automatic via uptime monitoring

### 2. Failed Requests (Error Rate)

**Calculation**:
```
Budget Consumed = (Failed Requests / Total Requests) × Time Period
```

**Example**:
- 1000 requests in 1 hour
- 10 requests failed (1% error rate)
- SLO is 99.90% (0.1% error budget)
- 1% actual vs 0.1% budget = 10x over budget for that hour
- Consumes 10 minutes of budget

**Tracking**: Automatic via request success rate metrics

### 3. Latency SLO Violations

**Calculation**:
```
Budget Consumed = (Slow Requests / Total Requests) × Time Period
```

**Example**:
- SLO: P95 latency < 2 seconds
- 100 requests in 10 minutes
- 10 requests exceeded 2 seconds (10%)
- 10% of 10 minutes = 1 minute consumed

**Tracking**: Automatic via latency percentile metrics

### 4. Manual Adjustments

**Use Cases**:
- Planned maintenance (consumes budget)
- Excluded incidents (refunds budget)
- Testing/load testing (may be excluded)

**Process**: SRE team approval required for manual adjustments

---

## Budget Exhaustion Scenarios

### Scenario 1: Slow Burn (Chronic Issues)

**Symptoms**:
- Budget consumption at 1.5x - 2x sustainable rate
- Gradual decline over days/weeks
- No single large incident

**Root Causes**:
- Intermittent errors
- Gradual performance degradation
- Infrastructure capacity issues
- Toil accumulation

**Response**:
1. Analyze error patterns and trends
2. Identify top contributors to budget consumption
3. Prioritize fixes for high-impact issues
4. Automate toil-heavy operations
5. Increase monitoring granularity

**Prevention**:
- Weekly error budget reviews
- Proactive performance monitoring
- Capacity planning
- Technical debt sprints

---

### Scenario 2: Fast Burn (Acute Incident)

**Symptoms**:
- Budget consumption at 5x+ sustainable rate
- Rapid decline over hours
- Clear incident trigger

**Root Causes**:
- Bad deployment
- Infrastructure failure
- External dependency outage
- Traffic spike (DDoS, viral load)

**Response**:
1. Declare incident (follow incident response process)
2. Immediate rollback if deployment-related
3. Engage on-call team
4. Establish incident command
5. Continuous status updates

**Prevention**:
- Canary deployments
- Automated rollbacks
- Rate limiting
- Circuit breakers
- Chaos engineering

---

### Scenario 3: Budget Reset Without Recovery

**Symptoms**:
- Budget resets at end of period
- Service still unstable
- Immediately starts consuming new budget

**Root Causes**:
- Structural reliability issues not addressed
- Insufficient post-incident follow-through
- Technical debt accumulation
- Inadequate capacity

**Response**:
1. Emergency architecture review
2. Reliability-focused sprint (1-2 weeks)
3. External expert consultation if needed
4. Potential service redesign
5. Increase SRE staffing on service

**Prevention**:
- Comprehensive post-incident reviews
- Action item tracking and enforcement
- Regular architecture reviews
- Proactive capacity planning

---

## Cross-Team Collaboration

### Engineering Team Responsibilities

1. **Design for Reliability**:
   - Build with error budgets in mind
   - Implement graceful degradation
   - Add comprehensive monitoring

2. **Respect Budget Status**:
   - Follow change policies for current status
   - Prioritize stability when budget is low
   - Participate in post-incident reviews

3. **Proactive Communication**:
   - Notify SRE of high-risk changes
   - Share deployment plans
   - Report anomalies early

### SRE Team Responsibilities

1. **Monitor and Report**:
   - Track error budgets in real-time
   - Weekly reports to engineering
   - Monthly executive summaries

2. **Enforce Policy**:
   - Approve/reject changes per policy
   - Escalate violations
   - Conduct post-incident reviews

3. **Enable Reliability**:
   - Provide tooling and automation
   - Conduct chaos engineering
   - Train teams on SRE practices

### Product Team Responsibilities

1. **Understand Trade-offs**:
   - Balance features vs reliability
   - Accept deployment delays when budget is low
   - Prioritize reliability work

2. **Support SRE Initiatives**:
   - Allocate time for technical debt
   - Include reliability in roadmaps
   - Celebrate reliability improvements

---

## Reporting and Communication

### Weekly Error Budget Report

**Audience**: Engineering teams

**Contents**:
- Current budget status per service
- Week-over-week trend
- Top budget consumers
- Upcoming changes risk assessment

**Format**: Slack post + dashboard link

---

### Monthly SLO Review

**Audience**: Engineering leadership

**Contents**:
- Monthly SLO achievement rates
- Error budget consumption breakdown
- Incidents and impact
- Improvement initiatives
- SLO adjustments (if needed)

**Format**: 30-minute meeting + written report

---

### Quarterly Business Review

**Audience**: Executive team, board

**Contents**:
- Quarterly reliability trends
- Customer impact metrics
- SRE team achievements
- Investment needs
- Strategic initiatives

**Format**: Executive presentation

---

## Policy Exceptions

### When to Grant Exceptions

Exceptions to error budget policy may be granted for:

1. **Security Patches**: Critical security vulnerabilities
2. **Regulatory Compliance**: Legal/compliance requirements
3. **Business Critical**: Executive-level priority (M&A, major launch)
4. **Customer Commitments**: Contractual obligations

### Exception Process

1. **Request**: Submit exception request with justification
2. **Risk Assessment**: SRE team evaluates impact
3. **Approval**: Requires approval from:
   - Warning status: VP Engineering
   - Critical status: CTO
   - Depleted status: CEO
4. **Documentation**: Exception logged with rationale
5. **Review**: Post-exception review within 48 hours

### Exception Template

```markdown
## Error Budget Policy Exception Request

**Requestor**: [Name, Team]
**Date**: [YYYY-MM-DD]
**Service**: [Service Name]
**Current Budget Status**: [Healthy/Warning/Critical/Depleted]

### Change Description
[What change is being requested]

### Justification
[Why this exception is needed]
- [ ] Security patch
- [ ] Regulatory compliance
- [ ] Business critical
- [ ] Customer commitment
- [ ] Other: ___________

### Risk Assessment
- **Probability of failure**: [Low/Medium/High]
- **Impact if failure**: [Low/Medium/High]
- **Rollback time**: [X minutes]
- **Customer impact**: [Yes/No, details]

### Mitigation Plan
- Rollback procedure: [...]
- Monitoring: [...]
- On-call coverage: [...]

### Approval
- [ ] SRE Lead: ____________ (Date: ______)
- [ ] VP Engineering: ____________ (Date: ______)
- [ ] CTO: ____________ (Date: ______)
```

---

## Metrics and KPIs

### Error Budget Health

Track across all services:

```
Average Budget Remaining = Σ(Budget Remaining) / Number of Services

Target: > 60% (organization-wide)
```

### Budget Exhaustion Rate

```
Budget Exhaustion Rate = Services with < 25% Budget / Total Services

Target: < 10%
```

### Policy Compliance

```
Policy Compliance = Approved Changes / Total Changes

Target: > 95%
```

### Mean Time to Recovery (MTTR)

```
MTTR = Σ(Incident Duration) / Number of Incidents

Target: < 30 minutes
```

---

## Continuous Improvement

### Retrospective Questions

After budget exhaustion or critical status:

1. **What consumed the budget?**
   - Incidents, errors, latency violations?
   - Root causes identified?

2. **Could we have prevented it?**
   - Monitoring gaps?
   - Testing gaps?
   - Process failures?

3. **How can we prevent recurrence?**
   - Technical changes?
   - Process improvements?
   - Training needs?

4. **Are our SLOs appropriate?**
   - Too aggressive (always depleting)?
   - Too lenient (never consuming)?
   - Aligned with user expectations?

### SLO Adjustment Criteria

Adjust SLOs if:

- **Consistently exceeding** (>99.99% for 3+ months)
  - Consider tightening SLO (more ambitious)
  - Frees up engineering time for features

- **Consistently missing** (<90% of months meet SLO)
  - Consider loosening SLO (more realistic)
  - Or invest in reliability improvements

- **Customer feedback misalignment**
  - Users complain despite meeting SLO → tighten
  - Users satisfied despite missing SLO → loosen

---

## References

- [SLI/SLO/SLA Definitions](SLI_SLO_SLA.md)
- [Incident Response Process](INCIDENT_RESPONSE.md)
- [Alerting Rules](ALERTING_RULES.md)
- [Post-Incident Review Template](templates/postmortem_template.md)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
