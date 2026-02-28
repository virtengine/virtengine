#!/usr/bin/env bash
# Regional smoke test for a specific VirtEngine deployment.
set -euo pipefail

REGION="${1:-}"
DOMAIN="${VE_DOMAIN:-virtengine.com}"
BLOCK_PROGRESS_WAIT_SECONDS="${VE_BLOCK_PROGRESS_WAIT_SECONDS:-8}"
REGION_ENV_KEY="$(printf '%s' "${REGION}" | tr '[:lower:]-' '[:upper:]_')"
PASS=0
FAIL=0

usage() {
  echo "Usage: $(basename "$0") <region>"
  echo "  Example: $(basename "$0") us-east-1"
}

if [ -z "${REGION}" ]; then
  usage
  exit 1
fi

log_info() { echo "[INFO] $*"; }
log_pass() { echo "[PASS] $*"; PASS=$((PASS + 1)); }
log_fail() { echo "[FAIL] $*"; FAIL=$((FAIL + 1)); }
fail() { log_fail "$*"; exit 1; }

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "Required command not found: $1"
  fi
}

kube_context() {
  local env_key="VE_KUBE_CONTEXT_${REGION_ENV_KEY}"
  echo "${!env_key:-virtengine-prod-${REGION}}"
}

kube_cluster_name() {
  local env_key="VE_KUBE_CLUSTER_NAME_${REGION_ENV_KEY}"
  echo "${!env_key:-virtengine-prod-${REGION}}"
}

rpc_url() {
  if [ -n "${VE_RPC_URL:-}" ]; then
    echo "${VE_RPC_URL}"
  else
    echo "https://rpc-${REGION}.${DOMAIN}"
  fi
}

extract_height() {
  local payload="$1"
  local height
  height="$(printf '%s' "${payload}" | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' | head -1 || true)"
  if [ -z "${height}" ]; then
    fail "RPC status payload missing latest_block_height"
  fi
  echo "${height}"
}

bootstrap_kubeconfig() {
  if kubectl --context="${CTX}" cluster-info >/dev/null 2>&1; then
    return 0
  fi

  require_cmd aws
  log_info "Bootstrapping kubeconfig for ${CLUSTER_NAME} in ${REGION}"
  aws eks update-kubeconfig --region "${REGION}" --name "${CLUSTER_NAME}" --alias "${CTX}" >/dev/null
  kubectl --context="${CTX}" cluster-info >/dev/null 2>&1 || fail "Cluster not reachable (context: ${CTX})"
}

check_nodes_ready() {
  local nodes ready total
  nodes="$(kubectl --context="${CTX}" get nodes --no-headers)"
  if [ -z "${nodes}" ]; then
    fail "Cluster returned no nodes"
  fi

  total="$(printf '%s\n' "${nodes}" | sed '/^$/d' | wc -l | tr -d ' ')"
  ready="$(printf '%s\n' "${nodes}" | grep -c ' Ready' || true)"
  if [ "${ready}" -ne "${total}" ]; then
    fail "Not all nodes are ready (${ready}/${total})"
  fi

  log_pass "All nodes ready (${ready}/${total})"
}

check_pods() {
  local namespace="$1"
  local label="$2"
  local pods total running

  pods="$(kubectl --context="${CTX}" -n "${namespace}" get pods -l "${label}" --no-headers 2>/dev/null || true)"
  if [ -z "${pods}" ]; then
    fail "No pods found for ${namespace}/${label}"
  fi

  total="$(printf '%s\n' "${pods}" | sed '/^$/d' | wc -l | tr -d ' ')"
  running="$(printf '%s\n' "${pods}" | grep -c 'Running' || true)"
  if [ "${running}" -ne "${total}" ]; then
    fail "${namespace}/${label} pods not fully running (${running}/${total})"
  fi

  log_pass "${namespace}/${label} pods running (${running}/${total})"
}

echo "=== Smoke Test: ${REGION} ==="
echo

require_cmd curl
require_cmd kubectl

CTX="$(kube_context)"
CLUSTER_NAME="$(kube_cluster_name)"
RPC_URL="$(rpc_url)"

log_info "Using cluster context ${CTX}"
bootstrap_kubeconfig
log_pass "Cluster reachable (${CTX})"

echo "--- Cluster Health ---"
check_nodes_ready

echo "--- Core Pods ---"
check_pods "virtengine" "app=virtengine"
check_pods "cockroachdb" "app.kubernetes.io/name=cockroachdb"
check_pods "monitoring" "app.kubernetes.io/name=prometheus"

echo "--- RPC Health ---"
status_one="$(curl -sf --connect-timeout 10 "${RPC_URL}/status")"
height_one="$(extract_height "${status_one}")"
log_info "Initial block height: ${height_one}"

curl -sf --connect-timeout 10 "${RPC_URL}/health" >/dev/null
log_pass "RPC health endpoint responding (${RPC_URL}/health)"

sleep "${BLOCK_PROGRESS_WAIT_SECONDS}"

status_two="$(curl -sf --connect-timeout 10 "${RPC_URL}/status")"
height_two="$(extract_height "${status_two}")"
if [ "${height_two}" -le "${height_one}" ]; then
  fail "Block height did not advance (${height_one} -> ${height_two})"
fi
log_pass "Block height advanced (${height_one} -> ${height_two})"

echo
echo "======================================="
echo "Smoke Test: ${REGION}"
echo "======================================="
echo "Passed: ${PASS}"
echo "Failed: ${FAIL}"
echo "======================================="

if [ "${FAIL}" -gt 0 ]; then
  echo "RESULT: FAILED"
  exit 1
fi

echo "RESULT: PASSED"
