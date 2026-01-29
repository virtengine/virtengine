## Release PR

**Version:** vX.Y.Z
**Target Network:** Mainnet / Testnet
**Release Issue:** #XXXX

---

## Summary

<!-- Brief description of what this release includes -->

## Changes Included

### Features
- #PR_NUMBER - Description

### Bug Fixes
- #PR_NUMBER - Description

### Breaking Changes
- #PR_NUMBER - Description

---

## Release Checklist

### Pre-Merge
- [ ] Version tag follows semver (vMAJOR.MINOR.PATCH[-PRERELEASE])
- [ ] CHANGELOG.md updated with all changes
- [ ] All included PRs are merged
- [ ] CI passes on all platforms
- [ ] Upgrade tests pass: `cd tests/upgrade && UPGRADE_TO=vX.Y.Z make test`

### Post-Merge (for Release Manager)
- [ ] Create release tag
- [ ] Verify GitHub Actions release workflow completes
- [ ] Verify Docker images published
- [ ] Verify binary artifacts downloadable
- [ ] Update release issue with status

---

## Rollback Information

**Previous stable version:** vX.Y.W
**Rollback procedure:** See [RELEASE.md Rollback Procedures](../RELEASE.md#rollback-procedures)

---

**Reference:** [RELEASE.md](../RELEASE.md) | [Release Issue Template](../.github/ISSUE_TEMPLATE/release.md)
