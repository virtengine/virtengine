#!/usr/bin/env bash
# Guard: the sharded govulncheck gate must not combine -show with -format json.
#
# govulncheck v1.1.4 rejects that combination: it prints
#   "the -show flag is not supported for json output"
# and writes NOTHING to stdout, so govulncheck-report.json is empty and every
# shard classifies as verdict=error ("no JSON report was written"). That is how
# all four shards went red the moment the gate was sharded (#1053), and #1058
# removed the flag.
#
# #1058 recorded the reason in a WORKFLOW COMMENT. A comment is not a guard: the
# next person to "restore progress reporting" re-adds the flag and the only
# signal is a 35-second red on all four shards. This asserts it instead.
#
# Behavioural where it can be (against the real binary when one is on PATH) and
# source-level otherwise, so it fails in either direction: re-adding -show, or
# the gate's invocation being renamed away from under it.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="$ROOT/.github/workflows/security.yaml"

fail=0
ok()  { echo "  PASS  $*"; }
bad() { echo "  FAIL  $*"; fail=1; }

echo "== source: the sharded govulncheck invocation =="

if [ ! -f "$WORKFLOW" ]; then
  bad "workflow not found: $WORKFLOW"
  exit 1
fi

# Lines that actually INVOKE govulncheck with json output. Comment lines are
# excluded: the workflow comment that warns about -show must not itself be read
# as an invocation (that false positive fired on this guard's first cut).
mapfile -t json_lines < <(grep -nE 'govulncheck[[:space:]].*-format[[:space:]]+json' "$WORKFLOW" \
  | grep -vE ':[[:space:]]*#')

if [ "${#json_lines[@]}" -eq 0 ]; then
  bad "no 'govulncheck ... -format json' invocation found -- did the gate get renamed?"
else
  for line in "${json_lines[@]}"; do
    if [[ "$line" == *"-show"* ]]; then
      bad "-show is combined with -format json (json output refuses it and writes nothing): $line"
    else
      ok "json invocation carries no -show: $(echo "$line" | cut -c1-95)"
    fi
  done
fi

echo
echo "== behaviour: the real binary (skipped when govulncheck is absent) =="

GV="$(command -v govulncheck || true)"
if [ -z "$GV" ]; then
  echo "  SKIP  govulncheck not on PATH -- the source check above is the guard here"
else
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  # A one-package module so the probe is seconds, not minutes.
  ( cd "$tmp" && printf 'module probe\n\ngo 1.21\n' > go.mod && mkdir -p p && \
    printf 'package p\n\nfunc F() int { return 1 }\n' > p/p.go )

  ( cd "$tmp" && "$GV" -show=verbose -format json ./p/... > bad.out 2>bad.err )
  if grep -q 'the -show flag is not supported for json output' "$tmp/bad.out" 2>/dev/null \
     || grep -q 'the -show flag is not supported for json output' "$tmp/bad.err" 2>/dev/null; then
    ok "binary still refuses -show with json -- the guard's premise holds"
  else
    echo "  NOTE  this govulncheck no longer refuses -show with json; re-check the"
    echo "        premise before trusting this guard"
  fi

  ( cd "$tmp" && "$GV" -format json ./p/... > good.json 2>good.err )
  good_bytes=$(wc -c < "$tmp/good.json")
  if [ "$good_bytes" -gt 0 ]; then
    ok "json alone writes a report ($good_bytes bytes)"
  else
    bad "json alone wrote an EMPTY report -- a shard would classify as verdict=error"
  fi
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "RESULT: govulncheck invocation guard PASSED"
else
  echo "RESULT: govulncheck invocation guard FAILED"
fi
exit "$fail"
