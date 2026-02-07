# Slopdev Workshop Phase 2: Multi-Workspace Orchestration

## Goal

Enable multiple autonomous workspaces to coordinate through a shared Telegram bot and vibe-kanban, with a primary coordinator handling routing and planning.

## Bootstrap (Windows-first)

1. Ensure each workspace has its own repo checkout + monitor running:
   - `pnpm -C scripts/codex-monitor install`
   - `pnpm -C scripts/codex-monitor start -- --args "-MaxParallel 6"`
2. Set the shared Telegram bot env vars:
   - `TELEGRAM_BOT_TOKEN`
   - `TELEGRAM_CHAT_ID`
3. Set the local workspace identity:
   - `VE_WORKSPACE_ID=primary` (or `eng-1`, `eng-2`, etc.)
4. Optionally override registry path:
   - `VE_WORKSPACE_REGISTRY_PATH=scripts/codex-monitor/workspace-registry.json`

Linux/WSL can reuse the same registry with different `host` values. Use the same bot token and chat id to keep a unified message bus.

## Workspace Registry

Registry lives at `scripts/codex-monitor/workspace-registry.json` and defines:

- `id`: stable workspace id used for routing
- `role`: `coordinator` or `agent`
- `capabilities`: domain hints (go, portal, infra, etc.)
- `model_priorities`: used by task planner to route work
- `vk_endpoint_url`: vibe-kanban endpoint for that workspace
- `telegram_prefix`: optional short tag for filtering

### Adding a Workspace

1. Add a new entry to `scripts/codex-monitor/workspace-registry.json`.
2. Set `VE_WORKSPACE_ID` on the new machine to match the `id`.
3. Run the monitor. No code changes required.

## Telegram Commands

- `/agent --workspace <id> <message>`  
  Routes work to a specific workspace. If `--workspace` is omitted, the primary coordinator handles it.

- `/handoff --to <id> <message>`  
  Sends a queued handoff to a target workspace.

- `/status`  
  Primary coordinator replies with the current status summary.

Routing helpers:
- Mention routing: include `@<workspace-id>` or `#<workspace-id>` in the message.
- Prefix filters: start a message with `[HQ]` or `ENG1:` to target a workspace prefix.

## Heartbeats + Planner Automation

- Each workspace sends a Telegram heartbeat every 5 minutes (configurable via `TELEGRAM_HEARTBEAT_MIN`).
- Planner triggers when backlog per agent is low, idle agents are detected, or stalled tasks accumulate.

Planner thresholds:
- `TASK_PLANNER_BACKLOG_PER_AGENT` (default: 10)
- `TASK_PLANNER_COOLDOWN_MIN` (default: 30)
- `TASK_PLANNER_ERROR_THRESHOLD` (default: 1)

## Shared Status & Logs

- Task planner outputs are stored in `scripts/codex-monitor/logs/task-planner-*.md`.
- Workspace message digests persist to `.cache/telegram-digest.json`.
- Progress tracking remains in `_docs/ralph/progress.md` (update per planning run).

## Model Priority Config

Set `model_priorities` in the registry entry. Example:

```
["CODEX_DEFAULT", "COPILOT_CLAUDE_OPUS_4_6"]
```

The planner uses these hints to route analysis-heavy work to stronger models.
