#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo="${VE_PROTO_REPO:-/src}"
sdk="$repo/sdk"
mode="${1:-all}"
go_template="${VE_PROTO_GO_TEMPLATE:-buf.gen.go.yaml}"
openapi_template="${VE_PROTO_OPENAPI_TEMPLATE:-buf.gen.openapi.yaml}"
ts_template="${VE_PROTO_TS_TEMPLATE:-buf.gen.ts.yaml}"

export PATH="/go/bin:/usr/local/bin${VE_PROTO_NODE_BIN:+:$VE_PROTO_NODE_BIN}:/usr/bin:/bin:${PATH:-}"
export GOWORK=off
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
export LC_ALL=C
export LANG=C
export TZ=UTC
export SOURCE_DATE_EPOCH=0
export BUF_CACHE_DIR="${BUF_CACHE_DIR:-/cache/buf}"
export GOMODCACHE="${GOMODCACHE:-/cache/go-mod}"
export GOCACHE="${GOCACHE:-/cache/go-build}"
export npm_config_cache="${npm_config_cache:-/cache/npm}"

cd "$sdk"

install_typescript() {
  npm --prefix ts ci --ignore-scripts --no-audit --no-fund
}

build_descriptor() {
  mkdir -p artifacts/proto
  buf build --as-file-descriptor-set -o artifacts/proto/virtengine.binpb
  (
    cd "$repo"
    sha256sum sdk/artifacts/proto/virtengine.binpb > sdk/artifacts/proto/virtengine.binpb.sha256
  )
}

generate_go() {
  buf generate --template "$go_template"
  source_root="$sdk/.generation/go/github.com/virtengine/virtengine/sdk/go/node"
  test -d "$source_root"
  find "$source_root" -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' \) -print0 |
    while IFS= read -r -d '' source; do
      destination="$sdk/go/node/${source#"$source_root/"}"
      mkdir -p "$(dirname "$destination")"
      cp "$source" "$destination"
    done
  if find "$sdk/go/node" -name gateway_stub.go -print -quit | grep -q .; then
    echo "production gateway_stub.go remains" >&2
    exit 1
  fi
}

generate_openapi() {
  buf generate --template "$openapi_template"
  node "$repo/scripts/protoinventory/compose-openapi.mjs"
}

generate_typescript() {
  install_typescript
  rm -rf ts/src/generated
  PROTO_SOURCE=node buf generate --template "$ts_template" proto/node
  PROTO_SOURCE=cosmos buf generate --template "$ts_template" go/vendor/github.com/cosmos/cosmos-sdk/proto
  PROTO_SOURCE=ibc-go buf generate --template "$ts_template" go/vendor/github.com/cosmos/ibc-go/v10/proto
  PROTO_SOURCE=provider buf generate --template "$ts_template" proto/provider
  node --experimental-strip-types --no-warnings ts/script/fix-ts-proto-generated-types.ts
  k8s_source="ts/src/generated/protos/k8s_io/apimachinery/pkg/api/resource/generated.ts"
  k8s_compat="ts/src/generated/protos/k8s.io/apimachinery/pkg/api/resource/generated.ts"
  if [[ -f "$k8s_source" ]]; then
    mkdir -p "$(dirname "$k8s_compat")"
    cp "$k8s_source" "$k8s_compat"
  fi
  npm --prefix ts run build
}

generate_inventory() {
  node "$repo/scripts/protoinventory/inventory.mjs"
}

case "$mode" in
  all)
    build_descriptor
    generate_go
    generate_openapi
    generate_typescript
    generate_inventory
    ;;
  go) generate_go ;;
  descriptor) build_descriptor ;;
  openapi) generate_openapi ;;
  ts) generate_typescript ;;
  inventory) generate_inventory ;;
  verify)
    buf build >/dev/null
    node --test "$repo/scripts/protoinventory/inventory.test.mjs"
    generate_inventory
    ;;
  python|rust)
    echo "$mode protobuf output is not a supported release contract; see sdk/generation/toolchain.json" >&2
    exit 2
    ;;
  *)
    echo "usage: generate.sh [all|go|descriptor|openapi|ts|inventory|verify]" >&2
    exit 2
    ;;
esac
