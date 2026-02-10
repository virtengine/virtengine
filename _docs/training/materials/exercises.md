# Training Exercises

**Version:** 1.0.0  
**Date:** 2026-01-31

---

## Overview

This document catalogs all exercises available in the VirtEngine operator training program. Exercises are organized by training track and difficulty level.

---

## Exercise Categories

| Category | Description | Count |
|----------|-------------|-------|
| Core Concepts | Foundation exercises for all operators | 15 |
| Validator | Validator-specific exercises | 20 |
| Provider | Provider daemon exercises | 18 |
| Security | Security-focused exercises | 16 |
| Incident Response | Incident handling practice | 12 |
| Capstone | Comprehensive assessments | 4 |

---

## Core Concepts Exercises

### Architecture Understanding

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CORE-EX-001 | Module Dependency Mapping | 30 min | Easy |
| CORE-EX-002 | Transaction Flow Tracing | 45 min | Medium |
| CORE-EX-003 | State Machine Analysis | 60 min | Medium |
| CORE-EX-004 | Keeper Pattern Implementation | 90 min | Hard |

#### CORE-EX-001: Module Dependency Mapping

**Objective:** Create a dependency diagram of VirtEngine modules.

**Instructions:**
1. Review the `x/` directory structure
2. Identify dependencies between modules
3. Create a visual dependency graph
4. Document critical dependency paths

**Deliverable:** Dependency diagram with annotations

**Grading Criteria:**
- All modules identified (25%)
- Dependencies correctly mapped (35%)
- Critical paths identified (25%)
- Clear documentation (15%)

---

### Security Fundamentals

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CORE-EX-005 | Key Generation Practice | 20 min | Easy |
| CORE-EX-006 | Encryption Envelope Analysis | 45 min | Medium |
| CORE-EX-007 | Signature Verification | 30 min | Medium |
| CORE-EX-008 | Security Configuration Audit | 60 min | Medium |

#### CORE-EX-006: Encryption Envelope Analysis

**Objective:** Analyze the structure of a VEID encryption envelope.

**Instructions:**
1. Review `x/encryption/types/envelope.go`
2. Identify all envelope fields
3. Explain the purpose of each field
4. Document the encryption flow

**Deliverable:** Written analysis with diagrams

---

### Incident Response

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CORE-EX-009 | Severity Classification Practice | 20 min | Easy |
| CORE-EX-010 | Status Update Writing | 30 min | Easy |
| CORE-EX-011 | Timeline Construction | 45 min | Medium |
| CORE-EX-012 | Runbook Navigation | 30 min | Easy |

---

## Validator Exercises

### Setup and Configuration

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| VAL-EX-001 | Node Initialization | 45 min | Easy |
| VAL-EX-002 | Configuration Optimization | 60 min | Medium |
| VAL-EX-003 | Sentry Architecture Setup | 90 min | Hard |
| VAL-EX-004 | State Sync Configuration | 45 min | Medium |

#### VAL-EX-001: Node Initialization

**Objective:** Initialize a validator node from scratch.

**Instructions:**
```bash
# 1. Initialize node with custom moniker
virtengine init "MyTrainingValidator" --chain-id training-1

# 2. Generate validator key
virtengine keys add validator --keyring-backend test

# 3. Configure genesis
# Add your validator account to genesis

# 4. Verify configuration
virtengine validate-genesis
```

**Verification:**
- [ ] Node initialized successfully
- [ ] Keys generated and backed up
- [ ] Genesis validates
- [ ] Node can start

---

### Consensus Operations

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| VAL-EX-005 | Block Production Analysis | 45 min | Medium |
| VAL-EX-006 | Voting Power Calculation | 30 min | Easy |
| VAL-EX-007 | Slashing Conditions Review | 45 min | Medium |
| VAL-EX-008 | Consensus Troubleshooting | 60 min | Hard |

---

### VEID Operations

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| VAL-EX-009 | Identity Key Setup | 30 min | Easy |
| VAL-EX-010 | ML Model Verification | 45 min | Medium |
| VAL-EX-011 | Determinism Testing | 60 min | Hard |
| VAL-EX-012 | Scoring Pipeline Analysis | 90 min | Hard |

#### VAL-EX-011: Determinism Testing

**Objective:** Verify ML scoring determinism.

**Instructions:**
1. Run scoring on identical input 10 times
2. Compare all outputs
3. Document any variance
4. Identify determinism configuration

**Expected:** All 10 outputs must be identical

---

### Key Management

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| VAL-EX-013 | Key Backup Procedure | 30 min | Easy |
| VAL-EX-014 | Key Rotation Practice | 60 min | Medium |
| VAL-EX-015 | Recovery Simulation | 45 min | Medium |
| VAL-EX-016 | HSM Configuration | 90 min | Hard |

---

### Troubleshooting

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| VAL-EX-017 | Sync Issue Diagnosis | 45 min | Medium |
| VAL-EX-018 | Peer Connection Issues | 30 min | Medium |
| VAL-EX-019 | Resource Exhaustion | 45 min | Medium |
| VAL-EX-020 | Log Analysis Practice | 60 min | Medium |

---

## Provider Exercises

### Daemon Setup

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| PROV-EX-001 | Daemon Installation | 45 min | Easy |
| PROV-EX-002 | Configuration Setup | 30 min | Easy |
| PROV-EX-003 | Key Manager Setup | 45 min | Medium |
| PROV-EX-004 | Connectivity Verification | 30 min | Easy |

---

### Infrastructure Adapters

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| PROV-EX-005 | Kubernetes Adapter Config | 60 min | Medium |
| PROV-EX-006 | Namespace Isolation Setup | 45 min | Medium |
| PROV-EX-007 | Resource Quota Configuration | 30 min | Easy |
| PROV-EX-008 | Multi-Adapter Setup | 90 min | Hard |

#### PROV-EX-005: Kubernetes Adapter Configuration

**Objective:** Configure the Kubernetes adapter for workload deployment.

**Instructions:**
1. Create kubeconfig with limited permissions
2. Configure adapter in provider config
3. Deploy test workload
4. Verify lifecycle management

**Verification:**
- [ ] Adapter connects to K8s
- [ ] Test workload deploys
- [ ] Lifecycle transitions work
- [ ] Cleanup successful

---

### Marketplace Operations

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| PROV-EX-009 | Pricing Strategy Design | 45 min | Medium |
| PROV-EX-010 | Bid Filter Configuration | 30 min | Easy |
| PROV-EX-011 | Escrow Monitoring | 30 min | Easy |
| PROV-EX-012 | Settlement Verification | 45 min | Medium |

---

### Operations

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| PROV-EX-013 | Monitoring Dashboard Setup | 60 min | Medium |
| PROV-EX-014 | Alert Configuration | 45 min | Medium |
| PROV-EX-015 | Capacity Planning | 60 min | Medium |
| PROV-EX-016 | Usage Report Analysis | 30 min | Easy |
| PROV-EX-017 | Workload Troubleshooting | 60 min | Hard |
| PROV-EX-018 | Adapter Failover Testing | 90 min | Hard |

---

## Security Exercises

### Configuration Auditing

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| SEC-EX-001 | Permission Audit | 30 min | Easy |
| SEC-EX-002 | Network Configuration Audit | 45 min | Medium |
| SEC-EX-003 | Secrets Audit | 45 min | Medium |
| SEC-EX-004 | Comprehensive Security Audit | 90 min | Hard |

---

### Threat Modeling

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| SEC-EX-005 | Attack Surface Mapping | 60 min | Medium |
| SEC-EX-006 | STRIDE Analysis | 60 min | Medium |
| SEC-EX-007 | Risk Assessment | 45 min | Medium |
| SEC-EX-008 | Mitigation Planning | 45 min | Medium |

#### SEC-EX-006: STRIDE Analysis

**Objective:** Perform STRIDE analysis on validator infrastructure.

**STRIDE Categories:**
- **S**poofing - Identity attacks
- **T**ampering - Data modification
- **R**epudiation - Deniability
- **I**nformation Disclosure - Data leaks
- **D**enial of Service - Availability
- **E**levation of Privilege - Authorization

**Instructions:**
1. Map validator components
2. Identify threats for each STRIDE category
3. Rate risk (High/Medium/Low)
4. Propose mitigations

---

### Incident Response

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| SEC-EX-009 | Key Compromise Drill | 60 min | Hard |
| SEC-EX-010 | Intrusion Detection | 45 min | Medium |
| SEC-EX-011 | Evidence Collection | 45 min | Medium |
| SEC-EX-012 | Incident Report Writing | 30 min | Easy |

---

### Advanced Security

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| SEC-EX-013 | Penetration Test Planning | 60 min | Hard |
| SEC-EX-014 | Vulnerability Assessment | 90 min | Hard |
| SEC-EX-015 | Security Hardening | 60 min | Medium |
| SEC-EX-016 | Compliance Checklist | 45 min | Medium |

---

## Incident Response Exercises

### Detection and Triage

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| IR-EX-001 | Alert Interpretation | 30 min | Easy |
| IR-EX-002 | Severity Assessment | 20 min | Easy |
| IR-EX-003 | Initial Response Actions | 30 min | Medium |

---

### Response and Resolution

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| IR-EX-004 | Troubleshooting Methodology | 45 min | Medium |
| IR-EX-005 | Rollback Procedures | 30 min | Medium |
| IR-EX-006 | Communication Practice | 30 min | Easy |
| IR-EX-007 | Coordination Exercise | 60 min | Hard |

---

### Post-Incident

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| IR-EX-008 | Timeline Reconstruction | 45 min | Medium |
| IR-EX-009 | 5 Whys Analysis | 30 min | Easy |
| IR-EX-010 | Postmortem Writing | 60 min | Medium |
| IR-EX-011 | Action Item Definition | 30 min | Easy |
| IR-EX-012 | Metrics Analysis | 45 min | Medium |

---

## Capstone Exercises

### Validator Capstone

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CAP-001 | Full Validator Deployment | 4 hours | Hard |

**Scenario:** Deploy and operate a validator through a complete lifecycle.

**Tasks:**
1. Initialize and configure validator
2. Join testnet consensus
3. Set up monitoring and alerting
4. Perform key rotation
5. Handle simulated incident
6. Document operations

---

### Provider Capstone

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CAP-002 | Full Provider Deployment | 4 hours | Hard |

**Scenario:** Deploy and operate a provider through marketplace lifecycle.

**Tasks:**
1. Set up provider daemon
2. Configure Kubernetes adapter
3. Set pricing and start bidding
4. Win lease and deploy workload
5. Monitor usage and settlement
6. Handle simulated failure

---

### Security Capstone

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CAP-003 | Security Assessment | 3 hours | Hard |

**Scenario:** Complete security assessment of infrastructure.

**Tasks:**
1. Perform comprehensive audit
2. Identify vulnerabilities
3. Create threat model
4. Develop mitigation plan
5. Simulate security incident
6. Write assessment report

---

### Master Operator Capstone

| ID | Title | Duration | Difficulty |
|----|-------|----------|------------|
| CAP-004 | Complex Multi-System Scenario | 6 hours | Expert |

**Scenario:** Manage complex incident affecting multiple systems.

**Tasks:**
1. Detect multi-system degradation
2. Coordinate response across teams
3. Implement mitigations
4. Communicate to stakeholders
5. Resolve root cause
6. Complete postmortem

---

## Exercise Submission

### Submission Process

1. Complete exercise in lab environment
2. Document your work
3. Submit via training portal
4. Receive feedback within 48 hours

### Grading Criteria

| Aspect | Weight |
|--------|--------|
| Correctness | 40% |
| Methodology | 25% |
| Documentation | 20% |
| Completeness | 15% |

### Passing Scores

| Difficulty | Passing Score |
|------------|---------------|
| Easy | 70% |
| Medium | 75% |
| Hard | 80% |
| Expert | 85% |

---

## Contact

**Exercise Support:**
- Email: training@virtengine.com
- Slack: #operator-training
- Office Hours: Tue/Thu 2-4 PM UTC

---

**Document Owner**: Training Team  
**Last Updated**: 2026-01-31  
**Version**: 1.0.0
