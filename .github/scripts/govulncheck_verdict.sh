#!/usr/bin/env bash
# Classify a govulncheck run as clean / vulnerable / scanner-error.
#
# Why this exists
# ---------------
# `govulncheck -format json` ALWAYS writes the config object first, even when
# package loading fails, so a non-empty report does not prove the scan ran.
# When a cgo dependency is unavailable (libudev, pulled in through
# zondax/troian hid) the report is a ~290-byte config-only document.
#
# The previous gate treated ANY nonzero exit as "govulncheck reported
# vulnerabilities", so a scanner that never loaded a single package was
# reported to reviewers as a security finding. A control that cannot tell
# "we are clean" apart from "the scanner did not run" is worse than no control:
# it manufactures false findings and hides true ones behind the same message.
#
# JSON mode also exits 0 when findings exist (exit 3 is text-mode only), so the
# text-mode status carries the verdict; any other nonzero status is a scanner
# error and must never be reported as a vulnerability.
#
# Usage: govulncheck_verdict.sh <report.json> <json_exit> <text_exit>
#   Prints `verdict=clean|vulns|error` on stdout; the reason goes to stderr.
#   Always exits 0 -- the caller decides how to fail the job.
set -uo pipefail

report="${1:-}"
json_status="${2:-0}"
text_status="${3:-0}"

if [ -z "$report" ]; then
  echo "usage: $0 <report.json> <json_exit> <text_exit>" >&2
  exit 2
fi

emit() {
  echo "verdict=$1"
  echo "reason=$2" >&2
}

if [ ! -s "$report" ]; then
  emit error "no JSON report was written"
  exit 0
fi

# A real scan emits at least one osv/finding/progress message. Their absence
# means package loading failed and the report is the config object alone.
if ! grep -qE '"(osv|finding|progress)"' "$report"; then
  emit error "could not load packages: JSON report contains no scan results (config only)"
  exit 0
fi

if [ "$text_status" -eq 3 ]; then
  emit vulns "govulncheck reported reachable vulnerabilities"
elif [ "$json_status" -ne 0 ] || [ "$text_status" -ne 0 ]; then
  emit error "govulncheck did not complete (json exit $json_status, text exit $text_status)"
else
  emit clean "no reachable vulnerabilities"
fi

exit 0
