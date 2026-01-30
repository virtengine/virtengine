#!/bin/bash
# Secret Detection Scanner
# Part of VirtEngine Penetration Testing Program

set -e

echo "========================================"
echo "VirtEngine Secret Detection Scan"
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
REPORT_FILE="$REPORT_DIR/secret_scan_$DATE.txt"

echo "Report will be saved to: $REPORT_FILE"
echo ""

# Run gitleaks if available
echo "=== Gitleaks Scan ===" | tee -a "$REPORT_FILE"

if command -v gitleaks &> /dev/null; then
    echo "Running gitleaks..." | tee -a "$REPORT_FILE"
    gitleaks detect --source . --verbose 2>&1 | tee -a "$REPORT_FILE" || true
else
    echo -e "${YELLOW}gitleaks not installed. Install from: https://github.com/gitleaks/gitleaks${NC}" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"

# Run truffleHog if available
echo "=== TruffleHog Scan ===" | tee -a "$REPORT_FILE"

if command -v trufflehog &> /dev/null; then
    echo "Running trufflehog..." | tee -a "$REPORT_FILE"
    trufflehog filesystem . 2>&1 | tee -a "$REPORT_FILE" || true
else
    echo -e "${YELLOW}trufflehog not installed${NC}" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"

# Manual pattern scanning
echo "=== Pattern-Based Secret Detection ===" | tee -a "$REPORT_FILE"

# High-entropy string detection
echo "Checking for high-entropy strings..." | tee -a "$REPORT_FILE"

# Common secret patterns
declare -a PATTERNS=(
    # API Keys
    "AIza[0-9A-Za-z_-]{35}"                    # Google API Key
    "AKIA[0-9A-Z]{16}"                         # AWS Access Key
    "sk_live_[0-9a-zA-Z]{24}"                  # Stripe Secret Key
    "sk_test_[0-9a-zA-Z]{24}"                  # Stripe Test Key
    "ghp_[0-9a-zA-Z]{36}"                      # GitHub Personal Access Token
    "gho_[0-9a-zA-Z]{36}"                      # GitHub OAuth Token
    "xox[baprs]-[0-9a-zA-Z-]{10,}"            # Slack Token
    
    # Private Keys
    "-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----"
    "-----BEGIN PGP PRIVATE KEY BLOCK-----"
    
    # Certificates
    "-----BEGIN CERTIFICATE-----"
    
    # Database connection strings
    "postgres://[^:]+:[^@]+@"
    "mysql://[^:]+:[^@]+@"
    "mongodb://[^:]+:[^@]+@"
    "redis://:[^@]+@"
    
    # Generic secrets
    "password[\"']?\\s*[:=]\\s*[\"'][^\"']{8,}[\"']"
    "secret[\"']?\\s*[:=]\\s*[\"'][^\"']{8,}[\"']"
    "token[\"']?\\s*[:=]\\s*[\"'][^\"']{20,}[\"']"
)

for pattern in "${PATTERNS[@]}"; do
    echo "Checking pattern: ${pattern:0:30}..." | tee -a "$REPORT_FILE"
    grep -rnoE "$pattern" . 2>/dev/null | \
        grep -v "vendor/" | \
        grep -v "node_modules/" | \
        grep -v ".git/" | \
        grep -v "_test.go" | \
        head -5 | tee -a "$REPORT_FILE" || true
done

echo "" | tee -a "$REPORT_FILE"

# Check .env files
echo "=== Environment File Check ===" | tee -a "$REPORT_FILE"

find . -name ".env*" -type f 2>/dev/null | while read envfile; do
    if [ -f "$envfile" ]; then
        echo -e "${YELLOW}Found env file: $envfile${NC}" | tee -a "$REPORT_FILE"
        # Check if it's in .gitignore
        if grep -q "$(basename $envfile)" .gitignore 2>/dev/null; then
            echo "  [OK] Listed in .gitignore" | tee -a "$REPORT_FILE"
        else
            echo -e "  ${RED}[WARN] NOT in .gitignore${NC}" | tee -a "$REPORT_FILE"
        fi
    fi
done

echo "" | tee -a "$REPORT_FILE"

# Check for credentials in config files
echo "=== Config File Credential Check ===" | tee -a "$REPORT_FILE"

for config in $(find . -name "*.yaml" -o -name "*.yml" -o -name "*.json" -o -name "*.toml" 2>/dev/null | grep -v vendor/ | grep -v node_modules/); do
    secrets_found=$(grep -inE "(password|secret|api_key|token|credential)\\s*[:=]" "$config" 2>/dev/null | grep -v '""' | grep -v "null" | grep -v '${' | head -3)
    if [ -n "$secrets_found" ]; then
        echo -e "${YELLOW}Potential secrets in $config:${NC}" | tee -a "$REPORT_FILE"
        echo "$secrets_found" | tee -a "$REPORT_FILE"
    fi
done

echo "" | tee -a "$REPORT_FILE"

# Check git history for secrets (limited check)
echo "=== Git History Check (last 100 commits) ===" | tee -a "$REPORT_FILE"

if [ -d ".git" ]; then
    git log --oneline -100 --all --diff-filter=A --name-only 2>/dev/null | \
        grep -E "\.(pem|key|p12|pfx|jks|keystore)$" | \
        head -20 | tee -a "$REPORT_FILE" || true
else
    echo "Not a git repository" | tee -a "$REPORT_FILE"
fi

echo "" | tee -a "$REPORT_FILE"
echo "=== Scan Complete ===" | tee -a "$REPORT_FILE"
echo "Report saved to: $REPORT_FILE"

# Summary
echo ""
echo "========================================"
echo "Scan Summary"
echo "========================================"
if grep -q "WARN" "$REPORT_FILE" 2>/dev/null; then
    echo -e "${RED}Potential issues found - review $REPORT_FILE${NC}"
else
    echo -e "${GREEN}No obvious secrets detected${NC}"
fi
