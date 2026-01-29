---
name: Release
about: Track a new release through the release process
title: 'Release vX.Y.Z'
labels: 'release'
assignees: ''
---

## Release Information

| Field | Value |
|-------|-------|
| **Version** | vX.Y.Z |
| **Target Network** | Mainnet / Testnet |
| **Release Type** | Major / Minor / Patch / Emergency |
| **Release Manager** | @username |
| **Target Date** | YYYY-MM-DD |

## Included Changes

List major changes and PRs included in this release:

### Features
- [ ] #PR_NUMBER - Feature description

### Bug Fixes
- [ ] #PR_NUMBER - Bug fix description

### Breaking Changes
- [ ] #PR_NUMBER - Breaking change description

## Pre-Release Checklist (T-7 days)

- [ ] Create release branch/tag plan
- [ ] Identify all PRs for inclusion
- [ ] Review and merge pending PRs
- [ ] Update CHANGELOG.md unreleased section
- [ ] Verify dependency versions in go.mod
- [ ] Run security audit
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint-go`

## Release Candidate Checklist (T-5 days)

- [ ] Create RC tag: `git tag -a vX.Y.Z-rc.1 -m "Release candidate vX.Y.Z-rc.1"`
- [ ] Push tag: `git push origin vX.Y.Z-rc.1`
- [ ] Verify GitHub Actions workflow completes successfully
- [ ] Verify Docker images published to ghcr.io
- [ ] Verify binary downloads work from GitHub Releases
- [ ] Announce RC to community (Discord/Telegram)
- [ ] Request validator testing

## Validation Checklist (T-3 days)

- [ ] Run upgrade tests: `cd tests/upgrade && UPGRADE_TO=vX.Y.Z make test`
- [ ] Verify upgrade path from previous stable version
- [ ] Verify no regression in key functionality
- [ ] Test on localnet: `./scripts/localnet.sh start`
- [ ] Review community feedback from RC testing
- [ ] Address critical issues (if any, create RC2)

## Final Release Checklist (T-0)

- [ ] Get release manager approval
- [ ] Get core team sign-off (minimum 2 approvals)
- [ ] Create final tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Verify GitHub Release page created with correct assets
- [ ] Verify Docker images tagged with final version
- [ ] Verify Homebrew tap updated (mainnet releases only)
- [ ] Test installation on clean system

## Post-Release Checklist (T+1)

- [ ] Announce release on Discord
- [ ] Announce release on Telegram
- [ ] Post to Twitter/social media
- [ ] Update documentation site
- [ ] Monitor validator upgrade progress
- [ ] Watch for post-release issues (check Discord, GitHub Issues)
- [ ] Close this release issue

## Rollback Plan

In case of critical issues:

**Rollback Version:** vX.Y.W (previous stable)
**Rollback Height:** N/A (or specific height if needed)

Rollback procedure:
1. Announce halt on Discord/Telegram
2. Validators stop nodes
3. Replace binary with previous version
4. Remove `~/.virtengine/data/upgrade-info.json` if upgrade occurred
5. Coordinate restart

## Notes

<!-- Add any additional notes, risks, or considerations for this release -->

---

**Reference Documentation:**
- [RELEASE.md](../RELEASE.md) - Complete release process
- [ADR-001: Network Upgrades](../_docs/adr/adr-001-network-upgrades.md) - Upgrade implementation
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Commit message format
