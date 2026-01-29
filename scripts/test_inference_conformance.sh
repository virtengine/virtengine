#!/bin/bash
# Test Go-Python inference conformance
# Run both implementations on same inputs and compare outputs
#
# Task Reference: VE-3006 - Go-Python Conformance Testing
#
# Usage:
#   ./scripts/test_inference_conformance.sh
#   ./scripts/test_inference_conformance.sh --generate-only
#   ./scripts/test_inference_conformance.sh --verify path/to/go_outputs.json

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Paths
ML_CONFORMANCE_DIR="${PROJECT_ROOT}/ml/conformance"
PKG_INFERENCE_DIR="${PROJECT_ROOT}/pkg/inference"
TEST_VECTORS_FILE="${ML_CONFORMANCE_DIR}/test_vectors.json"
GO_OUTPUTS_FILE="${PROJECT_ROOT}/tmp/go_inference_outputs.json"

# Functions
print_header() {
    echo -e "\n${BLUE}============================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check Python is available
check_python() {
    if command -v python3 &> /dev/null; then
        PYTHON_CMD="python3"
    elif command -v python &> /dev/null; then
        PYTHON_CMD="python"
    else
        print_error "Python not found. Please install Python 3.8+"
        exit 1
    fi
    
    print_info "Using Python: ${PYTHON_CMD}"
}

# Check Go is available
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go not found. Please install Go 1.21+"
        exit 1
    fi
    
    print_info "Using Go: $(go version)"
}

# Generate Python test vectors
generate_test_vectors() {
    print_header "Generating Python Test Vectors"
    
    cd "${PROJECT_ROOT}"
    
    # Ensure output directory exists
    mkdir -p "${ML_CONFORMANCE_DIR}"
    
    # Run generation script
    ${PYTHON_CMD} -m ml.conformance.generate_test_vectors
    
    if [[ -f "${TEST_VECTORS_FILE}" ]]; then
        VECTOR_COUNT=$(${PYTHON_CMD} -c "import json; print(len(json.load(open('${TEST_VECTORS_FILE}'))['vectors']))")
        print_success "Generated ${VECTOR_COUNT} test vectors to ${TEST_VECTORS_FILE}"
    else
        print_error "Failed to generate test vectors"
        exit 1
    fi
}

# Run Go conformance tests
run_go_tests() {
    print_header "Running Go Conformance Tests"
    
    cd "${PROJECT_ROOT}"
    
    # Run tests with verbose output
    echo -e "\nRunning: go test -v ./pkg/inference/... -run Conformance"
    echo ""
    
    if go test -v ./pkg/inference/... -run "Test.*Conformance\|Test.*Determinism\|Test.*Match"; then
        print_success "Go conformance tests passed"
    else
        print_error "Go conformance tests failed"
        exit 1
    fi
}

# Run Go benchmark tests
run_go_benchmarks() {
    print_header "Running Go Benchmark Tests"
    
    cd "${PROJECT_ROOT}"
    
    echo -e "\nRunning: go test -bench=. ./pkg/inference/... -benchmem -run=^$"
    echo ""
    
    go test -bench=. ./pkg/inference/... -benchmem -run=^$ || true
    
    print_success "Benchmarks complete"
}

# Verify Go outputs against Python expectations
verify_go_outputs() {
    local go_outputs_file="${1:-${GO_OUTPUTS_FILE}}"
    
    print_header "Verifying Go Outputs Against Python"
    
    if [[ ! -f "${go_outputs_file}" ]]; then
        print_error "Go outputs file not found: ${go_outputs_file}"
        print_info "Generate Go outputs first by running inference tests"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
    
    ${PYTHON_CMD} -m ml.conformance.verify_go_output "${go_outputs_file}" --verbose
    
    if [[ $? -eq 0 ]]; then
        print_success "All outputs match!"
    else
        print_error "Output verification failed"
        exit 1
    fi
}

# Print summary
print_summary() {
    print_header "Conformance Test Summary"
    
    echo "Test Vectors: ${TEST_VECTORS_FILE}"
    
    if [[ -f "${TEST_VECTORS_FILE}" ]]; then
        VECTOR_COUNT=$(${PYTHON_CMD} -c "import json; print(len(json.load(open('${TEST_VECTORS_FILE}'))['vectors']))")
        echo "Vector Count: ${VECTOR_COUNT}"
        
        VECTORS_HASH=$(${PYTHON_CMD} -c "import json; print(json.load(open('${TEST_VECTORS_FILE}'))['vectors_hash'][:16])...")
        echo "Vectors Hash: ${VECTORS_HASH}"
    fi
    
    echo ""
    print_success "Conformance testing complete!"
}

# Main entry point
main() {
    print_header "VE-3006: Go-Python Inference Conformance Testing"
    
    # Check dependencies
    check_python
    check_go
    
    # Parse arguments
    GENERATE_ONLY=false
    VERIFY_FILE=""
    BENCHMARKS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --generate-only)
                GENERATE_ONLY=true
                shift
                ;;
            --verify)
                VERIFY_FILE="${2:-}"
                shift 2
                ;;
            --benchmarks)
                BENCHMARKS=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --generate-only    Only generate test vectors, don't run tests"
                echo "  --verify <file>    Verify Go outputs from specified JSON file"
                echo "  --benchmarks       Run performance benchmarks"
                echo "  --help             Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Run workflow
    if [[ -n "${VERIFY_FILE}" ]]; then
        verify_go_outputs "${VERIFY_FILE}"
    elif [[ "${GENERATE_ONLY}" == "true" ]]; then
        generate_test_vectors
    else
        generate_test_vectors
        run_go_tests
        
        if [[ "${BENCHMARKS}" == "true" ]]; then
            run_go_benchmarks
        fi
        
        print_summary
    fi
}

main "$@"
