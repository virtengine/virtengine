#!/usr/bin/env bash
# Copyright 2024 VirtEngine Authors
# SPDX-License-Identifier: Apache-2.0
#
# compute_model_hash.sh — Compute deterministic SHA256 hashes of ML model files
# for VEID governance. Only hashes the frozen graph / model weights, excluding
# metadata files (checksums, manifests, READMEs) to ensure hash stability.
#
# Usage:
#   ./scripts/compute_model_hash.sh <model_dir> [model_type]
#   ./scripts/compute_model_hash.sh ml/facial_verification/weights face_verification
#   ./scripts/compute_model_hash.sh ml/liveness_detection/weights liveness
#
# Output: JSON with hash, version, timestamp on stdout.

set -euo pipefail

# Model file extensions to include in hash computation (frozen graphs / weights)
MODEL_EXTENSIONS="pb|h5|tflite|onnx|savedmodel|pt|pth|bin|safetensors"

# Files to exclude from hash computation (metadata, not model weights)
EXCLUDE_PATTERNS="README|LICENSE|CHANGELOG|manifest|checksum|metadata|\.txt$|\.md$|\.json$|\.yaml$|\.yml$"

usage() {
    echo "Usage: $0 <model_dir> [model_type]" >&2
    echo "" >&2
    echo "Arguments:" >&2
    echo "  model_dir    Path to the directory containing model files" >&2
    echo "  model_type   Optional model type (e.g., face_verification, liveness, ocr)" >&2
    echo "               Defaults to directory name if not specified" >&2
    echo "" >&2
    echo "Output: JSON object with hash, model_type, version, timestamp" >&2
    exit 1
}

if [ $# -lt 1 ]; then
    usage
fi

MODEL_DIR="$1"
MODEL_TYPE="${2:-$(basename "$MODEL_DIR")}"

if [ ! -d "$MODEL_DIR" ]; then
    echo "{\"error\": \"directory not found: $MODEL_DIR\"}" >&2
    exit 1
fi

# Find model files, sorted for deterministic ordering
MODEL_FILES=$(find "$MODEL_DIR" -type f \
    | grep -iE "\.(${MODEL_EXTENSIONS})$" \
    | grep -ivE "${EXCLUDE_PATTERNS}" \
    | LC_ALL=C sort)

if [ -z "$MODEL_FILES" ]; then
    echo "{\"error\": \"no model files found in $MODEL_DIR\"}" >&2
    exit 1
fi

# Compute combined SHA256: hash each file in sorted order, then hash the hashes
COMBINED=""
FILE_COUNT=0
for f in $MODEL_FILES; do
    FILE_HASH=$(sha256sum "$f" | awk '{print $1}')
    COMBINED="${COMBINED}${FILE_HASH}"
    FILE_COUNT=$((FILE_COUNT + 1))
done

# Final deterministic hash of all file hashes concatenated
FINAL_HASH=$(printf '%s' "$COMBINED" | sha256sum | awk '{print $1}')

# Extract version from directory structure or default
VERSION="unknown"
if [ -f "$MODEL_DIR/version.txt" ]; then
    VERSION=$(cat "$MODEL_DIR/version.txt" | tr -d '[:space:]')
elif [ -f "$MODEL_DIR/../version.txt" ]; then
    VERSION=$(cat "$MODEL_DIR/../version.txt" | tr -d '[:space:]')
fi

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat <<EOF
{
  "sha256_hash": "${FINAL_HASH}",
  "model_type": "${MODEL_TYPE}",
  "version": "${VERSION}",
  "file_count": ${FILE_COUNT},
  "timestamp": "${TIMESTAMP}",
  "directory": "${MODEL_DIR}"
}
EOF
