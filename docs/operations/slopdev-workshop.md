# Slopdev Workshop — Multi-Workspace Orchestration

This guide documents the Phase 2 Slopdev workshop setup for coordinating multiple workspaces via Telegram + Vibe-Kanban.

## Overview

- **Telegram bot** acts as the central communication bus.
- **Each workspace** runs its own local repo + monitor process.
- **Task routing** happens via `/agent` commands or `@workspace` mentions.
- **Planner automation** triggers when backlog metrics fall below thresholds.

## Workspace Registry

Workspaces are defined in `scripts/codex-monitor/workspaces.json`.

Example:

```json
{
  "version": 1,
  "default_workspace": "primary",
  "workspaces": [
    {
      "id": "primary",
      "name": "Primary Coordinator",
      "role": "coordinator",
      "capabilities": ["planning", "triage"],
      "model_priorities": ["CODEX:DEFAULT", "COPILOT:CLAUDE_OPUS_4_6"],
      "vk_workspace_id": "primary",
      "mentions": ["primary", "coord", "coordinator"]
    }
  ]
}
```

### Add a new workspace

1. Add a new entry to `scripts/codex-monitor/workspaces.json`.
2. Ensure the `vk_workspace_id` matches the workspace ID in Vibe-Kanban.
3. Add `mentions` for `@workspace` routing (optional).
4. Restart `pnpm -C scripts/codex-monitor start` on the workspace.

No code changes are required for new workspaces.

## Telegram Commands

- `/workspaces` — list configured workspaces.
- `/agent --workspace <id> <task>` — route task to a workspace. Defaults to the `default_workspace`.
- `/handoff @workspace <message>` — send a handoff to one or more workspaces.
- `/digest [count]` — show recent telegram digest for the local workspace.

### Mention routing

Messages that include `@workspace` or `[ws:workspace]` are routed to those workspaces.

Examples:

- `@eng-01 investigate failing build`
- `[ws:primary] plan next backlog`
- `@all sync status updates`

## Planner Automation

The monitor triggers the task planner when:

- **Backlog per agent** falls below `PLANNER_BACKLOG_PER_AGENT` (default 10).
- **Idle workspaces** are detected while backlog remains.
- **Stalled attempts** exceed `PLANNER_STALLED_MIN` (default 45 minutes).

Configurable env vars:

- `PLANNER_BACKLOG_PER_AGENT` (default `10`)
- `PLANNER_IDLE_TRIGGER_MIN` (default `15`)
- `PLANNER_STALLED_MIN` (default `45`)
- `PLANNER_COOLDOWN_MIN` (default `60`)

## Heartbeat

The monitor sends a Telegram heartbeat every 5 minutes by default. Configure with:

- `TELEGRAM_INTERVAL_MIN` (default `5`)

Heartbeat messages include:

- Task counts (running/review/error)
- Workspace load (active/idle/busy)
- Backlog per agent

## Model Priority Routing

Workspace-specific model routing is controlled in `workspaces.json`:

```json
"model_priorities": ["CODEX:DEFAULT", "COPILOT:CLAUDE_OPUS_4_6"]
```

The bot picks the first available executor in the list when routing tasks.

## Required Environment Variables

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`
- `VK_ENDPOINT_URL` (vibe-kanban API)
- Optional: `VE_WORKSPACE_ID`, `VE_WORKSPACE_REGISTRY`
