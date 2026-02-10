# Incident Tracking and Metrics

## Overview

This document defines the incident tracking system and key metrics for measuring and improving incident response effectiveness. What gets measured gets improved.

## Table of Contents

1. [Incident Tracking System](#incident-tracking-system)
2. [Key Metrics](#key-metrics)
3. [Dashboards](#dashboards)
4. [Reporting](#reporting)
5. [Analysis and Insights](#analysis-and-insights)
6. [Continuous Improvement](#continuous-improvement)

---

## Incident Tracking System

### Tracking Objectives

**Why Track Incidents**:
- Measure response effectiveness
- Identify patterns and trends
- Track action item completion
- Demonstrate reliability improvements
- Support capacity planning
- Guide priority decisions

**What to Track**:
- Incident timeline and duration
- Impact and affected users
- Root cause and resolution
- Response metrics (TTD, TTA, TTR)
- Action items and completion
- Cost and resource utilization

---

### Tracking Tools

**Primary Tools**:

| Tool | Purpose | Data Captured |
|------|---------|---------------|
| **PagerDuty** | Incident lifecycle | Alert time, acknowledgment, escalations |
| **Jira** | Incident tickets | Details, action items, postmortem links |
| **Slack** | Communication logs | Timeline, decisions, coordination |
| **Grafana** | Metrics visualization | System metrics during incident |
| **Postmortem Docs** | Learnings | Root cause, action items, lessons |

**Data Flow**:
```
Monitoring Alert → PagerDuty Incident → Jira Ticket → Postmortem Doc
                                    ↓
                              Slack Channel
                                    ↓
                           Metrics Dashboard
```

---

### Incident Ticket Structure

**Jira Ticket Template**:

```
Title: [YYYY-MM-DD] [SEV-X] [Service] - [Brief Description]

Example: [2026-01-30] [SEV-2] [API] - Connection Pool Exhaustion

Fields:
- Incident ID: INCIDENT-12345
- Severity: SEV-1 | SEV-2 | SEV-3 | SEV-4
- Service: [Service Name]
- Start Time: YYYY-MM-DD HH:MM:SS UTC
- End Time: YYYY-MM-DD HH:MM:SS UTC
- Duration: [X] minutes
- Impact: [User impact description]
- Root Cause: [Brief description]
- Resolution: [What fixed it]
- Postmortem: [Link to postmortem doc]
- Action Items: [Links to follow-up tickets]
- Status: Open | In Progress | Resolved | Closed

Labels:
- incident
- sev-[1-4]
- service:[service-name]
- root-cause:[category]
- detection:[method]
```

---

### Incident Taxonomy

**Categorization Dimensions**:

**By Service**:
- Blockchain Core
- VEID/Identity
- Provider Daemon
- API Gateway
- Database
- Network/Infrastructure
- Security
- ML/Inference

**By Root Cause**:
- Code Bug
- Configuration Error
- Capacity/Scaling
- External Dependency
- Human Error
- Hardware Failure
- Security Incident
- Unknown

**By Detection Method**:
- Automated Monitoring
- User Report
- Internal Discovery
- External Notification

**By Impact Type**:
- Complete Outage
- Degraded Performance
- Feature Unavailable
- Data Loss/Corruption
- Security Breach

---

## Key Metrics

### Response Time Metrics

#### 1. Time to Detect (TTD)

**Definition**: Time from incident start to alert firing

```
TTD = Alert Time - Incident Start Time
```

**Target**:
- Critical systems: < 5 minutes
- Important systems: < 15 minutes

**Measurement**:
- Incident Start: First error in logs or metrics anomaly
- Alert Time: PagerDuty alert timestamp

**Example**:
```
Incident Start: 14:23:00 UTC (first error in logs)
Alert Fired: 14:24:30 UTC
TTD = 1.5 minutes ✅ (under 5 min target)
```

**Improvements to Reduce TTD**:
- More sensitive alerting thresholds
- Predictive anomaly detection
- Synthetic monitoring
- User session monitoring

---

#### 2. Time to Acknowledge (TTA)

**Definition**: Time from alert to human acknowledgment

```
TTA = Acknowledgment Time - Alert Time
```

**Target**: < 5 minutes (all severities)

**Measurement**:
- Alert Time: PagerDuty alert timestamp
- Acknowledgment: Engineer clicks "Acknowledge" in PagerDuty

**Example**:
```
Alert Fired: 14:24:30 UTC
Acknowledged: 14:26:00 UTC
TTA = 1.5 minutes ✅
```

**Improvements to Reduce TTA**:
- Multiple notification channels (phone, SMS, push)
- Clear on-call schedule
- Backup on-call (auto-escalation)
- Alert fatigue reduction

---

#### 3. Time to Respond (TTR-initial)

**Definition**: Time from alert to beginning active investigation

```
TTR-initial = Investigation Start Time - Alert Time
```

**Target**: < 10 minutes

**Measurement**:
- Alert Time: PagerDuty alert timestamp
- Investigation Start: First message in incident channel or log entry

**Example**:
```
Alert Fired: 14:24:30 UTC
Investigation Started: 14:28:00 UTC
TTR-initial = 3.5 minutes ✅
```

---

#### 4. Time to Resolve (TTR)

**Definition**: Total time from incident start to full resolution

```
TTR = Resolution Time - Incident Start Time
```

**Targets**:
- SEV-1: < 1 hour
- SEV-2: < 2 hours
- SEV-3: < 24 hours
- SEV-4: < 1 week

**Measurement**:
- Incident Start: First user impact
- Resolution: Service fully restored and stable

**Example**:
```
Incident Start: 14:23:00 UTC
Resolved: 15:10:00 UTC
TTR = 47 minutes ✅ (SEV-2, under 2 hour target)
```

---

#### 5. Mean Time to Recovery (MTTR)

**Definition**: Average TTR across all incidents

```
MTTR = Sum(TTR for all incidents) / Number of incidents
```

**Target**:
- SEV-1/SEV-2: < 30 minutes
- All severities: < 2 hours

**Example**:
```
Last 10 incidents:
- SEV-1: 45 min, 62 min
- SEV-2: 35 min, 28 min, 41 min, 52 min
- SEV-3: 2 hours, 3 hours, 1.5 hours, 4 hours

MTTR (SEV-1/2 only) = (45+62+35+28+41+52) / 6 = 43.8 minutes
```

---

### Frequency Metrics

#### 1. Incident Rate

**Definition**: Number of incidents per time period

```
Incident Rate = Total Incidents / Time Period
```

**Targets**:
- SEV-1: < 1 per quarter
- SEV-2: < 2 per month
- SEV-3: < 10 per month

**Example**:
```
Q1 2026:
- SEV-1: 1 incident
- SEV-2: 4 incidents
- SEV-3: 23 incidents
- SEV-4: 67 incidents

SEV-1 Rate: 1 per quarter ✅
SEV-2 Rate: 1.33 per month ✅
```

---

#### 2. Repeat Incident Rate

**Definition**: Percentage of incidents that are repeats of previous incidents

```
Repeat Rate = (Repeat Incidents / Total Incidents) × 100%
```

**Target**: < 10%

**Classification**: Incident is "repeat" if same root cause within 90 days

**Example**:
```
Q1 2026: 28 total incidents (SEV-1 to SEV-3)
Repeat incidents: 2
- Database connection pool (again)
- API timeout (same bug reintroduced)

Repeat Rate = (2 / 28) × 100% = 7.1% ✅
```

**Action**: High repeat rate indicates action items not being completed

---

#### 3. False Positive Rate

**Definition**: Percentage of alerts that don't require action

```
False Positive Rate = (False Positive Alerts / Total Alerts) × 100%
```

**Target**: < 10%

**Example**:
```
Last week: 156 alerts
False positives: 23 (auto-resolved, known noise, etc.)

False Positive Rate = (23 / 156) × 100% = 14.7% ⚠️ (above target)
```

**Action**: High false positive rate causes alert fatigue

---

### Impact Metrics

#### 1. User Impact

**Metrics**:
- Number of users affected
- Percentage of user base affected
- Geographic distribution of impact
- Duration of impact per user

**Example**:
```
Incident: API Outage
- Total users affected: 1,200
- Total active users (at time): 8,000
- Impact percentage: 15%
- Duration: 47 minutes
- User-minutes of downtime: 1,200 × 47 = 56,400
```

---

#### 2. Business Impact

**Metrics**:
- Revenue impact ($ lost)
- SLO/SLA violations
- Error budget consumed
- Customer complaints
- Support ticket volume
- Churn risk (high-value customers affected)

**Example**:
```
Incident: Payment Processing Failure
- Duration: 35 minutes
- Failed transactions: 127
- Revenue impact: $8,450
- SLO violation: Payment Success Rate (99.9% → 94.2%)
- Error budget consumed: 45% of monthly budget
- Support tickets: 18
- Escalations: 2 (high-value customers)
```

---

#### 3. Error Budget Impact

**Definition**: How much error budget each incident consumed

```
Error Budget Consumed = (Incident Duration / Budget Period) × 100%
```

**Example**:
```
Monthly availability SLO: 99.9% (43.2 minutes downtime allowed)
Incident duration: 47 minutes

Budget consumed = (47 / 43.2) × 100% = 108.8% ⚠️
(Exceeded monthly budget in single incident!)
```

---

### Quality Metrics

#### 1. Detection Quality

**Metrics**:
- Automated detection rate (vs user reports)
- Time to detect (TTD)
- Alert accuracy (true positive rate)

**Target**: 90% of incidents detected by automated monitoring

**Example**:
```
Last month: 24 incidents
- Automated detection: 21 (87.5%)
- User reports: 2 (8.3%)
- Internal discovery: 1 (4.2%)

Automated detection rate: 87.5% ⚠️ (below 90% target)
```

---

#### 2. Response Quality

**Metrics**:
- Runbook usage rate
- Escalation appropriateness
- Communication timeliness
- Documentation completeness

**Example**:
```
Incident: Database Connection Pool
- Runbook used: Yes ✅
- Runbook helpful: Partially (needed updates)
- Escalation: Appropriate (DB team engaged)
- Communication: Updates every 10 minutes ✅
- Timeline documented: Complete ✅
- Postmortem quality: 8/10
```

---

#### 3. Postmortem Quality

**Metrics**:
- Postmortem completion rate
- Time to postmortem (days)
- Action item completion rate
- Action item completion time

**Targets**:
- 100% of SEV-1/SEV-2 have postmortems
- Published within 5 business days
- 90% of action items completed within deadline

**Example**:
```
Q1 2026:
- SEV-1 incidents: 1
- SEV-2 incidents: 4
- Postmortems completed: 5/5 (100%) ✅
- Avg time to postmortem: 3.2 days ✅
- Action items created: 47
- Action items completed: 39 (83%) ⚠️
- Avg completion time: 18 days
```

---

## Dashboards

### Real-Time Incident Dashboard

**Purpose**: Monitor active incidents and current status

**Panels**:

1. **Active Incidents**
   - Current SEV-1/SEV-2 incidents
   - Duration
   - Incident Commander
   - Status (investigating/identified/monitoring)

2. **System Health**
   - All services status
   - Error rates
   - Latency percentiles
   - Availability metrics

3. **Recent Incidents** (last 24 hours)
   - Timeline
   - Severity
   - Duration
   - Status

4. **On-Call Status**
   - Current on-call engineer
   - Alerts in last hour
   - Response time (last 5 alerts)

**URL**: https://grafana.virtengine.com/d/incident-dashboard

---

### Historical Metrics Dashboard

**Purpose**: Track trends and measure improvement over time

**Panels**:

1. **Incident Volume** (time series)
   - Incidents per week by severity
   - Trend line
   - Target threshold

2. **MTTR Trend** (time series)
   - Average TTR per week
   - By severity
   - Trend line
   - Target threshold

3. **Detection Metrics**
   - Time to Detect (TTD) distribution
   - Time to Acknowledge (TTA) distribution
   - Detection method breakdown

4. **Impact Metrics**
   - User impact (user-minutes of downtime)
   - Error budget consumption
   - Revenue impact

5. **Quality Metrics**
   - Postmortem completion rate
   - Action item completion rate
   - Repeat incident rate
   - False positive rate

6. **Heatmap**
   - Incidents by day of week and hour
   - Identify patterns

**URL**: https://grafana.virtengine.com/d/incident-metrics

---

### Executive Dashboard

**Purpose**: High-level metrics for leadership

**Panels**:

1. **Monthly Summary**
   - Total incidents by severity
   - MTTR by severity
   - User impact (total downtime)
   - Error budget status

2. **Availability**
   - Monthly availability percentage
   - SLO compliance
   - Trend over 12 months

3. **Top Incidents** (by impact)
   - Severity
   - Duration
   - User impact
   - Status (resolved/action items in progress)

4. **Improvement Trends**
   - MTTR improvement (%)
   - Repeat incident rate
   - Detection time improvement

**URL**: https://grafana.virtengine.com/d/executive-dashboard

---

## Reporting

### Weekly Incident Report

**Audience**: Engineering team

**Distribution**: Every Monday via email + Slack

**Template**:

```
🚨 Weekly Incident Report: [Date Range]

SUMMARY:
- Total incidents: [N]
  - SEV-1: [N]
  - SEV-2: [N]
  - SEV-3: [N]
  - SEV-4: [N]
- MTTR: [X] minutes (target: 30 min)
- User impact: [X] user-minutes downtime

NOTABLE INCIDENTS:
1. [Date] SEV-2: [Service] - [Brief description]
   - Duration: [X] min
   - Root cause: [Brief]
   - Postmortem: [Link]

TOP ISSUES:
1. [Issue category]: [N] incidents
2. [Issue category]: [N] incidents

ACTION ITEMS:
- In progress: [N]
- Completed this week: [N]
- Overdue: [N] ⚠️

IMPROVEMENTS:
- [Notable improvement or achievement]

Dashboard: https://grafana.virtengine.com/d/incident-metrics
```

---

### Monthly Incident Review

**Audience**: Engineering leadership + Product

**Distribution**: First week of month

**Contents**:

1. **Executive Summary**
   - Key metrics vs targets
   - Major incidents
   - Trends and patterns

2. **Detailed Metrics**
   - All key metrics with trends
   - Comparison to previous month
   - Year-over-year comparison

3. **Root Cause Analysis**
   - Breakdown by category
   - Repeat incidents
   - Systemic issues

4. **Improvement Initiatives**
   - Action item completion
   - Infrastructure improvements
   - Process improvements

5. **Looking Ahead**
   - Upcoming risks
   - Planned improvements
   - Resource needs

---

### Quarterly Business Review

**Audience**: Executive team, Board (optional)

**Contents**:

1. **Reliability Highlights**
   - Availability vs SLO
   - Incident trends
   - MTTR improvements
   - Major achievements

2. **Business Impact**
   - Downtime cost ($ impact)
   - Customer satisfaction impact
   - SLA credits issued

3. **Major Incidents**
   - Deep dive on SEV-1 incidents
   - Lessons learned
   - Improvements implemented

4. **Investments and ROI**
   - Infrastructure improvements
   - Tooling investments
   - Team growth
   - Measurable impact

5. **Forward Looking**
   - Upcoming initiatives
   - Risk mitigation
   - Resource requests

---

## Analysis and Insights

### Pattern Recognition

**Common Patterns to Identify**:

1. **Temporal Patterns**
   - Time of day (e.g., 2-4am incidents)
   - Day of week (e.g., Monday mornings)
   - Seasonal (e.g., end of month traffic)

2. **Correlation Patterns**
   - After deployments
   - During traffic spikes
   - Following configuration changes
   - External dependency issues

3. **Root Cause Patterns**
   - Same component repeatedly failing
   - Capacity issues
   - Configuration drift

**Action**: Use patterns to prioritize improvements

---

### Root Cause Distribution

**Example Analysis**:

```
Q1 2026 Root Cause Breakdown:
- Code bugs: 35% (9 incidents)
- Capacity/scaling: 27% (7 incidents)
- Configuration errors: 19% (5 incidents)
- External dependencies: 12% (3 incidents)
- Unknown: 7% (2 incidents)

Action: Focus on:
1. Better testing (catch bugs before production)
2. Capacity planning and auto-scaling
3. Configuration management and validation
```

---

### Service Reliability Ranking

**Rank services by incident rate and impact**:

| Service | Incidents | MTTR | User Impact | Reliability Score |
|---------|-----------|------|-------------|-------------------|
| API Gateway | 12 | 35 min | High | ⚠️ 6.5/10 |
| Database | 8 | 52 min | Critical | ⚠️ 7.0/10 |
| VEID | 3 | 25 min | Medium | ✅ 8.5/10 |
| Provider Daemon | 6 | 40 min | Medium | ⚠️ 7.5/10 |
| Blockchain Core | 2 | 15 min | Low | ✅ 9.0/10 |

**Action**: Prioritize improvements for lowest-scoring services

---

### Cost of Downtime Analysis

**Calculate total cost of incidents**:

```
Cost Components:
1. Lost revenue (failed transactions)
2. SLA credits issued
3. Engineering time (response + follow-up)
4. Customer support time
5. Reputation/churn risk

Example (Q1 2026):
- Lost revenue: $45,000
- SLA credits: $12,000
- Engineering time: 240 hours × $150/hr = $36,000
- Support time: 80 hours × $75/hr = $6,000
- Total: $99,000

Cost per incident: $99,000 / 26 incidents = $3,808

ROI Calculation:
- Investment in reliability: $200,000/year
- Estimated incident reduction: 50%
- Estimated cost savings: $200,000/year
- ROI: 100% (breakeven)
```

---

## Continuous Improvement

### Improvement Framework

**Monthly Improvement Cycle**:

1. **Measure** (Week 1)
   - Collect all metrics
   - Generate reports
   - Identify trends

2. **Analyze** (Week 2)
   - Review patterns
   - Root cause analysis
   - Prioritize issues

3. **Plan** (Week 3)
   - Define improvement initiatives
   - Assign owners
   - Set deadlines

4. **Execute** (Week 4)
   - Implement improvements
   - Track progress
   - Measure impact

---

### Key Performance Indicators (KPIs)

**Primary KPIs**:

1. **Availability**: 99.9% uptime (monthly)
2. **MTTR**: < 30 minutes (SEV-1/SEV-2)
3. **Incident Rate**: < 2 SEV-2 per month
4. **Action Item Completion**: > 90% within deadline
5. **Repeat Incidents**: < 10%

**Review Cadence**:
- Weekly: Team review
- Monthly: Leadership review
- Quarterly: Executive review

---

### Improvement Tracking

**Track improvement initiatives**:

| Initiative | Metric Impact | Status | ROI |
|------------|---------------|--------|-----|
| Auto-scaling implementation | Reduced capacity incidents by 60% | Complete | High |
| Alerting refinement | Reduced false positives from 18% → 8% | Complete | Medium |
| Runbook improvements | Reduced MTTR by 12 minutes | In Progress | High |
| Chaos engineering | Early detection of issues | Planned | High |

---

## Automated Reporting Setup

### Grafana Alerts

**Configure alerts on metric thresholds**:

```yaml
# Example: High Incident Rate Alert
alert: HighIncidentRate
expr: sum(increase(incidents_total[7d])) > 10
for: 1d
annotations:
  summary: "Incident rate above target"
  description: "{{ $value }} incidents in last 7 days (target: < 10)"
```

---

### Scheduled Reports

**Automate report generation**:

1. **Weekly Report**
   - Grafana snapshot
   - Jira query (incidents last week)
   - Automated email

2. **Monthly Report**
   - Grafana dashboard PDF export
   - Jira metrics export
   - Action item status

---

## References

- [Incident Response Process](INCIDENT_RESPONSE.md)
- [SLI/SLO/SLA Framework](SLI_SLO_SLA.md)
- [Error Budget Policy](ERROR_BUDGET_POLICY.md)
- [Grafana Dashboards](https://grafana.virtengine.com)
- [PagerDuty Analytics](https://virtengine.pagerduty.com/analytics)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Next Review**: 2026-04-30
