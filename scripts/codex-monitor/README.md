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

### Options

- `--script <path>`: Path to the PowerShell script. Default: `scripts/ve-orchestrator.ps1`.
- `--args "<args>"`: Arguments passed to the script. Default: `-MaxParallel 6`.
- `--restart-delay <ms>`: Delay before restart. Default: 10000.
- `--max-restarts <n>`: Max restarts (0 = unlimited).
- `--log-dir <path>`: Log directory. Default: `scripts/codex-monitor/logs`.
- `--watch-path <path>`: File or directory to watch for restarts. Default: `scripts/ve-orchestrator.ps1`.
- `--no-codex`: Disable Codex SDK analysis.
- `--no-watch`: Disable file watching.

## Telegram heartbeat

The monitor can post a lightweight heartbeat and ingest Telegram commands.

### Environment variables

- `TELEGRAM_BOT_TOKEN`: Telegram bot token.
- `TELEGRAM_CHAT_ID`: Chat ID to send updates to and read commands from.
- `TELEGRAM_HEARTBEAT_INTERVAL_MIN`: Heartbeat interval in minutes (default: `5`).
- `TELEGRAM_HEARTBEAT_ENABLED`: Set to `false`/`0` to disable heartbeats (default: enabled).
- `TELEGRAM_QUIET_HOURS`: Quiet hours range like `22-7` (optional).
- `TELEGRAM_QUIET_HOURS_START` / `TELEGRAM_QUIET_HOURS_END`: Quiet hour range (0-23).
- `TELEGRAM_QUIET_HOURS_TZ`: IANA timezone (default: `UTC`).

### Telegram commands

- `/status` — post a heartbeat immediately.
- `/heartbeat on|off` — toggle heartbeat messages.
- `/quiet 22-7 [Timezone]` — enable quiet hours.
- `/quiet off` — disable quiet hours.
- `/ping` — health check.

## Notes

- The Codex SDK uses environment configuration from your Codex setup (API key, base URL).
- If Codex analysis fails, the error is written next to the log file.
