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
# --- module cache: one root, two views ---------------------------------------
# modvendor resolves every module as $GOPATH/pkg/mod/<path>@<version> and reads
# GOPATH ONLY -- it never consults GOMODCACHE (grep GOMODCACHE in
# goware/modvendor => 0 hits). The golang base image sets GOPATH=/go while this
# script put the module cache at /cache/go-mod, so modvendor searched a
# directory nothing ever populates and aborted the run:
#
#   Error! "/go/pkg/mod/cloud.google.com/go@v0.123.0" module path does not exist, check $GOPATH/pkg/mod
#
# Derive both views from ONE root so they cannot disagree. The root defaults to
# /cache -- the directory proto-generate.sh actually mounts, so the cache still
# persists across runs -- and NOT to an inherited GOPATH, because the base image
# presets GOPATH=/go, which is container-local and would silently throw the
# module cache away on every run. A caller that already supplies GOMODCACHE at a
# proper <root>/pkg/mod keeps its own root.
if [[ "${GOMODCACHE:-}" == */pkg/mod ]]; then
  export GOPATH="${GOPATH:-${GOMODCACHE%/pkg/mod}}"
else
  export GOPATH="${VE_PROTO_GOPATH:-/cache}"
  export GOMODCACHE="$GOPATH/pkg/mod"
fi
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
  # Ensure vendor directory has proto files for cosmos-sdk and ibc-go.
  # modvendor reads vendor/modules.txt, which only exists once the module has
  # actually been vendored. CI's `contracts` job runs this script in the image
  # with no vendor/ present, so modvendor aborted the whole generation:
  #
  #   Whoops, cannot find vendor/modules.txt, first run `go mod vendor` and try again
  #
  # The --proto sources copied below and the TS templates both read from
  # go/vendor/..., so the vendor tree is a precondition of this step, not an
  # optional optimisation. Stage it when it is missing.
  (cd "$sdk/go" && [ -f vendor/modules.txt ] || go mod vendor)
  # modvendor stats $GOPATH/pkg/mod/<path>@<version> for EVERY `# <path> <version>`
  # stanza in vendor/modules.txt - including modules that contribute NO vendored
  # package at all. `go mod vendor` writes such a stanza for every module named in
  # go.mod (github.com/golang/mock is a `// indirect` require that no imported
  # package uses) but never downloads the source of one that provides no package,
  # so modvendor aborted the whole generation before copying anything:
  #
  #   Error! "/cache/pkg/mod/github.com/golang/mock@v1.7.0-rc.1" module path does not exist, check $GOPATH/pkg/mod
  #
  # Reproduced offline against the gate's own inputs: `go mod vendor` in sdk/go
  # writes the golang/mock stanza with no package lines, and modvendor exits 1 on
  # the existence check for it. Stage the source of every package-less stanza
  # first - modvendor only stats that directory and copies nothing from it - so an
  # unused go.mod entry cannot red the contracts job.
  (
    cd "$sdk/go"
    awk '
      /^# / && NF == 3 { if (path != "" && !has_pkg) print path "@" ver; path = $2; ver = $3; has_pkg = 0; next }
      /^#/            { next }
      NF              { has_pkg = 1 }
      END             { if (path != "" && !has_pkg) print path "@" ver }
    ' vendor/modules.txt | xargs -r go mod download
  )
  # modvendor keeps a matched file only when it sits under a directory that is
  # ALSO a package of that module (main.go: importPathIntersect, then
  # strings.Index(vendorFile, path) == 0), so a module whose ROOT package is not
  # vendored contributes nothing from its proto/ tree. cosmos-sdk and ibc-go are
  # exactly that, and the buf steps below then die with
  #
  #   Failure: Module "path: "go/vendor/github.com/cosmos/cosmos-sdk/proto"" had no .proto files
  #
  # (measured: run 36525756868, step "Regenerate all contracts" after the
  # package-less-stanza fix landed). `-include` appends these directories to the
  # module's package list - the case its own help text documents - which makes the
  # filter keep the proto trees the TS templates below read.
  (cd "$sdk/go" && modvendor -copy="**/*.proto" \
    -include="github.com/cosmos/cosmos-sdk/proto,github.com/cosmos/ibc-go/v10/proto" -v)
  # modvendor copies ONE module's proto tree at a time, so a vendored tree keeps
  # its own .proto files but loses the buf DEPENDENCIES its upstream buf.yaml
  # declared. buf resolves an import path relative to the module root it was
  # pointed at, so those imports then do not exist and buf exits 100:
  #
  #   go/vendor/github.com/cosmos/cosmos-sdk/proto/cosmos/auth/v1beta1/auth.proto:5:8:
  #     import "cosmos_proto/cosmos.proto": file does not exist
  #
  # (measured: PR run 36686760313 job contracts step 7, and byte-identically on
  # main run 36530441688 - inherited, not branch-introduced). BOTH vendored roots
  # need it, and each has a different dependency set: the cosmos-sdk tree has 7
  # unresolved imports and the ibc-go tree 16, of which the google/protobuf/*
  # ones are well-known types buf resolves itself:
  #
  #   cosmos_sdk: cosmos_proto/, gogoproto/, google/api/
  #   ibc_go:     all of the above PLUS cosmos/ and tendermint/ (from cosmos-sdk)
  #
  # Stage each dependency tree INSIDE the vendored root (a sibling directory is
  # invisible to buf - measured) and resolve every module through
  # `go list -m -f {{.Dir}}` so the repo's `replace` directives are honoured and
  # no version is hardcoded. Each tree is staged under BOTH roots, because
  # ibc-go imports cosmos-sdk's own protos and staging per-root would only move
  # the failure to the next buf step.
  #
  # The two paths are NOT the same and conflating them is the bug this shape
  # avoids: $3 is where the tree lives INSIDE the module (cosmos-sdk's
  # amino/ is at proto/amino), $4 is the import path buf must resolve it at
  # (amino/amino.proto). Passing the module-internal path as the import path
  # stages the tree at .../proto/proto/cosmos_proto and the import still fails
  # (measured).
  stage_proto_dependency() {
    local root="$1" module="$2" import_path="$3" module_subdir="$4"
    local dir target
    dir="$(cd "$sdk/go" && GOWORK=off go list -m -f '{{.Dir}}' "$module")"
    if [[ -z "$dir" || ! -d "$dir/$module_subdir" ]]; then
      # `go list -m -f '{{.Dir}}'` is NOT a reliable way to locate a module in
      # this image: it reports an empty Dir for a module that is required
      # plainly, is already downloaded, and has no replace directive
      # (measured twice - contracts runs 36718051031 and 36722480279, both
      # printing `dir=[]` for github.com/cosmos/cosmos-proto next to a
      # successful `go: downloading github.com/cosmos/cosmos-proto
      # v1.0.0-beta.5`). So: download, then GLOB the module cache, which is
      # correct by construction because the shell expands whatever go wrote and
      # the version is the only variable part.
      #
      # Locate the module in \$GOMODCACHE. For a REPLACED module neither
      # {{.Path}}@{{.Version}} nor the required version names the directory:
      # go.mod:292-293 replace cosmos-sdk -> virtengine/cosmos-sdk
      # v0.53.4-virtengine.2 and gogoproto -> virtengine/gogoproto
      # v1.7.0-virtengine.1, and the cache is keyed on the REPLACEMENT's path
      # and version (measured: the cache holds
      # github.com/cosmos/cosmos-sdk@v0.50.11 / v0.53.0 / v0.54.3 and
      # github.com/cosmos/gogoproto@v1.7.2 while the build uses
      # github.com/virtengine/gogoproto@v1.7.0-virtengine.1 - so both the
      # original path AND the original version point at the wrong tree).
      # Try the replacement identity first, then the plain one, and require the
      # subdir to exist in the winner - that is what makes the choice safe.
      local want replace_hit plain_hit
      want="$(cd "$sdk/go" && GOWORK=off go list -m -f '{{with .Replace}}{{.Path}}@{{.Version}}{{else}}{{.Path}}@{{.Version}}{{end}}' "$module" 2>/dev/null | head -1)"
      if [[ -z "$want" || "$want" == "$module@" ]]; then
        echo "proto dependency $module has no version selected by sdk/go" >&2
        return 1
      fi
      (cd "$sdk/go" && GOWORK=off go mod download "$module")
      replace_hit="$GOMODCACHE/$want"
      plain_hit="$GOMODCACHE/$(cd "$sdk/go" && GOWORK=off go list -m -f '{{.Path}}@{{.Version}}' "$module" 2>/dev/null | head -1)"
      local hit
      for hit in "$replace_hit" "$plain_hit"; do
        if [[ -n "$hit" && -d "$hit/$module_subdir" ]]; then
          dir="$hit"
          break
        fi
      done
      if [[ -z "$dir" || ! -d "$dir/$module_subdir" ]]; then
        echo "proto dependency $module (want $want) has no $module_subdir in \$GOMODCACHE" >&2
        return 1
      fi
    fi
    target="$root/$import_path"
    # Skip a tree the vendored root already has. cosmos-sdk's own proto/ ships
    # amino/, cosmos/ and tendermint/ already, and re-staging them there made
    # buf abort with `symbol "amino.name" already defined` (measured).
    if [[ -d "$target" ]]; then
      return 0
    fi
    mkdir -p "$(dirname "$target")"
    # `cp -r src target` COPIES INTO target when target exists, producing
    # target/amino/amino.proto - a nested copy buf then sees as a duplicate
    # symbol (measured). The parent is created, the leaf must not exist, so copy
    # the CONTENTS: the `-T` flag is GNU-only, do it portably.
    mkdir -p "$target"
    cp -r "$dir/$module_subdir/." "$target/"
  }
  stage_all_proto_dependencies() {
    local root="$1"
    stage_proto_dependency "$root" github.com/cosmos/cosmos-proto "cosmos_proto" "proto/cosmos_proto"
    stage_proto_dependency "$root" github.com/cosmos/gogoproto "gogoproto" "gogoproto"
    stage_proto_dependency "$root" github.com/grpc-ecosystem/grpc-gateway \
      "google/api" "third_party/googleapis/google/api"
    stage_proto_dependency "$root" github.com/cosmos/cosmos-sdk "amino" "proto/amino"
    stage_proto_dependency "$root" github.com/cosmos/cosmos-sdk "cosmos" "proto/cosmos"
    stage_proto_dependency "$root" github.com/cosmos/cosmos-sdk "tendermint" "proto/tendermint"
  }
  local vendored_cosmos="$sdk/go/vendor/github.com/cosmos/cosmos-sdk/proto"
  local vendored_ibc="$sdk/go/vendor/github.com/cosmos/ibc-go/v10/proto"
  stage_all_proto_dependencies "$vendored_cosmos"
  stage_all_proto_dependencies "$vendored_ibc"
  # One import cannot be staged: the ibc-go tree imports
  # `cosmos/ics23/v1/proofs.proto`, and NO pinned Go module ships that .proto -
  # the virtengine cosmos-sdk fork has 0 ics23 protos (so do upstream v0.53.0
  # and v0.54.3), and github.com/cosmos/ics23/go ships no .proto at all
  # (measured). Upstream resolves it from the BSR: ibc-go's own proto/buf.yaml
  # pins `buf.build/cosmos/ics23`, and modvendor copies only **/*.proto, so that
  # dependency set is dropped along with the lock. Declare it explicitly, at
  # the module root buf is pointed at, and let buf resolve it from the registry.
  # The buf.build.lock that `buf dep update` writes is build output, not a
  # source artifact, so it is not committed - the gate re-resolves each run.
  # The config is written into a fresh writable directory that IS the ibc-go
  # module root, not into the vendored tree: the module cache is mode 444
  # (read-only) and modvendor's tree inherits that, so writing buf.yaml in place
  # fails with `Permission denied` (measured). The staging block above already
  # creates that directory's contents, and the tree is inside $sdk (writable), so
  # `path: .` names a module at its own context directory - the only shape buf
  # accepts, since a module path may not escape it (`invalid module path:
  # ../v10/proto: is outside the context directory`, measured).
  local ics23_dep_dir="$vendored_ibc"
  printf 'version: v2\nmodules:\n  - path: .\ndeps:\n  - buf.build/cosmos/ics23\n' \
    > "$ics23_dep_dir/buf.yaml"
  ( cd "$ics23_dep_dir" && buf dep update ) \
    || echo "warning: could not resolve buf.build/cosmos/ics23 from the registry" >&2
  install_typescript
  rm -rf ts/src/generated
  PROTO_SOURCE=node buf generate --template "$ts_template" proto/node
  PROTO_SOURCE=cosmos buf generate --template "$ts_template" go/vendor/github.com/cosmos/cosmos-sdk/proto
  # The ibc-go step needs `buf.build/cosmos/ics23`, which no pinned Go module
  # ships, so it gets its own config. Two measured constraints decide the shape:
  #   * a module path may not escape its context directory, so the config must
  #     name a module at or below its own directory;
  #   * buf resolves a template's RELATIVE `local:` plugin paths against the
  #     WORKING DIRECTORY, so this step must keep cwd=$sdk. `cd`-ing to the
  #     config's directory (the shape that worked for `buf build`) made the
  #     template's ["node","ts/script/protoc-gen-sdk-object.ts"] resolve to
  #     /src/sdk/go/vendor/github.com/cosmos/ts/script/... and the step died with
  #     `Cannot find module` (measured: contracts run 36735951768).
  # So: put the config INSIDE the ibc-go root (module `path: .`), pass it with
  # --config, AND name the input directory on the command line.
  #
  # The positional input is not optional. With `--config` and NO input, buf
  # takes the module's own `path` - `.` - and compiles the whole CONTEXT
  # directory, which for a config inside the shared vendor tree also contains
  # the sibling cosmos-sdk root, so every symbol they share is defined twice:
  #
  #   cosmos/group/v1/types.proto:296:3: symbol
  #   "cosmos.group.v1.PROPOSAL_EXECUTOR_RESULT_FAILURE" already defined at
  #   go/vendor/github.com/cosmos/cosmos-sdk/proto/cosmos/group/v1/types.proto:296:3
  #
  # (measured: contracts run 36743453258, and reproduced locally - `EXIT=100`
  # with and without the sibling tree, `EXIT=0` once the input is named
  # explicitly). Naming the directory restricts the build to that tree while the
  # config still supplies the ics23 dependency.
  # `buf dep update` must run first, or the dep is declared but unresolved and
  # `cosmos/ics23/v1/proofs.proto` still does not resolve (measured).
  # Drop the config afterwards so it cannot perturb the generated-drift check.
  PROTO_SOURCE=ibc-go buf generate --template "$ts_template" \
    --config "$ics23_dep_dir/buf.yaml" "$vendored_ibc" \
    || { echo "ibc-go contract generation failed" >&2; return 1; }
  PROTO_SOURCE=provider buf generate --template "$ts_template" proto/provider
  node --experimental-strip-types --no-warnings ts/script/fix-ts-proto-generated-types.ts
  # The parent config is build scaffolding inside the vendor tree; leaving it
  # would make the next run's `go mod vendor`/`modvendor` see a stray module and
  # would show up in the generated-drift check.
  rm -f "$ics23_dep_dir/buf.yaml" "$ics23_dep_dir/buf.lock"
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
