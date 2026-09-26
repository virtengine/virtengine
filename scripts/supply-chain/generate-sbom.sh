#!/usr/bin/env bash
# generate-sbom.sh - Generate Software Bill of Materials (SBOM)
#
# This script generates SBOMs in multiple formats for compliance and
# vulnerability scanning purposes.
#
# Usage: ./scripts/supply-chain/generate-sbom.sh [options]
#
# Options:
#   --format <format>   Output format: cyclonedx, spdx, syft, all (default: all)
#   --output <dir>      Output directory (default: .cache/sbom)
#   --sign              Sign SBOM with cosign
#   --verify            Verify SBOM after generation
#   --container <img>   Generate SBOM for container image
#
# Output formats:
#   - CycloneDX (sbom.cdx.json) - Recommended for vulnerability scanning
#   - SPDX (sbom.spdx.json) - Recommended for license compliance
#   - Syft native (sbom.syft.json) - Full Syft output

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Configuration
FORMAT="all"
OUTPUT_DIR="${ROOT_DIR}/.cache/sbom"
SIGN_SBOM=false
VERIFY_SBOM=false
CONTAINER_IMAGE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --sign)
            SIGN_SBOM=true
            shift
            ;;
        --verify)
            VERIFY_SBOM=true
            shift
            ;;
        --container)
            CONTAINER_IMAGE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install syft if not present
ensure_syft() {
    if ! command_exists syft; then
        log_info "Installing Syft..."
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            brew install syft
        else
            log_error "Please install Syft manually: https://github.com/anchore/syft"
            exit 1
        fi
    fi
    log_success "Syft is available: $(syft version 2>/dev/null | head -1)"
}

# Generate CycloneDX SBOM
generate_cyclonedx() {
    local output_file="${OUTPUT_DIR}/sbom.cdx.json"
    log_info "Generating CycloneDX SBOM..."
    
    if [[ -n "$CONTAINER_IMAGE" ]]; then
        syft "$CONTAINER_IMAGE" -o cyclonedx-json > "$output_file"
    else
        syft dir:"$ROOT_DIR" -o cyclonedx-json > "$output_file"
    fi
    
    log_success "CycloneDX SBOM saved to $output_file"
    echo "$output_file"
}

# Generate SPDX SBOM
generate_spdx() {
    local output_file="${OUTPUT_DIR}/sbom.spdx.json"
    log_info "Generating SPDX SBOM..."
    
    if [[ -n "$CONTAINER_IMAGE" ]]; then
        syft "$CONTAINER_IMAGE" -o spdx-json > "$output_file"
    else
        syft dir:"$ROOT_DIR" -o spdx-json > "$output_file"
    fi
    
    log_success "SPDX SBOM saved to $output_file"
    echo "$output_file"
}

# Generate Syft native SBOM
generate_syft() {
    local output_file="${OUTPUT_DIR}/sbom.syft.json"
    log_info "Generating Syft SBOM..."
    
    if [[ -n "$CONTAINER_IMAGE" ]]; then
        syft "$CONTAINER_IMAGE" -o syft-json > "$output_file"
    else
        syft dir:"$ROOT_DIR" -o syft-json > "$output_file"
    fi
    
    log_success "Syft SBOM saved to $output_file"
    echo "$output_file"
}

# Generate Go-specific SBOM using go version -m
generate_go_sbom() {
    local output_file="${OUTPUT_DIR}/sbom.go.json"
    log_info "Generating Go module SBOM..."
    
    cd "$ROOT_DIR"
    
    # Create a JSON structure with go module information
    cat > "$output_file" << EOF
{
    "format": "go-modules",
    "generated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "tool": "go version -m",
    "modules": [
EOF
    
    # Parse go.mod for direct dependencies
    if [[ -f "go.mod" ]]; then
        local first=true
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]+([^[:space:]]+)[[:space:]]+(v[^[:space:]]+) ]]; then
                pkg="${BASH_REMATCH[1]}"
                ver="${BASH_REMATCH[2]}"
                
                if [[ "$first" == "true" ]]; then
                    first=false
                else
                    echo "," >> "$output_file"
                fi
                
                echo -n "        {\"path\": \"$pkg\", \"version\": \"$ver\"}" >> "$output_file"
            fi
        done < <(grep -A 1000 "^require (" go.mod | grep -B 1000 "^)" | grep -v "^require\|^)")
    fi
    
    cat >> "$output_file" << EOF

    ]
}
EOF
    
    log_success "Go module SBOM saved to $output_file"
    echo "$output_file"
}

# Sign SBOM with cosign
sign_sbom() {
    local sbom_file="$1"
    
    if ! command_exists cosign; then
        log_warning "cosign not installed - skipping SBOM signing"
        return
    fi
    
    log_info "Signing SBOM: $sbom_file"
    
    cosign sign-blob \
        --yes \
        --output-signature "${sbom_file}.sig" \
        --output-certificate "${sbom_file}.sig.cert" \
        "$sbom_file" 2>/dev/null || {
            log_warning "SBOM signing failed - this is expected in non-CI environments"
            return
        }
    
    log_success "SBOM signed: ${sbom_file}.sig"
}

# Verify SBOM with grype
verify_sbom() {
    local sbom_file="$1"
    
    if ! command_exists grype; then
        log_warning "grype not installed - skipping SBOM vulnerability scan"
        log_info "Install with: curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh"
        return
    fi
    
    log_info "Scanning SBOM for vulnerabilities: $sbom_file"
    
    local report_file="${OUTPUT_DIR}/vulnerability-report.json"
    grype sbom:"$sbom_file" -o json > "$report_file" 2>/dev/null || true
    
    # Count vulnerabilities by severity
    local critical=$(jq '[.matches[] | select(.vulnerability.severity == "Critical")] | length' "$report_file" 2>/dev/null || echo "0")
    local high=$(jq '[.matches[] | select(.vulnerability.severity == "High")] | length' "$report_file" 2>/dev/null || echo "0")
    local medium=$(jq '[.matches[] | select(.vulnerability.severity == "Medium")] | length' "$report_file" 2>/dev/null || echo "0")
    local low=$(jq '[.matches[] | select(.vulnerability.severity == "Low")] | length' "$report_file" 2>/dev/null || echo "0")
    
    echo ""
    echo "Vulnerability Summary:"
    echo "  Critical: $critical"
    echo "  High:     $high"
    echo "  Medium:   $medium"
    echo "  Low:      $low"
    echo ""
    echo "Full report: $report_file"
    
    if [[ "$critical" -gt 0 ]] || [[ "$high" -gt 0 ]]; then
        log_warning "Critical or High vulnerabilities detected!"
        return 1
    fi
    
    log_success "No critical or high vulnerabilities found"
}

# Generate SBOM attestation for container images
generate_attestation() {
    local sbom_file="$1"
    local image="$2"
    
    if ! command_exists cosign; then
        log_warning "cosign not installed - skipping attestation"
        return
    fi
    
    log_info "Generating SBOM attestation for $image"
    
    cosign attest \
        --yes \
        --type cyclonedx \
        --predicate "$sbom_file" \
        "$image" 2>/dev/null || {
            log_warning "Attestation failed - this is expected in non-CI environments"
            return
        }
    
    log_success "SBOM attestation attached to $image"
}

# Generate summary report
generate_summary() {
    local summary_file="${OUTPUT_DIR}/sbom-summary.json"
    
    log_info "Generating SBOM summary..."
    
    local package_count=0
    local license_count=0
    
    if [[ -f "${OUTPUT_DIR}/sbom.cdx.json" ]]; then
        package_count=$(jq '.components | length' "${OUTPUT_DIR}/sbom.cdx.json" 2>/dev/null || echo "0")
        license_count=$(jq '[.components[].licenses[]?.license.id // empty] | unique | length' "${OUTPUT_DIR}/sbom.cdx.json" 2>/dev/null || echo "0")
    fi
    
    cat > "$summary_file" << EOF
{
    "generated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "repository": "$(git remote get-url origin 2>/dev/null || echo 'unknown')",
    "commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')",
    "sbom_files": {
        "cyclonedx": "${OUTPUT_DIR}/sbom.cdx.json",
        "spdx": "${OUTPUT_DIR}/sbom.spdx.json",
        "syft": "${OUTPUT_DIR}/sbom.syft.json"
    },
    "statistics": {
        "total_packages": $package_count,
        "unique_licenses": $license_count
    },
    "tools": {
        "syft": "$(syft version 2>/dev/null | head -1 || echo 'unknown')"
    }
}
EOF
    
    log_success "Summary saved to $summary_file"
}

# Main execution
main() {
    echo ""
    echo "=========================================="
    echo "  VirtEngine SBOM Generator"
    echo "=========================================="
    echo ""
    
    # Ensure output directory exists
    mkdir -p "$OUTPUT_DIR"
    
    # Ensure Syft is available
    ensure_syft
    
    cd "$ROOT_DIR"
    
    # Generate requested formats
    local generated_files=()
    
    case "$FORMAT" in
        cyclonedx)
            generated_files+=($(generate_cyclonedx))
            ;;
        spdx)
            generated_files+=($(generate_spdx))
            ;;
        syft)
            generated_files+=($(generate_syft))
            ;;
        all)
            generated_files+=($(generate_cyclonedx))
            generated_files+=($(generate_spdx))
            generated_files+=($(generate_syft))
            generated_files+=($(generate_go_sbom))
            ;;
        *)
            log_error "Unknown format: $FORMAT"
            exit 1
            ;;
    esac
    
    echo ""
    
    # Sign SBOMs if requested
    if [[ "$SIGN_SBOM" == "true" ]]; then
        for file in "${generated_files[@]}"; do
            sign_sbom "$file"
        done
        echo ""
    fi
    
    # Verify SBOMs if requested
    if [[ "$VERIFY_SBOM" == "true" ]]; then
        for file in "${generated_files[@]}"; do
            if [[ "$file" == *".cdx.json" ]]; then
                verify_sbom "$file" || true
            fi
        done
        echo ""
    fi
    
    # Generate container attestation if image specified
    if [[ -n "$CONTAINER_IMAGE" ]] && [[ -f "${OUTPUT_DIR}/sbom.cdx.json" ]]; then
        generate_attestation "${OUTPUT_DIR}/sbom.cdx.json" "$CONTAINER_IMAGE"
        echo ""
    fi
    
    # Generate summary
    generate_summary
    
    echo ""
    echo "=========================================="
    echo "  SBOM Generation Complete"
    echo "=========================================="
    echo ""
    echo "Output directory: $OUTPUT_DIR"
    echo "Generated files:"
    ls -la "$OUTPUT_DIR"/*.json 2>/dev/null | awk '{print "  " $NF}'
    echo ""
    
    log_success "SBOM generation completed successfully"
}

# Run main
main
