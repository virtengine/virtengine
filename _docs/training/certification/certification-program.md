# VirtEngine Operator Certification Program

**Version:** 1.0.0  
**Date:** 2026-01-31

---

## Overview

The VirtEngine Operator Certification Program validates the knowledge and skills required to operate VirtEngine infrastructure. Certifications demonstrate competency in specific operational domains and are required for certain operational privileges.

---

## Available Certifications

| Certification | Code | Duration | Prerequisites |
|---------------|------|----------|---------------|
| Certified Validator Operator | CVO | 2 years | Validator Track completion |
| Certified Provider Operator | CPO | 2 years | Provider Track completion |
| Certified Security Specialist | CSS | 2 years | Security Track completion |
| Master Operator | MO | 2 years | All tracks + 1 year experience |

---

## Certification Requirements

### Certified Validator Operator (CVO)

**Purpose:** Demonstrates competency to operate VirtEngine validator nodes.

**Training Requirements:**
- [ ] Core Modules (12 hours)
  - [ ] Architecture Overview (4 hours)
  - [ ] Security Fundamentals (4 hours)
  - [ ] Incident Response Basics (4 hours)
- [ ] Validator Track (28 hours)
  - [ ] Validator Operator Training (16 hours)
  - [ ] VEID Operations (8 hours)
  - [ ] Validator Key Management (4 hours)
- [ ] Labs
  - [ ] Lab Environment Setup (2 hours)
  - [ ] Validator Operations Lab (4 hours)
  - [ ] Incident Simulation Lab (4 hours)

**Examination:**
- Written exam: 60 questions, 90 minutes, 80% passing
- Practical exam: 3 scenarios, 2 hours, pass all

**Experience Requirement:**
- Supervised validator operations for 2+ weeks
- OR 3+ months prior blockchain operations experience

### Certified Provider Operator (CPO)

**Purpose:** Demonstrates competency to operate VirtEngine provider infrastructure.

**Training Requirements:**
- [ ] Core Modules (12 hours)
  - [ ] Architecture Overview (4 hours)
  - [ ] Security Fundamentals (4 hours)
  - [ ] Incident Response Basics (4 hours)
- [ ] Provider Track (28 hours)
  - [ ] Provider Daemon Training (16 hours)
  - [ ] Infrastructure Adapters (8 hours)
  - [ ] Marketplace Operations (4 hours)
- [ ] Labs
  - [ ] Lab Environment Setup (2 hours)
  - [ ] Provider Operations Lab (4 hours)
  - [ ] Incident Simulation Lab (4 hours)

**Examination:**
- Written exam: 60 questions, 90 minutes, 80% passing
- Practical exam: 3 scenarios, 2 hours, pass all

**Experience Requirement:**
- Supervised provider operations for 2+ weeks
- OR 3+ months prior infrastructure operations experience

### Certified Security Specialist (CSS)

**Purpose:** Demonstrates competency in VirtEngine security operations.

**Training Requirements:**
- [ ] Core Modules (12 hours)
- [ ] Security Track (16 hours)
  - [ ] Security Best Practices (8 hours)
  - [ ] Threat Modeling (4 hours)
  - [ ] Security Incident Response (4 hours)
- [ ] Labs
  - [ ] Security Assessment Lab (4 hours)
  - [ ] Incident Simulation Lab (4 hours)

**Examination:**
- Written exam: 50 questions, 75 minutes, 85% passing
- Practical exam: Security assessment, 3 hours, pass

**Experience Requirement:**
- 6+ months security operations experience
- OR relevant security certification (CISSP, OSCP, etc.)

### Master Operator (MO)

**Purpose:** Demonstrates comprehensive VirtEngine operational expertise.

**Requirements:**
- [ ] Hold CVO certification
- [ ] Hold CPO certification
- [ ] Hold CSS certification (or equivalent experience)
- [ ] 1+ year operational experience
- [ ] Demonstrated leadership during SEV-1/2 incidents
- [ ] Contributed to operational documentation or training

**Examination:**
- Written exam: 100 questions, 150 minutes, 85% passing
- Practical exam: Complex multi-system scenario, 4 hours
- Panel interview with senior operators

---

## Examination Details

### Written Examination

**Format:**
- Multiple choice (60%)
- Scenario-based questions (30%)
- Short answer (10%)

**Topics by Certification:**

| Topic | CVO | CPO | CSS | MO |
|-------|-----|-----|-----|-----|
| Architecture | 20% | 20% | 15% | 15% |
| Security | 15% | 15% | 40% | 20% |
| Operations | 35% | 35% | 15% | 25% |
| Troubleshooting | 20% | 20% | 20% | 25% |
| Incident Response | 10% | 10% | 10% | 15% |

**Rules:**
- Closed book (no external resources)
- No electronic devices
- 90-150 minutes depending on certification
- One retake allowed after 2-week waiting period

### Practical Examination

**Format:**
- Hands-on scenarios in test environment
- Real-time problem solving
- Examiner observation

**Scoring Criteria:**
- Correct identification of issues (25%)
- Appropriate actions taken (25%)
- Effective troubleshooting methodology (20%)
- Communication and documentation (15%)
- Time management (15%)

**Sample Scenarios:**

| Certification | Scenario Examples |
|---------------|-------------------|
| CVO | Node sync failure, consensus issues, key rotation |
| CPO | Workload deployment failure, adapter configuration, bid pricing |
| CSS | Security audit, incident investigation, key compromise response |
| MO | Complex multi-system outage, security + operations hybrid |

---

## Certification Process

### Step 1: Complete Training

1. Register for training at training.virtengine.com
2. Complete all required modules
3. Complete all required labs
4. Obtain training completion certificate

### Step 2: Schedule Examination

1. Request exam scheduling via training portal
2. Verify eligibility (training + experience requirements)
3. Select exam date (available monthly)
4. Receive confirmation with exam instructions

### Step 3: Take Examinations

**Written Exam:**
- Online proctored or in-person
- Results within 48 hours

**Practical Exam:**
- Scheduled separately after passing written
- In-person or supervised remote
- Results within 1 week

### Step 4: Certification Award

Upon passing:
1. Receive digital certificate
2. Added to certified operator registry
3. Badge for LinkedIn/professional profiles
4. Access to certified operator resources

---

## Certification Maintenance

### Continuing Education Requirements

To maintain certification, complete annually:

| Certification | Annual CE Hours | Refresher Training |
|---------------|-----------------|-------------------|
| CVO | 16 hours | Quarterly (2 hours each) |
| CPO | 16 hours | Quarterly (2 hours each) |
| CSS | 20 hours | Quarterly (3 hours each) |
| MO | 24 hours | All tracks |

### Acceptable CE Activities

| Activity | CE Hours |
|----------|----------|
| Quarterly refresher training | 2-3 hours |
| Incident command participation | 2 hours per incident |
| Training material development | 4 hours per module |
| Conference presentations | 4 hours |
| Approved external training | Varies |

### Recertification

Certifications expire after 2 years. To recertify:

**Option 1: CE Completion**
- Complete all required CE hours
- Pass abbreviated exam (30 questions, 70% passing)

**Option 2: Full Re-examination**
- Take full written exam
- Practical exam waived if active operations experience

---

## Examination Preparation

### Study Materials

| Resource | Location |
|----------|----------|
| Training Modules | `_docs/training/` |
| Runbooks | `docs/operations/runbooks/` |
| Architecture Docs | `_docs/architecture/` |
| Practice Questions | `_docs/training/certification/practice/` |

### Preparation Checklist

**2 Weeks Before:**
- [ ] Complete all training modules
- [ ] Review all lab exercises
- [ ] Take practice exams
- [ ] Identify weak areas

**1 Week Before:**
- [ ] Review weak areas
- [ ] Practice hands-on scenarios
- [ ] Review runbooks
- [ ] Rest and prepare mentally

**Day Before:**
- [ ] Confirm exam logistics
- [ ] Prepare required materials
- [ ] Get adequate sleep

### Practice Exam Questions

#### Sample CVO Questions

1. **What is the correct procedure when a validator key may be compromised?**
   - a) Continue operating and monitor
   - b) Immediately stop the validator and rotate keys
   - c) Wait for confirmation from other validators
   - d) Restart the node

2. **Which file contains the validator's private signing key?**
   - a) config.toml
   - b) priv_validator_key.json
   - c) node_key.json
   - d) genesis.json

3. **What is the purpose of sentry nodes?**
   - a) Increase transaction throughput
   - b) Protect validator from DDoS and hide its IP
   - c) Store additional blockchain data
   - d) Process identity verification requests

#### Sample CPO Questions

1. **What state transition is NOT valid for a workload?**
   - a) Pending → Deploying
   - b) Running → Stopped
   - c) Stopped → Pending
   - d) Running → Paused

2. **Which metric is used for CPU usage billing?**
   - a) CPUCycles
   - b) CPUMilliSeconds
   - c) CPUPercent
   - d) CPUUnits

3. **What triggers automatic settlement in the escrow module?**
   - a) Every block
   - b) Every N blocks (configurable)
   - c) Only on lease close
   - d) Never (manual only)

#### Sample CSS Questions

1. **What encryption algorithm does VirtEngine use for identity envelopes?**
   - a) AES-256-GCM
   - b) X25519-XSalsa20-Poly1305
   - c) RSA-2048
   - d) ChaCha20-Poly1305

2. **How many signatures are required for VEID scope uploads?**
   - a) 1
   - b) 2
   - c) 3
   - d) 4

3. **What is the recommended file permission for validator keys?**
   - a) 644
   - b) 755
   - c) 600
   - d) 777

---

## Certification Benefits

### For Individuals

- Industry-recognized credential
- Career advancement opportunities
- Access to certified operator community
- Early access to new features and documentation
- Priority support channels

### For Organizations

- Verified operator competency
- Reduced operational risk
- Insurance and compliance requirements
- Partnership eligibility
- Listing in certified operators directory

---

## Policies

### Code of Conduct

Certified operators agree to:
- Maintain confidentiality of exam content
- Report security vulnerabilities responsibly
- Support fellow operators professionally
- Maintain accurate certification status

### Exam Integrity

Violations result in certification revocation:
- Cheating or sharing exam content
- Misrepresenting certification status
- Falsifying experience requirements

### Appeals Process

1. Submit appeal within 30 days of exam
2. Include specific concerns and evidence
3. Review by certification committee
4. Response within 2 weeks

---

## Contact

**Certification Questions:**
- Email: certification@virtengine.com
- Slack: #operator-certification

**Training Support:**
- Email: training@virtengine.com
- Slack: #operator-training

**Exam Scheduling:**
- Portal: training.virtengine.com/certification

---

**Document Owner**: Training Team  
**Last Updated**: 2026-01-31  
**Version**: 1.0.0
