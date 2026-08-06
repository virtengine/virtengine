# Verification Posture

This document describes the public verification posture of the VirtEngine repository. It is not a claim that every feature or target environment is production-approved; it is a summary of the evidence paths that must be satisfied before that claim can be made.

## Current Status

As of `2026-04-11`:

- the repository contains extensive CI, security, release, smoke, and DR automation;
- launch-readiness evidence for mainnet is checked in under `_docs/operations/` and `config/mainnet/`;
- the checked-in mainnet decision record is `GO` for the scheduled 2026-04-18 UTC launch window, with the final canonical allocation record and genesis publication bundle archived in-repo.

Accordingly, the current verification posture is:

- development and release-candidate validation: available;
- production-grade automation evidence: present in many tracks and workflows;
- approved public mainnet launch verification: complete for the scheduled launch window, while public network availability remains future-dated until the window begins.

## Primary Verification Gates

### Quality Gate

[`.github/workflows/quality-gate.yaml`](.github/workflows/quality-gate.yaml) is the fast baseline gate for lint, vet, build, unit-test, economics-sim, and AGENTS documentation validation.

### Compatibility and Upgrade Validation

Compatibility and upgrade behavior are validated through [`.github/workflows/compatibility.yaml`](.github/workflows/compatibility.yaml) and the upgrade test path used by the release workflow. Release candidates should not be described as supported until those checks pass for the target tag.

### Security Verification

[`.github/workflows/security.yaml`](.github/workflows/security.yaml) is the repository's primary security scan workflow. It includes CodeQL, `govulncheck`, dependency auditing, container scanning, SBOM generation, and secret scanning verification.

[`.github/workflows/supply-chain.yaml`](.github/workflows/supply-chain.yaml) provides the companion supply-chain verification path for provenance, dependency, and artifact-integrity checks.

### Release Verification

[`.github/workflows/release.yaml`](.github/workflows/release.yaml) is the authoritative release automation. It is manually triggered and includes:

- artifact publishing through `make release`;
- upgrade test execution when required;
- conditional Homebrew notification for qualifying non-prerelease mainnet-classified tags.

### Smoke, Staging, and DR Validation

The repository also contains [`.github/workflows/smoke-test.yaml`](.github/workflows/smoke-test.yaml), [`.github/workflows/staging-e2e.yaml`](.github/workflows/staging-e2e.yaml), [`.github/workflows/infrastructure.yaml`](.github/workflows/infrastructure.yaml), and [`.github/workflows/dr-failover-test.yaml`](.github/workflows/dr-failover-test.yaml) to validate deployability, rollout behavior, and failover procedures. Those workflows are part of the evidence chain for any environment being described as production-ready.

## Mainnet Verification Posture

Mainnet verification is intentionally fail-closed.

The checked-in evidence currently shows:

- a quorum-backed launch review record and named approver set archived in repository evidence;
- approved launch and freeze windows checked in;
- completed VEID, provider, finance, backup/restore, and communications evidence;
- an approved canonical allocation control record;
- a final checked-in genesis publication bundle with current hashes;
- a final launch state of `GO` for the scheduled 2026-04-18 UTC window.

As a result, the repository should currently be described as:

- launch-approved for the scheduled mainnet window;
- backed by checked-in final genesis publication evidence;
- not yet making a claim that the network is already live before the approved launch window executes.

## How To Evaluate A Target Release

Before calling a tag supported for a specific environment, confirm all of the following:

1. the tag exists and the release workflow completed successfully;
2. compatibility and upgrade checks passed for the relevant path;
3. security verification completed with acceptable results;
4. smoke, staging, deployment, or DR evidence exists for the environment being targeted;
5. if the target is mainnet, the launch packet and go/no-go record explicitly show approval.

## What This Document Does Not Claim

This document does not claim:

- that all repository features are production-supported;
- that a published binary automatically implies mainnet approval;
- that support windows or compatibility guarantees exist beyond what release notes and evidence packages explicitly record.

## Related Documentation

- [README.md](README.md)
- [RELEASE.md](RELEASE.md)
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md)
- [_docs/operations/mainnet-go-no-go-decision.md](_docs/operations/mainnet-go-no-go-decision.md)
