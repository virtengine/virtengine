# Communication Templates

## Overview

This document provides standardized communication templates for incident response. Consistent messaging ensures clarity, reduces confusion, and maintains trust with stakeholders.

## Table of Contents

1. [Internal Communications](#internal-communications)
2. [External Communications](#external-communications)
3. [Stakeholder Communications](#stakeholder-communications)
4. [Post-Incident Communications](#post-incident-communications)
5. [Special Situations](#special-situations)

---

## Internal Communications

### Incident Declaration (Slack)

**Template: SEV-1/SEV-2**

```
🚨 INCIDENT DECLARED 🚨

Severity: [SEV-1 | SEV-2]
Service: [Service Name]
Impact: [Brief description of user impact]
Start Time: [HH:MM UTC]

Incident Commander: @[username]
Incident Channel: #incident-YYYY-MM-DD-HHMM
Status Page: https://status.virtengine.com

Current Status: [Investigating | Identified | Monitoring]

All available responders, please join the incident channel.
```

**Example**:

```
🚨 INCIDENT DECLARED 🚨

Severity: SEV-1
Service: VirtEngine API
Impact: Complete API outage - all requests returning 500 errors
Start Time: 14:23 UTC

Incident Commander: @bob-jones
Incident Channel: #incident-2026-01-30-1423
Status Page: https://status.virtengine.com

Current Status: Investigating

All available responders, please join the incident channel.
```

---

**Template: SEV-3/SEV-4**

```
⚠️ Incident: [SEV-3 | SEV-4]

Service: [Service Name]
Impact: [Brief description]
Assignee: @[username]
Ticket: [JIRA-123]

No immediate action required. Updates in #sre-team.
```

---

### Incident Updates (Slack)

**Template: Investigation Update**

```
[HH:MM] Incident Update - Investigation

Findings:
- [What we've discovered]
- [Relevant metrics, logs, patterns]

Hypothesis:
- [Current theory of root cause]

Next Steps:
- [What we're doing next]
- [Who is working on what]

ETA: [Time estimate or "Unknown"]
```

**Example**:

```
[14:35] Incident Update - Investigation

Findings:
- Database connection pool exhausted (100/100 connections)
- Recent deployment: API v1.2.3 at 14:15 UTC
- Error logs show "cannot acquire connection" errors
- Traffic spike from 500 RPS → 800 RPS at 14:20

Hypothesis:
- New feature in v1.2.3 increased database query load
- Connection pool size insufficient for current traffic

Next Steps:
- @alice analyzing query patterns in v1.2.3
- @bob preparing rollback plan
- @carol investigating connection pool metrics

ETA: 10-15 minutes for diagnosis
```

---

**Template: Mitigation Update**

```
[HH:MM] Incident Update - Mitigation in Progress

Action: [What we're doing]
Method: [Rollback | Config Change | Hotfix | etc.]
Risk: [Low | Medium | High]
ETA: [Time estimate]

Rollback Plan: [How we'll revert if it fails]

Monitoring: [What we're watching]
```

**Example**:

```
[14:50] Incident Update - Mitigation in Progress

Action: Increasing database connection pool size
Method: Configuration change (100 → 500 connections)
Risk: Low (requires API restart, ~30 seconds downtime)
ETA: 5 minutes

Rollback Plan: Revert configuration via Ansible if issue persists

Monitoring: Connection pool utilization, error rate, API latency
```

---

**Template: Resolution**

```
[HH:MM] ✅ INCIDENT RESOLVED

Resolution: [What fixed it]
Duration: [Total incident time]
Impact: [Summary of customer impact]

Metrics:
- Error Rate: [Current vs Normal]
- Latency: [Current vs Normal]
- Availability: [Current status]

Post-Incident:
- Postmortem scheduled: [Date/Time]
- Action items: [Count] created
- Incident ticket: [JIRA-123]

Thank you to all responders! 🙏
```

**Example**:

```
[15:10] ✅ INCIDENT RESOLVED

Resolution: Increased database connection pool to 500 connections
Duration: 47 minutes (14:23 - 15:10 UTC)
Impact: ~1,200 users affected, complete API outage

Metrics:
- Error Rate: 0.3% (normal: < 0.5%) ✅
- Latency: P95 1.8s (normal: < 2s) ✅
- Availability: 100% (all health checks passing) ✅

Post-Incident:
- Postmortem scheduled: 2026-01-31 10:00 AM
- Action items: 12 created
- Incident ticket: INCIDENT-12345

Thank you to all responders! 🙏
Special thanks to @alice @bob @carol for quick response.
```

---

### Escalation Request (Slack)

**Template**:

```
🚨 ESCALATION REQUEST 🚨

To: @[escalation-target]
From: @[your-username]
Incident: #incident-YYYY-MM-DD-HHMM
Severity: [SEV-X]

Reason:
- [Why escalating]
- [What's blocking resolution]
- [What decision/expertise needed]

Current Status:
- [Brief status]
- [Actions attempted]
- [Current impact]

Requested Action:
- [Specific ask]

Please join incident channel: #incident-YYYY-MM-DD-HHMM
```

---

### War Room Announcement

**Template**:

```
🎙️ WAR ROOM ACTIVE

Zoom Link: [URL]
Passcode: [Code]

All responders, please join for real-time coordination.

Roles:
- Incident Commander: @[name]
- Tech Lead: @[name]
- Comms Lead: @[name]
- Scribe: @[name]

Slack updates will continue in parallel for those not on call.
```

---

## External Communications

### Status Page Updates

**Template: Investigating**

```
Status: Investigating
Affected Services: [Service Name(s)]

We are currently investigating an issue affecting [service name]. 
Users may experience [specific symptoms: errors, slowness, unavailability].

This incident affects [scope: all users, US region, specific features].

We are actively investigating and will provide updates every 15 minutes.

Next update: [HH:MM UTC]
```

**Example**:

```
Status: Investigating
Affected Services: VirtEngine API

We are currently investigating an issue affecting the VirtEngine API.
Users may experience errors when making API requests (HTTP 500 responses).

This incident affects all users globally.

We are actively investigating and will provide updates every 15 minutes.

Next update: 14:40 UTC
```

---

**Template: Identified**

```
Status: Identified
Affected Services: [Service Name(s)]

We have identified the cause of the issue affecting [service name].

Root Cause: [Brief non-technical explanation]

Our team is implementing a fix. We expect service restoration within [time estimate or "the next hour"].

We will continue to provide updates every 15 minutes.

Next update: [HH:MM UTC]
```

**Example**:

```
Status: Identified
Affected Services: VirtEngine API

We have identified the cause of the issue affecting the VirtEngine API.

Root Cause: Database connection limit exceeded due to higher than expected traffic.

Our team is implementing a fix. We expect service restoration within the next 20 minutes.

We will continue to provide updates every 15 minutes.

Next update: 15:00 UTC
```

---

**Template: Monitoring**

```
Status: Monitoring
Affected Services: [Service Name(s)]

A fix has been applied for the issue affecting [service name].

We are currently monitoring the service to ensure stability before declaring full resolution.

Service functionality should be restored, but we are watching closely for any remaining issues.

If you continue to experience problems, please contact support.

Next update: [HH:MM UTC]
```

---

**Template: Resolved**

```
Status: Resolved
Affected Services: [Service Name(s)]

The issue affecting [service name] has been fully resolved.

Duration: [X] minutes ([Start] - [End] UTC)
Impact: [Brief description of what users experienced]

Root Cause: [Brief non-technical explanation]
Resolution: [What we did to fix it]

We apologize for the disruption. A detailed post-incident report will be published within 5 business days.

Thank you for your patience.
```

**Example**:

```
Status: Resolved
Affected Services: VirtEngine API

The issue affecting the VirtEngine API has been fully resolved.

Duration: 47 minutes (14:23 - 15:10 UTC)
Impact: Users experienced errors when making API requests. All requests are now processing normally.

Root Cause: Database connection limit exceeded during a traffic spike.
Resolution: Increased database connection pool capacity to handle higher load.

We apologize for the disruption. A detailed post-incident report will be published within 5 business days at https://status.virtengine.com/incidents/[ID]

Thank you for your patience.
```

---

### Social Media

**Template: Acknowledgment (Twitter/X)**

```
We're aware of an issue affecting [service]. Our team is investigating. 
Status updates: https://status.virtengine.com
```

**Template: Resolution (Twitter/X)**

```
The issue affecting [service] has been resolved. Service is fully operational. 
We apologize for the disruption. 
Post-mortem: https://status.virtengine.com/incidents/[ID]
```

---

## Stakeholder Communications

### Customer Support Team

**Template: Initial Alert**

```
Subject: [SEV-X] INCIDENT ALERT - [Service Name]

Team,

We are experiencing a [SEV-X] incident affecting [service name].

IMPACT:
- [What customers are experiencing]
- [Which customers are affected: all, subset, region]
- [Started at: HH:MM UTC]

STATUS:
- Current status: [Investigating | Identified | Fixing]
- ETA: [Time estimate or "Unknown"]
- Status page: https://status.virtengine.com

SUPPORT GUIDANCE:
- Acknowledge customer reports
- Direct customers to status page
- Do NOT provide ETAs beyond what's on status page
- Do NOT provide technical details of root cause
- Escalate urgent cases to: [contact]

FAQ for Customer Inquiries:
Q: When will this be fixed?
A: We're actively working on it. Updates at https://status.virtengine.com

Q: What's causing this?
A: Our engineering team is investigating. Details will be shared in our post-incident report.

Q: Will I get a refund/credit?
A: We'll provide information on service credits after the incident is resolved.

Incident Channel (internal only): #incident-YYYY-MM-DD-HHMM
Next update: [HH:MM UTC]

- SRE Team
```

---

**Template: Support Update**

```
Subject: [SEV-X] UPDATE - [Service Name]

Team,

Update on the incident affecting [service name]:

STATUS: [Current status]
PROGRESS: [What's changed since last update]
ETA: [Updated estimate or "Still investigating"]

UPDATED FAQ:
[Any new Q&A based on customer inquiries]

ESCALATION:
[Any special instructions for urgent cases]

Next update: [HH:MM UTC]

- SRE Team
```

---

**Template: Resolution Notification**

```
Subject: [SEV-X] RESOLVED - [Service Name]

Team,

The incident affecting [service name] has been resolved.

SUMMARY:
- Duration: [X] minutes
- Impact: [What customers experienced]
- Resolution: [Brief explanation]

CUSTOMER COMMUNICATION:
- Service is fully operational
- Post-incident report will be published within 5 days
- Service credits (if applicable) will be processed within [timeframe]

SUPPORT GUIDANCE:
- Normal support processes resume
- Direct remaining questions to post-incident report (when published)
- Escalate any ongoing issues immediately

Thank you for your support during this incident.

Post-incident report: [Link when available]

- SRE Team
```

---

### Engineering Leadership

**Template: Executive Summary (Email)**

```
Subject: [SEV-X] Incident Summary - [Service Name]

Leadership Team,

We experienced a [SEV-X] incident affecting [service name].

BUSINESS IMPACT:
- Duration: [X] minutes ([Start] - [End] UTC)
- Customers affected: [Number/Percentage]
- Revenue impact: [$ estimate or "Minimal"]
- SLO impact: [Which SLOs violated, error budget consumed]

ROOT CAUSE:
- [Brief non-technical explanation]

RESOLUTION:
- [What we did to restore service]

NEXT STEPS:
- Post-incident review scheduled: [Date/Time]
- Action items to prevent recurrence: [Count]
- Customer communication: [Status page updated, emails sent]

LESSONS LEARNED:
- [Key takeaway 1]
- [Key takeaway 2]

Full postmortem will be available by [Date].

Please let me know if you have any questions.

[Your Name]
SRE Team
```

---

### Partner/Vendor Notification

**Template: External Dependency Issue**

```
Subject: Service Degradation Due to [Vendor Name] Outage

Dear [Vendor Contact],

We are currently experiencing service degradation due to an outage in your [service/region].

DETAILS:
- VirtEngine services affected: [List]
- Started: [HH:MM UTC]
- Impact: [Description]

YOUR STATUS:
- Your status page: [URL]
- Reported issue: [Description from their status page]

REQUEST:
- ETA for resolution
- Priority escalation for VirtEngine (Customer ID: [ID])
- Direct contact for updates

We have implemented [failover/workaround] but are dependent on your service for full restoration.

Please respond urgently.

Thank you,
[Your Name]
VirtEngine SRE Team
[Contact Info]
```

---

## Post-Incident Communications

### Customer Email (Optional for Major Incidents)

**Template**:

```
Subject: Update on [Date] Service Disruption

Dear VirtEngine Customers,

On [Date] at [Time UTC], VirtEngine experienced [brief description of incident]. We know you depend on our service, and we sincerely apologize for this disruption.

WHAT HAPPENED:
[Non-technical explanation of what customers experienced]

IMPACT:
- Duration: [X] minutes ([Start] - [End] UTC)
- Affected services: [List]
- Affected regions: [List or "Global"]

ROOT CAUSE:
[Honest, non-technical explanation of why it happened]

RESOLUTION:
[What we did to fix it]

WHAT WE'RE DOING TO PREVENT THIS:
[Specific improvements being implemented]
- [Action 1]
- [Action 2]
- [Action 3]

SERVICE CREDITS (if applicable):
[Information about credits or SLA compensation]

We take reliability extremely seriously. A full technical post-incident report is available at [URL].

If you have any questions or concerns, please contact our support team at support@virtengine.com.

Thank you for your patience and continued trust in VirtEngine.

Sincerely,
[Name]
[Title]
VirtEngine
```

---

### Public Post-Incident Report

**Template** (Blog Post):

```markdown
# Post-Incident Report: [Date] Service Disruption

**Date**: [YYYY-MM-DD]
**Author**: [Name, Title]
**Status**: Final

## Summary

On [Date] at [Time UTC], VirtEngine experienced a [duration]-minute service disruption affecting [services]. This post explains what happened, why it happened, and what we're doing to prevent it from happening again.

## What Happened

[Detailed timeline in user-friendly language]

## Impact

- **Duration**: [X] minutes
- **Affected Services**: [List]
- **Users Impacted**: [Number/Percentage]
- **Error Rate**: [Peak error rate]

## Root Cause

[Technical but accessible explanation of root cause]

## Resolution

[How we fixed it, including any temporary mitigations]

## What We're Doing Differently

We've identified [N] action items to prevent similar incidents:

1. **[Action 1]**: [Explanation of what and why]
   - Owner: [Team]
   - Timeline: [Date]

2. **[Action 2]**: [Explanation]
   - Owner: [Team]
   - Timeline: [Date]

[Continue for all major actions]

## Timeline (UTC)

| Time | Event |
|------|-------|
| [HH:MM] | [Event] |
| [HH:MM] | [Event] |

## Conclusion

We take reliability seriously and are committed to learning from every incident. Thank you for your patience and continued trust in VirtEngine.

If you have questions about this incident, please contact support@virtengine.com.

---

*For technical details, see our internal postmortem at [internal link].*
```

---

## Special Situations

### Security Incident Communication

**IMPORTANT**: Security incidents require special handling.

**Internal Communication**:
- Use PRIVATE channel: `#security-incident-[ID]` (not regular incident channel)
- Limit access to need-to-know basis
- Do NOT share details publicly until cleared by Security Team and Legal

**External Communication**:

**Template: Status Page (Investigating)**

```
Status: Investigating
Affected Services: [Service Name]

We are investigating a technical issue affecting [service name].
Users may experience [symptoms - be vague, don't mention security].

Updates will be provided as we learn more.
```

**Template: Status Page (Resolved)**

```
Status: Resolved
Affected Services: [Service Name]

The technical issue affecting [service name] has been resolved.

Duration: [X] minutes

We are conducting a thorough review and will provide more details as appropriate.
```

**Customer Communication (if data breach)**:

```
Subject: Important Security Notice - VirtEngineFollow legal and compliance team guidance exactly.
Include:
- What happened (facts only, verified)
- What data was affected
- What we've done to secure systems
- What customers should do
- Where to get more information
- Contact for questions

Coordinate with: Legal, CISO, PR, Customer Support
```

---

### Data Loss Communication

**Template: Internal (Immediate)**

```
🚨 DATA LOSS INCIDENT 🚨

Severity: SEV-1
Type: Data Loss/Corruption

Affected Data: [Description]
Scope: [Number of records/users affected]
Recovery Status: [Investigating | Partial | Full]

CRITICAL:
- All write operations to [system] HALTED
- Backups being validated
- Recovery plan being developed

Incident Commander: @[name]
Database Lead: @[name]
Legal Notified: [Yes/No]

Incident Channel: #incident-[ID] (PRIVATE)

This is a CONFIDENTIAL incident. Do not discuss outside incident channel.
```

**Template: Customer Communication**

```
Subject: Important Notice Regarding Your VirtEngine Data

Dear VirtEngine Customer,

We are writing to inform you of an issue that may have affected your data.

WHAT HAPPENED:
[Honest explanation of what occurred]

YOUR DATA:
- Affected data: [What data types]
- Status: [Recovered | Partially Recovered | Lost]
- Scope: [What was affected, what was not]

RECOVERY:
[What we've done to recover data and secure systems]

WHAT YOU SHOULD DO:
[Specific actions customers should take]

We understand the seriousness of this issue and are taking full responsibility. 

Support: [Contact information for questions]
Compensation: [If applicable]

We sincerely apologize for this incident.

[Name, Title]
VirtEngine
```

---

### Planned Maintenance Communication

**Template: Announcement (7 days before)**

```
Subject: Scheduled Maintenance - [Date]

Dear VirtEngine Customers,

We will be performing scheduled maintenance to improve our infrastructure.

MAINTENANCE WINDOW:
- Date: [YYYY-MM-DD]
- Time: [HH:MM - HH:MM UTC] ([Duration])
- Timezone converter: https://time.is/UTC

IMPACT:
- Services affected: [List]
- Expected downtime: [Duration]
- [Service] will be unavailable during this window

PREPARATION:
[Any actions customers should take beforehand]

We apologize for any inconvenience. This maintenance is necessary to [brief explanation of benefit].

Questions? Contact support@virtengine.com

Thank you,
VirtEngine Operations Team
```

**Template: Reminder (24 hours before)**

```
Subject: REMINDER: Maintenance Tomorrow - [Date]

This is a reminder that VirtEngine will undergo scheduled maintenance tomorrow.

WHEN: [Date] at [HH:MM UTC] ([Duration])
IMPACT: [Brief summary]
STATUS PAGE: https://status.virtengine.com

Please plan accordingly.

Questions? Contact support@virtengine.com
```

---

## Communication Best Practices

### Do's ✅

- **Be honest**: Don't hide or minimize issues
- **Be specific**: Include concrete details (times, numbers, scope)
- **Be timely**: Update regularly, even if no new information
- **Be empathetic**: Acknowledge customer frustration
- **Be accountable**: Take responsibility, don't blame others
- **Be clear**: Use simple language, avoid jargon

### Don'ts ❌

- **Don't speculate**: Only share confirmed information
- **Don't give false hope**: Don't promise ETAs you can't meet
- **Don't blame**: No finger-pointing (internal or external)
- **Don't overshare**: Technical details for internal use only
- **Don't disappear**: Maintain regular updates even during long incidents
- **Don't forget**: Always close the loop with final summary

---

## Language Guidelines

### Status Page Language

**Good**:
- "We are investigating reports of errors"
- "Service is degraded"
- "A fix has been applied"
- "Service is fully restored"

**Avoid**:
- "We're having some issues" (too vague)
- "Everything is broken" (too alarming)
- "It's probably fixed" (uncertain)
- "The database crashed" (too technical)

### Customer Communication

**Good**:
- "We understand this caused disruption to your business"
- "We take full responsibility"
- "We're implementing [specific improvements]"

**Avoid**:
- "Sorry for the inconvenience" (minimizing)
- "It wasn't our fault" (deflecting)
- "These things happen" (dismissive)

---

## References

- [Incident Response Process](INCIDENT_RESPONSE.md)
- [Escalation Procedures](ESCALATION_PROCEDURES.md)
- [Postmortem Template](templates/postmortem_template.md)
- [Status Page](https://status.virtengine.com)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Next Review**: 2026-04-30
