#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
modules=(
  .
  sdk/go
  sdk/specs
  infra/tests
  infra/terraform/tests
  infra/dr/tests
  sdk/generation
)

policy="$repo/scripts/supply-chain/go-module-policy.json"

for module in "${modules[@]}"; do
  echo "==> go mod verify: $module"
  (cd "$repo/$module" && GOWORK=off go mod verify)
done

node "$repo/scripts/supply-chain/verify-go-module-policy.mjs" "$policy"

if [[ "${VE_VERIFY_EMPTY_CACHE:-0}" == "1" ]]; then
  cache="$(mktemp -d)"
  trap 'rm -rf "$cache"' EXIT
  for module in "${modules[@]}"; do
    echo "==> empty-cache download: $module"
    (cd "$repo/$module" && GOWORK=off GOMODCACHE="$cache/mod" GOCACHE="$cache/build" go mod download all)
  done
fi

sdk_vendor="$(mktemp -d)"
trap 'rm -rf "$sdk_vendor"' EXIT
(
  cd "$repo/sdk/go"
  GOWORK=off go mod vendor -o "$sdk_vendor"
  test -s "$sdk_vendor/modules.txt"
)

for checksum in github.com/golang/glog github.com/moby/spdystream github.com/mxk/go-flowrate; do
  grep -q "^${checksum} " "$repo/go.sum" "$repo/sdk/go/go.sum" || {
    echo "missing checksum metadata for $checksum" >&2
    exit 1
  }
done

echo "module/checksum/vendor verification passed"
