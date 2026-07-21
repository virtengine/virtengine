#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
# Native Linux/WSL equivalent for validating the container workflow.
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin="${VE_PROTO_BIN:-$repo/.cache/proto-generation/bin}"
node_root="${VE_PROTO_NODE_ROOT:-$repo/.cache/proto-generation/node-24.12.0}"
protoc_root="${VE_PROTO_PROTOC_ROOT:-$repo/.cache/proto-generation/protoc-29.1}"
mode="${1:-all}"

export GOWORK=off
export GOTOOLCHAIN=go1.25.5
export GOBIN="$bin"
export VE_PROTO_NODE_BIN="$node_root/bin"
export PATH="$bin:$VE_PROTO_NODE_BIN:$protoc_root/bin:$PATH"
export BUF_CACHE_DIR="${BUF_CACHE_DIR:-$repo/.cache/proto-generation/buf}"
export GOMODCACHE="${GOMODCACHE:-$repo/.cache/proto-generation/go-mod}"
export GOCACHE="${GOCACHE:-$repo/.cache/proto-generation/go-build}"
export npm_config_cache="${npm_config_cache:-$repo/.cache/proto-generation/npm}"

mkdir -p "$bin" "$BUF_CACHE_DIR" "$GOMODCACHE" "$GOCACHE" "$npm_config_cache"
case "$(uname -m)" in
	x86_64)
		protoc_arch=x86_64
		protoc_sha=00c83fe9722d85e96c81b941b29f17a744b33b4ce66e0f18009fd8937de22c60
		node_arch=x64
		node_sha=bdebee276e58d0ef5448f3d5ac12c67daa963dd5e0a9bb621a53d1cefbc852fd
		;;
	aarch64|arm64)
		protoc_arch=aarch_64
		protoc_sha=1f74a3f3355de7c0666bc125611c13532c2598f853521d0d3e621a5b09f24799
		node_arch=arm64
		node_sha=a06d42807fb500f7459e5f3fa6cb431447352826ee6f07e14adfeec58a1b3210
		;;
	*)
		echo "unsupported architecture $(uname -m)" >&2
		exit 1
		;;
esac
if [[ ! -x "$protoc_root/bin/protoc" ]]; then
	archive="$(mktemp)"
	trap 'rm -f "$archive"' EXIT
	curl --fail --location --silent --show-error \
		"https://github.com/protocolbuffers/protobuf/releases/download/v29.1/protoc-29.1-linux-${protoc_arch}.zip" \
		--output "$archive"
	echo "$protoc_sha  $archive" | sha256sum --check --strict
	rm -rf "$protoc_root"
	mkdir -p "$protoc_root"
	unzip -q "$archive" 'bin/protoc' 'include/*' -d "$protoc_root"
	rm -f "$archive"
	trap - EXIT
fi
if [[ ! -x "$VE_PROTO_NODE_BIN/node" ]]; then
	archive="$(mktemp)"
	trap 'rm -f "$archive"' EXIT
	curl --fail --location --silent --show-error \
		"https://nodejs.org/dist/v24.12.0/node-v24.12.0-linux-${node_arch}.tar.xz" \
		--output "$archive"
	echo "$node_sha  $archive" | sha256sum --check --strict
	rm -rf "$node_root"
	mkdir -p "$node_root"
	tar -xJf "$archive" --strip-components=1 -C "$node_root"
	rm -f "$archive"
	trap - EXIT
fi
pushd "$repo/sdk/generation" >/dev/null
go mod verify
go install github.com/bufbuild/buf/cmd/buf
go install github.com/cosmos/gogoproto/protoc-gen-gocosmos
go install github.com/goware/modvendor
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
popd >/dev/null

# Container templates use /go/bin to prohibit host plugin discovery. For this
# equivalent path, materialize templates with the isolated absolute bin path.
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
sed "s#/go/bin/#$bin/#g" "$repo/sdk/buf.gen.go.yaml" > "$tmp/buf.gen.go.yaml"
sed "s#/go/bin/#$bin/#g" "$repo/sdk/buf.gen.openapi.yaml" > "$tmp/buf.gen.openapi.yaml"
sed "s#/src/sdk/#$repo/sdk/#g" "$repo/sdk/buf.gen.ts.yaml" > "$tmp/buf.gen.ts.yaml"

export VE_PROTO_GO_TEMPLATE="$tmp/buf.gen.go.yaml"
export VE_PROTO_OPENAPI_TEMPLATE="$tmp/buf.gen.openapi.yaml"
export VE_PROTO_TS_TEMPLATE="$tmp/buf.gen.ts.yaml"
export VE_PROTO_REPO="$repo"
exec "$repo/sdk/generation/generate.sh" "$mode"
