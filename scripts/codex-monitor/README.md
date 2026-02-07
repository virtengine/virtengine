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

## Notes

- The Codex SDK uses environment configuration from your Codex setup (API key, base URL).
- If Codex analysis fails, the error is written next to the log file.
