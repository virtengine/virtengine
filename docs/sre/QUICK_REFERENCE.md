# SRE Quick Reference Guide

One-page reference for common SRE tasks and information.

---

## 🚨 Emergency Contacts

| Role | Contact | When to Use |
|------|---------|-------------|
| **On-Call** | PagerDuty | SEV-1/SEV-2 incidents |
| **IC (Incident Commander)** | PagerDuty Escalation | SEV-1 incidents |
| **SRE Lead** | Slack: #sre-team | Non-urgent questions |
| **Security Team** | security@virtengine.com | Security incidents |

---

## 🎯 Key SLO Targets

| Service | Metric | Target |
|---------|--------|--------|
| **Blockchain Node** | Uptime | 99.95% |
| | TX Confirmation P95 | < 10s |
| | Throughput | ≥ 50 TPS |
| **Provider Daemon** | Uptime | 99.90% |
| | Provisioning P95 | < 300s |
| | Deployment Success | ≥ 99% |
| **API Services** | Availability | 99.90% |
| | Query P95 | < 2s |
| | Error Rate | < 0.5% |

---

## 📊 Dashboards

| Dashboard | URL | Use For |
|-----------|-----|---------|
| **SRE Overview** | `grafana.virtengine.com/d/sre-overview` | General health |
| **Error Budgets** | `grafana.virtengine.com/d/error-budgets` | Budget status |
| **Capacity** | `grafana.virtengine.com/d/capacity` | Resource usage |
| **Incidents** | `grafana.virtengine.com/d/incidents` | Incident tracking |

---

## 🔴 Incident Response

### Severity Levels

| Severity | Definition | Response Time | Example |
|----------|-----------|---------------|---------|
| **SEV-1** | Complete outage | < 5 min | API down for all users |
| **SEV-2** | Major degradation | < 15 min | 50% error rate |
| **SEV-3** | Minor issues | < 1 hour | Slow queries |
| **SEV-4** | Cosmetic | < 24 hours | UI formatting |

### Incident Checklist

**When Alert Fires**:
1. ⏱️ Acknowledge within 2 minutes
2. 🔍 Check dashboard for symptoms
3. 📱 Page additional help if SEV-1/SEV-2
4. 💬 Create #incident-YYYY-MM-DD-HHMM channel
5. 📢 Update status page
6. 🛠️ Begin investigation

**Quick Investigation**:
- Recent deployments? `kubectl rollout history`
- Metrics anomaly? Check Grafana
- Error spike? Check logs in Kibana
- Traffic spike? Check request rate

**Quick Fixes**:
- Rollback: `kubectl rollout undo deployment/X`
- Scale up: `kubectl scale deployment/X --replicas=N`
- Restart: `kubectl rollout restart deployment/X`

---

## 💰 Error Budget Status

### Budget Formula
```
Error Budget = (1 - SLO) × 28 days

Example: 99.90% SLO
= (1 - 0.9990) × 28 days
= 40.32 minutes/month
```

### Status Actions

| Status | Remaining | Action |
|--------|-----------|--------|
| 🟢 **Healthy** | > 50% | All changes allowed |
| 🟡 **Warning** | 25-50% | SRE approval for features |
| 🔴 **Critical** | 5-25% | Bug fixes only |
| ⚫ **Depleted** | < 5% | **CHANGE FREEZE** |

---

## 🔧 Common Commands

### Check Service Health
```bash
# Pod status
kubectl get pods -n virtengine

# Logs
kubectl logs -f deployment/virtengine-node -n virtengine

# Recent events
kubectl get events -n virtengine --sort-by='.lastTimestamp'
```

### Deployment Operations
```bash
# Rollout status
kubectl rollout status deployment/api-server

# History
kubectl rollout history deployment/api-server

# Rollback
kubectl rollout undo deployment/api-server

# Restart
kubectl rollout restart deployment/api-server
```

### Metrics Queries
```promql
# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Latency P95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error budget remaining
(1 - (error_count / total_requests)) / error_budget_target
```

---

## 📋 Runbooks Location

**Path**: `docs/sre/runbooks/`

Common Runbooks:
- `high-error-rate.md`
- `database-connection-pool.md`
- `deployment-failure.md`
- `disk-space-full.md`
- `memory-leak.md`

**Template**: `docs/sre/runbooks/template.md`

---

## 📝 Creating Postmortem

1. **Use Template**: `docs/sre/templates/postmortem_template.md`
2. **Timeline**: Document every action
3. **Root Cause**: Technical + process failures
4. **Action Items**: Assign owners and dates
5. **Review**: Within 48 hours
6. **Publish**: Within 5 days

**Blameless**: Focus on systems, not people!

---

## 🎓 Training Resources

| Role | Training Path | Duration |
|------|---------------|----------|
| **On-Call** | On-Call Engineer Training | 3 weeks |
| **IC** | Incident Commander Training | 4 modules |
| **Developer** | Developer SRE Training | 2-4 weeks |

**Documentation**: `docs/sre/SRE_TRAINING.md`

---

## 🤖 Automation Priorities

| Priority | Task | Savings |
|----------|------|---------|
| **P0** | Deployment pipeline | 40 hrs/month |
| **P0** | Alert auto-remediation | 40 hrs/month |
| **P0** | Config management | 24 hrs/month |
| **P1** | Certificate rotation | 24 hrs/year |
| **P1** | DB maintenance | 4 hrs/month |

---

## 📞 Escalation Path

```
Alert Fires
    ↓
On-Call Engineer
    ↓ (if SEV-1/SEV-2 or no response in 5 min)
Secondary On-Call
    ↓ (if SEV-1 or complex)
Incident Commander
    ↓ (if critical business impact)
VP Engineering
    ↓ (if major outage)
CTO/CEO
```

---

## 🔍 Troubleshooting Checklist

**The 5 Whys**:
1. What is the symptom?
2. What changed recently?
3. What does the data show?
4. What is the root cause?
5. How do we prevent recurrence?

**USE Method** (Resources):
- **U**tilization: How busy is it?
- **S**aturation: Is work queuing?
- **E**rrors: Are there errors?

**RED Method** (Services):
- **R**ate: Requests per second
- **E**rrors: Error rate
- **D**uration: Latency

---

## 📈 Key Metrics

| Metric | Formula | Target |
|--------|---------|--------|
| **MTTR** | Total downtime / # incidents | < 30 min |
| **MTBF** | Total uptime / # failures | > 24 hours |
| **TTD** | Alert time - incident start | < 5 min |
| **TTA** | Ack time - alert time | < 2 min |
| **Toil %** | Toil hours / total hours | < 50% |

---

## 🛡️ Error Budget Policy

### Decision Tree

```
What's the error budget status?

> 50% (Healthy)
    → Deploy freely
    → Innovation encouraged

25-50% (Warning)
    → Features need SRE approval
    → Proceed with caution

5-25% (Critical)
    → Bug fixes only
    → VP approval required

< 5% (Depleted)
    → CHANGE FREEZE
    → Emergency fixes only (CTO approval)
```

---

## 🎯 This Week's Goals

**Check Weekly in SRE Sync**:
- [ ] Any SEV-1/SEV-2 incidents?
- [ ] Error budgets healthy?
- [ ] Toil < 50%?
- [ ] Capacity concerns?
- [ ] Action items on track?

---

## 📚 Documentation Index

| Document | Use For |
|----------|---------|
| [README](README.md) | Overview and getting started |
| [SLI/SLO/SLA](SLI_SLO_SLA.md) | Service level targets |
| [Error Budget Policy](ERROR_BUDGET_POLICY.md) | Error budget rules |
| [Toil Management](TOIL_MANAGEMENT.md) | Automation priorities |
| [Capacity Planning](CAPACITY_PLANNING.md) | Resource forecasting |
| [Performance Budgets](PERFORMANCE_BUDGETS.md) | Performance targets |
| [Reliability Testing](RELIABILITY_TESTING.md) | Testing strategy |
| [Incident Response](INCIDENT_RESPONSE.md) | Incident procedures |
| [SRE Training](SRE_TRAINING.md) | Training programs |

---

## 💡 Quick Tips

**Performance**:
- Always check P95/P99, not just average
- "Fast on average" can hide bad tail latency

**Incidents**:
- Communicate early and often
- Status page first, then investigate
- Document everything (helps postmortem)

**Error Budgets**:
- Use them! Don't hoard budget
- Budget lets you innovate safely
- Depleted budget = focus on stability

**Toil**:
- If you do it > 3 times, automate it
- Toil compounds - address early
- Calculate ROI before automating

**Monitoring**:
- Alert on symptoms, not causes
- Alerts should be actionable
- Too many alerts = alert fatigue

---

**Questions?** → #sre-team on Slack

**Last Updated**: 2026-01-29
