# Version Control and Branching

This document covers version control practices specific to VirtEngine development.

For comprehensive release management, see [RELEASE.md](../RELEASE.md).

## Branch Strategy

| Branch | Purpose | Version Pattern |
|--------|---------|-----------------|
| `main` | Active development | Odd minor versions (v0.9.x, v0.11.x) |
| `mainnet/main` | Stable production releases | Even minor versions (v0.8.x, v0.10.x) |

## Version Numbering

VirtEngine uses semantic versioning with network-specific conventions:

- **Odd minor versions** (1, 3, 5, 7, 9): Testnet releases
- **Even minor versions** (0, 2, 4, 6, 8): Mainnet releases

Examples:
- `v0.9.0` → Testnet
- `v0.10.0` → Mainnet

## Merging into `mainnet/main`

When a new mainnet release is needed, `mainnet/main` is often far behind `main`.
This procedure performs the merge without conflicts while preserving history:

```shell
git checkout main
git merge -s ours mainnet/main
git checkout mainnet/main
git merge main
git push origin mainnet/main
```

## Semver Utilities

The repository includes utilities for version handling:

```bash
# Validate version format
./script/semver.sh validate v0.10.0

# Check if mainnet version (even minor)
./script/mainnet-from-tag.sh v0.10.0

# Check if pre-release
./script/is_prerelease.sh v0.10.0-rc.1

# Bump version components
./script/semver.sh bump patch v0.10.0  # → v0.10.1
./script/semver.sh bump minor v0.10.0  # → v0.11.0
./script/semver.sh bump major v0.10.0  # → v1.0.0
```

## Related Documentation

- [RELEASE.md](../RELEASE.md) - Complete release management process
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [ADR-001: Network Upgrades](./adr/adr-001-network-upgrades.md) - Network upgrade implementation
