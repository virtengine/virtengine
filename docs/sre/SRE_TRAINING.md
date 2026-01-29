# SRE Training and Education

## Overview

This document outlines the training program for Site Reliability Engineering at VirtEngine. Effective SRE practices require continuous learning and skill development across the organization.

## Table of Contents
1. [Training Philosophy](#training-philosophy)
2. [On-Call Engineer Training](#on-call-engineer-training)
3. [Incident Commander Training](#incident-commander-training)
4. [Developer SRE Training](#developer-sre-training)
5. [Continuous Education](#continuous-education)
6. [Certification and Advancement](#certification-and-advancement)

---

## Training Philosophy

### Core Principles

1. **Learn by Doing**: Hands-on practice beats theory
2. **Safe to Fail**: Training environment allows mistakes
3. **Continuous Learning**: SRE practices evolve constantly
4. **Cross-Functional**: Everyone learns SRE basics
5. **Blameless**: Mistakes are learning opportunities

### Training Outcomes

**Technical Skills**:
- System architecture understanding
- Debugging and troubleshooting
- Monitoring and observability
- Automation and tooling

**Operational Skills**:
- Incident response
- Communication under pressure
- Decision making
- Prioritization

**Cultural Skills**:
- Blameless postmortems
- Data-driven decisions
- Collaboration
- Continuous improvement

---

## On-Call Engineer Training

### Prerequisites

**Required**:
- [ ] 3+ months with VirtEngine
- [ ] Familiarity with VirtEngine architecture
- [ ] Access to all production systems
- [ ] Linux/Unix proficiency
- [ ] Git and deployment tools knowledge

**Recommended**:
- [ ] Programming experience (Go preferred)
- [ ] Kubernetes familiarity
- [ ] Prometheus/Grafana experience
- [ ] Previous on-call experience

---

### Week 1: Foundations

#### Day 1: SRE Philosophy and Practices

**Duration**: 4 hours

**Topics**:
- What is SRE?
- SRE vs DevOps vs Traditional Ops
- SLIs, SLOs, and Error Budgets
- Toil and Automation

**Activities**:
- Read Google SRE Book Chapter 1-3
- Review [VirtEngine SLI/SLO/SLA Framework](SLI_SLO_SLA.md)
- Calculate error budgets for example scenarios

**Assessment**: Quiz on SRE fundamentals

---

#### Day 2: VirtEngine Architecture Deep Dive

**Duration**: 6 hours

**Topics**:
- Blockchain node architecture
- Provider daemon components
- API service design
- Database architecture
- Network topology

**Activities**:
- Architecture walkthrough presentation
- Service dependency mapping exercise
- Deploy local development environment
- Run all services locally

**Assessment**: Architecture diagram creation

---

#### Day 3: Monitoring and Observability

**Duration**: 6 hours

**Topics**:
- Prometheus metrics
- Grafana dashboards
- Alerting rules
- Log aggregation
- Distributed tracing

**Activities**:
- Explore Grafana dashboards
- Write custom PromQL queries
- Create test alert
- Analyze sample traces in Jaeger
- Practice log searching in Kibana

**Assessment**: Create dashboard for new metric

---

#### Day 4: Incident Response Basics

**Duration**: 6 hours

**Topics**:
- Incident lifecycle
- Severity classification
- Communication protocols
- Troubleshooting methodology
- Postmortem process

**Activities**:
- Read [Incident Response Process](INCIDENT_RESPONSE.md)
- Review past postmortems
- Practice incident communication
- Tabletop incident exercise

**Assessment**: Mock incident response

---

#### Day 5: Tools and Access

**Duration**: 4 hours

**Topics**:
- PagerDuty configuration
- Slack incident channels
- Deployment tools
- Debug tools
- Runbook navigation

**Activities**:
- Set up PagerDuty alerts
- Join all necessary Slack channels
- Test SSH access to all systems
- Run practice deployment
- Walk through all runbooks

**Assessment**: Access verification checklist

---

### Week 2: Hands-On Practice

#### Days 6-8: Shadow Shifts (24 hours)

**Objective**: Observe experienced on-call engineer

**Activities**:
- Shadow on-call engineer for full shift
- Observe alert handling
- Participate in troubleshooting (guided)
- Ask questions about decisions made
- Document learnings

**Requirements**:
- Shadow at least 2 different engineers
- Cover both daytime and off-hours
- Minimum 24 hours total shadow time

**Deliverable**: Shadow shift notes and questions

---

#### Days 9-10: Simulated Incidents

**Objective**: Practice incident response in safe environment

**Scenarios**:

**Scenario 1: High Error Rate** (2 hours)
- Simulated: API error rate spike
- Task: Investigate and resolve
- Skills: Metrics analysis, log searching, troubleshooting

**Scenario 2: Deployment Failure** (2 hours)
- Simulated: Bad deployment causing errors
- Task: Rollback and verify
- Skills: Deployment tools, verification

**Scenario 3: Database Issues** (2 hours)
- Simulated: Database connection pool exhaustion
- Task: Identify and mitigate
- Skills: Database debugging, configuration

**Scenario 4: Network Problems** (2 hours)
- Simulated: Network latency spike
- Task: Diagnose and resolve
- Skills: Network troubleshooting

**Assessment**: Successfully resolve 3/4 scenarios

---

### Week 3: Supervised On-Call

**Objective**: Handle real alerts with backup support

**Setup**:
- Trainee receives all alerts
- Experienced engineer on standby (shadow support)
- Trainee handles incidents with guidance
- Debrief after each incident

**Duration**: One full week (7 days)

**Support**:
- Shadow engineer available 24/7
- Can take over if needed
- Provides guidance but lets trainee lead

**Deliverable**: Handled incidents log

---

### Certification

**Requirements**:
- [ ] Complete all training modules
- [ ] Pass all assessments
- [ ] Successfully shadow 24+ hours
- [ ] Resolve simulated incidents
- [ ] Complete supervised shift
- [ ] Sign off from 2 senior engineers

**Certification**: "VirtEngine On-Call Engineer Certified"

**Next Steps**: Join on-call rotation

---

## Incident Commander Training

### Prerequisites

**Required**:
- [ ] 6+ months as on-call engineer
- [ ] Responded to 5+ incidents
- [ ] Deep technical knowledge
- [ ] Strong communication skills
- [ ] Completed on-call engineer training

**Recommended**:
- [ ] Previous leadership experience
- [ ] Experience with crisis management
- [ ] Public speaking skills

---

### Module 1: Incident Command Fundamentals

**Duration**: 4 hours

**Topics**:
- Role of Incident Commander
- Authority and responsibility
- Decision making under pressure
- Delegation and coordination
- Communication strategies

**Activities**:
- Read incident command case studies
- Watch recorded incidents
- Analyze IC decision points
- Practice communication scripts

**Assessment**: Written exam on IC role

---

### Module 2: Incident Command Practice

**Duration**: 8 hours (4 sessions × 2 hours)

**Simulated Incidents**:

**Session 1: SEV-2 API Outage**
- Role: IC
- Team: 3 responders (simulated)
- Duration: 1.5 hours
- Focus: Communication and coordination

**Session 2: SEV-1 Database Corruption**
- Role: IC
- Team: 5 responders (simulated)
- Duration: 2 hours
- Focus: High-pressure decision making

**Session 3: SEV-2 Multi-Service Degradation**
- Role: IC
- Team: 4 responders (simulated)
- Duration: 1.5 hours
- Focus: Complexity management

**Session 4: SEV-1 Security Incident**
- Role: IC
- Team: 6 responders (simulated)
- Duration: 2 hours
- Focus: Stakeholder communication

**Assessment**: Evaluated by senior IC

---

### Module 3: Shadow Incident Command

**Duration**: 3 incidents minimum

**Objective**: Observe experienced IC

**Activities**:
- Shadow IC during real incidents
- Participate in decision discussions
- Assist with coordination tasks
- Debrief after incident

**Requirements**:
- Shadow at least 3 real incidents
- Mix of severities (SEV-1, SEV-2, SEV-3)
- Written reflection on each

**Deliverable**: IC shadow notes

---

### Module 4: Supervised Incident Command

**Duration**: 2 incidents minimum

**Objective**: Lead incident with IC backup

**Setup**:
- Trainee acts as IC
- Experienced IC shadows
- Shadow can intervene if needed
- Full debrief after resolution

**Requirements**:
- Lead at least 1 SEV-3
- Successfully manage incident to resolution
- Positive feedback from shadow IC

**Deliverable**: Incident leadership report

---

### Certification

**Requirements**:
- [ ] Complete all IC modules
- [ ] Pass all assessments
- [ ] Shadow 3+ incidents as IC trainee
- [ ] Lead 2+ incidents under supervision
- [ ] Recommendation from 2 senior ICs
- [ ] Approval from SRE Lead

**Certification**: "VirtEngine Incident Commander Certified"

**Next Steps**: Join IC rotation

---

## Developer SRE Training

### Objective

Enable developers to build reliable services and participate in SRE practices.

### Module 1: SRE for Developers (2 hours)

**Topics**:
- Why SRE matters for developers
- SLIs and SLOs for your service
- Error budgets and trade-offs
- Observability best practices
- Performance budgets

**Activities**:
- Define SLIs for a feature
- Calculate error budget impact
- Review observability code

**Deliverable**: SLI proposal for your service

---

### Module 2: Instrumentation Workshop (4 hours)

**Topics**:
- Prometheus metrics patterns
- Structured logging
- Distributed tracing
- Error tracking
- Performance profiling

**Activities**:
- Add metrics to sample code
- Implement structured logging
- Add trace spans
- Create custom dashboard

**Deliverable**: Instrumented service example

---

### Module 3: Reliability Patterns (4 hours)

**Topics**:
- Circuit breakers
- Retries and backoff
- Timeouts and deadlines
- Graceful degradation
- Rate limiting

**Activities**:
- Implement circuit breaker
- Add retry logic with backoff
- Test failure scenarios
- Code review of reliability patterns

**Deliverable**: Reliability patterns in code

---

### Module 4: Load Testing (2 hours)

**Topics**:
- Load testing fundamentals
- VirtEngine load testing framework
- Writing load tests
- Analyzing results
- Performance budgets

**Activities**:
- Write load test for API endpoint
- Run load test and analyze results
- Identify performance bottlenecks

**Deliverable**: Load test for your service

---

### Module 5: On-Call Lite (2 hours)

**Topics**:
- Incident response basics
- Runbook creation
- Debugging production issues
- Rollback procedures
- Blameless postmortems

**Activities**:
- Write runbook for your service
- Practice rollback
- Review past incidents

**Deliverable**: Runbook for your service

---

### Certification

**Requirements**:
- [ ] Complete all developer modules
- [ ] Instrument a service
- [ ] Create load tests
- [ ] Write runbook

**Certification**: "VirtEngine Developer SRE Trained"

**Benefits**:
- Can participate in on-call (optional)
- Invited to SRE sync meetings
- Access to SRE tools and dashboards

---

## Continuous Education

### Monthly Learning Sessions

**Format**: 1-hour lunch & learn

**Topics** (Rotating):
- SRE industry trends
- New tools and technologies
- Case studies from other companies
- Postmortem learning sessions
- Advanced troubleshooting techniques

**Frequency**: Last Friday of each month

---

### Quarterly Workshops

**Format**: Half-day hands-on workshops

**Q1**: Chaos Engineering
- Introduction to chaos engineering
- Chaos Mesh hands-on
- Design chaos experiments
- Run experiments in staging

**Q2**: Advanced Monitoring
- PromQL deep dive
- Custom metrics and exporters
- Advanced Grafana techniques
- Anomaly detection

**Q3**: Performance Optimization
- Profiling techniques
- Database optimization
- Caching strategies
- Load testing at scale

**Q4**: Incident Response Excellence
- Advanced IC techniques
- Crisis communication
- Complex incident scenarios
- Gameday retrospective

---

### External Training

**Budget**: $2,000/engineer/year for external training

**Recommended Courses**:
- [Google Cloud SRE Certification](https://cloud.google.com/certification/cloud-sre)
- [Linux Foundation Kubernetes Administrator](https://training.linuxfoundation.org/certification/certified-kubernetes-administrator-cka/)
- [HashiCorp Terraform Certification](https://www.hashicorp.com/certification/terraform-associate)

**Recommended Conferences**:
- SREcon (USENIX)
- KubeCon + CloudNativeCon
- VelocityConf

**Process**:
1. Submit training request to SRE Lead
2. Justification: How does this benefit VirtEngine?
3. Approval: Manager + Budget approval
4. Post-training: Share learnings with team

---

### Book Club

**Format**: Bi-weekly 30-minute discussions

**Current Book**: Site Reliability Engineering (Google)

**Schedule**:
- Week 1-2: Chapters 1-4 (Intro, SLOs, Monitoring)
- Week 3-4: Chapters 5-8 (Toil, Incident Response)
- Week 5-6: Chapters 9-12 (On-Call, Postmortems)
- Week 7-8: Chapters 13-16 (Capacity, Testing)

**Future Books**:
- The Site Reliability Workbook
- Database Reliability Engineering
- Seeking SRE
- The Phoenix Project

---

## Certification and Advancement

### Career Ladder

**Level 1: On-Call Engineer**
- Certified on-call
- Responds to incidents
- Executes runbooks
- Participates in postmortems

**Level 2: Senior On-Call Engineer**
- 1+ year experience
- IC certified
- Leads SEV-3 incidents
- Creates runbooks
- Mentors junior engineers

**Level 3: Incident Commander**
- 2+ years experience
- Leads SEV-1/SEV-2 incidents
- Designs reliability improvements
- Contributes to SRE strategy

**Level 4: SRE Lead**
- 3+ years experience
- Owns SRE program
- Sets SLOs and policies
- Manages SRE team
- Strategic planning

---

### Skill Matrix

| Skill | L1 | L2 | L3 | L4 |
|-------|----|----|----|----|
| **Incident Response** | Execute | Lead SEV-3 | Lead SEV-1/2 | Program Owner |
| **Monitoring** | Read | Create | Design | Strategy |
| **Automation** | Execute | Create | Architect | Program |
| **Communication** | Status Updates | Stakeholders | Leadership | Executive |
| **Architecture** | Understand | Design | Lead | Strategy |
| **Mentoring** | N/A | 1-2 engineers | 3-5 engineers | Team |

---

### Advancement Criteria

**L1 → L2**:
- [ ] 1+ year as on-call
- [ ] Responded to 20+ incidents
- [ ] Created 5+ runbooks
- [ ] Led 3+ SEV-3 incidents
- [ ] Completed IC training
- [ ] Mentored 1+ junior engineer

**L2 → L3**:
- [ ] 2+ years total SRE experience
- [ ] IC certified
- [ ] Led 10+ SEV-2 incidents
- [ ] Led 2+ SEV-1 incidents
- [ ] Contributed to SRE strategy
- [ ] Mentored 2+ engineers

**L3 → L4**:
- [ ] 3+ years total SRE experience
- [ ] Demonstrated leadership
- [ ] Owned major SRE initiatives
- [ ] Industry recognition (talks, articles)
- [ ] Team management skills

---

## Training Resources

### Internal

**Documentation**:
- [SRE Documentation](README.md)
- [Runbooks](runbooks/)
- [Past Postmortems](postmortems/)
- [Architecture Docs](../architecture/)

**Code**:
- Observability package: `pkg/observability/`
- SRE tools: `pkg/sre/`
- Load tests: `tests/load/`

**Recordings**:
- Past incident reviews
- Training sessions
- Gameday exercises

---

### External

**Books**:
- [Site Reliability Engineering](https://sre.google/sre-book/) (Free online)
- [The Site Reliability Workbook](https://sre.google/workbook/)
- [Database Reliability Engineering](https://www.oreilly.com/library/view/database-reliability-engineering/9781491925935/)

**Online Courses**:
- [Coursera: Site Reliability Engineering](https://www.coursera.org/learn/site-reliability-engineering-slos)
- [LinkedIn Learning: SRE Foundations](https://www.linkedin.com/learning/site-reliability-engineering-foundations)
- [Udemy: Kubernetes for SRE](https://www.udemy.com/topic/kubernetes/)

**Communities**:
- [SRE Weekly Newsletter](https://sreweekly.com/)
- [r/sre subreddit](https://reddit.com/r/sre)
- [CNCF Slack #sre](https://slack.cncf.io/)

**Podcasts**:
- [On-Call Nightmares Podcast](https://www.oncallnightmares.com/)
- [SRE Path](https://srepath.com/)

---

## Training Metrics

### Success Metrics

**Completion Rate**:
```
Completion Rate = Completed Training / Started Training

Target: > 90%
```

**Time to Proficiency**:
```
Time to Proficiency = Date of Certification - Start Date

Target: < 4 weeks (on-call), < 8 weeks (IC)
```

**Retention Rate**:
```
Retention Rate = Stayed in Role After 1 Year / Total Certified

Target: > 80%
```

**Incident Performance**:
```
Trainee MTTR vs Team Average MTTR

Target: Within 20% of team average within 3 months
```

---

### Feedback and Improvement

**Training Feedback Survey** (After each module):
- Content quality (1-5)
- Instructor effectiveness (1-5)
- Hands-on practice (1-5)
- Suggestions for improvement

**Quarterly Training Review**:
- Analyze completion rates
- Review feedback
- Identify gaps
- Update curriculum

**Annual Curriculum Refresh**:
- Industry trend analysis
- Technology updates
- Best practice evolution
- Stakeholder input

---

## Getting Started

### New Hire SRE Training

**Week 1**: General VirtEngine onboarding
**Week 2-4**: On-call engineer training (this program)
**Week 5**: Supervised on-call shift
**Week 6**: Join on-call rotation

### Existing Engineer → On-Call

**Week 1**: Days 1-5 of on-call training
**Week 2**: Days 6-10 (shadow and simulation)
**Week 3**: Supervised shift
**Week 4**: Join rotation

### Developer SRE Training

**Self-paced**: Complete 5 modules at your own pace
**Target**: 2-4 weeks
**Support**: Office hours with SRE team (Tuesdays, Thursdays)

---

## Contact

**Questions?**
- Slack: #sre-training
- Email: sre-training@virtengine.com
- SRE Lead: [Name]

**Training Coordinator**: [Name]

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
