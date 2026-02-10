# On-Call Rotation Setup and Management

## Overview

This document defines the on-call rotation structure, responsibilities, and operational procedures for VirtEngine. A well-managed on-call rotation ensures 24/7 coverage while maintaining engineer health and effectiveness.

## Table of Contents

1. [On-Call Structure](#on-call-structure)
2. [Responsibilities](#responsibilities)
3. [Requirements and Eligibility](#requirements-and-eligibility)
4. [Rotation Schedule](#rotation-schedule)
5. [Handoff Procedures](#handoff-procedures)
6. [Compensation and Time Off](#compensation-and-time-off)
7. [Tools and Access](#tools-and-access)
8. [Escalation](#escalation)
9. [Best Practices](#best-practices)

---

## On-Call Structure

### Rotation Tiers

VirtEngine uses a multi-tier on-call structure:

| Tier | Name | Purpose | Response Time |
|------|------|---------|---------------|
| **Tier 1** | Primary On-Call | First responder to all alerts | < 5 minutes |
| **Tier 2** | Secondary On-Call | Backup if primary doesn't respond | Automatic escalation at 5 min |
| **Tier 3** | Incident Commander | SEV-1/SEV-2 coordination | < 10 minutes when paged |
| **Tier 4** | Engineering Manager | Escalation for complex issues | < 30 minutes when paged |

### Coverage Model

**24/7 Follow-the-Sun**:
- Americas: 6am - 6pm UTC-8 (San Francisco)
- EMEA: 6am - 6pm UTC+0 (London)
- APAC: 6am - 6pm UTC+8 (Singapore)
- Night coverage: Rotates among all regions

**Alternative: 24/7 Full Rotation**:
- Single engineer on-call for full week
- Includes nights and weekends
- Higher compensation required

**Current Model**: Follow-the-Sun with night rotation

---

## Responsibilities

### Primary On-Call Engineer

**Core Responsibilities**:

1. **Alert Response**
   - Acknowledge all alerts within 5 minutes
   - Triage and assess severity
   - Declare incident if necessary
   - Begin investigation and mitigation

2. **Incident Management**
   - Lead response for SEV-3/SEV-4 incidents
   - Coordinate with Incident Commander for SEV-1/SEV-2
   - Document timeline and actions
   - Update status page as appropriate

3. **Communication**
   - Provide regular updates in incident channel
   - Escalate when needed
   - Notify stakeholders per escalation procedures
   - Update on-call log

4. **Maintenance and Improvements**
   - Fix urgent issues discovered during shift
   - Create tickets for non-urgent issues
   - Update runbooks based on learnings
   - Report tooling gaps

**Time Commitment**:
- Be available to respond within 5 minutes
- No expectation to be at computer 24/7
- Mobile access sufficient for acknowledgment
- Laptop required for investigation

**NOT Responsible For**:
- ❌ Being on-site/at office
- ❌ Solving all problems alone (escalate!)
- ❌ Working on projects during on-call
- ❌ Responding to non-urgent requests

---

### Secondary On-Call Engineer

**Responsibilities**:
- Serve as backup if primary doesn't respond
- Automatic escalation after 5 minutes
- Support primary on complex incidents
- Cover if primary needs break during incident

**Escalation Trigger**:
- Primary doesn't acknowledge within 5 minutes
- Primary requests backup
- SEV-1 incident (automatic escalation)

---

### Incident Commander (On-Call)

**Responsibilities**:
- Lead SEV-1/SEV-2 incident response
- Coordinate all responders
- Make high-level decisions
- Manage communication to executives
- Ensure incident timeline documented
- Lead postmortem process

**Selection Criteria**:
- Senior SRE or Engineering Lead
- Incident response experience
- Strong communication skills
- Deep technical knowledge

**Not Required**:
- Don't need to be most technical person
- Don't need to solve problem personally
- Focus is coordination, not implementation

---

## Requirements and Eligibility

### Prerequisites

**Before First On-Call Shift**:

- [ ] Completed on-call training program
- [ ] Shadowed 2 full on-call shifts
- [ ] Read all runbooks and procedures
- [ ] Access to all production systems verified
- [ ] PagerDuty configured and tested
- [ ] Responded to 3+ incidents (as shadow)
- [ ] Passed on-call readiness review
- [ ] Emergency contacts registered

**Technical Requirements**:
- Deep understanding of VirtEngine architecture
- Proficiency with monitoring/logging tools
- Strong debugging skills
- Familiarity with common failure modes
- Basic knowledge of all major components

**Soft Skills**:
- Calm under pressure
- Clear communication
- Good judgment
- Willingness to escalate
- Team player

---

### On-Call Training Program

**Week 1-2: Preparation**
- Read all incident response documentation
- Review runbooks for all services
- Verify tool access (Grafana, Prometheus, PagerDuty, logs)
- Attend on-call orientation session
- Set up on-call equipment (laptop, phone, chargers)

**Week 3-4: Shadowing**
- Shadow 2 complete on-call shifts
- Observe real incident responses
- Practice using tools
- Ask questions and take notes

**Week 5: Simulation**
- Complete 2 incident response drills
- Practice declaring incidents
- Practice escalation procedures
- Get feedback from trainers

**Week 6: Graduation**
- Pass on-call readiness quiz
- Pass hands-on incident simulation
- Manager approval
- Added to on-call rotation

---

## Rotation Schedule

### Schedule Structure

**Shift Duration**: 1 week (Monday 9am - Monday 9am local time)

**Example Schedule** (4-person rotation):

| Week | Primary On-Call | Secondary On-Call | IC On-Call |
|------|----------------|-------------------|------------|
| Week 1 | Alice | Bob | Carol |
| Week 2 | Bob | Carol | Dave |
| Week 3 | Carol | Dave | Alice |
| Week 4 | Dave | Alice | Bob |

**Rotation Cycle**: 4 weeks between shifts (for 4-person team)

---

### Schedule Management

**Tool**: PagerDuty

**Schedule Creation**:
1. Create schedules for each tier (Primary, Secondary, IC)
2. Configure rotation rules
3. Add all eligible engineers
4. Set up timezone handling
5. Configure escalation policies

**Schedule Visibility**:
- Published 3 months in advance
- Accessible in PagerDuty and Slack
- Calendar integration available
- Mobile app shows your next shift

**Schedule Changes**:
- Use PagerDuty override feature
- Document reason in on-call channel
- Notify team of changes
- Get manager approval for permanent changes

---

### Holidays and Time Off

**Policy**:
- On-call engineers may request shift swaps
- Holiday weeks should have volunteers (bonus pay)
- Cannot be scheduled for on-call during PTO
- Must have manager approval for shift swaps

**Shift Swap Procedure**:
1. Find engineer to swap with
2. Create PagerDuty override
3. Notify in #oncall-schedule channel
4. Document in shift log
5. Both engineers confirm

**Holiday Coverage**:
- Extra compensation: 2x on-call pay
- Volunteer-based if possible
- Mandatory rotation if insufficient volunteers
- Limit to 1 holiday shift per year

---

## Handoff Procedures

### Shift Start Handoff

**Timing**: Monday 9am local time (15-30 minute handoff call)

**Outgoing Engineer Provides**:

1. **Incident Summary**
   - Active incidents (if any)
   - Recent incidents (past week)
   - Recurring issues to watch for

2. **System Status**
   - Any ongoing degradations
   - Scheduled maintenance
   - Recent deployments
   - Known issues with workarounds

3. **Open Actions**
   - Tickets created during shift
   - Follow-up actions required
   - Escalated issues

4. **Notes and Tips**
   - New runbooks or procedures
   - Tool changes
   - Team member availability

**Incoming Engineer Confirms**:
- [ ] Received handoff notes
- [ ] PagerDuty shows correct schedule
- [ ] Can receive test page
- [ ] No questions or concerns
- [ ] Ready to take over

**Handoff Template** (Slack message):

```
📋 ON-CALL HANDOFF - Week of [Date]

Outgoing: @alice
Incoming: @bob

INCIDENTS THIS WEEK:
- [Date] SEV-2: API latency spike (resolved, postmortem: LINK)
- [Date] SEV-3: Database slow query (fixed, ticket: JIRA-123)

SYSTEM STATUS:
- All systems healthy ✅
- Known issue: Occasional provider-daemon restarts (JIRA-456, non-urgent)
- Scheduled: Database maintenance on [Date] at [Time] (automated)

OPEN ACTIONS:
- Monitor JIRA-456 (provider-daemon restarts)
- Follow up with Alice on postmortem action items (due Friday)

NOTES:
- New runbook added for database connection issues
- Bob is on PTO Thu-Fri (escalate to Carol if needed)

Ready to hand off? [Incoming confirms]
```

---

### Shift End Handoff

**Responsibilities**:

**Outgoing Engineer Must**:
- Complete handoff document
- Schedule handoff call
- Transfer any active incidents
- Update on-call log
- Document lessons learned

**Cannot Hand Off**:
- Active SEV-1 incidents (stay until resolved)
- Unacknowledged alerts
- Ongoing investigations without documentation

---

## Compensation and Time Off

### On-Call Compensation

**Standard Model**:
- **Base On-Call Pay**: $X per shift (regardless of alerts)
- **Incident Response Pay**: $Y per hour actively responding
- **SEV-1 Response Bonus**: $Z per incident

**Alternative Model**:
- **Time Off in Lieu (TOIL)**: 1 day off per on-call week
- **Incident Response TOIL**: Hours worked count towards TOIL

**Weekend Coverage**:
- Weekend shifts: 1.5x on-call pay
- Holiday shifts: 2x on-call pay

**Note**: Actual compensation determined by HR policy

---

### Time Off After Incidents

**Policy**:
- After SEV-1 incident > 4 hours: Take rest of day off
- After overnight incident > 2 hours: Take next day off
- Manager approval not required (inform team)
- No TOIL deduction (separate from on-call TOIL)

**Rationale**: Well-rested engineers are more effective

---

## Tools and Access

### Required Tools

**Communication**:
- [x] Slack (mobile app + desktop)
- [x] PagerDuty (mobile app configured)
- [x] Zoom (for war rooms)
- [x] Email (for escalations)

**Monitoring and Alerting**:
- [x] Grafana (dashboards access)
- [x] Prometheus (query access)
- [x] PagerDuty (alert management)

**Logging and Tracing**:
- [x] Elasticsearch/Kibana (log access)
- [x] Jaeger (distributed tracing)

**Incident Management**:
- [x] Jira (ticket creation)
- [x] Status page (update access)
- [x] Runbook repository (read/write)

**Infrastructure**:
- [x] AWS Console (production read, emergency write)
- [x] Kubernetes (production clusters)
- [x] Database (read-only, emergency write)
- [x] Deployment tools (Ansible, CI/CD)

---

### Access Requirements

**Production Access**:
- Read access: Always available
- Write access: Emergency only (with audit logging)
- Database access: Read-only (write requires approval)
- SSH access: Limited to specific jump hosts

**Security**:
- MFA required for all production access
- SSH keys rotated quarterly
- Access logs audited
- VPN required for remote access

---

### Tool Setup Checklist

**Before First Shift**:

- [ ] PagerDuty profile complete (phone, email, SMS)
- [ ] PagerDuty mobile app installed and tested
- [ ] Test page received successfully
- [ ] Grafana dashboards bookmarked
- [ ] Kibana log queries saved
- [ ] Runbooks bookmarked/downloaded
- [ ] Slack notifications configured
- [ ] Emergency contacts list saved
- [ ] VPN credentials tested
- [ ] Laptop fully charged
- [ ] Phone charger accessible
- [ ] Backup internet available (mobile hotspot)

---

## Escalation

### When to Escalate

**Immediate Escalation Required**:
- ❗ SEV-1 incident (automatic)
- ❗ Security incident
- ❗ Data loss or corruption
- ❗ Incident duration > 30 minutes (SEV-2)
- ❗ Unsure how to proceed

**Consider Escalation**:
- Need specialized expertise
- Multiple systems affected
- Customer impact growing
- Mitigation attempts failing
- Executive decision required

**Don't Hesitate**:
- Escalating is NOT a failure
- Better to escalate early
- Experienced engineers escalate too
- It's a team effort

---

### How to Escalate

**Primary → Secondary**:
- Page via PagerDuty (automatic after 5 min)
- Or manual page if need help sooner
- Message in incident channel

**To Incident Commander**:
- For SEV-1/SEV-2
- Page via PagerDuty IC rotation
- Provide incident summary

**To Subject Matter Expert**:
- Page via PagerDuty team-specific rotation
- Or message in Slack with @mention
- Provide context: "Need database expertise"

**To Management**:
- Use escalation procedures (see ESCALATION_PROCEDURES.md)
- For executive decisions
- For complex multi-team issues

---

## Best Practices

### During On-Call Week

✅ **Do**:
- Keep laptop and phone charged
- Stay near reliable internet
- Limit alcohol consumption
- Have backup communication (mobile hotspot)
- Update handoff notes daily
- Take breaks during long incidents
- Document everything
- Escalate when needed

❌ **Don't**:
- Travel without backup plan
- Ignore alerts (acknowledge immediately)
- Try to be a hero (escalate!)
- Work on risky changes
- Forget to eat/sleep during incidents
- Skip handoff procedures

---

### Response Workflow

**When Alert Fires**:

1. **Acknowledge** (< 5 minutes)
   - Stop alert from re-firing
   - Reduces noise for team

2. **Assess** (5-10 minutes)
   - Is this a real issue?
   - What's the user impact?
   - What's the severity?

3. **Declare or Resolve** (10-15 minutes)
   - Declare incident if needed
   - Or resolve if false positive
   - Document decision

4. **Respond** (ongoing)
   - Follow incident response procedures
   - Communicate regularly
   - Escalate if needed

---

### Self-Care

**During On-Call Week**:
- Maintain normal sleep schedule
- Take breaks during long incidents
- Ask for relief if exhausted
- Don't skip meals
- Exercise and stress relief

**After Major Incident**:
- Take time off as needed
- Debrief with team
- Process emotions (incidents are stressful)
- Celebrate resolution

**Burnout Prevention**:
- Limit on-call to 1 week per month
- Rotate fairly
- Provide adequate compensation
- Support team members

---

## Metrics and Reporting

### On-Call Metrics

**Alert Metrics**:
- Total alerts per shift
- Alerts requiring response
- False positive rate
- Alert fatigue score

**Response Metrics**:
- Time to acknowledge
- Time to resolve
- Escalation rate
- After-hours incidents

**Shift Metrics**:
- Incidents per shift
- SEV-1/SEV-2 count
- Total time responding
- Sleep disruption

**Quality Metrics**:
- Runbook usage
- Documentation completeness
- Handoff quality
- Team satisfaction

---

### Reporting

**Weekly On-Call Report**:
```
On-Call Report: Week of [Date]
On-Call Engineer: @alice

INCIDENTS:
- Total: 3
- SEV-2: 1 (API latency)
- SEV-3: 2 (minor issues)

ALERTS:
- Total alerts: 47
- Action required: 12
- False positives: 8 (17%)

TIME SPENT:
- Total response time: 6 hours
- After-hours: 2 hours
- Longest incident: 2 hours

IMPROVEMENTS:
- Updated runbook for database issues
- Created ticket for noisy alerts (JIRA-789)
- Identified monitoring gap (disk usage alerts)

HANDOFF:
- All incidents resolved ✅
- One known issue (JIRA-456)
- System healthy
```

---

## Schedule and Contacts

### Current Rotation

**Primary On-Call**: See PagerDuty schedule
**Secondary On-Call**: See PagerDuty schedule
**IC On-Call**: See PagerDuty schedule

**PagerDuty**: https://virtengine.pagerduty.com

**Schedule**: https://virtengine.pagerduty.com/schedules

---

### Emergency Contacts

**On-Call Hotline**: [Phone number]
**Engineering Manager**: [Phone]
**VP Engineering**: [Phone]
**CISO (Security)**: [Phone]

**Internal**: See company directory

---

## References

- [Incident Response Process](INCIDENT_RESPONSE.md)
- [Escalation Procedures](ESCALATION_PROCEDURES.md)
- [Communication Templates](COMMUNICATION_TEMPLATES.md)
- [Incident Drills](INCIDENT_DRILLS.md)
- [Runbooks](runbooks/README.md)
- [PagerDuty Documentation](https://support.pagerduty.com/)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Next Review**: 2026-04-30
