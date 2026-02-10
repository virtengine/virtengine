#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Configure Windows Firewall to prevent popup prompts during Go test runs.

.DESCRIPTION
    Go compiles test binaries to unique paths under the Go build cache
    (e.g., %LOCALAPPDATA%\go-build\<hash>\b001\*.test.exe). Each recompilation
    produces a new hash, and Windows Firewall treats every test binary as an
    unknown program — potentially displaying a blocking popup.

    Two-layer fix:

    1. CODE FIX (already applied):
       testutil/network/network.go binds to 127.0.0.1 instead of 0.0.0.0.
       Windows Firewall never prompts for loopback-only listeners.

    2. THIS SCRIPT (defense-in-depth):
       - Disables firewall notification popups for Private/Domain profiles
         via registry. Popups are silently allowed instead of blocking.
       - Adds firewall rules for known Go executables (go.exe, binaries
         in GOPATH/bin and .cache/bin). Note: Windows Firewall does NOT
         support wildcard paths, so per-hash test binaries cannot be
         pre-authorized — the notification suppression handles those.

    Run once after setting up the dev environment:
        pwsh scripts/setup-firewall.ps1

    Or via Make:
        make setup-firewall

.NOTES
    - Only creates rules if they don't already exist
    - Rules persist across reboots (no need to re-run)
    - Safe to run multiple times (idempotent)
    - Requires Administrator for rule creation / registry changes
#>

param(
    [switch]$Remove,
    [switch]$Check,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

# ── Path detection ───────────────────────────────────────────────────────────

$goBuildCache = & go env GOCACHE 2>$null
if (-not $goBuildCache) {
    $goBuildCache = Join-Path $env:LOCALAPPDATA "go-build"
}

$goRoot = & go env GOROOT 2>$null
$goExe = ""
if ($goRoot -and (Test-Path (Join-Path $goRoot "bin\go.exe"))) {
    $goExe = Join-Path $goRoot "bin\go.exe"
}

$goPath = & go env GOPATH 2>$null
if (-not $goPath) {
    $goPath = Join-Path $env:USERPROFILE "go"
}

# Repo-local cache ($PSScriptRoot is the scripts/ dir, one level up is repo root)
$repoRoot = Split-Path -Parent $PSScriptRoot
$repoCacheBin = Join-Path $repoRoot ".cache\bin"

# ── Build rule list (specific executables only — no wildcards) ───────────────

$rules = @()

# Go binary itself
if ($goExe -and (Test-Path $goExe)) {
    $rules += @{
        Name        = "VirtEngine-GoExe-In"
        DisplayName = "VirtEngine: Go toolchain (Inbound)"
        Description = "Allow Go toolchain to accept connections (test servers, build cache)"
        Direction   = "Inbound"
        Program     = $goExe
    }
    $rules += @{
        Name        = "VirtEngine-GoExe-Out"
        DisplayName = "VirtEngine: Go toolchain (Outbound)"
        Description = "Allow Go toolchain outbound connections (module downloads)"
        Direction   = "Outbound"
        Program     = $goExe
    }
}

# Known executables from GOPATH/bin
$gopathBin = Join-Path $goPath "bin"
if (Test-Path $gopathBin) {
    foreach ($exe in (Get-ChildItem -Path $gopathBin -Filter "*.exe" -ErrorAction SilentlyContinue)) {
        $safeName = $exe.BaseName -replace '[^a-zA-Z0-9_-]', '_'
        $rules += @{
            Name        = "VirtEngine-GoPathBin-$safeName"
            DisplayName = "VirtEngine: $($exe.Name) (GOPATH)"
            Description = "Allow $($exe.Name) from GOPATH/bin"
            Direction   = "Inbound"
            Program     = $exe.FullName
        }
    }
}

# Known executables from repo .cache/bin
if (Test-Path $repoCacheBin) {
    foreach ($exe in (Get-ChildItem -Path $repoCacheBin -Filter "*.exe" -ErrorAction SilentlyContinue)) {
        $safeName = $exe.BaseName -replace '[^a-zA-Z0-9_-]', '_'
        $rules += @{
            Name        = "VirtEngine-DevCache-$safeName"
            DisplayName = "VirtEngine: $($exe.Name) (dev cache)"
            Description = "Allow $($exe.Name) from .cache/bin"
            Direction   = "Inbound"
            Program     = $exe.FullName
        }
    }
}

# ── Helpers ──────────────────────────────────────────────────────────────────

function Test-IsAdmin {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = [Security.Principal.WindowsPrincipal]$identity
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-NotificationsDisabled {
    try {
        foreach ($profile in @("StandardProfile", "DomainProfile")) {
            $regPath = "HKLM:\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\FirewallPolicy\$profile"
            $val = Get-ItemProperty -Path $regPath -Name "DisableNotifications" -ErrorAction SilentlyContinue
            if (-not $val -or $val.DisableNotifications -ne 1) {
                return $false
            }
        }
        return $true
    } catch {
        return $false
    }
}

# ── Check mode ───────────────────────────────────────────────────────────────

if ($Check) {
    Write-Host "=== VirtEngine Firewall Status ===" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Go build cache:  $goBuildCache"
    Write-Host "Go executable:   $(if ($goExe) { $goExe } else { '(not found)' })"
    Write-Host "GOPATH:          $goPath"
    Write-Host "Repo cache:      $repoCacheBin"
    Write-Host ""

    $allGood = $true

    # Check notification suppression (most important)
    if (Test-NotificationsDisabled) {
        Write-Host "  [OK] Firewall notifications disabled (Private/Domain)" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] Firewall notifications still enabled (popups will appear)" -ForegroundColor Red
        $allGood = $false
    }

    # Check executable rules
    foreach ($rule in $rules) {
        $existing = Get-NetFirewallRule -Name $rule.Name -ErrorAction SilentlyContinue
        if ($existing) {
            Write-Host "  [OK] $($rule.DisplayName)" -ForegroundColor Green
        } else {
            Write-Host "  [MISSING] $($rule.DisplayName)" -ForegroundColor Yellow
            # Rules for specific executables are nice-to-have, not critical
        }
    }

    # Check loopback binding in test code
    $networkFile = Join-Path $repoRoot "testutil\network\network.go"
    if (Test-Path $networkFile) {
        $content = Get-Content $networkFile -Raw
        if ($content -match '0\.0\.0\.0') {
            Write-Host "  [WARN] testutil/network/network.go still binds to 0.0.0.0" -ForegroundColor Red
            $allGood = $false
        } else {
            Write-Host "  [OK] testutil/network/network.go uses loopback binding" -ForegroundColor Green
        }
    }

    Write-Host ""
    if ($allGood) {
        Write-Host "Firewall is configured. No popups expected." -ForegroundColor Green
    } else {
        Write-Host "Configuration incomplete. Run: pwsh scripts/setup-firewall.ps1" -ForegroundColor Yellow
    }
    exit ($allGood ? 0 : 1)
}

# ── Remove mode ──────────────────────────────────────────────────────────────

if ($Remove) {
    if (-not (Test-IsAdmin)) {
        Write-Host "ERROR: Requires Administrator privileges." -ForegroundColor Red
        exit 1
    }

    Write-Host "Removing VirtEngine firewall configuration..." -ForegroundColor Yellow

    # Remove named rules
    foreach ($rule in $rules) {
        $existing = Get-NetFirewallRule -Name $rule.Name -ErrorAction SilentlyContinue
        if ($existing) {
            Remove-NetFirewallRule -Name $rule.Name
            Write-Host "  Removed: $($rule.DisplayName)" -ForegroundColor Green
        }
    }

    # Remove any other VirtEngine-* rules
    Get-NetFirewallRule -Name "VirtEngine-*" -ErrorAction SilentlyContinue |
        ForEach-Object {
            Remove-NetFirewallRule -Name $_.Name
            Write-Host "  Removed: $($_.DisplayName)" -ForegroundColor Green
        }

    # Re-enable notifications
    foreach ($profile in @("StandardProfile", "DomainProfile")) {
        $regPath = "HKLM:\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\FirewallPolicy\$profile"
        Set-ItemProperty -Path $regPath -Name "DisableNotifications" -Value 0 -ErrorAction SilentlyContinue
    }
    Write-Host "  Re-enabled firewall notifications" -ForegroundColor Green

    Write-Host "Done." -ForegroundColor Green
    exit 0
}

# ── Install mode (default) ──────────────────────────────────────────────────

# Self-elevate if needed
if (-not (Test-IsAdmin)) {
    Write-Host ""
    Write-Host "=== VirtEngine Firewall Setup ===" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Configures Windows Firewall to prevent popups during Go tests." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Changes:" -ForegroundColor White
    Write-Host "  1. Suppress firewall notification popups (Private/Domain profiles)" -ForegroundColor DarkGray
    Write-Host "  2. Allow Go toolchain and known binaries through firewall" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Note: Test code already binds to 127.0.0.1 (loopback) which doesn't" -ForegroundColor DarkGray
    Write-Host "trigger popups. These settings are a safety net for edge cases." -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Requesting elevation..." -ForegroundColor Cyan

    $scriptPath = $PSCommandPath
    $argList = "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`""
    if ($Force) { $argList += " -Force" }

    try {
        Start-Process pwsh -Verb RunAs -ArgumentList $argList -Wait
        Write-Host ""
        if (Test-NotificationsDisabled) {
            Write-Host "Firewall configured successfully." -ForegroundColor Green
        } else {
            Write-Host "Elevation completed. Verify with: pwsh scripts/setup-firewall.ps1 -Check" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "Elevation cancelled or failed: $_" -ForegroundColor Red
        Write-Host ""
        Write-Host "Run manually from an admin terminal:" -ForegroundColor Yellow
        Write-Host "  pwsh scripts/setup-firewall.ps1" -ForegroundColor White
        exit 1
    }
    exit 0
}

# ── Running as admin — apply configuration ──────────────────────────────────

Write-Host ""
Write-Host "=== Applying VirtEngine Firewall Configuration ===" -ForegroundColor Cyan
Write-Host ""

$changed = 0
$skipped = 0

# STEP 1: Disable firewall notification popups.
# This is the PRIMARY mechanism — Windows Firewall will silently allow
# connections from unrecognized programs instead of displaying a blocking
# popup. This handles Go test binaries (which get unique paths per build)
# that cannot be pre-authorized via firewall rules.
Write-Host "--- Step 1: Disable notification popups ---" -ForegroundColor Yellow

foreach ($entry in @(
    @{ Name = "StandardProfile"; Label = "Private" },
    @{ Name = "DomainProfile"; Label = "Domain" }
)) {
    $regPath = "HKLM:\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\FirewallPolicy\$($entry.Name)"
    try {
        $current = Get-ItemProperty -Path $regPath -Name "DisableNotifications" -ErrorAction SilentlyContinue
        if ($current -and $current.DisableNotifications -eq 1 -and -not $Force) {
            Write-Host "  [EXISTS] $($entry.Label) notifications already disabled" -ForegroundColor DarkGray
            $skipped++
        } else {
            Set-ItemProperty -Path $regPath -Name "DisableNotifications" -Value 1 -Type DWord
            Write-Host "  [SET] $($entry.Label) notifications disabled" -ForegroundColor Green
            $changed++
        }
    } catch {
        Write-Host "  [FAILED] $($entry.Label): $_" -ForegroundColor Red
    }
}

# STEP 2: Add rules for specific known executables (no wildcards).
# Windows Firewall's New-NetFirewallRule -Program does NOT support wildcard
# paths (e.g., "C:\go-build\*"). We only add rules for fixed-path binaries.
Write-Host ""
Write-Host "--- Step 2: Firewall rules for known executables ---" -ForegroundColor Yellow

if ($rules.Count -eq 0) {
    Write-Host "  (no known executables found to add rules for)" -ForegroundColor DarkGray
}

foreach ($rule in $rules) {
    $existing = Get-NetFirewallRule -Name $rule.Name -ErrorAction SilentlyContinue
    if ($existing -and -not $Force) {
        Write-Host "  [EXISTS] $($rule.DisplayName)" -ForegroundColor DarkGray
        $skipped++
        continue
    }

    if ($existing -and $Force) {
        Remove-NetFirewallRule -Name $rule.Name -ErrorAction SilentlyContinue
    }

    if (-not (Test-Path $rule.Program)) {
        Write-Host "  [SKIP] $($rule.DisplayName) — not found: $($rule.Program)" -ForegroundColor DarkGray
        continue
    }

    try {
        New-NetFirewallRule `
            -Name $rule.Name `
            -DisplayName $rule.DisplayName `
            -Description $rule.Description `
            -Direction $rule.Direction `
            -Action Allow `
            -Program $rule.Program `
            -Protocol TCP `
            -Profile Private, Domain `
            -Enabled True `
            -ErrorAction Stop | Out-Null

        Write-Host "  [CREATED] $($rule.DisplayName)" -ForegroundColor Green
        $changed++
    } catch {
        Write-Host "  [FAILED] $($rule.DisplayName): $_" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Done: $changed configured, $skipped already existed." -ForegroundColor Cyan
Write-Host ""
Write-Host "Go test binaries will no longer trigger firewall popups." -ForegroundColor Green
Write-Host "  Verify: pwsh scripts/setup-firewall.ps1 -Check" -ForegroundColor DarkGray
Write-Host "  Remove: pwsh scripts/setup-firewall.ps1 -Remove" -ForegroundColor DarkGray
