# Incident Response Process

## Overview

This document defines the incident response process for VirtEngine services. An effective incident response minimizes impact, accelerates recovery, and facilitates learning.

## Table of Contents
1. [Incident Definition](#incident-definition)
2. [Severity Levels](#severity-levels)
3. [Incident Lifecycle](#incident-lifecycle)
4. [Roles and Responsibilities](#roles-and-responsibilities)
5. [Communication](#communication)
6. [Tools and Systems](#tools-and-systems)
7. [Post-Incident Review](#post-incident-review)

---

## Incident Definition

### What is an Incident?

An **incident** is an unplanned interruption or reduction in quality of service that requires immediate response to restore normal operations.

**Examples of Incidents**:
- Service outage or significant degradation
- SLO violation
- Security breach or suspected compromise
- Data loss or corruption
- Critical bug affecting users
- Infrastructure failure

**NOT Incidents**:
- Known issues with workarounds
- Planned maintenance
- Non-urgent bugs
- Feature requests
- Performance optimization opportunities

---

## Severity Levels

### SEV-1: Critical

**Definition**: Complete service outage or critical security breach

**Criteria**:
- Complete service unavailable for all users
- Major functionality completely broken
- Active security breach or data loss
- Revenue impact > $10,000/hour
- SLO budget depleted

**Response**:
- Page entire on-call rotation immediately
- Establish incident command within 5 minutes
- Executive team notified within 15 minutes
- Status page updated immediately
- All hands on deck (pull engineers from other work)

**SLA**: Response < 5 minutes, Resolution target < 1 hour

**Recent Example**: Complete API outage affecting all users

---

### SEV-2: High

**Definition**: Significant service degradation or SLO violation

**Criteria**:
- Major functionality degraded or broken
- Affecting significant subset of users (> 10%)
- SLO violated (error budget consuming rapidly)
- Revenue impact $1,000-$10,000/hour
- Security vulnerability discovered (not actively exploited)

**Response**:
- Page on-call engineer
- Establish incident command within 15 minutes
- Engineering leadership notified within 30 minutes
- Status page updated within 15 minutes
- Dedicated engineering resources assigned

**SLA**: Response < 15 minutes, Resolution target < 2 hours

**Recent Example**: Database connection pool exhaustion (47-minute outage)

---

### SEV-3: Medium

**Definition**: Minor service degradation, no SLO impact

**Criteria**:
- Non-critical functionality affected
- Small subset of users affected (< 10%)
- Performance degraded but within SLO
- Revenue impact < $1,000/hour
- Intermittent errors

**Response**:
- Alert on-call engineer (no page)
- Investigate during business hours
- Engineering lead notified
- Status page optional
- Fix during next deployment window

**SLA**: Response < 1 hour, Resolution target < 24 hours

**Recent Example**: Slow query causing intermittent timeouts for 5% of requests

---

### SEV-4: Low

**Definition**: Minimal impact, cosmetic issues

**Criteria**:
- Minor cosmetic issues
- Very small user impact
- No functionality broken
- No revenue impact
- Can be fixed in regular development cycle

**Response**:
- Create ticket in backlog
- No immediate action required
- Fix in regular sprint

**SLA**: Response < 24 hours, Resolution target < 1 week

**Recent Example**: Dashboard formatting issue affecting 1 metric display

---

## Incident Lifecycle

### Phase 1: Detection (0-5 minutes)

**Goal**: Identify that an incident is occurring

**Activities**:
1. **Automated Detection**: Monitoring alerts fire
2. **Manual Detection**: User report, internal discovery
3. **Alert Routing**: Page reaches on-call engineer
4. **Acknowledgement**: On-call acknowledges page

**Tools**:
- Prometheus/Grafana for monitoring
- PagerDuty for alerting
- Status page for user reports

**Success Criteria**:
- Incident detected within 5 minutes of start
- Alert reaches on-call within 1 minute
- On-call acknowledges within 2 minutes

---

### Phase 2: Response (5-15 minutes)

**Goal**: Assemble team and begin triage

**Activities**:
1. **Initial Assessment**: Determine severity
2. **Incident Declaration**: Declare incident with severity level
3. **Team Assembly**: Page additional responders if needed
4. **Communication Setup**: Open incident Slack channel
5. **Incident Command**: Assign Incident Commander (IC)
6. **Status Page**: Update status page (SEV-1/SEV-2)

**Checklist**:
- [ ] Severity determined
- [ ] Incident Slack channel created (#incident-YYYY-MM-DD-HHMM)
- [ ] Incident Commander assigned
- [ ] Status page updated
- [ ] Initial stakeholders notified

**Communication Template**:
```
🚨 INCIDENT DECLARED 🚨

Severity: SEV-2
Service: API Service
Impact: Complete outage for all users
Incident Channel: #incident-2026-01-29-1423
Incident Commander: @bob-jones
Status Page: https://status.virtengine.com

Join the incident channel if you can help.
```

---

### Phase 3: Investigation (15 minutes - 1 hour)

**Goal**: Identify root cause and mitigation strategy

**Activities**:
1. **Gather Information**: Metrics, logs, traces
2. **Form Hypotheses**: What could cause this?
3. **Test Hypotheses**: Validate theories
4. **Identify Root Cause**: What is the underlying issue?
5. **Plan Mitigation**: How can we restore service?

**Investigation Checklist**:
- [ ] Recent deployments reviewed
- [ ] Metrics analyzed (CPU, memory, disk, network)
- [ ] Logs reviewed for errors
- [ ] Database health checked
- [ ] External dependencies checked
- [ ] Traffic patterns analyzed

**Communication**:
- Update incident channel every 10-15 minutes
- Share findings, hypotheses, and next steps
- Ask for help if needed

**Example Update**:
```
[14:35] Bob (IC): Update - Investigation
- Recent deployment: API v1.2.3 at 14:15
- Symptom: DB connection pool exhausted (100/100)
- Hypothesis: New feature increased query load
- Next step: Attempting rollback
```

---

### Phase 4: Mitigation (1-2 hours)

**Goal**: Restore service to operational state

**Strategies**:

**1. Rollback** (Preferred)
- Roll back recent deployment
- Fastest path to recovery
- Downside: Lose new features

**2. Configuration Change**
- Adjust parameters (connection pool, timeout, etc.)
- Faster than code change
- Good for resource-related issues

**3. Traffic Reduction**
- Enable rate limiting
- Disable non-critical features
- Reduce load on system

**4. Failover**
- Switch to backup region/datacenter
- For infrastructure failures
- May have data staleness

**5. Scale Up**
- Add more instances/resources
- For capacity issues
- Takes time to provision

**6. Hotfix**
- Deploy emergency fix
- For bugs requiring code change
- Risky, thorough testing needed

**Decision Tree**:
```
Can we rollback?
├─ Yes → ROLLBACK
└─ No → Can we change config?
    ├─ Yes → CONFIG CHANGE
    └─ No → Can we reduce traffic?
        ├─ Yes → RATE LIMIT
        └─ No → HOTFIX or FAILOVER
```

**Communication**:
```
[14:50] Bob (IC): Mitigation in Progress
- Strategy: Increase DB connection pool (100 → 500)
- ETA: 5 minutes
- Risk: Low (config change, requires restart)
- Rollback plan: Revert config if issue persists
```

---

### Phase 5: Resolution (2+ hours)

**Goal**: Confirm service is fully restored and stable

**Activities**:
1. **Apply Fix**: Execute mitigation plan
2. **Verify Health**: Check all health metrics
3. **Monitor Stability**: Watch for 15-30 minutes
4. **Confirm Recovery**: All SLIs back to normal
5. **Declare Resolution**: Service restored

**Resolution Checklist**:
- [ ] Fix applied successfully
- [ ] Health checks passing
- [ ] Error rate back to normal
- [ ] Latency back to normal
- [ ] SLI metrics within target
- [ ] No new errors in logs
- [ ] Synthetic checks passing
- [ ] User reports stopped

**Communication**:
```
[15:10] Bob (IC): INCIDENT RESOLVED ✅
- Mitigation: Connection pool increased to 500
- Service Status: Fully restored
- Error Rate: < 0.5% (normal)
- Latency: P95 < 2s (normal)
- Monitoring: Continuing for 30 minutes to confirm stability
```

---

### Phase 6: Post-Incident (24-48 hours)

**Goal**: Learn from the incident and prevent recurrence

**Activities**:
1. **Close Incident**: Mark incident as resolved
2. **Schedule Postmortem**: Within 48 hours of resolution
3. **Collect Data**: Metrics, logs, timeline
4. **Write Postmortem**: Use blameless template
5. **Action Items**: Create tickets for improvements
6. **Review Meeting**: Discuss learnings with team

**Postmortem Deadline**: Within 5 business days of incident

**Communication**:
```
[15:30] Bob (IC): Incident Closed
- Total Duration: 47 minutes
- Impact: 1,200 users, 78% error budget consumed
- Postmortem: Scheduled for 2026-01-30 10:00 AM
- Postmortem Doc: [Link to draft]
- Action Items: 12 items created
- Thank you to all responders! 🙏
```

---

## Roles and Responsibilities

### On-Call Engineer (Primary Responder)

**Responsibilities**:
- First responder to all incidents
- Acknowledge alerts within 5 minutes
- Perform initial triage and investigation
- Determine severity
- Declare incident if necessary
- Act as Incident Commander for SEV-3/SEV-4
- Escalate to IC for SEV-1/SEV-2

**Skills Required**:
- Deep knowledge of VirtEngine architecture
- Debugging and troubleshooting skills
- Access to all production systems
- Calm under pressure

**Shift Duration**: 1 week (Monday 9am - Monday 9am)

**Escalation**: Secondary on-call if no response within 5 minutes

---

### Incident Commander (IC)

**Responsibilities**:
- Lead incident response for SEV-1/SEV-2
- Coordinate all response activities
- Make high-level decisions
- Delegate tasks to responders
- Maintain communication with stakeholders
- Ensure incident timeline is documented
- Declare resolution when service is restored

**Authority**:
- Can pull any engineer from other work
- Can make emergency changes without normal approval
- Can escalate to executive team
- Final decision maker during incident

**Who**: Senior SRE or Engineering Lead

**Qualities**:
- Strong communication skills
- Calm under pressure
- Deep technical knowledge
- Leadership experience

---

### Subject Matter Expert (SME)

**Responsibilities**:
- Provide deep technical expertise on specific components
- Investigate and diagnose issues in their domain
- Propose mitigation strategies
- Execute fixes under IC direction
- Document technical details for postmortem

**Examples**:
- Database SME for database issues
- Network SME for network issues
- Security SME for security incidents

---

### Communications Lead

**Responsibilities**:
- Manage external communication (status page, social media)
- Interface with customer support team
- Prepare stakeholder updates
- Draft incident summaries for leadership
- Handle press inquiries (with leadership approval)

**Updates**:
- Status page: Every 15 minutes during active incident
- Support team: Immediately when incident declared
- Leadership: Every 30 minutes for SEV-1, hourly for SEV-2
- Social media: As appropriate

---

### Scribe

**Responsibilities**:
- Document incident timeline in detail
- Record all actions taken
- Log all hypotheses and findings
- Capture communications
- Create timeline for postmortem

**Format**:
```
[14:23] Error rate spike detected
[14:24] Alert: APIErrorRateHigh fired
[14:26] On-call paged
[14:28] Alice acknowledged page
[14:30] Alice: "Checking recent deployments"
[14:32] Alice: "Connection pool exhausted, 100/100 connections"
...
```

---

## Communication

### Internal Communication

#### Incident Slack Channel

**Naming Convention**: `#incident-YYYY-MM-DD-HHMM`

**Example**: `#incident-2026-01-29-1423`

**Purpose**:
- Central coordination point
- All incident-related discussion
- Status updates
- Hypothesis sharing
- Action coordination

**Pinned Items**:
- Incident severity and description
- Incident Commander name
- Link to status page
- Link to dashboard
- Link to incident ticket

---

#### Update Cadence

**SEV-1**:
- Every 10 minutes during active response
- Every 30 minutes during monitoring

**SEV-2**:
- Every 15 minutes during active response
- Every hour during monitoring

**SEV-3/4**:
- Significant milestones only
- At least once at start and resolution

---

### External Communication

#### Status Page

**URL**: https://status.virtengine.com

**Update Template**:

**Investigating**:
```
We are currently investigating issues with [Service Name].
Users may experience [specific impact].
Updates will be provided every 15 minutes.
```

**Identified**:
```
We have identified the cause of the issue affecting [Service Name].
[Brief description of cause].
We are working on a fix. ETA: [time estimate or "unknown"].
```

**Monitoring**:
```
A fix has been applied and we are monitoring the results.
Service should be restored shortly.
```

**Resolved**:
```
The issue affecting [Service Name] has been resolved.
Total duration: [X] minutes.
A full post-incident review will be published within 5 days.
```

---

#### Stakeholder Notification

**Leadership (CTO, CEO)**:
- SEV-1: Immediately
- SEV-2: Within 30 minutes
- SEV-3: Daily summary

**Customer Support Team**:
- All SEV-1/SEV-2: Immediately
- Provide FAQ for user questions
- Keep updated on resolution progress

**Engineering Teams**:
- Incident Slack channel (all can monitor)
- Direct page only if specific expertise needed

---

## Tools and Systems

### Monitoring and Alerting

**Prometheus**: Metrics collection
- Endpoint: https://prometheus.virtengine.com
- Alerts defined in: `/etc/prometheus/alerts/`

**Grafana**: Visualization
- URL: https://grafana.virtengine.com
- Incident Dashboard: https://grafana.virtengine.com/d/incident-overview

**PagerDuty**: On-call management
- URL: https://virtengine.pagerduty.com
- Escalation policies configured
- Integration with Prometheus Alertmanager

---

### Logging and Tracing

**Elasticsearch**: Log aggregation
- URL: https://logs.virtengine.com
- Retention: 30 days

**Jaeger**: Distributed tracing
- URL: https://tracing.virtengine.com
- Retention: 7 days

---

### Incident Management

**PagerDuty Incidents**:
- Create incident for tracking
- Link to Slack channel
- Timeline auto-generated

**Jira Tickets**:
- Create ticket for postmortem tracking
- Template: "INCIDENT-YYYY-MM-DD: [Description]"
- Link to postmortem doc

---

### Communication Tools

**Slack**:
- Incident channels: `#incident-*`
- General SRE: `#sre-team`
- Alerts: `#sre-alerts`

**Status Page**:
- Statuspage.io or similar
- API for automated updates

**Zoom**:
- War room for complex incidents
- Recorded for postmortem

---

## Post-Incident Review

### Postmortem Meeting

**Timing**: Within 48 hours of resolution

**Duration**: 60 minutes

**Attendees**:
- All incident responders
- Service owners
- Engineering leads
- SRE team

**Agenda**:
1. **Timeline Review** (15 min)
   - Walk through what happened
   - Clarify any gaps or confusion

2. **Root Cause Discussion** (15 min)
   - What was the underlying cause?
   - Why did it happen?
   - What allowed it to happen?

3. **Response Review** (15 min)
   - What went well?
   - What went poorly?
   - How can we improve response?

4. **Action Items** (15 min)
   - What can prevent recurrence?
   - What can improve detection?
   - What can improve response?
   - Assign owners and deadlines

---

### Postmortem Document

**Template**: [Use postmortem_template.md](templates/postmortem_template.md)

**Distribution**:
- Engineering team (all)
- Product team
- Leadership team
- Public blog (sanitized version, optional)

**Timeline**:
- Draft: Within 3 days of incident
- Review: 2 business days for comments
- Final: Published within 5 business days

---

### Action Item Tracking

**Requirements**:
- Every action item must have:
  - Clear description
  - Owner assigned
  - Due date
  - Priority (P0-P3)
  - Acceptance criteria

**Follow-up**:
- Review in weekly SRE sync
- 30-day review: Check completion
- 90-day review: Assess effectiveness

**Metrics**:
- Action item completion rate: > 90% within deadline
- Time to complete P0 items: < 7 days

---

## Incident Response Training

### On-Call Training

**Required Before First Shift**:
- [ ] Shadow 2 on-call shifts
- [ ] Complete incident response training
- [ ] Read all runbooks
- [ ] Access to all systems verified
- [ ] PagerDuty configured
- [ ] Practice gameday exercise

**Ongoing Training**:
- Quarterly gameday exercises
- Runbook reviews
- Postmortem review meetings

---

### Incident Commander Training

**Prerequisites**:
- 6+ months as on-call engineer
- Responded to 5+ incidents
- Deep technical knowledge
- Strong communication skills

**Training Program**:
- Shadow 3 incidents as IC trainee
- Lead SEV-3 incident under supervision
- Complete IC training course
- Pass IC certification exercise

---

## Incident Metrics

### Response Metrics

**Time to Detect (TTD)**:
```
TTD = Alert Time - Incident Start Time

Target: < 5 minutes
```

**Time to Acknowledge (TTA)**:
```
TTA = Acknowledgement Time - Alert Time

Target: < 2 minutes
```

**Time to Resolve (TTR)**:
```
TTR = Resolution Time - Incident Start Time

Target:
- SEV-1: < 1 hour
- SEV-2: < 2 hours
- SEV-3: < 24 hours
```

**Mean Time to Recovery (MTTR)**:
```
MTTR = Average TTR across all incidents

Target: < 30 minutes (SEV-1/SEV-2)
```

---

### Incident Frequency

**Incident Rate**:
```
Incidents per Week = Total Incidents / Weeks

Target: < 2 SEV-1/SEV-2 per month
```

**Repeat Incidents**:
```
Repeat Rate = Repeat Incidents / Total Incidents

Target: < 10%
```

---

## References

- [Postmortem Template](templates/postmortem_template.md)
- [SLI/SLO/SLA Framework](SLI_SLO_SLA.md)
- [Error Budget Policy](ERROR_BUDGET_POLICY.md)
- [Runbooks](runbooks/README.md)
- [Google SRE Book - Incident Management](https://sre.google/sre-book/managing-incidents/)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
