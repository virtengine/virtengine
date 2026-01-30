#!/usr/bin/env python3
"""Create HPC proto directory structure and copy proto files.

This script creates the HPC module proto directories and copies the proto files
from the temporary .proto.txt files to their final locations.

Usage:
    python create_network_security_dirs.py

    Or on Windows:
    python3 create_network_security_dirs.py

After running, you can generate the Go files by:
    cd sdk
    buf generate
"""
import os
import shutil
import sys

def main():
    # Create directories
    dirs = [
        "sdk/proto/node/virtengine/hpc/v1",
        "sdk/go/node/hpc/v1",
    ]

    for d in dirs:
        os.makedirs(d, exist_ok=True)
        print(f"Created: {d}")

    # Copy proto files
    proto_files = {
        "hpc_types.proto.txt": "sdk/proto/node/virtengine/hpc/v1/types.proto",
        "hpc_tx.proto.txt": "sdk/proto/node/virtengine/hpc/v1/tx.proto",
        "hpc_query.proto.txt": "sdk/proto/node/virtengine/hpc/v1/query.proto",
        "hpc_genesis.proto.txt": "sdk/proto/node/virtengine/hpc/v1/genesis.proto",
    }

    success = True
    for src, dst in proto_files.items():
        if os.path.exists(src):
            shutil.copy2(src, dst)
            print(f"Copied: {src} -> {dst}")
        else:
            print(f"Warning: Source file not found: {src}")
            success = False

    if success:
        print("\nHPC proto files created successfully!")
        print("\nCreated files:")
        proto_dir = "sdk/proto/node/virtengine/hpc/v1"
        if os.path.isdir(proto_dir):
            for f in os.listdir(proto_dir):
                print(f"  - {proto_dir}/{f}")
        print("\nNext steps:")
        print("  1. cd sdk")
        print("  2. buf generate (or ./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go)")
    else:
        print("\nSome files were not found. Please check the source files exist.")
        sys.exit(1)

if __name__ == "__main__":
    main()
