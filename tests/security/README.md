# VirtEngine Security Evidence Suite

**Scope:** audited crypto, identity, attestation, consensus, and operator-reproducible security checks
**Last Updated:** 2026-04-11

## Purpose

`tests/security/` is the repository-level security evidence suite. It is not a generic penetration-test wish list. The files in this directory either:

- exercise real production code paths for audited controls, or
- provide deterministic regression coverage for shared security boundaries, or
- reproduce the local security bar documented in [SECURITY.md](/C:/Users/jON/Documents/source/repos/virtengine-gh/virtengine/SECURITY.md).

## Evidence Map

| Area | Primary Coverage | What It Proves |
| --- | --- | --- |
| Crypto envelopes and salt binding | `audit_crypto_contract_test.go` | malformed envelopes are rejected, salt binding requires valid signatures and fresh timestamps, attestation parsers reject malformed artifacts |
| Identity + MFA | `identity_integration_test.go`, `mfa_enforcement_test.go` | wallet rebinding requires linked signatures, serialized MFA proofs gate sensitive key-rotation flows |
| Consensus verification | `blockchain/consensus_test.go` | score/model/input-hash divergence is rejected, result hashing is deterministic, unhealthy or mismatched verifier state fails closed |
| Attestation lifecycle | `attestation_e2e_test.go` | heartbeat replay is rejected, stale rotation keys are rejected after overlap closes |
| API and generic state guards | `api/*.go`, `blockchain/state_machine_test.go` | authentication, rate-limit, authority, and authz regression expectations stay deterministic in the repo suite |

## Commands

```bash
# Unit / contract coverage
go test -tags=security ./tests/security/...

# Integration coverage
go test -tags='security,integration' ./tests/security/...

# E2E coverage
go test -tags='security,e2e.integration' ./tests/security/...

# Full local reproduction path
bash ./tests/security/scripts/reproduce_security_checks.sh full
```

## Fail-Closed Scripts

`tests/security/scripts/` contains the shell entry points used by the docs:

- `static_analysis.sh` runs `gosec` and `go vet` against `./tests/security/...`
- `scan_dependencies.sh` runs `govulncheck` against `./tests/security/...` and then the repository dependency risk assessor
- `secret_scan.sh` runs a pinned `gitleaks dir` current-tree scan with the repository config
- `reproduce_security_checks.sh` runs the test matrix plus the scripts above

All four scripts are intended to fail closed.

The Go-based scripts pin `GOTOOLCHAIN=go1.25.9+auto` unless the caller overrides it, so Git Bash and PowerShell reproduce the same toolchain behavior.

## Audit Follow-Up References

The current audit/runbook mapping lives in:

- `SECURITY.md`
- `_docs/audits/security-audit-report-2026-02-06.md`
- `_docs/training/security/security-incident-response.md`
