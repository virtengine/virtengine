# Codex-Monitor — AGENTS Guide

This guide is a fast, code-referenced map of the `scripts/codex-monitor/` module for AI agents. Every section points to the exact source lines to keep changes safe and grounded.

---

## Module Overview

**Purpose:** codex-monitor supervises VirtEngine's autonomous coding fleet — it schedules task attempts, runs PR automation, self-heals failures, and reports status via Telegram. The core supervisor (`monitor.mjs`) wires configuration, executor selection, fleet coordination, Telegram notifications, autofix, and maintenance sweeps into a single process loop.
**Primary entrypoints:** `cli.mjs` → `monitor.mjs` (supervisor), `ve-orchestrator.ps1` (task runner), `ve-kanban.ps1` (VK API wrapper).

**Code references:**

- Supervisor imports and orchestration wiring: `monitor.mjs:L14-L101`
- Task planner and preflight gating in monitor: `monitor.mjs:L400-L509`
- Orchestrator state + loop metrics: `ve-orchestrator.ps1:L144-L189`
- VK API wrapper and task attempt submission: `ve-kanban.ps1:L1-L45`, `ve-kanban.ps1:L310-L345`

---

### Start codex-monitor with defaults

```bash
node scripts/codex-monitor/cli.mjs
```

---

## 2. Architecture & Components

| Component                 | Role                                                                     | Key references                                                                                       |
| ------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| **monitor.mjs**           | Supervisor loop, smart PR flow, maintenance scheduling, fleet sync       | `monitor.mjs:L4047-L4440`, `monitor.mjs:L6960-L7050`                                                 |
| **telegram-bot.mjs**      | Telegram polling, batching/live digest, command queueing                 | `telegram-bot.mjs:L1-L120`, `telegram-bot.mjs:L185-L206`                                             |
| **ve-orchestrator.ps1**   | PowerShell task orchestration loop (parallel slots, retries, merge gate) | `ve-orchestrator.ps1:L4841-L4924`, `ve-orchestrator.ps1:L3306-L3504`                                 |
| **ve-kanban.ps1**         | VK CLI wrapper (list/submit/rebase/archive attempts)                     | `ve-kanban.ps1:L1-L45`, `ve-kanban.ps1:L310-L372`                                                    |
| **fleet-coordinator.mjs** | Multi-workstation coordination, fleet state persistence                  | `fleet-coordinator.mjs:L1-L21`, `fleet-coordinator.mjs:L181-L210`, `fleet-coordinator.mjs:L745-L769` |
| **autofix.mjs**           | Error loop detection + guarded auto-fix execution                        | `autofix.mjs:L1-L28`, `autofix.mjs:L66-L159`                                                         |
| **codex-shell.mjs**       | Persistent Codex SDK agent sessions                                      | `codex-shell.mjs:L1-L36`, `codex-shell.mjs:L104-L199`                                                |
| **copilot-shell.mjs**     | Persistent Copilot SDK agent sessions                                    | `copilot-shell.mjs:L1-L33`, `copilot-shell.mjs:L105-L166`                                            |
| **config.mjs**            | Unified config loader (CLI/env/.env/json/defaults)                       | `config.mjs:L4-L14`, `config.mjs:L81-L101`                                                           |

---

## 3. Critical Workflows

### Task lifecycle (creation → completion)

1. **Submit attempts:** `ve-kanban.ps1` posts `/api/task-attempts`, creating worktrees and starting agents (`Submit-VKTaskAttempt`).
   - `ve-kanban.ps1:L310-L345`
2. **Slot scheduling & merge gate:** `Fill-ParallelSlots` enforces capacity, avoids conflicts, and starts tasks with executor cycling.
   - `ve-orchestrator.ps1:L4841-L4924`
3. **Tracking + completion:** tracked attempts get PR checks/merge decisions; merged PRs trigger task completion updates.
   - `ve-orchestrator.ps1:L3306-L3504`

### PR flow (smartPRFlow + rebase logic)

- `smartPRFlow` handles branch status, stale detection, rebases, conflict resolution, PR creation, and fast/slow failure handling.
  - `monitor.mjs:L4047-L4440`
- Downstream rebases happen only for **active** tasks to avoid spam.
  - `monitor.mjs:L3453-L3554`

### Workspace creation, initialization, cleanup

- Attempt submission creates worktrees in VK and starts executors.
  - `ve-kanban.ps1:L310-L345`
- Completed/cancelled workspace cleanup: remove temp worktrees and prune git metadata.
  - `ve-orchestrator.ps1:L3223-L3304`
- Orphaned worktree cleanup with safety checks.
  - `workspace-reaper.mjs:L11-L178`
- Monitor-level maintenance sweep runs at startup and every 5 minutes.
  - `monitor.mjs:L6960-L6990`

### Notification batching and delivery

- Notification batching + immediate priority rules.
  - `telegram-bot.mjs:L95-L107`
- Live digest window configuration and debounce.
  - `telegram-bot.mjs:L108-L119`
- Serialized command queues for bot actions.
  - `telegram-bot.mjs:L185-L206`

### Error handling & recovery patterns

- Circuit breaker for repeated failures and monitor restart throttling.
  - `monitor.mjs:L400-L427`
- Autofix guardrails: max attempts + cooldown + dev/npm mode behavior.
  - `autofix.mjs:L1-L20`

---

## 4. Known Gotchas & Bug Patterns

Each gotcha includes root cause + fix location.

1. **NO_CHANGES infinite loop** ✅ FIXED
   - **Root cause:** orchestrator asked agents to reply "NO_CHANGES" if no work needed, but never actually parsed the response.
   - **Original issue:** fresh tasks with no commits were repeatedly prompted; agent responded NO_CHANGES but orchestrator ignored it and kept looping.
   - **Fix:** `Get-AttemptFailureCategory` now detects NO_CHANGES in agent output and immediately archives the attempt. Task descriptions with "superseded by", "already completed", etc. are detected during task selection and skipped entirely.
   - References: `ve-orchestrator.ps1:L2449-L2452` (detection), `ve-orchestrator.ps1:L3563-L3580` (archive on NO_CHANGES), `ve-orchestrator.ps1:L5206-L5229` (task description check)

2. **Already-merged task infinite retry loop** ✅ FIXED
   - **Root cause:** when a task was completed and merged (or committed directly to main) and the remote branch was deleted, the orchestrator entered the "no remote branch" retry path instead of detecting the task was already done. Four compounding gaps caused infinite loops:
     1. No "already merged into base" check before entering retry path
     2. `$script:TaskFollowUpCounts` (global safety cap) was declared but never incremented — dead code
     3. Task description "superseded by" check only ran at task selection, not for existing tracked attempts
     4. `monitor.mjs` also reacted to orchestrator's "No remote branch" log, causing dual-trigger
   - **Fix (orchestrator):**
     - `Test-BranchMergedIntoBase` function checks GitHub for merged PRs + `git merge-base --is-ancestor` fallback
     - `Test-TaskDescriptionAlreadyComplete` checks task description for "superseded by", "already completed", etc.
     - Both checks run in `Process-CompletedAttempts` before the retry logic
     - The "branch doesn't exist locally" fallback also checks for already-merged before marking `manual_review`
     - References: `ve-orchestrator.ps1` — `Test-BranchMergedIntoBase`, `Test-TaskDescriptionAlreadyComplete`, `Process-CompletedAttempts`
   - **Fix (follow-up cap):**
     - `$script:TaskFollowUpCounts` now wired up in both `Try-SendFollowUp` and `Try-SendFollowUpNewSession`
     - After `MAX_FOLLOWUPS_PER_TASK` (6) follow-ups, the attempt is archived and marked `manual_review`
     - References: `ve-orchestrator.ps1` — `Try-SendFollowUp`, `Try-SendFollowUpNewSession`
   - **Fix (monitor.mjs):**
     - `resolveAndTriggerSmartPR` checks `mergedBranchCache` + `isBranchMerged()` before triggering `smartPRFlow`
     - `smartPRFlow` Step 0 checks for merged branch and task description completion signals before any work
     - References: `monitor.mjs` — `resolveAndTriggerSmartPR`, `smartPRFlow`

3. **Zombie workspace cleanup**
   - **Root cause:** completed/cancelled tasks left temp worktrees behind, causing “ghost” workspaces and stale git metadata.
   - **Fix:** prune git worktrees and remove VK worktree directories on completion.
   - References: `ve-orchestrator.ps1:L3223-L3304`

4. **Stale worktree path corruption**
   - **Root cause:** worktree directories deleted without pruning `.git/worktrees` metadata, causing path resolution errors.
   - **Fix:** setup scripts prune worktrees and note that VK worktree paths live under `.git/worktrees/` or `vibe-kanban`.
   - References: `setup.mjs:L530-L539`, `ve-orchestrator.ps1:L2099-L2104`

5. **Credential helper corruption**
   - **Root cause:** VK containers run `gh auth setup-git`, writing a container-only helper path to `.git/config`.
   - **Fix:** remove local helper overrides on startup; rely on global helper or GH_TOKEN.
   - References: `ve-orchestrator.ps1:L461-L475`, `setup.mjs:L518-L527`

6. **Fresh vs crashed task detection**
   - **Root cause:** no distinction between “fresh task (0 commits)” and “crashed task (commits exist but not pushed)”.
   - **Fix:** check commit counts, restart fresh tasks, or prompt push for crashed ones.
   - References: `ve-orchestrator.ps1:L571-L605`, `ve-orchestrator.ps1:L3383-L3424`

7. **Rebase spam on completed tasks**
   - **Root cause:** downstream rebases attempted on archived/completed tasks.
   - **Fix:** filter VK attempts to active tasks only; skip archived attempts before rebase.
   - References: `monitor.mjs:L3495-L3506`, `monitor.mjs:L3543-L3548`

---

## 5. Configuration & Environment

**Config loading order:** CLI → env vars → `.env` → `codex-monitor.config.json` → defaults.
References: `config.mjs:L4-L14`, `config.mjs:L81-L101`

**Required env vars (core):**

- Telegram: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`
  - `.env.example:L11-L23`
- Vibe-Kanban: `VK_BASE_URL`, `VK_TARGET_BRANCH`, project/repo IDs
  - `.env.example:L87-L116`, `.env.example:L142-L146`
- Executor routing: `EXECUTORS`, `EXECUTOR_DISTRIBUTION`, failover config
  - `.env.example:L66-L85`
- Primary agent: `PRIMARY_AGENT`, `COPILOT_SDK_DISABLED`, `CLAUDE_SDK_DISABLED`
  - `.env.example:L157-L193`

**VK workspace environment & PATH propagation:**

- Setup scripts inject PATH entries and GH tokens into VK executor profiles.
  - `setup.mjs:L487-L529`, `setup.mjs:L1284-L1300`

**Git auth patterns:**

- Copilot shell pulls token from env or `gh auth status`.
  - `copilot-shell.mjs:L71-L102`
- Preflight enforces `git`, `gh`, `node`, `pwsh` on PATH.
  - `preflight.mjs:L143-L183`

---

## 6. State Management

**Orchestrator state & metrics:**

- Persistent state + status files in `.cache/` (`ve-orchestrator-state.json`, `ve-orchestrator-status.json`, success metrics).
  - `ve-orchestrator.ps1:L144-L180`

**Planner + fleet state:**

- Task planner state written under monitor logs.
  - `monitor.mjs:L429-L434`
- Fleet state persisted in `.cache/codex-monitor/fleet-state.json`.
  - `fleet-coordinator.mjs:L745-L769`

**Completed task archive:**

- Completed task archive stored under `.cache/completed-tasks/` with grouped JSON files.
  - `task-archiver.mjs:L1-L48`

---

## 7. Testing

**Test runner:** Vitest with Node environment and `tests/**/*.test.mjs` pattern.

- `vitest.config.mjs:L1-L9`

**Scripts:**

- `npm run test` → `vitest run`, with `pretest` syntax checks.
  - `package.json:L62-L70`

**Test coverage examples:** `tests/` includes `autofix.test.mjs`, `fleet-coordinator.test.mjs`, `workspace-reaper.test.mjs`, `vk-api.test.mjs`, etc.  
Use `npm run test` from `scripts/codex-monitor/`.

---

## 8. Implementation Patterns

### Adding a new executor

1. Update executor config parsing in `config.mjs` (executor list + failover).
   - `config.mjs:L203-L255`
2. Ensure VK executor profiles are wired in setup (`setup.mjs` auto-config).
   - `setup.mjs:L1284-L1300`
3. Add executor cycling or routing logic in `ve-kanban.ps1` or `task-complexity` if needed.
   - `ve-kanban.ps1:L41-L46`, `ve-kanban.ps1:L323-L337`

### Extending notification logic

1. Add message formats and priorities in `telegram-bot.mjs`.
   - `telegram-bot.mjs:L95-L119`
2. Ensure monitor uses `notify` or `sendTelegramMessage` on key events.
   - `monitor.mjs:L4047-L4440`

### Modifying PR flow behavior

- Edit `smartPRFlow` (rebase/PR creation/fast-fail logic) and rebase filters.
  - `monitor.mjs:L4047-L4440`, `monitor.mjs:L3453-L3554`

### Adding new autofix patterns

- Extend error parsing and signatures in `autofix.mjs`, then add tests.
  - `autofix.mjs:L66-L159`

---

## Quick Orientation Checklist (10-minute ramp-up)

- Read `monitor.mjs` smart PR + maintenance paths (`monitor.mjs:L4047-L4440`, `monitor.mjs:L6960-L6990`).
- Review orchestrator slot fill + PR completion (`ve-orchestrator.ps1:L4841-L4924`, `ve-orchestrator.ps1:L3306-L3504`).
- Check Telegram batching and bot queues (`telegram-bot.mjs:L95-L206`).
- Run `npm run test` in `scripts/codex-monitor/` to validate changes (`package.json:L62-L70`).
