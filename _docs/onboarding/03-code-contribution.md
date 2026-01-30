# Code Contribution Guidelines

This guide explains how to contribute code to VirtEngine effectively.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Branch Strategy](#branch-strategy)
3. [Commit Conventions](#commit-conventions)
4. [Pull Request Process](#pull-request-process)
5. [Code Standards](#code-standards)
6. [Documentation Requirements](#documentation-requirements)
7. [License and Copyright](#license-and-copyright)

---

## Getting Started

### Before Contributing

1. **Set up your environment** - Follow the [Environment Setup Guide](./01-environment-setup.md)
2. **Understand the architecture** - Read the [Architecture Overview](./02-architecture-overview.md)
3. **Check existing issues** - Look for open issues or discussions
4. **Propose significant changes** - Large features require a proposal first

### Types of Contributions

| Type | Process |
|------|---------|
| Typo/Doc fix | Direct PR (no issue needed) |
| Bug fix | Link to issue, direct PR |
| Small feature | Open issue first, then PR |
| Large feature | Formal proposal required |
| Breaking change | RFC + governance approval |

---

## Branch Strategy

### Branch Structure

```
main            ← Active development (odd versions: v0.9.x, v0.11.x)
  └── feature/*  ← Feature branches
  └── fix/*      ← Bug fix branches
  └── docs/*     ← Documentation branches

mainnet/main    ← Stable releases (even versions: v0.10.x, v0.12.x)
  └── hotfix/*   ← Critical fixes only
```

### Branch Naming

```bash
# Features
feature/veid-enhanced-scoring
feature/market-batch-orders

# Bug fixes
fix/escrow-timeout-handling
fix/mfa-rate-limit

# Documentation
docs/onboarding-guide
docs/api-reference-update

# Refactoring
refactor/keeper-interface-cleanup
```

### Targeting Branches

| Change Type | Target Branch |
|-------------|---------------|
| New features | `main` |
| Bug fixes | `main` (backported if needed) |
| Release-specific fixes | `mainnet/main` |
| Hotfixes | Release branch |

---

## Commit Conventions

VirtEngine uses [Conventional Commits](https://www.conventionalcommits.org/).

### Format

```
type(scope): description

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Code formatting (no logic change) |
| `refactor` | Code restructuring (no behavior change) |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `build` | Build system or dependencies |
| `ci` | CI/CD configuration |
| `chore` | Maintenance tasks |
| `revert` | Reverting a previous commit |

### Scopes

| Scope | Description |
|-------|-------------|
| `veid` | Identity verification module |
| `mfa` | Multi-factor authentication module |
| `encryption` | Encryption module |
| `market` | Marketplace module |
| `escrow` | Escrow module |
| `roles` | Roles module |
| `hpc` | HPC module |
| `provider` | Provider daemon |
| `sdk` | SDK packages |
| `cli` | CLI commands |
| `app` | Application wiring |
| `deps` | Dependencies |
| `ci` | CI/CD |
| `api` | API changes |

### Examples

```bash
# Feature with scope
feat(veid): add identity verification flow

# Bug fix with scope
fix(market): resolve bid race condition

# Documentation (no scope needed)
docs: update contributing guidelines

# Breaking change (note the !)
feat(api)!: change response format for provider endpoints

# With body explaining context
fix(escrow): handle timeout edge case

The escrow module was not properly handling timeouts when
the provider went offline during lease finalization.

Closes #123

# Dependency update
chore(deps): bump cosmos-sdk to v0.53.1
```

### Sign-Off Requirement

Every commit must be signed off:

```bash
git commit -s -m "feat(veid): add identity verification flow"
```

This adds:
```
Signed-off-by: Your Name <your.email@example.com>
```

Configure Git to sign automatically:

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### Local Validation

Install commitlint for local validation:

```bash
npm install --save-dev @commitlint/cli @commitlint/config-conventional
```

The repository includes `commitlint.config.js` with project-specific rules.

---

## Pull Request Process

### Before Opening a PR

1. **Ensure tests pass**:
   ```bash
   go test ./x/... ./pkg/...
   make lint-go
   ```

2. **Rebase on latest main**:
   ```bash
   git fetch origin
   git rebase origin/main
   ```

3. **Squash WIP commits**:
   ```bash
   git rebase -i HEAD~N  # N = number of commits to squash
   ```

### PR Title Format

Follow the same format as commits:

```
feat(veid): add enhanced identity scoring
fix(market): resolve bid ordering race condition
docs: add developer onboarding guide
```

### PR Description Template

```markdown
## Summary

Brief description of changes.

## Changes

- Added X
- Modified Y
- Fixed Z

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Related Issues

Closes #123
Related to #456

## Checklist

- [ ] Code follows project style guidelines
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] Commits are signed off
- [ ] Commit messages follow conventions
```

### Review Process

1. **Automated checks** - CI runs tests, lints, and builds
2. **Code review** - At least 1 approval required (2 for critical paths)
3. **Address feedback** - Push additional commits or amend
4. **Final approval** - Maintainer approves and merges

### After Merge

- Delete your feature branch
- Update local main: `git pull origin main`
- Close related issues if not auto-closed

---

## Code Standards

### Go Code Style

Follow the standard Go style with these additions:

```go
// Package comments are required
// Package keeper implements the market module's state management.
package keeper

// Exported types need documentation
// Keeper provides access to the market module's state.
type Keeper struct {
    cdc       codec.BinaryCodec
    storeKey  storetypes.StoreKey
    authority string
}

// Exported methods need documentation
// CreateOrder creates a new marketplace order.
func (k Keeper) CreateOrder(ctx sdk.Context, msg *types.MsgCreateOrder) (*types.Order, error) {
    // Implementation
}
```

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to create order: %w", err)
}

// Good: Use sentinel errors for expected conditions
var ErrOrderNotFound = errors.New("order not found")

// Good: Return early on errors
order, err := k.GetOrder(ctx, orderID)
if err != nil {
    return nil, err
}
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, short | `keeper`, `types` |
| Interfaces | -er suffix or I prefix | `IKeeper`, `Reader` |
| Structs | PascalCase | `OrderState` |
| Methods | PascalCase | `CreateOrder` |
| Variables | camelCase | `orderID` |
| Constants | PascalCase or ALL_CAPS | `MaxOrderQuantity` |

### Imports Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party
    "github.com/cosmos/cosmos-sdk/types"
    
    // Internal packages
    "github.com/virtengine/virtengine/x/market/types"
)
```

### Linting

Run linting before committing:

```bash
make lint-go
```

Configuration is in `.golangci.yml`.

---

## Documentation Requirements

### Code Documentation

- All exported types, functions, and methods must have doc comments
- Complex algorithms should have inline comments
- Include examples for non-obvious usage

### README Updates

Update README when:
- Adding new features
- Changing CLI commands
- Modifying configuration options
- Adding new dependencies

### Changelog

Add entries to CHANGELOG.md for:
- New features
- Bug fixes
- Breaking changes
- Deprecations

Format:
```markdown
## [Unreleased]

### Added
- New identity verification flow (#123)

### Fixed
- Bid race condition in market module (#456)

### Changed
- Updated escrow timeout handling (#789)
```

---

## License and Copyright

### License

VirtEngine is licensed under Apache 2.0.

### Copyright Headers

Add copyright headers to new files:

```go
// Copyright (c) VirtEngine Author(s) 2019. All rights reserved.
// Licensed under the Apache 2.0 license. See LICENSE file in the project root for full license information.
```

### Contribution Agreement

By contributing, you agree to:
- License your contribution under Apache 2.0
- Grant the project rights to use your contribution
- Certify the Developer Certificate of Origin (via sign-off)

---

## Proposal Process

For significant changes, submit a proposal first.

### What Requires a Proposal

- New modules
- Breaking API changes
- Major architectural changes
- New external dependencies
- Protocol changes

### Proposal Template

```markdown
# Proposal: [Title]

## Summary

One paragraph explaining the proposal.

## Motivation

Why is this needed? What problem does it solve?

## Design

Technical details of the proposed implementation.

## Alternatives Considered

What other approaches were evaluated?

## Impact

- Migration requirements
- Breaking changes
- Performance implications
- Security considerations

## Timeline

Estimated effort and milestones.
```

### Proposal Workflow

1. Open issue with proposal
2. Discussion period (minimum 1 week)
3. Core team review
4. `design/approved` label added
5. Implementation can begin

---

## Quick Reference

### Common Commands

```bash
# Build
make virtengine

# Test
go test ./x/... ./pkg/...
make test-integration

# Lint
make lint-go

# Format
go fmt ./...

# Generate mocks
make generate
```

### Checklist Before PR

- [ ] Tests pass locally
- [ ] Linting passes
- [ ] Commits are signed off
- [ ] Commit messages follow conventions
- [ ] Documentation updated if needed
- [ ] PR description is complete
- [ ] Related issues are linked

---

## Related Documentation

- [Testing Guide](./04-testing-guide.md) - Writing and running tests
- [Code Review Checklist](./05-code-review-checklist.md) - Review standards
- [Module Development](./06-module-development.md) - Building modules
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Full contribution guidelines
