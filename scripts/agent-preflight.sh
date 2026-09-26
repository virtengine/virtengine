#!/usr/bin/env bash
# Agent pre-flight check — run before git push
# Usage: ./scripts/agent-preflight.sh
set -euo pipefail

echo "=== Agent Pre-flight Check ==="

# ── Non-interactive git config (prevent editor popups) ──
git config --local core.editor ":" 2>/dev/null || true
git config --local merge.autoEdit false 2>/dev/null || true
export GIT_EDITOR=":"
export GIT_MERGE_AUTOEDIT="no"

# Detect what changed
CHANGED_FILES=$(git diff --cached --name-only 2>/dev/null || git diff --name-only HEAD~1 2>/dev/null || echo "")

if [ -z "$CHANGED_FILES" ]; then
    echo "No changed files detected. Skipping pre-flight."
    exit 0
fi

HAS_GO=false
HAS_PORTAL=false
HAS_GOMOD=false
HAS_BOSUN=false
ERRORS=0

echo "$CHANGED_FILES" | grep -q '\.go$' && HAS_GO=true || true
echo "$CHANGED_FILES" | grep -q '^portal/' && HAS_PORTAL=true || true
echo "$CHANGED_FILES" | grep -qE '^go\.(mod|sum)$' && HAS_GOMOD=true || true
echo "$CHANGED_FILES" | grep -q '^scripts/bosun/' && HAS_BOSUN=true || true

if $HAS_GO || $HAS_GOMOD; then
    echo "--- Go checks ---"

    if $HAS_GOMOD; then
        echo "  module/workspace/vendor policy..."
        ./scripts/verify-modules.sh 2>&1 || { echo "FAIL: module verification"; ERRORS=$((ERRORS+1)); }
    fi

    # Get changed Go packages
    GO_PKGS=$(echo "$CHANGED_FILES" | grep '\.go$' | xargs -I{} dirname {} | sort -u | sed 's|^|./|' || true)

    if [ -n "$GO_PKGS" ]; then
        echo "  gofmt..."
        echo "$CHANGED_FILES" | grep '\.go$' | xargs gofmt -w 2>&1 || true

        echo "  go vet..."
        echo "$GO_PKGS" | xargs go vet 2>&1 || { echo "FAIL: go vet"; ERRORS=$((ERRORS+1)); }

        echo "  go build..."
        go build ./cmd/... 2>&1 || { echo "FAIL: go build"; ERRORS=$((ERRORS+1)); }

        echo "  go test (changed packages)..."
        echo "$GO_PKGS" | xargs go test -short -count=1 2>&1 || { echo "FAIL: go test"; ERRORS=$((ERRORS+1)); }
    fi
fi

if $HAS_PORTAL; then
    echo "--- Portal checks ---"

    if [ ! -d "portal/node_modules" ]; then
        echo "  pnpm install..."
        pnpm -C portal install 2>&1 || { echo "FAIL: pnpm install"; ERRORS=$((ERRORS+1)); }
    fi

    echo "  ESLint..."
    pnpm -C portal lint 2>&1 || { echo "FAIL: eslint"; ERRORS=$((ERRORS+1)); }

    echo "  TypeScript..."
    pnpm -C portal type-check 2>&1 || { echo "FAIL: tsc"; ERRORS=$((ERRORS+1)); }

    echo "  Tests..."
    pnpm -C portal test 2>&1 || { echo "FAIL: portal tests"; ERRORS=$((ERRORS+1)); }
fi

if $HAS_BOSUN; then
    echo "--- Bosun checks ---"

    if [ ! -d "scripts/bosun/node_modules" ]; then
        echo "  npm install..."
        cd scripts/bosun
        npm install 2>&1 || { echo "FAIL: npm install"; ERRORS=$((ERRORS+1)); }
        cd - >/dev/null
    fi

    echo "  Prepublish check..."
    cd scripts/bosun
    node prepublish-check.mjs 2>&1 || { echo "FAIL: prepublish check"; ERRORS=$((ERRORS+1)); }
    cd - >/dev/null
fi

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "=== PRE-FLIGHT FAILED: $ERRORS error(s) ==="
    echo "Fix the issues above before pushing."
    exit 1
else
    echo "=== PRE-FLIGHT PASSED ==="
    exit 0
fi
