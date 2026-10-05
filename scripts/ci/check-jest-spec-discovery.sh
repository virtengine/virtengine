#!/usr/bin/env bash
# t_631af860 -- assert sdk/ts jest spec discovery does not collapse to zero.
#
# WHY THIS EXISTS
#   jest interpolated an ABSOLUTE `<rootDir>` prefix into testMatch. On a Windows
#   checkout path containing a glob-special segment (e.g. `virtengine/.worktrees/<name>/`)
#   the interpolated pattern ends up MIXED-separator and matches nothing, so
#   `--listTests` reported 0 of 40 specs while 40 spec files sat on disk. Two agents
#   measured this suite wrong from such a path.
#
#   This guard asserts non-zero discovery for the tree it runs in. To cover the
#   path-dependent shape as well, pass extra checkout roots explicitly:
#       scripts/ci/check-jest-spec-discovery.sh
#       scripts/ci/check-jest-spec-discovery.sh C:/.../virtengine/.worktrees/t_foo
#   A named probe that is absent, or has no node_modules, is reported UNPROBED and
#   FAILS the run: an unprobed shape is not a passing shape.
#
#   Root cause, read from the installed tree (jest 30.5.2):
#     jest-config/build/index.js:1887-1899  case 'testMatch'
#     jest-config/build/index.js:2845      escapeGlobCharacters   (root escaped FIRST)
#     jest-config/build/index.js:2848      replaceRootDirInPath   (native separators)
#     jest-config/build/index.js:2862      _replaceRootDirTags
#     jest-util/build/index.js:1009        replacePathSepForGlob
#         return path.replaceAll(/\\(?![$()+.?^{}])/g, '/');
#   Escaping happens BEFORE the separator rewrite, so in `...virtengine\.worktrees/...`
#   the backslash now precedes a glob escape `.`, the negative lookahead can never
#   fire, the separator survives, and the pattern is mixed -- matching no crawled path.
#
# EXIT CODES
#   0  every probed tree discovers at least one spec
#   1  a probed tree discovered ZERO specs, or a named probe was unprobeable
#       (this is the regression the guard exists to catch)
#   2  precondition unmet (node_modules absent for this tree)
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIN_SPECS=1

log() { printf '%s\n' "$*"; }

count_specs() {  # $1 = sdk/ts dir -> echoes discovered unit-spec count (exactly one line)
  # grep -c prints 0 AND exits non-zero when there are no matches; the pipeline below
  # must still emit a single bare integer, or a caller doing `[ "$n" -eq 0 ]` gets
  # "integer expression expected" instead of a verdict.
  local n
  n="$( ( cd "$1" 2>/dev/null && npx --no-install jest --selectProjects unit --listTests 2>/dev/null ) \
        | grep -c '\.spec\.ts$' )" || n=0
  case "$n" in
    ''|*[!0-9]*) n=0 ;;
  esac
  printf '%s' "$n"
}

# ------------------------------------------------------------------ selftest
# A guard that cannot fail is decoration. Drive the assertion over a tree with no
# jest at all: it must resolve to ZERO. Without this row, a green run is no
# evidence that the check bites.
if [ "${1:-}" = "--selftest" ]; then
  log "selftest: proving the zero-discovery assertion is reachable"

  probe="$(mktemp -d 2>/dev/null)" || { log "selftest FAIL: mktemp unavailable"; exit 2; }
  mkdir -p "$probe/sdk/ts/src"
  printf 'describe("x", () => { it("y", () => {}); });\n' > "$probe/sdk/ts/src/a.spec.ts"

  n="$(count_specs "$probe/sdk/ts" 2>/dev/null || echo 0)"
  rm -rf "$probe"
  if [ "$n" -ne 0 ]; then
    log "selftest FAIL: the RED fixture discovered $n specs; it must discover 0"
    exit 1
  fi
  log "selftest GREEN: a tree with no jest resolves to 0 specs (the RED condition)"

  # The real tree must be discoverable, or the selftest proves nothing.
  if [ ! -d "$REPO_ROOT/sdk/ts/node_modules" ]; then
    log "selftest BLOCKED: $REPO_ROOT/sdk/ts/node_modules is absent"
    exit 2
  fi
  real="$(count_specs "$REPO_ROOT/sdk/ts")"
  if [ "${real:-0}" -lt "$MIN_SPECS" ]; then
    log "selftest FAIL: the real tree discovers ${real:-0} specs; the guard would be red"
    exit 1
  fi
  log "selftest GREEN: the real tree discovers $real specs (must be > 0)"
  exit 0
fi

# ----------------------------------------------------------------- assertion
if [ ! -d "$REPO_ROOT/sdk/ts/node_modules" ]; then
  log "BLOCKED: $REPO_ROOT/sdk/ts/node_modules is absent; run the sdk/ts install first"
  exit 2
fi

rc=0

n="$(count_specs "$REPO_ROOT/sdk/ts")"
if [ "${n:-0}" -ge "$MIN_SPECS" ]; then
  log "PASS  this checkout: $n unit specs discovered"
else
  log "FAIL  this checkout: ${n:-0} unit specs discovered (>= $MIN_SPECS required)"
  rc=1
fi

# Additional checkout roots, e.g. a .worktrees path (each must itself contain sdk/ts).
for arg in "$@"; do
  root="${arg%/}"
  dir="$root/sdk/ts"
  [ "$dir" = "$REPO_ROOT/sdk/ts" ] && continue
  if [ ! -d "$dir" ]; then
    log "UNPROBED $root -- no sdk/ts there; an unprobed shape is not a passing shape"
    rc=1
    continue
  fi
  if [ ! -d "$dir/node_modules" ]; then
    log "UNPROBED $root -- node_modules absent; install before trusting this shape"
    rc=1
    continue
  fi
  n="$(count_specs "$dir")"
  if [ "${n:-0}" -ge "$MIN_SPECS" ]; then
    log "PASS  $root: $n unit specs discovered"
  else
    log "FAIL  $root: ${n:-0} unit specs discovered (>= $MIN_SPECS required)"
    rc=1
  fi
done

if [ "$rc" -eq 0 ]; then
  log "GREEN: jest spec discovery is non-zero from every probed checkout-path shape"
else
  log "RED: see the FAIL/UNPROBED rows above. A zero here means the suite was measured"
  log "     as EMPTY, not passed -- see sdk/ts/jest.config.ts (t_631af860)."
fi
exit "$rc"
