# VirtEngine repo guide for agents

## CRITICAL

{READ THE FOLLOWING}
Always work on tasks longer than you think are needed to accomodate edge case scenarios, testing, ensuring it works, implementing connecting logic and ensuring high quality of standards.

ALWAYS Ensure tests pass, build is clear with 0 warnings before deciding a task is fully completed and ready for merge. This is non negotiable.

When working on a task, do not stop until it is COMPLETELY done. Continue working on this task until it is 100% DONE END-TO-END with no questions asked.

Before finishing a task - ensure that you create a commit based on following convention (Ensuring Linting and Formatting are done precommit if possible) & trigger a git push --set-upstream origin ve/branch-name & git push passess all prepush hooks!

If you are running **outside** vibe-kanban (no `VE_TASK_TITLE` env var), you do not need to push changes.
You might notice that your task instructions might have already been implemented by another agent previously - if this is the case, your goal switches to analyzing if the previous agent has completed the task END TO END sufficiently - and if not, fix any issues or problems the previous agent might have missed.
If you notice unexpected modified/untracked files, assume another agent/subagent is working on a separate task; do NOT revert them. Continue with your task and avoid touching those files unless explicitly instructed.

### PR Creation & Merge (vibe-kanban automation)

If you are running as a vibe-kanban task agent (you'll have `VE_TASK_TITLE` and `VE_BRANCH_NAME` env vars set), you are responsible for **creating the PR** after your push. The **orchestrator handles merges** when CI passes. In this case:

- Focus on code quality, tests, and a clean commit
- After committing, sometimes a precommit will automatically trigger formatting changes (prettier, lint, etc) - please add these files to a secondary commit.
- Run `gh pr create` after your push to open the PR (do not bypass prepush hooks, if the issue is caused by an upstream branch - fix it)
- Ensure you consistently merge upstream changes before any git push, and fix conflicts if they exist
- Do NOT manually run `gh pr merge` (orchestrator merges after CI)

If you are running **outside** vibe-kanban (no `VE_TASK_TITLE` env var), you do not need to push changes.

If you are running **inside** vibe-kanban, and your prompt does not contain enough detail - check if the task exists already in \_docs/ralph/tasks folder

You should have all commands as needed available in shell, for example go, gh, pip, npm, git, etc. Consider increasing time outs when running long running commands such as git push, go test when running large test packages (running test on all packages could need more than 20minute timeout, only run tests on modules you actually changed instead), etc. Avoid running long CLI tasks when unnecessary, do not bypass verifications for git commit & git push - resolve any lint or unit test errors that you may encounter with these hooks.

## Agent-Specific Instructions

- **Codex agents:** See `.codex/instructions.md` for Codex-specific tooling, sandbox constraints, and workflow.
- **Copilot agents:** See `.github/copilot-instructions.md` for VS Code integration, MCP servers, and module patterns.

## MANDATORY Pre-Push Checklist

**Every agent MUST complete the checklist in [.github/AGENT_PREFLIGHT.md](.github/AGENT_PREFLIGHT.md) before committing and pushing.**

Failure to follow this checklist is the #1 cause of failed tasks. The pre-push hooks WILL reject your push if these steps are skipped.

You can also run the automated pre-flight script before pushing:

- **Linux/macOS/WSL:** `./scripts/agent-preflight.sh`
- **Windows (PowerShell):** `pwsh scripts/agent-preflight.ps1`

## Pre-commit automation (do this every time)

- Go formatting and linting are enforced before commit. The pre-commit hook auto-runs `gofmt` on staged `.go` files and runs `golangci-lint` on the staged Go packages.
- Portal frontend formatting is enforced before commit. If you modify `portal/` TypeScript/JS/CSS/JSON/MD files, ensure `portal/node_modules` exists (run `pnpm -C portal install` once) so the pre-commit hook can run Prettier and auto-add formatted files to the commit.
- SDK TypeScript formatting/linting is enforced before commit. If you modify `sdk/ts` files, ensure `sdk/ts/node_modules` exists (run `pnpm -C sdk/ts install` once) so `lint-staged` can auto-fix and stage changes.
- If you need to bypass a check for an emergency, use the documented `VE_HOOK_SKIP_*` env vars, but do not bypass for normal work.

## Pre-push quality gate (smart — runs only relevant checks)

The pre-push hook detects which files changed and only runs checks for the affected categories:

**Go changes** (`.go` files or `go.mod`/`go.sum`):

- `go vet` on changed packages
- `gofmt` auto-format
- `golangci-lint` (diff-only)
- `go mod vendor` sync
- Build binaries (`make bins`)
- Go unit tests (changed packages only)

**Portal/Frontend changes** (`portal/`, `lib/portal/`, `lib/capture/`, `lib/admin/`):

- Prettier auto-format
- ESLint (`pnpm -C portal lint`) — mirrors **Portal CI / Lint & Type Check**
- TypeScript type-check (`pnpm -C portal type-check`) — mirrors **Portal CI / Lint & Type Check**
- Portal unit tests (`pnpm -C portal test` + `pnpm -C lib/portal test`) — mirrors **Portal CI / Unit Tests**

**JS dependency changes** (`pnpm-lock.yaml`, `package.json`):

- pnpm lockfile validation (`pnpm install --frozen-lockfile`)
- All portal checks above (lint, type-check, tests)

**Docs-only changes** (`.md`, `_docs/`, `docs/`, `.github/`):

- No checks — push proceeds immediately

**Skip env vars:**

- `VE_HOOK_SKIP_PORTAL=1` — skip all portal checks (ESLint, TypeScript, tests)
- `VE_HOOK_SKIP_VET=1`, `VE_HOOK_SKIP_FMT=1`, `VE_HOOK_SKIP_LINT=1`, `VE_HOOK_SKIP_BUILD=1`, `VE_HOOK_SKIP_TEST=1`, `VE_HOOK_SKIP_MOD=1`, `VE_HOOK_SKIP_PNPM=1`
- `VE_HOOK_QUICK=1` — vet + build only (Go)

Commit files:

📝 Conventional Commits Format:
type(scope): description

✅ Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

📌 Examples:

feat(veid): add identity verification flow

fix(market): resolve bid race condition

docs: update contributing guidelines

chore(deps): bump cosmos-sdk to v0.53.1

⚠️ Breaking changes: add '!' after type/scope

feat(api)!: change response format

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

## GSD Framework & Orchestration

This repo uses the GSD (Get Stuff Done) framework for autonomous development with dual-agent orchestration.

**Orchestrator Workflow (`scripts/codex-monitor/ve-orchestrator.ps1`):**

1.  **Task Source:** Vibe-Kanban board (auto-detected by project name "virtengine").
2.  **Agent Cycling:** Alternates 50/50 between Codex (DEFAULT profile) and Copilot (CLAUDE_OPUS_4_6 profile) to avoid rate-limiting.
3.  **Merge Gate:** New tasks only start after previous task PRs are merged and confirmed.
4.  **Executor:**
    - Codex agents: Use shell + codex-cli MCP for subtasks. See `.codex/instructions.md`.
    - Copilot agents: Use VS Code tools + `runSubagent`. See `.github/copilot-instructions.md`.
    - Both: **MUST** pass pre-push hooks (go vet, lint, build, tests) before task completion.
5.  **Cleanup:** Vibe-kanban cleanup script handles PR creation + auto-merge.

## MCP Servers & Tool Usage

VirtEngine integrates several Model Context Protocol (MCP) servers for enhanced development capabilities. Use these tools when appropriate for your task:

### Context7 - Library Documentation

**When to use:**

- Looking up documentation for any programming library or framework
- Need code examples for specific APIs (Cosmos SDK, TensorFlow, gRPC, etc.)
- Understanding third-party dependencies or their usage patterns
- Researching best practices for libraries used in the project

**Examples:**

- "How do I use Cosmos SDK's keeper pattern?"
- "Show me TensorFlow deterministic ops configuration"
- "What's the correct way to implement IBC message handlers?"

### GitHub MCP Server

**When to use:**

- Creating or managing issues and pull requests
- Searching for code patterns across GitHub repositories
- Reviewing PR diffs, comments, and reviews
- Managing branches, commits, and releases
- Searching for issues, PRs, or code in VirtEngine or related repos

**Examples:**

- "Create an issue for the identity verification bug"
- "Search for usage of EncryptionEnvelope across the codebase"
- "Show me recent PRs to the mainnet/main branch"
- "Find implementations of IKeeper interface in Cosmos SDK repos"

### Playwright MCP - Browser Automation & Testing

**When to use:**

- Testing web UIs or frontend components
- Automating browser-based workflows (e.g., testing provider dashboard)
- Scraping documentation or web content for development
- E2E testing scenarios involving web interfaces

**Examples:**

- "Test the provider registration flow in the web UI"
- "Automate checking the validator dashboard"
- "Verify the VEID identity upload workflow"

### Vibe-Kanban - Task & Project Management

**When to use:**

- Creating or updating tasks/tickets for development work
- Listing project tasks and their status
- Starting workspace sessions for specific tasks
- Managing repositories and their scripts (setup, cleanup, dev server)

**Examples:**

- "Create a task for implementing HPC module tests"
- "List all in-progress tasks"
- "Update the VEID encryption task to completed"

### Exa - Web Search & Code Context

**When to use:**

- Searching the web for up-to-date information (APIs, protocols, security advisories)
- Finding code examples or implementations from public sources
- Researching new technologies or approaches not in local documentation
- Getting real-time context about external dependencies or standards

**Examples:**

- "Search for recent Cosmos SDK v0.53 migration guides"
- "Find examples of X25519 encryption in Go"
- "Look up the latest CometBFT security advisories"

### Chrome DevTools MCP

**When to use:**

- Debugging browser-based components or frontends
- Inspecting network requests from web clients
- Profiling performance of web UIs
- Capturing console logs or errors from browser environments

**Examples:**

- "Inspect the network requests when connecting to the provider daemon"
- "Profile the provider dashboard page load performance"
- "Capture console errors from the identity verification UI"

### Dev Manager MCP

**When to use:**

- Managing development environment setup
- Coordinating multi-service local development
- Handling development server lifecycles
- Managing development tooling and configurations

**Examples:**

- "Set up the development environment for provider daemon"
- "Start all required services for local testing"
- "Configure the development environment for ML pipeline work"

### General MCP Usage Guidelines

1. **Choose the right tool:** Select the MCP server that best matches your task's domain
2. **Combine when needed:** Some tasks may benefit from multiple MCP servers (e.g., GitHub + Context7 for researching API usage in other repos)
3. **Respect rate limits:** If one tool is rate-limited, switch to alternatives or manual approaches
4. **Security first:** Never expose sensitive data (API keys, private keys, credentials) through MCP tools
5. **Verify results:** Always validate outputs from MCP tools, especially for code examples or documentation

### Task-to-Tool Mapping

| Task Domain               | Primary Tool             | Secondary Tool   |
| ------------------------- | ------------------------ | ---------------- |
| Library/API Documentation | Context7                 | Exa (web search) |
| GitHub Operations         | GitHub MCP               | -                |
| Task Management           | Vibe-Kanban              | -                |
| Web Testing               | Playwright               | Chrome DevTools  |
| Research/Discovery        | Exa                      | Context7         |
| Development Setup         | Dev Manager              | -                |
| Code Search (external)    | GitHub MCP (search_code) | Exa              |
| Browser Debugging         | Chrome DevTools          | Playwright       |
