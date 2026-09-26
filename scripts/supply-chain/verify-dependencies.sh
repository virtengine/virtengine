#!/usr/bin/env bash
# verify-dependencies.sh - Verify dependency integrity and security
#
# This script performs comprehensive verification of all project dependencies
# including checksum verification, lockfile integrity, and vulnerability scanning.
#
# Usage: ./scripts/supply-chain/verify-dependencies.sh [options]
#
# Options:
#   --go          Verify Go dependencies only
#   --python      Verify Python dependencies only
#   --npm         Verify npm dependencies only
#   --all         Verify all dependencies (default)
#   --strict      Exit with error on any warning
#   --verbose     Enable verbose output
#
# Exit codes:
#   0 - All verifications passed
#   1 - Verification failed (security issue)
#   2 - Verification warning (non-blocking)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Configuration
STRICT_MODE=false
VERBOSE=false
VERIFY_GO=false
VERIFY_PYTHON=false
VERIFY_NPM=false
VERIFY_ALL=true
EXIT_CODE=0

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --go)
            VERIFY_GO=true
            VERIFY_ALL=false
            shift
            ;;
        --python)
            VERIFY_PYTHON=true
            VERIFY_ALL=false
            shift
            ;;
        --npm)
            VERIFY_NPM=true
            VERIFY_ALL=false
            shift
            ;;
        --all)
            VERIFY_ALL=true
            shift
            ;;
        --strict)
            STRICT_MODE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
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
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    if [[ "$STRICT_MODE" == "true" ]]; then
        EXIT_CODE=2
    fi
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    EXIT_CODE=1
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Verify Go dependencies
verify_go_deps() {
    log_info "Verifying Go dependencies..."
    
    cd "$ROOT_DIR"
    
    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod not found"
        return 1
    fi
    
    # Check if go.sum exists
    if [[ ! -f "go.sum" ]]; then
        log_error "go.sum not found - dependencies not locked"
        return 1
    fi
    
    # Verify module checksums
    log_info "Verifying module checksums..."
    if go mod verify 2>/dev/null; then
        log_success "Go module checksums verified"
    else
        log_error "Go module checksum verification failed"
        return 1
    fi
    
    # Check for version ranges in go.mod (should be exact versions)
    log_info "Checking for exact version pinning..."
    if grep -E "^\s*require\s+.+\s+v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$" go.mod >/dev/null 2>&1; then
        log_verbose "Found properly pinned versions"
    fi
    
    # Check for unpinned versions (like "latest" or ranges)
    if grep -E "^\s*(//\s*)?.+\s+(latest|>=|<=|>|<|~|\^)" go.mod 2>/dev/null | grep -v "^//" | head -5; then
        log_warning "Found potentially unpinned dependency versions in go.mod"
    else
        log_success "All Go dependencies appear to be pinned to exact versions"
    fi
    
    # Check go.mod is tidy
    log_info "Checking go.mod is tidy..."
    cp go.mod go.mod.backup
    cp go.sum go.sum.backup
    
    if go mod tidy 2>/dev/null; then
        if diff -q go.mod go.mod.backup >/dev/null 2>&1 && \
           diff -q go.sum go.sum.backup >/dev/null 2>&1; then
            log_success "go.mod and go.sum are tidy"
        else
            log_warning "go.mod or go.sum would change after 'go mod tidy'"
            # Restore original files
            mv go.mod.backup go.mod
            mv go.sum.backup go.sum
        fi
    else
        log_error "Failed to run 'go mod tidy'"
        mv go.mod.backup go.mod 2>/dev/null || true
        mv go.sum.backup go.sum 2>/dev/null || true
    fi
    rm -f go.mod.backup go.sum.backup
    
    # Run govulncheck if available
    if command_exists govulncheck; then
        log_info "Running govulncheck..."
        if govulncheck ./... 2>/dev/null; then
            log_success "No known vulnerabilities found in Go dependencies"
        else
            log_warning "govulncheck found potential vulnerabilities"
        fi
    else
        log_warning "govulncheck not installed - skipping vulnerability check"
        log_verbose "Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
    fi
    
    # Verify vendor directory if it exists
    if [[ -d "vendor" ]]; then
        log_info "Verifying vendor directory..."
        if go mod vendor 2>/dev/null && git diff --quiet vendor/ 2>/dev/null; then
            log_success "Vendor directory is up to date"
        else
            log_warning "Vendor directory may be out of sync"
        fi
    fi
    
    return 0
}

# Verify Python dependencies
verify_python_deps() {
    log_info "Verifying Python dependencies..."
    
    local python_dirs=("$ROOT_DIR/ml" "$ROOT_DIR/tests/python" "$ROOT_DIR/sdk/python")
    
    for dir in "${python_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            log_info "Checking Python dependencies in $dir..."
            
            # Check for requirements.txt
            if [[ -f "$dir/requirements.txt" ]]; then
                log_verbose "Found requirements.txt in $dir"
                
                # Check for exact version pinning
                log_info "Checking for exact version pinning in $dir/requirements.txt..."
                local unpinned=$(grep -E "^[^#].*[><=~!]" "$dir/requirements.txt" | grep -v "==" | head -5 || true)
                if [[ -n "$unpinned" ]]; then
                    log_warning "Found unpinned or range versions in $dir/requirements.txt:"
                    echo "$unpinned"
                else
                    log_success "All Python dependencies in $dir are pinned to exact versions"
                fi
                
                # Check for hash pinning
                if grep -q "sha256:" "$dir/requirements.txt" 2>/dev/null; then
                    log_success "Hash pinning found in $dir/requirements.txt"
                else
                    log_warning "No hash pinning in $dir/requirements.txt - consider using pip-compile"
                fi
                
                # Run pip-audit if available
                if command_exists pip-audit; then
                    log_info "Running pip-audit on $dir/requirements.txt..."
                    if pip-audit -r "$dir/requirements.txt" 2>/dev/null; then
                        log_success "No known vulnerabilities in $dir Python dependencies"
                    else
                        log_warning "pip-audit found potential vulnerabilities in $dir"
                    fi
                else
                    log_verbose "pip-audit not installed - skipping Python vulnerability check"
                fi
            fi
            
            # Check for requirements-deterministic.txt (ML specific)
            if [[ -f "$dir/requirements-deterministic.txt" ]]; then
                log_info "Found deterministic requirements file in $dir"
                if grep -q "tensorflow" "$dir/requirements-deterministic.txt" 2>/dev/null; then
                    log_success "TensorFlow deterministic requirements present"
                fi
            fi
        fi
    done
    
    return 0
}

# Verify npm dependencies
verify_npm_deps() {
    log_info "Verifying npm dependencies..."
    
    local npm_dirs=("$ROOT_DIR/lib/portal" "$ROOT_DIR/sdk/js")
    
    for dir in "${npm_dirs[@]}"; do
        if [[ -d "$dir" ]] && [[ -f "$dir/package.json" ]]; then
            log_info "Checking npm dependencies in $dir..."
            
            cd "$dir"
            
            # Check for package-lock.json
            if [[ ! -f "package-lock.json" ]]; then
                log_error "package-lock.json not found in $dir - dependencies not locked"
                continue
            fi
            
            log_success "package-lock.json found in $dir"
            
            # Check lockfileVersion
            local lockfile_version=$(grep -o '"lockfileVersion": [0-9]*' package-lock.json | grep -o '[0-9]*' || echo "1")
            if [[ "$lockfile_version" -ge 2 ]]; then
                log_success "Using modern lockfile version ($lockfile_version)"
            else
                log_warning "Consider upgrading to lockfileVersion 2+ for better security"
            fi
            
            # Run npm audit if npm is available
            if command_exists npm; then
                log_info "Running npm audit..."
                if npm audit --audit-level=high 2>/dev/null; then
                    log_success "No high/critical vulnerabilities in $dir npm dependencies"
                else
                    log_warning "npm audit found high/critical vulnerabilities in $dir"
                fi
            else
                log_warning "npm not installed - skipping npm vulnerability check"
            fi
            
            cd "$ROOT_DIR"
        fi
    done
    
    return 0
}

# Verify dependency sources
verify_dependency_sources() {
    log_info "Verifying dependency sources..."
    
    cd "$ROOT_DIR"
    
    # Check for suspicious sources in go.mod
    if [[ -f "go.mod" ]]; then
        log_info "Checking Go module sources..."
        
        # Check for replace directives pointing to non-official sources
        local suspicious_replaces=$(grep -E "^\s*replace\s+" go.mod | \
            grep -vE "(github\.com/(virtengine|cosmos|akash-network|regen-network)|golang\.org)" | \
            head -5 || true)
        
        if [[ -n "$suspicious_replaces" ]]; then
            log_warning "Found non-standard replace directives in go.mod:"
            echo "$suspicious_replaces"
        else
            log_success "All Go module sources appear to be from trusted providers"
        fi
    fi
    
    return 0
}

# Generate verification report
generate_report() {
    log_info "Generating verification report..."
    
    local report_file="$ROOT_DIR/.cache/dependency-verification-report.json"
    mkdir -p "$(dirname "$report_file")"
    
    cat > "$report_file" << EOF
{
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "git_branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')",
    "verification_result": "$([[ $EXIT_CODE -eq 0 ]] && echo 'passed' || echo 'failed')",
    "exit_code": $EXIT_CODE,
    "strict_mode": $STRICT_MODE,
    "verified_ecosystems": {
        "go": $([[ "$VERIFY_GO" == "true" || "$VERIFY_ALL" == "true" ]] && echo 'true' || echo 'false'),
        "python": $([[ "$VERIFY_PYTHON" == "true" || "$VERIFY_ALL" == "true" ]] && echo 'true' || echo 'false'),
        "npm": $([[ "$VERIFY_NPM" == "true" || "$VERIFY_ALL" == "true" ]] && echo 'true' || echo 'false')
    }
}
EOF
    
    log_success "Verification report saved to $report_file"
}

# Main execution
main() {
    echo ""
    echo "=========================================="
    echo "  VirtEngine Dependency Verification"
    echo "=========================================="
    echo ""
    
    # Verify selected ecosystems
    if [[ "$VERIFY_ALL" == "true" ]] || [[ "$VERIFY_GO" == "true" ]]; then
        verify_go_deps || true
        echo ""
    fi
    
    if [[ "$VERIFY_ALL" == "true" ]] || [[ "$VERIFY_PYTHON" == "true" ]]; then
        verify_python_deps || true
        echo ""
    fi
    
    if [[ "$VERIFY_ALL" == "true" ]] || [[ "$VERIFY_NPM" == "true" ]]; then
        verify_npm_deps || true
        echo ""
    fi
    
    # Always verify dependency sources
    verify_dependency_sources || true
    echo ""
    
    # Generate report
    generate_report
    
    # Summary
    echo ""
    echo "=========================================="
    echo "  Verification Summary"
    echo "=========================================="
    
    if [[ $EXIT_CODE -eq 0 ]]; then
        log_success "All dependency verifications passed"
    elif [[ $EXIT_CODE -eq 2 ]]; then
        log_warning "Dependency verification completed with warnings"
    else
        log_error "Dependency verification failed"
    fi
    
    exit $EXIT_CODE
}

# Run main
main
