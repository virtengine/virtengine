# VirtEngine repo guide for agents

## Overview
- Primary language: Go (Cosmos SDK-based chain + services).
- Secondary components: Python ML pipelines and SDKs.
- Repo includes chain modules (`x/*`), shared packages (`pkg/*`), app wiring (`app/`), CLI (`cmd/`), ML tooling (`ml/`), and SDKs (`sdk/`).

## Key paths
- `app/`: app configuration and module wiring.
- `cmd/`: binaries (e.g., `provider-daemon`).
- `x/`: blockchain modules (keepers, types, msgs).
- `pkg/`: shared libraries and runtime integrations.
- `tests/`: integration/e2e tests.
- `ml/`: ML pipelines, training, and evaluation.
- `_docs/` and `docs/`: architecture, operations, and testing guides.
- `scripts/`: localnet and utility scripts.

## Environment & tooling
- Go 1.21.0+ for core builds; localnet/testing docs mention Go 1.22+.
- `make` is required; repo uses a `.cache` toolchain under the repo root.
- `direnv` is used for environment management; see `_docs/development-environment.md`.
- CGO dependencies exist (libusb/libhid), so a C/C++ compiler is required.

## Build
```bash
make virtengine
```
Build outputs go to `.cache/bin`.

## Tests
Unit tests:
```bash
go test ./x/... ./pkg/...
```
Integration tests:
```bash
go test -tags="e2e.integration" ./tests/integration/...
```
E2E tests:
```bash
make test-integration
```

For detailed guidance, see `_docs/testing-guide.md`.

## Localnet (integration environment)
```bash
./scripts/localnet.sh start
./scripts/localnet.sh test
```
Windows users should run localnet in WSL2 as noted in `_docs/development-environment.md`.

## Contribution rules
- Target PRs against `main` unless a release-branch bug fix.
- Use Conventional Commits; sign-off required (`git commit -s`).
- Add copyright headers to new files when missing.
- Follow proposal process for new features per `CONTRIBUTING.md`.

## Commit instructions
Conventional Commits format:
```
type(scope): description
```

Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Valid scopes: `veid`, `mfa`, `encryption`, `market`, `escrow`, `roles`, `hpc`, `provider`, `sdk`, `cli`, `app`, `deps`, `ci`, `api`

Examples:
```
feat(veid): add identity verification flow
fix(market): resolve bid race condition
docs: update contributing guidelines
chore(deps): bump cosmos-sdk to v0.53.1
```

Breaking changes: add `!` after type/scope
```
feat(api)!: change response format
```

## Repository hygiene
- Do not commit generated caches or large binaries (ML weights live under `ml/*/weights/` and should stay out of git).
- Prefer existing make targets/scripts; avoid reimplementing workflows.
