# Security Policy

VirtEngine treats security claims as evidence-backed. The current production bar is whatever can be reproduced by the repository test suites, fail-closed scripts, and incident/runbook material referenced below.

## Supported Versions

| Version | Supported | Notes |
| --- | --- | --- |
| `0.9.x` | Yes | Active development on `main` |
| `0.8.x` | Yes | Stable release line on `mainnet/main` |
| `0.7.x` | Critical fixes only | No new feature work |
| `< 0.7.x` | No | Unsupported |

## Reporting a Vulnerability

Send reports to [security@virtengine.com](mailto:security@virtengine.com).

Include:

- affected component, module, or workflow
- precise reproduction steps
- expected impact and severity estimate
- logs, hashes, screenshots, or proof-of-concept material
- any mitigation or patch suggestions you already validated

Initial acknowledgement target: **48 hours**.

## Severity Targets

| Severity | Examples | Initial Response |
| --- | --- | --- |
| Critical | consensus manipulation, unauthorized fund movement, signing-key compromise, complete MFA/VEID bypass | 24 hours |
| High | validator/provider privilege escalation, replayable attestation or identity flows, sensitive data disclosure | 7 days |
| Medium | bounded logic bugs, non-critical authz drift, incomplete audit logging | 30 days |
| Low | defense-in-depth gaps, documentation/runbook issues, low-impact hygiene issues | 90 days |

## Evidence and Reproduction

### Security Test Matrix

```bash
# Contract / unit coverage
go test -tags=security ./tests/security/...

# Integration coverage
go test -tags='security,integration' ./tests/security/...

# E2E coverage
go test -tags='security,e2e.integration' ./tests/security/...
```

### Local Security Checks

```bash
# Full local reproduction path for the security suite
bash ./tests/security/scripts/reproduce_security_checks.sh full

# Static analysis only (security suite)
bash ./tests/security/scripts/static_analysis.sh

# Dependency and vulnerability checks (security suite plus repository dependency policy)
bash ./tests/security/scripts/scan_dependencies.sh

# Secret detection
bash ./tests/security/scripts/secret_scan.sh

# Workflow/security policy validation
python .github/scripts/validate_security_policies.py
```

### Audit-Backed Control Areas

| Control Area | Primary Evidence |
| --- | --- |
| Crypto envelope validation and salt binding | `tests/security/audit_crypto_contract_test.go` |
| MFA-gated identity recovery and wallet rebinding | `tests/security/identity_integration_test.go`, `tests/security/mfa_enforcement_test.go` |
| Consensus result matching and deterministic hashing | `tests/security/blockchain/consensus_test.go` |
| Enclave heartbeat replay and stale-key rejection | `tests/security/attestation_e2e_test.go` |
| Supply-chain signing, SBOM, provenance, and verification | `SUPPLY_CHAIN_SECURITY.md`, `.github/workflows/security.yaml`, `.github/workflows/supply-chain.yaml` |

## External Audit Status

The current external cryptography and identity audit evidence is:

- public summary: `_docs/audits/security-audit-report-2026-02-06.md`
- engagement record: `_docs/audits/external-security-audit-engagement.md`
- incident procedures and follow-up drills: `_docs/training/security/security-incident-response.md`

The repository does **not** claim a standing or blanket audit for every subsystem. Infrastructure and frontend security workstreams are tracked separately.

## Supply Chain Security

Release-critical supply-chain requirements are documented in `SUPPLY_CHAIN_SECURITY.md`. The required bar includes:

- pinned workflow/tool versions
- fail-closed vulnerability and secret scanning
- SBOM generation
- cosign signing and verification
- provenance generation and verification
- dependency risk assessment with no silent allow-through

## Incident Procedures

Use the internal runbooks for operational response:

- `_docs/training/security/security-incident-response.md`
- `_docs/training/security/security-best-practices.md`
- `_docs/training/security/threat-modeling.md`

These docs now point back to concrete tests or commands for the audited crypto, identity, consensus, and attestation surfaces.

## Scope

### In Scope

- `x/veid`, `x/mfa`, `x/enclave`, and consensus-adjacent app decorators
- provider and enclave signing/attestation paths
- release and supply-chain verification workflows
- audited crypto, key rotation, replay protection, and security logging surfaces

### Out of Scope

- unsupported versions
- third-party services that are not operated by VirtEngine
- purely theoretical issues without a reproducible path
- social engineering or physical-access issues not tied to VirtEngine-operated systems

---

_Last updated: 2026-04-11_
