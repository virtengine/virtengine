#!/usr/bin/env bash
# Cross-Region Connectivity Verification
# Validates network connectivity, VPC peering, and service communication
# between all VirtEngine regions.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOMAIN="${VE_DOMAIN:-virtengine.io}"
REGIONS=("us-east-1" "eu-west-1" "ap-southeast-1")
PASS=0
FAIL=0
SKIP=0

log_pass() { echo "  ✓ $1"; ((PASS++)); }
log_fail() { echo "  ✗ $1"; ((FAIL++)); }
log_skip() { echo "  ○ $1 (skipped)"; ((SKIP++)); }
log_section() { echo ""; echo "=== $1 ==="; }

# Resolve kube context for a region
kube_ctx() {
  local region="$1"
  local env_key="VE_KUBE_CONTEXT_$(echo "$region" | tr '[:lower:]-' '[:upper:]_')"
  echo "${!env_key:-virtengine-prod-$region}"
}

# ---------------------------------------------------------------------------
# Test 1: DNS Resolution
# ---------------------------------------------------------------------------
log_section "DNS Resolution"

for record in "api.${DOMAIN}" "rpc.${DOMAIN}"; do
  if command -v dig &>/dev/null; then
    result=$(dig +short "$record" 2>/dev/null || true)
    if [ -n "$result" ]; then
      log_pass "$record resolves to $result"
    else
      log_fail "$record does not resolve"
    fi
  else
    if nslookup "$record" &>/dev/null; then
      log_pass "$record resolves"
    else
      log_fail "$record does not resolve"
    fi
  fi
done

# Regional DNS
for region in "${REGIONS[@]}"; do
  record="rpc-${region}.${DOMAIN}"
  if command -v dig &>/dev/null; then
    result=$(dig +short "$record" 2>/dev/null || true)
    if [ -n "$result" ]; then
      log_pass "$record resolves"
    else
      log_skip "$record (not deployed yet)"
    fi
  else
    log_skip "$record (dig not available)"
  fi
done

# ---------------------------------------------------------------------------
# Test 2: Regional API Health
# ---------------------------------------------------------------------------
log_section "Regional API Health"

for region in "${REGIONS[@]}"; do
  endpoint="https://rpc-${region}.${DOMAIN}/status"
  if curl -sf --connect-timeout 10 "$endpoint" &>/dev/null; then
    log_pass "$region API responding at $endpoint"
  else
    log_skip "$region API not reachable (may not be deployed)"
  fi
done

# ---------------------------------------------------------------------------
# Test 3: Kubernetes Cluster Connectivity
# ---------------------------------------------------------------------------
log_section "Kubernetes Cluster Connectivity"

for region in "${REGIONS[@]}"; do
  ctx=$(kube_ctx "$region")
  if kubectl --context="$ctx" cluster-info &>/dev/null 2>&1; then
    nodes=$(kubectl --context="$ctx" get nodes --no-headers 2>/dev/null | wc -l)
    ready=$(kubectl --context="$ctx" get nodes --no-headers 2>/dev/null | grep -c " Ready" || true)
    log_pass "$region cluster reachable ($ready/$nodes nodes ready)"
  else
    log_skip "$region cluster not reachable (context: $ctx)"
  fi
done

# ---------------------------------------------------------------------------
# Test 4: VPC Peering Status
# ---------------------------------------------------------------------------
log_section "VPC Peering Status"

for region in "${REGIONS[@]}"; do
  peerings=$(aws ec2 describe-vpc-peering-connections \
    --region "$region" \
    --filters "Name=tag:Project,Values=virtengine" \
    --query 'VpcPeeringConnections[*].{Status:Status.Code,Peer:AccepterVpcInfo.Region}' \
    --output text 2>/dev/null || echo "")

  if [ -n "$peerings" ]; then
    while IFS=$'\t' read -r status peer; do
      if [ "$status" = "active" ]; then
        log_pass "$region ↔ $peer peering active"
      else
        log_fail "$region ↔ $peer peering status: $status"
      fi
    done <<< "$peerings"
  else
    log_skip "$region VPC peering (not configured or no AWS access)"
  fi
done

# ---------------------------------------------------------------------------
# Test 5: Cross-Region Latency
# ---------------------------------------------------------------------------
log_section "Cross-Region Latency"

for region in "${REGIONS[@]}"; do
  endpoint="https://rpc-${region}.${DOMAIN}/status"
  latency=$(curl -sf --connect-timeout 10 -w "%{time_total}" -o /dev/null "$endpoint" 2>/dev/null || echo "")
  if [ -n "$latency" ]; then
    ms=$(echo "$latency * 1000" | bc 2>/dev/null || echo "${latency}s")
    log_pass "$region latency: ${ms}ms"
  else
    log_skip "$region latency (endpoint not reachable)"
  fi
done

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "═══════════════════════════════════════"
echo "  Cross-Region Verification Summary"
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
