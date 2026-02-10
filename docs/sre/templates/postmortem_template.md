# Postmortem: [INCIDENT-ID] [Brief Description]

**Date**: YYYY-MM-DD
**Authors**: [Names]
**Status**: Draft | Under Review | Final
**Severity**: SEV-1 | SEV-2 | SEV-3 | SEV-4

---

## Executive Summary

[2-3 sentence summary of what happened, impact, and resolution]

**Example**:
> On 2026-01-29 at 14:23 UTC, the VirtEngine API service experienced a complete outage lasting 47 minutes due to a database connection pool exhaustion. During this time, all API requests returned 500 errors, affecting approximately 1,200 users and consuming 78% of our monthly error budget. The issue was resolved by restarting the API service and increasing the database connection pool size.

---

## Impact

### User Impact
- **Affected Users**: [Number or percentage]
- **Geographic Impact**: [Regions affected]
- **Duration**: [Total downtime]
- **Degraded Period**: [Time of degraded service, if applicable]

### Business Impact
- **Revenue Impact**: [If applicable]
- **Reputation Impact**: [User complaints, social media, press coverage]
- **SLO Impact**: [Which SLOs violated, error budget consumed]

### Example
```
User Impact:
- ~1,200 active users affected (15% of daily active users)
- Global impact (all regions)
- Complete outage: 47 minutes (14:23-15:10 UTC)
- Degraded performance: Additional 20 minutes (slow responses)

Business Impact:
- Estimated revenue impact: $2,500 in failed transactions
- 18 user complaints via support tickets
- 2 tweets mentioning the outage
- SLO Impact: API Availability SLO violated (99.90% → 99.73%)
- Error Budget: 78% consumed (31.2 of 40.32 minutes monthly budget)
```

---

## Timeline (all times UTC)

| Time | Event |
|------|-------|
| 14:15 | [Event leading to incident] |
| 14:23 | [Incident begins - first symptoms] |
| 14:24 | [Automated alert fires] |
| 14:26 | [On-call engineer paged] |
| 14:28 | [Engineer acknowledges page] |
| 14:35 | [Investigation begins] |
| 14:42 | [Root cause identified] |
| 14:50 | [Fix applied] |
| 15:10 | [Service fully restored] |
| 15:30 | [Incident closed, monitoring continues] |

### Detailed Timeline Example

```
14:15 - Deployment of API v1.2.3 completes successfully
14:20 - Traffic starts increasing due to scheduled marketing email
14:23 - Database connection pool reaches maximum (100 connections)
14:23 - API requests start failing with "cannot acquire connection" errors
14:24 - Alert: "APIErrorRateHigh" fires (error rate: 45%)
14:24 - Alert: "APIAvailabilityLow" fires (availability: 55%)
14:26 - On-call engineer (Alice) paged via PagerDuty
14:28 - Alice acknowledges page, begins investigation
14:30 - Alice checks recent deployments, notes v1.2.3 deployed 15 min ago
14:32 - Alice reviews metrics, sees connection pool exhaustion
14:35 - Alice joins #incident-response Slack channel
14:35 - Alice declares SEV-2 incident
14:37 - Incident Commander (Bob) joins, takes command
14:38 - Bob requests rollback to v1.2.2
14:42 - Rollback initiated but fails (database migration not reversible)
14:45 - Bob pivots to mitigation: increase connection pool
14:47 - Configuration change prepared (pool: 100 → 500)
14:50 - Configuration deployed, API restarting
14:52 - API health checks pass
14:55 - Error rate dropping (45% → 5%)
15:00 - Error rate back to normal (<0.5%)
15:05 - Monitoring confirms stability
15:10 - Incident declared resolved
15:30 - Incident closed, postmortem scheduled
```

---

## Root Cause

### What Happened

[Detailed technical explanation of the root cause]

**Template**:
1. What component failed?
2. Why did it fail?
3. What was the underlying cause?
4. Why didn't existing safeguards catch this?

### Example

```
Root Cause: Database Connection Pool Exhaustion

1. The API service's database connection pool was configured with a maximum of 100 connections.

2. Deployment v1.2.3 introduced a new feature that queries the database 3x more frequently per request (user profile enrichment).

3. When traffic spiked from 500 RPS to 800 RPS (due to marketing email), the connection pool exhausted:
   - 800 requests/sec × 3 queries/request = 2,400 queries/sec
   - Each query holds a connection for ~40ms on average
   - Peak connections needed: 2,400 × 0.04 = 96 connections
   - Pool size: 100 connections
   - No headroom for variance

4. The pool exhausted during normal traffic variance (800 RPS → 850 RPS spike).

5. Existing safeguards failed:
   - Load testing only tested up to 600 RPS (didn't catch the issue)
   - Connection pool monitoring existed but alert threshold was too high (> 95%)
   - Performance review didn't catch the 3x query increase
```

### Contributing Factors

[What made this worse or prevented earlier detection?]

**Example**:
```
1. Marketing email was unannounced (no capacity planning)
2. Load testing only tested up to 75% of production capacity
3. Connection pool monitoring alert threshold too high (95% vs 80%)
4. No query count tracking in performance budgets
5. Rollback failed due to irreversible database migration
6. On-call engineer unfamiliar with this specific component
```

---

## Detection

### How was the incident detected?

- [ ] Automated monitoring alert
- [ ] User report
- [ ] Internal discovery
- [ ] External notification

### Detection Quality

- **Time to Detect (TTD)**: [Time from incident start to first alert]
- **Time to Acknowledge (TTA)**: [Time from alert to human acknowledgement]
- **Detection Method**: [What alerted/how was it found?]

### Example

```
Detection Method: Automated monitoring alert

Timeline:
- 14:23 - Incident began (error rate spike)
- 14:24 - Alert fired (1 minute TTD) ✅
- 14:26 - On-call paged (2 minutes TTA)
- 14:28 - Page acknowledged (4 minutes total)

Quality: Good
- Alert fired quickly (1 min TTD)
- Alert was accurate (true positive)
- Page reached on-call successfully

Improvement Opportunity:
- Alert threshold could be more sensitive (10% errors vs current 20%)
- Predictive alerting could have warned before total failure
```

---

## Response

### Responders

| Role | Name | Joined At |
|------|------|-----------|
| On-Call Engineer | Alice Smith | 14:28 |
| Incident Commander | Bob Jones | 14:37 |
| SME - Database | Carol White | 14:45 |
| Communications Lead | Dave Brown | 14:40 |

### Response Actions

**What was done to mitigate/resolve?**

1. [Action 1]
2. [Action 2]
3. [Action 3]

### Example

```
Response Actions:

1. Investigation (14:28-14:42)
   - Reviewed recent deployments
   - Analyzed metrics and logs
   - Identified connection pool exhaustion

2. Attempted Rollback (14:38-14:42) - FAILED
   - Rollback command issued
   - Failed due to irreversible database migration
   - Pivoted to alternative mitigation

3. Connection Pool Increase (14:42-14:50)
   - Configuration change prepared (100 → 500 connections)
   - Change reviewed and approved
   - Deployed via Ansible playbook
   - API service restarted

4. Monitoring and Verification (14:50-15:10)
   - Health checks monitored
   - Error rate tracked
   - Connection pool utilization confirmed healthy
   - Service declared stable

5. Communication (14:40-15:15)
   - Status page updated (investigating → identified → resolved)
   - Internal Slack updates every 10 minutes
   - Support team notified
   - Post-resolution summary posted
```

### What Went Well

- ✅ Alerts fired quickly and accurately
- ✅ On-call responded promptly
- ✅ Incident command structure established quickly
- ✅ Root cause identified in < 20 minutes
- ✅ Alternative mitigation found when rollback failed
- ✅ Clear communication maintained throughout

### What Went Poorly

- ❌ Rollback failed (database migration not reversible)
- ❌ Load testing didn't catch the issue
- ❌ On-call engineer unfamiliar with component (learning curve)
- ❌ No automated mitigation available
- ❌ Marketing campaign not communicated to SRE team

---

## Resolution

### How was service restored?

[Describe the final fix that resolved the incident]

### Example

```
Resolution: Increased database connection pool size from 100 to 500 connections

Immediate Fix (Applied during incident):
- Configuration change: db.pool.max_connections = 500
- API service restart to apply configuration
- Service restored at 15:10 UTC

Temporary vs Permanent:
- ✅ Permanent fix (configuration change deployed to production)
- Configuration committed to version control
- Change propagated to all environments
```

### Verification

How did we know it was fixed?

**Example**:
```
Verification Steps:
1. Health checks passed (all API instances healthy)
2. Error rate returned to baseline (<0.5%)
3. Response latency returned to normal (P95 < 2s)
4. Connection pool utilization: 45% (healthy headroom)
5. No new errors in logs
6. Synthetic monitoring checks passing
7. User reports stopped
```

---

## Action Items

### Prevent Recurrence

| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| [Action to prevent this specific issue] | [Name] | YYYY-MM-DD | P0-P3 | Open/In Progress/Done |

### Example

| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| Implement query count tracking in performance budgets | Alice | 2026-02-05 | P0 | Open |
| Increase load testing to 150% of production capacity | Bob | 2026-02-10 | P0 | Open |
| Add connection pool predictive alerting (trend-based) | Alice | 2026-02-12 | P1 | Open |
| Document rollback procedures and limitations | Carol | 2026-02-08 | P1 | Open |
| Implement automated connection pool scaling | Alice | 2026-02-28 | P2 | Open |
| Establish SRE <> Marketing communication process | Dave | 2026-02-15 | P1 | Open |
| Create runbook for connection pool issues | Bob | 2026-02-07 | P2 | Open |

### Improve Detection

| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| Lower connection pool alert threshold (95% → 80%) | Alice | 2026-02-01 | P0 | Open |
| Add predictive alerting for connection pool trend | Alice | 2026-02-12 | P1 | Open |
| Implement query-per-request tracking | Bob | 2026-02-10 | P1 | Open |

### Improve Response

| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| Add automated connection pool scaling playbook | Alice | 2026-02-20 | P2 | Open |
| Cross-train on-call engineers on database components | Carol | 2026-02-28 | P2 | Open |
| Improve rollback testing (include DB migrations) | Bob | 2026-02-15 | P1 | Open |

### Process Improvements

| Action | Owner | Due Date | Priority | Status |
|--------|-------|----------|----------|--------|
| Require SRE review for marketing campaigns | Dave | 2026-02-05 | P0 | Open |
| Add "queries per request" to performance review checklist | Alice | 2026-02-03 | P1 | Open |
| Expand load testing coverage to include traffic spikes | Bob | 2026-02-10 | P0 | Open |

---

## Lessons Learned

### What did we learn?

**Technical Learnings**:
1. [Learning 1]
2. [Learning 2]

**Process Learnings**:
1. [Learning 1]
2. [Learning 2]

**Cultural Learnings**:
1. [Learning 1]

### Example

```
Technical Learnings:
1. Database connection pool sizing must account for queries-per-request, not just request rate
2. Load testing should test 150% of expected peak capacity
3. Irreversible database migrations prevent rollback; need alternative mitigation strategies
4. Connection pool exhaustion can happen within normal traffic variance when headroom is low

Process Learnings:
1. Marketing campaigns create traffic spikes; SRE must be notified in advance
2. Performance review process missed a 3x increase in database queries
3. Load testing coverage was insufficient (only 75% of production capacity)
4. Rollback procedures need to be tested regularly, especially with database migrations

Cultural Learnings:
1. Incident response was calm and effective due to regular gameday practice
2. Cross-team communication (Engineering, SRE, Support) worked well
3. Blameless culture allowed quick root cause identification without fear
```

---

## Supporting Information

### Metrics and Graphs

[Links to dashboards, graphs, screenshots]

**Example**:
```
- Error Rate Graph: https://grafana.example.com/d/api-errors?from=1738166580000&to=1738169400000
- Connection Pool Graph: https://grafana.example.com/d/db-pool?from=1738166580000&to=1738169400000
- Latency Distribution: https://grafana.example.com/d/api-latency?from=1738166580000&to=1738169400000
- Alert History: https://pagerduty.com/incidents/INCIDENT-12345
```

### Logs and Traces

[Links to relevant log queries or trace IDs]

**Example**:
```
- Error Logs: https://kibana.example.com/app/discover#/?_g=(time:(from:'2026-01-29T14:20:00Z',to:'2026-01-29T15:30:00Z'))
- Trace Example (Failed Request): https://jaeger.example.com/trace/abc123def456
- Database Query Logs: /var/log/postgresql/postgres-2026-01-29.log lines 5432-8901
```

### References

[Links to runbooks, documentation, related incidents]

**Example**:
```
- Runbook: Connection Pool Exhaustion: docs/runbooks/db-connection-pool.md
- Related Incident: INCIDENT-8901 (Database slowness, 2025-11-12)
- Performance Budget Documentation: docs/sre/PERFORMANCE_BUDGETS.md
- Load Testing Framework: tests/load/README.md
```

---

## Blameless Culture Reminder

> "Failure is inevitable in complex systems. Our goal is to learn from failure, not to assign blame. Assume good faith, focus on systemic improvements, and use this as an opportunity to make VirtEngine more reliable."

**Principles**:
- ✅ **Do** focus on what happened and why
- ✅ **Do** identify system weaknesses
- ✅ **Do** propose concrete improvements
- ❌ **Don't** assign blame to individuals
- ❌ **Don't** use language like "human error" (ask "why was this error possible?")
- ❌ **Don't** focus on "who" but rather "what" and "why"

---

## Approvals

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Author | [Name] | YYYY-MM-DD | |
| Reviewer (SRE Lead) | [Name] | YYYY-MM-DD | |
| Reviewer (Eng Lead) | [Name] | YYYY-MM-DD | |
| Approver (VP Eng) | [Name] | YYYY-MM-DD | |

---

## Appendix

### Incident Severity Definitions

**SEV-1 (Critical)**:
- Complete service outage
- Major functionality broken for all users
- Data loss or security breach
- Revenue impact > $10,000/hour

**SEV-2 (High)**:
- Significant service degradation
- Major functionality broken for subset of users
- SLO violation
- Revenue impact $1,000-$10,000/hour

**SEV-3 (Medium)**:
- Minor service degradation
- Non-critical functionality broken
- No SLO violation
- Revenue impact < $1,000/hour

**SEV-4 (Low)**:
- Minimal impact
- Cosmetic issues
- No user impact
- No revenue impact

---

**Document Version**: 1.0
**Last Updated**: YYYY-MM-DD
**Next Review**: [Schedule review of action items in 30 days]
