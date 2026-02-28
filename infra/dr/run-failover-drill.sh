#!/usr/bin/env bash
set -euo pipefail

MODE="rehearsal"
OUTPUT_DIR=""
RUN_LIVE_VALIDATION="false"

usage() {
  cat <<'EOF'
Usage: run-failover-drill.sh [--mode rehearsal|live] [--output-dir DIR] [--live-validation]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="$2"
      shift 2
      ;;
    --output-dir)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --live-validation)
      RUN_LIVE_VALIDATION="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "$MODE" != "rehearsal" && "$MODE" != "live" ]]; then
  echo "[dr-drill] ERROR: unsupported mode: $MODE" >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-$ROOT_DIR/.dr-drill-output}"
mkdir -p "$OUTPUT_DIR"

LOG_FILE="$OUTPUT_DIR/failover-drill.log"
SUMMARY_FILE="$OUTPUT_DIR/failover-drill-summary.md"
EVIDENCE_FILE="$OUTPUT_DIR/failover-drill-evidence.json"

STARTED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
START_EPOCH="$(date -u +%s)"
GIT_SHA="$(git -C "$ROOT_DIR" rev-parse HEAD)"
STATUS="success"
LIVE_STATUS="not-run"

run_rehearsal() {
  bash "$ROOT_DIR/scripts/dr/failover-test.sh" 2>&1 | tee "$LOG_FILE"
}

run_live_validation() {
  local test_filter='Test(RegionalFailover_|DatabaseReplication_|Observability_)'
  set +e
  go test ./infra/dr/tests -count=1 -run "$test_filter" 2>&1 | tee -a "$LOG_FILE"
  local rc=${PIPESTATUS[0]}
  set -e
  if [[ $rc -eq 0 ]]; then
    LIVE_STATUS="passed"
  else
    LIVE_STATUS="failed"
    return "$rc"
  fi
}

if ! run_rehearsal; then
  STATUS="failed"
fi

if [[ "$STATUS" == "success" && "$MODE" == "live" && "$RUN_LIVE_VALIDATION" == "true" ]]; then
  if ! run_live_validation; then
    STATUS="failed"
  fi
fi

ENDED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
END_EPOCH="$(date -u +%s)"
DURATION_SECONDS="$((END_EPOCH - START_EPOCH))"
LOG_SHA256="$(sha256sum "$LOG_FILE" | cut -d' ' -f1)"

cat > "$SUMMARY_FILE" <<EOF
# DR Failover Drill Summary

- Mode: $MODE
- Status: $STATUS
- Started: $STARTED_AT
- Ended: $ENDED_AT
- Duration Seconds: $DURATION_SECONDS
- Git SHA: $GIT_SHA
- Log SHA256: $LOG_SHA256
- Live Validation: $LIVE_STATUS
EOF

cat > "$EVIDENCE_FILE" <<EOF
{
  "mode": "$MODE",
  "status": "$STATUS",
  "started_at": "$STARTED_AT",
  "ended_at": "$ENDED_AT",
  "duration_seconds": $DURATION_SECONDS,
  "git_sha": "$GIT_SHA",
  "log_sha256": "$LOG_SHA256",
  "live_validation": "$LIVE_STATUS"
}
EOF

if [[ "$STATUS" != "success" ]]; then
  exit 1
fi
