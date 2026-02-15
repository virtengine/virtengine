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
$hasCodexMonitor = $changedFiles | Where-Object { $_ -match '^scripts/codex-monitor/' }
$errors = 0

function Test-CodexMonitorBranchSafety {
    if (-not (Test-Path "scripts/codex-monitor/git-safety.mjs")) {
        Write-Host "  [WARN] git-safety.mjs not found, skipping destructive-branch safety check" -ForegroundColor Yellow
        return $true
    }

    $script = @"
import { pathToFileURL } from 'node:url';
const repoRoot = process.cwd();
const mod = await import(pathToFileURL(repoRoot + '/scripts/codex-monitor/git-safety.mjs').href);
const safety = mod.evaluateBranchSafetyForPush(repoRoot, { baseBranch: 'main' });
if (!safety.safe) {
  console.error('[git-safety] ' + safety.reason);
  process.exit(2);
}
console.log('[git-safety] ok');
"@

    node --input-type=module -e $script 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAIL: destructive branch safety check" -ForegroundColor Red
        Write-Host "  Set VE_ALLOW_DESTRUCTIVE_PUSH=1 only for deliberate emergency bypasses." -ForegroundColor DarkGray
        return $false
    }
    return $true
}

# ── Windows Firewall check (non-blocking) ──────────────────────────────────
if (($IsWindows -or ($env:OS -eq "Windows_NT")) -and ($hasGo -or $hasGoMod)) {
    $fwScript = Join-Path $PSScriptRoot "setup-firewall.ps1"
    if (Test-Path $fwScript) {
        & pwsh -NoProfile -ExecutionPolicy Bypass -File $fwScript -Check 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  [WARN] Windows Firewall not configured — Go tests may trigger popups" -ForegroundColor Yellow
            Write-Host "  Run: make setup-firewall" -ForegroundColor DarkGray
        }
    }
}

if ($hasGo -or $hasGoMod) {
    Write-Host "--- Go checks ---" -ForegroundColor Yellow

    if ($hasGoMod) {
        Write-Host "  go mod tidy..."
        go mod tidy 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go mod tidy" -ForegroundColor Red; $errors++ }

        Write-Host "  go mod vendor..."
        go mod vendor 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go mod vendor" -ForegroundColor Red; $errors++ }
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

if ($hasCodexMonitor) {
    Write-Host "--- Codex Monitor checks ---" -ForegroundColor Yellow

    if (-not (Test-Path "scripts/codex-monitor/node_modules")) {
        Write-Host "  npm install..."
        Push-Location scripts/codex-monitor
        npm install 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: npm install" -ForegroundColor Red; $errors++ }
        Pop-Location
    }

    Write-Host "  Prepublish check..."
    Push-Location scripts/codex-monitor
    node prepublish-check.mjs 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: prepublish check" -ForegroundColor Red; $errors++ }

    Write-Host "  ESM module smoke check (hook-profiles, git-safety)..."
    node --input-type=module -e "await import(new URL('./hook-profiles.mjs', import.meta.url).href); await import(new URL('./git-safety.mjs', import.meta.url).href);" 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: codex-monitor module smoke check" -ForegroundColor Red; $errors++ }
    Pop-Location

    Write-Host "  Branch safety check (deletion-heavy push guard)..."
    if (-not (Test-CodexMonitorBranchSafety)) { $errors++ }
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
