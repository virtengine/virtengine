#!/usr/bin/env bash
set -euo pipefail

OUTPUT_DIR="${1:-./sim-output}"
SCENARIO="${SCENARIO:-baseline}"
RUNS="${RUNS:-50}"

echo "Running economics simulation suite (${SCENARIO})..."
go run ./cmd/ve-sim suite --scenario "${SCENARIO}" --output-dir "${OUTPUT_DIR}" --runs "${RUNS}"

echo "Validating economics metrics..."
go run ./cmd/ve-sim check --metrics "${OUTPUT_DIR}/metrics.json"

echo "Suite completed. Outputs in ${OUTPUT_DIR}"

