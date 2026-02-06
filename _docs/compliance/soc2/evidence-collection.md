# Evidence Collection

Evidence collection supports SOC 2 Type II testing and ongoing control monitoring. Evidence is collected monthly and on significant operational events.

## Automated Collection

Automation is handled by scripts/compliance/collect-soc2-evidence.sh using the manifest scripts/compliance/soc2-evidence-manifest.yaml. The script gathers non sensitive configuration, process metadata, and documentation references.

Outputs are stored in _build/compliance/soc2/ with a timestamped folder that includes:

- control-matrix.json (generated reference for audit sampling)
- git metadata snapshots (log and diff summaries)
- system tooling versions (go, node, pnpm when present)
- documentation checksums for key policies

## Manual Evidence

Some evidence remains manual and must be added to the monthly evidence packet:

- Access review approvals
- Vendor assessment approvals
- Incident postmortems and corrective action tracking
- Business continuity and disaster recovery test results

## Evidence Retention

- Retain evidence for at least the audit period plus one year.
- Store evidence in the compliance evidence system and not in git.
