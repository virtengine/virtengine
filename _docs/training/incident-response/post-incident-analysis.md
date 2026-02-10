# Post-Incident Analysis Training

**Module Duration:** 4 hours  
**Prerequisites:** Incident Response Training  
**Track:** Incident Response

---

## Learning Objectives

By the end of this module, you will be able to:

- [ ] Facilitate blameless postmortem meetings
- [ ] Construct accurate incident timelines
- [ ] Apply root cause analysis techniques
- [ ] Write effective postmortem documents
- [ ] Identify and track action items
- [ ] Extract organizational learning from incidents

---

## Table of Contents

1. [Blameless Culture](#blameless-culture)
2. [Timeline Construction](#timeline-construction)
3. [Root Cause Analysis](#root-cause-analysis)
4. [Postmortem Document](#postmortem-document)
5. [Action Items](#action-items)
6. [Learning and Metrics](#learning-and-metrics)
7. [Exercises](#exercises)

---

## Blameless Culture

### Why Blameless?

```
┌─────────────────────────────────────────────────────────────────┐
│                 Blame vs. Blameless                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Blame Culture                   Blameless Culture             │
│   ─────────────────               ─────────────────             │
│   • "Who made this mistake?"     • "What allowed this?"        │
│   • Fear of reporting            • Safe to report              │
│   • Hide failures                • Share failures              │
│   • Punish individuals           • Improve systems             │
│   • Same mistakes repeat         • Learn and prevent           │
│                                                                 │
│   Result: Problems hidden,       Result: Problems surfaced,    │
│           incidents repeat               systems improve       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Blameless Principles

1. **Assume good intentions**: Everyone was trying to do the right thing
2. **Focus on systems**: People make mistakes; systems should prevent impact
3. **Seek understanding**: Why did actions seem reasonable at the time?
4. **Encourage reporting**: More reports = more learning opportunities
5. **Celebrate learning**: Treat postmortems as improvement opportunities

### Facilitation Guidelines

**DO:**
- Use "we" language, not "you"
- Ask "what" and "how" questions
- Focus on timeline facts
- Encourage all perspectives
- Thank people for honesty

**DON'T:**
- Ask "why did you do that?"
- Assign individual blame
- Use loaded language ("failure", "mistake")
- Dismiss contributing factors
- Rush the process

---

## Timeline Construction

### Importance of Timelines

A good timeline:
- Establishes facts objectively
- Shows cause and effect
- Reveals decision points
- Identifies delays
- Provides learning opportunities

### Timeline Template

```
┌─────────────────────────────────────────────────────────────────┐
│                    Incident Timeline                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Time (UTC)    Event                            Source         │
│   ──────────    ─────                            ──────         │
│   14:23:01      Deployment started               Deploy log     │
│   14:25:33      Error rate began increasing      Prometheus     │
│   14:27:00      Alert fired (ErrorRateHigh)      PagerDuty      │
│   14:28:15      On-call acknowledged             PagerDuty      │
│   14:30:00      Started investigation            Slack          │
│   14:35:42      Identified deployment as cause   Slack          │
│   14:37:00      Initiated rollback               Deploy log     │
│   14:39:15      Rollback completed               Deploy log     │
│   14:42:00      Error rate normalized            Prometheus     │
│   14:45:00      Incident resolved                PagerDuty      │
│                                                                 │
│   Key Metrics:                                                  │
│   • Time to Detect (TTD): 4 minutes                            │
│   • Time to Acknowledge: 5 minutes                             │
│   • Time to Mitigate: 14 minutes                               │
│   • Time to Resolve: 22 minutes                                │
│   • Total Duration: 22 minutes                                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Data Sources

| Source | Information Type |
|--------|-----------------|
| Monitoring (Prometheus) | Metric changes, alert times |
| Logs (Kibana) | Error messages, stack traces |
| Deployment logs | Release times, changes |
| PagerDuty | Alert and acknowledgment times |
| Slack | Communication, decisions |
| Git | Code changes |
| Calendar | Meetings, planned activities |

### Timeline Best Practices

1. **Be precise**: Use exact timestamps (UTC)
2. **Be factual**: Record what happened, not interpretations
3. **Be complete**: Include both automated and human actions
4. **Cite sources**: Link to evidence for each event
5. **Note gaps**: Identify periods with no data

---

## Root Cause Analysis

### The 5 Whys Technique

Start with the problem and ask "why" repeatedly:

```
┌─────────────────────────────────────────────────────────────────┐
│                      5 Whys Example                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Problem: VEID scoring returned incorrect results              │
│                                                                 │
│   Why #1: The ML model produced different scores on validators  │
│           └─ Because different validators had different models  │
│                                                                 │
│   Why #2: Why did validators have different models?             │
│           └─ Some validators didn't download the upgrade        │
│                                                                 │
│   Why #3: Why didn't they download the upgrade?                │
│           └─ The notification email went to spam                │
│                                                                 │
│   Why #4: Why did it go to spam?                               │
│           └─ Email was sent from a new domain without SPF      │
│                                                                 │
│   Why #5: Why wasn't SPF configured?                           │
│           └─ New domain setup process doesn't include SPF       │
│                                                                 │
│   Root Cause: Domain provisioning process lacks email config   │
│   Action: Add SPF/DKIM/DMARC to domain setup checklist         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Fishbone (Ishikawa) Diagram

```
                                    ┌─────────────────┐
                                    │    PROBLEM:     │
          ┌─────────────────────────│  Service        │
          │                         │  Outage         │
          │                         └────────┬────────┘
          │                                  │
    People│           Process                │        Technology
    ──────┴───────    ───────                │        ──────────
          │                │                 │              │
    ┌─────┴─────┐    ┌─────┴─────┐          │        ┌─────┴─────┐
    │ Training  │    │ No rollback│          │        │ Memory    │
    │ gap       │    │ procedure  │          │        │ leak      │
    └───────────┘    └───────────┘          │        └───────────┘
                                             │
    ┌───────────┐    ┌───────────┐          │        ┌───────────┐
    │ Fatigue   │    │ Rushed    │          │        │ No        │
    │           │    │ deploy    │          │        │ canary    │
    └───────────┘    └───────────┘          │        └───────────┘
          │                │                 │              │
    ──────┴───────    ─────┴─────           │        ──────┴─────
    Environment       Communication         │        Tools
```

### Contributing Factors

Root causes rarely stand alone. Identify contributing factors:

| Category | Questions to Ask |
|----------|-----------------|
| **Process** | Was the process followed? Was it adequate? |
| **Technology** | What failed? Was it detected? |
| **People** | Was training adequate? Were they fatigued? |
| **Environment** | What external factors contributed? |
| **Communication** | Was information shared effectively? |

### Avoid These Mistakes

| Mistake | Example | Better Approach |
|---------|---------|-----------------|
| **Stopping too early** | "Operator error" | Ask why the error was possible |
| **Counterfactuals** | "If only we had..." | Focus on what actually happened |
| **Hindsight bias** | "Should have known" | What information was available? |
| **Single cause** | "The bug caused it" | What allowed the bug to have impact? |

---

## Postmortem Document

### Postmortem Template

```markdown
# Postmortem: [Incident Title]

**Date**: YYYY-MM-DD
**Authors**: [Names]
**Status**: Draft | Review | Final
**Severity**: SEV-[1/2/3/4]

## Executive Summary
[2-3 sentence summary of what happened and impact]

## Impact
- **Duration**: X hours Y minutes
- **Users Affected**: N users / N% of traffic
- **Revenue Impact**: $X (if applicable)
- **SLO Impact**: X% of error budget consumed

## Timeline
| Time (UTC) | Event |
|------------|-------|
| HH:MM | [Event description] |

## Root Cause
[Detailed explanation of the root cause]

## Contributing Factors
1. [Factor 1]
2. [Factor 2]

## Resolution
[What was done to resolve the incident]

## What Went Well
- [Positive aspect 1]
- [Positive aspect 2]

## What Could Be Improved
- [Improvement area 1]
- [Improvement area 2]

## Action Items
| Action | Owner | Priority | Due Date | Status |
|--------|-------|----------|----------|--------|
| [Action] | [Name] | P[1/2/3] | YYYY-MM-DD | Open |

## Lessons Learned
[Key takeaways for the organization]

## References
- [Link to incident Slack thread]
- [Link to relevant dashboards]
- [Link to related postmortems]
```

### Writing Tips

1. **Be specific**: "Error rate increased from 0.1% to 15%" not "Errors went up"
2. **Quantify impact**: Use numbers wherever possible
3. **Show your work**: Explain how conclusions were reached
4. **Stay factual**: Avoid emotional language
5. **Be actionable**: Every problem should have an action item

### Review Process

```
┌─────────────────────────────────────────────────────────────────┐
│                 Postmortem Review Process                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Day 0-1              Day 2-3              Day 4-5             │
│   ───────              ───────              ───────             │
│   Draft                Review               Finalize            │
│                                                                 │
│   ┌─────────┐         ┌─────────┐         ┌─────────┐          │
│   │ Author  │────────▶│ Team    │────────▶│ Publish │          │
│   │ writes  │         │ reviews │         │ & share │          │
│   │ draft   │         │ & adds  │         │         │          │
│   └─────────┘         └─────────┘         └─────────┘          │
│                                                                 │
│   Deadline: Postmortem complete within 5 business days         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Action Items

### Effective Action Items

Good action items are:
- **Specific**: Clear what needs to be done
- **Measurable**: How do we know it's complete?
- **Assigned**: Single owner responsible
- **Time-bound**: Has a due date
- **Prioritized**: P1/P2/P3

### Action Item Examples

| ❌ Bad | ✅ Good |
|--------|---------|
| "Improve monitoring" | "Add alert for error rate > 5% on VEID scoring" |
| "Better documentation" | "Update runbook with new rollback procedure" |
| "More testing" | "Add integration test for deployment validation" |
| "Fix the bug" | "Patch memory leak in bid engine (JIRA-123)" |

### Action Item Tracking

```
┌─────────────────────────────────────────────────────────────────┐
│                 Action Item Lifecycle                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Created      Assigned      In Progress     Verified           │
│   ───────      ────────      ───────────     ────────           │
│      │             │              │              │              │
│      ▼             ▼              ▼              ▼              │
│   ┌─────┐      ┌─────┐       ┌─────┐        ┌─────┐            │
│   │ New │─────▶│ Open│──────▶│ WIP │───────▶│Done │            │
│   └─────┘      └─────┘       └─────┘        └─────┘            │
│                                                                 │
│   Tracking: All action items tracked in JIRA with              │
│             "postmortem" label and link to postmortem          │
│                                                                 │
│   Review: Weekly review of open postmortem actions             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Priority Guidelines

| Priority | Definition | SLA |
|----------|------------|-----|
| **P1** | Prevents incident recurrence | 1 week |
| **P2** | Significantly reduces risk | 2 weeks |
| **P3** | Improves detection/response | 1 month |

---

## Learning and Metrics

### Incident Metrics

Track these metrics over time:

| Metric | Definition | Target |
|--------|------------|--------|
| **MTTD** | Mean Time to Detect | < 5 minutes |
| **MTTA** | Mean Time to Acknowledge | < 10 minutes |
| **MTTM** | Mean Time to Mitigate | < 30 minutes |
| **MTTR** | Mean Time to Resolve | < 2 hours |
| **MTBF** | Mean Time Between Failures | Increasing |

### Postmortem Metrics

| Metric | Target |
|--------|--------|
| Postmortem completion rate | 100% for SEV-1/2 |
| Time to complete postmortem | < 5 business days |
| Action item completion rate | > 90% on time |
| Repeat incidents | < 5% |

### Learning Dissemination

```
┌─────────────────────────────────────────────────────────────────┐
│                 Learning Dissemination                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   Immediate (Day of):                                           │
│   • Post summary in #engineering                                │
│   • Update relevant runbooks                                    │
│                                                                 │
│   Short-term (Week of):                                         │
│   • Complete and publish postmortem                             │
│   • Review in team meeting                                      │
│   • Update training materials if needed                         │
│                                                                 │
│   Long-term (Monthly):                                          │
│   • Aggregate learning in monthly review                        │
│   • Update architecture documentation                           │
│   • Incorporate into onboarding                                 │
│                                                                 │
│   Quarterly:                                                    │
│   • Trend analysis presentation                                 │
│   • System reliability review                                   │
│   • Refresher training updates                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Exercises

### Exercise 1: 5 Whys Analysis

**Scenario**: A validator was slashed for missing blocks.

Perform 5 Whys analysis starting with:
"The validator missed 100 consecutive blocks"

<details>
<summary>Example Solution</summary>

1. Why missed blocks? → Node was offline
2. Why offline? → Out of memory crash
3. Why OOM? → Memory leak in VEID scoring
4. Why leak? → Not caught in testing
5. Why not caught? → No memory leak tests

Root Cause: Lack of memory leak testing
Action: Add memory profiling to CI/CD pipeline

</details>

### Exercise 2: Write a Postmortem

**Scenario**: Provider daemon stopped bidding for 2 hours due to expired TLS certificate.

Write the following sections:
- Executive Summary
- Impact
- Root Cause
- Action Items (at least 3)

### Exercise 3: Critique Postmortem

Review this action item and improve it:

> "We should probably add more monitoring to prevent this from happening again."

<details>
<summary>Improved Version</summary>

**Original Issue**: No alert for TLS certificate expiration

**Action Items**:
1. Add Prometheus alert for cert expiry < 30 days (Owner: SRE, Due: 2 days, P1)
2. Document cert renewal procedure in runbook (Owner: Ops, Due: 1 week, P2)
3. Automate cert renewal with cert-manager (Owner: Infra, Due: 2 weeks, P2)

</details>

---

## Key Takeaways

1. **Blameless culture** enables learning from failures
2. **Accurate timelines** establish facts objectively
3. **Root cause analysis** goes beyond surface symptoms
4. **Postmortems** document learning for the organization
5. **Action items** must be SMART (Specific, Measurable, Assigned, Realistic, Time-bound)
6. **Track metrics** to measure improvement over time

---

## Next Steps

- Review recent postmortems in `docs/postmortems/`
- Practice facilitating a postmortem meeting
- Complete [Incident Simulation Lab](../labs/incident-simulation-lab.md)

---

**Document Owner**: Training Team  
**Last Updated**: 2026-01-31  
**Version**: 1.0.0
