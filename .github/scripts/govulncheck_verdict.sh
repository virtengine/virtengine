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
# JSON mode also exits 0 when findings exist (exit 3 is text-mode only), so when
# a text-mode run is supplied its status carries the verdict; any other nonzero
# status is a scanner error and must never be reported as a vulnerability.
#
# Scanning the whole module twice is what made this job unrunnable: govulncheck
# ./... over this monorepo takes longer than a hosted runner will stay alive, and
# the second (text-mode) pass was pure duplication. Pass an empty or "-" text
# status to derive the verdict from the JSON report alone, which already carries
# one `finding` object per REACHABLE vulnerability. The fixture regression that
# motivated this script is unaffected: a config-only report has no `finding`
# object and is still classified `error`.
#
# Usage: govulncheck_verdict.sh <report.json> <json_exit> [text_exit]
#   Prints `verdict=clean|vulns|error` on stdout; the reason goes to stderr.
#   Always exits 0 -- the caller decides how to fail the job.
set -uo pipefail

report="${1:-}"
json_status="${2:-0}"
text_status="${3:-0}"

# "-" (or omitted) means "no second text-mode scan was run".
derive_from_json=0
if [ -z "$text_status" ] || [ "$text_status" = "-" ]; then
  derive_from_json=1
  text_status=0
fi

# A reachable vulnerability is emitted as a `finding` object. govulncheck writes
# a pretty-printed stream of concatenated JSON objects, so in a real report the
# key sits alone on its own line; a compact one-line-per-message stream puts it
# mid-line. Matching the key immediately followed by a colon covers both without
# reading a quoted value that merely contains the word "finding", because such a
# value is followed by the value's own characters, not a colon.
count_findings() {
  grep -oE '"finding"[[:space:]]*:' "$1" 2>/dev/null | wc -l | tr -d ' '
}

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
elif [ "$derive_from_json" -eq 1 ]; then
  n="$(count_findings "$report")"
  n="${n:-0}"
  if [ "$n" -gt 0 ]; then
    emit vulns "govulncheck JSON report contains $n reachable finding(s)"
  elif [ "$json_status" -ne 0 ]; then
    emit error "govulncheck did not complete (json exit $json_status)"
  else
    emit clean "no reachable vulnerabilities"
  fi
elif [ "$json_status" -ne 0 ] || [ "$text_status" -ne 0 ]; then
  emit error "govulncheck did not complete (json exit $json_status, text exit $text_status)"
else
  emit clean "no reachable vulnerabilities"
fi

exit 0
