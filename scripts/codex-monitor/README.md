# Codex Monitor

Long-running wrapper that restarts the orchestrator and uses the Codex SDK to analyze failures.

## Setup

```bash
pnpm -C scripts/codex-monitor install
```

If dependencies are missing, the monitor will attempt to run `pnpm install` (or `npm install`) automatically on first run.

## Run

```bash
pnpm -C scripts/codex-monitor start -- --args "-MaxParallel 6"
```

## Multi-Workspace Routing

Workspaces are configured in `scripts/codex-monitor/workspaces.json`.
The Telegram bot supports:

- `/workspaces` to list workspaces
- `/agent [--workspace <id>] <task>` to route tasks (defaults to the coordinator)
- `/handoff @workspace <message>` to share context
- `@workspace` mentions to hand off tasks

See `docs/operations/slopdev-workshop.md` for setup details.

Environment highlights:

- `VE_WORKSPACE_ID` selects the local workspace entry.
- `VE_WORKSPACE_REGISTRY` overrides the registry path.
- `TELEGRAM_INTERVAL_MIN` controls heartbeat cadence (monitor).
- `TELEGRAM_DIGEST_MAX` caps entries per workspace digest file.

### Options

- `--script <path>`: Path to the PowerShell script. Default: `scripts/ve-orchestrator.ps1`.
- `--args "<args>"`: Arguments passed to the script. Default: `-MaxParallel 6`.
- `--restart-delay <ms>`: Delay before restart. Default: 10000.
- `--max-restarts <n>`: Max restarts (0 = unlimited).
- `--log-dir <path>`: Log directory. Default: `scripts/codex-monitor/logs`.
- `--watch-path <path>`: File or directory to watch for restarts. Default: `scripts/ve-orchestrator.ps1`.
- `--no-codex`: Disable Codex SDK analysis.
- `--no-watch`: Disable file watching.
- `--no-preflight`: Disable preflight checks (git config, gh auth, disk space, clean worktree, toolchain versions).
- `--preflight-retry <ms>`: Retry interval after preflight failure (default: 300000).

### Telegram heartbeat

When the Telegram bot is enabled, it posts a workspace heartbeat every 5 minutes by default.

Environment overrides:

- `TELEGRAM_HEARTBEAT_INTERVAL_MIN` — heartbeat interval in minutes (default: 5).
- `TELEGRAM_HEARTBEAT_DISABLED` — set to `1` or `true` to disable heartbeat.
- `TELEGRAM_HEARTBEAT_QUIET_HOURS` — quiet window(s) like `22:00-07:00` or `8-9,18:00-20:00`.
- `TELEGRAM_HEARTBEAT_TIMEZONE` — optional IANA timezone (e.g., `America/Los_Angeles`).

## Notes

- The Codex SDK uses environment configuration from your Codex setup (API key, base URL).
- If Codex analysis fails, the error is written next to the log file.
- On startup, codex-monitor runs preflight checks and blocks the orchestrator until the issues are resolved.
- Task planner automation env vars:
  - `TASK_PLANNER_PER_CAPITA_THRESHOLD` (default: 1) triggers when backlog-per-slot falls below threshold.
  - `TASK_PLANNER_IDLE_SLOT_THRESHOLD` (default: 1) triggers when idle slots meet/exceed threshold.
  - `TASK_PLANNER_DEDUP_HOURS` (default: 24) prevents repeated planner runs within the window.
