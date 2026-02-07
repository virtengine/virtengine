# @virtengine/codex-monitor

> AI-powered orchestrator supervisor with multi-executor failover, smart PR flow, and Telegram notifications. Install globally, point at any repo, and let AI agents ship code autonomously.

## Install

```bash
npm install -g @virtengine/codex-monitor
```

## Quick Start

```bash
cd your-project
codex-monitor          # auto-detects first run → launches setup wizard
```

That's it. On first run, the setup wizard walks you through everything: executors, AI provider, Telegram, Vibe-Kanban, and agent templates.

## What It Does

```
┌──────────────────────────────────────────────────────────────────┐
│                      @virtengine/codex-monitor                   │
│                                                                  │
│  cli.mjs ────── entry point, first-run detection                 │
│       │                                                          │
│  config.mjs ── unified config (env + JSON + CLI)                 │
│       │                                                          │
│  monitor.mjs ── orchestrator supervisor                          │
│       │    └── log analysis, error detection                     │
│       │    └── smart PR flow (VK API)                            │
│       │    └── executor scheduling & failover                    │
│       │    └── task planner auto-trigger                         │
│       │                                                          │
│       ├── telegram-bot.mjs ── interactive chatbot                │
│       │       └── /status /tasks /restart /stop /agents          │
│       │       └── free-text AI chat                              │
│       │                                                          │
│       ├── codex-shell.mjs ── persistent Codex session            │
│       │       └── MCP tool access (GitHub, VK, files)            │
│       │                                                          │
│       ├── autofix.mjs ── error loop detection + auto-fix         │
│       │                                                          │
│       └── maintenance.mjs ── singleton lock, cleanup             │
│                                                                  │
│  Integrations:                                                   │
│    • Vibe-Kanban API (task management, PR creation)              │
│    • Codex SDK (AI analysis, autofix)                            │
│    • Telegram Bot API (notifications, commands)                  │
│    • GitHub CLI (PR operations fallback)                         │
└──────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Multi-executor scheduling** — Configure N AI agents (Copilot, Codex, custom) with weighted distribution and automatic failover
- **Smart PR flow** — Auto-rebase, resolve merge conflicts, create PRs via Vibe-Kanban API
- **Crash analysis** — Codex SDK reads logs and diagnoses root causes
- **Error loop detection** — 4+ repeating errors in 10 minutes triggers AI autofix
- **Stale attempt cleanup** — Detects dead attempts (0 commits, far behind main) and archives them
- **Telegram chatbot** — Real-time notifications + interactive commands
- **Auto-setup** — First-run wizard configures everything; generates agent templates, wires Vibe-Kanban
- **Multi-repo support** — Manage separate backend/frontend repos from one monitor instance
- **Works with any orchestrator** — Wraps PowerShell, Bash, or any long-running CLI script

## Configuration

Configuration loads from (highest priority first):

1. **CLI flags** (`--script ./myorch.ps1`)
2. **Environment variables** (`ORCHESTRATOR_SCRIPT=...`)
3. **`.env` file** (in codex-monitor directory)
4. **`codex-monitor.config.json`** (project config)
5. **Built-in defaults**

### Setup Wizard

```bash
codex-monitor --setup
```

The wizard configures:

- Project identity (name, repo slug)
- Multi-repo setup (backend + frontend repos)
- Executor profiles (AI models, weights, roles)
- Failover strategy and distribution mode
- AI provider (OpenAI, Azure, Ollama, custom)
- Telegram bot
- Vibe-Kanban connection
- **Orchestrator scripts** — Auto-detects bundled `ve-orchestrator.ps1` and `ve-kanban.ps1` and offers to use them
- Agent template files (AGENTS.md, orchestrator.agent.md, task-planner.agent.md)
- VK workspace scripts (setup, cleanup)

### Executor Configuration

Executors are the AI agents that work on tasks. Configure as many as you want with weights and roles:

```json
// codex-monitor.config.json
{
  "projectName": "my-app",
  "executors": [
    {
      "name": "copilot-claude",
      "executor": "COPILOT",
      "variant": "CLAUDE_OPUS_4_6",
      "weight": 40,
      "role": "primary",
      "enabled": true
    },
    {
      "name": "codex-default",
      "executor": "CODEX",
      "variant": "DEFAULT",
      "weight": 35,
      "role": "backup",
      "enabled": true
    },
    {
      "name": "copilot-gpt",
      "executor": "COPILOT",
      "variant": "GPT_4_1",
      "weight": 25,
      "role": "tertiary",
      "enabled": true
    }
  ],
  "failover": {
    "strategy": "next-in-line",
    "maxRetries": 3,
    "cooldownMinutes": 5,
    "disableOnConsecutiveFailures": 3
  },
  "distribution": "weighted"
}
```

Or via environment variable shorthand:

```env
EXECUTORS=COPILOT:CLAUDE_OPUS_4_6:40,CODEX:DEFAULT:35,COPILOT:GPT_4_1:25
```

#### Distribution Modes

| Mode           | Behavior                                          |
| -------------- | ------------------------------------------------- |
| `weighted`     | Distribute tasks by configured weight percentages |
| `round-robin`  | Alternate between executors equally               |
| `primary-only` | Always use primary; others only for failover      |

#### Failover Strategies

| Strategy          | Behavior                                             |
| ----------------- | ---------------------------------------------------- |
| `next-in-line`    | Use the next executor by role (primary → backup → …) |
| `weighted-random` | Randomly select from remaining by weight             |
| `round-robin`     | Cycle through remaining executors                    |

When an executor fails `disableOnConsecutiveFailures` times in a row, it enters cooldown for `cooldownMinutes` minutes. Tasks automatically route to the next available executor.

### Multi-Repo Support

For projects with separate repositories (e.g., backend + frontend):

```json
{
  "repositories": [
    {
      "name": "backend",
      "path": "/path/to/backend",
      "slug": "org/backend",
      "primary": true
    },
    {
      "name": "frontend",
      "path": "/path/to/frontend",
      "slug": "org/frontend"
    }
  ]
}
```

### Environment Variables

See [.env.example](.env.example) for the full reference. Key variables:

| Variable                | Default                  | Description                        |
| ----------------------- | ------------------------ | ---------------------------------- |
| `PROJECT_NAME`          | auto-detected            | Project name for display           |
| `GITHUB_REPO`           | auto-detected            | GitHub repo slug (`org/repo`)      |
| `OPENAI_API_KEY`        | —                        | API key for Codex analysis         |
| `TELEGRAM_BOT_TOKEN`    | —                        | Telegram bot token from @BotFather |
| `TELEGRAM_CHAT_ID`      | —                        | Telegram chat ID                   |
| `VK_BASE_URL`           | `http://127.0.0.1:54089` | Vibe-Kanban API endpoint           |
| `EXECUTORS`             | Copilot+Codex 50/50      | Executor shorthand (see above)     |
| `EXECUTOR_DISTRIBUTION` | `weighted`               | Distribution mode                  |
| `FAILOVER_STRATEGY`     | `next-in-line`           | Failover behavior                  |
| `MAX_PARALLEL`          | `6`                      | Max concurrent agent slots         |

## CLI Reference

```
codex-monitor [options]
```

| Option                  | Description                               |
| ----------------------- | ----------------------------------------- |
| `--setup`               | Run the interactive setup wizard          |
| `--help`                | Show help                                 |
| `--version`             | Show version                              |
| `--script <path>`       | Path to the orchestrator script           |
| `--args "<args>"`       | Arguments passed to the script            |
| `--restart-delay <ms>`  | Delay before restart (default: `10000`)   |
| `--max-restarts <n>`    | Max restarts, 0 = unlimited               |
| `--log-dir <path>`      | Log directory (default: `./logs`)         |
| `--no-codex`            | Disable Codex SDK analysis                |
| `--no-autofix`          | Disable automatic error fixing            |
| `--no-telegram-bot`     | Disable the interactive Telegram bot      |
| `--no-vk-spawn`         | Don't auto-spawn Vibe-Kanban              |
| `--no-watch`            | Disable file watching for auto-restart    |
| `--no-echo-logs`        | Don't echo orchestrator output to console |
| `--config-dir <path>`   | Directory containing config files         |
| `--repo-root <path>`    | Repository root (auto-detected)           |
| `--project-name <name>` | Project name for display                  |
| `--repo <org/repo>`     | GitHub repo slug                          |

## Telegram Bot

### Setup

1. Open Telegram → search **@BotFather** → `/newbot` → copy token
2. Set `TELEGRAM_BOT_TOKEN` in `.env`
3. Run `codex-monitor-chat-id` → send a message to your bot → copy the chat ID
4. Set `TELEGRAM_CHAT_ID` in `.env`

### Commands

| Command               | Description                                   |
| --------------------- | --------------------------------------------- |
| `/status`             | Current orchestrator status + attempt summary |
| `/tasks`              | List active tasks with progress               |
| `/agents`             | Show executor utilization & health            |
| `/health`             | System health check                           |
| `/restart`            | Restart the orchestrator                      |
| `/stop`               | Gracefully stop the orchestrator              |
| `/reattempt <id>`     | Re-queue a failed task                        |
| `/plan <description>` | Trigger the AI task planner                   |
| Free text             | Chat with Codex AI about the project          |

## Smart PR Flow

```
Agent finishes task
        │
   Check branch status (VK API)
        │
   ┌─── 0 commits + far behind? → Archive stale attempt
   │
   ├─── Uncommitted changes? → Prompt agent to commit
   │
   └─── Commits exist:
         │
      Rebase onto main (VK API)
         │
      ┌── Conflicts? → Auto-resolve (VK API)
      │       │
      │   Still conflicts? → Prompt agent to resolve
      │
      └── Clean rebase:
            │
         Create PR (VK API → triggers pre-push hooks)
            │
         ┌── Success → Notify via Telegram
         ├── Fast fail (<2s) → Worktree issue → prompt agent
         └── Slow fail (>30s) → Pre-push failure → prompt agent to fix
```

## Agent Templates

The setup wizard generates agent template files for your project:

| File                                   | Purpose                                                            |
| -------------------------------------- | ------------------------------------------------------------------ |
| `AGENTS.md`                            | Root-level guide for AI agents (commit conventions, quality gates) |
| `.github/agents/orchestrator.agent.md` | Task orchestrator agent prompt (for Copilot/Codex)                 |
| `.github/agents/task-planner.agent.md` | Auto-creates tasks when backlog is low                             |

These are generic starting points. Customize them for your project's specific needs (build commands, test framework, architecture patterns).

## AI Provider Examples

**OpenAI (default):**

```env
OPENAI_API_KEY=sk-...
```

**Azure OpenAI:**

```env
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/openai/deployments/your-deployment
CODEX_MODEL=your-deployment-name
```

**Local model (Ollama):**

```env
OPENAI_API_KEY=ollama
OPENAI_BASE_URL=http://localhost:11434/v1
CODEX_MODEL=codex
```

## File Structure

```
codex-monitor/
├── cli.mjs                      # CLI entry point + first-run detection
├── config.mjs                   # Unified config loader (env + JSON + CLI)
├── monitor.mjs                  # Main supervisor (log analysis, PR flow)
├── telegram-bot.mjs             # Interactive Telegram chatbot
├── codex-shell.mjs              # Persistent Codex SDK session
├── autofix.mjs                  # Error loop detection + auto-fix
├── maintenance.mjs              # Singleton lock, process cleanup
├── setup.mjs                    # Interactive setup wizard
├── get-telegram-chat-id.mjs     # Telegram chat ID helper
├── ve-orchestrator.ps1          # Default task orchestrator (bundled)
├── ve-kanban.ps1                # Vibe-Kanban CLI wrapper (bundled)
├── codex-monitor.config.json    # Project config (generated by setup)
├── .env                         # Environment variables (generated by setup)
├── .env.example                 # Environment variable reference
├── package.json                 # NPM package definition
└── logs/                        # Auto-created log directory
```

## Default Orchestrator Scripts

The setup wizard automatically detects and offers to use default orchestrator scripts bundled with codex-monitor:

- **ve-orchestrator.ps1** — Main task orchestrator with parallel execution, auto-merge, and CI monitoring
- **ve-kanban.ps1** — Vibe-Kanban CLI wrapper for task management operations

These scripts live directly in the `codex-monitor/` directory, making it self-contained and portable. During setup, if detected, you'll be prompted to use them. If you decline or they're not found, codex-monitor runs in direct Vibe-Kanban mode (manages tasks directly without an external orchestrator script).

### Using Default Scripts

The default scripts expect to be run from the repository root and require:
- Vibe-Kanban API running (`vibe-kanban` CLI installed)
- GitHub CLI (`gh`) for PR operations
- PowerShell 7+ for cross-platform support

Example invocation:
```bash
# Via codex-monitor (recommended)
codex-monitor --script ./ve-orchestrator.ps1 --args "-MaxParallel 6"

# Or directly from the codex-monitor directory
cd scripts/codex-monitor
pwsh ve-orchestrator.ps1 -MaxParallel 6 -PollIntervalSec 90
```

## Troubleshooting

### First-run setup doesn't launch

The auto-detection checks for `.env` or `codex-monitor.config.json`. If either exists, setup won't auto-launch. Run `codex-monitor --setup` manually.

### Telegram 409 errors

> `Conflict: terminated by other getUpdates request`

Only one process can poll a Telegram bot. The monitor auto-disables its polling when `telegram-bot.mjs` is active. Ensure only one monitor instance runs (singleton lock prevents duplicates).

### Executor keeps failing over

Check the executor summary via `/agents` in Telegram. An executor enters cooldown after consecutive failures. Increase `FAILOVER_COOLDOWN_MIN` or `FAILOVER_DISABLE_AFTER` if failovers are too aggressive.

### "Agent must push before PR"

The Smart PR flow handles this automatically. The monitor detects this log line and triggers VK API flow (rebase → resolve conflicts → create PR).

### Vibe-Kanban not reachable

The monitor auto-spawns VK if not running. Set `VK_NO_SPAWN=1` to manage VK separately. Verify `VK_BASE_URL` matches your setup.

## License

MIT
