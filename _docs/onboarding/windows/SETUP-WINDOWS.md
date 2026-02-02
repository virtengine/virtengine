# VirtEngine Development Environment Setup for Windows (Git Bash)

## Quick Start

Run the setup script in Git Bash:

```bash
./scripts/setup-env-gitbash.sh
```

This script will check all dependencies and guide you through the setup process.

## Prerequisites

### Required Tools

1. **Git for Windows** (includes Git Bash)
   - Download: https://git-scm.com/download/win
   - Minimum version: 2.30+

2. **Go Programming Language**
   - Download: https://go.dev/dl/
   - Minimum version: **1.21.0** (1.22+ recommended for localnet)
   - Ensure Go is in your PATH

3. **Node.js and npm**
   - Download: https://nodejs.org/
   - LTS version recommended
   - npm comes bundled with Node.js

4. **Make** (GNU Make)
   - Install via Git Bash pacman: `pacman -S make`
   - Or use the version bundled with Git for Windows

5. **direnv** (version 2.32.0+)
   - Download from: https://github.com/direnv/direnv/releases
   - Extract `direnv.exe` to a directory in your PATH (e.g., `C:\Program Files\Git\usr\bin`)
   - Add to your `~/.bashrc`:
     ```bash
     eval "$(direnv hook bash)"
     ```

### Recommended Tools (for Git Bash)

Install these via Git Bash's pacman:

```bash
# Update package database
pacman -Sy

# Install recommended tools
pacman -S curl wget jq unzip coreutils pv lz4
```

### Optional Tools

- **Docker Desktop** (for running localnet)
  - Download: https://www.docker.com/products/docker-desktop
  - Required for integration testing and local blockchain network
  - Ensure WSL 2 backend is enabled in Docker Desktop settings

## Manual Setup Steps

If you prefer to set up manually instead of using the script:

### 1. Configure Git Bash Environment

Add to your `~/.bashrc` or `~/.bash_profile`:

```bash
# direnv integration
eval "$(direnv hook bash)"

# Ensure GOPATH is set
export GOPATH=$(go env GOPATH)
export PATH="$GOPATH/bin:$PATH"
```

### 2. Clone and Setup VirtEngine

```bash
# Clone the repository (if not already done)
git clone https://github.com/virtengine/virtengine.git
cd virtengine

# Allow direnv to load the environment
direnv allow .
```

The `.envrc` file will automatically:

- Check for required dependencies
- Set up environment variables
- Create `.cache` directories for build tools
- Configure git hooks for code quality

### 3. Configure direnv Auto-Allow (Optional)

To avoid manually running `direnv allow` each time:

Create/edit `~/.config/direnv/direnv.toml`:

```toml
[whitelist]
prefix = [
    "/c/Users/YOUR_USERNAME/path/to/virtengine"
]
```

Replace `YOUR_USERNAME` and the path with your actual setup.

### 4. Setup Git Hooks

Git hooks are automatically configured when direnv loads, but you can manually set them:

```bash
git config core.hooksPath .githooks
chmod +x .githooks/*
```

## Building VirtEngine

### Build the main binary

```bash
make virtengine
```

The binary will be created in `.cache/bin/virtengine`.

### Run tests

```bash
# Unit tests
make test

# Integration tests (requires more setup)
make test-integration

# Linting
make lint-go
```

## Running Localnet (Local Development Network)

**Note:** Localnet is best run in WSL2 on Windows for full compatibility.

### Using Docker Desktop

If you have Docker Desktop installed:

```bash
# Make scripts executable
chmod +x scripts/localnet.sh scripts/init-chain.sh

# Start localnet
./scripts/localnet.sh start

# Check status
./scripts/localnet.sh status

# View logs
./scripts/localnet.sh logs

# Stop localnet
./scripts/localnet.sh stop
```

### Using WSL2 (Recommended for Localnet)

For the best experience with localnet and integration tests:

1. Install WSL2: `wsl --install`
2. Install Ubuntu or your preferred distro
3. Clone the repo in WSL2
4. Follow the Linux setup instructions in `_docs/development-environment.md`

## Common Issues and Solutions

### Issue: `direnv: command not found`

**Solution:** Ensure direnv is installed and in your PATH. Add this to `~/.bashrc`:

```bash
eval "$(direnv hook bash)"
```

### Issue: `make: command not found`

**Solution:** Install make via pacman in Git Bash:

```bash
pacman -S make
```

### Issue: `.envrc` errors about missing tools

**Solution:** Install the missing tools listed in the Prerequisites section.

### Issue: Permission denied when running scripts

**Solution:** Make scripts executable:

```bash
chmod +x setup-env-gitbash.sh
chmod +x scripts/*.sh
```

### Issue: Docker not available

**Solution:**

- Install Docker Desktop for Windows
- Enable WSL 2 backend in Docker settings
- Or use WSL2 for localnet development

### Issue: `No git tags found`

This is expected for new clones. The `.envrc` will create cache directories manually. You can optionally create a tag:

```bash
git tag v0.1.0
```

## Environment Variables

Key environment variables set by `.envrc`:

- `VIRTENGINE_ROOT` - Project root directory
- `GOPATH` - Go workspace path
- `GOWORK` - Go workspace file path
- `VE_DEVCACHE` - Build tools cache directory
- `VE_DEVCACHE_BIN` - Build tools binary directory
- `PATH` - Extended with `.cache/bin` and other tool paths

## Next Steps

After setup:

1. **Test the build:**

   ```bash
   make virtengine
   ```

2. **Run unit tests:**

   ```bash
   make test
   ```

3. **Explore the documentation:**
   - `_docs/development-environment.md` - Detailed dev environment info
   - `_docs/testing-guide.md` - Testing strategies
   - `_docs/developer-guide.md` - Development workflows
   - `CONTRIBUTING.md` - Contribution guidelines

4. **Start developing:**
   - See `_docs/developer-guide.md` for development workflows
   - Use conventional commits (enforced by git hooks)
   - Run `make lint-go` before committing

## Additional Resources

- **VirtEngine Documentation:** `_docs/` directory
- **Go Documentation:** https://go.dev/doc/
- **Cosmos SDK Documentation:** https://docs.cosmos.network/
- **direnv Documentation:** https://direnv.net/

## Getting Help

- Check `_docs/` for detailed documentation
- Review `AGENTS.md` for repo structure and guidelines
- Check existing issues on GitHub
- Join the VirtEngine community channels

## Windows-Specific Notes

1. **Path Separators:** Git Bash uses Unix-style `/` paths. Windows `\` paths are automatically converted.

2. **Line Endings:** Configure git to handle line endings:

   ```bash
   git config --global core.autocrlf true
   ```

3. **Symbolic Links:** Some operations may require "Developer Mode" on Windows 10/11.

4. **Performance:** File I/O is slower than Linux. Consider WSL2 for intensive development.

5. **Localnet:** For full localnet functionality, use WSL2 or Docker Desktop with WSL2 backend.

6. **CGO Dependencies:** The project uses CGO (libusb/libhid). Ensure you have:
   - MinGW-w64 (usually bundled with Git for Windows)
   - Build tools available

7. **Terminal:** Use Git Bash, not PowerShell or CMD, for running shell scripts.
