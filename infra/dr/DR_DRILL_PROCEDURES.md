# VirtEngine Disaster Recovery Drills

This document provides procedures for conducting DR drills to validate recovery capabilities and ensure team readiness.

## Table of Contents

1. [Overview](#overview)
2. [Drill Types](#drill-types)
3. [Quarterly Drill Schedule](#quarterly-drill-schedule)
4. [Tabletop Exercises](#tabletop-exercises)
5. [Automated Test Drills](#automated-test-drills)
6. [Full Failover Drills](#full-failover-drills)
7. [Post-Drill Procedures](#post-drill-procedures)

---

## Overview

### Purpose

DR drills ensure:
- Recovery procedures are tested and validated
- Team members are familiar with their roles
- RTO/RPO targets are achievable
- Documentation is accurate and complete
- Issues are identified and remediated

### Drill Philosophy

| Principle | Description |
|-----------|-------------|
| **Progressive Complexity** | Start with tabletop exercises, progress to full drills |
| **Scheduled and Unscheduled** | Mix planned drills with surprise tests |
| **Documented Learning** | Every drill produces actionable improvements |
| **No Blame** | Focus on process improvement, not individual mistakes |
| **Realistic Scenarios** | Test actual failure conditions |

### Drill Frequency

| Drill Type | Frequency | Duration | Participants |
|------------|-----------|----------|--------------|
| Tabletop | Monthly | 1-2 hours | SRE team, management |
| Automated Tests | Daily | 15-30 min | Automated |
| Partial Failover | Quarterly | 2-4 hours | Full SRE team |
| Full Failover | Semi-annual | 4-8 hours | All technical staff |
| Surprise Test | Quarterly | Varies | On-call engineers |

---

## Drill Types

### 1. Tabletop Exercise

**Objective:** Validate understanding of procedures without executing them.

**Format:**
- Facilitator presents disaster scenario
- Team walks through response procedures
- Discussion of decision points and alternatives
- Identification of gaps in documentation or training

**Sample Scenarios:**
1. **Region Outage:** "US-East-1 experiences complete network partition"
2. **Database Corruption:** "CockroachDB reports data inconsistency in replication"
3. **Key Compromise:** "Validator key potentially exposed in compromised instance"
4. **Cascading Failure:** "Single validator failure triggers cluster-wide instability"

**Duration:** 60-90 minutes

**Outcome:** Updated runbooks, identified training needs, improved procedures

---

### 2. Automated Test Drills

**Objective:** Continuous validation of DR infrastructure and procedures.

**Execution:**
```bash
# Daily automated DR tests
cd /opt/virtengine
./scripts/dr/dr-test.sh --test all --report --notify
```

**Tests Performed:**
- Backup integrity verification
- Backup age checks
- Cross-region connectivity
- DNS health
- S3 access validation
- Secrets Manager access
- State sync endpoint availability

**Monitoring:**
- Tests run via Kubernetes CronJob daily at 06:00 UTC
- Results published to S3: `s3://virtengine-dr-test-results/`
- Failures trigger Slack/PagerDuty alerts
- Prometheus metrics: `dr_test_failures`, `dr_test_duration_seconds`

**Success Criteria:**
- All tests pass
- Execution time < 10 minutes
- No manual intervention required

---

### 3. Partial Failover Drill (Quarterly)

**Objective:** Test failover procedures in non-production or isolated environment.

**Pre-Drill Preparation (T-7 days):**
1. Schedule drill date/time with stakeholders
2. Notify user-facing teams (no customer impact expected)
3. Review and update runbooks
4. Ensure backup verification is current
5. Verify monitoring dashboards functional

**Execution Steps:**

#### Phase 1: Pre-Checks (T-30 min)

```bash
# Verify all regions healthy
for REGION in us-east-1 eu-west-1 ap-southeast-1; do
  kubectl --context=virtengine-prod-${REGION} cluster-info
  curl -sf https://rpc-${REGION}.virtengine.io/status | jq '.result.sync_info'
done

# Verify backup freshness
./scripts/dr/dr-test.sh --test backup

# Record baseline metrics
curl -sf https://prometheus.virtengine.io/api/v1/query?query=up | jq '.data.result'
```

#### Phase 2: Simulated Failure (T+0)

**Scenario:** Primary region (us-east-1) becomes unavailable

```bash
# Simulate region failure by blocking traffic
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch file://disable-us-east-1.json

# Record failover start time
START_TIME=$(date +%s)
echo $START_TIME > /tmp/failover_start_time
```

#### Phase 3: Execute Failover (T+5 min)

```bash
# Follow regional-failover runbook
cd /opt/virtengine/infra/dr/runbooks
ansible-playbook regional-failover.yaml \
  --extra-vars "failing_region=us-east-1 target_region=eu-west-1"
```

**Key Steps (from runbook):**
1. Verify backup freshness in target region (< 5 min RPO)
2. Verify target region cluster health
3. Verify database replication caught up
4. Scale validators in target region
5. Update DNS to point to target region
6. Enable health checks for target region

#### Phase 4: Validation (T+15 min)

```bash
# Verify API accessibility
curl -sf https://api.virtengine.io/cosmos/base/tendermint/v1beta1/node_info

# Verify block production
HEIGHT1=$(curl -sf https://rpc.virtengine.io/status | jq -r '.result.sync_info.latest_block_height')
sleep 10
HEIGHT2=$(curl -sf https://rpc.virtengine.io/status | jq -r '.result.sync_info.latest_block_height')
[ "$HEIGHT2" -gt "$HEIGHT1" ] && echo "Block production: OK"

# Verify database writes
kubectl --context=virtengine-prod-eu-west-1 -n cockroachdb exec cockroachdb-0 -- \
  cockroach sql --certs-dir=/cockroach/cockroach-certs \
  -e "CREATE TABLE IF NOT EXISTS system.dr_test (id INT PRIMARY KEY, ts TIMESTAMP DEFAULT now()); INSERT INTO system.dr_test VALUES (1) ON CONFLICT (id) DO UPDATE SET ts = now();"
```

#### Phase 5: Rollback (T+30 min)

```bash
# Restore primary region
aws route53 change-resource-record-sets \
  --hosted-zone-id Z1234567890ABC \
  --change-batch file://restore-us-east-1.json

# Re-balance validators across regions
kubectl --context=virtengine-prod-us-east-1 -n virtengine scale deployment virtengine-validator --replicas=3
kubectl --context=virtengine-prod-eu-west-1 -n virtengine scale deployment virtengine-validator --replicas=2
```

#### Phase 6: Post-Drill Analysis (T+60 min)

Calculate RTO:
```bash
END_TIME=$(date +%s)
START_TIME=$(cat /tmp/failover_start_time)
RTO=$((END_TIME - START_TIME))
echo "Achieved RTO: ${RTO} seconds (target: 900 seconds)"
```

**Success Criteria:**
- RTO < 15 minutes (900 seconds)
- All services accessible from target region
- Block production continues
- Database writes successful
- No data loss (RPO = 0)

---

### 4. Full Failover Drill (Semi-Annual)

**Objective:** Test complete regional failover with actual user traffic.

**Risk Mitigation:**
- Execute during maintenance window
- Pre-announce to users 48 hours in advance
- Have rollback plan ready
- Keep primary region in standby (not destroyed)

**Additional Considerations:**
1. **Customer Communication:**
   - Announce maintenance window 48 hours prior
   - Provide status page updates
   - Have support team on standby

2. **Extended Validation:**
   - Monitor for 2 hours post-failover
   - Test all user-facing features
   - Verify provider daemon functionality
   - Check escrow settlement
   - Validate identity verification

3. **Performance Validation:**
   - Compare latency metrics (primary vs secondary)
   - Monitor database replication lag
   - Check API response times
   - Verify autoscaling behavior

**Documentation:**
- Record all commands executed
- Screenshot key metrics
- Time-stamp each phase
- Capture any issues encountered

---

### 5. Surprise Drill (Quarterly)

**Objective:** Test on-call engineer response without advance notice.

**Execution:**
- DR coordinator triggers simulated outage
- On-call engineer receives PagerDuty alert
- Engineer follows incident response procedures
- Coordinator observes and documents response

**Evaluation Criteria:**
- Time to acknowledge alert (< 5 minutes)
- Time to begin investigation (< 10 minutes)
- Correct procedure identification (within 15 minutes)
- Communication with stakeholders (immediate)
- Escalation decisions (as needed)

**Debrief:**
- Review timeline with engineer
- Identify knowledge gaps
- Update documentation/training
- No negative consequences (learning exercise)

---

## Post-Drill Procedures

### Immediate Actions (Within 1 hour)

1. **Document Results:**
   ```bash
   # Generate drill report
   cat > /tmp/drill_report_$(date +%Y%m%d).md << EOF
   # DR Drill Report - $(date +%F)
   
   ## Scenario
   [Describe scenario tested]
   
   ## Timeline
   - Start: [timestamp]
   - Detection: [timestamp]
   - Response: [timestamp]
   - Resolution: [timestamp]
   - RTO Achieved: [duration]
   
   ## Success Metrics
   - RTO: [actual vs target]
   - RPO: [actual vs target]
   - All systems operational: [yes/no]
   
   ## Issues Encountered
   [List any issues]
   
   ## Action Items
   [List follow-up tasks]
   EOF
   
   # Upload to S3
   aws s3 cp /tmp/drill_report_$(date +%Y%m%d).md \
     s3://virtengine-dr-test-results/drills/
   ```

2. **Update Metrics:**
   ```bash
   # Record drill metrics in Prometheus
   curl -X POST http://pushgateway.virtengine.io:9091/metrics/job/dr_drill \
     --data-binary @- << EOF
   dr_drill_rto_seconds $(cat /tmp/failover_duration)
   dr_drill_success 1
   dr_drill_timestamp $(date +%s)
   EOF
   ```

3. **Restore Normal Operations:**
   - Verify all regions in correct state
   - Re-enable any disabled monitoring
   - Update status page
   - Notify stakeholders of completion

### Follow-Up Actions (Within 1 week)

1. **Team Debrief Meeting:**
   - Review what went well
   - Discuss challenges encountered
   - Identify improvement opportunities
   - Assign action items

2. **Update Documentation:**
   - Revise runbooks based on learnings
   - Update timing estimates
   - Add missing steps
   - Clarify ambiguous procedures

3. **Training Updates:**
   - Create training materials for identified gaps
   - Schedule knowledge sharing sessions
   - Update onboarding materials

4. **Process Improvements:**
   - Automate manual steps where possible
   - Improve monitoring/alerting
   - Enhance tooling
   - Update escalation procedures

### Quarterly Review (Every 3 months)

1. **Trend Analysis:**
   - Compare RTO across drills
   - Identify recurring issues
   - Track improvement over time
   - Update targets if consistently met

2. **Compliance Verification:**
   - Ensure drill frequency met
   - Verify documentation complete
   - Audit action item completion
   - Report to stakeholders

3. **Plan Next Quarter:**
   - Schedule upcoming drills
   - Select scenarios for tabletop exercises
   - Assign facilitators/coordinators
   - Update calendar

---

## Drill Scenarios Library

### Scenario 1: Complete Region Failure
**Description:** Primary region experiences complete outage (power, network, or AWS service failure)
**Components Affected:** All services in region
**Expected RTO:** 15 minutes
**Expected RPO:** 0 (continuous replication)

### Scenario 2: Database Replication Failure
**Description:** CockroachDB replication stops between regions
**Components Affected:** Database writes may be lost if primary region fails
**Expected RTO:** 5 minutes (promote secondary)
**Expected RPO:** Up to 5 minutes (last successful replication)

### Scenario 3: DNS/Load Balancer Failure
**Description:** Global DNS or load balancer becomes unavailable
**Components Affected:** Client access to API/RPC endpoints
**Expected RTO:** 10 minutes (update DNS)
**Expected RPO:** 0 (no data loss)

### Scenario 4: Key Compromise
**Description:** Validator private key potentially exposed
**Components Affected:** Single validator security
**Expected RTO:** 15 minutes (rotate keys, restart validator)
**Expected RPO:** 0 (keys backed up)

### Scenario 5: Data Corruption
**Description:** Software bug causes state corruption
**Components Affected:** Blockchain state integrity
**Expected RTO:** 1 hour (restore from backup)
**Expected RPO:** Up to 4 hours (last backup)

### Scenario 6: Multi-Region Network Partition
**Description:** Cross-region connectivity lost
**Components Affected:** Validator consensus, database replication
**Expected RTO:** 30 minutes (wait for partition heal or manual intervention)
**Expected RPO:** 0 (eventual consistency when partition resolves)

---

## Metrics and Reporting

### Key Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Drill Frequency | 100% of scheduled drills completed | Count vs plan |
| RTO Achievement | 90% of drills meet RTO target | Actual time vs target |
| RPO Achievement | 100% of drills meet RPO target | Data loss vs target |
| Issue Resolution | 100% of action items closed within 30 days | Open vs closed |
| Team Participation | 100% of SRE team participates annually | Individual records |

### Quarterly Report Template

```markdown
# DR Drill Quarterly Report - Q[N] [YEAR]

## Executive Summary
- Drills Conducted: [count]
- Average RTO Achieved: [duration]
- RPO Compliance: [percentage]
- Critical Issues Found: [count]
- Action Items Closed: [percentage]

## Drills Conducted
[Table of drill dates, types, and outcomes]

## RTO/RPO Trends
[Chart showing RTO over time]

## Issues and Resolutions
[List of issues found and remediation status]

## Recommendations
[Process improvements, resource needs, etc.]

## Next Quarter Plan
[Scheduled drills and scenarios]
```

---

## Emergency Contact List

| Role | Primary | Secondary | Phone | Slack |
|------|---------|-----------|-------|-------|
| **DR Coordinator** | [Name] | [Name] | [Phone] | @handle |
| **SRE Lead** | [Name] | [Name] | [Phone] | @handle |
| **Platform Engineer** | [Name] | [Name] | [Phone] | @handle |
| **CTO** | [Name] | [Name] | [Phone] | @handle |
| **Security Lead** | [Name] | [Name] | [Phone] | @handle |

**Escalation Path:**
1. On-call Engineer (immediate)
2. SRE Lead (< 15 minutes)
3. Platform Engineer / DR Coordinator (< 30 minutes)
4. CTO (< 1 hour or if unresolved)

**Communication Channels:**
- **Incident Channel:** `#incident-response`
- **DR Channel:** `#dr-drills`
- **Status Page:** https://status.virtengine.io
- **PagerDuty:** https://virtengine.pagerduty.com

---

## Appendices

### Appendix A: Pre-Drill Checklist

- [ ] Drill date/time scheduled and communicated
- [ ] Stakeholders notified
- [ ] Backup verification current (< 24 hours)
- [ ] Monitoring dashboards verified functional
- [ ] Runbooks reviewed and updated
- [ ] Team roles assigned
- [ ] Rollback plan documented
- [ ] Communication templates prepared

### Appendix B: During-Drill Checklist

- [ ] Start time recorded
- [ ] All steps documented
- [ ] Screenshots captured
- [ ] Issues logged
- [ ] Decisions documented with rationale
- [ ] Metrics collected
- [ ] Communication sent to stakeholders

### Appendix C: Post-Drill Checklist

- [ ] End time recorded
- [ ] RTO/RPO calculated
- [ ] Drill report created
- [ ] Metrics updated in Prometheus
- [ ] Normal operations restored
- [ ] Stakeholders notified of completion
- [ ] Debrief meeting scheduled
- [ ] Action items created and assigned
