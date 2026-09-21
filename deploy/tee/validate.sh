#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RENDERED="$(mktemp)"
trap 'rm -f "$RENDERED"' EXIT

KUBECTL_BIN="${KUBECTL_BIN:-}"
if [[ -z "$KUBECTL_BIN" ]]; then
  if command -v kubectl >/dev/null 2>&1; then
    KUBECTL_BIN="$(command -v kubectl)"
  elif command -v kubectl.exe >/dev/null 2>&1; then
    KUBECTL_BIN="$(command -v kubectl.exe)"
  else
    echo "kubectl or kubectl.exe is required" >&2
    exit 1
  fi
fi

KUSTOMIZE_ROOT="$ROOT"
if [[ "$KUBECTL_BIN" == *.exe ]]; then
  if command -v wslpath >/dev/null 2>&1; then
    KUSTOMIZE_ROOT="$(wslpath -w "$ROOT")"
  elif command -v cygpath >/dev/null 2>&1; then
    KUSTOMIZE_ROOT="$(cygpath -w "$ROOT")"
  fi
fi

"$KUBECTL_BIN" kustomize "$KUSTOMIZE_ROOT" >"$RENDERED"

for required in \
  "kind: Deployment" \
  "name: tee-enclave-sgx" \
  "name: tee-enclave-sev-snp" \
  "kind: ExternalSecret" \
  "name: tee-attestation-material" \
  "name: tee-measurement-allowlist" \
  "kind: ServiceMonitor" \
  "kind: PrometheusRule"; do
  grep -q "$required" "$RENDERED"
done

if grep -Eiq '[Pp][Ee][Nn][Dd][Ii][Nn][Gg]|[Tt][Bb][Dd]|placeh[o]lder|not published y[e]t' "$RENDERED"; then
  echo "rendered manifests contain forbidden launch tokens" >&2
  exit 1
fi

echo "deploy/tee manifests rendered and passed structural validation"
