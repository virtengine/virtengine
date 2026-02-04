# Build and Test Script
$ErrorActionPreference = 'Continue'
Set-Location 'C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine'

Write-Host "=== BUILD START ==="
$buildOutput = go build ./x/bme/... ./x/oracle/... 2>&1
$buildCode = $LASTEXITCODE
if ($buildOutput) { Write-Host $buildOutput }
Write-Host "BUILD EXIT CODE: $buildCode"
if ($buildCode -eq 0) { Write-Host "BUILD: PASS" } else { Write-Host "BUILD: FAIL" }

Write-Host ""
Write-Host "=== TEST START ==="
go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1
$testCode = $LASTEXITCODE
Write-Host "TEST EXIT CODE: $testCode"
if ($testCode -eq 0) { Write-Host "TEST: PASS" } else { Write-Host "TEST: FAIL" }

Write-Host ""
Write-Host "=== SUMMARY ==="
Write-Host "Build: $(if($buildCode -eq 0){'PASS'}else{'FAIL'})"
Write-Host "Test: $(if($testCode -eq 0){'PASS'}else{'FAIL'})"
