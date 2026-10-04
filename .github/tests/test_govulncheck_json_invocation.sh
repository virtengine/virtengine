#!/usr/bin/env bash
# Guard: the sharded govulncheck gate must not combine -show with -format json.
#
# govulncheck v1.1.4 rejects that combination: it prints
#   "the -show flag is not supported for json output"
# and writes NOTHING to stdout, so govulncheck-report.json is empty and every
# shard classifies as verdict=error ("no JSON report was written"). That is how
# all four shards went red the moment the gate was sharded -- the wrapper's
# classification was correct, the invocation was impossible.
#
# This is a behavioural check against the real binary when one is available, plus
# a source check that pins the invocation, so it fails in either direction:
# re-adding -show, or the flag becoming mandatory.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="$ROOT/.github/workflows/security.yaml"

fail=0
ok()   { echo "  PASS  $*"; }
bad()  { echo "  FAIL  $*"; fail=1; }

echo "== source: the sharded govulncheck invocation =="

if [ ! -f "$WORKFLOW" ]; then
  bad "workflow not found: $WORKFLOW"
  exit 1
fi

# Every govulncheck line that writes a JSON report. Only COMMAND lines are
# considered: a line whose first non-space character is `#` is prose, and a
# comment that WARNS about -show must not be read as an invocation of it.
mapfile -t json_lines < <(grep -nE 'govulncheck[[:space:]].*-format[[:space:]]+json' "$WORKFLOW" \
  | grep -vE ':[[:space:]]*#')

if [ "${#json_lines[@]}" -eq 0 ]; then
  bad "no 'govulncheck ... -format json' invocation found -- did the gate get renamed?"
else
  for line in "${json_lines[@]}"; do
    if [[ "$line" == *"-show"* ]]; then
      bad "-show is combined with -format json (json output ignores it and writes nothing): $line"
    else
      ok "json invocation carries no -show: $(echo "$line" | cut -c1-90)"
    fi
  done
fi

echo
echo "== behaviour: the real binary (skipped when govulncheck is absent) =="

GV="$(command -v govulncheck || true)"
if [ -z "$GV" ]; then
  echo "  SKIP  govulncheck not on PATH -- source check above is the guard here"
else
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  # A tiny package so the probe is fast; /tmp is fine because govulncheck reads
  # only GOFLAGS/module state for this shape.
  ( cd "$tmp" && printf 'module probe\n\ngo 1.21\n' > go.mod && mkdir -p p && \
    printf 'package p\n\nfunc F() int { return 1 }\n' > p/p.go )

  ( cd "$tmp" && "$GV" -show=verbose -format json ./p/... > bad.json 2>bad.err )
  bad_rc=$?
  bad_bytes=$(wc -c < "$tmp/bad.json")

  if grep -q 'the -show flag is not supported for json output' "$tmp/bad.err" \
     || grep -q 'the -show flag is not supported for json output' "$tmp/bad.json"; then
    ok "binary still rejects -show with json (the reason for this guard)"
  else
    echo "  NOTE  this govulncheck no longer rejects -show with json; the guard's"
    echo "        premise changed -- re-check before relying on it (rc=$bad_rc, bytes=$bad_bytes)"
  fi

  ( cd "$tmp" && "$GV" -format json ./p/... > good.json 2>good.err )
  good_bytes=$(wc -c < "$tmp/good.json")
  if [ "$good_bytes" -gt 0 ]; then
    ok "json alone writes a report ($good_bytes bytes)"
  else
    bad "json alone wrote an empty report -- the shard would classify as error"
  fi
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "RESULT: govulncheck invocation guard PASSED"
else
  echo "RESULT: govulncheck invocation guard FAILED"
fi
exit "$fail"
