# VirtEngine Module Security Audit Scope

## Overview
This document defines the scope for a third-party security audit of VirtEngine's
Cosmos SDK modules. The audit focuses on consensus-critical code paths, state
transition logic, economic invariants, and authorization/validation boundaries.

## Objectives
- Identify vulnerabilities in module state machines and message handling.
- Validate authorization, signature checks, and governance authority usage.
- Confirm correctness of economic calculations and invariants.
- Review on-chain audit trails and event emission for forensics.
- Provide remediation recommendations and re-test verification.

## In-Scope Modules

### x/veid (Identity Verification)
- Scope registration and encryption envelope handling.
- Three-signature scheme validation (user + client + salt binding).
- ML score aggregation and deterministic validation.
- Key rotation workflows and re-verification logic.
- Privacy-preserving identity storage and redaction.

Critical areas:
- Envelope parsing and replay protection.
- Signature verification bypass or downgrade.
- Score manipulation via malformed inputs.
- Cross-scope data leakage or state confusion.

### x/mfa (Multi-Factor Authentication)
- Challenge generation and validation flows.
- Time-based verification and expiration.
- Recovery and fallback mechanisms.
- Rate limiting and abuse controls.

Critical areas:
- Challenge prediction or reuse.
- Time skew and clock manipulation.
- Recovery bypass and stale challenge acceptance.

### x/encryption (Cryptographic Operations)
- X25519-XSalsa20-Poly1305 usage.
- Nonce handling and envelope structure.
- Key derivation flows and metadata validation.

Critical areas:
- Key/nonce reuse vulnerabilities.
- Ciphertext malleability and length validation.
- Cross-module key mixing or downgrade paths.

### x/market (Marketplace)
- Order matching and validation.
- Bid/ask handling and price calculations.
- Lease state transitions and settlement paths.

Critical areas:
- Front-running vectors and price manipulation.
- State transition gaps allowing double allocation.
- Precision and overflow/underflow conditions.

### x/escrow (Financial)
- Escrow creation/funding lifecycle.
- Settlement calculations and fee handling.
- Dispute resolution and slashing logic.

Critical areas:
- Fund extraction or locked funds.
- Double-spend or premature release scenarios.
- Rounding errors affecting payouts.

### x/roles (Access Control)
- Role assignment and revocation.
- Governance authority checks.
- Role-based gating across modules.

Critical areas:
- Privilege escalation via role misassignment.
- Authority mismatch or replay of role updates.

### x/provider (Provider Registry)
- Provider registration and updates.
- Offering lifecycle and verification gating.
- Provider-level audit trails.

Critical areas:
- Unauthorized provider state changes.
- Fake offerings or invalid capacity claims.

### x/audit (Provider Attribute Auditing)
- Audit record creation and integrity.
- Attribute signing and validation.
- Governance approval flows for auditors.

Critical areas:
- Tampering with audit records or proof chains.
- Unauthorized audit submissions.

## Cross-Module Review Areas
- Message validation consistency and shared interfaces.
- Event emission for state transitions and audit trail completeness.
- Authority checks for governance-only actions.
- Determinism and consensus safety for any ML or cryptographic operations.

## Out of Scope
- Frontend applications (portal/admin UI).
- Off-chain provider daemon services (unless explicitly requested).
- ML training pipelines and dataset curation.
- Infrastructure and cloud deployment security.
- Third-party dependency security (covered by supply chain audits).

## Required Artifacts Provided to Auditor
- SECURITY_SCOPE.md (overall system scope).
- Module READMEs and architecture diagrams.
- Governance authority and parameter documentation.
- Test coverage report and known issues list.

## Deliverables
- Detailed finding report with severity classifications.
- Reproduction steps and impacted code paths.
- Remediation recommendations for each finding.
- Retest verification confirming fixes.
- Executive summary for leadership and stakeholders.

## Timeline
- Estimated duration: 4-6 weeks.
- Draft report expected by week 4.
- Final report and retest verification by week 6.

## Points of Contact
- Security Lead: security@virtengine.io
- Engineering Lead: platform@virtengine.io
- Audit Coordinator: audit@virtengine.io
