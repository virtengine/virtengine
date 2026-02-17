# VirtEngine AGENTS System — Repo Guide

## Module Overview

- Purpose: Describe how AGENTS.md files document modules and how agents should work in this repo.
- Use when: Updating documentation, adding new modules, or onboarding contributors and agents.
- Key entry points:
  - Template: `docs/templates/AGENTS.template.md`
  - Standards: `docs/AGENTS_STANDARDS.md`
  - Index: `docs/AGENTS_INDEX.md`
  - Validator: `scripts/validate-agents-docs.mjs`

### CRITICAL

{READ THE FOLLOWING}
Always work on tasks longer than you think are needed to accomodate edge case scenarios, testing, ensuring it works, implementing connecting logic and ensuring high quality of standards.

NEVER START A PLAN AND THEN ASK FOR USER INPUT, OR ASK FOR USER INPUT DURING A COMPLETION OF A TASK - ONCE YOU ARE ASSIGNED A PROPER TASK - YOU MUST COMPLETE IT END TO END. IF YOU ARE ASSIGNED A TASK OF AN ALREADY COMPLETED TASK THAT YOU ARE ANALYZING TO ENSURE COMPLETENESS, IF YOU DO FIND GAPS DON'T ASK FOR CONFIRMATIONS - PROCEED TO COMPLETE THE MISSING GAPS THAT YOU IDENTIFIED DURING YOUR ANALYSIS.

ALWAYS Ensure tests pass, build is clear with 0 warnings before deciding a task is fully completed and ready for merge. This is non negotiable.

When working on a task, do not stop until it is COMPLETELY done. Continue working on this task until it is 100% DONE END-TO-END with no questions asked.

If you are COPILOT & HAVE ACCESS TO COPILOT SUBAGENT: MAKE SURE YOU USE SUBAGENTS, DO NOT MAKE CHANGES DIRECTLY - CONTAIN YOUR CONTEXT AND ACT AS AN ORCHESTRATOR - RUN THEM IN PARALLEL. IF YOU ARE LAUNCHED FROM CODEX, THEN YOU DO NOT HAVE SUBAGENTS - BUT YOU CAN STILL USE CODEX-CLI MCP SERVERS TO LAUNCH SUBAGENTS IF NEEDED.

## NEVER PLAN AND NOT EXECUTE THE TASKS MADE IN THE PLAN

DO NOT ASK USER FOR CONFIRMATION TO CONTINUE WITH THE PLAN YOU JUST MADE, IF YOU USED [[PLAN]] FIRST -> THEN CONTINUE AND
EXECUTE THE TASK IMMEDIATELY.

### Completing a TASK <IMPORTANT GIT INFO>

Before finishing a task - ensure that you create a commit based on following convention (Ensuring Linting and Formatting are done precommit, such as gofmt, prettier, golangci-lint) & git merge origin <<upstream branch>> --no-edit & trigger a git push --set-upstream origin ve/branch-name & git push passess all prepush hooks! <<upstream branch>> is the default branch that your branch was created from - MAKE sure that you have MERGED any upstream changes before pushing as otherwise there could be an unaddressed CONFLICT. <<URGENT>>

If you are running **outside** vibe-kanban (no `VE_TASK_TITLE` env var), you do not need to push changes.

You might notice that your task instructions might have already been implemented by another agent previously - if this is the case, your goal switches to analyzing if the previous agent has completed the task END TO END sufficiently - and if not, fix any issues or problems the previous agent might have missed.

If you notice unexpected modified/untracked files, assume another agent/subagent is working on a separate task; do NOT revert them. Continue with your task and avoid touching those files unless explicitly instructed.

## Architecture

- AGENTS.md files live alongside modules and summarize key entry points, concepts, and testing steps.
- System documentation lives in `docs/`: & development oriented documentation lives in `_docs/` 
  - Template: `docs/templates/AGENTS.template.md`
  - Standards: `docs/AGENTS_STANDARDS.md`
  - Index: `docs/AGENTS_INDEX.md`
- Module-specific docs (examples):
  - `api/AGENTS.md`
  - `sdk/ts/AGENTS.md`
  - `x/provider/AGENTS.md`
  - `scripts/codex-monitor/AGENTS.md`

## Core Concepts

- Every module AGENTS.md uses the required sections listed in the template.
- Code references use `path/to/file.ext:line` format to anchor guidance.
- Update AGENTS.md in the same PR as code changes to keep documentation in sync.

## Usage Examples

### Find module guidance

1. Open `docs/AGENTS_INDEX.md`.
2. Jump to the module link you need.

### Reference a code entry point

- Example: `x/provider/keeper/keeper.go:48`

## Implementation Patterns

- Adding a new module AGENTS.md:
  1. Copy `docs/templates/AGENTS.template.md` into the module directory.
  2. Replace placeholders with module-specific content.
  3. Add the file to `docs/AGENTS_INDEX.md`.
  4. Run `node scripts/validate-agents-docs.mjs`.

### PR Creation & Merge (vibe-kanban automation)

If you are running as a vibe-kanban task agent (you'll have `VE_TASK_TITLE` and `VE_BRANCH_NAME` env vars set), you are responsible for **creating the PR** after your push. The **orchestrator handles merges** when CI passes. In this case:

- Focus on code quality, tests, and a clean commit
- Before commiting - always run auto formatting tools such as prettier, lint, etc.
- After committing, sometimes a precommit will automatically trigger formatting changes (prettier, lint, etc) - please add these files to a secondary commit.
- Ensure you consistently merge upstream changes (from the original branch your branch was created from) before any git push, and fix conflicts if they exist <IMPORTANT>
- Run `gh pr create` after your push to open the PR (do not bypass prepush hooks, if the issue is caused by an upstream branch - fix it)
- Do NOT manually run `gh pr merge` (orchestrator merges after CI)
  If you are running **outside** vibe-kanban (no `VE_TASK_TITLE` env var), you do not need to push changes.

You should have all commands as needed available in shell, for example any powershell commands or any go, gh, pip, npm, git, etc. Consider increasing time outs when running long running commands such as git push, go test when running large test packages (running test on all packages could need more than 20minute timeout, only run tests on modules you actually changed instead), etc. Avoid running long CLI tasks when unnecessary, do not bypass verifications for git commit & git push - resolve any lint or unit test errors that you may encounter with these hooks.

### Commit Message Conventions

VirtEngine uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages and PR titles.

Format:

```
type(scope): description
```

Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Valid scopes: `veid`, `mfa`, `encryption`, `market`, `escrow`, `roles`, `hpc`, `provider`, `sdk`, `cli`, `app`, `deps`, `ci`, `api`

### Agent-Specific Instructions

- **Codex agents:** See `.codex/instructions.md` for Codex-specific tooling, sandbox constraints, and workflow.
- **Copilot agents:** See `.github/copilot-instructions.md` for VS Code integration, MCP servers, and module patterns.

### Branch Strategy

- `main` - active development (odd minor versions like v0.9.x)
- `mainnet/main` - stable releases (even minor versions like v0.8.x)

## Configuration

- AGENTS.md files have no runtime configuration.
- The validator runs from repo root with Node.js and no additional config.

## Testing

### MANDATORY Pre-Push Checklist

**Every agent MUST complete the checklist in [.github/AGENT_PREFLIGHT.md](.github/AGENT_PREFLIGHT.md) before committing and pushing.**

Failure to follow this checklist is the #1 cause of failed tasks. The pre-push hooks WILL reject your push if these steps are skipped.

You can also run the automated pre-flight script before pushing:

- **Linux/macOS/WSL:** `./scripts/agent-preflight.sh`
- **Windows (PowerShell):** `pwsh scripts/agent-preflight.ps1`

### Pre-commit automation (do this every time)

- Go formatting and linting are enforced before commit. The pre-commit hook auto-runs `gofmt` on staged `.go` files and runs `golangci-lint` on the staged Go packages.
- Portal frontend formatting is enforced before commit. If you modify `portal/` TypeScript/JS/CSS/JSON/MD files, ensure `portal/node_modules` exists (run `pnpm -C portal install` once) so the pre-commit hook can run Prettier and auto-add formatted files to the commit.
- SDK TypeScript formatting/linting is enforced before commit. If you modify `sdk/ts` files, ensure `sdk/ts/node_modules` exists (run `pnpm -C sdk/ts install` once) so `lint-staged` can auto-fix and stage changes.
- If you need to bypass a check for an emergency, use the documented `VE_HOOK_SKIP_*` env vars, but do not bypass for normal work.

### Pre-push quality gate (smart — runs only relevant checks)

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

### Build & Test References

```bash
make virtengine

go test ./x/... ./pkg/...

go test -tags="e2e.integration" ./tests/integration/...

make test-integration
```

### Documentation Validation

- Run `node scripts/validate-agents-docs.mjs` before push.

## Troubleshooting

- Validator fails on missing section: ensure required headings match the template.
- Validator fails on broken link: update the relative path or anchor.
- Pre-push hooks fail for portal: ensure `pnpm -C portal install` ran.
- Pre-push hooks fail for SDK: ensure `pnpm -C sdk/ts install` ran.

## MEMORY FILE
