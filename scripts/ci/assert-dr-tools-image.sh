#!/usr/bin/env bash
#
# Assert the runtime contract of a dr-tools image against the DR CronJobs that
# consume it.
#
#   Usage: scripts/ci/assert-dr-tools-image.sh <image-ref>
#
# The CronJobs in infra/kubernetes/dr/backup-cronjobs.yaml execute repo-owned
# scripts *directly* (command: ["/scripts/dr/x.sh"]) and mount no volume at
# /scripts, so the image under test is the sole supplier of both the scripts and
# their toolchain. This script is the single source of truth for that contract
# and is executed against BOTH
#
#   * the image built from the current tree (pull_request / validation runs), and
#   * the published image pulled back by its digest (post-publish proof),
#
# so a change that would break the DR CronJobs cannot pass either path.
#
# Exit codes: 0 contract holds, 1 contract violated, 2 usage/manifest error.

set -euo pipefail

IMG="${1:-}"
if [[ -z "$IMG" ]]; then
  echo "usage: $0 <image-ref>" >&2
  exit 2
fi

CRONJOB_MANIFEST="${CRONJOB_MANIFEST:-infra/kubernetes/dr/backup-cronjobs.yaml}"
if [[ ! -f "$CRONJOB_MANIFEST" ]]; then
  echo "CronJob manifest not found: $CRONJOB_MANIFEST" >&2
  exit 2
fi

failures=0
record() { # record <ok|FAIL> <description>
  if [[ "$1" == "ok" ]]; then
    echo "  ok    $2"
  else
    echo "  FAIL  $2"
    failures=$((failures + 1))
  fi
}
check() { # check <description> <command...>
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then record ok "$desc"; else record FAIL "$desc"; fi
}
in_image() { # in_image <shell-command>
  docker run --rm --entrypoint /bin/sh "$IMG" -c "$1"
}

echo "asserting runtime contract of $IMG"

# ---------------------------------------------------------------------------
# 1. Identity: the CronJob securityContext is runAsUser 1000 / runAsGroup 1000
#    with runAsNonRoot true.
# ---------------------------------------------------------------------------
uid="$(docker run --rm --entrypoint id "$IMG" -u 2>/dev/null || echo UNKNOWN)"
if [[ "$uid" == "1000" ]]; then
  record ok "runs as uid 1000"
else
  record FAIL "runs as uid 1000 (got: $uid)"
fi

# ---------------------------------------------------------------------------
# 2. readOnlyRootFilesystem: true and the only writable mount is the /tmp
#    emptyDir, so the image must run with a read-only rootfs and a writable
#    /tmp. The tmpfs is mounted mode=1777 because that is what kubelet gives an
#    emptyDir (world-writable, no setgid); docker's bare `--tmpfs /tmp` defaults
#    to root-owned mode=2755, which a uid-1000 container legitimately cannot
#    write, so it would report a false failure.
# ---------------------------------------------------------------------------
if docker run --rm --read-only --tmpfs /tmp:rw,mode=1777 --entrypoint /bin/sh "$IMG" \
  -c 'test -w /tmp && echo probe > /tmp/.dr-probe' >/dev/null 2>&1; then
  record ok "/tmp writable for uid 1000 under a read-only rootfs"
else
  record FAIL "/tmp writable for uid 1000 under a read-only rootfs"
fi

# ---------------------------------------------------------------------------
# 3. Toolchain the scripts shell out to.
# ---------------------------------------------------------------------------
for tool in bash jq curl openssl aws virtengine sha256sum tar gzip; do
  if in_image "command -v $tool" >/dev/null 2>&1; then
    record ok "tool present: $tool"
  else
    record FAIL "tool present: $tool"
  fi
done

# ---------------------------------------------------------------------------
# 4. Every command: path in the CronJob manifest must exist and be executable.
#    The exec bit is what the pod relies on (bare exec form, no `sh -c`).
# ---------------------------------------------------------------------------
mapfile -t entrypoints < <(
  grep -oE 'command: \["/scripts/dr/[^"]+"' "$CRONJOB_MANIFEST" |
    sed 's/.*"\(.*\)"/\1/' | sort -u
)
if [[ "${#entrypoints[@]}" -ge 3 ]]; then
  record ok "manifest declares ${#entrypoints[@]} dr entrypoints"
else
  record FAIL "manifest declares ${#entrypoints[@]} dr entrypoints (expected >= 3)"
fi
for ep in "${entrypoints[@]}"; do
  if in_image "test -x '$ep'"; then
    record ok "executable in image: $ep"
  else
    record FAIL "executable in image: $ep"
  fi
done

# ---------------------------------------------------------------------------
# 5. Negative control. `test -x` must be able to fail: a file the Dockerfile
#    pins to 0644 has to come back non-executable, otherwise every assertion in
#    section 4 is vacuous.
# ---------------------------------------------------------------------------
if in_image 'test -x /scripts/dr/README.md'; then
  record FAIL "negative control: /scripts/dr/README.md (pinned 0644) is executable"
else
  record ok "negative control: /scripts/dr/README.md (pinned 0644) is not executable"
fi

# ---------------------------------------------------------------------------
# 6. Every shipped script must be valid bash with LF line endings. A CRLF blob
#    would still be +x and would fail at runtime with "\\r: command not found",
#    which is exactly the class of break the mode/exec assertions cannot see.
# ---------------------------------------------------------------------------
# shellcheck disable=SC2016  # the $f must stay unexpanded: it is evaluated inside the container
if in_image 'for f in /scripts/dr/*.sh; do bash -n "$f" || exit 1; done'; then
  record ok "all /scripts/dr/*.sh pass bash -n"
else
  record FAIL "all /scripts/dr/*.sh pass bash -n"
fi
# The detector is self-tested first. The image's /bin/sh is busybox, and busybox
# grep has no -U option: it exits 2 with a usage error and writes nothing, so a
# naive `grep -lU <CR> ... | wc -l` returns 0 and reports "no CR bytes" for a
# file that is full of them. Asserting the detector against a known-CRLF and a
# known-LF file means a toolchain change cannot make this check silently vacuous.
# shellcheck disable=SC2016  # $(printf "\r") must stay unexpanded: it runs inside the container
detector="$(in_image '
  printf "x\r\n" > /tmp/.cr-self
  printf "x\n"   > /tmp/.lf-self
  if grep -q "$(printf "\r")" /tmp/.cr-self && ! grep -q "$(printf "\r")" /tmp/.lf-self; then
    echo yes
  else
    echo no
  fi
  rm -f /tmp/.cr-self /tmp/.lf-self' 2>/dev/null || echo no)"
if [[ "$detector" == "yes" ]]; then
  record ok "CR detector self-test (flags a CRLF file, clears an LF file)"
else
  record FAIL "CR detector self-test (flags a CRLF file, clears an LF file)"
fi
# shellcheck disable=SC2016  # $f must stay unexpanded: it is evaluated inside the container
cr="$(in_image 'n=0; for f in /scripts/dr/*.sh; do grep -q "$(printf "\r")" "$f" && n=$((n+1)); done; echo "$n"' 2>/dev/null || echo unknown)"
if [[ "$cr" == "0" ]]; then
  record ok "no CR bytes in /scripts/dr/*.sh"
else
  record FAIL "no CR bytes in /scripts/dr/*.sh (found in ${cr} file(s))"
fi

echo
if [[ "$failures" -gt 0 ]]; then
  echo "CONTRACT VIOLATED: $failures check(s) failed for $IMG"
  exit 1
fi
echo "CONTRACT HOLDS: $IMG"
