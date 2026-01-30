#!/usr/bin/env bash
# VirtEngine Git Hook Utilities
# Sourced by individual hook scripts — provides colors, env loading, timing.

set -euo pipefail

# --- Colors (disabled if not a terminal) ---
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' BOLD='' NC=''
fi

hook_info()  { echo -e "${BLUE}[hook]${NC} $*"; }
hook_pass()  { echo -e "${GREEN}[PASS]${NC} $*"; }
hook_fail()  { echo -e "${RED}[FAIL]${NC} $*"; }
hook_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
hook_step()  { echo -e "${BLUE}  ->  ${NC} $*"; }
hook_header() { echo -e "${BOLD}$*${NC}"; }

# --- Timing ---
hook_time_start() { date +%s; }
hook_time_elapsed() {
    local start=$1
    local end
    end=$(date +%s)
    echo $(( end - start ))
}

# --- Environment loading ---
# Git hooks run in a subshell where direnv hasn't evaluated.
# We need VE_DIRENV_SET, VE_ROOT, GOTOOLCHAIN, and tool paths for make to work.
ensure_direnv_env() {
    if [ "${VE_DIRENV_SET:-}" = "1" ]; then
        return 0
    fi

    local repo_root
    repo_root="$(git rev-parse --show-toplevel)"

    # Try direnv export first (cleanest)
    if command -v direnv &>/dev/null; then
        eval "$(cd "$repo_root" && direnv export bash 2>/dev/null)" || true
    fi

    if [ "${VE_DIRENV_SET:-}" = "1" ]; then
        return 0
    fi

    # Fallback: manually set the critical variables from .env
    hook_warn "direnv not active. Loading environment manually..."

    export VIRTENGINE_ROOT="$repo_root"

    # Source .env for cache paths (uses variable substitution)
    export GO111MODULE=on
    export VIRTENGINE_DEVCACHE_BASE="${VIRTENGINE_ROOT}/.cache"
    export VIRTENGINE_DEVCACHE="${VIRTENGINE_DEVCACHE_BASE}"
    export VIRTENGINE_DEVCACHE_BIN="${VIRTENGINE_DEVCACHE}/bin"
    export VIRTENGINE_DEVCACHE_INCLUDE="${VIRTENGINE_DEVCACHE}/include"
    export VIRTENGINE_DEVCACHE_VERSIONS="${VIRTENGINE_DEVCACHE}/versions"
    export VIRTENGINE_DEVCACHE_NODE_MODULES="${VIRTENGINE_DEVCACHE}"
    export VIRTENGINE_DEVCACHE_NODE_BIN="${VIRTENGINE_DEVCACHE_NODE_MODULES}/node_modules/.bin"
    export VIRTENGINE_RUN="${VIRTENGINE_DEVCACHE}/run"
    export VIRTENGINE_RUN_BIN="${VIRTENGINE_RUN}/bin"
    export ROOT_DIR="${VIRTENGINE_ROOT}"

    # Set VE_ aliases for Makefile compatibility
    export VE_DIRENV_SET=1
    export VE_ROOT="$VIRTENGINE_ROOT"
    export VE_DEVCACHE="$VIRTENGINE_DEVCACHE"
    export VE_DEVCACHE_BIN="$VIRTENGINE_DEVCACHE_BIN"
    export VE_DEVCACHE_INCLUDE="$VIRTENGINE_DEVCACHE_INCLUDE"
    export VE_DEVCACHE_VERSIONS="$VIRTENGINE_DEVCACHE_VERSIONS"
    export VE_DEVCACHE_NODE_MODULES="$VIRTENGINE_DEVCACHE_NODE_MODULES"
    export VE_RUN="$VIRTENGINE_RUN"
    export VE_RUN_BIN="$VIRTENGINE_RUN_BIN"
    export VIRTENGINE_DIRENV_SET=1
    export VIRTENGINE="$VIRTENGINE_DEVCACHE_BIN/virtengine"

    # Extract GOTOOLCHAIN from go.mod
    export GOTOOLCHAIN
    GOTOOLCHAIN=$(grep '^toolchain' "$repo_root/go.mod" 2>/dev/null | awk '{print $2}' || echo "local")

    # GOWORK
    if [ -f "$repo_root/go.work" ]; then
        export GOWORK="$repo_root/go.work"
    else
        export GOWORK=off
    fi

    # GOPATH
    if [ -z "${GOPATH:-}" ]; then
        GOPATH=$(go env GOPATH 2>/dev/null || echo "$HOME/go")
        export GOPATH
    fi

    # Add tool paths
    export PATH="$VE_DEVCACHE_BIN:$VIRTENGINE_DEVCACHE_NODE_BIN:$PATH"

    # Ensure cache directories exist
    mkdir -p "$VE_DEVCACHE" "$VE_DEVCACHE_BIN" "$VE_DEVCACHE_INCLUDE" \
             "$VE_DEVCACHE_VERSIONS" "$VE_RUN_BIN" 2>/dev/null || true
}
