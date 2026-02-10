# Installing Dependencies for VirtEngine on Windows

This guide helps you install the required tools for VirtEngine development on Windows.

## Quick Installation Summary

### Option 1: Using Chocolatey (Recommended)

Install Chocolatey first, then use it to install all dependencies:

```powershell
# Run in PowerShell (Admin)
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install dependencies
choco install -y git golang nodejs make jq curl wget
```

### Option 2: Using Scoop

```powershell
# Run in PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Install dependencies
scoop install git go nodejs make jq curl wget
```

### Option 3: Manual Installation (Detailed Below)

## Detailed Installation Instructions

### 1. Git for Windows ✓ (Already Installed)

You already have Git 2.51.1 installed.

### 2. GNU Make (Required)

Make is **required** for building VirtEngine. Choose one option:

#### Option A: Install via Chocolatey (Easiest)

```powershell
# In PowerShell (Admin)
choco install make
```

#### Option B: Install via Scoop

```powershell
# In PowerShell
scoop install make
```

#### Option C: Download Pre-built Binary

1. Download Make for Windows: https://gnuwin32.sourceforge.net/packages/make.htm
2. Or download from: https://github.com/mbuilov/gnumake-windows/releases
3. Extract `make.exe` to `C:\Program Files\Git\usr\bin\`
4. Verify: Open Git Bash and run `make --version`

#### Option D: Install MSYS2 (Most Complete)

1. Download MSYS2: https://www.msys2.org/
2. Run the installer
3. Open MSYS2 terminal and run:
   ```bash
   pacman -Syu
   pacman -S make mingw-w64-x86_64-toolchain
   ```
4. Add MSYS2 bin to PATH: `C:\msys64\usr\bin`

### 3. Go (Check Version)

Check your Go installation:

```bash
# In Git Bash
go version
```

**Required:** Go 1.21.0 or higher (1.22+ recommended)

If not installed or version is too old:

1. Download from: https://go.dev/dl/
2. Run the installer (`.msi` file)
3. Verify installation: `go version`
4. Ensure GOPATH is set: `go env GOPATH`

### 4. Node.js and npm (Check Version)

Check your Node.js installation:

```bash
# In Git Bash
node --version
npm --version
```

If not installed:

1. Download from: https://nodejs.org/
2. Install the LTS version
3. Verify: `node --version` and `npm --version`

### 5. direnv (Required for Environment Management)

**Required version:** 2.32.0 or higher

#### Installation:

1. Download the latest release from: https://github.com/direnv/direnv/releases
2. Download `direnv.windows-amd64.exe`
3. Rename to `direnv.exe`
4. Move to a directory in your PATH (e.g., `C:\Program Files\Git\usr\bin\`)

#### Configuration:

Add to your `~/.bashrc` (in Git Bash):

```bash
# Open Git Bash and edit .bashrc
nano ~/.bashrc

# Add this line:
eval "$(direnv hook bash)"

# Save and reload:
source ~/.bashrc
```

### 6. Command-Line Tools

#### jq (JSON processor)

**Option A: Chocolatey**

```powershell
choco install jq
```

**Option B: Scoop**

```powershell
scoop install jq
```

**Option C: Manual**

1. Download from: https://jqlang.github.io/jq/download/
2. Download `jq-windows-amd64.exe`
3. Rename to `jq.exe`
4. Move to `C:\Program Files\Git\usr\bin\`

#### curl (Usually included with Git for Windows)

Check if installed: `curl --version`

If not installed:

```powershell
choco install curl
```

#### wget

**Option A: Chocolatey**

```powershell
choco install wget
```

**Option B: Download**

1. Download from: https://eternallybored.org/misc/wget/
2. Extract `wget.exe` to `C:\Program Files\Git\usr\bin\`

#### unzip (Usually included with Git for Windows)

Check if installed: `unzip --version`

If not installed, it's typically at `/usr/bin/unzip` in Git Bash.

### 7. Optional Tools (Recommended)

#### pv (Pipe Viewer)

Download from: http://www.ivarch.com/programs/pv.shtml

#### lz4 (Compression)

**Chocolatey:**

```powershell
choco install lz4
```

**Scoop:**

```powershell
scoop install lz4
```

### 8. Docker Desktop (For Localnet)

**Required for:** Running local blockchain network and integration tests

1. Download Docker Desktop: https://www.docker.com/products/docker-desktop
2. Install with WSL 2 backend enabled
3. Start Docker Desktop
4. Verify: `docker --version` and `docker-compose --version`

**Note:** For best localnet experience, consider using WSL2 (see below).

## Post-Installation Verification

After installing dependencies, run the setup script:

```bash
# In Git Bash
cd /c/Users/YOUR_USERNAME/virtengine/virtengine
./scripts/setup-env-gitbash.sh
```

This will verify all dependencies are correctly installed.

## Using WSL2 (Recommended Alternative)

For the best development experience on Windows, consider using WSL2:

### Install WSL2

```powershell
# In PowerShell (Admin)
wsl --install
```

This installs Ubuntu by default. After reboot:

```bash
# In WSL2 Ubuntu terminal
sudo apt update
sudo apt install -y git build-essential curl wget jq nodejs npm direnv

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install direnv
curl -sfL https://direnv.net/install.sh | bash

# Clone and setup VirtEngine
cd ~
git clone https://github.com/virtengine/virtengine.git
cd virtengine
direnv allow .
```

### WSL2 Advantages

- Better compatibility with shell scripts
- Faster file I/O
- Native Linux tools
- Easier Docker integration
- Better localnet support

## Troubleshooting

### Issue: Command not found after installation

**Solution:** Restart Git Bash terminal or add the tool's directory to PATH.

### Issue: direnv not working

**Solution:** Ensure `eval "$(direnv hook bash)"` is in your `~/.bashrc` and you've restarted your terminal.

### Issue: Make version too old

**Solution:** Ensure you're installing GNU Make 4.0+, not the old 3.x versions.

### Issue: Go commands not working

**Solution:** Ensure Go bin directory is in PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: Permission denied when running scripts

**Solution:**

```bash
chmod +x scripts/setup-env-gitbash.sh
chmod +x scripts/*.sh
```

### Issue: Docker not starting

**Solution:**

- Ensure Hyper-V is enabled (Windows 10 Pro+) or WSL2 is installed (Windows 10 Home+)
- Check Docker Desktop is running
- Restart Docker Desktop

## Quick Install Script (PowerShell - Admin)

Run the install script (located in `scripts/install-deps.ps1`) in PowerShell (Admin):

```powershell
cd C:\Users\YOUR_USERNAME\virtengine\virtengine
.\scripts\install-deps.ps1
```

Or use the manual Chocolatey commands:

```powershell
# Install Chocolatey
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install all dependencies
choco install -y git golang nodejs make jq curl wget docker-desktop
```

## Next Steps

After installing all dependencies:

1. **Restart your terminal** (Git Bash)
2. **Navigate to VirtEngine directory:**
   ```bash
   cd /c/Users/YOUR_USERNAME/virtengine/virtengine
   ```
3. **Run the setup script:**
   ```bash
   ./scripts/setup-env-gitbash.sh
   ```
4. **Follow the on-screen instructions**
5. **Test the build:**
   ```bash
   make virtengine
   ```

## Getting Help

- Check `SETUP-WINDOWS.md` for Windows-specific setup details
- Review `_docs/development-environment.md` for general development setup
- See `CONTRIBUTING.md` for contribution guidelines
