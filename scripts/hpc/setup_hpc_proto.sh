#!/bin/bash
# ============================================================
# HPC Proto Setup Script (Unix/WSL2)
# Copies proto files to correct location and runs buf generate
# ============================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROTO_DIR="$SCRIPT_DIR/proto"

echo "====================================================="
echo "HPC Proto Generation Setup"
echo "====================================================="
echo ""

# Step 1: Create HPC proto directories
echo "Step 1: Creating HPC proto directories..."
mkdir -p "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1"
mkdir -p "$REPO_ROOT/sdk/go/node/hpc/v1"
echo "  Created: sdk/proto/node/virtengine/hpc/v1"
echo "  Created: sdk/go/node/hpc/v1"
echo ""

# Step 2: Copy proto files
echo "Step 2: Copying proto files..."
cp "$PROTO_DIR/hpc_types.proto.txt" "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1/types.proto"
echo "  Copied: types.proto"
cp "$PROTO_DIR/hpc_tx.proto.txt" "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1/tx.proto"
echo "  Copied: tx.proto"
cp "$PROTO_DIR/hpc_query.proto.txt" "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1/query.proto"
echo "  Copied: query.proto"
cp "$PROTO_DIR/hpc_genesis.proto.txt" "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1/genesis.proto"
echo "  Copied: genesis.proto"
echo ""

# Step 3: Verify files
echo "Step 3: Verifying proto files..."
ls -la "$REPO_ROOT/sdk/proto/node/virtengine/hpc/v1/"
echo ""

# Step 4: Run proto generation (optional)
if [[ "$1" == "--generate" ]]; then
    echo "Step 4: Running proto generation..."
    cd "$REPO_ROOT/sdk"
    if command -v buf &> /dev/null; then
        buf generate
        echo "Proto generation complete!"
    else
        echo "buf not found. Running protocgen.sh..."
        ./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go
    fi
    cd ..
fi

echo ""
echo "====================================================="
echo "SUCCESS: Proto files copied to correct location"
echo "====================================================="
echo ""
echo "Proto files are now in:"
echo "  sdk/proto/node/virtengine/hpc/v1/types.proto"
echo "  sdk/proto/node/virtengine/hpc/v1/tx.proto"
echo "  sdk/proto/node/virtengine/hpc/v1/query.proto"
echo "  sdk/proto/node/virtengine/hpc/v1/genesis.proto"
echo ""
echo "Next steps:"
echo "  1. cd sdk"
echo "  2. buf generate (or run: ./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go)"
echo "  3. Generated Go files will be in: sdk/go/node/hpc/v1/"
echo ""
echo "Or run this script with --generate flag:"
echo "  ./scripts/hpc/setup_hpc_proto.sh --generate"
echo ""
echo "After generation, clean up temporary files:"
echo "  rm scripts/hpc/proto/hpc_*.proto.txt scripts/hpc/setup_hpc_proto.* scripts/hpc/setup_hpc_dirs.js scripts/hpc/create_dirs.go scripts/hpc/create_network_security_dirs.py"
echo ""
echo "====================================================="
