# @virtengine/codex-monitor

Autonomous supervisor for AI coding workflows.

`codex-monitor` watches task execution, routes work across agent executors, handles retries/failover, automates PR lifecycle, and keeps you in control through Telegram (with optional WhatsApp and container isolation).

---

## Why codex-monitor

AI coding agents are fast, but unattended loops are expensive:

- silent failures
- repeated retries with no progress
- stale worktrees and merge drift
- disconnected notifications

`codex-monitor` is the control plane that keeps delivery moving:

- task routing and executor failover
- monitored orchestration and auto-recovery
- conflict/PR lifecycle automation
- live bot control (`/status`, `/tasks`, `/pause`, `/resume`, etc.)

---

## Install

```bash
npm install -g @virtengine/codex-monitor
```

Requires:

- Node.js 18+
- git
- PowerShell (`pwsh`) for VirtEngine default scripts
- GitHub CLI (`gh`) recommended

---

## Quick start

```bash
cd your-repo
codex-monitor
```

First run launches setup automatically.

You can also run setup directly:

```bash
codex-monitor --setup
```

---

## Setup modes (new)

The setup wizard now starts with two modes:

- **Recommended**
  - prompts only for important decisions (project identity, executor preset/model profile, AI provider, Telegram, board/execution mode)
  - keeps advanced knobs on proven defaults
  - writes a standardized `.env` based on `.env.example` so all options remain documented

- **Advanced**
  - full control over repository layout, failover/distribution, hook targets/overrides, orchestrator path, VK wiring details, and optional channels

---

## How codex-monitor can run

### 1) Standard foreground supervisor

```bash
codex-monitor
```

### 2) Daemon mode

```bash
codex-monitor --daemon
codex-monitor --daemon-status
codex-monitor --stop-daemon
```

### 3) Startup service (auto-start on login)

```bash
codex-monitor --enable-startup
codex-monitor --startup-status
codex-monitor --disable-startup
```

### 4) Interactive shell mode

```bash
codex-monitor --shell
```

### 5) Sentinel companion mode (Telegram watchdog)

```bash
codex-monitor --sentinel
codex-monitor --sentinel-status
codex-monitor --sentinel-stop
```

---

## Execution architecture modes

Configured by `EXECUTOR_MODE`:

- `internal` (recommended)
  - tasks run through internal agent pool in monitor process
- `vk`
  - task execution delegated to VK orchestrator flow
- `hybrid`
  - internal + VK behavior for mixed/overflow scenarios

Task board backend (`KANBAN_BACKEND`):

- `vk` - Vibe-Kanban (default)
- `github` - GitHub Issues with shared state persistence
- `jira` (scaffolded, not yet implemented)

**GitHub adapter enhancements:**
The GitHub Issues adapter now supports multi-agent coordination via structured state persistence:

- Claim tracking with `codex:claimed`, `codex:working`, `codex:stale` labels
- Heartbeat mechanism to detect stale/abandoned claims
- Task exclusion via `codex:ignore` label
- Structured comments with JSON state for agent coordination

See [KANBAN_GITHUB_ENHANCEMENT.md](./KANBAN_GITHUB_ENHANCEMENT.md) for details.

**Jira adapter (future):**
The Jira adapter is scaffolded with detailed JSDoc and implementation guidance:

- Method stubs: `persistSharedStateToIssue()`, `readSharedStateFromIssue()`, `markTaskIgnored()`
- Planned approach using Jira custom fields, labels, or structured comments
- Compatible API surface with GitHub adapter for drop-in replacement

See [JIRA_INTEGRATION.md](./JIRA_INTEGRATION.md) for implementation guide.

---

## Channels and control surfaces

### Telegram (primary control channel)

Core controls include:

- `/help` (inline keyboard)
- `/status`, `/tasks`, `/agents`, `/threads`, `/worktrees`
- `/pause`, `/resume`, `/restart`, `/retry`
- `/executor`, `/sdk`, `/kanban`, `/maxparallel`

### WhatsApp (optional)

Enable in env and authenticate:

```bash
codex-monitor --whatsapp-auth
# or
codex-monitor --whatsapp-auth --pairing-code
```

Telegram status commands include:

- `/whatsapp`
- `/container`

---

## Container isolation (optional)

`container-runner` can isolate agent executions with:

- Docker
- Podman
- Apple Container (`container`) on macOS

Key env vars:

- `CONTAINER_ENABLED=1`
- `CONTAINER_RUNTIME=auto|docker|podman|container`
- `CONTAINER_IMAGE=node:22-slim`
- `MAX_CONCURRENT_CONTAINERS=3`

---

## Configuration model

Load order (highest priority first):

1. CLI flags
2. environment variables
3. `.env`
4. `codex-monitor.config.json`
5. built-in defaults

### Files

- `.env` — runtime/environment settings
- `codex-monitor.config.json` — structured config (executors, failover, repos, profiles)
- `.codex-monitor/agents/*.md` — prompt templates scaffolded by setup

### SDK transport defaults

`codex-monitor` supports explicit transport selectors per SDK shell:

- `CODEX_TRANSPORT=sdk|auto|cli`
- `COPILOT_TRANSPORT=sdk|auto|cli|url`
- `CLAUDE_TRANSPORT=sdk|auto|cli`

Setup now defaults all three to `sdk` for predictable persistent-session behavior.
`auto` remains available when you intentionally want endpoint/CLI auto-detection.

### Recommended profile split

- Local development profile:
  - `DEVMODE=true`
  - `DEVMODE_MONITOR_MONITOR_ENABLED=true`
  - `*_TRANSPORT=sdk`
- End-user stability profile:
  - `DEVMODE=false`
  - `DEVMODE_MONITOR_MONITOR_ENABLED=false`
  - `*_TRANSPORT=sdk`

---

## Recommended configuration path

If you want a strong baseline with minimal decisions:

1. Run `codex-monitor --setup`
2. Pick **Recommended** mode
3. Choose executor preset that matches your token budget and speed goals
4. Configure AI provider credentials
5. Configure Telegram
6. Keep defaults for hooks/VK/orchestrator unless you already have a custom flow

This gives you a standardized `.env` with full inline documentation and sane defaults.

---

## Advanced configuration path

Use **Advanced** mode when you need:

- custom multi-repo topology
- custom failover/distribution behavior
- manual orchestrator path/args
- custom hook policy and event overrides
- explicit VK URL/port and wiring behavior
- explicit optional channel/runtime tuning

---

## Key config examples

### Executor config (`codex-monitor.config.json`)

```json
{
  "$schema": "./codex-monitor.schema.json",
  "projectName": "my-project",
  "executors": [
    {
      "name": "copilot-claude",
      "executor": "COPILOT",
      "variant": "CLAUDE_OPUS_4_6",
      "weight": 50,
      "role": "primary",
      "enabled": true
    },
    {
      "name": "codex-default",
      "executor": "CODEX",
      "variant": "DEFAULT",
      "weight": 50,
      "role": "backup",
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

### Env shorthand for executors

```env
EXECUTORS=COPILOT:CLAUDE_OPUS_4_6:50,CODEX:DEFAULT:50
```

### Minimal local env

```env
PROJECT_NAME=my-project
GITHUB_REPO=myorg/myrepo
KANBAN_BACKEND=vk
EXECUTOR_MODE=internal
VK_BASE_URL=http://127.0.0.1:54089
VK_RECOVERY_PORT=54089
MAX_PARALLEL=6
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

For full variable documentation see `.env.example`.

---

## Useful commands

```bash
codex-monitor --help
codex-monitor --setup
codex-monitor --doctor
codex-monitor --update
codex-monitor --no-update-check
codex-monitor --no-auto-update
codex-monitor --no-telegram-bot
codex-monitor --telegram-commands
codex-monitor --no-vk-spawn
codex-monitor --vk-ensure-interval 60000
codex-monitor --script ./my-orchestrator.ps1
codex-monitor --args "-MaxParallel 6"
```

`--doctor` validates effective config (.env + config JSON + process env overrides), reports actionable fixes, and exits non-zero when blocking issues are found.

---

## Validation and tests

From `scripts/codex-monitor`:

```bash
npm run syntax:check
npm run test
```

Focused tests (example):

```bash
npx vitest run tests/whatsapp-channel.test.mjs tests/container-runner.test.mjs tests/telegram-buttons.test.mjs
```

---

## Notes on generated `.env`

The setup wizard now standardizes `.env` generation by applying your selected values onto `.env.example`.

Benefits:

- all options stay documented in your generated file
- chosen values are explicitly activated/uncommented
- unchosen options remain visible as commented documentation
- easier upgrades over time as new options are added

---

## Troubleshooting quick checks

- `codex-monitor --help` for supported flags
- `codex-monitor --setup` to re-run configuration safely
- verify `.env` + `codex-monitor.config.json` are in your config directory
- verify `gh auth status` for PR operations
- verify Telegram token/chat id with `codex-monitor-chat-id`

---

## License

Apache 2.0
