#!/bin/bash
# Proto generation script for Windows/WSL
# Run this directly in WSL: bash proto-gen-go.sh

set -e

# Setup environment
VIRTENGINE_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export VIRTENGINE_ROOT
export VIRTENGINE_DEVCACHE="$VIRTENGINE_ROOT/sdk/.cache"
export VIRTENGINE_DEVCACHE_BIN="$VIRTENGINE_DEVCACHE/bin"
export VIRTENGINE_DEVCACHE_VERSIONS="$VIRTENGINE_DEVCACHE/versions"
export VIRTENGINE_DEVCACHE_INCLUDE="$VIRTENGINE_DEVCACHE/include"
export VIRTENGINE_DEVCACHE_TMP="$VIRTENGINE_DEVCACHE/tmp"
export VIRTENGINE_TS_ROOT="$VIRTENGINE_ROOT/sdk/ts"
export VIRTENGINE_TS_NODE_MODULES="$VIRTENGINE_TS_ROOT/node_modules"
export VIRTENGINE_TS_NODE_BIN="$VIRTENGINE_TS_NODE_MODULES/.bin"
export PATH="$VIRTENGINE_DEVCACHE_BIN:$PATH"
export VIRTENGINE_DIRENV_SET=1
export GOTOOLCHAIN=go1.25.5

cd "$VIRTENGINE_ROOT/sdk"

echo "=== VirtEngine Proto Generation ==="
echo "VIRTENGINE_ROOT: $VIRTENGINE_ROOT"
echo ""

# Check if vendor directory is in sync
echo "Step 1: Syncing vendor directory..."
(cd go && GO111MODULE=on go mod tidy)
(cd go && GOWORK=off go mod vendor)

echo "Step 2: Verifying modules..."
(cd go && go mod verify) || {
    echo "Module verification failed. Running go mod vendor again..."
    (cd go && GOWORK=off go mod vendor)
}

echo "Step 3: Running modvendor for proto files..."
if command -v modvendor &> /dev/null; then
    (cd go && modvendor -copy="**/*.proto" -v)
else
    echo "Installing modvendor..."
    go install github.com/goware/modvendor@v0.5.0
    (cd go && modvendor -copy="**/*.proto" -v)
fi

echo "Step 4: Setting up include symlinks..."
mkdir -p .cache/include/k8s
ln -snf ../../../go/vendor/k8s.io .cache/include/k8s/k8s.io 2>/dev/null || true

echo "Step 5: Running protocgen..."
GO_MOD_NAME=$(cd go && GOWORK=off go list -m | head -n 1)
echo "Module name: $GO_MOD_NAME"
./script/protocgen.sh go "$GO_MOD_NAME" go

echo ""
echo "=== Proto generation complete! ==="
