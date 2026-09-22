#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
#
# Verify that every `<file>.sha256` integrity sidecar in the tree records the
# SHA-256 of the bytes git actually stores for `<file>`.
#
# Why this exists
# ---------------
# `x/veid/zk/params/params_metadata.json.sha256` twice recorded the digest of the
# metadata file *as checked out on a Windows working tree* (CRLF, 21261041...) rather
# than the digest of the committed blob (LF, a4e05914...). Git stores and `go:embed`
# embeds the LF form, because `.gitattributes` pins
# `/x/veid/zk/params/*.json text eol=lf`.
#
# Nothing in the repo checked a sidecar before it landed, so the bad digest reached
# `main` twice and made `loadArtifactSet` fail on the embedded bundle. `keeper.NewKeeper`
# then panicked with "failed to initialize VEID ZK proof system", which reddened the
# `Go Tests` job and 16 packages (x/veid, x/veid/keeper, x/cert, x/deployment,
# x/escrow/keeper, x/market/keeper, x/provider, x/settlement/client/cli, upgrades,
# tests/compatibility, ...) plus 4 integration packages.
#
# The check hashes the git blob, never the working tree, so it is immune to the
# autocrlf/eol environment that produced the bad values in the first place, and it
# fails identically on Linux, macOS and Windows.
#
# Usage:
#   bash scripts/veid-params-sidecar-verify.sh               # scan the whole tree
#   bash scripts/veid-params-sidecar-verify.sh <repo-root>   # scan a specific tree
#
set -euo pipefail

if [[ -n "${1:-}" ]]; then
  repo="$1"
else
  repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

if [[ ! -d "$repo/.git" && ! -f "$repo/.git" ]]; then
  echo "error: $repo is not a git repository" >&2
  exit 2
fi

cd "$repo"

checked=0
failures=0
# Artifacts skipped by an explicit, reviewed decision (path<TAB>reason).
declare -A allowlist=()
# Allow repo-local opt-out for genuinely unverifiable vendored artifacts.
if [[ -n "${VE_SIDECAR_ALLOWLIST:-}" ]]; then
  while IFS=$'\t' read -r path reason; do
    [[ -z "${path:-}" ]] && continue
    allowlist["$path"]="${reason:-allowlisted}"
  done < <(printf '%s\n' "$VE_SIDECAR_ALLOWLIST")
fi

# Enumerate tracked sidecars, then resolve each artifact through the index so the
# path is repository-relative even when a caller passes an absolute repo root.
while IFS= read -r sidecar; do
  [[ -z "$sidecar" ]] && continue

  if [[ -v "allowlist[$sidecar]" ]]; then
    echo "skip: ${sidecar} (${allowlist[$sidecar]})"
    continue
  fi

  artifact="${sidecar%.sha256}"

  if ! git cat-file -e "HEAD:${artifact}" 2>/dev/null; then
    echo "FAIL: ${sidecar}: tracked sidecar has no tracked artifact"
    failures=$((failures + 1))
    continue
  fi

  # Expected: the digest recorded in the sidecar record. Accept either the binary-mode
  # marker (`*name`) or the text-mode form (` name`).
  record="$(git cat-file -p "HEAD:${sidecar}")"
  recorded="$(printf '%s\n' "$record" | awk 'NR == 1 { print $1 }')"
  label="$(printf '%s\n' "$record" | awk 'NR == 1 { print $2 }')"
  label="${label#\*}"
  label="${label#./}"

  if [[ ! "$recorded" =~ ^[0-9a-f]{64}$ ]]; then
    echo "FAIL: ${sidecar}: first field is not a 64-char lowercase SHA-256 hex digest"
    failures=$((failures + 1))
    continue
  fi

  base="${artifact##*/}"
  if [[ "$label" != "$artifact" && "$label" != "$base" ]]; then
    echo "FAIL: ${sidecar}: checksum label '${label}' names neither '${artifact}' nor '${base}'"
    failures=$((failures + 1))
    continue
  fi

  # Actual: the digest of the committed blob. This is exactly what git stores and
  # what go:embed embeds -- deliberately not the working-tree bytes.
  actual="$(git cat-file -p "HEAD:${artifact}" | sha256sum | awk '{ print $1 }')"

  if [[ "$recorded" != "$actual" ]]; then
    echo "FAIL: ${sidecar}: expected ${recorded} got ${actual}"
    echo "      artifact: ${artifact}"
    echo "      fix: printf '%s  %s\\n' \"\$(git cat-file -p HEAD:${artifact} | sha256sum | cut -d' ' -f1)\" \"${artifact}\" > ${sidecar}"
    failures=$((failures + 1))
    continue
  fi

  echo "ok:   ${sidecar} -> ${artifact}"
  checked=$((checked + 1))
done < <(git ls-files '*.sha256' | LC_ALL=C sort)

echo
if (( failures > 0 )); then
  echo "sha256 sidecar integrity: ${failures} mismatched, ${checked} verified"
  exit 1
fi

echo "sha256 sidecar integrity: 0 mismatched, ${checked} verified"
