#!/usr/bin/env bash
# Tests for .github/scripts/govulncheck_verdict.sh
#
# The regression these lock down: a scanner that never loaded a single package
# (a ~290-byte config-only JSON report) was reported to reviewers as
# "govulncheck reported vulnerabilities". A security control must be able to
# distinguish "we are clean" from "the scanner did not run"; conflating them
# manufactures false findings and hides true ones behind the same message.
#
# Run: bash .github/tests/test_govulncheck_verdict.sh
set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$HERE/../scripts/govulncheck_verdict.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

PASS=0
FAIL=0

# assert_verdict <name> <expected> <report> <json_exit> <text_exit>
assert_verdict() {
  local name="$1" expected="$2" report="$3" json_exit="$4" text_exit="$5"
  local actual
  actual="$(bash "$SCRIPT" "$report" "$json_exit" "$text_exit" 2>/dev/null | sed -n 's/^verdict=//p')"
  if [ "$actual" = "$expected" ]; then
    echo "  PASS  $name -> $actual"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $name -> got '$actual', want '$expected'"
    FAIL=$((FAIL + 1))
  fi
}

# --- fixtures ---------------------------------------------------------------

# Exactly what `govulncheck -format json ./...` writes when package loading
# fails: the config object and nothing else. Byte-for-byte shape taken from the
# real artifact uploaded by the Go Vulnerability Scan job on main.
cat > "$TMP/config-only.json" <<'EOF'
{
  "config": {
    "protocol_version": "v1.0.0",
    "scanner_name": "govulncheck",
    "scanner_version": "v1.1.4",
    "db": "https://vuln.go.dev",
    "db_last_modified": "2026-09-16T18:00:43Z",
    "go_version": "go1.25.14",
    "scan_level": "symbol",
    "scan_mode": "source"
  }
}
EOF

# A scan that loaded packages and found nothing still emits progress messages.
cat > "$TMP/clean.json" <<'EOF'
{"config":{"scanner_name":"govulncheck"}}
{"progress":{"message":"Scanning your code"}}
EOF

# A scan that loaded packages and found a reachable vulnerability.
cat > "$TMP/vulns.json" <<'EOF'
{"config":{"scanner_name":"govulncheck"}}
{"osv":{"id":"GO-2026-6355","summary":"Prevent DoS"}}
{"finding":{"osv":"GO-2026-6355"}}
EOF

: > "$TMP/empty.json"

# --- cases ------------------------------------------------------------------

echo "config-only report (scanner never loaded packages)"
assert_verdict "text exit 1"   error "$TMP/config-only.json" 1 1
assert_verdict "text exit 3"   error "$TMP/config-only.json" 1 3
assert_verdict "both exit 0"   error "$TMP/config-only.json" 0 0

echo "clean scan"
assert_verdict "exit 0/0"      clean "$TMP/clean.json" 0 0

echo "scan with findings"
assert_verdict "text exit 3"   vulns "$TMP/vulns.json" 0 3
assert_verdict "json nonzero"  vulns "$TMP/vulns.json" 1 3

echo "scanner error that is not a load failure"
assert_verdict "text exit 1"   error "$TMP/vulns.json" 0 1
assert_verdict "json exit 1"   error "$TMP/vulns.json" 1 0

echo "missing / empty report"
assert_verdict "empty file"    error "$TMP/empty.json" 1 1
assert_verdict "absent file"   error "$TMP/does-not-exist.json" 1 1

echo
echo "RESULT: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
