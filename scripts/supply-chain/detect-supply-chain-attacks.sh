#!/usr/bin/env bash
# detect-supply-chain-attacks.sh - Detect potential supply chain attack vectors
#
# This script checks for common supply chain attack patterns including:
# - Dependency confusion
# - Typosquatting
# - Suspicious package changes
# - Compromised maintainer signals
#
# Usage: ./scripts/supply-chain/detect-supply-chain-attacks.sh [options]
#
# Options:
#   --confusion       Check for dependency confusion vulnerabilities
#   --typosquatting   Check for typosquatting risks
#   --changes         Check for suspicious package changes
#   --all             Run all checks (default)
#   --json            Output results as JSON
#   --verbose         Enable verbose output
#
# Exit codes:
#   0 - No issues detected
#   1 - Potential supply chain attack detected (critical)
#   2 - Warning level issues detected

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Configuration
CHECK_CONFUSION=false
CHECK_TYPOSQUATTING=false
CHECK_CHANGES=false
CHECK_ALL=true
OUTPUT_JSON=false
VERBOSE=false
EXIT_CODE=0

# Known internal package prefixes (should not exist in public registries)
INTERNAL_PREFIXES=(
    "github.com/virtengine/internal"
    "github.com/virtengine/private"
    "@virtengine/internal"
    "@virtengine/private"
)

# Known typosquatting patterns
TYPOSQUAT_PATTERNS=(
    # Common Cosmos SDK typosquats
    "cosmos-sdk|cosmoss-sdk|cosmos-skd|cosmod-sdk"
    "cometbft|commetbft|cometbf|comebtft"
    # Common crypto typosquats
    "golang.org/x/crypto|golang.org/x/crypt|golang.org/x/cyrpto"
    # Common testing typosquats
    "testify|testfy|testyfy|testifi"
)

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --confusion)
            CHECK_CONFUSION=true
            CHECK_ALL=false
            shift
            ;;
        --typosquatting)
            CHECK_TYPOSQUATTING=true
            CHECK_ALL=false
            shift
            ;;
        --changes)
            CHECK_CHANGES=true
            CHECK_ALL=false
            shift
            ;;
        --all)
            CHECK_ALL=true
            shift
            ;;
        --json)
            OUTPUT_JSON=true
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
    if [[ "$OUTPUT_JSON" != "true" ]]; then
        echo -e "${BLUE}[INFO]${NC} $1"
    fi
}

log_success() {
    if [[ "$OUTPUT_JSON" != "true" ]]; then
        echo -e "${GREEN}[SAFE]${NC} $1"
    fi
}

log_warning() {
    if [[ "$OUTPUT_JSON" != "true" ]]; then
        echo -e "${YELLOW}[WARN]${NC} $1"
    fi
    EXIT_CODE=$((EXIT_CODE > 2 ? EXIT_CODE : 2))
}

log_critical() {
    if [[ "$OUTPUT_JSON" != "true" ]]; then
        echo -e "${RED}[CRITICAL]${NC} $1"
    fi
    EXIT_CODE=1
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]] && [[ "$OUTPUT_JSON" != "true" ]]; then
        echo -e "${MAGENTA}[DEBUG]${NC} $1"
    fi
}

# Initialize JSON output
declare -a JSON_FINDINGS=()

add_finding() {
    local severity="$1"
    local category="$2"
    local message="$3"
    local details="${4:-}"
    
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        JSON_FINDINGS+=("{\"severity\": \"$severity\", \"category\": \"$category\", \"message\": \"$message\", \"details\": \"$details\"}")
    fi
}

# Check for dependency confusion vulnerabilities
check_dependency_confusion() {
    log_info "Checking for dependency confusion vulnerabilities..."
    
    cd "$ROOT_DIR"
    local findings=0
    
    # Check Go modules for internal-looking packages that might be confused
    if [[ -f "go.mod" ]]; then
        log_verbose "Analyzing go.mod for internal package patterns..."
        
        for prefix in "${INTERNAL_PREFIXES[@]}"; do
            if grep -q "$prefix" go.mod 2>/dev/null; then
                log_critical "Found internal package prefix in go.mod: $prefix"
                log_critical "This could be vulnerable to dependency confusion if not properly scoped"
                add_finding "critical" "dependency-confusion" "Internal package prefix found" "$prefix"
                ((findings++))
            fi
        done
        
        # Check for packages without proper namespacing
        local unscoped=$(grep -E "^\s+[a-zA-Z0-9_-]+\s+v" go.mod | grep -v "github.com\|golang.org\|google.golang.org\|gopkg.in\|k8s.io" || true)
        if [[ -n "$unscoped" ]]; then
            log_warning "Found packages without explicit namespace (potential confusion risk):"
            echo "$unscoped" | head -5
            add_finding "warning" "dependency-confusion" "Unscoped packages found" "$(echo "$unscoped" | head -5 | tr '\n' ' ')"
            ((findings++))
        fi
    fi
    
    # Check npm packages for scope confusion
    for package_json in $(find "$ROOT_DIR" -name "package.json" -not -path "*/node_modules/*" 2>/dev/null); do
        local dir=$(dirname "$package_json")
        log_verbose "Checking $package_json..."
        
        # Look for dependencies that should be scoped but aren't
        if grep -q '"@virtengine/' "$package_json" 2>/dev/null; then
            # Check if there are also unscoped virtengine packages
            if grep -E '"virtengine-' "$package_json" 2>/dev/null; then
                log_warning "Mixed scoped and unscoped VirtEngine packages in $package_json"
                add_finding "warning" "dependency-confusion" "Mixed package scoping" "$package_json"
                ((findings++))
            fi
        fi
    done
    
    if [[ $findings -eq 0 ]]; then
        log_success "No dependency confusion vulnerabilities detected"
    fi
    
    return 0
}

# Check for typosquatting risks
check_typosquatting() {
    log_info "Checking for typosquatting risks..."
    
    cd "$ROOT_DIR"
    local findings=0
    
    if [[ -f "go.mod" ]]; then
        log_verbose "Analyzing go.mod for known typosquatting patterns..."
        
        # Extract all dependency names
        local deps=$(grep -E "^\s+[a-zA-Z]" go.mod | awk '{print $1}' | sort -u)
        
        # Check against known legitimate packages and their typosquats
        for pattern in "${TYPOSQUAT_PATTERNS[@]}"; do
            IFS='|' read -ra variants <<< "$pattern"
            local legitimate="${variants[0]}"
            
            for i in "${!variants[@]}"; do
                if [[ $i -eq 0 ]]; then
                    continue  # Skip the legitimate package
                fi
                local typosquat="${variants[$i]}"
                
                if echo "$deps" | grep -qi "$typosquat" 2>/dev/null; then
                    log_critical "Potential typosquat detected: '$typosquat' (expected: '$legitimate')"
                    add_finding "critical" "typosquatting" "Potential typosquat package" "$typosquat -> $legitimate"
                    ((findings++))
                fi
            done
        done
        
        # Check for suspicious single-character variations
        local common_packages=("cosmos-sdk" "cometbft" "ibc-go" "grpc" "protobuf")
        for pkg in "${common_packages[@]}"; do
            # Look for packages with similar names but different by 1-2 characters
            local similar=$(echo "$deps" | grep -i "$(echo "$pkg" | sed 's/./&./g' | sed 's/\.$//g')" | head -5 || true)
            if [[ -n "$similar" ]] && [[ "$(echo "$similar" | wc -l)" -gt 1 ]]; then
                log_verbose "Multiple similar packages found for '$pkg': $similar"
            fi
        done
    fi
    
    # Check npm packages
    for package_json in $(find "$ROOT_DIR" -name "package.json" -not -path "*/node_modules/*" 2>/dev/null); do
        log_verbose "Checking npm packages in $package_json..."
        
        # Common npm typosquats for blockchain projects
        local npm_typosquats=(
            "ethers|eters|etehrs"
            "cosmjs|cosmsj|cosmj"
            "@cosmjs|@cosmsj"
        )
        
        local content=$(cat "$package_json" 2>/dev/null)
        for pattern in "${npm_typosquats[@]}"; do
            IFS='|' read -ra variants <<< "$pattern"
            local legitimate="${variants[0]}"
            
            for i in "${!variants[@]}"; do
                if [[ $i -eq 0 ]]; then continue; fi
                local typosquat="${variants[$i]}"
                
                if echo "$content" | grep -qi "\"$typosquat\"" 2>/dev/null; then
                    log_critical "Potential npm typosquat detected: '$typosquat' (expected: '$legitimate')"
                    add_finding "critical" "typosquatting" "Potential npm typosquat" "$typosquat -> $legitimate"
                    ((findings++))
                fi
            done
        done
    done
    
    if [[ $findings -eq 0 ]]; then
        log_success "No typosquatting risks detected"
    fi
    
    return 0
}

# Check for suspicious package changes
check_suspicious_changes() {
    log_info "Checking for suspicious package changes..."
    
    cd "$ROOT_DIR"
    local findings=0
    
    # Check if go.sum has unexpected changes
    if [[ -f "go.sum" ]]; then
        log_verbose "Analyzing go.sum..."
        
        # Count entries per package (multiple entries might indicate version conflicts)
        local multi_version=$(awk '{print $1}' go.sum | sort | uniq -c | sort -rn | awk '$1 > 4 {print $2, $1}' | head -10)
        if [[ -n "$multi_version" ]]; then
            log_warning "Packages with many version entries in go.sum (may indicate churn):"
            echo "$multi_version"
            add_finding "warning" "suspicious-changes" "High version churn detected" "$(echo "$multi_version" | tr '\n' ' ')"
            ((findings++))
        fi
        
        # Check for very recent packages (could indicate malicious additions)
        # This checks for packages added in the last git commit
        if command -v git &>/dev/null; then
            local recent_changes=$(git diff HEAD~1 -- go.sum 2>/dev/null | grep "^+" | grep -v "^+++" | head -10 || true)
            if [[ -n "$recent_changes" ]]; then
                log_info "Recently added dependencies in last commit:"
                echo "$recent_changes"
            fi
        fi
    fi
    
    # Check for packages from suspicious domains
    if [[ -f "go.mod" ]]; then
        local suspicious_domains=$(grep -E "^\s+[a-zA-Z]" go.mod | \
            grep -vE "(github\.com|golang\.org|google\.golang\.org|gopkg\.in|k8s\.io|cosmossdk\.io)" | \
            awk '{print $1}' | head -10 || true)
        
        if [[ -n "$suspicious_domains" ]]; then
            log_warning "Dependencies from non-standard domains (verify legitimacy):"
            echo "$suspicious_domains"
            add_finding "warning" "suspicious-changes" "Non-standard domain dependencies" "$(echo "$suspicious_domains" | tr '\n' ' ')"
            ((findings++))
        fi
    fi
    
    # Check for packages with git submodules (potential supply chain vector)
    if [[ -f ".gitmodules" ]]; then
        log_warning "Git submodules detected - these can be supply chain attack vectors"
        log_info "Submodules:"
        cat .gitmodules
        add_finding "warning" "suspicious-changes" "Git submodules present" ".gitmodules"
        ((findings++))
    fi
    
    if [[ $findings -eq 0 ]]; then
        log_success "No suspicious package changes detected"
    fi
    
    return 0
}

# Check for known vulnerable maintainer patterns
check_maintainer_risks() {
    log_info "Checking for maintainer risk indicators..."
    
    cd "$ROOT_DIR"
    
    # Check if any forked dependencies have diverged significantly
    if [[ -f "go.mod" ]]; then
        local forks=$(grep "=>" go.mod | grep -v "^\s*//" || true)
        
        if [[ -n "$forks" ]]; then
            log_info "Found forked dependencies (verify fork authenticity):"
            echo "$forks"
            
            # These are the known legitimate forks for VirtEngine
            local known_forks=(
                "github.com/akash-network/cosmos-sdk"
                "github.com/akash-network/cometbft"
                "github.com/akash-network/ledger-go"
                "github.com/akash-network/gogoproto"
                "github.com/cosmos/keyring"
                "github.com/regen-network/protobuf"
            )
            
            while IFS= read -r fork_line; do
                local fork_target=$(echo "$fork_line" | awk '{print $NF}')
                local is_known=false
                
                for known in "${known_forks[@]}"; do
                    if [[ "$fork_target" == *"$known"* ]]; then
                        is_known=true
                        break
                    fi
                done
                
                if [[ "$is_known" != "true" ]]; then
                    log_warning "Unrecognized fork: $fork_target - verify legitimacy"
                    add_finding "warning" "maintainer-risk" "Unrecognized fork dependency" "$fork_target"
                fi
            done <<< "$forks"
        fi
    fi
    
    return 0
}

# Generate summary report
generate_report() {
    if [[ "$OUTPUT_JSON" == "true" ]]; then
        echo "{"
        echo "  \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\","
        echo "  \"repository\": \"$(git remote get-url origin 2>/dev/null || echo 'unknown')\","
        echo "  \"commit\": \"$(git rev-parse HEAD 2>/dev/null || echo 'unknown')\","
        echo "  \"exit_code\": $EXIT_CODE,"
        echo "  \"findings\": ["
        
        local first=true
        for finding in "${JSON_FINDINGS[@]}"; do
            if [[ "$first" == "true" ]]; then
                first=false
            else
                echo ","
            fi
            echo -n "    $finding"
        done
        echo ""
        echo "  ]"
        echo "}"
    else
        echo ""
        echo "=========================================="
        echo "  Supply Chain Attack Detection Summary"
        echo "=========================================="
        echo ""
        
        if [[ $EXIT_CODE -eq 0 ]]; then
            log_success "No supply chain attack indicators detected"
        elif [[ $EXIT_CODE -eq 2 ]]; then
            log_warning "Warning-level issues detected - review recommended"
        else
            log_critical "CRITICAL: Potential supply chain attack indicators detected!"
            log_critical "Please review the findings above immediately"
        fi
    fi
}

# Main execution
main() {
    if [[ "$OUTPUT_JSON" != "true" ]]; then
        echo ""
        echo "=========================================="
        echo "  VirtEngine Supply Chain Attack Detection"
        echo "=========================================="
        echo ""
    fi
    
    # Run selected checks
    if [[ "$CHECK_ALL" == "true" ]] || [[ "$CHECK_CONFUSION" == "true" ]]; then
        check_dependency_confusion || true
        [[ "$OUTPUT_JSON" != "true" ]] && echo ""
    fi
    
    if [[ "$CHECK_ALL" == "true" ]] || [[ "$CHECK_TYPOSQUATTING" == "true" ]]; then
        check_typosquatting || true
        [[ "$OUTPUT_JSON" != "true" ]] && echo ""
    fi
    
    if [[ "$CHECK_ALL" == "true" ]] || [[ "$CHECK_CHANGES" == "true" ]]; then
        check_suspicious_changes || true
        [[ "$OUTPUT_JSON" != "true" ]] && echo ""
    fi
    
    # Always check maintainer risks
    if [[ "$CHECK_ALL" == "true" ]]; then
        check_maintainer_risks || true
        [[ "$OUTPUT_JSON" != "true" ]] && echo ""
    fi
    
    # Generate report
    generate_report
    
    exit $EXIT_CODE
}

# Run main
main
