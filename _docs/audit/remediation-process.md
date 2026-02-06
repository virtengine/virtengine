# VirtEngine Audit Finding Remediation Process

## Purpose
Define a consistent workflow for triaging, remediating, and verifying
third-party security audit findings. This process aligns with the mainnet
launch gate: all Critical and High findings must be remediated and verified
by auditors before launch.

## Roles
- Audit Coordinator: Owns intake, tracking, and communications.
- Security Lead: Final severity confirmation and risk acceptance.
- Engineering Lead: Owns remediation delivery and test coverage.
- Auditor: Verifies fixes and provides retest confirmation.

## Severity Definitions
- CRITICAL: Consensus break, fund loss, key compromise, total bypass.
- HIGH: Severe security impact with realistic exploit path.
- MEDIUM: Moderate impact, reduced exploitability or partial control.
- LOW: Minor issues or best-practice violations.
- INFO: Observations and suggestions.

## SLA Targets
- CRITICAL: Triage within 24 hours, fix within 7 days.
- HIGH: Triage within 48 hours, fix within 14 days.
- MEDIUM: Triage within 5 days, fix within 30 days.
- LOW/INFO: Triage within 10 days, fix as capacity allows.

## Workflow

### 1) Intake
- Auditor submits finding via secure channel or GitHub issue labeled
  udit-finding.
- Audit Coordinator logs the finding in the audit tracker.
- Security Lead confirms severity and impact.

### 2) Triage
- Assign engineering owner and due date.
- Determine affected components and risk class.
- Decide if mitigation is needed immediately (e.g., config change).

### 3) Remediation
- Implement fix and add regression tests.
- Document the root cause and resolution.
- Update audit tracker with PR/commit references.

### 4) Verification
- Auditor re-tests the fix in the approved environment.
- If verified, mark the finding as VERIFIED in tracker.
- If failed, return to remediation with updated notes.

### 5) Closure
- Include final verification evidence.
- Record version/commit where fix shipped.
- Close related issues and update disclosure plan if needed.

## Evidence Requirements
- Reproduction steps (before fix).
- Code changes with PR/commit references.
- Test results or logs demonstrating fix.
- Auditor verification statement.

## Risk Acceptance
- Any decision to mark a finding as WONTFIX must include:
  - documented compensating controls,
  - explicit executive approval,
  - auditor acknowledgment (when applicable).

## Communication Cadence
- Daily status updates during active audit window.
- Weekly summary to stakeholders (severity counts and progress).
- Immediate escalation for Critical findings.

## Tracking
- Primary tracker: tools/audit-tracker output JSON.
- Secondary tracker: GitHub issues labeled udit-finding.

## Retest Window
- Schedule a formal retest window with the auditor.
- Bundle fixes where possible to reduce retest overhead.

## Post-Audit Actions
- Integrate findings into security backlog.
- Update security docs and runbooks.
- Prepare public disclosure and attestation artifacts.
