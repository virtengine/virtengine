# Incident Response Drills and Simulations

## Overview

Incident response drills (gamedays) are structured exercises to test and improve incident response capabilities. Regular drills ensure teams are prepared for real incidents and surface gaps in processes, tools, and knowledge.

## Table of Contents

1. [Drill Philosophy](#drill-philosophy)
2. [Drill Types](#drill-types)
3. [Drill Scenarios](#drill-scenarios)
4. [Planning a Drill](#planning-a-drill)
5. [Running a Drill](#running-a-drill)
6. [Evaluation and Learning](#evaluation-and-learning)
7. [Drill Schedule](#drill-schedule)
8. [Example Drills](#example-drills)

---

## Drill Philosophy

### Purpose

**Why We Run Drills**:
- Validate incident response procedures work in practice
- Build muscle memory for high-pressure situations
- Identify gaps in documentation, tools, and training
- Practice cross-team coordination
- Build confidence in handling real incidents
- Surface assumptions that may be incorrect

**Core Principles**:
- **Blameless**: Drills are for learning, not evaluation
- **Realistic**: Simulate real conditions as closely as possible
- **Safe**: No risk to production systems (use test environments)
- **Documented**: Capture learnings and improve processes
- **Regular**: Consistent schedule builds competency

---

### Types of Learning

**Technical Skills**:
- Debugging production systems
- Using monitoring and alerting tools
- Applying mitigation strategies
- Coordinating deployments under pressure

**Communication Skills**:
- Incident declaration and escalation
- Status updates to stakeholders
- Coordination across teams
- Customer communication

**Process Skills**:
- Incident command structure
- Role assignment and delegation
- Timeline documentation
- Postmortem creation

---

## Drill Types

### 1. Tabletop Exercise (Low Risk)

**Description**: Discussion-based walkthrough of incident scenarios without actually breaking anything.

**Duration**: 1-2 hours

**Participants**: 5-10 people (IC, responders, SMEs)

**How It Works**:
1. Facilitator presents scenario
2. Team discusses what they would do
3. Facilitator injects new information/complications
4. Team adapts response
5. Debrief and capture learnings

**When to Use**:
- Introducing new team members
- Testing new procedures
- Complex multi-team scenarios
- Regulatory compliance review

**Example**:
> "A database replication lag is growing. What do you check first? Now the primary database is showing high CPU. What do you do? Now you discover the backup system hasn't run in 3 days..."

---

### 2. Wheel of Misfortune (Medium Risk)

**Description**: Structured role-play of past incidents with randomized participant roles.

**Duration**: 1-2 hours

**Participants**: 3-8 people (rotating roles)

**How It Works**:
1. Select past incident postmortem
2. Randomly assign roles (IC, on-call, SME)
3. Facilitator narrates incident timeline
4. Participants respond as they would have
5. Compare to actual response
6. Discuss improvements

**When to Use**:
- Training new incident commanders
- Learning from past incidents
- Practicing escalation procedures
- Cross-training on different roles

**Example**:
> Use the database connection pool incident (INCIDENT-12345) with roles randomly assigned. Participants don't know the outcome.

---

### 3. Controlled Chaos (Medium-High Risk)

**Description**: Intentionally break non-production systems to practice real incident response.

**Duration**: 2-4 hours

**Participants**: 8-15 people (full incident response team)

**How It Works**:
1. Facilitator breaks staging/test environment
2. Monitoring alerts fire (real alerts)
3. Team responds as if production incident
4. All normal incident procedures followed
5. Debrief and action items

**When to Use**:
- Quarterly team training
- Testing new monitoring/alerting
- Validating runbooks
- Preparing for high-risk changes

**Safety Requirements**:
- ⚠️ **NEVER use production**
- Test environment must be production-like
- Clearly announce "DRILL IN PROGRESS"
- Have abort/revert plan ready

---

### 4. Production Gameday (High Risk)

**Description**: Intentional failure injection in production with extensive safety controls.

**Duration**: 4-8 hours

**Participants**: Full engineering team on standby

**How It Works**:
1. Extensive planning (weeks in advance)
2. Customer notification of maintenance window
3. Controlled failure injection
4. Real incident response
5. Immediate rollback capability
6. Executive approval required

**When to Use**:
- Testing disaster recovery
- Validating multi-region failover
- Chaos engineering initiatives
- Major infrastructure changes

**Safety Requirements**:
- ⚠️ **Requires executive approval**
- Customer communication required
- Maintenance window scheduled
- Full rollback plan tested
- Incident Commander assigned
- Abort criteria defined

**Example**:
> Simulate AWS region failure by blocking all traffic to us-west-2 to test automatic failover to us-east-1.

---

### 5. On-Call Shadow

**Description**: New on-call engineer shadows experienced engineer during real incidents.

**Duration**: Full on-call shift (1 week)

**Participants**: 2 people (experienced + new)

**How It Works**:
1. New engineer joins all pages/incidents
2. Experienced engineer leads response
3. New engineer observes and asks questions
4. New engineer takes increasing responsibility
5. Debrief after each incident

**When to Use**:
- Training new on-call engineers
- Onboarding to SRE team
- Learning new service/component

---

## Drill Scenarios

### Scenario 1: Complete API Outage (SEV-1)

**Objective**: Practice SEV-1 response, escalation, and communication

**Setup**:
- API service in staging returns 500 errors
- Monitoring alerts fire
- All requests failing

**Complications to Inject**:
- Initial mitigation (restart) doesn't work
- Database appears healthy (red herring)
- Rollback is blocked by migration
- Need executive approval for extended downtime

**Skills Practiced**:
- SEV-1 declaration
- Incident command structure
- Executive escalation
- Status page updates
- Multi-strategy mitigation

**Duration**: 2 hours

**Evaluation Criteria**:
- Time to detect: < 5 minutes
- Time to declare SEV-1: < 10 minutes
- Incident Commander assigned: < 15 minutes
- Status page updated: < 20 minutes
- Executive notified: < 30 minutes

---

### Scenario 2: Database Corruption (SEV-1)

**Objective**: Practice data loss response and recovery procedures

**Setup**:
- Test database shows corruption in critical table
- Backup restoration required
- Potential data loss window

**Complications to Inject**:
- Latest backup also corrupted (need older backup)
- Restore taking longer than expected
- Customer data potentially affected
- Legal notification required

**Skills Practiced**:
- Database recovery procedures
- Legal/compliance escalation
- Customer communication
- Data loss assessment
- Backup validation

**Duration**: 3 hours

---

### Scenario 3: Security Incident (SEV-1)

**Objective**: Practice security incident response

**Setup**:
- Alerts indicate unauthorized access attempts
- Suspicious activity in logs
- Potential data breach

**Complications to Inject**:
- Attack is ongoing (need to contain)
- Multiple attack vectors
- Legal notification required
- Press inquiry received

**Skills Practiced**:
- Security incident procedures
- System isolation
- Forensics preservation
- Legal coordination
- PR/communications

**Duration**: 4 hours

**Note**: Requires CISO involvement

---

### Scenario 4: Third-Party Outage (SEV-2)

**Objective**: Practice handling external dependencies

**Setup**:
- AWS region degradation
- Service impacted but not down
- No ETA from vendor

**Complications to Inject**:
- Failover partially works but has issues
- Customer complaints increasing
- Vendor communication is slow
- Executive asking for updates

**Skills Practiced**:
- External dependency handling
- Partial failover
- Customer communication
- Vendor escalation
- Executive reporting

**Duration**: 2 hours

---

### Scenario 5: Gradual Degradation (SEV-2)

**Objective**: Practice detection and diagnosis of subtle issues

**Setup**:
- Slowly increasing error rate (5% → 15%)
- Latency creeping up
- No obvious cause in logs
- Intermittent failures

**Complications to Inject**:
- Multiple possible causes
- Recent deployment (red herring)
- Memory leak causing issue
- Requires deep investigation

**Skills Practiced**:
- Detection of subtle issues
- Hypothesis-driven debugging
- Metric analysis
- Log correlation
- Gradual mitigation

**Duration**: 2-3 hours

---

### Scenario 6: Multi-Team Incident (SEV-2)

**Objective**: Practice cross-team coordination

**Setup**:
- Issue spans blockchain, provider daemon, and identity services
- Requires coordination across 3 teams
- No single team can resolve alone

**Complications to Inject**:
- Teams have different theories
- Need joint debugging session
- Competing priorities
- Communication challenges

**Skills Practiced**:
- Multi-team coordination
- Incident command with multiple SMEs
- Consensus building
- Joint troubleshooting

**Duration**: 3 hours

---

## Planning a Drill

### 4-6 Weeks Before

**Planning Phase**:

1. **Select Scenario**
   - Choose from library or create new
   - Based on recent incidents or risk areas
   - Consider team training needs

2. **Define Objectives**
   - What skills to practice
   - What to test (procedures, tools, communication)
   - Success criteria

3. **Assign Roles**
   - Drill Facilitator (runs the drill)
   - Observers (evaluate response)
   - Participants (respond to incident)
   - Red Team (inject complications)

4. **Prepare Environment**
   - Set up test/staging environment
   - Prepare failure injection
   - Test monitoring/alerting

5. **Get Approval**
   - Management approval
   - Participant availability
   - Environment reservations

---

### 2 Weeks Before

**Preparation Phase**:

1. **Announce Drill**
   - Date and time
   - Duration
   - Participants required
   - Objectives (high-level)

2. **Pre-Drill Materials**
   - Reminder of incident response procedures
   - Relevant runbooks
   - Tool access verification

3. **Test Setup**
   - Validate failure injection works
   - Test monitoring/alerting
   - Dry run with facilitator team

---

### 1 Week Before

**Final Preparation**:

1. **Confirmation**
   - Confirm participant availability
   - Verify environment ready
   - Review scenario with facilitators

2. **Logistics**
   - Calendar invites sent
   - Zoom links ready
   - Slack channels prepared
   - Documentation templates ready

---

### Day Of

**Execution Phase**:

1. **Pre-Drill Brief** (15 minutes)
   - Objectives
   - Rules of engagement
   - Abort criteria
   - Safety reminders

2. **Drill Execution** (2-4 hours)
   - Run scenario
   - Inject complications
   - Observe and document

3. **Debrief** (30-60 minutes)
   - What went well
   - What went poorly
   - Action items
   - Learnings

---

## Running a Drill

### Facilitator Checklist

**Before Drill Starts**:

- [ ] Environment ready and validated
- [ ] Monitoring/alerting configured
- [ ] Participants confirmed present
- [ ] Observers assigned
- [ ] Timeline doc ready for logging
- [ ] Abort plan ready
- [ ] Clear "DRILL IN PROGRESS" announcements

**During Drill**:

- [ ] Announce "DRILL START" clearly
- [ ] Inject initial failure
- [ ] Log all participant actions and decisions
- [ ] Inject complications at appropriate times
- [ ] Provide hints if participants stuck
- [ ] Monitor for safety issues
- [ ] Keep drill on schedule

**After Drill Concludes**:

- [ ] Announce "DRILL END" clearly
- [ ] Restore environment
- [ ] Collect participant feedback
- [ ] Schedule debrief
- [ ] Document observations

---

### Participant Guidelines

**During Drill**:

✅ **Do**:
- Respond as if real incident
- Follow all normal procedures
- Communicate in incident channel
- Update status page (if part of drill)
- Ask for help when stuck
- Document actions in timeline

❌ **Don't**:
- Try to "cheat" by investigating before drill
- Skip steps to save time
- Assume facilitator will give hints
- Give up if stuck (escalate instead)
- Ignore safety procedures

---

### Observer Guidelines

**What to Observe**:

- Time to detect
- Time to declare incident
- Communication clarity and frequency
- Escalation appropriateness
- Use of runbooks and documentation
- Hypothesis generation and testing
- Decision-making process
- Role clarity (IC, responders, etc.)

**How to Observe**:

- Take detailed notes
- Don't interrupt or help
- Log timestamps for key events
- Record good practices
- Note gaps or confusion
- Document action items

---

## Evaluation and Learning

### Debrief Structure

**1. Recap** (10 minutes)
- Facilitator reviews timeline
- Recap key decisions and actions
- Clarify any confusion

**2. What Went Well** (15 minutes)
- Celebrate successes
- Note good practices
- Reinforce positive behaviors

**3. What Went Poorly** (20 minutes)
- Identify gaps and issues
- Discuss what was confusing
- Surface tool/process problems

**4. Lessons Learned** (10 minutes)
- Key takeaways
- Surprises or unexpected findings
- Process improvements needed

**5. Action Items** (15 minutes)
- Concrete improvements
- Assign owners and deadlines
- Prioritize action items

---

### Evaluation Metrics

**Response Metrics**:
- Time to detect: [actual] vs [target]
- Time to declare: [actual] vs [target]
- Time to escalate: [actual] vs [target]
- Time to resolve: [actual] vs [target]

**Process Metrics**:
- Procedures followed: Y/N for each step
- Runbooks used: Which ones, were they helpful?
- Tools used: What worked, what didn't?
- Communication quality: Clear, timely, accurate?

**Team Metrics**:
- Roles clear: Did everyone know their role?
- Coordination smooth: How well did team work together?
- Escalation appropriate: Right people engaged at right time?
- Documentation quality: Timeline complete and accurate?

---

### Action Item Template

| Issue Identified | Action Item | Owner | Priority | Due Date |
|------------------|-------------|-------|----------|----------|
| [What went wrong] | [How to fix] | [Who] | [P0-P3] | [Date] |

**Example**:

| Issue Identified | Action Item | Owner | Priority | Due Date |
|------------------|-------------|-------|----------|----------|
| Rollback procedure unclear | Document rollback procedures for irreversible migrations | Alice | P1 | 2026-02-15 |
| On-call didn't know DB tooling | Create DB troubleshooting runbook | Carol | P0 | 2026-02-07 |
| Status page update delayed | Add status page to incident checklist | Bob | P2 | 2026-02-10 |

---

## Drill Schedule

### Quarterly Schedule (Recommended)

**Q1**: Database failure drill (tabletop)
**Q2**: Security incident drill (controlled chaos)
**Q3**: Multi-region failover drill (tabletop)
**Q4**: Complete outage drill (controlled chaos)

### Monthly Schedule

**Week 1**: Tabletop exercise (1 hour)
**Week 2**: Wheel of misfortune (1 hour)
**Week 3**: Controlled chaos drill (2-4 hours)
**Week 4**: Drill retrospective and planning

---

### Annual Calendar

| Month | Drill Type | Scenario | Duration |
|-------|-----------|----------|----------|
| Jan | Tabletop | API Outage | 1 hour |
| Feb | Controlled Chaos | Database Corruption | 3 hours |
| Mar | Wheel of Misfortune | Past Incidents Review | 1 hour |
| Apr | Tabletop | Security Incident | 2 hours |
| May | Controlled Chaos | Third-Party Outage | 2 hours |
| Jun | Production Gameday | Multi-Region Failover | 4 hours |
| Jul | Wheel of Misfortune | Past Incidents Review | 1 hour |
| Aug | Tabletop | Gradual Degradation | 2 hours |
| Sep | Controlled Chaos | Multi-Team Incident | 3 hours |
| Oct | Wheel of Misfortune | Past Incidents Review | 1 hour |
| Nov | Tabletop | Data Loss | 2 hours |
| Dec | Controlled Chaos | Complete Outage | 3 hours |

---

## Example Drills

### Example 1: Database Connection Pool Exhaustion

**Type**: Controlled Chaos

**Duration**: 2 hours

**Setup**:
1. Configure staging API with small connection pool (20 connections)
2. Set up load generator to gradually increase traffic
3. Configure monitoring to alert on connection pool utilization

**Scenario Timeline**:

| Time | Event |
|------|-------|
| 00:00 | Drill start announced |
| 00:05 | Load generator starts at 50 RPS |
| 00:10 | Traffic increases to 100 RPS |
| 00:15 | Connection pool reaches 80% (alert fires) |
| 00:20 | Traffic increases to 150 RPS |
| 00:22 | Connection pool exhausted, errors start |
| 00:23 | Error rate alert fires |
| [Participants respond] | |
| 00:45 | Complication: Rollback blocked |
| [Participants adapt] | |
| ~01:00 | Expected resolution |
| 01:15 | Debrief starts |

**Evaluation Points**:
- Did alert fire at correct threshold?
- Was incident declared promptly?
- Was root cause identified quickly?
- Was mitigation appropriate?
- Was communication clear?

**Expected Learnings**:
- Practice connection pool troubleshooting
- Understand when rollback isn't possible
- Practice alternative mitigation strategies

---

### Example 2: Security Incident Tabletop

**Type**: Tabletop Exercise

**Duration**: 90 minutes

**Setup**:
- Conference room or Zoom call
- Facilitator with scenario notes
- Participants with laptops for reference

**Scenario**:
> "At 2:30 AM, you receive an alert: 'Unusual API access patterns detected from IP 198.51.100.42'. The alert shows 10,000 requests in 5 minutes to the /api/users endpoint. What do you do?"

**Facilitator Injects** (based on participant responses):

**Response 1**: "Check the logs"
→ "Logs show the requests are using valid API keys from 20 different user accounts"

**Response 2**: "Block the IP address"
→ "New requests appear from different IP addresses"

**Response 3**: "Revoke the API keys"
→ "How do you determine which keys to revoke? Do you notify the users?"

**Response 4**: "Escalate to security team"
→ "Security team asks: Have you preserved logs? Have you isolated affected systems?"

**Continue scenario** through:
- Investigation
- Containment
- Legal notification decision
- Customer communication
- Recovery
- Postmortem planning

**Debrief Questions**:
- What went well in your response?
- What would you do differently?
- What procedures/runbooks would help?
- What tools are missing?
- Who should be notified in this scenario?

---

## Drill Best Practices

### For Facilitators

✅ **Do**:
- Keep scenario realistic
- Let participants struggle (within reason)
- Inject complications gradually
- Document everything
- Stay neutral (don't give answers)
- Adapt based on participant level

❌ **Don't**:
- Make scenario impossible
- Give up and provide solutions
- Skip debrief
- Criticize participants
- Rush through scenario

---

### For Participants

✅ **Do**:
- Take it seriously
- Follow real procedures
- Communicate clearly
- Ask for help when needed
- Document your actions
- Learn from mistakes

❌ **Don't**:
- Try to "win" the drill
- Take shortcuts
- Ignore safety rules
- Get frustrated
- Skip documentation

---

### For Leadership

✅ **Do**:
- Support drill program
- Provide time for drills
- Participate when appropriate
- Act on action items
- Celebrate learning

❌ **Don't**:
- Use drills for performance reviews
- Cancel drills due to "busy" schedule
- Ignore drill findings
- Blame individuals for drill outcomes

---

## References

- [Incident Response Process](INCIDENT_RESPONSE.md)
- [On-Call Rotation](ON_CALL_ROTATION.md)
- [Escalation Procedures](ESCALATION_PROCEDURES.md)
- [Google SRE Book - Testing for Reliability](https://sre.google/sre-book/testing-reliability/)
- [Wheel of Misfortune](https://dastergon.gr/wheel-of-misfortune/)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-30
**Version**: 1.0.0
**Next Review**: 2026-04-30
