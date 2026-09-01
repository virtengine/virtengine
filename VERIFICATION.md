# Verification Posture

This document describes the public verification posture of the VirtEngine repository. It is not a claim that every feature or target environment is production-approved; it is a summary of the evidence paths that must be satisfied before that claim can be made.

## Current Status

As of `2026-09-01`:

- the repository contains extensive CI, security, release, smoke, and DR automation;
- **TestNet is planned for the January 2027 launch window** to produce public, multi-operator pre-production evidence;
- **MainNet is planned for the March 2027 launch window**, contingent on TestNet exit criteria and a fresh production go/no-go approval;
- April 2026 MainNet readiness material remains archived under `_docs/operations/` and `config/mainnet/`, but the windows did not proceed and those records are not current launch authority.

Accordingly, the current verification posture is:

- development and release-candidate validation: available;
- production-grade automation evidence: present in many tracks and workflows;
- TestNet launch verification: to be refreshed and completed for the January 2027 window;
- MainNet launch verification: not complete until TestNet evidence is accepted, launch-blocking findings are closed and re-tested, final artifacts are verified, and a fresh `GO` decision is recorded for March 2027.

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

## Network Launch Verification Posture

Launch verification is intentionally fail-closed and environment-specific.

For **TestNet**, the January 2027 launch package must validate public
multi-operator consensus, validator and provider onboarding, VEID, marketplace
and settlement flows, upgrades, rollback, backup/restore, monitoring, and
incident response. TestNet may be reset, its tokens have no production value,
and its state is not guaranteed to migrate to MainNet.

For **MainNet**, the March 2027 launch package must additionally show that
TestNet exit criteria passed, every launch-blocking finding was closed and
re-tested, canonical genesis and production artifacts were finalized and
verified, and release management recorded a fresh `GO` decision. TestNet
success does not automatically authorize MainNet.

The separation reserves February for observation, remediation, repeated
validation, security and operational review, release freeze, and final
validator coordination. See
[docs/NETWORK_LAUNCH_SCHEDULE.md](docs/NETWORK_LAUNCH_SCHEDULE.md).

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
- [docs/NETWORK_LAUNCH_SCHEDULE.md](docs/NETWORK_LAUNCH_SCHEDULE.md)
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md)
- [_docs/operations/network-launch-schedule.md](_docs/operations/network-launch-schedule.md)
