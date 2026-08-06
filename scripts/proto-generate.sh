#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
image="virtengine-proto-gen:1.0.0"
mode="${1:-all}"
cache="${VE_PROTO_CACHE:-$repo/.cache/proto-generation}"

mkdir -p "$cache/buf" "$cache/go-mod" "$cache/go-build" "$cache/npm" "$cache/home" "$cache/npm-prefix"

docker build \
  --file "$repo/sdk/generation/Dockerfile" \
  --tag "$image" \
  "$repo"

docker run --rm \
  --user "$(id -u):$(id -g)" \
  --env HOME=/cache/home \
  --env npm_config_prefix=/cache/npm-prefix \
  --volume "$repo:/src" \
  --volume "$cache:/cache" \
  --workdir /src \
  "$image" "$mode"
