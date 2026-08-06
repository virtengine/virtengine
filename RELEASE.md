# Release Management Process

This document describes the current release process for VirtEngine as it exists in this repository today. It intentionally distinguishes between release-capable automation and an approved public mainnet launch.

## Current Public Release State

As of `2026-04-11`:

- the repository contains release automation, upgrade tests, and network-tag helper scripts;
- mainnet launch preparation artifacts exist under `config/mainnet/` and `_docs/operations/`;
- the checked-in mainnet decision record is `HOLD`: execution evidence and named approvals are checked in, but no public document should describe a mainnet general-availability line as approved or live until signed allocation addresses are inserted and the final genesis bundle is published.

Operationally, that means:

- a tag can still be built and published through the release workflow;
- operators must not infer that every published tag is approved for mainnet production use;
- production approval requires the launch packet, readiness checklist, and go/no-go package to move to an explicit approved state.

## Branching and Source of Truth

The repository currently develops from `main`.

- `main` is the active integration branch for code and documentation.
- Legacy tooling and workflows may still reference `mainnet/main` for tag classification or compatibility checks.
- This checkout does not use `mainnet/main` as the authoritative public release branch, so contributors should not document it as the current stable line.

If the project later restores a dedicated stable branch, this document and [_docs/version-control.md](_docs/version-control.md) should be updated together.

## Semantic Versioning

VirtEngine uses semantic versioning:

```text
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

The repository still includes helper scripts that classify tags by minor version:

```bash
./script/semver.sh validate v0.10.0
./script/is_prerelease.sh v0.10.0-rc.1
./script/mainnet-from-tag.sh v0.10.0
```

Those helpers are implementation details for the release pipeline. They do not, by themselves, establish that a tag is approved for public mainnet rollout.

## Release Types

### Development Tag

A semver tag that packages the current codebase for testing, validation, or staged rollout.

### Release Candidate

A pre-release tag with an `-rc.N` suffix used for upgrade testing and staged validation.

### Production-Approved Release

A published tag that is also backed by:

- completed launch or upgrade evidence,
- the required security and operations approvals,
- an explicit go decision for the target environment.

Until the go/no-go package records that approval, public docs should treat release tags as artifacts to validate, not proof of a launched mainnet.

## Release Workflow

The authoritative release automation is [`.github/workflows/release.yaml`](.github/workflows/release.yaml).

Current behavior:

- trigger: `workflow_dispatch`
- publish step: `make release`
- post-publish upgrade test: `tests/upgrade`
- Homebrew notification: only when the tag is classified as mainnet and not a prerelease

This workflow is manual. It is not triggered automatically by every tag push.

## Release Procedure

### 1. Prepare the candidate

1. Confirm the target commit on `main`.
2. Validate the intended tag with `./script/semver.sh`.
3. Update public docs if the release changes compatibility, verification posture, or support commitments.

### 2. Create the tag

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

Use `-rc.N` for release candidates.

### 3. Run the release workflow

Run the `release` workflow manually and confirm:

- build and publish succeeded;
- upgrade validation ran when required;
- release artifacts were generated;
- Homebrew notification only fired when the tag met the workflow conditions.

### 4. Validate release evidence

Before describing a tag as production-ready, confirm:

- [VERIFICATION.md](VERIFICATION.md) still matches the actual verification posture;
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md) still matches supported compatibility windows;
- the launch or upgrade evidence package contains named approvals and execution artifacts for the target environment.

### 5. Communicate accurately

Public announcements and docs must distinguish among:

- a tagged build artifact,
- a validated staging or rehearsal candidate,
- an approved production rollout.

Do not collapse those states into one label.

## Upgrade and Compatibility Testing

The release workflow runs upgrade testing from `tests/upgrade` when `script/upgrades.sh test-required` indicates that an upgrade path must be validated.

Representative local commands:

```bash
cd tests/upgrade
UPGRADE_BINARY_VERSION=vX.Y.Z make test
```

Release managers should also review:

- compatibility workflow results,
- security workflow results,
- any smoke or environment-specific rollout checks relevant to the target deployment.

## Rollback and Retraction

Rollback and DR execution guidance lives in:

- [docs/operations/DEPLOYMENT_GUIDE.md](docs/operations/DEPLOYMENT_GUIDE.md)
- [scripts/rollback](scripts/rollback)
- [scripts/dr](scripts/dr)

If a published tag must be withdrawn:

1. stop describing it as a supported production target;
2. update support and verification docs;
3. use `go mod edit -retract` when the module version itself must be retracted;
4. publish corrected operational guidance for operators already testing or deploying the tag.

## What This Document Does Not Claim

This document does not claim:

- that there is an approved public mainnet launch today;
- that the repository currently operates a dedicated stable `mainnet/main` release branch;
- that a monthly or quarterly release calendar is committed and active;
- that every semver tag automatically carries production support.

Those claims must be backed by checked-in evidence, not inherited from older process text.

## Related Documentation

- [README.md](README.md)
- [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md)
- [VERIFICATION.md](VERIFICATION.md)
- [_docs/version-control.md](_docs/version-control.md)
- [_docs/operations/mainnet-go-no-go-decision.md](_docs/operations/mainnet-go-no-go-decision.md)
