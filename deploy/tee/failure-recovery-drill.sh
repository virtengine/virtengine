#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
  shift
fi

if [[ $# -lt 1 ]]; then
  echo "usage: $0 [--dry-run] <sgx|sev-snp>" >&2
  exit 1
fi

PLATFORM="$1"
NAMESPACE="${NAMESPACE:-virtengine}"
DEPLOYMENT="tee-enclave-${PLATFORM}"

case "$PLATFORM" in
  sgx|sev-snp) ;;
  *)
    echo "unsupported platform: $PLATFORM" >&2
    exit 1
    ;;
esac

echo "Running failure-recovery drill for ${DEPLOYMENT} in namespace ${NAMESPACE}"

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '+'
    printf ' %q' "$@"
    printf '\n'
    return 0
  fi

  "$@"
}

run kubectl rollout status "deployment/${DEPLOYMENT}" -n "$NAMESPACE" --timeout=5m
run kubectl get externalsecret tee-attestation-material tee-measurement-allowlist -n "$NAMESPACE"

echo "Forcing ExternalSecret refresh"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
run kubectl annotate externalsecret tee-attestation-material -n "$NAMESPACE" "force-sync=${STAMP}" --overwrite
run kubectl annotate externalsecret tee-measurement-allowlist -n "$NAMESPACE" "force-sync=${STAMP}" --overwrite

echo "Restarting deployment to pick up rotated signer material"
run kubectl rollout restart "deployment/${DEPLOYMENT}" -n "$NAMESPACE"
run kubectl rollout status "deployment/${DEPLOYMENT}" -n "$NAMESPACE" --timeout=10m

echo "Checking readiness and alert surfaces"
run kubectl get pods -n "$NAMESPACE" -l "virtengine.com/tee-platform=${PLATFORM}"
run kubectl get prometheusrule tee-enclave -n "$NAMESPACE"
run kubectl get servicemonitor tee-enclave -n "$NAMESPACE"

echo "Drill complete: verify no active TEEEnclaveSignatureFailures or TEEEnclaveStaleAttestations alerts remain firing."
