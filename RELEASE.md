# Release Management Process

This document describes the comprehensive release management process for VirtEngine, including branching strategy, versioning, automation, testing, and rollback procedures.

## Table of Contents

- [Branching Strategy](#branching-strategy)
- [Semantic Versioning](#semantic-versioning)
- [Release Types](#release-types)
- [Release Process](#release-process)
- [Release Candidate Testing](#release-candidate-testing)
- [Changelog Generation](#changelog-generation)
- [Release Automation](#release-automation)
- [Production Release Checklist](#production-release-checklist)
- [Rollback Procedures](#rollback-procedures)
- [Release Calendar and Planning](#release-calendar-and-planning)
- [Emergency Releases](#emergency-releases)

---

## Branching Strategy

VirtEngine uses a dual-branch strategy to maintain both active development and stable production releases.

### Primary Branches

| Branch         | Purpose                    | Version Pattern                       |
| -------------- | -------------------------- | ------------------------------------- |
| `main`         | Active development         | Odd minor versions (v0.9.x, v0.11.x)  |
| `mainnet/main` | Stable production releases | Even minor versions (v0.8.x, v0.10.x) |

### Branch Flow

```
main (development)
  │
  ├─── feature/* branches
  │      └─── PR → main
  │
  ├─── v0.9.0-rc.1 (testnet pre-release)
  ├─── v0.9.0-rc.2
  └─── v0.9.0 (testnet release)

mainnet/main (production)
  │
  ├─── v0.10.0-rc.1 (mainnet pre-release)
  └─── v0.10.0 (mainnet release)
```

### Branch Protection Rules

- **main**: Requires PR review, passing CI, conventional commit format
- **mainnet/main**: Requires 2 PR reviews, passing CI, release manager approval

### Merging to mainnet/main

When preparing a mainnet release from development:

```bash
git checkout main
git merge -s ours mainnet/main
git checkout mainnet/main
git merge main
git push origin mainnet/main
```

This preserves history while resolving conflicts in favor of development changes.

---

## Semantic Versioning

VirtEngine follows [Semantic Versioning 2.0.0](https://semver.org/) with blockchain-specific conventions.

### Version Format

```
vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
```

### Version Components

| Component      | When to Increment                                           |
| -------------- | ----------------------------------------------------------- |
| **MAJOR**      | State machine breaking changes, consensus breaking changes  |
| **MINOR**      | New features, API additions (odd = testnet, even = mainnet) |
| **PATCH**      | Bug fixes, performance improvements                         |
| **PRERELEASE** | Release candidates (rc.1, rc.2), alphas, betas              |
| **BUILD**      | Build metadata (e.g., +network.mainnet)                     |

### Network Determination

The minor version determines the target network:

- **Odd minor** (1, 3, 5, 7, 9...): Testnet releases
- **Even minor** (0, 2, 4, 6, 8...): Mainnet releases

Examples:

- `v0.9.0` → Testnet
- `v0.10.0` → Mainnet
- `v0.9.1-rc.1` → Testnet pre-release
- `v0.10.0-rc.1` → Mainnet pre-release

### Version Validation

Use the semver utility to validate versions:

```bash
# Validate version format
./script/semver.sh validate v0.10.0

# Check if pre-release
./script/is_prerelease.sh v0.10.0-rc.1

# Check if mainnet
./script/mainnet-from-tag.sh v0.10.0
```

### Version Enforcement

1. **Git Tags**: All releases must be tagged with valid semver
2. **CI Validation**: Release workflow validates tag format
3. **go.mod**: Module version reflects release tag
4. **Upgrade Names**: Network upgrade names must match semver tags

---

## Release Types

### Stable Release

Full production release for mainnet or testnet.

**Characteristics:**

- No prerelease suffix
- Fully tested upgrade path
- Complete documentation
- Published Docker images
- Homebrew tap updated

**Example:** `v0.10.0`

### Release Candidate (RC)

Pre-production release for testing.

**Characteristics:**

- `-rc.N` suffix
- Available for community testing
- May have known issues
- Docker images tagged with RC suffix

**Example:** `v0.10.0-rc.1`, `v0.10.0-rc.2`

### Patch Release

Bug fixes for an existing release.

**Characteristics:**

- Patch version increment only
- No new features
- Backwards compatible
- Minimal risk

**Example:** `v0.10.1`

### Emergency/Hotfix Release

Critical security or consensus fixes.

**Characteristics:**

- Expedited process
- Minimal scope
- Immediate deployment needed
- May skip RC phase

**Example:** `v0.10.2` (fixing critical issue in v0.10.1)

---

## Release Process

### Phase 1: Planning (2 weeks before)

1. **Create Release Issue**
   - Document target features
   - List included PRs
   - Identify breaking changes
   - Assign release manager

2. **Feature Freeze**
   - No new features after freeze date
   - Only bug fixes and documentation

3. **Dependency Audit**
   - Review go.mod for updates
   - Check for security advisories
   - Update pinned dependencies

### Phase 2: Release Candidate (1 week before)

1. **Create RC Tag**

   ```bash
   git checkout main  # or mainnet/main for mainnet
   git pull origin main
   git tag -a v0.10.0-rc.1 -m "Release candidate v0.10.0-rc.1"
   git push origin v0.10.0-rc.1
   ```

2. **Automated Release**
   - GitHub Actions triggers on tag push
   - GoReleaser builds binaries
   - Docker images published
   - Changelog generated

3. **Community Testing**
   - Announce RC in Discord/Telegram
   - Request validator testing
   - Collect feedback

### Phase 3: Validation (3-5 days)

1. **Integration Testing**

   ```bash
   cd tests/upgrade
   UPGRADE_TO=v0.10.0 make test
   ```

2. **Network Upgrade Testing**
   - Test upgrade from previous version
   - Verify state migrations
   - Check consensus compatibility

3. **Security Review**
   - Audit critical changes
   - Verify no credential exposure
   - Check for dependency vulnerabilities

### Phase 4: Production Release

1. **Final Approval**
   - Release manager sign-off
   - Core team approval
   - Security team clearance (if applicable)

2. **Create Release Tag**

   ```bash
   git tag -a v0.10.0 -m "Release v0.10.0"
   git push origin v0.10.0
   ```

3. **Verify Release**
   - Check GitHub Release page
   - Verify Docker images
   - Confirm Homebrew tap update
   - Test binary downloads

### Phase 5: Post-Release

1. **Announce Release**
   - Blog post (major releases)
   - Discord/Telegram announcement
   - Twitter/social media

2. **Monitor Upgrade**
   - Watch validator adoption
   - Monitor for issues
   - Track upgrade percentage

3. **Documentation**
   - Update user guides
   - Update API documentation
   - Archive release notes

---

## Release Candidate Testing

### Automated Testing

The release workflow automatically runs upgrade tests:

```yaml
# .github/workflows/release.yaml
test-network-upgrade-on-release:
  runs-on: gh-runner-test
  needs: publish
  steps:
    - name: run test
      env:
        UPGRADE_BINARY_VERSION: ${{ env.RELEASE_TAG }}
      run: |
        cd tests/upgrade
        make test
```

### Manual Testing Checklist

#### Build Verification

- [ ] Binary builds successfully on Linux (amd64, arm64)
- [ ] Binary builds successfully on macOS (universal)
- [ ] Docker image builds and runs
- [ ] Version command shows correct version

#### Functional Testing

- [ ] Node starts from genesis
- [ ] Node syncs from existing chain
- [ ] Transactions execute correctly
- [ ] Query endpoints respond
- [ ] CLI commands work as expected

#### Upgrade Testing

- [ ] Upgrade handler executes without error
- [ ] State migrations complete successfully
- [ ] No consensus failures after upgrade
- [ ] Rollback tested (if applicable)

#### Network Testing

- [ ] P2P connections establish
- [ ] Block production continues
- [ ] Validators sign blocks
- [ ] Governance proposals work

### Test Configuration

Create test configuration for each upgrade:

```json
// tests/upgrade/upgrade-v0.10.0.json
{
  "upgrade_name": "v0.10.0",
  "genesis_version": "v0.9.0",
  "validators": [{ "moniker": "validator0", "voting_power": 1000000 }]
}
```

Run tests:

```bash
cd tests/upgrade
UPGRADE_TO=v0.10.0 CONFIG_FILE=upgrade-v0.10.0.json make test
```

---

## Changelog Generation

### Automatic Generation

Changelogs are automatically generated using [git-chglog](https://github.com/git-chglog/git-chglog):

```bash
make gen-changelog
```

This runs `./script/genchangelog.sh` which:

1. Detects network type from tag (mainnet/testnet)
2. Filters commits by network-specific patterns
3. Generates markdown changelog

### Changelog Configuration

Configuration in `.chglog/config.yaml`:

```yaml
style: github
template: CHANGELOG.tpl.md
options:
  commit_groups:
    title_maps:
      feat: Features
      fix: Bug Fixes
      perf: Performance Improvements
      refactor: Code Refactoring
  header:
    pattern: "^(\\w*)(?:\\(([\\w\\$\\.\\-\\*\\s]*)\\))?\\:\\s(.*)$"
```

### Manual Changelog (CHANGELOG.md)

The root `CHANGELOG.md` uses [Keep a Changelog](https://keepachangelog.com/) format with stanzas:

| Stanza                 | Description                           |
| ---------------------- | ------------------------------------- |
| Features               | New features                          |
| Improvements           | Changes in existing functionality     |
| Deprecated             | Soon-to-be removed features           |
| Bug Fixes              | Any bug fixes                         |
| Client Breaking        | Breaking CLI commands and REST routes |
| API Breaking           | Breaking exported APIs                |
| State Machine Breaking | Changes affecting AppState            |

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer(s)]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Scopes: `veid`, `mfa`, `encryption`, `market`, `escrow`, `roles`, `hpc`, `provider`, `sdk`, `cli`, `app`

---

## Release Automation

### GoReleaser Configuration

The `.goreleaser.yaml` defines:

- Multi-platform builds (linux/darwin, amd64/arm64)
- Universal macOS binaries
- ZIP archives with checksums
- Docker images (multi-arch)
- DEB/RPM packages
- GitHub releases

### Release Workflow

```yaml
# .github/workflows/release.yaml
on:
  workflow_dispatch: # Manual trigger only

jobs:
  publish:
    steps:
      - name: Make and publish
        run: make release
        env:
          GORELEASER_RELEASE: true
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Docker Images

Published to GitHub Container Registry:

```
ghcr.io/virtengine/virtengine:<version>
ghcr.io/virtengine/virtengine:<version>-amd64
ghcr.io/virtengine/virtengine:<version>-arm64
ghcr.io/virtengine/virtengine:latest  # stable releases only
ghcr.io/virtengine/virtengine:stable  # mainnet stable only
```

### Homebrew Tap

Stable mainnet releases trigger Homebrew tap update:

```yaml
notify-homebrew:
  if: ${{ secrets.GORELEASER_ACCESS_TOKEN != '' }}
  steps:
    - name: notify homebrew
      uses: benc-uk/workflow-dispatch@v1
      with:
        repo: virtengine/homebrew-tap
        workflow: virtengine
        inputs: '{"tag": "${{ env.RELEASE_TAG }}"}'
```

---

## Production Release Checklist

Use this checklist for each production release. Copy to release issue.

### Pre-Release (T-7 days)

- [ ] Create release branch/tag plan
- [ ] Identify all PRs for inclusion
- [ ] Review and merge pending PRs
- [ ] Update CHANGELOG.md unreleased section
- [ ] Verify dependency versions
- [ ] Run security audit: `make audit`
- [ ] Run full test suite: `make test`

### Release Candidate (T-5 days)

- [ ] Create RC tag: `git tag -a vX.Y.Z-rc.1 -m "..."`
- [ ] Push tag: `git push origin vX.Y.Z-rc.1`
- [ ] Verify GitHub Actions completes
- [ ] Verify Docker images published
- [ ] Verify binary downloads work
- [ ] Announce RC to community
- [ ] Begin validator testing period

### Validation (T-3 days)

- [ ] Run upgrade tests: `UPGRADE_TO=vX.Y.Z make test`
- [ ] Verify no regression in key functionality
- [ ] Confirm upgrade path works from previous version
- [ ] Review community feedback from RC testing
- [ ] Address critical issues (if any, may need RC2)

### Final Release (T-0)

- [ ] Get release manager approval
- [ ] Get core team sign-off
- [ ] Create final tag: `git tag -a vX.Y.Z -m "..."`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Verify GitHub Release created
- [ ] Verify Docker images with final tag
- [ ] Verify Homebrew tap updated (mainnet only)
- [ ] Test installation on clean system

### Post-Release (T+1)

- [ ] Announce release on Discord/Telegram
- [ ] Post to Twitter/social media
- [ ] Update documentation site
- [ ] Monitor validator upgrade progress
- [ ] Watch for post-release issues
- [ ] Close release issue

---

## Rollback Procedures

### When to Rollback

- Consensus failures on mainnet
- Critical security vulnerability
- Data corruption or state issues
- Widespread validator crashes

### Rollback Decision Tree

```
Issue Detected
    │
    ├── Consensus Failure?
    │   ├── Yes → Immediate Rollback
    │   └── No ↓
    │
    ├── Security Critical?
    │   ├── Yes → Expedited Rollback
    │   └── No ↓
    │
    ├── Affecting >10% validators?
    │   ├── Yes → Consider Rollback
    │   └── No → Patch Forward
    │
    └── Data Corruption?
        ├── Yes → Rollback + State Recovery
        └── No → Investigate & Decide
```

### Rollback Types

#### 1. Binary Rollback (No State Change)

For issues not affecting chain state:

```bash
# Stop the node
systemctl stop virtengined

# Replace binary with previous version
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
wget https://github.com/virtengine/virtengine/releases/download/v0.9.0/virtengine_v0.9.0_linux_${ARCH}.zip
unzip -o virtengine_v0.9.0_linux_${ARCH}.zip -d /usr/local/bin/

# Remove upgrade marker (if upgrade occurred)
rm -f ~/.virtengine/data/upgrade-info.json

# Start with previous version
systemctl start virtengined
```

#### 2. State Rollback (Coordinated)

For consensus-affecting issues requiring chain rollback:

1. **Halt the Network**

   ```bash
   # Coordinate with validators
   systemctl stop virtengined
   ```

2. **Identify Rollback Height**

   ```bash
   # Find last known good block
   virtengine query block --height <TARGET_HEIGHT>
   ```

3. **Export State (if needed)**

   ```bash
   virtengine export --height <TARGET_HEIGHT> > genesis_rollback.json
   ```

4. **Rollback Data**

   ```bash
   # Backup current data
   cp -r ~/.virtengine/data ~/.virtengine/data.backup

   # Reset to target height
   virtengine rollback --hard
   ```

5. **Coordinate Restart**
   - Ensure 2/3+ voting power ready
   - Synchronize restart time
   - Monitor consensus resumption

#### 3. Emergency Fork

For catastrophic issues requiring chain fork:

1. Halt all validators
2. Export state at safe height
3. Modify genesis with fixes
4. Distribute new genesis
5. Coordinated chain restart
6. Announce fork and block height

### Rollback Communication

Template for rollback announcement:

```
🚨 URGENT: Network Rollback Required

Issue: [Brief description]
Affected versions: vX.Y.Z
Rollback to: vX.Y.W (or height H)
Coordinator: [Name/Handle]

Instructions:
1. Stop your node immediately
2. [Specific rollback steps]
3. Wait for coordination signal
4. Restart when announced

ETA: [Time estimate]
Coordination channel: [Discord/Telegram link]
```

### Version Retraction

For released versions with critical issues:

```go
// go.mod
retract (
    v0.10.1 // Contains critical consensus bug
    [v0.10.2, v0.10.3] // Range retraction
)
```

The upgrade script checks retracted versions:

```bash
# script/upgrades.sh automatically handles retracted versions
retracted_versions=$(go mod edit --json | jq -cr .Retract)
```

---

## Release Calendar and Planning

### Release Cadence

| Release Type   | Cadence   | Notice Period |
| -------------- | --------- | ------------- |
| Major (vX.0.0) | As needed | 4 weeks       |
| Minor (vX.Y.0) | Monthly   | 2 weeks       |
| Patch (vX.Y.Z) | As needed | 1 week        |
| Emergency      | Immediate | None          |

### Quarterly Planning

Each quarter, plan releases:

1. **Q Start**: Identify target features
2. **Month 1**: Development sprint
3. **Month 2**: Integration and testing
4. **Month 3**: RC, validation, release

### Release Calendar Template

```markdown
## Q1 2026 Release Calendar

### January

- Jan 6: Feature freeze for v0.10.0
- Jan 8: v0.10.0-rc.1 release
- Jan 13: RC testing complete
- Jan 15: v0.10.0 production release

### February

- Feb 3: Feature freeze for v0.10.1
- Feb 5: v0.10.1-rc.1 release
- Feb 10: v0.10.1 production release

### March

- Mar 3: Feature freeze for v0.11.0 (testnet)
- Mar 5: v0.11.0-rc.1 release
- Mar 12: v0.11.0 production release
```

### Release Roles

| Role                | Responsibilities                                             |
| ------------------- | ------------------------------------------------------------ |
| **Release Manager** | Coordinates release, owns checklist, makes go/no-go decision |
| **QA Lead**         | Oversees testing, validates acceptance criteria              |
| **Security Lead**   | Reviews security implications, clears release                |
| **Communications**  | Drafts announcements, updates documentation                  |
| **On-Call**         | Monitors post-release, handles incidents                     |

### Governance Integration

For network upgrades requiring on-chain governance:

1. **Submit Proposal** (T-14 days before target)

   ```bash
   virtengine tx gov submit-proposal software-upgrade v0.10.0 \
     --title "Upgrade to v0.10.0" \
     --description "..." \
     --upgrade-height <TARGET_HEIGHT> \
     --upgrade-info '{"binaries":{"linux/amd64":"..."}}' \
     --deposit 1000000uve \
     --from validator
   ```

2. **Voting Period** (T-14 to T-7)
   - Community discussion
   - Validator voting
   - Quorum and threshold met

3. **Upgrade Execution** (T-0)
   - Chain halts at upgrade height
   - Validators upgrade binaries
   - Chain resumes with new version

---

## Emergency Releases

### Triggering Conditions

- Active exploitation of vulnerability
- Consensus failure affecting block production
- Critical data integrity issues
- Regulatory compliance requirements

### Emergency Process

1. **Triage (0-2 hours)**
   - Confirm issue severity
   - Assemble response team
   - Establish communication channel

2. **Fix Development (2-8 hours)**
   - Minimal scope fix
   - Focused testing
   - Security review

3. **Expedited Release (8-12 hours)**
   - Skip RC for critical fixes
   - Direct tag and release
   - Immediate notification

4. **Post-Incident (24-48 hours)**
   - Root cause analysis
   - Process improvements
   - Public disclosure (if security)

### Security Disclosure

For security vulnerabilities:

1. **Private Disclosure** to core team
2. **Patch Development** under embargo
3. **Coordinated Release** with validators
4. **Public Disclosure** after upgrade adoption

See [SECURITY.md](./SECURITY.md) for vulnerability reporting.

---

## Related Documentation

- [CONTRIBUTING.md](./CONTRIBUTING.md) - Contribution guidelines and commit format
- [CHANGELOG.md](./CHANGELOG.md) - Version history
- [ADR-001: Network Upgrades](./_docs/adr/adr-001-network-upgrades.md) - Upgrade implementation guide
- [\_docs/version-control.md](./_docs/version-control.md) - Branch merge procedures

---

## Appendix: Quick Reference

### Common Commands

```bash
# Validate version
./script/semver.sh validate v0.10.0

# Check if mainnet version
./script/mainnet-from-tag.sh v0.10.0

# Check if pre-release
./script/is_prerelease.sh v0.10.0-rc.1

# Generate changelog
make gen-changelog

# Build release locally (no publish)
GORELEASER_SKIP=publish make release

# Run upgrade tests
cd tests/upgrade && UPGRADE_TO=v0.10.0 make test

# Create signed tag
git tag -s -a v0.10.0 -m "Release v0.10.0"
```

### Version Examples

| Tag          | Network | Type        | Stable |
| ------------ | ------- | ----------- | ------ |
| v0.9.0       | Testnet | Release     | Yes    |
| v0.9.1-rc.1  | Testnet | Pre-release | No     |
| v0.10.0      | Mainnet | Release     | Yes    |
| v0.10.0-rc.1 | Mainnet | Pre-release | No     |
| v0.10.1      | Mainnet | Patch       | Yes    |
| v1.0.0       | Mainnet | Major       | Yes    |
