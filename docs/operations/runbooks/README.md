# VirtEngine Operator Runbooks

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Overview

This directory contains comprehensive operational runbooks for VirtEngine production environments. These runbooks are designed for operators, SREs, and on-call engineers responsible for maintaining VirtEngine infrastructure.

## Runbook Index

### Setup and Maintenance

| Runbook | Description | Audience |
|---------|-------------|----------|
| [Validator Setup](VALIDATOR_SETUP.md) | Complete validator node setup and maintenance | Validators |
| [Provider Operations](PROVIDER_OPERATIONS.md) | Provider daemon setup and daily operations | Providers |
| [Performance Tuning](PERFORMANCE_TUNING.md) | System optimization and tuning guidelines | SRE/Operators |

### Incident Response

| Runbook | Description | Audience |
|---------|-------------|----------|
| [Incident Response](INCIDENT_RESPONSE.md) | Incident classification and response procedures | On-call |
| [Troubleshooting Guide](TROUBLESHOOTING.md) | Common issues and resolution steps | On-call/SRE |
| [On-Call Quick Reference](../ON_CALL_RUNBOOK.md) | Quick reference for on-call engineers | On-call |
| [On-Call Rotation](../../sre/ON_CALL_ROTATION.md) | Full on-call rotation setup and management | SRE/Management |

### Disaster Recovery

| Runbook | Description | Audience |
|---------|-------------|----------|
| [Disaster Recovery](DISASTER_RECOVERY.md) | DR procedures for various failure scenarios | SRE/Management |
| [Backup and Restore](BACKUP_RESTORE.md) | Backup strategies and restoration procedures | SRE/Operators |
| [Upgrade Procedures](UPGRADE_PROCEDURES.md) | Chain upgrades, hotfixes, and rollbacks | SRE/Validators |

## Quick Reference

### Severity Levels

| Level | Definition | Response Time | Escalation |
|-------|------------|---------------|------------|
| **SEV-1** | Complete service outage | 5 minutes | Immediate page |
| **SEV-2** | Major feature unavailable | 15 minutes | Page primary |
| **SEV-3** | Degraded performance | 1 hour | Slack notification |
| **SEV-4** | Minor issue | 24 hours | Ticket queue |

### Emergency Contacts

| Role | Contact | Availability |
|------|---------|--------------|
| Primary On-Call | PagerDuty | 24/7 |
| Secondary On-Call | PagerDuty escalation | 24/7 |
| Security Team | security@virtengine.com | 24/7 |
| Infrastructure Lead | Slack: #sre-escalation | Business hours |

### Critical Commands

```bash
# Check chain status
virtengine status

# Check node sync
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Check validator status
virtengine query staking validators --status bonded | head -50

# Check provider daemon health
curl -s http://localhost:8443/health

# Emergency: Stop validator
sudo systemctl stop virtengine

# Emergency: Pause provider
virtengine tx provider set-status --status PAUSED --from provider
```

### Critical Dashboards

| Dashboard | URL |
|-----------|-----|
| Chain Health | https://grafana.virtengine.com/d/chain-health |
| VEID Scoring | https://grafana.virtengine.com/d/veid-scoring |
| Marketplace | https://grafana.virtengine.com/d/marketplace |
| Provider Health | https://grafana.virtengine.com/d/provider-health |
| Error Budget | https://grafana.virtengine.com/d/error-budget |

## Using These Runbooks

### During an Incident

1. **Identify severity** using the severity matrix
2. **Find the relevant runbook** in the index above
3. **Follow the diagnosis and resolution steps** in order
4. **Escalate** if unable to resolve within expected timeframe
5. **Document actions** in the incident channel

### For Planned Maintenance

1. **Review the upgrade or maintenance runbook** at least 24 hours ahead
2. **Create a change ticket** with planned steps
3. **Schedule a maintenance window** and notify stakeholders
4. **Execute the runbook** following each step
5. **Verify success** using validation commands
6. **Document any deviations** from the planned procedure

### For Training

1. Use runbooks as training material for new team members
2. Conduct quarterly incident drills using these runbooks
3. Review and update runbooks after each incident

## Runbook Maintenance

### Update Schedule

- **Weekly**: Review for accuracy after incidents
- **Monthly**: Full review of all runbooks
- **Quarterly**: Major version update and stakeholder review

### Contributing

To update a runbook:

1. Create a branch with your changes
2. Test procedures in staging environment
3. Get peer review from another SRE
4. Update version number and date
5. Merge and announce changes in #sre channel

### Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-30 | SRE Team | Initial release |

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
