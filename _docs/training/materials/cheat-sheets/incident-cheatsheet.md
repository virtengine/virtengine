# Incident Response Cheat Sheet

**Quick Reference for Incident Responders**

---

## Severity Levels

| Level | Impact | Response Time | Examples |
|-------|--------|---------------|----------|
| **SEV-1** | Complete outage, data risk | Immediate | Chain halt, key compromise |
| **SEV-2** | Major feature unavailable | < 15 min | VEID failed, consensus issues |
| **SEV-3** | Degraded performance | < 1 hour | High latency, partial failures |
| **SEV-4** | Minor issues | Next business day | UI bugs, non-critical errors |

---

## Incident Lifecycle

```
Detection → Triage → Response → Resolution → Postmortem
    │          │         │           │            │
   Alert    Severity   Mitigate    Fix root    Document
   fires    assess     impact      cause       learning
```

---

## First 15 Minutes

### 1. Acknowledge (2 min)
```
✓ Acknowledge alert in PagerDuty
✓ Join #incident-active Slack
✓ Announce: "Investigating [alert name]"
```

### 2. Assess Severity (5 min)
```
Questions:
• Is the service completely down?
• What percentage of users affected?
• Is data at risk?
• Is it security-related?
```

### 3. Assemble Team (3 min)
```
SEV-1/2: Page additional responders
SEV-3/4: Continue solo, escalate if needed
```

### 4. Initial Communication (5 min)
```
Post status update:
• What we know
• Current impact
• Actions being taken
• Next update time
```

---

## Roles

| Role | Responsibilities |
|------|------------------|
| **IC** | Coordinate, decide, communicate |
| **Tech Lead** | Debug, fix, verify |
| **Comms** | Status updates, stakeholders |
| **Scribe** | Document timeline |

---

## Status Update Template

```markdown
**Incident**: [Brief description]
**Severity**: SEV-[1/2/3/4]
**Status**: [Investigating/Identified/Monitoring/Resolved]
**Time**: [UTC]

### Impact
[What users are experiencing]

### Current Actions
[What we're doing]

### Next Update
[Time of next update]
```

---

## Communication Schedule

| Severity | First Update | Ongoing |
|----------|--------------|---------|
| SEV-1 | 15 min | Every 30 min |
| SEV-2 | 30 min | Every 1 hour |
| SEV-3 | 1 hour | Every 2 hours |
| SEV-4 | 4 hours | As needed |

---

## Quick Diagnostics

### Node Issues
```bash
# Status
virtengine status | jq '.SyncInfo'

# Logs
journalctl -u virtengine -n 100 --no-pager

# Resources
ps aux | grep virtengine
df -h
free -m
```

### Provider Issues
```bash
# Daemon status
systemctl status provider-daemon

# API health
curl -s localhost:8443/api/v1/status

# Workloads
curl -s localhost:8443/api/v1/workloads | jq '.[] | {id, state}'
```

### Metrics Check
```bash
# Prometheus queries
curl -s 'localhost:9090/api/v1/query?query=up' | jq .

# Error rate
curl -s 'localhost:9090/api/v1/query?query=rate(errors_total[5m])' | jq .
```

---

## Runbook Quick Links

| Issue | Runbook |
|-------|---------|
| Node down | `docs/operations/runbooks/node-down.md` |
| Block stalled | `docs/operations/runbooks/block-stalled.md` |
| High error rate | `docs/operations/runbooks/high-error-rate.md` |
| VEID issues | `docs/operations/runbooks/veid-non-deterministic.md` |
| Provider deploy fail | `docs/operations/runbooks/provider-deployment.md` |
| SLO burning | `docs/operations/runbooks/slo-budget-burning.md` |

---

## Escalation

### When to Escalate
- No progress after 30 min
- Need expertise you don't have
- Impact increasing
- Security concern

### How to Escalate
```
1. Page next tier in PagerDuty
2. Summarize situation clearly
3. State what you need
4. Hand off or support as needed
```

---

## Post-Incident

### Immediate
- [ ] Verify service restored
- [ ] Send resolution update
- [ ] Schedule postmortem (within 48h for SEV-1/2)

### Postmortem Must-Haves
- Timeline with sources
- Root cause analysis (5 Whys)
- Action items with owners
- What went well
- What to improve

---

## Emergency Contacts

| Role | Contact |
|------|---------|
| On-call primary | PagerDuty schedule |
| On-call secondary | PagerDuty schedule |
| Security | security@virtengine.com |
| Engineering Lead | [Slack DM] |

---

## Channels

| Channel | Purpose |
|---------|---------|
| #incident-active | Active incident coordination |
| #on-call | On-call questions |
| #engineering | Broader updates |

---

*Print this. Tape it to your monitor. You'll thank yourself at 3 AM.*
