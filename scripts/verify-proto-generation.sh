#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo"

before="$(mktemp)"
after="$(mktemp)"
trap 'rm -f "$before" "$after"' EXIT

tracked_hashes() {
  find sdk/go/node sdk/ts/src/generated sdk/artifacts/proto api/openapi \
    -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*.ts' -o -name '*.binpb' -o -name '*.sha256' -o -name 'inventory.json' -o -name 'virtengine-proto.swagger.json' \) \
    -print0 | sort -z | xargs -0 sha256sum
}

tracked_hashes > "$before"
"$repo/scripts/proto-generate.sh" all
tracked_hashes > "$after"
diff -u "$before" "$after"

if find sdk/go/node -name gateway_stub.go -print -quit | grep -q .; then
  echo "production gateway_stub.go remains" >&2
  exit 1
fi

node --test scripts/protoinventory/inventory.test.mjs
git diff --check
