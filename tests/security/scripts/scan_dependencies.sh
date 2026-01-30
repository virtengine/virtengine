#!/bin/bash
# Dependency Vulnerability Scanner
# Part of VirtEngine Penetration Testing Program

set -e

echo "========================================"
echo "VirtEngine Dependency Vulnerability Scan"
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
REPORT_FILE="$REPORT_DIR/dependency_scan_$DATE.txt"

echo "Report will be saved to: $REPORT_FILE"
echo ""

# Go vulnerability check
echo "=== Go Dependency Scan ===" | tee -a "$REPORT_FILE"

if command -v govulncheck &> /dev/null; then
    echo "Running govulncheck..." | tee -a "$REPORT_FILE"
    govulncheck ./... 2>&1 | tee -a "$REPORT_FILE" || true
else
    echo -e "${YELLOW}govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest${NC}" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"

# Check go.sum for known vulnerable versions
echo "=== Checking go.mod for known issues ===" | tee -a "$REPORT_FILE"

# Known vulnerable patterns to check
declare -a VULN_PATTERNS=(
    "golang.org/x/crypto v0.0"  # Old crypto versions
    "golang.org/x/net v0.0"     # Old net versions
    "github.com/dgrijalva/jwt-go" # Vulnerable JWT library
    "gopkg.in/yaml.v2 v2.2."    # Old YAML with issues
)

for pattern in "${VULN_PATTERNS[@]}"; do
    if grep -q "$pattern" go.mod 2>/dev/null; then
        echo -e "${RED}[WARN] Potentially vulnerable dependency: $pattern${NC}" | tee -a "$REPORT_FILE"
    fi
done

echo "" | tee -a "$REPORT_FILE"

# Python dependency scan (if applicable)
if [ -d "ml" ]; then
    echo "=== Python Dependency Scan (ml/) ===" | tee -a "$REPORT_FILE"
    
    if command -v pip-audit &> /dev/null; then
        for req_file in ml/*/requirements*.txt; do
            if [ -f "$req_file" ]; then
                echo "Scanning: $req_file" | tee -a "$REPORT_FILE"
                pip-audit -r "$req_file" 2>&1 | tee -a "$REPORT_FILE" || true
            fi
        done
    else
        echo -e "${YELLOW}pip-audit not installed. Install with: pip install pip-audit${NC}" | tee -a "$REPORT_FILE"
    fi
fi

echo "" | tee -a "$REPORT_FILE"

# Node.js dependency scan (if applicable)
if [ -f "package.json" ] || [ -f "lib/portal/package.json" ]; then
    echo "=== Node.js Dependency Scan ===" | tee -a "$REPORT_FILE"
    
    if command -v npm &> /dev/null; then
        for pkg_dir in . lib/portal sdk/ts; do
            if [ -f "$pkg_dir/package.json" ]; then
                echo "Scanning: $pkg_dir" | tee -a "$REPORT_FILE"
                (cd "$pkg_dir" && npm audit --json 2>/dev/null || true) | tee -a "$REPORT_FILE"
            fi
        done
    else
        echo -e "${YELLOW}npm not installed${NC}" | tee -a "$REPORT_FILE"
    fi
fi

echo "" | tee -a "$REPORT_FILE"

# Container image scanning (if Docker available)
if command -v trivy &> /dev/null; then
    echo "=== Container Image Scan ===" | tee -a "$REPORT_FILE"
    
    # Scan Dockerfiles
    for dockerfile in $(find . -name "Dockerfile*" -type f 2>/dev/null); do
        echo "Scanning: $dockerfile" | tee -a "$REPORT_FILE"
        trivy config "$dockerfile" 2>&1 | tee -a "$REPORT_FILE" || true
    done
else
    echo "=== Container scanning skipped (trivy not installed) ===" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"
echo "=== Scan Complete ===" | tee -a "$REPORT_FILE"
echo "Report saved to: $REPORT_FILE"
