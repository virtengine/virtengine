#!/usr/bin/env bash
set -euo pipefail

SPEC_FILE="api/openapi/portal_api.yaml"
TS_OUTPUT_DIR="lib/portal/src/provider-api/generated"
TS_OUTPUT_FILE="$TS_OUTPUT_DIR/types.ts"
GO_OUTPUT_DIR="pkg/provider_daemon/api/generated"
GO_OUTPUT_FILE="$GO_OUTPUT_DIR/types.go"
DOC_OUTPUT_FILE="docs/api/openapi/provider-portal.html"

mkdir -p "$TS_OUTPUT_DIR" "$GO_OUTPUT_DIR"

npx --yes @redocly/cli@latest lint "$SPEC_FILE"

npx --yes openapi-typescript@latest "$SPEC_FILE" -o "$TS_OUTPUT_FILE"

go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0 \
  -generate types \
  -package generated \
  "$SPEC_FILE" > "$GO_OUTPUT_FILE"

gofmt -w "$GO_OUTPUT_FILE"

npx --yes @redocly/cli@latest build-docs "$SPEC_FILE" -o "$DOC_OUTPUT_FILE"

echo "Provider API types and docs generated."
