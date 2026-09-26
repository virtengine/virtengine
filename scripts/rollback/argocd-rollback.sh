#!/usr/bin/env bash
# VirtEngine ArgoCD rollback helper with evidence capture.

set -euo pipefail

APP_NAME=""
REVISION=""
ARGOCD_SERVER="${ARGOCD_SERVER:-argocd.virtengine.com}"
TIMEOUT="${TIMEOUT:-600}"
DRY_RUN="false"
AUTO_APPROVE="false"
ARTIFACT_DIR=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

usage() {
  cat <<'EOF'
VirtEngine ArgoCD Rollback Script

Usage:
  argocd-rollback.sh <app-name> [revision] [--artifact-dir DIR] [--dry-run] [--yes]
EOF
}

fail() {
  log_error "$*"
  exit 1
}

check_prerequisites() {
  local commands=(argocd kubectl jq)
  for cmd in "${commands[@]}"; do
    command -v "$cmd" >/dev/null 2>&1 || fail "$cmd not found in PATH"
  done
}

previous_revision() {
  argocd app get "$APP_NAME" -o json | jq -r 'if (.status.history | length) > 1 then .status.history[-2].id else empty end'
}

capture_artifacts() {
  mkdir -p "$ARTIFACT_DIR"
  argocd app get "$APP_NAME" -o json > "$ARTIFACT_DIR/app.json"
  argocd app history "$APP_NAME" -o json > "$ARTIFACT_DIR/history.json"
  argocd app manifests "$APP_NAME" > "$ARTIFACT_DIR/manifests.yaml"

  local destination_namespace
  destination_namespace="$(jq -r '.spec.destination.namespace // "virtengine"' "$ARTIFACT_DIR/app.json")"
  kubectl get all -n "$destination_namespace" -l "app.kubernetes.io/part-of=virtengine-platform" -o yaml \
    > "$ARTIFACT_DIR/pre-rollback-cluster-snapshot.yaml" 2>/dev/null || true
}

write_summary() {
  local status="$1"
  cat > "$ARTIFACT_DIR/summary.md" <<EOF
# ArgoCD Rollback Summary

- Application: $APP_NAME
- Target Revision: $REVISION
- Dry Run: $DRY_RUN
- Status: $status
- ArgoCD Server: $ARGOCD_SERVER
- Timeout Seconds: $TIMEOUT
EOF
}

confirm_rollback() {
  if [[ "$AUTO_APPROVE" == "true" || "$DRY_RUN" == "true" ]]; then
    return 0
  fi

  echo
  log_warn "About to roll back ArgoCD application"
  echo "  Application: $APP_NAME"
  echo "  Target revision: $REVISION"
  echo
  read -r -p "Type 'yes' to continue: " confirm
  [[ "$confirm" == "yes" ]] || fail "rollback cancelled"
}

if [[ $# -lt 1 ]]; then
  usage >&2
  exit 1
fi

APP_NAME="$1"
shift

if [[ $# -gt 0 && ! "$1" =~ ^-- ]]; then
  REVISION="$1"
  shift
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --artifact-dir)
      ARTIFACT_DIR="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN="true"
      shift
      ;;
    --yes)
      AUTO_APPROVE="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown option: $1"
      ;;
  esac
done

check_prerequisites
argocd app list >/dev/null 2>&1 || fail "not logged into ArgoCD. Run: argocd login $ARGOCD_SERVER"
argocd app get "$APP_NAME" >/dev/null 2>&1 || fail "application '$APP_NAME' not found"

if [[ -z "$REVISION" ]]; then
  REVISION="$(previous_revision)"
  [[ -n "$REVISION" ]] || fail "no previous revision available for rollback"
fi

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT_DIR="${ARTIFACT_DIR:-$(pwd)/output/rollback/argocd/$APP_NAME/$TIMESTAMP}"
capture_artifacts
confirm_rollback

if [[ "$DRY_RUN" == "true" ]]; then
  write_summary "resolved"
  log_info "Dry run complete. Artifacts written to $ARTIFACT_DIR"
  exit 0
fi

argocd app rollback "$APP_NAME" "$REVISION" --prune
argocd app wait "$APP_NAME" --timeout "$TIMEOUT" --health --sync
argocd app get "$APP_NAME" -o json > "$ARTIFACT_DIR/post-rollback-app.json"
write_summary "completed"

log_info "Rollback completed. Artifacts written to $ARTIFACT_DIR"
