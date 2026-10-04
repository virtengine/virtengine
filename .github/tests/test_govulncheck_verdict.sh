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

# The REAL report shape: govulncheck writes a pretty-printed stream of
# concatenated JSON objects, so `"finding"` is a key on its own line at two
# spaces of indent, NOT a one-line-per-message stream. The single-scan mode
# must find findings in this shape. Trimmed from a real run against
# ./pkg/verification/... on Go 1.26.8, which reported 4 findings.
cat > "$TMP/real-vulns.json" <<'EOF'
{
  "config": {
    "scanner_name": "govulncheck"
  }
}
{
  "SBOM": {
    "data": "..."
  }
}
{
  "progress": {
    "message": "Scanning your code"
  }
}
{
  "osv": {
    "id": "GO-2022-0635",
    "summary": "Denial of service in golang.org/x/crypto/ssh"
  }
}
{
  "finding": {
    "osv": "GO-2022-0635",
    "fixed_by": "v0.0.0-20220315160706-3147a52a75dd",
    "trace": [
      {
        "module": "golang.org/x/crypto"
      }
    ]
  }
}
{
  "finding": {
    "osv": "GO-2026-5841",
    "trace": []
  }
}
EOF

# A clean real-shaped report: the scan ran and found nothing reachable.
cat > "$TMP/real-clean.json" <<'EOF'
{
  "config": {
    "scanner_name": "govulncheck"
  }
}
{
  "SBOM": {
    "data": "..."
  }
}
{
  "progress": {
    "message": "Scanning your code"
  }
}
{
  "osv": {
    "id": "GO-2024-2887",
    "summary": "A module in the dependency graph has a known issue"
  }
}
EOF

# An OSV entry whose summary text contains the word "finding" must NOT be
# counted as a result: the word appears inside a quoted value, not as a key.
cat > "$TMP/osv-mentions-finding.json" <<'EOF'
{
  "config": {
    "scanner_name": "govulncheck"
  }
}
{
  "progress": {
    "message": "Scanning your code"
  }
}
{
  "osv": {
    "id": "GO-2026-9999",
    "summary": "this module has a finding in its parser"
  }
}
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

# The single-scan mode: text status is "-" (no second govulncheck pass), so the
# verdict must come from the JSON report. This is the mode the job now uses.
echo "single-scan mode (text exit '-')"
assert_verdict "real report w/ findings"  vulns "$TMP/real-vulns.json" 0 -
assert_verdict "real report clean"       clean "$TMP/real-clean.json" 0 -
assert_verdict "osv word 'finding'"      clean "$TMP/osv-mentions-finding.json" 0 -
assert_verdict "config-only still error" error "$TMP/config-only.json" 0 -
assert_verdict "empty still error"       error "$TMP/empty.json" 0 -
assert_verdict "json nonzero = error"    error "$TMP/real-clean.json" 1 -
assert_verdict "absent file = error"     error "$TMP/does-not-exist.json" 0 -

# The word "finding" inside a quoted value must not survive the tightened grep,
# but the legacy one-line fixture (a genuine finding key) still must.
echo "single-scan mode on the legacy one-line fixtures"
assert_verdict "one-line finding key"  vulns "$TMP/vulns.json" 0 -
assert_verdict "one-line progress key" clean "$TMP/clean.json" 0 -

echo
echo "RESULT: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
