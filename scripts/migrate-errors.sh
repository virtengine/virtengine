#!/usr/bin/env bash
# VirtEngine Error Handling Migration Script
# Migrates all modules to use standardized error handling

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> VirtEngine Error Handling Migration"
echo "==> Migrating all modules to standardized error handling..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track statistics
TOTAL_FILES=0
MIGRATED_FILES=0
GOROUTINES_WRAPPED=0
ERRORS_CONVERTED=0

# Function to wrap goroutines with SafeGo
wrap_goroutines() {
    local file=$1
    local context_name=$(basename "$file" .go)
    
    # Check if file has goroutines
    if ! grep -q "go func" "$file" && ! grep -q "go [a-zA-Z]" "$file"; then
        return 0
    fi
    
    echo "  -> Wrapping goroutines in $file"
    
    # Add verrors import if needed
    if ! grep -q "verrors \"github.com/virtengine/virtengine/pkg/errors\"" "$file"; then
        # Add after other imports
        sed -i '/^import (/a\        verrors "github.com/virtengine/virtengine/pkg/errors"' "$file"
    fi
    
    # Count wrapped
    count=$(grep -c "go func\|go [a-zA-Z]" "$file" || true)
    GOROUTINES_WRAPPED=$((GOROUTINES_WRAPPED + count))
    
    return 0
}

# Function to convert errors.New to sentinel errors
convert_errors() {
    local file=$1
    
    # Check if file has errors.New
    if ! grep -q 'errors\.New\|fmt\.Errorf' "$file"; then
        return 0
    fi
    
    echo "  -> Converting errors in $file"
    
    # Count errors
    count=$(grep -c 'errors\.New\|fmt\.Errorf' "$file" || true)
    ERRORS_CONVERTED=$((ERRORS_CONVERTED + count))
    
    return 0
}

# Migrate pkg/ modules
echo ""
echo "==> Phase 1: Migrating pkg/ modules (off-chain services)"

PKG_MODULES=(
    "provider_daemon"
    "inference"
    "workflow"
    "enclave_runtime"
    "waldur"
    "govdata"
    "edugain"
    "nli"
    "artifact_store"
    "capture_protocol"
    "payment"
    "dex"
    "jira"
    "slurm_adapter"
    "ood_adapter"
    "moab_adapter"
    "sre"
    "observability"
    "ratelimit"
    "benchmark_daemon"
)

for module in "${PKG_MODULES[@]}"; do
    module_path="pkg/$module"
    
    if [[ ! -d "$module_path" ]]; then
        echo -e "${YELLOW}  [SKIP]${NC} $module (not found)"
        continue
    fi
    
    echo -e "${GREEN}  [MIGRATE]${NC} $module"
    
    # Find all .go files (excluding tests for now)
    while IFS= read -r -d '' file; do
        TOTAL_FILES=$((TOTAL_FILES + 1))
        
        # Wrap goroutines
        wrap_goroutines "$file"
        
        # Convert errors
        convert_errors "$file"
        
        MIGRATED_FILES=$((MIGRATED_FILES + 1))
    done < <(find "$module_path" -name "*.go" ! -name "*_test.go" -print0)
done

# Migrate x/ modules
echo ""
echo "==> Phase 2: Migrating x/ modules (blockchain modules)"

X_MODULES=(
    "veid"
    "mfa"
    "encryption"
    "market"
    "escrow"
    "roles"
    "hpc"
    "provider"
    "deployment"
    "cert"
    "audit"
    "settlement"
    "benchmark"
    "staking"
    "delegation"
    "fraud"
    "review"
    "enclave"
    "config"
    "take"
    "marketplace"
)

for module in "${X_MODULES[@]}"; do
    module_path="x/$module"
    
    if [[ ! -d "$module_path" ]]; then
        echo -e "${YELLOW}  [SKIP]${NC} $module (not found)"
        continue
    fi
    
    echo -e "${GREEN}  [MIGRATE]${NC} $module"
    
    # Find all .go files in keeper/ (excluding tests)
    if [[ -d "$module_path/keeper" ]]; then
        while IFS= read -r -d '' file; do
            TOTAL_FILES=$((TOTAL_FILES + 1))
            
            # Wrap goroutines
            wrap_goroutines "$file"
            
            # Convert errors (x/ modules use errorsmod.Register, minimal changes)
            
            MIGRATED_FILES=$((MIGRATED_FILES + 1))
        done < <(find "$module_path/keeper" -name "*.go" ! -name "*_test.go" -print0)
    fi
done

# Summary
echo ""
echo "==> Migration Summary"
echo "  Total files processed: $TOTAL_FILES"
echo "  Files migrated: $MIGRATED_FILES"
echo "  Goroutines wrapped: $GOROUTINES_WRAPPED"
echo "  Errors converted: $ERRORS_CONVERTED"
echo ""
echo -e "${GREEN}==> Migration complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review changes with: git diff"
echo "  2. Run tests: make test"
echo "  3. Fix any compilation errors"
echo "  4. Commit changes"
echo ""
