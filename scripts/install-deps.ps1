# VirtEngine Dependencies Installation Script for Windows
# Run this script in PowerShell as Administrator

#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "VirtEngine Development Dependencies Installer for Windows" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""

# Function to check if a command exists
function Test-CommandExists {
    param($command)
    $null = Get-Command $command -ErrorAction SilentlyContinue
    return $?
}

# Function to print status
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    switch ($Status) {
        "OK" { Write-Host "[OK] " -ForegroundColor Green -NoNewline; Write-Host $Message }
        "WARN" { Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline; Write-Host $Message }
        "ERROR" { Write-Host "[ERROR] " -ForegroundColor Red -NoNewline; Write-Host $Message }
        "INFO" { Write-Host "[INFO] " -ForegroundColor Cyan -NoNewline; Write-Host $Message }
    }
}

# Check if running as Administrator
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Status "ERROR" "This script must be run as Administrator"
    Write-Host "Please right-click on PowerShell and select 'Run as Administrator'"
    exit 1
}

Write-Status "OK" "Running as Administrator"
Write-Host ""

# Check for Chocolatey
Write-Host "Checking for Chocolatey package manager..." -ForegroundColor Cyan
if (Test-CommandExists choco) {
    $chocoVersion = (choco --version)
    Write-Status "OK" "Chocolatey is already installed: $chocoVersion"
} else {
    Write-Status "WARN" "Chocolatey is not installed. Installing..."
    
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    
    try {
        Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        
        # Refresh environment
        $env:ChocolateyInstall = Convert-Path "$((Get-Command choco).Path)\..\.."
        Import-Module "$env:ChocolateyInstall\helpers\chocolateyProfile.psm1"
        Update-SessionEnvironment
        
        Write-Status "OK" "Chocolatey installed successfully"
    } catch {
        Write-Status "ERROR" "Failed to install Chocolatey: $_"
        Write-Host "Please install Chocolatey manually from: https://chocolatey.org/install"
        exit 1
    }
}

Write-Host ""
Write-Host "Installing VirtEngine dependencies..." -ForegroundColor Cyan
Write-Host ""

# List of packages to install
$packages = @(
    @{Name="git"; DisplayName="Git for Windows"; Required=$true},
    @{Name="golang"; DisplayName="Go Programming Language"; Required=$true; MinVersion="1.21.0"},
    @{Name="nodejs"; DisplayName="Node.js and npm"; Required=$true},
    @{Name="make"; DisplayName="GNU Make"; Required=$true},
    @{Name="jq"; DisplayName="jq (JSON processor)"; Required=$true},
    @{Name="curl"; DisplayName="curl"; Required=$false},
    @{Name="wget"; DisplayName="wget"; Required=$false},
    @{Name="docker-desktop"; DisplayName="Docker Desktop"; Required=$false}
)

$installCount = 0
$skipCount = 0
$errorCount = 0

foreach ($package in $packages) {
    $pkgName = $package.Name
    $displayName = $package.DisplayName
    $required = $package.Required
    
    Write-Host "Checking $displayName..." -NoNewline
    
    # Check if already installed via choco
    $chocoList = choco list --local-only $pkgName
    if ($chocoList -match $pkgName) {
        Write-Host " " -NoNewline
        Write-Status "OK" "Already installed"
        $skipCount++
    } else {
        Write-Host ""
        
        if ($required) {
            Write-Status "INFO" "Installing $displayName (required)..."
        } else {
            Write-Status "INFO" "Installing $displayName (optional)..."
        }
        
        try {
            choco install $pkgName -y --no-progress
            if ($LASTEXITCODE -eq 0) {
                Write-Status "OK" "$displayName installed successfully"
                $installCount++
            } else {
                Write-Status "WARN" "$displayName installation had warnings"
                $installCount++
            }
        } catch {
            $errMsg = $_.Exception.Message
            Write-Status "ERROR" "Failed to install ${displayName}: $errMsg"
            $errorCount++
            if ($required) {
                Write-Host "This is a required package. Installation cannot continue."
                exit 1
            }
        }
    }
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Installing direnv (manual download required)" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

# direnv requires manual installation
$direnvPath = "C:\Program Files\Git\usr\bin\direnv.exe"
if (Test-Path $direnvPath) {
    Write-Status "OK" "direnv is already installed at $direnvPath"
} else {
    Write-Status "WARN" "direnv needs to be installed manually"
    Write-Host ""
    Write-Host "Please follow these steps to install direnv:" -ForegroundColor Yellow
    Write-Host "1. Download from: https://github.com/direnv/direnv/releases" -ForegroundColor Yellow
    Write-Host "2. Download 'direnv.windows-amd64.exe'" -ForegroundColor Yellow
    Write-Host "3. Rename to 'direnv.exe'" -ForegroundColor Yellow
    Write-Host "4. Move to: C:\Program Files\Git\usr\bin\" -ForegroundColor Yellow
    Write-Host ""
    
    $response = Read-Host "Would you like to open the direnv releases page now? (Y/N)"
    if ($response -eq 'Y' -or $response -eq 'y') {
        Start-Process "https://github.com/direnv/direnv/releases"
    }
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Installation Summary" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Status "INFO" "Packages installed: $installCount"
Write-Status "INFO" "Packages already present: $skipCount"
if ($errorCount -gt 0) {
    Write-Status "WARN" "Packages failed: $errorCount"
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Post-Installation Steps" -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "1. " -NoNewline
Write-Host "Refresh your PATH:" -ForegroundColor Yellow
Write-Host "   Close and reopen PowerShell/Git Bash to refresh environment variables"
Write-Host ""

Write-Host "2. " -NoNewline
Write-Host "Configure direnv in Git Bash:" -ForegroundColor Yellow
Write-Host "   Open Git Bash and add to ~/.bashrc:"
Write-Host '   eval "$(direnv hook bash)"' -ForegroundColor Cyan
Write-Host ""

Write-Host "3. " -NoNewline
Write-Host "Verify installations in Git Bash:" -ForegroundColor Yellow
Write-Host "   git --version" -ForegroundColor Cyan
Write-Host "   go version" -ForegroundColor Cyan
Write-Host "   node --version" -ForegroundColor Cyan
Write-Host "   make --version" -ForegroundColor Cyan
Write-Host "   jq --version" -ForegroundColor Cyan
Write-Host "   direnv --version" -ForegroundColor Cyan
Write-Host ""

Write-Host "4. " -NoNewline
Write-Host "Run VirtEngine setup script:" -ForegroundColor Yellow
Write-Host "   cd /c/Users/YOUR_USERNAME/virtengine/virtengine" -ForegroundColor Cyan
Write-Host "   ./setup-env-gitbash.sh" -ForegroundColor Cyan
Write-Host ""

Write-Host "5. " -NoNewline
Write-Host "Enable pnpm and install frontend deps:" -ForegroundColor Yellow
Write-Host "   corepack enable" -ForegroundColor Cyan
Write-Host "   corepack prepare pnpm@latest --activate" -ForegroundColor Cyan
Write-Host "   pnpm -C portal install" -ForegroundColor Cyan
Write-Host "   pnpm -C sdk/ts install" -ForegroundColor Cyan
Write-Host ""

Write-Host "6. " -NoNewline
Write-Host "Build VirtEngine:" -ForegroundColor Yellow
Write-Host "   make virtengine" -ForegroundColor Cyan
Write-Host ""

if ($errorCount -eq 0) {
    Write-Host "=================================================================" -ForegroundColor Green
    Write-Host "Installation completed successfully!" -ForegroundColor Green
    Write-Host "=================================================================" -ForegroundColor Green
} else {
    Write-Host "=================================================================" -ForegroundColor Yellow
    Write-Host "Installation completed with warnings" -ForegroundColor Yellow
    Write-Host "Please review the errors above and install missing packages manually" -ForegroundColor Yellow
    Write-Host "=================================================================" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "For detailed setup instructions, see:" -ForegroundColor Cyan
Write-Host "  - SETUP-WINDOWS.md" -ForegroundColor Cyan
Write-Host "  - INSTALL-DEPS-WINDOWS.md" -ForegroundColor Cyan
Write-Host "  - _docs/development-environment.md" -ForegroundColor Cyan
Write-Host ""

# Ask if user wants to open Git Bash
$response = Read-Host "Would you like to open Git Bash now? (Y/N)"
if ($response -eq 'Y' -or $response -eq 'y') {
    $gitBashPath = "C:\Program Files\Git\git-bash.exe"
    if (Test-Path $gitBashPath) {
        Start-Process $gitBashPath
    } else {
        Write-Status "WARN" "Git Bash not found at expected location: $gitBashPath"
    }
}

Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
