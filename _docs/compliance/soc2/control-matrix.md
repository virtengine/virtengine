# Control Matrix

Control IDs map objectives to implementation and evidence. Evidence paths reference repo documents or automated evidence snapshots.

| Control ID | Objective | Control Description | Evidence | Owner | Frequency |
| --- | --- | --- | --- | --- | --- |
| SOC2-CC1.1 | Governance | Security and compliance policies approved and reviewed quarterly. | SECURITY.md, PRIVACY_POLICY.md, _docs/security-guidelines.md | Security | Quarterly |
| SOC2-CC2.1 | Communication | Security responsibilities and incident response documented and communicated. | _docs/operations/incident-response.md, _docs/security-checklist.md | Security | Quarterly |
| SOC2-CC3.1 | Risk Assessment | Risk register maintained with remediation tracking. | _docs/security-review-8d-ml-verification.md, _docs/threat-model.md | Security | Quarterly |
| SOC2-CC4.1 | Monitoring | Continuous monitoring of logs, alerts, and SLOs with documented reviews. | _docs/slos-and-playbooks.md, _docs/operations/monitoring.md | SRE | Monthly |
| SOC2-CC5.1 | Control Activities | Change management via PR review, CI, and approvals. | CONTRIBUTING.md, _docs/version-control.md | Engineering | Per change |
| SOC2-CC6.1 | Logical Access | Access reviews, least privilege, and MFA enforced. | _docs/security-guidelines.md, _docs/onboarding/access-control.md | Security | Quarterly |
| SOC2-CC7.1 | System Operations | Incident response procedures and post-incident reviews. | _docs/runbooks, _docs/operations/incident-response.md | SRE | Per incident |
| SOC2-CC8.1 | Change Management | Releases documented with rollback and approval controls. | RELEASE.md, CHANGELOG.md | Engineering | Per release |
| SOC2-CC9.1 | Vendor Risk | Vendor assessments performed for critical providers. | _docs/security/vendor-risk.md, SUPPLY_CHAIN_SECURITY.md | Security | Annual |
| SOC2-A1.1 | Availability | Capacity planning and availability targets defined with SLOs. | _docs/slos-and-playbooks.md | SRE | Quarterly |
| SOC2-A1.2 | Availability | Backup and disaster recovery procedures documented and tested. | _docs/disaster-recovery.md, _docs/business-continuity.md | SRE | Annual |
| SOC2-C1.1 | Confidentiality | Data classification and encryption requirements documented. | _docs/data-classification.md, _docs/key-management.md | Security | Annual |
| SOC2-C1.2 | Confidentiality | Data retention and archival controls defined. | _docs/data-retention-policy.md, _docs/data-archival-guide.md | Compliance | Annual |
| SOC2-EVIDENCE | Evidence | Automated evidence collection executed and stored. | scripts/compliance/collect-soc2-evidence.sh output | Compliance | Monthly |

Notes:
- Evidence references may include repository documentation and generated evidence bundles.
- Evidence bundles are stored outside git in _build/compliance/soc2/.
- Control testing evidence is stored in the compliance evidence system for the audit window.
