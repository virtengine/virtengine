# Codex-Monitor ? AGENTS Guide

This guide is a fast, code-referenced map of the `scripts/codex-monitor/` module for AI agents. Every section points to the exact source lines to keep changes safe and grounded.

## Module Overview
- Purpose: codex-monitor supervises VirtEngine's autonomous coding fleet ? it schedules task attempts, runs PR automation, self-heals failures, and reports status via Telegram.
- Use when: Updating task orchestration, PR automation, executor routing, or Telegram notification behaviors.
- Key entry points:
  - Supervisor loop: `scripts/codex-monitor/monitor.mjs:14`
  - CLI entry: `scripts/codex-monitor/cli.mjs:1`
  - Orchestrator loop: `scripts/codex-monitor/ve-orchestrator.ps1:144`
  - VK API wrapper: `scripts/codex-monitor/ve-kanban.ps1:1`

**Code references:**
- Supervisor imports and orchestration wiring: `scripts/codex-monitor/monitor.mjs:14`
- Task planner and preflight gating in monitor: `scripts/codex-monitor/monitor.mjs:400`
- Orchestrator state + loop metrics: `scripts/codex-monitor/ve-orchestrator.ps1:144`
- VK API wrapper and task attempt submission: `scripts/codex-monitor/ve-kanban.ps1:1`

## Architecture
- Entry points and data flow: `cli.mjs` ? `monitor.mjs` ? `ve-orchestrator.ps1` / `ve-kanban.ps1`.
- Key components:
  | Component | Role | Key references |
  |---|---|---|
  | **monitor.mjs** | Supervisor loop, smart PR flow, maintenance scheduling, fleet sync | `scripts/codex-monitor/monitor.mjs:4047` |
  | **telegram-bot.mjs** | Telegram polling, batching/live digest, command queueing | `scripts/codex-monitor/telegram-bot.mjs:1` |
  | **ve-orchestrator.ps1** | PowerShell task orchestration loop (parallel slots, retries, merge gate) | `scripts/codex-monitor/ve-orchestrator.ps1:4841` |
  | **ve-kanban.ps1** | VK CLI wrapper (list/submit/rebase/archive attempts) | `scripts/codex-monitor/ve-kanban.ps1:1` |
  | **fleet-coordinator.mjs** | Multi-workstation coordination, fleet state persistence | `scripts/codex-monitor/fleet-coordinator.mjs:1` |
  | **autofix.mjs** | Error loop detection + guarded auto-fix execution | `scripts/codex-monitor/autofix.mjs:1` |
  | **codex-shell.mjs** | Persistent Codex SDK agent sessions | `scripts/codex-monitor/codex-shell.mjs:1` |
  | **copilot-shell.mjs** | Persistent Copilot SDK agent sessions | `scripts/codex-monitor/copilot-shell.mjs:1` |
  | **config.mjs** | Unified config loader (CLI/env/.env/json/defaults) | `scripts/codex-monitor/config.mjs:4` |

## Core Concepts
- **Task lifecycle:** VK attempt submission ? slot scheduling ? PR checks/merge decisions ? completion updates.
- **Smart PR flow:** `smartPRFlow` handles branch status, stale detection, rebases, PR creation, and fast/slow failure handling.
- **Workspace hygiene:** automated cleanup of worktrees + metadata, including orphan detection and pruning.
- **Notification batching:** Telegram bot batches low-priority updates but pushes critical failures immediately.
- **Autofix guardrails:** repeated failure detection is capped with cooldowns to prevent infinite loops.

## Usage Examples

### Start codex-monitor with defaults
```bash
node scripts/codex-monitor/cli.mjs
```

### Run codex-monitor tests
```bash
cd scripts/codex-monitor
npm run test
```

## Implementation Patterns
- Adding a new executor:
  - Update executor config parsing: `scripts/codex-monitor/config.mjs:203`
  - Wire executor profiles in setup: `scripts/codex-monitor/setup.mjs:1284`
  - Add routing logic in `ve-kanban.ps1`: `scripts/codex-monitor/ve-kanban.ps1:323`
- Extending notification logic:
  - Add message formats in `telegram-bot.mjs`: `scripts/codex-monitor/telegram-bot.mjs:95`
  - Ensure monitor uses `notify` or `sendTelegramMessage`: `scripts/codex-monitor/monitor.mjs:4047`
- Modifying PR flow behavior:
  - Edit `smartPRFlow` + rebase filters: `scripts/codex-monitor/monitor.mjs:4047`
- Adding new autofix patterns:
  - Extend error parsing in `autofix.mjs`: `scripts/codex-monitor/autofix.mjs:66`

## Configuration
- Config loading order: CLI ? env vars ? `.env` ? `codex-monitor.config.json` ? defaults.
- Required env vars (core):
  - Telegram: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`
  - Vibe-Kanban: `VK_BASE_URL`, `VK_TARGET_BRANCH`, project/repo IDs
  - Executor routing: `EXECUTORS`, `EXECUTOR_DISTRIBUTION`
- VK workspace environment & PATH propagation:
  - Setup scripts inject PATH entries and GH tokens into VK executor profiles: `scripts/codex-monitor/setup.mjs:487`

## Testing
- Test runner: Vitest with Node environment and `tests/**/*.test.mjs` pattern (`scripts/codex-monitor/vitest.config.mjs:1`).
- Suggested commands:
  - `cd scripts/codex-monitor && npm run test`

## Troubleshooting
- **NO_CHANGES infinite loop**
  - Cause: orchestrator previously ignored the NO_CHANGES response.
  - Fix: `Get-AttemptFailureCategory` now archives immediately (`scripts/codex-monitor/ve-orchestrator.ps1:3563`).
- **Already-merged task infinite retry loop**
  - Cause: merged tasks retried without branch checks.
  - Fix: `Test-BranchMergedIntoBase` + `Test-TaskDescriptionAlreadyComplete` (`scripts/codex-monitor/ve-orchestrator.ps1:571`).
- **Zombie workspace cleanup**
  - Cause: temp worktrees left behind after completion.
  - Fix: prune worktrees on completion (`scripts/codex-monitor/ve-orchestrator.ps1:3223`).
- **Stale worktree path corruption**
  - Cause: `.git/worktrees` metadata not pruned.
  - Fix: prune worktrees during setup (`scripts/codex-monitor/setup.mjs:530`).
- **Credential helper corruption**
  - Cause: containerized `gh auth setup-git` writes container-only helpers.
  - Fix: remove local helper overrides on startup (`scripts/codex-monitor/ve-orchestrator.ps1:461`).
- **Rebase spam on completed tasks**
  - Cause: downstream rebases attempted on archived/completed tasks.
  - Fix: skip archived attempts (`scripts/codex-monitor/monitor.mjs:3495`).
