#!/usr/bin/env bash
# Falsify the GHCR pull-credential guard in BOTH directions on the real tree.
#
# Direction 1 (negative): remove imagePullSecrets from each of the four
# deploy/kubernetes ServiceAccounts in turn; the guard MUST fail.
# Direction 2 (positive): the unmodified tree MUST pass.
# Direction 3: the placeholder annotation must fail when the placeholder digest
# is swapped for a real one, and must fail when removed while the digest stays.
#
# Restore is by FILE SNAPSHOT, not `git checkout`: this script has to be
# runnable against a pinned `git archive` export (which is how the guard gets
# verified against an exact SHA rather than a moving shared tree). `git
# checkout` fails silently in such an export -- the old 2>/dev/null hid it --
# so every mutation would compound onto the previous one and the closing
# "tree restored" run would prove nothing. Here every restore is byte-compared
# against the snapshot and a mismatch aborts non-zero.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TESTS="$REPO/.github/tests"
cd "$TESTS" || exit 1

# `python` here is NATIVE Windows Python, which cannot open the MSYS-style path
# bash's `pwd` prints (/c/Users/...). Shell tools are happy with /c/..., so the
# two get different forms of the same directory: REPO for grep/cp/cmp,
# REPO_PY for every path handed to python. On Linux cygpath is absent and the
# two are identical, so nothing here is Windows-specific behaviour.
if command -v cygpath >/dev/null 2>&1; then
  REPO_PY="$(cygpath -m "$REPO")"
else
  REPO_PY="$REPO"
fi

fail=0

# Every file this script is allowed to mutate.
FILES=(
  "deploy/kubernetes/base/virtengine-node-deployment.yaml"
  "deploy/kubernetes/base/provider-daemon-deployment.yaml"
  "deploy/kubernetes/base/tee-enclave-deployment.yaml"
  "deploy/kubernetes/base/veid-inference-deployment.yaml"
  "infra/kubernetes/dr/backup-cronjobs.yaml"
)

SNAP="$(mktemp -d)" || exit 1
trap 'rm -rf "$SNAP"' EXIT

snapshot() {
  local f
  for f in "${FILES[@]}"; do
    if [ ! -f "$REPO/$f" ]; then
      echo "missing expected manifest: $f"
      exit 1
    fi
    mkdir -p "$SNAP/$(dirname "$f")"
    cp "$REPO/$f" "$SNAP/$f"
  done
}

restore() {
  local f
  for f in "${FILES[@]}"; do
    cp "$SNAP/$f" "$REPO/$f" || { echo "restore failed: $f"; exit 1; }
    cmp -s "$SNAP/$f" "$REPO/$f" || { echo "restore MISMATCH: $f"; exit 1; }
  done
}

run_guard() {
  python -m unittest test_ghcr_image_pull_credential_policy 2>&1
}

report() { # name expected(FAIL|PASS) output
  local name="$1" want="$2" out="$3" got
  got="$(printf '%s\n' "$out" | grep -oE '^(OK|FAILED)' | head -1)"
  case "$got" in
    OK) got=PASS ;;
    FAILED*) got=FAIL ;;
  esac
  if [ "$got" = "$want" ]; then
    echo "PASS  $name -> $got (expected $want)"
  else
    echo "FAIL  $name -> ${got:-<no verdict line>} (expected $want)"
    printf '%s\n' "$out" | tail -25
    fail=1
  fi
}

snapshot
echo "pinned manifest digests under test:"
( cd "$SNAP" && sha256sum "${FILES[@]}" ) | sed 's/^/  /'

echo "=============================================================="
echo "Direction 2 (positive): unmodified tree must PASS"
echo "=============================================================="
restore
out="$(run_guard)"
report "unmodified tree" PASS "$out"

for sa_file in "${FILES[@]}"; do
  echo
  echo "=============================================================="
  echo "Direction 1 (negative): strip imagePullSecrets from $sa_file"
  echo "=============================================================="
  restore
  before="$(grep -c 'imagePullSecrets:' "$REPO/$sa_file")"
  # Delete the whole key plus its list items, wherever it sits.
  python - "$REPO_PY/$sa_file" <<'PY'
import re, sys
p = sys.argv[1]
text = open(p, encoding="utf-8").read()
new = re.sub(r"(?m)^imagePullSecrets:\n(?:[ \t]+-[^\n]*\n)+", "", text)
assert new != text, "no imagePullSecrets block removed"
open(p, "w", encoding="utf-8", newline="").write(new)
PY
  after="$(grep -c 'imagePullSecrets:' "$REPO/$sa_file" 2>/dev/null || echo 0)"
  echo "imagePullSecrets occurrences: $before -> $after"
  out="$(run_guard)"
  report "stripped $sa_file" FAIL "$out"
  echo "$out" | grep -oE 'AssertionError:.*' | head -2 | cut -c1-220
done

echo
echo "=============================================================="
echo "Direction 3a: placeholder digest replaced by a real one,"
echo "            placeholder annotation left behind -> must FAIL"
echo "=============================================================="
restore
python - "$REPO_PY/deploy/kubernetes/base/veid-inference-deployment.yaml" <<'PY'
import sys
p = sys.argv[1]
ZERO = "0" * 64
REAL = "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"
t = open(p, encoding="utf-8").read()
assert ZERO in t
open(p, "w", encoding="utf-8", newline="").write(t.replace(ZERO, REAL))
PY
out="$(run_guard)"
report "stale placeholder annotation" FAIL "$out"
echo "$out" | grep -oE 'is annotated.*' | head -1 | cut -c1-220

echo
echo "=============================================================="
echo "Direction 3b: annotation removed while the placeholder digest"
echo "            stays -> must FAIL (the unpullable pin is now undeclared)"
echo "=============================================================="
restore
python - "$REPO_PY/deploy/kubernetes/base/veid-inference-deployment.yaml" <<'PY'
import sys
p = sys.argv[1]
t = open(p, encoding="utf-8").read()
marker = "virtengine.com/image-pin-status: placeholder-awaiting-release"
assert marker in t
open(p, "w", encoding="utf-8", newline="").write(
    "\n".join(l for l in t.split("\n") if marker not in l)
)
PY
out="$(run_guard)"
report "undeclared placeholder digest" FAIL "$out"
echo "$out" | grep -oE 'AssertionError:.*' | head -1 | cut -c1-260

echo
echo "=============================================================="
echo "Restoring the tree"
echo "=============================================================="
restore
out="$(run_guard)"
report "tree restored" PASS "$out"

echo
if [ "$fail" -eq 0 ]; then
  echo "ALL DIRECTIONS CONFIRMED"
else
  echo "AT LEAST ONE DIRECTION FAILED"
fi
exit "$fail"