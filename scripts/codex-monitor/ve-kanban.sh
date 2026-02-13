#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PS1_SCRIPT="${SCRIPT_DIR}/ve-kanban.ps1"

if [[ ! -f "${PS1_SCRIPT}" ]]; then
  echo "[ve-kanban.sh] Missing script: ${PS1_SCRIPT}" >&2
  exit 1
fi

PWSH_BIN="${PWSH_PATH:-}"
if [[ -z "${PWSH_BIN}" ]]; then
  if command -v pwsh >/dev/null 2>&1; then
    PWSH_BIN="pwsh"
  elif command -v powershell >/dev/null 2>&1; then
    PWSH_BIN="powershell"
  fi
fi

if [[ -z "${PWSH_BIN}" ]]; then
  echo "[ve-kanban.sh] PowerShell runtime not found (pwsh/powershell)." >&2
  echo "[ve-kanban.sh] Install pwsh or run codex-monitor in internal executor mode." >&2
  exit 1
fi

exec "${PWSH_BIN}" -File "${PS1_SCRIPT}" "$@"
