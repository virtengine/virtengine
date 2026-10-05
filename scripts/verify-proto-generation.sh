#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo"

before="$(mktemp)"
after="$(mktemp)"
trap 'rm -f "$before" "$after"' EXIT

# The two ajv validators belong here for the same reason they are named in
# proto-generation.yaml's drift pathspec: `compile:validators`, reached via
# `npm --prefix ts run build` at the end of generate_typescript(), writes both
# of them, so this two-run byte-identity comparison is meaningless without them.
# Named as files rather than as sdk/ts/src/sdk + sdk/ts/src/sdl, because those
# trees are otherwise hand-written SDK source.
validator_jwt="sdk/ts/src/sdk/provider/auth/jwt/validateJwtPayload.ts"
validator_sdl="sdk/ts/src/sdl/SDL/validateSDL/validateSDLInput.ts"

tracked_hashes() {
  find sdk/go/node sdk/ts/src/generated sdk/artifacts/proto api/openapi \
    -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' -o -name '*.ts' -o -name '*.binpb' -o -name '*.sha256' -o -name 'inventory.json' -o -name 'virtengine-proto.swagger.json' \) \
    -print0 | sort -z | xargs -0 sha256sum
  # The generators under sdk/ts/src/sdk and sdk/ts/src/sdl are named files, not
  # directories, so they cannot join the find above without also sweeping in
  # hand-written SDK source.
  sha256sum "$validator_jwt" "$validator_sdl"
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
