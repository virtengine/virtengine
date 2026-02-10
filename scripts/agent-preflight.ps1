#!/usr/bin/env pwsh
# Agent pre-flight check — run before git push
# Usage: pwsh scripts/agent-preflight.ps1
$ErrorActionPreference = "Continue"

Write-Host "=== Agent Pre-flight Check ===" -ForegroundColor Cyan

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
$errors = 0

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

    if ($goPkgs) {
        Write-Host "  gofmt..."
        $hasGo | ForEach-Object { gofmt -w $_ 2>&1 | Out-Null }

        Write-Host "  go vet..."
        go vet @($goPkgs) 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go vet" -ForegroundColor Red; $errors++ }

        Write-Host "  go build..."
        go build ./cmd/... 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go build" -ForegroundColor Red; $errors++ }

        Write-Host "  go test (changed packages)..."
        go test -short -count=1 @($goPkgs) 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { Write-Host "FAIL: go test" -ForegroundColor Red; $errors++ }
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
