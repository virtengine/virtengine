# Lab Environment Setup

**Duration:** 2 hours  
**Prerequisites:** Basic command-line knowledge  
**Difficulty:** Beginner

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites Checklist](#prerequisites-checklist)
3. [Exercise 1: Install Required Tools](#exercise-1-install-required-tools)
4. [Exercise 2: Clone and Build VirtEngine](#exercise-2-clone-and-build-virtengine)
5. [Exercise 3: Start the Localnet](#exercise-3-start-the-localnet)
6. [Exercise 4: Verify Environment](#exercise-4-verify-environment)
7. [Troubleshooting](#troubleshooting)
8. [Cleanup](#cleanup)

---

## Overview

This lab guides you through setting up a complete VirtEngine local development environment. By the end, you'll have:

- All required development tools installed
- VirtEngine built from source
- A running localnet with validators, providers, and supporting services
- A verified environment ready for subsequent labs

### Environment Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       Local Development Environment                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │  VirtEngine     │  │   Provider      │  │    Waldur       │         │
│  │  Chain Node     │  │   Daemon        │  │    API          │         │
│  │  :26657 :1317   │  │   :8443         │  │    :8080        │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
│                                                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐         │
│  │   API Gateway   │  │   Prometheus    │  │    Grafana      │         │
│  │   (Kong) :8000  │  │   :9095         │  │    :3002        │         │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘         │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Prerequisites Checklist

Before starting, ensure you have:

| Requirement | Minimum Version | Check Command |
|-------------|-----------------|---------------|
| Operating System | Ubuntu 22.04 / macOS 12+ / Windows 10 (WSL2) | - |
| Docker | 20.10+ | `docker --version` |
| Docker Compose | 2.0+ | `docker compose version` |
| Go | 1.21+ | `go version` |
| Git | 2.30+ | `git --version` |
| Make | 4.0+ | `make --version` |
| curl | 7.0+ | `curl --version` |
| jq | 1.6+ | `jq --version` |

> **Windows Users:** Run all commands in WSL2. Ensure Docker Desktop is configured to use WSL2 backend.

---

## Exercise 1: Install Required Tools

### Objective
Install all tools required for VirtEngine development.

### Duration
30 minutes

### Instructions

#### Linux (Ubuntu/Debian)

```bash
# Update package list
sudo apt update

# Install essential tools
sudo apt install -y \
    build-essential \
    curl \
    wget \
    jq \
    git \
    ca-certificates \
    gnupg \
    lsb-release

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add user to docker group (logout/login required)
sudo usermod -aG docker $USER
newgrp docker

# Install Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
rm go1.22.0.linux-amd64.tar.gz

# Add Go to PATH (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install direnv (optional but recommended)
curl -sfL https://direnv.net/install.sh | bash
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
source ~/.bashrc
```

#### macOS

```bash
# Install Homebrew if not present
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install tools
brew install curl wget jq direnv coreutils make npm go@1.22

# Note: macOS make is outdated; use Homebrew version
export PATH="$(brew --prefix)/opt/make/libexec/gnubin:$PATH"
export PATH="$(brew --prefix)/opt/go@1.22/bin:$PATH"

# Add to shell profile (add to ~/.zshrc or ~/.bashrc)
echo 'export PATH="$(brew --prefix)/opt/make/libexec/gnubin:$PATH"' >> ~/.zshrc
echo 'export PATH="$(brew --prefix)/opt/go@1.22/bin:$PATH"' >> ~/.zshrc
echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc
source ~/.zshrc

# Install Docker Desktop from https://www.docker.com/products/docker-desktop/
# After installation, ensure it's running
```

#### Windows (WSL2)

```powershell
# In PowerShell (Administrator)
wsl --install -d Ubuntu-22.04

# Then follow Linux instructions inside WSL2

# Install Docker Desktop for Windows
# Enable WSL2 backend in Docker Desktop settings
```

### Verification

Run these commands to verify all tools are installed:

```bash
echo "=== Environment Verification ==="
echo "Docker: $(docker --version)"
echo "Docker Compose: $(docker compose version)"
echo "Go: $(go version)"
echo "Git: $(git --version)"
echo "Make: $(make --version | head -1)"
echo "jq: $(jq --version)"
echo "curl: $(curl --version | head -1)"

# Verify Docker is running
docker ps
```

### Expected Output

```
=== Environment Verification ===
Docker: Docker version 24.0.5, build ced0996
Docker Compose: Docker Compose version v2.20.2
Go: go version go1.22.0 linux/amd64
Git: git version 2.34.1
Make: GNU Make 4.3
jq: jq-1.6
curl: curl 7.81.0

CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES
```

---

## Exercise 2: Clone and Build VirtEngine

### Objective
Clone the VirtEngine repository and build the main binary.

### Duration
30 minutes

### Instructions

#### Step 1: Clone Repository

```bash
# Create workspace directory
mkdir -p ~/virtengine-lab
cd ~/virtengine-lab

# Clone repository
git clone https://github.com/virtengine/virtengine.git
cd virtengine

# Verify clone
ls -la
```

#### Step 2: Review Build Configuration

```bash
# Check Makefile targets
make help

# View cache directory setup
cat make/setup-cache.mk | head -50
```

#### Step 3: Build VirtEngine Binary

```bash
# Build the main binary
make virtengine

# This will:
# 1. Create .cache directory with build tools
# 2. Download dependencies
# 3. Compile the virtengine binary
```

#### Step 4: Verify Build

```bash
# Check binary location
ls -la .cache/bin/

# Verify binary works
.cache/bin/virtengine version

# Add to PATH (optional)
export PATH="$PWD/.cache/bin:$PATH"
virtengine version
```

### Expected Output

```
$ make virtengine
==> Installing build tools...
==> Downloading dependencies...
go: downloading github.com/cosmos/cosmos-sdk v0.53.0
go: downloading github.com/cometbft/cometbft v0.38.0
...
==> Building virtengine...
==> Build complete: .cache/bin/virtengine

$ .cache/bin/virtengine version
virtengine version 1.0.0
git commit: abc123def
go version: go1.22.0
```

### Common Build Issues

| Issue | Solution |
|-------|----------|
| `make: command not found` | Install make: `sudo apt install build-essential` |
| `go: command not found` | Add Go to PATH |
| CGO errors on Linux | Install `libusb-dev`: `sudo apt install libusb-1.0-0-dev` |
| CGO errors on macOS | Install Xcode CLI tools: `xcode-select --install` |
| Timeout downloading deps | Check network/proxy settings |

---

## Exercise 3: Start the Localnet

### Objective
Start the complete local development network using Docker Compose.

### Duration
30 minutes

### Instructions

#### Step 1: Review Localnet Script

```bash
# View available commands
./scripts/localnet.sh help
```

**Output:**
```
VirtEngine Local Development Network

Usage: ./scripts/localnet.sh [command]

Commands:
  start     Start the localnet (default)
  stop      Stop all services
  restart   Restart all services
  reset     Stop, clean data, and restart
  status    Show service status
  logs      Tail logs from all services
  test      Run integration tests
  shell     Open shell in test-runner container
  help      Show this help message
```

#### Step 2: Make Scripts Executable

```bash
chmod +x scripts/localnet.sh scripts/init-chain.sh
```

#### Step 3: Start Localnet

```bash
# Start all services
./scripts/localnet.sh start
```

This command will:
1. Build Docker images (first run takes 5-10 minutes)
2. Start all services in background
3. Wait for chain to be ready
4. Display service URLs

#### Step 4: Monitor Startup

```bash
# Watch logs during startup
./scripts/localnet.sh logs

# Press Ctrl+C to exit logs
```

#### Step 5: Verify Services

```bash
# Check service status
./scripts/localnet.sh status
```

### Expected Output

```
═══════════════════════════════════════════════════════════════
                 VirtEngine Local Development Network          
═══════════════════════════════════════════════════════════════

  Services:
    • Chain RPC:      http://localhost:26657
    • Chain REST:     http://localhost:1317
    • Chain gRPC:     localhost:9090
    • Waldur API:     http://localhost:8080
    • Portal:         http://localhost:3000
    • Provider API:   https://localhost:8443
    • API Gateway:    http://localhost:8000
    • Dev Portal:     http://localhost:3001/portal
    • Prometheus:     http://localhost:9095
    • Grafana:        http://localhost:3002

  Chain Info:
    • Chain ID:       virtengine-localnet-1
    • Keyring:        test

═══════════════════════════════════════════════════════════════
```

---

## Exercise 4: Verify Environment

### Objective
Verify all components are working correctly.

### Duration
30 minutes

### Instructions

#### Step 1: Test Chain RPC

```bash
# Check chain status
curl -s http://localhost:26657/status | jq '.result.sync_info'
```

**Expected Output:**
```json
{
  "latest_block_hash": "ABC123...",
  "latest_app_hash": "DEF456...",
  "latest_block_height": "100",
  "latest_block_time": "2026-01-24T12:00:00.000Z",
  "catching_up": false
}
```

#### Step 2: Test REST API

```bash
# Get node info
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info | jq '.default_node_info.moniker'

# List validators
curl -s http://localhost:1317/cosmos/staking/v1beta1/validators | jq '.validators | length'
```

#### Step 3: Test gRPC

```bash
# Install grpcurl if not present
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List available services
grpcurl -plaintext localhost:9090 list

# Query bank balances (example)
grpcurl -plaintext -d '{"address": "virtengine1..."}' localhost:9090 cosmos.bank.v1beta1.Query/AllBalances
```

#### Step 4: Check Test Accounts

The localnet creates pre-funded test accounts:

```bash
# Open shell in test container
./scripts/localnet.sh shell

# Inside container, list accounts
virtengine keys list --keyring-backend test

# Check balance
virtengine query bank balances $(virtengine keys show validator -a --keyring-backend test)
```

**Expected Accounts:**
| Account | Purpose |
|---------|---------|
| validator | Chain validator |
| alice | Test user |
| bob | Test user |
| charlie | Test user |
| provider | Test provider |
| operator | Test operator |

#### Step 5: Access Monitoring

```bash
# Open Grafana in browser
echo "Grafana: http://localhost:3002"
echo "Default credentials: admin/admin"

# Open Prometheus
echo "Prometheus: http://localhost:9095"

# Test Prometheus metrics endpoint
curl -s http://localhost:9095/api/v1/targets | jq '.data.activeTargets | length'
```

#### Step 6: Run Integration Tests

```bash
# Run basic integration tests
./scripts/localnet.sh test
```

### Verification Checklist

Complete this checklist before proceeding to other labs:

| Component | Test | Expected |
|-----------|------|----------|
| Chain RPC | `curl http://localhost:26657/status` | Returns JSON with block height |
| REST API | `curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info` | Returns node info |
| gRPC | `grpcurl -plaintext localhost:9090 list` | Lists gRPC services |
| Prometheus | `curl http://localhost:9095/-/ready` | Returns 200 OK |
| Grafana | Open http://localhost:3002 | Login page loads |
| Integration Tests | `./scripts/localnet.sh test` | Tests pass |

---

## Troubleshooting

### Docker Issues

#### Docker daemon not running

```bash
# Linux
sudo systemctl start docker
sudo systemctl enable docker

# macOS - Start Docker Desktop application

# Windows - Start Docker Desktop application
```

#### Port conflicts

```bash
# Check what's using a port
sudo lsof -i :26657
# or
sudo netstat -tulpn | grep 26657

# Kill conflicting process
sudo kill <PID>

# Alternative: Edit docker-compose.yaml to use different ports
```

#### Out of disk space

```bash
# Clean Docker resources
docker system prune -a --volumes

# Check disk usage
df -h
docker system df
```

### Build Issues

#### Go module download failures

```bash
# Clean module cache
go clean -modcache

# Set proxy if behind firewall
export GOPROXY=https://proxy.golang.org,direct

# Retry build
make virtengine
```

#### CGO compilation errors

```bash
# Linux - Install C compiler and libraries
sudo apt install build-essential libusb-1.0-0-dev

# macOS - Install Xcode tools
xcode-select --install
```

### Localnet Issues

#### Chain fails to start

```bash
# View detailed logs
docker compose logs virtengine-node

# Reset and try again
./scripts/localnet.sh reset
```

#### Services fail health checks

```bash
# Check container status
docker compose ps

# Restart specific service
docker compose restart <service-name>

# View service logs
docker compose logs -f <service-name>
```

#### Cannot connect to services

```bash
# Verify containers are running
docker compose ps

# Check if chain is synced
curl http://localhost:26657/status | jq '.result.sync_info.catching_up'
# Should return "false"

# Check Docker network
docker network ls
docker network inspect virtengine_default
```

### WSL2-Specific Issues

#### Docker socket not accessible

```bash
# Ensure Docker Desktop WSL integration is enabled
# Docker Desktop > Settings > Resources > WSL Integration

# Restart WSL
wsl --shutdown
# Open new WSL terminal
```

#### Slow performance

```bash
# Move project to WSL filesystem (not /mnt/c)
cd ~
mkdir -p virtengine-lab
cd virtengine-lab
git clone https://github.com/virtengine/virtengine.git
```

---

## Cleanup

### Stop Localnet

```bash
# Stop all services (preserves data)
./scripts/localnet.sh stop
```

### Full Reset

```bash
# Stop services and delete all data
./scripts/localnet.sh reset
```

### Remove Docker Resources

```bash
# Remove containers, networks, volumes
docker compose down -v --remove-orphans

# Remove images
docker rmi $(docker images -q 'virtengine*')
```

### Remove Build Artifacts

```bash
# Remove cache directory
rm -rf .cache

# Remove compiled binaries
rm -rf bin/
```

---

## Next Steps

After completing this lab, you're ready to proceed to:

1. **[Validator Operations Lab](validator-ops-lab.md)** - Set up and operate validator nodes
2. **[Provider Operations Lab](provider-ops-lab.md)** - Configure and run provider daemon
3. **[Incident Simulation Lab](incident-simulation-lab.md)** - Practice incident response
4. **[Security Assessment Lab](security-assessment-lab.md)** - Perform security audits

---

## Quick Reference

### Essential Commands

```bash
# Start environment
./scripts/localnet.sh start

# Check status
./scripts/localnet.sh status

# View logs
./scripts/localnet.sh logs

# Stop environment
./scripts/localnet.sh stop

# Reset everything
./scripts/localnet.sh reset

# Open shell
./scripts/localnet.sh shell
```

### Service URLs

| Service | URL |
|---------|-----|
| Chain RPC | http://localhost:26657 |
| Chain REST | http://localhost:1317 |
| Chain gRPC | localhost:9090 |
| Prometheus | http://localhost:9095 |
| Grafana | http://localhost:3002 |
| API Gateway | http://localhost:8000 |

### Test Account Mnemonics

> ⚠️ **WARNING**: These are for testing only. Never use in production!

The localnet uses the `test` keyring backend with pre-generated accounts. Access accounts inside the test container:

```bash
./scripts/localnet.sh shell
virtengine keys list --keyring-backend test
```

---

*Lab Version: 1.0.0*  
*Last Updated: 2026-01-24*  
*Maintainer: VirtEngine Training Team*
