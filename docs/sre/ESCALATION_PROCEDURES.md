# Escalation Procedures

## Overview

This document defines the escalation procedures for VirtEngine incidents. Proper escalation ensures the right people are engaged at the right time to minimize impact and accelerate resolution.

## Table of Contents

1. [Escalation Matrix](#escalation-matrix)
2. [Escalation Decision Trees](#escalation-decision-trees)
3. [Contact Information](#contact-information)
4. [Escalation Paths](#escalation-paths)
5. [Special Escalations](#special-escalations)
6. [Communication Protocols](#communication-protocols)

---

## Escalation Matrix

### By Severity Level

| Severity | Initial Response | Escalation Tier 1 | Escalation Tier 2 | Escalation Tier 3 | Executive Notification |
|----------|------------------|-------------------|-------------------|-------------------|----------------------|
| **SEV-1** | On-Call Engineer (immediate) | Senior SRE + Incident Commander (5 min) | Engineering Director + All SMEs (15 min) | VP Engineering + CTO (30 min) | CEO (immediate) |
| **SEV-2** | On-Call Engineer (15 min) | Incident Commander + Team Lead (30 min) | Engineering Director (1 hour) | VP Engineering (2 hours) | CTO (daily summary) |
| **SEV-3** | On-Call Engineer (1 hour) | Team Lead (business hours) | Engineering Director (if escalated) | - | Weekly summary |
| **SEV-4** | On-Call Engineer (24 hours) | Team Lead (if needed) | - | - | Monthly summary |

### By Domain

| Domain | Primary Contact | Secondary Contact | SME | Management |
|--------|----------------|-------------------|-----|------------|
| **Blockchain Core** | SRE On-Call | Senior Blockchain Engineer | Blockchain Architect | VP Engineering |
| **VEID/Identity** | Identity Team On-Call | Identity Tech Lead | ML Engineer | Identity Product Lead |
| **Provider Daemon** | Infrastructure On-Call | Provider Team Lead | Infrastructure Architect | VP Engineering |
| **Database/Storage** | Database SRE | Database Administrator | Senior SRE | Infrastructure Director |
| **Network** | Network SRE | Senior Network Engineer | Network Architect | Infrastructure Director |
| **Security** | Security On-Call | CISO | Security Engineer | CTO |
| **ML/Inference** | ML Engineer On-Call | ML Team Lead | Senior ML Engineer | VP Engineering |

---

## Escalation Decision Trees

### When to Escalate

```
┌─────────────────────────────────────┐
│  Incident Detected                  │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Can I resolve this quickly?        │
│  (< 15 minutes for SEV-1/2)        │
└──────────────┬──────────────────────┘
               │
         ┌─────┴─────┐
         │           │
        YES          NO
         │           │
         ▼           ▼
    ┌────────┐  ┌────────────────────┐
    │Continue│  │ ESCALATE           │
    │Working │  │ to Tier 1          │
    └────────┘  └─────────┬──────────┘
                          │
                          ▼
                ┌────────────────────────┐
                │ Tier 1 can resolve?    │
                └─────────┬──────────────┘
                          │
                    ┌─────┴─────┐
                    │           │
                   YES          NO
                    │           │
                    ▼           ▼
                ┌────────┐  ┌───────────┐
                │Resolve │  │ ESCALATE  │
                │        │  │ to Tier 2 │
                └────────┘  └───────────┘
```

### Escalation Triggers

**Automatic Escalation** (no decision needed):

1. **SEV-1 declared** → Immediately page:
   - Senior SRE on-call
   - Incident Commander
   - Service owner
   - Engineering Director

2. **Incident duration > 30 minutes** (SEV-1/SEV-2) → Escalate to:
   - Engineering Director
   - VP Engineering (if SEV-1)

3. **SLO budget depleted** → Notify:
   - Engineering Director
   - VP Engineering
   - Product Lead

4. **Security incident** → Immediately notify:
   - CISO
   - Security Team
   - Legal (for data breach)
   - CTO

5. **Data loss or corruption** → Immediately notify:
   - Database Team Lead
   - Engineering Director
   - VP Engineering
   - Legal

**Manual Escalation** (responder decides):

- Need specialized expertise not available
- Incident complexity exceeds current responder capability
- External dependencies blocking resolution
- Need executive decision or authorization
- Multi-team coordination required
- Incident scope expanding

---

## Contact Information

### On-Call Rotations

**Access PagerDuty**: https://virtengine.pagerduty.com

**Schedules**:
- `Primary SRE On-Call` - First responder
- `Secondary SRE On-Call` - Backup (auto-escalation after 5 min)
- `Incident Commander Rotation` - Senior SREs
- `Security On-Call` - Security team
- `Database On-Call` - Database specialists

### Key Personnel

| Role | Name | Phone | Email | PagerDuty | Slack |
|------|------|-------|-------|-----------|-------|
| **VP Engineering** | [Name] | [Phone] | [Email] | @vp-eng | @vp-eng |
| **Engineering Director** | [Name] | [Phone] | [Email] | @eng-director | @eng-director |
| **CISO** | [Name] | [Phone] | [Email] | @ciso | @ciso |
| **Database Team Lead** | [Name] | [Phone] | [Email] | @db-lead | @db-lead |
| **Infrastructure Lead** | [Name] | [Phone] | [Email] | @infra-lead | @infra-lead |
| **Blockchain Architect** | [Name] | [Phone] | [Email] | @blockchain-arch | @blockchain-arch |

**Note**: Phone numbers and emails stored in internal directory (not in public docs)

### Executive Contacts

| Role | Contact Method | When to Contact |
|------|---------------|-----------------|
| **CTO** | Slack DM, Phone | SEV-1 incidents, security breaches, major outages |
| **CEO** | Via CTO | Complete service outage > 1 hour, security breach with data loss, press inquiry |
| **Legal** | Email + Phone | Security breach, data loss, compliance violation, subpoena |
| **PR/Communications** | Email + Phone | Press inquiry, public incident, social media attention |

---

## Escalation Paths

### SEV-1 Escalation Path

```
[00:00] Incident Detected
   ↓
[00:01] On-Call Engineer Paged
   ↓
[00:03] On-Call Acknowledges
   ↓
[00:05] Senior SRE + IC Paged (automatic)
   ↓
[00:10] Engineering Director Notified (automatic)
   ↓
[00:15] VP Engineering Notified (automatic)
   ↓
[00:30] CTO Notified (if not resolved)
   ↓
[01:00] CEO Notified (if not resolved)
```

### SEV-2 Escalation Path

```
[00:00] Incident Detected
   ↓
[00:15] On-Call Engineer Assesses
   ↓
[00:30] Team Lead Notified
   ↓
[01:00] Engineering Director Notified
   ↓
[02:00] VP Engineering Notified (if not resolved)
```

### Security Incident Escalation

```
[00:00] Security Event Detected
   ↓
[00:01] Security On-Call Paged (immediate)
   ↓
[00:05] CISO Notified (immediate)
   ↓
[00:10] Legal Notified (if data breach suspected)
   ↓
[00:15] CTO Notified
   ↓
[00:30] CEO Notified (if confirmed breach)
   ↓
[01:00] PR Team Notified (if public disclosure needed)
```

---

## Special Escalations

### Security Incidents

**Criteria**:
- Unauthorized access detected
- Data breach suspected or confirmed
- Vulnerability actively exploited
- Malware or ransomware detected
- DDoS attack
- Insider threat

**Immediate Actions**:
1. Page Security On-Call (PagerDuty: `security-oncall`)
2. Notify CISO (Slack: `@ciso`, Phone)
3. Create private incident channel: `#security-incident-[ID]`
4. Do NOT use public incident channels
5. Contain the threat (isolate affected systems)

**Escalation Chain**:
```
Security On-Call → CISO → CTO → CEO → Legal → PR (if needed)
```

**Legal Notification Required**:
- Confirmed data breach
- Customer PII exposed
- Compliance violation (GDPR, HIPAA, etc.)
- Law enforcement involvement

---

### Data Loss or Corruption

**Criteria**:
- Data deleted or modified unexpectedly
- Database corruption detected
- Backup failure during incident
- Irrecoverable data loss

**Immediate Actions**:
1. STOP all write operations if possible
2. Page Database On-Call + Database Team Lead
3. Notify Engineering Director immediately
4. Take database snapshots/backups
5. Isolate affected systems

**Escalation Chain**:
```
On-Call → Database Team Lead → Engineering Director → VP Engineering → CTO
```

**Additional Notifications**:
- Legal (if customer data affected)
- Product Team (for user communication)
- Customer Success (for impacted customers)

---

### External Dependencies

**Criteria**:
- AWS/Cloud provider outage
- Third-party API outage
- DNS/CDN failure
- Payment processor issue

**Immediate Actions**:
1. Confirm external issue (check status pages)
2. Notify customers via status page
3. Activate failover/backup systems if available
4. Contact vendor support
5. Monitor vendor status pages

**Escalation Chain**:
```
On-Call → Infrastructure Lead → Engineering Director → VP Engineering
```

**Communication**:
- Update status page with external attribution
- Prepare customer communication
- Engage vendor account team if needed

---

### Multi-Team Coordination

**Criteria**:
- Issue spans multiple services/teams
- Requires coordination across Engineering, Product, Support
- Customer-facing with complex resolution

**Immediate Actions**:
1. Establish Incident Commander (IC)
2. Create cross-functional incident channel
3. Assign liaison from each team
4. IC coordinates all activities
5. Regular sync meetings (every 15-30 min)

**Escalation Chain**:
```
On-Call → Incident Commander → Team Leads → Engineering Director
```

**Roles**:
- **IC**: Overall coordination
- **Tech Lead**: Technical decisions
- **Comms Lead**: Customer communication
- **Support Liaison**: Handle customer inquiries
- **Product Liaison**: Business impact assessment

---

## Communication Protocols

### Escalation Notification Template

**Initial Escalation (Slack)**:

```
🚨 ESCALATION REQUEST 🚨

Incident: #incident-2026-01-30-1530
Severity: SEV-1
Current Responder: @alice-smith
Escalation To: @eng-director

Reason for Escalation:
- Incident duration > 30 minutes
- Need authorization for full database restore
- Potential data loss affecting 5,000+ users

Current Status:
- Database corruption detected in user_profiles table
- Backup restoration in progress
- ETA unknown, need executive decision on rollback window

Action Requested:
- Approve 2-hour maintenance window
- Notify customers of extended downtime
- Authorize full database restore from 6-hour-old backup

Incident Channel: #incident-2026-01-30-1530
```

**Phone Escalation Script**:

```
"Hi [Name], this is [Your Name] from VirtEngine SRE.

We have a [SEVERITY] incident affecting [SERVICE].

Impact: [Brief description of user impact]

Duration: [How long it's been going on]

I'm escalating because: [Reason]

We need: [Specific ask or decision]

Incident channel: [Slack channel name]

Can you join the incident channel now?"
```

### Status Updates During Escalation

**To Escalated Personnel**:
- Update every 10 minutes for SEV-1
- Update every 15-30 minutes for SEV-2
- Include: what's happening, what we've tried, what we need

**To Executive Team**:
- Update every 30 minutes for SEV-1
- Update every hour for SEV-2
- Include: business impact, customer impact, ETA

**To Customers (via Status Page)**:
- Update every 15 minutes for SEV-1
- Update every 30 minutes for SEV-2
- Be honest but avoid technical jargon
- Provide ETA or "investigating"

---

## Escalation Metrics

### Key Metrics

**Time to Escalate (TTE)**:
```
TTE = Escalation Time - Incident Start Time

Target:
- SEV-1: < 5 minutes to Tier 1
- SEV-2: < 15 minutes to Tier 1
```

**Escalation Accuracy**:
```
Appropriate Escalations / Total Escalations

Target: > 90%
```

**Resolution After Escalation**:
```
Average Time to Resolution After Escalation

Target:
- SEV-1: < 30 minutes after escalation
- SEV-2: < 1 hour after escalation
```

### Escalation Anti-Patterns

**Over-Escalation**:
- ❌ Escalating before attempting basic troubleshooting
- ❌ Escalating SEV-3/4 incidents to executives
- ❌ Paging entire team for non-critical issues

**Under-Escalation**:
- ❌ Waiting too long to escalate (trying to be a hero)
- ❌ Not escalating when lacking expertise
- ❌ Not escalating security incidents immediately

**Proper Escalation**:
- ✅ Escalate when you're stuck or need expertise
- ✅ Escalate based on severity and duration
- ✅ Escalate proactively when impact is growing
- ✅ Escalate security incidents immediately

---

## Training and Drills

### Escalation Training

**All On-Call Engineers Must Complete**:
- [ ] Review escalation matrix
- [ ] Practice escalation phone calls (role-play)
- [ ] Shadow 2 escalations
- [ ] Understand when NOT to escalate
- [ ] Know all contact methods

**Annual Refresher**:
- Review updated escalation procedures
- Practice gameday with escalations
- Update contact information
- Review recent escalation incidents

### Escalation Drills

**Quarterly Gameday Exercises**:
1. Simulate SEV-1 incident requiring escalation
2. Practice calling executives (with advance notice)
3. Test PagerDuty escalation policies
4. Verify contact information is current
5. Review escalation effectiveness

---

## References

- [Incident Response Process](INCIDENT_RESPONSE.md)
- [On-Call Rotation Setup](ON_CALL_ROTATION.md)
- [Communication Templates](COMMUNICATION_TEMPLATES.md)
- [PagerDuty Escalation Policies](https://virtengine.pagerduty.com/escalation_policies)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Next Review**: 2026-04-30
