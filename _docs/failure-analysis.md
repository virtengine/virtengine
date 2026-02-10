# VirtEngine Orchestrator — Failure Analysis Report

> Generated from 191 orchestrator logs, 258 agent PRs, and orchestrator source analysis.

---

## Executive Summary

| Metric                              | Value              |
| ----------------------------------- | ------------------ |
| Total agent PRs created             | 258                |
| PRs merged (success)                | 244                |
| PRs rejected (closed without merge) | 13                 |
| PRs open (pending)                  | 1                  |
| **PR merge rate**                   | **94.9%**          |
| Orchestrator restarts (logs)        | 191                |
| Rapid restarts (<5 min gap)         | ~40+ (crash loops) |

---

## Category 1: CI/CD Failures

**Frequency: 27 occurrences across 191 logs (14.1%)**

### 1A. CI checks failing on agent PRs

The most common CI failures are:

- **Portal lint/type-check failures** — Agent changes TypeScript code that fails ESLint or `tsc --noEmit`
- **Go build failures** — Missing imports, type mismatches introduced by agents
- **Go test failures** — Agents break existing tests or write failing new tests
- **Pre-push hook failures** — `golangci-lint`, `go vet`, or Prettier violations

### 1B. Copilot sub-PR cascade pattern

12 of the 13 rejected PRs follow the `copilot/sub-pr-*` naming pattern. This occurs when:

1. Agent creates a PR that fails CI
2. Orchestrator requests `@copilot` fix via PR comment
3. Copilot creates a `copilot/sub-pr-{N}` branch to fix CI
4. The sub-PR itself often fails or the original PR already got superseded
5. Multiple sub-PRs are created for the same issue (`sub-pr-312`, `sub-pr-312-again`, `sub-pr-312-another-one`)

**Root cause**: Copilot fix PRs target the agent's branch, but if main has moved significantly, the fixes become stale. The orchestrator creates endless sub-PR variants.

### 1C. Pre-push hook blocking agent commits

Pre-push hooks run `go vet`, `golangci-lint`, `go mod vendor`, `make bins`, and Go unit tests. Agents unfamiliar with these gates get blocked silently and produce empty branches.

**Evidence**: 6 occurrences of "Agent didn't push branch" = agent work was done locally but pre-push rejected it.

---

## Category 2: Agent-Related Mixups

**Frequency: 9 task failures + 6 missing branches + 3 manual reviews = 18 incidents**

### 2A. Empty branches / no commits

1 explicit "no commits" case and 6 "missing remote branch" cases show agents that:

- Completed their work locally but failed to `git push`
- Had their push rejected by pre-push hooks
- Ran out of context window before pushing

### 2B. Context window exhaustion

The orchestrator has `Test-ContextWindowError` detection and `Try-RecoverContextWindow` logic. When Copilot agents hit context limits:

- Work is incomplete
- No push happens
- Recovery is attempted but requires a 10-minute cooldown

### 2C. Agent working on wrong files / scope creep

Some agents modify files outside their task scope, creating merge conflicts with other parallel agents. 6 parallel task slots means 6 agents can step on each other's toes.

### 2D. Stale/stuck operations

1 occurrence of stale/stuck operation detected. The maintenance module now handles this with:

- 15-minute git push reaper
- Stale orchestrator process killer
- Worktree cleanup

---

## Category 3: Missing Crucial Instructions

### 3A. Pre-push hooks not understood

Agents don't always know about the pre-push quality gates. The `AGENTS.md` documents them, but agents may not read all instructions before starting work. This leads to:

- Failed `git push` with no error visible to the orchestrator
- Agent reports "completed" but no branch exists remotely

### 3B. Dependency installation

Portal changes require `pnpm -C portal install` before lint/test can pass. Agents sometimes skip this step, causing Prettier and ESLint failures.

### 3C. Go mod vendor sync

After modifying `go.mod`, agents must run `go mod vendor`. The pre-push hook enforces this, but agents don't always do it proactively.

### 3D. Conventional commit format

While documented, agents occasionally create non-conforming commit messages that get rejected by `commitlint`.

---

## Category 4: Conflicts from Slow Merge Strategy

**Frequency: 9 merge conflict states + 9 PR merge failures = 18 incidents**

### 4A. Parallel agent conflict

With 6 parallel slots, two agents can modify the same file simultaneously—especially:

- `go.mod` / `go.sum` (dependency changes)
- `portal/package.json` / `pnpm-lock.yaml`
- Shared types in `x/*/types/`
- `app/` module wiring

When Agent A's PR merges first, Agent B's PR becomes DIRTY/CONFLICTING.

### 4B. Merge state BEHIND

The orchestrator detects BEHIND state and requests rebase, but this requires the VK workspace to be active. If the workspace has already been archived, rebase fails silently.

### 4C. Serial merge bottleneck

The orchestrator merges PRs one at a time. When 6 agents complete simultaneously:

1. PR #1 merges → main advances
2. PR #2 is now BEHIND → needs rebase → add to queue
3. PR #3 BEHIND → rebase queue grows
4. By the time PR #6 gets to merge, it's several commits behind and may have conflicts

---

## Category 5: Infrastructure / Orchestrator Crashes

**Frequency: 36 VK failures + 22 timeouts + 19 ParserErrors + 9 PS crashes + 4 mutex = 90 incidents**

### 5A. Vibe-kanban API unreachable (36 occurrences)

The VK API at `http://192.168.0.161:54089` is frequently unreachable. The orchestrator waits 90s and retries, but if VK is down for extended periods, the orchestrator accumulates restarts.

### 5B. PowerShell ParserError (19 occurrences)

Two specific parser bugs caused crash loops:

- `$Branch:` without `${}` escaping (line 437) — fixed
- String interpolation issues in log messages

### 5C. PS type/method crashes (9 occurrences)

The `op_Addition` crash in `Get-OrderedTodoTasks` caused a sustained crash loop (3 rapid restarts at 3:51-3:52 PM). Fixed by switching to `List[object]`.

### 5D. Mutex contention (4 occurrences)

Multiple orchestrator instances attempting to run simultaneously. The mutex guard now properly exits surplus instances.

### 5E. Timeouts (22 occurrences)

Various timeout sources: GitHub API, VK API, `gh` CLI commands, git operations. The orchestrator has rate limit detection but timeouts still cause cycle delays.

---

## Failure Prioritization (Impact × Frequency)

| #   | Failure Class          | Impact                | Frequency   | Priority                  |
| --- | ---------------------- | --------------------- | ----------- | ------------------------- |
| 1   | VK API unreachable     | High (blocks all)     | 36          | **Critical**              |
| 2   | CI check failures      | Medium (blocks merge) | 27          | **High**                  |
| 3   | Timeouts               | Medium (delays)       | 22          | **High**                  |
| 4   | PS Parser/Type crashes | High (crash loop)     | 28          | **Medium** (mostly fixed) |
| 5   | Merge conflicts        | Medium (blocks merge) | 18          | **Medium**                |
| 6   | Agent missing branches | High (task wasted)    | 6           | **Medium**                |
| 7   | Copilot sub-PR cascade | Low (wasted effort)   | 12 rejected | **Low**                   |
| 8   | Mutex contention       | Low (self-resolves)   | 4           | **Low**                   |

---

## Recommended Reliability Improvements

### Immediate (implement now)

1. **Success rate tracking** — Add first-shot success metric to status.json and Telegram reports
2. **Limit copilot sub-PR retries** — Cap at 1 sub-PR attempt per original PR, close stale ones
3. **VK health check with exponential backoff** — Instead of fixed 90s retry, use 30s→60s→120s→300s

### Short-term (next sprint)

4. **Pre-flight validation before task assignment** — Check that the agent's target files don't overlap with other active agent branches
5. **Agent instructions reinforcement** — Add a mandatory pre-push checklist to the task prompt itself
6. **Merge queue ordering** — Merge oldest PRs first to minimize rebase churn

### Medium-term

7. **Conflict-aware task scheduling** — Don't assign tasks that touch the same module to parallel agents
8. **CI pre-check** — Run a lightweight lint/build check before pushing (agent-side)
9. **Auto-rebase on BEHIND** — Instead of requesting VK rebase, run `git rebase origin/main` directly

---

## Success Rate Tracking (Implementation)

The orchestrator now tracks:

- `tasks_first_shot_success` — Tasks merged on first attempt (no copilot fix, no manual intervention)
- `tasks_needed_fix` — Tasks that required copilot sub-PR or manual fix
- `tasks_failed` — Tasks that were abandoned or rejected
- `first_shot_rate` — Percentage: `first_shot_success / (first_shot_success + needed_fix + failed)`

This is exposed in:

1. `ve-orchestrator-status.json` → `success_metrics` object
2. Telegram periodic report (every 5 completed tasks)
3. Console cycle banner
