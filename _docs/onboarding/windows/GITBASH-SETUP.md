# Git Bash Environment Setup for VirtEngine on Windows

This directory contains scripts and documentation to help you set up the VirtEngine development environment on Windows using Git Bash.

## Quick Start

### 1. Install Dependencies (PowerShell - Admin Required)

Open PowerShell as Administrator and run:

```powershell
cd C:\Users\YOUR_USERNAME\virtengine\virtengine
.\scripts\install-deps.ps1
```

This will install:

- Chocolatey (package manager)
- Git for Windows
- Go (1.21+)
- Node.js and npm
- GNU Make
- jq, curl, wget
- Docker Desktop (optional)

### 2. Manual Step: Install direnv

After running the script above, manually install direnv:

1. Download from: https://github.com/direnv/direnv/releases
2. Get `direnv.windows-amd64.exe` (latest version)
3. Rename to `direnv.exe`
4. Move to: `C:\Program Files\Git\usr\bin\`

### 3. Configure direnv in Git Bash

Open Git Bash and add direnv hook to your shell:

```bash
# Edit ~/.bashrc
nano ~/.bashrc

# Add this line at the end:
eval "$(direnv hook bash)"

# Save (Ctrl+O, Enter, Ctrl+X) and reload:
source ~/.bashrc
```

### 4. Run VirtEngine Setup Script

```bash
# Navigate to VirtEngine directory
cd /c/Users/YOUR_USERNAME/virtengine/virtengine

# Make script executable
chmod +x scripts/setup-env-gitbash.sh

# Run setup script
./scripts/setup-env-gitbash.sh
```

The setup script will:

- Verify all dependencies
- Configure direnv for this project
- Set up git hooks
- Create `.cache` directories
- Test the build environment

### 5. Allow direnv for the Project

```bash
direnv allow .
```

### 6. Test the Build

```bash
# Build VirtEngine
make virtengine

# Run tests
make test
```

## Files in This Setup

| File                                               | Purpose                                                   |
| -------------------------------------------------- | --------------------------------------------------------- |
| `scripts/install-deps.ps1`                         | PowerShell script to install dependencies via Chocolatey  |
| `scripts/setup-env-gitbash.sh`                     | Bash script to verify and configure the dev environment   |
| `_docs/onboarding/windows/SETUP-WINDOWS.md`        | Comprehensive Windows setup guide                         |
| `_docs/onboarding/windows/INSTALL-DEPS-WINDOWS.md` | Detailed dependency installation instructions             |
| `.envrc`                                           | direnv configuration (auto-loads when entering directory) |

## What Gets Installed

### Required Tools

- **Git for Windows** - Version control and Git Bash environment
- **Go 1.21+** - Programming language for VirtEngine
- **Node.js & npm** - JavaScript runtime for build tools
- **GNU Make 4+** - Build system
- **direnv 2.32+** - Environment management
- **jq** - JSON processor for scripts
- **curl & wget** - Download utilities
- **unzip** - Archive extraction

### Optional Tools

- **Docker Desktop** - For running localnet and integration tests
- **pv** - Pipe viewer for progress monitoring
- **lz4** - Compression utility

## Directory Structure After Setup

```
virtengine/
├── .cache/                    # Build tools cache (created by direnv)
│   ├── bin/                   # Cached build tool binaries
│   ├── versions/              # Version tracking
│   └── run/                   # Runtime directories
├── .githooks/                 # Git hooks for code quality
├── scripts/
│   ├── setup-env-gitbash.sh   # Environment setup script
│   └── install-deps.ps1       # Dependency installation script
├── _docs/onboarding/windows/  # Windows setup documentation
└── ...
```

## Environment Variables

When direnv loads, it sets:

- `VIRTENGINE_ROOT` - Project root directory
- `GOPATH` - Go workspace
- `VE_DEVCACHE` - Cache directory (`.cache`)
- `VE_DEVCACHE_BIN` - Cached binaries (`.cache/bin`)
- `PATH` - Extended with cache bin directories

## Common Commands

```bash
# Build the main binary
make virtengine

# Run unit tests
make test

# Run integration tests
make test-integration

# Run linters
make lint-go

# Generate code
make generate

# Clean cache
make clean-cache

# Start localnet (requires Docker)
./scripts/localnet.sh start

# Stop localnet
./scripts/localnet.sh stop
```

## Troubleshooting

### Issue: "make: command not found"

**Solution:** Make was not installed. Run `install-deps.ps1` in PowerShell (Admin) or install manually:

```powershell
choco install make
```

### Issue: "direnv: command not found"

**Solution:** Install direnv manually (see step 2 above) and ensure `eval "$(direnv hook bash)"` is in `~/.bashrc`.

### Issue: ".envrc is blocked"

**Solution:** Run `direnv allow .` in the project directory.

### Issue: "Go version too old"

**Solution:** Update Go to 1.21+ from https://go.dev/dl/

### Issue: Scripts have wrong line endings

**Solution:** Configure git line endings:

```bash
git config --global core.autocrlf true
```

### Issue: Permission denied on scripts

**Solution:** Make scripts executable:

```bash
chmod +x setup-env-gitbash.sh
chmod +x scripts/*.sh
```

## Alternative: Using WSL2

For the best experience, consider using WSL2 instead of Git Bash:

```powershell
# In PowerShell (Admin)
wsl --install
```

After WSL2 is installed, follow the Linux setup instructions in `_docs/development-environment.md`.

### WSL2 Advantages

- Better script compatibility
- Faster file I/O
- Native Linux tools
- Easier Docker integration
- Full localnet support

## Verification Checklist

After setup, verify everything works:

- [ ] Git Bash opens without errors
- [ ] `git --version` shows Git 2.30+
- [ ] `go version` shows Go 1.21+
- [ ] `node --version` and `npm --version` work
- [ ] `make --version` shows GNU Make 4+
- [ ] `jq --version` works
- [ ] `direnv --version` shows 2.32+
- [ ] `direnv allow .` in project directory succeeds
- [ ] Environment variables are set (`echo $VIRTENGINE_ROOT`)
- [ ] `make virtengine` builds successfully
- [ ] Git hooks are configured (`git config core.hooksPath` shows `.githooks`)

## Next Steps

After successful setup:

1. **Read the documentation:**
   - `_docs/developer-guide.md` - Development workflows
   - `_docs/testing-guide.md` - Testing strategies
   - `CONTRIBUTING.md` - Contribution guidelines

2. **Start developing:**
   - Create a feature branch
   - Make changes
   - Run tests: `make test`
   - Run linters: `make lint-go`
   - Commit with conventional commits

3. **Explore the codebase:**
   - `x/` - Blockchain modules
   - `pkg/` - Shared packages
   - `cmd/` - CLI binaries
   - `app/` - Application wiring

4. **Join the community:**
   - Check GitHub issues
   - Review AGENTS.md for AI development guidance
   - Read architecture docs in `_docs/`

## Getting Help

- **Setup issues:** Check `INSTALL-DEPS-WINDOWS.md`
- **Development questions:** See `_docs/developer-guide.md`
- **Testing:** Review `_docs/testing-guide.md`
- **Contributing:** Read `CONTRIBUTING.md`
- **Architecture:** Explore `_docs/architecture.md`

## Important Notes for Windows

1. **Use Git Bash** for all shell commands, not PowerShell or CMD
2. **Line endings:** Git should be configured for Windows (`core.autocrlf=true`)
3. **Paths:** Git Bash uses Unix-style paths (`/c/Users/...`)
4. **Localnet:** Requires Docker Desktop or WSL2
5. **Performance:** WSL2 offers better performance than Git Bash for intensive builds

## Resources

- VirtEngine Docs: `_docs/` directory
- Git for Windows: https://git-scm.com/download/win
- Go Downloads: https://go.dev/dl/
- Node.js: https://nodejs.org/
- direnv: https://direnv.net/
- Docker Desktop: https://www.docker.com/products/docker-desktop
- Chocolatey: https://chocolatey.org/
