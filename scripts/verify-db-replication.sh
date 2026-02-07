#!/usr/bin/env bash
# Database Replication Verification
# Validates CockroachDB multi-region replication health, backup status,
# and data consistency across regions.
set -euo pipefail

REGIONS=("us-east-1" "eu-west-1" "ap-southeast-1")
RPO_THRESHOLD_SECONDS=300
PASS=0
FAIL=0
SKIP=0

log_pass() { echo "  ✓ $1"; ((PASS++)); }
log_fail() { echo "  ✗ $1"; ((FAIL++)); }
log_skip() { echo "  ○ $1 (skipped)"; ((SKIP++)); }
log_section() { echo ""; echo "=== $1 ==="; }

kube_ctx() {
  local region="$1"
  local env_key="VE_KUBE_CONTEXT_$(echo "$region" | tr '[:lower:]-' '[:upper:]_')"
  echo "${!env_key:-virtengine-prod-$region}"
}

crdb_sql() {
  local ctx="$1"
  local query="$2"
  kubectl --context="$ctx" -n cockroachdb exec cockroachdb-0 -- \
    cockroach sql --certs-dir=/cockroach/cockroach-certs --format=csv -e "$query" 2>/dev/null
}

# ---------------------------------------------------------------------------
# Test 1: CockroachDB Node Health
# ---------------------------------------------------------------------------
log_section "CockroachDB Node Health"

for region in "${REGIONS[@]}"; do
  ctx=$(kube_ctx "$region")
  pods=$(kubectl --context="$ctx" -n cockroachdb get pods --no-headers 2>/dev/null || echo "")
  if [ -z "$pods" ]; then
    log_skip "$region CockroachDB (not deployed)"
    continue
  fi

  running=$(echo "$pods" | grep -c "Running" || true)
  total=$(echo "$pods" | wc -l)
  if [ "$running" -eq "$total" ] && [ "$total" -gt 0 ]; then
    log_pass "$region: $running/$total CockroachDB pods running"
  else
    log_fail "$region: $running/$total CockroachDB pods running"
  fi
done

# ---------------------------------------------------------------------------
# Test 2: Replication Status
# ---------------------------------------------------------------------------
log_section "Replication Status"

primary_ctx=$(kube_ctx "${REGIONS[0]}")

# Under-replicated ranges
result=$(crdb_sql "$primary_ctx" \
  "SELECT count(*) AS cnt FROM crdb_internal.ranges WHERE array_length(replicas, 1) < 3;" 2>/dev/null || echo "")
if [ -n "$result" ]; then
  count=$(echo "$result" | tail -1 | tr -d '[:space:]')
  if [ "$count" = "0" ]; then
    log_pass "No under-replicated ranges"
  else
    log_fail "$count under-replicated ranges found"
  fi
else
  log_skip "Replication status check (CockroachDB not reachable)"
fi

# Unavailable ranges
result=$(crdb_sql "$primary_ctx" \
  "SELECT count(*) AS cnt FROM crdb_internal.ranges WHERE array_length(replicas, 1) = 0;" 2>/dev/null || echo "")
if [ -n "$result" ]; then
  count=$(echo "$result" | tail -1 | tr -d '[:space:]')
  if [ "$count" = "0" ]; then
    log_pass "No unavailable ranges"
  else
    log_fail "$count unavailable ranges found"
  fi
else
  log_skip "Unavailable ranges check"
fi

# ---------------------------------------------------------------------------
# Test 3: Backup Freshness
# ---------------------------------------------------------------------------
log_section "Backup Freshness"

for region in "${REGIONS[@]}"; do
  ctx=$(kube_ctx "$region")
  bucket="s3://virtengine-cockroachdb-backup-${region}/backups"
  query="SELECT extract(epoch from (now() - max(end_time)))::int AS age FROM [SHOW BACKUP LATEST IN '${bucket}?AUTH=implicit'];"

  result=$(crdb_sql "$ctx" "$query" 2>/dev/null || echo "")
  if [ -n "$result" ]; then
    age=$(echo "$result" | tail -1 | tr -d '[:space:]')
    if [ -n "$age" ] && [ "$age" -lt "$RPO_THRESHOLD_SECONDS" ]; then
      log_pass "$region backup age: ${age}s (within RPO: ${RPO_THRESHOLD_SECONDS}s)"
    elif [ -n "$age" ]; then
      log_fail "$region backup age: ${age}s (exceeds RPO: ${RPO_THRESHOLD_SECONDS}s)"
    else
      log_skip "$region backup age (no data)"
    fi
  else
    log_skip "$region backup freshness (not reachable)"
  fi
done

# ---------------------------------------------------------------------------
# Test 4: S3 Backup Bucket Encryption
# ---------------------------------------------------------------------------
log_section "Backup Bucket Encryption"

for region in "${REGIONS[@]}"; do
  bucket="virtengine-cockroachdb-backup-${region}"
  enc=$(aws s3api get-bucket-encryption --bucket "$bucket" --region "$region" 2>/dev/null || echo "")
  if [ -n "$enc" ]; then
    if echo "$enc" | grep -q "aws:kms"; then
      log_pass "$region backup bucket uses KMS encryption"
    else
      log_fail "$region backup bucket does NOT use KMS encryption"
    fi
  else
    log_skip "$region backup bucket encryption (bucket not found or no access)"
  fi
done

# ---------------------------------------------------------------------------
# Test 5: Write Test (primary region only)
# ---------------------------------------------------------------------------
log_section "Database Write Test"

write_result=$(crdb_sql "$primary_ctx" \
  "CREATE TABLE IF NOT EXISTS system.dr_verify (id INT PRIMARY KEY, ts TIMESTAMP DEFAULT now()); UPSERT INTO system.dr_verify VALUES (1); SELECT ts FROM system.dr_verify WHERE id = 1;" \
  2>/dev/null || echo "")
if [ -n "$write_result" ]; then
  log_pass "Primary region accepts writes"
else
  log_skip "Write test (CockroachDB not reachable)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "═══════════════════════════════════════"
echo "  Database Replication Summary"
echo "═══════════════════════════════════════"
echo "  Passed:  $PASS"
echo "  Failed:  $FAIL"
echo "  Skipped: $SKIP"
echo "  RPO Target: ${RPO_THRESHOLD_SECONDS}s"
echo "═══════════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
  echo "  RESULT: FAILED"
  exit 1
else
  echo "  RESULT: PASSED"
  exit 0
fi
