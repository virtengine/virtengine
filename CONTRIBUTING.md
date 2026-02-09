# Contributing

## Guidelines

Guidelines for contributing.

### I want to contribute on GitHub

#### I've found a typo

- A Pull Request is not necessary. Raise an [Issue](https://github.com/virtengine/support/issues) and we'll fix it as soon as we can.

#### I have a (great) idea

The VirtEngine maintainers would like to make VirtEngine the best it can be and welcome new contributions that align with the project's goals. Our time is limited so we'd like to make sure we agree on the proposed work before you spend time doing it. Saying "no" is hard which is why we'd rather say "yes" ahead of time. You need to raise a proposal.

Every feature carries a cost - a cost if developed wrong, a cost to carry and maintain it and if it wasn't needed in the first place then this is an unnecessary burden. See [Yagni from Martin Fowler](https://martinfowler.com/bliki/Yagni.html). The best proposals are defensible with real data and are more than a hypothesis.

**Please do not raise a proposal after doing the work - this is counter to the spirit of the project. It is hard to be objective about something which has already been done**

What makes a good proposal?

- Brief summary including motivation/context
- Any design changes
- Pros + Cons
- Effort required up front
- Effort required for CI/CD, release, ongoing maintenance
- Migration strategy / backwards-compatibility
- Mock-up screenshots or examples of how the CLI would work
- Clear examples of how to reproduce any issue the proposal is addressing

Once your proposal receives a `design/approved` label you may go ahead and start work on your Pull Request.

If you are proposing a new tool or service please do due diligence. Does this tool already exist in a 3rd party project or library? Can we reuse it? For example: a timer / CRON-type scheduler for invoking functions is a well-solved problem, do we need to reinvent the wheel?

Every effort will be made to work with contributors who do not follow the process. Your PR may be closed or marked as `invalid` if it is left inactive, or the proposal cannot move into a `design/approved` status.

#### Paperwork for Pull Requests

Ensure that you base and target your PR on the `main` branch.
All feature additions and all bug fixes must be targeted against main. Exception is for bug fixes which are only related to a released version. In that case, the related bug fix PRs must target against the release branch.
If needed, we backport a commit from main to a release branch (excluding consensus breaking feature, API breaking and similar).

#### Testing

Tests can be executed by running `make test` at the top level of the Cosmos SDK repository.

Please follow style guide on [this blog post](https://blog.alexellis.io/golang-writing-unit-tests/) from [The Go Programming Language](https://www.amazon.co.uk/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440)

If you are making changes to code that use goroutines, consider adding `goleak` to your test to help ensure that we are not leaking any goroutines. Simply add

```go
defer goleak.VerifyNoLeaks(t)
```

at the very beginning of the test, and it will fail the test if it detects goroutines that were opened but never cleaned up at the end of the test.

#### Required Checks and Branch Protection

All pull requests must pass the following required checks before merge:

**Quality Gate (quality-gate.yaml):**
- ✅ **quality-gate / lint** - golangci-lint with all enabled linters
- ✅ **quality-gate / vet** - go vet on all packages
- ✅ **quality-gate / build** - all binaries must build successfully
- ✅ **quality-gate / test-go** - core package tests must pass
- ✅ **quality-gate / gosec-scan** - gosec with severity gating (if Go files changed)
- ✅ **quality-gate / gitleaks-scan** - gitleaks on full PR diff (blocking)
- ⚠️ **quality-gate / agents-docs** - AGENTS docs validation (warning only)

**ML Determinism (quality-gate.yaml):** (if ML code changed)
- ✅ **quality-gate / ml-determinism-check** - Go inference determinism verification

**Portal CI (quality-gate.yaml):** (if portal code changed)
- ✅ **quality-gate / portal-lint** - ESLint and TypeScript checks
- ✅ **quality-gate / portal-test** - Unit tests
- ✅ **quality-gate / portal-build** - Next.js build

**Security Details:**

- **gosec baseline**: New high/critical severity findings fail the build. Medium findings generate warnings. Baseline is tracked in `.gosec-baseline.json`.
- **gitleaks**: Scans full PR diff (not just last commit). Allowlist for known false positives in `.gitleaks.toml`.
- **ML determinism**: Critical for blockchain consensus - ensures inference is reproducible across platforms with CPU-only mode and fixed seed.

**Setting Up Branch Protection:**

Repository maintainers should configure branch protection for `main` and `mainnet/main` with:

```
Required status checks:
- quality-gate / lint
- quality-gate / vet
- quality-gate / build
- quality-gate / test-go
- quality-gate / gosec-scan (if Go files changed)
- quality-gate / gitleaks-scan
- quality-gate / ml-determinism-check (if ML files changed)
- quality-gate / portal-lint (if portal changed)
- quality-gate / portal-test (if portal changed)
- quality-gate / portal-build (if portal changed)

Require branches to be up to date before merging: Yes
Require linear history: Yes (recommended)
```

If you don't have permissions to configure branch protection via the GitHub API, ensure the repository admin manually configures these settings via GitHub Settings → Branches → Branch protection rules.

#### I have a question, a suggestion or need help

If you have a simple question you can [join the VirtEngine community](https://virtengine.com/community) and ask there, but please bear in mind that contributors may live in a different timezone or be working to a different timeline to you. If you have an urgent request then let them know about this.

If you have a deeply technical request or need help debugging your application then you should prepare a simple, public GitHub repository with the minimum amount of code required to reproduce the issue.

If you feel there is an issue with VirtEngine or were unable to get the help you needed from the Slack channels then raise an issue on one of the GitHub repositories.

#### Setting expectations, support and SLAs

- What kind of support can I expect for free?

  If you are using one of the Open Source projects within the virtengine repository, then help is offered on a good-will basis by volunteers. You can also request help from employees of DET-IO Pty. Ltd. who host the VirtEngine testnet.

  Please be respectful of volunteer time, it is often limited to evenings and weekends. The person you are requesting help from may not reside in your timezone.

  The VirtEngine chat is the best place to ask questions, suggest features, and to get help. The GitHub issue tracker can be used for suspected issues with the codebase or deployment artifacts.

- What is the SLA for my Issue?

  Issues are examined, triaged and answered on a best effort basis by volunteers and community contributors. This means that you may receive an initial response within any time period such as: 1 minute, 1 hour, 1 day, or 1 week. There is no implicit meaning to the time between you raising an issue and it being answered or resolved.

  If you see an issue which does not have a response or does not have a resolution, it does not mean that it is not important, or that it is being ignored. It simply means it has not been worked on by a volunteer yet.

  Please take responsibility for following up on your Issues if you feel further action is required.

- What is the SLA for my Pull Request?

  In a similar way to Issues, Pull Requests are triaged, reviewed, and considered by a team of volunteers - the Core Team, Members Team and the Project Lead. There are dozens of components that make up the VirtEngine project and a limited amount of people. Sometimes PRs may become blocked or require further action.

  Please take responsibility for following up on your Pull Requests if you feel further action is required.

- Why may your PR be delayed?
  - The contributing guide was not followed in some way

  - The commits are not signed-off

  - The commits need to be rebased

  - Changes have been requested

  More information, a use-case, or context may be required for the change to be accepted.

- What if I need more than that?

  If you're a company using any of these projects, you can get the following through a support agreement with DET-IO Pty. Ltd. so that the time can be paid for to help your business.

  A support agreement can be tailored to your needs, you may benefit from support, if you need any of the following:
  - responses within N hours/days on issues/PRs
  - feature prioritization
  - urgent help
  - 1:1 consultations
  - or any other level of professional services

#### I need to add a dependency

The concept of `vendoring` is used in projects written in Go. This means that a copy of the source-code of dependencies is stored within each repository in the `vendor` folder. It allows for a repeatable build and isolates change.

The chosen tool for vendoring code in the project is [dep](https://github.com/golang/dep).

> Note: despite the availability of [Go modules](https://github.com/golang/go/wiki/Modules) in Go 1.11, they are not being used in the project at this time. If and when the decision is made to move, a complete overhaul of all repositories will need to be made in a coordinated fashion, including regression and integration testing. This is not a trivial task.

### How are releases made?

For detailed release management procedures, see [RELEASE.md](./RELEASE.md).

**Quick Summary:**

1. **Branching Strategy:**
   - `main` branch for development (odd minor versions like v0.9.x)
   - `mainnet/main` branch for production (even minor versions like v0.10.x)

2. **Release Process:**
   - Release candidates are published for community testing (e.g., v0.10.0-rc.1)
   - After validation, stable releases are tagged (e.g., v0.10.0)
   - GitHub Actions triggers GoReleaser to build and publish

3. **Artifacts Published:**
   - GitHub Releases: binaries for Linux (amd64, arm64) and macOS (universal)
   - Docker images: `ghcr.io/virtengine/virtengine:<version>`
   - Homebrew: `virtengine/homebrew-tap` (mainnet stable releases only)

4. **Requesting a Release:**
   - For bug fixes: mention urgency in your PR
   - For features: coordinate with release manager via GitHub Issue
   - For emergency fixes: contact core team directly

See also:

- [RELEASE.md](./RELEASE.md) - Complete release management process
- [\_docs/version-control.md](./_docs/version-control.md) - Version control practices
- [ADR-001: Network Upgrades](./_docs/adr/adr-001-network-upgrades.md) - Network upgrade implementation

## Governance

VirtEngine is an independent open-source project which was created by the DET-IO Pty. Ltd. in 2016. The project is maintained and developed by DET-IO Pty. Ltd.

DET-IO hosts and sponsors the development and maintenance of VirtEngine. DET-IO provides professional services, consultation and support. Contact us at [virtengine.com/contact](https://virtengine.com/contact) to find out more.

#### Project Lead

Responsibility for the project starts with the _Project Lead_, who delegates specific responsibilities and the corresponding authority to the Core and Members team.

Some duties include:

- Setting overall technical & community leadership
- Engaging end-user community to advocate needs of end-users and to capture case-studies
- Defining and curating roadmap for VirtEngine
- Building a community and team of contributors
- Community & media briefings, out-bound communications, partnerships, relationship management and events

### How do I become a maintainer?

In the VirtEngine community there are three levels of structure or maintainership:

- Core Team (GitHub org)
- The rest of the community.

#### Core Team

The VirtEngine system was designed and is primarily being led by Jonathan Philipos.

## Community

This project is written in Golang but many of the community contributions so far have been through blogging, speaking engagements, helping to test and drive the backlog of VirtEngine. If you'd like to help in any way then that would be more than welcome whatever your level of experience.

### Roadmap

- Browse open issues in [support](https://github.com/virtengine/support/issues)

## License

This project is licensed under the Apache 2.0 License.

### Copyright notice

It is important to state that you retain copyright for your contributions, but agree to license them for usage by the project and author(s) under the Apache 2.0 license. Git retains history of authorship, but we use a catch-all statement rather than individual names.

Please add a Copyright notice to new files you add where this is not already present.

```
// Copyright (c) VirtEngine Author(s) 2019. All rights reserved.
// Licensed under the Apache 2.0 license. See LICENSE file in the project root for full license information.
```

### Sign your work

> Note: every commit in your PR or Patch must be signed-off.

The sign-off is a simple line at the end of the explanation for a patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](http://developercertificate.org/)):

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages. This ensures a consistent commit history and enables automatic changelog generation.

#### Format

```
type(scope): description

[optional body]

[optional footer(s)]
```

#### Types

| Type       | Description                                           |
| ---------- | ----------------------------------------------------- |
| `feat`     | A new feature                                         |
| `fix`      | A bug fix                                             |
| `docs`     | Documentation only changes                            |
| `style`    | Changes that do not affect the meaning of the code    |
| `refactor` | Code change that neither fixes a bug nor adds feature |
| `perf`     | Performance improvement                               |
| `test`     | Adding missing tests or correcting existing tests     |
| `build`    | Changes affecting build system or external deps       |
| `ci`       | Changes to CI configuration files and scripts         |
| `chore`    | Other changes that don't modify src or test files     |
| `revert`   | Reverts a previous commit                             |

#### Scopes (Optional)

Use module names as scopes when applicable:

- **Blockchain modules:** `veid`, `mfa`, `encryption`, `market`, `escrow`, `roles`, `hpc`
- **Infrastructure:** `provider`, `sdk`, `cli`, `app`
- **Development:** `deps`, `ci`, `api`, `ml`, `tests`

#### Examples

```bash
# Feature with scope
feat(veid): add identity verification flow

# Bug fix with scope
fix(market): resolve bid race condition

# Documentation without scope
docs: update contributing guidelines

# Breaking change (note the !)
feat(api)!: change response format for provider endpoints

# Dependency update
chore(deps): bump cosmos-sdk to v0.53.1

# With body and footer
fix(escrow): handle timeout edge case

The escrow module was not properly handling timeouts when
the provider went offline during lease finalization.

Closes #123
```

#### Local Validation

Install commitlint for local validation:

```bash
npm install --save-dev @commitlint/cli @commitlint/config-conventional
```

The repository includes a `commitlint.config.js` with project-specific rules.

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

Please sign your commits with `git commit -s` so that commits are traceable.

This is different from digital signing using GPG, GPG is not required for
making contributions to the project.

If you forgot to sign your work and want to fix that, see the following
guide: [Git: Rewriting History](https://git-scm.com/book/en/v2/Git-Tools-Rewriting-History)
