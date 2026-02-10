# VirtEngine Git Bash Environment Setup - Summary

## Created Files

This setup process created the following files to help you configure the VirtEngine development environment on Windows:

### 1. **install-deps.ps1**

- **Purpose:** PowerShell script to automatically install dependencies
- **Usage:** Run in PowerShell (Administrator)
- **What it does:** Installs Chocolatey and all required packages (Git, Go, Node.js, Make, jq, etc.)

### 2. **setup-env-gitbash.sh**

- **Purpose:** Bash script to verify and configure the environment
- **Usage:** Run in Git Bash after installing dependencies
- **What it does:** Checks dependencies, configures direnv, sets up git hooks, creates cache directories

### 3. **GITBASH-SETUP.md**

- **Purpose:** Quick start guide for Git Bash setup
- **Contents:** Step-by-step instructions, verification checklist, troubleshooting

### 4. **SETUP-WINDOWS.md**

- **Purpose:** Comprehensive Windows setup documentation
- **Contents:** Detailed setup steps, environment configuration, localnet instructions

### 5. **INSTALL-DEPS-WINDOWS.md**

- **Purpose:** Detailed dependency installation guide
- **Contents:** Manual installation instructions for each tool, troubleshooting, alternatives

## Current System Status

✅ **Already Installed:**

- Git for Windows 2.51.1
- Node.js v20.18.0
- npm 10.8.2

❌ **Missing (Required):**

- Go 1.21.0+ (programming language for VirtEngine)
- GNU Make 4+ (build system)
- direnv 2.32+ (environment management)
- jq (JSON processor)

❌ **Missing (Recommended):**

- Docker Desktop (for localnet)
- curl, wget (download utilities)
- unzip (archive extraction)

## Quick Setup Path

### Option 1: Automated Installation (Recommended)

1. **Open PowerShell as Administrator**
2. **Run the installation script:**
   ```powershell
   cd C:\Users\YOUR_USERNAME\virtengine\virtengine
   .\scripts\install-deps.ps1
   ```
3. **Manual step - Install direnv:**
   - Download from: https://github.com/direnv/direnv/releases
   - Get `direnv.windows-amd64.exe`
   - Rename to `direnv.exe`
   - Move to `C:\Program Files\Git\usr\bin\`

4. **Open Git Bash and configure direnv:**

   ```bash
   echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
   source ~/.bashrc
   ```

5. **Run the Git Bash setup script:**

   ```bash
   cd /c/Users/jonathan/virtengine/virtengine
   ./setup-env-gitbash.sh
   ```

6. **Allow direnv and test build:**
   ```bash
   direnv allow .
   make virtengine
   ```

### Option 2: Manual Installation

Follow the detailed instructions in **INSTALL-DEPS-WINDOWS.md** to install each dependency manually.

### Option 3: Use WSL2 (Best Experience)

For the best development experience on Windows:

```powershell
# In PowerShell (Admin)
wsl --install

# After reboot, in WSL2:
sudo apt update
sudo apt install -y git build-essential golang nodejs npm direnv
cd /mnt/c/Users/jonathan/virtengine/virtengine
direnv allow .
make virtengine
```

## What Happens During Setup

### 1. Dependency Installation (install-deps.ps1)

- Installs Chocolatey package manager
- Installs Git, Go, Node.js, Make, jq, curl, wget
- Optionally installs Docker Desktop
- Provides instructions for direnv installation

### 2. Environment Configuration (setup-env-gitbash.sh)

- Verifies all dependencies are installed
- Checks versions meet minimum requirements
- Configures direnv hook in ~/.bashrc
- Sets up direnv auto-allow (optional)
- Allows direnv for the project directory

### 3. Project Setup (.envrc auto-loads)

- Sets GOPATH and VIRTENGINE_ROOT
- Creates `.cache` directories
- Adds cache bins to PATH
- Installs git hooks for code quality
- Configures Go workspace

### 4. Build Environment Ready

- Can run `make virtengine` to build
- Can run `make test` for unit tests
- Can run `./scripts/localnet.sh start` for integration testing (requires Docker)

## Environment Variables Set by .envrc

When you enter the VirtEngine directory, direnv automatically sets:

```bash
VIRTENGINE_ROOT=/c/Users/jonathan/virtengine/virtengine
GOPATH=<your go path>
GOWORK=<repo>/go.work  # or "off" if disabled
VE_DEVCACHE=<repo>/.cache
VE_DEVCACHE_BIN=<repo>/.cache/bin
PATH=<extended with cache bins>
```

## Next Steps After Setup

1. **Verify installation:**

   ```bash
   git --version
   go version
   node --version
   make --version
   jq --version
   direnv --version
   ```

2. **Test build:**

   ```bash
   make virtengine
   ```

3. **Run tests:**

   ```bash
   make test
   ```

4. **Read documentation:**
   - `_docs/developer-guide.md` - Development workflows
   - `_docs/testing-guide.md` - Testing strategies
   - `CONTRIBUTING.md` - How to contribute

5. **Start developing:**
   - Create a feature branch
   - Make changes following conventional commits
   - Run tests and linters before committing

## Troubleshooting Quick Reference

| Issue                       | Solution                                              |
| --------------------------- | ----------------------------------------------------- |
| "make: command not found"   | Run `install-deps.ps1` or `choco install make`        |
| "go: command not found"     | Install Go from https://go.dev/dl/                    |
| "direnv: command not found" | Install direnv manually (see INSTALL-DEPS-WINDOWS.md) |
| ".envrc is blocked"         | Run `direnv allow .` in project directory             |
| Scripts won't execute       | Run `chmod +x *.sh` in Git Bash                       |
| Git hooks not working       | Run `git config core.hooksPath .githooks`             |
| Docker not available        | Install Docker Desktop or use WSL2                    |

## Getting Help

- **General setup:** GITBASH-SETUP.md
- **Dependency installation:** INSTALL-DEPS-WINDOWS.md
- **Windows-specific issues:** SETUP-WINDOWS.md
- **Development environment:** \_docs/development-environment.md
- **Testing:** \_docs/testing-guide.md
- **Contributing:** CONTRIBUTING.md

## Important Notes

- **Always use Git Bash** for shell commands (not PowerShell or CMD)
- **direnv must be configured** in ~/.bashrc with the hook
- **Run setup scripts in order:** scripts/install-deps.ps1 → scripts/setup-env-gitbash.sh
- **Restart terminals** after installing new tools
- **WSL2 recommended** for best experience with localnet and testing

## Resources Created

Setup files are organized as follows:

```
virtengine/
├── scripts/
│   ├── install-deps.ps1          # PowerShell dependency installer
│   └── setup-env-gitbash.sh      # Git Bash environment setup
└── _docs/onboarding/windows/
    ├── README.md                 # This file (summary)
    ├── GITBASH-SETUP.md          # Quick start guide
    ├── SETUP-WINDOWS.md          # Comprehensive Windows guide
    └── INSTALL-DEPS-WINDOWS.md   # Detailed dependency instructions
```

## Support

If you encounter issues not covered in these documents:

1. Check the `_docs/` directory for additional documentation
2. Review existing GitHub issues
3. Check CONTRIBUTING.md for development guidelines
4. Review AGENTS.md for repository structure and patterns
