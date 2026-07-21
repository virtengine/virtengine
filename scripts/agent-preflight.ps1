#!/usr/bin/env pwsh
# Agent pre-flight check — run before git push
# Usage: pwsh scripts/agent-preflight.ps1
$ErrorActionPreference = "Continue"

Write-Host "=== Agent Pre-flight Check ===" -ForegroundColor Cyan

# ── Non-interactive git config (prevent editor popups) ──────────────────────
git config --local core.editor ":" 2>$null
git config --local merge.autoEdit false 2>$null
$env:GIT_EDITOR = ":"
$env:GIT_MERGE_AUTOEDIT = "no"

$changedFiles = git diff --cached --name-only 2>$null
if (-not $changedFiles) {
    $changedFiles = git diff --name-only HEAD~1 2>$null
}
if (-not $changedFiles) {
    Write-Host "No changed files detected. Skipping pre-flight."
    exit 0
}

$hasGo = $changedFiles | Where-Object { $_ -match '\.go$' }
$hasPortal = $changedFiles | Where-Object { $_ -match '^portal/' }
$hasGoMod = $changedFiles | Where-Object { $_ -match '^go\.(mod|sum)$' }
$hasBosun = $changedFiles | Where-Object { $_ -match '^scripts/bosun/' }
$errors = 0

# ── Windows Firewall check (non-blocking) ──────────────────────────────────
if (($IsWindows -or ($env:OS -eq "Windows_NT")) -and ($hasGo -or $hasGoMod)) {
    $fwScript = Join-Path $PSScriptRoot "setup-firewall.ps1"
    if (Test-Path $fwScript) {
        $fwResult = & pwsh -NoProfile -ExecutionPolicy Bypass -File $fwScript -Check 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  [WARN] Windows Firewall not configured — Go tests may trigger popups" -ForegroundColor Yellow
            Write-Host "  Run: make setup-firewall" -ForegroundColor DarkGray
        }
    }
}

if ($hasGo -or $hasGoMod) {
    Write-Host "--- Go checks ---" -ForegroundColor Yellow

    if ($hasGoMod) {
        Write-Host "  module/workspace/vendor policy..."
        $gitBash = "C:\Program Files\Git\bin\bash.exe"
        if (-not (Test-Path $gitBash)) { $gitBash = "bash" }
        & $gitBash ./scripts/verify-modules.sh 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: module verification" -ForegroundColor Red; $errors++ }
    }

    $goPkgs = $hasGo | ForEach-Object { "./" + (Split-Path -Parent $_) } | Sort-Object -Unique | Where-Object { $_ -ne "./" }
    $goPkgs = $goPkgs | Where-Object { Test-Path $_ } | Where-Object {
        (Get-ChildItem -Path $_ -Filter *.go -File -ErrorAction SilentlyContinue).Count -gt 0
    }
    $goPkgs = $goPkgs | ForEach-Object { $_ -replace '\\\\', '/' }

    $nonTestGoPkgs = $goPkgs | Where-Object { $_ -notmatch '^\\.\\/tests\\b' }

    if ($goPkgs -and ($nonTestGoPkgs.Count -gt 0)) {
        Write-Host "  gofmt..."
        $hasGo | ForEach-Object { gofmt -w $_ 2>&1 | Out-Null }

        Write-Host "  go vet..."
        go vet @($nonTestGoPkgs) 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go vet" -ForegroundColor Red; $errors++ }

        Write-Host "  go build..."
        go build ./cmd/... 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go build" -ForegroundColor Red; $errors++ }

        Write-Host "  go test (changed packages)..."
        go test -short -count=1 @($nonTestGoPkgs) 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go test" -ForegroundColor Red; $errors++ }
    }
    elseif ($goPkgs) {
        Write-Host "  Skipping Go build/vet/test for tests-only changes." -ForegroundColor DarkGray
    }
}

if ($hasPortal) {
    Write-Host "--- Portal checks ---" -ForegroundColor Yellow

    if (-not (Test-Path "portal/node_modules")) {
        Write-Host "  pnpm install..."
        pnpm -C portal install 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: pnpm install" -ForegroundColor Red; $errors++ }
    }

    Write-Host "  ESLint..."
    pnpm -C portal lint 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: eslint" -ForegroundColor Red; $errors++ }

    Write-Host "  TypeScript..."
    pnpm -C portal type-check 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: tsc" -ForegroundColor Red; $errors++ }

    Write-Host "  Tests..."
    pnpm -C portal test 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: portal tests" -ForegroundColor Red; $errors++ }
}

if ($hasBosun) {
    Write-Host "--- Bosun checks ---" -ForegroundColor Yellow

    if (-not (Test-Path "scripts/bosun/node_modules")) {
        Write-Host "  npm install..."
        Push-Location scripts/bosun
        npm install 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: npm install" -ForegroundColor Red; $errors++ }
        Pop-Location
    }

    Write-Host "  Prepublish check..."
    Push-Location scripts/bosun
    node prepublish-check.mjs 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: prepublish check" -ForegroundColor Red; $errors++ }
    Pop-Location
}

Write-Host ""
if ($errors -gt 0) {
    Write-Host "=== PRE-FLIGHT FAILED: $errors error(s) ===" -ForegroundColor Red
    Write-Host "Fix the issues above before pushing."
    exit 1
}
else {
    Write-Host "=== PRE-FLIGHT PASSED ===" -ForegroundColor Green
    exit 0
}
