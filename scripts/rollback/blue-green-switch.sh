#!/usr/bin/env bash
# VirtEngine blue/green traffic switch helper with Prometheus-backed health gates.

set -euo pipefail

APP=""
TARGET=""
MODE="gradual"
AUTO_APPROVE="false"
DRY_RUN="false"
NAMESPACE="${NAMESPACE:-virtengine}"
WEIGHT_STEP="${WEIGHT_STEP:-10}"
WAIT_SECONDS="${WAIT_SECONDS:-30}"
ERROR_RATE_THRESHOLD="${ERROR_RATE_THRESHOLD:-0.02}"
PROMETHEUS_URL="${PROMETHEUS_URL:-}"
PROMETHEUS_QUERY_TEMPLATE="${PROMETHEUS_QUERY_TEMPLATE:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

usage() {
  cat <<'EOF'
VirtEngine Blue/Green Traffic Switch Script

Usage:
  blue-green-switch.sh <app> <target-version> [--mode gradual|instant] [--dry-run] [--yes]

Arguments:
  app              Application name (virtengine-node, provider-daemon)
  target-version   Target version to switch to (blue, green)

Options:
  --mode MODE      gradual (default) or instant
  --dry-run        Print the operations without patching the VirtualService
  --yes            Skip the interactive confirmation prompt

Environment Variables:
  NAMESPACE                  Kubernetes namespace (default: virtengine)
  WEIGHT_STEP                Traffic weight increment for gradual switches (default: 10)
  WAIT_SECONDS               Wait time between gradual switch steps (default: 30)
  PROMETHEUS_URL             Prometheus base URL, required for gradual switches
  ERROR_RATE_THRESHOLD       Maximum tolerated 5xx ratio during gradual switches (default: 0.02)
  PROMETHEUS_QUERY_TEMPLATE  Optional query template using {app}, {target}, and {namespace}
EOF
}

fail() {
  log_error "$*"
  exit 1
}

require_commands() {
  local commands=(kubectl jq)
  for cmd in "${commands[@]}"; do
    command -v "$cmd" >/dev/null 2>&1 || fail "$cmd not found in PATH"
  done
  if [[ "$MODE" == "gradual" ]]; then
    command -v curl >/dev/null 2>&1 || fail "curl not found in PATH"
  fi
}

replace_query_tokens() {
  local template="$1"
  template="${template//\{app\}/$APP}"
  template="${template//\{target\}/$TARGET}"
  template="${template//\{namespace\}/$NAMESPACE}"
  echo "$template"
}

url_encode() {
  local raw="$1"
  if command -v python >/dev/null 2>&1; then
    python - <<'PY' "$raw"
import sys
from urllib.parse import quote_plus
print(quote_plus(sys.argv[1]))
PY
  else
    pwsh -NoProfile -Command "[uri]::EscapeDataString($args[0])" -- "$raw"
  fi
}

prometheus_query() {
  local query="$1"
  local encoded
  encoded="$(url_encode "$query")"
  curl -fsSL "${PROMETHEUS_URL%/}/api/v1/query?query=${encoded}"
}

default_error_rate_query() {
  local service_fqdn="${APP}-${TARGET}.${NAMESPACE}.svc.cluster.local"
  cat <<EOF
(
  sum(rate(istio_requests_total{reporter="destination",destination_service_name="${service_fqdn}",response_code=~"5.."}[5m]))
)
/
clamp_min(
  sum(rate(istio_requests_total{reporter="destination",destination_service_name="${service_fqdn}"}[5m])),
  1
)
EOF
}

get_current_weights() {
  kubectl get virtualservice "$APP" -n "$NAMESPACE" -o json | \
    jq -r '.spec.http[-1].route[] | "\(.destination.host): \(.weight)"'
}

check_target_health() {
  local ready
  ready="$(kubectl get pods -n "$NAMESPACE" \
    -l "app.kubernetes.io/name=$APP,app.kubernetes.io/version=$TARGET" \
    -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}')"

  [[ -n "$ready" ]] || fail "no candidate pods found for $APP/$TARGET"
  [[ "$ready" == *"False"* ]] && fail "target pods are not ready"
}

update_weights() {
  local blue_weight="$1"
  local green_weight="$2"
  local patch
  patch="[
    {\"op\": \"replace\", \"path\": \"/spec/http/-1/route/0/weight\", \"value\": ${blue_weight}},
    {\"op\": \"replace\", \"path\": \"/spec/http/-1/route/1/weight\", \"value\": ${green_weight}}
  ]"

  log_info "Setting weights: blue=${blue_weight}, green=${green_weight}"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "$patch"
    return 0
  fi

  kubectl patch virtualservice "$APP" -n "$NAMESPACE" --type=json -p="$patch" >/dev/null
}

check_error_rate() {
  [[ -n "$PROMETHEUS_URL" ]] || fail "PROMETHEUS_URL is required for gradual switches"

  local query
  if [[ -n "$PROMETHEUS_QUERY_TEMPLATE" ]]; then
    query="$(replace_query_tokens "$PROMETHEUS_QUERY_TEMPLATE")"
  else
    query="$(default_error_rate_query)"
  fi

  local result
  result="$(prometheus_query "$query")"
  local status
  status="$(echo "$result" | jq -r '.status')"
  [[ "$status" == "success" ]] || fail "Prometheus query failed"

  local value
  value="$(echo "$result" | jq -r '.data.result[0].value[1] // "0"')"
  awk -v value="$value" -v threshold="$ERROR_RATE_THRESHOLD" 'BEGIN { exit (value <= threshold ? 0 : 1) }'
}

instant_rollback() {
  if [[ "$TARGET" == "green" ]]; then
    update_weights 100 0
  else
    update_weights 0 100
  fi
}

gradual_switch() {
  local blue_weight=100
  local green_weight=0

  if [[ "$TARGET" == "blue" ]]; then
    blue_weight=0
    green_weight=100
  fi

  while true; do
    if [[ "$TARGET" == "green" ]]; then
      blue_weight=$(( blue_weight - WEIGHT_STEP ))
      green_weight=$(( green_weight + WEIGHT_STEP ))
      (( green_weight > 100 )) && green_weight=100
      (( blue_weight < 0 )) && blue_weight=0
    else
      green_weight=$(( green_weight - WEIGHT_STEP ))
      blue_weight=$(( blue_weight + WEIGHT_STEP ))
      (( blue_weight > 100 )) && blue_weight=100
      (( green_weight < 0 )) && green_weight=0
    fi

    update_weights "$blue_weight" "$green_weight"

    if [[ "$TARGET" == "green" && "$green_weight" -eq 100 ]] || [[ "$TARGET" == "blue" && "$blue_weight" -eq 100 ]]; then
      break
    fi

    log_info "Waiting ${WAIT_SECONDS}s before the next increment"
    [[ "$DRY_RUN" == "true" ]] || sleep "$WAIT_SECONDS"

    if ! check_error_rate; then
      log_error "Error rate exceeded threshold during gradual switch"
      instant_rollback
      exit 1
    fi
  done
}

confirm_switch() {
  if [[ "$AUTO_APPROVE" == "true" || "$DRY_RUN" == "true" ]]; then
    return 0
  fi

  echo
  log_warn "About to switch traffic"
  echo "  Application: $APP"
  echo "  Namespace: $NAMESPACE"
  echo "  Target: $TARGET"
  echo "  Mode: $MODE"
  echo
  get_current_weights
  echo
  read -r -p "Type 'yes' to continue: " confirm
  [[ "$confirm" == "yes" ]] || fail "switch cancelled"
}

if [[ $# -lt 2 ]]; then
  usage >&2
  exit 1
fi

APP="$1"
TARGET="$2"
shift 2

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="$2"
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

[[ "$TARGET" =~ ^(blue|green)$ ]] || fail "target must be blue or green"
[[ "$MODE" =~ ^(gradual|instant)$ ]] || fail "--mode must be gradual or instant"

require_commands
kubectl get virtualservice "$APP" -n "$NAMESPACE" >/dev/null 2>&1 || fail "VirtualService '$APP' not found in namespace '$NAMESPACE'"
check_target_health
confirm_switch

if [[ "$MODE" == "instant" ]]; then
  if [[ "$TARGET" == "green" ]]; then
    update_weights 0 100
  else
    update_weights 100 0
  fi
else
  gradual_switch
fi

log_info "Traffic switch completed"
