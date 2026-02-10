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
ERRORS=0

echo "$CHANGED_FILES" | grep -q '\.go$' && HAS_GO=true || true
echo "$CHANGED_FILES" | grep -q '^portal/' && HAS_PORTAL=true || true
echo "$CHANGED_FILES" | grep -qE '^go\.(mod|sum)$' && HAS_GOMOD=true || true

if $HAS_GO || $HAS_GOMOD; then
    echo "--- Go checks ---"

    if $HAS_GOMOD; then
        echo "  go mod tidy..."
        go mod tidy 2>&1 || { echo "FAIL: go mod tidy"; ERRORS=$((ERRORS+1)); }
        echo "  go mod vendor..."
        go mod vendor 2>&1 || { echo "FAIL: go mod vendor"; ERRORS=$((ERRORS+1)); }
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

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "=== PRE-FLIGHT FAILED: $ERRORS error(s) ==="
    echo "Fix the issues above before pushing."
    exit 1
else
    echo "=== PRE-FLIGHT PASSED ==="
    exit 0
fi
