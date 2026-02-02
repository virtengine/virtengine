#!/usr/bin/env bash
# VirtEngine Development Environment Setup for Git Bash on Windows
# This script sets up the required tools and dependencies for VirtEngine development

set -e

echo "==================================================================="
echo "VirtEngine Development Environment Setup (Git Bash on Windows)"
echo "==================================================================="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running in Git Bash
if [[ ! "$OSTYPE" =~ "msys" ]] && [[ ! "$OSTYPE" =~ "cygwin" ]]; then
    echo -e "${RED}Error: This script must be run in Git Bash on Windows${NC}"
    exit 1
fi

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "ok" ]; then
        echo -e "${GREEN}✓${NC} $message"
    elif [ "$status" = "warn" ]; then
        echo -e "${YELLOW}⚠${NC} $message"
    else
        echo -e "${RED}✗${NC} $message"
    fi
}

# Check required tools
echo "Checking required dependencies..."
echo ""

# Git
if command_exists git; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    print_status "ok" "Git installed: $GIT_VERSION"
else
    print_status "error" "Git is not installed"
    echo "  Install from: https://git-scm.com/download/win"
    exit 1
fi

# Make
if command_exists make; then
    MAKE_VERSION=$(make --version | head -n 1)
    print_status "ok" "Make installed: $MAKE_VERSION"
else
    print_status "error" "Make is not installed"
    echo "  Install via: pacman -S make (in Git Bash)"
    exit 1
fi

# Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    GO_MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f2)
    
    if [ "$GO_MAJOR" -ge 1 ] && [ "$GO_MINOR" -ge 21 ]; then
        print_status "ok" "Go installed: $GO_VERSION (>= 1.21 required)"
    else
        print_status "warn" "Go version $GO_VERSION found, but 1.21+ recommended"
    fi
else
    print_status "error" "Go is not installed"
    echo "  Install from: https://go.dev/dl/"
    exit 1
fi

# curl
if command_exists curl; then
    print_status "ok" "curl is installed"
else
    print_status "error" "curl is not installed"
    echo "  Install via: pacman -S curl (in Git Bash)"
    exit 1
fi

# wget
if command_exists wget; then
    print_status "ok" "wget is installed"
else
    print_status "warn" "wget is not installed (recommended)"
    echo "  Install via: pacman -S wget (in Git Bash)"
fi

# jq
if command_exists jq; then
    print_status "ok" "jq is installed"
else
    print_status "error" "jq is not installed"
    echo "  Install via: pacman -S jq (in Git Bash)"
    exit 1
fi

# npm/node
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    NODE_VERSION=$(node --version)
    print_status "ok" "npm/node installed: npm $NPM_VERSION, node $NODE_VERSION"
else
    print_status "error" "npm/node is not installed"
    echo "  Install from: https://nodejs.org/"
    exit 1
fi

# direnv
if command_exists direnv; then
    DIRENV_VERSION=$(direnv version)
    DIRENV_MAJOR=$(echo $DIRENV_VERSION | cut -d. -f1)
    DIRENV_MINOR=$(echo $DIRENV_VERSION | cut -d. -f2)
    
    if [ "$DIRENV_MAJOR" -ge 2 ] && [ "$DIRENV_MINOR" -ge 32 ]; then
        print_status "ok" "direnv installed: $DIRENV_VERSION (>= 2.32 required)"
    else
        print_status "error" "direnv version $DIRENV_VERSION is too old (need 2.32+)"
        echo "  Download latest from: https://github.com/direnv/direnv/releases"
        exit 1
    fi
else
    print_status "error" "direnv is not installed"
    echo "  Install instructions:"
    echo "  1. Download from: https://github.com/direnv/direnv/releases"
    echo "  2. Extract binary to a directory in your PATH"
    echo "  3. Add to ~/.bashrc: eval \"\$(direnv hook bash)\""
    exit 1
fi

# unzip
if command_exists unzip; then
    print_status "ok" "unzip is installed"
else
    print_status "error" "unzip is not installed"
    echo "  Install via: pacman -S unzip (in Git Bash)"
    exit 1
fi

# readlink
if command_exists readlink; then
    print_status "ok" "readlink is installed"
else
    print_status "warn" "readlink is not installed (may be needed)"
    echo "  Install via: pacman -S coreutils (in Git Bash)"
fi

# pv
if command_exists pv; then
    print_status "ok" "pv is installed"
else
    print_status "warn" "pv is not installed (may be needed)"
    echo "  Install via: pacman -S pv (in Git Bash)"
fi

# lz4
if command_exists lz4; then
    print_status "ok" "lz4 is installed"
else
    print_status "warn" "lz4 is not installed (may be needed)"
    echo "  Install via: pacman -S lz4 (in Git Bash)"
fi

# Docker (for localnet)
if command_exists docker; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | tr -d ',')
    print_status "ok" "Docker installed: $DOCKER_VERSION"
else
    print_status "warn" "Docker is not installed (needed for localnet)"
    echo "  Install Docker Desktop: https://www.docker.com/products/docker-desktop"
fi

# docker-compose
if command_exists docker-compose; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $3}' | tr -d ',')
    print_status "ok" "docker-compose installed: $COMPOSE_VERSION"
else
    print_status "warn" "docker-compose not found (usually included with Docker Desktop)"
fi

echo ""
echo "==================================================================="
echo "Environment Configuration"
echo "==================================================================="
echo ""

# Check GOPATH
if [ -z "$GOPATH" ]; then
    export GOPATH=$(go env GOPATH)
    echo "GOPATH not set, using: $GOPATH"
else
    print_status "ok" "GOPATH: $GOPATH"
fi

# Check direnv hook
if grep -q "direnv hook bash" ~/.bashrc 2>/dev/null; then
    print_status "ok" "direnv hook configured in ~/.bashrc"
else
    print_status "warn" "direnv hook not found in ~/.bashrc"
    echo ""
    echo "To enable direnv, add this to your ~/.bashrc:"
    echo "  eval \"\$(direnv hook bash)\""
    echo ""
    read -p "Would you like to add it now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
        print_status "ok" "direnv hook added to ~/.bashrc"
        echo "  Restart your shell or run: source ~/.bashrc"
    fi
fi

# Setup direnv whitelist (optional)
DIRENV_CONFIG="${XDG_CONFIG_HOME:-$HOME/.config}/direnv/direnv.toml"
if [ -f "$DIRENV_CONFIG" ]; then
    print_status "ok" "direnv config exists: $DIRENV_CONFIG"
else
    print_status "warn" "direnv config not found"
    echo ""
    read -p "Would you like to create a direnv config with auto-allow for this project? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        mkdir -p "$(dirname "$DIRENV_CONFIG")"
        cat > "$DIRENV_CONFIG" << EOF
[whitelist]
prefix = [
    "$(pwd | sed 's|\\|/|g')"
]
EOF
        print_status "ok" "direnv config created with auto-allow for: $(pwd)"
    fi
fi

echo ""
echo "==================================================================="
echo "VirtEngine Project Setup"
echo "==================================================================="
echo ""

# Check if in virtengine directory
if [ ! -f "go.mod" ] || ! grep -q "github.com/virtengine/virtengine" go.mod 2>/dev/null; then
    print_status "error" "Not in VirtEngine project directory"
    exit 1
fi

print_status "ok" "In VirtEngine project directory"

# Check git tags (needed for cache setup)
if git describe --tags >/dev/null 2>&1; then
    GIT_TAG=$(git describe --tags)
    print_status "ok" "Git tags found: $GIT_TAG"
else
    print_status "warn" "No git tags found"
    echo "  The .envrc will create cache directories manually"
    echo "  Consider running: git tag v0.1.0"
fi

# Allow direnv for this directory
echo ""
echo "Allowing direnv for this directory..."
direnv allow .

echo ""
echo "Waiting for direnv to load environment..."
sleep 2

# Source the environment
eval "$(direnv export bash)"

echo ""
echo "==================================================================="
echo "Testing Build Environment"
echo "==================================================================="
echo ""

# Test make
echo "Testing make cache setup..."
if make cache; then
    print_status "ok" "Make cache setup successful"
else
    print_status "warn" "Make cache setup had issues (may need git tags)"
fi

# Check for .cache directory
if [ -d ".cache" ]; then
    print_status "ok" ".cache directory created"
    if [ -d ".cache/bin" ]; then
        print_status "ok" ".cache/bin directory exists"
    fi
fi

echo ""
echo "==================================================================="
echo "Git Hooks Setup"
echo "==================================================================="
echo ""

# Setup git hooks
if [ "$(git config core.hooksPath 2>/dev/null)" = ".githooks" ]; then
    print_status "ok" "Git hooks already configured"
else
    git config core.hooksPath .githooks
    chmod +x .githooks/* 2>/dev/null || true
    print_status "ok" "Git hooks configured to use .githooks/"
fi

echo ""
echo "==================================================================="
echo "Setup Complete!"
echo "==================================================================="
echo ""
echo "Next steps:"
echo "  1. Restart your Git Bash terminal (or run: source ~/.bashrc)"
echo "  2. Navigate back to this directory"
echo "  3. Run 'direnv allow .' if needed"
echo "  4. Test the build: make virtengine"
echo "  5. Run tests: make test"
echo "  6. Start localnet: ./scripts/localnet.sh start (requires Docker)"
echo ""
echo "Useful commands:"
echo "  make virtengine         - Build the main binary"
echo "  make test               - Run unit tests"
echo "  make test-integration   - Run integration tests"
echo "  make lint-go            - Run Go linters"
echo ""
echo "For Windows development, note:"
echo "  - Localnet requires Docker Desktop or WSL2"
echo "  - Some scripts may need WSL2 for full compatibility"
echo "  - See _docs/development-environment.md for details"
echo ""
