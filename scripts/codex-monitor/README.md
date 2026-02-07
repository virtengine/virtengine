# Codex Monitor

> AI-powered orchestrator supervisor — wraps any task orchestrator script, auto-restarts on failure, analyzes crashes with the Codex SDK, manages PRs via the Vibe-Kanban API, and sends real-time Telegram notifications.

## What It Does

Codex Monitor is a long-running Node.js process that:

1. **Supervises** a task orchestrator script (PowerShell, Bash, or any CLI)
2. **Auto-restarts** the orchestrator on crash with configurable delays
3. **Analyzes failures** using the Codex SDK — diagnoses root causes from logs
4. **Auto-fixes** repeating errors by detecting error loops and invoking AI
5. **Creates PRs automatically** using the Vibe-Kanban API (rebase → resolve conflicts → push → create PR)
6. **Sends Telegram notifications** for errors, merges, milestones, and status summaries
7. **Provides a Telegram chatbot** for interactive commands (`/status`, `/tasks`, `/restart`, `/stop`, free-text AI chat)
8. **Manages Vibe-Kanban** lifecycle — spawns, monitors, and restarts the task board backend

## Benefits

- **Zero-touch operation**: Set up once, walk away. The monitor handles crashes, retries, and PR creation.
- **Smart PR flow**: Automatically rebases, resolves merge conflicts, runs prepush hooks, and creates PRs — so agents don't get stuck on "agent must push before PR".
- **AI-powered crash analysis**: When the orchestrator crashes, Codex reads the log and provides a diagnosis.
- **Loop detection**: If the same error repeats 4+ times in 10 minutes, Codex autofix kicks in automatically.
- **Stale attempt cleanup**: Detects dead attempts (0 commits, many behind main) and archives them.
- **Real-time visibility**: Telegram notifications keep you informed without watching a terminal.
- **Works with any orchestrator**: While built for VirtEngine's `ve-orchestrator.ps1`, it can wrap any long-running script.

## Quick Start

### 1. Install

```bash
# In your project
cd scripts/codex-monitor   # or wherever you place it
npm install
```

### 2. Configure

Run the interactive setup wizard:

```bash
node setup.mjs
```

Or copy the example environment file and edit it:

```bash
cp .env.example .env
# Edit .env with your values
```

### 3. Run

```bash
node monitor.mjs
```

With options:

```bash
node monitor.mjs --args "-MaxParallel 6" --restart-delay 10000
```

## Configuration

All configuration is done through environment variables (or a `.env` file).

### Required

| Variable             | Description                                                            |
| -------------------- | ---------------------------------------------------------------------- |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token from @BotFather                                     |
| `TELEGRAM_CHAT_ID`   | Your Telegram chat ID (use `node get-telegram-chat-id.mjs` to find it) |

### Recommended

| Variable         | Default                  | Description                                       |
| ---------------- | ------------------------ | ------------------------------------------------- |
| `OPENAI_API_KEY` | —                        | OpenAI (or compatible) API key for Codex analysis |
| `VK_BASE_URL`    | `http://127.0.0.1:54089` | Vibe-Kanban API endpoint                          |
| `GITHUB_REPO`    | auto-detected            | GitHub repo slug (`org/repo`)                     |
| `MAX_PARALLEL`   | `6`                      | Maximum concurrent agent slots                    |

### Optional

| Variable                            | Default                     | Description                               |
| ----------------------------------- | --------------------------- | ----------------------------------------- |
| `OPENAI_BASE_URL`                   | `https://api.openai.com/v1` | Custom API base (Azure, local, etc.)      |
| `CODEX_MODEL`                       | SDK default                 | Model to use for analysis                 |
| `CODEX_SDK_DISABLED`                | `0`                         | Set to `1` to disable all AI features     |
| `VK_RECOVERY_PORT`                  | `54089`                     | Vibe-Kanban API port                      |
| `VK_PUBLIC_URL`                     | —                           | Public URL for VK links in Telegram       |
| `VK_NO_SPAWN`                       | `0`                         | Set to `1` to prevent auto-spawning VK    |
| `TELEGRAM_INTERVAL_MIN`             | `10`                        | Minutes between periodic status summaries |
| `TASK_PLANNER_PER_CAPITA_THRESHOLD` | `1`                         | Trigger planner when backlog-per-slot < N |
| `TASK_PLANNER_IDLE_SLOT_THRESHOLD`  | `1`                         | Trigger planner when idle slots ≥ N       |
| `TASK_PLANNER_DEDUP_HOURS`          | `24`                        | Don't re-trigger planner within N hours   |
| `RESTART_DELAY_MS`                  | `10000`                     | Delay before restarting orchestrator      |
| `MAX_RESTARTS`                      | `0`                         | Max restarts (0 = unlimited)              |

See [.env.example](.env.example) for the full reference.

## CLI Options

```
node monitor.mjs [options]
```

| Option                      | Description                                                         |
| --------------------------- | ------------------------------------------------------------------- |
| `--script <path>`           | Path to the orchestrator script (default: `../ve-orchestrator.ps1`) |
| `--args "<args>"`           | Arguments passed to the script (default: `-MaxParallel 6`)          |
| `--restart-delay <ms>`      | Delay before restart (default: `10000`)                             |
| `--max-restarts <n>`        | Max restarts, 0 = unlimited (default: `0`)                          |
| `--log-dir <path>`          | Log directory (default: `./logs`)                                   |
| `--watch-path <path>`       | File to watch for auto-restart (default: script path)               |
| `--no-codex`                | Disable Codex SDK analysis                                          |
| `--no-watch`                | Disable file watching                                               |
| `--no-autofix`              | Disable automatic error fixing                                      |
| `--no-telegram-bot`         | Disable the interactive Telegram bot                                |
| `--no-vk-spawn`             | Don't auto-spawn Vibe-Kanban                                        |
| `--no-echo-logs`            | Don't echo orchestrator output to console                           |
| `--vk-ensure-interval <ms>` | VK health check interval (default: `60000`)                         |

## Setting Up Telegram

### 1. Create a Bot

1. Open Telegram and search for **@BotFather**
2. Send `/newbot` and follow the prompts
3. Copy the bot token (looks like `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11`)
4. Set `TELEGRAM_BOT_TOKEN` in your `.env`

### 2. Get Your Chat ID

```bash
node get-telegram-chat-id.mjs
```

This starts a temporary listener. Send any message to your bot — the script will print your chat ID. Set `TELEGRAM_CHAT_ID` in your `.env`.

### 3. Bot Commands

Once running, the Telegram bot responds to:

| Command               | Description                                     |
| --------------------- | ----------------------------------------------- |
| `/status`             | Current orchestrator status and attempt summary |
| `/tasks`              | List active tasks with progress                 |
| `/agents`             | Show agent slot utilization                     |
| `/health`             | System health check                             |
| `/restart`            | Restart the orchestrator                        |
| `/stop`               | Gracefully stop the orchestrator                |
| `/reattempt <id>`     | Re-queue a failed task                          |
| `/plan <description>` | Trigger the AI task planner                     |
| Free text             | Chat with Codex AI about the project            |

## Smart PR Flow

When an agent finishes a task, Codex Monitor automatically handles PR creation:

```
Agent finishes task
        ↓
   Check branch status (VK API)
        ↓
   ┌─── 0 commits + far behind? → Archive stale attempt
   │
   ├─── Uncommitted changes only? → Prompt agent to commit
   │
   └─── Commits exist:
         ↓
      Rebase onto main (VK API)
         ↓
      ┌── Conflicts? → Auto-resolve (VK API)
      │       ↓
      │   Still conflicts? → Prompt agent to resolve
      │
      └── Clean rebase:
            ↓
         Create PR (VK API, triggers prepush hooks)
            ↓
         ┌── Success → Notify via Telegram
         │
         ├── Fast fail (<2s) → Worktree issue, prompt agent
         │
         └── Slow fail (>30s) → Prepush hook failure, prompt agent to fix
```

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                     codex-monitor                         │
│                                                           │
│  monitor.mjs ─── orchestrator supervisor                  │
│       │     └─── log analysis, error detection            │
│       │     └─── smart PR flow (VK API)                   │
│       │     └─── task planner auto-trigger                │
│       │                                                   │
│       ├── telegram-bot.mjs ─── interactive chatbot        │
│       │       └─── commands, free-text AI chat            │
│       │                                                   │
│       ├── codex-shell.mjs ─── persistent Codex session    │
│       │       └─── MCP tool access (GitHub, VK, files)    │
│       │                                                   │
│       ├── autofix.mjs ─── error loop detection + fix      │
│       │                                                   │
│       └── maintenance.mjs ─── singleton lock, cleanup     │
│                                                           │
│  Integrations:                                            │
│    • Vibe-Kanban API (task management, PR creation)       │
│    • Codex SDK (AI analysis, autofix)                     │
│    • Telegram Bot API (notifications, commands)           │
│    • GitHub CLI (PR operations fallback)                  │
└──────────────────────────────────────────────────────────┘
```

## Using with a Different Repository

Codex Monitor is designed to work with any project that uses Vibe-Kanban for task management:

1. **Run the setup wizard** in your project:

   ```bash
   node setup.mjs
   ```

2. **Point to your orchestrator script** via `--script`:

   ```bash
   node monitor.mjs --script ./my-orchestrator.ps1
   ```

3. **Configure your Vibe-Kanban instance** — set `VK_BASE_URL` to your Kanban API.

4. **Customize AI provider** — set `OPENAI_API_KEY` and optionally `OPENAI_BASE_URL` for:
   - OpenAI (default)
   - Azure OpenAI
   - Local models (Ollama, vLLM, etc.)
   - Any OpenAI-compatible endpoint

### Provider Examples

**OpenAI (default):**

```env
OPENAI_API_KEY=sk-...
```

**Azure OpenAI:**

```env
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment
```

**Local model (Ollama):**

```env
OPENAI_API_KEY=ollama
OPENAI_BASE_URL=http://localhost:11434/v1
CODEX_MODEL=codex
```

## Agent Support Matrix

| Agent Type | Subagents           | Codex Shell | Notes                        |
| ---------- | ------------------- | ----------- | ---------------------------- |
| Copilot    | Yes (`runSubagent`) | Yes         | Full MCP tool access         |
| Codex CLI  | No                  | Yes         | Single-session, sandbox mode |
| Custom     | —                   | Via API     | Bring your own agent         |

The Codex SDK remains the primary driver for the monitor's core AI features (analysis, autofix, shell). When using Copilot as an agent executor, subagent support enables parallel task delegation.

## Troubleshooting

### Telegram 409 errors

> `Conflict: terminated by other getUpdates request`

Only one process can poll a Telegram bot at a time. Codex Monitor auto-disables its internal polling when the telegram-bot module is active. If you see this error, ensure only one monitor instance is running (the singleton lock prevents duplicates automatically).

### "Agent must push before PR"

The Smart PR flow handles this automatically. When this log line appears, the monitor detects it and triggers the VK API flow (rebase → resolve conflicts → create PR → fallback to agent prompt).

### Vibe-Kanban not reachable

The monitor auto-spawns Vibe-Kanban if not running. If it keeps failing:

- Check that the vibe-kanban binary is on your PATH
- Verify `VK_BASE_URL` and `VK_RECOVERY_PORT` match your setup
- Set `VK_NO_SPAWN=1` to manage VK separately

### Codex analysis not working

- Ensure `OPENAI_API_KEY` is set
- Check that `CODEX_SDK_DISABLED` is not `1`
- The Codex SDK auto-installs on first run; check npm/pnpm availability

## File Structure

```
codex-monitor/
├── monitor.mjs              # Main supervisor (entry point)
├── telegram-bot.mjs         # Interactive Telegram chatbot
├── codex-shell.mjs          # Persistent Codex SDK session
├── autofix.mjs              # Error loop detection + auto-fix
├── maintenance.mjs          # Singleton lock, process cleanup
├── setup.mjs                # Interactive setup wizard
├── get-telegram-chat-id.mjs # Telegram chat ID helper
├── .env.example             # Environment variable reference
├── package.json             # NPM package definition
└── logs/                    # Auto-created log directory
```

## License

MIT
