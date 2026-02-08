# SOC 2 Type II Compliance Program

This directory defines VirtEngine SOC 2 Type II compliance program for the Security, Availability, and Confidentiality trust services criteria. It documents control objectives, control mappings, evidence collection, and audit readiness activities.

## Scope

- System: VirtEngine core chain, provider services, portal, and supporting infrastructure
- Criteria: Security, Availability, Confidentiality
- Period: Rolling 12-month audit window aligned to the compliance calendar

## Status

- Report status: Obtained
- Audit period: 2025-02-01 to 2026-01-31
- Report date: 2026-02-06
- Auditor: Approved third party auditor
- Evidence system: Compliance evidence system (restricted access)

## Program Artifacts

- Control objectives: _docs/compliance/soc2/control-objectives.md
- Control matrix: _docs/compliance/soc2/control-matrix.md
- Evidence collection: _docs/compliance/soc2/evidence-collection.md
- Continuous monitoring: _docs/compliance/soc2/continuous-monitoring.md
- Audit readiness: _docs/compliance/soc2/audit-readiness.md
- Gap analysis: _docs/compliance/soc2/gap-analysis.md
- Auditor engagement: _docs/compliance/soc2/auditor-engagement.md
- Continuous compliance program: _docs/compliance/soc2/continuous-compliance-program.md
- Type II report register: _docs/compliance/soc2/type-ii-report.md

## Control Implementation and Testing

- Controls are implemented per the control matrix and validated during audit readiness.
- Control testing evidence for the audit window is stored in the compliance evidence system.
- Continuous monitoring evidence is collected monthly and attached to evidence packets.

## Evidence Automation

Automation is handled by the script below. It collects evidence snapshots without secrets and writes to _build/compliance/soc2/.

- Script: scripts/compliance/collect-soc2-evidence.sh
- Manifest: scripts/compliance/soc2-evidence-manifest.yaml

## Owner and Cadence

- Compliance owner: Security and Risk
- Evidence cadence: Monthly plus ad hoc for major releases or incidents
- Review cadence: Quarterly control effectiveness review
