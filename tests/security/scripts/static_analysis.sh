#!/bin/bash
# Static Security Analysis
# Part of VirtEngine Penetration Testing Program

set -e

echo "========================================"
echo "VirtEngine Static Security Analysis"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

REPORT_DIR="./tests/security/reports"
mkdir -p "$REPORT_DIR"

DATE=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$REPORT_DIR/static_analysis_$DATE.txt"

echo "Report will be saved to: $REPORT_FILE"
echo ""

# Go security linting with gosec
echo "=== Go Security Linting (gosec) ===" | tee -a "$REPORT_FILE"

if command -v gosec &> /dev/null; then
    echo "Running gosec..." | tee -a "$REPORT_FILE"
    gosec -fmt text -severity medium ./... 2>&1 | tee -a "$REPORT_FILE" || true
else
    echo -e "${YELLOW}gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest${NC}" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"

# Check for hardcoded secrets
echo "=== Hardcoded Secret Detection ===" | tee -a "$REPORT_FILE"

# Patterns to search for
declare -a SECRET_PATTERNS=(
    "password\s*=\s*[\"'][^\"']+[\"']"
    "api[_-]?key\s*=\s*[\"'][^\"']+[\"']"
    "secret\s*=\s*[\"'][^\"']+[\"']"
    "private[_-]?key\s*="
    "-----BEGIN.*PRIVATE KEY-----"
    "aws_access_key_id"
    "aws_secret_access_key"
)

for pattern in "${SECRET_PATTERNS[@]}"; do
    echo "Checking for: $pattern" | tee -a "$REPORT_FILE"
    grep -rn --include="*.go" --include="*.yaml" --include="*.yml" --include="*.json" \
        -iE "$pattern" . 2>/dev/null | grep -v "_test.go" | grep -v "vendor/" | \
        head -20 | tee -a "$REPORT_FILE" || true
done

echo "" | tee -a "$REPORT_FILE"

# Check for dangerous functions
echo "=== Dangerous Function Usage ===" | tee -a "$REPORT_FILE"

declare -a DANGEROUS_FUNCS=(
    "time.Now()"           # Consensus safety - should use ctx.BlockTime()
    "rand.Read"            # Should use crypto/rand
    "unsafe.Pointer"       # Unsafe memory access
    "reflect.SliceHeader"  # Unsafe memory access
    "os.Exec"              # Command execution
    "exec.Command"         # Command execution
    "panic("               # In production paths
)

for func in "${DANGEROUS_FUNCS[@]}"; do
    echo "Checking for: $func" | tee -a "$REPORT_FILE"
    grep -rn --include="*.go" "$func" . 2>/dev/null | \
        grep -v "_test.go" | grep -v "vendor/" | \
        head -10 | tee -a "$REPORT_FILE" || true
done

echo "" | tee -a "$REPORT_FILE"

# Check for TODO/FIXME security comments
echo "=== Security TODOs ===" | tee -a "$REPORT_FILE"

grep -rn --include="*.go" -E "(TODO|FIXME).*(security|auth|crypt|key|secret|password)" . 2>/dev/null | \
    grep -v "vendor/" | tee -a "$REPORT_FILE" || true

echo "" | tee -a "$REPORT_FILE"

# Check for proper error handling
echo "=== Error Handling Check ===" | tee -a "$REPORT_FILE"

# Check for ignored errors
grep -rn --include="*.go" "_ = " . 2>/dev/null | \
    grep -v "_test.go" | grep -v "vendor/" | \
    head -20 | tee -a "$REPORT_FILE" || true

echo "" | tee -a "$REPORT_FILE"

# Run go vet
echo "=== Go Vet ===" | tee -a "$REPORT_FILE"
go vet ./... 2>&1 | tee -a "$REPORT_FILE" || true

echo "" | tee -a "$REPORT_FILE"

# Run staticcheck if available
echo "=== Staticcheck ===" | tee -a "$REPORT_FILE"

if command -v staticcheck &> /dev/null; then
    staticcheck ./... 2>&1 | tee -a "$REPORT_FILE" || true
else
    echo -e "${YELLOW}staticcheck not installed. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest${NC}" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"
echo "=== Analysis Complete ===" | tee -a "$REPORT_FILE"
echo "Report saved to: $REPORT_FILE"
