#!/usr/bin/env bash
# verify-pinned-images.sh — fail loudly on floating/mutable container base images.
#
# Purpose: every EXTERNAL base image referenced by the Dockerfiles scanned in the
# Security workflow's container-scan matrix must be pinned to an immutable digest:
#
#   FROM golang:1.25.14-alpine@sha256:<64 hex> AS build
#
# Floating tags silently change what we ship and what Trivy scans, so one
# upstream rebuild can flip four gates at once with no repo diff to blame.
# This script is enforced by the Policy Validation job in security.yaml.
#
# Rules per FROM instruction:
#   - FROM scratch            -> allowed (no digest exists for scratch)
#   - FROM <local-stage>      -> allowed (multi-stage internal reference,
#                                resolved from an earlier AS <name> in the file)
#   - FROM <img>@sha256:<hex> -> allowed, unless the digest is a known
#                                placeholder (see DENY_DIGESTS below)
#   - anything else           -> FAIL (bare name, :tag, :latest, variables)
#
# Usage:
#   scripts/supply-chain/verify-pinned-images.sh [--self-test] [Dockerfile ...]
#   No file args: checks the four security-scanned Dockerfiles under _build/.
#   --self-test: runs built-in positive/negative fixtures and exits.
#
# Exit status: 0 when every external base is pinned, 1 otherwise.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Default scope: exactly the container-scan matrix in .github/workflows/security.yaml.
DEFAULT_FILES=(
  "_build/Dockerfile.virtengine"
  "_build/Dockerfile.provider-daemon"
  "_build/Dockerfile.test"
  "_build/Dockerfile.veid-pipeline"
)

# Digests that match the FORMAT but are not real pins. The veid-pipeline base
# once carried an all-zero-filled placeholder digest; reject it by value so a
# "pinned-looking" line can never sneak back in. (All-zero digests rejected too.)
DENY_DIGESTS=(
  "0b22a29f55c0f3e4d2e8c79f9eba1d2a0a6a6a0b0b0b0b0b0b0b0b0b0b0b0b0b"
)

failures=0

lower() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]'
}

check_file() {
  local file="$1"
  local lineno=0
  local line ref rest alias
  local -a stages=()
  local -a froms=()

  if [[ ! -f "$file" ]]; then
    echo "ERROR: $file: file not found" >&2
    failures=$((failures + 1))
    return
  fi

  # Pass 1: collect local stage names from "AS <name>".
  while IFS= read -r line || [[ -n "$line" ]]; do
    lineno=$((lineno + 1))
    line="${line%%#*}"
    if [[ "$line" =~ ^[[:space:]]*[Ff][Rr][Oo][Mm][[:space:]]+([^[:space:]]+)([[:space:]]+[Aa][Ss][[:space:]]+([A-Za-z0-9_.-]+))? ]]; then
      alias="${BASH_REMATCH[3]:-}"
      if [[ -n "$alias" ]]; then
        stages+=("$(lower "$alias"):$lineno")
      fi
      froms+=("$lineno:${BASH_REMATCH[1]}")
    fi
  done < "$file"

  # Pass 2: validate every FROM reference.
  local entry fl ref_low ok s
  for entry in ${froms[@]+"${froms[@]}"}; do
    fl="${entry%%:*}"
    ref="${entry#*:}"
    # Strip --platform/--param flags: the reference is the last token.
    ref="${ref##* }"
    ref_low="$(lower "$ref")"

    # scratch needs no digest.
    if [[ "$ref_low" == "scratch" ]]; then
      continue
    fi

    # Local multi-stage reference needs no digest.
    ok=0
    for s in ${stages[@]+"${stages[@]}"}; do
      if [[ "$ref_low" == "${s%%:*}" ]]; then
        ok=1
        break
      fi
    done
    if [[ "$ok" == "1" ]]; then
      continue
    fi

    # External image: require @sha256:<64 hex>.
    if [[ "$ref" =~ @sha256:([0-9a-fA-F]{64})$ ]]; then
      local digest
      digest="$(lower "${BASH_REMATCH[1]}")"
      if [[ "$digest" =~ ^0+$ ]]; then
        echo "ERROR: $file:$fl: placeholder all-zero digest is not a pin: $ref" >&2
        failures=$((failures + 1))
        continue
      fi
      local denied
      for denied in "${DENY_DIGESTS[@]}"; do
        if [[ "$digest" == "$denied" ]]; then
          echo "ERROR: $file:$fl: known placeholder digest is not a pin: $ref" >&2
          failures=$((failures + 1))
          ok=2
          break
        fi
      done
      if [[ "$ok" == "2" ]]; then
        continue
      fi
      continue
    fi

    echo "ERROR: $file:$fl: floating base image (pin to name:tag@sha256:<digest>): $ref" >&2
    failures=$((failures + 1))
  done
}

self_test() {
  SELFTEST_TMP="$(mktemp -d)"
  trap 'rm -rf "$SELFTEST_TMP"' EXIT
  local tmpdir="$SELFTEST_TMP"

  cat > "$tmpdir/unresolved.dockerfile" <<'EOF'
FROM golang:1.25.14-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS build
RUN echo hi
FROM alpine:3.20@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc
COPY --from=build /out/app /app
# "base" is never defined with AS in this file: an unknown bare ref must fail.
FROM base AS extra
EOF
  cat > "$tmpdir/floating.dockerfile" <<'EOF'
FROM alpine:3.20
EOF
  cat > "$tmpdir/placeholder.dockerfile" <<'EOF'
FROM python:3.11.7-slim-bookworm@sha256:0b22a29f55c0f3e4d2e8c79f9eba1d2a0a6a6a0b0b0b0b0b0b0b0b0b0b0b0b0b AS base
EOF
  cat > "$tmpdir/stages.dockerfile" <<'EOF'
FROM python:3.11.7-slim-bookworm@sha256:53d6284a40eae6b625f22870f5faba6c54f2a28db9027408f4dee111f1e885a2 AS base
FROM base AS python-deps
FROM python-deps AS models
FROM gcr.io/distroless/python3-debian12@sha256:2fdb05402a2cf21cf78fdb3ba4c5db167241e9e498140f5bf689d7efb773731f AS final
EOF
  cat > "$tmpdir/scratch.dockerfile" <<'EOF'
FROM scratch
EOF

  local passed=0 failed=0
  local script_path="${BASH_SOURCE[0]}"

  # Positive: fully-pinned multi-stage file must pass.
  if bash "$script_path" "$tmpdir/stages.dockerfile" >/dev/null 2>&1; then
    echo "self-test PASS: pinned multi-stage file accepted"
    passed=$((passed + 1))
  else
    echo "self-test FAIL: pinned multi-stage file rejected" >&2
    failed=$((failed + 1))
  fi

  # Negative: floating tag must fail.
  if bash "$script_path" "$tmpdir/floating.dockerfile" >/dev/null 2>&1; then
    echo "self-test FAIL: floating tag accepted (must be rejected)" >&2
    failed=$((failed + 1))
  else
    echo "self-test PASS: floating tag rejected"
    passed=$((passed + 1))
  fi

  # Negative: placeholder digest must fail.
  if bash "$script_path" "$tmpdir/placeholder.dockerfile" >/dev/null 2>&1; then
    echo "self-test FAIL: placeholder digest accepted (must be rejected)" >&2
    failed=$((failed + 1))
  else
    echo "self-test PASS: placeholder digest rejected"
    passed=$((passed + 1))
  fi

  # Negative: unknown stage / scratch-like bare ref must fail.
  if bash "$script_path" "$tmpdir/unresolved.dockerfile" >/dev/null 2>&1; then
    echo "self-test FAIL: unresolved stage ref accepted (must be rejected)" >&2
    failed=$((failed + 1))
  else
    echo "self-test PASS: unresolved stage ref rejected"
    passed=$((passed + 1))
  fi

  # Positive: FROM scratch needs no digest.
  if bash "$script_path" "$tmpdir/scratch.dockerfile" >/dev/null 2>&1; then
    echo "self-test PASS: scratch accepted"
    passed=$((passed + 1))
  else
    echo "self-test FAIL: scratch rejected" >&2
    failed=$((failed + 1))
  fi

  echo "self-test: $passed passed, $failed failed"
  if [[ "$failed" -gt 0 ]]; then
    return 1
  fi
}

if [[ "${1:-}" == "--self-test" ]]; then
  self_test
  exit $?
fi

files=()
if [[ $# -eq 0 ]]; then
  for f in "${DEFAULT_FILES[@]}"; do
    files+=("$REPO_ROOT/$f")
  done
else
  files=("$@")
fi

for f in "${files[@]}"; do
  check_file "$f"
done

if [[ "$failures" -gt 0 ]]; then
  echo "verify-pinned-images: $failures unpinned base image(s) — pin to @sha256 digests" >&2
  exit 1
fi
echo "verify-pinned-images: all base images pinned (${#files[@]} file(s))"
