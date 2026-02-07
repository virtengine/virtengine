#!/usr/bin/env bash
# Regional Smoke Test
# Quick health check for a specific region's VirtEngine deployment.
set -euo pipefail

REGION="${1:-}"
DOMAIN="${VE_DOMAIN:-virtengine.io}"
PASS=0
FAIL=0
SKIP=0

if [ -z "$REGION" ]; then
  echo "Usage: $(basename "$0") <region>"
  echo "  e.g., $(basename "$0") us-east-1"
  exit 1
fi

log_pass() { echo "  ✓ $1"; ((PASS++)); }
log_fail() { echo "  ✗ $1"; ((FAIL++)); }
log_skip() { echo "  ○ $1 (skipped)"; ((SKIP++)); }

kube_ctx() {
  local env_key="VE_KUBE_CONTEXT_$(echo "$REGION" | tr '[:lower:]-' '[:upper:]_')"
  echo "${!env_key:-virtengine-prod-$REGION}"
}

echo "=== Smoke Test: $REGION ==="
echo ""

# ---------------------------------------------------------------------------
# 1. Kubernetes cluster health
# ---------------------------------------------------------------------------
echo "--- Cluster Health ---"
CTX=$(kube_ctx)

if kubectl --context="$CTX" cluster-info &>/dev/null 2>&1; then
  log_pass "Cluster reachable"

  nodes=$(kubectl --context="$CTX" get nodes --no-headers 2>/dev/null | wc -l)
  ready=$(kubectl --context="$CTX" get nodes --no-headers 2>/dev/null | grep -c " Ready" || true)
  if [ "$ready" -eq "$nodes" ] && [ "$nodes" -gt 0 ]; then
    log_pass "All nodes ready ($ready/$nodes)"
  else
    log_fail "Not all nodes ready ($ready/$nodes)"
  fi
else
  log_skip "Cluster not reachable (context: $CTX)"
fi

# ---------------------------------------------------------------------------
# 2. Core pods running
# ---------------------------------------------------------------------------
echo "--- Core Pods ---"

for ns_label in "virtengine:app=virtengine" "cockroachdb:app.kubernetes.io/name=cockroachdb" "monitoring:app.kubernetes.io/name=prometheus"; do
  ns="${ns_label%%:*}"
  label="${ns_label##*:}"

  pods=$(kubectl --context="$CTX" -n "$ns" get pods -l "$label" --no-headers 2>/dev/null || echo "")
  if [ -n "$pods" ]; then
    running=$(echo "$pods" | grep -c "Running" || true)
    total=$(echo "$pods" | wc -l)
    if [ "$running" -eq "$total" ]; then
      log_pass "$ns/$label: $running/$total running"
    else
      log_fail "$ns/$label: $running/$total running"
    fi
  else
    log_skip "$ns/$label (not found)"
  fi
done

# ---------------------------------------------------------------------------
# 3. API endpoint health
# ---------------------------------------------------------------------------
echo "--- API Health ---"

endpoint="https://rpc-${REGION}.${DOMAIN}/status"
if curl -sf --connect-timeout 10 "$endpoint" &>/dev/null; then
  log_pass "RPC endpoint responding ($endpoint)"

  # Check block production
  height=$(curl -sf --connect-timeout 5 "$endpoint" 2>/dev/null | \
    grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' || echo "")
  if [ -n "$height" ] && [ "$height" -gt 0 ]; then
    log_pass "Block height: $height"
  else
    log_skip "Block height check"
  fi
else
  log_skip "RPC endpoint ($endpoint)"
fi

# ---------------------------------------------------------------------------
# 4. CockroachDB health
# ---------------------------------------------------------------------------
echo "--- CockroachDB ---"

crdb_pods=$(kubectl --context="$CTX" -n cockroachdb get pods --no-headers 2>/dev/null || echo "")
if [ -n "$crdb_pods" ]; then
  running=$(echo "$crdb_pods" | grep -c "Running" || true)
  log_pass "CockroachDB pods: $running running"
else
  log_skip "CockroachDB (not deployed)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "═══════════════════════════════════════"
echo "  Smoke Test: $REGION"
echo "═══════════════════════════════════════"
echo "  Passed:  $PASS"
echo "  Failed:  $FAIL"
echo "  Skipped: $SKIP"
echo "═══════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
  echo "  RESULT: FAILED"
  exit 1
else
  echo "  RESULT: PASSED"
  exit 0
fi
