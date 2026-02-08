# Slopes Slopdev Multi-Workspace Operator Guide

This guide documents how to bootstrap, add workspaces, configure Telegram notifications, tune model priority, and recover from disconnects/logouts in the Slopes slopdev multi-workspace environment.

## Scope and terminology

- **Workspace**: A vibe-kanban task attempt that creates a Git worktree and runs an agent session.
- **Orchestrator**: `scripts/codex-monitor/ve-orchestrator.ps1`, the loop that assigns tasks and monitors progress.
- **Monitor**: `scripts/codex-monitor/monitor.mjs`, long-running wrapper with restart + diagnostics.

## Bootstrap (Windows)

1. **Prereqs**
   - PowerShell 7+ (`pwsh`), Git, Node.js 18+, pnpm.
   - Vibe-kanban API reachable (default `http://127.0.0.1:54089`).

2. **Install monitor dependencies**

```bash
pnpm -C scripts/codex-monitor install
```

3. **Configure environment variables**

```bash
setx VK_BASE_URL "http://127.0.0.1:54089"
setx VK_PROJECT_NAME "virtengine"
setx VK_TARGET_BRANCH "origin/main"
setx GH_OWNER "virtengine"
setx GH_REPO "virtengine"
```

Optional (if your VK service advertises a public UI URL):

```bash
setx VK_PUBLIC_URL "http://<vk-host>:<port>"
```

4. **Start the monitor (recommended)**

```bash
pnpm -C scripts/codex-monitor start -- --args "-MaxParallel 6"
```

This launches `scripts/codex-monitor/ve-orchestrator.ps1` and auto-restarts on failures.

## Bootstrap (Linux/WSL - future)

1. Install `pwsh`, Git, Node.js, and pnpm in your distro.
2. Ensure the VK API is reachable from WSL (`VK_BASE_URL`).
3. Install monitor deps:

```bash
pnpm -C scripts/codex-monitor install
```

4. Run:

```bash
pnpm -C scripts/codex-monitor start -- --args "-MaxParallel 6"
```

## Adding a workspace

Use the vibe-kanban CLI wrapper to create new attempts (workspaces). Each attempt spins up a new Git worktree with a unique branch.

### Add a specific workspace (task ID)

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 submit <task-id>
```

### Add the next N tasks as workspaces

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 submit-next --count 2
```

### See active workspaces

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 status
```

## Telegram bot configuration

The codex monitor posts important status updates to Telegram.

### Required environment variables

- `TELEGRAM_BOT_TOKEN` - BotFather token
- `TELEGRAM_CHAT_ID` - numeric chat ID (user or group)
- `TELEGRAM_INTERVAL_MIN` - minimum minutes between heartbeat posts (default: 10)

### Get a chat ID

1. Send a message to your bot (or add it to a group and send a message).
2. Run:

```bash
set TELEGRAM_BOT_TOKEN=<token>
node scripts/codex-monitor/get-telegram-chat-id.mjs
```

3. Set the chosen ID:

```bash
setx TELEGRAM_CHAT_ID "<chat-id>"
```

### Verify Telegram notifications

Start the monitor and confirm the first status post arrives in the chat.

## Model priority configuration examples

The orchestrator alternates between executor profiles defined in `scripts/codex-monitor/ve-kanban.ps1` (used by `ve-orchestrator.ps1`). Update the list to control priority or weighting.

### Default (50/50 Codex/Copilot)

```powershell
$script:VK_EXECUTORS = @(
    @{ executor = "CODEX"; variant = "DEFAULT" }
    @{ executor = "COPILOT"; variant = "CLAUDE_OPUS_4_6" }
)
```

### Prioritize Codex (2:1)

```powershell
$script:VK_EXECUTORS = @(
    @{ executor = "CODEX"; variant = "DEFAULT" }
    @{ executor = "CODEX"; variant = "DEFAULT" }
    @{ executor = "COPILOT"; variant = "CLAUDE_OPUS_4_6" }
)
```

### Codex-only fallback

```powershell
$script:VK_EXECUTORS = @(
    @{ executor = "CODEX"; variant = "DEFAULT" }
)
```

After edits, restart the monitor so the orchestrator reloads the updated profile list.

## Troubleshooting disconnects and logouts

### VK API unreachable

Symptoms:
- Orchestrator logs show repeated VK retries
- Status shows no new workspaces created

Steps:
1. Verify the VK service is up (`VK_BASE_URL`).
2. Ensure port `54089` (default) is reachable from the host running the orchestrator.
3. Restart the monitor after VK is healthy.

### Workspace archived or missing

Symptoms:
- Task is BEHIND but rebase fails
- Workspace disappears from `status`

Steps:
1. List archived attempts:

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 archived
```

2. Unarchive if needed:

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 unarchive <attempt-id>
```

3. Rebase the attempt:

```bash
pwsh scripts/codex-monitor/ve-kanban.ps1 rebase <attempt-id>
```

### Agent session disconnected

Symptoms:
- No new logs from workspace
- Monitor reports crashed child process

Steps:
1. Restart the monitor (it auto-relaunches the orchestrator).
2. Verify the workspace still exists via `pwsh scripts/codex-monitor/ve-kanban.ps1 status`.
3. If the attempt is stale, archive it in VK (or via the CLI helper) and resubmit the task:

```bash
pwsh -Command ". .\\scripts\\ve-kanban.ps1; Archive-VKAttempt -AttemptId <attempt-id>"
pwsh scripts/codex-monitor/ve-kanban.ps1 submit <task-id>
```

### GitHub auth expired

Symptoms:
- `gh` commands fail in logs
- PR creation fails repeatedly

Steps:
1. Re-authenticate:

```bash
gh auth login
```

2. Restart the monitor.

## Recovery checklist

- Confirm VK API is reachable.
- Confirm monitor is running and posting status.
- Ensure workspaces are active (`status`).
- Unarchive or rebase stuck attempts.
- Resubmit tasks if the worktree was lost.

## Related references

- `scripts/codex-monitor/ve-orchestrator.ps1`
- `scripts/codex-monitor/ve-kanban.ps1`
- `scripts/codex-monitor/README.md`
- `_docs/failure-analysis.md`
