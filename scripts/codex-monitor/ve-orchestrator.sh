#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PS1_SCRIPT="${SCRIPT_DIR}/ve-orchestrator.ps1"

if [[ ! -f "${PS1_SCRIPT}" ]]; then
  echo "[ve-orchestrator.sh] Missing script: ${PS1_SCRIPT}" >&2
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
  echo "[ve-orchestrator.sh] PowerShell runtime not found (pwsh/powershell)." >&2
  echo "[ve-orchestrator.sh] Install pwsh or use internal executor mode (EXECUTOR_MODE=internal)." >&2
  exit 1
fi

exec "${PWSH_BIN}" -File "${PS1_SCRIPT}" "$@"
