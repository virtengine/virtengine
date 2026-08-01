# VirtEngine Development Environment Setup for Windows

## Quick Start

VirtEngine supports native Windows builds and unit tests from PowerShell. Install Go `1.25.8`, then run:

```powershell
$env:CGO_ENABLED = "0"
go build -mod=readonly -o .\bin\virtengine.exe .\cmd\virtengine
go test -mod=readonly -short -count=1 ./...
```

Use `scripts/localnet.ps1` with Docker Desktop to run the complete local environment; WSL is not required by the launcher.

## Prerequisites

### Required Tools

1. **Git for Windows** (includes Git Bash)
   - Download: https://git-scm.com/download/win
   - Minimum version: 2.30+

2. **Go Programming Language**
   - Download: https://go.dev/dl/
   - Required version: **1.25.8**
   - Ensure Go is in your PATH

3. **Node.js and npm**
   - Download: https://nodejs.org/
   - LTS version recommended
   - npm comes bundled with Node.js

4. **Docker Desktop** (complete localnet only)
   - Use Docker Desktop configured for Linux containers.
   - Docker Desktop itself still requires a virtualization backend (WSL2 or Hyper-V); this is a Docker requirement, not a VirtEngine shell requirement.

5. **C compiler** (optional, race tests and CGO features)
   - Install a supported MinGW-w64 toolchain and set `CC` to its `gcc.exe`.
   - Standard builds and short unit tests use `CGO_ENABLED=0` and do not require it.

### Optional Tools

These are needed only for specific workflows:

- Node.js 20+ and pnpm `10.28.2` for portal and TypeScript SDK work.
- GNU Make and Git Bash for legacy Make and Bash-only development helpers.
- `direnv` for the Git Bash workflow. It is not required by native PowerShell commands.

## Manual Setup Steps

If you prefer to set up manually instead of using the script:

### 1. Configure PowerShell Environment

Open a new PowerShell window after installing Go, then verify the toolchain:

```powershell
go version
git --version
```

### 2. Clone and Setup VirtEngine

```powershell
# Clone the repository (if not already done)
git clone https://github.com/virtengine/virtengine.git
Set-Location virtengine
```

The PowerShell build and test commands do not depend on `.envrc`. Configure Git hooks with `git config core.hooksPath .githooks`; use `pwsh scripts/agent-preflight.ps1` before pushing.

## Building VirtEngine

### Build the main binary

```powershell
New-Item -ItemType Directory -Force .\bin | Out-Null
$env:CGO_ENABLED = "0"
go build -mod=readonly -o .\bin\virtengine.exe .\cmd\virtengine
```

The binary is created at `.\bin\virtengine.exe`.

### Run tests

```powershell
# Unit tests
$env:CGO_ENABLED = "0"
go test -mod=readonly -short -count=1 ./...

# Race tests require MinGW-w64 and CGO_ENABLED=1.
go test -mod=readonly -race -count=1 ./pkg/provider_daemon
```

## Running Localnet (Local Development Network)

The localnet launcher is native PowerShell. Docker Desktop runs the Linux service containers; it must be configured with a functional Linux-container backend.

### Using Docker Desktop

If Docker Desktop is running:

```powershell
# Start and inspect the localnet
pwsh .\scripts\localnet.ps1 start
pwsh .\scripts\localnet.ps1 status
pwsh .\scripts\localnet.ps1 logs virtengine-node

# Integration tests and lifecycle
pwsh .\scripts\localnet.ps1 test
pwsh .\scripts\localnet.ps1 stop
```

`start`, `stop`, `restart`, `update`, `reset`, `status`, `logs`, `test`, `shell`, and `create-admin` mirror the Bash launcher. Use `-Foreground` to retain Compose output and `reset -Force` to skip the destructive-action prompt.

## Common Issues and Solutions

### Issue: `go: command not found`

**Solution:** Install Go `1.25.8`, then open a new PowerShell window so its installation directory is added to `PATH`.

### Issue: `.envrc` errors about missing tools

**Solution:** Install the missing tools listed in the Prerequisites section.

### Issue: Docker not available

**Solution:**

- Install Docker Desktop for Windows
- Configure a Linux-container backend (WSL2 or Hyper-V)
- Ensure hardware virtualization is enabled before running the complete localnet

### Issue: `No git tags found`

This is expected for new clones. The `.envrc` will create cache directories manually. You can optionally create a tag:

```bash
git tag v0.1.0
```

## Environment Variables

Key environment variables:

- `CHAIN_ID` - Overrides the Docker localnet chain ID.
- `LOG_LEVEL` - Overrides localnet logging (default `info`).
- `CGO_ENABLED` - Set to `1` only when a Windows C compiler is configured.
- `CC` - Path to MinGW-w64 `gcc.exe` for CGO and race tests.

## Next Steps

After setup:

1. **Test the build:**

   ```powershell
   $env:CGO_ENABLED = "0"
   go build -mod=readonly -o .\bin\virtengine.exe .\cmd\virtengine
   ```

2. **Run unit tests:**

   ```powershell
   $env:CGO_ENABLED = "0"
   go test -mod=readonly -short -count=1 ./...
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

4. **Performance:** Native Go builds and unit tests run directly on Windows. Docker localnet performance depends on Docker Desktop's selected virtualization backend.

5. **Localnet:** Use `pwsh .\scripts\localnet.ps1 start`. The launcher does not require Git Bash or WSL.

6. **CGO Dependencies:** The project uses CGO (libusb/libhid). Ensure you have:
   - MinGW-w64 (usually bundled with Git for Windows)
   - Build tools available

7. **Terminal:** Use PowerShell for the native workflow. Use Git Bash only for legacy `.sh` helpers that do not yet have a PowerShell equivalent.
