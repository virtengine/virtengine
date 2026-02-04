$ErrorActionPreference = "Continue"
Set-Location "C:\Users\jonathan\AppData\Local\Temp\vibe-kanban\worktrees\ed79-feat-bme-oracle\virtengine"

$resultsFile = "results.txt"

# Run build command
Write-Host "Starting build..."
$buildOutput = go build ./x/bme/... ./x/oracle/... 2>&1 | Out-String
$buildExitCode = $LASTEXITCODE
Write-Host "Build completed with exit code: $buildExitCode"

# Run test command  
Write-Host "Starting tests..."
$testOutput = go test ./x/bme/keeper/... ./x/oracle/keeper/... 2>&1 | Out-String
$testExitCode = $LASTEXITCODE
Write-Host "Tests completed with exit code: $testExitCode"

# Determine status
$buildStatus = if ($buildExitCode -eq 0) { "PASSED" } else { "FAILED" }
$testStatus = if ($testExitCode -eq 0) { "PASSED" } else { "FAILED" }

# Write results
$results = @"
=== BUILD RESULTS ===
$(if ([string]::IsNullOrWhiteSpace($buildOutput)) { "Build passed with no output" } else { $buildOutput.Trim() })
BUILD EXIT CODE: $buildExitCode
BUILD STATUS: $buildStatus

=== TEST RESULTS ===
$($testOutput.Trim())
TEST EXIT CODE: $testExitCode
TEST STATUS: $testStatus
"@

$results | Out-File -FilePath $resultsFile -Encoding UTF8
Write-Host "Results written to $resultsFile"
Get-Content $resultsFile
