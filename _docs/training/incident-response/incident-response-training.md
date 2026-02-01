# Incident Response Training

**Module Duration:** 8 hours  
**Target Audience:** On-call engineers, SREs, Engineering leads  
**Prerequisites:** Basic understanding of VirtEngine architecture  
**Version:** 1.0.0  
**Last Updated:** 2026-01-30

---

## Module Overview

This comprehensive training module prepares engineers to effectively respond to incidents in VirtEngine production environments. By the end of this training, participants will be able to confidently detect, triage, respond to, and resolve incidents while maintaining clear communication with stakeholders.

### Learning Objectives

Upon completing this module, you will be able to:

1. **Recognize** the full incident lifecycle from detection to post-incident review
2. **Classify** incidents using the severity framework (SEV-1 through SEV-4)
3. **Execute** proper escalation procedures based on incident severity
4. **Perform** the Incident Commander role for SEV-1/SEV-2 incidents
5. **Communicate** effectively with stakeholders during incidents
6. **Navigate** incident response tools including PagerDuty, Slack, and monitoring dashboards
7. **Apply** structured problem-solving techniques under pressure
8. **Document** incidents for effective post-incident analysis

---

## Training Schedule

| Session | Topic                                | Duration  |
| ------- | ------------------------------------ | --------- |
| 1       | Incident Lifecycle Overview          | 1 hour    |
| 2       | Severity Classification & Escalation | 1.5 hours |
| 3       | Incident Commander Role              | 1.5 hours |
| 4       | Communication Protocols              | 1 hour    |
| 5       | Tools and Access                     | 1 hour    |
| 6       | Role-Playing Exercises               | 1.5 hours |
| 7       | Assessment & Certification           | 0.5 hours |

---

## Session 1: Incident Lifecycle Overview (1 hour)

### What is an Incident?

An **incident** is an unplanned interruption or reduction in quality of service that requires immediate response to restore normal operations.

**Examples of Incidents:**

- Complete service outage
- Significant performance degradation
- SLO violations
- Security breaches
- Data loss or corruption
- Consensus failures (chain halt)

**NOT Incidents:**

- Known issues with documented workarounds
- Planned maintenance windows
- Non-urgent bugs in backlog
- Feature requests
- Minor cosmetic issues

### The Incident Lifecycle

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  DETECTION  │ → │   TRIAGE    │ → │  RESPONSE   │ → │ RESOLUTION  │ → │ POST-MORTEM │
│  (0-5 min)  │   │  (5-15 min) │   │ (15+ min)   │   │ (varies)    │   │ (24-48 hrs) │
└─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘
```

---

### Phase 1: Detection (0-5 minutes)

**Goal:** Identify that an incident is occurring as quickly as possible.

**Detection Sources:**
| Source | Description | Example |
|--------|-------------|---------|
| Automated Monitoring | Prometheus/Grafana alerts | `APIErrorRateHigh` alert fires |
| PagerDuty | On-call paging | Primary on-call receives page |
| User Reports | Customer-reported issues | Support ticket: "Can't deploy" |
| Internal Discovery | Engineer notices problem | "Dashboard is slow" |

**Key Metrics:**

- **Time to Detect (TTD):** Time from incident start to alert firing
- **Target:** < 5 minutes for SEV-1/SEV-2

**Detection Checklist:**

- [ ] Alert received and acknowledged
- [ ] Initial symptom identified
- [ ] Affected service(s) determined
- [ ] Preliminary scope understood

---

### Phase 2: Triage (5-15 minutes)

**Goal:** Assess severity and assemble the response team.

**Triage Activities:**

1. **Severity Assessment:** Determine SEV-1, SEV-2, SEV-3, or SEV-4
2. **Impact Assessment:** Who is affected? How many users?
3. **Incident Declaration:** Officially declare an incident
4. **Team Assembly:** Page additional responders if needed
5. **Communication Setup:** Create incident Slack channel

**Triage Questions:**

- What service(s) are affected?
- What is the user impact?
- Is the issue getting worse?
- Are there any recent changes (deployments, configs)?
- Does this match any known issues?

**Triage Checklist:**

- [ ] Severity determined
- [ ] Incident Slack channel created (#incident-YYYY-MM-DD-HHMM)
- [ ] Incident Commander assigned
- [ ] Status page updated (SEV-1/SEV-2)
- [ ] Initial stakeholders notified

---

### Phase 3: Response (15+ minutes)

**Goal:** Investigate root cause and implement mitigation.

**Response Activities:**

**1. Investigation**

- Gather information (metrics, logs, traces)
- Form hypotheses about root cause
- Test hypotheses systematically
- Identify contributing factors

**2. Mitigation Strategy**
| Strategy | When to Use | Risk |
|----------|-------------|------|
| Rollback | Recent deployment caused issue | Low |
| Config Change | Parameter adjustment needed | Low |
| Traffic Reduction | System overloaded | Medium |
| Failover | Infrastructure failure | Medium |
| Scale Up | Capacity issues | Medium |
| Hotfix | Code bug requires fix | High |

**Response Checklist:**

- [ ] Investigation underway
- [ ] Root cause hypothesis formed
- [ ] Mitigation strategy selected
- [ ] Regular updates being posted
- [ ] Timeline being documented

---

### Phase 4: Resolution

**Goal:** Confirm service is fully restored and stable.

**Resolution Activities:**

1. Apply the selected mitigation
2. Verify health metrics return to normal
3. Monitor for stability (15-30 minutes)
4. Confirm all SLIs are within targets
5. Declare incident resolved

**Resolution Verification:**

```bash
# Check key health metrics
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Verify error rate is normal
curl -s "http://prometheus:9090/api/v1/query?query=virtengine_api_error_rate_5m"

# Check SLO status
curl -s "http://prometheus:9090/api/v1/query?query=virtengine_slo_error_budget_remaining"
```

**Resolution Checklist:**

- [ ] Fix applied successfully
- [ ] Health checks passing
- [ ] Error rate back to normal
- [ ] Latency back to normal
- [ ] SLI metrics within target
- [ ] No new errors in logs
- [ ] User reports stopped

---

### Phase 5: Post-Incident (24-48 hours)

**Goal:** Learn from the incident and prevent recurrence.

**Post-Incident Activities:**

1. Close incident officially
2. Schedule postmortem meeting (within 48 hours)
3. Collect all incident data (logs, metrics, timeline)
4. Write blameless postmortem document
5. Create action items with owners and deadlines
6. Present learnings to team

**Post-Incident Checklist:**

- [ ] Incident ticket closed
- [ ] Postmortem meeting scheduled
- [ ] Timeline documented
- [ ] Root cause identified
- [ ] Action items created
- [ ] Postmortem published

> **Reference:** See [Post-Incident Analysis Training](./post-incident-analysis.md) for detailed postmortem procedures.

---

## Session 2: Severity Classification & Escalation (1.5 hours)

### Severity Framework

VirtEngine uses a four-level severity classification system:

### SEV-1: Critical

**Definition:** Complete service outage or critical security breach

**Criteria:**
| Factor | Threshold |
|--------|-----------|
| User Impact | 100% of users affected |
| Functionality | Core service completely unavailable |
| Security | Active breach or data loss |
| Revenue Impact | > $10,000/hour |
| Error Budget | Depleted |
| Chain Status | Chain halted |

**Response Requirements:**

- Page entire on-call rotation immediately
- Establish incident command within 5 minutes
- Executive team notified within 15 minutes
- Status page updated immediately
- All hands on deck (pull engineers from other work)

**SLA:**

- Response Time: < 5 minutes
- Resolution Target: < 1 hour

**VirtEngine SEV-1 Examples:**

- Complete chain halt (BlockProductionStalled)
- All API endpoints returning 5xx errors
- VEID consensus divergence across validators
- Active security breach with data exfiltration
- Complete provider daemon cluster failure

---

### SEV-2: High

**Definition:** Significant service degradation or SLO violation

**Criteria:**
| Factor | Threshold |
|--------|-----------|
| User Impact | > 10% of users affected |
| Functionality | Major feature unavailable |
| Security | Vulnerability discovered (not exploited) |
| Revenue Impact | $1,000-$10,000/hour |
| Error Budget | Consuming rapidly |

**Response Requirements:**

- Page primary on-call engineer
- Establish incident command within 15 minutes
- Engineering leadership notified within 30 minutes
- Status page updated within 15 minutes
- Dedicated engineering resources assigned

**SLA:**

- Response Time: < 15 minutes
- Resolution Target: < 2 hours

**VirtEngine SEV-2 Examples:**

- Low validator count (approaching 2/3 threshold)
- Provider deployment success rate < 90%
- VEID verification latency > 10 seconds
- Database connection pool exhaustion
- High error rate (> 5%) on API endpoints

---

### SEV-3: Medium

**Definition:** Minor service degradation, no SLO impact

**Criteria:**
| Factor | Threshold |
|--------|-----------|
| User Impact | < 10% of users affected |
| Functionality | Non-critical feature affected |
| Performance | Degraded but within SLO |
| Revenue Impact | < $1,000/hour |

**Response Requirements:**

- Alert on-call engineer (no page)
- Investigate during business hours
- Engineering lead notified
- Status page optional
- Fix during next deployment window

**SLA:**

- Response Time: < 1 hour
- Resolution Target: < 24 hours

**VirtEngine SEV-3 Examples:**

- Single node down in multi-node cluster
- Slow queries affecting subset of requests
- Intermittent timeout errors (< 5% of requests)
- Secondary monitoring dashboard unavailable

---

### SEV-4: Low

**Definition:** Minimal impact, cosmetic issues

**Criteria:**

- Minor cosmetic issues
- Very small user impact
- No functionality broken
- No revenue impact
- Can be fixed in regular development cycle

**Response Requirements:**

- Create ticket in backlog
- No immediate action required
- Fix in regular sprint

**SLA:**

- Response Time: < 24 hours
- Resolution Target: < 1 week

**VirtEngine SEV-4 Examples:**

- Dashboard formatting issues
- Non-critical log warnings
- Documentation errors
- Minor UI inconsistencies

---

### Severity Decision Tree

```
                    Is service completely down?
                          /           \
                        Yes            No
                        /               \
                   SEV-1            Are >10% of users affected?
                                       /            \
                                     Yes             No
                                     /                \
                              Is SLO violated?    Is any user affected?
                                 /     \             /        \
                               Yes     No          Yes        No
                               /        \           \          \
                           SEV-2     SEV-2       SEV-3      SEV-4
```

---

### Escalation Procedures

#### Escalation Matrix

| Severity | First Responder | Escalate To       | Notify           | Timeline    |
| -------- | --------------- | ----------------- | ---------------- | ----------- |
| SEV-1    | Primary On-Call | IC + All SMEs     | CTO, CEO         | Immediately |
| SEV-2    | Primary On-Call | IC + Relevant SME | Engineering Lead | 15 minutes  |
| SEV-3    | Primary On-Call | Secondary On-Call | Team Lead        | 1 hour      |
| SEV-4    | Primary On-Call | N/A               | N/A              | Backlog     |

#### When to Escalate

**Escalate Immediately When:**

- Cannot identify root cause within 15 minutes
- Incident is getting worse
- You lack expertise in affected area
- Decision requires higher authority
- Customer communication needed
- Security implications suspected

**Escalation Channels:**

| Channel               | Use Case               | Contact                        |
| --------------------- | ---------------------- | ------------------------------ |
| PagerDuty             | Page on-call engineers | Auto-routes to current on-call |
| Slack #sre-escalation | Rapid coordination     | @sre-team                      |
| Phone Tree            | SEV-1 emergencies      | See emergency contacts         |
| Email                 | Non-urgent escalation  | sre@virtengine.com             |

#### Escalation Template

```
ESCALATION: [Severity] - [Brief Description]

Incident: #incident-2026-01-30-1423
Current IC: [Your Name]
Duration: [Time since incident start]

Current Status:
- [What is happening]
- [What has been tried]
- [Why escalation is needed]

Request:
- [Specific help needed]
- [Expertise required]

Incident Channel: #incident-2026-01-30-1423
```

---

### Exercise 2.1: Severity Classification

**Classify the following scenarios:**

1. The VirtEngine API is returning 503 errors for all requests
2. Provider deployments in us-east-1 are failing, us-west-2 is fine
3. A typo was found in the status page text
4. Block production has stopped for 2 minutes
5. 8% of VEID verifications are timing out
6. A validator reported 10 missed blocks in the last hour

**Answers:**

1. SEV-1 (complete outage)
2. SEV-2 (major feature, partial outage)
3. SEV-4 (cosmetic)
4. SEV-1 (chain halt)
5. SEV-3 (< 10% impact, within SLO)
6. SEV-3 (single validator, no chain impact)

---

## Session 3: Incident Commander Role (1.5 hours)

### What is an Incident Commander?

The **Incident Commander (IC)** is the single point of coordination during a major incident. The IC leads the response effort, makes decisions, and ensures effective communication.

### IC Responsibilities

| Responsibility      | Description                                       |
| ------------------- | ------------------------------------------------- |
| **Coordination**    | Lead and organize all response activities         |
| **Decision Making** | Make high-level decisions about response strategy |
| **Delegation**      | Assign tasks to responders                        |
| **Communication**   | Maintain communication with all stakeholders      |
| **Documentation**   | Ensure timeline is being captured                 |
| **Resolution**      | Declare when incident is resolved                 |

### IC Authority

During an active incident, the IC has authority to:

- Pull any engineer from other work
- Make emergency changes without normal approval process
- Escalate to executive team
- Approve emergency spending (within limits)
- Make final decisions during incident

### When to Have an IC

| Severity | IC Required? | Who Acts as IC?                |
| -------- | ------------ | ------------------------------ |
| SEV-1    | Yes          | Senior SRE or Engineering Lead |
| SEV-2    | Yes          | Senior SRE or Engineering Lead |
| SEV-3    | Optional     | On-call engineer               |
| SEV-4    | No           | N/A                            |

---

### IC Selection

**Who Can Be an IC:**

- Senior SRE engineers (SRE-III and above)
- Engineering leads
- Trained IC rotation members

**IC Prerequisites:**

- [ ] 6+ months as on-call engineer
- [ ] Responded to 5+ incidents
- [ ] Deep technical knowledge of VirtEngine
- [ ] Strong communication skills
- [ ] Completed IC training (this course)
- [ ] Passed IC certification exercise

**IC Rotation:**

- ICs are scheduled weekly in PagerDuty
- If scheduled IC is unavailable, on-call declares incident and pages IC rotation
- IC can hand off to another IC if incident extends beyond shift

---

### IC Workflow

#### 1. Taking Command

When you become IC:

```
1. Announce in incident channel:
   "[Name] is now Incident Commander for this incident"

2. Pin key information:
   - Incident severity
   - Current status
   - Link to status page
   - Link to dashboard

3. Assess current state:
   - What do we know?
   - What are we doing?
   - Who is working on what?

4. Assign roles:
   - Scribe (documents timeline)
   - Communications Lead (status updates)
   - Technical Lead (investigation)
```

#### 2. Running the Incident

**Every 10-15 Minutes (SEV-1) or 15-30 Minutes (SEV-2):**

1. **Status Check:** Get updates from all responders
2. **Assess Progress:** Are we making progress toward resolution?
3. **Adjust Strategy:** Change approach if not working
4. **Communicate:** Post update to incident channel
5. **Update Stakeholders:** Status page, leadership

**IC Update Template:**

```
[HH:MM] IC Update:

Status: [Investigating | Identified | Mitigating | Monitoring | Resolved]

What we know:
- [Key finding 1]
- [Key finding 2]

Current actions:
- @alice: [What Alice is doing]
- @bob: [What Bob is doing]

Next steps:
- [Planned action 1]
- [Planned action 2]

ETA: [Estimate or "Unknown"]
```

#### 3. Decision Making

**Decision Framework:**

```
┌─────────────────────────────────────────────────┐
│            Can we roll back?                     │
│               /        \                         │
│             Yes         No                       │
│              │           │                       │
│           ROLLBACK    Can we change config?      │
│                         /        \               │
│                       Yes         No             │
│                        │           │             │
│                   CONFIG        Can we reduce    │
│                   CHANGE        traffic?         │
│                                  /     \         │
│                                Yes      No       │
│                                 │        │       │
│                            RATE      HOTFIX or   │
│                            LIMIT     FAILOVER    │
└─────────────────────────────────────────────────┘
```

**Key Decision Points:**

- Rollback vs. forward fix
- Single vs. multi-strategy mitigation
- When to escalate to leadership
- When to invoke disaster recovery
- When to declare resolution

#### 4. Handing Off

If you need to hand off IC to another person:

```
1. Announce handoff:
   "IC handoff from [old IC] to [new IC]"

2. Brief new IC:
   - Current status
   - What's been tried
   - Current working hypothesis
   - Who is doing what
   - Outstanding decisions

3. Confirm handoff:
   "[New IC] confirms IC role"

4. Old IC remains available for questions
```

---

### IC Checklist

**At Incident Start:**

- [ ] Announce IC role
- [ ] Confirm severity
- [ ] Assign Scribe and Communications Lead
- [ ] Confirm status page is updated
- [ ] Notify stakeholders per escalation matrix

**During Incident:**

- [ ] Regular status updates (every 10-15 min for SEV-1/SEV-2)
- [ ] Track all actions in timeline
- [ ] Ensure responders have clear tasks
- [ ] Monitor for escalation needs
- [ ] Manage resource allocation

**At Resolution:**

- [ ] Confirm all health metrics are normal
- [ ] Monitor for 15-30 minutes stability
- [ ] Announce resolution
- [ ] Update status page to "Resolved"
- [ ] Schedule postmortem
- [ ] Thank responders

---

### IC Anti-Patterns

**What NOT to Do as IC:**

| Anti-Pattern            | Why It's Bad                | What to Do Instead              |
| ----------------------- | --------------------------- | ------------------------------- |
| Debugging code yourself | You lose coordination focus | Delegate technical work         |
| Not communicating       | Stakeholders panic          | Update every 10-15 minutes      |
| Making decisions alone  | You might miss something    | Consult SMEs, then decide       |
| Not documenting         | Postmortem is impossible    | Assign Scribe, track everything |
| Not escalating          | Problem gets worse          | Escalate early, escalate often  |
| Micromanaging           | Slows responders down       | Trust your team, check progress |

---

### Exercise 3.1: IC Role Play

**Scenario:**
You are the IC for a SEV-2 incident. The VEID verification service is returning errors for approximately 15% of requests. You have two engineers available: Alice (VEID expert) and Bob (infrastructure specialist).

**Tasks:**

1. Write your initial announcement taking IC role
2. Assign tasks to Alice and Bob
3. Write a 15-minute status update
4. Decide: The error rate is now 25% and increasing. What do you do?

---

## Session 4: Communication Protocols (1 hour)

### Communication Principles

**Effective incident communication is:**

- **Clear:** No jargon, no ambiguity
- **Concise:** Get to the point quickly
- **Accurate:** Verify before sharing
- **Timely:** Regular updates, not radio silence
- **Actionable:** What should people do?

---

### Internal Communication

#### Incident Slack Channel

**Naming Convention:** `#incident-YYYY-MM-DD-HHMM`

**Example:** `#incident-2026-01-30-1423`

**Pinned Information:**

```
📌 INCIDENT INFORMATION

Severity: SEV-2
Service: VEID Verification
Impact: 15% of verification requests failing
Incident Commander: @alice
Scribe: @bob
Status Page: https://status.virtengine.com
Dashboard: https://grafana.virtengine.com/d/veid-health

Started: 2026-01-30 14:23 UTC
```

**Update Cadence:**

| Severity | Update Frequency | Stakeholder Update |
| -------- | ---------------- | ------------------ |
| SEV-1    | Every 10 minutes | Every 30 minutes   |
| SEV-2    | Every 15 minutes | Every hour         |
| SEV-3    | At milestones    | At resolution      |
| SEV-4    | At resolution    | N/A                |

---

#### Internal Update Template

```
[HH:MM] [STATUS] - [Brief Description]

Impact: [Who is affected, how]
Current Status: [What's happening now]
Actions: [What we're doing about it]
Next Update: [When to expect next update]
```

**Example:**

```
[14:45] INVESTIGATING - VEID Verification Errors

Impact: ~15% of verification requests failing for all users
Current Status: Investigating ML inference service logs
Actions:
- @alice checking inference pod health
- @bob reviewing recent deployments
Next Update: 15:00 UTC
```

---

#### Incident Declaration Template

```
🚨 INCIDENT DECLARED 🚨

Severity: [SEV-1/SEV-2/SEV-3]
Service: [Affected Service]
Impact: [User-facing impact description]

Incident Channel: #incident-YYYY-MM-DD-HHMM
Incident Commander: @[IC Name]
Status Page: https://status.virtengine.com

Join the incident channel if you can help.
```

---

### External Communication

#### Status Page Updates

**URL:** https://status.virtengine.com

**Status States:**

| State         | When to Use                         |
| ------------- | ----------------------------------- |
| Investigating | Aware of issue, determining cause   |
| Identified    | Root cause known, working on fix    |
| Monitoring    | Fix applied, watching for stability |
| Resolved      | Issue fully resolved                |

**Status Page Templates:**

**Investigating:**

```
We are currently investigating issues with [Service Name].
Users may experience [specific impact].
We will provide updates as more information becomes available.
```

**Identified:**

```
We have identified the cause of the issue affecting [Service Name].
[Brief description of cause - no sensitive details].
We are implementing a fix. ETA: [time estimate or "working to determine"].
```

**Monitoring:**

```
A fix has been implemented for [Service Name].
We are monitoring the results and expect service to be fully restored shortly.
```

**Resolved:**

```
The issue affecting [Service Name] has been resolved.
Duration: [X] minutes
Impact: [brief summary]
A detailed post-incident review will be published within 5 business days.

We apologize for any inconvenience.
```

---

#### Stakeholder Notification Matrix

| Stakeholder             | SEV-1            | SEV-2            | SEV-3          | SEV-4  |
| ----------------------- | ---------------- | ---------------- | -------------- | ------ |
| CTO                     | Immediately      | 30 min           | Daily summary  | N/A    |
| CEO                     | 15 min           | 1 hour           | Weekly summary | N/A    |
| Customer Support        | Immediately      | Immediately      | As needed      | N/A    |
| Engineering Teams       | Incident channel | Incident channel | Team channel   | Ticket |
| Customers (Status Page) | Immediately      | 15 min           | Optional       | N/A    |

---

### Communication Anti-Patterns

| Anti-Pattern                       | Impact                       | Correct Approach                     |
| ---------------------------------- | ---------------------------- | ------------------------------------ |
| "We're looking into it" repeatedly | Stakeholders lose confidence | Provide specific actions being taken |
| Technical jargon to customers      | Confusion                    | Use simple language                  |
| No updates for 30+ minutes         | Panic, assumptions           | Set and meet update schedule         |
| Promising resolution time          | Disappointment if missed     | Say "investigating" until confident  |
| Blaming individuals                | Damaged trust                | Use blameless language               |

---

### Exercise 4.1: Communication Practice

**Scenario:** VirtEngine marketplace orders are failing with a 30% error rate. You need to write:

1. An initial status page update
2. A Slack update 15 minutes into the incident
3. A resolution message after the fix is applied

---

## Session 5: Tools and Access (1 hour)

### Monitoring and Alerting

#### Prometheus

**URL:** https://prometheus.virtengine.com

**Common Queries:**

```promql
# API error rate
sum(rate(virtengine_api_requests_total{status=~"5.."}[5m])) /
sum(rate(virtengine_api_requests_total[5m]))

# Block production rate
rate(virtengine_chain_block_height[5m])

# VEID verification latency P95
histogram_quantile(0.95, rate(virtengine_veid_verification_duration_seconds_bucket[5m]))

# Provider deployment success rate
sum(rate(virtengine_provider_deployment_success_total[5m])) /
sum(rate(virtengine_provider_deployment_total[5m]))
```

---

#### Grafana

**URL:** https://grafana.virtengine.com

**Key Dashboards:**

| Dashboard         | URL                  | Use Case                             |
| ----------------- | -------------------- | ------------------------------------ |
| Chain Health      | /d/chain-health      | Chain status, block time, validators |
| API Overview      | /d/api-overview      | Request rate, error rate, latency    |
| VEID Scoring      | /d/veid-scoring      | Verification rate, ML inference      |
| Marketplace       | /d/marketplace       | Orders, bids, leases                 |
| Provider Health   | /d/provider-health   | Deployment success, adapter status   |
| Error Budget      | /d/error-budget      | SLO status, budget remaining         |
| Incident Overview | /d/incident-overview | Key metrics during incidents         |

---

#### PagerDuty

**URL:** https://virtengine.pagerduty.com

**Key Actions:**

| Action            | How                                                    |
| ----------------- | ------------------------------------------------------ |
| Acknowledge alert | Click "Acknowledge" or reply to SMS/call               |
| Escalate          | Click "Escalate" → Select escalation policy            |
| Reassign          | Click "Reassign" → Select user                         |
| Add responders    | Click "Add Responders" → Select users/teams            |
| Snooze            | Click "Snooze" → Set duration                          |
| Resolve           | Click "Resolve" (only when incident is truly resolved) |

**Escalation Policies:**

- `Primary On-Call` - Main on-call rotation
- `Secondary On-Call` - Backup if primary doesn't respond
- `IC Rotation` - Incident Commander rotation
- `Management` - Engineering leadership

---

### Logging

#### Log Locations

| Log Type        | Location                             | Retention |
| --------------- | ------------------------------------ | --------- |
| API Logs        | Loki/Elasticsearch                   | 30 days   |
| Chain Node Logs | `/var/log/virtengine/virtengine.log` | 14 days   |
| VEID Logs       | `/var/log/virtengine/veid.log`       | 14 days   |
| Provider Daemon | `/var/log/virtengine/provider.log`   | 14 days   |

#### Common Log Queries

```bash
# Search for errors in chain logs
journalctl -u virtengined -p err -n 100 --no-pager

# Search for specific error code
grep "ERR_CONSENSUS" /var/log/virtengine/virtengine.log | tail -50

# Loki query for API errors
{service="virtengine-api"} |= "error" | json | line_format "{{.level}} {{.error}}"
```

---

### Incident Management

#### Slack Channels

| Channel           | Purpose                      |
| ----------------- | ---------------------------- |
| `#incident-*`     | Active incident coordination |
| `#sre-team`       | SRE team discussion          |
| `#sre-alerts`     | Alert notifications          |
| `#sre-escalation` | Escalation requests          |
| `#oncall-handoff` | Shift handoff notes          |

#### Incident Ticket Template

```
Title: INCIDENT-YYYY-MM-DD: [Brief Description]

Severity: [SEV-1/SEV-2/SEV-3]
Duration: [Start time] - [End time]
Incident Commander: [Name]

Summary:
[Brief description of what happened]

Timeline:
[Key events with timestamps]

Root Cause:
[What caused the incident]

Resolution:
[How it was resolved]

Action Items:
- [ ] [Action 1] - Owner: [Name], Due: [Date]
- [ ] [Action 2] - Owner: [Name], Due: [Date]

Postmortem: [Link to postmortem document]
```

---

### Access Verification Checklist

Before going on-call, verify you have access to:

- [ ] PagerDuty (can receive and acknowledge pages)
- [ ] Prometheus (can query metrics)
- [ ] Grafana (can view dashboards)
- [ ] Slack (in #sre-alerts, #sre-team channels)
- [ ] SSH access to production nodes
- [ ] Kubernetes cluster access
- [ ] VPN (if required)
- [ ] Status page (can update)
- [ ] Jira/ticket system (can create tickets)
- [ ] Chain node access (virtengined CLI)

---

### Exercise 5.1: Tool Navigation

Using the staging environment:

1. Find the current API error rate in Prometheus
2. Open the Chain Health dashboard in Grafana
3. Acknowledge a test alert in PagerDuty
4. Create a test incident channel in Slack
5. Verify you can SSH to a staging node
6. Run `virtengined status` and interpret the output

---

## Session 6: Role-Playing Exercises (1.5 hours)

### Exercise 6.1: Tabletop - API Outage

**Scenario:**
At 3:15 PM, you receive a PagerDuty alert: `APIErrorRateHigh - Error rate at 45%`

Walk through the response:

1. What is your first action?
2. What severity would you assign?
3. Who would you page?
4. What are your first diagnostic steps?

**Inject 1:** All recent deployments were 3+ hours ago
**Inject 2:** Database connections look healthy
**Inject 3:** Logs show `connection refused` to ML inference service

Continue the exercise with each inject.

---

### Exercise 6.2: Role Play - IC Practice

**Setup:**

- 1 person as Incident Commander
- 2 people as responders
- 1 person as facilitator (injects complications)
- 1 person as observer (provides feedback)

**Scenario:**
Block production has slowed to 1 block per 45 seconds (normally 6 seconds). Validator count appears normal.

**Practice:**

- IC takes command
- IC assigns tasks
- IC manages 15-minute updates
- IC makes decisions on mitigation
- Facilitator injects: "A major validator just went offline"

---

### Exercise 6.3: Mock Incident

**Full simulation using staging environment:**

1. Facilitator triggers failure in staging
2. Team receives real alert
3. Full incident response follows
4. Use actual tools (Slack, Grafana, etc.)
5. Debrief after exercise

**Evaluation Criteria:**

- Time to acknowledge alert: < 2 minutes
- Time to create incident channel: < 5 minutes
- Time to assign IC: < 10 minutes
- Regular updates posted: Every 15 minutes
- Resolution declared appropriately

---

## Session 7: Assessment & Certification (30 minutes)

### Knowledge Assessment

Complete the written assessment covering:

1. Severity classification (10 questions)
2. Escalation procedures (5 questions)
3. IC responsibilities (5 questions)
4. Communication protocols (5 questions)
5. Tool usage (5 questions)

**Passing Score:** 80%

---

### Practical Assessment

Demonstrate competency by:

1. **Shadow Shift:** Complete 2 on-call shadow shifts
2. **Mock Incident:** Successfully respond to a mock incident as participant
3. **IC Exercise:** Successfully lead a tabletop exercise as IC

---

### Certification

Upon successful completion:

- Added to on-call rotation pool
- Added to `@oncall-certified` Slack group
- Issued Incident Response Certification (valid 1 year)
- Eligible for IC training after 6 months

---

### Recertification

**Annual Requirements:**

- Respond to 5+ real incidents
- Attend quarterly gameday exercises
- Complete refresher quiz
- Review updated procedures

---

## Appendices

### Appendix A: Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│                 INCIDENT RESPONSE QUICK REFERENCE           │
├─────────────────────────────────────────────────────────────┤
│ SEVERITY LEVELS                                             │
│   SEV-1: Complete outage, chain halt      → 5 min response  │
│   SEV-2: Major degradation, SLO violated  → 15 min response │
│   SEV-3: Minor impact, no SLO impact      → 1 hour response │
│   SEV-4: Cosmetic, minimal impact         → 24 hour response│
├─────────────────────────────────────────────────────────────┤
│ FIRST ACTIONS                                               │
│   1. Acknowledge alert                                      │
│   2. Determine severity                                     │
│   3. Create incident channel (#incident-YYYY-MM-DD-HHMM)    │
│   4. Assign IC (SEV-1/SEV-2)                               │
│   5. Update status page (SEV-1/SEV-2)                      │
├─────────────────────────────────────────────────────────────┤
│ KEY URLS                                                    │
│   Grafana:    https://grafana.virtengine.com               │
│   Prometheus: https://prometheus.virtengine.com            │
│   PagerDuty:  https://virtengine.pagerduty.com             │
│   Status:     https://status.virtengine.com                │
├─────────────────────────────────────────────────────────────┤
│ ESCALATION                                                  │
│   Primary On-Call: PagerDuty auto-routes                   │
│   IC Rotation:     PagerDuty "IC Rotation" policy          │
│   Management:      Slack #sre-escalation                   │
│   Security:        security@virtengine.com (24/7)          │
└─────────────────────────────────────────────────────────────┘
```

---

### Appendix B: Related Runbooks

For specific incident types, refer to:

| Alert/Incident         | Runbook                                                                                                           |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Node Down              | [docs/operations/runbooks/node-down.md](../../../docs/operations/runbooks/node-down.md)                           |
| High Error Rate        | [docs/operations/runbooks/high-error-rate.md](../../../docs/operations/runbooks/high-error-rate.md)               |
| Block Stalled          | [docs/operations/runbooks/block-stalled.md](../../../docs/operations/runbooks/block-stalled.md)                   |
| Low Validators         | [docs/operations/runbooks/low-validators.md](../../../docs/operations/runbooks/low-validators.md)                 |
| VEID Non-Deterministic | [docs/operations/runbooks/veid-non-deterministic.md](../../../docs/operations/runbooks/veid-non-deterministic.md) |
| Provider Deployment    | [docs/operations/runbooks/provider-deployment.md](../../../docs/operations/runbooks/provider-deployment.md)       |
| SLO Budget Burning     | [docs/operations/runbooks/slo-budget-burning.md](../../../docs/operations/runbooks/slo-budget-burning.md)         |

---

### Appendix C: Incident Response Checklist

**Print and keep handy during on-call shifts**

**DETECTION**

- [ ] Alert acknowledged within 2 minutes
- [ ] Initial assessment complete

**TRIAGE**

- [ ] Severity determined
- [ ] Incident channel created
- [ ] IC assigned (SEV-1/SEV-2)
- [ ] Status page updated

**RESPONSE**

- [ ] Investigation underway
- [ ] Hypothesis formed
- [ ] Updates posted every 10-15 minutes
- [ ] Escalation if needed

**RESOLUTION**

- [ ] Fix applied
- [ ] Health verified
- [ ] Stability monitored (15-30 min)
- [ ] Resolution announced
- [ ] Status page updated

**POST-INCIDENT**

- [ ] Incident ticket created
- [ ] Postmortem scheduled
- [ ] Timeline documented
- [ ] Action items created

---

## References

- [Incident Response Process](../../../docs/sre/INCIDENT_RESPONSE.md)
- [Incident Drills](../../../docs/sre/INCIDENT_DRILLS.md)
- [Runbook Procedures Training](./runbook-procedures.md)
- [Post-Incident Analysis Training](./post-incident-analysis.md)
- [Google SRE Book - Managing Incidents](https://sre.google/sre-book/managing-incidents/)
- [PagerDuty Incident Response Guide](https://response.pagerduty.com/)

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
