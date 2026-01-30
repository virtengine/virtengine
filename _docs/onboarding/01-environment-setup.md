# Development Environment Setup

This guide walks you through setting up a complete development environment for VirtEngine.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Operating System Setup](#operating-system-setup)
3. [Install Dependencies](#install-dependencies)
4. [Clone and Build](#clone-and-build)
5. [IDE Configuration](#ide-configuration)
6. [Localnet Setup](#localnet-setup)
7. [Verification](#verification)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Tools

| Tool | Minimum Version | Recommended | Purpose |
|------|-----------------|-------------|---------|
| Go | 1.21.0 | 1.22+ | Core language runtime |
| GNU Make | 4.0 | 4.3+ | Build system |
| Docker | 20.10 | 24.0+ | Containerization |
| Docker Compose | 2.0 | 2.20+ | Multi-container orchestration |
| Node.js | 18 | 20 LTS | TypeScript SDK development |
| Python | 3.10 | 3.11+ | ML pipelines |
| Git | 2.30 | 2.40+ | Version control |
| direnv | 2.32 | 2.33+ | Environment management |

### System Requirements

- **CPU**: 4+ cores recommended
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 20GB free space (includes Docker images)
- **Network**: Internet access for dependencies

---

## Operating System Setup

### macOS

macOS ships with an outdated version of `make`. You must install GNU Make 4+:

```bash
# Install Homebrew if not present
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install curl wget jq direnv coreutils make npm go

# IMPORTANT: Add GNU make to PATH (add to ~/.zshrc or ~/.bashrc)
export PATH="$(brew --prefix)/opt/make/libexec/gnubin:$PATH"

# Verify make version (must be 4.0+)
make --version
```

### Linux (Debian/Ubuntu)

```bash
# Update package lists
sudo apt update

# Install dependencies
sudo apt install -y \
    git \
    curl \
    wget \
    jq \
    build-essential \
    ca-certificates \
    npm \
    direnv \
    gcc \
    libusb-1.0-0-dev \
    libhidapi-dev

# Install Go (if not using a package manager)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### Windows (WSL2)

VirtEngine development requires WSL2 on Windows:

```powershell
# Enable WSL2 (run as Administrator)
wsl --install -d Ubuntu-22.04

# Restart and open Ubuntu terminal
```

Then follow the Linux setup instructions inside WSL2.

> **Important**: Run all development commands from within WSL2. Ensure Docker Desktop is configured to use the WSL2 backend.

---

## Install Dependencies

### CGO Dependencies

VirtEngine requires CGO for Ledger device support:

```bash
# macOS
brew install libusb hidapi

# Linux (Debian/Ubuntu)
sudo apt install -y libusb-1.0-0-dev libhidapi-dev
```

### direnv Configuration

Set up direnv for automatic environment loading:

```bash
# Add to shell config (~/.bashrc, ~/.zshrc)
eval "$(direnv hook bash)"  # or zsh

# Whitelist VirtEngine directories
mkdir -p ~/.config/direnv
cat >> ~/.config/direnv/direnv.toml << 'EOF'
[whitelist]
prefix = [
    "~/code/virtengine",
    "~/go/src/github.com/virtengine"
]
EOF
```

---

## Clone and Build

### Clone Repository

```bash
# Using SSH (recommended)
git clone git@github.com:virtengine/virtengine.git
cd virtengine

# Or using HTTPS
git clone https://github.com/virtengine/virtengine.git
cd virtengine
```

### Build the Binary

```bash
# Build the main virtengine binary
make virtengine

# Build all binaries
make bins

# Verify build
.cache/bin/virtengine version
```

### Build Cache

The build system creates a `.cache` directory for tools:

```
.cache/
├── bin/           # Build tools (protoc, golangci-lint, etc.)
├── run/           # Work directories for examples
└── versions/      # Tool version tracking
```

You can customize the cache location:

```bash
export VE_DEVCACHE=/path/to/custom/cache
```

---

## IDE Configuration

### VS Code (Recommended)

Install recommended extensions:

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.go",
    "ms-python.python",
    "hashicorp.terraform",
    "redhat.vscode-yaml",
    "esbenp.prettier-vscode"
  ]
}
```

Configure Go settings:

```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "gofumpt",
  "go.testFlags": ["-v"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

### GoLand

1. Open the project root directory
2. Go to **Settings > Go > GOROOT** and set your Go installation
3. Enable **Go Modules** integration
4. Configure **golangci-lint** as the external linter

### Neovim

If using Neovim with LSP:

```lua
-- lua/lsp/go.lua
require('lspconfig').gopls.setup{
  settings = {
    gopls = {
      analyses = {
        unusedparams = true,
      },
      staticcheck = true,
    },
  },
}
```

---

## Localnet Setup

The localnet provides a complete local development environment.

### Start Localnet

```bash
# Make scripts executable (first time only)
chmod +x scripts/localnet.sh scripts/init-chain.sh

# Start all services
./scripts/localnet.sh start
```

### Available Services

| Service | Port | Description |
|---------|------|-------------|
| Chain RPC | 26657 | Tendermint RPC |
| Chain REST | 1317 | Cosmos REST API |
| Chain gRPC | 9090 | gRPC endpoint |
| Provider Daemon | 8443 | Provider API |
| API Gateway | 8000 | Kong proxy |
| Prometheus | 9095 | Metrics |
| Grafana | 3002 | Dashboards |

### Localnet Commands

```bash
# Start localnet
./scripts/localnet.sh start

# Stop all services
./scripts/localnet.sh stop

# View status
./scripts/localnet.sh status

# View logs (all services)
./scripts/localnet.sh logs

# View specific service logs
./scripts/localnet.sh logs virtengine-node

# Run integration tests
./scripts/localnet.sh test

# Reset (delete all data)
./scripts/localnet.sh reset

# Open shell in test container
./scripts/localnet.sh shell
```

### Test Accounts

The localnet creates these test accounts:

| Account | Purpose |
|---------|---------|
| validator | Chain validator |
| alice | Test user |
| bob | Test user |
| charlie | Test user |
| provider | Test provider |
| operator | Test operator |

---

## Verification

Run these commands to verify your setup:

```bash
# 1. Check Go version
go version
# Expected: go version go1.21+ ...

# 2. Check Make version
make --version
# Expected: GNU Make 4.0+

# 3. Build project
make virtengine
# Expected: Binary created in .cache/bin/

# 4. Run unit tests
go test ./x/veid/... -v
# Expected: Tests pass

# 5. Run linter
make lint-go
# Expected: No errors

# 6. Start localnet
./scripts/localnet.sh start
# Expected: Services start successfully

# 7. Query chain status
curl http://localhost:26657/status
# Expected: JSON response with chain info
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `make: *** No rule to make target` | Ensure GNU Make 4+ is installed and in PATH |
| `go: module not found` | Run `go mod tidy` |
| `cgo: C compiler not found` | Install build-essential (Linux) or Xcode CLI (macOS) |
| `permission denied` on scripts | Run `chmod +x scripts/*.sh` |
| Port already in use | Check for conflicting services, edit `docker-compose.yaml` |
| Docker permission denied | Add user to docker group: `sudo usermod -aG docker $USER` |

### CGO Errors

If you see CGO-related errors:

```bash
# macOS
brew install libusb hidapi
export CGO_ENABLED=1

# Linux
sudo apt install -y libusb-1.0-0-dev libhidapi-dev
export CGO_ENABLED=1
```

### Module Cache Issues

```bash
# Clear module cache
go clean -modcache

# Re-download modules
go mod download
```

### Build Cache Issues

```bash
# Clear build cache
make clean

# Remove tool cache
rm -rf .cache

# Rebuild
make virtengine
```

### Docker Issues

```bash
# Reset Docker state
docker system prune -af
docker volume prune -f

# Restart Docker service
sudo systemctl restart docker
```

---

## Next Steps

Once your environment is set up:

1. Read the [Architecture Overview](./02-architecture-overview.md)
2. Review [Contribution Guidelines](./03-code-contribution.md)
3. Explore the codebase starting with `x/veid/` module

## Related Documentation

- [Development Environment Reference](../development-environment.md) - Detailed environment configuration
- [Localnet Guide](../development-environment.md#local-development-network-localnet) - Advanced localnet usage
- [Testing Guide](./04-testing-guide.md) - Running and writing tests
